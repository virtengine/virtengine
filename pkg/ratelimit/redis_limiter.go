package ratelimit

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// RedisRateLimiter implements RateLimiter using Redis
type RedisRateLimiter struct {
	client *redis.Client
	config RateLimitConfig
	logger zerolog.Logger
	mu     sync.RWMutex

	// Metrics tracking
	metrics *metricsTracker
}

// metricsTracker tracks rate limiting metrics
type metricsTracker struct {
	mu                sync.RWMutex
	totalRequests     uint64
	allowedRequests   uint64
	blockedRequests   uint64
	bypassAttempts    uint64
	byLimitType       map[LimitType]*LimitTypeMetrics
	blockedIPCounts   map[string]uint64
	blockedUserCounts map[string]uint64
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(ctx context.Context, config RateLimitConfig, logger zerolog.Logger) (*RedisRateLimiter, error) {
	opts, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisRateLimiter{
		client: client,
		config: config,
		logger: logger.With().Str("component", "rate-limiter").Logger(),
		metrics: &metricsTracker{
			byLimitType:       make(map[LimitType]*LimitTypeMetrics),
			blockedIPCounts:   make(map[string]uint64),
			blockedUserCounts: make(map[string]uint64),
		},
	}, nil
}

// Allow checks if a request is allowed based on rate limits
func (r *RedisRateLimiter) Allow(ctx context.Context, key string, limitType LimitType) (bool, *RateLimitResult, error) {
	r.mu.RLock()
	enabled := r.config.Enabled
	r.mu.RUnlock()

	if !enabled {
		return true, &RateLimitResult{
			Allowed:    true,
			Identifier: key,
			LimitType:  limitType,
		}, nil
	}

	// Check if whitelisted
	if r.IsWhitelisted(key, limitType) {
		return true, &RateLimitResult{
			Allowed:    true,
			Identifier: key,
			LimitType:  limitType,
		}, nil
	}

	// Check if banned
	banned, err := r.IsBanned(ctx, key)
	if err != nil {
		r.logger.Error().Err(err).Str("key", key).Msg("failed to check ban status")
	}
	if banned {
		r.recordMetric(false, limitType, key)
		return false, &RateLimitResult{
			Allowed:    false,
			Limit:      0,
			Remaining:  0,
			RetryAfter: 3600, // 1 hour
			LimitType:  limitType,
			Identifier: key,
		}, nil
	}

	// Get applicable limits
	rules := r.getLimitsForType(limitType)

	// Apply graceful degradation multiplier
	load, _ := r.GetCurrentLoad(ctx)
	multiplier := r.getDegradationMultiplier(load, "")
	rules = r.applyMultiplier(rules, multiplier)

	// Check all time windows
	allowed, result, err := r.checkLimits(ctx, key, limitType, rules)
	if err != nil {
		return false, nil, err
	}

	// Record metrics
	r.recordMetric(allowed, limitType, key)

	// Detect bypass attempts
	if !allowed && r.config.BypassDetection.Enabled {
		if err := r.RecordBypassAttempt(ctx, key, "rate_limit_exceeded"); err != nil {
			r.logger.Warn().Err(err).Str("key", key).Msg("failed to record bypass attempt")
		}
	}

	return allowed, result, nil
}

// AllowEndpoint checks if a request to a specific endpoint is allowed
func (r *RedisRateLimiter) AllowEndpoint(ctx context.Context, endpoint string, identifier string, limitType LimitType) (bool, *RateLimitResult, error) {
	// First check general limit
	allowed, result, err := r.Allow(ctx, identifier, limitType)
	if err != nil {
		return false, nil, err
	}
	if !allowed {
		return false, result, nil
	}

	// Then check endpoint-specific limit
	endpointRules := r.getEndpointLimits(endpoint)
	if endpointRules == nil {
		// No specific limits for this endpoint
		return true, result, nil
	}

	// Apply graceful degradation
	load, _ := r.GetCurrentLoad(ctx)
	multiplier := r.getDegradationMultiplier(load, endpoint)
	adjustedRules := r.applyMultiplier(*endpointRules, multiplier)

	endpointKey := fmt.Sprintf("%s:endpoint:%s", identifier, endpoint)
	allowed, endpointResult, err := r.checkLimits(ctx, endpointKey, LimitTypeEndpoint, adjustedRules)
	if err != nil {
		return false, nil, err
	}

	r.recordMetric(allowed, LimitTypeEndpoint, endpointKey)

	if !allowed {
		return false, endpointResult, nil
	}

	return true, result, nil
}

// checkLimits checks all time window limits using token bucket algorithm
func (r *RedisRateLimiter) checkLimits(ctx context.Context, key string, limitType LimitType, rules LimitRules) (bool, *RateLimitResult, error) {
	windows := []struct {
		duration time.Duration
		limit    int
		name     string
	}{
		{time.Second, rules.RequestsPerSecond, "second"},
		{time.Minute, rules.RequestsPerMinute, "minute"},
		{time.Hour, rules.RequestsPerHour, "hour"},
		{24 * time.Hour, rules.RequestsPerDay, "day"},
	}

	for _, window := range windows {
		if window.limit == 0 {
			continue
		}

		redisKey := fmt.Sprintf("%s:%s:%s:%s", r.config.RedisPrefix, limitType, key, window.name)

		// Use token bucket algorithm with Redis
		allowed, remaining, resetAt, err := r.tokenBucket(ctx, redisKey, window.limit, rules.BurstSize, window.duration)
		if err != nil {
			return false, nil, fmt.Errorf("failed to check %s limit: %w", window.name, err)
		}

		if !allowed {
			retryAfter := int(time.Until(resetAt).Seconds())
			if retryAfter < 0 {
				retryAfter = 0
			}

			return false, &RateLimitResult{
				Allowed:    false,
				Limit:      window.limit,
				Remaining:  remaining,
				RetryAfter: retryAfter,
				ResetAt:    resetAt,
				LimitType:  limitType,
				Identifier: key,
			}, nil
		}
	}

	// All checks passed
	return true, &RateLimitResult{
		Allowed:    true,
		LimitType:  limitType,
		Identifier: key,
	}, nil
}

// tokenBucket implements token bucket algorithm using Redis
func (r *RedisRateLimiter) tokenBucket(ctx context.Context, key string, limit int, burst int, window time.Duration) (bool, int, time.Time, error) {
	if burst == 0 {
		burst = limit
	}

	now := time.Now()
	resetAt := now.Add(window)

	// Lua script for atomic token bucket operation
	script := `
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local burst = tonumber(ARGV[2])
		local window = tonumber(ARGV[3])
		local now = tonumber(ARGV[4])

		local current = redis.call('GET', key)
		if not current then
			redis.call('SET', key, burst - 1, 'EX', window)
			return {1, burst - 1, now + window}
		end

		current = tonumber(current)
		if current > 0 then
			redis.call('DECR', key)
			local ttl = redis.call('TTL', key)
			return {1, current - 1, now + ttl}
		end

		local ttl = redis.call('TTL', key)
		return {0, 0, now + ttl}
	`

	result, err := r.client.Eval(ctx, script, []string{key}, limit, burst, int(window.Seconds()), now.Unix()).Result()
	if err != nil {
		return false, 0, resetAt, err
	}

	resultSlice := result.([]interface{})
	allowed := resultSlice[0].(int64) == 1
	remainingInt64 := resultSlice[1].(int64)
	resetTimestamp := resultSlice[2].(int64)

	// Safe conversion with bounds checking to prevent integer overflow
	remaining := int(remainingInt64)
	if int64(remaining) != remainingInt64 {
		// Overflow occurred, clamp to max int
		if remainingInt64 > 0 {
			remaining = int(^uint(0) >> 1) // max int
		} else {
			remaining = 0
		}
	}

	return allowed, remaining, time.Unix(resetTimestamp, 0), nil
}

// RecordBypassAttempt records a potential bypass attempt
func (r *RedisRateLimiter) RecordBypassAttempt(ctx context.Context, identifier string, reason string) error {
	key := fmt.Sprintf("%s:bypass:%s", r.config.RedisPrefix, identifier)

	attempt := BypassAttempt{
		Identifier: identifier,
		Timestamp:  time.Now(),
		Reason:     reason,
		Count:      1,
	}

	// Increment counter
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return err
	}

	// Safe conversion with bounds checking
	attemptCount := int(count)
	if int64(attemptCount) != count {
		attemptCount = int(^uint(0) >> 1) // max int on overflow
	}
	attempt.Count = attemptCount

	// Set expiry to 1 minute
	r.client.Expire(ctx, key, time.Minute)

	// Record in metrics
	r.metrics.mu.Lock()
	r.metrics.bypassAttempts++
	r.metrics.mu.Unlock()

	// Check if threshold exceeded
	if count > int64(r.config.BypassDetection.MaxFailedAttemptsPerMinute) {
		r.logger.Warn().
			Str("identifier", identifier).
			Int64("count", count).
			Str("reason", reason).
			Msg("bypass attempt threshold exceeded")

		// Auto-ban
		if err := r.Ban(ctx, identifier, r.config.BypassDetection.BanDuration, "automated ban: excessive rate limit violations"); err != nil {
			return err
		}
	}

	// Alert if alert threshold exceeded
	if count > int64(r.config.BypassDetection.AlertThreshold) {
		r.logger.Error().
			Str("identifier", identifier).
			Int64("count", count).
			Str("reason", reason).
			Msg("ALERT: Potential DDoS attack detected")
	}

	return nil
}

