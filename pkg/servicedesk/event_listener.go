// Package servicedesk provides event listening for support module chain events.
//
// VE-12B: Chain event listener for automatic service desk synchronization.
package servicedesk

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"cosmossdk.io/log"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	rpctypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"

	supporttypes "github.com/virtengine/virtengine/x/support/types"
)

// ChainEventListener listens for support-related events from the chain
// and forwards them to the service desk bridge for external sync.
type ChainEventListener struct {
	bridge     IBridge
	client     *rpchttp.HTTP
	logger     log.Logger
	config     *EventListenerConfig
	wsEndpoint string
	decryptor  PayloadDecryptor

	mu       sync.RWMutex
	running  bool
	stopCh   chan struct{}
	wg       sync.WaitGroup
	eventSub <-chan rpctypes.ResultEvent
}

// EventListenerConfig holds configuration for the chain event listener
type EventListenerConfig struct {
	// BufferSize is the event buffer size
	BufferSize int `json:"buffer_size"`

	// ReconnectDelay is the delay before reconnecting on disconnect
	ReconnectDelay time.Duration `json:"reconnect_delay"`

	// MaxReconnectAttempts is the maximum reconnection attempts
	MaxReconnectAttempts int `json:"max_reconnect_attempts"`

	// SubscriberID is the subscription ID
	SubscriberID string `json:"subscriber_id"`

	// CometWS is the websocket endpoint path
	CometWS string `json:"comet_ws"`

	// EventQuery is the CometBFT query for events
	EventQuery string `json:"event_query"`

	// EventTypes are the specific event types to listen for
	EventTypes []string `json:"event_types"`
}

// DefaultEventListenerConfig returns a default event listener configuration
func DefaultEventListenerConfig() *EventListenerConfig {
	return &EventListenerConfig{
		BufferSize:           1000,
		ReconnectDelay:       5 * time.Second,
		MaxReconnectAttempts: 10,
		SubscriberID:         "servicedesk",
		CometWS:              "/websocket",
		EventTypes: []string{
			string(supporttypes.SupportEventTypeRequestCreated),
			string(supporttypes.SupportEventTypeRequestUpdated),
			string(supporttypes.SupportEventTypeStatusChanged),
			string(supporttypes.SupportEventTypeResponseAdded),
			string(supporttypes.SupportEventTypeRequestArchived),
			string(supporttypes.SupportEventTypeRequestPurged),
			string(supporttypes.SupportEventTypeExternalTicketLinked),
		},
	}
}

