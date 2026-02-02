// Package offramp provides fiat off-ramp integration for token-to-fiat payouts.
//
// VE-5E: Fiat off-ramp integration for PayPal/ACH payouts
package offramp

import (
	"errors"
	"time"

	"github.com/virtengine/virtengine/pkg/payment"
)

// ============================================================================
// Errors
// ============================================================================

var (
	// ErrProviderNotConfigured is returned when the off-ramp provider is not configured
	ErrProviderNotConfigured = errors.New("off-ramp provider not configured")

	// ErrKYCNotVerified is returned when user has not completed KYC
	ErrKYCNotVerified = errors.New("KYC verification required for payouts")

	// ErrAMLCheckFailed is returned when AML screening fails
	ErrAMLCheckFailed = errors.New("AML screening failed")

	// ErrPayoutNotFound is returned when payout intent doesn't exist
	ErrPayoutNotFound = errors.New("payout not found")

	// ErrPayoutAlreadyProcessed is returned when payout was already processed
	ErrPayoutAlreadyProcessed = errors.New("payout already processed")

	// ErrPayoutAmountBelowMinimum is returned when amount is too small
	ErrPayoutAmountBelowMinimum = errors.New("payout amount below minimum")

	// ErrPayoutAmountAboveMaximum is returned when amount exceeds limits
	ErrPayoutAmountAboveMaximum = errors.New("payout amount above maximum")

	// ErrPayoutLimitExceeded is returned when daily/monthly limit exceeded
	ErrPayoutLimitExceeded = errors.New("payout limit exceeded")

	// ErrInvalidPayoutDestination is returned for invalid payout destination
	ErrInvalidPayoutDestination = errors.New("invalid payout destination")

	// ErrPayoutFailed is returned when payout fails at provider
	ErrPayoutFailed = errors.New("payout failed")

	// ErrPayoutPending is returned when payout is still pending
	ErrPayoutPending = errors.New("payout is pending")

	// ErrPayoutReversed is returned when payout was reversed
	ErrPayoutReversed = errors.New("payout was reversed")

	// ErrWebhookSignatureInvalid is returned for invalid webhook signatures
	ErrWebhookSignatureInvalid = errors.New("webhook signature verification failed")

	// ErrReconciliationMismatch is returned when reconciliation finds discrepancy
	ErrReconciliationMismatch = errors.New("reconciliation mismatch detected")

	// ErrProviderUnavailable is returned when provider is unreachable
	ErrProviderUnavailable = errors.New("off-ramp provider unavailable")

	// ErrInsufficientBalance is returned when treasury has insufficient balance
	ErrInsufficientBalance = errors.New("insufficient treasury balance for payout")

	// ErrDailyLimitExceeded is returned when daily payout limit exceeded
	ErrDailyLimitExceeded = errors.New("daily payout limit exceeded")

	// ErrMonthlyLimitExceeded is returned when monthly payout limit exceeded
	ErrMonthlyLimitExceeded = errors.New("monthly payout limit exceeded")
)

// ============================================================================
// Provider Types
// ============================================================================

// ProviderType identifies the off-ramp provider
type ProviderType string

const (
	// ProviderPayPal represents PayPal Payouts
	ProviderPayPal ProviderType = "paypal"

	// ProviderACH represents ACH bank transfers
	ProviderACH ProviderType = "ach"

	// ProviderWire represents wire transfers
	ProviderWire ProviderType = "wire"
)

// String returns the string representation
func (p ProviderType) String() string {
	return string(p)
}

// IsValid checks if the provider type is valid
func (p ProviderType) IsValid() bool {
	switch p {
	case ProviderPayPal, ProviderACH, ProviderWire:
		return true
	default:
		return false
	}
}

// ============================================================================
// Payout Status Types
// ============================================================================

// PayoutStatus represents the status of a payout
type PayoutStatus string

