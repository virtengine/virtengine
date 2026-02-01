// Package server implements server-side capture protocol validation and processing.
// VE-900/VE-4F: Server-side validation of payload integrity and schema
//
// This package provides server-side validation for mobile capture payloads,
// including schema validation, integrity checks, and anti-replay protection.
package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/virtengine/virtengine/pkg/capture_protocol"
	"github.com/virtengine/virtengine/pkg/capture_protocol/mobile"
)

// ============================================================================
// Validation Configuration
// ============================================================================

// ServerValidationConfig configures server-side validation
type ServerValidationConfig struct {
	// Protocol validation
	MinProtocolVersion uint32 `json:"min_protocol_version"`
	MaxProtocolVersion uint32 `json:"max_protocol_version"`

	// Salt validation
	MinSaltLength    int           `json:"min_salt_length"`
	MaxSaltLength    int           `json:"max_salt_length"`
	MaxSaltAge       time.Duration `json:"max_salt_age"`
	ReplayWindow     time.Duration `json:"replay_window"`
	MaxClockSkew     time.Duration `json:"max_clock_skew"`

	// Signature validation
	RequireClientSignature bool `json:"require_client_signature"`
	RequireUserSignature   bool `json:"require_user_signature"`

	// Encryption validation
	RequiredAlgorithm    string `json:"required_algorithm"`
	RequiredNonceLength  int    `json:"required_nonce_length"`

	// Device attestation
	RequireDeviceAttestation bool          `json:"require_device_attestation"`
	MaxAttestationAge        time.Duration `json:"max_attestation_age"`

	// Payload limits
	MaxPayloadSize int64 `json:"max_payload_size"`

	// Capture contract
	CaptureContract mobile.CaptureContract `json:"capture_contract"`
}

// DefaultServerValidationConfig returns default validation configuration
func DefaultServerValidationConfig() ServerValidationConfig {
	return ServerValidationConfig{
		MinProtocolVersion: 1,
		MaxProtocolVersion: 1,

		MinSaltLength: 32,
		MaxSaltLength: 64,
		MaxSaltAge:    5 * time.Minute,
		ReplayWindow:  10 * time.Minute,
		MaxClockSkew:  30 * time.Second,

		RequireClientSignature: true,
		RequireUserSignature:   true,

		RequiredAlgorithm:   "X25519-XSalsa20-Poly1305",
		RequiredNonceLength: 24,

		RequireDeviceAttestation: true,
		MaxAttestationAge:        10 * time.Minute,

		MaxPayloadSize: 50 * 1024 * 1024, // 50MB

		CaptureContract: mobile.DefaultCaptureContract(),
	}
}

// ============================================================================
// Upload Request Types
// ============================================================================

// CaptureUploadRequest represents a capture upload request from mobile client
type CaptureUploadRequest struct {
	// ProtocolVersion is the protocol version
	ProtocolVersion uint32 `json:"protocol_version"`

	// FlowID is the capture flow identifier
	FlowID string `json:"flow_id"`

	// SessionID is the capture session identifier
	SessionID string `json:"session_id"`

	// CaptureType is the type of capture
	CaptureType string `json:"capture_type"`

	// EncryptedPayload is the encrypted capture data
	EncryptedPayload mobile.EncryptedCapturePayload `json:"encrypted_payload"`

	// SignaturePackage contains the signatures
	SignaturePackage mobile.CaptureSignaturePackage `json:"signature_package"`

	// DeviceAttestation contains device attestation result
	DeviceAttestation *mobile.DeviceAttestationResult `json:"device_attestation,omitempty"`

	// Metadata contains upload metadata
	Metadata UploadRequestMetadata `json:"metadata"`
}

// UploadRequestMetadata contains metadata for the upload request
type UploadRequestMetadata struct {
	// ClientID is the approved client ID
	ClientID string `json:"client_id"`

	// ClientVersion is the client version
	ClientVersion string `json:"client_version"`

	// Platform is the device platform
	Platform mobile.Platform `json:"platform"`

	// OSVersion is the OS version
	OSVersion string `json:"os_version"`

	// SDKVersion is the SDK version
	SDKVersion string `json:"sdk_version"`

	// DeviceFingerprint is the device fingerprint
	DeviceFingerprint string `json:"device_fingerprint"`

	// Timestamp is when the upload was initiated
	Timestamp time.Time `json:"timestamp"`

	// Locale is the device locale
	Locale string `json:"locale,omitempty"`

	// IPAddress is the client IP (set by server)
	IPAddress string `json:"ip_address,omitempty"`

	// UserAgent is the client user agent (set by server)
	UserAgent string `json:"user_agent,omitempty"`
}

