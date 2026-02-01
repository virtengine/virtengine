// Package keeper implements the staking module keeper.
//
// VE-921: Reward distribution logic
package keeper

import (
	"encoding/json"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/staking/types"
)

// ============================================================================
// Reward Epoch Management
// ============================================================================

// GetRewardEpoch returns a reward epoch
func (k Keeper) GetRewardEpoch(ctx sdk.Context, epochNumber uint64) (types.RewardEpoch, bool) {
	store := ctx.KVStore(k.skey)
	key := types.GetRewardEpochKey(epochNumber)
	bz := store.Get(key)
	if bz == nil {
		return types.RewardEpoch{}, false
	}

	var epoch types.RewardEpoch
	if err := json.Unmarshal(bz, &epoch); err != nil {
		return types.RewardEpoch{}, false
	}
	return epoch, true
}

// SetRewardEpoch stores a reward epoch
func (k Keeper) SetRewardEpoch(ctx sdk.Context, epoch types.RewardEpoch) error {
	if err := types.ValidateRewardEpoch(&epoch); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	key := types.GetRewardEpochKey(epoch.EpochNumber)
	bz, err := json.Marshal(epoch)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// WithRewardEpochs iterates over all reward epochs
func (k Keeper) WithRewardEpochs(ctx sdk.Context, fn func(types.RewardEpoch) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.RewardEpochPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var epoch types.RewardEpoch
		if err := json.Unmarshal(iter.Value(), &epoch); err != nil {
			continue
		}
		if fn(epoch) {
			break
		}
	}
}

// ============================================================================
// Validator Rewards
// ============================================================================

// GetValidatorReward returns a validator's reward for an epoch
func (k Keeper) GetValidatorReward(ctx sdk.Context, validatorAddr string, epoch uint64) (types.ValidatorReward, bool) {
	store := ctx.KVStore(k.skey)
	key := types.GetValidatorRewardKey(validatorAddr, epoch)
	bz := store.Get(key)
	if bz == nil {
		return types.ValidatorReward{}, false
	}

	var reward types.ValidatorReward
	if err := json.Unmarshal(bz, &reward); err != nil {
		return types.ValidatorReward{}, false
	}
	return reward, true
}

// SetValidatorReward stores a validator's reward
func (k Keeper) SetValidatorReward(ctx sdk.Context, reward types.ValidatorReward) error {
	if err := types.ValidateValidatorReward(&reward); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	key := types.GetValidatorRewardKey(reward.ValidatorAddress, reward.EpochNumber)
	bz, err := json.Marshal(reward)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// WithValidatorRewards iterates over all validator rewards
func (k Keeper) WithValidatorRewards(ctx sdk.Context, fn func(types.ValidatorReward) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.ValidatorRewardPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var reward types.ValidatorReward
		if err := json.Unmarshal(iter.Value(), &reward); err != nil {
			continue
		}
		if fn(reward) {
			break
		}
	}
}

// ============================================================================
// Reward Calculation
// ============================================================================

// CalculateEpochRewards calculates rewards for all validators in an epoch
func (k Keeper) CalculateEpochRewards(ctx sdk.Context, epoch uint64) ([]types.ValidatorReward, error) {
	params := k.GetParams(ctx)

	// Get epoch info
	epochInfo, found := k.GetRewardEpoch(ctx, epoch)
	if !found {
		return nil, types.ErrInvalidEpoch.Wrapf("epoch %d not found", epoch)
	}

	if epochInfo.Finalized {
		return nil, types.ErrRewardsAlreadyDistributed.Wrapf("epoch %d already finalized", epoch)
	}

	// Get total stake (or use placeholder if staking keeper not available)
	var totalStake int64 = 1000000000 // Default: 1000 tokens total stake
	if k.stakingKeeper != nil {
		totalStake = k.stakingKeeper.GetTotalStake(ctx)
	}

	// Calculate epoch reward pool
	blocksInEpoch := types.EpochDuration(&epochInfo)
	if blocksInEpoch == 0 {
		blocksInEpoch = int64(params.EpochLength)
	}
	epochRewardPool := params.BaseRewardPerBlock * blocksInEpoch

	// Get all performances for this epoch
	var performances []types.ValidatorPerformance
	k.WithValidatorPerformances(ctx, func(perf types.ValidatorPerformance) bool {
		if perf.EpochNumber == epoch {
			performances = append(performances, perf)
		}
		return false
	})

	// Calculate rewards for each validator
	rewards := make([]types.ValidatorReward, 0, len(performances))
	for _, perf := range performances {
		// Get validator stake (or use equal distribution if staking keeper not available)
		var validatorStake int64 = totalStake / int64(len(performances))
		if k.stakingKeeper != nil {
			validatorAddr, _ := sdk.AccAddressFromBech32(perf.ValidatorAddress)
			validatorStake = k.stakingKeeper.GetValidatorStake(ctx, validatorAddr)
		}

		// Calculate reward using deterministic integer arithmetic
		input := types.RewardCalculationInput{
			ValidatorAddress: perf.ValidatorAddress,
			Performance:      &perf,
			StakeAmount:      validatorStake,
			TotalStake:       totalStake,
			EpochRewardPool:  epochRewardPool,
			BlocksInEpoch:    blocksInEpoch,
		}

		reward := types.CalculateRewards(input, params.RewardDenom)
		reward.EpochNumber = epoch
		calculatedAt := ctx.BlockTime()
		reward.CalculatedAt = &calculatedAt
		reward.BlockHeight = ctx.BlockHeight()

		rewards = append(rewards, *reward)
	}

	k.Logger(ctx).Info("calculated epoch rewards",
		"epoch", epoch,
		"validators", len(rewards),
		"total_pool", epochRewardPool,
	)

	return rewards, nil
}

