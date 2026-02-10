// Package hardware provides a comprehensive hardware abstraction layer for TEE integration.
//
// This package defines unified types and interfaces for interacting with various
// Trusted Execution Environment (TEE) platforms including Intel SGX, AMD SEV-SNP,
// and AWS Nitro Enclaves. The abstraction allows VirtEngine to work seamlessly
// across different hardware platforms with automatic detection and fallback.
//
// Task Reference: VE-2029 - Hardware TEE Integration Layer
package hardware

import (
	"errors"
	"fmt"
	"time"
)

// =============================================================================
// Platform Enumeration
// =============================================================================

// Platform represents a TEE hardware platform type.
type Platform int

const (
	// PlatformUnknown represents an unknown or undetected platform.
	PlatformUnknown Platform = iota

	// PlatformSGX represents Intel Software Guard Extensions.
	PlatformSGX

	// PlatformSEVSNP represents AMD Secure Encrypted Virtualization with
	// Secure Nested Paging.
	PlatformSEVSNP

	// PlatformNitro represents AWS Nitro Enclaves.
	PlatformNitro

	// PlatformSimulated represents a simulated TEE for testing and development.
	PlatformSimulated
)

// String returns the string representation of the platform.
func (p Platform) String() string {
	switch p {
	case PlatformSGX:
		return "sgx"
	case PlatformSEVSNP:
		return "sev-snp"
	case PlatformNitro:
		return "nitro"
	case PlatformSimulated:
		return "simulated"
	default:
		return "unknown"
	}
}

// ParsePlatform converts a string to a Platform value.
func ParsePlatform(s string) Platform {
	switch s {
	case "sgx", "SGX", "intel-sgx":
		return PlatformSGX
	case "sev-snp", "SEV-SNP", "sev", "SEV", "amd-sev":
		return PlatformSEVSNP
	case "nitro", "Nitro", "aws-nitro":
		return PlatformNitro
	case "simulated", "sim", "simulation":
		return PlatformSimulated
	default:
		return PlatformUnknown
	}
}

// IsHardware returns true if this platform represents real hardware (not simulated).
func (p Platform) IsHardware() bool {
	return p == PlatformSGX || p == PlatformSEVSNP || p == PlatformNitro
}

// =============================================================================
// TCB Status
// =============================================================================

// TCBStatus represents the status of the Trusted Computing Base.
type TCBStatus int

const (
	// TCBStatusUnknown indicates the TCB status could not be determined.
	TCBStatusUnknown TCBStatus = iota

	// TCBStatusUpToDate indicates the TCB is fully up to date.
	TCBStatusUpToDate

	// TCBStatusOutOfDate indicates the TCB needs security updates.
	TCBStatusOutOfDate

	// TCBStatusConfigurationNeeded indicates configuration changes are required.
	TCBStatusConfigurationNeeded

	// TCBStatusRevoked indicates the TCB has been revoked due to security issues.
	TCBStatusRevoked

	// TCBStatusSWHardeningNeeded indicates software hardening updates are needed.
	TCBStatusSWHardeningNeeded
)

// String returns the string representation of the TCB status.
func (s TCBStatus) String() string {
	switch s {
	case TCBStatusUpToDate:
		return "UpToDate"
	case TCBStatusOutOfDate:
		return "OutOfDate"
	case TCBStatusConfigurationNeeded:
		return "ConfigurationNeeded"
	case TCBStatusRevoked:
		return "Revoked"
	case TCBStatusSWHardeningNeeded:
		return "SWHardeningNeeded"
	default:
		return "Unknown"
	}
}

// IsSecure returns true if the TCB status is acceptable for production use.
func (s TCBStatus) IsSecure() bool {
	return s == TCBStatusUpToDate || s == TCBStatusSWHardeningNeeded
}

// =============================================================================
// Hardware Capabilities
// =============================================================================

// HardwareCapabilities provides detailed information about detected TEE hardware.
type HardwareCapabilities struct {
	// Platform identification
	Platform Platform

	// SGX-specific capabilities
	SGX SGXCapabilities

	// SEV-SNP-specific capabilities
	SEVSNP SEVSNPCapabilities

	// Nitro-specific capabilities
	Nitro NitroCapabilities

	// Detection metadata
	DetectedAt      time.Time
	DetectionErrors []string

	// Overall status
	Available bool
	TCBStatus TCBStatus
}

