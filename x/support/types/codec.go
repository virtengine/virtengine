package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	supportv1 "github.com/virtengine/virtengine/sdk/go/node/support/v1"
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
	legacy.RegisterAminoMsg(cdc, &supportv1.MsgRegisterExternalTicket{}, "support/MsgRegisterExternalTicket")
	legacy.RegisterAminoMsg(cdc, &supportv1.MsgUpdateExternalTicket{}, "support/MsgUpdateExternalTicket")
	legacy.RegisterAminoMsg(cdc, &supportv1.MsgRemoveExternalTicket{}, "support/MsgRemoveExternalTicket")
	legacy.RegisterAminoMsg(cdc, &supportv1.MsgUpdateParams{}, "support/MsgUpdateParams")
	legacy.RegisterAminoMsg(cdc, &supportv1.MsgCreateSupportRequest{}, "support/MsgCreateSupportRequest")
	legacy.RegisterAminoMsg(cdc, &supportv1.MsgUpdateSupportRequest{}, "support/MsgUpdateSupportRequest")
	legacy.RegisterAminoMsg(cdc, &supportv1.MsgAddSupportResponse{}, "support/MsgAddSupportResponse")
	legacy.RegisterAminoMsg(cdc, &supportv1.MsgArchiveSupportRequest{}, "support/MsgArchiveSupportRequest")
}

// RegisterInterfaces registers the interfaces types with the interface registry.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&supportv1.MsgRegisterExternalTicket{},
		&supportv1.MsgUpdateExternalTicket{},
		&supportv1.MsgRemoveExternalTicket{},
		&supportv1.MsgUpdateParams{},
		&supportv1.MsgCreateSupportRequest{},
		&supportv1.MsgUpdateSupportRequest{},
		&supportv1.MsgAddSupportResponse{},
		&supportv1.MsgArchiveSupportRequest{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &supportv1.Msg_serviceDesc)
}
