// Package dex provides DEX integration for crypto-to-fiat conversions.
//
// VE-905: DEX integration for crypto-to-fiat conversions
package dex

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
	"errors"
	"time"

	sdkmath "cosmossdk.io/math"
)

// ============================================================================
// Errors
// ============================================================================

var (
	// ErrAdapterNotFound is returned when a DEX adapter is not registered
	ErrAdapterNotFound = errors.New("dex adapter not found")

	// ErrPriceFeedStale is returned when price data is too old
	ErrPriceFeedStale = errors.New("price feed data is stale")

	// ErrInsufficientLiquidity is returned when pool has insufficient liquidity
	ErrInsufficientLiquidity = errors.New("insufficient liquidity")

	// ErrSlippageExceeded is returned when slippage exceeds tolerance
	ErrSlippageExceeded = errors.New("slippage tolerance exceeded")

	// ErrSwapFailed is returned when swap execution fails
	ErrSwapFailed = errors.New("swap execution failed")

	// ErrQuoteExpired is returned when a swap quote has expired
	ErrQuoteExpired = errors.New("swap quote expired")

	// ErrUnsupportedPair is returned for unsupported trading pairs
	ErrUnsupportedPair = errors.New("unsupported trading pair")

	// ErrOffRampFailed is returned when off-ramp operation fails
	ErrOffRampFailed = errors.New("off-ramp operation failed")

	// ErrKYCRequired is returned when KYC verification is needed
	ErrKYCRequired = errors.New("KYC verification required for off-ramp")

	// ErrAmountTooSmall is returned when amount is below minimum
	ErrAmountTooSmall = errors.New("amount below minimum threshold")

	// ErrAmountTooLarge is returned when amount exceeds limits
	ErrAmountTooLarge = errors.New("amount exceeds maximum threshold")

	// ErrCircuitBreakerTripped is returned when price deviation triggers safety
	ErrCircuitBreakerTripped = errors.New("circuit breaker tripped: abnormal price movement")

	// ErrProviderUnavailable is returned when a DEX provider is offline
	ErrProviderUnavailable = errors.New("dex provider unavailable")
)

// ============================================================================
// Token and Pair Types
// ============================================================================

// Token represents a cryptocurrency token
type Token struct {
	// Symbol is the token symbol (e.g., "UVE", "USDC")
	Symbol string `json:"symbol"`

	// Denom is the on-chain denomination
	Denom string `json:"denom"`

	// Decimals is the token decimal precision
	Decimals uint8 `json:"decimals"`

	// ChainID is the chain where the token resides
	ChainID string `json:"chain_id"`

	// ContractAddress is the token contract (if applicable)
	ContractAddress string `json:"contract_address,omitempty"`

	// IsNative indicates if this is a native chain token
	IsNative bool `json:"is_native"`
}

// TradingPair represents a trading pair
type TradingPair struct {
	// BaseToken is the base token (being sold)
	BaseToken Token `json:"base_token"`

	// QuoteToken is the quote token (being bought)
	QuoteToken Token `json:"quote_token"`

	// MinOrderSize is the minimum order size
	MinOrderSize sdkmath.Int `json:"min_order_size"`

	// MaxOrderSize is the maximum order size
	MaxOrderSize sdkmath.Int `json:"max_order_size"`
}

// PairID returns a unique identifier for the trading pair
func (p TradingPair) PairID() string {
	return p.BaseToken.Symbol + "/" + p.QuoteToken.Symbol
}

// Reverse returns the inverse trading pair
func (p TradingPair) Reverse() TradingPair {
	return TradingPair{
		BaseToken:    p.QuoteToken,
		QuoteToken:   p.BaseToken,
		MinOrderSize: p.MinOrderSize,
		MaxOrderSize: p.MaxOrderSize,
	}
}

// ============================================================================
// Price Types
// ============================================================================

// Price represents a token price
type Price struct {
	// Pair is the trading pair
	Pair TradingPair `json:"pair"`

	// Rate is the exchange rate (quote/base)
	Rate sdkmath.LegacyDec `json:"rate"`

	// Timestamp is when the price was fetched
	Timestamp time.Time `json:"timestamp"`

	// Source is the price data source
	Source string `json:"source"`

	// Confidence is the price confidence score (0.0-1.0)
	Confidence float64 `json:"confidence"`
}

// IsStale checks if the price data is stale
func (p Price) IsStale(maxAge time.Duration) bool {
	return time.Since(p.Timestamp) > maxAge
}

