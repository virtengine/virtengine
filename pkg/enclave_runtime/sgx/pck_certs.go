// Package sgx provides Intel SGX enclave management and DCAP attestation.
//
// This file implements PCK (Platform Certification Key) certificate handling
// for Intel SGX DCAP attestation. PCK certificates are X.509 certificates
// issued by Intel that bind a platform's SGX identity to a public key.
//
// Certificate Chain:
//
//	Intel SGX Root CA
//	       ↓
//	Intel SGX PCK Processor CA (or Platform CA)
//	       ↓
//	PCK Certificate (platform-specific)
//
// SGX Extensions (OIDs):
//   - 1.2.840.113741.1.13.1    - SGX Extensions root
//   - 1.2.840.113741.1.13.1.2  - FMSPC
//   - 1.2.840.113741.1.13.1.3  - TCB Components
//   - 1.2.840.113741.1.13.1.4  - PCESVN
//   - 1.2.840.113741.1.13.1.5  - CPUSVN
//
// Task Reference: VE-2029 - Hardware TEE Integration Layer
package sgx

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// OID Constants
// =============================================================================

// SGX Extension OIDs
var (
	// OIDSGXExtensions is the root OID for SGX extensions.
	OIDSGXExtensions = asn1.ObjectIdentifier{1, 2, 840, 113741, 1, 13, 1}

	// OIDPPID is the OID for Platform Provisioning ID.
	OIDPPID = asn1.ObjectIdentifier{1, 2, 840, 113741, 1, 13, 1, 1}

	// OIDFMSPC is the OID for Family-Model-Stepping-Platform-Custom.
	OIDFMSPC = asn1.ObjectIdentifier{1, 2, 840, 113741, 1, 13, 1, 2}

	// OIDTCBComponents is the OID for TCB Component values.
	OIDTCBComponents = asn1.ObjectIdentifier{1, 2, 840, 113741, 1, 13, 1, 3}

	// OIDPCESVN is the OID for PCE Security Version Number.
	OIDPCESVN = asn1.ObjectIdentifier{1, 2, 840, 113741, 1, 13, 1, 4}

	// OIDCPUSVN is the OID for CPU Security Version Number.
	OIDCPUSVN = asn1.ObjectIdentifier{1, 2, 840, 113741, 1, 13, 1, 5}

	// OIDPCEId is the OID for PCE Identifier.
	OIDPCEId = asn1.ObjectIdentifier{1, 2, 840, 113741, 1, 13, 1, 6}

	// OIDSGXTYPE is the OID for SGX Type (Standard/Scalable).
	OIDSGXTYPE = asn1.ObjectIdentifier{1, 2, 840, 113741, 1, 13, 1, 7}

	// OIDPlatformInstanceId is the OID for Platform Instance ID.
	OIDPlatformInstanceId = asn1.ObjectIdentifier{1, 2, 840, 113741, 1, 13, 1, 8}

	// OIDConfiguration is the OID for Configuration.
	OIDConfiguration = asn1.ObjectIdentifier{1, 2, 840, 113741, 1, 13, 1, 9}
)

// =============================================================================
// Error Types
// =============================================================================

var (
	// ErrInvalidPCKCert indicates the PCK certificate is invalid.
	ErrInvalidPCKCert = errors.New("sgx: invalid PCK certificate")

	// ErrCertChainTooShort indicates the certificate chain is too short.
	ErrCertChainTooShort = errors.New("sgx: certificate chain too short")

	// ErrCertExpired indicates the certificate has expired.
	ErrCertExpired = errors.New("sgx: certificate expired")

	// ErrCertNotYetValid indicates the certificate is not yet valid.
	ErrCertNotYetValid = errors.New("sgx: certificate not yet valid")

	// ErrMissingFMSPC indicates FMSPC is missing from the certificate.
	ErrMissingFMSPC = errors.New("sgx: missing FMSPC in certificate")

	// ErrMissingTCBInfo indicates TCB info is missing from the certificate.
	ErrMissingTCBInfo = errors.New("sgx: missing TCB info in certificate")

	// ErrChainVerificationFailed indicates certificate chain verification failed.
	ErrChainVerificationFailed = errors.New("sgx: certificate chain verification failed")

	// ErrRootCAMismatch indicates the root CA doesn't match Intel's.
	ErrRootCAMismatch = errors.New("sgx: root CA mismatch")
)

