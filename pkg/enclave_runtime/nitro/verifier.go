// Package nitro provides AWS Nitro Enclave integration for VirtEngine TEE.
//
// This file implements attestation document verification for AWS Nitro Enclaves.
// Verification includes:
// - COSE_Sign1 signature verification
// - Certificate chain validation against AWS Nitro Root CA
// - PCR value validation
// - Document freshness checks
//
// Task Reference: VE-2029 - Hardware TEE Integration Layer
package nitro

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha512"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// =============================================================================
// Verification Constants
// =============================================================================

const (
	// DefaultMaxDocumentAge is the default maximum age for attestation documents
	DefaultMaxDocumentAge = 5 * time.Minute

	// DefaultMaxClockSkew is the default allowed clock skew
	DefaultMaxClockSkew = 30 * time.Second

	// AWS Nitro Root CA Subject
	AWSNitroRootCASubject = "CN=aws.nitro-enclaves"

	// Signature size for ES384 (96 bytes: 48 bytes R + 48 bytes S)
	ES384SignatureSize = 96
)

// AWS Nitro Enclave Attestation Root CA (PEM format)
// This is the root of trust for all Nitro attestation documents
const AWSNitroRootCAPEM = `-----BEGIN CERTIFICATE-----
MIICETCCAZagAwIBAgIRAPkxdWgbkK/hHUbMtOTn+FYwCgYIKoZIzj0EAwMwSTEL
MAkGA1UEBhMCVVMxDzANBgNVBAoMBkFtYXpvbjEMMAoGA1UECwwDQVdTMRswGQYD
VQQDDBJhd3Mubml0cm8tZW5jbGF2ZXMwHhcNMTkxMDI4MTMyODA1WhcNNDkxMDI4
MTQyODA1WjBJMQswCQYDVQQGEwJVUzEPMA0GA1UECgwGQW1hem9uMQwwCgYDVQQL
DANBV1MxGzAZBgNVBAMMEmF3cy5uaXRyby1lbmNsYXZlczB2MBAGByqGSM49AgEG
BSuBBAAiA2IABPwCVOumCMHzaHDimtqQvkY4MpJzbolL//Zy2YlES1BR5TSksfbb
48C8WBoyt7F2Bw7eEtaaP+ohG2bnUs990d0JX28TcPQXCEPZ3BABIeTPYwEoCWZE
h8l5YoQwTcU/9KNCMEAwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUkCW1DdkF
R+eWw5b6cp3PmanfS5YwDgYDVR0PAQH/BAQDAgGGMAoGCCqGSM49BAMDA2kAMGYC
MQCjfy+Rocm9Xue4YnwWmNJVA44fA0P5W2OpYow9OYCVRaEevL8uO1XYru5xtMPW
rfMCMQCi85sWBbJwKKXdS6BptQFuZbT73o/gBh1qUxl/nNr12UO8Yfwr6wPLb+6N
IwLz3/Y=
-----END CERTIFICATE-----`

// =============================================================================
// Verification Errors
// =============================================================================

var (
	// ErrVerificationFailed is returned when verification fails
	ErrVerificationFailed = errors.New("verification failed")

	// ErrSignatureInvalid is returned when signature verification fails
	ErrSignatureInvalid = errors.New("invalid signature")

	// ErrCertificateChainInvalid is returned when certificate chain is invalid
	ErrCertificateChainInvalid = errors.New("invalid certificate chain")

	// ErrRootCAMismatch is returned when root CA doesn't match AWS Nitro
	ErrRootCAMismatch = errors.New("root CA mismatch")

	// ErrPCRMismatch is returned when PCR values don't match expected
	ErrPCRMismatch = errors.New("PCR value mismatch")

	// ErrDocumentTooOld is returned when document timestamp is too old
	ErrDocumentTooOld = errors.New("document too old")

	// ErrDocumentFromFuture is returned when document is from the future
	ErrDocumentFromFuture = errors.New("document from the future")

	// ErrNoRootCA is returned when root CA is not configured
	ErrNoRootCA = errors.New("root CA not configured")
)

