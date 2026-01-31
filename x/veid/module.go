package veid

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

	"github.com/virtengine/virtengine/x/veid/keeper"
	"github.com/virtengine/virtengine/x/veid/types"
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

// AppModuleBasic defines the basic application module used by the veid module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the veid module's name
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the veid module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (b AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the veid module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	// Use standard JSON encoding for stub types since they don't have proper proto marshaling
	defaultGenesis := types.DefaultGenesisState()
	bz, err := json.Marshal(defaultGenesis)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal default genesis state: %v", err))
	}
	return bz
}

// ValidateGenesis performs genesis state validation for the veid module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	if bz == nil {
		return nil
	}
	// Use standard JSON decoding for stub types since they don't have proper proto unmarshaling
	var data types.GenesisState
	if err := json.Unmarshal(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %v", types.ModuleName, err)
	}

	return data.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the veid module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// Register gRPC gateway routes when query client is available
	// This would be implemented with protobuf-generated code in a full implementation
}

// RegisterGRPCRoutes registers the gRPC Gateway routes for the veid module.
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

// AppModule implements an application module for the veid module.
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

// Name returns the veid module name
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

// QuerierRoute returns the veid module's querier route name.
func (am AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

// RegisterServices registers the module's services
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewGRPCQuerier(am.keeper))
}

// RegisterQueryService registers a GRPC query service to respond to the
// module-specific GRPC queries.
func (am AppModule) RegisterQueryService(server grpc.Server) {
	types.RegisterQueryServer(server, keeper.NewGRPCQuerier(am.keeper))
}

// BeginBlock performs no-op
func (am AppModule) BeginBlock(_ context.Context) error {
	return nil
}

// EndBlock returns the end blocker for the veid module.
func (am AppModule) EndBlock(_ context.Context) error {
	return nil
}

// InitGenesis performs genesis initialization for the veid module.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	// Use standard JSON decoding for stub types since they don't have proper proto unmarshaling
	var genesisState types.GenesisState
	if err := json.Unmarshal(data, &genesisState); err != nil {
		panic(fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err))
	}
	InitGenesis(ctx, am.keeper, &genesisState)
}

// ExportGenesis returns the exported genesis state as raw bytes for the veid module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	// Use standard JSON encoding for stub types since they don't have proper proto marshaling
	bz, err := json.Marshal(gs)
	if err != nil {
		panic(fmt.Errorf("failed to marshal %s genesis state: %w", types.ModuleName, err))
	}
	return bz
}

// ConsensusVersion returns the consensus version for the veid module.
func (AppModule) ConsensusVersion() uint64 {
	return 1
}

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the veid module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	// Simulation genesis state generation would go here
}

// RegisterStoreDecoder registers a decoder for veid module's types
func (am AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {
	// Store decoder registration would go here
}

// WeightedOperations returns the all the veid module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
