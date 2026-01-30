// Package enclave_runtime provides TEE enclave implementations.
//
// This file defines the interface and stubs for real TEE integration
// (Intel SGX and AMD SEV-SNP). The actual implementations will be
// added as part of the full TEE integration project.
//
// Task Reference: VE-2023 - TEE integration planning and POC
package enclave_runtime

import (
	"context"
	"errors"
	"time"
)

// =============================================================================
// Platform Types
// =============================================================================

// PlatformType identifies the TEE platform
type PlatformType string

const (
	// PlatformSimulated is for testing/development only (NOT SECURE)
	PlatformSimulated PlatformType = "simulated"

	// PlatformSGX is Intel Software Guard Extensions
	PlatformSGX PlatformType = "sgx"

	// PlatformSEVSNP is AMD SEV-SNP (Secure Encrypted Virtualization)
	PlatformSEVSNP PlatformType = "sev-snp"

	// PlatformNitro is AWS Nitro Enclaves
	PlatformNitro PlatformType = "nitro"
)

// String returns the string representation
func (p PlatformType) String() string {
	return string(p)
}

// IsSecure returns true if the platform provides real security guarantees
func (p PlatformType) IsSecure() bool {
	switch p {
	case PlatformSGX, PlatformSEVSNP, PlatformNitro:
		return true
	default:
		return false
	}
}

// =============================================================================
// Attestation Types
// =============================================================================

// AttestationReport contains a platform-specific attestation report
type AttestationReport struct {
	// Platform identifies the TEE platform
	Platform PlatformType

	// Measurement is the enclave/VM measurement hash
	Measurement []byte

	// ReportData is user-supplied data bound to the report (nonce)
	ReportData []byte

	// PlatformInfo contains platform-specific information
	// - SGX: CPUSVN, ISV SVN, etc.
	// - SEV-SNP: TCB version, guest policy, etc.
	PlatformInfo []byte

	// Signature is the platform signature over the report
	Signature []byte

	// CertChain is the certificate chain for verification
	// - SGX: PCK cert chain to Intel root
	// - SEV-SNP: VCEK cert chain to AMD root
	CertChain [][]byte

	// Timestamp is when the report was generated
	Timestamp time.Time
}

// Validate performs basic validation of the attestation report
func (r *AttestationReport) Validate() error {
	if r.Platform == "" {
		return errors.New("platform is required")
	}
	if len(r.Measurement) == 0 {
		return errors.New("measurement is required")
	}
	if len(r.Signature) == 0 {
		return errors.New("signature is required")
	}
	return nil
}

// =============================================================================
// Real Enclave Service Interface
// =============================================================================

// RealEnclaveService extends EnclaveService with real TEE capabilities
type RealEnclaveService interface {
	EnclaveService

	// GetAttestationReport generates a platform-specific attestation report
	// The nonce is included in ReportData to prevent replay attacks
	GetAttestationReport(nonce []byte) (*AttestationReport, error)

	// VerifyPeerAttestation verifies another enclave's attestation report
	VerifyPeerAttestation(report *AttestationReport) error

	// GetPlatformType returns the TEE platform type
	GetPlatformType() PlatformType

	// IsPlatformSecure returns true if this is a real (non-simulated) TEE
	IsPlatformSecure() bool

	// GetTCBInfo returns Trusted Computing Base information
	GetTCBInfo() (*TCBInfo, error)
}

// TCBInfo contains Trusted Computing Base information
type TCBInfo struct {
	// Platform is the TEE platform
	Platform PlatformType

	// Version is the TCB version string
	Version string

	// SVN is the Security Version Number
	SVN uint64

	// Flags contains platform-specific flags
	Flags map[string]bool

	// LastUpdated is when the TCB was last updated
	LastUpdated time.Time
}

// =============================================================================
// SGX Enclave Service (Stub)
// =============================================================================

// SGXEnclaveConfig configures the SGX enclave service
type SGXEnclaveConfig struct {
	// EnclavePath is the path to the signed enclave binary
	EnclavePath string

	// SPIDHex is the Service Provider ID (for EPID attestation)
	SPIDHex string

	// DCAPEnabled enables DCAP attestation (recommended)
	DCAPEnabled bool

	// QuoteProviderLibrary is the path to the quote provider library
	QuoteProviderLibrary string

	// Debug enables debug mode (NOT FOR PRODUCTION)
	Debug bool

	// MaxEPCPages is the maximum EPC pages to request
	MaxEPCPages int
}