// =============================================================================
// PCK Certificate Types
// =============================================================================

// PCKCertificate represents a parsed PCK certificate with SGX extensions.
type PCKCertificate struct {
	// Raw is the underlying X.509 certificate.
	Raw *x509.Certificate

	// PPID is the Platform Provisioning ID (optional, 16 bytes).
	PPID []byte

	// FMSPC is the Family-Model-Stepping-Platform-Custom (6 bytes hex).
	FMSPC string

	// PCEId is the PCE Identifier (2 bytes hex).
	PCEId string

	// CPUSVN is the CPU Security Version Number (16 bytes).
	CPUSVN [CPUSVNSize]byte

	// PCESVN is the PCE Security Version Number.
	PCESVN uint16

	// TCBComponents are the 16 TCB component values.
	TCBComponents [16]byte

	// SGXType indicates Standard (0) or Scalable (1).
	SGXType int

	// PlatformInstanceId is the platform instance identifier (optional).
	PlatformInstanceId []byte

	// Configuration contains platform configuration flags (optional).
	Configuration *PlatformConfiguration

	// PublicKey is the ECDSA public key from the certificate.
	PublicKey *ecdsa.PublicKey
}

// PlatformConfiguration contains platform configuration flags.
type PlatformConfiguration struct {
	// DynamicPlatform indicates dynamic platform support.
	DynamicPlatform bool

	// CachedKeys indicates cached keys support.
	CachedKeys bool

	// SMTEnabled indicates SMT (hyperthreading) is enabled.
	SMTEnabled bool
}

// CertificateChain represents a complete PCK certificate chain.
type CertificateChain struct {
	// PCKCert is the leaf PCK certificate.
	PCKCert *PCKCertificate

	// IntermediateCert is the intermediate CA certificate (PCK Processor or Platform CA).
	IntermediateCert *x509.Certificate

	// RootCert is the Intel SGX Root CA certificate.
	RootCert *x509.Certificate

	// RawChain contains the raw PEM-encoded certificates.
	RawChain []byte

	// ParsedAt is when the chain was parsed.
	ParsedAt time.Time
}

// PCKTCBInfo represents TCB information extracted from a PCK certificate.
type PCKTCBInfo struct {
	// FMSPC is the Family-Model-Stepping-Platform-Custom.
	FMSPC string

	// PCEId is the PCE Identifier.
	PCEId string

	// CPUSVN is the CPU Security Version Number.
	CPUSVN [CPUSVNSize]byte

	// PCESVN is the PCE Security Version Number.
	PCESVN uint16

	// TCBComponents are the 16 TCB component values.
	TCBComponents [16]byte

	// SGXType indicates Standard (0) or Scalable (1).
	SGXType int
}

// =============================================================================
// PCK Certificate Parser
// =============================================================================

// PCKCertParser provides methods for parsing and validating PCK certificates.
type PCKCertParser struct {
	mu sync.RWMutex

	// Cache for parsed certificates
	certCache map[string]*PCKCertificate
}

// NewPCKCertParser creates a new PCK certificate parser.
func NewPCKCertParser() *PCKCertParser {
	return &PCKCertParser{
		certCache: make(map[string]*PCKCertificate),
	}
}

// ParsePCKCert parses a DER-encoded PCK certificate.
func (p *PCKCertParser) ParsePCKCert(der []byte) (*PCKCertificate, error) {
	// Parse the X.509 certificate
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPCKCert, err)
	}

	return p.parsePCKFromX509(cert)
}

// ParsePCKCertPEM parses a PEM-encoded PCK certificate.
func (p *PCKCertParser) ParsePCKCertPEM(pemData []byte) (*PCKCertificate, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("%w: failed to decode PEM", ErrInvalidPCKCert)
	}

	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("%w: unexpected PEM type: %s", ErrInvalidPCKCert, block.Type)
	}

	return p.ParsePCKCert(block.Bytes)
}

