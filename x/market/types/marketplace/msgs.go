package marketplace

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	marketplacev1 "github.com/virtengine/virtengine/sdk/go/node/marketplace/v1"
)

// Type aliases to generated protobuf types
type (
	MsgCreateOffering              = marketplacev1.MsgCreateOffering
	MsgCreateOfferingResponse      = marketplacev1.MsgCreateOfferingResponse
	MsgUpdateOffering              = marketplacev1.MsgUpdateOffering
	MsgUpdateOfferingResponse      = marketplacev1.MsgUpdateOfferingResponse
	MsgDeactivateOffering          = marketplacev1.MsgDeactivateOffering
	MsgDeactivateOfferingResponse  = marketplacev1.MsgDeactivateOfferingResponse
	MsgAcceptBid                   = marketplacev1.MsgAcceptBid
	MsgAcceptBidResponse           = marketplacev1.MsgAcceptBidResponse
	MsgTerminateAllocation         = marketplacev1.MsgTerminateAllocation
	MsgTerminateAllocationResponse = marketplacev1.MsgTerminateAllocationResponse
	MsgWaldurCallback              = marketplacev1.MsgWaldurCallback
	MsgWaldurCallbackResponse      = marketplacev1.MsgWaldurCallbackResponse
	MsgServer                      = marketplacev1.MsgServer
	UnimplementedMsgServer         = marketplacev1.UnimplementedMsgServer
)

// Message type constants
const (
	TypeMsgCreateOffering      = "create_offering"
	TypeMsgUpdateOffering      = "update_offering"
	TypeMsgDeactivateOffering  = "deactivate_offering"
	TypeMsgAcceptBid           = "accept_bid"
	TypeMsgTerminateAllocation = "terminate_allocation"
	TypeMsgWaldurCallback      = "waldur_callback"
)

var (
	_ sdk.Msg = &MsgCreateOffering{}
	_ sdk.Msg = &MsgUpdateOffering{}
	_ sdk.Msg = &MsgDeactivateOffering{}
	_ sdk.Msg = &MsgAcceptBid{}
	_ sdk.Msg = &MsgTerminateAllocation{}
	_ sdk.Msg = &MsgWaldurCallback{}

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
