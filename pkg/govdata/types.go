// Package govdata provides government data source integration for identity verification.
//
// VE-909: Government data integration for authoritative identity verification
package govdata

import (
	"errors"
	"time"
)

// ============================================================================
// Errors
// ============================================================================

var (
	// ErrServiceNotInitialized is returned when the service is not initialized
	ErrServiceNotInitialized = errors.New("govdata service not initialized")

	// ErrJurisdictionNotSupported is returned when jurisdiction is not configured
	ErrJurisdictionNotSupported = errors.New("jurisdiction not supported")

	// ErrDocumentTypeNotSupported is returned when document type is not supported
	ErrDocumentTypeNotSupported = errors.New("document type not supported for jurisdiction")

	// ErrAdapterNotFound is returned when no adapter is found for the request
	ErrAdapterNotFound = errors.New("no adapter found for request")

	// ErrAdapterUnavailable is returned when adapter is temporarily unavailable
	ErrAdapterUnavailable = errors.New("data source adapter unavailable")

	// ErrVerificationFailed is returned when verification fails
	ErrVerificationFailed = errors.New("document verification failed")

	// ErrVerificationTimeout is returned when verification times out
	ErrVerificationTimeout = errors.New("verification request timed out")

	// ErrInvalidRequest is returned when request is invalid
	ErrInvalidRequest = errors.New("invalid verification request")

	// ErrMissingRequiredField is returned when required field is missing
	ErrMissingRequiredField = errors.New("required verification field missing")

	// ErrInvalidDocumentNumber is returned when document number format is invalid
	ErrInvalidDocumentNumber = errors.New("invalid document number format")

	// ErrConsentRequired is returned when user consent is required
	ErrConsentRequired = errors.New("user consent required for verification")

	// ErrConsentExpired is returned when user consent has expired
	ErrConsentExpired = errors.New("user consent has expired")

	// ErrRateLimitExceeded is returned when rate limit is exceeded
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrUnauthorized is returned when requester is not authorized
	ErrUnauthorized = errors.New("unauthorized verification request")

	// ErrAuditLogFailed is returned when audit logging fails
	ErrAuditLogFailed = errors.New("failed to create audit log entry")

	// ErrDataSourceError is returned when government data source returns error
	ErrDataSourceError = errors.New("government data source error")

	// ErrDataRetentionViolation is returned when retention policy is violated
	ErrDataRetentionViolation = errors.New("data retention policy violation")

	// ErrVEIDIntegrationFailed is returned when VEID integration fails
	ErrVEIDIntegrationFailed = errors.New("VEID integration failed")

	// ErrDocumentExpired is returned when document has expired
	ErrDocumentExpired = errors.New("document has expired")

	// ErrDocumentRevoked is returned when document has been revoked
	ErrDocumentRevoked = errors.New("document has been revoked")
)

// ============================================================================
// Constants
// ============================================================================

// GovDataVersion is the current version of the govdata package
const GovDataVersion = "1.0.0"

// ============================================================================
// Document Types
// ============================================================================

// DocumentType represents the type of government document
type DocumentType string

const (
	// DocumentTypeDriversLicense is a driver's license
	DocumentTypeDriversLicense DocumentType = "drivers_license"

	// DocumentTypeStateID is a state-issued ID card
	DocumentTypeStateID DocumentType = "state_id"

	// DocumentTypePassport is a passport
	DocumentTypePassport DocumentType = "passport"

	// DocumentTypeBirthCertificate is a birth certificate
	DocumentTypeBirthCertificate DocumentType = "birth_certificate"

	// DocumentTypeNationalID is a national identity card
	DocumentTypeNationalID DocumentType = "national_id"

	// DocumentTypeTaxID is a tax identification number
	DocumentTypeTaxID DocumentType = "tax_id"

	// DocumentTypeVoterID is a voter registration ID
	DocumentTypeVoterID DocumentType = "voter_id"

	// DocumentTypeMilitaryID is a military ID
	DocumentTypeMilitaryID DocumentType = "military_id"

	// DocumentTypeResidencePermit is a residence permit
	DocumentTypeResidencePermit DocumentType = "residence_permit"

	// DocumentTypeVisaDocument is a visa document
	DocumentTypeVisaDocument DocumentType = "visa_document"
)

