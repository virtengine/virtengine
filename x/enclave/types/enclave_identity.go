package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// TEEType represents the type of Trusted Execution Environment
type TEEType string

const (
	// TEETypeSGX is Intel SGX
	TEETypeSGX TEEType = "SGX"

	// TEETypeSEVSNP is AMD SEV-SNP
	TEETypeSEVSNP TEEType = "SEV-SNP"

	// TEETypeNitro is AWS Nitro Enclaves
	TEETypeNitro TEEType = "NITRO"

	// TEETypeTrustZone is ARM TrustZone (future)
	TEETypeTrustZone TEEType = "TRUSTZONE"
)

// AllTEETypes returns all valid TEE types
func AllTEETypes() []TEEType {
	return []TEEType{TEETypeSGX, TEETypeSEVSNP, TEETypeNitro, TEETypeTrustZone}
}

// IsValidTEEType checks if a TEE type is valid
func IsValidTEEType(teeType TEEType) bool {
	for _, t := range AllTEETypes() {
		if t == teeType {
			return true
		}
	}
	return false
}

// EnclaveIdentity represents a validator's enclave identity record
type EnclaveIdentity struct {
	// ValidatorAddress is the validator operator address
	ValidatorAddress string `json:"validator_address"`

	// TEEType is the type of TEE (SGX, SEV-SNP, NITRO)
	TEEType TEEType `json:"tee_type"`

	// MeasurementHash is the enclave measurement (MRENCLAVE for SGX)
	MeasurementHash []byte `json:"measurement_hash"`

	// SignerHash is the signer measurement (MRSIGNER for SGX)
	SignerHash []byte `json:"signer_hash,omitempty"`

	// EncryptionPubKey is the enclave's public key for encryption
	EncryptionPubKey []byte `json:"encryption_pub_key"`

	// SigningPubKey is the enclave's public key for signing attestations
	SigningPubKey []byte `json:"signing_pub_key"`

	// AttestationQuote is the raw attestation quote from the TEE
	AttestationQuote []byte `json:"attestation_quote"`

	// AttestationChain is the certificate chain for attestation verification
	AttestationChain [][]byte `json:"attestation_chain"`

	// ISVProdID is the Independent Software Vendor Product ID
	ISVProdID uint16 `json:"isv_prod_id"`

	// ISVSVN is the Independent Software Vendor Security Version Number
	ISVSVN uint16 `json:"isv_svn"`

	// QuoteVersion is the attestation quote format version
	QuoteVersion uint32 `json:"quote_version"`

	// DebugMode indicates if the enclave is in debug mode (must be false for production)
	DebugMode bool `json:"debug_mode"`

	// Epoch is the registration epoch
	Epoch uint64 `json:"epoch"`

	// ExpiryHeight is the block height when this identity expires
	ExpiryHeight int64 `json:"expiry_height"`

	// RegisteredAt is the timestamp when this identity was registered
	RegisteredAt time.Time `json:"registered_at"`

	// UpdatedAt is the timestamp when this identity was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// Status is the current status of the enclave identity
	Status EnclaveIdentityStatus `json:"status"`
}

// EnclaveIdentityStatus represents the status of an enclave identity
type EnclaveIdentityStatus string

const (
	// EnclaveIdentityStatusActive indicates the enclave identity is active
	EnclaveIdentityStatusActive EnclaveIdentityStatus = "active"

	// EnclaveIdentityStatusPending indicates the enclave identity is pending verification
	EnclaveIdentityStatusPending EnclaveIdentityStatus = "pending"

	// EnclaveIdentityStatusExpired indicates the enclave identity has expired
	EnclaveIdentityStatusExpired EnclaveIdentityStatus = "expired"

	// EnclaveIdentityStatusRevoked indicates the enclave identity has been revoked
	EnclaveIdentityStatusRevoked EnclaveIdentityStatus = "revoked"

	// EnclaveIdentityStatusRotating indicates key rotation is in progress
	EnclaveIdentityStatusRotating EnclaveIdentityStatus = "rotating"
)

// Validate validates the enclave identity
func (e *EnclaveIdentity) Validate() error {
	if e.ValidatorAddress == "" {
		return ErrInvalidEnclaveIdentity.Wrap("validator address cannot be empty")
	}

	if !IsValidTEEType(e.TEEType) {
		return ErrInvalidEnclaveIdentity.Wrapf("invalid TEE type: %s", e.TEEType)
	}

	if len(e.MeasurementHash) == 0 {
		return ErrInvalidEnclaveIdentity.Wrap("measurement hash cannot be empty")
	}

	if len(e.MeasurementHash) != 32 {
		return ErrInvalidEnclaveIdentity.Wrapf("measurement hash must be 32 bytes, got %d", len(e.MeasurementHash))
	}

	if len(e.EncryptionPubKey) == 0 {
		return ErrInvalidEnclaveIdentity.Wrap("encryption public key cannot be empty")
	}

	if len(e.SigningPubKey) == 0 {
		return ErrInvalidEnclaveIdentity.Wrap("signing public key cannot be empty")
	}

	if len(e.AttestationQuote) == 0 {
		return ErrInvalidEnclaveIdentity.Wrap("attestation quote cannot be empty")
	}

	if e.DebugMode {
		return ErrInvalidEnclaveIdentity.Wrap("debug mode must be disabled for production enclaves")
	}

	if e.ExpiryHeight <= 0 {
		return ErrInvalidEnclaveIdentity.Wrap("expiry height must be positive")
	}

	return nil
}

