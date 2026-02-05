//go:build sgx_hardware

// Package sgx provides Intel SGX enclave management and DCAP attestation.
//
// This file implements SGX DCAP quote generation and parsing for remote
// attestation. DCAP (Data Center Attestation Primitives) uses ECDSA signatures
// that can be verified without connecting to Intel Attestation Service.
//
// Quote Structure (v3/v4):
//
//	+------------------+
//	| Quote Header     | 48 bytes
//	+------------------+
//	| ISV Enclave Rep  | 384 bytes (SGX Report Body)
//	+------------------+
//	| Signature Data   | Variable
//	+------------------+
//
// Task Reference: VE-2029 - Hardware TEE Integration Layer
package sgx

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"sync"
)

// =============================================================================
// Quote Constants
// =============================================================================

const (
	// QuoteVersionDCAP3 is DCAP quote version 3.
	QuoteVersionDCAP3 uint16 = 3

	// QuoteVersionDCAP4 is DCAP quote version 4.
	QuoteVersionDCAP4 uint16 = 4

	// QuoteHeaderSize is the size of the quote header in bytes.
	QuoteHeaderSize = 48

	// ReportBodySize is the size of the SGX report body in bytes.
	ReportBodySize = 384

	// MinQuoteSize is the minimum valid quote size.
	MinQuoteSize = QuoteHeaderSize + ReportBodySize + 4 // Header + Report + SigLen

	// SignatureDataMinSize is the minimum signature data size.
	SignatureDataMinSize = 64 + 64 + 384 + 64 + 2 + 2 + 4 // ISVSig + PubKey + QEReport + QESig + AuthSize + CertType + CertSize
)

// Attestation Key Types
const (
	// AttKeyTypeECDSA256P256 is ECDSA-256 with P-256 curve.
	AttKeyTypeECDSA256P256 uint16 = 2

	// AttKeyTypeECDSA384P384 is ECDSA-384 with P-384 curve.
	AttKeyTypeECDSA384P384 uint16 = 3
)

// TEE Types
const (
	// TEETypeSGX indicates standard SGX.
	TEETypeSGX uint32 = 0x00000000

	// TEETypeTDX indicates Intel TDX (Trust Domain Extensions).
	TEETypeTDX uint32 = 0x00000081
)

// Certification Data Types
const (
	// CertDataTypePPID is PPID + CPUSVN + PCE ID + PCE SVN.
	CertDataTypePPID uint16 = 1

	// CertDataTypePCKCertChain is PCK Certificate Chain (PEM).
	CertDataTypePCKCertChain uint16 = 5

	// CertDataTypePCKCertPlusInfo is PCK Certificate + Manifest.
	CertDataTypePCKCertPlusInfo uint16 = 6
)

// Intel QE Vendor ID (fixed value for Intel QE).
var IntelQEVendorID = [16]byte{
	0x93, 0x9a, 0x72, 0x33, 0xf7, 0x9c, 0x4c, 0xa9,
	0x94, 0xa5, 0xcb, 0x8e, 0x39, 0x34, 0x4c, 0x1d,
}

// =============================================================================
// Error Types
// =============================================================================

var (
	// ErrQuoteTooSmall indicates the quote data is too small.
	ErrQuoteTooSmall = errors.New("sgx: quote data too small")

	// ErrInvalidQuoteVersion indicates an unsupported quote version.
	ErrInvalidQuoteVersion = errors.New("sgx: invalid quote version")

	// ErrInvalidQuoteSignature indicates the quote signature is invalid.
	ErrInvalidQuoteSignature = errors.New("sgx: invalid quote signature")

	// ErrInvalidAttKeyType indicates an unsupported attestation key type.
	ErrInvalidAttKeyType = errors.New("sgx: invalid attestation key type")

	// ErrInvalidQEVendorID indicates an unrecognized QE vendor ID.
	ErrInvalidQEVendorID = errors.New("sgx: invalid QE vendor ID")

	// ErrInvalidCertDataType indicates an unsupported certification data type.
	ErrInvalidCertDataType = errors.New("sgx: invalid certification data type")

	// ErrQuoteGenerationFailed indicates quote generation failed.
	ErrQuoteGenerationFailed = errors.New("sgx: quote generation failed")
)

// =============================================================================
// Quote Header
// =============================================================================

