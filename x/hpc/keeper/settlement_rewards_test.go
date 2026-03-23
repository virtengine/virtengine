package keeper_test

import (
	"bytes"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/hpc/types"
)

func TestDistributeJobRewardsFromSettlement_NodeLevelRewards(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)

	providerA := sdk.AccAddress(bytes.Repeat([]byte{11}, 20)).String()
	providerB := sdk.AccAddress(bytes.Repeat([]byte{12}, 20)).String()
	customer := sdk.AccAddress(bytes.Repeat([]byte{13}, 20)).String()

	job := types.HPCJob{
		JobID:             "job-settle-node-1",
		OfferingID:        "off-1",
		ClusterID:         "cluster-node-rewards",
		ProviderAddress:   providerA,
		CustomerAddress:   customer,
		State:             types.JobStateCompleted,
		QueueName:         "default",
		WorkloadSpec:      types.JobWorkloadSpec{ContainerImage: "repo/image:latest"},
		Resources:         types.JobResources{Nodes: 2},
		MaxRuntimeSeconds: 3600,
		AgreedPrice:       sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1_000))),
	}
	mustSetJob(t, ctx, k, job)

	mustSetNode(t, ctx, k, types.NodeMetadata{
		NodeID:          "node-fast",
		ClusterID:       job.ClusterID,
		ProviderAddress: providerA,
		Region:          "us-east-1",
		AvgLatencyMs:    10,
		Active:          true,
	})
	mustSetNode(t, ctx, k, types.NodeMetadata{
		NodeID:          "node-slow",
		ClusterID:       job.ClusterID,
		ProviderAddress: providerB,
		Region:          "us-east-1",
		AvgLatencyMs:    30,
		Active:          true,
	})

	record := &types.HPCAccountingRecord{
		JobID:           job.JobID,
		ClusterID:       job.ClusterID,
		ProviderAddress: providerA,
		CustomerAddress: customer,
		ProviderReward:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1_000))),
		BillableAmount:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1_000))),
		UsageMetrics: types.HPCDetailedMetrics{
			NodesUsed: 2,
		},
		FormulaVersion: "v1",
	}

	rewardRecord, err := k.DistributeJobRewardsFromSettlement(ctx, job.JobID, record)
	require.NoError(t, err)
	require.Len(t, rewardRecord.Recipients, 2)

	byNode := map[string]types.HPCRewardRecipient{}
	total := sdk.NewCoins()
	for _, recipient := range rewardRecord.Recipients {
		require.Equal(t, "node_operator", recipient.RecipientType)
		byNode[recipient.NodeID] = recipient
		total = total.Add(recipient.Amount...)
	}

	require.Equal(t, record.ProviderReward, total)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(954))), byNode["node-fast"].Amount)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(46))), byNode["node-slow"].Amount)
}

func TestDistributeJobRewardsFromSettlement_FallsBackToProvider(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)

	provider := sdk.AccAddress(bytes.Repeat([]byte{21}, 20)).String()
	customer := sdk.AccAddress(bytes.Repeat([]byte{22}, 20)).String()

	job := types.HPCJob{
		JobID:             "job-settle-provider-fallback",
		OfferingID:        "off-2",
		ClusterID:         "cluster-provider-fallback",
		ProviderAddress:   provider,
		CustomerAddress:   customer,
		State:             types.JobStateCompleted,
		QueueName:         "default",
		WorkloadSpec:      types.JobWorkloadSpec{ContainerImage: "repo/image:latest"},
		Resources:         types.JobResources{Nodes: 1},
		MaxRuntimeSeconds: 3600,
		AgreedPrice:       sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(500))),
	}
	mustSetJob(t, ctx, k, job)

	mustSetNode(t, ctx, k, types.NodeMetadata{
		NodeID:          "node-inactive",
		ClusterID:       job.ClusterID,
		ProviderAddress: provider,
		Region:          "us-east-1",
		AvgLatencyMs:    15,
		Active:          false,
	})

	record := &types.HPCAccountingRecord{
		JobID:           job.JobID,
		ClusterID:       job.ClusterID,
		ProviderAddress: provider,
		CustomerAddress: customer,
		ProviderReward:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(500))),
		BillableAmount:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(500))),
		UsageMetrics: types.HPCDetailedMetrics{
			NodesUsed: 1,
		},
		FormulaVersion: "v1",
	}

	rewardRecord, err := k.DistributeJobRewardsFromSettlement(ctx, job.JobID, record)
	require.NoError(t, err)
	require.Len(t, rewardRecord.Recipients, 1)

	recipient := rewardRecord.Recipients[0]
	require.Equal(t, "provider", recipient.RecipientType)
	require.Equal(t, provider, recipient.Address)
	require.Equal(t, record.ProviderReward, recipient.Amount)
}