// NewChainEventListener creates a new chain event listener
func NewChainEventListener(bridge IBridge, wsEndpoint string, logger log.Logger, config *EventListenerConfig) *ChainEventListener {
	if config == nil {
		config = DefaultEventListenerConfig()
	}
	if config.CometWS == "" {
		config.CometWS = "/websocket"
	}
	if config.SubscriberID == "" {
		config.SubscriberID = "servicedesk"
	}
	if config.EventQuery == "" {
		config.EventQuery = defaultSupportEventQuery(config.EventTypes)
	}

	return &ChainEventListener{
		bridge:     bridge,
		wsEndpoint: wsEndpoint,
		logger:     logger.With("component", "chain_event_listener"),
		config:     config,
		decryptor:  bridge.Decryptor(),
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
		_ = l.client.Unsubscribe(ctx, l.config.SubscriberID, l.config.EventQuery)
		_ = l.client.Stop()
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
	l.logger.Debug("connecting to chain websocket", "endpoint", l.wsEndpoint)

	rpc, err := rpchttp.New(l.wsEndpoint, l.config.CometWS)
	if err != nil {
		return fmt.Errorf("create comet rpc client: %w", err)
	}
	if err := rpc.Start(); err != nil {
		return fmt.Errorf("start comet rpc client: %w", err)
	}

	l.client = rpc

	sub, err := rpc.Subscribe(ctx, l.config.SubscriberID, l.config.EventQuery, l.config.BufferSize)
	if err != nil {
		return fmt.Errorf("subscribe to support events: %w", err)
	}
	l.eventSub = sub

	for {
		select {
		case <-l.stopCh:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-sub:
			if !ok {
				return fmt.Errorf("event subscription closed")
			}
			data, ok := msg.Data.(tmtypes.EventDataTx)
			if !ok {
				continue
			}
			txHash := fmt.Sprintf("%X", tmtypes.Tx(data.Tx).Hash())
			envelopes, err := ExtractSupportEvents(data.Result.Events)
			if err != nil {
				l.logger.Error("failed to extract support events", "error", err)
				continue
			}
			for _, env := range envelopes {
				if !l.isSupportEvent(env.EventType) {
					continue
				}
				if err := l.handleSupportEnvelope(ctx, env, data.Height, txHash); err != nil {
					l.logger.Error("failed to process support event", "error", err)
				}
			}
		}
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

func (l *ChainEventListener) handleSupportEnvelope(ctx context.Context, env SupportEventEnvelope, blockHeight int64, txHash string) error {
	switch env.EventType {
	case string(supporttypes.SupportEventTypeRequestCreated):
		return l.handleSupportRequestCreated(ctx, env, blockHeight, txHash)
	case string(supporttypes.SupportEventTypeRequestUpdated):
		return l.handleSupportRequestUpdated(ctx, env, blockHeight, txHash)
	case string(supporttypes.SupportEventTypeStatusChanged):
		return l.handleSupportStatusChanged(ctx, env, blockHeight, txHash)
	case string(supporttypes.SupportEventTypeResponseAdded):
		return l.handleSupportResponseAdded(ctx, env, blockHeight, txHash)
	default:
		return nil
	}
}

func (l *ChainEventListener) handleSupportRequestCreated(ctx context.Context, env SupportEventEnvelope, blockHeight int64, txHash string) error {
	var payload supporttypes.SupportRequestCreatedEvent
	if err := json.Unmarshal([]byte(env.PayloadJSON), &payload); err != nil {
		return fmt.Errorf("decode request created payload: %w", err)
	}

	subject := ""
	description := ""
	if l.decryptor != nil && payload.Payload != nil {
		decoded, err := l.decryptor.DecryptSupportRequestPayload(ctx, payload.Payload)
		if err != nil {
			l.logger.Error("failed to decrypt support payload", "error", err)
		} else if decoded != nil {
			subject = decoded.Subject
			description = decoded.Description
		}
	}

	event := &TicketCreatedEvent{
		TicketID:        payload.TicketID,
		TicketNumber:    payload.TicketNumber,
		CustomerAddress: payload.Submitter,
		Category:        payload.Category,
		Priority:        payload.Priority,
		Subject:         subject,
		Description:     description,
		BlockHeight:     blockHeight,
		Timestamp:       time.Unix(payload.Timestamp, 0).UTC(),
		TxHash:          txHash,
	}
	if payload.RelatedEntity != nil {
		event.RelatedEntity = &RelatedEntity{
			Type: string(payload.RelatedEntity.Type),
			ID:   payload.RelatedEntity.ID,
		}
	}

	return l.bridge.HandleTicketCreated(ctx, event)
}

func (l *ChainEventListener) handleSupportRequestUpdated(ctx context.Context, env SupportEventEnvelope, blockHeight int64, txHash string) error {
	var payload supporttypes.SupportRequestUpdatedEvent
	if err := json.Unmarshal([]byte(env.PayloadJSON), &payload); err != nil {
		return fmt.Errorf("decode request updated payload: %w", err)
	}

	changes := map[string]interface{}{}
	if payload.Status != "" {
		changes["status"] = payload.Status
	}
	if payload.Priority != "" {
		changes["priority"] = payload.Priority
	}
	if payload.AssignedAgent != "" {
		changes["assignee"] = payload.AssignedAgent
	}
	if l.decryptor != nil && payload.Payload != nil {
		decoded, err := l.decryptor.DecryptSupportRequestPayload(ctx, payload.Payload)
		if err != nil {
			l.logger.Error("failed to decrypt updated payload", "error", err)
		} else if decoded != nil {
			changes["subject"] = decoded.Subject
			changes["description"] = decoded.Description
		}
	}

	event := &TicketUpdatedEvent{
		TicketID:    payload.TicketID,
		UpdatedBy:   payload.UpdatedBy,
		Changes:     changes,
		BlockHeight: blockHeight,
		Timestamp:   time.Unix(payload.Timestamp, 0).UTC(),
		TxHash:      txHash,
	}

	return l.bridge.HandleTicketUpdated(ctx, event)
}

func (l *ChainEventListener) handleSupportStatusChanged(ctx context.Context, env SupportEventEnvelope, blockHeight int64, txHash string) error {
	var payload supporttypes.SupportStatusChangedEvent
	if err := json.Unmarshal([]byte(env.PayloadJSON), &payload); err != nil {
		return fmt.Errorf("decode status payload: %w", err)
	}
	event := &TicketUpdatedEvent{
		TicketID:    payload.TicketID,
		UpdatedBy:   payload.UpdatedBy,
		Changes:     map[string]interface{}{"status": payload.NewStatus},
		BlockHeight: blockHeight,
		Timestamp:   time.Unix(payload.Timestamp, 0).UTC(),
		TxHash:      txHash,
	}
	return l.bridge.HandleTicketUpdated(ctx, event)
}

func (l *ChainEventListener) handleSupportResponseAdded(ctx context.Context, env SupportEventEnvelope, blockHeight int64, txHash string) error {
	var payload supporttypes.SupportResponseAddedEvent
	if err := json.Unmarshal([]byte(env.PayloadJSON), &payload); err != nil {
		return fmt.Errorf("decode response payload: %w", err)
	}

	message := ""
	if l.decryptor != nil && payload.Payload != nil {
		decoded, err := l.decryptor.DecryptSupportResponsePayload(ctx, payload.Payload)
		if err != nil {
			l.logger.Error("failed to decrypt response payload", "error", err)
		} else if decoded != nil {
			message = decoded.Message
		}
	}

	event := &TicketResponseAddedEvent{
		TicketID:    payload.TicketID,
		Author:      payload.Author,
		IsAgent:     payload.IsAgent,
		Message:     message,
		BlockHeight: blockHeight,
		Timestamp:   time.Unix(payload.Timestamp, 0).UTC(),
		TxHash:      txHash,
	}
	return l.bridge.HandleTicketResponseAdded(ctx, event)
}

func defaultSupportEventQuery(eventTypes []string) string {
	if len(eventTypes) == 0 {
		return "tm.event='Tx' AND support_event.event_type='support_request_created'"
	}
	clauses := make([]string, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		clauses = append(clauses, fmt.Sprintf("support_event.event_type='%s'", eventType))
	}
	return fmt.Sprintf("tm.event='Tx' AND (%s)", strings.Join(clauses, " OR "))
}

// IChainEventListener is the interface for chain event listeners
type IChainEventListener interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Ensure ChainEventListener implements IChainEventListener
var _ IChainEventListener = (*ChainEventListener)(nil)
