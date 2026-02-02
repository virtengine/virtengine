package types

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
)

// ============================================================================
// Market Integration Types (VE-3028)
// ============================================================================

// MarketType represents different marketplace segments with different VEID requirements
type MarketType string

const (
	// MarketTypeCompute represents compute resource marketplace
	MarketTypeCompute MarketType = "compute"

	// MarketTypeStorage represents storage marketplace
	MarketTypeStorage MarketType = "storage"

	// MarketTypeHPC represents high-performance computing marketplace
	MarketTypeHPC MarketType = "hpc"

	// MarketTypeGPU represents GPU resource marketplace
	MarketTypeGPU MarketType = "gpu"

	// MarketTypeTEE represents trusted execution environment marketplace
	MarketTypeTEE MarketType = "tee"
)

// AllMarketTypes returns all valid market types
func AllMarketTypes() []MarketType {
	return []MarketType{
		MarketTypeCompute,
		MarketTypeStorage,
		MarketTypeHPC,
		MarketTypeGPU,
		MarketTypeTEE,
	}
}

// IsValidMarketType checks if a market type is valid
func IsValidMarketType(marketType MarketType) bool {
	for _, t := range AllMarketTypes() {
		if t == marketType {
			return true
		}
	}
	return false
}

// VEIDLevel represents the required VEID verification level for market participation
type VEIDLevel int

const (
	// VEIDLevelNone indicates no VEID verification required
	VEIDLevelNone VEIDLevel = iota

	// VEIDLevelBasic indicates basic VEID verification (email/SMS)
	VEIDLevelBasic

	// VEIDLevelStandard indicates standard VEID verification (ID document + selfie)
	VEIDLevelStandard

	// VEIDLevelPremium indicates premium VEID verification (full biometric + liveness)
	VEIDLevelPremium

	// VEIDLevelEnterprise indicates enterprise VEID verification (domain + organization)
	VEIDLevelEnterprise
)

// String returns the string representation of a VEIDLevel
func (l VEIDLevel) String() string {
	switch l {
	case VEIDLevelNone:
		return "none"
	case VEIDLevelBasic:
		return string(IdentityTierBasic)
	case VEIDLevelStandard:
		return string(IdentityTierStandard)
	case VEIDLevelPremium:
		return string(IdentityTierPremium)
	case VEIDLevelEnterprise:
		return "enterprise"
	default:
		return "unknown"
	}
}

// IsValid returns true if the level is valid
func (l VEIDLevel) IsValid() bool {
	return l >= VEIDLevelNone && l <= VEIDLevelEnterprise
}

// MinScore returns the minimum identity score required for this level
func (l VEIDLevel) MinScore() uint32 {
	switch l {
	case VEIDLevelNone:
		return 0
	case VEIDLevelBasic:
		return ThresholdBasic // 50
	case VEIDLevelStandard:
		return ThresholdStandard // 70
	case VEIDLevelPremium:
		return ThresholdPremium // 85
	case VEIDLevelEnterprise:
		return ThresholdPremium // 85 (plus additional scope requirements)
	default:
		return 0
	}
}

// ParseVEIDLevel parses a string into a VEIDLevel
func ParseVEIDLevel(s string) (VEIDLevel, error) {
	switch s {
	case "none":
		return VEIDLevelNone, nil
	case string(IdentityTierBasic):
		return VEIDLevelBasic, nil
	case string(IdentityTierStandard):
		return VEIDLevelStandard, nil
	case string(IdentityTierPremium):
		return VEIDLevelPremium, nil
	case "enterprise":
		return VEIDLevelEnterprise, nil
	default:
		return VEIDLevelNone, errorsmod.Wrapf(ErrInvalidParams, "invalid VEID level: %s", s)
	}
}

// ============================================================================
// Market VEID Requirements
// ============================================================================

// MarketVEIDRequirements defines the VEID requirements for marketplace participation
type MarketVEIDRequirements struct {
	// MarketType is the type of marketplace these requirements apply to
	MarketType MarketType `json:"market_type"`

	// MinTrustScore is the minimum identity score required (0-100)
	MinTrustScore sdkmath.LegacyDec `json:"min_trust_score"`

	// RequiredScopes are the scope types that must be verified
	RequiredScopes []ScopeType `json:"required_scopes"`

	// RequiredLevels maps scope types to their required verification levels
	RequiredLevels map[ScopeType]VerificationStatus `json:"required_levels"`

	// AllowDelegation indicates if delegated identities can participate
	AllowDelegation bool `json:"allow_delegation"`

	// MaxDelegationAge is the maximum age of a delegation for it to be valid
	MaxDelegationAge time.Duration `json:"max_delegation_age"`

	// RequireActiveIdentity indicates if the identity must be non-expired
	RequireActiveIdentity bool `json:"require_active_identity"`

	// RequireUnlockedIdentity indicates if the identity must not be locked
	RequireUnlockedIdentity bool `json:"require_unlocked_identity"`

	// RequiresMFA indicates if multi-factor authentication is required
	RequiresMFA bool `json:"requires_mfa"`

	// ProviderRequirements are additional requirements for providers (empty = same as tenant)
	ProviderRequirements *ProviderVEIDRequirements `json:"provider_requirements,omitempty"`

	// CreatedAt is when these requirements were created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when these requirements were last updated
	UpdatedAt time.Time `json:"updated_at"`

	// Authority is the address that set these requirements
	Authority string `json:"authority"`
}

