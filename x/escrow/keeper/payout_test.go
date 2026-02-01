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

// PayoutTestSuite tests the payout keeper functionality
type PayoutTestSuite struct {
	suite.Suite
}

func TestPayoutTestSuite(t *testing.T) {
	suite.Run(t, new(PayoutTestSuite))
}

// TestPayoutRecordCreation tests creating a payout record
func (s *PayoutTestSuite) TestPayoutRecordCreation() {
	now := time.Now().UTC()
	provider := "virtengine1provider..."

	record := billing.NewPayoutRecord(
		"payout-001",
		"settlement-001",
		provider,
		sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)), // gross
		sdk.NewCoins(sdk.NewInt64Coin("uvirt", 80)),   // fees
		sdk.NewCoins(sdk.NewInt64Coin("uvirt", 920)),  // net
		sdk.NewCoins(sdk.NewInt64Coin("uvirt", 920)),  // payout amount
		100,
		now,
	)

	s.Require().NotNil(record)
	s.Require().Equal("payout-001", record.PayoutID)
	s.Require().Equal("settlement-001", record.SettlementID)
	s.Require().Equal(provider, record.Provider)
	s.Require().Equal(billing.PayoutStatusPending, record.Status)
	s.Require().True(record.GrossAmount.AmountOf("uvirt").Equal(sdkmath.NewInt(1000)))
	s.Require().True(record.FeeAmount.AmountOf("uvirt").Equal(sdkmath.NewInt(80)))
	s.Require().True(record.NetAmount.AmountOf("uvirt").Equal(sdkmath.NewInt(920)))
}

// TestPayoutRecordValidation tests payout record validation
func (s *PayoutTestSuite) TestPayoutRecordValidation() {
	now := time.Now().UTC()

	// Use valid bech32 addresses for testing
	validProvider := sdk.AccAddress(make([]byte, 20)).String()

	tests := []struct {
		name        string
		record      billing.PayoutRecord
		expectError bool
		errContains string
	}{
		{
			name: "valid record with settlement_id",
			record: billing.PayoutRecord{
				PayoutID:     "payout-001",
				SettlementID: "settlement-001",
				Provider:     validProvider,
				PayoutAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
				FeeAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 80)),
				NetAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 920)),
				PayoutDate:   now,
			},
			expectError: false,
		},
		{
			name: "valid record with invoice_ids",
			record: billing.PayoutRecord{
				PayoutID:     "payout-002",
				Provider:     validProvider,
				PayoutAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
				FeeAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 80)),
				NetAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 920)),
				InvoiceIDs:   []string{"invoice-001"},
				PayoutDate:   now,
			},
			expectError: false,
		},
		{
			name: "missing payout_id",
			record: billing.PayoutRecord{
				SettlementID: "settlement-001",
				Provider:     validProvider,
				PayoutAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
				FeeAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 80)),
				NetAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 920)),
				PayoutDate:   now,
			},
			expectError: true,
			errContains: "payout_id is required",
		},
		{
			name: "missing both invoice_ids and settlement_id",
			record: billing.PayoutRecord{
				PayoutID:     "payout-003",
				Provider:     validProvider,
				PayoutAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
				FeeAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 80)),
				NetAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 920)),
				PayoutDate:   now,
			},
			expectError: true,
			errContains: "at least one invoice_id or settlement_id is required",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := tt.record.Validate()
			if tt.expectError {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tt.errContains)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// TestPayoutStatusTransitions tests payout status values
func (s *PayoutTestSuite) TestPayoutStatusTransitions() {
	statuses := []billing.PayoutStatus{
		billing.PayoutStatusPending,
		billing.PayoutStatusProcessing,
		billing.PayoutStatusCompleted,
		billing.PayoutStatusFailed,
		billing.PayoutStatusCancelled,
		billing.PayoutStatusRefunded,
	}

	expectedNames := []string{
		"pending",
		"processing",
		"completed",
		"failed",
		"cancelled",
		"refunded",
	}

	for i, status := range statuses {
		s.Require().Equal(expectedNames[i], status.String())
	}
}

// TestPayoutCalculation tests payout calculation structure
func (s *PayoutTestSuite) TestPayoutCalculation() {
	now := time.Now().UTC()

	calc := billing.PayoutCalculation{
		SettlementID:   "settlement-001",
		Provider:       "provider-001",
		GrossAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
		TotalFees:      sdk.NewCoins(sdk.NewInt64Coin("uvirt", 80)),
		NetAmount:      sdk.NewCoins(sdk.NewInt64Coin("uvirt", 920)),
		HoldbackAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 100)),
		PayableAmount:  sdk.NewCoins(sdk.NewInt64Coin("uvirt", 820)),
		CalculatedAt:   now,
		BlockHeight:    100,
	}

	s.Require().Equal("settlement-001", calc.SettlementID)
	s.Require().True(calc.PayableAmount.AmountOf("uvirt").Equal(sdkmath.NewInt(820)))
}

