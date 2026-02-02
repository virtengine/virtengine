// Package enclave_runtime provides TEE enclave implementations.
//
// This file defines configuration types for loading enclave settings from
// application configuration files (TOML/YAML). These structs can be embedded
// in the main application configuration.
//
// Task Reference: SECURITY-002 - Real TEE Enclave Implementation
package enclave_runtime

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// modeSELinuxPermissive is the permissive attestation mode value
const modeSELinuxPermissive = "permissive"

// EnclaveRuntimeConfig is the top-level configuration for enclave runtime
// This can be embedded in app.toml or loaded from a separate config file
type EnclaveRuntimeConfig struct {
	// Enabled controls whether enclave scoring is enabled
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Platform selects the TEE platform: "auto", "sgx", "sev-snp", "nitro", "simulated"
	Platform string `mapstructure:"platform" json:"platform"`

	// Mode controls hardware requirements: "auto", "require", "simulate"
	Mode string `mapstructure:"mode" json:"mode"`

	// AllowedMeasurements lists the hex-encoded measurements that are trusted
	// Format: "platform:measurement" (e.g., "sgx:a1b2c3...")
	AllowedMeasurements []string `mapstructure:"allowed_measurements" json:"allowed_measurements"`

	// AttestationMode controls how attestation is verified:
	// "strict" = reject unknown measurements
	// "permissive" = log warning but accept (testnet only)
	AttestationMode string `mapstructure:"attestation_mode" json:"attestation_mode"`

	// AttestationCacheSeconds is how long to cache attestation reports
	AttestationCacheSeconds int `mapstructure:"attestation_cache_seconds" json:"attestation_cache_seconds"`

	// Endpoint is the gRPC endpoint for remote TEE service (if applicable)
	Endpoint string `mapstructure:"endpoint" json:"endpoint"`

	// MaxConcurrentRequests limits concurrent scoring requests
	MaxConcurrentRequests int `mapstructure:"max_concurrent_requests" json:"max_concurrent_requests"`

	// RequestTimeoutMs is the timeout for scoring requests in milliseconds
	RequestTimeoutMs int64 `mapstructure:"request_timeout_ms" json:"request_timeout_ms"`

	// Platform-specific configurations
	SGX   SGXConfig      `mapstructure:"sgx" json:"sgx"`
	SEV   SEVConfig      `mapstructure:"sev" json:"sev"`
	Nitro NitroConfigApp `mapstructure:"nitro" json:"nitro"`
}

// SGXConfig contains Intel SGX-specific configuration
type SGXConfig struct {
	// EnclavePath is the path to the signed enclave binary
	EnclavePath string `mapstructure:"enclave_path" json:"enclave_path"`

	// DCAPEnabled enables DCAP attestation (recommended over EPID)
	DCAPEnabled bool `mapstructure:"dcap_enabled" json:"dcap_enabled"`

	// SPIDHex is the Service Provider ID for EPID attestation (if DCAP disabled)
	SPIDHex string `mapstructure:"spid_hex" json:"spid_hex"`

	// QuoteProviderLibrary is the path to the quote provider library
	QuoteProviderLibrary string `mapstructure:"quote_provider_library" json:"quote_provider_library"`

	// MaxEPCPages is the maximum EPC pages to request
	MaxEPCPages int `mapstructure:"max_epc_pages" json:"max_epc_pages"`

	// Debug enables SGX debug mode (NOT FOR PRODUCTION)
	Debug bool `mapstructure:"debug" json:"debug"`
}

// SEVConfig contains AMD SEV-SNP-specific configuration
type SEVConfig struct {
	// Endpoint is the gRPC endpoint for the SEV-SNP enclave service
	Endpoint string `mapstructure:"endpoint" json:"endpoint"`

	// CertChainPath is the path to AMD certificate chain
	CertChainPath string `mapstructure:"cert_chain_path" json:"cert_chain_path"`

	// VCEKCachePath is the path to cache VCEK certificates
	VCEKCachePath string `mapstructure:"vcek_cache_path" json:"vcek_cache_path"`

	// MinTCBVersion is the minimum required TCB version
	MinTCBVersion string `mapstructure:"min_tcb_version" json:"min_tcb_version"`

	// AllowDebugPolicy allows debug-enabled guests (NOT FOR PRODUCTION)
	AllowDebugPolicy bool `mapstructure:"allow_debug_policy" json:"allow_debug_policy"`
}