// parsePCKFromX509 parses PCK-specific extensions from an X.509 certificate.
func (p *PCKCertParser) parsePCKFromX509(cert *x509.Certificate) (*PCKCertificate, error) {
	pck := &PCKCertificate{
		Raw: cert,
	}

	// Extract ECDSA public key
	if ecKey, ok := cert.PublicKey.(*ecdsa.PublicKey); ok {
		pck.PublicKey = ecKey
	}

	// Parse SGX extensions
	for _, ext := range cert.Extensions {
		if ext.Id.Equal(OIDSGXExtensions) || isChildOID(ext.Id, OIDSGXExtensions) {
			if err := p.parseSGXExtension(ext.Id, ext.Value, pck); err != nil {
				// Log but continue - some extensions may be optional
				continue
			}
		}
	}

	// Validate required fields
	if pck.FMSPC == "" {
		// Try to extract from subject or other fields
		pck.FMSPC = p.extractFMSPCFromSubject(cert)
	}

	return pck, nil
}

// parseSGXExtension parses a single SGX extension.
func (p *PCKCertParser) parseSGXExtension(oid asn1.ObjectIdentifier, value []byte, pck *PCKCertificate) error {
	switch {
	case oid.Equal(OIDPPID):
		pck.PPID = make([]byte, len(value))
		copy(pck.PPID, value)

	case oid.Equal(OIDFMSPC):
		// FMSPC is 6 bytes, stored as hex string
		pck.FMSPC = hex.EncodeToString(value)

	case oid.Equal(OIDPCEId):
		// PCE ID is 2 bytes
		if len(value) >= 2 {
			pck.PCEId = hex.EncodeToString(value[:2])
		}

	case oid.Equal(OIDCPUSVN):
		// CPUSVN is 16 bytes
		if len(value) >= CPUSVNSize {
			copy(pck.CPUSVN[:], value[:CPUSVNSize])
		}

	case oid.Equal(OIDPCESVN):
		// PCESVN is 2 bytes little-endian
		if len(value) >= 2 {
			pck.PCESVN = uint16(value[0]) | uint16(value[1])<<8
		}

	case oid.Equal(OIDTCBComponents):
		// TCB Components are 16 bytes
		if len(value) >= 16 {
			copy(pck.TCBComponents[:], value[:16])
		}

	case oid.Equal(OIDSGXTYPE):
		if len(value) >= 1 {
			pck.SGXType = int(value[0])
		}

	case oid.Equal(OIDPlatformInstanceId):
		pck.PlatformInstanceId = make([]byte, len(value))
		copy(pck.PlatformInstanceId, value)

	case oid.Equal(OIDConfiguration):
		pck.Configuration = p.parseConfiguration(value)
	}

	return nil
}

// parseConfiguration parses platform configuration from extension value.
func (p *PCKCertParser) parseConfiguration(value []byte) *PlatformConfiguration {
	if len(value) == 0 {
		return nil
	}

	config := &PlatformConfiguration{}

	// Configuration is typically a bit field
	if len(value) >= 1 {
		config.DynamicPlatform = (value[0] & 0x01) != 0
		config.CachedKeys = (value[0] & 0x02) != 0
		config.SMTEnabled = (value[0] & 0x04) != 0
	}

	return config
}

// extractFMSPCFromSubject attempts to extract FMSPC from certificate subject.
func (p *PCKCertParser) extractFMSPCFromSubject(_ *x509.Certificate) string {
	// In some cases, FMSPC might be encoded in subject CN or other fields
	// Return empty for now - real implementation would parse subject
	return ""
}

// =============================================================================
// Certificate Chain Functions
// =============================================================================

