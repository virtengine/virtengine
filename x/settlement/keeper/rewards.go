package keeper

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/virtengine/virtengine/x/settlement/types"
)

const (
	usageTypeCPU     = "cpu"
	usageTypeMemory  = "memory"
	usageTypeStorage = "storage"
	usageTypeGPU     = "gpu"
	usageTypeNetwork = "network"
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

	// For now, we create a placeholder distribution with the module account as recipient
	// This would be integrated with the distribution module in production

	params := k.GetParams(ctx)
	rewardPoolAddr := params.RewardPoolAddress
	if rewardPoolAddr == "" {
		rewardPoolAddr = authtypes.NewModuleAddress(authtypes.FeeCollectorName).String()
	}

	if _, err := sdk.AccAddressFromBech32(rewardPoolAddr); err != nil {
		return nil, types.ErrInvalidReward.Wrap("invalid reward pool address")
	}

	// Generate distribution ID
	seq := k.incrementDistributionSequence(ctx)
	distributionID := generateIDWithTimestamp("staking", seq, ctx.BlockTime().Unix())

	rewardRecipient := types.RewardRecipient{
		Address: rewardPoolAddr,
		Amount:  sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1))), // Minimum valid amount
		Reason:  "staking rewards to reward pool",
	}

	dist := types.NewRewardDistribution(
		distributionID,
		epoch,
		types.RewardSourceStaking,
		[]types.RewardRecipient{rewardRecipient},
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
		RecipientCount: safeUint32FromInt(len(dist.Recipients)),
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

	// Aggregate rewards by provider
	providerRewards := make(map[string]sdk.Coins)
	providerUsage := make(map[string]uint64)
	providerOrderIDs := make(map[string][]string)

	for _, usage := range usageRecords {
		providerUsage[usage.Provider] += usage.UsageUnits
		providerOrderIDs[usage.Provider] = append(providerOrderIDs[usage.Provider], usage.OrderID)
	}

	params := k.GetParams(ctx)
	for _, usage := range usageRecords {
		rewardAmount := k.calculateUsageReward(usage, params)
		if rewardAmount.IsZero() {
			continue
		}
		providerRewards[usage.Provider] = providerRewards[usage.Provider].Add(rewardAmount...)
	}

	recipients := make([]types.RewardRecipient, 0, len(providerRewards))

	for provider, rewards := range providerRewards {
		units := providerUsage[provider]
		if rewards.IsZero() {
			continue
		}

		recipients = append(recipients, types.RewardRecipient{
			Address:     provider,
			Amount:      rewards,
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
		RecipientCount: safeUint32FromInt(len(recipients)),
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

// DistributeUsageRewards distributes usage-based rewards for settled usage records.
func (k Keeper) DistributeUsageRewards(ctx sdk.Context, usageRecords []types.UsageRecord) (*types.RewardDistribution, error) {
	return k.distributeUsageRewardsWithMetadata(ctx, usageRecords, nil)
}

// DistributeUsageRewardsForSettlement distributes usage rewards for a settlement.
func (k Keeper) DistributeUsageRewardsForSettlement(
	ctx sdk.Context,
	settlementID string,
	usageRecords []types.UsageRecord,
) (*types.RewardDistribution, error) {
	if settlementID == "" {
		return nil, types.ErrInvalidReward.Wrap("settlement_id is required")
	}

	if existing, found := k.findUsageRewardDistributionBySettlement(ctx, settlementID); found {
		return &existing, nil
	}

	metadata := map[string]string{
		"settlement_id": settlementID,
	}

	return k.distributeUsageRewardsWithMetadata(ctx, usageRecords, metadata)
}

func (k Keeper) distributeUsageRewardsWithMetadata(
	ctx sdk.Context,
	usageRecords []types.UsageRecord,
	metadata map[string]string,
) (*types.RewardDistribution, error) {
	if len(usageRecords) == 0 {
		return nil, types.ErrInvalidReward.Wrap("no usage records provided")
	}

	params := k.GetParams(ctx)
	if params.UsageRewardRateBps == 0 {
		return nil, types.ErrInvalidReward.Wrap("usage reward rate is zero")
	}

	sorted := make([]types.UsageRecord, len(usageRecords))
	copy(sorted, usageRecords)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Provider == sorted[j].Provider {
			if sorted[i].UsageType == sorted[j].UsageType {
				return sorted[i].UsageID < sorted[j].UsageID
			}
			return sorted[i].UsageType < sorted[j].UsageType
		}
		return sorted[i].Provider < sorted[j].Provider
	})

	recipients := make([]types.RewardRecipient, 0, len(sorted))
	for _, usage := range sorted {
		rewardAmount := k.calculateUsageReward(usage, params)
		if rewardAmount.IsZero() {
			continue
		}

		normalizedType := normalizeUsageType(usage.UsageType)
		recipients = append(recipients, types.RewardRecipient{
			Address:     usage.Provider,
			Amount:      rewardAmount,
			Reason:      fmt.Sprintf("usage_%s_reward", normalizedType),
			UsageUnits:  usage.UsageUnits,
			ReferenceID: usage.UsageID,
		})
	}

	if len(recipients) == 0 {
		return nil, types.ErrInvalidReward.Wrap("no rewards to distribute")
	}

	seq := k.incrementDistributionSequence(ctx)
	distributionID := generateIDWithTimestamp("usage", seq, ctx.BlockTime().Unix())
	epoch := k.calculateCurrentEpoch(ctx)

	dist := types.NewRewardDistribution(
		distributionID,
		epoch,
		types.RewardSourceUsage,
		recipients,
		ctx.BlockTime(),
		ctx.BlockHeight(),
	)

	if len(metadata) > 0 {
		for key, value := range metadata {
			dist.Metadata[key] = value
		}
	}

	if err := k.SetRewardDistribution(ctx, *dist); err != nil {
		return nil, err
	}

	err := ctx.EventManager().EmitTypedEvent(&types.EventRewardsDistributed{
		DistributionID: distributionID,
		EpochNumber:    epoch,
		Source:         string(types.RewardSourceUsage),
		TotalRewards:   dist.TotalRewards.String(),
		RecipientCount: safeUint32FromInt(len(recipients)),
		DistributedAt:  ctx.BlockTime().Unix(),
	})
	if err != nil {
		k.Logger(ctx).Error("failed to emit usage rewards distributed event", "error", err)
	}

	k.Logger(ctx).Info("usage rewards distributed",
		"distribution_id", distributionID,
		"recipients", len(recipients),
		"total", dist.TotalRewards.String(),
	)

	return dist, nil
}

func (k Keeper) findUsageRewardDistributionBySettlement(ctx sdk.Context, settlementID string) (types.RewardDistribution, bool) {
	var foundDist types.RewardDistribution
	found := false

	k.WithRewardDistributions(ctx, func(dist types.RewardDistribution) bool {
		if dist.Source != types.RewardSourceUsage {
			return false
		}
		if dist.Metadata != nil && dist.Metadata["settlement_id"] == settlementID {
			foundDist = dist
			found = true
			return true
		}
		return false
	})

	return foundDist, found
}

func (k Keeper) calculateUsageReward(usage types.UsageRecord, params types.Params) sdk.Coins {
	if usage.TotalCost == nil || usage.TotalCost.IsZero() {
		return sdk.NewCoins()
	}

	resourceMultiplier := usageResourceMultiplierBps(params, usage.UsageType)
	slaMultiplier := usageSLAMultiplierBps(params, usage)
	ackMultiplier := usageAcknowledgementMultiplierBps(params, usage)

	totalRewards := sdk.NewCoins()
	for _, coin := range usage.TotalCost {
		amount := sdkmath.LegacyNewDecFromInt(coin.Amount)
		reward := amount.
			MulInt64(int64(params.UsageRewardRateBps)).
			QuoInt64(10000).
			MulInt64(int64(resourceMultiplier)).
			QuoInt64(10000).
			MulInt64(int64(slaMultiplier)).
			QuoInt64(10000).
			MulInt64(int64(ackMultiplier)).
			QuoInt64(10000)

		rewardAmt := reward.TruncateInt()
		if rewardAmt.IsPositive() {
			totalRewards = totalRewards.Add(sdk.NewCoin(coin.Denom, rewardAmt))
		}
	}

	return totalRewards
}

func usageResourceMultiplierBps(params types.Params, usageType string) uint32 {
	switch normalizeUsageType(usageType) {
	case usageTypeCPU:
		return params.UsageRewardCPUMultiplierBps
	case usageTypeMemory:
		return params.UsageRewardMemoryMultiplierBps
	case usageTypeStorage:
		return params.UsageRewardStorageMultiplierBps
	case usageTypeGPU:
		return params.UsageRewardGPUMultiplierBps
	case usageTypeNetwork:
		return params.UsageRewardNetworkMultiplierBps
	default:
		return params.UsageRewardCPUMultiplierBps
	}
}

func usageSLAMultiplierBps(params types.Params, usage types.UsageRecord) uint32 {
	if params.UsageGracePeriod == 0 {
		return params.UsageRewardSLAOnTimeMultiplierBps
	}

	submittedAt := usage.SubmittedAt
	if submittedAt.IsZero() {
		submittedAt = usage.PeriodEnd
	}

	grace := safeDurationFromSeconds(params.UsageGracePeriod)
	if submittedAt.After(usage.PeriodEnd.Add(grace)) {
		return params.UsageRewardSLALateMultiplierBps
	}

	return params.UsageRewardSLAOnTimeMultiplierBps
}

func usageAcknowledgementMultiplierBps(params types.Params, usage types.UsageRecord) uint32 {
	if usage.CustomerAcknowledged {
		return params.UsageRewardAcknowledgedMultiplierBps
	}
	return params.UsageRewardUnacknowledgedMultiplierBps
}

func normalizeUsageType(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "compute", "cpu", "cpu_core_hours", "core-hour", "core_hours", "corehours", "cpu_hours":
		return usageTypeCPU
	case "memory", "ram", "mem", "memory_gb_hours", "gb-hour", "gb_hours":
		return usageTypeMemory
	case "storage", "disk", "storage_gb_hours":
		return usageTypeStorage
	case "gpu", "gpu_hours", "gpu-hour":
		return usageTypeGPU
	case "network", "bandwidth", "network_gb":
		return usageTypeNetwork
	default:
		return normalized
	}
}

func safeDurationFromSeconds(seconds uint64) time.Duration {
	maxSeconds := uint64(^uint64(0)>>1) / uint64(time.Second)
	if seconds > maxSeconds {
		return time.Duration(maxSeconds) * time.Second
	}
	return time.Duration(seconds) * time.Second
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
		RecipientCount: safeUint32FromInt(len(recipients)),
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
	if err := k.SetClaimableRewards(ctx, claimer, rewards); err != nil {
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
	blockHeight := ctx.BlockHeight()
	if blockHeight < 0 {
		blockHeight = 0
	}
	blockInEpoch := safeUint64FromInt64(blockHeight) % params.StakingRewardEpochLength

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

func safeUint32FromInt(value int) uint32 {
	if value < 0 {
		return 0
	}
	if value > int(^uint32(0)) {
		return ^uint32(0)
	}
	return uint32(value)
}

func safeUint64FromInt64(value int64) uint64 {
	if value < 0 {
		return 0
	}
	if value > int64(^uint64(0)>>1) {
		return ^uint64(0)
	}
	//nolint:gosec // range checked above
	return uint64(value)
}
