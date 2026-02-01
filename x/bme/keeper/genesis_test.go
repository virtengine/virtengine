package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	types "github.com/virtengine/virtengine/sdk/go/node/bme/v1"
	"github.com/virtengine/virtengine/x/bme/keeper"
)

func TestGenesisInit(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Create genesis state
	genesisState := &types.GenesisState{
		Params: types.Params{
			CircuitBreakerWarnThreshold: 9600,
			CircuitBreakerHaltThreshold: 9100,
			MinEpochBlocks:              20,
			EpochBlocksBackoff:          15,
			MintSpreadBps:               30,
			SettleSpreadBps:             5,
		},
		State: types.GenesisVaultState{
			TotalBurned:   sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(100))),
			TotalMinted:   sdk.NewCoins(sdk.NewCoin("uvact", math.NewInt(90))),
			RemintCredits: sdk.NewCoins(),
		},
		Ledger: &types.GenesisLedgerState{
			Records:        []types.GenesisLedgerRecord{},
			PendingRecords: []types.GenesisLedgerPendingRecord{},
		},
	}

	// Init genesis
	keeper.InitGenesis(ctx, k, genesisState)

	// Verify params were set
	params := k.GetParams(ctx)
	require.Equal(t, genesisState.Params, params)

	// Verify state was set
	state := k.GetState(ctx)
	require.True(t, genesisState.State.TotalBurned.Equal(state.TotalBurned))
	require.True(t, genesisState.State.TotalMinted.Equal(state.TotalMinted))
}

func TestGenesisExport(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Set custom params
	customParams := types.Params{
		CircuitBreakerWarnThreshold: 9600,
		CircuitBreakerHaltThreshold: 9100,
		MinEpochBlocks:              20,
		EpochBlocksBackoff:          15,
		MintSpreadBps:               30,
		SettleSpreadBps:             5,
	}
	err := k.SetParams(ctx, customParams)
	require.NoError(t, err)

	// Set custom state
	customState := types.State{
		Balances:      sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(500))),
		TotalBurned:   sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(50))),
		TotalMinted:   sdk.NewCoins(sdk.NewCoin("uvact", math.NewInt(45))),
		RemintCredits: sdk.NewCoins(),
	}
	err = k.SetState(ctx, customState)
	require.NoError(t, err)

	// Export genesis
	exported := keeper.ExportGenesis(ctx, k)
	require.NotNil(t, exported)
	require.Equal(t, customParams, exported.Params)
	require.True(t, customState.TotalBurned.Equal(exported.State.TotalBurned))
	require.True(t, customState.TotalMinted.Equal(exported.State.TotalMinted))
}

func TestGenesisRoundTrip(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Create custom genesis state
	genesisState := &types.GenesisState{
		Params: types.Params{
			CircuitBreakerWarnThreshold: 8000,
			CircuitBreakerHaltThreshold: 7000,
			MinEpochBlocks:              100,
			EpochBlocksBackoff:          50,
			MintSpreadBps:               50,
			SettleSpreadBps:             10,
		},
		State: types.GenesisVaultState{
			TotalBurned:   sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(500))),
			TotalMinted:   sdk.NewCoins(sdk.NewCoin("uvact", math.NewInt(450))),
			RemintCredits: sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(100))),
		},
		Ledger: &types.GenesisLedgerState{
			Records:        []types.GenesisLedgerRecord{},
			PendingRecords: []types.GenesisLedgerPendingRecord{},
		},
	}

	// Init genesis
	keeper.InitGenesis(ctx, k, genesisState)

	// Export genesis
	exported := keeper.ExportGenesis(ctx, k)

	// Verify round-trip
	require.Equal(t, genesisState.Params, exported.Params)
	require.True(t, genesisState.State.TotalBurned.Equal(exported.State.TotalBurned))
	require.True(t, genesisState.State.TotalMinted.Equal(exported.State.TotalMinted))
	require.True(t, genesisState.State.RemintCredits.Equal(exported.State.RemintCredits))
}
