package provider_daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
)

// OrderListenerConfig configures chain order event listening.
type OrderListenerConfig struct {
	Enabled         bool
	ProviderAddress string
	CometRPC        string
	CometWS         string
	SubscriberID    string
	EventBuffer     int
	EventQuery      string
	CheckpointFile  string
	ReplayOnStart   bool
	ReplayLookback  int64
	MaxReplayPages  int
}

// DefaultOrderListenerConfig returns defaults for order listener.
func DefaultOrderListenerConfig() OrderListenerConfig {
	return OrderListenerConfig{
		CometWS:        "/websocket",
		EventBuffer:    100,
		ReplayOnStart:  true,
		ReplayLookback: 5000,
		MaxReplayPages: 10,
	}
}

// OrderListener listens for marketplace order events and routes them.
type OrderListener struct {
	cfg        OrderListenerConfig
	subscriber EventSubscriber
	router     *OrderRouter
	checkpoint *EventCheckpointState
	store      *EventCheckpointStore
}

// NewOrderListener creates a new order listener.
func NewOrderListener(cfg OrderListenerConfig, router *OrderRouter) (*OrderListener, error) {
	if cfg.CometRPC == "" {
		return nil, fmt.Errorf("comet RPC endpoint is required")
	}
	if cfg.SubscriberID == "" {
		cfg.SubscriberID = fmt.Sprintf("order-listener-%d", time.Now().UnixNano())
	}
	if cfg.EventQuery == "" {
		cfg.EventQuery = defaultOrderRoutingEventQuery()
	}
	if cfg.EventBuffer <= 0 {
		cfg.EventBuffer = 100
	}

	var store *EventCheckpointStore
	if cfg.CheckpointFile != "" {
		ck, err := NewEventCheckpointStore(cfg.CheckpointFile)
		if err != nil {
			return nil, err
		}
		store = ck
	}

	subCfg := DefaultEventSubscriberConfig()
	subCfg.CometRPC = cfg.CometRPC
	subCfg.CometWS = cfg.CometWS
	subCfg.EventBuffer = cfg.EventBuffer
	subCfg.SubscriberID = cfg.SubscriberID
	subCfg.CheckpointStore = store

	sub, err := NewCometEventSubscriber(subCfg)
	if err != nil {
		return nil, err
	}

	checkpoint := &EventCheckpointState{SubscriberID: cfg.SubscriberID}
	if store != nil {
		loaded, err := store.Load(cfg.SubscriberID)
		if err == nil {
			checkpoint = loaded
		}
	}

	return &OrderListener{
		cfg:        cfg,
		subscriber: sub,
		router:     router,
		checkpoint: checkpoint,
		store:      store,
	}, nil
}

// Start begins listening for order events.
func (l *OrderListener) Start(ctx context.Context) error {
	if l == nil || l.router == nil || !l.cfg.Enabled {
		return nil
	}

	if l.cfg.ReplayOnStart {
		if err := l.replayMissed(ctx); err != nil {
			log.Printf("[order-listener] replay failed: %v", err)
		}
	}

	events, err := l.subscriber.Subscribe(ctx, l.cfg.SubscriberID, l.cfg.EventQuery)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			_ = l.subscriber.Close()
			return ctx.Err()
		case event, ok := <-events:
			if !ok {
				return nil
			}
			if event.Sequence <= l.checkpoint.LastSequence {
				continue
			}
			if err := l.handleMarketplaceEvent(event); err != nil {
				log.Printf("[order-listener] event %s failed: %v", event.ID, err)
				continue
			}
			l.checkpoint.LastSequence = event.Sequence
			l.saveCheckpoint()
		}
	}
}

func (l *OrderListener) handleMarketplaceEvent(event MarketplaceEvent) error {
	switch event.Type {
	case EventTypeOrderCreated:
		req, ok := parseOrderRoutingRequest(event.Data, event.ID, event.Sequence)
		if !ok {
			return fmt.Errorf("invalid order_created payload")
		}
		if l.cfg.ProviderAddress != "" && req.ProviderAddress != "" &&
			!strings.EqualFold(req.ProviderAddress, l.cfg.ProviderAddress) {
			return nil
		}
		return l.router.Enqueue(req)
	case EventType("order_state_changed"):
		orderID, state := parseOrderStateChange(event.Data)
		if orderID != "" && state != "" {
			l.router.UpdateChainState(orderID, state)
		}
		return nil
	default:
		return nil
	}
}

