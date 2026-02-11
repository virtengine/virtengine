package keeper_test

import (
	"bytes"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/hpc/types"
)

func TestJobCostEstimationBeforeSubmission(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	providerAddr := sdk.AccAddress(bytes.Repeat([]byte{1}, 20)).String()
	customerAddr := sdk.AccAddress(bytes.Repeat([]byte{2}, 20)).String()

	job := types.HPCJob{
		JobID:           "job-estimate-1",
		ProviderAddress: providerAddr,
		CustomerAddress: customerAddr,
		Resources: types.JobResources{
			Nodes:           2,
			CPUCoresPerNode: 8,
			MemoryGBPerNode: 32,
			GPUsPerNode:     1,
		},
		State: types.JobStateRunning,
	}
	mustSetJob(t, ctx, k, job)

	metrics := &types.HPCDetailedMetrics{
		WallClockSeconds: 1800,
		CPUCoreSeconds:   28800,
		MemoryGBSeconds:  57600,
		GPUSeconds:       3600,
		NodeHours:        sdkmath.LegacyNewDec(1),
		NodesUsed:        2,
	}

	billable, breakdown, err := k.CalculateInterimBilling(ctx, job.JobID, metrics)
	require.NoError(t, err)
	require.NotNil(t, breakdown)
	require.False(t, billable.IsZero())
}

func TestUsageMeteringDuringExecution(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	providerAddr := sdk.AccAddress(bytes.Repeat([]byte{3}, 20)).String()
	customerAddr := sdk.AccAddress(bytes.Repeat([]byte{4}, 20)).String()

	job := types.HPCJob{
		JobID:           "job-meter-1",
		ClusterID:       "cluster-meter",
		ProviderAddress: providerAddr,
		CustomerAddress: customerAddr,
		Resources: types.JobResources{
			Nodes: 1,
		},
		State: types.JobStateRunning,
	}
	mustSetJob(t, ctx, k, job)

	snap1 := &types.HPCUsageSnapshot{
		SnapshotID:      "snap-1",
		JobID:           job.JobID,
		ClusterID:       job.ClusterID,
		SchedulerType:   "SLURM",
		SnapshotType:    types.SnapshotTypeInterim,
		SequenceNumber:  1,
		ProviderAddress: providerAddr,
		CustomerAddress: customerAddr,
		Metrics: types.HPCDetailedMetrics{
			WallClockSeconds: 300,
			CPUCoreSeconds:   1200,
		},
		CumulativeMetrics: types.HPCDetailedMetrics{
			WallClockSeconds: 300,
			CPUCoreSeconds:   1200,
		},
		JobState:          types.JobStateRunning,
		SnapshotTime:      ctx.BlockTime(),
		ProviderSignature: "sig-1",
	}
	require.NoError(t, k.CreateUsageSnapshot(ctx, snap1))

	snap2 := &types.HPCUsageSnapshot{
		SnapshotID:      "snap-2",
		JobID:           job.JobID,
		ClusterID:       job.ClusterID,
		SchedulerType:   "SLURM",
		SnapshotType:    types.SnapshotTypeInterim,
		SequenceNumber:  2,
		ProviderAddress: providerAddr,
		CustomerAddress: customerAddr,
		Metrics: types.HPCDetailedMetrics{
			WallClockSeconds: 900,
			CPUCoreSeconds:   3600,
		},
		CumulativeMetrics: types.HPCDetailedMetrics{
			WallClockSeconds: 1200,
			CPUCoreSeconds:   4800,
		},
		JobState:          types.JobStateRunning,
		SnapshotTime:      ctx.BlockTime().Add(10 * time.Minute),
		ProviderSignature: "sig-2",
	}
	require.NoError(t, k.CreateUsageSnapshot(ctx, snap2))

	latest, found := k.GetLatestUsageSnapshot(ctx, job.JobID)
	require.True(t, found)
	require.Equal(t, int64(1200), latest.CumulativeMetrics.WallClockSeconds)

	record, err := k.CalculateJobBilling(ctx, job.JobID)
	require.NoError(t, err)
	require.Equal(t, latest.CumulativeMetrics.WallClockSeconds, record.UsageMetrics.WallClockSeconds)
}

