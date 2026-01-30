package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message type constants for model messages
const (
	TypeMsgRegisterModel      = "register_model"
	TypeMsgProposeModelUpdate = "propose_model_update"
	TypeMsgReportModelVersion = "report_model_version"
	TypeMsgActivateModel      = "activate_model"
	TypeMsgDeprecateModel     = "deprecate_model"
	TypeMsgRevokeModel        = "revoke_model"
)

var (
	_ sdk.Msg = &MsgRegisterModel{}
	_ sdk.Msg = &MsgProposeModelUpdate{}
	_ sdk.Msg = &MsgReportModelVersion{}
	_ sdk.Msg = &MsgActivateModel{}
	_ sdk.Msg = &MsgDeprecateModel{}
	_ sdk.Msg = &MsgRevokeModel{}
)

// ============================================================================
// MsgRegisterModel
// ============================================================================

// Route returns the route for the message
func (msg *MsgRegisterModel) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgRegisterModel) Type() string { return TypeMsgRegisterModel }

// ValidateBasic validates the message
func (msg *MsgRegisterModel) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrap("invalid authority address")
	}

	// ModelInfo is a value type (not pointer)
	if msg.ModelInfo.ModelId == "" {
		return ErrModelNotFound.Wrap("model_id cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgRegisterModel) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{signer}
}

// ============================================================================
// MsgProposeModelUpdate
// ============================================================================

// Route returns the route for the message
func (msg *MsgProposeModelUpdate) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgProposeModelUpdate) Type() string { return TypeMsgProposeModelUpdate }

// ValidateBasic validates the message
func (msg *MsgProposeModelUpdate) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Proposer); err != nil {
		return ErrInvalidAddress.Wrap("invalid proposer address")
	}

	// Proposal is a value type (not pointer), validation is done via its own Validate method if needed
	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgProposeModelUpdate) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Proposer)
	return []sdk.AccAddress{signer}
}

// ============================================================================
// MsgReportModelVersion
// ============================================================================

// Route returns the route for the message
func (msg *MsgReportModelVersion) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgReportModelVersion) Type() string { return TypeMsgReportModelVersion }

// ValidateBasic validates the message
func (msg *MsgReportModelVersion) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ValidatorAddress); err != nil {
		return ErrInvalidAddress.Wrap("invalid validator address")
	}

	if len(msg.ModelVersions) == 0 {
		return ErrModelVersionMismatch.Wrap("model_versions cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgReportModelVersion) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.ValidatorAddress)
	return []sdk.AccAddress{signer}
}

// ============================================================================
// MsgActivateModel
// ============================================================================

// Route returns the route for the message
func (msg *MsgActivateModel) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgActivateModel) Type() string { return TypeMsgActivateModel }

// ValidateBasic validates the message
func (msg *MsgActivateModel) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrap("invalid authority address")
	}

	if msg.ModelId == "" {
		return ErrModelNotFound.Wrap("model_id cannot be empty")
	}

	if msg.ModelType == "" {
		return ErrModelNotFound.Wrap("model_type cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgActivateModel) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{signer}
}

// ============================================================================
// MsgDeprecateModel
// ============================================================================

// Route returns the route for the message
func (msg *MsgDeprecateModel) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgDeprecateModel) Type() string { return TypeMsgDeprecateModel }

// ValidateBasic validates the message
func (msg *MsgDeprecateModel) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrap("invalid authority address")
	}

	if msg.ModelId == "" {
		return ErrModelNotFound.Wrap("model_id cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgDeprecateModel) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{signer}
}

// ============================================================================
// MsgRevokeModel
// ============================================================================

// Route returns the route for the message
func (msg *MsgRevokeModel) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgRevokeModel) Type() string { return TypeMsgRevokeModel }

// ValidateBasic validates the message
func (msg *MsgRevokeModel) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrap("invalid authority address")
	}

	if msg.ModelId == "" {
		return ErrModelNotFound.Wrap("model_id cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgRevokeModel) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{signer}
}
