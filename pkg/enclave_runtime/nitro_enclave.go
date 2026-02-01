// Package enclave_runtime provides TEE enclave implementations.
//
// This file implements the AWS Nitro Enclaves service interface for VirtEngine VEID.
// AWS Nitro Enclaves provide isolated compute environments on AWS instances
// with cryptographic attestation via the AWS Nitro Attestation process.
//
// Task Reference: VE-2025 - AWS Nitro Enclave Implementation
//
// IMPORTANT: This is a POC implementation. Real AWS Nitro Enclave calls are stubbed
// and marked with TODO comments. Full implementation requires:
// - Instance types: c5.xlarge, c5a.xlarge, m5.xlarge, r5.xlarge or larger
// - Enclave-enabled AMI
// - nitro-cli installed and /dev/nitro_enclaves device available
// - vsock kernel module for enclave communication
package enclave_runtime

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"golang.org/x/crypto/hkdf"
)

// hashAlgorithmSHA384 is the SHA384 hash algorithm string
const hashAlgorithmSHA384 = "SHA384"

// =============================================================================
// Nitro Enclave Constants and Types
// =============================================================================

const (
	// NitroDefaultCID is the default Context ID for enclave vsock communication
	NitroDefaultCID = 16

	// NitroDefaultVsockPort is the default vsock port for enclave communication
	NitroDefaultVsockPort = 5000

	// NitroPCRCount is the number of PCRs in Nitro attestation (0-15)
	NitroPCRCount = 16

	// NitroPCRDigestSize is the size of each PCR (SHA-384)
	NitroPCRDigestSize = 48

	// NitroMaxUserData is the maximum size of user data in attestation
	NitroMaxUserData = 1024

	// NitroModuleIDSize is the maximum size of module ID
	NitroModuleIDSize = 64

	// NitroMaxNonceSize is the maximum nonce size
	NitroMaxNonceSize = 64

	// NitroAttestationDocVersion is the version of attestation document format
	NitroAttestationDocVersion = 1

	// NitroDefaultMemoryMB is default enclave memory in MB
	NitroDefaultMemoryMB = 2048

	// NitroDefaultCPUCount is default number of vCPUs for enclave
	NitroDefaultCPUCount = 2

	// NitroNSMDevicePath is the path to the Nitro Security Module device
	NitroNSMDevicePath = "/dev/nsm"

	// NitroEnclavesDevicePath is the path to the Nitro Enclaves device
	NitroEnclavesDevicePath = "/dev/nitro_enclaves"
)

// NitroPCRIndex represents the index of a Platform Configuration Register
type NitroPCRIndex uint8

// Nitro PCR indices and their meanings
const (
	// PCR0 - Enclave Image File (EIF) measurement
	NitroPCR0EIF NitroPCRIndex = 0

	// PCR1 - Linux kernel and boot ramfs measurement
	NitroPCR1Kernel NitroPCRIndex = 1

	// PCR2 - User application measurement
	NitroPCR2App NitroPCRIndex = 2

	// PCR3 - IAM role attached to the parent instance (if any)
	NitroPCR3IAMRole NitroPCRIndex = 3

	// PCR4 - Instance ID of the parent EC2 instance
	NitroPCR4InstanceID NitroPCRIndex = 4

	// PCR8 - Signing certificate for signed enclave images
	NitroPCR8SigningCert NitroPCRIndex = 8
)

// NitroPCR represents a single Platform Configuration Register value
type NitroPCR [NitroPCRDigestSize]byte

// String returns hex representation
func (p NitroPCR) String() string {
	return fmt.Sprintf("%x", p[:])
}

// IsZero returns true if PCR is all zeros
func (p NitroPCR) IsZero() bool {
	for _, b := range p {
		if b != 0 {
			return false
		}
	}
	return true
}

// NitroPCRSet represents the set of PCR values (PCR0-PCR15)
type NitroPCRSet struct {
	PCRs [NitroPCRCount]NitroPCR
}

// Get returns the PCR value at the given index
func (s *NitroPCRSet) Get(index NitroPCRIndex) NitroPCR {
	if index >= NitroPCRCount {
		return NitroPCR{}
	}
	return s.PCRs[index]
}

// Set sets the PCR value at the given index
func (s *NitroPCRSet) Set(index NitroPCRIndex, value NitroPCR) {
	if index < NitroPCRCount {
		s.PCRs[index] = value
	}
}

// Digest computes a combined digest of all PCRs
func (s *NitroPCRSet) Digest() []byte {
	h := sha512.New384()
	for i := range s.PCRs {
		h.Write(s.PCRs[i][:])
	}
	return h.Sum(nil)
}

