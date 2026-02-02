package pricefeed

import (
	"errors"
	"time"

	sdkmath "cosmossdk.io/math"
)

// ============================================================================
// Errors
// ============================================================================

var (
	// ErrPriceFeedUnavailable is returned when no price feed sources are available
	ErrPriceFeedUnavailable = errors.New("price feed unavailable")

	// ErrPriceNotFound is returned when a price for the requested pair is not found
	ErrPriceNotFound = errors.New("price not found for requested pair")

	// ErrPriceStale is returned when the cached price is too old
	ErrPriceStale = errors.New("price data is stale")

	// ErrInvalidPair is returned for unsupported trading pairs
	ErrInvalidPair = errors.New("invalid or unsupported trading pair")

	// ErrRateLimitExceeded is returned when API rate limits are hit
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrSourceUnhealthy is returned when a price source is not responding
	ErrSourceUnhealthy = errors.New("price source is unhealthy")

	// ErrAllSourcesFailed is returned when all configured sources fail
	ErrAllSourcesFailed = errors.New("all price feed sources failed")

	// ErrPriceDeviation is returned when prices deviate too much between sources
	ErrPriceDeviation = errors.New("price deviation exceeds threshold")
)

// ============================================================================
// Price Types
// ============================================================================

// PriceData represents a price quote from a price feed source
type PriceData struct {
	// BaseAsset is the asset being priced (e.g., "virtengine", "uve")
	BaseAsset string `json:"base_asset"`

	// QuoteAsset is the quote currency (e.g., "usd", "eur")
	QuoteAsset string `json:"quote_asset"`

	// Price is the price of 1 unit of base asset in quote currency
	Price sdkmath.LegacyDec `json:"price"`

	// Timestamp is when the price was fetched
	Timestamp time.Time `json:"timestamp"`

	// Source is the identifier of the price source
	Source string `json:"source"`

	// Confidence is a 0-1 score indicating price reliability (optional)
	Confidence float64 `json:"confidence,omitempty"`

	// Volume24h is the 24-hour trading volume (optional)
	Volume24h sdkmath.LegacyDec `json:"volume_24h,omitempty"`

	// MarketCap is the market capitalization (optional)
	MarketCap sdkmath.LegacyDec `json:"market_cap,omitempty"`

	// LastUpdatedAt is when the source last updated this price
	LastUpdatedAt time.Time `json:"last_updated_at,omitempty"`
}

// IsValid checks if the price data is valid
func (p PriceData) IsValid() bool {
	if p.BaseAsset == "" || p.QuoteAsset == "" {
		return false
	}
	if p.Price.IsNil() || p.Price.IsNegative() || p.Price.IsZero() {
		return false
	}
	if p.Timestamp.IsZero() {
		return false
	}
	return true
}

// Age returns how old the price data is
func (p PriceData) Age() time.Duration {
	return time.Since(p.Timestamp)
}

// IsStale checks if the price is older than the given threshold
func (p PriceData) IsStale(maxAge time.Duration) bool {
	return p.Age() > maxAge
}

// ============================================================================
// Source Types
// ============================================================================

// SourceType identifies the type of price feed source
type SourceType string

const (
	// SourceTypeCoinGecko is the CoinGecko API source
	SourceTypeCoinGecko SourceType = "coingecko"

	// SourceTypeChainlink is the Chainlink oracle source
	SourceTypeChainlink SourceType = "chainlink"

	// SourceTypePyth is the Pyth network source
	SourceTypePyth SourceType = "pyth"

	// SourceTypeMock is a mock source for testing
	SourceTypeMock SourceType = "mock"
)

// String returns the string representation
func (s SourceType) String() string {
	return string(s)
}

// IsValid checks if the source type is valid
func (s SourceType) IsValid() bool {
	switch s {
	case SourceTypeCoinGecko, SourceTypeChainlink, SourceTypePyth, SourceTypeMock:
		return true
	default:
		return false
	}
}

// SourceHealth represents the health status of a price source
type SourceHealth struct {
	// Source is the source identifier
	Source string `json:"source"`

	// Type is the source type
	Type SourceType `json:"type"`

	// Healthy indicates if the source is responding
	Healthy bool `json:"healthy"`

	// LastCheck is when health was last checked
	LastCheck time.Time `json:"last_check"`

	// LastSuccess is when the source last returned valid data
	LastSuccess time.Time `json:"last_success"`

	// LastError is the most recent error (if any)
	LastError string `json:"last_error,omitempty"`

	// Latency is the average response time
	Latency time.Duration `json:"latency"`

	// RequestCount is total requests made
	RequestCount int64 `json:"request_count"`

	// ErrorCount is total errors
	ErrorCount int64 `json:"error_count"`
}

// ErrorRate returns the error rate as a percentage
func (h SourceHealth) ErrorRate() float64 {
	if h.RequestCount == 0 {
		return 0
	}
	return float64(h.ErrorCount) / float64(h.RequestCount) * 100
}

// ============================================================================
// Aggregation Types
// ============================================================================

// AggregationStrategy defines how to combine prices from multiple sources
type AggregationStrategy string

const (
	// StrategyPrimary uses the first healthy source in priority order
	StrategyPrimary AggregationStrategy = "primary"

	// StrategyMedian uses the median price across all sources
	StrategyMedian AggregationStrategy = "median"

	// StrategyWeighted weights prices by source confidence/volume
	StrategyWeighted AggregationStrategy = "weighted"
)

// String returns the string representation
func (s AggregationStrategy) String() string {
	return string(s)
}

// IsValid checks if the strategy is valid
func (s AggregationStrategy) IsValid() bool {
	switch s {
	case StrategyPrimary, StrategyMedian, StrategyWeighted:
		return true
	default:
		return false
	}
}

// AggregatedPrice represents a price derived from multiple sources
type AggregatedPrice struct {
	PriceData

	// Strategy is the aggregation strategy used
	Strategy AggregationStrategy `json:"strategy"`

	// SourcePrices contains prices from each individual source
	SourcePrices []PriceData `json:"source_prices"`

	// Deviation is the max percentage deviation between sources
	Deviation float64 `json:"deviation"`

	// SourceCount is the number of sources that contributed
	SourceCount int `json:"source_count"`
}

// ============================================================================
// Asset Mapping Types
// ============================================================================

// AssetMapping maps internal asset identifiers to external source identifiers
type AssetMapping struct {
	// InternalID is the asset identifier used internally (e.g., "uve")
	InternalID string `json:"internal_id"`

	// CoinGeckoID is the CoinGecko asset ID (e.g., "virtengine")
	CoinGeckoID string `json:"coingecko_id,omitempty"`

	// ChainlinkFeedAddress is the Chainlink price feed contract address
	ChainlinkFeedAddress string `json:"chainlink_feed_address,omitempty"`

	// PythPriceID is the Pyth price feed ID (32-byte hex)
	PythPriceID string `json:"pyth_price_id,omitempty"`

	// Decimals is the number of decimals for this asset
	Decimals int `json:"decimals"`
}

// DefaultAssetMappings returns default mappings for common assets
func DefaultAssetMappings() []AssetMapping {
	return []AssetMapping{
		{
			InternalID:  "uve",
			CoinGeckoID: "virtengine",
			Decimals:    6,
			// Chainlink/Pyth IDs would be added when feeds are available
		},
		{
			InternalID:           "atom",
			CoinGeckoID:          "cosmos",
			ChainlinkFeedAddress: "0x...", // Example
			Decimals:             6,
		},
		{
			InternalID:  "usdc",
			CoinGeckoID: "usd-coin",
			Decimals:    6,
		},
	}
}
