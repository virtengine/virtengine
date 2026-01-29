// Package enclave_runtime provides TEE enclave implementations.
//
// This file implements the Intel SGX enclave service interface for VirtEngine VEID.
// The implementation provides POC stubs for SGX operations including:
// - DCAP remote attestation
// - Key derivation using SGX sealing
// - Measurement verification
//
// Task Reference: VE-2023 - TEE Integration Planning and POC
//
// IMPORTANT: This is a POC implementation. Real SGX hardware calls are stubbed
// and marked with TODO comments. Full implementation requires:
// - Intel SGX SDK or Gramine LibOS
// - DCAP Quote Provider Library
// - Access to Intel PCS/PCCS for collateral
package enclave_runtime

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/hkdf"
)

// =============================================================================
// SGX Constants and Types
// =============================================================================

const (
	// SGX quote version for DCAP
	SGXQuoteVersionDCAP = 3

	// Measurement sizes
	SGXMREnclaveSize = 32
	SGXMRSignerSize  = 32

	// Key sizes
	SGXSealKeySize     = 16
	SGXReportDataSize  = 64
	SGXTargetInfoSize  = 512
	SGXReportSize      = 432
	SGXQuoteHeaderSize = 48

	// Key policy flags
	SGXKeyPolicyMREnclave = 0x0001
	SGXKeyPolicyMRSigner  = 0x0002
	SGXKeyPolicyNoISVSVN  = 0x0004

	// Attribute flags
	SGXFlagInitted    = 0x0001
	SGXFlagDebug      = 0x0002
	SGXFlagMode64Bit  = 0x0004
	SGXFlagProvisionKey = 0x0010
)

// SGXMeasurement represents an enclave measurement (MRENCLAVE or MRSIGNER)
type SGXMeasurement [32]byte

// String returns hex representation
func (m SGXMeasurement) String() string {
	return fmt.Sprintf("%x", m[:])
}

// SGXAttributes represents enclave attributes
type SGXAttributes struct {
	Flags uint64
	Xfrm  uint64
}

// IsDebug returns true if debug mode is enabled
func (a SGXAttributes) IsDebug() bool {
	return (a.Flags & SGXFlagDebug) != 0
}

// SGXReportBody represents the body of an SGX report
type SGXReportBody struct {
	CPUSVN      [16]byte
	MiscSelect  uint32
	Reserved1   [12]byte
	ISVExtProdID [16]byte
	Attributes  SGXAttributes
	MREnclave   SGXMeasurement
	Reserved2   [32]byte
	MRSigner    SGXMeasurement
	Reserved3   [32]byte
	ConfigID    [64]byte
	ISVProdID   uint16
	ISVSVN      uint16
	ConfigSVN   uint16
	Reserved4   [42]byte
	ISVFamilyID [16]byte
	ReportData  [64]byte
}

// SGXQuoteHeader represents the header of a DCAP quote
type SGXQuoteHeader struct {
	Version      uint16
	AttKeyType   uint16
	TEEType      uint32
	Reserved     uint32
	QEVendorID   [16]byte
	UserData     [20]byte
}

// SGXQuote represents a full DCAP attestation quote
type SGXQuote struct {
	Header          SGXQuoteHeader
	ReportBody      SGXReportBody
	SignatureLength uint32
	Signature       []byte
}

// =============================================================================
// SGX Enclave Service Implementation
// =============================================================================

// SGXEnclaveServiceImpl implements the EnclaveService interface for Intel SGX
type SGXEnclaveServiceImpl struct {
	mu sync.RWMutex

	// Configuration
	config        SGXEnclaveConfig
	runtimeConfig RuntimeConfig

	// State
	initialized   bool
	startTime     time.Time
	activeReqs    int
	totalProc     uint64
	currentEpoch  uint64
	lastError     string

	// Simulated enclave state (in real impl, these are inside SGX)
	mrEnclave      SGXMeasurement
	mrSigner       SGXMeasurement
	sealKey        []byte
	encryptionKey  []byte
	signingKey     []byte
	encryptPubKey  []byte
	signingPubKey  []byte
}

// Compile-time interface check
var _ EnclaveService = (*SGXEnclaveServiceImpl)(nil)

