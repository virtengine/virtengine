package types

import (
	"context"

	"github.com/cosmos/gogoproto/grpc"
)

type MsgServer interface {
	UpsertPolicy(context.Context, *MsgUpsertPolicy) (*MsgUpsertPolicyResponse, error)
	SetActivePolicy(context.Context, *MsgSetActivePolicy) (*MsgSetActivePolicyResponse, error)
	PausePolicy(context.Context, *MsgPausePolicy) (*MsgPausePolicyResponse, error)
	ResumePolicy(context.Context, *MsgResumePolicy) (*MsgResumePolicyResponse, error)
	DeprecatePolicy(context.Context, *MsgDeprecatePolicy) (*MsgDeprecatePolicyResponse, error)
	UpdateParams(context.Context, *MsgUpdateParams) (*MsgUpdateParamsResponse, error)
}

func RegisterMsgServer(s grpc.Server, srv MsgServer) {
	_ = s
	_ = srv
}
