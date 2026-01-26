package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the necessary interfaces and types
// for the enclave module on the provided LegacyAmino codec.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterEnclaveIdentity{}, "enclave/MsgRegisterEnclaveIdentity", nil)
	cdc.RegisterConcrete(&MsgRotateEnclaveIdentity{}, "enclave/MsgRotateEnclaveIdentity", nil)
	cdc.RegisterConcrete(&MsgProposeMeasurement{}, "enclave/MsgProposeMeasurement", nil)
	cdc.RegisterConcrete(&MsgRevokeMeasurement{}, "enclave/MsgRevokeMeasurement", nil)
}

// RegisterInterfaces registers the interfaces for the enclave module
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgRegisterEnclaveIdentity{},
		&MsgRotateEnclaveIdentity{},
		&MsgProposeMeasurement{},
		&MsgRevokeMeasurement{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
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
