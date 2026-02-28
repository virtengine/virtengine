// Package types registers a minimal protobuf descriptor for workload template query services.
package types

import (
	"bytes"
	"compress/gzip"

	gogoproto "github.com/cosmos/gogoproto/proto"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const hpcWorkloadTemplateQueryProtoFile = "virtengine/hpc/v1/query_template.proto"

func init() {
	gogoproto.RegisterType((*QueryWorkloadTemplateRequest)(nil), "virtengine.hpc.v1.QueryWorkloadTemplateRequest")
	gogoproto.RegisterType((*QueryWorkloadTemplateResponse)(nil), "virtengine.hpc.v1.QueryWorkloadTemplateResponse")
	gogoproto.RegisterType((*QueryWorkloadTemplatesRequest)(nil), "virtengine.hpc.v1.QueryWorkloadTemplatesRequest")
	gogoproto.RegisterType((*QueryWorkloadTemplatesResponse)(nil), "virtengine.hpc.v1.QueryWorkloadTemplatesResponse")
	gogoproto.RegisterType((*QueryWorkloadTemplatesByTypeRequest)(nil), "virtengine.hpc.v1.QueryWorkloadTemplatesByTypeRequest")
	gogoproto.RegisterType((*QueryWorkloadTemplatesByTypeResponse)(nil), "virtengine.hpc.v1.QueryWorkloadTemplatesByTypeResponse")
	gogoproto.RegisterType((*QueryWorkloadTemplatesByPublisherRequest)(nil), "virtengine.hpc.v1.QueryWorkloadTemplatesByPublisherRequest")
	gogoproto.RegisterType((*QueryWorkloadTemplatesByPublisherResponse)(nil), "virtengine.hpc.v1.QueryWorkloadTemplatesByPublisherResponse")
	gogoproto.RegisterType((*QueryApprovedWorkloadTemplatesRequest)(nil), "virtengine.hpc.v1.QueryApprovedWorkloadTemplatesRequest")
	gogoproto.RegisterType((*QueryApprovedWorkloadTemplatesResponse)(nil), "virtengine.hpc.v1.QueryApprovedWorkloadTemplatesResponse")
	gogoproto.RegisterType((*QueryWorkloadTemplateUsageRequest)(nil), "virtengine.hpc.v1.QueryWorkloadTemplateUsageRequest")
	gogoproto.RegisterType((*QueryWorkloadTemplateUsageResponse)(nil), "virtengine.hpc.v1.QueryWorkloadTemplateUsageResponse")
	gogoproto.RegisterType((*QuerySearchWorkloadTemplatesRequest)(nil), "virtengine.hpc.v1.QuerySearchWorkloadTemplatesRequest")
	gogoproto.RegisterType((*QuerySearchWorkloadTemplatesResponse)(nil), "virtengine.hpc.v1.QuerySearchWorkloadTemplatesResponse")

	fd := &descriptorpb.FileDescriptorProto{
		Syntax:  protov2.String("proto3"),
		Name:    protov2.String(hpcWorkloadTemplateQueryProtoFile),
		Package: protov2.String("virtengine.hpc.v1"),
		MessageType: []*descriptorpb.DescriptorProto{
			workloadTemplateDescriptorMessage("QueryWorkloadTemplateRequest"),
			workloadTemplateDescriptorMessage("QueryWorkloadTemplateResponse"),
			workloadTemplateDescriptorMessage("QueryWorkloadTemplatesRequest"),
			workloadTemplateDescriptorMessage("QueryWorkloadTemplatesResponse"),
			workloadTemplateDescriptorMessage("QueryWorkloadTemplatesByTypeRequest"),
			workloadTemplateDescriptorMessage("QueryWorkloadTemplatesByTypeResponse"),
			workloadTemplateDescriptorMessage("QueryWorkloadTemplatesByPublisherRequest"),
			workloadTemplateDescriptorMessage("QueryWorkloadTemplatesByPublisherResponse"),
			workloadTemplateDescriptorMessage("QueryApprovedWorkloadTemplatesRequest"),
			workloadTemplateDescriptorMessage("QueryApprovedWorkloadTemplatesResponse"),
			workloadTemplateDescriptorMessage("QueryWorkloadTemplateUsageRequest"),
			workloadTemplateDescriptorMessage("QueryWorkloadTemplateUsageResponse"),
			workloadTemplateDescriptorMessage("QuerySearchWorkloadTemplatesRequest"),
			workloadTemplateDescriptorMessage("QuerySearchWorkloadTemplatesResponse"),
		},
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name: protov2.String("WorkloadTemplateQuery"),
				Method: []*descriptorpb.MethodDescriptorProto{
					workloadTemplateDescriptorMethod("WorkloadTemplate", "QueryWorkloadTemplateRequest", "QueryWorkloadTemplateResponse"),
					workloadTemplateDescriptorMethod("WorkloadTemplates", "QueryWorkloadTemplatesRequest", "QueryWorkloadTemplatesResponse"),
					workloadTemplateDescriptorMethod("WorkloadTemplatesByType", "QueryWorkloadTemplatesByTypeRequest", "QueryWorkloadTemplatesByTypeResponse"),
					workloadTemplateDescriptorMethod("WorkloadTemplatesByPublisher", "QueryWorkloadTemplatesByPublisherRequest", "QueryWorkloadTemplatesByPublisherResponse"),
					workloadTemplateDescriptorMethod("ApprovedWorkloadTemplates", "QueryApprovedWorkloadTemplatesRequest", "QueryApprovedWorkloadTemplatesResponse"),
					workloadTemplateDescriptorMethod("WorkloadTemplateUsage", "QueryWorkloadTemplateUsageRequest", "QueryWorkloadTemplateUsageResponse"),
					workloadTemplateDescriptorMethod("SearchWorkloadTemplates", "QuerySearchWorkloadTemplatesRequest", "QuerySearchWorkloadTemplatesResponse"),
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

	gogoproto.RegisterFile(hpcWorkloadTemplateQueryProtoFile, buf.Bytes())
}

func workloadTemplateDescriptorMessage(name string) *descriptorpb.DescriptorProto {
	return &descriptorpb.DescriptorProto{
		Name: protov2.String(name),
	}
}

func workloadTemplateDescriptorMethod(name, input, output string) *descriptorpb.MethodDescriptorProto {
	return &descriptorpb.MethodDescriptorProto{
		Name:       protov2.String(name),
		InputType:  protov2.String(".virtengine.hpc.v1." + input),
		OutputType: protov2.String(".virtengine.hpc.v1." + output),
	}
}
