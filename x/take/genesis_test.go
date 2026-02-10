package take

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
	"github.com/virtengine/virtengine/x/take/keeper"
)

func setupGenesisKeeper(t *testing.T) (keeper.IKeeper, sdk.Context) {
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
	return keeper.NewKeeper(cdc, storeKey, authority), ctx
}

func TestDefaultGenesisState_Validate(t *testing.T) {
	gs := DefaultGenesisState()
	require.Equal(t, types.DefaultParams(), gs.Params)
	require.NoError(t, ValidateGenesis(gs))
}

func TestValidateGenesis_CustomParams(t *testing.T) {
	gs := &types.GenesisState{
		Params: types.Params{
			DefaultTakeRate: 15,
			DenomTakeRates: types.DenomTakeRates{
				{Denom: "uve", Rate: 3},
				{Denom: "ufoo", Rate: 6},
			},
		},
	}
	require.NoError(t, ValidateGenesis(gs))
}

func TestValidateGenesis_InvalidParams(t *testing.T) {
	gs := &types.GenesisState{
		Params: types.Params{
			DefaultTakeRate: 200,
			DenomTakeRates: types.DenomTakeRates{
				{Denom: "uve", Rate: 2},
			},
		},
	}
	require.Error(t, ValidateGenesis(gs))
}

func TestGenesis_ExportImportRoundTrip(t *testing.T) {
	keeper, ctx := setupGenesisKeeper(t)

	gs := &types.GenesisState{
		Params: types.Params{
			DefaultTakeRate: 12,
			DenomTakeRates: types.DenomTakeRates{
				{Denom: "uve", Rate: 2},
				{Denom: "ufoo", Rate: 7},
			},
		},
	}

	InitGenesis(ctx, keeper, gs)

	exported := ExportGenesis(ctx, keeper)
	require.Equal(t, gs, exported)
}
