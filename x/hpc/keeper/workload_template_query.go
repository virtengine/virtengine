package keeper

import (
	"context"
	"encoding/json"
	"math"
	"strings"

	"cosmossdk.io/store/prefix"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"

	"github.com/virtengine/virtengine/x/hpc/types"
)

// WorkloadTemplateQueryServer provides query helpers for workload templates.
type WorkloadTemplateQueryServer struct {
	Keeper
}

// NewQueryServerImpl returns a workload template query server implementation.
func NewQueryServerImpl(k Keeper) *WorkloadTemplateQueryServer {
	return &WorkloadTemplateQueryServer{Keeper: k}
}

// GetWorkloadTemplate returns a specific workload template by ID/version.
func (q *WorkloadTemplateQueryServer) GetWorkloadTemplate(ctx context.Context, req *types.QueryGetWorkloadTemplateRequest) (*types.QueryGetWorkloadTemplateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.TemplateId == "" || req.Version == "" {
		return nil, status.Error(codes.InvalidArgument, "template_id and version required")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	template, found := q.GetWorkloadTemplateByVersion(sdkCtx, req.TemplateId, req.Version)
	if !found {
		return nil, status.Error(codes.NotFound, "template not found")
	}

	return &types.QueryGetWorkloadTemplateResponse{Template: &template}, nil
}

// ListWorkloadTemplates lists templates (latest versions) or all versions for a template ID.
func (q *WorkloadTemplateQueryServer) ListWorkloadTemplates(ctx context.Context, req *types.QueryListWorkloadTemplatesRequest) (*types.QueryListWorkloadTemplatesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	templates := make([]*types.WorkloadTemplate, 0)

	if req.TemplateId != "" {
		// List all versions for a specific template ID.
		store := prefix.NewStore(sdkCtx.KVStore(q.skey), types.WorkloadTemplateVersionPrefix)
		prefixKey := []byte(req.TemplateId + "/")

		pageRes, err := sdkquery.FilteredPaginate(store, req.Pagination, func(key, value []byte, accumulate bool) (bool, error) {
			if !strings.HasPrefix(string(key), string(prefixKey)) {
				return false, nil
			}

			var template types.WorkloadTemplate
			if err := json.Unmarshal(value, &template); err != nil {
				return false, err
			}
			if accumulate {
				templates = append(templates, &template)
			}
			return true, nil
		})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		return &types.QueryListWorkloadTemplatesResponse{
			Templates:  templates,
			Pagination: pageRes,
		}, nil
	}

	// List latest versions (one per template ID).
	store := prefix.NewStore(sdkCtx.KVStore(q.skey), types.WorkloadTemplatePrefix)
	pageRes, err := sdkquery.FilteredPaginate(store, req.Pagination, func(_ []byte, value []byte, accumulate bool) (bool, error) {
		var template types.WorkloadTemplate
		if err := json.Unmarshal(value, &template); err != nil {
			return false, err
		}
		if accumulate {
			templates = append(templates, &template)
		}
		return true, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryListWorkloadTemplatesResponse{
		Templates:  templates,
		Pagination: pageRes,
	}, nil
}

// ListWorkloadTemplatesByType lists templates by workload type.
func (q *WorkloadTemplateQueryServer) ListWorkloadTemplatesByType(ctx context.Context, req *types.QueryListWorkloadTemplatesByTypeRequest) (*types.QueryListWorkloadTemplatesByTypeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(q.skey), types.WorkloadTemplatePrefix)

	templates := make([]*types.WorkloadTemplate, 0)
	pageRes, err := sdkquery.FilteredPaginate(store, req.Pagination, func(_ []byte, value []byte, accumulate bool) (bool, error) {
		var template types.WorkloadTemplate
		if err := json.Unmarshal(value, &template); err != nil {
			return false, err
		}
		match := template.Type == req.Type
		if accumulate && match {
			templates = append(templates, &template)
		}
		return match, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryListWorkloadTemplatesByTypeResponse{
		Templates:  templates,
		Pagination: pageRes,
	}, nil
}

// ListWorkloadTemplatesByPublisher lists templates by publisher address.
func (q *WorkloadTemplateQueryServer) ListWorkloadTemplatesByPublisher(ctx context.Context, req *types.QueryListWorkloadTemplatesByPublisherRequest) (*types.QueryListWorkloadTemplatesByPublisherResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(q.skey), types.WorkloadTemplatePrefix)

	templates := make([]*types.WorkloadTemplate, 0)
	pageRes, err := sdkquery.FilteredPaginate(store, req.Pagination, func(_ []byte, value []byte, accumulate bool) (bool, error) {
		var template types.WorkloadTemplate
		if err := json.Unmarshal(value, &template); err != nil {
			return false, err
		}
		match := template.Publisher == req.Publisher
		if accumulate && match {
			templates = append(templates, &template)
		}
		return match, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryListWorkloadTemplatesByPublisherResponse{
		Templates:  templates,
		Pagination: pageRes,
	}, nil
}

// ListApprovedWorkloadTemplates lists approved templates.
func (q *WorkloadTemplateQueryServer) ListApprovedWorkloadTemplates(ctx context.Context, req *types.QueryListApprovedWorkloadTemplatesRequest) (*types.QueryListApprovedWorkloadTemplatesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(q.skey), types.WorkloadTemplatePrefix)

	templates := make([]*types.WorkloadTemplate, 0)
	pageRes, err := sdkquery.FilteredPaginate(store, req.Pagination, func(_ []byte, value []byte, accumulate bool) (bool, error) {
		var template types.WorkloadTemplate
		if err := json.Unmarshal(value, &template); err != nil {
			return false, err
		}
		match := template.ApprovalStatus.CanBeUsed()
		if accumulate && match {
			templates = append(templates, &template)
		}
		return match, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryListApprovedWorkloadTemplatesResponse{
		Templates:  templates,
		Pagination: pageRes,
	}, nil
}

// WorkloadTemplateUsage returns usage statistics for a template version.
func (q *WorkloadTemplateQueryServer) WorkloadTemplateUsage(ctx context.Context, req *types.QueryWorkloadTemplateUsageRequest) (*types.QueryWorkloadTemplateUsageResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.TemplateId == "" || req.Version == "" {
		return nil, status.Error(codes.InvalidArgument, "template_id and version required")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	var totalUses uint64
	var activeJobs uint64
	var completedJobs uint64
	var failedJobs uint64

	q.WithJobs(sdkCtx, func(job types.HPCJob) bool {
		if !job.WorkloadSpec.IsPreconfigured {
			return false
		}

		templateID, version := splitTemplateIDVersion(job.WorkloadSpec.PreconfiguredWorkloadID)
		if templateID != req.TemplateId || version != req.Version {
			return false
		}

		totalUses++

		if types.IsTerminalJobState(job.State) {
			switch job.State {
			case types.JobStateCompleted:
				completedJobs++
			case types.JobStateFailed:
				failedJobs++
			}
		} else {
			activeJobs++
		}

		return false
	})

	return &types.QueryWorkloadTemplateUsageResponse{
		TemplateId:    req.TemplateId,
		Version:       req.Version,
		TotalUses:     safeInt64FromUint64(totalUses),
		ActiveJobs:    safeInt64FromUint64(activeJobs),
		CompletedJobs: safeInt64FromUint64(completedJobs),
		FailedJobs:    safeInt64FromUint64(failedJobs),
	}, nil
}

// SearchWorkloadTemplates searches templates by query string.
func (q *WorkloadTemplateQueryServer) SearchWorkloadTemplates(ctx context.Context, req *types.QuerySearchWorkloadTemplatesRequest) (*types.QuerySearchWorkloadTemplatesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(q.skey), types.WorkloadTemplatePrefix)

	query := strings.ToLower(strings.TrimSpace(req.Query))
	templates := make([]*types.WorkloadTemplate, 0)

	pageRes, err := sdkquery.FilteredPaginate(store, req.Pagination, func(_ []byte, value []byte, accumulate bool) (bool, error) {
		var template types.WorkloadTemplate
		if err := json.Unmarshal(value, &template); err != nil {
			return false, err
		}

		match := query == "" ||
			strings.Contains(strings.ToLower(template.TemplateID), query) ||
			strings.Contains(strings.ToLower(template.Name), query) ||
			strings.Contains(strings.ToLower(template.Description), query) ||
			templateHasTag(&template, query)

		if accumulate && match {
			templates = append(templates, &template)
		}
		return match, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QuerySearchWorkloadTemplatesResponse{
		Templates:  templates,
		Pagination: pageRes,
	}, nil
}

func templateHasTag(template *types.WorkloadTemplate, query string) bool {
	if query == "" {
		return false
	}
	for _, tag := range template.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

func splitTemplateIDVersion(value string) (string, string) {
	if value == "" {
		return "", ""
	}
	if strings.Contains(value, "@") {
		parts := strings.SplitN(value, "@", 2)
		return parts[0], parts[1]
	}
	return value, ""
}

func safeInt64FromUint64(val uint64) int64 {
	if val > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(val)
}
