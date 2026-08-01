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
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"

	"github.com/virtengine/virtengine/x/encryption/types"
)

// KeyPair contains independent X25519 encryption and Ed25519 signing keys.
type KeyPair struct {
	PublicKey         [32]byte
	PrivateKey        [32]byte
	SigningPublicKey  [ed25519.PublicKeySize]byte
	SigningPrivateKey [ed25519.PrivateKeySize]byte
}

// RecipientInfo describes a recipient for multi-recipient envelopes.
type RecipientInfo struct {
	PublicKey  []byte
	KeyID      string
	KeyVersion uint32
}

func (kp *KeyPair) validate() error {
	if kp == nil {
		return fmt.Errorf("sender key pair required")
	}
	var derivedPublicKey [32]byte
	curve25519.ScalarBaseMult(&derivedPublicKey, &kp.PrivateKey)
	if derivedPublicKey != kp.PublicKey {
		return fmt.Errorf("sender X25519 public key does not match private key")
	}
	derivedSigningPublicKey := ed25519.PrivateKey(kp.SigningPrivateKey[:]).Public().(ed25519.PublicKey)
	if !bytes.Equal(derivedSigningPublicKey, kp.SigningPublicKey[:]) {
		return fmt.Errorf("sender Ed25519 public key does not match private key")
	}
	return nil
}

func validateRecipient(recipient RecipientInfo) (string, error) {
	if len(recipient.PublicKey) != 32 {
		return "", fmt.Errorf("invalid recipient public key size: expected 32, got %d", len(recipient.PublicKey))
	}
	testScalar := [32]byte{1}
	if _, err := curve25519.X25519(testScalar[:], recipient.PublicKey); err != nil {
		return "", fmt.Errorf("invalid low-order recipient public key: %w", err)
	}
	fingerprint := types.ComputeKeyFingerprint(recipient.PublicKey)
	keyID := recipient.KeyID
	if keyID == "" {
		keyID = types.FormatRecipientKeyID(fingerprint, recipient.KeyVersion)
	}
	if types.NormalizeRecipientKeyID(keyID) != fingerprint {
		return "", fmt.Errorf("recipient key ID does not match public key fingerprint")
	}
	return keyID, nil
}

