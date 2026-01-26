package keeper

import (
	"strconv"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"pkg.akt.dev/node/x/settlement/types"
)

// DistributeStakingRewards distributes staking rewards for an epoch
func (k Keeper) DistributeStakingRewards(ctx sdk.Context, epoch uint64) (*types.RewardDistribution, error) {
	k.Logger(ctx).Info("distributing staking rewards", "epoch", epoch)

	// Check if rewards were already distributed for this epoch
	existingDists := k.GetRewardsByEpoch(ctx, epoch)
	for _, dist := range existingDists {
		if dist.Source == types.RewardSourceStaking {
			return nil, types.ErrInvalidEpoch.Wrapf("staking rewards already distributed for epoch %d", epoch)
		}
	}

	// In a real implementation, you would:
	// 1. Query the staking module for validator/delegator stakes
	// 2. Calculate rewards based on stake weight
	// 3. Pull rewards from a reward pool

	// For now, we create a placeholder distribution
	// This would be integrated with the distribution module in production

	params := k.GetParams(ctx)
	_ = params

	// Placeholder: Get accumulated platform fees to distribute
	// In production, you would track platform fees in a separate account

	// Generate distribution ID
	seq := k.incrementDistributionSequence(ctx)
	distributionID := generateIDWithTimestamp("staking", seq, ctx.BlockTime().Unix())

	// Create empty distribution if no rewards to distribute
	dist := types.NewRewardDistribution(
		distributionID,
		epoch,
		types.RewardSourceStaking,
		[]types.RewardRecipient{},
		ctx.BlockTime(),
		ctx.BlockHeight(),
	)

	dist.Metadata = map[string]string{
		"epoch":  strconv.FormatUint(epoch, 10),
		"source": "staking",
	}

	// Save distribution
	if err := k.SetRewardDistribution(ctx, *dist); err != nil {
		return nil, err
	}

	// Emit event
	err := ctx.EventManager().EmitTypedEvent(&types.EventRewardsDistributed{
		DistributionID: distributionID,
		EpochNumber:    epoch,
		Source:         string(types.RewardSourceStaking),
		TotalRewards:   dist.TotalRewards.String(),
		RecipientCount: uint32(len(dist.Recipients)),
		DistributedAt:  ctx.BlockTime().Unix(),
	})
	if err != nil {
		k.Logger(ctx).Error("failed to emit staking rewards distributed event", "error", err)
	}

	k.Logger(ctx).Info("staking rewards distributed",
		"distribution_id", distributionID,
		"epoch", epoch,
		"recipients", len(dist.Recipients),
	)

	return dist, nil
}