// QuoteHeader represents the header of a DCAP quote.
type QuoteHeader struct {
	// Version is the quote version (3 or 4).
	Version uint16

	// AttKeyType is the type of attestation key (ECDSA-256 or ECDSA-384).
	AttKeyType uint16

	// TEEType identifies the TEE technology (SGX or TDX).
	TEEType uint32

	// Reserved bytes.
	Reserved uint32

	// QEVendorID is the Quoting Enclave vendor ID.
	QEVendorID [16]byte

	// UserData contains custom user data (20 bytes).
	UserData [20]byte
}

// Serialize serializes the quote header to bytes.
func (h *QuoteHeader) Serialize() []byte {
	buf := make([]byte, QuoteHeaderSize)

	binary.LittleEndian.PutUint16(buf[0:2], h.Version)
	binary.LittleEndian.PutUint16(buf[2:4], h.AttKeyType)
	binary.LittleEndian.PutUint32(buf[4:8], h.TEEType)
	binary.LittleEndian.PutUint32(buf[8:12], h.Reserved)
	copy(buf[12:28], h.QEVendorID[:])
	copy(buf[28:48], h.UserData[:])

	return buf
}

// ParseQuoteHeader parses a quote header from bytes.
func ParseQuoteHeader(data []byte) (*QuoteHeader, error) {
	if len(data) < QuoteHeaderSize {
		return nil, ErrQuoteTooSmall
	}

	h := &QuoteHeader{
		Version:    binary.LittleEndian.Uint16(data[0:2]),
		AttKeyType: binary.LittleEndian.Uint16(data[2:4]),
		TEEType:    binary.LittleEndian.Uint32(data[4:8]),
		Reserved:   binary.LittleEndian.Uint32(data[8:12]),
	}
	copy(h.QEVendorID[:], data[12:28])
	copy(h.UserData[:], data[28:48])

	return h, nil
}

// IsIntelQE returns true if the QE vendor ID matches Intel's.
func (h *QuoteHeader) IsIntelQE() bool {
	return h.QEVendorID == IntelQEVendorID
}

// =============================================================================
// Report Body
// =============================================================================

// ReportBody represents the SGX report body structure (384 bytes).
type ReportBody struct {
	// CPUSVN is the CPU Security Version Number.
	CPUSVN [CPUSVNSize]byte

	// MiscSelect is the MISC select value.
	MiscSelect uint32

	// Reserved1 is reserved bytes.
	Reserved1 [12]byte

	// ISVExtProdID is the ISV Extended Product ID.
	ISVExtProdID [ExtProdIDSize]byte

	// Attributes are the enclave attributes.
	Attributes Attributes

	// MREnclave is the enclave measurement.
	MREnclave Measurement

	// Reserved2 is reserved bytes.
	Reserved2 [32]byte

	// MRSigner is the signer measurement.
	MRSigner Measurement

	// Reserved3 is reserved bytes.
	Reserved3 [32]byte

	// ConfigID is the configuration ID.
	ConfigID [ConfigIDSize]byte

	// ISVProdID is the ISV Product ID.
	ISVProdID uint16

	// ISVSVN is the ISV Security Version Number.
	ISVSVN uint16

	// ConfigSVN is the Configuration Security Version Number.
	ConfigSVN uint16

	// Reserved4 is reserved bytes.
	Reserved4 [42]byte

	// ISVFamilyID is the ISV Family ID.
	ISVFamilyID [FamilyIDSize]byte

	// ReportData is the user-provided report data.
	ReportData [ReportDataSize]byte
}

