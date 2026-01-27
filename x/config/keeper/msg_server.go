package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/config/types"
)

// Error message constants
const (
	errMsgInvalidSenderAddr = "invalid sender address"
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
func (ms msgServer) RegisterApprovedClient(ctx sdk.Context, msg *types.MsgRegisterApprovedClient) (*types.MsgRegisterApprovedClientResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	// Check if sender is authorized (admin or governance)
	if !ms.keeper.IsAdmin(ctx, sender) {
		return nil, types.ErrUnauthorized.Wrap("only admin or governance can register clients")
	}

	// Create the approved client
	client := types.NewApprovedClient(
		msg.ClientID,
		msg.Name,
		msg.Description,
		msg.PublicKey,
		msg.KeyType,
		msg.MinVersion,
		msg.MaxVersion,
		msg.AllowedScopes,
		sender.String(),
		ctx.BlockTime(),
	)

	// Register the client
	if err := ms.keeper.RegisterClient(ctx, *client); err != nil {
		return nil, err
	}

	// Add to status index
	ms.keeper.addClientToStatusIndex(ctx, client.ClientID, types.ClientStatusActive)

	// Emit event
	_ = ctx.EventManager().EmitTypedEvent(&types.EventClientRegistered{
		ClientID:     msg.ClientID,
		Name:         msg.Name,
		KeyType:      string(msg.KeyType),
		MinVersion:   msg.MinVersion,
		MaxVersion:   msg.MaxVersion,
		RegisteredBy: sender.String(),
		RegisteredAt: ctx.BlockTime().Unix(),
	})

	return &types.MsgRegisterApprovedClientResponse{
		ClientID:     msg.ClientID,
		RegisteredAt: ctx.BlockTime().Unix(),
	}, nil
}

// UpdateApprovedClient handles updating an approved client
func (ms msgServer) UpdateApprovedClient(ctx sdk.Context, msg *types.MsgUpdateApprovedClient) (*types.MsgUpdateApprovedClientResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	// Check if sender is authorized
	if !ms.keeper.IsAdmin(ctx, sender) {
		return nil, types.ErrUnauthorized.Wrap("only admin or governance can update clients")
	}

	// Update the client
	if err := ms.keeper.UpdateClient(
		ctx,
		msg.ClientID,
		msg.Name,
		msg.Description,
		msg.MinVersion,
		msg.MaxVersion,
		msg.AllowedScopes,
		sender.String(),
	); err != nil {
		return nil, err
	}

	// Emit event
	_ = ctx.EventManager().EmitTypedEvent(&types.EventClientUpdated{
		ClientID:  msg.ClientID,
		UpdatedBy: sender.String(),
		UpdatedAt: ctx.BlockTime().Unix(),
	})

	return &types.MsgUpdateApprovedClientResponse{
		ClientID:  msg.ClientID,
		UpdatedAt: ctx.BlockTime().Unix(),
	}, nil
}

// SuspendApprovedClient handles suspending an approved client
func (ms msgServer) SuspendApprovedClient(ctx sdk.Context, msg *types.MsgSuspendApprovedClient) (*types.MsgSuspendApprovedClientResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	// Check if sender is authorized
	if !ms.keeper.IsAdmin(ctx, sender) {
		return nil, types.ErrUnauthorized.Wrap("only admin or governance can suspend clients")
	}

	// Suspend the client
	if err := ms.keeper.SuspendClient(ctx, msg.ClientID, msg.Reason, sender.String()); err != nil {
		return nil, err
	}

	// Emit event
	_ = ctx.EventManager().EmitTypedEvent(&types.EventClientSuspended{
		ClientID:    msg.ClientID,
		SuspendedBy: sender.String(),
		SuspendedAt: ctx.BlockTime().Unix(),
		Reason:      msg.Reason,
	})

	return &types.MsgSuspendApprovedClientResponse{
		ClientID:    msg.ClientID,
		SuspendedAt: ctx.BlockTime().Unix(),
	}, nil
}

// RevokeApprovedClient handles revoking an approved client
func (ms msgServer) RevokeApprovedClient(ctx sdk.Context, msg *types.MsgRevokeApprovedClient) (*types.MsgRevokeApprovedClientResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	// Check if sender is authorized
	if !ms.keeper.IsAdmin(ctx, sender) {
		return nil, types.ErrUnauthorized.Wrap("only admin or governance can revoke clients")
	}

	// Revoke the client
	if err := ms.keeper.RevokeClient(ctx, msg.ClientID, msg.Reason, sender.String()); err != nil {
		return nil, err
	}

	// Emit event
	_ = ctx.EventManager().EmitTypedEvent(&types.EventClientRevoked{
		ClientID:  msg.ClientID,
		RevokedBy: sender.String(),
		RevokedAt: ctx.BlockTime().Unix(),
		Reason:    msg.Reason,
	})

	return &types.MsgRevokeApprovedClientResponse{
		ClientID:  msg.ClientID,
		RevokedAt: ctx.BlockTime().Unix(),
	}, nil
}

// ReactivateApprovedClient handles reactivating a suspended client
func (ms msgServer) ReactivateApprovedClient(ctx sdk.Context, msg *types.MsgReactivateApprovedClient) (*types.MsgReactivateApprovedClientResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	// Check if sender is authorized
	if !ms.keeper.IsAdmin(ctx, sender) {
		return nil, types.ErrUnauthorized.Wrap("only admin or governance can reactivate clients")
	}

	// Reactivate the client
	if err := ms.keeper.ReactivateClient(ctx, msg.ClientID, msg.Reason, sender.String()); err != nil {
		return nil, err
	}

	// Emit event
	_ = ctx.EventManager().EmitTypedEvent(&types.EventClientReactivated{
		ClientID:      msg.ClientID,
		ReactivatedBy: sender.String(),
		ReactivatedAt: ctx.BlockTime().Unix(),
		Reason:        msg.Reason,
	})

	return &types.MsgReactivateApprovedClientResponse{
		ClientID:      msg.ClientID,
		ReactivatedAt: ctx.BlockTime().Unix(),
	}, nil
}

// UpdateParams handles updating module parameters
func (ms msgServer) UpdateParams(ctx sdk.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if msg.Authority != ms.keeper.GetAuthority() {
		return nil, types.ErrUnauthorized.Wrapf("invalid authority; expected %s, got %s", ms.keeper.GetAuthority(), msg.Authority)
	}

	if err := ms.keeper.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
