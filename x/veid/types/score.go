package types

import (
	"crypto/sha256"
	"time"
)

// ============================================================================
// Score Tier Constants (aligned with veid-flow-spec.md)
// ============================================================================

const (
	// TierUnverified represents accounts with no verification or score < 50
	TierUnverified int = 0

	// TierBasic represents accounts with score 50-69
	TierBasic int = 1

	// TierStandard represents accounts with score 70-84
	TierStandard int = 2

	// TierPremium represents accounts with score 85-100
	TierPremium int = 3
)

// Score threshold constants (from veid-flow-spec.md)
const (
	// ThresholdBasic is the minimum score for basic marketplace access
	ThresholdBasic uint32 = 50

	// ThresholdStandard is the minimum score for standard/provider access
	ThresholdStandard uint32 = 70

	// ThresholdPremium is the minimum score for premium/validator access
	ThresholdPremium uint32 = 85

	// MaxScore is the maximum possible identity score
	MaxScore uint32 = 100
)

// AccountStatus represents the overall account verification status
type AccountStatus string

const (
	// AccountStatusUnknown indicates an uninitialized status
	AccountStatusUnknown AccountStatus = "unknown"

	// AccountStatusPending indicates verification is in progress
	AccountStatusPending AccountStatus = "pending"

	// AccountStatusInProgress indicates active ML scoring
	AccountStatusInProgress AccountStatus = "in_progress"

	// AccountStatusVerified indicates account is verified
	AccountStatusVerified AccountStatus = "verified"

	// AccountStatusRejected indicates verification was rejected
	AccountStatusRejected AccountStatus = "rejected"

	// AccountStatusExpired indicates verification has expired
	AccountStatusExpired AccountStatus = "expired"

	// AccountStatusNeedsAdditionalFactor indicates additional verification needed
	AccountStatusNeedsAdditionalFactor AccountStatus = "needs_additional_factor"
)

// AllAccountStatuses returns all valid account statuses
func AllAccountStatuses() []AccountStatus {
	return []AccountStatus{
		AccountStatusUnknown,
		AccountStatusPending,
		AccountStatusInProgress,
		AccountStatusVerified,
		AccountStatusRejected,
		AccountStatusExpired,
		AccountStatusNeedsAdditionalFactor,
	}
}

// IsValidAccountStatus checks if a status is valid
func IsValidAccountStatus(status AccountStatus) bool {
	for _, s := range AllAccountStatuses() {
		if s == status {
			return true
		}
	}
	return false
}

// ============================================================================
// Identity Score Types
// ============================================================================

// IdentityScore represents the current identity score for an account
type IdentityScore struct {
	// AccountAddress is the blockchain address this score belongs to
	AccountAddress string `json:"account_address"`

	// Score is the current identity score (0-100)
	Score uint32 `json:"score"`

	// Status is the current verification status
	Status AccountStatus `json:"status"`

	// ModelVersion is the ML model version used to compute this score
	// e.g., "v1.0.0", "v2.1.0"
	ModelVersion string `json:"model_version"`

	// ComputedAt is when this score was computed
	ComputedAt time.Time `json:"computed_at"`

	// ExpiresAt is when this score expires (optional)
	// After expiration, re-verification may be required
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// VerificationHash is the SHA-256 hash of inputs used for scoring
	// This allows verification that the same inputs produce the same score
	VerificationHash []byte `json:"verification_hash"`

	// BlockHeight is the block height when the score was recorded
	BlockHeight int64 `json:"block_height"`
}

// NewIdentityScore creates a new identity score
func NewIdentityScore(
	accountAddress string,
	score uint32,
	status AccountStatus,
	modelVersion string,
	computedAt time.Time,
	blockHeight int64,
	verificationInputs []byte,
) *IdentityScore {
	hash := sha256.Sum256(verificationInputs)
	return &IdentityScore{
		AccountAddress:   accountAddress,
		Score:            score,
		Status:           status,
		ModelVersion:     modelVersion,
		ComputedAt:       computedAt,
		VerificationHash: hash[:],
		BlockHeight:      blockHeight,
	}
}

// Validate validates the identity score
func (s *IdentityScore) Validate() error {
	if s.AccountAddress == "" {
		return ErrInvalidAddress.Wrap("account address cannot be empty")
	}

	if s.Score > MaxScore {
		return ErrInvalidScore.Wrapf("score %d exceeds maximum %d", s.Score, MaxScore)
	}

	if !IsValidAccountStatus(s.Status) {
		return ErrInvalidVerificationStatus.Wrapf("invalid status: %s", s.Status)
	}

	if s.ComputedAt.IsZero() {
		return ErrInvalidScore.Wrap("computed_at cannot be zero")
	}

	return nil
}

// GetTier returns the tier based on the current score and status
func (s *IdentityScore) GetTier() int {
	return ComputeTierFromScoreValue(s.Score, s.Status)
}

