// Package keeper implements the HPC module keeper.
//
// VE-504: Rewards distribution for HPC contributors
package keeper

import (
	"encoding/json"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/hpc/types"
)

// ============================================================================
// Job Accounting
// ============================================================================

// CreateJobAccounting creates job accounting for a job
func (k Keeper) CreateJobAccounting(ctx sdk.Context, accounting *types.JobAccounting) error {
	if err := accounting.Validate(); err != nil {
		return err
	}

	accounting.CreatedAt = ctx.BlockTime()
	accounting.BlockHeight = ctx.BlockHeight()
	accounting.SettlementStatus = "pending"

	return k.SetJobAccounting(ctx, *accounting)
}

// FinalizeJobAccounting finalizes job accounting and triggers reward distribution
func (k Keeper) FinalizeJobAccounting(ctx sdk.Context, jobID string) error {
	accounting, exists := k.GetJobAccounting(ctx, jobID)
	if !exists {
		return types.ErrJobAccountingNotFound
	}

	job, exists := k.GetJob(ctx, jobID)
	if !exists {
		return types.ErrJobNotFound
	}

	if !types.IsTerminalJobState(job.State) {
		return types.ErrInvalidJob.Wrap("job is not in terminal state")
	}

	now := ctx.BlockTime()
	accounting.FinalizedAt = &now
	accounting.JobCompletionStatus = job.State
	accounting.SettlementStatus = "finalized"

	if err := k.SetJobAccounting(ctx, accounting); err != nil {
		return err
	}

	// Trigger reward distribution if job completed successfully
	if job.State == types.JobStateCompleted {
		_, err := k.DistributeJobRewards(ctx, jobID)
		if err != nil {
			k.Logger(ctx).Error("failed to distribute job rewards", "jobID", jobID, "error", err)
			// Don't fail the finalization, just log the error
		}
	}

	return nil
}

// GetJobAccounting retrieves job accounting by job ID
func (k Keeper) GetJobAccounting(ctx sdk.Context, jobID string) (types.JobAccounting, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GetJobAccountingKey(jobID))
	if bz == nil {
		return types.JobAccounting{}, false
	}

	var accounting types.JobAccounting
	if err := json.Unmarshal(bz, &accounting); err != nil {
		return types.JobAccounting{}, false
	}
	return accounting, true
}

// SetJobAccounting stores job accounting
func (k Keeper) SetJobAccounting(ctx sdk.Context, accounting types.JobAccounting) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(accounting)
	if err != nil {
		return err
	}
	store.Set(types.GetJobAccountingKey(accounting.JobID), bz)
	return nil
}

// ============================================================================
// Rewards Distribution
// ============================================================================

