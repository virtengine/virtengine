package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ApplyGovernedVerificationResult applies a prepared verification result through the
// same governed artifact and issuance checks used by the normal verification pipeline.
func (k Keeper) ApplyGovernedVerificationResult(
	ctx sdk.Context,
	addr sdk.AccAddress,
	request *types.VerificationRequest,
	result *types.VerificationResult,
) error {
	return k.applyVerificationResult(ctx, addr, request, result)
}
