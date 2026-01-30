package keeper

import (
	"context"
	"fmt"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/pkg/cache"
	mv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	types "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
)

// ============================================================================
// Market Cache Layer
// PERF-003: Caching Layer Implementation
// ============================================================================

// MarketCache provides caching for frequently accessed market data.
// It wraps the keeper and caches orders, bids, and leases.
type MarketCache struct {
	keeper *Keeper

	// Order cache: orderID string -> Order
	orderCache cache.LoadingCache[string, cachedOrder]

	// Bid cache: bidID string -> Bid
	bidCache cache.LoadingCache[string, cachedBid]

	// Lease cache: leaseID string -> Lease
	leaseCache cache.LoadingCache[string, cachedLease]

	// Configuration
	config MarketCacheConfig

	// Metrics
	metrics *cache.Metrics

	// State
	mu     sync.RWMutex
	closed bool
}

// cachedOrder holds a cached order with found status.
type cachedOrder struct {
	Order types.Order
	Found bool
}

// cachedBid holds a cached bid with found status.
type cachedBid struct {
	Bid   types.Bid
	Found bool
}

// cachedLease holds a cached lease with found status.
type cachedLease struct {
	Lease mv1.Lease
	Found bool
}

// MarketCacheConfig configures the market cache layer.
type MarketCacheConfig struct {
	// Enabled determines if caching is enabled.
	Enabled bool

	// OrderTTL is the TTL for cached orders.
	OrderTTL time.Duration

	// BidTTL is the TTL for cached bids.
	BidTTL time.Duration

	// LeaseTTL is the TTL for cached leases.
	LeaseTTL time.Duration

	// MaxOrderEntries is the maximum number of cached orders.
	MaxOrderEntries int

	// MaxBidEntries is the maximum number of cached bids.
	MaxBidEntries int

	// MaxLeaseEntries is the maximum number of cached leases.
	MaxLeaseEntries int

	// EnableMetrics enables cache metrics collection.
	EnableMetrics bool
}

// DefaultMarketCacheConfig returns the default market cache configuration.
func DefaultMarketCacheConfig() MarketCacheConfig {
	return MarketCacheConfig{
		Enabled:         true,
		OrderTTL:        1 * time.Minute,  // Orders are relatively dynamic
		BidTTL:          30 * time.Second, // Bids change frequently
		LeaseTTL:        2 * time.Minute,  // Leases are more stable
		MaxOrderEntries: 5000,
		MaxBidEntries:   10000,
		MaxLeaseEntries: 5000,
		EnableMetrics:   true,
	}
}

// NewMarketCache creates a new market cache wrapper around a keeper.
func NewMarketCache(keeper *Keeper, config MarketCacheConfig) *MarketCache {
	mc := &MarketCache{
		keeper: keeper,
		config: config,
	}

	if !config.Enabled {
		return mc
	}

	// Create order cache
	mc.orderCache = cache.NewMemoryCache(
		cache.WithMaxSize[string, cachedOrder](config.MaxOrderEntries),
		cache.WithDefaultTTL[string, cachedOrder](config.OrderTTL),
	)

	// Create bid cache
	mc.bidCache = cache.NewMemoryCache(
		cache.WithMaxSize[string, cachedBid](config.MaxBidEntries),
		cache.WithDefaultTTL[string, cachedBid](config.BidTTL),
	)

	// Create lease cache
	mc.leaseCache = cache.NewMemoryCache(
		cache.WithMaxSize[string, cachedLease](config.MaxLeaseEntries),
		cache.WithDefaultTTL[string, cachedLease](config.LeaseTTL),
	)

	// Setup metrics
	if config.EnableMetrics {
		mc.metrics = cache.NewMetrics("market_cache", 1*time.Minute)
		mc.metrics.RegisterCache("orders", mc.orderCache.(cache.MetricsProvider))
		mc.metrics.RegisterCache("bids", mc.bidCache.(cache.MetricsProvider))
		mc.metrics.RegisterCache("leases", mc.leaseCache.(cache.MetricsProvider))
	}

	return mc
}

// GetOrder returns an order by ID, using cache if available.
func (mc *MarketCache) GetOrder(ctx sdk.Context, id mv1.OrderID) (types.Order, bool) {
	if !mc.config.Enabled || mc.orderCache == nil {
		return mc.keeper.GetOrder(ctx, id)
	}

	key := id.String()
	bgCtx := context.Background()

	// Try to get from cache with loader
	cached, err := mc.orderCache.GetOrLoad(bgCtx, key, func(_ context.Context, k string) (cachedOrder, error) {
		order, found := mc.keeper.GetOrder(ctx, id)
		return cachedOrder{Order: order, Found: found}, nil
	})

	if err != nil {
		return mc.keeper.GetOrder(ctx, id)
	}

	return cached.Order, cached.Found
}