// =============================================================================
// Verifier Configuration
// =============================================================================

// VerifierConfig configures the attestation verifier
type VerifierConfig struct {
	// RootCA is the AWS Nitro Root CA certificate
	// If nil, uses the embedded AWS Nitro Root CA
	RootCA *x509.Certificate

	// AllowedPCRs maps PCR indices to expected values
	// Only specified PCRs are validated; others are ignored
	AllowedPCRs map[int][]byte

	// MaxDocumentAge is the maximum age for attestation documents
	// Default: 5 minutes
	MaxDocumentAge time.Duration

	// MaxClockSkew is the maximum allowed clock skew
	// Default: 30 seconds
	MaxClockSkew time.Duration

	// RequireUserData requires user data to be present
	RequireUserData bool

	// ExpectedUserData is the expected user data (if RequireUserData is true)
	ExpectedUserData []byte

	// RequireNonce requires nonce to be present
	RequireNonce bool

	// ExpectedNonce is the expected nonce (if RequireNonce is true)
	ExpectedNonce []byte

	// SkipSignatureVerification skips signature verification (for testing only)
	SkipSignatureVerification bool

	// SkipCertificateChainVerification skips cert chain verification (for testing)
	SkipCertificateChainVerification bool

	// AllowSimulated allows simulated attestation documents
	AllowSimulated bool
}

// DefaultVerifierConfig returns a config with default settings
func DefaultVerifierConfig() *VerifierConfig {
	rootCA, _ := parseAWSNitroRootCA()
	return &VerifierConfig{
		RootCA:         rootCA,
		AllowedPCRs:    make(map[int][]byte),
		MaxDocumentAge: DefaultMaxDocumentAge,
		MaxClockSkew:   DefaultMaxClockSkew,
	}
}

// =============================================================================
// Verification Result
// =============================================================================

// VerificationResult contains the result of attestation verification
type VerificationResult struct {
	// Valid is true if all verification checks passed
	Valid bool

	// Document is the parsed attestation document
	Document *AttestationDocument

	// Certificate is the parsed attestation certificate
	Certificate *x509.Certificate

	// CABundle is the parsed certificate chain
	CABundle []*x509.Certificate

	// PCRDigest is the combined digest of core PCRs
	PCRDigest []byte

	// Timestamp is the document timestamp
	Timestamp time.Time

	// UserData is the extracted user data
	UserData []byte

	// PublicKey is the extracted public key
	PublicKey []byte

	// ModuleID is the enclave module ID
	ModuleID string

	// Warnings contains non-fatal issues detected
	Warnings []string

	// Error contains the error if verification failed
	Error error
}

// =============================================================================
// Verifier
// =============================================================================

// Verifier verifies AWS Nitro attestation documents
type Verifier struct {
	mu     sync.RWMutex
	config *VerifierConfig
}

// NewVerifier creates a new attestation verifier with default config
func NewVerifier() *Verifier {
	return &Verifier{
		config: DefaultVerifierConfig(),
	}
}

// NewVerifierWithConfig creates a verifier with custom config
func NewVerifierWithConfig(config *VerifierConfig) *Verifier {
	if config == nil {
		config = DefaultVerifierConfig()
	}
	if config.MaxDocumentAge == 0 {
		config.MaxDocumentAge = DefaultMaxDocumentAge
	}
	if config.MaxClockSkew == 0 {
		config.MaxClockSkew = DefaultMaxClockSkew
	}
	if config.RootCA == nil {
		config.RootCA, _ = parseAWSNitroRootCA()
	}
	return &Verifier{
		config: config,
	}
}

// SetAllowedPCRs sets the expected PCR values for verification
func (v *Verifier) SetAllowedPCRs(pcrs map[int][]byte) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.config.AllowedPCRs = pcrs
}

