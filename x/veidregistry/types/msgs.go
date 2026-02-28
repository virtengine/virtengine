package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	TypeMsgUpsertVerifierVersion    = "upsert_verifier_version"
	TypeMsgApproveVerifierVersion   = "approve_verifier_version"
	TypeMsgCancelVerifierVersion    = "cancel_verifier_version"
	TypeMsgRetireVerifierVersion    = "retire_verifier_version"
	TypeMsgReportValidatorReadiness = "report_validator_readiness"
	TypeMsgUpdateVerifierRegParams  = "update_params"
)

var (
	_ sdk.Msg = &MsgUpsertVerifierVersion{}
	_ sdk.Msg = &MsgApproveVerifierVersion{}
	_ sdk.Msg = &MsgCancelVerifierVersion{}
	_ sdk.Msg = &MsgRetireVerifierVersion{}
	_ sdk.Msg = &MsgReportValidatorReadiness{}
	_ sdk.Msg = &MsgUpdateParams{}
)

type MsgUpsertVerifierVersion struct {
	Authority string          `json:"authority"`
	Verifier  VerifierVersion `json:"verifier"`
	Changelog string          `json:"changelog,omitempty"`
}

func NewMsgUpsertVerifierVersion(authority string, verifier VerifierVersion, changelog string) *MsgUpsertVerifierVersion {
	return &MsgUpsertVerifierVersion{Authority: authority, Verifier: verifier, Changelog: changelog}
}

func (msg MsgUpsertVerifierVersion) Route() string { return RouterKey }
func (msg MsgUpsertVerifierVersion) Type() string  { return TypeMsgUpsertVerifierVersion }
func (msg MsgUpsertVerifierVersion) ValidateBasic() error {
	if strings.TrimSpace(msg.Authority) == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	if err := msg.Verifier.Validate(); err != nil {
		return err
	}
	switch VerifierStatus(msg.Verifier.Status) {
	case VerifierStatusProposed:
		return nil
	case "":
		return nil
	default:
		return fmt.Errorf("upsert only supports proposed verifier state")
	}
}
func (msg MsgUpsertVerifierVersion) GetSigners() []sdk.AccAddress { return nil }
func (msg MsgUpsertVerifierVersion) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

type MsgApproveVerifierVersion struct {
	Authority            string `json:"authority"`
	VerifierID           string `json:"verifier_id"`
	GovernanceProposalID uint64 `json:"governance_proposal_id"`
	ActivationHeight     int64  `json:"activation_height"`
	SecurityFix          bool   `json:"security_fix"`
}

func NewMsgApproveVerifierVersion(authority, verifierID string, proposalID uint64, activationHeight int64, securityFix bool) *MsgApproveVerifierVersion {
	return &MsgApproveVerifierVersion{
		Authority:            authority,
		VerifierID:           verifierID,
		GovernanceProposalID: proposalID,
		ActivationHeight:     activationHeight,
		SecurityFix:          securityFix,
	}
}

func (msg MsgApproveVerifierVersion) Route() string { return RouterKey }
func (msg MsgApproveVerifierVersion) Type() string  { return TypeMsgApproveVerifierVersion }
func (msg MsgApproveVerifierVersion) ValidateBasic() error {
	if strings.TrimSpace(msg.Authority) == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	if strings.TrimSpace(msg.VerifierID) == "" {
		return fmt.Errorf("verifier_id cannot be empty")
	}
	if msg.GovernanceProposalID == 0 {
		return fmt.Errorf("governance_proposal_id cannot be zero")
	}
	if msg.ActivationHeight < 0 {
		return fmt.Errorf("activation_height cannot be negative")
	}
	return nil
}
func (msg MsgApproveVerifierVersion) GetSigners() []sdk.AccAddress { return nil }
func (msg MsgApproveVerifierVersion) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

type MsgCancelVerifierVersion struct {
	Authority  string `json:"authority"`
	VerifierID string `json:"verifier_id"`
	Reason     string `json:"reason,omitempty"`
}

func NewMsgCancelVerifierVersion(authority, verifierID, reason string) *MsgCancelVerifierVersion {
	return &MsgCancelVerifierVersion{Authority: authority, VerifierID: verifierID, Reason: reason}
}

func (msg MsgCancelVerifierVersion) Route() string { return RouterKey }
func (msg MsgCancelVerifierVersion) Type() string  { return TypeMsgCancelVerifierVersion }
func (msg MsgCancelVerifierVersion) ValidateBasic() error {
	if strings.TrimSpace(msg.Authority) == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	if strings.TrimSpace(msg.VerifierID) == "" {
		return fmt.Errorf("verifier_id cannot be empty")
	}
	return nil
}
func (msg MsgCancelVerifierVersion) GetSigners() []sdk.AccAddress { return nil }
func (msg MsgCancelVerifierVersion) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

