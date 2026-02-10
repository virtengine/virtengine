// Package email provides email verification attestation creation.
//
// This file implements signed verification attestation creation with:
// - Cryptographic signing of verification results
// - Attestation lifecycle management
// - Integration with the signer service
// - On-chain submission preparation
//
// Task Reference: VE-3F - Email Verification Delivery + Attestation
package email

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/virtengine/virtengine/pkg/errors"
	"github.com/virtengine/virtengine/pkg/verification/audit"
	"github.com/virtengine/virtengine/pkg/verification/signer"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Attestation Service Configuration
// ============================================================================

// AttestationConfig contains configuration for the attestation service.
type AttestationConfig struct {
	// ValidityDays is how long attestations remain valid
	ValidityDays int `json:"validity_days"`

	// IncludeMetadata determines if metadata is included in attestations
	IncludeMetadata bool `json:"include_metadata"`

	// RequireLivenessScore requires liveness check for email attestations
	RequireLivenessScore bool `json:"require_liveness_score"`

	// MinVerificationScore is the minimum score for attestation creation
	MinVerificationScore uint32 `json:"min_verification_score"`

	// OrganizationalEmailBonus is the score bonus for organizational emails
	OrganizationalEmailBonus uint32 `json:"organizational_email_bonus"`

	// VerifiedDomainBonus is the score bonus for verified domains
	VerifiedDomainBonus uint32 `json:"verified_domain_bonus"`

	// EnableChainSubmission enables automatic on-chain submission
	EnableChainSubmission bool `json:"enable_chain_submission"`
}

// DefaultAttestationConfig returns the default attestation configuration.
func DefaultAttestationConfig() AttestationConfig {
	return AttestationConfig{
		ValidityDays:             365,
		IncludeMetadata:          true,
		RequireLivenessScore:     false,
		MinVerificationScore:     50,
		OrganizationalEmailBonus: 10,
		VerifiedDomainBonus:      5,
		EnableChainSubmission:    true,
	}
}

// Validate validates the configuration.
func (c *AttestationConfig) Validate() error {
	if c.ValidityDays <= 0 {
		return errors.Wrap(ErrInvalidConfig, "validity_days must be positive")
	}
	if c.MinVerificationScore > 100 {
		return errors.Wrap(ErrInvalidConfig, "min_verification_score cannot exceed 100")
	}
	return nil
}

// ============================================================================
// Attestation Service
// ============================================================================

// AttestationService handles email verification attestation creation and management.
type AttestationService struct {
	config  AttestationConfig
	signer  signer.SignerService
	auditor audit.AuditLogger
	metrics *Metrics
	logger  zerolog.Logger

	// State
	mu                sync.RWMutex
	attestationCache  map[string]*veidtypes.VerificationAttestation
	pendingSubmission []string
	closed            bool
}

// AttestationServiceOption is a functional option for configuring the attestation service.
type AttestationServiceOption func(*AttestationService)

// WithAttestationAuditor sets the audit logger.
func WithAttestationAuditor(auditor audit.AuditLogger) AttestationServiceOption {
	return func(s *AttestationService) {
		s.auditor = auditor
	}
}

// WithAttestationMetrics sets the metrics collector.
func WithAttestationMetrics(m *Metrics) AttestationServiceOption {
	return func(s *AttestationService) {
		s.metrics = m
	}
}

