package types

import (
	"time"
)

// Error message constants
const errMsgAccountAddrEmpty = "account address cannot be empty"

// IdentityRecord represents a user's complete identity record on-chain
type IdentityRecord struct {
	// AccountAddress is the blockchain address that owns this identity
	AccountAddress string `json:"account_address"`

	// ScopeRefs are lightweight references to the scopes owned by this identity
	// Full scope data is stored separately for efficiency
	ScopeRefs []ScopeRef `json:"scope_refs"`

	// CurrentScore is the current identity score (0-100)
	// This is computed from verified scopes using the ML pipeline
	CurrentScore uint32 `json:"current_score"`

	// ScoreVersion is the ML model version used to compute the current score
	ScoreVersion string `json:"score_version"`

	// LastVerifiedAt is when the identity was last verified
	LastVerifiedAt *time.Time `json:"last_verified_at,omitempty"`

	// CreatedAt is when this identity record was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when this identity record was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// Tier is the current identity tier based on score and verified scopes
	Tier IdentityTier `json:"tier"`

	// Flags contains any flags on this identity (fraud, suspicious, etc.)
	Flags []string `json:"flags,omitempty"`

	// Locked indicates if the identity is locked (e.g., for security reasons)
	Locked bool `json:"locked"`

	// LockedReason is the reason for locking (if locked)
	LockedReason string `json:"locked_reason,omitempty"`
}

// NewIdentityRecord creates a new identity record
func NewIdentityRecord(accountAddress string, createdAt time.Time) *IdentityRecord {
	return &IdentityRecord{
		AccountAddress: accountAddress,
		ScopeRefs:      make([]ScopeRef, 0),
		CurrentScore:   0,
		ScoreVersion:   "",
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
		Tier:           IdentityTierUnverified,
		Flags:          make([]string, 0),
		Locked:         false,
	}
}

// Validate validates the identity record
func (r *IdentityRecord) Validate() error {
	if r.AccountAddress == "" {
		return ErrInvalidAddress.Wrap(errMsgAccountAddrEmpty)
	}

	if r.CurrentScore > 100 {
		return ErrInvalidScore.Wrap("current score cannot exceed 100")
	}

	if !IsValidIdentityTier(r.Tier) {
		return ErrInvalidTier.Wrapf("invalid tier: %s", r.Tier)
	}

	if r.CreatedAt.IsZero() {
		return ErrInvalidIdentityRecord.Wrap("created_at cannot be zero")
	}

	return nil
}

// AddScopeRef adds a scope reference to the identity record
func (r *IdentityRecord) AddScopeRef(ref ScopeRef) {
	// Check if scope already exists, update if so
	for i, existing := range r.ScopeRefs {
		if existing.ScopeID == ref.ScopeID {
			r.ScopeRefs[i] = ref
			return
		}
	}
	r.ScopeRefs = append(r.ScopeRefs, ref)
}

// RemoveScopeRef removes a scope reference from the identity record
func (r *IdentityRecord) RemoveScopeRef(scopeID string) bool {
	for i, ref := range r.ScopeRefs {
		if ref.ScopeID == scopeID {
			r.ScopeRefs = append(r.ScopeRefs[:i], r.ScopeRefs[i+1:]...)
			return true
		}
	}
	return false
}

// GetScopeRef returns a scope reference by ID
func (r *IdentityRecord) GetScopeRef(scopeID string) (ScopeRef, bool) {
	for _, ref := range r.ScopeRefs {
		if ref.ScopeID == scopeID {
			return ref, true
		}
	}
	return ScopeRef{}, false
}

// GetScopeRefsByType returns all scope references of a specific type
func (r *IdentityRecord) GetScopeRefsByType(scopeType ScopeType) []ScopeRef {
	var refs []ScopeRef
	for _, ref := range r.ScopeRefs {
		if ref.ScopeType == scopeType {
			refs = append(refs, ref)
		}
	}
	return refs
}

// CountVerifiedScopes returns the number of verified scopes
func (r *IdentityRecord) CountVerifiedScopes() int {
	count := 0
	for _, ref := range r.ScopeRefs {
		if ref.Status == VerificationStatusVerified {
			count++
		}
	}
	return count
}

// HasVerifiedScope checks if the identity has a verified scope of a specific type
func (r *IdentityRecord) HasVerifiedScope(scopeType ScopeType) bool {
	for _, ref := range r.ScopeRefs {
		if ref.ScopeType == scopeType && ref.Status == VerificationStatusVerified {
			return true
		}
	}
	return false
}

// UpdateTier updates the identity tier based on current score
func (r *IdentityRecord) UpdateTier() {
	r.Tier = ComputeTierFromScore(r.CurrentScore)
}

