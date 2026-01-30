package marketplace

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	marketplacev1 "github.com/virtengine/virtengine/sdk/go/node/marketplace/v1"
)

// Type aliases to generated protobuf types
type (
	MsgWaldurCallback         = marketplacev1.MsgWaldurCallback
	MsgWaldurCallbackResponse = marketplacev1.MsgWaldurCallbackResponse
	MsgServer                 = marketplacev1.MsgServer
	UnimplementedMsgServer    = marketplacev1.UnimplementedMsgServer
)

// Message type constants
const (
	TypeMsgWaldurCallback = "waldur_callback"
)

var (
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
