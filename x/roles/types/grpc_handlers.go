// Package types provides gRPC handler implementations for the roles module.
//
// This file provides the MsgServer interface using local message types,
// and bridges to the generated protobuf handlers for gRPC registration.
package types

import (
	"context"
	"math"

	ggrpc "google.golang.org/grpc"

	"github.com/cosmos/gogoproto/grpc"

	mfav1 "github.com/virtengine/virtengine/sdk/go/node/mfa/v1"
	rolesv1 "github.com/virtengine/virtengine/sdk/go/node/roles/v1"
	mfatypes "github.com/virtengine/virtengine/x/mfa/types"
)

// MsgServer is the interface for the message server using local types.
// This interface is implemented by the keeper's msg_server.
type MsgServer interface {
	AssignRole(ctx context.Context, msg *MsgAssignRole) (*MsgAssignRoleResponse, error)
	RevokeRole(ctx context.Context, msg *MsgRevokeRole) (*MsgRevokeRoleResponse, error)
	SetAccountState(ctx context.Context, msg *MsgSetAccountState) (*MsgSetAccountStateResponse, error)
	NominateAdmin(ctx context.Context, msg *MsgNominateAdmin) (*MsgNominateAdminResponse, error)
	UpdateParams(ctx context.Context, msg *MsgUpdateParams) (*MsgUpdateParamsResponse, error)
}

// msgServerAdapter adapts the local MsgServer interface to the generated proto MsgServer
type msgServerAdapter struct {
	srv MsgServer
}

// Ensure msgServerAdapter implements rolesv1.MsgServer
var _ rolesv1.MsgServer = (*msgServerAdapter)(nil)

// AssignRole adapts the local type to the proto type
func (a *msgServerAdapter) AssignRole(ctx context.Context, req *rolesv1.MsgAssignRole) (*rolesv1.MsgAssignRoleResponse, error) {
	localReq := &MsgAssignRole{
		Sender:  req.Sender,
		Address: req.Address,
		Role:    req.Role,
	}
	_, err := a.srv.AssignRole(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &rolesv1.MsgAssignRoleResponse{}, nil
}

// RevokeRole adapts the local type to the proto type
func (a *msgServerAdapter) RevokeRole(ctx context.Context, req *rolesv1.MsgRevokeRole) (*rolesv1.MsgRevokeRoleResponse, error) {
	localReq := &MsgRevokeRole{
		Sender:  req.Sender,
		Address: req.Address,
		Role:    req.Role,
	}
	_, err := a.srv.RevokeRole(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &rolesv1.MsgRevokeRoleResponse{}, nil
}

// SetAccountState adapts the local type to the proto type
func (a *msgServerAdapter) SetAccountState(ctx context.Context, req *rolesv1.MsgSetAccountState) (*rolesv1.MsgSetAccountStateResponse, error) {
	localReq := &MsgSetAccountState{
		Sender:            req.Sender,
		Address:           req.Address,
		State:             req.State,
		Reason:            req.Reason,
		DeviceFingerprint: req.DeviceFingerprint,
	}
	// Handle MFAProof conversion if present
	if req.MfaProof != nil {
		localReq.MFAProof = convertMFAProofFromProto(req.MfaProof)
	}
	_, err := a.srv.SetAccountState(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &rolesv1.MsgSetAccountStateResponse{}, nil
}

// NominateAdmin adapts the local type to the proto type
func (a *msgServerAdapter) NominateAdmin(ctx context.Context, req *rolesv1.MsgNominateAdmin) (*rolesv1.MsgNominateAdminResponse, error) {
	localReq := &MsgNominateAdmin{
		Sender:  req.Sender,
		Address: req.Address,
	}
	_, err := a.srv.NominateAdmin(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &rolesv1.MsgNominateAdminResponse{}, nil
}

// UpdateParams adapts the local type to the proto type
func (a *msgServerAdapter) UpdateParams(ctx context.Context, req *rolesv1.MsgUpdateParams) (*rolesv1.MsgUpdateParamsResponse, error) {
	localReq := &MsgUpdateParams{
		Authority: req.Authority,
		Params:    convertParamsFromProto(&req.Params),
	}
	_, err := a.srv.UpdateParams(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &rolesv1.MsgUpdateParamsResponse{}, nil
}

// RegisterMsgServer registers the MsgServer implementation with the grpc.Server.
// This adapts the local MsgServer interface to the generated proto MsgServer.
func RegisterMsgServer(s grpc.Server, srv MsgServer) {
	adapter := &msgServerAdapter{srv: srv}
	rolesv1.RegisterMsgServer(s, adapter)
}

// _Msg_serviceDesc_grpc is the proper grpc.ServiceDesc for Msg service.
// This is used for direct gRPC registration without the adapter.
//
//nolint:unused // Reserved for direct gRPC registration
var _Msg_serviceDesc_grpc = ggrpc.ServiceDesc{
	ServiceName: "virtengine.roles.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []ggrpc.MethodDesc{
		{MethodName: "AssignRole", Handler: _Msg_AssignRole_Handler},
		{MethodName: "RevokeRole", Handler: _Msg_RevokeRole_Handler},
		{MethodName: "SetAccountState", Handler: _Msg_SetAccountState_Handler},
		{MethodName: "NominateAdmin", Handler: _Msg_NominateAdmin_Handler},
		{MethodName: "UpdateParams", Handler: _Msg_UpdateParams_Handler},
	},
	Streams:  []ggrpc.StreamDesc{},
	Metadata: "virtengine/roles/v1/tx.proto",
}

// The following handlers are used by _Msg_serviceDesc_grpc for direct gRPC registration.
// They are marked unused because the service descriptor is reserved for future use.

//nolint:unused // Handler for _Msg_serviceDesc_grpc
func _Msg_AssignRole_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgAssignRole)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).AssignRole(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.roles.v1.Msg/AssignRole"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).AssignRole(ctx, req.(*MsgAssignRole))
	})
}

