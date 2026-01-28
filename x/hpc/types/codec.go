// Package types contains types for the HPC module.
package types

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	grpc "google.golang.org/grpc"
)

// MsgServer defines the HPC module Msg service
type MsgServer interface {
	RegisterCluster(context.Context, *MsgRegisterCluster) (*MsgRegisterClusterResponse, error)
	UpdateCluster(context.Context, *MsgUpdateCluster) (*MsgUpdateClusterResponse, error)
	DeregisterCluster(context.Context, *MsgDeregisterCluster) (*MsgDeregisterClusterResponse, error)
	CreateOffering(context.Context, *MsgCreateOffering) (*MsgCreateOfferingResponse, error)
	UpdateOffering(context.Context, *MsgUpdateOffering) (*MsgUpdateOfferingResponse, error)
	SubmitJob(context.Context, *MsgSubmitJob) (*MsgSubmitJobResponse, error)
	CancelJob(context.Context, *MsgCancelJob) (*MsgCancelJobResponse, error)
	ReportJobStatus(context.Context, *MsgReportJobStatus) (*MsgReportJobStatusResponse, error)
	UpdateNodeMetadata(context.Context, *MsgUpdateNodeMetadata) (*MsgUpdateNodeMetadataResponse, error)
	FlagDispute(context.Context, *MsgFlagDispute) (*MsgFlagDisputeResponse, error)
	ResolveDispute(context.Context, *MsgResolveDispute) (*MsgResolveDisputeResponse, error)
}

// RegisterMsgServer registers the MsgServer implementation with a gRPC server
func RegisterMsgServer(s grpc.ServiceRegistrar, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc_grpc, srv)
}

// _Msg_serviceDesc_grpc is the service descriptor for grpc registration
var _Msg_serviceDesc_grpc = grpc.ServiceDesc{
	ServiceName: "virtengine.hpc.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams:     []grpc.StreamDesc{},
	Metadata:    "virtengine/hpc/v1/tx.proto",
}

// RegisterLegacyAminoCodec registers the account types and interface
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterCluster{}, "hpc/MsgRegisterCluster", nil)
	cdc.RegisterConcrete(&MsgUpdateCluster{}, "hpc/MsgUpdateCluster", nil)
	cdc.RegisterConcrete(&MsgDeregisterCluster{}, "hpc/MsgDeregisterCluster", nil)
	cdc.RegisterConcrete(&MsgCreateOffering{}, "hpc/MsgCreateOffering", nil)
	cdc.RegisterConcrete(&MsgUpdateOffering{}, "hpc/MsgUpdateOffering", nil)
	cdc.RegisterConcrete(&MsgSubmitJob{}, "hpc/MsgSubmitJob", nil)
	cdc.RegisterConcrete(&MsgCancelJob{}, "hpc/MsgCancelJob", nil)
	cdc.RegisterConcrete(&MsgReportJobStatus{}, "hpc/MsgReportJobStatus", nil)
	cdc.RegisterConcrete(&MsgUpdateNodeMetadata{}, "hpc/MsgUpdateNodeMetadata", nil)
	cdc.RegisterConcrete(&MsgFlagDispute{}, "hpc/MsgFlagDispute", nil)
	cdc.RegisterConcrete(&MsgResolveDispute{}, "hpc/MsgResolveDispute", nil)
}

// RegisterInterfaces registers the x/hpc interfaces types with the interface registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// NOTE: These are stub message types without proper protobuf generation.
	// They don't have proper typeURLs (XXX_MessageName() methods), so we cannot
	// register them with RegisterImplementations. This will cause typeURL "/" conflicts.
	//
	// Once proper .proto files are generated with protoc-gen-gogo, this should be:
	//
	// registry.RegisterImplementations((*sdk.Msg)(nil),
	//     &MsgRegisterCluster{},
	//     &MsgUpdateCluster{},
	//     &MsgDeregisterCluster{},
	//     &MsgCreateOffering{},
	//     &MsgUpdateOffering{},
	//     &MsgSubmitJob{},
	//     &MsgCancelJob{},
	//     &MsgReportJobStatus{},
	//     &MsgUpdateNodeMetadata{},
	//     &MsgFlagDispute{},
	//     &MsgResolveDispute{},
	// )
	//
	// msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
	_ = registry // suppress unused variable warning
	_ = msgservice.RegisterMsgServiceDesc
}

// _Msg_serviceDesc is a placeholder for the message service descriptor
// This would normally be generated from protobuf definitions
var _Msg_serviceDesc = struct {
	ServiceName string
}{
	ServiceName: "virtengine.hpc.v1.Msg",
}
