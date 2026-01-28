package types

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/gogoproto/grpc"
	ggrpc "google.golang.org/grpc"
)

// ModuleCdc is the codec for the benchmark module
var ModuleCdc = codec.NewLegacyAmino()

func init() {
	RegisterLegacyAminoCodec(ModuleCdc)
}

// RegisterLegacyAminoCodec registers the benchmark types
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSubmitBenchmarks{}, "benchmark/MsgSubmitBenchmarks", nil)
	cdc.RegisterConcrete(&MsgRequestChallenge{}, "benchmark/MsgRequestChallenge", nil)
	cdc.RegisterConcrete(&MsgRespondChallenge{}, "benchmark/MsgRespondChallenge", nil)
	cdc.RegisterConcrete(&MsgFlagProvider{}, "benchmark/MsgFlagProvider", nil)
	cdc.RegisterConcrete(&MsgUnflagProvider{}, "benchmark/MsgUnflagProvider", nil)
	cdc.RegisterConcrete(&MsgResolveAnomalyFlag{}, "benchmark/MsgResolveAnomalyFlag", nil)
}

// RegisterInterfaces registers the x/benchmark interfaces types with the interface registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// NOTE: These are stub message types without proper protobuf generation.
	// They don't have proper typeURLs (XXX_MessageName() methods), so we cannot
	// register them with RegisterImplementations. This will cause typeURL "/" conflicts.
	//
	// Once proper .proto files are generated with protoc-gen-gogo, this should be:
	//
	// registry.RegisterImplementations((*sdk.Msg)(nil),
	//     &MsgSubmitBenchmarks{},
	//     &MsgRequestChallenge{},
	//     &MsgRespondChallenge{},
	//     &MsgFlagProvider{},
	//     &MsgUnflagProvider{},
	//     &MsgResolveAnomalyFlag{},
	// )
	//
	// msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)

	// Suppress unused import warning
	_ = msgservice.RegisterMsgServiceDesc
}

// MsgServer is the interface for the benchmark message server
type MsgServer interface {
	// SubmitBenchmarks submits one or more benchmark reports
	SubmitBenchmarks(ctx context.Context, msg *MsgSubmitBenchmarks) (*MsgSubmitBenchmarksResponse, error)
	// RequestChallenge creates a new benchmark challenge for a provider
	RequestChallenge(ctx context.Context, msg *MsgRequestChallenge) (*MsgRequestChallengeResponse, error)
	// RespondChallenge responds to a benchmark challenge
	RespondChallenge(ctx context.Context, msg *MsgRespondChallenge) (*MsgRespondChallengeResponse, error)
	// FlagProvider flags a provider for performance issues
	FlagProvider(ctx context.Context, msg *MsgFlagProvider) (*MsgFlagProviderResponse, error)
	// UnflagProvider removes a flag from a provider
	UnflagProvider(ctx context.Context, msg *MsgUnflagProvider) (*MsgUnflagProviderResponse, error)
	// ResolveAnomalyFlag resolves an anomaly flag
	ResolveAnomalyFlag(ctx context.Context, msg *MsgResolveAnomalyFlag) (*MsgResolveAnomalyFlagResponse, error)
}

// RegisterMsgServer registers the MsgServer implementation with the grpc.Server.
func RegisterMsgServer(s grpc.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc_grpc, srv)
}

// _Msg_serviceDesc_grpc is the proper grpc.ServiceDesc for Msg service.
var _Msg_serviceDesc_grpc = ggrpc.ServiceDesc{
	ServiceName: "virtengine.benchmark.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []ggrpc.MethodDesc{
		{MethodName: "SubmitBenchmarks", Handler: _Msg_SubmitBenchmarks_Handler},
		{MethodName: "RequestChallenge", Handler: _Msg_RequestChallenge_Handler},
		{MethodName: "RespondChallenge", Handler: _Msg_RespondChallenge_Handler},
		{MethodName: "FlagProvider", Handler: _Msg_FlagProvider_Handler},
		{MethodName: "UnflagProvider", Handler: _Msg_UnflagProvider_Handler},
		{MethodName: "ResolveAnomalyFlag", Handler: _Msg_ResolveAnomalyFlag_Handler},
	},
	Streams:  []ggrpc.StreamDesc{},
	Metadata: "virtengine/benchmark/v1/msg.proto",
}

func _Msg_SubmitBenchmarks_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSubmitBenchmarks)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SubmitBenchmarks(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.benchmark.v1.Msg/SubmitBenchmarks"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SubmitBenchmarks(ctx, req.(*MsgSubmitBenchmarks))
	})
}

func _Msg_RequestChallenge_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRequestChallenge)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RequestChallenge(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.benchmark.v1.Msg/RequestChallenge"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RequestChallenge(ctx, req.(*MsgRequestChallenge))
	})
}

func _Msg_RespondChallenge_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRespondChallenge)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RespondChallenge(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.benchmark.v1.Msg/RespondChallenge"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RespondChallenge(ctx, req.(*MsgRespondChallenge))
	})
}

func _Msg_FlagProvider_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgFlagProvider)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).FlagProvider(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.benchmark.v1.Msg/FlagProvider"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).FlagProvider(ctx, req.(*MsgFlagProvider))
	})
}

func _Msg_UnflagProvider_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUnflagProvider)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UnflagProvider(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.benchmark.v1.Msg/UnflagProvider"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UnflagProvider(ctx, req.(*MsgUnflagProvider))
	})
}

func _Msg_ResolveAnomalyFlag_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgResolveAnomalyFlag)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).ResolveAnomalyFlag(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.benchmark.v1.Msg/ResolveAnomalyFlag"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).ResolveAnomalyFlag(ctx, req.(*MsgResolveAnomalyFlag))
	})
}
