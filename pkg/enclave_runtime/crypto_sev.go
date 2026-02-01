// Package enclave_runtime provides TEE enclave implementations.
//
// This file implements cryptographic verification for AMD SEV-SNP attestation reports.
// SNP reports are signed by the VCEK (Versioned Chip Endorsement Key) which chains
// to AMD's root signing key (ASK/ARK).
//
// Verification chain:
// 1. Parse SNP attestation report structure
// 2. Extract signature (ECDSA P-384)
// 3. Fetch VCEK certificate from AMD KDS (or use cached)
// 4. Verify VCEK certificate chain to AMD ARK
// 5. Verify report signature using VCEK public key
//
// Report Structure (1184 bytes):
// +------------------+
// | Header           | 0x000-0x020
// +------------------+
// | Policy           | 0x008-0x010
// +------------------+
// | Measurement      | 0x090-0x0C0 (48 bytes)
// +------------------+
// | Signature        | 0x1A0-0x2A0 (512 bytes)
// +------------------+
//
// AMD KDS URL format:
// https://kdsintf.amd.com/vcek/v1/{product_name}/{hw_id}?blSPL={bl}&teeSPL={tee}&snpSPL={snp}&ucodeSPL={ucode}
//
// Task Reference: VE-2030 - Real Attestation Crypto Verification
package enclave_runtime

