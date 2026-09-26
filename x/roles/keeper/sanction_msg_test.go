package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/roles/keeper"
	"github.com/virtengine/virtengine/x/roles/types"
)

// ---------------------------------------------------------------------------
// On-chain reachability of the sanction model.
//
// The keeper state machine is only meaningful if a transaction can drive it.
// These tests exercise the MsgServer path end to end: a moderator submits
// MsgImposeSanction, a second distinct moderator submits MsgConfirmSanction,
// and the account-state projection follows. They are the regression guard for
// the defect where every keeper transition existed but had no message handler,
// so no account could be suspended or terminated by any means.
// ---------------------------------------------------------------------------

// newMsgServerFixture builds a msgServer over the sanction fixture's keeper.
func newMsgServerFixture(t *testing.T) (sanctionFixture, types.MsgServer) {
	t.Helper()
	f := setupSanctionFixture(t)
	return f, keeper.NewMsgServerImpl(f.k)
}

// asGoCtx converts an sdk.Context into the context.Context the MsgServer takes.
func asGoCtx(ctx sdk.Context) context.Context { return sdk.WrapSDKContext(ctx) }

func TestMsgServer_ImposeSanctionIsRecordedPendingAndDoesNotBind(t *testing.T) {
	f, ms := newMsgServerFixture(t)
	now := f.ctx.BlockTime().Unix()

	resp, err := ms.ImposeSanction(asGoCtx(f.ctx), types.NewMsgImposeSanction(
		f.modA.String(),
		f.target.String(),
		types.SanctionScopeAccount.String(),
		"",
		types.SanctionKindSuspension.String(),
		types.SanctionReasonFraudConfirmed.String(),
		testJustification,
		"You are suspended. You may appeal within 30 days.",
		3600,
	))
	require.NoError(t, err)

	// The response must not claim the sanction is in force.
	require.Equal(t, types.SanctionStatusPendingReview.String(), resp.Status,
		"a single moderator's suspension must be recorded pending review")
	require.NotEmpty(t, resp.SanctionID)

	record, found := f.k.GetSanction(f.ctx, resp.SanctionID)
	require.True(t, found)
	require.Equal(t, types.SanctionStatusPendingReview, record.Status)
	require.False(t, record.InForce(), "a pending suspension must not be in force")
	require.Equal(t, now+3600, record.ExpiresAt,
		"duration_seconds must resolve against block time")

	// The account state projection must not have suspended anyone.
	require.Equal(t, types.AccountStateActive, f.k.EffectiveAccountState(f.ctx, f.target))
}

func TestMsgServer_SanctionCannotBeConfirmedByItsOwnImposer(t *testing.T) {
	f, ms := newMsgServerFixture(t)

	resp, err := ms.ImposeSanction(asGoCtx(f.ctx), types.NewMsgImposeSanction(
		f.modA.String(), f.target.String(),
		types.SanctionScopeAccount.String(), "",
		types.SanctionKindSuspension.String(),
		types.SanctionReasonFraudConfirmed.String(),
		testJustification, "You are suspended.", 3600,
	))
	require.NoError(t, err)

	// The imposing moderator cannot co-sign their own suspension.
	_, err = ms.ConfirmSanction(asGoCtx(f.ctx),
		types.NewMsgConfirmSanction(f.modA.String(), resp.SanctionID, 0))
	require.Error(t, err, "a single moderator must not be able to confirm their own sanction")

	record, found := f.k.GetSanction(f.ctx, resp.SanctionID)
	require.True(t, found)
	require.Equal(t, types.SanctionStatusPendingReview, record.Status,
		"a refused confirmation must leave the sanction pending, not active")
	require.Equal(t, types.AccountStateActive, f.k.EffectiveAccountState(f.ctx, f.target))
}

