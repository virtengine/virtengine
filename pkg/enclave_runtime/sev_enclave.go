// Package enclave_runtime provides TEE enclave implementations.
//
// This file implements the AMD SEV-SNP enclave service interface for VirtEngine VEID.
// The implementation provides POC stubs for SEV-SNP operations including:
// - SNP attestation report generation
// - Memory encryption verification
// - Launch measurement verification
// - vTPM-based key derivation
//
// Task Reference: VE-2023 - TEE Integration Planning and POC
//
// IMPORTANT: This is a POC implementation. Real SEV-SNP hardware calls are stubbed
// and marked with TODO comments. Full implementation requires:
// - AMD EPYC processor with SEV-SNP support (Milan or later)
// - Linux kernel 6.0+ with SNP patches
// - Access to /dev/sev-guest device
// - AMD KDS (Key Distribution Server) for VCEK certificates
package enclave_runtime

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/hkdf"
)

// =============================================================================
// SEV-SNP Constants and Types
// =============================================================================

const (
	// SNP attestation report version
	SNPReportVersion = 2

	// Measurement sizes
	SNPLaunchDigestSize = 48
	SNPReportDataSize   = 64
	SNPChipIDSize       = 64

	// SNP signature size (ECDSA P-384)
	SNPSignatureSize = 512

	// Guest policy flags
	SNPPolicyNoDebug        = 0x0001
	SNPPolicyNoKeyShare     = 0x0002
	SNPPolicySingleSocket   = 0x0010
	SNPPolicySMTAllowed     = 0x0020
	SNPPolicyMigrationAgent = 0x0040

	// Platform info flags
	SNPPlatformSMTEnabled  = 0x0001
	SNPPlatformTSMEEnabled = 0x0002
)

// SNPGuestPolicy represents the SNP guest policy
type SNPGuestPolicy struct {
	ABIMinor       uint8
	ABIMajor       uint8
	SMT            bool // Allow SMT (simultaneous multi-threading)
	MigrationAgent bool // Allow migration
	Debug          bool // Allow debug (MUST be false for production)
	SingleSocket   bool // Restrict to single socket
}

// ToUint64 serializes the policy to a uint64
func (p SNPGuestPolicy) ToUint64() uint64 {
	var policy uint64
	policy |= uint64(p.ABIMinor)
	policy |= uint64(p.ABIMajor) << 8
	if !p.Debug {
		policy |= SNPPolicyNoDebug
	}
	if p.SingleSocket {
		policy |= SNPPolicySingleSocket
	}
	if p.SMT {
		policy |= SNPPolicySMTAllowed
	}
	if p.MigrationAgent {
		policy |= SNPPolicyMigrationAgent
	}
	return policy
}

// SNPTCBVersion represents the TCB (Trusted Computing Base) version
type SNPTCBVersion struct {
	BootLoader uint8
	TEE        uint8
	Reserved   [4]uint8
	SNP        uint8
	Microcode  uint8
}

// ToUint64 serializes the TCB version
func (t SNPTCBVersion) ToUint64() uint64 {
	return uint64(t.BootLoader) |
		uint64(t.TEE)<<8 |
		uint64(t.SNP)<<48 |
		uint64(t.Microcode)<<56
}

// SNPLaunchDigest represents the launch measurement (48 bytes, SHA-384)
type SNPLaunchDigest [48]byte

// String returns hex representation
func (d SNPLaunchDigest) String() string {
	return fmt.Sprintf("%x", d[:])
}

// SNPChipID represents the unique chip identifier
type SNPChipID [64]byte

// SNPAttestationReport represents an SNP attestation report
type SNPAttestationReport struct {
	// Header fields
	Version  uint32
	GuestSVN uint32
	Policy   SNPGuestPolicy
	FamilyID [16]byte
	ImageID  [16]byte

	// TCB information
	CurrentTCB       SNPTCBVersion
	PlatformInfo     uint64
	AuthorKeyEnabled uint32
	Reserved1        uint32

	// Measurements
	LaunchDigest    SNPLaunchDigest
	ReportData      [64]byte
	HostData        [32]byte
	IDKeyDigest     [48]byte
	AuthorKeyDigest [48]byte
	ReportID        [32]byte
	ReportIDMA      [32]byte
	ReportedTCB     SNPTCBVersion
	Reserved2       [24]byte
	ChipID          SNPChipID

	// Signature
	Signature [512]byte
}

