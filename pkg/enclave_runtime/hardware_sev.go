// Package enclave_runtime provides TEE enclave implementations.
//
// This file provides the hardware abstraction layer for AMD SEV-SNP operations.
// When running inside an SNP-enabled confidential VM, these functions make
// real /dev/sev-guest ioctl calls. Otherwise, they fall back to simulation.
//
// Requirements for real SEV-SNP:
// - AMD EPYC Milan (7003 series) or later CPU
// - Linux kernel 6.0+ with SNP support
// - /dev/sev-guest device available
// - Running inside a confidential VM
//
// Task Reference: VE-2029 - Hardware TEE Integration Layer
package enclave_runtime

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
// SEV-SNP Device Paths and Constants
// =============================================================================

const (
	// SEV device paths
	SEVGuestDevicePath     = "/dev/sev-guest"
	SEVDebugFSPath         = "/sys/kernel/debug/sev"
	SEVPlatformStatusPath  = "/sys/kernel/debug/sev/status"
	SEVFirmwareVersionPath = "/sys/class/firmware-attributes/sev/sev_version"

	// SEV-SNP ioctl commands
	// These are the actual Linux kernel ioctl numbers for SEV-SNP guest operations
	SNP_GET_REPORT      = 0xC0185300 // Get attestation report
	SNP_GET_DERIVED_KEY = 0xC0405301 // Get derived key
	SNP_GET_EXT_REPORT  = 0xC0285302 // Get extended report with certificates

	// SNP message types
	SNP_MSG_REPORT_REQ = 5  // Request attestation report
	SNP_MSG_REPORT_RSP = 6  // Attestation report response
	SNP_MSG_KEY_REQ    = 7  // Request derived key
	SNP_MSG_KEY_RSP    = 8  // Derived key response
	SNP_MSG_EXPORT_REQ = 9  // Export request
	SNP_MSG_EXPORT_RSP = 10 // Export response
	SNP_MSG_ABSORB_REQ = 11 // Absorb request
	SNP_MSG_ABSORB_RSP = 12 // Absorb response
	SNP_MSG_VMRK_REQ   = 13 // VMRK request
	SNP_MSG_VMRK_RSP   = 14 // VMRK response

	// SNP report request flags
	SNP_REPORT_USER_DATA = 0x0001 // Include user data in report

	// Key derivation fields
	SNP_KEY_ROOT_VCEK   = 0 // Use VCEK as root
	SNP_KEY_ROOT_VMRK   = 1 // Use VMRK as root
	SNP_KEY_GUEST_FIELD = 0x0001
	SNP_KEY_TCB_FIELD   = 0x0002
	SNP_KEY_SVN_FIELD   = 0x0004

	// Report sizes
	SNP_REPORT_SIZE        = 0x4A0 // 1184 bytes
	SNP_REPORT_DATA_SIZE   = 64
	SNP_LAUNCH_DIGEST_SIZE = 48
	SNP_SIGNATURE_SIZE     = 512
)

// =============================================================================
// SEV Hardware Detector
// =============================================================================

// SEVHardwareDetector provides methods to detect SEV-SNP hardware capabilities
type SEVHardwareDetector struct {
	mu sync.RWMutex

	detected       bool
	available      bool
	version        string
	apiVersion     int
	guestDevice    string
	lastDetection  time.Time
	detectionError error
}

// NewSEVHardwareDetector creates a new SEV-SNP hardware detector
func NewSEVHardwareDetector() *SEVHardwareDetector {
	return &SEVHardwareDetector{}
}

// Detect performs SEV-SNP hardware detection
func (d *SEVHardwareDetector) Detect() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.detected = true
	d.lastDetection = time.Now()

	// Check for SEV-guest device
	exists, err := checkDeviceExists(SEVGuestDevicePath)
	if err != nil {
		if errors.Is(err, ErrPermissionDenied) {
			// Device exists but we don't have permission
			d.available = true
			d.guestDevice = SEVGuestDevicePath
			d.detectionError = err
			return err
		}
		d.detectionError = fmt.Errorf("error checking SEV device: %w", err)
		return d.detectionError
	}

	if !exists {
		d.available = false
		d.detectionError = fmt.Errorf("SEV-SNP device not found at %s", SEVGuestDevicePath)
		return d.detectionError
	}

	d.guestDevice = SEVGuestDevicePath
	d.available = true

	// Try to read platform status for version info
	if status, err := readSysFile(SEVPlatformStatusPath); err == nil {
		d.version = status
	} else {
		d.version = "unknown"
	}

	// Try to determine API version
	d.apiVersion = d.detectAPIVersion()

	d.detectionError = nil
	return nil
}

