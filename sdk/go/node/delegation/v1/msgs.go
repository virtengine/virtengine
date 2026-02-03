// Package v1 provides additional methods for generated delegation types.
package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// sdk.Msg interface methods for MsgUpdateParams

func (msg *MsgUpdateParams) ValidateBasic() error {
	if msg.Authority == "" {
		return ErrInvalidAddress.Wrap("authority address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrapf("invalid authority address: %v", err)
	}

	return nil
}

func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

// sdk.Msg interface methods for MsgDelegate

func (msg *MsgDelegate) ValidateBasic() error {
	if msg.Delegator == "" {
		return ErrInvalidAddress.Wrap("delegator address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Delegator); err != nil {
		return ErrInvalidAddress.Wrapf("invalid delegator address: %v", err)
	}

	if msg.Validator == "" {
		return ErrInvalidAddress.Wrap("validator address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Validator); err != nil {
		return ErrInvalidAddress.Wrapf("invalid validator address: %v", err)
	}

	if !msg.Amount.IsValid() {
		return ErrInvalidAmount.Wrapf("invalid amount: %v", msg.Amount)
	}

	if !msg.Amount.IsPositive() {
		return ErrInvalidAmount.Wrap("amount must be positive")
	}

	return nil
}

func (msg *MsgDelegate) GetSigners() []sdk.AccAddress {
	delegator, _ := sdk.AccAddressFromBech32(msg.Delegator)
	return []sdk.AccAddress{delegator}
}

// sdk.Msg interface methods for MsgUndelegate

func (msg *MsgUndelegate) ValidateBasic() error {
	if msg.Delegator == "" {
		return ErrInvalidAddress.Wrap("delegator address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Delegator); err != nil {
		return ErrInvalidAddress.Wrapf("invalid delegator address: %v", err)
	}

	if msg.Validator == "" {
		return ErrInvalidAddress.Wrap("validator address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Validator); err != nil {
		return ErrInvalidAddress.Wrapf("invalid validator address: %v", err)
	}

	if !msg.Amount.IsValid() {
		return ErrInvalidAmount.Wrapf("invalid amount: %v", msg.Amount)
	}

	if !msg.Amount.IsPositive() {
		return ErrInvalidAmount.Wrap("amount must be positive")
	}

	return nil
}

func (msg *MsgUndelegate) GetSigners() []sdk.AccAddress {
	delegator, _ := sdk.AccAddressFromBech32(msg.Delegator)
	return []sdk.AccAddress{delegator}
}

// sdk.Msg interface methods for MsgRedelegate

func (msg *MsgRedelegate) ValidateBasic() error {
	if msg.Delegator == "" {
		return ErrInvalidAddress.Wrap("delegator address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Delegator); err != nil {
		return ErrInvalidAddress.Wrapf("invalid delegator address: %v", err)
	}

	if msg.SrcValidator == "" {
		return ErrInvalidAddress.Wrap("source validator address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.SrcValidator); err != nil {
		return ErrInvalidAddress.Wrapf("invalid source validator address: %v", err)
	}

	if msg.DstValidator == "" {
		return ErrInvalidAddress.Wrap("destination validator address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.DstValidator); err != nil {
		return ErrInvalidAddress.Wrapf("invalid destination validator address: %v", err)
	}

	if msg.SrcValidator == msg.DstValidator {
		return ErrSelfRedelegation.Wrap("source and destination validators cannot be the same")
	}

	if !msg.Amount.IsValid() {
		return ErrInvalidAmount.Wrapf("invalid amount: %v", msg.Amount)
	}

	if !msg.Amount.IsPositive() {
		return ErrInvalidAmount.Wrap("amount must be positive")
	}

	return nil
}

func (msg *MsgRedelegate) GetSigners() []sdk.AccAddress {
	delegator, _ := sdk.AccAddressFromBech32(msg.Delegator)
	return []sdk.AccAddress{delegator}
}

// sdk.Msg interface methods for MsgClaimRewards

func (msg *MsgClaimRewards) ValidateBasic() error {
	if msg.Delegator == "" {
		return ErrInvalidAddress.Wrap("delegator address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Delegator); err != nil {
		return ErrInvalidAddress.Wrapf("invalid delegator address: %v", err)
	}

	if msg.Validator == "" {
		return ErrInvalidAddress.Wrap("validator address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Validator); err != nil {
		return ErrInvalidAddress.Wrapf("invalid validator address: %v", err)
	}

	return nil
}

func (msg *MsgClaimRewards) GetSigners() []sdk.AccAddress {
	delegator, _ := sdk.AccAddressFromBech32(msg.Delegator)
	return []sdk.AccAddress{delegator}
}

// sdk.Msg interface methods for MsgClaimAllRewards

func (msg *MsgClaimAllRewards) ValidateBasic() error {
	if msg.Delegator == "" {
		return ErrInvalidAddress.Wrap("delegator address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Delegator); err != nil {
		return ErrInvalidAddress.Wrapf("invalid delegator address: %v", err)
	}

	return nil
}

func (msg *MsgClaimAllRewards) GetSigners() []sdk.AccAddress {
	delegator, _ := sdk.AccAddressFromBech32(msg.Delegator)
	return []sdk.AccAddress{delegator}
}
