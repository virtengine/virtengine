// Package enclave_runtime provides TEE enclave implementations.
//
// This file provides production configuration and environment-based
// mode detection for TEE adapters. It enables automatic detection of
// production vs development environments and enforces security policies.
//
// Environment Variables:
//   - VIRTENGINE_TEE_MODE: "production", "development", or "testing"
//   - VIRTENGINE_TEE_PLATFORM: Force specific platform ("sgx", "sev-snp", "nitro")
//   - VIRTENGINE_TEE_REQUIRE_HARDWARE: "true" to fail if no hardware available
//   - VIRTENGINE_TEE_ALLOW_DEBUG: "true" to allow debug enclaves (NEVER in production)
//   - VIRTENGINE_TEE_MEASUREMENT_ALLOWLIST: Path to measurement allowlist JSON
//   - VIRTENGINE_TEE_ATTESTATION_ENDPOINT: Remote attestation service endpoint
//
// Task Reference: TEE-HARDWARE-001 - Deploy TEE adapters on production hardware
package enclave_runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// =============================================================================
// Production Mode Constants
// =============================================================================

// TEEMode represents the operational mode of the TEE subsystem
type TEEMode int

const (
	// TEEModeProduction enforces strict security policies
	TEEModeProduction TEEMode = iota
	// TEEModeDevelopment allows simulation and debug enclaves
	TEEModeDevelopment
	// TEEModeTesting is permissive for automated testing
	TEEModeTesting
)

// String returns the string representation of TEEMode
func (m TEEMode) String() string {
	switch m {
	case TEEModeProduction:
		return "production"
	case TEEModeDevelopment:
		return "development"
	case TEEModeTesting:
		return "testing"
	default:
		return "unknown"
	}
}

// ParseTEEMode parses a string into TEEMode
func ParseTEEMode(s string) (TEEMode, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "production", "prod":
		return TEEModeProduction, nil
	case "development", "dev":
		return TEEModeDevelopment, nil
	case "testing", "test":
		return TEEModeTesting, nil
	default:
		return TEEModeDevelopment, fmt.Errorf("unknown TEE mode: %s", s)
	}
}

// =============================================================================
// Environment Variable Names
// =============================================================================

const (
	EnvTEEMode                 = "VIRTENGINE_TEE_MODE"
	EnvTEEPlatform             = "VIRTENGINE_TEE_PLATFORM"
	EnvTEERequireHardware      = "VIRTENGINE_TEE_REQUIRE_HARDWARE"
	EnvTEEAllowDebug           = "VIRTENGINE_TEE_ALLOW_DEBUG"
	EnvTEEMeasurementAllowlist = "VIRTENGINE_TEE_MEASUREMENT_ALLOWLIST"
	EnvTEEAttestationEndpoint  = "VIRTENGINE_TEE_ATTESTATION_ENDPOINT"
	EnvTEECertCachePath        = "VIRTENGINE_TEE_CERT_CACHE_PATH"
	EnvTEEMinTCBVersion        = "VIRTENGINE_TEE_MIN_TCB_VERSION"
)

// =============================================================================
// Production Configuration
// =============================================================================

// ProductionConfig holds production-specific TEE configuration
type ProductionConfig struct {
	// Mode determines security policy strictness
	Mode TEEMode `json:"mode"`

	// ForcePlatform overrides automatic platform detection
	ForcePlatform string `json:"force_platform,omitempty"`

	// RequireHardware fails if no TEE hardware is available
	RequireHardware bool `json:"require_hardware"`

	// AllowDebug permits debug enclaves (NEVER true in production)
	AllowDebug bool `json:"allow_debug"`

	// MeasurementAllowlistPath is the path to trusted measurements JSON
	MeasurementAllowlistPath string `json:"measurement_allowlist_path,omitempty"`

	// AttestationEndpoint is the remote attestation verification service
	AttestationEndpoint string `json:"attestation_endpoint,omitempty"`

	// CertCachePath is where to cache attestation certificates
	CertCachePath string `json:"cert_cache_path,omitempty"`

	// MinTCBVersion is the minimum TCB version to accept
	MinTCBVersion string `json:"min_tcb_version,omitempty"`

	// AttestationCacheDuration is how long to cache attestation results
	AttestationCacheDuration time.Duration `json:"attestation_cache_duration"`

	// StrictMeasurementCheck rejects unknown measurements
	StrictMeasurementCheck bool `json:"strict_measurement_check"`

	// RequireRemoteAttestation requires verification via remote service
	RequireRemoteAttestation bool `json:"require_remote_attestation"`

	// Intel SGX specific
	SGX SGXProductionConfig `json:"sgx,omitempty"`

	// AMD SEV-SNP specific
	SEVSNP SEVSNPProductionConfig `json:"sev_snp,omitempty"`

	// AWS Nitro specific
	Nitro NitroProductionConfig `json:"nitro,omitempty"`
}