type MsgRetireVerifierVersion struct {
	Authority  string `json:"authority"`
	VerifierID string `json:"verifier_id"`
	Reason     string `json:"reason,omitempty"`
}

func NewMsgRetireVerifierVersion(authority, verifierID, reason string) *MsgRetireVerifierVersion {
	return &MsgRetireVerifierVersion{Authority: authority, VerifierID: verifierID, Reason: reason}
}

func (msg MsgRetireVerifierVersion) Route() string { return RouterKey }
func (msg MsgRetireVerifierVersion) Type() string  { return TypeMsgRetireVerifierVersion }
func (msg MsgRetireVerifierVersion) ValidateBasic() error {
	if strings.TrimSpace(msg.Authority) == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	if strings.TrimSpace(msg.VerifierID) == "" {
		return fmt.Errorf("verifier_id cannot be empty")
	}
	return nil
}
func (msg MsgRetireVerifierVersion) GetSigners() []sdk.AccAddress { return nil }
func (msg MsgRetireVerifierVersion) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

type MsgReportValidatorReadiness struct {
	ValidatorAddress  string `json:"validator_address"`
	VerifierID        string `json:"verifier_id"`
	ConformancePassed bool   `json:"conformance_passed"`
	ImplementationID  string `json:"implementation_id,omitempty"`
	Organization      string `json:"organization,omitempty"`
}

func NewMsgReportValidatorReadiness(
	validatorAddress string,
	verifierID string,
	conformancePassed bool,
	implementationID string,
	organization string,
) *MsgReportValidatorReadiness {
	return &MsgReportValidatorReadiness{
		ValidatorAddress:  validatorAddress,
		VerifierID:        verifierID,
		ConformancePassed: conformancePassed,
		ImplementationID:  implementationID,
		Organization:      organization,
	}
}

func (msg MsgReportValidatorReadiness) Route() string { return RouterKey }
func (msg MsgReportValidatorReadiness) Type() string  { return TypeMsgReportValidatorReadiness }
func (msg MsgReportValidatorReadiness) ValidateBasic() error {
	if strings.TrimSpace(msg.ValidatorAddress) == "" {
		return fmt.Errorf("validator_address cannot be empty")
	}
	if strings.TrimSpace(msg.VerifierID) == "" {
		return fmt.Errorf("verifier_id cannot be empty")
	}
	return nil
}
func (msg MsgReportValidatorReadiness) GetSigners() []sdk.AccAddress { return nil }
func (msg MsgReportValidatorReadiness) GetSignBytes() []byte {
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
func (msg MsgUpdateParams) Type() string  { return TypeMsgUpdateVerifierRegParams }
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

type MsgUpsertVerifierVersionResponse struct{}
type MsgApproveVerifierVersionResponse struct{}
type MsgCancelVerifierVersionResponse struct{}
type MsgRetireVerifierVersionResponse struct{}
type MsgReportValidatorReadinessResponse struct{}
type MsgUpdateParamsResponse struct{}

func (*MsgUpsertVerifierVersion) ProtoMessage()    {}
func (m *MsgUpsertVerifierVersion) Reset()         { *m = MsgUpsertVerifierVersion{} }
func (m *MsgUpsertVerifierVersion) String() string { return fmt.Sprintf("%+v", *m) }

func (*MsgApproveVerifierVersion) ProtoMessage()    {}
func (m *MsgApproveVerifierVersion) Reset()         { *m = MsgApproveVerifierVersion{} }
func (m *MsgApproveVerifierVersion) String() string { return fmt.Sprintf("%+v", *m) }

func (*MsgCancelVerifierVersion) ProtoMessage()    {}
func (m *MsgCancelVerifierVersion) Reset()         { *m = MsgCancelVerifierVersion{} }
func (m *MsgCancelVerifierVersion) String() string { return fmt.Sprintf("%+v", *m) }

func (*MsgRetireVerifierVersion) ProtoMessage()    {}
func (m *MsgRetireVerifierVersion) Reset()         { *m = MsgRetireVerifierVersion{} }
func (m *MsgRetireVerifierVersion) String() string { return fmt.Sprintf("%+v", *m) }

func (*MsgReportValidatorReadiness) ProtoMessage()    {}
func (m *MsgReportValidatorReadiness) Reset()         { *m = MsgReportValidatorReadiness{} }
func (m *MsgReportValidatorReadiness) String() string { return fmt.Sprintf("%+v", *m) }

func (*MsgUpdateParams) ProtoMessage()    {}
func (m *MsgUpdateParams) Reset()         { *m = MsgUpdateParams{} }
func (m *MsgUpdateParams) String() string { return fmt.Sprintf("%+v", *m) }
