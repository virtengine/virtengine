// Package types contains types for the delegation module.
//
// VE-922: Delegation module codec
package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the delegation types on the provided LegacyAmino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register messages
	cdc.RegisterConcrete(&MsgUpdateParams{}, "virt_delegation/UpdateParams", nil)
	cdc.RegisterConcrete(&MsgDelegate{}, "virt_delegation/Delegate", nil)
	cdc.RegisterConcrete(&MsgUndelegate{}, "virt_delegation/Undelegate", nil)
	cdc.RegisterConcrete(&MsgRedelegate{}, "virt_delegation/Redelegate", nil)
	cdc.RegisterConcrete(&MsgClaimRewards{}, "virt_delegation/ClaimRewards", nil)
	cdc.RegisterConcrete(&MsgClaimAllRewards{}, "virt_delegation/ClaimAllRewards", nil)
}

// RegisterInterfaces registers the delegation types and interfaces
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// NOTE: These are stub message types without proper protobuf generation.
	// They don't have proper typeURLs (XXX_MessageName() methods), so we cannot
	// register them with RegisterImplementations. This will cause typeURL "/" conflicts.
	//
	// Once proper .proto files are generated with protoc-gen-gogo, this should be:
	//
	// registry.RegisterImplementations((*sdk.Msg)(nil),
	//     &MsgUpdateParams{},
	//     &MsgDelegate{},
	//     &MsgUndelegate{},
	//     &MsgRedelegate{},
	//     &MsgClaimRewards{},
	//     &MsgClaimAllRewards{},
	// )
	//
	// msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
	_ = registry // suppress unused variable warning
	_ = msgservice.RegisterMsgServiceDesc
}

// _Msg_serviceDesc is a placeholder for the message service descriptor
var _Msg_serviceDesc = struct {
	ServiceName string
}{
	ServiceName: "virtengine.delegation.v1.Msg",
}
