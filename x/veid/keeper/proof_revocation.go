package keeper

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// RevokeSelectiveDisclosureProof revokes a selective disclosure proof by ID.
func (k Keeper) RevokeSelectiveDisclosureProof(ctx sdk.Context, proofID string, reason string) error {
	if proofID == "" {
		return types.ErrInvalidProof.Wrap("proof_id cannot be empty")
	}

	revocation := types.ProofRevocation{
		ProofID:   proofID,
		RevokedAt: ctx.BlockTime().Unix(),
		Reason:    reason,
	}

	bz, err := json.Marshal(&revocation)
	if err != nil {
		return types.ErrProofGenerationFailed.Wrap(err.Error())
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.ProofRevocationKey(proofID), bz)
	return nil
}

// GetProofRevocation returns a proof revocation record if present.
func (k Keeper) GetProofRevocation(ctx sdk.Context, proofID string) (types.ProofRevocation, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ProofRevocationKey(proofID))
	if bz == nil {
		return types.ProofRevocation{}, false
	}

	var revocation types.ProofRevocation
	if err := json.Unmarshal(bz, &revocation); err != nil {
		return types.ProofRevocation{}, false
	}

	return revocation, true
}

// IsProofRevoked returns true if the proof has been revoked.
func (k Keeper) IsProofRevoked(ctx sdk.Context, proofID string) bool {
	_, found := k.GetProofRevocation(ctx, proofID)
	return found
}
