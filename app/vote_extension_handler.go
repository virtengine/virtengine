package app

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	veidkeeper "github.com/virtengine/virtengine/x/veid/keeper"
)

const activeVEIDVoteExtensionCarrierVersion = veidkeeper.ActiveVoteExtensionCarrierVersion

// newVEIDVoteExtensionHandlers returns the active canonical protobuf carrier.
func newVEIDVoteExtensionHandlers(keeper *veidkeeper.Keeper) (sdk.ExtendVoteHandler, sdk.VerifyVoteExtensionHandler) {
	if keeper == nil {
		panic("VEID keeper is required")
	}
	return func(ctx sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
			return keeper.ExtendVote(ctx, req, nil)
		}, func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
			return keeper.VerifyVoteExtension(ctx, req, nil)
		}
}
