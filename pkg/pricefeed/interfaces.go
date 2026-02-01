package pricefeed

import (
	"context"
	"io"
)

// Provider is the interface for a price feed source
type Provider interface {
	// Name returns the unique name of this provider
	Name() string

	// Type returns the provider type (coingecko, chainlink, pyth, etc.)
	Type() SourceType

	// GetPrice fetches the current price for an asset pair
	// baseAsset: the asset being priced (e.g., "virtengine", "uve")
	// quoteAsset: the quote currency (e.g., "usd", "eur")
	GetPrice(ctx context.Context, baseAsset, quoteAsset string) (PriceData, error)

	// GetPrices fetches prices for multiple asset pairs in a single request
	// Returns a map of "baseAsset/quoteAsset" -> PriceData
	GetPrices(ctx context.Context, pairs []AssetPair) (map[string]PriceData, error)

	// IsHealthy checks if the provider is responding correctly
	IsHealthy(ctx context.Context) bool

	// Health returns detailed health information
	Health() SourceHealth

	io.Closer
}

// AssetPair represents a trading pair
type AssetPair struct {
	Base  string `json:"base"`
	Quote string `json:"quote"`
}

// String returns the pair as "base/quote"
func (p AssetPair) String() string {
	return p.Base + "/" + p.Quote
}

// Aggregator combines multiple providers and provides a unified price interface
type Aggregator interface {
	// GetPrice returns the aggregated price from configured sources
	GetPrice(ctx context.Context, baseAsset, quoteAsset string) (AggregatedPrice, error)

	// GetPrices returns aggregated prices for multiple pairs
	GetPrices(ctx context.Context, pairs []AssetPair) (map[string]AggregatedPrice, error)

	// GetPriceFromSource returns price from a specific source
	GetPriceFromSource(ctx context.Context, source string, baseAsset, quoteAsset string) (PriceData, error)

	// ListSources returns all configured price sources
	ListSources() []SourceHealth

	// AddProvider adds a new price provider
	AddProvider(provider Provider) error

	// RemoveProvider removes a price provider by name
	RemoveProvider(name string) error

	// RefreshCache forces a cache refresh for the given pairs
	RefreshCache(ctx context.Context, pairs []AssetPair) error

	io.Closer
}

// Cache provides caching for price data
type Cache interface {
	// Get retrieves a cached price
	Get(key string) (PriceData, bool)

	// Set stores a price in the cache
	Set(key string, price PriceData)

	// Delete removes a price from the cache
	Delete(key string)

	// Clear clears all cached prices
	Clear()

	// Stats returns cache statistics
	Stats() CacheStats
}

// CacheStats contains cache performance statistics
type CacheStats struct {
	// Hits is the number of cache hits
	Hits int64 `json:"hits"`

	// Misses is the number of cache misses
	Misses int64 `json:"misses"`

	// Size is the number of items in cache
	Size int `json:"size"`

	// EvictedCount is the number of items evicted
	EvictedCount int64 `json:"evicted_count"`

	// StaleCount is the number of stale items returned
	StaleCount int64 `json:"stale_count"`
}

// HitRate returns the cache hit rate as a percentage
func (s CacheStats) HitRate() float64 {
	total := s.Hits + s.Misses
	if total == 0 {
		return 0
	}
	return float64(s.Hits) / float64(total) * 100
}

// Metrics collects price feed metrics for observability
type Metrics interface {
	// RecordFetch records a price fetch operation
	RecordFetch(source string, baseAsset, quoteAsset string, latency float64, success bool)

	// RecordCacheHit records a cache hit
	RecordCacheHit(baseAsset, quoteAsset string)

	// RecordCacheMiss records a cache miss
	RecordCacheMiss(baseAsset, quoteAsset string)

	// RecordPriceDeviation records price deviation between sources
	RecordPriceDeviation(baseAsset, quoteAsset string, deviation float64)

	// RecordSourceHealth records source health status
	RecordSourceHealth(source string, healthy bool, latency float64)
}

