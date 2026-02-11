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
		UsageID:           "usage-1",
		OrderID:           "order-settle",
		Provider:          s.provider.String(),
		Customer:          s.depositor.String(),
		UsageUnits:        1000,
		UsageType:         "compute",
		TotalCost:         sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(500))),
		PeriodStart:       s.ctx.BlockTime().Add(-time.Hour),
		PeriodEnd:         s.ctx.BlockTime(),
		SubmittedAt:       s.ctx.BlockTime(),
		ProviderSignature: []byte("provider-signature"),
	}
	err = s.keeper.RecordUsage(s.ctx, usage)
	s.Require().NoError(err)

	// Settle the order
	settlement, err := s.keeper.SettleOrder(s.ctx, "order-settle", []string{"usage-1"}, false)
	s.Require().NoError(err)
	s.Require().NotNil(settlement)
	s.Require().Equal("order-settle", settlement.OrderID)
	s.Require().Equal(types.SettlementTypeUsageBased, settlement.SettlementType)

	// Verify usage is marked as settled
	updatedUsage, found := s.keeper.GetUsageRecord(s.ctx, "usage-1")
	s.Require().True(found)
	s.Require().True(updatedUsage.Settled)

	params := s.keeper.GetParams(s.ctx)
	height := s.ctx.BlockHeight()
	if height < 0 {
		height = 0
	}
	epoch := uint64(height) / params.StakingRewardEpochLength //nolint:gosec // non-negative height checked above
	distributions := s.keeper.GetRewardsByEpoch(s.ctx, epoch)

	foundUsageRewards := false
	for _, dist := range distributions {
		if dist.Source == types.RewardSourceUsage {
			foundUsageRewards = true
			break
		}
	}
	s.Require().True(foundUsageRewards, "expected usage rewards distribution")
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
	s.Require().Equal(types.SettlementTypeFinal, settlement.SettlementType)

	// Verify escrow is released
	escrow, found := s.keeper.GetEscrow(s.ctx, escrowID)
	s.Require().True(found)
	s.Require().Equal(types.EscrowStateReleased, escrow.State)
}

func (s *KeeperTestSuite) TestRecordUsage() {
	// Set up: Create and activate an escrow for valid test cases
	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10000)))
	escrowID, err := s.keeper.CreateEscrow(s.ctx, "order-usage-1", s.depositor, amount, time.Hour*24, nil)
	s.Require().NoError(err)
	err = s.keeper.ActivateEscrow(s.ctx, escrowID, "lease-usage-1", s.provider)
	s.Require().NoError(err)

	testCases := []struct {
		name        string
		usage       *types.UsageRecord
		expectError bool
	}{
		{
			name: "valid usage record",
			usage: &types.UsageRecord{
				OrderID:           "order-usage-1",
				Provider:          s.provider.String(),
				Customer:          s.depositor.String(),
				UsageUnits:        1000,
				UsageType:         "compute",
				TotalCost:         sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(500))),
				PeriodStart:       s.ctx.BlockTime().Add(-time.Hour),
				PeriodEnd:         s.ctx.BlockTime(),
				SubmittedAt:       s.ctx.BlockTime(),
				ProviderSignature: []byte("provider-signature-data"),
			},
			expectError: false,
		},
		{
			name: "empty order ID",
			usage: &types.UsageRecord{
				OrderID:           "",
				Provider:          s.provider.String(),
				Customer:          s.depositor.String(),
				UsageUnits:        1000,
				UsageType:         "compute",
				TotalCost:         sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(500))),
				ProviderSignature: []byte("provider-signature-data"),
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

