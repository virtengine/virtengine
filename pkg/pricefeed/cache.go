package pricefeed

import (
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================================
// In-Memory Cache Implementation
// ============================================================================

// cacheEntry represents a cached price with metadata
type cacheEntry struct {
	Price     PriceData
	CachedAt  time.Time
	ExpiresAt time.Time
}

// IsExpired checks if the entry has expired
func (e cacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// InMemoryCache is a thread-safe in-memory cache for price data
type InMemoryCache struct {
	entries      map[string]cacheEntry
	mu           sync.RWMutex
	ttl          time.Duration
	maxSize      int
	allowStale   bool
	staleMaxAge  time.Duration
	hits         atomic.Int64
	misses       atomic.Int64
	staleCount   atomic.Int64
	evictedCount atomic.Int64
}

// NewInMemoryCache creates a new in-memory cache
func NewInMemoryCache(cfg CacheConfig) *InMemoryCache {
	c := &InMemoryCache{
		entries:     make(map[string]cacheEntry),
		ttl:         cfg.TTL,
		maxSize:     cfg.MaxSize,
		allowStale:  cfg.AllowStale,
		staleMaxAge: cfg.StaleMaxAge,
	}

	// Start background cleanup goroutine
	go c.cleanupLoop()

	return c
}

// cacheKey generates a cache key for an asset pair
func cacheKey(baseAsset, quoteAsset string) string {
	return baseAsset + "/" + quoteAsset
}

// Get retrieves a cached price
func (c *InMemoryCache) Get(key string) (PriceData, bool) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()

	if !ok {
		c.misses.Add(1)
		return PriceData{}, false
	}

	if entry.IsExpired() {
		// Check if stale data is acceptable
		if c.allowStale && time.Since(entry.CachedAt) <= c.staleMaxAge {
			c.staleCount.Add(1)
			c.hits.Add(1)
			return entry.Price, true
		}
		c.misses.Add(1)
		return PriceData{}, false
	}

	c.hits.Add(1)
	return entry.Price, true
}

// Set stores a price in the cache
func (c *InMemoryCache) Set(key string, price PriceData) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict oldest entries if at capacity
	if len(c.entries) >= c.maxSize {
		c.evictOldest()
	}

	c.entries[key] = cacheEntry{
		Price:     price,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// Delete removes a price from the cache
func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}

// Clear clears all cached prices
func (c *InMemoryCache) Clear() {
	c.mu.Lock()
	c.entries = make(map[string]cacheEntry)
	c.mu.Unlock()
}

// Stats returns cache statistics
func (c *InMemoryCache) Stats() CacheStats {
	c.mu.RLock()
	size := len(c.entries)
	c.mu.RUnlock()

	return CacheStats{
		Hits:         c.hits.Load(),
		Misses:       c.misses.Load(),
		Size:         size,
		EvictedCount: c.evictedCount.Load(),
		StaleCount:   c.staleCount.Load(),
	}
}

// evictOldest removes the oldest entry (must be called with lock held)
func (c *InMemoryCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.entries {
		if oldestKey == "" || entry.CachedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.CachedAt
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
		c.evictedCount.Add(1)
	}
}

// cleanupLoop periodically removes expired entries
func (c *InMemoryCache) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup removes all expired entries
func (c *InMemoryCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		// Remove entries that are beyond stale max age
		if now.Sub(entry.CachedAt) > c.staleMaxAge {
			delete(c.entries, key)
			c.evictedCount.Add(1)
		}
	}
}

// GetStale retrieves a stale price if available (for fallback scenarios)
func (c *InMemoryCache) GetStale(key string) (PriceData, bool) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()

	if !ok {
		return PriceData{}, false
	}

	// Return stale data regardless of expiry, up to staleMaxAge
	if time.Since(entry.CachedAt) <= c.staleMaxAge {
		c.staleCount.Add(1)
		return entry.Price, true
	}

	return PriceData{}, false
}

// GetEntry returns the full cache entry for debugging
func (c *InMemoryCache) GetEntry(key string) (cacheEntry, bool) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	return entry, ok
}
