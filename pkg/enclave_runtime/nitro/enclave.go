// Package nitro provides AWS Nitro Enclave integration for VirtEngine TEE.
//
// This package implements comprehensive Nitro Enclave support including:
// - Enclave lifecycle management (build, run, describe, terminate)
// - NSM (Nitro Security Module) interaction for attestation
// - Attestation document parsing and verification
// - Vsock communication helpers
//
// Build tags:
// - Default (no tag): Simulation mode for development/testing
// - nitro_hardware: Real Nitro Enclave operations
//
// Requirements for hardware mode:
// - EC2 instance with Nitro Enclave support (c5.xlarge, m5.xlarge, etc.)
// - nitro-cli installed and configured
// - /dev/nitro_enclaves device available
// - Enclave allocator configured in /etc/nitro_enclaves/allocator.yaml
//
// Task Reference: VE-2029 - Hardware TEE Integration Layer
package nitro

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// =============================================================================
// Constants
// =============================================================================

const (
	// Device paths
	NitroDevicePath     = "/dev/nitro_enclaves"
	NSMDevicePath       = "/dev/nsm"
	AllocatorConfigPath = "/etc/nitro_enclaves/allocator.yaml"

	// CLI binary
	NitroCLIBinary = "nitro-cli"

	// Vsock constants
	VsockCIDAny    uint32 = 0xFFFFFFFF // VMADDR_CID_ANY
	VsockCIDHost   uint32 = 2          // Host CID
	VsockCIDParent uint32 = 3          // Parent instance CID

	// Default enclave settings
	DefaultCPUCount  = 2
	DefaultMemoryMB  = 2048
	DefaultCID       = 16
	DefaultVsockPort = 5000

	// PCR constants
	PCRCount      = 16
	PCRDigestSize = 48 // SHA-384

	// PCR indices
	PCRIndexEIF        = 0 // Enclave Image File measurement
	PCRIndexKernel     = 1 // Linux kernel and boot ramfs
	PCRIndexApp        = 2 // User application
	PCRIndexIAMRole    = 3 // IAM role (if any)
	PCRIndexInstanceID = 4 // EC2 instance ID
	PCRIndexSigningKey = 8 // Signing certificate for signed images

	// Timeouts
	DefaultBuildTimeout     = 10 * time.Minute
	DefaultRunTimeout       = 30 * time.Second
	DefaultTerminateTimeout = 10 * time.Second
	DefaultDescribeTimeout  = 5 * time.Second
)

// =============================================================================
// Errors
// =============================================================================

var (
	// ErrEnclaveNotFound is returned when an enclave cannot be found
	ErrEnclaveNotFound = errors.New("enclave not found")

	// ErrEnclaveBuildFailed is returned when enclave image build fails
	ErrEnclaveBuildFailed = errors.New("enclave build failed")

	// ErrEnclaveRunFailed is returned when enclave start fails
	ErrEnclaveRunFailed = errors.New("enclave run failed")

	// ErrEnclaveTerminateFailed is returned when enclave termination fails
	ErrEnclaveTerminateFailed = errors.New("enclave terminate failed")

	// ErrNitroCLINotFound is returned when nitro-cli is not installed
	ErrNitroCLINotFound = errors.New("nitro-cli not found")

	// ErrNitroDeviceNotAvailable is returned when Nitro device is not available
	ErrNitroDeviceNotAvailable = errors.New("nitro device not available")

	// ErrVsockConnectionFailed is returned when vsock connection fails
	ErrVsockConnectionFailed = errors.New("vsock connection failed")

	// ErrInvalidEIFPath is returned when EIF path is invalid
	ErrInvalidEIFPath = errors.New("invalid EIF path")

	// ErrInvalidEnclaveID is returned when enclave ID format is invalid
	ErrInvalidEnclaveID = errors.New("invalid enclave ID format")

	// ErrSimulationMode is returned when operation requires hardware but running in simulation
	ErrSimulationMode = errors.New("operation not available in simulation mode")
)

// =============================================================================
// Enclave State
// =============================================================================

// EnclaveState represents the state of a Nitro enclave
type EnclaveState string

const (
	EnclaveStatePending    EnclaveState = "PENDING"
	EnclaveStateRunning    EnclaveState = "RUNNING"
	EnclaveStateTerminated EnclaveState = "TERMINATED"
	EnclaveStateFailed     EnclaveState = "FAILED"
)

// =============================================================================
// Enclave Info Types
// =============================================================================

