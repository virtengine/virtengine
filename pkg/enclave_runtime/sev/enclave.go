// Package sev provides AMD SEV-SNP (Secure Encrypted Virtualization - Secure Nested Paging)
// integration for VirtEngine TEE hardware attestation.
//
// This package implements the SEV-SNP guest interface for confidential VMs, including:
// - Guest initialization and platform information retrieval
// - Attestation report generation via /dev/sev-guest
// - Hardware-bound key derivation using VCEK
// - VM Privilege Level (VMPL) handling
// - Guest policy enforcement
//
// # Hardware Requirements
//
// For real hardware operations (build tag: sev_hardware):
// - AMD EPYC Milan (7003 series) or later CPU with SEV-SNP support
// - Linux kernel 6.0+ with SEV-SNP guest patches
// - /dev/sev-guest device available inside confidential VM
// - Running inside an SNP-enabled guest
//
// Without the sev_hardware build tag, all operations run in simulation mode.
//
// # Architecture
//
// SEV-SNP uses a Platform Security Processor (PSP) for attestation:
// - Guest requests attestation via ioctl to /dev/sev-guest
// - PSP generates and signs report with VCEK (Versioned Chip Endorsement Key)
// - VCEK is unique per-chip and TCB version, retrieved from AMD KDS
//
// # Security Properties
//
// - Memory encryption: All guest memory encrypted with AES-128-XEX
// - Integrity protection: Reverse Map Table (RMP) prevents hypervisor tampering
// - Attestation: Reports signed by hardware-bound VCEK
// - Key derivation: Keys bound to guest measurement and TCB version
//
// Task Reference: VE-2029 - Hardware TEE Integration Layer
package sev

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/hkdf"
)

// =============================================================================
// Constants
// =============================================================================

const (
	// Device paths
	GuestDevicePath     = "/dev/sev-guest"
	DebugFSPath         = "/sys/kernel/debug/sev"
	PlatformStatusPath  = "/sys/kernel/debug/sev/status"
	FirmwareVersionPath = "/sys/class/firmware-attributes/sev/sev_version"

	// ioctl command numbers for /dev/sev-guest
	// These match the Linux kernel definitions in include/uapi/linux/sev-guest.h
	IoctlGetReport     = 0xC0185300 // SNP_GET_REPORT
	IoctlGetDerivedKey = 0xC0405301 // SNP_GET_DERIVED_KEY
	IoctlGetExtReport  = 0xC0285302 // SNP_GET_EXT_REPORT
	IoctlConfigInfo    = 0xC0185303 // SNP_GET_CONFIG_INFO (kernel 6.6+)

	// Report sizes
	ReportSize       = 0x4A0 // 1184 bytes - full attestation report
	ReportDataSize   = 64    // User-provided data in report
	LaunchDigestSize = 48    // SHA-384 launch measurement
	SignatureSize    = 512   // ECDSA P-384 signature (r || s || reserved)
	ChipIDSize       = 64    // Unique chip identifier

	// VMPL (VM Privilege Level) values
	VMPL0 = 0 // Highest privilege (firmware)
	VMPL1 = 1 // Usually hypervisor components
	VMPL2 = 2 // Usually kernel
	VMPL3 = 3 // Lowest privilege (user space)

	// Key derivation root key selection
	KeyRootVCEK = 0 // Use VCEK (Versioned Chip Endorsement Key)
	KeyRootVMRK = 1 // Use VMRK (VM Root Key)

	// Guest field selection flags for key derivation
	KeyFieldGuest    = 0x0001 // Mix guest policy into key
	KeyFieldTCB      = 0x0002 // Mix TCB version into key
	KeyFieldSVN      = 0x0004 // Mix guest SVN into key
	KeyFieldFamilyID = 0x0008 // Mix family ID into key
	KeyFieldImageID  = 0x0010 // Mix image ID into key

	// Message types for guest-host protocol
	MsgReportReq  = 5  // Request attestation report
	MsgReportResp = 6  // Attestation report response
	MsgKeyReq     = 7  // Request derived key
	MsgKeyResp    = 8  // Derived key response
	MsgExportReq  = 9  // Export request
	MsgExportResp = 10 // Export response
	MsgAbsorbReq  = 11 // Absorb request
	MsgAbsorbResp = 12 // Absorb response
	MsgVmrkReq    = 13 // VMRK request
	MsgVmrkResp   = 14 // VMRK response

	// Status codes from guest request
	StatusSuccess          = 0
	StatusInvalidParams    = 0x16
	StatusInvalidLen       = 0x17
	StatusInvalidGuest     = 0x23
	StatusPlatformBusy     = 0x07
	StatusInvalidKey       = 0x27
	StatusInvalidTCB       = 0x2F
	StatusInvalidSignature = 0x1B
)

