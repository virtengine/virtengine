package types

import (
	"encoding/hex"
	"fmt"

	v1 "github.com/virtengine/virtengine/sdk/go/node/enclave/v1"
)

// DefaultParams returns the default enclave parameters
func DefaultParams() v1.Params {
	return v1.Params{
		MaxEnclaveKeysPerValidator: 2,                                                                  // Allow current + rotating key
		DefaultExpiryBlocks:        1000,                                                               // ~1.5 hours at 5s blocks
		KeyRotationOverlapBlocks:   100,                                                                // ~8 minutes overlap
		MinQuoteVersion:            3,                                                                  // DCAP v3 minimum
		AllowedTeeTypes:            []v1.TEEType{v1.TEETypeSGX, v1.TEETypeSEVSNP, v1.TEETypeNitro},
		ScoreTolerance:             0,                                                                  // Exact match by default
		RequireAttestationChain:    true,
		MaxAttestationAge:          10000,                                                              // ~14 hours
		EnableCommitteeMode:        false,
		CommitteeSize:              0,
		CommitteeEpochBlocks:       10000,                                                              // ~14 hours per epoch
		EnableMeasurementCleanup:   false,                                                              // Disabled by default
		MaxRegistrationsPerBlock:   0,                                                                  // Unlimited by default
		RegistrationCooldownBlocks: 0,                                                                  // No cooldown by default
		HealthCheckParams:          DefaultHealthCheckParams(),
	}
}

// ValidateParams validates the parameters
func ValidateParams(p *v1.Params) error {
	if p.MaxEnclaveKeysPerValidator == 0 {
		return fmt.Errorf("max_enclave_keys_per_validator must be positive")
	}

	if p.DefaultExpiryBlocks <= 0 {
		return fmt.Errorf("default_expiry_blocks must be positive")
	}

	if p.KeyRotationOverlapBlocks <= 0 {
		return fmt.Errorf("key_rotation_overlap_blocks must be positive")
	}

	if p.MinQuoteVersion == 0 {
		return fmt.Errorf("min_quote_version must be positive")
	}

	if len(p.AllowedTeeTypes) == 0 {
		return fmt.Errorf("allowed_tee_types cannot be empty")
	}

	for _, teeType := range p.AllowedTeeTypes {
		if !IsValidTEEType(teeType) {
			return fmt.Errorf("invalid TEE type in allowed list: %s", teeType)
		}
	}

	if p.ScoreTolerance > 100 {
		return fmt.Errorf("score_tolerance cannot exceed 100")
	}

	if p.MaxAttestationAge <= 0 {
		return fmt.Errorf("max_attestation_age must be positive")
	}

	if p.EnableCommitteeMode && p.CommitteeSize == 0 {
		return fmt.Errorf("committee_size must be positive when committee mode is enabled")
	}

	if err := p.HealthCheckParams.Validate(); err != nil {
		return fmt.Errorf("invalid health check params: %w", err)
	}

	return nil
}

// IsTEETypeAllowed checks if a TEE type is in the allowed list
func IsTEETypeAllowed(p *v1.Params, teeType v1.TEEType) bool {
	for _, allowed := range p.AllowedTeeTypes {
		if allowed == teeType {
			return true
		}
	}
	return false
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *v1.GenesisState {
	return &v1.GenesisState{
		EnclaveIdentities:     []v1.EnclaveIdentity{},
		MeasurementAllowlist:  []v1.MeasurementRecord{},
		KeyRotations:          []v1.KeyRotationRecord{},
		EnclaveHealthStatuses: []EnclaveHealthStatus{},
		Params:                DefaultParams(),
	}
}

// ValidateGenesis validates the genesis state
func ValidateGenesis(g *v1.GenesisState) error {
	if err := ValidateParams(&g.Params); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	seenValidators := make(map[string]bool)
	for i, identity := range g.EnclaveIdentities {
		if err := ValidateEnclaveIdentity(&identity); err != nil {
			return fmt.Errorf("invalid enclave identity at index %d: %w", i, err)
		}

		if seenValidators[identity.ValidatorAddress] {
			return fmt.Errorf("duplicate enclave identity for validator %s", identity.ValidatorAddress)
		}
		seenValidators[identity.ValidatorAddress] = true
	}

	seenMeasurements := make(map[string]bool)
	for i, measurement := range g.MeasurementAllowlist {
		if err := ValidateMeasurementRecord(&measurement); err != nil {
			return fmt.Errorf("invalid measurement at index %d: %w", i, err)
		}

		hashHex := hex.EncodeToString(measurement.MeasurementHash)
		if seenMeasurements[hashHex] {
			return fmt.Errorf("duplicate measurement in allowlist: %s", hashHex)
		}
		seenMeasurements[hashHex] = true
	}

	for i, rotation := range g.KeyRotations {
		if err := ValidateKeyRotationRecord(&rotation); err != nil {
			return fmt.Errorf("invalid key rotation at index %d: %w", i, err)
		}
	}

	seenHealthValidators := make(map[string]bool)
	for i, health := range g.EnclaveHealthStatuses {
		if err := health.Validate(); err != nil {
			return fmt.Errorf("invalid health status at index %d: %w", i, err)
		}

		if seenHealthValidators[health.ValidatorAddress] {
			return fmt.Errorf("duplicate health status for validator %s", health.ValidatorAddress)
		}
		seenHealthValidators[health.ValidatorAddress] = true
	}

	return nil
}
