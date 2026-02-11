package marketplace

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	marketplacev1 "github.com/virtengine/virtengine/sdk/go/node/marketplace/v1"
	marketplacetypes "github.com/virtengine/virtengine/x/market/types/marketplace"
	marketplacekeeper "github.com/virtengine/virtengine/x/market/types/marketplace/keeper"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.AppModule      = AppModule{}

	_ appmodule.AppModule = AppModule{}
)

// ModuleName is the name of the marketplace module.
const ModuleName = marketplacetypes.ModuleName

// AppModuleBasic defines the basic application module used by the marketplace module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the module's name.
func (AppModuleBasic) Name() string {
	return marketplacetypes.ModuleName
}

// RegisterLegacyAminoCodec registers amino types.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	marketplacetypes.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module interfaces.
func (b AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	marketplacetypes.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	defaultGenesis := marketplacetypes.DefaultGenesisState()
	bz, err := json.Marshal(defaultGenesis)
	if err != nil {
		panic(fmt.Errorf("failed to marshal %s genesis state: %w", marketplacetypes.ModuleName, err))
	}
	return bz
}

// ValidateGenesis performs genesis state validation.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	if bz == nil {
		return nil
	}
	var data marketplacetypes.GenesisState
	if err := json.Unmarshal(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", marketplacetypes.ModuleName, err)
	}
	return data.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// TODO: RegisterQueryHandlerClient not generated for marketplace v1 - needs proto regeneration
	// if err := marketplacev1.RegisterQueryHandlerClient(context.Background(), mux, marketplacev1.NewQueryClient(clientCtx)); err != nil {
	// 	panic(fmt.Errorf("couldn't register marketplace grpc routes: %s", err.Error()))
	// }
}

// GetTxCmd returns the root tx command for the module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// GetQueryCmd returns the root query command for the module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
}

// AppModule implements the marketplace module.
type AppModule struct {
	AppModuleBasic
	keeper marketplacekeeper.IKeeper
}

// NewAppModule creates a new AppModule.
func NewAppModule(cdc codec.Codec, k marketplacekeeper.IKeeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         k,
	}
}

// Name returns the module's name.
func (am AppModule) Name() string {
	return marketplacetypes.ModuleName
}

// RegisterInvariants registers module invariants.
//
//nolint:staticcheck // sdk.InvariantRegistry is deprecated in upstream SDK
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	marketplacetypes.RegisterMsgServer(cfg.MsgServer(), marketplacekeeper.NewMsgServerImpl(am.keeper))
	marketplacev1.RegisterQueryServer(cfg.QueryServer(), marketplacekeeper.NewQueryServerImpl(am.keeper))
}

// InitGenesis performs genesis initialization.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesisState marketplacetypes.GenesisState
	if err := json.Unmarshal(data, &genesisState); err != nil {
		panic(fmt.Errorf("failed to unmarshal %s genesis state: %w", marketplacetypes.ModuleName, err))
	}
	InitGenesis(ctx, am.keeper, &genesisState)
}

// ExportGenesis returns exported genesis state as raw bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	bz, err := json.Marshal(gs)
	if err != nil {
		panic(fmt.Errorf("failed to marshal %s genesis state: %w", marketplacetypes.ModuleName, err))
	}
	return bz
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// IsOnePerModuleType implements depinject.OnePerModuleType.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements appmodule.AppModule.
func (am AppModule) IsAppModule() {}

// BeginBlock is a no-op for the marketplace module.
func (am AppModule) BeginBlock(ctx context.Context) error {
	_ = ctx
	return nil
}

// EndBlock is a no-op for the marketplace module.
func (am AppModule) EndBlock(ctx context.Context) error {
	_ = ctx
	return nil
}

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the marketplace module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	bz, err := json.Marshal(marketplacetypes.DefaultGenesisState())
	if err != nil {
		panic(fmt.Errorf("failed to marshal %s genesis state: %w", marketplacetypes.ModuleName, err))
	}
	simState.GenState[marketplacetypes.ModuleName] = bz
}

// RegisterStoreDecoder registers a decoder for marketplace module's types.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the marketplace module operations with their respective weights.
func (am AppModule) WeightedOperations(_ module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
