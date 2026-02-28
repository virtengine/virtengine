package types

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

func init() {
	const (
		pkgName  = "virtengine.hpc.v1"
		fileName = hpcWorkloadTemplateQueryProtoFile
	)

	messageNames := []string{
		"QueryWorkloadTemplateRequest",
		"QueryWorkloadTemplateResponse",
		"QueryWorkloadTemplatesRequest",
		"QueryWorkloadTemplatesResponse",
		"QueryWorkloadTemplatesByTypeRequest",
		"QueryWorkloadTemplatesByTypeResponse",
		"QueryWorkloadTemplatesByPublisherRequest",
		"QueryWorkloadTemplatesByPublisherResponse",
		"QueryApprovedWorkloadTemplatesRequest",
		"QueryApprovedWorkloadTemplatesResponse",
		"QueryWorkloadTemplateUsageRequest",
		"QueryWorkloadTemplateUsageResponse",
		"QuerySearchWorkloadTemplatesRequest",
		"QuerySearchWorkloadTemplatesResponse",
	}

	methods := []struct {
		name   string
		input  string
		output string
	}{
		{"WorkloadTemplate", "QueryWorkloadTemplateRequest", "QueryWorkloadTemplateResponse"},
		{"WorkloadTemplates", "QueryWorkloadTemplatesRequest", "QueryWorkloadTemplatesResponse"},
		{"WorkloadTemplatesByType", "QueryWorkloadTemplatesByTypeRequest", "QueryWorkloadTemplatesByTypeResponse"},
		{"WorkloadTemplatesByPublisher", "QueryWorkloadTemplatesByPublisherRequest", "QueryWorkloadTemplatesByPublisherResponse"},
		{"ApprovedWorkloadTemplates", "QueryApprovedWorkloadTemplatesRequest", "QueryApprovedWorkloadTemplatesResponse"},
		{"WorkloadTemplateUsage", "QueryWorkloadTemplateUsageRequest", "QueryWorkloadTemplateUsageResponse"},
		{"SearchWorkloadTemplates", "QuerySearchWorkloadTemplatesRequest", "QuerySearchWorkloadTemplatesResponse"},
	}

	file := &descriptorpb.FileDescriptorProto{
		Syntax:  proto.String("proto3"),
		Name:    proto.String(fileName),
		Package: proto.String(pkgName),
	}

	for _, name := range messageNames {
		file.MessageType = append(file.MessageType, &descriptorpb.DescriptorProto{Name: proto.String(name)})
	}

	svc := &descriptorpb.ServiceDescriptorProto{Name: proto.String("WorkloadTemplateQuery")}
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
