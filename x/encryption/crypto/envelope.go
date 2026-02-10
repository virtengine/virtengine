// Package crypto provides encryption and decryption helpers for use with the
// VirtEngine encryption module. These functions are designed for OFF-CHAIN use
// by clients that need to create encrypted envelopes or decrypt received data.
//
// SECURITY NOTICE:
// - Never store private keys on-chain
// - Use crypto/rand for all random generation
// - Nonces must be unique per encryption
// - This package does not persist any key material
package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/nacl/box"

	"github.com/virtengine/virtengine/x/encryption/types"
)

// KeyPair represents an X25519 key pair for encryption
type KeyPair struct {
	PublicKey  [32]byte
	PrivateKey [32]byte
}

// RecipientInfo describes a recipient for multi-recipient envelopes.
type RecipientInfo struct {
	PublicKey  []byte
	KeyID      string
	KeyVersion uint32
}

// GenerateKeyPair generates a new X25519 key pair using crypto/rand
func GenerateKeyPair() (*KeyPair, error) {
	var privateKey [32]byte
	if _, err := io.ReadFull(rand.Reader, privateKey[:]); err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	var publicKey [32]byte
	curve25519.ScalarBaseMult(&publicKey, &privateKey)

	return &KeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}, nil
}

// Fingerprint returns the key fingerprint for this key pair
func (kp *KeyPair) Fingerprint() string {
	return types.ComputeKeyFingerprint(kp.PublicKey[:])
}

// GenerateNonce generates a random 24-byte nonce for XSalsa20
func GenerateNonce() ([24]byte, error) {
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nonce, fmt.Errorf("failed to generate nonce: %w", err)
	}
	return nonce, nil
}

// CreateEnvelope creates an encrypted payload envelope for a single recipient
// using X25519-XSalsa20-Poly1305 (NaCl box).
//
// Parameters:
//   - plaintext: The data to encrypt
//   - recipientPublicKey: The recipient's X25519 public key (32 bytes)
//   - senderKeyPair: The sender's key pair for ephemeral key exchange
//
// Returns the encrypted envelope ready for storage on-chain.
func CreateEnvelope(plaintext []byte, recipientPublicKey []byte, senderKeyPair *KeyPair) (*types.EncryptedPayloadEnvelope, error) {
	return CreateEnvelopeWithRecipient(plaintext, RecipientInfo{PublicKey: recipientPublicKey}, senderKeyPair)
}

// CreateEnvelopeWithRecipient creates an encrypted payload envelope for a single recipient
// with an optional versioned key ID.
func CreateEnvelopeWithRecipient(plaintext []byte, recipient RecipientInfo, senderKeyPair *KeyPair) (*types.EncryptedPayloadEnvelope, error) {
	if len(recipient.PublicKey) == 0 {
		return nil, fmt.Errorf("recipient public key required")
	}
	if len(recipient.PublicKey) != 32 {
		return nil, fmt.Errorf("invalid recipient public key size: expected 32, got %d", len(recipient.PublicKey))
	}

	// Generate nonce
	nonce, err := GenerateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Convert recipient public key to array
	var recipientPubKeyArr [32]byte
	copy(recipientPubKeyArr[:], recipient.PublicKey)

	// Encrypt using NaCl box
	ciphertext := box.Seal(nil, plaintext, &nonce, &recipientPubKeyArr, &senderKeyPair.PrivateKey)

	// Compute recipient key fingerprint
	recipientFingerprint := types.ComputeKeyFingerprint(recipient.PublicKey)
	recipientKeyID := recipient.KeyID
	if recipientKeyID == "" {
		recipientKeyID = types.FormatRecipientKeyID(recipientFingerprint, recipient.KeyVersion)
	}

	// Create envelope
	envelope := &types.EncryptedPayloadEnvelope{
		Version:             types.EnvelopeVersion,
		AlgorithmID:         types.AlgorithmX25519XSalsa20Poly1305,
		AlgorithmVersion:    types.AlgorithmVersionV1,
		RecipientKeyIDs:     []string{recipientKeyID},
		RecipientPublicKeys: [][]byte{append([]byte(nil), recipient.PublicKey...)},
		Nonce:               nonce[:],
		Ciphertext:          ciphertext,
		SenderPubKey:        senderKeyPair.PublicKey[:],
		Metadata:            make(map[string]string),
	}

	// Generate signature over the signing payload
	signature, err := signEnvelope(envelope, &senderKeyPair.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign envelope: %w", err)
	}
	envelope.SenderSignature = signature

	return envelope, nil
}

