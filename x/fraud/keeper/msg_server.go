// Package keeper implements the Fraud module keeper.
//
// VE-2018: MsgServer implementation for fraud module
// VE-3053: Fixed to use proto-generated types correctly
package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/fraud/types"
)

// Error message constants for msg_server
const (
	errMsgInvalidReporterAddr  = "invalid reporter address"
	errMsgInvalidModeratorAddr = "invalid moderator address"
	errMsgInvalidAuthorityAddr = "invalid authority address"
)

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of the fraud MsgServer interface
// This returns a MsgServerImpl that can be wrapped by RegisterMsgServer.
func NewMsgServerImpl(k Keeper) types.MsgServerImpl {
	return &msgServer{keeper: k}
}

var _ types.MsgServerImpl = (*msgServer)(nil)

// SubmitFraudReport handles submitting a new fraud report
func (ms *msgServer) SubmitFraudReport(goCtx context.Context, msg *types.MsgSubmitFraudReport) (*types.MsgSubmitFraudReportResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate reporter address
	reporterAddr, err := sdk.AccAddressFromBech32(msg.Reporter)
	if err != nil {
		return nil, types.ErrInvalidReporter.Wrap(errMsgInvalidReporterAddr)
	}

	// Check if reporter is a provider
	if !ms.keeper.IsProvider(ctx, reporterAddr) {
		return nil, types.ErrUnauthorizedReporter
	}

	// Convert proto evidence to local type
	evidence := make([]types.EncryptedEvidence, len(msg.Evidence))
	for i, e := range msg.Evidence {
		evidence[i] = types.EncryptedEvidenceFromProto(&e)
	}

	// Create the fraud report using local types
	report := &types.FraudReport{
		Reporter:        msg.Reporter,
		ReportedParty:   msg.ReportedParty,
		Category:        types.FraudCategoryFromProto(msg.Category),
		Description:     msg.Description,
		Evidence:        evidence,
		RelatedOrderIDs: msg.RelatedOrderIds,
		Status:          types.FraudReportStatusSubmitted,
		SubmittedAt:     ctx.BlockTime(),
		UpdatedAt:       ctx.BlockTime(),
		BlockHeight:     ctx.BlockHeight(),
	}

	// Submit the report through the keeper
	if err := ms.keeper.SubmitFraudReport(ctx, report); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("fraud report submitted via message",
		"report_id", report.ID,
		"reporter", msg.Reporter,
		"reported_party", msg.ReportedParty,
		"category", msg.Category.String(),
	)

	return &types.MsgSubmitFraudReportResponse{
		ReportId: report.ID,
	}, nil
}

// AssignModerator handles assigning a moderator to a fraud report
func (ms *msgServer) AssignModerator(goCtx context.Context, msg *types.MsgAssignModerator) (*types.MsgAssignModeratorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate moderator address
	moderatorAddr, err := sdk.AccAddressFromBech32(msg.Moderator)
	if err != nil {
		return nil, types.ErrUnauthorizedModerator.Wrap(errMsgInvalidModeratorAddr)
	}

	// Check if moderator has permission
	if !ms.keeper.IsModerator(ctx, moderatorAddr) {
		return nil, types.ErrUnauthorizedModerator.Wrap("sender is not a moderator")
	}

	// Assign moderator through the keeper
	if err := ms.keeper.AssignModerator(ctx, msg.ReportId, msg.AssignTo); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("moderator assigned via message",
		"report_id", msg.ReportId,
		"moderator", msg.Moderator,
		"assigned_to", msg.AssignTo,
	)

	return &types.MsgAssignModeratorResponse{}, nil
}

// UpdateReportStatus handles updating the status of a fraud report
func (ms *msgServer) UpdateReportStatus(goCtx context.Context, msg *types.MsgUpdateReportStatus) (*types.MsgUpdateReportStatusResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate moderator address
	moderatorAddr, err := sdk.AccAddressFromBech32(msg.Moderator)
	if err != nil {
		return nil, types.ErrUnauthorizedModerator.Wrap(errMsgInvalidModeratorAddr)
	}

	// Check if moderator has permission
	if !ms.keeper.IsModerator(ctx, moderatorAddr) {
		return nil, types.ErrUnauthorizedModerator.Wrap("sender is not a moderator")
	}

	// Convert proto status to local type
	newStatus := types.FraudReportStatusFromProto(msg.NewStatus)

	// Update status through the keeper
	if err := ms.keeper.UpdateReportStatus(ctx, msg.ReportId, newStatus, msg.Moderator, msg.Notes); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("report status updated via message",
		"report_id", msg.ReportId,
		"moderator", msg.Moderator,
		"new_status", msg.NewStatus.String(),
	)

	return &types.MsgUpdateReportStatusResponse{}, nil
}

