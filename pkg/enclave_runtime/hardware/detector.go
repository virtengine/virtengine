// Package hardware provides a comprehensive hardware abstraction layer for TEE integration.
//
// This file implements unified hardware detection across Intel SGX, AMD SEV-SNP,
// and AWS Nitro platforms. The detector probes for available hardware using
// CPUID instructions, device file checks, and sysfs probing.
//
// Task Reference: VE-2029 - Hardware TEE Integration Layer
package hardware

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

// =============================================================================
// CPUID Constants for SGX Detection
// =============================================================================

const (
	// CPUID leaf and bit positions for SGX detection
	cpuidLeafExtendedFeatures = 7
	cpuidSubleafDefault       = 0
	cpuidSGXBitEBX            = 2  // CPUID.07H.EBX[2] = SGX
	cpuidFLCBitECX            = 30 // CPUID.07H.ECX[30] = FLC

	cpuidLeafSGXInfo    = 0x12
	cpuidSGX1BitEAX     = 0 // CPUID.12H.0.EAX[0] = SGX1
	cpuidSGX2BitEAX     = 1 // CPUID.12H.0.EAX[1] = SGX2
	cpuidEnclavesBitEAX = 2 // CPUID.12H.0.EAX[2] = ENCLV instructions
)

// =============================================================================
// Unified Detector
// =============================================================================

// UnifiedDetector provides comprehensive hardware detection across all TEE platforms.
// It caches detection results and provides methods to query capabilities.
type UnifiedDetector struct {
	mu sync.RWMutex

	// Cached capabilities
	capabilities *HardwareCapabilities

	// Detection state
	detected        bool
	lastDetection   time.Time
	detectionErrors []string

	// Platform-specific detectors
	sgxDetector   *SGXDetector
	sevDetector   *SEVDetector
	nitroDetector *NitroDetector
}

// NewUnifiedDetector creates a new unified hardware detector.
func NewUnifiedDetector() *UnifiedDetector {
	return &UnifiedDetector{
		sgxDetector:   NewSGXDetector(),
		sevDetector:   NewSEVDetector(),
		nitroDetector: NewNitroDetector(),
	}
}

// Detect probes all TEE platforms and caches the results.
// This is safe to call multiple times; subsequent calls return cached results.
func (d *UnifiedDetector) Detect() (*HardwareCapabilities, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Return cached results if available and recent
	if d.detected && time.Since(d.lastDetection) < 5*time.Minute {
		return d.capabilities, nil
	}

	return d.detectInternal()
}

// ForceDetect forces a re-detection of hardware capabilities.
// This should be called if hardware state may have changed (e.g., driver loaded).
func (d *UnifiedDetector) ForceDetect() (*HardwareCapabilities, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.detectInternal()
}

// detectInternal performs the actual hardware detection.
// Must be called with mu held.
func (d *UnifiedDetector) detectInternal() (*HardwareCapabilities, error) {
	caps := &HardwareCapabilities{
		DetectedAt:      time.Now(),
		DetectionErrors: make([]string, 0),
		Platform:        PlatformSimulated,
	}

	// Only detect hardware on Linux
	if runtime.GOOS != "linux" {
		caps.DetectionErrors = append(caps.DetectionErrors,
			fmt.Sprintf("TEE hardware detection only supported on Linux (current: %s)", runtime.GOOS))
		d.capabilities = caps
		d.detected = true
		d.lastDetection = time.Now()
		return caps, nil
	}

	// Detect each platform
	var wg sync.WaitGroup
	var sgxCaps SGXCapabilities
	var sevCaps SEVSNPCapabilities
	var nitroCaps NitroCapabilities
	var sgxErr, sevErr, nitroErr error

	wg.Add(3)

	go func() {
		defer wg.Done()
		sgxCaps, sgxErr = d.sgxDetector.Detect()
	}()

	go func() {
		defer wg.Done()
		sevCaps, sevErr = d.sevDetector.Detect()
	}()

	go func() {
		defer wg.Done()
		nitroCaps, nitroErr = d.nitroDetector.Detect()
	}()

	wg.Wait()

	// Collect results and errors
	caps.SGX = sgxCaps
	if sgxErr != nil {
		caps.DetectionErrors = append(caps.DetectionErrors, fmt.Sprintf("SGX: %v", sgxErr))
	}

	caps.SEVSNP = sevCaps
	if sevErr != nil {
		caps.DetectionErrors = append(caps.DetectionErrors, fmt.Sprintf("SEV-SNP: %v", sevErr))
	}

	caps.Nitro = nitroCaps
	if nitroErr != nil {
		caps.DetectionErrors = append(caps.DetectionErrors, fmt.Sprintf("Nitro: %v", nitroErr))
	}

	// Set overall availability
	caps.Available = caps.HasAnyHardware()

	// Determine recommended platform
	caps.Platform = caps.GetRecommendedPlatform()

	// Determine overall TCB status
	caps.TCBStatus = d.aggregateTCBStatus(caps)

	d.capabilities = caps
	d.detected = true
	d.lastDetection = time.Now()
	d.detectionErrors = caps.DetectionErrors

	return caps, nil
}

