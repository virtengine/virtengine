package keeper

import (
	"testing"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// TestIsZeroRequirements_IgnoresUnlockedIdentity documents a real gap found while
// auditing the order-creation path.
//
// DefaultVEIDGatingRequirements() sets RequireUnlockedIdentity: true, but
// isZeroRequirements() does not consult that field. CreateOrder only runs the VEID
// gating check when the requirements are non-zero, so with the defaults the guard
// short-circuits and NO identity check runs at all — not even the unlocked-identity
// requirement the defaults ask for.
//
// This is under-enforcement (too little gating, never too much), and tightening the
// guard would change which orders are accepted, i.e. block validity. That is
// consensus-breaking and needs explicit sign-off, so this test pins the current
// behaviour and will fail loudly if and when the guard is corrected — at which point
// the assertion should be inverted deliberately rather than discovered by accident.
//
// See the eligibility map in the PR body for the surrounding analysis.
func TestIsZeroRequirements_IgnoresUnlockedIdentity(t *testing.T) {
	defaults := DefaultVEIDGatingRequirements()

	if !defaults.RequireUnlockedIdentity {
		t.Fatal("expected the defaults to request an unlocked identity")
	}

	if !isZeroRequirements(defaults) {
		t.Fatal("documented gap closed: isZeroRequirements now honours RequireUnlockedIdentity — " +
			"the order-creation guard has been tightened. Update this test and the consensus-review note.")
	}
}

// TestIsZeroRequirements_DetectsRealGating confirms the guard still recognises the
// requirement fields it does inspect, so the gap above is specifically the missing
// unlocked-identity clause and not a generally broken predicate.
func TestIsZeroRequirements_DetectsRealGating(t *testing.T) {
	base := DefaultVEIDGatingRequirements()

	cases := []struct {
		name     string
		mutate   func(r *VEIDGatingRequirements)
		wantZero bool
	}{
		{
			name:     "defaults are treated as no gating",
			mutate:   func(*VEIDGatingRequirements) {},
			wantZero: true,
		},
		{
			name:     "a score floor is gating",
			mutate:   func(r *VEIDGatingRequirements) { r.MinCustomerScore = 1 },
			wantZero: false,
		},
		{
			name:     "a tier floor above unverified is gating",
			mutate:   func(r *VEIDGatingRequirements) { r.MinCustomerTier = veidtypes.TierBasic },
			wantZero: false,
		},
		{
			name: "required scopes are gating",
			mutate: func(r *VEIDGatingRequirements) {
				r.RequiredScopes = []veidtypes.ScopeType{veidtypes.ScopeTypeIDDocument}
			},
			wantZero: false,
		},
		{
			name:     "a verified-status requirement is gating",
			mutate:   func(r *VEIDGatingRequirements) { r.RequireVerifiedStatus = true },
			wantZero: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := base
			tc.mutate(&r)
			if got := isZeroRequirements(r); got != tc.wantZero {
				t.Errorf("isZeroRequirements() = %v, want %v", got, tc.wantZero)
			}
		})
	}
}