// SGXProductionConfig holds SGX-specific production settings
type SGXProductionConfig struct {
	// EnclavePath is the path to the signed enclave binary
	EnclavePath string `json:"enclave_path"`

	// PCCSEndpoint is the Provisioning Certificate Caching Service URL
	PCCSEndpoint string `json:"pccs_endpoint,omitempty"`

	// QuoteProviderLibrary is the path to the DCAP quote provider
	QuoteProviderLibrary string `json:"quote_provider_library,omitempty"`

	// RequireFLC requires Flexible Launch Control support
	RequireFLC bool `json:"require_flc"`

	// AllowedMRENCLAVEs is the list of trusted MRENCLAVE values (hex)
	AllowedMRENCLAVEs []string `json:"allowed_mrenclaves,omitempty"`

	// AllowedMRSIGNERs is the list of trusted MRSIGNER values (hex)
	AllowedMRSIGNERs []string `json:"allowed_mrsigners,omitempty"`
}

// SEVSNPProductionConfig holds SEV-SNP-specific production settings
type SEVSNPProductionConfig struct {
	// KDSBaseURL is the AMD Key Distribution Server base URL
	KDSBaseURL string `json:"kds_base_url"`

	// ProductName is the AMD processor product name (e.g., "Milan", "Genoa")
	ProductName string `json:"product_name"`

	// CertChainPath is the path to AMD certificate chain
	CertChainPath string `json:"cert_chain_path,omitempty"`

	// VCEKCachePath is the path to cache VCEK certificates
	VCEKCachePath string `json:"vcek_cache_path,omitempty"`

	// MinTCB specifies minimum TCB component versions
	MinTCB TCBRequirements `json:"min_tcb"`

	// AllowedLaunchDigests is the list of trusted launch measurements (hex)
	AllowedLaunchDigests []string `json:"allowed_launch_digests,omitempty"`
}

// TCBRequirements specifies minimum TCB component versions
type TCBRequirements struct {
	BootLoader uint8 `json:"boot_loader"`
	TEE        uint8 `json:"tee"`
	SNP        uint8 `json:"snp"`
	Microcode  uint8 `json:"microcode"`
}

// NitroProductionConfig holds Nitro-specific production settings
type NitroProductionConfig struct {
	// EnclaveImagePath is the path to the Enclave Image File (EIF)
	EnclaveImagePath string `json:"enclave_image_path"`

	// CPUCount is the number of vCPUs to allocate
	CPUCount int `json:"cpu_count"`

	// MemoryMB is the memory to allocate in MB
	MemoryMB int `json:"memory_mb"`

	// AllowedPCR0s is the list of trusted PCR0 values (hex)
	AllowedPCR0s []string `json:"allowed_pcr0s,omitempty"`

	// AllowedPCR1s is the list of trusted PCR1 values (hex)
	AllowedPCR1s []string `json:"allowed_pcr1s,omitempty"`

	// AllowedPCR2s is the list of trusted PCR2 values (hex)
	AllowedPCR2s []string `json:"allowed_pcr2s,omitempty"`

	// KMSKeyARN is the AWS KMS key ARN for cryptographic operations
	KMSKeyARN string `json:"kms_key_arn,omitempty"`
}

// =============================================================================
// Default Production Configurations
// =============================================================================

