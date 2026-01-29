// Package marketplace contains proto.Message stub implementations for the marketplace module.
//
// These are temporary stub implementations until proper protobuf generation is set up.
// They implement the proto.Message interface required by Cosmos SDK.
package marketplace

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
)

func init() {
	proto.RegisterType((*MsgWaldurCallback)(nil), "virtengine.marketplace.v1.MsgWaldurCallback")
	proto.RegisterType((*MsgWaldurCallbackResponse)(nil), "virtengine.marketplace.v1.MsgWaldurCallbackResponse")
}

// Proto.Message interface stubs for MsgWaldurCallback
func (m *MsgWaldurCallback) ProtoMessage()  {}
func (m *MsgWaldurCallback) Reset()         { *m = MsgWaldurCallback{} }
func (m *MsgWaldurCallback) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgWaldurCallbackResponse
func (m *MsgWaldurCallbackResponse) ProtoMessage()  {}
func (m *MsgWaldurCallbackResponse) Reset()         { *m = MsgWaldurCallbackResponse{} }
func (m *MsgWaldurCallbackResponse) String() string { return fmt.Sprintf("%+v", *m) }
