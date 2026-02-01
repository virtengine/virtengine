// Package enclave_runtime provides the interface and implementations for
// Trusted Execution Environment (TEE) enclave operations.
//
// This package defines the enclave runtime service that receives encrypted
// identity data, performs decryption and ML scoring inside the enclave,
// and returns only enclave-signed results without exposing plaintext.
//
// Security Properties:
// - Decryption keys are generated inside the enclave and sealed
// - Plaintext data never leaves the enclave boundary
// - All results include cryptographic attestation
// - Memory is scrubbed after processing
package enclave_runtime

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

const (
	// statusNeedsReview is the needs_review verification status
	statusNeedsReview = "needs_review"
	// statusRejected is the rejected verification status
	statusRejected = "rejected"
)

// ErrEnclaveNotInitialized is returned when enclave operations are attempted before initialization
var ErrEnclaveNotInitialized = errors.New("enclave not initialized")

// ErrEnclaveUnavailable is returned when the enclave is temporarily unavailable
var ErrEnclaveUnavailable = errors.New("enclave temporarily unavailable")

// ErrInputTooLarge is returned when input exceeds maximum size
var ErrInputTooLarge = errors.New("input exceeds maximum size")

// ErrTimeout is returned when enclave operation exceeds timeout
var ErrTimeout = errors.New("enclave operation timeout")

// ErrDecryptionFailed is returned when decryption fails
var ErrDecryptionFailed = errors.New("decryption failed")

// ErrScoringFailed is returned when ML scoring fails
var ErrScoringFailed = errors.New("scoring failed")

// RuntimeConfig configures the enclave runtime
type RuntimeConfig struct {
	// MaxInputSize is the maximum input size in bytes
	MaxInputSize int64

	// MaxExecutionTimeMs is the maximum execution time in milliseconds
	MaxExecutionTimeMs int64

	// MaxConcurrentRequests is the maximum number of concurrent requests
	MaxConcurrentRequests int

	// ScrubIntervalMs is the interval for memory scrubbing in milliseconds
	ScrubIntervalMs int64

	// ModelPath is the path to the ML model inside the enclave
	ModelPath string

	// KeyRotationEpoch is the epoch for key rotation
	KeyRotationEpoch uint64
}

// DefaultRuntimeConfig returns the default runtime configuration
func DefaultRuntimeConfig() RuntimeConfig {
	return RuntimeConfig{
		MaxInputSize:          10 * 1024 * 1024, // 10 MB
		MaxExecutionTimeMs:    1000,             // 1 second
		MaxConcurrentRequests: 4,
		ScrubIntervalMs:       0, // Scrub after each request
		ModelPath:             "/enclave/model/veid_scoring_v1.bin",
		KeyRotationEpoch:      1000,
	}
}

// ScoringRequest represents a request to score identity data
type ScoringRequest struct {
	// RequestID is a unique identifier for this request
	RequestID string

	// Ciphertext is the encrypted identity payload
	Ciphertext []byte

	// WrappedKey is the data encryption key wrapped for this enclave
	WrappedKey []byte

	// Nonce is the encryption nonce
	Nonce []byte

	// Metadata contains optional request metadata
	Metadata map[string]string

	// BlockHeight is the block height for context
	BlockHeight int64

	// ScopeID is the identity scope being scored
	ScopeID string

	// AccountAddress is the account owning the identity
	AccountAddress string
}

// Validate validates the scoring request
func (r *ScoringRequest) Validate(config RuntimeConfig) error {
	if r.RequestID == "" {
		return errors.New("request ID required")
	}

	if len(r.Ciphertext) == 0 {
		return errors.New("ciphertext required")
	}

	if int64(len(r.Ciphertext)) > config.MaxInputSize {
		return ErrInputTooLarge
	}

	if len(r.WrappedKey) == 0 {
		return errors.New("wrapped key required")
	}

	if len(r.Nonce) == 0 {
		return errors.New("nonce required")
	}

	if r.ScopeID == "" {
		return errors.New("scope ID required")
	}

	if r.AccountAddress == "" {
		return errors.New("account address required")
	}

	return nil
}

// ScoringResult represents the output from enclave scoring
type ScoringResult struct {
	// RequestID matches the request
	RequestID string

	// Score is the computed identity score (0-100)
	Score uint32

	// Status is the verification status
	Status string

	// ReasonCodes are structured reason codes for the score
	ReasonCodes []string

	// EvidenceHashes are hashes of evidence artifacts
	EvidenceHashes [][]byte

	// ModelVersionHash is the hash of the ML model used
	ModelVersionHash []byte

	// InputHash is the hash of the input data
	InputHash []byte

	// EnclaveSignature is the signature from the enclave
	EnclaveSignature []byte

	// MeasurementHash is the enclave measurement
	MeasurementHash []byte

	// AttestationReference is a reference to the attestation
	AttestationReference []byte

	// ProcessingTimeMs is the processing time in milliseconds
	ProcessingTimeMs int64

	// Error contains any error message (empty on success)
	Error string
}