// aggregateTCBStatus determines the overall TCB status from individual platforms.
func (d *UnifiedDetector) aggregateTCBStatus(caps *HardwareCapabilities) TCBStatus {
	// Use the status of the recommended platform
	switch caps.Platform {
	case PlatformSGX:
		return caps.SGX.TCBStatus
	case PlatformSEVSNP:
		return caps.SEVSNP.TCBStatus
	case PlatformNitro:
		// Nitro doesn't have TCB status in the same way
		return TCBStatusUpToDate
	default:
		return TCBStatusUnknown
	}
}

// GetCapabilities returns the cached hardware capabilities.
// Returns nil if detection has not been performed.
func (d *UnifiedDetector) GetCapabilities() *HardwareCapabilities {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.capabilities
}

// GetRecommendedPlatform returns the recommended platform based on detected hardware.
func (d *UnifiedDetector) GetRecommendedPlatform() Platform {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.capabilities == nil {
		return PlatformSimulated
	}
	return d.capabilities.GetRecommendedPlatform()
}

// IsHardwareAvailable returns true if any real TEE hardware is available.
func (d *UnifiedDetector) IsHardwareAvailable() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.capabilities == nil {
		return false
	}
	return d.capabilities.HasAnyHardware()
}

// GetErrors returns any errors encountered during detection.
func (d *UnifiedDetector) GetErrors() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.detectionErrors
}

// =============================================================================
// SGX Detector
// =============================================================================

// SGXDetector handles Intel SGX hardware detection.
type SGXDetector struct {
	mu sync.Mutex
}

// NewSGXDetector creates a new SGX detector.
func NewSGXDetector() *SGXDetector {
	return &SGXDetector{}
}

// Detect performs SGX hardware detection.
func (d *SGXDetector) Detect() (SGXCapabilities, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	caps := SGXCapabilities{}

	// Check CPUID for SGX support
	hasSGX := d.hasCPUIDSGX()
	hasFLC := d.hasCPUIDFLC()
	hasSGX2 := d.hasCPUIDSGX2()

	// Check device files regardless of CPUID (driver might still be present)
	enclaveExists, enclaveErr := checkDeviceExists(SGXDeviceEnclave)
	provisionExists, _ := checkDeviceExists(SGXDeviceProvision)
	legacyExists, _ := checkDeviceExists(SGXDeviceLegacy)

	// Determine device path
	var devicePath string
	if enclaveExists {
		devicePath = SGXDeviceEnclave
	} else if legacyExists {
		devicePath = SGXDeviceLegacy
	}

	if devicePath == "" && !hasSGX {
		return caps, fmt.Errorf("SGX not supported by CPU and no SGX device found")
	}

	if enclaveErr != nil && devicePath == "" {
		return caps, fmt.Errorf("error checking SGX device: %w", enclaveErr)
	}

	caps.Available = devicePath != "" || hasSGX
	caps.EnclaveDevice = devicePath
	caps.FLCSupported = hasFLC

	if provisionExists {
		caps.ProvisionDevice = SGXDeviceProvision
		caps.DCAPAvailable = true
	}

	// Determine SGX version
	caps.Version = 1
	if hasSGX2 {
		caps.Version = 2
	}

	// Check for AESM socket (indicates Intel PSW is installed)
	if _, err := os.Stat(SGXAESMSocket); err == nil {
		caps.AESMSocketPath = SGXAESMSocket
	}

	// Read EPC size from sysfs if available
	caps.EPCSize = d.readEPCSize()

	// Default TCB status
	caps.TCBStatus = TCBStatusUnknown

	return caps, nil
}

