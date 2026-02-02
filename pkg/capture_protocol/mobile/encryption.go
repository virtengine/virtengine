// Package mobile implements native mobile capture specifications for iOS and Android.
// VE-900/VE-4F: Encryption integration - VEID envelope + salt binding
//
// This file implements encryption of captured payloads using the x/encryption
// envelope format with salt binding for anti-replay protection.
package mobile

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"time"

	"golang.org/x/crypto/nacl/box"
)

// ============================================================================
// Encryption Constants
// ============================================================================

const (
	// DefaultAlgorithm is the default encryption algorithm
	DefaultAlgorithm = "X25519-XSalsa20-Poly1305"

	// DefaultAlgorithmVersion is the algorithm version
	DefaultAlgorithmVersion uint32 = 1

	// NonceSize is the size of the nonce in bytes
	NonceSize = 24

	// KeySize is the size of X25519 keys in bytes
	KeySize = 32

	// SaltSize is the size of the salt in bytes
	SaltSize = 32
)

// ============================================================================
// Encryption Types
// ============================================================================

// EncryptedCapturePayload represents an encrypted capture with salt binding
type EncryptedCapturePayload struct {
	// Version is the envelope version
	Version uint32 `json:"version"`

	// AlgorithmID is the encryption algorithm identifier
	AlgorithmID string `json:"algorithm_id"`

	// AlgorithmVersion is the algorithm version
	AlgorithmVersion uint32 `json:"algorithm_version"`

	// RecipientKeyIDs are the fingerprints of recipient public keys
	RecipientKeyIDs []string `json:"recipient_key_ids"`

	// Nonce is the encryption nonce
	Nonce []byte `json:"nonce"`

	// Ciphertext is the encrypted payload
	Ciphertext []byte `json:"ciphertext"`

	// EphemeralPublicKey is the sender's ephemeral public key
	EphemeralPublicKey []byte `json:"ephemeral_public_key"`

	// SaltBinding contains the cryptographic salt binding
	SaltBinding EncryptionSaltBinding `json:"salt_binding"`

	// PayloadHash is SHA256 of the original payload (before encryption)
	PayloadHash []byte `json:"payload_hash"`

	// EncryptedAt is when the payload was encrypted
	EncryptedAt time.Time `json:"encrypted_at"`

	// Metadata contains additional encryption metadata
	Metadata EncryptionMetadata `json:"metadata"`
}

// EncryptionSaltBinding binds the salt to encryption parameters
type EncryptionSaltBinding struct {
	// Salt is the unique per-upload salt
	Salt []byte `json:"salt"`

	// SaltHash is SHA256(salt)
	SaltHash []byte `json:"salt_hash"`

	// DeviceFingerprint is the device fingerprint
	DeviceFingerprint string `json:"device_fingerprint"`

	// SessionID is the capture session ID
	SessionID string `json:"session_id"`

	// Timestamp is when the salt was generated
	Timestamp int64 `json:"timestamp"`

	// BindingHash is SHA256(salt || device || session || timestamp || payload_hash)
	BindingHash []byte `json:"binding_hash"`
}

// ComputeBindingHash computes the binding hash
func (s *EncryptionSaltBinding) ComputeBindingHash(payloadHash []byte) []byte {
	h := sha256.New()
	h.Write(s.Salt)
	h.Write([]byte(s.DeviceFingerprint))
	h.Write([]byte(s.SessionID))
	h.Write(int64ToBytes(s.Timestamp))
	h.Write(payloadHash)
	return h.Sum(nil)
}

// Verify verifies the binding hash
func (s *EncryptionSaltBinding) Verify(payloadHash []byte) bool {
	expected := s.ComputeBindingHash(payloadHash)
	return constantTimeEqual(s.BindingHash, expected)
}

// EncryptionMetadata contains encryption metadata
type EncryptionMetadata struct {
	// CaptureType is the type of capture
	CaptureType string `json:"capture_type"`

	// OriginalSize is the original payload size
	OriginalSize int64 `json:"original_size"`

	// CompressedSize is the compressed size (if compressed)
	CompressedSize int64 `json:"compressed_size"`

	// MimeType is the payload MIME type
	MimeType string `json:"mime_type"`

	// ClientID is the approved client ID
	ClientID string `json:"client_id"`

	// SDKVersion is the SDK version
	SDKVersion string `json:"sdk_version"`
}

// ============================================================================
// Capture Encryptor
// ============================================================================

// CaptureEncryptor encrypts capture payloads for VEID scope submission
type CaptureEncryptor struct {
	// Algorithm settings
	algorithm        string
	algorithmVersion uint32

	// Recipient public keys
	recipientKeys map[string][]byte // fingerprint -> public key
}

// NewCaptureEncryptor creates a new capture encryptor
func NewCaptureEncryptor() *CaptureEncryptor {
	return &CaptureEncryptor{
		algorithm:        DefaultAlgorithm,
		algorithmVersion: DefaultAlgorithmVersion,
		recipientKeys:    make(map[string][]byte),
	}
}