// TestPayoutSummary tests payout summary aggregation
func (s *PayoutTestSuite) TestPayoutSummary() {
	now := time.Now().UTC()
	periodStart := now.Add(-24 * time.Hour)
	periodEnd := now

	summary := billing.NewPayoutSummary("provider-001", periodStart, periodEnd, 100, now)

	s.Require().NotNil(summary)
	s.Require().Equal("provider-001", summary.Provider)
	s.Require().Equal(uint32(0), summary.TotalPayouts)
	s.Require().True(summary.TotalGrossAmount.Empty())

	// Add a completed payout
	payout := &billing.PayoutRecord{
		PayoutID:     "payout-001",
		Provider:     "provider-001",
		GrossAmount:  sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
		PayoutAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 920)),
		FeeAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 80)),
		NetAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 920)),
		Status:       billing.PayoutStatusCompleted,
	}

	summary.AddPayout(payout)

	s.Require().Equal(uint32(1), summary.TotalPayouts)
	s.Require().Equal(uint32(1), summary.CompletedPayouts)
	s.Require().True(summary.TotalGrossAmount.AmountOf("uvirt").Equal(sdkmath.NewInt(1000)))
	s.Require().True(summary.CompletedAmount.AmountOf("uvirt").Equal(sdkmath.NewInt(920)))

	// Add a pending payout
	pendingPayout := &billing.PayoutRecord{
		PayoutID:     "payout-002",
		Provider:     "provider-001",
		GrossAmount:  sdk.NewCoins(sdk.NewInt64Coin("uvirt", 500)),
		PayoutAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 460)),
		FeeAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 40)),
		NetAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 460)),
		Status:       billing.PayoutStatusPending,
	}

	summary.AddPayout(pendingPayout)

	s.Require().Equal(uint32(2), summary.TotalPayouts)
	s.Require().Equal(uint32(1), summary.CompletedPayouts)
	s.Require().Equal(uint32(1), summary.PendingPayouts)
	s.Require().True(summary.PendingAmount.AmountOf("uvirt").Equal(sdkmath.NewInt(460)))
}

// TestPayoutStoreKeys tests payout store key builders
func (s *PayoutTestSuite) TestPayoutStoreKeys() {
	payoutID := "payout-001"
	provider := "provider-001"
	settlementID := "settlement-001"

	// Test payout record key
	key := billing.BuildPayoutRecordKey(payoutID)
	s.Require().NotEmpty(key)

	// Parse payout record key
	parsedID, err := billing.ParsePayoutRecordKey(key)
	s.Require().NoError(err)
	s.Require().Equal(payoutID, parsedID)

	// Test payout by provider key
	byProviderKey := billing.BuildPayoutRecordByProviderKey(provider, payoutID)
	s.Require().NotEmpty(byProviderKey)

	// Test payout by provider prefix
	byProviderPrefix := billing.BuildPayoutRecordByProviderPrefix(provider)
	s.Require().NotEmpty(byProviderPrefix)

	// Test payout by status key
	byStatusKey := billing.BuildPayoutRecordByStatusKey(billing.PayoutStatusPending, payoutID)
	s.Require().NotEmpty(byStatusKey)

	// Test payout by status prefix
	byStatusPrefix := billing.BuildPayoutRecordByStatusPrefix(billing.PayoutStatusPending)
	s.Require().NotEmpty(byStatusPrefix)

	// Test payout by settlement key
	bySettlementKey := billing.BuildPayoutRecordBySettlementKey(settlementID, payoutID)
	s.Require().NotEmpty(bySettlementKey)

	// Test payout by settlement prefix
	bySettlementPrefix := billing.BuildPayoutRecordBySettlementPrefix(settlementID)
	s.Require().NotEmpty(bySettlementPrefix)
}

// TestNextPayoutID tests payout ID generation
func (s *PayoutTestSuite) TestNextPayoutID() {
	id1 := billing.NextPayoutID(0, "PROV")
	s.Require().Equal("PROV-PAY-00000001", id1)

	id2 := billing.NextPayoutID(99, "PROV")
	s.Require().Equal("PROV-PAY-00000100", id2)

	id3 := billing.NextPayoutID(12345678, "TEST")
	s.Require().Equal("TEST-PAY-12345679", id3)
}
