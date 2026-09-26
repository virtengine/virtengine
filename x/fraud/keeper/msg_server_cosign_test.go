package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/mock"

	"github.com/virtengine/virtengine/x/fraud/types"
)

// t_9169400a: A suspension or termination arriving on the wire must not lock an
// account out on the strength of one moderator. These tests exercise the
// message path (not just the keeper) because that is what a moderator actually
// submits.

// submitReportForCoSign submits a report and returns its ID.
func (s *MsgServerTestSuite) submitReportForCoSign(
	reporter sdk.AccAddress, reported string,
) string {
	s.providerKeeper.On("IsProvider", mock.Anything, reporter).Return(true)
	s.rolesKeeper.On("HasRole", mock.Anything, reporter, mock.Anything).Return(true)

	resp, err := s.msgServer.SubmitFraudReport(s.ctx, &types.MsgSubmitFraudReport{
		Reporter:      reporter.String(),
		ReportedParty: reported,
		Category:      types.FraudCategoryPBPaymentFraud,
		Description:   "Payment fraud description for co-sign review",
		Evidence:      validEvidence(),
	})
	s.Require().NoError(err)
	return resp.ReportId
}

// TestResolveFraudReport_SuspensionProposesInsteadOfApplying pins the wire
// behaviour: one moderator resolving a suspension records a proposal awaiting a
// distinct second reviewer and leaves the report unresolved.
func (s *MsgServerTestSuite) TestResolveFraudReport_SuspensionProposesInsteadOfApplying() {
	reporterAddr := sdk.AccAddress([]byte("reporter-susp-cosign"))
	moderatorAddr := sdk.AccAddress([]byte("moderator-susp-cos"))
	secondAddr := sdk.AccAddress([]byte("moderator-susp-two"))

	reportID := s.submitReportForCoSign(reporterAddr, "cosmos1reported")
	s.rolesKeeper.On("IsModerator", mock.Anything, moderatorAddr).Return(true)
	s.rolesKeeper.On("IsModerator", mock.Anything, secondAddr).Return(true)

	resp, err := s.msgServer.ResolveFraudReport(s.ctx, &types.MsgResolveFraudReport{
		ReportId:   reportID,
		Moderator:  moderatorAddr.String(),
		Resolution: types.ResolutionTypePBSuspension,
		Notes:      "Suspending pending second review",
	})
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	// The report is not resolved by one moderator.
	report, found := s.keeper.GetFraudReport(s.ctx, reportID)
	s.Require().True(found)
	s.Require().NotEqual(types.FraudReportStatusResolved, report.Status,
		"one moderator must not resolve a suspension on the wire")
	s.Require().NotEqual(types.ResolutionTypeSuspension, report.Resolution)

	// A proposal is waiting for a distinct reviewer.
	pending, found := s.keeper.GetPendingResolution(s.ctx, reportID)
	s.Require().True(found, "a pending resolution should be recorded")
	s.Require().Equal(moderatorAddr.String(), pending.ProposedBy)
	s.Require().Equal(types.ResolutionTypeSuspension, pending.Resolution)

	// The proposing moderator cannot confirm it.
	_, err = s.keeper.ConfirmResolution(s.ctx, reportID, moderatorAddr.String())
	s.Require().ErrorIs(err, types.ErrSecondReviewerMustDiffer)

	// A distinct moderator can, and only then does it take effect.
	_, err = s.keeper.ConfirmResolution(s.ctx, reportID, secondAddr.String())
	s.Require().NoError(err)

	report, found = s.keeper.GetFraudReport(s.ctx, reportID)
	s.Require().True(found)
	s.Require().Equal(types.FraudReportStatusResolved, report.Status)
	s.Require().Equal(types.ResolutionTypeSuspension, report.Resolution)
}

