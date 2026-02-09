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
	_ sdk.Msg = &MsgSubmitSSOVerificationProof{}
	_ sdk.Msg = &MsgSubmitEmailVerificationProof{}
	_ sdk.Msg = &MsgSubmitSMSVerificationProof{}
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

// Route returns the route for the message
func (msg *MsgSubmitSSOVerificationProof) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgSubmitSSOVerificationProof) Type() string { return "submit_sso_verification_proof" }

// ValidateBasic validates the message
func (msg *MsgSubmitSSOVerificationProof) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.AccountAddress); err != nil {
		return ErrInvalidAddress.Wrap("invalid account address")
	}
	if msg.LinkageId == "" {
		return ErrInvalidSSO.Wrap("linkage_id cannot be empty")
	}
	if len(msg.AttestationData) == 0 {
		return ErrInvalidAttestation.Wrap("attestation_data cannot be empty")
	}
	if msg.EvidenceStorageRef == "" {
		return ErrInvalidPayload.Wrap("evidence_storage_ref cannot be empty")
	}
	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgSubmitSSOVerificationProof) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.AccountAddress)
	return []sdk.AccAddress{signer}
}

// Route returns the route for the message
func (msg *MsgSubmitEmailVerificationProof) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgSubmitEmailVerificationProof) Type() string { return "submit_email_verification_proof" }

// ValidateBasic validates the message
func (msg *MsgSubmitEmailVerificationProof) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.AccountAddress); err != nil {
		return ErrInvalidAddress.Wrap("invalid account address")
	}
	if msg.VerificationId == "" {
		return ErrInvalidEmail.Wrap("verification_id cannot be empty")
	}
	if msg.EmailHash == "" {
		return ErrInvalidEmail.Wrap("email_hash cannot be empty")
	}
	if msg.Nonce == "" {
		return ErrInvalidEmail.Wrap("nonce cannot be empty")
	}
	if len(msg.AttestationData) == 0 {
		return ErrInvalidAttestation.Wrap("attestation_data cannot be empty")
	}
	if len(msg.AccountSignature) == 0 {
		return ErrInvalidBindingSignature.Wrap("account_signature cannot be empty")
	}
	if msg.EvidenceStorageRef == "" {
		return ErrInvalidPayload.Wrap("evidence_storage_ref cannot be empty")
	}
	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgSubmitEmailVerificationProof) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.AccountAddress)
	return []sdk.AccAddress{signer}
}

// Route returns the route for the message
func (msg *MsgSubmitSMSVerificationProof) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgSubmitSMSVerificationProof) Type() string { return "submit_sms_verification_proof" }

// ValidateBasic validates the message
func (msg *MsgSubmitSMSVerificationProof) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.AccountAddress); err != nil {
		return ErrInvalidAddress.Wrap("invalid account address")
	}
	if msg.VerificationId == "" {
		return ErrInvalidPhone.Wrap("verification_id cannot be empty")
	}
	if msg.PhoneHash == "" || msg.PhoneHashSalt == "" {
		return ErrInvalidPhone.Wrap("phone_hash and phone_hash_salt cannot be empty")
	}
	if len(msg.AttestationData) == 0 {
		return ErrInvalidAttestation.Wrap("attestation_data cannot be empty")
	}
	if len(msg.AccountSignature) == 0 {
		return ErrInvalidBindingSignature.Wrap("account_signature cannot be empty")
	}
	if msg.EvidenceStorageRef == "" {
		return ErrInvalidPayload.Wrap("evidence_storage_ref cannot be empty")
	}
	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgSubmitSMSVerificationProof) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.AccountAddress)
	return []sdk.AccAddress{signer}
}
