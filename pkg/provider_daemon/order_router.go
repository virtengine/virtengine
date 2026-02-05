package provider_daemon

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/waldur"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// OrderRouterConfig configures order routing from chain to Waldur.
type OrderRouterConfig struct {
	Enabled            bool
	ProviderAddress    string
	WaldurBaseURL      string
	WaldurToken        string
	WaldurProjectUUID  string
	WaldurOfferingMap  map[string]string
	OrderCallbackURL   string
	StateFile          string
	OperationTimeout   time.Duration
	RetryInterval      time.Duration
	MaxRetries         int
	RetryBackoff       time.Duration
	MaxRetryBackoff    time.Duration
	ReplayOnStartup    bool
	DeadLetterAlerting bool
}

// DefaultOrderRouterConfig returns defaults for order routing.
func DefaultOrderRouterConfig() OrderRouterConfig {
	return OrderRouterConfig{
		OperationTimeout:   45 * time.Second,
		RetryInterval:      30 * time.Second,
		MaxRetries:         5,
		RetryBackoff:       5 * time.Second,
		MaxRetryBackoff:    5 * time.Minute,
		DeadLetterAlerting: true,
	}
}

// OrderRoutingAlert contains manual intervention alert data.
type OrderRoutingAlert struct {
	OrderID      string
	OfferingID   string
	Provider     string
	Attempts     int
	LastError    string
	DeadLettered bool
	RecordedAt   time.Time
}

// OrderRoutingAlertManager handles manual intervention alerts.
type OrderRoutingAlertManager interface {
	Alert(ctx context.Context, alert OrderRoutingAlert) error
}

// DefaultOrderRoutingAlertManager logs alerts.
type DefaultOrderRoutingAlertManager struct{}

// Alert logs the alert for manual intervention.
func (m *DefaultOrderRoutingAlertManager) Alert(_ context.Context, alert OrderRoutingAlert) error {
	log.Printf("[order-router] ALERT order=%s offering=%s provider=%s attempts=%d dead_lettered=%t error=%s",
		alert.OrderID, alert.OfferingID, alert.Provider, alert.Attempts, alert.DeadLettered, alert.LastError)
	return nil
}

// WaldurOrderAPI captures Waldur order operations.
type WaldurOrderAPI interface {
	CreateOrder(ctx context.Context, req waldur.CreateOrderRequest) (*waldur.Order, error)
	ApproveOrderByProvider(ctx context.Context, orderUUID string) error
	SetOrderBackendID(ctx context.Context, orderUUID string, backendID string) error
	GetOrder(ctx context.Context, orderUUID string) (*waldur.Order, error)
	GetOfferingByBackendID(ctx context.Context, backendID string) (*waldur.Offering, error)
}

// OrderRouter routes chain orders to Waldur.
type OrderRouter struct {
	cfg        OrderRouterConfig
	client     WaldurOrderAPI
	stateStore *OrderRoutingStateStore
	state      *OrderRoutingState
	alerts     OrderRoutingAlertManager

	mu     sync.RWMutex
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewOrderRouter creates an order router with Waldur client and state store.
func NewOrderRouter(cfg OrderRouterConfig) (*OrderRouter, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if cfg.ProviderAddress == "" {
		return nil, fmt.Errorf("provider address is required")
	}
	if cfg.WaldurBaseURL == "" {
		return nil, fmt.Errorf("waldur base URL is required")
	}
	if cfg.WaldurToken == "" {
		return nil, fmt.Errorf("waldur token is required")
	}
	if cfg.WaldurProjectUUID == "" {
		return nil, fmt.Errorf("waldur project UUID is required")
	}
	if cfg.StateFile == "" {
		cfg.StateFile = "data/waldur_order_state.json"
	}
	if cfg.OperationTimeout == 0 {
		cfg.OperationTimeout = 45 * time.Second
	}
	if cfg.RetryInterval == 0 {
		cfg.RetryInterval = 30 * time.Second
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 5
	}
	if cfg.RetryBackoff == 0 {
		cfg.RetryBackoff = 5 * time.Second
	}
	if cfg.MaxRetryBackoff == 0 {
		cfg.MaxRetryBackoff = 5 * time.Minute
	}

	waldurCfg := waldur.DefaultConfig()
	waldurCfg.BaseURL = cfg.WaldurBaseURL
	waldurCfg.Token = cfg.WaldurToken
	waldurClient, err := waldur.NewClient(waldurCfg)
	if err != nil {
		return nil, err
	}
	marketplaceClient := waldur.NewMarketplaceClient(waldurClient)
	return NewOrderRouterWithClient(cfg, marketplaceClient, nil, nil)
}

// NewOrderRouterWithClient allows injecting a Waldur client and state store (for tests).
func NewOrderRouterWithClient(
	cfg OrderRouterConfig,
	client WaldurOrderAPI,
	store *OrderRoutingStateStore,
	alerts OrderRoutingAlertManager,
) (*OrderRouter, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if client == nil {
		return nil, fmt.Errorf("waldur client is required")
	}
	if store == nil {
		store = NewOrderRoutingStateStore(cfg.StateFile)
	}
	state, err := store.Load()
	if err != nil {
		return nil, err
	}
	if alerts == nil {
		alerts = &DefaultOrderRoutingAlertManager{}
	}
	return &OrderRouter{
		cfg:        cfg,
		client:     client,
		stateStore: store,
		state:      state,
		alerts:     alerts,
		stopCh:     make(chan struct{}),
	}, nil
}

// Start launches background retry processing.
func (r *OrderRouter) Start(ctx context.Context) error {
	if r == nil || !r.cfg.Enabled {
		return nil
	}

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		r.retryLoop(ctx)
	}()
	log.Printf("[order-router] started")
	return nil
}

