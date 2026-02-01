// Package dex provides DEX integration for crypto-to-fiat conversions.
//
// VE-905: DEX integration for crypto-to-fiat conversions
package dex

import (
	"time"

	sdkmath "cosmossdk.io/math"
)

// Config holds the DEX service configuration
type Config struct {
	// PriceFeed configures price feed behavior
	PriceFeed PriceFeedConfig `json:"price_feed"`

	// Swap configures swap execution
	Swap SwapConfig `json:"swap"`

	// OffRamp configures off-ramp behavior
	OffRamp OffRampConfig `json:"off_ramp"`

	// CircuitBreaker configures safety circuit breakers
	CircuitBreaker CircuitBreakerConfig `json:"circuit_breaker"`

	// Adapters lists enabled DEX adapters
	Adapters []AdapterConfig `json:"adapters"`
}

// PriceFeedConfig configures price feed behavior
type PriceFeedConfig struct {
	// UpdateInterval is how often to refresh prices
	UpdateInterval time.Duration `json:"update_interval"`

	// MaxPriceAge is the maximum acceptable price age
	MaxPriceAge time.Duration `json:"max_price_age"`

	// MinSources is the minimum number of price sources required
	MinSources int `json:"min_sources"`

	// MaxDeviation is the maximum allowed deviation between sources
	MaxDeviation float64 `json:"max_deviation"`

	// TWAPWindow is the TWAP calculation window
	TWAPWindow time.Duration `json:"twap_window"`

	// VWAPWindow is the VWAP calculation window
	VWAPWindow time.Duration `json:"vwap_window"`

	// CacheEnabled enables price caching
	CacheEnabled bool `json:"cache_enabled"`

	// CacheTTL is the cache time-to-live
	CacheTTL time.Duration `json:"cache_ttl"`
}

// SwapConfig configures swap execution
type SwapConfig struct {
	// DefaultSlippage is the default slippage tolerance
	DefaultSlippage float64 `json:"default_slippage"`

	// MaxSlippage is the maximum allowed slippage
	MaxSlippage float64 `json:"max_slippage"`

	// QuoteValidityPeriod is how long quotes remain valid
	QuoteValidityPeriod time.Duration `json:"quote_validity_period"`

	// MaxHops is the maximum number of swap hops
	MaxHops int `json:"max_hops"`

	// GasMultiplier is applied to gas estimates
	GasMultiplier float64 `json:"gas_multiplier"`

	// MinLiquidityRatio is the minimum pool liquidity ratio for swaps
	MinLiquidityRatio float64 `json:"min_liquidity_ratio"`

	// EnableRouteFinding enables multi-hop route optimization
	EnableRouteFinding bool `json:"enable_route_finding"`

	// EnableMEVProtection enables MEV protection features
	EnableMEVProtection bool `json:"enable_mev_protection"`
}

// OffRampConfig configures off-ramp behavior
type OffRampConfig struct {
	// MinAmount is the minimum off-ramp amount (in base units)
	MinAmount sdkmath.Int `json:"min_amount"`

	// MaxAmount is the maximum off-ramp amount (in base units)
	MaxAmount sdkmath.Int `json:"max_amount"`

	// MinVEIDScore is the minimum VEID score required
	MinVEIDScore int64 `json:"min_veid_score"`

	// QuoteValidityPeriod is how long off-ramp quotes remain valid
	QuoteValidityPeriod time.Duration `json:"quote_validity_period"`

	// SupportedCurrencies is the list of supported fiat currencies
	SupportedCurrencies []FiatCurrency `json:"supported_currencies"`

	// SupportedMethods is the list of supported payment methods
	SupportedMethods []PaymentMethod `json:"supported_methods"`

	// Providers is the list of off-ramp provider configs
	Providers []OffRampProviderConfig `json:"providers"`
}

// OffRampProviderConfig configures an off-ramp partner
type OffRampProviderConfig struct {
	// Name is the provider name
	Name string `json:"name"`

	// Enabled indicates if the provider is active
	Enabled bool `json:"enabled"`

	// APIEndpoint is the provider API endpoint
	APIEndpoint string `json:"api_endpoint"`

	// SupportedCurrencies lists supported fiat currencies
	SupportedCurrencies []FiatCurrency `json:"supported_currencies"`

	// SupportedMethods lists supported payment methods
	SupportedMethods []PaymentMethod `json:"supported_methods"`

	// FeePercent is the provider fee percentage
	FeePercent sdkmath.LegacyDec `json:"fee_percent"`

	// MinAmount is the provider minimum amount
	MinAmount sdkmath.Int `json:"min_amount"`

	// MaxAmount is the provider maximum amount
	MaxAmount sdkmath.Int `json:"max_amount"`
}

