package provider_daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"

	verrors "github.com/virtengine/virtengine/pkg/errors"
)

// CometEventSubscriber implements EventSubscriber using CometBFT WebSocket.
type CometEventSubscriber struct {
	cfg             EventSubscriberConfig
	rpcClient       *rpchttp.HTTP
	checkpointStore *EventCheckpointStore
	checkpoint      *EventCheckpointState

	mu             sync.RWMutex
	connected      bool
	lastEventTime  time.Time
	reconnectCount int
	lastError      string
	usingFallback  bool
	closed         bool

	cancelFn context.CancelFunc
	wg       sync.WaitGroup
}

// NewCometEventSubscriber creates a new CometBFT event subscriber.
func NewCometEventSubscriber(cfg EventSubscriberConfig) (*CometEventSubscriber, error) {
	if cfg.CometRPC == "" {
		return nil, fmt.Errorf("comet RPC endpoint is required")
	}
	if cfg.SubscriberID == "" {
		cfg.SubscriberID = fmt.Sprintf("provider-stream-%d", time.Now().UnixNano())
	}
	if cfg.CometWS == "" {
		cfg.CometWS = "/websocket"
	}
	if cfg.EventBuffer <= 0 {
		cfg.EventBuffer = 100
	}
	if cfg.ReconnectDelay <= 0 {
		cfg.ReconnectDelay = time.Second
	}
	if cfg.MaxReconnectDelay <= 0 {
		cfg.MaxReconnectDelay = time.Minute
	}
	if cfg.ReconnectBackoffFactor <= 0 {
		cfg.ReconnectBackoffFactor = 2.0
	}

	sub := &CometEventSubscriber{
		cfg:             cfg,
		checkpointStore: cfg.CheckpointStore,
	}

	return sub, nil
}

// connect establishes the WebSocket connection.
func (s *CometEventSubscriber) connect(ctx context.Context) error {
	rpc, err := rpchttp.New(s.cfg.CometRPC, s.cfg.CometWS)
	if err != nil {
		return fmt.Errorf("create rpc client: %w", err)
	}

	if err := rpc.Start(); err != nil {
		return fmt.Errorf("start rpc client: %w", err)
	}

	s.mu.Lock()
	s.rpcClient = rpc
	s.connected = true
	s.lastError = ""
	s.mu.Unlock()

	// Load checkpoint if available
	if s.checkpointStore != nil {
		checkpoint, err := s.checkpointStore.Load(s.cfg.SubscriberID)
		if err != nil {
			log.Printf("[stream-client] failed to load checkpoint: %v", err)
		} else {
			s.checkpoint = checkpoint
		}
	}

	return nil
}

// reconnect handles reconnection with exponential backoff.
func (s *CometEventSubscriber) reconnect(ctx context.Context) error {
	delay := s.cfg.ReconnectDelay

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		s.mu.Lock()
		if s.closed {
			s.mu.Unlock()
			return fmt.Errorf("subscriber closed")
		}
		s.reconnectCount++
		count := s.reconnectCount
		s.mu.Unlock()

		log.Printf("[stream-client] reconnection attempt %d after %v", count, delay)

		// Wait before reconnecting
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}

		if err := s.connect(ctx); err != nil {
			s.mu.Lock()
			s.lastError = err.Error()
			s.mu.Unlock()
			log.Printf("[stream-client] reconnect failed: %v", err)

			// Exponential backoff
			delay = time.Duration(float64(delay) * s.cfg.ReconnectBackoffFactor)
			if delay > s.cfg.MaxReconnectDelay {
				delay = s.cfg.MaxReconnectDelay
			}
			continue
		}

		log.Printf("[stream-client] reconnected successfully")
		return nil
	}
}

// Subscribe starts listening for events matching the given query.
func (s *CometEventSubscriber) Subscribe(ctx context.Context, subscriberID string, query string) (<-chan MarketplaceEvent, error) {
	if err := s.connect(ctx); err != nil {
		return nil, err
	}

	eventCh := make(chan MarketplaceEvent, s.cfg.EventBuffer)

	subCtx, cancel := context.WithCancel(ctx)
	s.cancelFn = cancel

	s.wg.Add(1)
	verrors.SafeGo("provider-stream:subscribe", func() {
		defer s.wg.Done()
		defer close(eventCh)
		s.subscriptionLoop(subCtx, subscriberID, query, eventCh)
	})

	return eventCh, nil
}

