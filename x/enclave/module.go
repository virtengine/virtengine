package enclave

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

	"github.com/virtengine/virtengine/x/enclave/keeper"
	"github.com/virtengine/virtengine/x/enclave/types"
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

// AppModuleBasic defines the basic application module used by the enclave module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the enclave module's name
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the enclave module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (b AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the enclave module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	// Use standard JSON encoding for stub types since they don't have proper proto marshaling
	defaultGenesis := types.DefaultGenesisState()
	bz, err := json.Marshal(defaultGenesis)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal default genesis state: %v", err))
	}
	return bz
}

// ValidateGenesis performs genesis state validation for the enclave module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	if bz == nil {
		return nil
	}
	// Use standard JSON decoding for stub types since they don't have proper proto marshaling
	var data types.GenesisState
	if err := json.Unmarshal(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %v", types.ModuleName, err)
	}

	return types.ValidateGenesis(&data)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the enclave module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// REST API routes will be automatically generated and registered once proper
	// protobuf definitions are generated (see ENCLAVE-ENH-001 task).
	// The gRPC-Gateway will expose all query endpoints as REST endpoints:
	// - GET /virtengine/enclave/v1/identity/{validator_address}
	// - GET /virtengine/enclave/v1/keys/active
	// - GET /virtengine/enclave/v1/measurements
	// - etc.
}

// RegisterGRPCRoutes registers the gRPC Gateway routes for the enclave module.
func (AppModuleBasic) RegisterGRPCRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// gRPC routes will be automatically registered once proper
	// protobuf definitions and service registration is complete.
}

// GetQueryCmd returns the root query command of this module
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	panic("virtengine modules do not export cli commands via cosmos interface")
}

// GetTxCmd returns the transaction commands for this module
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	panic("virtengine modules do not export cli commands via cosmos interface")
}

// AppModule implements an application module for the enclave module.
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
	}
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// ConsensusVersion implements module.HasConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return 1 }

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))
}

// InitGenesis performs genesis initialization for the enclave module.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	// Use standard JSON decoding for stub types since they don't have proper proto marshaling
	var genesisState types.GenesisState
	if err := json.Unmarshal(data, &genesisState); err != nil {
		panic(fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err))
	}
	InitGenesis(ctx, am.keeper, &genesisState)
}

// ExportGenesis returns the exported genesis state as raw bytes for the enclave module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	// Use standard JSON encoding for stub types since they don't have proper proto marshaling
	gs := ExportGenesis(ctx, am.keeper)
	bz, err := json.Marshal(gs)
	if err != nil {
		panic(fmt.Errorf("failed to marshal %s genesis state: %w", types.ModuleName, err))
	}
	return bz
}

// BeginBlock performs begin block logic for the enclave module
func (am AppModule) BeginBlock(goCtx context.Context) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	am.keeper.BeginBlocker(ctx)
	return nil
}

// EndBlock performs end block logic for the enclave module
func (am AppModule) EndBlock(goCtx context.Context) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	am.keeper.EndBlocker(ctx)
	return nil
}

// GenerateGenesisState implements module.AppModuleSimulation
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	// Simulation genesis state generation
}

// RegisterStoreDecoder implements module.AppModuleSimulation
func (am AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {
	// Register store decoder for simulation
}

// WeightedOperations implements module.AppModuleSimulation
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return nil
}

// RegisterMsgServer is a helper to register MsgServer
func RegisterMsgServer(s grpc.Server, srv types.MsgServer) {
	// This would be implemented with protobuf-generated code in production
}

// RegisterQueryServer is a helper to register QueryServer
func RegisterQueryServer(s grpc.Server, srv types.QueryServer) {
	// This would be implemented with protobuf-generated code in production
}
