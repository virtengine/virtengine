package cache

import (
	"container/list"
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// MemoryCache is an in-memory LRU cache implementation with TTL support.
// It is safe for concurrent access.
type MemoryCache[K comparable, V any] struct {
	// Configuration
	maxSize    int
	defaultTTL time.Duration

	// Internal storage
	mu       sync.RWMutex
	items    map[K]*list.Element
	lruList  *list.List
	closed   atomic.Bool

	// Metrics
	hits        atomic.Uint64
	misses      atomic.Uint64
	evictions   atomic.Uint64
	expirations atomic.Uint64
	createdAt   time.Time
	lastAccess  atomic.Int64

	// Cleanup
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
	cleanupWg       sync.WaitGroup

	// Invalidation listeners
	listenersMu sync.RWMutex
	listeners   []InvalidationListener
}

// memoryCacheEntry is the internal entry type stored in the LRU list.
type memoryCacheEntry[K comparable, V any] struct {
	key       K
	value     V
	expiresAt time.Time
	createdAt time.Time
}

// MemoryCacheOption is a functional option for configuring MemoryCache.
type MemoryCacheOption[K comparable, V any] func(*MemoryCache[K, V])

// WithMaxSize sets the maximum cache size.
func WithMaxSize[K comparable, V any](size int) MemoryCacheOption[K, V] {
	return func(c *MemoryCache[K, V]) {
		if size > 0 {
			c.maxSize = size
		}
	}
}

// WithDefaultTTL sets the default TTL for cache entries.
func WithDefaultTTL[K comparable, V any](ttl time.Duration) MemoryCacheOption[K, V] {
	return func(c *MemoryCache[K, V]) {
		c.defaultTTL = ttl
	}
}

// WithCleanupInterval sets how often expired entries are cleaned up.
func WithCleanupInterval[K comparable, V any](interval time.Duration) MemoryCacheOption[K, V] {
	return func(c *MemoryCache[K, V]) {
		if interval > 0 {
			c.cleanupInterval = interval
		}
	}
}

// NewMemoryCache creates a new in-memory LRU cache.
func NewMemoryCache[K comparable, V any](opts ...MemoryCacheOption[K, V]) *MemoryCache[K, V] {
	c := &MemoryCache[K, V]{
		maxSize:         1000,
		defaultTTL:      5 * time.Minute,
		items:           make(map[K]*list.Element),
		lruList:         list.New(),
		createdAt:       time.Now(),
		cleanupInterval: 1 * time.Minute,
		stopCleanup:     make(chan struct{}),
	}

	for _, opt := range opts {
		opt(c)
	}

	// Start background cleanup
	c.cleanupWg.Add(1)
	go c.cleanupLoop()

	return c
}

// Get retrieves a value from the cache.
func (c *MemoryCache[K, V]) Get(ctx context.Context, key K) (V, error) {
	var zero V

	if c.closed.Load() {
		return zero, ErrCacheClosed
	}

	c.lastAccess.Store(time.Now().UnixNano())

	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		c.misses.Add(1)
		return zero, ErrCacheMiss
	}

	entry := elem.Value.(*memoryCacheEntry[K, V])

	// Check expiration
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		c.removeElement(elem)
		c.expirations.Add(1)
		c.misses.Add(1)
		return zero, ErrCacheMiss
	}

	// Move to front (most recently used)
	c.lruList.MoveToFront(elem)
	c.hits.Add(1)

	return entry.value, nil
}

// Set stores a value with the default TTL.
func (c *MemoryCache[K, V]) Set(ctx context.Context, key K, value V) error {
	return c.SetWithTTL(ctx, key, value, c.defaultTTL)
}

// SetWithTTL stores a value with a specific TTL.
func (c *MemoryCache[K, V]) SetWithTTL(ctx context.Context, key K, value V, ttl time.Duration) error {
	if c.closed.Load() {
		return ErrCacheClosed
	}

	c.lastAccess.Store(time.Now().UnixNano())

	now := time.Now()
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = now.Add(ttl)
	}

	entry := &memoryCacheEntry[K, V]{
		key:       key,
		value:     value,
		expiresAt: expiresAt,
		createdAt: now,
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Update existing entry
	if elem, ok := c.items[key]; ok {
		c.lruList.MoveToFront(elem)
		elem.Value = entry
		return nil
	}

	// Evict if at capacity
	for c.lruList.Len() >= c.maxSize {
		c.evictOldest()
	}

	// Add new entry
	elem := c.lruList.PushFront(entry)
	c.items[key] = elem

	return nil
}

// Delete removes a value from the cache.
func (c *MemoryCache[K, V]) Delete(ctx context.Context, key K) error {
	if c.closed.Load() {
		return ErrCacheClosed
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.removeElement(elem)
		c.notifyInvalidation(key, "explicit delete")
	}

	return nil
}

