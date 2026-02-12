package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/virtengine/virtengine/sdk/go/node/audit/v1"
)

// QueryServer implements the audit log query server
type QueryServer struct {
	keeper Keeper
}

// NewQueryServer creates a new query server
func NewQueryServer(keeper Keeper) types.QueryAuditLogServer {
	return &QueryServer{keeper: keeper}
}

var _ types.QueryAuditLogServer = &QueryServer{}

// QueryLogEntries queries all audit log entries with optional filters
func (q *QueryServer) QueryLogEntries(goCtx context.Context, req *types.QueryLogEntriesRequest) (*types.QueryLogEntriesResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidRequest.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var filter types.ExportFilter
	if req.Filter != nil {
		filter = *req.Filter
	}

	// Determine limit from pagination
	limit := int64(0)
	if req.Pagination != nil && req.Pagination.Limit > 0 {
		// Safe conversion: pagination limit is always reasonable
		limit = int64(req.Pagination.Limit) //nolint:gosec
	}

	// Query logs
	entries, err := q.keeper.QueryLogs(ctx, filter, limit)
	if err != nil {
		return nil, err
	}

	return &types.QueryLogEntriesResponse{
		Entries:    entries,
		Pagination: nil, // Simplified - no pagination cursor for now
	}, nil
}

// QueryLogEntry queries a specific audit log entry by ID
func (q *QueryServer) QueryLogEntry(goCtx context.Context, req *types.QueryLogEntryRequest) (*types.QueryLogEntryResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidRequest.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	entry, found := q.keeper.GetLogEntry(ctx, req.Id)
	if !found {
		return nil, types.ErrNotFound.Wrapf("audit log entry %s not found", req.Id)
	}

	return &types.QueryLogEntryResponse{
		Entry: entry,
	}, nil
}

// QueryExportJobs queries all export jobs
func (q *QueryServer) QueryExportJobs(goCtx context.Context, req *types.QueryExportJobsRequest) (*types.QueryExportJobsResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidRequest.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	jobs, err := q.keeper.QueryExportJobs(ctx, req.Requester, req.Status)
	if err != nil {
		return nil, err
	}

	return &types.QueryExportJobsResponse{
		Jobs:       jobs,
		Pagination: nil, // Simplified - no pagination cursor for now
	}, nil
}

// QueryExportJob queries a specific export job by ID
func (q *QueryServer) QueryExportJob(goCtx context.Context, req *types.QueryExportJobRequest) (*types.QueryExportJobResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidRequest.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	job, found := q.keeper.GetExportJob(ctx, req.Id)
	if !found {
		return nil, types.ErrNotFound.Wrapf("export job %s not found", req.Id)
	}

	return &types.QueryExportJobResponse{
		Job: job,
	}, nil
}

// QueryParams queries the audit log module parameters
func (q *QueryServer) QueryParams(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidRequest.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	params := q.keeper.GetAuditLogParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}