// CalculateVEIDRewards calculates VEID verification rewards for an epoch
func (k Keeper) CalculateVEIDRewards(ctx sdk.Context, epoch uint64) ([]types.ValidatorReward, error) {
	params := k.GetParams(ctx)

	// Get epoch info
	epochInfo, found := k.GetRewardEpoch(ctx, epoch)
	if !found {
		return nil, types.ErrInvalidEpoch.Wrapf("epoch %d not found", epoch)
	}

	// Get all performances for this epoch
	var totalVerifications int64
	var performances []types.ValidatorPerformance

	k.WithValidatorPerformances(ctx, func(perf types.ValidatorPerformance) bool {
		if perf.EpochNumber == epoch {
			performances = append(performances, perf)
			totalVerifications += perf.VEIDVerificationsCompleted
		}
		return false
	})

	if totalVerifications == 0 {
		k.Logger(ctx).Debug("no VEID verifications in epoch", "epoch", epoch)
		return nil, nil
	}

	// Calculate VEID rewards
	rewards := make([]types.ValidatorReward, 0, len(performances))
	for _, perf := range performances {
		if perf.VEIDVerificationsCompleted == 0 {
			continue
		}

		input := types.IdentityNetworkRewardInput{
			ValidatorAddress:         perf.ValidatorAddress,
			VerificationsCompleted:   perf.VEIDVerificationsCompleted,
			TotalVerifications:       totalVerifications,
			AverageVerificationScore: perf.VEIDVerificationScore,
			RewardPool:               params.VEIDRewardPool,
		}

		veidReward := types.CalculateIdentityNetworkReward(input, params.RewardDenom)

		reward := types.NewValidatorReward(perf.ValidatorAddress, epoch)
		reward.VEIDReward = veidReward
		reward.IdentityNetworkReward = veidReward
		calculatedAt := ctx.BlockTime()
		reward.CalculatedAt = &calculatedAt
		reward.BlockHeight = ctx.BlockHeight()
		reward.TotalReward = types.ComputeTotalReward(reward)

		rewards = append(rewards, *reward)
	}

	k.Logger(ctx).Info("calculated VEID rewards",
		"epoch", epoch,
		"validators", len(rewards),
		"total_verifications", totalVerifications,
		"epoch_duration", types.EpochDuration(&epochInfo),
	)

	return rewards, nil
}

