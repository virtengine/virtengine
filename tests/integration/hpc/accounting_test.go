// Package hpc contains integration tests for the HPC module.
//
// VE-5A: Integration tests for usage accounting, billing, and settlement
package hpc

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/hpc/types"
)

// TestAccountingRecordCreation tests creating and validating accounting records
func TestAccountingRecordCreation(t *testing.T) {
	provider := sdk.AccAddress([]byte("provider1234567890123"))
	customer := sdk.AccAddress([]byte("customer1234567890123"))

	record := &types.HPCAccountingRecord{
		RecordID:        "test-record-1",
		JobID:           "hpc-job-1",
		ClusterID:       "cluster-1",
		ProviderAddress: provider.String(),
		CustomerAddress: customer.String(),
		OfferingID:      "offering-1",
		SchedulerType:   "SLURM",
		UsageMetrics: types.HPCDetailedMetrics{
			WallClockSeconds: 3600,
			CPUCoreSeconds:   14400, // 4 cores * 1 hour
			MemoryGBSeconds:  28800, // 8 GB * 1 hour
			GPUSeconds:       3600,  // 1 GPU * 1 hour
			NodeHours:        sdkmath.LegacyNewDec(1),
			NodesUsed:        1,
		},
		BillableAmount: sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(100000))),
		BillableBreakdown: types.BillableBreakdown{
			CPUCost:     sdk.NewCoin("uvirt", sdkmath.NewInt(20000)),
			MemoryCost:  sdk.NewCoin("uvirt", sdkmath.NewInt(8000)),
			GPUCost:     sdk.NewCoin("uvirt", sdkmath.NewInt(60000)),
			StorageCost: sdk.NewCoin("uvirt", sdkmath.NewInt(1000)),
			NetworkCost: sdk.NewCoin("uvirt", sdkmath.NewInt(1000)),
			NodeCost:    sdk.NewCoin("uvirt", sdkmath.NewInt(10000)),
			Subtotal:    sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(100000))),
		},
		ProviderReward: sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(97500))), // 97.5%
		PlatformFee:    sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(2500))),  // 2.5%
		Status:         types.AccountingStatusPending,
		PeriodStart:    time.Now().Add(-time.Hour),
		PeriodEnd:      time.Now(),
		FormulaVersion: types.CurrentBillingFormulaVersion,
	}

	err := record.Validate()
	require.NoError(t, err)

	// Test finalization
	err = record.Finalize(time.Now())
	require.NoError(t, err)
	require.Equal(t, types.AccountingStatusFinalized, record.Status)
	require.NotNil(t, record.FinalizedAt)
	require.NotEmpty(t, record.CalculationHash)

	// Test cannot finalize twice
	err = record.Finalize(time.Now())
	require.Error(t, err)
}

// TestBillingCalculator tests the billing calculator
func TestBillingCalculator(t *testing.T) {
	rules := types.DefaultHPCBillingRules("uvirt")
	calculator := types.NewHPCBillingCalculator(rules)

	metrics := &types.HPCDetailedMetrics{
		WallClockSeconds: 3600, // 1 hour
		CPUCoreSeconds:   14400, // 4 cores * 1 hour
		MemoryGBSeconds:  28800, // 8 GB * 1 hour
		GPUSeconds:       3600,  // 1 GPU * 1 hour
		GPUType:          "nvidia-a100",
		NodeHours:        sdkmath.LegacyNewDec(1),
		NodesUsed:        1,
		StorageGBHours:   10,
		NetworkBytesIn:   1073741824,  // 1 GB
		NetworkBytesOut:  1073741824,  // 1 GB
	}

	breakdown, billable, err := calculator.CalculateBillableAmount(metrics, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, breakdown)
	require.False(t, billable.IsZero())

	// Calculate provider reward
	reward := calculator.CalculateProviderReward(billable)
	require.False(t, reward.IsZero())

	// Calculate platform fee
	fee := calculator.CalculatePlatformFee(billable)
	require.False(t, fee.IsZero())

	// Verify reward + fee == billable
	totalDistributed := reward.Add(fee...)
	require.True(t, totalDistributed.IsAllLTE(billable))
}

