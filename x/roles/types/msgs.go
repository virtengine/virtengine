package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	mfatypes "github.com/virtengine/virtengine/x/mfa/types"
)

// Error message constants
const (
	errMsgInvalidSenderAddress = "invalid sender address"
	errMsgInvalidTargetAddress = "invalid target address"
)

const (
	TypeMsgAssignRole      = "assign_role"
	TypeMsgRevokeRole      = "revoke_role"
	TypeMsgSetAccountState = "set_account_state"
	TypeMsgNominateAdmin   = "nominate_admin"

	TypeMsgImposeSanction        = "impose_sanction"
	TypeMsgConfirmSanction       = "confirm_sanction"
	TypeMsgRevokeSanction        = "revoke_sanction"
	TypeMsgOpenSanctionAppeal    = "open_sanction_appeal"
	TypeMsgResolveSanctionAppeal = "resolve_sanction_appeal"
)

var (
	_ sdk.Msg = &MsgAssignRole{}
	_ sdk.Msg = &MsgRevokeRole{}
	_ sdk.Msg = &MsgSetAccountState{}
	_ sdk.Msg = &MsgNominateAdmin{}
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgImposeSanction{}
	_ sdk.Msg = &MsgConfirmSanction{}
	_ sdk.Msg = &MsgRevokeSanction{}
	_ sdk.Msg = &MsgOpenSanctionAppeal{}
	_ sdk.Msg = &MsgResolveSanctionAppeal{}
)

// MsgAssignRole is the message for assigning a role to an account
type MsgAssignRole struct {
	Sender  string `json:"sender"`
	Address string `json:"address"`
	Role    string `json:"role"`
}

// NewMsgAssignRole creates a new MsgAssignRole
func NewMsgAssignRole(sender, address, role string) *MsgAssignRole {
	return &MsgAssignRole{
		Sender:  sender,
		Address: address,
		Role:    role,
	}
}

// Route returns the route for the message
func (msg MsgAssignRole) Route() string { return RouterKey }

// Type returns the type for the message
func (msg MsgAssignRole) Type() string { return TypeMsgAssignRole }

// ValidateBasic validates the message
func (msg MsgAssignRole) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidTargetAddress)
	}
	if _, err := RoleFromString(msg.Role); err != nil {
		return ErrInvalidRole.Wrap(err.Error())
	}
	return nil
}

// GetSigners returns the signers for the message
func (msg MsgAssignRole) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns the sign bytes for the message
func (msg MsgAssignRole) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// MsgRevokeRole is the message for revoking a role from an account
type MsgRevokeRole struct {
	Sender  string `json:"sender"`
	Address string `json:"address"`
	Role    string `json:"role"`
}

// NewMsgRevokeRole creates a new MsgRevokeRole
func NewMsgRevokeRole(sender, address, role string) *MsgRevokeRole {
	return &MsgRevokeRole{
		Sender:  sender,
		Address: address,
		Role:    role,
	}
}

// Route returns the route for the message
func (msg MsgRevokeRole) Route() string { return RouterKey }

// Type returns the type for the message
func (msg MsgRevokeRole) Type() string { return TypeMsgRevokeRole }

// ValidateBasic validates the message
func (msg MsgRevokeRole) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidTargetAddress)
	}
	if _, err := RoleFromString(msg.Role); err != nil {
		return ErrInvalidRole.Wrap(err.Error())
	}
	return nil
}

// GetSigners returns the signers for the message
func (msg MsgRevokeRole) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns the sign bytes for the message
func (msg MsgRevokeRole) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// MsgSetAccountState is the message for setting an account's state
type MsgSetAccountState struct {
	Sender  string `json:"sender"`
	Address string `json:"address"`
	State   string `json:"state"`
	Reason  string `json:"reason"`

	// MFAProof is proof of MFA for sensitive account recovery operations
	MFAProof *mfatypes.MFAProof `json:"mfa_proof,omitempty"`

	// DeviceFingerprint is the client device fingerprint (optional)
	DeviceFingerprint string `json:"device_fingerprint,omitempty"`
}