// hasCPUIDSGX checks if the CPU supports SGX via CPUID.
func (d *SGXDetector) hasCPUIDSGX() bool {
	result := cpuid(cpuidLeafExtendedFeatures, cpuidSubleafDefault)
	return (result.EBX & (1 << cpuidSGXBitEBX)) != 0
}

// hasCPUIDFLC checks if the CPU supports Flexible Launch Control.
func (d *SGXDetector) hasCPUIDFLC() bool {
	result := cpuid(cpuidLeafExtendedFeatures, cpuidSubleafDefault)
	return (result.ECX & (1 << cpuidFLCBitECX)) != 0
}

// hasCPUIDSGX2 checks if the CPU supports SGX2.
func (d *SGXDetector) hasCPUIDSGX2() bool {
	result := cpuid(cpuidLeafSGXInfo, cpuidSubleafDefault)
	return (result.EAX & (1 << cpuidSGX2BitEAX)) != 0
}

// readEPCSize attempts to read EPC size from sysfs.
func (d *SGXDetector) readEPCSize() uint64 {
	// EPC size might be available in various locations
	paths := []string{
		"/sys/devices/virtual/misc/sgx_enclave/epc_size",
		"/sys/module/intel_sgx/parameters/epc_size",
	}

	for _, path := range paths {
		if data, err := os.ReadFile(path); err == nil {
			var size uint64
			if _, err := fmt.Sscanf(string(data), "%d", &size); err == nil {
				return size
			}
		}
	}
	return 0
}

// =============================================================================
// SEV Detector
// =============================================================================

// SEVDetector handles AMD SEV-SNP hardware detection.
type SEVDetector struct {
	mu sync.Mutex
}

// NewSEVDetector creates a new SEV detector.
func NewSEVDetector() *SEVDetector {
	return &SEVDetector{}
}

// Detect performs SEV-SNP hardware detection.
func (d *SEVDetector) Detect() (SEVSNPCapabilities, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	caps := SEVSNPCapabilities{}

	// Check for SEV-guest device
	exists, err := checkDeviceExists(SEVGuestDevice)
	if err != nil {
		return caps, fmt.Errorf("error checking SEV device: %w", err)
	}

	if !exists {
		return caps, fmt.Errorf("SEV-SNP device not found at %s", SEVGuestDevice)
	}

	caps.Available = true
	caps.GuestDevice = SEVGuestDevice

	// Read platform status for version info
	if status, err := readSysFile(SEVPlatformStatus); err == nil {
		caps.Version = status
	} else {
		caps.Version = "unknown"
	}

	// Detect API version
	caps.APIVersion = d.detectAPIVersion()

	// Default TCB status
	caps.TCBStatus = TCBStatusUnknown

	return caps, nil
}

// detectAPIVersion attempts to detect the SEV API version.
func (d *SEVDetector) detectAPIVersion() int {
	firmwarePath := "/sys/class/firmware-attributes/sev/sev_version"
	if version, err := readSysFile(firmwarePath); err == nil {
		var major, minor int
		if _, err := fmt.Sscanf(version, "%d.%d", &major, &minor); err == nil {
			return major*100 + minor
		}
	}
	// Default to Milan API version
	return 151
}

// =============================================================================
// Nitro Detector
// =============================================================================

