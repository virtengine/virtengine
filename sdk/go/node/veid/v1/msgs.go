package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgUploadScope{}
	_ sdk.Msg = &MsgRevokeScope{}
	_ sdk.Msg = &MsgRequestVerification{}
	_ sdk.Msg = &MsgUpdateVerificationStatus{}
	_ sdk.Msg = &MsgUpdateScore{}
)

// Route returns the route for the message
func (msg *MsgUploadScope) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgUploadScope) Type() string { return "upload_scope" }

// ValidateBasic validates the message
func (msg *MsgUploadScope) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.ScopeId == "" {
		return ErrInvalidScope.Wrap("scope_id cannot be empty")
	}

	if len(msg.ScopeId) > 128 {
		return ErrInvalidScope.Wrap("scope_id exceeds maximum length")
	}

	if msg.ScopeType == ScopeTypeUnspecified {
		return ErrInvalidScopeType.Wrapf("invalid scope type: %s", msg.ScopeType)
	}

	if err := msg.EncryptedPayload.Validate(); err != nil {
		return ErrInvalidPayload.Wrap(err.Error())
	}

	if len(msg.Salt) < 16 {
		return ErrInvalidSalt.Wrap("salt must be at least 16 bytes")
	}

	if len(msg.Salt) > 64 {
		return ErrInvalidSalt.Wrap("salt cannot exceed 64 bytes")
	}

	if msg.DeviceFingerprint == "" {
		return ErrInvalidDeviceInfo.Wrap("device fingerprint cannot be empty")
	}

	if msg.ClientId == "" {
		return ErrInvalidClientID.Wrap("client_id cannot be empty")
	}

	if len(msg.ClientSignature) == 0 {
		return ErrInvalidClientSignature.Wrap("client signature cannot be empty")
	}

	if len(msg.UserSignature) == 0 {
		return ErrInvalidUserSignature.Wrap("user signature cannot be empty")
	}

	if len(msg.PayloadHash) != 32 {
		return ErrInvalidPayloadHash.Wrap("payload hash must be 32 bytes (SHA256)")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgUploadScope) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// Route returns the route for the message
func (msg *MsgRevokeScope) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgRevokeScope) Type() string { return "revoke_scope" }

// ValidateBasic validates the message
func (msg *MsgRevokeScope) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.ScopeId == "" {
		return ErrInvalidScope.Wrap("scope_id cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgRevokeScope) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// Route returns the route for the message
func (msg *MsgRequestVerification) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgRequestVerification) Type() string { return "request_verification" }

// ValidateBasic validates the message
func (msg *MsgRequestVerification) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.ScopeId == "" {
		return ErrInvalidScope.Wrap("scope_id cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgRequestVerification) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// Route returns the route for the message
func (msg *MsgUpdateVerificationStatus) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgUpdateVerificationStatus) Type() string { return "update_verification_status" }

// ValidateBasic validates the message
func (msg *MsgUpdateVerificationStatus) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if _, err := sdk.AccAddressFromBech32(msg.AccountAddress); err != nil {
		return ErrInvalidAddress.Wrap("invalid account address")
	}

	if msg.ScopeId == "" {
		return ErrInvalidScope.Wrap("scope_id cannot be empty")
	}

	if msg.NewStatus == VerificationStatusUnknown {
		return ErrInvalidVerificationStatus.Wrapf("invalid status: %s", msg.NewStatus)
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgUpdateVerificationStatus) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// Route returns the route for the message
func (msg *MsgUpdateScore) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgUpdateScore) Type() string { return "update_score" }

// ValidateBasic validates the message
func (msg *MsgUpdateScore) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if _, err := sdk.AccAddressFromBech32(msg.AccountAddress); err != nil {
		return ErrInvalidAddress.Wrap("invalid account address")
	}

	if msg.NewScore > 100 {
		return ErrInvalidScore.Wrap("score cannot exceed 100")
	}

	if msg.ScoreVersion == "" {
		return ErrInvalidScore.Wrap("score version cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgUpdateScore) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