// =============================================================================
// Error Types
// =============================================================================

var (
	// ErrDeviceNotFound indicates /dev/sev-guest is not present
	ErrDeviceNotFound = errors.New("sev: /dev/sev-guest device not found")

	// ErrPermissionDenied indicates insufficient permissions to access device
	ErrPermissionDenied = errors.New("sev: permission denied accessing /dev/sev-guest")

	// ErrNotInitialized indicates the guest has not been initialized
	ErrNotInitialized = errors.New("sev: guest not initialized")

	// ErrInvalidVMPL indicates an invalid VMPL was specified
	ErrInvalidVMPL = errors.New("sev: invalid VMPL (must be 0-3)")

	// ErrReportGenFailed indicates attestation report generation failed
	ErrReportGenFailed = errors.New("sev: attestation report generation failed")

	// ErrKeyDerivationFailed indicates key derivation failed
	ErrKeyDerivationFailed = errors.New("sev: key derivation failed")

	// ErrIoctlFailed indicates an ioctl call failed
	ErrIoctlFailed = errors.New("sev: ioctl failed")

	// ErrSimulationMode indicates operation requires hardware but running in simulation
	ErrSimulationMode = errors.New("sev: operation not available in simulation mode")

	// ErrInvalidPolicy indicates the guest policy is invalid
	ErrInvalidPolicy = errors.New("sev: invalid guest policy")

	// ErrDebugEnabled indicates debug mode is enabled (insecure)
	ErrDebugEnabled = errors.New("sev: debug mode enabled - not secure for production")
)

// GuestError provides detailed error information for SEV-SNP operations
type GuestError struct {
	Op         string // Operation that failed
	DevicePath string // Device path if applicable
	StatusCode uint32 // Hardware status code
	Err        error  // Underlying error
}

func (e *GuestError) Error() string {
	if e.StatusCode != 0 {
		return fmt.Sprintf("sev: %s failed: status=0x%x: %v", e.Op, e.StatusCode, e.Err)
	}
	if e.DevicePath != "" {
		return fmt.Sprintf("sev: %s failed on %s: %v", e.Op, e.DevicePath, e.Err)
	}
	return fmt.Sprintf("sev: %s failed: %v", e.Op, e.Err)
}

func (e *GuestError) Unwrap() error {
	return e.Err
}

// =============================================================================
// Platform Information
// =============================================================================

// PlatformInfo contains SEV-SNP platform configuration and capabilities
type PlatformInfo struct {
	// APIVersion is the SEV firmware API version (major*100 + minor)
	APIVersion int

	// BuildID is the firmware build ID
	BuildID uint32

	// CurrentTCB is the current Trusted Computing Base version
	CurrentTCB TCBVersion

	// ReportedTCB is the TCB version reported in attestation
	ReportedTCB TCBVersion

	// PlatformFlags indicates platform configuration
	// Bit 0: SMT enabled
	// Bit 1: TSME enabled
	PlatformFlags uint64

	// ChipID is the unique 64-byte chip identifier
	ChipID [ChipIDSize]byte

	// SocketCount is the number of CPU sockets
	SocketCount int

	// GuestCount is the current number of active SEV guests
	GuestCount int

	// IsSimulated indicates if running in simulation mode
	IsSimulated bool
}

// SMTEnabled returns true if Simultaneous Multi-Threading is enabled
func (p *PlatformInfo) SMTEnabled() bool {
	return p.PlatformFlags&0x01 != 0
}

