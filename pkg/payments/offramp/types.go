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

// IsTerminal returns true when the payout no longer changes.
func (s Status) IsTerminal() bool {
	switch s {
	case StatusCompleted, StatusFailed, StatusCancelled:
		return true
	default:
		return false
	}
}

// QuoteRequest requests a fiat payout quote.
type QuoteRequest struct {
	CryptoSymbol         string             `json:"crypto_symbol"`
	CryptoDenom          string             `json:"crypto_denom"`
	CryptoDecimals       uint8              `json:"crypto_decimals"`
	CryptoAmount         sdkmath.Int        `json:"crypto_amount"`
	FiatCurrency         string             `json:"fiat_currency"`
	PaymentMethod        string             `json:"payment_method"`
	Sender               string             `json:"sender"`
	Destination          string             `json:"destination"`
	Jurisdiction         string             `json:"jurisdiction,omitempty"`
	BeneficiaryReference string             `json:"beneficiary_reference,omitempty"`
	CorrelationID        string             `json:"correlation_id,omitempty"`
	Compliance           ComplianceDecision `json:"compliance,omitempty"`
}

// ComplianceDecision binds a payout to independently recorded KYC and
// sanctions decisions. It contains references only, never identity evidence.
type ComplianceDecision struct {
	Reference         string    `json:"reference,omitempty"`
	KYCDecision       string    `json:"kyc_decision,omitempty"`
	SanctionsDecision string    `json:"sanctions_decision,omitempty"`
	ValidUntil        time.Time `json:"valid_until,omitempty"`
	Revoked           bool      `json:"revoked,omitempty"`
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
	AuditFields  map[string]string `json:"audit_fields,omitempty"`
}

// IsExpired returns true when the quote is no longer valid.
func (q Quote) IsExpired(now time.Time) bool {
	return !q.ExpiresAt.IsZero() && !now.Before(q.ExpiresAt)
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
	ID              string            `json:"id"`
	QuoteID         string            `json:"quote_id"`
	Status          Status            `json:"status"`
	Provider        string            `json:"provider"`
	FiatAmount      sdkmath.LegacyDec `json:"fiat_amount"`
	CryptoAmount    sdkmath.Int       `json:"crypto_amount"`
	Fee             sdkmath.Int       `json:"fee"`
	Reference       string            `json:"reference"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	InitiatedAt     time.Time         `json:"initiated_at"`
	CompletedAt     *time.Time        `json:"completed_at,omitempty"`
	StatusUpdatedAt time.Time         `json:"status_updated_at"`
	FailureReason   string            `json:"failure_reason,omitempty"`
	FailureCode     string            `json:"failure_code,omitempty"`
	Retryable       bool              `json:"retryable,omitempty"`
	AuditFields     map[string]string `json:"audit_fields,omitempty"`
	// DailyReservationKey and DailyReservationOperationID identify the exact
	// durable corridor reservation created before provider initiation. They
	// contain no beneficiary data and permit restart-safe terminal release.
	DailyReservationKey         string `json:"daily_reservation_key,omitempty"`
	DailyReservationOperationID string `json:"daily_reservation_operation_id,omitempty"`
}

// PayoutInitiationState is the durable write-ahead state for one exact payout
// POST. Prepared is intentionally treated as ambiguous after a restart because
// the process may have stopped after the partner accepted the request.
type PayoutInitiationState string

const (
	PayoutInitiationPrepared          PayoutInitiationState = "prepared"
	PayoutInitiationAmbiguous         PayoutInitiationState = "ambiguous"
	PayoutInitiationAccepted          PayoutInitiationState = "accepted"
	PayoutInitiationNoPayout          PayoutInitiationState = "no_payout"
	PayoutInitiationTerminalFailed    PayoutInitiationState = "terminal_failed"
	PayoutInitiationTerminalCancelled PayoutInitiationState = "terminal_cancelled"
)

// PayoutInitiationRecord contains only privacy-safe immutable bindings needed
// to reconcile an uncertain POST and retain its exact daily-limit reservation.
type PayoutInitiationRecord struct {
	Provider                    string                `json:"provider"`
	QuoteID                     string                `json:"quote_id"`
	OperationBinding            string                `json:"operation_binding"`
	RequestBinding              string                `json:"request_binding"`
	FiatAmount                  string                `json:"fiat_amount"`
	CryptoAmount                string                `json:"crypto_amount"`
	Fee                         string                `json:"fee"`
	Metadata                    map[string]string     `json:"metadata"`
	DailyReservationKey         string                `json:"daily_reservation_key"`
	DailyReservationOperationID string                `json:"daily_reservation_operation_id"`
	State                       PayoutInitiationState `json:"state"`
	PayoutID                    string                `json:"payout_id,omitempty"`
	PreparedAt                  time.Time             `json:"prepared_at"`
}

// IsTerminal returns true when the payout is in a final state.
func (r PayoutResult) IsTerminal() bool {
	return r.Status.IsTerminal()
}

// MetadataLookupAdapter optionally supports idempotent payout lookup by metadata.
type MetadataLookupAdapter interface {
	FindPayoutByMetadata(ctx context.Context, metadata map[string]string) (PayoutResult, error)
}

// PayoutBindingRecoveryAdapter restores volatile adapter state from one
// durably known nonterminal payout. Implementations must recover by immutable
// metadata/correlation and reject no-match, ambiguity, or binding mismatch.
type PayoutBindingRecoveryAdapter interface {
	RestorePayoutBinding(ctx context.Context, expected PayoutResult) (PayoutResult, error)
}

// PayoutRepository persists payout state and idempotency lookup keys.
// Production bridges require a durable implementation.
type PayoutRepository interface {
	GetPayout(ctx context.Context, payoutID string) (PayoutResult, error)
	FindPayout(ctx context.Context, provider string, metadata map[string]string) (PayoutResult, error)
	PutPayout(ctx context.Context, result PayoutResult) error
	Durable() bool
}

// PayoutInitiationRepository durably stores the write-ahead binding for an
// exact provider POST. Production adapters require a durable implementation.
type PayoutInitiationRepository interface {
	GetPayoutInitiation(ctx context.Context, provider string, metadata map[string]string) (PayoutInitiationRecord, error)
	PutPayoutInitiation(ctx context.Context, record PayoutInitiationRecord) error
	Durable() bool
}

// ReplacementQuoteBridge optionally exposes an explicit alternate quote. A
// caller must commit the returned quote before initiating it; InitiatePayout
// itself never substitutes a different provider or quote.
type ReplacementQuoteBridge interface {
	GetReplacementQuote(ctx context.Context, previous Quote) (Quote, error)
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
	FindPayoutByMetadata(ctx context.Context, provider string, metadata map[string]string) (PayoutResult, error)
	Cancel(ctx context.Context, payoutID string) error
	ListProviders() []string
}
