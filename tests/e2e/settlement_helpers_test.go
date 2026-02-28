//go:build e2e.integration

package e2e

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/app"
	escrowkeeper "github.com/virtengine/virtengine/x/escrow/keeper"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
)

type e2eInvoiceKeeperSource interface {
	NewInvoiceKeeper() escrowkeeper.InvoiceKeeper
}

func fundE2EAccount(t *testing.T, app *app.VirtEngineApp, ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) {
	t.Helper()

	bank := app.Keepers.Cosmos.Bank
	require.NoError(t, bank.MintCoins(ctx, minttypes.ModuleName, coins))
	require.NoError(t, bank.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, coins))
}

func createActiveSettlementEscrowE2E(
	t *testing.T,
	app *app.VirtEngineApp,
	ctx sdk.Context,
	orderID string,
	leaseID string,
	customer sdk.AccAddress,
	provider sdk.AccAddress,
	amount sdk.Coins,
	expiresIn time.Duration,
) string {
	t.Helper()

	keeper := &app.Keepers.VirtEngine.Settlement
	escrowID, err := keeper.CreateEscrow(ctx, orderID, customer, amount, expiresIn, nil)
	require.NoError(t, err)
	require.NoError(t, keeper.ActivateEscrow(ctx, escrowID, leaseID, provider))

	return escrowID
}

func recordSettlementUsageE2E(
	t *testing.T,
	app *app.VirtEngineApp,
	ctx sdk.Context,
	orderID string,
	leaseID string,
	provider sdk.AccAddress,
	customer sdk.AccAddress,
	usageType string,
	usageUnits uint64,
	totalCost sdk.Coins,
	periodStart time.Time,
	periodEnd time.Time,
) *settlementtypes.UsageRecord {
	t.Helper()

	record := &settlementtypes.UsageRecord{
		OrderID:           orderID,
		LeaseID:           leaseID,
		Provider:          provider.String(),
		Customer:          customer.String(),
		UsageUnits:        usageUnits,
		UsageType:         usageType,
		TotalCost:         totalCost,
		PeriodStart:       periodStart,
		PeriodEnd:         periodEnd,
		SubmittedAt:       ctx.BlockTime(),
		ProviderSignature: []byte("e2e-settlement-signature"),
	}
	require.NoError(t, app.Keepers.VirtEngine.Settlement.RecordUsage(ctx, record))

	return record
}

func requireSettlementSplit(t *testing.T, settlement *settlementtypes.SettlementRecord) {
	t.Helper()

	require.True(
		t,
		settlement.ProviderShare.Add(settlement.PlatformFee...).Add(settlement.ValidatorFee...).Equal(settlement.TotalAmount),
		"expected split %s + %s + %s to equal %s",
		settlement.ProviderShare.String(),
		settlement.PlatformFee.String(),
		settlement.ValidatorFee.String(),
		settlement.TotalAmount.String(),
	)
}

func requireE2EInvoiceKeeper(t *testing.T, app *app.VirtEngineApp) escrowkeeper.InvoiceKeeper {
	t.Helper()

	source, ok := app.Keepers.VirtEngine.Escrow.(e2eInvoiceKeeperSource)
	require.True(t, ok, "escrow keeper must expose invoice keeper integration")

	return source.NewInvoiceKeeper()
}