// IsExpired checks if the score has expired
func (s *IdentityScore) IsExpired() bool {
	if s.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*s.ExpiresAt)
}

// IsAboveThreshold checks if the score is above a given threshold
func (s *IdentityScore) IsAboveThreshold(threshold uint32) bool {
	return s.Score >= threshold && s.Status == AccountStatusVerified
}

// ============================================================================
// Score History Types
// ============================================================================

// ScoreHistoryEntry represents a single entry in the score history
type ScoreHistoryEntry struct {
	// Score is the score value at this point
	Score uint32 `json:"score"`

	// Status is the verification status at this point
	Status AccountStatus `json:"status"`

	// ModelVersion is the ML model version used
	ModelVersion string `json:"model_version"`

	// ComputedAt is when this score was computed
	ComputedAt time.Time `json:"computed_at"`

	// BlockHeight is the block height when recorded
	BlockHeight int64 `json:"block_height"`

	// Reason is an optional description of the score change
	Reason string `json:"reason,omitempty"`

	// VerificationHash is the hash of verification inputs
	VerificationHash []byte `json:"verification_hash,omitempty"`
}

// NewScoreHistoryEntry creates a new score history entry
func NewScoreHistoryEntry(
	score uint32,
	status AccountStatus,
	modelVersion string,
	computedAt time.Time,
	blockHeight int64,
	reason string,
) *ScoreHistoryEntry {
	return &ScoreHistoryEntry{
		Score:        score,
		Status:       status,
		ModelVersion: modelVersion,
		ComputedAt:   computedAt,
		BlockHeight:  blockHeight,
		Reason:       reason,
	}
}

// ScoreHistory represents the complete score history for an account
type ScoreHistory struct {
	// AccountAddress is the account this history belongs to
	AccountAddress string `json:"account_address"`

	// Entries are the historical score entries (ordered by time, newest first)
	Entries []ScoreHistoryEntry `json:"entries"`
}

// NewScoreHistory creates a new score history
func NewScoreHistory(accountAddress string) *ScoreHistory {
	return &ScoreHistory{
		AccountAddress: accountAddress,
		Entries:        make([]ScoreHistoryEntry, 0),
	}
}

// AddEntry adds a new entry to the history (prepends to maintain newest-first order)
func (h *ScoreHistory) AddEntry(entry ScoreHistoryEntry) {
	h.Entries = append([]ScoreHistoryEntry{entry}, h.Entries...)
}

// GetLatest returns the most recent history entry
func (h *ScoreHistory) GetLatest() (ScoreHistoryEntry, bool) {
	if len(h.Entries) == 0 {
		return ScoreHistoryEntry{}, false
	}
	return h.Entries[0], true
}

// ============================================================================
// Required Scopes for Offerings
// ============================================================================

// OfferingType represents a type of marketplace offering
type OfferingType string

const (
	// OfferingTypeBasic represents basic tier offerings
	OfferingTypeBasic OfferingType = "basic"

	// OfferingTypeStandard represents standard tier offerings
	OfferingTypeStandard OfferingType = "standard"

	// OfferingTypePremium represents premium tier offerings
	OfferingTypePremium OfferingType = "premium"

	// OfferingTypeProvider represents provider registration
	OfferingTypeProvider OfferingType = "provider"

	// OfferingTypeValidator represents validator registration
	OfferingTypeValidator OfferingType = "validator"

	// OfferingTypeTEE represents TEE (Trusted Execution Environment) offerings
	OfferingTypeTEE OfferingType = "tee"

	// OfferingTypeHPC represents high-performance computing offerings
	OfferingTypeHPC OfferingType = "hpc"

	// OfferingTypeGPU represents GPU compute offerings
	OfferingTypeGPU OfferingType = "gpu"

	// OfferingTypeCompute represents general compute offerings
	OfferingTypeCompute OfferingType = "compute"

	// OfferingTypeStorage represents storage offerings
	OfferingTypeStorage OfferingType = "storage"
)

// AllOfferingTypes returns all valid offering types
func AllOfferingTypes() []OfferingType {
	return []OfferingType{
		OfferingTypeBasic,
		OfferingTypeStandard,
		OfferingTypePremium,
		OfferingTypeProvider,
		OfferingTypeValidator,
		OfferingTypeTEE,
		OfferingTypeHPC,
		OfferingTypeGPU,
		OfferingTypeCompute,
		OfferingTypeStorage,
	}
}

// IsValidOfferingType checks if an offering type is valid
func IsValidOfferingType(offeringType OfferingType) bool {
	for _, t := range AllOfferingTypes() {
		if t == offeringType {
			return true
		}
	}
	return false
}

