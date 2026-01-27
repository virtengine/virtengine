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
	legacy.RegisterAminoMsg(cdc, &MsgAssignRole{}, "roles/MsgAssignRole")
	legacy.RegisterAminoMsg(cdc, &MsgRevokeRole{}, "roles/MsgRevokeRole")
	legacy.RegisterAminoMsg(cdc, &MsgSetAccountState{}, "roles/MsgSetAccountState")
	legacy.RegisterAminoMsg(cdc, &MsgNominateAdmin{}, "roles/MsgNominateAdmin")
}

// RegisterInterfaces registers the interfaces types with the interface registry.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAssignRole{},
		&MsgRevokeRole{},
		&MsgSetAccountState{},
		&MsgNominateAdmin{},
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
	ServiceName: "virtengine.roles.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []struct {
		MethodName string
		Handler    interface{}
	}{
		{MethodName: "AssignRole", Handler: nil},
		{MethodName: "RevokeRole", Handler: nil},
		{MethodName: "SetAccountState", Handler: nil},
		{MethodName: "NominateAdmin", Handler: nil},
	},
	Streams:  []struct{}{},
	Metadata: "virtengine/roles/v1/msg.proto",
}

// MsgServer is the interface for the message server
type MsgServer interface {
	AssignRole(ctx sdk.Context, msg *MsgAssignRole) (*MsgAssignRoleResponse, error)
	RevokeRole(ctx sdk.Context, msg *MsgRevokeRole) (*MsgRevokeRoleResponse, error)
	SetAccountState(ctx sdk.Context, msg *MsgSetAccountState) (*MsgSetAccountStateResponse, error)
	NominateAdmin(ctx sdk.Context, msg *MsgNominateAdmin) (*MsgNominateAdminResponse, error)
}

// RegisterMsgServer registers the MsgServer
// This is a stub implementation until proper protobuf generation is set up.
func RegisterMsgServer(s grpc.Server, impl MsgServer) {
	// Registration is a no-op for now since we don't have proper protobuf generated code
	_ = s
	_ = impl
}
