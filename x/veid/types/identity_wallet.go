// Package types contains the Identity Wallet on-chain primitive.
// This file provides the unified Identity Wallet API as specified in VE-209.
//
// The Identity Wallet is the first-class on-chain identity primitive that:
// - References encrypted scope envelopes (not the raw data)
// - Stores derived feature hashes for verification matching
// - Maintains verification history
// - Tracks current score and status
// - Provides scope-level consent toggles and revocation
//
// All wallet operations require cryptographic binding via signatures.
package types

import (
	"crypto/sha256"
	"time"
)

// ============================================================================
// Identity Wallet API (VE-209 Specification)
// ============================================================================

// This file defines the public API for the Identity Wallet primitive.
// The actual implementation is split across:
// - wallet.go: Core types (IdentityWallet, ScopeReference, WalletStatus, etc.)
// - consent.go: ConsentSettings and ScopeConsent types
// - derived_features.go: DerivedFeatures type and hash utilities
// - wallet_msgs.go: Cosmos SDK message types for wallet operations
// - wallet_query.go: Query request/response types

// ============================================================================
// Factory Functions for Identity Wallet Creation
// ============================================================================

// CreateIdentityWalletParams contains parameters for creating a new identity wallet.
// This is used by the keeper's CreateIdentityWallet method.
type CreateIdentityWalletParams struct {
	// AccountAddress is the blockchain address to bind the wallet to
	AccountAddress string

	// BindingSignature is the signature proving ownership of the account
	// Signs: SHA256("VEID_WALLET_BINDING:" + wallet_id + ":" + account_address)
	BindingSignature []byte

	// BindingPubKey is the public key used for the binding signature
	BindingPubKey []byte

	// InitialConsent contains optional initial consent settings
	InitialConsent *ConsentSettings

	// Metadata contains optional wallet metadata
	Metadata map[string]string
}