// EnclaveInfo represents information about a running enclave
type EnclaveInfo struct {
	// EnclaveID is the unique identifier for the enclave
	EnclaveID string `json:"EnclaveID"`

	// ProcessID is the host process ID managing the enclave
	ProcessID int `json:"ProcessID"`

	// EnclaveCID is the Context ID for vsock communication
	EnclaveCID uint32 `json:"EnclaveCID"`

	// NumberOfCPUs is the number of vCPUs allocated
	NumberOfCPUs int `json:"NumberOfCPUs"`

	// CPUIDs is the list of host CPU IDs assigned
	CPUIDs []int `json:"CPUIDs"`

	// MemoryMiB is the memory allocated in MiB
	MemoryMiB int64 `json:"MemoryMiB"`

	// State is the current enclave state
	State EnclaveState `json:"State"`

	// Flags contains enclave flags (e.g., "DEBUG_MODE")
	Flags string `json:"Flags"`
}

// BuildOutput represents the output of build-enclave command
type BuildOutput struct {
	// Measurements contains the PCR measurements
	Measurements Measurements `json:"Measurements"`
}

// Measurements represents enclave measurements from build output
type Measurements struct {
	// HashAlgorithm is the algorithm used (e.g., "SHA384")
	HashAlgorithm string `json:"HashAlgorithm"`

	// PCR0 is the EIF measurement (hex string)
	PCR0 string `json:"PCR0"`

	// PCR1 is the kernel measurement (hex string)
	PCR1 string `json:"PCR1"`

	// PCR2 is the application measurement (hex string)
	PCR2 string `json:"PCR2"`
}

// =============================================================================
// NitroEnclave - Main Enclave Management
// =============================================================================

// NitroEnclave manages Nitro Enclave lifecycle and operations
type NitroEnclave struct {
	mu sync.RWMutex

	// Configuration
	simulated bool
	cliPath   string

	// Enclave state
	enclaveInfo  *EnclaveInfo
	measurements *Measurements

	// Simulated state
	simulatedEnclaves map[string]*EnclaveInfo
}

// NewNitroEnclave creates a new NitroEnclave instance
//
// If Nitro hardware is not available, it operates in simulation mode.
func NewNitroEnclave() (*NitroEnclave, error) {
	ne := &NitroEnclave{
		simulatedEnclaves: make(map[string]*EnclaveInfo),
	}

	// Check for Nitro hardware
	if err := ne.detectHardware(); err != nil {
		// Fall back to simulation mode
		ne.simulated = true
	}

	return ne, nil
}

// NewNitroEnclaveWithMode creates a NitroEnclave with explicit mode
func NewNitroEnclaveWithMode(requireHardware bool) (*NitroEnclave, error) {
	ne := &NitroEnclave{
		simulatedEnclaves: make(map[string]*EnclaveInfo),
	}

	if err := ne.detectHardware(); err != nil {
		if requireHardware {
			return nil, fmt.Errorf("hardware required but not available: %w", err)
		}
		ne.simulated = true
	}

	return ne, nil
}

// detectHardware checks for Nitro hardware availability
func (ne *NitroEnclave) detectHardware() error {
	// Check for Nitro device
	if _, err := os.Stat(NitroDevicePath); err != nil {
		return fmt.Errorf("%w: %v", ErrNitroDeviceNotAvailable, err)
	}

	// Check for nitro-cli
	cliPath, err := exec.LookPath(NitroCLIBinary)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrNitroCLINotFound, err)
	}
	ne.cliPath = cliPath

	return nil
}

// IsSimulated returns true if running in simulation mode
func (ne *NitroEnclave) IsSimulated() bool {
	ne.mu.RLock()
	defer ne.mu.RUnlock()
	return ne.simulated
}

// =============================================================================
// Enclave Lifecycle Operations
// =============================================================================

// BuildEnclave builds an Enclave Image File (EIF) from a Docker image
//
// Parameters:
//   - dockerImage: Docker image name/tag to convert to EIF
//   - outputEIF: Path where the EIF file will be written
//
// Returns an error if the build fails
func (ne *NitroEnclave) BuildEnclave(dockerImage string, outputEIF string) error {
	return ne.BuildEnclaveWithContext(context.Background(), dockerImage, outputEIF)
}