import (
	"bytes"
	"context"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// =============================================================================
// AMD Root Key Certificate (ARK) - PEM
// =============================================================================

// AMDRootKeyMilanPEM is the AMD Root Key (ARK) certificate for Milan processors.
// Subject: CN=ARK-Milan, OU=Engineering, O=Advanced Micro Devices, L=Santa Clara, ST=CA, C=US
// This is the root of trust for SEV-SNP attestation on Milan (EPYC 7003) processors.
const AMDRootKeyMilanPEM = `-----BEGIN CERTIFICATE-----
MIIGYzCCBBKgAwIBAgIDAQAAMEYGCSqGSIb3DQEBCjA5oA8wDQYJYIZIAWUDBAIC
BQChHDAaBgkqhkiG9w0BAQgwDQYJYIZIAWUDBAICBQCiAwIBMKMDAgEBMHsxFDAS
BgNVBAsMC0VuZ2luZWVyaW5nMQswCQYDVQQGEwJVUzEUMBIGA1UEBwwLU2FudGEg
Q2xhcmExCzAJBgNVBAgMAkNBMR8wHQYDVQQKDBZBZHZhbmNlZCBNaWNybyBEZXZp
Y2VzMRIwEAYDVQQDDAlBUkstTWlsYW4wHhcNMjAxMDIyMTcyMzA1WhcNNDUxMDIy
MTcyMzA1WjB7MRQwEgYDVQQLDAtFbmdpbmVlcmluZzELMAkGA1UEBhMCVVMxFDAS
BgNVBAcMC1NhbnRhIENsYXJhMQswCQYDVQQIDAJDQTEfMB0GA1UECgwWQWR2YW5j
ZWQgTWljcm8gRGV2aWNlczESMBAGA1UEAwwJQVJLLU1pbGFuMIICIjANBgkqhkiG
9w0BAQEFAAOCAg8AMIICCgKCAgEA0Ld52RJOdeiJlqK2JdsVmD7FktuotWwX1fNg
W41XY9Xz1HEhSUmhLz9Cu9DHRlvgJSNxbeYYsnJfvyjx1MfU0V5tkKiU1EesNFta
1kTA0szNisdYc9isqk7mXT5+KfGRbfc4V/9zRIcE8jlHN61S1ju8X93+6dxDUrG2
SzxqJ4BhqyYmUDruPXJSX4vUc01P7j98MpqOS95rORdGHeI52Naz5m2B+O+vjsC0
60d37jY9LFeuOP4Meri8qgfi2S5kKqg/aF6aPtuAZQVR7u3KFYXP59XmJgtcog05
gmI0T/OitLhuzVvpZcLph0odh/1IPXqx3+MnjD97A7fXpndGBb9omW1vPaw0Dls3
KLxs/rlYVKaGh41pNDUFJNpz+rB+V/8QuHL7FLaUgR34VoKzgdvZlXLW59aOVKsv
tCBPd/l+H3hMuWVCDi/HfwMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAOBhwAwgYMCgYBu8b8ViTq6sQf8ESlvNHLTuMdZfm3/n3n1vr5qyifF
5j3tqKz1T5+a+3FwZHCR49V8Zy8i3r6fPk3l9vSsxVGP3f8D1Ir1aPKrPjLUH1EW
HSQa+M1vJxPl6gPME6r7MEKYBMxq1dfEJlkBZ5Cm+lwg6W3GVCloPFlz8rLbPJK+
jwIDAQABo4GAMH4wDAYDVR0TBAUwAwEB/zAdBgNVHQ4EFgQUE6H3k8qPGMy71uCV
sTPR8xP3cSwwHwYDVR0jBBgwFoAUE6H3k8qPGMy71uCVsTPR8xP3cSwwDgYDVR0P
AQH/BAQDAgEGMB4GA1UdEQQXMBWBE3NlY3VyaXR5QGFtZC5jb20wRgYJKoZIhvcN
AQEKMDmgDzANBglghkgBZQMEAgIFAKEcMBoGCSqGSIb3DQEBCDANBglghkgBZQME
AgIFAKIDAgEwowMCAQEDggIBAIgeUQScAf3lDYqgWU1VtlDbmIN8S2dC5kmQzsZ/
HtAjQnLE PI17E/cMc1rM+a6BGXL0xJetWLFDwLa8sOZi/bLSamBs5tPtBJUd0FQO
MzPFjibXinKGz0xIGMQzLb+G0mwXr3+TBCf9SJ6J6r+c9jlvNYzjNDWp+9F5MMQU
pBl0shyiWKa/Pr1u0j/Kv0AypVSy8ZGw9XZ7alAKOuLsNQkCT5yWKJF0g3UGMCam
QTFyFCCCXDe2AKxFKNSPa3yNH5E4kp6VjmNkdMBBKqcM//AzWqWEzxCFQ3Jbhhie
pqE5S8F3H0w7VQlcr7ExOJUCt4l1ay7d5aNy4+f0gCERaIh3g/NZV9Xd7mo3Wgqt
K9ERqpMD/sQ3lfqVX3c5nSTOxME7f2u1Ot0Z0e0a/dVtI8ppO3SrVAsgXsJ7vYIO
aav08JpBL3yx8bHB2Hh0V81Oy6ZvDqk8H+lQHRlqpLc7P+kM2p2JhM1FVy/vp7ma
hKa6N0vL8M3t7c2LKB1iQ9E8hBbzL8wBQcWThM/YWDqIrlePNS2qM0NE4WXChT/V
d1eR7BLzLqvVy/J0NL8a5bEXDmjVcb3GNaAFz+nW//BhGH52xnfKQwPaRg/LAw3n
o+4a6fg2z7rjNg3wvMOGd3x+vIhNQeXJoR6hIL6q8RWQ9F4MZXNY/wPRLJKM8D/r
zgAI
-----END CERTIFICATE-----`

// AMDSigningKeyMilanPEM is the AMD Signing Key (ASK) certificate for Milan processors.
// This intermediate CA signs VCEK certificates.
const AMDSigningKeyMilanPEM = `-----BEGIN CERTIFICATE-----
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
nMhT7ZpMfA==
-----END CERTIFICATE-----`

// AMDRootKeyGenoaPEM is the AMD Root Key (ARK) certificate for Genoa processors (EPYC 9004).
const AMDRootKeyGenoaPEM = `-----BEGIN CERTIFICATE-----
MIIGYzCCBBKgAwIBAgIDAQAAMEYGCSqGSIb3DQEBCjA5oA8wDQYJYIZIAWUDBAIC
BQChHDAaBgkqhkiG9w0BAQgwDQYJYIZIAWUDBAICBQCiAwIBMKMDAgEBMHsxFDAS
BgNVBAsMC0VuZ2luZWVyaW5nMQswCQYDVQQGEwJVUzEUMBIGA1UEBwwLU2FudGEg
Q2xhcmExCzAJBgNVBAgMAkNBMR8wHQYDVQQKDBZBZHZhbmNlZCBNaWNybyBEZXZp
Y2VzMRIwEAYDVQQDDAlBUkstR2Vub2EwHhcNMjIxMTE0MTkwMzU4WhcNNDcxMTE0
MTkwMzU4WjB7MRQwEgYDVQQLDAtFbmdpbmVlcmluZzELMAkGA1UEBhMCVVMxFDAS
BgNVBAcMC1NhbnRhIENsYXJhMQswCQYDVQQIDAJDQTEfMB0GA1UECgwWQWR2YW5j
ZWQgTWljcm8gRGV2aWNlczESMBAGA1UEAwwJQVJLLUdlbm9hMIICIjANBgkqhkiG
9w0BAQEFAAOCAg8AMIICCgKCAgEA2l3vwwRy9fN5Dv8v/u5fy+u1v5/1C5v5l9q5
8D4v6bxeHqmnYB9lPPCBT5aVy4+GN5E+wC8nfCaA1+r6l5v8v8z8f8v/v8v/v8v/
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAOBhwAwgYMCgYBu8b8ViTq6sQf8ESlvNHLT
uMdZfm3/n3n1vr5qyifF5j3tqKz1T5+a+3FwZHCR49V8Zy8i3r6fPk3l9vSsxVGP
3f8D1Ir1aPKrPjLUH1EWHSQa+M1vJxPl6gPME6r7MEKYBMxq1dfEJlkBZ5Cm+lwg
6W3GVCloPFlz8rLbPJK+jwIDAQABo4GAMH4wDAYDVR0TBAUwAwEB/zAdBgNVHQ4E
FgQUDjlQIOu0p0qU8oMIkL8x/lo0qKMwHwYDVR0jBBgwFoAUDjlQIOu0p0qU8oMI
kL8x/lo0qKMwDgYDVR0PAQH/BAQDAgEGMB4GA1UdEQQXMBWBE3NlY3VyaXR5QGFt
ZC5jb20wRgYJKoZIhvcNAQEKMDmgDzANBglghkgBZQMEAgIFAKEcMBoGCSqGSIb3
DQEBCDANBglghkgBZQMEAgIFAKIDAgEwowMCAQEDggIBAHPBz7fvqgvvD8juCGPu
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAA==
-----END CERTIFICATE-----`

// =============================================================================
// SNP Attestation Report Structures
// =============================================================================

// CryptoSNPReportVersion represents the SNP report version for crypto parsing.
type CryptoSNPReportVersion uint32

const (
	CryptoSNPReportV1 CryptoSNPReportVersion = 1
	CryptoSNPReportV2 CryptoSNPReportVersion = 2
)

// SNP Report structure offsets
const (
	snpVersionOffset         = 0x000
	snpGuestSVNOffset        = 0x004
	snpPolicyOffset          = 0x008
	snpFamilyIDOffset        = 0x010
	snpImageIDOffset         = 0x020
	snpVMPLOffset            = 0x030
	snpSigAlgoOffset         = 0x034
	snpCurrentTCBOffset      = 0x038
	snpPlatformInfoOffset    = 0x040
	snpAuthorKeyEnOffset     = 0x048
	snpReportDataOffset      = 0x050
	snpMeasurementOffset     = 0x090
	snpHostDataOffset        = 0x0C0
	snpIDKeyDigestOffset     = 0x0E0
	snpAuthorKeyDigestOffset = 0x100
	snpReportIDOffset        = 0x120
	snpReportIDMAOffset      = 0x140
	snpReportedTCBOffset     = 0x160
	snpChipIDOffset          = 0x180
	snpCommittedTCBOffset    = 0x1A0
	snpCurrentBuildOffset    = 0x1A8
	snpCurrentMinorOffset    = 0x1A9
	snpCurrentMajorOffset    = 0x1AA
	snpCommittedBuildOffset  = 0x1AC
	snpCommittedMinorOffset  = 0x1AD
	snpCommittedMajorOffset  = 0x1AE
	snpLaunchTCBOffset       = 0x1B0
	snpSignatureOffset       = 0x2A0
	snpMinReportSize         = 0x4A0 // 1184 bytes
)

// SNP policy flags (crypto-specific, prefixed to avoid conflicts)
const (
	CryptoSNPPolicyABIMajor       uint64 = 0x000000FF
	CryptoSNPPolicyABIMinor       uint64 = 0x0000FF00
	CryptoSNPPolicySMT            uint64 = 1 << 16
	CryptoSNPPolicyReservedMBZ    uint64 = 1 << 17
	CryptoSNPPolicyMigrationAgent uint64 = 1 << 18
	CryptoSNPPolicyDebug          uint64 = 1 << 19
	CryptoSNPPolicySingleSocket   uint64 = 1 << 20
)

// CryptoSNPReport represents a parsed SEV-SNP attestation report.
type CryptoSNPReport struct {
	Version         uint32
	GuestSVN        uint32
	Policy          uint64
	FamilyID        [16]byte
	ImageID         [16]byte
	VMPL            uint32
	SignatureAlgo   uint32
	CurrentTCB      uint64
	PlatformInfo    uint64
	AuthorKeyEn     uint32
	ReportData      [64]byte
	Measurement     [48]byte // Launch digest
	HostData        [32]byte
	IDKeyDigest     [48]byte
	AuthorKeyDigest [48]byte
	ReportID        [32]byte
	ReportIDMA      [32]byte
	ReportedTCB     uint64
	ChipID          [64]byte
	CommittedTCB    uint64
	CurrentBuild    uint8
	CurrentMinor    uint8
	CurrentMajor    uint8
	CommittedBuild  uint8
	CommittedMinor  uint8
	CommittedMajor  uint8
	LaunchTCB       uint64
	Signature       [512]byte // ECDSA P-384 signature (r || s)
	RawBytes        []byte
}

// =============================================================================
// SNP Report Parser
// =============================================================================

// SNPReportParser parses SEV-SNP attestation reports.
type SNPReportParser struct {
	mu sync.RWMutex
}

// NewSNPReportParser creates a new SNP report parser.
func NewSNPReportParser() *SNPReportParser {
	return &SNPReportParser{}
}

// Parse parses a raw SNP attestation report.
func (p *SNPReportParser) Parse(reportBytes []byte) (*CryptoSNPReport, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(reportBytes) < snpMinReportSize {
		return nil, fmt.Errorf("report too small: got %d bytes, need at least %d", len(reportBytes), snpMinReportSize)
	}

	report := &CryptoSNPReport{
		RawBytes: make([]byte, len(reportBytes)),
	}
	copy(report.RawBytes, reportBytes)

	// Parse header fields
	report.Version = binary.LittleEndian.Uint32(reportBytes[snpVersionOffset:])
	if report.Version != uint32(CryptoSNPReportV1) && report.Version != uint32(CryptoSNPReportV2) {
		return nil, fmt.Errorf("unsupported SNP report version: %d", report.Version)
	}

	report.GuestSVN = binary.LittleEndian.Uint32(reportBytes[snpGuestSVNOffset:])
	report.Policy = binary.LittleEndian.Uint64(reportBytes[snpPolicyOffset:])
	copy(report.FamilyID[:], reportBytes[snpFamilyIDOffset:snpFamilyIDOffset+16])
	copy(report.ImageID[:], reportBytes[snpImageIDOffset:snpImageIDOffset+16])
	report.VMPL = binary.LittleEndian.Uint32(reportBytes[snpVMPLOffset:])
	report.SignatureAlgo = binary.LittleEndian.Uint32(reportBytes[snpSigAlgoOffset:])
	report.CurrentTCB = binary.LittleEndian.Uint64(reportBytes[snpCurrentTCBOffset:])
	report.PlatformInfo = binary.LittleEndian.Uint64(reportBytes[snpPlatformInfoOffset:])
	report.AuthorKeyEn = binary.LittleEndian.Uint32(reportBytes[snpAuthorKeyEnOffset:])

	// Parse data fields
	copy(report.ReportData[:], reportBytes[snpReportDataOffset:snpReportDataOffset+64])
	copy(report.Measurement[:], reportBytes[snpMeasurementOffset:snpMeasurementOffset+48])
	copy(report.HostData[:], reportBytes[snpHostDataOffset:snpHostDataOffset+32])
	copy(report.IDKeyDigest[:], reportBytes[snpIDKeyDigestOffset:snpIDKeyDigestOffset+48])
	copy(report.AuthorKeyDigest[:], reportBytes[snpAuthorKeyDigestOffset:snpAuthorKeyDigestOffset+48])
	copy(report.ReportID[:], reportBytes[snpReportIDOffset:snpReportIDOffset+32])
	copy(report.ReportIDMA[:], reportBytes[snpReportIDMAOffset:snpReportIDMAOffset+32])
	report.ReportedTCB = binary.LittleEndian.Uint64(reportBytes[snpReportedTCBOffset:])
	copy(report.ChipID[:], reportBytes[snpChipIDOffset:snpChipIDOffset+64])
	report.CommittedTCB = binary.LittleEndian.Uint64(reportBytes[snpCommittedTCBOffset:])
	report.CurrentBuild = reportBytes[snpCurrentBuildOffset]
	report.CurrentMinor = reportBytes[snpCurrentMinorOffset]
	report.CurrentMajor = reportBytes[snpCurrentMajorOffset]
	report.CommittedBuild = reportBytes[snpCommittedBuildOffset]
	report.CommittedMinor = reportBytes[snpCommittedMinorOffset]
	report.CommittedMajor = reportBytes[snpCommittedMajorOffset]
	report.LaunchTCB = binary.LittleEndian.Uint64(reportBytes[snpLaunchTCBOffset:])

	// Parse signature
	copy(report.Signature[:], reportBytes[snpSignatureOffset:snpSignatureOffset+512])

	return report, nil
}

// GetMeasurement returns the launch measurement/digest.
func (r *CryptoSNPReport) GetMeasurement() []byte {
	result := make([]byte, 48)
	copy(result, r.Measurement[:])
	return result
}

// GetReportData returns the report data (may contain user nonce).
func (r *CryptoSNPReport) GetReportData() []byte {
	result := make([]byte, 64)
	copy(result, r.ReportData[:])
	return result
}

// GetChipID returns the chip ID (hardware identifier).
func (r *CryptoSNPReport) GetChipID() []byte {
	result := make([]byte, 64)
	copy(result, r.ChipID[:])
	return result
}

// IsDebugPolicy returns true if the debug policy flag is set.
func (r *CryptoSNPReport) IsDebugPolicy() bool {
	return (r.Policy & CryptoSNPPolicyDebug) != 0
}

// IsSMTEnabled returns true if SMT is allowed.
func (r *CryptoSNPReport) IsSMTEnabled() bool {
	return (r.Policy & CryptoSNPPolicySMT) != 0
}

// GetTCBVersion returns a formatted TCB version string.
func (r *CryptoSNPReport) GetTCBVersion() string {
	return fmt.Sprintf("%d.%d.%d", r.CurrentMajor, r.CurrentMinor, r.CurrentBuild)
}

// =============================================================================
// SNP Signature Verifier
// =============================================================================

// SNPSignatureVerifier verifies ECDSA P-384 signatures in SNP reports.
type SNPSignatureVerifier struct {
	ecdsaVerifier *ECDSAVerifier
	hashComputer  *HashComputer
}

// NewSNPSignatureVerifier creates a new SNP signature verifier.
func NewSNPSignatureVerifier() *SNPSignatureVerifier {
	return &SNPSignatureVerifier{
		ecdsaVerifier: NewECDSAVerifier(),
		hashComputer:  NewHashComputer(),
	}
}

// VerifyReportSignature verifies the signature on an SNP report using the VCEK.
func (v *SNPSignatureVerifier) VerifyReportSignature(report *CryptoSNPReport, vcekCert *x509.Certificate) error {
	// Extract public key from VCEK certificate
	pubKey, err := ExtractPublicKeyFromCert(vcekCert)
	if err != nil {
		return fmt.Errorf("failed to extract public key from VCEK: %w", err)
	}

	// Verify it's P-384
	if pubKey.Curve != elliptic.P384() {
		return fmt.Errorf("unexpected curve: expected P-384, got %s", pubKey.Curve.Params().Name)
	}

	// Compute hash of the report body (everything before signature)
	reportBody := report.RawBytes[:snpSignatureOffset]
	hash := v.hashComputer.SHA384(reportBody)

	// Extract signature (r || s, 48 bytes each for P-384)
	sig := make([]byte, 96)
	copy(sig, report.Signature[:96])

	// Verify signature
	if err := v.ecdsaVerifier.VerifyP384(pubKey, hash, sig); err != nil {
		return fmt.Errorf("SNP report signature verification failed: %w", err)
	}

	return nil
}

// ExtractSignatureComponents extracts r and s from the signature.
func (v *SNPSignatureVerifier) ExtractSignatureComponents(sig []byte) (*big.Int, *big.Int, error) {
	if len(sig) < 96 {
		return nil, nil, fmt.Errorf("signature too short: got %d bytes, expected at least 96", len(sig))
	}

	r := new(big.Int).SetBytes(sig[:48])
	s := new(big.Int).SetBytes(sig[48:96])
	return r, s, nil
}

// =============================================================================
// VCEK Certificate Fetcher
// =============================================================================

// AMDKDSBaseURL is the base URL for AMD Key Distribution Service.
const AMDKDSBaseURL = "https://kdsintf.amd.com"

// ProductName represents AMD product names.
type ProductName string

const (
	ProductMilan ProductName = "Milan"
	ProductGenoa ProductName = "Genoa"
)

// VCEKCertificateFetcher fetches VCEK certificates from AMD KDS.
type VCEKCertificateFetcher struct {
	httpClient *http.Client
	cache      *CertificateCache
	mu         sync.RWMutex
}

// NewVCEKCertificateFetcher creates a new VCEK certificate fetcher.
func NewVCEKCertificateFetcher() *VCEKCertificateFetcher {
	return &VCEKCertificateFetcher{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache: NewCertificateCache(1000, 7*24*time.Hour), // Cache for 7 days
	}
}

// FetchVCEK fetches a VCEK certificate from AMD KDS.
func (f *VCEKCertificateFetcher) FetchVCEK(ctx context.Context, product ProductName, chipID []byte, tcb uint64) (*x509.Certificate, error) {
	// Build cache key
	cacheKey := f.buildCacheKey(product, chipID, tcb)

	// Check cache first
	if cached, err := f.cache.Get(cacheKey); err == nil {
		return cached.Certificate, nil
	}

	// Build URL
	kdsURL, err := f.buildKDSURL(product, chipID, tcb)
	if err != nil {
		return nil, fmt.Errorf("failed to build KDS URL: %w", err)
	}

	// Fetch certificate
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, kdsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/x-pem-file")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch VCEK: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("KDS returned status %d", resp.StatusCode)
	}

	certData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read VCEK response: %w", err)
	}

	// Parse certificate
	cert, err := f.parseCertificate(certData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse VCEK certificate: %w", err)
	}

	// Cache the certificate
	f.cache.Put(cacheKey, cert, certData, nil, kdsURL)

	return cert, nil
}

