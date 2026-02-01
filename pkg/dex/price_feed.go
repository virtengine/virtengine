// Package dex provides DEX integration for crypto-to-fiat conversions.
//
// VE-905: DEX integration for crypto-to-fiat conversions
package dex

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
	"context"
	"sort"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"
)

// priceFeedImpl implements the PriceFeed interface
type priceFeedImpl struct {
	cfg        PriceFeedConfig
	sources    map[string]PriceSource
	sourcesMu  sync.RWMutex
	cache      *priceCache
	history    *priceHistory
	subs       map[string][]PriceCallback
	subsMu     sync.RWMutex
}

// priceCache provides thread-safe price caching
type priceCache struct {
	prices map[string]Price
	mu     sync.RWMutex
	ttl    time.Duration
}

// priceHistory stores historical price data for TWAP/VWAP
type priceHistory struct {
	entries map[string][]priceEntry
	mu      sync.RWMutex
	maxAge  time.Duration
}

type priceEntry struct {
	price     sdkmath.LegacyDec
	volume    sdkmath.Int
	timestamp time.Time
}

// newPriceFeed creates a new price feed implementation
func newPriceFeed(cfg PriceFeedConfig) *priceFeedImpl {
	return &priceFeedImpl{
		cfg:     cfg,
		sources: make(map[string]PriceSource),
		cache: &priceCache{
			prices: make(map[string]Price),
			ttl:    cfg.CacheTTL,
		},
		history: &priceHistory{
			entries: make(map[string][]priceEntry),
			maxAge:  cfg.VWAPWindow,
		},
		subs: make(map[string][]PriceCallback),
	}
}

// pairKey generates a cache key for a trading pair
func pairKey(baseSymbol, quoteSymbol string) string {
	return baseSymbol + "/" + quoteSymbol
}

// GetPrice fetches the current price for a trading pair
func (p *priceFeedImpl) GetPrice(ctx context.Context, baseSymbol, quoteSymbol string) (Price, error) {
	key := pairKey(baseSymbol, quoteSymbol)

	// Check cache first
	if p.cfg.CacheEnabled {
		if cached, ok := p.cache.get(key); ok {
			return cached, nil
		}
	}

	// Fetch from sources
	prices, err := p.fetchFromSources(ctx, baseSymbol, quoteSymbol)
	if err != nil {
		return Price{}, err
	}

	if len(prices) == 0 {
		return Price{}, ErrUnsupportedPair
	}

	// Use median price
	price := p.selectMedianPrice(prices)

	// Update cache
	if p.cfg.CacheEnabled {
		p.cache.set(key, price)
	}

	// Update history
	p.history.add(key, price.Rate, sdkmath.ZeroInt())

	return price, nil
}

// GetPriceAggregate fetches aggregated price data from multiple sources
func (p *priceFeedImpl) GetPriceAggregate(ctx context.Context, baseSymbol, quoteSymbol string) (PriceAggregate, error) {
	key := pairKey(baseSymbol, quoteSymbol)

	prices, err := p.fetchFromSources(ctx, baseSymbol, quoteSymbol)
	if err != nil {
		return PriceAggregate{}, err
	}

	if len(prices) == 0 {
		return PriceAggregate{}, ErrUnsupportedPair
	}

	// Calculate median
	median := p.selectMedianPrice(prices)

	// Get TWAP and VWAP from history
	twap := p.calculateTWAP(key, p.cfg.TWAPWindow)
	vwap := p.calculateVWAP(key, p.cfg.VWAPWindow)

	return PriceAggregate{
		Pair: TradingPair{
			BaseToken:  Token{Symbol: baseSymbol},
			QuoteToken: Token{Symbol: quoteSymbol},
		},
		MedianPrice:  median.Rate,
		TWAP:         twap,
		VWAP:         vwap,
		Sources:      prices,
		AggregatedAt: time.Now().UTC(),
	}, nil
}

