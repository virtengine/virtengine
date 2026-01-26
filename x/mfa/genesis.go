package mfa

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"pkg.akt.dev/node/x/mfa/keeper"
	"pkg.akt.dev/node/x/mfa/types"
)

// InitGenesis initializes the mfa module's state from a genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data *types.GenesisState) {
	k.InitGenesis(ctx, data)
}

// ExportGenesis exports the mfa module's state to a genesis state.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return k.ExportGenesis(ctx)
}
