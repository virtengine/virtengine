// Package payment provides payment gateway integration for Visa/Mastercard.
//
// VE-906: Payment gateway integration for fiat-to-crypto onramp
package payment

import (
	"errors"
	"time"

	sdkmath "cosmossdk.io/math"
)

// ============================================================================
// Errors
// ============================================================================

var (
	// ErrGatewayNotConfigured is returned when the payment gateway is not configured
	ErrGatewayNotConfigured = errors.New("payment gateway not configured")

	// ErrInvalidCardToken is returned for invalid or expired card tokens
	ErrInvalidCardToken = errors.New("invalid or expired card token")

	// ErrPaymentDeclined is returned when payment is declined by issuer
	ErrPaymentDeclined = errors.New("payment declined by issuer")

	// ErrInsufficientFunds is returned when card has insufficient funds
	ErrInsufficientFunds = errors.New("insufficient funds")

	// ErrSCARequired is returned when 3D Secure authentication is required
	ErrSCARequired = errors.New("3D Secure authentication required")

	// ErrSCAFailed is returned when 3D Secure authentication fails
	ErrSCAFailed = errors.New("3D Secure authentication failed")

	// ErrPaymentIntentNotFound is returned when payment intent doesn't exist
	ErrPaymentIntentNotFound = errors.New("payment intent not found")

	// ErrRefundNotAllowed is returned when refund is not permitted
	ErrRefundNotAllowed = errors.New("refund not allowed for this payment")

	// ErrRefundAmountExceeds is returned when refund amount exceeds captured amount
	ErrRefundAmountExceeds = errors.New("refund amount exceeds captured amount")

	// ErrWebhookSignatureInvalid is returned for invalid webhook signatures
	ErrWebhookSignatureInvalid = errors.New("webhook signature verification failed")

	// ErrWebhookEventUnknown is returned for unknown webhook event types
	ErrWebhookEventUnknown = errors.New("unknown webhook event type")

	// ErrDuplicateIdempotencyKey is returned for duplicate requests
	ErrDuplicateIdempotencyKey = errors.New("duplicate idempotency key")

	// ErrCardExpired is returned when the card has expired
	ErrCardExpired = errors.New("card has expired")

	// ErrInvalidCurrency is returned for unsupported currencies
	ErrInvalidCurrency = errors.New("unsupported currency")

	// ErrAmountBelowMinimum is returned when amount is too small
	ErrAmountBelowMinimum = errors.New("amount below minimum threshold")

	// ErrAmountAboveMaximum is returned when amount exceeds limits
	ErrAmountAboveMaximum = errors.New("amount above maximum threshold")

	// ErrRateLimitExceeded is returned when rate limit is hit
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrGatewayUnavailable is returned when gateway is unreachable
	ErrGatewayUnavailable = errors.New("payment gateway unavailable")

	// ErrDisputeInProgress is returned when payment is under dispute
	ErrDisputeInProgress = errors.New("payment has active dispute")

	// ErrQuoteExpired is returned when a conversion quote has expired
	ErrQuoteExpired = errors.New("conversion quote expired")
)

// ============================================================================
// Gateway Types
// ============================================================================

// GatewayType identifies the payment gateway provider
type GatewayType string

const (
	// GatewayStripe represents Stripe payment gateway
	GatewayStripe GatewayType = "stripe"

	// GatewayAdyen represents Adyen payment gateway
	GatewayAdyen GatewayType = "adyen"
)

// String returns the string representation
func (g GatewayType) String() string {
	return string(g)
}

// IsValid checks if the gateway type is valid
func (g GatewayType) IsValid() bool {
	switch g {
	case GatewayStripe, GatewayAdyen:
		return true
	default:
		return false
	}
}

// ============================================================================
// Currency and Amount Types
// ============================================================================

// Currency represents an ISO 4217 currency code
type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
	CurrencyGBP Currency = "GBP"
	CurrencyJPY Currency = "JPY"
	CurrencyAUD Currency = "AUD"
	CurrencyCAD Currency = "CAD"
	CurrencyCHF Currency = "CHF"
)

// IsValid checks if the currency code is valid
func (c Currency) IsValid() bool {
	switch c {
	case CurrencyUSD, CurrencyEUR, CurrencyGBP, CurrencyJPY,
		CurrencyAUD, CurrencyCAD, CurrencyCHF:
		return true
	default:
		return false
	}
}

