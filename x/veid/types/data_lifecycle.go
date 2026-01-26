package types

import (
	"time"
)

// ============================================================================
// Data Lifecycle and Retention Policy Types
// ============================================================================

// RetentionPolicyVersion is the current version of the retention policy format
const RetentionPolicyVersion uint32 = 1

// RetentionType defines how data should be retained
type RetentionType string

const (
	// RetentionTypeDuration indicates data should be retained for a specific duration
	RetentionTypeDuration RetentionType = "duration"

	// RetentionTypeBlockCount indicates data should be retained for a specific number of blocks
	RetentionTypeBlockCount RetentionType = "block_count"

	// RetentionTypeIndefinite indicates data should be retained indefinitely
	RetentionTypeIndefinite RetentionType = "indefinite"

	// RetentionTypeUntilRevoked indicates data should be retained until explicitly revoked
	RetentionTypeUntilRevoked RetentionType = "until_revoked"

	// RetentionTypeVerificationCycle indicates data should be retained until next verification
	RetentionTypeVerificationCycle RetentionType = "verification_cycle"
)

// AllRetentionTypes returns all valid retention types
func AllRetentionTypes() []RetentionType {
	return []RetentionType{
		RetentionTypeDuration,
		RetentionTypeBlockCount,
		RetentionTypeIndefinite,
		RetentionTypeUntilRevoked,
		RetentionTypeVerificationCycle,
	}
}

// IsValidRetentionType checks if a retention type is valid
func IsValidRetentionType(t RetentionType) bool {
	for _, valid := range AllRetentionTypes() {
		if t == valid {
			return true
		}
	}
	return false
}

// ArtifactType represents a type of identity artifact
type ArtifactType string

const (
	// ArtifactTypeRawImage represents a raw captured image
	ArtifactTypeRawImage ArtifactType = "raw_image"

	// ArtifactTypeProcessedImage represents a processed/enhanced image
	ArtifactTypeProcessedImage ArtifactType = "processed_image"

	// ArtifactTypeFaceEmbedding represents a face embedding vector
	ArtifactTypeFaceEmbedding ArtifactType = "face_embedding"

	// ArtifactTypeDocumentHash represents a document field hash
	ArtifactTypeDocumentHash ArtifactType = "document_hash"

	// ArtifactTypeBiometricHash represents a biometric data hash
	ArtifactTypeBiometricHash ArtifactType = "biometric_hash"

	// ArtifactTypeVerificationRecord represents a verification record
	ArtifactTypeVerificationRecord ArtifactType = "verification_record"

	// ArtifactTypeOCRData represents OCR extracted data
	ArtifactTypeOCRData ArtifactType = "ocr_data"
)

// AllArtifactTypes returns all valid artifact types
func AllArtifactTypes() []ArtifactType {
	return []ArtifactType{
		ArtifactTypeRawImage,
		ArtifactTypeProcessedImage,
		ArtifactTypeFaceEmbedding,
		ArtifactTypeDocumentHash,
		ArtifactTypeBiometricHash,
		ArtifactTypeVerificationRecord,
		ArtifactTypeOCRData,
	}
}

// IsValidArtifactType checks if an artifact type is valid
func IsValidArtifactType(t ArtifactType) bool {
	for _, valid := range AllArtifactTypes() {
		if t == valid {
			return true
		}
	}
	return false
}

// RetentionPolicy defines retention rules for identity artifacts
type RetentionPolicy struct {
	// Version is the policy format version
	Version uint32 `json:"version"`

	// PolicyID is a unique identifier for this policy
	PolicyID string `json:"policy_id"`

	// RetentionType specifies how retention is calculated
	RetentionType RetentionType `json:"retention_type"`

	// DurationSeconds is the retention duration in seconds (for duration type)
	DurationSeconds int64 `json:"duration_seconds,omitempty"`

	// BlockCount is the number of blocks to retain (for block_count type)
	BlockCount int64 `json:"block_count,omitempty"`

	// ExpiresAt is the calculated expiration time
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// ExpiresAtBlock is the calculated expiration block height
	ExpiresAtBlock *int64 `json:"expires_at_block,omitempty"`

	// CreatedAt is when this policy was created
	CreatedAt time.Time `json:"created_at"`

	// CreatedAtBlock is the block height when created
	CreatedAtBlock int64 `json:"created_at_block"`

	// DeleteOnExpiry indicates if data should be auto-deleted on expiry
	DeleteOnExpiry bool `json:"delete_on_expiry"`

	// NotifyBeforeExpiry is the duration before expiry to send notifications
	NotifyBeforeExpirySeconds int64 `json:"notify_before_expiry_seconds,omitempty"`

	// ExtensionAllowed indicates if the policy can be extended
	ExtensionAllowed bool `json:"extension_allowed"`

	// MaxExtensions is the maximum number of extensions allowed
	MaxExtensions uint32 `json:"max_extensions"`

	// CurrentExtensions is the current number of extensions applied
	CurrentExtensions uint32 `json:"current_extensions"`
}

