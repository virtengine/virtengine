package types

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

func init() {
	const (
		pkgName  = "virtengine.delegation.v1"
		fileName = "virtengine/delegation/v1/query.proto"
	)

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

	methods := []struct {
		name   string
		input  string
		output string
	}{
		{"Params", "QueryParamsRequest", "QueryParamsResponse"},
		{"Delegation", "QueryDelegationRequest", "QueryDelegationResponse"},
		{"DelegatorDelegations", "QueryDelegatorDelegationsRequest", "QueryDelegatorDelegationsResponse"},
		{"ValidatorDelegations", "QueryValidatorDelegationsRequest", "QueryValidatorDelegationsResponse"},
		{"UnbondingDelegation", "QueryUnbondingDelegationRequest", "QueryUnbondingDelegationResponse"},
		{"DelegatorUnbondingDelegations", "QueryDelegatorUnbondingDelegationsRequest", "QueryDelegatorUnbondingDelegationsResponse"},
		{"Redelegation", "QueryRedelegationRequest", "QueryRedelegationResponse"},
		{"DelegatorRedelegations", "QueryDelegatorRedelegationsRequest", "QueryDelegatorRedelegationsResponse"},
		{"Redelegations", "QueryRedelegationsRequest", "QueryRedelegationsResponse"},
		{"DelegatorRewards", "QueryDelegatorRewardsRequest", "QueryDelegatorRewardsResponse"},
		{"HistoricalRewards", "QueryHistoricalRewardsRequest", "QueryHistoricalRewardsResponse"},
		{"SlashingEvents", "QuerySlashingEventsRequest", "QuerySlashingEventsResponse"},
		{"DelegatorAllRewards", "QueryDelegatorAllRewardsRequest", "QueryDelegatorAllRewardsResponse"},
		{"ValidatorShares", "QueryValidatorSharesRequest", "QueryValidatorSharesResponse"},
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