// Validate performs basic validation of the SNP report
func (r *SNPAttestationReport) Validate() error {
	if r.Version < SNPReportVersion {
		return fmt.Errorf("unsupported SNP report version: %d", r.Version)
	}
	if r.Policy.Debug {
		return errors.New("debug mode enabled - not secure for production")
	}
	return nil
}

// =============================================================================
// SEV-SNP Enclave Service Implementation
// =============================================================================

// SEVSNPEnclaveServiceImpl implements the EnclaveService interface for AMD SEV-SNP
type SEVSNPEnclaveServiceImpl struct {
	mu sync.RWMutex

	// Configuration
	config        SEVSNPConfig
	runtimeConfig RuntimeConfig
	hardwareMode  HardwareMode

	// Hardware backend (nil if using simulation)
	hardwareBackend *SEVHardwareBackend

	// State
	initialized  bool
	startTime    time.Time
	activeReqs   int
	totalProc    uint64
	currentEpoch uint64
	lastError    string

	// Simulated CVM state (in real impl, these are inside the confidential VM)
	launchDigest   SNPLaunchDigest
	chipID         SNPChipID
	guestPolicy    SNPGuestPolicy
	currentTCB     SNPTCBVersion
	vcekPrivateKey []byte // Simulated VCEK (Versioned Chip Endorsement Key)
	encryptionKey  []byte
	signingKey     []byte
	encryptPubKey  []byte
	signingPubKey  []byte
}

// Compile-time interface check
var _ EnclaveService = (*SEVSNPEnclaveServiceImpl)(nil)

// NewSEVSNPEnclaveServiceImpl creates a new SEV-SNP enclave service implementation
//
// This is a POC implementation that simulates SEV-SNP operations.
// For production use, this must be replaced with actual /dev/sev-guest ioctls.
func NewSEVSNPEnclaveServiceImpl(config SEVSNPConfig) (*SEVSNPEnclaveServiceImpl, error) {
	return NewSEVSNPEnclaveServiceImplWithMode(config, HardwareModeAuto)
}

// NewSEVSNPEnclaveServiceImplWithMode creates a new SEV-SNP enclave service with explicit hardware mode
func NewSEVSNPEnclaveServiceImplWithMode(config SEVSNPConfig, mode HardwareMode) (*SEVSNPEnclaveServiceImpl, error) {
	if config.Endpoint == "" {
		return nil, errors.New("endpoint required for SEV-SNP service")
	}

	// Validate configuration
	if config.AllowDebugPolicy {
		fmt.Println("WARNING: SEV-SNP debug policy allowed - NOT SECURE FOR PRODUCTION")
	}

	svc := &SEVSNPEnclaveServiceImpl{
		config:       config,
		currentEpoch: 1,
		hardwareMode: mode,
	}

	// Initialize hardware backend based on mode
	if mode != HardwareModeSimulate {
		backend := NewSEVHardwareBackend()
		if backend.IsAvailable() {
			if err := backend.Initialize(); err == nil {
				svc.hardwareBackend = backend
				fmt.Println("INFO: SEV-SNP hardware backend initialized successfully")
			} else if mode == HardwareModeRequire {
				return nil, fmt.Errorf("SEV-SNP hardware required but initialization failed: %w", err)
			} else {
				fmt.Printf("INFO: SEV-SNP hardware initialization failed, using simulation: %v\n", err)
			}
		} else if mode == HardwareModeRequire {
			return nil, fmt.Errorf("%w: SEV-SNP hardware required but not available", ErrHardwareNotAvailable)
		} else {
			fmt.Println("INFO: SEV-SNP hardware not available, using simulation mode")
		}
	} else {
		fmt.Println("INFO: SEV-SNP running in forced simulation mode")
	}

	return svc, nil
}

// =============================================================================
// EnclaveService Interface Implementation
// =============================================================================