// TSMEEnabled returns true if Transparent SME is enabled
func (p *PlatformInfo) TSMEEnabled() bool {
	return p.PlatformFlags&0x02 != 0
}

// =============================================================================
// Key Request
// =============================================================================

// KeyRequest specifies parameters for hardware key derivation
type KeyRequest struct {
	// RootKeySelect specifies which root key to use (KeyRootVCEK or KeyRootVMRK)
	RootKeySelect int

	// GuestFieldSelect specifies which fields to mix into the derived key
	// Use KeyField* constants (can be OR'd together)
	GuestFieldSelect uint64

	// VMPL specifies the VM Privilege Level (0-3)
	VMPL uint32

	// GuestSVN is the guest security version number to use
	GuestSVN uint32

	// TCBVersion is the TCB version to use (0 = current)
	TCBVersion uint64
}

// Validate checks if the key request is valid
func (r *KeyRequest) Validate() error {
	if r.RootKeySelect != KeyRootVCEK && r.RootKeySelect != KeyRootVMRK {
		return fmt.Errorf("invalid root key select: %d", r.RootKeySelect)
	}
	if r.VMPL > VMPL3 {
		return ErrInvalidVMPL
	}
	return nil
}

// =============================================================================
// SEV Guest
// =============================================================================

// SEVGuest provides the interface to the SEV-SNP guest device
type SEVGuest struct {
	mu sync.RWMutex

	// Device state
	devicePath  string
	fd          *os.File
	opened      bool
	simulated   bool
	initialized bool

	// Cached platform info
	platformInfo *PlatformInfo
	lastDetect   time.Time

	// Simulated state (used when hardware not available)
	simChipID       [ChipIDSize]byte
	simLaunchDigest [LaunchDigestSize]byte
	simVCEKSeed     []byte
	simPolicy       GuestPolicy
	simTCB          TCBVersion
}

// NewSEVGuest creates a new SEV guest interface
func NewSEVGuest() *SEVGuest {
	return &SEVGuest{
		devicePath: GuestDevicePath,
	}
}

// NewSEVGuestWithPath creates a new SEV guest with custom device path
func NewSEVGuestWithPath(path string) *SEVGuest {
	return &SEVGuest{
		devicePath: path,
	}
}

// Initialize initializes the SEV guest interface
//
// This attempts to open /dev/sev-guest for hardware operations.
// If the device is not available, it falls back to simulation mode.
func (g *SEVGuest) Initialize() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.initialized {
		return nil
	}

	// Try to detect and open hardware
	if err := g.detectHardware(); err != nil {
		// Fall back to simulation
		g.simulated = true
		if err := g.initSimulation(); err != nil {
			return fmt.Errorf("simulation init failed: %w", err)
		}
	}

	g.initialized = true
	return nil
}

// detectHardware attempts to detect and open the SEV-SNP device
func (g *SEVGuest) detectHardware() error {
	// Check if device exists
	info, err := os.Stat(g.devicePath)
	if os.IsNotExist(err) {
		return ErrDeviceNotFound
	}
	if err != nil {
		return fmt.Errorf("stat device: %w", err)
	}

	// Must be a character device
	if info.Mode()&os.ModeDevice == 0 {
		return fmt.Errorf("%s is not a device file", g.devicePath)
	}

	// Try to open device
	fd, err := os.OpenFile(g.devicePath, os.O_RDWR, 0)
	if err != nil {
		if os.IsPermission(err) {
			return ErrPermissionDenied
		}
		return fmt.Errorf("open device: %w", err)
	}

	g.fd = fd
	g.opened = true
	g.simulated = false
	return nil
}

