package offramp

import (
	"context"
	"time"

	sdkmath "cosmossdk.io/math"
)

// Status represents the state of an off-ramp payout.
type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
	StatusCancelled  Status = "cancelled"
)

// QuoteRequest requests a fiat payout quote.
type QuoteRequest struct {
	CryptoSymbol  string      `json:"crypto_symbol"`
	CryptoDenom   string      `json:"crypto_denom"`
	CryptoAmount  sdkmath.Int `json:"crypto_amount"`
	FiatCurrency  string      `json:"fiat_currency"`
	PaymentMethod string      `json:"payment_method"`
	Sender        string      `json:"sender"`
	Destination   string      `json:"destination"`
}

// Quote represents an off-ramp quote.
type Quote struct {
	ID           string            `json:"id"`
	Request      QuoteRequest      `json:"request"`
	FiatAmount   sdkmath.LegacyDec `json:"fiat_amount"`
	ExchangeRate sdkmath.LegacyDec `json:"exchange_rate"`
	Fee          sdkmath.Int       `json:"fee"`
	Provider     string            `json:"provider"`
	ExpiresAt    time.Time         `json:"expires_at"`
	CreatedAt    time.Time         `json:"created_at"`
}

// IsExpired returns true when the quote is no longer valid.
func (q Quote) IsExpired(now time.Time) bool {
	return !q.ExpiresAt.IsZero() && now.After(q.ExpiresAt)
}

// PayoutRequest executes a fiat payout using an accepted quote.
type PayoutRequest struct {
	Quote       Quote             `json:"quote"`
	CryptoTxRef string            `json:"crypto_tx_ref"`
	Destination string            `json:"destination"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// PayoutResult is the result of an off-ramp payout.
type PayoutResult struct {
	ID            string            `json:"id"`
	QuoteID       string            `json:"quote_id"`
	Status        Status            `json:"status"`
	Provider      string            `json:"provider"`
	FiatAmount    sdkmath.LegacyDec `json:"fiat_amount"`
	CryptoAmount  sdkmath.Int       `json:"crypto_amount"`
	Fee           sdkmath.Int       `json:"fee"`
	Reference     string            `json:"reference"`
	InitiatedAt   time.Time         `json:"initiated_at"`
	CompletedAt   *time.Time        `json:"completed_at,omitempty"`
	FailureReason string            `json:"failure_reason,omitempty"`
}

// Adapter defines the off-ramp provider interface.
type Adapter interface {
	Name() string
	GetQuote(ctx context.Context, req QuoteRequest) (Quote, error)
	InitiatePayout(ctx context.Context, req PayoutRequest) (PayoutResult, error)
	GetStatus(ctx context.Context, payoutID string) (PayoutResult, error)
	Cancel(ctx context.Context, payoutID string) error
	SupportsCurrency(currency string) bool
	SupportsMethod(method string) bool
	IsHealthy(ctx context.Context) bool
}

// Bridge aggregates multiple adapters.
type Bridge interface {
	RegisterAdapter(adapter Adapter) error
	GetQuote(ctx context.Context, req QuoteRequest) (Quote, error)
	InitiatePayout(ctx context.Context, quote Quote, cryptoTxRef string, destination string, metadata map[string]string) (PayoutResult, error)
	GetStatus(ctx context.Context, payoutID string) (PayoutResult, error)
	Cancel(ctx context.Context, payoutID string) error
	ListProviders() []string
}