// DistributeJobRewards distributes rewards for a completed HPC job
// Uses deterministic fixed-point arithmetic for on-chain calculations
func (k Keeper) DistributeJobRewards(ctx sdk.Context, jobID string) (*types.HPCRewardRecord, error) {
	job, exists := k.GetJob(ctx, jobID)
	if !exists {
		return nil, types.ErrJobNotFound
	}

	accounting, exists := k.GetJobAccounting(ctx, jobID)
	if !exists {
		return nil, types.ErrJobAccountingNotFound
	}

	params := k.GetParams(ctx)

	// Calculate rewards using deterministic fixed-point math
	totalCost := accounting.TotalCost

	// Parse rates as fixed-point integers (scale: 1000000)
	platformFeeRate := parseFixedPoint(params.PlatformFeeRate)
	providerRewardRate := parseFixedPoint(params.ProviderRewardRate)
	nodeRewardRate := parseFixedPoint(params.NodeRewardRate)

	// Calculate splits using fixed-point arithmetic
	// platformFee = totalCost * platformFeeRate / 1000000
	// providerReward = totalCost * providerRewardRate / 1000000
	// nodeRewards = totalCost * nodeRewardRate / 1000000

	recipients := []types.HPCRewardRecipient{}

	// Calculate platform fee
	platformFee := calculateFixedPointShare(totalCost, platformFeeRate)
	if !platformFee.IsZero() {
		recipients = append(recipients, types.HPCRewardRecipient{
			Address:            k.authority, // Platform receives fee
			Amount:             platformFee,
			RecipientType:      "platform",
			ContributionWeight: params.PlatformFeeRate,
			Reason:             "HPC platform fee",
		})
	}

	// Calculate provider reward
	providerReward := calculateFixedPointShare(totalCost, providerRewardRate)
	if !providerReward.IsZero() {
		recipients = append(recipients, types.HPCRewardRecipient{
			Address:            job.ProviderAddress,
			Amount:             providerReward,
			RecipientType:      "provider",
			ContributionWeight: params.ProviderRewardRate,
			Reason:             "HPC job provider reward",
		})
	}

	// Calculate node rewards
	nodes := k.GetNodesByCluster(ctx, job.ClusterID)
	activeNodes := []types.NodeMetadata{}
	for _, node := range nodes {
		if node.Active {
			activeNodes = append(activeNodes, node)
		}
	}

	if len(activeNodes) > 0 {
		nodePool := calculateFixedPointShare(totalCost, nodeRewardRate)
		perNodeReward := divideCoinsEqually(nodePool, len(activeNodes))

		for _, node := range activeNodes {
			if !perNodeReward.IsZero() {
				// Calculate contribution weight for this node
				contributionWeight := formatFixedPoint(FixedPointScale / int64(len(activeNodes)))

				recipients = append(recipients, types.HPCRewardRecipient{
					Address:            node.ProviderAddress,
					Amount:             perNodeReward,
					RecipientType:      "node_operator",
					NodeID:             node.NodeID,
					ContributionWeight: contributionWeight,
					Reason:             fmt.Sprintf("HPC node reward for node %s", node.NodeID),
				})
			}
		}
	}

	// Calculate total reward
	totalReward := sdk.NewCoins()
	for _, r := range recipients {
		totalReward = totalReward.Add(r.Amount...)
	}

	// Create reward record
	rewardID := fmt.Sprintf("hpc-reward-%s", jobID)
	reward := &types.HPCRewardRecord{
		RewardID:               rewardID,
		JobID:                  jobID,
		ClusterID:              job.ClusterID,
		Source:                 types.HPCRewardSourceJobCompletion,
		TotalReward:            totalReward,
		Recipients:             recipients,
		ReferencedUsageRecords: accounting.SignedUsageRecordIDs,
		JobCompletionStatus:    job.State,
		FormulaVersion:         params.RewardFormulaVersion,
		CalculationDetails: types.RewardCalculationDetails{
			TotalUsageValue:         formatCoinsValue(totalCost),
			PlatformFeeRate:         params.PlatformFeeRate,
			NodeContributionFormula: "equal_split",
			InputMetrics: map[string]string{
				"node_hours":     fmt.Sprintf("%d", accounting.UsageMetrics.NodeHours),
				"cpu_seconds":    fmt.Sprintf("%d", accounting.UsageMetrics.CPUCoreSeconds),
				"memory_seconds": fmt.Sprintf("%d", accounting.UsageMetrics.MemoryGBSeconds),
				"gpu_seconds":    fmt.Sprintf("%d", accounting.UsageMetrics.GPUSeconds),
			},
		},
		Disputed:    false,
		IssuedAt:    ctx.BlockTime(),
		BlockHeight: ctx.BlockHeight(),
	}

	if err := reward.Validate(); err != nil {
		return nil, err
	}

	if err := k.SetHPCReward(ctx, *reward); err != nil {
		return nil, err
	}

	// Execute actual token transfers
	for _, recipient := range recipients {
		recipientAddr, err := sdk.AccAddressFromBech32(recipient.Address)
		if err != nil {
			k.Logger(ctx).Error("invalid recipient address", "address", recipient.Address, "error", err)
			continue
		}

		// Transfer from module to recipient
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddr, recipient.Amount); err != nil {
			k.Logger(ctx).Error("failed to transfer reward", "recipient", recipient.Address, "amount", recipient.Amount, "error", err)
			// Continue with other recipients
		}
	}

	return reward, nil
}

// GetHPCReward retrieves an HPC reward by ID
func (k Keeper) GetHPCReward(ctx sdk.Context, rewardID string) (types.HPCRewardRecord, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GetHPCRewardKey(rewardID))
	if bz == nil {
		return types.HPCRewardRecord{}, false
	}

	var reward types.HPCRewardRecord
	if err := json.Unmarshal(bz, &reward); err != nil {
		return types.HPCRewardRecord{}, false
	}
	return reward, true
}

// GetRewardsByJob retrieves rewards by job ID
func (k Keeper) GetRewardsByJob(ctx sdk.Context, jobID string) []types.HPCRewardRecord {
	var rewards []types.HPCRewardRecord
	k.WithHPCRewards(ctx, func(reward types.HPCRewardRecord) bool {
		if reward.JobID == jobID {
			rewards = append(rewards, reward)
		}
		return false
	})
	return rewards
}

// SetHPCReward stores an HPC reward record
func (k Keeper) SetHPCReward(ctx sdk.Context, reward types.HPCRewardRecord) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(reward)
	if err != nil {
		return err
	}
	store.Set(types.GetHPCRewardKey(reward.RewardID), bz)
	return nil
}