// RequiredScopes defines the scopes required for an offering type
type RequiredScopes struct {
	// OfferingType is the type of offering
	OfferingType OfferingType `json:"offering_type"`

	// MinimumScore is the minimum required score
	MinimumScore uint32 `json:"minimum_score"`

	// RequiredScopeTypes are the scope types that must be verified
	RequiredScopeTypes []ScopeType `json:"required_scope_types"`

	// RequiresMFA indicates if MFA is required for this offering type
	RequiresMFA bool `json:"requires_mfa"`

	// Description is a human-readable description
	Description string `json:"description"`
}

// GetRequiredScopesForOffering returns the required scopes for an offering type
// Based on veid-flow-spec.md requirements
func GetRequiredScopesForOffering(offeringType OfferingType) RequiredScopes {
	switch offeringType {
	case OfferingTypeBasic:
		return RequiredScopes{
			OfferingType:       OfferingTypeBasic,
			MinimumScore:       ThresholdBasic,
			RequiredScopeTypes: []ScopeType{ScopeTypeIDDocument, ScopeTypeSelfie},
			RequiresMFA:        false,
			Description:        "Basic marketplace access with limited features",
		}
	case OfferingTypeStandard:
		return RequiredScopes{
			OfferingType:       OfferingTypeStandard,
			MinimumScore:       ThresholdStandard,
			RequiredScopeTypes: []ScopeType{ScopeTypeIDDocument, ScopeTypeSelfie, ScopeTypeEmailProof},
			RequiresMFA:        false,
			Description:        "Standard marketplace access with full ordering capabilities",
		}
	case OfferingTypePremium:
		return RequiredScopes{
			OfferingType:       OfferingTypePremium,
			MinimumScore:       ThresholdPremium,
			RequiredScopeTypes: []ScopeType{ScopeTypeIDDocument, ScopeTypeSelfie, ScopeTypeFaceVideo},
			RequiresMFA:        true,
			Description:        "Premium access with high-value transaction capabilities",
		}
	case OfferingTypeProvider:
		return RequiredScopes{
			OfferingType:       OfferingTypeProvider,
			MinimumScore:       ThresholdStandard,
			RequiredScopeTypes: []ScopeType{ScopeTypeIDDocument, ScopeTypeSelfie, ScopeTypeDomainVerify},
			RequiresMFA:        true,
			Description:        "Service provider registration and offering creation",
		}
	case OfferingTypeValidator:
		return RequiredScopes{
			OfferingType:       OfferingTypeValidator,
			MinimumScore:       ThresholdPremium,
			RequiredScopeTypes: []ScopeType{ScopeTypeIDDocument, ScopeTypeSelfie, ScopeTypeFaceVideo, ScopeTypeDomainVerify},
			RequiresMFA:        true,
			Description:        "Validator node registration and operation",
		}
	default:
		return RequiredScopes{
			OfferingType:       offeringType,
			MinimumScore:       MaxScore,
			RequiredScopeTypes: []ScopeType{},
			RequiresMFA:        true,
			Description:        "Unknown offering type - maximum requirements",
		}
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// ComputeTierFromScoreValue calculates the tier from a score value
// This is aligned with veid-flow-spec.md tier definitions
func ComputeTierFromScoreValue(score uint32, status AccountStatus) int {
	// Non-verified accounts are always Tier 0
	if status != AccountStatusVerified {
		return TierUnverified
	}

	switch {
	case score >= ThresholdPremium:
		return TierPremium
	case score >= ThresholdStandard:
		return TierStandard
	case score >= ThresholdBasic:
		return TierBasic
	default:
		return TierUnverified
	}
}

// TierToString converts a tier number to a string representation
func TierToString(tier int) string {
	switch tier {
	case TierUnverified:
		return "unverified"
	case TierBasic:
		return "basic"
	case TierStandard:
		return "standard"
	case TierPremium:
		return "premium"
	default:
		return "unknown"
	}
}

// TierFromString converts a string to a tier number
func TierFromString(s string) int {
	switch s {
	case "basic":
		return TierBasic
	case "standard":
		return TierStandard
	case "premium":
		return TierPremium
	default:
		return TierUnverified
	}
}

// GetMinimumScoreForTier returns the minimum score required for a tier
func GetMinimumScoreForTier(tier int) uint32 {
	switch tier {
	case TierBasic:
		return ThresholdBasic
	case TierStandard:
		return ThresholdStandard
	case TierPremium:
		return ThresholdPremium
	default:
		return 0
	}
}

// AccountStatusFromVerificationStatus converts VerificationStatus to AccountStatus
func AccountStatusFromVerificationStatus(vs VerificationStatus) AccountStatus {
	switch vs {
	case VerificationStatusUnknown:
		return AccountStatusUnknown
	case VerificationStatusPending:
		return AccountStatusPending
	case VerificationStatusInProgress:
		return AccountStatusInProgress
	case VerificationStatusVerified:
		return AccountStatusVerified
	case VerificationStatusRejected:
		return AccountStatusRejected
	case VerificationStatusExpired:
		return AccountStatusExpired
	default:
		return AccountStatusUnknown
	}
}
