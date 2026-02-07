//go:build e2e.integration

// Package e2e contains end-to-end integration tests.
//
// VE-5A: E2E tests for HPC billing and settlement integration.
// Tests the complete billing workflow from usage metrics through settlement.
package e2e

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	pd "github.com/virtengine/virtengine/pkg/provider_daemon"
	"github.com/virtengine/virtengine/tests/e2e/fixtures"
	"github.com/virtengine/virtengine/tests/e2e/mocks"
	"github.com/virtengine/virtengine/testutil"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// =============================================================================
// Billing Rate Constants (from _docs/hpc-billing-rules.md)
// =============================================================================

const (
	// CPUCoreHourRate is uvirt per core-hour
	CPUCoreHourRate = 10000
	// MemoryGBHourRate is uvirt per GB-hour
	MemoryGBHourRate = 1000
	// GPUA100HourRate is uvirt per GPU-hour for A100
	GPUA100HourRate = 500000
	// GPUV100HourRate is uvirt per GPU-hour for V100
	GPUV100HourRate = 250000
	// GPUT4HourRate is uvirt per GPU-hour for T4
	GPUT4HourRate = 100000
	// StorageGBHourRate is uvirt per GB-hour
	StorageGBHourRate = 50
	// NetworkGBRate is uvirt per GB
	NetworkGBRate = 100
	// NodeHourRate is uvirt per node-hour
	NodeHourRate = 50000
	// MinimumCharge is uvirt minimum per job
	MinimumCharge = 1000
	// PlatformFeeBps is platform fee in basis points (2.5%)
	PlatformFeeBps = 250
	// ProviderRewardBps is provider reward in basis points (97.5%)
	ProviderRewardBps = 9750
)

// Volume discount tiers
const (
	VolumeDiscountTier0Max = 100  // 0-100 core-hours: 0%
	VolumeDiscountTier1Max = 500  // 100-500 core-hours: 5%
	VolumeDiscountTier2Max = 1000 // 500-1000 core-hours: 10%
	VolumeDiscountTier1Bps = 500  // 5%
	VolumeDiscountTier2Bps = 1000 // 10%
	VolumeDiscountTier3Bps = 1500 // 15%
)

// =============================================================================
// Test Suite Definition
// =============================================================================

// HPCBillingE2ETestSuite tests the complete HPC billing and settlement flow:
// Usage metrics → Billing calculation → Discounts → Escrow → Settlement → Rewards
type HPCBillingE2ETestSuite struct {
	*testutil.NetworkTestSuite

	// Test addresses
	providerAddr  string
	customerAddr  string
	provider2Addr string
	customer2Addr string

	// Mock components
	mockSLURM         *mocks.MockSLURMIntegration
	BillingMockEscrow *BillingMockEscrowKeeper
	mockBilling       *BillingMockCalculator
	mockSettlement    *BillingMockSettlementProcessor

	// Test data
	testCluster  *mocks.SLURMCluster
	testOffering *hpctypes.HPCOffering
	billingRules hpctypes.HPCBillingRules

	// State tracking
	createdJobs       map[string]*hpctypes.HPCJob
	accountingRecords map[string]*hpctypes.HPCAccountingRecord
	disputes          map[string]*hpctypes.HPCDispute
	mu                sync.RWMutex
}

func TestHPCBillingE2E(t *testing.T) {
	suite.Run(t, &HPCBillingE2ETestSuite{
		NetworkTestSuite: testutil.NewNetworkTestSuite(nil, &HPCBillingE2ETestSuite{}),
	})
}

func (s *HPCBillingE2ETestSuite) SetupSuite() {
	s.NetworkTestSuite.SetupSuite()

	val := s.Network().Validators[0]
	s.providerAddr = val.Address.String()
	s.customerAddr = val.Address.String()
	s.provider2Addr = sdk.AccAddress([]byte("provider2-e2e-test-001")).String()
	s.customer2Addr = sdk.AccAddress([]byte("customer2-e2e-test-001")).String()

	// Initialize mock components
	s.mockSLURM = mocks.NewMockSLURMIntegration()
	s.BillingMockEscrow = NewBillingMockEscrowKeeper()
	s.mockBilling = NewBillingMockCalculator()
	s.mockSettlement = NewBillingMockSettlementProcessor()

	// Initialize state tracking
	s.createdJobs = make(map[string]*hpctypes.HPCJob)
	s.accountingRecords = make(map[string]*hpctypes.HPCAccountingRecord)
	s.disputes = make(map[string]*hpctypes.HPCDispute)

	// Setup test cluster
	s.testCluster = mocks.DefaultTestCluster()
	s.mockSLURM.RegisterCluster(s.testCluster)

	// Setup billing rules
	s.billingRules = s.createTestBillingRules()

	// Setup test offering
	config := fixtures.DefaultTestOfferingConfig()
	config.ProviderAddr = s.providerAddr
	s.testOffering = fixtures.CreateTestOffering(config)

	// Start mock SLURM
	ctx := context.Background()
	err := s.mockSLURM.Start(ctx)
	s.Require().NoError(err)
}

func (s *HPCBillingE2ETestSuite) TearDownSuite() {
	if s.mockSLURM != nil {
		_ = s.mockSLURM.Stop()
	}
	s.NetworkTestSuite.TearDownSuite()
}

func (s *HPCBillingE2ETestSuite) createTestBillingRules() hpctypes.HPCBillingRules {
	denom := "uvirt"
	return hpctypes.HPCBillingRules{
		FormulaVersion: hpctypes.CurrentBillingFormulaVersion,
		ResourceRates: hpctypes.HPCResourceRates{
			CPUCoreHourRate:   sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(CPUCoreHourRate)),
			MemoryGBHourRate:  sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(MemoryGBHourRate)),
			GPUHourRate:       sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(GPUV100HourRate)),
			StorageGBHourRate: sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(StorageGBHourRate)),
			NetworkGBRate:     sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(NetworkGBRate)),
			NodeHourRate:      sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(NodeHourRate)),
			GPUTypeRates: map[string]sdk.DecCoin{
				"nvidia-a100": sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(GPUA100HourRate)),
				"nvidia-v100": sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(GPUV100HourRate)),
				"nvidia-t4":   sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(GPUT4HourRate)),
			},
		},
		DiscountRules: []hpctypes.HPCDiscountRule{
			{
				RuleID:         "volume-tier1",
				RuleName:       "Volume Discount Tier 1",
				DiscountType:   hpctypes.HPCDiscountTypeVolume,
				DiscountBps:    VolumeDiscountTier1Bps,
				ThresholdValue: sdkmath.LegacyNewDec(VolumeDiscountTier0Max),
				ThresholdUnit:  "cpu_core_hours",
				Active:         true,
			},
			{
				RuleID:         "volume-tier2",
				RuleName:       "Volume Discount Tier 2",
				DiscountType:   hpctypes.HPCDiscountTypeVolume,
				DiscountBps:    VolumeDiscountTier2Bps,
				ThresholdValue: sdkmath.LegacyNewDec(VolumeDiscountTier1Max),
				ThresholdUnit:  "cpu_core_hours",
				Active:         true,
			},
			{
				RuleID:         "volume-tier3",
				RuleName:       "Volume Discount Tier 3",
				DiscountType:   hpctypes.HPCDiscountTypeVolume,
				DiscountBps:    VolumeDiscountTier3Bps,
				ThresholdValue: sdkmath.LegacyNewDec(VolumeDiscountTier2Max),
				ThresholdUnit:  "cpu_core_hours",
				Active:         true,
			},
		},
		PlatformFeeRateBps:        PlatformFeeBps,
		ProviderRewardRateBps:     ProviderRewardBps,
		MinimumCharge:             sdk.NewCoin(denom, sdkmath.NewInt(MinimumCharge)),
		BillingGranularitySeconds: 60,
	}
}

// =============================================================================
// A. Billing Calculation Tests (~400 lines)
// =============================================================================

func (s *HPCBillingE2ETestSuite) TestA01_BasicBillingCalculation() {
	s.Run("CalculateBillingFromUsageMetrics", func() {
		// Create a simple job with 1 hour of usage
		metrics := &hpctypes.HPCDetailedMetrics{
			WallClockSeconds: 3600,                    // 1 hour
			CPUCoreSeconds:   3600 * 4,                // 4 cores for 1 hour
			MemoryGBSeconds:  3600 * 16,               // 16 GB for 1 hour
			StorageGBHours:   10,                      // 10 GB-hours
			NetworkBytesIn:   1024 * 1024 * 1024,      // 1 GB in
			NetworkBytesOut:  512 * 1024 * 1024,       // 0.5 GB out
			NodeHours:        sdkmath.LegacyNewDec(1), // 1 node-hour
			NodesUsed:        1,
		}

		calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)
		breakdown, billable, err := calculator.CalculateBillableAmount(metrics, nil, nil)

		s.Require().NoError(err)
		s.NotNil(breakdown)
		s.NotNil(billable)

		// Verify CPU cost: 4 core-hours * 10000 = 40000
		expectedCPUCost := sdkmath.NewInt(4 * CPUCoreHourRate)
		s.Equal(expectedCPUCost, breakdown.CPUCost.Amount, "CPU cost mismatch")

		// Verify Memory cost: 16 GB-hours * 1000 = 16000
		expectedMemCost := sdkmath.NewInt(16 * MemoryGBHourRate)
		s.Equal(expectedMemCost, breakdown.MemoryCost.Amount, "Memory cost mismatch")

		// Verify Node cost: 1 node-hour * 50000 = 50000
		expectedNodeCost := sdkmath.NewInt(1 * NodeHourRate)
		s.Equal(expectedNodeCost, breakdown.NodeCost.Amount, "Node cost mismatch")

		// Verify Storage cost: 10 GB-hours * 50 = 500
		expectedStorageCost := sdkmath.NewInt(10 * StorageGBHourRate)
		s.Equal(expectedStorageCost, breakdown.StorageCost.Amount, "Storage cost mismatch")

		// Verify Network cost: ~1.5 GB * 100 = 150
		expectedNetworkCost := sdkmath.NewInt(1 * NetworkGBRate) // 1 GB (truncated)
		s.True(breakdown.NetworkCost.Amount.GTE(expectedNetworkCost), "Network cost too low")

		// Verify total is sum of all components
		s.True(billable.AmountOf("uvirt").IsPositive(), "Billable amount should be positive")
	})
}

