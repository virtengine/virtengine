// Package enclave_runtime provides TEE enclave implementations.
//
// This file provides a factory for creating enclave services based on
// configuration and hardware availability. It handles automatic platform
// detection and graceful fallback to simulation mode.
//
// Task Reference: SECURITY-002 - Real TEE Enclave Implementation
package enclave_runtime

import (
	"fmt"
	"sync"
)

// EnclaveFactory creates enclave services based on configuration and hardware availability
type EnclaveFactory struct {
	mu sync.RWMutex

	// Hardware capabilities (cached)
	capabilities *HardwareCapabilities

	// Default configuration
	defaultConfig RuntimeConfig

	// Default hardware mode
	defaultMode HardwareMode

	// Default enclave configuration
	sgxConfig   *SGXEnclaveConfig
	sevConfig   *SEVSNPConfig
	nitroConfig *NitroEnclaveConfig
}

// NewEnclaveFactory creates a new enclave factory with default settings
func NewEnclaveFactory() *EnclaveFactory {
	caps := DetectHardware()
	return &EnclaveFactory{
		capabilities:  &caps,
		defaultConfig: DefaultRuntimeConfig(),
		defaultMode:   HardwareModeAuto,
	}
}

// NewEnclaveFactoryWithConfig creates a factory with custom configuration
func NewEnclaveFactoryWithConfig(config EnclaveFactoryConfig) *EnclaveFactory {
	caps := DetectHardware()
	factory := &EnclaveFactory{
		capabilities:  &caps,
		defaultConfig: config.RuntimeConfig,
		defaultMode:   config.HardwareMode,
		sgxConfig:     config.SGXConfig,
		sevConfig:     config.SEVConfig,
		nitroConfig:   config.NitroConfig,
	}
	if factory.defaultConfig.MaxInputSize == 0 {
		factory.defaultConfig = DefaultRuntimeConfig()
	}
	return factory
}

// EnclaveFactoryConfig configures the enclave factory
type EnclaveFactoryConfig struct {
	// HardwareMode controls simulation vs hardware
	HardwareMode HardwareMode

	// RuntimeConfig for all enclave services
	RuntimeConfig RuntimeConfig

	// Platform-specific configurations
	SGXConfig   *SGXEnclaveConfig
	SEVConfig   *SEVSNPConfig
	NitroConfig *NitroEnclaveConfig
}

// CreateService creates an enclave service based on the preferred platform
// It automatically selects the best available platform based on hardware detection
func (f *EnclaveFactory) CreateService() (EnclaveService, error) {
	f.mu.RLock()
	caps := f.capabilities
	mode := f.defaultMode
	f.mu.RUnlock()

	// In simulate mode, always return simulated service
	if mode == HardwareModeSimulate {
		return f.createSimulatedService()
	}

	// Try to create a service based on available hardware
	if caps.SGXAvailable {
		svc, err := f.createSGXService()
		if err == nil {
			return svc, nil
		}
		if mode == HardwareModeRequire {
			return nil, fmt.Errorf("SGX hardware required but failed to initialize: %w", err)
		}
		// Fall through to try other platforms
	}

	if caps.SEVSNPAvailable {
		svc, err := f.createSEVService()
		if err == nil {
			return svc, nil
		}
		if mode == HardwareModeRequire {
			return nil, fmt.Errorf("SEV-SNP hardware required but failed to initialize: %w", err)
		}
		// Fall through to try other platforms
	}

	if caps.NitroAvailable {
		svc, err := f.createNitroService()
		if err == nil {
			return svc, nil
		}
		if mode == HardwareModeRequire {
			return nil, fmt.Errorf("Nitro hardware required but failed to initialize: %w", err)
		}
		// Fall through to simulation
	}

	// If require mode and no hardware worked
	if mode == HardwareModeRequire {
		return nil, fmt.Errorf("%w: no TEE hardware available", ErrHardwareNotAvailable)
	}

	// Fall back to simulation
	return f.createSimulatedService()
}

// CreateServiceForPlatform creates an enclave service for a specific platform
func (f *EnclaveFactory) CreateServiceForPlatform(platform AttestationType) (EnclaveService, error) {
	f.mu.RLock()
	mode := f.defaultMode
	f.mu.RUnlock()

	switch platform {
	case AttestationTypeSGX:
		return f.createSGXService()
	case AttestationTypeSEVSNP:
		return f.createSEVService()
	case AttestationTypeNitro:
		return f.createNitroService()
	case AttestationTypeSimulated:
		if mode == HardwareModeRequire {
			return nil, fmt.Errorf("hardware required but simulated requested")
		}
		return f.createSimulatedService()
	default:
		return nil, fmt.Errorf("unknown platform: %s", platform)
	}
}

