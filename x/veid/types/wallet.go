package types

import (
	"bytes"
	"crypto/sha256"
	"time"
)

// ============================================================================
// Identity Wallet Types
// ============================================================================

// WalletStatus represents the overall status of an identity wallet
type WalletStatus string

const (
	// WalletStatusActive indicates the wallet is active and usable
	WalletStatusActive WalletStatus = "active"

	// WalletStatusSuspended indicates the wallet is temporarily suspended
	WalletStatusSuspended WalletStatus = "suspended"

	// WalletStatusRevoked indicates the wallet has been revoked
	WalletStatusRevoked WalletStatus = "revoked"

	// WalletStatusExpired indicates the wallet verification has expired
	WalletStatusExpired WalletStatus = "expired"
)

// AllWalletStatuses returns all valid wallet statuses
func AllWalletStatuses() []WalletStatus {
	return []WalletStatus{
		WalletStatusActive,
		WalletStatusSuspended,
		WalletStatusRevoked,
		WalletStatusExpired,
	}
}

// IsValidWalletStatus checks if a wallet status is valid
func IsValidWalletStatus(status WalletStatus) bool {
	for _, s := range AllWalletStatuses() {
		if s == status {
			return true
		}
	}
	return false
}

// ScopeRefStatus represents the status of a scope reference within a wallet
type ScopeRefStatus string

const (
	// ScopeRefStatusActive indicates the scope reference is active
	ScopeRefStatusActive ScopeRefStatus = "active"

	// ScopeRefStatusRevoked indicates the scope reference has been revoked
	ScopeRefStatusRevoked ScopeRefStatus = "revoked"

	// ScopeRefStatusExpired indicates the scope reference has expired
	ScopeRefStatusExpired ScopeRefStatus = "expired"

	// ScopeRefStatusPending indicates the scope is pending verification
	ScopeRefStatusPending ScopeRefStatus = "pending"
)

// AllScopeRefStatuses returns all valid scope reference statuses
func AllScopeRefStatuses() []ScopeRefStatus {
	return []ScopeRefStatus{
		ScopeRefStatusActive,
		ScopeRefStatusRevoked,
		ScopeRefStatusExpired,
		ScopeRefStatusPending,
	}
}

// IsValidScopeRefStatus checks if a scope reference status is valid
func IsValidScopeRefStatus(status ScopeRefStatus) bool {
	for _, s := range AllScopeRefStatuses() {
		if s == status {
			return true
		}
	}
	return false
}

// VerificationHistoryEntry represents a single verification event in the wallet's history
type VerificationHistoryEntry struct {
	// EntryID is a unique identifier for this history entry
	EntryID string `json:"entry_id"`

	// Timestamp is when this verification occurred
	Timestamp time.Time `json:"timestamp"`

	// BlockHeight is the block height when this was recorded
	BlockHeight int64 `json:"block_height"`

	// PreviousScore is the score before this verification
	PreviousScore uint32 `json:"previous_score"`

	// NewScore is the score after this verification
	NewScore uint32 `json:"new_score"`

	// PreviousStatus is the status before this verification
	PreviousStatus AccountStatus `json:"previous_status"`

	// NewStatus is the status after this verification
	NewStatus AccountStatus `json:"new_status"`

	// ScopesEvaluated lists the scope IDs that were evaluated
	ScopesEvaluated []string `json:"scopes_evaluated,omitempty"`

	// ModelVersion is the ML model version used for this verification
	ModelVersion string `json:"model_version"`

	// ValidatorAddress is the address of the validator that performed this verification
	ValidatorAddress string `json:"validator_address,omitempty"`

	// Reason is an optional reason/description for this verification
	Reason string `json:"reason,omitempty"`
}

