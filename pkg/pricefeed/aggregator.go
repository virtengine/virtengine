package pricefeed

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"
)

// ============================================================================
// Aggregator Implementation
// ============================================================================

// PriceFeedAggregator implements the Aggregator interface
type PriceFeedAggregator struct {
	config    Config
	providers map[string]Provider
	cache     *InMemoryCache
	metrics   Metrics
	mu        sync.RWMutex
	closed    bool

	// Background health checker
	healthCheckStop chan struct{}
}

// NewAggregator creates a new price feed aggregator
func NewAggregator(cfg Config) (*PriceFeedAggregator, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	agg := &PriceFeedAggregator{
		config:          cfg,
		providers:       make(map[string]Provider),
		healthCheckStop: make(chan struct{}),
	}

	// Initialize cache if enabled
	if cfg.CacheConfig.Enabled {
		agg.cache = NewInMemoryCache(cfg.CacheConfig)
	}

	// Initialize providers
	for _, provCfg := range cfg.EnabledProviders() {
		provider, err := createProvider(provCfg, cfg.RetryConfig)
		if err != nil {
			// Log error but continue with other providers
			continue
		}
		agg.providers[provCfg.Name] = provider
	}

	if len(agg.providers) == 0 {
		return nil, ErrPriceFeedUnavailable
	}

	// Start background health checker
	if cfg.HealthCheckInterval > 0 {
		go agg.healthCheckLoop()
	}

	return agg, nil
}

// createProvider creates a provider based on configuration
func createProvider(cfg ProviderConfig, retryCfg RetryConfig) (Provider, error) {
	switch cfg.Type {
	case SourceTypeCoinGecko:
		if cfg.CoinGeckoConfig == nil {
			return nil, fmt.Errorf("CoinGecko config required")
		}
		return NewCoinGeckoProvider(cfg.Name, *cfg.CoinGeckoConfig, retryCfg)

	case SourceTypeChainlink:
		if cfg.ChainlinkConfig == nil {
			return nil, fmt.Errorf("Chainlink config required")
		}
		return NewChainlinkProvider(cfg.Name, *cfg.ChainlinkConfig, retryCfg)

	case SourceTypePyth:
		if cfg.PythConfig == nil {
			return nil, fmt.Errorf("Pyth config required")
		}
		return NewPythProvider(cfg.Name, *cfg.PythConfig, retryCfg)

	default:
		return nil, fmt.Errorf("unsupported provider type: %s", cfg.Type)
	}
}

// GetPrice returns the aggregated price from configured sources
func (a *PriceFeedAggregator) GetPrice(ctx context.Context, baseAsset, quoteAsset string) (AggregatedPrice, error) {
	a.mu.RLock()
	if a.closed {
		a.mu.RUnlock()
		return AggregatedPrice{}, ErrPriceFeedUnavailable
	}
	a.mu.RUnlock()

	cacheKey := cacheKey(baseAsset, quoteAsset)

	// Check cache first
	if a.cache != nil {
		if cached, ok := a.cache.Get(cacheKey); ok {
			if a.metrics != nil {
				a.metrics.RecordCacheHit(baseAsset, quoteAsset)
			}
			return AggregatedPrice{
				PriceData:    cached,
				Strategy:     a.config.Strategy,
				SourcePrices: []PriceData{cached},
				SourceCount:  1,
			}, nil
		}
		if a.metrics != nil {
			a.metrics.RecordCacheMiss(baseAsset, quoteAsset)
		}
	}

	// Fetch from providers based on strategy
	switch a.config.Strategy {
	case StrategyPrimary:
		return a.getPricePrimary(ctx, baseAsset, quoteAsset)
	case StrategyMedian:
		return a.getPriceMedian(ctx, baseAsset, quoteAsset)
	case StrategyWeighted:
		return a.getPriceWeighted(ctx, baseAsset, quoteAsset)
	default:
		return a.getPricePrimary(ctx, baseAsset, quoteAsset)
	}
}

