// Package govdata provides government data source integration for identity verification.
//
// VE-909: Government data integration for authoritative identity verification
package govdata

import (
	"context"
	"time"
)

// ============================================================================
// Service Interface
// ============================================================================

// Service defines the main government data service interface
type Service interface {
	// Lifecycle methods

	// Start starts the service
	Start(ctx context.Context) error

	// Stop stops the service gracefully
	Stop(ctx context.Context) error

	// IsHealthy returns true if the service is healthy
	IsHealthy() bool

	// GetStatus returns the service status
	GetStatus() ServiceStatus

	// Verification methods

	// VerifyDocument verifies a government document
	VerifyDocument(ctx context.Context, req *VerificationRequest) (*VerificationResponse, error)

	// VerifyDocumentBatch verifies multiple documents in batch
	VerifyDocumentBatch(ctx context.Context, req *BatchVerificationRequest) (*BatchVerificationResponse, error)

	// GetVerification retrieves a previous verification result by ID
	GetVerification(ctx context.Context, requestID string) (*VerificationResponse, error)

	// ListVerifications lists verifications for a wallet
	ListVerifications(ctx context.Context, walletAddress string, opts ListOptions) ([]VerificationResponse, error)

	// Consent methods

	// GrantConsent records user consent for verification
	GrantConsent(ctx context.Context, consent *Consent) (*Consent, error)

	// RevokeConsent revokes user consent
	RevokeConsent(ctx context.Context, consentID string) error

	// GetConsent retrieves consent by ID
	GetConsent(ctx context.Context, consentID string) (*Consent, error)

	// ListConsents lists consents for a wallet
	ListConsents(ctx context.Context, walletAddress string) ([]Consent, error)

	// ValidateConsent validates consent for a verification request
	ValidateConsent(ctx context.Context, consentID string, req *VerificationRequest) error

	// Jurisdiction methods

	// GetJurisdiction returns a jurisdiction by code
	GetJurisdiction(ctx context.Context, code string) (*Jurisdiction, error)

	// ListJurisdictions lists all supported jurisdictions
	ListJurisdictions(ctx context.Context) ([]Jurisdiction, error)

	// IsJurisdictionSupported checks if a jurisdiction is supported
	IsJurisdictionSupported(ctx context.Context, code string) bool

	// GetSupportedDocuments returns supported documents for a jurisdiction
	GetSupportedDocuments(ctx context.Context, jurisdiction string) ([]DocumentType, error)

	// Audit methods

	// GetAuditLog retrieves an audit log entry
	GetAuditLog(ctx context.Context, auditID string) (*AuditLogEntry, error)

	// ListAuditLogs lists audit logs with filtering
	ListAuditLogs(ctx context.Context, filter AuditLogFilter) ([]AuditLogEntry, error)

	// ExportAuditLogs exports audit logs
	ExportAuditLogs(ctx context.Context, filter AuditLogFilter) ([]byte, error)

	// VEID integration methods

	// CreateVEIDScope creates a VEID scope from verification
	CreateVEIDScope(ctx context.Context, verification *VerificationResponse) (*VEIDScope, error)

	// EnrichVEID enriches an existing VEID with government verification
	EnrichVEID(ctx context.Context, verification *VerificationResponse, veidID string) error

	// GetVEIDScopes retrieves VEID scopes for a wallet
	GetVEIDScopes(ctx context.Context, walletAddress string) ([]VEIDScope, error)
}

// ListOptions contains options for listing
type ListOptions struct {
	// Offset is the offset for pagination
	Offset int `json:"offset"`

	// Limit is the maximum results to return
	Limit int `json:"limit"`

	// SortBy is the field to sort by
	SortBy string `json:"sort_by"`

	// SortDesc sorts in descending order
	SortDesc bool `json:"sort_desc"`

	// Status filters by status
	Status *VerificationStatus `json:"status,omitempty"`

	// Since filters by time
	Since *time.Time `json:"since,omitempty"`

	// Until filters by time
	Until *time.Time `json:"until,omitempty"`
}

