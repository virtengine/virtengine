// Package types provides types for the VEID module.
//
// VE-225: Security controls - tokenization/pseudonymization + retention policy enforcement
// This file defines types for privacy-by-design controls including tokenization,
// pseudonymization, and retention policy enforcement for identity data flows.
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// SecurityControlsVersion is the current version of security controls format
const SecurityControlsVersion uint32 = 1

// ============================================================================
// Tokenization Types
// ============================================================================

// TokenType identifies the type of tokenized data
type TokenType string

const (
	// TokenTypeIdentityRef tokenizes identity references
	TokenTypeIdentityRef TokenType = "identity_ref"

	// TokenTypeScopeRef tokenizes scope references
	TokenTypeScopeRef TokenType = "scope_ref"

	// TokenTypeArtifactRef tokenizes artifact references
	TokenTypeArtifactRef TokenType = "artifact_ref"

	// TokenTypeVerificationRef tokenizes verification references
	TokenTypeVerificationRef TokenType = "verification_ref"

	// TokenTypeSessionRef tokenizes session references
	TokenTypeSessionRef TokenType = "session_ref"
)

// AllTokenTypes returns all valid token types
func AllTokenTypes() []TokenType {
	return []TokenType{
		TokenTypeIdentityRef,
		TokenTypeScopeRef,
		TokenTypeArtifactRef,
		TokenTypeVerificationRef,
		TokenTypeSessionRef,
	}
}

// IsValidTokenType checks if a token type is valid
func IsValidTokenType(t TokenType) bool {
	for _, valid := range AllTokenTypes() {
		if t == valid {
			return true
		}
	}
	return false
}

// ExternalToken represents a tokenized reference for external systems
// External systems see only the token, never the underlying data
type ExternalToken struct {
	// Token is the opaque token value
	Token string `json:"token"`

	// TokenType identifies what kind of data this represents
	TokenType TokenType `json:"token_type"`

	// IssuedAt is when this token was issued
	IssuedAt time.Time `json:"issued_at"`

	// ExpiresAt is when this token expires
	ExpiresAt time.Time `json:"expires_at"`

	// Scope limits what the token can be used for
	Scope string `json:"scope,omitempty"`

	// IsRevoked indicates if the token has been revoked
	IsRevoked bool `json:"is_revoked"`
}

