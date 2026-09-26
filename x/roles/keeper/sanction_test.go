package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/roles/keeper"
	"github.com/virtengine/virtengine/x/roles/types"
)

// ---------------------------------------------------------------------------
// Sanction model tests.
//
// These cover the card's DONE-WHEN clauses:
//   1. scope/expiry/notice/appeal transitions
//   2. a single moderator cannot suspend/terminate without a distinct second
//      reviewer, and an emergency hold expires unless reviewed
//   3. an expired suspension restores the preceding state with no manual action
// ---------------------------------------------------------------------------

const testJustification = "Confirmed fraudulent activity with supporting evidence."

// sanctionFixture wires a keeper with three distinct moderator identities.
type sanctionFixture struct {
	ctx    sdk.Context
	k      keeper.Keeper
	modA   sdk.AccAddress
	modB   sdk.AccAddress
	modC   sdk.AccAddress
	target sdk.AccAddress
}

func setupSanctionFixture(t *testing.T) sanctionFixture {
	t.Helper()

	ctx, k := setupKeeper(t)

	f := sanctionFixture{
		ctx:    ctx,
		k:      k,
		modA:   sdk.AccAddress([]byte("moderator_a_address1")),
		modB:   sdk.AccAddress([]byte("moderator_b_address2")),
		modC:   sdk.AccAddress([]byte("moderator_c_address3")),
		target: sdk.AccAddress([]byte("sanction_target_acct1")),
	}

	for _, mod := range []sdk.AccAddress{f.modA, f.modB, f.modC} {
		require.NoError(t, k.AssignRole(ctx, mod, types.RoleModerator, mod))
		require.True(t, k.IsModerator(ctx, mod), "moderator role must be effective")
	}
	require.False(t, k.IsModerator(ctx, f.target), "target must not be a moderator")

	return f
}

// at advances the fixture's block time to the given Unix second.
func (f sanctionFixture) at(unix int64) sanctionFixture {
	f.ctx = f.ctx.WithBlockTime(time.Unix(unix, 0))
	return f
}

func suspensionProposal(subject string, expiresAt int64) types.Sanction {
	return types.Sanction{
		Subject:       subject,
		Scope:         types.SanctionScopeAccount,
		Kind:          types.SanctionKindSuspension,
		ReasonCode:    types.SanctionReasonFraudConfirmed,
		Justification: testJustification,
		Notice:        "You are suspended. You may appeal within 30 days.",
		ExpiresAt:     expiresAt,
	}
}

func terminationProposal(subject string) types.Sanction {
	return types.Sanction{
		Subject:       subject,
		Scope:         types.SanctionScopeAccount,
		Kind:          types.SanctionKindTermination,
		ReasonCode:    types.SanctionReasonFraudConfirmed,
		Justification: testJustification,
		Notice:        "Your account is terminated. You may appeal.",
	}
}

func emergencyHoldProposal(subject string, expiresAt int64) types.Sanction {
	return types.Sanction{
		Subject:       subject,
		Scope:         types.SanctionScopeAccount,
		Kind:          types.SanctionKindEmergencyHold,
		ReasonCode:    types.SanctionReasonFraudSuspected,
		Justification: testJustification,
		Notice:        "Emergency hold placed pending review.",
		ExpiresAt:     expiresAt,
	}
}

// ---------------------------------------------------------------------------
// DONE WHEN 2 (first half): a single moderator cannot suspend or terminate.
// ---------------------------------------------------------------------------

func TestSanction_SingleModeratorCannotSuspend(t *testing.T) {
	f := setupSanctionFixture(t)
	now := f.ctx.BlockTime().Unix()

	// One moderator proposes a suspension, alone.
	proposal := suspensionProposal(f.target.String(), now+3600)
	sanction, err := f.k.ImposeSanction(f.ctx, proposal, f.modA)
	require.NoError(t, err)

	// It is recorded, but not in force: no account-state effect yet.
	require.Equal(t, types.SanctionStatusPendingReview, sanction.Status)
	require.False(t, sanction.InForce(), "a pending suspension must not be in force")

	state, found := f.k.GetAccountState(f.ctx, f.target)
	if found {
		require.NotEqual(t, types.AccountStateSuspended, state.State,
			"a single moderator must not be able to suspend an account")
	}
	require.Equal(t, types.AccountStateActive, f.k.EffectiveAccountState(f.ctx, f.target))
	require.True(t, f.k.IsAccountOperational(f.ctx, f.target),
		"the account must remain operational after a single-handed suspension")

	// The imposing moderator cannot confirm their own action.
	_, err = f.k.ConfirmSanction(f.ctx, sanction.ID, f.modA, 0)
	require.ErrorIs(t, err, types.ErrSecondReviewerMustDiffer)

	// A distinct moderator can.
	confirmed, err := f.k.ConfirmSanction(f.ctx, sanction.ID, f.modB, 0)
	require.NoError(t, err)
	require.Equal(t, types.SanctionStatusActive, confirmed.Status)
	require.Equal(t, f.modB.String(), confirmed.SecondReviewer)

	require.Equal(t, types.AccountStateSuspended, f.k.EffectiveAccountState(f.ctx, f.target))
	require.False(t, f.k.IsAccountOperational(f.ctx, f.target))
}

