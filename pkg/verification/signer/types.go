// Package signer provides the verification attestation signing service.
//
// This package implements the SignerService interface for creating verifiable
// attestations with support for key rotation, HSM/Vault integration, and
// audit logging.
//
// Task Reference: VE-2B - Verification Shared Infrastructure
package signer

import (
	"context"
	"time"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Signer Service Interface
// ============================================================================

// SignerService defines the interface for the verification signer service.
// It handles attestation signing with key rotation and audit logging.
type SignerService interface {
	// SignAttestation signs a verification attestation with the active key.
	SignAttestation(ctx context.Context, attestation *veidtypes.VerificationAttestation) error

	// VerifyAttestation verifies an attestation signature.
	VerifyAttestation(ctx context.Context, attestation *veidtypes.VerificationAttestation) (bool, error)

	// GetActiveKey returns the currently active signing key info.
	GetActiveKey(ctx context.Context) (*veidtypes.SignerKeyInfo, error)

	// GetKeyByID returns key info by key ID.
	GetKeyByID(ctx context.Context, keyID string) (*veidtypes.SignerKeyInfo, error)

	// GetKeyByFingerprint returns key info by fingerprint.
	GetKeyByFingerprint(ctx context.Context, fingerprint string) (*veidtypes.SignerKeyInfo, error)

	// ListKeys returns all keys for this signer.
	ListKeys(ctx context.Context) ([]*veidtypes.SignerKeyInfo, error)

	// RotateKey initiates a key rotation to a new key.
	RotateKey(ctx context.Context, req *KeyRotationRequest) (*veidtypes.KeyRotationRecord, error)

	// CompleteRotation completes a key rotation.
	CompleteRotation(ctx context.Context, rotationID string) error

	// RevokeKey revokes a key immediately (emergency only).
	RevokeKey(ctx context.Context, keyID string, reason veidtypes.KeyRevocationReason) error

	// GetRotationStatus returns the status of a key rotation.
	GetRotationStatus(ctx context.Context, rotationID string) (*veidtypes.KeyRotationRecord, error)

	// GetSignerInfo returns information about this signer.
	GetSignerInfo(ctx context.Context) (*veidtypes.SignerRegistryEntry, error)

	// HealthCheck returns the health status of the signer service.
	HealthCheck(ctx context.Context) (*HealthStatus, error)

	// Close closes the signer service and releases resources.
	Close() error
}

// ============================================================================
// Request/Response Types
// ============================================================================

// KeyRotationRequest contains parameters for initiating key rotation.
type KeyRotationRequest struct {
	// Reason for the rotation
	Reason veidtypes.KeyRevocationReason `json:"reason"`

	// InitiatedBy indicates who initiated the rotation
	InitiatedBy string `json:"initiated_by"`

	// OverrideOverlapSeconds overrides the default overlap period
	OverrideOverlapSeconds *int64 `json:"override_overlap_seconds,omitempty"`

	// NewKeyAlgorithm specifies the algorithm for the new key (optional)
	NewKeyAlgorithm veidtypes.AttestationProofType `json:"new_key_algorithm,omitempty"`

	// Notes contains optional notes about the rotation
	Notes string `json:"notes,omitempty"`

	// Emergency indicates if this is an emergency rotation
	Emergency bool `json:"emergency,omitempty"`
}

// HealthStatus represents the health status of the signer service.
type HealthStatus struct {
	// Healthy indicates if the service is healthy
	Healthy bool `json:"healthy"`

	// Status is a human-readable status message
	Status string `json:"status"`

	// Details contains detailed health information
	Details map[string]interface{} `json:"details,omitempty"`

	// Timestamp is when the health check was performed
	Timestamp time.Time `json:"timestamp"`

	// ActiveKeyID is the current active key ID
	ActiveKeyID string `json:"active_key_id,omitempty"`

	// KeyExpiresAt is when the active key expires
	KeyExpiresAt *time.Time `json:"key_expires_at,omitempty"`

	// RotationPending indicates if a rotation is in progress
	RotationPending bool `json:"rotation_pending"`

	// KeyAge is the age of the active key
	KeyAge time.Duration `json:"key_age,omitempty"`

	// Warnings contains any health warnings
	Warnings []string `json:"warnings,omitempty"`
}

// SignerConfig contains configuration for the signer service.
type SignerConfig struct {
	// SignerID is the unique identifier for this signer
	SignerID string `json:"signer_id"`

	// SignerName is the human-readable name
	SignerName string `json:"signer_name"`

	// ValidatorAddress is the associated validator address (if applicable)
	ValidatorAddress string `json:"validator_address,omitempty"`

	// KeyStorageType specifies the key storage backend
	KeyStorageType KeyStorageType `json:"key_storage_type"`

	// KeyStorageConfig contains backend-specific configuration
	KeyStorageConfig map[string]string `json:"key_storage_config,omitempty"`

	// DefaultAlgorithm is the default signing algorithm
	DefaultAlgorithm veidtypes.AttestationProofType `json:"default_algorithm"`

	// KeyPolicy is the key rotation policy
	KeyPolicy veidtypes.SignerKeyPolicy `json:"key_policy"`

	// AuditLogEnabled enables audit logging
	AuditLogEnabled bool `json:"audit_log_enabled"`

	// MetricsEnabled enables Prometheus metrics
	MetricsEnabled bool `json:"metrics_enabled"`

	// ServiceEndpoint is the optional service endpoint URL
	ServiceEndpoint string `json:"service_endpoint,omitempty"`
}

// KeyStorageType identifies the key storage backend.
type KeyStorageType string

const (
	// KeyStorageMemory stores keys in memory (testing only)
	KeyStorageMemory KeyStorageType = "memory"

	// KeyStorageFile stores keys in encrypted files
	KeyStorageFile KeyStorageType = "file"

	// KeyStorageVault stores keys in HashiCorp Vault
	KeyStorageVault KeyStorageType = "vault"

	// KeyStorageHSM stores keys in an HSM
	KeyStorageHSM KeyStorageType = "hsm"

	// KeyStorageKMS stores keys in cloud KMS (AWS/GCP/Azure)
	KeyStorageKMS KeyStorageType = "kms"
)

// DefaultSignerConfig returns the default signer configuration.
func DefaultSignerConfig() SignerConfig {
	return SignerConfig{
		SignerID:         "",
		SignerName:       "verification-signer",
		KeyStorageType:   KeyStorageMemory,
		DefaultAlgorithm: veidtypes.ProofTypeEd25519,
		KeyPolicy:        veidtypes.DefaultSignerKeyPolicy(),
		AuditLogEnabled:  true,
		MetricsEnabled:   true,
	}
}

// Validate validates the signer configuration.
func (c *SignerConfig) Validate() error {
	if c.SignerID == "" {
		return ErrInvalidConfig.Wrap("signer_id is required")
	}

	if c.SignerName == "" {
		return ErrInvalidConfig.Wrap("signer_name is required")
	}

	if !veidtypes.IsValidProofType(c.DefaultAlgorithm) {
		return ErrInvalidConfig.Wrapf("invalid default_algorithm: %s", c.DefaultAlgorithm)
	}

	if err := c.KeyPolicy.Validate(); err != nil {
		return ErrInvalidConfig.Wrapf("invalid key_policy: %v", err)
	}

	return nil
}

// ============================================================================
// Signing Request/Response
// ============================================================================

// SignRequest contains parameters for a signing operation.
type SignRequest struct {
	// Data is the data to sign
	Data []byte `json:"data"`

	// KeyID optionally specifies which key to use (defaults to active)
	KeyID string `json:"key_id,omitempty"`

	// Purpose describes the purpose of the signature (for audit)
	Purpose string `json:"purpose,omitempty"`

	// RequestID is an optional request identifier for correlation
	RequestID string `json:"request_id,omitempty"`
}

// SignResponse contains the result of a signing operation.
type SignResponse struct {
	// Signature is the cryptographic signature
	Signature []byte `json:"signature"`

	// KeyID is the key that was used for signing
	KeyID string `json:"key_id"`

	// KeyFingerprint is the fingerprint of the signing key
	KeyFingerprint string `json:"key_fingerprint"`

	// Algorithm is the signature algorithm used
	Algorithm veidtypes.AttestationProofType `json:"algorithm"`

	// SignedAt is when the signature was created
	SignedAt time.Time `json:"signed_at"`
}

// VerifyRequest contains parameters for a verification operation.
type VerifyRequest struct {
	// Data is the original data that was signed
	Data []byte `json:"data"`

	// Signature is the signature to verify
	Signature []byte `json:"signature"`

	// KeyFingerprint identifies the key to use for verification
	KeyFingerprint string `json:"key_fingerprint"`

	// AllowRotatingKeys allows verification with rotating keys
	AllowRotatingKeys bool `json:"allow_rotating_keys"`
}

// VerifyResponse contains the result of a verification operation.
type VerifyResponse struct {
	// Valid indicates if the signature is valid
	Valid bool `json:"valid"`

	// KeyID is the key that was used for verification
	KeyID string `json:"key_id"`

	// KeyState is the current state of the key
	KeyState veidtypes.SignerKeyState `json:"key_state"`

	// VerifiedAt is when the verification was performed
	VerifiedAt time.Time `json:"verified_at"`

	// Error contains an error message if verification failed
	Error string `json:"error,omitempty"`
}

