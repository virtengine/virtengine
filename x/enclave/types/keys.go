package types

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "enclave"

	// StoreKey is the store key string for enclave module
	StoreKey = ModuleName

	// RouterKey is the message route for enclave module
	RouterKey = ModuleName

	// QuerierRoute is the querier route for enclave module
	QuerierRoute = ModuleName
)

// Store key prefixes
var (
	// PrefixEnclaveIdentity is the prefix for enclave identity storage
	// Key: PrefixEnclaveIdentity | validator_address -> EnclaveIdentity
	PrefixEnclaveIdentity = []byte{0x01}

	// PrefixMeasurementAllowlist is the prefix for approved enclave measurements
	// Key: PrefixMeasurementAllowlist | measurement_hash -> MeasurementRecord
	PrefixMeasurementAllowlist = []byte{0x02}

	// PrefixEnclaveKeyByFingerprint is the prefix for looking up validator by enclave key fingerprint
	// Key: PrefixEnclaveKeyByFingerprint | fingerprint -> validator_address
	PrefixEnclaveKeyByFingerprint = []byte{0x03}

	// PrefixParams is the prefix for module parameters
	PrefixParams = []byte{0x04}

	// PrefixKeyRotation is the prefix for key rotation records
	// Key: PrefixKeyRotation | validator_address | epoch -> KeyRotationRecord
	PrefixKeyRotation = []byte{0x05}

	// PrefixAttestedResult is the prefix for attested scoring results
	// Key: PrefixAttestedResult | block_height | scope_id -> AttestedScoringResult
	PrefixAttestedResult = []byte{0x06}

	// PrefixEnclaveHealth is the prefix for enclave health status
	// Key: PrefixEnclaveHealth | validator_address -> EnclaveHealthStatus
	PrefixEnclaveHealth = []byte{0x07}

	// PrefixHealthCheckParams is the prefix for health check parameters
	PrefixHealthCheckParams = []byte{0x08}
)

// EnclaveIdentityKey returns the store key for a validator's enclave identity
func EnclaveIdentityKey(validatorAddr []byte) []byte {
	key := make([]byte, 0, len(PrefixEnclaveIdentity)+len(validatorAddr))
	key = append(key, PrefixEnclaveIdentity...)
	key = append(key, validatorAddr...)
	return key
}

// MeasurementAllowlistKey returns the store key for a measurement in the allowlist
func MeasurementAllowlistKey(measurementHash []byte) []byte {
	key := make([]byte, 0, len(PrefixMeasurementAllowlist)+len(measurementHash))
	key = append(key, PrefixMeasurementAllowlist...)
	key = append(key, measurementHash...)
	return key
}

// EnclaveKeyByFingerprintKey returns the store key for looking up validator by key fingerprint
func EnclaveKeyByFingerprintKey(fingerprint []byte) []byte {
	key := make([]byte, 0, len(PrefixEnclaveKeyByFingerprint)+len(fingerprint))
	key = append(key, PrefixEnclaveKeyByFingerprint...)
	key = append(key, fingerprint...)
	return key
}

// ParamsKey returns the store key for module parameters
func ParamsKey() []byte {
	return PrefixParams
}

// KeyRotationKey returns the store key for a key rotation record
func KeyRotationKey(validatorAddr []byte, epoch uint64) []byte {
	key := make([]byte, 0, len(PrefixKeyRotation)+len(validatorAddr)+8)
	key = append(key, PrefixKeyRotation...)
	key = append(key, validatorAddr...)
	// Append epoch as big-endian bytes
	key = append(key,
		byte(epoch>>56), byte(epoch>>48), byte(epoch>>40), byte(epoch>>32),
		byte(epoch>>24), byte(epoch>>16), byte(epoch>>8), byte(epoch),
	)
	return key
}

// AttestedResultKey returns the store key for an attested scoring result
func AttestedResultKey(blockHeight int64, scopeID string) []byte {
	scopeIDBytes := []byte(scopeID)
	key := make([]byte, 0, len(PrefixAttestedResult)+8+len(scopeIDBytes))
	key = append(key, PrefixAttestedResult...)
	// Append block height as big-endian bytes
	key = append(key,
		byte(blockHeight>>56), byte(blockHeight>>48), byte(blockHeight>>40), byte(blockHeight>>32),
		byte(blockHeight>>24), byte(blockHeight>>16), byte(blockHeight>>8), byte(blockHeight),
	)
	key = append(key, scopeIDBytes...)
	return key
}

// EnclaveHealthKey returns the store key for an enclave health status
func EnclaveHealthKey(validatorAddr []byte) []byte {
	key := make([]byte, 0, len(PrefixEnclaveHealth)+len(validatorAddr))
	key = append(key, PrefixEnclaveHealth...)
	key = append(key, validatorAddr...)
	return key
}

// HealthCheckParamsKey returns the store key for health check parameters
func HealthCheckParamsKey() []byte {
	return PrefixHealthCheckParams
}
