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
	legacy.RegisterAminoMsg(cdc, &MsgCreateTicket{}, "support/MsgCreateTicket")
	legacy.RegisterAminoMsg(cdc, &MsgAssignTicket{}, "support/MsgAssignTicket")
	legacy.RegisterAminoMsg(cdc, &MsgRespondToTicket{}, "support/MsgRespondToTicket")
	legacy.RegisterAminoMsg(cdc, &MsgResolveTicket{}, "support/MsgResolveTicket")
	legacy.RegisterAminoMsg(cdc, &MsgCloseTicket{}, "support/MsgCloseTicket")
	legacy.RegisterAminoMsg(cdc, &MsgReopenTicket{}, "support/MsgReopenTicket")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "support/MsgUpdateParams")
}

// RegisterInterfaces registers the interfaces types with the interface registry.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateTicket{},
		&MsgAssignTicket{},
		&MsgRespondToTicket{},
		&MsgResolveTicket{},
		&MsgCloseTicket{},
		&MsgReopenTicket{},
		&MsgUpdateParams{},
	)
	// Note: In a full implementation, msgservice.RegisterMsgServiceDesc would be called
	// with the protobuf-generated service descriptor. For now, we register messages
	// individually above which is sufficient for Cosmos SDK module operation.
}
