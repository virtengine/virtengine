// Package types contains types for the Fraud module.
//
// VE-912: Fraud reporting flow - Codec registration
package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	fraudv1 "github.com/virtengine/virtengine/sdk/go/node/fraud/v1"
)

// ModuleCdc is the codec for the module
var ModuleCdc = codec.NewProtoCodec(types.NewInterfaceRegistry())

// RegisterLegacyAminoCodec registers the necessary interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&fraudv1.MsgSubmitFraudReport{}, "fraud/MsgSubmitFraudReport", nil)
	cdc.RegisterConcrete(&fraudv1.MsgAssignModerator{}, "fraud/MsgAssignModerator", nil)
	cdc.RegisterConcrete(&fraudv1.MsgUpdateReportStatus{}, "fraud/MsgUpdateReportStatus", nil)
	cdc.RegisterConcrete(&fraudv1.MsgResolveFraudReport{}, "fraud/MsgResolveFraudReport", nil)
	cdc.RegisterConcrete(&fraudv1.MsgRejectFraudReport{}, "fraud/MsgRejectFraudReport", nil)
	cdc.RegisterConcrete(&fraudv1.MsgEscalateFraudReport{}, "fraud/MsgEscalateFraudReport", nil)
	cdc.RegisterConcrete(&fraudv1.MsgUpdateParams{}, "fraud/MsgUpdateParams", nil)
}

// RegisterInterfaces registers the interfaces for the module
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&fraudv1.MsgSubmitFraudReport{},
		&fraudv1.MsgAssignModerator{},
		&fraudv1.MsgUpdateReportStatus{},
		&fraudv1.MsgResolveFraudReport{},
		&fraudv1.MsgRejectFraudReport{},
		&fraudv1.MsgEscalateFraudReport{},
		&fraudv1.MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &fraudv1.Msg_serviceDesc)
}
