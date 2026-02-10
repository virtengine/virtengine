// Copyright 2024 The VirtEngine Authors
// This file is part of the VirtEngine library.

package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/escrow/types/billing"
)

// SettlementIntegrationTestSuite tests the settlement integration keeper
type SettlementIntegrationTestSuite struct {
	suite.Suite
}

func TestSettlementIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(SettlementIntegrationTestSuite))
}

func (s *SettlementIntegrationTestSuite) SetupTest() {
	// Basic setup - we'll use mock stores for testing
}

// TestDefaultFeeConfig tests that the default fee config is valid
func (s *SettlementIntegrationTestSuite) TestDefaultFeeConfig() {
	config := billing.DefaultFeeConfig()

	s.Require().NotNil(config)

	// Platform fee should be 2.5% (250 basis points)
	expectedPlatform := sdkmath.LegacyNewDecWithPrec(250, 4)
	s.Require().Equal(expectedPlatform, config.PlatformFeeRate)

	// Network fee should be 0.5% (50 basis points)
	expectedNetwork := sdkmath.LegacyNewDecWithPrec(50, 4)
	s.Require().Equal(expectedNetwork, config.NetworkFeeRate)

	// Community fee should be 1% (100 basis points)
	expectedCommunity := sdkmath.LegacyNewDecWithPrec(100, 4)
	s.Require().Equal(expectedCommunity, config.CommunityPoolRate)

	// Take fee should be 4% (400 basis points)
	expectedTake := sdkmath.LegacyNewDecWithPrec(400, 4)
	s.Require().Equal(expectedTake, config.TakeRate)

	// Total fee should be 8%
	total := config.PlatformFeeRate.Add(config.NetworkFeeRate).Add(config.CommunityPoolRate).Add(config.TakeRate)
	expectedTotal := sdkmath.LegacyNewDecWithPrec(800, 4)
	s.Require().Equal(expectedTotal, total)
}

// TestFeeBreakdown tests fee breakdown calculation
func (s *SettlementIntegrationTestSuite) TestFeeBreakdown() {
	config := billing.DefaultFeeConfig()

	// Test with 1000 uvirt
	grossAmount := sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000))
	breakdown := config.CalculateFees(grossAmount)

	s.Require().NotNil(breakdown)

	// Platform fee: 1000 * 2.5% = 25
	s.Require().True(breakdown.PlatformFee.AmountOf("uvirt").Equal(sdkmath.NewInt(25)))

	// Network fee: 1000 * 0.5% = 5
	s.Require().True(breakdown.NetworkFee.AmountOf("uvirt").Equal(sdkmath.NewInt(5)))

	// Community fee: 1000 * 1% = 10
	s.Require().True(breakdown.CommunityFee.AmountOf("uvirt").Equal(sdkmath.NewInt(10)))

	// Take fee: 1000 * 4% = 40
	s.Require().True(breakdown.TakeFee.AmountOf("uvirt").Equal(sdkmath.NewInt(40)))

	// Total fees: 80
	s.Require().True(breakdown.TotalFees.AmountOf("uvirt").Equal(sdkmath.NewInt(80)))

	// Net amount: 1000 - 80 = 920
	s.Require().True(breakdown.NetAmount.AmountOf("uvirt").Equal(sdkmath.NewInt(920)))
}

// TestSettlementRecordStatus tests settlement record status transitions
func (s *SettlementIntegrationTestSuite) TestSettlementRecordStatus() {
	now := time.Now().UTC()
	config := billing.DefaultFeeConfig()
	grossAmount := sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000))
	breakdown := config.CalculateFees(grossAmount)

	// Use valid bech32 addresses for testing
	provider := sdk.AccAddress(make([]byte, 20)).String()
	customer := sdk.AccAddress(make([]byte, 20)).String()

	record := billing.NewSettlementRecord(
		"settlement-001",
		"invoice-001",
		"escrow-001",
		provider,
		customer,
		grossAmount,
		breakdown,
		100,
		now,
	)

	s.Require().Equal(billing.SettlementStatusPending, record.Status)

	// Validate the record
	err := record.Validate()
	s.Require().NoError(err)
}

// TestSettlementRecordHoldback tests holdback functionality
func (s *SettlementIntegrationTestSuite) TestSettlementRecordHoldback() {
	now := time.Now().UTC()
	config := billing.DefaultFeeConfig()
	grossAmount := sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000))
	breakdown := config.CalculateFees(grossAmount)

	// Use valid bech32 addresses for testing
	provider := sdk.AccAddress(make([]byte, 20)).String()
	customer := sdk.AccAddress(make([]byte, 20)).String()

	record := billing.NewSettlementRecord(
		"settlement-002",
		"invoice-002",
		"escrow-002",
		provider,
		customer,
		grossAmount,
		breakdown,
		100,
		now,
	)

	// Set a holdback
	holdbackAmount := sdk.NewCoins(sdk.NewInt64Coin("uvirt", 100))
	err := record.SetHoldback(holdbackAmount, "Dispute pending")
	s.Require().NoError(err)

	s.Require().True(record.HoldbackAmount.AmountOf("uvirt").Equal(sdkmath.NewInt(100)))
	s.Require().Equal(billing.SettlementStatusHeldBack, record.Status)

	// Release holdback
	err = record.ReleaseHoldback()
	s.Require().NoError(err)
	s.Require().True(record.HoldbackAmount.Empty())
	s.Require().Equal(billing.SettlementStatusPending, record.Status)
}

