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
	cdc.RegisterConcrete(&UpsertPolicyProposal{}, "issuancepolicy/UpsertPolicyProposal", nil)
	cdc.RegisterConcrete(&SetActivePolicyProposal{}, "issuancepolicy/SetActivePolicyProposal", nil)
	cdc.RegisterConcrete(&UpdateParamsProposal{}, "issuancepolicy/UpdateParamsProposal", nil)
	legacy.RegisterAminoMsg(cdc, &MsgUpsertPolicy{}, "ip/upsert")
	legacy.RegisterAminoMsg(cdc, &MsgSetActivePolicy{}, "ip/activate")
	legacy.RegisterAminoMsg(cdc, &MsgPausePolicy{}, "ip/pause")
	legacy.RegisterAminoMsg(cdc, &MsgResumePolicy{}, "ip/resume")
	legacy.RegisterAminoMsg(cdc, &MsgDeprecatePolicy{}, "ip/deprecate")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "ip/params")
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {}
