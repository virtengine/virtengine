// Package keeper implements the staking module keeper.
//
// VE-921: Reward distribution logic
package keeper

import (
	"encoding/json"
	"fmt"
	"math/big"
	"sort"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	delegationtypes "github.com/virtengine/virtengine/x/delegation/types"
	"github.com/virtengine/virtengine/x/staking/types"
)

type rewardComponent uint8

const (
	rewardComponentBlock rewardComponent = iota
	rewardComponentVEID
	rewardComponentUptime
)

type rewardAllocation struct {
	performance  types.ValidatorPerformance
	stake        int64
	blockWeight  *big.Int
	veidWeight   *big.Int
	uptimeWeight *big.Int
}

type rewardRemainder struct {
	validatorIndex int
	validatorAddr  string
	component      rewardComponent
	remainder      *big.Int
}

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

// CalculateEpochRewards calculates rewards for all validators in an epoch.
func (k Keeper) CalculateEpochRewards(ctx sdk.Context, epoch uint64) ([]types.ValidatorReward, error) {
	params := k.GetParams(ctx)

	epochInfo, found := k.GetRewardEpoch(ctx, epoch)
	if !found {
		return nil, types.ErrInvalidEpoch.Wrapf("epoch %d not found", epoch)
	}

	if epochInfo.Finalized {
		return nil, types.ErrRewardsAlreadyDistributed.Wrapf("epoch %d already finalized", epoch)
	}

	// Calculate epoch reward pool
	blocksInEpoch := types.EpochDuration(&epochInfo)
	if blocksInEpoch == 0 {
		blocksInEpoch = safeInt64FromUint64Rewards(params.EpochLength)
	}
	epochRewardPool := params.BaseRewardPerBlock * blocksInEpoch
	if epochRewardPool <= 0 {
		return []types.ValidatorReward{}, nil
	}

	var performances []types.ValidatorPerformance
	k.WithValidatorPerformances(ctx, func(perf types.ValidatorPerformance) bool {
		if perf.EpochNumber == epoch {
			performances = append(performances, perf)
		}
		return false
	})
	if len(performances) == 0 {
		return []types.ValidatorReward{}, nil
	}

	stakesByValidator, totalStake := k.buildStakeDistribution(ctx, performances)

	sort.Slice(performances, func(i, j int) bool {
		return performances[i].ValidatorAddress < performances[j].ValidatorAddress
	})

	allocations := make([]rewardAllocation, 0, len(performances))
	for _, perf := range performances {
		validatorStake, ok := stakesByValidator[perf.ValidatorAddress]
		if !ok || validatorStake <= 0 {
			validatorStake = minVirtualStakeUnits
		}

		blockScore, veidScore, uptimeScore := rewardComponentScores(perf)
		allocations = append(allocations, rewardAllocation{
			performance:  perf,
			stake:        validatorStake,
			blockWeight:  scaledRewardWeight(validatorStake, blockScore, types.WeightBlockProposal),
			veidWeight:   scaledRewardWeight(validatorStake, veidScore, types.WeightVEIDVerification),
			uptimeWeight: scaledRewardWeight(validatorStake, uptimeScore, types.WeightUptime),
		})
	}

	if len(allocations) == 0 {
		return []types.ValidatorReward{}, nil
	}

	rewards := allocateEpochRewards(allocations, totalStake, epochRewardPool, params.RewardDenom)
	calculatedAt := ctx.BlockTime()
	for idx := range rewards {
		rewards[idx].EpochNumber = epoch
		rewards[idx].CalculatedAt = &calculatedAt
		rewards[idx].BlockHeight = ctx.BlockHeight()
	}

	k.Logger(ctx).Info("calculated epoch rewards",
		"epoch", epoch,
		"validators", len(rewards),
		"total_pool", epochRewardPool,
		"total_stake", totalStake,
	)

	return rewards, nil
}