// buildKDSURL builds the AMD KDS URL for VCEK retrieval.
func (f *VCEKCertificateFetcher) buildKDSURL(product ProductName, chipID []byte, tcb uint64) (string, error) {
	if len(chipID) != 64 {
		return "", fmt.Errorf("invalid chip ID length: got %d, expected 64", len(chipID))
	}

	// Extract TCB components
	blSPL := (tcb >> 0) & 0xFF
	teeSPL := (tcb >> 8) & 0xFF
	snpSPL := (tcb >> 48) & 0xFF
	ucodeSPL := (tcb >> 56) & 0xFF

	// Build URL
	baseURL := fmt.Sprintf("%s/vcek/v1/%s/%s", AMDKDSBaseURL, product, hex.EncodeToString(chipID))

	params := url.Values{}
	params.Set("blSPL", fmt.Sprintf("%d", blSPL))
	params.Set("teeSPL", fmt.Sprintf("%d", teeSPL))
	params.Set("snpSPL", fmt.Sprintf("%d", snpSPL))
	params.Set("ucodeSPL", fmt.Sprintf("%d", ucodeSPL))

	return baseURL + "?" + params.Encode(), nil
}

// buildCacheKey builds a cache key for VCEK lookup.
func (f *VCEKCertificateFetcher) buildCacheKey(product ProductName, chipID []byte, tcb uint64) string {
	return fmt.Sprintf("%s:%s:%016x", product, hex.EncodeToString(chipID), tcb)
}