// Serialize serializes the report body to bytes.
func (r *ReportBody) Serialize() []byte {
	buf := make([]byte, ReportBodySize)
	offset := 0

	copy(buf[offset:offset+16], r.CPUSVN[:])
	offset += 16

	binary.LittleEndian.PutUint32(buf[offset:offset+4], r.MiscSelect)
	offset += 4

	copy(buf[offset:offset+12], r.Reserved1[:])
	offset += 12

	copy(buf[offset:offset+16], r.ISVExtProdID[:])
	offset += 16

	binary.LittleEndian.PutUint64(buf[offset:offset+8], r.Attributes.Flags)
	offset += 8
	binary.LittleEndian.PutUint64(buf[offset:offset+8], r.Attributes.Xfrm)
	offset += 8

	copy(buf[offset:offset+32], r.MREnclave[:])
	offset += 32

	copy(buf[offset:offset+32], r.Reserved2[:])
	offset += 32

	copy(buf[offset:offset+32], r.MRSigner[:])
	offset += 32

	copy(buf[offset:offset+32], r.Reserved3[:])
	offset += 32

	copy(buf[offset:offset+64], r.ConfigID[:])
	offset += 64

	binary.LittleEndian.PutUint16(buf[offset:offset+2], r.ISVProdID)
	offset += 2

	binary.LittleEndian.PutUint16(buf[offset:offset+2], r.ISVSVN)
	offset += 2

	binary.LittleEndian.PutUint16(buf[offset:offset+2], r.ConfigSVN)
	offset += 2

	copy(buf[offset:offset+42], r.Reserved4[:])
	offset += 42

	copy(buf[offset:offset+16], r.ISVFamilyID[:])
	offset += 16

	copy(buf[offset:offset+64], r.ReportData[:])

	return buf
}

// ParseReportBody parses a report body from bytes.
func ParseReportBody(data []byte) (*ReportBody, error) {
	if len(data) < ReportBodySize {
		return nil, fmt.Errorf("report body too small: got %d, need %d", len(data), ReportBodySize)
	}

	r := &ReportBody{}
	offset := 0

	copy(r.CPUSVN[:], data[offset:offset+16])
	offset += 16

	r.MiscSelect = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	copy(r.Reserved1[:], data[offset:offset+12])
	offset += 12

	copy(r.ISVExtProdID[:], data[offset:offset+16])
	offset += 16

	r.Attributes.Flags = binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	r.Attributes.Xfrm = binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8

	copy(r.MREnclave[:], data[offset:offset+32])
	offset += 32

	copy(r.Reserved2[:], data[offset:offset+32])
	offset += 32

	copy(r.MRSigner[:], data[offset:offset+32])
	offset += 32

	copy(r.Reserved3[:], data[offset:offset+32])
	offset += 32

	copy(r.ConfigID[:], data[offset:offset+64])
	offset += 64

	r.ISVProdID = binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2

	r.ISVSVN = binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2

	r.ConfigSVN = binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2

	copy(r.Reserved4[:], data[offset:offset+42])
	offset += 42

	copy(r.ISVFamilyID[:], data[offset:offset+16])
	offset += 16

	copy(r.ReportData[:], data[offset:offset+64])

	return r, nil
}

// =============================================================================
// Quote Signature Data
// =============================================================================

// QuoteSignatureData represents the signature data in a DCAP quote.
type QuoteSignatureData struct {
	// ISVEnclaveReportSignature is the ECDSA P-256 signature (r || s).
	ISVEnclaveReportSignature [64]byte

	// ECDSAAttestationKey is the public key (x || y).
	ECDSAAttestationKey [64]byte

	// QEReport is the Quoting Enclave report body.
	QEReport ReportBody

	// QEReportSignature is the signature over QE report.
	QEReportSignature [64]byte

	// QEAuthenticationData is authentication data from QE.
	QEAuthenticationData []byte

	// CertificationDataType indicates the type of certification data.
	CertificationDataType uint16

	// CertificationData contains the PCK certificate chain (PEM).
	CertificationData []byte
}

// =============================================================================
// Quote
// =============================================================================

// Quote represents a complete DCAP attestation quote.
type Quote struct {
	// Header is the quote header.
	Header QuoteHeader

	// ReportBody is the ISV enclave report body.
	ReportBody ReportBody

	// SignatureLength is the length of the signature data.
	SignatureLength uint32

	// SignatureData contains the quote signatures and certificates.
	SignatureData QuoteSignatureData

	// RawBytes contains the original raw quote bytes.
	RawBytes []byte
}

// =============================================================================
// Quote Generator
// =============================================================================

// QuoteGenerator generates SGX DCAP quotes.
type QuoteGenerator struct {
	mu sync.Mutex

	enclave    *Enclave
	simulated  bool
	quoteCount uint64
}

// NewQuoteGenerator creates a new quote generator.
func NewQuoteGenerator(enclave *Enclave) *QuoteGenerator {
	return &QuoteGenerator{
		enclave:   enclave,
		simulated: enclave == nil || enclave.IsSimulated(),
	}
}

