// Package types provides gRPC handlers for the staking module.
package types

import (
	"context"

	grpc1 "github.com/cosmos/gogoproto/grpc"
	grpc "google.golang.org/grpc"
)

// QueryServer is the interface for the staking query service.
type QueryServer interface {
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	ValidatorPerformance(context.Context, *QueryValidatorPerformanceRequest) (*QueryValidatorPerformanceResponse, error)
	ValidatorPerformances(context.Context, *QueryValidatorPerformancesRequest) (*QueryValidatorPerformancesResponse, error)
	ValidatorReward(context.Context, *QueryValidatorRewardRequest) (*QueryValidatorRewardResponse, error)
	ValidatorRewards(context.Context, *QueryValidatorRewardsRequest) (*QueryValidatorRewardsResponse, error)
	RewardEpoch(context.Context, *QueryRewardEpochRequest) (*QueryRewardEpochResponse, error)
	SlashRecords(context.Context, *QuerySlashRecordsRequest) (*QuerySlashRecordsResponse, error)
	SigningInfo(context.Context, *QuerySigningInfoRequest) (*QuerySigningInfoResponse, error)
	CurrentEpoch(context.Context, *QueryCurrentEpochRequest) (*QueryCurrentEpochResponse, error)
}

// RegisterQueryServer registers the QueryServer implementation with the grpc.Server.
func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_Params_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Params(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.staking.v1.Query/Params"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Params(ctx, req.(*QueryParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_ValidatorPerformance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryValidatorPerformanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ValidatorPerformance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.staking.v1.Query/ValidatorPerformance"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ValidatorPerformance(ctx, req.(*QueryValidatorPerformanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_ValidatorPerformances_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryValidatorPerformancesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ValidatorPerformances(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.staking.v1.Query/ValidatorPerformances"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ValidatorPerformances(ctx, req.(*QueryValidatorPerformancesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_ValidatorReward_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryValidatorRewardRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ValidatorReward(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.staking.v1.Query/ValidatorReward"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ValidatorReward(ctx, req.(*QueryValidatorRewardRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_ValidatorRewards_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryValidatorRewardsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ValidatorRewards(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.staking.v1.Query/ValidatorRewards"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ValidatorRewards(ctx, req.(*QueryValidatorRewardsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_RewardEpoch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryRewardEpochRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).RewardEpoch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.staking.v1.Query/RewardEpoch"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).RewardEpoch(ctx, req.(*QueryRewardEpochRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_SlashRecords_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QuerySlashRecordsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).SlashRecords(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.staking.v1.Query/SlashRecords"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).SlashRecords(ctx, req.(*QuerySlashRecordsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_SigningInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QuerySigningInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).SigningInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.staking.v1.Query/SigningInfo"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).SigningInfo(ctx, req.(*QuerySigningInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_CurrentEpoch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryCurrentEpochRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).CurrentEpoch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.staking.v1.Query/CurrentEpoch"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).CurrentEpoch(ctx, req.(*QueryCurrentEpochRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "virtengine.staking.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "Params", Handler: _Query_Params_Handler},
		{MethodName: "ValidatorPerformance", Handler: _Query_ValidatorPerformance_Handler},
		{MethodName: "ValidatorPerformances", Handler: _Query_ValidatorPerformances_Handler},
		{MethodName: "ValidatorReward", Handler: _Query_ValidatorReward_Handler},
		{MethodName: "ValidatorRewards", Handler: _Query_ValidatorRewards_Handler},
		{MethodName: "RewardEpoch", Handler: _Query_RewardEpoch_Handler},
		{MethodName: "SlashRecords", Handler: _Query_SlashRecords_Handler},
		{MethodName: "SigningInfo", Handler: _Query_SigningInfo_Handler},
		{MethodName: "CurrentEpoch", Handler: _Query_CurrentEpoch_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "virtengine/staking/v1/query.proto",
}