// TestTreasuryAllocationValidation tests treasury allocation validation
func (s *SettlementIntegrationTestSuite) TestTreasuryAllocationValidation() {
	now := time.Now().UTC()

	allocation := billing.TreasuryAllocation{
		AllocationID: "alloc-001",
		InvoiceID:    "invoice-001",
		SettlementID: "settlement-001",
		FeeType:      billing.FeeTypePlatform,
		Amount:       sdk.NewCoins(sdk.NewInt64Coin("uvirt", 25)),
		Destination:  "virtengine1treasury...",
		BlockHeight:  100,
		Timestamp:    now,
		Status:       billing.TreasuryAllocationStatusPending,
	}

	err := allocation.Validate()
	s.Require().NoError(err)

	// Test invalid allocation (missing invoice ID)
	invalidAlloc := &billing.TreasuryAllocation{
		AllocationID: "alloc-002",
		SettlementID: "settlement-001",
		FeeType:      billing.FeeTypePlatform,
		Amount:       sdk.NewCoins(sdk.NewInt64Coin("uvirt", 25)),
		BlockHeight:  100,
		Timestamp:    now,
	}

	err = invalidAlloc.Validate()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "invoice_id is required")
}

// TestTreasurySummary tests treasury summary aggregation
func (s *SettlementIntegrationTestSuite) TestTreasurySummary() {
	now := time.Now().UTC()
	periodStart := now.Add(-24 * time.Hour)
	periodEnd := now

	summary := billing.NewTreasurySummary(periodStart, periodEnd, 100, now)

	s.Require().Equal(uint32(0), summary.TotalSettlements)
	s.Require().True(summary.TotalGrossVolume.Empty())
	s.Require().True(summary.TotalFees.Empty())

	// Add some allocations
	summary.TotalSettlements = 5
	summary.TotalGrossVolume = sdk.NewCoins(sdk.NewInt64Coin("uvirt", 10000))
	summary.TotalFees = sdk.NewCoins(sdk.NewInt64Coin("uvirt", 800))
	summary.TotalPayouts = sdk.NewCoins(sdk.NewInt64Coin("uvirt", 9200))
	summary.TotalPlatformFees = sdk.NewCoins(sdk.NewInt64Coin("uvirt", 250))
	summary.TotalNetworkFees = sdk.NewCoins(sdk.NewInt64Coin("uvirt", 50))
	summary.TotalCommunityFees = sdk.NewCoins(sdk.NewInt64Coin("uvirt", 100))
	summary.TotalTakeFees = sdk.NewCoins(sdk.NewInt64Coin("uvirt", 400))

	s.Require().Equal(uint32(5), summary.TotalSettlements)
	s.Require().True(summary.TotalPlatformFees.AmountOf("uvirt").Equal(sdkmath.NewInt(250)))
}

// TestSettlementStoreKeys tests the settlement store key builders
func (s *SettlementIntegrationTestSuite) TestSettlementStoreKeys() {
	settlementID := "settlement-001"
	invoiceID := "invoice-001"
	provider := "provider-001"

	// Test settlement record key
	key := billing.BuildSettlementRecordKey(settlementID)
	s.Require().NotEmpty(key)
	s.Require().Contains(string(key), settlementID)

	// Test settlement by invoice key
	byInvoiceKey := billing.BuildSettlementByInvoiceKey(invoiceID, settlementID)
	s.Require().NotEmpty(byInvoiceKey)

	// Test settlement by provider prefix
	byProviderPrefix := billing.BuildSettlementByProviderPrefix(provider)
	s.Require().NotEmpty(byProviderPrefix)

	// Test treasury allocation key
	allocKey := billing.BuildTreasuryAllocationKey("alloc-001")
	s.Require().NotEmpty(allocKey)
}

// TestSettlementComplete tests completing a settlement
func (s *SettlementIntegrationTestSuite) TestSettlementComplete() {
	now := time.Now().UTC()
	config := billing.DefaultFeeConfig()
	grossAmount := sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000))
	breakdown := config.CalculateFees(grossAmount)

	// Use valid bech32 addresses for testing
	provider := sdk.AccAddress(make([]byte, 20)).String()
	customer := sdk.AccAddress(make([]byte, 20)).String()

	record := billing.NewSettlementRecord(
		"settlement-003",
		"invoice-003",
		"escrow-003",
		provider,
		customer,
		grossAmount,
		breakdown,
		100,
		now,
	)

	// Complete the settlement
	err := record.Complete(now.Add(time.Hour))
	s.Require().NoError(err)
	s.Require().Equal(billing.SettlementStatusCompleted, record.Status)
	s.Require().NotNil(record.SettledAt)

	// Should not be able to complete again
	err = record.Complete(now.Add(2 * time.Hour))
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "terminal state")
}
