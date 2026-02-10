package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	v1 "github.com/virtengine/virtengine/sdk/go/node/enclave/v1"
)

// AllTEETypes returns all valid TEE types
func AllTEETypes() []v1.TEEType {
	return []v1.TEEType{v1.TEETypeSGX, v1.TEETypeSEVSNP, v1.TEETypeNitro, v1.TEETypeTrustZone}
}

// IsValidTEEType checks if a TEE type is valid
func IsValidTEEType(teeType v1.TEEType) bool {
	switch teeType {
	case v1.TEETypeSGX, v1.TEETypeSEVSNP, v1.TEETypeNitro, v1.TEETypeTrustZone:
		return true
	}
	return false
}

// ValidateEnclaveIdentity validates the enclave identity
func ValidateEnclaveIdentity(e *v1.EnclaveIdentity) error {
	if e.ValidatorAddress == "" {
		return ErrInvalidEnclaveIdentity.Wrap("validator address cannot be empty")
	}

	if !IsValidTEEType(e.TeeType) {
		return ErrInvalidEnclaveIdentity.Wrapf("invalid TEE type: %s", e.TeeType)
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
func KeyFingerprint(pubKey []byte) string {
	h := sha256.Sum256(pubKey)
	return hex.EncodeToString(h[:])
}

// IsIdentityExpired checks if the enclave identity has expired at the given block height
func IsIdentityExpired(e *v1.EnclaveIdentity, currentHeight int64) bool {
	return currentHeight >= e.ExpiryHeight
}

// ValidateMeasurementRecord validates the measurement record
func ValidateMeasurementRecord(m *v1.MeasurementRecord) error {
	if len(m.MeasurementHash) == 0 {
		return ErrInvalidMeasurement.Wrap("measurement hash cannot be empty")
	}

	if len(m.MeasurementHash) != 32 {
		return ErrInvalidMeasurement.Wrapf("measurement hash must be 32 bytes, got %d", len(m.MeasurementHash))
	}

	if !IsValidTEEType(m.TeeType) {
		return ErrInvalidMeasurement.Wrapf("invalid TEE type: %s", m.TeeType)
	}

	if m.Description == "" {
		return ErrInvalidMeasurement.Wrap("description cannot be empty")
	}

	return nil
}

// IsMeasurementValid checks if the measurement is valid (not revoked and not expired)
func IsMeasurementValid(m *v1.MeasurementRecord, currentHeight int64) bool {
	if m.Revoked {
		return false
	}
	if m.ExpiryHeight > 0 && currentHeight >= m.ExpiryHeight {
		return false
	}
	return true
}

// ValidateKeyRotationRecord validates the key rotation record
func ValidateKeyRotationRecord(k *v1.KeyRotationRecord) error {
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
func IsInOverlapPeriod(k *v1.KeyRotationRecord, currentHeight int64) bool {
	return currentHeight >= k.OverlapStartHeight && currentHeight < k.OverlapEndHeight
}

// MeasurementHashHex returns a measurement hash as hex string
func MeasurementHashHex(hash []byte) string {
	return hex.EncodeToString(hash)
}
