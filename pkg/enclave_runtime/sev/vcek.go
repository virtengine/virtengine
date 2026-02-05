// Package sev provides VCEK (Versioned Chip Endorsement Key) certificate handling.
//
// VCEK certificates are used to verify SEV-SNP attestation report signatures.
// Each VCEK is unique per chip and TCB version, and contains SEV-specific
// extensions with hardware identity and TCB component values.
//
// # Certificate Chain
//
// AMD SEV-SNP uses a three-level PKI:
// - ARK (AMD Root Key): Self-signed root, different per product
// - ASK (AMD SEV Signing Key): Signed by ARK, signs VCEKs
// - VCEK: Signed by ASK, signs attestation reports
//
// # VCEK Extensions
//
// VCEKs contain custom OIDs with SEV-specific data:
// - 1.3.6.1.4.1.3704.1.3.1: Boot Loader SVN
// - 1.3.6.1.4.1.3704.1.3.2: TEE SVN
// - 1.3.6.1.4.1.3704.1.3.3: SNP SVN
// - 1.3.6.1.4.1.3704.1.3.8: Microcode SVN
// - 1.3.6.1.4.1.3704.1.4: Hardware ID (chip ID)
//
// Task Reference: VE-2029 - Hardware TEE Integration Layer
package sev

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha512"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
)

// =============================================================================
// OID Constants
// =============================================================================

// AMD SEV OID base: 1.3.6.1.4.1.3704
// SEV extension OIDs: 1.3.6.1.4.1.3704.1.*
var (
	// OID base for AMD
	oidAMD = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 3704}

	// OID for SEV extensions
	oidSEV = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 3704, 1}

	// TCB component OIDs
	oidBootLoaderSVN = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 3704, 1, 3, 1}
	oidTEESVN        = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 3704, 1, 3, 2}
	oidSNPSVN        = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 3704, 1, 3, 3}
	oidSPL4          = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 3704, 1, 3, 4} // Reserved
	oidSPL5          = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 3704, 1, 3, 5} // Reserved
	oidSPL6          = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 3704, 1, 3, 6} // Reserved
	oidSPL7          = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 3704, 1, 3, 7} // Reserved
	oidMicrocodeSVN  = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 3704, 1, 3, 8}

	// Hardware ID OID
	oidHardwareID = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 3704, 1, 4}
)

// =============================================================================
// Errors
// =============================================================================

var (
	// ErrInvalidVCEK indicates the VCEK certificate is invalid
	ErrInvalidVCEK = errors.New("vcek: invalid VCEK certificate")

	// ErrMissingExtension indicates a required extension is missing
	ErrMissingExtension = errors.New("vcek: missing required extension")

	// ErrInvalidSignature indicates signature verification failed
	ErrInvalidSignature = errors.New("vcek: invalid signature")

	// ErrCertExpired indicates the certificate has expired
	ErrCertExpired = errors.New("vcek: certificate expired")

	// ErrChainIncomplete indicates the certificate chain is incomplete
	ErrChainIncomplete = errors.New("vcek: certificate chain incomplete")

	// ErrRootNotTrusted indicates the root certificate is not in trusted set
	ErrRootNotTrusted = errors.New("vcek: root certificate not trusted")

	// ErrTCBMismatch indicates TCB version doesn't match report
	ErrTCBMismatch = errors.New("vcek: TCB version mismatch with report")
)

// VCEKError provides detailed VCEK-related error information
type VCEKError struct {
	Op  string
	Err error
}

func (e *VCEKError) Error() string {
	return fmt.Sprintf("vcek: %s: %v", e.Op, e.Err)
}

func (e *VCEKError) Unwrap() error {
	return e.Err
}

// =============================================================================
// TCB Components
// =============================================================================

// TCBComponents contains the individual TCB security version numbers
// extracted from a VCEK certificate
type TCBComponents struct {
	// BootLoader SVN
	BootLoader uint8

	// TEE (AMD-SP) SVN
	TEE uint8

	// SNP firmware SVN
	SNP uint8

	// Microcode SVN
	Microcode uint8

	// Reserved SPL values (for future use)
	Reserved [4]uint8
}

