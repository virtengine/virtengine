// Package staking implements the staking module for VirtEngine.
//
// VE-921: Genesis initialization
package staking

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"pkg.akt.dev/node/x/staking/keeper"
	"pkg.akt.dev/node/x/staking/types"
)

// InitGenesis initializes the staking module's state from a genesis state
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data *types.GenesisState) {
	// Set parameters
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	// Set current epoch
	k.SetCurrentEpoch(ctx, data.CurrentEpoch)

	// Set slash sequence
	k.SetNextSlashSequence(ctx, data.SlashSequence)

	// Initialize validator performances
	for _, perf := range data.ValidatorPerformances {
		if err := k.SetValidatorPerformance(ctx, perf); err != nil {
			panic(err)
		}
	}

	// Initialize slash records
	for _, record := range data.SlashRecords {
		if err := k.SetSlashRecord(ctx, record); err != nil {
			panic(err)
		}
	}

	// Initialize reward epochs
	for _, epoch := range data.RewardEpochs {
		if err := k.SetRewardEpoch(ctx, epoch); err != nil {
			panic(err)
		}
	}

	// Initialize validator rewards
	for _, reward := range data.ValidatorRewards {
		if err := k.SetValidatorReward(ctx, reward); err != nil {
			panic(err)
		}
	}

	// Initialize validator signing infos
	for _, info := range data.ValidatorSigningInfos {
		if err := k.SetValidatorSigningInfo(ctx, info); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis exports the staking module's state to a genesis state
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	gs := &types.GenesisState{
		Params:       k.GetParams(ctx),
		CurrentEpoch: k.GetCurrentEpoch(ctx),
	}

	// Export validator performances
	k.WithValidatorPerformances(ctx, func(perf types.ValidatorPerformance) bool {
		gs.ValidatorPerformances = append(gs.ValidatorPerformances, perf)
		return false
	})

	// Export slash records
	k.WithSlashRecords(ctx, func(record types.SlashRecord) bool {
		gs.SlashRecords = append(gs.SlashRecords, record)
		return false
	})

	// Export reward epochs
	k.WithRewardEpochs(ctx, func(epoch types.RewardEpoch) bool {
		gs.RewardEpochs = append(gs.RewardEpochs, epoch)
		return false
	})

	// Export validator rewards
	k.WithValidatorRewards(ctx, func(reward types.ValidatorReward) bool {
		gs.ValidatorRewards = append(gs.ValidatorRewards, reward)
		return false
	})

	return gs
}