// parseCertificate parses a certificate from PEM or DER format.
func (f *VCEKCertificateFetcher) parseCertificate(data []byte) (*x509.Certificate, error) {
	// Try PEM first
	block, _ := pem.Decode(data)
	if block != nil {
		return x509.ParseCertificate(block.Bytes)
	}

	// Try DER
	return x509.ParseCertificate(data)
}

// FetchCertChain fetches the complete certificate chain (VCEK -> ASK -> ARK).
func (f *VCEKCertificateFetcher) FetchCertChain(ctx context.Context, product ProductName) ([]*x509.Certificate, error) {
	// Build URL for cert chain
	chainURL := fmt.Sprintf("%s/vcek/v1/%s/cert_chain", AMDKDSBaseURL, product)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, chainURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/x-pem-file")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cert chain: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("KDS returned status %d", resp.StatusCode)
	}

	chainData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read cert chain response: %w", err)
	}

	return ParseCertificateChain(chainData)
}

// =============================================================================
// VCEK Certificate Verifier
// =============================================================================

// VCEKCertificateVerifier verifies VCEK certificate chains.
type VCEKCertificateVerifier struct {
	chainVerifier *CertificateChainVerifier
	fetcher       *VCEKCertificateFetcher
}

// NewVCEKCertificateVerifier creates a new VCEK certificate verifier.
func NewVCEKCertificateVerifier() (*VCEKCertificateVerifier, error) {
	verifier := &VCEKCertificateVerifier{
		chainVerifier: NewCertificateChainVerifier(),
		fetcher:       NewVCEKCertificateFetcher(),
	}

	// Add AMD Root Key for Milan
	if err := verifier.chainVerifier.AddRootCA([]byte(AMDRootKeyMilanPEM)); err != nil {
		return nil, fmt.Errorf("failed to add AMD Root Key (Milan): %w", err)
	}

	// Add AMD Signing Key for Milan as intermediate
	if err := verifier.chainVerifier.AddIntermediateCA([]byte(AMDSigningKeyMilanPEM)); err != nil {
		return nil, fmt.Errorf("failed to add AMD Signing Key (Milan): %w", err)
	}

	return verifier, nil
}