// NitroConfigApp contains AWS Nitro-specific configuration for app.toml
// Note: This is distinct from NitroEnclaveConfig used by the enclave service
type NitroConfigApp struct {
	// EnclaveImagePath is the path to the Enclave Image File (EIF)
	EnclaveImagePath string `mapstructure:"enclave_image_path" json:"enclave_image_path"`

	// CPUCount is the number of vCPUs to allocate
	CPUCount int `mapstructure:"cpu_count" json:"cpu_count"`

	// MemoryMB is the memory to allocate in MB
	MemoryMB int `mapstructure:"memory_mb" json:"memory_mb"`

	// VsockPort is the vsock port for enclave communication
	VsockPort uint32 `mapstructure:"vsock_port" json:"vsock_port"`

	// DebugMode enables debug mode (NOT FOR PRODUCTION)
	DebugMode bool `mapstructure:"debug_mode" json:"debug_mode"`
}

// DefaultEnclaveRuntimeConfig returns sensible defaults for enclave configuration
func DefaultEnclaveRuntimeConfig() EnclaveRuntimeConfig {
	return EnclaveRuntimeConfig{
		Enabled:                 true,
		Platform:                "auto",
		Mode:                    "auto",
		AllowedMeasurements:     []string{},
		AttestationMode:         "strict",
		AttestationCacheSeconds: 300,
		Endpoint:                "unix:///var/run/veid-enclave.sock",
		MaxConcurrentRequests:   4,
		RequestTimeoutMs:        5000,
		SGX: SGXConfig{
			EnclavePath: "/opt/virtengine/enclaves/veid_scorer.signed.so",
			DCAPEnabled: true,
			MaxEPCPages: 256,
			Debug:       false,
		},
		SEV: SEVConfig{
			Endpoint:         "unix:///var/run/veid-enclave.sock",
			CertChainPath:    "/opt/virtengine/certs/amd-sev",
			VCEKCachePath:    "/var/cache/virtengine/vcek",
			MinTCBVersion:    "1.51",
			AllowDebugPolicy: false,
		},
		Nitro: NitroConfigApp{
			EnclaveImagePath: "/opt/virtengine/enclaves/veid_scorer.eif",
			CPUCount:         2,
			MemoryMB:         2048,
			VsockPort:        5000,
			DebugMode:        false,
		},
	}
}

// Validate validates the enclave runtime configuration
func (c *EnclaveRuntimeConfig) Validate() error {
	if !c.Enabled {
		return nil // Nothing to validate if disabled
	}

	// Validate platform
	platform := strings.ToLower(c.Platform)
	switch platform {
	case "auto", "sgx", "sev-snp", "nitro", "simulated":
		// Valid platforms
	default:
		return fmt.Errorf("invalid platform: %s (must be auto, sgx, sev-snp, nitro, or simulated)", c.Platform)
	}

	// Validate mode
	mode := strings.ToLower(c.Mode)
	switch mode {
	case "auto", "require", "simulate":
		// Valid modes
	default:
		return fmt.Errorf("invalid mode: %s (must be auto, require, or simulate)", c.Mode)
	}

	// Validate attestation mode
	attestationMode := strings.ToLower(c.AttestationMode)
	switch attestationMode {
	case "strict", "permissive":
		// Valid
	default:
		return fmt.Errorf("invalid attestation_mode: %s (must be strict or permissive)", c.AttestationMode)
	}

	// Warn about permissive mode
	if attestationMode == modeSELinuxPermissive {
		fmt.Println("WARNING: Enclave attestation mode is 'permissive' - this should only be used for testing")
	}

	// Validate concurrency
	if c.MaxConcurrentRequests <= 0 {
		c.MaxConcurrentRequests = 4
	}
	if c.MaxConcurrentRequests > 100 {
		return errors.New("max_concurrent_requests cannot exceed 100")
	}

	// Validate timeout
	if c.RequestTimeoutMs <= 0 {
		c.RequestTimeoutMs = 5000
	}
	if c.RequestTimeoutMs > 60000 {
		return errors.New("request_timeout_ms cannot exceed 60000 (60 seconds)")
	}

	// Platform-specific validation
	switch platform {
	case "sgx":
		if c.SGX.EnclavePath == "" {
			return errors.New("sgx.enclave_path is required when platform is 'sgx'")
		}
		if c.SGX.Debug {
			fmt.Println("WARNING: SGX debug mode enabled - NOT SECURE FOR PRODUCTION")
		}
	case "sev-snp":
		if c.SEV.Endpoint == "" {
			return errors.New("sev.endpoint is required when platform is 'sev-snp'")
		}
		if c.SEV.AllowDebugPolicy {
			fmt.Println("WARNING: SEV-SNP debug policy allowed - NOT SECURE FOR PRODUCTION")
		}
	case "nitro":
		if c.Nitro.EnclaveImagePath == "" {
			return errors.New("nitro.enclave_image_path is required when platform is 'nitro'")
		}
		if c.Nitro.DebugMode {
			fmt.Println("WARNING: Nitro debug mode enabled - NOT SECURE FOR PRODUCTION")
		}
	}

	return nil
}

