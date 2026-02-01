// Package servicedesk provides event listening for support module chain events.
//
// VE-12B: Chain event listener for automatic service desk synchronization.
package servicedesk

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/rpc/client"
	ctypes "github.com/cometbft/cometbft/types"
)

// ChainEventListener listens for support-related events from the chain
// and forwards them to the service desk bridge for external sync.
type ChainEventListener struct {
	bridge     IBridge
	client     client.Client
	logger     log.Logger
	config     *EventListenerConfig
	wsEndpoint string

	mu       sync.RWMutex
	running  bool
	stopCh   chan struct{}
	wg       sync.WaitGroup
	eventSub ctypes.Subscription
}

// EventListenerConfig holds configuration for the chain event listener
type EventListenerConfig struct {
	// BufferSize is the event buffer size
	BufferSize int `json:"buffer_size"`

	// ReconnectDelay is the delay before reconnecting on disconnect
	ReconnectDelay time.Duration `json:"reconnect_delay"`

	// MaxReconnectAttempts is the maximum reconnection attempts
	MaxReconnectAttempts int `json:"max_reconnect_attempts"`

	// EventTypes are the specific event types to listen for
	EventTypes []string `json:"event_types"`
}

// DefaultEventListenerConfig returns a default event listener configuration
func DefaultEventListenerConfig() *EventListenerConfig {
	return &EventListenerConfig{
		BufferSize:           1000,
		ReconnectDelay:       5 * time.Second,
		MaxReconnectAttempts: 10,
		EventTypes: []string{
			"external_ticket_registered",
			"external_ticket_updated",
			"external_ticket_removed",
			"support.ticket.created",
			"support.ticket.updated",
			"support.ticket.closed",
		},
	}
}

// NewChainEventListener creates a new chain event listener
func NewChainEventListener(bridge IBridge, wsEndpoint string, logger log.Logger, config *EventListenerConfig) *ChainEventListener {
	if config == nil {
		config = DefaultEventListenerConfig()
	}

	return &ChainEventListener{
		bridge:     bridge,
		wsEndpoint: wsEndpoint,
		logger:     logger.With("component", "chain_event_listener"),
		config:     config,
		stopCh:     make(chan struct{}),
	}
}

// Start starts the event listener
func (l *ChainEventListener) Start(ctx context.Context) error {
	l.mu.Lock()
	if l.running {
		l.mu.Unlock()
		return nil
	}
	l.running = true
	l.mu.Unlock()

	l.logger.Info("starting chain event listener", "endpoint", l.wsEndpoint)

	// Start the event loop
	l.wg.Add(1)
	go l.eventLoop(ctx)

	return nil
}

// Stop stops the event listener
func (l *ChainEventListener) Stop(ctx context.Context) error {
	l.mu.Lock()
	if !l.running {
		l.mu.Unlock()
		return nil
	}
	l.running = false
	l.mu.Unlock()

	l.logger.Info("stopping chain event listener")

	// Signal stop
	close(l.stopCh)

	// Unsubscribe from events
	if l.eventSub != nil && l.client != nil {
		_ = l.client.Unsubscribe(ctx, "servicedesk", "tm.event = 'Tx'")
	}

	// Wait for goroutines
	done := make(chan struct{})
	go func() {
		l.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	}

	l.logger.Info("chain event listener stopped")
	return nil
}

// eventLoop is the main event processing loop
func (l *ChainEventListener) eventLoop(ctx context.Context) {
	defer l.wg.Done()

	for attempt := 0; attempt < l.config.MaxReconnectAttempts; attempt++ {
		select {
		case <-l.stopCh:
			return
		case <-ctx.Done():
			return
		default:
		}

		if err := l.connectAndListen(ctx); err != nil {
			l.logger.Error("event listener connection error",
				"error", err,
				"attempt", attempt+1,
			)

			// Wait before reconnecting
			select {
			case <-l.stopCh:
				return
			case <-ctx.Done():
				return
			case <-time.After(l.config.ReconnectDelay):
			}
		}
	}

	l.logger.Error("max reconnect attempts reached, stopping event listener")
}

