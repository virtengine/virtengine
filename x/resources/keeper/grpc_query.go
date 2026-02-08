package keeper

import (
	"context"
	"encoding/json"
	"strconv"

	"cosmossdk.io/store/prefix"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"

	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
	"github.com/virtengine/virtengine/x/resources/types"
)

// Querier implements the gRPC QueryServer for the resources module.
type Querier struct {
	Keeper
	resourcesv1.UnimplementedQueryServer
}

// NewQuerier returns a new Querier.
func NewQuerier(k Keeper) *Querier {
	return &Querier{Keeper: k}
}

var _ resourcesv1.QueryServer = (*Querier)(nil)

// AvailableResources returns eligible inventories for a request.
func (q *Querier) AvailableResources(ctx context.Context, req *resourcesv1.QueryAvailableResourcesRequest) (*resourcesv1.QueryAvailableResourcesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	candidates := q.selectInventoryCandidates(sdkCtx, req.Request, int(req.Request.MaxCandidates))
	out := make([]resourcesv1.AvailableResource, 0, len(candidates))

	for _, candidate := range candidates {
		out = append(out, resourcesv1.AvailableResource{
			Inventory: candidate.inventory,
			Score:     strconv.FormatInt(candidate.combinedScore, 10),
		})
	}

	return &resourcesv1.QueryAvailableResourcesResponse{Candidates: out}, nil
}

// Allocation returns an allocation by ID.
func (q *Querier) Allocation(ctx context.Context, req *resourcesv1.QueryAllocationRequest) (*resourcesv1.QueryAllocationResponse, error) {
	if req == nil || req.AllocationId == "" {
		return nil, status.Error(codes.InvalidArgument, "allocation_id required")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	allocation, found := q.GetAllocation(sdkCtx, req.AllocationId)
	if !found {
		return nil, status.Error(codes.NotFound, "allocation not found")
	}

	return &resourcesv1.QueryAllocationResponse{Allocation: allocation}, nil
}

// AllocationHistory returns lifecycle events for an allocation.
func (q *Querier) AllocationHistory(ctx context.Context, req *resourcesv1.QueryAllocationHistoryRequest) (*resourcesv1.QueryAllocationHistoryResponse, error) {
	if req == nil || req.AllocationId == "" {
		return nil, status.Error(codes.InvalidArgument, "allocation_id required")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(q.skey), types.AllocationEventPrefix(req.AllocationId))

	var events []resourcesv1.AllocationEvent
	pageRes, err := sdkquery.Paginate(store, req.Pagination, func(_ []byte, value []byte) error {
		var event resourcesv1.AllocationEvent
		if err := json.Unmarshal(value, &event); err != nil {
			return err
		}
		events = append(events, event)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &resourcesv1.QueryAllocationHistoryResponse{Events: events, Pagination: pageRes}, nil
}

// AllocationsByProvider returns allocations for a provider.
func (q *Querier) AllocationsByProvider(ctx context.Context, req *resourcesv1.QueryAllocationsByProviderRequest) (*resourcesv1.QueryAllocationsByProviderResponse, error) {
	if req == nil || req.ProviderAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "provider_address required")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(q.skey), types.AllocationProviderPrefix(req.ProviderAddress))

	var allocations []resourcesv1.ResourceAllocation
	pageRes, err := sdkquery.Paginate(store, req.Pagination, func(key []byte, _ []byte) error {
		allocationID := string(key)
		allocation, found := q.GetAllocation(sdkCtx, allocationID)
		if found {
			allocations = append(allocations, allocation)
		}
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &resourcesv1.QueryAllocationsByProviderResponse{Allocations: allocations, Pagination: pageRes}, nil
}

// Params returns module params.
func (q *Querier) Params(ctx context.Context, _ *resourcesv1.QueryParamsRequest) (*resourcesv1.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := q.GetParams(sdkCtx)
	return &resourcesv1.QueryParamsResponse{Params: params.ToProto()}, nil
}
