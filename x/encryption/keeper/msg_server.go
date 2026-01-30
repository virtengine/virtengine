package keeper

import (
	"context"

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

var _ types.MsgServer = &msgServer{}

// RegisterRecipientKey registers a new recipient public key
func (ms *msgServer) RegisterRecipientKey(goCtx context.Context, msg *types.MsgRegisterRecipientKey) (*types.MsgRegisterRecipientKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	// Register the key
	fingerprint, err := ms.keeper.RegisterRecipientKey(ctx, sender, msg.PublicKey, msg.AlgorithmId, msg.Label)
	if err != nil {
		return nil, err
	}

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventKeyRegisteredPB{
		Address:      sender.String(),
		Fingerprint:  fingerprint,
		Algorithm:    msg.AlgorithmId,
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
func (ms *msgServer) RevokeRecipientKey(goCtx context.Context, msg *types.MsgRevokeRecipientKey) (*types.MsgRevokeRecipientKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	// Revoke the key
	if err := ms.keeper.RevokeRecipientKey(ctx, sender, msg.KeyFingerprint); err != nil {
		return nil, err
	}

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventKeyRevokedPB{
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
func (ms *msgServer) UpdateKeyLabel(goCtx context.Context, msg *types.MsgUpdateKeyLabel) (*types.MsgUpdateKeyLabelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

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
	err = ctx.EventManager().EmitTypedEvent(&types.EventKeyUpdatedPB{
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
