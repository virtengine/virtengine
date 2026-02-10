package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
)

// MsgServer type alias
type MsgServer = resourcesv1.MsgServer

// RegisterMsgServer function alias
var RegisterMsgServer = resourcesv1.RegisterMsgServer

// ModuleCdc is the codec for the module
var ModuleCdc = codec.NewLegacyAmino()

func init() {
	RegisterLegacyAminoCodec(ModuleCdc)
}

// RegisterLegacyAminoCodec registers the resources module types.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgProviderHeartbeat{}, "resources/MsgProviderHeartbeat")
	legacy.RegisterAminoMsg(cdc, &MsgAllocateResources{}, "resources/MsgAllocateResources")
	legacy.RegisterAminoMsg(cdc, &MsgActivateAllocation{}, "resources/MsgActivateAllocation")
	legacy.RegisterAminoMsg(cdc, &MsgReleaseAllocation{}, "resources/MsgReleaseAllocation")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "resources/MsgUpdateParams")
}

// RegisterInterfaces registers interfaces for the resources module.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgProviderHeartbeat{},
		&MsgAllocateResources{},
		&MsgActivateAllocation{},
		&MsgReleaseAllocation{},
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &resourcesv1.Msg_serviceDesc)
}
