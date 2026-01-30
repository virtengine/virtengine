package cache

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestMemoryCache_BasicOperations(t *testing.T) {
	cache := NewMemoryCache[string, string](
		WithMaxSize[string, string](100),
		WithDefaultTTL[string, string](5*time.Minute),
	)
	defer cache.Close()

	ctx := context.Background()

	// Test Set and Get
	if err := cache.Set(ctx, "key1", "value1"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err := cache.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "value1" {
		t.Errorf("expected 'value1', got '%s'", val)
	}

	// Test Exists
	if !cache.Exists(ctx, "key1") {
		t.Error("expected key1 to exist")
	}
	if cache.Exists(ctx, "nonexistent") {
		t.Error("expected nonexistent to not exist")
	}

	// Test Delete
	if err := cache.Delete(ctx, "key1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if cache.Exists(ctx, "key1") {
		t.Error("expected key1 to be deleted")
	}

	// Test cache miss
	_, err = cache.Get(ctx, "nonexistent")
	if err != ErrCacheMiss {
		t.Errorf("expected ErrCacheMiss, got %v", err)
	}
}

func TestMemoryCache_TTLExpiration(t *testing.T) {
	cache := NewMemoryCache[string, string](
		WithMaxSize[string, string](100),
		WithDefaultTTL[string, string](50*time.Millisecond),
		WithCleanupInterval[string, string](10*time.Millisecond),
	)
	defer cache.Close()

	ctx := context.Background()

	// Set a value
	if err := cache.Set(ctx, "key1", "value1"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Should exist immediately
	if !cache.Exists(ctx, "key1") {
		t.Error("expected key1 to exist")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired
	_, err := cache.Get(ctx, "key1")
	if err != ErrCacheMiss {
		t.Errorf("expected ErrCacheMiss after expiration, got %v", err)
	}
}

func TestMemoryCache_LRUEviction(t *testing.T) {
	cache := NewMemoryCache[string, string](
		WithMaxSize[string, string](3),
		WithDefaultTTL[string, string](5*time.Minute),
	)
	defer cache.Close()

	ctx := context.Background()

	// Fill cache to capacity
	cache.Set(ctx, "key1", "value1")
	cache.Set(ctx, "key2", "value2")
	cache.Set(ctx, "key3", "value3")

	// Access key1 to make it recently used
	cache.Get(ctx, "key1")

	// Add new item, should evict key2 (least recently used)
	cache.Set(ctx, "key4", "value4")

	// key2 should be evicted
	if cache.Exists(ctx, "key2") {
		t.Error("expected key2 to be evicted")
	}

	// key1, key3, key4 should exist
	if !cache.Exists(ctx, "key1") {
		t.Error("expected key1 to exist")
	}
	if !cache.Exists(ctx, "key3") {
		t.Error("expected key3 to exist")
	}
	if !cache.Exists(ctx, "key4") {
		t.Error("expected key4 to exist")
	}
}

func TestMemoryCache_Stats(t *testing.T) {
	cache := NewMemoryCache[string, string](
		WithMaxSize[string, string](100),
		WithDefaultTTL[string, string](5*time.Minute),
	)
	defer cache.Close()

	ctx := context.Background()

	// Initial stats
	stats := cache.Stats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Error("expected zero initial stats")
	}

	// Generate hits and misses
	cache.Set(ctx, "key1", "value1")
	cache.Get(ctx, "key1") // hit
	cache.Get(ctx, "key1") // hit
	cache.Get(ctx, "miss") // miss

	stats = cache.Stats()
	if stats.Hits != 2 {
		t.Errorf("expected 2 hits, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.Misses)
	}
	if stats.Size != 1 {
		t.Errorf("expected size 1, got %d", stats.Size)
	}

	// Check hit rate
	hitRate := stats.HitRate()
	expected := float64(2) / float64(3) * 100
	if hitRate != expected {
		t.Errorf("expected hit rate %.2f, got %.2f", expected, hitRate)
	}
}

func TestMemoryCache_Clear(t *testing.T) {
	cache := NewMemoryCache[string, string](
		WithMaxSize[string, string](100),
		WithDefaultTTL[string, string](5*time.Minute),
	)
	defer cache.Close()

	ctx := context.Background()

	// Add items
	cache.Set(ctx, "key1", "value1")
	cache.Set(ctx, "key2", "value2")
	cache.Set(ctx, "key3", "value3")

	if cache.Size(ctx) != 3 {
		t.Errorf("expected size 3, got %d", cache.Size(ctx))
	}

	// Clear
	if err := cache.Clear(ctx); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	if cache.Size(ctx) != 0 {
		t.Errorf("expected size 0 after clear, got %d", cache.Size(ctx))
	}
}

func TestMemoryCache_GetOrLoad(t *testing.T) {
	cache := NewMemoryCache[string, string](
		WithMaxSize[string, string](100),
		WithDefaultTTL[string, string](5*time.Minute),
	)
	defer cache.Close()

	ctx := context.Background()
	loadCount := 0

	loader := func(ctx context.Context, key string) (string, error) {
		loadCount++
		return "loaded:" + key, nil
	}

	// First call should load
	val, err := cache.GetOrLoad(ctx, "key1", loader)
	if err != nil {
		t.Fatalf("GetOrLoad failed: %v", err)
	}
	if val != "loaded:key1" {
		t.Errorf("expected 'loaded:key1', got '%s'", val)
	}
	if loadCount != 1 {
		t.Errorf("expected 1 load, got %d", loadCount)
	}

	// Second call should use cache
	val, err = cache.GetOrLoad(ctx, "key1", loader)
	if err != nil {
		t.Fatalf("GetOrLoad failed: %v", err)
	}
	if val != "loaded:key1" {
		t.Errorf("expected 'loaded:key1', got '%s'", val)
	}
	if loadCount != 1 {
		t.Errorf("expected 1 load (cached), got %d", loadCount)
	}
}

func TestMemoryCache_Concurrent(t *testing.T) {
	cache := NewMemoryCache[string, int](
		WithMaxSize[string, int](1000),
		WithDefaultTTL[string, int](5*time.Minute),
	)
	defer cache.Close()

	ctx := context.Background()
	var wg sync.WaitGroup
	numGoroutines := 100
	numOps := 100

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := "key" + string(rune('a'+id%26))
				cache.Set(ctx, key, id*numOps+j)
				cache.Get(ctx, key)
				if j%10 == 0 {
					cache.Delete(ctx, key)
				}
			}
		}(i)
	}

	wg.Wait()

	// Should not panic or have data races
	stats := cache.Stats()
	t.Logf("Concurrent test stats: hits=%d, misses=%d, size=%d", stats.Hits, stats.Misses, stats.Size)
}

