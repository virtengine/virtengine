// Package v1 provides additional methods for generated marketplace types.
package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// sdk.Msg interface methods for MsgWaldurCallback

func (msg *MsgWaldurCallback) ValidateBasic() error {
	if msg.Sender == "" {
		return ErrInvalidAddress.Wrap("sender address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrapf("invalid sender address: %v", err)
	}

	if msg.CallbackType == "" {
		return ErrInvalidCallback.Wrap("callback_type is required")
	}

	if msg.ResourceId == "" {
		return ErrInvalidCallback.Wrap("resource_id is required")
	}

	if msg.Status == "" {
		return ErrInvalidCallback.Wrap("status is required")
	}

	return nil
}

func (msg *MsgWaldurCallback) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

