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
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the Fraud module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var data types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return data.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the Fraud module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// gRPC gateway routes will be registered here when proto definitions are added
}

// GetTxCmd returns the root tx command for the Fraud module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil // CLI commands to be implemented
}

// GetQueryCmd returns the root query command for the Fraud module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil // CLI commands to be implemented
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
	// Register query server
	queryServer := keeper.NewQueryServer(am.keeper)
	_ = queryServer // Will be registered when proto is generated
}

// RegisterInvariants registers module invariants
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	// Invariants to be implemented
}

// InitGenesis performs genesis initialization
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	InitGenesis(ctx, am.keeper, &genesisState)
}

// ExportGenesis returns the exported genesis state
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(gs)
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
