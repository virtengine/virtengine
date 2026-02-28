package marketplace

import (
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/stretchr/testify/require"
	grpc "google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/module"

	marketplacekeeper "github.com/virtengine/virtengine/x/market/types/marketplace/keeper"
)

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