// connectAndListen connects to the chain and listens for events
func (l *ChainEventListener) connectAndListen(ctx context.Context) error {
	// In a real implementation, this would:
	// 1. Connect to CometBFT WebSocket
	// 2. Subscribe to transaction events
	// 3. Filter for support module events
	// 4. Parse and forward to bridge

	l.logger.Debug("connecting to chain websocket", "endpoint", l.wsEndpoint)

	// For now, this is a placeholder that would be implemented
	// with actual CometBFT client connection

	// Simulate waiting for events
	for {
		select {
		case <-l.stopCh:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// processEvent processes a single chain event
func (l *ChainEventListener) processEvent(ctx context.Context, event abci.Event, blockHeight int64, txHash string) error {
	eventType := event.Type

	// Check if this is a support module event
	if !l.isSupportEvent(eventType) {
		return nil
	}

	l.logger.Debug("processing support event",
		"type", eventType,
		"block_height", blockHeight,
		"tx_hash", txHash,
	)

	// Parse event attributes
	attrs := make(map[string]string)
	for _, attr := range event.Attributes {
		attrs[attr.Key] = attr.Value
	}

	// Route to appropriate handler
	switch eventType {
	case "external_ticket_registered", "support.ticket.created":
		return l.handleTicketCreated(ctx, attrs, blockHeight, txHash)
	case "external_ticket_updated", "support.ticket.updated":
		return l.handleTicketUpdated(ctx, attrs, blockHeight, txHash)
	case "external_ticket_removed", "support.ticket.closed":
		return l.handleTicketClosed(ctx, attrs, blockHeight, txHash)
	default:
		l.logger.Debug("unhandled event type", "type", eventType)
		return nil
	}
}

// isSupportEvent checks if an event type is a support module event
func (l *ChainEventListener) isSupportEvent(eventType string) bool {
	for _, et := range l.config.EventTypes {
		if et == eventType {
			return true
		}
	}
	return false
}

// handleTicketCreated handles a ticket created event
func (l *ChainEventListener) handleTicketCreated(ctx context.Context, attrs map[string]string, blockHeight int64, txHash string) error {
	event := &TicketCreatedEvent{
		TicketID:        attrs["ticket_id"],
		CustomerAddress: attrs["customer_address"],
		ProviderAddress: attrs["provider_address"],
		Category:        attrs["category"],
		Priority:        attrs["priority"],
		Subject:         attrs["subject"],
		Description:     attrs["description"],
		BlockHeight:     blockHeight,
		Timestamp:       time.Now(),
		TxHash:          txHash,
	}

	// Validate required fields
	if event.TicketID == "" {
		event.TicketID = attrs["resource_id"]
	}
	if event.TicketID == "" {
		return fmt.Errorf("ticket_id not found in event attributes")
	}

	return l.bridge.HandleTicketCreated(ctx, event)
}

// handleTicketUpdated handles a ticket updated event
func (l *ChainEventListener) handleTicketUpdated(ctx context.Context, attrs map[string]string, blockHeight int64, txHash string) error {
	event := &TicketUpdatedEvent{
		TicketID:    attrs["ticket_id"],
		UpdatedBy:   attrs["updated_by"],
		BlockHeight: blockHeight,
		Timestamp:   time.Now(),
		TxHash:      txHash,
	}

	// Parse changes
	event.Changes = make(map[string]interface{})
	if status := attrs["status"]; status != "" {
		event.Changes["status"] = status
	}
	if priority := attrs["priority"]; priority != "" {
		event.Changes["priority"] = priority
	}
	if assignee := attrs["assignee"]; assignee != "" {
		event.Changes["assignee"] = assignee
	}

	if event.TicketID == "" {
		event.TicketID = attrs["resource_id"]
	}
	if event.TicketID == "" {
		return fmt.Errorf("ticket_id not found in event attributes")
	}

	return l.bridge.HandleTicketUpdated(ctx, event)
}

// handleTicketClosed handles a ticket closed event
func (l *ChainEventListener) handleTicketClosed(ctx context.Context, attrs map[string]string, blockHeight int64, txHash string) error {
	event := &TicketClosedEvent{
		TicketID:    attrs["ticket_id"],
		ClosedBy:    attrs["closed_by"],
		Resolution:  attrs["resolution"],
		BlockHeight: blockHeight,
		Timestamp:   time.Now(),
		TxHash:      txHash,
	}

	if event.TicketID == "" {
		event.TicketID = attrs["resource_id"]
	}
	if event.TicketID == "" {
		return fmt.Errorf("ticket_id not found in event attributes")
	}

	return l.bridge.HandleTicketClosed(ctx, event)
}

// IChainEventListener is the interface for chain event listeners
type IChainEventListener interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Ensure ChainEventListener implements IChainEventListener
var _ IChainEventListener = (*ChainEventListener)(nil)