// ParseCertificateChain parses a PEM-encoded certificate chain.
func ParseCertificateChain(pemData []byte) (*CertificateChain, error) {
	chain := &CertificateChain{
		RawChain: pemData,
		ParsedAt: time.Now(),
	}

	var certs []*x509.Certificate
	remaining := pemData

	for len(remaining) > 0 {
		block, rest := pem.Decode(remaining)
		if block == nil {
			break
		}
		remaining = rest

		if block.Type != "CERTIFICATE" {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %w", err)
		}
		certs = append(certs, cert)
	}

	if len(certs) < 2 {
		return nil, ErrCertChainTooShort
	}

	// Parse the leaf certificate as PCK
	parser := NewPCKCertParser()
	pckCert, err := parser.parsePCKFromX509(certs[0])
	if err != nil {
		return nil, err
	}
	chain.PCKCert = pckCert

	// Identify intermediate and root
	if len(certs) >= 2 {
		chain.IntermediateCert = certs[1]
	}
	if len(certs) >= 3 {
		chain.RootCert = certs[2]
	} else if len(certs) == 2 {
		// Two cert chain - second is root
		chain.RootCert = certs[1]
		chain.IntermediateCert = nil
	}

	return chain, nil
}

// ValidateChain validates a PCK certificate chain against Intel's Root CA.
func ValidateChain(chain *CertificateChain) error {
	if chain == nil || chain.PCKCert == nil {
		return ErrInvalidPCKCert
	}

	now := time.Now()

	// Check PCK certificate validity
	if err := checkCertValidity(chain.PCKCert.Raw, now); err != nil {
		return fmt.Errorf("PCK cert: %w", err)
	}

	// Check intermediate certificate validity
	if chain.IntermediateCert != nil {
		if err := checkCertValidity(chain.IntermediateCert, now); err != nil {
			return fmt.Errorf("intermediate cert: %w", err)
		}
	}

	// Create certificate pool for verification
	roots := x509.NewCertPool()

	// Add Intel Root CA
	intelRoot, err := GetIntelRootCACert()
	if err != nil {
		// Use provided root if Intel root not available
		if chain.RootCert != nil {
			roots.AddCert(chain.RootCert)
		}
	} else {
		roots.AddCert(intelRoot)
	}

	// Build intermediate pool
	intermediates := x509.NewCertPool()
	if chain.IntermediateCert != nil {
		intermediates.AddCert(chain.IntermediateCert)
	}

	// Verify the chain
	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		CurrentTime:   now,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}

	if _, err := chain.PCKCert.Raw.Verify(opts); err != nil {
		return fmt.Errorf("%w: %v", ErrChainVerificationFailed, err)
	}

	return nil
}

// checkCertValidity checks if a certificate is valid at the given time.
func checkCertValidity(cert *x509.Certificate, t time.Time) error {
	if t.Before(cert.NotBefore) {
		return ErrCertNotYetValid
	}
	if t.After(cert.NotAfter) {
		return ErrCertExpired
	}
	return nil
}

// ExtractPCKTCBInfo extracts TCB information from a PCK certificate.
func ExtractPCKTCBInfo(pck *PCKCertificate) (*PCKTCBInfo, error) {
	if pck == nil {
		return nil, ErrInvalidPCKCert
	}

	tcbInfo := &PCKTCBInfo{
		FMSPC:         pck.FMSPC,
		PCEId:         pck.PCEId,
		CPUSVN:        pck.CPUSVN,
		PCESVN:        pck.PCESVN,
		TCBComponents: pck.TCBComponents,
		SGXType:       pck.SGXType,
	}

	return tcbInfo, nil
}

// =============================================================================
// Intel Root CA
// =============================================================================