// WithHPCRewards iterates over all HPC rewards
func (k Keeper) WithHPCRewards(ctx sdk.Context, fn func(types.HPCRewardRecord) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.HPCRewardPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var reward types.HPCRewardRecord
		if err := json.Unmarshal(iter.Value(), &reward); err != nil {
			continue
		}
		if fn(reward) {
			break
		}
	}
}

// ============================================================================
// Disputes
// ============================================================================

// FlagDispute flags a dispute for moderation
func (k Keeper) FlagDispute(ctx sdk.Context, dispute *types.HPCDispute) error {
	// Verify job exists
	job, exists := k.GetJob(ctx, dispute.JobID)
	if !exists {
		return types.ErrJobNotFound
	}

	// Only customer or provider can flag dispute
	if dispute.DisputerAddress != job.CustomerAddress && dispute.DisputerAddress != job.ProviderAddress {
		return types.ErrUnauthorized
	}

	// Generate dispute ID
	seq := k.incrementSequence(ctx, types.SequenceKeyDispute)
	dispute.DisputeID = fmt.Sprintf("hpc-dispute-%d", seq)
	dispute.Status = types.DisputeStatusPending
	dispute.CreatedAt = ctx.BlockTime()
	dispute.BlockHeight = ctx.BlockHeight()

	if err := dispute.Validate(); err != nil {
		return err
	}

	// Mark related reward as disputed if applicable
	if dispute.RewardID != "" {
		reward, exists := k.GetHPCReward(ctx, dispute.RewardID)
		if exists {
			reward.Disputed = true
			reward.DisputeID = dispute.DisputeID
			_ = k.SetHPCReward(ctx, reward)
		}
	}

	return k.SetDispute(ctx, *dispute)
}

// ResolveDispute resolves a dispute
func (k Keeper) ResolveDispute(ctx sdk.Context, disputeID string, status types.DisputeStatus, resolution string, resolverAddr sdk.AccAddress) error {
	dispute, exists := k.GetDispute(ctx, disputeID)
	if !exists {
		return types.ErrDisputeNotFound
	}

	// Only pending or under_review disputes can be resolved
	if dispute.Status != types.DisputeStatusPending && dispute.Status != types.DisputeStatusUnderReview {
		return types.ErrInvalidDispute.Wrap("dispute is not in resolvable state")
	}

	now := ctx.BlockTime()
	dispute.Status = status
	dispute.Resolution = resolution
	dispute.ResolverAddress = resolverAddr.String()
	dispute.ResolvedAt = &now

	return k.SetDispute(ctx, dispute)
}

// GetDispute retrieves a dispute by ID
func (k Keeper) GetDispute(ctx sdk.Context, disputeID string) (types.HPCDispute, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GetDisputeKey(disputeID))
	if bz == nil {
		return types.HPCDispute{}, false
	}

	var dispute types.HPCDispute
	if err := json.Unmarshal(bz, &dispute); err != nil {
		return types.HPCDispute{}, false
	}
	return dispute, true
}

// SetDispute stores a dispute
func (k Keeper) SetDispute(ctx sdk.Context, dispute types.HPCDispute) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(dispute)
	if err != nil {
		return err
	}
	store.Set(types.GetDisputeKey(dispute.DisputeID), bz)
	return nil
}

// WithDisputes iterates over all disputes
func (k Keeper) WithDisputes(ctx sdk.Context, fn func(types.HPCDispute) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.DisputePrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var dispute types.HPCDispute
		if err := json.Unmarshal(iter.Value(), &dispute); err != nil {
			continue
		}
		if fn(dispute) {
			break
		}
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// calculateFixedPointShare calculates a share using fixed-point arithmetic
// share = coins * rate / 1000000
func calculateFixedPointShare(coins sdk.Coins, rate int64) sdk.Coins {
	result := sdk.NewCoins()
	for _, coin := range coins {
		// Use integer arithmetic to avoid floating point
		// newAmount = amount * rate / FixedPointScale
		newAmount := coin.Amount.MulRaw(rate).QuoRaw(FixedPointScale)
		if newAmount.IsPositive() {
			result = result.Add(sdk.NewCoin(coin.Denom, newAmount))
		}
	}
	return result
}

// divideCoinsEqually divides coins equally among n recipients
func divideCoinsEqually(coins sdk.Coins, n int) sdk.Coins {
	if n <= 0 {
		return sdk.NewCoins()
	}
	result := sdk.NewCoins()
	for _, coin := range coins {
		newAmount := coin.Amount.QuoRaw(int64(n))
		if newAmount.IsPositive() {
			result = result.Add(sdk.NewCoin(coin.Denom, newAmount))
		}
	}
	return result
}

// formatCoinsValue formats coins to a string value
func formatCoinsValue(coins sdk.Coins) string {
	if len(coins) == 0 {
		return "0"
	}
	return coins.String()
}