// CalculateVEIDRewards calculates VEID verification rewards for an epoch
func (k Keeper) CalculateVEIDRewards(ctx sdk.Context, epoch uint64) ([]types.ValidatorReward, error) {
	params := k.GetParams(ctx)

	epochInfo, found := k.GetRewardEpoch(ctx, epoch)
	if !found {
		return nil, types.ErrInvalidEpoch.Wrapf("epoch %d not found", epoch)
	}

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

// DistributeRewards distributes rewards for an epoch.
func (k Keeper) DistributeRewards(ctx sdk.Context, epoch uint64) error {
	params := k.GetParams(ctx)

	stakingRewards, err := k.CalculateEpochRewards(ctx, epoch)
	if err != nil {
		return err
	}

	veidRewards, err := k.CalculateVEIDRewards(ctx, epoch)
	if err != nil {
		k.Logger(ctx).Error("failed to calculate VEID rewards", "error", err)
	}

	rewardsByValidator := make(map[string]*types.ValidatorReward, len(stakingRewards)+len(veidRewards))
	validatorAddresses := make([]string, 0, len(stakingRewards)+len(veidRewards))

	mergeReward := func(in types.ValidatorReward) {
		if existing, ok := rewardsByValidator[in.ValidatorAddress]; ok {
			existing.BlockProposalReward = existing.BlockProposalReward.Add(in.BlockProposalReward...)
			existing.VEIDReward = existing.VEIDReward.Add(in.VEIDReward...)
			existing.UptimeReward = existing.UptimeReward.Add(in.UptimeReward...)
			existing.IdentityNetworkReward = existing.IdentityNetworkReward.Add(in.IdentityNetworkReward...)
			if in.PerformanceScore > existing.PerformanceScore {
				existing.PerformanceScore = in.PerformanceScore
			}
			if in.StakeWeight != "" {
				existing.StakeWeight = in.StakeWeight
			}
			existing.TotalReward = types.ComputeTotalReward(existing)
			return
		}

		rewardCopy := in
		rewardCopy.TotalReward = types.ComputeTotalReward(&rewardCopy)
		rewardsByValidator[in.ValidatorAddress] = &rewardCopy
		validatorAddresses = append(validatorAddresses, in.ValidatorAddress)
	}

	for _, reward := range stakingRewards {
		mergeReward(reward)
	}
	for _, reward := range veidRewards {
		mergeReward(reward)
	}

	sort.Strings(validatorAddresses)

	epochInfo, found := k.GetRewardEpoch(ctx, epoch)
	if !found {
		epochInfo = *types.NewRewardEpoch(epoch, ctx.BlockHeight(), ctx.BlockTime())
	}

	var totalDistributed sdk.Coins
	var totalBlockRewards sdk.Coins
	var totalVEIDRewards sdk.Coins
	var totalUptimeRewards sdk.Coins

	for _, validatorAddrStr := range validatorAddresses {
		reward := rewardsByValidator[validatorAddrStr]
		if reward == nil {
			continue
		}

		if err := k.SetValidatorReward(ctx, *reward); err != nil {
			k.Logger(ctx).Error("failed to store validator reward",
				"validator", reward.ValidatorAddress,
				"error", err,
			)
			continue
		}

		validatorPayout := reward.TotalReward
		delegatorPayout := sdk.NewCoins()
		if k.stakingKeeper != nil && !reward.TotalReward.IsZero() {
			commissionCoins, delegatorCoins, distErr := k.stakingKeeper.DistributeValidatorRewardsToDelegators(ctx, reward.ValidatorAddress, epoch, reward.TotalReward)
			if distErr != nil {
				return distErr
			}
			validatorPayout = commissionCoins
			delegatorPayout = delegatorCoins
		}

		if k.bankKeeper != nil && !reward.TotalReward.IsZero() {
			if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, reward.TotalReward); err != nil {
				k.Logger(ctx).Error("failed to mint rewards",
					"validator", reward.ValidatorAddress,
					"amount", reward.TotalReward,
					"error", err,
				)
				continue
			}

			if !delegatorPayout.IsZero() {
				if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, delegationtypes.ModuleName, delegatorPayout); err != nil {
					k.Logger(ctx).Error("failed to route delegator rewards",
						"validator", reward.ValidatorAddress,
						"amount", delegatorPayout,
						"error", err,
					)
					continue
				}
			}

			if !validatorPayout.IsZero() {
				validatorAddr, err := sdk.AccAddressFromBech32(reward.ValidatorAddress)
				if err != nil {
					k.Logger(ctx).Error("invalid validator address", "address", reward.ValidatorAddress, "error", err)
					continue
				}

				if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, validatorAddr, validatorPayout); err != nil {
					k.Logger(ctx).Error("failed to distribute rewards",
						"validator", reward.ValidatorAddress,
						"amount", validatorPayout,
						"error", err,
					)
					continue
				}
			}
		}

		totalDistributed = totalDistributed.Add(reward.TotalReward...)
		totalBlockRewards = totalBlockRewards.Add(reward.BlockProposalReward...)
		totalVEIDRewards = totalVEIDRewards.Add(reward.VEIDReward...)
		totalUptimeRewards = totalUptimeRewards.Add(reward.UptimeReward...)

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

	epochInfo.TotalRewardsDistributed = totalDistributed
	epochInfo.BlockProposalRewards = totalBlockRewards
	epochInfo.VEIDRewards = totalVEIDRewards
	epochInfo.UptimeRewards = totalUptimeRewards
	epochInfo.ValidatorCount = int64(len(validatorAddresses))
	epochInfo.EndHeight = ctx.BlockHeight()
	endTime := ctx.BlockTime()
	epochInfo.EndTime = &endTime
	epochInfo.Finalized = true

	if err := k.SetRewardEpoch(ctx, epochInfo); err != nil {
		return err
	}

	k.Logger(ctx).Info("distributed epoch rewards",
		"epoch", epoch,
		"validators", len(validatorAddresses),
		"total_distributed", totalDistributed.String(),
		"denom", params.RewardDenom,
	)

	return nil
}

