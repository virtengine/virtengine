// Package enclave_runtime provides TEE enclave implementations.
//
// This file provides common utilities and types for the hardware abstraction layer
// that enables VirtEngine to interact with real TEE hardware (Intel SGX, AMD SEV-SNP,
// AWS Nitro) when available, with graceful fallback to simulation.
//
// Task Reference: VE-2029 - Hardware TEE Integration Layer
package enclave_runtime

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

// =============================================================================
// Hardware Mode Configuration
// =============================================================================

// HardwareMode controls whether to use real hardware or simulation
type HardwareMode int

const (
	// HardwareModeAuto auto-detects available hardware and uses it if present
	HardwareModeAuto HardwareMode = iota

	// HardwareModeSimulate forces simulation mode even if hardware is available
	HardwareModeSimulate

	// HardwareModeRequire requires real hardware and fails if unavailable
	HardwareModeRequire
)

// String returns a human-readable representation of the hardware mode
func (m HardwareMode) String() string {
	switch m {
	case HardwareModeAuto:
		return "auto"
	case HardwareModeSimulate:
		return "simulate"
	case HardwareModeRequire:
		return "require"
	default:
		return fmt.Sprintf("unknown(%d)", m)
	}
}

// =============================================================================
// Hardware Capabilities
// =============================================================================

// HardwareCapabilities represents detected TEE hardware on the system
type HardwareCapabilities struct {
	// SGX capabilities
	SGXAvailable    bool   // SGX hardware detected
	SGXVersion      int    // 1 or 2 (SGX2 supports dynamic memory)
	SGXFLCSupported bool   // Flexible Launch Control support
	SGXDriverPath   string // Path to SGX device (/dev/sgx_enclave)
	SGXProvisionPath string // Path to provision device (/dev/sgx_provision)

	// SEV-SNP capabilities
	SEVSNPAvailable  bool   // SEV-SNP hardware detected
	SEVSNPVersion    string // SEV-SNP version string (e.g., "1.51")
	SEVGuestDevice   string // Path to guest device (/dev/sev-guest)
	SEVAPIVersion    int    // SEV API version

	// Nitro capabilities
	NitroAvailable   bool   // Nitro hardware detected
	NitroVersion     string // Nitro CLI version
	NitroDevice      string // Path to Nitro device (/dev/nitro_enclaves)
	NitroCLIPath     string // Path to nitro-cli binary

	// Preferred backend based on detection priority
	PreferredBackend AttestationType

	// Detection metadata
	DetectedAt      time.Time
	DetectionErrors []string
}

// String returns a summary of hardware capabilities
func (h *HardwareCapabilities) String() string {
	var available []string
	if h.SGXAvailable {
		available = append(available, fmt.Sprintf("SGX%d", h.SGXVersion))
	}
	if h.SEVSNPAvailable {
		available = append(available, fmt.Sprintf("SEV-SNP(%s)", h.SEVSNPVersion))
	}
	if h.NitroAvailable {
		available = append(available, fmt.Sprintf("Nitro(%s)", h.NitroVersion))
	}
	if len(available) == 0 {
		return "No TEE hardware detected"
	}
	return fmt.Sprintf("TEE Hardware: %v (preferred: %s)", available, h.PreferredBackend)
}

// HasAnyHardware returns true if any TEE hardware is available
func (h *HardwareCapabilities) HasAnyHardware() bool {
	return h.SGXAvailable || h.SEVSNPAvailable || h.NitroAvailable
}

// =============================================================================
// Hardware Detection
// =============================================================================

var (
	// cachedCapabilities stores the detected hardware capabilities
	cachedCapabilities *HardwareCapabilities
	capabilitiesMu     sync.RWMutex
	capabilitiesOnce   sync.Once
)

// DetectHardware probes for all available TEE hardware on the system.
// Results are cached after the first call.
func DetectHardware() HardwareCapabilities {
	capabilitiesOnce.Do(func() {
		cachedCapabilities = detectHardwareInternal()
	})

	capabilitiesMu.RLock()
	defer capabilitiesMu.RUnlock()
	return *cachedCapabilities
}

// RefreshHardwareDetection forces a re-detection of hardware capabilities.
// This is useful if hardware state may have changed (e.g., driver loaded).
func RefreshHardwareDetection() HardwareCapabilities {
	capabilitiesMu.Lock()
	defer capabilitiesMu.Unlock()

	cachedCapabilities = detectHardwareInternal()
	return *cachedCapabilities
}

// detectHardwareInternal performs the actual hardware detection
func detectHardwareInternal() *HardwareCapabilities {
	caps := &HardwareCapabilities{
		DetectedAt:       time.Now(),
		DetectionErrors:  make([]string, 0),
		PreferredBackend: AttestationTypeSimulated,
	}

	// Only detect hardware on Linux
	if runtime.GOOS != "linux" {
		caps.DetectionErrors = append(caps.DetectionErrors,
			fmt.Sprintf("TEE hardware detection only supported on Linux (current: %s)", runtime.GOOS))
		return caps
	}

	// Detect SGX
	detectSGXCapabilities(caps)

	// Detect SEV-SNP
	detectSEVCapabilities(caps)

	// Detect Nitro
	detectNitroCapabilities(caps)

	// Determine preferred backend (priority: SGX > SEV-SNP > Nitro > Simulated)
	caps.PreferredBackend = determinePreferredBackend(caps)

	return caps
}