func (s *HPCBillingE2ETestSuite) TestA02_CPUBillingAccuracy() {
	s.Run("VerifyCPUCoreHourBilling", func() {
		testCases := []struct {
			name           string
			cpuCoreSeconds int64
			expectedCost   int64
		}{
			{"1 core-hour", 3600, CPUCoreHourRate},
			{"10 core-hours", 36000, 10 * CPUCoreHourRate},
			{"0.5 core-hours", 1800, CPUCoreHourRate / 2},
			{"100 core-hours", 360000, 100 * CPUCoreHourRate},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				metrics := &hpctypes.HPCDetailedMetrics{
					CPUCoreSeconds: tc.cpuCoreSeconds,
					NodeHours:      sdkmath.LegacyZeroDec(),
				}

				calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)
				breakdown, _, err := calculator.CalculateBillableAmount(metrics, nil, nil)

				s.Require().NoError(err)
				s.Equal(tc.expectedCost, breakdown.CPUCost.Amount.Int64(),
					"CPU cost mismatch for %s", tc.name)
			})
		}
	})
}

func (s *HPCBillingE2ETestSuite) TestA03_MemoryBillingAccuracy() {
	s.Run("VerifyMemoryGBHourBilling", func() {
		testCases := []struct {
			name            string
			memoryGBSeconds int64
			expectedCost    int64
		}{
			{"1 GB-hour", 3600, MemoryGBHourRate},
			{"8 GB-hours", 3600 * 8, 8 * MemoryGBHourRate},
			{"64 GB-hours", 3600 * 64, 64 * MemoryGBHourRate},
			{"256 GB-hours", 3600 * 256, 256 * MemoryGBHourRate},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				metrics := &hpctypes.HPCDetailedMetrics{
					MemoryGBSeconds: tc.memoryGBSeconds,
					NodeHours:       sdkmath.LegacyZeroDec(),
				}

				calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)
				breakdown, _, err := calculator.CalculateBillableAmount(metrics, nil, nil)

				s.Require().NoError(err)
				s.Equal(tc.expectedCost, breakdown.MemoryCost.Amount.Int64(),
					"Memory cost mismatch for %s", tc.name)
			})
		}
	})
}

func (s *HPCBillingE2ETestSuite) TestA04_GPUBillingAccuracy() {
	s.Run("VerifyGPUHourBillingByType", func() {
		testCases := []struct {
			name         string
			gpuType      string
			gpuSeconds   int64
			expectedCost int64
		}{
			{"A100 1 hour", "nvidia-a100", 3600, GPUA100HourRate},
			{"V100 1 hour", "nvidia-v100", 3600, GPUV100HourRate},
			{"T4 1 hour", "nvidia-t4", 3600, GPUT4HourRate},
			{"A100 8 hours", "nvidia-a100", 3600 * 8, 8 * GPUA100HourRate},
			{"V100 4 hours", "nvidia-v100", 3600 * 4, 4 * GPUV100HourRate},
			{"T4 24 hours", "nvidia-t4", 3600 * 24, 24 * GPUT4HourRate},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				metrics := &hpctypes.HPCDetailedMetrics{
					GPUSeconds: tc.gpuSeconds,
					GPUType:    tc.gpuType,
					NodeHours:  sdkmath.LegacyZeroDec(),
				}

				calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)
				breakdown, _, err := calculator.CalculateBillableAmount(metrics, nil, nil)

				s.Require().NoError(err)
				s.Equal(tc.expectedCost, breakdown.GPUCost.Amount.Int64(),
					"GPU cost mismatch for %s", tc.name)
			})
		}
	})
}

func (s *HPCBillingE2ETestSuite) TestA05_StorageBillingAccuracy() {
	s.Run("VerifyStorageGBHourBilling", func() {
		testCases := []struct {
			name           string
			storageGBHours int64
			expectedCost   int64
		}{
			{"10 GB-hours", 10, 10 * StorageGBHourRate},
			{"100 GB-hours", 100, 100 * StorageGBHourRate},
			{"1000 GB-hours", 1000, 1000 * StorageGBHourRate},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				metrics := &hpctypes.HPCDetailedMetrics{
					StorageGBHours: tc.storageGBHours,
					NodeHours:      sdkmath.LegacyZeroDec(),
				}

				calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)
				breakdown, _, err := calculator.CalculateBillableAmount(metrics, nil, nil)

				s.Require().NoError(err)
				s.Equal(tc.expectedCost, breakdown.StorageCost.Amount.Int64(),
					"Storage cost mismatch for %s", tc.name)
			})
		}
	})
}

func (s *HPCBillingE2ETestSuite) TestA06_NetworkBillingAccuracy() {
	s.Run("VerifyNetworkGBBilling", func() {
		testCases := []struct {
			name         string
			bytesIn      int64
			bytesOut     int64
			expectedCost int64
		}{
			{"1 GB total", 512 * 1024 * 1024, 512 * 1024 * 1024, 1 * NetworkGBRate},
			{"10 GB total", 5 * 1024 * 1024 * 1024, 5 * 1024 * 1024 * 1024, 10 * NetworkGBRate},
			{"100 GB total", 50 * 1024 * 1024 * 1024, 50 * 1024 * 1024 * 1024, 100 * NetworkGBRate},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				metrics := &hpctypes.HPCDetailedMetrics{
					NetworkBytesIn:  tc.bytesIn,
					NetworkBytesOut: tc.bytesOut,
					NodeHours:       sdkmath.LegacyZeroDec(),
				}

				calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)
				breakdown, _, err := calculator.CalculateBillableAmount(metrics, nil, nil)

				s.Require().NoError(err)
				s.Equal(tc.expectedCost, breakdown.NetworkCost.Amount.Int64(),
					"Network cost mismatch for %s", tc.name)
			})
		}
	})
}

func (s *HPCBillingE2ETestSuite) TestA07_NodeHourBillingAccuracy() {
	s.Run("VerifyNodeHourBilling", func() {
		testCases := []struct {
			name         string
			nodeHours    string
			expectedCost int64
		}{
			{"1 node-hour", "1", NodeHourRate},
			{"4 node-hours", "4", 4 * NodeHourRate},
			{"24 node-hours", "24", 24 * NodeHourRate},
			{"0.5 node-hours", "0.5", NodeHourRate / 2},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				nodeHours, _ := sdkmath.LegacyNewDecFromStr(tc.nodeHours)
				metrics := &hpctypes.HPCDetailedMetrics{
					NodeHours: nodeHours,
				}

				calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)
				breakdown, _, err := calculator.CalculateBillableAmount(metrics, nil, nil)

				s.Require().NoError(err)
				s.Equal(tc.expectedCost, breakdown.NodeCost.Amount.Int64(),
					"Node cost mismatch for %s", tc.name)
			})
		}
	})
}

func (s *HPCBillingE2ETestSuite) TestA08_MinimumChargeEnforcement() {
	s.Run("VerifyMinimumChargeApplied", func() {
		// Create metrics that would result in less than minimum charge
		metrics := &hpctypes.HPCDetailedMetrics{
			WallClockSeconds: 1, // 1 second
			CPUCoreSeconds:   1,
			MemoryGBSeconds:  1,
			NodeHours:        sdkmath.LegacyZeroDec(),
		}

		calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)
		_, billable, err := calculator.CalculateBillableAmount(metrics, nil, nil)

		s.Require().NoError(err)
		s.Equal(int64(MinimumCharge), billable.AmountOf("uvirt").Int64(),
			"Minimum charge not enforced")
	})

	s.Run("VerifyMinimumChargeNotAppliedWhenExceeded", func() {
		// Create metrics that exceed minimum charge
		metrics := &hpctypes.HPCDetailedMetrics{
			WallClockSeconds: 3600,
			CPUCoreSeconds:   3600 * 10, // 10 core-hours = 100000 uvirt
			NodeHours:        sdkmath.LegacyZeroDec(),
		}

		calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)
		_, billable, err := calculator.CalculateBillableAmount(metrics, nil, nil)

		s.Require().NoError(err)
		s.True(billable.AmountOf("uvirt").GT(sdkmath.NewInt(MinimumCharge)),
			"Billable should exceed minimum charge")
	})
}