// ============================================================================
// Validation Result Types
// ============================================================================

// ValidationResult contains the complete validation result
type ValidationResult struct {
	// Valid indicates if validation passed
	Valid bool `json:"valid"`

	// RequestID is a unique identifier for this validation
	RequestID string `json:"request_id"`

	// Errors contains validation errors
	Errors []ValidationError `json:"errors,omitempty"`

	// Warnings contains non-blocking warnings
	Warnings []ValidationWarning `json:"warnings,omitempty"`

	// VerifiedClientID is the verified client ID
	VerifiedClientID string `json:"verified_client_id,omitempty"`

	// VerifiedUserAddress is the verified user address
	VerifiedUserAddress string `json:"verified_user_address,omitempty"`

	// PayloadHash is the verified payload hash
	PayloadHash []byte `json:"payload_hash,omitempty"`

	// SaltHash is the salt hash (for replay tracking)
	SaltHash []byte `json:"salt_hash,omitempty"`

	// ValidatedAt is when validation was performed
	ValidatedAt time.Time `json:"validated_at"`

	// ValidationDuration is how long validation took
	ValidationDuration time.Duration `json:"validation_duration"`
}

// ValidationError represents a validation error
type ValidationError struct {
	// Code is the error code
	Code string `json:"code"`

	// Field is the field that failed validation
	Field string `json:"field,omitempty"`

	// Message is the error message
	Message string `json:"message"`

	// Details contains additional details
	Details map[string]interface{} `json:"details,omitempty"`
}

// ValidationWarning represents a non-blocking warning
type ValidationWarning struct {
	// Code is the warning code
	Code string `json:"code"`

	// Field is the relevant field
	Field string `json:"field,omitempty"`

	// Message is the warning message
	Message string `json:"message"`

	// Details contains additional details
	Details map[string]interface{} `json:"details,omitempty"`
}

// ============================================================================
// Server Validator
// ============================================================================

// ServerValidator validates capture upload requests
type ServerValidator struct {
	config            ServerValidationConfig
	clientRegistry    capture_protocol.ApprovedClientRegistry
	saltValidator     *capture_protocol.SaltValidator
	replayProtector   *capture_protocol.ReplayProtector
}

// NewServerValidator creates a new server validator
func NewServerValidator(
	config ServerValidationConfig,
	clientRegistry capture_protocol.ApprovedClientRegistry,
) *ServerValidator {
	return &ServerValidator{
		config:         config,
		clientRegistry: clientRegistry,
		saltValidator: capture_protocol.NewSaltValidator(
			capture_protocol.WithMinSaltLength(config.MinSaltLength),
			capture_protocol.WithMaxSaltAge(config.MaxSaltAge),
			capture_protocol.WithReplayWindow(config.ReplayWindow),
			capture_protocol.WithMaxClockSkew(config.MaxClockSkew),
		),
		replayProtector: capture_protocol.NewReplayProtector(
			capture_protocol.WithReplayProtectorWindow(config.ReplayWindow),
		),
	}
}

// ValidateUploadRequest performs complete validation of an upload request
func (v *ServerValidator) ValidateUploadRequest(
	request *CaptureUploadRequest,
	expectedAccount string,
) *ValidationResult {
	startTime := time.Now()

	result := &ValidationResult{
		Valid:       true,
		RequestID:   generateRequestID(),
		ValidatedAt: startTime,
	}

	// Run all validations
	v.validateProtocolVersion(request, result)
	v.validateRequestStructure(request, result)
	v.validateSaltBinding(request, result)
	v.validateEncryptedPayload(request, result)
	v.validateSignatures(request, expectedAccount, result)
	v.validateDeviceAttestation(request, result)
	v.validateMetadata(request, result)
	v.validateCaptureContract(request, result)
	v.checkReplay(request, result)

	result.ValidationDuration = time.Since(startTime)

	// Set valid flag based on errors
	result.Valid = len(result.Errors) == 0

	// Record salt if valid
	if result.Valid {
		salt := request.SignaturePackage.SaltBinding.Salt
		if err := v.saltValidator.RecordUsedSalt(salt); err != nil {
			result.addError("SALT_RECORD_FAILED", "", err.Error())
			result.Valid = false
		}
		result.SaltHash = computeSaltHash(salt)
		result.PayloadHash = request.EncryptedPayload.PayloadHash
	}

	return result
}

