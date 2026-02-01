// Package payment provides payment gateway integration for Visa/Mastercard.
//
// VE-906: Payment gateway integration for fiat-to-crypto onramp
package payment

import (
	"context"
)

// ============================================================================
// Gateway Interface
// ============================================================================

// Gateway defines the interface for payment gateway adapters.
// Each adapter implements gateway-specific logic (Stripe, Adyen, etc.)
type Gateway interface {
	// Name returns the gateway name
	Name() string

	// Type returns the gateway type
	Type() GatewayType

	// IsHealthy checks if the gateway is operational
	IsHealthy(ctx context.Context) bool

	// Close releases gateway resources
	Close() error

	// ---- Customer Management ----

	// CreateCustomer creates a new customer in the gateway
	CreateCustomer(ctx context.Context, req CreateCustomerRequest) (Customer, error)

	// GetCustomer retrieves a customer by ID
	GetCustomer(ctx context.Context, customerID string) (Customer, error)

	// UpdateCustomer updates customer details
	UpdateCustomer(ctx context.Context, customerID string, req UpdateCustomerRequest) (Customer, error)

	// DeleteCustomer deletes a customer
	DeleteCustomer(ctx context.Context, customerID string) error

	// ---- Payment Methods ----

	// AttachPaymentMethod attaches a tokenized card to a customer
	AttachPaymentMethod(ctx context.Context, customerID string, token CardToken) (string, error)

	// DetachPaymentMethod removes a payment method from a customer
	DetachPaymentMethod(ctx context.Context, paymentMethodID string) error

	// ListPaymentMethods lists a customer's payment methods
	ListPaymentMethods(ctx context.Context, customerID string) ([]CardToken, error)

	// ---- Payment Intents ----

	// CreatePaymentIntent creates a new payment intent
	CreatePaymentIntent(ctx context.Context, req PaymentIntentRequest) (PaymentIntent, error)

	// GetPaymentIntent retrieves a payment intent
	GetPaymentIntent(ctx context.Context, paymentIntentID string) (PaymentIntent, error)

	// ConfirmPaymentIntent confirms a payment intent
	ConfirmPaymentIntent(ctx context.Context, paymentIntentID string, paymentMethodID string) (PaymentIntent, error)

	// CancelPaymentIntent cancels a payment intent
	CancelPaymentIntent(ctx context.Context, paymentIntentID string, reason string) (PaymentIntent, error)

	// CapturePaymentIntent captures an authorized payment
	CapturePaymentIntent(ctx context.Context, paymentIntentID string, amount *Amount) (PaymentIntent, error)

	// ---- Refunds ----

	// CreateRefund creates a refund for a payment
	CreateRefund(ctx context.Context, req RefundRequest) (Refund, error)

	// GetRefund retrieves a refund by ID
	GetRefund(ctx context.Context, refundID string) (Refund, error)

	// ---- Webhooks ----

	// ValidateWebhook verifies a webhook signature
	ValidateWebhook(payload []byte, signature string) error

	// ParseWebhookEvent parses a webhook event
	ParseWebhookEvent(payload []byte) (WebhookEvent, error)
}

