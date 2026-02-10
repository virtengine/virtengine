package types

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

func init() {
	const (
		pkgName  = "virtengine.staking.v1"
		fileName = "virtengine/staking/v1/query.proto"
	)

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

	methods := []struct {
		name   string
		input  string
		output string
	}{
		{"Params", "QueryParamsRequest", "QueryParamsResponse"},
		{"ValidatorPerformance", "QueryValidatorPerformanceRequest", "QueryValidatorPerformanceResponse"},
		{"ValidatorPerformances", "QueryValidatorPerformancesRequest", "QueryValidatorPerformancesResponse"},
		{"ValidatorReward", "QueryValidatorRewardRequest", "QueryValidatorRewardResponse"},
		{"ValidatorRewards", "QueryValidatorRewardsRequest", "QueryValidatorRewardsResponse"},
		{"RewardEpoch", "QueryRewardEpochRequest", "QueryRewardEpochResponse"},
		{"SlashRecords", "QuerySlashRecordsRequest", "QuerySlashRecordsResponse"},
		{"SigningInfo", "QuerySigningInfoRequest", "QuerySigningInfoResponse"},
		{"CurrentEpoch", "QueryCurrentEpochRequest", "QueryCurrentEpochResponse"},
	}

	file := &descriptorpb.FileDescriptorProto{
		Syntax:  proto.String("proto3"),
		Name:    proto.String(fileName),
		Package: proto.String(pkgName),
	}

	for _, name := range messageNames {
		file.MessageType = append(file.MessageType, &descriptorpb.DescriptorProto{Name: proto.String(name)})
	}

	svc := &descriptorpb.ServiceDescriptorProto{Name: proto.String("Query")}
	for _, method := range methods {
		svc.Method = append(svc.Method, &descriptorpb.MethodDescriptorProto{
			Name:       proto.String(method.name),
			InputType:  proto.String("." + pkgName + "." + method.input),
			OutputType: proto.String("." + pkgName + "." + method.output),
		})
	}
	file.Service = []*descriptorpb.ServiceDescriptorProto{svc}

	fd, err := protodesc.NewFile(file, protoregistry.GlobalFiles)
	if err != nil {
		return
	}

	_ = protoregistry.GlobalFiles.RegisterFile(fd)
}