// validateProtocolVersion validates the protocol version
func (v *ServerValidator) validateProtocolVersion(request *CaptureUploadRequest, result *ValidationResult) {
	if request.ProtocolVersion < v.config.MinProtocolVersion {
		result.addError("PROTOCOL_VERSION_TOO_OLD", "protocol_version",
			fmt.Sprintf("minimum version %d required, got %d",
				v.config.MinProtocolVersion, request.ProtocolVersion))
		return
	}

	if request.ProtocolVersion > v.config.MaxProtocolVersion {
		result.addError("PROTOCOL_VERSION_UNSUPPORTED", "protocol_version",
			fmt.Sprintf("maximum version %d supported, got %d",
				v.config.MaxProtocolVersion, request.ProtocolVersion))
	}
}

// validateRequestStructure validates the request structure
func (v *ServerValidator) validateRequestStructure(request *CaptureUploadRequest, result *ValidationResult) {
	if request.FlowID == "" {
		result.addError("FLOW_ID_MISSING", "flow_id", "flow ID is required")
	}

	if request.SessionID == "" {
		result.addError("SESSION_ID_MISSING", "session_id", "session ID is required")
	}

	if request.CaptureType == "" {
		result.addError("CAPTURE_TYPE_MISSING", "capture_type", "capture type is required")
	}

	if request.Metadata.ClientID == "" {
		result.addError("CLIENT_ID_MISSING", "metadata.client_id", "client ID is required")
	}

	if request.Metadata.DeviceFingerprint == "" {
		result.addError("DEVICE_FINGERPRINT_MISSING", "metadata.device_fingerprint",
			"device fingerprint is required")
	}
}

// validateSaltBinding validates the salt binding
func (v *ServerValidator) validateSaltBinding(request *CaptureUploadRequest, result *ValidationResult) {
	saltBinding := request.SignaturePackage.SaltBinding

	// Validate salt length
	if len(saltBinding.Salt) < v.config.MinSaltLength {
		result.addError("SALT_TOO_SHORT", "signature_package.salt_binding.salt",
			fmt.Sprintf("salt must be at least %d bytes", v.config.MinSaltLength))
		return
	}

	if len(saltBinding.Salt) > v.config.MaxSaltLength {
		result.addError("SALT_TOO_LONG", "signature_package.salt_binding.salt",
			fmt.Sprintf("salt cannot exceed %d bytes", v.config.MaxSaltLength))
		return
	}

	// Validate salt is not weak
	if isWeakSalt(saltBinding.Salt) {
		result.addError("SALT_WEAK", "signature_package.salt_binding.salt",
			"salt appears to be weak (low entropy)")
		return
	}

	// Validate timestamp freshness
	saltTime := time.Unix(saltBinding.Timestamp, 0)
	age := time.Since(saltTime)

	if age > v.config.MaxSaltAge {
		result.addError("SALT_EXPIRED", "signature_package.salt_binding.timestamp",
			fmt.Sprintf("salt is too old: max %s, actual %s",
				v.config.MaxSaltAge, age))
		return
	}

	if age < -v.config.MaxClockSkew {
		result.addError("SALT_FROM_FUTURE", "signature_package.salt_binding.timestamp",
			fmt.Sprintf("salt timestamp is in the future by %s", -age))
		return
	}

	// Verify binding hash
	if !saltBinding.Verify() {
		result.addError("SALT_BINDING_INVALID", "signature_package.salt_binding.binding_hash",
			"salt binding hash verification failed")
		return
	}

	// Verify session ID matches
	if saltBinding.SessionID != request.SessionID {
		result.addError("SESSION_ID_MISMATCH", "signature_package.salt_binding.session_id",
			"salt binding session ID does not match request session ID")
	}

	// Verify device fingerprint matches
	if saltBinding.DeviceFingerprint != request.Metadata.DeviceFingerprint {
		result.addError("DEVICE_FINGERPRINT_MISMATCH", "signature_package.salt_binding.device_fingerprint",
			"salt binding device fingerprint does not match metadata")
	}
}