// NitroAttestationDocument represents the AWS Nitro attestation document
// The actual document is CBOR-encoded with COSE Sign1 signature
type NitroAttestationDocument struct {
	// ModuleID is the enclave image ID
	ModuleID string

	// Timestamp is the Unix timestamp in milliseconds when attestation was generated
	Timestamp uint64

	// Digest is the digest algorithm used (SHA384)
	Digest string

	// PCRs contains the Platform Configuration Register values
	PCRs NitroPCRSet

	// Certificate is the DER-encoded EC certificate for the enclave
	Certificate []byte

	// CABundle contains the certificate chain (array of DER-encoded certs)
	CABundle [][]byte

	// PublicKey is optional user-provided public key (COSE_Key format)
	PublicKey []byte

	// UserData is optional user-provided data (up to 1KB)
	UserData []byte

	// Nonce is optional nonce to prevent replay attacks
	Nonce []byte
}

// Validate validates the attestation document structure
func (d *NitroAttestationDocument) Validate() error {
	if d.ModuleID == "" {
		return errors.New("module ID is required")
	}
	if len(d.ModuleID) > NitroModuleIDSize {
		return errors.New("module ID too long")
	}
	if d.Timestamp == 0 {
		return errors.New("timestamp is required")
	}
	if d.Digest != hashAlgorithmSHA384 {
		return errors.New("digest must be SHA384")
	}
	if len(d.Certificate) == 0 {
		return errors.New("certificate is required")
	}
	if len(d.UserData) > NitroMaxUserData {
		return errors.New("user data too large")
	}
	if len(d.Nonce) > NitroMaxNonceSize {
		return errors.New("nonce too large")
	}
	// PCR0, PCR1, PCR2 must be non-zero (core measurements)
	if d.PCRs.Get(NitroPCR0EIF).IsZero() {
		return errors.New("PCR0 (EIF measurement) must be present")
	}
	return nil
}

// GetPCRDigest returns a hash of PCR0, PCR1, and PCR2 (core measurements)
func (d *NitroAttestationDocument) GetPCRDigest() []byte {
	h := sha512.New384()
	h.Write(d.PCRs.PCRs[NitroPCR0EIF][:])
	h.Write(d.PCRs.PCRs[NitroPCR1Kernel][:])
	h.Write(d.PCRs.PCRs[NitroPCR2App][:])
	return h.Sum(nil)
}

// NitroEnclaveConfig configures the Nitro enclave service
type NitroEnclaveConfig struct {
	// EnclaveImagePath is the path to the Enclave Image File (EIF)
	EnclaveImagePath string

	// CPUCount is the number of vCPUs to allocate to the enclave
	CPUCount int

	// MemoryMB is the amount of memory in MB to allocate
	MemoryMB int

	// DebugMode enables debug mode (NOT FOR PRODUCTION)
	DebugMode bool

	// CID is the Context ID for vsock communication (default: 16)
	CID uint32

	// VsockPort is the vsock port for enclave communication (default: 5000)
	VsockPort uint32

	// EnclaveID is an optional identifier for the enclave instance
	EnclaveID string

	// AllowedPCR0 is the expected PCR0 value (EIF measurement) for verification
	// If empty, PCR0 verification is skipped (not recommended for production)
	AllowedPCR0 []byte

	// AllowedPCR1 is the expected PCR1 value (kernel measurement)
	AllowedPCR1 []byte

	// AllowedPCR2 is the expected PCR2 value (application measurement)
	AllowedPCR2 []byte
}

// Validate validates the enclave configuration
func (c *NitroEnclaveConfig) Validate() error {
	if c.EnclaveImagePath == "" {
		return errors.New("enclave image path is required")
	}
	if c.CPUCount <= 0 {
		c.CPUCount = NitroDefaultCPUCount
	}
	if c.MemoryMB <= 0 {
		c.MemoryMB = NitroDefaultMemoryMB
	}
	if c.CID == 0 {
		c.CID = NitroDefaultCID
	}
	if c.VsockPort == 0 {
		c.VsockPort = NitroDefaultVsockPort
	}
	return nil
}

// NitroEnclaveState represents the state of a Nitro enclave
type NitroEnclaveState int

const (
	NitroEnclaveStateStopped NitroEnclaveState = iota
	NitroEnclaveStateStarting
	NitroEnclaveStateRunning
	NitroEnclaveStateStopping
	NitroEnclaveStateFailed
)