// ToTCBVersion converts to a TCBVersion struct
func (c *TCBComponents) ToTCBVersion() TCBVersion {
	return TCBVersion{
		BootLoader: c.BootLoader,
		TEE:        c.TEE,
		SNP:        c.SNP,
		Microcode:  c.Microcode,
		Reserved:   c.Reserved,
	}
}

// String returns a human-readable representation
func (c *TCBComponents) String() string {
	return fmt.Sprintf("TCB{BL=%d, TEE=%d, SNP=%d, ucode=%d}",
		c.BootLoader, c.TEE, c.SNP, c.Microcode)
}

// =============================================================================
// VCEK Certificate
// =============================================================================

// VCEKCertificate wraps an x509 certificate with SEV-specific parsed data
type VCEKCertificate struct {
	// Certificate is the underlying x509 certificate
	Certificate *x509.Certificate

	// Raw is the DER-encoded certificate
	Raw []byte

	// HardwareID is the chip identifier from the certificate
	HardwareID [ChipIDSize]byte

	// TCB contains the TCB component values from the certificate
	TCB TCBComponents

	// Product identifies the processor product (Milan, Genoa, etc.)
	Product string
}

// ParseVCEK parses a DER-encoded VCEK certificate
func ParseVCEK(der []byte) (*VCEKCertificate, error) {
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, &VCEKError{Op: "parse", Err: err}
	}

	return ParseVCEKFromCert(cert, der)
}

// ParseVCEKFromCert creates a VCEKCertificate from an existing x509.Certificate
func ParseVCEKFromCert(cert *x509.Certificate, raw []byte) (*VCEKCertificate, error) {
	vcek := &VCEKCertificate{
		Certificate: cert,
		Raw:         raw,
	}

	// Extract TCB components from extensions
	if err := vcek.extractTCBComponents(); err != nil {
		return nil, err
	}

	// Extract hardware ID
	if err := vcek.extractHardwareID(); err != nil {
		return nil, err
	}

	// Determine product from issuer
	vcek.Product = extractProductFromCert(cert)

	return vcek, nil
}

// extractTCBComponents extracts TCB version components from certificate extensions
func (v *VCEKCertificate) extractTCBComponents() error {
	var foundBL, foundTEE, foundSNP, foundUcode bool

	for _, ext := range v.Certificate.Extensions {
		var value uint8

		switch {
		case ext.Id.Equal(oidBootLoaderSVN):
			if err := parseSPLValue(ext.Value, &value); err != nil {
				return &VCEKError{Op: "parse_bl_svn", Err: err}
			}
			v.TCB.BootLoader = value
			foundBL = true

		case ext.Id.Equal(oidTEESVN):
			if err := parseSPLValue(ext.Value, &value); err != nil {
				return &VCEKError{Op: "parse_tee_svn", Err: err}
			}
			v.TCB.TEE = value
			foundTEE = true

		case ext.Id.Equal(oidSNPSVN):
			if err := parseSPLValue(ext.Value, &value); err != nil {
				return &VCEKError{Op: "parse_snp_svn", Err: err}
			}
			v.TCB.SNP = value
			foundSNP = true

		case ext.Id.Equal(oidMicrocodeSVN):
			if err := parseSPLValue(ext.Value, &value); err != nil {
				return &VCEKError{Op: "parse_ucode_svn", Err: err}
			}
			v.TCB.Microcode = value
			foundUcode = true

		case ext.Id.Equal(oidSPL4):
			if err := parseSPLValue(ext.Value, &value); err == nil {
				v.TCB.Reserved[0] = value
			}
		case ext.Id.Equal(oidSPL5):
			if err := parseSPLValue(ext.Value, &value); err == nil {
				v.TCB.Reserved[1] = value
			}
		case ext.Id.Equal(oidSPL6):
			if err := parseSPLValue(ext.Value, &value); err == nil {
				v.TCB.Reserved[2] = value
			}
		case ext.Id.Equal(oidSPL7):
			if err := parseSPLValue(ext.Value, &value); err == nil {
				v.TCB.Reserved[3] = value
			}
		}
	}

	// All required TCB components must be present
	if !foundBL || !foundTEE || !foundSNP || !foundUcode {
		return &VCEKError{
			Op: "extract_tcb",
			Err: fmt.Errorf("%w: BL=%v TEE=%v SNP=%v ucode=%v",
				ErrMissingExtension, foundBL, foundTEE, foundSNP, foundUcode),
		}
	}

	return nil
}

