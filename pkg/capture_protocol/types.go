// Package capture_protocol implements the VirtEngine Mobile Capture Protocol v1.
// This protocol defines secure upload mechanisms with salt binding and
// anti-replay protection for identity document and selfie captures.
//
// Protocol Features:
//   - Per-upload salt generation and binding
//   - Dual signature scheme (client + user)
//   - Anti-replay protection via salt caching
//   - Approved client verification
//   - Key rotation support
//
// See _docs/protocols/mobile-capture-protocol-v1.md for full specification.
package capture_protocol

import (
	"crypto/sha256"
	"time"
)

// ProtocolVersion is the current protocol version
const ProtocolVersion uint32 = 1

// Protocol constants
const (
	// MinSaltLength is the minimum required salt length in bytes
	MinSaltLength = 32

	// MaxSaltLength is the maximum allowed salt length in bytes
	MaxSaltLength = 64

	// DefaultMaxSaltAge is the default maximum age for a salt
	DefaultMaxSaltAge = 5 * time.Minute

	// DefaultReplayWindow is the default window for replay detection
	DefaultReplayWindow = 10 * time.Minute

	// DefaultMaxClockSkew is the default allowed clock skew
	DefaultMaxClockSkew = 30 * time.Second

	// AlgorithmEd25519 is the Ed25519 signature algorithm
	AlgorithmEd25519 = "ed25519"

	// AlgorithmSecp256k1 is the secp256k1 signature algorithm
	AlgorithmSecp256k1 = "secp256k1"
)

// CapturePayload represents a complete capture upload with all required
// bindings and signatures for the Mobile Capture Protocol v1.
type CapturePayload struct {
	// Version is the protocol version
	Version uint32 `json:"version"`

	// PayloadHash is the SHA256 hash of the encrypted content
	PayloadHash []byte `json:"payload_hash"`

	// Salt is the per-upload unique salt (32 bytes)
	Salt []byte `json:"salt"`

	// SaltBinding contains the cryptographic binding of the salt
	SaltBinding SaltBinding `json:"salt_binding"`

	// ClientSignature is the signature from the approved client
	ClientSignature SignatureProof `json:"client_signature"`

	// UserSignature is the signature from the user's account
	UserSignature SignatureProof `json:"user_signature"`

	// CaptureMetadata contains metadata about the capture
	CaptureMetadata CaptureMetadata `json:"capture_metadata"`

	// Timestamp is when the payload was created
	Timestamp time.Time `json:"timestamp"`
}

// SaltBinding represents the cryptographic binding of a salt to device,
// session, and timestamp to prevent replay attacks.
type SaltBinding struct {
	// Salt is the unique per-upload salt
	Salt []byte `json:"salt"`

	// DeviceID is the bound device fingerprint hash
	DeviceID string `json:"device_id"`

	// SessionID is the bound capture session identifier
	SessionID string `json:"session_id"`

	// Timestamp is the Unix timestamp when binding was created
	Timestamp int64 `json:"timestamp"`

	// BindingHash is SHA256(salt || device_id || session_id || timestamp)
	BindingHash []byte `json:"binding_hash"`
}

// ComputeBindingHash computes the expected binding hash for verification
func (sb *SaltBinding) ComputeBindingHash() []byte {
	h := sha256.New()
	h.Write(sb.Salt)
	h.Write([]byte(sb.DeviceID))
	h.Write([]byte(sb.SessionID))
	h.Write(int64ToBytes(sb.Timestamp))
	return h.Sum(nil)
}

// VerifyBindingHash verifies that the stored binding hash is correct
func (sb *SaltBinding) VerifyBindingHash() bool {
	expected := sb.ComputeBindingHash()
	return constantTimeEqual(sb.BindingHash, expected)
}

// SignatureProof contains a cryptographic signature with metadata
// for verification purposes.
type SignatureProof struct {
	// PublicKey is the signer's public key
	PublicKey []byte `json:"public_key"`

	// Signature is the cryptographic signature
	Signature []byte `json:"signature"`

	// Algorithm is the signature algorithm ("ed25519" or "secp256k1")
	Algorithm string `json:"algorithm"`

	// KeyID is the client ID (for clients) or account address (for users)
	KeyID string `json:"key_id"`

	// SignedData is the data that was signed
	SignedData []byte `json:"signed_data"`
}

// CaptureMetadata contains metadata about a capture event.
type CaptureMetadata struct {
	// DeviceFingerprint is a hash of device identifiers
	DeviceFingerprint string `json:"device_fingerprint"`

	// ClientID is the approved client identifier
	ClientID string `json:"client_id"`

	// ClientVersion is the client application version
	ClientVersion string `json:"client_version"`

	// SessionID is the capture session identifier
	SessionID string `json:"session_id"`

	// DocumentType is the type of document captured
	DocumentType string `json:"document_type"`

	// QualityScore is the quality validation score (0-100)
	QualityScore uint32 `json:"quality_score"`

	// CaptureTimestamp is when the image was captured
	CaptureTimestamp int64 `json:"capture_timestamp"`

	// GeoHint is an optional geographic hint (country code)
	GeoHint string `json:"geo_hint,omitempty"`
}

