// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0
//
// Package inferencepb provides the gRPC service interface for VEID inference.
// VE-219: Deterministic identity verification runtime

package inferencepb

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InferenceServiceClient is the client API for InferenceService.
type InferenceServiceClient interface {
	// GetModelInfo returns information about the loaded model.
	GetModelInfo(ctx context.Context, in *GetModelInfoRequest, opts ...grpc.CallOption) (*GetModelInfoResponse, error)

	// ComputeScore runs ML inference to compute an identity score.
	ComputeScore(ctx context.Context, in *ComputeScoreRequest, opts ...grpc.CallOption) (*ComputeScoreResponse, error)

	// HealthCheck checks if the service is healthy and ready.
	HealthCheck(ctx context.Context, in *HealthCheckRequest, opts ...grpc.CallOption) (*HealthCheckResponse, error)

	// GetMetrics returns service metrics for observability.
	GetMetrics(ctx context.Context, in *GetMetricsRequest, opts ...grpc.CallOption) (*GetMetricsResponse, error)

	// VerifyDeterminism runs a determinism check with known test vectors.
	VerifyDeterminism(ctx context.Context, in *VerifyDeterminismRequest, opts ...grpc.CallOption) (*VerifyDeterminismResponse, error)
}

// InferenceServiceServer is the server API for InferenceService.
type InferenceServiceServer interface {
	// GetModelInfo returns information about the loaded model.
	GetModelInfo(context.Context, *GetModelInfoRequest) (*GetModelInfoResponse, error)

	// ComputeScore runs ML inference to compute an identity score.
	ComputeScore(context.Context, *ComputeScoreRequest) (*ComputeScoreResponse, error)

	// HealthCheck checks if the service is healthy and ready.
	HealthCheck(context.Context, *HealthCheckRequest) (*HealthCheckResponse, error)

	// GetMetrics returns service metrics for observability.
	GetMetrics(context.Context, *GetMetricsRequest) (*GetMetricsResponse, error)

	// VerifyDeterminism runs a determinism check with known test vectors.
	VerifyDeterminism(context.Context, *VerifyDeterminismRequest) (*VerifyDeterminismResponse, error)

	mustEmbedUnimplementedInferenceServiceServer()
}

// UnimplementedInferenceServiceServer must be embedded to have forward compatible implementations.
type UnimplementedInferenceServiceServer struct{}

func (UnimplementedInferenceServiceServer) GetModelInfo(context.Context, *GetModelInfoRequest) (*GetModelInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetModelInfo not implemented")
}

func (UnimplementedInferenceServiceServer) ComputeScore(context.Context, *ComputeScoreRequest) (*ComputeScoreResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ComputeScore not implemented")
}

func (UnimplementedInferenceServiceServer) HealthCheck(context.Context, *HealthCheckRequest) (*HealthCheckResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HealthCheck not implemented")
}

func (UnimplementedInferenceServiceServer) GetMetrics(context.Context, *GetMetricsRequest) (*GetMetricsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMetrics not implemented")
}

func (UnimplementedInferenceServiceServer) VerifyDeterminism(context.Context, *VerifyDeterminismRequest) (*VerifyDeterminismResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VerifyDeterminism not implemented")
}

func (UnimplementedInferenceServiceServer) mustEmbedUnimplementedInferenceServiceServer() {}

// UnsafeInferenceServiceServer may be embedded to opt out of forward compatibility.
type UnsafeInferenceServiceServer interface {
	mustEmbedUnimplementedInferenceServiceServer()
}

// ServiceName is the full service name for the InferenceService.
const ServiceName = "inference.v1.InferenceService"

// inferenceServiceClient implements InferenceServiceClient.
type inferenceServiceClient struct {
	cc grpc.ClientConnInterface
}

// NewInferenceServiceClient creates a new InferenceServiceClient.
func NewInferenceServiceClient(cc grpc.ClientConnInterface) InferenceServiceClient {
	return &inferenceServiceClient{cc}
}

func (c *inferenceServiceClient) GetModelInfo(ctx context.Context, in *GetModelInfoRequest, opts ...grpc.CallOption) (*GetModelInfoResponse, error) {
	out := new(GetModelInfoResponse)
	err := c.cc.Invoke(ctx, "/inference.v1.InferenceService/GetModelInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *inferenceServiceClient) ComputeScore(ctx context.Context, in *ComputeScoreRequest, opts ...grpc.CallOption) (*ComputeScoreResponse, error) {
	out := new(ComputeScoreResponse)
	err := c.cc.Invoke(ctx, "/inference.v1.InferenceService/ComputeScore", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *inferenceServiceClient) HealthCheck(ctx context.Context, in *HealthCheckRequest, opts ...grpc.CallOption) (*HealthCheckResponse, error) {
	out := new(HealthCheckResponse)
	err := c.cc.Invoke(ctx, "/inference.v1.InferenceService/HealthCheck", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *inferenceServiceClient) GetMetrics(ctx context.Context, in *GetMetricsRequest, opts ...grpc.CallOption) (*GetMetricsResponse, error) {
	out := new(GetMetricsResponse)
	err := c.cc.Invoke(ctx, "/inference.v1.InferenceService/GetMetrics", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *inferenceServiceClient) VerifyDeterminism(ctx context.Context, in *VerifyDeterminismRequest, opts ...grpc.CallOption) (*VerifyDeterminismResponse, error) {
	out := new(VerifyDeterminismResponse)
	err := c.cc.Invoke(ctx, "/inference.v1.InferenceService/VerifyDeterminism", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// InferenceServiceDesc is the service descriptor for InferenceService.
var InferenceServiceDesc = grpc.ServiceDesc{
	ServiceName: ServiceName,
	HandlerType: (*InferenceServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetModelInfo",
			Handler:    getModelInfoHandler,
		},
		{
			MethodName: "ComputeScore",
			Handler:    computeScoreHandler,
		},
		{
			MethodName: "HealthCheck",
			Handler:    healthCheckHandler,
		},
		{
			MethodName: "GetMetrics",
			Handler:    getMetricsHandler,
		},
		{
			MethodName: "VerifyDeterminism",
			Handler:    verifyDeterminismHandler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/inference/proto/inference.proto",
}

// RegisterInferenceServiceServer registers a service implementation with a gRPC server.
func RegisterInferenceServiceServer(s grpc.ServiceRegistrar, srv InferenceServiceServer) {
	s.RegisterService(&InferenceServiceDesc, srv)
}

func getModelInfoHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetModelInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InferenceServiceServer).GetModelInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/inference.v1.InferenceService/GetModelInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InferenceServiceServer).GetModelInfo(ctx, req.(*GetModelInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func computeScoreHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ComputeScoreRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InferenceServiceServer).ComputeScore(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/inference.v1.InferenceService/ComputeScore",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InferenceServiceServer).ComputeScore(ctx, req.(*ComputeScoreRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func healthCheckHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HealthCheckRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InferenceServiceServer).HealthCheck(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/inference.v1.InferenceService/HealthCheck",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InferenceServiceServer).HealthCheck(ctx, req.(*HealthCheckRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func getMetricsHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMetricsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InferenceServiceServer).GetMetrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/inference.v1.InferenceService/GetMetrics",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InferenceServiceServer).GetMetrics(ctx, req.(*GetMetricsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func verifyDeterminismHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VerifyDeterminismRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InferenceServiceServer).VerifyDeterminism(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/inference.v1.InferenceService/VerifyDeterminism",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InferenceServiceServer).VerifyDeterminism(ctx, req.(*VerifyDeterminismRequest))
	}
	return interceptor(ctx, in, info, handler)
}
