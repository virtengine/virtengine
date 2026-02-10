// Package keeper implements the delegation module keeper.
//
// VE-922: Historical rewards queries and range claims
package keeper

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/delegation/types"
)

// ClaimRewardsInRange claims rewards for a delegator from a validator in a block height range.
func (k Keeper) ClaimRewardsInRange(ctx sdk.Context, delegatorAddr, validatorAddr string, startHeight, endHeight int64) (sdk.Coins, error) {
	if startHeight < 0 || endHeight < 0 {
		return nil, types.ErrInvalidAmount.Wrap("start_height and end_height must be non-negative")
	}
	if endHeight != 0 && startHeight > endHeight {
		return nil, types.ErrInvalidAmount.Wrap("start_height cannot be greater than end_height")
	}

	// If no range provided, default to all unclaimed rewards.
	if startHeight == 0 && endHeight == 0 {
		return k.ClaimRewards(ctx, delegatorAddr, validatorAddr)
	}

	if endHeight == 0 {
		endHeight = int64(^uint64(0) >> 1)
	}

	unclaimedRewards := k.GetDelegatorValidatorRewardsInRange(ctx, delegatorAddr, validatorAddr, startHeight, endHeight, true)
	if len(unclaimedRewards) == 0 {
		return sdk.NewCoins(), nil
	}

	totalReward := big.NewInt(0)
	now := ctx.BlockTime()

	for _, reward := range unclaimedRewards {
		rewardAmount := reward.GetRewardBigInt()
		totalReward.Add(totalReward, rewardAmount)

		reward.Claimed = true
		reward.ClaimedAt = &now

		if err := k.SetDelegatorReward(ctx, reward); err != nil {
			return nil, fmt.Errorf("failed to update reward: %w", err)
		}
	}

	if totalReward.Sign() <= 0 {
		return sdk.NewCoins(), nil
	}

	delegatorAccAddr, err := sdk.AccAddressFromBech32(delegatorAddr)
	if err != nil {
		return nil, types.ErrInvalidDelegator.Wrapf("invalid delegator address: %v", err)
	}

	rewardCoins := sdk.NewCoins(sdk.NewCoin(DefaultRewardDenom, math.NewIntFromBigInt(totalReward)))
	if k.bankKeeper != nil {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, delegatorAccAddr, rewardCoins); err != nil {
			return nil, fmt.Errorf("failed to transfer rewards: %w", err)
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaimReward,
			sdk.NewAttribute(types.AttributeKeyDelegator, delegatorAddr),
			sdk.NewAttribute(types.AttributeKeyValidator, validatorAddr),
			sdk.NewAttribute(types.AttributeKeyReward, totalReward.String()),
		),
	)

	return rewardCoins, nil
}

// GetDelegatorValidatorRewardsInRange returns rewards for a delegator/validator in a height range.
func (k Keeper) GetDelegatorValidatorRewardsInRange(
	ctx sdk.Context,
	delegatorAddr, validatorAddr string,
	startHeight, endHeight int64,
	unclaimedOnly bool,
) []types.DelegatorReward {
	var rewards []types.DelegatorReward

	k.WithDelegatorRewards(ctx, func(reward types.DelegatorReward) bool {
		if reward.DelegatorAddress != delegatorAddr || reward.ValidatorAddress != validatorAddr {
			return false
		}
		if unclaimedOnly && reward.Claimed {
			return false
		}
		if reward.Height < startHeight || reward.Height > endHeight {
			return false
		}
		rewards = append(rewards, reward)
		return false
	})

	return rewards
}