func TestSettlementAfterJobCompletion(t *testing.T) {
	ctx, k, bank := setupHPCKeeper(t)

	providerAddr := sdk.AccAddress(bytes.Repeat([]byte{5}, 20)).String()
	customerAddr := sdk.AccAddress(bytes.Repeat([]byte{6}, 20)).String()
	start := time.Unix(1_700_000_000, 0).UTC()
	end := start.Add(2 * time.Hour)

	job := types.HPCJob{
		JobID:           "job-settle-1",
		ClusterID:       "cluster-settle",
		ProviderAddress: providerAddr,
		CustomerAddress: customerAddr,
		Resources: types.JobResources{
			Nodes:           1,
			CPUCoresPerNode: 4,
			MemoryGBPerNode: 16,
			GPUsPerNode:     1,
		},
		State:       types.JobStateCompleted,
		CreatedAt:   start,
		StartedAt:   &start,
		CompletedAt: &end,
	}
	mustSetJob(t, ctx, k, job)
	ctx = ctx.WithBlockTime(end)

	record := &types.HPCAccountingRecord{
		RecordID:        "acct-settle-1",
		JobID:           job.JobID,
		ClusterID:       job.ClusterID,
		ProviderAddress: providerAddr,
		CustomerAddress: customerAddr,
		UsageMetrics: types.HPCDetailedMetrics{
			WallClockSeconds: 7200,
			CPUCoreSeconds:   28800,
			MemoryGBSeconds:  115200,
			GPUSeconds:       7200,
			NodeHours:        sdkmath.LegacyNewDec(2),
			NodesUsed:        1,
		},
		BillableAmount: sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1_000_000))),
		BillableBreakdown: types.BillableBreakdown{
			Subtotal: sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1_000_000))),
		},
		ProviderReward: sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(800_000))),
		PlatformFee:    sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(200_000))),
		Status:         types.AccountingStatusPending,
		PeriodStart:    start,
		PeriodEnd:      end,
		FormulaVersion: types.CurrentBillingFormulaVersion,
	}
	require.NoError(t, k.CreateAccountingRecord(ctx, record))

	result, err := k.ProcessJobSettlement(ctx, job.JobID)
	require.NoError(t, err)
	require.True(t, result.Success)
	require.NotEmpty(t, result.SettlementID)

	records := k.GetAccountingRecordsByJob(ctx, job.JobID)
	require.Len(t, records, 1)
	require.Equal(t, types.AccountingStatusSettled, records[0].Status)

	transfers := bank.Transfers()
	require.NotEmpty(t, transfers)
}

func TestBillingInvoiceGeneration(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	billingKeeper := NewMockBillingKeeper()
	k.SetBillingKeeper(billingKeeper)

	providerAddr := sdk.AccAddress(bytes.Repeat([]byte{9}, 20)).String()
	customerAddr := sdk.AccAddress(bytes.Repeat([]byte{10}, 20)).String()
	start := time.Unix(1_700_010_000, 0).UTC()
	end := start.Add(1 * time.Hour)

	job := types.HPCJob{
		JobID:           "job-invoice-1",
		ClusterID:       "cluster-invoice",
		ProviderAddress: providerAddr,
		CustomerAddress: customerAddr,
		EscrowID:        "escrow-invoice",
		State:           types.JobStateCompleted,
		CreatedAt:       start,
		StartedAt:       &start,
		CompletedAt:     &end,
	}
	mustSetJob(t, ctx, k, job)

	record := &types.HPCAccountingRecord{
		RecordID:        "acct-invoice-1",
		JobID:           job.JobID,
		ClusterID:       job.ClusterID,
		ProviderAddress: providerAddr,
		CustomerAddress: customerAddr,
		UsageMetrics: types.HPCDetailedMetrics{
			WallClockSeconds: 3600,
			CPUCoreSeconds:   14400,
			MemoryGBSeconds:  28800,
			GPUSeconds:       3600,
			NodeHours:        sdkmath.LegacyNewDec(1),
			NodesUsed:        1,
		},
		BillableAmount: sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(500_000))),
		BillableBreakdown: types.BillableBreakdown{
			Subtotal: sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(500_000))),
		},
		ProviderReward: sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(400_000))),
		PlatformFee:    sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(100_000))),
		Status:         types.AccountingStatusPending,
		PeriodStart:    start,
		PeriodEnd:      end,
		FormulaVersion: types.CurrentBillingFormulaVersion,
	}
	require.NoError(t, k.CreateAccountingRecord(ctx, record))

	invoiceID, err := k.GenerateInvoiceForJob(ctx, record.RecordID)
	require.NoError(t, err)
	require.NotEmpty(t, invoiceID)

	require.NotEmpty(t, billingKeeper.CreatedInvoices())
	require.NotEmpty(t, billingKeeper.UsageRecords())
}

func TestRefundOnJobCancellation(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	providerAddr := sdk.AccAddress(bytes.Repeat([]byte{7}, 20)).String()
	customerAddr := sdk.AccAddress(bytes.Repeat([]byte{8}, 20)).String()

	job := types.HPCJob{
		JobID:           "job-refund-1",
		ClusterID:       "cluster-refund",
		ProviderAddress: providerAddr,
		CustomerAddress: customerAddr,
		State:           types.JobStateCancelled,
		EscrowID:        "escrow-1",
		AgreedPrice:     sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1_000_000))),
	}
	mustSetJob(t, ctx, k, job)

	record := &types.HPCAccountingRecord{
		JobID:           job.JobID,
		ClusterID:       job.ClusterID,
		ProviderAddress: providerAddr,
		CustomerAddress: customerAddr,
		BillableAmount:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(800_000))),
		BillableBreakdown: types.BillableBreakdown{
			Subtotal: sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(800_000))),
		},
		ProviderReward: sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(640_000))),
		PlatformFee:    sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(160_000))),
		Status:         types.AccountingStatusPending,
		PeriodStart:    ctx.BlockTime().Add(-time.Hour),
		PeriodEnd:      ctx.BlockTime(),
		FormulaVersion: types.CurrentBillingFormulaVersion,
	}

	require.NoError(t, k.CreateAccountingRecord(ctx, record))

	result, err := k.ProcessJobSettlement(ctx, job.JobID)
	require.NoError(t, err)
	require.True(t, result.Success)
	require.False(t, result.RefundAmount.IsZero())
}
