// Package keeper contains tests for the Fraud module keeper.
//
// t_9169400a: These tests cover the second-reviewer requirement for
// suspension/termination resolutions, which strip an account's access
// network-wide and must not be driven by a single moderator.
package keeper

import (
	"errors"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/fraud/types"
)

// fraudCoSignFixture wires a keeper with a submitted report and three distinct
// moderator identities so the two-person rule can actually be exercised.
type fraudCoSignFixture struct {
	k        Keeper
	ctx      sdk.Context
	roles    *MockRolesKeeper
	report   types.FraudReport
	reporter sdk.AccAddress
	modA     sdk.AccAddress
	modB     sdk.AccAddress
	modC     sdk.AccAddress
	target   sdk.AccAddress
}

func setupFraudCoSignFixture(t *testing.T) fraudCoSignFixture {
	t.Helper()

	k, ctx, mockRoles, mockProvider := setupKeeper(t)

	f := fraudCoSignFixture{
		k:        k,
		ctx:      ctx,
		roles:    mockRoles,
		reporter: sdk.AccAddress("cosmos1reporter______"),
		modA:     sdk.AccAddress("cosmos1moderator_a___"),
		modB:     sdk.AccAddress("cosmos1moderator_b___"),
		modC:     sdk.AccAddress("cosmos1moderator_c___"),
		target:   sdk.AccAddress("cosmos1reported______"),
	}

	mockProvider.SetProvider(f.reporter.String())
	for _, mod := range []sdk.AccAddress{f.modA, f.modB, f.modC} {
		mockRoles.SetModerator(mod.String())
		if !k.IsModerator(ctx, mod) {
			t.Fatalf("moderator %s is not recognised", mod)
		}
	}

	report := types.NewFraudReport(
		"",
		f.reporter.String(),
		f.target.String(),
		types.FraudCategoryPaymentFraud,
		testFraudDescription,
		createValidEvidence(),
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)
	if err := k.SubmitFraudReport(ctx, report); err != nil {
		t.Fatalf("SubmitFraudReport() error = %v", err)
	}

	stored, found := k.GetFraudReport(ctx, report.ID)
	if !found {
		t.Fatal("submitted report not found")
	}
	f.report = stored

	return f
}

// at advances the fixture's block time to the given Unix second.
func (f fraudCoSignFixture) at(unix int64) fraudCoSignFixture {
	f.ctx = f.ctx.WithBlockTime(time.Unix(unix, 0))
	return f
}

// ---------------------------------------------------------------------------
// DONE WHEN 2: a single moderator cannot suspend or terminate.
// ---------------------------------------------------------------------------

func TestFraud_SingleModeratorCannotResolveSuspension(t *testing.T) {
	f := setupFraudCoSignFixture(t)

	for _, resolution := range []types.ResolutionType{
		types.ResolutionTypeSuspension,
		types.ResolutionTypeTermination,
	} {
		t.Run(resolution.String(), func(t *testing.T) {
			err := f.k.ResolveFraudReport(f.ctx, f.report.ID, resolution, "notes", f.modA.String())
			if err == nil {
				t.Fatalf("ResolveFraudReport(%s) succeeded for a single moderator; want %v",
					resolution, types.ErrSecondReviewerRequired)
			}
			if !errIs(err, types.ErrSecondReviewerRequired) {
				t.Fatalf("ResolveFraudReport(%s) error = %v; want %v",
					resolution, err, types.ErrSecondReviewerRequired)
			}

			// Nothing happened to the report.
			stored, found := f.k.GetFraudReport(f.ctx, f.report.ID)
			if !found {
				t.Fatal("report disappeared")
			}
			if stored.Status == types.FraudReportStatusResolved {
				t.Fatalf("report was resolved by a single moderator via %s", resolution)
			}
			if stored.Resolution == resolution {
				t.Fatalf("resolution %s took effect for a single moderator", resolution)
			}
		})
	}
}

