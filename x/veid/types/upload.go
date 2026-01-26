package types

import (
	"crypto/sha256"
	"encoding/hex"
)

// UploadMetadata contains metadata about an identity scope upload
// This includes cryptographic binding via salt, device information,
// and dual signatures (approved client + user)
type UploadMetadata struct {
	// Salt is a per-upload unique salt for cryptographic binding
	// This prevents replay attacks and ensures each upload is unique
	Salt []byte `json:"salt"`

	// SaltHash is the SHA256 hash of the salt (for quick lookup without exposing salt)
	SaltHash []byte `json:"salt_hash"`

	// DeviceFingerprint is a hash/identifier of the device used for upload
	// Used for device binding and anomaly detection
	DeviceFingerprint string `json:"device_fingerprint"`

	// ClientID is the identifier of the approved client that facilitated the upload
	ClientID string `json:"client_id"`

	// ClientSignature is the cryptographic signature from the approved client
	// This proves the upload came through an approved client application
	ClientSignature []byte `json:"client_signature"`

	// UserSignature is the cryptographic signature from the user's account
	// This proves the user authorized this upload
	UserSignature []byte `json:"user_signature"`

	// PayloadHash is the SHA256 hash of the encrypted payload
	// Used for integrity verification without decryption
	PayloadHash []byte `json:"payload_hash"`

	// UploadNonce is a unique nonce for this upload session
	UploadNonce []byte `json:"upload_nonce"`

	// CaptureTimestamp is when the data was captured (from client)
	// This may differ from upload time due to offline capture
	CaptureTimestamp int64 `json:"capture_timestamp,omitempty"`

	// GeoHint is an optional geographic hint (coarse location, country code)
	// Used for fraud detection, not for precise location tracking
	GeoHint string `json:"geo_hint,omitempty"`
}

// NewUploadMetadata creates a new upload metadata instance
func NewUploadMetadata(
	salt []byte,
	deviceFingerprint string,
	clientID string,
	clientSignature []byte,
	userSignature []byte,
	payloadHash []byte,
) *UploadMetadata {
	return &UploadMetadata{
		Salt:              salt,
		SaltHash:          ComputeSaltHash(salt),
		DeviceFingerprint: deviceFingerprint,
		ClientID:          clientID,
		ClientSignature:   clientSignature,
		UserSignature:     userSignature,
		PayloadHash:       payloadHash,
	}
}

// Validate validates the upload metadata
func (m *UploadMetadata) Validate() error {
	if len(m.Salt) == 0 {
		return ErrInvalidSalt.Wrap("salt cannot be empty")
	}

	if len(m.Salt) < 16 {
		return ErrInvalidSalt.Wrap("salt must be at least 16 bytes")
	}

	if len(m.Salt) > 64 {
		return ErrInvalidSalt.Wrap("salt cannot exceed 64 bytes")
	}

	if m.DeviceFingerprint == "" {
		return ErrInvalidDeviceInfo.Wrap("device fingerprint cannot be empty")
	}

	if len(m.DeviceFingerprint) > 256 {
		return ErrInvalidDeviceInfo.Wrap("device fingerprint exceeds maximum length")
	}

	if m.ClientID == "" {
		return ErrInvalidClientID.Wrap("client_id cannot be empty")
	}

	if len(m.ClientID) > 128 {
		return ErrInvalidClientID.Wrap("client_id exceeds maximum length")
	}

	if len(m.ClientSignature) == 0 {
		return ErrInvalidClientSignature.Wrap("client signature cannot be empty")
	}

	if len(m.UserSignature) == 0 {
		return ErrInvalidUserSignature.Wrap("user signature cannot be empty")
	}

	if len(m.PayloadHash) == 0 {
		return ErrInvalidPayloadHash.Wrap("payload hash cannot be empty")
	}

	if len(m.PayloadHash) != 32 {
		return ErrInvalidPayloadHash.Wrap("payload hash must be 32 bytes (SHA256)")
	}

	// Verify salt hash matches salt
	expectedHash := ComputeSaltHash(m.Salt)
	if len(m.SaltHash) > 0 && !bytesEqual(m.SaltHash, expectedHash) {
		return ErrInvalidSalt.Wrap("salt hash does not match salt")
	}

	return nil
}

// ComputeSaltHash computes the SHA256 hash of a salt
func ComputeSaltHash(salt []byte) []byte {
	hash := sha256.Sum256(salt)
	return hash[:]
}

// SaltHashHex returns the salt hash as a hex string
func (m *UploadMetadata) SaltHashHex() string {
	return hex.EncodeToString(m.SaltHash)
}

// PayloadHashHex returns the payload hash as a hex string
func (m *UploadMetadata) PayloadHashHex() string {
	return hex.EncodeToString(m.PayloadHash)
}

// SigningPayload returns the bytes that should be signed by the client
// This ensures the signature covers all relevant metadata
func (m *UploadMetadata) SigningPayload() []byte {
	h := sha256.New()

	h.Write(m.Salt)
	h.Write([]byte(m.DeviceFingerprint))
	h.Write([]byte(m.ClientID))
	h.Write(m.PayloadHash)

	if len(m.UploadNonce) > 0 {
		h.Write(m.UploadNonce)
	}

	return h.Sum(nil)
}

// UserSigningPayload returns the bytes that should be signed by the user
// This includes the client signature to create a signature chain
func (m *UploadMetadata) UserSigningPayload() []byte {
	h := sha256.New()

	h.Write(m.SigningPayload())
	h.Write(m.ClientSignature)

	return h.Sum(nil)
}

// bytesEqual is a constant-time comparison of two byte slices
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

// ApprovedClient represents an approved client that can facilitate uploads
type ApprovedClient struct {
	// ClientID is the unique identifier for the client
	ClientID string `json:"client_id"`

	// Name is a human-readable name for the client
	Name string `json:"name"`

	// PublicKey is the client's public key for signature verification
	PublicKey []byte `json:"public_key"`

	// Algorithm is the signature algorithm used by the client
	Algorithm string `json:"algorithm"`

	// Active indicates if the client is currently approved
	Active bool `json:"active"`

	// RegisteredAt is when the client was registered
	RegisteredAt int64 `json:"registered_at"`

	// DeactivatedAt is when the client was deactivated (if applicable)
	DeactivatedAt int64 `json:"deactivated_at,omitempty"`

	// Metadata contains optional additional client information
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewApprovedClient creates a new approved client
func NewApprovedClient(clientID, name string, publicKey []byte, algorithm string, registeredAt int64) *ApprovedClient {
	return &ApprovedClient{
		ClientID:     clientID,
		Name:         name,
		PublicKey:    publicKey,
		Algorithm:    algorithm,
		Active:       true,
		RegisteredAt: registeredAt,
		Metadata:     make(map[string]string),
	}
}

// Validate validates the approved client
func (c *ApprovedClient) Validate() error {
	if c.ClientID == "" {
		return ErrInvalidClientID.Wrap("client_id cannot be empty")
	}

	if c.Name == "" {
		return ErrInvalidClientID.Wrap("client name cannot be empty")
	}

	if len(c.PublicKey) == 0 {
		return ErrInvalidClientID.Wrap("client public key cannot be empty")
	}

	if c.Algorithm == "" {
		return ErrInvalidClientID.Wrap("client algorithm cannot be empty")
	}

	return nil
}
