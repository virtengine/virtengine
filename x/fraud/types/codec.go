// Package types contains types for the Fraud module.
//
// VE-912: Fraud reporting flow - Codec registration
package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// ModuleCdc is the codec for the module
var ModuleCdc = codec.NewProtoCodec(types.NewInterfaceRegistry())

// RegisterLegacyAminoCodec registers the necessary interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSubmitFraudReport{}, "fraud/MsgSubmitFraudReport", nil)
	cdc.RegisterConcrete(&MsgAssignModerator{}, "fraud/MsgAssignModerator", nil)
	cdc.RegisterConcrete(&MsgUpdateReportStatus{}, "fraud/MsgUpdateReportStatus", nil)
	cdc.RegisterConcrete(&MsgResolveFraudReport{}, "fraud/MsgResolveFraudReport", nil)
	cdc.RegisterConcrete(&MsgRejectFraudReport{}, "fraud/MsgRejectFraudReport", nil)
	cdc.RegisterConcrete(&MsgEscalateFraudReport{}, "fraud/MsgEscalateFraudReport", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "fraud/MsgUpdateParams", nil)
}

// RegisterInterfaces registers the interfaces for the module
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSubmitFraudReport{},
		&MsgAssignModerator{},
		&MsgUpdateReportStatus{},
		&MsgResolveFraudReport{},
		&MsgRejectFraudReport{},
		&MsgEscalateFraudReport{},
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// _Msg_serviceDesc is the gRPC service descriptor for Msg service
var _Msg_serviceDesc = struct {
	ServiceName string
	HandlerType interface{}
	Methods     []struct {
		MethodName string
		Handler    interface{}
	}
	Streams  []struct{}
	Metadata string
}{
	ServiceName: "virtengine.fraud.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods:     []struct{ MethodName string; Handler interface{} }{},
	Streams:     []struct{}{},
	Metadata:    "virtengine/fraud/v1/tx.proto",
}

// MsgServer is the server API for Msg service
type MsgServer interface {
	SubmitFraudReport(sdk.Context, *MsgSubmitFraudReport) (*MsgSubmitFraudReportResponse, error)
	AssignModerator(sdk.Context, *MsgAssignModerator) (*MsgAssignModeratorResponse, error)
	UpdateReportStatus(sdk.Context, *MsgUpdateReportStatus) (*MsgUpdateReportStatusResponse, error)
	ResolveFraudReport(sdk.Context, *MsgResolveFraudReport) (*MsgResolveFraudReportResponse, error)
	RejectFraudReport(sdk.Context, *MsgRejectFraudReport) (*MsgRejectFraudReportResponse, error)
	EscalateFraudReport(sdk.Context, *MsgEscalateFraudReport) (*MsgEscalateFraudReportResponse, error)
	UpdateParams(sdk.Context, *MsgUpdateParams) (*MsgUpdateParamsResponse, error)
}

// Message response types
type MsgSubmitFraudReportResponse struct {
	ReportID string `json:"report_id"`
}

type MsgAssignModeratorResponse struct{}

type MsgUpdateReportStatusResponse struct{}

type MsgResolveFraudReportResponse struct{}

type MsgRejectFraudReportResponse struct{}

type MsgEscalateFraudReportResponse struct{}

type MsgUpdateParamsResponse struct{}