// IsActive checks if the identity is active and not locked
func (r *IdentityRecord) IsActive() bool {
	return !r.Locked && r.CurrentScore > 0
}

// IdentityTier represents the verification tier of an identity
type IdentityTier string

// Identity tier constants
const (
	// IdentityTierUnverified is the initial state with no verification
	IdentityTierUnverified IdentityTier = "unverified"

	// IdentityTierBasic is for minimally verified identities (score 50-69)
	IdentityTierBasic IdentityTier = "basic"

	// IdentityTierStandard is for standard verified identities (score 70-84)
	IdentityTierStandard IdentityTier = "standard"

	// IdentityTierVerified is for fully verified identities (score 60-84)
	//
	// Deprecated: Use IdentityTierStandard for score 70-84
	IdentityTierVerified IdentityTier = "verified"

	// IdentityTierTrusted is for highly trusted identities (score 85-100)
	//
	// Deprecated: Use IdentityTierPremium instead
	IdentityTierTrusted IdentityTier = "trusted"

	// IdentityTierPremium is for premium verified identities (score 85-100)
	IdentityTierPremium IdentityTier = "premium"
)

// AllIdentityTiers returns all valid identity tiers
func AllIdentityTiers() []IdentityTier {
	return []IdentityTier{
		IdentityTierUnverified,
		IdentityTierBasic,
		IdentityTierStandard,
		IdentityTierVerified,
		IdentityTierTrusted,
		IdentityTierPremium,
	}
}

// IsValidIdentityTier checks if a tier is valid
func IsValidIdentityTier(tier IdentityTier) bool {
	for _, t := range AllIdentityTiers() {
		if t == tier {
			return true
		}
	}
	return false
}

// ComputeTierFromScore determines the identity tier from a score
func ComputeTierFromScore(score uint32) IdentityTier {
	switch {
	case score == 0:
		return IdentityTierUnverified
	case score < 30:
		return IdentityTierBasic
	case score < 60:
		return IdentityTierStandard
	case score < 85:
		return IdentityTierVerified
	default:
		return IdentityTierTrusted
	}
}

// TierMinimumScore returns the minimum score for a tier
func TierMinimumScore(tier IdentityTier) uint32 {
	switch tier {
	case IdentityTierUnverified:
		return 0
	case IdentityTierBasic:
		return 1
	case IdentityTierStandard:
		return 30
	case IdentityTierVerified:
		return 60
	case IdentityTierTrusted:
		return 85
	default:
		return 0
	}
}

// SimpleIdentityWallet represents a simplified identity wallet for portable identity
// See IdentityWallet in wallet.go for the full on-chain wallet type
type SimpleIdentityWallet struct {
	// WalletID is the unique identifier for this wallet
	WalletID string `json:"wallet_id"`

	// OwnerAddress is the blockchain address that owns this wallet
	OwnerAddress string `json:"owner_address"`

	// IdentityRecord is the associated identity record
	IdentityRecordAddress string `json:"identity_record_address"`

	// PublicKeyFingerprint is the fingerprint of the public key bound to this wallet
	PublicKeyFingerprint string `json:"public_key_fingerprint"`

	// CreatedAt is when this wallet was created
	CreatedAt time.Time `json:"created_at"`

	// Active indicates if this wallet is active
	Active bool `json:"active"`
}

// NewSimpleIdentityWallet creates a new simplified identity wallet
func NewSimpleIdentityWallet(walletID, ownerAddress, identityRecordAddress, pubKeyFingerprint string, createdAt time.Time) *SimpleIdentityWallet {
	return &SimpleIdentityWallet{
		WalletID:              walletID,
		OwnerAddress:          ownerAddress,
		IdentityRecordAddress: identityRecordAddress,
		PublicKeyFingerprint:  pubKeyFingerprint,
		CreatedAt:             createdAt,
		Active:                true,
	}
}

// Validate validates the simplified identity wallet
func (w *SimpleIdentityWallet) Validate() error {
	if w.WalletID == "" {
		return ErrInvalidWallet.Wrap("wallet_id cannot be empty")
	}

	if w.OwnerAddress == "" {
		return ErrInvalidAddress.Wrap("owner address cannot be empty")
	}

	if w.IdentityRecordAddress == "" {
		return ErrInvalidAddress.Wrap("identity record address cannot be empty")
	}

	if w.PublicKeyFingerprint == "" {
		return ErrInvalidWallet.Wrap("public key fingerprint cannot be empty")
	}

	if w.CreatedAt.IsZero() {
		return ErrInvalidWallet.Wrap("created_at cannot be zero")
	}

	return nil
}
