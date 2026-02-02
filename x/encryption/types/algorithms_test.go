package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSupportedAlgorithms(t *testing.T) {
	algorithms := SupportedAlgorithms()

	assert.NotEmpty(t, algorithms)
	assert.Contains(t, algorithms, AlgorithmX25519XSalsa20Poly1305)
}

func TestIsAlgorithmSupported(t *testing.T) {
	tests := []struct {
		algorithm string
		expected  bool
	}{
		{AlgorithmX25519XSalsa20Poly1305, true},
		{"UNKNOWN-ALGO", false},
		{"", false},
		{"aes-256-gcm", false},
	}

	for _, tc := range tests {
		t.Run(tc.algorithm, func(t *testing.T) {
			result := IsAlgorithmSupported(tc.algorithm)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestDefaultAlgorithm(t *testing.T) {
	alg := DefaultAlgorithm()

	assert.Equal(t, AlgorithmX25519XSalsa20Poly1305, alg)
	assert.True(t, IsAlgorithmSupported(alg))
}

func TestGetAlgorithmInfo(t *testing.T) {
	tests := []struct {
		algorithm  string
		expectErr  bool
		keySize    int
		nonceSize  int
		deprecated bool
	}{
		{
			algorithm:  AlgorithmX25519XSalsa20Poly1305,
			expectErr:  false,
			keySize:    32,
			nonceSize:  24,
			deprecated: false,
		},
		{
			algorithm:  AlgorithmAgeX25519,
			expectErr:  false,
			keySize:    32,
			nonceSize:  16,
			deprecated: false,
		},
		{
			algorithm: "UNKNOWN",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.algorithm, func(t *testing.T) {
			info, err := GetAlgorithmInfo(tc.algorithm)

			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.algorithm, info.ID)
			assert.Equal(t, tc.keySize, info.KeySize)
			assert.Equal(t, tc.nonceSize, info.NonceSize)
			assert.Equal(t, tc.deprecated, info.Deprecated)
			assert.NotEmpty(t, info.Description)
		})
	}
}

func TestValidateAlgorithmParams(t *testing.T) {
	tests := []struct {
		name      string
		algorithm string
		pubKey    []byte
		nonce     []byte
		expectErr bool
	}{
		{
			name:      "valid X25519 params",
			algorithm: AlgorithmX25519XSalsa20Poly1305,
			pubKey:    make([]byte, 32),
			nonce:     make([]byte, 24),
			expectErr: false,
		},
		{
			name:      "wrong public key size",
			algorithm: AlgorithmX25519XSalsa20Poly1305,
			pubKey:    make([]byte, 16),
			nonce:     make([]byte, 24),
			expectErr: true,
		},
		{
			name:      "wrong nonce size",
			algorithm: AlgorithmX25519XSalsa20Poly1305,
			pubKey:    make([]byte, 32),
			nonce:     make([]byte, 16),
			expectErr: true,
		},
		{
			name:      "unknown algorithm",
			algorithm: "UNKNOWN",
			pubKey:    make([]byte, 32),
			nonce:     make([]byte, 24),
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateAlgorithmParams(tc.algorithm, tc.pubKey, tc.nonce)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAlgorithmConstants(t *testing.T) {
	// Verify expected constant values
	assert.Equal(t, 32, X25519PublicKeySize)
	assert.Equal(t, 32, X25519PrivateKeySize)
	assert.Equal(t, 24, XSalsa20NonceSize)
	assert.Equal(t, 16, Poly1305TagSize)
	assert.Equal(t, 20, KeyFingerprintSize)
}