// DefaultProductionConfig returns a secure default production configuration
func DefaultProductionConfig() ProductionConfig {
	return ProductionConfig{
		Mode:                     TEEModeProduction,
		RequireHardware:          true,
		AllowDebug:               false,
		CertCachePath:            "/var/cache/virtengine/tee-certs",
		AttestationCacheDuration: 5 * time.Minute,
		StrictMeasurementCheck:   true,
		RequireRemoteAttestation: false, // Enable when remote service is deployed
		SGX: SGXProductionConfig{
			EnclavePath:          "/opt/virtengine/enclaves/veid_scorer.signed.so",
			PCCSEndpoint:         "https://localhost:8081/sgx/certification/v4/",
			QuoteProviderLibrary: "/usr/lib/x86_64-linux-gnu/libsgx_dcap_ql.so",
			RequireFLC:           true,
		},
		SEVSNP: SEVSNPProductionConfig{
			KDSBaseURL:  "https://kdsintf.amd.com/vcek/v1",
			ProductName: "Milan",
			MinTCB: TCBRequirements{
				BootLoader: 2,
				TEE:        0,
				SNP:        8,
				Microcode:  115,
			},
		},
		Nitro: NitroProductionConfig{
			EnclaveImagePath: "/opt/virtengine/enclaves/veid_scorer.eif",
			CPUCount:         2,
			MemoryMB:         2048,
		},
	}
}

// DefaultDevelopmentConfig returns a permissive development configuration
func DefaultDevelopmentConfig() ProductionConfig {
	return ProductionConfig{
		Mode:                     TEEModeDevelopment,
		RequireHardware:          false,
		AllowDebug:               true,
		CertCachePath:            "/tmp/virtengine/tee-certs",
		AttestationCacheDuration: 1 * time.Hour,
		StrictMeasurementCheck:   false,
		RequireRemoteAttestation: false,
		SGX: SGXProductionConfig{
			EnclavePath: "./build/enclaves/veid_scorer.signed.so",
			RequireFLC:  false,
		},
		SEVSNP: SEVSNPProductionConfig{
			KDSBaseURL:  "https://kdsintf.amd.com/vcek/v1",
			ProductName: "Milan",
		},
		Nitro: NitroProductionConfig{
			EnclaveImagePath: "./build/enclaves/veid_scorer.eif",
			CPUCount:         2,
			MemoryMB:         512,
		},
	}
}

// =============================================================================
// Configuration Loading
// =============================================================================

var (
	globalProductionConfig *ProductionConfig
	productionConfigMu     sync.RWMutex
	productionConfigOnce   sync.Once
)

// LoadProductionConfigFromEnv loads configuration from environment variables
func LoadProductionConfigFromEnv() (*ProductionConfig, error) {
	config := DefaultDevelopmentConfig()

	// Load mode
	if modeStr := os.Getenv(EnvTEEMode); modeStr != "" {
		mode, err := ParseTEEMode(modeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %s: %w", EnvTEEMode, err)
		}
		config.Mode = mode

		// Apply mode-specific defaults
		if mode == TEEModeProduction {
			config = DefaultProductionConfig()
		}
	}

	// Override with specific environment variables
	if platform := os.Getenv(EnvTEEPlatform); platform != "" {
		config.ForcePlatform = strings.ToLower(platform)
	}

	if requireHW := os.Getenv(EnvTEERequireHardware); requireHW != "" {
		config.RequireHardware = strings.ToLower(requireHW) == "true"
	}

	if allowDebug := os.Getenv(EnvTEEAllowDebug); allowDebug != "" {
		config.AllowDebug = strings.ToLower(allowDebug) == "true"
	}

	if allowlistPath := os.Getenv(EnvTEEMeasurementAllowlist); allowlistPath != "" {
		config.MeasurementAllowlistPath = allowlistPath
	}

	if endpoint := os.Getenv(EnvTEEAttestationEndpoint); endpoint != "" {
		config.AttestationEndpoint = endpoint
	}

	if certCache := os.Getenv(EnvTEECertCachePath); certCache != "" {
		config.CertCachePath = certCache
	}

	if minTCB := os.Getenv(EnvTEEMinTCBVersion); minTCB != "" {
		config.MinTCBVersion = minTCB
	}

	return &config, nil
}