// Ban temporarily bans an identifier
func (r *RedisRateLimiter) Ban(ctx context.Context, identifier string, duration time.Duration, reason string) error {
	key := fmt.Sprintf("%s:ban:%s", r.config.RedisPrefix, identifier)

	ban := BanRecord{
		Identifier: identifier,
		BannedAt:   time.Now(),
		ExpiresAt:  time.Now().Add(duration),
		Reason:     reason,
		Permanent:  duration == 0,
	}

	data, err := json.Marshal(ban)
	if err != nil {
		return err
	}

	if duration == 0 {
		// Permanent ban
		return r.client.Set(ctx, key, data, 0).Err()
	}

	r.logger.Warn().
		Str("identifier", identifier).
		Dur("duration", duration).
		Str("reason", reason).
		Msg("identifier banned")

	return r.client.Set(ctx, key, data, duration).Err()
}

// IsBanned checks if an identifier is banned
func (r *RedisRateLimiter) IsBanned(ctx context.Context, identifier string) (bool, error) {
	key := fmt.Sprintf("%s:ban:%s", r.config.RedisPrefix, identifier)
	exists, err := r.client.Exists(ctx, key).Result()
	return exists > 0, err
}

// GetMetrics returns rate limiting metrics
func (r *RedisRateLimiter) GetMetrics(ctx context.Context) (*Metrics, error) {
	r.metrics.mu.RLock()
	defer r.metrics.mu.RUnlock()

	// Count banned identifiers
	pattern := fmt.Sprintf("%s:ban:*", r.config.RedisPrefix)
	var cursor uint64
	bannedCount := 0
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			break
		}
		bannedCount += len(keys)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	// Get top blocked IPs
	topBlockedIPs := r.getTopBlocked(r.metrics.blockedIPCounts, 10)
	topBlockedUsers := r.getTopBlocked(r.metrics.blockedUserCounts, 10)

	// Get current load
	load, _ := r.GetCurrentLoad(ctx)

	return &Metrics{
		TotalRequests:     r.metrics.totalRequests,
		AllowedRequests:   r.metrics.allowedRequests,
		BlockedRequests:   r.metrics.blockedRequests,
		BypassAttempts:    r.metrics.bypassAttempts,
		BannedIdentifiers: bannedCount,
		CurrentLoad:       load,
		ByLimitType:       r.metrics.byLimitType,
		TopBlockedIPs:     topBlockedIPs,
		TopBlockedUsers:   topBlockedUsers,
	}, nil
}