// NewRetentionPolicyDuration creates a duration-based retention policy
func NewRetentionPolicyDuration(
	policyID string,
	durationSeconds int64,
	createdAt time.Time,
	createdAtBlock int64,
	deleteOnExpiry bool,
) *RetentionPolicy {
	expiresAt := createdAt.Add(time.Duration(durationSeconds) * time.Second)
	return &RetentionPolicy{
		Version:          RetentionPolicyVersion,
		PolicyID:         policyID,
		RetentionType:    RetentionTypeDuration,
		DurationSeconds:  durationSeconds,
		ExpiresAt:        &expiresAt,
		CreatedAt:        createdAt,
		CreatedAtBlock:   createdAtBlock,
		DeleteOnExpiry:   deleteOnExpiry,
		ExtensionAllowed: true,
		MaxExtensions:    3,
	}
}

// NewRetentionPolicyBlockCount creates a block-count-based retention policy
func NewRetentionPolicyBlockCount(
	policyID string,
	blockCount int64,
	createdAt time.Time,
	createdAtBlock int64,
	deleteOnExpiry bool,
) *RetentionPolicy {
	expiresAtBlock := createdAtBlock + blockCount
	return &RetentionPolicy{
		Version:          RetentionPolicyVersion,
		PolicyID:         policyID,
		RetentionType:    RetentionTypeBlockCount,
		BlockCount:       blockCount,
		ExpiresAtBlock:   &expiresAtBlock,
		CreatedAt:        createdAt,
		CreatedAtBlock:   createdAtBlock,
		DeleteOnExpiry:   deleteOnExpiry,
		ExtensionAllowed: true,
		MaxExtensions:    3,
	}
}

// NewRetentionPolicyIndefinite creates an indefinite retention policy
func NewRetentionPolicyIndefinite(
	policyID string,
	createdAt time.Time,
	createdAtBlock int64,
) *RetentionPolicy {
	return &RetentionPolicy{
		Version:          RetentionPolicyVersion,
		PolicyID:         policyID,
		RetentionType:    RetentionTypeIndefinite,
		CreatedAt:        createdAt,
		CreatedAtBlock:   createdAtBlock,
		DeleteOnExpiry:   false,
		ExtensionAllowed: false,
	}
}

// NewRetentionPolicyUntilRevoked creates an until-revoked retention policy
func NewRetentionPolicyUntilRevoked(
	policyID string,
	createdAt time.Time,
	createdAtBlock int64,
) *RetentionPolicy {
	return &RetentionPolicy{
		Version:          RetentionPolicyVersion,
		PolicyID:         policyID,
		RetentionType:    RetentionTypeUntilRevoked,
		CreatedAt:        createdAt,
		CreatedAtBlock:   createdAtBlock,
		DeleteOnExpiry:   true,
		ExtensionAllowed: false,
	}
}

