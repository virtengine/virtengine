// Package offramp provides fiat off-ramp integration for token-to-fiat payouts.
//
// VE-5E: Fiat off-ramp integration for PayPal/ACH payouts
package offramp

import (
	"context"
)

// ============================================================================
// Off-Ramp Provider Interface
// ============================================================================

// Provider defines the interface for off-ramp payout providers.
// Each provider implements the specific logic for PayPal, ACH, etc.
type Provider interface {
	// Name returns the provider name
	Name() string

	// Type returns the provider type
	Type() ProviderType

	// IsHealthy checks if the provider is operational
	IsHealthy(ctx context.Context) bool

	// Close releases provider resources
	Close() error

	// ---- Payout Operations ----

	// CreatePayout initiates a payout to the destination
	CreatePayout(ctx context.Context, intent *PayoutIntent) error

	// GetPayoutStatus retrieves the current status of a payout
	GetPayoutStatus(ctx context.Context, providerPayoutID string) (PayoutStatus, error)

	// CancelPayout cancels a pending payout
	CancelPayout(ctx context.Context, providerPayoutID string) error

	// ---- Webhooks ----

	// ValidateWebhook verifies a webhook signature
	ValidateWebhook(payload []byte, signature string) error

	// ParseWebhookEvent parses a webhook event
	ParseWebhookEvent(payload []byte) (*WebhookEvent, error)

	// ---- Reconciliation ----

	// GetSettlementReport retrieves a settlement report for reconciliation
	GetSettlementReport(ctx context.Context, req SettlementReportRequest) (*SettlementReport, error)
}

// SettlementReportRequest is a request for a settlement report
type SettlementReportRequest struct {
	// StartDate is the start of the reporting period
	StartDate string `json:"start_date"`

	// EndDate is the end of the reporting period
	EndDate string `json:"end_date"`

	// Format is the report format (csv, json)
	Format string `json:"format,omitempty"`
}

// SettlementReport represents a provider's settlement report
type SettlementReport struct {
	// ReportID is the unique report identifier
	ReportID string `json:"report_id"`

	// Provider is the provider that generated the report
	Provider ProviderType `json:"provider"`

	// StartDate is the start of the reporting period
	StartDate string `json:"start_date"`

	// EndDate is the end of the reporting period
	EndDate string `json:"end_date"`

	// TotalPayouts is the total number of payouts
	TotalPayouts int `json:"total_payouts"`

	// TotalAmount is the total amount paid out
	TotalAmount int64 `json:"total_amount"`

	// TotalFees is the total fees charged
	TotalFees int64 `json:"total_fees"`

	// Transactions contains individual transaction records
	Transactions []SettlementTransaction `json:"transactions"`

	// GeneratedAt is when the report was generated
	GeneratedAt string `json:"generated_at"`
}

// SettlementTransaction represents a single settlement transaction
type SettlementTransaction struct {
	// TransactionID is the provider's transaction ID
	TransactionID string `json:"transaction_id"`

	// PayoutID is our payout ID
	PayoutID string `json:"payout_id"`

	// Amount is the transaction amount
	Amount int64 `json:"amount"`

	// Fee is the fee charged
	Fee int64 `json:"fee"`

	// Status is the transaction status
	Status string `json:"status"`

	// ProcessedAt is when the transaction was processed
	ProcessedAt string `json:"processed_at"`
}

// ============================================================================
// KYC Gate Interface
// ============================================================================

// KYCGate defines the interface for KYC verification checks.
type KYCGate interface {
	// CheckKYCStatus checks the KYC status for an account
	CheckKYCStatus(ctx context.Context, accountAddress string, veidID string) (KYCCheckResult, error)

	// GetVerificationLevel returns the verification level for an account
	GetVerificationLevel(ctx context.Context, accountAddress string) (KYCVerificationLevel, error)

	// RequireVerification returns an error if verification is required
	RequireVerification(ctx context.Context, accountAddress string, requiredLevel KYCVerificationLevel) error
}

// KYCCheckResult contains the result of a KYC check
type KYCCheckResult struct {
	// Status is the KYC status
	Status KYCStatus `json:"status"`

	// Level is the verification level
	Level KYCVerificationLevel `json:"level"`

	// VEIDID is the verified identity ID
	VEIDID string `json:"veid_id"`

	// VerifiedAt is when verification completed
	VerifiedAt string `json:"verified_at,omitempty"`

	// ExpiresAt is when verification expires
	ExpiresAt string `json:"expires_at,omitempty"`

	// RequiresRevalidation indicates if revalidation is needed
	RequiresRevalidation bool `json:"requires_revalidation"`

	// Message contains additional information
	Message string `json:"message,omitempty"`
}

// ============================================================================
// AML Screener Interface
// ============================================================================

