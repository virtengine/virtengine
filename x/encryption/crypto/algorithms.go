package crypto

import (
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/nacl/box"

	"pkg.akt.dev/node/x/encryption/types"
)

// Algorithm represents an encryption algorithm implementation
type Algorithm interface {
	// ID returns the algorithm identifier
	ID() string

	// KeySize returns the expected public key size in bytes
	KeySize() int

	// NonceSize returns the expected nonce size in bytes
	NonceSize() int

	// GenerateKeyPair generates a new key pair
	GenerateKeyPair() (*KeyPair, error)

	// Encrypt encrypts plaintext to a recipient
	Encrypt(plaintext []byte, recipientPubKey []byte, senderKeyPair *KeyPair) (ciphertext []byte, nonce []byte, err error)

	// Decrypt decrypts ciphertext from a sender
	Decrypt(ciphertext []byte, nonce []byte, senderPubKey []byte, recipientPrivateKey []byte) (plaintext []byte, err error)
}

// X25519XSalsa20Poly1305 implements the X25519-XSalsa20-Poly1305 algorithm
type X25519XSalsa20Poly1305 struct{}

// NewX25519XSalsa20Poly1305 creates a new instance of the algorithm
func NewX25519XSalsa20Poly1305() *X25519XSalsa20Poly1305 {
	return &X25519XSalsa20Poly1305{}
}

// ID returns the algorithm identifier
func (a *X25519XSalsa20Poly1305) ID() string {
	return types.AlgorithmX25519XSalsa20Poly1305
}

// KeySize returns the expected public key size
func (a *X25519XSalsa20Poly1305) KeySize() int {
	return types.X25519PublicKeySize
}

// NonceSize returns the expected nonce size
func (a *X25519XSalsa20Poly1305) NonceSize() int {
	return types.XSalsa20NonceSize
}

// GenerateKeyPair generates a new X25519 key pair
func (a *X25519XSalsa20Poly1305) GenerateKeyPair() (*KeyPair, error) {
	return GenerateKeyPair()
}

// Encrypt encrypts plaintext to a recipient using NaCl box
func (a *X25519XSalsa20Poly1305) Encrypt(plaintext []byte, recipientPubKey []byte, senderKeyPair *KeyPair) ([]byte, []byte, error) {
	if len(recipientPubKey) != a.KeySize() {
		return nil, nil, fmt.Errorf("invalid recipient public key size: expected %d, got %d", a.KeySize(), len(recipientPubKey))
	}

	// Generate nonce
	nonce, err := GenerateNonce()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Convert recipient public key
	var recipientPubKeyArr [32]byte
	copy(recipientPubKeyArr[:], recipientPubKey)

	// Encrypt
	ciphertext := box.Seal(nil, plaintext, &nonce, &recipientPubKeyArr, &senderKeyPair.PrivateKey)

	return ciphertext, nonce[:], nil
}

// Decrypt decrypts ciphertext from a sender using NaCl box
func (a *X25519XSalsa20Poly1305) Decrypt(ciphertext []byte, nonce []byte, senderPubKey []byte, recipientPrivateKey []byte) ([]byte, error) {
	if len(nonce) != a.NonceSize() {
		return nil, fmt.Errorf("invalid nonce size: expected %d, got %d", a.NonceSize(), len(nonce))
	}

	if len(senderPubKey) != a.KeySize() {
		return nil, fmt.Errorf("invalid sender public key size: expected %d, got %d", a.KeySize(), len(senderPubKey))
	}

	if len(recipientPrivateKey) != types.X25519PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: expected %d, got %d", types.X25519PrivateKeySize, len(recipientPrivateKey))
	}

	// Convert to arrays
	var nonceArr [24]byte
	copy(nonceArr[:], nonce)

	var senderPubKeyArr [32]byte
	copy(senderPubKeyArr[:], senderPubKey)

	var privateKeyArr [32]byte
	copy(privateKeyArr[:], recipientPrivateKey)

	// Decrypt
	plaintext, ok := box.Open(nil, ciphertext, &nonceArr, &senderPubKeyArr, &privateKeyArr)
	if !ok {
		return nil, fmt.Errorf("decryption failed: authentication error")
	}

	return plaintext, nil
}

// GetAlgorithm returns an algorithm implementation by ID
func GetAlgorithm(algorithmID string) (Algorithm, error) {
	switch algorithmID {
	case types.AlgorithmX25519XSalsa20Poly1305:
		return NewX25519XSalsa20Poly1305(), nil
	case types.AlgorithmAgeX25519:
		return nil, fmt.Errorf("AGE-X25519 algorithm not yet implemented")
	default:
		return nil, fmt.Errorf("unknown algorithm: %s", algorithmID)
	}
}

// DeriveSharedSecret derives a shared secret from a private key and peer's public key
// using X25519 key agreement
func DeriveSharedSecret(privateKey, peerPublicKey []byte) ([]byte, error) {
	if len(privateKey) != 32 {
		return nil, fmt.Errorf("invalid private key size: expected 32, got %d", len(privateKey))
	}
	if len(peerPublicKey) != 32 {
		return nil, fmt.Errorf("invalid public key size: expected 32, got %d", len(peerPublicKey))
	}

	var privateKeyArr [32]byte
	copy(privateKeyArr[:], privateKey)

	var peerPubKeyArr [32]byte
	copy(peerPubKeyArr[:], peerPublicKey)

	sharedSecret, err := curve25519.X25519(privateKeyArr[:], peerPubKeyArr[:])
	if err != nil {
		return nil, fmt.Errorf("key exchange failed: %w", err)
	}

	return sharedSecret, nil
}

// PrecomputeSharedKey precomputes a shared key for box operations
// This can be used to speed up multiple encryptions/decryptions with the same peer
func PrecomputeSharedKey(privateKey, peerPublicKey []byte) ([32]byte, error) {
	var sharedKey [32]byte

	if len(privateKey) != 32 {
		return sharedKey, fmt.Errorf("invalid private key size: expected 32, got %d", len(privateKey))
	}
	if len(peerPublicKey) != 32 {
		return sharedKey, fmt.Errorf("invalid public key size: expected 32, got %d", len(peerPublicKey))
	}

	var privateKeyArr [32]byte
	copy(privateKeyArr[:], privateKey)

	var peerPubKeyArr [32]byte
	copy(peerPubKeyArr[:], peerPublicKey)

	box.Precompute(&sharedKey, &peerPubKeyArr, &privateKeyArr)

	return sharedKey, nil
}

// EncryptWithSharedKey encrypts using a precomputed shared key
func EncryptWithSharedKey(plaintext []byte, sharedKey *[32]byte) (ciphertext []byte, nonce [24]byte, err error) {
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nil, nonce, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext = box.SealAfterPrecomputation(nil, plaintext, &nonce, sharedKey)
	return ciphertext, nonce, nil
}

// DecryptWithSharedKey decrypts using a precomputed shared key
func DecryptWithSharedKey(ciphertext []byte, nonce *[24]byte, sharedKey *[32]byte) ([]byte, error) {
	plaintext, ok := box.OpenAfterPrecomputation(nil, ciphertext, nonce, sharedKey)
	if !ok {
		return nil, fmt.Errorf("decryption failed")
	}
	return plaintext, nil
}

// ZeroBytes securely zeros a byte slice
func ZeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

// ZeroKey securely zeros a 32-byte key array
func ZeroKey(key *[32]byte) {
	for i := range key {
		key[i] = 0
	}
}
