// Package dex provides DEX integration for crypto-to-fiat conversions.
//
// VE-905: DEX integration for crypto-to-fiat conversions
package dex

import (
	"context"
	"time"

	sdkmath "cosmossdk.io/math"
)

// ============================================================================
// DEX Adapter Interface
// ============================================================================

// Adapter defines the interface for DEX protocol adapters.
// Each adapter implements protocol-specific logic for a DEX (Uniswap, Osmosis, etc.)
type Adapter interface {
	// Name returns the adapter name
	Name() string

	// Type returns the adapter type (uniswap_v2, uniswap_v3, osmosis, curve, etc.)
	Type() string

	// ChainID returns the chain ID this adapter operates on
	ChainID() string

	// IsHealthy checks if the adapter is operational
	IsHealthy(ctx context.Context) bool

	// GetSupportedPairs returns all supported trading pairs
	GetSupportedPairs(ctx context.Context) ([]TradingPair, error)

	// GetPrice fetches the current price for a trading pair
	GetPrice(ctx context.Context, baseSymbol, quoteSymbol string) (Price, error)

	// GetPool fetches liquidity pool information
	GetPool(ctx context.Context, poolID string) (LiquidityPool, error)

	// ListPools lists liquidity pools matching the query
	ListPools(ctx context.Context, query PoolQuery) ([]LiquidityPool, error)

	// GetSwapQuote generates a swap quote
	GetSwapQuote(ctx context.Context, request SwapRequest) (SwapQuote, error)

	// ExecuteSwap executes a swap (requires signed transaction)
	ExecuteSwap(ctx context.Context, quote SwapQuote, signedTx []byte) (SwapResult, error)

	// EstimateGas estimates gas for a swap
	EstimateGas(ctx context.Context, request SwapRequest) (uint64, error)

	// Close releases adapter resources
	Close() error
}

// ============================================================================
// Price Feed Interface
// ============================================================================

// PriceFeed provides price data aggregation from multiple sources
type PriceFeed interface {
	// GetPrice fetches the current price for a trading pair
	GetPrice(ctx context.Context, baseSymbol, quoteSymbol string) (Price, error)

	// GetPriceAggregate fetches aggregated price data from multiple sources
	GetPriceAggregate(ctx context.Context, baseSymbol, quoteSymbol string) (PriceAggregate, error)

	// GetTWAP fetches the time-weighted average price
	GetTWAP(ctx context.Context, baseSymbol, quoteSymbol string, window time.Duration) (sdkmath.LegacyDec, error)

	// GetVWAP fetches the volume-weighted average price
	GetVWAP(ctx context.Context, baseSymbol, quoteSymbol string, window time.Duration) (sdkmath.LegacyDec, error)

	// SubscribePrice subscribes to price updates
	SubscribePrice(ctx context.Context, baseSymbol, quoteSymbol string, callback PriceCallback) (Subscription, error)

	// RegisterSource registers a price data source
	RegisterSource(source PriceSource) error

	// UnregisterSource removes a price data source
	UnregisterSource(name string) error
}

// PriceCallback is called when price updates are received
type PriceCallback func(Price)

// Subscription represents a price subscription
type Subscription interface {
	// Unsubscribe cancels the subscription
	Unsubscribe()
}

// PriceSource provides price data from a single source
type PriceSource interface {
	// Name returns the source name
	Name() string

	// GetPrice fetches the current price
	GetPrice(ctx context.Context, baseSymbol, quoteSymbol string) (Price, error)

	// IsHealthy checks if the source is operational
	IsHealthy(ctx context.Context) bool
}

// ============================================================================
// Swap Executor Interface
// ============================================================================

// SwapExecutor handles swap quote generation and execution
type SwapExecutor interface {
	// GetQuote generates a swap quote with optimal routing
	GetQuote(ctx context.Context, request SwapRequest) (SwapQuote, error)

	// ExecuteSwap executes a previously quoted swap
	ExecuteSwap(ctx context.Context, quote SwapQuote, signedTx []byte) (SwapResult, error)

	// FindBestRoute finds the optimal swap route
	FindBestRoute(ctx context.Context, request SwapRequest) (SwapRoute, error)

	// ValidateQuote validates if a quote is still executable
	ValidateQuote(ctx context.Context, quote SwapQuote) error

	// GetSlippageEstimate estimates potential slippage
	GetSlippageEstimate(ctx context.Context, request SwapRequest) (float64, error)
}