// MinorUnitFactor returns the factor for converting to minor units
// (e.g., cents for USD, pence for GBP)
func (c Currency) MinorUnitFactor() int64 {
	switch c {
	case CurrencyJPY:
		return 1 // JPY has no minor units
	default:
		return 100 // Most currencies use 2 decimal places
	}
}

// Amount represents a monetary amount in minor units (e.g., cents)
type Amount struct {
	// Value is the amount in minor units (e.g., 1000 = $10.00 for USD)
	Value int64 `json:"value"`

	// Currency is the ISO 4217 currency code
	Currency Currency `json:"currency"`
}

// NewAmount creates a new Amount in minor units
func NewAmount(value int64, currency Currency) Amount {
	return Amount{
		Value:    value,
		Currency: currency,
	}
}

// NewAmountFromMajor creates an Amount from major units (e.g., dollars)
func NewAmountFromMajor(major float64, currency Currency) Amount {
	return Amount{
		Value:    int64(major * float64(currency.MinorUnitFactor())),
		Currency: currency,
	}
}

// Major returns the amount in major units (e.g., dollars)
func (a Amount) Major() float64 {
	return float64(a.Value) / float64(a.Currency.MinorUnitFactor())
}

// IsZero checks if the amount is zero
func (a Amount) IsZero() bool {
	return a.Value == 0
}

// IsPositive checks if the amount is positive
func (a Amount) IsPositive() bool {
	return a.Value > 0
}

// Add adds two amounts (must be same currency)
func (a Amount) Add(other Amount) (Amount, error) {
	if a.Currency != other.Currency {
		return Amount{}, ErrInvalidCurrency
	}
	return Amount{Value: a.Value + other.Value, Currency: a.Currency}, nil
}

// Sub subtracts amount (must be same currency)
func (a Amount) Sub(other Amount) (Amount, error) {
	if a.Currency != other.Currency {
		return Amount{}, ErrInvalidCurrency
	}
	return Amount{Value: a.Value - other.Value, Currency: a.Currency}, nil
}

// ============================================================================
// Card Types (Tokens Only - No Raw PAN Data)
// ============================================================================

// CardBrand represents card network brands
type CardBrand string

const (
	CardBrandVisa       CardBrand = "visa"
	CardBrandMastercard CardBrand = "mastercard"
	CardBrandAmex       CardBrand = "amex"
	CardBrandDiscover   CardBrand = "discover"
	CardBrandUnknown    CardBrand = "unknown"
)

// IsSupported checks if the card brand is supported
func (b CardBrand) IsSupported() bool {
	switch b {
	case CardBrandVisa, CardBrandMastercard:
		return true
	default:
		return false
	}
}

// CardToken represents a tokenized card (from gateway's client-side SDK)
// NEVER contains actual card numbers - only gateway-provided tokens
type CardToken struct {
	// Token is the gateway-provided payment method token
	Token string `json:"token"`

	// Gateway is which gateway issued this token
	Gateway GatewayType `json:"gateway"`

	// Last4 is the last 4 digits of the card (for display only)
	Last4 string `json:"last4"`

	// Brand is the card brand
	Brand CardBrand `json:"brand"`

	// ExpiryMonth is the expiry month (1-12)
	ExpiryMonth int `json:"expiry_month"`

	// ExpiryYear is the 4-digit expiry year
	ExpiryYear int `json:"expiry_year"`

	// Fingerprint is a unique card identifier (for detecting duplicates)
	Fingerprint string `json:"fingerprint"`

	// Country is the card issuing country (ISO 3166-1 alpha-2)
	Country string `json:"country"`

	// CreatedAt is when the token was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when the token expires (optional)
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// IsExpired checks if the card (not token) is expired
func (c CardToken) IsExpired() bool {
	now := time.Now()
	// Card expires at end of expiry month
	expiryDate := time.Date(c.ExpiryYear, time.Month(c.ExpiryMonth)+1, 0, 23, 59, 59, 0, time.UTC)
	return now.After(expiryDate)
}

// IsTokenExpired checks if the token itself is expired
func (c CardToken) IsTokenExpired() bool {
	if c.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*c.ExpiresAt)
}

// MaskedNumber returns a masked card number for display (e.g., "•••• •••• •••• 4242")
func (c CardToken) MaskedNumber() string {
	return "•••• •••• •••• " + c.Last4
}

// ============================================================================
// Customer Types
// ============================================================================

