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

// FraudReport returns a fraud report by ID
func (q QueryServer) FraudReport(ctx context.Context, req *QueryFraudReportRequest) (*QueryFraudReportResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidReportID.Wrap("request is nil")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	report, found := q.GetFraudReport(sdkCtx, req.ReportID)
	if !found {
		return nil, types.ErrReportNotFound
	}

	return &QueryFraudReportResponse{Report: report}, nil
}

// FraudReportsByReporter returns all fraud reports by a reporter
func (q QueryServer) FraudReportsByReporter(ctx context.Context, req *QueryFraudReportsByReporterRequest) (*QueryFraudReportsByReporterResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidReporter.Wrap("request is nil")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	reports := q.GetFraudReportsByReporter(sdkCtx, req.Reporter)

	return &QueryFraudReportsByReporterResponse{Reports: reports}, nil
}

// FraudReportsByReportedParty returns all fraud reports against a party
func (q QueryServer) FraudReportsByReportedParty(ctx context.Context, req *QueryFraudReportsByReportedPartyRequest) (*QueryFraudReportsByReportedPartyResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidReportedParty.Wrap("request is nil")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	reports := q.GetFraudReportsByReportedParty(sdkCtx, req.ReportedParty)

	return &QueryFraudReportsByReportedPartyResponse{Reports: reports}, nil
}

// FraudReportsByStatus returns all fraud reports with a specific status
func (q QueryServer) FraudReportsByStatus(ctx context.Context, req *QueryFraudReportsByStatusRequest) (*QueryFraudReportsByStatusResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidStatus.Wrap("request is nil")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	reports := q.GetFraudReportsByStatus(sdkCtx, req.Status)

	return &QueryFraudReportsByStatusResponse{Reports: reports}, nil
}

// ModeratorQueue returns the moderator queue
func (q QueryServer) ModeratorQueue(ctx context.Context, req *QueryModeratorQueueRequest) (*QueryModeratorQueueResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	entries := q.GetModeratorQueue(sdkCtx)

	return &QueryModeratorQueueResponse{Entries: entries}, nil
}

// AuditLogs returns audit logs for a report
func (q QueryServer) AuditLogs(ctx context.Context, req *QueryAuditLogsRequest) (*QueryAuditLogsResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidReportID.Wrap("request is nil")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	logs := q.GetAuditLogsForReport(sdkCtx, req.ReportID)

	return &QueryAuditLogsResponse{Logs: logs}, nil
}

// Params returns the module parameters
func (q QueryServer) Params(ctx context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := q.GetParams(sdkCtx)

	return &QueryParamsResponse{Params: params}, nil
}

// Query request/response types
type QueryFraudReportRequest struct {
	ReportID string `json:"report_id"`
}

type QueryFraudReportResponse struct {
	Report types.FraudReport `json:"report"`
}

type QueryFraudReportsByReporterRequest struct {
	Reporter string `json:"reporter"`
}

type QueryFraudReportsByReporterResponse struct {
	Reports []types.FraudReport `json:"reports"`
}

type QueryFraudReportsByReportedPartyRequest struct {
	ReportedParty string `json:"reported_party"`
}

type QueryFraudReportsByReportedPartyResponse struct {
	Reports []types.FraudReport `json:"reports"`
}

type QueryFraudReportsByStatusRequest struct {
	Status types.FraudReportStatus `json:"status"`
}

type QueryFraudReportsByStatusResponse struct {
	Reports []types.FraudReport `json:"reports"`
}

type QueryModeratorQueueRequest struct{}

type QueryModeratorQueueResponse struct {
	Entries []types.ModeratorQueueEntry `json:"entries"`
}

type QueryAuditLogsRequest struct {
	ReportID string `json:"report_id"`
}

type QueryAuditLogsResponse struct {
	Logs []types.FraudAuditLog `json:"logs"`
}

type QueryParamsRequest struct{}

type QueryParamsResponse struct {
	Params types.Params `json:"params"`
}
