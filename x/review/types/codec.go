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

	// TODO: Enable when protobuf generation is complete
	// msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
	_ = msgservice.RegisterMsgServiceDesc
}

// _Msg_serviceDesc is the gRPC service descriptor for the Msg service
var _Msg_serviceDesc = struct {
	ServiceName string
	HandlerType interface{}
	Methods     []struct {
		MethodName string
		Handler    interface{}
	}
	Streams  []interface{}
	Metadata interface{}
}{
	ServiceName: "virtengine.review.v1.Msg",
	HandlerType: nil,
	Methods:     nil,
	Streams:     nil,
	Metadata:    nil,
}