// LoadProductionConfigFromFile loads configuration from a JSON file
func LoadProductionConfigFromFile(path string) (*ProductionConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ProductionConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// GetGlobalProductionConfig returns the global production configuration
func GetGlobalProductionConfig() *ProductionConfig {
	productionConfigOnce.Do(func() {
		config, err := LoadProductionConfigFromEnv()
		if err != nil {
			fmt.Printf("WARNING: Failed to load production config from env: %v\n", err)
			defaultConfig := DefaultDevelopmentConfig()
			config = &defaultConfig
		}
		globalProductionConfig = config
	})

	productionConfigMu.RLock()
	defer productionConfigMu.RUnlock()
	return globalProductionConfig
}

// SetGlobalProductionConfig sets the global production configuration
func SetGlobalProductionConfig(config *ProductionConfig) {
	productionConfigMu.Lock()
	defer productionConfigMu.Unlock()
	globalProductionConfig = config
}

// =============================================================================
// Configuration Validation
// =============================================================================

// Validate validates the production configuration
func (c *ProductionConfig) Validate() error {
	var errs []string

	// Production mode checks
	if c.Mode == TEEModeProduction {
		if c.AllowDebug {
			errs = append(errs, "AllowDebug must be false in production mode")
		}
		if !c.RequireHardware {
			errs = append(errs, "RequireHardware must be true in production mode")
		}
		if !c.StrictMeasurementCheck {
			errs = append(errs, "StrictMeasurementCheck must be true in production mode")
		}
	}

	// Validate platform if specified
	if c.ForcePlatform != "" {
		switch c.ForcePlatform {
		case "sgx", "sev-snp", "nitro", "simulated":
			// Valid
		default:
			errs = append(errs, fmt.Sprintf("invalid ForcePlatform: %s", c.ForcePlatform))
		}

		// Can't force simulated in production
		if c.Mode == TEEModeProduction && c.ForcePlatform == "simulated" {
			errs = append(errs, "cannot force simulated platform in production mode")
		}
	}

	// Validate SGX config
	if c.ForcePlatform == "sgx" || c.ForcePlatform == "" {
		if c.SGX.EnclavePath == "" {
			errs = append(errs, "SGX.EnclavePath is required")
		}
	}

	// Validate SEV-SNP config
	if c.ForcePlatform == "sev-snp" || c.ForcePlatform == "" {
		if c.SEVSNP.KDSBaseURL == "" {
			errs = append(errs, "SEVSNP.KDSBaseURL is required")
		}
		if c.SEVSNP.ProductName == "" {
			errs = append(errs, "SEVSNP.ProductName is required")
		}
	}

	// Validate Nitro config
	if c.ForcePlatform == "nitro" || c.ForcePlatform == "" {
		if c.Nitro.EnclaveImagePath == "" {
			errs = append(errs, "Nitro.EnclaveImagePath is required")
		}
		if c.Nitro.CPUCount <= 0 {
			c.Nitro.CPUCount = 2
		}
		if c.Nitro.MemoryMB <= 0 {
			c.Nitro.MemoryMB = 2048
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errs, "; "))
	}

	return nil
}

// IsProductionReady returns true if the configuration is suitable for production
func (c *ProductionConfig) IsProductionReady() (bool, []string) {
	var issues []string

	if c.Mode != TEEModeProduction {
		issues = append(issues, fmt.Sprintf("mode is %s (expected production)", c.Mode))
	}

	if c.AllowDebug {
		issues = append(issues, "debug enclaves are allowed")
	}

	if !c.RequireHardware {
		issues = append(issues, "hardware is not required")
	}

	if !c.StrictMeasurementCheck {
		issues = append(issues, "strict measurement check is disabled")
	}

	if c.MeasurementAllowlistPath == "" {
		issues = append(issues, "no measurement allowlist configured")
	}

	return len(issues) == 0, issues
}

// =============================================================================
// Factory Integration
// =============================================================================

// ToEnclaveFactoryConfig converts ProductionConfig to EnclaveFactoryConfig
func (c *ProductionConfig) ToEnclaveFactoryConfig() EnclaveFactoryConfig {
	// Determine hardware mode
	var hwMode HardwareMode
	switch {
	case c.RequireHardware:
		hwMode = HardwareModeRequire
	case c.Mode == TEEModeProduction:
		hwMode = HardwareModeAuto
	default:
		hwMode = HardwareModeAuto
	}

	return EnclaveFactoryConfig{
		HardwareMode: hwMode,
		RuntimeConfig: RuntimeConfig{
			MaxInputSize:          10 * 1024 * 1024,
			MaxExecutionTimeMs:    5000,
			MaxConcurrentRequests: 4,
			ScrubIntervalMs:       0,
			ModelPath:             "/enclave/model/veid_scoring_v1.bin",
			KeyRotationEpoch:      1000,
		},
		SGXConfig: &SGXEnclaveConfig{
			EnclavePath:          c.SGX.EnclavePath,
			DCAPEnabled:          true,
			QuoteProviderLibrary: c.SGX.QuoteProviderLibrary,
			Debug:                c.AllowDebug,
		},
		SEVConfig: &SEVSNPConfig{
			Endpoint:         "unix:///var/run/veid-enclave.sock",
			CertChainPath:    c.SEVSNP.CertChainPath,
			VCEKCachePath:    c.SEVSNP.VCEKCachePath,
			MinTCBVersion:    c.MinTCBVersion,
			AllowDebugPolicy: c.AllowDebug,
		},
		NitroConfig: &NitroEnclaveConfig{
			EnclaveImagePath: c.Nitro.EnclaveImagePath,
			CPUCount:         c.Nitro.CPUCount,
			MemoryMB:         c.Nitro.MemoryMB,
			DebugMode:        c.AllowDebug,
		},
	}
}