func (s NitroEnclaveState) String() string {
	switch s {
	case NitroEnclaveStateStopped:
		return "stopped"
	case NitroEnclaveStateStarting:
		return "starting"
	case NitroEnclaveStateRunning:
		return "running"
	case NitroEnclaveStateStopping:
		return "stopping"
	case NitroEnclaveStateFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// =============================================================================
// Nitro Enclave Service Implementation
// =============================================================================

// NitroEnclaveServiceImpl implements the EnclaveService interface for AWS Nitro Enclaves
type NitroEnclaveServiceImpl struct {
	mu sync.RWMutex

	// Configuration
	config        NitroEnclaveConfig
	runtimeConfig RuntimeConfig
	hardwareMode  HardwareMode

	// Hardware backend (nil if using simulation)
	hardwareBackend *NitroHardwareBackend

	// State
	initialized  bool
	startTime    time.Time
	activeReqs   int
	totalProc    uint64
	currentEpoch uint64
	lastError    string
	enclaveState NitroEnclaveState

	// Simulated enclave state (in real impl, these are inside the Nitro enclave)
	enclaveID     string
	pcrSet        NitroPCRSet
	moduleID      string
	certificate   []byte
	caBundle      [][]byte
	encryptionKey []byte
	signingKey    []byte
	encryptPubKey []byte
	signingPubKey []byte
}

// Compile-time interface check
var _ EnclaveService = (*NitroEnclaveServiceImpl)(nil)

// NewNitroEnclaveServiceImpl creates a new Nitro enclave service implementation
//
// This is a POC implementation that simulates Nitro Enclave operations.
// For production use, this must be replaced with actual nitro-cli and NSM calls.
func NewNitroEnclaveServiceImpl(config NitroEnclaveConfig) (*NitroEnclaveServiceImpl, error) {
	return NewNitroEnclaveServiceImplWithMode(config, HardwareModeAuto)
}

// NewNitroEnclaveServiceImplWithMode creates a new Nitro enclave service with explicit hardware mode
func NewNitroEnclaveServiceImplWithMode(config NitroEnclaveConfig, mode HardwareMode) (*NitroEnclaveServiceImpl, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Validate configuration
	if config.DebugMode {
		// Allow debug mode for development, but log warning
		fmt.Println("WARNING: Nitro enclave debug mode enabled - NOT SECURE FOR PRODUCTION")
	}

	svc := &NitroEnclaveServiceImpl{
		config:       config,
		currentEpoch: 1,
		enclaveState: NitroEnclaveStateStopped,
		hardwareMode: mode,
	}

	// Initialize hardware backend based on mode
	if mode != HardwareModeSimulate {
		backend := NewNitroHardwareBackend()
		if backend.IsAvailable() {
			if err := backend.Initialize(); err == nil {
				svc.hardwareBackend = backend
				fmt.Println("INFO: Nitro hardware backend initialized successfully")
			} else if mode == HardwareModeRequire {
				return nil, fmt.Errorf("Nitro hardware required but initialization failed: %w", err)
			} else {
				fmt.Printf("INFO: Nitro hardware initialization failed, using simulation: %v\n", err)
			}
		} else if mode == HardwareModeRequire {
			return nil, fmt.Errorf("%w: Nitro hardware required but not available", ErrHardwareNotAvailable)
		} else {
			fmt.Println("INFO: Nitro hardware not available, using simulation mode")
		}
	} else {
		fmt.Println("INFO: Nitro running in forced simulation mode")
	}

	return svc, nil
}

// =============================================================================
// EnclaveService Interface Implementation
// =============================================================================

// Initialize initializes the Nitro enclave
func (n *NitroEnclaveServiceImpl) Initialize(config RuntimeConfig) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.initialized {
		return errors.New("enclave already initialized")
	}

	n.runtimeConfig = config
	n.startTime = time.Now()
	n.enclaveState = NitroEnclaveStateStarting

	// TODO: Real Nitro implementation would:
	// 1. Check for /dev/nitro_enclaves device
	// 2. Run: nitro-cli run-enclave --eif-path <path> --cpu-count <n> --memory <mb>
	// 3. Parse enclave ID from output
	// 4. Establish vsock connection to enclave
	// 5. Initialize enclave application via vsock

	// Simulate enclave start
	if err := n.simulateEnclaveStart(); err != nil {
		n.enclaveState = NitroEnclaveStateFailed
		n.lastError = err.Error()
		return fmt.Errorf("failed to start enclave: %w", err)
	}

	// Derive enclave keys
	if err := n.deriveEnclaveKeys(); err != nil {
		n.enclaveState = NitroEnclaveStateFailed
		n.lastError = err.Error()
		return fmt.Errorf("failed to derive keys: %w", err)
	}

	n.enclaveState = NitroEnclaveStateRunning
	n.initialized = true
	return nil
}

// Score performs identity scoring inside the Nitro enclave
func (n *NitroEnclaveServiceImpl) Score(ctx context.Context, request *ScoringRequest) (*ScoringResult, error) {
	n.mu.Lock()
	if !n.initialized {
		n.mu.Unlock()
		return nil, ErrEnclaveNotInitialized
	}
	if n.activeReqs >= n.runtimeConfig.MaxConcurrentRequests {
		n.mu.Unlock()
		return nil, ErrEnclaveUnavailable
	}
	n.activeReqs++
	n.mu.Unlock()

	defer func() {
		n.mu.Lock()
		n.activeReqs--
		n.totalProc++
		n.mu.Unlock()
	}()

	// Validate request
	if err := request.Validate(n.runtimeConfig); err != nil {
		return &ScoringResult{RequestID: request.RequestID, Error: err.Error()}, nil
	}

	startTime := time.Now()

	// TODO: Real Nitro implementation would:
	// 1. Connect to enclave via vsock (CID, port)
	// 2. Send encrypted payload via vsock
	// 3. Inside enclave: decrypt with sealed key
	// 4. Inside enclave: run ML scoring
	// 5. Inside enclave: sign result with enclave key
	// 6. Receive signed result via vsock

	// Simulate enclave scoring via vsock (POC)
	result := n.simulateEnclaveScoring(request)
	result.ProcessingTimeMs = time.Since(startTime).Milliseconds()

	return result, nil
}

// GetMeasurement returns the enclave measurement (PCR digest)
func (n *NitroEnclaveServiceImpl) GetMeasurement() ([]byte, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if !n.initialized {
		return nil, ErrEnclaveNotInitialized
	}

	// Return combined PCR0/PCR1/PCR2 digest
	h := sha512.New384()
	h.Write(n.pcrSet.PCRs[NitroPCR0EIF][:])
	h.Write(n.pcrSet.PCRs[NitroPCR1Kernel][:])
	h.Write(n.pcrSet.PCRs[NitroPCR2App][:])
	return h.Sum(nil), nil
}

// GetEncryptionPubKey returns the enclave's encryption public key
func (n *NitroEnclaveServiceImpl) GetEncryptionPubKey() ([]byte, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if !n.initialized {
		return nil, ErrEnclaveNotInitialized
	}

	return n.encryptPubKey, nil
}

// GetSigningPubKey returns the enclave's signing public key
func (n *NitroEnclaveServiceImpl) GetSigningPubKey() ([]byte, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if !n.initialized {
		return nil, ErrEnclaveNotInitialized
	}

	return n.signingPubKey, nil
}

// GenerateAttestation generates a Nitro attestation document
func (n *NitroEnclaveServiceImpl) GenerateAttestation(reportData []byte) ([]byte, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if !n.initialized {
		return nil, ErrEnclaveNotInitialized
	}

	if len(reportData) > NitroMaxUserData {
		return nil, errors.New("report data too large")
	}

	// Use hardware backend if available
	if n.hardwareBackend != nil {
		attestation, err := n.hardwareBackend.GetAttestation(reportData)
		if err != nil {
			n.lastError = fmt.Sprintf("hardware attestation failed: %v", err)
			// Fall back to simulation
		} else {
			return attestation, nil
		}
	}

	// TODO: Real Nitro implementation would:
	// 1. Open /dev/nsm (Nitro Security Module)
	// 2. Send GetAttestation request with user_data and public_key
	// 3. Receive CBOR-encoded COSE Sign1 attestation document
	// 4. Optionally verify the document locally

	doc, err := n.simulateAttestationDocument(reportData)
	if err != nil {
		return nil, fmt.Errorf("attestation generation failed: %w", err)
	}

	// Serialize the document (in real impl, this is CBOR with COSE Sign1)
	serialized := n.serializeAttestationDocument(doc)

	return serialized, nil
}

// RotateKeys initiates key rotation
func (n *NitroEnclaveServiceImpl) RotateKeys() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.initialized {
		return ErrEnclaveNotInitialized
	}

	// Increment epoch
	n.currentEpoch++

	// TODO: Real Nitro implementation would:
	// 1. Send key rotation command to enclave via vsock
	// 2. Enclave generates new key pair
	// 3. Enclave returns new public key
	// 4. Update local public key cache

	// Re-derive keys with new epoch
	if err := n.deriveEnclaveKeys(); err != nil {
		n.lastError = fmt.Sprintf("key rotation failed: %v", err)
		return fmt.Errorf("key rotation failed: %w", err)
	}

	return nil
}