// NewSGXEnclaveServiceImpl creates a new SGX enclave service implementation
//
// This is a POC implementation that simulates SGX operations.
// For production use, this must be replaced with actual SGX SDK calls.
func NewSGXEnclaveServiceImpl(config SGXEnclaveConfig) (*SGXEnclaveServiceImpl, error) {
	if config.EnclavePath == "" {
		return nil, errors.New("enclave path required")
	}

	// Validate configuration
	if config.Debug {
		// Allow debug mode for development, but log warning
		fmt.Println("WARNING: SGX debug mode enabled - NOT SECURE FOR PRODUCTION")
	}

	return &SGXEnclaveServiceImpl{
		config:       config,
		currentEpoch: 1,
	}, nil
}

// =============================================================================
// EnclaveService Interface Implementation
// =============================================================================

// Initialize initializes the SGX enclave
func (s *SGXEnclaveServiceImpl) Initialize(config RuntimeConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.initialized {
		return errors.New("enclave already initialized")
	}

	s.runtimeConfig = config
	s.startTime = time.Now()

	// TODO: Real SGX implementation would:
	// 1. Load the signed enclave binary (sgx_create_enclave)
	// 2. Initialize enclave state via ECALL
	// 3. Verify SIGSTRUCT and enclave measurement
	// 4. Set up the Quote Provider for DCAP

	// Simulate enclave creation and measurement
	if err := s.simulateEnclaveCreation(); err != nil {
		return fmt.Errorf("failed to create enclave: %w", err)
	}

	// Derive enclave keys
	if err := s.deriveEnclaveKeys(); err != nil {
		return fmt.Errorf("failed to derive keys: %w", err)
	}

	s.initialized = true
	return nil
}

// Score performs identity scoring inside the SGX enclave
func (s *SGXEnclaveServiceImpl) Score(ctx context.Context, request *ScoringRequest) (*ScoringResult, error) {
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

	// TODO: Real SGX implementation would:
	// 1. ECALL into enclave with encrypted payload
	// 2. Inside enclave: decrypt with sealed key
	// 3. Inside enclave: run ML scoring
	// 4. Inside enclave: sign result with enclave key
	// 5. ORET with signed result

	// Simulate enclave scoring (POC)
	result := s.simulateEnclaveScoring(request)
	result.ProcessingTimeMs = time.Since(startTime).Milliseconds()

	return result, nil
}

// GetMeasurement returns the enclave measurement (MRENCLAVE)
func (s *SGXEnclaveServiceImpl) GetMeasurement() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return nil, ErrEnclaveNotInitialized
	}

	return s.mrEnclave[:], nil
}

// GetEncryptionPubKey returns the enclave's encryption public key
func (s *SGXEnclaveServiceImpl) GetEncryptionPubKey() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return nil, ErrEnclaveNotInitialized
	}

	return s.encryptPubKey, nil
}

// GetSigningPubKey returns the enclave's signing public key
func (s *SGXEnclaveServiceImpl) GetSigningPubKey() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return nil, ErrEnclaveNotInitialized
	}

	return s.signingPubKey, nil
}

// GenerateAttestation generates a DCAP attestation quote
func (s *SGXEnclaveServiceImpl) GenerateAttestation(reportData []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return nil, ErrEnclaveNotInitialized
	}

	if len(reportData) > SGXReportDataSize {
		return nil, errors.New("report data too large")
	}

	// TODO: Real SGX implementation would:
	// 1. ECALL to generate report for QE target info
	// 2. Call sgx_qe_get_quote() to get DCAP quote
	// 3. Optionally fetch collateral from PCS/PCCS

	quote, err := s.simulateDCAPQuoteGeneration(reportData)
	if err != nil {
		return nil, fmt.Errorf("quote generation failed: %w", err)
	}

	return quote, nil
}

// RotateKeys initiates key rotation
func (s *SGXEnclaveServiceImpl) RotateKeys() error {
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
func (s *SGXEnclaveServiceImpl) GetStatus() EnclaveStatus {
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

// Shutdown gracefully shuts down the enclave
func (s *SGXEnclaveServiceImpl) Shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.initialized {
		return nil
	}

	// TODO: Real SGX implementation would:
	// 1. ECALL to seal critical state
	// 2. sgx_destroy_enclave()
	// 3. Clean up Quote Provider

	// Securely clear keys from memory
	s.scrubKeys()

	s.initialized = false
	return nil
}

// =============================================================================
// SGX-Specific Methods
// =============================================================================

// GetMRSigner returns the enclave signer measurement
func (s *SGXEnclaveServiceImpl) GetMRSigner() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return nil, ErrEnclaveNotInitialized
	}

	return s.mrSigner[:], nil
}

