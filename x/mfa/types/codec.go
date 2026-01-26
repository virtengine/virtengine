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
	legacy.RegisterAminoMsg(cdc, &MsgEnrollFactor{}, "mfa/MsgEnrollFactor")
	legacy.RegisterAminoMsg(cdc, &MsgRevokeFactor{}, "mfa/MsgRevokeFactor")
	legacy.RegisterAminoMsg(cdc, &MsgSetMFAPolicy{}, "mfa/MsgSetMFAPolicy")
	legacy.RegisterAminoMsg(cdc, &MsgCreateChallenge{}, "mfa/MsgCreateChallenge")
	legacy.RegisterAminoMsg(cdc, &MsgVerifyChallenge{}, "mfa/MsgVerifyChallenge")
	legacy.RegisterAminoMsg(cdc, &MsgAddTrustedDevice{}, "mfa/MsgAddTrustedDevice")
	legacy.RegisterAminoMsg(cdc, &MsgRemoveTrustedDevice{}, "mfa/MsgRemoveTrustedDevice")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateSensitiveTxConfig{}, "mfa/MsgUpdateSensitiveTxConfig")
}

// RegisterInterfaces registers the interfaces types with the interface registry.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgEnrollFactor{},
		&MsgRevokeFactor{},
		&MsgSetMFAPolicy{},
		&MsgCreateChallenge{},
		&MsgVerifyChallenge{},
		&MsgAddTrustedDevice{},
		&MsgRemoveTrustedDevice{},
		&MsgUpdateSensitiveTxConfig{},
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
	ServiceName: "virtengine.mfa.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []struct {
		MethodName string
		Handler    interface{}
	}{
		{MethodName: "EnrollFactor", Handler: nil},
		{MethodName: "RevokeFactor", Handler: nil},
		{MethodName: "SetMFAPolicy", Handler: nil},
		{MethodName: "CreateChallenge", Handler: nil},
		{MethodName: "VerifyChallenge", Handler: nil},
		{MethodName: "AddTrustedDevice", Handler: nil},
		{MethodName: "RemoveTrustedDevice", Handler: nil},
		{MethodName: "UpdateSensitiveTxConfig", Handler: nil},
	},
	Streams:  []struct{}{},
	Metadata: "virtengine/mfa/v1/tx.proto",
}
