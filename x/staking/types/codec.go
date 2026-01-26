// Package types contains types for the staking module.
//
// VE-921: Staking rewards codec
package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the staking types on the provided LegacyAmino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register messages
	cdc.RegisterConcrete(&MsgUpdateParams{}, "virt_staking/UpdateParams", nil)
	cdc.RegisterConcrete(&MsgSlashValidator{}, "virt_staking/SlashValidator", nil)
	cdc.RegisterConcrete(&MsgUnjailValidator{}, "virt_staking/UnjailValidator", nil)
	cdc.RegisterConcrete(&MsgRecordPerformance{}, "virt_staking/RecordPerformance", nil)
}

// RegisterInterfaces registers the staking types and interfaces
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateParams{},
		&MsgSlashValidator{},
		&MsgUnjailValidator{},
		&MsgRecordPerformance{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// _Msg_serviceDesc is a placeholder for the message service descriptor
var _Msg_serviceDesc = struct {
	ServiceName string
}{
	ServiceName: "virtengine.staking.v1.Msg",
}