// getPricePrimary uses the first healthy source in priority order
func (a *PriceFeedAggregator) getPricePrimary(ctx context.Context, baseAsset, quoteAsset string) (AggregatedPrice, error) {
	providers := a.getProvidersByPriority()

	var lastErr error
	for _, provider := range providers {
		if !provider.IsHealthy(ctx) {
			continue
		}

		price, err := provider.GetPrice(ctx, baseAsset, quoteAsset)
		if err != nil {
			lastErr = err
			continue
		}

		// Cache the result
		if a.cache != nil {
			a.cache.Set(cacheKey(baseAsset, quoteAsset), price)
		}

		return AggregatedPrice{
			PriceData:    price,
			Strategy:     StrategyPrimary,
			SourcePrices: []PriceData{price},
			SourceCount:  1,
		}, nil
	}

	// Try stale cache as fallback
	if a.cache != nil && a.config.CacheConfig.AllowStale {
		if stale, ok := a.cache.GetStale(cacheKey(baseAsset, quoteAsset)); ok {
			return AggregatedPrice{
				PriceData:    stale,
				Strategy:     StrategyPrimary,
				SourcePrices: []PriceData{stale},
				SourceCount:  1,
			}, nil
		}
	}

	if lastErr != nil {
		return AggregatedPrice{}, lastErr
	}
	return AggregatedPrice{}, ErrAllSourcesFailed
}

// getPriceMedian uses the median price across all sources
func (a *PriceFeedAggregator) getPriceMedian(ctx context.Context, baseAsset, quoteAsset string) (AggregatedPrice, error) {
	prices, err := a.fetchFromAllProviders(ctx, baseAsset, quoteAsset)
	if err != nil {
		return AggregatedPrice{}, err
	}

	if len(prices) == 0 {
		return AggregatedPrice{}, ErrAllSourcesFailed
	}

	// Sort by price
	sort.Slice(prices, func(i, j int) bool {
		return prices[i].Price.LT(prices[j].Price)
	})

	// Get median
	medianIdx := len(prices) / 2
	medianPrice := prices[medianIdx]

	// Check deviation
	if len(prices) > 1 {
		deviation := a.calculateDeviation(prices)
		if deviation > a.config.MaxPriceDeviation {
			if a.metrics != nil {
				a.metrics.RecordPriceDeviation(baseAsset, quoteAsset, deviation)
			}
			return AggregatedPrice{}, ErrPriceDeviation
		}
	}

	// Cache the result
	if a.cache != nil {
		a.cache.Set(cacheKey(baseAsset, quoteAsset), medianPrice)
	}

	return AggregatedPrice{
		PriceData:    medianPrice,
		Strategy:     StrategyMedian,
		SourcePrices: prices,
		Deviation:    a.calculateDeviation(prices),
		SourceCount:  len(prices),
	}, nil
}

