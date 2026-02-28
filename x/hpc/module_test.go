package hpc

import (
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/stretchr/testify/require"
	grpc "google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/virtengine/virtengine/x/hpc/client/cli"
	"github.com/virtengine/virtengine/x/hpc/keeper"
)

func TestRegisterServicesRegistersHPCQueryServers(t *testing.T) {
	server := grpc.NewServer()
	cfg := module.NewConfigurator(nil, server, server)

	am := NewAppModule(nil, keeper.Keeper{})
	am.RegisterServices(cfg)

	services := server.GetServiceInfo()
	_, ok := services["virtengine.hpc.v1.Query"]
	require.True(t, ok, "hpc query service must be registered")

	templateService, ok := services["virtengine.hpc.v1.WorkloadTemplateQuery"]
	require.True(t, ok, "workload template query service must be registered")
	require.Len(t, templateService.Methods, 7)
}

func TestRegisterGRPCGatewayRoutesDoesNotPanic(t *testing.T) {
	basic := AppModuleBasic{}
	require.NotPanics(t, func() {
		basic.RegisterGRPCGatewayRoutes(client.Context{}, runtime.NewServeMux())
	})
}

func TestGetQueryCmdIncludesTemplateQueries(t *testing.T) {
	cmd := cli.GetQueryCmd()
	templateCmd, _, err := cmd.Find([]string{"template"})
	require.NoError(t, err)
	require.NotNil(t, templateCmd)
	require.Equal(t, "template", templateCmd.Name())
}