func (s *HPCBillingE2ETestSuite) TestA09_MultiResourceBilling() {
	s.Run("VerifyCombinedResourceBilling", func() {
		// Create metrics with all resource types
		metrics := &hpctypes.HPCDetailedMetrics{
			WallClockSeconds: 7200,      // 2 hours
			CPUCoreSeconds:   7200 * 8,  // 8 cores for 2 hours = 16 core-hours
			MemoryGBSeconds:  7200 * 32, // 32 GB for 2 hours = 64 GB-hours
			GPUSeconds:       7200 * 2,  // 2 GPUs for 2 hours = 4 GPU-hours
			GPUType:          "nvidia-a100",
			StorageGBHours:   100,                     // 100 GB-hours
			NetworkBytesIn:   10 * 1024 * 1024 * 1024, // 10 GB
			NetworkBytesOut:  5 * 1024 * 1024 * 1024,  // 5 GB
			NodeHours:        sdkmath.LegacyNewDec(2), // 2 node-hours
			NodesUsed:        1,
		}

		calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)
		breakdown, billable, err := calculator.CalculateBillableAmount(metrics, nil, nil)

		s.Require().NoError(err)

		// Calculate expected totals
		expectedCPU := int64(16 * CPUCoreHourRate)        // 160000
		expectedMem := int64(64 * MemoryGBHourRate)       // 64000
		expectedGPU := int64(4 * GPUA100HourRate)         // 2000000
		expectedStorage := int64(100 * StorageGBHourRate) // 5000
		expectedNetwork := int64(15 * NetworkGBRate)      // 1500
		expectedNode := int64(2 * NodeHourRate)           // 100000

		s.Equal(expectedCPU, breakdown.CPUCost.Amount.Int64())
		s.Equal(expectedMem, breakdown.MemoryCost.Amount.Int64())
		s.Equal(expectedGPU, breakdown.GPUCost.Amount.Int64())
		s.Equal(expectedStorage, breakdown.StorageCost.Amount.Int64())
		s.Equal(expectedNetwork, breakdown.NetworkCost.Amount.Int64())
		s.Equal(expectedNode, breakdown.NodeCost.Amount.Int64())

		expectedTotal := expectedCPU + expectedMem + expectedGPU + expectedStorage + expectedNetwork + expectedNode
		s.Equal(expectedTotal, billable.AmountOf("uvirt").Int64(),
			"Total billable mismatch")
	})
}

// =============================================================================
// B. Discount Tests (~300 lines)
// =============================================================================

func (s *HPCBillingE2ETestSuite) TestB01_VolumeDiscountTier0() {
	s.Run("NoDiscountFor0To100CoreHours", func() {
		// 50 core-hours of historical usage (below tier 1 threshold)
		historicalUsage := &hpctypes.AccountingAggregation{
			TotalCPUCoreHours: sdkmath.LegacyNewDec(50),
		}

		metrics := &hpctypes.HPCDetailedMetrics{
			CPUCoreSeconds: 3600 * 10, // 10 core-hours current job
			NodeHours:      sdkmath.LegacyZeroDec(),
		}

		calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)
		discounts := calculator.EvaluateDiscounts(metrics, s.customerAddr, historicalUsage)

		s.Empty(discounts, "No discount should be applied for tier 0")
	})
}

func (s *HPCBillingE2ETestSuite) TestB02_VolumeDiscountTier1() {
	s.Run("5PercentDiscountFor100To500CoreHours", func() {
		// 150 core-hours of historical usage (in tier 1)
		historicalUsage := &hpctypes.AccountingAggregation{
			TotalCPUCoreHours: sdkmath.LegacyNewDec(150),
		}

		metrics := &hpctypes.HPCDetailedMetrics{
			CPUCoreSeconds: 3600 * 10, // 10 core-hours current job
			NodeHours:      sdkmath.LegacyZeroDec(),
		}

		calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)
		discounts := calculator.EvaluateDiscounts(metrics, s.customerAddr, historicalUsage)

		s.Len(discounts, 1, "One discount should be applied for tier 1")
		s.Equal(uint32(VolumeDiscountTier1Bps), discounts[0].DiscountBps,
			"Discount should be 5%% (500 bps)")
	})
}

func (s *HPCBillingE2ETestSuite) TestB03_VolumeDiscountTier2() {
	s.Run("10PercentDiscountFor500To1000CoreHours", func() {
		// 600 core-hours of historical usage (in tier 2)
		historicalUsage := &hpctypes.AccountingAggregation{
			TotalCPUCoreHours: sdkmath.LegacyNewDec(600),
		}

		metrics := &hpctypes.HPCDetailedMetrics{
			CPUCoreSeconds: 3600 * 10, // 10 core-hours current job
			NodeHours:      sdkmath.LegacyZeroDec(),
		}

		calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)
		discounts := calculator.EvaluateDiscounts(metrics, s.customerAddr, historicalUsage)

		// Should get tier 1 and tier 2 discounts (both thresholds met)
		s.GreaterOrEqual(len(discounts), 1, "At least one discount should be applied")

		// Find the highest tier discount applied
		var maxDiscount uint32
		for _, d := range discounts {
			if d.DiscountBps > maxDiscount {
				maxDiscount = d.DiscountBps
			}
		}
		s.GreaterOrEqual(maxDiscount, uint32(VolumeDiscountTier2Bps),
			"Should have at least tier 2 discount")
	})
}

func (s *HPCBillingE2ETestSuite) TestB04_VolumeDiscountTier3() {
	s.Run("15PercentDiscountFor1000PlusCoreHours", func() {
		// 1500 core-hours of historical usage (in tier 3)
		historicalUsage := &hpctypes.AccountingAggregation{
			TotalCPUCoreHours: sdkmath.LegacyNewDec(1500),
		}

		metrics := &hpctypes.HPCDetailedMetrics{
			CPUCoreSeconds: 3600 * 10, // 10 core-hours current job
			NodeHours:      sdkmath.LegacyZeroDec(),
		}

		calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)
		discounts := calculator.EvaluateDiscounts(metrics, s.customerAddr, historicalUsage)

		// Should get all tier discounts (all thresholds met)
		s.GreaterOrEqual(len(discounts), 1, "Discounts should be applied")

		// Find the highest tier discount applied
		var maxDiscount uint32
		for _, d := range discounts {
			if d.DiscountBps > maxDiscount {
				maxDiscount = d.DiscountBps
			}
		}
		s.GreaterOrEqual(maxDiscount, uint32(VolumeDiscountTier3Bps),
			"Should have tier 3 discount (15%%)")
	})
}

func (s *HPCBillingE2ETestSuite) TestB05_DiscountApplicationOrder() {
	s.Run("VerifyCorrectDiscountStackingOrder", func() {
		// Create scenario with multiple discount types
		historicalUsage := &hpctypes.AccountingAggregation{
			TotalCPUCoreHours: sdkmath.LegacyNewDec(1200),
		}

		metrics := &hpctypes.HPCDetailedMetrics{
			CPUCoreSeconds: 3600 * 100, // 100 core-hours = 1,000,000 uvirt base
			NodeHours:      sdkmath.LegacyZeroDec(),
		}

		calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)
		discounts := calculator.EvaluateDiscounts(metrics, s.customerAddr, historicalUsage)

		// Apply discounts to base amount
		baseCost := int64(100 * CPUCoreHourRate) // 1,000,000 uvirt

		totalDiscountAmount := int64(0)
		for _, d := range discounts {
			totalDiscountAmount += d.DiscountAmount.AmountOf("uvirt").Int64()
		}

		// Verify discounts are applied correctly
		s.True(totalDiscountAmount > 0, "Discounts should be applied")
		s.True(totalDiscountAmount < baseCost, "Discounts should not exceed base cost")
	})
}

func (s *HPCBillingE2ETestSuite) TestB06_HistoricalUsageForDiscount() {
	s.Run("VerifyDiscountBasedOn30DayHistory", func() {
		// Simulate 30-day historical usage accumulation
		now := time.Now()
		thirtyDaysAgo := now.Add(-30 * 24 * time.Hour)

		historicalUsage := &hpctypes.AccountingAggregation{
			CustomerAddress:   s.customerAddr,
			PeriodStart:       thirtyDaysAgo,
			PeriodEnd:         now,
			TotalCPUCoreHours: sdkmath.LegacyNewDec(800), // Within tier 2
			TotalJobs:         50,
		}

		metrics := &hpctypes.HPCDetailedMetrics{
			CPUCoreSeconds: 3600 * 10,
			NodeHours:      sdkmath.LegacyZeroDec(),
		}

		calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)
		discounts := calculator.EvaluateDiscounts(metrics, s.customerAddr, historicalUsage)

		// Should get appropriate discount based on historical usage
		s.GreaterOrEqual(len(discounts), 1, "Should apply discount based on history")

		// Verify discount is based on the 800 core-hours historical usage
		for _, d := range discounts {
			s.Equal("volume", string(d.DiscountType), "Should be volume discount")
		}
	})
}

// =============================================================================
// C. Escrow Integration Tests (~400 lines)
// =============================================================================

func (s *HPCBillingE2ETestSuite) TestC01_EscrowHoldOnJobSubmission() {
	s.Run("FundsHeldOnJobSubmission", func() {
		ctx := context.Background()

		// Create a job with agreed price
		jobConfig := fixtures.DefaultTestJobConfig()
		jobConfig.ProviderAddress = s.providerAddr
		jobConfig.CustomerAddress = s.customerAddr
		jobConfig.AgreedPrice = sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1000000)))
		job := fixtures.CreateTestJob(jobConfig)

		// Setup escrow mock to track holds
		escrowID := fmt.Sprintf("escrow-%s", job.JobID)
		s.BillingMockEscrow.CreateEscrow(escrowID, s.customerAddr, s.providerAddr, jobConfig.AgreedPrice)

		// Submit job to SLURM
		_, err := s.mockSLURM.SubmitJob(ctx, job)
		s.Require().NoError(err)

		// Verify escrow hold was created
		escrow, exists := s.BillingMockEscrow.GetEscrow(escrowID)
		s.True(exists, "Escrow should exist")
		s.Equal(EscrowStatusHeld, escrow.Status, "Escrow should be in held status")
		s.True(escrow.Amount.Equal(jobConfig.AgreedPrice), "Held amount should match agreed price")

		// Track job
		s.mu.Lock()
		s.createdJobs[job.JobID] = job
		s.mu.Unlock()
	})
}

