package config

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"pkg.akt.dev/node/x/config/types"
)

// InitGenesis initializes the genesis state for the config module
func InitGenesis(ctx sdk.Context, k Keeper, data *types.GenesisState) error {
	return k.InitGenesis(ctx, data)
}

// ExportGenesis exports the genesis state for the config module
func ExportGenesis(ctx sdk.Context, k Keeper) *types.GenesisState {
	return k.ExportGenesis(ctx)
}