// ============================================================================
// Off-Ramp Bridge Interface
// ============================================================================

// OffRampBridge handles crypto-to-fiat conversions
type OffRampBridge interface {
	// GetQuote generates an off-ramp quote
	GetQuote(ctx context.Context, request OffRampRequest) (OffRampQuote, error)

	// InitiateOffRamp initiates an off-ramp operation
	InitiateOffRamp(ctx context.Context, quote OffRampQuote, signedTx []byte) (OffRampResult, error)

	// GetStatus fetches the status of an off-ramp operation
	GetStatus(ctx context.Context, offRampID string) (OffRampResult, error)

	// CancelOffRamp cancels a pending off-ramp operation
	CancelOffRamp(ctx context.Context, offRampID string) error

	// ListOperations lists off-ramp operations for an address
	ListOperations(ctx context.Context, address string, limit, offset int) ([]OffRampResult, error)

	// GetSupportedCurrencies returns supported fiat currencies
	GetSupportedCurrencies(ctx context.Context) ([]FiatCurrency, error)

	// GetSupportedMethods returns supported payment methods
	GetSupportedMethods(ctx context.Context, currency FiatCurrency) ([]PaymentMethod, error)

	// ValidateKYC validates if an address has sufficient KYC
	ValidateKYC(ctx context.Context, address string, veIDScore int64) error
}

// ============================================================================
// Off-Ramp Provider Interface
// ============================================================================

// OffRampProvider represents a third-party off-ramp partner
type OffRampProvider interface {
	// Name returns the provider name
	Name() string

	// GetQuote generates a quote from this provider
	GetQuote(ctx context.Context, request OffRampRequest) (OffRampQuote, error)

	// Execute executes an off-ramp with this provider
	Execute(ctx context.Context, quote OffRampQuote, signedTx []byte) (OffRampResult, error)

	// GetStatus gets operation status from this provider
	GetStatus(ctx context.Context, operationID string) (OffRampResult, error)

	// Cancel cancels an operation with this provider
	Cancel(ctx context.Context, operationID string) error

	// IsHealthy checks if the provider is operational
	IsHealthy(ctx context.Context) bool

	// SupportsCurrency checks if the provider supports a currency
	SupportsCurrency(currency FiatCurrency) bool

	// SupportsMethod checks if the provider supports a payment method
	SupportsMethod(method PaymentMethod) bool
}

// ============================================================================
// Service Interface
// ============================================================================

// Service is the main DEX service interface
type Service interface {
	// Adapter management
	RegisterAdapter(adapter Adapter) error
	UnregisterAdapter(name string) error
	GetAdapter(name string) (Adapter, error)
	ListAdapters() []string

	// Price feed
	GetPrice(ctx context.Context, baseSymbol, quoteSymbol string) (Price, error)
	GetPriceAggregate(ctx context.Context, baseSymbol, quoteSymbol string) (PriceAggregate, error)
	GetTWAP(ctx context.Context, baseSymbol, quoteSymbol string, window time.Duration) (sdkmath.LegacyDec, error)

	// Liquidity pools
	GetPool(ctx context.Context, dex, poolID string) (LiquidityPool, error)
	ListPools(ctx context.Context, query PoolQuery) ([]LiquidityPool, error)

	// Swaps
	GetSwapQuote(ctx context.Context, request SwapRequest) (SwapQuote, error)
	ExecuteSwap(ctx context.Context, quote SwapQuote, signedTx []byte) (SwapResult, error)
	FindBestRoute(ctx context.Context, request SwapRequest) (SwapRoute, error)

	// Off-ramp
	GetOffRampQuote(ctx context.Context, request OffRampRequest) (OffRampQuote, error)
	InitiateOffRamp(ctx context.Context, quote OffRampQuote, signedTx []byte) (OffRampResult, error)
	GetOffRampStatus(ctx context.Context, offRampID string) (OffRampResult, error)
	CancelOffRamp(ctx context.Context, offRampID string) error

	// Health
	IsHealthy(ctx context.Context) bool

	// Lifecycle
	Start(ctx context.Context) error
	Stop() error
}