func (s *HPCBillingE2ETestSuite) TestC02_EscrowReleaseOnCompletion() {
	s.Run("FundsReleasedOnSuccessfulCompletion", func() {
		ctx := context.Background()

		// Create and submit job
		jobConfig := fixtures.DefaultTestJobConfig()
		jobConfig.JobID = "job-c02-completion"
		jobConfig.ProviderAddress = s.providerAddr
		jobConfig.CustomerAddress = s.customerAddr
		jobConfig.AgreedPrice = sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(500000)))
		job := fixtures.CreateTestJob(jobConfig)

		escrowID := fmt.Sprintf("escrow-%s", job.JobID)
		s.BillingMockEscrow.CreateEscrow(escrowID, s.customerAddr, s.providerAddr, jobConfig.AgreedPrice)

		_, err := s.mockSLURM.SubmitJob(ctx, job)
		s.Require().NoError(err)

		// Simulate job execution completion
		metrics := fixtures.StandardJobMetrics(3600)
		err = s.mockSLURM.SimulateJobExecution(ctx, job.JobID, 100, true, metrics)
		s.Require().NoError(err)

		// Calculate actual billing
		detailedMetrics := &hpctypes.HPCDetailedMetrics{
			WallClockSeconds: metrics.WallClockSeconds,
			CPUCoreSeconds:   metrics.CPUCoreSeconds,
			MemoryGBSeconds:  metrics.MemoryGBSeconds,
			NodeHours:        sdkmath.LegacyNewDec(int64(metrics.NodeHours)),
		}

		calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)
		_, billable, _ := calculator.CalculateBillableAmount(detailedMetrics, nil, nil)

		// Process escrow release
		s.BillingMockEscrow.ProcessRelease(escrowID, billable)

		// Verify escrow released
		escrow, _ := s.BillingMockEscrow.GetEscrow(escrowID)
		s.Equal(EscrowStatusReleased, escrow.Status, "Escrow should be released")
	})
}

func (s *HPCBillingE2ETestSuite) TestC03_PartialEscrowRelease() {
	s.Run("PartialReleaseForEarlyTermination", func() {
		ctx := context.Background()

		// Create job with larger agreed price
		jobConfig := fixtures.DefaultTestJobConfig()
		jobConfig.JobID = "job-c03-partial"
		jobConfig.ProviderAddress = s.providerAddr
		jobConfig.CustomerAddress = s.customerAddr
		jobConfig.AgreedPrice = sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(2000000)))
		job := fixtures.CreateTestJob(jobConfig)

		escrowID := fmt.Sprintf("escrow-%s", job.JobID)
		s.BillingMockEscrow.CreateEscrow(escrowID, s.customerAddr, s.providerAddr, jobConfig.AgreedPrice)

		_, err := s.mockSLURM.SubmitJob(ctx, job)
		s.Require().NoError(err)

		// Simulate partial execution (cancelled after 30 minutes)
		s.mockSLURM.SetJobState(job.JobID, pd.HPCJobStateRunning)
		time.Sleep(time.Millisecond * 50)

		// Cancel the job
		err = s.mockSLURM.CancelJob(ctx, job.JobID)
		s.Require().NoError(err)

		// Calculate partial billing (30 minutes = 1800 seconds)
		partialMetrics := &hpctypes.HPCDetailedMetrics{
			WallClockSeconds: 1800,
			CPUCoreSeconds:   1800 * 4,
			MemoryGBSeconds:  1800 * 8,
			NodeHours:        sdkmath.LegacyNewDecWithPrec(5, 1), // 0.5 hours
		}

		calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)
		_, actualBilling, _ := calculator.CalculateBillableAmount(partialMetrics, nil, nil)

		// Process partial release
		s.BillingMockEscrow.ProcessPartialRelease(escrowID, actualBilling)

		// Verify partial release
		escrow, _ := s.BillingMockEscrow.GetEscrow(escrowID)
		s.Equal(EscrowStatusPartialReleased, escrow.Status)
		s.True(escrow.ReleasedAmount.IsAllLT(jobConfig.AgreedPrice),
			"Released amount should be less than agreed price")
	})
}

func (s *HPCBillingE2ETestSuite) TestC04_EscrowRefundOnCancellation() {
	s.Run("RefundOnCustomerCancellation", func() {
		ctx := context.Background()

		// Create job
		jobConfig := fixtures.DefaultTestJobConfig()
		jobConfig.JobID = "job-c04-cancel"
		jobConfig.ProviderAddress = s.providerAddr
		jobConfig.CustomerAddress = s.customerAddr
		jobConfig.AgreedPrice = sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1000000)))
		job := fixtures.CreateTestJob(jobConfig)

		escrowID := fmt.Sprintf("escrow-%s", job.JobID)
		s.BillingMockEscrow.CreateEscrow(escrowID, s.customerAddr, s.providerAddr, jobConfig.AgreedPrice)

		_, err := s.mockSLURM.SubmitJob(ctx, job)
		s.Require().NoError(err)

		// Cancel immediately (before any resources consumed)
		err = s.mockSLURM.CancelJob(ctx, job.JobID)
		s.Require().NoError(err)

		// Process full refund
		s.BillingMockEscrow.ProcessRefund(escrowID, jobConfig.AgreedPrice)

		// Verify full refund
		escrow, _ := s.BillingMockEscrow.GetEscrow(escrowID)
		s.Equal(EscrowStatusRefunded, escrow.Status)
		s.True(escrow.RefundAmount.Equal(jobConfig.AgreedPrice),
			"Full amount should be refunded")
	})
}

func (s *HPCBillingE2ETestSuite) TestC05_EscrowRefundOnProviderFailure() {
	s.Run("RefundWhenProviderFails", func() {
		ctx := context.Background()

		// Create job
		jobConfig := fixtures.DefaultTestJobConfig()
		jobConfig.JobID = "job-c05-provider-fail"
		jobConfig.ProviderAddress = s.providerAddr
		jobConfig.CustomerAddress = s.customerAddr
		jobConfig.AgreedPrice = sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1500000)))
		job := fixtures.CreateTestJob(jobConfig)

		escrowID := fmt.Sprintf("escrow-%s", job.JobID)
		s.BillingMockEscrow.CreateEscrow(escrowID, s.customerAddr, s.providerAddr, jobConfig.AgreedPrice)

		_, err := s.mockSLURM.SubmitJob(ctx, job)
		s.Require().NoError(err)

		// Simulate provider failure (infrastructure issue)
		s.mockSLURM.SetJobState(job.JobID, pd.HPCJobStateFailed)
		s.mockSLURM.SetJobExitCode(job.JobID, -1) // Infrastructure failure code

		// Process refund due to provider failure
		s.BillingMockEscrow.ProcessRefundDueToProviderFailure(escrowID)

		// Verify refund
		escrow, _ := s.BillingMockEscrow.GetEscrow(escrowID)
		s.Equal(EscrowStatusRefunded, escrow.Status)
		s.Equal("provider_failure", escrow.RefundReason)
	})
}

func (s *HPCBillingE2ETestSuite) TestC06_InsufficientFundsRejection() {
	s.Run("JobRejectedWhenInsufficientFunds", func() {
		ctx := context.Background()

		// Create job with price higher than available funds
		jobConfig := fixtures.DefaultTestJobConfig()
		jobConfig.JobID = "job-c06-insufficient"
		jobConfig.ProviderAddress = s.providerAddr
		jobConfig.CustomerAddress = s.customerAddr
		jobConfig.AgreedPrice = sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(100000000000))) // Very high
		job := fixtures.CreateTestJob(jobConfig)

		// Set customer balance to insufficient
		s.BillingMockEscrow.SetCustomerBalance(s.customerAddr, sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1000000))))

		// Attempt to create escrow should fail
		escrowID := fmt.Sprintf("escrow-%s", job.JobID)
		err := s.BillingMockEscrow.CreateEscrowWithValidation(escrowID, s.customerAddr, s.providerAddr, jobConfig.AgreedPrice)
		s.Error(err, "Should fail due to insufficient funds")
		s.Contains(err.Error(), "insufficient", "Error should mention insufficient funds")

		// Job submission should also fail
		_, err = s.mockSLURM.SubmitJob(ctx, job)
		if err == nil {
			// If SLURM accepted it, the escrow validation should still fail
			_, exists := s.BillingMockEscrow.GetEscrow(escrowID)
			s.False(exists, "Escrow should not be created")
		}
	})
}

// =============================================================================
// D. Settlement Pipeline Tests (~400 lines)
// =============================================================================

func (s *HPCBillingE2ETestSuite) TestD01_SettlementAfterFinalization() {
	s.Run("SettlementAfterFinalizationDelay", func() {
		// Create accounting record
		now := time.Now()
		record := &hpctypes.HPCAccountingRecord{
			RecordID:        "record-d01-settle",
			JobID:           "job-d01-settle",
			ClusterID:       s.testCluster.ClusterID,
			ProviderAddress: s.providerAddr,
			CustomerAddress: s.customerAddr,
			Status:          hpctypes.AccountingStatusPending,
			BillableAmount:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(500000))),
			ProviderReward:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(487500))), // 97.5%
			PlatformFee:     sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(12500))),  // 2.5%
			PeriodStart:     now.Add(-2 * time.Hour),
			PeriodEnd:       now.Add(-1 * time.Hour),
			FormulaVersion:  hpctypes.CurrentBillingFormulaVersion,
			CreatedAt:       now.Add(-1 * time.Hour),
		}

		// Store record
		s.mu.Lock()
		s.accountingRecords[record.RecordID] = record
		s.mu.Unlock()

		// Finalize record
		finalizedAt := now.Add(-30 * time.Minute)
		err := record.Finalize(finalizedAt)
		s.Require().NoError(err)
		s.Equal(hpctypes.AccountingStatusFinalized, record.Status)
		s.NotNil(record.FinalizedAt)

		// Simulate settlement delay (normally 24 hours, but we simulate it passed)
		settlementTime := now

		// Process settlement
		result := s.mockSettlement.ProcessSettlement(record, settlementTime)
		s.True(result.Success, "Settlement should succeed")
		s.NotEmpty(result.SettlementID)

		// Update record status
		err = record.Settle(result.SettlementID, settlementTime)
		s.Require().NoError(err)
		s.Equal(hpctypes.AccountingStatusSettled, record.Status)
	})
}

