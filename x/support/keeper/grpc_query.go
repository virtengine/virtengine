package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	types "github.com/virtengine/virtengine/x/support/types" //nolint:staticcheck // Deprecated types retained for compatibility.
)

// GRPCQuerier implements the gRPC query interface for external ticket refs
type GRPCQuerier struct {
	Keeper
}

var _ types.QueryServer = GRPCQuerier{}

// SupportRequest returns a support request by ID
func (q GRPCQuerier) SupportRequest(c context.Context, req *types.QuerySupportRequestRequest) (*types.QuerySupportRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.TicketID == "" {
		return nil, status.Error(codes.InvalidArgument, "ticket_id is required")
	}
	if req.ViewerAddress == "" {
		return nil, status.Error(codes.PermissionDenied, "viewer_address is required")
	}

	ctx := sdk.UnwrapSDKContext(c)
	id, err := types.ParseSupportRequestID(req.TicketID)
	if err != nil {
		return nil, types.ErrInvalidSupportRequest.Wrap(err.Error())
	}

	request, found := q.GetSupportRequest(ctx, id)
	if !found {
		return nil, types.ErrSupportRequestNotFound
	}

	viewerAddr, err := sdk.AccAddressFromBech32(req.ViewerAddress)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(err.Error())
	}
	if err := q.requireSupportPayloadAccess(ctx, &request.Payload, viewerAddr, req.ViewerKeyID); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	return &types.QuerySupportRequestResponse{Request: request}, nil
}

// SupportRequestsBySubmitter returns support requests by submitter
func (q GRPCQuerier) SupportRequestsBySubmitter(c context.Context, req *types.QuerySupportRequestsBySubmitterRequest) (*types.QuerySupportRequestsBySubmitterResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.SubmitterAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "submitter_address is required")
	}

	ctx := sdk.UnwrapSDKContext(c)
	addr, err := sdk.AccAddressFromBech32(req.SubmitterAddress)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(err.Error())
	}

	viewer := req.ViewerAddress
	if viewer == "" {
		viewer = req.SubmitterAddress
	}
	viewerAddr, err := sdk.AccAddressFromBech32(viewer)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(err.Error())
	}

	requests := q.GetSupportRequestsBySubmitter(ctx, addr)
	if req.Status != "" {
		statusFilter := types.SupportStatusFromString(req.Status)
		if !statusFilter.IsValid() {
			return nil, types.ErrInvalidSupportRequest.Wrapf("invalid status: %s", req.Status)
		}
		filtered := make([]types.SupportRequest, 0, len(requests))
		for _, r := range requests {
			if r.Status == statusFilter {
				filtered = append(filtered, r)
			}
		}
		requests = filtered
	}

	for _, request := range requests {
		if err := q.requireSupportPayloadAccess(ctx, &request.Payload, viewerAddr, req.ViewerKeyID); err != nil {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
	}

	return &types.QuerySupportRequestsBySubmitterResponse{Requests: requests}, nil
}

// SupportResponsesByRequest returns responses for a request
func (q GRPCQuerier) SupportResponsesByRequest(c context.Context, req *types.QuerySupportResponsesByRequestRequest) (*types.QuerySupportResponsesByRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.TicketID == "" {
		return nil, status.Error(codes.InvalidArgument, "ticket_id is required")
	}
	if req.ViewerAddress == "" {
		return nil, status.Error(codes.PermissionDenied, "viewer_address is required")
	}

	ctx := sdk.UnwrapSDKContext(c)
	id, err := types.ParseSupportRequestID(req.TicketID)
	if err != nil {
		return nil, types.ErrInvalidSupportRequest.Wrap(err.Error())
	}

	responses := q.GetSupportResponses(ctx, id)
	viewerAddr, err := sdk.AccAddressFromBech32(req.ViewerAddress)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(err.Error())
	}
	for _, response := range responses {
		if err := q.requireSupportPayloadAccess(ctx, &response.Payload, viewerAddr, req.ViewerKeyID); err != nil {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
	}
	return &types.QuerySupportResponsesByRequestResponse{Responses: responses}, nil
}

// ExternalRef returns a single external ticket reference by resource type and ID
func (q GRPCQuerier) ExternalRef(c context.Context, req *types.QueryExternalRefRequest) (*types.QueryExternalRefResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ResourceType == "" {
		return nil, status.Error(codes.InvalidArgument, "resource_type is required")
	}

	if req.ResourceID == "" {
		return nil, status.Error(codes.InvalidArgument, "resource_id is required")
	}

	ctx := sdk.UnwrapSDKContext(c)

	resourceType := types.ResourceType(req.ResourceType)
	if !resourceType.IsValid() {
		return nil, types.ErrInvalidResourceType.Wrapf("invalid resource type: %s", req.ResourceType)
	}

	ref, found := q.GetExternalRef(ctx, resourceType, req.ResourceID)
	if !found {
		return nil, types.ErrRefNotFound.Wrapf("ref for %s/%s not found", req.ResourceType, req.ResourceID)
	}

	return &types.QueryExternalRefResponse{
		Ref: ref,
	}, nil
}

// ExternalRefsByOwner returns all external refs created by a given owner
func (q GRPCQuerier) ExternalRefsByOwner(c context.Context, req *types.QueryExternalRefsByOwnerRequest) (*types.QueryExternalRefsByOwnerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.OwnerAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "owner_address is required")
	}

	ctx := sdk.UnwrapSDKContext(c)

	ownerAddr, err := sdk.AccAddressFromBech32(req.OwnerAddress)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(err.Error())
	}

	refs := q.GetExternalRefsByOwner(ctx, ownerAddr)

	// Filter by resource type if provided
	if req.ResourceType != "" {
		resourceType := types.ResourceType(req.ResourceType)
		if !resourceType.IsValid() {
			return nil, types.ErrInvalidResourceType.Wrapf("invalid resource type: %s", req.ResourceType)
		}

		var filteredRefs []types.ExternalTicketRef
		for _, ref := range refs {
			if ref.ResourceType == resourceType {
				filteredRefs = append(filteredRefs, ref)
			}
		}
		refs = filteredRefs
	}

	return &types.QueryExternalRefsByOwnerResponse{
		Refs: refs,
	}, nil
}

// Params returns the module parameters
func (q GRPCQuerier) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	params := q.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}