func TestSanction_SingleModeratorCannotTerminate(t *testing.T) {
	f := setupSanctionFixture(t)

	proposal := terminationProposal(f.target.String())
	sanction, err := f.k.ImposeSanction(f.ctx, proposal, f.modA)
	require.NoError(t, err)
	require.Equal(t, types.SanctionStatusPendingReview, sanction.Status)
	require.False(t, sanction.InForce())

	require.Equal(t, types.AccountStateActive, f.k.EffectiveAccountState(f.ctx, f.target))

	// Self-confirmation is refused.
	_, err = f.k.ConfirmSanction(f.ctx, sanction.ID, f.modA, 0)
	require.ErrorIs(t, err, types.ErrSecondReviewerMustDiffer)

	// A non-moderator cannot confirm either.
	_, err = f.k.ConfirmSanction(f.ctx, sanction.ID, f.target, 0)
	require.ErrorIs(t, err, types.ErrSecondReviewerUnauthorized)

	// A distinct moderator puts it into force.
	confirmed, err := f.k.ConfirmSanction(f.ctx, sanction.ID, f.modC, 0)
	require.NoError(t, err)
	require.Equal(t, types.SanctionStatusActive, confirmed.Status)
	require.Equal(t, types.AccountStateTerminated, f.k.EffectiveAccountState(f.ctx, f.target))
}

func TestSanction_ConfirmRequiresPendingRecord(t *testing.T) {
	f := setupSanctionFixture(t)
	now := f.ctx.BlockTime().Unix()

	sanction, err := f.k.ImposeSanction(f.ctx, suspensionProposal(f.target.String(), now+3600), f.modA)
	require.NoError(t, err)

	_, err = f.k.ConfirmSanction(f.ctx, sanction.ID, f.modB, 0)
	require.NoError(t, err)

	// Confirming twice is not a transition.
	_, err = f.k.ConfirmSanction(f.ctx, sanction.ID, f.modC, 0)
	require.ErrorIs(t, err, types.ErrInvalidSanctionTransition)

	_, err = f.k.ConfirmSanction(f.ctx, "sanction-does-not-exist", f.modB, 0)
	require.ErrorIs(t, err, types.ErrSanctionNotFound)
}

// ---------------------------------------------------------------------------
// DONE WHEN 2 (second half): an emergency hold expires unless reviewed.
// ---------------------------------------------------------------------------

func TestSanction_EmergencyHoldBindsButExpiresUnreviewed(t *testing.T) {
	f := setupSanctionFixture(t)
	now := f.ctx.BlockTime().Unix()

	// A single moderator may place an emergency hold alone...
	hold, err := f.k.ImposeSanction(f.ctx, emergencyHoldProposal(f.target.String(), now+3600), f.modA)
	require.NoError(t, err)
	require.Equal(t, types.SanctionStatusPendingReview, hold.Status)

	// ...and it binds immediately, which is its purpose.
	require.True(t, hold.InForce())
	require.Equal(t, types.AccountStateSuspended, f.k.EffectiveAccountState(f.ctx, f.target))

	// Unreviewed, it lapses automatically at its window.
	lapsed := f.at(now + 3601)
	require.NoError(t, lapsed.k.ProcessSanctionExpiry(lapsed.ctx))

	stored, found := lapsed.k.GetSanction(lapsed.ctx, hold.ID)
	require.True(t, found)
	require.Equal(t, types.SanctionStatusExpired, stored.Status)
	require.False(t, stored.InForce())

	// The account is restored with no manual intervention.
	require.Equal(t, types.AccountStateActive, lapsed.k.EffectiveAccountState(lapsed.ctx, f.target))
	require.True(t, lapsed.k.IsAccountOperational(lapsed.ctx, f.target))
}