func TestMemoryCache_InvalidationListener(t *testing.T) {
	cache := NewMemoryCache[string, string](
		WithMaxSize[string, string](100),
		WithDefaultTTL[string, string](5*time.Minute),
	)
	defer cache.Close()

	ctx := context.Background()
	var events []InvalidationEvent
	var mu sync.Mutex

	cache.OnInvalidate(func(event InvalidationEvent) {
		mu.Lock()
		events = append(events, event)
		mu.Unlock()
	})

	// Add and delete
	cache.Set(ctx, "key1", "value1")
	cache.Delete(ctx, "key1")

	// Give async notification time
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
	if len(events) > 0 && events[0].Reason != "explicit delete" {
		t.Errorf("expected reason 'explicit delete', got '%s'", events[0].Reason)
	}
	mu.Unlock()
}

func TestRedisCache_WithMock(t *testing.T) {
	mockClient := NewMockRedisClient()
	config := RedisConfig{
		URL:         "localhost:6379",
		KeyPrefix:   "test:",
		DialTimeout: 5 * time.Second,
	}

	cache, err := NewRedisCache[string, string](mockClient, config)
	if err != nil {
		t.Fatalf("NewRedisCache failed: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	// Test Set and Get
	if err := cache.Set(ctx, "key1", "value1"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err := cache.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "value1" {
		t.Errorf("expected 'value1', got '%s'", val)
	}

	// Test Exists
	if !cache.Exists(ctx, "key1") {
		t.Error("expected key1 to exist")
	}

	// Test Delete
	if err := cache.Delete(ctx, "key1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if cache.Exists(ctx, "key1") {
		t.Error("expected key1 to be deleted")
	}

	// Test cache miss
	_, err = cache.Get(ctx, "nonexistent")
	if err != ErrCacheMiss {
		t.Errorf("expected ErrCacheMiss, got %v", err)
	}
}

func TestCacheEntry_IsExpired(t *testing.T) {
	// Not expired
	entry := CacheEntry[string]{
		Value:     "test",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	if entry.IsExpired() {
		t.Error("expected entry not to be expired")
	}

	// Expired
	entry.ExpiresAt = time.Now().Add(-1 * time.Hour)
	if !entry.IsExpired() {
		t.Error("expected entry to be expired")
	}

	// Never expires
	entry.ExpiresAt = time.Time{}
	if entry.IsExpired() {
		t.Error("expected entry with zero ExpiresAt not to be expired")
	}
}

func TestCacheStats_HitRate(t *testing.T) {
	tests := []struct {
		name     string
		hits     uint64
		misses   uint64
		expected float64
	}{
		{"no accesses", 0, 0, 0},
		{"all hits", 100, 0, 100},
		{"all misses", 0, 100, 0},
		{"50% hit rate", 50, 50, 50},
		{"75% hit rate", 75, 25, 75},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := CacheStats{
				Hits:   tt.hits,
				Misses: tt.misses,
			}
			if rate := stats.HitRate(); rate != tt.expected {
				t.Errorf("expected %.2f, got %.2f", tt.expected, rate)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	// Valid config
	config := DefaultConfig()
	if err := config.Validate(); err != nil {
		t.Errorf("expected valid config, got error: %v", err)
	}

	// Disabled config should skip validation
	config.Enabled = false
	config.MaxSize = -1 // Invalid, but should be ignored when disabled
	if err := config.Validate(); err != nil {
		t.Errorf("expected disabled config to skip validation, got error: %v", err)
	}

	// Invalid backend
	config.Enabled = true
	config.Backend = "invalid"
	if err := config.Validate(); err == nil {
		t.Error("expected error for invalid backend")
	}

	// Invalid max size for memory backend
	config.Backend = BackendMemory
	config.MaxSize = 0
	if err := config.Validate(); err == nil {
		t.Error("expected error for invalid max size")
	}
}

func TestManager_GetOrCreateMemoryCache(t *testing.T) {
	config := DefaultConfig()
	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer manager.Close()

	// Create cache
	cache1 := GetOrCreateMemoryCache[string, string](manager, "test-cache")
	if cache1 == nil {
		t.Fatal("expected cache to be created")
	}

	// Get same cache
	cache2 := GetOrCreateMemoryCache[string, string](manager, "test-cache")
	if cache1 != cache2 {
		t.Error("expected same cache instance")
	}

	// Create different cache
	cache3 := GetOrCreateMemoryCache[string, string](manager, "other-cache")
	if cache1 == cache3 {
		t.Error("expected different cache instance")
	}

	// Check cache names
	names := manager.CacheNames()
	if len(names) != 2 {
		t.Errorf("expected 2 caches, got %d", len(names))
	}
}

func TestMetrics_RegisterAndReport(t *testing.T) {
	metrics := NewMetrics("test", 100*time.Millisecond)

	cache := NewMemoryCache[string, string](
		WithMaxSize[string, string](100),
		WithDefaultTTL[string, string](5*time.Minute),
	)
	defer cache.Close()

	metrics.RegisterCache("test-cache", cache)

	// Add some data
	ctx := context.Background()
	cache.Set(ctx, "key1", "value1")
	cache.Get(ctx, "key1") // hit
	cache.Get(ctx, "miss") // miss

	// Get report
	report := metrics.GetReport()

	if len(report.Caches) != 1 {
		t.Errorf("expected 1 cache, got %d", len(report.Caches))
	}

	stats, ok := report.Caches["test-cache"]
	if !ok {
		t.Fatal("expected test-cache in report")
	}

	if stats.Hits != 1 {
		t.Errorf("expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.Misses)
	}
}

func BenchmarkMemoryCache_Get(b *testing.B) {
	cache := NewMemoryCache[string, string](
		WithMaxSize[string, string](10000),
		WithDefaultTTL[string, string](5*time.Minute),
	)
	defer cache.Close()

	ctx := context.Background()

	// Pre-populate
	for i := 0; i < 1000; i++ {
		cache.Set(ctx, "key"+string(rune(i)), "value")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(ctx, "key"+string(rune(i%1000)))
	}
}

func BenchmarkMemoryCache_Set(b *testing.B) {
	cache := NewMemoryCache[string, string](
		WithMaxSize[string, string](10000),
		WithDefaultTTL[string, string](5*time.Minute),
	)
	defer cache.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(ctx, "key"+string(rune(i%10000)), "value")
	}
}

func BenchmarkMemoryCache_ConcurrentAccess(b *testing.B) {
	cache := NewMemoryCache[string, string](
		WithMaxSize[string, string](10000),
		WithDefaultTTL[string, string](5*time.Minute),
	)
	defer cache.Close()

	ctx := context.Background()

	// Pre-populate
	for i := 0; i < 1000; i++ {
		cache.Set(ctx, "key"+string(rune(i)), "value")
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "key" + string(rune(i%1000))
			if i%2 == 0 {
				cache.Get(ctx, key)
			} else {
				cache.Set(ctx, key, "value")
			}
			i++
		}
	})
}