// SetAllowedPCRsFromHex sets PCRs from hex strings
func (v *Verifier) SetAllowedPCRsFromHex(pcrs map[int]string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	decoded := make(map[int][]byte, len(pcrs))
	for idx, hexStr := range pcrs {
		val, err := hex.DecodeString(hexStr)
		if err != nil {
			return fmt.Errorf("PCR%d: invalid hex: %w", idx, err)
		}
		if len(val) != PCRDigestSize {
			return fmt.Errorf("PCR%d: invalid size %d", idx, len(val))
		}
		decoded[idx] = val
	}
	v.config.AllowedPCRs = decoded
	return nil
}

// =============================================================================
// Main Verification Entry Point
// =============================================================================

// VerifyDocument verifies an attestation document
func (v *Verifier) VerifyDocument(doc *AttestationDocument) (*VerificationResult, error) {
	v.mu.RLock()
	config := v.config
	v.mu.RUnlock()

	result := &VerificationResult{
		Document: doc,
		Warnings: make([]string, 0),
	}

	// Step 1: Validate document structure
	if err := ValidateDocument(doc); err != nil {
		result.Error = fmt.Errorf("document validation failed: %w", err)
		return result, result.Error
	}

	// Extract basic info
	result.ModuleID = doc.Payload.ModuleID
	result.Timestamp = time.UnixMilli(int64(doc.Payload.Timestamp))
	result.UserData = doc.Payload.UserData
	result.PublicKey = doc.Payload.PublicKey
	result.PCRDigest = GetPCRDigest(doc.Payload.PCRs)

	// Check for simulated document
	if isSimulatedDocument(doc) {
		if !config.AllowSimulated {
			result.Error = errors.New("simulated attestation not allowed")
			return result, result.Error
		}
		result.Warnings = append(result.Warnings, "simulated attestation document")
		result.Valid = true
		return result, nil
	}

	// Step 2: Parse certificates
	cert, err := GetCertificate(doc)
	if err != nil {
		result.Error = fmt.Errorf("certificate parsing failed: %w", err)
		return result, result.Error
	}
	result.Certificate = cert

	caBundle, err := GetCABundle(doc)
	if err != nil {
		result.Error = fmt.Errorf("CA bundle parsing failed: %w", err)
		return result, result.Error
	}
	result.CABundle = caBundle

	// Step 3: Verify signature
	if !config.SkipSignatureVerification {
		if err := v.VerifySignature(doc, cert); err != nil {
			result.Error = fmt.Errorf("signature verification failed: %w", err)
			return result, result.Error
		}
	}

	// Step 4: Verify certificate chain
	if !config.SkipCertificateChainVerification {
		if err := v.VerifyCertificateChain(append([]*x509.Certificate{cert}, caBundle...)); err != nil {
			result.Error = fmt.Errorf("certificate chain verification failed: %w", err)
			return result, result.Error
		}
	}

	// Step 5: Verify document freshness
	if err := v.verifyFreshness(doc, config); err != nil {
		result.Error = err
		return result, result.Error
	}

	// Step 6: Validate PCRs
	if len(config.AllowedPCRs) > 0 {
		if err := v.ValidatePCRs(doc.Payload.PCRs, config.AllowedPCRs); err != nil {
			result.Error = err
			return result, result.Error
		}
	}

	// Step 7: Validate user data
	if config.RequireUserData {
		if len(doc.Payload.UserData) == 0 {
			result.Error = errors.New("user data required but not present")
			return result, result.Error
		}
		if len(config.ExpectedUserData) > 0 {
			if !bytes.Equal(doc.Payload.UserData, config.ExpectedUserData) {
				result.Error = errors.New("user data mismatch")
				return result, result.Error
			}
		}
	}

	// Step 8: Validate nonce
	if config.RequireNonce {
		if len(doc.Payload.Nonce) == 0 {
			result.Error = errors.New("nonce required but not present")
			return result, result.Error
		}
		if len(config.ExpectedNonce) > 0 {
			if err := ValidateNonce(doc, config.ExpectedNonce); err != nil {
				result.Error = err
				return result, result.Error
			}
		}
	}

	result.Valid = true
	return result, nil
}