func TestMsgServer_SecondDistinctReviewerActivatesTheSanction(t *testing.T) {
	f, ms := newMsgServerFixture(t)
	now := f.ctx.BlockTime().Unix()

	resp, err := ms.ImposeSanction(asGoCtx(f.ctx), types.NewMsgImposeSanction(
		f.modA.String(), f.target.String(),
		types.SanctionScopeAccount.String(), "",
		types.SanctionKindSuspension.String(),
		types.SanctionReasonFraudConfirmed.String(),
		testJustification, "You are suspended.", 3600,
	))
	require.NoError(t, err)

	_, err = ms.ConfirmSanction(asGoCtx(f.ctx),
		types.NewMsgConfirmSanction(f.modB.String(), resp.SanctionID, 0))
	require.NoError(t, err, "a distinct second reviewer must be able to confirm")

	record, found := f.k.GetSanction(f.ctx, resp.SanctionID)
	require.True(t, found)
	require.Equal(t, types.SanctionStatusActive, record.Status)
	require.Equal(t, f.modB.String(), record.SecondReviewer)
	require.Equal(t, types.AccountStateSuspended, f.k.EffectiveAccountState(f.ctx, f.target),
		"a confirmed suspension must project a suspended account state")

	// And the expiry path still restores Active with no manual intervention.
	expired := f.at(now + 3601)
	require.NoError(t, f.k.ProcessSanctionExpiry(expired.ctx))
	require.Equal(t, types.AccountStateActive, f.k.EffectiveAccountState(expired.ctx, f.target),
		"an expired suspension must restore Active without manual intervention")
}

// A pending emergency hold binds immediately, so it must also be liftable
// immediately. This is the defect where RevokeSanction guarded on the record's
// *status* (Active only) rather than on whether it was *in force*, leaving one
// moderator able to suspend an account for the full 72h window with no way out.
func TestMsgServer_PendingEmergencyHoldCanBeRevokedByAnyModerator(t *testing.T) {
	f, ms := newMsgServerFixture(t)
	now := f.ctx.BlockTime().Unix()

	resp, err := ms.ImposeSanction(asGoCtx(f.ctx), types.NewMsgImposeSanction(
		f.modA.String(), f.target.String(),
		types.SanctionScopeAccount.String(), "",
		types.SanctionKindEmergencyHold.String(),
		types.SanctionReasonFraudSuspected.String(),
		testJustification, "Emergency hold pending review.", 0,
	))
	require.NoError(t, err)

	record, found := f.k.GetSanction(f.ctx, resp.SanctionID)
	require.True(t, found)
	require.Equal(t, types.SanctionStatusPendingReview, record.Status)
	require.True(t, record.InForce(), "a pending emergency hold binds immediately")
	require.Equal(t, types.AccountStateSuspended, f.k.EffectiveAccountState(f.ctx, f.target))

	// A different moderator must be able to lift it before the window closes.
	_, err = ms.RevokeSanction(asGoCtx(f.ctx),
		types.NewMsgRevokeSanction(f.modB.String(), resp.SanctionID, "hold was mistaken"))
	require.NoError(t, err, "a pending emergency hold must be revocable before it expires")

	revoked, found := f.k.GetSanction(f.ctx, resp.SanctionID)
	require.True(t, found)
	require.Equal(t, types.SanctionStatusRevoked, revoked.Status)
	require.Equal(t, types.AccountStateActive, f.k.EffectiveAccountState(f.ctx, f.target),
		"revoking a mistaken hold must restore the account immediately")

	// An expired hold is not in force, so it is no longer revocable: the record
	// is already terminal and the account is Active either way.
	_ = now
}