// Initialize initializes the SEV-SNP confidential VM service
func (s *SEVSNPEnclaveServiceImpl) Initialize(config RuntimeConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.initialized {
		return errors.New("enclave already initialized")
	}

	s.runtimeConfig = config
	s.startTime = time.Now()

	// TODO: Real SEV-SNP implementation would:
	// 1. Verify we're running in an SNP guest (check /dev/sev-guest)
	// 2. Fetch launch measurement from platform
	// 3. Initialize gRPC server for enclave communication
	// 4. Set up connection to AMD KDS for VCEK certificates

	// Simulate CVM initialization
	if err := s.simulateCVMInitialization(); err != nil {
		return fmt.Errorf("failed to initialize CVM: %w", err)
	}

	// Derive enclave keys
	if err := s.deriveEnclaveKeys(); err != nil {
		return fmt.Errorf("failed to derive keys: %w", err)
	}

	s.initialized = true
	return nil
}

// Score performs identity scoring inside the SEV-SNP CVM
func (s *SEVSNPEnclaveServiceImpl) Score(ctx context.Context, request *ScoringRequest) (*ScoringResult, error) {
	s.mu.Lock()
	if !s.initialized {
		s.mu.Unlock()
		return nil, ErrEnclaveNotInitialized
	}
	if s.activeReqs >= s.runtimeConfig.MaxConcurrentRequests {
		s.mu.Unlock()
		return nil, ErrEnclaveUnavailable
	}
	s.activeReqs++
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.activeReqs--
		s.totalProc++
		s.mu.Unlock()
	}()

	// Validate request
	if err := request.Validate(s.runtimeConfig); err != nil {
		return &ScoringResult{RequestID: request.RequestID, Error: err.Error()}, nil
	}

	startTime := time.Now()

	// TODO: Real SEV-SNP implementation would:
	// 1. Receive encrypted payload via gRPC/vsock
	// 2. Decrypt with keys stored in encrypted memory
	// 3. Run ML scoring (TensorFlow Lite) in protected VM
	// 4. Sign result with enclave key
	// 5. Generate attestation report binding result

	// Simulate CVM scoring (POC)
	result := s.simulateCVMScoring(request)
	result.ProcessingTimeMs = time.Since(startTime).Milliseconds()

	return result, nil
}

// GetMeasurement returns the launch measurement (launch digest)
func (s *SEVSNPEnclaveServiceImpl) GetMeasurement() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return nil, ErrEnclaveNotInitialized
	}

	return s.launchDigest[:], nil
}

// GetEncryptionPubKey returns the CVM's encryption public key
func (s *SEVSNPEnclaveServiceImpl) GetEncryptionPubKey() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return nil, ErrEnclaveNotInitialized
	}

	return s.encryptPubKey, nil
}

// GetSigningPubKey returns the CVM's signing public key
func (s *SEVSNPEnclaveServiceImpl) GetSigningPubKey() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return nil, ErrEnclaveNotInitialized
	}

	return s.signingPubKey, nil
}

// GenerateAttestation generates an SNP attestation report
func (s *SEVSNPEnclaveServiceImpl) GenerateAttestation(reportData []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return nil, ErrEnclaveNotInitialized
	}

	if len(reportData) > SNPReportDataSize {
		return nil, errors.New("report data too large")
	}

	// Use hardware backend if available
	if s.hardwareBackend != nil {
		attestation, err := s.hardwareBackend.GetAttestation(reportData)
		if err != nil {
			s.lastError = fmt.Sprintf("hardware attestation failed: %v", err)
			// Fall back to simulation
		} else {
			return attestation, nil
		}
	}

	// TODO: Real SEV-SNP implementation would:
	// 1. Open /dev/sev-guest
	// 2. Send SNP_GUEST_REQUEST ioctl with MSG_REPORT_REQ
	// 3. Receive attestation report signed by VCEK

	report, err := s.simulateSNPReportGeneration(reportData)
	if err != nil {
		return nil, fmt.Errorf("report generation failed: %w", err)
	}

	return report, nil
}

// RotateKeys initiates key rotation
func (s *SEVSNPEnclaveServiceImpl) RotateKeys() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.initialized {
		return ErrEnclaveNotInitialized
	}

	// Increment epoch
	s.currentEpoch++

	// Re-derive keys with new epoch
	if err := s.deriveEnclaveKeys(); err != nil {
		s.lastError = fmt.Sprintf("key rotation failed: %v", err)
		return fmt.Errorf("key rotation failed: %w", err)
	}

	return nil
}

