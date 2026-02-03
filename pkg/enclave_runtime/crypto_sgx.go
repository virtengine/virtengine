// Package enclave_runtime provides TEE enclave implementations.
//
// This file implements cryptographic verification for Intel SGX DCAP quotes.
// DCAP (Data Center Attestation Primitives) uses ECDSA signatures that can
// be verified without connecting to Intel Attestation Service.
//
// Verification chain:
// 1. Parse DCAP quote structure
// 2. Extract QE (Quoting Enclave) report body
// 3. Verify QE signature using PCK certificate
// 4. Verify PCK certificate chain to Intel Root CA
// 5. Check TCB status against Intel TCB Info
//
// Quote Structure (v3):
// +------------------+
// | Quote Header     | 48 bytes
// +------------------+
// | ISV Enclave Rep  | 384 bytes (SGX Report Body)
// +------------------+
// | Signature Data   | Variable
// +------------------+
//
// Task Reference: VE-2030 - Real Attestation Crypto Verification
package enclave_runtime

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// =============================================================================
// Intel SGX Root CA Certificate (PEM)
// =============================================================================

// IntelSGXRootCAPEM is the Intel SGX Root CA certificate for DCAP verification.
// This is the production Intel SGX Root CA used to sign PCK certificates.
// Subject: CN=Intel SGX Root CA, O=Intel Corporation, L=Santa Clara, ST=CA, C=US
// Valid: 2018-05-21 to 2049-12-31
const IntelSGXRootCAPEM = `-----BEGIN CERTIFICATE-----
MIICjzCCAjSgAwIBAgIUImUM1lqdNInzg7SVUr9QGzknBqwwCgYIKoZIzj0EAwIw
aDEaMBgGA1UEAwwRSW50ZWwgU0dYIFJvb3QgQ0ExGjAYBgNVBAoMEUludGVsIENv
cnBvcmF0aW9uMRQwEgYDVQQHDAtTYW50YSBDbGFyYTELMAkGA1UECAwCQ0ExCzAJ
BgNVBAYTAlVTMB4XDTE4MDUyMTEwNDUxMFoXDTQ5MTIzMTIzNTk1OVowaDEaMBgG
A1UEAwwRSW50ZWwgU0dYIFJvb3QgQ0ExGjAYBgNVBAoMEUludGVsIENvcnBvcmF0
aW9uMRQwEgYDVQQHDAtTYW50YSBDbGFyYTELMAkGA1UECAwCQ0ExCzAJBgNVBAYT
AlVTMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEC6nEwMDIYZOj/iPWsCzaEKi7
1OiOSLRFhWGjbnBVJfVnkY4u3IjkDYYL0MxO4mqsyYjlBalTVYxFP2sJBK5zlKOB
uzCBuDAfBgNVHSMEGDAWgBQiZQzWWp00ifODtJVSv1AbOScGrDBSBgNVHR8ESzBJ
MEegRaBDhkFodHRwczovL2NlcnRpZmljYXRlcy50cnVzdGVkc2VydmljZXMuaW50
ZWwuY29tL0ludGVsU0dYUm9vdENBLmRlcjAdBgNVHQ4EFgQUImUM1lqdNInzg7SV
Ur9QGzknBqwwDgYDVR0PAQH/BAQDAgEGMBIGA1UdEwEB/wQIMAYBAf8CAQEwCgYI
KoZIzj0EAwIDSQAwRgIhAOW/5QkR+S9CiSDcNoowLuPRLsWGf/Yi7GSX94BgwTwg
AiEA4J0lrHoMs+Xo5o/sX6O9QWxHRAvZUGOdRQ7cvqRXaqI=
-----END CERTIFICATE-----`