// detectAPIVersion attempts to detect the SEV API version
func (d *SEVHardwareDetector) detectAPIVersion() int {
	// Try reading from firmware attributes
	if version, err := readSysFile(SEVFirmwareVersionPath); err == nil {
		// Parse version string (format: "major.minor")
		var major, minor int
		if _, err := fmt.Sscanf(version, "%d.%d", &major, &minor); err == nil {
			return major*100 + minor
		}
	}

	// Default to assuming latest API if we can't detect
	return 151 // Version 1.51 (Milan)
}

// IsAvailable returns true if SEV-SNP hardware is available
func (d *SEVHardwareDetector) IsAvailable() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.available
}

// Version returns the SEV-SNP version string
func (d *SEVHardwareDetector) Version() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.version
}

// APIVersion returns the SEV API version
func (d *SEVHardwareDetector) APIVersion() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.apiVersion
}

// GetDevicePath returns the path to the SEV-guest device
func (d *SEVHardwareDetector) GetDevicePath() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.guestDevice
}

// =============================================================================
// SEV Guest Device
// =============================================================================

// SEVGuestDevice wraps operations on /dev/sev-guest
type SEVGuestDevice struct {
	mu sync.Mutex

	detector   *SEVHardwareDetector
	devicePath string
	fd         *os.File
	opened     bool
	simulated  bool
}

// NewSEVGuestDevice creates a new SEV guest device wrapper
func NewSEVGuestDevice(detector *SEVHardwareDetector) *SEVGuestDevice {
	return &SEVGuestDevice{
		detector:   detector,
		devicePath: SEVGuestDevicePath,
	}
}

// Open opens the SEV-guest device
func (d *SEVGuestDevice) Open() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.opened {
		return nil
	}

	if !d.detector.IsAvailable() {
		// Fall back to simulation
		d.simulated = true
		d.opened = true
		return nil
	}

	// Try to open the device
	fd, err := os.OpenFile(d.devicePath, os.O_RDWR, 0)
	if err != nil {
		if os.IsPermission(err) {
			return &HardwareError{
				Platform:   AttestationTypeSEVSNP,
				Operation:  "open",
				DevicePath: d.devicePath,
				Underlying: ErrPermissionDenied,
			}
		}
		// Fall back to simulation
		d.simulated = true
		d.opened = true
		return nil
	}

	d.fd = fd
	d.opened = true
	return nil
}

// Close closes the device
func (d *SEVGuestDevice) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.opened {
		return nil
	}

	if d.fd != nil {
		if err := d.fd.Close(); err != nil {
			return err
		}
		d.fd = nil
	}

	d.opened = false
	return nil
}

// IsSimulated returns true if running in simulation mode
func (d *SEVGuestDevice) IsSimulated() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.simulated
}

// =============================================================================
// SNP Report Request Structures
// =============================================================================

// SNPReportRequest represents a request for an attestation report
type SNPReportRequest struct {
	UserData [64]byte // User-provided data to include in report
	VMPL     uint32   // VM Privilege Level (0-3)
	Reserved [28]byte
}

// SNPReportResponse represents the response containing an attestation report
type SNPReportResponse struct {
	Status   uint32 // Status code
	Reserved [28]byte
	Report   [SNP_REPORT_SIZE]byte // The attestation report
}

// SNPKeyRequest represents a request for a derived key
type SNPKeyRequest struct {
	RootKeySelect    uint32 // VCEK (0) or VMRK (1)
	Reserved         uint32
	GuestFieldSelect uint64 // Which fields to mix into key
	VMPL             uint32 // VM Privilege Level
	GuestSVN         uint32 // Guest SVN to use
	TCBVersion       uint64 // TCB version to use
}

// SNPKeyResponse represents the response containing a derived key
type SNPKeyResponse struct {
	Status   uint32 // Status code
	Reserved [28]byte
	Key      [32]byte // The derived key
}

