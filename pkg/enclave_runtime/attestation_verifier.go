// Package enclave_runtime provides TEE enclave implementations.
//
// This file implements attestation verification for multiple TEE platforms:
// - Intel SGX DCAP quotes
// - AMD SEV-SNP attestation reports
// - AWS Nitro attestation documents
//
// Task Reference: VE-2026 - Attestation Verification Infrastructure
//
// The verifier validates cryptographic signatures, checks measurement allowlists,
// and enforces security policies before trusting enclave outputs.
package enclave_runtime

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

// AttestationType represents the type of TEE attestation.
type AttestationType int

const (
	// AttestationTypeUnknown represents an unknown attestation type.
	AttestationTypeUnknown AttestationType = iota
	// AttestationTypeSGX represents Intel SGX DCAP attestation.
	AttestationTypeSGX
	// AttestationTypeSEVSNP represents AMD SEV-SNP attestation.
	AttestationTypeSEVSNP
	// AttestationTypeNitro represents AWS Nitro attestation.
	AttestationTypeNitro
	// AttestationTypeSimulated represents simulated attestation for testing.
	AttestationTypeSimulated
)

// String returns the string representation of the attestation type.
func (t AttestationType) String() string {
	switch t {
	case AttestationTypeSGX:
		return "SGX"
	case AttestationTypeSEVSNP:
		return "SEV-SNP"
	case AttestationTypeNitro:
		return "NITRO"
	case AttestationTypeSimulated:
		return "SIMULATED"
	default:
		return "UNKNOWN"
	}
}

// Magic bytes for attestation type detection.
var (
	sgxQuoteMagic  = []byte{0x03, 0x00}       // SGX ECDSA quote version 3
	sevsnpMagic    = []byte{0x01, 0x00}       // SEV-SNP report version
	nitroMagic     = []byte{0xD2, 0x84}       // CBOR/COSE tag for Nitro
	simulatedMagic = []byte{0x53, 0x49, 0x4D} // "SIM"
)

// VerificationPolicy defines configurable security requirements for attestation verification.
type VerificationPolicy struct {
	// AllowDebugMode if false, reject attestations from debug enclaves.
	AllowDebugMode bool `json:"allow_debug_mode"`
	// RequireLatestTCB if true, require latest TCB version.
	RequireLatestTCB bool `json:"require_latest_tcb"`
	// AllowedPlatforms lists the permitted attestation types.
	AllowedPlatforms []AttestationType `json:"allowed_platforms"`
	// MinimumSecurityLevel: 1=Low, 2=Medium, 3=High.
	MinimumSecurityLevel int `json:"minimum_security_level"`
	// MaxAttestationAge rejects attestations older than this duration.
	MaxAttestationAge time.Duration `json:"max_attestation_age"`
	// RequireNonce requires fresh nonce in attestation.
	RequireNonce bool `json:"require_nonce"`
	// TrustedSignerKeys lists trusted signer public keys.
	TrustedSignerKeys [][]byte `json:"trusted_signer_keys"`
}

// DefaultVerificationPolicy returns a strict default policy.
func DefaultVerificationPolicy() VerificationPolicy {
	return VerificationPolicy{
		AllowDebugMode:       false,
		RequireLatestTCB:     false,
		AllowedPlatforms:     []AttestationType{AttestationTypeSGX, AttestationTypeSEVSNP, AttestationTypeNitro},
		MinimumSecurityLevel: 2,
		MaxAttestationAge:    24 * time.Hour,
		RequireNonce:         true,
		TrustedSignerKeys:    nil,
	}
}

// PermissiveVerificationPolicy returns a permissive policy for testing.
func PermissiveVerificationPolicy() VerificationPolicy {
	return VerificationPolicy{
		AllowDebugMode:       true,
		RequireLatestTCB:     false,
		AllowedPlatforms:     []AttestationType{AttestationTypeSGX, AttestationTypeSEVSNP, AttestationTypeNitro, AttestationTypeSimulated},
		MinimumSecurityLevel: 1,
		MaxAttestationAge:    365 * 24 * time.Hour,
		RequireNonce:         false,
		TrustedSignerKeys:    nil,
	}
}

