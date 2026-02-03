package pricefeed

import (
	"time"
)

// ============================================================================
// Configuration
// ============================================================================

// Config contains the configuration for the price feed service
type Config struct {
	// Providers is the list of enabled price providers in priority order
	Providers []ProviderConfig `json:"providers"`

	// Strategy is the aggregation strategy to use
	Strategy AggregationStrategy `json:"strategy"`

	// CacheConfig is the caching configuration
	CacheConfig CacheConfig `json:"cache_config"`

	// RetryConfig is the retry configuration
	RetryConfig RetryConfig `json:"retry_config"`

	// AssetMappings maps internal asset IDs to external source IDs
	AssetMappings []AssetMapping `json:"asset_mappings"`

	// MaxPriceDeviation is the max allowed price deviation between sources (0-1)
	// If deviation exceeds this, ErrPriceDeviation is returned
	MaxPriceDeviation float64 `json:"max_price_deviation"`

	// StaleThreshold is the max age for cached prices before they're considered stale
	StaleThreshold time.Duration `json:"stale_threshold"`

	// HealthCheckInterval is how often to check source health
	HealthCheckInterval time.Duration `json:"health_check_interval"`

	// EnableMetrics enables Prometheus metrics collection
	EnableMetrics bool `json:"enable_metrics"`

	// EnableLogging enables debug logging
	EnableLogging bool `json:"enable_logging"`
}

// ProviderConfig contains configuration for a single price provider
type ProviderConfig struct {
	// Name is the unique name for this provider instance
	Name string `json:"name"`

	// Type is the provider type
	Type SourceType `json:"type"`

	// Enabled controls whether this provider is active
	Enabled bool `json:"enabled"`

	// Priority is the priority order (lower = higher priority)
	Priority int `json:"priority"`

	// Weight is the weight for weighted aggregation (0-1)
	Weight float64 `json:"weight"`

	// CoinGeckoConfig is config for CoinGecko provider
	CoinGeckoConfig *CoinGeckoConfig `json:"coingecko_config,omitempty"`

	// ChainlinkConfig is config for Chainlink provider
	ChainlinkConfig *ChainlinkConfig `json:"chainlink_config,omitempty"`

	// PythConfig is config for Pyth provider
	PythConfig *PythConfig `json:"pyth_config,omitempty"`

	// RequestTimeout is the timeout for requests to this provider
	RequestTimeout time.Duration `json:"request_timeout"`
}

// CoinGeckoConfig contains CoinGecko-specific configuration
type CoinGeckoConfig struct {
	// APIURL is the base API URL
	APIURL string `json:"api_url"`

	// APIKey is the optional API key for Pro tier
	APIKey string `json:"api_key,omitempty"`

	// RateLimitPerMinute is the rate limit (free tier: 10-30/min)
	RateLimitPerMinute int `json:"rate_limit_per_minute"`

	// UsePro indicates whether to use the Pro API
	UsePro bool `json:"use_pro"`
}

// ChainlinkConfig contains Chainlink-specific configuration
type ChainlinkConfig struct {
	// RPCURL is the Ethereum RPC URL to query feeds
	RPCURL string `json:"rpc_url"`

	// NetworkID is the chain ID (1 = mainnet, 11155111 = sepolia)
	NetworkID int `json:"network_id"`

	// FeedAddresses maps asset pairs to feed contract addresses
	FeedAddresses map[string]string `json:"feed_addresses"`
}

// PythConfig contains Pyth-specific configuration
type PythConfig struct {
	// HermesURL is the Pyth Hermes API URL
	HermesURL string `json:"hermes_url"`

	// PriceServiceURL is the Pyth price service URL
	PriceServiceURL string `json:"price_service_url"`

	// PriceIDs maps asset pairs to Pyth price feed IDs
	PriceIDs map[string]string `json:"price_ids"`
}

// CacheConfig contains caching configuration
type CacheConfig struct {
	// Enabled enables caching
	Enabled bool `json:"enabled"`

	// TTL is the time-to-live for cached prices
	TTL time.Duration `json:"ttl"`

	// MaxSize is the maximum number of cached prices
	MaxSize int `json:"max_size"`

	// AllowStale allows returning stale prices when sources are unavailable
	AllowStale bool `json:"allow_stale"`

	// StaleMaxAge is the max age for stale prices (if AllowStale is true)
	StaleMaxAge time.Duration `json:"stale_max_age"`
}

