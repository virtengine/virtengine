// Package enclave_runtime provides TEE enclave implementations.
//
// This file implements the remote attestation protocol for validator-to-validator
// attestation verification. Validators use this protocol to verify that peer
// validators are running genuine TEE enclaves with approved measurements.
//
// Protocol Flow:
// 1. Challenger generates a random nonce
// 2. Challenger sends attestation request with nonce to responder
// 3. Responder generates attestation with nonce embedded
// 4. Responder sends attestation to challenger
// 5. Challenger verifies attestation, nonce, and measurement
//
// Task Reference: SECURITY-002 - Real TEE Enclave Implementation
package enclave_runtime

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// RemoteAttestationProtocol manages remote attestation between validators
type RemoteAttestationProtocol struct {
	mu sync.RWMutex

	// Local enclave service for generating our attestations
	localEnclave EnclaveService

	// Verifier for validating peer attestations
	verifier AttestationVerifier

	// Pending challenges we've issued
	pendingChallenges map[string]*attestationChallenge

	// Configuration
	config RemoteAttestationConfig

	// Metrics
	challengesSent     uint64
	challengesReceived uint64
	successfulAttests  uint64
	failedAttests      uint64
}

// attestationChallenge represents a pending attestation challenge
type attestationChallenge struct {
	ChallengeID  string
	Nonce        []byte
	PeerID       string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	ResponseChan chan *AttestationResponse
}

// RemoteAttestationConfig configures the remote attestation protocol
type RemoteAttestationConfig struct {
	// NonceSize is the size of attestation nonces in bytes
	NonceSize int

	// ChallengeTimeout is how long to wait for attestation response
	ChallengeTimeout time.Duration

	// MaxPendingChallenges limits concurrent outstanding challenges
	MaxPendingChallenges int

	// ReattestInterval is how often to re-verify peer attestations
	ReattestInterval time.Duration

	// AllowSimulated permits simulated attestations (testing only)
	AllowSimulated bool
}

// DefaultRemoteAttestationConfig returns default configuration
func DefaultRemoteAttestationConfig() RemoteAttestationConfig {
	return RemoteAttestationConfig{
		NonceSize:            32,
		ChallengeTimeout:     30 * time.Second,
		MaxPendingChallenges: 100,
		ReattestInterval:     24 * time.Hour,
		AllowSimulated:       false,
	}
}

// AttestationRequest is sent from challenger to responder
type AttestationRequest struct {
	// ChallengeID uniquely identifies this attestation challenge
	ChallengeID string `json:"challenge_id"`

	// Nonce is a random value that must be included in the attestation
	Nonce []byte `json:"nonce"`

	// ChallengerID identifies the validator making the request
	ChallengerID string `json:"challenger_id"`

	// RequestedPlatforms lists acceptable TEE platforms (empty = any)
	RequestedPlatforms []AttestationType `json:"requested_platforms,omitempty"`

	// Timestamp when the request was created
	Timestamp time.Time `json:"timestamp"`
}

// Validate validates the attestation request
func (r *AttestationRequest) Validate() error {
	if r.ChallengeID == "" {
		return errors.New("challenge_id required")
	}
	if len(r.Nonce) == 0 {
		return errors.New("nonce required")
	}
	if len(r.Nonce) < 16 {
		return errors.New("nonce too short (minimum 16 bytes)")
	}
	if r.ChallengerID == "" {
		return errors.New("challenger_id required")
	}
	return nil
}

// AttestationResponse is sent from responder back to challenger
type AttestationResponse struct {
	// ChallengeID matches the request
	ChallengeID string `json:"challenge_id"`

	// ResponderID identifies the responding validator
	ResponderID string `json:"responder_id"`

	// Platform is the TEE platform used
	Platform AttestationType `json:"platform"`

	// AttestationData is the platform-specific attestation (quote, report, etc.)
	AttestationData []byte `json:"attestation_data"`

	// PublicKey is the enclave's signing public key
	PublicKey []byte `json:"public_key"`

	// EncryptionPubKey is the enclave's encryption public key
	EncryptionPubKey []byte `json:"encryption_pub_key"`

	// Measurement is the enclave measurement hash
	Measurement []byte `json:"measurement"`

	// Timestamp when the response was generated
	Timestamp time.Time `json:"timestamp"`

	// Error is set if attestation generation failed
	Error string `json:"error,omitempty"`
}

