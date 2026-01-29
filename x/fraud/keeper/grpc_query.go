// Package keeper implements the Fraud module keeper.
//
// VE-912: Fraud reporting flow - gRPC query handlers
package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/fraud/types"
)

// QueryServer implements the gRPC query service
type QueryServer struct {
	Keeper
}

// NewQueryServer creates a new gRPC query server
func NewQueryServer(k Keeper) QueryServer {
	return QueryServer{Keeper: k}
}

// Ensure QueryServer implements types.QueryServer
var _ types.QueryServer = QueryServer{}

// FraudReport returns a fraud report by ID
func (q QueryServer) FraudReport(ctx context.Context, req *types.QueryFraudReportRequest) (*types.QueryFraudReportResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidReportID.Wrap("request is nil")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	report, found := q.GetFraudReport(sdkCtx, req.ReportID)
	if !found {
		return nil, types.ErrReportNotFound
	}

	return &types.QueryFraudReportResponse{Report: &report}, nil
}

// FraudReports returns all fraud reports with optional status filter
func (q QueryServer) FraudReports(ctx context.Context, req *types.QueryFraudReportsRequest) (*types.QueryFraudReportsResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidStatus.Wrap("request is nil")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	reports := q.GetFraudReportsByStatus(sdkCtx, req.Status)

	result := make([]*types.FraudReport, len(reports))
	for i := range reports {
		result[i] = &reports[i]
	}

	return &types.QueryFraudReportsResponse{Reports: result}, nil
}

// FraudReportsByReporter returns all fraud reports by a reporter
func (q QueryServer) FraudReportsByReporter(ctx context.Context, req *types.QueryFraudReportsByReporterRequest) (*types.QueryFraudReportsByReporterResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidReporter.Wrap("request is nil")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	reports := q.GetFraudReportsByReporter(sdkCtx, req.Reporter)

	result := make([]*types.FraudReport, len(reports))
	for i := range reports {
		result[i] = &reports[i]
	}

	return &types.QueryFraudReportsByReporterResponse{Reports: result}, nil
}

// FraudReportsByReportedParty returns all fraud reports against a party
func (q QueryServer) FraudReportsByReportedParty(ctx context.Context, req *types.QueryFraudReportsByReportedPartyRequest) (*types.QueryFraudReportsByReportedPartyResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidReportedParty.Wrap("request is nil")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	reports := q.GetFraudReportsByReportedParty(sdkCtx, req.ReportedParty)

	result := make([]*types.FraudReport, len(reports))
	for i := range reports {
		result[i] = &reports[i]
	}

	return &types.QueryFraudReportsByReportedPartyResponse{Reports: result}, nil
}

// ModeratorQueue returns the moderator queue
func (q QueryServer) ModeratorQueue(ctx context.Context, req *types.QueryModeratorQueueRequest) (*types.QueryModeratorQueueResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	entries := q.GetModeratorQueue(sdkCtx)

	result := make([]*types.ModeratorQueueEntry, len(entries))
	for i := range entries {
		result[i] = &entries[i]
	}

	return &types.QueryModeratorQueueResponse{Entries: result}, nil
}

// Params returns the module parameters
func (q QueryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := q.GetParams(sdkCtx)

	return &types.QueryParamsResponse{Params: params}, nil
}
