package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/virtengine/virtengine/x/support/types"
)

// GRPCQuerier implements the gRPC query interface for external ticket refs
type GRPCQuerier struct {
	Keeper
}

var _ types.QueryServer = GRPCQuerier{}

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

	ref, found := q.Keeper.GetExternalRef(ctx, resourceType, req.ResourceID)
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

	refs := q.Keeper.GetExternalRefsByOwner(ctx, ownerAddr)

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
	params := q.Keeper.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}
