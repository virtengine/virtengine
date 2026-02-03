// Package cache provides a generic caching layer for VirtEngine.
//
// PERF-003: Caching Layer Implementation
//
// This package provides:
// - Generic cache interface with TTL support
// - In-memory LRU cache implementation
// - Redis distributed cache for multi-validator setups
// - Cache hit rate monitoring and metrics
// - Automatic invalidation strategies
package cache

import (
	"context"
	"errors"
	"time"
)

// Common errors
var (
	ErrCacheMiss     = errors.New("cache: key not found")
	ErrCacheClosed   = errors.New("cache: cache is closed")
	ErrInvalidTTL    = errors.New("cache: invalid TTL")
	ErrKeyEmpty      = errors.New("cache: key cannot be empty")
	ErrValueNil      = errors.New("cache: value cannot be nil")
	ErrNotSupported  = errors.New("cache: operation not supported")
	ErrSerialization = errors.New("cache: serialization error")
)

// Cache defines the interface for a generic cache implementation.
// All implementations must be safe for concurrent access.
type Cache[K comparable, V any] interface {
	// Get retrieves a value from the cache.
	// Returns ErrCacheMiss if the key is not found or expired.
	Get(ctx context.Context, key K) (V, error)

	// Set stores a value in the cache with the default TTL.
	Set(ctx context.Context, key K, value V) error

	// SetWithTTL stores a value in the cache with a specific TTL.
	// A TTL of 0 means the entry never expires.
	SetWithTTL(ctx context.Context, key K, value V, ttl time.Duration) error

	// Delete removes a value from the cache.
	Delete(ctx context.Context, key K) error

	// Exists checks if a key exists in the cache.
	Exists(ctx context.Context, key K) bool

	// Clear removes all entries from the cache.
	Clear(ctx context.Context) error

	// Size returns the number of items in the cache.
	Size(ctx context.Context) int

	// Close releases any resources used by the cache.
	Close() error
}

// CacheWithMetrics extends Cache with metrics collection capabilities.
type CacheWithMetrics[K comparable, V any] interface {
	Cache[K, V]

	// Stats returns current cache statistics.
	Stats() CacheStats
}

// CacheStats contains cache performance metrics.
type CacheStats struct {
	// Hits is the number of cache hits.
	Hits uint64 `json:"hits"`

	// Misses is the number of cache misses.
	Misses uint64 `json:"misses"`

	// Evictions is the number of entries evicted due to capacity.
	Evictions uint64 `json:"evictions"`

	// Expirations is the number of entries expired due to TTL.
	Expirations uint64 `json:"expirations"`

	// Size is the current number of items in the cache.
	Size int `json:"size"`

	// MaxSize is the maximum capacity of the cache.
	MaxSize int `json:"max_size"`

	// CreatedAt is when the cache was created.
	CreatedAt time.Time `json:"created_at"`

	// LastAccessedAt is the last time the cache was accessed.
	LastAccessedAt time.Time `json:"last_accessed_at"`
}

// HitRate returns the cache hit rate as a percentage (0-100).
func (s CacheStats) HitRate() float64 {
	total := s.Hits + s.Misses
	if total == 0 {
		return 0
	}
	return float64(s.Hits) / float64(total) * 100
}

// CacheEntry represents a single cache entry with metadata.
type CacheEntry[V any] struct {
	// Value is the cached value.
	Value V

	// CreatedAt is when the entry was created.
	CreatedAt time.Time

	// ExpiresAt is when the entry expires (zero means never).
	ExpiresAt time.Time

	// AccessCount is the number of times this entry was accessed.
	AccessCount uint64
}

// IsExpired returns true if the entry has expired.
func (e CacheEntry[V]) IsExpired() bool {
	if e.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(e.ExpiresAt)
}

// TTL returns the remaining time-to-live for the entry.
func (e CacheEntry[V]) TTL() time.Duration {
	if e.ExpiresAt.IsZero() {
		return 0 // Never expires
	}
	remaining := time.Until(e.ExpiresAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Serializer defines the interface for serializing/deserializing cache values.
// Used by distributed cache backends like Redis.
type Serializer[V any] interface {
	// Serialize converts a value to bytes.
	Serialize(v V) ([]byte, error)

	// Deserialize converts bytes back to a value.
	Deserialize(data []byte) (V, error)
}

// KeyFormatter defines the interface for formatting cache keys.
// Useful for adding prefixes or namespacing keys.
type KeyFormatter[K comparable] interface {
	// Format converts a key to its string representation.
	Format(key K) string

	// Parse converts a string back to a key.
	Parse(s string) (K, error)
}

// StringKeyFormatter is a simple formatter for string keys.
type StringKeyFormatter struct {
	Prefix string
}

// Format implements KeyFormatter.
func (f StringKeyFormatter) Format(key string) string {
	if f.Prefix == "" {
		return key
	}
	return f.Prefix + ":" + key
}

// Parse implements KeyFormatter.
func (f StringKeyFormatter) Parse(s string) (string, error) {
	if f.Prefix == "" {
		return s, nil
	}
	if len(s) <= len(f.Prefix)+1 {
		return "", errors.New("invalid formatted key")
	}
	return s[len(f.Prefix)+1:], nil
}

// CacheLoader is a function type for loading values into cache on miss.
type CacheLoader[K comparable, V any] func(ctx context.Context, key K) (V, error)

// LoadingCache extends Cache with automatic value loading on cache miss.
type LoadingCache[K comparable, V any] interface {
	Cache[K, V]

	// GetOrLoad retrieves a value from cache, or loads it using the loader on miss.
	GetOrLoad(ctx context.Context, key K, loader CacheLoader[K, V]) (V, error)
}

// InvalidationEvent represents a cache invalidation event.
type InvalidationEvent struct {
	// Key is the cache key being invalidated.
	Key string `json:"key"`

	// Pattern is the pattern for bulk invalidation (optional).
	Pattern string `json:"pattern,omitempty"`

	// Reason describes why the cache was invalidated.
	Reason string `json:"reason"`

	// Timestamp is when the invalidation occurred.
	Timestamp time.Time `json:"timestamp"`
}

// InvalidationListener is a callback for cache invalidation events.
type InvalidationListener func(event InvalidationEvent)

// ObservableCache extends Cache with invalidation event notifications.
type ObservableCache[K comparable, V any] interface {
	Cache[K, V]

	// OnInvalidate registers a listener for invalidation events.
	OnInvalidate(listener InvalidationListener)
}
