package cache

import (
	"errors"
	"time"
)

// Backend represents the cache backend type.
type Backend string

const (
	// BackendMemory is an in-memory LRU cache.
	BackendMemory Backend = "memory"

	// BackendRedis is a Redis-based distributed cache.
	BackendRedis Backend = "redis"
)

// Config holds the configuration for the cache system.
type Config struct {
	// Enabled determines if caching is enabled.
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Backend specifies the cache backend to use.
	Backend Backend `json:"backend" yaml:"backend"`

	// DefaultTTL is the default time-to-live for cache entries.
	DefaultTTL time.Duration `json:"default_ttl" yaml:"default_ttl"`

	// MaxSize is the maximum number of entries for memory cache.
	MaxSize int `json:"max_size" yaml:"max_size"`

	// Redis configuration (only used when Backend is "redis").
	Redis RedisConfig `json:"redis" yaml:"redis"`

	// Metrics configuration.
	Metrics MetricsConfig `json:"metrics" yaml:"metrics"`

	// Specific cache TTLs for different data types.
	TTLs TTLConfig `json:"ttls" yaml:"ttls"`
}

// RedisConfig holds Redis-specific configuration.
type RedisConfig struct {
	// URL is the Redis connection URL.
	URL string `json:"url" yaml:"url"`

	// Password is the Redis password (optional).
	Password string `json:"password" yaml:"password"`

	// Database is the Redis database number.
	Database int `json:"database" yaml:"database"`

	// PoolSize is the maximum number of connections.
	PoolSize int `json:"pool_size" yaml:"pool_size"`

	// MinIdleConns is the minimum number of idle connections.
	MinIdleConns int `json:"min_idle_conns" yaml:"min_idle_conns"`

	// MaxRetries is the maximum number of retries on failure.
	MaxRetries int `json:"max_retries" yaml:"max_retries"`

	// DialTimeout is the timeout for establishing connections.
	DialTimeout time.Duration `json:"dial_timeout" yaml:"dial_timeout"`

	// ReadTimeout is the timeout for read operations.
	ReadTimeout time.Duration `json:"read_timeout" yaml:"read_timeout"`

	// WriteTimeout is the timeout for write operations.
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`

	// KeyPrefix is added to all cache keys.
	KeyPrefix string `json:"key_prefix" yaml:"key_prefix"`

	// TLSEnabled enables TLS for Redis connections.
	TLSEnabled bool `json:"tls_enabled" yaml:"tls_enabled"`
}

// MetricsConfig holds metrics collection configuration.
type MetricsConfig struct {
	// Enabled determines if metrics collection is enabled.
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Namespace is the metrics namespace prefix.
	Namespace string `json:"namespace" yaml:"namespace"`

	// ReportInterval is how often to report metrics.
	ReportInterval time.Duration `json:"report_interval" yaml:"report_interval"`
}

// TTLConfig holds TTL settings for different cache types.
type TTLConfig struct {
	// VEIDScore TTL for VEID score caching (relatively stable).
	VEIDScore time.Duration `json:"veid_score" yaml:"veid_score"`

	// ProviderInfo TTL for provider information caching.
	ProviderInfo time.Duration `json:"provider_info" yaml:"provider_info"`

	// ProviderOfferings TTL for provider offerings caching.
	ProviderOfferings time.Duration `json:"provider_offerings" yaml:"provider_offerings"`

	// MarketOrder TTL for active market orders caching.
	MarketOrder time.Duration `json:"market_order" yaml:"market_order"`

	// MarketBid TTL for market bids caching.
	MarketBid time.Duration `json:"market_bid" yaml:"market_bid"`

	// Lease TTL for active leases caching.
	Lease time.Duration `json:"lease" yaml:"lease"`
}

// DefaultConfig returns the default cache configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:    true,
		Backend:    BackendMemory,
		DefaultTTL: 5 * time.Minute,
		MaxSize:    10000,
		Redis: RedisConfig{
			URL:          "localhost:6379",
			Database:     0,
			PoolSize:     10,
			MinIdleConns: 2,
			MaxRetries:   3,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			KeyPrefix:    "virtengine:",
		},
		Metrics: MetricsConfig{
			Enabled:        true,
			Namespace:      "virtengine_cache",
			ReportInterval: 1 * time.Minute,
		},
		TTLs: DefaultTTLConfig(),
	}
}

// DefaultTTLConfig returns the default TTL configuration.
func DefaultTTLConfig() TTLConfig {
	return TTLConfig{
		VEIDScore:         5 * time.Minute,  // VEID scores are relatively stable
		ProviderInfo:      10 * time.Minute, // Provider info changes infrequently
		ProviderOfferings: 10 * time.Minute, // Provider offerings are relatively stable
		MarketOrder:       1 * time.Minute,  // Orders are more dynamic
		MarketBid:         30 * time.Second, // Bids change frequently
		Lease:             2 * time.Minute,  // Leases are relatively stable
	}
}

// Validate validates the cache configuration.
func (c *Config) Validate() error {
	if !c.Enabled {
		return nil // No validation needed when disabled
	}

	if c.Backend != BackendMemory && c.Backend != BackendRedis {
		return errors.New("cache: invalid backend, must be 'memory' or 'redis'")
	}

	if c.DefaultTTL < 0 {
		return errors.New("cache: default TTL cannot be negative")
	}

	if c.Backend == BackendMemory && c.MaxSize <= 0 {
		return errors.New("cache: max size must be positive for memory backend")
	}

	if c.Backend == BackendRedis {
		if err := c.Redis.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates the Redis configuration.
func (c *RedisConfig) Validate() error {
	if c.URL == "" {
		return errors.New("cache: Redis URL is required")
	}

	if c.PoolSize < 0 {
		return errors.New("cache: Redis pool size cannot be negative")
	}

	if c.MaxRetries < 0 {
		return errors.New("cache: Redis max retries cannot be negative")
	}

	return nil
}