// CreateHardwareAwareService creates a service that implements HardwareAwareEnclaveService
func (f *EnclaveFactory) CreateHardwareAwareService() (HardwareAwareEnclaveService, error) {
	f.mu.RLock()
	caps := f.capabilities
	mode := f.defaultMode
	f.mu.RUnlock()

	// In simulate mode, return a wrapper around simulated service
	if mode == HardwareModeSimulate {
		svc, err := f.createSimulatedService()
		if err != nil {
			return nil, err
		}
		return &simulatedHardwareWrapper{
			EnclaveService: svc,
			mode:           HardwareModeSimulate,
		}, nil
	}

	// Try SGX
	if caps.SGXAvailable {
		svc, err := f.createSGXServiceWithMode(mode)
		if err == nil {
			return svc, nil
		}
	}

	// Try SEV-SNP
	if caps.SEVSNPAvailable {
		svc, err := f.createSEVServiceWithMode(mode)
		if err == nil {
			return svc, nil
		}
	}

	// Try Nitro
	if caps.NitroAvailable {
		svc, err := f.createNitroServiceWithMode(mode)
		if err == nil {
			return svc, nil
		}
	}

	// Fall back to simulation if allowed
	if mode != HardwareModeRequire {
		svc, err := f.createSimulatedService()
		if err != nil {
			return nil, err
		}
		return &simulatedHardwareWrapper{
			EnclaveService: svc,
			mode:           HardwareModeSimulate,
		}, nil
	}

	return nil, ErrHardwareNotAvailable
}

// createSimulatedService creates a simulated enclave service
func (f *EnclaveFactory) createSimulatedService() (EnclaveService, error) {
	f.mu.RLock()
	config := f.defaultConfig
	f.mu.RUnlock()

	svc := NewSimulatedEnclaveService()
	if err := svc.Initialize(config); err != nil {
		return nil, fmt.Errorf("failed to initialize simulated service: %w", err)
	}
	return svc, nil
}

// createSGXService creates an SGX enclave service
func (f *EnclaveFactory) createSGXService() (EnclaveService, error) {
	return f.createSGXServiceWithMode(f.defaultMode)
}

// createSGXServiceWithMode creates an SGX enclave service with explicit mode
func (f *EnclaveFactory) createSGXServiceWithMode(mode HardwareMode) (*SGXEnclaveServiceImpl, error) {
	f.mu.RLock()
	config := f.sgxConfig
	runtimeConfig := f.defaultConfig
	f.mu.RUnlock()

	// Use default SGX config if not provided
	if config == nil {
		config = &SGXEnclaveConfig{
			EnclavePath: "/opt/virtengine/enclaves/veid_scorer.signed.so",
			DCAPEnabled: true,
			Debug:       false,
			MaxEPCPages: 256,
		}
	}

	svc, err := NewSGXEnclaveServiceImplWithMode(*config, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to create SGX service: %w", err)
	}

	if err := svc.Initialize(runtimeConfig); err != nil {
		return nil, fmt.Errorf("failed to initialize SGX service: %w", err)
	}

	return svc, nil
}

// createSEVService creates a SEV-SNP enclave service
func (f *EnclaveFactory) createSEVService() (EnclaveService, error) {
	return f.createSEVServiceWithMode(f.defaultMode)
}

// createSEVServiceWithMode creates a SEV-SNP service with explicit mode
func (f *EnclaveFactory) createSEVServiceWithMode(mode HardwareMode) (*SEVSNPEnclaveServiceImpl, error) {
	f.mu.RLock()
	config := f.sevConfig
	runtimeConfig := f.defaultConfig
	f.mu.RUnlock()

	// Use default SEV config if not provided
	if config == nil {
		config = &SEVSNPConfig{
			Endpoint:         "unix:///var/run/veid-enclave.sock",
			CertChainPath:    "/opt/virtengine/certs/amd-sev",
			VCEKCachePath:    "/var/cache/virtengine/vcek",
			MinTCBVersion:    "1.51",
			AllowDebugPolicy: false,
		}
	}

	svc, err := NewSEVSNPEnclaveServiceImplWithMode(*config, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to create SEV-SNP service: %w", err)
	}

	if err := svc.Initialize(runtimeConfig); err != nil {
		return nil, fmt.Errorf("failed to initialize SEV-SNP service: %w", err)
	}

	return svc, nil
}