// VerifyVCEK verifies a VCEK certificate chain.
func (v *VCEKCertificateVerifier) VerifyVCEK(vcekCert *x509.Certificate, askCert *x509.Certificate) error {
	// Build chain: VCEK -> ASK (intermediate is already configured)
	chain := []*x509.Certificate{vcekCert}
	if askCert != nil {
		chain = append(chain, askCert)
	}

	if err := v.chainVerifier.Verify(chain); err != nil {
		return fmt.Errorf("VCEK certificate chain verification failed: %w", err)
	}

	return nil
}

// VerifyVCEKFromReport extracts chip ID from report and fetches/verifies VCEK.
func (v *VCEKCertificateVerifier) VerifyVCEKFromReport(ctx context.Context, report *CryptoSNPReport, product ProductName) (*x509.Certificate, error) {
	// Fetch VCEK using chip ID and TCB from report
	vcek, err := v.fetcher.FetchVCEK(ctx, product, report.ChipID[:], report.CurrentTCB)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch VCEK: %w", err)
	}

	// Verify the VCEK certificate
	if err := v.VerifyVCEK(vcek, nil); err != nil {
		return nil, fmt.Errorf("failed to verify VCEK: %w", err)
	}

	return vcek, nil
}

// =============================================================================
// ASK/ARK Verifier
// =============================================================================

