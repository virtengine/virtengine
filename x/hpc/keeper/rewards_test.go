package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/hpc/types"
)

// VE-504: Rewards distribution for HPC contributors tests

const (
	// Fixed-point scale for deterministic math
	RewardScale = int64(1000000)
)

// TestFixedPointRewardCalculation tests deterministic reward calculations
func TestFixedPointRewardCalculation(t *testing.T) {
	testCases := []struct {
		name           string
		totalAmount    int64
		rate           int64 // Fixed-point rate (scale 1,000,000)
		expectedAmount int64
	}{
		{
			name:           "5% platform fee",
			totalAmount:    1000000, // 1 token
			rate:           50000,   // 5%
			expectedAmount: 50000,
		},
		{
			name:           "70% provider reward",
			totalAmount:    1000000,
			rate:           700000, // 70%
			expectedAmount: 700000,
		},
		{
			name:           "25% node reward",
			totalAmount:    1000000,
			rate:           250000, // 25%
			expectedAmount: 250000,
		},
		{
			name:           "large amount precision",
			totalAmount:    1000000000000, // 1 million tokens
			rate:           333333,        // 33.3333%
			expectedAmount: 333333000000,
		},
		{
			name:           "small amount precision",
			totalAmount:    100,
			rate:           500000, // 50%
			expectedAmount: 50,
		},
		{
			name:           "zero rate",
			totalAmount:    1000000,
			rate:           0,
			expectedAmount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateFixedPointShare(tc.totalAmount, tc.rate)
			require.Equal(t, tc.expectedAmount, result,
				"expected %d, got %d", tc.expectedAmount, result)
		})
	}
}

// calculateFixedPointShare mirrors keeper's calculation for testing
func calculateFixedPointShare(total, rate int64) int64 {
	return (total * rate) / RewardScale
}

// TestRewardDistributionSumsCorrectly tests that all rewards sum to total
func TestRewardDistributionSumsCorrectly(t *testing.T) {
	testCases := []struct {
		name         string
		totalAmount  int64
		platformRate int64
		providerRate int64
		nodeRate     int64
	}{
		{
			name:         "standard distribution",
			totalAmount:  1000000,
			platformRate: 50000,  // 5%
			providerRate: 700000, // 70%
			nodeRate:     250000, // 25%
		},
		{
			name:         "equal distribution",
			totalAmount:  1000000,
			platformRate: 333333,
			providerRate: 333333,
			nodeRate:     333334,
		},
		{
			name:         "large amount",
			totalAmount:  1000000000000,
			platformRate: 50000,
			providerRate: 700000,
			nodeRate:     250000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			platformShare := calculateFixedPointShare(tc.totalAmount, tc.platformRate)
			providerShare := calculateFixedPointShare(tc.totalAmount, tc.providerRate)
			nodeShare := calculateFixedPointShare(tc.totalAmount, tc.nodeRate)

			totalDistributed := platformShare + providerShare + nodeShare

			// Allow for small rounding differences (< 3 units)
			diff := tc.totalAmount - totalDistributed
			require.LessOrEqual(t, diff, int64(3),
				"distribution should sum to total (diff: %d)", diff)
			require.GreaterOrEqual(t, diff, int64(0),
				"distribution should not exceed total")
		})
	}
}