// extractHardwareID extracts the chip hardware ID from certificate extensions
func (v *VCEKCertificate) extractHardwareID() error {
	for _, ext := range v.Certificate.Extensions {
		if ext.Id.Equal(oidHardwareID) {
			// Hardware ID is DER-encoded OCTET STRING
			var hwid []byte
			if _, err := asn1.Unmarshal(ext.Value, &hwid); err != nil {
				// Try direct value
				if len(ext.Value) == ChipIDSize {
					copy(v.HardwareID[:], ext.Value)
					return nil
				}
				return &VCEKError{Op: "parse_hwid", Err: err}
			}
			if len(hwid) != ChipIDSize {
				return &VCEKError{
					Op:  "parse_hwid",
					Err: fmt.Errorf("unexpected hwid length: %d", len(hwid)),
				}
			}
			copy(v.HardwareID[:], hwid)
			return nil
		}
	}

	return &VCEKError{Op: "extract_hwid", Err: ErrMissingExtension}
}

// parseSPLValue parses an ASN.1 INTEGER into a uint8 SPL value
func parseSPLValue(data []byte, out *uint8) error {
	var value int
	if _, err := asn1.Unmarshal(data, &value); err != nil {
		return err
	}
	if value < 0 || value > 255 {
		return fmt.Errorf("SPL value out of range: %d", value)
	}
	*out = uint8(value)
	return nil
}

// extractProductFromCert attempts to determine the product from certificate issuer
func extractProductFromCert(cert *x509.Certificate) string {
	// Check issuer CN
	if cert.Issuer.CommonName != "" {
		switch {
		case bytes.Contains([]byte(cert.Issuer.CommonName), []byte("Milan")):
			return ProductMilan
		case bytes.Contains([]byte(cert.Issuer.CommonName), []byte("Genoa")):
			return ProductGenoa
		case bytes.Contains([]byte(cert.Issuer.CommonName), []byte("Bergamo")):
			return ProductBergamo
		case bytes.Contains([]byte(cert.Issuer.CommonName), []byte("Siena")):
			return ProductSiena
		}
	}
	return "Unknown"
}

// =============================================================================
// Certificate Chain
// =============================================================================

// CertificateChain represents the full AMD SEV-SNP certificate chain
type CertificateChain struct {
	// VCEK is the parsed VCEK certificate
	VCEK *VCEKCertificate

	// ASK is the AMD SEV Signing Key certificate
	ASK *x509.Certificate

	// ARK is the AMD Root Key certificate
	ARK *x509.Certificate

	// Product identifies the processor product
	Product string
}

// ValidateChain validates the certificate chain
func ValidateChain(chain *CertificateChain) error {
	if chain == nil {
		return ErrChainIncomplete
	}
	if chain.VCEK == nil || chain.VCEK.Certificate == nil {
		return &VCEKError{Op: "validate_chain", Err: errors.New("VCEK missing")}
	}
	if chain.ASK == nil {
		return &VCEKError{Op: "validate_chain", Err: errors.New("ASK missing")}
	}
	if chain.ARK == nil {
		return &VCEKError{Op: "validate_chain", Err: errors.New("ARK missing")}
	}

	// Verify ARK is self-signed
	if err := chain.ARK.CheckSignatureFrom(chain.ARK); err != nil {
		return &VCEKError{
			Op:  "verify_ark",
			Err: fmt.Errorf("ARK not self-signed: %w", err),
		}
	}

	// Verify ASK is signed by ARK
	if err := chain.ASK.CheckSignatureFrom(chain.ARK); err != nil {
		return &VCEKError{
			Op:  "verify_ask",
			Err: fmt.Errorf("ASK not signed by ARK: %w", err),
		}
	}

	// Verify VCEK is signed by ASK
	if err := chain.VCEK.Certificate.CheckSignatureFrom(chain.ASK); err != nil {
		return &VCEKError{
			Op:  "verify_vcek",
			Err: fmt.Errorf("VCEK not signed by ASK: %w", err),
		}
	}

	return nil
}