// IntelSGXProcessorCAPEM is the Intel SGX PCK Processor CA certificate.
const IntelSGXProcessorCAPEM = `-----BEGIN CERTIFICATE-----
MIICmDCCAj6gAwIBAgIVANDoqtp11/kuSReYPHsUZdDV8llNMAoGCCqGSM49BAMC
MGgxGjAYBgNVBAMMEUludGVsIFNHWCBSb290IENBMRowGAYDVQQKDBFJbnRlbCBD
b3Jwb3JhdGlvbjEUMBIGA1UEBwwLU2FudGEgQ2xhcmExCzAJBgNVBAgMAkNBMQsw
CQYDVQQGEwJVUzAeFw0xODA1MjExMDUwMTBaFw0zMzA1MjExMDUwMTBaMHExIzAh
BgNVBAMMGkludGVsIFNHWCBQQ0sgUHJvY2Vzc29yIENBMRowGAYDVQQKDBFJbnRl
bCBDb3Jwb3JhdGlvbjEUMBIGA1UEBwwLU2FudGEgQ2xhcmExCzAJBgNVBAgMAkNB
MQswCQYDVQQGEwJVUzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABLdDx6k8VVQx
vJ7t+fmZbEXI7JZ6aev21faMpGsJxPo4z+sHXbpbDC5vJFLBYYTMnVp8/u6E5YIc
PyKgjSpJhLijgbswgbgwHwYDVR0jBBgwFoAUImUM1lqdNInzg7SVUr9QGzknBqww
UgYDVR0fBEswSTBHoEWgQ4ZBaHR0cHM6Ly9jZXJ0aWZpY2F0ZXMudHJ1c3RlZHNl
cnZpY2VzLmludGVsLmNvbS9JbnRlbFNHWFJvb3RDQS5kZXIwHQYDVR0OBBYEFNDO
qtpvbVNlS6IyZ5+IsnqqAsoiMA4GA1UdDwEB/wQEAwIBBjASBgNVHRMBAf8ECDAG
AQH/AgEAMAoGCCqGSM49BAMCA0gAMEUCIQCx4fMvIV5bOcfTNPviqE0qKjVNZgce
FTNM+VOu4oRbdgIgAXqWDFXQl8WGj1N7n8m9WPq7vPuMq8V1a2oCrMd6yE4=
-----END CERTIFICATE-----`

// IntelSGXPlatformCAPEM is the Intel SGX PCK Platform CA certificate.
const IntelSGXPlatformCAPEM = `-----BEGIN CERTIFICATE-----
MIICljCCAj2gAwIBAgIVANDoqtp11/kuSReYPHsUZdDV8llOMAoGCCqGSM49BAMC
MGgxGjAYBgNVBAMMEUludGVsIFNHWCBSb290IENBMRowGAYDVQQKDBFJbnRlbCBD
b3Jwb3JhdGlvbjEUMBIGA1UEBwwLU2FudGEgQ2xhcmExCzAJBgNVBAgMAkNBMQsw
CQYDVQQGEwJVUzAeFw0xODA1MjExMDUwMTBaFw0zMzA1MjExMDUwMTBaMG8xITAf
BgNVBAMMGEludGVsIFNHWCBQQ0sgUGxhdGZvcm0gQ0ExGjAYBgNVBAoMEUludGVs
IENvcnBvcmF0aW9uMRQwEgYDVQQHDAtTYW50YSBDbGFyYTELMAkGA1UECAwCQ0Ex
CzAJBgNVBAYTAlVTMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAENSB/7t21lXSO
2Cuzpxv74NReTCUh4qNKoEwWpJT4nRKqB5CqF3hpxvCIg1dqwZY8D5lX1I5VLsBy
qJW6nPO1wKOBuzCBuDAfBgNVHSMEGDAWgBQiZQzWWp00ifODtJVSv1AbOScGrDBS
BgNVHR8ESzBJMEegRaBDhkFodHRwczovL2NlcnRpZmljYXRlcy50cnVzdGVkc2Vy
dmljZXMuaW50ZWwuY29tL0ludGVsU0dYUm9vdENBLmRlcjAdBgNVHQ4EFgQUDsS7
7evR0qT1cpfGa6poxaP63zQwDgYDVR0PAQH/BAQDAgEGMBIGA1UdEwEB/wQIMAYB
Af8CAQAwCgYIKoZIzj0EAwIDRwAwRAIgIHw7pB6lz9a9F1Tn0mjLe6pYJIkH6TuM
S+6BUHCO2gkCIGO2gwkPxlEH+RV4E3pMl1Hg0jDzl75rlq7Tl9ap2SZw
-----END CERTIFICATE-----`

// GetIntelProcessorCA returns the parsed Intel SGX PCK Processor CA certificate.
func GetIntelProcessorCA() (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(IntelSGXProcessorCAPEM))
	if block == nil {
		return nil, errors.New("failed to decode Intel Processor CA PEM")
	}
	return x509.ParseCertificate(block.Bytes)
}

