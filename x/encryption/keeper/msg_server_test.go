package keeper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/encryption/types"
)

func TestNewMsgRegisterRecipientKey(t *testing.T) {
	publicKey := make([]byte, 32)
	msg := types.NewMsgRegisterRecipientKey("cosmos1abc...", publicKey, types.AlgorithmX25519XSalsa20Poly1305, "my-key")

	assert.Equal(t, "cosmos1abc...", msg.Sender)
	assert.Equal(t, publicKey, msg.PublicKey)
	assert.Equal(t, types.AlgorithmX25519XSalsa20Poly1305, msg.AlgorithmId)
	assert.Equal(t, "my-key", msg.Label)
}

func TestMsgRegisterRecipientKey_ValidateBasic(t *testing.T) {
	tests := []struct {
		name      string
		msg       types.MsgRegisterRecipientKey
		expectErr bool
	}{
		{
			name: "valid message",
			msg: types.MsgRegisterRecipientKey{
				Sender:      "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
				PublicKey:   make([]byte, 32),
				AlgorithmId: types.AlgorithmX25519XSalsa20Poly1305,
				Label:       "test",
			},
			expectErr: false,
		},
		{
			name: "invalid sender address",
			msg: types.MsgRegisterRecipientKey{
				Sender:      "invalid",
				PublicKey:   make([]byte, 32),
				AlgorithmId: types.AlgorithmX25519XSalsa20Poly1305,
			},
			expectErr: true,
		},
		{
			name: "empty public key",
			msg: types.MsgRegisterRecipientKey{
				Sender:      "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
				PublicKey:   []byte{},
				AlgorithmId: types.AlgorithmX25519XSalsa20Poly1305,
			},
			expectErr: true,
		},
		{
			name: "wrong public key size",
			msg: types.MsgRegisterRecipientKey{
				Sender:      "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
				PublicKey:   make([]byte, 16),
				AlgorithmId: types.AlgorithmX25519XSalsa20Poly1305,
			},
			expectErr: true,
		},
		{
			name: "unsupported algorithm",
			msg: types.MsgRegisterRecipientKey{
				Sender:      "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
				PublicKey:   make([]byte, 32),
				AlgorithmId: "UNKNOWN",
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateMsgRegisterRecipientKey(&tc.msg)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgRevokeRecipientKey_ValidateBasic(t *testing.T) {
	tests := []struct {
		name      string
		msg       types.MsgRevokeRecipientKey
		expectErr bool
	}{
		{
			name: "valid message",
			msg: types.MsgRevokeRecipientKey{
				Sender:         "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
				KeyFingerprint: "abc123def456",
			},
			expectErr: false,
		},
		{
			name: "invalid sender address",
			msg: types.MsgRevokeRecipientKey{
				Sender:         "invalid",
				KeyFingerprint: "abc123def456",
			},
			expectErr: true,
		},
		{
			name: "empty fingerprint",
			msg: types.MsgRevokeRecipientKey{
				Sender:         "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
				KeyFingerprint: "",
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateMsgRevokeRecipientKey(&tc.msg)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgUpdateKeyLabel_ValidateBasic(t *testing.T) {
	tests := []struct {
		name      string
		msg       types.MsgUpdateKeyLabel
		expectErr bool
	}{
		{
			name: "valid message",
			msg: types.MsgUpdateKeyLabel{
				Sender:         "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
				KeyFingerprint: "abc123def456",
				Label:          "new-label",
			},
			expectErr: false,
		},
		{
			name: "valid message with empty label",
			msg: types.MsgUpdateKeyLabel{
				Sender:         "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
				KeyFingerprint: "abc123def456",
				Label:          "", // Empty label is allowed
			},
			expectErr: false,
		},
		{
			name: "invalid sender address",
			msg: types.MsgUpdateKeyLabel{
				Sender:         "invalid",
				KeyFingerprint: "abc123def456",
				Label:          "test",
			},
			expectErr: true,
		},
		{
			name: "empty fingerprint",
			msg: types.MsgUpdateKeyLabel{
				Sender:         "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
				KeyFingerprint: "",
				Label:          "test",
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateMsgUpdateKeyLabel(&tc.msg)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