// VerifyRaw verifies a raw CBOR-encoded attestation document
func (v *Verifier) VerifyRaw(data []byte) (*VerificationResult, error) {
	doc, err := ParseDocument(data)
	if err != nil {
		return &VerificationResult{
			Error: fmt.Errorf("failed to parse document: %w", err),
		}, err
	}
	return v.VerifyDocument(doc)
}

// =============================================================================
// Signature Verification
// =============================================================================

// VerifySignature verifies the COSE_Sign1 signature
func (v *Verifier) VerifySignature(doc *AttestationDocument, cert *x509.Certificate) error {
	if doc == nil || cert == nil {
		return ErrSignatureInvalid
	}

	// Get public key from certificate
	pubKey, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("%w: certificate does not contain ECDSA public key", ErrSignatureInvalid)
	}

	// Verify curve is P-384 (for ES384)
	if pubKey.Curve != elliptic.P384() {
		return fmt.Errorf("%w: certificate uses wrong curve (expected P-384)", ErrSignatureInvalid)
	}

	// Build Sig_structure for COSE_Sign1
	// Sig_structure = ["Signature1", protected, external_aad, payload]
	sigStructure := buildSigStructure(doc.Protected, nil, doc.RawPayload)

	// Hash the Sig_structure with SHA-384
	hash := sha512.Sum384(sigStructure)

	// Parse signature (R || S format, each 48 bytes for P-384)
	if len(doc.Signature) != ES384SignatureSize {
		return fmt.Errorf("%w: invalid signature length %d", ErrSignatureInvalid, len(doc.Signature))
	}

	r := new(big.Int).SetBytes(doc.Signature[:48])
	s := new(big.Int).SetBytes(doc.Signature[48:])

	// Verify signature
	if !ecdsa.Verify(pubKey, hash[:], r, s) {
		return ErrSignatureInvalid
	}

	return nil
}

// buildSigStructure builds the COSE Sig_structure
func buildSigStructure(protected, externalAAD, payload []byte) []byte {
	writer := newCBORWriter()

	// Array of 4 elements
	writer.writeArrayHeader(4)

	// Context string
	writer.writeTextString("Signature1")

	// Protected header
	writer.writeByteString(protected)

	// External AAD (usually empty)
	if externalAAD == nil {
		externalAAD = []byte{}
	}
	writer.writeByteString(externalAAD)

	// Payload
	writer.writeByteString(payload)

	return writer.bytes()
}

// =============================================================================
// Certificate Chain Verification
// =============================================================================

// VerifyCertificateChain verifies the certificate chain against AWS Nitro Root CA
func (v *Verifier) VerifyCertificateChain(certs []*x509.Certificate) error {
	v.mu.RLock()
	rootCA := v.config.RootCA
	v.mu.RUnlock()

	if rootCA == nil {
		return ErrNoRootCA
	}

	if len(certs) == 0 {
		return fmt.Errorf("%w: no certificates provided", ErrCertificateChainInvalid)
	}

	// Build root pool
	rootPool := x509.NewCertPool()
	rootPool.AddCert(rootCA)

	// Build intermediate pool
	intermediatePool := x509.NewCertPool()
	for i := 1; i < len(certs); i++ {
		intermediatePool.AddCert(certs[i])
	}

	// Verify the leaf certificate
	opts := x509.VerifyOptions{
		Roots:         rootPool,
		Intermediates: intermediatePool,
		CurrentTime:   time.Now(),
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}

	if _, err := certs[0].Verify(opts); err != nil {
		return fmt.Errorf("%w: %v", ErrCertificateChainInvalid, err)
	}

	return nil
}