// GetPlatformType returns PlatformSGX
func (s *SGXEnclaveServiceImpl) GetPlatformType() PlatformType {
	return PlatformSGX
}

// IsPlatformSecure returns true (SGX is secure)
func (s *SGXEnclaveServiceImpl) IsPlatformSecure() bool {
	return !s.config.Debug
}

// GenerateDCAPReport generates an SGX report for the Quote Enclave
func (s *SGXEnclaveServiceImpl) GenerateDCAPReport(targetInfo []byte, reportData []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return nil, ErrEnclaveNotInitialized
	}

	// TODO: Real SGX implementation would call EREPORT instruction
	// This generates a report that can be verified by the QE

	report := s.simulateSGXReport(targetInfo, reportData)
	return report, nil
}

// SealData seals data to the enclave measurement
func (s *SGXEnclaveServiceImpl) SealData(plaintext []byte, aad []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return nil, ErrEnclaveNotInitialized
	}

	// TODO: Real SGX implementation would call sgx_seal_data()
	// This encrypts data with a key derived from enclave measurement

	sealed, err := s.simulateSealData(plaintext, aad)
	if err != nil {
		return nil, fmt.Errorf("seal failed: %w", err)
	}

	return sealed, nil
}

// UnsealData unseals data previously sealed by this enclave
func (s *SGXEnclaveServiceImpl) UnsealData(sealed []byte) ([]byte, []byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return nil, nil, ErrEnclaveNotInitialized
	}

	// TODO: Real SGX implementation would call sgx_unseal_data()
	// This decrypts data only if enclave measurement matches

	plaintext, aad, err := s.simulateUnsealData(sealed)
	if err != nil {
		return nil, nil, fmt.Errorf("unseal failed: %w", err)
	}

	return plaintext, aad, nil
}

// VerifyMeasurement checks if a measurement is in the allowlist
func (s *SGXEnclaveServiceImpl) VerifyMeasurement(mrEnclave []byte) bool {
	// TODO: In production, check against on-chain governance allowlist
	// For POC, accept any non-zero measurement

	if len(mrEnclave) != SGXMREnclaveSize {
		return false
	}

	// Check if all zeros (invalid)
	allZero := true
	for _, b := range mrEnclave {
		if b != 0 {
			allZero = false
			break
		}
	}

	return !allZero
}

// =============================================================================
// Simulation Methods (POC Only - Replace with Real SGX Calls)
// =============================================================================

// simulateEnclaveCreation simulates SGX enclave creation
func (s *SGXEnclaveServiceImpl) simulateEnclaveCreation() error {
	// Simulate MRENCLAVE (hash of enclave code)
	// In real SGX, this is computed during enclave loading
	enclaveCode := []byte("veid_scoring_enclave_v1_" + s.config.EnclavePath)
	mrEnclave := sha256.Sum256(enclaveCode)
	copy(s.mrEnclave[:], mrEnclave[:])

	// Simulate MRSIGNER (hash of signing key)
	// In real SGX, this is the hash of the ISV's public key
	signerKey := []byte("virtengine_enclave_signer_key")
	mrSigner := sha256.Sum256(signerKey)
	copy(s.mrSigner[:], mrSigner[:])

	// Derive seal key (in real SGX, this uses EGETKEY with key_policy)
	// Key is bound to MRENCLAVE and platform
	sealKeyMaterial := append(s.mrEnclave[:], []byte("seal_key_derive")...)
	s.sealKey = sha256Bytes(sealKeyMaterial)[:SGXSealKeySize]

	return nil
}

// deriveEnclaveKeys derives encryption and signing keys
func (s *SGXEnclaveServiceImpl) deriveEnclaveKeys() error {
	// Derive keys using HKDF from seal key and epoch
	salt := append(s.mrEnclave[:], s.mrSigner[:]...)

	// Encryption key (X25519 seed)
	encInfo := fmt.Sprintf("encryption_key_epoch_%d", s.currentEpoch)
	s.encryptionKey = s.hkdfDerive(s.sealKey, salt, []byte(encInfo), 32)

	// Signing key (Ed25519 seed)
	sigInfo := fmt.Sprintf("signing_key_epoch_%d", s.currentEpoch)
	s.signingKey = s.hkdfDerive(s.sealKey, salt, []byte(sigInfo), 32)

	// Derive public keys (simplified - in real impl use proper EC ops)
	s.encryptPubKey = sha256Bytes(append([]byte("X25519_pub_"), s.encryptionKey...))
	s.signingPubKey = sha256Bytes(append([]byte("Ed25519_pub_"), s.signingKey...))

	return nil
}

