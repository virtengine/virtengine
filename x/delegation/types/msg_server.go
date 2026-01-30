// Package types contains types for the delegation module.
//
// VE-922: Delegation module MsgServer interface using local types
package types

import (
	"context"

	grpc "google.golang.org/grpc"
)

// MsgServer defines the delegation Msg service using local types
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
	s.RegisterService(&_Msg_serviceDesc_local, srv)
}

// _Msg_serviceDesc_local is the service descriptor for local type grpc registration
var _Msg_serviceDesc_local = grpc.ServiceDesc{
	ServiceName: "virtengine.delegation.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams:     []grpc.StreamDesc{},
	Metadata:    "virtengine/delegation/v1/tx.proto",
}
