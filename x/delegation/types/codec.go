// Package types contains types for the delegation module.
//
// VE-922: Delegation module codec
package types

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	grpc "google.golang.org/grpc"
)

// MsgServer defines the delegation Msg service
type MsgServer interface {
	Delegate(context.Context, *MsgDelegate) (*MsgDelegateResponse, error)
	Undelegate(context.Context, *MsgUndelegate) (*MsgUndelegateResponse, error)
	Redelegate(context.Context, *MsgRedelegate) (*MsgRedelegateResponse, error)
	ClaimRewards(context.Context, *MsgClaimRewards) (*MsgClaimRewardsResponse, error)
	ClaimAllRewards(context.Context, *MsgClaimAllRewards) (*MsgClaimAllRewardsResponse, error)
	UpdateParams(context.Context, *MsgUpdateParams) (*MsgUpdateParamsResponse, error)
}

// RegisterMsgServer registers the MsgServer implementation with a gRPC server
func RegisterMsgServer(s grpc.ServiceRegistrar, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc_grpc, srv)
}

// _Msg_serviceDesc_grpc is the service descriptor for grpc registration
var _Msg_serviceDesc_grpc = grpc.ServiceDesc{
	ServiceName: "virtengine.delegation.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams:     []grpc.StreamDesc{},
	Metadata:    "virtengine/delegation/v1/tx.proto",
}

// RegisterLegacyAminoCodec registers the delegation types on the provided LegacyAmino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register messages
	cdc.RegisterConcrete(&MsgUpdateParams{}, "virt_delegation/UpdateParams", nil)
	cdc.RegisterConcrete(&MsgDelegate{}, "virt_delegation/Delegate", nil)
	cdc.RegisterConcrete(&MsgUndelegate{}, "virt_delegation/Undelegate", nil)
	cdc.RegisterConcrete(&MsgRedelegate{}, "virt_delegation/Redelegate", nil)
	cdc.RegisterConcrete(&MsgClaimRewards{}, "virt_delegation/ClaimRewards", nil)
	cdc.RegisterConcrete(&MsgClaimAllRewards{}, "virt_delegation/ClaimAllRewards", nil)
}

// RegisterInterfaces registers the delegation types and interfaces
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// NOTE: These are stub message types without proper protobuf generation.
	// They don't have proper typeURLs (XXX_MessageName() methods), so we cannot
	// register them with RegisterImplementations. This will cause typeURL "/" conflicts.
	//
	// Once proper .proto files are generated with protoc-gen-gogo, this should be:
	//
	// registry.RegisterImplementations((*sdk.Msg)(nil),
	//     &MsgUpdateParams{},
	//     &MsgDelegate{},
	//     &MsgUndelegate{},
	//     &MsgRedelegate{},
	//     &MsgClaimRewards{},
	//     &MsgClaimAllRewards{},
	// )
	//
	// msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
	_ = registry // suppress unused variable warning
	_ = msgservice.RegisterMsgServiceDesc
}

// _Msg_serviceDesc is a placeholder for the message service descriptor
var _Msg_serviceDesc = struct {
	ServiceName string
}{
	ServiceName: "virtengine.delegation.v1.Msg",
}