// Stop stops background processing.
func (r *OrderRouter) Stop() {
	if r == nil {
		return
	}
	close(r.stopCh)
	r.wg.Wait()
}

// ProcessOrderCreated routes a chain order to Waldur.
func (r *OrderRouter) ProcessOrderCreated(ctx context.Context, event marketplace.OrderCreatedEvent) error {
	if !r.cfg.Enabled {
		return nil
	}
	if !strings.EqualFold(event.ProviderAddress, r.cfg.ProviderAddress) {
		r.markSkipped(event.OrderID)
		return nil
	}
	if event.OrderID == "" || event.OfferingID == "" {
		return fmt.Errorf("invalid order event: missing order or offering id")
	}

	record := r.getOrCreateRecord(event)
	if record.DeadLettered {
		return fmt.Errorf("order %s is dead-lettered", event.OrderID)
	}
	if record.LastSequence >= event.Sequence && record.WaldurOrderUUID != "" {
		r.markSkipped(event.OrderID)
		return nil
	}

	err := r.createOrUpdateWaldurOrder(ctx, &record, event)
	if err != nil {
		return r.handleFailure(ctx, &record, err)
	}

	r.markProcessed(&record, event)
	return nil
}

func (r *OrderRouter) createOrUpdateWaldurOrder(ctx context.Context, record *OrderRoutingRecord, event marketplace.OrderCreatedEvent) error {
	opCtx, cancel := r.operationContext(ctx)
	defer cancel()

	waldurOfferingUUID, err := r.resolveOfferingUUID(opCtx, event.OfferingID)
	if err != nil {
		return err
	}

	record.WaldurOfferingUUID = waldurOfferingUUID

	if record.WaldurOrderUUID != "" {
		order, err := r.client.GetOrder(opCtx, record.WaldurOrderUUID)
		if err == nil && order != nil {
			if order.ResourceUUID != "" && record.WaldurResourceUUID == "" {
				record.WaldurResourceUUID = order.ResourceUUID
			}
			return nil
		}
	}

	attributes := map[string]interface{}{
		"order_id":         event.OrderID,
		"provider_address": event.ProviderAddress,
		"customer_address": event.CustomerAddress,
		"offering_id":      event.OfferingID,
		"requested_qty":    event.Quantity,
		"max_bid_price":    event.MaxBidPrice,
	}
	if event.Region != "" {
		attributes["region"] = event.Region
	}
	if event.ExpiresAt != nil {
		attributes["expires_at"] = event.ExpiresAt.UTC().Format(time.RFC3339)
	}

	orderName := fmt.Sprintf("ve-order-%s", sanitizeOrderID(event.OrderID))
	req := waldur.CreateOrderRequest{
		OfferingUUID:   waldurOfferingUUID,
		ProjectUUID:    r.cfg.WaldurProjectUUID,
		CallbackURL:    r.cfg.OrderCallbackURL,
		RequestComment: fmt.Sprintf("virtengine order %s", event.OrderID),
		Attributes:     attributes,
		Name:           orderName,
		Description:    fmt.Sprintf("VirtEngine order %s", event.OrderID),
	}

	order, err := r.client.CreateOrder(opCtx, req)
	if err != nil {
		return err
	}
	record.WaldurOrderUUID = order.UUID
	record.WaldurResourceUUID = order.ResourceUUID

	if err := r.client.ApproveOrderByProvider(opCtx, order.UUID); err != nil {
		if !errors.Is(err, waldur.ErrConflict) {
			return err
		}
	}

	if err := r.client.SetOrderBackendID(opCtx, order.UUID, event.OrderID); err != nil {
		if !errors.Is(err, waldur.ErrConflict) {
			return err
		}
	}

	return nil
}