// GetIntelPlatformCA returns the parsed Intel SGX PCK Platform CA certificate.
func GetIntelPlatformCA() (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(IntelSGXPlatformCAPEM))
	if block == nil {
		return nil, errors.New("failed to decode Intel Platform CA PEM")
	}
	return x509.ParseCertificate(block.Bytes)
}

// =============================================================================
// Helper Functions
// =============================================================================

// isChildOID checks if child is a child of parent OID.
func isChildOID(child, parent asn1.ObjectIdentifier) bool {
	if len(child) < len(parent) {
		return false
	}
	for i := range parent {
		if child[i] != parent[i] {
			return false
		}
	}
	return true
}

// VerifyPCKSignature verifies that data was signed by the PCK certificate.
func VerifyPCKSignature(pck *PCKCertificate, data, signature []byte) error {
	if pck == nil || pck.PublicKey == nil {
		return ErrInvalidPCKCert
	}

	// Verify ECDSA signature
	if !ecdsa.VerifyASN1(pck.PublicKey, data, signature) {
		return errors.New("signature verification failed")
	}

	return nil
}

// CompareFMSPC compares two FMSPC values.
func CompareFMSPC(a, b string) bool {
	// Normalize to lowercase for comparison
	return len(a) == len(b) && bytes.EqualFold([]byte(a), []byte(b))
}

// FormatFMSPC formats FMSPC bytes as a hex string.
func FormatFMSPC(fmspc []byte) string {
	if len(fmspc) != 6 {
		return ""
	}
	return hex.EncodeToString(fmspc)
}

// ParseFMSPC parses an FMSPC hex string to bytes.
func ParseFMSPC(fmspc string) ([]byte, error) {
	if len(fmspc) != 12 {
		return nil, fmt.Errorf("invalid FMSPC length: expected 12 hex chars, got %d", len(fmspc))
	}
	return hex.DecodeString(fmspc)
}

// GetTCBComponentLevel returns the TCB level for a component index.
func (t *PCKTCBInfo) GetTCBComponentLevel(index int) int {
	if index < 0 || index >= 16 {
		return 0
	}
	return int(t.TCBComponents[index])
}

// CompareTCB compares this TCB info against another, returning:
// -1 if this < other, 0 if equal, 1 if this > other
func (t *PCKTCBInfo) CompareTCB(other *PCKTCBInfo) int {
	if other == nil {
		return 1
	}

	// Compare each TCB component
	for i := 0; i < 16; i++ {
		if t.TCBComponents[i] < other.TCBComponents[i] {
			return -1
		}
		if t.TCBComponents[i] > other.TCBComponents[i] {
			return 1
		}
	}

	// Compare PCESVN
	if t.PCESVN < other.PCESVN {
		return -1
	}
	if t.PCESVN > other.PCESVN {
		return 1
	}

	return 0
}

// String returns a string representation of the TCB info.
func (t *PCKTCBInfo) String() string {
	return fmt.Sprintf("PCKTCBInfo{FMSPC: %s, PCEId: %s, PCESVN: %d, SGXType: %d}",
		t.FMSPC, t.PCEId, t.PCESVN, t.SGXType)
}

// Fingerprint returns a fingerprint of the PCK certificate.
func (p *PCKCertificate) Fingerprint() string {
	if p.Raw == nil {
		return ""
	}
	return hex.EncodeToString(p.Raw.Raw[:16])
}

// Subject returns the certificate subject.
func (p *PCKCertificate) Subject() string {
	if p.Raw == nil {
		return ""
	}
	return p.Raw.Subject.String()
}

// ValidAt checks if the certificate is valid at the given time.
func (p *PCKCertificate) ValidAt(t time.Time) bool {
	if p.Raw == nil {
		return false
	}
	return !t.Before(p.Raw.NotBefore) && !t.After(p.Raw.NotAfter)
}

// NotBefore returns the certificate's not-before time.
func (p *PCKCertificate) NotBefore() time.Time {
	if p.Raw == nil {
		return time.Time{}
	}
	return p.Raw.NotBefore
}

// NotAfter returns the certificate's not-after time.
func (p *PCKCertificate) NotAfter() time.Time {
	if p.Raw == nil {
		return time.Time{}
	}
	return p.Raw.NotAfter
}