// GetStatus returns the enclave status
func (n *NitroEnclaveServiceImpl) GetStatus() EnclaveStatus {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var uptime int64
	if n.initialized {
		uptime = int64(time.Since(n.startTime).Seconds())
	}

	return EnclaveStatus{
		Initialized:    n.initialized,
		Available:      n.initialized && n.activeReqs < n.runtimeConfig.MaxConcurrentRequests,
		CurrentEpoch:   n.currentEpoch,
		ActiveRequests: n.activeReqs,
		TotalProcessed: n.totalProc,
		LastError:      n.lastError,
		Uptime:         uptime,
	}
}

// Shutdown gracefully shuts down the enclave
func (n *NitroEnclaveServiceImpl) Shutdown() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.initialized {
		return nil
	}

	n.enclaveState = NitroEnclaveStateStopping

	// TODO: Real Nitro implementation would:
	// 1. Send shutdown command to enclave via vsock
	// 2. Wait for enclave to acknowledge
	// 3. Run: nitro-cli terminate-enclave --enclave-id <id>
	// 4. Verify enclave has stopped

	// Securely clear keys from memory
	n.scrubKeys()

	n.enclaveState = NitroEnclaveStateStopped
	n.initialized = false
	return nil
}

// =============================================================================
// Nitro-Specific Methods
// =============================================================================

// GetPlatformType returns PlatformNitro
func (n *NitroEnclaveServiceImpl) GetPlatformType() PlatformType {
	return PlatformNitro
}

