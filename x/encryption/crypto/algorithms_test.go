package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pkg.akt.dev/node/x/encryption/types"
)

func TestX25519XSalsa20Poly1305_ID(t *testing.T) {
	alg := NewX25519XSalsa20Poly1305()

	assert.Equal(t, types.AlgorithmX25519XSalsa20Poly1305, alg.ID())
}

func TestX25519XSalsa20Poly1305_KeySize(t *testing.T) {
	alg := NewX25519XSalsa20Poly1305()

	assert.Equal(t, types.X25519PublicKeySize, alg.KeySize())
}

func TestX25519XSalsa20Poly1305_NonceSize(t *testing.T) {
	alg := NewX25519XSalsa20Poly1305()

	assert.Equal(t, types.XSalsa20NonceSize, alg.NonceSize())
}

func TestX25519XSalsa20Poly1305_EncryptDecrypt(t *testing.T) {
	alg := NewX25519XSalsa20Poly1305()

	sender, err := alg.GenerateKeyPair()
	require.NoError(t, err)

	recipient, err := alg.GenerateKeyPair()
	require.NoError(t, err)

	plaintext := []byte("Hello, encrypted world!")

	// Encrypt
	ciphertext, nonce, err := alg.Encrypt(plaintext, recipient.PublicKey[:], sender)
	require.NoError(t, err)
	require.NotEmpty(t, ciphertext)
	require.Len(t, nonce, alg.NonceSize())

	// Ciphertext should be different from plaintext
	assert.NotEqual(t, plaintext, ciphertext)

	// Decrypt
	decrypted, err := alg.Decrypt(ciphertext, nonce, sender.PublicKey[:], recipient.PrivateKey[:])
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestX25519XSalsa20Poly1305_Encrypt_InvalidKeySize(t *testing.T) {
	alg := NewX25519XSalsa20Poly1305()

	sender, err := alg.GenerateKeyPair()
	require.NoError(t, err)

	plaintext := []byte("test")

	// Wrong recipient key size
	_, _, err = alg.Encrypt(plaintext, make([]byte, 16), sender)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid recipient public key size")
}

func TestX25519XSalsa20Poly1305_Decrypt_InvalidNonceSize(t *testing.T) {
	alg := NewX25519XSalsa20Poly1305()

	sender, err := alg.GenerateKeyPair()
	require.NoError(t, err)

	recipient, err := alg.GenerateKeyPair()
	require.NoError(t, err)

	// Wrong nonce size
	_, err = alg.Decrypt([]byte("ciphertext"), make([]byte, 16), sender.PublicKey[:], recipient.PrivateKey[:])
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid nonce size")
}

func TestX25519XSalsa20Poly1305_Decrypt_InvalidKeySize(t *testing.T) {
	alg := NewX25519XSalsa20Poly1305()

	sender, err := alg.GenerateKeyPair()
	require.NoError(t, err)

	plaintext := []byte("test")

	ciphertext, nonce, err := alg.Encrypt(plaintext, sender.PublicKey[:], sender)
	require.NoError(t, err)

	// Wrong sender public key size
	_, err = alg.Decrypt(ciphertext, nonce, make([]byte, 16), sender.PrivateKey[:])
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sender public key size")

	// Wrong private key size
	_, err = alg.Decrypt(ciphertext, nonce, sender.PublicKey[:], make([]byte, 16))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid private key size")
}

func TestX25519XSalsa20Poly1305_Decrypt_TamperedCiphertext(t *testing.T) {
	alg := NewX25519XSalsa20Poly1305()

	sender, err := alg.GenerateKeyPair()
	require.NoError(t, err)

	recipient, err := alg.GenerateKeyPair()
	require.NoError(t, err)

	plaintext := []byte("test message")

	ciphertext, nonce, err := alg.Encrypt(plaintext, recipient.PublicKey[:], sender)
	require.NoError(t, err)

	// Tamper with ciphertext
	ciphertext[0] ^= 0xFF

	_, err = alg.Decrypt(ciphertext, nonce, sender.PublicKey[:], recipient.PrivateKey[:])
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authentication error")
}

func TestGetAlgorithm(t *testing.T) {
	tests := []struct {
		algorithmID string
		expectErr   bool
	}{
		{types.AlgorithmX25519XSalsa20Poly1305, false},
		{types.AlgorithmAgeX25519, true}, // Not yet implemented
		{"UNKNOWN", true},
	}

	for _, tc := range tests {
		t.Run(tc.algorithmID, func(t *testing.T) {
			alg, err := GetAlgorithm(tc.algorithmID)

			if tc.expectErr {
				require.Error(t, err)
				assert.Nil(t, alg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, alg)
				assert.Equal(t, tc.algorithmID, alg.ID())
			}
		})
	}
}

func TestDeriveSharedSecret(t *testing.T) {
	kp1, err := GenerateKeyPair()
	require.NoError(t, err)

	kp2, err := GenerateKeyPair()
	require.NoError(t, err)

	// Both sides should derive the same shared secret
	secret1, err := DeriveSharedSecret(kp1.PrivateKey[:], kp2.PublicKey[:])
	require.NoError(t, err)
	require.NotEmpty(t, secret1)

	secret2, err := DeriveSharedSecret(kp2.PrivateKey[:], kp1.PublicKey[:])
	require.NoError(t, err)

	assert.Equal(t, secret1, secret2)
}

func TestDeriveSharedSecret_InvalidKeySize(t *testing.T) {
	kp, err := GenerateKeyPair()
	require.NoError(t, err)

	// Invalid private key size
	_, err = DeriveSharedSecret(make([]byte, 16), kp.PublicKey[:])
	require.Error(t, err)

	// Invalid public key size
	_, err = DeriveSharedSecret(kp.PrivateKey[:], make([]byte, 16))
	require.Error(t, err)
}

func TestPrecomputeSharedKey(t *testing.T) {
	kp1, err := GenerateKeyPair()
	require.NoError(t, err)

	kp2, err := GenerateKeyPair()
	require.NoError(t, err)

	// Precompute shared key
	sharedKey, err := PrecomputeSharedKey(kp1.PrivateKey[:], kp2.PublicKey[:])
	require.NoError(t, err)
	assert.NotEqual(t, [32]byte{}, sharedKey)
}

func TestEncryptDecryptWithSharedKey(t *testing.T) {
	kp1, err := GenerateKeyPair()
	require.NoError(t, err)

	kp2, err := GenerateKeyPair()
	require.NoError(t, err)

	// Both parties compute the same shared key
	sharedKey1, err := PrecomputeSharedKey(kp1.PrivateKey[:], kp2.PublicKey[:])
	require.NoError(t, err)

	sharedKey2, err := PrecomputeSharedKey(kp2.PrivateKey[:], kp1.PublicKey[:])
	require.NoError(t, err)

	assert.Equal(t, sharedKey1, sharedKey2)

	// Encrypt with shared key
	plaintext := []byte("message encrypted with shared key")
	ciphertext, nonce, err := EncryptWithSharedKey(plaintext, &sharedKey1)
	require.NoError(t, err)

	// Decrypt with same shared key
	decrypted, err := DecryptWithSharedKey(ciphertext, &nonce, &sharedKey2)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestZeroBytes(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	ZeroBytes(data)

	for _, b := range data {
		assert.Equal(t, byte(0), b)
	}
}

func TestZeroKey(t *testing.T) {
	key := [32]byte{1, 2, 3, 4, 5}
	ZeroKey(&key)

	for _, b := range key {
		assert.Equal(t, byte(0), b)
	}
}