func TestSanction_EmergencyHoldReviewedBeforeWindowPersists(t *testing.T) {
	f := setupSanctionFixture(t)
	now := f.ctx.BlockTime().Unix()

	hold, err := f.k.ImposeSanction(f.ctx, emergencyHoldProposal(f.target.String(), now+3600), f.modA)
	require.NoError(t, err)

	// Reviewed within the window by a distinct moderator, who chooses the
	// duration the interim hold is converted into: it survives that window.
	confirmed, err := f.k.ConfirmSanction(f.ctx, hold.ID, f.modB, now+86400)
	require.NoError(t, err)
	require.Equal(t, types.SanctionStatusActive, confirmed.Status)
	require.Equal(t, now+86400, confirmed.ExpiresAt)

	after := f.at(now + 3601)
	require.NoError(t, after.k.ProcessSanctionExpiry(after.ctx))

	stored, found := after.k.GetSanction(after.ctx, hold.ID)
	require.True(t, found)
	require.Equal(t, types.SanctionStatusActive, stored.Status,
		"a reviewed hold must not be reaped by its own review window")
	require.Equal(t, types.AccountStateSuspended, after.k.EffectiveAccountState(after.ctx, f.target))
}

func TestSanction_EmergencyHoldRejectedAfterWindow(t *testing.T) {
	f := setupSanctionFixture(t)
	now := f.ctx.BlockTime().Unix()

	hold, err := f.k.ImposeSanction(f.ctx, emergencyHoldProposal(f.target.String(), now+3600), f.modA)
	require.NoError(t, err)

	// Reviewing after the window closes the hold instead of confirming it.
	late := f.at(now + 3601)
	_, err = late.k.ConfirmSanction(late.ctx, hold.ID, late.modB, 0)
	require.ErrorIs(t, err, types.ErrEmergencyHoldExpired)

	stored, found := late.k.GetSanction(late.ctx, hold.ID)
	require.True(t, found)
	require.Equal(t, types.SanctionStatusExpired, stored.Status)
	require.Equal(t, types.AccountStateActive, late.k.EffectiveAccountState(late.ctx, f.target))
}

func TestSanction_EmergencyHoldWindowIsBounded(t *testing.T) {
	f := setupSanctionFixture(t)
	now := f.ctx.BlockTime().Unix()

	// A single-handed hold may not exceed the maximum unreviewed window.
	_, err := f.k.ImposeSanction(f.ctx,
		emergencyHoldProposal(f.target.String(), now+types.MaxEmergencyHoldSeconds+1), f.modA)
	require.ErrorIs(t, err, types.ErrEmergencyHoldTooLong)

	// It must expire in the future.
	_, err = f.k.ImposeSanction(f.ctx,
		emergencyHoldProposal(f.target.String(), now-1), f.modA)
	require.ErrorIs(t, err, types.ErrInvalidSanction)

	// Omitted expiry defaults to the default hold window.
	hold, err := f.k.ImposeSanction(f.ctx, emergencyHoldProposal(f.target.String(), 0), f.modA)
	require.NoError(t, err)
	require.Equal(t, now+types.DefaultEmergencyHoldSeconds, hold.ExpiresAt)
}

// ---------------------------------------------------------------------------
// DONE WHEN 3: an expired suspension restores the preceding state.
// ---------------------------------------------------------------------------

func TestSanction_ExpiredSuspensionRestoresActive(t *testing.T) {
	f := setupSanctionFixture(t)
	now := f.ctx.BlockTime().Unix()

	sanction, err := f.k.ImposeSanction(f.ctx, suspensionProposal(f.target.String(), now+3600), f.modA)
	require.NoError(t, err)

	_, err = f.k.ConfirmSanction(f.ctx, sanction.ID, f.modB, 0)
	require.NoError(t, err)
	require.Equal(t, types.AccountStateSuspended, f.k.EffectiveAccountState(f.ctx, f.target))

	// Nothing happens before the window closes.
	before := f.at(now + 3599)
	require.NoError(t, before.k.ProcessSanctionExpiry(before.ctx))
	require.Equal(t, types.AccountStateSuspended, before.k.EffectiveAccountState(before.ctx, f.target))

	// At expiry, the EndBlocker restores the account with no transaction.
	after := f.at(now + 3600)
	require.NoError(t, after.k.ProcessSanctionExpiry(after.ctx))

	require.Equal(t, types.AccountStateActive, after.k.EffectiveAccountState(after.ctx, f.target))
	require.True(t, after.k.IsAccountOperational(after.ctx, f.target))

	state, found := after.k.GetAccountState(after.ctx, f.target)
	require.True(t, found)
	require.Equal(t, types.AccountStateActive, state.State)
	require.Contains(t, state.Reason, "sanction")
}

