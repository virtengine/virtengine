package roles

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

	"github.com/virtengine/virtengine/x/roles/keeper"
	"github.com/virtengine/virtengine/x/roles/types"
)

var (
	_ module.AppModuleBasic   = AppModuleBasic{}
	_ module.HasGenesisBasics = AppModuleBasic{}

	_ appmodule.AppModule        = AppModule{}
	_ appmodule.HasEndBlocker    = AppModule{}
	_ module.HasConsensusVersion = AppModule{}
	_ module.HasGenesis          = AppModule{}
	_ module.HasServices         = AppModule{}

	_ module.AppModuleSimulation = AppModule{}
)

// AppModuleBasic defines the basic application module used by the roles module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the roles module's name
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the roles module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (b AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the roles module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	// Keep JSON genesis encoding aligned with the local module state structs.
	defaultGenesis := types.DefaultGenesisState()
	bz, err := json.Marshal(defaultGenesis)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal default genesis state: %v", err))
	}
	return bz
}

// ValidateGenesis performs genesis state validation for the roles module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	if bz == nil {
		return nil
	}
	// Keep JSON genesis decoding aligned with the local module state structs.
	var data types.GenesisState
	if err := json.Unmarshal(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %v", types.ModuleName, err)
	}

	return data.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the roles module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(fmt.Sprintf("couldn't register roles grpc routes: %s", err.Error()))
	}
}

// RegisterGRPCRoutes registers the gRPC Gateway routes for the roles module.
func (AppModuleBasic) RegisterGRPCRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	AppModuleBasic{}.RegisterGRPCGatewayRoutes(clientCtx, mux)
}

// GetQueryCmd returns the root query command of this module
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	panic("virtengine modules do not export cli commands via cosmos interface")
}

// GetTxCmd returns the transaction commands for this module
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	panic("virtengine modules do not export cli commands via cosmos interface")
}

// AppModule implements an application module for the roles module.
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

// Name returns the roles module name
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

// QuerierRoute returns the roles module's querier route name.
func (am AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

// RegisterServices registers the module's services
func (am AppModule) RegisterServices(cfg module.Configurator) {
	queryServer := keeper.GRPCQuerier{Keeper: am.keeper}
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), queryServer)

	// v1 -> v2: introduces the sanction store (prefixes 0x06-0x08).
	//
	// The migration is deliberately a no-op: sanctions did not exist before
	// v2, so there is no pre-existing state to transform and every account's
	// effective state already equals the projection of an empty sanction set
	// (active). Registering it is still required, because a missing migration
	// for a bumped ConsensusVersion makes the upgrade handler fail closed.
	if err := cfg.RegisterMigration(types.ModuleName, 1, func(_ sdk.Context) error {
		return nil
	}); err != nil {
		panic(err)
	}
}

// RegisterQueryService registers a GRPC query service to respond to the
// module-specific GRPC queries.
func (am AppModule) RegisterQueryService(server grpc.Server) {
	queryServer := keeper.GRPCQuerier{Keeper: am.keeper}
	types.RegisterQueryServer(server, queryServer)
}

// BeginBlock performs no-op
func (am AppModule) BeginBlock(_ context.Context) error {
	return nil
}

// EndBlock lapses due sanctions and re-projects the affected account states.
//
// This is the mechanism behind "an expired suspension restores the preceding
// state without manual intervention": a time-limited sanction clears itself on
// the first block after its expiry, with no transaction and no operator action.
func (am AppModule) EndBlock(ctx context.Context) error {
	return am.keeper.ProcessSanctionExpiry(sdk.UnwrapSDKContext(ctx))
}

// InitGenesis performs genesis initialization for the roles module.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	// Keep JSON genesis decoding aligned with the local module state structs.
	var genesisState types.GenesisState
	if err := json.Unmarshal(data, &genesisState); err != nil {
		panic(fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err))
	}
	InitGenesis(ctx, am.keeper, &genesisState)
}

// ExportGenesis returns the exported genesis state as raw bytes for the roles module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	// Keep JSON genesis encoding aligned with the local module state structs.
	bz, err := json.Marshal(gs)
	if err != nil {
		panic(fmt.Errorf("failed to marshal %s genesis state: %w", types.ModuleName, err))
	}
	return bz
}

// ConsensusVersion implements module.AppModule#ConsensusVersion
//
// Version 2 adds scoped, time-limited, appealable sanction records
// (store prefixes 0x06-0x08). The bump is required so the v1 -> v2 migration
// runs on upgrade; see RegisterServices.
func (am AppModule) ConsensusVersion() uint64 {
	return 2
}

// RegisterStoreDecoder registers a decoder for roles module's types.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {
	// No custom store decoder needed for roles module
}

// WeightedOperations doesn't return any roles module operation.
func (am AppModule) WeightedOperations(_ module.SimulationState) []simtypes.WeightedOperation {
	return []simtypes.WeightedOperation{}
}

// GenerateGenesisState creates a randomized GenState of the roles module.
func (AppModule) GenerateGenesisState(_ *module.SimulationState) {
	// Simulation genesis state generation not implemented
}