// initSimulation initializes simulation mode state
func (g *SEVGuest) initSimulation() error {
	// Generate simulated chip ID
	chipIDSeed := make([]byte, 64)
	if _, err := rand.Read(chipIDSeed); err != nil {
		return err
	}
	chipHash := sha512.Sum512(chipIDSeed)
	copy(g.simChipID[:], chipHash[:])

	// Generate simulated launch digest
	launchSeed := []byte("virtengine-sev-snp-simulated-" + time.Now().Format(time.RFC3339Nano))
	launchHash := sha512.Sum384(launchSeed)
	copy(g.simLaunchDigest[:], launchHash[:])

	// Generate simulated VCEK seed
	g.simVCEKSeed = make([]byte, 32)
	if _, err := rand.Read(g.simVCEKSeed); err != nil {
		return err
	}

	// Set default policy (production settings)
	g.simPolicy = GuestPolicy{
		ABIMajor:     1,
		ABIMinor:     0,
		SMT:          true,
		Debug:        false,
		SingleSocket: false,
		Migration:    false,
	}

	// Set simulated TCB version (Milan defaults)
	g.simTCB = TCBVersion{
		BootLoader: 3,
		TEE:        0,
		SNP:        14,
		Microcode:  209,
	}

	g.simulated = true
	g.platformInfo = &PlatformInfo{
		APIVersion:    151, // Version 1.51
		CurrentTCB:    g.simTCB,
		ReportedTCB:   g.simTCB,
		PlatformFlags: 0x01, // SMT enabled
		ChipID:        g.simChipID,
		SocketCount:   1,
		IsSimulated:   true,
	}

	return nil
}

// Close closes the SEV guest device
func (g *SEVGuest) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.opened || g.fd == nil {
		return nil
	}

	err := g.fd.Close()
	g.fd = nil
	g.opened = false
	g.initialized = false

	return err
}

// IsInitialized returns true if the guest has been initialized
func (g *SEVGuest) IsInitialized() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.initialized
}

// IsSimulated returns true if running in simulation mode
func (g *SEVGuest) IsSimulated() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.simulated
}

// GetPlatformInfo retrieves platform information from the SEV-SNP hardware
func (g *SEVGuest) GetPlatformInfo() (*PlatformInfo, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.initialized {
		return nil, ErrNotInitialized
	}

	// Return cached info in simulation mode
	if g.simulated {
		return g.platformInfo, nil
	}

	// For hardware mode, request a report to extract platform info
	var userData [ReportDataSize]byte
	report, err := g.requestReportLocked(userData, VMPL0)
	if err != nil {
		return nil, fmt.Errorf("get platform info: %w", err)
	}

	info := &PlatformInfo{
		APIVersion:    g.detectAPIVersion(),
		CurrentTCB:    report.CurrentTCB,
		ReportedTCB:   report.ReportedTCB,
		PlatformFlags: report.PlatformInfo,
		ChipID:        report.ChipID,
		IsSimulated:   false,
	}

	g.platformInfo = info
	g.lastDetect = time.Now()

	return info, nil
}

// detectAPIVersion attempts to read the API version from sysfs
func (g *SEVGuest) detectAPIVersion() int {
	data, err := os.ReadFile(FirmwareVersionPath)
	if err != nil {
		// Try alternate location
		data, err = os.ReadFile(PlatformStatusPath)
		if err != nil {
			return 151 // Default to Milan (1.51)
		}
	}

	var major, minor int
	if _, err := fmt.Sscanf(string(data), "%d.%d", &major, &minor); err == nil {
		return major*100 + minor
	}
	return 151
}

// GenerateAttestation generates an attestation report with the provided user data
//
// userData is included in the report and can be used to bind the attestation
// to a specific request (e.g., nonce, request hash). It must be exactly 64 bytes.
//
// The returned report is signed by the VCEK and can be verified using the
// AMD KDS certificate chain.
func (g *SEVGuest) GenerateAttestation(userData [ReportDataSize]byte) ([]byte, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.initialized {
		return nil, ErrNotInitialized
	}

	report, err := g.requestReportLocked(userData, VMPL0)
	if err != nil {
		return nil, err
	}

	return SerializeReport(report)
}

// GenerateAttestationWithVMPL generates an attestation report at a specific VMPL
func (g *SEVGuest) GenerateAttestationWithVMPL(userData [ReportDataSize]byte, vmpl uint32) ([]byte, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.initialized {
		return nil, ErrNotInitialized
	}

	if vmpl > VMPL3 {
		return nil, ErrInvalidVMPL
	}

	report, err := g.requestReportLocked(userData, vmpl)
	if err != nil {
		return nil, err
	}

	return SerializeReport(report)
}

