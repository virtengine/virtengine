package simulation

import (
	"github.com/cosmos/cosmos-sdk/types/module"

	types "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
)

// RandomizedGenState generates a random GenesisState for supply
func RandomizedGenState(simState *module.SimulationState) {
	providerGenesis := &types.GenesisState{}

	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(providerGenesis)
}
