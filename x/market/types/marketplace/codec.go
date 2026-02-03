package marketplace

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	marketplacev1 "github.com/virtengine/virtengine/sdk/go/node/marketplace/v1"
)

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterLegacyAminoCodec(amino)
}

// RegisterLegacyAminoCodec registers amino types.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgCreateOffering{}, "marketplace/MsgCreateOffering")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateOffering{}, "marketplace/MsgUpdateOffering")
	legacy.RegisterAminoMsg(cdc, &MsgDeactivateOffering{}, "marketplace/MsgDeactivateOffering")
	legacy.RegisterAminoMsg(cdc, &MsgAcceptBid{}, "marketplace/MsgAcceptBid")
	legacy.RegisterAminoMsg(cdc, &MsgTerminateAllocation{}, "marketplace/MsgTerminateAllocation")
	legacy.RegisterAminoMsg(cdc, &MsgWaldurCallback{}, "marketplace/MsgWaldurCallback")
}

// RegisterInterfaces registers module interfaces.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateOffering{},
		&MsgUpdateOffering{},
		&MsgDeactivateOffering{},
		&MsgAcceptBid{},
		&MsgTerminateAllocation{},
		&MsgWaldurCallback{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &marketplacev1.Msg_serviceDesc)
}