func (r *OrderRouter) resolveOfferingUUID(ctx context.Context, offeringID string) (string, error) {
	if offeringID == "" {
		return "", fmt.Errorf("offering ID is required")
	}
	if r.cfg.WaldurOfferingMap != nil {
		if mapped := r.cfg.WaldurOfferingMap[offeringID]; mapped != "" {
			return mapped, nil
		}
	}
	offering, err := r.client.GetOfferingByBackendID(ctx, offeringID)
	if err != nil {
		return "", fmt.Errorf("resolve waldur offering for %s: %w", offeringID, err)
	}
	if offering == nil || offering.UUID == "" {
		return "", fmt.Errorf("waldur offering not found for %s", offeringID)
	}
	return offering.UUID, nil
}

func (r *OrderRouter) operationContext(parent context.Context) (context.Context, context.CancelFunc) {
	timeout := r.cfg.OperationTimeout
	if timeout <= 0 {
		timeout = 45 * time.Second
	}
	return context.WithTimeout(parent, timeout)
}

func (r *OrderRouter) getOrCreateRecord(event marketplace.OrderCreatedEvent) OrderRoutingRecord {
	r.mu.Lock()
	defer r.mu.Unlock()

	record := r.state.Records[event.OrderID]
	if record == nil {
		record = &OrderRoutingRecord{
			OrderID:         event.OrderID,
			OfferingID:      event.OfferingID,
			ProviderAddress: event.ProviderAddress,
			CustomerAddress: event.CustomerAddress,
			CreatedAt:       time.Now().UTC(),
		}
		r.state.Records[event.OrderID] = record
	}
	return *record
}

func (r *OrderRouter) markProcessed(record *OrderRoutingRecord, event marketplace.OrderCreatedEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	record.LastSequence = event.Sequence
	record.LastState = "created"
	record.RetryCount = 0
	record.LastError = ""
	record.UpdatedAt = now
	record.LastAttemptAt = now

	r.state.Metrics.Processed++
	r.state.Metrics.LastSequence = event.Sequence
	r.state.Metrics.LastProcessed = event.OrderID

	r.state.Records[record.OrderID] = record
	if err := r.stateStore.Save(r.state); err != nil {
		log.Printf("[order-router] failed to save state: %v", err)
	}
}

func (r *OrderRouter) markSkipped(orderID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state.Metrics.Skipped++
	r.state.Metrics.LastProcessed = orderID
	_ = r.stateStore.Save(r.state)
}

func (r *OrderRouter) handleFailure(ctx context.Context, record *OrderRoutingRecord, err error) error {
	if err == nil {
		return nil
	}

	retryable := isRetryableOrderError(err)

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	record.LastError = err.Error()
	record.UpdatedAt = now
	record.LastAttemptAt = now
	if retryable {
		record.RetryCount++
		r.state.Metrics.Retried++
	} else {
		record.DeadLettered = true
		record.RetryCount++
		r.state.Metrics.DeadLettered++
		r.state.DeadLetterQueue = append(r.state.DeadLetterQueue, &OrderDeadLetterItem{
			OrderID:         record.OrderID,
			OfferingID:      record.OfferingID,
			ProviderAddress: record.ProviderAddress,
			Reason:          "fatal_error",
			Attempts:        record.RetryCount,
			DeadLetteredAt:  now,
			LastError:       record.LastError,
		})
		if r.cfg.DeadLetterAlerting && r.alerts != nil {
			_ = r.alerts.Alert(ctx, OrderRoutingAlert{
				OrderID:      record.OrderID,
				OfferingID:   record.OfferingID,
				Provider:     record.ProviderAddress,
				Attempts:     record.RetryCount,
				LastError:    record.LastError,
				DeadLettered: true,
				RecordedAt:   now,
			})
		}
	}
	r.state.Metrics.Failed++
	r.state.Records[record.OrderID] = record
	_ = r.stateStore.Save(r.state)

	if retryable {
		return err
	}
	return fmt.Errorf("order %s dead-lettered: %w", record.OrderID, err)
}