// Validate validates the retention policy
func (p *RetentionPolicy) Validate() error {
	if p.Version == 0 || p.Version > RetentionPolicyVersion {
		return ErrInvalidParams.Wrapf("unsupported retention policy version: %d", p.Version)
	}

	if p.PolicyID == "" {
		return ErrInvalidParams.Wrap("policy_id cannot be empty")
	}

	if !IsValidRetentionType(p.RetentionType) {
		return ErrInvalidParams.Wrapf("invalid retention_type: %s", p.RetentionType)
	}

	switch p.RetentionType {
	case RetentionTypeDuration:
		if p.DurationSeconds <= 0 {
			return ErrInvalidParams.Wrap("duration_seconds must be positive")
		}
		if p.ExpiresAt == nil {
			return ErrInvalidParams.Wrap("expires_at required for duration type")
		}

	case RetentionTypeBlockCount:
		if p.BlockCount <= 0 {
			return ErrInvalidParams.Wrap("block_count must be positive")
		}
		if p.ExpiresAtBlock == nil {
			return ErrInvalidParams.Wrap("expires_at_block required for block_count type")
		}
	}

	if p.CreatedAt.IsZero() {
		return ErrInvalidParams.Wrap("created_at cannot be zero")
	}

	if p.CreatedAtBlock < 0 {
		return ErrInvalidParams.Wrap("created_at_block cannot be negative")
	}

	return nil
}

// IsExpired checks if the policy has expired
func (p *RetentionPolicy) IsExpired(now time.Time) bool {
	switch p.RetentionType {
	case RetentionTypeDuration:
		if p.ExpiresAt != nil {
			return now.After(*p.ExpiresAt)
		}
	case RetentionTypeIndefinite, RetentionTypeUntilRevoked:
		return false
	}
	return false
}

// IsExpiredAtBlock checks if the policy has expired at a specific block height
func (p *RetentionPolicy) IsExpiredAtBlock(blockHeight int64) bool {
	switch p.RetentionType {
	case RetentionTypeBlockCount:
		if p.ExpiresAtBlock != nil {
			return blockHeight >= *p.ExpiresAtBlock
		}
	case RetentionTypeIndefinite, RetentionTypeUntilRevoked:
		return false
	}
	return false
}

// CanExtend checks if the policy can be extended
func (p *RetentionPolicy) CanExtend() bool {
	if !p.ExtensionAllowed {
		return false
	}
	return p.CurrentExtensions < p.MaxExtensions
}

// Extend extends the retention policy by the original duration/block count
func (p *RetentionPolicy) Extend() error {
	if !p.CanExtend() {
		return ErrInvalidParams.Wrap("policy cannot be extended")
	}

	switch p.RetentionType {
	case RetentionTypeDuration:
		if p.ExpiresAt != nil && p.DurationSeconds > 0 {
			newExpiry := p.ExpiresAt.Add(time.Duration(p.DurationSeconds) * time.Second)
			p.ExpiresAt = &newExpiry
		}
	case RetentionTypeBlockCount:
		if p.ExpiresAtBlock != nil && p.BlockCount > 0 {
			newExpiry := *p.ExpiresAtBlock + p.BlockCount
			p.ExpiresAtBlock = &newExpiry
		}
	default:
		return ErrInvalidParams.Wrapf("cannot extend retention type: %s", p.RetentionType)
	}

	p.CurrentExtensions++
	return nil
}

// TimeUntilExpiry returns the duration until expiry (for duration-based policies)
func (p *RetentionPolicy) TimeUntilExpiry(now time.Time) time.Duration {
	if p.RetentionType != RetentionTypeDuration || p.ExpiresAt == nil {
		return 0
	}
	if now.After(*p.ExpiresAt) {
		return 0
	}
	return p.ExpiresAt.Sub(now)
}

// ShouldNotify checks if a notification should be sent
func (p *RetentionPolicy) ShouldNotify(now time.Time) bool {
	if p.NotifyBeforeExpirySeconds <= 0 {
		return false
	}

	if p.RetentionType != RetentionTypeDuration || p.ExpiresAt == nil {
		return false
	}

	notifyTime := p.ExpiresAt.Add(-time.Duration(p.NotifyBeforeExpirySeconds) * time.Second)
	return now.After(notifyTime) && now.Before(*p.ExpiresAt)
}

// ============================================================================
// Data Lifecycle Rules
// ============================================================================

// DataLifecycleRulesVersion is the current version of lifecycle rules
const DataLifecycleRulesVersion uint32 = 1

// DataLifecycleRules defines the rules for data retention and disposal
type DataLifecycleRules struct {
	// Version is the rules format version
	Version uint32 `json:"version"`

	// ArtifactPolicies maps artifact types to their retention policies
	ArtifactPolicies map[ArtifactType]*ArtifactRetentionRule `json:"artifact_policies"`
}

