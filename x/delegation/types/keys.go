// Package types contains types for the delegation module.
//
// VE-922: Delegated staking for non-validators
package types

const (
	// ModuleName is the name of the delegation module
	ModuleName = "virt_delegation"

	// StoreKey is the store key for the delegation module
	StoreKey = ModuleName

	// RouterKey is the router key for the delegation module
	RouterKey = ModuleName

	// QuerierRoute is the querier route for the delegation module
	QuerierRoute = ModuleName
)

// Key prefixes for delegation store
var (
	// DelegationPrefix is the prefix for delegation storage
	DelegationPrefix = []byte{0x01}

	// UnbondingDelegationPrefix is the prefix for unbonding delegation storage
	UnbondingDelegationPrefix = []byte{0x02}

	// RedelegationPrefix is the prefix for redelegation storage
	RedelegationPrefix = []byte{0x03}

	// ValidatorDelegationsPrefix is the prefix for validator delegations index
	ValidatorDelegationsPrefix = []byte{0x04}

	// DelegatorValidatorsPrefix is the prefix for delegator validators index
	DelegatorValidatorsPrefix = []byte{0x05}

	// DelegatorRewardsPrefix is the prefix for delegator rewards
	DelegatorRewardsPrefix = []byte{0x06}

	// ValidatorSharesPrefix is the prefix for validator total shares
	ValidatorSharesPrefix = []byte{0x07}

	// UnbondingQueuePrefix is the prefix for unbonding queue
	UnbondingQueuePrefix = []byte{0x08}

	// RedelegationQueuePrefix is the prefix for redelegation queue
	RedelegationQueuePrefix = []byte{0x09}

	// ParamsKey is the key for module parameters
	ParamsKey = []byte{0x10}

	// SequenceKeyDelegation is the sequence key for delegation IDs
	SequenceKeyDelegation = []byte{0x20}

	// SequenceKeyUnbonding is the sequence key for unbonding IDs
	SequenceKeyUnbonding = []byte{0x21}

	// SequenceKeyRedelegation is the sequence key for redelegation IDs
	SequenceKeyRedelegation = []byte{0x22}
)

// GetDelegationKey returns the key for a delegation
func GetDelegationKey(delegatorAddr, validatorAddr string) []byte {
	return append(DelegationPrefix, []byte(delegatorAddr+":"+validatorAddr)...)
}

// GetUnbondingDelegationKey returns the key for an unbonding delegation
func GetUnbondingDelegationKey(unbondingID string) []byte {
	return append(UnbondingDelegationPrefix, []byte(unbondingID)...)
}

// GetRedelegationKey returns the key for a redelegation
func GetRedelegationKey(redelegationID string) []byte {
	return append(RedelegationPrefix, []byte(redelegationID)...)
}

// GetValidatorDelegationsKey returns the key for validator delegations index
func GetValidatorDelegationsKey(validatorAddr string) []byte {
	return append(ValidatorDelegationsPrefix, []byte(validatorAddr)...)
}

// GetDelegatorValidatorsKey returns the key for delegator validators index
func GetDelegatorValidatorsKey(delegatorAddr string) []byte {
	return append(DelegatorValidatorsPrefix, []byte(delegatorAddr)...)
}

// GetDelegatorRewardsKey returns the key for delegator rewards
func GetDelegatorRewardsKey(delegatorAddr, validatorAddr string, epoch uint64) []byte {
	addrPart := delegatorAddr + ":" + validatorAddr + ":"
	key := make([]byte, 0, len(DelegatorRewardsPrefix)+len(addrPart)+8)
	key = append(key, DelegatorRewardsPrefix...)
	key = append(key, []byte(addrPart)...)
	return append(key, uint64ToBytes(epoch)...)
}

// GetValidatorSharesKey returns the key for validator total shares
func GetValidatorSharesKey(validatorAddr string) []byte {
	return append(ValidatorSharesPrefix, []byte(validatorAddr)...)
}

// GetUnbondingQueueKey returns the key for an unbonding queue entry
func GetUnbondingQueueKey(completionTime int64) []byte {
	return append(UnbondingQueuePrefix, int64ToBytes(completionTime)...)
}

// GetRedelegationQueueKey returns the key for a redelegation queue entry
func GetRedelegationQueueKey(completionTime int64) []byte {
	return append(RedelegationQueuePrefix, int64ToBytes(completionTime)...)
}

// uint64ToBytes converts uint64 to big-endian bytes
func uint64ToBytes(n uint64) []byte {
	b := make([]byte, 8)
	b[0] = byte(n >> 56)
	b[1] = byte(n >> 48)
	b[2] = byte(n >> 40)
	b[3] = byte(n >> 32)
	b[4] = byte(n >> 24)
	b[5] = byte(n >> 16)
	b[6] = byte(n >> 8)
	b[7] = byte(n)
	return b
}

// int64ToBytes converts int64 to big-endian bytes
func int64ToBytes(n int64) []byte {
	if n < 0 {
		return uint64ToBytes(0)
	}
	return uint64ToBytes(uint64(n))
}
