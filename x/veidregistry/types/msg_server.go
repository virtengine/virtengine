package types

import (
	"context"

	"github.com/cosmos/gogoproto/grpc"
)

type MsgServer interface {
	UpsertVerifierVersion(context.Context, *MsgUpsertVerifierVersion) (*MsgUpsertVerifierVersionResponse, error)
	ApproveVerifierVersion(context.Context, *MsgApproveVerifierVersion) (*MsgApproveVerifierVersionResponse, error)
	CancelVerifierVersion(context.Context, *MsgCancelVerifierVersion) (*MsgCancelVerifierVersionResponse, error)
	RetireVerifierVersion(context.Context, *MsgRetireVerifierVersion) (*MsgRetireVerifierVersionResponse, error)
	ReportValidatorReadiness(context.Context, *MsgReportValidatorReadiness) (*MsgReportValidatorReadinessResponse, error)
	UpdateParams(context.Context, *MsgUpdateParams) (*MsgUpdateParamsResponse, error)
}

func RegisterMsgServer(s grpc.Server, srv MsgServer) {
	_ = s
	_ = srv
}