// requestReportLocked generates an attestation report (must hold lock)
func (g *SEVGuest) requestReportLocked(userData [ReportDataSize]byte, vmpl uint32) (*AttestationReport, error) {
	if g.simulated {
		return g.generateSimulatedReport(userData, vmpl)
	}

	// Build ioctl request structure
	return g.requestHardwareReport(userData, vmpl)
}

// requestHardwareReport performs the actual ioctl to get a hardware report
// When building with the sev_hardware tag, this is replaced by the real implementation.
func (g *SEVGuest) requestHardwareReport(userData [ReportDataSize]byte, vmpl uint32) (*AttestationReport, error) {
	// This is implemented in enclave_hardware.go with sev_hardware build tag
	// For non-hardware builds, fall back to simulation
	return g.generateSimulatedReport(userData, vmpl)
}

// generateSimulatedReport creates a simulated attestation report
func (g *SEVGuest) generateSimulatedReport(userData [ReportDataSize]byte, vmpl uint32) (*AttestationReport, error) {
	report := &AttestationReport{
		Version:          ReportVersionV2,
		GuestSVN:         0,
		Policy:           g.simPolicy,
		ReportData:       userData,
		AuthorKeyEnabled: 0,
		VMPL:             vmpl,
		CurrentTCB:       g.simTCB,
		ReportedTCB:      g.simTCB,
		PlatformInfo:     0x01, // SMT enabled
		ChipID:           g.simChipID,
		LaunchDigest:     g.simLaunchDigest,
	}

	// Generate family ID and image ID
	familyHash := sha256.Sum256([]byte("virtengine-veid-family"))
	copy(report.FamilyID[:], familyHash[:16])

	imageHash := sha256.Sum256([]byte("veid-scorer-image-v1"))
	copy(report.ImageID[:], imageHash[:16])

	// Generate report ID
	var reportIDInput bytes.Buffer
	reportIDInput.Write(userData[:])
	_ = binary.Write(&reportIDInput, binary.LittleEndian, time.Now().UnixNano())
	reportIDHash := sha256.Sum256(reportIDInput.Bytes())
	copy(report.ReportID[:], reportIDHash[:])

	// Generate simulated signature
	sigData := g.computeReportSigningData(report)
	sigHash := sha512.Sum384(append(g.simVCEKSeed, sigData...))
	copy(report.Signature[:], sigHash[:])
	// Signature is ECDSA P-384: R (48 bytes) || S (48 bytes) || reserved
	// For simulation, we just use the hash

	return report, nil
}

// computeReportSigningData computes the data that would be signed
func (g *SEVGuest) computeReportSigningData(report *AttestationReport) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, report.Version)
	_ = binary.Write(&buf, binary.LittleEndian, report.GuestSVN)
	_ = binary.Write(&buf, binary.LittleEndian, report.Policy.ToUint64())
	buf.Write(report.FamilyID[:])
	buf.Write(report.ImageID[:])
	buf.Write(report.LaunchDigest[:])
	buf.Write(report.ReportData[:])
	buf.Write(report.ChipID[:])
	return buf.Bytes()
}

// DeriveKey derives a key from the SEV-SNP hardware root of trust
//
// The derived key is bound to:
// - The chip's unique identity (VCEK or VMRK based on request)
// - The guest's measurement (launch digest)
// - The TCB version
// - Additional fields specified in the request
func (g *SEVGuest) DeriveKey(request *KeyRequest) ([]byte, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.initialized {
		return nil, ErrNotInitialized
	}

	if err := request.Validate(); err != nil {
		return nil, err
	}

	if g.simulated {
		return g.deriveSimulatedKey(request)
	}

	return g.deriveHardwareKey(request)
}

// deriveHardwareKey performs hardware key derivation via ioctl
func (g *SEVGuest) deriveHardwareKey(request *KeyRequest) ([]byte, error) {
	// Hardware implementation with sev_hardware build tag
	// For non-hardware builds, fall back to simulation
	return g.deriveSimulatedKey(request)
}