// GetStatus returns the enclave status
func (s *SEVSNPEnclaveServiceImpl) GetStatus() EnclaveStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var uptime int64
	if s.initialized {
		uptime = int64(time.Since(s.startTime).Seconds())
	}

	return EnclaveStatus{
		Initialized:    s.initialized,
		Available:      s.initialized && s.activeReqs < s.runtimeConfig.MaxConcurrentRequests,
		CurrentEpoch:   s.currentEpoch,
		ActiveRequests: s.activeReqs,
		TotalProcessed: s.totalProc,
		LastError:      s.lastError,
		Uptime:         uptime,
	}
}

// Shutdown gracefully shuts down the CVM service
func (s *SEVSNPEnclaveServiceImpl) Shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.initialized {
		return nil
	}

	// TODO: Real SEV-SNP implementation would:
	// 1. Seal critical state to encrypted storage
	// 2. Close /dev/sev-guest
	// 3. Clean up gRPC connections

	// Securely clear keys from memory
	s.scrubKeys()

	s.initialized = false
	return nil
}

// =============================================================================
// SEV-SNP-Specific Methods
// =============================================================================

// GetPlatformType returns PlatformSEVSNP
func (s *SEVSNPEnclaveServiceImpl) GetPlatformType() PlatformType {
	return PlatformSEVSNP
}

// IsPlatformSecure returns true (SEV-SNP is secure when debug is disabled)
func (s *SEVSNPEnclaveServiceImpl) IsPlatformSecure() bool {
	return !s.guestPolicy.Debug
}

// IsHardwareEnabled returns true if real SEV-SNP hardware is being used
func (s *SEVSNPEnclaveServiceImpl) IsHardwareEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hardwareBackend != nil
}

// GetHardwareMode returns the configured hardware mode
func (s *SEVSNPEnclaveServiceImpl) GetHardwareMode() HardwareMode {
	return s.hardwareMode
}

// DeriveKey derives a key from the SEV-SNP root of trust
func (s *SEVSNPEnclaveServiceImpl) DeriveKey(context []byte, keySize int) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return nil, ErrEnclaveNotInitialized
	}

	// Use hardware backend if available
	if s.hardwareBackend != nil {
		key, err := s.hardwareBackend.DeriveKey(context, keySize)
		if err != nil {
			s.lastError = fmt.Sprintf("hardware key derivation failed: %v", err)
			// Fall back to simulation
		} else {
			return key, nil
		}
	}

	// Simulated key derivation
	return s.hkdfDerive(s.vcekPrivateKey, s.launchDigest[:], context, keySize), nil
}

// GetChipID returns the unique chip identifier
func (s *SEVSNPEnclaveServiceImpl) GetChipID() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return nil, ErrEnclaveNotInitialized
	}

	return s.chipID[:], nil
}

// GetTCBVersion returns the current TCB version
func (s *SEVSNPEnclaveServiceImpl) GetTCBVersion() (*SNPTCBVersion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return nil, ErrEnclaveNotInitialized
	}

	tcb := s.currentTCB
	return &tcb, nil
}

// GetGuestPolicy returns the guest policy
func (s *SEVSNPEnclaveServiceImpl) GetGuestPolicy() (*SNPGuestPolicy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return nil, ErrEnclaveNotInitialized
	}

	policy := s.guestPolicy
	return &policy, nil
}

// VerifyMemoryEncryption verifies that memory encryption is active
func (s *SEVSNPEnclaveServiceImpl) VerifyMemoryEncryption() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return ErrEnclaveNotInitialized
	}

	// TODO: Real implementation would:
	// 1. Read /sys/kernel/debug/x86/sev to verify SEV is active
	// 2. Check cpuid for SEV-SNP feature flags
	// 3. Verify guest policy has encryption enabled

	// POC: Always report encrypted in simulation
	return nil
}

// VerifyLaunchMeasurement verifies the launch measurement matches expected
func (s *SEVSNPEnclaveServiceImpl) VerifyLaunchMeasurement(expected []byte) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return ErrEnclaveNotInitialized
	}

	if len(expected) != SNPLaunchDigestSize {
		return fmt.Errorf("invalid expected measurement size: got %d, want %d", len(expected), SNPLaunchDigestSize)
	}

	// Compare measurements
	for i := 0; i < SNPLaunchDigestSize; i++ {
		if s.launchDigest[i] != expected[i] {
			return errors.New("launch measurement mismatch")
		}
	}

	return nil
}

