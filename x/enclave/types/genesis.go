package types

import "fmt"

// Params defines the parameters for the enclave module
type Params struct {
	// MaxEnclaveKeysPerValidator is the maximum number of enclave keys a validator can have
	MaxEnclaveKeysPerValidator uint32 `json:"max_enclave_keys_per_validator"`

	// DefaultExpiryBlocks is the default number of blocks until enclave identity expires
	DefaultExpiryBlocks int64 `json:"default_expiry_blocks"`

	// KeyRotationOverlapBlocks is the default overlap period for key rotations
	KeyRotationOverlapBlocks int64 `json:"key_rotation_overlap_blocks"`

	// MinQuoteVersion is the minimum attestation quote version required
	MinQuoteVersion uint32 `json:"min_quote_version"`

	// AllowedTEETypes is the list of allowed TEE types
	AllowedTEETypes []TEEType `json:"allowed_tee_types"`

	// ScoreTolerance is the maximum allowed score difference for consensus
	ScoreTolerance uint32 `json:"score_tolerance"`

	// RequireAttestationChain indicates if attestation chain verification is required
	RequireAttestationChain bool `json:"require_attestation_chain"`

	// MaxAttestationAge is the maximum age of attestation in blocks
	MaxAttestationAge int64 `json:"max_attestation_age"`

	// EnableCommitteeMode enables committee-based identity processing
	EnableCommitteeMode bool `json:"enable_committee_mode"`

	// CommitteeSize is the size of the identity committee (if committee mode enabled)
	CommitteeSize uint32 `json:"committee_size,omitempty"`

	// CommitteeEpochBlocks is the number of blocks per committee epoch
	CommitteeEpochBlocks int64 `json:"committee_epoch_blocks,omitempty"`

	// EnableMeasurementCleanup enables automatic cleanup of expired measurements
	EnableMeasurementCleanup bool `json:"enable_measurement_cleanup"`

	// MaxRegistrationsPerBlock limits registrations per block (0 = unlimited)
	MaxRegistrationsPerBlock uint32 `json:"max_registrations_per_block"`

	// RegistrationCooldownBlocks enforces cooldown between re-registrations
	RegistrationCooldownBlocks int64 `json:"registration_cooldown_blocks"`
}

// DefaultParams returns the default enclave parameters
func DefaultParams() Params {
	return Params{
		MaxEnclaveKeysPerValidator: 2,                                        // Allow current + rotating key
		DefaultExpiryBlocks:        1000,                                     // ~1.5 hours at 5s blocks
		KeyRotationOverlapBlocks:   100,                                      // ~8 minutes overlap
		MinQuoteVersion:            3,                                        // DCAP v3 minimum
		AllowedTEETypes:            []TEEType{TEETypeSGX, TEETypeSEVSNP, TEETypeNitro},
		ScoreTolerance:             0,                                        // Exact match by default
		RequireAttestationChain:    true,
		MaxAttestationAge:          10000,                                    // ~14 hours
		EnableCommitteeMode:        false,
		CommitteeSize:              0,
		CommitteeEpochBlocks:       10000,                                    // ~14 hours per epoch
		EnableMeasurementCleanup:   false,                                    // Disabled by default
		MaxRegistrationsPerBlock:   0,                                        // Unlimited by default
		RegistrationCooldownBlocks: 0,                                        // No cooldown by default
	}
}

// Validate validates the parameters
func (p Params) Validate() error {
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

	if len(p.AllowedTEETypes) == 0 {
		return fmt.Errorf("allowed_tee_types cannot be empty")
	}

	for _, teeType := range p.AllowedTEETypes {
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

	return nil
}

// IsTEETypeAllowed checks if a TEE type is in the allowed list
func (p Params) IsTEETypeAllowed(teeType TEEType) bool {
	for _, allowed := range p.AllowedTEETypes {
		if allowed == teeType {
			return true
		}
	}
	return false
}

// GenesisState defines the enclave module's genesis state
type GenesisState struct {
	// EnclaveIdentities is the list of registered enclave identities
	EnclaveIdentities []EnclaveIdentity `json:"enclave_identities"`

	// MeasurementAllowlist is the list of approved enclave measurements
	MeasurementAllowlist []MeasurementRecord `json:"measurement_allowlist"`

	// KeyRotations is the list of active key rotations
	KeyRotations []KeyRotationRecord `json:"key_rotations,omitempty"`

	// Params is the module parameters
	Params Params `json:"params"`
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		EnclaveIdentities:    []EnclaveIdentity{},
		MeasurementAllowlist: []MeasurementRecord{},
		KeyRotations:         []KeyRotationRecord{},
		Params:               DefaultParams(),
	}
}

// Validate validates the genesis state
func (g *GenesisState) Validate() error {
	if err := g.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	seenValidators := make(map[string]bool)
	for i, identity := range g.EnclaveIdentities {
		if err := identity.Validate(); err != nil {
			return fmt.Errorf("invalid enclave identity at index %d: %w", i, err)
		}

		if seenValidators[identity.ValidatorAddress] {
			return fmt.Errorf("duplicate enclave identity for validator %s", identity.ValidatorAddress)
		}
		seenValidators[identity.ValidatorAddress] = true
	}

	seenMeasurements := make(map[string]bool)
	for i, measurement := range g.MeasurementAllowlist {
		if err := measurement.Validate(); err != nil {
			return fmt.Errorf("invalid measurement at index %d: %w", i, err)
		}

		hashHex := measurement.MeasurementHashHex()
		if seenMeasurements[hashHex] {
			return fmt.Errorf("duplicate measurement in allowlist: %s", hashHex)
		}
		seenMeasurements[hashHex] = true
	}

	for i, rotation := range g.KeyRotations {
		if err := rotation.Validate(); err != nil {
			return fmt.Errorf("invalid key rotation at index %d: %w", i, err)
		}
	}

	return nil
}

// ProtoMessage implements proto.Message
func (*GenesisState) ProtoMessage() {}

// Reset implements proto.Message
func (g *GenesisState) Reset() { *g = GenesisState{} }

// String implements proto.Message
func (g *GenesisState) String() string { return fmt.Sprintf("%+v", *g) }
