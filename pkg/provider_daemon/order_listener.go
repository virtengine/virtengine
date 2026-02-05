package provider_daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// OrderListenerConfig configures chain order event listening.
type OrderListenerConfig struct {
	Enabled          bool
	ProviderAddress  string
	CometRPC         string
	CometWS          string
	SubscriberID     string
	EventBuffer      int
	EventQuery       string
	CheckpointFile   string
	ReplayOnStart    bool
	ReplayPageSize   int
	OperationTimeout time.Duration
}

// DefaultOrderListenerConfig returns default order listener config.
func DefaultOrderListenerConfig() OrderListenerConfig {
	return OrderListenerConfig{
		CometWS:          "/websocket",
		EventBuffer:      100,
		EventQuery:       "tm.event='Tx' AND marketplace_event.event_type='order_created'",
		ReplayOnStart:    true,
		ReplayPageSize:   100,
		OperationTimeout: 30 * time.Second,
	}
}

// OrderListener subscribes to order events and forwards them to the router.
type OrderListener struct {
	cfg             OrderListenerConfig
	router          *OrderRouter
	checkpointStore *EventCheckpointStore
	checkpoint      *EventCheckpointState
	rpcClient       *rpchttp.HTTP

	mu     sync.RWMutex
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewOrderListener creates a new order listener.
func NewOrderListener(cfg OrderListenerConfig, router *OrderRouter) (*OrderListener, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if cfg.ProviderAddress == "" {
		return nil, fmt.Errorf("provider address is required")
	}
	if cfg.CometRPC == "" {
		return nil, fmt.Errorf("comet RPC endpoint is required")
	}
	if cfg.SubscriberID == "" {
		cfg.SubscriberID = fmt.Sprintf("provider-order-listener-%d", time.Now().UnixNano())
	}
	if cfg.EventBuffer <= 0 {
		cfg.EventBuffer = 100
	}
	if cfg.EventQuery == "" {
		cfg.EventQuery = "tm.event='Tx' AND marketplace_event.event_type='order_created'"
	}
	if cfg.CheckpointFile == "" {
		cfg.CheckpointFile = "data/order_routing_checkpoint.json"
	}
	if cfg.ReplayPageSize <= 0 {
		cfg.ReplayPageSize = 100
	}
	if cfg.OperationTimeout == 0 {
		cfg.OperationTimeout = 30 * time.Second
	}

	checkpointStore, err := NewEventCheckpointStore(cfg.CheckpointFile)
	if err != nil {
		return nil, err
	}
	checkpoint, err := checkpointStore.Load(cfg.SubscriberID)
	if err != nil {
		return nil, err
	}

	return &OrderListener{
		cfg:             cfg,
		router:          router,
		checkpointStore: checkpointStore,
		checkpoint:      checkpoint,
		stopCh:          make(chan struct{}),
	}, nil
}

// Start begins listening for order events.
func (l *OrderListener) Start(ctx context.Context) error {
	if l == nil || !l.cfg.Enabled {
		return nil
	}
	rpc, err := rpchttp.New(l.cfg.CometRPC, l.cfg.CometWS)
	if err != nil {
		return fmt.Errorf("create rpc client: %w", err)
	}
	if err := rpc.Start(); err != nil {
		return fmt.Errorf("start rpc client: %w", err)
	}
	l.rpcClient = rpc

	if l.cfg.ReplayOnStart {
		if err := l.replayFromSequence(ctx, l.checkpoint.LastSequence+1, 0); err != nil {
			log.Printf("[order-listener] replay failed: %v", err)
		}
	}

	sub, err := rpc.Subscribe(ctx, l.cfg.SubscriberID, l.cfg.EventQuery, l.cfg.EventBuffer)
	if err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}

	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		l.eventLoop(ctx, sub)
	}()
	log.Printf("[order-listener] subscribed: %s", l.cfg.EventQuery)
	return nil
}

// Stop stops the listener.
func (l *OrderListener) Stop() {
	if l == nil {
		return
	}
	close(l.stopCh)
	l.wg.Wait()
	if l.rpcClient != nil {
		_ = l.rpcClient.Stop()
	}
}

