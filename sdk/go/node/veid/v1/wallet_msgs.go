package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message type constants for wallet messages
const (
	TypeMsgCreateIdentityWallet  = "create_identity_wallet"
	TypeMsgAddScopeToWallet      = "add_scope_to_wallet"
	TypeMsgRevokeScopeFromWallet = "revoke_scope_from_wallet"
	TypeMsgUpdateConsentSettings = "update_consent_settings"
	TypeMsgRebindWallet          = "rebind_wallet"
	TypeMsgUpdateDerivedFeatures = "update_derived_features"
)

var (
	_ sdk.Msg = &MsgCreateIdentityWallet{}
	_ sdk.Msg = &MsgAddScopeToWallet{}
	_ sdk.Msg = &MsgRevokeScopeFromWallet{}
	_ sdk.Msg = &MsgUpdateConsentSettings{}
	_ sdk.Msg = &MsgRebindWallet{}
	_ sdk.Msg = &MsgUpdateDerivedFeatures{}
)

// ============================================================================
// MsgCreateIdentityWallet
// ============================================================================

// Route returns the route for the message
func (msg *MsgCreateIdentityWallet) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgCreateIdentityWallet) Type() string { return TypeMsgCreateIdentityWallet }

// ValidateBasic validates the message
func (msg *MsgCreateIdentityWallet) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if len(msg.BindingSignature) == 0 {
		return ErrWalletBindingFailed.Wrap("binding signature cannot be empty")
	}

	if len(msg.BindingPubKey) == 0 {
		return ErrWalletBindingFailed.Wrap("binding public key cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgCreateIdentityWallet) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// ============================================================================
// MsgAddScopeToWallet
// ============================================================================

// Route returns the route for the message
func (msg *MsgAddScopeToWallet) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgAddScopeToWallet) Type() string { return TypeMsgAddScopeToWallet }

// ValidateBasic validates the message
func (msg *MsgAddScopeToWallet) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.ScopeId == "" {
		return ErrInvalidScope.Wrap("scope_id cannot be empty")
	}

	if msg.ScopeType == ScopeTypeUnspecified {
		return ErrInvalidScopeType.Wrap("invalid scope type")
	}

	if len(msg.EnvelopeHash) == 0 {
		return ErrInvalidPayloadHash.Wrap("envelope_hash cannot be empty")
	}

	if len(msg.UserSignature) == 0 {
		return ErrInvalidUserSignature.Wrap("user_signature cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgAddScopeToWallet) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// ============================================================================
// MsgRevokeScopeFromWallet
// ============================================================================

// Route returns the route for the message
func (msg *MsgRevokeScopeFromWallet) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgRevokeScopeFromWallet) Type() string { return TypeMsgRevokeScopeFromWallet }

// ValidateBasic validates the message
func (msg *MsgRevokeScopeFromWallet) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.ScopeId == "" {
		return ErrInvalidScope.Wrap("scope_id cannot be empty")
	}

	if len(msg.UserSignature) == 0 {
		return ErrInvalidUserSignature.Wrap("user_signature cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgRevokeScopeFromWallet) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// ============================================================================
// MsgUpdateConsentSettings
// ============================================================================

// Route returns the route for the message
func (msg *MsgUpdateConsentSettings) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgUpdateConsentSettings) Type() string { return TypeMsgUpdateConsentSettings }

// ValidateBasic validates the message
func (msg *MsgUpdateConsentSettings) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.NewSettings == nil {
		return ErrInvalidConsent.Wrap("new_settings cannot be nil")
	}

	if len(msg.UserSignature) == 0 {
		return ErrInvalidUserSignature.Wrap("user_signature cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgUpdateConsentSettings) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// ============================================================================
// MsgRebindWallet
// ============================================================================

// Route returns the route for the message
func (msg *MsgRebindWallet) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgRebindWallet) Type() string { return TypeMsgRebindWallet }

// ValidateBasic validates the message
func (msg *MsgRebindWallet) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.CurrentOwner); err != nil {
		return ErrInvalidAddress.Wrap("invalid current owner address")
	}

	if _, err := sdk.AccAddressFromBech32(msg.NewOwner); err != nil {
		return ErrInvalidAddress.Wrap("invalid new owner address")
	}

	if len(msg.CurrentOwnerSignature) == 0 {
		return ErrInvalidUserSignature.Wrap("current_owner_signature cannot be empty")
	}

	if len(msg.NewOwnerSignature) == 0 {
		return ErrInvalidUserSignature.Wrap("new_owner_signature cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgRebindWallet) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.CurrentOwner)
	return []sdk.AccAddress{signer}
}

// ============================================================================
// MsgUpdateDerivedFeatures
// ============================================================================

// Route returns the route for the message
func (msg *MsgUpdateDerivedFeatures) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgUpdateDerivedFeatures) Type() string { return TypeMsgUpdateDerivedFeatures }

// ValidateBasic validates the message
func (msg *MsgUpdateDerivedFeatures) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.ScopeId == "" {
		return ErrInvalidScope.Wrap("scope_id cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgUpdateDerivedFeatures) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}