// CreateCustomerRequest is a request to create a customer
type CreateCustomerRequest struct {
	Email       string            `json:"email"`
	Name        string            `json:"name,omitempty"`
	Phone       string            `json:"phone,omitempty"`
	VEIDAddress string            `json:"veid_address"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// UpdateCustomerRequest is a request to update a customer
type UpdateCustomerRequest struct {
	Email                  *string           `json:"email,omitempty"`
	Name                   *string           `json:"name,omitempty"`
	Phone                  *string           `json:"phone,omitempty"`
	DefaultPaymentMethodID *string           `json:"default_payment_method_id,omitempty"`
	Metadata               map[string]string `json:"metadata,omitempty"`
}

// ============================================================================
// Token Manager Interface
// ============================================================================

// TokenManager handles card tokenization operations
// Note: All tokenization happens client-side via gateway SDKs.
// This interface manages the server-side token lifecycle.
type TokenManager interface {
	// ValidateToken validates a card token
	ValidateToken(ctx context.Context, token CardToken) error

	// GetTokenDetails retrieves token details from gateway
	GetTokenDetails(ctx context.Context, tokenID string) (CardToken, error)

	// RefreshToken refreshes an expiring token
	RefreshToken(ctx context.Context, tokenID string) (CardToken, error)

	// RevokeToken revokes a token
	RevokeToken(ctx context.Context, tokenID string) error
}

// ============================================================================
// Webhook Handler Interface
// ============================================================================

// WebhookHandler processes payment webhook events
type WebhookHandler interface {
	// HandleEvent processes a webhook event
	HandleEvent(ctx context.Context, event WebhookEvent) error

	// RegisterHandler registers a handler for a specific event type
	RegisterHandler(eventType WebhookEventType, handler EventHandler)

	// UnregisterHandler removes a handler
	UnregisterHandler(eventType WebhookEventType)
}

// EventHandler is a function that handles a specific webhook event
type EventHandler func(ctx context.Context, event WebhookEvent) error

// ============================================================================
// SCA Handler Interface
// ============================================================================

// SCAHandler handles 3D Secure / Strong Customer Authentication flows
type SCAHandler interface {
	// InitiateSCA initiates 3D Secure authentication
	InitiateSCA(ctx context.Context, paymentIntent PaymentIntent) (SCAChallenge, error)

	// CompleteSCA completes 3D Secure after customer authentication
	CompleteSCA(ctx context.Context, challengeID string, result SCAResult) (PaymentIntent, error)

	// GetSCAStatus gets the current SCA status
	GetSCAStatus(ctx context.Context, paymentIntentID string) (SCAResult, error)
}

// SCAChallenge represents a 3D Secure challenge
type SCAChallenge struct {
	// ID is the challenge ID
	ID string `json:"id"`

	// PaymentIntentID is the associated payment intent
	PaymentIntentID string `json:"payment_intent_id"`

	// RedirectURL is where to redirect the customer
	RedirectURL string `json:"redirect_url"`

	// ReturnURL is where customer returns after auth
	ReturnURL string `json:"return_url"`

	// ExpiresAt is when the challenge expires
	ExpiresAt string `json:"expires_at"`

	// ThreeDSVersion is the 3DS protocol version
	ThreeDSVersion string `json:"three_ds_version"`
}

// ============================================================================
// Dispute Handler Interface
// ============================================================================

// DisputeHandler handles payment disputes and chargebacks
type DisputeHandler interface {
	// GetDispute retrieves a dispute by ID
	GetDispute(ctx context.Context, disputeID string) (Dispute, error)

	// ListDisputes lists disputes for a payment
	ListDisputes(ctx context.Context, paymentIntentID string) ([]Dispute, error)

	// SubmitEvidence submits evidence for a dispute
	SubmitEvidence(ctx context.Context, disputeID string, evidence DisputeEvidence) error

	// AcceptDispute accepts (concedes) a dispute
	AcceptDispute(ctx context.Context, disputeID string) error
}

// DisputeEvidence contains evidence for dispute response
type DisputeEvidence struct {
	// ProductDescription describes what was sold
	ProductDescription string `json:"product_description,omitempty"`

	// CustomerEmail is the customer's email
	CustomerEmail string `json:"customer_email,omitempty"`

	// CustomerPurchaseIP is the customer's IP address
	CustomerPurchaseIP string `json:"customer_purchase_ip,omitempty"`

	// BillingAddress is the billing address used
	BillingAddress string `json:"billing_address,omitempty"`

	// Receipt is a receipt or invoice
	Receipt []byte `json:"receipt,omitempty"`

	// UncategorizedText is any other evidence
	UncategorizedText string `json:"uncategorized_text,omitempty"`
}

// ============================================================================
// Conversion Service Interface
// ============================================================================

// ConversionService handles fiat-to-crypto conversions
type ConversionService interface {
	// GetConversionRate gets the current conversion rate
	GetConversionRate(ctx context.Context, fromCurrency Currency, toCrypto string) (ConversionRate, error)

	// CreateConversionQuote creates a conversion quote
	CreateConversionQuote(ctx context.Context, req ConversionQuoteRequest) (ConversionQuote, error)

	// ExecuteConversion executes a conversion after payment succeeds
	ExecuteConversion(ctx context.Context, quote ConversionQuote, paymentIntentID string) error
}

// ConversionQuoteRequest is a request to create a conversion quote
type ConversionQuoteRequest struct {
	// FiatAmount is the fiat amount to convert
	FiatAmount Amount `json:"fiat_amount"`

	// CryptoDenom is the target crypto denomination
	CryptoDenom string `json:"crypto_denom"`

	// DestinationAddress is the blockchain address
	DestinationAddress string `json:"destination_address"`
}

// ============================================================================
// Treasury Transfer Interface (PAY-002)
// ============================================================================

// TreasuryTransfer handles on-chain crypto transfers from treasury
type TreasuryTransfer interface {
	// SendFromTreasury transfers crypto from treasury to destination
	SendFromTreasury(ctx context.Context, req TreasuryTransferRequest) (*TreasuryTransferResult, error)

	// GetTreasuryBalance returns the treasury balance for a given denom
	GetTreasuryBalance(ctx context.Context, denom string) (TreasuryBalance, error)

	// ValidateAddress validates a blockchain address
	ValidateAddress(ctx context.Context, address string) error
}

// TreasuryTransferRequest is a request to transfer from treasury
type TreasuryTransferRequest struct {
	// DestinationAddress is the recipient address
	DestinationAddress string `json:"destination_address"`

	// Amount is the amount to transfer (in smallest units)
	Amount int64 `json:"amount"`

	// Denom is the token denomination
	Denom string `json:"denom"`

	// Memo is an optional transaction memo
	Memo string `json:"memo,omitempty"`

	// IdempotencyKey prevents duplicate transfers
	IdempotencyKey string `json:"idempotency_key"`
}

// TreasuryBalance represents treasury balance for a denom
type TreasuryBalance struct {
	// Denom is the token denomination
	Denom string `json:"denom"`

	// Available is the available balance
	Available int64 `json:"available"`

	// Reserved is the reserved balance (pending transfers)
	Reserved int64 `json:"reserved"`

	// Total is the total balance
	Total int64 `json:"total"`
}

// ============================================================================
// Conversion Executor Interface (PAY-002)
// ============================================================================

// ConversionExecutor handles conversion execution with idempotency
type ConversionExecutor interface {
	// ExecuteConversion executes a conversion with idempotency guarantees
	ExecuteConversion(ctx context.Context, req ConversionExecutionRequest) (*ConversionExecutionResult, error)

	// GetLedgerEntry retrieves a ledger entry by ID
	GetLedgerEntry(ctx context.Context, id string) (*ConversionLedgerEntry, error)

	// GetLedgerEntryByIdempotencyKey retrieves a ledger entry by idempotency key
	GetLedgerEntryByIdempotencyKey(ctx context.Context, key string) (*ConversionLedgerEntry, error)

	// RetryFailedConversion retries a failed conversion
	RetryFailedConversion(ctx context.Context, id string) (*ConversionExecutionResult, error)

	// ReconcileConversion manually reconciles a stuck conversion
	ReconcileConversion(ctx context.Context, id string, txHash string, blockHeight int64) error

	// RefundConversion refunds a failed conversion
	RefundConversion(ctx context.Context, id string, reason string) error

	// ListPendingConversions lists conversions ready for execution
	ListPendingConversions(ctx context.Context) ([]*ConversionLedgerEntry, error)

	// ListConversionsForReconciliation lists conversions needing manual reconciliation
	ListConversionsForReconciliation(ctx context.Context) ([]*ConversionLedgerEntry, error)
}

// ConversionLedgerStore persists conversion ledger entries
type ConversionLedgerStore interface {
	// Save saves or updates a ledger entry
	Save(ctx context.Context, entry *ConversionLedgerEntry) error

	// GetByID retrieves an entry by ID
	GetByID(ctx context.Context, id string) (*ConversionLedgerEntry, error)

	// GetByIdempotencyKey retrieves an entry by idempotency key
	GetByIdempotencyKey(ctx context.Context, key string) (*ConversionLedgerEntry, error)

	// ListByStatus lists entries by status
	ListByStatus(ctx context.Context, status ConversionStatus) ([]*ConversionLedgerEntry, error)

	// ListPendingReadyForExecution lists pending entries ready for execution
	ListPendingReadyForExecution(ctx context.Context) ([]*ConversionLedgerEntry, error)
}

// ============================================================================
// Payment Service Interface
// ============================================================================

// Service is the main payment service interface combining all functionality
type Service interface {
	Gateway
	TokenManager
	WebhookHandler
	SCAHandler
	DisputeHandler
	ConversionService

	// GetGateway returns the underlying gateway adapter
	GetGateway() Gateway

	// GetConfig returns the service configuration
	GetConfig() Config
}

