package types

import (
	"time"
)

// ============================================================================
// Consent Settings Types
// ============================================================================

// ScopeConsent represents consent configuration for a specific scope
type ScopeConsent struct {
	// ScopeID is the identifier of the scope this consent applies to
	ScopeID string `json:"scope_id"`

	// Granted indicates if consent is currently granted
	Granted bool `json:"granted"`

	// GrantedAt is when consent was granted (nil if never granted)
	GrantedAt *time.Time `json:"granted_at,omitempty"`

	// RevokedAt is when consent was revoked (nil if currently granted or never revoked)
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	// ExpiresAt is when this consent expires (nil for no expiration)
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// Purpose describes the purpose for which consent was granted
	Purpose string `json:"purpose,omitempty"`

	// GrantedToProviders lists specific providers consent was granted to
	// Empty means consent is general (not provider-specific)
	GrantedToProviders []string `json:"granted_to_providers,omitempty"`

	// Restrictions contains any restrictions on this consent
	Restrictions []string `json:"restrictions,omitempty"`
}

// NewScopeConsent creates a new scope consent
func NewScopeConsent(scopeID string, granted bool, purpose string) *ScopeConsent {
	consent := &ScopeConsent{
		ScopeID:            scopeID,
		Granted:            granted,
		Purpose:            purpose,
		GrantedToProviders: make([]string, 0),
		Restrictions:       make([]string, 0),
	}
	if granted {
		now := time.Now()
		consent.GrantedAt = &now
	}
	return consent
}

// Validate validates the scope consent
func (sc *ScopeConsent) Validate() error {
	if sc.ScopeID == "" {
		return ErrInvalidScope.Wrap("scope_id cannot be empty in consent")
	}

	// If consent is granted, GrantedAt should be set
	if sc.Granted && sc.GrantedAt == nil {
		return ErrInvalidWallet.Wrap("granted_at must be set when consent is granted")
	}

	return nil
}

// IsActive checks if the consent is currently active (granted and not expired)
func (sc *ScopeConsent) IsActive() bool {
	if !sc.Granted {
		return false
	}

	if sc.ExpiresAt != nil && time.Now().After(*sc.ExpiresAt) {
		return false
	}

	return true
}

// Grant grants consent with the given purpose and optional expiration
func (sc *ScopeConsent) Grant(purpose string, expiresAt *time.Time) {
	now := time.Now()
	sc.Granted = true
	sc.GrantedAt = &now
	sc.RevokedAt = nil
	sc.Purpose = purpose
	sc.ExpiresAt = expiresAt
}

// Revoke revokes the consent
func (sc *ScopeConsent) Revoke() {
	now := time.Now()
	sc.Granted = false
	sc.RevokedAt = &now
}

// AddProviderGrant adds a provider-specific consent grant
func (sc *ScopeConsent) AddProviderGrant(providerAddress string) {
	for _, p := range sc.GrantedToProviders {
		if p == providerAddress {
			return // Already granted
		}
	}
	sc.GrantedToProviders = append(sc.GrantedToProviders, providerAddress)
}

// RemoveProviderGrant removes a provider-specific consent grant
func (sc *ScopeConsent) RemoveProviderGrant(providerAddress string) bool {
	for i, p := range sc.GrantedToProviders {
		if p == providerAddress {
			sc.GrantedToProviders = append(sc.GrantedToProviders[:i], sc.GrantedToProviders[i+1:]...)
			return true
		}
	}
	return false
}

// IsGrantedToProvider checks if consent is granted to a specific provider
func (sc *ScopeConsent) IsGrantedToProvider(providerAddress string) bool {
	if !sc.IsActive() {
		return false
	}

	// If no specific providers listed, consent is general
	if len(sc.GrantedToProviders) == 0 {
		return true
	}

	for _, p := range sc.GrantedToProviders {
		if p == providerAddress {
			return true
		}
	}
	return false
}

// ConsentSettings represents the global consent configuration for an identity wallet
type ConsentSettings struct {
	// ScopeConsents contains per-scope consent settings
	// Key is scopeID
	ScopeConsents map[string]ScopeConsent `json:"scope_consents"`

	// ShareWithProviders allows providers to access non-sensitive identity metadata
	ShareWithProviders bool `json:"share_with_providers"`

	// ShareForVerification allows the identity to be used for verification requests
	ShareForVerification bool `json:"share_for_verification"`

	// AllowReVerification allows the identity to be re-verified without explicit request
	AllowReVerification bool `json:"allow_re_verification"`

	// AllowDerivedFeatureSharing allows sharing of derived feature hashes
	AllowDerivedFeatureSharing bool `json:"allow_derived_feature_sharing"`

	// GlobalExpiresAt sets a global expiration for all consents (optional)
	GlobalExpiresAt *time.Time `json:"global_expires_at,omitempty"`

	// LastUpdatedAt is when consent settings were last modified
	LastUpdatedAt time.Time `json:"last_updated_at"`

	// ConsentVersion tracks consent settings version for audit
	ConsentVersion uint32 `json:"consent_version"`
}

// NewConsentSettings creates new consent settings with secure defaults
func NewConsentSettings() ConsentSettings {
	return ConsentSettings{
		ScopeConsents:              make(map[string]ScopeConsent),
		ShareWithProviders:         false, // Secure default: no sharing
		ShareForVerification:       false, // Secure default: explicit opt-in required
		AllowReVerification:        false, // Secure default: explicit opt-in required
		AllowDerivedFeatureSharing: false, // Secure default: no feature sharing
		LastUpdatedAt:              time.Now(),
		ConsentVersion:             1,
	}
}