// TestNodeRewardProportionalDistribution tests fair distribution among nodes
func TestNodeRewardProportionalDistribution(t *testing.T) {
	testCases := []struct {
		name           string
		totalNodePool  int64
		nodeContribs   []int64 // CPU-seconds contributed
		expectedShares []int64
	}{
		{
			name:           "equal contribution",
			totalNodePool:  1000000,
			nodeContribs:   []int64{100, 100, 100, 100},
			expectedShares: []int64{250000, 250000, 250000, 250000},
		},
		{
			name:           "proportional contribution",
			totalNodePool:  1000000,
			nodeContribs:   []int64{500, 300, 200},
			expectedShares: []int64{500000, 300000, 200000},
		},
		{
			name:           "single node",
			totalNodePool:  1000000,
			nodeContribs:   []int64{100},
			expectedShares: []int64{1000000},
		},
		{
			name:           "uneven distribution",
			totalNodePool:  1000000,
			nodeContribs:   []int64{70, 20, 10},
			expectedShares: []int64{700000, 200000, 100000},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Calculate total contribution
			var totalContrib int64
			for _, c := range tc.nodeContribs {
				totalContrib += c
			}

			// Calculate each node's share
			shares := make([]int64, len(tc.nodeContribs))
			for i, contrib := range tc.nodeContribs {
				// nodeShare = (contrib / totalContrib) * totalNodePool
				// Using fixed-point: (contrib * scale / totalContrib) * totalNodePool / scale
				rate := (contrib * RewardScale) / totalContrib
				shares[i] = calculateFixedPointShare(tc.totalNodePool, rate)
			}

			for i, expected := range tc.expectedShares {
				// Allow 1 unit tolerance for rounding
				diff := shares[i] - expected
				if diff < 0 {
					diff = -diff
				}
				require.LessOrEqual(t, diff, int64(1),
					"node %d: expected %d, got %d", i, expected, shares[i])
			}
		})
	}
}