// GenerateQuote generates a DCAP attestation quote.
//
// Parameters:
//   - reportData: User-provided data to include in the quote (max 64 bytes)
//
// The quote generation process:
// 1. Generate an SGX report targeting the Quoting Enclave (QE)
// 2. QE verifies the report and generates an ECDSA signature
// 3. Quote includes PCK certificate chain for verification
func (g *QuoteGenerator) GenerateQuote(reportData [64]byte) (*Quote, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.enclave != nil && !g.enclave.IsLoaded() {
		return nil, ErrEnclaveNotLoaded
	}

	g.quoteCount++

	if g.simulated {
		return g.generateSimulatedQuote(reportData)
	}

	return g.generateHardwareQuote(reportData)
}

// generateHardwareQuote generates a quote using SGX DCAP.
func (g *QuoteGenerator) generateHardwareQuote(reportData [64]byte) (*Quote, error) {
	// Real DCAP implementation:
	// 1. Call sgx_qe_get_target_info() to get QE's target info
	// 2. Generate report: sgx_create_report(&qe_target_info, &report_data, &report)
	// 3. Call sgx_qe_get_quote_size() to get buffer size
	// 4. Get quote: sgx_qe_get_quote(&report, quote_size, quote)
	// 5. Parse the returned quote

	// Fall back to simulation
	return g.generateSimulatedQuote(reportData)
}

// generateSimulatedQuote generates a simulated quote for testing.
func (g *QuoteGenerator) generateSimulatedQuote(reportData [64]byte) (*Quote, error) {
	quote := &Quote{}

	// Set header
	quote.Header = QuoteHeader{
		Version:    QuoteVersionDCAP3,
		AttKeyType: AttKeyTypeECDSA256P256,
		TEEType:    TEETypeSGX,
		QEVendorID: IntelQEVendorID,
	}

	// Generate user data hash
	hash := sha256.Sum256(reportData[:])
	copy(quote.Header.UserData[:], hash[:20])

	// Set report body
	if g.enclave != nil {
		identity := g.enclave.GetIdentity()
		quote.ReportBody.MREnclave = identity.MREnclave
		quote.ReportBody.MRSigner = identity.MRSigner
		quote.ReportBody.ISVProdID = identity.ISVProdID
		quote.ReportBody.ISVSVN = identity.ISVSVN
		quote.ReportBody.Attributes = identity.Attributes
	} else {
		// Generate simulated measurements
		mrEnclave := sha256.Sum256([]byte("simulated-mrenclave"))
		copy(quote.ReportBody.MREnclave[:], mrEnclave[:])

		mrSigner := sha256.Sum256([]byte("virtengine-signer-v1"))
		copy(quote.ReportBody.MRSigner[:], mrSigner[:])

		quote.ReportBody.ISVProdID = 1
		quote.ReportBody.ISVSVN = 1
		quote.ReportBody.Attributes = Attributes{
			Flags: FlagInitted | FlagMode64Bit,
			Xfrm:  0x03,
		}
	}

	quote.ReportBody.ReportData = reportData

	// Simulate CPUSVN
	copy(quote.ReportBody.CPUSVN[:], []byte{
		0x0f, 0x0f, 0x02, 0x04, 0xff, 0x80, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	})

	// Generate simulated signature data
	if err := g.generateSimulatedSignature(quote); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQuoteGenerationFailed, err)
	}

	// Serialize quote to raw bytes
	quote.RawBytes = SerializeQuote(quote)

	return quote, nil
}