// GetBid returns a bid by ID, using cache if available.
func (mc *MarketCache) GetBid(ctx sdk.Context, id mv1.BidID) (types.Bid, bool) {
	if !mc.config.Enabled || mc.bidCache == nil {
		return mc.keeper.GetBid(ctx, id)
	}

	key := id.String()
	bgCtx := context.Background()

	// Try to get from cache with loader
	cached, err := mc.bidCache.GetOrLoad(bgCtx, key, func(_ context.Context, k string) (cachedBid, error) {
		bid, found := mc.keeper.GetBid(ctx, id)
		return cachedBid{Bid: bid, Found: found}, nil
	})

	if err != nil {
		return mc.keeper.GetBid(ctx, id)
	}

	return cached.Bid, cached.Found
}

// GetLease returns a lease by ID, using cache if available.
func (mc *MarketCache) GetLease(ctx sdk.Context, id mv1.LeaseID) (mv1.Lease, bool) {
	if !mc.config.Enabled || mc.leaseCache == nil {
		return mc.keeper.GetLease(ctx, id)
	}

	key := id.String()
	bgCtx := context.Background()

	// Try to get from cache with loader
	cached, err := mc.leaseCache.GetOrLoad(bgCtx, key, func(_ context.Context, k string) (cachedLease, error) {
		lease, found := mc.keeper.GetLease(ctx, id)
		return cachedLease{Lease: lease, Found: found}, nil
	})

	if err != nil {
		return mc.keeper.GetLease(ctx, id)
	}

	return cached.Lease, cached.Found
}

// InvalidateOrder removes an order from the cache.
func (mc *MarketCache) InvalidateOrder(id mv1.OrderID) {
	if !mc.config.Enabled || mc.orderCache == nil {
		return
	}
	_ = mc.orderCache.Delete(context.Background(), id.String())
}

// InvalidateBid removes a bid from the cache.
func (mc *MarketCache) InvalidateBid(id mv1.BidID) {
	if !mc.config.Enabled || mc.bidCache == nil {
		return
	}
	_ = mc.bidCache.Delete(context.Background(), id.String())
}

// InvalidateLease removes a lease from the cache.
func (mc *MarketCache) InvalidateLease(id mv1.LeaseID) {
	if !mc.config.Enabled || mc.leaseCache == nil {
		return
	}
	_ = mc.leaseCache.Delete(context.Background(), id.String())
}

// InvalidateAll clears all caches.
func (mc *MarketCache) InvalidateAll() {
	if !mc.config.Enabled {
		return
	}

	bgCtx := context.Background()
	if mc.orderCache != nil {
		_ = mc.orderCache.Clear(bgCtx)
	}
	if mc.bidCache != nil {
		_ = mc.bidCache.Clear(bgCtx)
	}
	if mc.leaseCache != nil {
		_ = mc.leaseCache.Clear(bgCtx)
	}
}

// Stats returns cache statistics.
func (mc *MarketCache) Stats() MarketCacheStats {
	stats := MarketCacheStats{
		Enabled: mc.config.Enabled,
	}

	if !mc.config.Enabled || mc.metrics == nil {
		return stats
	}

	if orderStats, ok := mc.metrics.GetCacheStats("orders"); ok {
		stats.OrderHits = orderStats.Hits
		stats.OrderMisses = orderStats.Misses
		stats.OrderSize = orderStats.Size
		stats.OrderHitRate = orderStats.HitRate()
	}

	if bidStats, ok := mc.metrics.GetCacheStats("bids"); ok {
		stats.BidHits = bidStats.Hits
		stats.BidMisses = bidStats.Misses
		stats.BidSize = bidStats.Size
		stats.BidHitRate = bidStats.HitRate()
	}

	if leaseStats, ok := mc.metrics.GetCacheStats("leases"); ok {
		stats.LeaseHits = leaseStats.Hits
		stats.LeaseMisses = leaseStats.Misses
		stats.LeaseSize = leaseStats.Size
		stats.LeaseHitRate = leaseStats.HitRate()
	}

	return stats
}

