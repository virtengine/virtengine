package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	taketype "github.com/virtengine/virtengine/sdk/go/node/take/v1"
	"github.com/virtengine/virtengine/x/take"
)

func TestGenesisState_DefaultParamsValid(t *testing.T) {
	gs := &taketype.GenesisState{
		Params: taketype.DefaultParams(),
	}

	require.NoError(t, take.ValidateGenesis(gs))
}

func TestGenesisState_CustomParamsValid(t *testing.T) {
	gs := &taketype.GenesisState{
		Params: taketype.Params{
			DefaultTakeRate: 18,
			DenomTakeRates: taketype.DenomTakeRates{
				{Denom: "uve", Rate: 4},
				{Denom: "ufoo", Rate: 7},
			},
		},
	}

	require.NoError(t, take.ValidateGenesis(gs))
}
