package enclave_runtime

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	// WorkloadBindingDomainV1 is the fixed domain separator for the v1 verifier
	// contract/fixture boundary. It does not certify hardware, collateral, signer
	// ownership, replay consumption, or production readiness.
	WorkloadBindingDomainV1 = "virtengine.workload-binding"
	// WorkloadBindingVersionV1 is the fixed version for this verifier
	// contract/fixture boundary. It does not certify hardware, collateral, signer
	// ownership, replay consumption, or production readiness.
	WorkloadBindingVersionV1 = "1"

	maxWorkloadBindingIdentifierLength = 256
	maxMeasurementValueHexLength       = 1024
)

// WorkloadBindingCollateralReferenceV1 is an opaque verifier contract/fixture
// reference. It does not redefine cloud or storage contracts or certify hardware,
// collateral, signer ownership, replay consumption, or production readiness.
type WorkloadBindingCollateralReferenceV1 struct {
	Kind         string `json:"kind"`
	OpaqueID     string `json:"opaque_id"`
	SHA256Digest string `json:"sha256_digest"`
}

// WorkloadBindingV1 is a deterministic verifier contract and fixture boundary.
// It does not certify hardware, collateral, signer ownership, replay consumption,
// or production readiness.
type WorkloadBindingV1 struct {
	Domain                   string                               `json:"domain"`
	Version                  string                               `json:"version"`
	ChainID                  string                               `json:"chain_id"`
	WorkloadID               string                               `json:"workload_id"`
	Platform                 AttestationType                      `json:"platform"`
	MeasurementSHA256        string                               `json:"measurement_sha256"`
	MeasurementValueHex      string                               `json:"measurement_value_hex"`
	ReceiptSignerKeyID       string                               `json:"receipt_signer_key_id"`
	ReceiptSignerFingerprint string                               `json:"receipt_signer_sha256"`
	ModelDigest              string                               `json:"model_sha256"`
	RuntimeDigest            string                               `json:"runtime_sha256"`
	Nonce                    string                               `json:"nonce_hex"`
	ProfileID                string                               `json:"profile_id"`
	ProfileDigest            string                               `json:"profile_sha256"`
	ActivationHeight         int64                                `json:"activation_height"`
	ExpiryHeight             int64                                `json:"expiry_height"`
	Collateral               WorkloadBindingCollateralReferenceV1 `json:"collateral"`
	DebugMode                bool                                 `json:"debug_mode"`
}

// Validate fail-closed validates this verifier contract/fixture boundary and its
// chain-height applicability. It does not certify hardware, collateral, signer
// ownership, replay consumption, or production readiness. Activation is inclusive,
// expiry is exclusive, and no wall clock is consulted.
func (b WorkloadBindingV1) Validate(chainID string, height int64) error {
	if b.Domain != WorkloadBindingDomainV1 {
		return fmt.Errorf("invalid workload binding domain")
	}
	if b.Version != WorkloadBindingVersionV1 {
		return fmt.Errorf("invalid workload binding version")
	}
	identifiers := []struct {
		name  string
		value string
	}{
		{"chain ID", b.ChainID},
		{"workload ID", b.WorkloadID},
		{"receipt signer key ID", b.ReceiptSignerKeyID},
		{"profile ID", b.ProfileID},
		{"collateral kind", b.Collateral.Kind},
		{"collateral opaque ID", b.Collateral.OpaqueID},
	}
	for _, identifier := range identifiers {
		if err := validateWorkloadBindingIdentifier(identifier.value); err != nil {
			return fmt.Errorf("invalid %s: %w", identifier.name, err)
		}
	}
	if chainID != b.ChainID {
		return fmt.Errorf("chain ID mismatch")
	}
	if b.Platform != AttestationTypeSGX && b.Platform != AttestationTypeSEVSNP && b.Platform != AttestationTypeNitro {
		return fmt.Errorf("unsupported attestation platform %s", b.Platform)
	}
	commitments := []struct {
		name  string
		value string
	}{
		{"measurement digest", b.MeasurementSHA256},
		{"receipt signer fingerprint", b.ReceiptSignerFingerprint},
		{"model digest", b.ModelDigest},
		{"runtime digest", b.RuntimeDigest},
		{"nonce", b.Nonce},
		{"profile digest", b.ProfileDigest},
		{"collateral digest", b.Collateral.SHA256Digest},
	}
	for _, commitment := range commitments {
		if !isLowerSHA256Hex(commitment.value) {
			return fmt.Errorf("invalid %s", commitment.name)
		}
	}
	measurement, err := decodeMeasurementValue(b.MeasurementValueHex)
	if err != nil {
		return err
	}
	measurementDigest := sha256.Sum256(measurement)
	if hex.EncodeToString(measurementDigest[:]) != b.MeasurementSHA256 {
		return fmt.Errorf("measurement digest does not commit to measurement value")
	}
	if b.ActivationHeight <= 0 {
		return fmt.Errorf("activation height must be positive")
	}
	if b.ExpiryHeight <= b.ActivationHeight {
		return fmt.Errorf("expiry height must exceed activation height")
	}
	if height < b.ActivationHeight || height >= b.ExpiryHeight {
		return fmt.Errorf("height outside binding activation window")
	}
	if b.DebugMode {
		return fmt.Errorf("debug mode is forbidden")
	}
	return nil
}

// CanonicalBytes returns a fresh fixed-order encoding for this verifier
// contract/fixture boundary. It does not certify hardware, collateral, signer
// ownership, replay consumption, or production readiness.
func (b WorkloadBindingV1) CanonicalBytes() ([]byte, error) {
	if err := b.Validate(b.ChainID, b.ActivationHeight); err != nil {
		return nil, err
	}
	canonical, err := json.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("marshal workload binding: %w", err)
	}
	return append([]byte(nil), canonical...), nil
}

// ChallengeDigest returns a fresh SHA-256 verifier contract/fixture challenge.
// It does not certify hardware, collateral, signer ownership, replay consumption,
// or production readiness.
func (b WorkloadBindingV1) ChallengeDigest() ([]byte, error) {
	canonical, err := b.CanonicalBytes()
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(canonical)
	return append([]byte(nil), digest[:]...), nil
}

// ChallengeDigestHex returns the lowercase verifier contract/fixture challenge.
// It does not certify hardware, collateral, signer ownership, replay consumption,
// or production readiness.
func (b WorkloadBindingV1) ChallengeDigestHex() (string, error) {
	digest, err := b.ChallengeDigest()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(digest), nil
}

func validateWorkloadBindingIdentifier(value string) error {
	if value == "" || len(value) > maxWorkloadBindingIdentifierLength || strings.TrimSpace(value) != value {
		return fmt.Errorf("must be non-empty, trimmed, and at most %d bytes", maxWorkloadBindingIdentifierLength)
	}
	for _, character := range []byte(value) {
		if character < 0x20 || character > 0x7e {
			return fmt.Errorf("must contain only printable ASCII")
		}
	}
	return nil
}

func isLowerSHA256Hex(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	for _, character := range []byte(value) {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return false
		}
	}
	return true
}

func decodeMeasurementValue(value string) ([]byte, error) {
	if value == "" || len(value) > maxMeasurementValueHexLength || len(value)%2 != 0 {
		return nil, fmt.Errorf("invalid measurement value hex")
	}
	for _, character := range []byte(value) {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return nil, fmt.Errorf("invalid measurement value hex")
		}
	}
	decoded, err := hex.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("invalid measurement value hex: %w", err)
	}
	return decoded, nil
}