// AMLScreener defines the interface for AML screening.
type AMLScreener interface {
	// Screen performs AML screening on a user
	Screen(ctx context.Context, req AMLScreenRequest) (*AMLScreenResult, error)

	// GetScreeningStatus retrieves the status of a screening
	GetScreeningStatus(ctx context.Context, screeningID string) (*AMLScreenResult, error)
}

// AMLScreenRequest is a request to screen a user
type AMLScreenRequest struct {
	// AccountAddress is the blockchain account address
	AccountAddress string `json:"account_address"`

	// VEIDID is the verified identity ID
	VEIDID string `json:"veid_id"`

	// FullName is the user's full name
	FullName string `json:"full_name"`

	// DateOfBirth is the user's date of birth
	DateOfBirth string `json:"date_of_birth,omitempty"`

	// Country is the user's country
	Country string `json:"country"`

	// PayoutAmount is the payout amount being requested
	PayoutAmount int64 `json:"payout_amount"`

	// PayoutCurrency is the payout currency
	PayoutCurrency string `json:"payout_currency"`
}

// AMLScreenResult contains the result of AML screening
type AMLScreenResult struct {
	// ScreeningID is the unique screening identifier
	ScreeningID string `json:"screening_id"`

	// Status is the AML status
	Status AMLStatus `json:"status"`

	// RiskScore is the risk score (0-100)
	RiskScore int `json:"risk_score"`

	// Matches contains any matches found
	Matches []AMLMatch `json:"matches,omitempty"`

	// ScreenedAt is when screening was performed
	ScreenedAt string `json:"screened_at"`

	// ExpiresAt is when the screening expires
	ExpiresAt string `json:"expires_at,omitempty"`

	// ReviewRequired indicates if manual review is needed
	ReviewRequired bool `json:"review_required"`

	// Notes contains screening notes
	Notes string `json:"notes,omitempty"`
}

// AMLMatch represents a match from AML screening
type AMLMatch struct {
	// Type is the match type (sanctions, pep, adverse_media)
	Type string `json:"type"`

	// ListName is the name of the list matched
	ListName string `json:"list_name"`

	// MatchScore is the confidence score (0-100)
	MatchScore int `json:"match_score"`

	// MatchedName is the name that matched
	MatchedName string `json:"matched_name"`

	// Details contains match details
	Details string `json:"details,omitempty"`
}

// ============================================================================
// Payout Store Interface
// ============================================================================

// PayoutStore defines the interface for payout storage.
type PayoutStore interface {
	// Save saves or updates a payout intent
	Save(ctx context.Context, payout *PayoutIntent) error

	// GetByID retrieves a payout by ID
	GetByID(ctx context.Context, id string) (*PayoutIntent, error)

	// GetByIdempotencyKey retrieves a payout by idempotency key
	GetByIdempotencyKey(ctx context.Context, key string) (*PayoutIntent, error)

	// GetByProviderPayoutID retrieves a payout by provider payout ID
	GetByProviderPayoutID(ctx context.Context, providerPayoutID string) (*PayoutIntent, error)

	// ListByAccount lists payouts for an account
	ListByAccount(ctx context.Context, accountAddress string, limit int) ([]*PayoutIntent, error)

	// ListByStatus lists payouts by status
	ListByStatus(ctx context.Context, status PayoutStatus, limit int) ([]*PayoutIntent, error)

	// ListPendingReconciliation lists payouts pending reconciliation
	ListPendingReconciliation(ctx context.Context) ([]*PayoutIntent, error)

	// Delete deletes a payout (for testing only)
	Delete(ctx context.Context, id string) error
}

// ============================================================================
// Reconciliation Store Interface
// ============================================================================

// ReconciliationStore defines the interface for reconciliation storage.
type ReconciliationStore interface {
	// Save saves a reconciliation record
	Save(ctx context.Context, record *ReconciliationRecord) error

	// GetByPayoutID retrieves a reconciliation record by payout ID
	GetByPayoutID(ctx context.Context, payoutID string) (*ReconciliationRecord, error)

	// ListByStatus lists records by status
	ListByStatus(ctx context.Context, status ReconciliationStatus) ([]*ReconciliationRecord, error)

	// ListMismatches lists records with mismatches
	ListMismatches(ctx context.Context) ([]*ReconciliationRecord, error)
}

// ============================================================================
// Payout Limits Store Interface
// ============================================================================

// LimitsStore defines the interface for payout limits storage.
type LimitsStore interface {
	// GetLimits retrieves limits for an account
	GetLimits(ctx context.Context, accountAddress string) (*PayoutLimits, error)

	// UpdateUsage updates the usage for an account
	UpdateUsage(ctx context.Context, accountAddress string, amount int64) error

	// ResetDailyUsage resets daily usage for all accounts
	ResetDailyUsage(ctx context.Context) error

	// ResetMonthlyUsage resets monthly usage for all accounts
	ResetMonthlyUsage(ctx context.Context) error
}

