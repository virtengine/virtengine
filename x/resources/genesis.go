package resources

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/resources/keeper"
	"github.com/virtengine/virtengine/x/resources/types"
)

// InitGenesis initializes the module state from genesis data.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, gs *types.GenesisState) {
	if gs == nil {
		gs = types.DefaultGenesisState()
	}
	if err := k.SetParams(ctx, gs.Params); err != nil {
		panic(err)
	}

	for _, inventory := range gs.Inventories {
		if err := k.SetInventory(ctx, inventory); err != nil {
			panic(err)
		}
	}
	for _, allocation := range gs.Allocations {
		if err := k.SetAllocation(ctx, allocation); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis exports module state to genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	gs := types.DefaultGenesisState()
	gs.Params = k.GetParams(ctx)

	k.WithInventories(ctx, func(inv types.ResourceInventory) bool {
		gs.Inventories = append(gs.Inventories, inv)
		return false
	})
	k.WithAllocations(ctx, func(allocation types.ResourceAllocation) bool {
		gs.Allocations = append(gs.Allocations, allocation)
		return false
	})

	return gs
}

// MustMarshalGenesis marshals genesis state.
func MustMarshalGenesis(gs *types.GenesisState) []byte {
	bz, err := json.Marshal(gs)
	if err != nil {
		panic(err)
	}
	return bz
}
