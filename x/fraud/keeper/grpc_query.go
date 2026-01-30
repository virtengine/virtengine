// Package keeper implements the Fraud module keeper.
//
// VE-912: Fraud reporting flow - gRPC query handlers
// VE-3053: Fixed to implement proto-generated QueryServer interface
package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/fraud/types"
)

// queryServer implements the gRPC query service
type queryServer struct {
	Keeper
}

// NewQueryServer creates a new gRPC query server
func NewQueryServer(k Keeper) types.QueryServerImpl {
	return &queryServer{Keeper: k}
}

var _ types.QueryServerImpl = (*queryServer)(nil)

// Params returns the module parameters
func (q *queryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := q.GetParams(sdkCtx)

	return &types.QueryParamsResponse{
		Params: *types.ParamsToProto(&params),
	}, nil
}

// FraudReport returns a fraud report by ID
func (q *queryServer) FraudReport(ctx context.Context, req *types.QueryFraudReportRequest) (*types.QueryFraudReportResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidReportID.Wrap("request is nil")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	report, found := q.GetFraudReport(sdkCtx, req.ReportId)
	if !found {
		return nil, types.ErrReportNotFound
	}

	return &types.QueryFraudReportResponse{
		Report: *types.FraudReportToProto(&report),
	}, nil
}

// FraudReports returns all fraud reports with optional filters
func (q *queryServer) FraudReports(ctx context.Context, req *types.QueryFraudReportsRequest) (*types.QueryFraudReportsResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidReportID.Wrap("request is nil")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	var localReports []types.FraudReport

	// Apply filters
	if req.Status != types.FraudReportStatusPBUnspecified {
		status := types.FraudReportStatusFromProto(req.Status)
		localReports = q.GetFraudReportsByStatus(sdkCtx, status)
	} else {
		// Get all reports
		q.WithFraudReports(sdkCtx, func(report types.FraudReport) bool {
			// Apply category filter if specified
			if req.Category != types.FraudCategoryPBUnspecified {
				if report.Category == types.FraudCategoryFromProto(req.Category) {
					localReports = append(localReports, report)
				}
			} else {
				localReports = append(localReports, report)
			}
			return false
		})
	}

	// Convert to proto types
	protoReports := make([]types.FraudReportPB, len(localReports))
	for i, r := range localReports {
		protoReports[i] = *types.FraudReportToProto(&r)
	}

	return &types.QueryFraudReportsResponse{
		Reports: protoReports,
	}, nil
}

// FraudReportsByReporter returns fraud reports submitted by a reporter
func (q *queryServer) FraudReportsByReporter(ctx context.Context, req *types.QueryFraudReportsByReporterRequest) (*types.QueryFraudReportsByReporterResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidReporter.Wrap("request is nil")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	localReports := q.GetFraudReportsByReporter(sdkCtx, req.Reporter)

	// Convert to proto types
	protoReports := make([]types.FraudReportPB, len(localReports))
	for i, r := range localReports {
		protoReports[i] = *types.FraudReportToProto(&r)
	}

	return &types.QueryFraudReportsByReporterResponse{
		Reports: protoReports,
	}, nil
}

// FraudReportsByReportedParty returns fraud reports against a reported party
func (q *queryServer) FraudReportsByReportedParty(ctx context.Context, req *types.QueryFraudReportsByReportedPartyRequest) (*types.QueryFraudReportsByReportedPartyResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidReportedParty.Wrap("request is nil")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	localReports := q.GetFraudReportsByReportedParty(sdkCtx, req.ReportedParty)

	// Convert to proto types
	protoReports := make([]types.FraudReportPB, len(localReports))
	for i, r := range localReports {
		protoReports[i] = *types.FraudReportToProto(&r)
	}

	return &types.QueryFraudReportsByReportedPartyResponse{
		Reports: protoReports,
	}, nil
}

// AuditLog returns the audit log for a report
func (q *queryServer) AuditLog(ctx context.Context, req *types.QueryAuditLogRequest) (*types.QueryAuditLogResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidReportID.Wrap("request is nil")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	localLogs := q.GetAuditLogsForReport(sdkCtx, req.ReportId)

	// Convert to proto types
	protoLogs := make([]types.FraudAuditLogPB, len(localLogs))
	for i, l := range localLogs {
		protoLogs[i] = *types.FraudAuditLogToProto(&l)
	}

	return &types.QueryAuditLogResponse{
		AuditLogs: protoLogs,
	}, nil
}

// ModeratorQueue returns the moderator queue entries
func (q *queryServer) ModeratorQueue(ctx context.Context, req *types.QueryModeratorQueueRequest) (*types.QueryModeratorQueueResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	localEntries := q.GetModeratorQueue(sdkCtx)

	// Filter by category and assigned_to if specified
	var filteredEntries []types.ModeratorQueueEntry
	for _, entry := range localEntries {
		if req != nil && req.Category != types.FraudCategoryPBUnspecified {
			if entry.Category != types.FraudCategoryFromProto(req.Category) {
				continue
			}
		}
		if req != nil && req.AssignedTo != "" {
			if entry.AssignedTo != req.AssignedTo {
				continue
			}
		}
		filteredEntries = append(filteredEntries, entry)
	}

	// Convert to proto types
	protoEntries := make([]types.ModeratorQueueEntryPB, len(filteredEntries))
	for i, e := range filteredEntries {
		protoEntries[i] = *types.ModeratorQueueEntryToProto(&e)
	}

	return &types.QueryModeratorQueueResponse{
		QueueEntries: protoEntries,
	}, nil
}