// GetHardwareMode converts the mode string to HardwareMode enum
func (c *EnclaveRuntimeConfig) GetHardwareMode() HardwareMode {
	switch strings.ToLower(c.Mode) {
	case "require":
		return HardwareModeRequire
	case "simulate":
		return HardwareModeSimulate
	default:
		return HardwareModeAuto
	}
}

// GetPlatformType converts the platform string to AttestationType enum
func (c *EnclaveRuntimeConfig) GetPlatformType() AttestationType {
	switch strings.ToLower(c.Platform) {
	case "sgx":
		return AttestationTypeSGX
	case "sev-snp":
		return AttestationTypeSEVSNP
	case "nitro":
		return AttestationTypeNitro
	case "simulated":
		return AttestationTypeSimulated
	default:
		return AttestationTypeUnknown
	}
}

// GetRequestTimeout returns the request timeout as a Duration
func (c *EnclaveRuntimeConfig) GetRequestTimeout() time.Duration {
	return time.Duration(c.RequestTimeoutMs) * time.Millisecond
}

// GetAttestationCacheDuration returns the attestation cache duration
func (c *EnclaveRuntimeConfig) GetAttestationCacheDuration() time.Duration {
	return time.Duration(c.AttestationCacheSeconds) * time.Second
}

// ToRuntimeConfig converts to RuntimeConfig
func (c *EnclaveRuntimeConfig) ToRuntimeConfig() RuntimeConfig {
	return RuntimeConfig{
		MaxInputSize:          10 * 1024 * 1024, // 10 MB
		MaxExecutionTimeMs:    c.RequestTimeoutMs,
		MaxConcurrentRequests: c.MaxConcurrentRequests,
		ScrubIntervalMs:       0, // Scrub after each request
		ModelPath:             "/enclave/model/veid_scoring_v1.bin",
		KeyRotationEpoch:      1000,
	}
}

// ToSGXEnclaveConfig converts to SGXEnclaveConfig
func (c *EnclaveRuntimeConfig) ToSGXEnclaveConfig() *SGXEnclaveConfig {
	return &SGXEnclaveConfig{
		EnclavePath:          c.SGX.EnclavePath,
		SPIDHex:              c.SGX.SPIDHex,
		DCAPEnabled:          c.SGX.DCAPEnabled,
		QuoteProviderLibrary: c.SGX.QuoteProviderLibrary,
		Debug:                c.SGX.Debug,
		MaxEPCPages:          c.SGX.MaxEPCPages,
	}
}

// ToSEVSNPConfig converts to SEVSNPConfig
func (c *EnclaveRuntimeConfig) ToSEVSNPConfig() *SEVSNPConfig {
	return &SEVSNPConfig{
		Endpoint:         c.SEV.Endpoint,
		CertChainPath:    c.SEV.CertChainPath,
		VCEKCachePath:    c.SEV.VCEKCachePath,
		MinTCBVersion:    c.SEV.MinTCBVersion,
		AllowDebugPolicy: c.SEV.AllowDebugPolicy,
	}
}

// ToNitroEnclaveConfig converts to NitroEnclaveConfig
func (c *EnclaveRuntimeConfig) ToNitroEnclaveConfig() *NitroEnclaveConfig {
	return &NitroEnclaveConfig{
		EnclaveImagePath: c.Nitro.EnclaveImagePath,
		CPUCount:         c.Nitro.CPUCount,
		MemoryMB:         c.Nitro.MemoryMB,
		DebugMode:        c.Nitro.DebugMode,
		VsockPort:        c.Nitro.VsockPort,
	}
}

// CreateFactory creates an EnclaveFactory from this configuration
func (c *EnclaveRuntimeConfig) CreateFactory() (*EnclaveFactory, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return NewEnclaveFactoryWithConfig(EnclaveFactoryConfig{
		HardwareMode:  c.GetHardwareMode(),
		RuntimeConfig: c.ToRuntimeConfig(),
		SGXConfig:     c.ToSGXEnclaveConfig(),
		SEVConfig:     c.ToSEVSNPConfig(),
		NitroConfig:   c.ToNitroEnclaveConfig(),
	}), nil
}

