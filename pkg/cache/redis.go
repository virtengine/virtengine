package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// RedisCache is a distributed cache implementation using Redis.
// It provides multi-validator cache coherency for VirtEngine.
type RedisCache[K comparable, V any] struct {
	// Configuration
	config     RedisConfig
	defaultTTL time.Duration
	keyPrefix  string

	// Client interface (allows mocking)
	client RedisClient

	// Serialization
	serializer   Serializer[V]
	keyFormatter KeyFormatter[K]

	// State
	closed atomic.Bool

	// Metrics
	hits        atomic.Uint64
	misses      atomic.Uint64
	evictions   atomic.Uint64
	expirations atomic.Uint64
	createdAt   time.Time
	lastAccess  atomic.Int64

	// Invalidation
	listenersMu sync.RWMutex
	listeners   []InvalidationListener
}

// RedisClient defines the interface for Redis operations.
// This allows for easy testing with mock implementations.
type RedisClient interface {
	// Get retrieves a value by key.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value with TTL.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes a key.
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists.
	Exists(ctx context.Context, key string) (bool, error)

	// Keys returns all keys matching a pattern.
	Keys(ctx context.Context, pattern string) ([]string, error)

	// FlushDB clears all keys.
	FlushDB(ctx context.Context) error

	// DBSize returns the number of keys.
	DBSize(ctx context.Context) (int64, error)

	// Ping tests connectivity.
	Ping(ctx context.Context) error

	// Close closes the connection.
	Close() error
}

// RedisCacheOption is a functional option for configuring RedisCache.
type RedisCacheOption[K comparable, V any] func(*RedisCache[K, V])

// WithRedisSerializer sets a custom serializer.
func WithRedisSerializer[K comparable, V any](s Serializer[V]) RedisCacheOption[K, V] {
	return func(c *RedisCache[K, V]) {
		c.serializer = s
	}
}

// WithRedisKeyFormatter sets a custom key formatter.
func WithRedisKeyFormatter[K comparable, V any](f KeyFormatter[K]) RedisCacheOption[K, V] {
	return func(c *RedisCache[K, V]) {
		c.keyFormatter = f
	}
}

// WithRedisDefaultTTL sets the default TTL.
func WithRedisDefaultTTL[K comparable, V any](ttl time.Duration) RedisCacheOption[K, V] {
	return func(c *RedisCache[K, V]) {
		c.defaultTTL = ttl
	}
}

// NewRedisCache creates a new Redis-based cache.
func NewRedisCache[K comparable, V any](client RedisClient, config RedisConfig, opts ...RedisCacheOption[K, V]) (*RedisCache[K, V], error) {
	if client == nil {
		return nil, errors.New("cache: Redis client is required")
	}

	c := &RedisCache[K, V]{
		config:     config,
		client:     client,
		defaultTTL: 5 * time.Minute,
		keyPrefix:  config.KeyPrefix,
		createdAt:  time.Now(),
	}

	for _, opt := range opts {
		opt(c)
	}

	// Set default serializer if not provided
	if c.serializer == nil {
		c.serializer = &JSONSerializer[V]{}
	}

	// Test connectivity
	ctx, cancel := context.WithTimeout(context.Background(), config.DialTimeout)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		return nil, fmt.Errorf("cache: failed to connect to Redis: %w", err)
	}

	return c, nil
}

// Get retrieves a value from Redis.
func (c *RedisCache[K, V]) Get(ctx context.Context, key K) (V, error) {
	var zero V

	if c.closed.Load() {
		return zero, ErrCacheClosed
	}

	c.lastAccess.Store(time.Now().UnixNano())

	keyStr := c.formatKey(key)
	data, err := c.client.Get(ctx, keyStr)
	if err != nil {
		c.misses.Add(1)
		if isNotFound(err) {
			return zero, ErrCacheMiss
		}
		return zero, fmt.Errorf("cache: Redis get failed: %w", err)
	}

	value, err := c.serializer.Deserialize(data)
	if err != nil {
		c.misses.Add(1)
		return zero, fmt.Errorf("%w: %v", ErrSerialization, err)
	}

	c.hits.Add(1)
	return value, nil
}

// Set stores a value with the default TTL.
func (c *RedisCache[K, V]) Set(ctx context.Context, key K, value V) error {
	return c.SetWithTTL(ctx, key, value, c.defaultTTL)
}

