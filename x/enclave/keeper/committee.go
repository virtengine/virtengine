package keeper

import (
	"crypto/sha256"
	"encoding/binary"
	"math/rand"

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
	if uint32(len(allKeys)) <= params.CommitteeSize {
		return allKeys
	}

	// Deterministic selection using epoch and block hash as seed
	seed := k.computeCommitteeSeed(ctx, epoch)

	// Fisher-Yates shuffle with deterministic RNG
	rng := rand.New(rand.NewSource(int64(binary.BigEndian.Uint64(seed[:]))))
	shuffled := make([]types.EnclaveIdentity, len(allKeys))
	copy(shuffled, allKeys)

	for i := len(shuffled) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
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