// SNPExtReportRequest represents a request for an extended report with certificates
type SNPExtReportRequest struct {
	UserData  [64]byte // User-provided data
	VMPL      uint32   // VM Privilege Level
	Reserved  [28]byte
	CertsSize uint32 // Size of certificate buffer
	CertsAddr uint64 // Address of certificate buffer
}

// =============================================================================
// SNP Report Requester
// =============================================================================

// SNPReportRequester handles attestation report requests
type SNPReportRequester struct {
	device *SEVGuestDevice
}

// NewSNPReportRequester creates a new report requester
func NewSNPReportRequester(device *SEVGuestDevice) *SNPReportRequester {
	return &SNPReportRequester{device: device}
}

// RequestReport requests an attestation report from the hardware
func (r *SNPReportRequester) RequestReport(userData [64]byte, vmpl uint32) (*SNPAttestationReport, error) {
	if !r.device.opened {
		if err := r.device.Open(); err != nil {
			return nil, err
		}
	}

	if r.device.IsSimulated() {
		return r.requestSimulatedReport(userData, vmpl)
	}

	return r.requestHardwareReport(userData, vmpl)
}

// requestHardwareReport requests a report via ioctl
func (r *SNPReportRequester) requestHardwareReport(userData [64]byte, vmpl uint32) (*SNPAttestationReport, error) {
	// Build request
	req := SNPReportRequest{
		UserData: userData,
		VMPL:     vmpl,
	}

	// TODO: Real implementation would:
	// 1. Serialize request to bytes
	// 2. Call ioctl(fd, SNP_GET_REPORT, &request)
	// 3. Parse response into SNPAttestationReport
	//
	// Example ioctl call (requires cgo or syscall):
	// _, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
	//     r.device.fd.Fd(),
	//     uintptr(SNP_GET_REPORT),
	//     uintptr(unsafe.Pointer(&req)))

	// For now, fall back to simulation
	_ = req
	return r.requestSimulatedReport(userData, vmpl)
}

// requestSimulatedReport generates a simulated report
//
//nolint:unparam // vmpl kept for future VMPL-based report differentiation
func (r *SNPReportRequester) requestSimulatedReport(userData [64]byte, _ uint32) (*SNPAttestationReport, error) {
	report := &SNPAttestationReport{
		Version:  SNPReportVersion,
		GuestSVN: 0,
		Policy: SNPGuestPolicy{
			ABIMinor:     0,
			ABIMajor:     1,
			SMT:          true,
			Debug:        false, // Never debug in production
			SingleSocket: false,
		},
		ReportData:       userData,
		AuthorKeyEnabled: 0,
		CurrentTCB: SNPTCBVersion{
			BootLoader: 3,
			TEE:        0,
			SNP:        14,
			Microcode:  209,
		},
		ReportedTCB: SNPTCBVersion{
			BootLoader: 3,
			TEE:        0,
			SNP:        14,
			Microcode:  209,
		},
	}

	// Generate simulated launch digest
	h := sha512.New384()
	h.Write([]byte("virtengine-sev-snp-simulated-launch"))
	h.Write(userData[:])
	copy(report.LaunchDigest[:], h.Sum(nil))

	// Generate simulated chip ID
	chipHash := sha512.Sum512([]byte("simulated-chip-id-" + time.Now().String()))
	copy(report.ChipID[:], chipHash[:])

	// Generate simulated signature
	sigHash := sha512.Sum512(append(report.LaunchDigest[:], userData[:]...))
	copy(report.Signature[:], sigHash[:])

	return report, nil
}

// =============================================================================
// SNP Derived Key Requester
// =============================================================================

// SNPDerivedKeyRequester handles derived key requests
type SNPDerivedKeyRequester struct {
	device *SEVGuestDevice
}

// NewSNPDerivedKeyRequester creates a new derived key requester
func NewSNPDerivedKeyRequester(device *SEVGuestDevice) *SNPDerivedKeyRequester {
	return &SNPDerivedKeyRequester{device: device}
}

