// Package ratelimit provides comprehensive rate limiting for API endpoints
package ratelimit

import (
	"context"
	"time"
)

// LimitType represents the type of rate limit
type LimitType string

const (
	// LimitTypeIP is IP-based rate limiting
	LimitTypeIP LimitType = "ip"
	// LimitTypeUser is user/account-based rate limiting
	LimitTypeUser LimitType = "user"
	// LimitTypeEndpoint is endpoint-based rate limiting
	LimitTypeEndpoint LimitType = "endpoint"
	// LimitTypeGlobal is global system-wide rate limiting
	LimitTypeGlobal LimitType = "global"
)

// RateLimitConfig defines configuration for a rate limiter
type RateLimitConfig struct {
	// RedisURL is the connection string for Redis
	RedisURL string `json:"redis_url"`

	// RedisPrefix is the key prefix for Redis keys
	RedisPrefix string `json:"redis_prefix"`

	// Enabled determines if rate limiting is active
	Enabled bool `json:"enabled"`

	// IPLimits defines IP-based rate limits
	IPLimits LimitRules `json:"ip_limits"`

	// UserLimits defines user-based rate limits
	UserLimits LimitRules `json:"user_limits"`

	// EndpointLimits defines per-endpoint rate limits
	EndpointLimits map[string]LimitRules `json:"endpoint_limits"`

	// GlobalLimits defines system-wide rate limits
	GlobalLimits LimitRules `json:"global_limits"`

	// WhitelistedIPs is a list of IPs exempt from rate limiting
	WhitelistedIPs []string `json:"whitelisted_ips"`

	// WhitelistedUsers is a list of users exempt from rate limiting
	WhitelistedUsers []string `json:"whitelisted_users"`

	// BypassDetection configures bypass attempt detection
	BypassDetection BypassDetectionConfig `json:"bypass_detection"`

	// GracefulDegradation configures graceful degradation settings
	GracefulDegradation DegradationConfig `json:"graceful_degradation"`
}

// LimitRules defines the rules for a rate limit
type LimitRules struct {
	// RequestsPerSecond is the max requests per second
	RequestsPerSecond int `json:"requests_per_second"`

	// RequestsPerMinute is the max requests per minute
	RequestsPerMinute int `json:"requests_per_minute"`

	// RequestsPerHour is the max requests per hour
	RequestsPerHour int `json:"requests_per_hour"`

	// RequestsPerDay is the max requests per day
	RequestsPerDay int `json:"requests_per_day"`

	// BurstSize is the max burst size (token bucket)
	BurstSize int `json:"burst_size"`
}

// BypassDetectionConfig configures bypass attempt detection
type BypassDetectionConfig struct {
	// Enabled determines if bypass detection is active
	Enabled bool `json:"enabled"`

	// MaxFailedAttemptsPerMinute is the threshold for failed rate limit attempts
	MaxFailedAttemptsPerMinute int `json:"max_failed_attempts_per_minute"`

	// BanDuration is how long to ban detected bypass attempts
	BanDuration time.Duration `json:"ban_duration"`

	// AlertThreshold is the number of bypass attempts before alerting
	AlertThreshold int `json:"alert_threshold"`
}

// DegradationConfig configures graceful degradation under load
type DegradationConfig struct {
	// Enabled determines if graceful degradation is active
	Enabled bool `json:"enabled"`

	// LoadThresholds defines load levels and corresponding actions
	LoadThresholds []LoadThreshold `json:"load_thresholds"`

	// ReadOnlyMode enables read-only mode under extreme load
	ReadOnlyMode bool `json:"read_only_mode"`
}

// LoadThreshold defines a load level and action
type LoadThreshold struct {
	// LoadPercentage is the load level (0-100)
	LoadPercentage int `json:"load_percentage"`

	// RateLimitMultiplier is how much to reduce limits (e.g., 0.5 = 50% of normal)
	RateLimitMultiplier float64 `json:"rate_limit_multiplier"`

	// Priority defines which endpoints to prioritize
	Priority []string `json:"priority"`
}

// RateLimiter defines the interface for rate limiting
type RateLimiter interface {
	// Allow checks if a request is allowed
	Allow(ctx context.Context, key string, limitType LimitType) (bool, *RateLimitResult, error)

	// AllowEndpoint checks if a request to a specific endpoint is allowed
	AllowEndpoint(ctx context.Context, endpoint string, identifier string, limitType LimitType) (bool, *RateLimitResult, error)

	// RecordBypassAttempt records a potential bypass attempt
	RecordBypassAttempt(ctx context.Context, identifier string, reason string) error

	// GetMetrics returns rate limiting metrics
	GetMetrics(ctx context.Context) (*Metrics, error)

	// UpdateConfig updates the rate limiter configuration
	UpdateConfig(config RateLimitConfig) error

	// IsWhitelisted checks if an identifier is whitelisted
	IsWhitelisted(identifier string, limitType LimitType) bool

	// Ban temporarily bans an identifier
	Ban(ctx context.Context, identifier string, duration time.Duration, reason string) error

	// IsBanned checks if an identifier is banned
	IsBanned(ctx context.Context, identifier string) (bool, error)

	// GetCurrentLoad returns the current system load
	GetCurrentLoad(ctx context.Context) (float64, error)

	// Close closes the rate limiter
	Close() error
}