func safeInt64FromUint64Rewards(value uint64) int64 {
	if value > uint64(^uint64(0)>>1) {
		return int64(^uint64(0) >> 1)
	}
	//nolint:gosec // range checked above
	return int64(value)
}

// DistributeIdentityNetworkRewards distributes identity network rewards for an epoch
func (k Keeper) DistributeIdentityNetworkRewards(ctx sdk.Context, epoch uint64) error {
	params := k.GetParams(ctx)

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
			if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, reward); err != nil {
				k.Logger(ctx).Error("failed to mint identity network rewards", "error", err)
				continue
			}

			validatorAddr, err := sdk.AccAddressFromBech32(perf.ValidatorAddress)
			if err != nil {
				k.Logger(ctx).Error("invalid validator address", "address", perf.ValidatorAddress, "error", err)
				continue
			}
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, validatorAddr, reward); err != nil {
				k.Logger(ctx).Error("failed to distribute identity network rewards", "error", err)
				continue
			}
		}

		totalDistributed = totalDistributed.Add(reward...)

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

func rewardComponentScores(perf types.ValidatorPerformance) (int64, int64, int64) {
	blockScore := int64(0)
	if perf.BlocksExpected > 0 {
		blockScore = (perf.BlocksProposed * types.MaxPerformanceScore) / perf.BlocksExpected
		if blockScore > types.MaxPerformanceScore {
			blockScore = types.MaxPerformanceScore
		}
	} else if perf.BlocksProposed > 0 {
		blockScore = types.MaxPerformanceScore
	}

	veidScore := clampPerformanceScore(perf.VEIDVerificationScore)
	if perf.VEIDVerificationsExpected > 0 {
		completionRate := (perf.VEIDVerificationsCompleted * types.MaxPerformanceScore) / perf.VEIDVerificationsExpected
		if completionRate > types.MaxPerformanceScore {
			completionRate = types.MaxPerformanceScore
		}
		veidScore = (completionRate + veidScore) / 2
	}

	uptimeScore := types.MaxPerformanceScore
	totalTime := perf.UptimeSeconds + perf.DowntimeSeconds
	if totalTime > 0 {
		uptimeScore = (perf.UptimeSeconds * types.MaxPerformanceScore) / totalTime
	}

	return clampPerformanceScore(blockScore), clampPerformanceScore(veidScore), clampPerformanceScore(uptimeScore)
}

func clampPerformanceScore(score int64) int64 {
	if score < 0 {
		return 0
	}
	if score > types.MaxPerformanceScore {
		return types.MaxPerformanceScore
	}
	return score
}

func scaledRewardWeight(stake, score, componentWeight int64) *big.Int {
	if stake <= 0 || score <= 0 || componentWeight <= 0 {
		return big.NewInt(0)
	}

	weight := big.NewInt(stake)
	weight.Mul(weight, big.NewInt(score))
	weight.Mul(weight, big.NewInt(componentWeight))
	return weight
}