// hkdfDerive performs HKDF key derivation
func (s *SGXEnclaveServiceImpl) hkdfDerive(secret, salt, info []byte, length int) []byte {
	reader := hkdf.New(sha256.New, secret, salt, info)
	key := make([]byte, length)
	_, _ = reader.Read(key)
	return key
}

// simulateEnclaveScoring simulates scoring inside the enclave
func (s *SGXEnclaveServiceImpl) simulateEnclaveScoring(request *ScoringRequest) *ScoringResult {
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
		sha256Bytes([]byte("sgx_face_embedding")),
		sha256Bytes([]byte("sgx_document_features")),
	}

	// Generate model version hash
	modelVersionHash := sha256Bytes([]byte("sgx_veid_model_v1.0.0"))

	// Sign result inside enclave
	signingPayload := s.computeSigningPayload(request.RequestID, score, status, inputHash[:])
	enclaveSignature := s.signInsideEnclave(signingPayload)

	return &ScoringResult{
		RequestID:            request.RequestID,
		Score:                score,
		Status:               status,
		ReasonCodes:          []string{"sgx_score_" + status},
		EvidenceHashes:       evidenceHashes,
		ModelVersionHash:     modelVersionHash,
		InputHash:            inputHash[:],
		EnclaveSignature:     enclaveSignature,
		MeasurementHash:      s.mrEnclave[:],
		AttestationReference: sha256Bytes([]byte("sgx_dcap_attestation_ref")),
	}
}

// computeSigningPayload computes the payload to sign
func (s *SGXEnclaveServiceImpl) computeSigningPayload(requestID string, score uint32, status string, inputHash []byte) []byte {
	h := sha256.New()
	h.Write([]byte(requestID))
	h.Write([]byte{byte(score >> 24), byte(score >> 16), byte(score >> 8), byte(score)})
	h.Write([]byte(status))
	h.Write(inputHash)
	h.Write(s.mrEnclave[:])
	return h.Sum(nil)
}

// signInsideEnclave simulates signing inside the enclave
func (s *SGXEnclaveServiceImpl) signInsideEnclave(payload []byte) []byte {
	// In real SGX, this uses Ed25519 with sealed key
	h := sha256.New()
	h.Write(s.signingKey)
	h.Write(payload)
	return h.Sum(nil)
}

// simulateDCAPQuoteGeneration simulates DCAP quote generation
func (s *SGXEnclaveServiceImpl) simulateDCAPQuoteGeneration(reportData []byte) ([]byte, error) {
	// Build quote header
	header := SGXQuoteHeader{
		Version:    SGXQuoteVersionDCAP,
		AttKeyType: 2, // ECDSA-256-with-P-256 curve
		TEEType:    0, // SGX
	}
	copy(header.QEVendorID[:], []byte("INTEL_SGX_QE_ID_"))
	copy(header.UserData[:], reportData[:min(20, len(reportData))])

	// Build report body
	reportBody := SGXReportBody{
		MREnclave: s.mrEnclave,
		MRSigner:  s.mrSigner,
		ISVProdID: 1,
		ISVSVN:    1,
		Attributes: SGXAttributes{
			Flags: SGXFlagInitted | SGXFlagMode64Bit,
			Xfrm:  3,
		},
	}

	// Copy report data
	copy(reportBody.ReportData[:], reportData)

	// Serialize quote (simplified)
	quote := make([]byte, 0, 1024)

	// Header
	quote = append(quote, byte(header.Version), byte(header.Version>>8))
	quote = append(quote, byte(header.AttKeyType), byte(header.AttKeyType>>8))
	quote = binary.LittleEndian.AppendUint32(quote, header.TEEType)
	quote = binary.LittleEndian.AppendUint32(quote, header.Reserved)
	quote = append(quote, header.QEVendorID[:]...)
	quote = append(quote, header.UserData[:]...)

	// Report body (simplified - actual format is 384 bytes)
	quote = append(quote, reportBody.CPUSVN[:]...)
	quote = append(quote, reportBody.MREnclave[:]...)
	quote = append(quote, reportBody.MRSigner[:]...)
	quote = binary.LittleEndian.AppendUint16(quote, reportBody.ISVProdID)
	quote = binary.LittleEndian.AppendUint16(quote, reportBody.ISVSVN)
	quote = append(quote, reportBody.ReportData[:]...)

	// Signature (simulated ECDSA)
	sigPayload := sha256.Sum256(quote)
	signature := s.signInsideEnclave(sigPayload[:])
	quote = binary.LittleEndian.AppendUint32(quote, uint32(len(signature)))
	quote = append(quote, signature...)

	return quote, nil
}

