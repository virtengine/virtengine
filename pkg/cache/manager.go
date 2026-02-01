package cache

import (
	"context"
	"fmt"
	"sync"
)

// Manager manages multiple cache instances and provides a centralized
// way to create, access, and monitor caches.
type Manager struct {
	// Configuration
	config Config

	// Caches
	mu     sync.RWMutex
	caches map[string]interface{} // name -> cache instance

	// Metrics
	metrics *Metrics

	// Redis client (shared across all Redis caches)
	redisClient RedisClient
}

// NewManager creates a new cache manager.
func NewManager(config Config) (*Manager, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	m := &Manager{
		config: config,
		caches: make(map[string]interface{}),
	}

	if config.Metrics.Enabled {
		m.metrics = NewMetrics(config.Metrics.Namespace, config.Metrics.ReportInterval)
	}

	return m, nil
}

// SetRedisClient sets the Redis client for distributed caches.
func (m *Manager) SetRedisClient(client RedisClient) {
	m.redisClient = client
}

// GetOrCreateMemoryCache gets or creates an in-memory cache with the given name.
func GetOrCreateMemoryCache[K comparable, V any](m *Manager, name string, opts ...MemoryCacheOption[K, V]) *MemoryCache[K, V] {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if cache exists
	if existing, ok := m.caches[name]; ok {
		if cache, ok := existing.(*MemoryCache[K, V]); ok {
			return cache
		}
	}

	// Create new cache with config defaults
	defaultOpts := []MemoryCacheOption[K, V]{
		WithMaxSize[K, V](m.config.MaxSize),
		WithDefaultTTL[K, V](m.config.DefaultTTL),
	}
	opts = append(defaultOpts, opts...)

	cache := NewMemoryCache(opts...)
	m.caches[name] = cache

	// Register with metrics
	if m.metrics != nil {
		m.metrics.RegisterCache(name, cache)
	}

	return cache
}

// GetOrCreateRedisCache gets or creates a Redis cache with the given name.
func GetOrCreateRedisCache[K comparable, V any](m *Manager, name string, opts ...RedisCacheOption[K, V]) (*RedisCache[K, V], error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if cache exists
	if existing, ok := m.caches[name]; ok {
		if cache, ok := existing.(*RedisCache[K, V]); ok {
			return cache, nil
		}
	}

	// Ensure Redis client is available
	if m.redisClient == nil {
		return nil, fmt.Errorf("cache: Redis client not configured")
	}

	// Create Redis config with cache-specific prefix
	redisConfig := m.config.Redis
	redisConfig.KeyPrefix = m.config.Redis.KeyPrefix + name + ":"

	// Create new cache
	defaultOpts := []RedisCacheOption[K, V]{
		WithRedisDefaultTTL[K, V](m.config.DefaultTTL),
	}
	opts = append(defaultOpts, opts...)

	cache, err := NewRedisCache(m.redisClient, redisConfig, opts...)
	if err != nil {
		return nil, err
	}

	m.caches[name] = cache

	// Register with metrics
	if m.metrics != nil {
		m.metrics.RegisterCache(name, cache)
	}

	return cache, nil
}

// GetMetrics returns the metrics collector.
func (m *Manager) GetMetrics() *Metrics {
	return m.metrics
}

// GetMetricsReport returns a current metrics report for all caches.
func (m *Manager) GetMetricsReport() MetricsReport {
	if m.metrics == nil {
		return MetricsReport{}
	}
	return m.metrics.GetReport()
}

// StartMetrics starts periodic metrics reporting.
func (m *Manager) StartMetrics() {
	if m.metrics != nil {
		m.metrics.Start()
	}
}

// StopMetrics stops periodic metrics reporting.
func (m *Manager) StopMetrics() {
	if m.metrics != nil {
		m.metrics.Stop()
	}
}

// Close closes all caches and releases resources.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.StopMetrics()

	var errs []error
	for name, cache := range m.caches {
		if closer, ok := cache.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				errs = append(errs, fmt.Errorf("cache %s: %w", name, err))
			}
		}
	}

	m.caches = make(map[string]interface{})

	if len(errs) > 0 {
		return fmt.Errorf("cache manager close errors: %v", errs)
	}

	return nil
}

// InvalidateAll clears all caches.
func (m *Manager) InvalidateAll(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errs []error
	for name, cache := range m.caches {
		if clearer, ok := cache.(interface{ Clear(context.Context) error }); ok {
			if err := clearer.Clear(ctx); err != nil {
				errs = append(errs, fmt.Errorf("cache %s: %w", name, err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("cache invalidation errors: %v", errs)
	}

	return nil
}

// Size returns the total number of items across all caches.
func (m *Manager) Size(ctx context.Context) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := 0
	for _, cache := range m.caches {
		if sizer, ok := cache.(interface{ Size(context.Context) int }); ok {
			total += sizer.Size(ctx)
		}
	}

	return total
}

// CacheNames returns the names of all registered caches.
func (m *Manager) CacheNames() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.caches))
	for name := range m.caches {
		names = append(names, name)
	}
	return names
}

// IsEnabled returns whether caching is enabled.
func (m *Manager) IsEnabled() bool {
	return m.config.Enabled
}

// Config returns the cache configuration.
func (m *Manager) Config() Config {
	return m.config
}

// Global cache manager instance
var (
	globalManager     *Manager
	globalManagerOnce sync.Once
	globalManagerMu   sync.Mutex
)

// GetGlobalManager returns the global cache manager instance.
// It is lazily initialized with default configuration.
func GetGlobalManager() *Manager {
	globalManagerOnce.Do(func() {
		var err error
		globalManager, err = NewManager(DefaultConfig())
		if err != nil {
			panic(fmt.Sprintf("failed to create global cache manager: %v", err))
		}
	})
	return globalManager
}

// SetGlobalManager sets the global cache manager instance.
// This should only be called during application initialization.
func SetGlobalManager(m *Manager) {
	globalManagerMu.Lock()
	defer globalManagerMu.Unlock()
	globalManager = m
}