// GetCurrentLoad returns the current system load (0-100)
func (r *RedisRateLimiter) GetCurrentLoad(ctx context.Context) (float64, error) {
	key := fmt.Sprintf("%s:metrics:requests_per_second", r.config.RedisPrefix)

	// Get requests in last second
	count, err := r.client.Get(ctx, key).Int64()
	if err != nil && err != redis.Nil {
		return 0, err
	}

	// Calculate load as percentage of global limit
	if r.config.GlobalLimits.RequestsPerSecond == 0 {
		return 0, nil
	}

	load := (float64(count) / float64(r.config.GlobalLimits.RequestsPerSecond)) * 100
	if load > 100 {
		load = 100
	}

	return load, nil
}

// IsWhitelisted checks if an identifier is whitelisted
func (r *RedisRateLimiter) IsWhitelisted(identifier string, limitType LimitType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	switch limitType {
	case LimitTypeIP:
		for _, ip := range r.config.WhitelistedIPs {
			if r.matchesIPPattern(identifier, ip) {
				return true
			}
		}
	case LimitTypeUser:
		for _, user := range r.config.WhitelistedUsers {
			if user == identifier {
				return true
			}
		}
	}

	return false
}

// UpdateConfig updates the rate limiter configuration
func (r *RedisRateLimiter) UpdateConfig(config RateLimitConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.config = config
	r.logger.Info().Msg("rate limiter configuration updated")

	return nil
}