// NewVerificationHistoryEntry creates a new verification history entry
func NewVerificationHistoryEntry(
	entryID string,
	timestamp time.Time,
	blockHeight int64,
	previousScore, newScore uint32,
	previousStatus, newStatus AccountStatus,
	modelVersion string,
) *VerificationHistoryEntry {
	return &VerificationHistoryEntry{
		EntryID:        entryID,
		Timestamp:      timestamp,
		BlockHeight:    blockHeight,
		PreviousScore:  previousScore,
		NewScore:       newScore,
		PreviousStatus: previousStatus,
		NewStatus:      newStatus,
		ModelVersion:   modelVersion,
	}
}

// ScopeReference represents a reference to an encrypted scope within the wallet
type ScopeReference struct {
	// ScopeID is the unique identifier of the scope
	ScopeID string `json:"scope_id"`

	// ScopeType indicates what kind of identity data this scope contains
	ScopeType ScopeType `json:"scope_type"`

	// EnvelopeHash is the SHA-256 hash of the encrypted envelope
	// This allows verification without exposing the encrypted content
	EnvelopeHash []byte `json:"envelope_hash"`

	// AddedAt is when this scope was added to the wallet
	AddedAt time.Time `json:"added_at"`

	// Status is the current status of this scope reference
	Status ScopeRefStatus `json:"status"`

	// ConsentGranted indicates if consent has been granted for this scope
	ConsentGranted bool `json:"consent_granted"`

	// RevocationReason is the reason for revocation (if revoked)
	RevocationReason string `json:"revocation_reason,omitempty"`

	// RevokedAt is when this scope was revoked (if revoked)
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	// ExpiresAt is when this scope reference expires (optional)
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// NewScopeReference creates a new scope reference
func NewScopeReference(
	scopeID string,
	scopeType ScopeType,
	envelopeHash []byte,
	addedAt time.Time,
) *ScopeReference {
	return &ScopeReference{
		ScopeID:        scopeID,
		ScopeType:      scopeType,
		EnvelopeHash:   envelopeHash,
		AddedAt:        addedAt,
		Status:         ScopeRefStatusPending,
		ConsentGranted: false,
	}
}

// Validate validates the scope reference
func (sr *ScopeReference) Validate() error {
	if sr.ScopeID == "" {
		return ErrInvalidScope.Wrap("scope_id cannot be empty")
	}

	if !IsValidScopeType(sr.ScopeType) {
		return ErrInvalidScopeType.Wrapf("invalid scope type: %s", sr.ScopeType)
	}

	if len(sr.EnvelopeHash) == 0 {
		return ErrInvalidPayloadHash.Wrap("envelope_hash cannot be empty")
	}

	if sr.AddedAt.IsZero() {
		return ErrInvalidScope.Wrap("added_at cannot be zero")
	}

	if !IsValidScopeRefStatus(sr.Status) {
		return ErrInvalidScope.Wrapf("invalid status: %s", sr.Status)
	}

	return nil
}

// IsActive checks if the scope reference is active
func (sr *ScopeReference) IsActive() bool {
	if sr.Status != ScopeRefStatusActive {
		return false
	}

	if sr.ExpiresAt != nil && time.Now().After(*sr.ExpiresAt) {
		return false
	}

	return true
}

// IdentityWallet represents a user-controlled identity container
// This is the first-class on-chain identity primitive that references
// encrypted scopes and derived features, bound to the user's account key(s).
type IdentityWallet struct {
	// WalletID is the unique identifier for this wallet
	// Typically derived from the account address
	WalletID string `json:"wallet_id"`

	// AccountAddress is the blockchain address bound to this wallet
	AccountAddress string `json:"account_address"`

	// CreatedAt is when this wallet was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when this wallet was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// Status is the current status of the wallet
	Status WalletStatus `json:"status"`

	// ScopeRefs are references to encrypted scope envelopes
	// These are opaque references - the actual encrypted data is stored separately
	ScopeRefs []ScopeReference `json:"scope_refs"`

	// DerivedFeatures contains hashes of derived features for verification matching
	DerivedFeatures DerivedFeatures `json:"derived_features"`

	// CurrentScore is the current identity verification score (0-100)
	CurrentScore uint32 `json:"current_score"`

	// ScoreStatus is the current verification status
	ScoreStatus AccountStatus `json:"score_status"`

	// VerificationHistory contains the history of verification events
	VerificationHistory []VerificationHistoryEntry `json:"verification_history"`

	// ConsentSettings contains the consent configuration for this wallet
	ConsentSettings ConsentSettings `json:"consent_settings"`

	// BindingSignature is the user's signature over (WalletID + AccountAddress)
	// This cryptographically binds the wallet to the user's account
	BindingSignature []byte `json:"binding_signature"`

	// BindingPubKey is the public key used to create the binding signature
	// This is captured at wallet creation and updated on key rotation
	BindingPubKey []byte `json:"binding_pub_key"`

	// LastBindingAt is when the wallet was last bound/rebound
	LastBindingAt time.Time `json:"last_binding_at"`

	// Tier is the current identity tier based on score
	Tier IdentityTier `json:"tier"`

	// Metadata contains additional wallet metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewIdentityWallet creates a new identity wallet
func NewIdentityWallet(
	walletID string,
	accountAddress string,
	createdAt time.Time,
	bindingSignature []byte,
	bindingPubKey []byte,
) *IdentityWallet {
	return &IdentityWallet{
		WalletID:            walletID,
		AccountAddress:      accountAddress,
		CreatedAt:           createdAt,
		UpdatedAt:           createdAt,
		Status:              WalletStatusActive,
		ScopeRefs:           make([]ScopeReference, 0),
		DerivedFeatures:     NewDerivedFeatures(),
		CurrentScore:        0,
		ScoreStatus:         AccountStatusUnknown,
		VerificationHistory: make([]VerificationHistoryEntry, 0),
		ConsentSettings:     NewConsentSettings(),
		BindingSignature:    bindingSignature,
		BindingPubKey:       bindingPubKey,
		LastBindingAt:       createdAt,
		Tier:                IdentityTierUnverified,
		Metadata:            make(map[string]string),
	}
}

// Validate validates the identity wallet
func (w *IdentityWallet) Validate() error {
	if w.WalletID == "" {
		return ErrInvalidWallet.Wrap("wallet_id cannot be empty")
	}

	if w.AccountAddress == "" {
		return ErrInvalidAddress.Wrap(errMsgAccountAddrEmpty)
	}

	if !IsValidWalletStatus(w.Status) {
		return ErrInvalidWallet.Wrapf("invalid wallet status: %s", w.Status)
	}

	if w.CreatedAt.IsZero() {
		return ErrInvalidWallet.Wrap("created_at cannot be zero")
	}

	if w.CurrentScore > MaxScore {
		return ErrInvalidScore.Wrapf("score %d exceeds maximum %d", w.CurrentScore, MaxScore)
	}

	if len(w.BindingSignature) == 0 {
		return ErrInvalidWallet.Wrap("binding_signature cannot be empty")
	}

	if len(w.BindingPubKey) == 0 {
		return ErrInvalidWallet.Wrap("binding_pub_key cannot be empty")
	}

	// Validate scope references
	for i, sr := range w.ScopeRefs {
		if err := sr.Validate(); err != nil {
			return ErrInvalidWallet.Wrapf("invalid scope reference at index %d: %v", i, err)
		}
	}

	// Validate derived features
	if err := w.DerivedFeatures.Validate(); err != nil {
		return ErrInvalidWallet.Wrapf("invalid derived features: %v", err)
	}

	// Validate consent settings
	if err := w.ConsentSettings.Validate(); err != nil {
		return ErrInvalidWallet.Wrapf("invalid consent settings: %v", err)
	}

	return nil
}

// GetBindingMessage returns the message that should be signed for wallet binding
func (w *IdentityWallet) GetBindingMessage() []byte {
	return GetWalletBindingMessage(w.WalletID, w.AccountAddress)
}

// GetWalletBindingMessage returns the canonical message for wallet binding signature
func GetWalletBindingMessage(walletID, accountAddress string) []byte {
	msg := []byte("VEID_WALLET_BINDING:" + walletID + ":" + accountAddress)
	hash := sha256.Sum256(msg)
	return hash[:]
}

// AddScopeReference adds a scope reference to the wallet
func (w *IdentityWallet) AddScopeReference(ref ScopeReference) {
	// Check if scope already exists, update if so
	for i, existing := range w.ScopeRefs {
		if existing.ScopeID == ref.ScopeID {
			w.ScopeRefs[i] = ref
			w.UpdatedAt = time.Now()
			return
		}
	}
	w.ScopeRefs = append(w.ScopeRefs, ref)
	w.UpdatedAt = time.Now()
}

// GetScopeReference returns a scope reference by ID
func (w *IdentityWallet) GetScopeReference(scopeID string) (ScopeReference, bool) {
	for _, ref := range w.ScopeRefs {
		if ref.ScopeID == scopeID {
			return ref, true
		}
	}
	return ScopeReference{}, false
}

// RevokeScopeReference revokes a scope reference
func (w *IdentityWallet) RevokeScopeReference(scopeID string, reason string, revokedAt time.Time) bool {
	for i, ref := range w.ScopeRefs {
		if ref.ScopeID == scopeID {
			w.ScopeRefs[i].Status = ScopeRefStatusRevoked
			w.ScopeRefs[i].RevocationReason = reason
			w.ScopeRefs[i].RevokedAt = &revokedAt
			w.ScopeRefs[i].ConsentGranted = false
			w.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// GetActiveScopeRefs returns all active scope references
func (w *IdentityWallet) GetActiveScopeRefs() []ScopeReference {
	var active []ScopeReference
	for _, ref := range w.ScopeRefs {
		if ref.IsActive() {
			active = append(active, ref)
		}
	}
	return active
}

// GetScopeRefsByType returns all scope references of a specific type
func (w *IdentityWallet) GetScopeRefsByType(scopeType ScopeType) []ScopeReference {
	var refs []ScopeReference
	for _, ref := range w.ScopeRefs {
		if ref.ScopeType == scopeType {
			refs = append(refs, ref)
		}
	}
	return refs
}

// AddVerificationHistoryEntry adds a verification event to the history
func (w *IdentityWallet) AddVerificationHistoryEntry(entry VerificationHistoryEntry) {
	w.VerificationHistory = append(w.VerificationHistory, entry)
	w.UpdatedAt = time.Now()
}

// UpdateScore updates the wallet's current score and status
func (w *IdentityWallet) UpdateScore(score uint32, status AccountStatus) {
	w.CurrentScore = score
	w.ScoreStatus = status
	w.Tier = TierFromScore(score, status)
	w.UpdatedAt = time.Now()
}

// TierFromScore determines the identity tier from score and status
func TierFromScore(score uint32, status AccountStatus) IdentityTier {
	if status != AccountStatusVerified {
		return IdentityTierUnverified
	}

	switch {
	case score >= ThresholdPremium:
		return IdentityTierPremium
	case score >= ThresholdStandard:
		return IdentityTierStandard
	case score >= ThresholdBasic:
		return IdentityTierBasic
	default:
		return IdentityTierUnverified
	}
}

// Rebind rebinds the wallet with a new signature and public key
// This is used during key rotation
func (w *IdentityWallet) Rebind(newSignature, newPubKey []byte, rebindAt time.Time) {
	w.BindingSignature = newSignature
	w.BindingPubKey = newPubKey
	w.LastBindingAt = rebindAt
	w.UpdatedAt = rebindAt
}

// IsActive checks if the wallet is active
func (w *IdentityWallet) IsActive() bool {
	return w.Status == WalletStatusActive
}

// VerifyBinding verifies that a signature matches the wallet binding
func (w *IdentityWallet) VerifyBinding(pubKey, signature []byte) bool {
	return bytes.Equal(w.BindingPubKey, pubKey) && bytes.Equal(w.BindingSignature, signature)
}

// PublicWalletInfo represents non-sensitive public metadata about a wallet
// This is returned by query endpoints without exposing encrypted data
type PublicWalletInfo struct {
	// WalletID is the wallet identifier
	WalletID string `json:"wallet_id"`

	// AccountAddress is the bound account
	AccountAddress string `json:"account_address"`

	// Status is the wallet status
	Status WalletStatus `json:"status"`

	// CreatedAt is when the wallet was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the wallet was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// CurrentScore is the verification score
	CurrentScore uint32 `json:"current_score"`

	// ScoreStatus is the verification status
	ScoreStatus AccountStatus `json:"score_status"`

	// Tier is the identity tier
	Tier IdentityTier `json:"tier"`

	// ScopeCount is the number of scopes in the wallet
	ScopeCount int `json:"scope_count"`

	// ActiveScopeCount is the number of active scopes
	ActiveScopeCount int `json:"actiVIRTENGINE_scope_count"`

	// LastVerificationAt is when the wallet was last verified
	LastVerificationAt *time.Time `json:"last_verification_at,omitempty"`

	// ConsentShareWithProviders indicates if sharing with providers is enabled
	ConsentShareWithProviders bool `json:"consent_share_with_providers"`

	// ConsentShareForVerification indicates if sharing for verification is enabled
	ConsentShareForVerification bool `json:"consent_share_for_verification"`
}

// ToPublicInfo converts an IdentityWallet to PublicWalletInfo
func (w *IdentityWallet) ToPublicInfo() PublicWalletInfo {
	var lastVerification *time.Time
	if len(w.VerificationHistory) > 0 {
		lastVerification = &w.VerificationHistory[len(w.VerificationHistory)-1].Timestamp
	}

	return PublicWalletInfo{
		WalletID:                    w.WalletID,
		AccountAddress:              w.AccountAddress,
		Status:                      w.Status,
		CreatedAt:                   w.CreatedAt,
		UpdatedAt:                   w.UpdatedAt,
		CurrentScore:                w.CurrentScore,
		ScoreStatus:                 w.ScoreStatus,
		Tier:                        w.Tier,
		ScopeCount:                  len(w.ScopeRefs),
		ActiveScopeCount:            len(w.GetActiveScopeRefs()),
		LastVerificationAt:          lastVerification,
		ConsentShareWithProviders:   w.ConsentSettings.ShareWithProviders,
		ConsentShareForVerification: w.ConsentSettings.ShareForVerification,
	}
}

// WalletScopeInfo represents non-sensitive scope information
type WalletScopeInfo struct {
	// ScopeID is the scope identifier
	ScopeID string `json:"scope_id"`

	// ScopeType is the type of scope
	ScopeType ScopeType `json:"scope_type"`

	// Status is the scope status
	Status ScopeRefStatus `json:"status"`

	// AddedAt is when the scope was added
	AddedAt time.Time `json:"added_at"`

	// ConsentGranted indicates if consent is granted
	ConsentGranted bool `json:"consent_granted"`

	// ExpiresAt is when the scope expires (if applicable)
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// GetScopesPublicInfo returns non-sensitive info about all scopes
func (w *IdentityWallet) GetScopesPublicInfo() []WalletScopeInfo {
	infos := make([]WalletScopeInfo, 0, len(w.ScopeRefs))
	for _, ref := range w.ScopeRefs {
		infos = append(infos, WalletScopeInfo{
			ScopeID:        ref.ScopeID,
			ScopeType:      ref.ScopeType,
			Status:         ref.Status,
			AddedAt:        ref.AddedAt,
			ConsentGranted: ref.ConsentGranted,
			ExpiresAt:      ref.ExpiresAt,
		})
	}
	return infos
}