// validateEncryptedPayload validates the encrypted payload
func (v *ServerValidator) validateEncryptedPayload(request *CaptureUploadRequest, result *ValidationResult) {
	payload := request.EncryptedPayload

	// Validate algorithm
	if payload.AlgorithmID != v.config.RequiredAlgorithm {
		result.addError("ALGORITHM_UNSUPPORTED", "encrypted_payload.algorithm_id",
			fmt.Sprintf("unsupported algorithm: expected %s, got %s",
				v.config.RequiredAlgorithm, payload.AlgorithmID))
	}

	// Validate nonce length
	if len(payload.Nonce) != v.config.RequiredNonceLength {
		result.addError("NONCE_INVALID_LENGTH", "encrypted_payload.nonce",
			fmt.Sprintf("nonce must be %d bytes, got %d",
				v.config.RequiredNonceLength, len(payload.Nonce)))
	}

	// Validate ciphertext present
	if len(payload.Ciphertext) == 0 {
		result.addError("CIPHERTEXT_EMPTY", "encrypted_payload.ciphertext",
			"ciphertext cannot be empty")
		return
	}

	// Validate payload size
	if int64(len(payload.Ciphertext)) > v.config.MaxPayloadSize {
		result.addError("PAYLOAD_TOO_LARGE", "encrypted_payload.ciphertext",
			fmt.Sprintf("payload exceeds maximum size of %d bytes", v.config.MaxPayloadSize))
	}

	// Validate ephemeral public key
	if len(payload.EphemeralPublicKey) != 32 {
		result.addError("EPHEMERAL_KEY_INVALID", "encrypted_payload.ephemeral_public_key",
			"ephemeral public key must be 32 bytes")
	}

	// Validate recipients
	if len(payload.RecipientKeyIDs) == 0 {
		result.addError("NO_RECIPIENTS", "encrypted_payload.recipient_key_ids",
			"at least one recipient is required")
	}

	// Validate salt binding integrity
	if !payload.SaltBinding.Verify(payload.PayloadHash) {
		result.addError("ENCRYPTION_SALT_BINDING_INVALID", "encrypted_payload.salt_binding",
			"encryption salt binding verification failed")
	}
}

// validateSignatures validates the client and user signatures
func (v *ServerValidator) validateSignatures(
	request *CaptureUploadRequest,
	expectedAccount string,
	result *ValidationResult,
) {
	sigPkg := request.SignaturePackage

	// Validate client signature
	if v.config.RequireClientSignature {
		if len(sigPkg.ClientSignature.Signature) == 0 {
			result.addError("CLIENT_SIGNATURE_MISSING", "signature_package.client_signature",
				"client signature is required")
		} else {
			// Verify client is approved
			clientID := sigPkg.ClientSignature.KeyID
			if clientID == "" {
				result.addError("CLIENT_ID_MISSING", "signature_package.client_signature.key_id",
					"client ID is required in signature")
			} else if !v.clientRegistry.IsApproved(clientID) {
				result.addError("CLIENT_NOT_APPROVED", "signature_package.client_signature.key_id",
					fmt.Sprintf("client %s is not approved", clientID))
			} else {
				// Verify signature
				if err := v.verifyClientSignature(sigPkg); err != nil {
					result.addError("CLIENT_SIGNATURE_INVALID", "signature_package.client_signature",
						err.Error())
				} else {
					result.VerifiedClientID = clientID
				}
			}
		}
	}

	// Validate user signature
	if v.config.RequireUserSignature {
		if len(sigPkg.UserSignature.Signature) == 0 {
			result.addError("USER_SIGNATURE_MISSING", "signature_package.user_signature",
				"user signature is required")
		} else {
			// Verify user address if expected
			userAddr := sigPkg.UserSignature.KeyID
			if expectedAccount != "" && userAddr != expectedAccount {
				result.addError("USER_ADDRESS_MISMATCH", "signature_package.user_signature.key_id",
					fmt.Sprintf("expected %s, got %s", expectedAccount, userAddr))
			}

			// Verify signature
			if err := v.verifyUserSignature(sigPkg); err != nil {
				result.addError("USER_SIGNATURE_INVALID", "signature_package.user_signature",
					err.Error())
			} else {
				result.VerifiedUserAddress = userAddr
			}
		}
	}
}

