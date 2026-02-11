// Package keeper implements the delegation module keeper.
//
// VE-922: Reward distribution for delegators
package keeper

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/delegation/types"
)

// DefaultRewardDenom is the fallback denomination for rewards
const DefaultRewardDenom = "uve"

// DistributeValidatorRewardsToDelegators distributes a validator's rewards to their delegators.
// Returns the validator commission and total delegator rewards.
// This should be called after validator rewards are calculated in the staking module.
func (k Keeper) DistributeValidatorRewardsToDelegators(ctx sdk.Context, validatorAddr string, epoch uint64, validatorReward sdk.Coins) (sdk.Coins, sdk.Coins, error) {
	if validatorReward.IsZero() {
		return sdk.NewCoins(), sdk.NewCoins(), nil
	}

	params := k.GetParams(ctx)
	rewardDenom := params.RewardDenom
	if rewardDenom == "" {
		rewardDenom = DefaultRewardDenom
	}

	if validatorReward.Len() > 1 {
		return nil, nil, types.ErrInvalidAmount.Wrap("validator reward must use a single denom")
	}
	if validatorReward.Len() == 1 && validatorReward[0].Denom != rewardDenom {
		return nil, nil, types.ErrInvalidAmount.Wrapf("unexpected reward denom: %s", validatorReward[0].Denom)
	}

	rewardAmount := validatorReward.AmountOf(rewardDenom)
	if !rewardAmount.IsPositive() {
		return sdk.NewCoins(), sdk.NewCoins(), nil
	}

	rewardBig := rewardAmount.BigInt()

	// Get validator shares to determine proportions
	valShares, found := k.GetValidatorShares(ctx, validatorAddr)
	if !found || valShares.GetTotalSharesBigInt().Sign() == 0 {
		commissionCoins := sdk.NewCoins(sdk.NewCoin(rewardDenom, math.NewIntFromBigInt(rewardBig)))
		return commissionCoins, sdk.NewCoins(), nil // No delegations
	}

	// Calculate commission (goes to validator) using params commission rate
	// commission = validatorReward * commissionRate / BasisPointsMax
	commission := new(big.Int).Mul(rewardBig, big.NewInt(params.ValidatorCommissionRate))
	commission.Div(commission, big.NewInt(BasisPointsMax))

	// Distributable = validatorReward - commission
	distributable := new(big.Int).Sub(rewardBig, commission)
	if distributable.Sign() <= 0 {
		commissionCoins := sdk.NewCoins(sdk.NewCoin(rewardDenom, math.NewIntFromBigInt(rewardBig)))
		return commissionCoins, sdk.NewCoins(), nil
	}

	totalShares := valShares.GetTotalSharesBigInt()

	// Get all delegations for this validator
	delegations := k.GetValidatorDelegations(ctx, validatorAddr)
	totalDelegatorReward := big.NewInt(0)

	for _, del := range delegations {
		delegatorShares := del.GetSharesBigInt()
		if delegatorShares.Sign() <= 0 {
			continue
		}

		// delegatorReward = distributable * delegatorShares / totalShares
		delegatorReward := new(big.Int).Mul(distributable, delegatorShares)
		delegatorReward.Div(delegatorReward, totalShares)

		if delegatorReward.Sign() <= 0 {
			continue
		}

		totalDelegatorReward.Add(totalDelegatorReward, delegatorReward)

		// Create delegator reward record
		reward := types.NewDelegatorReward(
			del.DelegatorAddress,
			validatorAddr,
			epoch,
			delegatorReward.String(),
			del.Shares,
			valShares.TotalShares,
			ctx.BlockTime(),
			ctx.BlockHeight(),
		)

		if err := k.SetDelegatorReward(ctx, *reward); err != nil {
			return nil, nil, fmt.Errorf("failed to set delegator reward: %w", err)
		}

		// Emit reward event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeDistributeReward,
				sdk.NewAttribute(types.AttributeKeyDelegator, del.DelegatorAddress),
				sdk.NewAttribute(types.AttributeKeyValidator, validatorAddr),
				sdk.NewAttribute(types.AttributeKeyEpoch, fmt.Sprintf("%d", epoch)),
				sdk.NewAttribute(types.AttributeKeyReward, delegatorReward.String()),
			),
		)
	}

	// Assign any remainder to commission to preserve totals.
	remainder := new(big.Int).Sub(distributable, totalDelegatorReward)
	if remainder.Sign() > 0 {
		commission.Add(commission, remainder)
	}

	commissionCoins := sdk.NewCoins(sdk.NewCoin(rewardDenom, math.NewIntFromBigInt(commission)))
	delegatorCoins := sdk.NewCoins()
	if totalDelegatorReward.Sign() > 0 {
		delegatorCoins = sdk.NewCoins(sdk.NewCoin(rewardDenom, math.NewIntFromBigInt(totalDelegatorReward)))
	}

	k.Logger(ctx).Info("distributed validator rewards to delegators",
		"validator", validatorAddr,
		"epoch", epoch,
		"total_reward", rewardBig.String(),
		"commission", commission.String(),
		"delegator_total", totalDelegatorReward.String(),
		"delegation_count", len(delegations),
	)

	return commissionCoins, delegatorCoins, nil
}