// CreateMultiRecipientEnvelope creates an encrypted payload envelope for multiple recipients.
// This uses a symmetric data encryption key (DEK) that is encrypted separately for each recipient.
//
// Parameters:
//   - plaintext: The data to encrypt
//   - recipientPublicKeys: List of recipient X25519 public keys (each 32 bytes)
//   - senderKeyPair: The sender's key pair
//
// Returns the encrypted envelope with separate encrypted keys for each recipient.
func CreateMultiRecipientEnvelope(plaintext []byte, recipientPublicKeys [][]byte, senderKeyPair *KeyPair) (*types.EncryptedPayloadEnvelope, error) {
	if len(recipientPublicKeys) == 0 {
		return nil, fmt.Errorf("at least one recipient required")
	}

	// For single recipient, use simple box
	if len(recipientPublicKeys) == 1 {
		return CreateEnvelope(plaintext, recipientPublicKeys[0], senderKeyPair)
	}

	recipients := make([]RecipientInfo, len(recipientPublicKeys))
	for i, pubKey := range recipientPublicKeys {
		recipients[i] = RecipientInfo{PublicKey: pubKey}
	}

	return CreateMultiRecipientEnvelopeWithRecipients(plaintext, recipients, senderKeyPair)
}

// CreateMultiRecipientEnvelopeWithRecipients creates an encrypted payload envelope for multiple recipients
// using optional versioned key IDs.
func CreateMultiRecipientEnvelopeWithRecipients(plaintext []byte, recipients []RecipientInfo, senderKeyPair *KeyPair) (*types.EncryptedPayloadEnvelope, error) {
	if len(recipients) == 0 {
		return nil, fmt.Errorf("at least one recipient required")
	}

	if len(recipients) == 1 {
		return CreateEnvelopeWithRecipient(plaintext, recipients[0], senderKeyPair)
	}

	// Generate a random Data Encryption Key (DEK)
	var dek [32]byte
	if _, err := io.ReadFull(rand.Reader, dek[:]); err != nil {
		return nil, fmt.Errorf("failed to generate DEK: %w", err)
	}

	// Generate nonce for data encryption
	dataNonce, err := GenerateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate data nonce: %w", err)
	}

	// Encrypt data with DEK using XSalsa20-Poly1305
	// For multi-recipient, we use secretbox-style encryption with the DEK
	ciphertext := xsalsa20Poly1305Encrypt(plaintext, &dek, &dataNonce)

	// Encrypt DEK for each recipient
	recipientKeyIDs := make([]string, len(recipients))
	encryptedKeys := make([][]byte, len(recipients))
	recipientPubKeys := make([][]byte, len(recipients))
	wrappedKeys := make([]types.WrappedKeyEntry, len(recipients))

	for i, recipient := range recipients {
		recipientPubKey := recipient.PublicKey
		if len(recipientPubKey) != 32 {
			return nil, fmt.Errorf("invalid recipient public key size at index %d: expected 32, got %d", i, len(recipientPubKey))
		}

		var recipientPubKeyArr [32]byte
		copy(recipientPubKeyArr[:], recipientPubKey)

		// Generate unique nonce for key encryption
		keyNonce, err := GenerateNonce()
		if err != nil {
			return nil, fmt.Errorf("failed to generate key nonce for recipient %d: %w", i, err)
		}

		// Encrypt DEK to this recipient
		encryptedDEK := box.Seal(keyNonce[:], dek[:], &keyNonce, &recipientPubKeyArr, &senderKeyPair.PrivateKey)
		encryptedKeys[i] = encryptedDEK

		fingerprint := types.ComputeKeyFingerprint(recipientPubKey)
		keyID := recipient.KeyID
		if keyID == "" {
			keyID = types.FormatRecipientKeyID(fingerprint, recipient.KeyVersion)
		}
		recipientKeyIDs[i] = keyID
		recipientPubKeys[i] = append([]byte(nil), recipientPubKey...)
		wrappedKeys[i] = types.WrappedKeyEntry{
			RecipientID: keyID,
			WrappedKey:  encryptedDEK,
		}
	}

	// Create envelope
	envelope := &types.EncryptedPayloadEnvelope{
		Version:             types.EnvelopeVersion,
		AlgorithmID:         types.AlgorithmX25519XSalsa20Poly1305,
		AlgorithmVersion:    types.AlgorithmVersionV1,
		RecipientKeyIDs:     recipientKeyIDs,
		RecipientPublicKeys: recipientPubKeys,
		EncryptedKeys:       encryptedKeys,
		WrappedKeys:         wrappedKeys,
		Nonce:               dataNonce[:],
		Ciphertext:          ciphertext,
		SenderPubKey:        senderKeyPair.PublicKey[:],
		Metadata:            make(map[string]string),
	}

	// Add metadata to indicate multi-recipient mode
	envelope.Metadata["_mode"] = "multi-recipient"

	// Generate signature
	signature, err := signEnvelope(envelope, &senderKeyPair.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign envelope: %w", err)
	}
	envelope.SenderSignature = signature

	// Clear DEK from memory
	for i := range dek {
		dek[i] = 0
	}

	return envelope, nil
}