// NewMsgSetAccountState creates a new MsgSetAccountState
func NewMsgSetAccountState(sender, address, state, reason string) *MsgSetAccountState {
	return &MsgSetAccountState{
		Sender:  sender,
		Address: address,
		State:   state,
		Reason:  reason,
	}
}

// Route returns the route for the message
func (msg MsgSetAccountState) Route() string { return RouterKey }

// Type returns the type for the message
func (msg MsgSetAccountState) Type() string { return TypeMsgSetAccountState }

// ValidateBasic validates the message
func (msg MsgSetAccountState) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidTargetAddress)
	}
	if _, err := AccountStateFromString(msg.State); err != nil {
		return ErrInvalidAccountState.Wrap(err.Error())
	}
	if msg.MFAProof != nil {
		if err := msg.MFAProof.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// GetSigners returns the signers for the message
func (msg MsgSetAccountState) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns the sign bytes for the message
func (msg MsgSetAccountState) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetMFAProof returns the MFA proof for gating.
func (msg MsgSetAccountState) GetMFAProof() *mfatypes.MFAProof {
	return msg.MFAProof
}

// GetDeviceFingerprint returns the device fingerprint for gating.
func (msg MsgSetAccountState) GetDeviceFingerprint() string {
	return msg.DeviceFingerprint
}

// MsgNominateAdmin is the message for nominating an administrator (GenesisAccount only)
type MsgNominateAdmin struct {
	Sender  string `json:"sender"`
	Address string `json:"address"`
}

// NewMsgNominateAdmin creates a new MsgNominateAdmin
func NewMsgNominateAdmin(sender, address string) *MsgNominateAdmin {
	return &MsgNominateAdmin{
		Sender:  sender,
		Address: address,
	}
}

// Route returns the route for the message
func (msg MsgNominateAdmin) Route() string { return RouterKey }

// Type returns the type for the message
func (msg MsgNominateAdmin) Type() string { return TypeMsgNominateAdmin }

// ValidateBasic validates the message
func (msg MsgNominateAdmin) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidTargetAddress)
	}
	return nil
}

// GetSigners returns the signers for the message
func (msg MsgNominateAdmin) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns the sign bytes for the message
func (msg MsgNominateAdmin) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// MsgUpdateParams is the message for updating module parameters (governance only)
type MsgUpdateParams struct {
	Authority string `json:"authority"`
	Params    Params `json:"params"`
}

// NewMsgUpdateParams creates a new MsgUpdateParams
func NewMsgUpdateParams(authority string, params Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

// Route returns the route for the message
func (msg MsgUpdateParams) Route() string { return RouterKey }

// Type returns the type for the message
func (msg MsgUpdateParams) Type() string { return "update_params" }

// ValidateBasic validates the message
func (msg MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrap("invalid authority address")
	}
	return msg.Params.Validate()
}