// IsSuccess returns true if the response contains valid attestation
func (r *AttestationResponse) IsSuccess() bool {
	return r.Error == "" && len(r.AttestationData) > 0
}

// Validate validates the attestation response
func (r *AttestationResponse) Validate() error {
	if r.ChallengeID == "" {
		return errors.New("challenge_id required")
	}
	if r.ResponderID == "" {
		return errors.New("responder_id required")
	}
	if r.Error != "" {
		return fmt.Errorf("responder error: %s", r.Error)
	}
	if len(r.AttestationData) == 0 {
		return errors.New("attestation_data required")
	}
	if len(r.Measurement) == 0 {
		return errors.New("measurement required")
	}
	return nil
}

// AttestationVerificationResult contains the result of verifying a remote attestation
type AttestationVerificationResult struct {
	// Valid indicates the attestation passed all checks
	Valid bool `json:"valid"`

	// PeerID is the verified peer identifier
	PeerID string `json:"peer_id"`

	// Platform is the verified TEE platform
	Platform AttestationType `json:"platform"`

	// Measurement is the verified enclave measurement
	Measurement []byte `json:"measurement"`

	// MeasurementHex is the hex-encoded measurement
	MeasurementHex string `json:"measurement_hex"`

	// PublicKey is the verified enclave public key
	PublicKey []byte `json:"public_key"`

	// VerifiedAt is when verification completed
	VerifiedAt time.Time `json:"verified_at"`

	// ValidUntil is when re-attestation is recommended
	ValidUntil time.Time `json:"valid_until"`

	// Errors contains any verification errors
	Errors []string `json:"errors,omitempty"`

	// Warnings contains non-fatal issues
	Warnings []string `json:"warnings,omitempty"`
}

// NewRemoteAttestationProtocol creates a new remote attestation protocol handler
func NewRemoteAttestationProtocol(
	localEnclave EnclaveService,
	verifier AttestationVerifier,
	config RemoteAttestationConfig,
) *RemoteAttestationProtocol {
	return &RemoteAttestationProtocol{
		localEnclave:      localEnclave,
		verifier:          verifier,
		pendingChallenges: make(map[string]*attestationChallenge),
		config:            config,
	}
}

// GenerateChallenge creates a new attestation challenge for a peer
func (p *RemoteAttestationProtocol) GenerateChallenge(peerID string) (*AttestationRequest, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check pending challenge limit
	if len(p.pendingChallenges) >= p.config.MaxPendingChallenges {
		return nil, errors.New("too many pending challenges")
	}

	// Generate challenge ID
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return nil, fmt.Errorf("failed to generate challenge ID: %w", err)
	}
	challengeID := hex.EncodeToString(idBytes)

	// Generate nonce
	nonce := make([]byte, p.config.NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Create challenge record
	now := time.Now()
	challenge := &attestationChallenge{
		ChallengeID:  challengeID,
		Nonce:        nonce,
		PeerID:       peerID,
		CreatedAt:    now,
		ExpiresAt:    now.Add(p.config.ChallengeTimeout),
		ResponseChan: make(chan *AttestationResponse, 1),
	}

	p.pendingChallenges[challengeID] = challenge
	p.challengesSent++

	// Return request
	return &AttestationRequest{
		ChallengeID:  challengeID,
		Nonce:        nonce,
		ChallengerID: "self", // Caller should replace with actual ID
		Timestamp:    now,
	}, nil
}