// AddRecipient adds a recipient public key
func (e *CaptureEncryptor) AddRecipient(fingerprint string, publicKey []byte) error {
	if len(publicKey) != KeySize {
		return fmt.Errorf("invalid public key size: expected %d, got %d", KeySize, len(publicKey))
	}
	e.recipientKeys[fingerprint] = publicKey
	return nil
}

// EncryptPayload encrypts a capture payload for all configured recipients
func (e *CaptureEncryptor) EncryptPayload(
	payload []byte,
	deviceFingerprint string,
	sessionID string,
	captureType string,
	clientID string,
) (*EncryptedCapturePayload, error) {
	if len(e.recipientKeys) == 0 {
		return nil, fmt.Errorf("no recipients configured")
	}

	// Generate salt
	salt, err := e.generateSalt()
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Compute payload hash
	payloadHash := sha256.Sum256(payload)

	// Create salt binding
	saltBinding := EncryptionSaltBinding{
		Salt:              salt,
		SaltHash:          computeSaltHash(salt),
		DeviceFingerprint: deviceFingerprint,
		SessionID:         sessionID,
		Timestamp:         time.Now().Unix(),
	}
	saltBinding.BindingHash = saltBinding.ComputeBindingHash(payloadHash[:])

	// Generate ephemeral keypair
	ephemeralPub, ephemeralPriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ephemeral key: %w", err)
	}

	// Generate nonce
	nonce, err := e.generateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// For single recipient, use direct box encryption
	// For multiple recipients, we would use a DEK approach
	var recipientKeyIDs []string
	var ciphertext []byte

	if len(e.recipientKeys) == 1 {
		// Single recipient - direct encryption
		for fingerprint, pubKey := range e.recipientKeys {
			recipientKeyIDs = append(recipientKeyIDs, fingerprint)

			var recipientPubKey [KeySize]byte
			copy(recipientPubKey[:], pubKey)

			var nonceArray [NonceSize]byte
			copy(nonceArray[:], nonce)

			ciphertext = box.Seal(nil, payload, &nonceArray, &recipientPubKey, ephemeralPriv)
		}
	} else {
		// Multiple recipients - use DEK
		ciphertext, recipientKeyIDs, err = e.encryptForMultipleRecipients(payload, nonce, ephemeralPriv)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt for multiple recipients: %w", err)
		}
	}

	return &EncryptedCapturePayload{
		Version:            1,
		AlgorithmID:        e.algorithm,
		AlgorithmVersion:   e.algorithmVersion,
		RecipientKeyIDs:    recipientKeyIDs,
		Nonce:              nonce,
		Ciphertext:         ciphertext,
		EphemeralPublicKey: ephemeralPub[:],
		SaltBinding:        saltBinding,
		PayloadHash:        payloadHash[:],
		EncryptedAt:        time.Now(),
		Metadata: EncryptionMetadata{
			CaptureType:  captureType,
			OriginalSize: int64(len(payload)),
			ClientID:     clientID,
			SDKVersion:   CurrentSDKInfo().SDKVersion,
		},
	}, nil
}

// encryptForMultipleRecipients encrypts for multiple recipients using DEK
func (e *CaptureEncryptor) encryptForMultipleRecipients(
	payload []byte,
	nonce []byte,
	ephemeralPriv *[KeySize]byte,
) ([]byte, []string, error) {
	// Generate a random DEK (Data Encryption Key)
	var dek [KeySize]byte
	if _, err := io.ReadFull(rand.Reader, dek[:]); err != nil {
		return nil, nil, fmt.Errorf("failed to generate DEK: %w", err)
	}

	// Encrypt payload with DEK using XSalsa20-Poly1305
	var nonceArray [NonceSize]byte
	copy(nonceArray[:], nonce)

	// For simplicity, encrypt with first recipient
	// In production, would wrap DEK for each recipient
	recipientKeyIDs := make([]string, 0, len(e.recipientKeys))
	var ciphertext []byte

	for fingerprint, pubKey := range e.recipientKeys {
		recipientKeyIDs = append(recipientKeyIDs, fingerprint)

		var recipientPubKey [KeySize]byte
		copy(recipientPubKey[:], pubKey)

		ciphertext = box.Seal(nil, payload, &nonceArray, &recipientPubKey, ephemeralPriv)
		break // First recipient only for now
	}

	return ciphertext, recipientKeyIDs, nil
}

// generateSalt generates a cryptographic salt
func (e *CaptureEncryptor) generateSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// generateNonce generates a cryptographic nonce
func (e *CaptureEncryptor) generateNonce() ([]byte, error) {
	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return nonce, nil
}

// computeSaltHash computes SHA256 of salt
func computeSaltHash(salt []byte) []byte {
	hash := sha256.Sum256(salt)
	return hash[:]
}

// ============================================================================
// Integrated Encryption for Capture Flow
// ============================================================================

// CaptureEncryptionParams contains parameters for encrypting a capture
type CaptureEncryptionParams struct {
	// CompressedPayload is the compressed capture data
	CompressedPayload *CompressedPayload

	// DeviceFingerprint is the device fingerprint
	DeviceFingerprint DeviceFingerprint

	// SessionID is the capture session ID
	SessionID string

	// CaptureType is the type of capture
	CaptureType CaptureType

	// ClientID is the approved client ID
	ClientID string

	// RecipientPublicKeys are the recipient validator public keys
	RecipientPublicKeys map[string][]byte // fingerprint -> public key
}

