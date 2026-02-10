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
	MsgRotateKey                    = encryptionv1.MsgRotateKey
	MsgRegisterRecipientKeyResponse = encryptionv1.MsgRegisterRecipientKeyResponse
	MsgRevokeRecipientKeyResponse   = encryptionv1.MsgRevokeRecipientKeyResponse
	MsgUpdateKeyLabelResponse       = encryptionv1.MsgUpdateKeyLabelResponse
	MsgRotateKeyResponse            = encryptionv1.MsgRotateKeyResponse
)

// Message type constants
const (
	TypeMsgRegisterRecipientKey = "register_recipient_key"
	TypeMsgRevokeRecipientKey   = "revoke_recipient_key"
	TypeMsgUpdateKeyLabel       = "update_key_label"
	TypeMsgRotateKey            = "rotate_key"
)

// Error message constants
const (
	errMsgInvalidSenderAddress = "invalid sender address"
)

var (
	_ sdk.Msg = &MsgRegisterRecipientKey{}
	_ sdk.Msg = &MsgRevokeRecipientKey{}
	_ sdk.Msg = &MsgUpdateKeyLabel{}
	_ sdk.Msg = &MsgRotateKey{}
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

// NewMsgRotateKey creates a new MsgRotateKey
func NewMsgRotateKey(sender, oldKeyFingerprint string, newPublicKey []byte, newAlgorithmID, newLabel, reason string, newKeyTTLSeconds uint64) *MsgRotateKey {
	return &MsgRotateKey{
		Sender:            sender,
		OldKeyFingerprint: oldKeyFingerprint,
		NewPublicKey:      newPublicKey,
		NewAlgorithmId:    newAlgorithmID,
		NewLabel:          newLabel,
		Reason:            reason,
		NewKeyTtlSeconds:  newKeyTTLSeconds,
	}
}

// ValidateMsgRotateKey validates the message
func ValidateMsgRotateKey(msg *MsgRotateKey) error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}

	if len(msg.OldKeyFingerprint) == 0 {
		return ErrInvalidKeyFingerprint.Wrap("old key fingerprint cannot be empty")
	}

	if len(msg.NewPublicKey) == 0 {
		return ErrInvalidPublicKey.Wrap("new public key cannot be empty")
	}

	algorithmID := msg.NewAlgorithmId
	if algorithmID == "" {
		algorithmID = DefaultAlgorithm()
	}

	if !IsAlgorithmSupported(algorithmID) {
		return ErrUnsupportedAlgorithm.Wrapf("algorithm %s is not supported", algorithmID)
	}

	algInfo, err := GetAlgorithmInfo(algorithmID)
	if err != nil {
		return err
	}

	if len(msg.NewPublicKey) != algInfo.KeySize {
		return ErrInvalidPublicKey.Wrapf("public key size must be %d bytes for %s",
			algInfo.KeySize, algorithmID)
	}

	return nil
}