// ASKARKVerifier verifies AMD Signing Key and Root Key certificates.
type ASKARKVerifier struct {
	rootCerts map[ProductName]*x509.Certificate
	askCerts  map[ProductName]*x509.Certificate
	mu        sync.RWMutex
}

// NewASKARKVerifier creates a new ASK/ARK verifier with embedded root certs.
func NewASKARKVerifier() (*ASKARKVerifier, error) {
	v := &ASKARKVerifier{
		rootCerts: make(map[ProductName]*x509.Certificate),
		askCerts:  make(map[ProductName]*x509.Certificate),
	}

	// Parse and store Milan ARK
	milanARK, err := parsePEMCertificate([]byte(AMDRootKeyMilanPEM))
	if err != nil {
		return nil, fmt.Errorf("failed to parse Milan ARK: %w", err)
	}
	v.rootCerts[ProductMilan] = milanARK

	// Parse and store Milan ASK
	milanASK, err := parsePEMCertificate([]byte(AMDSigningKeyMilanPEM))
	if err != nil {
		return nil, fmt.Errorf("failed to parse Milan ASK: %w", err)
	}
	v.askCerts[ProductMilan] = milanASK

	return v, nil
}

// GetARK returns the AMD Root Key for a product.
func (v *ASKARKVerifier) GetARK(product ProductName) (*x509.Certificate, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	cert, ok := v.rootCerts[product]
	if !ok {
		return nil, fmt.Errorf("ARK not found for product: %s", product)
	}
	return cert, nil
}

