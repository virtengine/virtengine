// Package types contains types for the Fraud module.
//
// VE-912: Fraud reporting flow - Codec registration
package types

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/gogoproto/grpc"
	ggrpc "google.golang.org/grpc"
)

// ModuleCdc is the codec for the module
var ModuleCdc = codec.NewProtoCodec(types.NewInterfaceRegistry())

// RegisterLegacyAminoCodec registers the necessary interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSubmitFraudReport{}, "fraud/MsgSubmitFraudReport", nil)
	cdc.RegisterConcrete(&MsgAssignModerator{}, "fraud/MsgAssignModerator", nil)
	cdc.RegisterConcrete(&MsgUpdateReportStatus{}, "fraud/MsgUpdateReportStatus", nil)
	cdc.RegisterConcrete(&MsgResolveFraudReport{}, "fraud/MsgResolveFraudReport", nil)
	cdc.RegisterConcrete(&MsgRejectFraudReport{}, "fraud/MsgRejectFraudReport", nil)
	cdc.RegisterConcrete(&MsgEscalateFraudReport{}, "fraud/MsgEscalateFraudReport", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "fraud/MsgUpdateParams", nil)
}

// RegisterInterfaces registers the interfaces for the module
func RegisterInterfaces(registry types.InterfaceRegistry) {
	// NOTE: These are stub message types without proper protobuf generation.
	// They don't have proper typeURLs (XXX_MessageName() methods), so we cannot
	// register them with RegisterImplementations. This will cause typeURL "/" conflicts.
	//
	// Once proper .proto files are generated with protoc-gen-gogo, this should be:
	//
	// registry.RegisterImplementations((*sdk.Msg)(nil),
	//     &MsgSubmitFraudReport{},
	//     &MsgAssignModerator{},
	//     &MsgUpdateReportStatus{},
	//     &MsgResolveFraudReport{},
	//     &MsgRejectFraudReport{},
	//     &MsgEscalateFraudReport{},
	//     &MsgUpdateParams{},
	// )
	//
	// msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
	_ = registry // suppress unused variable warning
	_ = msgservice.RegisterMsgServiceDesc
}

// _Msg_serviceDesc is the gRPC service descriptor for Msg service
var _Msg_serviceDesc = struct {
	ServiceName string
	HandlerType interface{}
	Methods     []struct {
		MethodName string
		Handler    interface{}
	}
	Streams  []struct{}
	Metadata string
}{
	ServiceName: "virtengine.fraud.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []struct {
		MethodName string
		Handler    interface{}
	}{},
	Streams:  []struct{}{},
	Metadata: "virtengine/fraud/v1/tx.proto",
}

// MsgServer is the server API for Msg service.
// Uses context.Context for gRPC compatibility (sdk.Context is unwrapped internally).
type MsgServer interface {
	SubmitFraudReport(context.Context, *MsgSubmitFraudReport) (*MsgSubmitFraudReportResponse, error)
	AssignModerator(context.Context, *MsgAssignModerator) (*MsgAssignModeratorResponse, error)
	UpdateReportStatus(context.Context, *MsgUpdateReportStatus) (*MsgUpdateReportStatusResponse, error)
	ResolveFraudReport(context.Context, *MsgResolveFraudReport) (*MsgResolveFraudReportResponse, error)
	RejectFraudReport(context.Context, *MsgRejectFraudReport) (*MsgRejectFraudReportResponse, error)
	EscalateFraudReport(context.Context, *MsgEscalateFraudReport) (*MsgEscalateFraudReportResponse, error)
	UpdateParams(context.Context, *MsgUpdateParams) (*MsgUpdateParamsResponse, error)
}

// Message response types
type MsgSubmitFraudReportResponse struct {
	ReportID string `json:"report_id"`
}

type MsgAssignModeratorResponse struct{}

type MsgUpdateReportStatusResponse struct{}

type MsgResolveFraudReportResponse struct{}

type MsgRejectFraudReportResponse struct{}

type MsgEscalateFraudReportResponse struct{}

type MsgUpdateParamsResponse struct{}

// RegisterMsgServer registers the MsgServer implementation with the grpc.Server.
func RegisterMsgServer(s grpc.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc_grpc, srv)
}

// _Msg_serviceDesc_grpc is the proper grpc.ServiceDesc for Msg service.
var _Msg_serviceDesc_grpc = ggrpc.ServiceDesc{
	ServiceName: "virtengine.fraud.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []ggrpc.MethodDesc{
		{MethodName: "SubmitFraudReport", Handler: _Msg_SubmitFraudReport_Handler},
		{MethodName: "AssignModerator", Handler: _Msg_AssignModerator_Handler},
		{MethodName: "UpdateReportStatus", Handler: _Msg_UpdateReportStatus_Handler},
		{MethodName: "ResolveFraudReport", Handler: _Msg_ResolveFraudReport_Handler},
		{MethodName: "RejectFraudReport", Handler: _Msg_RejectFraudReport_Handler},
		{MethodName: "EscalateFraudReport", Handler: _Msg_EscalateFraudReport_Handler},
		{MethodName: "UpdateParams", Handler: _Msg_UpdateParams_Handler},
	},
	Streams:  []ggrpc.StreamDesc{},
	Metadata: "virtengine/fraud/v1/tx.proto",
}

// Handler functions for MsgServer methods
func _Msg_SubmitFraudReport_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSubmitFraudReport)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SubmitFraudReport(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.fraud.v1.Msg/SubmitFraudReport"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SubmitFraudReport(ctx, req.(*MsgSubmitFraudReport))
	})
}

func _Msg_AssignModerator_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgAssignModerator)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).AssignModerator(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.fraud.v1.Msg/AssignModerator"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).AssignModerator(ctx, req.(*MsgAssignModerator))
	})
}

func _Msg_UpdateReportStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateReportStatus)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateReportStatus(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.fraud.v1.Msg/UpdateReportStatus"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateReportStatus(ctx, req.(*MsgUpdateReportStatus))
	})
}

func _Msg_ResolveFraudReport_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgResolveFraudReport)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).ResolveFraudReport(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.fraud.v1.Msg/ResolveFraudReport"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).ResolveFraudReport(ctx, req.(*MsgResolveFraudReport))
	})
}

func _Msg_RejectFraudReport_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRejectFraudReport)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RejectFraudReport(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.fraud.v1.Msg/RejectFraudReport"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RejectFraudReport(ctx, req.(*MsgRejectFraudReport))
	})
}

func _Msg_EscalateFraudReport_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgEscalateFraudReport)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).EscalateFraudReport(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.fraud.v1.Msg/EscalateFraudReport"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).EscalateFraudReport(ctx, req.(*MsgEscalateFraudReport))
	})
}

func _Msg_UpdateParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateParams(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.fraud.v1.Msg/UpdateParams"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateParams(ctx, req.(*MsgUpdateParams))
	})
}

// QueryServer is the server API for Query service.
type QueryServer interface{}

// RegisterQueryServer registers the QueryServer implementation with the grpc.Server.
func RegisterQueryServer(s grpc.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc_grpc, srv)
}

// _Query_serviceDesc_grpc is the proper grpc.ServiceDesc for Query service.
var _Query_serviceDesc_grpc = ggrpc.ServiceDesc{
	ServiceName: "virtengine.fraud.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods:     []ggrpc.MethodDesc{},
	Streams:     []ggrpc.StreamDesc{},
	Metadata:    "virtengine/fraud/v1/query.proto",
}
