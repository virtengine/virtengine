package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	TypeMsgUpsertPolicy         = "upsert_policy"
	TypeMsgSetActivePolicy      = "set_active_policy"
	TypeMsgPauseIssuancePolicy  = "pause_policy"
	TypeMsgResumeIssuancePolicy = "resume_policy"
	TypeMsgDeprecatePolicy      = "deprecate_policy"
	TypeMsgUpdateIssuanceParams = "update_params"
)

var (
	_ sdk.Msg = &MsgUpsertPolicy{}
	_ sdk.Msg = &MsgSetActivePolicy{}
	_ sdk.Msg = &MsgPausePolicy{}
	_ sdk.Msg = &MsgResumePolicy{}
	_ sdk.Msg = &MsgDeprecatePolicy{}
	_ sdk.Msg = &MsgUpdateParams{}
)

type MsgUpsertPolicy struct {
	Authority string         `json:"authority"`
	Policy    IssuancePolicy `json:"policy"`
}

func NewMsgUpsertPolicy(authority string, policy IssuancePolicy) *MsgUpsertPolicy {
	return &MsgUpsertPolicy{Authority: authority, Policy: policy}
}

func (msg MsgUpsertPolicy) Route() string { return RouterKey }
func (msg MsgUpsertPolicy) Type() string  { return TypeMsgUpsertPolicy }
func (msg MsgUpsertPolicy) ValidateBasic() error {
	if strings.TrimSpace(msg.Authority) == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	return msg.Policy.Validate()
}
func (msg MsgUpsertPolicy) GetSigners() []sdk.AccAddress { return nil }
func (msg MsgUpsertPolicy) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

type MsgSetActivePolicy struct {
	Authority string `json:"authority"`
	PolicyID  string `json:"policy_id"`
}

func NewMsgSetActivePolicy(authority, policyID string) *MsgSetActivePolicy {
	return &MsgSetActivePolicy{Authority: authority, PolicyID: policyID}
}

func (msg MsgSetActivePolicy) Route() string { return RouterKey }
func (msg MsgSetActivePolicy) Type() string  { return TypeMsgSetActivePolicy }
func (msg MsgSetActivePolicy) ValidateBasic() error {
	if strings.TrimSpace(msg.Authority) == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	if strings.TrimSpace(msg.PolicyID) == "" {
		return fmt.Errorf("policy_id cannot be empty")
	}
	return nil
}
func (msg MsgSetActivePolicy) GetSigners() []sdk.AccAddress { return nil }
func (msg MsgSetActivePolicy) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

type MsgPausePolicy struct {
	Authority string `json:"authority"`
	PolicyID  string `json:"policy_id,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

func NewMsgPausePolicy(authority, policyID, reason string) *MsgPausePolicy {
	return &MsgPausePolicy{Authority: authority, PolicyID: policyID, Reason: reason}
}

func (msg MsgPausePolicy) Route() string { return RouterKey }
func (msg MsgPausePolicy) Type() string  { return TypeMsgPauseIssuancePolicy }
func (msg MsgPausePolicy) ValidateBasic() error {
	if strings.TrimSpace(msg.Authority) == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	return nil
}
func (msg MsgPausePolicy) GetSigners() []sdk.AccAddress { return nil }
func (msg MsgPausePolicy) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

type MsgResumePolicy struct {
	Authority string `json:"authority"`
	PolicyID  string `json:"policy_id,omitempty"`
}

func NewMsgResumePolicy(authority, policyID string) *MsgResumePolicy {
	return &MsgResumePolicy{Authority: authority, PolicyID: policyID}
}

func (msg MsgResumePolicy) Route() string { return RouterKey }
func (msg MsgResumePolicy) Type() string  { return TypeMsgResumeIssuancePolicy }
func (msg MsgResumePolicy) ValidateBasic() error {
	if strings.TrimSpace(msg.Authority) == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	return nil
}
func (msg MsgResumePolicy) GetSigners() []sdk.AccAddress { return nil }
func (msg MsgResumePolicy) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

type MsgDeprecatePolicy struct {
	Authority string `json:"authority"`
	PolicyID  string `json:"policy_id"`
}

func NewMsgDeprecatePolicy(authority, policyID string) *MsgDeprecatePolicy {
	return &MsgDeprecatePolicy{Authority: authority, PolicyID: policyID}
}

func (msg MsgDeprecatePolicy) Route() string { return RouterKey }
func (msg MsgDeprecatePolicy) Type() string  { return TypeMsgDeprecatePolicy }
func (msg MsgDeprecatePolicy) ValidateBasic() error {
	if strings.TrimSpace(msg.Authority) == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	if strings.TrimSpace(msg.PolicyID) == "" {
		return fmt.Errorf("policy_id cannot be empty")
	}
	return nil
}
func (msg MsgDeprecatePolicy) GetSigners() []sdk.AccAddress { return nil }
func (msg MsgDeprecatePolicy) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

type MsgUpdateParams struct {
	Authority string `json:"authority"`
	Params    Params `json:"params"`
}

func NewMsgUpdateParams(authority string, params Params) *MsgUpdateParams {
	return &MsgUpdateParams{Authority: authority, Params: params}
}

func (msg MsgUpdateParams) Route() string { return RouterKey }
func (msg MsgUpdateParams) Type() string  { return TypeMsgUpdateIssuanceParams }
func (msg MsgUpdateParams) ValidateBasic() error {
	if strings.TrimSpace(msg.Authority) == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	return msg.Params.Validate()
}
func (msg MsgUpdateParams) GetSigners() []sdk.AccAddress { return nil }
func (msg MsgUpdateParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

type MsgUpsertPolicyResponse struct{}
type MsgSetActivePolicyResponse struct{}
type MsgPausePolicyResponse struct{}
type MsgResumePolicyResponse struct{}
type MsgDeprecatePolicyResponse struct{}
type MsgUpdateParamsResponse struct{}

func (*MsgUpsertPolicy) ProtoMessage()    {}
func (m *MsgUpsertPolicy) Reset()         { *m = MsgUpsertPolicy{} }
func (m *MsgUpsertPolicy) String() string { return fmt.Sprintf("%+v", *m) }

func (*MsgSetActivePolicy) ProtoMessage()    {}
func (m *MsgSetActivePolicy) Reset()         { *m = MsgSetActivePolicy{} }
func (m *MsgSetActivePolicy) String() string { return fmt.Sprintf("%+v", *m) }

func (*MsgPausePolicy) ProtoMessage()    {}
func (m *MsgPausePolicy) Reset()         { *m = MsgPausePolicy{} }
func (m *MsgPausePolicy) String() string { return fmt.Sprintf("%+v", *m) }

func (*MsgResumePolicy) ProtoMessage()    {}
func (m *MsgResumePolicy) Reset()         { *m = MsgResumePolicy{} }
func (m *MsgResumePolicy) String() string { return fmt.Sprintf("%+v", *m) }

func (*MsgDeprecatePolicy) ProtoMessage()    {}
func (m *MsgDeprecatePolicy) Reset()         { *m = MsgDeprecatePolicy{} }
func (m *MsgDeprecatePolicy) String() string { return fmt.Sprintf("%+v", *m) }

func (*MsgUpdateParams) ProtoMessage()    {}
func (m *MsgUpdateParams) Reset()         { *m = MsgUpdateParams{} }
func (m *MsgUpdateParams) String() string { return fmt.Sprintf("%+v", *m) }
