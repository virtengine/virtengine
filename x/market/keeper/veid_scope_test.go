package keeper_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"

	"github.com/virtengine/virtengine/x/market/keeper"
)

// stubVEIDKeeper is a minimal VEIDKeeper used to drive gating outcomes deterministically.
type stubVEIDKeeper struct {
	score  uint32
	record veidtypes.IdentityRecord
	found  bool
	scopes []veidtypes.IdentityScope
}

func (s stubVEIDKeeper) GetVEIDScore(_ sdk.Context, _ sdk.AccAddress) (uint32, bool) {
	return s.score, s.found
}

func (s stubVEIDKeeper) GetIdentityRecord(_ sdk.Context, _ sdk.AccAddress) (veidtypes.IdentityRecord, bool) {
	return s.record, s.found
}

func (s stubVEIDKeeper) GetScopesByType(_ sdk.Context, _ sdk.AccAddress, _ veidtypes.ScopeType) []veidtypes.IdentityScope {
	return s.scopes
}

// TestFailedVEIDCheckNeverSetsGlobalAccountState is DONE-WHEN criterion 2: a failed VEID
// check must never set a global account state by itself.
//
// The identity signal is an input to a listing/order requirement. It is not an account
// verdict, so no matter how the check fails (locked, unverified, absent, low score) the
// roles-module account state must be left exactly as it was.
func TestFailedVEIDCheckNeverSetsGlobalAccountState(t *testing.T) {
	customer := sdk.AccAddress(bytes.Repeat([]byte{3}, 20))

	baseRecord := func() veidtypes.IdentityRecord {
		rec := veidtypes.NewIdentityRecord(customer.String(), sdk.Context{}.BlockTime())
		return *rec
	}

	strict := keeper.VEIDGatingRequirements{
		MinCustomerScore:        70,
		RequireVerifiedStatus:   true,
		RequireUnlockedIdentity: true,
	}

	tests := []struct {
		name   string
		keeper stubVEIDKeeper
	}{
		{
			name: "no identity record at all",
			keeper: stubVEIDKeeper{
				found: false,
			},
		},
		{
			name: "identity locked",
			keeper: func() stubVEIDKeeper {
				rec := baseRecord()
				rec.Locked = true
				rec.LockedReason = "under review"
				return stubVEIDKeeper{found: true, score: 90, record: rec}
			}(),
		},
		{
			name: "identity unverified with zero score",
			keeper: func() stubVEIDKeeper {
				rec := baseRecord()
				return stubVEIDKeeper{found: true, score: 0, record: rec}
			}(),
		},
		{
			name: "score below the required minimum",
			keeper: func() stubVEIDKeeper {
				rec := baseRecord()
				return stubVEIDKeeper{found: true, score: 25, record: rec}
			}(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, k, suite := setupKeeper(t)
			rolesKeeper := suite.App().Keepers.VirtEngine.Roles

			// Baseline: no global account state has been written, and the account is
			// operational (absence of a record means operational).
			_, foundBefore := rolesKeeper.GetAccountState(ctx, customer)
			require.False(t, foundBefore, "fixture should start with no global account state")
			require.True(t, rolesKeeper.IsAccountOperational(ctx, customer))

			k.SetVEIDKeeper(tc.keeper)

			result, err := k.CheckVEIDGating(ctx, customer, strict)
			require.Error(t, err, "gating should fail for %s", tc.name)
			require.False(t, result.Passed)

			// The failed check must not have reached global account state.
			_, foundAfter := rolesKeeper.GetAccountState(ctx, customer)
			require.False(t, foundAfter,
				"a failed VEID check must not write global account state (%s)", tc.name)
			require.True(t, rolesKeeper.IsAccountOperational(ctx, customer),
				"a failed VEID check must not suspend the account (%s)", tc.name)
		})
	}
}

// TestPassingVEIDCheckAlsoLeavesAccountStateUntouched proves the symmetric case: even a
// passing identity check is not a reason to write global account state, so a marketplace
// identity signal can never be the thing that switches an account on or off.
func TestPassingVEIDCheckAlsoLeavesAccountStateUntouched(t *testing.T) {
	ctx, k, suite := setupKeeper(t)
	rolesKeeper := suite.App().Keepers.VirtEngine.Roles

	customer := sdk.AccAddress(bytes.Repeat([]byte{4}, 20))
	require.True(t, rolesKeeper.IsAccountOperational(ctx, customer))

	rec := veidtypes.NewIdentityRecord(customer.String(), ctx.BlockTime())
	rec.CurrentScore = 90
	k.SetVEIDKeeper(stubVEIDKeeper{found: true, score: 90, record: *rec})

	// Requirements that the identity satisfies.
	lenient := keeper.VEIDGatingRequirements{
		MinCustomerScore:        10,
		RequireUnlockedIdentity: true,
	}

	result, err := k.CheckVEIDGating(ctx, customer, lenient)
	require.NoError(t, err)
	require.True(t, result.Passed)

	_, found := rolesKeeper.GetAccountState(ctx, customer)
	require.False(t, found, "a passing VEID check must not write global account state either")
}
