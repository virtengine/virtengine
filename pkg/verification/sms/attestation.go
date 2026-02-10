// Package sms provides attestation service for SMS verification.
//
// This file implements signed verification attestations for SMS phone verification,
// creating cryptographically signed proofs that can be stored on-chain.
//
// Task Reference: VE-4C - SMS Verification Delivery + Anti-Fraud
package sms

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/virtengine/virtengine/pkg/verification/signer"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// SMS Attestation Service Interface
// ============================================================================

// SMSAttestationService defines the interface for creating SMS verification attestations
type SMSAttestationService interface {
	// CreateAttestation creates a signed attestation for a verified SMS challenge
	CreateAttestation(ctx context.Context, challenge *SMSChallenge) (*SMSAttestation, error)

	// VerifyAttestation verifies an SMS attestation signature
	VerifyAttestation(ctx context.Context, attestation *SMSAttestation) (bool, error)

	// GetAttestation retrieves an attestation by ID
	GetAttestation(ctx context.Context, attestationID string) (*SMSAttestation, error)

	// RevokeAttestation revokes an existing attestation
	RevokeAttestation(ctx context.Context, attestationID string, reason string) error

	// ListAttestations lists attestations for an account
	ListAttestations(ctx context.Context, accountAddress string) ([]*SMSAttestation, error)

	// Close closes the service
	Close() error
}

// ============================================================================
// SMS Attestation Types
// ============================================================================

// SMSAttestation represents a signed SMS verification attestation
type SMSAttestation struct {
	// Core attestation
	*veidtypes.VerificationAttestation

	// SMS-specific fields
	SMSData SMSAttestationData `json:"sms_data"`
}

// SMSAttestationData contains SMS-specific attestation data
type SMSAttestationData struct {
	// ChallengeID is the ID of the verified challenge
	ChallengeID string `json:"challenge_id"`

	// PhoneHash is the hashed phone number (never plaintext)
	PhoneHash string `json:"phone_hash"`

	// CountryCode is the country code (for regional analytics)
	CountryCode string `json:"country_code"`

	// CarrierType is the detected carrier type
	CarrierType CarrierType `json:"carrier_type"`

	// IsVoIP indicates if VoIP was detected
	IsVoIP bool `json:"is_voip"`

	// VerificationAttempts is how many attempts were made
	VerificationAttempts uint32 `json:"verification_attempts"`

	// VerifiedAt is when verification succeeded
	VerifiedAt time.Time `json:"verified_at"`

	// ValidUntil is when this verification expires
	ValidUntil time.Time `json:"valid_until"`

	// RiskScore is the fraud risk score at verification time
	RiskScore uint32 `json:"risk_score"`

	// Provider is which SMS provider was used
	Provider string `json:"provider,omitempty"`
}

// SMSAttestationConfig contains configuration for the attestation service
type SMSAttestationConfig struct {
	// ValidityDays is how long the attestation remains valid
	ValidityDays int `json:"validity_days"`

	// ScoreThreshold is the minimum score to create attestation
	ScoreThreshold uint32 `json:"score_threshold"`

	// BaseScore is the base score for SMS verification
	BaseScore uint32 `json:"base_score"`

	// MobileBonus is bonus score for mobile (non-VoIP) numbers
	MobileBonus uint32 `json:"mobile_bonus"`

	// SigningEnabled indicates if signing is enabled
	SigningEnabled bool `json:"signing_enabled"`

	// StorageEnabled indicates if attestations should be stored
	StorageEnabled bool `json:"storage_enabled"`

	// MetricsEnabled indicates if metrics should be collected
	MetricsEnabled bool `json:"metrics_enabled"`
}

// DefaultSMSAttestationConfig returns the default configuration
func DefaultSMSAttestationConfig() SMSAttestationConfig {
	return SMSAttestationConfig{
		ValidityDays:   365,
		ScoreThreshold: 50,
		BaseScore:      70,
		MobileBonus:    15,
		SigningEnabled: true,
		StorageEnabled: true,
		MetricsEnabled: true,
	}
}

// ============================================================================
// Default SMS Attestation Service Implementation
// ============================================================================

// DefaultSMSAttestationService implements SMSAttestationService
type DefaultSMSAttestationService struct {
	config  SMSAttestationConfig
	signer  signer.SignerService
	logger  zerolog.Logger
	metrics *Metrics

	// State
	mu           sync.RWMutex
	attestations map[string]*SMSAttestation // In-memory storage
	accountIndex map[string][]string        // account -> attestation IDs
}