// IsValid checks if the document type is valid
func (dt DocumentType) IsValid() bool {
	switch dt {
	case DocumentTypeDriversLicense, DocumentTypeStateID, DocumentTypePassport,
		DocumentTypeBirthCertificate, DocumentTypeNationalID, DocumentTypeTaxID,
		DocumentTypeVoterID, DocumentTypeMilitaryID, DocumentTypeResidencePermit,
		DocumentTypeVisaDocument:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (dt DocumentType) String() string {
	return string(dt)
}

// ============================================================================
// Data Source Types
// ============================================================================

// DataSourceType represents the type of government data source
type DataSourceType string

const (
	// DataSourceDMV is a Department of Motor Vehicles
	DataSourceDMV DataSourceType = "dmv"

	// DataSourcePassport is a passport authority
	DataSourcePassport DataSourceType = "passport"

	// DataSourceVitalRecords is a vital records office
	DataSourceVitalRecords DataSourceType = "vital_records"

	// DataSourceNationalRegistry is a national registry
	DataSourceNationalRegistry DataSourceType = "national_registry"

	// DataSourceTaxAuthority is a tax authority
	DataSourceTaxAuthority DataSourceType = "tax_authority"

	// DataSourceImmigration is an immigration authority
	DataSourceImmigration DataSourceType = "immigration"

	// DataSourceDVS is Australia's Document Verification Service
	DataSourceDVS DataSourceType = "dvs"

	// DataSourceGovUK is UK's GOV.UK Verify service
	DataSourceGovUK DataSourceType = "govuk_verify"

	// DataSourceEIDAS is EU's eIDAS framework
	DataSourceEIDAS DataSourceType = "eidas"

	// DataSourcePCTF is Canada's Pan-Canadian Trust Framework
	DataSourcePCTF DataSourceType = "pctf"
)

// IsValid checks if the data source type is valid
func (dst DataSourceType) IsValid() bool {
	switch dst {
	case DataSourceDMV, DataSourcePassport, DataSourceVitalRecords,
		DataSourceNationalRegistry, DataSourceTaxAuthority, DataSourceImmigration,
		DataSourceDVS, DataSourceGovUK, DataSourceEIDAS, DataSourcePCTF:
		return true
	default:
		return false
	}
}

// ============================================================================
// Verification Status
// ============================================================================

// VerificationStatus represents the status of a verification request
type VerificationStatus string

const (
	// VerificationStatusPending is a pending verification
	VerificationStatusPending VerificationStatus = "pending"

	// VerificationStatusVerified is a successful verification
	VerificationStatusVerified VerificationStatus = "verified"

	// VerificationStatusNotVerified is a failed verification
	VerificationStatusNotVerified VerificationStatus = "not_verified"

	// VerificationStatusPartialMatch is a partial match
	VerificationStatusPartialMatch VerificationStatus = "partial_match"

	// VerificationStatusNotFound is when document/person not found
	VerificationStatusNotFound VerificationStatus = "not_found"

	// VerificationStatusExpired is when document has expired
	VerificationStatusExpired VerificationStatus = "expired"

	// VerificationStatusRevoked is when document has been revoked
	VerificationStatusRevoked VerificationStatus = "revoked"

	// VerificationStatusError is when verification error occurred
	VerificationStatusError VerificationStatus = "error"

	// VerificationStatusTimeout is when verification timed out
	VerificationStatusTimeout VerificationStatus = "timeout"

	// VerificationStatusRateLimited is when rate limit was exceeded
	VerificationStatusRateLimited VerificationStatus = "rate_limited"
)

// IsSuccess returns true if verification was successful
func (vs VerificationStatus) IsSuccess() bool {
	return vs == VerificationStatusVerified || vs == VerificationStatusPartialMatch
}

// ============================================================================
// Field Match Result
// ============================================================================

// FieldMatchResult represents the result of matching a single field
type FieldMatchResult string

const (
	// FieldMatchExact is an exact match
	FieldMatchExact FieldMatchResult = "exact"

	// FieldMatchFuzzy is a fuzzy/partial match
	FieldMatchFuzzy FieldMatchResult = "fuzzy"

	// FieldMatchNoMatch is no match
	FieldMatchNoMatch FieldMatchResult = "no_match"

	// FieldMatchNotChecked is when field was not checked
	FieldMatchNotChecked FieldMatchResult = "not_checked"

	// FieldMatchUnavailable is when field data is unavailable
	FieldMatchUnavailable FieldMatchResult = "unavailable"
)

// ============================================================================
// Verification Request
// ============================================================================

// VerificationRequest represents a request to verify government data
type VerificationRequest struct {
	// RequestID is the unique request identifier
	RequestID string `json:"request_id"`

	// WalletAddress is the requester's wallet address
	WalletAddress string `json:"wallet_address"`

	// DocumentType is the type of document to verify
	DocumentType DocumentType `json:"document_type"`

	// Jurisdiction is the jurisdiction code (ISO 3166-1 alpha-2 or alpha-2 + subdivision)
	Jurisdiction string `json:"jurisdiction"`

	// Fields contains the fields to verify
	Fields VerificationFields `json:"fields"`

	// ConsentID is the ID of the user's consent record
	ConsentID string `json:"consent_id"`

	// VerifyOnly specifies which fields to verify (empty = all)
	VerifyOnly []string `json:"verify_only,omitempty"`

	// Metadata contains additional request metadata
	Metadata map[string]string `json:"metadata,omitempty"`

	// RequestedAt is when the request was created
	RequestedAt time.Time `json:"requested_at"`

	// ExpiresAt is when the request expires
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// Validate validates the verification request
func (r *VerificationRequest) Validate() error {
	if r.WalletAddress == "" {
		return ErrMissingRequiredField
	}
	if !r.DocumentType.IsValid() {
		return ErrDocumentTypeNotSupported
	}
	if r.Jurisdiction == "" {
		return ErrMissingRequiredField
	}
	if r.Fields.DocumentNumber == "" {
		return ErrMissingRequiredField
	}
	return nil
}

// VerificationFields contains the fields to verify
type VerificationFields struct {
	// DocumentNumber is the document number
	DocumentNumber string `json:"document_number"`

	// FirstName is the first name
	FirstName string `json:"first_name,omitempty"`

	// MiddleName is the middle name
	MiddleName string `json:"middle_name,omitempty"`

	// LastName is the last name
	LastName string `json:"last_name"`

	// DateOfBirth is the date of birth
	DateOfBirth time.Time `json:"date_of_birth,omitempty"`

	// Gender is the gender
	Gender string `json:"gender,omitempty"`

	// Address contains address fields
	Address *AddressFields `json:"address,omitempty"`

	// IssueDate is when the document was issued
	IssueDate time.Time `json:"issue_date,omitempty"`

	// ExpirationDate is when the document expires
	ExpirationDate time.Time `json:"expiration_date,omitempty"`

	// IssuingAuthority is the issuing authority
	IssuingAuthority string `json:"issuing_authority,omitempty"`

	// DocumentClass is the document class (e.g., license class)
	DocumentClass string `json:"document_class,omitempty"`

	// Nationality is the nationality
	Nationality string `json:"nationality,omitempty"`

	// PlaceOfBirth is the place of birth
	PlaceOfBirth string `json:"place_of_birth,omitempty"`
}

// AddressFields contains address verification fields
type AddressFields struct {
	// Street is the street address
	Street string `json:"street,omitempty"`

	// City is the city
	City string `json:"city,omitempty"`

	// State is the state/province
	State string `json:"state,omitempty"`

	// PostalCode is the postal code
	PostalCode string `json:"postal_code,omitempty"`

	// Country is the country
	Country string `json:"country,omitempty"`
}

// ============================================================================
// Verification Response
// ============================================================================

// VerificationResponse represents the result of a verification request
type VerificationResponse struct {
	// RequestID is the original request ID
	RequestID string `json:"request_id"`

	// Status is the verification status
	Status VerificationStatus `json:"status"`

	// Confidence is the overall confidence score (0.0 - 1.0)
	Confidence float64 `json:"confidence"`

	// FieldResults contains per-field match results
	FieldResults map[string]FieldVerificationResult `json:"field_results"`

	// DataSourceType is the type of data source used
	DataSourceType DataSourceType `json:"data_source_type"`

	// Jurisdiction is the jurisdiction that performed verification
	Jurisdiction string `json:"jurisdiction"`

	// DocumentValid indicates if document is currently valid
	DocumentValid bool `json:"document_valid"`

	// DocumentExpiresAt is when the document expires (if available)
	DocumentExpiresAt *time.Time `json:"document_expires_at,omitempty"`

	// VerifiedAt is when verification was completed
	VerifiedAt time.Time `json:"verified_at"`

	// ExpiresAt is when this verification result expires
	ExpiresAt time.Time `json:"expires_at"`

	// AuditLogID is the ID of the audit log entry
	AuditLogID string `json:"audit_log_id"`

	// ErrorCode is the error code if status is error
	ErrorCode string `json:"error_code,omitempty"`

	// ErrorMessage is the error message if status is error
	ErrorMessage string `json:"error_message,omitempty"`

	// Warnings contains any verification warnings
	Warnings []string `json:"warnings,omitempty"`
}

// FieldVerificationResult contains the result for a single field
type FieldVerificationResult struct {
	// FieldName is the name of the field
	FieldName string `json:"field_name"`

	// Match is the match result
	Match FieldMatchResult `json:"match"`

	// Confidence is the confidence for this field (0.0 - 1.0)
	Confidence float64 `json:"confidence"`

	// Note contains any notes about the verification
	Note string `json:"note,omitempty"`
}

// ============================================================================
// Consent
// ============================================================================

// Consent represents user consent for government data verification
type Consent struct {
	// ID is the unique consent ID
	ID string `json:"id"`

	// WalletAddress is the user's wallet address
	WalletAddress string `json:"wallet_address"`

	// DocumentTypes lists consented document types
	DocumentTypes []DocumentType `json:"document_types"`

	// Jurisdictions lists consented jurisdictions
	Jurisdictions []string `json:"jurisdictions"`

	// Purpose describes the verification purpose
	Purpose string `json:"purpose"`

	// GrantedAt is when consent was granted
	GrantedAt time.Time `json:"granted_at"`

	// ExpiresAt is when consent expires
	ExpiresAt time.Time `json:"expires_at"`

	// RevokedAt is when consent was revoked (if revoked)
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	// Signature is the cryptographic signature of consent
	Signature string `json:"signature"`

	// Active indicates if consent is currently active
	Active bool `json:"active"`
}

// IsValid checks if consent is currently valid
func (c *Consent) IsValid() bool {
	if !c.Active || c.RevokedAt != nil {
		return false
	}
	if time.Now().After(c.ExpiresAt) {
		return false
	}
	return true
}

// ============================================================================
// Jurisdiction
// ============================================================================

// Jurisdiction represents a supported jurisdiction
type Jurisdiction struct {
	// Code is the jurisdiction code (ISO 3166-1 alpha-2 or with subdivision)
	Code string `json:"code"`

	// Name is the jurisdiction name
	Name string `json:"name"`

	// Country is the country code
	Country string `json:"country"`

	// Subdivision is the subdivision code (state, province, etc.)
	Subdivision string `json:"subdivision,omitempty"`

	// SupportedDocuments lists supported document types
	SupportedDocuments []DocumentType `json:"supported_documents"`

	// DataSources lists available data sources
	DataSources []DataSourceType `json:"data_sources"`

	// RetentionPolicy specifies data retention rules
	RetentionPolicy RetentionPolicy `json:"retention_policy"`

	// GDPRApplicable indicates if GDPR applies
	GDPRApplicable bool `json:"gdpr_applicable"`

	// CCPAApplicable indicates if CCPA applies
	CCPAApplicable bool `json:"ccpa_applicable"`

	// Active indicates if jurisdiction is active
	Active bool `json:"active"`

	// RequiresConsent indicates if explicit consent is required
	RequiresConsent bool `json:"requires_consent"`
}

// RetentionPolicy specifies data retention rules
type RetentionPolicy struct {
	// ResultRetentionDays is how long to keep verification results
	ResultRetentionDays int `json:"result_retention_days"`

	// AuditLogRetentionDays is how long to keep audit logs
	AuditLogRetentionDays int `json:"audit_log_retention_days"`

	// ConsentRetentionDays is how long to keep consent records
	ConsentRetentionDays int `json:"consent_retention_days"`

	// AutoPurge indicates if data should be auto-purged
	AutoPurge bool `json:"auto_purge"`
}

// ============================================================================
// Audit Log
// ============================================================================

// AuditLogEntry represents an audit log entry for government data access
type AuditLogEntry struct {
	// ID is the unique audit log entry ID
	ID string `json:"id"`

	// RequestID is the verification request ID
	RequestID string `json:"request_id"`

	// Action is the action performed
	Action AuditAction `json:"action"`

	// WalletAddress is the requesting wallet address
	WalletAddress string `json:"wallet_address"`

	// Jurisdiction is the jurisdiction accessed
	Jurisdiction string `json:"jurisdiction"`

	// DocumentType is the document type accessed
	DocumentType DocumentType `json:"document_type"`

	// DataSource is the data source accessed
	DataSource DataSourceType `json:"data_source"`

	// Status is the request status
	Status VerificationStatus `json:"status"`

	// ConsentID is the associated consent ID
	ConsentID string `json:"consent_id"`

	// IPAddress is the requester's IP address (hashed)
	IPAddress string `json:"ip_address_hash"`

	// Timestamp is when the access occurred
	Timestamp time.Time `json:"timestamp"`

	// Duration is how long the request took
	Duration time.Duration `json:"duration"`

	// ErrorCode is the error code if failed
	ErrorCode string `json:"error_code,omitempty"`

	// Metadata contains additional audit metadata
	Metadata map[string]string `json:"metadata,omitempty"`

	// RetentionExpiresAt is when this log entry should be purged
	RetentionExpiresAt time.Time `json:"retention_expires_at"`
}

// AuditAction represents the type of audit action
type AuditAction string

const (
	// AuditActionVerify is a verification request
	AuditActionVerify AuditAction = "verify"

	// AuditActionConsentGrant is consent granted
	AuditActionConsentGrant AuditAction = "consent_grant"

	// AuditActionConsentRevoke is consent revoked
	AuditActionConsentRevoke AuditAction = "consent_revoke"

	// AuditActionDataAccess is data access
	AuditActionDataAccess AuditAction = "data_access"

	// AuditActionDataPurge is data purge
	AuditActionDataPurge AuditAction = "data_purge"

	// AuditActionAdminOverride is admin override
	AuditActionAdminOverride AuditAction = "admin_override"
)

// ============================================================================
// VEID Integration Types
// ============================================================================

// VEIDScope represents a VEID scope from government verification
type VEIDScope struct {
	// ID is the unique scope ID
	ID string `json:"id"`

	// WalletAddress is the associated wallet address
	WalletAddress string `json:"wallet_address"`

	// DocumentType is the verified document type
	DocumentType DocumentType `json:"document_type"`

	// Jurisdiction is the verification jurisdiction
	Jurisdiction string `json:"jurisdiction"`

	// DataSource is the data source used
	DataSource DataSourceType `json:"data_source"`

	// VerificationStatus is the verification status
	VerificationStatus VerificationStatus `json:"verification_status"`

	// Confidence is the verification confidence
	Confidence float64 `json:"confidence"`

	// ScoreContribution is the contribution to VEID score
	ScoreContribution float64 `json:"score_contribution"`

	// FieldsVerified lists which fields were verified
	FieldsVerified []string `json:"fields_verified"`

	// VerifiedAt is when verification was performed
	VerifiedAt time.Time `json:"verified_at"`

	// ExpiresAt is when this scope expires
	ExpiresAt time.Time `json:"expires_at"`

	// CreatedAt is when scope was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when scope was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// Status is the scope status
	Status string `json:"status"`
}

// ============================================================================
// Service Status
// ============================================================================

// ServiceStatus represents the service status
type ServiceStatus struct {
	// Version is the service version
	Version string `json:"version"`

	// Healthy indicates if service is healthy
	Healthy bool `json:"healthy"`

	// StartedAt is when service started
	StartedAt time.Time `json:"started_at"`

	// Uptime is the service uptime
	Uptime time.Duration `json:"uptime"`

	// ActiveAdapters is the number of active adapters
	ActiveAdapters int `json:"active_adapters"`

	// TotalVerifications is the total verifications performed
	TotalVerifications int64 `json:"total_verifications"`

	// SuccessfulVerifications is the successful verifications
	SuccessfulVerifications int64 `json:"successful_verifications"`

	// FailedVerifications is the failed verifications
	FailedVerifications int64 `json:"failed_verifications"`

	// AdapterStatus contains per-adapter status
	AdapterStatus map[string]AdapterStatus `json:"adapter_status"`

	// LastError is the last error (if any)
	LastError string `json:"last_error,omitempty"`
}

// AdapterStatus represents an adapter's status
type AdapterStatus struct {
	// Type is the adapter type
	Type DataSourceType `json:"type"`

	// Jurisdiction is the jurisdiction served
	Jurisdiction string `json:"jurisdiction"`

	// Available indicates if adapter is available
	Available bool `json:"available"`

	// LastCheck is the last health check time
	LastCheck time.Time `json:"last_check"`

	// LastSuccess is the last successful verification
	LastSuccess *time.Time `json:"last_success,omitempty"`

	// ErrorCount is the error count since last success
	ErrorCount int `json:"error_count"`

	// AverageLatency is the average request latency
	AverageLatency time.Duration `json:"average_latency"`
}

// ============================================================================
// Batch Verification
// ============================================================================

// BatchVerificationRequest represents a batch verification request
type BatchVerificationRequest struct {
	// BatchID is the unique batch ID
	BatchID string `json:"batch_id"`

	// Requests contains the individual requests
	Requests []VerificationRequest `json:"requests"`

	// FailFast stops on first failure if true
	FailFast bool `json:"fail_fast"`

	// MaxConcurrent is the maximum concurrent requests
	MaxConcurrent int `json:"max_concurrent"`

	// Timeout is the overall batch timeout
	Timeout time.Duration `json:"timeout"`
}

// BatchVerificationResponse represents a batch verification response
type BatchVerificationResponse struct {
	// BatchID is the batch ID
	BatchID string `json:"batch_id"`

	// Responses contains individual responses
	Responses []VerificationResponse `json:"responses"`

	// TotalRequests is the total request count
	TotalRequests int `json:"total_requests"`

	// SuccessCount is the successful verification count
	SuccessCount int `json:"success_count"`

	// FailureCount is the failed verification count
	FailureCount int `json:"failure_count"`

	// Duration is the total batch duration
	Duration time.Duration `json:"duration"`
}
