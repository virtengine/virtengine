package testutil

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/testutil/state"
	escrowkeeper "github.com/virtengine/virtengine/x/escrow/keeper"
	"github.com/virtengine/virtengine/x/escrow/types/billing"
)

// SetupEscrowKeeper creates a test escrow keeper with necessary dependencies using the existing test suite
func SetupEscrowKeeper(t *testing.T) (sdk.Context, escrowkeeper.Keeper) {
	t.Helper()

	suite := state.SetupTestSuite(t)
	return suite.Context(), suite.EscrowKeeper()
}

// AccAddress generates a test account address
func AccAddress(t *testing.T) sdk.AccAddress {
	t.Helper()
	return sdk.AccAddress("test_" + t.Name() + "_addr")
}

// CreateTestUsageRecord creates a test usage record
func CreateTestUsageRecord(t *testing.T, leaseID, provider, customer string, amount int64) *billing.UsageRecord {
	t.Helper()

	now := time.Now().UTC()
	return &billing.UsageRecord{
		RecordID:     "usage-" + leaseID + "-1",
		LeaseID:      leaseID,
		Provider:     provider,
		Customer:     customer,
		StartTime:    now.Add(-24 * time.Hour),
		EndTime:      now,
		ResourceType: billing.UsageTypeCPU,
		UsageAmount:  sdkmath.LegacyNewDec(100),
		UnitPrice:    sdk.NewDecCoin("uakt", sdkmath.NewInt(10)),
		TotalAmount:  sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(amount))),
		Status:       billing.UsageRecordStatusPending,
		BlockHeight:  1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