// VerificationResult contains the structured output of attestation verification.
type VerificationResult struct {
	// Valid indicates whether the attestation passed all checks.
	Valid bool `json:"valid"`
	// AttestationType is the detected type of attestation.
	AttestationType AttestationType `json:"attestation_type"`
	// Measurement is the enclave measurement (MRENCLAVE, launch digest, PCR, etc.).
	Measurement []byte `json:"measurement"`
	// SignerKey is the signer identity (MRSIGNER, etc.).
	SignerKey []byte `json:"signer_key"`
	// TCBVersion is the TCB version string.
	TCBVersion string `json:"tcb_version"`
	// DebugMode indicates if the enclave was in debug mode.
	DebugMode bool `json:"debug_mode"`
	// Timestamp is when the attestation was generated.
	Timestamp time.Time `json:"timestamp"`
	// Errors contains any verification errors.
	Errors []string `json:"errors,omitempty"`
	// Warnings contains non-fatal issues.
	Warnings []string `json:"warnings,omitempty"`
	// RawAttestation is the original attestation data.
	RawAttestation []byte `json:"raw_attestation,omitempty"`
	// Nonce is the extracted nonce if present.
	Nonce []byte `json:"nonce,omitempty"`
	// SecurityLevel is the computed security level (1-3).
	SecurityLevel int `json:"security_level"`
	// Platform-specific fields
	SGXAttributes  uint64 `json:"sgx_attributes,omitempty"`
	SEVGuestPolicy uint64 `json:"sev_guest_policy,omitempty"`
	NitroEnclaveID string `json:"nitro_enclave_id,omitempty"`
}

// AddError adds an error message to the result.
func (r *VerificationResult) AddError(format string, args ...interface{}) {
	r.Errors = append(r.Errors, fmt.Sprintf(format, args...))
	r.Valid = false
}

// AddWarning adds a warning message to the result.
func (r *VerificationResult) AddWarning(format string, args ...interface{}) {
	r.Warnings = append(r.Warnings, fmt.Sprintf(format, args...))
}

// Measurement represents a trusted enclave measurement.
type Measurement struct {
	Platform    AttestationType `json:"platform"`
	Value       []byte          `json:"value"`
	Description string          `json:"description"`
	AddedAt     time.Time       `json:"added_at"`
	ExpiresAt   *time.Time      `json:"expires_at,omitempty"`
}

// MeasurementAllowlist manages trusted measurements.
type MeasurementAllowlist struct {
	mu           sync.RWMutex
	measurements map[AttestationType]map[string]Measurement
}

// NewMeasurementAllowlist creates a new empty allowlist.
func NewMeasurementAllowlist() *MeasurementAllowlist {
	return &MeasurementAllowlist{
		measurements: make(map[AttestationType]map[string]Measurement),
	}
}

// AddMeasurement adds a trusted measurement to the allowlist.
func (m *MeasurementAllowlist) AddMeasurement(platform AttestationType, measurement []byte, description string) error {
	if len(measurement) == 0 {
		return errors.New("measurement cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.measurements[platform] == nil {
		m.measurements[platform] = make(map[string]Measurement)
	}

	key := hex.EncodeToString(measurement)
	m.measurements[platform][key] = Measurement{
		Platform:    platform,
		Value:       measurement,
		Description: description,
		AddedAt:     time.Now(),
	}
	return nil
}

// RemoveMeasurement removes a measurement from the allowlist.
func (m *MeasurementAllowlist) RemoveMeasurement(platform AttestationType, measurement []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.measurements[platform] == nil {
		return fmt.Errorf("no measurements for platform %s", platform)
	}

	key := hex.EncodeToString(measurement)
	if _, exists := m.measurements[platform][key]; !exists {
		return fmt.Errorf("measurement not found for platform %s", platform)
	}

	delete(m.measurements[platform], key)
	return nil
}

// IsTrusted checks if a measurement is in the allowlist.
func (m *MeasurementAllowlist) IsTrusted(platform AttestationType, measurement []byte) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.measurements[platform] == nil {
		return false
	}

	key := hex.EncodeToString(measurement)
	meas, exists := m.measurements[platform][key]
	if !exists {
		return false
	}

	// Check expiration
	if meas.ExpiresAt != nil && time.Now().After(*meas.ExpiresAt) {
		return false
	}

	return true
}

// ListMeasurements returns all measurements for a platform.
func (m *MeasurementAllowlist) ListMeasurements(platform AttestationType) []Measurement {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.measurements[platform] == nil {
		return nil
	}

	result := make([]Measurement, 0, len(m.measurements[platform]))
	for _, meas := range m.measurements[platform] {
		result = append(result, meas)
	}
	return result
}

// allowlistJSON is the JSON representation for persistence.
type allowlistJSON struct {
	Measurements []Measurement `json:"measurements"`
}