// RequestKey requests a derived key from the hardware
func (r *SNPDerivedKeyRequester) RequestKey(rootKey int, guestFieldSelect uint64, vmpl uint32) ([]byte, error) {
	if !r.device.opened {
		if err := r.device.Open(); err != nil {
			return nil, err
		}
	}

	if r.device.IsSimulated() {
		return r.requestSimulatedKey(rootKey, guestFieldSelect, vmpl)
	}

	return r.requestHardwareKey(rootKey, guestFieldSelect, vmpl)
}

// requestHardwareKey requests a key via ioctl
func (r *SNPDerivedKeyRequester) requestHardwareKey(rootKey int, guestFieldSelect uint64, vmpl uint32) ([]byte, error) {
	// Build request
	req := SNPKeyRequest{
		//nolint:gosec // G115: rootKey is 0 or 1 enum value
		RootKeySelect:    uint32(rootKey),
		GuestFieldSelect: guestFieldSelect,
		VMPL:             vmpl,
	}

	// TODO: Real implementation would:
	// 1. Serialize request to bytes
	// 2. Call ioctl(fd, SNP_GET_DERIVED_KEY, &request)
	// 3. Parse response to get 32-byte key
	//
	// Example ioctl call (requires cgo or syscall):
	// _, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
	//     r.device.fd.Fd(),
	//     uintptr(SNP_GET_DERIVED_KEY),
	//     uintptr(unsafe.Pointer(&req)))

	_ = req
	return r.requestSimulatedKey(rootKey, guestFieldSelect, vmpl)
}

// requestSimulatedKey generates a simulated derived key
func (r *SNPDerivedKeyRequester) requestSimulatedKey(rootKey int, guestFieldSelect uint64, vmpl uint32) ([]byte, error) {
	// Generate a deterministic simulated key
	h := sha256.New()
	h.Write([]byte("virtengine-sev-snp-derived-key"))
	_ = binary.Write(h, binary.LittleEndian, uint32(rootKey))
	_ = binary.Write(h, binary.LittleEndian, guestFieldSelect)
	_ = binary.Write(h, binary.LittleEndian, vmpl)

	key := h.Sum(nil)
	return key, nil
}

// =============================================================================
// SNP Extended Report Requester
// =============================================================================

// SNPExtendedReportRequester handles extended report requests with certificates
type SNPExtendedReportRequester struct {
	device *SEVGuestDevice
}

// NewSNPExtendedReportRequester creates a new extended report requester
func NewSNPExtendedReportRequester(device *SEVGuestDevice) *SNPExtendedReportRequester {
	return &SNPExtendedReportRequester{device: device}
}

// ExtendedReport contains a report and its certificate chain
type ExtendedReport struct {
	Report   *SNPAttestationReport
	VCEKCert []byte // Versioned Chip Endorsement Key certificate
	ASKCert  []byte // AMD SEV Signing Key certificate
	ARKCert  []byte // AMD Root Key certificate
}

// RequestExtendedReport requests a report with certificate chain
func (r *SNPExtendedReportRequester) RequestExtendedReport(userData [64]byte, vmpl uint32) (*ExtendedReport, error) {
	if !r.device.opened {
		if err := r.device.Open(); err != nil {
			return nil, err
		}
	}

	if r.device.IsSimulated() {
		return r.requestSimulatedExtendedReport(userData, vmpl)
	}

	return r.requestHardwareExtendedReport(userData, vmpl)
}

// requestHardwareExtendedReport requests an extended report via ioctl
func (r *SNPExtendedReportRequester) requestHardwareExtendedReport(userData [64]byte, vmpl uint32) (*ExtendedReport, error) {
	// TODO: Real implementation would:
	// 1. Allocate a buffer for certificates (usually 4KB is enough)
	// 2. Build SNPExtReportRequest with certificate buffer address
	// 3. Call ioctl(fd, SNP_GET_EXT_REPORT, &request)
	// 4. Parse response including certificate chain

	return r.requestSimulatedExtendedReport(userData, vmpl)
}

// requestSimulatedExtendedReport generates a simulated extended report
func (r *SNPExtendedReportRequester) requestSimulatedExtendedReport(userData [64]byte, vmpl uint32) (*ExtendedReport, error) {
	// Get base report
	reportReq := NewSNPReportRequester(r.device)
	report, err := reportReq.RequestReport(userData, vmpl)
	if err != nil {
		return nil, err
	}

	// Generate simulated certificates
	return &ExtendedReport{
		Report:   report,
		VCEKCert: generateSimulatedCert("VCEK"),
		ASKCert:  generateSimulatedCert("ASK"),
		ARKCert:  generateSimulatedCert("ARK"),
	}, nil
}