// TestResolveFraudReport_TerminationProposesInsteadOfApplying is the same
// guarantee for termination, the harshest outcome in the enum.
func (s *MsgServerTestSuite) TestResolveFraudReport_TerminationProposesInsteadOfApplying() {
	reporterAddr := sdk.AccAddress([]byte("reporter-term-cosign"))
	moderatorAddr := sdk.AccAddress([]byte("moderator-term-cos"))

	reportID := s.submitReportForCoSign(reporterAddr, "cosmos1reported")
	s.rolesKeeper.On("IsModerator", mock.Anything, moderatorAddr).Return(true)

	_, err := s.msgServer.ResolveFraudReport(s.ctx, &types.MsgResolveFraudReport{
		ReportId:   reportID,
		Moderator:  moderatorAddr.String(),
		Resolution: types.ResolutionTypePBTermination,
		Notes:      "Terminating",
	})
	s.Require().NoError(err)

	report, found := s.keeper.GetFraudReport(s.ctx, reportID)
	s.Require().True(found)
	s.Require().NotEqual(types.FraudReportStatusResolved, report.Status)
	s.Require().NotEqual(types.ResolutionTypeTermination, report.Resolution)

	pending, found := s.keeper.GetPendingResolution(s.ctx, reportID)
	s.Require().True(found)
	s.Require().Equal(types.ResolutionTypeTermination, pending.Resolution)
}

// TestResolveFraudReport_ResponseTellsTheCallerItIsOnlyPending pins the wire
// contract. The proposing moderator must not be told, by a successful
// transaction, that a suspension was applied when nothing was applied at all:
// the response has to say the resolution is pending a second reviewer.
func (s *MsgServerTestSuite) TestResolveFraudReport_ResponseTellsTheCallerItIsOnlyPending() {
	reporterAddr := sdk.AccAddress([]byte("reporter-resp-cosign"))
	moderatorAddr := sdk.AccAddress([]byte("moderator-resp-cos"))

	reportID := s.submitReportForCoSign(reporterAddr, "cosmos1reported")
	s.rolesKeeper.On("IsModerator", mock.Anything, moderatorAddr).Return(true)

	resp, err := s.msgServer.ResolveFraudReport(s.ctx, &types.MsgResolveFraudReport{
		ReportId:   reportID,
		Moderator:  moderatorAddr.String(),
		Resolution: types.ResolutionTypePBSuspension,
		Notes:      "Suspending pending second review",
	})
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().True(resp.RequiresSecondReviewer,
		"the response must report that a second reviewer is required")
	s.Require().True(resp.Pending,
		"the response must report that nothing was applied yet")
	s.Require().NotZero(resp.PendingExpiresAt,
		"the response must carry the lapse time of the proposal")

	pending, found := s.keeper.GetPendingResolution(s.ctx, reportID)
	s.Require().True(found)
	s.Require().Equal(resp.PendingExpiresAt, pending.ExpiresAt)
}

// A warning needs no co-signature, so its response must NOT claim to be
// pending. This is the other half of the honesty contract.
func (s *MsgServerTestSuite) TestResolveFraudReport_AppliedResponseIsNotMarkedPending() {
	reporterAddr := sdk.AccAddress([]byte("reporter-applied-resp"))
	moderatorAddr := sdk.AccAddress([]byte("moderator-applied-cos"))

	reportID := s.submitReportForCoSign(reporterAddr, "cosmos1reported")
	s.rolesKeeper.On("IsModerator", mock.Anything, moderatorAddr).Return(true)

	resp, err := s.msgServer.ResolveFraudReport(s.ctx, &types.MsgResolveFraudReport{
		ReportId:   reportID,
		Moderator:  moderatorAddr.String(),
		Resolution: types.ResolutionTypePBWarning,
		Notes:      "Warning issued",
	})
	s.Require().NoError(err)
	s.Require().False(resp.Pending, "an applied resolution must not be marked pending")
	s.Require().False(resp.RequiresSecondReviewer)
}