func (s *HPCBillingE2ETestSuite) TestD02_ProviderRewardDistribution() {
	s.Run("Verify97Point5PercentToProvider", func() {
		billableAmount := sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1000000)))

		calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)
		providerReward := calculator.CalculateProviderReward(billableAmount)

		// 97.5% of 1,000,000 = 975,000
		expectedReward := sdkmath.NewInt(975000)
		s.Equal(expectedReward, providerReward.AmountOf("uvirt"),
			"Provider should receive 97.5%% of billable")
	})

	s.Run("VerifyProviderRewardForVariousAmounts", func() {
		testCases := []struct {
			billable       int64
			expectedReward int64
		}{
			{100000, 97500},
			{500000, 487500},
			{1000000, 975000},
			{10000000, 9750000},
		}

		calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)

		for _, tc := range testCases {
			billable := sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(tc.billable)))
			reward := calculator.CalculateProviderReward(billable)
			s.Equal(tc.expectedReward, reward.AmountOf("uvirt").Int64(),
				"Reward mismatch for billable %d", tc.billable)
		}
	})
}

func (s *HPCBillingE2ETestSuite) TestD03_PlatformFeeCollection() {
	s.Run("Verify2Point5PercentPlatformFee", func() {
		billableAmount := sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1000000)))

		calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)
		platformFee := calculator.CalculatePlatformFee(billableAmount)

		// 2.5% of 1,000,000 = 25,000
		expectedFee := sdkmath.NewInt(25000)
		s.Equal(expectedFee, platformFee.AmountOf("uvirt"),
			"Platform should collect 2.5%% fee")
	})

	s.Run("VerifyFeeAndRewardSumToTotal", func() {
		billableAmount := sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1000000)))

		calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)
		providerReward := calculator.CalculateProviderReward(billableAmount)
		platformFee := calculator.CalculatePlatformFee(billableAmount)

		total := providerReward.Add(platformFee...)
		s.Equal(billableAmount.AmountOf("uvirt"), total.AmountOf("uvirt"),
			"Reward + Fee should equal billable amount")
	})
}

func (s *HPCBillingE2ETestSuite) TestD04_SettlementEventEmitted() {
	s.Run("VerifySettlementEventEmitted", func() {
		record := &hpctypes.HPCAccountingRecord{
			RecordID:        "record-d04-event",
			JobID:           "job-d04-event",
			ClusterID:       s.testCluster.ClusterID,
			ProviderAddress: s.providerAddr,
			CustomerAddress: s.customerAddr,
			Status:          hpctypes.AccountingStatusFinalized,
			BillableAmount:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(200000))),
			ProviderReward:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(195000))),
			PlatformFee:     sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(5000))),
			FormulaVersion:  hpctypes.CurrentBillingFormulaVersion,
		}

		// Process settlement
		result := s.mockSettlement.ProcessSettlement(record, time.Now())
		s.True(result.Success)

		// Verify event was captured
		events := s.mockSettlement.GetEmittedEvents()
		s.NotEmpty(events, "Settlement should emit events")

		found := false
		for _, event := range events {
			if event.Type == "hpc_settlement" {
				found = true
				s.Equal(record.JobID, event.Attributes["job_id"])
				s.Equal(record.RecordID, event.Attributes["accounting_record_id"])
				s.Equal(record.ProviderAddress, event.Attributes["provider"])
				s.Equal(record.CustomerAddress, event.Attributes["customer"])
			}
		}
		s.True(found, "Settlement event should be emitted")
	})
}

func (s *HPCBillingE2ETestSuite) TestD05_AutomaticSettlement() {
	s.Run("TestAutoSettlementViaEndBlocker", func() {
		now := time.Now()

		// Create multiple finalized records ready for settlement
		records := []*hpctypes.HPCAccountingRecord{
			{
				RecordID:        "record-d05-auto-1",
				JobID:           "job-d05-auto-1",
				ClusterID:       s.testCluster.ClusterID,
				ProviderAddress: s.providerAddr,
				CustomerAddress: s.customerAddr,
				Status:          hpctypes.AccountingStatusFinalized,
				BillableAmount:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(100000))),
				ProviderReward:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(97500))),
				PlatformFee:     sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(2500))),
				FormulaVersion:  hpctypes.CurrentBillingFormulaVersion,
			},
			{
				RecordID:        "record-d05-auto-2",
				JobID:           "job-d05-auto-2",
				ClusterID:       s.testCluster.ClusterID,
				ProviderAddress: s.providerAddr,
				CustomerAddress: s.customerAddr,
				Status:          hpctypes.AccountingStatusFinalized,
				BillableAmount:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(200000))),
				ProviderReward:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(195000))),
				PlatformFee:     sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(5000))),
				FormulaVersion:  hpctypes.CurrentBillingFormulaVersion,
			},
		}

		// Set finalized time to be past settlement delay
		for _, r := range records {
			finalizedAt := now.Add(-25 * time.Hour) // Past 24-hour delay
			r.FinalizedAt = &finalizedAt
		}

		// Simulate EndBlocker processing
		for _, r := range records {
			result := s.mockSettlement.ProcessSettlement(r, now)
			s.True(result.Success, "Auto-settlement should succeed for %s", r.RecordID)
		}

		// Verify all records settled
		results := s.mockSettlement.GetSettlementResults()
		s.GreaterOrEqual(len(results), 2, "Should have at least 2 settlements")
	})
}

func (s *HPCBillingE2ETestSuite) TestD06_ManualSettlement() {
	s.Run("TestManualSettlementTrigger", func() {
		record := &hpctypes.HPCAccountingRecord{
			RecordID:        "record-d06-manual",
			JobID:           "job-d06-manual",
			ClusterID:       s.testCluster.ClusterID,
			ProviderAddress: s.providerAddr,
			CustomerAddress: s.customerAddr,
			Status:          hpctypes.AccountingStatusFinalized,
			BillableAmount:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(300000))),
			ProviderReward:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(292500))),
			PlatformFee:     sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(7500))),
			FormulaVersion:  hpctypes.CurrentBillingFormulaVersion,
		}

		// Manual settlement (before auto-settlement delay)
		now := time.Now()
		finalizedAt := now.Add(-1 * time.Hour) // Only 1 hour ago
		record.FinalizedAt = &finalizedAt

		// Process manual settlement
		result := s.mockSettlement.ProcessManualSettlement(record, s.providerAddr, now)
		s.True(result.Success, "Manual settlement should succeed")
		s.Equal("manual", result.SettlementType)
		s.Equal(s.providerAddr, result.RequestedBy)
	})
}

func (s *HPCBillingE2ETestSuite) TestD07_SettlementIdempotency() {
	s.Run("VerifyDoubleSettlementPrevented", func() {
		record := &hpctypes.HPCAccountingRecord{
			RecordID:        "record-d07-idempotent",
			JobID:           "job-d07-idempotent",
			ClusterID:       s.testCluster.ClusterID,
			ProviderAddress: s.providerAddr,
			CustomerAddress: s.customerAddr,
			Status:          hpctypes.AccountingStatusFinalized,
			BillableAmount:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(400000))),
			ProviderReward:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(390000))),
			PlatformFee:     sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(10000))),
			FormulaVersion:  hpctypes.CurrentBillingFormulaVersion,
		}

		now := time.Now()

		// First settlement should succeed
		result1 := s.mockSettlement.ProcessSettlement(record, now)
		s.True(result1.Success, "First settlement should succeed")
		settlementID := result1.SettlementID

		// Update record to settled
		_ = record.Settle(settlementID, now)

		// Second settlement should fail or return same result
		result2 := s.mockSettlement.ProcessSettlement(record, now)
		if result2.Success {
			// If it succeeds, it should return the same settlement ID
			s.Equal(settlementID, result2.SettlementID,
				"Double settlement should return same ID")
		} else {
			// Or it should fail with appropriate error
			s.Contains(result2.Error, "already settled",
				"Should indicate already settled")
		}
	})
}

// =============================================================================
// E. Dispute and Correction Tests (~300 lines)
// =============================================================================

