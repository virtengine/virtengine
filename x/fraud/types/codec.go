// Package types contains types for the Fraud module.
//
// VE-912: Fraud reporting flow - Codec registration
// VE-3053: Fixed proto method descriptor panic
package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/gogoproto/grpc"

	fraudv1 "github.com/virtengine/virtengine/sdk/go/node/fraud/v1"
)

var (
	amino = codec.NewLegacyAmino()
	// ModuleCdc is the codec for the module
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterLegacyAminoCodec(amino)
}

// RegisterLegacyAminoCodec registers the necessary interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &fraudv1.MsgSubmitFraudReport{}, "fraud/MsgSubmitFraudReport")
	legacy.RegisterAminoMsg(cdc, &fraudv1.MsgAssignModerator{}, "fraud/MsgAssignModerator")
	legacy.RegisterAminoMsg(cdc, &fraudv1.MsgUpdateReportStatus{}, "fraud/MsgUpdateReportStatus")
	legacy.RegisterAminoMsg(cdc, &fraudv1.MsgResolveFraudReport{}, "fraud/MsgResolveFraudReport")
	legacy.RegisterAminoMsg(cdc, &fraudv1.MsgRejectFraudReport{}, "fraud/MsgRejectFraudReport")
	legacy.RegisterAminoMsg(cdc, &fraudv1.MsgEscalateFraudReport{}, "fraud/MsgEscalateFraudReport")
	legacy.RegisterAminoMsg(cdc, &fraudv1.MsgUpdateParams{}, "fraud/MsgUpdateParams")
}

// RegisterInterfaces registers the interfaces for the module
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&fraudv1.MsgSubmitFraudReport{},
		&fraudv1.MsgAssignModerator{},
		&fraudv1.MsgUpdateReportStatus{},
		&fraudv1.MsgResolveFraudReport{},
		&fraudv1.MsgRejectFraudReport{},
		&fraudv1.MsgEscalateFraudReport{},
		&fraudv1.MsgUpdateParams{},
	)

	// Register the Msg service descriptor for proper gRPC routing
	msgservice.RegisterMsgServiceDesc(registry, &fraudv1.Msg_serviceDesc)
}

// RegisterMsgServer registers the MsgServer implementation with the grpc.Server.
// This delegates to the generated proto registration function.
func RegisterMsgServer(s grpc.Server, srv MsgServerImpl) {
	fraudv1.RegisterMsgServer(s, NewMsgServerAdapter(srv))
}

// RegisterQueryServer registers the QueryServer implementation with the grpc.Server.
// This delegates to the generated proto registration function.
func RegisterQueryServer(s grpc.Server, srv QueryServerImpl) {
	fraudv1.RegisterQueryServer(s, NewQueryServerAdapter(srv))
}
