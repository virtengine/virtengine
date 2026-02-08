package keeper_test

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
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/resources/keeper"
	"github.com/virtengine/virtengine/x/resources/types"
)

const testAuthority = "virtengine1authorityxxxxxxxxxxxxxxxxxxxxxxxxx"

func setupKeeper(t *testing.T) (keeper.Keeper, sdk.Context) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)
	legacyAmino := codec.NewLegacyAmino()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	paramsKey := storetypes.NewKVStoreKey(paramtypes.StoreKey)
	paramsTKey := storetypes.NewTransientStoreKey(paramtypes.TStoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(paramsKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(paramsTKey, storetypes.StoreTypeTransient, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	ctx := sdk.NewContext(stateStore, cmtproto.Header{Height: 1, Time: time.Now().UTC()}, false, log.NewNopLogger())
	subspace := paramtypes.NewSubspace(cdc, legacyAmino, paramsKey, paramsTKey, types.ModuleName).WithKeyTable(types.ParamKeyTable())

	k := keeper.NewKeeper(cdc, storeKey, subspace, testAuthority)
	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))

	return k, ctx
}