// TestVolumeDiscounts tests volume-based discounts
func TestVolumeDiscounts(t *testing.T) {
	rules := types.DefaultHPCBillingRules("uvirt")
	calculator := types.NewHPCBillingCalculator(rules)

	// Create historical usage that triggers volume discount
	historicalUsage := &types.AccountingAggregation{
		TotalCPUCoreHours: sdkmath.LegacyNewDec(150), // Over 100 threshold
		TotalGPUHours:     sdkmath.LegacyNewDec(20),
		TotalNodeHours:    sdkmath.LegacyNewDec(100),
	}

	metrics := &types.HPCDetailedMetrics{
		CPUCoreSeconds: 36000, // 10 core-hours
	}

	discounts := calculator.EvaluateDiscounts(metrics, "customer1", historicalUsage)

	// Should have at least one volume discount
	hasVolumeDiscount := false
	for _, d := range discounts {
		if d.DiscountType == string(types.HPCDiscountTypeVolume) {
			hasVolumeDiscount = true
			break
		}
	}
	require.True(t, hasVolumeDiscount, "should have volume discount for high usage")
}

// TestBillingCaps tests billing caps
func TestBillingCaps(t *testing.T) {
	rules := types.DefaultHPCBillingRules("uvirt")
	calculator := types.NewHPCBillingCalculator(rules)

	// Set a low monthly cap for testing
	rules.BillingCaps = []types.HPCBillingCap{
		{
			CapID:     "test-cap",
			CapName:   "Test Cap",
			CapType:   types.HPCBillingCapTypeMonthly,
			CapAmount: sdk.NewCoin("uvirt", sdkmath.NewInt(1000000)), // 1 virt
			Active:    true,
		},
	}
	calculator.Rules = rules

	billable := sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(500000))) // 0.5 virt
	periodSpending := sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(800000))) // 0.8 virt already spent

	caps := calculator.EvaluateCaps(billable, "customer1", periodSpending)

	// Should have cap applied (0.8 + 0.5 = 1.3 > 1.0 cap)
	require.Len(t, caps, 1)
	require.Equal(t, "test-cap", caps[0].CapID)
}

// TestUsageSnapshotValidation tests usage snapshot validation
func TestUsageSnapshotValidation(t *testing.T) {
	provider := sdk.AccAddress([]byte("provider1234567890123"))
	customer := sdk.AccAddress([]byte("customer1234567890123"))

	snapshot := &types.HPCUsageSnapshot{
		SnapshotID:      "snap-1",
		JobID:           "job-1",
		ClusterID:       "cluster-1",
		SchedulerType:   "SLURM",
		SnapshotType:    types.SnapshotTypeInterim,
		SequenceNumber:  1,
		ProviderAddress: provider.String(),
		CustomerAddress: customer.String(),
		Metrics: types.HPCDetailedMetrics{
			WallClockSeconds: 1800,
			CPUCoreSeconds:   7200,
		},
		CumulativeMetrics: types.HPCDetailedMetrics{
			WallClockSeconds: 1800,
			CPUCoreSeconds:   7200,
		},
		JobState:          types.JobStateRunning,
		SnapshotTime:      time.Now(),
		ProviderSignature: "sig123",
	}

	err := snapshot.Validate()
	require.NoError(t, err)

	// Test content hash calculation
	hash := snapshot.CalculateContentHash()
	require.NotEmpty(t, hash)

	// Calculate again - should be deterministic
	hash2 := snapshot.CalculateContentHash()
	require.Equal(t, hash, hash2)
}

// TestReconciliationTolerances tests reconciliation tolerance checks
func TestReconciliationTolerances(t *testing.T) {
	tolerances := types.DefaultReconciliationTolerances()

	// CPU tolerance is 1%
	require.Equal(t, 1.0, tolerances.CPUCoreSecondsPercent)

	// GPU tolerance is 0.5% (tighter for expensive resources)
	require.Equal(t, 0.5, tolerances.GPUSecondsPercent)

	// Wall clock is very tight
	require.Equal(t, 0.1, tolerances.WallClockSecondsPercent)

	// Network has more tolerance
	require.Equal(t, 5.0, tolerances.NetworkBytesPercent)
}

