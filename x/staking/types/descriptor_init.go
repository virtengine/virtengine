package types

import (
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

func init() {
	const filePath = "virtengine/staking/v1/query.proto"
	if _, err := protoregistry.GlobalFiles.FindFileByPath(filePath); err == nil {
		return
	}

	messageNames := []string{
		"QueryParamsRequest",
		"QueryParamsResponse",
		"QueryValidatorPerformanceRequest",
		"QueryValidatorPerformanceResponse",
		"QueryValidatorPerformancesRequest",
		"QueryValidatorPerformancesResponse",
		"QueryValidatorRewardRequest",
		"QueryValidatorRewardResponse",
		"QueryValidatorRewardsRequest",
		"QueryValidatorRewardsResponse",
		"QueryRewardEpochRequest",
		"QueryRewardEpochResponse",
		"QuerySlashRecordsRequest",
		"QuerySlashRecordsResponse",
		"QuerySigningInfoRequest",
		"QuerySigningInfoResponse",
		"QueryCurrentEpochRequest",
		"QueryCurrentEpochResponse",
	}

	messages := make([]*descriptorpb.DescriptorProto, 0, len(messageNames))
	for _, name := range messageNames {
		messages = append(messages, &descriptorpb.DescriptorProto{
			Name: proto.String(name),
		})
	}

	methods := []*descriptorpb.MethodDescriptorProto{
		methodDesc("Params", "QueryParamsRequest", "QueryParamsResponse"),
		methodDesc("ValidatorPerformance", "QueryValidatorPerformanceRequest", "QueryValidatorPerformanceResponse"),
		methodDesc("ValidatorPerformances", "QueryValidatorPerformancesRequest", "QueryValidatorPerformancesResponse"),
		methodDesc("ValidatorReward", "QueryValidatorRewardRequest", "QueryValidatorRewardResponse"),
		methodDesc("ValidatorRewards", "QueryValidatorRewardsRequest", "QueryValidatorRewardsResponse"),
		methodDesc("RewardEpoch", "QueryRewardEpochRequest", "QueryRewardEpochResponse"),
		methodDesc("SlashRecords", "QuerySlashRecordsRequest", "QuerySlashRecordsResponse"),
		methodDesc("SigningInfo", "QuerySigningInfoRequest", "QuerySigningInfoResponse"),
		methodDesc("CurrentEpoch", "QueryCurrentEpochRequest", "QueryCurrentEpochResponse"),
	}

	fdProto := &descriptorpb.FileDescriptorProto{
		Syntax:      proto.String("proto3"),
		Name:        proto.String(filePath),
		Package:     proto.String("virtengine.staking.v1"),
		MessageType: messages,
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name:   proto.String("Query"),
				Method: methods,
			},
		},
	}

	fd, err := protodesc.NewFile(fdProto, protoregistry.GlobalFiles)
	if err != nil {
		return
	}
	_ = protoregistry.GlobalFiles.RegisterFile(fd)
}

func methodDesc(name, input, output string) *descriptorpb.MethodDescriptorProto {
	pkg := ".virtengine.staking.v1."
	return &descriptorpb.MethodDescriptorProto{
		Name:       proto.String(name),
		InputType:  proto.String(pkg + strings.TrimPrefix(input, pkg)),
		OutputType: proto.String(pkg + strings.TrimPrefix(output, pkg)),
	}
}