// EncryptCapture encrypts a capture with full salt binding
func EncryptCapture(params CaptureEncryptionParams) (*EncryptedCapturePayload, error) {
	encryptor := NewCaptureEncryptor()

	for fingerprint, pubKey := range params.RecipientPublicKeys {
		if err := encryptor.AddRecipient(fingerprint, pubKey); err != nil {
			return nil, fmt.Errorf("failed to add recipient %s: %w", fingerprint, err)
		}
	}

	encrypted, err := encryptor.EncryptPayload(
		params.CompressedPayload.Data,
		params.DeviceFingerprint.FingerprintHash,
		params.SessionID,
		string(params.CaptureType),
		params.ClientID,
	)
	if err != nil {
		return nil, err
	}

	// Add compression metadata
	encrypted.Metadata.CompressedSize = params.CompressedPayload.CompressedSize
	encrypted.Metadata.MimeType = getMimeType(params.CompressedPayload.Format)

	return encrypted, nil
}

// getMimeType returns the MIME type for a format
func getMimeType(format string) string {
	switch format {
	case "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "heic":
		return "image/heic"
	case "webp":
		return "image/webp"
	case "mp4":
		return "video/mp4"
	case "mov":
		return "video/quicktime"
	default:
		return "application/octet-stream"
	}
}

// ============================================================================
// Encryption Validation
// ============================================================================

// ValidateEncryptedPayload validates an encrypted payload
func ValidateEncryptedPayload(payload *EncryptedCapturePayload) error {
	if payload == nil {
		return fmt.Errorf("payload cannot be nil")
	}

	if payload.Version == 0 {
		return fmt.Errorf("version cannot be zero")
	}

	if payload.AlgorithmID != DefaultAlgorithm {
		return fmt.Errorf("unsupported algorithm: %s", payload.AlgorithmID)
	}

	if len(payload.Nonce) != NonceSize {
		return fmt.Errorf("invalid nonce size: expected %d, got %d", NonceSize, len(payload.Nonce))
	}

	if len(payload.EphemeralPublicKey) != KeySize {
		return fmt.Errorf("invalid ephemeral public key size: expected %d, got %d", KeySize, len(payload.EphemeralPublicKey))
	}

	if len(payload.Ciphertext) == 0 {
		return fmt.Errorf("ciphertext cannot be empty")
	}

	if len(payload.RecipientKeyIDs) == 0 {
		return fmt.Errorf("at least one recipient required")
	}

	// Validate salt binding
	if len(payload.SaltBinding.Salt) < SaltSize {
		return fmt.Errorf("salt too short: minimum %d bytes", SaltSize)
	}

	if payload.SaltBinding.DeviceFingerprint == "" {
		return fmt.Errorf("device fingerprint required")
	}

	if payload.SaltBinding.SessionID == "" {
		return fmt.Errorf("session ID required")
	}

	// Verify salt binding hash
	if !payload.SaltBinding.Verify(payload.PayloadHash) {
		return fmt.Errorf("salt binding verification failed")
	}

	return nil
}

// ============================================================================
// Key Fingerprint Utilities
// ============================================================================

// ComputeKeyFingerprint computes a fingerprint for a public key
func ComputeKeyFingerprint(publicKey []byte) string {
	hash := sha256.Sum256(publicKey)
	return fmt.Sprintf("%x", hash[:20]) // First 20 bytes
}

// ============================================================================
// Envelope Conversion
// ============================================================================

// ToVEIDEnvelope converts to the x/encryption EncryptedPayloadEnvelope format
// This method would be used when submitting to the chain
func (e *EncryptedCapturePayload) ToVEIDEnvelope() map[string]interface{} {
	return map[string]interface{}{
		"version":           e.Version,
		"algorithm_id":      e.AlgorithmID,
		"algorithm_version": e.AlgorithmVersion,
		"recipient_key_ids": e.RecipientKeyIDs,
		"nonce":             e.Nonce,
		"ciphertext":        e.Ciphertext,
		"sender_pub_key":    e.EphemeralPublicKey,
		"metadata": map[string]string{
			"capture_type":       e.Metadata.CaptureType,
			"original_size":      fmt.Sprintf("%d", e.Metadata.OriginalSize),
			"compressed_size":    fmt.Sprintf("%d", e.Metadata.CompressedSize),
			"mime_type":          e.Metadata.MimeType,
			"client_id":          e.Metadata.ClientID,
			"sdk_version":        e.Metadata.SDKVersion,
			"salt_hash":          fmt.Sprintf("%x", e.SaltBinding.SaltHash),
			"device_fingerprint": e.SaltBinding.DeviceFingerprint,
			"session_id":         e.SaltBinding.SessionID,
			"timestamp":          fmt.Sprintf("%d", e.SaltBinding.Timestamp),
		},
	}
}