// subscriptionLoop manages the subscription with automatic reconnection.
func (s *CometEventSubscriber) subscriptionLoop(ctx context.Context, subscriberID, query string, eventCh chan<- MarketplaceEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		s.mu.RLock()
		rpc := s.rpcClient
		s.mu.RUnlock()

		if rpc == nil {
			if err := s.reconnect(ctx); err != nil {
				log.Printf("[stream-client] reconnect failed permanently: %v", err)
				return
			}
			s.mu.RLock()
			rpc = s.rpcClient
			s.mu.RUnlock()
		}

		sub, err := rpc.Subscribe(ctx, subscriberID, query, s.cfg.EventBuffer)
		if err != nil {
			log.Printf("[stream-client] subscribe failed: %v", err)
			s.mu.Lock()
			s.connected = false
			s.lastError = err.Error()
			s.rpcClient = nil
			s.mu.Unlock()
			continue
		}

		log.Printf("[stream-client] subscribed to: %s", query)
		s.processEvents(ctx, sub, eventCh)

		// If we exit processEvents, connection was lost
		s.mu.Lock()
		s.connected = false
		s.rpcClient = nil
		s.mu.Unlock()
	}
}

// processEvents reads events from the subscription channel.
func (s *CometEventSubscriber) processEvents(ctx context.Context, sub <-chan ctypes.ResultEvent, eventCh chan<- MarketplaceEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-sub:
			if !ok {
				log.Printf("[stream-client] subscription channel closed")
				return
			}

			event, err := s.parseEvent(msg)
			if err != nil {
				log.Printf("[stream-client] parse event error: %v", err)
				continue
			}

			// Skip if already processed (checkpoint-based dedup)
			s.mu.RLock()
			checkpoint := s.checkpoint
			s.mu.RUnlock()

			if checkpoint != nil && event.Sequence <= checkpoint.LastSequence {
				continue
			}

			select {
			case eventCh <- event:
				s.mu.Lock()
				s.lastEventTime = time.Now()
				s.mu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}
}

// parseEvent converts a CometBFT event to MarketplaceEvent.
func (s *CometEventSubscriber) parseEvent(msg ctypes.ResultEvent) (MarketplaceEvent, error) {
	event := MarketplaceEvent{
		Timestamp: time.Now().UTC(),
		Data:      make(map[string]interface{}),
	}

	data, ok := msg.Data.(tmtypes.EventDataTx)
	if !ok {
		return event, fmt.Errorf("unexpected event data type: %T", msg.Data)
	}

	event.BlockHeight = data.Height

	// Extract marketplace events from ABCI events
	envelopes, err := ExtractMarketplaceEvents(data.Result.Events)
	if err != nil {
		return event, err
	}

	if len(envelopes) == 0 {
		return event, fmt.Errorf("no marketplace events in transaction")
	}

	// Use the first envelope (most transactions have one event)
	env := envelopes[0]
	event.Type = eventTypeFromString(env.EventType)
	event.ID = env.EventID
	event.Sequence = env.Sequence

	if env.PayloadJSON != "" {
		if err := json.Unmarshal([]byte(env.PayloadJSON), &event.Data); err != nil {
			log.Printf("[stream-client] failed to parse payload: %v", err)
		}
	}

	return event, nil
}

// eventTypeFromString maps string event types to EventType.
func eventTypeFromString(s string) EventType {
	switch s {
	case "order_created":
		return EventTypeOrderCreated
	case "order_closed":
		return EventTypeOrderClosed
	case "bid_created":
		return EventTypeBidCreated
	case "bid_accepted":
		return EventTypeBidAccepted
	case "lease_created":
		return EventTypeLeaseCreated
	case "lease_closed":
		return EventTypeLeaseClosed
	case "config_updated":
		return EventTypeConfigUpdated
	default:
		return EventType(s)
	}
}