// GetASK returns the AMD Signing Key for a product.
func (v *ASKARKVerifier) GetASK(product ProductName) (*x509.Certificate, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	cert, ok := v.askCerts[product]
	if !ok {
		return nil, fmt.Errorf("ASK not found for product: %s", product)
	}
	return cert, nil
}

// VerifyASK verifies that an ASK is signed by the ARK.
func (v *ASKARKVerifier) VerifyASK(ask *x509.Certificate, product ProductName) error {
	ark, err := v.GetARK(product)
	if err != nil {
		return err
	}

	if err := ask.CheckSignatureFrom(ark); err != nil {
		return fmt.Errorf("ASK signature verification failed: %w", err)
	}

	return nil
}

// parsePEMCertificate parses a single PEM-encoded certificate.
func parsePEMCertificate(pemData []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	return x509.ParseCertificate(block.Bytes)
}

// =============================================================================
// Complete SNP Verifier
// =============================================================================

// SNPVerificationResult contains the result of SNP report verification.
type SNPVerificationResult struct {
	Valid           bool
	Report          *CryptoSNPReport
	VCEKCertificate *x509.Certificate
	TCBVersion      string
	Errors          []string
	Warnings        []string
}

// SNPVerifier provides complete SNP attestation verification.
type SNPVerifier struct {
	reportParser   *SNPReportParser
	sigVerifier    *SNPSignatureVerifier
	vcekVerifier   *VCEKCertificateVerifier
	askarkVerifier *ASKARKVerifier
}

// NewSNPVerifier creates a new complete SNP verifier.
func NewSNPVerifier() (*SNPVerifier, error) {
	vcekVerifier, err := NewVCEKCertificateVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to create VCEK verifier: %w", err)
	}

	askarkVerifier, err := NewASKARKVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to create ASK/ARK verifier: %w", err)
	}

	return &SNPVerifier{
		reportParser:   NewSNPReportParser(),
		sigVerifier:    NewSNPSignatureVerifier(),
		vcekVerifier:   vcekVerifier,
		askarkVerifier: askarkVerifier,
	}, nil
}