// Close closes the rate limiter
func (r *RedisRateLimiter) Close() error {
	return r.client.Close()
}

// Helper methods

func (r *RedisRateLimiter) getLimitsForType(limitType LimitType) LimitRules {
	r.mu.RLock()
	defer r.mu.RUnlock()

	switch limitType {
	case LimitTypeIP:
		return r.config.IPLimits
	case LimitTypeUser:
		return r.config.UserLimits
	case LimitTypeGlobal:
		return r.config.GlobalLimits
	default:
		return r.config.IPLimits
	}
}

func (r *RedisRateLimiter) getEndpointLimits(endpoint string) *LimitRules {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Try exact match first
	if rules, ok := r.config.EndpointLimits[endpoint]; ok {
		return &rules
	}

	// Try pattern matching
	for pattern, rules := range r.config.EndpointLimits {
		if r.matchesPattern(endpoint, pattern) {
			return &rules
		}
	}

	return nil
}

func (r *RedisRateLimiter) matchesPattern(endpoint, pattern string) bool {
	// Simple wildcard matching (supports * at end)
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(endpoint, prefix)
	}
	return endpoint == pattern
}

func (r *RedisRateLimiter) matchesIPPattern(ip, pattern string) bool {
	// Check for exact match
	if ip == pattern {
		return true
	}

	// Check for CIDR match
	if strings.Contains(pattern, "/") {
		_, ipNet, err := net.ParseCIDR(pattern)
		if err != nil {
			return false
		}
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil {
			return false
		}
		return ipNet.Contains(parsedIP)
	}

	return false
}

func (r *RedisRateLimiter) getDegradationMultiplier(load float64, endpoint string) float64 {
	if !r.config.GracefulDegradation.Enabled {
		return 1.0
	}

	// Find applicable threshold
	for i := len(r.config.GracefulDegradation.LoadThresholds) - 1; i >= 0; i-- {
		threshold := r.config.GracefulDegradation.LoadThresholds[i]
		if load >= float64(threshold.LoadPercentage) {
			// Check if endpoint is priority
			if endpoint != "" && len(threshold.Priority) > 0 {
				for _, pattern := range threshold.Priority {
					if r.matchesPattern(endpoint, pattern) {
						return 1.0 // Priority endpoints not degraded
					}
				}
			}
			return threshold.RateLimitMultiplier
		}
	}

	return 1.0
}

func (r *RedisRateLimiter) applyMultiplier(rules LimitRules, multiplier float64) LimitRules {
	if multiplier == 1.0 {
		return rules
	}

	return LimitRules{
		RequestsPerSecond: int(float64(rules.RequestsPerSecond) * multiplier),
		RequestsPerMinute: int(float64(rules.RequestsPerMinute) * multiplier),
		RequestsPerHour:   int(float64(rules.RequestsPerHour) * multiplier),
		RequestsPerDay:    int(float64(rules.RequestsPerDay) * multiplier),
		BurstSize:         int(float64(rules.BurstSize) * multiplier),
	}
}

func (r *RedisRateLimiter) recordMetric(allowed bool, limitType LimitType, key string) {
	r.metrics.mu.Lock()
	defer r.metrics.mu.Unlock()

	r.metrics.totalRequests++
	if allowed {
		r.metrics.allowedRequests++
	} else {
		r.metrics.blockedRequests++

		// Track by type
		if limitType == LimitTypeIP {
			r.metrics.blockedIPCounts[key]++
		} else if limitType == LimitTypeUser {
			r.metrics.blockedUserCounts[key]++
		}
	}

	// Track by limit type
	if _, ok := r.metrics.byLimitType[limitType]; !ok {
		r.metrics.byLimitType[limitType] = &LimitTypeMetrics{}
	}
	metrics := r.metrics.byLimitType[limitType]
	metrics.Requests++
	if allowed {
		metrics.Allowed++
	} else {
		metrics.Blocked++
	}
}

func (r *RedisRateLimiter) getTopBlocked(counts map[string]uint64, limit int) []string {
	type kv struct {
		Key   string
		Value uint64
	}

	var sorted []kv
	for k, v := range counts {
		sorted = append(sorted, kv{k, v})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	var result []string
	for i := 0; i < len(sorted) && i < limit; i++ {
		result = append(result, sorted[i].Key)
	}

	return result
}