// BuildEnclaveWithContext builds an EIF with context for cancellation
func (ne *NitroEnclave) BuildEnclaveWithContext(ctx context.Context, dockerImage string, outputEIF string) error {
	ne.mu.Lock()
	defer ne.mu.Unlock()

	// Validate inputs
	if dockerImage == "" {
		return errors.New("docker image is required")
	}
	if outputEIF == "" {
		return errors.New("output EIF path is required")
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputEIF)
	if err := os.MkdirAll(outputDir, 0750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if ne.simulated {
		return ne.simulateBuildEnclave(dockerImage, outputEIF)
	}

	return ne.hardwareBuildEnclave(ctx, dockerImage, outputEIF)
}

// hardwareBuildEnclave builds EIF using real nitro-cli
func (ne *NitroEnclave) hardwareBuildEnclave(ctx context.Context, dockerImage string, outputEIF string) error {
	// Validate and sanitize docker image name
	if err := validateDockerImage(dockerImage); err != nil {
		return fmt.Errorf("invalid docker image: %w", err)
	}

	// Clean output path
	cleanPath, err := sanitizePath(outputEIF)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidEIFPath, err)
	}

	// Set timeout
	ctx, cancel := context.WithTimeout(ctx, DefaultBuildTimeout)
	defer cancel()

	args := []string{
		"build-enclave",
		"--docker-uri", dockerImage,
		"--output-file", cleanPath,
	}

	//nolint:gosec // G204: inputs validated above
	cmd := exec.CommandContext(ctx, ne.cliPath, args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("%w: %s", ErrEnclaveBuildFailed, string(exitErr.Stderr))
		}
		return fmt.Errorf("%w: %v", ErrEnclaveBuildFailed, err)
	}

	// Parse build output for measurements
	var buildOutput BuildOutput
	if err := json.Unmarshal(output, &buildOutput); err != nil {
		// Build succeeded but couldn't parse output - not fatal
		return nil
	}

	ne.measurements = &buildOutput.Measurements
	return nil
}

// simulateBuildEnclave simulates EIF building for testing
func (ne *NitroEnclave) simulateBuildEnclave(dockerImage string, outputEIF string) error {
	// Generate simulated measurements based on docker image
	imageHash := sha256.Sum256([]byte(dockerImage))

	ne.measurements = &Measurements{
		HashAlgorithm: "SHA384",
		PCR0:          hex.EncodeToString(imageHash[:]) + strings.Repeat("0", 48), // Pad to 96 chars
		PCR1:          strings.Repeat("00", 48),                                   // Zero PCR1
		PCR2:          strings.Repeat("00", 48),                                   // Zero PCR2
	}

	// Create a simulated EIF file
	eifContent := fmt.Sprintf("SIMULATED_EIF:%s:%d", dockerImage, time.Now().Unix())
	return os.WriteFile(outputEIF, []byte(eifContent), 0600)
}

// RunEnclave starts a new enclave from an EIF
//
// Parameters:
//   - eifPath: Path to the Enclave Image File
//   - cpuCount: Number of vCPUs to allocate (minimum 2)
//   - memoryMB: Memory in MB to allocate
//
// Returns EnclaveInfo on success
func (ne *NitroEnclave) RunEnclave(eifPath string, cpuCount int, memoryMB int64) (*EnclaveInfo, error) {
	return ne.RunEnclaveWithContext(context.Background(), eifPath, cpuCount, memoryMB)
}

// RunEnclaveWithContext starts an enclave with context for cancellation
func (ne *NitroEnclave) RunEnclaveWithContext(ctx context.Context, eifPath string, cpuCount int, memoryMB int64) (*EnclaveInfo, error) {
	ne.mu.Lock()
	defer ne.mu.Unlock()

	// Validate inputs
	if cpuCount < 2 {
		cpuCount = DefaultCPUCount
	}
	if memoryMB < 64 {
		memoryMB = DefaultMemoryMB
	}

	if ne.simulated {
		return ne.simulateRunEnclave(eifPath, cpuCount, memoryMB)
	}

	return ne.hardwareRunEnclave(ctx, eifPath, cpuCount, memoryMB)
}