// RateLimitResult contains the result of a rate limit check
type RateLimitResult struct {
	// Allowed indicates if the request is allowed
	Allowed bool `json:"allowed"`

	// Limit is the rate limit ceiling
	Limit int `json:"limit"`

	// Remaining is the number of requests remaining
	Remaining int `json:"remaining"`

	// RetryAfter is when the client can retry (seconds)
	RetryAfter int `json:"retry_after"`

	// ResetAt is when the limit resets
	ResetAt time.Time `json:"reset_at"`

	// LimitType is the type of limit applied
	LimitType LimitType `json:"limit_type"`

	// Identifier is the key that was rate limited
	Identifier string `json:"identifier"`
}

// Metrics contains rate limiting metrics
type Metrics struct {
	// TotalRequests is the total number of requests processed
	TotalRequests uint64 `json:"total_requests"`

	// AllowedRequests is the number of allowed requests
	AllowedRequests uint64 `json:"allowed_requests"`

	// BlockedRequests is the number of blocked requests
	BlockedRequests uint64 `json:"blocked_requests"`

	// BypassAttempts is the number of detected bypass attempts
	BypassAttempts uint64 `json:"bypass_attempts"`

	// BannedIdentifiers is the number of currently banned identifiers
	BannedIdentifiers int `json:"banned_identifiers"`

	// CurrentLoad is the current system load (0-100)
	CurrentLoad float64 `json:"current_load"`

	// ByLimitType breaks down metrics by limit type
	ByLimitType map[LimitType]*LimitTypeMetrics `json:"by_limit_type"`

	// TopBlockedIPs is a list of most blocked IPs
	TopBlockedIPs []string `json:"top_blocked_ips"`

	// TopBlockedUsers is a list of most blocked users
	TopBlockedUsers []string `json:"top_blocked_users"`
}

// LimitTypeMetrics contains metrics for a specific limit type
type LimitTypeMetrics struct {
	// Requests is the total requests for this limit type
	Requests uint64 `json:"requests"`

	// Allowed is the allowed requests for this limit type
	Allowed uint64 `json:"allowed"`

	// Blocked is the blocked requests for this limit type
	Blocked uint64 `json:"blocked"`
}

// BypassAttempt represents a detected bypass attempt
type BypassAttempt struct {
	// Identifier is the IP or user attempting bypass
	Identifier string `json:"identifier"`

	// Timestamp is when the attempt occurred
	Timestamp time.Time `json:"timestamp"`

	// Reason is why it was flagged as a bypass attempt
	Reason string `json:"reason"`

	// Count is the number of attempts
	Count int `json:"count"`
}

// BanRecord represents a banned identifier
type BanRecord struct {
	// Identifier is the banned IP or user
	Identifier string `json:"identifier"`

	// BannedAt is when the ban started
	BannedAt time.Time `json:"banned_at"`

	// ExpiresAt is when the ban expires
	ExpiresAt time.Time `json:"expires_at"`

	// Reason is the reason for the ban
	Reason string `json:"reason"`

	// Permanent indicates if the ban is permanent
	Permanent bool `json:"permanent"`
}

// DefaultConfig returns a default rate limit configuration
func DefaultConfig() RateLimitConfig {
	return RateLimitConfig{
		RedisURL:    "redis://localhost:6379/0",
		RedisPrefix: "virtengine:ratelimit",
		Enabled:     true,
		IPLimits: LimitRules{
			RequestsPerSecond: 10,
			RequestsPerMinute: 300,
			RequestsPerHour:   10000,
			RequestsPerDay:    100000,
			BurstSize:         20,
		},
		UserLimits: LimitRules{
			RequestsPerSecond: 50,
			RequestsPerMinute: 1000,
			RequestsPerHour:   50000,
			RequestsPerDay:    500000,
			BurstSize:         100,
		},
		EndpointLimits: map[string]LimitRules{
			"/market/*": {
				RequestsPerSecond: 20,
				RequestsPerMinute: 600,
				RequestsPerHour:   20000,
				BurstSize:         40,
			},
			"/veid/*": {
				RequestsPerSecond: 5,
				RequestsPerMinute: 100,
				RequestsPerHour:   1000,
				BurstSize:         10,
			},
		},
		GlobalLimits: LimitRules{
			RequestsPerSecond: 10000,
			RequestsPerMinute: 300000,
			RequestsPerHour:   5000000,
			BurstSize:         20000,
		},
		WhitelistedIPs:   []string{},
		WhitelistedUsers: []string{},
		BypassDetection: BypassDetectionConfig{
			Enabled:                    true,
			MaxFailedAttemptsPerMinute: 100,
			BanDuration:                time.Hour,
			AlertThreshold:             50,
		},
		GracefulDegradation: DegradationConfig{
			Enabled: true,
			LoadThresholds: []LoadThreshold{
				{
					LoadPercentage:      80,
					RateLimitMultiplier: 0.7,
					Priority:            []string{"/veid/*", "/market/order/*"},
				},
				{
					LoadPercentage:      90,
					RateLimitMultiplier: 0.5,
					Priority:            []string{"/veid/*"},
				},
				{
					LoadPercentage:      95,
					RateLimitMultiplier: 0.3,
					Priority:            []string{"/veid/verify"},
				},
			},
			ReadOnlyMode: true,
		},
	}
}