// ResolveFraudReport handles resolving a fraud report
func (ms *msgServer) ResolveFraudReport(goCtx context.Context, msg *types.MsgResolveFraudReport) (*types.MsgResolveFraudReportResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate moderator address
	moderatorAddr, err := sdk.AccAddressFromBech32(msg.Moderator)
	if err != nil {
		return nil, types.ErrUnauthorizedModerator.Wrap(errMsgInvalidModeratorAddr)
	}

	// Check if moderator has permission
	if !ms.keeper.IsModerator(ctx, moderatorAddr) {
		return nil, types.ErrUnauthorizedModerator.Wrap("sender is not a moderator")
	}

	// Convert proto resolution to local type
	resolution := types.ResolutionTypeFromProto(msg.Resolution)

	// Resolve report through the keeper
	if err := ms.keeper.ResolveFraudReport(ctx, msg.ReportId, resolution, msg.Notes, msg.Moderator); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("fraud report resolved via message",
		"report_id", msg.ReportId,
		"moderator", msg.Moderator,
		"resolution", msg.Resolution.String(),
	)

	return &types.MsgResolveFraudReportResponse{}, nil
}

// RejectFraudReport handles rejecting a fraud report
func (ms *msgServer) RejectFraudReport(goCtx context.Context, msg *types.MsgRejectFraudReport) (*types.MsgRejectFraudReportResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate moderator address
	moderatorAddr, err := sdk.AccAddressFromBech32(msg.Moderator)
	if err != nil {
		return nil, types.ErrUnauthorizedModerator.Wrap(errMsgInvalidModeratorAddr)
	}

	// Check if moderator has permission
	if !ms.keeper.IsModerator(ctx, moderatorAddr) {
		return nil, types.ErrUnauthorizedModerator.Wrap("sender is not a moderator")
	}

	// Reject report through the keeper
	if err := ms.keeper.RejectFraudReport(ctx, msg.ReportId, msg.Notes, msg.Moderator); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("fraud report rejected via message",
		"report_id", msg.ReportId,
		"moderator", msg.Moderator,
	)

	return &types.MsgRejectFraudReportResponse{}, nil
}

// EscalateFraudReport handles escalating a fraud report
func (ms *msgServer) EscalateFraudReport(goCtx context.Context, msg *types.MsgEscalateFraudReport) (*types.MsgEscalateFraudReportResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate moderator address
	moderatorAddr, err := sdk.AccAddressFromBech32(msg.Moderator)
	if err != nil {
		return nil, types.ErrUnauthorizedModerator.Wrap(errMsgInvalidModeratorAddr)
	}

	// Check if moderator has permission
	if !ms.keeper.IsModerator(ctx, moderatorAddr) {
		return nil, types.ErrUnauthorizedModerator.Wrap("sender is not a moderator")
	}

	// Escalate report through the keeper
	if err := ms.keeper.EscalateFraudReport(ctx, msg.ReportId, msg.Reason, msg.Moderator); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("fraud report escalated via message",
		"report_id", msg.ReportId,
		"moderator", msg.Moderator,
		"reason", msg.Reason,
	)

	return &types.MsgEscalateFraudReportResponse{}, nil
}

// UpdateParams handles updating module parameters (governance only)
func (ms *msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate authority address
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return nil, types.ErrInvalidReporter.Wrap(errMsgInvalidAuthorityAddr)
	}

	// Verify authority matches module authority
	if msg.Authority != ms.keeper.GetAuthority() {
		return nil, types.ErrUnauthorizedModerator.Wrap("unauthorized: sender is not the module authority")
	}

	// Convert proto params to local type
	params := types.ParamsFromProto(&msg.Params)

	// Update params through the keeper
	if err := ms.keeper.SetParams(ctx, *params); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("fraud module params updated via message",
		"authority", msg.Authority,
	)

	return &types.MsgUpdateParamsResponse{}, nil
}