// FetchVCEKCertificate fetches the VCEK certificate from AMD KDS
func (s *SEVSNPEnclaveServiceImpl) FetchVCEKCertificate() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return nil, ErrEnclaveNotInitialized
	}

	// TODO: Real implementation would:
	// 1. Extract chip_id from attestation report
	// 2. Query AMD KDS: https://kdsintf.amd.com/vcek/v1/{product_name}/{chip_id}?blSPL=..&teeSPL=..&snpSPL=..&ucodeSPL=..
	// 3. Cache certificate locally

	// POC: Return simulated certificate
	cert := sha512.Sum512(s.chipID[:])
	return cert[:], nil
}

// GenerateExtendedReport generates an extended attestation report with certificate chain
func (s *SEVSNPEnclaveServiceImpl) GenerateExtendedReport(reportData []byte) ([]byte, [][]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return nil, nil, ErrEnclaveNotInitialized
	}

	// Generate base report
	report, err := s.simulateSNPReportGeneration(reportData)
	if err != nil {
		return nil, nil, err
	}

	// TODO: Real implementation would:
	// 1. Use SNP_EXTENDED_REPORT ioctl
	// 2. Include VCEK certificate chain
	// 3. Include TCB certificates

	// POC: Return simulated certificate chain
	vcek, _ := s.FetchVCEKCertificate()
	ask := sha256Bytes([]byte("AMD_SEV_SIGNING_KEY")) // AMD SEV Signing Key
	ark := sha256Bytes([]byte("AMD_ROOT_KEY"))        // AMD Root Key

	certChain := [][]byte{vcek, ask, ark}

	return report, certChain, nil
}

// =============================================================================
// Simulation Methods (POC Only - Replace with Real SEV-SNP Calls)
// =============================================================================

// simulateCVMInitialization simulates SEV-SNP CVM initialization
func (s *SEVSNPEnclaveServiceImpl) simulateCVMInitialization() error {
	// Simulate launch measurement (SHA-384 of guest initial state)
	// In real SEV-SNP, this is computed by the PSP during launch
	guestImage := []byte("veid_scoring_cvm_v1_" + s.config.Endpoint)
	digest := sha512.Sum384(guestImage)
	copy(s.launchDigest[:], digest[:])

	// Simulate chip ID (unique per-CPU)
	chipIDSeed := []byte("amd_epyc_milan_" + time.Now().Format(time.RFC3339))
	chipHash := sha512.Sum512(chipIDSeed)
	copy(s.chipID[:], chipHash[:])

	// Set guest policy (production settings)
	s.guestPolicy = SNPGuestPolicy{
		ABIMinor:       0,
		ABIMajor:       1,
		SMT:            true,
		Debug:          s.config.AllowDebugPolicy, // Should be false in production
		SingleSocket:   false,
		MigrationAgent: false,
	}

	// Set TCB version
	s.currentTCB = SNPTCBVersion{
		BootLoader: 2,
		TEE:        0,
		SNP:        8,
		Microcode:  115,
	}

	// Simulate VCEK derivation
	// In real SEV-SNP, VCEK is a per-chip key used to sign attestation reports
	vcekSeed := append(s.chipID[:], []byte("vcek_derive")...)
	vcekHash := sha512.Sum512(vcekSeed)
	s.vcekPrivateKey = vcekHash[:32]

	return nil
}

// deriveEnclaveKeys derives encryption and signing keys
func (s *SEVSNPEnclaveServiceImpl) deriveEnclaveKeys() error {
	// Derive keys using HKDF from launch measurement and epoch
	// In real SEV-SNP, keys would be derived from vTPM or launch secret
	salt := append(s.launchDigest[:], s.chipID[:32]...)

	// Encryption key (X25519 seed)
	encInfo := fmt.Sprintf("encryption_key_epoch_%d", s.currentEpoch)
	s.encryptionKey = s.hkdfDerive(s.vcekPrivateKey, salt, []byte(encInfo), 32)

	// Signing key (Ed25519 seed)
	sigInfo := fmt.Sprintf("signing_key_epoch_%d", s.currentEpoch)
	s.signingKey = s.hkdfDerive(s.vcekPrivateKey, salt, []byte(sigInfo), 32)

	// Derive public keys (simplified - in real impl use proper EC ops)
	s.encryptPubKey = sha256Bytes(append([]byte("SNP_X25519_pub_"), s.encryptionKey...))
	s.signingPubKey = sha256Bytes(append([]byte("SNP_Ed25519_pub_"), s.signingKey...))

	return nil
}

