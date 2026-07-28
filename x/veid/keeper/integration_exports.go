package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ApplyGovernedVerificationResult exposes the governed artifact/issuance
// integration boundary. The underlying method rejects every call that is not
// an authorized FinalizeBlock consensus system transaction, so this does not
// permit local or ordinary-transaction score mutation.
func (k Keeper) ApplyGovernedVerificationResult(
	ctx sdk.Context,
	addr sdk.AccAddress,
	request *types.VerificationRequest,
	result *types.VerificationResult,
) error {
	return k.applyVerificationResult(ctx, addr, request, result)
}
