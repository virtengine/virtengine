package types

import (
	"time"
)

// FactorCombination represents a set of factors that must ALL be satisfied (AND logic)
type FactorCombination struct {
	// Factors is the list of factor types that must all be verified
	Factors []FactorType `json:"factors"`

	// MinSecurityLevel is the minimum security level required for the combination
	MinSecurityLevel FactorSecurityLevel `json:"min_security_level,omitempty"`
}

// Validate validates the factor combination
func (fc *FactorCombination) Validate() error {
	if len(fc.Factors) == 0 {
		return ErrInvalidPolicy.Wrap("factor combination must have at least one factor")
	}

	seen := make(map[FactorType]bool)
	for _, ft := range fc.Factors {
		if !ft.IsValid() {
			return ErrInvalidFactorType.Wrapf("invalid factor type in combination: %d", ft)
		}
		if seen[ft] {
			return ErrInvalidPolicy.Wrapf("duplicate factor type in combination: %s", ft.String())
		}
		seen[ft] = true
	}

	return nil
}

// HasFactor returns true if the combination includes the given factor type
func (fc *FactorCombination) HasFactor(ft FactorType) bool {
	for _, f := range fc.Factors {
		if f == ft {
			return true
		}
	}
	return false
}

// GetSecurityLevel returns the combined security level of all factors
func (fc *FactorCombination) GetSecurityLevel() FactorSecurityLevel {
	var maxLevel FactorSecurityLevel
	for _, ft := range fc.Factors {
		level := ft.GetSecurityLevel()
		if level > maxLevel {
			maxLevel = level
		}
	}
	return maxLevel
}

// TrustedDevicePolicy defines how trusted devices can reduce MFA requirements
type TrustedDevicePolicy struct {
	// Enabled indicates if trusted device reduction is enabled
	Enabled bool `json:"enabled"`

	// TrustDuration is how long a device remains trusted (in seconds)
	TrustDuration int64 `json:"trust_duration"`

	// ReducedFactors is the factor combination to use for trusted devices
	// If empty, MFA is skipped entirely for trusted devices
	ReducedFactors *FactorCombination `json:"reduced_factors,omitempty"`

	// MaxTrustedDevices is the maximum number of trusted devices per account
	MaxTrustedDevices uint32 `json:"max_trusted_devices"`

	// RequireReauthForSensitive if true, still requires full MFA for critical actions
	RequireReauthForSensitive bool `json:"require_reauth_for_sensitive"`
}

// Validate validates the trusted device policy
func (p *TrustedDevicePolicy) Validate() error {
	if !p.Enabled {
		return nil
	}

	if p.TrustDuration <= 0 {
		return ErrInvalidPolicy.Wrap("trust_duration must be positive when enabled")
	}

	if p.MaxTrustedDevices == 0 {
		return ErrInvalidPolicy.Wrap("max_trusted_devices must be positive when enabled")
	}

	if p.ReducedFactors != nil {
		if err := p.ReducedFactors.Validate(); err != nil {
			return ErrInvalidPolicy.Wrapf("invalid reduced_factors: %v", err)
		}
	}

	return nil
}

// MFAPolicy defines the MFA requirements for an account
type MFAPolicy struct {
	// AccountAddress is the account this policy applies to
	AccountAddress string `json:"account_address"`

	// RequiredFactors is a list of factor combinations (OR logic between combinations)
	// The user can satisfy ANY ONE of these combinations
	// Each combination uses AND logic internally
	RequiredFactors []FactorCombination `json:"required_factors"`

	// TrustedDeviceRule defines how trusted devices affect MFA requirements
	TrustedDeviceRule *TrustedDevicePolicy `json:"trusted_device_rule,omitempty"`

	// RecoveryFactors defines factor combinations for account recovery
	// Typically more stringent than standard MFA
	RecoveryFactors []FactorCombination `json:"recovery_factors,omitempty"`

	// KeyRotationFactors defines factor combinations for key rotation
	// Should include current key proof + additional factors
	KeyRotationFactors []FactorCombination `json:"key_rotation_factors,omitempty"`

	// SessionDuration is how long an MFA session remains valid (in seconds)
	SessionDuration int64 `json:"session_duration"`

	// VEIDThreshold is the minimum VEID score required for VEID-based factors
	VEIDThreshold uint32 `json:"veid_threshold,omitempty"`

	// Enabled indicates if MFA is enabled for this account
	Enabled bool `json:"enabled"`

	// CreatedAt is when the policy was created
	CreatedAt int64 `json:"created_at"`

	// UpdatedAt is when the policy was last updated
	UpdatedAt int64 `json:"updated_at"`
}

