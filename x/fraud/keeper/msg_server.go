// Package keeper implements the Fraud module keeper.
//
// VE-2018: MsgServer implementation for fraud module
// VE-3053: Fixed to use proto-generated types correctly
package keeper

import (
	"context"
	"crypto/sha256"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/fraud/types"
	settlementkeeper "github.com/virtengine/virtengine/x/settlement/keeper"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
)

// Error message constants for msg_server
const (
	errMsgInvalidReporterAddr  = "invalid reporter address"
	errMsgInvalidModeratorAddr = "invalid moderator address"
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
	cacheCtx, write := ctx.CacheContext()
	response, err := ms.submitFraudReport(cacheCtx, msg)
	if err != nil {
		return nil, err
	}
	write()
	return response, nil
}

func (ms *msgServer) submitFraudReport(ctx sdk.Context, msg *types.MsgSubmitFraudReport) (*types.MsgSubmitFraudReportResponse, error) {
	reporterAddr, err := sdk.AccAddressFromBech32(msg.Reporter)
	if err != nil {
		return nil, types.ErrInvalidReporter.Wrap(errMsgInvalidReporterAddr)
	}
	if !ms.keeper.IsProvider(ctx, reporterAddr) {
		return nil, types.ErrUnauthorizedReporter
	}

	evidence := make([]types.EncryptedEvidence, len(msg.Evidence))
	for i, item := range msg.Evidence {
		evidence[i] = types.EncryptedEvidenceFromProto(&item)
	}
	report := &types.FraudReport{
		Reporter: msg.Reporter, ReportedParty: msg.ReportedParty,
		Category: types.FraudCategoryFromProto(msg.Category), Description: msg.Description,
		Evidence: evidence, RelatedOrderIDs: msg.RelatedOrderIds,
		Status: types.FraudReportStatusSubmitted, SubmittedAt: ctx.BlockTime(), UpdatedAt: ctx.BlockTime(), BlockHeight: ctx.BlockHeight(),
	}
	if err := ms.keeper.SubmitFraudReport(ctx, report); err != nil {
		return nil, err
	}

	if len(report.RelatedOrderIDs) > 0 && ms.keeper.financialCases != nil && ms.keeper.financialCases.IsFinancialCasesActive(ctx) {
		orderID := report.RelatedOrderIDs[0]
		evidenceHash := sha256.Sum256([]byte(report.ContentHash))
		idempotencyHash := sha256.Sum256([]byte("fraud/financial-case/v1\x00" + report.ID + "\x00" + orderID))
		idempotency := idempotencyHash[:]
		financialCase, _, _, err := ms.keeper.financialCases.OpenFinancialCase(ctx, settlementkeeper.FinancialCaseOpenRequest{
			Subject:  settlementtypes.FinancialSubject{Type: settlementtypes.FinancialSubjectTypeOrder, PrimaryId: orderID, OrderId: orderID},
			Claimant: report.Reporter, Respondent: report.ReportedParty, IdempotencyKey: idempotency, TrustedAdapter: true,
			Claim: settlementtypes.FinancialClaim{ClaimType: settlementtypes.FinancialClaimTypeFraud, Claimant: report.Reporter, SourceModule: "fraud", SourceReference: report.ID, EvidenceHash: evidenceHash[:], EncryptedReference: "fraud://report/" + report.ID + "/evidence-root/" + fmt.Sprintf("%x", evidenceHash[:]), IdempotencyKey: idempotency},
		})
		if err != nil {
			return nil, err
		}
		report.FinancialCaseID, report.FinancialCaseStatus = financialCase.CaseId, financialCase.Status.String()
		if err := ms.keeper.SetFraudReport(ctx, *report); err != nil {
			return nil, err
		}
	}

	ms.keeper.Logger(ctx).Info("fraud report submitted via message", "report_id", report.ID, "reporter", msg.Reporter, "reported_party", msg.ReportedParty, "category", msg.Category.String())
	return &types.MsgSubmitFraudReportResponse{ReportId: report.ID}, nil
}