// IsPlatformSecure returns true (Nitro is secure when not in debug mode)
func (n *NitroEnclaveServiceImpl) IsPlatformSecure() bool {
	return !n.config.DebugMode
}

// IsHardwareEnabled returns true if real Nitro hardware is being used
func (n *NitroEnclaveServiceImpl) IsHardwareEnabled() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.hardwareBackend != nil
}

// GetHardwareMode returns the configured hardware mode
func (n *NitroEnclaveServiceImpl) GetHardwareMode() HardwareMode {
	return n.hardwareMode
}

// LaunchEnclave launches a Nitro enclave using the hardware backend
func (n *NitroEnclaveServiceImpl) LaunchEnclave(ctx context.Context) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.hardwareBackend != nil {
		config := NitroHWEnclaveConfig{
			EIFPath:   n.config.EnclaveImagePath,
			CPUCount:  n.config.CPUCount,
			MemoryMB:  int64(n.config.MemoryMB),
			DebugMode: n.config.DebugMode,
			VsockPort: n.config.VsockPort,
		}
		_, err := n.hardwareBackend.RunAndConnect(ctx, config)
		if err != nil {
			n.lastError = fmt.Sprintf("hardware enclave launch failed: %v", err)
			return err
		}
		enclaveID, _ := n.hardwareBackend.GetEnclaveInfo()
		n.enclaveID = enclaveID
		return nil
	}

	// Simulated launch
	return n.simulateEnclaveStart()
}

// GetEnclaveID returns the enclave instance ID
func (n *NitroEnclaveServiceImpl) GetEnclaveID() (string, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if !n.initialized {
		return "", ErrEnclaveNotInitialized
	}

	return n.enclaveID, nil
}

// GetPCR returns a specific PCR value
func (n *NitroEnclaveServiceImpl) GetPCR(index NitroPCRIndex) (NitroPCR, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if !n.initialized {
		return NitroPCR{}, ErrEnclaveNotInitialized
	}

	if index >= NitroPCRCount {
		return NitroPCR{}, errors.New("invalid PCR index")
	}

	return n.pcrSet.Get(index), nil
}

// GetPCRSet returns all PCR values
func (n *NitroEnclaveServiceImpl) GetPCRSet() (*NitroPCRSet, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if !n.initialized {
		return nil, ErrEnclaveNotInitialized
	}

	// Return a copy
	pcrsCopy := n.pcrSet
	return &pcrsCopy, nil
}

// VerifyPCRs verifies PCR values against expected values in config
func (n *NitroEnclaveServiceImpl) VerifyPCRs() error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if !n.initialized {
		return ErrEnclaveNotInitialized
	}

	// Verify PCR0 if configured
	if len(n.config.AllowedPCR0) > 0 {
		pcr0 := n.pcrSet.Get(NitroPCR0EIF)
		if !bytesEqual(pcr0[:], n.config.AllowedPCR0) {
			return errors.New("PCR0 mismatch: enclave image not trusted")
		}
	}

	// Verify PCR1 if configured
	if len(n.config.AllowedPCR1) > 0 {
		pcr1 := n.pcrSet.Get(NitroPCR1Kernel)
		if !bytesEqual(pcr1[:], n.config.AllowedPCR1) {
			return errors.New("PCR1 mismatch: kernel measurement not trusted")
		}
	}

	// Verify PCR2 if configured
	if len(n.config.AllowedPCR2) > 0 {
		pcr2 := n.pcrSet.Get(NitroPCR2App)
		if !bytesEqual(pcr2[:], n.config.AllowedPCR2) {
			return errors.New("PCR2 mismatch: application measurement not trusted")
		}
	}

	return nil
}

// GetEnclaveState returns the current enclave state
func (n *NitroEnclaveServiceImpl) GetEnclaveState() NitroEnclaveState {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.enclaveState
}

// VerifyMeasurement verifies the measurement matches
func (n *NitroEnclaveServiceImpl) VerifyMeasurement(expected []byte) bool {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if !n.initialized {
		return false
	}

	// Get combined PCR digest
	h := sha512.New384()
	h.Write(n.pcrSet.PCRs[NitroPCR0EIF][:])
	h.Write(n.pcrSet.PCRs[NitroPCR1Kernel][:])
	h.Write(n.pcrSet.PCRs[NitroPCR2App][:])
	actual := h.Sum(nil)

	return bytesEqual(actual, expected)
}

// =============================================================================
// Simulation Methods (POC - Replace with real implementation)
// =============================================================================

