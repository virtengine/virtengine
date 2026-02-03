package keeper_test

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	types "github.com/virtengine/virtengine/sdk/go/node/bme/v1"
	"github.com/virtengine/virtengine/x/bme/keeper"
)

func setupKeeper(t testing.TB) (keeper.IKeeper, sdk.Context) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	k := keeper.NewKeeper(cdc, storeKey, authority, nil, nil)
	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	// Initialize with default params
	err := k.SetParams(ctx, types.DefaultParams())
	require.NoError(t, err)

	return k, ctx
}

func TestKeeperGetSetParams(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Test GetParams returns default params
	params := k.GetParams(ctx)
	require.Equal(t, types.DefaultParams(), params)

	// Test SetParams with custom params
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

	// Verify custom params are stored
	storedParams := k.GetParams(ctx)
	require.Equal(t, customParams, storedParams)
}

func TestKeeperGetSetState(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Test GetState returns empty state
	state := k.GetState(ctx)
	require.Empty(t, state.Balances)
	require.Empty(t, state.TotalBurned)
	require.Empty(t, state.TotalMinted)
	require.Empty(t, state.RemintCredits)

	// Test SetState
	customState := types.State{
		Balances:      sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(1000))),
		TotalBurned:   sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(500))),
		TotalMinted:   sdk.NewCoins(sdk.NewCoin("uvact", math.NewInt(400))),
		RemintCredits: sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(100))),
	}
	err := k.SetState(ctx, customState)
	require.NoError(t, err)

	// Verify custom state is stored
	storedState := k.GetState(ctx)
	require.True(t, customState.Balances.Equal(storedState.Balances))
	require.True(t, customState.TotalBurned.Equal(storedState.TotalBurned))
	require.True(t, customState.TotalMinted.Equal(storedState.TotalMinted))
	require.True(t, customState.RemintCredits.Equal(storedState.RemintCredits))
}

func TestQuerierParams(t *testing.T) {
	k, ctx := setupKeeper(t)

	querier := k.NewQuerier()

	resp, err := querier.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, types.DefaultParams(), resp.Params)
}

func TestQuerierVaultState(t *testing.T) {
	k, ctx := setupKeeper(t)

	querier := k.NewQuerier()

	resp, err := querier.VaultState(ctx, &types.QueryVaultStateRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQuerierStatus(t *testing.T) {
	k, ctx := setupKeeper(t)

	querier := k.NewQuerier()

	resp, err := querier.Status(ctx, &types.QueryStatusRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, types.MintStatusHealthy, resp.Status)
	require.True(t, resp.MintsAllowed)
	require.True(t, resp.RefundsAllowed)
}
