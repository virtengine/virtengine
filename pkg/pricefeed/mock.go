package pricefeed

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	sdkmath "cosmossdk.io/math"
)

// ============================================================================
// Mock Provider for Testing
// ============================================================================

// MockProvider is a mock implementation of Provider for testing
type MockProvider struct {
	name         string
	prices       map[string]PriceData
	pricesMu     sync.RWMutex
	healthy      bool
	healthyMu    sync.RWMutex
	errorOnGet   error
	requestCount atomic.Int64
	closed       atomic.Bool
	getLatency   time.Duration
	onGetPrice   func(ctx context.Context, baseAsset, quoteAsset string) (PriceData, error)
}

// NewMockProvider creates a new mock provider
func NewMockProvider(name string) *MockProvider {
	return &MockProvider{
		name:    name,
		prices:  make(map[string]PriceData),
		healthy: true,
	}
}

// Name returns the provider name
func (m *MockProvider) Name() string {
	return m.name
}

// Type returns the provider type
func (m *MockProvider) Type() SourceType {
	return SourceTypeMock
}

// SetPrice sets a mock price for testing
func (m *MockProvider) SetPrice(baseAsset, quoteAsset string, price sdkmath.LegacyDec) {
	m.pricesMu.Lock()
	defer m.pricesMu.Unlock()
	key := baseAsset + "/" + quoteAsset
	m.prices[key] = PriceData{
		BaseAsset:  baseAsset,
		QuoteAsset: quoteAsset,
		Price:      price,
		Timestamp:  time.Now().UTC(),
		Source:     m.name,
		Confidence: 0.95,
	}
}

// SetPriceData sets a full price data object
func (m *MockProvider) SetPriceData(data PriceData) {
	m.pricesMu.Lock()
	defer m.pricesMu.Unlock()
	key := data.BaseAsset + "/" + data.QuoteAsset
	m.prices[key] = data
}

// SetError sets an error to return on GetPrice
func (m *MockProvider) SetError(err error) {
	m.errorOnGet = err
}

// SetHealthy sets the health status
func (m *MockProvider) SetHealthy(healthy bool) {
	m.healthyMu.Lock()
	defer m.healthyMu.Unlock()
	m.healthy = healthy
}

// SetLatency sets artificial latency for testing
func (m *MockProvider) SetLatency(d time.Duration) {
	m.getLatency = d
}

// SetOnGetPrice sets a custom handler for GetPrice
func (m *MockProvider) SetOnGetPrice(fn func(ctx context.Context, baseAsset, quoteAsset string) (PriceData, error)) {
	m.onGetPrice = fn
}

// GetPrice returns a mock price
func (m *MockProvider) GetPrice(ctx context.Context, baseAsset, quoteAsset string) (PriceData, error) {
	if m.closed.Load() {
		return PriceData{}, ErrSourceUnhealthy
	}

	m.requestCount.Add(1)

	// Simulate latency
	if m.getLatency > 0 {
		select {
		case <-time.After(m.getLatency):
		case <-ctx.Done():
			return PriceData{}, ctx.Err()
		}
	}

	// Check custom handler
	if m.onGetPrice != nil {
		return m.onGetPrice(ctx, baseAsset, quoteAsset)
	}

	// Check for error
	if m.errorOnGet != nil {
		return PriceData{}, m.errorOnGet
	}

	m.pricesMu.RLock()
	defer m.pricesMu.RUnlock()

	key := baseAsset + "/" + quoteAsset
	if price, ok := m.prices[key]; ok {
		return price, nil
	}

	return PriceData{}, ErrPriceNotFound
}

// GetPrices returns mock prices for multiple pairs
func (m *MockProvider) GetPrices(ctx context.Context, pairs []AssetPair) (map[string]PriceData, error) {
	if m.closed.Load() {
		return nil, ErrSourceUnhealthy
	}

	if m.errorOnGet != nil {
		return nil, m.errorOnGet
	}

	results := make(map[string]PriceData)
	for _, pair := range pairs {
		price, err := m.GetPrice(ctx, pair.Base, pair.Quote)
		if err == nil {
			results[pair.String()] = price
		}
	}

	return results, nil
}

// IsHealthy returns the mock health status
func (m *MockProvider) IsHealthy(ctx context.Context) bool {
	if m.closed.Load() {
		return false
	}
	m.healthyMu.RLock()
	defer m.healthyMu.RUnlock()
	return m.healthy
}

// Health returns mock health information
func (m *MockProvider) Health() SourceHealth {
	m.healthyMu.RLock()
	healthy := m.healthy
	m.healthyMu.RUnlock()

	return SourceHealth{
		Source:       m.name,
		Type:         SourceTypeMock,
		Healthy:      healthy && !m.closed.Load(),
		LastCheck:    time.Now(),
		LastSuccess:  time.Now(),
		RequestCount: m.requestCount.Load(),
	}
}

// RequestCount returns the number of requests made
func (m *MockProvider) RequestCount() int64 {
	return m.requestCount.Load()
}

// Close closes the mock provider
func (m *MockProvider) Close() error {
	m.closed.Store(true)
	return nil
}

// Reset resets the mock provider state
func (m *MockProvider) Reset() {
	m.pricesMu.Lock()
	m.prices = make(map[string]PriceData)
	m.pricesMu.Unlock()
	m.errorOnGet = nil
	m.SetHealthy(true)
	m.requestCount.Store(0)
	m.closed.Store(false)
	m.getLatency = 0
	m.onGetPrice = nil
}