// simulateEnclaveStart simulates starting a Nitro enclave using nitro-cli
func (n *NitroEnclaveServiceImpl) simulateEnclaveStart() error {
	// TODO: Real implementation would:
	// cmd := exec.Command("nitro-cli", "run-enclave",
	//     "--eif-path", n.config.EnclaveImagePath,
	//     "--cpu-count", strconv.Itoa(n.config.CPUCount),
	//     "--memory", strconv.Itoa(n.config.MemoryMB),
	//     "--enclave-cid", strconv.FormatUint(uint64(n.config.CID), 10),
	// )
	// if n.config.DebugMode {
	//     cmd.Args = append(cmd.Args, "--debug-mode")
	// }
	// output, err := cmd.Output()
	// Parse JSON output for enclave ID and state

	// Generate simulated enclave ID
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return err
	}
	n.enclaveID = fmt.Sprintf("i-%x", idBytes)

	// Generate module ID from image path
	moduleHash := sha256.Sum256([]byte(n.config.EnclaveImagePath))
	n.moduleID = fmt.Sprintf("veid-nitro-%x", moduleHash[:8])

	// Generate simulated PCR values
	if err := n.simulatePCRMeasurement(); err != nil {
		return err
	}

	// Generate simulated certificate and CA bundle
	if err := n.simulateCertificateGeneration(); err != nil {
		return err
	}

	return nil
}

// simulatePCRMeasurement simulates PCR measurement from Nitro
func (n *NitroEnclaveServiceImpl) simulatePCRMeasurement() error {
	// PCR0: EIF measurement (hash of enclave image)
	pcr0Hash := sha512.Sum384([]byte("nitro_eif_" + n.config.EnclaveImagePath))
	copy(n.pcrSet.PCRs[NitroPCR0EIF][:], pcr0Hash[:])

	// PCR1: Kernel measurement
	pcr1Hash := sha512.Sum384([]byte("nitro_kernel_linux_5.10"))
	copy(n.pcrSet.PCRs[NitroPCR1Kernel][:], pcr1Hash[:])

	// PCR2: Application measurement
	pcr2Hash := sha512.Sum384([]byte("nitro_app_veid_scorer_v1"))
	copy(n.pcrSet.PCRs[NitroPCR2App][:], pcr2Hash[:])

	// PCR3: IAM role (optional - leave zeros if no role)
	// PCR4: Instance ID
	instanceHash := sha512.Sum384([]byte("nitro_instance_id_" + n.enclaveID))
	copy(n.pcrSet.PCRs[NitroPCR4InstanceID][:], instanceHash[:])

	// PCR8: Signing certificate (if signed EIF)
	signingHash := sha512.Sum384([]byte("nitro_signing_cert_virtengine"))
	copy(n.pcrSet.PCRs[NitroPCR8SigningCert][:], signingHash[:])

	return nil
}

// simulateCertificateGeneration simulates certificate generation
func (n *NitroEnclaveServiceImpl) simulateCertificateGeneration() error {
	// Generate simulated enclave certificate (in real impl, from NSM)
	certData := sha256.Sum256([]byte("nitro_enclave_cert_" + n.enclaveID))
	n.certificate = certData[:]

	// Generate simulated CA bundle
	n.caBundle = [][]byte{
		sha256Bytes([]byte("nitro_ca_root")),
		sha256Bytes([]byte("nitro_ca_intermediate")),
		sha256Bytes([]byte("nitro_enclave_signer")),
	}

	return nil
}

// deriveEnclaveKeys derives encryption and signing keys using HKDF
func (n *NitroEnclaveServiceImpl) deriveEnclaveKeys() error {
	// In real Nitro, keys are generated inside the enclave
	// and only public keys are exported via vsock

	// Generate seed from PCR measurements and epoch
	seed := make([]byte, 0, 128)
	seed = append(seed, n.pcrSet.PCRs[NitroPCR0EIF][:]...)
	seed = append(seed, n.pcrSet.PCRs[NitroPCR2App][:]...)
	seed = append(seed, byte(n.currentEpoch))

	// Derive key using HKDF
	hkdfReader := hkdf.New(sha256.New, seed, []byte("nitro_enclave_salt"), []byte("nitro_key_derivation"))

	// Derive encryption key (32 bytes for X25519)
	n.encryptionKey = make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, n.encryptionKey); err != nil {
		return fmt.Errorf("failed to derive encryption key: %w", err)
	}

	// Derive signing key (32 bytes for Ed25519 seed)
	n.signingKey = make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, n.signingKey); err != nil {
		return fmt.Errorf("failed to derive signing key: %w", err)
	}

	// Generate public keys (simulated - in real impl, use proper crypto)
	encPubHash := sha256.Sum256(append([]byte("nitro_enc_pub_"), n.encryptionKey...))
	n.encryptPubKey = encPubHash[:]

	sigPubHash := sha256.Sum256(append([]byte("nitro_sig_pub_"), n.signingKey...))
	n.signingPubKey = sigPubHash[:]

	return nil
}

