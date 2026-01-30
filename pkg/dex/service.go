// Package dex provides DEX integration for crypto-to-fiat conversions.
//
// VE-905: DEX integration for crypto-to-fiat conversions
package dex

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"
)

// Ensure service implements Service interface
var _ Service = (*service)(nil)

// service is the main DEX service implementation
type service struct {
	cfg        Config
	adapters   map[string]Adapter
	adaptersMu sync.RWMutex

	priceFeed  *priceFeedImpl
	swapExec   *swapExecutorImpl
	offRamp    *offRampBridgeImpl
	breaker    *circuitBreaker

	started bool
	stopCh  chan struct{}
	wg      sync.WaitGroup
	mu      sync.RWMutex
}

// NewService creates a new DEX service
func NewService(cfg Config) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	s := &service{
		cfg:      cfg,
		adapters: make(map[string]Adapter),
		stopCh:   make(chan struct{}),
	}

	// Initialize price feed
	s.priceFeed = newPriceFeed(cfg.PriceFeed)

	// Initialize swap executor
	s.swapExec = newSwapExecutor(cfg.Swap, s)

	// Initialize off-ramp bridge
	s.offRamp = newOffRampBridge(cfg.OffRamp)

	// Initialize circuit breaker
	s.breaker = newCircuitBreaker(cfg.CircuitBreaker)

	return s, nil
}

// ============================================================================
// Adapter Management
// ============================================================================

// RegisterAdapter registers a DEX adapter
func (s *service) RegisterAdapter(adapter Adapter) error {
	if adapter == nil {
		return errors.New("adapter cannot be nil")
	}

	s.adaptersMu.Lock()
	defer s.adaptersMu.Unlock()

	name := adapter.Name()
	if _, exists := s.adapters[name]; exists {
		return fmt.Errorf("adapter %q already registered", name)
	}

	s.adapters[name] = adapter

	// Register adapter as a price source
	if err := s.priceFeed.registerAdapterSource(adapter); err != nil {
		delete(s.adapters, name)
		return fmt.Errorf("failed to register adapter as price source: %w", err)
	}

	return nil
}

// UnregisterAdapter removes a DEX adapter
func (s *service) UnregisterAdapter(name string) error {
	s.adaptersMu.Lock()
	defer s.adaptersMu.Unlock()

	adapter, exists := s.adapters[name]
	if !exists {
		return ErrAdapterNotFound
	}

	// Unregister from price feed
	s.priceFeed.UnregisterSource(name)

	// Close the adapter
	if err := adapter.Close(); err != nil {
		return fmt.Errorf("failed to close adapter: %w", err)
	}

	delete(s.adapters, name)
	return nil
}

// GetAdapter returns a specific adapter
func (s *service) GetAdapter(name string) (Adapter, error) {
	s.adaptersMu.RLock()
	defer s.adaptersMu.RUnlock()

	adapter, exists := s.adapters[name]
	if !exists {
		return nil, ErrAdapterNotFound
	}
	return adapter, nil
}

