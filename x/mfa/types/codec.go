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

// _Query_serviceDesc is the grpc.ServiceDesc for Query service.
var _Query_serviceDesc = struct {
	ServiceName string
	HandlerType interface{}
	Methods     []struct {
		MethodName string
		Handler    interface{}
	}
	Streams  []struct{}
	Metadata interface{}
}{
	ServiceName: "virtengine.mfa.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []struct {
		MethodName string
		Handler    interface{}
	}{
		{MethodName: "Params", Handler: nil},
		{MethodName: "MFAPolicy", Handler: nil},
		{MethodName: "AccountFactors", Handler: nil},
	},
	Streams:  []struct{}{},
	Metadata: "virtengine/mfa/v1/query.proto",
}

// RegisterMsgServer registers the MsgServer implementation with the grpc.Server.
// This is a stub implementation until proper protobuf generation is set up.
func RegisterMsgServer(s grpc.Server, srv MsgServer) {
	// Registration is a no-op for now since we don't have proper protobuf generated code
	_ = s
	_ = srv
}

// RegisterQueryServer registers the QueryServer implementation with the grpc.Server.
// This is a stub implementation until proper protobuf generation is set up.
func RegisterQueryServer(s grpc.Server, srv QueryServer) {
	// Registration is a no-op for now since we don't have proper protobuf generated code
	_ = s
	_ = srv
}