func allocateEpochRewards(allocations []rewardAllocation, totalStake, epochRewardPool int64, denom string) []types.ValidatorReward {
	rewards := make([]types.ValidatorReward, len(allocations))

	for idx, allocation := range allocations {
		reward := types.NewValidatorReward(allocation.performance.ValidatorAddress, 0)
		reward.PerformanceScore = allocation.performance.OverallScore
		reward.StakeWeight = formatStakeWeight(allocation.stake, totalStake)
		rewards[idx] = *reward
	}

	if totalStake <= 0 || epochRewardPool <= 0 {
		return rewards
	}

	denominator := big.NewInt(totalStake)
	denominator.Mul(denominator, big.NewInt(types.TotalWeight))
	denominator.Mul(denominator, big.NewInt(types.MaxPerformanceScore))
	if denominator.Sign() == 0 {
		return rewards
	}

	pool := big.NewInt(epochRewardPool)
	remainders := make([]rewardRemainder, 0, len(allocations)*3)
	allocated := big.NewInt(0)
	totalNumerator := big.NewInt(0)

	for idx, allocation := range allocations {
		setComponentAllocation(&rewards[idx], denom, pool, denominator, allocation.blockWeight, rewardComponentBlock, idx, &remainders, totalNumerator, allocated)
		setComponentAllocation(&rewards[idx], denom, pool, denominator, allocation.veidWeight, rewardComponentVEID, idx, &remainders, totalNumerator, allocated)
		setComponentAllocation(&rewards[idx], denom, pool, denominator, allocation.uptimeWeight, rewardComponentUptime, idx, &remainders, totalNumerator, allocated)
	}

	totalEarned := new(big.Int).Quo(totalNumerator, denominator)
	leftover := new(big.Int).Sub(totalEarned, allocated)
	if leftover.Sign() > 0 {
		sort.SliceStable(remainders, func(i, j int) bool {
			cmp := remainders[i].remainder.Cmp(remainders[j].remainder)
			if cmp != 0 {
				return cmp > 0
			}
			if remainders[i].validatorAddr != remainders[j].validatorAddr {
				return remainders[i].validatorAddr < remainders[j].validatorAddr
			}
			return remainders[i].component < remainders[j].component
		})

		one := big.NewInt(1)
		for idx := 0; idx < len(remainders) && leftover.Sign() > 0; idx++ {
			entry := remainders[idx]
			incrementRewardComponent(&rewards[entry.validatorIndex], denom, entry.component)
			leftover.Sub(leftover, one)
		}
	}

	for idx := range rewards {
		rewards[idx].TotalReward = types.ComputeTotalReward(&rewards[idx])
	}

	return rewards
}

func setComponentAllocation(
	reward *types.ValidatorReward,
	denom string,
	pool *big.Int,
	totalWeight *big.Int,
	weight *big.Int,
	component rewardComponent,
	validatorIndex int,
	remainders *[]rewardRemainder,
	totalNumerator *big.Int,
	allocated *big.Int,
) {
	if weight == nil || weight.Sign() == 0 {
		*remainders = append(*remainders, rewardRemainder{
			validatorIndex: validatorIndex,
			validatorAddr:  reward.ValidatorAddress,
			component:      component,
			remainder:      big.NewInt(0),
		})
		return
	}

	product := new(big.Int).Mul(pool, weight)
	totalNumerator.Add(totalNumerator, product)
	quotient := new(big.Int)
	remainder := new(big.Int)
	quotient.QuoRem(product, totalWeight, remainder)
	setRewardComponent(reward, denom, component, quotient)
	allocated.Add(allocated, quotient)
	*remainders = append(*remainders, rewardRemainder{
		validatorIndex: validatorIndex,
		validatorAddr:  reward.ValidatorAddress,
		component:      component,
		remainder:      remainder,
	})
}

func setRewardComponent(reward *types.ValidatorReward, denom string, component rewardComponent, amount *big.Int) {
	coins := coinsFromBigInt(denom, amount)
	switch component {
	case rewardComponentBlock:
		reward.BlockProposalReward = coins
	case rewardComponentVEID:
		reward.VEIDReward = coins
	case rewardComponentUptime:
		reward.UptimeReward = coins
	}
}

func incrementRewardComponent(reward *types.ValidatorReward, denom string, component rewardComponent) {
	switch component {
	case rewardComponentBlock:
		reward.BlockProposalReward = incrementCoinSet(reward.BlockProposalReward, denom)
	case rewardComponentVEID:
		reward.VEIDReward = incrementCoinSet(reward.VEIDReward, denom)
	case rewardComponentUptime:
		reward.UptimeReward = incrementCoinSet(reward.UptimeReward, denom)
	}
}

func coinsFromBigInt(denom string, amount *big.Int) sdk.Coins {
	if amount == nil || amount.Sign() <= 0 {
		return sdk.NewCoins()
	}
	return sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewIntFromBigInt(amount)))
}

func incrementCoinSet(coins sdk.Coins, denom string) sdk.Coins {
	amount := coins.AmountOf(denom)
	if amount.IsNil() {
		amount = sdkmath.ZeroInt()
	}
	amount = amount.AddRaw(1)
	return sdk.NewCoins(sdk.NewCoin(denom, amount))
}

func formatStakeWeight(stakeAmount, totalStake int64) string {
	if stakeAmount <= 0 || totalStake <= 0 {
		return "0"
	}

	stake := big.NewInt(stakeAmount)
	stake.Mul(stake, big.NewInt(types.FixedPointScale))
	stake.Div(stake, big.NewInt(totalStake))
	return stake.String()
}
