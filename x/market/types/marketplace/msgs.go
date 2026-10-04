package marketplace

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	marketplacev1 "github.com/virtengine/virtengine/sdk/go/node/marketplace/v1"
)

// Type aliases to generated protobuf types
type (
	MsgCreateOffering                = marketplacev1.MsgCreateOffering
	MsgCreateOfferingResponse        = marketplacev1.MsgCreateOfferingResponse
	MsgUpdateOffering                = marketplacev1.MsgUpdateOffering
	MsgUpdateOfferingResponse        = marketplacev1.MsgUpdateOfferingResponse
	MsgDeactivateOffering            = marketplacev1.MsgDeactivateOffering
	MsgDeactivateOfferingResponse    = marketplacev1.MsgDeactivateOfferingResponse
	MsgAcceptBid                     = marketplacev1.MsgAcceptBid
	MsgAcceptBidResponse             = marketplacev1.MsgAcceptBidResponse
	MsgTerminateAllocation           = marketplacev1.MsgTerminateAllocation
	MsgTerminateAllocationResponse   = marketplacev1.MsgTerminateAllocationResponse
	MsgResizeAllocation              = marketplacev1.MsgResizeAllocation
	MsgResizeAllocationResponse      = marketplacev1.MsgResizeAllocationResponse
	MsgPauseAllocation               = marketplacev1.MsgPauseAllocation
	MsgPauseAllocationResponse       = marketplacev1.MsgPauseAllocationResponse
	MsgWaldurCallback                = marketplacev1.MsgWaldurCallback
	MsgWaldurCallbackResponse        = marketplacev1.MsgWaldurCallbackResponse
	MsgCreateOrder                   = marketplacev1.MsgCreateOrder
	MsgCreateOrderResponse           = marketplacev1.MsgCreateOrderResponse
	MsgPlaceBid                      = marketplacev1.MsgPlaceBid
	MsgPlaceBidResponse              = marketplacev1.MsgPlaceBidResponse
	MsgWithdrawBid                   = marketplacev1.MsgWithdrawBid
	MsgWithdrawBidResponse           = marketplacev1.MsgWithdrawBidResponse
	MsgRegisterWaldurSource          = marketplacev1.MsgRegisterWaldurSource
	MsgRegisterWaldurSourceResponse  = marketplacev1.MsgRegisterWaldurSourceResponse
	MsgIngestWaldurOffering          = marketplacev1.MsgIngestWaldurOffering
	MsgIngestWaldurOfferingResponse  = marketplacev1.MsgIngestWaldurOfferingResponse
	MsgSetOfferingVisibility         = marketplacev1.MsgSetOfferingVisibility
	MsgSetOfferingVisibilityResponse = marketplacev1.MsgSetOfferingVisibilityResponse
	MsgAckWaldurCommand              = marketplacev1.MsgAckWaldurCommand
	MsgAckWaldurCommandResponse      = marketplacev1.MsgAckWaldurCommandResponse
	MsgServer                        = marketplacev1.MsgServer
	UnimplementedMsgServer           = marketplacev1.UnimplementedMsgServer
)

// Message type constants
const (
	TypeMsgCreateOffering        = "create_offering"
	TypeMsgUpdateOffering        = "update_offering"
	TypeMsgDeactivateOffering    = "deactivate_offering"
	TypeMsgAcceptBid             = "accept_bid"
	TypeMsgTerminateAllocation   = "terminate_allocation"
	TypeMsgResizeAllocation      = "resize_allocation"
	TypeMsgPauseAllocation       = "pause_allocation"
	TypeMsgWaldurCallback        = "waldur_callback"
	TypeMsgCreateOrder           = "create_order"
	TypeMsgPlaceBid              = "place_bid"
	TypeMsgWithdrawBid           = "withdraw_bid"
	TypeMsgRegisterWaldurSource  = "register_waldur_source"
	TypeMsgIngestWaldurOffering  = "ingest_waldur_offering"
	TypeMsgSetOfferingVisibility = "set_offering_visibility"
	TypeMsgAckWaldurCommand      = "ack_waldur_command"
)

var (
	_ sdk.Msg = &MsgCreateOffering{}
	_ sdk.Msg = &MsgUpdateOffering{}
	_ sdk.Msg = &MsgDeactivateOffering{}
	_ sdk.Msg = &MsgAcceptBid{}
	_ sdk.Msg = &MsgTerminateAllocation{}
	_ sdk.Msg = &MsgResizeAllocation{}
	_ sdk.Msg = &MsgPauseAllocation{}
	_ sdk.Msg = &MsgWaldurCallback{}
	_ sdk.Msg = &MsgCreateOrder{}
	_ sdk.Msg = &MsgPlaceBid{}
	_ sdk.Msg = &MsgWithdrawBid{}
	_ sdk.Msg = &MsgRegisterWaldurSource{}
	_ sdk.Msg = &MsgIngestWaldurOffering{}
	_ sdk.Msg = &MsgSetOfferingVisibility{}
	_ sdk.Msg = &MsgAckWaldurCommand{}

	// RegisterMsgServer registers the MsgServer on a grpc server.
	RegisterMsgServer = marketplacev1.RegisterMsgServer
)

