package marketplace

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/gogoproto/grpc"
)

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterLegacyAminoCodec(amino)
}

// RegisterLegacyAminoCodec registers amino types.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgWaldurCallback{}, "marketplace/MsgWaldurCallback")
}

// RegisterInterfaces registers module interfaces.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgWaldurCallback{})
	_ = msgservice.RegisterMsgServiceDesc
}

// _Msg_serviceDesc is the grpc.ServiceDesc for Msg service.
var _Msg_serviceDesc = struct {
	ServiceName string
	HandlerType interface{}
	Methods     []struct {
		MethodName string
		Handler    interface{}
	}
	Streams  []struct{}
	Metadata interface{}
}{
	ServiceName: "virtengine.marketplace.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []struct {
		MethodName string
		Handler    interface{}
	}{
		{MethodName: "SubmitWaldurCallback", Handler: nil},
	},
	Streams:  []struct{}{},
	Metadata: "virtengine/marketplace/v1/tx.proto",
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	SubmitWaldurCallback(ctx sdk.Context, msg *MsgWaldurCallback) (*MsgWaldurCallbackResponse, error)
}

// RegisterMsgServer registers the MsgServer.
// This is a stub implementation until proper protobuf generation is set up.
func RegisterMsgServer(s grpc.Server, impl MsgServer) {
	_ = s
	_ = impl
}
