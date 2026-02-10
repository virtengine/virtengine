package v1beta4

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

func init() {
	const (
		pkgName  = "virtengine.provider.v1beta4"
		fileName = "virtengine/provider/v1beta4/service.proto"
	)

	messageNames := []string{
		"MsgCreateProvider",
		"MsgCreateProviderResponse",
		"MsgUpdateProvider",
		"MsgUpdateProviderResponse",
		"MsgDeleteProvider",
		"MsgDeleteProviderResponse",
		"MsgGenerateDomainVerificationToken",
		"MsgGenerateDomainVerificationTokenResponse",
		"MsgVerifyProviderDomain",
		"MsgVerifyProviderDomainResponse",
		"MsgRequestDomainVerification",
		"MsgRequestDomainVerificationResponse",
		"MsgConfirmDomainVerification",
		"MsgConfirmDomainVerificationResponse",
		"MsgRevokeDomainVerification",
		"MsgRevokeDomainVerificationResponse",
	}

	methods := []struct {
		name   string
		input  string
		output string
	}{
		{"CreateProvider", "MsgCreateProvider", "MsgCreateProviderResponse"},
		{"UpdateProvider", "MsgUpdateProvider", "MsgUpdateProviderResponse"},
		{"DeleteProvider", "MsgDeleteProvider", "MsgDeleteProviderResponse"},
		{"GenerateDomainVerificationToken", "MsgGenerateDomainVerificationToken", "MsgGenerateDomainVerificationTokenResponse"},
		{"VerifyProviderDomain", "MsgVerifyProviderDomain", "MsgVerifyProviderDomainResponse"},
		{"RequestDomainVerification", "MsgRequestDomainVerification", "MsgRequestDomainVerificationResponse"},
		{"ConfirmDomainVerification", "MsgConfirmDomainVerification", "MsgConfirmDomainVerificationResponse"},
		{"RevokeDomainVerification", "MsgRevokeDomainVerification", "MsgRevokeDomainVerificationResponse"},
	}

	file := &descriptorpb.FileDescriptorProto{
		Syntax:  proto.String("proto3"),
		Name:    proto.String(fileName),
		Package: proto.String(pkgName),
	}

	for _, name := range messageNames {
		file.MessageType = append(file.MessageType, &descriptorpb.DescriptorProto{Name: proto.String(name)})
	}

	svc := &descriptorpb.ServiceDescriptorProto{Name: proto.String("Msg")}
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