// CreateFactory creates an EnclaveFactory from this production config
func (c *ProductionConfig) CreateFactory() (*EnclaveFactory, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("invalid production config: %w", err)
	}

	factoryConfig := c.ToEnclaveFactoryConfig()
	return NewEnclaveFactoryWithConfig(factoryConfig), nil
}

// ToVerificationPolicy converts ProductionConfig to VerificationPolicy
func (c *ProductionConfig) ToVerificationPolicy() VerificationPolicy {
	policy := DefaultVerificationPolicy()

	policy.AllowDebugMode = c.AllowDebug
	policy.RequireLatestTCB = c.Mode == TEEModeProduction
	policy.MaxAttestationAge = c.AttestationCacheDuration

	// Set allowed platforms based on mode
	if c.Mode == TEEModeProduction {
		policy.AllowedPlatforms = []AttestationType{
			AttestationTypeSGX,
			AttestationTypeSEVSNP,
			AttestationTypeNitro,
		}
	} else {
		policy.AllowedPlatforms = []AttestationType{
			AttestationTypeSGX,
			AttestationTypeSEVSNP,
			AttestationTypeNitro,
			AttestationTypeSimulated,
		}
	}

	return policy
}

// =============================================================================
// Production Service Initializer
// =============================================================================

// ProductionEnclaveService wraps an enclave service with production controls
type ProductionEnclaveService struct {
	service  EnclaveService
	config   *ProductionConfig
	verifier *UniversalAttestationVerifier
	mu       sync.RWMutex
}

// NewProductionEnclaveService creates a production-ready enclave service
func NewProductionEnclaveService(config *ProductionConfig) (*ProductionEnclaveService, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid production config: %w", err)
	}

	// Create factory from config
	factory, err := config.CreateFactory()
	if err != nil {
		return nil, fmt.Errorf("failed to create factory: %w", err)
	}

	// Create the underlying service
	service, err := factory.CreateService()
	if err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	// Load measurement allowlist if configured
	var allowlist *MeasurementAllowlist
	if config.MeasurementAllowlistPath != "" {
		allowlist = NewMeasurementAllowlist()
		if err := allowlist.LoadFromJSON(config.MeasurementAllowlistPath); err != nil {
			if config.Mode == TEEModeProduction {
				return nil, fmt.Errorf("failed to load measurement allowlist: %w", err)
			}
			fmt.Printf("WARNING: Failed to load measurement allowlist: %v\n", err)
		}
	} else {
		allowlist = NewMeasurementAllowlist()
	}

	// Create attestation verifier
	verifier := NewUniversalAttestationVerifier(allowlist)

	return &ProductionEnclaveService{
		service:  service,
		config:   config,
		verifier: verifier,
	}, nil
}

// GetService returns the underlying enclave service
func (p *ProductionEnclaveService) GetService() EnclaveService {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.service
}

// GetConfig returns the production configuration
func (p *ProductionEnclaveService) GetConfig() *ProductionConfig {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.config
}

// GetVerifier returns the attestation verifier
func (p *ProductionEnclaveService) GetVerifier() *UniversalAttestationVerifier {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.verifier
}

// VerifyAttestation verifies an attestation using the configured policy
func (p *ProductionEnclaveService) VerifyAttestation(attestation []byte, nonce []byte) (*VerificationResult, error) {
	p.mu.RLock()
	policy := p.config.ToVerificationPolicy()
	verifier := p.verifier
	p.mu.RUnlock()

	return verifier.Verify(attestation, nonce, policy)
}

// Shutdown gracefully shuts down the production service
func (p *ProductionEnclaveService) Shutdown() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.service != nil {
		return p.service.Shutdown()
	}
	return nil
}

// =============================================================================
// Production Status Reporting
// =============================================================================