// TokenMapping maps an external token to an internal reference
// This is stored internally and never exposed externally
type TokenMapping struct {
	// Token is the external token
	Token string `json:"token"`

	// TokenType identifies the token type
	TokenType TokenType `json:"token_type"`

	// InternalReference is the actual internal reference
	InternalReference string `json:"internal_reference"`

	// AccountAddress is the owning account
	AccountAddress string `json:"account_address"`

	// CreatedAt is when this mapping was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when this mapping expires
	ExpiresAt time.Time `json:"expires_at"`

	// IsRevoked indicates if this mapping is revoked
	IsRevoked bool `json:"is_revoked"`

	// RevokedAt is when this mapping was revoked
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	// UsageCount tracks how many times this token was used
	UsageCount uint64 `json:"usage_count"`

	// LastUsedAt is when this token was last used
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// NewTokenMapping creates a new token mapping
func NewTokenMapping(
	token string,
	tokenType TokenType,
	internalRef string,
	accountAddress string,
	createdAt time.Time,
	ttlSeconds int64,
) *TokenMapping {
	expiresAt := createdAt.Add(time.Duration(ttlSeconds) * time.Second)
	return &TokenMapping{
		Token:             token,
		TokenType:         tokenType,
		InternalReference: internalRef,
		AccountAddress:    accountAddress,
		CreatedAt:         createdAt,
		ExpiresAt:         expiresAt,
	}
}

// IsValid returns true if the token mapping is still valid
func (m *TokenMapping) IsValid(now time.Time) bool {
	return !m.IsRevoked && now.Before(m.ExpiresAt)
}

// RecordUsage records a token usage
func (m *TokenMapping) RecordUsage(usedAt time.Time) {
	m.UsageCount++
	m.LastUsedAt = &usedAt
}

// Revoke revokes the token mapping
func (m *TokenMapping) Revoke(revokedAt time.Time) {
	m.IsRevoked = true
	m.RevokedAt = &revokedAt
}

// GenerateToken creates a secure token for a given internal reference
func GenerateToken(internalRef string, salt string, createdAt time.Time) string {
	data := fmt.Sprintf("%s|%s|%d", internalRef, salt, createdAt.UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// ============================================================================
// Pseudonymization Types
// ============================================================================

// PseudonymType identifies the type of pseudonymized identifier
type PseudonymType string

const (
	// PseudonymTypeAccount pseudonymizes account addresses
	PseudonymTypeAccount PseudonymType = "account"

	// PseudonymTypeValidator pseudonymizes validator addresses
	PseudonymTypeValidator PseudonymType = "validator"

	// PseudonymTypeProvider pseudonymizes provider addresses
	PseudonymTypeProvider PseudonymType = "provider"

	// PseudonymTypeSession pseudonymizes session IDs
	PseudonymTypeSession PseudonymType = "session"
)

// AllPseudonymTypes returns all valid pseudonym types
func AllPseudonymTypes() []PseudonymType {
	return []PseudonymType{
		PseudonymTypeAccount,
		PseudonymTypeValidator,
		PseudonymTypeProvider,
		PseudonymTypeSession,
	}
}

// Pseudonym represents a pseudonymized identifier for logs and telemetry
type Pseudonym struct {
	// Value is the pseudonymized value (never the real identifier)
	Value string `json:"value"`

	// Type identifies what kind of identifier this represents
	Type PseudonymType `json:"type"`

	// Context provides additional non-identifying context
	Context string `json:"context,omitempty"`
}

// PseudonymizationConfig defines how identifiers should be pseudonymized
type PseudonymizationConfig struct {
	// Enabled indicates if pseudonymization is enabled
	Enabled bool `json:"enabled"`

	// Salt is the secret salt used for pseudonymization
	// SECURITY: This should be stored securely and rotated periodically
	Salt string `json:"salt"`

	// RotationIntervalSeconds is how often the salt should be rotated
	RotationIntervalSeconds int64 `json:"rotation_interval_seconds"`

	// LastRotatedAt is when the salt was last rotated
	LastRotatedAt time.Time `json:"last_rotated_at"`

	// PreservePrefix preserves a prefix of the original identifier
	PreservePrefix int `json:"preserVIRTENGINE_prefix"`
}

// DefaultPseudonymizationConfig returns the default configuration
func DefaultPseudonymizationConfig() *PseudonymizationConfig {
	return &PseudonymizationConfig{
		Enabled:                 true,
		RotationIntervalSeconds: 86400 * 30, // 30 days
		PreservePrefix:          4,          // Preserve first 4 chars for debugging
	}
}

// Pseudonymize creates a pseudonym for an identifier
func Pseudonymize(identifier string, identType PseudonymType, salt string, preservePrefix int) *Pseudonym {
	// Combine identifier with salt and type
	data := fmt.Sprintf("%s|%s|%s", identifier, salt, identType)
	hash := sha256.Sum256([]byte(data))
	hashHex := hex.EncodeToString(hash[:16]) // Use first 16 bytes

	// Optionally preserve a prefix
	prefix := ""
	if preservePrefix > 0 && len(identifier) >= preservePrefix {
		prefix = identifier[:preservePrefix] + "_"
	}

	return &Pseudonym{
		Value: prefix + hashHex,
		Type:  identType,
	}
}

// ============================================================================
// Retention Policy Enforcement Types
// ============================================================================

// RetentionRuleVersion is the current version of retention rules format
const RetentionRuleVersion uint32 = 1

// RetentionRule defines a retention rule for a specific artifact type
type RetentionRule struct {
	// Version is the rule format version
	Version uint32 `json:"version"`

	// RuleID is a unique identifier for this rule
	RuleID string `json:"rule_id"`

	// ArtifactType is the type of artifact this rule applies to
	ArtifactType ArtifactType `json:"artifact_type"`

	// RetentionType defines how retention is calculated
	RetentionType RetentionType `json:"retention_type"`

	// RetentionDurationSeconds is the retention duration (for duration type)
	RetentionDurationSeconds int64 `json:"retention_duration_seconds,omitempty"`

	// RetentionBlockCount is the block count (for block_count type)
	RetentionBlockCount int64 `json:"retention_block_count,omitempty"`

	// AutoDelete indicates if artifacts should be auto-deleted on expiry
	AutoDelete bool `json:"auto_delete"`

	// NotifyBeforeExpirySeconds is when to notify before expiry
	NotifyBeforeExpirySeconds int64 `json:"notify_before_expiry_seconds,omitempty"`

	// IsEnabled indicates if this rule is active
	IsEnabled bool `json:"is_enabled"`

	// Priority determines rule precedence (higher = higher priority)
	Priority uint32 `json:"priority"`

	// CreatedAt is when this rule was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when this rule was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// NewRetentionRule creates a new retention rule
func NewRetentionRule(
	ruleID string,
	artifactType ArtifactType,
	retentionType RetentionType,
	durationSeconds int64,
	autoDelete bool,
	createdAt time.Time,
) *RetentionRule {
	return &RetentionRule{
		Version:                  RetentionRuleVersion,
		RuleID:                   ruleID,
		ArtifactType:             artifactType,
		RetentionType:            retentionType,
		RetentionDurationSeconds: durationSeconds,
		AutoDelete:               autoDelete,
		IsEnabled:                true,
		Priority:                 100,
		CreatedAt:                createdAt,
		UpdatedAt:                createdAt,
	}
}

// Validate validates the retention rule
func (r *RetentionRule) Validate() error {
	if r.Version == 0 || r.Version > RetentionRuleVersion {
		return ErrInvalidRetention.Wrapf("unsupported version: %d", r.Version)
	}
	if r.RuleID == "" {
		return ErrInvalidRetention.Wrap("rule_id cannot be empty")
	}
	if !IsValidArtifactType(r.ArtifactType) {
		return ErrInvalidRetention.Wrapf("invalid artifact type: %s", r.ArtifactType)
	}
	if !IsValidRetentionType(r.RetentionType) {
		return ErrInvalidRetention.Wrapf("invalid retention type: %s", r.RetentionType)
	}
	if r.RetentionType == RetentionTypeDuration && r.RetentionDurationSeconds <= 0 {
		return ErrInvalidRetention.Wrap("duration must be positive for duration retention")
	}
	if r.RetentionType == RetentionTypeBlockCount && r.RetentionBlockCount <= 0 {
		return ErrInvalidRetention.Wrap("block count must be positive for block_count retention")
	}
	return nil
}

// DefaultRetentionRules returns the default retention rules for artifact types
func DefaultRetentionRules(createdAt time.Time) []RetentionRule {
	return []RetentionRule{
		// Raw images: 90 days (short retention, high sensitivity)
		{
			Version:                  RetentionRuleVersion,
			RuleID:                   "default_raw_image",
			ArtifactType:             ArtifactTypeRawImage,
			RetentionType:            RetentionTypeDuration,
			RetentionDurationSeconds: 90 * 24 * 3600, // 90 days
			AutoDelete:               true,
			IsEnabled:                true,
			Priority:                 100,
			CreatedAt:                createdAt,
			UpdatedAt:                createdAt,
		},
		// Processed images: 180 days
		{
			Version:                  RetentionRuleVersion,
			RuleID:                   "default_processed_image",
			ArtifactType:             ArtifactTypeProcessedImage,
			RetentionType:            RetentionTypeDuration,
			RetentionDurationSeconds: 180 * 24 * 3600, // 180 days
			AutoDelete:               true,
			IsEnabled:                true,
			Priority:                 100,
			CreatedAt:                createdAt,
			UpdatedAt:                createdAt,
		},
		// Face embeddings: 1 year
		{
			Version:                  RetentionRuleVersion,
			RuleID:                   "default_face_embedding",
			ArtifactType:             ArtifactTypeFaceEmbedding,
			RetentionType:            RetentionTypeDuration,
			RetentionDurationSeconds: 365 * 24 * 3600, // 1 year
			AutoDelete:               true,
			IsEnabled:                true,
			Priority:                 100,
			CreatedAt:                createdAt,
			UpdatedAt:                createdAt,
		},
		// Document hashes: indefinite (low sensitivity, needed for re-verification)
		{
			Version:       RetentionRuleVersion,
			RuleID:        "default_document_hash",
			ArtifactType:  ArtifactTypeDocumentHash,
			RetentionType: RetentionTypeIndefinite,
			AutoDelete:    false,
			IsEnabled:     true,
			Priority:      100,
			CreatedAt:     createdAt,
			UpdatedAt:     createdAt,
		},
		// Verification records: 2 years
		{
			Version:                  RetentionRuleVersion,
			RuleID:                   "default_verification_record",
			ArtifactType:             ArtifactTypeVerificationRecord,
			RetentionType:            RetentionTypeDuration,
			RetentionDurationSeconds: 2 * 365 * 24 * 3600, // 2 years
			AutoDelete:               false,               // Keep for audit
			IsEnabled:                true,
			Priority:                 100,
			CreatedAt:                createdAt,
			UpdatedAt:                createdAt,
		},
		// OCR data: 180 days
		{
			Version:                  RetentionRuleVersion,
			RuleID:                   "default_ocr_data",
			ArtifactType:             ArtifactTypeOCRData,
			RetentionType:            RetentionTypeDuration,
			RetentionDurationSeconds: 180 * 24 * 3600, // 180 days
			AutoDelete:               true,
			IsEnabled:                true,
			Priority:                 100,
			CreatedAt:                createdAt,
			UpdatedAt:                createdAt,
		},
	}
}

// RetentionEnforcementResult represents the result of retention enforcement
type RetentionEnforcementResult struct {
	// EnforcementID is a unique identifier for this enforcement run
	EnforcementID string `json:"enforcement_id"`

	// RunAt is when enforcement was run
	RunAt time.Time `json:"run_at"`

	// ArtifactsScanned is the number of artifacts scanned
	ArtifactsScanned uint64 `json:"artifacts_scanned"`

	// ArtifactsExpired is the number of artifacts found expired
	ArtifactsExpired uint64 `json:"artifacts_expired"`

	// ArtifactsDeleted is the number of artifacts deleted
	ArtifactsDeleted uint64 `json:"artifacts_deleted"`

	// ArtifactsFailed is the number of artifacts that failed to delete
	ArtifactsFailed uint64 `json:"artifacts_failed"`

	// NotificationsSent is the number of expiry notifications sent
	NotificationsSent uint64 `json:"notifications_sent"`

	// ByType maps artifact types to their enforcement counts
	ByType map[ArtifactType]uint64 `json:"by_type"`

	// Duration is how long enforcement took
	DurationMs int64 `json:"duration_ms"`

	// BlockHeight is the block height at enforcement time
	BlockHeight int64 `json:"block_height"`
}

// NewRetentionEnforcementResult creates a new enforcement result
func NewRetentionEnforcementResult(enforcementID string, runAt time.Time, blockHeight int64) *RetentionEnforcementResult {
	return &RetentionEnforcementResult{
		EnforcementID: enforcementID,
		RunAt:         runAt,
		BlockHeight:   blockHeight,
		ByType:        make(map[ArtifactType]uint64),
	}
}

// ============================================================================
// Audit and Compliance Types
// ============================================================================

// SecurityAuditEvent represents a security-related audit event
// Uses pseudonymized identifiers for privacy
type SecurityAuditEvent struct {
	// EventID is a unique identifier for this event
	EventID string `json:"event_id"`

	// EventType is the type of security event
	EventType string `json:"event_type"`

	// AccountPseudonym is the pseudonymized account identifier
	AccountPseudonym string `json:"account_pseudonym"`

	// Action is the action taken
	Action string `json:"action"`

	// ResourceType is the type of resource affected
	ResourceType string `json:"resource_type"`

	// ResourcePseudonym is the pseudonymized resource identifier
	ResourcePseudonym string `json:"resource_pseudonym,omitempty"`

	// Outcome is the outcome (success/failure)
	Outcome string `json:"outcome"`

	// Metadata contains non-sensitive metadata
	Metadata map[string]string `json:"metadata,omitempty"`

	// Timestamp is when this event occurred
	Timestamp time.Time `json:"timestamp"`

	// BlockHeight is the block height when this occurred
	BlockHeight int64 `json:"block_height"`
}

// NewSecurityAuditEvent creates a new security audit event with pseudonymized identifiers
func NewSecurityAuditEvent(
	eventID string,
	eventType string,
	accountAddress string,
	action string,
	outcome string,
	salt string,
	timestamp time.Time,
	blockHeight int64,
) *SecurityAuditEvent {
	// Pseudonymize the account address for the audit log
	pseudonym := Pseudonymize(accountAddress, PseudonymTypeAccount, salt, 4)

	return &SecurityAuditEvent{
		EventID:          eventID,
		EventType:        eventType,
		AccountPseudonym: pseudonym.Value,
		Action:           action,
		Outcome:          outcome,
		Metadata:         make(map[string]string),
		Timestamp:        timestamp,
		BlockHeight:      blockHeight,
	}
}

// AddMetadata adds non-sensitive metadata to the event
func (e *SecurityAuditEvent) AddMetadata(key, value string) {
	e.Metadata[key] = value
}
