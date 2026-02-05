// Package sev provides AMD SEV-SNP attestation report handling.
//
// This file implements parsing, serialization, and validation of SNP attestation
// reports according to the AMD SEV-SNP ABI specification.
//
// # Report Structure
//
// The attestation report is 1184 bytes and contains:
// - Header: Version, guest SVN, policy, family ID, image ID
// - TCB info: Current and reported TCB versions
// - Measurements: Launch digest, report data, host data
// - Identifiers: Report ID, chip ID
// - Signature: ECDSA P-384 signature from VCEK
//
// All multi-byte fields are little-endian as per AMD specification.
//
// Task Reference: VE-2029 - Hardware TEE Integration Layer
package sev

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// =============================================================================
// Constants
// =============================================================================

const (
	// Report versions
	ReportVersionV1 = 1 // Initial version (deprecated)
	ReportVersionV2 = 2 // Current version

	// Report structure offsets (bytes)
	offsetVersion          = 0
	offsetGuestSVN         = 4
	offsetPolicy           = 8
	offsetFamilyID         = 16
	offsetImageID          = 32
	offsetVMPL             = 48
	offsetSigAlgo          = 52
	offsetCurrentTCB       = 56
	offsetPlatformInfo     = 64
	offsetAuthorKeyEnabled = 72
	offsetReserved1        = 76
	offsetReportData       = 80
	offsetMeasurement      = 144 // Launch digest
	offsetHostData         = 192
	offsetIDKeyDigest      = 224
	offsetAuthorKeyDigest  = 272
	offsetReportID         = 320
	offsetReportIDMA       = 352
	offsetReportedTCB      = 384
	offsetReserved2        = 392
	offsetChipID           = 416
	offsetCommittedTCB     = 480
	offsetCurrentBuild     = 488
	offsetCurrentVersion   = 489
	offsetLaunchTCB        = 496
	offsetReserved3        = 504
	offsetSignature        = 672

	// Signature algorithm values
	SigAlgoECDSAP384SHA384 = 1

	// Guest policy bit positions
	policyABIMinorMask      = 0x00000000000000FF
	policyABIMajorMask      = 0x000000000000FF00
	policyABIMajorShift     = 8
	policySMTMask           = 0x0000000000010000
	policySMTShift          = 16
	policyReserved1Mask     = 0x0000000000020000
	policyMigrationMAMask   = 0x0000000000040000
	policyMigrationMAShift  = 18
	policyDebugMask         = 0x0000000000080000
	policyDebugShift        = 19
	policySingleSocketMask  = 0x0000000000100000
	policySingleSocketShift = 20
	policyCXLAllowMask      = 0x0000000000200000
	policyCXLAllowShift     = 21
	policyMemAESMask        = 0x0000000000400000
	policyMemAESShift       = 22
	policyRASPMask          = 0x0000000000800000
	policyRASPShift         = 23
)

// =============================================================================
// Guest Policy
// =============================================================================

// GuestPolicy represents the SNP guest policy configuration
type GuestPolicy struct {
	// ABIMinor is the minimum supported ABI minor version
	ABIMinor uint8

	// ABIMajor is the minimum supported ABI major version
	ABIMajor uint8

	// SMT indicates if Simultaneous Multi-Threading is allowed
	SMT bool

	// Migration indicates if migration agent is allowed
	Migration bool

	// Debug indicates if debug mode is enabled (MUST be false for production)
	Debug bool

	// SingleSocket restricts guest to single socket
	SingleSocket bool

	// CXLAllow allows CXL devices
	CXLAllow bool

	// MemAES256 uses AES-256 for memory encryption (vs AES-128)
	MemAES256 bool

	// RASP enables Restricted Async Stack Pointer (SNP rev 1.55+)
	RASP bool
}

// ToUint64 serializes the policy to a 64-bit value
func (p GuestPolicy) ToUint64() uint64 {
	var policy uint64

	policy |= uint64(p.ABIMinor)
	policy |= uint64(p.ABIMajor) << policyABIMajorShift

	if p.SMT {
		policy |= policySMTMask
	}
	if p.Migration {
		policy |= policyMigrationMAMask
	}
	if p.Debug {
		policy |= policyDebugMask
	}
	if p.SingleSocket {
		policy |= policySingleSocketMask
	}
	if p.CXLAllow {
		policy |= policyCXLAllowMask
	}
	if p.MemAES256 {
		policy |= policyMemAESMask
	}
	if p.RASP {
		policy |= policyRASPMask
	}

	return policy
}