// Customer represents a customer in the payment gateway
type Customer struct {
	// ID is the gateway-assigned customer ID
	ID string `json:"id"`

	// Email is the customer's email
	Email string `json:"email"`

	// Name is the customer's full name
	Name string `json:"name,omitempty"`

	// Phone is the customer's phone number
	Phone string `json:"phone,omitempty"`

	// VEIDAddress is the VirtEngine blockchain address
	VEIDAddress string `json:"veid_address"`

	// DefaultPaymentMethodID is the default payment method
	DefaultPaymentMethodID string `json:"default_payment_method_id,omitempty"`

	// Metadata is additional key-value metadata
	Metadata map[string]string `json:"metadata,omitempty"`

	// CreatedAt is when the customer was created
	CreatedAt time.Time `json:"created_at"`
}

// BillingAddress represents a billing address
type BillingAddress struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"` // ISO 3166-1 alpha-2
}

// ============================================================================
// Payment Intent Types
// ============================================================================

// PaymentIntentStatus represents the status of a payment intent
type PaymentIntentStatus string

const (
	// PaymentIntentStatusRequiresPaymentMethod needs payment method
	PaymentIntentStatusRequiresPaymentMethod PaymentIntentStatus = "requires_payment_method"

	// PaymentIntentStatusRequiresConfirmation ready for confirmation
	PaymentIntentStatusRequiresConfirmation PaymentIntentStatus = "requires_confirmation"

	// PaymentIntentStatusRequiresAction needs customer action (3DS)
	PaymentIntentStatusRequiresAction PaymentIntentStatus = "requires_action"

	// PaymentIntentStatusProcessing being processed
	PaymentIntentStatusProcessing PaymentIntentStatus = "processing"

	// PaymentIntentStatusSucceeded payment succeeded
	PaymentIntentStatusSucceeded PaymentIntentStatus = "succeeded"

	// PaymentIntentStatusCanceled payment was canceled
	PaymentIntentStatusCanceled PaymentIntentStatus = "canceled"

	// PaymentIntentStatusFailed payment failed
	PaymentIntentStatusFailed PaymentIntentStatus = "failed"
)

// IsFinal checks if the status is a terminal state
func (s PaymentIntentStatus) IsFinal() bool {
	switch s {
	case PaymentIntentStatusSucceeded, PaymentIntentStatusCanceled, PaymentIntentStatusFailed:
		return true
	default:
		return false
	}
}

// IsSuccessful checks if payment was successful
func (s PaymentIntentStatus) IsSuccessful() bool {
	return s == PaymentIntentStatusSucceeded
}

