// Package govdata provides government data source integration for identity verification.
//
// SECURITY-004: Privacy and PII protection utilities for GDPR/CCPA compliance
package govdata

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// Privacy Errors
// ============================================================================

var (
	// ErrDataNotFound is returned when data is not found
	ErrDataNotFound = errors.New("data not found")

	// ErrErasureNotAllowed is returned when erasure is not allowed
	ErrErasureNotAllowed = errors.New("erasure not allowed due to legal hold")

	// ErrConsentNotGranted is returned when consent is not granted
	ErrConsentNotGranted = errors.New("consent not granted for data processing")

	// ErrDataMinimizationViolation is returned when data minimization is violated
	ErrDataMinimizationViolation = errors.New("data minimization principle violated")
)

// ============================================================================
// Privacy Configuration
// ============================================================================

// PrivacyConfig contains privacy and data protection configuration
type PrivacyConfig struct {
	// Enabled indicates if privacy protection is enabled
	Enabled bool `json:"enabled" yaml:"enabled"`

	// GDPRMode enables GDPR-specific handling
	GDPRMode bool `json:"gdpr_mode" yaml:"gdpr_mode"`

	// CCPAMode enables CCPA-specific handling
	CCPAMode bool `json:"ccpa_mode" yaml:"ccpa_mode"`

	// DataMinimization enables data minimization enforcement
	DataMinimization bool `json:"data_minimization" yaml:"data_minimization"`

	// AutoPurge enables automatic data purge after retention period
	AutoPurge bool `json:"auto_purge" yaml:"auto_purge"`

	// PurgeIntervalHours is hours between purge runs
	PurgeIntervalHours int `json:"purge_interval_hours" yaml:"purge_interval_hours"`

	// PIIEncryptionRequired requires PII to be encrypted
	PIIEncryptionRequired bool `json:"pii_encryption_required" yaml:"pii_encryption_required"`

	// AuditDataAccess audits all data access
	AuditDataAccess bool `json:"audit_data_access" yaml:"audit_data_access"`

	// SubjectAccessRequestEnabled enables GDPR Article 15 SAR
	SubjectAccessRequestEnabled bool `json:"subject_access_request_enabled" yaml:"subject_access_request_enabled"`

	// RightToErasureEnabled enables GDPR Article 17 erasure
	RightToErasureEnabled bool `json:"right_to_erasure_enabled" yaml:"right_to_erasure_enabled"`

	// DataPortabilityEnabled enables GDPR Article 20 portability
	DataPortabilityEnabled bool `json:"data_portability_enabled" yaml:"data_portability_enabled"`

	// LegalHoldDays is days data must be kept for legal reasons
	LegalHoldDays int `json:"legal_hold_days" yaml:"legal_hold_days"`
}

// DefaultPrivacyConfig returns default privacy configuration
func DefaultPrivacyConfig() PrivacyConfig {
	return PrivacyConfig{
		Enabled:                     true,
		GDPRMode:                    true,
		CCPAMode:                    true,
		DataMinimization:            true,
		AutoPurge:                   true,
		PurgeIntervalHours:          24,
		PIIEncryptionRequired:       true,
		AuditDataAccess:             true,
		SubjectAccessRequestEnabled: true,
		RightToErasureEnabled:       true,
		DataPortabilityEnabled:      true,
		LegalHoldDays:               365 * 7, // 7 years for financial/identity records
	}
}

// ============================================================================
// PII Classification
// ============================================================================

// PIICategory represents categories of personally identifiable information
type PIICategory string

const (
	// PIICategoryBasic is basic identity info (name, DOB)
	PIICategoryBasic PIICategory = "basic_identity"

	// PIICategoryContact is contact information
	PIICategoryContact PIICategory = "contact"

	// PIICategoryDocument is document identifiers
	PIICategoryDocument PIICategory = "document"

	// PIICategoryBiometric is biometric data
	PIICategoryBiometric PIICategory = "biometric"

	// PIICategoryFinancial is financial information
	PIICategoryFinancial PIICategory = "financial"

	// PIICategorySensitive is sensitive personal data (GDPR Article 9)
	PIICategorySensitive PIICategory = "sensitive"

	// PIICategoryGenetic is genetic data
	PIICategoryGenetic PIICategory = "genetic"

	// PIICategoryHealth is health data
	PIICategoryHealth PIICategory = "health"
)

// PIIField represents a field containing PII
type PIIField struct {
	// Name is the field name
	Name string `json:"name"`

	// Category is the PII category
	Category PIICategory `json:"category"`

	// Value is the field value (may be hashed/encrypted)
	Value string `json:"value"`

	// IsHashed indicates if value is hashed
	IsHashed bool `json:"is_hashed"`

	// IsEncrypted indicates if value is encrypted
	IsEncrypted bool `json:"is_encrypted"`

	// RetentionExpiresAt is when the field should be purged
	RetentionExpiresAt time.Time `json:"retention_expires_at"`
}

