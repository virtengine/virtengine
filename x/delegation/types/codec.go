// Package types contains types for the delegation module.
//
// VE-922: Delegation module codec
package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	delegationv1 "github.com/virtengine/virtengine/sdk/go/node/delegation/v1"
)

// Type alias for MsgServer from generated proto
type MsgServer = delegationv1.MsgServer

// RegisterMsgServer registers the MsgServer implementation with a gRPC server
var RegisterMsgServer = delegationv1.RegisterMsgServer

// ModuleCdc is the codec for the module
var ModuleCdc = codec.NewLegacyAmino()

func init() {
	RegisterLegacyAminoCodec(ModuleCdc)
}

// RegisterLegacyAminoCodec registers the delegation types on the provided LegacyAmino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register messages
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "virtengine/delegation/MsgUpdateParams")
	legacy.RegisterAminoMsg(cdc, &MsgDelegate{}, "virtengine/delegation/MsgDelegate")
	legacy.RegisterAminoMsg(cdc, &MsgUndelegate{}, "virtengine/delegation/MsgUndelegate")
	legacy.RegisterAminoMsg(cdc, &MsgRedelegate{}, "virtengine/delegation/MsgRedelegate")
	legacy.RegisterAminoMsg(cdc, &MsgClaimRewards{}, "virtengine/delegation/MsgClaimRewards")
	legacy.RegisterAminoMsg(cdc, &MsgClaimAllRewards{}, "virtengine/delegation/MsgClaimAllRewards")
}

// RegisterInterfaces registers the delegation types and interfaces
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateParams{},
		&MsgDelegate{},
		&MsgUndelegate{},
		&MsgRedelegate{},
		&MsgClaimRewards{},
		&MsgClaimAllRewards{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &delegationv1.Msg_serviceDesc)
}
