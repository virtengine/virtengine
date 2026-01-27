// Package types contains types for the staking module.
//
// VE-921: Staking rewards codec
package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
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
	// NOTE: These are stub message types without proper protobuf generation.
	// They don't have proper typeURLs (XXX_MessageName() methods), so we cannot
	// register them with RegisterImplementations. This will cause typeURL "/" conflicts.
	//
	// Once proper .proto files are generated with protoc-gen-gogo, this should be:
	//
	// registry.RegisterImplementations((*sdk.Msg)(nil),
	//     &MsgUpdateParams{},
	//     &MsgSlashValidator{},
	//     &MsgUnjailValidator{},
	//     &MsgRecordPerformance{},
	// )
	//
	// msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
	_ = registry // suppress unused variable warning
}

// _Msg_serviceDesc is a placeholder for the message service descriptor
// NOTE: This is a stub - proper proto generation will create the real ServiceDesc
var _Msg_serviceDesc = struct {
	ServiceName string
}{
	ServiceName: "virtengine.staking.v1.Msg",
}

// Ensure msgservice import is used
var _ = msgservice.RegisterMsgServiceDesc
