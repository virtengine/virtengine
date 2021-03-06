package simulation

import (
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/virtengine/virtengine/x/cert/types"
)

func RandomizedGenState(simState *module.SimulationState) {
	deploymentGenesis := &types.GenesisState{}

	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(deploymentGenesis)
}