// ArtifactRetentionRule defines retention rules for a specific artifact type
type ArtifactRetentionRule struct {
	// ArtifactType is the type of artifact this rule applies to
	ArtifactType ArtifactType `json:"artifact_type"`

	// AllowOnChain indicates if this artifact type can be stored on-chain
	AllowOnChain bool `json:"allow_on_chain"`

	// RequireEncryption indicates if the artifact must be encrypted
	RequireEncryption bool `json:"require_encryption"`

	// MaxRetentionDays is the maximum retention duration in days
	// 0 means no limit
	MaxRetentionDays uint32 `json:"max_retention_days"`

	// DefaultRetentionDays is the default retention duration in days
	DefaultRetentionDays uint32 `json:"default_retention_days"`

	// DeleteAfterVerification indicates if artifact should be deleted after verification
	DeleteAfterVerification bool `json:"delete_after_verification"`

	// AllowOffChainStorage indicates if artifact can be stored off-chain
	AllowOffChainStorage bool `json:"allow_off_chain_storage"`

	// RequireUserConsent indicates if user consent is required for storage
	RequireUserConsent bool `json:"require_user_consent"`

	// Description is a human-readable description of the rule
	Description string `json:"description"`
}

// DefaultDataLifecycleRules returns the default lifecycle rules
// These rules implement privacy-by-design and data minimization principles
func DefaultDataLifecycleRules() *DataLifecycleRules {
	return &DataLifecycleRules{
		Version: DataLifecycleRulesVersion,
		ArtifactPolicies: map[ArtifactType]*ArtifactRetentionRule{
			// Raw images NEVER stored on-chain
			ArtifactTypeRawImage: {
				ArtifactType:            ArtifactTypeRawImage,
				AllowOnChain:            false, // NEVER on-chain
				RequireEncryption:       true,
				MaxRetentionDays:        30,  // Max 30 days off-chain
				DefaultRetentionDays:    7,   // Default 7 days
				DeleteAfterVerification: true, // Delete after processing
				AllowOffChainStorage:    true,
				RequireUserConsent:      true,
				Description:             "Raw captured images are never stored on-chain and are deleted after verification",
			},
			// Processed images - temporary only
			ArtifactTypeProcessedImage: {
				ArtifactType:            ArtifactTypeProcessedImage,
				AllowOnChain:            false,
				RequireEncryption:       true,
				MaxRetentionDays:        7,
				DefaultRetentionDays:    1,
				DeleteAfterVerification: true,
				AllowOffChainStorage:    true,
				RequireUserConsent:      true,
				Description:             "Processed images are temporary and deleted after embedding extraction",
			},
			// Face embeddings - stored as hashes on-chain only
			ArtifactTypeFaceEmbedding: {
				ArtifactType:            ArtifactTypeFaceEmbedding,
				AllowOnChain:            true, // Hash only, not raw embedding
				RequireEncryption:       true, // Raw embedding encrypted off-chain
				MaxRetentionDays:        0,    // No limit for hashes
				DefaultRetentionDays:    365,
				DeleteAfterVerification: false,
				AllowOffChainStorage:    true,
				RequireUserConsent:      true,
				Description:             "Face embedding hash stored on-chain; encrypted embedding stored off-chain",
			},
			// Document hashes - stored on-chain
			ArtifactTypeDocumentHash: {
				ArtifactType:            ArtifactTypeDocumentHash,
				AllowOnChain:            true,
				RequireEncryption:       false, // Hashes don't need encryption
				MaxRetentionDays:        0,
				DefaultRetentionDays:    365,
				DeleteAfterVerification: false,
				AllowOffChainStorage:    false,
				RequireUserConsent:      true,
				Description:             "Document field hashes stored on-chain for verification",
			},
			// Biometric hashes - stored on-chain
			ArtifactTypeBiometricHash: {
				ArtifactType:            ArtifactTypeBiometricHash,
				AllowOnChain:            true,
				RequireEncryption:       false,
				MaxRetentionDays:        0,
				DefaultRetentionDays:    365,
				DeleteAfterVerification: false,
				AllowOffChainStorage:    false,
				RequireUserConsent:      true,
				Description:             "Biometric data hashes stored on-chain for verification",
			},
			// Verification records - stored on-chain
			ArtifactTypeVerificationRecord: {
				ArtifactType:            ArtifactTypeVerificationRecord,
				AllowOnChain:            true,
				RequireEncryption:       false,
				MaxRetentionDays:        0,
				DefaultRetentionDays:    0, // Indefinite
				DeleteAfterVerification: false,
				AllowOffChainStorage:    true,
				RequireUserConsent:      false,
				Description:             "Verification records stored on-chain for audit and consensus",
			},
			// OCR data - temporary only
			ArtifactTypeOCRData: {
				ArtifactType:            ArtifactTypeOCRData,
				AllowOnChain:            false, // Only hashes on-chain
				RequireEncryption:       true,
				MaxRetentionDays:        30,
				DefaultRetentionDays:    7,
				DeleteAfterVerification: true,
				AllowOffChainStorage:    true,
				RequireUserConsent:      true,
				Description:             "OCR extracted data is hashed; raw data deleted after processing",
			},
		},
	}
}