// SGXCapabilities contains Intel SGX-specific capability information.
type SGXCapabilities struct {
	Available        bool
	Version          int    // 1 or 2 (SGX2 supports dynamic memory)
	FLCSupported     bool   // Flexible Launch Control support
	EPCSize          uint64 // Enclave Page Cache size in bytes
	MaxEnclaveSize   uint64 // Maximum enclave size
	EnclaveDevice    string // Path to /dev/sgx_enclave
	ProvisionDevice  string // Path to /dev/sgx_provision
	DCAPAvailable    bool   // DCAP quote generation available
	AESMSocketPath   string // Path to AESM socket
	TCBStatus        TCBStatus
	CPUSVNComponents [16]byte // CPU Security Version Number components
}

// SEVSNPCapabilities contains AMD SEV-SNP-specific capability information.
type SEVSNPCapabilities struct {
	Available     bool
	Version       string // SEV-SNP version string
	APIVersion    int    // SEV API version (e.g., 151 for 1.51)
	GuestDevice   string // Path to /dev/sev-guest
	TCBVersion    TCBVersionInfo
	ChipID        [64]byte // Unique chip identifier
	PlatformFlags uint64   // Platform capability flags
	TCBStatus     TCBStatus
}

// TCBVersionInfo contains detailed TCB version information for SEV-SNP.
type TCBVersionInfo struct {
	BootLoader uint8
	TEE        uint8
	SNP        uint8
	Microcode  uint8
}

// NitroCapabilities contains AWS Nitro-specific capability information.
type NitroCapabilities struct {
	Available      bool
	Version        string // nitro-cli version
	DevicePath     string // Path to /dev/nitro_enclaves
	CLIPath        string // Path to nitro-cli binary
	MaxCPUs        int    // Maximum CPUs available for enclaves
	MaxMemoryMiB   int64  // Maximum memory available for enclaves
	NSMAvailable   bool   // Nitro Security Module available
	VsockSupported bool   // Vsock communication supported
}

// String returns a summary of hardware capabilities.
func (c *HardwareCapabilities) String() string {
	var platforms []string
	if c.SGX.Available {
		platforms = append(platforms, fmt.Sprintf("SGX%d", c.SGX.Version))
	}
	if c.SEVSNP.Available {
		platforms = append(platforms, fmt.Sprintf("SEV-SNP(%s)", c.SEVSNP.Version))
	}
	if c.Nitro.Available {
		platforms = append(platforms, fmt.Sprintf("Nitro(%s)", c.Nitro.Version))
	}
	if len(platforms) == 0 {
		return "No TEE hardware detected"
	}
	return fmt.Sprintf("TEE Hardware: %v (recommended: %s)", platforms, c.Platform)
}

// HasAnyHardware returns true if any real TEE hardware is available.
func (c *HardwareCapabilities) HasAnyHardware() bool {
	return c.SGX.Available || c.SEVSNP.Available || c.Nitro.Available
}

// GetRecommendedPlatform returns the recommended platform based on capabilities.
func (c *HardwareCapabilities) GetRecommendedPlatform() Platform {
	// Priority: SGX with FLC > SEV-SNP > Nitro > SGX without FLC > Simulated
	if c.SGX.Available && c.SGX.FLCSupported && c.SGX.DCAPAvailable {
		return PlatformSGX
	}
	if c.SEVSNP.Available {
		return PlatformSEVSNP
	}
	if c.Nitro.Available {
		return PlatformNitro
	}
	if c.SGX.Available {
		return PlatformSGX
	}
	return PlatformSimulated
}

// =============================================================================
// Hardware Information
// =============================================================================

// HardwareInfo provides detailed hardware identification and status information.
type HardwareInfo struct {
	// Platform type
	Platform Platform

	// Human-readable name
	Name string

	// Vendor information
	Vendor string

	// Firmware/driver version
	Version string

	// Unique hardware identifier (platform-specific)
	HardwareID []byte

	// Current TCB status
	TCBStatus TCBStatus

	// Platform-specific measurements
	Measurements map[string][]byte

	// Device paths
	DevicePaths []string

	// Feature flags
	Features map[string]bool

	// Last update timestamp
	LastUpdated time.Time
}

// =============================================================================
// Attestation Quote Interface
// =============================================================================

// AttestationQuote provides a platform-agnostic interface for attestation quotes.
type AttestationQuote interface {
	// Platform returns the TEE platform that generated this quote.
	Platform() Platform

	// RawBytes returns the raw quote bytes for transmission.
	RawBytes() []byte

	// UserData returns the user-provided data embedded in the quote.
	UserData() []byte

	// Measurement returns the enclave/VM measurement (MRENCLAVE, launch digest, etc.).
	Measurement() []byte

	// SignerID returns the signer identity (MRSIGNER for SGX, empty for others).
	SignerID() []byte

	// Timestamp returns when the quote was generated.
	Timestamp() time.Time

	// Verify performs basic structural verification of the quote.
	// Full verification requires platform-specific certificate chains.
	Verify() error
}