// TestConfirmFraudResolution_AppliesTheProposedSuspension is the regression
// guard for the defect where the co-signature had no message at all: the
// proposal was recorded, the transaction reported success, and the only
// ConfirmResolution caller in the tree was a test. This drives the same flow
// entirely through the MsgServer.
func (s *MsgServerTestSuite) TestConfirmFraudResolution_AppliesTheProposedSuspension() {
	reporterAddr := sdk.AccAddress([]byte("reporter-conf-apply"))
	moderatorAddr := sdk.AccAddress([]byte("moderator-conf-one"))
	secondAddr := sdk.AccAddress([]byte("moderator-conf-two"))

	reportID := s.submitReportForCoSign(reporterAddr, "cosmos1reported")
	s.rolesKeeper.On("IsModerator", mock.Anything, moderatorAddr).Return(true)
	s.rolesKeeper.On("IsModerator", mock.Anything, secondAddr).Return(true)

	_, err := s.msgServer.ResolveFraudReport(s.ctx, &types.MsgResolveFraudReport{
		ReportId:   reportID,
		Moderator:  moderatorAddr.String(),
		Resolution: types.ResolutionTypePBSuspension,
		Notes:      "Suspending pending second review",
	})
	s.Require().NoError(err)

	// The proposer cannot co-sign.
	_, err = s.msgServer.ConfirmFraudResolution(s.ctx, &types.MsgConfirmFraudResolution{
		ReportId: reportID,
		Reviewer: moderatorAddr.String(),
	})
	s.Require().Error(err, "the proposing moderator must not confirm their own resolution")

	report, found := s.keeper.GetFraudReport(s.ctx, reportID)
	s.Require().True(found)
	s.Require().NotEqual(types.FraudReportStatusResolved, report.Status,
		"a refused confirmation must leave the report unresolved")

	// A distinct moderator can, and only then does the resolution apply.
	resp, err := s.msgServer.ConfirmFraudResolution(s.ctx, &types.MsgConfirmFraudResolution{
		ReportId: reportID,
		Reviewer: secondAddr.String(),
	})
	s.Require().NoError(err, "a distinct second reviewer must be able to confirm on the wire")
	s.Require().Equal(types.ResolutionTypePBSuspension, resp.Resolution)

	report, found = s.keeper.GetFraudReport(s.ctx, reportID)
	s.Require().True(found)
	s.Require().Equal(types.FraudReportStatusResolved, report.Status)
	s.Require().Equal(types.ResolutionTypeSuspension, report.Resolution)

	// The proposal is consumed, so it cannot be confirmed twice.
	_, found = s.keeper.GetPendingResolution(s.ctx, reportID)
	s.Require().False(found, "a confirmed proposal must be cleared")
}

// TestConfirmFraudResolution_UnknownReportFails proves the new message does not
// report success for a report that has no proposal.
func (s *MsgServerTestSuite) TestConfirmFraudResolution_UnknownReportFails() {
	reviewer := sdk.AccAddress([]byte("moderator-conf-missing"))
	s.rolesKeeper.On("IsModerator", mock.Anything, reviewer).Return(true)

	_, err := s.msgServer.ConfirmFraudResolution(s.ctx, &types.MsgConfirmFraudResolution{
		ReportId: "no-such-report",
		Reviewer: reviewer.String(),
	})
	s.Require().Error(err, "confirming a report with no proposal must fail")
}

// TestResolveFraudReport_WarningStillAppliesImmediately guards against the
// co-signature rule over-reaching: advisory resolutions are unaffected.
func (s *MsgServerTestSuite) TestResolveFraudReport_WarningStillAppliesImmediately() {
	reporterAddr := sdk.AccAddress([]byte("reporter-warn-cosign"))
	moderatorAddr := sdk.AccAddress([]byte("moderator-warn-cos"))

	reportID := s.submitReportForCoSign(reporterAddr, "cosmos1reported")
	s.rolesKeeper.On("IsModerator", mock.Anything, moderatorAddr).Return(true)

	_, err := s.msgServer.ResolveFraudReport(s.ctx, &types.MsgResolveFraudReport{
		ReportId:   reportID,
		Moderator:  moderatorAddr.String(),
		Resolution: types.ResolutionTypePBWarning,
		Notes:      "Warning issued",
	})
	s.Require().NoError(err)

	report, found := s.keeper.GetFraudReport(s.ctx, reportID)
	s.Require().True(found)
	s.Require().Equal(types.FraudReportStatusResolved, report.Status)
	s.Require().Equal(types.ResolutionTypeWarning, report.Resolution)

	// No co-signature is required for it.
	_, found = s.keeper.GetPendingResolution(s.ctx, reportID)
	s.Require().False(found, "a warning must not require a second reviewer")
}