// ProductionStatus represents the status of the production TEE service
type ProductionStatus struct {
	Mode                 string    `json:"mode"`
	Platform             string    `json:"platform"`
	HardwareEnabled      bool      `json:"hardware_enabled"`
	Initialized          bool      `json:"initialized"`
	DebugMode            bool      `json:"debug_mode"`
	ProductionReady      bool      `json:"production_ready"`
	ProductionIssues     []string  `json:"production_issues,omitempty"`
	MeasurementsLoaded   int       `json:"measurements_loaded"`
	AttestationEndpoint  string    `json:"attestation_endpoint,omitempty"`
	LastHealthCheck      time.Time `json:"last_health_check"`
	HealthCheckStatus    string    `json:"health_check_status"`
}

// GetStatus returns the current production status
func (p *ProductionEnclaveService) GetStatus() ProductionStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()

	status := ProductionStatus{
		Mode:            p.config.Mode.String(),
		DebugMode:       p.config.AllowDebug,
		LastHealthCheck: time.Now(),
	}

	// Get service status
	if p.service != nil {
		svcStatus := p.service.GetStatus()
		status.Initialized = svcStatus.Initialized

		// Try to get platform type via interface assertion
		type platformAware interface {
			GetPlatformType() PlatformType
		}
		if pa, ok := p.service.(platformAware); ok {
			status.Platform = string(pa.GetPlatformType())
		} else {
			status.Platform = "unknown"
		}

		// Check if hardware is enabled
		if hwService, ok := p.service.(HardwareAwareEnclaveService); ok {
			status.HardwareEnabled = hwService.IsHardwareEnabled()
		}
	}

	// Check production readiness
	ready, issues := p.config.IsProductionReady()
	status.ProductionReady = ready
	status.ProductionIssues = issues

	// Set health check status
	if status.Initialized && (status.HardwareEnabled || p.config.Mode != TEEModeProduction) {
		status.HealthCheckStatus = "healthy"
	} else if status.Initialized {
		status.HealthCheckStatus = "degraded"
	} else {
		status.HealthCheckStatus = "unhealthy"
	}

	return status
}

// =============================================================================
// Utility Functions
// =============================================================================

// MustGetProductionService creates a production service or panics
func MustGetProductionService() *ProductionEnclaveService {
	config := GetGlobalProductionConfig()
	service, err := NewProductionEnclaveService(config)
	if err != nil {
		panic(fmt.Sprintf("failed to create production enclave service: %v", err))
	}
	return service
}

// IsProductionMode returns true if running in production mode
func IsProductionMode() bool {
	config := GetGlobalProductionConfig()
	return config != nil && config.Mode == TEEModeProduction
}

// RequireProductionHardware returns an error if production mode but no hardware
func RequireProductionHardware() error {
	config := GetGlobalProductionConfig()
	if config == nil || config.Mode != TEEModeProduction {
		return nil
	}

	caps := DetectHardware()
	if !caps.HasAnyHardware() {
		return errors.New("production mode requires TEE hardware, but none detected")
	}

	return nil
}

// ParseMinTCBVersion parses a TCB version string into TCBRequirements
func ParseMinTCBVersion(s string) (TCBRequirements, error) {
	var tcb TCBRequirements

	// Format: "BL.TEE.SNP.UC" or "bootloader.tee.snp.microcode"
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return tcb, fmt.Errorf("invalid TCB version format: %s (expected BL.TEE.SNP.UC)", s)
	}

	bl, err := strconv.ParseUint(parts[0], 10, 8)
	if err != nil {
		return tcb, fmt.Errorf("invalid bootloader version: %s", parts[0])
	}
	tcb.BootLoader = uint8(bl)

	tee, err := strconv.ParseUint(parts[1], 10, 8)
	if err != nil {
		return tcb, fmt.Errorf("invalid TEE version: %s", parts[1])
	}
	tcb.TEE = uint8(tee)

	snp, err := strconv.ParseUint(parts[2], 10, 8)
	if err != nil {
		return tcb, fmt.Errorf("invalid SNP version: %s", parts[2])
	}
	tcb.SNP = uint8(snp)

	uc, err := strconv.ParseUint(parts[3], 10, 8)
	if err != nil {
		return tcb, fmt.Errorf("invalid microcode version: %s", parts[3])
	}
	tcb.Microcode = uint8(uc)

	return tcb, nil
}
