package mobile

import (
	"crypto/sha256"
	"time"
)

// ============================================================================
// Approved-Client Key Signing Integration
// VE-900: Integration with VE-201 approved client system
// ============================================================================

// ClientSigningConfig configures client key signing
type ClientSigningConfig struct {
	// ClientID is the approved client identifier
	ClientID string `json:"client_id"`

	// ClientVersion is the app version
	ClientVersion string `json:"client_version"`

	// KeyType is the signing key type
	KeyType SigningKeyType `json:"key_type"`

	// UseSecureEnclave uses hardware-backed keys (iOS Secure Enclave / Android StrongBox)
	UseSecureEnclave bool `json:"use_secure_enclave"`

	// RequireUserPresence requires biometric for signing
	RequireUserPresence bool `json:"require_user_presence"`
}

// SigningKeyType represents the signing key algorithm
type SigningKeyType string

const (
	// KeyTypeEd25519 is Ed25519 signing
	KeyTypeEd25519 SigningKeyType = "ed25519"

	// KeyTypeSecp256k1 is secp256k1 signing (Cosmos-compatible)
	KeyTypeSecp256k1 SigningKeyType = "secp256k1"
)

// ============================================================================
// Mobile Client Key Provider Interface
// ============================================================================

// MobileClientKeyProvider provides client signing operations on mobile
type MobileClientKeyProvider interface {
	// GetClientID returns the registered client ID
	GetClientID() string

	// GetClientVersion returns the app version
	GetClientVersion() string

	// GetPublicKey returns the client's public key
	GetPublicKey() ([]byte, error)

	// GetKeyType returns the signing key type
	GetKeyType() SigningKeyType

	// Sign signs data with the client key
	Sign(data []byte) ([]byte, error)

	// IsHardwareBacked returns true if key is hardware-backed
	IsHardwareBacked() bool

	// RequiresUserPresence returns true if signing requires biometric
	RequiresUserPresence() bool
}

// MobileUserKeyProvider provides user signing operations on mobile
type MobileUserKeyProvider interface {
	// GetAccountAddress returns the user's account address
	GetAccountAddress() string

	// GetPublicKey returns the user's public key
	GetPublicKey() ([]byte, error)

	// GetKeyType returns the signing key type
	GetKeyType() SigningKeyType

	// Sign signs data with the user key
	Sign(data []byte) ([]byte, error)

	// IsHardwareBacked returns true if key is hardware-backed
	IsHardwareBacked() bool

	// RequiresUserPresence returns true if signing requires biometric
	RequiresUserPresence() bool
}

// ============================================================================
// Capture Signature Package
// ============================================================================

// CaptureSignaturePackage contains all signatures for a capture upload
type CaptureSignaturePackage struct {
	// ProtocolVersion is the protocol version
	ProtocolVersion uint32 `json:"protocol_version"`

	// Salt is the per-upload unique salt
	Salt []byte `json:"salt"`

	// SaltBinding binds salt to device/session/time
	SaltBinding MobileSaltBinding `json:"salt_binding"`

	// PayloadHash is SHA256 of the encrypted payload
	PayloadHash []byte `json:"payload_hash"`

	// ClientSignature is the approved client signature
	ClientSignature MobileSignature `json:"client_signature"`

	// UserSignature is the user's signature
	UserSignature MobileSignature `json:"user_signature"`

	// CaptureMetadata contains capture information
	CaptureMetadata MobileCaptureMetadata `json:"capture_metadata"`

	// OriginProof proves live camera capture
	OriginProof *CaptureOriginProof `json:"origin_proof,omitempty"`

	// Timestamp is when package was created
	Timestamp time.Time `json:"timestamp"`
}

// MobileSaltBinding represents salt binding for mobile captures
type MobileSaltBinding struct {
	// Salt is the unique salt
	Salt []byte `json:"salt"`

	// DeviceFingerprint is the device fingerprint hash
	DeviceFingerprint string `json:"device_fingerprint"`

	// SessionID is the capture session ID
	SessionID string `json:"session_id"`

	// Timestamp is Unix timestamp
	Timestamp int64 `json:"timestamp"`

	// BindingHash is SHA256(salt || device || session || timestamp)
	BindingHash []byte `json:"binding_hash"`
}

// ComputeBindingHash computes the salt binding hash
func (sb *MobileSaltBinding) ComputeBindingHash() []byte {
	h := sha256.New()
	h.Write(sb.Salt)
	h.Write([]byte(sb.DeviceFingerprint))
	h.Write([]byte(sb.SessionID))
	h.Write(int64ToBytes(sb.Timestamp))
	return h.Sum(nil)
}