func (r *OrderRouter) retryLoop(ctx context.Context) {
	ticker := time.NewTicker(r.cfg.RetryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-r.stopCh:
			return
		case <-ticker.C:
			r.retryPending(ctx)
		}
	}
}

func (r *OrderRouter) retryPending(ctx context.Context) {
	r.mu.RLock()
	records := make([]OrderRoutingRecord, 0, len(r.state.Records))
	for _, record := range r.state.Records {
		if record.DeadLettered || record.RetryCount == 0 {
			continue
		}
		records = append(records, *record)
	}
	r.mu.RUnlock()

	for _, record := range records {
		if r.shouldRetry(record) {
			event := marketplace.OrderCreatedEvent{
				BaseMarketplaceEvent: marketplace.BaseMarketplaceEvent{
					Sequence: record.LastSequence,
				},
				OrderID:         record.OrderID,
				OfferingID:      record.OfferingID,
				ProviderAddress: record.ProviderAddress,
				CustomerAddress: record.CustomerAddress,
			}
			if err := r.createOrUpdateWaldurOrder(ctx, &record, event); err != nil {
				_ = r.handleFailure(ctx, &record, err)
				continue
			}
			r.mu.Lock()
			record.LastError = ""
			record.RetryCount = 0
			record.UpdatedAt = time.Now().UTC()
			r.state.Records[record.OrderID] = &record
			r.state.Metrics.Recovered++
			_ = r.stateStore.Save(r.state)
			r.mu.Unlock()
			log.Printf("[order-router] recovered order %s", record.OrderID)
		}
	}
}

func (r *OrderRouter) shouldRetry(record OrderRoutingRecord) bool {
	if record.RetryCount <= 0 || record.DeadLettered {
		return false
	}
	if record.RetryCount >= r.cfg.MaxRetries {
		r.mu.Lock()
		record.DeadLettered = true
		r.state.Metrics.DeadLettered++
		r.state.Records[record.OrderID] = &record
		r.state.DeadLetterQueue = append(r.state.DeadLetterQueue, &OrderDeadLetterItem{
			OrderID:         record.OrderID,
			OfferingID:      record.OfferingID,
			ProviderAddress: record.ProviderAddress,
			Reason:          "max_retries",
			Attempts:        record.RetryCount,
			DeadLetteredAt:  time.Now().UTC(),
			LastError:       record.LastError,
		})
		_ = r.stateStore.Save(r.state)
		r.mu.Unlock()
		if r.cfg.DeadLetterAlerting && r.alerts != nil {
			_ = r.alerts.Alert(context.Background(), OrderRoutingAlert{
				OrderID:      record.OrderID,
				OfferingID:   record.OfferingID,
				Provider:     record.ProviderAddress,
				Attempts:     record.RetryCount,
				LastError:    record.LastError,
				DeadLettered: true,
				RecordedAt:   time.Now().UTC(),
			})
		}
		return false
	}

	delay := retryDelay(r.cfg.RetryBackoff, r.cfg.MaxRetryBackoff, record.RetryCount)
	return time.Since(record.LastAttemptAt) >= delay
}

func retryDelay(base, max time.Duration, attempts int) time.Duration {
	if attempts <= 0 {
		return base
	}
	shift := attempts - 1
	if shift > 30 {
		shift = 30
	}
	delay := base * time.Duration(1<<shift)
	if delay > max {
		delay = max
	}
	jitterRange := delay/10 + 1
	jitter := time.Duration(0)
	if jitterRange > 0 {
		if n, err := rand.Int(rand.Reader, big.NewInt(int64(jitterRange))); err == nil {
			jitter = time.Duration(n.Int64())
		}
	}
	return delay + jitter
}

func sanitizeOrderID(orderID string) string {
	replacer := strings.NewReplacer("/", "-", ":", "-", " ", "-")
	return replacer.Replace(orderID)
}

func isRetryableOrderError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, waldur.ErrRateLimited) || errors.Is(err, waldur.ErrServerError) || errors.Is(err, waldur.ErrTimeout) {
		return true
	}
	if errors.Is(err, waldur.ErrUnauthorized) || errors.Is(err, waldur.ErrForbidden) {
		return false
	}
	if errors.Is(err, waldur.ErrNotFound) {
		return false
	}
	return true
}