// AuditLogFilter contains audit log filtering options
type AuditLogFilter struct {
	// WalletAddress filters by wallet address
	WalletAddress string `json:"wallet_address,omitempty"`

	// Jurisdiction filters by jurisdiction
	Jurisdiction string `json:"jurisdiction,omitempty"`

	// Action filters by action
	Action *AuditAction `json:"action,omitempty"`

	// Status filters by status
	Status *VerificationStatus `json:"status,omitempty"`

	// Since filters by time
	Since time.Time `json:"since,omitempty"`

	// Until filters by time
	Until time.Time `json:"until,omitempty"`

	// Offset for pagination
	Offset int `json:"offset"`

	// Limit for pagination
	Limit int `json:"limit"`
}

// ============================================================================
// Data Source Adapter Interface
// ============================================================================

// DataSourceAdapter defines the interface for government data source adapters
type DataSourceAdapter interface {
	// Type returns the adapter type
	Type() DataSourceType

	// Jurisdiction returns the jurisdiction served
	Jurisdiction() string

	// SupportedDocuments returns supported document types
	SupportedDocuments() []DocumentType

	// SupportsDocument checks if a document type is supported
	SupportsDocument(docType DocumentType) bool

	// IsAvailable checks if the adapter is available
	IsAvailable(ctx context.Context) bool

	// HealthCheck performs a health check
	HealthCheck(ctx context.Context) error

	// Verify performs the verification
	Verify(ctx context.Context, req *VerificationRequest) (*VerificationResponse, error)

	// GetLastError returns the last error
	GetLastError() error

	// GetStats returns adapter statistics
	GetStats() AdapterStatus
}

// ============================================================================
// Consent Manager Interface
// ============================================================================

// ConsentManager manages user consent for government data access
type ConsentManager interface {
	// Grant grants consent
	Grant(ctx context.Context, consent *Consent) (*Consent, error)

	// Revoke revokes consent
	Revoke(ctx context.Context, consentID string) error

	// Get retrieves consent by ID
	Get(ctx context.Context, consentID string) (*Consent, error)

	// List lists consents for a wallet
	List(ctx context.Context, walletAddress string) ([]Consent, error)

	// Validate validates consent for a request
	Validate(ctx context.Context, consentID string, req *VerificationRequest) error

	// IsConsentValid checks if consent is valid
	IsConsentValid(ctx context.Context, consentID string) bool

	// CleanupExpired removes expired consent records
	CleanupExpired(ctx context.Context) (int, error)
}

// ============================================================================
// Audit Logger Interface
// ============================================================================

// AuditLogger provides audit logging for government data access
type AuditLogger interface {
	// Log logs an audit entry
	Log(ctx context.Context, entry *AuditLogEntry) error

	// Get retrieves an audit log entry
	Get(ctx context.Context, auditID string) (*AuditLogEntry, error)

	// List lists audit log entries
	List(ctx context.Context, filter AuditLogFilter) ([]AuditLogEntry, error)

	// Export exports audit logs
	Export(ctx context.Context, filter AuditLogFilter, format string) ([]byte, error)

	// Purge purges old audit logs
	Purge(ctx context.Context, before time.Time) (int, error)

	// Count counts audit log entries
	Count(ctx context.Context, filter AuditLogFilter) (int64, error)
}

// ============================================================================
// Verification Result Store Interface
// ============================================================================

// VerificationStore stores verification results
type VerificationStore interface {
	// Store stores a verification result
	Store(ctx context.Context, result *VerificationResponse) error

	// Get retrieves a verification by request ID
	Get(ctx context.Context, requestID string) (*VerificationResponse, error)

	// List lists verifications for a wallet
	List(ctx context.Context, walletAddress string, opts ListOptions) ([]VerificationResponse, error)

	// Delete deletes a verification result
	Delete(ctx context.Context, requestID string) error

	// Purge purges expired verification results
	Purge(ctx context.Context, before time.Time) (int, error)
}

// ============================================================================
// VEID Integrator Interface
// ============================================================================

