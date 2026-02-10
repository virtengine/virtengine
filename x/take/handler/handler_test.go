package handler_test

import (
	"testing"
	"time"

	"cosmossdk.io/log"
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
	"github.com/virtengine/virtengine/x/take/handler"
	tkeeper "github.com/virtengine/virtengine/x/take/keeper"
)

func setupHandlerKeeper(t *testing.T) (tkeeper.IKeeper, sdk.Context) {
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
	return tkeeper.NewKeeper(cdc, storeKey, authority), ctx
}

func TestUpdateParams_RejectsInvalidAuthority(t *testing.T) {
	keeper, ctx := setupHandlerKeeper(t)
	server := handler.NewMsgServerImpl(keeper)

	req := &types.MsgUpdateParams{
		Authority: "virtengine1invalidauthority",
		Params:    types.DefaultParams(),
	}

	_, err := server.UpdateParams(ctx, req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid authority")
}

func TestUpdateParams_SetsParams(t *testing.T) {
	keeper, ctx := setupHandlerKeeper(t)
	server := handler.NewMsgServerImpl(keeper)

	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	params := types.Params{
		DefaultTakeRate: 25,
		DenomTakeRates: types.DenomTakeRates{
			{Denom: "uve", Rate: 4},
			{Denom: "ufoo", Rate: 9},
		},
	}

	_, err := server.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: authority,
		Params:    params,
	})
	require.NoError(t, err)
	require.Equal(t, params, keeper.GetParams(ctx))
}

func TestUpdateParams_InvalidParams(t *testing.T) {
	keeper, ctx := setupHandlerKeeper(t)
	server := handler.NewMsgServerImpl(keeper)

	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	params := types.Params{
		DefaultTakeRate: 20,
		DenomTakeRates: types.DenomTakeRates{
			{Denom: "ufoo", Rate: 2},
		},
	}

	_, err := server.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: authority,
		Params:    params,
	})
	require.Error(t, err)
}