// getPriceWeighted weights prices by source confidence/volume
func (a *PriceFeedAggregator) getPriceWeighted(ctx context.Context, baseAsset, quoteAsset string) (AggregatedPrice, error) {
	prices, err := a.fetchFromAllProviders(ctx, baseAsset, quoteAsset)
	if err != nil {
		return AggregatedPrice{}, err
	}

	if len(prices) == 0 {
		return AggregatedPrice{}, ErrAllSourcesFailed
	}

	// Calculate weighted average
	var totalWeight float64
	weightedSum := sdkmath.LegacyZeroDec()

	providerWeights := make(map[string]float64)
	for _, cfg := range a.config.Providers {
		providerWeights[cfg.Name] = cfg.Weight
	}

	for _, p := range prices {
		weight := providerWeights[p.Source]
		if weight == 0 {
			weight = 1.0 / float64(len(prices)) // Default equal weight
		}
		// Adjust weight by confidence
		weight *= p.Confidence

		totalWeight += weight
		weightDec := sdkmath.LegacyMustNewDecFromStr(fmt.Sprintf("%f", weight))
		contribution := p.Price.MulTruncate(weightDec)
		weightedSum = weightedSum.Add(contribution)
	}

	if totalWeight == 0 {
		return AggregatedPrice{}, ErrAllSourcesFailed
	}

	totalWeightDec := sdkmath.LegacyMustNewDecFromStr(fmt.Sprintf("%f", totalWeight))
	avgPrice := weightedSum.QuoTruncate(totalWeightDec)

	result := PriceData{
		BaseAsset:  baseAsset,
		QuoteAsset: quoteAsset,
		Price:      avgPrice,
		Timestamp:  time.Now().UTC(),
		Source:     "aggregated",
		Confidence: a.calculateAggregatedConfidence(prices),
	}

	// Cache the result
	if a.cache != nil {
		a.cache.Set(cacheKey(baseAsset, quoteAsset), result)
	}

	return AggregatedPrice{
		PriceData:    result,
		Strategy:     StrategyWeighted,
		SourcePrices: prices,
		Deviation:    a.calculateDeviation(prices),
		SourceCount:  len(prices),
	}, nil
}

// fetchFromAllProviders fetches prices from all healthy providers
func (a *PriceFeedAggregator) fetchFromAllProviders(ctx context.Context, baseAsset, quoteAsset string) ([]PriceData, error) {
	a.mu.RLock()
	providers := make([]Provider, 0, len(a.providers))
	for _, p := range a.providers {
		providers = append(providers, p)
	}
	a.mu.RUnlock()

	var prices []PriceData
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, provider := range providers {
		if !provider.IsHealthy(ctx) {
			continue
		}

		wg.Add(1)
		go func(p Provider) {
			defer wg.Done()
			price, err := p.GetPrice(ctx, baseAsset, quoteAsset)
			if err == nil && price.IsValid() {
				mu.Lock()
				prices = append(prices, price)
				mu.Unlock()
			}
		}(provider)
	}

	wg.Wait()

	if len(prices) == 0 {
		// Try stale cache as fallback
		if a.cache != nil && a.config.CacheConfig.AllowStale {
			if stale, ok := a.cache.GetStale(cacheKey(baseAsset, quoteAsset)); ok {
				return []PriceData{stale}, nil
			}
		}
		return nil, ErrAllSourcesFailed
	}

	return prices, nil
}

// calculateDeviation calculates the max percentage deviation between prices
func (a *PriceFeedAggregator) calculateDeviation(prices []PriceData) float64 {
	if len(prices) < 2 {
		return 0
	}

	var minPrice, maxPrice sdkmath.LegacyDec
	for i, p := range prices {
		if i == 0 || p.Price.LT(minPrice) {
			minPrice = p.Price
		}
		if i == 0 || p.Price.GT(maxPrice) {
			maxPrice = p.Price
		}
	}

	if minPrice.IsZero() {
		return 0
	}

	deviation := maxPrice.Sub(minPrice).Quo(minPrice)
	devFloat, _ := deviation.Float64()
	return devFloat
}

// calculateAggregatedConfidence calculates aggregated confidence
func (a *PriceFeedAggregator) calculateAggregatedConfidence(prices []PriceData) float64 {
	if len(prices) == 0 {
		return 0
	}

	var sum float64
	for _, p := range prices {
		sum += p.Confidence
	}
	return sum / float64(len(prices))
}

// getProvidersByPriority returns providers sorted by priority
func (a *PriceFeedAggregator) getProvidersByPriority() []Provider {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Get enabled providers config sorted by priority
	configs := a.config.EnabledProviders()

	providers := make([]Provider, 0, len(configs))
	for _, cfg := range configs {
		if p, ok := a.providers[cfg.Name]; ok {
			providers = append(providers, p)
		}
	}

	return providers
}