// NitroDetector handles AWS Nitro Enclave hardware detection.
type NitroDetector struct {
	mu sync.Mutex
}

// NewNitroDetector creates a new Nitro detector.
func NewNitroDetector() *NitroDetector {
	return &NitroDetector{}
}

// Detect performs Nitro Enclave hardware detection.
func (d *NitroDetector) Detect() (NitroCapabilities, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	caps := NitroCapabilities{}

	// Check for Nitro device
	exists, err := checkDeviceExists(NitroDevice)
	if err != nil {
		return caps, fmt.Errorf("error checking Nitro device: %w", err)
	}

	if !exists {
		return caps, fmt.Errorf("Nitro device not found at %s", NitroDevice)
	}

	caps.Available = true
	caps.DevicePath = NitroDevice

	// Check for NSM device
	if nsmExists, _ := checkDeviceExists(NitroNSMDevice); nsmExists {
		caps.NSMAvailable = true
	}

	// Check for nitro-cli
	cliPath, found := checkExecutableExists("nitro-cli")
	if found {
		caps.CLIPath = cliPath
		// Get version (simplified)
		caps.Version = d.getNitroCLIVersion(cliPath)
	}

	// Vsock is generally available when Nitro is available
	caps.VsockSupported = true

	return caps, nil
}

// getNitroCLIVersion gets the nitro-cli version.
func (d *NitroDetector) getNitroCLIVersion(cliPath string) string {
	// In a real implementation, we'd exec nitro-cli --version
	// For now, return a placeholder
	_ = cliPath
	return "unknown"
}

// =============================================================================
// Utility Functions
// =============================================================================

// checkDeviceExists checks if a device file exists and is accessible.
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
	// Check if it's a device file
	if info.Mode()&os.ModeDevice != 0 || info.Mode()&os.ModeCharDevice != 0 {
		return true, nil
	}
	return false, fmt.Errorf("path exists but is not a device: %s", path)
}

// checkExecutableExists checks if an executable exists in PATH or at an absolute path.
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
		"/sbin/" + name,
	}

	for _, path := range paths {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, true
		}
	}

	return "", false
}

// readSysFile reads a value from a sysfs file.
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
// CPUID Implementation
// =============================================================================

// CPUIDResult holds the result of a CPUID instruction.
type CPUIDResult struct {
	EAX, EBX, ECX, EDX uint32
}

// cpuid executes the CPUID instruction.
// This is a stub implementation. Real implementation would use assembly or cgo.
// On non-x86 platforms or when assembly is not available, returns zeros.
func cpuid(eax, ecx uint32) CPUIDResult {
	// Stub implementation - real implementation would be in assembly
	// For x86_64, this would look like:
	//
	// func cpuid(eax, ecx uint32) (CPUIDResult) {
	//     var result CPUIDResult
	//     asm volatile (
	//         "cpuid"
	//         : "=a"(result.EAX), "=b"(result.EBX), "=c"(result.ECX), "=d"(result.EDX)
	//         : "a"(eax), "c"(ecx)
	//     )
	//     return result
	// }
	//
	// For now, return zeros which means "not supported"
	_ = eax
	_ = ecx
	return CPUIDResult{}
}

// =============================================================================
// Global Detector Instance
// =============================================================================

var (
	globalDetector     *UnifiedDetector
	globalDetectorOnce sync.Once
)

// GetGlobalDetector returns the global unified detector instance.
// This instance is lazily initialized and cached.
func GetGlobalDetector() *UnifiedDetector {
	globalDetectorOnce.Do(func() {
		globalDetector = NewUnifiedDetector()
	})
	return globalDetector
}

// DetectHardware is a convenience function that uses the global detector.
func DetectHardware() (*HardwareCapabilities, error) {
	return GetGlobalDetector().Detect()
}

// RefreshHardwareDetection forces a re-detection using the global detector.
func RefreshHardwareDetection() (*HardwareCapabilities, error) {
	return GetGlobalDetector().ForceDetect()
}