// Validate validates the MFA policy
func (p *MFAPolicy) Validate() error {
	if p.AccountAddress == "" {
		return ErrInvalidAddress.Wrap("account address cannot be empty")
	}

	if p.Enabled && len(p.RequiredFactors) == 0 {
		return ErrInvalidPolicy.Wrap("enabled policy must have at least one required factor combination")
	}

	for i, fc := range p.RequiredFactors {
		if err := fc.Validate(); err != nil {
			return ErrInvalidPolicy.Wrapf("invalid required_factors[%d]: %v", i, err)
		}
	}

	for i, fc := range p.RecoveryFactors {
		if err := fc.Validate(); err != nil {
			return ErrInvalidPolicy.Wrapf("invalid recovery_factors[%d]: %v", i, err)
		}
	}

	for i, fc := range p.KeyRotationFactors {
		if err := fc.Validate(); err != nil {
			return ErrInvalidPolicy.Wrapf("invalid key_rotation_factors[%d]: %v", i, err)
		}
	}

	if p.TrustedDeviceRule != nil {
		if err := p.TrustedDeviceRule.Validate(); err != nil {
			return err
		}
	}

	if p.SessionDuration < 0 {
		return ErrInvalidPolicy.Wrap("session_duration cannot be negative")
	}

	return nil
}

// GetRequiredFactorsForAction returns the factor combinations required for a given action
func (p *MFAPolicy) GetRequiredFactorsForAction(action SensitiveTransactionType) []FactorCombination {
	switch action {
	case SensitiveTxAccountRecovery:
		if len(p.RecoveryFactors) > 0 {
			return p.RecoveryFactors
		}
	case SensitiveTxKeyRotation:
		if len(p.KeyRotationFactors) > 0 {
			return p.KeyRotationFactors
		}
	}

	return p.RequiredFactors
}

// CanUseTrustedDevice returns true if a trusted device can reduce MFA for this action
func (p *MFAPolicy) CanUseTrustedDevice(action SensitiveTransactionType, isTrustedDevice bool) bool {
	if p.TrustedDeviceRule == nil || !p.TrustedDeviceRule.Enabled {
		return false
	}

	if !isTrustedDevice {
		return false
	}

	// Critical actions always require full MFA
	if p.TrustedDeviceRule.RequireReauthForSensitive {
		switch action {
		case SensitiveTxAccountRecovery, SensitiveTxKeyRotation,
			SensitiveTxValidatorRegistration, SensitiveTxRoleAssignment:
			return false
		}
	}

	return true
}

// GetReducedFactors returns the reduced factor requirements for trusted devices
func (p *MFAPolicy) GetReducedFactors() *FactorCombination {
	if p.TrustedDeviceRule == nil || !p.TrustedDeviceRule.Enabled {
		return nil
	}
	return p.TrustedDeviceRule.ReducedFactors
}

// DefaultMFAPolicy returns a default MFA policy for a new account
func DefaultMFAPolicy(accountAddress string) *MFAPolicy {
	now := time.Now().Unix()
	return &MFAPolicy{
		AccountAddress: accountAddress,
		RequiredFactors: []FactorCombination{
			{Factors: []FactorType{FactorTypeFIDO2}},                // FIDO2 alone is sufficient
			{Factors: []FactorType{FactorTypeTOTP, FactorTypeVEID}}, // Or TOTP + VEID
		},
		RecoveryFactors: []FactorCombination{
			{Factors: []FactorType{FactorTypeVEID, FactorTypeFIDO2, FactorTypeSMS}},
			{Factors: []FactorType{FactorTypeVEID, FactorTypeFIDO2, FactorTypeEmail}},
		},
		KeyRotationFactors: []FactorCombination{
			{Factors: []FactorType{FactorTypeFIDO2, FactorTypeTOTP}},
		},
		TrustedDeviceRule: &TrustedDevicePolicy{
			Enabled:                   true,
			TrustDuration:             30 * 24 * 60 * 60, // 30 days
			MaxTrustedDevices:         5,
			RequireReauthForSensitive: true,
			ReducedFactors:            nil, // Skip MFA for low-risk actions on trusted devices
		},
		SessionDuration: 15 * 60, // 15 minutes
		VEIDThreshold:   50,      // Default minimum VEID score
		Enabled:         false,   // Disabled by default, user must opt-in
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// PolicyMatch represents the result of checking if available factors match a policy
type PolicyMatch struct {
	// Matched indicates if the available factors satisfy the policy
	Matched bool

	// MatchedCombination is the factor combination that was satisfied (if matched)
	MatchedCombination *FactorCombination

	// MissingFactors is the list of factors still needed (if not matched)
	MissingFactors []FactorType

	// AvailableOptions are other combinations that could satisfy the policy
	AvailableOptions []FactorCombination
}

// CheckFactorsMatch checks if the provided factors satisfy any combination in the policy
func (p *MFAPolicy) CheckFactorsMatch(availableFactors []FactorType) PolicyMatch {
	result := PolicyMatch{
		Matched:          false,
		AvailableOptions: p.RequiredFactors,
	}

	// Create a set of available factors for quick lookup
	available := make(map[FactorType]bool)
	for _, ft := range availableFactors {
		available[ft] = true
	}

	// Check each combination (OR logic)
	for i, combo := range p.RequiredFactors {
		allPresent := true
		var missing []FactorType

		// Check all factors in this combination (AND logic)
		for _, requiredFactor := range combo.Factors {
			if !available[requiredFactor] {
				allPresent = false
				missing = append(missing, requiredFactor)
			}
		}

		if allPresent {
			result.Matched = true
			result.MatchedCombination = &p.RequiredFactors[i]
			result.MissingFactors = nil
			return result
		}

		// Track the combination with fewest missing factors
		if result.MissingFactors == nil || len(missing) < len(result.MissingFactors) {
			result.MissingFactors = missing
		}
	}

	return result
}
