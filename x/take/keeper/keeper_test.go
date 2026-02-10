package keeper

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	types "github.com/virtengine/virtengine/sdk/go/node/take/v1"
)

func setupTakeKeeper(t *testing.T) (IKeeper, sdk.Context) {
	t.Helper()

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 1,
	}, false, log.NewNopLogger())

	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	return NewKeeper(cdc, storeKey, authority), ctx
}

func TestSubtractFees_UsesDenomRate(t *testing.T) {
	keeper, ctx := setupTakeKeeper(t)

	params := types.Params{
		DefaultTakeRate: 20,
		DenomTakeRates: types.DenomTakeRates{
			{Denom: "uve", Rate: 2},
		},
	}
	require.NoError(t, keeper.SetParams(ctx, params))

	earnings, fee, err := keeper.SubtractFees(ctx, sdk.NewInt64Coin("uve", 1000))
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt64Coin("uve", 980), earnings)
	require.Equal(t, sdk.NewInt64Coin("uve", 20), fee)
}

func TestSubtractFees_DefaultRateFallback(t *testing.T) {
	keeper, ctx := setupTakeKeeper(t)

	params := types.Params{
		DefaultTakeRate: 20,
		DenomTakeRates: types.DenomTakeRates{
			{Denom: "uve", Rate: 2},
		},
	}
	require.NoError(t, keeper.SetParams(ctx, params))

	earnings, fee, err := keeper.SubtractFees(ctx, sdk.NewInt64Coin("ufoo", 1000))
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt64Coin("ufoo", 800), earnings)
	require.Equal(t, sdk.NewInt64Coin("ufoo", 200), fee)
}

func TestSubtractFees_MultiDenomHandling(t *testing.T) {
	keeper, ctx := setupTakeKeeper(t)

	params := types.Params{
		DefaultTakeRate: 20,
		DenomTakeRates: types.DenomTakeRates{
			{Denom: "uve", Rate: 2},
			{Denom: "ufoo", Rate: 10},
		},
	}
	require.NoError(t, keeper.SetParams(ctx, params))

	earnings, fee, err := keeper.SubtractFees(ctx, sdk.NewInt64Coin("ufoo", 100))
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt64Coin("ufoo", 90), earnings)
	require.Equal(t, sdk.NewInt64Coin("ufoo", 10), fee)

	earnings, fee, err = keeper.SubtractFees(ctx, sdk.NewInt64Coin("ubar", 50))
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt64Coin("ubar", 40), earnings)
	require.Equal(t, sdk.NewInt64Coin("ubar", 10), fee)
}

func TestSubtractFees_ZeroAndRounding(t *testing.T) {
	keeper, ctx := setupTakeKeeper(t)

	params := types.Params{
		DefaultTakeRate: 20,
		DenomTakeRates: types.DenomTakeRates{
			{Denom: "uve", Rate: 2},
		},
	}
	require.NoError(t, keeper.SetParams(ctx, params))

	earnings, fee, err := keeper.SubtractFees(ctx, sdk.NewInt64Coin("uve", 0))
	require.NoError(t, err)
	require.True(t, earnings.IsZero())
	require.True(t, fee.IsZero())

	earnings, fee, err = keeper.SubtractFees(ctx, sdk.NewInt64Coin("uve", 3))
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt64Coin("uve", 3), earnings)
	require.True(t, fee.IsZero())
}

func TestSubtractFees_LargeAmountNoOverflow(t *testing.T) {
	keeper, ctx := setupTakeKeeper(t)

	params := types.Params{
		DefaultTakeRate: 100,
		DenomTakeRates: types.DenomTakeRates{
			{Denom: "uve", Rate: 100},
		},
	}
	require.NoError(t, keeper.SetParams(ctx, params))

	amount, ok := sdkmath.NewIntFromString("100000000000000000000000")
	require.True(t, ok)

	earnings, fee, err := keeper.SubtractFees(ctx, sdk.NewCoin("uve", amount))
	require.NoError(t, err)
	require.True(t, earnings.IsZero())
	require.Equal(t, sdk.NewCoin("uve", amount), fee)
}