// IsSuccess returns true if the scoring succeeded
func (r *ScoringResult) IsSuccess() bool {
	return r.Error == ""
}

// SigningPayload returns the bytes that were signed
func (r *ScoringResult) SigningPayload() []byte {
	h := sha256.New()
	h.Write([]byte(r.RequestID))
	h.Write([]byte{byte(r.Score >> 24), byte(r.Score >> 16), byte(r.Score >> 8), byte(r.Score)})
	h.Write([]byte(r.Status))
	h.Write(r.ModelVersionHash)
	h.Write(r.InputHash)
	for _, eh := range r.EvidenceHashes {
		h.Write(eh)
	}
	h.Write(r.MeasurementHash)
	return h.Sum(nil)
}

// EnclaveService defines the interface for enclave operations
type EnclaveService interface {
	// Initialize initializes the enclave runtime
	Initialize(config RuntimeConfig) error

	// Score performs identity scoring inside the enclave
	Score(ctx context.Context, request *ScoringRequest) (*ScoringResult, error)

	// GetMeasurement returns the enclave measurement hash
	GetMeasurement() ([]byte, error)

	// GetEncryptionPubKey returns the enclave's encryption public key
	GetEncryptionPubKey() ([]byte, error)

	// GetSigningPubKey returns the enclave's signing public key
	GetSigningPubKey() ([]byte, error)

	// GenerateAttestation generates an attestation quote
	GenerateAttestation(reportData []byte) ([]byte, error)

	// RotateKeys initiates key rotation
	RotateKeys() error

	// GetStatus returns the enclave status
	GetStatus() EnclaveStatus

	// Shutdown gracefully shuts down the enclave
	Shutdown() error
}

// HardwareAwareEnclaveService extends EnclaveService with hardware status methods
type HardwareAwareEnclaveService interface {
	EnclaveService

	// IsHardwareEnabled returns true if real TEE hardware is being used
	IsHardwareEnabled() bool

	// GetHardwareMode returns the configured hardware mode
	GetHardwareMode() HardwareMode
}

// HardwareEnclaveConfig holds the configuration for creating hardware-aware enclave services
type HardwareEnclaveConfig struct {
	// Type specifies which TEE platform to use
	Type AttestationType

	// HardwareMode controls hardware vs simulation behavior
	HardwareMode HardwareMode

	// SGX-specific configuration
	SGXConfig *SGXEnclaveConfig

	// SEV-SNP-specific configuration
	SEVConfig *SEVSNPConfig

	// Nitro-specific configuration
	NitroConfig *NitroEnclaveConfig

	// Runtime configuration
	RuntimeConfig RuntimeConfig
}

// EnclaveStatus represents the status of the enclave
type EnclaveStatus struct {
	// Initialized indicates if the enclave is initialized
	Initialized bool

	// Available indicates if the enclave is available for requests
	Available bool

	// CurrentEpoch is the current key epoch
	CurrentEpoch uint64

	// ActiveRequests is the number of currently processing requests
	ActiveRequests int

	// TotalProcessed is the total number of requests processed
	TotalProcessed uint64

	// LastError is the last error encountered (if any)
	LastError string

	// Uptime is the uptime in seconds
	Uptime int64
}

// SimulatedEnclaveService provides a simulated enclave for testing and development
// WARNING: This is NOT secure and should only be used for development
type SimulatedEnclaveService struct {
	mu               sync.RWMutex
	config           RuntimeConfig
	initialized      bool
	encryptionPubKey []byte
	signingPubKey    []byte
	measurementHash  []byte
	epoch            uint64
	activeRequests   int
	totalProcessed   uint64
	startTime        time.Time
	lastError        string
}

// NewSimulatedEnclaveService creates a new simulated enclave service
// WARNING: For development/testing only - not secure
func NewSimulatedEnclaveService() *SimulatedEnclaveService {
	return &SimulatedEnclaveService{}
}

// Initialize initializes the simulated enclave
func (s *SimulatedEnclaveService) Initialize(config RuntimeConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = config
	s.initialized = true
	s.startTime = time.Now()
	s.epoch = 1

	// Generate simulated keys
	s.encryptionPubKey = generateSimulatedKey("encryption")
	s.signingPubKey = generateSimulatedKey("signing")
	s.measurementHash = generateSimulatedMeasurement()

	return nil
}

// Score simulates identity scoring
func (s *SimulatedEnclaveService) Score(ctx context.Context, request *ScoringRequest) (*ScoringResult, error) {
	s.mu.Lock()
	if !s.initialized {
		s.mu.Unlock()
		return nil, ErrEnclaveNotInitialized
	}
	if s.activeRequests >= s.config.MaxConcurrentRequests {
		s.mu.Unlock()
		return nil, ErrEnclaveUnavailable
	}
	s.activeRequests++
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.activeRequests--
		s.totalProcessed++
		s.mu.Unlock()
	}()

	// Validate request
	if err := request.Validate(s.config); err != nil {
		return &ScoringResult{
			RequestID: request.RequestID,
			Error:     err.Error(),
		}, nil
	}

	startTime := time.Now()

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(s.config.MaxExecutionTimeMs)*time.Millisecond)
	defer cancel()

	// Simulate processing
	resultCh := make(chan *ScoringResult, 1)
	verrors.SafeGo("simulated-scoring", func() {
		defer func() {}() // WG Done if needed
		result := s.simulateScoring(request)
		result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
		resultCh <- result
	})

	select {
	case <-timeoutCtx.Done():
		return &ScoringResult{
			RequestID: request.RequestID,
			Error:     ErrTimeout.Error(),
		}, nil
	case result := <-resultCh:
		return result, nil
	}
}