func TestSanction_TerminationReactivationRequiresReview(t *testing.T) {
	f := setupSanctionFixture(t)
	now := f.ctx.BlockTime().Unix()

	termination, err := f.k.ImposeSanction(f.ctx, terminationProposal(f.target.String()), f.modA)
	require.NoError(t, err)
	_, err = f.k.ConfirmSanction(f.ctx, termination.ID, f.modB, 0)
	require.NoError(t, err)
	require.Equal(t, types.AccountStateTerminated, f.k.EffectiveAccountState(f.ctx, f.target))

	// The bare administrative path still cannot reactivate a terminated account.
	err = f.k.SetAccountState(f.ctx, f.target, types.AccountStateActive, "manual reactivation", f.modA)
	require.ErrorIs(t, err, types.ErrInvalidStateTransition)

	// The reviewed path can: revoking the sanction restores the account.
	_, err = f.k.RevokeSanction(f.ctx, termination.ID, f.modC, "reinstated on review")
	require.NoError(t, err)
	require.Equal(t, types.AccountStateActive, f.k.EffectiveAccountState(f.ctx, f.target))
	require.True(t, f.k.IsAccountOperational(f.ctx, f.target))

	_ = now
}

// ---------------------------------------------------------------------------
// DONE WHEN 1: scope, notice and validation of each field.
// ---------------------------------------------------------------------------

func TestSanction_ProposalValidation(t *testing.T) {
	f := setupSanctionFixture(t)
	now := f.ctx.BlockTime().Unix()

	base := suspensionProposal(f.target.String(), now+3600)

	t.Run("JustificationTooShort", func(t *testing.T) {
		p := base
		p.Justification = "too short"
		_, err := f.k.ImposeSanction(f.ctx, p, f.modA)
		require.ErrorIs(t, err, types.ErrInvalidSanctionJustification)
	})

	t.Run("NoticeRequired", func(t *testing.T) {
		p := base
		p.Notice = ""
		_, err := f.k.ImposeSanction(f.ctx, p, f.modA)
		require.ErrorIs(t, err, types.ErrSanctionNoticeRequired)
	})

	t.Run("UnknownScope", func(t *testing.T) {
		p := base
		p.Scope = types.SanctionScope(99)
		_, err := f.k.ImposeSanction(f.ctx, p, f.modA)
		require.ErrorIs(t, err, types.ErrInvalidSanctionScope)
	})

	t.Run("UnknownReasonCode", func(t *testing.T) {
		p := base
		p.ReasonCode = types.SanctionReasonCode(99)
		_, err := f.k.ImposeSanction(f.ctx, p, f.modA)
		require.ErrorIs(t, err, types.ErrInvalidSanctionReason)
	})

	t.Run("OrderScopeRequiresRef", func(t *testing.T) {
		p := base
		p.Scope = types.SanctionScopeOrder
		p.ScopeRef = ""
		_, err := f.k.ImposeSanction(f.ctx, p, f.modA)
		require.ErrorIs(t, err, types.ErrInvalidSanction)
	})

	t.Run("AccountScopeRejectsRef", func(t *testing.T) {
		p := base
		p.Scope = types.SanctionScopeAccount
		p.ScopeRef = "order-1"
		_, err := f.k.ImposeSanction(f.ctx, p, f.modA)
		require.ErrorIs(t, err, types.ErrInvalidSanction)
	})

	t.Run("SuspensionRequiresExpiry", func(t *testing.T) {
		p := base
		p.ExpiresAt = 0
		_, err := f.k.ImposeSanction(f.ctx, p, f.modA)
		require.ErrorIs(t, err, types.ErrInvalidSanction)
	})

	t.Run("InvalidSubjectAddress", func(t *testing.T) {
		p := base
		p.Subject = "not-a-bech32-address"
		_, err := f.k.ImposeSanction(f.ctx, p, f.modA)
		require.Error(t, err)
	})

	t.Run("NonModeratorCannotImpose", func(t *testing.T) {
		_, err := f.k.ImposeSanction(f.ctx, base, f.target)
		require.ErrorIs(t, err, types.ErrSecondReviewerUnauthorized)
	})

	t.Run("ScopedSanctionIsRecorded", func(t *testing.T) {
		p := base
		p.Scope = types.SanctionScopeOrder
		p.ScopeRef = "order-42"
		s, err := f.k.ImposeSanction(f.ctx, p, f.modA)
		require.NoError(t, err)
		require.Equal(t, types.SanctionScopeOrder, s.Scope)
		require.Equal(t, "order-42", s.ScopeRef)
	})
}

