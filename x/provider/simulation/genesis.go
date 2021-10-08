package simulation

import (
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/virtengine/virtengine/x/provider/types"
)

// RandomizedGenState generates a random GenesisState for supply
func RandomizedGenState(simState *module.SimulationState) {
	providerGenesis := &types.GenesisState{}

	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(providerGenesis)
}