// =============================================================================
// Sealing Key Interface
// =============================================================================

// SealingKey provides a platform-agnostic interface for hardware-bound key derivation.
type SealingKey interface {
	// Platform returns the TEE platform providing this key.
	Platform() Platform

	// Derive derives a key of the specified size using the given context.
	// The derived key is bound to the current enclave/VM identity.
	Derive(context []byte, size int) ([]byte, error)

	// Seal encrypts data using the hardware-bound sealing key.
	Seal(plaintext []byte) ([]byte, error)

	// Unseal decrypts data that was previously sealed.
	Unseal(ciphertext []byte) ([]byte, error)

	// Policy returns the key derivation policy (platform-specific).
	Policy() KeyPolicy
}

// KeyPolicy specifies how sealing keys are bound to identity.
type KeyPolicy int

const (
	// KeyPolicyEnclave binds the key to the exact enclave measurement.
	// For SGX: MRENCLAVE, for SEV-SNP: launch digest
	KeyPolicyEnclave KeyPolicy = iota

	// KeyPolicySigner binds the key to the enclave signer.
	// For SGX: MRSIGNER, allows key migration across enclave versions.
	KeyPolicySigner

	// KeyPolicyPlatform binds the key to the platform identity.
	// Allows key access from any enclave on the same machine.
	KeyPolicyPlatform
)

// String returns the string representation of the key policy.
func (p KeyPolicy) String() string {
	switch p {
	case KeyPolicyEnclave:
		return "enclave"
	case KeyPolicySigner:
		return "signer"
	case KeyPolicyPlatform:
		return "platform"
	default:
		return "unknown"
	}
}

// =============================================================================
// Backend Interface
// =============================================================================

// Backend defines the interface that all TEE hardware backends must implement.
// This is the core abstraction that enables platform-agnostic TEE operations.
type Backend interface {
	// Platform returns the TEE platform type for this backend.
	Platform() Platform

	// IsAvailable returns true if this hardware is currently available.
	IsAvailable() bool

	// Initialize sets up the hardware backend and any required resources.
	Initialize() error

	// Shutdown cleanly releases all resources associated with this backend.
	Shutdown() error

	// GetAttestation generates a hardware attestation quote with the given nonce.
	// The nonce is incorporated into the attestation to prevent replay attacks.
	GetAttestation(nonce []byte) ([]byte, error)

	// DeriveKey derives a key from the hardware root of trust.
	// The context is mixed into the derivation to produce unique keys.
	DeriveKey(context []byte, keySize int) ([]byte, error)

	// Seal encrypts data using hardware-backed sealing.
	// The sealed data can only be unsealed on the same platform.
	Seal(plaintext []byte) ([]byte, error)

	// Unseal decrypts data that was previously sealed by this backend.
	Unseal(ciphertext []byte) ([]byte, error)

	// HealthCheck verifies the backend is functioning correctly.
	HealthCheck() error

	// GetCapabilities returns detailed capability information.
	GetCapabilities() *HardwareCapabilities

	// GetInfo returns hardware identification and status information.
	GetInfo() *HardwareInfo
}

// =============================================================================
// Error Types
// =============================================================================

// Sentinel errors for hardware operations.
var (
	// ErrHardwareNotFound indicates the required TEE hardware was not detected.
	ErrHardwareNotFound = errors.New("TEE hardware not found")

	// ErrNotSupported indicates the requested operation is not supported
	// on the current platform.
	ErrNotSupported = errors.New("operation not supported on this platform")

	// ErrPermissionDenied indicates insufficient permissions to access
	// the TEE hardware devices.
	ErrPermissionDenied = errors.New("permission denied for TEE hardware access")

	// ErrInitializationFailed indicates the hardware backend failed to initialize.
	ErrInitializationFailed = errors.New("TEE hardware initialization failed")

	// ErrNotInitialized indicates an operation was attempted before initialization.
	ErrNotInitialized = errors.New("TEE hardware not initialized")

	// ErrAttestationFailed indicates attestation quote generation failed.
	ErrAttestationFailed = errors.New("attestation generation failed")

	// ErrKeyDerivationFailed indicates key derivation failed.
	ErrKeyDerivationFailed = errors.New("key derivation failed")

	// ErrSealingFailed indicates data sealing failed.
	ErrSealingFailed = errors.New("data sealing failed")

	// ErrUnsealingFailed indicates data unsealing failed.
	ErrUnsealingFailed = errors.New("data unsealing failed")

	// ErrQuoteVerificationFailed indicates quote verification failed.
	ErrQuoteVerificationFailed = errors.New("quote verification failed")

	// ErrBackendBusy indicates the backend is currently busy processing.
	ErrBackendBusy = errors.New("TEE backend is busy")

	// ErrHealthCheckFailed indicates the health check failed.
	ErrHealthCheckFailed = errors.New("TEE health check failed")
)