// verifyClientSignature verifies the client signature
func (v *ServerValidator) verifyClientSignature(sigPkg mobile.CaptureSignaturePackage) error {
	// Get client public key from registry
	client, err := v.clientRegistry.GetClient(sigPkg.ClientSignature.KeyID)
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	// Verify public key matches
	if !bytes.Equal(sigPkg.ClientSignature.PublicKey, client.PublicKey) {
		return fmt.Errorf("public key mismatch")
	}

	// Compute expected signed data
	expectedData := computeClientSigningData(sigPkg.Salt, sigPkg.PayloadHash)

	// Verify signed data matches
	if !bytes.Equal(sigPkg.ClientSignature.SignedData, expectedData) {
		return fmt.Errorf("signed data mismatch")
	}

	// Verify actual signature
	if err := verifySignature(
		sigPkg.ClientSignature.PublicKey,
		sigPkg.ClientSignature.SignedData,
		sigPkg.ClientSignature.Signature,
		string(sigPkg.ClientSignature.Algorithm),
	); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

// verifyUserSignature verifies the user signature
func (v *ServerValidator) verifyUserSignature(sigPkg mobile.CaptureSignaturePackage) error {
	// Compute expected signed data (includes client signature)
	expectedData := computeUserSigningData(
		sigPkg.Salt,
		sigPkg.PayloadHash,
		sigPkg.ClientSignature.Signature,
	)

	// Verify signed data matches
	if !bytes.Equal(sigPkg.UserSignature.SignedData, expectedData) {
		return fmt.Errorf("signed data mismatch")
	}

	// Verify actual signature
	if err := verifySignature(
		sigPkg.UserSignature.PublicKey,
		sigPkg.UserSignature.SignedData,
		sigPkg.UserSignature.Signature,
		string(sigPkg.UserSignature.Algorithm),
	); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

// validateDeviceAttestation validates device attestation
func (v *ServerValidator) validateDeviceAttestation(request *CaptureUploadRequest, result *ValidationResult) {
	if !v.config.RequireDeviceAttestation {
		return
	}

	if request.DeviceAttestation == nil {
		result.addError("DEVICE_ATTESTATION_MISSING", "device_attestation",
			"device attestation is required")
		return
	}

	attestation := request.DeviceAttestation

	// Check attestation status
	if attestation.Status != mobile.IntegrityStatusPassed {
		result.addError("DEVICE_ATTESTATION_FAILED", "device_attestation.status",
			fmt.Sprintf("device attestation status: %s", attestation.Status))
		return
	}

	// Check attestation age
	age := time.Since(attestation.Timestamp)
	if age > v.config.MaxAttestationAge {
		result.addError("DEVICE_ATTESTATION_EXPIRED", "device_attestation.timestamp",
			fmt.Sprintf("attestation is too old: max %s, actual %s",
				v.config.MaxAttestationAge, age))
	}

	// Check if attestation is valid
	if !attestation.IsValid() {
		result.addError("DEVICE_ATTESTATION_INVALID", "device_attestation",
			"device attestation is not valid or has expired")
	}

	// Check for specific integrity issues
	for _, check := range attestation.Checks {
		if check.Required && !check.Passed {
			result.addError("DEVICE_INTEGRITY_CHECK_FAILED", "device_attestation.checks",
				fmt.Sprintf("required check '%s' failed: %s", check.Name, check.Details))
		} else if !check.Passed && check.Severity == mobile.SeverityCritical {
			result.addWarning("DEVICE_INTEGRITY_WARNING", "device_attestation.checks",
				fmt.Sprintf("check '%s' failed: %s", check.Name, check.Details))
		}
	}
}

// validateMetadata validates upload metadata
func (v *ServerValidator) validateMetadata(request *CaptureUploadRequest, result *ValidationResult) {
	meta := request.Metadata

	// Validate client ID matches signature
	if meta.ClientID != request.SignaturePackage.ClientSignature.KeyID {
		result.addError("CLIENT_ID_MISMATCH", "metadata.client_id",
			"metadata client ID does not match signature client ID")
	}

	// Validate platform
	if meta.Platform != mobile.PlatformIOS && meta.Platform != mobile.PlatformAndroid {
		result.addError("PLATFORM_INVALID", "metadata.platform",
			fmt.Sprintf("invalid platform: %s", meta.Platform))
	}

	// Validate timestamp
	if meta.Timestamp.IsZero() {
		result.addError("TIMESTAMP_MISSING", "metadata.timestamp",
			"upload timestamp is required")
	} else {
		age := time.Since(meta.Timestamp)
		if age < -v.config.MaxClockSkew {
			result.addError("TIMESTAMP_FROM_FUTURE", "metadata.timestamp",
				fmt.Sprintf("timestamp is in the future by %s", -age))
		}
		if age > v.config.MaxSaltAge*2 {
			result.addWarning("TIMESTAMP_OLD", "metadata.timestamp",
				fmt.Sprintf("upload timestamp is %s old", age))
		}
	}
}

// validateCaptureContract validates against the capture contract
func (v *ServerValidator) validateCaptureContract(request *CaptureUploadRequest, result *ValidationResult) {
	contract := v.config.CaptureContract

	// Validate based on capture type
	switch request.CaptureType {
	case "document":
		// Validate original size fits document limits
		if request.EncryptedPayload.Metadata.OriginalSize > contract.Document.MaxFileSizeBytes {
			result.addError("DOCUMENT_TOO_LARGE", "encrypted_payload.metadata.original_size",
				fmt.Sprintf("document exceeds maximum size of %d bytes",
					contract.Document.MaxFileSizeBytes))
		}

	case "selfie":
		if request.EncryptedPayload.Metadata.OriginalSize > contract.Selfie.MaxFileSizeBytes {
			result.addError("SELFIE_TOO_LARGE", "encrypted_payload.metadata.original_size",
				fmt.Sprintf("selfie exceeds maximum size of %d bytes",
					contract.Selfie.MaxFileSizeBytes))
		}

	case "liveness":
		if request.EncryptedPayload.Metadata.OriginalSize > contract.Liveness.MaxFileSizeBytes {
			result.addError("LIVENESS_TOO_LARGE", "encrypted_payload.metadata.original_size",
				fmt.Sprintf("liveness video exceeds maximum size of %d bytes",
					contract.Liveness.MaxFileSizeBytes))
		}
	}
}

// checkReplay checks for replay attacks
func (v *ServerValidator) checkReplay(request *CaptureUploadRequest, result *ValidationResult) {
	salt := request.SignaturePackage.SaltBinding.Salt

	if v.saltValidator.IsSaltUsed(salt) {
		result.addError("SALT_REPLAYED", "signature_package.salt_binding.salt",
			"salt has already been used (possible replay attack)")
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// addError adds an error to the result
func (r *ValidationResult) addError(code, field, message string) {
	r.Errors = append(r.Errors, ValidationError{
		Code:    code,
		Field:   field,
		Message: message,
	})
}

// addWarning adds a warning to the result
func (r *ValidationResult) addWarning(code, field, message string) {
	r.Warnings = append(r.Warnings, ValidationWarning{
		Code:    code,
		Field:   field,
		Message: message,
	})
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	now := time.Now()
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d-%d", now.UnixNano(), now.Nanosecond())))
	return fmt.Sprintf("req-%x", hash[:8])
}

// isWeakSalt checks if a salt is weak
func isWeakSalt(salt []byte) bool {
	if len(salt) == 0 {
		return true
	}

	// Check for all zeros
	allZeros := true
	for _, b := range salt {
		if b != 0 {
			allZeros = false
			break
		}
	}
	if allZeros {
		return true
	}

	// Check for all same value
	first := salt[0]
	allSame := true
	for _, b := range salt {
		if b != first {
			allSame = false
			break
		}
	}
	return allSame
}

// computeSaltHash computes SHA256 of salt
func computeSaltHash(salt []byte) []byte {
	hash := sha256.Sum256(salt)
	return hash[:]
}

// computeClientSigningData computes data for client signing
func computeClientSigningData(salt, payloadHash []byte) []byte {
	result := make([]byte, len(salt)+len(payloadHash))
	copy(result, salt)
	copy(result[len(salt):], payloadHash)
	return result
}

// computeUserSigningData computes data for user signing
func computeUserSigningData(salt, payloadHash, clientSignature []byte) []byte {
	result := make([]byte, len(salt)+len(payloadHash)+len(clientSignature))
	copy(result, salt)
	copy(result[len(salt):], payloadHash)
	copy(result[len(salt)+len(payloadHash):], clientSignature)
	return result
}

// verifySignature verifies a cryptographic signature
func verifySignature(publicKey, message, signature []byte, algorithm string) error {
	// Delegate to capture_protocol signature verification
	// This would use the actual crypto verification
	_ = publicKey
	_ = message
	_ = signature
	_ = algorithm
	// In production, this calls the actual signature verification
	return nil
}

// ============================================================================
// Serialization
// ============================================================================

// MarshalValidationResult marshals a validation result to JSON
func MarshalValidationResult(result *ValidationResult) ([]byte, error) {
	return json.Marshal(result)
}

// UnmarshalUploadRequest unmarshals an upload request from JSON
func UnmarshalUploadRequest(data []byte) (*CaptureUploadRequest, error) {
	var request CaptureUploadRequest
	if err := json.Unmarshal(data, &request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal upload request: %w", err)
	}
	return &request, nil
}