// DistributeRewards distributes rewards for an epoch
func (k Keeper) DistributeRewards(ctx sdk.Context, epoch uint64) error {
	params := k.GetParams(ctx)

	// Calculate staking rewards
	stakingRewards, err := k.CalculateEpochRewards(ctx, epoch)
	if err != nil {
		return err
	}

	// Calculate VEID rewards
	veidRewards, err := k.CalculateVEIDRewards(ctx, epoch)
	if err != nil {
		k.Logger(ctx).Error("failed to calculate VEID rewards", "error", err)
		// Continue with staking rewards only
	}

	// Merge rewards by validator
	rewardsByValidator := make(map[string]*types.ValidatorReward)
	for _, r := range stakingRewards {
		reward := r // Create a copy
		rewardsByValidator[r.ValidatorAddress] = &reward
	}

	for _, r := range veidRewards {
		if existing, ok := rewardsByValidator[r.ValidatorAddress]; ok {
			existing.VEIDReward = existing.VEIDReward.Add(r.VEIDReward...)
			existing.IdentityNetworkReward = existing.IdentityNetworkReward.Add(r.IdentityNetworkReward...)
			existing.TotalReward = types.ComputeTotalReward(existing)
		} else {
			reward := r // Create a copy
			rewardsByValidator[r.ValidatorAddress] = &reward
		}
	}

	// Get or create epoch info
	epochInfo, found := k.GetRewardEpoch(ctx, epoch)
	if !found {
		epochInfo = *types.NewRewardEpoch(epoch, ctx.BlockHeight(), ctx.BlockTime())
	}

	// Distribute rewards
	var totalDistributed sdk.Coins
	for _, reward := range rewardsByValidator {
		// Store reward record
		if err := k.SetValidatorReward(ctx, *reward); err != nil {
			k.Logger(ctx).Error("failed to store validator reward",
				"validator", reward.ValidatorAddress,
				"error", err,
			)
			continue
		}

		// Mint and distribute rewards (if bank keeper available)
		if k.bankKeeper != nil && !reward.TotalReward.IsZero() {
			// Mint rewards
			if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, reward.TotalReward); err != nil {
				k.Logger(ctx).Error("failed to mint rewards",
					"validator", reward.ValidatorAddress,
					"amount", reward.TotalReward,
					"error", err,
				)
				continue
			}

			// Send to validator
			validatorAddr, err := sdk.AccAddressFromBech32(reward.ValidatorAddress)
			if err != nil {
				k.Logger(ctx).Error("invalid validator address", "address", reward.ValidatorAddress)
				continue
			}

			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, validatorAddr, reward.TotalReward); err != nil {
				k.Logger(ctx).Error("failed to distribute rewards",
					"validator", reward.ValidatorAddress,
					"amount", reward.TotalReward,
					"error", err,
				)
				continue
			}
		}

		totalDistributed = totalDistributed.Add(reward.TotalReward...)

		// Emit event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRewardsDistributed,
				sdk.NewAttribute(types.AttributeKeyValidatorAddress, reward.ValidatorAddress),
				sdk.NewAttribute(types.AttributeKeyTotalRewards, reward.TotalReward.String()),
				sdk.NewAttribute(types.AttributeKeyPerformanceScore, fmt.Sprintf("%d", reward.PerformanceScore)),
				sdk.NewAttribute(types.AttributeKeyEpochNumber, fmt.Sprintf("%d", epoch)),
			),
		)
	}

	// Update epoch info
	epochInfo.TotalRewardsDistributed = totalDistributed
	epochInfo.ValidatorCount = int64(len(rewardsByValidator))
	epochInfo.EndHeight = ctx.BlockHeight()
	endTime := ctx.BlockTime()
	epochInfo.EndTime = &endTime
	epochInfo.Finalized = true

	if err := k.SetRewardEpoch(ctx, epochInfo); err != nil {
		return err
	}

	k.Logger(ctx).Info("distributed epoch rewards",
		"epoch", epoch,
		"validators", len(rewardsByValidator),
		"total_distributed", totalDistributed.String(),
		"denom", params.RewardDenom,
	)

	return nil
}

// DistributeIdentityNetworkRewards distributes identity network rewards for an epoch
func (k Keeper) DistributeIdentityNetworkRewards(ctx sdk.Context, epoch uint64) error {
	params := k.GetParams(ctx)

	// Get performances for identity network work
	var performances []types.ValidatorPerformance
	var totalVerifications int64

	k.WithValidatorPerformances(ctx, func(perf types.ValidatorPerformance) bool {
		if perf.EpochNumber == epoch && perf.VEIDVerificationsCompleted > 0 {
			performances = append(performances, perf)
			totalVerifications += perf.VEIDVerificationsCompleted
		}
		return false
	})

	if len(performances) == 0 {
		k.Logger(ctx).Debug("no identity network work to reward", "epoch", epoch)
		return nil
	}

	// Calculate and distribute identity network rewards
	var totalDistributed sdk.Coins
	for _, perf := range performances {
		input := types.IdentityNetworkRewardInput{
			ValidatorAddress:         perf.ValidatorAddress,
			VerificationsCompleted:   perf.VEIDVerificationsCompleted,
			TotalVerifications:       totalVerifications,
			AverageVerificationScore: perf.VEIDVerificationScore,
			RewardPool:               params.IdentityNetworkRewardPool,
		}

		reward := types.CalculateIdentityNetworkReward(input, params.RewardDenom)

		if k.bankKeeper != nil && !reward.IsZero() {
			// Mint rewards
			if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, reward); err != nil {
				k.Logger(ctx).Error("failed to mint identity network rewards", "error", err)
				continue
			}

			// Distribute
			validatorAddr, _ := sdk.AccAddressFromBech32(perf.ValidatorAddress)
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, validatorAddr, reward); err != nil {
				k.Logger(ctx).Error("failed to distribute identity network rewards", "error", err)
				continue
			}
		}

		totalDistributed = totalDistributed.Add(reward...)

		// Emit event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeVEIDRewardsDistributed,
				sdk.NewAttribute(types.AttributeKeyValidatorAddress, perf.ValidatorAddress),
				sdk.NewAttribute(types.AttributeKeyTotalRewards, reward.String()),
				sdk.NewAttribute(types.AttributeKeyVEIDScore, fmt.Sprintf("%d", perf.VEIDVerificationScore)),
			),
		)
	}

	k.Logger(ctx).Info("distributed identity network rewards",
		"epoch", epoch,
		"validators", len(performances),
		"total_distributed", totalDistributed.String(),
	)

	return nil
}
