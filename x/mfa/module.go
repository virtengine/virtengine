package mfa

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/gogoproto/grpc"

	"github.com/virtengine/virtengine/x/mfa/keeper"
	"github.com/virtengine/virtengine/x/mfa/types"
)

var (
	_ module.AppModuleBasic   = AppModuleBasic{}
	_ module.HasGenesisBasics = AppModuleBasic{}

	_ appmodule.AppModule        = AppModule{}
	_ module.HasConsensusVersion = AppModule{}
	_ module.HasGenesis          = AppModule{}
	_ module.HasServices         = AppModule{}

	_ module.AppModuleSimulation = AppModule{}
)

// AppModuleBasic defines the basic application module used by the mfa module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the mfa module's name
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the mfa module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (b AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the mfa module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the mfa module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	if bz == nil {
		return nil
	}

	var data types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %v", types.ModuleName, err)
	}

	return data.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the mfa module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// Register gRPC gateway routes when query client is available
	// This would be implemented with protobuf-generated code in a full implementation
}

// RegisterGRPCRoutes registers the gRPC Gateway routes for the mfa module.
func (AppModuleBasic) RegisterGRPCRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// Register gRPC routes when query client is available
}

// GetQueryCmd returns the root query command of this module
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	panic("virtengine modules do not export cli commands via cosmos interface")
}

// GetTxCmd returns the transaction commands for this module
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	panic("virtengine modules do not export cli commands via cosmos interface")
}

// AppModule implements an application module for the mfa module.
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, k keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         k,
	}
}

// Name returns the mfa module name
func (AppModule) Name() string {
	return types.ModuleName
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {
	// This function is intentionally empty - it's a marker interface implementation
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {
	// This function is intentionally empty - it's a marker interface implementation
}

// QuerierRoute returns the mfa module's querier route name.
func (am AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

// RegisterServices registers the module's services
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerWithContext(am.keeper))
	// TODO: GRPCQuerier doesn't match QueryServer interface - wrong signature for GetAllSensitiveTxConfigs
	// types.RegisterQueryServer(cfg.QueryServer(), keeper.GRPCQuerier{Keeper: am.keeper})
}

// RegisterQueryService registers a GRPC query service to respond to the
// module-specific GRPC queries.
func (am AppModule) RegisterQueryService(server grpc.Server) {
	// TODO: GRPCQuerier doesn't match QueryServer interface - wrong signature for GetAllSensitiveTxConfigs
	// types.RegisterQueryServer(server, keeper.GRPCQuerier{Keeper: am.keeper})
}

// BeginBlock performs no-op
func (am AppModule) BeginBlock(_ context.Context) error {
	return nil
}

// EndBlock returns the end blocker for the mfa module.
func (am AppModule) EndBlock(ctx context.Context) error {
	// Clean up expired challenges and sessions
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	hooks := keeper.NewMFAGatingHooks(am.keeper)
	hooks.CleanupExpiredData(sdkCtx)
	return nil
}

// InitGenesis performs genesis initialization for the mfa module.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	am.keeper.InitGenesis(ctx, &genesisState)
}

// ExportGenesis returns the exported genesis state as raw bytes for the mfa module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion returns the mfa module's consensus version.
func (am AppModule) ConsensusVersion() uint64 {
	return 1
}

// GenerateGenesisState creates a randomized GenesisState of the mfa module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	// Generate default genesis for simulation
}

// RegisterStoreDecoder registers a decoder for the mfa module's types.
func (am AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {
	// Register store decoder for simulation
}

// WeightedOperations returns the all the mfa module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