// HardwareError wraps a hardware-specific error with additional context.
type HardwareError struct {
	Platform   Platform
	Operation  string
	DevicePath string
	Underlying error
}

// Error implements the error interface.
func (e *HardwareError) Error() string {
	if e.DevicePath != "" {
		return fmt.Sprintf("%s hardware error during %s on %s: %v",
			e.Platform, e.Operation, e.DevicePath, e.Underlying)
	}
	return fmt.Sprintf("%s hardware error during %s: %v",
		e.Platform, e.Operation, e.Underlying)
}

// Unwrap returns the underlying error for errors.Is/As support.
func (e *HardwareError) Unwrap() error {
	return e.Underlying
}

// NewHardwareError creates a new HardwareError with the given details.
func NewHardwareError(platform Platform, operation string, err error) *HardwareError {
	return &HardwareError{
		Platform:   platform,
		Operation:  operation,
		Underlying: err,
	}
}

// NewHardwareErrorWithDevice creates a new HardwareError with device path information.
func NewHardwareErrorWithDevice(platform Platform, operation, devicePath string, err error) *HardwareError {
	return &HardwareError{
		Platform:   platform,
		Operation:  operation,
		DevicePath: devicePath,
		Underlying: err,
	}
}

// =============================================================================
// Device Path Constants
// =============================================================================

const (
	// SGX device paths
	SGXDeviceEnclave   = "/dev/sgx_enclave"
	SGXDeviceProvision = "/dev/sgx_provision"
	SGXDeviceLegacy    = "/dev/isgx"
	SGXAESMSocket      = "/var/run/aesmd/aesm.socket"

	// SEV-SNP device paths
	SEVGuestDevice    = "/dev/sev-guest"
	SEVDebugFSPath    = "/sys/kernel/debug/sev"
	SEVPlatformStatus = "/sys/kernel/debug/sev/status"

	// Nitro device paths
	NitroDevice          = "/dev/nitro_enclaves"
	NitroNSMDevice       = "/dev/nsm"
	NitroAllocatorConfig = "/etc/nitro_enclaves/allocator.yaml"
)

// =============================================================================
// Configuration Types
// =============================================================================

// Config provides configuration options for the hardware abstraction layer.
type Config struct {
	// PreferredPlatform specifies which platform to prefer when multiple are available.
	// Set to PlatformUnknown to auto-select.
	PreferredPlatform Platform

	// RequireHardware if true, fails initialization if no real hardware is available.
	RequireHardware bool

	// AllowSimulation if true, falls back to simulation when hardware is unavailable.
	AllowSimulation bool

	// Platform-specific configuration
	SGXConfig   SGXConfig
	SEVConfig   SEVConfig
	NitroConfig NitroConfig

	// Health check configuration
	HealthCheckInterval time.Duration
	HealthCheckTimeout  time.Duration
}

// SGXConfig contains Intel SGX-specific configuration.
type SGXConfig struct {
	EnclavePath          string
	DebugMode            bool
	KeyPolicy            KeyPolicy
	QuoteProviderLibrary string
}

// SEVConfig contains AMD SEV-SNP-specific configuration.
type SEVConfig struct {
	CertChainPath string
	VCEKCachePath string
	MinTCBVersion TCBVersionInfo
}

// NitroConfig contains AWS Nitro-specific configuration.
type NitroConfig struct {
	EnclaveImagePath string
	CPUs             int
	MemoryMiB        int64
	DebugMode        bool
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		PreferredPlatform:   PlatformUnknown, // Auto-select
		RequireHardware:     false,
		AllowSimulation:     true,
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		SGXConfig: SGXConfig{
			EnclavePath: "/opt/virtengine/enclaves/veid_scorer.signed.so",
			KeyPolicy:   KeyPolicyEnclave,
		},
		SEVConfig: SEVConfig{
			CertChainPath: "/etc/virtengine/sev/cert-chain.pem",
			VCEKCachePath: "/var/cache/virtengine/vcek",
		},
		NitroConfig: NitroConfig{
			CPUs:      2,
			MemoryMiB: 512,
		},
	}
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
	if c.RequireHardware && c.AllowSimulation {
		return errors.New("RequireHardware and AllowSimulation are mutually exclusive")
	}
	if c.HealthCheckInterval < time.Second {
		return errors.New("HealthCheckInterval must be at least 1 second")
	}
	if c.HealthCheckTimeout < 100*time.Millisecond {
		return errors.New("HealthCheckTimeout must be at least 100ms")
	}
	return nil
}