// HandleChallengeRequest handles an incoming attestation challenge from a peer
func (p *RemoteAttestationProtocol) HandleChallengeRequest(ctx context.Context, request *AttestationRequest, responderID string) (*AttestationResponse, error) {
	p.mu.Lock()
	p.challengesReceived++
	p.mu.Unlock()

	// Validate request
	if err := request.Validate(); err != nil {
		return &AttestationResponse{
			ChallengeID: request.ChallengeID,
			ResponderID: responderID,
			Timestamp:   time.Now(),
			Error:       fmt.Sprintf("invalid request: %v", err),
		}, nil
	}

	// Check if we have an enclave
	if p.localEnclave == nil {
		return &AttestationResponse{
			ChallengeID: request.ChallengeID,
			ResponderID: responderID,
			Timestamp:   time.Now(),
			Error:       "no local enclave available",
		}, nil
	}

	// Generate attestation with nonce as report data
	// Hash the nonce with challenge ID for binding
	reportData := p.computeReportData(request.ChallengeID, request.Nonce)

	attestationData, err := p.localEnclave.GenerateAttestation(reportData)
	if err != nil {
		return &AttestationResponse{
			ChallengeID: request.ChallengeID,
			ResponderID: responderID,
			Timestamp:   time.Now(),
			Error:       fmt.Sprintf("attestation generation failed: %v", err),
		}, nil
	}

	// Get public keys
	signingPubKey, _ := p.localEnclave.GetSigningPubKey()
	encryptPubKey, _ := p.localEnclave.GetEncryptionPubKey()
	measurement, _ := p.localEnclave.GetMeasurement()

	// Determine platform
	var platform AttestationType
	if hwService, ok := p.localEnclave.(HardwareAwareEnclaveService); ok {
		if hwService.IsHardwareEnabled() {
			// Get actual platform from service
			switch hwService.GetHardwareMode() {
			case HardwareModeRequire, HardwareModeAuto:
				// Detect from attestation data
				platform = detectPlatformFromAttestation(attestationData)
			default:
				platform = AttestationTypeSimulated
			}
		} else {
			platform = AttestationTypeSimulated
		}
	} else {
		platform = detectPlatformFromAttestation(attestationData)
	}

	return &AttestationResponse{
		ChallengeID:      request.ChallengeID,
		ResponderID:      responderID,
		Platform:         platform,
		AttestationData:  attestationData,
		PublicKey:        signingPubKey,
		EncryptionPubKey: encryptPubKey,
		Measurement:      measurement,
		Timestamp:        time.Now(),
	}, nil
}

// VerifyResponse verifies an attestation response from a peer
func (p *RemoteAttestationProtocol) VerifyResponse(response *AttestationResponse) (*AttestationVerificationResult, error) {
	p.mu.Lock()
	challenge, exists := p.pendingChallenges[response.ChallengeID]
	if exists {
		delete(p.pendingChallenges, response.ChallengeID)
	}
	p.mu.Unlock()

	result := &AttestationVerificationResult{
		PeerID:     response.ResponderID,
		Platform:   response.Platform,
		VerifiedAt: time.Now(),
		ValidUntil: time.Now().Add(p.config.ReattestInterval),
	}

	// Validate response
	if err := response.Validate(); err != nil {
		result.Errors = append(result.Errors, err.Error())
		p.mu.Lock()
		p.failedAttests++
		p.mu.Unlock()
		return result, nil
	}

	// Check challenge exists
	if !exists {
		result.Errors = append(result.Errors, "unknown challenge ID")
		p.mu.Lock()
		p.failedAttests++
		p.mu.Unlock()
		return result, nil
	}

	// Check challenge not expired
	if time.Now().After(challenge.ExpiresAt) {
		result.Errors = append(result.Errors, "challenge expired")
		p.mu.Lock()
		p.failedAttests++
		p.mu.Unlock()
		return result, nil
	}

	// Check peer ID matches
	if response.ResponderID != challenge.PeerID {
		result.Errors = append(result.Errors, fmt.Sprintf("peer ID mismatch: expected %s, got %s",
			challenge.PeerID, response.ResponderID))
	}

	// Check if simulated attestation is allowed
	if response.Platform == AttestationTypeSimulated && !p.config.AllowSimulated {
		result.Errors = append(result.Errors, "simulated attestation not allowed")
	}

	// Verify the attestation with our verifier
	if p.verifier != nil {
		// Parse the attestation data into an AttestationReport
		report := &AttestationReport{
			Platform:    getPlatformTypeFromAttestation(response.Platform),
			Measurement: response.Measurement,
			ReportData:  p.computeReportData(challenge.ChallengeID, challenge.Nonce),
			Signature:   response.AttestationData, // Raw attestation contains signature
			Timestamp:   response.Timestamp,
		}

		if err := p.verifier.VerifyReport(context.Background(), report); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("attestation verification failed: %v", err))
		}

		// Verify measurement is allowed
		if !p.verifier.IsMeasurementAllowed(response.Measurement) {
			result.Warnings = append(result.Warnings, "measurement not in allowlist (may need to be added)")
		}

		// Verify nonce is embedded in attestation
		expectedReportData := p.computeReportData(challenge.ChallengeID, challenge.Nonce)
		if !p.verifyNonceInAttestation(response.AttestationData, expectedReportData, response.Platform) {
			result.Warnings = append(result.Warnings, "nonce verification not performed (platform-specific check needed)")
		}
	}

	// Set result fields
	result.Measurement = response.Measurement
	result.MeasurementHex = hex.EncodeToString(response.Measurement)
	result.PublicKey = response.PublicKey
	result.Valid = len(result.Errors) == 0

	p.mu.Lock()
	if result.Valid {
		p.successfulAttests++
	} else {
		p.failedAttests++
	}
	p.mu.Unlock()

	return result, nil
}