// GetTWAP fetches the time-weighted average price
func (p *priceFeedImpl) GetTWAP(ctx context.Context, baseSymbol, quoteSymbol string, window time.Duration) (sdkmath.LegacyDec, error) {
	key := pairKey(baseSymbol, quoteSymbol)
	twap := p.calculateTWAP(key, window)
	if twap.IsZero() {
		// Fetch current price if no history
		price, err := p.GetPrice(ctx, baseSymbol, quoteSymbol)
		if err != nil {
			return sdkmath.LegacyDec{}, err
		}
		return price.Rate, nil
	}
	return twap, nil
}

// GetVWAP fetches the volume-weighted average price
func (p *priceFeedImpl) GetVWAP(ctx context.Context, baseSymbol, quoteSymbol string, window time.Duration) (sdkmath.LegacyDec, error) {
	key := pairKey(baseSymbol, quoteSymbol)
	vwap := p.calculateVWAP(key, window)
	if vwap.IsZero() {
		// Fall back to TWAP if no volume data
		return p.GetTWAP(ctx, baseSymbol, quoteSymbol, window)
	}
	return vwap, nil
}

// SubscribePrice subscribes to price updates
func (p *priceFeedImpl) SubscribePrice(ctx context.Context, baseSymbol, quoteSymbol string, callback PriceCallback) (Subscription, error) {
	key := pairKey(baseSymbol, quoteSymbol)

	p.subsMu.Lock()
	p.subs[key] = append(p.subs[key], callback)
	idx := len(p.subs[key]) - 1
	p.subsMu.Unlock()

	return &priceSubscription{
		feed: p,
		key:  key,
		idx:  idx,
	}, nil
}

// RegisterSource registers a price data source
func (p *priceFeedImpl) RegisterSource(source PriceSource) error {
	p.sourcesMu.Lock()
	defer p.sourcesMu.Unlock()

	p.sources[source.Name()] = source
	return nil
}

// UnregisterSource removes a price data source
func (p *priceFeedImpl) UnregisterSource(name string) error {
	p.sourcesMu.Lock()
	defer p.sourcesMu.Unlock()

	delete(p.sources, name)
	return nil
}

// registerAdapterSource registers a DEX adapter as a price source
func (p *priceFeedImpl) registerAdapterSource(adapter Adapter) error {
	source := &adapterPriceSource{adapter: adapter}
	return p.RegisterSource(source)
}

// fetchFromSources fetches prices from all sources
//
//nolint:unparam // result 1 (error) reserved for future aggregation failures
func (p *priceFeedImpl) fetchFromSources(ctx context.Context, baseSymbol, quoteSymbol string) ([]Price, error) {
	p.sourcesMu.RLock()
	sources := make([]PriceSource, 0, len(p.sources))
	for _, source := range p.sources {
		sources = append(sources, source)
	}
	p.sourcesMu.RUnlock()

	var prices []Price
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, source := range sources {
		if !source.IsHealthy(ctx) {
			continue
		}

		wg.Add(1)
		src := source // capture loop variable
		verrors.SafeGo("price-feed", func() {
			defer wg.Done()
			price, err := src.GetPrice(ctx, baseSymbol, quoteSymbol)
			if err != nil {
				return
			}
			if !price.IsStale(p.cfg.MaxPriceAge) {
				mu.Lock()
				prices = append(prices, price)
				mu.Unlock()
			}
		})
	}

	wg.Wait()
	return prices, nil
}

// selectMedianPrice selects the median price from a list
func (p *priceFeedImpl) selectMedianPrice(prices []Price) Price {
	if len(prices) == 0 {
		return Price{}
	}
	if len(prices) == 1 {
		return prices[0]
	}

	// Sort by rate
	sort.Slice(prices, func(i, j int) bool {
		return prices[i].Rate.LT(prices[j].Rate)
	})

	// Return median
	return prices[len(prices)/2]
}

// calculateTWAP calculates the time-weighted average price
func (p *priceFeedImpl) calculateTWAP(key string, window time.Duration) sdkmath.LegacyDec {
	p.history.mu.RLock()
	entries := p.history.entries[key]
	p.history.mu.RUnlock()

	if len(entries) == 0 {
		return sdkmath.LegacyZeroDec()
	}

	cutoff := time.Now().Add(-window)
	var sum sdkmath.LegacyDec = sdkmath.LegacyZeroDec()
	var count int

	for _, e := range entries {
		if e.timestamp.After(cutoff) {
			sum = sum.Add(e.price)
			count++
		}
	}

	if count == 0 {
		return sdkmath.LegacyZeroDec()
	}

	return sum.Quo(sdkmath.LegacyNewDec(int64(count)))
}