const (
	// PayoutStatusPending is the initial status when payout is created
	PayoutStatusPending PayoutStatus = "pending"

	// PayoutStatusKYCRequired indicates KYC verification is needed
	PayoutStatusKYCRequired PayoutStatus = "kyc_required"

	// PayoutStatusAMLPending indicates AML screening is in progress
	PayoutStatusAMLPending PayoutStatus = "aml_pending"

	// PayoutStatusApproved indicates payout passed all checks
	PayoutStatusApproved PayoutStatus = "approved"

	// PayoutStatusProcessing indicates payout is being processed by provider
	PayoutStatusProcessing PayoutStatus = "processing"

	// PayoutStatusSucceeded indicates payout completed successfully
	PayoutStatusSucceeded PayoutStatus = "succeeded"

	// PayoutStatusFailed indicates payout failed
	PayoutStatusFailed PayoutStatus = "failed"

	// PayoutStatusCanceled indicates payout was canceled
	PayoutStatusCanceled PayoutStatus = "canceled"

	// PayoutStatusReversed indicates payout was reversed
	PayoutStatusReversed PayoutStatus = "reversed"

	// PayoutStatusOnHold indicates payout is on hold for review
	PayoutStatusOnHold PayoutStatus = "on_hold"
)

// String returns the string representation
func (s PayoutStatus) String() string {
	return string(s)
}

// IsTerminal returns true if status is a terminal state
func (s PayoutStatus) IsTerminal() bool {
	switch s {
	case PayoutStatusSucceeded, PayoutStatusFailed, PayoutStatusCanceled, PayoutStatusReversed:
		return true
	default:
		return false
	}
}

// IsSuccessful returns true if payout succeeded
func (s PayoutStatus) IsSuccessful() bool {
	return s == PayoutStatusSucceeded
}

// CanCancel returns true if payout can be canceled
func (s PayoutStatus) CanCancel() bool {
	switch s {
	case PayoutStatusPending, PayoutStatusKYCRequired, PayoutStatusAMLPending, PayoutStatusApproved, PayoutStatusOnHold:
		return true
	default:
		return false
	}
}

// ============================================================================
// Destination Types
// ============================================================================

// DestinationType identifies the payout destination type
type DestinationType string

const (
	// DestinationPayPalEmail is PayPal email payout
	DestinationPayPalEmail DestinationType = "paypal_email"

	// DestinationPayPalID is PayPal ID payout
	DestinationPayPalID DestinationType = "paypal_id"

	// DestinationBankAccount is ACH bank account
	DestinationBankAccount DestinationType = "bank_account"

	// DestinationDebitCard is debit card push
	DestinationDebitCard DestinationType = "debit_card"
)

// PayoutDestination represents where the payout will be sent
type PayoutDestination struct {
	// Type is the destination type
	Type DestinationType `json:"type"`

	// Email is the PayPal email (for PayPal payouts)
	Email string `json:"email,omitempty"`

	// PayPalID is the PayPal account ID (for PayPal payouts)
	PayPalID string `json:"paypal_id,omitempty"`

	// BankAccount contains bank account details (for ACH)
	BankAccount *BankAccountDetails `json:"bank_account,omitempty"`

	// DebitCard contains debit card details (for card push)
	DebitCard *DebitCardDetails `json:"debit_card,omitempty"`
}

// BankAccountDetails contains bank account information
type BankAccountDetails struct {
	// AccountHolderName is the name on the account
	AccountHolderName string `json:"account_holder_name"`

	// AccountHolderType is "individual" or "company"
	AccountHolderType string `json:"account_holder_type"`

	// RoutingNumber is the ABA routing number
	RoutingNumber string `json:"routing_number"`

	// AccountNumber is the bank account number
	AccountNumber string `json:"account_number"`

	// AccountType is "checking" or "savings"
	AccountType string `json:"account_type"`

	// BankName is the name of the bank
	BankName string `json:"bank_name,omitempty"`

	// Country is the country code (e.g., "US")
	Country string `json:"country"`
}

// DebitCardDetails contains debit card information
type DebitCardDetails struct {
	// CardToken is the tokenized card reference
	CardToken string `json:"card_token"`

	// Last4 is the last 4 digits of the card
	Last4 string `json:"last4"`

	// Brand is the card brand (visa, mastercard, etc.)
	Brand string `json:"brand"`
}

// ============================================================================
// Payout Intent
// ============================================================================