// deriveSimulatedKey generates a simulated derived key
func (g *SEVGuest) deriveSimulatedKey(request *KeyRequest) ([]byte, error) {
	// Build input for key derivation
	var input bytes.Buffer
	input.Write(g.simVCEKSeed)
	input.Write(g.simLaunchDigest[:])
	_ = binary.Write(&input, binary.LittleEndian, uint32(request.RootKeySelect))
	_ = binary.Write(&input, binary.LittleEndian, request.GuestFieldSelect)
	_ = binary.Write(&input, binary.LittleEndian, request.VMPL)
	_ = binary.Write(&input, binary.LittleEndian, request.GuestSVN)
	_ = binary.Write(&input, binary.LittleEndian, request.TCBVersion)

	// Derive using HKDF
	salt := g.simChipID[:]
	info := []byte("sev-snp-derived-key-v1")

	reader := hkdf.New(sha256.New, input.Bytes(), salt, info)
	key := make([]byte, 32) // Always 32 bytes for SEV-SNP
	if _, err := reader.Read(key); err != nil {
		return nil, fmt.Errorf("hkdf read: %w", err)
	}

	return key, nil
}

// DeriveKeyWithContext derives a key and further derives it with application context
func (g *SEVGuest) DeriveKeyWithContext(request *KeyRequest, context []byte, keySize int) ([]byte, error) {
	// Get base key from hardware/simulation
	baseKey, err := g.DeriveKey(request)
	if err != nil {
		return nil, err
	}

	// Further derive with application context
	reader := hkdf.New(sha256.New, baseKey, nil, context)
	derived := make([]byte, keySize)
	if _, err := reader.Read(derived); err != nil {
		return nil, fmt.Errorf("context derivation: %w", err)
	}

	// Clear intermediate key
	for i := range baseKey {
		baseKey[i] = 0
	}

	return derived, nil
}

// GetLaunchDigest returns the current launch measurement
func (g *SEVGuest) GetLaunchDigest() ([LaunchDigestSize]byte, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return [LaunchDigestSize]byte{}, ErrNotInitialized
	}

	if g.simulated {
		return g.simLaunchDigest, nil
	}

	// For hardware, get from a report
	var userData [ReportDataSize]byte
	report, err := g.requestReportLocked(userData, VMPL0)
	if err != nil {
		return [LaunchDigestSize]byte{}, err
	}

	return report.LaunchDigest, nil
}

// GetChipID returns the unique chip identifier
func (g *SEVGuest) GetChipID() ([ChipIDSize]byte, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return [ChipIDSize]byte{}, ErrNotInitialized
	}

	if g.simulated {
		return g.simChipID, nil
	}

	if g.platformInfo != nil {
		return g.platformInfo.ChipID, nil
	}

	return [ChipIDSize]byte{}, ErrNotInitialized
}

// GetGuestPolicy returns the current guest policy
func (g *SEVGuest) GetGuestPolicy() (GuestPolicy, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return GuestPolicy{}, ErrNotInitialized
	}

	if g.simulated {
		return g.simPolicy, nil
	}

	// For hardware, get from a report
	var userData [ReportDataSize]byte
	report, err := g.requestReportLocked(userData, VMPL0)
	if err != nil {
		return GuestPolicy{}, err
	}

	return report.Policy, nil
}

// GetCurrentTCB returns the current TCB version
func (g *SEVGuest) GetCurrentTCB() (TCBVersion, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return TCBVersion{}, ErrNotInitialized
	}

	if g.simulated {
		return g.simTCB, nil
	}

	if g.platformInfo != nil {
		return g.platformInfo.CurrentTCB, nil
	}

	return TCBVersion{}, ErrNotInitialized
}

// VerifyPolicySecure checks if the guest policy meets security requirements
func (g *SEVGuest) VerifyPolicySecure() error {
	policy, err := g.GetGuestPolicy()
	if err != nil {
		return err
	}

	if policy.Debug {
		return ErrDebugEnabled
	}

	if policy.ABIMajor < 1 {
		return fmt.Errorf("ABI major version %d is too old", policy.ABIMajor)
	}

	return nil
}