// ============================================================================
// Data Subject Rights (GDPR Articles 15-22)
// ============================================================================

// SubjectAccessRequest represents a GDPR Article 15 subject access request
type SubjectAccessRequest struct {
	// ID is the request ID
	ID string `json:"id"`

	// SubjectID is the data subject identifier (wallet address)
	SubjectID string `json:"subject_id"`

	// RequestedAt is when the request was made
	RequestedAt time.Time `json:"requested_at"`

	// ResponseDeadline is the deadline for response (30 days per GDPR)
	ResponseDeadline time.Time `json:"response_deadline"`

	// Status is the request status
	Status string `json:"status"`

	// DataCategories are the categories of data requested
	DataCategories []PIICategory `json:"data_categories"`

	// VerificationProof is proof of identity verification
	VerificationProof string `json:"verification_proof,omitempty"`
}

// ErasureRequest represents a GDPR Article 17 right to erasure request
type ErasureRequest struct {
	// ID is the request ID
	ID string `json:"id"`

	// SubjectID is the data subject identifier
	SubjectID string `json:"subject_id"`

	// RequestedAt is when the request was made
	RequestedAt time.Time `json:"requested_at"`

	// Reason is the reason for erasure
	Reason ErasureReason `json:"reason"`

	// DataCategories are the categories to erase
	DataCategories []PIICategory `json:"data_categories"`

	// Status is the request status
	Status ErasureStatus `json:"status"`

	// CompletedAt is when erasure was completed
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// RejectionReason is reason for rejection (if rejected)
	RejectionReason string `json:"rejection_reason,omitempty"`
}

// ErasureReason represents reasons for erasure under GDPR
type ErasureReason string

const (
	// ErasureReasonNoLongerNecessary - data no longer necessary
	ErasureReasonNoLongerNecessary ErasureReason = "no_longer_necessary"

	// ErasureReasonWithdrawConsent - consent withdrawn
	ErasureReasonWithdrawConsent ErasureReason = "withdraw_consent"

	// ErasureReasonObjectProcessing - objection to processing
	ErasureReasonObjectProcessing ErasureReason = "object_processing"

	// ErasureReasonUnlawfulProcessing - unlawful processing
	ErasureReasonUnlawfulProcessing ErasureReason = "unlawful_processing"

	// ErasureReasonLegalObligation - legal obligation
	ErasureReasonLegalObligation ErasureReason = "legal_obligation"
)

// ErasureStatus represents the status of an erasure request
type ErasureStatus string

const (
	// ErasureStatusPending - request is pending
	ErasureStatusPending ErasureStatus = "pending"

	// ErasureStatusInProgress - erasure in progress
	ErasureStatusInProgress ErasureStatus = "in_progress"

	// ErasureStatusCompleted - erasure completed
	ErasureStatusCompleted ErasureStatus = "completed"

	// ErasureStatusRejected - request rejected (legal hold, etc.)
	ErasureStatusRejected ErasureStatus = "rejected"

	// ErasureStatusPartial - partially completed
	ErasureStatusPartial ErasureStatus = "partial"
)

// ============================================================================
// Privacy Manager Interface
// ============================================================================

// PrivacyManager manages privacy and data protection
type PrivacyManager interface {
	// HashPII hashes PII for privacy-preserving storage
	HashPII(value string, salt []byte) string

	// AnonymizePII anonymizes PII by replacing with pseudonym
	AnonymizePII(value string) string

	// ValidateDataMinimization validates data minimization compliance
	ValidateDataMinimization(fields map[string]string, requiredFields []string) error

	// ProcessSubjectAccessRequest processes a GDPR SAR
	ProcessSubjectAccessRequest(ctx context.Context, req *SubjectAccessRequest) ([]byte, error)

	// ProcessErasureRequest processes a GDPR erasure request
	ProcessErasureRequest(ctx context.Context, req *ErasureRequest) error

	// GetRetentionPolicy returns the retention policy for a jurisdiction
	GetRetentionPolicy(jurisdiction string) (*RetentionPolicy, error)

	// SchedulePurge schedules data for purge
	SchedulePurge(dataType string, subjectID string, purgeAt time.Time) error

	// ExportData exports data for portability (GDPR Article 20)
	ExportData(ctx context.Context, subjectID string) ([]byte, error)
}

// ============================================================================
// Privacy Manager Implementation
// ============================================================================

// privacyManager implements PrivacyManager
type privacyManager struct {
	config           PrivacyConfig
	retentionPolicies map[string]*RetentionPolicy
	erasureRequests  map[string]*ErasureRequest
	sarRequests      map[string]*SubjectAccessRequest
	purgeTasks       map[string]time.Time
	legalHolds       map[string]time.Time
	mu               sync.RWMutex
}

