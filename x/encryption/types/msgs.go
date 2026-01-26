package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message type constants
const (
	TypeMsgRegisterRecipientKey = "register_recipient_key"
	TypeMsgRevokeRecipientKey   = "revoke_recipient_key"
	TypeMsgUpdateKeyLabel       = "update_key_label"
)

// Error message constants
const (
	errMsgInvalidSenderAddress = "invalid sender address"
)

var (
	_ sdk.Msg = &MsgRegisterRecipientKey{}
	_ sdk.Msg = &MsgRevokeRecipientKey{}
	_ sdk.Msg = &MsgUpdateKeyLabel{}
)

// MsgRegisterRecipientKey is the message for registering a recipient public key
type MsgRegisterRecipientKey struct {
	// Sender is the account registering the key (must match the key owner)
	Sender string `json:"sender"`

	// PublicKey is the X25519 public key bytes (32 bytes)
	PublicKey []byte `json:"public_key"`

	// AlgorithmID specifies which algorithm this key is for
	AlgorithmID string `json:"algorithm_id"`

	// Label is an optional human-readable label for the key
	Label string `json:"label,omitempty"`
}

// NewMsgRegisterRecipientKey creates a new MsgRegisterRecipientKey
func NewMsgRegisterRecipientKey(sender string, publicKey []byte, algorithmID, label string) *MsgRegisterRecipientKey {
	return &MsgRegisterRecipientKey{
		Sender:      sender,
		PublicKey:   publicKey,
		AlgorithmID: algorithmID,
		Label:       label,
	}
}

// Route returns the route for the message
func (msg MsgRegisterRecipientKey) Route() string { return RouterKey }

// Type returns the type for the message
func (msg MsgRegisterRecipientKey) Type() string { return TypeMsgRegisterRecipientKey }

// ValidateBasic validates the message
func (msg MsgRegisterRecipientKey) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}

	if len(msg.PublicKey) == 0 {
		return ErrInvalidPublicKey.Wrap("public key cannot be empty")
	}

	if !IsAlgorithmSupported(msg.AlgorithmID) {
		return ErrUnsupportedAlgorithm.Wrapf("algorithm %s is not supported", msg.AlgorithmID)
	}

	// Validate key size for algorithm
	algInfo, err := GetAlgorithmInfo(msg.AlgorithmID)
	if err != nil {
		return err
	}

	if len(msg.PublicKey) != algInfo.KeySize {
		return ErrInvalidPublicKey.Wrapf("public key size must be %d bytes for %s",
			algInfo.KeySize, msg.AlgorithmID)
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg MsgRegisterRecipientKey) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns the sign bytes for the message
func (msg MsgRegisterRecipientKey) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// MsgRevokeRecipientKey is the message for revoking a recipient public key
type MsgRevokeRecipientKey struct {
	// Sender is the account revoking the key (must own the key)
	Sender string `json:"sender"`

	// KeyFingerprint is the fingerprint of the key to revoke
	KeyFingerprint string `json:"key_fingerprint"`
}

// NewMsgRevokeRecipientKey creates a new MsgRevokeRecipientKey
func NewMsgRevokeRecipientKey(sender, keyFingerprint string) *MsgRevokeRecipientKey {
	return &MsgRevokeRecipientKey{
		Sender:         sender,
		KeyFingerprint: keyFingerprint,
	}
}

// Route returns the route for the message
func (msg MsgRevokeRecipientKey) Route() string { return RouterKey }

// Type returns the type for the message
func (msg MsgRevokeRecipientKey) Type() string { return TypeMsgRevokeRecipientKey }

// ValidateBasic validates the message
func (msg MsgRevokeRecipientKey) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}

	if len(msg.KeyFingerprint) == 0 {
		return ErrInvalidKeyFingerprint.Wrap("key fingerprint cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg MsgRevokeRecipientKey) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns the sign bytes for the message
func (msg MsgRevokeRecipientKey) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// MsgUpdateKeyLabel is the message for updating a key's label
type MsgUpdateKeyLabel struct {
	// Sender is the account updating the key (must own the key)
	Sender string `json:"sender"`

	// KeyFingerprint is the fingerprint of the key to update
	KeyFingerprint string `json:"key_fingerprint"`

	// Label is the new label for the key
	Label string `json:"label"`
}

// NewMsgUpdateKeyLabel creates a new MsgUpdateKeyLabel
func NewMsgUpdateKeyLabel(sender, keyFingerprint, label string) *MsgUpdateKeyLabel {
	return &MsgUpdateKeyLabel{
		Sender:         sender,
		KeyFingerprint: keyFingerprint,
		Label:          label,
	}
}

// Route returns the route for the message
func (msg MsgUpdateKeyLabel) Route() string { return RouterKey }

// Type returns the type for the message
func (msg MsgUpdateKeyLabel) Type() string { return TypeMsgUpdateKeyLabel }

// ValidateBasic validates the message
func (msg MsgUpdateKeyLabel) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}

	if len(msg.KeyFingerprint) == 0 {
		return ErrInvalidKeyFingerprint.Wrap("key fingerprint cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg MsgUpdateKeyLabel) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns the sign bytes for the message
func (msg MsgUpdateKeyLabel) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// Message response types

// MsgRegisterRecipientKeyResponse is the response for MsgRegisterRecipientKey
type MsgRegisterRecipientKeyResponse struct {
	// KeyFingerprint is the fingerprint of the registered key
	KeyFingerprint string `json:"key_fingerprint"`
}

// MsgRevokeRecipientKeyResponse is the response for MsgRevokeRecipientKey
type MsgRevokeRecipientKeyResponse struct{}

// MsgUpdateKeyLabelResponse is the response for MsgUpdateKeyLabel
type MsgUpdateKeyLabelResponse struct{}