// NewSMSAttestationService creates a new SMS attestation service
func NewSMSAttestationService(
	config SMSAttestationConfig,
	signerSvc signer.SignerService,
	logger zerolog.Logger,
) (*DefaultSMSAttestationService, error) {
	service := &DefaultSMSAttestationService{
		config:       config,
		signer:       signerSvc,
		logger:       logger.With().Str("component", "sms_attestation").Logger(),
		metrics:      DefaultMetrics,
		attestations: make(map[string]*SMSAttestation),
		accountIndex: make(map[string][]string),
	}

	return service, nil
}

// CreateAttestation creates a signed attestation for a verified SMS challenge
func (s *DefaultSMSAttestationService) CreateAttestation(ctx context.Context, challenge *SMSChallenge) (*SMSAttestation, error) {
	if challenge == nil {
		return nil, ErrInvalidRequest
	}

	if challenge.Status != StatusVerified {
		return nil, fmt.Errorf("challenge not verified: status=%s", challenge.Status)
	}

	if challenge.VerifiedAt == nil {
		return nil, fmt.Errorf("challenge has no verification timestamp")
	}

	now := time.Now()

	// Generate nonce for attestation
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Calculate score
	score := s.calculateScore(challenge)
	if score < s.config.ScoreThreshold {
		s.logger.Warn().
			Str("challenge_id", challenge.ChallengeID).
			Uint32("score", score).
			Uint32("threshold", s.config.ScoreThreshold).
			Msg("score below threshold, attestation may have limited value")
	}

	// Calculate validity
	validityDuration := time.Duration(s.config.ValidityDays) * 24 * time.Hour
	validUntil := now.Add(validityDuration)

	// Create issuer (will be updated by signer)
	issuer := veidtypes.NewAttestationIssuer("pending", "")

	// Create subject
	subject := veidtypes.NewAttestationSubject(challenge.AccountAddress)
	subject.ScopeID = challenge.ChallengeID
	subject.RequestID = challenge.ChallengeID

	// Create base attestation
	baseAttestation := veidtypes.NewVerificationAttestation(
		issuer,
		subject,
		veidtypes.AttestationTypeSMSVerification,
		nonce,
		now,
		validityDuration,
		score,
		calculateConfidence(challenge),
	)

	// Add verification proofs
	proofHash := createChallengeProofHash(challenge)
	verificationProof := veidtypes.NewVerificationProofDetail(
		"sms_otp_verification",
		proofHash,
		score,
		s.config.ScoreThreshold,
		*challenge.VerifiedAt,
	)
	baseAttestation.AddVerificationProof(verificationProof)

	// Add carrier verification proof if available
	if challenge.CarrierType != "" {
		carrierProof := veidtypes.NewVerificationProofDetail(
			"carrier_verification",
			HashPhoneNumber(string(challenge.CarrierType)+challenge.PhoneHash),
			carrierTypeScore(challenge.CarrierType),
			0, // No threshold for carrier verification
			now,
		)
		baseAttestation.AddVerificationProof(carrierProof)
	}

	// Add VoIP check proof
	voipProof := veidtypes.NewVerificationProofDetail(
		"voip_check",
		HashPhoneNumber(fmt.Sprintf("%t:%s", challenge.IsVoIP, challenge.PhoneHash)),
		voipCheckScore(challenge.IsVoIP),
		50, // VoIP check threshold
		now,
	)
	baseAttestation.AddVerificationProof(voipProof)

	// Set metadata
	baseAttestation.SetMetadata("country_code", challenge.CountryCode)
	baseAttestation.SetMetadata("carrier_type", string(challenge.CarrierType))
	baseAttestation.SetMetadata("is_voip", fmt.Sprintf("%t", challenge.IsVoIP))
	baseAttestation.SetMetadata("risk_score", fmt.Sprintf("%d", challenge.RiskScore))
	if challenge.Provider != "" {
		baseAttestation.SetMetadata("provider", challenge.Provider)
	}

	// Create SMS attestation
	smsAttestation := &SMSAttestation{
		VerificationAttestation: baseAttestation,
		SMSData: SMSAttestationData{
			ChallengeID:          challenge.ChallengeID,
			PhoneHash:            challenge.PhoneHash,
			CountryCode:          challenge.CountryCode,
			CarrierType:          challenge.CarrierType,
			IsVoIP:               challenge.IsVoIP,
			VerificationAttempts: challenge.Attempts,
			VerifiedAt:           *challenge.VerifiedAt,
			ValidUntil:           validUntil,
			RiskScore:            challenge.RiskScore,
			Provider:             challenge.Provider,
		},
	}

	// Sign the attestation if signing is enabled
	if s.config.SigningEnabled && s.signer != nil {
		if err := s.signer.SignAttestation(ctx, baseAttestation); err != nil {
			s.logger.Error().Err(err).Str("challenge_id", challenge.ChallengeID).Msg("failed to sign attestation")
			// Continue without signature - can be signed later
		}
	}

	// Store the attestation
	if s.config.StorageEnabled {
		s.storeAttestation(smsAttestation)
	}

	// Record metrics
	if s.config.MetricsEnabled && s.metrics != nil {
		s.metrics.RecordAttestationCreated(challenge.CountryCode)
	}

	s.logger.Info().
		Str("attestation_id", smsAttestation.ID).
		Str("challenge_id", challenge.ChallengeID).
		Str("account", challenge.AccountAddress).
		Uint32("score", score).
		Time("valid_until", validUntil).
		Msg("SMS attestation created")

	return smsAttestation, nil
}

