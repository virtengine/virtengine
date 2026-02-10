package keeper_test

import (
	"bytes"
	"strconv"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/hpc/keeper"
	"github.com/virtengine/virtengine/x/hpc/types"
)

func TestDistributeJobRewardsProviderCalculation(t *testing.T) {
	ctx, k, bank := setupHPCKeeper(t)
	providerAddr := sdk.AccAddress(bytes.Repeat([]byte{1}, 20)).String()
	customerAddr := sdk.AccAddress(bytes.Repeat([]byte{2}, 20)).String()

	job := types.HPCJob{
		JobID:           "job-reward-1",
		ClusterID:       "cluster-reward-1",
		ProviderAddress: providerAddr,
		CustomerAddress: customerAddr,
		State:           types.JobStateCompleted,
	}
	mustSetJob(t, ctx, k, job)

	accounting := &types.JobAccounting{
		JobID:           job.JobID,
		ClusterID:       job.ClusterID,
		ProviderAddress: providerAddr,
		CustomerAddress: customerAddr,
		TotalCost:       sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1_000_000))),
		UsageMetrics: types.HPCUsageMetrics{
			NodeHours:        10,
			CPUCoreSeconds:   3600,
			MemoryGBSeconds:  7200,
			GPUSeconds:       1800,
			NodesUsed:        2,
			WallClockSeconds: 3600,
		},
		JobCompletionStatus: types.JobStateCompleted,
	}
	require.NoError(t, k.CreateJobAccounting(ctx, accounting))

	mustSetNode(t, ctx, k, types.NodeMetadata{
		NodeID:          "node-1",
		ClusterID:       job.ClusterID,
		ProviderAddress: providerAddr,
		Region:          "us-east-1",
		Active:          true,
	})
	mustSetNode(t, ctx, k, types.NodeMetadata{
		NodeID:          "node-2",
		ClusterID:       job.ClusterID,
		ProviderAddress: providerAddr,
		Region:          "us-east-1",
		Active:          true,
	})

	reward, err := k.DistributeJobRewards(ctx, job.JobID)
	require.NoError(t, err)

	params := k.GetParams(ctx)
	platformRate := parseFixedPointRate(t, params.PlatformFeeRate)
	providerRate := parseFixedPointRate(t, params.ProviderRewardRate)
	nodeRate := parseFixedPointRate(t, params.NodeRewardRate)

	platformExpected := applyRate(accounting.TotalCost, platformRate)
	providerExpected := applyRate(accounting.TotalCost, providerRate)
	nodePool := applyRate(accounting.TotalCost, nodeRate)
	perNode := divideCoins(nodePool, 2)

	recipientTotals := sumRewardsByType(reward.Recipients)
	require.Equal(t, platformExpected, recipientTotals["platform"])
	require.Equal(t, providerExpected, recipientTotals["provider"])
	require.Equal(t, nodePool, recipientTotals["node_operator"])
	require.Equal(t, platformExpected.Add(providerExpected...).Add(nodePool...), reward.TotalReward)

	transfers := bank.Transfers()
	require.Len(t, transfers, len(reward.Recipients))
	require.True(t, recipientTotals["node_operator"].IsAllGTE(perNode))
}

func TestDistributeJobRewardsEarlyCompletionBonus(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	providerAddr := sdk.AccAddress(bytes.Repeat([]byte{3}, 20)).String()
	customerAddr := sdk.AccAddress(bytes.Repeat([]byte{4}, 20)).String()

	baseJob := types.HPCJob{
		JobID:           "job-base",
		ClusterID:       "cluster-base",
		ProviderAddress: providerAddr,
		CustomerAddress: customerAddr,
		State:           types.JobStateCompleted,
	}
	bonusJob := types.HPCJob{
		JobID:           "job-bonus",
		ClusterID:       "cluster-base",
		ProviderAddress: providerAddr,
		CustomerAddress: customerAddr,
		State:           types.JobStateCompleted,
	}
	mustSetJob(t, ctx, k, baseJob)
	mustSetJob(t, ctx, k, bonusJob)

	baseAccounting := &types.JobAccounting{
		JobID:               baseJob.JobID,
		ClusterID:           baseJob.ClusterID,
		ProviderAddress:     providerAddr,
		CustomerAddress:     customerAddr,
		TotalCost:           sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1_000_000))),
		JobCompletionStatus: types.JobStateCompleted,
	}
	bonusAccounting := &types.JobAccounting{
		JobID:               bonusJob.JobID,
		ClusterID:           bonusJob.ClusterID,
		ProviderAddress:     providerAddr,
		CustomerAddress:     customerAddr,
		TotalCost:           sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1_100_000))),
		JobCompletionStatus: types.JobStateCompleted,
	}

	require.NoError(t, k.CreateJobAccounting(ctx, baseAccounting))
	require.NoError(t, k.CreateJobAccounting(ctx, bonusAccounting))

	baseReward, err := k.DistributeJobRewards(ctx, baseJob.JobID)
	require.NoError(t, err)
	bonusReward, err := k.DistributeJobRewards(ctx, bonusJob.JobID)
	require.NoError(t, err)

	require.True(t, bonusReward.TotalReward.IsAllGT(baseReward.TotalReward))
}