// Validate validates the lifecycle rules
func (r *DataLifecycleRules) Validate() error {
	if r.Version == 0 || r.Version > DataLifecycleRulesVersion {
		return ErrInvalidParams.Wrapf("unsupported lifecycle rules version: %d", r.Version)
	}

	if len(r.ArtifactPolicies) == 0 {
		return ErrInvalidParams.Wrap("artifact_policies cannot be empty")
	}

	for artifactType, rule := range r.ArtifactPolicies {
		if !IsValidArtifactType(artifactType) {
			return ErrInvalidParams.Wrapf("invalid artifact type: %s", artifactType)
		}
		if rule == nil {
			return ErrInvalidParams.Wrapf("nil rule for artifact type: %s", artifactType)
		}
		if rule.ArtifactType != artifactType {
			return ErrInvalidParams.Wrapf("mismatched artifact type in rule: %s vs %s", rule.ArtifactType, artifactType)
		}
	}

	return nil
}

// GetRule returns the retention rule for an artifact type
func (r *DataLifecycleRules) GetRule(artifactType ArtifactType) (*ArtifactRetentionRule, bool) {
	rule, found := r.ArtifactPolicies[artifactType]
	return rule, found
}

// CanStoreOnChain checks if an artifact type can be stored on-chain
func (r *DataLifecycleRules) CanStoreOnChain(artifactType ArtifactType) bool {
	rule, found := r.ArtifactPolicies[artifactType]
	if !found {
		return false
	}
	return rule.AllowOnChain
}

// RequiresEncryption checks if an artifact type requires encryption
func (r *DataLifecycleRules) RequiresEncryption(artifactType ArtifactType) bool {
	rule, found := r.ArtifactPolicies[artifactType]
	if !found {
		return true // Default to requiring encryption
	}
	return rule.RequireEncryption
}

// ShouldDeleteAfterVerification checks if an artifact should be deleted after verification
func (r *DataLifecycleRules) ShouldDeleteAfterVerification(artifactType ArtifactType) bool {
	rule, found := r.ArtifactPolicies[artifactType]
	if !found {
		return true // Default to deleting after verification
	}
	return rule.DeleteAfterVerification
}

// CreateRetentionPolicy creates a retention policy for an artifact based on rules
func (r *DataLifecycleRules) CreateRetentionPolicy(
	artifactType ArtifactType,
	policyID string,
	createdAt time.Time,
	createdAtBlock int64,
) (*RetentionPolicy, error) {
	rule, found := r.ArtifactPolicies[artifactType]
	if !found {
		return nil, ErrInvalidParams.Wrapf("no rule for artifact type: %s", artifactType)
	}

	if rule.DefaultRetentionDays == 0 {
		// Indefinite retention
		return NewRetentionPolicyIndefinite(policyID, createdAt, createdAtBlock), nil
	}

	durationSeconds := int64(rule.DefaultRetentionDays) * 24 * 60 * 60
	return NewRetentionPolicyDuration(
		policyID,
		durationSeconds,
		createdAt,
		createdAtBlock,
		rule.DeleteAfterVerification,
	), nil
}
