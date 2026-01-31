package keeper

import (
	"crypto/sha256"
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/enclave/types"
)

// GetCommitteeEnclaveKeys returns enclave keys for the identity committee
// This implements deterministic committee selection based on epoch
func (k Keeper) GetCommitteeEnclaveKeys(ctx sdk.Context, epoch uint64) []types.EnclaveIdentity {
	params := k.GetParams(ctx)
	if !params.EnableCommitteeMode {
		return k.GetActiveValidatorEnclaveKeys(ctx)
	}

	allKeys := k.GetActiveValidatorEnclaveKeys(ctx)
	// #nosec G115 - len() is always non-negative, safe conversion to uint32
	if uint32(len(allKeys)) <= params.CommitteeSize {
		return allKeys
	}

	// Deterministic selection using cryptographic seed
	// Use Fisher-Yates shuffle with crypto-derived randomness
	seed := k.computeCommitteeSeed(ctx, epoch)

	shuffled := make([]types.EnclaveIdentity, len(allKeys))
	copy(shuffled, allKeys)

	// Fisher-Yates shuffle using cryptographic hash for index selection
	for i := len(shuffled) - 1; i > 0; i-- {
		// Derive deterministic random index using HKDF-style expansion
		j := cryptoRandIndex(seed, uint64(i), i+1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled[:params.CommitteeSize]
}

// computeCommitteeSeed creates a deterministic seed for committee selection
func (k Keeper) computeCommitteeSeed(ctx sdk.Context, epoch uint64) []byte {
	h := sha256.New()

	// Include chain ID for cross-chain uniqueness
	h.Write([]byte(ctx.ChainID()))

	// Include committee selection prefix
	h.Write([]byte("enclave-committee-selection"))

	// Include epoch
	epochBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(epochBytes, epoch)
	h.Write(epochBytes)

	// Include block hash for additional entropy
	// Note: We use the previous block hash to ensure determinism
	// across all validators
	if ctx.BlockHeight() > 0 {
		prevBlockHash := ctx.BlockHeader().LastBlockId.Hash
		h.Write(prevBlockHash)
	}

	return h.Sum(nil)
}

// cryptoRandIndex generates a deterministic random index in [0, max) using
// cryptographic hash expansion. Uses rejection sampling to avoid modulo bias.
func cryptoRandIndex(seed []byte, counter uint64, max int) int {
	if max <= 0 {
		return 0
	}
	if max == 1 {
		return 0
	}

	// Use rejection sampling to avoid modulo bias
	// We need enough bits to cover max, then reject values >= max
	maxUint64 := uint64(max)

	// Compute how many times max fits in 2^64
	// We reject values >= (2^64 / max) * max to avoid bias
	limit := (^uint64(0) / maxUint64) * maxUint64

	for attempt := uint64(0); ; attempt++ {
		// Hash seed || counter || attempt to get random bytes
		h := sha256.New()
		h.Write(seed)

		counterBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(counterBytes, counter)
		h.Write(counterBytes)

		attemptBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(attemptBytes, attempt)
		h.Write(attemptBytes)

		hash := h.Sum(nil)
		val := binary.BigEndian.Uint64(hash[:8])

		if val < limit {
			// #nosec G115 - result is always < max which is an int, safe conversion
			return int(val % maxUint64)
		}
		// Rejection: try again with next attempt value
	}
}

// GetCommitteeEpoch returns the current committee epoch based on block height
func (k Keeper) GetCommitteeEpoch(ctx sdk.Context) uint64 {
	params := k.GetParams(ctx)
	if !params.EnableCommitteeMode {
		return 0
	}

	// Each epoch lasts for CommitteeEpochBlocks blocks
	// Default: 10000 blocks (~14 hours at 5s block time)
	epochBlocks := int64(10000)
	if params.CommitteeEpochBlocks > 0 {
		epochBlocks = params.CommitteeEpochBlocks
	}

	// #nosec G115 - block heights are always non-negative, safe conversion
	return uint64(ctx.BlockHeight() / epochBlocks)
}

// IsValidatorInCommittee checks if a validator is in the current committee
func (k Keeper) IsValidatorInCommittee(ctx sdk.Context, validatorAddr sdk.AccAddress) bool {
	params := k.GetParams(ctx)
	if !params.EnableCommitteeMode {
		// If committee mode is disabled, all validators are "in committee"
		return true
	}

	epoch := k.GetCommitteeEpoch(ctx)
	committee := k.GetCommitteeEnclaveKeys(ctx, epoch)

	validatorAddrStr := validatorAddr.String()
	for _, member := range committee {
		if member.ValidatorAddress == validatorAddrStr {
			return true
		}
	}

	return false
}
