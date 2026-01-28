package types

import (
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
)

var (
	_ sdk.Msg = &MsgAssignRole{}
	_ sdk.Msg = &MsgRevokeRole{}
	_ sdk.Msg = &MsgSetAccountState{}
	_ sdk.Msg = &MsgNominateAdmin{}
	_ sdk.Msg = &MsgUpdateParams{}
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