// SubscribeOrders subscribes to order-related events for the given provider.
func (s *CometEventSubscriber) SubscribeOrders(ctx context.Context, providerAddress string) (<-chan OrderEvent, error) {
	query := buildOrderQuery(providerAddress)
	baseCh, err := s.Subscribe(ctx, s.cfg.SubscriberID+"-orders", query)
	if err != nil {
		return nil, err
	}

	orderCh := make(chan OrderEvent, s.cfg.EventBuffer)

	s.wg.Add(1)
	verrors.SafeGo("provider-stream:order-filter", func() {
		defer s.wg.Done()
		defer close(orderCh)

		for event := range baseCh {
			if event.Type != EventTypeOrderCreated && event.Type != EventTypeOrderClosed {
				continue
			}

			orderEvent := OrderEvent{
				MarketplaceEvent: event,
			}

			// Parse order from event data
			if order, err := parseOrderFromEvent(event); err == nil {
				orderEvent.Order = order
			}

			select {
			case orderCh <- orderEvent:
			case <-ctx.Done():
				return
			}
		}
	})

	return orderCh, nil
}

// SubscribeConfig subscribes to provider config changes.
func (s *CometEventSubscriber) SubscribeConfig(ctx context.Context, providerAddress string) (<-chan ConfigEvent, error) {
	query := buildConfigQuery(providerAddress)
	baseCh, err := s.Subscribe(ctx, s.cfg.SubscriberID+"-config", query)
	if err != nil {
		return nil, err
	}

	configCh := make(chan ConfigEvent, s.cfg.EventBuffer)

	s.wg.Add(1)
	verrors.SafeGo("provider-stream:config-filter", func() {
		defer s.wg.Done()
		defer close(configCh)

		for event := range baseCh {
			if event.Type != EventTypeConfigUpdated {
				continue
			}

			configEvent := ConfigEvent{
				MarketplaceEvent: event,
			}

			// Parse config from event data
			if config, err := parseConfigFromEvent(event); err == nil {
				configEvent.Config = config
			}

			select {
			case configCh <- configEvent:
			case <-ctx.Done():
				return
			}
		}
	})

	return configCh, nil
}

// LastCheckpoint returns the last processed event sequence.
func (s *CometEventSubscriber) LastCheckpoint() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.checkpoint != nil {
		return s.checkpoint.LastSequence
	}
	return 0
}

// SetCheckpoint sets the checkpoint for replay on reconnect.
func (s *CometEventSubscriber) SetCheckpoint(sequence uint64) {
	s.mu.Lock()
	if s.checkpoint == nil {
		s.checkpoint = &EventCheckpointState{
			SubscriberID: s.cfg.SubscriberID,
		}
	}
	s.checkpoint.LastSequence = sequence
	s.mu.Unlock()

	if s.checkpointStore != nil {
		if err := s.checkpointStore.Save(s.checkpoint); err != nil {
			log.Printf("[stream-client] checkpoint save failed: %v", err)
		}
	}
}

// Close terminates all subscriptions.
func (s *CometEventSubscriber) Close() error {
	s.mu.Lock()
	s.closed = true
	if s.cancelFn != nil {
		s.cancelFn()
	}
	rpc := s.rpcClient
	s.mu.Unlock()

	s.wg.Wait()

	if rpc != nil {
		return rpc.Stop()
	}
	return nil
}

// Status returns the current connection status.
func (s *CometEventSubscriber) Status() SubscriberStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := SubscriberStatus{
		Connected:      s.connected,
		LastEventTime:  s.lastEventTime,
		ReconnectCount: s.reconnectCount,
		LastError:      s.lastError,
		UsingFallback:  s.usingFallback,
	}
	if s.checkpoint != nil {
		status.LastCheckpoint = s.checkpoint.LastSequence
	}
	return status
}

// SetFallbackMode marks the subscriber as using fallback polling.
func (s *CometEventSubscriber) SetFallbackMode(fallback bool) {
	s.mu.Lock()
	s.usingFallback = fallback
	s.mu.Unlock()
}

// buildOrderQuery constructs a CometBFT query for order events.
func buildOrderQuery(providerAddress string) string {
	return fmt.Sprintf(
		"tm.event='Tx' AND (marketplace_event.event_type='order_created' OR marketplace_event.event_type='order_closed')",
	)
}

// buildConfigQuery constructs a CometBFT query for config events.
func buildConfigQuery(providerAddress string) string {
	return fmt.Sprintf(
		"tm.event='Tx' AND provider_config.provider_address='%s'",
		providerAddress,
	)
}