func TestSanction_WarningHasNoStateEffect(t *testing.T) {
	f := setupSanctionFixture(t)
	now := f.ctx.BlockTime().Unix()

	warning, err := f.k.ImposeSanction(f.ctx, types.Sanction{
		Subject:       f.target.String(),
		Scope:         types.SanctionScopeAccount,
		Kind:          types.SanctionKindWarning,
		ReasonCode:    types.SanctionReasonTermsViolation,
		Justification: testJustification,
		ExpiresAt:     now + 86400,
	}, f.modA)
	require.NoError(t, err)

	// A warning needs no second reviewer and no notice...
	require.Equal(t, types.SanctionStatusActive, warning.Status)
	require.False(t, warning.Kind.RequiresNotice())
	require.Empty(t, warning.Notice)

	// ...and has no account-state effect.
	require.True(t, warning.InForce())
	_, projects := warning.EffectedAccountState()
	require.False(t, projects)
	require.Equal(t, types.AccountStateActive, f.k.EffectiveAccountState(f.ctx, f.target))
}

// ---------------------------------------------------------------------------
// DONE WHEN 1: appeal semantics.
// ---------------------------------------------------------------------------

func TestSanction_AppealBlocksEscalationUntilReviewed(t *testing.T) {
	f := setupSanctionFixture(t)
	now := f.ctx.BlockTime().Unix()

	suspension, err := f.k.ImposeSanction(f.ctx, suspensionProposal(f.target.String(), now+7200), f.modA)
	require.NoError(t, err)
	_, err = f.k.ConfirmSanction(f.ctx, suspension.ID, f.modB, 0)
	require.NoError(t, err)

	// Only the sanctioned party may appeal.
	_, err = f.k.OpenAppeal(f.ctx, suspension.ID, f.modA, testJustification)
	require.ErrorIs(t, err, types.ErrSanctionSubjectMismatch)

	appeal, err := f.k.OpenAppeal(f.ctx, suspension.ID, f.target, testJustification)
	require.NoError(t, err)
	require.Equal(t, types.SanctionStatusAppealPending, appeal.Status)
	require.Equal(t, suspension.ID, appeal.AppealOf)

	// A second appeal is refused while one is open.
	_, err = f.k.OpenAppeal(f.ctx, suspension.ID, f.target, testJustification)
	require.ErrorIs(t, err, types.ErrAppealAlreadyOpen)

	// Escalation beyond the challenged severity is blocked until review.
	_, err = f.k.ImposeSanction(f.ctx, terminationProposal(f.target.String()), f.modC)
	require.ErrorIs(t, err, types.ErrAppealAlreadyOpen)

	// The imposing reviewer may not decide the appeal.
	_, err = f.k.ResolveAppeal(f.ctx, appeal.ID, f.modB, false, "denied")
	require.ErrorIs(t, err, types.ErrSecondReviewerMustDiffer)

	// The appellant may not decide their own appeal.
	_, err = f.k.ResolveAppeal(f.ctx, appeal.ID, f.target, true, "granted")
	require.ErrorIs(t, err, types.ErrSecondReviewerUnauthorized)
}