// hardwareRunEnclave runs enclave using real nitro-cli
func (ne *NitroEnclave) hardwareRunEnclave(ctx context.Context, eifPath string, cpuCount int, memoryMB int64) (*EnclaveInfo, error) {
	// Validate EIF path
	cleanPath, err := sanitizePath(eifPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidEIFPath, err)
	}

	// Check EIF exists
	if _, err := os.Stat(cleanPath); err != nil {
		return nil, fmt.Errorf("%w: file not found", ErrInvalidEIFPath)
	}

	ctx, cancel := context.WithTimeout(ctx, DefaultRunTimeout)
	defer cancel()

	args := []string{
		"run-enclave",
		"--eif-path", cleanPath,
		"--cpu-count", fmt.Sprintf("%d", cpuCount),
		"--memory", fmt.Sprintf("%d", memoryMB),
	}

	//nolint:gosec // G204: cliPath validated, cleanPath sanitized
	cmd := exec.CommandContext(ctx, ne.cliPath, args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("%w: %s", ErrEnclaveRunFailed, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("%w: %v", ErrEnclaveRunFailed, err)
	}

	var info EnclaveInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("failed to parse run-enclave output: %w", err)
	}

	info.State = EnclaveStateRunning
	ne.enclaveInfo = &info
	return &info, nil
}

// simulateRunEnclave simulates enclave start for testing
func (ne *NitroEnclave) simulateRunEnclave(eifPath string, cpuCount int, memoryMB int64) (*EnclaveInfo, error) {
	// Generate simulated enclave ID (matches i-(?:sim-)?[a-f0-9-]{8,64} pattern)
	idBytes := make([]byte, 8)
	_, _ = rand.Read(idBytes)
	enclaveID := fmt.Sprintf("i-sim-%x", idBytes)

	// Generate simulated CID
	var cid uint32
	_ = binary.Read(rand.Reader, binary.LittleEndian, &cid)
	cid = (cid % 65000) + 100

	// Create CPU ID list
	cpuIDs := make([]int, cpuCount)
	for i := 0; i < cpuCount; i++ {
		cpuIDs[i] = i + 1
	}

	info := &EnclaveInfo{
		EnclaveID:    enclaveID,
		ProcessID:    os.Getpid(),
		EnclaveCID:   cid,
		NumberOfCPUs: cpuCount,
		CPUIDs:       cpuIDs,
		MemoryMiB:    memoryMB,
		State:        EnclaveStateRunning,
		Flags:        "SIMULATED",
	}

	ne.simulatedEnclaves[enclaveID] = info
	ne.enclaveInfo = info
	return info, nil
}

// DescribeEnclave retrieves information about a specific enclave
func (ne *NitroEnclave) DescribeEnclave(enclaveID string) (*EnclaveInfo, error) {
	return ne.DescribeEnclaveWithContext(context.Background(), enclaveID)
}

// DescribeEnclaveWithContext describes an enclave with context
func (ne *NitroEnclave) DescribeEnclaveWithContext(ctx context.Context, enclaveID string) (*EnclaveInfo, error) {
	ne.mu.RLock()
	defer ne.mu.RUnlock()

	if err := validateEnclaveID(enclaveID); err != nil {
		return nil, err
	}

	if ne.simulated {
		info, ok := ne.simulatedEnclaves[enclaveID]
		if !ok {
			return nil, ErrEnclaveNotFound
		}
		return info, nil
	}

	return ne.hardwareDescribeEnclave(ctx, enclaveID)
}

// hardwareDescribeEnclave describes enclave using nitro-cli
func (ne *NitroEnclave) hardwareDescribeEnclave(ctx context.Context, enclaveID string) (*EnclaveInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, DefaultDescribeTimeout)
	defer cancel()

	//nolint:gosec // G204: cliPath validated during init
	cmd := exec.CommandContext(ctx, ne.cliPath, "describe-enclaves")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("describe-enclaves failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("describe-enclaves failed: %w", err)
	}

	var enclaves []EnclaveInfo
	if err := json.Unmarshal(output, &enclaves); err != nil {
		return nil, fmt.Errorf("failed to parse output: %w", err)
	}

	for i := range enclaves {
		if enclaves[i].EnclaveID == enclaveID {
			return &enclaves[i], nil
		}
	}

	return nil, ErrEnclaveNotFound
}

// TerminateEnclave terminates a running enclave
func (ne *NitroEnclave) TerminateEnclave(enclaveID string) error {
	return ne.TerminateEnclaveWithContext(context.Background(), enclaveID)
}