// PriceAggregate represents aggregated price data from multiple sources
type PriceAggregate struct {
	// Pair is the trading pair
	Pair TradingPair `json:"pair"`

	// MedianPrice is the median price across sources
	MedianPrice sdkmath.LegacyDec `json:"median_price"`

	// TWAP is the time-weighted average price
	TWAP sdkmath.LegacyDec `json:"twap"`

	// VWAP is the volume-weighted average price
	VWAP sdkmath.LegacyDec `json:"vwap"`

	// High24h is the 24-hour high
	High24h sdkmath.LegacyDec `json:"high_24h"`

	// Low24h is the 24-hour low
	Low24h sdkmath.LegacyDec `json:"low_24h"`

	// Volume24h is the 24-hour trading volume
	Volume24h sdkmath.Int `json:"volume_24h"`

	// Sources is the list of contributing price sources
	Sources []Price `json:"sources"`

	// AggregatedAt is when the aggregation was performed
	AggregatedAt time.Time `json:"aggregated_at"`
}

// ============================================================================
// Swap Types
// ============================================================================

// SwapType defines the type of swap operation
type SwapType string

const (
	// SwapTypeExactIn specifies exact input amount
	SwapTypeExactIn SwapType = "exact_in"

	// SwapTypeExactOut specifies exact output amount
	SwapTypeExactOut SwapType = "exact_out"
)

// SwapRequest represents a swap request
type SwapRequest struct {
	// FromToken is the token being sold
	FromToken Token `json:"from_token"`

	// ToToken is the token being bought
	ToToken Token `json:"to_token"`

	// Amount is the swap amount (in/out depends on Type)
	Amount sdkmath.Int `json:"amount"`

	// Type is the swap type (exact in or exact out)
	Type SwapType `json:"type"`

	// SlippageTolerance is the maximum acceptable slippage (0.0-1.0)
	SlippageTolerance float64 `json:"slippage_tolerance"`

	// Deadline is the transaction deadline
	Deadline time.Time `json:"deadline"`

	// Sender is the sender's address
	Sender string `json:"sender"`

	// Recipient is the recipient's address (optional, defaults to sender)
	Recipient string `json:"recipient,omitempty"`

	// PreferredDEX is the preferred DEX adapter (optional)
	PreferredDEX string `json:"preferred_dex,omitempty"`
}

// Validate validates the swap request
func (r SwapRequest) Validate() error {
	if r.FromToken.Symbol == "" || r.ToToken.Symbol == "" {
		return errors.New("from_token and to_token are required")
	}
	if r.FromToken.Symbol == r.ToToken.Symbol {
		return errors.New("from_token and to_token must be different")
	}
	if !r.Amount.IsPositive() {
		return errors.New("amount must be positive")
	}
	if r.SlippageTolerance < 0 || r.SlippageTolerance > 1.0 {
		return errors.New("slippage_tolerance must be between 0 and 1")
	}
	if r.Sender == "" {
		return errors.New("sender is required")
	}
	return nil
}

// SwapRoute represents a swap route through one or more pools
type SwapRoute struct {
	// Hops is the list of swap hops
	Hops []SwapHop `json:"hops"`

	// TotalGas is the estimated total gas cost
	TotalGas uint64 `json:"total_gas"`

	// PriceImpact is the expected price impact
	PriceImpact float64 `json:"price_impact"`
}

// SwapHop represents a single hop in a multi-hop swap
type SwapHop struct {
	// PoolID is the liquidity pool identifier
	PoolID string `json:"pool_id"`

	// DEX is the DEX adapter name
	DEX string `json:"dex"`

	// FromToken is the input token for this hop
	FromToken Token `json:"from_token"`

	// ToToken is the output token for this hop
	ToToken Token `json:"to_token"`

	// AmountIn is the input amount
	AmountIn sdkmath.Int `json:"amount_in"`

	// AmountOut is the expected output amount
	AmountOut sdkmath.Int `json:"amount_out"`

	// Fee is the pool fee for this hop
	Fee sdkmath.LegacyDec `json:"fee"`
}

// SwapQuote represents a swap quote with execution details
type SwapQuote struct {
	// ID is the unique quote identifier
	ID string `json:"id"`

	// Request is the original swap request
	Request SwapRequest `json:"request"`

	// Route is the optimal swap route
	Route SwapRoute `json:"route"`

	// InputAmount is the exact input amount
	InputAmount sdkmath.Int `json:"input_amount"`

	// OutputAmount is the expected output amount
	OutputAmount sdkmath.Int `json:"output_amount"`

	// MinOutputAmount is the minimum output with slippage
	MinOutputAmount sdkmath.Int `json:"min_output_amount"`

	// Rate is the effective exchange rate
	Rate sdkmath.LegacyDec `json:"rate"`

	// PriceImpact is the total price impact
	PriceImpact float64 `json:"price_impact"`

	// TotalFee is the total fee amount
	TotalFee sdkmath.Int `json:"total_fee"`

	// GasEstimate is the estimated gas cost
	GasEstimate uint64 `json:"gas_estimate"`

	// ExpiresAt is when the quote expires
	ExpiresAt time.Time `json:"expires_at"`

	// CreatedAt is when the quote was created
	CreatedAt time.Time `json:"created_at"`
}