// ParseGuestPolicy parses a 64-bit policy value into a GuestPolicy struct
func ParseGuestPolicy(raw uint64) GuestPolicy {
	return GuestPolicy{
		ABIMinor:     uint8(raw & policyABIMinorMask),
		ABIMajor:     uint8((raw & policyABIMajorMask) >> policyABIMajorShift),
		SMT:          raw&policySMTMask != 0,
		Migration:    raw&policyMigrationMAMask != 0,
		Debug:        raw&policyDebugMask != 0,
		SingleSocket: raw&policySingleSocketMask != 0,
		CXLAllow:     raw&policyCXLAllowMask != 0,
		MemAES256:    raw&policyMemAESMask != 0,
		RASP:         raw&policyRASPMask != 0,
	}
}

// Validate checks if the policy is valid for production use
func (p GuestPolicy) Validate() error {
	if p.Debug {
		return ErrDebugEnabled
	}
	if p.ABIMajor < 1 {
		return fmt.Errorf("ABI major version %d is below minimum", p.ABIMajor)
	}
	return nil
}

// String returns a human-readable representation
func (p GuestPolicy) String() string {
	return fmt.Sprintf("GuestPolicy{ABI=%d.%d, SMT=%v, Debug=%v, SingleSocket=%v, Migration=%v}",
		p.ABIMajor, p.ABIMinor, p.SMT, p.Debug, p.SingleSocket, p.Migration)
}

// =============================================================================
// TCB Version
// =============================================================================

// TCBVersion represents the Trusted Computing Base version
type TCBVersion struct {
	// BootLoader SVN
	BootLoader uint8

	// TEE (AMD-SP) SVN
	TEE uint8

	// Reserved bytes
	Reserved [4]uint8

	// SNP firmware SVN
	SNP uint8

	// Microcode SVN
	Microcode uint8
}

// ToUint64 serializes the TCB version to a 64-bit value
func (t TCBVersion) ToUint64() uint64 {
	return uint64(t.BootLoader) |
		uint64(t.TEE)<<8 |
		uint64(t.Reserved[0])<<16 |
		uint64(t.Reserved[1])<<24 |
		uint64(t.Reserved[2])<<32 |
		uint64(t.Reserved[3])<<40 |
		uint64(t.SNP)<<48 |
		uint64(t.Microcode)<<56
}

// ParseTCBVersion parses a 64-bit TCB version value
func ParseTCBVersion(raw uint64) TCBVersion {
	return TCBVersion{
		BootLoader: uint8(raw),
		TEE:        uint8(raw >> 8),
		Reserved: [4]uint8{
			uint8(raw >> 16),
			uint8(raw >> 24),
			uint8(raw >> 32),
			uint8(raw >> 40),
		},
		SNP:       uint8(raw >> 48),
		Microcode: uint8(raw >> 56),
	}
}

// Compare compares two TCB versions
// Returns -1 if t < other, 0 if equal, 1 if t > other
func (t TCBVersion) Compare(other TCBVersion) int {
	// Compare component by component (bootloader, tee, snp, microcode)
	if t.BootLoader != other.BootLoader {
		if t.BootLoader < other.BootLoader {
			return -1
		}
		return 1
	}
	if t.TEE != other.TEE {
		if t.TEE < other.TEE {
			return -1
		}
		return 1
	}
	if t.SNP != other.SNP {
		if t.SNP < other.SNP {
			return -1
		}
		return 1
	}
	if t.Microcode != other.Microcode {
		if t.Microcode < other.Microcode {
			return -1
		}
		return 1
	}
	return 0
}

// MeetsMinimum returns true if this TCB meets minimum requirements
func (t TCBVersion) MeetsMinimum(min TCBVersion) bool {
	return t.BootLoader >= min.BootLoader &&
		t.TEE >= min.TEE &&
		t.SNP >= min.SNP &&
		t.Microcode >= min.Microcode
}

// String returns a human-readable representation
func (t TCBVersion) String() string {
	return fmt.Sprintf("TCB{BL=%d, TEE=%d, SNP=%d, ucode=%d}",
		t.BootLoader, t.TEE, t.SNP, t.Microcode)
}

// =============================================================================
// Report Request/Response Structures
// =============================================================================

// ReportRequest is the input structure for SNP_GET_REPORT ioctl
type ReportRequest struct {
	// UserData is included verbatim in the report (64 bytes)
	UserData [ReportDataSize]byte

	// VMPL is the VM Privilege Level (0-3)
	VMPL uint32

	// Reserved padding
	Reserved [28]byte
}