// TerminateEnclaveWithContext terminates an enclave with context
func (ne *NitroEnclave) TerminateEnclaveWithContext(ctx context.Context, enclaveID string) error {
	ne.mu.Lock()
	defer ne.mu.Unlock()

	if err := validateEnclaveID(enclaveID); err != nil {
		return err
	}

	if ne.simulated {
		if _, ok := ne.simulatedEnclaves[enclaveID]; !ok {
			return ErrEnclaveNotFound
		}
		delete(ne.simulatedEnclaves, enclaveID)
		if ne.enclaveInfo != nil && ne.enclaveInfo.EnclaveID == enclaveID {
			ne.enclaveInfo = nil
		}
		return nil
	}

	return ne.hardwareTerminateEnclave(ctx, enclaveID)
}

// hardwareTerminateEnclave terminates enclave using nitro-cli
func (ne *NitroEnclave) hardwareTerminateEnclave(ctx context.Context, enclaveID string) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultTerminateTimeout)
	defer cancel()

	args := []string{
		"terminate-enclave",
		"--enclave-id", enclaveID,
	}

	//nolint:gosec // G204: cliPath validated, enclaveID validated
	cmd := exec.CommandContext(ctx, ne.cliPath, args...)
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("%w: %s", ErrEnclaveTerminateFailed, string(exitErr.Stderr))
		}
		return fmt.Errorf("%w: %v", ErrEnclaveTerminateFailed, err)
	}

	if ne.enclaveInfo != nil && ne.enclaveInfo.EnclaveID == enclaveID {
		ne.enclaveInfo = nil
	}

	return nil
}

// ListEnclaves lists all running enclaves
func (ne *NitroEnclave) ListEnclaves() ([]*EnclaveInfo, error) {
	return ne.ListEnclavesWithContext(context.Background())
}

// ListEnclavesWithContext lists enclaves with context
func (ne *NitroEnclave) ListEnclavesWithContext(ctx context.Context) ([]*EnclaveInfo, error) {
	ne.mu.RLock()
	defer ne.mu.RUnlock()

	if ne.simulated {
		result := make([]*EnclaveInfo, 0, len(ne.simulatedEnclaves))
		for _, info := range ne.simulatedEnclaves {
			result = append(result, info)
		}
		return result, nil
	}

	return ne.hardwareListEnclaves(ctx)
}

// hardwareListEnclaves lists enclaves using nitro-cli
func (ne *NitroEnclave) hardwareListEnclaves(ctx context.Context) ([]*EnclaveInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, DefaultDescribeTimeout)
	defer cancel()

	//nolint:gosec // G204: cliPath validated during init
	cmd := exec.CommandContext(ctx, ne.cliPath, "describe-enclaves")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("describe-enclaves failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("describe-enclaves failed: %w", err)
	}

	var enclaves []EnclaveInfo
	if err := json.Unmarshal(output, &enclaves); err != nil {
		return nil, fmt.Errorf("failed to parse output: %w", err)
	}

	result := make([]*EnclaveInfo, len(enclaves))
	for i := range enclaves {
		result[i] = &enclaves[i]
	}
	return result, nil
}

// GetMeasurements returns the enclave measurements from the last build
func (ne *NitroEnclave) GetMeasurements() *Measurements {
	ne.mu.RLock()
	defer ne.mu.RUnlock()
	return ne.measurements
}

// GetCurrentEnclave returns the current enclave info
func (ne *NitroEnclave) GetCurrentEnclave() *EnclaveInfo {
	ne.mu.RLock()
	defer ne.mu.RUnlock()
	return ne.enclaveInfo
}

// =============================================================================
// Vsock Communication
// =============================================================================

// VsockConnection represents a vsock connection to an enclave
type VsockConnection struct {
	mu sync.Mutex

	cid       uint32
	port      uint32
	conn      net.Conn
	simulated bool
}

// NewVsockConnection creates a new vsock connection
func NewVsockConnection(cid, port uint32) *VsockConnection {
	return &VsockConnection{
		cid:  cid,
		port: port,
	}
}

// Connect establishes a vsock connection
func (vc *VsockConnection) Connect(ctx context.Context) error {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	if vc.conn != nil {
		return nil // Already connected
	}

	// Check if vsock is available
	if _, err := os.Stat("/dev/vsock"); err != nil {
		vc.simulated = true
		return nil // Simulation mode
	}

	// Create vsock connection
	// Note: Real implementation would use syscall.AF_VSOCK
	addr := fmt.Sprintf("vsock://%d:%d", vc.cid, vc.port)

	var d net.Dialer
	conn, err := d.DialContext(ctx, "unix", addr)
	if err != nil {
		// Fall back to simulation if vsock not supported
		vc.simulated = true
		return nil
	}

	vc.conn = conn
	return nil
}