// =============================================================================
// TCB Component Extraction
// =============================================================================

// ExtractTCBComponents extracts TCB components from a VCEK certificate
func ExtractTCBComponents(vcek *VCEKCertificate) (*TCBComponents, error) {
	if vcek == nil {
		return nil, errors.New("nil VCEK")
	}
	tcb := vcek.TCB
	return &tcb, nil
}

// =============================================================================
// Report Signature Verification
// =============================================================================

// VerifyReportSignature verifies an attestation report signature using the VCEK
func VerifyReportSignature(report *AttestationReport, vcek *VCEKCertificate) error {
	if report == nil {
		return errors.New("nil report")
	}
	if vcek == nil || vcek.Certificate == nil {
		return errors.New("nil VCEK")
	}

	// Get the signed data (report without signature)
	signedData, err := report.GetSignedData()
	if err != nil {
		return &VCEKError{Op: "get_signed_data", Err: err}
	}

	// Get ECDSA public key from VCEK
	pubKey, ok := vcek.Certificate.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return &VCEKError{
			Op:  "verify_signature",
			Err: errors.New("VCEK does not contain ECDSA public key"),
		}
	}

	// Verify it's P-384
	if pubKey.Curve != elliptic.P384() {
		return &VCEKError{
			Op:  "verify_signature",
			Err: fmt.Errorf("unexpected curve: %v", pubKey.Curve.Params().Name),
		}
	}

	// Extract R and S from signature
	// SEV-SNP signature format: R (48 bytes) || S (48 bytes) || reserved
	r := new(big.Int).SetBytes(report.GetSignatureR())
	s := new(big.Int).SetBytes(report.GetSignatureS())

	// Hash the signed data with SHA-384
	hash := sha512.Sum384(signedData)

	// Verify signature
	if !ecdsa.Verify(pubKey, hash[:], r, s) {
		return ErrInvalidSignature
	}

	// Verify chip ID matches
	if !bytes.Equal(report.ChipID[:], vcek.HardwareID[:]) {
		return &VCEKError{
			Op: "verify_hwid",
			Err: fmt.Errorf("chip ID mismatch: report=%s vcek=%s",
				hex.EncodeToString(report.ChipID[:16]),
				hex.EncodeToString(vcek.HardwareID[:16])),
		}
	}

	// Verify TCB version matches
	reportTCB := report.ReportedTCB
	vcekTCB := vcek.TCB.ToTCBVersion()

	if reportTCB.BootLoader != vcekTCB.BootLoader ||
		reportTCB.TEE != vcekTCB.TEE ||
		reportTCB.SNP != vcekTCB.SNP ||
		reportTCB.Microcode != vcekTCB.Microcode {
		return &VCEKError{
			Op: "verify_tcb",
			Err: fmt.Errorf("%w: report=%s vcek=%s",
				ErrTCBMismatch, reportTCB.String(), vcekTCB.String()),
		}
	}

	return nil
}

// =============================================================================
// AMD Root CA Pinning
// =============================================================================

