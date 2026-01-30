package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// RegisterLegacyAminoCodec registers the necessary interfaces and types
// for the enclave module on the provided LegacyAmino codec.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterEnclaveIdentity{}, "enclave/MsgRegisterEnclaveIdentity", nil)
	cdc.RegisterConcrete(&MsgRotateEnclaveIdentity{}, "enclave/MsgRotateEnclaveIdentity", nil)
	cdc.RegisterConcrete(&MsgProposeMeasurement{}, "enclave/MsgProposeMeasurement", nil)
	cdc.RegisterConcrete(&MsgRevokeMeasurement{}, "enclave/MsgRevokeMeasurement", nil)
	cdc.RegisterConcrete(&AddMeasurementProposal{}, "enclave/AddMeasurementProposal", nil)
	cdc.RegisterConcrete(&RevokeMeasurementProposal{}, "enclave/RevokeMeasurementProposal", nil)
}

// RegisterInterfaces registers the interfaces for the enclave module
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// NOTE: These are stub message types without proper protobuf generation.
	// They don't have proper typeURLs (XXX_MessageName() methods), so we cannot
	// register them with RegisterImplementations. This will cause typeURL "/" conflicts.
	//
	// Once proper .proto files are generated with protoc-gen-gogo, this should be:
	//
	// registry.RegisterImplementations(
	//     (*sdk.Msg)(nil),
	//     &MsgRegisterEnclaveIdentity{},
	//     &MsgRotateEnclaveIdentity{},
	//     &MsgProposeMeasurement{},
	//     &MsgRevokeMeasurement{},
	// )
	//
	// msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
	registry.RegisterImplementations(
		(*govv1beta1.Content)(nil),
		&AddMeasurementProposal{},
		&RevokeMeasurementProposal{},
	)

	_ = registry // suppress unused variable warning
	_ = msgservice.RegisterMsgServiceDesc
}

// MsgService descriptor for enclave module messages
var _Msg_serviceDesc = struct {
	ServiceName string
	HandlerType interface{}
	Methods     []struct {
		MethodName   string
		Handler      interface{}
		MethodDesc   interface{}
		ServerStream bool
		ClientStream bool
	}
	Streams  []interface{}
	Metadata interface{}
}{
	ServiceName: "virtengine.enclave.v1.Msg",
}