// OpenEnvelope decrypts an encrypted payload envelope using the recipient's private key.
//
// Parameters:
//   - envelope: The encrypted envelope to decrypt
//   - recipientPrivateKey: The recipient's X25519 private key (32 bytes)
//
// Returns the decrypted plaintext.
func OpenEnvelope(envelope *types.EncryptedPayloadEnvelope, recipientPrivateKey []byte) ([]byte, error) {
	if envelope == nil {
		return nil, types.ErrInvalidEnvelope.Wrap("envelope cannot be nil")
	}

	if err := envelope.Validate(); err != nil {
		return nil, err
	}

	if len(recipientPrivateKey) != 32 {
		return nil, fmt.Errorf("invalid private key size: expected 32, got %d", len(recipientPrivateKey))
	}

	// Check algorithm
	if envelope.AlgorithmID != types.AlgorithmX25519XSalsa20Poly1305 {
		return nil, types.ErrUnsupportedAlgorithm.Wrapf("cannot decrypt %s", envelope.AlgorithmID)
	}

	// Convert keys to arrays
	var privateKeyArr [32]byte
	copy(privateKeyArr[:], recipientPrivateKey)

	var senderPubKeyArr [32]byte
	copy(senderPubKeyArr[:], envelope.SenderPubKey)

	var nonce [24]byte
	copy(nonce[:], envelope.Nonce)

	// Check if multi-recipient mode
	if mode, ok := envelope.Metadata["_mode"]; ok && mode == "multi-recipient" {
		return openMultiRecipientEnvelope(envelope, &privateKeyArr, &senderPubKeyArr)
	}

	// Single recipient: direct box.Open
	plaintext, ok := box.Open(nil, envelope.Ciphertext, &nonce, &senderPubKeyArr, &privateKeyArr)
	if !ok {
		return nil, types.ErrDecryptionFailed.Wrap("failed to decrypt envelope")
	}

	return plaintext, nil
}

// openMultiRecipientEnvelope decrypts a multi-recipient envelope
func openMultiRecipientEnvelope(envelope *types.EncryptedPayloadEnvelope, recipientPrivateKey, senderPubKey *[32]byte) ([]byte, error) {
	// Derive our public key to find our encrypted key
	var ourPublicKey [32]byte
	curve25519.ScalarBaseMult(&ourPublicKey, recipientPrivateKey)
	ourFingerprint := types.ComputeKeyFingerprint(ourPublicKey[:])

	// Find our encrypted DEK
	var encryptedDEK []byte
	for _, entry := range envelope.WrappedKeys {
		if types.NormalizeRecipientKeyID(entry.RecipientID) == ourFingerprint {
			encryptedDEK = entry.WrappedKey
			break
		}
	}
	if encryptedDEK == nil {
		for i, keyID := range envelope.RecipientKeyIDs {
			if types.NormalizeRecipientKeyID(keyID) == ourFingerprint {
				if i < len(envelope.EncryptedKeys) {
					encryptedDEK = envelope.EncryptedKeys[i]
				}
				break
			}
		}
	}

	if encryptedDEK == nil {
		return nil, types.ErrNotRecipient.Wrap("no encrypted key found for this recipient")
	}

	// Extract nonce from encrypted DEK (first 24 bytes)
	if len(encryptedDEK) < 24 {
		return nil, types.ErrDecryptionFailed.Wrap("encrypted key too short")
	}

	var keyNonce [24]byte
	copy(keyNonce[:], encryptedDEK[:24])

	// Decrypt DEK
	dek, ok := box.Open(nil, encryptedDEK[24:], &keyNonce, senderPubKey, recipientPrivateKey)
	if !ok {
		return nil, types.ErrDecryptionFailed.Wrap("failed to decrypt data encryption key")
	}

	if len(dek) != 32 {
		return nil, types.ErrDecryptionFailed.Wrap("invalid DEK size")
	}

	// Decrypt data with DEK
	var dekArr [32]byte
	copy(dekArr[:], dek)

	var dataNonce [24]byte
	copy(dataNonce[:], envelope.Nonce)

	plaintext, err := xsalsa20Poly1305Decrypt(envelope.Ciphertext, &dekArr, &dataNonce)
	if err != nil {
		return nil, types.ErrDecryptionFailed.Wrap(err.Error())
	}

	// Clear DEK from memory
	for i := range dekArr {
		dekArr[i] = 0
	}

	return plaintext, nil
}

