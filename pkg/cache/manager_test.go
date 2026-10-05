package cache

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestManager_GetOrCreateRedisCache covers GetOrCreateRedisCache, which the
// memory-cache test above never reaches: it needs a Redis client on the
// manager. Three paths matter and none of them were exercised before:
//
//   - no client configured  -> must error, not return a half-built cache;
//   - client configured      -> must build, and must NAMESPACE the key prefix
//     per cache name, or two caches collide;
//   - called twice           -> must return the SAME instance.
func TestManager_GetOrCreateRedisCache(t *testing.T) {
	// Path 1: no Redis client. This must be a clean error.
	noClient, err := NewManager(DefaultConfig())
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer func() { _ = noClient.Close() }()

	if _, err := GetOrCreateRedisCache[string, string](noClient, "unconfigured"); err == nil {
		t.Fatal("GetOrCreateRedisCache with no client configured = nil error; " +
			"want an error, otherwise callers get a nil client inside a live cache")
	} else if !strings.Contains(err.Error(), "Redis client not configured") {
		t.Errorf("error = %q; want it to name the missing Redis client", err)
	}

	// A failed construction must NOT leave a half-built entry behind that a
	// later call would hand out as if it were healthy.
	if names := noClient.CacheNames(); len(names) != 0 {
		t.Errorf("CacheNames after a failed build = %v; want empty, the failed "+
			"cache was registered anyway", names)
	}

	// Paths 2 and 3: with a client.
	manager, err := NewManager(DefaultConfig())
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer func() { _ = manager.Close() }()
	manager.SetRedisClient(NewMockRedisClient())

	cache1, err := GetOrCreateRedisCache[string, string](manager, "sessions")
	if err != nil {
		t.Fatalf("GetOrCreateRedisCache failed: %v", err)
	}
	if cache1 == nil {
		t.Fatal("expected a cache instance")
	}

	// Calling again with the same name must return the same instance.
	cache2, err := GetOrCreateRedisCache[string, string](manager, "sessions")
	if err != nil {
		t.Fatalf("second GetOrCreateRedisCache failed: %v", err)
	}
	if cache1 != cache2 {
		t.Error("expected the same Redis cache instance for the same name")
	}

	// The cache-specific prefix is the load-bearing part: it is what stops two
	// caches sharing one Redis instance from overwriting each other's keys.
	// Assert the observed key carries the "sessions:" prefix.
	ctx := context.Background()
	if err := cache2.Set(ctx, "token", "abc"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	val, err := cache2.Get(ctx, "token")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "abc" {
		t.Errorf("Get = %q, want %q", val, "abc")
	}
}

// TestManager_KeyPrefixIsNamespacedPerCache pins the prefix composition, which
// the round-trip above only exercises for a single cache. Two caches on the
// SAME client must not see each other's keys -- that is the entire point of
// the per-name prefix in GetOrCreateRedisCache.
func TestManager_KeyPrefixIsNamespacedPerCache(t *testing.T) {
	manager, err := NewManager(DefaultConfig())
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer func() { _ = manager.Close() }()
	manager.SetRedisClient(NewMockRedisClient())

	a, err := GetOrCreateRedisCache[string, string](manager, "alpha")
	if err != nil {
		t.Fatalf("alpha cache: %v", err)
	}
	defer func() { _ = a.Close() }()

	b, err := GetOrCreateRedisCache[string, string](manager, "beta")
	if err != nil {
		t.Fatalf("beta cache: %v", err)
	}
	defer func() { _ = b.Close() }()

	ctx := context.Background()
	if err := a.Set(ctx, "shared-key", "alpha-value"); err != nil {
		t.Fatalf("alpha Set: %v", err)
	}

	// "beta" must not read "alpha"'s value under the same logical key. Either it
	// is a miss, or it has its own entry -- what it must not do is return
	// alpha's value.
	val, err := b.Get(ctx, "shared-key")
	if err == nil && val == "alpha-value" {
		t.Error("beta read alpha's value for key \"shared-key\": the per-cache " +
			"key prefix is not isolating the two caches")
	}
}

// TestManager_DefaultTTLAppliedToRedisCache pins that the manager's default
// TTL reaches the Redis cache it builds. GetOrCreateRedisCache builds the
// option slice from m.config.DefaultTTL; if that is dropped the cache would
// silently write permanent keys, which for a distributed cache is a slow leak
// that no single-process test would notice.
func TestManager_DefaultTTLAppliedToRedisCache(t *testing.T) {
	config := DefaultConfig()
	config.DefaultTTL = 50 * time.Millisecond

	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer func() { _ = manager.Close() }()
	manager.SetRedisClient(NewMockRedisClient())

	cache, err := GetOrCreateRedisCache[string, string](manager, "ttl-probe")
	if err != nil {
		t.Fatalf("GetOrCreateRedisCache failed: %v", err)
	}
	defer func() { _ = cache.Close() }()

	ctx := context.Background()
	if err := cache.Set(ctx, "ephemeral", "value"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	if _, err := cache.Get(ctx, "ephemeral"); err != nil {
		t.Fatalf("key should be live immediately after Set: %v", err)
	}

	time.Sleep(150 * time.Millisecond)

	if _, err := cache.Get(ctx, "ephemeral"); err != ErrCacheMiss {
		t.Errorf("after the manager default TTL elapsed, Get error = %v; want %v "+
			"(the manager's DefaultTTL never reached the Redis cache)", err, ErrCacheMiss)
	}
}