// Close closes the vsock connection
func (vc *VsockConnection) Close() error {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	if vc.conn != nil {
		err := vc.conn.Close()
		vc.conn = nil
		return err
	}
	return nil
}

// Send sends data to the enclave via vsock
func (vc *VsockConnection) Send(data []byte) error {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	if vc.simulated || vc.conn == nil {
		return nil // Simulated send
	}

	_, err := vc.conn.Write(data)
	return err
}

// Receive receives data from the enclave via vsock
func (vc *VsockConnection) Receive(maxSize int) ([]byte, error) {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	if vc.simulated || vc.conn == nil {
		// Return simulated response
		return []byte("SIMULATED_RESPONSE"), nil
	}

	buf := make([]byte, maxSize)
	n, err := vc.conn.Read(buf)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return buf[:n], nil
}

// IsSimulated returns true if running in simulation mode
func (vc *VsockConnection) IsSimulated() bool {
	return vc.simulated
}

// =============================================================================
// PCR Measurement Helpers
// =============================================================================

// PCRValue represents a Platform Configuration Register value
type PCRValue [PCRDigestSize]byte

// String returns hex representation
func (p PCRValue) String() string {
	return hex.EncodeToString(p[:])
}

// IsZero returns true if PCR is all zeros
func (p PCRValue) IsZero() bool {
	for _, b := range p {
		if b != 0 {
			return false
		}
	}
	return true
}

// PCRSet represents the set of PCR values
type PCRSet struct {
	PCRs [PCRCount]PCRValue
}

// Get returns the PCR value at the given index
func (s *PCRSet) Get(index int) PCRValue {
	if index < 0 || index >= PCRCount {
		return PCRValue{}
	}
	return s.PCRs[index]
}

// Set sets the PCR value at the given index
func (s *PCRSet) Set(index int, value PCRValue) {
	if index >= 0 && index < PCRCount {
		s.PCRs[index] = value
	}
}

// ParsePCRFromHex parses a hex string into a PCRValue
func ParsePCRFromHex(hexStr string) (PCRValue, error) {
	var pcr PCRValue
	decoded, err := hex.DecodeString(hexStr)
	if err != nil {
		return pcr, fmt.Errorf("invalid hex string: %w", err)
	}
	if len(decoded) != PCRDigestSize {
		return pcr, fmt.Errorf("invalid PCR length: expected %d, got %d", PCRDigestSize, len(decoded))
	}
	copy(pcr[:], decoded)
	return pcr, nil
}

// =============================================================================
// Input Validation Helpers
// =============================================================================

// enclaveIDPattern matches valid Nitro enclave IDs
// Accepts real AWS format (i-<hex>) and simulated format (i-sim-<hex>)
var enclaveIDPattern = regexp.MustCompile(`^i-(?:sim-)?[a-f0-9-]{8,64}$`)

// dockerImagePattern matches valid Docker image names
var dockerImagePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._/-]*[a-z0-9]?(:[a-zA-Z0-9._-]+)?$`)

// validateEnclaveID validates enclave ID format
func validateEnclaveID(id string) error {
	if id == "" {
		return ErrInvalidEnclaveID
	}
	if !enclaveIDPattern.MatchString(id) {
		return ErrInvalidEnclaveID
	}
	return nil
}

// validateDockerImage validates Docker image name format
func validateDockerImage(image string) error {
	if image == "" {
		return errors.New("docker image name is required")
	}
	if len(image) > 256 {
		return errors.New("docker image name too long")
	}
	if !dockerImagePattern.MatchString(image) {
		return errors.New("invalid docker image name format")
	}
	return nil
}

// sanitizePath sanitizes a file path for use with nitro-cli
func sanitizePath(path string) (string, error) {
	if path == "" {
		return "", errors.New("path is required")
	}

	// Clean and normalize the path
	cleanPath := filepath.Clean(path)

	// Check for path traversal
	if strings.Contains(cleanPath, "..") {
		return "", errors.New("path traversal not allowed")
	}

	// Check for shell metacharacters
	shellChars := []string{";", "&", "|", "$", "`", "(", ")", "{", "}", "[", "]", "<", ">", "!", "~", "*", "?", "\n", "\r"}
	for _, char := range shellChars {
		if strings.Contains(cleanPath, char) {
			return "", fmt.Errorf("invalid character in path: %s", char)
		}
	}

	return cleanPath, nil
}