// Exists checks if a key exists in the cache.
func (c *MemoryCache[K, V]) Exists(ctx context.Context, key K) bool {
	if c.closed.Load() {
		return false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	elem, ok := c.items[key]
	if !ok {
		return false
	}

	entry := elem.Value.(*memoryCacheEntry[K, V])
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		return false
	}

	return true
}

// Clear removes all entries from the cache.
func (c *MemoryCache[K, V]) Clear(ctx context.Context) error {
	if c.closed.Load() {
		return ErrCacheClosed
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[K]*list.Element)
	c.lruList.Init()

	c.notifyInvalidation(*new(K), "cache cleared")

	return nil
}

// Size returns the number of items in the cache.
func (c *MemoryCache[K, V]) Size(ctx context.Context) int {
	if c.closed.Load() {
		return 0
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.lruList.Len()
}

// Close releases resources and stops background tasks.
func (c *MemoryCache[K, V]) Close() error {
	if c.closed.Swap(true) {
		return nil // Already closed
	}

	close(c.stopCleanup)
	c.cleanupWg.Wait()

	c.mu.Lock()
	c.items = nil
	c.lruList = nil
	c.mu.Unlock()

	return nil
}

// Stats returns current cache statistics.
func (c *MemoryCache[K, V]) Stats() CacheStats {
	c.mu.RLock()
	size := 0
	if c.lruList != nil {
		size = c.lruList.Len()
	}
	c.mu.RUnlock()

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
		MaxSize:        c.maxSize,
		CreatedAt:      c.createdAt,
		LastAccessedAt: lastAccessTime,
	}
}

// OnInvalidate registers a listener for cache invalidation events.
func (c *MemoryCache[K, V]) OnInvalidate(listener InvalidationListener) {
	c.listenersMu.Lock()
	defer c.listenersMu.Unlock()
	c.listeners = append(c.listeners, listener)
}

// GetOrLoad retrieves a value or loads it using the provided loader.
func (c *MemoryCache[K, V]) GetOrLoad(ctx context.Context, key K, loader CacheLoader[K, V]) (V, error) {
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

// evictOldest removes the least recently used entry.
// Must be called with lock held.
func (c *MemoryCache[K, V]) evictOldest() {
	elem := c.lruList.Back()
	if elem == nil {
		return
	}

	c.removeElement(elem)
	c.evictions.Add(1)
}

// removeElement removes an element from the cache.
// Must be called with lock held.
func (c *MemoryCache[K, V]) removeElement(elem *list.Element) {
	entry := elem.Value.(*memoryCacheEntry[K, V])
	delete(c.items, entry.key)
	c.lruList.Remove(elem)
}

// cleanupLoop periodically removes expired entries.
func (c *MemoryCache[K, V]) cleanupLoop() {
	defer c.cleanupWg.Done()

	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopCleanup:
			return
		}
	}
}

// cleanup removes all expired entries.
func (c *MemoryCache[K, V]) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lruList == nil {
		return
	}

	now := time.Now()
	var toRemove []*list.Element

	for elem := c.lruList.Back(); elem != nil; elem = elem.Prev() {
		entry := elem.Value.(*memoryCacheEntry[K, V])
		if !entry.expiresAt.IsZero() && now.After(entry.expiresAt) {
			toRemove = append(toRemove, elem)
		}
	}

	for _, elem := range toRemove {
		c.removeElement(elem)
		c.expirations.Add(1)
	}
}

// notifyInvalidation notifies all listeners of a cache invalidation.
func (c *MemoryCache[K, V]) notifyInvalidation(key K, reason string) {
	c.listenersMu.RLock()
	listeners := make([]InvalidationListener, len(c.listeners))
	copy(listeners, c.listeners)
	c.listenersMu.RUnlock()

	if len(listeners) == 0 {
		return
	}

	event := InvalidationEvent{
		Key:       formatKey(key),
		Reason:    reason,
		Timestamp: time.Now(),
	}

	for _, listener := range listeners {
		go listener(event)
	}
}

// formatKey converts a key to a string for event reporting.
func formatKey[K comparable](key K) string {
	switch k := any(key).(type) {
	case string:
		return k
	case []byte:
		return string(k)
	default:
		return ""
	}
}

// Compile-time interface checks
var (
	_ Cache[string, any]            = (*MemoryCache[string, any])(nil)
	_ CacheWithMetrics[string, any] = (*MemoryCache[string, any])(nil)
	_ LoadingCache[string, any]     = (*MemoryCache[string, any])(nil)
	_ ObservableCache[string, any]  = (*MemoryCache[string, any])(nil)
)

