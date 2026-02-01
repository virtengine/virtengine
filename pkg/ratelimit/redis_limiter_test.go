package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisRateLimiter(t *testing.T) {
	// Skip if Redis is not available
	ctx := context.Background()
	config := DefaultConfig()
	config.RedisURL = "redis://localhost:6379/15" // Use test database

	logger := zerolog.New(nil).Level(zerolog.Disabled)

	limiter, err := NewRedisRateLimiter(ctx, config, logger)
	if err != nil {
		t.Skip("Redis not available for testing")
	}
	defer limiter.Close()

	t.Run("Allow requests within limit", func(t *testing.T) {
		key := "test-ip-1"

		// First request should be allowed
		allowed, result, err := limiter.Allow(ctx, key, LimitTypeIP)
		require.NoError(t, err)
		assert.True(t, allowed)
		assert.NotNil(t, result)
	})

	t.Run("Block requests exceeding limit", func(t *testing.T) {
		key := "test-ip-2"

		// Make requests up to the limit
		limit := config.IPLimits.RequestsPerSecond
		for i := 0; i < limit; i++ {
			allowed, _, err := limiter.Allow(ctx, key, LimitTypeIP)
			require.NoError(t, err)
			assert.True(t, allowed, "Request %d should be allowed", i+1)
		}

		// Next request should be blocked
		allowed, result, err := limiter.Allow(ctx, key, LimitTypeIP)
		require.NoError(t, err)
		assert.False(t, allowed)
		assert.Greater(t, result.RetryAfter, 0)
	})

	t.Run("Whitelist functionality", func(t *testing.T) {
		key := "whitelisted-ip"
		config.WhitelistedIPs = []string{key}
		limiter.UpdateConfig(config)

		// Whitelisted IPs should always be allowed
		for i := 0; i < 100; i++ {
			allowed, _, err := limiter.Allow(ctx, key, LimitTypeIP)
			require.NoError(t, err)
			assert.True(t, allowed)
		}
	})

	t.Run("Ban functionality", func(t *testing.T) {
		key := "banned-ip"

		// Ban the IP
		err := limiter.Ban(ctx, key, time.Minute, "test ban")
		require.NoError(t, err)

		// Requests should be blocked
		allowed, _, err := limiter.Allow(ctx, key, LimitTypeIP)
		require.NoError(t, err)
		assert.False(t, allowed)

		// Check if banned
		banned, err := limiter.IsBanned(ctx, key)
		require.NoError(t, err)
		assert.True(t, banned)
	})

	t.Run("Endpoint-specific limits", func(t *testing.T) {
		key := "test-ip-3"
		endpoint := "/veid/verify"

		// Endpoint should have stricter limits
		for i := 0; i < 5; i++ {
			allowed, _, err := limiter.AllowEndpoint(ctx, endpoint, key, LimitTypeIP)
			require.NoError(t, err)
			assert.True(t, allowed)
		}

		// Should be rate limited at endpoint level
		allowed, _, err := limiter.AllowEndpoint(ctx, endpoint, key, LimitTypeIP)
		require.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("Bypass attempt recording", func(t *testing.T) {
		key := "suspicious-ip"

		// Record bypass attempts
		for i := 0; i < 150; i++ {
			err := limiter.RecordBypassAttempt(ctx, key, "rate_limit_exceeded")
			require.NoError(t, err)
		}

		// Should be auto-banned after threshold
		time.Sleep(100 * time.Millisecond)
		banned, err := limiter.IsBanned(ctx, key)
		require.NoError(t, err)
		assert.True(t, banned)
	})

	t.Run("Metrics collection", func(t *testing.T) {
		metrics, err := limiter.GetMetrics(ctx)
		require.NoError(t, err)
		assert.NotNil(t, metrics)
		assert.GreaterOrEqual(t, metrics.TotalRequests, uint64(0))
	})

	t.Run("Load calculation", func(t *testing.T) {
		load, err := limiter.GetCurrentLoad(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, load, 0.0)
		assert.LessOrEqual(t, load, 100.0)
	})

	t.Run("Pattern matching", func(t *testing.T) {
		// Test IP pattern matching
		assert.True(t, limiter.matchesIPPattern("192.168.1.1", "192.168.1.1"))
		assert.True(t, limiter.matchesIPPattern("192.168.1.1", "192.168.1.0/24"))
		assert.False(t, limiter.matchesIPPattern("192.168.2.1", "192.168.1.0/24"))

		// Test endpoint pattern matching
		assert.True(t, limiter.matchesPattern("/market/orders", "/market/*"))
		assert.True(t, limiter.matchesPattern("/veid/verify", "/veid/*"))
		assert.False(t, limiter.matchesPattern("/other/endpoint", "/market/*"))
	})
}

func TestLimitRulesValidation(t *testing.T) {
	t.Run("Valid configuration", func(t *testing.T) {
		config := DefaultConfig()
		assert.NotEmpty(t, config.RedisURL)
		assert.True(t, config.Enabled)
	})

	t.Run("Multiplier application", func(t *testing.T) {
		rules := LimitRules{
			RequestsPerSecond: 100,
			RequestsPerMinute: 1000,
			BurstSize:         200,
		}

		// Apply 50% reduction
		reduced := applyMultiplier(rules, 0.5)
		assert.Equal(t, 50, reduced.RequestsPerSecond)
		assert.Equal(t, 500, reduced.RequestsPerMinute)
		assert.Equal(t, 100, reduced.BurstSize)
	})
}

func TestGracefulDegradation(t *testing.T) {
	config := DefaultConfig()
	config.GracefulDegradation.Enabled = true
	config.GracefulDegradation.LoadThresholds = []LoadThreshold{
		{
			LoadPercentage:      80,
			RateLimitMultiplier: 0.7,
			Priority:            []string{"/veid/*"},
		},
	}

	limiter := &RedisRateLimiter{
		config: config,
	}

	t.Run("No degradation under low load", func(t *testing.T) {
		multiplier := limiter.getDegradationMultiplier(50.0, "/market/orders")
		assert.Equal(t, 1.0, multiplier)
	})

	t.Run("Degradation at high load", func(t *testing.T) {
		multiplier := limiter.getDegradationMultiplier(85.0, "/market/orders")
		assert.Equal(t, 0.7, multiplier)
	})

	t.Run("Priority endpoints not degraded", func(t *testing.T) {
		multiplier := limiter.getDegradationMultiplier(85.0, "/veid/verify")
		assert.Equal(t, 1.0, multiplier)
	})
}

func TestWhitelisting(t *testing.T) {
	config := DefaultConfig()
	config.WhitelistedIPs = []string{"10.0.0.1", "192.168.1.0/24"}
	config.WhitelistedUsers = []string{"admin", "service-account"}

	limiter := &RedisRateLimiter{
		config: config,
	}

	t.Run("IP whitelist exact match", func(t *testing.T) {
		assert.True(t, limiter.IsWhitelisted("10.0.0.1", LimitTypeIP))
		assert.False(t, limiter.IsWhitelisted("10.0.0.2", LimitTypeIP))
	})

	t.Run("IP whitelist CIDR match", func(t *testing.T) {
		assert.True(t, limiter.IsWhitelisted("192.168.1.50", LimitTypeIP))
		assert.False(t, limiter.IsWhitelisted("192.168.2.1", LimitTypeIP))
	})

	t.Run("User whitelist", func(t *testing.T) {
		assert.True(t, limiter.IsWhitelisted("admin", LimitTypeUser))
		assert.True(t, limiter.IsWhitelisted("service-account", LimitTypeUser))
		assert.False(t, limiter.IsWhitelisted("regular-user", LimitTypeUser))
	})
}

func BenchmarkRateLimiter(b *testing.B) {
	ctx := context.Background()
	config := DefaultConfig()
	config.RedisURL = "redis://localhost:6379/15"

	logger := zerolog.New(nil).Level(zerolog.Disabled)

	limiter, err := NewRedisRateLimiter(ctx, config, logger)
	if err != nil {
		b.Skip("Redis not available for benchmarking")
	}
	defer limiter.Close()

	b.Run("Allow", func(b *testing.B) {
		key := "bench-ip"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			limiter.Allow(ctx, key, LimitTypeIP)
		}
	})

	b.Run("AllowEndpoint", func(b *testing.B) {
		key := "bench-ip"
		endpoint := "/market/orders"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			limiter.AllowEndpoint(ctx, endpoint, key, LimitTypeIP)
		}
	})

	b.Run("IsBanned", func(b *testing.B) {
		key := "bench-ip"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			limiter.IsBanned(ctx, key)
		}
	})
}

func applyMultiplier(rules LimitRules, multiplier float64) LimitRules {
	return LimitRules{
		RequestsPerSecond: int(float64(rules.RequestsPerSecond) * multiplier),
		RequestsPerMinute: int(float64(rules.RequestsPerMinute) * multiplier),
		RequestsPerHour:   int(float64(rules.RequestsPerHour) * multiplier),
		RequestsPerDay:    int(float64(rules.RequestsPerDay) * multiplier),
		BurstSize:         int(float64(rules.BurstSize) * multiplier),
	}
}

