package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/encryption/types"
)

// Error message constants
const (
	errMsgInvalidSenderAddr = "invalid sender address"
)

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of the encryption MsgServer interface
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{keeper: k}
}

var _ types.MsgServer = msgServer{}

// RegisterRecipientKey registers a new recipient public key
func (ms msgServer) RegisterRecipientKey(ctx sdk.Context, msg *types.MsgRegisterRecipientKey) (*types.MsgRegisterRecipientKeyResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	// Register the key
	fingerprint, err := ms.keeper.RegisterRecipientKey(ctx, sender, msg.PublicKey, msg.AlgorithmID, msg.Label)
	if err != nil {
		return nil, err
	}

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventKeyRegistered{
		Address:      sender.String(),
		Fingerprint:  fingerprint,
		Algorithm:    msg.AlgorithmID,
		Label:        msg.Label,
		RegisteredAt: ctx.BlockTime().Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgRegisterRecipientKeyResponse{
		KeyFingerprint: fingerprint,
	}, nil
}

// RevokeRecipientKey revokes a recipient's public key
func (ms msgServer) RevokeRecipientKey(ctx sdk.Context, msg *types.MsgRevokeRecipientKey) (*types.MsgRevokeRecipientKeyResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	// Revoke the key
	if err := ms.keeper.RevokeRecipientKey(ctx, sender, msg.KeyFingerprint); err != nil {
		return nil, err
	}

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventKeyRevoked{
		Address:     sender.String(),
		Fingerprint: msg.KeyFingerprint,
		RevokedBy:   sender.String(),
		RevokedAt:   ctx.BlockTime().Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgRevokeRecipientKeyResponse{}, nil
}

// UpdateKeyLabel updates a key's label
func (ms msgServer) UpdateKeyLabel(ctx sdk.Context, msg *types.MsgUpdateKeyLabel) (*types.MsgUpdateKeyLabelResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	// Get current label for event
	keys := ms.keeper.GetRecipientKeys(ctx, sender)
	var oldLabel string
	for _, key := range keys {
		if key.KeyFingerprint == msg.KeyFingerprint {
			oldLabel = key.Label
			break
		}
	}

	// Update the label
	if err := ms.keeper.UpdateKeyLabel(ctx, sender, msg.KeyFingerprint, msg.Label); err != nil {
		return nil, err
	}

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventKeyUpdated{
		Address:     sender.String(),
		Fingerprint: msg.KeyFingerprint,
		Field:       "label",
		OldValue:    oldLabel,
		NewValue:    msg.Label,
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateKeyLabelResponse{}, nil
}

// MsgServerWithContext wraps msgServer for gRPC context handling
type MsgServerWithContext struct {
	msgServer
}

// NewMsgServerWithContext returns a wrapped message server with context support
func NewMsgServerWithContext(k Keeper) MsgServerWithContext {
	return MsgServerWithContext{
		msgServer: msgServer{keeper: k},
	}
}
