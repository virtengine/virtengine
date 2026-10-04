package keeper

import (
	"testing"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// TestIsZeroRequirements_HonoursUnlockedIdentity records the corrected guard.
//
// History: DefaultVEIDGatingRequirements() used to set RequireUnlockedIdentity: true
// while isZeroRequirements() did not consult that field, so the defaults were
// reported as "zero requirements" and the order-creation guard short-circuited:
// NO identity check ran, not even the unlocked-identity requirement the defaults
// asked for. The constraint-resolution work in this change corrects both halves
// together so the predicate and the default agree with each other.
//
// Block validity is preserved for every existing path, and this is the reason the
// correction does not need a consensus-breaking sign-off:
//
//   - Before: defaults carried RequireUnlockedIdentity=true but the field was
//     ignored, so isZeroRequirements(defaults) == true and the check was skipped.
//   - After: the default is a TRUE zero (RequireUnlockedIdentity=false), so
//     isZeroRequirements(defaults) == true and the check is still skipped.
//
// The live order path (getVEIDGatingRequirementsForOrder, which returns the
// defaults) therefore accepts exactly the same orders as before. What changes is
// that a requirement which explicitly asks for an unlocked identity is now
// honoured instead of silently discarded.
//
// The previous revision of this test pinned the buggy behaviour on purpose and
// asked for the assertion to be inverted deliberately if the guard was ever
// corrected. That is what this revision does.
func TestIsZeroRequirements_HonoursUnlockedIdentity(t *testing.T) {
	defaults := DefaultVEIDGatingRequirements()

	// The default must be a true zero: no field may obligate a buyer, otherwise
	// the default would gate every order on the chain.
	if defaults.RequireUnlockedIdentity {
		t.Fatal("defaults must not request an unlocked identity (no blanket gating)")
	}
	if !isZeroRequirements(defaults) {
		t.Fatal("defaults must remain 'no gating' so existing order acceptance is unchanged")
	}

	// An unlocked-identity-only requirement must NOT be mistaken for "no
	// requirements": it constrains the buyer and has to reach the checker.
	unlockedOnly := VEIDGatingRequirements{RequireUnlockedIdentity: true}
	if isZeroRequirements(unlockedOnly) {
		t.Fatal("an unlocked-only requirement must not be treated as zero requirements")
	}
}

// TestIsZeroRequirements_DetectsRealGating confirms the guard recognises each
// requirement field, including the unlocked-identity clause that used to be
// missing from the predicate.
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
		{
			name:     "an unlocked-identity requirement is gating",
			mutate:   func(r *VEIDGatingRequirements) { r.RequireUnlockedIdentity = true },
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