// ValidateEnvelopeSignature verifies the sender's signature on an envelope.
// Note: This is a simplified signature scheme using the signing payload hash.
// In production, consider using Ed25519 for signatures.
func ValidateEnvelopeSignature(envelope *types.EncryptedPayloadEnvelope) (bool, error) {
	if envelope == nil {
		return false, types.ErrInvalidEnvelope.Wrap("envelope cannot be nil")
	}

	if len(envelope.SenderSignature) == 0 {
		return false, types.ErrInvalidSignature.Wrap("no signature present")
	}

	if len(envelope.SenderPubKey) != 32 {
		return false, types.ErrInvalidPublicKey.Wrap("invalid sender public key")
	}

	// Compute expected signature
	payload := envelope.SigningPayload()
	expectedSig := computeSignature(payload, envelope.SenderPubKey)

	// Compare signatures
	if len(envelope.SenderSignature) != len(expectedSig) {
		return false, nil
	}

	// Constant-time comparison
	var diff byte
	for i := range expectedSig {
		diff |= envelope.SenderSignature[i] ^ expectedSig[i]
	}

	return diff == 0, nil
}

// signEnvelope creates a signature binding for the envelope.
//
// SECURITY WARNING (SECURITY-001):
// This implementation uses a SHA256 binding (H(payload || publicKey)) which provides
// integrity binding but NOT authentication. Anyone with knowledge of the public key
// can compute the same binding value.
//
// REMEDIATION REQUIRED: Replace with Ed25519 signature using sender's private key:
//  1. Change signature to: Ed25519.Sign(privateKey, payload)
//  2. Update CreateEnvelope to accept signing key
//  3. Update ValidateEnvelopeSignature to use Ed25519.Verify
//
// Current limitations:
//   - No proof of sender identity (anyone can forge "signatures")
//   - Suitable only when sender authentication is not required
//   - Must NOT be used for transaction authorization
//
//nolint:unparam // result 1 (error) reserved for future signing failures
func signEnvelope(envelope *types.EncryptedPayloadEnvelope, _ *[32]byte) ([]byte, error) {
	payload := envelope.SigningPayload()

	// Create binding: H(payload || publicKey)
	// WARNING: This is NOT a true signature - see security note above
	h := sha256.New()
	h.Write(payload)
	h.Write(envelope.SenderPubKey)

	return h.Sum(nil), nil
}

// computeSignature computes the expected signature for verification.
//
// SECURITY WARNING: See signEnvelope security notes. This function only verifies
// integrity binding, not sender authentication.
func computeSignature(payload, senderPubKey []byte) []byte {
	// WARNING: This is a simplified scheme - does not verify sender identity
	h := sha256.New()
	h.Write(payload)
	h.Write(senderPubKey)
	return h.Sum(nil)
}

// xsalsa20Poly1305Encrypt encrypts data using XSalsa20-Poly1305 with a symmetric key
func xsalsa20Poly1305Encrypt(plaintext []byte, key *[32]byte, nonce *[24]byte) []byte {
	// Use box.SealAfterPrecomputation with a zero peer key for symmetric encryption
	// This is a simplified approach; for production, use secretbox
	var zeroKey [32]byte
	var sharedKey [32]byte
	box.Precompute(&sharedKey, &zeroKey, key)

	return box.SealAfterPrecomputation(nil, plaintext, nonce, &sharedKey)
}

// xsalsa20Poly1305Decrypt decrypts data using XSalsa20-Poly1305 with a symmetric key
func xsalsa20Poly1305Decrypt(ciphertext []byte, key *[32]byte, nonce *[24]byte) ([]byte, error) {
	var zeroKey [32]byte
	var sharedKey [32]byte
	box.Precompute(&sharedKey, &zeroKey, key)

	plaintext, ok := box.OpenAfterPrecomputation(nil, ciphertext, nonce, &sharedKey)
	if !ok {
		return nil, fmt.Errorf("decryption failed")
	}

	return plaintext, nil
}