func TestSanction_AppealGrantedRevokesSanctionAndRestoresAccount(t *testing.T) {
	f := setupSanctionFixture(t)
	now := f.ctx.BlockTime().Unix()

	termination, err := f.k.ImposeSanction(f.ctx, terminationProposal(f.target.String()), f.modA)
	require.NoError(t, err)
	_, err = f.k.ConfirmSanction(f.ctx, termination.ID, f.modB, 0)
	require.NoError(t, err)
	require.Equal(t, types.AccountStateTerminated, f.k.EffectiveAccountState(f.ctx, f.target))

	appeal, err := f.k.OpenAppeal(f.ctx, termination.ID, f.target, testJustification)
	require.NoError(t, err)

	// A third, uninvolved moderator grants it.
	_, err = f.k.ResolveAppeal(f.ctx, appeal.ID, f.modC, true, "evidence did not support termination")
	require.NoError(t, err)

	stored, found := f.k.GetSanction(f.ctx, termination.ID)
	require.True(t, found)
	require.Equal(t, types.SanctionStatusRevoked, stored.Status)

	resolved, found := f.k.GetSanction(f.ctx, appeal.ID)
	require.True(t, found)
	require.Equal(t, types.SanctionStatusAppealGranted, resolved.Status)

	// Reactivation through a reviewed path: the account returns to active.
	require.Equal(t, types.AccountStateActive, f.k.EffectiveAccountState(f.ctx, f.target))
	require.True(t, f.k.IsAccountOperational(f.ctx, f.target))

	_ = now
}

func TestSanction_AppealDeniedLeavesSanctionInForce(t *testing.T) {
	f := setupSanctionFixture(t)
	now := f.ctx.BlockTime().Unix()

	suspension, err := f.k.ImposeSanction(f.ctx, suspensionProposal(f.target.String(), now+7200), f.modA)
	require.NoError(t, err)
	_, err = f.k.ConfirmSanction(f.ctx, suspension.ID, f.modB, 0)
	require.NoError(t, err)

	appeal, err := f.k.OpenAppeal(f.ctx, suspension.ID, f.target, testJustification)
	require.NoError(t, err)

	_, err = f.k.ResolveAppeal(f.ctx, appeal.ID, f.modC, false, "evidence supported the suspension")
	require.NoError(t, err)

	stored, found := f.k.GetSanction(f.ctx, suspension.ID)
	require.True(t, found)
	require.Equal(t, types.SanctionStatusActive, stored.Status)
	require.Equal(t, types.AccountStateSuspended, f.k.EffectiveAccountState(f.ctx, f.target))

	// The appeal is no longer open, so escalation is no longer blocked.
	_, err = f.k.ImposeSanction(f.ctx, terminationProposal(f.target.String()), f.modC)
	require.NoError(t, err)
}

func TestSanction_AppealOfAppealRefused(t *testing.T) {
	f := setupSanctionFixture(t)
	now := f.ctx.BlockTime().Unix()

	suspension, err := f.k.ImposeSanction(f.ctx, suspensionProposal(f.target.String(), now+7200), f.modA)
	require.NoError(t, err)
	_, err = f.k.ConfirmSanction(f.ctx, suspension.ID, f.modB, 0)
	require.NoError(t, err)

	appeal, err := f.k.OpenAppeal(f.ctx, suspension.ID, f.target, testJustification)
	require.NoError(t, err)

	_, err = f.k.OpenAppeal(f.ctx, appeal.ID, f.target, testJustification)
	require.ErrorIs(t, err, types.ErrSanctionNotAppealable)

	// A lapsed sanction cannot be appealed either.
	expired, err := f.k.ImposeSanction(f.ctx, suspensionProposal(f.target.String(), now+1), f.modA)
	require.NoError(t, err)
	later := f.at(now + 2)
	require.NoError(t, later.k.ProcessSanctionExpiry(later.ctx))
	_, err = later.k.OpenAppeal(later.ctx, expired.ID, f.target, testJustification)
	require.ErrorIs(t, err, types.ErrSanctionNotAppealable)
}

// ---------------------------------------------------------------------------
// Projection and ordering invariants.
// ---------------------------------------------------------------------------