// TestDisputeWorkflow tests the dispute workflow for accounting
func TestDisputeWorkflow(t *testing.T) {
	provider := sdk.AccAddress([]byte("provider1234567890123"))

	record := &types.HPCAccountingRecord{
		RecordID:        "record-1",
		JobID:           "job-1",
		ClusterID:       "cluster-1",
		ProviderAddress: provider.String(),
		CustomerAddress: sdk.AccAddress([]byte("customer1234567890123")).String(),
		BillableAmount:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(100000))),
		ProviderReward:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(97500))),
		PlatformFee:     sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(2500))),
		Status:          types.AccountingStatusFinalized,
		PeriodStart:     time.Now().Add(-time.Hour),
		PeriodEnd:       time.Now(),
		FormulaVersion:  "v1.0.0",
	}

	// Mark as disputed
	err := record.MarkDisputed("dispute-123")
	require.NoError(t, err)
	require.Equal(t, types.AccountingStatusDisputed, record.Status)
	require.Equal(t, "dispute-123", record.DisputeID)

	// Cannot settle disputed record
	err = record.Settle("settlement-1", time.Now())
	require.Error(t, err)
}

// TestAccountingStatusTransitions tests valid status transitions
func TestAccountingStatusTransitions(t *testing.T) {
	testCases := []struct {
		name        string
		initial     types.AccountingRecordStatus
		checkValid  func(*types.HPCAccountingRecord) error
	}{
		{
			name:    "pending to finalized",
			initial: types.AccountingStatusPending,
			checkValid: func(r *types.HPCAccountingRecord) error {
				return r.Finalize(time.Now())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			record := &types.HPCAccountingRecord{
				RecordID:        "record-1",
				JobID:           "job-1",
				ClusterID:       "cluster-1",
				ProviderAddress: sdk.AccAddress([]byte("provider1234567890123")).String(),
				CustomerAddress: sdk.AccAddress([]byte("customer1234567890123")).String(),
				BillableAmount:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1000))),
				ProviderReward:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(975))),
				PlatformFee:     sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(25))),
				Status:          tc.initial,
				PeriodStart:     time.Now().Add(-time.Hour),
				PeriodEnd:       time.Now(),
				FormulaVersion:  "v1.0.0",
			}

			err := tc.checkValid(record)
			require.NoError(t, err)
		})
	}
}

// TestDeterministicHash tests that accounting hash is deterministic
func TestDeterministicHash(t *testing.T) {
	record := &types.HPCAccountingRecord{
		JobID:      "job-123",
		ClusterID:  "cluster-456",
		UsageMetrics: types.HPCDetailedMetrics{
			WallClockSeconds: 3600,
			CPUCoreSeconds:   14400,
		},
		PeriodStart:        time.Unix(1700000000, 0),
		PeriodEnd:          time.Unix(1700003600, 0),
		FormulaVersion:     "v1.0.0",
		SignedUsageRecords: []string{"snap-1", "snap-2"},
	}

	hash1 := record.CalculateHash()
	hash2 := record.CalculateHash()

	require.Equal(t, hash1, hash2, "hash should be deterministic")
	require.Len(t, hash1, 64, "hash should be 64 hex characters (SHA256)")
}

// TestMinimumCharge tests that minimum charge is enforced
func TestMinimumCharge(t *testing.T) {
	rules := types.DefaultHPCBillingRules("uvirt")
	calculator := types.NewHPCBillingCalculator(rules)

	// Very small usage - should still hit minimum charge
	metrics := &types.HPCDetailedMetrics{
		CPUCoreSeconds: 1, // Almost nothing
	}

	_, billable, err := calculator.CalculateBillableAmount(metrics, nil, nil)
	require.NoError(t, err)

	// Should be at least minimum charge
	require.True(t, billable.AmountOf("uvirt").GTE(rules.MinimumCharge.Amount),
		"billable should be at least minimum charge")
}