func TestDistributeJobRewardsFromSettlement_SelectsBestProximityNodes(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)

	providerA := sdk.AccAddress(bytes.Repeat([]byte{31}, 20)).String()
	providerB := sdk.AccAddress(bytes.Repeat([]byte{32}, 20)).String()
	providerC := sdk.AccAddress(bytes.Repeat([]byte{33}, 20)).String()
	customer := sdk.AccAddress(bytes.Repeat([]byte{34}, 20)).String()

	job := types.HPCJob{
		JobID:             "job-settle-proximity",
		OfferingID:        "off-3",
		ClusterID:         "cluster-settle-proximity",
		ProviderAddress:   providerA,
		CustomerAddress:   customer,
		State:             types.JobStateCompleted,
		QueueName:         "default",
		WorkloadSpec:      types.JobWorkloadSpec{ContainerImage: "repo/image:latest"},
		Resources:         types.JobResources{Nodes: 2},
		MaxRuntimeSeconds: 3600,
		AgreedPrice:       sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1_000))),
	}
	mustSetJob(t, ctx, k, job)

	mustSetNode(t, ctx, k, types.NodeMetadata{
		NodeID:              "node-fast",
		ClusterID:           job.ClusterID,
		ProviderAddress:     providerA,
		Region:              "us-east-1",
		AvgLatencyMs:        5,
		LatencyMeasurements: []types.LatencyMeasurement{{TargetNodeID: "node-mid", LatencyMs: 120}, {TargetNodeID: "node-close", LatencyMs: 4}},
		Active:              true,
	})
	mustSetNode(t, ctx, k, types.NodeMetadata{
		NodeID:              "node-mid",
		ClusterID:           job.ClusterID,
		ProviderAddress:     providerB,
		Region:              "us-east-1",
		AvgLatencyMs:        6,
		LatencyMeasurements: []types.LatencyMeasurement{{TargetNodeID: "node-fast", LatencyMs: 120}, {TargetNodeID: "node-close", LatencyMs: 40}},
		Active:              true,
	})
	mustSetNode(t, ctx, k, types.NodeMetadata{
		NodeID:              "node-close",
		ClusterID:           job.ClusterID,
		ProviderAddress:     providerC,
		Region:              "us-east-1",
		AvgLatencyMs:        8,
		LatencyMeasurements: []types.LatencyMeasurement{{TargetNodeID: "node-fast", LatencyMs: 4}, {TargetNodeID: "node-mid", LatencyMs: 40}},
		Active:              true,
	})

	record := &types.HPCAccountingRecord{
		RecordID:        "acct-proximity",
		JobID:           job.JobID,
		ClusterID:       job.ClusterID,
		ProviderAddress: providerA,
		CustomerAddress: customer,
		ProviderReward:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1_000))),
		BillableAmount:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1_000))),
		UsageMetrics: types.HPCDetailedMetrics{
			NodesUsed: 2,
		},
		FormulaVersion: "v1",
	}

	rewardRecord, err := k.DistributeJobRewardsFromSettlement(ctx, job.JobID, record)
	require.NoError(t, err)
	require.Len(t, rewardRecord.Recipients, 2)
	require.NoError(t, rewardRecord.Validate())
	require.NotEmpty(t, rewardRecord.RewardID)
	require.Equal(t, ctx.BlockHeight(), rewardRecord.BlockHeight)

	byNode := map[string]types.HPCRewardRecipient{}
	for _, recipient := range rewardRecord.Recipients {
		byNode[recipient.NodeID] = recipient
	}

	require.Contains(t, byNode, "node-fast")
	require.Contains(t, byNode, "node-close")
	require.NotContains(t, byNode, "node-mid")

	stored, found := k.GetHPCReward(ctx, rewardRecord.RewardID)
	require.True(t, found)
	require.Equal(t, rewardRecord.RewardID, stored.RewardID)
	require.Equal(t, rewardRecord.TotalReward, stored.TotalReward)

	byJob := k.GetRewardsByJob(ctx, job.JobID)
	require.Len(t, byJob, 1)
	require.Equal(t, rewardRecord.RewardID, byJob[0].RewardID)
}