// ListAdapters returns names of all registered adapters
func (s *service) ListAdapters() []string {
	s.adaptersMu.RLock()
	defer s.adaptersMu.RUnlock()

	names := make([]string, 0, len(s.adapters))
	for name := range s.adapters {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// getHealthyAdapters returns all healthy adapters sorted by priority
func (s *service) getHealthyAdapters(ctx context.Context) []Adapter {
	s.adaptersMu.RLock()
	defer s.adaptersMu.RUnlock()

	var healthy []Adapter
	for _, adapter := range s.adapters {
		if adapter.IsHealthy(ctx) {
			healthy = append(healthy, adapter)
		}
	}
	return healthy
}

// ============================================================================
// Price Feed Operations
// ============================================================================

// GetPrice fetches the current price for a trading pair
func (s *service) GetPrice(ctx context.Context, baseSymbol, quoteSymbol string) (Price, error) {
	if err := s.checkHealth(ctx); err != nil {
		return Price{}, err
	}
	return s.priceFeed.GetPrice(ctx, baseSymbol, quoteSymbol)
}

// GetPriceAggregate fetches aggregated price data
func (s *service) GetPriceAggregate(ctx context.Context, baseSymbol, quoteSymbol string) (PriceAggregate, error) {
	if err := s.checkHealth(ctx); err != nil {
		return PriceAggregate{}, err
	}
	return s.priceFeed.GetPriceAggregate(ctx, baseSymbol, quoteSymbol)
}

// GetTWAP fetches the time-weighted average price
func (s *service) GetTWAP(ctx context.Context, baseSymbol, quoteSymbol string, window time.Duration) (sdkmath.LegacyDec, error) {
	if err := s.checkHealth(ctx); err != nil {
		return sdkmath.LegacyDec{}, err
	}
	return s.priceFeed.GetTWAP(ctx, baseSymbol, quoteSymbol, window)
}

// ============================================================================
// Liquidity Pool Operations
// ============================================================================

// GetPool fetches a liquidity pool by ID
func (s *service) GetPool(ctx context.Context, dex, poolID string) (LiquidityPool, error) {
	if err := s.checkHealth(ctx); err != nil {
		return LiquidityPool{}, err
	}

	adapter, err := s.GetAdapter(dex)
	if err != nil {
		return LiquidityPool{}, err
	}

	return adapter.GetPool(ctx, poolID)
}

// ListPools lists liquidity pools matching the query
func (s *service) ListPools(ctx context.Context, query PoolQuery) ([]LiquidityPool, error) {
	if err := s.checkHealth(ctx); err != nil {
		return nil, err
	}

	// If DEX is specified, query only that adapter
	if query.DEX != "" {
		adapter, err := s.GetAdapter(query.DEX)
		if err != nil {
			return nil, err
		}
		return adapter.ListPools(ctx, query)
	}

	// Query all adapters
	var allPools []LiquidityPool
	adapters := s.getHealthyAdapters(ctx)

	for _, adapter := range adapters {
		pools, err := adapter.ListPools(ctx, query)
		if err != nil {
			continue // Skip failed adapters
		}
		allPools = append(allPools, pools...)
	}

	// Apply limit if specified
	if query.Limit > 0 && len(allPools) > query.Limit {
		allPools = allPools[:query.Limit]
	}

	return allPools, nil
}

// ============================================================================
// Swap Operations
// ============================================================================

// GetSwapQuote generates a swap quote
func (s *service) GetSwapQuote(ctx context.Context, request SwapRequest) (SwapQuote, error) {
	if err := s.checkHealth(ctx); err != nil {
		return SwapQuote{}, err
	}

	if err := request.Validate(); err != nil {
		return SwapQuote{}, fmt.Errorf("invalid swap request: %w", err)
	}

	// Check circuit breaker
	if s.breaker.IsTripped() {
		return SwapQuote{}, ErrCircuitBreakerTripped
	}

	return s.swapExec.GetQuote(ctx, request)
}

// ExecuteSwap executes a previously quoted swap
func (s *service) ExecuteSwap(ctx context.Context, quote SwapQuote, signedTx []byte) (SwapResult, error) {
	if err := s.checkHealth(ctx); err != nil {
		return SwapResult{}, err
	}

	// Check circuit breaker
	if s.breaker.IsTripped() {
		return SwapResult{}, ErrCircuitBreakerTripped
	}

	result, err := s.swapExec.ExecuteSwap(ctx, quote, signedTx)
	if err != nil {
		s.breaker.RecordFailure()
		return SwapResult{}, err
	}

	s.breaker.RecordSuccess()
	return result, nil
}

// FindBestRoute finds the optimal swap route
func (s *service) FindBestRoute(ctx context.Context, request SwapRequest) (SwapRoute, error) {
	if err := s.checkHealth(ctx); err != nil {
		return SwapRoute{}, err
	}

	if err := request.Validate(); err != nil {
		return SwapRoute{}, fmt.Errorf("invalid swap request: %w", err)
	}

	return s.swapExec.FindBestRoute(ctx, request)
}

// ============================================================================
// Off-Ramp Operations
// ============================================================================

// GetOffRampQuote generates an off-ramp quote
func (s *service) GetOffRampQuote(ctx context.Context, request OffRampRequest) (OffRampQuote, error) {
	if err := s.checkHealth(ctx); err != nil {
		return OffRampQuote{}, err
	}

	if err := request.Validate(); err != nil {
		return OffRampQuote{}, fmt.Errorf("invalid off-ramp request: %w", err)
	}

	// Validate KYC/VEID requirements
	if err := s.offRamp.ValidateKYC(ctx, request.Sender, request.VEIDScore); err != nil {
		return OffRampQuote{}, err
	}

	return s.offRamp.GetQuote(ctx, request)
}

// InitiateOffRamp initiates an off-ramp operation
func (s *service) InitiateOffRamp(ctx context.Context, quote OffRampQuote, signedTx []byte) (OffRampResult, error) {
	if err := s.checkHealth(ctx); err != nil {
		return OffRampResult{}, err
	}

	return s.offRamp.InitiateOffRamp(ctx, quote, signedTx)
}

// GetOffRampStatus fetches the status of an off-ramp operation
func (s *service) GetOffRampStatus(ctx context.Context, offRampID string) (OffRampResult, error) {
	return s.offRamp.GetStatus(ctx, offRampID)
}

// CancelOffRamp cancels a pending off-ramp operation
func (s *service) CancelOffRamp(ctx context.Context, offRampID string) error {
	return s.offRamp.CancelOffRamp(ctx, offRampID)
}

// ============================================================================
// Health & Lifecycle
// ============================================================================

// checkHealth verifies service is operational
func (s *service) checkHealth(ctx context.Context) error {
	s.mu.RLock()
	started := s.started
	s.mu.RUnlock()

	if !started {
		return errors.New("service not started")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}

// IsHealthy checks if the service is operational
func (s *service) IsHealthy(ctx context.Context) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.started {
		return false
	}

	// Check at least one adapter is healthy
	adapters := s.getHealthyAdapters(ctx)
	return len(adapters) > 0
}

// Start starts the DEX service
func (s *service) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return errors.New("service already started")
	}

	// Start price feed background updater
	s.wg.Add(1)
	go s.priceFeed.runUpdater(ctx, s.stopCh, &s.wg)

	s.started = true
	return nil
}

// Stop stops the DEX service
func (s *service) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return nil
	}

	// Signal stop
	close(s.stopCh)

	// Wait for goroutines
	s.wg.Wait()

	// Close all adapters
	s.adaptersMu.Lock()
	for name, adapter := range s.adapters {
		_ = adapter.Close()
		delete(s.adapters, name)
	}
	s.adaptersMu.Unlock()

	s.started = false
	return nil
}