// newPrivacyManager creates a new privacy manager
func newPrivacyManager(config PrivacyConfig) *privacyManager {
	return &privacyManager{
		config:            config,
		retentionPolicies: make(map[string]*RetentionPolicy),
		erasureRequests:   make(map[string]*ErasureRequest),
		sarRequests:       make(map[string]*SubjectAccessRequest),
		purgeTasks:        make(map[string]time.Time),
		legalHolds:        make(map[string]time.Time),
	}
}

// HashPII creates a privacy-preserving hash of PII
func (p *privacyManager) HashPII(value string, salt []byte) string {
	if salt == nil {
		salt = make([]byte, 16)
		_, _ = rand.Read(salt)
	}

	// Use SHA-256 with salt for hashing
	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(value))
	hash := h.Sum(nil)

	return hex.EncodeToString(hash)
}

// AnonymizePII creates an anonymous pseudonym for PII
func (p *privacyManager) AnonymizePII(value string) string {
	// Generate a consistent but non-reversible pseudonym
	h := sha256.Sum256([]byte(value))
	return "ANON-" + base64.RawURLEncoding.EncodeToString(h[:8])
}

// ValidateDataMinimization ensures only required data is collected
func (p *privacyManager) ValidateDataMinimization(fields map[string]string, requiredFields []string) error {
	if !p.config.DataMinimization {
		return nil
	}

	requiredSet := make(map[string]bool)
	for _, f := range requiredFields {
		requiredSet[f] = true
	}

	// Check for unnecessary fields
	var unnecessaryFields []string
	for field := range fields {
		if !requiredSet[field] {
			unnecessaryFields = append(unnecessaryFields, field)
		}
	}

	if len(unnecessaryFields) > 0 {
		return fmt.Errorf("%w: unnecessary fields provided: %v",
			ErrDataMinimizationViolation, unnecessaryFields)
	}

	return nil
}

// ProcessSubjectAccessRequest processes a GDPR Article 15 SAR
func (p *privacyManager) ProcessSubjectAccessRequest(ctx context.Context, req *SubjectAccessRequest) ([]byte, error) {
	if !p.config.SubjectAccessRequestEnabled {
		return nil, errors.New("subject access requests not enabled")
	}

	p.mu.Lock()
	p.sarRequests[req.ID] = req
	req.Status = "processing"
	p.mu.Unlock()

	// In production, this would:
	// 1. Verify the requester's identity
	// 2. Collect all data held about the subject
	// 3. Format in a structured, commonly used format (JSON)
	// 4. Deliver within 30 days

	// For now, return a placeholder response
	response := map[string]interface{}{
		"subject_id":   req.SubjectID,
		"request_id":   req.ID,
		"data_held":    []string{},
		"purposes":     []string{"identity_verification"},
		"recipients":   []string{},
		"retention":    "see_privacy_policy",
		"source":       "user_provided",
		"generated_at": time.Now().Format(time.RFC3339),
	}

	req.Status = "completed"

	// Marshal to JSON (simplified)
	return []byte(fmt.Sprintf("%+v", response)), nil
}

// ProcessErasureRequest processes a GDPR Article 17 erasure request
func (p *privacyManager) ProcessErasureRequest(ctx context.Context, req *ErasureRequest) error {
	if !p.config.RightToErasureEnabled {
		return errors.New("right to erasure not enabled")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check for legal hold
	if holdExpires, hasHold := p.legalHolds[req.SubjectID]; hasHold {
		if time.Now().Before(holdExpires) {
			req.Status = ErasureStatusRejected
			req.RejectionReason = "data under legal hold until " + holdExpires.Format(time.RFC3339)
			p.erasureRequests[req.ID] = req
			return ErrErasureNotAllowed
		}
	}

	// Process erasure
	req.Status = ErasureStatusInProgress
	p.erasureRequests[req.ID] = req

	// In production, this would:
	// 1. Identify all data stores containing the subject's data
	// 2. Delete or anonymize data in each store
	// 3. Notify any third parties who received the data
	// 4. Document the erasure for compliance records

	// Mark as completed
	now := time.Now()
	req.Status = ErasureStatusCompleted
	req.CompletedAt = &now

	return nil
}

// GetRetentionPolicy returns the retention policy for a jurisdiction
func (p *privacyManager) GetRetentionPolicy(jurisdiction string) (*RetentionPolicy, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if policy, ok := p.retentionPolicies[jurisdiction]; ok {
		return policy, nil
	}

	// Return default based on jurisdiction
	switch {
	case strings.HasPrefix(jurisdiction, "EU") || jurisdiction == "GB":
		// GDPR jurisdictions - shorter retention
		return &RetentionPolicy{
			ResultRetentionDays:   30,
			AuditLogRetentionDays: 365,
			ConsentRetentionDays:  365,
			AutoPurge:             true,
		}, nil
	case strings.HasPrefix(jurisdiction, "US-CA"):
		// CCPA jurisdiction
		return &RetentionPolicy{
			ResultRetentionDays:   90,
			AuditLogRetentionDays: 365 * 7,
			ConsentRetentionDays:  365 * 7,
			AutoPurge:             true,
		}, nil
	default:
		// Default retention
		return &RetentionPolicy{
			ResultRetentionDays:   90,
			AuditLogRetentionDays: 365 * 7,
			ConsentRetentionDays:  365 * 7,
			AutoPurge:             true,
		}, nil
	}
}

// SchedulePurge schedules data for purge
func (p *privacyManager) SchedulePurge(dataType string, subjectID string, purgeAt time.Time) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := dataType + ":" + subjectID
	p.purgeTasks[key] = purgeAt

	return nil
}

