package provider_daemon

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/waldur"
)

// OrderRoutingConfig configures order routing to Waldur.
type OrderRoutingConfig struct {
	Enabled          bool
	ProviderAddress  string
	WaldurBaseURL    string
	WaldurToken      string
	WaldurProjectID  string
	OrderCallbackURL string
	OfferingMap      map[string]string

	MaxRetries   int
	RetryBackoff time.Duration
	MaxBackoff   time.Duration
	QueueSize    int
	WorkerCount  int

	StateFile string

	OperationTimeout time.Duration

	AlertSink OrderRoutingAlertSink
}

// DefaultOrderRoutingConfig returns default routing configuration.
func DefaultOrderRoutingConfig() OrderRoutingConfig {
	return OrderRoutingConfig{
		Enabled:          false,
		MaxRetries:       5,
		RetryBackoff:     5 * time.Second,
		MaxBackoff:       5 * time.Minute,
		QueueSize:        256,
		WorkerCount:      4,
		StateFile:        "data/waldur_order_state.json",
		OperationTimeout: 45 * time.Second,
	}
}

// OrderRoutingAlertSink provides alerting for order routing failures.
type OrderRoutingAlertSink interface {
	CreateAlert(ctx context.Context, alertType AlertType, severity AlertSeverity, message string, allocationID, orderID string, details map[string]string) (*Alert, error)
}

// OrderRoutingRequest captures order data needed for routing.
type OrderRoutingRequest struct {
	OrderID         string
	CustomerAddress string
	OfferingID      string
	ProviderAddress string
	Region          string
	Quantity        uint32
	MaxBidPrice     uint64
	ExpiresAt       *time.Time
	EventID         string
	Sequence        uint64
}

// OrderRoutingTask wraps an order routing request for processing.
type OrderRoutingTask struct {
	Request    OrderRoutingRequest
	ReceivedAt time.Time
}

// WaldurOfferingResolver resolves a chain offering ID into a Waldur offering UUID.
type WaldurOfferingResolver interface {
	ResolveOfferingUUID(ctx context.Context, offeringID string) (string, error)
}

// MapOfferingResolver resolves offerings using a static map and backend lookup fallback.
type MapOfferingResolver struct {
	Map         map[string]string
	Marketplace *waldur.MarketplaceClient
}

// ResolveOfferingUUID resolves offering UUID from map or backend lookup.
func (r *MapOfferingResolver) ResolveOfferingUUID(ctx context.Context, offeringID string) (string, error) {
	if offeringID == "" {
		return "", fmt.Errorf("offering ID is required")
	}
	if r != nil && r.Map != nil {
		if uuid := r.Map[offeringID]; uuid != "" {
			return uuid, nil
		}
	}
	if r != nil && r.Marketplace != nil {
		offering, err := r.Marketplace.GetOfferingByBackendID(ctx, offeringID)
		if err == nil && offering != nil {
			return offering.UUID, nil
		}
		return "", err
	}
	return "", fmt.Errorf("no offering mapping for %s", offeringID)
}

// OrderRouter routes on-chain orders to Waldur.
type OrderRouter struct {
	cfg         OrderRoutingConfig
	store       *WaldurOrderStore
	state       *WaldurOrderState
	marketplace *waldur.MarketplaceClient
	resolver    WaldurOfferingResolver
	tasks       chan OrderRoutingTask
	stopCh      chan struct{}
	wg          sync.WaitGroup
	mu          sync.Mutex
}

// NewOrderRouter creates a new order router.
func NewOrderRouter(cfg OrderRoutingConfig, resolver WaldurOfferingResolver) (*OrderRouter, error) {
	if cfg.WaldurBaseURL == "" || cfg.WaldurToken == "" {
		return nil, fmt.Errorf("waldur base URL and token are required")
	}
	if cfg.WaldurProjectID == "" {
		return nil, fmt.Errorf("waldur project ID is required")
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 256
	}
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 4
	}
	if cfg.RetryBackoff <= 0 {
		cfg.RetryBackoff = 5 * time.Second
	}
	if cfg.MaxBackoff <= 0 {
		cfg.MaxBackoff = 5 * time.Minute
	}

	store, err := NewWaldurOrderStore(cfg.StateFile)
	if err != nil {
		return nil, err
	}
	state, err := store.Load()
	if err != nil {
		return nil, err
	}

	wcfg := waldur.DefaultConfig()
	wcfg.BaseURL = cfg.WaldurBaseURL
	wcfg.Token = cfg.WaldurToken
	client, err := waldur.NewClient(wcfg)
	if err != nil {
		return nil, err
	}
	marketplace := waldur.NewMarketplaceClient(client)

	if resolver == nil {
		resolver = &MapOfferingResolver{Map: cfg.OfferingMap, Marketplace: marketplace}
	}

	return &OrderRouter{
		cfg:         cfg,
		store:       store,
		state:       state,
		marketplace: marketplace,
		resolver:    resolver,
		tasks:       make(chan OrderRoutingTask, cfg.QueueSize),
		stopCh:      make(chan struct{}),
	}, nil
}