// simulateSGXReport simulates SGX report generation
func (s *SGXEnclaveServiceImpl) simulateSGXReport(targetInfo []byte, reportData []byte) []byte {
	// In real SGX, this calls EREPORT instruction
	report := make([]byte, SGXReportSize)

	// Copy measurement
	copy(report[64:96], s.mrEnclave[:])
	copy(report[128:160], s.mrSigner[:])

	// Copy report data
	if len(reportData) > 0 {
		copy(report[320:384], reportData[:min(64, len(reportData))])
	}

	// Compute MAC (simulated)
	mac := sha256.Sum256(append(report, targetInfo...))
	copy(report[384:], mac[:16])

	return report
}

// simulateSealData simulates SGX data sealing
func (s *SGXEnclaveServiceImpl) simulateSealData(plaintext []byte, aad []byte) ([]byte, error) {
	// TODO: Real implementation uses sgx_seal_data with AES-GCM
	// Key is derived from EGETKEY with key_policy = MRENCLAVE

	// Generate nonce
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Simulate AES-GCM encryption (in POC, just XOR with key hash)
	// NEVER use this in production - use real AES-GCM
	keyStream := sha256Bytes(append(s.sealKey, nonce...))
	ciphertext := make([]byte, len(plaintext))
	for i := range plaintext {
		ciphertext[i] = plaintext[i] ^ keyStream[i%len(keyStream)]
	}

	// Build sealed blob
	sealed := make([]byte, 0, 64+len(ciphertext)+16)
	sealed = append(sealed, []byte("SEAL")...)           // Magic
	sealed = append(sealed, 0x01, 0x00)                  // Version
	sealed = append(sealed, byte(PlatformSGX[0]), 0x00)  // Platform
	sealed = append(sealed, s.mrEnclave[:]...)           // Measurement
	sealed = append(sealed, nonce...)                    // Nonce
	sealed = append(sealed, ciphertext...)               // Ciphertext

	// Auth tag (simulated)
	tag := sha256.Sum256(append(sealed, aad...))
	sealed = append(sealed, tag[:16]...)

	return sealed, nil
}

// simulateUnsealData simulates SGX data unsealing
func (s *SGXEnclaveServiceImpl) simulateUnsealData(sealed []byte) ([]byte, []byte, error) {
	// Verify minimum length
	if len(sealed) < 64 {
		return nil, nil, errors.New("sealed data too short")
	}

	// Verify magic
	if string(sealed[:4]) != "SEAL" {
		return nil, nil, errors.New("invalid sealed data format")
	}

	// Extract measurement and verify
	measurement := sealed[8:40]
	if !s.VerifyMeasurement(measurement) {
		return nil, nil, errors.New("measurement mismatch - cannot unseal")
	}

	// Extract nonce and ciphertext
	nonce := sealed[40:52]
	ciphertext := sealed[52 : len(sealed)-16]

	// Simulate decryption
	keyStream := sha256Bytes(append(s.sealKey, nonce...))
	plaintext := make([]byte, len(ciphertext))
	for i := range ciphertext {
		plaintext[i] = ciphertext[i] ^ keyStream[i%len(keyStream)]
	}

	return plaintext, nil, nil
}

// scrubKeys securely clears keys from memory
func (s *SGXEnclaveServiceImpl) scrubKeys() {
	// Use ScrubBytes from memory_scrub.go
	if s.sealKey != nil {
		ScrubBytes(s.sealKey)
	}
	if s.encryptionKey != nil {
		ScrubBytes(s.encryptionKey)
	}
	if s.signingKey != nil {
		ScrubBytes(s.signingKey)
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