// IsExpired checks if the quote has expired
func (q SwapQuote) IsExpired() bool {
	return time.Now().After(q.ExpiresAt)
}

// SwapResult represents the result of a swap execution
type SwapResult struct {
	// QuoteID is the quote that was executed
	QuoteID string `json:"quote_id"`

	// TxHash is the transaction hash
	TxHash string `json:"tx_hash"`

	// InputAmount is the actual input amount
	InputAmount sdkmath.Int `json:"input_amount"`

	// OutputAmount is the actual output amount
	OutputAmount sdkmath.Int `json:"output_amount"`

	// Fee is the actual fee paid
	Fee sdkmath.Int `json:"fee"`

	// GasUsed is the actual gas used
	GasUsed uint64 `json:"gas_used"`

	// ExecutedAt is when the swap was executed
	ExecutedAt time.Time `json:"executed_at"`

	// Route is the route that was used
	Route SwapRoute `json:"route"`
}

// ============================================================================
// Liquidity Pool Types
// ============================================================================

// PoolType defines the type of liquidity pool
type PoolType string

const (
	// PoolTypeConstantProduct is a constant product AMM (x*y=k)
	PoolTypeConstantProduct PoolType = "constant_product"

	// PoolTypeStableSwap is a stableswap curve (Curve-style)
	PoolTypeStableSwap PoolType = "stable_swap"

	// PoolTypeConcentrated is concentrated liquidity (Uniswap v3 style)
	PoolTypeConcentrated PoolType = "concentrated"
)

