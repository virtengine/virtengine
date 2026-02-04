//go:build e2e.integration

// Package e2e contains end-to-end integration tests.
//
// HPC-E2E-001: Settlement and billing pipeline tests.
package e2e

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/tests/e2e/fixtures"
	"github.com/virtengine/virtengine/testutil"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// HPCSettlementE2ETestSuite validates billing and settlement behaviors.
type HPCSettlementE2ETestSuite struct {
	*testutil.NetworkTestSuite

	providerAddr string
	customerAddr string

	billingRules   hpctypes.HPCBillingRules
	mockEscrow     *BillingMockEscrowKeeper
	mockSettlement *BillingMockSettlementProcessor
}

func TestHPCSettlementE2E(t *testing.T) {
	suite.Run(t, &HPCSettlementE2ETestSuite{
		NetworkTestSuite: testutil.NewNetworkTestSuite(nil, &HPCSettlementE2ETestSuite{}),
	})
}

func (s *HPCSettlementE2ETestSuite) SetupSuite() {
	s.NetworkTestSuite.SetupSuite()
	val := s.Network().Validators[0]
	s.providerAddr = val.Address.String()
	s.customerAddr = val.Address.String()

	s.billingRules = newSettlementBillingRules()
	s.mockEscrow = NewBillingMockEscrowKeeper()
	s.mockSettlement = NewBillingMockSettlementProcessor()
}

func (s *HPCSettlementE2ETestSuite) TearDownSuite() {
	s.NetworkTestSuite.TearDownSuite()
}

func newSettlementBillingRules() hpctypes.HPCBillingRules {
	denom := "uvirt"
	return hpctypes.HPCBillingRules{
		FormulaVersion: hpctypes.CurrentBillingFormulaVersion,
		ResourceRates: hpctypes.HPCResourceRates{
			CPUCoreHourRate:   sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(CPUCoreHourRate)),
			MemoryGBHourRate:  sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(MemoryGBHourRate)),
			GPUHourRate:       sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(GPUA100HourRate)),
			StorageGBHourRate: sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(StorageGBHourRate)),
			NetworkGBRate:     sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(NetworkGBRate)),
			NodeHourRate:      sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(NodeHourRate)),
			GPUTypeRates: map[string]sdk.DecCoin{
				"nvidia-a100": sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(GPUA100HourRate)),
				"nvidia-v100": sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(GPUV100HourRate)),
				"nvidia-t4":   sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(GPUT4HourRate)),
			},
		},
		PlatformFeeRateBps:        PlatformFeeBps,
		ProviderRewardRateBps:     ProviderRewardBps,
		MinimumCharge:             sdk.NewCoin(denom, sdkmath.NewInt(MinimumCharge)),
		BillingGranularitySeconds: 60,
	}
}

func (s *HPCSettlementE2ETestSuite) TestUsageCalculationAccuracy() {
	metrics := fixtures.StandardJobMetrics(3600)
	calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)
	breakdown, billable, err := calculator.CalculateBillableAmount(&hpctypes.HPCDetailedMetrics{
		WallClockSeconds: metrics.WallClockSeconds,
		CPUCoreSeconds:   metrics.CPUCoreSeconds,
		MemoryGBSeconds:  metrics.MemoryGBSeconds,
		GPUSeconds:       metrics.GPUSeconds,
		StorageGBHours:   metrics.StorageGBHours,
		NetworkBytesIn:   metrics.NetworkBytesIn,
		NetworkBytesOut:  metrics.NetworkBytesOut,
		NodeHours:        sdkmath.LegacyNewDec(1),
		NodesUsed:        metrics.NodesUsed,
	}, nil, nil)

	s.Require().NoError(err)
	s.True(billable.AmountOf("uvirt").IsPositive())
	s.True(breakdown.CPUCost.Amount.IsPositive())
	s.True(breakdown.MemoryCost.Amount.IsPositive())
}

func (s *HPCSettlementE2ETestSuite) TestEscrowHoldAndRelease() {
	amount := sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(500000)))
	s.mockEscrow.SetCustomerBalance(s.customerAddr, amount)

	err := s.mockEscrow.CreateEscrowWithValidation("escrow-1", s.customerAddr, s.providerAddr, amount)
	s.Require().NoError(err)

	escrow, ok := s.mockEscrow.GetEscrow("escrow-1")
	s.True(ok)
	s.Equal(EscrowStatusHeld, escrow.Status)

	s.mockEscrow.ProcessRelease("escrow-1", amount)
	escrow, ok = s.mockEscrow.GetEscrow("escrow-1")
	s.True(ok)
	s.Equal(EscrowStatusReleased, escrow.Status)
}

func (s *HPCSettlementE2ETestSuite) TestSettlementTiming() {
	now := time.Now()
	const settlementDelaySec = 86400
	finalizedAt := now.Add(-time.Duration(settlementDelaySec)*time.Second - time.Minute)
	record := &hpctypes.HPCAccountingRecord{
		RecordID:        "record-settle-1",
		JobID:           "job-settle-1",
		ClusterID:       "cluster-1",
		ProviderAddress: s.providerAddr,
		CustomerAddress: s.customerAddr,
		Status:          hpctypes.AccountingStatusFinalized,
		BillableAmount:  sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1000000))),
		FinalizedAt:     &finalizedAt,
	}

	settleAllowed := now.After(record.FinalizedAt.Add(time.Duration(settlementDelaySec) * time.Second))
	s.True(settleAllowed)

	result := s.mockSettlement.ProcessSettlement(record, now)
	s.True(result.Success)
}

func (s *HPCSettlementE2ETestSuite) TestPartialJobBilling() {
	calculator := hpctypes.NewHPCBillingCalculator(s.billingRules)
	metrics := &hpctypes.HPCDetailedMetrics{
		WallClockSeconds: 600,
		CPUCoreSeconds:   600 * 2,
		MemoryGBSeconds:  600 * 4,
		NodeHours:        sdkmath.LegacyNewDecWithPrec(1, 1),
		NodesUsed:        1,
	}

	_, billable, err := calculator.CalculateBillableAmount(metrics, nil, nil)
	s.Require().NoError(err)
	// Minimum charge enforced
	s.True(billable.AmountOf("uvirt").GTE(sdkmath.NewInt(MinimumCharge)))
}

func (s *HPCSettlementE2ETestSuite) TestRefundScenario() {
	amount := sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(200000)))
	s.mockEscrow.SetCustomerBalance(s.customerAddr, amount)

	err := s.mockEscrow.CreateEscrowWithValidation("escrow-refund", s.customerAddr, s.providerAddr, amount)
	s.Require().NoError(err)

	partial := sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(50000)))
	s.mockEscrow.ProcessPartialRelease("escrow-refund", partial)
	escrow, ok := s.mockEscrow.GetEscrow("escrow-refund")
	s.True(ok)
	s.Equal(EscrowStatusPartialReleased, escrow.Status)
	s.True(escrow.RefundAmount.IsAllGT(sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(0)))))
}