// ToBytes serializes the request for ioctl
func (r *ReportRequest) ToBytes() []byte {
	buf := make([]byte, 96)
	copy(buf[0:64], r.UserData[:])
	binary.LittleEndian.PutUint32(buf[64:68], r.VMPL)
	return buf
}

// ReportResponse is the output structure from SNP_GET_REPORT ioctl
type ReportResponse struct {
	// Status is the operation status code
	Status uint32

	// Reserved padding
	Reserved [28]byte

	// ReportData contains the raw attestation report
	ReportData [ReportSize]byte
}

// ParseReportResponse parses the ioctl response
func ParseReportResponse(data []byte) (*ReportResponse, error) {
	if len(data) < 32+ReportSize {
		return nil, fmt.Errorf("response too short: %d bytes", len(data))
	}

	resp := &ReportResponse{
		Status: binary.LittleEndian.Uint32(data[0:4]),
	}
	copy(resp.Reserved[:], data[4:32])
	copy(resp.ReportData[:], data[32:32+ReportSize])

	return resp, nil
}

// =============================================================================
// Attestation Report
// =============================================================================

// AttestationReport represents a parsed SNP attestation report
type AttestationReport struct {
	// Version is the report format version (currently 2)
	Version uint32

	// GuestSVN is the guest security version number
	GuestSVN uint32

	// Policy is the guest policy
	Policy GuestPolicy

	// FamilyID identifies the guest family (16 bytes)
	FamilyID [16]byte

	// ImageID identifies the guest image (16 bytes)
	ImageID [16]byte

	// VMPL is the VM Privilege Level that generated this report
	VMPL uint32

	// SignatureAlgo is the signature algorithm (1 = ECDSA-P384-SHA384)
	SignatureAlgo uint32

	// CurrentTCB is the current platform TCB version
	CurrentTCB TCBVersion

	// PlatformInfo contains platform configuration flags
	PlatformInfo uint64

	// AuthorKeyEnabled indicates if author key is present
	AuthorKeyEnabled uint32

	// ReportData is user-provided data (64 bytes)
	ReportData [ReportDataSize]byte

	// LaunchDigest is the SHA-384 of the guest launch measurement (48 bytes)
	LaunchDigest [LaunchDigestSize]byte

	// HostData is data provided by the host (32 bytes)
	HostData [32]byte

	// IDKeyDigest is the SHA-384 of the ID key (48 bytes)
	IDKeyDigest [48]byte

	// AuthorKeyDigest is the SHA-384 of the author key (48 bytes)
	AuthorKeyDigest [48]byte

	// ReportID is the unique report identifier (32 bytes)
	ReportID [32]byte

	// ReportIDMA is the report ID for migration agent (32 bytes)
	ReportIDMA [32]byte

	// ReportedTCB is the TCB version reported in attestation
	ReportedTCB TCBVersion

	// ChipID is the unique chip identifier (64 bytes)
	ChipID [ChipIDSize]byte

	// CommittedTCB is the committed TCB version
	CommittedTCB TCBVersion

	// CurrentBuild is the current firmware build number
	CurrentBuild uint8

	// CurrentVersion is the current version minor.major
	CurrentVersion uint8

	// LaunchTCB is the TCB at guest launch
	LaunchTCB TCBVersion

	// Signature is the ECDSA P-384 signature (512 bytes: R || S || reserved)
	Signature [SignatureSize]byte
}