// SetWithTTL stores a value with a specific TTL.
func (c *RedisCache[K, V]) SetWithTTL(ctx context.Context, key K, value V, ttl time.Duration) error {
	if c.closed.Load() {
		return ErrCacheClosed
	}

	c.lastAccess.Store(time.Now().UnixNano())

	data, err := c.serializer.Serialize(value)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSerialization, err)
	}

	keyStr := c.formatKey(key)
	if err := c.client.Set(ctx, keyStr, data, ttl); err != nil {
		return fmt.Errorf("cache: Redis set failed: %w", err)
	}

	return nil
}

// Delete removes a value from Redis.
func (c *RedisCache[K, V]) Delete(ctx context.Context, key K) error {
	if c.closed.Load() {
		return ErrCacheClosed
	}

	keyStr := c.formatKey(key)
	if err := c.client.Delete(ctx, keyStr); err != nil {
		return fmt.Errorf("cache: Redis delete failed: %w", err)
	}

	c.notifyInvalidation(keyStr, "explicit delete")
	return nil
}

// Exists checks if a key exists in Redis.
func (c *RedisCache[K, V]) Exists(ctx context.Context, key K) bool {
	if c.closed.Load() {
		return false
	}

	keyStr := c.formatKey(key)
	exists, err := c.client.Exists(ctx, keyStr)
	if err != nil {
		return false
	}

	return exists
}

// Clear removes all entries with the cache prefix.
func (c *RedisCache[K, V]) Clear(ctx context.Context) error {
	if c.closed.Load() {
		return ErrCacheClosed
	}

	// Get all keys with our prefix
	pattern := c.keyPrefix + "*"
	keys, err := c.client.Keys(ctx, pattern)
	if err != nil {
		return fmt.Errorf("cache: Redis keys failed: %w", err)
	}

	// Delete all matching keys
	for _, key := range keys {
		_ = c.client.Delete(ctx, key)
	}

	c.notifyInvalidation("", "cache cleared")
	return nil
}

// Size returns the approximate number of keys with our prefix.
func (c *RedisCache[K, V]) Size(ctx context.Context) int {
	if c.closed.Load() {
		return 0
	}

	pattern := c.keyPrefix + "*"
	keys, err := c.client.Keys(ctx, pattern)
	if err != nil {
		return 0
	}

	return len(keys)
}

// Close closes the Redis connection.
func (c *RedisCache[K, V]) Close() error {
	if c.closed.Swap(true) {
		return nil // Already closed
	}

	return c.client.Close()
}

// Stats returns current cache statistics.
func (c *RedisCache[K, V]) Stats() CacheStats {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	size := c.Size(ctx)

	lastAccessNano := c.lastAccess.Load()
	var lastAccessTime time.Time
	if lastAccessNano > 0 {
		lastAccessTime = time.Unix(0, lastAccessNano)
	}

	return CacheStats{
		Hits:           c.hits.Load(),
		Misses:         c.misses.Load(),
		Evictions:      c.evictions.Load(),
		Expirations:    c.expirations.Load(),
		Size:           size,
		MaxSize:        -1, // Redis doesn't have a fixed max size
		CreatedAt:      c.createdAt,
		LastAccessedAt: lastAccessTime,
	}
}

// OnInvalidate registers a listener for cache invalidation events.
func (c *RedisCache[K, V]) OnInvalidate(listener InvalidationListener) {
	c.listenersMu.Lock()
	defer c.listenersMu.Unlock()
	c.listeners = append(c.listeners, listener)
}

// GetOrLoad retrieves a value or loads it using the provided loader.
func (c *RedisCache[K, V]) GetOrLoad(ctx context.Context, key K, loader CacheLoader[K, V]) (V, error) {
	// Try cache first
	if val, err := c.Get(ctx, key); err == nil {
		return val, nil
	}

	// Load value
	val, err := loader(ctx, key)
	if err != nil {
		var zero V
		return zero, err
	}

	// Store in cache
	_ = c.Set(ctx, key, val)

	return val, nil
}

// InvalidatePattern invalidates all keys matching a pattern.
func (c *RedisCache[K, V]) InvalidatePattern(ctx context.Context, pattern string) error {
	if c.closed.Load() {
		return ErrCacheClosed
	}

	fullPattern := c.keyPrefix + pattern
	keys, err := c.client.Keys(ctx, fullPattern)
	if err != nil {
		return fmt.Errorf("cache: Redis keys failed: %w", err)
	}

	for _, key := range keys {
		_ = c.client.Delete(ctx, key)
	}

	c.notifyInvalidation("", fmt.Sprintf("pattern invalidation: %s", pattern))
	return nil
}