func TestDistributeJobRewardsSLAPenalty(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	providerAddr := sdk.AccAddress(bytes.Repeat([]byte{5}, 20)).String()
	customerAddr := sdk.AccAddress(bytes.Repeat([]byte{6}, 20)).String()

	baseJob := types.HPCJob{
		JobID:           "job-sla-base",
		ClusterID:       "cluster-sla",
		ProviderAddress: providerAddr,
		CustomerAddress: customerAddr,
		State:           types.JobStateCompleted,
	}
	penaltyJob := types.HPCJob{
		JobID:           "job-sla-penalty",
		ClusterID:       "cluster-sla",
		ProviderAddress: providerAddr,
		CustomerAddress: customerAddr,
		State:           types.JobStateCompleted,
	}
	mustSetJob(t, ctx, k, baseJob)
	mustSetJob(t, ctx, k, penaltyJob)

	baseAccounting := &types.JobAccounting{
		JobID:               baseJob.JobID,
		ClusterID:           baseJob.ClusterID,
		ProviderAddress:     providerAddr,
		CustomerAddress:     customerAddr,
		TotalCost:           sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1_000_000))),
		JobCompletionStatus: types.JobStateCompleted,
	}
	penaltyAccounting := &types.JobAccounting{
		JobID:               penaltyJob.JobID,
		ClusterID:           penaltyJob.ClusterID,
		ProviderAddress:     providerAddr,
		CustomerAddress:     customerAddr,
		TotalCost:           sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(850_000))),
		JobCompletionStatus: types.JobStateCompleted,
	}

	require.NoError(t, k.CreateJobAccounting(ctx, baseAccounting))
	require.NoError(t, k.CreateJobAccounting(ctx, penaltyAccounting))

	baseReward, err := k.DistributeJobRewards(ctx, baseJob.JobID)
	require.NoError(t, err)
	penaltyReward, err := k.DistributeJobRewards(ctx, penaltyJob.JobID)
	require.NoError(t, err)

	require.True(t, penaltyReward.TotalReward.IsAllLT(baseReward.TotalReward))
}

func TestDistributeJobRewardsMultiProviderSplit(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	jobProvider := sdk.AccAddress(bytes.Repeat([]byte{7}, 20)).String()
	providerA := sdk.AccAddress(bytes.Repeat([]byte{8}, 20)).String()
	providerB := sdk.AccAddress(bytes.Repeat([]byte{9}, 20)).String()
	customerAddr := sdk.AccAddress(bytes.Repeat([]byte{10}, 20)).String()

	job := types.HPCJob{
		JobID:           "job-multi",
		ClusterID:       "cluster-multi",
		ProviderAddress: jobProvider,
		CustomerAddress: customerAddr,
		State:           types.JobStateCompleted,
	}
	mustSetJob(t, ctx, k, job)

	accounting := &types.JobAccounting{
		JobID:               job.JobID,
		ClusterID:           job.ClusterID,
		ProviderAddress:     jobProvider,
		CustomerAddress:     customerAddr,
		TotalCost:           sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(2_000_000))),
		JobCompletionStatus: types.JobStateCompleted,
	}
	require.NoError(t, k.CreateJobAccounting(ctx, accounting))

	mustSetNode(t, ctx, k, types.NodeMetadata{
		NodeID:          "node-a",
		ClusterID:       job.ClusterID,
		ProviderAddress: providerA,
		Region:          "us-east-2",
		Active:          true,
	})
	mustSetNode(t, ctx, k, types.NodeMetadata{
		NodeID:          "node-b",
		ClusterID:       job.ClusterID,
		ProviderAddress: providerB,
		Region:          "us-east-2",
		Active:          true,
	})

	reward, err := k.DistributeJobRewards(ctx, job.JobID)
	require.NoError(t, err)

	nodeRewards := map[string]sdk.Coins{}
	for _, recipient := range reward.Recipients {
		if recipient.RecipientType == "node_operator" {
			nodeRewards[recipient.Address] = nodeRewards[recipient.Address].Add(recipient.Amount...)
		}
	}

	require.Len(t, nodeRewards, 2)
	require.Equal(t, nodeRewards[providerA], nodeRewards[providerB])
}

func parseFixedPointRate(t *testing.T, rate string) int64 {
	value, err := strconv.ParseInt(rate, 10, 64)
	require.NoError(t, err)
	return value
}

func applyRate(total sdk.Coins, rate int64) sdk.Coins {
	out := sdk.NewCoins()
	for _, coin := range total {
		amount := coin.Amount.MulRaw(rate).QuoRaw(keeper.FixedPointScale)
		if amount.IsPositive() {
			out = out.Add(sdk.NewCoin(coin.Denom, amount))
		}
	}
	return out
}

func divideCoins(total sdk.Coins, count int64) sdk.Coins {
	out := sdk.NewCoins()
	for _, coin := range total {
		amount := coin.Amount.QuoRaw(count)
		if amount.IsPositive() {
			out = out.Add(sdk.NewCoin(coin.Denom, amount))
		}
	}
	return out
}

func sumRewardsByType(recipients []types.HPCRewardRecipient) map[string]sdk.Coins {
	result := map[string]sdk.Coins{}
	for _, recipient := range recipients {
		result[recipient.RecipientType] = result[recipient.RecipientType].Add(recipient.Amount...)
	}
	return result
}