// ParseReport parses raw attestation report bytes into an AttestationReport
func ParseReport(data []byte) (*AttestationReport, error) {
	if len(data) < ReportSize {
		return nil, fmt.Errorf("report data too short: got %d, want %d", len(data), ReportSize)
	}

	r := &AttestationReport{}

	// Parse header fields
	r.Version = binary.LittleEndian.Uint32(data[offsetVersion:])
	r.GuestSVN = binary.LittleEndian.Uint32(data[offsetGuestSVN:])
	r.Policy = ParseGuestPolicy(binary.LittleEndian.Uint64(data[offsetPolicy:]))
	copy(r.FamilyID[:], data[offsetFamilyID:offsetFamilyID+16])
	copy(r.ImageID[:], data[offsetImageID:offsetImageID+16])
	r.VMPL = binary.LittleEndian.Uint32(data[offsetVMPL:])
	r.SignatureAlgo = binary.LittleEndian.Uint32(data[offsetSigAlgo:])

	// Parse TCB info
	r.CurrentTCB = ParseTCBVersion(binary.LittleEndian.Uint64(data[offsetCurrentTCB:]))
	r.PlatformInfo = binary.LittleEndian.Uint64(data[offsetPlatformInfo:])
	r.AuthorKeyEnabled = binary.LittleEndian.Uint32(data[offsetAuthorKeyEnabled:])

	// Parse measurements
	copy(r.ReportData[:], data[offsetReportData:offsetReportData+ReportDataSize])
	copy(r.LaunchDigest[:], data[offsetMeasurement:offsetMeasurement+LaunchDigestSize])
	copy(r.HostData[:], data[offsetHostData:offsetHostData+32])
	copy(r.IDKeyDigest[:], data[offsetIDKeyDigest:offsetIDKeyDigest+48])
	copy(r.AuthorKeyDigest[:], data[offsetAuthorKeyDigest:offsetAuthorKeyDigest+48])

	// Parse identifiers
	copy(r.ReportID[:], data[offsetReportID:offsetReportID+32])
	copy(r.ReportIDMA[:], data[offsetReportIDMA:offsetReportIDMA+32])
	r.ReportedTCB = ParseTCBVersion(binary.LittleEndian.Uint64(data[offsetReportedTCB:]))

	// Parse chip ID
	copy(r.ChipID[:], data[offsetChipID:offsetChipID+ChipIDSize])

	// Parse additional TCB fields (available in newer firmware)
	if len(data) >= offsetCommittedTCB+8 {
		r.CommittedTCB = ParseTCBVersion(binary.LittleEndian.Uint64(data[offsetCommittedTCB:]))
	}
	if len(data) >= offsetCurrentBuild+1 {
		r.CurrentBuild = data[offsetCurrentBuild]
	}
	if len(data) >= offsetCurrentVersion+1 {
		r.CurrentVersion = data[offsetCurrentVersion]
	}
	if len(data) >= offsetLaunchTCB+8 {
		r.LaunchTCB = ParseTCBVersion(binary.LittleEndian.Uint64(data[offsetLaunchTCB:]))
	}

	// Parse signature
	copy(r.Signature[:], data[offsetSignature:offsetSignature+SignatureSize])

	return r, nil
}

// SerializeReport serializes an AttestationReport to bytes
func SerializeReport(r *AttestationReport) ([]byte, error) {
	if r == nil {
		return nil, errors.New("nil report")
	}

	buf := make([]byte, ReportSize)

	// Header
	binary.LittleEndian.PutUint32(buf[offsetVersion:], r.Version)
	binary.LittleEndian.PutUint32(buf[offsetGuestSVN:], r.GuestSVN)
	binary.LittleEndian.PutUint64(buf[offsetPolicy:], r.Policy.ToUint64())
	copy(buf[offsetFamilyID:], r.FamilyID[:])
	copy(buf[offsetImageID:], r.ImageID[:])
	binary.LittleEndian.PutUint32(buf[offsetVMPL:], r.VMPL)
	binary.LittleEndian.PutUint32(buf[offsetSigAlgo:], r.SignatureAlgo)

	// TCB info
	binary.LittleEndian.PutUint64(buf[offsetCurrentTCB:], r.CurrentTCB.ToUint64())
	binary.LittleEndian.PutUint64(buf[offsetPlatformInfo:], r.PlatformInfo)
	binary.LittleEndian.PutUint32(buf[offsetAuthorKeyEnabled:], r.AuthorKeyEnabled)

	// Measurements
	copy(buf[offsetReportData:], r.ReportData[:])
	copy(buf[offsetMeasurement:], r.LaunchDigest[:])
	copy(buf[offsetHostData:], r.HostData[:])
	copy(buf[offsetIDKeyDigest:], r.IDKeyDigest[:])
	copy(buf[offsetAuthorKeyDigest:], r.AuthorKeyDigest[:])

	// Identifiers
	copy(buf[offsetReportID:], r.ReportID[:])
	copy(buf[offsetReportIDMA:], r.ReportIDMA[:])
	binary.LittleEndian.PutUint64(buf[offsetReportedTCB:], r.ReportedTCB.ToUint64())

	// Chip ID
	copy(buf[offsetChipID:], r.ChipID[:])

	// Additional TCB
	binary.LittleEndian.PutUint64(buf[offsetCommittedTCB:], r.CommittedTCB.ToUint64())
	buf[offsetCurrentBuild] = r.CurrentBuild
	buf[offsetCurrentVersion] = r.CurrentVersion
	binary.LittleEndian.PutUint64(buf[offsetLaunchTCB:], r.LaunchTCB.ToUint64())

	// Signature
	copy(buf[offsetSignature:], r.Signature[:])

	return buf, nil
}