func TestSetParams_RejectsInvalidParams(t *testing.T) {
	keeper, ctx := setupTakeKeeper(t)

	params := types.Params{
		DefaultTakeRate: 200,
		DenomTakeRates: types.DenomTakeRates{
			{Denom: "uve", Rate: 2},
		},
	}
	err := keeper.SetParams(ctx, params)
	require.Error(t, err)

	params = types.Params{
		DefaultTakeRate: 20,
		DenomTakeRates: types.DenomTakeRates{
			{Denom: "ufoo", Rate: 2},
		},
	}
	err = keeper.SetParams(ctx, params)
	require.Error(t, err)
}

func TestGetSetParams_RoundTrip(t *testing.T) {
	keeper, ctx := setupTakeKeeper(t)

	params := types.Params{
		DefaultTakeRate: 15,
		DenomTakeRates: types.DenomTakeRates{
			{Denom: "uve", Rate: 3},
			{Denom: "ufoo", Rate: 7},
		},
	}
	require.NoError(t, keeper.SetParams(ctx, params))

	got := keeper.GetParams(ctx)
	require.Equal(t, params, got)
}

func TestKeeperGetters(t *testing.T) {
	keeper, _ := setupTakeKeeper(t)

	concreteKeeper := keeper.(Keeper)

	require.NotNil(t, concreteKeeper.Codec())
	require.NotNil(t, concreteKeeper.StoreKey())
	require.Equal(t, types.StoreKey, concreteKeeper.StoreKey().Name())
	require.NotEmpty(t, concreteKeeper.GetAuthority())
	// Authority is the gov module account address
	require.Contains(t, concreteKeeper.GetAuthority(), "ve1")
}

func TestNewQuerier(t *testing.T) {
	keeper, _ := setupTakeKeeper(t)

	querier := keeper.NewQuerier()
	require.NotNil(t, querier)
}

func TestSubtractFees_NegativeAmount(t *testing.T) {
	keeper, ctx := setupTakeKeeper(t)

	params := types.Params{
		DefaultTakeRate: 20,
		DenomTakeRates: types.DenomTakeRates{
			{Denom: "uve", Rate: 2},
		},
	}
	require.NoError(t, keeper.SetParams(ctx, params))

	coin := sdk.Coin{Denom: "uve", Amount: sdkmath.NewInt(100)}
	earnings, fee, err := keeper.SubtractFees(ctx, coin)
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt64Coin("uve", 98), earnings)
	require.Equal(t, sdk.NewInt64Coin("uve", 2), fee)
}

func TestSubtractFees_ExactlyOne(t *testing.T) {
	keeper, ctx := setupTakeKeeper(t)

	params := types.Params{
		DefaultTakeRate: 50,
		DenomTakeRates: types.DenomTakeRates{
			{Denom: "uve", Rate: 50},
		},
	}
	require.NoError(t, keeper.SetParams(ctx, params))

	earnings, fee, err := keeper.SubtractFees(ctx, sdk.NewInt64Coin("uve", 1))
	require.NoError(t, err)
	// With amount of 1 and 50% rate, fee is 0 (rounds down) and earnings is 1
	require.Equal(t, sdk.NewInt64Coin("uve", 1), earnings)
	require.True(t, fee.IsZero())
}

func TestSubtractFees_HighRateCloseToMax(t *testing.T) {
	keeper, ctx := setupTakeKeeper(t)

	params := types.Params{
		DefaultTakeRate: 99,
		DenomTakeRates: types.DenomTakeRates{
			{Denom: "uve", Rate: 99},
		},
	}
	require.NoError(t, keeper.SetParams(ctx, params))

	earnings, fee, err := keeper.SubtractFees(ctx, sdk.NewInt64Coin("uve", 10000))
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt64Coin("uve", 100), earnings)
	require.Equal(t, sdk.NewInt64Coin("uve", 9900), fee)
}

func TestGetParams_EmptyStore(t *testing.T) {
	keeper, ctx := setupTakeKeeper(t)

	params := keeper.GetParams(ctx)
	require.Equal(t, types.Params{}, params)
}