func TestSanction_ProjectionPicksHarshestInForce(t *testing.T) {
	f := setupSanctionFixture(t)
	now := f.ctx.BlockTime().Unix()

	// Suspend, then terminate alongside it.
	suspension, err := f.k.ImposeSanction(f.ctx, suspensionProposal(f.target.String(), now+7200), f.modA)
	require.NoError(t, err)
	_, err = f.k.ConfirmSanction(f.ctx, suspension.ID, f.modB, 0)
	require.NoError(t, err)

	termination, err := f.k.ImposeSanction(f.ctx, terminationProposal(f.target.String()), f.modA)
	require.NoError(t, err)
	_, err = f.k.ConfirmSanction(f.ctx, termination.ID, f.modB, 0)
	require.NoError(t, err)

	require.Equal(t, types.AccountStateTerminated, f.k.EffectiveAccountState(f.ctx, f.target))

	// Clearing the harshest falls back to the suspension.
	_, err = f.k.RevokeSanction(f.ctx, termination.ID, f.modC, "cleared")
	require.NoError(t, err)
	require.Equal(t, types.AccountStateSuspended, f.k.EffectiveAccountState(f.ctx, f.target))

	// Clearing that too restores active.
	_, err = f.k.RevokeSanction(f.ctx, suspension.ID, f.modC, "cleared")
	require.NoError(t, err)
	require.Equal(t, types.AccountStateActive, f.k.EffectiveAccountState(f.ctx, f.target))
}

func TestSanction_ProjectionIsDeterministic(t *testing.T) {
	// The projection reduces over sanctions; running it repeatedly over the
	// same store must produce the same answer with no dependence on iteration
	// order, and must be idempotent.
	f := setupSanctionFixture(t)
	now := f.ctx.BlockTime().Unix()

	for i := 0; i < 3; i++ {
		_, err := f.k.ImposeSanction(f.ctx, types.Sanction{
			Subject:       f.target.String(),
			Scope:         types.SanctionScopeAccount,
			Kind:          types.SanctionKindWarning,
			ReasonCode:    types.SanctionReasonTermsViolation,
			Justification: testJustification,
			ExpiresAt:     now + 86400 + int64(i),
		}, f.modA)
		require.NoError(t, err)
	}

	for i := 0; i < 10; i++ {
		state, err := f.k.ProjectAccountState(f.ctx, f.target)
		require.NoError(t, err)
		require.Equal(t, types.AccountStateActive, state)
	}

	// Sequence IDs are allocated monotonically and never reused.
	next := f.k.GetNextSanctionSequence(f.ctx)
	require.Equal(t, uint64(4), next, "three sanctions must consume sequences 1..3")

	all := f.k.GetAllSanctions(f.ctx)
	require.Len(t, all, 3)
}

func TestSanction_UnsactionedAccountIsUnaffected(t *testing.T) {
	f := setupSanctionFixture(t)

	// An account with an administrative state but no sanctions keeps that state.
	other := sdk.AccAddress([]byte("untouched_account_12"))
	require.NoError(t, f.k.SetAccountState(f.ctx, other, types.AccountStateActive, "initial", f.modA))

	state, err := f.k.ProjectAccountState(f.ctx, other)
	require.NoError(t, err)
	require.Equal(t, types.AccountStateActive, state)

	// And a previously terminated account without sanctions is left alone.
	terminated := sdk.AccAddress([]byte("terminated_account_1"))
	require.NoError(t, f.k.SetAccountState(f.ctx, terminated, types.AccountStateActive, "initial", f.modA))
	require.NoError(t, f.k.SetAccountState(f.ctx, terminated, types.AccountStateTerminated, "legacy", f.modA))

	require.NoError(t, f.k.ProcessSanctionExpiry(f.ctx))

	stored, found := f.k.GetAccountState(f.ctx, terminated)
	require.True(t, found)
	require.Equal(t, types.AccountStateTerminated, stored.State,
		"the projection must not resurrect an account with no sanctions")
}

func TestSanction_SubjectIndexIsScoped(t *testing.T) {
	f := setupSanctionFixture(t)
	now := f.ctx.BlockTime().Unix()

	other := sdk.AccAddress([]byte("second_sanction_subj"))

	_, err := f.k.ImposeSanction(f.ctx, suspensionProposal(f.target.String(), now+3600), f.modA)
	require.NoError(t, err)
	_, err = f.k.ImposeSanction(f.ctx, suspensionProposal(other.String(), now+3600), f.modA)
	require.NoError(t, err)

	require.Len(t, f.k.GetSanctionsForSubject(f.ctx, f.target.String()), 1)
	require.Len(t, f.k.GetSanctionsForSubject(f.ctx, other.String()), 1)

	// The projection is per-subject and does not leak across accounts.
	require.Equal(t, types.AccountStateActive, f.k.EffectiveAccountState(f.ctx, other))
	require.Len(t, f.k.GetAllSanctions(f.ctx), 2)
}