// PaymentIntent represents a payment intent
type PaymentIntent struct {
	// ID is the unique payment intent ID
	ID string `json:"id"`

	// Gateway is which gateway processed this
	Gateway GatewayType `json:"gateway"`

	// Amount is the payment amount
	Amount Amount `json:"amount"`

	// Status is the current status
	Status PaymentIntentStatus `json:"status"`

	// CustomerID is the customer making the payment
	CustomerID string `json:"customer_id"`

	// PaymentMethodID is the payment method used
	PaymentMethodID string `json:"payment_method_id,omitempty"`

	// Description is a description of the payment
	Description string `json:"description,omitempty"`

	// Metadata is additional key-value metadata
	Metadata map[string]string `json:"metadata,omitempty"`

	// ClientSecret is for client-side confirmation
	ClientSecret string `json:"client_secret,omitempty"`

	// RequiresSCA indicates if 3DS authentication is needed
	RequiresSCA bool `json:"requires_sca"`

	// SCARedirectURL is the URL for 3DS authentication
	SCARedirectURL string `json:"sca_redirect_url,omitempty"`

	// SCAStatus is the 3DS authentication status
	SCAStatus SCAStatus `json:"sca_status,omitempty"`

	// ReceiptEmail is where to send receipt
	ReceiptEmail string `json:"receipt_email,omitempty"`

	// StatementDescriptor appears on customer statement
	StatementDescriptor string `json:"statement_descriptor,omitempty"`

	// CapturedAmount is the amount captured (for auth-capture flows)
	CapturedAmount Amount `json:"captured_amount"`

	// RefundedAmount is the amount refunded
	RefundedAmount Amount `json:"refunded_amount"`

	// FailureCode is the error code if failed
	FailureCode string `json:"failure_code,omitempty"`

	// FailureMessage is the error message if failed
	FailureMessage string `json:"failure_message,omitempty"`

	// CreatedAt is when the intent was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the intent was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// CanRefund checks if this payment can be refunded
func (p PaymentIntent) CanRefund() bool {
	if p.Status != PaymentIntentStatusSucceeded {
		return false
	}
	remaining := p.CapturedAmount.Value - p.RefundedAmount.Value
	return remaining > 0
}

// RefundableAmount returns the maximum refundable amount
func (p PaymentIntent) RefundableAmount() Amount {
	remaining := p.CapturedAmount.Value - p.RefundedAmount.Value
	if remaining < 0 {
		remaining = 0
	}
	return Amount{Value: remaining, Currency: p.Amount.Currency}
}

// PaymentIntentRequest is a request to create a payment intent
type PaymentIntentRequest struct {
	// Amount is the payment amount
	Amount Amount `json:"amount"`

	// CustomerID is the customer ID
	CustomerID string `json:"customer_id"`

	// PaymentMethodID is the payment method to use (optional)
	PaymentMethodID string `json:"payment_method_id,omitempty"`

	// Description is a description of the payment
	Description string `json:"description,omitempty"`

	// Metadata is additional metadata
	Metadata map[string]string `json:"metadata,omitempty"`

	// ReceiptEmail where to send receipt
	ReceiptEmail string `json:"receipt_email,omitempty"`

	// StatementDescriptor for customer bank statement
	StatementDescriptor string `json:"statement_descriptor,omitempty"`

	// CaptureMethod: "automatic" or "manual"
	CaptureMethod string `json:"capture_method,omitempty"`

	// ReturnURL for 3DS redirect
	ReturnURL string `json:"return_url,omitempty"`

	// IdempotencyKey to prevent duplicate charges
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

// ============================================================================
// 3D Secure / SCA Types
// ============================================================================

// SCAStatus represents 3D Secure authentication status
type SCAStatus string

const (
	// SCAStatusNotRequired authentication not required
	SCAStatusNotRequired SCAStatus = "not_required"

	// SCAStatusPending waiting for authentication
	SCAStatusPending SCAStatus = "pending"

	// SCAStatusSucceeded authentication succeeded
	SCAStatusSucceeded SCAStatus = "succeeded"

	// SCAStatusFailed authentication failed
	SCAStatusFailed SCAStatus = "failed"
)

// SCAResult contains 3D Secure authentication details
type SCAResult struct {
	// Status is the authentication status
	Status SCAStatus `json:"status"`

	// Version is the 3DS version used (e.g., "2.2.0")
	Version string `json:"version"`

	// ECI is the Electronic Commerce Indicator
	ECI string `json:"eci,omitempty"`

	// CAVV is the Cardholder Authentication Verification Value
	CAVV string `json:"cavv,omitempty"`

	// TransactionID is the 3DS transaction ID
	TransactionID string `json:"transaction_id,omitempty"`

	// Liability indicates who bears fraud liability
	Liability string `json:"liability,omitempty"` // "issuer" or "merchant"
}

// ============================================================================
// Refund Types
// ============================================================================

// RefundStatus represents refund status
type RefundStatus string

const (
	RefundStatusPending   RefundStatus = "pending"
	RefundStatusSucceeded RefundStatus = "succeeded"
	RefundStatusFailed    RefundStatus = "failed"
	RefundStatusCanceled  RefundStatus = "canceled"
)

// RefundReason represents the reason for refund
type RefundReason string

const (
	RefundReasonDuplicate           RefundReason = "duplicate"
	RefundReasonFraudulent          RefundReason = "fraudulent"
	RefundReasonRequestedByCustomer RefundReason = "requested_by_customer"
	RefundReasonServiceNotProvided  RefundReason = "service_not_provided"
	RefundReasonOther               RefundReason = "other"
)

// Refund represents a refund
type Refund struct {
	// ID is the refund ID
	ID string `json:"id"`

	// PaymentIntentID is the original payment
	PaymentIntentID string `json:"payment_intent_id"`

	// Amount is the refunded amount
	Amount Amount `json:"amount"`

	// Status is the refund status
	Status RefundStatus `json:"status"`

	// Reason is the refund reason
	Reason RefundReason `json:"reason"`

	// Metadata is additional metadata
	Metadata map[string]string `json:"metadata,omitempty"`

	// FailureReason if refund failed
	FailureReason string `json:"failure_reason,omitempty"`

	// CreatedAt when refund was created
	CreatedAt time.Time `json:"created_at"`
}

// RefundRequest is a request to create a refund
type RefundRequest struct {
	// PaymentIntentID is the payment to refund
	PaymentIntentID string `json:"payment_intent_id"`

	// Amount is the amount to refund (nil for full refund)
	Amount *Amount `json:"amount,omitempty"`

	// Reason for the refund
	Reason RefundReason `json:"reason"`

	// Metadata is additional metadata
	Metadata map[string]string `json:"metadata,omitempty"`

	// IdempotencyKey to prevent duplicates
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

// ============================================================================
// Webhook Types
// ============================================================================

// WebhookEventType identifies the type of webhook event
type WebhookEventType string

const (
	WebhookEventPaymentIntentSucceeded    WebhookEventType = "payment_intent.succeeded"
	WebhookEventPaymentIntentFailed       WebhookEventType = "payment_intent.payment_failed"
	WebhookEventPaymentIntentCanceled     WebhookEventType = "payment_intent.canceled"
	WebhookEventPaymentIntentProcessing   WebhookEventType = "payment_intent.processing"
	WebhookEventChargeRefunded            WebhookEventType = "charge.refunded"
	WebhookEventChargeDisputeCreated      WebhookEventType = "charge.dispute.created"
	WebhookEventChargeDisputeUpdated      WebhookEventType = "charge.dispute.updated"
	WebhookEventChargeDisputeClosed       WebhookEventType = "charge.dispute.closed"
	WebhookEventChargeDisputeFundsWithdrawn  WebhookEventType = "charge.dispute.funds_withdrawn"
	WebhookEventChargeDisputeFundsReinstated WebhookEventType = "charge.dispute.funds_reinstated"
	WebhookEventPaymentMethodAttached     WebhookEventType = "payment_method.attached"
	WebhookEventPaymentMethodDetached     WebhookEventType = "payment_method.detached"
	WebhookEventCustomerCreated           WebhookEventType = "customer.created"
	WebhookEventCustomerDeleted           WebhookEventType = "customer.deleted"
)

// WebhookEvent represents a webhook event from the payment gateway
type WebhookEvent struct {
	// ID is the event ID
	ID string `json:"id"`

	// Type is the event type
	Type WebhookEventType `json:"type"`

	// Gateway is which gateway sent this
	Gateway GatewayType `json:"gateway"`

	// Payload is the raw event payload
	Payload []byte `json:"payload"`

	// Data is the parsed event data (type depends on event type)
	Data interface{} `json:"data"`

	// Signature is the webhook signature for verification
	Signature string `json:"signature"`

	// Timestamp is when the event occurred
	Timestamp time.Time `json:"timestamp"`

	// APIVersion is the gateway API version
	APIVersion string `json:"api_version,omitempty"`
}

// ============================================================================
// Dispute Types
// ============================================================================

// DisputeStatus represents dispute status
type DisputeStatus string

const (
	DisputeStatusOpen             DisputeStatus = "open"
	DisputeStatusNeedsResponse    DisputeStatus = "needs_response"
	DisputeStatusUnderReview      DisputeStatus = "under_review"
	DisputeStatusWon              DisputeStatus = "won"
	DisputeStatusLost             DisputeStatus = "lost"
	DisputeStatusAccepted         DisputeStatus = "accepted"
	DisputeStatusExpired          DisputeStatus = "expired"
)

// IsFinal returns true if the dispute is in a terminal state
func (s DisputeStatus) IsFinal() bool {
	switch s {
	case DisputeStatusWon, DisputeStatusLost, DisputeStatusAccepted, DisputeStatusExpired:
		return true
	default:
		return false
	}
}

// DisputeReason represents dispute reason
type DisputeReason string

const (
	DisputeReasonFraudulent        DisputeReason = "fraudulent"
	DisputeReasonDuplicate         DisputeReason = "duplicate"
	DisputeReasonProductNotReceived DisputeReason = "product_not_received"
	DisputeReasonUnrecognized      DisputeReason = "unrecognized"
	DisputeReasonGeneral           DisputeReason = "general"
)

// Dispute represents a payment dispute/chargeback
type Dispute struct {
	// ID is the dispute ID
	ID string `json:"id"`

	// Gateway is which gateway processed this dispute
	Gateway GatewayType `json:"gateway,omitempty"`

	// PaymentIntentID is the disputed payment
	PaymentIntentID string `json:"payment_intent_id"`

	// ChargeID is the underlying charge/transaction ID
	ChargeID string `json:"charge_id,omitempty"`

	// Amount is the disputed amount
	Amount Amount `json:"amount"`

	// Status is the dispute status
	Status DisputeStatus `json:"status"`

	// Reason is the dispute reason
	Reason DisputeReason `json:"reason"`

	// EvidenceDueBy is when evidence must be submitted
	EvidenceDueBy time.Time `json:"evidence_due_by"`

	// IsRefundable indicates if dispute can be refunded
	IsRefundable bool `json:"is_refundable"`

	// NetworkReasonCode is the card network reason code
	NetworkReasonCode string `json:"network_reason_code,omitempty"`

	// Metadata is additional key-value metadata
	Metadata map[string]string `json:"metadata,omitempty"`

	// CreatedAt when dispute was opened
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt when dispute was last updated
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// ============================================================================
// Conversion Types (Fiat-to-Crypto)
// ============================================================================

// ConversionRate represents an exchange rate for fiat-to-crypto
type ConversionRate struct {
	// FromCurrency is the fiat currency
	FromCurrency Currency `json:"from_currency"`

	// ToCrypto is the cryptocurrency symbol (e.g., "UVE")
	ToCrypto string `json:"to_crypto"`

	// Rate is the conversion rate (crypto per fiat unit)
	Rate sdkmath.LegacyDec `json:"rate"`

	// Timestamp when rate was fetched
	Timestamp time.Time `json:"timestamp"`

	// Source of the rate data
	Source string `json:"source"`
}

// ConversionQuote represents a fiat-to-crypto conversion quote
type ConversionQuote struct {
	// ID is the quote ID
	ID string `json:"id"`

	// FiatAmount is the fiat amount being converted
	FiatAmount Amount `json:"fiat_amount"`

	// CryptoAmount is the crypto amount to receive
	CryptoAmount sdkmath.Int `json:"crypto_amount"`

	// CryptoDenom is the cryptocurrency denomination
	CryptoDenom string `json:"crypto_denom"`

	// Rate is the applied conversion rate
	Rate ConversionRate `json:"rate"`

	// Fee is any conversion fee
	Fee Amount `json:"fee"`

	// ExpiresAt when quote expires
	ExpiresAt time.Time `json:"expires_at"`

	// DestinationAddress is the blockchain address to receive crypto
	DestinationAddress string `json:"destination_address"`
}

// IsExpired checks if the quote has expired
func (q ConversionQuote) IsExpired() bool {
	return time.Now().After(q.ExpiresAt)
}

// ============================================================================
// Conversion Ledger Types (PAY-002)
// ============================================================================

// ConversionStatus represents the status of a conversion execution
type ConversionStatus string

const (
	// ConversionStatusPending indicates conversion is pending execution
	ConversionStatusPending ConversionStatus = "pending"

	// ConversionStatusExecuting indicates conversion is being executed
	ConversionStatusExecuting ConversionStatus = "executing"

	// ConversionStatusCompleted indicates conversion completed successfully
	ConversionStatusCompleted ConversionStatus = "completed"

	// ConversionStatusFailed indicates conversion failed
	ConversionStatusFailed ConversionStatus = "failed"

	// ConversionStatusRefunded indicates conversion was refunded
	ConversionStatusRefunded ConversionStatus = "refunded"

	// ConversionStatusReconciling indicates conversion needs manual reconciliation
	ConversionStatusReconciling ConversionStatus = "reconciling"
)

// IsFinal checks if the status is a terminal state
func (s ConversionStatus) IsFinal() bool {
	switch s {
	case ConversionStatusCompleted, ConversionStatusFailed, ConversionStatusRefunded:
		return true
	default:
		return false
	}
}

// IsRetryable checks if conversion can be retried
func (s ConversionStatus) IsRetryable() bool {
	switch s {
	case ConversionStatusPending, ConversionStatusFailed, ConversionStatusReconciling:
		return true
	default:
		return false
	}
}

// ConversionErrorCode categorizes conversion failures
type ConversionErrorCode string

const (
	// ConversionErrorNone indicates no error
	ConversionErrorNone ConversionErrorCode = ""

	// ConversionErrorQuoteExpired indicates quote expired before execution
	ConversionErrorQuoteExpired ConversionErrorCode = "quote_expired"

	// ConversionErrorPaymentNotSucceeded indicates payment was not successful
	ConversionErrorPaymentNotSucceeded ConversionErrorCode = "payment_not_succeeded"

	// ConversionErrorInsufficientTreasury indicates treasury has insufficient funds
	ConversionErrorInsufficientTreasury ConversionErrorCode = "insufficient_treasury"

	// ConversionErrorTransferFailed indicates on-chain transfer failed
	ConversionErrorTransferFailed ConversionErrorCode = "transfer_failed"

	// ConversionErrorInvalidAddress indicates destination address is invalid
	ConversionErrorInvalidAddress ConversionErrorCode = "invalid_address"

	// ConversionErrorDuplicate indicates duplicate execution attempt
	ConversionErrorDuplicate ConversionErrorCode = "duplicate"

	// ConversionErrorInternal indicates internal system error
	ConversionErrorInternal ConversionErrorCode = "internal_error"
)

// IsRetryable checks if the error is retryable
func (e ConversionErrorCode) IsRetryable() bool {
	switch e {
	case ConversionErrorInsufficientTreasury, ConversionErrorTransferFailed, ConversionErrorInternal:
		return true
	default:
		return false
	}
}

// ConversionLedgerEntry represents a conversion execution record with full audit trail
type ConversionLedgerEntry struct {
	// ID is the unique ledger entry ID
	ID string `json:"id"`

	// IdempotencyKey prevents duplicate executions
	IdempotencyKey string `json:"idempotency_key"`

	// QuoteID is the associated conversion quote
	QuoteID string `json:"quote_id"`

	// PaymentIntentID is the payment that funded this conversion
	PaymentIntentID string `json:"payment_intent_id"`

	// Status is the current conversion status
	Status ConversionStatus `json:"status"`

	// FiatAmount is the fiat amount converted
	FiatAmount Amount `json:"fiat_amount"`

	// CryptoAmount is the crypto amount transferred
	CryptoAmount sdkmath.Int `json:"crypto_amount"`

	// CryptoDenom is the cryptocurrency denomination
	CryptoDenom string `json:"crypto_denom"`

	// DestinationAddress is the blockchain address receiving funds
	DestinationAddress string `json:"destination_address"`

	// TreasuryAddress is the treasury address funds were sent from
	TreasuryAddress string `json:"treasury_address,omitempty"`

	// TxHash is the on-chain transaction hash (when completed)
	TxHash string `json:"tx_hash,omitempty"`

	// BlockHeight is the block height of the transfer (when completed)
	BlockHeight int64 `json:"block_height,omitempty"`

	// ErrorCode categorizes any failure
	ErrorCode ConversionErrorCode `json:"error_code,omitempty"`

	// ErrorMessage provides detailed error information
	ErrorMessage string `json:"error_message,omitempty"`

	// RetryCount tracks retry attempts
	RetryCount int `json:"retry_count"`

	// MaxRetries is the maximum retry attempts
	MaxRetries int `json:"max_retries"`

	// LastRetryAt is when the last retry was attempted
	LastRetryAt *time.Time `json:"last_retry_at,omitempty"`

	// NextRetryAt is when the next retry should be attempted
	NextRetryAt *time.Time `json:"next_retry_at,omitempty"`

	// ConversionRate is the rate used for conversion
	ConversionRate ConversionRate `json:"conversion_rate"`

	// Fee is the conversion fee charged
	Fee Amount `json:"fee"`

	// CustomerID is the payment customer ID
	CustomerID string `json:"customer_id,omitempty"`

	// Metadata is additional context data
	Metadata map[string]string `json:"metadata,omitempty"`

	// CreatedAt is when the entry was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the entry was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// CompletedAt is when conversion completed (success or final failure)
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// NewConversionLedgerEntry creates a new ledger entry from a quote and payment
func NewConversionLedgerEntry(
	id string,
	idempotencyKey string,
	quote ConversionQuote,
	paymentIntentID string,
	maxRetries int,
) *ConversionLedgerEntry {
	now := time.Now()
	return &ConversionLedgerEntry{
		ID:                 id,
		IdempotencyKey:     idempotencyKey,
		QuoteID:            quote.ID,
		PaymentIntentID:    paymentIntentID,
		Status:             ConversionStatusPending,
		FiatAmount:         quote.FiatAmount,
		CryptoAmount:       quote.CryptoAmount,
		CryptoDenom:        quote.CryptoDenom,
		DestinationAddress: quote.DestinationAddress,
		ConversionRate:     quote.Rate,
		Fee:                quote.Fee,
		RetryCount:         0,
		MaxRetries:         maxRetries,
		CreatedAt:          now,
		UpdatedAt:          now,
		Metadata:           make(map[string]string),
	}
}

// MarkExecuting transitions to executing state
func (e *ConversionLedgerEntry) MarkExecuting() error {
	if e.Status != ConversionStatusPending && e.Status != ConversionStatusReconciling {
		return errors.New("cannot execute: invalid status transition")
	}
	e.Status = ConversionStatusExecuting
	e.UpdatedAt = time.Now()
	return nil
}

// MarkCompleted transitions to completed state
func (e *ConversionLedgerEntry) MarkCompleted(txHash string, blockHeight int64, treasuryAddr string) {
	now := time.Now()
	e.Status = ConversionStatusCompleted
	e.TxHash = txHash
	e.BlockHeight = blockHeight
	e.TreasuryAddress = treasuryAddr
	e.ErrorCode = ConversionErrorNone
	e.ErrorMessage = ""
	e.UpdatedAt = now
	e.CompletedAt = &now
}

// MarkFailed transitions to failed state
func (e *ConversionLedgerEntry) MarkFailed(errCode ConversionErrorCode, errMsg string) {
	now := time.Now()
	e.Status = ConversionStatusFailed
	e.ErrorCode = errCode
	e.ErrorMessage = errMsg
	e.UpdatedAt = now
	e.CompletedAt = &now
}

// MarkForRetry schedules a retry with exponential backoff
func (e *ConversionLedgerEntry) MarkForRetry(errCode ConversionErrorCode, errMsg string, baseDelay time.Duration) bool {
	if e.RetryCount >= e.MaxRetries {
		e.MarkFailed(errCode, errMsg)
		return false
	}

	e.RetryCount++
	now := time.Now()
	e.LastRetryAt = &now
	e.ErrorCode = errCode
	e.ErrorMessage = errMsg
	e.Status = ConversionStatusPending
	e.UpdatedAt = now

	// Exponential backoff: baseDelay * 2^(retryCount-1)
	delay := baseDelay * time.Duration(1<<(e.RetryCount-1))
	nextRetry := now.Add(delay)
	e.NextRetryAt = &nextRetry

	return true
}

// MarkReconciling transitions to reconciling state for manual intervention
func (e *ConversionLedgerEntry) MarkReconciling(errCode ConversionErrorCode, errMsg string) {
	e.Status = ConversionStatusReconciling
	e.ErrorCode = errCode
	e.ErrorMessage = errMsg
	e.UpdatedAt = time.Now()
}

// MarkRefunded transitions to refunded state
func (e *ConversionLedgerEntry) MarkRefunded() {
	now := time.Now()
	e.Status = ConversionStatusRefunded
	e.UpdatedAt = now
	e.CompletedAt = &now
}

// CanRetry checks if this entry can be retried
func (e *ConversionLedgerEntry) CanRetry() bool {
	if !e.Status.IsRetryable() {
		return false
	}
	if e.RetryCount >= e.MaxRetries {
		return false
	}
	if e.NextRetryAt != nil && time.Now().Before(*e.NextRetryAt) {
		return false
	}
	return e.ErrorCode.IsRetryable()
}

// IsReadyForExecution checks if entry is ready to execute
func (e *ConversionLedgerEntry) IsReadyForExecution() bool {
	if e.Status != ConversionStatusPending {
		return false
	}
	if e.NextRetryAt != nil && time.Now().Before(*e.NextRetryAt) {
		return false
	}
	return true
}

// ConversionExecutionRequest is a request to execute a conversion
type ConversionExecutionRequest struct {
	// Quote is the conversion quote to execute
	Quote ConversionQuote `json:"quote"`

	// PaymentIntentID is the successful payment funding this conversion
	PaymentIntentID string `json:"payment_intent_id"`

	// IdempotencyKey prevents duplicate executions
	IdempotencyKey string `json:"idempotency_key"`

	// Metadata is additional context
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ConversionExecutionResult is the result of a conversion execution
type ConversionExecutionResult struct {
	// LedgerEntry is the created/updated ledger entry
	LedgerEntry *ConversionLedgerEntry `json:"ledger_entry"`

	// Success indicates if the conversion was successful
	Success bool `json:"success"`

	// TxHash is the transaction hash if successful
	TxHash string `json:"tx_hash,omitempty"`

	// Error is the error if failed
	Error error `json:"-"`

	// ErrorCode is the error category
	ErrorCode ConversionErrorCode `json:"error_code,omitempty"`

	// AlreadyCompleted indicates this was a duplicate idempotent request
	AlreadyCompleted bool `json:"already_completed"`
}

// TreasuryTransferResult is the result of a treasury transfer operation
type TreasuryTransferResult struct {
	// TxHash is the transaction hash
	TxHash string `json:"tx_hash"`

	// BlockHeight is the block height
	BlockHeight int64 `json:"block_height"`

	// TreasuryAddress is the source address
	TreasuryAddress string `json:"treasury_address"`

	// Success indicates if transfer succeeded
	Success bool `json:"success"`

	// Error is any error that occurred
	Error error `json:"-"`
}
