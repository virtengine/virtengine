package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
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

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
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
func RegisterMsgServer(s interface{ RegisterService(desc interface{}, impl interface{}) }, impl MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, impl)
}