// ProviderVEIDRequirements defines additional VEID requirements for providers
type ProviderVEIDRequirements struct {
	// MinTrustScore is the minimum identity score required for providers
	MinTrustScore sdkmath.LegacyDec `json:"min_trust_score"`

	// RequiredScopes are additional scope types that providers must have verified
	RequiredScopes []ScopeType `json:"required_scopes"`

	// RequireDomainVerification indicates if domain verification is required
	RequireDomainVerification bool `json:"require_domain_verification"`

	// RequireActiveStake indicates if the provider must have an active stake
	RequireActiveStake bool `json:"require_active_stake"`
}

// NewMarketVEIDRequirements creates a new MarketVEIDRequirements with defaults
func NewMarketVEIDRequirements(marketType MarketType, now time.Time) *MarketVEIDRequirements {
	return &MarketVEIDRequirements{
		MarketType:              marketType,
		MinTrustScore:           sdkmath.LegacyNewDec(int64(ThresholdBasic)),
		RequiredScopes:          []ScopeType{},
		RequiredLevels:          make(map[ScopeType]VerificationStatus),
		AllowDelegation:         true,
		MaxDelegationAge:        30 * 24 * time.Hour, // 30 days
		RequireActiveIdentity:   true,
		RequireUnlockedIdentity: true,
		CreatedAt:               now,
		UpdatedAt:               now,
	}
}

// Validate validates the requirements
func (r *MarketVEIDRequirements) Validate() error {
	if !IsValidMarketType(r.MarketType) {
		return errorsmod.Wrapf(ErrInvalidParams, "invalid market type: %s", r.MarketType)
	}

	if r.MinTrustScore.IsNegative() || r.MinTrustScore.GT(sdkmath.LegacyNewDec(100)) {
		return errorsmod.Wrap(ErrInvalidScore, "min trust score must be between 0 and 100")
	}

	for _, scope := range r.RequiredScopes {
		if !IsValidScopeType(scope) {
			return errorsmod.Wrapf(ErrInvalidScopeType, "invalid scope type: %s", scope)
		}
	}

	for scope, status := range r.RequiredLevels {
		if !IsValidScopeType(scope) {
			return errorsmod.Wrapf(ErrInvalidScopeType, "invalid scope type in required levels: %s", scope)
		}
		if !IsValidVerificationStatus(status) {
			return errorsmod.Wrapf(ErrInvalidVerificationStatus, "invalid status for scope %s: %s", scope, status)
		}
	}

	if r.MaxDelegationAge < 0 {
		return errorsmod.Wrap(ErrInvalidParams, "max delegation age cannot be negative")
	}

	if r.CreatedAt.IsZero() {
		return errorsmod.Wrap(ErrInvalidParams, "created_at cannot be zero")
	}

	return nil
}

// ============================================================================
// Participant VEID Status
// ============================================================================

// ParticipantVEIDStatus represents the VEID status of a marketplace participant
type ParticipantVEIDStatus struct {
	// Address is the participant's blockchain address
	Address string `json:"address"`

	// IsVerified indicates if the participant has any verified identity
	IsVerified bool `json:"is_verified"`

	// TrustScore is the participant's current identity score
	TrustScore sdkmath.LegacyDec `json:"trust_score"`

	// Tier is the participant's current identity tier
	Tier IdentityTier `json:"tier"`

	// MeetsRequirements indicates if the participant meets all requirements
	MeetsRequirements bool `json:"meets_requirements"`

	// MissingScopes lists scope types that are required but not verified
	MissingScopes []ScopeType `json:"missing_scopes,omitempty"`

	// VerifiedScopes lists scope types that are verified
	VerifiedScopes []ScopeType `json:"verified_scopes,omitempty"`

	// IsLocked indicates if the identity is locked
	IsLocked bool `json:"is_locked"`

	// IsExpired indicates if the identity verification has expired
	IsExpired bool `json:"is_expired"`

	// IsDelegated indicates if the participant is using delegated identity
	IsDelegated bool `json:"is_delegated"`

	// DelegatorAddress is the address of the delegator (if delegated)
	DelegatorAddress string `json:"delegator_address,omitempty"`

	// DelegationID is the ID of the delegation being used (if delegated)
	DelegationID string `json:"delegation_id,omitempty"`

	// LastVerifiedAt is when the identity was last verified
	LastVerifiedAt *time.Time `json:"last_verified_at,omitempty"`

	// EligibilityReason provides a human-readable explanation of eligibility status
	EligibilityReason string `json:"eligibility_reason,omitempty"`
}