// simulateVsockCommunication simulates vsock communication with enclave
//
//nolint:unused,unparam // Reserved for vsock communication implementation; result 1 (error) for future failures
func (n *NitroEnclaveServiceImpl) simulateVsockCommunication(request []byte) ([]byte, error) {
	// TODO: Real implementation would:
	// 1. Open vsock connection: net.Dial("vsock", fmt.Sprintf("%d:%d", n.config.CID, n.config.VsockPort))
	// 2. Send request length + request bytes
	// 3. Read response length + response bytes
	// 4. Close connection

	// Simulate processing delay
	// In real implementation, this would be the enclave processing time

	// Simulate response
	responseHash := sha256.Sum256(append([]byte("nitro_vsock_response_"), request...))
	return responseHash[:], nil
}

// simulateEnclaveScoring simulates ML scoring inside the Nitro enclave
func (n *NitroEnclaveServiceImpl) simulateEnclaveScoring(request *ScoringRequest) *ScoringResult {
	// Compute input hash
	inputHash := sha256.Sum256(append(request.Ciphertext, request.Nonce...))

	// Simulate ML scoring (deterministic based on input for testing)
	scoreHash := sha256.Sum256(append(inputHash[:], []byte(request.ScopeID)...))
	score := uint32(scoreHash[0] % 101) // 0-100

	// Determine status
	var status string
	switch {
	case score >= 80:
		status = "verified"
	case score >= 60:
		status = "pending_review"
	default:
		status = "rejected"
	}

	// Generate evidence hashes
	evidenceHashes := [][]byte{
		sha256Bytes([]byte("nitro_face_embedding")),
		sha256Bytes([]byte("nitro_document_features")),
	}

	// Generate model version hash
	modelVersionHash := sha256Bytes([]byte("nitro_veid_model_v1.0.0"))

	// Sign result inside enclave (simulated)
	signingPayload := n.computeSigningPayload(request.RequestID, score, status, inputHash[:])
	enclaveSignature := n.signInsideEnclave(signingPayload)

	// Get measurement hash (compute directly to avoid lock)
	measurementH := sha512.New384()
	measurementH.Write(n.pcrSet.PCRs[NitroPCR0EIF][:])
	measurementH.Write(n.pcrSet.PCRs[NitroPCR1Kernel][:])
	measurementH.Write(n.pcrSet.PCRs[NitroPCR2App][:])
	measurementHash := measurementH.Sum(nil)

	return &ScoringResult{
		RequestID:            request.RequestID,
		Score:                score,
		Status:               status,
		ReasonCodes:          []string{"nitro_score_" + status},
		EvidenceHashes:       evidenceHashes,
		ModelVersionHash:     modelVersionHash,
		InputHash:            inputHash[:],
		EnclaveSignature:     enclaveSignature,
		MeasurementHash:      measurementHash,
		AttestationReference: sha256Bytes([]byte("nitro_attestation_ref_" + n.enclaveID)),
	}
}

// computeSigningPayload computes the payload to sign
func (n *NitroEnclaveServiceImpl) computeSigningPayload(requestID string, score uint32, status string, inputHash []byte) []byte {
	h := sha256.New()
	h.Write([]byte(requestID))
	h.Write([]byte{byte(score >> 24), byte(score >> 16), byte(score >> 8), byte(score)})
	h.Write([]byte(status))
	h.Write(inputHash)
	h.Write(n.pcrSet.PCRs[NitroPCR0EIF][:])
	return h.Sum(nil)
}

// signInsideEnclave simulates signing inside the enclave
func (n *NitroEnclaveServiceImpl) signInsideEnclave(payload []byte) []byte {
	// In real Nitro, this uses Ed25519 with key stored in enclave memory
	h := sha256.New()
	h.Write(n.signingKey)
	h.Write(payload)
	return h.Sum(nil)
}

// simulateAttestationDocument generates a simulated NSM attestation document
func (n *NitroEnclaveServiceImpl) simulateAttestationDocument(userData []byte) (*NitroAttestationDocument, error) {
	// TODO: Real implementation would:
	// 1. Open /dev/nsm
	// 2. Send GetAttestation command with user_data, public_key, and nonce
	// 3. Receive CBOR-encoded COSE Sign1 document
	// 4. Parse and return the document

	doc := &NitroAttestationDocument{
		ModuleID: n.moduleID,
		//nolint:gosec // G115: UnixMilli timestamp is positive and fits in uint64
		Timestamp:   uint64(time.Now().UnixMilli()),
		Digest:      hashAlgorithmSHA384,
		PCRs:        n.pcrSet,
		Certificate: n.certificate,
		CABundle:    n.caBundle,
		PublicKey:   n.encryptPubKey,
		UserData:    userData,
		Nonce:       nil, // Nonce is optional
	}

	// Validate the document structure
	if err := doc.Validate(); err != nil {
		return nil, err
	}

	return doc, nil
}

