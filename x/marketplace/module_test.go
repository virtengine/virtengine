package marketplace

import (
	"bytes"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/stretchr/testify/require"
	grpc "google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	marketplacetypes "github.com/virtengine/virtengine/x/market/types/marketplace"
	marketplacekeeper "github.com/virtengine/virtengine/x/market/types/marketplace/keeper"
)

func TestInitGenesisActivatesCanonicalLifecycle(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	c := codec.NewProtoCodec(registry)
	key := storetypes.NewKVStoreKey(marketplacetypes.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())
	ctx := sdk.NewContext(stateStore, cmtproto.Header{Height: 1, Time: time.Unix(1, 0).UTC()}, false, log.NewNopLogger())
	keeper := marketplacekeeper.NewKeeper(c, key, sdk.AccAddress(bytes.Repeat([]byte{9}, 20)).String(), nil, nil, nil)

	InitGenesis(ctx, keeper, marketplacetypes.DefaultGenesisState())
	require.True(t, keeper.IsCanonicalLifecycleActive(ctx))
}

func TestRegisterServicesRegistersMarketplaceQueryServer(t *testing.T) {
	server := grpc.NewServer()
	cfg := module.NewConfigurator(nil, server, server)

	am := NewAppModule(nil, marketplacekeeper.Keeper{})
	am.RegisterServices(cfg)

	services := server.GetServiceInfo()
	queryService, ok := services["virtengine.marketplace.v1.Query"]
	require.True(t, ok, "marketplace query service must be registered")
	require.Len(t, queryService.Methods, 3)
}

func TestRegisterGRPCGatewayRoutesDoesNotPanic(t *testing.T) {
	basic := AppModuleBasic{}
	require.NotPanics(t, func() {
		basic.RegisterGRPCGatewayRoutes(client.Context{}, runtime.NewServeMux())
	})
}