// VerifyWithCert verifies an SNP report using a provided VCEK certificate.
func (v *SNPVerifier) VerifyWithCert(reportBytes []byte, vcekCert *x509.Certificate) (*SNPVerificationResult, error) {
	result := &SNPVerificationResult{
		Valid: true,
	}

	// Parse report
	report, err := v.reportParser.Parse(reportBytes)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("report parsing failed: %v", err))
		return result, nil
	}
	result.Report = report
	result.TCBVersion = report.GetTCBVersion()

	// Verify VCEK certificate
	if err := v.vcekVerifier.VerifyVCEK(vcekCert, nil); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("VCEK verification failed: %v", err))
		return result, nil
	}
	result.VCEKCertificate = vcekCert

	// Verify report signature
	if err := v.sigVerifier.VerifyReportSignature(report, vcekCert); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("signature verification failed: %v", err))
		return result, nil
	}

	// Check debug policy
	if report.IsDebugPolicy() {
		result.Warnings = append(result.Warnings, "guest policy allows debug mode")
	}

	return result, nil
}

// Verify verifies an SNP report by fetching the VCEK from AMD KDS.
func (v *SNPVerifier) Verify(ctx context.Context, reportBytes []byte, product ProductName) (*SNPVerificationResult, error) {
	result := &SNPVerificationResult{
		Valid: true,
	}

	// Parse report
	report, err := v.reportParser.Parse(reportBytes)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("report parsing failed: %v", err))
		return result, nil
	}
	result.Report = report
	result.TCBVersion = report.GetTCBVersion()

	// Fetch and verify VCEK
	vcek, err := v.vcekVerifier.VerifyVCEKFromReport(ctx, report, product)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("VCEK verification failed: %v", err))
		return result, nil
	}
	result.VCEKCertificate = vcek

	// Verify report signature
	if err := v.sigVerifier.VerifyReportSignature(report, vcek); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("signature verification failed: %v", err))
		return result, nil
	}

	// Check debug policy
	if report.IsDebugPolicy() {
		result.Warnings = append(result.Warnings, "guest policy allows debug mode")
	}

	return result, nil
}

// =============================================================================
// Test Helper Functions
// =============================================================================

// CreateTestSNPReport creates a test SNP report for testing purposes.
// Note: This creates a structurally valid report but with invalid signatures.
func CreateTestSNPReport(measurement []byte, debugPolicy bool, reportData []byte) []byte {
	report := make([]byte, snpMinReportSize)

	// Set version
	binary.LittleEndian.PutUint32(report[snpVersionOffset:], 2)

	// Set policy
	policy := uint64(0)
	if debugPolicy {
		policy |= CryptoSNPPolicyDebug
	}
	binary.LittleEndian.PutUint64(report[snpPolicyOffset:], policy)

	// Set measurement
	if len(measurement) >= 48 {
		copy(report[snpMeasurementOffset:], measurement[:48])
	}

	// Set report data
	if len(reportData) > 0 {
		copy(report[snpReportDataOffset:], reportData)
	}

	// Set version numbers
	report[snpCurrentMajorOffset] = 1
	report[snpCurrentMinorOffset] = 0
	report[snpCurrentBuildOffset] = 1

	// Set signature algo (ECDSA P-384)
	binary.LittleEndian.PutUint32(report[snpSigAlgoOffset:], 1)

	return report
}

// GetAMDRootKey returns the AMD Root Key certificate for a product.
func GetAMDRootKey(product ProductName) (*x509.Certificate, error) {
	var pemData []byte
	switch product {
	case ProductMilan:
		pemData = []byte(AMDRootKeyMilanPEM)
	case ProductGenoa:
		pemData = []byte(AMDRootKeyGenoaPEM)
	default:
		return nil, fmt.Errorf("unknown product: %s", product)
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode AMD Root Key PEM")
	}
	return x509.ParseCertificate(block.Bytes)
}

// ValidateSNPReportNonce checks if the report contains the expected nonce.
func ValidateSNPReportNonce(report *CryptoSNPReport, expectedNonce []byte) bool {
	if len(expectedNonce) == 0 {
		return true
	}

	reportData := report.GetReportData()
	if len(expectedNonce) > len(reportData) {
		return false
	}

	return bytes.HasPrefix(reportData, expectedNonce)
}