// DistributeProviderRewards distributes rewards to providers based on usage
func (k Keeper) DistributeProviderRewards(ctx sdk.Context, usageRecords []types.UsageRecord) (*types.RewardDistribution, error) {
	if len(usageRecords) == 0 {
		return nil, types.ErrInvalidReward.Wrap("no usage records provided")
	}

	k.Logger(ctx).Info("distributing provider rewards", "usage_records", len(usageRecords))

	// Aggregate usage by provider
	providerUsage := make(map[string]uint64)
	providerOrderIDs := make(map[string][]string)

	for _, usage := range usageRecords {
		providerUsage[usage.Provider] += usage.UsageUnits
		providerOrderIDs[usage.Provider] = append(providerOrderIDs[usage.Provider], usage.OrderID)
	}

	// Calculate rewards based on usage
	// In production, you would have a reward rate per usage unit
	params := k.GetParams(ctx)
	_ = params

	recipients := make([]types.RewardRecipient, 0, len(providerUsage))

	for provider, units := range providerUsage {
		// Calculate reward (placeholder: 1 token per 100 usage units)
		rewardAmount := sdkmath.NewInt(int64(units / 100))
		if rewardAmount.IsZero() {
			continue
		}

		recipients = append(recipients, types.RewardRecipient{
			Address:     provider,
			Amount:      sdk.NewCoins(sdk.NewCoin("uve", rewardAmount)),
			Reason:      "provider usage reward",
			UsageUnits:  units,
			ReferenceID: providerOrderIDs[provider][0], // First order as reference
		})
	}

	if len(recipients) == 0 {
		return nil, types.ErrInvalidReward.Wrap("no rewards to distribute")
	}

	// Generate distribution ID
	seq := k.incrementDistributionSequence(ctx)
	distributionID := generateIDWithTimestamp("provider", seq, ctx.BlockTime().Unix())

	epoch := k.calculateCurrentEpoch(ctx)

	// Create distribution
	dist := types.NewRewardDistribution(
		distributionID,
		epoch,
		types.RewardSourceProvider,
		recipients,
		ctx.BlockTime(),
		ctx.BlockHeight(),
	)

	// In production, you would transfer rewards from a pool
	// For now, we just record the distribution

	// Save distribution (this also adds claimable rewards)
	if err := k.SetRewardDistribution(ctx, *dist); err != nil {
		return nil, err
	}

	// Emit event
	err := ctx.EventManager().EmitTypedEvent(&types.EventRewardsDistributed{
		DistributionID: distributionID,
		EpochNumber:    epoch,
		Source:         string(types.RewardSourceProvider),
		TotalRewards:   dist.TotalRewards.String(),
		RecipientCount: uint32(len(recipients)),
		DistributedAt:  ctx.BlockTime().Unix(),
	})
	if err != nil {
		k.Logger(ctx).Error("failed to emit provider rewards distributed event", "error", err)
	}

	k.Logger(ctx).Info("provider rewards distributed",
		"distribution_id", distributionID,
		"recipients", len(recipients),
		"total", dist.TotalRewards.String(),
	)

	return dist, nil
}

// DistributeVerificationRewards distributes rewards for identity verifications
func (k Keeper) DistributeVerificationRewards(ctx sdk.Context, verificationResults []VerificationResult) (*types.RewardDistribution, error) {
	if len(verificationResults) == 0 {
		return nil, types.ErrInvalidReward.Wrap("no verification results provided")
	}

	k.Logger(ctx).Info("distributing verification rewards", "verifications", len(verificationResults))

	params := k.GetParams(ctx)

	// Parse base reward amount
	baseReward, err := strconv.ParseInt(params.VerificationRewardAmount, 10, 64)
	if err != nil {
		baseReward = 100 // Default
	}

	// Create recipients for validators who performed verifications
	recipients := make([]types.RewardRecipient, 0, len(verificationResults))

	for _, result := range verificationResults {
		// Calculate reward based on verification complexity
		// Higher scores = more thorough verification = higher reward
		scoreMultiplier := sdkmath.LegacyNewDecWithPrec(int64(result.Score), 2)
		rewardAmount := sdkmath.LegacyNewDec(baseReward).Mul(scoreMultiplier).Add(sdkmath.LegacyNewDec(baseReward)).TruncateInt()

		recipients = append(recipients, types.RewardRecipient{
			Address:           result.ValidatorAddress,
			Amount:            sdk.NewCoins(sdk.NewCoin("uve", rewardAmount)),
			Reason:            "verification reward",
			VerificationScore: result.Score,
			ReferenceID:       result.AccountAddress,
		})
	}

	if len(recipients) == 0 {
		return nil, types.ErrInvalidReward.Wrap("no rewards to distribute")
	}

	// Generate distribution ID
	seq := k.incrementDistributionSequence(ctx)
	distributionID := generateIDWithTimestamp("verify", seq, ctx.BlockTime().Unix())

	epoch := k.calculateCurrentEpoch(ctx)

	// Create distribution
	dist := types.NewRewardDistribution(
		distributionID,
		epoch,
		types.RewardSourceVerification,
		recipients,
		ctx.BlockTime(),
		ctx.BlockHeight(),
	)

	// Save distribution
	if err := k.SetRewardDistribution(ctx, *dist); err != nil {
		return nil, err
	}

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventRewardsDistributed{
		DistributionID: distributionID,
		EpochNumber:    epoch,
		Source:         string(types.RewardSourceVerification),
		TotalRewards:   dist.TotalRewards.String(),
		RecipientCount: uint32(len(recipients)),
		DistributedAt:  ctx.BlockTime().Unix(),
	})
	if err != nil {
		k.Logger(ctx).Error("failed to emit verification rewards distributed event", "error", err)
	}

	k.Logger(ctx).Info("verification rewards distributed",
		"distribution_id", distributionID,
		"recipients", len(recipients),
		"total", dist.TotalRewards.String(),
	)

	return dist, nil
}

