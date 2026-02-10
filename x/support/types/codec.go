package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterLegacyAminoCodec(amino)
}

// RegisterLegacyAminoCodec registers the necessary interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgRegisterExternalTicket{}, "support/MsgRegisterExternalTicket")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateExternalTicket{}, "support/MsgUpdateExternalTicket")
	legacy.RegisterAminoMsg(cdc, &MsgRemoveExternalTicket{}, "support/MsgRemoveExternalTicket")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "support/MsgUpdateParams")
	legacy.RegisterAminoMsg(cdc, &MsgCreateSupportRequest{}, "support/MsgCreateSupportRequest")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateSupportRequest{}, "support/MsgUpdateSupportRequest")
	legacy.RegisterAminoMsg(cdc, &MsgAddSupportResponse{}, "support/MsgAddSupportResponse")
	legacy.RegisterAminoMsg(cdc, &MsgArchiveSupportRequest{}, "support/MsgArchiveSupportRequest")
}

// RegisterInterfaces registers the interfaces types with the interface registry.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterExternalTicket{},
		&MsgUpdateExternalTicket{},
		&MsgRemoveExternalTicket{},
		&MsgUpdateParams{},
		&MsgCreateSupportRequest{},
		&MsgUpdateSupportRequest{},
		&MsgAddSupportResponse{},
		&MsgArchiveSupportRequest{},
	)
}