// =============================================================================
// PCR Validation
// =============================================================================

// ValidatePCRs validates PCR values against expected values
func (v *Verifier) ValidatePCRs(actual map[int][]byte, expected map[int][]byte) error {
	for idx, expectedValue := range expected {
		actualValue, ok := actual[idx]
		if !ok {
			return fmt.Errorf("%w: PCR%d not present in attestation", ErrPCRMismatch, idx)
		}
		if !bytes.Equal(actualValue, expectedValue) {
			return fmt.Errorf("%w: PCR%d value mismatch (expected %s, got %s)",
				ErrPCRMismatch, idx,
				hex.EncodeToString(expectedValue),
				hex.EncodeToString(actualValue))
		}
	}
	return nil
}

// =============================================================================
// Freshness Verification
// =============================================================================

// verifyFreshness verifies document timestamp is within acceptable range
func (v *Verifier) verifyFreshness(doc *AttestationDocument, config *VerifierConfig) error {
	if doc == nil || doc.Payload == nil {
		return ErrInvalidDocument
	}

	docTime := time.UnixMilli(int64(doc.Payload.Timestamp))
	now := time.Now()

	// Check if document is too old
	age := now.Sub(docTime)
	if age > config.MaxDocumentAge {
		return fmt.Errorf("%w: document age %v exceeds max %v", ErrDocumentTooOld, age, config.MaxDocumentAge)
	}

	// Check if document is from the future (with clock skew allowance)
	if docTime.After(now.Add(config.MaxClockSkew)) {
		return fmt.Errorf("%w: document timestamp %v is in the future", ErrDocumentFromFuture, docTime)
	}

	return nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// parseAWSNitroRootCA parses the embedded AWS Nitro Root CA
func parseAWSNitroRootCA() (*x509.Certificate, error) {
	block, _ := decodePEM([]byte(AWSNitroRootCAPEM))
	if block == nil {
		return nil, errors.New("failed to decode AWS Nitro Root CA PEM")
	}
	return x509.ParseCertificate(block)
}

// decodePEM decodes a PEM block (simple implementation)
func decodePEM(data []byte) ([]byte, []byte) {
	// Find BEGIN marker
	beginMarker := []byte("-----BEGIN ")
	endMarker := []byte("-----END ")

	beginIdx := bytes.Index(data, beginMarker)
	if beginIdx < 0 {
		return nil, data
	}

	// Find end of BEGIN line
	data = data[beginIdx:]
	newlineIdx := bytes.IndexByte(data, '\n')
	if newlineIdx < 0 {
		return nil, nil
	}
	data = data[newlineIdx+1:]

	// Find END marker
	endIdx := bytes.Index(data, endMarker)
	if endIdx < 0 {
		return nil, nil
	}

	// Extract base64 content
	base64Data := data[:endIdx]

	// Remove newlines and decode
	base64Data = bytes.ReplaceAll(base64Data, []byte("\n"), nil)
	base64Data = bytes.ReplaceAll(base64Data, []byte("\r"), nil)

	decoded, err := decodeBase64(base64Data)
	if err != nil {
		return nil, nil
	}

	return decoded, data[endIdx:]
}

// decodeBase64 decodes base64 data
func decodeBase64(data []byte) ([]byte, error) {
	// Standard base64 alphabet
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

	// Build decode table
	var decodeTable [256]byte
	for i := range decodeTable {
		decodeTable[i] = 0xff
	}
	for i, c := range alphabet {
		decodeTable[c] = byte(i)
	}
	decodeTable['='] = 0

	// Remove padding
	data = bytes.TrimRight(data, "=")

	// Calculate output size
	outputLen := len(data) * 6 / 8

	output := make([]byte, outputLen)
	outputIdx := 0
	buffer := 0
	bufferBits := 0

	for _, b := range data {
		val := decodeTable[b]
		if val == 0xff {
			continue // Skip invalid characters
		}

		buffer = (buffer << 6) | int(val)
		bufferBits += 6

		if bufferBits >= 8 {
			bufferBits -= 8
			if outputIdx < len(output) {
				output[outputIdx] = byte(buffer >> bufferBits)
				outputIdx++
			}
		}
	}

	return output[:outputIdx], nil
}

// isSimulatedDocument checks if a document is simulated
func isSimulatedDocument(doc *AttestationDocument) bool {
	if doc == nil || doc.Payload == nil {
		return false
	}
	// Check for simulation markers
	cert := doc.Payload.Certificate
	if len(cert) > 0 && bytes.Contains(cert, []byte("SIMULATED")) {
		return true
	}
	if bytes.HasPrefix([]byte(doc.Payload.ModuleID), []byte("sim-")) {
		return true
	}
	return false
}

// =============================================================================
// Convenience Functions
// =============================================================================

// VerifyAttestationDocument is a convenience function to verify a raw attestation
func VerifyAttestationDocument(data []byte, allowedPCRs map[int][]byte) (*VerificationResult, error) {
	config := DefaultVerifierConfig()
	config.AllowedPCRs = allowedPCRs
	verifier := NewVerifierWithConfig(config)
	return verifier.VerifyRaw(data)
}

// VerifyWithPCRHex verifies with PCRs specified as hex strings
func VerifyWithPCRHex(data []byte, allowedPCRs map[int]string) (*VerificationResult, error) {
	decoded, err := PCRMapFromHex(allowedPCRs)
	if err != nil {
		return nil, fmt.Errorf("invalid PCR values: %w", err)
	}
	return VerifyAttestationDocument(data, decoded)
}

// QuickVerify performs a quick verification with default settings
func QuickVerify(data []byte) (*VerificationResult, error) {
	verifier := NewVerifier()
	return verifier.VerifyRaw(data)
}

// =============================================================================
// PCR Policy Builder
// =============================================================================

// PCRPolicy represents a set of PCR requirements
type PCRPolicy struct {
	mu       sync.RWMutex
	expected map[int][]byte
}

// NewPCRPolicy creates a new PCR policy
func NewPCRPolicy() *PCRPolicy {
	return &PCRPolicy{
		expected: make(map[int][]byte),
	}
}

// AddPCR adds an expected PCR value
func (p *PCRPolicy) AddPCR(index int, value []byte) error {
	if index < 0 || index >= PCRCount {
		return fmt.Errorf("invalid PCR index: %d", index)
	}
	if len(value) != PCRDigestSize {
		return fmt.Errorf("invalid PCR value size: expected %d, got %d", PCRDigestSize, len(value))
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.expected[index] = make([]byte, len(value))
	copy(p.expected[index], value)
	return nil
}

// AddPCRHex adds an expected PCR value from hex string
func (p *PCRPolicy) AddPCRHex(index int, hexValue string) error {
	value, err := hex.DecodeString(hexValue)
	if err != nil {
		return fmt.Errorf("invalid hex: %w", err)
	}
	return p.AddPCR(index, value)
}

// GetExpected returns the expected PCR values
func (p *PCRPolicy) GetExpected() map[int][]byte {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make(map[int][]byte, len(p.expected))
	for k, v := range p.expected {
		result[k] = make([]byte, len(v))
		copy(result[k], v)
	}
	return result
}

// Validate validates PCRs against the policy
func (p *PCRPolicy) Validate(actual map[int][]byte) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for idx, expected := range p.expected {
		actual, ok := actual[idx]
		if !ok {
			return fmt.Errorf("PCR%d not present", idx)
		}
		if !bytes.Equal(actual, expected) {
			return fmt.Errorf("PCR%d mismatch", idx)
		}
	}
	return nil
}