// NewAttestationService creates a new attestation service.
func NewAttestationService(
	config AttestationConfig,
	signerSvc signer.SignerService,
	logger zerolog.Logger,
	opts ...AttestationServiceOption,
) (*AttestationService, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	if signerSvc == nil {
		return nil, errors.Wrap(ErrInvalidConfig, "signer service is required")
	}

	s := &AttestationService{
		config:            config,
		signer:            signerSvc,
		logger:            logger.With().Str("component", "email_attestation").Logger(),
		attestationCache:  make(map[string]*veidtypes.VerificationAttestation),
		pendingSubmission: make([]string, 0),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

// ============================================================================
// Attestation Creation
// ============================================================================

// AttestationRequest contains parameters for creating an attestation.
type AttestationRequest struct {
	// Challenge is the verified email challenge
	Challenge *EmailChallenge `json:"challenge"`

	// VerificationScore is the verification score (0-100)
	VerificationScore uint32 `json:"verification_score"`

	// ConfidenceScore is the confidence level (0-100)
	ConfidenceScore uint32 `json:"confidence_score"`

	// IsOrganizational indicates if this is an organizational email
	IsOrganizational bool `json:"is_organizational"`

	// IsDomainVerified indicates if the domain is verified
	IsDomainVerified bool `json:"is_domain_verified"`

	// RequestID is an optional request identifier
	RequestID string `json:"request_id,omitempty"`

	// AdditionalMetadata contains extra metadata to include
	AdditionalMetadata map[string]string `json:"additional_metadata,omitempty"`
}

// Validate validates the attestation request.
func (r *AttestationRequest) Validate() error {
	if r.Challenge == nil {
		return errors.Wrap(ErrInvalidRequest, "challenge is required")
	}
	if r.Challenge.Status != StatusVerified {
		return errors.Wrap(ErrInvalidRequest, "challenge must be verified")
	}
	if r.VerificationScore > 100 {
		return errors.Wrap(ErrInvalidRequest, "verification_score cannot exceed 100")
	}
	if r.ConfidenceScore > 100 {
		return errors.Wrap(ErrInvalidRequest, "confidence_score cannot exceed 100")
	}
	return nil
}

// AttestationResult contains the created attestation.
type AttestationResult struct {
	// Attestation is the created and signed attestation
	Attestation *veidtypes.VerificationAttestation `json:"attestation"`

	// AttestationID is the unique attestation ID
	AttestationID string `json:"attestation_id"`

	// AttestationHash is the SHA256 hash of the attestation
	AttestationHash string `json:"attestation_hash"`

	// Signature is the base64-encoded signature
	Signature string `json:"signature"`

	// SignerKeyID is the ID of the signing key
	SignerKeyID string `json:"signer_key_id"`

	// CreatedAt is when the attestation was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when the attestation expires
	ExpiresAt time.Time `json:"expires_at"`

	// ChainSubmitted indicates if submitted to chain
	ChainSubmitted bool `json:"chain_submitted"`
}

// CreateAttestation creates a signed verification attestation.
func (s *AttestationService) CreateAttestation(ctx context.Context, req *AttestationRequest) (*AttestationResult, error) {
	if s.closed {
		return nil, ErrServiceUnavailable
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check minimum score
	if req.VerificationScore < s.config.MinVerificationScore {
		return nil, errors.Wrapf(ErrAttestationFailed,
			"verification score %d below minimum %d", req.VerificationScore, s.config.MinVerificationScore)
	}

	// Calculate final score with bonuses
	finalScore := req.VerificationScore
	if req.IsOrganizational {
		finalScore += s.config.OrganizationalEmailBonus
	}
	if req.IsDomainVerified {
		finalScore += s.config.VerifiedDomainBonus
	}
	if finalScore > 100 {
		finalScore = 100
	}

	// Generate nonce
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return nil, errors.Wrap(ErrAttestationFailed, "failed to generate nonce")
	}

	now := time.Now()
	validityDuration := time.Duration(s.config.ValidityDays) * 24 * time.Hour

	// Create subject
	subject := veidtypes.NewAttestationSubject(req.Challenge.AccountAddress)
	subject.RequestID = req.Challenge.ChallengeID

	// Create attestation
	attestation := veidtypes.NewVerificationAttestation(
		veidtypes.AttestationIssuer{}, // Will be set by signer
		subject,
		veidtypes.AttestationTypeEmailVerification,
		nonceBytes,
		now,
		validityDuration,
		finalScore,
		req.ConfidenceScore,
	)

	// Add verification proof
	proofDetail := veidtypes.NewVerificationProofDetail(
		"email_verification",
		req.Challenge.EmailHash,
		finalScore,
		s.config.MinVerificationScore,
		now,
	)
	attestation.AddVerificationProof(proofDetail)

	// Add metadata
	if s.config.IncludeMetadata {
		attestation.SetMetadata("email_hash", truncateHash(req.Challenge.EmailHash))
		attestation.SetMetadata("domain_hash", truncateHash(req.Challenge.DomainHash))
		attestation.SetMetadata("method", string(req.Challenge.Method))
		attestation.SetMetadata("is_organizational", fmt.Sprintf("%t", req.IsOrganizational))
		attestation.SetMetadata("is_domain_verified", fmt.Sprintf("%t", req.IsDomainVerified))

		// Add custom metadata
		for k, v := range req.AdditionalMetadata {
			attestation.SetMetadata(k, v)
		}
	}

	// Sign the attestation
	if err := s.signer.SignAttestation(ctx, attestation); err != nil {
		return nil, errors.Wrap(ErrSigningFailed, err.Error())
	}

	// Calculate hash
	attestationHash, err := attestation.HashHex()
	if err != nil {
		s.logger.Warn().Err(err).Msg("failed to calculate attestation hash")
		attestationHash = ""
	}

	// Cache the attestation
	s.cacheAttestation(attestation)

	// Record metrics
	if s.metrics != nil {
		s.metrics.RecordAttestationCreated()
	}

	// Audit log
	if s.auditor != nil {
		s.auditor.Log(ctx, audit.Event{
			Type:      audit.EventTypeAttestationSigned,
			Timestamp: now,
			Actor:     req.Challenge.AccountAddress,
			Resource:  attestation.ID,
			Action:    "create_email_attestation",
			Details: map[string]interface{}{
				"challenge_id":    req.Challenge.ChallengeID,
				"score":           finalScore,
				"confidence":      req.ConfidenceScore,
				"is_org":          req.IsOrganizational,
				"domain_verified": req.IsDomainVerified,
				"key_id":          attestation.Proof.VerificationMethod,
			},
		})
	}

	s.logger.Info().
		Str("attestation_id", attestation.ID).
		Str("challenge_id", req.Challenge.ChallengeID).
		Str("account", req.Challenge.AccountAddress).
		Uint32("score", finalScore).
		Msg("email verification attestation created")

	return &AttestationResult{
		Attestation:     attestation,
		AttestationID:   attestation.ID,
		AttestationHash: attestationHash,
		Signature:       attestation.Proof.ProofValue,
		SignerKeyID:     attestation.Proof.VerificationMethod,
		CreatedAt:       now,
		ExpiresAt:       attestation.ExpiresAt,
	}, nil
}

// truncateHash truncates a hash for metadata (privacy).
func truncateHash(hash string) string {
	if len(hash) > 16 {
		return hash[:16] + "..."
	}
	return hash
}

// ============================================================================
// Attestation Retrieval
// ============================================================================

// GetAttestation retrieves an attestation by ID.
func (s *AttestationService) GetAttestation(ctx context.Context, attestationID string) (*veidtypes.VerificationAttestation, error) {
	s.mu.RLock()
	attestation, ok := s.attestationCache[attestationID]
	s.mu.RUnlock()

	if !ok {
		return nil, errors.Wrapf(ErrChallengeNotFound, "attestation not found: %s", attestationID)
	}

	return attestation, nil
}

// VerifyAttestation verifies an attestation's signature.
func (s *AttestationService) VerifyAttestation(ctx context.Context, attestation *veidtypes.VerificationAttestation) (bool, error) {
	return s.signer.VerifyAttestation(ctx, attestation)
}

// cacheAttestation stores an attestation in the cache.
func (s *AttestationService) cacheAttestation(attestation *veidtypes.VerificationAttestation) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.attestationCache[attestation.ID] = attestation

	if s.config.EnableChainSubmission {
		s.pendingSubmission = append(s.pendingSubmission, attestation.ID)
	}
}

// ============================================================================
// Batch Attestation
// ============================================================================

// BatchAttestationRequest contains multiple attestation requests.
type BatchAttestationRequest struct {
	// Requests are the individual attestation requests
	Requests []*AttestationRequest `json:"requests"`
}

// BatchAttestationResult contains the results of batch attestation creation.
type BatchAttestationResult struct {
	// TotalRequested is the number of attestations requested
	TotalRequested int `json:"total_requested"`

	// Successful is the number of successful creations
	Successful int `json:"successful"`

	// Failed is the number of failed creations
	Failed int `json:"failed"`

	// Results contains individual results
	Results []*AttestationResult `json:"results"`

	// Errors contains any errors
	Errors []string `json:"errors,omitempty"`
}

// CreateBatchAttestations creates multiple attestations.
func (s *AttestationService) CreateBatchAttestations(ctx context.Context, req *BatchAttestationRequest) (*BatchAttestationResult, error) {
	if s.closed {
		return nil, ErrServiceUnavailable
	}

	result := &BatchAttestationResult{
		TotalRequested: len(req.Requests),
		Results:        make([]*AttestationResult, len(req.Requests)),
		Errors:         make([]string, 0),
	}

	for i, attReq := range req.Requests {
		attResult, err := s.CreateAttestation(ctx, attReq)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("request %d: %v", i, err))
			continue
		}

		result.Successful++
		result.Results[i] = attResult
	}

	return result, nil
}