// serializeAttestationDocument serializes the attestation document
// In real implementation, this would be CBOR-encoded with COSE Sign1 signature
func (n *NitroEnclaveServiceImpl) serializeAttestationDocument(doc *NitroAttestationDocument) []byte {
	// TODO: Real implementation would use:
	// - github.com/fxamacker/cbor/v2 for CBOR encoding
	// - COSE Sign1 structure for signature

	// Simplified serialization for POC
	buf := make([]byte, 0, 2048)

	// Magic and version
	buf = append(buf, []byte("NITRO_ATTEST_V1")...)

	// Module ID (length-prefixed)
	buf = append(buf, byte(len(doc.ModuleID)))
	buf = append(buf, []byte(doc.ModuleID)...)

	// Timestamp
	buf = binary.BigEndian.AppendUint64(buf, doc.Timestamp)

	// PCRs (all 16)
	for i := 0; i < NitroPCRCount; i++ {
		buf = append(buf, doc.PCRs.PCRs[i][:]...)
	}

	// Certificate
	//nolint:gosec // G115: Certificate length fits in uint16
	buf = binary.BigEndian.AppendUint16(buf, uint16(len(doc.Certificate)))
	buf = append(buf, doc.Certificate...)

	// User data
	//nolint:gosec // G115: UserData length fits in uint16
	buf = binary.BigEndian.AppendUint16(buf, uint16(len(doc.UserData)))
	buf = append(buf, doc.UserData...)

	// Public key
	//nolint:gosec // G115: PublicKey length fits in uint16
	buf = binary.BigEndian.AppendUint16(buf, uint16(len(doc.PublicKey)))
	buf = append(buf, doc.PublicKey...)

	// Signature (simulated)
	signature := n.signInsideEnclave(buf)
	buf = append(buf, signature...)

	return buf
}

// scrubKeys securely clears keys from memory
func (n *NitroEnclaveServiceImpl) scrubKeys() {
	// Use ScrubBytes from memory_scrub.go
	if n.encryptionKey != nil {
		ScrubBytes(n.encryptionKey)
	}
	if n.signingKey != nil {
		ScrubBytes(n.signingKey)
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// bytesEqual performs constant-time comparison of byte slices
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

// =============================================================================
// Nitro Attestation Verification (for validators)
// =============================================================================

// VerifyNitroAttestation verifies a Nitro attestation document
func VerifyNitroAttestation(attestation []byte, expectedPCR0 []byte) error {
	// TODO: Real implementation would:
	// 1. Parse CBOR-encoded COSE Sign1 structure
	// 2. Extract and verify the certificate chain
	// 3. Verify the signature using the enclave certificate
	// 4. Check PCR values against expected measurements
	// 5. Verify timestamp is recent

	// POC: Basic format validation
	if len(attestation) < 50 {
		return errors.New("attestation too short")
	}

	// Check magic
	if string(attestation[:15]) != "NITRO_ATTEST_V1" {
		return errors.New("invalid attestation format")
	}

	// Extract and verify PCR0 if expected value provided
	if len(expectedPCR0) > 0 {
		// PCRs start at offset: 15 (magic) + 1 (module_id_len) + module_id + 8 (timestamp)
		moduleIDLen := int(attestation[15])
		pcrStart := 15 + 1 + moduleIDLen + 8

		if len(attestation) < pcrStart+NitroPCRDigestSize {
			return errors.New("attestation too short for PCR extraction")
		}

		pcr0 := attestation[pcrStart : pcrStart+NitroPCRDigestSize]
		if !bytesEqual(pcr0, expectedPCR0) {
			return errors.New("PCR0 mismatch: enclave not trusted")
		}
	}

	return nil
}

// ExtractPCRsFromAttestation extracts PCR values from an attestation document
func ExtractPCRsFromAttestation(attestation []byte) (*NitroPCRSet, error) {
	if len(attestation) < 50 {
		return nil, errors.New("attestation too short")
	}

	if string(attestation[:15]) != "NITRO_ATTEST_V1" {
		return nil, errors.New("invalid attestation format")
	}

	// Calculate PCR offset
	moduleIDLen := int(attestation[15])
	pcrStart := 15 + 1 + moduleIDLen + 8

	if len(attestation) < pcrStart+(NitroPCRCount*NitroPCRDigestSize) {
		return nil, errors.New("attestation too short for PCR extraction")
	}

	pcrSet := &NitroPCRSet{}
	for i := 0; i < NitroPCRCount; i++ {
		offset := pcrStart + (i * NitroPCRDigestSize)
		copy(pcrSet.PCRs[i][:], attestation[offset:offset+NitroPCRDigestSize])
	}

	return pcrSet, nil
}

// BindNitroResultToReport creates user data that binds a scoring result to attestation
func BindNitroResultToReport(result *ScoringResult, nonce []byte) []byte {
	h := sha256.New()
	h.Write([]byte(result.RequestID))
	h.Write(result.InputHash)
	h.Write(result.EnclaveSignature)
	h.Write(nonce)
	digest := h.Sum(nil)

	// User data format: digest (32 bytes) + nonce (up to 32 bytes)
	userData := make([]byte, 0, 64)
	userData = append(userData, digest...)
	if len(nonce) > 32 {
		userData = append(userData, nonce[:32]...)
	} else {
		userData = append(userData, nonce...)
	}

	return userData
}

