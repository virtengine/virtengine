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

// sdk.Msg interface methods for MsgCreateOrder

func (msg *MsgCreateOrder) ValidateBasic() error {
	if msg.Customer == "" {
		return ErrInvalidAddress.Wrap("customer address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Customer); err != nil {
		return ErrInvalidAddress.Wrapf("invalid customer address: %v", err)
	}

	if msg.AcquisitionMode != "direct" && msg.AcquisitionMode != "bid" {
		return ErrInvalidOrder.Wrapf("invalid acquisition mode: %s", msg.AcquisitionMode)
	}

	if msg.RequestedQuantity == 0 {
		return ErrInvalidOrder.Wrap("requested quantity must be positive")
	}

	return nil
}

func (msg *MsgCreateOrder) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Customer)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgPlaceBid

func (msg *MsgPlaceBid) ValidateBasic() error {
	if msg.Provider == "" {
		return ErrInvalidAddress.Wrap("provider address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Provider); err != nil {
		return ErrInvalidAddress.Wrapf("invalid provider address: %v", err)
	}

	if msg.OrderId == "" {
		return ErrInvalidBid.Wrap("order_id is required")
	}

	if msg.Price == 0 {
		return ErrInvalidBid.Wrap("bid price must be positive")
	}

	return nil
}

func (msg *MsgPlaceBid) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Provider)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgWithdrawBid

func (msg *MsgWithdrawBid) ValidateBasic() error {
	if msg.Provider == "" {
		return ErrInvalidAddress.Wrap("provider address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Provider); err != nil {
		return ErrInvalidAddress.Wrapf("invalid provider address: %v", err)
	}

	if msg.BidId == "" {
		return ErrInvalidBid.Wrap("bid_id is required")
	}

	return nil
}

func (msg *MsgWithdrawBid) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Provider)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgRegisterWaldurSource

func (msg *MsgRegisterWaldurSource) ValidateBasic() error {
	if msg.Authority == "" {
		return ErrInvalidAddress.Wrap("authority address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrapf("invalid authority address: %v", err)
	}

	if msg.InstanceId == "" {
		return ErrInvalidRequest.Wrap("instance_id is required")
	}

	if msg.PublicKey == "" {
		return ErrInvalidRequest.Wrap("public_key is required")
	}

	return nil
}

func (msg *MsgRegisterWaldurSource) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgIngestWaldurOffering

func (msg *MsgIngestWaldurOffering) ValidateBasic() error {
	if msg.Relayer == "" {
		return ErrInvalidAddress.Wrap("relayer address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Relayer); err != nil {
		return ErrInvalidAddress.Wrapf("invalid relayer address: %v", err)
	}

	if msg.Snapshot == nil {
		return ErrInvalidOffering.Wrap("snapshot is required")
	}

	if msg.Snapshot.Uuid == "" {
		return ErrInvalidOffering.Wrap("snapshot uuid is required")
	}

	if msg.Signature == "" {
		return ErrInvalidOffering.Wrap("signature is required")
	}

	return nil
}

func (msg *MsgIngestWaldurOffering) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Relayer)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgSetOfferingVisibility

func (msg *MsgSetOfferingVisibility) ValidateBasic() error {
	if msg.Provider == "" {
		return ErrInvalidAddress.Wrap("provider address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Provider); err != nil {
		return ErrInvalidAddress.Wrapf("invalid provider address: %v", err)
	}

	if msg.OfferingId == "" {
		return ErrInvalidOffering.Wrap("offering_id is required")
	}

	if msg.Visibility != "public" && msg.Visibility != "unlisted" && msg.Visibility != "private" {
		return ErrInvalidOffering.Wrapf("invalid visibility: %s", msg.Visibility)
	}

	return nil
}

func (msg *MsgSetOfferingVisibility) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Provider)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgAckWaldurCommand

func (msg *MsgAckWaldurCommand) ValidateBasic() error {
	if msg.Sender == "" {
		return ErrInvalidAddress.Wrap("sender address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrapf("invalid sender address: %v", err)
	}

	if msg.CommandId == "" {
		return ErrInvalidRequest.Wrap("command_id is required")
	}

	return nil
}

func (msg *MsgAckWaldurCommand) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}