// AssignModerator handles assigning a moderator to a fraud report
func (ms *msgServer) AssignModerator(goCtx context.Context, msg *types.MsgAssignModerator) (*types.MsgAssignModeratorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	cacheCtx, write := ctx.CacheContext()
	response, err := ms.assignModerator(cacheCtx, msg)
	if err != nil {
		return nil, err
	}
	write()
	return response, nil
}

func (ms *msgServer) assignModerator(ctx sdk.Context, msg *types.MsgAssignModerator) (*types.MsgAssignModeratorResponse, error) {

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
	cacheCtx, write := ctx.CacheContext()
	response, err := ms.updateReportStatus(cacheCtx, msg)
	if err != nil {
		return nil, err
	}
	write()
	return response, nil
}

func (ms *msgServer) updateReportStatus(ctx sdk.Context, msg *types.MsgUpdateReportStatus) (*types.MsgUpdateReportStatusResponse, error) {

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
	cacheCtx, write := ctx.CacheContext()
	response, err := ms.resolveFraudReport(cacheCtx, msg)
	if err != nil {
		return nil, err
	}
	write()
	return response, nil
}

func (ms *msgServer) resolveFraudReport(ctx sdk.Context, msg *types.MsgResolveFraudReport) (*types.MsgResolveFraudReportResponse, error) {

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

	report, found := ms.keeper.GetFraudReport(ctx, msg.ReportId)
	if !found {
		return nil, types.ErrReportNotFound
	}
	if report.FinancialCaseID != "" && ms.keeper.financialCases == nil {
		return nil, settlementtypes.ErrFinancialCaseEffect.Wrap("fraud financial-case keeper is required")
	}
	canonicalStatus := ""
	if report.FinancialCaseID != "" && ms.keeper.financialCases.IsFinancialCasesActive(ctx) {
		financialCase, caseFound := ms.keeper.financialCases.GetFinancialCase(ctx, report.FinancialCaseID)
		if !caseFound {
			return nil, settlementtypes.ErrFinancialCaseMalformedState.Wrap("fraud projection references missing case")
		}
		if settlementtypes.IsActiveFinancialCaseStatus(financialCase.Status) {
			reasonHash := sha256.Sum256([]byte(msg.Resolution.String() + ":" + msg.Notes))
			if err := ms.keeper.financialCases.EscalateFinancialCase(ctx, report.FinancialCaseID, "adapter:fraud:"+msg.Moderator, reasonHash[:]); err != nil {
				return nil, err
			}
			updatedCase, caseFound := ms.keeper.financialCases.GetFinancialCase(ctx, report.FinancialCaseID)
			if !caseFound {
				return nil, settlementtypes.ErrFinancialCaseMalformedState.Wrap("fraud projection references missing case after escalation")
			}
			report.FinancialCaseStatus = updatedCase.Status.String()
			if err := ms.keeper.SetFraudReport(ctx, report); err != nil {
				return nil, err
			}
			return &types.MsgResolveFraudReportResponse{}, nil
		}
		canonicalStatus = financialCase.Status.String()
		report.FinancialCaseStatus = canonicalStatus
		if err := ms.keeper.SetFraudReport(ctx, report); err != nil {
			return nil, err
		}
	}
	// Resolve the fraud/reputation projection only; settlement remains financial authority.
	if err := ms.keeper.ResolveFraudReport(ctx, msg.ReportId, resolution, msg.Notes, msg.Moderator); err != nil {
		return nil, err
	}
	if canonicalStatus != "" {
		report, found = ms.keeper.GetFraudReport(ctx, msg.ReportId)
		if !found {
			return nil, types.ErrReportNotFound
		}
		report.FinancialCaseStatus = canonicalStatus
		if err := ms.keeper.SetFraudReport(ctx, report); err != nil {
			return nil, err
		}
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
	cacheCtx, write := ctx.CacheContext()
	response, err := ms.rejectFraudReport(cacheCtx, msg)
	if err != nil {
		return nil, err
	}
	write()
	return response, nil
}

func (ms *msgServer) rejectFraudReport(ctx sdk.Context, msg *types.MsgRejectFraudReport) (*types.MsgRejectFraudReportResponse, error) {

	// Validate moderator address
	moderatorAddr, err := sdk.AccAddressFromBech32(msg.Moderator)
	if err != nil {
		return nil, types.ErrUnauthorizedModerator.Wrap(errMsgInvalidModeratorAddr)
	}

	// Check if moderator has permission
	if !ms.keeper.IsModerator(ctx, moderatorAddr) {
		return nil, types.ErrUnauthorizedModerator.Wrap("sender is not a moderator")
	}
	report, found := ms.keeper.GetFraudReport(ctx, msg.ReportId)
	if !found {
		return nil, types.ErrReportNotFound
	}
	if report.FinancialCaseID != "" {
		if ms.keeper.financialCases == nil {
			return nil, settlementtypes.ErrFinancialCaseEffect.Wrap("fraud financial-case keeper is required")
		}
		if ms.keeper.financialCases.IsFinancialCasesActive(ctx) {
			financialCase, caseFound := ms.keeper.financialCases.GetFinancialCase(ctx, report.FinancialCaseID)
			if !caseFound {
				return nil, settlementtypes.ErrFinancialCaseMalformedState.Wrap("fraud projection references missing case")
			}
			if settlementtypes.IsActiveFinancialCaseStatus(financialCase.Status) {
				return nil, settlementtypes.ErrLegacyFinancialMutationFenced.Wrap("canonical financial case remains active")
			}
		}
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
	cacheCtx, write := ctx.CacheContext()
	response, err := ms.escalateFraudReport(cacheCtx, msg)
	if err != nil {
		return nil, err
	}
	write()
	return response, nil
}

func (ms *msgServer) escalateFraudReport(ctx sdk.Context, msg *types.MsgEscalateFraudReport) (*types.MsgEscalateFraudReportResponse, error) {

	// Validate moderator address
	moderatorAddr, err := sdk.AccAddressFromBech32(msg.Moderator)
	if err != nil {
		return nil, types.ErrUnauthorizedModerator.Wrap(errMsgInvalidModeratorAddr)
	}

	// Check if moderator has permission
	if !ms.keeper.IsModerator(ctx, moderatorAddr) {
		return nil, types.ErrUnauthorizedModerator.Wrap("sender is not a moderator")
	}

	report, found := ms.keeper.GetFraudReport(ctx, msg.ReportId)
	if !found {
		return nil, types.ErrReportNotFound
	}
	if report.FinancialCaseID != "" && ms.keeper.financialCases == nil {
		return nil, settlementtypes.ErrFinancialCaseEffect.Wrap("fraud financial-case keeper is required")
	}
	if report.FinancialCaseID != "" && ms.keeper.financialCases.IsFinancialCasesActive(ctx) {
		reasonHash := sha256.Sum256([]byte(msg.Reason))
		if err := ms.keeper.financialCases.EscalateFinancialCase(ctx, report.FinancialCaseID, "adapter:fraud:"+msg.Moderator, reasonHash[:]); err != nil {
			return nil, err
		}
	}
	// Escalate fraud projection after the canonical case transition succeeds.
	if err := ms.keeper.EscalateFraudReport(ctx, msg.ReportId, msg.Reason, msg.Moderator); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("fraud report escalated via message",
		"report_id", msg.ReportId,
		"moderator", msg.Moderator,
	)

	return &types.MsgEscalateFraudReportResponse{}, nil
}

// UpdateParams handles updating module parameters (governance only)
func (ms *msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

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
