package types

// Query request/response types for enclave module

// QueryEnclaveIdentityRequest is the request for querying an enclave identity
type QueryEnclaveIdentityRequest struct {
	ValidatorAddress string `json:"validator_address"`
}

// QueryEnclaveIdentityResponse is the response for querying an enclave identity
type QueryEnclaveIdentityResponse struct {
	Identity *EnclaveIdentity `json:"identity,omitempty"`
}

// QueryActiveValidatorEnclaveKeysRequest is the request for querying active validator enclave keys
type QueryActiveValidatorEnclaveKeysRequest struct {
	// Empty - returns all active validator enclave keys
}

// QueryActiveValidatorEnclaveKeysResponse is the response for active validator enclave keys
type QueryActiveValidatorEnclaveKeysResponse struct {
	Identities []EnclaveIdentity `json:"identities"`
}

// QueryCommitteeEnclaveKeysRequest is the request for querying committee enclave keys
type QueryCommitteeEnclaveKeysRequest struct {
	// CommitteeEpoch is the epoch for which to get committee keys
	CommitteeEpoch uint64 `json:"committee_epoch,omitempty"`
}

// QueryCommitteeEnclaveKeysResponse is the response for committee enclave keys
type QueryCommitteeEnclaveKeysResponse struct {
	Identities []EnclaveIdentity `json:"identities"`
}

// QueryMeasurementAllowlistRequest is the request for querying the measurement allowlist
type QueryMeasurementAllowlistRequest struct {
	// TEEType optionally filters by TEE type
	TEEType string `json:"tee_type,omitempty"`

	// IncludeRevoked optionally includes revoked measurements
	IncludeRevoked bool `json:"include_revoked,omitempty"`
}

// QueryMeasurementAllowlistResponse is the response for the measurement allowlist
type QueryMeasurementAllowlistResponse struct {
	Measurements []MeasurementRecord `json:"measurements"`
}

// QueryMeasurementRequest is the request for querying a specific measurement
type QueryMeasurementRequest struct {
	MeasurementHash []byte `json:"measurement_hash"`
}

// QueryMeasurementResponse is the response for a specific measurement
type QueryMeasurementResponse struct {
	Measurement *MeasurementRecord `json:"measurement,omitempty"`
	IsAllowed   bool               `json:"is_allowed"`
}

// QueryKeyRotationRequest is the request for querying key rotation status
type QueryKeyRotationRequest struct {
	ValidatorAddress string `json:"validator_address"`
}

// QueryKeyRotationResponse is the response for key rotation status
type QueryKeyRotationResponse struct {
	Rotation       *KeyRotationRecord `json:"rotation,omitempty"`
	HasActiveRotation bool            `json:"has_active_rotation"`
}

// QueryValidKeySetRequest is the request for querying current valid key set
type QueryValidKeySetRequest struct {
	// ForBlockHeight is the block height to check validity for
	ForBlockHeight int64 `json:"for_block_height,omitempty"`
}

// ValidatorKeyInfo contains key information for a validator
type ValidatorKeyInfo struct {
	ValidatorAddress    string `json:"validator_address"`
	EncryptionKeyID     string `json:"encryption_key_id"`
	EncryptionPubKey    []byte `json:"encryption_pub_key"`
	MeasurementHash     []byte `json:"measurement_hash"`
	ExpiryHeight        int64  `json:"expiry_height"`
	IsInRotation        bool   `json:"is_in_rotation"`
}

// QueryValidKeySetResponse is the response for current valid key set
type QueryValidKeySetResponse struct {
	ValidatorKeys []ValidatorKeyInfo `json:"validator_keys"`
	TotalCount    int                `json:"total_count"`
}

// QueryParamsRequest is the request for querying module parameters
type QueryParamsRequest struct{}

// QueryParamsResponse is the response for module parameters
type QueryParamsResponse struct {
	Params Params `json:"params"`
}

// QueryAttestedResultRequest is the request for querying an attested result
type QueryAttestedResultRequest struct {
	BlockHeight int64  `json:"block_height"`
	ScopeID     string `json:"scope_id"`
}

// QueryAttestedResultResponse is the response for an attested result
type QueryAttestedResultResponse struct {
	Result *AttestedScoringResult `json:"result,omitempty"`
}