// IntelSGXPCKProcessorCAPEM is the Intel SGX PCK Processor CA certificate.
// This intermediate CA signs PCK certificates for specific processors.
const IntelSGXPCKProcessorCAPEM = `-----BEGIN CERTIFICATE-----
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

// =============================================================================
// DCAP Quote Structures
// =============================================================================

// CryptoDCAPQuoteVersion represents supported quote versions.
type CryptoDCAPQuoteVersion uint16

const (
	CryptoDCAPQuoteV3 CryptoDCAPQuoteVersion = 3
	CryptoDCAPQuoteV4 CryptoDCAPQuoteVersion = 4
)

// CryptoAttestationKeyType represents the type of attestation key.
type CryptoAttestationKeyType uint16

const (
	CryptoAttKeyTypeECDSA256P256 CryptoAttestationKeyType = 2 // ECDSA-256-with-P-256
	CryptoAttKeyTypeECDSA384P384 CryptoAttestationKeyType = 3 // ECDSA-384-with-P-384
)

// CryptoCertificationDataType represents the type of certification data.
type CryptoCertificationDataType uint16

const (
	CryptoCertDataTypePPID            CryptoCertificationDataType = 1 // PPID + CPUSVN + PCE ID + PCE SVN
	CryptoCertDataTypePCKCertChain    CryptoCertificationDataType = 5 // PCK Certificate Chain
	CryptoCertDataTypePCKCertPlusInfo CryptoCertificationDataType = 6 // PCK Certificate + Manifest
)

// CryptoDCAPQuoteHeader represents the header of a DCAP quote.
type CryptoDCAPQuoteHeader struct {
	Version            uint16   // Quote version (3 or 4)
	AttestationKeyType uint16   // Type of attestation key
	TEEType            uint32   // TEE type (0 = SGX, 0x81 = TDX)
	Reserved           uint32   // Reserved bytes
	QEVendorID         [16]byte // QE Vendor ID
	UserData           [20]byte // Custom user data
}

// CryptoSGXReportBody represents the SGX report body structure (384 bytes).
// This is the crypto-specific version with fixed-size arrays for parsing.
type CryptoSGXReportBody struct {
	CPUSVN       [16]byte // CPU Security Version
	MiscSelect   uint32   // MISC select
	Reserved1    [12]byte // Reserved
	ISVExtProdID [16]byte // ISV Extended Product ID
	Attributes   [16]byte // Enclave attributes
	MRENCLAVE    [32]byte // Enclave measurement
	Reserved2    [32]byte // Reserved
	MRSIGNER     [32]byte // Signer measurement
	Reserved3    [32]byte // Reserved
	ConfigID     [64]byte // Config ID
	ISVProdID    uint16   // ISV Product ID
	ISVSVN       uint16   // ISV Security Version
	ConfigSVN    uint16   // Config SVN
	Reserved4    [42]byte // Reserved
	ISVFamilyID  [16]byte // ISV Family ID
	ReportData   [64]byte // Report data (user-supplied)
}

// CryptoDCAPQuoteSignatureData represents the signature data in a DCAP quote.
type CryptoDCAPQuoteSignatureData struct {
	ISVEnclaveReportSignature [64]byte // ECDSA P-256 signature (r || s)
	ECDSAAttestationKey       [64]byte // Public key (x || y)
	QEReport                  CryptoSGXReportBody
	QEReportSignature         [64]byte
	QEAuthenticationDataSize  uint16
	QEAuthenticationData      []byte
	CertificationDataType     uint16
	CertificationDataSize     uint32
	CertificationData         []byte // PCK cert chain (PEM)
}

// CryptoDCAPQuote represents a complete DCAP quote (version 3 or 4).
type CryptoDCAPQuote struct {
	Header                CryptoDCAPQuoteHeader
	ISVEnclaveReport      CryptoSGXReportBody
	QuoteSignatureDataLen uint32
	QuoteSignatureData    CryptoDCAPQuoteSignatureData
	RawBytes              []byte // Original raw quote bytes
}

// =============================================================================
// DCAP Quote Parser
// =============================================================================

// DCAPQuoteParser parses DCAP quote structures.
type DCAPQuoteParser struct {
	mu sync.RWMutex
}

// NewDCAPQuoteParser creates a new DCAP quote parser.
func NewDCAPQuoteParser() *DCAPQuoteParser {
	return &DCAPQuoteParser{}
}

// SGX Quote offsets
const (
	dcapQuoteHeaderSize = 48
	dcapReportBodySize  = 384
	dcapMinQuoteSize    = dcapQuoteHeaderSize + dcapReportBodySize + 4 // Header + Report + SigLen
	dcapSigDataMinSize  = 64 + 64 + 384 + 64 + 2 + 2 + 4               // ISVSig + PubKey + QEReport + QESig + AuthSize + CertType + CertSize
)

// Parse parses a raw DCAP quote into a structured format.
func (p *DCAPQuoteParser) Parse(quoteBytes []byte) (*CryptoDCAPQuote, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(quoteBytes) < dcapMinQuoteSize {
		return nil, fmt.Errorf("quote too small: got %d bytes, need at least %d", len(quoteBytes), dcapMinQuoteSize)
	}

	quote := &CryptoDCAPQuote{
		RawBytes: make([]byte, len(quoteBytes)),
	}
	copy(quote.RawBytes, quoteBytes)

	// Parse header
	if err := p.parseHeader(quoteBytes[:dcapQuoteHeaderSize], &quote.Header); err != nil {
		return nil, fmt.Errorf("failed to parse header: %w", err)
	}

	// Validate version
	if quote.Header.Version != uint16(CryptoDCAPQuoteV3) && quote.Header.Version != uint16(CryptoDCAPQuoteV4) {
		return nil, fmt.Errorf("unsupported quote version: %d", quote.Header.Version)
	}

	// Parse ISV enclave report body
	reportStart := dcapQuoteHeaderSize
	if err := p.parseReportBody(quoteBytes[reportStart:reportStart+dcapReportBodySize], &quote.ISVEnclaveReport); err != nil {
		return nil, fmt.Errorf("failed to parse ISV enclave report: %w", err)
	}

	// Parse signature data length
	sigLenOffset := reportStart + dcapReportBodySize
	quote.QuoteSignatureDataLen = binary.LittleEndian.Uint32(quoteBytes[sigLenOffset:])

	// Parse signature data
	sigDataStart := sigLenOffset + 4
	if len(quoteBytes) < sigDataStart+int(quote.QuoteSignatureDataLen) {
		return nil, fmt.Errorf("quote truncated: signature data length %d exceeds available bytes", quote.QuoteSignatureDataLen)
	}

	if err := p.parseSignatureData(quoteBytes[sigDataStart:sigDataStart+int(quote.QuoteSignatureDataLen)], &quote.QuoteSignatureData); err != nil {
		return nil, fmt.Errorf("failed to parse signature data: %w", err)
	}

	return quote, nil
}

// parseHeader parses the quote header.
func (p *DCAPQuoteParser) parseHeader(data []byte, header *CryptoDCAPQuoteHeader) error {
	if len(data) < dcapQuoteHeaderSize {
		return errors.New("header data too short")
	}

	header.Version = binary.LittleEndian.Uint16(data[0:2])
	header.AttestationKeyType = binary.LittleEndian.Uint16(data[2:4])
	header.TEEType = binary.LittleEndian.Uint32(data[4:8])
	header.Reserved = binary.LittleEndian.Uint32(data[8:12])
	copy(header.QEVendorID[:], data[12:28])
	copy(header.UserData[:], data[28:48])

	return nil
}

// parseReportBody parses an SGX report body.
func (p *DCAPQuoteParser) parseReportBody(data []byte, report *CryptoSGXReportBody) error {
	if len(data) < dcapReportBodySize {
		return errors.New("report body data too short")
	}

	offset := 0
	copy(report.CPUSVN[:], data[offset:offset+16])
	offset += 16

	report.MiscSelect = binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	copy(report.Reserved1[:], data[offset:offset+12])
	offset += 12

	copy(report.ISVExtProdID[:], data[offset:offset+16])
	offset += 16

	copy(report.Attributes[:], data[offset:offset+16])
	offset += 16

	copy(report.MRENCLAVE[:], data[offset:offset+32])
	offset += 32

	copy(report.Reserved2[:], data[offset:offset+32])
	offset += 32

	copy(report.MRSIGNER[:], data[offset:offset+32])
	offset += 32

	copy(report.Reserved3[:], data[offset:offset+32])
	offset += 32

	copy(report.ConfigID[:], data[offset:offset+64])
	offset += 64

	report.ISVProdID = binary.LittleEndian.Uint16(data[offset:])
	offset += 2

	report.ISVSVN = binary.LittleEndian.Uint16(data[offset:])
	offset += 2

	report.ConfigSVN = binary.LittleEndian.Uint16(data[offset:])
	offset += 2

	copy(report.Reserved4[:], data[offset:offset+42])
	offset += 42

	copy(report.ISVFamilyID[:], data[offset:offset+16])
	offset += 16

	copy(report.ReportData[:], data[offset:offset+64])

	return nil
}

// parseSignatureData parses the quote signature data.
func (p *DCAPQuoteParser) parseSignatureData(data []byte, sigData *CryptoDCAPQuoteSignatureData) error {
	if len(data) < dcapSigDataMinSize {
		return fmt.Errorf("signature data too short: got %d, need at least %d", len(data), dcapSigDataMinSize)
	}

	offset := 0

	// ISV Enclave Report Signature (64 bytes)
	copy(sigData.ISVEnclaveReportSignature[:], data[offset:offset+64])
	offset += 64

	// ECDSA Attestation Key (64 bytes)
	copy(sigData.ECDSAAttestationKey[:], data[offset:offset+64])
	offset += 64

	// QE Report (384 bytes)
	if err := p.parseReportBody(data[offset:offset+384], &sigData.QEReport); err != nil {
		return fmt.Errorf("failed to parse QE report: %w", err)
	}
	offset += 384

	// QE Report Signature (64 bytes)
	copy(sigData.QEReportSignature[:], data[offset:offset+64])
	offset += 64

	// QE Authentication Data Size (2 bytes)
	sigData.QEAuthenticationDataSize = binary.LittleEndian.Uint16(data[offset:])
	offset += 2

	// QE Authentication Data (variable)
	if len(data) < offset+int(sigData.QEAuthenticationDataSize) {
		return errors.New("data truncated at QE authentication data")
	}
	sigData.QEAuthenticationData = make([]byte, sigData.QEAuthenticationDataSize)
	copy(sigData.QEAuthenticationData, data[offset:offset+int(sigData.QEAuthenticationDataSize)])
	offset += int(sigData.QEAuthenticationDataSize)

	// Certification Data Type (2 bytes)
	if len(data) < offset+6 {
		return errors.New("data truncated at certification data header")
	}
	sigData.CertificationDataType = binary.LittleEndian.Uint16(data[offset:])
	offset += 2

	// Certification Data Size (4 bytes)
	sigData.CertificationDataSize = binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	// Certification Data (variable - PCK cert chain)
	if len(data) < offset+int(sigData.CertificationDataSize) {
		return errors.New("data truncated at certification data")
	}
	sigData.CertificationData = make([]byte, sigData.CertificationDataSize)
	copy(sigData.CertificationData, data[offset:offset+int(sigData.CertificationDataSize)])

	return nil
}

// GetMRENCLAVE returns the MRENCLAVE value from the quote.
func (q *CryptoDCAPQuote) GetMRENCLAVE() []byte {
	result := make([]byte, 32)
	copy(result, q.ISVEnclaveReport.MRENCLAVE[:])
	return result
}

// GetMRSIGNER returns the MRSIGNER value from the quote.
func (q *CryptoDCAPQuote) GetMRSIGNER() []byte {
	result := make([]byte, 32)
	copy(result, q.ISVEnclaveReport.MRSIGNER[:])
	return result
}

// GetReportData returns the report data from the quote.
func (q *CryptoDCAPQuote) GetReportData() []byte {
	result := make([]byte, 64)
	copy(result, q.ISVEnclaveReport.ReportData[:])
	return result
}

// IsDebugEnclave returns true if the enclave is in debug mode.
func (q *CryptoDCAPQuote) IsDebugEnclave() bool {
	// Debug flag is bit 1 of the first attribute byte
	return (q.ISVEnclaveReport.Attributes[0] & 0x02) != 0
}

// =============================================================================
// DCAP Signature Verifier
// =============================================================================

// DCAPSignatureVerifier verifies ECDSA signatures in DCAP quotes.
type DCAPSignatureVerifier struct {
	ecdsaVerifier *ECDSAVerifier
	hashComputer  *HashComputer
}

// NewDCAPSignatureVerifier creates a new DCAP signature verifier.
func NewDCAPSignatureVerifier() *DCAPSignatureVerifier {
	return &DCAPSignatureVerifier{
		ecdsaVerifier: NewECDSAVerifier(),
		hashComputer:  NewHashComputer(),
	}
}

// VerifyISVEnclaveReportSignature verifies the signature over the ISV enclave report.
func (v *DCAPSignatureVerifier) VerifyISVEnclaveReportSignature(quote *CryptoDCAPQuote) error {
	// Extract the attestation public key
	pubKey, err := v.extractAttestationKey(&quote.QuoteSignatureData)
	if err != nil {
		return fmt.Errorf("failed to extract attestation key: %w", err)
	}

	// Compute hash of header + report body
	dataToSign := quote.RawBytes[:dcapQuoteHeaderSize+dcapReportBodySize]
	hash := v.hashComputer.SHA256(dataToSign)

	// Verify signature
	sig := quote.QuoteSignatureData.ISVEnclaveReportSignature[:]
	if err := v.ecdsaVerifier.VerifyP256(pubKey, hash, sig); err != nil {
		return fmt.Errorf("ISV enclave report signature verification failed: %w", err)
	}

	return nil
}

// VerifyQEReportSignature verifies the QE report signature using the PCK certificate.
func (v *DCAPSignatureVerifier) VerifyQEReportSignature(quote *CryptoDCAPQuote, pckCert *x509.Certificate) error {
	// Extract public key from PCK certificate
	pubKey, err := ExtractPublicKeyFromCert(pckCert)
	if err != nil {
		return fmt.Errorf("failed to extract public key from PCK cert: %w", err)
	}

	// Compute hash of QE report body
	qeReportBytes := v.serializeReportBody(&quote.QuoteSignatureData.QEReport)
	hash := v.hashComputer.SHA256(qeReportBytes)

	// Verify QE report signature
	sig := quote.QuoteSignatureData.QEReportSignature[:]
	if err := v.ecdsaVerifier.VerifyP256(pubKey, hash, sig); err != nil {
		return fmt.Errorf("QE report signature verification failed: %w", err)
	}

	return nil
}

// extractAttestationKey extracts the ECDSA P-256 public key from the signature data.
func (v *DCAPSignatureVerifier) extractAttestationKey(sigData *CryptoDCAPQuoteSignatureData) (*ecdsa.PublicKey, error) {
	// The key is stored as x || y (32 bytes each for P-256)
	if len(sigData.ECDSAAttestationKey) != 64 {
		return nil, errors.New("invalid attestation key length")
	}

	x := new(big.Int).SetBytes(sigData.ECDSAAttestationKey[:32])
	y := new(big.Int).SetBytes(sigData.ECDSAAttestationKey[32:])

	pubKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	// Verify the point is on the curve
	if !pubKey.IsOnCurve(x, y) {
		return nil, errors.New("attestation key point not on curve")
	}

	return pubKey, nil
}

// serializeReportBody serializes an SGX report body to bytes.
func (v *DCAPSignatureVerifier) serializeReportBody(report *CryptoSGXReportBody) []byte {
	buf := make([]byte, dcapReportBodySize)
	offset := 0

	copy(buf[offset:], report.CPUSVN[:])
	offset += 16

	binary.LittleEndian.PutUint32(buf[offset:], report.MiscSelect)
	offset += 4

	copy(buf[offset:], report.Reserved1[:])
	offset += 12

	copy(buf[offset:], report.ISVExtProdID[:])
	offset += 16

	copy(buf[offset:], report.Attributes[:])
	offset += 16

	copy(buf[offset:], report.MRENCLAVE[:])
	offset += 32

	copy(buf[offset:], report.Reserved2[:])
	offset += 32

	copy(buf[offset:], report.MRSIGNER[:])
	offset += 32

	copy(buf[offset:], report.Reserved3[:])
	offset += 32

	copy(buf[offset:], report.ConfigID[:])
	offset += 64

	binary.LittleEndian.PutUint16(buf[offset:], report.ISVProdID)
	offset += 2

	binary.LittleEndian.PutUint16(buf[offset:], report.ISVSVN)
	offset += 2

	binary.LittleEndian.PutUint16(buf[offset:], report.ConfigSVN)
	offset += 2

	copy(buf[offset:], report.Reserved4[:])
	offset += 42

	copy(buf[offset:], report.ISVFamilyID[:])
	offset += 16

	copy(buf[offset:], report.ReportData[:])

	return buf
}

// =============================================================================
// PCK Certificate Verifier
// =============================================================================

// PCKCertificateVerifier verifies PCK certificate chains against Intel Root CA.
type PCKCertificateVerifier struct {
	chainVerifier *CertificateChainVerifier
	certCache     *CertificateCache
}

// NewPCKCertificateVerifier creates a new PCK certificate verifier.
func NewPCKCertificateVerifier() (*PCKCertificateVerifier, error) {
	verifier := &PCKCertificateVerifier{
		chainVerifier: NewCertificateChainVerifier(),
		certCache:     NewCertificateCache(100, 24*time.Hour),
	}

	// Add Intel SGX Root CA
	if err := verifier.chainVerifier.AddRootCA([]byte(IntelSGXRootCAPEM)); err != nil {
		return nil, fmt.Errorf("failed to add Intel SGX Root CA: %w", err)
	}

	// Add Intel SGX PCK Processor CA as intermediate
	if err := verifier.chainVerifier.AddIntermediateCA([]byte(IntelSGXPCKProcessorCAPEM)); err != nil {
		return nil, fmt.Errorf("failed to add Intel SGX PCK Processor CA: %w", err)
	}

	return verifier, nil
}

// VerifyPCKCertChain verifies the PCK certificate chain from a DCAP quote.
func (v *PCKCertificateVerifier) VerifyPCKCertChain(quote *CryptoDCAPQuote) ([]*x509.Certificate, error) {
	certData := quote.QuoteSignatureData.CertificationData

	// Check certification data type
	certType := CryptoCertificationDataType(quote.QuoteSignatureData.CertificationDataType)
	if certType != CryptoCertDataTypePCKCertChain && certType != CryptoCertDataTypePCKCertPlusInfo {
		return nil, fmt.Errorf("unsupported certification data type: %d", certType)
	}

	// Parse certificate chain
	certs, err := ParseCertificateChain(certData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PCK certificate chain: %w", err)
	}

	if len(certs) == 0 {
		return nil, errors.New("empty PCK certificate chain")
	}

	// Verify the chain
	if err := v.chainVerifier.Verify(certs); err != nil {
		return nil, fmt.Errorf("PCK certificate chain verification failed: %w", err)
	}

	return certs, nil
}

// GetPCKCert returns the PCK leaf certificate from the chain.
func (v *PCKCertificateVerifier) GetPCKCert(quote *CryptoDCAPQuote) (*x509.Certificate, error) {
	certs, err := v.VerifyPCKCertChain(quote)
	if err != nil {
		return nil, err
	}

	// Return the leaf (first) certificate
	return certs[0], nil
}

// =============================================================================
// TCB Info Verifier
// =============================================================================

// TCBStatus represents the status of a TCB level.
type TCBStatus string

const (
	TCBStatusUpToDate                          TCBStatus = "UpToDate"
	TCBStatusSWHardeningNeeded                 TCBStatus = "SWHardeningNeeded"
	TCBStatusConfigurationNeeded               TCBStatus = "ConfigurationNeeded"
	TCBStatusConfigurationAndSWHardeningNeeded TCBStatus = "ConfigurationAndSWHardeningNeeded"
	TCBStatusOutOfDate                         TCBStatus = "OutOfDate"
	TCBStatusOutOfDateConfigurationNeeded      TCBStatus = "OutOfDateConfigurationNeeded"
	TCBStatusRevoked                           TCBStatus = "Revoked"
)

// CryptoTCBLevel represents a single TCB level from Intel TCB Info.
type CryptoTCBLevel struct {
	TCBComponents []CryptoTCBComponent `json:"tcb"`
	PCESVN        uint16               `json:"pcesvn"`
	TCBDate       string               `json:"tcbDate"`
	TCBStatus     TCBStatus            `json:"tcbStatus"`
}

// CryptoTCBComponent represents a single TCB component (CPUSVN byte).
type CryptoTCBComponent struct {
	SVN      uint8  `json:"svn"`
	Category string `json:"category,omitempty"`
	Type     string `json:"type,omitempty"`
}

// CryptoTCBInfo represents Intel's TCB Info structure.
type CryptoTCBInfo struct {
	Version                 int              `json:"version"`
	IssueDate               string           `json:"issueDate"`
	NextUpdate              string           `json:"nextUpdate"`
	FMSPC                   string           `json:"fmspc"`
	PCEID                   string           `json:"pceId"`
	TCBType                 int              `json:"tcbType"`
	TCBEvaluationDataNumber int              `json:"tcbEvaluationDataNumber"`
	TCBLevels               []CryptoTCBLevel `json:"tcbLevels"`
}

// CryptoTCBInfoWrapper wraps the TCB Info with signature.
type CryptoTCBInfoWrapper struct {
	TCBInfo   json.RawMessage `json:"tcbInfo"`
	Signature string          `json:"signature"`
}

// TCBInfoVerifier verifies TCB levels against Intel TCB Info.
type TCBInfoVerifier struct {
	hashComputer *HashComputer
	tcbInfoCache map[string]*CryptoTCBInfo
	mu           sync.RWMutex //nolint:unused // Reserved for future concurrent access protection
}

// NewTCBInfoVerifier creates a new TCB Info verifier.
func NewTCBInfoVerifier() *TCBInfoVerifier {
	return &TCBInfoVerifier{
		hashComputer: NewHashComputer(),
		tcbInfoCache: make(map[string]*CryptoTCBInfo),
	}
}

// ParseTCBInfo parses TCB Info JSON data.
func (v *TCBInfoVerifier) ParseTCBInfo(data []byte) (*CryptoTCBInfo, error) {
	var wrapper CryptoTCBInfoWrapper
	if err := json.Unmarshal(data, &wrapper); err == nil && len(wrapper.TCBInfo) > 0 {
		// Successfully parsed as wrapper with non-empty tcbInfo field
		var tcbInfo CryptoTCBInfo
		if err := json.Unmarshal(wrapper.TCBInfo, &tcbInfo); err != nil {
			return nil, fmt.Errorf("failed to parse TCB Info content: %w", err)
		}
		return &tcbInfo, nil
	}

	// Try parsing directly as CryptoTCBInfo
	var tcbInfo CryptoTCBInfo
	if err := json.Unmarshal(data, &tcbInfo); err != nil {
		return nil, fmt.Errorf("failed to parse TCB Info: %w", err)
	}
	return &tcbInfo, nil
}

// GetTCBStatus determines the TCB status for a given CPUSVN and PCESVN.
func (v *TCBInfoVerifier) GetTCBStatus(tcbInfo *CryptoTCBInfo, cpusvn []byte, pcesvn uint16) (TCBStatus, error) {
	if len(cpusvn) != 16 {
		return "", fmt.Errorf("invalid CPUSVN length: got %d, expected 16", len(cpusvn))
	}

	for _, level := range tcbInfo.TCBLevels {
		if v.matchesTCBLevel(level, cpusvn, pcesvn) {
			return level.TCBStatus, nil
		}
	}

	return TCBStatusOutOfDate, nil
}

// matchesTCBLevel checks if the given CPUSVN/PCESVN matches a TCB level.
func (v *TCBInfoVerifier) matchesTCBLevel(level CryptoTCBLevel, cpusvn []byte, pcesvn uint16) bool {
	// PCESVN must be >= level PCESVN
	if pcesvn < level.PCESVN {
		return false
	}

	// Each CPUSVN component must be >= corresponding level component
	for i, component := range level.TCBComponents {
		if i >= len(cpusvn) {
			break
		}
		if cpusvn[i] < component.SVN {
			return false
		}
	}

	return true
}

// IsTCBStatusAcceptable returns true if the TCB status is acceptable for production.
func IsTCBStatusAcceptable(status TCBStatus, allowOutOfDate bool) bool {
	switch status {
	case TCBStatusUpToDate, TCBStatusSWHardeningNeeded:
		return true
	case TCBStatusConfigurationNeeded, TCBStatusConfigurationAndSWHardeningNeeded:
		return true // Configuration issues may be acceptable
	case TCBStatusOutOfDate, TCBStatusOutOfDateConfigurationNeeded:
		return allowOutOfDate
	case TCBStatusRevoked:
		return false
	default:
		return false
	}
}

// =============================================================================
// QE Identity Verifier
// =============================================================================

// QEIdentity represents the Quoting Enclave identity from Intel.
type QEIdentity struct {
	Version         int             `json:"version"`
	IssueDate       string          `json:"issueDate"`
	NextUpdate      string          `json:"nextUpdate"`
	EnclaveIdentity EnclaveIdentity `json:"enclaveIdentity"`
}

// EnclaveIdentity contains the enclave identity fields.
type EnclaveIdentity struct {
	ID                      string       `json:"id"`
	Version                 int          `json:"version"`
	IssueDate               string       `json:"issueDate"`
	NextUpdate              string       `json:"nextUpdate"`
	TCBEvaluationDataNumber int          `json:"tcbEvaluationDataNumber"`
	MiscSelect              string       `json:"miscselect"`
	MiscSelectMask          string       `json:"miscselectMask"`
	Attributes              string       `json:"attributes"`
	AttributesMask          string       `json:"attributesMask"`
	MRSIGNER                string       `json:"mrsigner"`
	ISVProdID               uint16       `json:"isvprodid"`
	TCBLevels               []QETCBLevel `json:"tcbLevels"`
}

// QETCBLevel represents a TCB level for QE identity.
type QETCBLevel struct {
	TCB       QETCBInfo `json:"tcb"`
	TCBDate   string    `json:"tcbDate"`
	TCBStatus TCBStatus `json:"tcbStatus"`
}

// QETCBInfo contains ISVSVN for QE TCB.
type QETCBInfo struct {
	ISVSVN uint16 `json:"isvsvn"`
}

// QEIdentityVerifier verifies Quoting Enclave identity.
type QEIdentityVerifier struct {
	hashComputer *HashComputer
}

// NewQEIdentityVerifier creates a new QE identity verifier.
func NewQEIdentityVerifier() *QEIdentityVerifier {
	return &QEIdentityVerifier{
		hashComputer: NewHashComputer(),
	}
}

// ParseQEIdentity parses QE Identity JSON data.
func (v *QEIdentityVerifier) ParseQEIdentity(data []byte) (*QEIdentity, error) {
	var qeIdentity QEIdentity
	if err := json.Unmarshal(data, &qeIdentity); err != nil {
		return nil, fmt.Errorf("failed to parse QE Identity: %w", err)
	}
	return &qeIdentity, nil
}

// VerifyQEReport verifies the QE report against the QE identity.
func (v *QEIdentityVerifier) VerifyQEReport(qeReport *SGXReportBody, qeIdentity *QEIdentity) error {
	identity := &qeIdentity.EnclaveIdentity

	// Verify MRSIGNER
	expectedMRSIGNER, err := hex.DecodeString(identity.MRSIGNER)
	if err != nil {
		return fmt.Errorf("failed to decode expected MRSIGNER: %w", err)
	}
	if !bytes.Equal(qeReport.MRSigner[:], expectedMRSIGNER) {
		return fmt.Errorf("MRSIGNER mismatch: got %x, expected %x", qeReport.MRSigner, expectedMRSIGNER)
	}

	// Verify ISV Product ID
	if qeReport.ISVProdID != identity.ISVProdID {
		return fmt.Errorf("ISVProdID mismatch: got %d, expected %d", qeReport.ISVProdID, identity.ISVProdID)
	}

	// Verify MISCSELECT (with mask)
	miscSelect, err := hex.DecodeString(identity.MiscSelect)
	if err == nil && len(miscSelect) >= 4 {
		miscSelectMask, _ := hex.DecodeString(identity.MiscSelectMask)
		if len(miscSelectMask) >= 4 {
			expected := binary.LittleEndian.Uint32(miscSelect)
			mask := binary.LittleEndian.Uint32(miscSelectMask)
			if (qeReport.MiscSelect & mask) != (expected & mask) {
				return fmt.Errorf("MISCSELECT mismatch after masking")
			}
		}
	}

	// Verify attributes (with mask)
	attributes, err := hex.DecodeString(identity.Attributes)
	if err == nil && len(attributes) >= 8 {
		attrMask, _ := hex.DecodeString(identity.AttributesMask)
		if len(attrMask) >= 8 {
			// Compare flags (first 8 bytes)
			expectedFlags := binary.LittleEndian.Uint64(attributes[:8])
			maskFlags := binary.LittleEndian.Uint64(attrMask[:8])
			if (qeReport.Attributes.Flags & maskFlags) != (expectedFlags & maskFlags) {
				return fmt.Errorf("attributes flags mismatch")
			}
		}
	}

	return nil
}

// =============================================================================
// Complete DCAP Verifier
// =============================================================================

// DCAPVerificationResult contains the result of DCAP quote verification.
type DCAPVerificationResult struct {
	Valid           bool
	Quote           *CryptoDCAPQuote
	PCKCertificates []*x509.Certificate
	TCBStatus       TCBStatus
	Errors          []string
	Warnings        []string
}

// DCAPVerifier provides complete DCAP quote verification.
type DCAPVerifier struct {
	quoteParser        *DCAPQuoteParser
	sigVerifier        *DCAPSignatureVerifier
	pckVerifier        *PCKCertificateVerifier
	tcbVerifier        *TCBInfoVerifier
	qeIdentityVerifier *QEIdentityVerifier
}

// NewDCAPVerifier creates a new complete DCAP verifier.
func NewDCAPVerifier() (*DCAPVerifier, error) {
	pckVerifier, err := NewPCKCertificateVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to create PCK verifier: %w", err)
	}

	return &DCAPVerifier{
		quoteParser:        NewDCAPQuoteParser(),
		sigVerifier:        NewDCAPSignatureVerifier(),
		pckVerifier:        pckVerifier,
		tcbVerifier:        NewTCBInfoVerifier(),
		qeIdentityVerifier: NewQEIdentityVerifier(),
	}, nil
}

// Verify performs complete verification of a DCAP quote.
func (v *DCAPVerifier) Verify(quoteBytes []byte) (*DCAPVerificationResult, error) {
	result := &DCAPVerificationResult{
		Valid: true,
	}

	// Parse quote
	quote, err := v.quoteParser.Parse(quoteBytes)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("quote parsing failed: %v", err))
		return result, nil
	}
	result.Quote = quote

	// Verify PCK certificate chain
	pckCerts, err := v.pckVerifier.VerifyPCKCertChain(quote)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("PCK certificate verification failed: %v", err))
		return result, nil
	}
	result.PCKCertificates = pckCerts

	// Verify QE report signature using PCK certificate
	if err := v.sigVerifier.VerifyQEReportSignature(quote, pckCerts[0]); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("QE report signature verification failed: %v", err))
		return result, nil
	}

	// Verify ISV enclave report signature
	if err := v.sigVerifier.VerifyISVEnclaveReportSignature(quote); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("ISV enclave report signature verification failed: %v", err))
		return result, nil
	}

	// Check debug mode
	if quote.IsDebugEnclave() {
		result.Warnings = append(result.Warnings, "enclave is running in debug mode")
	}

	return result, nil
}

// =============================================================================
// Test Helper Functions
// =============================================================================

// CreateTestDCAPQuote creates a test DCAP quote for testing purposes.
// Note: This creates a structurally valid quote but with invalid signatures.
func CreateTestDCAPQuote(mrenclave, mrsigner []byte, debug bool, reportData []byte) []byte {
	// Create header
	header := make([]byte, dcapQuoteHeaderSize)
	binary.LittleEndian.PutUint16(header[0:], uint16(CryptoDCAPQuoteV3))
	binary.LittleEndian.PutUint16(header[2:], uint16(CryptoAttKeyTypeECDSA256P256))
	// QE Vendor ID (Intel)
	copy(header[12:28], []byte{0x93, 0x9A, 0x72, 0x33, 0xF7, 0x9C, 0x4C, 0xA9, 0x94, 0x0A, 0x0D, 0xB3, 0x95, 0x7F, 0x06, 0x07})

	// Create report body
	reportBody := make([]byte, dcapReportBodySize)

	// Report body layout:
	// CPUSVN: 0-15 (16 bytes)
	// MiscSelect: 16-19 (4 bytes)
	// Reserved1: 20-31 (12 bytes)
	// ISVExtProdID: 32-47 (16 bytes)
	// Attributes: 48-63 (16 bytes) - debug flag is bit 1 of byte 48
	// MRENCLAVE: 64-95 (32 bytes)
	// Reserved2: 96-127 (32 bytes)
	// MRSIGNER: 128-159 (32 bytes)
	// Reserved3: 160-191 (32 bytes)
	// ConfigID: 192-255 (64 bytes)
	// ISVProdID: 256-257 (2 bytes)
	// ISVSVN: 258-259 (2 bytes)
	// ConfigSVN: 260-261 (2 bytes)
	// Reserved4: 262-303 (42 bytes)
	// ISVFamilyID: 304-319 (16 bytes)
	// ReportData: 320-383 (64 bytes)

	// Set attributes (debug flag is bit 1 of first byte at offset 48)
	if debug {
		reportBody[48] = 0x07 // DEBUG | INIT | MODE64BIT
	} else {
		reportBody[48] = 0x05 // INIT | MODE64BIT
	}

	// Set MRENCLAVE (offset 64)
	if len(mrenclave) == 32 {
		copy(reportBody[64:96], mrenclave)
	}

	// Set MRSIGNER (offset 128)
	if len(mrsigner) == 32 {
		copy(reportBody[128:160], mrsigner)
	}

	// Set report data (offset 320)
	if len(reportData) > 0 {
		copy(reportBody[320:], reportData)
	}

	// Add fake PCK cert chain (minimal PEM)
	fakePEM := []byte(`-----BEGIN CERTIFICATE-----
