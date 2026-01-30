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

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterLegacyAminoCodec(amino)
}

// RegisterLegacyAminoCodec registers the delegation types on the provided LegacyAmino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register messages using generated proto types
	legacy.RegisterAminoMsg(cdc, &delegationv1.MsgDelegate{}, "delegation/MsgDelegate")
	legacy.RegisterAminoMsg(cdc, &delegationv1.MsgUndelegate{}, "delegation/MsgUndelegate")
	legacy.RegisterAminoMsg(cdc, &delegationv1.MsgRedelegate{}, "delegation/MsgRedelegate")
	legacy.RegisterAminoMsg(cdc, &delegationv1.MsgClaimRewards{}, "delegation/MsgClaimRewards")
	legacy.RegisterAminoMsg(cdc, &delegationv1.MsgClaimAllRewards{}, "delegation/MsgClaimAllRewards")
	legacy.RegisterAminoMsg(cdc, &delegationv1.MsgUpdateParams{}, "delegation/MsgUpdateParams")
}

// RegisterInterfaces registers the delegation types and interfaces
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&delegationv1.MsgDelegate{},
		&delegationv1.MsgUndelegate{},
		&delegationv1.MsgRedelegate{},
		&delegationv1.MsgClaimRewards{},
		&delegationv1.MsgClaimAllRewards{},
		&delegationv1.MsgUpdateParams{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &delegationv1.Msg_serviceDesc)
}