// GetSigners returns the signers for the message
func (msg MsgUpdateParams) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns the sign bytes for the message
func (msg MsgUpdateParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// MsgAssignRoleResponse is the response for MsgAssignRole
type MsgAssignRoleResponse struct{}

// MsgRevokeRoleResponse is the response for MsgRevokeRole
type MsgRevokeRoleResponse struct{}

// MsgSetAccountStateResponse is the response for MsgSetAccountState
type MsgSetAccountStateResponse struct{}

// MsgNominateAdminResponse is the response for MsgNominateAdmin
type MsgNominateAdminResponse struct{}

// MsgUpdateParamsResponse is the response for MsgUpdateParams
type MsgUpdateParamsResponse struct{}

// ============================================================================
// Sanction messages
//
// These mirror the generated rolesv1 types field-for-field, following the same
// pattern as the messages above. They are the on-chain entry points to the
// sanction state machine: without them the keeper's ImposeSanction,
// ConfirmSanction, RevokeSanction, OpenAppeal and ResolveAppeal would be
// unreachable and no account could be sanctioned at all.
// ============================================================================

// MsgImposeSanction records a scoped, time-limited sanction against an account.
//
// A suspension or termination it proposes takes effect only after a second,
// distinct moderator confirms it (MsgConfirmSanction). An emergency hold binds
// immediately but expires unless confirmed within its window.
type MsgImposeSanction struct {
	Sender          string `json:"sender"`
	Subject         string `json:"subject"`
	Scope           string `json:"scope"`
	ScopeRef        string `json:"scope_ref,omitempty"`
	Kind            string `json:"kind"`
	ReasonCode      string `json:"reason_code"`
	Justification   string `json:"justification"`
	Notice          string `json:"notice,omitempty"`
	DurationSeconds int64  `json:"duration_seconds,omitempty"`
}

// NewMsgImposeSanction creates a new MsgImposeSanction
func NewMsgImposeSanction(
	sender, subject, scope, scopeRef, kind, reasonCode, justification, notice string,
	durationSeconds int64,
) *MsgImposeSanction {
	return &MsgImposeSanction{
		Sender:          sender,
		Subject:         subject,
		Scope:           scope,
		ScopeRef:        scopeRef,
		Kind:            kind,
		ReasonCode:      reasonCode,
		Justification:   justification,
		Notice:          notice,
		DurationSeconds: durationSeconds,
	}
}

// Route returns the route for the message
func (msg MsgImposeSanction) Route() string { return RouterKey }

// Type returns the type for the message
func (msg MsgImposeSanction) Type() string { return TypeMsgImposeSanction }

// ValidateBasic validates the message
func (msg MsgImposeSanction) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}
	if _, err := sdk.AccAddressFromBech32(msg.Subject); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidTargetAddress)
	}
	if _, err := SanctionScopeFromString(msg.Scope); err != nil {
		return ErrInvalidSanctionScope.Wrap(err.Error())
	}
	if _, err := SanctionKindFromString(msg.Kind); err != nil {
		return ErrInvalidSanctionKind.Wrap(err.Error())
	}
	if _, err := SanctionReasonCodeFromString(msg.ReasonCode); err != nil {
		return ErrInvalidSanctionReason.Wrap(err.Error())
	}
	// A negative duration is meaningless and would produce an expiry in the past.
	if msg.DurationSeconds < 0 {
		return ErrInvalidSanction.Wrap("duration_seconds must not be negative")
	}
	return nil
}

// GetSigners returns the signers for the message
func (msg MsgImposeSanction) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns the sign bytes for the message
func (msg MsgImposeSanction) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// MsgImposeSanctionResponse is the response for MsgImposeSanction
type MsgImposeSanctionResponse struct {
	SanctionID string `json:"sanction_id"`
	Status     string `json:"status"`
}

// MsgConfirmSanction applies a pending sanction as its second, distinct reviewer
type MsgConfirmSanction struct {
	Reviewer      string `json:"reviewer"`
	SanctionID    string `json:"sanction_id"`
	ReviewedUntil int64  `json:"reviewed_until,omitempty"`
}

// NewMsgConfirmSanction creates a new MsgConfirmSanction
func NewMsgConfirmSanction(reviewer, sanctionID string, reviewedUntil int64) *MsgConfirmSanction {
	return &MsgConfirmSanction{
		Reviewer:      reviewer,
		SanctionID:    sanctionID,
		ReviewedUntil: reviewedUntil,
	}
}

// Route returns the route for the message
func (msg MsgConfirmSanction) Route() string { return RouterKey }

// Type returns the type for the message
func (msg MsgConfirmSanction) Type() string { return TypeMsgConfirmSanction }

// ValidateBasic validates the message
func (msg MsgConfirmSanction) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Reviewer); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}
	if strings.TrimSpace(msg.SanctionID) == "" {
		return ErrInvalidSanction.Wrap("sanction_id is required")
	}
	return nil
}

// GetSigners returns the signers for the message
func (msg MsgConfirmSanction) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Reviewer)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns the sign bytes for the message
func (msg MsgConfirmSanction) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// MsgConfirmSanctionResponse is the response for MsgConfirmSanction
type MsgConfirmSanctionResponse struct{}

// MsgRevokeSanction clears an in-force or pending sanction.
//
// Revocation is de-escalation, so it needs only one moderator-or-above actor.
type MsgRevokeSanction struct {
	Sender     string `json:"sender"`
	SanctionID string `json:"sanction_id"`
	Reason     string `json:"reason"`
}