//nolint:unused // Handler for _Msg_serviceDesc_grpc
func _Msg_RevokeRole_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRevokeRole)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RevokeRole(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.roles.v1.Msg/RevokeRole"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RevokeRole(ctx, req.(*MsgRevokeRole))
	})
}

//nolint:unused // Handler for _Msg_serviceDesc_grpc
func _Msg_SetAccountState_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSetAccountState)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SetAccountState(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.roles.v1.Msg/SetAccountState"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SetAccountState(ctx, req.(*MsgSetAccountState))
	})
}

//nolint:unused // Handler for _Msg_serviceDesc_grpc
func _Msg_NominateAdmin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgNominateAdmin)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).NominateAdmin(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.roles.v1.Msg/NominateAdmin"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).NominateAdmin(ctx, req.(*MsgNominateAdmin))
	})
}

//nolint:unused // Handler for _Msg_serviceDesc_grpc
func _Msg_UpdateParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateParams(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.roles.v1.Msg/UpdateParams"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateParams(ctx, req.(*MsgUpdateParams))
	})
}

// ============================================================================
// Conversion Functions
// ============================================================================

// convertMFAProofFromProto converts a proto MFAProof to the local mfatypes.MFAProof
func convertMFAProofFromProto(p *mfav1.MFAProof) *mfatypes.MFAProof {
	if p == nil {
		return nil
	}
	factors := make([]mfatypes.FactorType, len(p.VerifiedFactors))
	for i, f := range p.VerifiedFactors {
		factors[i] = safeFactorTypeFromProto(f)
	}
	return &mfatypes.MFAProof{
		SessionID:       p.SessionId,
		VerifiedFactors: factors,
		Timestamp:       p.Timestamp,
		Signature:       p.Signature,
	}
}

// convertParamsFromProto converts a proto Params to the local Params type
func convertParamsFromProto(p *rolesv1.Params) Params {
	if p == nil {
		return DefaultParams()
	}
	return Params{
		MaxRolesPerAccount: p.MaxRolesPerAccount,
		AllowSelfRevoke:    p.AllowSelfRevoke,
	}
}

func safeFactorTypeFromProto(value mfav1.FactorType) mfatypes.FactorType {
	intValue := int32(value)
	if intValue < 0 {
		return 0
	}
	if intValue > math.MaxUint8 {
		return mfatypes.FactorType(math.MaxUint8)
	}
	return mfatypes.FactorType(intValue)
}