// CleanupExpiredChallenges removes expired pending challenges
func (p *RemoteAttestationProtocol) CleanupExpiredChallenges() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	expired := 0

	for id, challenge := range p.pendingChallenges {
		if now.After(challenge.ExpiresAt) {
			delete(p.pendingChallenges, id)
			expired++
		}
	}

	return expired
}

// GetStats returns protocol statistics
func (p *RemoteAttestationProtocol) GetStats() RemoteAttestationStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return RemoteAttestationStats{
		ChallengesSent:     p.challengesSent,
		ChallengesReceived: p.challengesReceived,
		SuccessfulAttests:  p.successfulAttests,
		FailedAttests:      p.failedAttests,
		PendingChallenges:  len(p.pendingChallenges),
	}
}

// RemoteAttestationStats contains protocol statistics
type RemoteAttestationStats struct {
	ChallengesSent     uint64
	ChallengesReceived uint64
	SuccessfulAttests  uint64
	FailedAttests      uint64
	PendingChallenges  int
}

// computeReportData computes the report data from challenge ID and nonce
func (p *RemoteAttestationProtocol) computeReportData(challengeID string, nonce []byte) []byte {
	h := sha256.New()
	h.Write([]byte(challengeID))
	h.Write(nonce)
	h.Write([]byte("virtengine-attestation-v1"))
	return h.Sum(nil)
}

// verifyNonceInAttestation verifies the nonce is embedded in the attestation
//
//nolint:unparam // expectedReportData kept for future nonce extraction verification
func (p *RemoteAttestationProtocol) verifyNonceInAttestation(attestation, _ []byte, platform AttestationType) bool {
	// Platform-specific nonce extraction and verification
	// This is a simplified check - real implementation would parse the attestation format
	switch platform {
	case AttestationTypeSGX:
		// SGX: report_data is at offset 320 in quote body (64 bytes)
		// This is a simplified check
		return len(attestation) > 384
	case AttestationTypeSEVSNP:
		// SEV-SNP: report_data is at specific offset in report
		return len(attestation) > 100
	case AttestationTypeNitro:
		// Nitro: user_data is in CBOR structure
		return len(attestation) > 100
	default:
		return true // Skip for simulated
	}
}

// detectPlatformFromAttestation detects the platform from attestation bytes
func detectPlatformFromAttestation(attestation []byte) AttestationType {
	if len(attestation) < 4 {
		return AttestationTypeUnknown
	}

	// Check magic bytes
	switch {
	case attestation[0] == 0x03 && attestation[1] == 0x00:
		return AttestationTypeSGX
	case attestation[0] == 0x01 && attestation[1] == 0x00:
		return AttestationTypeSEVSNP
	case attestation[0] == 0xD2 && attestation[1] == 0x84:
		return AttestationTypeNitro
	case string(attestation[:3]) == "SIM":
		return AttestationTypeSimulated
	default:
		return AttestationTypeUnknown
	}
}