// GetPrices returns aggregated prices for multiple pairs
func (a *PriceFeedAggregator) GetPrices(ctx context.Context, pairs []AssetPair) (map[string]AggregatedPrice, error) {
	results := make(map[string]AggregatedPrice)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, pair := range pairs {
		wg.Add(1)
		go func(p AssetPair) {
			defer wg.Done()
			price, err := a.GetPrice(ctx, p.Base, p.Quote)
			if err == nil {
				mu.Lock()
				results[p.String()] = price
				mu.Unlock()
			}
		}(pair)
	}

	wg.Wait()
	return results, nil
}

// GetPriceFromSource returns price from a specific source
func (a *PriceFeedAggregator) GetPriceFromSource(ctx context.Context, source string, baseAsset, quoteAsset string) (PriceData, error) {
	a.mu.RLock()
	provider, ok := a.providers[source]
	a.mu.RUnlock()

	if !ok {
		return PriceData{}, fmt.Errorf("provider not found: %s", source)
	}

	return provider.GetPrice(ctx, baseAsset, quoteAsset)
}

// ListSources returns all configured price sources
func (a *PriceFeedAggregator) ListSources() []SourceHealth {
	a.mu.RLock()
	defer a.mu.RUnlock()

	sources := make([]SourceHealth, 0, len(a.providers))
	for _, p := range a.providers {
		sources = append(sources, p.Health())
	}
	return sources
}

// AddProvider adds a new price provider
func (a *PriceFeedAggregator) AddProvider(provider Provider) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.closed {
		return ErrPriceFeedUnavailable
	}

	name := provider.Name()
	if _, exists := a.providers[name]; exists {
		return fmt.Errorf("provider already exists: %s", name)
	}

	a.providers[name] = provider
	return nil
}

// RemoveProvider removes a price provider by name
func (a *PriceFeedAggregator) RemoveProvider(name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	provider, ok := a.providers[name]
	if !ok {
		return fmt.Errorf("provider not found: %s", name)
	}

	delete(a.providers, name)
	return provider.Close()
}

// RefreshCache forces a cache refresh for the given pairs
func (a *PriceFeedAggregator) RefreshCache(ctx context.Context, pairs []AssetPair) error {
	if a.cache == nil {
		return nil
	}

	for _, pair := range pairs {
		a.cache.Delete(cacheKey(pair.Base, pair.Quote))
	}

	// Fetch fresh prices
	_, err := a.GetPrices(ctx, pairs)
	return err
}

// healthCheckLoop periodically checks provider health
func (a *PriceFeedAggregator) healthCheckLoop() {
	ticker := time.NewTicker(a.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.checkHealth()
		case <-a.healthCheckStop:
			return
		}
	}
}

// checkHealth checks health of all providers
func (a *PriceFeedAggregator) checkHealth() {
	a.mu.RLock()
	providers := make([]Provider, 0, len(a.providers))
	for _, p := range a.providers {
		providers = append(providers, p)
	}
	a.mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, p := range providers {
		healthy := p.IsHealthy(ctx)
		if a.metrics != nil {
			latency := p.Health().Latency.Seconds()
			a.metrics.RecordSourceHealth(p.Name(), healthy, latency)
		}
	}
}

// SetMetrics sets the metrics collector
func (a *PriceFeedAggregator) SetMetrics(m Metrics) {
	a.metrics = m
}

// CacheStats returns cache statistics
func (a *PriceFeedAggregator) CacheStats() CacheStats {
	if a.cache == nil {
		return CacheStats{}
	}
	return a.cache.Stats()
}

// Close closes the aggregator and all providers
func (a *PriceFeedAggregator) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.closed {
		return nil
	}

	a.closed = true
	close(a.healthCheckStop)

	var errs []string
	for name, p := range a.providers {
		if err := p.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing providers: %s", strings.Join(errs, "; "))
	}

	return nil
}
