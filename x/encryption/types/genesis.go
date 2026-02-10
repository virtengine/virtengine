package types

import (
	encryptionv1 "github.com/virtengine/virtengine/sdk/go/node/encryption/v1"
)

// GenesisState is an alias to the generated proto type for the encryption module's genesis state.
type GenesisState = encryptionv1.GenesisState

// Params is an alias to the generated proto type for encryption module parameters.
type Params = encryptionv1.Params

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		RecipientKeys: []encryptionv1.RecipientKeyRecord{},
		Params:        DefaultParams(),
	}
}

// DefaultParams returns the default parameters
func DefaultParams() Params {
	return Params{
		MaxRecipientsPerEnvelope:     10,
		MaxKeysPerAccount:            5,
		AllowedAlgorithms:            SupportedAlgorithms(),
		RequireSignature:             true,
		RevocationGracePeriodSeconds: 604800, // 7 days
		KeyExpiryWarningSeconds:      []uint64{604800, 86400},
		RotationBatchSize:            100,
		DefaultKeyTtlSeconds:         0,
	}
}

// ValidateGenesis validates the genesis state
func ValidateGenesis(gs *GenesisState) error {
	// Validate recipient keys
	seen := make(map[string]bool)
	for _, key := range gs.RecipientKeys {
		if err := ValidateRecipientKeyRecord(&key); err != nil {
			return err
		}
		if seen[key.KeyFingerprint] {
			return ErrKeyAlreadyExists.Wrapf("duplicate key fingerprint: %s", key.KeyFingerprint)
		}
		seen[key.KeyFingerprint] = true
	}

	// Validate params
	return ValidateParams(&gs.Params)
}

// ValidateParams validates the params
func ValidateParams(p *Params) error {
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

	if p.RotationBatchSize == 0 {
		return ErrInvalidEnvelope.Wrap("rotation_batch_size must be greater than 0")
	}

	return nil
}

// IsAlgorithmAllowed checks if an algorithm is allowed by the params
func IsAlgorithmAllowed(p *Params, algorithmID string) bool {
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

// ValidateRecipientKeyRecord validates the proto RecipientKeyRecord type
func ValidateRecipientKeyRecord(r *encryptionv1.RecipientKeyRecord) error {
	if len(r.Address) == 0 {
		return ErrInvalidAddress.Wrap("address cannot be empty")
	}

	if len(r.PublicKey) == 0 {
		return ErrInvalidPublicKey.Wrap("public key cannot be empty")
	}

	if len(r.KeyFingerprint) == 0 {
		return ErrInvalidPublicKey.Wrap("key fingerprint cannot be empty")
	}

	if !IsAlgorithmSupported(r.AlgorithmId) {
		return ErrUnsupportedAlgorithm.Wrapf("algorithm %s is not supported", r.AlgorithmId)
	}

	algInfo, err := GetAlgorithmInfo(r.AlgorithmId)
	if err != nil {
		return err
	}

	if len(r.PublicKey) != algInfo.KeySize {
		return ErrInvalidPublicKey.Wrapf("public key size mismatch: expected %d, got %d",
			algInfo.KeySize, len(r.PublicKey))
	}

	return nil
}