func (l *OrderListener) eventLoop(ctx context.Context, sub <-chan coretypes.ResultEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-l.stopCh:
			return
		case msg, ok := <-sub:
			if !ok {
				log.Printf("[order-listener] subscription closed")
				return
			}
			if err := l.handleEvent(ctx, msg); err != nil {
				log.Printf("[order-listener] event handling failed: %v", err)
			}
		}
	}
}

func (l *OrderListener) handleEvent(ctx context.Context, msg coretypes.ResultEvent) error {
	data, ok := msg.Data.(tmtypes.EventDataTx)
	if !ok {
		return nil
	}

	envelopes, err := ExtractMarketplaceEvents(data.Result.Events)
	if err != nil {
		return err
	}
	for _, env := range envelopes {
		if env.EventType != string(marketplace.EventOrderCreated) {
			continue
		}
		if env.Sequence <= l.checkpoint.LastSequence {
			continue
		}
		if env.Sequence > l.checkpoint.LastSequence+1 {
			if err := l.replayFromSequence(ctx, l.checkpoint.LastSequence+1, env.Sequence-1); err != nil {
				log.Printf("[order-listener] replay gap failed: %v", err)
			}
		}
		orderEvent, err := parseOrderCreatedEvent(env.PayloadJSON)
		if err != nil {
			log.Printf("[order-listener] decode order event failed: %v", err)
			l.advanceCheckpoint(env.Sequence)
			continue
		}
		if !strings.EqualFold(orderEvent.ProviderAddress, l.cfg.ProviderAddress) {
			l.advanceCheckpoint(env.Sequence)
			continue
		}
		orderEvent.Sequence = env.Sequence
		if l.router != nil {
			if err := l.router.ProcessOrderCreated(ctx, orderEvent); err != nil {
				log.Printf("[order-listener] router error for %s: %v", orderEvent.OrderID, err)
			}
		}
		l.advanceCheckpoint(env.Sequence)
	}
	return nil
}

func (l *OrderListener) advanceCheckpoint(sequence uint64) {
	l.mu.Lock()
	l.checkpoint.LastSequence = sequence
	l.mu.Unlock()
	if err := l.checkpointStore.Save(l.checkpoint); err != nil {
		log.Printf("[order-listener] checkpoint save failed: %v", err)
	}
}

func (l *OrderListener) replayFromSequence(ctx context.Context, startSeq, endSeq uint64) error {
	if startSeq == 0 {
		startSeq = 1
	}
	if endSeq != 0 && endSeq < startSeq {
		return nil
	}
	if l.rpcClient == nil {
		return fmt.Errorf("rpc client not initialized")
	}

	page := 1
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		limit := l.cfg.ReplayPageSize
		result, err := l.rpcClient.TxSearch(ctx, l.cfg.EventQuery, true, &page, &limit, "asc")
		if err != nil {
			return err
		}

		for _, tx := range result.Txs {
			envelopes, err := ExtractMarketplaceEvents(tx.TxResult.Events)
			if err != nil {
				continue
			}
			for _, env := range envelopes {
				if env.EventType != string(marketplace.EventOrderCreated) {
					continue
				}
				if env.Sequence < startSeq {
					continue
				}
				if endSeq != 0 && env.Sequence > endSeq {
					continue
				}
				if env.Sequence <= l.checkpoint.LastSequence {
					continue
				}
				orderEvent, err := parseOrderCreatedEvent(env.PayloadJSON)
				if err != nil {
					continue
				}
				if !strings.EqualFold(orderEvent.ProviderAddress, l.cfg.ProviderAddress) {
					l.advanceCheckpoint(env.Sequence)
					continue
				}
				orderEvent.Sequence = env.Sequence
				if l.router != nil {
					if err := l.router.ProcessOrderCreated(ctx, orderEvent); err != nil {
						log.Printf("[order-listener] replay router error for %s: %v", orderEvent.OrderID, err)
					}
				}
				l.advanceCheckpoint(env.Sequence)
			}
		}

		if result.TotalCount <= len(result.Txs) || len(result.Txs) == 0 {
			break
		}
		page++
	}

	return nil
}

func parseOrderCreatedEvent(payload string) (marketplace.OrderCreatedEvent, error) {
	if payload == "" {
		return marketplace.OrderCreatedEvent{}, fmt.Errorf("empty payload")
	}
	var event marketplace.OrderCreatedEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		return marketplace.OrderCreatedEvent{}, err
	}
	return event, nil
}