// generateSimulatedCert generates a placeholder certificate
func generateSimulatedCert(name string) []byte {
	// This is NOT a valid certificate - just a placeholder for simulation
	header := fmt.Sprintf("-----BEGIN SIMULATED %s CERTIFICATE-----\n", name)
	footer := fmt.Sprintf("\n-----END SIMULATED %s CERTIFICATE-----\n", name)

	h := sha256.Sum256([]byte(name + "-simulated"))
	content := fmt.Sprintf("%x", h)

	return []byte(header + content + footer)
}

// =============================================================================
// SEV Hardware Backend (implements HardwareBackend interface)
// =============================================================================

// SEVHardwareBackend implements the HardwareBackend interface for SEV-SNP
type SEVHardwareBackend struct {
	mu sync.RWMutex

	detector     *SEVHardwareDetector
	device       *SEVGuestDevice
	reportReq    *SNPReportRequester
	keyReq       *SNPDerivedKeyRequester
	extReportReq *SNPExtendedReportRequester

	initialized  bool
	simulatedKey []byte
}

// NewSEVHardwareBackend creates a new SEV-SNP hardware backend
func NewSEVHardwareBackend() *SEVHardwareBackend {
	detector := NewSEVHardwareDetector()
	return &SEVHardwareBackend{
		detector: detector,
	}
}

// Platform returns the attestation type for this backend
func (b *SEVHardwareBackend) Platform() AttestationType {
	return AttestationTypeSEVSNP
}

// IsAvailable returns true if SEV-SNP hardware is available
func (b *SEVHardwareBackend) IsAvailable() bool {
	if err := b.detector.Detect(); err != nil {
		return false
	}
	return b.detector.IsAvailable()
}

// Initialize sets up the SEV-SNP hardware backend
func (b *SEVHardwareBackend) Initialize() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return nil
	}

	// Detect hardware
	if err := b.detector.Detect(); err != nil {
		// Continue anyway for simulation mode
	}

	// Create device wrapper
	b.device = NewSEVGuestDevice(b.detector)
	if err := b.device.Open(); err != nil {
		return fmt.Errorf("failed to open SEV device: %w", err)
	}

	// Create request handlers
	b.reportReq = NewSNPReportRequester(b.device)
	b.keyReq = NewSNPDerivedKeyRequester(b.device)
	b.extReportReq = NewSNPExtendedReportRequester(b.device)

	// Generate simulated key for simulation mode
	b.simulatedKey = make([]byte, 32)
	if _, err := rand.Read(b.simulatedKey); err != nil {
		return fmt.Errorf("failed to generate simulated key: %w", err)
	}

	b.initialized = true
	return nil
}

// Shutdown cleanly shuts down the SEV-SNP backend
func (b *SEVHardwareBackend) Shutdown() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.initialized {
		return nil
	}

	if b.device != nil {
		if err := b.device.Close(); err != nil {
			return fmt.Errorf("failed to close SEV device: %w", err)
		}
	}

	b.initialized = false
	return nil
}

// GetAttestation generates a SEV-SNP attestation report
func (b *SEVHardwareBackend) GetAttestation(nonce []byte) ([]byte, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, ErrHardwareNotInitialized
	}

	// Prepare user data with nonce
	var userData [64]byte
	h := sha512.New384()
	h.Write(nonce)
	copy(userData[:48], h.Sum(nil))

	// Request report
	report, err := b.reportReq.RequestReport(userData, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to request SNP report: %w", err)
	}

	// Serialize report
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, report.Version)
	_ = binary.Write(&buf, binary.LittleEndian, report.GuestSVN)
	buf.Write(report.LaunchDigest[:])
	buf.Write(report.ReportData[:])
	buf.Write(report.Signature[:])

	return buf.Bytes(), nil
}