// ParseMeasurements parses the allowed_measurements strings into Measurement objects
func (c *EnclaveRuntimeConfig) ParseMeasurements() ([]Measurement, error) {
	measurements := make([]Measurement, 0, len(c.AllowedMeasurements))

	for _, entry := range c.AllowedMeasurements {
		parts := strings.SplitN(entry, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid measurement format: %s (expected 'platform:hex_value')", entry)
		}

		platform := strings.ToLower(parts[0])
		hexValue := strings.TrimPrefix(parts[1], "0x")

		var platformType AttestationType
		switch platform {
		case "sgx":
			platformType = AttestationTypeSGX
		case "sev-snp", "sev":
			platformType = AttestationTypeSEVSNP
		case "nitro":
			platformType = AttestationTypeNitro
		case "simulated", "sim":
			platformType = AttestationTypeSimulated
		default:
			return nil, fmt.Errorf("unknown platform: %s", platform)
		}

		// Decode hex value
		value := make([]byte, len(hexValue)/2)
		for i := 0; i < len(hexValue)/2; i++ {
			var b byte
			_, err := fmt.Sscanf(hexValue[i*2:i*2+2], "%02x", &b)
			if err != nil {
				return nil, fmt.Errorf("invalid hex value in measurement: %s", entry)
			}
			value[i] = b
		}

		measurements = append(measurements, Measurement{
			Platform:    platformType,
			Value:       value,
			Description: fmt.Sprintf("Loaded from config: %s", entry),
			AddedAt:     time.Now(),
		})
	}

	return measurements, nil
}

// BuildMeasurementAllowlist creates a MeasurementAllowlist from the configuration
func (c *EnclaveRuntimeConfig) BuildMeasurementAllowlist() (*MeasurementAllowlist, error) {
	measurements, err := c.ParseMeasurements()
	if err != nil {
		return nil, err
	}

	allowlist := NewMeasurementAllowlist()
	for _, m := range measurements {
		if err := allowlist.AddMeasurement(m.Platform, m.Value, m.Description); err != nil {
			return nil, fmt.Errorf("failed to add measurement: %w", err)
		}
	}

	return allowlist, nil
}

// IsProductionReady returns true if the configuration is suitable for production
func (c *EnclaveRuntimeConfig) IsProductionReady() (bool, []string) {
	var issues []string

	// Check mode
	if c.Mode == "simulate" {
		issues = append(issues, "mode is 'simulate' - real hardware not required")
	}

	// Check attestation mode
	if c.AttestationMode == modeSELinuxPermissive {
		issues = append(issues, "attestation_mode is 'permissive' - unknown measurements will be accepted")
	}

	// Check measurements
	if len(c.AllowedMeasurements) == 0 {
		issues = append(issues, "no allowed_measurements defined - any measurement will be accepted")
	}

	// Check debug modes
	if c.SGX.Debug {
		issues = append(issues, "sgx.debug is enabled")
	}
	if c.SEV.AllowDebugPolicy {
		issues = append(issues, "sev.allow_debug_policy is enabled")
	}
	if c.Nitro.DebugMode {
		issues = append(issues, "nitro.debug_mode is enabled")
	}

	return len(issues) == 0, issues
}

// PrintConfigSummary prints a summary of the configuration
func (c *EnclaveRuntimeConfig) PrintConfigSummary() {
	fmt.Println("=== Enclave Runtime Configuration ===")
	fmt.Printf("Enabled: %v\n", c.Enabled)
	fmt.Printf("Platform: %s\n", c.Platform)
	fmt.Printf("Mode: %s\n", c.Mode)
	fmt.Printf("Attestation Mode: %s\n", c.AttestationMode)
	fmt.Printf("Attestation Cache: %ds\n", c.AttestationCacheSeconds)
	fmt.Printf("Max Concurrent Requests: %d\n", c.MaxConcurrentRequests)
	fmt.Printf("Request Timeout: %dms\n", c.RequestTimeoutMs)
	fmt.Printf("Allowed Measurements: %d\n", len(c.AllowedMeasurements))

	ready, issues := c.IsProductionReady()
	if ready {
		fmt.Println("Production Ready: YES")
	} else {
		fmt.Println("Production Ready: NO")
		for _, issue := range issues {
			fmt.Printf("  - %s\n", issue)
		}
	}
	fmt.Println("======================================")
}