// Verify verifies the binding hash
func (sb *MobileSaltBinding) Verify() bool {
	expected := sb.ComputeBindingHash()
	return constantTimeEqual(sb.BindingHash, expected)
}

// MobileSignature represents a signature in the mobile context
type MobileSignature struct {
	// PublicKey is the signer's public key
	PublicKey []byte `json:"public_key"`

	// Signature is the cryptographic signature
	Signature []byte `json:"signature"`

	// Algorithm is the signature algorithm
	Algorithm SigningKeyType `json:"algorithm"`

	// KeyID is client ID or user address
	KeyID string `json:"key_id"`

	// SignedData is the data that was signed
	SignedData []byte `json:"signed_data"`

	// IsHardwareBacked indicates if key was hardware-backed
	IsHardwareBacked bool `json:"is_hardware_backed"`

	// Timestamp is when signature was created
	Timestamp time.Time `json:"timestamp"`
}

// MobileCaptureMetadata contains mobile capture metadata
type MobileCaptureMetadata struct {
	// DeviceFingerprint is device identification hash
	DeviceFingerprint string `json:"device_fingerprint"`

	// ClientID is the approved client ID
	ClientID string `json:"client_id"`

	// ClientVersion is the app version
	ClientVersion string `json:"client_version"`

	// Platform is ios or android
	Platform Platform `json:"platform"`

	// OSVersion is the OS version
	OSVersion string `json:"os_version"`

	// SessionID is the capture session ID
	SessionID string `json:"session_id"`

	// CaptureType is document or selfie
	CaptureType CaptureType `json:"capture_type"`

	// DocumentType is the document type (if applicable)
	DocumentType DocumentType `json:"document_type,omitempty"`

	// DocumentSide is the document side (if applicable)
	DocumentSide DocumentSide `json:"document_side,omitempty"`

	// QualityScore is the quality validation score
	QualityScore int `json:"quality_score"`

	// LivenessVerified indicates if liveness was verified
	LivenessVerified bool `json:"liveness_verified"`

	// LivenessScore is the liveness confidence score
	LivenessScore float64 `json:"liveness_score,omitempty"`

	// GalleryBlocked indicates gallery uploads were blocked
	GalleryBlocked bool `json:"gallery_blocked"`

	// CaptureTimestamp is when image was captured
	CaptureTimestamp int64 `json:"capture_timestamp"`
}

// ============================================================================
// Signature Builder
// ============================================================================

// CaptureSignatureBuilder builds capture signature packages
type CaptureSignatureBuilder struct {
	clientKeyProvider MobileClientKeyProvider
	userKeyProvider   MobileUserKeyProvider
	config            SignatureBuilderConfig
}

// SignatureBuilderConfig configures the signature builder
type SignatureBuilderConfig struct {
	// ProtocolVersion is the protocol version
	ProtocolVersion uint32

	// RequireClientSignature requires client signature
	RequireClientSignature bool

	// RequireUserSignature requires user signature
	RequireUserSignature bool

	// RequireHardwareBackedKeys requires hardware-backed keys
	RequireHardwareBackedKeys bool
}

// DefaultSignatureBuilderConfig returns default config
func DefaultSignatureBuilderConfig() SignatureBuilderConfig {
	return SignatureBuilderConfig{
		ProtocolVersion:           1,
		RequireClientSignature:    true,
		RequireUserSignature:      true,
		RequireHardwareBackedKeys: false, // Not all devices support
	}
}

// NewCaptureSignatureBuilder creates a new signature builder
func NewCaptureSignatureBuilder(
	clientKeyProvider MobileClientKeyProvider,
	userKeyProvider MobileUserKeyProvider,
	config SignatureBuilderConfig,
) *CaptureSignatureBuilder {
	return &CaptureSignatureBuilder{
		clientKeyProvider: clientKeyProvider,
		userKeyProvider:   userKeyProvider,
		config:            config,
	}
}

// BuildSignatureParams contains parameters for building signatures
type BuildSignatureParams struct {
	// Salt is the per-upload salt
	Salt []byte

	// PayloadHash is the encrypted payload hash
	PayloadHash []byte

	// DeviceFingerprint is the device fingerprint
	DeviceFingerprint DeviceFingerprint

	// SessionID is the capture session ID
	SessionID string

	// CaptureResult is the capture result
	CaptureResult *CaptureResult

	// OriginProof is the origin proof
	OriginProof *CaptureOriginProof
}