// ValidateReport performs validation of the attestation report structure
//
// This validates the report format and policy, but does NOT verify the
// cryptographic signature. Use VerifyReportSignature for that.
func ValidateReport(r *AttestationReport) error {
	if r == nil {
		return errors.New("nil report")
	}

	// Check version
	if r.Version < ReportVersionV2 {
		return fmt.Errorf("unsupported report version: %d (minimum: %d)", r.Version, ReportVersionV2)
	}

	// Check signature algorithm
	if r.SignatureAlgo != SigAlgoECDSAP384SHA384 && r.SignatureAlgo != 0 {
		return fmt.Errorf("unsupported signature algorithm: %d", r.SignatureAlgo)
	}

	// Check VMPL
	if r.VMPL > VMPL3 {
		return fmt.Errorf("invalid VMPL: %d", r.VMPL)
	}

	// Validate policy
	if err := r.Policy.Validate(); err != nil {
		return fmt.Errorf("invalid policy: %w", err)
	}

	// Check for non-zero chip ID (indicates valid hardware)
	var zeroChipID [ChipIDSize]byte
	if bytes.Equal(r.ChipID[:], zeroChipID[:]) {
		return errors.New("chip ID is all zeros")
	}

	// Check for non-zero launch digest
	var zeroDigest [LaunchDigestSize]byte
	if bytes.Equal(r.LaunchDigest[:], zeroDigest[:]) {
		return errors.New("launch digest is all zeros")
	}

	return nil
}

// =============================================================================
// Report Data Helpers
// =============================================================================

// GetSignatureR returns the R component of the ECDSA signature (48 bytes)
func (r *AttestationReport) GetSignatureR() []byte {
	return r.Signature[0:48]
}

// GetSignatureS returns the S component of the ECDSA signature (48 bytes)
func (r *AttestationReport) GetSignatureS() []byte {
	return r.Signature[48:96]
}

// GetSignedData returns the data that is signed (everything before signature)
func (r *AttestationReport) GetSignedData() ([]byte, error) {
	data, err := SerializeReport(r)
	if err != nil {
		return nil, err
	}
	return data[:offsetSignature], nil
}

// SMTEnabled returns true if SMT is enabled on the platform
func (r *AttestationReport) SMTEnabled() bool {
	return r.PlatformInfo&0x01 != 0
}

// TSMEEnabled returns true if TSME is enabled
func (r *AttestationReport) TSMEEnabled() bool {
	return r.PlatformInfo&0x02 != 0
}

// IsDebug returns true if the guest is in debug mode
func (r *AttestationReport) IsDebug() bool {
	return r.Policy.Debug
}

// =============================================================================
// TCB Comparison Helpers
// =============================================================================

// TCBMeetsMinimum checks if the reported TCB meets minimum requirements
func (r *AttestationReport) TCBMeetsMinimum(minTCB TCBVersion) bool {
	return r.ReportedTCB.MeetsMinimum(minTCB)
}

// GetTCBComponents returns individual TCB component values for verification
func (r *AttestationReport) GetTCBComponents() (blSPL, teeSPL, snpSPL, ucodeSPL uint8) {
	return r.ReportedTCB.BootLoader,
		r.ReportedTCB.TEE,
		r.ReportedTCB.SNP,
		r.ReportedTCB.Microcode
}

// =============================================================================
// Extended Report
// =============================================================================

// ExtendedReportRequest is the input for SNP_GET_EXT_REPORT ioctl
type ExtendedReportRequest struct {
	// Base report request
	ReportRequest

	// CertsSize is the size of the certificate buffer
	CertsSize uint32

	// CertsAddr is the address of the certificate buffer
	CertsAddr uint64
}

// ExtendedReport contains an attestation report with certificate chain
type ExtendedReport struct {
	// Report is the parsed attestation report
	Report *AttestationReport

	// VCEKCert is the Versioned Chip Endorsement Key certificate (DER)
	VCEKCert []byte

	// ASKCert is the AMD SEV Signing Key certificate (DER)
	ASKCert []byte

	// ARKCert is the AMD Root Key certificate (DER)
	ARKCert []byte
}

// GetCertificateChain returns the certificate chain as a slice
// Order: VCEK, ASK, ARK (leaf to root)
func (e *ExtendedReport) GetCertificateChain() [][]byte {
	return [][]byte{e.VCEKCert, e.ASKCert, e.ARKCert}
}