// ClaimRewards claims rewards for a delegator from a specific validator
func (k Keeper) ClaimRewards(ctx sdk.Context, delegatorAddr, validatorAddr string) (sdk.Coins, error) {
	// Get all unclaimed rewards from this validator
	unclaimedRewards := k.GetDelegatorValidatorUnclaimedRewards(ctx, delegatorAddr, validatorAddr)
	if len(unclaimedRewards) == 0 {
		return sdk.NewCoins(), nil
	}

	totalReward := big.NewInt(0)
	now := ctx.BlockTime()

	for _, reward := range unclaimedRewards {
		rewardAmount := reward.GetRewardBigInt()
		totalReward.Add(totalReward, rewardAmount)

		// Mark reward as claimed
		reward.Claimed = true
		reward.ClaimedAt = &now

		if err := k.SetDelegatorReward(ctx, reward); err != nil {
			return nil, fmt.Errorf("failed to update reward: %w", err)
		}
	}

	if totalReward.Sign() <= 0 {
		return sdk.NewCoins(), nil
	}

	params := k.GetParams(ctx)

	// Transfer rewards to delegator
	delegatorAccAddr, err := sdk.AccAddressFromBech32(delegatorAddr)
	if err != nil {
		return nil, types.ErrInvalidDelegator.Wrapf("invalid delegator address: %v", err)
	}

	rewardDenom := params.RewardDenom
	if rewardDenom == "" {
		rewardDenom = DefaultRewardDenom
	}
	rewardCoins := sdk.NewCoins(sdk.NewCoin(rewardDenom, math.NewIntFromBigInt(totalReward)))

	if k.bankKeeper != nil {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, delegatorAccAddr, rewardCoins); err != nil {
			return nil, fmt.Errorf("failed to transfer rewards: %w", err)
		}
	}

	// Emit claim reward event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaimReward,
			sdk.NewAttribute(types.AttributeKeyDelegator, delegatorAddr),
			sdk.NewAttribute(types.AttributeKeyValidator, validatorAddr),
			sdk.NewAttribute(types.AttributeKeyReward, totalReward.String()),
		),
	)

	k.Logger(ctx).Info("rewards claimed",
		"delegator", delegatorAddr,
		"validator", validatorAddr,
		"amount", totalReward.String(),
		"reward_count", len(unclaimedRewards),
	)

	return rewardCoins, nil
}

// ClaimAllRewards claims all rewards for a delegator from all validators
func (k Keeper) ClaimAllRewards(ctx sdk.Context, delegatorAddr string) (sdk.Coins, error) {
	// Get all unclaimed rewards
	unclaimedRewards := k.GetDelegatorUnclaimedRewards(ctx, delegatorAddr)
	if len(unclaimedRewards) == 0 {
		return sdk.NewCoins(), nil
	}

	totalReward := big.NewInt(0)
	now := ctx.BlockTime()

	for _, reward := range unclaimedRewards {
		rewardAmount := reward.GetRewardBigInt()
		totalReward.Add(totalReward, rewardAmount)

		// Mark reward as claimed
		reward.Claimed = true
		reward.ClaimedAt = &now

		if err := k.SetDelegatorReward(ctx, reward); err != nil {
			return nil, fmt.Errorf("failed to update reward: %w", err)
		}
	}

	if totalReward.Sign() <= 0 {
		return sdk.NewCoins(), nil
	}

	params := k.GetParams(ctx)

	// Transfer rewards to delegator
	delegatorAccAddr, err := sdk.AccAddressFromBech32(delegatorAddr)
	if err != nil {
		return nil, types.ErrInvalidDelegator.Wrapf("invalid delegator address: %v", err)
	}

	rewardDenom := params.RewardDenom
	if rewardDenom == "" {
		rewardDenom = DefaultRewardDenom
	}
	rewardCoins := sdk.NewCoins(sdk.NewCoin(rewardDenom, math.NewIntFromBigInt(totalReward)))

	if k.bankKeeper != nil {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, delegatorAccAddr, rewardCoins); err != nil {
			return nil, fmt.Errorf("failed to transfer rewards: %w", err)
		}
	}

	// Emit claim reward event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaimReward,
			sdk.NewAttribute(types.AttributeKeyDelegator, delegatorAddr),
			sdk.NewAttribute(types.AttributeKeyReward, totalReward.String()),
		),
	)

	k.Logger(ctx).Info("all rewards claimed",
		"delegator", delegatorAddr,
		"amount", totalReward.String(),
		"reward_count", len(unclaimedRewards),
	)

	return rewardCoins, nil
}

// GetDelegatorTotalRewards returns the total unclaimed rewards for a delegator
func (k Keeper) GetDelegatorTotalRewards(ctx sdk.Context, delegatorAddr string) string {
	unclaimedRewards := k.GetDelegatorUnclaimedRewards(ctx, delegatorAddr)

	totalReward := big.NewInt(0)
	for _, reward := range unclaimedRewards {
		rewardAmount := reward.GetRewardBigInt()
		totalReward.Add(totalReward, rewardAmount)
	}

	return totalReward.String()
}

// GetDelegatorValidatorTotalRewards returns the total unclaimed rewards for a delegator from a specific validator
func (k Keeper) GetDelegatorValidatorTotalRewards(ctx sdk.Context, delegatorAddr, validatorAddr string) string {
	unclaimedRewards := k.GetDelegatorValidatorUnclaimedRewards(ctx, delegatorAddr, validatorAddr)

	totalReward := big.NewInt(0)
	for _, reward := range unclaimedRewards {
		rewardAmount := reward.GetRewardBigInt()
		totalReward.Add(totalReward, rewardAmount)
	}

	return totalReward.String()
}