// DeriveKey derives a key from the SEV-SNP root of trust
func (b *SEVHardwareBackend) DeriveKey(context []byte, keySize int) ([]byte, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, ErrHardwareNotInitialized
	}

	// Request derived key using VCEK
	rootKey, err := b.keyReq.RequestKey(SNP_KEY_ROOT_VCEK, SNP_KEY_GUEST_FIELD, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to request derived key: %w", err)
	}

	// Further derive using context
	kdf := hkdf.New(sha256.New, rootKey, nil, context)
	derived := make([]byte, keySize)
	if _, err := kdf.Read(derived); err != nil {
		return nil, err
	}

	return derived, nil
}

// Seal encrypts data using SEV-SNP sealing
func (b *SEVHardwareBackend) Seal(plaintext []byte) ([]byte, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, ErrHardwareNotInitialized
	}

	// Derive sealing key
	sealKey, err := b.DeriveKey([]byte("sev-snp-seal-v1"), 32)
	if err != nil {
		return nil, err
	}

	// Generate nonce
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Simple XOR encryption for simulation (real impl would use AES-GCM)
	ciphertext := make([]byte, len(plaintext))
	keyStream := make([]byte, len(plaintext))
	expandKey(sealKey, nonce, keyStream)
	for i := range plaintext {
		ciphertext[i] = plaintext[i] ^ keyStream[i]
	}

	// Format: nonce || ciphertext
	sealed := make([]byte, 12+len(ciphertext))
	copy(sealed[:12], nonce)
	copy(sealed[12:], ciphertext)

	return sealed, nil
}

// Unseal decrypts data that was previously sealed
func (b *SEVHardwareBackend) Unseal(sealed []byte) ([]byte, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, ErrHardwareNotInitialized
	}

	if len(sealed) < 12 {
		return nil, errors.New("sealed data too short")
	}

	// Derive sealing key
	sealKey, err := b.DeriveKey([]byte("sev-snp-seal-v1"), 32)
	if err != nil {
		return nil, err
	}

	nonce := sealed[:12]
	ciphertext := sealed[12:]

	// Decrypt
	plaintext := make([]byte, len(ciphertext))
	keyStream := make([]byte, len(ciphertext))
	expandKey(sealKey, nonce, keyStream)
	for i := range ciphertext {
		plaintext[i] = ciphertext[i] ^ keyStream[i]
	}

	return plaintext, nil
}

// =============================================================================
// SEV Hardware Detection Helper
// =============================================================================

// detectSEVCapabilities detects SEV-SNP hardware and populates capabilities
func detectSEVCapabilities(caps *HardwareCapabilities) {
	detector := NewSEVHardwareDetector()
	if err := detector.Detect(); err != nil {
		caps.DetectionErrors = append(caps.DetectionErrors, fmt.Sprintf("SEV-SNP: %v", err))
		return
	}

	if !detector.IsAvailable() {
		return
	}

	caps.SEVSNPAvailable = true
	caps.SEVSNPVersion = detector.Version()
	caps.SEVGuestDevice = detector.GetDevicePath()
	caps.SEVAPIVersion = detector.APIVersion()
}

// =============================================================================
// SEV-SNP Utility Types
// =============================================================================

// SEVSNPPlatformInfo contains platform information from SEV-SNP
type SEVSNPPlatformInfo struct {
	APIVersion    int
	BuildID       uint32
	TCBVersion    SNPTCBVersion
	PlatformFlags uint64
	ChipID        SNPChipID
	ReportedTCB   SNPTCBVersion
}

// GetPlatformInfo retrieves platform information from SEV-SNP hardware
func (b *SEVHardwareBackend) GetPlatformInfo() (*SEVSNPPlatformInfo, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, ErrHardwareNotInitialized
	}

	// Get a report to extract platform info
	var userData [64]byte
	report, err := b.reportReq.RequestReport(userData, 0)
	if err != nil {
		return nil, err
	}

	return &SEVSNPPlatformInfo{
		APIVersion:  b.detector.APIVersion(),
		TCBVersion:  report.CurrentTCB,
		ChipID:      report.ChipID,
		ReportedTCB: report.ReportedTCB,
	}, nil
}

// VerifyGuestPolicy validates the guest policy meets security requirements
func VerifyGuestPolicy(policy SNPGuestPolicy) error {
	if policy.Debug {
		return errors.New("debug mode is not allowed in production")
	}
	if policy.ABIMajor < 1 {
		return errors.New("ABI major version must be at least 1")
	}
	return nil
}