// CircuitBreakerConfig configures safety circuit breakers
type CircuitBreakerConfig struct {
	// Enabled enables circuit breaker functionality
	Enabled bool `json:"enabled"`

	// PriceDeviationThreshold triggers on price deviation from TWAP
	PriceDeviationThreshold float64 `json:"price_deviation_threshold"`

	// VolumeSpikeFactor triggers on abnormal volume
	VolumeSpikeFactor float64 `json:"volume_spike_factor"`

	// CooldownPeriod is the pause period after trigger
	CooldownPeriod time.Duration `json:"cooldown_period"`

	// MaxFailuresPerMinute triggers on repeated failures
	MaxFailuresPerMinute int `json:"max_failures_per_minute"`

	// RecoveryThreshold is the success count needed to reset
	RecoveryThreshold int `json:"recovery_threshold"`
}

// AdapterConfig configures a DEX adapter
type AdapterConfig struct {
	// Name is the adapter name
	Name string `json:"name"`

	// Type is the adapter type (uniswap, osmosis, curve, etc.)
	Type string `json:"type"`

	// Enabled indicates if the adapter is active
	Enabled bool `json:"enabled"`

	// Priority sets the adapter priority (lower = higher priority)
	Priority int `json:"priority"`

	// ChainID is the blockchain chain ID
	ChainID string `json:"chain_id"`

	// RPCEndpoint is the RPC endpoint for the chain
	RPCEndpoint string `json:"rpc_endpoint"`

	// ContractAddresses maps contract names to addresses
	ContractAddresses map[string]string `json:"contract_addresses"`

	// Timeout is the adapter operation timeout
	Timeout time.Duration `json:"timeout"`

	// MaxRetries is the maximum retry count
	MaxRetries int `json:"max_retries"`

	// RetryDelay is the delay between retries
	RetryDelay time.Duration `json:"retry_delay"`
}

// DefaultConfig returns sensible default configuration
func DefaultConfig() Config {
	return Config{
		PriceFeed: PriceFeedConfig{
			UpdateInterval: 30 * time.Second,
			MaxPriceAge:    5 * time.Minute,
			MinSources:     2,
			MaxDeviation:   0.05, // 5% maximum deviation
			TWAPWindow:     1 * time.Hour,
			VWAPWindow:     24 * time.Hour,
			CacheEnabled:   true,
			CacheTTL:       10 * time.Second,
		},
		Swap: SwapConfig{
			DefaultSlippage:     0.005, // 0.5%
			MaxSlippage:         0.10,  // 10%
			QuoteValidityPeriod: 30 * time.Second,
			MaxHops:             4,
			GasMultiplier:       1.2,
			MinLiquidityRatio:   0.01, // 1% of pool
			EnableRouteFinding:  true,
			EnableMEVProtection: true,
		},
		OffRamp: OffRampConfig{
			MinAmount:           sdkmath.NewInt(100_000_000), // 100 tokens (6 decimals)
			MaxAmount:           sdkmath.NewInt(100_000_000_000_000),
			MinVEIDScore:        500, // Minimum identity score
			QuoteValidityPeriod: 5 * time.Minute,
			SupportedCurrencies: []FiatCurrency{FiatUSD, FiatEUR, FiatGBP, FiatAUD},
			SupportedMethods:    []PaymentMethod{PaymentMethodBankTransfer, PaymentMethodSEPA, PaymentMethodACH},
		},
		CircuitBreaker: CircuitBreakerConfig{
			Enabled:                 true,
			PriceDeviationThreshold: 0.15, // 15% deviation
			VolumeSpikeFactor:       10.0, // 10x normal volume
			CooldownPeriod:          5 * time.Minute,
			MaxFailuresPerMinute:    10,
			RecoveryThreshold:       5,
		},
	}
}

// Validate validates the configuration
func (c Config) Validate() error {
	if c.PriceFeed.MinSources < 1 {
		return ErrInvalidConfig{Field: "price_feed.min_sources", Reason: "must be at least 1"}
	}
	if c.PriceFeed.MaxDeviation <= 0 || c.PriceFeed.MaxDeviation >= 1 {
		return ErrInvalidConfig{Field: "price_feed.max_deviation", Reason: "must be between 0 and 1"}
	}
	if c.Swap.DefaultSlippage <= 0 || c.Swap.DefaultSlippage > c.Swap.MaxSlippage {
		return ErrInvalidConfig{Field: "swap.default_slippage", Reason: "must be positive and less than max_slippage"}
	}
	if c.Swap.MaxHops < 1 {
		return ErrInvalidConfig{Field: "swap.max_hops", Reason: "must be at least 1"}
	}
	return nil
}

// ErrInvalidConfig represents a configuration error
type ErrInvalidConfig struct {
	Field  string
	Reason string
}

func (e ErrInvalidConfig) Error() string {
	return "invalid config: " + e.Field + ": " + e.Reason
}