// SGXEnclaveService connects to an Intel SGX enclave
type SGXEnclaveService struct {
	config      SGXEnclaveConfig
	initialized bool
	// TODO: Add SGX-specific fields
	// - enclave handle
	// - quote provider
	// - sealed keys
}

// Ensure SGXEnclaveService implements EnclaveService
var _ EnclaveService = (*SGXEnclaveService)(nil)

// NewSGXEnclaveService creates a new SGX enclave service
// NOTE: This now delegates to the full implementation in sgx_enclave.go
// The stub type SGXEnclaveService is deprecated in favor of SGXEnclaveServiceImpl
func NewSGXEnclaveService(config SGXEnclaveConfig) (EnclaveService, error) {
	// Create and return the full SGX implementation
	return NewSGXEnclaveServiceImpl(config)
}

// GetPlatformType returns PlatformSGX
func (s *SGXEnclaveService) GetPlatformType() PlatformType {
	return PlatformSGX
}

// IsPlatformSecure returns true for SGX
func (s *SGXEnclaveService) IsPlatformSecure() bool {
	return true
}

// Initialize initializes the SGX enclave
func (s *SGXEnclaveService) Initialize(_ RuntimeConfig) error {
	return errors.New("SGX enclave not yet implemented")
}

// Score performs scoring in the SGX enclave
func (s *SGXEnclaveService) Score(_ context.Context, _ *ScoringRequest) (*ScoringResult, error) {
	return nil, errors.New("SGX enclave not yet implemented")
}

// GetMeasurement returns the enclave measurement
func (s *SGXEnclaveService) GetMeasurement() ([]byte, error) {
	return nil, errors.New("SGX enclave not yet implemented")
}

// GetEncryptionPubKey returns the encryption public key
func (s *SGXEnclaveService) GetEncryptionPubKey() ([]byte, error) {
	return nil, errors.New("SGX enclave not yet implemented")
}

// GetSigningPubKey returns the signing public key
func (s *SGXEnclaveService) GetSigningPubKey() ([]byte, error) {
	return nil, errors.New("SGX enclave not yet implemented")
}

// GenerateAttestation generates an attestation quote
func (s *SGXEnclaveService) GenerateAttestation(_ []byte) ([]byte, error) {
	return nil, errors.New("SGX enclave not yet implemented")
}

// RotateKeys rotates enclave keys
func (s *SGXEnclaveService) RotateKeys() error {
	return errors.New("SGX enclave not yet implemented")
}

// GetStatus returns the enclave status
func (s *SGXEnclaveService) GetStatus() EnclaveStatus {
	return EnclaveStatus{Initialized: false, Available: false}
}

// Shutdown shuts down the enclave
func (s *SGXEnclaveService) Shutdown() error {
	return errors.New("SGX enclave not yet implemented")
}

// =============================================================================
// SEV-SNP Enclave Service (Stub)
// =============================================================================

// SEVSNPConfig configures the SEV-SNP confidential VM service
type SEVSNPConfig struct {
	// Endpoint is the gRPC endpoint for the enclave service
	Endpoint string

	// CertChainPath is the path to the AMD certificate chain
	CertChainPath string

	// VCEKCachePath is the path to cache VCEK certificates
	VCEKCachePath string

	// MinTCBVersion is the minimum required TCB version
	MinTCBVersion string

	// AllowDebugPolicy allows debug-enabled guests (NOT FOR PRODUCTION)
	AllowDebugPolicy bool
}

// SEVSNPEnclaveService connects to an AMD SEV-SNP confidential VM
type SEVSNPEnclaveService struct {
	config      SEVSNPConfig
	initialized bool
	// TODO: Add SEV-SNP-specific fields
	// - gRPC client
	// - attestation verifier
	// - guest policy
}

// Ensure SEVSNPEnclaveService implements EnclaveService
var _ EnclaveService = (*SEVSNPEnclaveService)(nil)

// NewSEVSNPEnclaveService creates a new SEV-SNP enclave service
// NOTE: This now delegates to the full implementation in sev_enclave.go
// The stub type SEVSNPEnclaveService is deprecated in favor of SEVSNPEnclaveServiceImpl
func NewSEVSNPEnclaveService(config SEVSNPConfig) (EnclaveService, error) {
	// Create and return the full SEV-SNP implementation
	return NewSEVSNPEnclaveServiceImpl(config)
}

// GetPlatformType returns PlatformSEVSNP
func (s *SEVSNPEnclaveService) GetPlatformType() PlatformType {
	return PlatformSEVSNP
}