// BuildSignaturePackage creates a complete signature package
func (b *CaptureSignatureBuilder) BuildSignaturePackage(
	params BuildSignatureParams,
) (*CaptureSignaturePackage, error) {
	now := time.Now()

	// Create salt binding
	saltBinding := MobileSaltBinding{
		Salt:              params.Salt,
		DeviceFingerprint: params.DeviceFingerprint.FingerprintHash,
		SessionID:         params.SessionID,
		Timestamp:         now.Unix(),
	}
	saltBinding.BindingHash = saltBinding.ComputeBindingHash()

	// Compute client signing data: salt || payload_hash
	clientSignedData := computeClientSigningData(params.Salt, params.PayloadHash)

	// Sign with client key
	clientSig, err := b.signWithClient(clientSignedData)
	if err != nil {
		return nil, ErrClientSigningFailed.Wrap(err)
	}

	// Compute user signing data: salt || payload_hash || client_signature
	userSignedData := computeUserSigningData(params.Salt, params.PayloadHash, clientSig.Signature)

	// Sign with user key
	userSig, err := b.signWithUser(userSignedData)
	if err != nil {
		return nil, ErrUserSigningFailed.Wrap(err)
	}

	// Build metadata
	metadata := MobileCaptureMetadata{
		DeviceFingerprint: params.DeviceFingerprint.FingerprintHash,
		ClientID:          b.clientKeyProvider.GetClientID(),
		ClientVersion:     b.clientKeyProvider.GetClientVersion(),
		Platform:          params.DeviceFingerprint.Platform,
		SessionID:         params.SessionID,
		CaptureTimestamp:  now.Unix(),
		GalleryBlocked:    true,
	}

	if params.CaptureResult != nil {
		metadata.CaptureType = params.CaptureResult.CaptureType
		metadata.QualityScore = params.CaptureResult.QualityResult.OverallScore
		if params.CaptureResult.LivenessResult != nil {
			metadata.LivenessVerified = params.CaptureResult.LivenessResult.Passed
			metadata.LivenessScore = params.CaptureResult.LivenessResult.Confidence
		}
	}

	return &CaptureSignaturePackage{
		ProtocolVersion: b.config.ProtocolVersion,
		Salt:            params.Salt,
		SaltBinding:     saltBinding,
		PayloadHash:     params.PayloadHash,
		ClientSignature: *clientSig,
		UserSignature:   *userSig,
		CaptureMetadata: metadata,
		OriginProof:     params.OriginProof,
		Timestamp:       now,
	}, nil
}

// signWithClient creates client signature
func (b *CaptureSignatureBuilder) signWithClient(data []byte) (*MobileSignature, error) {
	pubKey, err := b.clientKeyProvider.GetPublicKey()
	if err != nil {
		return nil, err
	}

	sig, err := b.clientKeyProvider.Sign(data)
	if err != nil {
		return nil, err
	}

	return &MobileSignature{
		PublicKey:        pubKey,
		Signature:        sig,
		Algorithm:        b.clientKeyProvider.GetKeyType(),
		KeyID:            b.clientKeyProvider.GetClientID(),
		SignedData:       data,
		IsHardwareBacked: b.clientKeyProvider.IsHardwareBacked(),
		Timestamp:        time.Now(),
	}, nil
}

// signWithUser creates user signature
func (b *CaptureSignatureBuilder) signWithUser(data []byte) (*MobileSignature, error) {
	pubKey, err := b.userKeyProvider.GetPublicKey()
	if err != nil {
		return nil, err
	}

	sig, err := b.userKeyProvider.Sign(data)
	if err != nil {
		return nil, err
	}

	return &MobileSignature{
		PublicKey:        pubKey,
		Signature:        sig,
		Algorithm:        b.userKeyProvider.GetKeyType(),
		KeyID:            b.userKeyProvider.GetAccountAddress(),
		SignedData:       data,
		IsHardwareBacked: b.userKeyProvider.IsHardwareBacked(),
		Timestamp:        time.Now(),
	}, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// computeClientSigningData computes data for client to sign
func computeClientSigningData(salt, payloadHash []byte) []byte {
	result := make([]byte, len(salt)+len(payloadHash))
	copy(result, salt)
	copy(result[len(salt):], payloadHash)
	return result
}

// computeUserSigningData computes data for user to sign
func computeUserSigningData(salt, payloadHash, clientSignature []byte) []byte {
	result := make([]byte, len(salt)+len(payloadHash)+len(clientSignature))
	copy(result, salt)
	copy(result[len(salt):], payloadHash)
	copy(result[len(salt)+len(payloadHash):], clientSignature)
	return result
}

// int64ToBytes converts int64 to big-endian bytes
func int64ToBytes(n int64) []byte {
	b := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		b[i] = byte(n & 0xff)
		n >>= 8
	}
	return b
}

// constantTimeEqual performs constant-time comparison
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