// createNitroService creates a Nitro enclave service
func (f *EnclaveFactory) createNitroService() (EnclaveService, error) {
	return f.createNitroServiceWithMode(f.defaultMode)
}

// createNitroServiceWithMode creates a Nitro service with explicit mode
func (f *EnclaveFactory) createNitroServiceWithMode(mode HardwareMode) (*NitroEnclaveServiceImpl, error) {
	f.mu.RLock()
	config := f.nitroConfig
	runtimeConfig := f.defaultConfig
	f.mu.RUnlock()

	// Use default Nitro config if not provided
	if config == nil {
		config = &NitroEnclaveConfig{
			EnclaveImagePath: "virtengine/veid-enclave:latest",
			CPUCount:         2,
			MemoryMB:         2048,
		}
	}

	svc, err := NewNitroEnclaveServiceImplWithMode(*config, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to create Nitro service: %w", err)
	}

	if err := svc.Initialize(runtimeConfig); err != nil {
		return nil, fmt.Errorf("failed to initialize Nitro service: %w", err)
	}

	return svc, nil
}

// GetCapabilities returns the detected hardware capabilities
func (f *EnclaveFactory) GetCapabilities() HardwareCapabilities {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if f.capabilities == nil {
		return HardwareCapabilities{}
	}
	return *f.capabilities
}

// RefreshCapabilities re-detects hardware capabilities
func (f *EnclaveFactory) RefreshCapabilities() HardwareCapabilities {
	caps := RefreshHardwareDetection()
	f.mu.Lock()
	f.capabilities = &caps
	f.mu.Unlock()
	return caps
}

// SetHardwareMode sets the default hardware mode
func (f *EnclaveFactory) SetHardwareMode(mode HardwareMode) {
	f.mu.Lock()
	f.defaultMode = mode
	f.mu.Unlock()
}

// simulatedHardwareWrapper wraps a simulated service to implement HardwareAwareEnclaveService
type simulatedHardwareWrapper struct {
	EnclaveService
	mode HardwareMode
}

// IsHardwareEnabled returns false for simulated service
func (w *simulatedHardwareWrapper) IsHardwareEnabled() bool {
	return false
}

// GetHardwareMode returns the configured hardware mode
func (w *simulatedHardwareWrapper) GetHardwareMode() HardwareMode {
	return w.mode
}

// =============================================================================
// Global Factory Instance
// =============================================================================

var (
	globalFactory     *EnclaveFactory
	globalFactoryOnce sync.Once
)

// GetGlobalFactory returns the global enclave factory instance
func GetGlobalFactory() *EnclaveFactory {
	globalFactoryOnce.Do(func() {
		globalFactory = NewEnclaveFactory()
	})
	return globalFactory
}

// SetGlobalFactory sets the global factory instance (useful for testing)
func SetGlobalFactory(factory *EnclaveFactory) {
	globalFactory = factory
}

// =============================================================================
// Convenience Functions
// =============================================================================

// MustCreateService creates an enclave service or panics
func MustCreateService() EnclaveService {
	svc, err := GetGlobalFactory().CreateService()
	if err != nil {
		panic(fmt.Sprintf("failed to create enclave service: %v", err))
	}
	return svc
}

// CreateProductionService creates a service configured for production use
// It requires real hardware and fails if none is available.
// This function uses the global production configuration loaded from environment
// variables and enforces production security policies.
func CreateProductionService() (EnclaveService, error) {
	// Load production configuration from environment
	config := GetGlobalProductionConfig()
	if config == nil {
		// Fall back to default production config
		defaultConfig := DefaultProductionConfig()
		config = &defaultConfig
	}

	// Validate configuration for production readiness
	if config.Mode == TEEModeProduction {
		ready, issues := config.IsProductionReady()
		if !ready {
			fmt.Printf("WARNING: Production config has issues: %v\n", issues)
		}
	}

	// Create factory from production config
	factory, err := config.CreateFactory()
	if err != nil {
		return nil, fmt.Errorf("failed to create production factory: %w", err)
	}

	return factory.CreateService()
}

// CreateProductionServiceWithConfig creates a production service with explicit config
func CreateProductionServiceWithConfig(config *ProductionConfig) (EnclaveService, error) {
	if config == nil {
		return nil, fmt.Errorf("production config cannot be nil")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid production config: %w", err)
	}

	factory, err := config.CreateFactory()
	if err != nil {
		return nil, fmt.Errorf("failed to create factory: %w", err)
	}

	return factory.CreateService()
}

