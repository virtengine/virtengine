// Package keeper implements the staking module keeper.
//
// VE-921: Block hooks for staking module
package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"pkg.akt.dev/node/x/staking/types"
)

// BeginBlocker is called at the beginning of every block
func (k Keeper) BeginBlocker(ctx sdk.Context) error {
	// Track block proposer performance
	proposer := ctx.BlockHeader().ProposerAddress
	if proposer != nil {
		proposerAddr := sdk.AccAddress(proposer).String()
		update := PerformanceUpdate{
			BlockProposed: true,
			BlockSigned:   true,
		}
		if err := k.UpdateValidatorPerformance(ctx, proposerAddr, update); err != nil {
			k.Logger(ctx).Debug("failed to update proposer performance", "error", err)
		}
	}

	return nil
}

// EndBlocker is called at the end of every block
func (k Keeper) EndBlocker(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	currentEpoch := k.GetCurrentEpoch(ctx)

	// Check if we need to create a new epoch
	epochInfo, found := k.GetRewardEpoch(ctx, currentEpoch)
	if !found {
		// Create new epoch
		epochInfo = *types.NewRewardEpoch(currentEpoch, ctx.BlockHeight(), ctx.BlockTime())
		if err := k.SetRewardEpoch(ctx, epochInfo); err != nil {
			k.Logger(ctx).Error("failed to create epoch", "error", err)
		}
	}

	// Check if current epoch should end
	blocksInEpoch := ctx.BlockHeight() - epochInfo.StartHeight
	if blocksInEpoch >= int64(params.EpochLength) && !epochInfo.Finalized {
		// Finalize current epoch and distribute rewards
		if err := k.DistributeRewards(ctx, currentEpoch); err != nil {
			k.Logger(ctx).Error("failed to distribute rewards", "error", err)
		}

		// Distribute identity network rewards
		if err := k.DistributeIdentityNetworkRewards(ctx, currentEpoch); err != nil {
			k.Logger(ctx).Error("failed to distribute identity network rewards", "error", err)
		}

		// Start new epoch
		newEpoch := currentEpoch + 1
		k.SetCurrentEpoch(ctx, newEpoch)

		newEpochInfo := types.NewRewardEpoch(newEpoch, ctx.BlockHeight(), ctx.BlockTime())
		if err := k.SetRewardEpoch(ctx, *newEpochInfo); err != nil {
			k.Logger(ctx).Error("failed to create new epoch", "error", err)
		}

		k.Logger(ctx).Info("epoch transition",
			"old_epoch", currentEpoch,
			"new_epoch", newEpoch,
			"blocks_in_epoch", blocksInEpoch,
		)
	}

	return nil
}