// PayoutIntent represents a request to send fiat to a user
type PayoutIntent struct {
	// ID is the unique payout identifier
	ID string `json:"id"`

	// Provider is the off-ramp provider
	Provider ProviderType `json:"provider"`

	// Status is the current payout status
	Status PayoutStatus `json:"status"`

	// AccountAddress is the blockchain account receiving crypto
	AccountAddress string `json:"account_address"`

	// VEIDID is the verified identity ID
	VEIDID string `json:"veid_id"`

	// CryptoAmount is the amount of crypto being converted
	CryptoAmount int64 `json:"crypto_amount"`

	// CryptoDenom is the crypto denomination
	CryptoDenom string `json:"crypto_denom"`

	// FiatAmount is the fiat amount to be paid out
	FiatAmount payment.Amount `json:"fiat_amount"`

	// ConversionRate is the rate at quote time
	ConversionRate string `json:"conversion_rate"`

	// Fee is the fee amount
	Fee payment.Amount `json:"fee"`

	// Destination is where to send the payout
	Destination PayoutDestination `json:"destination"`

	// ProviderPayoutID is the ID from the provider
	ProviderPayoutID string `json:"provider_payout_id,omitempty"`

	// ProviderBatchID is the batch ID from the provider (PayPal)
	ProviderBatchID string `json:"provider_batch_id,omitempty"`

	// SettlementID is the on-chain settlement ID
	SettlementID string `json:"settlement_id,omitempty"`

	// Description is a description of the payout
	Description string `json:"description,omitempty"`

	// Metadata contains custom metadata
	Metadata map[string]string `json:"metadata,omitempty"`

	// KYCStatus is the KYC verification status
	KYCStatus KYCStatus `json:"kyc_status"`

	// AMLStatus is the AML screening status
	AMLStatus AMLStatus `json:"aml_status"`

	// FailureCode is the failure code if failed
	FailureCode string `json:"failure_code,omitempty"`

	// FailureMessage is the failure message if failed
	FailureMessage string `json:"failure_message,omitempty"`

	// IdempotencyKey prevents duplicate payouts
	IdempotencyKey string `json:"idempotency_key"`

	// CreatedAt is when the payout was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the payout was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// CompletedAt is when the payout completed
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// ExpiresAt is when the payout quote expires
	ExpiresAt time.Time `json:"expires_at"`

	// AuditTrail contains audit entries
	AuditTrail []PayoutAuditEntry `json:"audit_trail,omitempty"`
}

// IsExpired returns true if the payout quote has expired
func (p *PayoutIntent) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

// AddAuditEntry adds an audit entry to the payout
func (p *PayoutIntent) AddAuditEntry(action, actor, details string) {
	entry := PayoutAuditEntry{
		Timestamp: time.Now(),
		Action:    action,
		Actor:     actor,
		Details:   details,
		Status:    p.Status,
	}
	p.AuditTrail = append(p.AuditTrail, entry)
}

// PayoutAuditEntry represents an audit log entry for a payout
type PayoutAuditEntry struct {
	// Timestamp is when the action occurred
	Timestamp time.Time `json:"timestamp"`

	// Action is the action taken
	Action string `json:"action"`

	// Actor is who took the action
	Actor string `json:"actor"`

	// Details contains additional information
	Details string `json:"details,omitempty"`

	// Status is the status after the action
	Status PayoutStatus `json:"status"`
}

// ============================================================================
// KYC/AML Types
// ============================================================================

// KYCStatus represents the KYC verification status
type KYCStatus string

const (
	// KYCStatusPending indicates KYC not yet started
	KYCStatusPending KYCStatus = "pending"

	// KYCStatusInProgress indicates KYC verification in progress
	KYCStatusInProgress KYCStatus = "in_progress"

	// KYCStatusVerified indicates KYC passed
	KYCStatusVerified KYCStatus = "verified"

	// KYCStatusFailed indicates KYC failed
	KYCStatusFailed KYCStatus = "failed"

	// KYCStatusExpired indicates KYC verification expired
	KYCStatusExpired KYCStatus = "expired"
)

// IsVerified returns true if KYC is verified
func (k KYCStatus) IsVerified() bool {
	return k == KYCStatusVerified
}

// AMLStatus represents the AML screening status
type AMLStatus string

const (
	// AMLStatusPending indicates AML not yet screened
	AMLStatusPending AMLStatus = "pending"

	// AMLStatusScreening indicates AML screening in progress
	AMLStatusScreening AMLStatus = "screening"

	// AMLStatusCleared indicates AML screening passed
	AMLStatusCleared AMLStatus = "cleared"

	// AMLStatusFlagged indicates AML screening flagged for review
	AMLStatusFlagged AMLStatus = "flagged"

	// AMLStatusRejected indicates AML screening rejected
	AMLStatusRejected AMLStatus = "rejected"
)

