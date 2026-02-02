package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/config/types"
)

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of the config MsgServer interface
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{keeper: k}
}

// NewMsgServerWithContext returns an implementation that uses context
func NewMsgServerWithContext(k Keeper) types.MsgServer {
	return &msgServer{keeper: k}
}

var _ types.MsgServer = msgServer{}

// RegisterApprovedClient handles registering a new approved client
func (ms msgServer) RegisterApprovedClient(goCtx context.Context, msg *types.MsgRegisterApprovedClient) (*types.MsgRegisterApprovedClientResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid authority address")
	}

	// Check if sender is authorized (admin or governance)
	if !ms.keeper.IsAdmin(ctx, authority) {
		return nil, types.ErrUnauthorized.Wrap("only admin or governance can register clients")
	}

	// Create the approved client - using simplified proto types
	client := &types.ApprovedClient{
		ClientID:      msg.ClientId,
		Name:          msg.Name,
		Description:   msg.Description,
		PublicKey:     []byte(msg.PublicKey),
		KeyType:       types.KeyTypeEd25519, // Default key type for proto messages
		MinVersion:    msg.VersionConstraint,
		AllowedScopes: msg.AllowedScopes,
		Status:        types.ClientStatusActive,
		RegisteredBy:  authority.String(),
		RegisteredAt:  ctx.BlockTime(),
		LastUpdatedAt: ctx.BlockTime(),
	}

	// Register the client
	if err := ms.keeper.RegisterClient(ctx, *client); err != nil {
		return nil, err
	}

	// Add to status index
	ms.keeper.addClientToStatusIndex(ctx, client.ClientID, types.ClientStatusActive)

	// Emit event using proto field names
	_ = ctx.EventManager().EmitTypedEvent(&types.EventClientRegistered{
		ClientId:     msg.ClientId,
		Name:         msg.Name,
		RegisteredAt: ctx.BlockTime().Unix(),
	})

	return &types.MsgRegisterApprovedClientResponse{}, nil
}

// UpdateApprovedClient handles updating an approved client
func (ms msgServer) UpdateApprovedClient(goCtx context.Context, msg *types.MsgUpdateApprovedClient) (*types.MsgUpdateApprovedClientResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid authority address")
	}

	// Check if sender is authorized
	if !ms.keeper.IsAdmin(ctx, authority) {
		return nil, types.ErrUnauthorized.Wrap("only admin or governance can update clients")
	}

	// Update the client
	if err := ms.keeper.UpdateClient(
		ctx,
		msg.ClientId,
		"",
		"",
		msg.VersionConstraint,
		"",
		msg.AllowedScopes,
		authority.String(),
	); err != nil {
		return nil, err
	}

	// Emit event using proto field names
	_ = ctx.EventManager().EmitTypedEvent(&types.EventClientUpdated{
		ClientId:  msg.ClientId,
		UpdatedAt: ctx.BlockTime().Unix(),
	})

	return &types.MsgUpdateApprovedClientResponse{}, nil
}

// SuspendApprovedClient handles suspending an approved client
func (ms msgServer) SuspendApprovedClient(goCtx context.Context, msg *types.MsgSuspendApprovedClient) (*types.MsgSuspendApprovedClientResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid authority address")
	}

	// Check if sender is authorized
	if !ms.keeper.IsAdmin(ctx, authority) {
		return nil, types.ErrUnauthorized.Wrap("only admin or governance can suspend clients")
	}

	// Suspend the client
	if err := ms.keeper.SuspendClient(ctx, msg.ClientId, msg.Reason, authority.String()); err != nil {
		return nil, err
	}

	// Emit event using proto field names
	_ = ctx.EventManager().EmitTypedEvent(&types.EventClientSuspended{
		ClientId:    msg.ClientId,
		Reason:      msg.Reason,
		SuspendedAt: ctx.BlockTime().Unix(),
	})

	return &types.MsgSuspendApprovedClientResponse{}, nil
}

// RevokeApprovedClient handles revoking an approved client
func (ms msgServer) RevokeApprovedClient(goCtx context.Context, msg *types.MsgRevokeApprovedClient) (*types.MsgRevokeApprovedClientResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid authority address")
	}

	// Check if sender is authorized
	if !ms.keeper.IsAdmin(ctx, authority) {
		return nil, types.ErrUnauthorized.Wrap("only admin or governance can revoke clients")
	}

	// Revoke the client
	if err := ms.keeper.RevokeClient(ctx, msg.ClientId, msg.Reason, authority.String()); err != nil {
		return nil, err
	}

	// Emit event using proto field names
	_ = ctx.EventManager().EmitTypedEvent(&types.EventClientRevoked{
		ClientId:  msg.ClientId,
		Reason:    msg.Reason,
		RevokedAt: ctx.BlockTime().Unix(),
	})

	return &types.MsgRevokeApprovedClientResponse{}, nil
}

// ReactivateApprovedClient handles reactivating a suspended client
func (ms msgServer) ReactivateApprovedClient(goCtx context.Context, msg *types.MsgReactivateApprovedClient) (*types.MsgReactivateApprovedClientResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid authority address")
	}

	// Check if sender is authorized
	if !ms.keeper.IsAdmin(ctx, authority) {
		return nil, types.ErrUnauthorized.Wrap("only admin or governance can reactivate clients")
	}

	// Reactivate the client (using empty reason since proto doesn't have it)
	if err := ms.keeper.ReactivateClient(ctx, msg.ClientId, "", authority.String()); err != nil {
		return nil, err
	}

	// Emit event using proto field names
	_ = ctx.EventManager().EmitTypedEvent(&types.EventClientReactivated{
		ClientId:      msg.ClientId,
		ReactivatedAt: ctx.BlockTime().Unix(),
	})

	return &types.MsgReactivateApprovedClientResponse{}, nil
}

// UpdateParams handles updating module parameters
func (ms msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Authority != ms.keeper.GetAuthority() {
		return nil, types.ErrUnauthorized.Wrapf("invalid authority; expected %s, got %s", ms.keeper.GetAuthority(), msg.Authority)
	}

	maxUint32 := uint64(^uint32(0))
	if msg.Params.MaxClients > maxUint32 {
		return nil, types.ErrInvalidProposal.Wrap("max_clients exceeds uint32")
	}

	// Convert proto Params to local Params type for storage
	localParams := types.Params{
		RequireClientSignature:  msg.Params.RequireClientSignature,
		RequireUserSignature:    true, // default
		RequireSaltBinding:      true, // default
		MaxClientsPerRegistrar:  uint32(msg.Params.MaxClients),
		AllowGovernanceOverride: true, // default
		DefaultMinVersion:       "1.0.0",
		AdminAddresses:          []string{},
	}

	if err := ms.keeper.SetParams(ctx, localParams); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