// TestRewardRecordValidation tests reward record validation
func TestRewardRecordValidation(t *testing.T) {
	validAddr := sdk.AccAddress(make([]byte, 20)).String()
	testCoins := sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1000000)))
	platformCoins := sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(50000)))
	providerCoins := sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(700000)))
	nodeCoins := sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(250000)))

	testCases := []struct {
		name        string
		record      types.HPCRewardRecord
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid record",
			record: types.HPCRewardRecord{
				RewardID:    "reward-001",
				JobID:       "job-001",
				ClusterID:   "cluster-001",
				Source:      types.HPCRewardSourceJobCompletion,
				TotalReward: testCoins,
				Recipients: []types.HPCRewardRecipient{
					{
						Address:       validAddr,
						Amount:        platformCoins,
						RecipientType: "platform",
					},
					{
						Address:       validAddr,
						Amount:        providerCoins,
						RecipientType: "provider",
					},
					{
						Address:       validAddr,
						Amount:        nodeCoins,
						RecipientType: "node",
					},
				},
				JobCompletionStatus: types.JobStateCompleted,
				IssuedAt:            time.Now(),
				BlockHeight:         100,
			},
			expectError: false,
		},
		{
			name: "missing job ID",
			record: types.HPCRewardRecord{
				RewardID:    "reward-001",
				JobID:       "",
				ClusterID:   "cluster-001",
				Source:      types.HPCRewardSourceJobCompletion,
				TotalReward: testCoins,
			},
			expectError: true,
			errorMsg:    "job_id",
		},
		{
			name: "missing cluster ID",
			record: types.HPCRewardRecord{
				RewardID:    "reward-001",
				JobID:       "job-001",
				ClusterID:   "",
				Source:      types.HPCRewardSourceJobCompletion,
				TotalReward: testCoins,
			},
			expectError: true,
			errorMsg:    "cluster_id",
		},
		{
			name: "empty recipients",
			record: types.HPCRewardRecord{
				RewardID:    "reward-001",
				JobID:       "job-001",
				ClusterID:   "cluster-001",
				Source:      types.HPCRewardSourceJobCompletion,
				TotalReward: testCoins,
				Recipients:  []types.HPCRewardRecipient{},
			},
			expectError: true,
			errorMsg:    "recipients",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.record.Validate()
			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestDisputeValidation tests HPC dispute validation
func TestDisputeValidation(t *testing.T) {
	validAddr := sdk.AccAddress(make([]byte, 20)).String()

	testCases := []struct {
		name        string
		dispute     types.HPCDispute
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid dispute",
			dispute: types.HPCDispute{
				DisputeID:       "dispute-001",
				JobID:           "job-001",
				DisputerAddress: validAddr,
				DisputeType:     "job_failure",
				Reason:          "job terminated early without completion",
				Evidence:        "logs showing premature termination",
				Status:          types.DisputeStatusPending,
				CreatedAt:       time.Now(),
			},
			expectError: false,
		},
		{
			name: "missing disputer",
			dispute: types.HPCDispute{
				DisputeID:       "dispute-001",
				JobID:           "job-001",
				DisputerAddress: "",
				DisputeType:     "job_failure",
				Reason:          "test",
			},
			expectError: true,
			errorMsg:    "disputer",
		},
		{
			name: "missing reason",
			dispute: types.HPCDispute{
				DisputeID:       "dispute-001",
				JobID:           "job-001",
				DisputerAddress: validAddr,
				DisputeType:     "job_failure",
				Reason:          "",
			},
			expectError: true,
			errorMsg:    "reason",
		},
		{
			name: "invalid status",
			dispute: types.HPCDispute{
				DisputeID:       "dispute-001",
				JobID:           "job-001",
				DisputerAddress: validAddr,
				DisputeType:     "job_failure",
				Reason:          "test",
				Status:          types.DisputeStatus("invalid"),
			},
			expectError: true,
			errorMsg:    "status",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.dispute.Validate()
			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestPartialJobRewardCalculation tests rewards for partial job completion
func TestPartialJobRewardCalculation(t *testing.T) {
	testCases := []struct {
		name              string
		totalBudget       int64
		completionPercent int64 // Fixed-point percentage
		expectedPayout    int64
	}{
		{
			name:              "full completion",
			totalBudget:       1000000,
			completionPercent: 1000000, // 100%
			expectedPayout:    1000000,
		},
		{
			name:              "half completion",
			totalBudget:       1000000,
			completionPercent: 500000, // 50%
			expectedPayout:    500000,
		},
		{
			name:              "early termination",
			totalBudget:       1000000,
			completionPercent: 250000, // 25%
			expectedPayout:    250000,
		},
		{
			name:              "minimal work",
			totalBudget:       1000000,
			completionPercent: 10000, // 1%
			expectedPayout:    10000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			payout := calculateFixedPointShare(tc.totalBudget, tc.completionPercent)
			require.Equal(t, tc.expectedPayout, payout)
		})
	}
}

// TestRewardIdempotency tests that reward distribution is idempotent
func TestRewardIdempotency(t *testing.T) {
	// Same inputs should always produce same outputs
	totalAmount := int64(1000000)
	platformRate := int64(50000)
	providerRate := int64(700000)
	nodeRate := int64(250000)

	// Calculate rewards multiple times
	results := make([][3]int64, 10)
	for i := 0; i < 10; i++ {
		results[i][0] = calculateFixedPointShare(totalAmount, platformRate)
		results[i][1] = calculateFixedPointShare(totalAmount, providerRate)
		results[i][2] = calculateFixedPointShare(totalAmount, nodeRate)
	}

	// All results should be identical
	for i := 1; i < 10; i++ {
		require.Equal(t, results[0], results[i],
			"reward calculation should be deterministic")
	}
}

// TestJobAccountingMetrics tests job accounting metrics validation
func TestJobAccountingMetrics(t *testing.T) {
	validAddr := sdk.AccAddress(make([]byte, 20)).String()
	testCoins := sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(100000)))

	testCases := []struct {
		name        string
		accounting  types.JobAccounting
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid accounting",
			accounting: types.JobAccounting{
				JobID:           "job-001",
				ClusterID:       "cluster-001",
				ProviderAddress: validAddr,
				CustomerAddress: validAddr,
				UsageMetrics: types.HPCUsageMetrics{
					CPUCoreSeconds:   3600,
					MemoryGBSeconds:  3600 * 1024,
					GPUSeconds:       0,
					WallClockSeconds: 1000,
				},
				TotalCost:           testCoins,
				JobCompletionStatus: types.JobStateCompleted,
			},
			expectError: false,
		},
		{
			name: "missing job ID",
			accounting: types.JobAccounting{
				JobID:           "",
				ClusterID:       "cluster-001",
				ProviderAddress: validAddr,
				CustomerAddress: validAddr,
			},
			expectError: true,
			errorMsg:    "job_id",
		},
		{
			name: "missing cluster ID",
			accounting: types.JobAccounting{
				JobID:           "job-001",
				ClusterID:       "",
				ProviderAddress: validAddr,
				CustomerAddress: validAddr,
			},
			expectError: true,
			errorMsg:    "cluster_id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.accounting.Validate()
			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