// calculateVWAP calculates the volume-weighted average price
func (p *priceFeedImpl) calculateVWAP(key string, window time.Duration) sdkmath.LegacyDec {
	p.history.mu.RLock()
	entries := p.history.entries[key]
	p.history.mu.RUnlock()

	if len(entries) == 0 {
		return sdkmath.LegacyZeroDec()
	}

	cutoff := time.Now().Add(-window)
	var sumPV sdkmath.LegacyDec = sdkmath.LegacyZeroDec()
	var sumV sdkmath.Int = sdkmath.ZeroInt()

	for _, e := range entries {
		if e.timestamp.After(cutoff) && e.volume.IsPositive() {
			pv := e.price.MulInt(e.volume)
			sumPV = sumPV.Add(pv)
			sumV = sumV.Add(e.volume)
		}
	}

	if sumV.IsZero() {
		return sdkmath.LegacyZeroDec()
	}

	return sumPV.QuoInt(sumV)
}

// notifySubscribers notifies price subscribers
//
//nolint:unused // Reserved for subscription-based price notifications
func (p *priceFeedImpl) notifySubscribers(key string, price Price) {
	p.subsMu.RLock()
	callbacks := p.subs[key]
	p.subsMu.RUnlock()

	for _, cb := range callbacks {
		callback := cb // capture loop variable
		verrors.SafeGo("price-notify", func() {
			callback(price)
		})
	}
}

// runUpdater runs the background price updater
func (p *priceFeedImpl) runUpdater(ctx context.Context, stopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(p.cfg.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-stopCh:
			return
		case <-ticker.C:
			p.cleanupHistory()
		}
	}
}

// cleanupHistory removes old price history entries
func (p *priceFeedImpl) cleanupHistory() {
	p.history.mu.Lock()
	defer p.history.mu.Unlock()

	cutoff := time.Now().Add(-p.history.maxAge)

	for key, entries := range p.history.entries {
		var filtered []priceEntry
		for _, e := range entries {
			if e.timestamp.After(cutoff) {
				filtered = append(filtered, e)
			}
		}
		if len(filtered) > 0 {
			p.history.entries[key] = filtered
		} else {
			delete(p.history.entries, key)
		}
	}
}

// ============================================================================
// Price Cache
// ============================================================================

func (c *priceCache) get(key string) (Price, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	price, ok := c.prices[key]
	if !ok {
		return Price{}, false
	}
	if price.IsStale(c.ttl) {
		return Price{}, false
	}
	return price, true
}

func (c *priceCache) set(key string, price Price) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.prices[key] = price
}

// ============================================================================
// Price History
// ============================================================================

func (h *priceHistory) add(key string, price sdkmath.LegacyDec, volume sdkmath.Int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.entries[key] = append(h.entries[key], priceEntry{
		price:     price,
		volume:    volume,
		timestamp: time.Now().UTC(),
	})
}

// ============================================================================
// Price Subscription
// ============================================================================

type priceSubscription struct {
	feed *priceFeedImpl
	key  string
	idx  int
}

func (s *priceSubscription) Unsubscribe() {
	s.feed.subsMu.Lock()
	defer s.feed.subsMu.Unlock()

	callbacks := s.feed.subs[s.key]
	if s.idx < len(callbacks) {
		// Remove callback by setting to nil (we don't reorder to preserve indices)
		callbacks[s.idx] = nil
	}
}

// ============================================================================
// Adapter Price Source
// ============================================================================

type adapterPriceSource struct {
	adapter Adapter
}

func (s *adapterPriceSource) Name() string {
	return s.adapter.Name()
}

func (s *adapterPriceSource) GetPrice(ctx context.Context, baseSymbol, quoteSymbol string) (Price, error) {
	return s.adapter.GetPrice(ctx, baseSymbol, quoteSymbol)
}

func (s *adapterPriceSource) IsHealthy(ctx context.Context) bool {
	return s.adapter.IsHealthy(ctx)
}

