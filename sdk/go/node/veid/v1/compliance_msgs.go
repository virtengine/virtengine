package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message type constants for compliance messages
const (
	TypeMsgSubmitComplianceCheck       = "submit_compliance_check"
	TypeMsgAttestCompliance            = "attest_compliance"
	TypeMsgUpdateComplianceParams      = "update_compliance_params"
	TypeMsgRegisterComplianceProvider  = "register_compliance_provider"
	TypeMsgDeactivateComplianceProvider = "deactivate_compliance_provider"
)

var (
	_ sdk.Msg = &MsgSubmitComplianceCheck{}
	_ sdk.Msg = &MsgAttestCompliance{}
	_ sdk.Msg = &MsgUpdateComplianceParams{}
	_ sdk.Msg = &MsgRegisterComplianceProvider{}
	_ sdk.Msg = &MsgDeactivateComplianceProvider{}
)

// ============================================================================
// MsgSubmitComplianceCheck
// ============================================================================

// Route returns the route for the message
func (msg *MsgSubmitComplianceCheck) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgSubmitComplianceCheck) Type() string { return TypeMsgSubmitComplianceCheck }

// ValidateBasic validates the message
func (msg *MsgSubmitComplianceCheck) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return ErrNotComplianceProvider.Wrap("invalid provider address")
	}

	if _, err := sdk.AccAddressFromBech32(msg.TargetAddress); err != nil {
		return ErrInvalidAddress.Wrap("invalid target address")
	}

	if len(msg.CheckResults) == 0 {
		return ErrComplianceCheckFailed.Wrap("at least one check result is required")
	}

	if msg.ProviderId == "" {
		return ErrNotComplianceProvider.Wrap("provider_id cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgSubmitComplianceCheck) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{signer}
}

// ============================================================================
// MsgAttestCompliance
// ============================================================================

// Route returns the route for the message
func (msg *MsgAttestCompliance) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgAttestCompliance) Type() string { return TypeMsgAttestCompliance }

// ValidateBasic validates the message
func (msg *MsgAttestCompliance) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ValidatorAddress); err != nil {
		return ErrInvalidAddress.Wrap("invalid validator address")
	}

	if _, err := sdk.AccAddressFromBech32(msg.TargetAddress); err != nil {
		return ErrInvalidAddress.Wrap("invalid target address")
	}

	if msg.AttestationType == "" {
		return ErrInsufficientAttestations.Wrap("attestation_type cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgAttestCompliance) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.ValidatorAddress)
	return []sdk.AccAddress{signer}
}

// ============================================================================
// MsgUpdateComplianceParams
// ============================================================================

// Route returns the route for the message
func (msg *MsgUpdateComplianceParams) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgUpdateComplianceParams) Type() string { return TypeMsgUpdateComplianceParams }

// ValidateBasic validates the message
func (msg *MsgUpdateComplianceParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrap("invalid authority address")
	}

	if msg.Params == nil {
		return ErrInvalidComplianceParams.Wrap("params cannot be nil")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgUpdateComplianceParams) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{signer}
}

// ============================================================================
// MsgRegisterComplianceProvider
// ============================================================================

// Route returns the route for the message
func (msg *MsgRegisterComplianceProvider) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgRegisterComplianceProvider) Type() string { return TypeMsgRegisterComplianceProvider }

// ValidateBasic validates the message
func (msg *MsgRegisterComplianceProvider) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrap("invalid authority address")
	}

	if msg.Provider == nil {
		return ErrNotComplianceProvider.Wrap("provider cannot be nil")
	}

	if msg.Provider.ProviderId == "" {
		return ErrNotComplianceProvider.Wrap("provider_id cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgRegisterComplianceProvider) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{signer}
}

// ============================================================================
// MsgDeactivateComplianceProvider
// ============================================================================

// Route returns the route for the message
func (msg *MsgDeactivateComplianceProvider) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgDeactivateComplianceProvider) Type() string { return TypeMsgDeactivateComplianceProvider }

// ValidateBasic validates the message
func (msg *MsgDeactivateComplianceProvider) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrap("invalid authority address")
	}

	if msg.ProviderId == "" {
		return ErrNotComplianceProvider.Wrap("provider_id cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgDeactivateComplianceProvider) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{signer}
}