// simulateScoring simulates the actual scoring process
func (s *SimulatedEnclaveService) simulateScoring(request *ScoringRequest) *ScoringResult {
	// Simulate decryption and scoring (in real enclave this happens securely)

	// Compute input hash
	inputHash := sha256.Sum256(request.Ciphertext)

	// Simulate score computation (in real implementation, ML model runs here)
	// Using a deterministic score based on input hash for reproducibility
	score := uint32(inputHash[0]) % 101

	// Determine status based on score
	var status string
	switch {
	case score >= 80:
		status = "verified"
	case score >= 50:
		status = statusNeedsReview
	default:
		status = statusRejected
	}

	// Generate evidence hashes
	evidenceHashes := [][]byte{
		sha256Bytes([]byte("face_embedding")),
		sha256Bytes([]byte("document_features")),
	}

	// Generate model version hash
	modelVersionHash := sha256Bytes([]byte("veid_model_v1.0.0"))

	// Generate enclave signature (simulated)
	signingPayload := sha256Bytes(append(inputHash[:], []byte(status)...))
	enclaveSignature := s.simulateSign(signingPayload)

	return &ScoringResult{
		RequestID:            request.RequestID,
		Score:                score,
		Status:               status,
		ReasonCodes:          []string{"score_" + status},
		EvidenceHashes:       evidenceHashes,
		ModelVersionHash:     modelVersionHash,
		InputHash:            inputHash[:],
		EnclaveSignature:     enclaveSignature,
		MeasurementHash:      s.measurementHash,
		AttestationReference: sha256Bytes([]byte("attestation_ref")),
	}
}

// simulateSign simulates enclave signing
func (s *SimulatedEnclaveService) simulateSign(payload []byte) []byte {
	// In real enclave, this uses the sealed signing key
	h := sha256.New()
	h.Write(s.signingPubKey)
	h.Write(payload)
	return h.Sum(nil)
}

// GetMeasurement returns the enclave measurement
func (s *SimulatedEnclaveService) GetMeasurement() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.initialized {
		return nil, ErrEnclaveNotInitialized
	}
	return s.measurementHash, nil
}

// GetEncryptionPubKey returns the encryption public key
func (s *SimulatedEnclaveService) GetEncryptionPubKey() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.initialized {
		return nil, ErrEnclaveNotInitialized
	}
	return s.encryptionPubKey, nil
}

// GetSigningPubKey returns the signing public key
func (s *SimulatedEnclaveService) GetSigningPubKey() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.initialized {
		return nil, ErrEnclaveNotInitialized
	}
	return s.signingPubKey, nil
}

// GenerateAttestation generates a simulated attestation
func (s *SimulatedEnclaveService) GenerateAttestation(reportData []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.initialized {
		return nil, ErrEnclaveNotInitialized
	}

	// Simulated attestation quote structure
	h := sha256.New()
	h.Write([]byte("SIMULATED_ATTESTATION_V1"))
	h.Write(s.measurementHash)
	h.Write(reportData)
	return h.Sum(nil), nil
}

// RotateKeys simulates key rotation
func (s *SimulatedEnclaveService) RotateKeys() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.initialized {
		return ErrEnclaveNotInitialized
	}

	s.epoch++
	s.encryptionPubKey = generateSimulatedKey("encryption_" + hex.EncodeToString([]byte{byte(s.epoch)}))
	s.signingPubKey = generateSimulatedKey("signing_" + hex.EncodeToString([]byte{byte(s.epoch)}))

	return nil
}

// GetStatus returns the enclave status
func (s *SimulatedEnclaveService) GetStatus() EnclaveStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var uptime int64
	if s.initialized {
		uptime = int64(time.Since(s.startTime).Seconds())
	}

	return EnclaveStatus{
		Initialized:    s.initialized,
		Available:      s.initialized && s.activeRequests < s.config.MaxConcurrentRequests,
		CurrentEpoch:   s.epoch,
		ActiveRequests: s.activeRequests,
		TotalProcessed: s.totalProcessed,
		LastError:      s.lastError,
		Uptime:         uptime,
	}
}

// Shutdown shuts down the simulated enclave
func (s *SimulatedEnclaveService) Shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.initialized = false
	return nil
}

// Helper functions

func generateSimulatedKey(seed string) []byte {
	h := sha256.Sum256([]byte("simulated_key_" + seed))
	return h[:]
}

func generateSimulatedMeasurement() []byte {
	h := sha256.Sum256([]byte("simulated_enclave_measurement_v1"))
	return h[:]
}

func sha256Bytes(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}
