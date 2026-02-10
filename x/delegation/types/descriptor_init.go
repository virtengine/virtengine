package types

import (
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

func init() {
	const filePath = "virtengine/delegation/v1/query.proto"
	if _, err := protoregistry.GlobalFiles.FindFileByPath(filePath); err == nil {
		return
	}

	messageNames := []string{
		"QueryParamsRequest",
		"QueryParamsResponse",
		"QueryDelegationRequest",
		"QueryDelegationResponse",
		"QueryDelegatorDelegationsRequest",
		"QueryDelegatorDelegationsResponse",
		"QueryValidatorDelegationsRequest",
		"QueryValidatorDelegationsResponse",
		"QueryUnbondingDelegationRequest",
		"QueryUnbondingDelegationResponse",
		"QueryDelegatorUnbondingDelegationsRequest",
		"QueryDelegatorUnbondingDelegationsResponse",
		"QueryRedelegationRequest",
		"QueryRedelegationResponse",
		"QueryDelegatorRedelegationsRequest",
		"QueryDelegatorRedelegationsResponse",
		"QueryRedelegationsRequest",
		"QueryRedelegationsResponse",
		"QueryDelegatorRewardsRequest",
		"QueryDelegatorRewardsResponse",
		"QueryHistoricalRewardsRequest",
		"QueryHistoricalRewardsResponse",
		"QuerySlashingEventsRequest",
		"QuerySlashingEventsResponse",
		"QueryDelegatorAllRewardsRequest",
		"QueryDelegatorAllRewardsResponse",
		"QueryValidatorSharesRequest",
		"QueryValidatorSharesResponse",
	}

	messages := make([]*descriptorpb.DescriptorProto, 0, len(messageNames))
	for _, name := range messageNames {
		messages = append(messages, &descriptorpb.DescriptorProto{
			Name: proto.String(name),
		})
	}

	methods := []*descriptorpb.MethodDescriptorProto{
		methodDesc("Params", "QueryParamsRequest", "QueryParamsResponse"),
		methodDesc("Delegation", "QueryDelegationRequest", "QueryDelegationResponse"),
		methodDesc("DelegatorDelegations", "QueryDelegatorDelegationsRequest", "QueryDelegatorDelegationsResponse"),
		methodDesc("ValidatorDelegations", "QueryValidatorDelegationsRequest", "QueryValidatorDelegationsResponse"),
		methodDesc("UnbondingDelegation", "QueryUnbondingDelegationRequest", "QueryUnbondingDelegationResponse"),
		methodDesc("DelegatorUnbondingDelegations", "QueryDelegatorUnbondingDelegationsRequest", "QueryDelegatorUnbondingDelegationsResponse"),
		methodDesc("Redelegation", "QueryRedelegationRequest", "QueryRedelegationResponse"),
		methodDesc("DelegatorRedelegations", "QueryDelegatorRedelegationsRequest", "QueryDelegatorRedelegationsResponse"),
		methodDesc("Redelegations", "QueryRedelegationsRequest", "QueryRedelegationsResponse"),
		methodDesc("DelegatorRewards", "QueryDelegatorRewardsRequest", "QueryDelegatorRewardsResponse"),
		methodDesc("HistoricalRewards", "QueryHistoricalRewardsRequest", "QueryHistoricalRewardsResponse"),
		methodDesc("SlashingEvents", "QuerySlashingEventsRequest", "QuerySlashingEventsResponse"),
		methodDesc("DelegatorAllRewards", "QueryDelegatorAllRewardsRequest", "QueryDelegatorAllRewardsResponse"),
		methodDesc("ValidatorShares", "QueryValidatorSharesRequest", "QueryValidatorSharesResponse"),
	}

	fdProto := &descriptorpb.FileDescriptorProto{
		Syntax:      proto.String("proto3"),
		Name:        proto.String(filePath),
		Package:     proto.String("virtengine.delegation.v1"),
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
	pkg := ".virtengine.delegation.v1."
	return &descriptorpb.MethodDescriptorProto{
		Name:       proto.String(name),
		InputType:  proto.String(pkg + strings.TrimPrefix(input, pkg)),
		OutputType: proto.String(pkg + strings.TrimPrefix(output, pkg)),
	}
}