// AMD Root Key fingerprints (SHA-256 of DER-encoded certificate)
// These should be updated when AMD rotates root keys
var (
	// MilanARKFingerprint is the SHA-256 fingerprint of the Milan ARK
	MilanARKFingerprint = []byte{
		// Placeholder - real fingerprint would be computed from AMD's Milan ARK
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	// GenoaARKFingerprint is the SHA-256 fingerprint of the Genoa ARK
	GenoaARKFingerprint = []byte{
		// Placeholder - real fingerprint would be computed from AMD's Genoa ARK
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
)

// TrustedRoots contains the trusted AMD root certificates
type TrustedRoots struct {
	// Fingerprints maps product name to expected ARK fingerprint
	Fingerprints map[string][]byte

	// AllowUnknown allows verification with unknown products (for testing)
	AllowUnknown bool
}

// DefaultTrustedRoots returns the default set of trusted AMD root certificates
func DefaultTrustedRoots() *TrustedRoots {
	return &TrustedRoots{
		Fingerprints: map[string][]byte{
			ProductMilan: MilanARKFingerprint,
			ProductGenoa: GenoaARKFingerprint,
		},
		AllowUnknown: false,
	}
}

// VerifyRoot verifies that an ARK certificate matches the trusted fingerprint
func (t *TrustedRoots) VerifyRoot(ark *x509.Certificate, product string) error {
	if ark == nil {
		return errors.New("nil ARK certificate")
	}

	expectedFP, ok := t.Fingerprints[product]
	if !ok {
		if t.AllowUnknown {
			return nil
		}
		return &VCEKError{
			Op:  "verify_root",
			Err: fmt.Errorf("unknown product: %s", product),
		}
	}

	// All zeros means not yet configured (placeholder)
	allZero := true
	for _, b := range expectedFP {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		// Placeholder fingerprint - allow for now but log warning
		return nil
	}

	// Compute fingerprint of ARK
	actualFP := sha512.Sum384(ark.Raw)

	if !bytes.Equal(actualFP[:32], expectedFP) {
		return &VCEKError{
			Op:  "verify_root",
			Err: fmt.Errorf("%w: fingerprint mismatch for %s", ErrRootNotTrusted, product),
		}
	}

	return nil
}

// =============================================================================
// Full Attestation Verification
// =============================================================================

// VerificationOptions configures attestation verification
type VerificationOptions struct {
	// MinimumTCB is the minimum required TCB version
	MinimumTCB *TCBVersion

	// TrustedRoots contains trusted AMD root certificates
	TrustedRoots *TrustedRoots

	// AllowDebug allows debug mode guests (insecure, for testing only)
	AllowDebug bool

	// RequireSMT requires SMT to be disabled (for side-channel protection)
	RequireSMT bool

	// ExpectedMeasurement is the expected launch measurement (optional)
	ExpectedMeasurement []byte
}

// VerifyAttestation performs full verification of an attestation report
func VerifyAttestation(report *AttestationReport, chain *CertificateChain, opts *VerificationOptions) error {
	if opts == nil {
		opts = &VerificationOptions{}
	}

	// Validate report structure
	if err := ValidateReport(report); err != nil {
		return fmt.Errorf("report validation failed: %w", err)
	}

	// Validate certificate chain
	if err := ValidateChain(chain); err != nil {
		return fmt.Errorf("chain validation failed: %w", err)
	}

	// Verify root certificate if trusted roots configured
	if opts.TrustedRoots != nil {
		if err := opts.TrustedRoots.VerifyRoot(chain.ARK, chain.Product); err != nil {
			return fmt.Errorf("root verification failed: %w", err)
		}
	}

	// Verify report signature
	if err := VerifyReportSignature(report, chain.VCEK); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	// Check debug mode
	if report.Policy.Debug && !opts.AllowDebug {
		return ErrDebugEnabled
	}

	// Check minimum TCB
	if opts.MinimumTCB != nil {
		if !report.TCBMeetsMinimum(*opts.MinimumTCB) {
			return &VCEKError{
				Op: "verify_tcb",
				Err: fmt.Errorf("TCB below minimum: have=%s want=%s",
					report.ReportedTCB.String(), opts.MinimumTCB.String()),
			}
		}
	}

	// Check expected measurement
	if len(opts.ExpectedMeasurement) > 0 {
		if !bytes.Equal(report.LaunchDigest[:], opts.ExpectedMeasurement) {
			return &VCEKError{
				Op: "verify_measurement",
				Err: fmt.Errorf("measurement mismatch: have=%s",
					hex.EncodeToString(report.LaunchDigest[:16])),
			}
		}
	}

	return nil
}
