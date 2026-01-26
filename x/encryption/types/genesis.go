package types

// GenesisState is the genesis state for the encryption module
type GenesisState struct {
	// RecipientKeys are the initial registered recipient keys
	RecipientKeys []RecipientKeyRecord `json:"recipient_keys"`

	// Params are the module parameters
	Params Params `json:"params"`
}

// Params defines the parameters for the encryption module
type Params struct {
	// MaxRecipientsPerEnvelope is the maximum number of recipients per envelope
	MaxRecipientsPerEnvelope uint32 `json:"max_recipients_per_envelope"`

	// MaxKeysPerAccount is the maximum number of keys an account can register
	MaxKeysPerAccount uint32 `json:"max_keys_per_account"`

	// AllowedAlgorithms is the list of allowed encryption algorithms
	// Empty means all supported algorithms are allowed
	AllowedAlgorithms []string `json:"allowed_algorithms"`

	// RequireSignature determines if envelope signatures are mandatory
	RequireSignature bool `json:"require_signature"`
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		RecipientKeys: []RecipientKeyRecord{},
		Params:        DefaultParams(),
	}
}

// DefaultParams returns the default parameters
func DefaultParams() Params {
	return Params{
		MaxRecipientsPerEnvelope: 10,
		MaxKeysPerAccount:        5,
		AllowedAlgorithms:        SupportedAlgorithms(),
		RequireSignature:         true,
	}
}

// Validate validates the genesis state
func (gs GenesisState) Validate() error {
	// Validate recipient keys
	seen := make(map[string]bool)
	for _, key := range gs.RecipientKeys {
		if err := key.Validate(); err != nil {
			return err
		}
		if seen[key.KeyFingerprint] {
			return ErrKeyAlreadyExists.Wrapf("duplicate key fingerprint: %s", key.KeyFingerprint)
		}
		seen[key.KeyFingerprint] = true
	}

	// Validate params
	return gs.Params.Validate()
}

// Validate validates the params
func (p Params) Validate() error {
	if p.MaxRecipientsPerEnvelope == 0 {
		return ErrInvalidEnvelope.Wrap("max_recipients_per_envelope must be greater than 0")
	}

	if p.MaxKeysPerAccount == 0 {
		return ErrInvalidPublicKey.Wrap("max_keys_per_account must be greater than 0")
	}

	// Validate allowed algorithms
	for _, alg := range p.AllowedAlgorithms {
		if !IsAlgorithmSupported(alg) {
			return ErrUnsupportedAlgorithm.Wrapf("algorithm %s is not supported", alg)
		}
	}

	return nil
}

// IsAlgorithmAllowed checks if an algorithm is allowed by the params
func (p Params) IsAlgorithmAllowed(algorithmID string) bool {
	// If no specific algorithms are configured, allow all supported
	if len(p.AllowedAlgorithms) == 0 {
		return IsAlgorithmSupported(algorithmID)
	}

	for _, alg := range p.AllowedAlgorithms {
		if alg == algorithmID {
			return true
		}
	}
	return false
}