// KeyFingerprint returns the fingerprint of the encryption public key
func (e *EnclaveIdentity) KeyFingerprint() string {
	h := sha256.Sum256(e.EncryptionPubKey)
	return hex.EncodeToString(h[:])
}

// IsExpired checks if the enclave identity has expired at the given block height
func (e *EnclaveIdentity) IsExpired(currentHeight int64) bool {
	return currentHeight >= e.ExpiryHeight
}

// MeasurementRecord represents an approved enclave measurement in the allowlist
type MeasurementRecord struct {
	// MeasurementHash is the enclave measurement hash
	MeasurementHash []byte `json:"measurement_hash"`

	// TEEType is the TEE type this measurement is for
	TEEType TEEType `json:"tee_type"`

	// Description is a human-readable description
	Description string `json:"description"`

	// MinISVSVN is the minimum required security version
	MinISVSVN uint16 `json:"min_isv_svn"`

	// AddedAt is when this measurement was added
	AddedAt time.Time `json:"added_at"`

	// AddedByProposal is the governance proposal ID that added this measurement
	AddedByProposal uint64 `json:"added_by_proposal"`

	// ExpiryHeight is when this measurement expires (0 for no expiry)
	ExpiryHeight int64 `json:"expiry_height,omitempty"`

	// Revoked indicates if this measurement has been revoked
	Revoked bool `json:"revoked"`

	// RevokedAt is when this measurement was revoked (if applicable)
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	// RevokedByProposal is the governance proposal that revoked this (if applicable)
	RevokedByProposal uint64 `json:"revoked_by_proposal,omitempty"`
}

// Validate validates the measurement record
func (m *MeasurementRecord) Validate() error {
	if len(m.MeasurementHash) == 0 {
		return ErrInvalidMeasurement.Wrap("measurement hash cannot be empty")
	}

	if len(m.MeasurementHash) != 32 {
		return ErrInvalidMeasurement.Wrapf("measurement hash must be 32 bytes, got %d", len(m.MeasurementHash))
	}

	if !IsValidTEEType(m.TEEType) {
		return ErrInvalidMeasurement.Wrapf("invalid TEE type: %s", m.TEEType)
	}

	if m.Description == "" {
		return ErrInvalidMeasurement.Wrap("description cannot be empty")
	}

	return nil
}

// IsValid checks if the measurement is valid (not revoked and not expired)
func (m *MeasurementRecord) IsValid(currentHeight int64) bool {
	if m.Revoked {
		return false
	}
	if m.ExpiryHeight > 0 && currentHeight >= m.ExpiryHeight {
		return false
	}
	return true
}

// MeasurementHashHex returns the measurement hash as a hex string
func (m *MeasurementRecord) MeasurementHashHex() string {
	return hex.EncodeToString(m.MeasurementHash)
}

// KeyRotationRecord represents a key rotation event
type KeyRotationRecord struct {
	// ValidatorAddress is the validator operator address
	ValidatorAddress string `json:"validator_address"`

	// Epoch is the epoch when rotation was initiated
	Epoch uint64 `json:"epoch"`

	// OldKeyFingerprint is the fingerprint of the old key
	OldKeyFingerprint string `json:"old_key_fingerprint"`

	// NewKeyFingerprint is the fingerprint of the new key
	NewKeyFingerprint string `json:"new_key_fingerprint"`

	// OverlapStartHeight is when both keys become valid
	OverlapStartHeight int64 `json:"overlap_start_height"`

	// OverlapEndHeight is when the old key becomes invalid
	OverlapEndHeight int64 `json:"overlap_end_height"`

	// InitiatedAt is when the rotation was initiated
	InitiatedAt time.Time `json:"initiated_at"`

	// CompletedAt is when the rotation was completed (old key invalidated)
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Status is the current status of the rotation
	Status KeyRotationStatus `json:"status"`
}

// KeyRotationStatus represents the status of a key rotation
type KeyRotationStatus string

const (
	// KeyRotationStatusPending indicates rotation is pending
	KeyRotationStatusPending KeyRotationStatus = "pending"

	// KeyRotationStatusActive indicates rotation is active (overlap period)
	KeyRotationStatusActive KeyRotationStatus = "active"

	// KeyRotationStatusCompleted indicates rotation is completed
	KeyRotationStatusCompleted KeyRotationStatus = "completed"

	// KeyRotationStatusCancelled indicates rotation was cancelled
	KeyRotationStatusCancelled KeyRotationStatus = "cancelled"
)

// Validate validates the key rotation record
func (k *KeyRotationRecord) Validate() error {
	if k.ValidatorAddress == "" {
		return fmt.Errorf("validator address cannot be empty")
	}

	if k.OldKeyFingerprint == "" {
		return fmt.Errorf("old key fingerprint cannot be empty")
	}

	if k.NewKeyFingerprint == "" {
		return fmt.Errorf("new key fingerprint cannot be empty")
	}

	if k.OverlapStartHeight >= k.OverlapEndHeight {
		return fmt.Errorf("overlap start height must be less than overlap end height")
	}

	return nil
}

// IsInOverlapPeriod checks if the current height is within the overlap period
func (k *KeyRotationRecord) IsInOverlapPeriod(currentHeight int64) bool {
	return currentHeight >= k.OverlapStartHeight && currentHeight < k.OverlapEndHeight
}