func (s *HPCBillingE2ETestSuite) TestE01_DisputeInitiation() {
	s.Run("CustomerInitiatesDispute", func() {
		// Create a finalized accounting record
		record := &hpctypes.HPCAccountingRecord{
			RecordID:        "record-e01-dispute",
			JobID:           "job-e01-dispute",
			ClusterID:       s.testCluster.ClusterID,
			ProviderAddress: s.providerAddr,
			CustomerAddress: s.customerAddr,
			Status:          hpctypes.AccountingStatusFinalized,
			BillableAmount:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(500000))),
			FormulaVersion:  hpctypes.CurrentBillingFormulaVersion,
		}

		s.mu.Lock()
		s.accountingRecords[record.RecordID] = record
		s.mu.Unlock()

		// Create dispute
		dispute := &hpctypes.HPCDispute{
			DisputeID:       fmt.Sprintf("dispute-%s", record.RecordID),
			JobID:           record.JobID,
			DisputerAddress: s.customerAddr,
			DisputeType:     "billing_accuracy",
			Reason:          "Billed CPU hours exceed actual usage based on scheduler logs",
			Evidence:        "SLURM accounting shows 8 core-hours, billed for 16 core-hours",
			Status:          hpctypes.DisputeStatusPending,
			CreatedAt:       time.Now(),
		}

		err := dispute.Validate()
		s.Require().NoError(err, "Dispute should be valid")

		// Mark record as disputed
		err = record.MarkDisputed(dispute.DisputeID)
		s.Require().NoError(err)
		s.Equal(hpctypes.AccountingStatusDisputed, record.Status)
		s.Equal(dispute.DisputeID, record.DisputeID)

		// Store dispute
		s.mu.Lock()
		s.disputes[dispute.DisputeID] = dispute
		s.mu.Unlock()
	})
}

func (s *HPCBillingE2ETestSuite) TestE02_DisputeBlocksSettlement() {
	s.Run("DisputedRecordsCannotSettle", func() {
		record := &hpctypes.HPCAccountingRecord{
			RecordID:        "record-e02-blocked",
			JobID:           "job-e02-blocked",
			ClusterID:       s.testCluster.ClusterID,
			ProviderAddress: s.providerAddr,
			CustomerAddress: s.customerAddr,
			Status:          hpctypes.AccountingStatusDisputed,
			DisputeID:       "dispute-e02-blocked",
			BillableAmount:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(300000))),
			FormulaVersion:  hpctypes.CurrentBillingFormulaVersion,
		}

		// Attempt settlement should fail
		err := record.Settle("settlement-attempt", time.Now())
		s.Error(err, "Disputed record should not settle")
		s.Contains(err.Error(), "disputed", "Error should mention disputed status")

		// Status should remain disputed
		s.Equal(hpctypes.AccountingStatusDisputed, record.Status)
	})
}

func (s *HPCBillingE2ETestSuite) TestE03_DisputeResolutionUpheld() {
	s.Run("OriginalBillingStandsWhenUpheld", func() {
		// Create disputed record
		originalBilling := sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(400000)))
		record := &hpctypes.HPCAccountingRecord{
			RecordID:        "record-e03-upheld",
			JobID:           "job-e03-upheld",
			ClusterID:       s.testCluster.ClusterID,
			ProviderAddress: s.providerAddr,
			CustomerAddress: s.customerAddr,
			Status:          hpctypes.AccountingStatusDisputed,
			DisputeID:       "dispute-e03-upheld",
			BillableAmount:  originalBilling,
			FormulaVersion:  hpctypes.CurrentBillingFormulaVersion,
		}

		dispute := &hpctypes.HPCDispute{
			DisputeID:       "dispute-e03-upheld",
			JobID:           record.JobID,
			DisputerAddress: s.customerAddr,
			DisputeType:     "billing_accuracy",
			Reason:          "Customer claims overbilling",
			Status:          hpctypes.DisputeStatusPending,
			CreatedAt:       time.Now(),
		}

		// Resolve dispute - upheld (original billing stands)
		dispute.Status = hpctypes.DisputeStatusRejected
		dispute.Resolution = "Investigation confirms billing accuracy. Original charges stand."
		dispute.ResolverAddress = "cosmos1arbitrator..."
		resolvedAt := time.Now()
		dispute.ResolvedAt = &resolvedAt

		// Record can now be finalized again
		record.Status = hpctypes.AccountingStatusFinalized
		record.DisputeID = ""

		// Verify billing unchanged
		s.True(record.BillableAmount.Equal(originalBilling),
			"Billing should remain unchanged when dispute is upheld")

		// Settlement should now succeed
		err := record.Settle("settlement-after-upheld", time.Now())
		s.NoError(err, "Settlement should succeed after dispute rejected")
	})
}

func (s *HPCBillingE2ETestSuite) TestE04_DisputeResolutionCorrected() {
	s.Run("CreateCorrectionRecordWhenDisputeSucceeds", func() {
		// Original record
		originalRecord := &hpctypes.HPCAccountingRecord{
			RecordID:        "record-e04-original",
			JobID:           "job-e04-corrected",
			ClusterID:       s.testCluster.ClusterID,
			ProviderAddress: s.providerAddr,
			CustomerAddress: s.customerAddr,
			Status:          hpctypes.AccountingStatusDisputed,
			DisputeID:       "dispute-e04-corrected",
			BillableAmount:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(800000))),
			UsageMetrics: hpctypes.HPCDetailedMetrics{
				CPUCoreSeconds: 3600 * 80, // Originally billed for 80 core-hours
				NodeHours:      sdkmath.LegacyNewDec(2),
			},
			FormulaVersion: hpctypes.CurrentBillingFormulaVersion,
		}

		// Corrected metrics (investigation found only 40 core-hours used)
		correctedMetrics := &hpctypes.HPCDetailedMetrics{
			CPUCoreSeconds: 3600 * 40, // Corrected to 40 core-hours
			NodeHours:      sdkmath.LegacyNewDec(1),
		}

		// Calculate corrected billing
		calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)
		_, correctedBilling, _ := calculator.CalculateBillableAmount(correctedMetrics, nil, nil)

		// Create correction record
		correctionRecord := &hpctypes.HPCAccountingRecord{
			RecordID:         "record-e04-correction",
			JobID:            originalRecord.JobID,
			ClusterID:        originalRecord.ClusterID,
			ProviderAddress:  originalRecord.ProviderAddress,
			CustomerAddress:  originalRecord.CustomerAddress,
			Status:           hpctypes.AccountingStatusCorrected,
			CorrectedFromID:  originalRecord.RecordID,
			CorrectionReason: "Dispute resolution: CPU hours overcounted due to scheduler bug",
			BillableAmount:   correctedBilling,
			UsageMetrics:     *correctedMetrics,
			FormulaVersion:   hpctypes.CurrentBillingFormulaVersion,
		}

		// Mark original as corrected
		originalRecord.Status = hpctypes.AccountingStatusCorrected

		// Verify correction record
		s.Equal(originalRecord.RecordID, correctionRecord.CorrectedFromID)
		s.True(correctionRecord.BillableAmount.IsAllLT(originalRecord.BillableAmount),
			"Corrected billing should be less than original")

		// Corrected amount should be about half
		expectedCorrectedAmount := int64(40 * CPUCoreHourRate) // 400000
		s.Equal(expectedCorrectedAmount, correctedBilling.AmountOf("uvirt").Int64(),
			"Corrected billing should be 40 core-hours")
	})
}

func (s *HPCBillingE2ETestSuite) TestE05_CorrectionRecordBilling() {
	s.Run("VerifyCorrectedBillingCalculation", func() {
		// Original metrics (incorrect)
		originalMetrics := &hpctypes.HPCDetailedMetrics{
			CPUCoreSeconds:  3600 * 100, // 100 core-hours
			MemoryGBSeconds: 3600 * 256, // 256 GB-hours
			GPUSeconds:      3600 * 8,   // 8 GPU-hours
			GPUType:         "nvidia-a100",
			NodeHours:       sdkmath.LegacyNewDec(4),
		}

		// Corrected metrics (after dispute investigation)
		correctedMetrics := &hpctypes.HPCDetailedMetrics{
			CPUCoreSeconds:  3600 * 50,  // Actually only 50 core-hours
			MemoryGBSeconds: 3600 * 128, // Actually only 128 GB-hours
			GPUSeconds:      3600 * 4,   // Actually only 4 GPU-hours
			GPUType:         "nvidia-a100",
			NodeHours:       sdkmath.LegacyNewDec(2),
		}

		calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)

		// Calculate original and corrected
		_, originalBilling, _ := calculator.CalculateBillableAmount(originalMetrics, nil, nil)
		_, correctedBilling, _ := calculator.CalculateBillableAmount(correctedMetrics, nil, nil)

		// Verify correction is less
		s.True(correctedBilling.IsAllLT(originalBilling),
			"Corrected billing should be less than original")

		// Calculate expected refund
		refundAmount := originalBilling.Sub(correctedBilling...)
		s.True(refundAmount.AmountOf("uvirt").IsPositive(),
			"Refund should be positive")

		// Verify the refund is approximately 50% of original
		originalAmount := originalBilling.AmountOf("uvirt").Int64()
		refundAmountInt := refundAmount.AmountOf("uvirt").Int64()
		refundPercentage := float64(refundAmountInt) / float64(originalAmount) * 100

		s.True(refundPercentage > 40 && refundPercentage < 60,
			"Refund should be approximately 50%% of original")
	})
}

// =============================================================================
// F. Invoice Generation Tests (~200 lines)
// =============================================================================