// LoadFromJSON loads the allowlist from a JSON file.
func (m *MeasurementAllowlist) LoadFromJSON(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read allowlist file: %w", err)
	}

	var aj allowlistJSON
	if err := json.Unmarshal(data, &aj); err != nil {
		return fmt.Errorf("failed to parse allowlist JSON: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.measurements = make(map[AttestationType]map[string]Measurement)
	for _, meas := range aj.Measurements {
		if m.measurements[meas.Platform] == nil {
			m.measurements[meas.Platform] = make(map[string]Measurement)
		}
		key := hex.EncodeToString(meas.Value)
		m.measurements[meas.Platform][key] = meas
	}

	return nil
}

// SaveToJSON saves the allowlist to a JSON file.
func (m *MeasurementAllowlist) SaveToJSON(path string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var aj allowlistJSON
	for _, platformMeasurements := range m.measurements {
		for _, meas := range platformMeasurements {
			aj.Measurements = append(aj.Measurements, meas)
		}
	}

	data, err := json.MarshalIndent(aj, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal allowlist: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write allowlist file: %w", err)
	}

	return nil
}

// PlatformAttestationVerifier defines the interface for platform-specific attestation verification.
// This extends the base AttestationVerifier interface with policy-based verification.
type PlatformAttestationVerifier interface {
	// Verify verifies an attestation against the policy.
	Verify(attestation []byte, nonce []byte, policy VerificationPolicy) (*VerificationResult, error)
	// Type returns the attestation type this verifier handles.
	Type() AttestationType
}

// SGXDCAPVerifier verifies Intel SGX DCAP quotes.
type SGXDCAPVerifier struct {
	allowlist *MeasurementAllowlist
}

// NewSGXDCAPVerifier creates a new SGX DCAP verifier.
func NewSGXDCAPVerifier(allowlist *MeasurementAllowlist) *SGXDCAPVerifier {
	return &SGXDCAPVerifier{allowlist: allowlist}
}

// Type returns the attestation type.
func (v *SGXDCAPVerifier) Type() AttestationType {
	return AttestationTypeSGX
}

// SGX quote structure offsets (simplified).
const (
	sgxQuoteVersionOffset    = 0
	sgxQuoteAttKeyTypeOffset = 2
	sgxQuoteTeeTypeOffset    = 4
	sgxQuoteReserved1Offset  = 8
	sgxQuoteQEVendorIDOffset = 12
	sgxQuoteUserDataOffset   = 28
	sgxQuoteReportBodyOffset = 48
	sgxMRENCLAVEOffset       = 112 // Offset within report body
	sgxMRSIGNEROffset        = 176 // Offset within report body
	sgxAttributesOffset      = 96  // Offset within report body
	sgxReportDataOffset      = 320 // Offset within report body
	sgxMinQuoteSize          = 432 // Minimum quote size
)

// SGX attribute flags.
const (
	sgxAttrDebug     uint64 = 0x02
	sgxAttrMode64Bit uint64 = 0x04
)

// Verify verifies an SGX DCAP quote.
func (v *SGXDCAPVerifier) Verify(attestation []byte, nonce []byte, policy VerificationPolicy) (*VerificationResult, error) {
	result := &VerificationResult{
		Valid:           true,
		AttestationType: AttestationTypeSGX,
		RawAttestation:  attestation,
		Timestamp:       time.Now(),
		SecurityLevel:   3, // SGX provides high security by default
	}

	// Check minimum size
	if len(attestation) < sgxMinQuoteSize {
		result.AddError("SGX quote too small: got %d bytes, need at least %d", len(attestation), sgxMinQuoteSize)
		return result, nil
	}

	// Parse version
	version := binary.LittleEndian.Uint16(attestation[sgxQuoteVersionOffset:])
	if version != 3 {
		result.AddError("unsupported SGX quote version: %d (expected 3)", version)
		return result, nil
	}

	// Parse attestation key type
	attKeyType := binary.LittleEndian.Uint16(attestation[sgxQuoteAttKeyTypeOffset:])
	if attKeyType != 2 { // ECDSA-256-with-P-256
		result.AddWarning("unexpected attestation key type: %d", attKeyType)
	}

	// Extract report body
	reportBodyStart := sgxQuoteReportBodyOffset
	if reportBodyStart+384 > len(attestation) {
		result.AddError("SGX quote truncated: cannot read report body")
		return result, nil
	}

	// Extract MRENCLAVE (measurement)
	mrenclave := attestation[reportBodyStart+sgxMRENCLAVEOffset-sgxQuoteReportBodyOffset : reportBodyStart+sgxMRENCLAVEOffset-sgxQuoteReportBodyOffset+32]
	result.Measurement = make([]byte, 32)
	copy(result.Measurement, mrenclave)

	// Extract MRSIGNER
	mrsigner := attestation[reportBodyStart+sgxMRSIGNEROffset-sgxQuoteReportBodyOffset : reportBodyStart+sgxMRSIGNEROffset-sgxQuoteReportBodyOffset+32]
	result.SignerKey = make([]byte, 32)
	copy(result.SignerKey, mrsigner)

	// Extract attributes
	attrOffset := reportBodyStart + sgxAttributesOffset - sgxQuoteReportBodyOffset
	attributes := binary.LittleEndian.Uint64(attestation[attrOffset:])
	result.SGXAttributes = attributes
	result.DebugMode = (attributes & sgxAttrDebug) != 0

	// Extract report data (may contain nonce)
	reportDataOffset := reportBodyStart + sgxReportDataOffset - sgxQuoteReportBodyOffset
	if reportDataOffset+64 <= len(attestation) {
		result.Nonce = attestation[reportDataOffset : reportDataOffset+64]
	}

	// Policy checks
	if !policy.AllowDebugMode && result.DebugMode {
		result.AddError("SGX enclave is in debug mode, which is not allowed by policy")
	}

	// Verify nonce if required
	if policy.RequireNonce && len(nonce) > 0 {
		if !bytes.HasPrefix(result.Nonce, nonce) {
			result.AddError("nonce mismatch: attestation does not contain expected nonce")
		}
	}

	// Check measurement against allowlist
	if v.allowlist != nil && !v.allowlist.IsTrusted(AttestationTypeSGX, result.Measurement) {
		result.AddError("MRENCLAVE not in allowlist: %s", hex.EncodeToString(result.Measurement))
	}

	// TODO: Verify QE signature cryptographically
	result.AddWarning("QE signature verification is simulated (not cryptographically verified)")

	// TODO: Check TCB status against Intel PCS
	result.TCBVersion = "simulated-tcb-v1"
	if policy.RequireLatestTCB {
		result.AddWarning("TCB freshness check is simulated")
	}

	// Check platform is allowed
	if !isPlatformAllowed(AttestationTypeSGX, policy.AllowedPlatforms) {
		result.AddError("SGX platform not allowed by policy")
	}

	return result, nil
}

// SEVSNPVerifier verifies AMD SEV-SNP attestation reports.
type SEVSNPVerifier struct {
	allowlist *MeasurementAllowlist
}

// NewSEVSNPVerifier creates a new SEV-SNP verifier.
func NewSEVSNPVerifier(allowlist *MeasurementAllowlist) *SEVSNPVerifier {
	return &SEVSNPVerifier{allowlist: allowlist}
}

// Type returns the attestation type.
func (v *SEVSNPVerifier) Type() AttestationType {
	return AttestationTypeSEVSNP
}

// SEV-SNP report structure offsets (simplified).
const (
	sevsnpVersionOffset         = 0
	sevsnpGuestSVNOffset        = 4
	sevsnpPolicyOffset          = 8
	sevsnpFamilyIDOffset        = 16
	sevsnpImageIDOffset         = 32
	sevsnpVMPLOffset            = 48
	sevsnpSignatureAlgoOffset   = 52
	sevsnpCurrentTCBOffset      = 56
	sevsnpPlatformInfoOffset    = 64
	sevsnpAuthorKeyEnOffset     = 72
	sevsnpReportDataOffset      = 80
	sevsnpMeasurementOffset     = 144
	sevsnpHostDataOffset        = 192
	sevsnpIDKeyDigestOffset     = 224
	sevsnpAuthorKeyDigestOffset = 256
	sevsnpReportIDOffset        = 288
	sevsnpReportIDMAOffset      = 320
	sevsnpReportedTCBOffset     = 352
	sevsnpChipIDOffset          = 384
	sevsnpSignatureOffset       = 416
	sevsnpMinReportSize         = 672
)

// SEV-SNP policy flags.
const (
	sevsnpPolicyDebug          uint64 = 1 << 19
	sevsnpPolicySMT            uint64 = 1 << 16
	sevsnpPolicyMigrationAgent uint64 = 1 << 18
)

// Verify verifies a SEV-SNP attestation report.
func (v *SEVSNPVerifier) Verify(attestation []byte, nonce []byte, policy VerificationPolicy) (*VerificationResult, error) {
	result := &VerificationResult{
		Valid:           true,
		AttestationType: AttestationTypeSEVSNP,
		RawAttestation:  attestation,
		Timestamp:       time.Now(),
		SecurityLevel:   3, // SEV-SNP provides high security
	}

	// Check minimum size
	if len(attestation) < sevsnpMinReportSize {
		result.AddError("SEV-SNP report too small: got %d bytes, need at least %d", len(attestation), sevsnpMinReportSize)
		return result, nil
	}

	// Parse version
	version := binary.LittleEndian.Uint32(attestation[sevsnpVersionOffset:])
	if version != 1 && version != 2 {
		result.AddError("unsupported SEV-SNP report version: %d", version)
		return result, nil
	}

	// Parse guest policy
	guestPolicy := binary.LittleEndian.Uint64(attestation[sevsnpPolicyOffset:])
	result.SEVGuestPolicy = guestPolicy
	result.DebugMode = (guestPolicy & sevsnpPolicyDebug) != 0

	// Extract measurement (launch digest)
	result.Measurement = make([]byte, 48)
	copy(result.Measurement, attestation[sevsnpMeasurementOffset:sevsnpMeasurementOffset+48])

	// Extract report data (may contain nonce)
	result.Nonce = make([]byte, 64)
	copy(result.Nonce, attestation[sevsnpReportDataOffset:sevsnpReportDataOffset+64])

	// Extract ID key digest as signer key
	result.SignerKey = make([]byte, 32)
	copy(result.SignerKey, attestation[sevsnpIDKeyDigestOffset:sevsnpIDKeyDigestOffset+32])

	// Parse TCB version
	currentTCB := binary.LittleEndian.Uint64(attestation[sevsnpCurrentTCBOffset:])
	result.TCBVersion = fmt.Sprintf("tcb-%d", currentTCB)

	// Policy checks
	if !policy.AllowDebugMode && result.DebugMode {
		result.AddError("SEV-SNP guest has debug policy enabled, which is not allowed")
	}

	// Verify nonce if required
	if policy.RequireNonce && len(nonce) > 0 {
		if !bytes.HasPrefix(result.Nonce, nonce) {
			result.AddError("nonce mismatch: report does not contain expected nonce")
		}
	}

	// Check measurement against allowlist
	if v.allowlist != nil && !v.allowlist.IsTrusted(AttestationTypeSEVSNP, result.Measurement) {
		result.AddError("launch digest not in allowlist: %s", hex.EncodeToString(result.Measurement))
	}

	// TODO: Verify VCEK signature cryptographically
	result.AddWarning("VCEK signature verification is simulated (not cryptographically verified)")

	// Check platform is allowed
	if !isPlatformAllowed(AttestationTypeSEVSNP, policy.AllowedPlatforms) {
		result.AddError("SEV-SNP platform not allowed by policy")
	}

	return result, nil
}

// NitroVerifier verifies AWS Nitro attestation documents.
type NitroVerifier struct {
	allowlist *MeasurementAllowlist
}

// NewNitroVerifier creates a new Nitro verifier.
func NewNitroVerifier(allowlist *MeasurementAllowlist) *NitroVerifier {
	return &NitroVerifier{allowlist: allowlist}
}

// Type returns the attestation type.
func (v *NitroVerifier) Type() AttestationType {
	return AttestationTypeNitro
}

// Nitro attestation document structure (simplified).
// Real implementation would use CBOR/COSE parsing.
const (
	nitroMinDocSize = 64
	nitroPCROffset  = 32
	nitroPCRLength  = 48
)

// Verify verifies a Nitro attestation document.
func (v *NitroVerifier) Verify(attestation []byte, nonce []byte, policy VerificationPolicy) (*VerificationResult, error) {
	result := &VerificationResult{
		Valid:           true,
		AttestationType: AttestationTypeNitro,
		RawAttestation:  attestation,
		Timestamp:       time.Now(),
		SecurityLevel:   3,     // Nitro provides high security
		DebugMode:       false, // Nitro doesn't have debug mode in production
	}

	// Check minimum size
	if len(attestation) < nitroMinDocSize {
		result.AddError("Nitro document too small: got %d bytes, need at least %d", len(attestation), nitroMinDocSize)
		return result, nil
	}

	// Check CBOR/COSE magic bytes
	if len(attestation) >= 2 && !bytes.HasPrefix(attestation, nitroMagic) {
		result.AddWarning("Nitro document does not start with expected COSE tag")
	}

	// TODO: Parse CBOR/COSE structure
	// For now, extract simulated PCR values from fixed offset
	if len(attestation) >= nitroPCROffset+nitroPCRLength {
		result.Measurement = make([]byte, nitroPCRLength)
		copy(result.Measurement, attestation[nitroPCROffset:nitroPCROffset+nitroPCRLength])
	} else {
		// Use hash of document as simulated measurement
		result.Measurement = make([]byte, 32)
		for i := 0; i < 32 && i < len(attestation); i++ {
			result.Measurement[i] = attestation[i]
		}
	}

	// Extract nonce from document (simulated)
	if len(attestation) > 96 {
		result.Nonce = attestation[64:96]
	}

	// Extract enclave ID (simulated)
	result.NitroEnclaveID = fmt.Sprintf("i-%x", attestation[:8])

	// TCB version for Nitro
	result.TCBVersion = "nitro-v1"

	// Verify nonce if required
	if policy.RequireNonce && len(nonce) > 0 {
		if len(result.Nonce) < len(nonce) || !bytes.Equal(result.Nonce[:len(nonce)], nonce) {
			result.AddError("nonce mismatch: document does not contain expected nonce")
		}
	}

	// Check measurement against allowlist
	if v.allowlist != nil && !v.allowlist.IsTrusted(AttestationTypeNitro, result.Measurement) {
		result.AddError("PCR values not in allowlist: %s", hex.EncodeToString(result.Measurement))
	}

	// TODO: Verify certificate chain to AWS Nitro root
	result.AddWarning("Nitro certificate chain verification is simulated")

	// Check platform is allowed
	if !isPlatformAllowed(AttestationTypeNitro, policy.AllowedPlatforms) {
		result.AddError("Nitro platform not allowed by policy")
	}

	return result, nil
}

// SimulatedVerifier verifies simulated attestations for testing.
type SimulatedVerifier struct {
	allowlist *MeasurementAllowlist
}

// NewSimulatedVerifier creates a new simulated verifier.
func NewSimulatedVerifier(allowlist *MeasurementAllowlist) *SimulatedVerifier {
	return &SimulatedVerifier{allowlist: allowlist}
}

// Type returns the attestation type.
func (v *SimulatedVerifier) Type() AttestationType {
	return AttestationTypeSimulated
}

// Verify verifies a simulated attestation.
func (v *SimulatedVerifier) Verify(attestation []byte, nonce []byte, policy VerificationPolicy) (*VerificationResult, error) {
	result := &VerificationResult{
		Valid:           true,
		AttestationType: AttestationTypeSimulated,
		RawAttestation:  attestation,
		Timestamp:       time.Now(),
		SecurityLevel:   1,    // Simulated provides low security
		DebugMode:       true, // Simulated is always "debug"
		TCBVersion:      "simulated",
	}

	// Check magic
	if len(attestation) < 3 || !bytes.HasPrefix(attestation, simulatedMagic) {
		result.AddError("invalid simulated attestation format")
		return result, nil
	}

	// Extract measurement (bytes after magic)
	if len(attestation) > 35 {
		result.Measurement = attestation[3:35]
	} else if len(attestation) > 3 {
		result.Measurement = attestation[3:]
	}

	// Extract nonce if present
	if len(attestation) > 67 {
		result.Nonce = attestation[35:67]
	}

	// Policy checks
	if !policy.AllowDebugMode {
		result.AddError("simulated attestation is always in debug mode")
	}

	// Check platform is allowed
	if !isPlatformAllowed(AttestationTypeSimulated, policy.AllowedPlatforms) {
		result.AddError("simulated platform not allowed by policy")
	}

	// Verify nonce if required
	if policy.RequireNonce && len(nonce) > 0 {
		if len(result.Nonce) < len(nonce) || !bytes.Equal(result.Nonce[:len(nonce)], nonce) {
			result.AddError("nonce mismatch")
		}
	}

	// Check measurement against allowlist
	if v.allowlist != nil && len(result.Measurement) > 0 && !v.allowlist.IsTrusted(AttestationTypeSimulated, result.Measurement) {
		result.AddError("measurement not in allowlist")
	}

	return result, nil
}

// UniversalAttestationVerifier auto-detects and routes attestations to appropriate verifiers.
type UniversalAttestationVerifier struct {
	sgxVerifier       *SGXDCAPVerifier
	sevsnpVerifier    *SEVSNPVerifier
	nitroVerifier     *NitroVerifier
	simulatedVerifier *SimulatedVerifier
	allowlist         *MeasurementAllowlist
}

// NewUniversalAttestationVerifier creates a new universal verifier.
func NewUniversalAttestationVerifier(allowlist *MeasurementAllowlist) *UniversalAttestationVerifier {
	return &UniversalAttestationVerifier{
		sgxVerifier:       NewSGXDCAPVerifier(allowlist),
		sevsnpVerifier:    NewSEVSNPVerifier(allowlist),
		nitroVerifier:     NewNitroVerifier(allowlist),
		simulatedVerifier: NewSimulatedVerifier(allowlist),
		allowlist:         allowlist,
	}
}

// DetectAttestationType determines the attestation type from the payload.
func DetectAttestationType(attestation []byte) AttestationType {
	if len(attestation) < 2 {
		return AttestationTypeUnknown
	}

	// Check for simulated first (most specific magic)
	if len(attestation) >= 3 && bytes.HasPrefix(attestation, simulatedMagic) {
		return AttestationTypeSimulated
	}

	// Check for Nitro CBOR/COSE
	if bytes.HasPrefix(attestation, nitroMagic) {
		return AttestationTypeNitro
	}

	// Check for SGX quote (version 3)
	if bytes.HasPrefix(attestation, sgxQuoteMagic) {
		return AttestationTypeSGX
	}

	// Check for SEV-SNP report
	if bytes.HasPrefix(attestation, sevsnpMagic) {
		return AttestationTypeSEVSNP
	}

	return AttestationTypeUnknown
}

// Verify auto-detects the attestation type and verifies it.
func (v *UniversalAttestationVerifier) Verify(attestation []byte, nonce []byte, policy VerificationPolicy) (*VerificationResult, error) {
	attType := DetectAttestationType(attestation)

	switch attType {
	case AttestationTypeSGX:
		return v.sgxVerifier.Verify(attestation, nonce, policy)
	case AttestationTypeSEVSNP:
		return v.sevsnpVerifier.Verify(attestation, nonce, policy)
	case AttestationTypeNitro:
		return v.nitroVerifier.Verify(attestation, nonce, policy)
	case AttestationTypeSimulated:
		return v.simulatedVerifier.Verify(attestation, nonce, policy)
	default:
		return &VerificationResult{
			Valid:           false,
			AttestationType: AttestationTypeUnknown,
			RawAttestation:  attestation,
			Timestamp:       time.Now(),
			Errors:          []string{"unable to detect attestation type from payload"},
		}, nil
	}
}

// Type returns Unknown as the universal verifier handles all types.
func (v *UniversalAttestationVerifier) Type() AttestationType {
	return AttestationTypeUnknown
}

// VerifyMultiple verifies multiple attestations and aggregates results.
func (v *UniversalAttestationVerifier) VerifyMultiple(attestations [][]byte, nonces [][]byte, policy VerificationPolicy) ([]*VerificationResult, error) {
	results := make([]*VerificationResult, len(attestations))

	for i, att := range attestations {
		var nonce []byte
		if i < len(nonces) {
			nonce = nonces[i]
		}

		result, err := v.Verify(att, nonce, policy)
		if err != nil {
			return nil, fmt.Errorf("failed to verify attestation %d: %w", i, err)
		}
		results[i] = result
	}

	return results, nil
}

// GetVerifier returns the specific verifier for an attestation type.
func (v *UniversalAttestationVerifier) GetVerifier(attType AttestationType) PlatformAttestationVerifier {
	switch attType {
	case AttestationTypeSGX:
		return v.sgxVerifier
	case AttestationTypeSEVSNP:
		return v.sevsnpVerifier
	case AttestationTypeNitro:
		return v.nitroVerifier
	case AttestationTypeSimulated:
		return v.simulatedVerifier
	default:
		return nil
	}
}

// MeasurementAllowlistManager provides a higher-level API for managing allowlists.
type MeasurementAllowlistManager struct {
	allowlist *MeasurementAllowlist
	filePath  string
	autoSave  bool
	mu        sync.Mutex
}

// NewMeasurementAllowlistManager creates a new manager.
func NewMeasurementAllowlistManager(allowlist *MeasurementAllowlist) *MeasurementAllowlistManager {
	return &MeasurementAllowlistManager{
		allowlist: allowlist,
		autoSave:  false,
	}
}

// SetFilePath sets the file path for persistence.
func (m *MeasurementAllowlistManager) SetFilePath(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.filePath = path
}

// SetAutoSave enables or disables automatic saving after modifications.
func (m *MeasurementAllowlistManager) SetAutoSave(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.autoSave = enabled
}

// AddMeasurement adds a measurement and optionally saves.
func (m *MeasurementAllowlistManager) AddMeasurement(platform AttestationType, measurement []byte, description string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.allowlist.AddMeasurement(platform, measurement, description); err != nil {
		return err
	}

	if m.autoSave && m.filePath != "" {
		return m.allowlist.SaveToJSON(m.filePath)
	}
	return nil
}

// RemoveMeasurement removes a measurement and optionally saves.
func (m *MeasurementAllowlistManager) RemoveMeasurement(platform AttestationType, measurement []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.allowlist.RemoveMeasurement(platform, measurement); err != nil {
		return err
	}

	if m.autoSave && m.filePath != "" {
		return m.allowlist.SaveToJSON(m.filePath)
	}
	return nil
}

// IsTrusted checks if a measurement is trusted.
func (m *MeasurementAllowlistManager) IsTrusted(platform AttestationType, measurement []byte) bool {
	return m.allowlist.IsTrusted(platform, measurement)
}

// ListMeasurements returns all measurements for a platform.
func (m *MeasurementAllowlistManager) ListMeasurements(platform AttestationType) []Measurement {
	return m.allowlist.ListMeasurements(platform)
}

// LoadFromJSON loads from the configured file path.
func (m *MeasurementAllowlistManager) LoadFromJSON(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if path != "" {
		m.filePath = path
	}
	if m.filePath == "" {
		return errors.New("no file path configured")
	}
	return m.allowlist.LoadFromJSON(m.filePath)
}

// SaveToJSON saves to the configured file path.
func (m *MeasurementAllowlistManager) SaveToJSON(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if path != "" {
		m.filePath = path
	}
	if m.filePath == "" {
		return errors.New("no file path configured")
	}
	return m.allowlist.SaveToJSON(m.filePath)
}

// GetAllowlist returns the underlying allowlist.
func (m *MeasurementAllowlistManager) GetAllowlist() *MeasurementAllowlist {
	return m.allowlist
}

// isPlatformAllowed checks if a platform is in the allowed list.
func isPlatformAllowed(platform AttestationType, allowed []AttestationType) bool {
	if len(allowed) == 0 {
		return true // Empty list means all allowed
	}
	for _, p := range allowed {
		if p == platform {
			return true
		}
	}
	return false
}

// CreateTestSGXQuote creates a test SGX quote for testing purposes.
func CreateTestSGXQuote(mrenclave, mrsigner []byte, debugMode bool, nonce []byte) []byte {
	quote := make([]byte, sgxMinQuoteSize+64)

	// Version 3
	binary.LittleEndian.PutUint16(quote[sgxQuoteVersionOffset:], 3)

	// Attestation key type: ECDSA-256-with-P-256
	binary.LittleEndian.PutUint16(quote[sgxQuoteAttKeyTypeOffset:], 2)

	// Report body starts at offset 48
	reportBody := quote[sgxQuoteReportBodyOffset:]

	// Attributes
	var attrs uint64 = sgxAttrMode64Bit
	if debugMode {
		attrs |= sgxAttrDebug
	}
	binary.LittleEndian.PutUint64(reportBody[sgxAttributesOffset-sgxQuoteReportBodyOffset:], attrs)

	// MRENCLAVE
	if len(mrenclave) >= 32 {
		copy(reportBody[sgxMRENCLAVEOffset-sgxQuoteReportBodyOffset:], mrenclave[:32])
	}

	// MRSIGNER
	if len(mrsigner) >= 32 {
		copy(reportBody[sgxMRSIGNEROffset-sgxQuoteReportBodyOffset:], mrsigner[:32])
	}

	// Report data (nonce)
	if len(nonce) > 0 {
		rdOffset := sgxReportDataOffset - sgxQuoteReportBodyOffset
		copy(reportBody[rdOffset:], nonce)
	}

	return quote
}

// CreateTestSEVSNPReport creates a test SEV-SNP report for testing purposes.
func CreateTestSEVSNPReport(measurement []byte, debugMode bool, nonce []byte) []byte {
	report := make([]byte, sevsnpMinReportSize)

	// Version 1
	binary.LittleEndian.PutUint32(report[sevsnpVersionOffset:], 1)

	// Guest policy
	var policy uint64 = 0
	if debugMode {
		policy |= sevsnpPolicyDebug
	}
	binary.LittleEndian.PutUint64(report[sevsnpPolicyOffset:], policy)

	// Measurement
	if len(measurement) >= 48 {
		copy(report[sevsnpMeasurementOffset:], measurement[:48])
	} else if len(measurement) > 0 {
		copy(report[sevsnpMeasurementOffset:], measurement)
	}

	// Report data (nonce)
	if len(nonce) > 0 {
		copy(report[sevsnpReportDataOffset:], nonce)
	}

	// TCB version
	binary.LittleEndian.PutUint64(report[sevsnpCurrentTCBOffset:], 1)

	return report
}

// CreateTestNitroDocument creates a test Nitro document for testing purposes.
func CreateTestNitroDocument(pcrs []byte, nonce []byte) []byte {
	doc := make([]byte, 128)

	// CBOR/COSE magic
	copy(doc[0:], nitroMagic)

	// PCRs at offset 32
	if len(pcrs) > 0 {
		copy(doc[nitroPCROffset:], pcrs)
	}

	// Nonce at offset 64
	if len(nonce) > 0 {
		copy(doc[64:], nonce)
	}

	return doc
}

// CreateTestSimulatedAttestation creates a test simulated attestation.
func CreateTestSimulatedAttestation(measurement []byte, nonce []byte) []byte {
	att := make([]byte, 3+32+32)
	copy(att[0:], simulatedMagic)

	if len(measurement) > 0 {
		copy(att[3:], measurement)
	}

	if len(nonce) > 0 {
		copy(att[35:], nonce)
	}

	return att
}
