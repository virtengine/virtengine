package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
<<<<<<< HEAD

	v1 "github.com/virtengine/virtengine/sdk/go/node/enclave/v1"
=======
>>>>>>> 481fa029457ab6a2454257716c9bd2651a9bb202
)

// RegisterLegacyAminoCodec registers the necessary interfaces and types
// for the enclave module on the provided LegacyAmino codec.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
<<<<<<< HEAD
	cdc.RegisterConcrete(&v1.MsgRegisterEnclaveIdentity{}, "enclave/MsgRegisterEnclaveIdentity", nil)
	cdc.RegisterConcrete(&v1.MsgRotateEnclaveIdentity{}, "enclave/MsgRotateEnclaveIdentity", nil)
	cdc.RegisterConcrete(&v1.MsgProposeMeasurement{}, "enclave/MsgProposeMeasurement", nil)
	cdc.RegisterConcrete(&v1.MsgRevokeMeasurement{}, "enclave/MsgRevokeMeasurement", nil)
	cdc.RegisterConcrete(&v1.MsgUpdateParams{}, "enclave/MsgUpdateParams", nil)
=======
	cdc.RegisterConcrete(&MsgRegisterEnclaveIdentity{}, "enclave/MsgRegisterEnclaveIdentity", nil)
	cdc.RegisterConcrete(&MsgRotateEnclaveIdentity{}, "enclave/MsgRotateEnclaveIdentity", nil)
	cdc.RegisterConcrete(&MsgProposeMeasurement{}, "enclave/MsgProposeMeasurement", nil)
	cdc.RegisterConcrete(&MsgRevokeMeasurement{}, "enclave/MsgRevokeMeasurement", nil)
>>>>>>> 481fa029457ab6a2454257716c9bd2651a9bb202
	cdc.RegisterConcrete(&AddMeasurementProposal{}, "enclave/AddMeasurementProposal", nil)
	cdc.RegisterConcrete(&RevokeMeasurementProposal{}, "enclave/RevokeMeasurementProposal", nil)
}

// RegisterInterfaces registers the interfaces for the enclave module
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
<<<<<<< HEAD
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&v1.MsgRegisterEnclaveIdentity{},
		&v1.MsgRotateEnclaveIdentity{},
		&v1.MsgProposeMeasurement{},
		&v1.MsgRevokeMeasurement{},
		&v1.MsgUpdateParams{},
	)
=======
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
>>>>>>> 481fa029457ab6a2454257716c9bd2651a9bb202
	registry.RegisterImplementations(
		(*govv1beta1.Content)(nil),
		&AddMeasurementProposal{},
		&RevokeMeasurementProposal{},
	)
<<<<<<< HEAD
=======

	_ = registry // suppress unused variable warning
	_ = msgservice.RegisterMsgServiceDesc
}
>>>>>>> 481fa029457ab6a2454257716c9bd2651a9bb202

	msgservice.RegisterMsgServiceDesc(registry, &v1.Msg_serviceDesc)
}