// ============================================================================
// Attestation for Verified Challenge
// ============================================================================

// CreateAttestationFromChallenge creates an attestation directly from a verified challenge.
// This is a convenience method that calculates scores automatically.
func (s *AttestationService) CreateAttestationFromChallenge(
	ctx context.Context,
	challenge *EmailChallenge,
) (*AttestationResult, error) {
	if challenge.Status != StatusVerified {
		return nil, errors.Wrap(ErrInvalidRequest, "challenge must be verified")
	}

	// Calculate verification score based on challenge properties
	score := s.calculateVerificationScore(challenge)
	confidence := s.calculateConfidenceScore(challenge)

	req := &AttestationRequest{
		Challenge:         challenge,
		VerificationScore: score,
		ConfidenceScore:   confidence,
		IsOrganizational:  challenge.IsOrganizational,
		IsDomainVerified:  false, // Would need domain verification check
		RequestID:         challenge.ChallengeID,
	}

	return s.CreateAttestation(ctx, req)
}

// calculateVerificationScore calculates the verification score for a challenge.
func (s *AttestationService) calculateVerificationScore(challenge *EmailChallenge) uint32 {
	// Base score for successful verification
	score := uint32(80)

	// Bonus for organizational email
	if challenge.IsOrganizational {
		score += 10
	}

	// Penalty for multiple attempts
	if challenge.Attempts > 1 {
		penalty := uint32(challenge.Attempts-1) * 5
		if penalty > 20 {
			penalty = 20
		}
		score -= penalty
	}

	// Ensure score is within bounds
	if score > 100 {
		score = 100
	}

	return score
}

