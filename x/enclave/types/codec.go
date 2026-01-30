package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	v1 "github.com/virtengine/virtengine/sdk/go/node/enclave/v1"
)

// RegisterLegacyAminoCodec registers the necessary interfaces and types
// for the enclave module on the provided LegacyAmino codec.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&v1.MsgRegisterEnclaveIdentity{}, "enclave/MsgRegisterEnclaveIdentity", nil)
	cdc.RegisterConcrete(&v1.MsgRotateEnclaveIdentity{}, "enclave/MsgRotateEnclaveIdentity", nil)
	cdc.RegisterConcrete(&v1.MsgProposeMeasurement{}, "enclave/MsgProposeMeasurement", nil)
	cdc.RegisterConcrete(&v1.MsgRevokeMeasurement{}, "enclave/MsgRevokeMeasurement", nil)
	cdc.RegisterConcrete(&v1.MsgUpdateParams{}, "enclave/MsgUpdateParams", nil)
	cdc.RegisterConcrete(&AddMeasurementProposal{}, "enclave/AddMeasurementProposal", nil)
	cdc.RegisterConcrete(&RevokeMeasurementProposal{}, "enclave/RevokeMeasurementProposal", nil)
}

// RegisterInterfaces registers the interfaces for the enclave module
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&v1.MsgRegisterEnclaveIdentity{},
		&v1.MsgRotateEnclaveIdentity{},
		&v1.MsgProposeMeasurement{},
		&v1.MsgRevokeMeasurement{},
		&v1.MsgUpdateParams{},
	)
	registry.RegisterImplementations(
		(*govv1beta1.Content)(nil),
		&AddMeasurementProposal{},
		&RevokeMeasurementProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &v1.Msg_serviceDesc)
}