// NewParticipantVEIDStatus creates a new empty ParticipantVEIDStatus
func NewParticipantVEIDStatus(address string) *ParticipantVEIDStatus {
	return &ParticipantVEIDStatus{
		Address:        address,
		TrustScore:     sdkmath.LegacyZeroDec(),
		Tier:           IdentityTierUnverified,
		MissingScopes:  []ScopeType{},
		VerifiedScopes: []ScopeType{},
	}
}

// ============================================================================
// Market Eligibility Results
// ============================================================================

// MarketEligibilityResult represents the result of a market eligibility check
type MarketEligibilityResult struct {
	// Eligible indicates if the participant is eligible
	Eligible bool `json:"eligible"`

	// Reason provides a human-readable explanation
	Reason string `json:"reason"`

	// ParticipantStatus contains detailed status information
	ParticipantStatus *ParticipantVEIDStatus `json:"participant_status,omitempty"`

	// CheckedAt is when the eligibility was checked
	CheckedAt time.Time `json:"checked_at"`

	// Requirements are the requirements that were checked against
	Requirements *MarketVEIDRequirements `json:"requirements,omitempty"`

	// ValidationErrors lists specific validation failures
	ValidationErrors []string `json:"validation_errors,omitempty"`
}

// NewMarketEligibilityResult creates a new MarketEligibilityResult
func NewMarketEligibilityResult(eligible bool, reason string, checkedAt time.Time) *MarketEligibilityResult {
	return &MarketEligibilityResult{
		Eligible:         eligible,
		Reason:           reason,
		CheckedAt:        checkedAt,
		ValidationErrors: []string{},
	}
}

// AddValidationError adds a validation error to the result
func (r *MarketEligibilityResult) AddValidationError(err string) {
	r.Eligible = false
	r.ValidationErrors = append(r.ValidationErrors, err)
}

// ============================================================================
// Market Integration Errors (VE-3028)
// Error codes: 1160-1169
// ============================================================================

var (
	// ErrMarketVEIDNotMet is returned when VEID requirements are not met
	ErrMarketVEIDNotMet = errorsmod.Register(ModuleName, 1160, "market VEID requirements not met")

	// ErrMarketRequirementsNotFound is returned when market requirements are not found
	ErrMarketRequirementsNotFound = errorsmod.Register(ModuleName, 1161, "market VEID requirements not found")

	// ErrInsufficientTrustScore is returned when trust score is too low
	ErrInsufficientTrustScore = errorsmod.Register(ModuleName, 1162, "insufficient trust score for market participation")

	// ErrMissingScopesForMarket is returned when required scopes are missing
	ErrMissingScopesForMarket = errorsmod.Register(ModuleName, 1163, "missing required verification scopes for market")

	// ErrDelegationNotAllowed is returned when delegation is not allowed for market
	ErrDelegationNotAllowed = errorsmod.Register(ModuleName, 1164, "delegation not allowed for this market type")

	// ErrDelegationTooOld is returned when the delegation is too old
	ErrDelegationTooOld = errorsmod.Register(ModuleName, 1165, "delegation exceeds maximum age for market participation")

	// ErrIdentityExpiredForMarket is returned when identity is expired
	ErrIdentityExpiredForMarket = errorsmod.Register(ModuleName, 1166, "identity verification expired for market participation")

	// ErrIdentityLockedForMarket is returned when identity is locked
	ErrIdentityLockedForMarket = errorsmod.Register(ModuleName, 1167, "identity is locked and cannot participate in market")

	// ErrProviderVEIDNotMet is returned when provider-specific VEID requirements are not met
	ErrProviderVEIDNotMet = errorsmod.Register(ModuleName, 1168, "provider VEID requirements not met")
)

// ============================================================================
// Store Key Prefixes for Market Integration
// ============================================================================

var (
	// PrefixMarketVEIDRequirements is the prefix for market VEID requirements storage
	// Key: PrefixMarketVEIDRequirements | market_type -> MarketVEIDRequirements
	PrefixMarketVEIDRequirements = []byte{0x40}

	// PrefixMarketParticipantStatus is the prefix for cached participant status
	// Key: PrefixMarketParticipantStatus | address | market_type -> ParticipantVEIDStatus
	PrefixMarketParticipantStatus = []byte{0x41}
)

// MarketRequirementsKey returns the store key for market requirements
func MarketRequirementsKey(marketType MarketType) []byte {
	return append(PrefixMarketVEIDRequirements, []byte(marketType)...)
}

// MarketParticipantStatusKey returns the store key for participant status
func MarketParticipantStatusKey(address string, marketType MarketType) []byte {
	key := make([]byte, 0, len(PrefixMarketParticipantStatus)+len(address)+1+len(marketType))
	key = append(key, PrefixMarketParticipantStatus...)
	key = append(key, []byte(address)...)
	key = append(key, byte('/'))
	return append(key, []byte(marketType)...)
}
