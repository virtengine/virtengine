package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
)

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterLegacyAminoCodec(amino)
}

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&AddVerifierVersionProposal{}, "veidregistry/AddVerifierVersionProposal", nil)
	cdc.RegisterConcrete(&ActivateVerifierProposal{}, "veidregistry/ActivateVerifierProposal", nil)
	cdc.RegisterConcrete(&UpdateParamsProposal{}, "veidregistry/UpdateParamsProposal", nil)
	legacy.RegisterAminoMsg(cdc, &MsgUpsertVerifierVersion{}, "vr/upsert")
	legacy.RegisterAminoMsg(cdc, &MsgApproveVerifierVersion{}, "vr/approve")
	legacy.RegisterAminoMsg(cdc, &MsgCancelVerifierVersion{}, "vr/cancel")
	legacy.RegisterAminoMsg(cdc, &MsgRetireVerifierVersion{}, "vr/retire")
	legacy.RegisterAminoMsg(cdc, &MsgReportValidatorReadiness{}, "vr/readiness")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "vr/params")
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {}