// determinePreferredBackend selects the best available TEE platform
func determinePreferredBackend(caps *HardwareCapabilities) AttestationType {
	// Priority order based on security properties and maturity
	if caps.SGXAvailable && caps.SGXFLCSupported {
		return AttestationTypeSGX
	}
	if caps.SEVSNPAvailable {
		return AttestationTypeSEVSNP
	}
	if caps.NitroAvailable {
		return AttestationTypeNitro
	}
	if caps.SGXAvailable {
		// SGX without FLC (legacy EPID attestation)
		return AttestationTypeSGX
	}
	return AttestationTypeSimulated
}

// =============================================================================
// Hardware Error Types
// =============================================================================

// ErrHardwareNotAvailable is returned when required hardware is not present
var ErrHardwareNotAvailable = errors.New("required TEE hardware not available")

// ErrHardwareNotInitialized is returned when hardware has not been initialized
var ErrHardwareNotInitialized = errors.New("TEE hardware not initialized")

// ErrHardwareOperationFailed is returned when a hardware operation fails
var ErrHardwareOperationFailed = errors.New("TEE hardware operation failed")

// ErrDeviceNotFound is returned when a required device file is not found
var ErrDeviceNotFound = errors.New("TEE device not found")

// ErrPermissionDenied is returned when permission to access hardware is denied
var ErrPermissionDenied = errors.New("permission denied for TEE hardware access")

// HardwareError wraps a hardware-specific error with context
type HardwareError struct {
	Platform   AttestationType
	Operation  string
	Underlying error
	DevicePath string
}

// Error implements the error interface
func (e *HardwareError) Error() string {
	if e.DevicePath != "" {
		return fmt.Sprintf("%s hardware error during %s on %s: %v",
			e.Platform, e.Operation, e.DevicePath, e.Underlying)
	}
	return fmt.Sprintf("%s hardware error during %s: %v",
		e.Platform, e.Operation, e.Underlying)
}

// Unwrap returns the underlying error
func (e *HardwareError) Unwrap() error {
	return e.Underlying
}

// =============================================================================
// Hardware Backend Interface
// =============================================================================

// HardwareBackend defines the interface that all TEE hardware backends must implement
type HardwareBackend interface {
	// Platform returns the attestation type for this backend
	Platform() AttestationType

	// IsAvailable returns true if this hardware is available
	IsAvailable() bool

	// Initialize sets up the hardware backend
	Initialize() error

	// Shutdown cleanly shuts down the hardware backend
	Shutdown() error

	// GetAttestation generates a hardware attestation with the given nonce
	GetAttestation(nonce []byte) ([]byte, error)

	// DeriveKey derives a key from the hardware root of trust
	DeriveKey(context []byte, keySize int) ([]byte, error)

	// Seal encrypts data using hardware-backed sealing
	Seal(plaintext []byte) ([]byte, error)

	// Unseal decrypts data that was previously sealed
	Unseal(ciphertext []byte) ([]byte, error)
}

// =============================================================================
// Hardware State Management
// =============================================================================

// HardwareState tracks the state of hardware backends
type HardwareState struct {
	mu sync.RWMutex

	sgxBackend   HardwareBackend
	sevBackend   HardwareBackend
	nitroBackend HardwareBackend

	mode             HardwareMode
	activeBackend    HardwareBackend
	initialized      bool
	initError        error
	lastHealthCheck  time.Time
	healthCheckError error
}

// NewHardwareState creates a new hardware state manager
func NewHardwareState(mode HardwareMode) *HardwareState {
	return &HardwareState{
		mode: mode,
	}
}

// Initialize initializes the hardware state with detected backends
func (s *HardwareState) Initialize() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.initialized {
		return nil
	}

	caps := DetectHardware()

	// Handle mode requirements
	switch s.mode {
	case HardwareModeRequire:
		if !caps.HasAnyHardware() {
			s.initError = fmt.Errorf("%w: mode is 'require' but no hardware detected",
				ErrHardwareNotAvailable)
			return s.initError
		}
	case HardwareModeSimulate:
		// Don't initialize any hardware backends
		s.initialized = true
		return nil
	}

	// Initialize backends based on detected hardware
	var initErrors []error

	if caps.SGXAvailable {
		backend := NewSGXHardwareBackend()
		if err := backend.Initialize(); err != nil {
			initErrors = append(initErrors, fmt.Errorf("SGX init: %w", err))
		} else {
			s.sgxBackend = backend
		}
	}

	if caps.SEVSNPAvailable {
		backend := NewSEVHardwareBackend()
		if err := backend.Initialize(); err != nil {
			initErrors = append(initErrors, fmt.Errorf("SEV init: %w", err))
		} else {
			s.sevBackend = backend
		}
	}

	if caps.NitroAvailable {
		backend := NewNitroHardwareBackend()
		if err := backend.Initialize(); err != nil {
			initErrors = append(initErrors, fmt.Errorf("Nitro init: %w", err))
		} else {
			s.nitroBackend = backend
		}
	}

	// Select active backend based on preference
	s.activeBackend = s.selectActiveBackend(caps.PreferredBackend)

	if s.mode == HardwareModeRequire && s.activeBackend == nil {
		s.initError = fmt.Errorf("%w: failed to initialize any hardware backend: %v",
			ErrHardwareNotAvailable, initErrors)
		return s.initError
	}

	s.initialized = true
	return nil
}