// formatKey formats a key with the cache prefix.
func (c *RedisCache[K, V]) formatKey(key K) string {
	var keyStr string

	if c.keyFormatter != nil {
		keyStr = c.keyFormatter.Format(key)
	} else {
		// Default formatting
		switch k := any(key).(type) {
		case string:
			keyStr = k
		case []byte:
			keyStr = string(k)
		default:
			keyStr = fmt.Sprintf("%v", key)
		}
	}

	if c.keyPrefix != "" {
		return c.keyPrefix + keyStr
	}
	return keyStr
}

// notifyInvalidation notifies all listeners of a cache invalidation.
func (c *RedisCache[K, V]) notifyInvalidation(key, reason string) {
	c.listenersMu.RLock()
	listeners := make([]InvalidationListener, len(c.listeners))
	copy(listeners, c.listeners)
	c.listenersMu.RUnlock()

	if len(listeners) == 0 {
		return
	}

	event := InvalidationEvent{
		Key:       key,
		Reason:    reason,
		Timestamp: time.Now(),
	}

	for _, listener := range listeners {
		go listener(event)
	}
}

// isNotFound checks if an error indicates a key was not found.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "redis: nil" || errors.Is(err, ErrCacheMiss)
}

// JSONSerializer is a JSON-based serializer.
type JSONSerializer[V any] struct{}

// Serialize implements Serializer.
func (s *JSONSerializer[V]) Serialize(v V) ([]byte, error) {
	return json.Marshal(v)
}

// Deserialize implements Serializer.
func (s *JSONSerializer[V]) Deserialize(data []byte) (V, error) {
	var v V
	err := json.Unmarshal(data, &v)
	return v, err
}

// MockRedisClient is a mock Redis client for testing.
type MockRedisClient struct {
	mu      sync.RWMutex
	data    map[string]mockEntry
	closed  bool
	pingErr error
}

type mockEntry struct {
	value     []byte
	expiresAt time.Time
}

// NewMockRedisClient creates a new mock Redis client.
func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		data: make(map[string]mockEntry),
	}
}

// Get implements RedisClient.
func (m *MockRedisClient) Get(ctx context.Context, key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, errors.New("redis: client is closed")
	}

	entry, ok := m.data[key]
	if !ok {
		return nil, errors.New("redis: nil")
	}

	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		return nil, errors.New("redis: nil")
	}

	return entry.value, nil
}

// Set implements RedisClient.
func (m *MockRedisClient) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return errors.New("redis: client is closed")
	}

	entry := mockEntry{value: value}
	if ttl > 0 {
		entry.expiresAt = time.Now().Add(ttl)
	}

	m.data[key] = entry
	return nil
}

// Delete implements RedisClient.
func (m *MockRedisClient) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return errors.New("redis: client is closed")
	}

	delete(m.data, key)
	return nil
}

// Exists implements RedisClient.
func (m *MockRedisClient) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return false, errors.New("redis: client is closed")
	}

	entry, ok := m.data[key]
	if !ok {
		return false, nil
	}

	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		return false, nil
	}

	return true, nil
}

// Keys implements RedisClient.
func (m *MockRedisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, errors.New("redis: client is closed")
	}

	var result []string
	for key := range m.data {
		// Simple prefix matching for mock
		if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
			prefix := pattern[:len(pattern)-1]
			if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
				result = append(result, key)
			}
		} else if key == pattern {
			result = append(result, key)
		}
	}

	return result, nil
}

// FlushDB implements RedisClient.
func (m *MockRedisClient) FlushDB(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return errors.New("redis: client is closed")
	}

	m.data = make(map[string]mockEntry)
	return nil
}

// DBSize implements RedisClient.
func (m *MockRedisClient) DBSize(ctx context.Context) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return 0, errors.New("redis: client is closed")
	}

	return int64(len(m.data)), nil
}

// Ping implements RedisClient.
func (m *MockRedisClient) Ping(ctx context.Context) error {
	if m.pingErr != nil {
		return m.pingErr
	}
	return nil
}

// Close implements RedisClient.
func (m *MockRedisClient) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

// SetPingError sets an error to be returned by Ping (for testing).
func (m *MockRedisClient) SetPingError(err error) {
	m.pingErr = err
}

// Compile-time interface checks
var (
	_ Cache[string, any]            = (*RedisCache[string, any])(nil)
	_ CacheWithMetrics[string, any] = (*RedisCache[string, any])(nil)
	_ LoadingCache[string, any]     = (*RedisCache[string, any])(nil)
	_ ObservableCache[string, any]  = (*RedisCache[string, any])(nil)
	_ RedisClient                   = (*MockRedisClient)(nil)
)