// Close releases cache resources.
func (mc *MarketCache) Close() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.closed {
		return nil
	}
	mc.closed = true

	if mc.metrics != nil {
		mc.metrics.Stop()
	}

	var errs []error
	if mc.orderCache != nil {
		if err := mc.orderCache.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if mc.bidCache != nil {
		if err := mc.bidCache.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if mc.leaseCache != nil {
		if err := mc.leaseCache.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("market cache close errors: %v", errs)
	}
	return nil
}

// MarketCacheStats contains market cache statistics.
type MarketCacheStats struct {
	Enabled      bool    `json:"enabled"`
	OrderHits    uint64  `json:"order_hits"`
	OrderMisses  uint64  `json:"order_misses"`
	OrderSize    int     `json:"order_size"`
	OrderHitRate float64 `json:"order_hit_rate"`
	BidHits      uint64  `json:"bid_hits"`
	BidMisses    uint64  `json:"bid_misses"`
	BidSize      int     `json:"bid_size"`
	BidHitRate   float64 `json:"bid_hit_rate"`
	LeaseHits    uint64  `json:"lease_hits"`
	LeaseMisses  uint64  `json:"lease_misses"`
	LeaseSize    int     `json:"lease_size"`
	LeaseHitRate float64 `json:"lease_hit_rate"`
}

// ============================================================================
// Cached Market Keeper Wrapper
// ============================================================================

// CachedMarketKeeper wraps a Market Keeper with caching support.
type CachedMarketKeeper struct {
	*Keeper
	cache *MarketCache
}

// NewCachedMarketKeeper creates a new cached market keeper wrapper.
func NewCachedMarketKeeper(keeper *Keeper, config MarketCacheConfig) *CachedMarketKeeper {
	return &CachedMarketKeeper{
		Keeper: keeper,
		cache:  NewMarketCache(keeper, config),
	}
}

// GetOrder returns an order using cache.
func (ck *CachedMarketKeeper) GetOrder(ctx sdk.Context, id mv1.OrderID) (types.Order, bool) {
	return ck.cache.GetOrder(ctx, id)
}

// GetBid returns a bid using cache.
func (ck *CachedMarketKeeper) GetBid(ctx sdk.Context, id mv1.BidID) (types.Bid, bool) {
	return ck.cache.GetBid(ctx, id)
}

// GetLease returns a lease using cache.
func (ck *CachedMarketKeeper) GetLease(ctx sdk.Context, id mv1.LeaseID) (mv1.Lease, bool) {
	return ck.cache.GetLease(ctx, id)
}

// OnOrderMatched updates order state and invalidates cache.
func (ck *CachedMarketKeeper) OnOrderMatched(ctx sdk.Context, order types.Order) {
	ck.Keeper.OnOrderMatched(ctx, order)
	ck.cache.InvalidateOrder(order.ID)
}

// OnBidMatched updates bid state and invalidates cache.
func (ck *CachedMarketKeeper) OnBidMatched(ctx sdk.Context, bid types.Bid) {
	ck.Keeper.OnBidMatched(ctx, bid)
	ck.cache.InvalidateBid(bid.ID)
}

// OnBidLost updates bid state and invalidates cache.
func (ck *CachedMarketKeeper) OnBidLost(ctx sdk.Context, bid types.Bid) {
	ck.Keeper.OnBidLost(ctx, bid)
	ck.cache.InvalidateBid(bid.ID)
}

// OnBidClosed updates bid state and invalidates cache.
func (ck *CachedMarketKeeper) OnBidClosed(ctx sdk.Context, bid types.Bid) error {
	err := ck.Keeper.OnBidClosed(ctx, bid)
	ck.cache.InvalidateBid(bid.ID)
	return err
}

// OnOrderClosed updates order state and invalidates cache.
func (ck *CachedMarketKeeper) OnOrderClosed(ctx sdk.Context, order types.Order) error {
	err := ck.Keeper.OnOrderClosed(ctx, order)
	ck.cache.InvalidateOrder(order.ID)
	return err
}

// OnLeaseClosed updates lease state and invalidates cache.
func (ck *CachedMarketKeeper) OnLeaseClosed(ctx sdk.Context, lease mv1.Lease, state mv1.Lease_State, reason mv1.LeaseClosedReason) error {
	err := ck.Keeper.OnLeaseClosed(ctx, lease, state, reason)
	ck.cache.InvalidateLease(lease.ID)
	return err
}

// CreateLease creates a lease and invalidates related caches.
func (ck *CachedMarketKeeper) CreateLease(ctx sdk.Context, bid types.Bid) error {
	err := ck.Keeper.CreateLease(ctx, bid)
	if err == nil {
		ck.cache.InvalidateBid(bid.ID)
		// Invalidate order cache as well since order state changes
		ck.cache.InvalidateOrder(mv1.OrderID(bid.ID))
	}
	return err
}

// CacheStats returns cache statistics.
func (ck *CachedMarketKeeper) CacheStats() MarketCacheStats {
	return ck.cache.Stats()
}

// Close closes the cache resources.
func (ck *CachedMarketKeeper) Close() error {
	return ck.cache.Close()
}