func TestMsgServer_AppealRoundTripRestoresAccountThroughReview(t *testing.T) {
	f, ms := newMsgServerFixture(t)

	resp, err := ms.ImposeSanction(asGoCtx(f.ctx), types.NewMsgImposeSanction(
		f.modA.String(), f.target.String(),
		types.SanctionScopeAccount.String(), "",
		types.SanctionKindSuspension.String(),
		types.SanctionReasonFraudConfirmed.String(),
		testJustification, "You are suspended.", 3600,
	))
	require.NoError(t, err)
	_, err = ms.ConfirmSanction(asGoCtx(f.ctx),
		types.NewMsgConfirmSanction(f.modB.String(), resp.SanctionID, 0))
	require.NoError(t, err)
	require.Equal(t, types.AccountStateSuspended, f.k.EffectiveAccountState(f.ctx, f.target))

	// Only the sanctioned party may appeal.
	_, err = ms.OpenSanctionAppeal(asGoCtx(f.ctx),
		types.NewMsgOpenSanctionAppeal(f.modC.String(), resp.SanctionID, "not my sanction to answer"))
	require.Error(t, err, "only the sanctioned party may appeal")

	appeal, err := ms.OpenSanctionAppeal(asGoCtx(f.ctx),
		types.NewMsgOpenSanctionAppeal(f.target.String(), resp.SanctionID,
			"The evidence relied on was already reviewed and rejected."))
	require.NoError(t, err)
	require.NotEmpty(t, appeal.AppealID)

	// A participant in the challenged decision may not decide the appeal.
	_, err = ms.ResolveSanctionAppeal(asGoCtx(f.ctx),
		types.NewMsgResolveSanctionAppeal(f.modA.String(), appeal.AppealID, true, "grant"))
	require.Error(t, err, "the imposer of the challenged sanction may not decide its appeal")
	_, err = ms.ResolveSanctionAppeal(asGoCtx(f.ctx),
		types.NewMsgResolveSanctionAppeal(f.modB.String(), appeal.AppealID, true, "grant"))
	require.Error(t, err, "the second reviewer may not decide the appeal they confirmed")

	// An uninvolved moderator may, and granting it revokes the sanction.
	_, err = ms.ResolveSanctionAppeal(asGoCtx(f.ctx),
		types.NewMsgResolveSanctionAppeal(f.modC.String(), appeal.AppealID, true, "appeal allowed"))
	require.NoError(t, err)
	require.Equal(t, types.AccountStateActive, f.k.EffectiveAccountState(f.ctx, f.target),
		"a granted appeal must restore the account through review")
}

func TestMsgServer_ModeratorCannotSanctionItself(t *testing.T) {
	f, ms := newMsgServerFixture(t)

	_, err := ms.ImposeSanction(asGoCtx(f.ctx), types.NewMsgImposeSanction(
		f.modA.String(), f.modA.String(),
		types.SanctionScopeAccount.String(), "",
		types.SanctionKindTermination.String(),
		types.SanctionReasonTermsViolation.String(),
		testJustification, "Notice of termination.", 0,
	))
	require.Error(t, err, "a moderator must not be able to sanction itself")
}

func TestMsgServer_ImposeSanctionRejectsUnknownEnumStrings(t *testing.T) {
	f, ms := newMsgServerFixture(t)

	for _, tc := range []struct {
		name                              string
		scope, kind, reason, justification string
	}{
		{"unknown scope", "planet", types.SanctionKindWarning.String(), types.SanctionReasonFraudConfirmed.String(), testJustification},
		{"unknown kind", types.SanctionScopeAccount.String(), "exile", types.SanctionReasonFraudConfirmed.String(), testJustification},
		{"unknown reason", types.SanctionScopeAccount.String(), types.SanctionKindWarning.String(), "because", testJustification},
		{"short justification", types.SanctionScopeAccount.String(), types.SanctionKindWarning.String(), types.SanctionReasonFraudConfirmed.String(), "too short"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ms.ImposeSanction(asGoCtx(f.ctx), types.NewMsgImposeSanction(
				f.modA.String(), f.target.String(),
				tc.scope, "", tc.kind, tc.reason, tc.justification, "", 0,
			))
			require.Error(t, err)
		})
	}
}