// Validate validates the create wallet parameters
func (p *CreateIdentityWalletParams) Validate() error {
	if p.AccountAddress == "" {
		return ErrInvalidAddress.Wrap(errMsgAccountAddrEmpty)
	}

	if len(p.BindingSignature) == 0 {
		return ErrInvalidUserSignature.Wrap("binding_signature cannot be empty")
	}

	if len(p.BindingPubKey) == 0 {
		return ErrInvalidWallet.Wrap("binding_pub_key cannot be empty")
	}

	if p.InitialConsent != nil {
		if err := p.InitialConsent.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ============================================================================
// Identity Wallet Update Operations
// ============================================================================

// UpdateIdentityWalletParams contains parameters for updating an identity wallet.
type UpdateIdentityWalletParams struct {
	// AccountAddress is the wallet's bound account address
	AccountAddress string

	// Score is the new identity score (optional, nil to keep existing)
	Score *uint32

	// Status is the new account status (optional, nil to keep existing)
	Status *AccountStatus

	// ModelVersion is the ML model version used for scoring
	ModelVersion string

	// ValidatorAddress is the validator performing the update
	ValidatorAddress string

	// ScopesEvaluated lists scopes that were evaluated
	ScopesEvaluated []string

	// Reason describes why the update occurred
	Reason string
}

// Validate validates the update wallet parameters
func (p *UpdateIdentityWalletParams) Validate() error {
	if p.AccountAddress == "" {
		return ErrInvalidAddress.Wrap(errMsgAccountAddrEmpty)
	}

	if p.Score != nil && *p.Score > MaxScore {
		return ErrInvalidScore.Wrapf("score %d exceeds maximum %d", *p.Score, MaxScore)
	}

	if p.Status != nil && !IsValidAccountStatus(*p.Status) {
		return ErrInvalidVerificationStatus.Wrapf("invalid status: %s", *p.Status)
	}

	return nil
}

// ============================================================================
// Scope Revocation Parameters
// ============================================================================

// RevokeScopeParams contains parameters for revoking a scope from a wallet.
type RevokeScopeParams struct {
	// AccountAddress is the wallet's bound account address
	AccountAddress string

	// ScopeID is the unique identifier of the scope to revoke
	ScopeID string

	// Reason is the reason for revocation
	Reason string

	// UserSignature authorizes the revocation
	// Signs: SHA256("VEID_REVOKE_SCOPE:" + account_address + ":" + scope_id)
	UserSignature []byte

	// RevokeConsent also revokes consent for this scope (default: true)
	RevokeConsent bool
}

// Validate validates the revoke scope parameters
func (p *RevokeScopeParams) Validate() error {
	if p.AccountAddress == "" {
		return ErrInvalidAddress.Wrap(errMsgAccountAddrEmpty)
	}

	if p.ScopeID == "" {
		return ErrInvalidScope.Wrap(errMsgScopeIDEmpty)
	}

	if len(p.UserSignature) == 0 {
		return ErrInvalidUserSignature.Wrap("user_signature cannot be empty")
	}

	return nil
}

// ============================================================================
// Consent Toggle Parameters
// ============================================================================

// ToggleScopeConsentParams contains parameters for toggling consent on a scope.
type ToggleScopeConsentParams struct {
	// AccountAddress is the wallet's bound account address
	AccountAddress string

	// ScopeID is the scope to toggle consent for
	ScopeID string

	// GrantConsent indicates whether to grant (true) or revoke (false) consent
	GrantConsent bool

	// Purpose describes why consent is being granted (required when granting)
	Purpose string

	// ExpiresAt is when consent expires (optional)
	ExpiresAt *time.Time

	// GrantedToProviders lists specific providers to grant consent to (optional)
	GrantedToProviders []string

	// UserSignature authorizes the consent change
	UserSignature []byte
}

// Validate validates the toggle consent parameters
func (p *ToggleScopeConsentParams) Validate() error {
	if p.AccountAddress == "" {
		return ErrInvalidAddress.Wrap(errMsgAccountAddrEmpty)
	}

	if p.ScopeID == "" {
		return ErrInvalidScope.Wrap(errMsgScopeIDEmpty)
	}

	if len(p.UserSignature) == 0 {
		return ErrInvalidUserSignature.Wrap("user_signature cannot be empty")
	}

	if p.GrantConsent && p.Purpose == "" {
		return ErrInvalidWallet.Wrap("purpose is required when granting consent")
	}

	return nil
}

// ============================================================================
// Identity Wallet Query Helpers
// ============================================================================

// WalletQueryFilter defines filters for wallet queries.
type WalletQueryFilter struct {
	// Status filters by wallet status
	Status *WalletStatus

	// MinScore filters by minimum score
	MinScore *uint32

	// MaxScore filters by maximum score
	MaxScore *uint32

	// Tier filters by identity tier
	Tier *IdentityTier

	// HasActiveScopes filters to wallets with active scopes
	HasActiveScopes *bool

	// ScopeType filters to wallets with specific scope types
	ScopeType *ScopeType
}

// Matches checks if a wallet matches the filter criteria.
func (f *WalletQueryFilter) Matches(wallet *IdentityWallet, now time.Time) bool {
	if f.Status != nil && wallet.Status != *f.Status {
		return false
	}

	if f.MinScore != nil && wallet.CurrentScore < *f.MinScore {
		return false
	}

	if f.MaxScore != nil && wallet.CurrentScore > *f.MaxScore {
		return false
	}

	if f.Tier != nil && wallet.Tier != *f.Tier {
		return false
	}

	if f.HasActiveScopes != nil {
		hasActive := len(wallet.GetActiveScopeRefs(now)) > 0
		if hasActive != *f.HasActiveScopes {
			return false
		}
	}

	if f.ScopeType != nil {
		found := false
		for _, ref := range wallet.ScopeRefs {
			if ref.ScopeType == *f.ScopeType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// ============================================================================
// Wallet Binding Utilities
// ============================================================================

// WalletBindingData contains the data used for cryptographic wallet binding.
type WalletBindingData struct {
	// WalletID is the unique wallet identifier
	WalletID string

	// AccountAddress is the bound account address
	AccountAddress string

	// BindingPubKey is the public key used for binding
	BindingPubKey []byte

	// BoundAt is when the binding was created/updated
	BoundAt time.Time
}

// ComputeBindingHash computes a hash of the binding data.
func (b *WalletBindingData) ComputeBindingHash() []byte {
	data := []byte(b.WalletID + ":" + b.AccountAddress)
	data = append(data, b.BindingPubKey...)
	hash := sha256.Sum256(data)
	return hash[:]
}

// ============================================================================
// Wallet Verification State
// ============================================================================

// WalletVerificationState represents the current verification state of a wallet.
type WalletVerificationState struct {
	// Score is the current identity score (0-100)
	Score uint32

	// Status is the current verification status
	Status AccountStatus

	// Tier is the current identity tier
	Tier IdentityTier

	// LastVerifiedAt is when the wallet was last verified
	LastVerifiedAt *time.Time

	// ModelVersion is the ML model version used for scoring
	ModelVersion string

	// ActiveScopeCount is the number of active scopes
	ActiveScopeCount int

	// TotalScopeCount is the total number of scopes
	TotalScopeCount int

	// HasRequiredScopes indicates if minimum required scopes are present
	HasRequiredScopes bool

	// ConsentsGranted is the number of scopes with active consent
	ConsentsGranted int
}

// GetVerificationState returns the current verification state of a wallet.
func (w *IdentityWallet) GetVerificationState(now time.Time) WalletVerificationState {
	var lastVerified *time.Time
	var modelVersion string

	if len(w.VerificationHistory) > 0 {
		lastEntry := w.VerificationHistory[len(w.VerificationHistory)-1]
		lastVerified = &lastEntry.Timestamp
		modelVersion = lastEntry.ModelVersion
	}

	activeScopes := w.GetActiveScopeRefs(now)
	activeConsents := w.ConsentSettings.GetActiveConsentsAt(now)

	// Check for required scopes (ID document and selfie minimum)
	hasIDDoc := false
	hasSelfie := false
	for _, ref := range activeScopes {
		if ref.ScopeType == ScopeTypeIDDocument {
			hasIDDoc = true
		}
		if ref.ScopeType == ScopeTypeSelfie {
			hasSelfie = true
		}
	}

	return WalletVerificationState{
		Score:             w.CurrentScore,
		Status:            w.ScoreStatus,
		Tier:              w.Tier,
		LastVerifiedAt:    lastVerified,
		ModelVersion:      modelVersion,
		ActiveScopeCount:  len(activeScopes),
		TotalScopeCount:   len(w.ScopeRefs),
		HasRequiredScopes: hasIDDoc && hasSelfie,
		ConsentsGranted:   len(activeConsents),
	}
}

// ============================================================================
// Wallet Eligibility Check
// ============================================================================

// WalletEligibility represents what a wallet is eligible for.
type WalletEligibility struct {
	// CanAccessBasic indicates eligibility for basic marketplace
	CanAccessBasic bool

	// CanAccessStandard indicates eligibility for standard features
	CanAccessStandard bool

	// CanAccessPremium indicates eligibility for premium features
	CanAccessPremium bool

	// CanRegisterAsProvider indicates eligibility to register as provider
	CanRegisterAsProvider bool

	// CanRegisterAsValidator indicates eligibility to register as validator
	CanRegisterAsValidator bool

	// MissingRequirements lists what's missing for higher tiers
	MissingRequirements []string
}

// GetEligibility returns what the wallet is eligible for.
func (w *IdentityWallet) GetEligibility(now time.Time) WalletEligibility {
	state := w.GetVerificationState(now)
	eligibility := WalletEligibility{
		MissingRequirements: make([]string, 0),
	}

	// Check basic requirements
	if state.Status == AccountStatusVerified && state.Score >= ThresholdBasic {
		eligibility.CanAccessBasic = true
	} else {
		if state.Status != AccountStatusVerified {
			eligibility.MissingRequirements = append(eligibility.MissingRequirements, "verified status required")
		}
		if state.Score < ThresholdBasic {
			eligibility.MissingRequirements = append(eligibility.MissingRequirements, "score must be at least 50")
		}
	}

	// Check standard requirements
	if state.Status == AccountStatusVerified && state.Score >= ThresholdStandard {
		eligibility.CanAccessStandard = true
	} else if state.Score < ThresholdStandard {
		eligibility.MissingRequirements = append(eligibility.MissingRequirements, "score must be at least 70 for standard access")
	}

	// Check premium requirements
	if state.Status == AccountStatusVerified && state.Score >= ThresholdPremium {
		eligibility.CanAccessPremium = true
	} else if state.Score < ThresholdPremium {
		eligibility.MissingRequirements = append(eligibility.MissingRequirements, "score must be at least 85 for premium access")
	}

	// Check provider eligibility (standard + domain verification)
	hasDomainVerify := false
	for _, ref := range w.GetActiveScopeRefs(now) {
		if ref.ScopeType == ScopeTypeDomainVerify {
			hasDomainVerify = true
			break
		}
	}

	if eligibility.CanAccessStandard && hasDomainVerify {
		eligibility.CanRegisterAsProvider = true
	} else if !hasDomainVerify && eligibility.CanAccessStandard {
		eligibility.MissingRequirements = append(eligibility.MissingRequirements, "domain verification required for provider registration")
	}

	// Check validator eligibility (premium + domain + liveness)
	hasLiveness := false
	for _, ref := range w.GetActiveScopeRefs(now) {
		if ref.ScopeType == ScopeTypeFaceVideo {
			hasLiveness = true
			break
		}
	}

	if eligibility.CanAccessPremium && hasDomainVerify && hasLiveness {
		eligibility.CanRegisterAsValidator = true
	} else if eligibility.CanAccessPremium {
		if !hasDomainVerify {
			eligibility.MissingRequirements = append(eligibility.MissingRequirements, "domain verification required for validator registration")
		}
		if !hasLiveness {
			eligibility.MissingRequirements = append(eligibility.MissingRequirements, "liveness verification required for validator registration")
		}
	}

	return eligibility
}

// ============================================================================
// Wallet Events for Indexing
// ============================================================================

// WalletEventType represents types of wallet events.
type WalletEventType string

const (
	// WalletEventCreated is emitted when a wallet is created
	WalletEventCreated WalletEventType = "wallet_created"

	// WalletEventScopeAdded is emitted when a scope is added
	WalletEventScopeAdded WalletEventType = "scope_added"

	// WalletEventScopeRevoked is emitted when a scope is revoked
	WalletEventScopeRevoked WalletEventType = "scope_revoked"

	// WalletEventConsentUpdated is emitted when consent is updated
	WalletEventConsentUpdated WalletEventType = "consent_updated"

	// WalletEventScoreUpdated is emitted when score is updated
	WalletEventScoreUpdated WalletEventType = "score_updated"

	// WalletEventRebound is emitted when wallet is rebound
	WalletEventRebound WalletEventType = "wallet_rebound"

	// WalletEventStatusChanged is emitted when wallet status changes
	WalletEventStatusChanged WalletEventType = "status_changed"
)

// WalletEvent represents an event that occurred on a wallet.
type WalletEvent struct {
	// EventType is the type of event
	EventType WalletEventType `json:"event_type"`

	// WalletID is the wallet this event occurred on
	WalletID string `json:"wallet_id"`

	// AccountAddress is the wallet's account address
	AccountAddress string `json:"account_address"`

	// BlockHeight is when the event occurred
	BlockHeight int64 `json:"block_height"`

	// Timestamp is when the event occurred
	Timestamp time.Time `json:"timestamp"`

	// Details contains event-specific details
	Details map[string]string `json:"details,omitempty"`
}

// NewWalletEvent creates a new wallet event.
func NewWalletEvent(
	eventType WalletEventType,
	walletID string,
	accountAddress string,
	blockHeight int64,
	timestamp time.Time,
) *WalletEvent {
	return &WalletEvent{
		EventType:      eventType,
		WalletID:       walletID,
		AccountAddress: accountAddress,
		BlockHeight:    blockHeight,
		Timestamp:      timestamp,
		Details:        make(map[string]string),
	}
}

// AddDetail adds a detail to the event.
func (e *WalletEvent) AddDetail(key, value string) {
	e.Details[key] = value
}