func TestFraud_ProposeThenConfirmRequiresDistinctReviewer(t *testing.T) {
	f := setupFraudCoSignFixture(t)

	// One moderator proposes a suspension; the report is untouched.
	pending, err := f.k.ProposeResolution(
		f.ctx, f.report.ID, types.ResolutionTypeSuspension, "suspension proposed", f.modA.String())
	if err != nil {
		t.Fatalf("ProposeResolution() error = %v", err)
	}
	if pending.ProposedBy != f.modA.String() {
		t.Fatalf("ProposedBy = %s; want %s", pending.ProposedBy, f.modA.String())
	}

	stored, _ := f.k.GetFraudReport(f.ctx, f.report.ID)
	if stored.Status == types.FraudReportStatusResolved {
		t.Fatal("a proposal must not resolve the report")
	}

	// The proposing moderator cannot confirm their own proposal.
	if _, err := f.k.ConfirmResolution(f.ctx, f.report.ID, f.modA.String()); !errIs(err, types.ErrSecondReviewerMustDiffer) {
		t.Fatalf("self-confirmation error = %v; want %v", err, types.ErrSecondReviewerMustDiffer)
	}

	// A non-moderator cannot confirm it either.
	if _, err := f.k.ConfirmResolution(f.ctx, f.report.ID, f.target.String()); !errIs(err, types.ErrUnauthorizedModerator) {
		t.Fatalf("non-moderator confirmation error = %v; want %v", err, types.ErrUnauthorizedModerator)
	}

	// A distinct moderator can, and only then does it take effect.
	resolution, err := f.k.ConfirmResolution(f.ctx, f.report.ID, f.modB.String())
	if err != nil {
		t.Fatalf("ConfirmResolution() error = %v", err)
	}
	if resolution != types.ResolutionTypeSuspension {
		t.Fatalf("confirmed resolution = %s; want %s", resolution, types.ResolutionTypeSuspension)
	}

	resolved, _ := f.k.GetFraudReport(f.ctx, f.report.ID)
	if resolved.Status != types.FraudReportStatusResolved {
		t.Fatalf("status = %s; want resolved", resolved.Status)
	}
	if resolved.Resolution != types.ResolutionTypeSuspension {
		t.Fatalf("resolution = %s; want suspension", resolved.Resolution)
	}

	// The pending record is consumed.
	if _, found := f.k.GetPendingResolution(f.ctx, f.report.ID); found {
		t.Fatal("pending resolution should be deleted after confirmation")
	}
}

func TestFraud_ConfirmWithoutProposalRejected(t *testing.T) {
	f := setupFraudCoSignFixture(t)

	if _, err := f.k.ConfirmResolution(f.ctx, f.report.ID, f.modB.String()); !errIs(err, types.ErrResolutionNotPending) {
		t.Fatalf("error = %v; want %v", err, types.ErrResolutionNotPending)
	}
}

func TestFraud_ProposeRejectsNonSecondReviewerResolution(t *testing.T) {
	f := setupFraudCoSignFixture(t)

	// A warning needs no co-signature and must not be routed through the
	// proposal path.
	if _, err := f.k.ProposeResolution(
		f.ctx, f.report.ID, types.ResolutionTypeWarning, "notes", f.modA.String()); !errIs(err, types.ErrInvalidResolution) {
		t.Fatalf("error = %v; want %v", err, types.ErrInvalidResolution)
	}
}

func TestFraud_ProposeRequiresModerator(t *testing.T) {
	f := setupFraudCoSignFixture(t)

	if _, err := f.k.ProposeResolution(
		f.ctx, f.report.ID, types.ResolutionTypeTermination, "notes", f.target.String()); !errIs(err, types.ErrUnauthorizedModerator) {
		t.Fatalf("error = %v; want %v", err, types.ErrUnauthorizedModerator)
	}
}

// ---------------------------------------------------------------------------
// The emergency-hold analogue: an unreviewed proposal expires on its own.
// ---------------------------------------------------------------------------

func TestFraud_UnreviewedProposalExpires(t *testing.T) {
	f := setupFraudCoSignFixture(t)
	now := f.ctx.BlockTime().Unix()

	pending, err := f.k.ProposeResolution(
		f.ctx, f.report.ID, types.ResolutionTypeTermination, "termination proposed", f.modA.String())
	if err != nil {
		t.Fatalf("ProposeResolution() error = %v", err)
	}
	if pending.ExpiresAt <= now {
		t.Fatalf("ExpiresAt = %d; want a window after %d", pending.ExpiresAt, now)
	}

	// Before the window closes it is still pending and reviewable.
	before := f.at(pending.ExpiresAt - 1)
	expired, err := before.k.ExpirePendingResolutions(before.ctx)
	if err != nil {
		t.Fatalf("ExpirePendingResolutions() error = %v", err)
	}
	if len(expired) != 0 {
		t.Fatalf("expired %d proposals early; want 0", len(expired))
	}
	if _, found := before.k.GetPendingResolution(before.ctx, f.report.ID); !found {
		t.Fatal("proposal vanished before its window closed")
	}

	// At the window it lapses, via the EndBlocker path...
	after := f.at(pending.ExpiresAt)
	if err := after.k.ProcessPendingResolutionExpiry(after.ctx); err != nil {
		t.Fatalf("ProcessPendingResolutionExpiry() error = %v", err)
	}
	if _, found := after.k.GetPendingResolution(after.ctx, f.report.ID); found {
		t.Fatal("lapsed proposal was not cleared")
	}

	// ...and the report is left as it was: a termination never happened.
	stored, _ := after.k.GetFraudReport(after.ctx, f.report.ID)
	if stored.Status == types.FraudReportStatusResolved {
		t.Fatal("report was resolved by an unreviewed proposal")
	}
	if stored.Resolution == types.ResolutionTypeTermination {
		t.Fatal("termination took effect without review")
	}

	// Confirming after the window is refused, and clears the record.
	late := f.at(pending.ExpiresAt + 1)
	if _, err := late.k.ConfirmResolution(late.ctx, f.report.ID, late.modB.String()); !errIs(err, types.ErrResolutionNotPending) {
		t.Fatalf("late confirmation error = %v; want %v", err, types.ErrResolutionNotPending)
	}
}