// ExportData exports data for portability (GDPR Article 20)
func (p *privacyManager) ExportData(ctx context.Context, subjectID string) ([]byte, error) {
	if !p.config.DataPortabilityEnabled {
		return nil, errors.New("data portability not enabled")
	}

	// In production, this would:
	// 1. Collect all data about the subject
	// 2. Format in a structured, machine-readable format (JSON/CSV)
	// 3. Include verification history, consent records, etc.

	export := map[string]interface{}{
		"subject_id":    subjectID,
		"exported_at":   time.Now().Format(time.RFC3339),
		"format":        "json",
		"version":       "1.0",
		"verifications": []interface{}{},
		"consents":      []interface{}{},
	}

	return []byte(fmt.Sprintf("%+v", export)), nil
}

// SetLegalHold places a legal hold on a subject's data
func (p *privacyManager) SetLegalHold(subjectID string, holdUntil time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.legalHolds[subjectID] = holdUntil
}

// RemoveLegalHold removes a legal hold
func (p *privacyManager) RemoveLegalHold(subjectID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.legalHolds, subjectID)
}

// ============================================================================
// PII Processing Utilities
// ============================================================================

// RedactPII redacts PII from a string for logging
func RedactPII(value string) string {
	if len(value) <= 4 {
		return strings.Repeat("*", len(value))
	}
	return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
}

// RedactDocumentNumber redacts a document number for logging
func RedactDocumentNumber(docNumber string) string {
	if len(docNumber) <= 4 {
		return strings.Repeat("*", len(docNumber))
	}
	return strings.Repeat("*", len(docNumber)-4) + docNumber[len(docNumber)-4:]
}

// RedactEmail redacts an email address
func RedactEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return RedactPII(email)
	}
	return RedactPII(parts[0]) + "@" + parts[1]
}

// HashDocumentNumber creates a salted hash of a document number
func HashDocumentNumber(docNumber string, salt []byte) string {
	if salt == nil {
		salt = make([]byte, 16)
		_, _ = rand.Read(salt)
	}

	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(strings.ToUpper(strings.TrimSpace(docNumber))))
	return hex.EncodeToString(h.Sum(nil))
}

// IsGDPRJurisdiction checks if a jurisdiction is subject to GDPR
func IsGDPRJurisdiction(jurisdiction string) bool {
	gdprCodes := map[string]bool{
		"AT": true, "BE": true, "BG": true, "HR": true, "CY": true,
		"CZ": true, "DK": true, "EE": true, "FI": true, "FR": true,
		"DE": true, "GR": true, "HU": true, "IE": true, "IT": true,
		"LV": true, "LT": true, "LU": true, "MT": true, "NL": true,
		"PL": true, "PT": true, "RO": true, "SK": true, "SI": true,
		"ES": true, "SE": true,
		"IS": true, "LI": true, "NO": true, // EEA
		"GB": true, // UK GDPR
		"EU": true, // Generic EU
	}

	// Check exact match or country prefix
	if gdprCodes[jurisdiction] {
		return true
	}
	if len(jurisdiction) > 2 {
		return gdprCodes[jurisdiction[:2]]
	}
	return false
}

// IsCCPAJurisdiction checks if a jurisdiction is subject to CCPA
func IsCCPAJurisdiction(jurisdiction string) bool {
	return jurisdiction == "US-CA" || jurisdiction == "CA"
}

// GetRequiredConsentScope returns required consent scope for jurisdiction
func GetRequiredConsentScope(jurisdiction string) []string {
	if IsGDPRJurisdiction(jurisdiction) {
		return []string{
			"identity_verification",
			"government_data_access",
			"data_retention",
			"cross_border_transfer",
		}
	}
	if IsCCPAJurisdiction(jurisdiction) {
		return []string{
			"identity_verification",
			"sale_opt_out",
		}
	}
	return []string{
		"identity_verification",
	}
}
