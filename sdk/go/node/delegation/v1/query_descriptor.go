package v1

import (
	"bytes"
	"compress/gzip"

	gogoproto "github.com/cosmos/gogoproto/proto"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

func init() {
	if _, err := gogoproto.HybridResolver.FindDescriptorByName(
		protoreflect.FullName("virtengine.delegation.v1.Query.Redelegations"),
	); err == nil {
		return
	}

	fd := &descriptorpb.FileDescriptorProto{
		Name:    protoString("virtengine/delegation/v1/query.proto"),
		Package: protoString("virtengine.delegation.v1"),
		Syntax:  protoString("proto3"),
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name: protoString("Query"),
				Method: []*descriptorpb.MethodDescriptorProto{
					{Name: protoString("Params"), InputType: protoString(".virtengine.delegation.v1.QueryParamsRequest"), OutputType: protoString(".virtengine.delegation.v1.QueryParamsResponse")},
					{Name: protoString("Delegation"), InputType: protoString(".virtengine.delegation.v1.QueryDelegationRequest"), OutputType: protoString(".virtengine.delegation.v1.QueryDelegationResponse")},
					{Name: protoString("DelegatorDelegations"), InputType: protoString(".virtengine.delegation.v1.QueryDelegatorDelegationsRequest"), OutputType: protoString(".virtengine.delegation.v1.QueryDelegatorDelegationsResponse")},
					{Name: protoString("ValidatorDelegations"), InputType: protoString(".virtengine.delegation.v1.QueryValidatorDelegationsRequest"), OutputType: protoString(".virtengine.delegation.v1.QueryValidatorDelegationsResponse")},
					{Name: protoString("UnbondingDelegation"), InputType: protoString(".virtengine.delegation.v1.QueryUnbondingDelegationRequest"), OutputType: protoString(".virtengine.delegation.v1.QueryUnbondingDelegationResponse")},
					{Name: protoString("DelegatorUnbondingDelegations"), InputType: protoString(".virtengine.delegation.v1.QueryDelegatorUnbondingDelegationsRequest"), OutputType: protoString(".virtengine.delegation.v1.QueryDelegatorUnbondingDelegationsResponse")},
					{Name: protoString("Redelegation"), InputType: protoString(".virtengine.delegation.v1.QueryRedelegationRequest"), OutputType: protoString(".virtengine.delegation.v1.QueryRedelegationResponse")},
					{Name: protoString("DelegatorRedelegations"), InputType: protoString(".virtengine.delegation.v1.QueryDelegatorRedelegationsRequest"), OutputType: protoString(".virtengine.delegation.v1.QueryDelegatorRedelegationsResponse")},
					{Name: protoString("Redelegations"), InputType: protoString(".virtengine.delegation.v1.QueryRedelegationsRequest"), OutputType: protoString(".virtengine.delegation.v1.QueryRedelegationsResponse")},
					{Name: protoString("DelegatorRewards"), InputType: protoString(".virtengine.delegation.v1.QueryDelegatorRewardsRequest"), OutputType: protoString(".virtengine.delegation.v1.QueryDelegatorRewardsResponse")},
					{Name: protoString("HistoricalRewards"), InputType: protoString(".virtengine.delegation.v1.QueryHistoricalRewardsRequest"), OutputType: protoString(".virtengine.delegation.v1.QueryHistoricalRewardsResponse")},
					{Name: protoString("SlashingEvents"), InputType: protoString(".virtengine.delegation.v1.QuerySlashingEventsRequest"), OutputType: protoString(".virtengine.delegation.v1.QuerySlashingEventsResponse")},
					{Name: protoString("DelegatorAllRewards"), InputType: protoString(".virtengine.delegation.v1.QueryDelegatorAllRewardsRequest"), OutputType: protoString(".virtengine.delegation.v1.QueryDelegatorAllRewardsResponse")},
					{Name: protoString("ValidatorShares"), InputType: protoString(".virtengine.delegation.v1.QueryValidatorSharesRequest"), OutputType: protoString(".virtengine.delegation.v1.QueryValidatorSharesResponse")},
				},
			},
		},
	}

	raw, err := protov2.Marshal(fd)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	if _, err := gzw.Write(raw); err != nil {
		panic(err)
	}
	if err := gzw.Close(); err != nil {
		panic(err)
	}

	gogoproto.RegisterFile(fd.GetName(), buf.Bytes())
	if file, err := (protodesc.FileOptions{AllowUnresolvable: true}).New(fd, protoregistry.GlobalFiles); err == nil {
		if _, err := protoregistry.GlobalFiles.FindFileByPath(fd.GetName()); err != nil {
			_ = protoregistry.GlobalFiles.RegisterFile(file)
		}
	}
}

func protoString(value string) *string {
	return &value
}