// hkdfDerive performs HKDF key derivation
func (s *SEVSNPEnclaveServiceImpl) hkdfDerive(secret, salt, info []byte, length int) []byte {
	reader := hkdf.New(sha256.New, secret, salt, info)
	key := make([]byte, length)
	_, _ = reader.Read(key)
	return key
}

// simulateCVMScoring simulates scoring inside the SEV-SNP CVM
func (s *SEVSNPEnclaveServiceImpl) simulateCVMScoring(request *ScoringRequest) *ScoringResult {
	// Compute input hash
	inputHash := sha256.Sum256(request.Ciphertext)

	// Simulate deterministic score based on input
	score := uint32(inputHash[0]) % 101

	// Determine status
	var status string
	switch {
	case score >= 80:
		status = "verified"
	case score >= 50:
		status = "needs_review"
	default:
		status = "rejected"
	}

	// Generate evidence hashes
	evidenceHashes := [][]byte{
		sha256Bytes([]byte("snp_face_embedding")),
		sha256Bytes([]byte("snp_document_features")),
	}

	// Generate model version hash
	modelVersionHash := sha256Bytes([]byte("snp_veid_model_v1.0.0"))

	// Sign result inside CVM
	signingPayload := s.computeSigningPayload(request.RequestID, score, status, inputHash[:])
	enclaveSignature := s.signInsideCVM(signingPayload)

	return &ScoringResult{
		RequestID:            request.RequestID,
		Score:                score,
		Status:               status,
		ReasonCodes:          []string{"snp_score_" + status},
		EvidenceHashes:       evidenceHashes,
		ModelVersionHash:     modelVersionHash,
		InputHash:            inputHash[:],
		EnclaveSignature:     enclaveSignature,
		MeasurementHash:      s.launchDigest[:],
		AttestationReference: sha256Bytes([]byte("snp_attestation_ref")),
	}
}

// computeSigningPayload computes the payload to sign
func (s *SEVSNPEnclaveServiceImpl) computeSigningPayload(requestID string, score uint32, status string, inputHash []byte) []byte {
	h := sha256.New()
	h.Write([]byte(requestID))
	h.Write([]byte{byte(score >> 24), byte(score >> 16), byte(score >> 8), byte(score)})
	h.Write([]byte(status))
	h.Write(inputHash)
	h.Write(s.launchDigest[:])
	return h.Sum(nil)
}

// signInsideCVM simulates signing inside the CVM
func (s *SEVSNPEnclaveServiceImpl) signInsideCVM(payload []byte) []byte {
	// In real SEV-SNP, this uses Ed25519 with keys stored in encrypted memory
	h := sha256.New()
	h.Write(s.signingKey)
	h.Write(payload)
	return h.Sum(nil)
}

// simulateSNPReportGeneration simulates SNP attestation report generation
func (s *SEVSNPEnclaveServiceImpl) simulateSNPReportGeneration(reportData []byte) ([]byte, error) {
	// Build attestation report
	report := &SNPAttestationReport{
		Version:      SNPReportVersion,
		GuestSVN:     1,
		Policy:       s.guestPolicy,
		CurrentTCB:   s.currentTCB,
		ReportedTCB:  s.currentTCB,
		PlatformInfo: SNPPlatformSMTEnabled,
		LaunchDigest: s.launchDigest,
		ChipID:       s.chipID,
	}

	// Copy report data
	copy(report.ReportData[:], reportData)

	// Generate family ID and image ID
	familyID := sha256.Sum256([]byte("virtengine_veid_family"))
	copy(report.FamilyID[:], familyID[:16])
	imageID := sha256.Sum256([]byte("veid_scorer_image_v1"))
	copy(report.ImageID[:], imageID[:16])

	// Generate report ID
	reportIDSeed := append(reportData, []byte(fmt.Sprintf("report_%d", time.Now().UnixNano()))...)
	reportID := sha256.Sum256(reportIDSeed)
	copy(report.ReportID[:], reportID[:])

	// Serialize report
	serialized := s.serializeSNPReport(report)

	// Sign with VCEK (simulated)
	signature := s.signWithVCEK(serialized)
	copy(report.Signature[:], signature)

	// Re-serialize with signature
	return s.serializeSNPReport(report), nil
}