// LiquidityPool represents a liquidity pool
type LiquidityPool struct {
	// ID is the pool identifier
	ID string `json:"id"`

	// DEX is the DEX adapter name
	DEX string `json:"dex"`

	// Type is the pool type
	Type PoolType `json:"type"`

	// Tokens is the list of tokens in the pool
	Tokens []Token `json:"tokens"`

	// Reserves maps token symbol to reserve amount
	Reserves map[string]sdkmath.Int `json:"reserves"`

	// TotalLiquidity is the total liquidity value (in quote currency)
	TotalLiquidity sdkmath.LegacyDec `json:"total_liquidity"`

	// Fee is the pool swap fee (0.0-1.0)
	Fee sdkmath.LegacyDec `json:"fee"`

	// Volume24h is the 24-hour volume
	Volume24h sdkmath.Int `json:"volume_24h"`

	// APY is the current APY for LPs
	APY sdkmath.LegacyDec `json:"apy"`

	// UpdatedAt is when pool data was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// PoolQuery specifies pool query parameters
type PoolQuery struct {
	// DEX filters by DEX adapter
	DEX string `json:"dex,omitempty"`

	// TokenSymbols filters pools containing these tokens
	TokenSymbols []string `json:"token_symbols,omitempty"`

	// MinLiquidity filters by minimum liquidity
	MinLiquidity sdkmath.LegacyDec `json:"min_liquidity,omitempty"`

	// PoolType filters by pool type
	PoolType PoolType `json:"pool_type,omitempty"`

	// Limit limits the number of results
	Limit int `json:"limit,omitempty"`

	// Offset for pagination
	Offset int `json:"offset,omitempty"`
}

// ============================================================================
// Off-Ramp Types
// ============================================================================

// FiatCurrency represents a fiat currency
type FiatCurrency string

const (
	FiatUSD FiatCurrency = "USD"
	FiatEUR FiatCurrency = "EUR"
	FiatGBP FiatCurrency = "GBP"
	FiatAUD FiatCurrency = "AUD"
	FiatCAD FiatCurrency = "CAD"
	FiatCHF FiatCurrency = "CHF"
	FiatJPY FiatCurrency = "JPY"
)

// OffRampStatus represents the status of an off-ramp operation
type OffRampStatus string

const (
	OffRampStatusPending    OffRampStatus = "pending"
	OffRampStatusProcessing OffRampStatus = "processing"
	OffRampStatusCompleted  OffRampStatus = "completed"
	OffRampStatusFailed     OffRampStatus = "failed"
	OffRampStatusCancelled  OffRampStatus = "cancelled"
)

// PaymentMethod represents a fiat payment method
type PaymentMethod string

const (
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodSEPA         PaymentMethod = "sepa"
	PaymentMethodACH          PaymentMethod = "ach"
	PaymentMethodWire         PaymentMethod = "wire"
	PaymentMethodCard         PaymentMethod = "card"
)

// OffRampRequest represents a crypto-to-fiat off-ramp request
type OffRampRequest struct {
	// CryptoToken is the cryptocurrency to sell
	CryptoToken Token `json:"crypto_token"`

	// CryptoAmount is the amount of crypto to sell
	CryptoAmount sdkmath.Int `json:"crypto_amount"`

	// FiatCurrency is the target fiat currency
	FiatCurrency FiatCurrency `json:"fiat_currency"`

	// PaymentMethod is the fiat payment method
	PaymentMethod PaymentMethod `json:"payment_method"`

	// BankDetails contains bank account details (encrypted)
	BankDetails *BankDetails `json:"bank_details,omitempty"`

	// Sender is the sender's blockchain address
	Sender string `json:"sender"`

	// VEIDScore is the user's VEID identity score (for KYC)
	VEIDScore int64 `json:"veid_score"`
}

// Validate validates the off-ramp request
func (r OffRampRequest) Validate() error {
	if r.CryptoToken.Symbol == "" {
		return errors.New("crypto_token is required")
	}
	if !r.CryptoAmount.IsPositive() {
		return errors.New("crypto_amount must be positive")
	}
	if r.FiatCurrency == "" {
		return errors.New("fiat_currency is required")
	}
	if r.PaymentMethod == "" {
		return errors.New("payment_method is required")
	}
	if r.Sender == "" {
		return errors.New("sender is required")
	}
	return nil
}

// BankDetails contains bank account information (stored encrypted)
type BankDetails struct {
	// AccountName is the bank account holder name
	AccountName string `json:"account_name"`

	// AccountNumber is the bank account number
	AccountNumber string `json:"account_number"`

	// RoutingNumber is the routing/sort code
	RoutingNumber string `json:"routing_number,omitempty"`

	// IBAN is the international bank account number
	IBAN string `json:"iban,omitempty"`

	// SWIFT is the SWIFT/BIC code
	SWIFT string `json:"swift,omitempty"`

	// BankName is the name of the bank
	BankName string `json:"bank_name"`

	// BankCountry is the country of the bank
	BankCountry string `json:"bank_country"`
}

// OffRampQuote represents a quote for an off-ramp operation
type OffRampQuote struct {
	// ID is the unique quote identifier
	ID string `json:"id"`

	// Request is the original off-ramp request
	Request OffRampRequest `json:"request"`

	// CryptoAmount is the crypto amount to be sold
	CryptoAmount sdkmath.Int `json:"crypto_amount"`

	// FiatAmount is the fiat amount to be received
	FiatAmount sdkmath.LegacyDec `json:"fiat_amount"`

	// ExchangeRate is the crypto-to-fiat rate
	ExchangeRate sdkmath.LegacyDec `json:"exchange_rate"`

	// Fee is the total fee in crypto
	Fee sdkmath.Int `json:"fee"`

	// FiatFee is the fiat processing fee
	FiatFee sdkmath.LegacyDec `json:"fiat_fee"`

	// Provider is the off-ramp partner name
	Provider string `json:"provider"`

	// EstimatedSettlement is the expected settlement time
	EstimatedSettlement time.Duration `json:"estimated_settlement"`

	// ExpiresAt is when the quote expires
	ExpiresAt time.Time `json:"expires_at"`

	// CreatedAt is when the quote was created
	CreatedAt time.Time `json:"created_at"`
}

// IsExpired checks if the quote has expired
func (q OffRampQuote) IsExpired() bool {
	return time.Now().After(q.ExpiresAt)
}

// OffRampResult represents the result of an off-ramp operation
type OffRampResult struct {
	// ID is the off-ramp operation ID
	ID string `json:"id"`

	// QuoteID is the quote that was used
	QuoteID string `json:"quote_id"`

	// Status is the current status
	Status OffRampStatus `json:"status"`

	// CryptoTxHash is the blockchain transaction hash
	CryptoTxHash string `json:"crypto_tx_hash"`

	// FiatReference is the fiat payment reference
	FiatReference string `json:"fiat_reference,omitempty"`

	// CryptoAmount is the crypto amount sold
	CryptoAmount sdkmath.Int `json:"crypto_amount"`

	// FiatAmount is the fiat amount received
	FiatAmount sdkmath.LegacyDec `json:"fiat_amount"`

	// Fee is the total fee paid
	Fee sdkmath.Int `json:"fee"`

	// Provider is the off-ramp partner
	Provider string `json:"provider"`

	// InitiatedAt is when the off-ramp was initiated
	InitiatedAt time.Time `json:"initiated_at"`

	// CompletedAt is when the off-ramp completed (if applicable)
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// FailureReason is the reason for failure (if applicable)
	FailureReason string `json:"failure_reason,omitempty"`
}