// NewMsgRevokeSanction creates a new MsgRevokeSanction
func NewMsgRevokeSanction(sender, sanctionID, reason string) *MsgRevokeSanction {
	return &MsgRevokeSanction{
		Sender:     sender,
		SanctionID: sanctionID,
		Reason:     reason,
	}
}

// Route returns the route for the message
func (msg MsgRevokeSanction) Route() string { return RouterKey }

// Type returns the type for the message
func (msg MsgRevokeSanction) Type() string { return TypeMsgRevokeSanction }

// ValidateBasic validates the message
func (msg MsgRevokeSanction) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}
	if strings.TrimSpace(msg.SanctionID) == "" {
		return ErrInvalidSanction.Wrap("sanction_id is required")
	}
	return nil
}

// GetSigners returns the signers for the message
func (msg MsgRevokeSanction) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns the sign bytes for the message
func (msg MsgRevokeSanction) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// MsgRevokeSanctionResponse is the response for MsgRevokeSanction
type MsgRevokeSanctionResponse struct{}

// MsgOpenSanctionAppeal opens an appeal against an in-force sanction
type MsgOpenSanctionAppeal struct {
	Subject       string `json:"subject"`
	SanctionID    string `json:"sanction_id"`
	Justification string `json:"justification"`
}

// NewMsgOpenSanctionAppeal creates a new MsgOpenSanctionAppeal
func NewMsgOpenSanctionAppeal(subject, sanctionID, justification string) *MsgOpenSanctionAppeal {
	return &MsgOpenSanctionAppeal{
		Subject:       subject,
		SanctionID:    sanctionID,
		Justification: justification,
	}
}

// Route returns the route for the message
func (msg MsgOpenSanctionAppeal) Route() string { return RouterKey }

// Type returns the type for the message
func (msg MsgOpenSanctionAppeal) Type() string { return TypeMsgOpenSanctionAppeal }

// ValidateBasic validates the message
func (msg MsgOpenSanctionAppeal) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Subject); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}
	if strings.TrimSpace(msg.SanctionID) == "" {
		return ErrInvalidSanction.Wrap("sanction_id is required")
	}
	return nil
}

// GetSigners returns the signers for the message
func (msg MsgOpenSanctionAppeal) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Subject)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns the sign bytes for the message
func (msg MsgOpenSanctionAppeal) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// MsgOpenSanctionAppealResponse is the response for MsgOpenSanctionAppeal
type MsgOpenSanctionAppealResponse struct {
	AppealID string `json:"appeal_id"`
}

// MsgResolveSanctionAppeal resolves an open appeal
type MsgResolveSanctionAppeal struct {
	Reviewer string `json:"reviewer"`
	AppealID string `json:"appeal_id"`
	Grant    bool   `json:"grant"`
	Notes    string `json:"notes"`
}

// NewMsgResolveSanctionAppeal creates a new MsgResolveSanctionAppeal
func NewMsgResolveSanctionAppeal(reviewer, appealID string, grant bool, notes string) *MsgResolveSanctionAppeal {
	return &MsgResolveSanctionAppeal{
		Reviewer: reviewer,
		AppealID: appealID,
		Grant:    grant,
		Notes:    notes,
	}
}

// Route returns the route for the message
func (msg MsgResolveSanctionAppeal) Route() string { return RouterKey }

// Type returns the type for the message
func (msg MsgResolveSanctionAppeal) Type() string { return TypeMsgResolveSanctionAppeal }

// ValidateBasic validates the message
func (msg MsgResolveSanctionAppeal) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Reviewer); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}
	if strings.TrimSpace(msg.AppealID) == "" {
		return ErrInvalidSanction.Wrap("appeal_id is required")
	}
	return nil
}

// GetSigners returns the signers for the message
func (msg MsgResolveSanctionAppeal) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Reviewer)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns the sign bytes for the message
func (msg MsgResolveSanctionAppeal) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// MsgResolveSanctionAppealResponse is the response for MsgResolveSanctionAppeal
type MsgResolveSanctionAppealResponse struct{}
