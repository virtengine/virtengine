package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/virtengine/virtengine/sdk/go/node/bme/v1"
)

// InitGenesis initializes the BME module's state from a genesis state.
func InitGenesis(ctx sdk.Context, keeper IKeeper, data *types.GenesisState) {
	if err := keeper.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	// Initialize vault state from genesis
	state := types.State{
		Balances:      sdk.Coins{},
		TotalBurned:   data.State.TotalBurned,
		TotalMinted:   data.State.TotalMinted,
		RemintCredits: data.State.RemintCredits,
	}

	if err := keeper.SetState(ctx, state); err != nil {
		panic(err)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper IKeeper) *types.GenesisState {
	params := keeper.GetParams(ctx)
	state := keeper.GetState(ctx)

	return &types.GenesisState{
		Params: params,
		State: types.GenesisVaultState{
			TotalBurned:   state.TotalBurned,
			TotalMinted:   state.TotalMinted,
			RemintCredits: state.RemintCredits,
		},
	}
}