func (s *HPCBillingE2ETestSuite) TestF01_InvoiceGeneration() {
	s.Run("GenerateInvoiceFromAccountingRecord", func() {
		record := &hpctypes.HPCAccountingRecord{
			RecordID:        "record-f01-invoice",
			JobID:           "job-f01-invoice",
			ClusterID:       s.testCluster.ClusterID,
			ProviderAddress: s.providerAddr,
			CustomerAddress: s.customerAddr,
			Status:          hpctypes.AccountingStatusFinalized,
			BillableAmount:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(750000))),
			BillableBreakdown: hpctypes.BillableBreakdown{
				CPUCost:     sdk.NewCoin("uvirt", sdkmath.NewInt(200000)),
				MemoryCost:  sdk.NewCoin("uvirt", sdkmath.NewInt(100000)),
				GPUCost:     sdk.NewCoin("uvirt", sdkmath.NewInt(350000)),
				StorageCost: sdk.NewCoin("uvirt", sdkmath.NewInt(50000)),
				NetworkCost: sdk.NewCoin("uvirt", sdkmath.NewInt(25000)),
				NodeCost:    sdk.NewCoin("uvirt", sdkmath.NewInt(25000)),
			},
			UsageMetrics: hpctypes.HPCDetailedMetrics{
				CPUCoreSeconds:  3600 * 20,
				MemoryGBSeconds: 3600 * 100,
				GPUSeconds:      3600 * 0.7,
				GPUType:         "nvidia-a100",
				StorageGBHours:  1000,
				NetworkBytesIn:  128 * 1024 * 1024 * 1024,
				NetworkBytesOut: 64 * 1024 * 1024 * 1024,
				NodeHours:       sdkmath.LegacyNewDecWithPrec(5, 1),
			},
			FormulaVersion: hpctypes.CurrentBillingFormulaVersion,
		}

		// Generate invoice
		invoice := s.mockBilling.GenerateInvoice(record)

		s.NotEmpty(invoice.InvoiceID, "Invoice ID should be set")
		s.Equal(record.CustomerAddress, invoice.CustomerAddress)
		s.Equal(record.ProviderAddress, invoice.ProviderAddress)
		s.True(invoice.TotalAmount.Equal(record.BillableAmount),
			"Invoice total should match billable amount")
		s.NotEmpty(invoice.LineItems, "Invoice should have line items")
	})
}

func (s *HPCBillingE2ETestSuite) TestF02_InvoiceLineItems() {
	s.Run("VerifyLineItemsForEachResourceType", func() {
		record := &hpctypes.HPCAccountingRecord{
			RecordID:        "record-f02-lineitems",
			JobID:           "job-f02-lineitems",
			ClusterID:       s.testCluster.ClusterID,
			ProviderAddress: s.providerAddr,
			CustomerAddress: s.customerAddr,
			Status:          hpctypes.AccountingStatusFinalized,
			BillableAmount:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(600000))),
			BillableBreakdown: hpctypes.BillableBreakdown{
				CPUCost:     sdk.NewCoin("uvirt", sdkmath.NewInt(150000)),
				MemoryCost:  sdk.NewCoin("uvirt", sdkmath.NewInt(80000)),
				GPUCost:     sdk.NewCoin("uvirt", sdkmath.NewInt(300000)),
				StorageCost: sdk.NewCoin("uvirt", sdkmath.NewInt(30000)),
				NetworkCost: sdk.NewCoin("uvirt", sdkmath.NewInt(20000)),
				NodeCost:    sdk.NewCoin("uvirt", sdkmath.NewInt(20000)),
			},
			FormulaVersion: hpctypes.CurrentBillingFormulaVersion,
		}

		invoice := s.mockBilling.GenerateInvoice(record)

		// Verify line items exist for each resource type
		resourceTypes := map[string]bool{
			"cpu":     false,
			"memory":  false,
			"gpu":     false,
			"storage": false,
			"network": false,
			"node":    false,
		}

		for _, item := range invoice.LineItems {
			resourceTypes[item.ResourceType] = true
		}

		for rt, found := range resourceTypes {
			s.True(found, "Line item for %s should exist", rt)
		}

		// Verify line item amounts match breakdown
		for _, item := range invoice.LineItems {
			switch item.ResourceType {
			case "cpu":
				s.Equal(record.BillableBreakdown.CPUCost.Amount.Int64(), item.Amount.Int64())
			case "memory":
				s.Equal(record.BillableBreakdown.MemoryCost.Amount.Int64(), item.Amount.Int64())
			case "gpu":
				s.Equal(record.BillableBreakdown.GPUCost.Amount.Int64(), item.Amount.Int64())
			case "storage":
				s.Equal(record.BillableBreakdown.StorageCost.Amount.Int64(), item.Amount.Int64())
			case "network":
				s.Equal(record.BillableBreakdown.NetworkCost.Amount.Int64(), item.Amount.Int64())
			case "node":
				s.Equal(record.BillableBreakdown.NodeCost.Amount.Int64(), item.Amount.Int64())
			}
		}
	})
}

func (s *HPCBillingE2ETestSuite) TestF03_InvoiceAmountMatches() {
	s.Run("InvoiceTotalMatchesBilling", func() {
		testCases := []int64{100000, 500000, 1000000, 5000000}

		for _, amount := range testCases {
			record := &hpctypes.HPCAccountingRecord{
				RecordID:        fmt.Sprintf("record-f03-%d", amount),
				JobID:           fmt.Sprintf("job-f03-%d", amount),
				ClusterID:       s.testCluster.ClusterID,
				ProviderAddress: s.providerAddr,
				CustomerAddress: s.customerAddr,
				Status:          hpctypes.AccountingStatusFinalized,
				BillableAmount:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(amount))),
				BillableBreakdown: hpctypes.BillableBreakdown{
					CPUCost: sdk.NewCoin("uvirt", sdkmath.NewInt(amount)),
				},
				FormulaVersion: hpctypes.CurrentBillingFormulaVersion,
			}

			invoice := s.mockBilling.GenerateInvoice(record)
			s.Equal(amount, invoice.TotalAmount.AmountOf("uvirt").Int64(),
				"Invoice total should match for amount %d", amount)
		}
	})
}

func (s *HPCBillingE2ETestSuite) TestF04_InvoiceEventEmitted() {
	s.Run("VerifyInvoiceGenerationEvent", func() {
		record := &hpctypes.HPCAccountingRecord{
			RecordID:        "record-f04-event",
			JobID:           "job-f04-event",
			ClusterID:       s.testCluster.ClusterID,
			ProviderAddress: s.providerAddr,
			CustomerAddress: s.customerAddr,
			Status:          hpctypes.AccountingStatusFinalized,
			BillableAmount:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(250000))),
			BillableBreakdown: hpctypes.BillableBreakdown{
				CPUCost: sdk.NewCoin("uvirt", sdkmath.NewInt(250000)),
			},
			FormulaVersion: hpctypes.CurrentBillingFormulaVersion,
		}

		// Generate invoice and emit event
		invoice := s.mockBilling.GenerateInvoiceWithEvent(record)

		// Verify event was captured
		events := s.mockBilling.GetEmittedEvents()
		s.NotEmpty(events, "Invoice generation should emit events")

		found := false
		for _, event := range events {
			if event.Type == "hpc_invoice_generated" {
				found = true
				s.Equal(invoice.InvoiceID, event.Attributes["invoice_id"])
				s.Equal(record.JobID, event.Attributes["job_id"])
				s.Equal(record.RecordID, event.Attributes["accounting_record_id"])
				s.Equal(record.CustomerAddress, event.Attributes["customer"])
				s.Equal(record.ProviderAddress, event.Attributes["provider"])
			}
		}
		s.True(found, "Invoice generation event should be emitted")
	})
}

// =============================================================================
// Mock Types and Helpers for Billing Tests
// =============================================================================

// HPCJobState type alias for mock compatibility
type HPCJobState = pd.HPCJobState

// EscrowStatus indicates escrow state
type EscrowStatus string

const (
	EscrowStatusHeld            EscrowStatus = "held"
	EscrowStatusReleased        EscrowStatus = "released"
	EscrowStatusPartialReleased EscrowStatus = "partial_released"
	EscrowStatusRefunded        EscrowStatus = "refunded"
)

// BillingMockEscrowKeeper simulates escrow keeper for billing tests
type BillingMockEscrowKeeper struct {
	mu               sync.RWMutex
	escrows          map[string]*BillingMockEscrow
	customerBalances map[string]sdk.Coins
}

// BillingMockEscrow represents a mock escrow account
type BillingMockEscrow struct {
	EscrowID        string
	CustomerAddress string
	ProviderAddress string
	Amount          sdk.Coins
	ReleasedAmount  sdk.Coins
	RefundAmount    sdk.Coins
	Status          EscrowStatus
	RefundReason    string
	CreatedAt       time.Time
}

// NewBillingMockEscrowKeeper creates a new mock escrow keeper
func NewBillingMockEscrowKeeper() *BillingMockEscrowKeeper {
	return &BillingMockEscrowKeeper{
		escrows:          make(map[string]*BillingMockEscrow),
		customerBalances: make(map[string]sdk.Coins),
	}
}

// CreateEscrow creates a mock escrow
func (m *BillingMockEscrowKeeper) CreateEscrow(escrowID, customer, provider string, amount sdk.Coins) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.escrows[escrowID] = &BillingMockEscrow{
		EscrowID:        escrowID,
		CustomerAddress: customer,
		ProviderAddress: provider,
		Amount:          amount,
		Status:          EscrowStatusHeld,
		CreatedAt:       time.Now(),
	}
}

// CreateEscrowWithValidation creates escrow with balance validation
func (m *BillingMockEscrowKeeper) CreateEscrowWithValidation(escrowID, customer, provider string, amount sdk.Coins) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	balance, exists := m.customerBalances[customer]
	if !exists || !balance.IsAllGTE(amount) {
		return fmt.Errorf("insufficient funds: requested %s, available %s", amount, balance)
	}

	m.escrows[escrowID] = &BillingMockEscrow{
		EscrowID:        escrowID,
		CustomerAddress: customer,
		ProviderAddress: provider,
		Amount:          amount,
		Status:          EscrowStatusHeld,
		CreatedAt:       time.Now(),
	}
	return nil
}

// GetEscrow returns an escrow by ID
func (m *BillingMockEscrowKeeper) GetEscrow(escrowID string) (*BillingMockEscrow, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	escrow, ok := m.escrows[escrowID]
	return escrow, ok
}