func TestFraud_ProposalReviewedInWindowTakesEffect(t *testing.T) {
	f := setupFraudCoSignFixture(t)

	pending, err := f.k.ProposeResolution(
		f.ctx, f.report.ID, types.ResolutionTypeTermination, "termination proposed", f.modA.String())
	if err != nil {
		t.Fatalf("ProposeResolution() error = %v", err)
	}

	// Reviewed one second inside the window by a distinct moderator.
	inWindow := f.at(pending.ExpiresAt - 1)
	if _, err := inWindow.k.ConfirmResolution(inWindow.ctx, f.report.ID, inWindow.modC.String()); err != nil {
		t.Fatalf("ConfirmResolution() inside window error = %v", err)
	}

	stored, _ := inWindow.k.GetFraudReport(inWindow.ctx, f.report.ID)
	if stored.Status != types.FraudReportStatusResolved {
		t.Fatalf("status = %s; want resolved", stored.Status)
	}
	if stored.Resolution != types.ResolutionTypeTermination {
		t.Fatalf("resolution = %s; want termination", stored.Resolution)
	}
}

// ---------------------------------------------------------------------------
// No regression: resolutions that never needed a co-signature still apply.
// ---------------------------------------------------------------------------

func TestFraud_ImmediateResolutionStillWorks(t *testing.T) {
	f := setupFraudCoSignFixture(t)

	for _, resolution := range []types.ResolutionType{
		types.ResolutionTypeWarning,
		types.ResolutionTypeRefund,
	} {
		t.Run(resolution.String(), func(t *testing.T) {
			// Each subtest submits a report that is distinct in content: develop's
			// dedup index rejects a byte-identical resubmission from the same
			// reporter, so reusing one description would fail the second subtest on
			// ErrDuplicateReport rather than on anything to do with the co-signature
			// rule under test.
			report := types.NewFraudReport(
				"",
				f.reporter.String(),
				f.target.String(),
				types.FraudCategoryResourceAbuse,
				testFraudDescription+" ["+resolution.String()+"]",
				createValidEvidence(),
				f.ctx.BlockHeight(),
				f.ctx.BlockTime(),
			)
			if err := f.k.SubmitFraudReport(f.ctx, report); err != nil {
				t.Fatalf("SubmitFraudReport() error = %v", err)
			}

			if err := f.k.ResolveFraudReport(f.ctx, report.ID, resolution, "notes", f.modA.String()); err != nil {
				t.Fatalf("ResolveFraudReport(%s) error = %v", resolution, err)
			}

			stored, _ := f.k.GetFraudReport(f.ctx, report.ID)
			if stored.Status != types.FraudReportStatusResolved {
				t.Fatalf("status = %s; want resolved", stored.Status)
			}
			if stored.Resolution != resolution {
				t.Fatalf("resolution = %s; want %s", stored.Resolution, resolution)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// The co-signature rule must be exhaustive over the resolution enum, so a new
// resolution cannot silently inherit single-moderator application.
// ---------------------------------------------------------------------------

func TestFraud_SecondReviewerClassificationIsExhaustive(t *testing.T) {
	expected := map[types.ResolutionType]bool{
		types.ResolutionTypeUnspecified: false,
		types.ResolutionTypeWarning:     false,
		types.ResolutionTypeSuspension:  true,
		types.ResolutionTypeTermination: true,
		types.ResolutionTypeRefund:      false,
		types.ResolutionTypeNoAction:    false,
	}

	for resolution, want := range expected {
		if got := resolution.RequiresSecondReviewer(); got != want {
			t.Errorf("%s.RequiresSecondReviewer() = %v; want %v", resolution, got, want)
		}
		if got := resolution.IsImmediate(); got != (resolution.IsValid() && !want) {
			t.Errorf("%s.IsImmediate() = %v; want %v", resolution, got, resolution.IsValid() && !want)
		}
	}

	if len(expected) != int(types.ResolutionTypeNoAction)+1 {
		t.Errorf("test table covers %d resolution values but the enum has %d; "+
			"update this test when the enum grows",
			len(expected), int(types.ResolutionTypeNoAction)+1)
	}
}

// errIs reports whether err is (or wraps) target. cosmos-sdk errors implement
// Is, so a plain equality check would miss every wrapped sentinel here.
func errIs(err, target error) bool {
	return errors.Is(err, target)
}
