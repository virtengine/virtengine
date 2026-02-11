package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/settlement/keeper"
	"github.com/virtengine/virtengine/x/settlement/types"
)

func TestEscrowSettlementInvariant(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000)))
	escrowID, err := suite.keeper.CreateEscrow(suite.ctx, "order-inv", suite.depositor, amount, time.Hour, nil)
	require.NoError(t, err)
	require.NoError(t, suite.keeper.ActivateEscrow(suite.ctx, escrowID, "lease-inv", suite.provider))

	usage := &types.UsageRecord{
		OrderID:           "order-inv",
		Provider:          suite.provider.String(),
		Customer:          suite.depositor.String(),
		UsageUnits:        10,
		UsageType:         "compute",
		TotalCost:         sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100))),
		PeriodStart:       suite.ctx.BlockTime().Add(-time.Hour),
		PeriodEnd:         suite.ctx.BlockTime(),
		ProviderSignature: []byte("sig"),
	}
	require.NoError(t, suite.keeper.RecordUsage(suite.ctx, usage))

	_, err = suite.keeper.SettleOrder(suite.ctx, "order-inv", []string{usage.UsageID}, false)
	require.NoError(t, err)

	invariant := keeper.EscrowSettlementReconciliationInvariant(suite.keeper)
	msg, broken := invariant(suite.ctx)
	require.False(t, broken, msg)
}

func TestEscrowSettlementInvariantDetectsMismatch(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000)))
	escrowID, err := suite.keeper.CreateEscrow(suite.ctx, "order-inv-bad", suite.depositor, amount, time.Hour, nil)
	require.NoError(t, err)
	require.NoError(t, suite.keeper.ActivateEscrow(suite.ctx, escrowID, "lease-inv-bad", suite.provider))

	escrow, found := suite.keeper.GetEscrow(suite.ctx, escrowID)
	require.True(t, found)
	escrow.TotalSettled = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(500)))
	require.NoError(t, suite.keeper.SetEscrow(suite.ctx, escrow))

	invariant := keeper.EscrowSettlementReconciliationInvariant(suite.keeper)
	_, broken := invariant(suite.ctx)
	require.True(t, broken)
}

func TestEscrowInvariantMultipleOrders(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(2000)))
	settledEscrowID, err := suite.keeper.CreateEscrow(suite.ctx, "order-multi-1", suite.depositor, amount, time.Hour, nil)
	require.NoError(t, err)
	require.NoError(t, suite.keeper.ActivateEscrow(suite.ctx, settledEscrowID, "lease-multi-1", suite.provider))

	usage := &types.UsageRecord{
		OrderID:           "order-multi-1",
		Provider:          suite.provider.String(),
		Customer:          suite.depositor.String(),
		UsageUnits:        20,
		UsageType:         "compute",
		TotalCost:         sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(500))),
		PeriodStart:       suite.ctx.BlockTime().Add(-time.Hour),
		PeriodEnd:         suite.ctx.BlockTime(),
		ProviderSignature: []byte("sig"),
	}
	require.NoError(t, suite.keeper.RecordUsage(suite.ctx, usage))
	_, err = suite.keeper.SettleOrder(suite.ctx, "order-multi-1", []string{usage.UsageID}, false)
	require.NoError(t, err)

	refundEscrowID, err := suite.keeper.CreateEscrow(suite.ctx, "order-multi-2", suite.depositor, amount, time.Hour, nil)
	require.NoError(t, err)
	err = suite.keeper.RefundEscrow(suite.ctx, refundEscrowID, "customer refund")
	require.NoError(t, err)

	invariant := keeper.EscrowSettlementReconciliationInvariant(suite.keeper)
	msg, broken := invariant(suite.ctx)
	require.False(t, broken, msg)
}
