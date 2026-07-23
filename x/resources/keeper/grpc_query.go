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
	candidates := q.selectInventoryCandidates(sdkCtx, req.Request, uint64ToInt(req.Request.MaxCandidates))
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

func (q *Querier) Reservation(ctx context.Context, req *resourcesv1.QueryReservationRequest) (*resourcesv1.QueryReservationResponse, error) {
	if req == nil || req.ReservationId == "" {
		return nil, status.Error(codes.InvalidArgument, "reservation_id required")
	}
	reservation, found := q.GetReservation(sdk.UnwrapSDKContext(ctx), req.ReservationId)
	if !found {
		return nil, status.Error(codes.NotFound, "reservation not found")
	}
	return &resourcesv1.QueryReservationResponse{Reservation: reservation}, nil
}

func (q *Querier) ReservationByOrder(ctx context.Context, req *resourcesv1.QueryReservationByOrderRequest) (*resourcesv1.QueryReservationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "order ID required")
	}
	return q.reservationByLineage(ctx, "order", req.GetOrderId())
}

func (q *Querier) ReservationByBid(ctx context.Context, req *resourcesv1.QueryReservationByBidRequest) (*resourcesv1.QueryReservationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "bid ID required")
	}
	return q.reservationByLineage(ctx, "bid", req.GetBidId())
}

func (q *Querier) ReservationByLease(ctx context.Context, req *resourcesv1.QueryReservationByLeaseRequest) (*resourcesv1.QueryReservationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "lease ID required")
	}
	return q.reservationByLineage(ctx, "lease", req.GetLeaseId())
}

func (q *Querier) ReservationByJob(ctx context.Context, req *resourcesv1.QueryReservationByJobRequest) (*resourcesv1.QueryReservationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "job ID required")
	}
	return q.reservationByLineage(ctx, "job", req.GetJobId())
}

func (q *Querier) ReservationByConsumer(ctx context.Context, req *resourcesv1.QueryReservationByConsumerRequest) (*resourcesv1.QueryReservationResponse, error) {
	if req == nil || req.ConsumerType == "" || req.ConsumerId == "" {
		return nil, status.Error(codes.InvalidArgument, "consumer type and ID required")
	}
	reservation, found := q.GetReservationByConsumer(sdk.UnwrapSDKContext(ctx), req.ConsumerType, req.ConsumerId)
	if !found {
		return nil, status.Error(codes.NotFound, "reservation not found")
	}
	return &resourcesv1.QueryReservationResponse{Reservation: reservation}, nil
}

func (q *Querier) ReservationsByProvider(ctx context.Context, req *resourcesv1.QueryReservationsByProviderRequest) (*resourcesv1.QueryReservationsResponse, error) {
	if req == nil || req.ProviderAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "provider_address required")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(q.skey), types.ReservationProviderPrefix(req.ProviderAddress))
	reservations := make([]resourcesv1.Reservation, 0)
	pageRes, err := sdkquery.Paginate(store, req.Pagination, func(key, _ []byte) error {
		reservation, found := q.GetReservation(sdkCtx, string(key))
		if found {
			reservations = append(reservations, reservation)
		}
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &resourcesv1.QueryReservationsResponse{Reservations: reservations, Pagination: pageRes}, nil
}

func (q *Querier) ReservationLineage(ctx context.Context, req *resourcesv1.QueryReservationLineageRequest) (*resourcesv1.QueryReservationLineageResponse, error) {
	if req == nil || req.ReservationId == "" {
		return nil, status.Error(codes.InvalidArgument, "reservation_id required")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	reservation, found := q.GetReservation(sdkCtx, req.ReservationId)
	if !found {
		return nil, status.Error(codes.NotFound, "reservation not found")
	}
	store := prefix.NewStore(sdkCtx.KVStore(q.skey), types.ReservationEventPrefix(req.ReservationId))
	events := make([]resourcesv1.ReservationEvent, 0)
	pageRes, err := sdkquery.Paginate(store, req.Pagination, func(_ []byte, value []byte) error {
		var event resourcesv1.ReservationEvent
		if err := json.Unmarshal(value, &event); err != nil {
			return err
		}
		events = append(events, event)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &resourcesv1.QueryReservationLineageResponse{Reservation: reservation, Events: events, Pagination: pageRes}, nil
}

func (q *Querier) reservationByLineage(ctx context.Context, kind, id string) (*resourcesv1.QueryReservationResponse, error) {
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, kind+" ID required")
	}
	reservation, found := q.GetReservationByLineage(sdk.UnwrapSDKContext(ctx), kind, id)
	if !found {
		return nil, status.Error(codes.NotFound, "reservation not found")
	}
	return &resourcesv1.QueryReservationResponse{Reservation: reservation}, nil
}

// Params returns module params.
func (q *Querier) Params(ctx context.Context, _ *resourcesv1.QueryParamsRequest) (*resourcesv1.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := q.GetParams(sdkCtx)
	return &resourcesv1.QueryParamsResponse{Params: params.ToProto()}, nil
}