// IsPlatformSecure returns true for SEV-SNP
func (s *SEVSNPEnclaveService) IsPlatformSecure() bool {
	return true
}

// Initialize initializes the SEV-SNP enclave
func (s *SEVSNPEnclaveService) Initialize(_ RuntimeConfig) error {
	return errors.New("SEV-SNP enclave not yet implemented")
}

// Score performs scoring in the SEV-SNP enclave
func (s *SEVSNPEnclaveService) Score(_ context.Context, _ *ScoringRequest) (*ScoringResult, error) {
	return nil, errors.New("SEV-SNP enclave not yet implemented")
}

// GetMeasurement returns the enclave measurement
func (s *SEVSNPEnclaveService) GetMeasurement() ([]byte, error) {
	return nil, errors.New("SEV-SNP enclave not yet implemented")
}

// GetEncryptionPubKey returns the encryption public key
func (s *SEVSNPEnclaveService) GetEncryptionPubKey() ([]byte, error) {
	return nil, errors.New("SEV-SNP enclave not yet implemented")
}

// GetSigningPubKey returns the signing public key
func (s *SEVSNPEnclaveService) GetSigningPubKey() ([]byte, error) {
	return nil, errors.New("SEV-SNP enclave not yet implemented")
}

// GenerateAttestation generates an attestation quote
func (s *SEVSNPEnclaveService) GenerateAttestation(_ []byte) ([]byte, error) {
	return nil, errors.New("SEV-SNP enclave not yet implemented")
}

// RotateKeys rotates enclave keys
func (s *SEVSNPEnclaveService) RotateKeys() error {
	return errors.New("SEV-SNP enclave not yet implemented")
}

// GetStatus returns the enclave status
func (s *SEVSNPEnclaveService) GetStatus() EnclaveStatus {
	return EnclaveStatus{Initialized: false, Available: false}
}

// Shutdown shuts down the enclave
func (s *SEVSNPEnclaveService) Shutdown() error {
	return errors.New("SEV-SNP enclave not yet implemented")
}

// =============================================================================
// Nitro Enclave Service (Stub)
// =============================================================================

// NitroConfig configures the AWS Nitro Enclave service
type NitroConfig struct {
	// EnclaveImageURI is the enclave image URI
	EnclaveImageURI string

	// NSMPath is the path to the Nitro Security Module
	NSMPath string

	// CPUCount is the number of CPUs to allocate
	CPUCount int

	// MemoryMB is the memory to allocate in MB
	MemoryMB int
}

// NitroEnclaveService connects to an AWS Nitro Enclave
type NitroEnclaveService struct {
	config      NitroConfig
	initialized bool
	// TODO: Add Nitro-specific fields
	// - vsock connection
	// - NSM client
	// - enclave CID
}

// Ensure NitroEnclaveService implements EnclaveService
var _ EnclaveService = (*NitroEnclaveService)(nil)

// NewNitroEnclaveService creates a new Nitro enclave service
func NewNitroEnclaveService(config NitroConfig) (*NitroEnclaveService, error) {
	return nil, errors.New("Nitro enclave not yet implemented - see _docs/tee-integration-plan.md")
}

// GetPlatformType returns PlatformNitro
func (s *NitroEnclaveService) GetPlatformType() PlatformType {
	return PlatformNitro
}

// IsPlatformSecure returns true for Nitro
func (s *NitroEnclaveService) IsPlatformSecure() bool {
	return true
}

// Initialize initializes the Nitro enclave
func (s *NitroEnclaveService) Initialize(_ RuntimeConfig) error {
	return errors.New("Nitro enclave not yet implemented")
}

// Score performs scoring in the Nitro enclave
func (s *NitroEnclaveService) Score(_ context.Context, _ *ScoringRequest) (*ScoringResult, error) {
	return nil, errors.New("Nitro enclave not yet implemented")
}

// GetMeasurement returns the enclave measurement
func (s *NitroEnclaveService) GetMeasurement() ([]byte, error) {
	return nil, errors.New("Nitro enclave not yet implemented")
}

// GetEncryptionPubKey returns the encryption public key
func (s *NitroEnclaveService) GetEncryptionPubKey() ([]byte, error) {
	return nil, errors.New("Nitro enclave not yet implemented")
}

// GetSigningPubKey returns the signing public key
func (s *NitroEnclaveService) GetSigningPubKey() ([]byte, error) {
	return nil, errors.New("Nitro enclave not yet implemented")
}