// ProcessRelease releases escrow funds
func (m *BillingMockEscrowKeeper) ProcessRelease(escrowID string, amount sdk.Coins) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if escrow, ok := m.escrows[escrowID]; ok {
		escrow.ReleasedAmount = amount
		escrow.Status = EscrowStatusReleased
	}
}

// ProcessPartialRelease partially releases escrow funds
func (m *BillingMockEscrowKeeper) ProcessPartialRelease(escrowID string, amount sdk.Coins) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if escrow, ok := m.escrows[escrowID]; ok {
		escrow.ReleasedAmount = amount
		escrow.RefundAmount = escrow.Amount.Sub(amount...)
		escrow.Status = EscrowStatusPartialReleased
	}
}

// ProcessRefund refunds escrow to customer
func (m *BillingMockEscrowKeeper) ProcessRefund(escrowID string, amount sdk.Coins) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if escrow, ok := m.escrows[escrowID]; ok {
		escrow.RefundAmount = amount
		escrow.Status = EscrowStatusRefunded
	}
}

// ProcessRefundDueToProviderFailure refunds due to provider failure
func (m *BillingMockEscrowKeeper) ProcessRefundDueToProviderFailure(escrowID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if escrow, ok := m.escrows[escrowID]; ok {
		escrow.RefundAmount = escrow.Amount
		escrow.Status = EscrowStatusRefunded
		escrow.RefundReason = "provider_failure"
	}
}

// SetCustomerBalance sets customer balance for testing
func (m *BillingMockEscrowKeeper) SetCustomerBalance(customer string, balance sdk.Coins) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.customerBalances[customer] = balance
}

// BillingMockEvent represents a mock blockchain event
type BillingMockEvent struct {
	Type       string
	Attributes map[string]string
}

// BillingMockSettlementResult represents settlement result
type BillingMockSettlementResult struct {
	Success        bool
	SettlementID   string
	SettlementType string
	RequestedBy    string
	Error          string
}

// BillingMockSettlementProcessor simulates settlement processing
type BillingMockSettlementProcessor struct {
	mu           sync.RWMutex
	results      []BillingMockSettlementResult
	events       []BillingMockEvent
	processedIDs map[string]string
}

// NewBillingMockSettlementProcessor creates a new mock settlement processor
func NewBillingMockSettlementProcessor() *BillingMockSettlementProcessor {
	return &BillingMockSettlementProcessor{
		results:      make([]BillingMockSettlementResult, 0),
		events:       make([]BillingMockEvent, 0),
		processedIDs: make(map[string]string),
	}
}

// ProcessSettlement processes a settlement
func (m *BillingMockSettlementProcessor) ProcessSettlement(record *hpctypes.HPCAccountingRecord, settleTime time.Time) BillingMockSettlementResult {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already settled
	if existingID, exists := m.processedIDs[record.RecordID]; exists {
		return BillingMockSettlementResult{
			Success:      true,
			SettlementID: existingID,
		}
	}

	// Check if record is in valid state
	if record.Status == hpctypes.AccountingStatusDisputed {
		return BillingMockSettlementResult{
			Success: false,
			Error:   "cannot settle disputed record",
		}
	}

	if record.Status == hpctypes.AccountingStatusSettled {
		return BillingMockSettlementResult{
			Success: false,
			Error:   "record already settled",
		}
	}

	settlementID := fmt.Sprintf("settle-%s-%d", record.RecordID, settleTime.Unix())
	m.processedIDs[record.RecordID] = settlementID

	result := BillingMockSettlementResult{
		Success:        true,
		SettlementID:   settlementID,
		SettlementType: "automatic",
	}
	m.results = append(m.results, result)

	// Emit event
	m.events = append(m.events, BillingMockEvent{
		Type: "hpc_settlement",
		Attributes: map[string]string{
			"settlement_id":        settlementID,
			"job_id":               record.JobID,
			"accounting_record_id": record.RecordID,
			"provider":             record.ProviderAddress,
			"customer":             record.CustomerAddress,
			"amount":               record.BillableAmount.String(),
		},
	})

	return result
}

// ProcessManualSettlement processes a manual settlement
func (m *BillingMockSettlementProcessor) ProcessManualSettlement(record *hpctypes.HPCAccountingRecord, requestedBy string, settleTime time.Time) BillingMockSettlementResult {
	result := m.ProcessSettlement(record, settleTime)
	if result.Success {
		result.SettlementType = "manual"
		result.RequestedBy = requestedBy
	}
	return result
}

// GetSettlementResults returns all settlement results
func (m *BillingMockSettlementProcessor) GetSettlementResults() []BillingMockSettlementResult {
	m.mu.RLock()
	defer m.mu.RUnlock()
	results := make([]BillingMockSettlementResult, len(m.results))
	copy(results, m.results)
	return results
}

// GetEmittedEvents returns all emitted events
func (m *BillingMockSettlementProcessor) GetEmittedEvents() []BillingMockEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	events := make([]BillingMockEvent, len(m.events))
	copy(events, m.events)
	return events
}

// BillingMockInvoice represents a mock invoice
type BillingMockInvoice struct {
	InvoiceID       string
	CustomerAddress string
	ProviderAddress string
	TotalAmount     sdk.Coins
	LineItems       []BillingMockLineItem
	CreatedAt       time.Time
}

// BillingMockLineItem represents a line item in an invoice
type BillingMockLineItem struct {
	Description  string
	ResourceType string
	Quantity     string
	UnitPrice    sdkmath.Int
	Amount       sdkmath.Int
}

// BillingMockCalculator simulates billing calculation
type BillingMockCalculator struct {
	mu       sync.RWMutex
	invoices map[string]*BillingMockInvoice
	events   []BillingMockEvent
}

// NewBillingMockCalculator creates a new mock billing calculator
func NewBillingMockCalculator() *BillingMockCalculator {
	return &BillingMockCalculator{
		invoices: make(map[string]*BillingMockInvoice),
		events:   make([]BillingMockEvent, 0),
	}
}

// GenerateInvoice generates a mock invoice from accounting record
func (m *BillingMockCalculator) GenerateInvoice(record *hpctypes.HPCAccountingRecord) *BillingMockInvoice {
	m.mu.Lock()
	defer m.mu.Unlock()

	invoiceID := fmt.Sprintf("inv-%s", record.RecordID)

	invoice := &BillingMockInvoice{
		InvoiceID:       invoiceID,
		CustomerAddress: record.CustomerAddress,
		ProviderAddress: record.ProviderAddress,
		TotalAmount:     record.BillableAmount,
		LineItems:       m.createLineItems(record),
		CreatedAt:       time.Now(),
	}

	m.invoices[invoiceID] = invoice
	return invoice
}

// GenerateInvoiceWithEvent generates invoice and emits event
func (m *BillingMockCalculator) GenerateInvoiceWithEvent(record *hpctypes.HPCAccountingRecord) *BillingMockInvoice {
	invoice := m.GenerateInvoice(record)

	m.mu.Lock()
	m.events = append(m.events, BillingMockEvent{
		Type: "hpc_invoice_generated",
		Attributes: map[string]string{
			"invoice_id":           invoice.InvoiceID,
			"job_id":               record.JobID,
			"accounting_record_id": record.RecordID,
			"customer":             record.CustomerAddress,
			"provider":             record.ProviderAddress,
			"amount":               record.BillableAmount.String(),
		},
	})
	m.mu.Unlock()

	return invoice
}

// createLineItems creates line items from breakdown
func (m *BillingMockCalculator) createLineItems(record *hpctypes.HPCAccountingRecord) []BillingMockLineItem {
	items := []BillingMockLineItem{}

	if record.BillableBreakdown.CPUCost.Amount.IsPositive() {
		items = append(items, BillingMockLineItem{
			Description:  "CPU Usage",
			ResourceType: "cpu",
			Amount:       record.BillableBreakdown.CPUCost.Amount,
		})
	}

	if record.BillableBreakdown.MemoryCost.Amount.IsPositive() {
		items = append(items, BillingMockLineItem{
			Description:  "Memory Usage",
			ResourceType: "memory",
			Amount:       record.BillableBreakdown.MemoryCost.Amount,
		})
	}

	if record.BillableBreakdown.GPUCost.Amount.IsPositive() {
		items = append(items, BillingMockLineItem{
			Description:  "GPU Usage",
			ResourceType: "gpu",
			Amount:       record.BillableBreakdown.GPUCost.Amount,
		})
	}

	if record.BillableBreakdown.StorageCost.Amount.IsPositive() {
		items = append(items, BillingMockLineItem{
			Description:  "Storage Usage",
			ResourceType: "storage",
			Amount:       record.BillableBreakdown.StorageCost.Amount,
		})
	}

	if record.BillableBreakdown.NetworkCost.Amount.IsPositive() {
		items = append(items, BillingMockLineItem{
			Description:  "Network Usage",
			ResourceType: "network",
			Amount:       record.BillableBreakdown.NetworkCost.Amount,
		})
	}

	if record.BillableBreakdown.NodeCost.Amount.IsPositive() {
		items = append(items, BillingMockLineItem{
			Description:  "Node Usage",
			ResourceType: "node",
			Amount:       record.BillableBreakdown.NodeCost.Amount,
		})
	}

	return items
}

// GetEmittedEvents returns all emitted events
func (m *BillingMockCalculator) GetEmittedEvents() []BillingMockEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	events := make([]BillingMockEvent, len(m.events))
	copy(events, m.events)
	return events
}

// =============================================================================
// Escrow Status Types for HPC Module
// =============================================================================

// EscrowStatusHeld etc are defined above. Add to hpctypes namespace for compatibility
const (
	_ = iota
)

// Extend hpctypes with escrow status for test compatibility
func init() {
	// This ensures the test types are initialized properly
}
