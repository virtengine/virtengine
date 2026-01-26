package types

// GenesisState is the genesis state for the mfa module
type GenesisState struct {
	// Params are the module parameters
	Params Params `json:"params"`

	// MFAPolicies are the initial MFA policies
	MFAPolicies []MFAPolicy `json:"mfa_policies"`

	// FactorEnrollments are the initial factor enrollments
	FactorEnrollments []FactorEnrollment `json:"factor_enrollments"`

	// SensitiveTxConfigs are the sensitive transaction configurations
	SensitiveTxConfigs []SensitiveTxConfig `json:"sensitiVIRTENGINE_tx_configs"`

	// TrustedDevices are the initial trusted devices
	TrustedDevices []TrustedDevice `json:"trusted_devices"`
}

// TrustedDevice represents a stored trusted device record
type TrustedDevice struct {
	// AccountAddress is the account that trusts this device
	AccountAddress string `json:"account_address"`

	// DeviceInfo contains the device information
	DeviceInfo DeviceInfo `json:"device_info"`

	// AddedAt is when the device was added
	AddedAt int64 `json:"added_at"`

	// LastUsedAt is when the device was last used
	LastUsedAt int64 `json:"last_used_at"`
}

// Params defines the parameters for the mfa module
type Params struct {
	// DefaultSessionDuration is the default MFA session duration in seconds
	DefaultSessionDuration int64 `json:"default_session_duration"`

	// MaxFactorsPerAccount is the maximum number of factors per account
	MaxFactorsPerAccount uint32 `json:"max_factors_per_account"`

	// MaxChallengeAttempts is the maximum verification attempts per challenge
	MaxChallengeAttempts uint32 `json:"max_challenge_attempts"`

	// ChallengeTTL is the challenge time-to-live in seconds
	ChallengeTTL int64 `json:"challenge_ttl"`

	// MaxTrustedDevices is the maximum trusted devices per account
	MaxTrustedDevices uint32 `json:"max_trusted_devices"`

	// TrustedDeviceTTL is the trusted device time-to-live in seconds
	TrustedDeviceTTL int64 `json:"trusted_device_ttl"`

	// MinVEIDScoreForMFA is the minimum VEID score to enable MFA
	MinVEIDScoreForMFA uint32 `json:"min_veid_score_for_mfa"`

	// RequireAtLeastOneFactor requires at least one factor when MFA is enabled
	RequireAtLeastOneFactor bool `json:"require_at_least_one_factor"`

	// AllowedFactorTypes lists the factor types allowed on this chain
	AllowedFactorTypes []FactorType `json:"allowed_factor_types"`
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:             DefaultParams(),
		MFAPolicies:        []MFAPolicy{},
		FactorEnrollments:  []FactorEnrollment{},
		SensitiveTxConfigs: GetDefaultSensitiveTxConfigs(),
		TrustedDevices:     []TrustedDevice{},
	}
}

// DefaultParams returns the default parameters
func DefaultParams() Params {
	return Params{
		DefaultSessionDuration:  15 * 60, // 15 minutes
		MaxFactorsPerAccount:    10,
		MaxChallengeAttempts:    3,
		ChallengeTTL:            5 * 60, // 5 minutes
		MaxTrustedDevices:       5,
		TrustedDeviceTTL:        30 * 24 * 60 * 60, // 30 days
		MinVEIDScoreForMFA:      50,
		RequireAtLeastOneFactor: true,
		AllowedFactorTypes: []FactorType{
			FactorTypeTOTP,
			FactorTypeFIDO2,
			FactorTypeSMS,
			FactorTypeEmail,
			FactorTypeVEID,
			FactorTypeTrustedDevice,
		},
	}
}

// Validate validates the genesis state
func (gs GenesisState) Validate() error {
	// Validate params
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// Validate MFA policies
	for i, policy := range gs.MFAPolicies {
		if err := policy.Validate(); err != nil {
			return ErrInvalidPolicy.Wrapf("invalid policy[%d]: %v", i, err)
		}
	}

	// Validate factor enrollments
	for i, enrollment := range gs.FactorEnrollments {
		if err := enrollment.Validate(); err != nil {
			return ErrInvalidEnrollment.Wrapf("invalid enrollment[%d]: %v", i, err)
		}
	}

	// Validate sensitive tx configs
	for i, config := range gs.SensitiveTxConfigs {
		if err := config.Validate(); err != nil {
			return ErrInvalidSensitiveTxConfig.Wrapf("invalid config[%d]: %v", i, err)
		}
	}

	return nil
}

// Validate validates the params
func (p Params) Validate() error {
	if p.DefaultSessionDuration <= 0 {
		return ErrInvalidPolicy.Wrap("default_session_duration must be positive")
	}

	if p.MaxFactorsPerAccount == 0 {
		return ErrInvalidPolicy.Wrap("max_factors_per_account must be positive")
	}

	if p.MaxChallengeAttempts == 0 {
		return ErrInvalidPolicy.Wrap("max_challenge_attempts must be positive")
	}

	if p.ChallengeTTL <= 0 {
		return ErrInvalidPolicy.Wrap("challenge_ttl must be positive")
	}

	if p.MaxTrustedDevices == 0 {
		return ErrInvalidPolicy.Wrap("max_trusted_devices must be positive")
	}

	if p.TrustedDeviceTTL <= 0 {
		return ErrInvalidPolicy.Wrap("trusted_device_ttl must be positive")
	}

	// Validate allowed factor types
	for _, ft := range p.AllowedFactorTypes {
		if !ft.IsValid() {
			return ErrInvalidFactorType.Wrapf("invalid allowed factor type: %d", ft)
		}
	}

	return nil
}

// IsFactorTypeAllowed returns true if the factor type is allowed
func (p Params) IsFactorTypeAllowed(ft FactorType) bool {
	for _, allowed := range p.AllowedFactorTypes {
		if allowed == ft {
			return true
		}
	}
	return false
}
