// Package types registers a minimal protobuf descriptor for staking query services.
package types

import (
	"bytes"
	"compress/gzip"

	gogoproto "github.com/cosmos/gogoproto/proto"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const stakingQueryProtoFile = "virtengine/staking/v1/query.proto"

func init() {
	fd := &descriptorpb.FileDescriptorProto{
		Syntax:  protov2.String("proto3"),
		Name:    protov2.String(stakingQueryProtoFile),
		Package: protov2.String("virtengine.staking.v1"),
		MessageType: []*descriptorpb.DescriptorProto{
			message("QueryParamsRequest"),
			message("QueryParamsResponse"),
			message("QueryValidatorPerformanceRequest"),
			message("QueryValidatorPerformanceResponse"),
			message("QueryValidatorPerformancesRequest"),
			message("QueryValidatorPerformancesResponse"),
			message("QueryValidatorRewardRequest"),
			message("QueryValidatorRewardResponse"),
			message("QueryValidatorRewardsRequest"),
			message("QueryValidatorRewardsResponse"),
			message("QueryRewardEpochRequest"),
			message("QueryRewardEpochResponse"),
			message("QuerySlashRecordsRequest"),
			message("QuerySlashRecordsResponse"),
			message("QuerySigningInfoRequest"),
			message("QuerySigningInfoResponse"),
			message("QueryCurrentEpochRequest"),
			message("QueryCurrentEpochResponse"),
		},
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name: protov2.String("Query"),
				Method: []*descriptorpb.MethodDescriptorProto{
					method("Params", "QueryParamsRequest", "QueryParamsResponse"),
					method("ValidatorPerformance", "QueryValidatorPerformanceRequest", "QueryValidatorPerformanceResponse"),
					method("ValidatorPerformances", "QueryValidatorPerformancesRequest", "QueryValidatorPerformancesResponse"),
					method("ValidatorReward", "QueryValidatorRewardRequest", "QueryValidatorRewardResponse"),
					method("ValidatorRewards", "QueryValidatorRewardsRequest", "QueryValidatorRewardsResponse"),
					method("RewardEpoch", "QueryRewardEpochRequest", "QueryRewardEpochResponse"),
					method("SlashRecords", "QuerySlashRecordsRequest", "QuerySlashRecordsResponse"),
					method("SigningInfo", "QuerySigningInfoRequest", "QuerySigningInfoResponse"),
					method("CurrentEpoch", "QueryCurrentEpochRequest", "QueryCurrentEpochResponse"),
				},
			},
		},
	}

	raw, err := protov2.Marshal(fd)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(raw); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}

	gogoproto.RegisterFile(stakingQueryProtoFile, buf.Bytes())
}

func message(name string) *descriptorpb.DescriptorProto {
	return &descriptorpb.DescriptorProto{
		Name: protov2.String(name),
	}
}

func method(name, input, output string) *descriptorpb.MethodDescriptorProto {
	return &descriptorpb.MethodDescriptorProto{
		Name:       protov2.String(name),
		InputType:  protov2.String(".virtengine.staking.v1." + input),
		OutputType: protov2.String(".virtengine.staking.v1." + output),
	}
}
