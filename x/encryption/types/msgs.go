package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	encryptionv1 "github.com/virtengine/virtengine/sdk/go/node/encryption/v1"
)

// Type aliases for generated protobuf message types
type (
	MsgRegisterRecipientKey         = encryptionv1.MsgRegisterRecipientKey
	MsgRevokeRecipientKey           = encryptionv1.MsgRevokeRecipientKey
	MsgUpdateKeyLabel               = encryptionv1.MsgUpdateKeyLabel
	MsgRegisterRecipientKeyResponse = encryptionv1.MsgRegisterRecipientKeyResponse
	MsgRevokeRecipientKeyResponse   = encryptionv1.MsgRevokeRecipientKeyResponse
	MsgUpdateKeyLabelResponse       = encryptionv1.MsgUpdateKeyLabelResponse
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

// NewMsgRegisterRecipientKey creates a new MsgRegisterRecipientKey
func NewMsgRegisterRecipientKey(sender string, publicKey []byte, algorithmID, label string) *MsgRegisterRecipientKey {
	return &MsgRegisterRecipientKey{
		Sender:      sender,
		PublicKey:   publicKey,
		AlgorithmId: algorithmID,
		Label:       label,
	}
}

// ValidateMsgRegisterRecipientKey validates the message
func ValidateMsgRegisterRecipientKey(msg *MsgRegisterRecipientKey) error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}

	if len(msg.PublicKey) == 0 {
		return ErrInvalidPublicKey.Wrap("public key cannot be empty")
	}

	if !IsAlgorithmSupported(msg.AlgorithmId) {
		return ErrUnsupportedAlgorithm.Wrapf("algorithm %s is not supported", msg.AlgorithmId)
	}

	// Validate key size for algorithm
	algInfo, err := GetAlgorithmInfo(msg.AlgorithmId)
	if err != nil {
		return err
	}

	if len(msg.PublicKey) != algInfo.KeySize {
		return ErrInvalidPublicKey.Wrapf("public key size must be %d bytes for %s",
			algInfo.KeySize, msg.AlgorithmId)
	}

	return nil
}

// NewMsgRevokeRecipientKey creates a new MsgRevokeRecipientKey
func NewMsgRevokeRecipientKey(sender, keyFingerprint string) *MsgRevokeRecipientKey {
	return &MsgRevokeRecipientKey{
		Sender:         sender,
		KeyFingerprint: keyFingerprint,
	}
}

// ValidateMsgRevokeRecipientKey validates the message
func ValidateMsgRevokeRecipientKey(msg *MsgRevokeRecipientKey) error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}

	if len(msg.KeyFingerprint) == 0 {
		return ErrInvalidKeyFingerprint.Wrap("key fingerprint cannot be empty")
	}

	return nil
}

// NewMsgUpdateKeyLabel creates a new MsgUpdateKeyLabel
func NewMsgUpdateKeyLabel(sender, keyFingerprint, label string) *MsgUpdateKeyLabel {
	return &MsgUpdateKeyLabel{
		Sender:         sender,
		KeyFingerprint: keyFingerprint,
		Label:          label,
	}
}

// ValidateMsgUpdateKeyLabel validates the message
func ValidateMsgUpdateKeyLabel(msg *MsgUpdateKeyLabel) error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}

	if len(msg.KeyFingerprint) == 0 {
		return ErrInvalidKeyFingerprint.Wrap("key fingerprint cannot be empty")
	}

	return nil
}