// GenerateAttestation generates an attestation quote
func (s *NitroEnclaveService) GenerateAttestation(_ []byte) ([]byte, error) {
	return nil, errors.New("Nitro enclave not yet implemented")
}

// RotateKeys rotates enclave keys
func (s *NitroEnclaveService) RotateKeys() error {
	return errors.New("Nitro enclave not yet implemented")
}

// GetStatus returns the enclave status
func (s *NitroEnclaveService) GetStatus() EnclaveStatus {
	return EnclaveStatus{Initialized: false, Available: false}
}

// Shutdown shuts down the enclave
func (s *NitroEnclaveService) Shutdown() error {
	return errors.New("Nitro enclave not yet implemented")
}

// =============================================================================
// Factory Function
// =============================================================================

// EnclaveConfig is the configuration for creating an enclave service
type EnclaveConfig struct {
	// Platform is the TEE platform to use
	Platform PlatformType

	// RuntimeConfig is the common runtime configuration
	RuntimeConfig RuntimeConfig

	// SGXConfig is SGX-specific configuration (if Platform == "sgx")
	SGXConfig *SGXEnclaveConfig

	// SEVSNPConfig is SEV-SNP-specific configuration (if Platform == "sev-snp")
	SEVSNPConfig *SEVSNPConfig

	// NitroConfig is Nitro-specific configuration (if Platform == "nitro")
	NitroConfig *NitroConfig
}

// CreateEnclaveService creates an enclave service based on configuration
func CreateEnclaveService(config EnclaveConfig) (EnclaveService, error) {
	switch config.Platform {
	case PlatformSimulated, "":
		// Default to simulated for development
		svc := NewSimulatedEnclaveService()
		if err := svc.Initialize(config.RuntimeConfig); err != nil {
			return nil, err
		}
		return svc, nil

	case PlatformSGX:
		if config.SGXConfig == nil {
			return nil, errors.New("SGX configuration required")
		}
		return NewSGXEnclaveService(*config.SGXConfig)

	case PlatformSEVSNP:
		if config.SEVSNPConfig == nil {
			return nil, errors.New("SEV-SNP configuration required")
		}
		return NewSEVSNPEnclaveService(*config.SEVSNPConfig)

	case PlatformNitro:
		if config.NitroConfig == nil {
			return nil, errors.New("Nitro configuration required")
		}
		return NewNitroEnclaveService(*config.NitroConfig)

	default:
		return nil, errors.New("unknown platform: " + string(config.Platform))
	}
}

// =============================================================================
// Attestation Verification
// =============================================================================

// AttestationVerifier verifies TEE attestation reports
type AttestationVerifier interface {
	// VerifyReport verifies an attestation report
	VerifyReport(ctx context.Context, report *AttestationReport) error

	// IsMeasurementAllowed checks if a measurement is in the allowlist
	IsMeasurementAllowed(measurement []byte) bool

	// AddAllowedMeasurement adds a measurement to the allowlist
	AddAllowedMeasurement(measurement []byte)
}

// SimpleAttestationVerifier is a basic verifier with measurement allowlist
type SimpleAttestationVerifier struct {
	allowedMeasurements map[string]bool
}

// NewSimpleAttestationVerifier creates a new simple attestation verifier
func NewSimpleAttestationVerifier() *SimpleAttestationVerifier {
	return &SimpleAttestationVerifier{
		allowedMeasurements: make(map[string]bool),
	}
}

// VerifyReport verifies an attestation report
func (v *SimpleAttestationVerifier) VerifyReport(ctx context.Context, report *AttestationReport) error {
	if err := report.Validate(); err != nil {
		return err
	}

	// Check measurement allowlist
	if !v.IsMeasurementAllowed(report.Measurement) {
		return errors.New("measurement not in allowlist")
	}

	// Platform-specific verification would go here
	// TODO: Implement SGX DCAP verification
	// TODO: Implement SEV-SNP verification
	// TODO: Implement Nitro verification

	return nil
}

// IsMeasurementAllowed checks if a measurement is in the allowlist
func (v *SimpleAttestationVerifier) IsMeasurementAllowed(measurement []byte) bool {
	key := string(measurement)
	return v.allowedMeasurements[key]
}

// AddAllowedMeasurement adds a measurement to the allowlist
func (v *SimpleAttestationVerifier) AddAllowedMeasurement(measurement []byte) {
	key := string(measurement)
	v.allowedMeasurements[key] = true
}