// serializeSNPReport serializes an SNP attestation report
func (s *SEVSNPEnclaveServiceImpl) serializeSNPReport(report *SNPAttestationReport) []byte {
	buf := make([]byte, 0, 1184) // Standard SNP report size

	// Header
	buf = binary.LittleEndian.AppendUint32(buf, report.Version)
	buf = binary.LittleEndian.AppendUint32(buf, report.GuestSVN)
	buf = binary.LittleEndian.AppendUint64(buf, report.Policy.ToUint64())
	buf = append(buf, report.FamilyID[:]...)
	buf = append(buf, report.ImageID[:]...)

	// TCB
	buf = binary.LittleEndian.AppendUint64(buf, report.CurrentTCB.ToUint64())
	buf = binary.LittleEndian.AppendUint64(buf, report.PlatformInfo)
	buf = binary.LittleEndian.AppendUint32(buf, report.AuthorKeyEnabled)
	buf = binary.LittleEndian.AppendUint32(buf, report.Reserved1)

	// Measurements
	buf = append(buf, report.LaunchDigest[:]...)
	buf = append(buf, report.ReportData[:]...)
	buf = append(buf, report.HostData[:]...)
	buf = append(buf, report.IDKeyDigest[:]...)
	buf = append(buf, report.AuthorKeyDigest[:]...)
	buf = append(buf, report.ReportID[:]...)
	buf = append(buf, report.ReportIDMA[:]...)
	buf = binary.LittleEndian.AppendUint64(buf, report.ReportedTCB.ToUint64())
	buf = append(buf, report.Reserved2[:]...)
	buf = append(buf, report.ChipID[:]...)

	// Signature
	buf = append(buf, report.Signature[:]...)

	return buf
}

// signWithVCEK signs data with the VCEK (ECDSA P-384 simulated)
func (s *SEVSNPEnclaveServiceImpl) signWithVCEK(data []byte) []byte {
	// In real SEV-SNP, the PSP signs the report with the VCEK
	// using ECDSA with P-384 curve
	h := sha512.New384()
	h.Write(s.vcekPrivateKey)
	h.Write(data)
	sig := h.Sum(nil)

	// Pad to SNP signature size (512 bytes for P-384 signature + metadata)
	result := make([]byte, SNPSignatureSize)
	copy(result, sig)
	return result
}

// VerifyReport verifies an SNP attestation report
func (s *SEVSNPEnclaveServiceImpl) VerifyReport(report []byte) error {
	if len(report) < 100 {
		return errors.New("report too short")
	}

	// TODO: Real implementation would:
	// 1. Fetch VCEK certificate from AMD KDS
	// 2. Verify certificate chain (VCEK -> ASK -> ARK)
	// 3. Verify report signature with VCEK public key
	// 4. Check TCB version against minimum requirements
	// 5. Validate guest policy (debug=false, etc.)
	// 6. Verify launch measurement against allowlist

	// POC: Basic format validation
	version := binary.LittleEndian.Uint32(report[:4])
	if version < SNPReportVersion {
		return fmt.Errorf("unsupported report version: %d", version)
	}

	return nil
}

// scrubKeys securely clears keys from memory
func (s *SEVSNPEnclaveServiceImpl) scrubKeys() {
	// Use ScrubBytes from memory_scrub.go
	if s.vcekPrivateKey != nil {
		ScrubBytes(s.vcekPrivateKey)
	}
	if s.encryptionKey != nil {
		ScrubBytes(s.encryptionKey)
	}
	if s.signingKey != nil {
		ScrubBytes(s.signingKey)
	}
}

// =============================================================================
// Extended Report Data Helpers
// =============================================================================

// BindResultToReport creates report data that binds a scoring result
func BindResultToReport(result *ScoringResult, nonce []byte) []byte {
	h := sha256.New()
	h.Write([]byte(result.RequestID))
	h.Write(result.InputHash)
	h.Write(result.EnclaveSignature)
	h.Write(nonce)
	digest := h.Sum(nil)

	// Report data is 64 bytes
	reportData := make([]byte, SNPReportDataSize)
	copy(reportData, digest)
	copy(reportData[32:], nonce[:min(32, len(nonce))])

	return reportData
}

// ExtractNonceFromReport extracts the nonce from report data
func ExtractNonceFromReport(reportData []byte) []byte {
	if len(reportData) < SNPReportDataSize {
		return nil
	}
	return reportData[32:64]
}
