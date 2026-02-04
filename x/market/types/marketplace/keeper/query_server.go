package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	query "github.com/cosmos/cosmos-sdk/types/query"

	marketplacev1 "github.com/virtengine/virtengine/sdk/go/node/marketplace/v1"
	marketplacetypes "github.com/virtengine/virtengine/x/market/types/marketplace"
)

type queryServer struct {
	keeper IKeeper
}

var _ marketplacev1.QueryServer = queryServer{}

// NewQueryServerImpl returns an implementation of the marketplace QueryServer interface.
func NewQueryServerImpl(k IKeeper) marketplacev1.QueryServer {
	return queryServer{keeper: k}
}

// OfferingPrice calculates pricing for a specific offering.
func (qs queryServer) OfferingPrice(
	goCtx context.Context,
	req *marketplacev1.QueryOfferingPriceRequest,
) (*marketplacev1.QueryOfferingPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.OfferingId == "" {
		return nil, status.Error(codes.InvalidArgument, "offering_id is required")
	}

	offeringID, err := marketplacetypes.ParseOfferingID(req.OfferingId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	offering, found := qs.keeper.GetOffering(ctx, offeringID)
	if !found {
		return nil, marketplacetypes.ErrOfferingNotFound
	}

	quantity := req.Quantity
	if quantity == 0 {
		quantity = 1
	}

	quote, err := marketplacetypes.CalculateOfferingPrice(offering, req.ResourceUnits, quantity)
	if err != nil {
		return nil, marketplacetypes.ErrPricingInvalid.Wrap(err.Error())
	}

	return &marketplacev1.QueryOfferingPriceResponse{
		Total: quote.Total,
	}, nil
}

// AllocationsByCustomer returns allocations for a customer.
func (qs queryServer) AllocationsByCustomer(
	goCtx context.Context,
	req *marketplacev1.QueryAllocationsByCustomerRequest,
) (*marketplacev1.QueryAllocationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.CustomerAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "customer_address is required")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	allocations := qs.keeper.GetAllocationsByCustomer(ctx, req.CustomerAddress)
	return paginateAllocations(allocations, req.Pagination)
}

// AllocationsByProvider returns allocations for a provider.
func (qs queryServer) AllocationsByProvider(
	goCtx context.Context,
	req *marketplacev1.QueryAllocationsByProviderRequest,
) (*marketplacev1.QueryAllocationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.ProviderAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "provider_address is required")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	allocations := qs.keeper.GetAllocationsByProvider(ctx, req.ProviderAddress)
	return paginateAllocations(allocations, req.Pagination)
}

func paginateAllocations(allocations []marketplacetypes.Allocation, pageReq *query.PageRequest) (*marketplacev1.QueryAllocationsResponse, error) {
	total := uint64(len(allocations))
	start := uint64(0)
	limit := uint64(len(allocations))
	if pageReq != nil {
		start = pageReq.Offset
		if pageReq.Limit > 0 {
			limit = pageReq.Limit
		}
	}
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	result := make([]marketplacev1.Allocation, 0, end-start)
	for _, allocation := range allocations[start:end] {
		result = append(result, allocationToProto(allocation))
	}

	resp := &marketplacev1.QueryAllocationsResponse{
		Allocations: result,
		Pagination: &query.PageResponse{
			Total: total,
		},
	}
	return resp, nil
}