// parseOrderFromEvent extracts an Order from event data.
func parseOrderFromEvent(event MarketplaceEvent) (Order, error) {
	order := Order{}

	if id, ok := event.Data["order_id"].(string); ok {
		order.OrderID = id
	}
	if addr, ok := event.Data["customer_address"].(string); ok {
		order.CustomerAddress = addr
	}
	if t, ok := event.Data["offering_type"].(string); ok {
		order.OfferingType = t
	}
	if region, ok := event.Data["region"].(string); ok {
		order.Region = region
	}
	if price, ok := event.Data["max_price"].(string); ok {
		order.MaxPrice = price
	}
	if currency, ok := event.Data["currency"].(string); ok {
		order.Currency = currency
	}

	// Parse requirements
	if reqs, ok := event.Data["requirements"].(map[string]interface{}); ok {
		if cpu, ok := reqs["cpu_cores"].(float64); ok {
			order.Requirements.CPUCores = int64(cpu)
		}
		if mem, ok := reqs["memory_gb"].(float64); ok {
			order.Requirements.MemoryGB = int64(mem)
		}
		if storage, ok := reqs["storage_gb"].(float64); ok {
			order.Requirements.StorageGB = int64(storage)
		}
		if gpus, ok := reqs["gpus"].(float64); ok {
			order.Requirements.GPUs = int64(gpus)
		}
		if gpuType, ok := reqs["gpu_type"].(string); ok {
			order.Requirements.GPUType = gpuType
		}
	}

	return order, nil
}

// parseConfigFromEvent extracts a ProviderConfig from event data.
func parseConfigFromEvent(event MarketplaceEvent) (*ProviderConfig, error) {
	config := &ProviderConfig{}

	if addr, ok := event.Data["provider_address"].(string); ok {
		config.ProviderAddress = addr
	}
	if active, ok := event.Data["active"].(bool); ok {
		config.Active = active
	}
	if version, ok := event.Data["version"].(float64); ok {
		config.Version = uint64(version)
	}

	// Parse pricing
	if pricing, ok := event.Data["pricing"].(map[string]interface{}); ok {
		if cpu, ok := pricing["cpu_price_per_core"].(string); ok {
			config.Pricing.CPUPricePerCore = cpu
		}
		if mem, ok := pricing["memory_price_per_gb"].(string); ok {
			config.Pricing.MemoryPricePerGB = mem
		}
		if storage, ok := pricing["storage_price_per_gb"].(string); ok {
			config.Pricing.StoragePricePerGB = storage
		}
		if network, ok := pricing["network_price_per_gb"].(string); ok {
			config.Pricing.NetworkPricePerGB = network
		}
		if gpu, ok := pricing["gpu_price_per_hour"].(string); ok {
			config.Pricing.GPUPricePerHour = gpu
		}
		if min, ok := pricing["min_bid_price"].(string); ok {
			config.Pricing.MinBidPrice = min
		}
		if markup, ok := pricing["bid_markup_percent"].(float64); ok {
			config.Pricing.BidMarkupPercent = markup
		}
		if currency, ok := pricing["currency"].(string); ok {
			config.Pricing.Currency = currency
		}
	}

	// Parse capacity
	if capacity, ok := event.Data["capacity"].(map[string]interface{}); ok {
		if cpu, ok := capacity["total_cpu_cores"].(float64); ok {
			config.Capacity.TotalCPUCores = int64(cpu)
		}
		if mem, ok := capacity["total_memory_gb"].(float64); ok {
			config.Capacity.TotalMemoryGB = int64(mem)
		}
		if storage, ok := capacity["total_storage_gb"].(float64); ok {
			config.Capacity.TotalStorageGB = int64(storage)
		}
		if gpus, ok := capacity["total_gpus"].(float64); ok {
			config.Capacity.TotalGPUs = int64(gpus)
		}
	}

	// Parse supported offerings and regions
	if offerings, ok := event.Data["supported_offerings"].([]interface{}); ok {
		for _, o := range offerings {
			if s, ok := o.(string); ok {
				config.SupportedOfferings = append(config.SupportedOfferings, s)
			}
		}
	}
	if regions, ok := event.Data["regions"].([]interface{}); ok {
		for _, r := range regions {
			if s, ok := r.(string); ok {
				config.Regions = append(config.Regions, s)
			}
		}
	}

	return config, nil
}