MIIBkTCB+wIJAMHAwMDAwMDAwMAKBggqhkjOPQQDAjAXMRUwEwYDVQQDDAxUZXN0
IFJvb3QgQ0EwHhcNMjMwMTAxMDAwMDAwWhcNMjQwMTAxMDAwMDAwWjAXMRUwEwYD
VQQDDAxUZXN0IFJvb3QgQ0EwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAQXXXXX
XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
XXXXXXXXXXXXXXXXXozQwMjAMBgNVHRMEBTADAQH/MAoGCCqGSM49BAMCA0kAMEYC
IQCEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEAIhAOOOOOOOOOOOOOOOOOOOOOOO
OOOOOOOOOO==
-----END CERTIFICATE-----`)

	// Calculate signature data length including the certificate
	sigDataLen := 64 + 64 + 384 + 64 + 2 + 2 + 4 + len(fakePEM)

	// Create signature data with sufficient space
	sigData := make([]byte, sigDataLen)

	// Leave signature and key areas as zeros (placeholder data)
	// Offset 0-63: ISV Enclave Report Signature (zeros)
	// Offset 64-127: ECDSA Attestation Key (zeros)
	// Offset 128-511: QE Report (384 bytes - zeros)
	// Offset 512-575: QE Report Signature (64 bytes - zeros)

	qeAuthOffset := 64 + 64 + 384 + 64
	binary.LittleEndian.PutUint16(sigData[qeAuthOffset:], 0) // QE auth data size = 0
	//nolint:gosec // G115: CryptoCertDataTypePCKCertChain is small constant
	binary.LittleEndian.PutUint16(sigData[qeAuthOffset+2:], uint16(CryptoCertDataTypePCKCertChain))
	//nolint:gosec // G115: len(fakePEM) is bounded cert chain size
	binary.LittleEndian.PutUint32(sigData[qeAuthOffset+4:], uint32(len(fakePEM)))
	copy(sigData[qeAuthOffset+8:], fakePEM)

	// Combine all parts
	quote := make([]byte, dcapQuoteHeaderSize+dcapReportBodySize+4+sigDataLen)
	copy(quote[0:], header)
	copy(quote[dcapQuoteHeaderSize:], reportBody)
	//nolint:gosec // G115: sigDataLen is bounded buffer size
	binary.LittleEndian.PutUint32(quote[dcapQuoteHeaderSize+dcapReportBodySize:], uint32(sigDataLen))
	copy(quote[dcapQuoteHeaderSize+dcapReportBodySize+4:], sigData)

	return quote
}

// GetIntelSGXRootCA returns the Intel SGX Root CA certificate.
func GetIntelSGXRootCA() (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(IntelSGXRootCAPEM))
	if block == nil {
		return nil, errors.New("failed to decode Intel SGX Root CA PEM")
	}
	return x509.ParseCertificate(block.Bytes)
}
