package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
	"github.com/virtengine/virtengine/x/resources/types"
)

// msgServer implements the MsgServer interface.
type msgServer struct {
	Keeper
}

var _ resourcesv1.MsgServer = (*msgServer)(nil)

// NewMsgServerImpl returns an implementation of the MsgServer interface.
func NewMsgServerImpl(keeper Keeper) resourcesv1.MsgServer {
	return &msgServer{Keeper: keeper}
}

// ProviderHeartbeat handles inventory heartbeats.
func (m msgServer) ProviderHeartbeat(ctx context.Context, msg *types.MsgProviderHeartbeat) (*types.MsgProviderHeartbeatResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if msg == nil {
		return nil, types.ErrInvalidRequest
	}
	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return nil, types.ErrInvalidRequest.Wrap("invalid provider address")
	}
	if msg.ResourceClass == types.ResourceClassUnspecified {
		return nil, types.ErrInvalidRequest.Wrap("resource class required")
	}
	if msg.Sequence == 0 {
		return nil, types.ErrInvalidRequest.Wrap("sequence must be positive")
	}

	inv, err := m.UpdateInventoryFromHeartbeat(sdkCtx, msg)
	if err != nil {
		return nil, err
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"resource_heartbeat",
			sdk.NewAttribute("provider", inv.ProviderAddress),
			sdk.NewAttribute("inventory_id", inv.InventoryId),
		),
	)

	return &types.MsgProviderHeartbeatResponse{Accepted: true, SequenceAck: msg.Sequence}, nil
}

// AllocateResources handles allocation requests.
func (m msgServer) AllocateResources(ctx context.Context, msg *types.MsgAllocateResources) (*types.MsgAllocateResourcesResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if msg == nil {
		return nil, types.ErrInvalidRequest
	}
	if _, err := sdk.AccAddressFromBech32(msg.RequesterAddress); err != nil {
		return nil, types.ErrInvalidRequest.Wrap("invalid requester address")
	}
	if err := validateRequest(msg.Request); err != nil {
		return nil, err
	}

	allocation, err := m.Keeper.AllocateResources(sdkCtx, msg.Request)
	if err != nil {
		return nil, err
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"resource_allocation",
			sdk.NewAttribute("allocation_id", allocation.AllocationId),
			sdk.NewAttribute("provider", allocation.ProviderAddress),
		),
	)

	return &types.MsgAllocateResourcesResponse{AllocationId: allocation.AllocationId}, nil
}

// ActivateAllocation acknowledges an allocation.
func (m msgServer) ActivateAllocation(ctx context.Context, msg *types.MsgActivateAllocation) (*types.MsgActivateAllocationResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if msg == nil {
		return nil, types.ErrInvalidRequest
	}
	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return nil, types.ErrInvalidRequest.Wrap("invalid provider address")
	}
	if msg.AllocationId == "" {
		return nil, types.ErrInvalidRequest.Wrap("allocation_id required")
	}

	allocation, err := m.Keeper.ActivateAllocation(sdkCtx, msg.AllocationId, msg.ProviderAddress)
	if err != nil {
		return nil, err
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"resource_allocation_active",
			sdk.NewAttribute("allocation_id", allocation.AllocationId),
			sdk.NewAttribute("provider", allocation.ProviderAddress),
		),
	)

	return &types.MsgActivateAllocationResponse{}, nil
}

// ReleaseAllocation releases an allocation.
func (m msgServer) ReleaseAllocation(ctx context.Context, msg *types.MsgReleaseAllocation) (*types.MsgReleaseAllocationResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if msg == nil {
		return nil, types.ErrInvalidRequest
	}
	if _, err := sdk.AccAddressFromBech32(msg.RequesterAddress); err != nil {
		return nil, types.ErrInvalidRequest.Wrap("invalid requester address")
	}
	if msg.AllocationId == "" {
		return nil, types.ErrInvalidRequest.Wrap("allocation_id required")
	}

	allocation, err := m.Keeper.ReleaseAllocation(sdkCtx, msg.AllocationId, msg.RequesterAddress, msg.Reason)
	if err != nil {
		return nil, err
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"resource_allocation_released",
			sdk.NewAttribute("allocation_id", allocation.AllocationId),
			sdk.NewAttribute("provider", allocation.ProviderAddress),
		),
	)

	return &types.MsgReleaseAllocationResponse{}, nil
}

// UpdateParams updates module params.
func (m msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if msg == nil {
		return nil, types.ErrInvalidRequest
	}
	if msg.Authority != m.GetAuthority() {
		return nil, types.ErrUnauthorized
	}
	if err := m.SetParams(sdkCtx, types.ParamsFromProto(msg.Params)); err != nil {
		return nil, err
	}
	return &types.MsgUpdateParamsResponse{}, nil
}

func validateRequest(req types.ResourceRequest) error {
	if req.ResourceClass == types.ResourceClassUnspecified {
		return types.ErrInvalidRequest.Wrap("resource_class required")
	}
	if req.Required.CpuCores == 0 && req.Required.MemoryGb == 0 && req.Required.StorageGb == 0 && req.Required.NetworkMbps == 0 && req.Required.Gpus == 0 {
		return types.ErrInvalidRequest.Wrap("required capacity missing")
	}
	if req.RequestId == "" {
		return types.ErrInvalidRequest.Wrap("request_id required")
	}
	return nil
}