// GenerateKeyPair generates a new X25519 key pair using crypto/rand
func GenerateKeyPair() (*KeyPair, error) {
	var privateKey [32]byte
	if _, err := io.ReadFull(rand.Reader, privateKey[:]); err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	var publicKey [32]byte
	curve25519.ScalarBaseMult(&publicKey, &privateKey)

	signingPublicKey, signingPrivateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signing key: %w", err)
	}
	var signingPublicKeyArray [ed25519.PublicKeySize]byte
	var signingPrivateKeyArray [ed25519.PrivateKeySize]byte
	copy(signingPublicKeyArray[:], signingPublicKey)
	copy(signingPrivateKeyArray[:], signingPrivateKey)

	return &KeyPair{
		PublicKey:         publicKey,
		PrivateKey:        privateKey,
		SigningPublicKey:  signingPublicKeyArray,
		SigningPrivateKey: signingPrivateKeyArray,
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
	if err := senderKeyPair.validate(); err != nil {
		return nil, err
	}
	recipientKeyID, err := validateRecipient(recipient)
	if err != nil {
		return nil, err
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

	// Create envelope
	envelope := &types.EncryptedPayloadEnvelope{
		Version:             types.EnvelopeVersionV2,
		AlgorithmID:         types.AlgorithmX25519XSalsa20Poly1305,
		AlgorithmVersion:    types.AlgorithmVersionV1,
		RecipientKeyIDs:     []string{recipientKeyID},
		RecipientPublicKeys: [][]byte{append([]byte(nil), recipient.PublicKey...)},
		Nonce:               nonce[:],
		Ciphertext:          ciphertext,
		SenderPubKey:        senderKeyPair.PublicKey[:],
		SenderSigningPubKey: senderKeyPair.SigningPublicKey[:],
		Metadata:            make(map[string]string),
	}

	// Generate signature over the signing payload
	signature, err := signEnvelope(envelope, senderKeyPair.SigningPrivateKey[:])
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
	if err := senderKeyPair.validate(); err != nil {
		return nil, err
	}
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
	defer ZeroKey(&dek)

	// Generate nonce for data encryption
	dataNonce, err := GenerateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate data nonce: %w", err)
	}

	// Encrypt the payload using the audited XSalsa20-Poly1305 secretbox construction.
	ciphertext := secretbox.Seal(nil, plaintext, &dataNonce, &dek)

	// Encrypt DEK for each recipient
	recipientKeyIDs := make([]string, len(recipients))
	recipientPubKeys := make([][]byte, len(recipients))
	wrappedKeys := make([]types.WrappedKeyEntry, len(recipients))
	seenRecipients := make(map[string]struct{}, len(recipients))

	for i, recipient := range recipients {
		recipientPubKey := recipient.PublicKey
		keyID, err := validateRecipient(recipient)
		if err != nil {
			return nil, fmt.Errorf("invalid recipient at index %d: %w", i, err)
		}
		normalizedID := types.NormalizeRecipientKeyID(keyID)
		if _, exists := seenRecipients[normalizedID]; exists {
			return nil, fmt.Errorf("duplicate recipient public key at index %d", i)
		}
		seenRecipients[normalizedID] = struct{}{}

		var recipientPubKeyArr [32]byte
		copy(recipientPubKeyArr[:], recipientPubKey)

		ephemeralPublicKey, ephemeralPrivateKey, err := box.GenerateKey(rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ephemeral key for recipient %d: %w", i, err)
		}

		keyNonce, err := GenerateNonce()
		if err != nil {
			return nil, fmt.Errorf("failed to generate key nonce for recipient %d: %w", i, err)
		}

		// Prefix the random nonce to the authenticated ephemeral box ciphertext.
		encryptedDEK := box.Seal(keyNonce[:], dek[:], &keyNonce, &recipientPubKeyArr, ephemeralPrivateKey)

		recipientKeyIDs[i] = keyID
		recipientPubKeys[i] = append([]byte(nil), recipientPubKey...)
		wrappedKeys[i] = types.WrappedKeyEntry{
			RecipientID:     keyID,
			WrappedKey:      encryptedDEK,
			Algorithm:       types.WrappedKeyAlgorithmX25519NaClBox,
			EphemeralPubKey: append([]byte(nil), ephemeralPublicKey[:]...),
		}
	}

	// Create envelope
	envelope := &types.EncryptedPayloadEnvelope{
		Version:             types.EnvelopeVersionV2,
		AlgorithmID:         types.AlgorithmX25519XSalsa20Poly1305,
		AlgorithmVersion:    types.AlgorithmVersionV1,
		RecipientKeyIDs:     recipientKeyIDs,
		RecipientPublicKeys: recipientPubKeys,
		WrappedKeys:         wrappedKeys,
		Nonce:               dataNonce[:],
		Ciphertext:          ciphertext,
		SenderPubKey:        senderKeyPair.PublicKey[:],
		SenderSigningPubKey: senderKeyPair.SigningPublicKey[:],
		Metadata:            make(map[string]string),
	}

	// Add metadata to indicate multi-recipient mode
	envelope.Metadata["_mode"] = "multi-recipient"

	// Generate signature
	signature, err := signEnvelope(envelope, senderKeyPair.SigningPrivateKey[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign envelope: %w", err)
	}
	envelope.SenderSignature = signature

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

	if envelope.Version != types.EnvelopeVersionV2 {
		return nil, types.ErrUnsupportedVersion.Wrap("unauthenticated v1 envelopes require OpenUnauthenticatedLegacyEnvelopeV1")
	}

	if err := envelope.Validate(); err != nil {
		return nil, err
	}

	valid, err := ValidateEnvelopeSignature(envelope)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, types.ErrInvalidSignature.Wrap("sender signature verification failed")
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
func openMultiRecipientEnvelope(envelope *types.EncryptedPayloadEnvelope, recipientPrivateKey, _ *[32]byte) ([]byte, error) {
	// Derive our public key to find our encrypted key
	var ourPublicKey [32]byte
	curve25519.ScalarBaseMult(&ourPublicKey, recipientPrivateKey)
	ourFingerprint := types.ComputeKeyFingerprint(ourPublicKey[:])

	// Find our encrypted DEK
	var wrappedKey *types.WrappedKeyEntry
	for _, entry := range envelope.WrappedKeys {
		if types.NormalizeRecipientKeyID(entry.RecipientID) == ourFingerprint {
			entryCopy := entry
			wrappedKey = &entryCopy
			break
		}
	}
	if wrappedKey == nil {
		return nil, types.ErrNotRecipient.Wrap("no encrypted key found for this recipient")
	}

	var keyNonce [24]byte
	copy(keyNonce[:], wrappedKey.WrappedKey[:24])
	var ephemeralPublicKey [32]byte
	copy(ephemeralPublicKey[:], wrappedKey.EphemeralPubKey)

	// Decrypt DEK
	dek, ok := box.Open(nil, wrappedKey.WrappedKey[24:], &keyNonce, &ephemeralPublicKey, recipientPrivateKey)
	if !ok {
		return nil, types.ErrDecryptionFailed.Wrap("failed to decrypt data encryption key")
	}

	if len(dek) != 32 {
		ZeroBytes(dek)
		return nil, types.ErrDecryptionFailed.Wrap("invalid DEK size")
	}
	defer ZeroBytes(dek)

	// Decrypt data with DEK
	var dekArr [32]byte
	copy(dekArr[:], dek)
	defer ZeroKey(&dekArr)

	var dataNonce [24]byte
	copy(dataNonce[:], envelope.Nonce)

	plaintext, ok := secretbox.Open(nil, envelope.Ciphertext, &dataNonce, &dekArr)
	if !ok {
		return nil, types.ErrDecryptionFailed.Wrap("payload authentication failed")
	}

	return plaintext, nil
}

// ValidateEnvelopeSignature verifies the sender's signature on an envelope.
func ValidateEnvelopeSignature(envelope *types.EncryptedPayloadEnvelope) (bool, error) {
	if envelope == nil {
		return false, types.ErrInvalidEnvelope.Wrap("envelope cannot be nil")
	}

	if len(envelope.SenderSignature) == 0 {
		return false, types.ErrInvalidSignature.Wrap("no signature present")
	}

	if envelope.Version != types.EnvelopeVersionV2 {
		return false, types.ErrInvalidSignature.Wrap("legacy v1 envelopes are unauthenticated")
	}

	if len(envelope.SenderSigningPubKey) != ed25519.PublicKeySize {
		return false, types.ErrInvalidPublicKey.Wrap("invalid sender signing public key")
	}
	if len(envelope.SenderSignature) != ed25519.SignatureSize {
		return false, nil
	}

	return ed25519.Verify(ed25519.PublicKey(envelope.SenderSigningPubKey), envelope.SigningPayload(), envelope.SenderSignature), nil
}

// ValidateEnvelopeSignatureWithExpectedKey authenticates the envelope against a trusted sender key.
func ValidateEnvelopeSignatureWithExpectedKey(envelope *types.EncryptedPayloadEnvelope, expectedSigningPublicKey []byte) (bool, error) {
	if len(expectedSigningPublicKey) != ed25519.PublicKeySize {
		return false, types.ErrInvalidPublicKey.Wrap("invalid expected sender signing public key")
	}
	if envelope == nil || !bytes.Equal(envelope.SenderSigningPubKey, expectedSigningPublicKey) {
		return false, types.ErrInvalidSignature.Wrap("sender signing public key does not match trusted key")
	}
	return ValidateEnvelopeSignature(envelope)
}

func signEnvelope(envelope *types.EncryptedPayloadEnvelope, privateKey ed25519.PrivateKey) ([]byte, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid Ed25519 private key size: %d", len(privateKey))
	}
	return ed25519.Sign(privateKey, envelope.SigningPayload()), nil
}

// OpenUnauthenticatedLegacyEnvelopeV1 decrypts a v1 envelope for migration only.
// V1 sender signatures are publicly forgeable and are intentionally not verified.
func OpenUnauthenticatedLegacyEnvelopeV1(envelope *types.EncryptedPayloadEnvelope, recipientPrivateKey []byte) ([]byte, error) {
	if envelope == nil || envelope.Version != types.EnvelopeVersionV1 {
		return nil, types.ErrUnsupportedVersion.Wrap("legacy v1 envelope required")
	}
	if err := envelope.Validate(); err != nil {
		return nil, err
	}
	if len(recipientPrivateKey) != 32 {
		return nil, fmt.Errorf("invalid private key size: expected 32, got %d", len(recipientPrivateKey))
	}
	var privateKey, senderPublicKey [32]byte
	copy(privateKey[:], recipientPrivateKey)
	copy(senderPublicKey[:], envelope.SenderPubKey)
	if envelope.Metadata["_mode"] == "multi-recipient" {
		return openUnauthenticatedLegacyMultiRecipientEnvelopeV1(envelope, &privateKey, &senderPublicKey)
	}
	var nonce [24]byte
	copy(nonce[:], envelope.Nonce)
	plaintext, ok := box.Open(nil, envelope.Ciphertext, &nonce, &senderPublicKey, &privateKey)
	if !ok {
		return nil, types.ErrDecryptionFailed.Wrap("failed to decrypt legacy envelope")
	}
	return plaintext, nil
}

func openUnauthenticatedLegacyMultiRecipientEnvelopeV1(envelope *types.EncryptedPayloadEnvelope, recipientPrivateKey, senderPublicKey *[32]byte) ([]byte, error) {
	var recipientPublicKey [32]byte
	curve25519.ScalarBaseMult(&recipientPublicKey, recipientPrivateKey)
	fingerprint := types.ComputeKeyFingerprint(recipientPublicKey[:])
	var encryptedDEK []byte
	for i, recipientID := range envelope.RecipientKeyIDs {
		if types.NormalizeRecipientKeyID(recipientID) == fingerprint && i < len(envelope.EncryptedKeys) {
			encryptedDEK = envelope.EncryptedKeys[i]
			break
		}
	}
	if len(encryptedDEK) < 24 {
		return nil, types.ErrNotRecipient.Wrap("no encrypted key found for this recipient")
	}
	var keyNonce [24]byte
	copy(keyNonce[:], encryptedDEK[:24])
	dek, ok := box.Open(nil, encryptedDEK[24:], &keyNonce, senderPublicKey, recipientPrivateKey)
	if !ok || len(dek) != 32 {
		ZeroBytes(dek)
		return nil, types.ErrDecryptionFailed.Wrap("failed to decrypt legacy data encryption key")
	}
	defer ZeroBytes(dek)
	var dekArray [32]byte
	copy(dekArray[:], dek)
	defer ZeroKey(&dekArray)
	var dataNonce [24]byte
	copy(dataNonce[:], envelope.Nonce)
	var zeroKey [32]byte
	var sharedKey [32]byte
	box.Precompute(&sharedKey, &zeroKey, &dekArray)
	defer ZeroKey(&sharedKey)
	plaintext, ok := box.OpenAfterPrecomputation(nil, envelope.Ciphertext, &dataNonce, &sharedKey)
	if !ok {
		return nil, types.ErrDecryptionFailed.Wrap("failed to decrypt legacy payload")
	}
	return plaintext, nil
}