// Validate validates the consent settings
func (cs *ConsentSettings) Validate() error {
	for scopeID, consent := range cs.ScopeConsents {
		if consent.ScopeID != scopeID {
			return ErrInvalidWallet.Wrapf("scope consent key %s does not match ScopeID %s", scopeID, consent.ScopeID)
		}
		if err := consent.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// SetScopeConsent sets consent for a specific scope
func (cs *ConsentSettings) SetScopeConsent(consent ScopeConsent) {
	cs.ScopeConsents[consent.ScopeID] = consent
	cs.LastUpdatedAt = time.Now()
	cs.ConsentVersion++
}

// GetScopeConsent returns consent settings for a specific scope
func (cs *ConsentSettings) GetScopeConsent(scopeID string) (ScopeConsent, bool) {
	consent, found := cs.ScopeConsents[scopeID]
	return consent, found
}

// GrantScopeConsent grants consent for a scope
func (cs *ConsentSettings) GrantScopeConsent(scopeID string, purpose string, expiresAt *time.Time) {
	consent, found := cs.ScopeConsents[scopeID]
	if !found {
		consent = ScopeConsent{
			ScopeID:            scopeID,
			GrantedToProviders: make([]string, 0),
			Restrictions:       make([]string, 0),
		}
	}
	consent.Grant(purpose, expiresAt)
	cs.ScopeConsents[scopeID] = consent
	cs.LastUpdatedAt = time.Now()
	cs.ConsentVersion++
}

// RevokeScopeConsent revokes consent for a scope
func (cs *ConsentSettings) RevokeScopeConsent(scopeID string) bool {
	consent, found := cs.ScopeConsents[scopeID]
	if !found {
		return false
	}
	consent.Revoke()
	cs.ScopeConsents[scopeID] = consent
	cs.LastUpdatedAt = time.Now()
	cs.ConsentVersion++
	return true
}

// IsScopeConsentActive checks if consent is active for a scope
func (cs *ConsentSettings) IsScopeConsentActive(scopeID string) bool {
	consent, found := cs.ScopeConsents[scopeID]
	if !found {
		return false
	}
	return consent.IsActive()
}

// RevokeAll revokes all consents
func (cs *ConsentSettings) RevokeAll() {
	for scopeID, consent := range cs.ScopeConsents {
		consent.Revoke()
		cs.ScopeConsents[scopeID] = consent
	}
	cs.ShareWithProviders = false
	cs.ShareForVerification = false
	cs.AllowReVerification = false
	cs.AllowDerivedFeatureSharing = false
	cs.LastUpdatedAt = time.Now()
	cs.ConsentVersion++
}

// SetGlobalSettings updates global consent settings
func (cs *ConsentSettings) SetGlobalSettings(
	shareWithProviders bool,
	shareForVerification bool,
	allowReVerification bool,
	allowDerivedFeatureSharing bool,
) {
	cs.ShareWithProviders = shareWithProviders
	cs.ShareForVerification = shareForVerification
	cs.AllowReVerification = allowReVerification
	cs.AllowDerivedFeatureSharing = allowDerivedFeatureSharing
	cs.LastUpdatedAt = time.Now()
	cs.ConsentVersion++
}

// GetActiveConsents returns all active scope consents
func (cs *ConsentSettings) GetActiveConsents() []ScopeConsent {
	var active []ScopeConsent
	for _, consent := range cs.ScopeConsents {
		if consent.IsActive() {
			active = append(active, consent)
		}
	}
	return active
}

// ConsentUpdateRequest represents a request to update consent settings
type ConsentUpdateRequest struct {
	// ScopeID is the scope to update consent for (empty for global settings)
	ScopeID string `json:"scope_id,omitempty"`

	// GrantConsent indicates whether to grant or revoke consent
	GrantConsent bool `json:"grant_consent"`

	// Purpose is the purpose for granting consent
	Purpose string `json:"purpose,omitempty"`

	// ExpiresAt is when the consent should expire
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// GlobalSettings contains global settings updates
	GlobalSettings *GlobalConsentUpdate `json:"global_settings,omitempty"`
}

// GlobalConsentUpdate represents updates to global consent settings
type GlobalConsentUpdate struct {
	ShareWithProviders         *bool `json:"share_with_providers,omitempty"`
	ShareForVerification       *bool `json:"share_for_verification,omitempty"`
	AllowReVerification        *bool `json:"allow_re_verification,omitempty"`
	AllowDerivedFeatureSharing *bool `json:"allow_derived_feature_sharing,omitempty"`
}

// ApplyConsentUpdate applies a consent update request to consent settings
func (cs *ConsentSettings) ApplyConsentUpdate(update ConsentUpdateRequest) {
	// Apply scope-specific consent
	if update.ScopeID != "" {
		if update.GrantConsent {
			cs.GrantScopeConsent(update.ScopeID, update.Purpose, update.ExpiresAt)
		} else {
			cs.RevokeScopeConsent(update.ScopeID)
		}
	}

	// Apply global settings if provided
	if update.GlobalSettings != nil {
		gs := update.GlobalSettings
		if gs.ShareWithProviders != nil {
			cs.ShareWithProviders = *gs.ShareWithProviders
		}
		if gs.ShareForVerification != nil {
			cs.ShareForVerification = *gs.ShareForVerification
		}
		if gs.AllowReVerification != nil {
			cs.AllowReVerification = *gs.AllowReVerification
		}
		if gs.AllowDerivedFeatureSharing != nil {
			cs.AllowDerivedFeatureSharing = *gs.AllowDerivedFeatureSharing
		}
		cs.LastUpdatedAt = time.Now()
		cs.ConsentVersion++
	}
}