// Start begins processing routing tasks.
func (r *OrderRouter) Start(ctx context.Context) {
	if r == nil {
		return
	}
	for i := 0; i < r.cfg.WorkerCount; i++ {
		r.wg.Add(1)
		go r.worker(ctx, i)
	}
	r.wg.Add(1)
	go r.retryLoop(ctx)
}

// Stop stops workers and waits for completion.
func (r *OrderRouter) Stop() {
	if r == nil {
		return
	}
	close(r.stopCh)
	r.wg.Wait()
}

// Enqueue schedules an order routing request.
func (r *OrderRouter) Enqueue(req OrderRoutingRequest) error {
	if r == nil {
		return fmt.Errorf("order router not configured")
	}
	if req.OrderID == "" {
		return fmt.Errorf("order ID is required")
	}
	select {
	case r.tasks <- OrderRoutingTask{Request: req, ReceivedAt: time.Now().UTC()}:
		return nil
	default:
		return fmt.Errorf("order routing queue full")
	}
}

func (r *OrderRouter) worker(ctx context.Context, idx int) {
	defer r.wg.Done()
	for {
		select {
		case <-r.stopCh:
			return
		case <-ctx.Done():
			return
		case task := <-r.tasks:
			if err := r.processTask(ctx, task); err != nil {
				log.Printf("[order-router] worker=%d order=%s error=%v", idx, task.Request.OrderID, err)
			}
		}
	}
}

func (r *OrderRouter) retryLoop(ctx context.Context) {
	defer r.wg.Done()
	ticker := time.NewTicker(r.cfg.RetryBackoff)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-r.stopCh:
			return
		case <-ticker.C:
			now := time.Now().UTC()
			r.mu.Lock()
			for _, record := range r.state.Orders {
				if record == nil || record.DeadLettered || record.OrderID == "" {
					continue
				}
				if record.NextAttemptAt == nil || record.NextAttemptAt.After(now) {
					continue
				}
				req := OrderRoutingRequest{
					OrderID:         record.OrderID,
					CustomerAddress: record.CustomerAddress,
					OfferingID:      record.OfferingID,
					ProviderAddress: record.ProviderAddress,
					Region:          record.Attributes["region"],
				}
				_ = r.Enqueue(req)
			}
			r.mu.Unlock()
		}
	}
}

func (r *OrderRouter) processTask(ctx context.Context, task OrderRoutingTask) error {
	req := task.Request
	if req.ProviderAddress != "" && r.cfg.ProviderAddress != "" &&
		!equalAddresses(req.ProviderAddress, r.cfg.ProviderAddress) {
		return nil
	}

	r.mu.Lock()
	record := r.state.Get(req.OrderID)
	if record != nil && record.DeadLettered {
		r.mu.Unlock()
		return nil
	}
	r.mu.Unlock()

	opCtx, cancel := r.operationContext(ctx)
	defer cancel()

	if err := r.routeOrder(opCtx, req); err != nil {
		r.handleFailure(req, err)
		return err
	}

	return nil
}

func (r *OrderRouter) routeOrder(ctx context.Context, req OrderRoutingRequest) error {
	r.mu.Lock()
	record := r.state.Get(req.OrderID)
	if record == nil {
		record = &WaldurOrderRecord{
			OrderID:         req.OrderID,
			CustomerAddress: req.CustomerAddress,
			OfferingID:      req.OfferingID,
			ProviderAddress: req.ProviderAddress,
			Attributes:      map[string]string{},
		}
	}
	if record.Attributes == nil {
		record.Attributes = map[string]string{}
	}
	if req.Region != "" {
		record.Attributes["region"] = req.Region
	}
	r.state.Upsert(record)
	_ = r.store.Save(r.state)
	r.mu.Unlock()

	if record.WaldurOrderUUID != "" {
		return r.refreshExisting(ctx, req, record)
	}

	offeringUUID, err := r.resolver.ResolveOfferingUUID(ctx, req.OfferingID)
	if err != nil {
		return err
	}

	attrs := map[string]interface{}{
		"ve_order_id":         req.OrderID,
		"ve_customer_address": req.CustomerAddress,
		"ve_offering_id":      req.OfferingID,
		"ve_provider_address": req.ProviderAddress,
	}
	if req.Region != "" {
		attrs["region"] = req.Region
	}
	if req.MaxBidPrice > 0 {
		attrs["max_bid_price"] = fmt.Sprintf("%d", req.MaxBidPrice)
	}

	name := fmt.Sprintf("ve-order-%s", req.OrderID)
	description := fmt.Sprintf("VirtEngine order %s", req.OrderID)

	limits := map[string]int{}
	if req.Quantity > 0 {
		limits["instances"] = int(req.Quantity)
	}

	order, err := r.marketplace.CreateOrder(ctx, waldur.CreateOrderRequest{
		OfferingUUID:   offeringUUID,
		ProjectUUID:    r.cfg.WaldurProjectID,
		CallbackURL:    r.cfg.OrderCallbackURL,
		RequestComment: fmt.Sprintf("virtengine order %s", req.OrderID),
		Attributes:     attrs,
		Limits:         limits,
		Type:           "Create",
		Name:           name,
		Description:    description,
	})
	if err != nil {
		return err
	}

	r.mu.Lock()
	record = r.state.Get(req.OrderID)
	if record == nil {
		record = &WaldurOrderRecord{OrderID: req.OrderID}
	}
	record.WaldurOrderUUID = order.UUID
	record.WaldurResource = order.ResourceUUID
	record.WaldurState = order.State
	record.OfferingUUID = offeringUUID
	record.ProjectUUID = r.cfg.WaldurProjectID
	r.state.Upsert(record)
	_ = r.store.Save(r.state)
	r.mu.Unlock()

	if err := r.marketplace.ApproveOrderByProvider(ctx, order.UUID); err != nil {
		return err
	}
	if err := r.marketplace.SetOrderBackendID(ctx, order.UUID, req.OrderID); err != nil {
		return err
	}

	r.mu.Lock()
	record = r.state.Get(req.OrderID)
	if record != nil {
		record.LastError = ""
		record.NextAttemptAt = nil
		record.RetryCount = 0
		record.UpdatedAt = time.Now().UTC()
		r.state.Upsert(record)
		_ = r.store.Save(r.state)
	}
	r.mu.Unlock()

	return nil
}

