// Package fraud implements the Fraud module for VirtEngine.
//
// VE-912: Fraud reporting flow
// This module provides fraud reporting from providers to moderators,
// featuring encrypted evidence storage, moderator queue routing,
// and comprehensive on-chain audit trail.
package fraud

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/virtengine/virtengine/x/fraud/client/cli"
	"github.com/virtengine/virtengine/x/fraud/keeper"
	"github.com/virtengine/virtengine/x/fraud/types"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.AppModule      = AppModule{}

	_ appmodule.AppModule       = AppModule{}
	_ appmodule.HasBeginBlocker = AppModule{}
	_ appmodule.HasEndBlocker   = AppModule{}
)

// AppModuleBasic defines the basic application module used by the Fraud module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the Fraud module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the Fraud module's types on the given LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (b AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the Fraud module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	// Use proto-generated types with proper codec marshaling
	defaultGenesis := types.DefaultGenesisStatePB()
	return cdc.MustMarshalJSON(defaultGenesis)
}

// ValidateGenesis performs genesis state validation for the Fraud module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	if bz == nil {
		return nil
	}
	// Use proto-generated types with proper codec unmarshaling
	var data types.GenesisStatePB
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	// Convert to local type for validation
	localData := types.GenesisStateFromProto(&data)
	return localData.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the Fraud module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// gRPC gateway routes will be registered here when proto definitions are added
}

// GetTxCmd returns the root tx command for the Fraud module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd returns the root query command for the Fraud module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// AppModule implements an application module for the Fraud module.
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

// Name returns the Fraud module's name.
func (am AppModule) Name() string {
	return types.ModuleName
}

// RegisterServices registers module services
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// Register MsgServer for transaction handling
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))

	// Register query server
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(am.keeper))
}

// RegisterInvariants registers module invariants
//
//nolint:staticcheck // sdk.InvariantRegistry is deprecated in upstream SDK
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	keeper.RegisterInvariants(ir, am.keeper)
}

// InitGenesis performs genesis initialization
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	// Use proto-generated types with proper codec unmarshaling
	var genesisStatePB types.GenesisStatePB
	cdc.MustUnmarshalJSON(data, &genesisStatePB)
	// Convert to local type for initialization
	genesisState := types.GenesisStateFromProto(&genesisStatePB)
	InitGenesis(ctx, am.keeper, genesisState)
}

// ExportGenesis returns the exported genesis state
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	// Export to local type then convert to proto for marshaling
	gs := ExportGenesis(ctx, am.keeper)
	gsPB := types.GenesisStateToProto(gs)
	return cdc.MustMarshalJSON(gsPB)
}

// ConsensusVersion returns the consensus version
func (am AppModule) ConsensusVersion() uint64 {
	return 1
}

// BeginBlock executes all ABCI BeginBlock logic
func (am AppModule) BeginBlock(ctx context.Context) error {
	// BeginBlock logic here
	return nil
}

// EndBlock executes all ABCI EndBlock logic
func (am AppModule) EndBlock(ctx context.Context) error {
	// EndBlock logic here - could include auto-escalation of stale reports
	return nil
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}
