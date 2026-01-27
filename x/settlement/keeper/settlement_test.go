//go:build ignore
// +build ignore

// TODO: This test file is excluded until settlement API is stabilized.

package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/settlement/types"
)

func (s *KeeperTestSuite) TestSettleOrder() {
	// Create and activate an escrow
	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10000)))
	escrowID, err := s.keeper.CreateEscrow(s.ctx, "order-settle", s.depositor, amount, time.Hour*24, nil)
	s.Require().NoError(err)

	err = s.keeper.ActivateEscrow(s.ctx, escrowID, "lease-settle", s.provider)
	s.Require().NoError(err)

	// Record some usage
	usage := &types.UsageRecord{
		UsageID:     "usage-1",
		OrderID:     "order-settle",
		Provider:    s.provider.String(),
		Customer:    s.depositor.String(),
		ComputeUsed: "1000",
		StorageUsed: "500",
		TotalCost:   sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(500))),
		PeriodStart: s.ctx.BlockTime().Add(-time.Hour),
		PeriodEnd:   s.ctx.BlockTime(),
		RecordedAt:  s.ctx.BlockTime(),
	}
	err = s.keeper.RecordUsage(s.ctx, usage)
	s.Require().NoError(err)

	// Settle the order
	settlement, err := s.keeper.SettleOrder(s.ctx, "order-settle", []string{"usage-1"}, false)
	s.Require().NoError(err)
	s.Require().NotNil(settlement)
	s.Require().Equal("order-settle", settlement.OrderID)
	s.Require().Equal(types.SettlementTypeUsageBased, settlement.Type)

	// Verify usage is marked as settled
	updatedUsage, found := s.keeper.GetUsageRecord(s.ctx, "usage-1")
	s.Require().True(found)
	s.Require().True(updatedUsage.Settled)
}

func (s *KeeperTestSuite) TestFinalSettlement() {
	// Create and activate an escrow
	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10000)))
	escrowID, err := s.keeper.CreateEscrow(s.ctx, "order-final", s.depositor, amount, time.Hour*24, nil)
	s.Require().NoError(err)

	err = s.keeper.ActivateEscrow(s.ctx, escrowID, "lease-final", s.provider)
	s.Require().NoError(err)

	// Settle as final
	settlement, err := s.keeper.SettleOrder(s.ctx, "order-final", nil, true)
	s.Require().NoError(err)
	s.Require().NotNil(settlement)
	s.Require().Equal(types.SettlementTypeFinal, settlement.Type)

	// Verify escrow is released
	escrow, found := s.keeper.GetEscrow(s.ctx, escrowID)
	s.Require().True(found)
	s.Require().Equal(types.EscrowStateReleased, escrow.State)
}

func (s *KeeperTestSuite) TestRecordUsage() {
	testCases := []struct {
		name        string
		usage       *types.UsageRecord
		expectError bool
	}{
		{
			name: "valid usage record",
			usage: &types.UsageRecord{
				OrderID:     "order-usage-1",
				Provider:    s.provider.String(),
				Customer:    s.depositor.String(),
				ComputeUsed: "1000",
				StorageUsed: "500",
				TotalCost:   sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(500))),
				PeriodStart: time.Now().Add(-time.Hour),
				PeriodEnd:   time.Now(),
				RecordedAt:  time.Now(),
			},
			expectError: false,
		},
		{
			name: "empty order ID",
			usage: &types.UsageRecord{
				OrderID:     "",
				Provider:    s.provider.String(),
				Customer:    s.depositor.String(),
				ComputeUsed: "1000",
				TotalCost:   sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(500))),
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := s.keeper.RecordUsage(s.ctx, tc.usage)
			if tc.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(tc.usage.UsageID)

				// Verify usage was stored
				usage, found := s.keeper.GetUsageRecord(s.ctx, tc.usage.UsageID)
				s.Require().True(found)
				s.Require().Equal(tc.usage.OrderID, usage.OrderID)
			}
		})
	}
}

