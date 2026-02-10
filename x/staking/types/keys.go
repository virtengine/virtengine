// Package types contains types for the staking module.
//
// VE-921: Staking rewards for validators including identity network rewards and slashing
package types

const (
	// ModuleName is the name of the staking module
	ModuleName = "virt_staking"

	// StoreKey is the store key for the staking module
	StoreKey = ModuleName

	// RouterKey is the router key for the staking module
	RouterKey = ModuleName

	// QuerierRoute is the querier route for the staking module
	QuerierRoute = ModuleName
)

// Key prefixes for staking store
var (
	// ValidatorPerformancePrefix is the prefix for validator performance storage
	ValidatorPerformancePrefix = []byte{0x01}

	// SlashingRecordPrefix is the prefix for slashing records
	SlashingRecordPrefix = []byte{0x02}

	// RewardEpochPrefix is the prefix for reward epoch storage
	RewardEpochPrefix = []byte{0x03}

	// ValidatorRewardPrefix is the prefix for validator reward storage
	ValidatorRewardPrefix = []byte{0x04}

	// ParamsKey is the key for module parameters
	ParamsKey = []byte{0x10}

	// SequenceKeySlash is the sequence key for slashing records
	SequenceKeySlash = []byte{0x20}

	// SequenceKeyRewardEpoch is the sequence key for reward epochs
	SequenceKeyRewardEpoch = []byte{0x21}

	// ValidatorMissedBlocksPrefix is the prefix for missed blocks tracking
	ValidatorMissedBlocksPrefix = []byte{0x30}

	// ValidatorSigningInfoPrefix is the prefix for signing info
	ValidatorSigningInfoPrefix = []byte{0x31}

	// DoubleSignEvidencePrefix is the prefix for double sign evidence
	DoubleSignEvidencePrefix = []byte{0x32}

	// InvalidAttestationPrefix is the prefix for invalid VEID attestations
	InvalidAttestationPrefix = []byte{0x33}
)

// GetValidatorPerformanceKey returns the key for a validator's performance record
func GetValidatorPerformanceKey(validatorAddr string) []byte {
	return append(ValidatorPerformancePrefix, []byte(validatorAddr)...)
}

// GetSlashingRecordKey returns the key for a slashing record
func GetSlashingRecordKey(slashID string) []byte {
	return append(SlashingRecordPrefix, []byte(slashID)...)
}

// GetRewardEpochKey returns the key for a reward epoch
func GetRewardEpochKey(epochNumber uint64) []byte {
	key := make([]byte, len(RewardEpochPrefix)+8)
	copy(key, RewardEpochPrefix)
	putUint64BE(key[len(RewardEpochPrefix):], epochNumber)
	return key
}

// GetValidatorRewardKey returns the key for a validator's reward record
func GetValidatorRewardKey(validatorAddr string, epochNumber uint64) []byte {
	key := make([]byte, len(ValidatorRewardPrefix)+len(validatorAddr)+1+8)
	copy(key, ValidatorRewardPrefix)
	key = append(key[:len(ValidatorRewardPrefix)], []byte(validatorAddr)...)
	key = append(key, byte(':'))
	putUint64BE(key[len(key)-8:], epochNumber)
	return key
}

// GetValidatorMissedBlocksKey returns the key for validator missed blocks
func GetValidatorMissedBlocksKey(validatorAddr string) []byte {
	return append(ValidatorMissedBlocksPrefix, []byte(validatorAddr)...)
}

// GetValidatorSigningInfoKey returns the key for validator signing info
func GetValidatorSigningInfoKey(validatorAddr string) []byte {
	return append(ValidatorSigningInfoPrefix, []byte(validatorAddr)...)
}

// GetDoubleSignEvidenceKey returns the key for double sign evidence
func GetDoubleSignEvidenceKey(evidenceID string) []byte {
	return append(DoubleSignEvidencePrefix, []byte(evidenceID)...)
}

// GetInvalidAttestationKey returns the key for invalid attestation records
func GetInvalidAttestationKey(recordID string) []byte {
	return append(InvalidAttestationPrefix, []byte(recordID)...)
}

// putUint64BE puts a big-endian uint64 into the slice
func putUint64BE(b []byte, v uint64) {
	b[0] = byte(v >> 56)
	b[1] = byte(v >> 48)
	b[2] = byte(v >> 40)
	b[3] = byte(v >> 32)
	b[4] = byte(v >> 24)
	b[5] = byte(v >> 16)
	b[6] = byte(v >> 8)
	b[7] = byte(v)
}
