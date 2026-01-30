// Package types contains types for the Review module.
//
// VE-911: Provider public reviews - Codec registration
package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	reviewv1 "github.com/virtengine/virtengine/sdk/go/node/review/v1"
)

// ModuleCdc is the codec for the module
var ModuleCdc = codec.NewLegacyAmino()

func init() {
	RegisterLegacyAminoCodec(ModuleCdc)
}

// RegisterLegacyAminoCodec registers the necessary interfaces and concrete types
// on the provided LegacyAmino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgSubmitReview{}, "virtengine/review/MsgSubmitReview")
	legacy.RegisterAminoMsg(cdc, &MsgDeleteReview{}, "virtengine/review/MsgDeleteReview")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "virtengine/review/MsgUpdateParams")
}

// RegisterInterfaces registers the module interface types
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSubmitReview{},
		&MsgDeleteReview{},
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &reviewv1.Msg_serviceDesc)
}