// ============================================================================
// Webhook Handler Interface
// ============================================================================

// WebhookHandler processes off-ramp webhook events.
type WebhookHandler interface {
	// HandleEvent processes a webhook event
	HandleEvent(ctx context.Context, event *WebhookEvent) error

	// RegisterHandler registers a handler for a specific event type
	RegisterHandler(eventType WebhookEventType, handler EventHandler)

	// UnregisterHandler removes a handler
	UnregisterHandler(eventType WebhookEventType)
}

// EventHandler is a function that handles a specific webhook event
type EventHandler func(ctx context.Context, event *WebhookEvent) error

// ============================================================================
// Off-Ramp Service Interface
// ============================================================================

// Service is the main off-ramp service interface combining all functionality.
type Service interface {
	// ---- Payout Operations ----

	// CreatePayoutQuote creates a quote for a payout
	CreatePayoutQuote(ctx context.Context, req CreatePayoutRequest) (*PayoutQuote, error)

	// ExecutePayout executes a payout from an approved quote
	ExecutePayout(ctx context.Context, quoteID string) (*PayoutIntent, error)

	// GetPayout retrieves a payout by ID
	GetPayout(ctx context.Context, payoutID string) (*PayoutIntent, error)

	// CancelPayout cancels a pending payout
	CancelPayout(ctx context.Context, payoutID string, reason string) error

	// ListPayouts lists payouts for an account
	ListPayouts(ctx context.Context, accountAddress string, limit int) ([]*PayoutIntent, error)

	// ---- KYC/AML ----

	// CheckPayoutEligibility checks if an account can make a payout
	CheckPayoutEligibility(ctx context.Context, accountAddress string, amount int64) (*EligibilityResult, error)

	// ---- Limits ----

	// GetPayoutLimits retrieves payout limits for an account
	GetPayoutLimits(ctx context.Context, accountAddress string) (*PayoutLimits, error)

	// ---- Webhooks ----

	// HandleWebhook handles an incoming webhook
	HandleWebhook(ctx context.Context, provider ProviderType, payload []byte, signature string) error

	// ---- Reconciliation ----

	// RunReconciliation runs the reconciliation job
	RunReconciliation(ctx context.Context) (*ReconciliationResult, error)

	// GetReconciliationRecord retrieves a reconciliation record
	GetReconciliationRecord(ctx context.Context, payoutID string) (*ReconciliationRecord, error)

	// ---- Health ----

	// HealthCheck returns the health status
	HealthCheck(ctx context.Context) (*HealthStatus, error)

	// Close closes the service
	Close() error
}

// EligibilityResult contains the result of eligibility check
type EligibilityResult struct {
	// Eligible indicates if the account can make a payout
	Eligible bool `json:"eligible"`

	// Reason explains why the account is not eligible
	Reason string `json:"reason,omitempty"`

	// KYCStatus is the current KYC status
	KYCStatus KYCStatus `json:"kyc_status"`

	// AMLStatus is the current AML status
	AMLStatus AMLStatus `json:"aml_status"`

	// Limits contains the current limits
	Limits *PayoutLimits `json:"limits,omitempty"`

	// RequiredActions lists actions needed for eligibility
	RequiredActions []string `json:"required_actions,omitempty"`
}

// ReconciliationResult contains the result of a reconciliation run
type ReconciliationResult struct {
	// PayoutsProcessed is the number of payouts processed
	PayoutsProcessed int `json:"payouts_processed"`

	// Matched is the number of matched records
	Matched int `json:"matched"`

	// Mismatched is the number of mismatched records
	Mismatched int `json:"mismatched"`

	// Missing is the number of missing records
	Missing int `json:"missing"`

	// Errors is the number of errors encountered
	Errors int `json:"errors"`

	// Duration is how long reconciliation took
	Duration string `json:"duration"`

	// Records contains individual reconciliation records
	Records []*ReconciliationRecord `json:"records,omitempty"`
}

// HealthStatus contains health status information
type HealthStatus struct {
	// Healthy indicates if the service is healthy
	Healthy bool `json:"healthy"`

	// Status is a status message
	Status string `json:"status"`

	// Providers contains provider health status
	Providers map[ProviderType]bool `json:"providers"`

	// LastReconciliation is when reconciliation last ran
	LastReconciliation string `json:"last_reconciliation,omitempty"`

	// PendingPayouts is the number of pending payouts
	PendingPayouts int `json:"pending_payouts"`

	// Warnings contains any warnings
	Warnings []string `json:"warnings,omitempty"`
}