func (s *KeeperTestSuite) TestDoubleSettlementPrevention() {
	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10000)))
	escrowID, err := s.keeper.CreateEscrow(s.ctx, "order-double", s.depositor, amount, time.Hour*24, nil)
	s.Require().NoError(err)
	s.Require().NoError(s.keeper.ActivateEscrow(s.ctx, escrowID, "lease-double", s.provider))

	usage := &types.UsageRecord{
		OrderID:           "order-double",
		Provider:          s.provider.String(),
		Customer:          s.depositor.String(),
		UsageUnits:        50,
		UsageType:         "compute",
		TotalCost:         sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(500))),
		PeriodStart:       s.ctx.BlockTime().Add(-time.Hour),
		PeriodEnd:         s.ctx.BlockTime(),
		ProviderSignature: []byte("provider-signature"),
	}
	s.Require().NoError(s.keeper.RecordUsage(s.ctx, usage))

	_, err = s.keeper.SettleOrder(s.ctx, "order-double", []string{usage.UsageID}, false)
	s.Require().NoError(err)

	_, err = s.keeper.SettleOrder(s.ctx, "order-double", []string{usage.UsageID}, false)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestOverRefundBlocked() {
	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000)))
	escrowID, err := s.keeper.CreateEscrow(s.ctx, "order-refund-block", s.depositor, amount, time.Hour*24, nil)
	s.Require().NoError(err)
	s.Require().NoError(s.keeper.ActivateEscrow(s.ctx, escrowID, "lease-refund-block", s.provider))

	_, err = s.keeper.SettleOrder(s.ctx, "order-refund-block", nil, true)
	s.Require().NoError(err)

	err = s.keeper.RefundEscrow(s.ctx, escrowID, "refund-after-release")
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestAcknowledgeUsage() {
	// Set up: Create and activate an escrow
	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10000)))
	escrowID, err := s.keeper.CreateEscrow(s.ctx, "order-ack", s.depositor, amount, time.Hour*24, nil)
	s.Require().NoError(err)
	err = s.keeper.ActivateEscrow(s.ctx, escrowID, "lease-ack", s.provider)
	s.Require().NoError(err)

	// Record usage first
	usage := &types.UsageRecord{
		OrderID:           "order-ack",
		Provider:          s.provider.String(),
		Customer:          s.depositor.String(),
		UsageUnits:        1000,
		UsageType:         "compute",
		TotalCost:         sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(500))),
		PeriodStart:       s.ctx.BlockTime().Add(-time.Hour),
		PeriodEnd:         s.ctx.BlockTime(),
		SubmittedAt:       s.ctx.BlockTime(),
		ProviderSignature: []byte("provider-signature"),
	}
	err = s.keeper.RecordUsage(s.ctx, usage)
	s.Require().NoError(err)

	// Acknowledge usage
	signature := []byte("customer_signature")
	err = s.keeper.AcknowledgeUsage(s.ctx, usage.UsageID, signature)
	s.Require().NoError(err)

	// Verify acknowledgment
	updatedUsage, found := s.keeper.GetUsageRecord(s.ctx, usage.UsageID)
	s.Require().True(found)
	s.Require().True(updatedUsage.CustomerAcknowledged)
	s.Require().Equal(signature, updatedUsage.CustomerSignature)
}

func (s *KeeperTestSuite) TestGetSettlementsByOrder() {
	// Create and activate an escrow
	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10000)))
	escrowID, err := s.keeper.CreateEscrow(s.ctx, "order-multi-settle", s.depositor, amount, time.Hour*24, nil)
	s.Require().NoError(err)
	err = s.keeper.ActivateEscrow(s.ctx, escrowID, "lease-multi-settle", s.provider)
	s.Require().NoError(err)

	// Record and settle multiple usage records
	for i := 0; i < 3; i++ {
		usage := &types.UsageRecord{
			OrderID:           "order-multi-settle",
			Provider:          s.provider.String(),
			Customer:          s.depositor.String(),
			UsageUnits:        1000,
			UsageType:         "compute",
			TotalCost:         sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100))),
			PeriodStart:       s.ctx.BlockTime().Add(-time.Hour),
			PeriodEnd:         s.ctx.BlockTime(),
			SubmittedAt:       s.ctx.BlockTime(),
			ProviderSignature: []byte("provider-signature"),
		}
		err := s.keeper.RecordUsage(s.ctx, usage)
		s.Require().NoError(err)
	}

	// Get all settlements for the order
	settlements := s.keeper.GetSettlementsByOrder(s.ctx, "order-multi-settle")
	// Initially no settlements
	s.Require().Len(settlements, 0)
}