func (r *OrderRouter) refreshExisting(ctx context.Context, req OrderRoutingRequest, record *WaldurOrderRecord) error {
	if record == nil || record.WaldurOrderUUID == "" {
		return errors.New("missing waldur order UUID")
	}

	order, err := r.marketplace.GetOrder(ctx, record.WaldurOrderUUID)
	if err != nil {
		return err
	}

	if err := r.marketplace.ApproveOrderByProvider(ctx, record.WaldurOrderUUID); err != nil && !errors.Is(err, waldur.ErrConflict) {
		return err
	}
	if err := r.marketplace.SetOrderBackendID(ctx, record.WaldurOrderUUID, req.OrderID); err != nil {
		return err
	}

	r.mu.Lock()
	record.WaldurState = order.State
	record.WaldurResource = order.ResourceUUID
	record.LastError = ""
	record.RetryCount = 0
	record.NextAttemptAt = nil
	record.UpdatedAt = time.Now().UTC()
	r.state.Upsert(record)
	_ = r.store.Save(r.state)
	r.mu.Unlock()
	return nil
}

func (r *OrderRouter) handleFailure(req OrderRoutingRequest, err error) {
	r.mu.Lock()
	record := r.state.MarkFailed(req.OrderID, err.Error())
	backoff := r.retryDelay(record.RetryCount)
	next := time.Now().UTC().Add(backoff)
	record.NextAttemptAt = &next
	r.state.Upsert(record)
	_ = r.store.Save(r.state)
	r.mu.Unlock()

	if record.RetryCount > r.cfg.MaxRetries {
		r.mu.Lock()
		r.state.DeadLetter(record, "max retries exceeded")
		_ = r.store.Save(r.state)
		r.mu.Unlock()

		if r.cfg.AlertSink != nil {
			details := map[string]string{
				"error":       err.Error(),
				"offering_id": req.OfferingID,
				"provider":    req.ProviderAddress,
			}
			_, _ = r.cfg.AlertSink.CreateAlert(context.Background(), AlertTypeServiceDegraded, AlertSeverityError,
				"order routing dead-lettered", "", req.OrderID, details)
		}
	}
}

func (r *OrderRouter) retryDelay(retryCount int) time.Duration {
	if retryCount <= 0 {
		return r.cfg.RetryBackoff
	}
	backoff := float64(r.cfg.RetryBackoff) * math.Pow(2, float64(retryCount-1))
	if max := float64(r.cfg.MaxBackoff); backoff > max {
		backoff = max
	}
	return time.Duration(backoff)
}

func (r *OrderRouter) operationContext(parent context.Context) (context.Context, context.CancelFunc) {
	timeout := r.cfg.OperationTimeout
	if timeout <= 0 {
		timeout = 45 * time.Second
	}
	return context.WithTimeout(parent, timeout)
}

func equalAddresses(a, b string) bool {
	return a == b || (a != "" && b != "" && a == b)
}

// UpdateChainState records the latest chain-side order state.
func (r *OrderRouter) UpdateChainState(orderID, state string) {
	if r == nil || orderID == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	record := r.state.Get(orderID)
	if record == nil {
		record = &WaldurOrderRecord{OrderID: orderID}
	}
	record.ChainState = state
	record.UpdatedAt = time.Now().UTC()
	r.state.Upsert(record)
	_ = r.store.Save(r.state)
}

// Store returns the underlying order store.
func (r *OrderRouter) Store() *WaldurOrderStore {
	if r == nil {
		return nil
	}
	return r.store
}