// RetryConfig contains retry configuration
type RetryConfig struct {
	// MaxRetries is the maximum number of retries
	MaxRetries int `json:"max_retries"`

	// InitialDelay is the initial retry delay
	InitialDelay time.Duration `json:"initial_delay"`

	// MaxDelay is the maximum retry delay
	MaxDelay time.Duration `json:"max_delay"`

	// BackoffFactor is the exponential backoff multiplier
	BackoffFactor float64 `json:"backoff_factor"`

	// RetryableErrors is a list of errors that trigger retries
	RetryableErrors []string `json:"retryable_errors"`
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() Config {
	return Config{
		Providers: []ProviderConfig{
			{
				Name:     "coingecko-primary",
				Type:     SourceTypeCoinGecko,
				Enabled:  true,
				Priority: 1,
				Weight:   0.5,
				CoinGeckoConfig: &CoinGeckoConfig{
					APIURL:             "https://api.coingecko.com/api/v3",
					RateLimitPerMinute: 10,
					UsePro:             false,
				},
				RequestTimeout: 10 * time.Second,
			},
			{
				Name:     "chainlink-primary",
				Type:     SourceTypeChainlink,
				Enabled:  true,
				Priority: 2,
				Weight:   0.3,
				ChainlinkConfig: &ChainlinkConfig{
					NetworkID: 1, // Ethereum mainnet
					FeedAddresses: map[string]string{
						// Example feeds - would be configured per deployment
						"ETH/USD": "0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419",
					},
				},
				RequestTimeout: 15 * time.Second,
			},
			{
				Name:     "pyth-primary",
				Type:     SourceTypePyth,
				Enabled:  true,
				Priority: 3,
				Weight:   0.2,
				PythConfig: &PythConfig{
					HermesURL:       "https://hermes.pyth.network",
					PriceServiceURL: "https://xc-mainnet.pyth.network",
					PriceIDs: map[string]string{
						// Example feeds - would be configured per deployment
						"ETH/USD": "0xff61491a931112ddf1bd8147cd1b641375f79f5825126d665480874634fd0ace",
					},
				},
				RequestTimeout: 10 * time.Second,
			},
		},
		Strategy: StrategyPrimary,
		CacheConfig: CacheConfig{
			Enabled:     true,
			TTL:         30 * time.Second,
			MaxSize:     1000,
			AllowStale:  true,
			StaleMaxAge: 5 * time.Minute,
		},
		RetryConfig: RetryConfig{
			MaxRetries:    3,
			InitialDelay:  100 * time.Millisecond,
			MaxDelay:      5 * time.Second,
			BackoffFactor: 2.0,
			RetryableErrors: []string{
				"timeout",
				"connection refused",
				"rate limit",
				"503",
				"504",
			},
		},
		AssetMappings:       DefaultAssetMappings(),
		MaxPriceDeviation:   0.05, // 5% max deviation
		StaleThreshold:      1 * time.Minute,
		HealthCheckInterval: 30 * time.Second,
		EnableMetrics:       true,
		EnableLogging:       false,
	}
}

// Validate validates the configuration
func (c Config) Validate() error {
	if len(c.Providers) == 0 {
		return ErrPriceFeedUnavailable
	}

	if !c.Strategy.IsValid() {
		return ErrInvalidPair
	}

	enabledCount := 0
	for _, p := range c.Providers {
		if p.Enabled {
			enabledCount++
			if !p.Type.IsValid() {
				return ErrInvalidPair
			}
			if p.RequestTimeout <= 0 {
				return ErrInvalidPair
			}
		}
	}

	if enabledCount == 0 {
		return ErrPriceFeedUnavailable
	}

	if c.MaxPriceDeviation <= 0 || c.MaxPriceDeviation > 1 {
		return ErrInvalidPair
	}

	return nil
}

// GetProviderConfig returns the config for a specific provider by name
func (c Config) GetProviderConfig(name string) (ProviderConfig, bool) {
	for _, p := range c.Providers {
		if p.Name == name {
			return p, true
		}
	}
	return ProviderConfig{}, false
}

// GetAssetMapping returns the mapping for an internal asset ID
func (c Config) GetAssetMapping(internalID string) (AssetMapping, bool) {
	for _, m := range c.AssetMappings {
		if m.InternalID == internalID {
			return m, true
		}
	}
	return AssetMapping{}, false
}

// EnabledProviders returns only enabled providers sorted by priority
func (c Config) EnabledProviders() []ProviderConfig {
	var enabled []ProviderConfig
	for _, p := range c.Providers {
		if p.Enabled {
			enabled = append(enabled, p)
		}
	}
	// Sort by priority (lower = higher priority)
	for i := 0; i < len(enabled)-1; i++ {
		for j := i + 1; j < len(enabled); j++ {
			if enabled[j].Priority < enabled[i].Priority {
				enabled[i], enabled[j] = enabled[j], enabled[i]
			}
		}
	}
	return enabled
}