// IsCleared returns true if AML is cleared
func (a AMLStatus) IsCleared() bool {
	return a == AMLStatusCleared
}

// KYCVerificationLevel represents the required verification level
type KYCVerificationLevel int

const (
	// KYCLevelBasic requires basic identity verification
	KYCLevelBasic KYCVerificationLevel = 1

	// KYCLevelEnhanced requires enhanced verification
	KYCLevelEnhanced KYCVerificationLevel = 2

	// KYCLevelFull requires full verification with document and biometric
	KYCLevelFull KYCVerificationLevel = 3
)

// ============================================================================
// Payout Request
// ============================================================================

// CreatePayoutRequest is a request to create a payout
type CreatePayoutRequest struct {
	// AccountAddress is the user's blockchain address
	AccountAddress string `json:"account_address"`

	// VEIDID is the verified identity ID
	VEIDID string `json:"veid_id"`

	// CryptoAmount is the amount of crypto to convert
	CryptoAmount int64 `json:"crypto_amount"`

	// CryptoDenom is the crypto denomination
	CryptoDenom string `json:"crypto_denom"`

	// Provider is the preferred provider (optional)
	Provider ProviderType `json:"provider,omitempty"`

	// Destination is where to send the payout
	Destination PayoutDestination `json:"destination"`

	// FiatCurrency is the target fiat currency
	FiatCurrency payment.Currency `json:"fiat_currency"`

	// Description is an optional description
	Description string `json:"description,omitempty"`

	// Metadata contains optional metadata
	Metadata map[string]string `json:"metadata,omitempty"`

	// IdempotencyKey prevents duplicate requests
	IdempotencyKey string `json:"idempotency_key"`
}

// ============================================================================
// Payout Response Types
// ============================================================================

// PayoutQuote represents a quote for a payout
type PayoutQuote struct {
	// QuoteID is the unique quote identifier
	QuoteID string `json:"quote_id"`

	// CryptoAmount is the amount of crypto to convert
	CryptoAmount int64 `json:"crypto_amount"`

	// CryptoDenom is the crypto denomination
	CryptoDenom string `json:"crypto_denom"`

	// FiatAmount is the fiat amount after conversion
	FiatAmount payment.Amount `json:"fiat_amount"`

	// ConversionRate is the exchange rate
	ConversionRate string `json:"conversion_rate"`

	// Fee is the fee amount
	Fee payment.Amount `json:"fee"`

	// NetAmount is the net amount after fees
	NetAmount payment.Amount `json:"net_amount"`

	// Provider is the payout provider
	Provider ProviderType `json:"provider"`

	// EstimatedArrival is when funds are expected to arrive
	EstimatedArrival time.Time `json:"estimated_arrival"`

	// ExpiresAt is when the quote expires
	ExpiresAt time.Time `json:"expires_at"`

	// CreatedAt is when the quote was created
	CreatedAt time.Time `json:"created_at"`
}

// IsExpired returns true if the quote has expired
func (q *PayoutQuote) IsExpired() bool {
	return time.Now().After(q.ExpiresAt)
}

// ============================================================================
// Webhook Types
// ============================================================================

// WebhookEventType identifies the type of webhook event
type WebhookEventType string

const (
	// WebhookPayoutCompleted indicates payout completed
	WebhookPayoutCompleted WebhookEventType = "payout.completed"

	// WebhookPayoutFailed indicates payout failed
	WebhookPayoutFailed WebhookEventType = "payout.failed"

	// WebhookPayoutReversed indicates payout reversed
	WebhookPayoutReversed WebhookEventType = "payout.reversed"

	// WebhookPayoutPending indicates payout is pending
	WebhookPayoutPending WebhookEventType = "payout.pending"

	// WebhookPayoutUnclaimed indicates payout unclaimed (PayPal)
	WebhookPayoutUnclaimed WebhookEventType = "payout.unclaimed"

	// WebhookPayoutReturned indicates payout returned (ACH)
	WebhookPayoutReturned WebhookEventType = "payout.returned"
)

