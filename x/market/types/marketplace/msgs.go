package marketplace

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	marketplacev1 "github.com/virtengine/virtengine/sdk/go/node/marketplace/v1"
)

// Type aliases to generated protobuf types
type (
	MsgWaldurCallback            = marketplacev1.MsgWaldurCallback
	MsgWaldurCallbackResponse    = marketplacev1.MsgWaldurCallbackResponse
	MsgCreateOffering            = marketplacev1.MsgCreateOffering
	MsgCreateOfferingResponse    = marketplacev1.MsgCreateOfferingResponse
	MsgUpdateOffering            = marketplacev1.MsgUpdateOffering
	MsgUpdateOfferingResponse    = marketplacev1.MsgUpdateOfferingResponse
	MsgDeprecateOffering         = marketplacev1.MsgDeprecateOffering
	MsgDeprecateOfferingResponse = marketplacev1.MsgDeprecateOfferingResponse
	MsgServer                    = marketplacev1.MsgServer
	UnimplementedMsgServer       = marketplacev1.UnimplementedMsgServer
)

// Message type constants
const (
	TypeMsgWaldurCallback    = "waldur_callback"
	TypeMsgCreateOffering    = "create_offering"
	TypeMsgUpdateOffering    = "update_offering"
	TypeMsgDeprecateOffering = "deprecate_offering"
)

var (
	_ sdk.Msg = &MsgWaldurCallback{}
	_ sdk.Msg = &MsgCreateOffering{}
	_ sdk.Msg = &MsgUpdateOffering{}
	_ sdk.Msg = &MsgDeprecateOffering{}

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
func NewMsgCreateOffering(sender string, offering []byte) *MsgCreateOffering {
	return &MsgCreateOffering{
		Sender:   sender,
		Offering: offering,
	}
}

// NewMsgUpdateOffering creates a new MsgUpdateOffering.
func NewMsgUpdateOffering(sender string, offeringID string, offering []byte) *MsgUpdateOffering {
	return &MsgUpdateOffering{
		Sender:     sender,
		OfferingId: offeringID,
		Offering:   offering,
	}
}

// NewMsgDeprecateOffering creates a new MsgDeprecateOffering.
func NewMsgDeprecateOffering(sender string, offeringID string, reason string) *MsgDeprecateOffering {
	return &MsgDeprecateOffering{
		Sender:     sender,
		OfferingId: offeringID,
		Reason:     reason,
	}
}
