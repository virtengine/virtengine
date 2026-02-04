// Package v1 provides additional methods for generated marketplace types.
package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// sdk.Msg interface methods for MsgCreateOffering

func (msg *MsgCreateOffering) ValidateBasic() error {
	if msg.Provider == "" {
		return ErrInvalidAddress.Wrap("provider address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Provider); err != nil {
		return ErrInvalidAddress.Wrapf("invalid provider address: %v", err)
	}

	if msg.Offering == nil {
		return ErrInvalidOffering.Wrap("offering is required")
	}

	return nil
}

func (msg *MsgCreateOffering) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Provider)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgUpdateOffering

func (msg *MsgUpdateOffering) ValidateBasic() error {
	if msg.Provider == "" {
		return ErrInvalidAddress.Wrap("provider address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Provider); err != nil {
		return ErrInvalidAddress.Wrapf("invalid provider address: %v", err)
	}

	if msg.OfferingId == "" {
		return ErrInvalidOffering.Wrap("offering_id is required")
	}

	if msg.Updates == nil {
		return ErrInvalidOffering.Wrap("updates are required")
	}

	return nil
}

func (msg *MsgUpdateOffering) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Provider)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgDeactivateOffering

func (msg *MsgDeactivateOffering) ValidateBasic() error {
	if msg.Provider == "" {
		return ErrInvalidAddress.Wrap("provider address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Provider); err != nil {
		return ErrInvalidAddress.Wrapf("invalid provider address: %v", err)
	}

	if msg.OfferingId == "" {
		return ErrInvalidOffering.Wrap("offering_id is required")
	}

	return nil
}

func (msg *MsgDeactivateOffering) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Provider)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgAcceptBid

func (msg *MsgAcceptBid) ValidateBasic() error {
	if msg.Customer == "" {
		return ErrInvalidAddress.Wrap("customer address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Customer); err != nil {
		return ErrInvalidAddress.Wrapf("invalid customer address: %v", err)
	}

	if msg.OrderId == "" {
		return ErrInvalidOrder.Wrap("order_id is required")
	}

	if msg.BidId == "" {
		return ErrInvalidBid.Wrap("bid_id is required")
	}

	return nil
}

func (msg *MsgAcceptBid) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Customer)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgTerminateAllocation

func (msg *MsgTerminateAllocation) ValidateBasic() error {
	if msg.Customer == "" {
		return ErrInvalidAddress.Wrap("customer address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Customer); err != nil {
		return ErrInvalidAddress.Wrapf("invalid customer address: %v", err)
	}

	if msg.AllocationId == "" {
		return ErrInvalidAllocation.Wrap("allocation_id is required")
	}

	return nil
}

func (msg *MsgTerminateAllocation) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Customer)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgResizeAllocation

func (msg *MsgResizeAllocation) ValidateBasic() error {
	if msg.Customer == "" {
		return ErrInvalidAddress.Wrap("customer address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Customer); err != nil {
		return ErrInvalidAddress.Wrapf("invalid customer address: %v", err)
	}

	if msg.AllocationId == "" {
		return ErrInvalidAllocation.Wrap("allocation_id is required")
	}

	if len(msg.ResourceUnits) == 0 {
		return ErrInvalidAllocation.Wrap("resource_units are required")
	}

	return nil
}

func (msg *MsgResizeAllocation) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Customer)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgPauseAllocation

func (msg *MsgPauseAllocation) ValidateBasic() error {
	if msg.Customer == "" {
		return ErrInvalidAddress.Wrap("customer address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Customer); err != nil {
		return ErrInvalidAddress.Wrapf("invalid customer address: %v", err)
	}

	if msg.AllocationId == "" {
		return ErrInvalidAllocation.Wrap("allocation_id is required")
	}

	return nil
}

func (msg *MsgPauseAllocation) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Customer)
	return []sdk.AccAddress{addr}
}

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
