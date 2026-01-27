package types

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

// RegisterLegacyAminoCodec registers the necessary interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgRegisterRecipientKey{}, "encryption/MsgRegisterRecipientKey")
	legacy.RegisterAminoMsg(cdc, &MsgRevokeRecipientKey{}, "encryption/MsgRevokeRecipientKey")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateKeyLabel{}, "encryption/MsgUpdateKeyLabel")
}

// RegisterInterfaces registers the interfaces types with the interface registry.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterRecipientKey{},
		&MsgRevokeRecipientKey{},
		&MsgUpdateKeyLabel{},
	)

	// TODO: Enable when protobuf generation is complete
	// msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
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
	ServiceName: "virtengine.encryption.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []struct {
		MethodName string
		Handler    interface{}
	}{
		{MethodName: "RegisterRecipientKey", Handler: nil},
		{MethodName: "RevokeRecipientKey", Handler: nil},
		{MethodName: "UpdateKeyLabel", Handler: nil},
	},
	Streams:  []struct{}{},
	Metadata: "virtengine/encryption/v1/msg.proto",
}

// MsgServer is the interface for the message server
type MsgServer interface {
	RegisterRecipientKey(ctx sdk.Context, msg *MsgRegisterRecipientKey) (*MsgRegisterRecipientKeyResponse, error)
	RevokeRecipientKey(ctx sdk.Context, msg *MsgRevokeRecipientKey) (*MsgRevokeRecipientKeyResponse, error)
	UpdateKeyLabel(ctx sdk.Context, msg *MsgUpdateKeyLabel) (*MsgUpdateKeyLabelResponse, error)
}

// RegisterMsgServer registers the MsgServer
// This is a stub implementation until proper protobuf generation is set up.
func RegisterMsgServer(s grpc.Server, impl MsgServer) {
	// Registration is a no-op for now since we don't have proper protobuf generated code
	_ = s
	_ = impl
}