// ClaimRewards claims accumulated rewards for an address
func (k Keeper) ClaimRewards(ctx sdk.Context, claimer sdk.AccAddress, source string) (sdk.Coins, error) {
	rewards, found := k.GetClaimableRewards(ctx, claimer)
	if !found || rewards.TotalClaimable.IsZero() {
		return sdk.Coins{}, types.ErrNoClaimableRewards.Wrapf("no claimable rewards for %s", claimer.String())
	}

	var claimed sdk.Coins
	var claimedEntries []types.RewardEntry

	if source != "" {
		// Claim from specific source
		rewardSource := types.RewardSource(source)
		if !types.IsValidRewardSource(rewardSource) {
			return sdk.Coins{}, types.ErrInvalidReward.Wrapf("invalid reward source: %s", source)
		}
		claimed, claimedEntries = rewards.ClaimBySource(rewardSource, ctx.BlockTime())
	} else {
		// Claim all
		claimed, claimedEntries = rewards.ClaimAll(ctx.BlockTime())
	}

	if claimed.IsZero() {
		return sdk.Coins{}, types.ErrNoClaimableRewards.Wrapf("no rewards to claim from source %s", source)
	}

	// Transfer rewards from module to claimer
	// In production, you would have the rewards already in the module account
	// For now, we just update the claimable rewards record

	// Save updated rewards
	if err := k.SetClaimableRewards(ctx, rewards); err != nil {
		return sdk.Coins{}, err
	}

	// Emit event
	err := ctx.EventManager().EmitTypedEvent(&types.EventRewardsClaimed{
		Claimer:       claimer.String(),
		ClaimedAmount: claimed.String(),
		Source:        source,
		ClaimedAt:     ctx.BlockTime().Unix(),
	})
	if err != nil {
		k.Logger(ctx).Error("failed to emit rewards claimed event", "error", err)
	}

	k.Logger(ctx).Info("rewards claimed",
		"claimer", claimer.String(),
		"amount", claimed.String(),
		"entries", len(claimedEntries),
	)

	return claimed, nil
}

// ProcessRewardExpiry removes expired reward entries
func (k Keeper) ProcessRewardExpiry(ctx sdk.Context) {
	// In production, you would iterate over all claimable rewards
	// and remove expired entries
	k.Logger(ctx).Debug("processing reward expiry")
}

// EndBlockerRewards processes end-of-block reward distributions
func (k Keeper) EndBlockerRewards(ctx sdk.Context) error {
	params := k.GetParams(ctx)

	// Check if we should distribute staking rewards (new epoch)
	currentEpoch := k.calculateCurrentEpoch(ctx)
	blockInEpoch := uint64(ctx.BlockHeight()) % params.StakingRewardEpochLength

	if blockInEpoch == 0 && currentEpoch > 0 {
		// New epoch started, distribute rewards for previous epoch
		previousEpoch := currentEpoch - 1
		if _, err := k.DistributeStakingRewards(ctx, previousEpoch); err != nil {
			k.Logger(ctx).Debug("no staking rewards to distribute", "epoch", previousEpoch)
		}
	}

	// Process expired rewards
	k.ProcessRewardExpiry(ctx)

	return nil
}