// WebhookEvent represents a webhook event from a provider
type WebhookEvent struct {
	// ID is the event ID
	ID string `json:"id"`

	// Type is the event type
	Type WebhookEventType `json:"type"`

	// Provider is the provider that sent the webhook
	Provider ProviderType `json:"provider"`

	// PayoutID is the payout ID
	PayoutID string `json:"payout_id"`

	// ProviderPayoutID is the provider's payout ID
	ProviderPayoutID string `json:"provider_payout_id"`

	// Status is the payout status
	Status PayoutStatus `json:"status"`

	// FailureCode is the failure code if failed
	FailureCode string `json:"failure_code,omitempty"`

	// FailureMessage is the failure message if failed
	FailureMessage string `json:"failure_message,omitempty"`

	// Payload is the raw webhook payload
	Payload []byte `json:"payload"`

	// Timestamp is when the event occurred
	Timestamp time.Time `json:"timestamp"`

	// ReceivedAt is when we received the webhook
	ReceivedAt time.Time `json:"received_at"`
}

// ============================================================================
// Reconciliation Types
// ============================================================================

// ReconciliationStatus represents reconciliation status
type ReconciliationStatus string

const (
	// ReconciliationPending indicates not yet reconciled
	ReconciliationPending ReconciliationStatus = "pending"

	// ReconciliationMatched indicates records match
	ReconciliationMatched ReconciliationStatus = "matched"

	// ReconciliationMismatch indicates records don't match
	ReconciliationMismatch ReconciliationStatus = "mismatch"

	// ReconciliationMissing indicates missing from provider
	ReconciliationMissing ReconciliationStatus = "missing"

	// ReconciliationReviewing indicates under manual review
	ReconciliationReviewing ReconciliationStatus = "reviewing"
)

// ReconciliationRecord represents a reconciliation entry
type ReconciliationRecord struct {
	// ID is the reconciliation record ID
	ID string `json:"id"`

	// PayoutID is the payout ID being reconciled
	PayoutID string `json:"payout_id"`

	// Status is the reconciliation status
	Status ReconciliationStatus `json:"status"`

	// OnChainAmount is the amount recorded on-chain
	OnChainAmount int64 `json:"on_chain_amount"`

	// ProviderAmount is the amount reported by provider
	ProviderAmount int64 `json:"provider_amount"`

	// Discrepancy is the difference if any
	Discrepancy int64 `json:"discrepancy"`

	// Notes contains reconciliation notes
	Notes string `json:"notes,omitempty"`

	// ReconciledAt is when reconciliation completed
	ReconciledAt *time.Time `json:"reconciled_at,omitempty"`

	// ReconciledBy is who performed the reconciliation
	ReconciledBy string `json:"reconciled_by,omitempty"`

	// CreatedAt is when the record was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the record was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// ============================================================================
// Rate Limit Types
// ============================================================================

// PayoutLimits defines payout limits for a user
type PayoutLimits struct {
	// DailyLimit is the daily payout limit in minor units
	DailyLimit int64 `json:"daily_limit"`

	// MonthlyLimit is the monthly payout limit in minor units
	MonthlyLimit int64 `json:"monthly_limit"`

	// PerTransactionLimit is the per-transaction limit
	PerTransactionLimit int64 `json:"per_transaction_limit"`

	// DailyUsed is the amount used today
	DailyUsed int64 `json:"daily_used"`

	// MonthlyUsed is the amount used this month
	MonthlyUsed int64 `json:"monthly_used"`

	// DailyRemaining is the remaining daily limit
	DailyRemaining int64 `json:"daily_remaining"`

	// MonthlyRemaining is the remaining monthly limit
	MonthlyRemaining int64 `json:"monthly_remaining"`

	// LastReset is when limits were last reset
	LastReset time.Time `json:"last_reset"`
}

// CanPayout checks if a payout of the given amount is allowed
func (l *PayoutLimits) CanPayout(amount int64) (bool, error) {
	if amount > l.PerTransactionLimit {
		return false, ErrPayoutAmountAboveMaximum
	}
	if amount > l.DailyRemaining {
		return false, ErrDailyLimitExceeded
	}
	if amount > l.MonthlyRemaining {
		return false, ErrMonthlyLimitExceeded
	}
	return true, nil
}