func (s *KeeperTestSuite) TestAcknowledgeUsage() {
	// Record usage first
	usage := &types.UsageRecord{
		OrderID:     "order-ack",
		Provider:    s.provider.String(),
		Customer:    s.depositor.String(),
		ComputeUsed: "1000",
		TotalCost:   sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(500))),
		PeriodStart: s.ctx.BlockTime().Add(-time.Hour),
		PeriodEnd:   s.ctx.BlockTime(),
		RecordedAt:  s.ctx.BlockTime(),
	}
	err := s.keeper.RecordUsage(s.ctx, usage)
	s.Require().NoError(err)

	// Acknowledge usage
	signature := []byte("customer_signature")
	err = s.keeper.AcknowledgeUsage(s.ctx, usage.UsageID, signature)
	s.Require().NoError(err)

	// Verify acknowledgment
	updatedUsage, found := s.keeper.GetUsageRecord(s.ctx, usage.UsageID)
	s.Require().True(found)
	s.Require().True(updatedUsage.Acknowledged)
	s.Require().Equal(signature, updatedUsage.CustomerSignature)
}

func (s *KeeperTestSuite) TestGetSettlementsByOrder() {
	// Create and activate an escrow
	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10000)))
	_, err := s.keeper.CreateEscrow(s.ctx, "order-multi-settle", s.depositor, amount, time.Hour*24, nil)
	s.Require().NoError(err)

	// Record and settle multiple usage records
	for i := 0; i < 3; i++ {
		usage := &types.UsageRecord{
			OrderID:     "order-multi-settle",
			Provider:    s.provider.String(),
			Customer:    s.depositor.String(),
			ComputeUsed: "1000",
			TotalCost:   sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100))),
			PeriodStart: s.ctx.BlockTime().Add(-time.Hour),
			PeriodEnd:   s.ctx.BlockTime(),
			RecordedAt:  s.ctx.BlockTime(),
		}
		err := s.keeper.RecordUsage(s.ctx, usage)
		s.Require().NoError(err)
	}

	// Get all settlements for the order
	settlements := s.keeper.GetSettlementsByOrder(s.ctx, "order-multi-settle")
	// Initially no settlements
	s.Require().Len(settlements, 0)
}

func TestSettlementValidation(t *testing.T) {
	testCases := []struct {
		name        string
		settlement  types.SettlementRecord
		expectError bool
	}{
		{
			name: "valid settlement",
			settlement: types.SettlementRecord{
				SettlementID: "settlement-1",
				OrderID:      "order-1",
				EscrowID:     "escrow-1",
				Provider:     "cosmos1provider...",
				Customer:     "cosmos1customer...",
				Amount:       sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
				Type:         types.SettlementTypePeriodic,
				SettledAt:    time.Now(),
			},
			expectError: false,
		},
		{
			name: "empty settlement ID",
			settlement: types.SettlementRecord{
				SettlementID: "",
				OrderID:      "order-1",
				EscrowID:     "escrow-1",
				Amount:       sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
				Type:         types.SettlementTypePeriodic,
			},
			expectError: true,
		},
		{
			name: "invalid type",
			settlement: types.SettlementRecord{
				SettlementID: "settlement-1",
				OrderID:      "order-1",
				EscrowID:     "escrow-1",
				Amount:       sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
				Type:         types.SettlementType("invalid"),
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.settlement.Validate()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUsageRecordValidation(t *testing.T) {
	testCases := []struct {
		name        string
		usage       types.UsageRecord
		expectError bool
	}{
		{
			name: "valid usage record",
			usage: types.UsageRecord{
				UsageID:     "usage-1",
				OrderID:     "order-1",
				Provider:    "cosmos1provider...",
				Customer:    "cosmos1customer...",
				ComputeUsed: "1000",
				TotalCost:   sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100))),
				PeriodStart: time.Now().Add(-time.Hour),
				PeriodEnd:   time.Now(),
			},
			expectError: false,
		},
		{
			name: "empty usage ID",
			usage: types.UsageRecord{
				UsageID:     "",
				OrderID:     "order-1",
				Provider:    "cosmos1provider...",
				ComputeUsed: "1000",
				TotalCost:   sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100))),
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.usage.Validate()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