// calculateConfidenceScore calculates the confidence score for a challenge.
func (s *AttestationService) calculateConfidenceScore(challenge *EmailChallenge) uint32 {
	// Base confidence
	confidence := uint32(90)

	// Reduce confidence for resends
	if challenge.ResendCount > 0 {
		reduction := uint32(challenge.ResendCount) * 10
		if reduction > 30 {
			reduction = 30
		}
		confidence -= reduction
	}

	// Reduce confidence for delivery issues
	if challenge.DeliveryStatus == DeliveryBounced {
		confidence -= 20
	}

	return confidence
}

// ============================================================================
// Pending Submission Management
// ============================================================================

// GetPendingSubmissions returns attestations pending chain submission.
func (s *AttestationService) GetPendingSubmissions() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]string, len(s.pendingSubmission))
	copy(result, s.pendingSubmission)
	return result
}

// MarkSubmitted marks an attestation as submitted to chain.
func (s *AttestationService) MarkSubmitted(attestationID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove from pending list
	for i, id := range s.pendingSubmission {
		if id == attestationID {
			s.pendingSubmission = append(s.pendingSubmission[:i], s.pendingSubmission[i+1:]...)
			break
		}
	}
}

// ClearSubmissionQueue clears all pending submissions.
func (s *AttestationService) ClearSubmissionQueue() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pendingSubmission = make([]string, 0)
}

// ============================================================================
// Attestation Export
// ============================================================================

// ExportAttestation exports an attestation as JSON.
func (s *AttestationService) ExportAttestation(ctx context.Context, attestationID string) ([]byte, error) {
	attestation, err := s.GetAttestation(ctx, attestationID)
	if err != nil {
		return nil, err
	}

	return attestation.ToJSON()
}

// ============================================================================
// Health and Lifecycle
// ============================================================================

// HealthCheck returns the health status of the attestation service.
func (s *AttestationService) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	status := &HealthStatus{
		Healthy:   true,
		Status:    "healthy",
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	// Check signer health
	signerHealth, err := s.signer.HealthCheck(ctx)
	if err != nil || !signerHealth.Healthy {
		status.Healthy = false
		status.Status = "signer unhealthy"
		if err != nil {
			status.Details["signer_error"] = err.Error()
		}
	}

	s.mu.RLock()
	status.Details["cached_attestations"] = len(s.attestationCache)
	status.Details["pending_submissions"] = len(s.pendingSubmission)
	s.mu.RUnlock()

	return status, nil
}

// Close closes the attestation service.
func (s *AttestationService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true
	s.logger.Info().Msg("attestation service closed")
	return nil
}

// ============================================================================
// Attestation ID Generation
// ============================================================================

// GenerateAttestationID generates a unique attestation ID.
func GenerateAttestationID(keyFingerprint string) string {
	nonceBytes := make([]byte, 8)
	rand.Read(nonceBytes) //nolint:errcheck
	nonceHex := hex.EncodeToString(nonceBytes)

	fp := keyFingerprint
	if len(fp) > 16 {
		fp = fp[:16]
	}

	return fmt.Sprintf("veid:attestation:%s:%s", fp, nonceHex)
}

// GenerateAttestationIDFromUUID generates an attestation ID using UUID.
func GenerateAttestationIDFromUUID() string {
	return fmt.Sprintf("veid:attestation:%s", uuid.New().String())
}
