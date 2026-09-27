package keeper

import (
	"context"
	"encoding/binary"
	"math"
	"sort"

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
		if len(pageReq.Key) == 8 {
			start = binary.BigEndian.Uint64(pageReq.Key)
		}
		if pageReq.Limit > 0 {
			limit = pageReq.Limit
		}
	}
	if start > total {
		start = total
	}
	end := total
	if limit <= math.MaxUint64-start {
		end = start + limit
	}
	if end > total {
		end = total
	}

	result := make([]marketplacev1.Allocation, 0, end-start)
	for _, allocation := range allocations[start:end] {
		result = append(result, allocationToProto(allocation))
	}

	var nextKey []byte
	if end < total {
		nextKey = make([]byte, 8)
		binary.BigEndian.PutUint64(nextKey, end)
	}
	resp := &marketplacev1.QueryAllocationsResponse{
		Allocations: result,
		Pagination: &query.PageResponse{
			Total:   total,
			NextKey: nextKey,
		},
	}
	return resp, nil
}

// Catalog returns active, browsable offerings across supply sources.
func (qs queryServer) Catalog(
	goCtx context.Context,
	req *marketplacev1.QueryCatalogRequest,
) (*marketplacev1.QueryCatalogResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	filter := CatalogFilter{
		Category:        marketplacetypes.OfferingCategory(req.Category),
		Regions:         append([]string(nil), req.Regions...),
		Backends:        append([]string(nil), req.Backends...),
		IncludeUnlisted: req.IncludeUnlisted,
		Source:          marketplacetypes.OfferingSource(req.Source),
	}
	offerings := qs.keeper.UnifiedCatalog(ctx, filter)

	total := uint64(len(offerings))
	start, limit, pageReq := queryPageBounds(req.Pagination, total)
	end := pageEnd(start, limit, total)

	result := make([]marketplacev1.Offering, 0, end-start)
	for _, offering := range offerings[start:end] {
		result = append(result, *offeringToProto(offering))
	}

	return &marketplacev1.QueryCatalogResponse{
		Offerings:  result,
		Pagination: pageResponse(total, end, pageReq),
	}, nil
}

// WaldurCommands lists durable commands for off-chain Waldur adapters.
func (qs queryServer) WaldurCommands(
	goCtx context.Context,
	req *marketplacev1.QueryWaldurCommandsRequest,
) (*marketplacev1.QueryWaldurCommandsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	commands := make([]marketplacetypes.WaldurCommand, 0)
	qs.keeper.WithWaldurCommands(ctx, func(command marketplacetypes.WaldurCommand) bool {
		if req.InstanceId != "" && command.InstanceID != req.InstanceId {
			return false
		}
		if req.PendingOnly && command.Acked {
			return false
		}
		commands = append(commands, command)
		return false
	})
	sort.SliceStable(commands, func(i, j int) bool { return commands[i].ID < commands[j].ID })

	total := uint64(len(commands))
	start, limit, pageReq := queryPageBounds(req.Pagination, total)
	end := pageEnd(start, limit, total)

	result := make([]marketplacev1.WaldurCommandSummary, 0, end-start)
	for _, command := range commands[start:end] {
		result = append(result, marketplacev1.WaldurCommandSummary{
			Id:                 command.ID,
			Kind:               string(command.Kind),
			InstanceId:         command.InstanceID,
			ChainEntityId:      command.ChainEntityID,
			Acked:              command.Acked,
			CreatedAt:          command.CreatedAt,
			WaldurOfferingUuid: command.WaldurOfferingUUID,
			BackendId:          command.BackendID,
		})
	}

	return &marketplacev1.QueryWaldurCommandsResponse{
		Commands:   result,
		Pagination: pageResponse(total, end, pageReq),
	}, nil
}

func queryPageBounds(pageReq *query.PageRequest, total uint64) (uint64, uint64, *query.PageRequest) {
	start := uint64(0)
	limit := total
	if pageReq != nil {
		start = pageReq.Offset
		if len(pageReq.Key) == 8 {
			start = binary.BigEndian.Uint64(pageReq.Key)
		}
		if pageReq.Limit > 0 {
			limit = pageReq.Limit
		}
	}
	if start > total {
		start = total
	}
	return start, limit, pageReq
}

func pageEnd(start, limit, total uint64) uint64 {
	end := total
	if limit <= math.MaxUint64-start {
		end = start + limit
	}
	if end > total {
		end = total
	}
	return end
}

func pageResponse(total, end uint64, pageReq *query.PageRequest) *query.PageResponse {
	var nextKey []byte
	if end < total {
		nextKey = make([]byte, 8)
		binary.BigEndian.PutUint64(nextKey, end)
	}
	_ = pageReq
	return &query.PageResponse{
		Total:   total,
		NextKey: nextKey,
	}
}
