package settlement

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

	"github.com/virtengine/virtengine/x/settlement/keeper"
	"github.com/virtengine/virtengine/x/settlement/types"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.AppModule      = AppModule{}

	_ appmodule.AppModule       = AppModule{}
	_ appmodule.HasBeginBlocker = AppModule{}
	_ appmodule.HasEndBlocker   = AppModule{}
)

// ModuleName is the name of the settlement module
const ModuleName = types.ModuleName

// AppModuleBasic defines the basic application module used by the settlement module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the settlement module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the settlement module's types on the given LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (b AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the settlement module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	defaultGenesis := types.DefaultGenesisState()
	bz, err := json.Marshal(defaultGenesis)
	if err != nil {
		panic(fmt.Errorf("failed to marshal %s genesis state: %w", types.ModuleName, err))
	}
	return bz
}

// ValidateGenesis performs genesis state validation for the settlement module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	if bz == nil {
		return nil
	}
	var data types.GenesisState
	if err := json.Unmarshal(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return data.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the settlement module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// gRPC gateway routes will be registered here when proto definitions are added
}

// GetTxCmd returns the root tx command for the settlement module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil // CLI commands to be implemented
}

// GetQueryCmd returns the root query command for the settlement module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil // CLI commands to be implemented
}

// AppModule implements an application module for the settlement module.
type AppModule struct {
	AppModuleBasic
	keeper keeper.IKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, k keeper.IKeeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         k,
	}
}

// Name returns the settlement module's name.
func (am AppModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants registers the settlement module invariants.
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {
	// Settlement module does not register invariants as state consistency
	// is enforced through escrow lifecycle validations
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// Message server and query server registration would go here
}

// InitGenesis performs genesis initialization for the settlement module.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesisState types.GenesisState
	if err := json.Unmarshal(data, &genesisState); err != nil {
		panic(fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err))
	}
	InitGenesis(ctx, am.keeper, &genesisState)
}

// ExportGenesis returns the exported genesis state as raw bytes for the settlement module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	bz, err := json.Marshal(gs)
	if err != nil {
		panic(fmt.Errorf("failed to marshal %s genesis state: %w", types.ModuleName, err))
	}
	return bz
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {
	// Required marker method for depinject module registration
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {
	// Required marker method for appmodule interface compliance
}

// BeginBlock returns the begin blocker for the settlement module.
func (am AppModule) BeginBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return BeginBlocker(sdkCtx, am.keeper)
}

// EndBlock returns the end blocker for the settlement module.
func (am AppModule) EndBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return EndBlocker(sdkCtx, am.keeper)
}

// BeginBlocker performs begin block operations for the settlement module
func BeginBlocker(ctx sdk.Context, k keeper.IKeeper) error {
	// Process expired escrows at the beginning of each block
	if err := k.ProcessExpiredEscrows(ctx); err != nil {
		return err
	}

	// Satisfy timelock conditions for escrows
	if err := k.SatisfyTimelockConditions(ctx); err != nil {
		return err
	}

	return nil
}

// EndBlocker performs end block operations for the settlement module
func EndBlocker(ctx sdk.Context, k keeper.IKeeper) error {
	// Auto-settle orders at the end of each block
	if err := k.AutoSettle(ctx); err != nil {
		return err
	}

	// Process reward distributions at the end of each block
	if err := k.EndBlockerRewards(ctx); err != nil {
		return err
	}

	return nil
}
