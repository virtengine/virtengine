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
	legacy.RegisterAminoMsg(cdc, &MsgResizeAllocation{}, "marketplace/MsgResizeAllocation")
	legacy.RegisterAminoMsg(cdc, &MsgPauseAllocation{}, "marketplace/MsgPauseAllocation")
	legacy.RegisterAminoMsg(cdc, &MsgWaldurCallback{}, "marketplace/MsgWaldurCallback")
	legacy.RegisterAminoMsg(cdc, &MsgCreateOrder{}, "marketplace/MsgCreateOrder")
	legacy.RegisterAminoMsg(cdc, &MsgPlaceBid{}, "marketplace/MsgPlaceBid")
	legacy.RegisterAminoMsg(cdc, &MsgWithdrawBid{}, "marketplace/MsgWithdrawBid")
	legacy.RegisterAminoMsg(cdc, &MsgRegisterWaldurSource{}, "marketplace/MsgRegisterWaldurSource")
	legacy.RegisterAminoMsg(cdc, &MsgIngestWaldurOffering{}, "marketplace/MsgIngestWaldurOffering")
	legacy.RegisterAminoMsg(cdc, &MsgSetOfferingVisibility{}, "marketplace/MsgSetOfferingVisibility")
	legacy.RegisterAminoMsg(cdc, &MsgAckWaldurCommand{}, "marketplace/MsgAckWaldurCommand")
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
		&MsgResizeAllocation{},
		&MsgPauseAllocation{},
		&MsgWaldurCallback{},
		&MsgCreateOrder{},
		&MsgPlaceBid{},
		&MsgWithdrawBid{},
		&MsgRegisterWaldurSource{},
		&MsgIngestWaldurOffering{},
		&MsgSetOfferingVisibility{},
		&MsgAckWaldurCommand{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &marketplacev1.Msg_serviceDesc)
}
