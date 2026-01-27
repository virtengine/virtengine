package simulation

import (
	"github.com/cosmos/cosmos-sdk/types/module"
	mv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"

	dtypes "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4"
	types "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
)

var minDeposit, _ = dtypes.DefaultParams().MinDepositFor("uakt")

// RandomizedGenState generates a random GenesisState for supply
func RandomizedGenState(simState *module.SimulationState) {
	marketGenesis := &types.GenesisState{
		Params: types.Params{
			BidMinDeposit: minDeposit,
			OrderMaxBids:  20,
		},
	}

	simState.GenState[mv1.ModuleName] = simState.Cdc.MustMarshalJSON(marketGenesis)
}