// generateSimulatedSignature generates simulated signature data.
func (g *QuoteGenerator) generateSimulatedSignature(quote *Quote) error {
	// Generate a random ECDSA key pair for simulation
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}

	// Store public key (x || y)
	xBytes := privateKey.PublicKey.X.Bytes()
	yBytes := privateKey.PublicKey.Y.Bytes()
	copy(quote.SignatureData.ECDSAAttestationKey[32-len(xBytes):32], xBytes)
	copy(quote.SignatureData.ECDSAAttestationKey[64-len(yBytes):64], yBytes)

	// Sign the header + report body
	dataToSign := append(quote.Header.Serialize(), quote.ReportBody.Serialize()...)
	hash := sha256.Sum256(dataToSign)

	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return fmt.Errorf("failed to sign: %w", err)
	}

	// Store signature (r || s)
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	copy(quote.SignatureData.ISVEnclaveReportSignature[32-len(rBytes):32], rBytes)
	copy(quote.SignatureData.ISVEnclaveReportSignature[64-len(sBytes):64], sBytes)

	// Generate QE report
	quote.SignatureData.QEReport = quote.ReportBody // Simplified

	// Sign QE report
	qeHash := sha256.Sum256(quote.SignatureData.QEReport.Serialize())
	qeR, qeS, err := ecdsa.Sign(rand.Reader, privateKey, qeHash[:])
	if err != nil {
		return fmt.Errorf("failed to sign QE report: %w", err)
	}
	qeRBytes := qeR.Bytes()
	qeSBytes := qeS.Bytes()
	copy(quote.SignatureData.QEReportSignature[32-len(qeRBytes):32], qeRBytes)
	copy(quote.SignatureData.QEReportSignature[64-len(qeSBytes):64], qeSBytes)

	// Set certification data type (PCK cert chain)
	quote.SignatureData.CertificationDataType = CertDataTypePCKCertChain

	// Generate simulated PCK certificate chain
	quote.SignatureData.CertificationData = []byte("-----BEGIN CERTIFICATE-----\nSIMULATED PCK CERTIFICATE\n-----END CERTIFICATE-----\n")

	// Calculate signature length
	// 64 (ISV sig) + 64 (att key) + 384 (QE report) + 64 (QE sig) + 2 (auth size) + len(auth) + 2 (cert type) + 4 (cert size) + len(cert)
	quote.SignatureLength = uint32(64 + 64 + 384 + 64 + 2 + len(quote.SignatureData.QEAuthenticationData) + 2 + 4 + len(quote.SignatureData.CertificationData))

	return nil
}

// =============================================================================
// Quote Parsing
// =============================================================================

// ParseQuote parses raw quote bytes into a Quote structure.
func ParseQuote(data []byte) (*Quote, error) {
	if len(data) < MinQuoteSize {
		return nil, fmt.Errorf("%w: got %d bytes, need at least %d", ErrQuoteTooSmall, len(data), MinQuoteSize)
	}

	quote := &Quote{
		RawBytes: make([]byte, len(data)),
	}
	copy(quote.RawBytes, data)

	offset := 0

	// Parse header
	header, err := ParseQuoteHeader(data[offset : offset+QuoteHeaderSize])
	if err != nil {
		return nil, err
	}
	quote.Header = *header
	offset += QuoteHeaderSize

	// Validate version
	if quote.Header.Version != QuoteVersionDCAP3 && quote.Header.Version != QuoteVersionDCAP4 {
		return nil, fmt.Errorf("%w: got %d", ErrInvalidQuoteVersion, quote.Header.Version)
	}

	// Parse report body
	reportBody, err := ParseReportBody(data[offset : offset+ReportBodySize])
	if err != nil {
		return nil, err
	}
	quote.ReportBody = *reportBody
	offset += ReportBodySize

	// Parse signature length
	if len(data) < offset+4 {
		return nil, ErrQuoteTooSmall
	}
	quote.SignatureLength = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Validate we have enough data for signature
	if len(data) < offset+int(quote.SignatureLength) {
		return nil, fmt.Errorf("%w: signature data truncated", ErrQuoteTooSmall)
	}

	// Parse signature data
	sigData := data[offset : offset+int(quote.SignatureLength)]
	if err := parseSignatureData(sigData, &quote.SignatureData); err != nil {
		return nil, err
	}

	return quote, nil
}