// NewMsgWaldurCallback creates a new MsgWaldurCallback.
func NewMsgWaldurCallback(sender string, callbackType string, resourceID string, status string, payload string, signature []byte) *MsgWaldurCallback {
	return &MsgWaldurCallback{
		Sender:       sender,
		CallbackType: callbackType,
		ResourceId:   resourceID,
		Status:       status,
		Payload:      payload,
		Signature:    signature,
	}
}

// NewMsgCreateOffering creates a new MsgCreateOffering.
func NewMsgCreateOffering(provider string, offering *marketplacev1.Offering) *MsgCreateOffering {
	return &MsgCreateOffering{
		Provider: provider,
		Offering: offering,
	}
}

// NewMsgUpdateOffering creates a new MsgUpdateOffering.
func NewMsgUpdateOffering(provider string, offeringID string, updates *marketplacev1.Offering) *MsgUpdateOffering {
	return &MsgUpdateOffering{
		Provider:   provider,
		OfferingId: offeringID,
		Updates:    updates,
	}
}

// NewMsgDeactivateOffering creates a new MsgDeactivateOffering.
func NewMsgDeactivateOffering(provider string, offeringID string) *MsgDeactivateOffering {
	return &MsgDeactivateOffering{
		Provider:   provider,
		OfferingId: offeringID,
	}
}

// NewMsgAcceptBid creates a new MsgAcceptBid.
func NewMsgAcceptBid(customer string, orderID string, bidID string) *MsgAcceptBid {
	return &MsgAcceptBid{
		Customer: customer,
		OrderId:  orderID,
		BidId:    bidID,
	}
}

// NewMsgTerminateAllocation creates a new MsgTerminateAllocation.
func NewMsgTerminateAllocation(customer string, allocationID string, reason string) *MsgTerminateAllocation {
	return &MsgTerminateAllocation{
		Customer:     customer,
		AllocationId: allocationID,
		Reason:       reason,
	}
}

// NewMsgResizeAllocation creates a new MsgResizeAllocation.
func NewMsgResizeAllocation(customer string, allocationID string, resourceUnits []marketplacev1.ResourceUnit, reason string) *MsgResizeAllocation {
	return &MsgResizeAllocation{
		Customer:      customer,
		AllocationId:  allocationID,
		ResourceUnits: resourceUnits,
		Reason:        reason,
	}
}

// NewMsgPauseAllocation creates a new MsgPauseAllocation.
func NewMsgPauseAllocation(customer string, allocationID string, reason string) *MsgPauseAllocation {
	return &MsgPauseAllocation{
		Customer:     customer,
		AllocationId: allocationID,
		Reason:       reason,
	}
}

// NewMsgCreateOrder creates a new MsgCreateOrder.
func NewMsgCreateOrder(customer string) *MsgCreateOrder {
	return &MsgCreateOrder{
		Customer:          customer,
		AcquisitionMode:   string(AcquisitionModeDirect),
		RequestedQuantity: 1,
	}
}

// NewMsgPlaceBid creates a new MsgPlaceBid.
func NewMsgPlaceBid(provider string, orderID string, price uint64) *MsgPlaceBid {
	return &MsgPlaceBid{
		Provider: provider,
		OrderId:  orderID,
		Price:    price,
	}
}

// NewMsgWithdrawBid creates a new MsgWithdrawBid.
func NewMsgWithdrawBid(provider string, bidID string) *MsgWithdrawBid {
	return &MsgWithdrawBid{
		Provider: provider,
		BidId:    bidID,
	}
}

// NewMsgRegisterWaldurSource creates a new MsgRegisterWaldurSource.
func NewMsgRegisterWaldurSource(authority string, instanceID string, publicKey string) *MsgRegisterWaldurSource {
	return &MsgRegisterWaldurSource{
		Authority:  authority,
		InstanceId: instanceID,
		PublicKey:  publicKey,
	}
}

// NewMsgSetOfferingVisibility creates a new MsgSetOfferingVisibility.
func NewMsgSetOfferingVisibility(provider string, offeringID string, visibility OfferingVisibility) *MsgSetOfferingVisibility {
	return &MsgSetOfferingVisibility{
		Provider:   provider,
		OfferingId: offeringID,
		Visibility: string(visibility),
	}
}

// NewMsgAckWaldurCommand creates a new MsgAckWaldurCommand.
func NewMsgAckWaldurCommand(sender string, commandID string) *MsgAckWaldurCommand {
	return &MsgAckWaldurCommand{
		Sender:    sender,
		CommandId: commandID,
	}
}
