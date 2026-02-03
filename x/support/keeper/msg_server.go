package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/virtengine/virtengine/x/support/types" //nolint:staticcheck // Deprecated types retained for compatibility.
)

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of the support MsgServer interface
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{keeper: k}
}

var _ types.MsgServer = msgServer{}

// RegisterExternalTicket registers a new external ticket reference
func (ms msgServer) RegisterExternalTicket(goCtx context.Context, msg *types.MsgRegisterExternalTicket) (*types.MsgRegisterExternalTicketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	// Create the reference
	ref := &types.ExternalTicketRef{
		ResourceID:       msg.ResourceID,
		ResourceType:     types.ResourceType(msg.ResourceType),
		ExternalSystem:   types.ExternalSystem(msg.ExternalSystem),
		ExternalTicketID: msg.ExternalTicketID,
		ExternalURL:      msg.ExternalURL,
		CreatedBy:        msg.Sender,
	}

	// Register the reference
	if err := ms.keeper.RegisterExternalRef(ctx, ref); err != nil {
		return nil, err
	}

	// Emit event
	err := ctx.EventManager().EmitTypedEvent(&types.EventExternalTicketRegistered{
		ResourceID:       msg.ResourceID,
		ResourceType:     msg.ResourceType,
		ExternalSystem:   msg.ExternalSystem,
		ExternalTicketID: msg.ExternalTicketID,
		ExternalURL:      msg.ExternalURL,
		CreatedBy:        msg.Sender,
		BlockHeight:      ctx.BlockHeight(),
		Timestamp:        ctx.BlockTime().Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgRegisterExternalTicketResponse{}, nil
}

// UpdateExternalTicket updates an existing external ticket reference
func (ms msgServer) UpdateExternalTicket(goCtx context.Context, msg *types.MsgUpdateExternalTicket) (*types.MsgUpdateExternalTicketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	// Get existing reference
	existing, found := ms.keeper.GetExternalRef(ctx, types.ResourceType(msg.ResourceType), msg.ResourceID)
	if !found {
		return nil, types.ErrRefNotFound.Wrapf("ref for %s/%s not found", msg.ResourceType, msg.ResourceID)
	}

	// Check authorization - only the creator can update
	if existing.CreatedBy != sender.String() {
		return nil, types.ErrUnauthorized.Wrap("only the creator can update this reference")
	}

	// Update the reference
	ref := &types.ExternalTicketRef{
		ResourceID:       msg.ResourceID,
		ResourceType:     types.ResourceType(msg.ResourceType),
		ExternalSystem:   existing.ExternalSystem, // Preserve original system
		ExternalTicketID: msg.ExternalTicketID,
		ExternalURL:      msg.ExternalURL,
		CreatedBy:        existing.CreatedBy,
	}

	if err := ms.keeper.UpdateExternalRef(ctx, ref); err != nil {
		return nil, err
	}

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventExternalTicketUpdated{
		ResourceID:       msg.ResourceID,
		ResourceType:     msg.ResourceType,
		ExternalTicketID: msg.ExternalTicketID,
		ExternalURL:      msg.ExternalURL,
		UpdatedBy:        msg.Sender,
		BlockHeight:      ctx.BlockHeight(),
		Timestamp:        ctx.BlockTime().Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateExternalTicketResponse{}, nil
}

// RemoveExternalTicket removes an external ticket reference
func (ms msgServer) RemoveExternalTicket(goCtx context.Context, msg *types.MsgRemoveExternalTicket) (*types.MsgRemoveExternalTicketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	// Get existing reference
	existing, found := ms.keeper.GetExternalRef(ctx, types.ResourceType(msg.ResourceType), msg.ResourceID)
	if !found {
		return nil, types.ErrRefNotFound.Wrapf("ref for %s/%s not found", msg.ResourceType, msg.ResourceID)
	}

	// Check authorization - only the creator can remove
	if existing.CreatedBy != sender.String() {
		return nil, types.ErrUnauthorized.Wrap("only the creator can remove this reference")
	}

	// Remove the reference
	if err := ms.keeper.RemoveExternalRef(ctx, types.ResourceType(msg.ResourceType), msg.ResourceID); err != nil {
		return nil, err
	}

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventExternalTicketRemoved{
		ResourceID:   msg.ResourceID,
		ResourceType: msg.ResourceType,
		RemovedBy:    msg.Sender,
		BlockHeight:  ctx.BlockHeight(),
		Timestamp:    ctx.BlockTime().Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgRemoveExternalTicketResponse{}, nil
}

// UpdateParams updates the module parameters (governance only)
func (ms msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify authority matches the module's expected authority
	if ms.keeper.GetAuthority() != msg.Authority {
		return nil, types.ErrUnauthorized.Wrapf("invalid authority; expected %s, got %s", ms.keeper.GetAuthority(), msg.Authority)
	}

	// Validate params
	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	// Set the new params
	if err := ms.keeper.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