// ApprovedClient represents a client registered in the approved client allowlist.
type ApprovedClient struct {
	// ClientID is the unique identifier for the client
	ClientID string `json:"client_id"`

	// Name is a human-readable name
	Name string `json:"name"`

	// PublicKey is the client's current active public key
	PublicKey []byte `json:"public_key"`

	// Algorithm is the signature algorithm used by the client
	Algorithm string `json:"algorithm"`

	// Active indicates if the client is currently active
	Active bool `json:"active"`

	// RegisteredAt is when the client was registered
	RegisteredAt time.Time `json:"registered_at"`

	// DeactivatedAt is when the client was deactivated (if applicable)
	DeactivatedAt *time.Time `json:"deactivated_at,omitempty"`

	// DeprecatedKey is the previous key (valid during overlap period)
	DeprecatedKey []byte `json:"deprecated_key,omitempty"`

	// DeprecatedKeyExpiry is when the deprecated key expires
	DeprecatedKeyExpiry *time.Time `json:"deprecated_key_expiry,omitempty"`
}

// IsKeyValid checks if a given public key is valid for this client
// This supports key rotation by accepting both active and deprecated keys
func (ac *ApprovedClient) IsKeyValid(publicKey []byte) bool {
	if !ac.Active {
		return false
	}

	// Check active key
	if constantTimeEqual(ac.PublicKey, publicKey) {
		return true
	}

	// Check deprecated key if still in overlap period
	if ac.DeprecatedKey != nil && ac.DeprecatedKeyExpiry != nil {
		if time.Now().Before(*ac.DeprecatedKeyExpiry) {
			if constantTimeEqual(ac.DeprecatedKey, publicKey) {
				return true
			}
		}
	}

	return false
}

// KeyState represents the state of a client key
type KeyState string

const (
	// KeyStatePending indicates a key that is registered but not yet active
	KeyStatePending KeyState = "pending"

	// KeyStateActive indicates a currently valid key
	KeyStateActive KeyState = "active"

	// KeyStateDeprecated indicates a key in overlap period (still valid)
	KeyStateDeprecated KeyState = "deprecated"

	// KeyStateRevoked indicates an invalid key (rejected)
	KeyStateRevoked KeyState = "revoked"
)

// ApprovedClientRegistry is the interface for managing approved clients
type ApprovedClientRegistry interface {
	// GetClient returns an approved client by ID
	GetClient(clientID string) (*ApprovedClient, error)

	// IsApproved checks if a client ID is in the approved list
	IsApproved(clientID string) bool

	// VerifyClientKey verifies a client's public key
	VerifyClientKey(clientID string, publicKey []byte) error
}

// ValidationConfig contains configuration for protocol validation
type ValidationConfig struct {
	// MinSaltLength is the minimum salt length in bytes
	MinSaltLength int `json:"min_salt_length"`

	// MaxSaltAge is the maximum age for salt timestamps
	MaxSaltAge time.Duration `json:"max_salt_age"`

	// ReplayWindow is the duration to cache salts for replay detection
	ReplayWindow time.Duration `json:"replay_window"`

	// MaxClockSkew is the maximum allowed clock skew
	MaxClockSkew time.Duration `json:"max_clock_skew"`

	// RequireClientSignature indicates if client signature is required
	RequireClientSignature bool `json:"require_client_signature"`

	// RequireUserSignature indicates if user signature is required
	RequireUserSignature bool `json:"require_user_signature"`
}

// DefaultValidationConfig returns the default validation configuration
func DefaultValidationConfig() ValidationConfig {
	return ValidationConfig{
		MinSaltLength:          MinSaltLength,
		MaxSaltAge:             DefaultMaxSaltAge,
		ReplayWindow:           DefaultReplayWindow,
		MaxClockSkew:           DefaultMaxClockSkew,
		RequireClientSignature: true,
		RequireUserSignature:   true,
	}
}

// ValidationResult contains the result of payload validation
type ValidationResult struct {
	// Valid indicates if the payload is valid
	Valid bool `json:"valid"`

	// Errors contains validation errors (if any)
	Errors []ValidationError `json:"errors,omitempty"`

	// ClientID is the verified client ID
	ClientID string `json:"client_id,omitempty"`

	// UserAddress is the verified user address
	UserAddress string `json:"user_address,omitempty"`

	// ValidatedAt is when validation was performed
	ValidatedAt time.Time `json:"validated_at"`
}

// ValidationError represents a specific validation failure
type ValidationError struct {
	// Code is the error code
	Code string `json:"code"`

	// Message is the error message
	Message string `json:"message"`

	// Field is the field that caused the error (if applicable)
	Field string `json:"field,omitempty"`
}

// Helper functions

// int64ToBytes converts an int64 to a byte slice (big-endian)
func int64ToBytes(n int64) []byte {
	b := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		b[i] = byte(n & 0xff)
		n >>= 8
	}
	return b
}

// constantTimeEqual performs constant-time comparison of two byte slices
func constantTimeEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

// ComputeClientSigningData computes the data that clients should sign
// SignedData = salt || payload_hash
func ComputeClientSigningData(salt, payloadHash []byte) []byte {
	result := make([]byte, len(salt)+len(payloadHash))
	copy(result, salt)
	copy(result[len(salt):], payloadHash)
	return result
}

// ComputeUserSigningData computes the data that users should sign
// SignedData = salt || payload_hash || client_signature
func ComputeUserSigningData(salt, payloadHash, clientSignature []byte) []byte {
	result := make([]byte, len(salt)+len(payloadHash)+len(clientSignature))
	copy(result, salt)
	copy(result[len(salt):], payloadHash)
	copy(result[len(salt)+len(payloadHash):], clientSignature)
	return result
}