// CreateDevelopmentService creates a service configured for development
// It uses simulation mode if no hardware is available
func CreateDevelopmentService() (EnclaveService, error) {
	factory := NewEnclaveFactoryWithConfig(EnclaveFactoryConfig{
		HardwareMode:  HardwareModeAuto,
		RuntimeConfig: DefaultRuntimeConfig(),
	})
	return factory.CreateService()
}

// GetRecommendedPlatform returns the recommended TEE platform for this system
func GetRecommendedPlatform() AttestationType {
	caps := DetectHardware()
	return caps.PreferredBackend
}

// IsHardwareAvailable returns true if any TEE hardware is available
func IsHardwareAvailable() bool {
	caps := DetectHardware()
	return caps.HasAnyHardware()
}

// PrintHardwareStatus prints the current hardware status to stdout
func PrintHardwareStatus() {
	caps := DetectHardware()
	config := GetGlobalProductionConfig()

	fmt.Println("=== VirtEngine TEE Hardware Status ===")
	fmt.Printf("Hardware Available: %v\n", caps.HasAnyHardware())
	fmt.Printf("Preferred Platform: %s\n", caps.PreferredBackend)
	if config != nil {
		fmt.Printf("TEE Mode: %s\n", config.Mode)
		fmt.Printf("Require Hardware: %v\n", config.RequireHardware)
		fmt.Printf("Allow Debug: %v\n", config.AllowDebug)
	}
	fmt.Println()

	if caps.SGXAvailable {
		fmt.Println("[Intel SGX]")
		fmt.Printf("  Available: true\n")
		fmt.Printf("  Version: SGX%d\n", caps.SGXVersion)
		fmt.Printf("  FLC Support: %v\n", caps.SGXFLCSupported)
		fmt.Printf("  Driver: %s\n", caps.SGXDriverPath)
		if caps.SGXProvisionPath != "" {
			fmt.Printf("  Provision: %s\n", caps.SGXProvisionPath)
		}
	} else {
		fmt.Println("[Intel SGX] Not available")
	}

	if caps.SEVSNPAvailable {
		fmt.Println("[AMD SEV-SNP]")
		fmt.Printf("  Available: true\n")
		fmt.Printf("  Version: %s\n", caps.SEVSNPVersion)
		fmt.Printf("  Device: %s\n", caps.SEVGuestDevice)
		fmt.Printf("  API Version: %d\n", caps.SEVAPIVersion)
	} else {
		fmt.Println("[AMD SEV-SNP] Not available")
	}

	if caps.NitroAvailable {
		fmt.Println("[AWS Nitro]")
		fmt.Printf("  Available: true\n")
		fmt.Printf("  Version: %s\n", caps.NitroVersion)
		fmt.Printf("  Device: %s\n", caps.NitroDevice)
		fmt.Printf("  CLI: %s\n", caps.NitroCLIPath)
	} else {
		fmt.Println("[AWS Nitro] Not available")
	}

	if len(caps.DetectionErrors) > 0 {
		fmt.Println()
		fmt.Println("Detection Errors:")
		for _, err := range caps.DetectionErrors {
			fmt.Printf("  - %s\n", err)
		}
	}

	// Production readiness check
	if config != nil {
		ready, issues := config.IsProductionReady()
		fmt.Println()
		if ready {
			fmt.Println("Production Ready: YES")
		} else {
			fmt.Println("Production Ready: NO")
			for _, issue := range issues {
				fmt.Printf("  - %s\n", issue)
			}
		}
	}

	fmt.Println("======================================")
}

// =============================================================================
// Production Mode Integration
// =============================================================================

// InitializeProductionTEE initializes TEE services for production deployment
// This is the main entry point for production deployments
func InitializeProductionTEE() (*ProductionEnclaveService, error) {
	// Load configuration from environment
	config, err := LoadProductionConfigFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to load production config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid production config: %w", err)
	}

	// Check production readiness
	if config.Mode == TEEModeProduction {
		ready, issues := config.IsProductionReady()
		if !ready {
			return nil, fmt.Errorf("production config not ready: %v", issues)
		}

		// Verify hardware is available
		if err := RequireProductionHardware(); err != nil {
			return nil, err
		}
	}

	// Create production service
	return NewProductionEnclaveService(config)
}

