package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptedPayloadEnvelope_NewEnvelope(t *testing.T) {
	envelope := NewEncryptedPayloadEnvelope()

	assert.Equal(t, EnvelopeVersion, envelope.Version)
	assert.Equal(t, DefaultAlgorithm(), envelope.AlgorithmID)
	assert.Empty(t, envelope.RecipientKeyIDs)
	assert.NotNil(t, envelope.Metadata)
}

func TestEncryptedPayloadEnvelope_Validate(t *testing.T) {
	tests := []struct {
		name      string
		envelope  *EncryptedPayloadEnvelope
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid envelope",
			envelope: &EncryptedPayloadEnvelope{
				Version:         1,
				AlgorithmID:     AlgorithmX25519XSalsa20Poly1305,
				RecipientKeyIDs: []string{"abc123def456"},
				Nonce:           make([]byte, XSalsa20NonceSize),
				Ciphertext:      []byte("encrypted data"),
				SenderPubKey:    make([]byte, X25519PublicKeySize),
			},
			expectErr: false,
		},
		{
			name: "zero version",
			envelope: &EncryptedPayloadEnvelope{
				Version:         0,
				AlgorithmID:     AlgorithmX25519XSalsa20Poly1305,
				RecipientKeyIDs: []string{"abc123def456"},
				Nonce:           make([]byte, XSalsa20NonceSize),
				Ciphertext:      []byte("encrypted data"),
				SenderPubKey:    make([]byte, X25519PublicKeySize),
			},
			expectErr: true,
			errMsg:    "version cannot be zero",
		},
		{
			name: "unsupported version",
			envelope: &EncryptedPayloadEnvelope{
				Version:         999,
				AlgorithmID:     AlgorithmX25519XSalsa20Poly1305,
				RecipientKeyIDs: []string{"abc123def456"},
				Nonce:           make([]byte, XSalsa20NonceSize),
				Ciphertext:      []byte("encrypted data"),
				SenderPubKey:    make([]byte, X25519PublicKeySize),
			},
			expectErr: true,
			errMsg:    "not supported",
		},
		{
			name: "unsupported algorithm",
			envelope: &EncryptedPayloadEnvelope{
				Version:         1,
				AlgorithmID:     "UNKNOWN-ALGO",
				RecipientKeyIDs: []string{"abc123def456"},
				Nonce:           make([]byte, XSalsa20NonceSize),
				Ciphertext:      []byte("encrypted data"),
				SenderPubKey:    make([]byte, X25519PublicKeySize),
			},
			expectErr: true,
			errMsg:    "not supported",
		},
		{
			name: "no recipients",
			envelope: &EncryptedPayloadEnvelope{
				Version:         1,
				AlgorithmID:     AlgorithmX25519XSalsa20Poly1305,
				RecipientKeyIDs: []string{},
				Nonce:           make([]byte, XSalsa20NonceSize),
				Ciphertext:      []byte("encrypted data"),
				SenderPubKey:    make([]byte, X25519PublicKeySize),
			},
			expectErr: true,
			errMsg:    "at least one recipient",
		},
		{
			name: "empty nonce",
			envelope: &EncryptedPayloadEnvelope{
				Version:         1,
				AlgorithmID:     AlgorithmX25519XSalsa20Poly1305,
				RecipientKeyIDs: []string{"abc123def456"},
				Nonce:           []byte{},
				Ciphertext:      []byte("encrypted data"),
				SenderPubKey:    make([]byte, X25519PublicKeySize),
			},
			expectErr: true,
			errMsg:    "nonce cannot be empty",
		},
		{
			name: "wrong nonce size",
			envelope: &EncryptedPayloadEnvelope{
				Version:         1,
				AlgorithmID:     AlgorithmX25519XSalsa20Poly1305,
				RecipientKeyIDs: []string{"abc123def456"},
				Nonce:           make([]byte, 16), // Wrong size
				Ciphertext:      []byte("encrypted data"),
				SenderPubKey:    make([]byte, X25519PublicKeySize),
			},
			expectErr: true,
			errMsg:    "nonce size mismatch",
		},
		{
			name: "empty ciphertext",
			envelope: &EncryptedPayloadEnvelope{
				Version:         1,
				AlgorithmID:     AlgorithmX25519XSalsa20Poly1305,
				RecipientKeyIDs: []string{"abc123def456"},
				Nonce:           make([]byte, XSalsa20NonceSize),
				Ciphertext:      []byte{},
				SenderPubKey:    make([]byte, X25519PublicKeySize),
			},
			expectErr: true,
			errMsg:    "ciphertext cannot be empty",
		},
		{
			name: "empty sender public key",
			envelope: &EncryptedPayloadEnvelope{
				Version:         1,
				AlgorithmID:     AlgorithmX25519XSalsa20Poly1305,
				RecipientKeyIDs: []string{"abc123def456"},
				Nonce:           make([]byte, XSalsa20NonceSize),
				Ciphertext:      []byte("encrypted data"),
				SenderPubKey:    []byte{},
			},
			expectErr: true,
			errMsg:    "sender public key required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.envelope.Validate()
			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEncryptedPayloadEnvelope_SigningPayload(t *testing.T) {
	envelope := &EncryptedPayloadEnvelope{
		Version:         1,
		AlgorithmID:     AlgorithmX25519XSalsa20Poly1305,
		RecipientKeyIDs: []string{"key1", "key2"},
		Nonce:           []byte("test-nonce-24-bytes!"),
		Ciphertext:      []byte("test ciphertext"),
		SenderPubKey:    make([]byte, 32),
	}

	payload1 := envelope.SigningPayload()
	payload2 := envelope.SigningPayload()

	// Same envelope should produce same payload
	assert.Equal(t, payload1, payload2)

	// Different ciphertext should produce different payload
	envelope.Ciphertext = []byte("different ciphertext")
	payload3 := envelope.SigningPayload()
	assert.NotEqual(t, payload1, payload3)
}

func TestEncryptedPayloadEnvelope_Recipients(t *testing.T) {
	envelope := &EncryptedPayloadEnvelope{
		RecipientKeyIDs: []string{"key1", "key2", "key3"},
	}

	// Test GetRecipientIndex
	assert.Equal(t, 0, envelope.GetRecipientIndex("key1"))
	assert.Equal(t, 1, envelope.GetRecipientIndex("key2"))
	assert.Equal(t, 2, envelope.GetRecipientIndex("key3"))
	assert.Equal(t, -1, envelope.GetRecipientIndex("key4"))

	// Test IsRecipient
	assert.True(t, envelope.IsRecipient("key1"))
	assert.True(t, envelope.IsRecipient("key2"))
	assert.False(t, envelope.IsRecipient("unknown"))
}

func TestEncryptedPayloadEnvelope_Metadata(t *testing.T) {
	envelope := NewEncryptedPayloadEnvelope()

	// Add metadata
	err := envelope.AddMetadata("key1", "value1")
	require.NoError(t, err)

	// Retrieve metadata
	val, ok := envelope.GetMetadata("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", val)

	// Non-existent key
	_, ok = envelope.GetMetadata("nonexistent")
	assert.False(t, ok)

	// Empty key should error
	err = envelope.AddMetadata("", "value")
	require.Error(t, err)
}

func TestRecipientKeyRecord_Validate(t *testing.T) {
	tests := []struct {
		name      string
		record    RecipientKeyRecord
		expectErr bool
	}{
		{
			name: "valid record",
			record: RecipientKeyRecord{
				Address:        "cosmos1xyz...",
				PublicKey:      make([]byte, 32),
				KeyFingerprint: "abc123",
				AlgorithmID:    AlgorithmX25519XSalsa20Poly1305,
			},
			expectErr: false,
		},
		{
			name: "empty address",
			record: RecipientKeyRecord{
				Address:        "",
				PublicKey:      make([]byte, 32),
				KeyFingerprint: "abc123",
				AlgorithmID:    AlgorithmX25519XSalsa20Poly1305,
			},
			expectErr: true,
		},
		{
			name: "empty public key",
			record: RecipientKeyRecord{
				Address:        "cosmos1xyz...",
				PublicKey:      []byte{},
				KeyFingerprint: "abc123",
				AlgorithmID:    AlgorithmX25519XSalsa20Poly1305,
			},
			expectErr: true,
		},
		{
			name: "wrong public key size",
			record: RecipientKeyRecord{
				Address:        "cosmos1xyz...",
				PublicKey:      make([]byte, 16), // Wrong size
				KeyFingerprint: "abc123",
				AlgorithmID:    AlgorithmX25519XSalsa20Poly1305,
			},
			expectErr: true,
		},
		{
			name: "unsupported algorithm",
			record: RecipientKeyRecord{
				Address:        "cosmos1xyz...",
				PublicKey:      make([]byte, 32),
				KeyFingerprint: "abc123",
				AlgorithmID:    "UNKNOWN",
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.record.Validate()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRecipientKeyRecord_IsActive(t *testing.T) {
	// Active key
	activeKey := RecipientKeyRecord{RevokedAt: 0}
	assert.True(t, activeKey.IsActive())

	// Revoked key
	revokedKey := RecipientKeyRecord{RevokedAt: 1234567890}
	assert.False(t, revokedKey.IsActive())
}

func TestComputeKeyFingerprint(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	key2[0] = 1 // Different key

	fp1 := ComputeKeyFingerprint(key1)
	fp2 := ComputeKeyFingerprint(key2)

	// Same key should produce same fingerprint
	assert.Equal(t, fp1, ComputeKeyFingerprint(key1))

	// Different keys should produce different fingerprints
	assert.NotEqual(t, fp1, fp2)

	// Fingerprint should be hex string of expected length
	assert.Len(t, fp1, KeyFingerprintSize*2) // hex encoding doubles length
}