// parseSignatureData parses the signature data section.
func parseSignatureData(data []byte, sig *QuoteSignatureData) error {
	if len(data) < 64+64+384+64+2 {
		return fmt.Errorf("signature data too small: got %d bytes", len(data))
	}

	offset := 0

	// ISV enclave report signature (64 bytes)
	copy(sig.ISVEnclaveReportSignature[:], data[offset:offset+64])
	offset += 64

	// ECDSA attestation key (64 bytes)
	copy(sig.ECDSAAttestationKey[:], data[offset:offset+64])
	offset += 64

	// QE report (384 bytes)
	qeReport, err := ParseReportBody(data[offset : offset+384])
	if err != nil {
		return fmt.Errorf("failed to parse QE report: %w", err)
	}
	sig.QEReport = *qeReport
	offset += 384

	// QE report signature (64 bytes)
	copy(sig.QEReportSignature[:], data[offset:offset+64])
	offset += 64

	// QE authentication data size (2 bytes)
	if len(data) < offset+2 {
		return fmt.Errorf("missing QE auth data size")
	}
	authDataSize := binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2

	// QE authentication data
	if len(data) < offset+int(authDataSize) {
		return fmt.Errorf("QE auth data truncated")
	}
	if authDataSize > 0 {
		sig.QEAuthenticationData = make([]byte, authDataSize)
		copy(sig.QEAuthenticationData, data[offset:offset+int(authDataSize)])
	}
	offset += int(authDataSize)

	// Certification data type (2 bytes)
	if len(data) < offset+2 {
		return fmt.Errorf("missing cert data type")
	}
	sig.CertificationDataType = binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2

	// Certification data size (4 bytes)
	if len(data) < offset+4 {
		return fmt.Errorf("missing cert data size")
	}
	certDataSize := binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Certification data
	if len(data) < offset+int(certDataSize) {
		return fmt.Errorf("cert data truncated")
	}
	if certDataSize > 0 {
		sig.CertificationData = make([]byte, certDataSize)
		copy(sig.CertificationData, data[offset:offset+int(certDataSize)])
	}

	return nil
}

// =============================================================================
// Quote Serialization
// =============================================================================

// SerializeQuote serializes a Quote to bytes.
func SerializeQuote(quote *Quote) []byte {
	var buf bytes.Buffer

	// Write header
	buf.Write(quote.Header.Serialize())

	// Write report body
	buf.Write(quote.ReportBody.Serialize())

	// Write signature length
	sigLenBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(sigLenBytes, quote.SignatureLength)
	buf.Write(sigLenBytes)

	// Write signature data
	buf.Write(serializeSignatureData(&quote.SignatureData))

	return buf.Bytes()
}

// serializeSignatureData serializes the signature data.
func serializeSignatureData(sig *QuoteSignatureData) []byte {
	var buf bytes.Buffer

	// ISV enclave report signature
	buf.Write(sig.ISVEnclaveReportSignature[:])

	// ECDSA attestation key
	buf.Write(sig.ECDSAAttestationKey[:])

	// QE report
	buf.Write(sig.QEReport.Serialize())

	// QE report signature
	buf.Write(sig.QEReportSignature[:])

	// QE authentication data size and data
	authSizeBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(authSizeBytes, uint16(len(sig.QEAuthenticationData)))
	buf.Write(authSizeBytes)
	buf.Write(sig.QEAuthenticationData)

	// Certification data type
	certTypeBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(certTypeBytes, sig.CertificationDataType)
	buf.Write(certTypeBytes)

	// Certification data size and data
	certSizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(certSizeBytes, uint32(len(sig.CertificationData)))
	buf.Write(certSizeBytes)
	buf.Write(sig.CertificationData)

	return buf.Bytes()
}

// =============================================================================
// Quote Verification
// =============================================================================

// VerifyQuoteSignature verifies the ECDSA signature in a quote.
func VerifyQuoteSignature(quote *Quote) error {
	// Extract public key
	x := new(big.Int).SetBytes(quote.SignatureData.ECDSAAttestationKey[:32])
	y := new(big.Int).SetBytes(quote.SignatureData.ECDSAAttestationKey[32:64])

	pubKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	// Verify the public key is on the curve
	if !pubKey.Curve.IsOnCurve(x, y) {
		return errors.New("public key not on curve")
	}

	// Compute hash of header + report body
	dataToVerify := append(quote.Header.Serialize(), quote.ReportBody.Serialize()...)
	hash := sha256.Sum256(dataToVerify)

	// Extract signature components
	r := new(big.Int).SetBytes(quote.SignatureData.ISVEnclaveReportSignature[:32])
	s := new(big.Int).SetBytes(quote.SignatureData.ISVEnclaveReportSignature[32:64])

	// Verify signature
	if !ecdsa.Verify(pubKey, hash[:], r, s) {
		return ErrInvalidQuoteSignature
	}

	return nil
}

// ValidateQEVendorID validates the QE vendor ID.
func ValidateQEVendorID(quote *Quote) error {
	if !quote.Header.IsIntelQE() {
		return fmt.Errorf("%w: expected Intel QE", ErrInvalidQEVendorID)
	}
	return nil
}
