// Package staking implements the staking module for VirtEngine.
//
// VE-921: Staking rewards module
package staking

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

	"github.com/virtengine/virtengine/x/staking/keeper"
	"github.com/virtengine/virtengine/x/staking/types"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.AppModule      = AppModule{}

	_ appmodule.AppModule       = AppModule{}
	_ appmodule.HasBeginBlocker = AppModule{}
	_ appmodule.HasEndBlocker   = AppModule{}
)

// AppModuleBasic defines the basic application module used by the staking module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the staking module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the staking module's types on the given LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (b AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the staking module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	// Use standard JSON encoding for stub types since they don't have proper proto marshaling
	defaultGenesis := types.DefaultGenesisState()
	bz, err := json.Marshal(defaultGenesis)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal default genesis state: %v", err))
	}
	return bz
}

// ValidateGenesis performs genesis state validation for the staking module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	if bz == nil {
		return nil
	}
	// Use standard JSON decoding for stub types since they don't have proper proto marshaling
	var data types.GenesisState
	if err := json.Unmarshal(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return types.ValidateGenesis(&data)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the staking module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// gRPC gateway routes will be registered here when proto definitions are added
}

// GetTxCmd returns the root tx command for the staking module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil // CLI commands to be implemented
}

// GetQueryCmd returns the root query command for the staking module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil // CLI commands to be implemented
}

// AppModule implements an application module for the staking module.
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

// Name returns the staking module's name.
func (am AppModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants registers the staking module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	RegisterInvariants(ir, am.keeper)
}

// RegisterInvariants registers all staking module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k keeper.Keeper) {
	ir.RegisterRoute(types.ModuleName, "validator-performance-consistency",
		ValidatorPerformanceConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "reward-non-negative",
		RewardNonNegativeInvariant(k))
}

// ValidatorPerformanceConsistencyInvariant checks that validator performance records
// have consistent data (uptime percentage between 0 and FixedPointScale, valid epoch numbers)
func ValidatorPerformanceConsistencyInvariant(k keeper.Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var invalidRecords []string

		k.WithValidatorPerformances(ctx, func(perf types.ValidatorPerformance) bool {
			// Check uptime is valid (0-100% in fixed-point scale, i.e., 0 to FixedPointScale)
			if uptimePercent := types.GetUptimePercent(&perf); uptimePercent < 0 || uptimePercent > types.FixedPointScale {
				invalidRecords = append(invalidRecords,
					fmt.Sprintf("validator %s has invalid uptime: %d (expected 0-%d)", perf.ValidatorAddress, uptimePercent, types.FixedPointScale))
			}

			// Check epoch number is positive
			if perf.EpochNumber == 0 {
				invalidRecords = append(invalidRecords,
					fmt.Sprintf("validator %s has zero epoch number", perf.ValidatorAddress))
			}

			// Check overall score is computed correctly (0 to MaxPerformanceScore, i.e., 10000)
			computedScore := types.ComputeOverallScore(&perf)
			if computedScore < 0 || computedScore > types.MaxPerformanceScore {
				invalidRecords = append(invalidRecords,
					fmt.Sprintf("validator %s has invalid overall score: %d (expected 0-%d)", perf.ValidatorAddress, computedScore, types.MaxPerformanceScore))
			}

			return false // continue iteration
		})

		if len(invalidRecords) > 0 {
			return sdk.FormatInvariant(types.ModuleName, "validator-performance-consistency",
				fmt.Sprintf("found %d invalid validator performance records:\n%s",
					len(invalidRecords), formatRecords(invalidRecords))), true
		}

		return sdk.FormatInvariant(types.ModuleName, "validator-performance-consistency",
			"all validator performance records are consistent"), false
	}
}

// RewardNonNegativeInvariant checks that all validator rewards are non-negative
func RewardNonNegativeInvariant(k keeper.Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var invalidRewards []string

		k.WithValidatorRewards(ctx, func(reward types.ValidatorReward) bool {
			// Check total reward is non-negative
			if reward.TotalReward.IsAnyNegative() {
				invalidRewards = append(invalidRewards,
					fmt.Sprintf("validator %s epoch %d has negative reward: %s",
						reward.ValidatorAddress, reward.EpochNumber, reward.TotalReward.String()))
			}

			return false // continue iteration
		})

		if len(invalidRewards) > 0 {
			return sdk.FormatInvariant(types.ModuleName, "reward-non-negative",
				fmt.Sprintf("found %d negative rewards:\n%s",
					len(invalidRewards), formatRecords(invalidRewards))), true
		}

		return sdk.FormatInvariant(types.ModuleName, "reward-non-negative",
			"all validator rewards are non-negative"), false
	}
}

// formatRecords formats a slice of record strings for invariant output
func formatRecords(records []string) string {
	result := ""
	for i, r := range records {
		if i > 0 {
			result += "\n"
		}
		result += "  - " + r
	}
	return result
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// Message server and query server registration would go here
}

// InitGenesis performs genesis initialization for the staking module.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	// Use standard JSON decoding for stub types since they don't have proper proto marshaling
	var genesisState types.GenesisState
	if err := json.Unmarshal(data, &genesisState); err != nil {
		panic(fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err))
	}
	InitGenesis(ctx, am.keeper, &genesisState)
}

// ExportGenesis returns the exported genesis state as raw bytes for the staking module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	// Use standard JSON encoding for stub types since they don't have proper proto marshaling
	gs := ExportGenesis(ctx, am.keeper)
	bz, err := json.Marshal(gs)
	if err != nil {
		panic(fmt.Errorf("failed to marshal %s genesis state: %w", types.ModuleName, err))
	}
	return bz
}

// ConsensusVersion returns the consensus state-breaking version for the module.
func (AppModule) ConsensusVersion() uint64 {
	return 1
}

// BeginBlock is called at the beginning of every block
func (am AppModule) BeginBlock(ctx context.Context) error {
	return am.keeper.BeginBlocker(sdk.UnwrapSDKContext(ctx))
}

// EndBlock is called at the end of every block
func (am AppModule) EndBlock(ctx context.Context) error {
	return am.keeper.EndBlocker(sdk.UnwrapSDKContext(ctx))
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
// This is a marker interface that does not require implementation.
func (am AppModule) IsOnePerModuleType() {
	// Marker interface method - no implementation needed
}

// IsAppModule implements the appmodule.AppModule interface.
// This is a marker interface that does not require implementation.
func (am AppModule) IsAppModule() {
	// Marker interface method - no implementation needed
}