func (s *KeeperTestSuite) TestBuildUsageSummary() {
	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10000)))
	escrowID, err := s.keeper.CreateEscrow(s.ctx, "order-summary", s.depositor, amount, time.Hour*24, nil)
	s.Require().NoError(err)
	err = s.keeper.ActivateEscrow(s.ctx, escrowID, "lease-summary", s.provider)
	s.Require().NoError(err)

	now := s.ctx.BlockTime()
	usages := []types.UsageRecord{
		{
			OrderID:           "order-summary",
			Provider:          s.provider.String(),
			Customer:          s.depositor.String(),
			UsageUnits:        100,
			UsageType:         "cpu",
			TotalCost:         sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(200))),
			PeriodStart:       now.Add(-2 * time.Hour),
			PeriodEnd:         now.Add(-time.Hour),
			SubmittedAt:       now.Add(-time.Hour),
			ProviderSignature: []byte("sig-1"),
		},
		{
			OrderID:           "order-summary",
			Provider:          s.provider.String(),
			Customer:          s.depositor.String(),
			UsageUnits:        50,
			UsageType:         "gpu",
			TotalCost:         sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(300))),
			PeriodStart:       now.Add(-time.Hour),
			PeriodEnd:         now,
			SubmittedAt:       now,
			ProviderSignature: []byte("sig-2"),
		},
	}

	for _, usage := range usages {
		u := usage
		err := s.keeper.RecordUsage(s.ctx, &u)
		s.Require().NoError(err)
	}

	summary, err := s.keeper.BuildUsageSummary(s.ctx, "order-summary", s.provider.String(), time.Time{}, time.Time{})
	s.Require().NoError(err)
	s.Require().Equal(uint64(150), summary.TotalUsage)
	s.Require().Equal(sdkmath.NewInt(500), summary.TotalCost.AmountOf("uve"))
	s.Require().Len(summary.ByUsageType, 2)
}

func TestSettlementValidation(t *testing.T) {
	validProvider := sdk.AccAddress([]byte("test_provider_______")).String()
	validCustomer := sdk.AccAddress([]byte("test_customer_______")).String()

	testCases := []struct {
		name        string
		settlement  types.SettlementRecord
		expectError bool
	}{
		{
			name: "valid settlement",
			settlement: types.SettlementRecord{
				SettlementID:   "settlement-1",
				OrderID:        "order-1",
				EscrowID:       "escrow-1",
				Provider:       validProvider,
				Customer:       validCustomer,
				TotalAmount:    sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
				ProviderShare:  sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(950))),
				PlatformFee:    sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(50))),
				SettlementType: types.SettlementTypePeriodic,
				SettledAt:      time.Now(),
			},
			expectError: false,
		},
		{
			name: "empty settlement ID",
			settlement: types.SettlementRecord{
				SettlementID:   "",
				OrderID:        "order-1",
				EscrowID:       "escrow-1",
				TotalAmount:    sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
				ProviderShare:  sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(950))),
				PlatformFee:    sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(50))),
				SettlementType: types.SettlementTypePeriodic,
			},
			expectError: true,
		},
		{
			name: "invalid type",
			settlement: types.SettlementRecord{
				SettlementID:   "settlement-1",
				OrderID:        "order-1",
				EscrowID:       "escrow-1",
				TotalAmount:    sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
				ProviderShare:  sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(950))),
				PlatformFee:    sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(50))),
				SettlementType: types.SettlementType("invalid"),
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
	validProvider := sdk.AccAddress([]byte("test_provider_______")).String()
	validCustomer := sdk.AccAddress([]byte("test_customer_______")).String()

	testCases := []struct {
		name        string
		usage       types.UsageRecord
		expectError bool
	}{
		{
			name: "valid usage record",
			usage: types.UsageRecord{
				UsageID:           "usage-1",
				OrderID:           "order-1",
				Provider:          validProvider,
				Customer:          validCustomer,
				UsageUnits:        1000,
				UsageType:         "compute",
				TotalCost:         sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100))),
				PeriodStart:       time.Now().Add(-time.Hour),
				PeriodEnd:         time.Now(),
				ProviderSignature: []byte("provider-signature"),
			},
			expectError: false,
		},
		{
			name: "empty usage ID",
			usage: types.UsageRecord{
				UsageID:    "",
				OrderID:    "order-1",
				Provider:   validProvider,
				UsageUnits: 1000,
				TotalCost:  sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100))),
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