// VEIDIntegrator integrates government verification with VEID
type VEIDIntegrator interface {
	// CreateScope creates a VEID scope from verification
	CreateScope(ctx context.Context, verification *VerificationResponse) (*VEIDScope, error)

	// EnrichIdentity enriches an existing VEID
	EnrichIdentity(ctx context.Context, verification *VerificationResponse, veidID string) error

	// GetScopes retrieves VEID scopes for a wallet
	GetScopes(ctx context.Context, walletAddress string) ([]VEIDScope, error)

	// RevokeScope revokes a VEID scope
	RevokeScope(ctx context.Context, scopeID string) error

	// ComputeScoreContribution computes the VEID score contribution
	ComputeScoreContribution(verification *VerificationResponse) float64

	// GetScopeStats returns scope statistics
	GetScopeStats(ctx context.Context, walletAddress string) (*VEIDScopeStats, error)
}

// VEIDScopeStats contains VEID scope statistics
type VEIDScopeStats struct {
	// TotalScopes is the total number of scopes
	TotalScopes int `json:"total_scopes"`

	// ActiveScopes is the number of active scopes
	ActiveScopes int `json:"active_scopes"`

	// ExpiredScopes is the number of expired scopes
	ExpiredScopes int `json:"expired_scopes"`

	// TotalScoreContribution is the total score contribution
	TotalScoreContribution float64 `json:"total_score_contribution"`

	// DocumentTypesVerified lists verified document types
	DocumentTypesVerified []DocumentType `json:"document_types_verified"`

	// JurisdictionsVerified lists verified jurisdictions
	JurisdictionsVerified []string `json:"jurisdictions_verified"`

	// LastVerificationAt is the last verification time
	LastVerificationAt *time.Time `json:"last_verification_at,omitempty"`
}

// ============================================================================
// Jurisdiction Registry Interface
// ============================================================================

// JurisdictionRegistry manages supported jurisdictions
type JurisdictionRegistry interface {
	// Get retrieves a jurisdiction by code
	Get(ctx context.Context, code string) (*Jurisdiction, error)

	// List lists all jurisdictions
	List(ctx context.Context) ([]Jurisdiction, error)

	// IsSupported checks if a jurisdiction is supported
	IsSupported(ctx context.Context, code string) bool

	// Register registers a new jurisdiction
	Register(ctx context.Context, jurisdiction *Jurisdiction) error

	// Update updates a jurisdiction
	Update(ctx context.Context, jurisdiction *Jurisdiction) error

	// Disable disables a jurisdiction
	Disable(ctx context.Context, code string) error

	// GetSupportedDocuments returns supported documents for jurisdiction
	GetSupportedDocuments(ctx context.Context, code string) ([]DocumentType, error)

	// GetDataSources returns data sources for jurisdiction
	GetDataSources(ctx context.Context, code string) ([]DataSourceType, error)
}

// ============================================================================
// Rate Limiter Interface
// ============================================================================

// RateLimiter provides rate limiting for verification requests
type RateLimiter interface {
	// Allow checks if a request is allowed
	Allow(ctx context.Context, walletAddress string) (bool, error)

	// GetRemaining returns remaining requests
	GetRemaining(ctx context.Context, walletAddress string) (RateLimitInfo, error)

	// Reset resets rate limits for a wallet
	Reset(ctx context.Context, walletAddress string) error
}

// RateLimitInfo contains rate limit information
type RateLimitInfo struct {
	// RemainingMinute is requests remaining this minute
	RemainingMinute int `json:"remaining_minute"`

	// RemainingHour is requests remaining this hour
	RemainingHour int `json:"remaining_hour"`

	// RemainingDay is requests remaining today
	RemainingDay int `json:"remaining_day"`

	// ResetsAt is when limits reset
	ResetsAt time.Time `json:"resets_at"`
}

// ============================================================================
// Adapter Factory Interface
// ============================================================================

// AdapterFactory creates data source adapters
type AdapterFactory interface {
	// Create creates an adapter for the given configuration
	Create(config AdapterConfig) (DataSourceAdapter, error)

	// CreateForJurisdiction creates adapters for a jurisdiction
	CreateForJurisdiction(jurisdiction string, configs []AdapterConfig) ([]DataSourceAdapter, error)
}