// calculateScore calculates the attestation score based on challenge data
func (s *DefaultSMSAttestationService) calculateScore(challenge *SMSChallenge) uint32 {
	score := s.config.BaseScore

	// Bonus for non-VoIP mobile numbers
	if !challenge.IsVoIP && challenge.CarrierType == CarrierTypeMobile {
		score += s.config.MobileBonus
	}

	// Penalty for VoIP
	if challenge.IsVoIP {
		if score > 30 {
			score -= 30
		} else {
			score = 0
		}
	}

	// Penalty for high risk score
	if challenge.RiskScore > 50 {
		penalty := (challenge.RiskScore - 50) / 2
		if score > penalty {
			score -= penalty
		} else {
			score = 0
		}
	}

	// Penalty for multiple attempts
	if challenge.Attempts > 1 {
		penalty := uint32((challenge.Attempts - 1) * 5)
		if score > penalty {
			score -= penalty
		}
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

// calculateConfidence calculates confidence score based on verification quality
func calculateConfidence(challenge *SMSChallenge) uint32 {
	confidence := uint32(80) // Base confidence

	// Higher confidence for mobile carriers
	if challenge.CarrierType == CarrierTypeMobile && !challenge.IsVoIP {
		confidence += 15
	}

	// Lower confidence for VoIP
	if challenge.IsVoIP {
		confidence -= 25
	}

	// Lower confidence for unknown carriers
	if challenge.CarrierType == CarrierTypeUnknown {
		confidence -= 10
	}

	// Adjust for risk score
	if challenge.RiskScore > 0 {
		confidence -= challenge.RiskScore / 5
	}

	// Cap at 0-100
	if confidence > 100 {
		confidence = 100
	}

	return confidence
}

// createChallengeProofHash creates a hash of challenge verification data
func createChallengeProofHash(challenge *SMSChallenge) string {
	data := fmt.Sprintf("%s:%s:%s:%d:%s",
		challenge.ChallengeID,
		challenge.PhoneHash,
		challenge.AccountAddress,
		challenge.VerifiedAt.Unix(),
		challenge.Nonce,
	)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// carrierTypeScore returns a score contribution for carrier type
func carrierTypeScore(carrierType CarrierType) uint32 {
	switch carrierType {
	case CarrierTypeMobile:
		return 100
	case CarrierTypeLandline:
		return 80
	case CarrierTypeVoIP:
		return 20
	default:
		return 50
	}
}

// voipCheckScore returns a score for VoIP check result
func voipCheckScore(isVoIP bool) uint32 {
	if isVoIP {
		return 0
	}
	return 100
}

// storeAttestation stores an attestation in memory
func (s *DefaultSMSAttestationService) storeAttestation(attestation *SMSAttestation) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.attestations[attestation.ID] = attestation

	// Update account index
	account := attestation.Subject.AccountAddress
	s.accountIndex[account] = append(s.accountIndex[account], attestation.ID)
}

// VerifyAttestation verifies an SMS attestation signature
func (s *DefaultSMSAttestationService) VerifyAttestation(ctx context.Context, attestation *SMSAttestation) (bool, error) {
	if attestation == nil || attestation.VerificationAttestation == nil {
		return false, ErrInvalidRequest
	}

	// Check expiration
	if attestation.IsExpired(time.Now()) {
		return false, nil
	}

	// Verify signature if signer is available
	if s.signer != nil {
		return s.signer.VerifyAttestation(ctx, attestation.VerificationAttestation)
	}

	// No signer available - validate structure only
	if err := attestation.Validate(); err != nil {
		return false, err
	}

	return true, nil
}

// GetAttestation retrieves an attestation by ID
func (s *DefaultSMSAttestationService) GetAttestation(ctx context.Context, attestationID string) (*SMSAttestation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	attestation, ok := s.attestations[attestationID]
	if !ok {
		return nil, fmt.Errorf("attestation not found: %s", attestationID)
	}

	return attestation, nil
}

// RevokeAttestation revokes an existing attestation
func (s *DefaultSMSAttestationService) RevokeAttestation(ctx context.Context, attestationID string, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	attestation, ok := s.attestations[attestationID]
	if !ok {
		return fmt.Errorf("attestation not found: %s", attestationID)
	}

	// Mark as expired by setting expiry to past
	attestation.ExpiresAt = time.Now().Add(-time.Hour)
	attestation.SetMetadata("revoked", "true")
	attestation.SetMetadata("revocation_reason", reason)
	attestation.SetMetadata("revoked_at", time.Now().UTC().Format(time.RFC3339))

	s.logger.Info().
		Str("attestation_id", attestationID).
		Str("reason", reason).
		Msg("attestation revoked")

	return nil
}

// ListAttestations lists attestations for an account
func (s *DefaultSMSAttestationService) ListAttestations(ctx context.Context, accountAddress string) ([]*SMSAttestation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids, ok := s.accountIndex[accountAddress]
	if !ok {
		return []*SMSAttestation{}, nil
	}

	result := make([]*SMSAttestation, 0, len(ids))
	for _, id := range ids {
		if attestation, ok := s.attestations[id]; ok {
			result = append(result, attestation)
		}
	}

	return result, nil
}

// Close closes the service
func (s *DefaultSMSAttestationService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear in-memory storage
	s.attestations = make(map[string]*SMSAttestation)
	s.accountIndex = make(map[string][]string)

	return nil
}

// Ensure DefaultSMSAttestationService implements SMSAttestationService
var _ SMSAttestationService = (*DefaultSMSAttestationService)(nil)

// ============================================================================
// Attestation Verification Helpers
// ============================================================================

// ValidateChallenge validates that a challenge can have an attestation created
func ValidateChallenge(challenge *SMSChallenge) error {
	if challenge == nil {
		return ErrInvalidRequest
	}

	if challenge.ChallengeID == "" {
		return fmt.Errorf("challenge ID is required")
	}

	if challenge.AccountAddress == "" {
		return fmt.Errorf("account address is required")
	}

	if challenge.PhoneHash == "" {
		return fmt.Errorf("phone hash is required")
	}

	if challenge.Status != StatusVerified {
		return fmt.Errorf("challenge not verified: %s", challenge.Status)
	}

	if challenge.VerifiedAt == nil {
		return fmt.Errorf("verification timestamp is required")
	}

	return nil
}

// IsAttestationValid checks if an attestation is currently valid
func IsAttestationValid(attestation *SMSAttestation, now time.Time) bool {
	if attestation == nil || attestation.VerificationAttestation == nil {
		return false
	}

	// Check if expired
	if attestation.IsExpired(now) {
		return false
	}

	// Check if revoked
	if revoked, ok := attestation.Metadata["revoked"]; ok && revoked == "true" {
		return false
	}

	// Check SMS-specific validity
	if now.After(attestation.SMSData.ValidUntil) {
		return false
	}

	return true
}

// AttestationScore returns the effective score considering age and validity
func AttestationScore(attestation *SMSAttestation, now time.Time) uint32 {
	if !IsAttestationValid(attestation, now) {
		return 0
	}

	score := attestation.Score

	// Apply age decay if attestation is old
	age := now.Sub(attestation.SMSData.VerifiedAt)
	ageYears := age.Hours() / (24 * 365)

	// 5% decay per year after first year
	if ageYears > 1 {
		decayPercent := uint32((ageYears - 1) * 5)
		if decayPercent > 25 {
			decayPercent = 25 // Max 25% decay
		}
		score = score * (100 - decayPercent) / 100
	}

	return score
}
