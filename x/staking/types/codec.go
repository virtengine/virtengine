// Package types contains types for the staking module.
//
// VE-921: Staking rewards codec
package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	stakingv1 "github.com/virtengine/virtengine/sdk/go/node/staking/v1"
)

// RegisterLegacyAminoCodec registers the staking types on the provided LegacyAmino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &stakingv1.MsgUpdateParams{}, "virt_staking/UpdateParams")
	legacy.RegisterAminoMsg(cdc, &stakingv1.MsgSlashValidator{}, "virt_staking/SlashValidator")
	legacy.RegisterAminoMsg(cdc, &stakingv1.MsgUnjailValidator{}, "virt_staking/UnjailValidator")
	legacy.RegisterAminoMsg(cdc, &stakingv1.MsgRecordPerformance{}, "virt_staking/RecordPerformance")
}

// RegisterInterfaces registers the staking types and interfaces
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&stakingv1.MsgUpdateParams{},
		&stakingv1.MsgSlashValidator{},
		&stakingv1.MsgUnjailValidator{},
		&stakingv1.MsgRecordPerformance{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &stakingv1.Msg_serviceDesc)
}