// selectActiveBackend chooses the best available backend
func (s *HardwareState) selectActiveBackend(preferred AttestationType) HardwareBackend {
	switch preferred {
	case AttestationTypeSGX:
		if s.sgxBackend != nil {
			return s.sgxBackend
		}
	case AttestationTypeSEVSNP:
		if s.sevBackend != nil {
			return s.sevBackend
		}
	case AttestationTypeNitro:
		if s.nitroBackend != nil {
			return s.nitroBackend
		}
	}

	// Fallback order
	if s.sgxBackend != nil {
		return s.sgxBackend
	}
	if s.sevBackend != nil {
		return s.sevBackend
	}
	if s.nitroBackend != nil {
		return s.nitroBackend
	}
	return nil
}

// GetActiveBackend returns the currently active backend
func (s *HardwareState) GetActiveBackend() HardwareBackend {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.activeBackend
}

// IsHardwareActive returns true if a hardware backend is active
func (s *HardwareState) IsHardwareActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.activeBackend != nil
}

// Shutdown shuts down all initialized backends
func (s *HardwareState) Shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var errs []error

	if s.sgxBackend != nil {
		if err := s.sgxBackend.Shutdown(); err != nil {
			errs = append(errs, fmt.Errorf("SGX shutdown: %w", err))
		}
	}

	if s.sevBackend != nil {
		if err := s.sevBackend.Shutdown(); err != nil {
			errs = append(errs, fmt.Errorf("SEV shutdown: %w", err))
		}
	}

	if s.nitroBackend != nil {
		if err := s.nitroBackend.Shutdown(); err != nil {
			errs = append(errs, fmt.Errorf("Nitro shutdown: %w", err))
		}
	}

	s.initialized = false
	s.activeBackend = nil

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}

// =============================================================================
// Utility Functions
// =============================================================================

// checkDeviceExists checks if a device file exists and is accessible
func checkDeviceExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		if os.IsPermission(err) {
			return true, ErrPermissionDenied
		}
		return false, err
	}
	// Check if it's a device file (mode type bits)
	if info.Mode()&os.ModeDevice != 0 || info.Mode()&os.ModeCharDevice != 0 {
		return true, nil
	}
	return false, fmt.Errorf("path exists but is not a device: %s", path)
}

// checkExecutableExists checks if an executable exists in PATH or at an absolute path
func checkExecutableExists(name string) (string, bool) {
	// Check absolute path first
	if info, err := os.Stat(name); err == nil && !info.IsDir() {
		return name, true
	}

	// Check common paths
	paths := []string{
		"/usr/bin/" + name,
		"/usr/local/bin/" + name,
		"/opt/bin/" + name,
	}

	for _, path := range paths {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, true
		}
	}

	return "", false
}

// readSysFile reads a value from a sysfs file
func readSysFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	// Trim newline
	result := string(data)
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}
	return result, nil
}

// =============================================================================
// CPUID Detection (Platform-Independent Stubs)
// =============================================================================

// CPUIDResult holds the result of a CPUID instruction
type CPUIDResult struct {
	EAX, EBX, ECX, EDX uint32
}

// cpuid executes the CPUID instruction (stub - real implementation in platform-specific file)
func cpuid(eax, ecx uint32) CPUIDResult {
	// This is a stub. Real implementation requires assembly or cgo.
	// On non-x86 platforms, return zeros.
	return CPUIDResult{}
}

// hasCPUIDSGX checks if the CPU supports SGX via CPUID
func hasCPUIDSGX() bool {
	// Check CPUID.07H.EBX[2] for SGX support
	result := cpuid(7, 0)
	return (result.EBX & (1 << 2)) != 0
}

// hasCPUIDSGX2 checks if the CPU supports SGX2 via CPUID
func hasCPUIDSGX2() bool {
	// Check CPUID.12H.0.EAX[1] for SGX2 support
	result := cpuid(0x12, 0)
	return (result.EAX & (1 << 1)) != 0
}

// hasCPUIDFLC checks if the CPU supports Flexible Launch Control
func hasCPUIDFLC() bool {
	// Check CPUID.07H.ECX[30] for FLC support
	result := cpuid(7, 0)
	return (result.ECX & (1 << 30)) != 0
}