func (l *OrderListener) saveCheckpoint() {
	if l.store == nil || l.checkpoint == nil {
		return
	}
	if err := l.store.Save(l.checkpoint); err != nil {
		log.Printf("[order-listener] checkpoint save failed: %v", err)
	}
}

func (l *OrderListener) replayMissed(ctx context.Context) error {
	if l.cfg.ReplayLookback <= 0 {
		return nil
	}

	rpc, err := rpchttp.New(l.cfg.CometRPC, l.cfg.CometWS)
	if err != nil {
		return err
	}
	if err := rpc.Start(); err != nil {
		return err
	}
	defer func() {
		if err := rpc.Stop(); err != nil {
			log.Printf("[order-listener] replay stop failed: %v", err)
		}
	}()

	page := 1
	perPage := 100
	pages := l.cfg.MaxReplayPages
	if pages <= 0 {
		pages = 10
	}

	query := l.cfg.EventQuery
	if l.cfg.ReplayLookback > 0 {
		status, err := rpc.Status(ctx)
		if err == nil && status != nil && status.SyncInfo.LatestBlockHeight > 0 {
			minHeight := status.SyncInfo.LatestBlockHeight - l.cfg.ReplayLookback
			if minHeight < 1 {
				minHeight = 1
			}
			query = fmt.Sprintf("%s AND tx.height>=%d", l.cfg.EventQuery, minHeight)
		}
	}

	for page <= pages {
		result, err := rpc.TxSearch(ctx, query, false, &page, &perPage, "asc")
		if err != nil {
			return err
		}
		if result == nil || len(result.Txs) == 0 {
			return nil
		}
		for _, tx := range result.Txs {
			if err := l.replayTx(ctx, tx); err != nil {
				log.Printf("[order-listener] replay tx failed: %v", err)
			}
		}
		if len(result.Txs) < perPage {
			break
		}
		page++
	}
	return nil
}

func (l *OrderListener) replayTx(ctx context.Context, tx *ctypes.ResultTx) error {
	if tx == nil {
		return nil
	}
	envelopes, err := ExtractMarketplaceEvents(tx.TxResult.Events)
	if err != nil {
		return err
	}
	for _, envelope := range envelopes {
		if envelope.Sequence <= l.checkpoint.LastSequence {
			continue
		}
		if envelope.EventType != "order_created" && envelope.EventType != "order_state_changed" {
			continue
		}
		payload := map[string]interface{}{}
		if envelope.PayloadJSON != "" {
			if err := json.Unmarshal([]byte(envelope.PayloadJSON), &payload); err != nil {
				continue
			}
		}
		event := MarketplaceEvent{
			Type:     EventType(envelope.EventType),
			ID:       envelope.EventID,
			Sequence: envelope.Sequence,
			Data:     payload,
		}
		if err := l.handleMarketplaceEvent(event); err != nil {
			log.Printf("[order-listener] replay event %s failed: %v", envelope.EventID, err)
			continue
		}
		l.checkpoint.LastSequence = envelope.Sequence
		l.saveCheckpoint()
	}
	return nil
}

func defaultOrderRoutingEventQuery() string {
	return "tm.event='Tx' AND (marketplace_event.event_type='order_created' OR marketplace_event.event_type='order_state_changed')"
}

func parseOrderRoutingRequest(data map[string]interface{}, eventID string, seq uint64) (OrderRoutingRequest, bool) {
	req := OrderRoutingRequest{
		EventID:  eventID,
		Sequence: seq,
	}
	if data == nil {
		return req, false
	}
	if v, ok := data["order_id"].(string); ok {
		req.OrderID = v
	}
	if v, ok := data["customer_address"].(string); ok {
		req.CustomerAddress = v
	}
	if v, ok := data["offering_id"].(string); ok {
		req.OfferingID = v
	}
	if v, ok := data["provider_address"].(string); ok {
		req.ProviderAddress = v
	}
	if v, ok := data["region"].(string); ok {
		req.Region = v
	}
	if v, ok := data["quantity"].(float64); ok {
		req.Quantity = uint32(v)
	}
	if v, ok := data["max_bid_price"].(float64); ok {
		req.MaxBidPrice = uint64(v)
	}
	if v, ok := data["expires_at"].(string); ok && v != "" {
		if ts, err := time.Parse(time.RFC3339, v); err == nil {
			req.ExpiresAt = &ts
		}
	}
	return req, req.OrderID != "" && req.OfferingID != ""
}

func parseOrderStateChange(data map[string]interface{}) (string, string) {
	if data == nil {
		return "", ""
	}
	orderID, _ := data["order_id"].(string)
	state, _ := data["new_state"].(string)
	return orderID, state
}
