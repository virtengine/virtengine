// Package types contains types for the HPC module.
package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

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