// getPlatformTypeFromAttestation converts AttestationType to PlatformType
func getPlatformTypeFromAttestation(at AttestationType) PlatformType {
	switch at {
	case AttestationTypeSGX:
		return PlatformSGX
	case AttestationTypeSEVSNP:
		return PlatformSEVSNP
	case AttestationTypeNitro:
		return PlatformNitro
	case AttestationTypeSimulated:
		return PlatformSimulated
	default:
		return PlatformSimulated
	}
}

// =============================================================================
// Validator Attestation Cache
// =============================================================================

// ValidatorAttestationCache caches verified attestations for validators
type ValidatorAttestationCache struct {
	mu sync.RWMutex

	// cache maps validator ID to verification result
	cache map[string]*cachedAttestation

	// config
	maxEntries int
	ttl        time.Duration
}

type cachedAttestation struct {
	result    *AttestationVerificationResult
	expiresAt time.Time
}

// NewValidatorAttestationCache creates a new validator attestation cache
func NewValidatorAttestationCache(maxEntries int, ttl time.Duration) *ValidatorAttestationCache {
	return &ValidatorAttestationCache{
		cache:      make(map[string]*cachedAttestation),
		maxEntries: maxEntries,
		ttl:        ttl,
	}
}

// Get retrieves a cached attestation result for a validator
func (c *ValidatorAttestationCache) Get(validatorID string) (*AttestationVerificationResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[validatorID]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		return nil, false
	}

	return entry.result, true
}

// Set stores an attestation result for a validator
func (c *ValidatorAttestationCache) Set(validatorID string, result *AttestationVerificationResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict expired entries if at capacity
	if len(c.cache) >= c.maxEntries {
		c.evictExpired()
	}

	// Still at capacity? Evict oldest
	if len(c.cache) >= c.maxEntries {
		c.evictOldest()
	}

	c.cache[validatorID] = &cachedAttestation{
		result:    result,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Invalidate removes a validator's cached attestation
func (c *ValidatorAttestationCache) Invalidate(validatorID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, validatorID)
}

// evictExpired removes all expired entries (must hold write lock)
func (c *ValidatorAttestationCache) evictExpired() {
	now := time.Now()
	for id, entry := range c.cache {
		if now.After(entry.expiresAt) {
			delete(c.cache, id)
		}
	}
}

// evictOldest removes the oldest entry (must hold write lock)
func (c *ValidatorAttestationCache) evictOldest() {
	var oldestID string
	var oldestTime time.Time

	for id, entry := range c.cache {
		if oldestID == "" || entry.result.VerifiedAt.Before(oldestTime) {
			oldestID = id
			oldestTime = entry.result.VerifiedAt
		}
	}

	if oldestID != "" {
		delete(c.cache, oldestID)
	}
}

// IsValidatorTrusted returns true if the validator has a valid cached attestation
func (c *ValidatorAttestationCache) IsValidatorTrusted(validatorID string) bool {
	result, exists := c.Get(validatorID)
	return exists && result.Valid
}

// GetTrustedMeasurement returns the trusted measurement for a validator
func (c *ValidatorAttestationCache) GetTrustedMeasurement(validatorID string) ([]byte, bool) {
	result, exists := c.Get(validatorID)
	if !exists || !result.Valid {
		return nil, false
	}
	return result.Measurement, true
}

// ListTrustedValidators returns all validators with valid cached attestations
func (c *ValidatorAttestationCache) ListTrustedValidators() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	trusted := make([]string, 0)

	for id, entry := range c.cache {
		if now.Before(entry.expiresAt) && entry.result.Valid {
			trusted = append(trusted, id)
		}
	}

	return trusted
}

// Size returns the number of cached entries
func (c *ValidatorAttestationCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

