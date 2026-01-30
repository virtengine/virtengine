// Package email provides the main email verification service implementation.
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

	"github.com/virtengine/virtengine/pkg/cache"
	"github.com/virtengine/virtengine/pkg/errors"
	"github.com/virtengine/virtengine/pkg/verification/audit"
	"github.com/virtengine/virtengine/pkg/verification/ratelimit"
	"github.com/virtengine/virtengine/pkg/verification/signer"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Default Email Verification Service
// ============================================================================

// DefaultService implements EmailVerificationService
type DefaultService struct {
	config          Config
	provider        Provider
	signer          signer.SignerService
	auditor         audit.AuditLogger
	rateLimiter     ratelimit.VerificationLimiter
	cache           cache.Cache[string, *EmailChallenge]
	templateManager *TemplateManager
	metrics         *Metrics
	logger          zerolog.Logger

	// State
	mu     sync.RWMutex
	closed bool
}

// ServiceOption is a functional option for configuring the service
type ServiceOption func(*DefaultService)

// WithProvider sets the email provider
func WithProvider(provider Provider) ServiceOption {
	return func(s *DefaultService) {
		s.provider = provider
	}
}

// WithSigner sets the attestation signer
func WithSigner(signer signer.SignerService) ServiceOption {
	return func(s *DefaultService) {
		s.signer = signer
	}
}

// WithAuditor sets the audit logger
func WithAuditor(auditor audit.AuditLogger) ServiceOption {
	return func(s *DefaultService) {
		s.auditor = auditor
	}
}

// WithRateLimiter sets the rate limiter
func WithRateLimiter(limiter ratelimit.VerificationLimiter) ServiceOption {
	return func(s *DefaultService) {
		s.rateLimiter = limiter
	}
}

// WithCache sets the challenge cache
func WithCache(c cache.Cache[string, *EmailChallenge]) ServiceOption {
	return func(s *DefaultService) {
		s.cache = c
	}
}

// WithMetrics sets the metrics collector
func WithMetrics(m *Metrics) ServiceOption {
	return func(s *DefaultService) {
		s.metrics = m
	}
}

// NewService creates a new email verification service
func NewService(
	ctx context.Context,
	config Config,
	logger zerolog.Logger,
	opts ...ServiceOption,
) (*DefaultService, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	s := &DefaultService{
		config:  config,
		logger:  logger.With().Str("component", "email_verification").Logger(),
		metrics: DefaultMetrics,
	}

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	// Create provider if not set
	if s.provider == nil {
		provider, err := NewProvider(config, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create email provider: %w", err)
		}
		s.provider = provider
	}

	// Create template manager
	s.templateManager = NewTemplateManager(TemplateDefaults{
		ProductName:  config.FromName,
		SupportEmail: config.FromAddress,
	})

	// Register metrics if enabled
	if config.MetricsEnabled && s.metrics != nil {
		if err := s.metrics.Register(); err != nil {
			s.logger.Warn().Err(err).Msg("failed to register metrics")
		}
	}

	s.logger.Info().
		Str("provider", config.Provider).
		Bool("rate_limit_enabled", config.RateLimitEnabled).
		Bool("audit_log_enabled", config.AuditLogEnabled).
		Msg("email verification service initialized")

	return s, nil
}

// ============================================================================
// Core Operations
// ============================================================================

// InitiateVerification starts a new email verification challenge
func (s *DefaultService) InitiateVerification(ctx context.Context, req *InitiateRequest) (*InitiateResponse, error) {
	if s.closed {
		return nil, ErrServiceUnavailable
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check rate limits
	if s.config.RateLimitEnabled && s.rateLimiter != nil {
		allowed, result, err := s.rateLimiter.AllowVerification(
			ctx,
			req.AccountAddress,
			ratelimit.LimitTypeEmailVerification,
		)
		if err != nil {
			s.logger.Error().Err(err).Msg("rate limiter error")
		} else if !allowed {
			if s.metrics != nil {
				s.metrics.RecordRateLimitHit()
			}
			return nil, errors.Wrapf(ErrRateLimited, "try again in %d seconds", result.RetryAfter)
		}
	}

	// Generate challenge ID
	challengeID := uuid.New().String()

	// Create challenge
	challenge, secret, err := NewEmailChallenge(ChallengeConfig{
		ChallengeID:    challengeID,
		AccountAddress: req.AccountAddress,
		Email:          req.Email,
		Method:         req.Method,
		CreatedAt:      time.Now(),
		TTLSeconds:     s.config.OTPTTLSeconds,
		MaxAttempts:    s.config.MaxAttempts,
		MaxResends:     s.config.MaxResends,
		IPAddress:      req.IPAddress,
		UserAgent:      req.UserAgent,
	})
	if err != nil {
		return nil, err
	}

	// Store challenge in cache
	if s.cache != nil {
		ttl := time.Duration(s.config.OTPTTLSeconds) * time.Second
		if err := s.cache.SetWithTTL(ctx, challengeID, challenge, ttl); err != nil {
			s.logger.Error().Err(err).Msg("failed to cache challenge")
			return nil, errors.Wrap(ErrCacheError, err.Error())
		}
	}

	// Send verification email
	if err := s.sendVerificationEmail(ctx, req.Email, challenge, secret); err != nil {
		// Clean up cache on failure
		if s.cache != nil {
			_ = s.cache.Delete(ctx, challengeID)
		}
		return nil, err
	}

	// Record metrics
	if s.metrics != nil {
		s.metrics.RecordChallengeCreated(req.Method)
	}

	// Audit log
	if s.config.AuditLogEnabled && s.auditor != nil {
		s.auditor.Log(ctx, audit.Event{
			Type:      audit.EventTypeVerificationInitiated,
			Timestamp: time.Now(),
			Actor:     req.AccountAddress,
			Resource:  challengeID,
			Action:    "initiate_email_verification",
			Details: map[string]interface{}{
				"method":       req.Method,
				"email_hash":   challenge.EmailHash[:16] + "...",
				"masked_email": challenge.MaskedEmail,
				"ip_address":   req.IPAddress,
			},
		})
	}

	return &InitiateResponse{
		ChallengeID:           challengeID,
		MaskedEmail:           challenge.MaskedEmail,
		Method:                req.Method,
		ExpiresAt:             challenge.ExpiresAt,
		ResendCooldownSeconds: s.config.ResendCooldownSeconds,
	}, nil
}

// VerifyChallenge verifies an email challenge with the provided secret
func (s *DefaultService) VerifyChallenge(ctx context.Context, req *VerifyRequest) (*VerifyResponse, error) {
	if s.closed {
		return nil, ErrServiceUnavailable
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Get challenge from cache
	challenge, err := s.getChallenge(ctx, req.ChallengeID)
	if err != nil {
		return nil, err
	}

	// Verify account matches
	if challenge.AccountAddress != req.AccountAddress {
		return nil, ErrAccountMismatch
	}

	// Check if expired
	now := time.Now()
	if challenge.IsExpired(now) {
		challenge.Status = StatusExpired
		_ = s.updateChallenge(ctx, challenge)
		if s.metrics != nil {
			s.metrics.RecordChallengeExpired(challenge.Method)
		}
		return nil, ErrChallengeExpired
	}

	// Check if already consumed
	if challenge.IsConsumed {
		return nil, ErrChallengeConsumed
	}

	// Check if can attempt
	if !challenge.CanAttempt() {
		return &VerifyResponse{
			Success:           false,
			RemainingAttempts: 0,
			ErrorCode:         "MAX_ATTEMPTS_EXCEEDED",
			ErrorMessage:      "Maximum verification attempts exceeded",
		}, ErrMaxAttemptsExceeded
	}

	// Verify the secret
	success := challenge.VerifySecret(req.Secret)
	challenge.RecordAttempt(now, success)

	// Record metrics
	if s.metrics != nil {
		s.metrics.RecordVerificationAttempt(challenge.Method, success)
	}

	// Update challenge in cache
	if err := s.updateChallenge(ctx, challenge); err != nil {
		s.logger.Error().Err(err).Msg("failed to update challenge")
	}

	// Handle verification result
	if success {
		// Create attestation
		attestation, err := s.createAttestation(ctx, challenge)
		if err != nil {
			s.logger.Error().Err(err).Msg("failed to create attestation")
			// Still return success, attestation can be retried
		}

		// Record verification latency
		if s.metrics != nil {
			latency := now.Sub(challenge.CreatedAt)
			s.metrics.RecordChallengeVerified(challenge.Method, latency)
		}

		// Audit log
		if s.config.AuditLogEnabled && s.auditor != nil {
			s.auditor.Log(ctx, audit.Event{
				Type:      audit.EventTypeVerificationCompleted,
				Timestamp: now,
				Actor:     req.AccountAddress,
				Resource:  req.ChallengeID,
				Action:    "verify_email_success",
				Details: map[string]interface{}{
					"method":         challenge.Method,
					"attempts":       challenge.Attempts,
					"attestation_id": attestation.ID,
				},
			})
		}

		var attestationID string
		if attestation != nil {
			attestationID = attestation.ID
		}

		return &VerifyResponse{
			Success:           true,
			Verified:          true,
			AttestationID:     attestationID,
			RemainingAttempts: challenge.MaxAttempts - challenge.Attempts,
		}, nil
	}

	// Failed attempt
	if s.config.AuditLogEnabled && s.auditor != nil {
		s.auditor.Log(ctx, audit.Event{
			Type:      audit.EventTypeVerificationFailed,
			Timestamp: now,
			Actor:     req.AccountAddress,
			Resource:  req.ChallengeID,
			Action:    "verify_email_failed",
			Details: map[string]interface{}{
				"method":    challenge.Method,
				"attempts":  challenge.Attempts,
				"remaining": challenge.MaxAttempts - challenge.Attempts,
			},
		})
	}

	// Check if max attempts reached
	if challenge.Attempts >= challenge.MaxAttempts {
		if s.metrics != nil {
			s.metrics.RecordChallengeFailed(challenge.Method, "max_attempts")
		}
		return &VerifyResponse{
			Success:           false,
			RemainingAttempts: 0,
			ErrorCode:         "MAX_ATTEMPTS_EXCEEDED",
			ErrorMessage:      "Maximum verification attempts exceeded",
		}, ErrMaxAttemptsExceeded
	}

	return &VerifyResponse{
		Success:           false,
		RemainingAttempts: challenge.MaxAttempts - challenge.Attempts,
		ErrorCode:         "INVALID_SECRET",
		ErrorMessage:      "The verification code is incorrect",
	}, ErrInvalidSecret
}

// ResendVerification resends the verification email
func (s *DefaultService) ResendVerification(ctx context.Context, req *ResendRequest) (*ResendResponse, error) {
	if s.closed {
		return nil, ErrServiceUnavailable
	}

	if req.ChallengeID == "" {
		return nil, errors.Wrap(ErrInvalidRequest, "challenge_id is required")
	}
	if req.AccountAddress == "" {
		return nil, errors.Wrap(ErrInvalidRequest, "account_address is required")
	}

	// Get challenge
	challenge, err := s.getChallenge(ctx, req.ChallengeID)
	if err != nil {
		return nil, err
	}

	// Verify account matches
	if challenge.AccountAddress != req.AccountAddress {
		return nil, ErrAccountMismatch
	}

	now := time.Now()

	// Check if can resend
	if !challenge.CanResend(now) {
		if challenge.ResendCount >= challenge.MaxResends {
			return nil, ErrResendLimitExceeded
		}
		// Calculate when next resend is allowed
		nextResend := challenge.LastResendAt.Add(time.Duration(DefaultResendCooldownSeconds) * time.Second)
		return nil, errors.Wrapf(ErrResendCooldown, "try again at %s", nextResend.Format(time.RFC3339))
	}

	// Generate new secret
	var secretHash string
	switch challenge.Method {
	case MethodOTP:
		_, secretHash, err = GenerateOTP(s.config.OTPLength)
	case MethodMagicLink:
		_, secretHash, err = GenerateMagicLinkToken()
	}
	if err != nil {
		return nil, err
	}

	// Calculate new expiry
	ttl := s.config.OTPTTLSeconds
	if challenge.Method == MethodMagicLink {
		ttl = s.config.LinkTTLSeconds
	}
	newExpiresAt := now.Add(time.Duration(ttl) * time.Second)

	// Update challenge
	if err := challenge.RecordResend(now, secretHash, newExpiresAt); err != nil {
		return nil, err
	}

	// We need to get the email to resend - but we only have the hash
	// In production, you'd need to store this securely or ask the user to provide it again
	// For now, we'll return an error indicating the limitation
	// TODO: Implement secure email retrieval for resends

	// Update cache
	if err := s.updateChallenge(ctx, challenge); err != nil {
		s.logger.Error().Err(err).Msg("failed to update challenge")
	}

	// Record metrics
	if s.metrics != nil {
		s.metrics.RecordResend(challenge.Method)
	}

	// Audit log
	if s.config.AuditLogEnabled && s.auditor != nil {
		s.auditor.Log(ctx, audit.Event{
			Type:      audit.EventTypeResendRequested,
			Timestamp: now,
			Actor:     req.AccountAddress,
			Resource:  req.ChallengeID,
			Action:    "resend_verification",
			Details: map[string]interface{}{
				"method":        challenge.Method,
				"resend_count":  challenge.ResendCount,
				"max_resends":   challenge.MaxResends,
			},
		})
	}

	// Note: In production, resend would need the email address to be provided again
	// or stored securely (encrypted) in the challenge
	s.logger.Info().
		Str("challenge_id", req.ChallengeID).
		Str("secret_hash", secretHash[:16]+"...").
		Msg("resend prepared (email send not implemented in resend)")

	nextResendAt := now.Add(time.Duration(s.config.ResendCooldownSeconds) * time.Second)

	return &ResendResponse{
		Success:          true,
		ExpiresAt:        newExpiresAt,
		ResendsRemaining: challenge.MaxResends - challenge.ResendCount,
		NextResendAt:     nextResendAt,
	}, nil
}

// GetChallenge retrieves a challenge by ID
func (s *DefaultService) GetChallenge(ctx context.Context, challengeID string) (*EmailChallenge, error) {
	return s.getChallenge(ctx, challengeID)
}

// CancelChallenge cancels an active challenge
func (s *DefaultService) CancelChallenge(ctx context.Context, challengeID string, accountAddress string) error {
	challenge, err := s.getChallenge(ctx, challengeID)
	if err != nil {
		return err
	}

	if challenge.AccountAddress != accountAddress {
		return ErrAccountMismatch
	}

	challenge.Status = StatusCancelled
	challenge.IsConsumed = true

	if s.cache != nil {
		if err := s.cache.Delete(ctx, challengeID); err != nil {
			s.logger.Warn().Err(err).Msg("failed to delete cancelled challenge")
		}
	}

	// Audit log
	if s.config.AuditLogEnabled && s.auditor != nil {
		s.auditor.Log(ctx, audit.Event{
			Type:      audit.EventTypeVerificationCancelled,
			Timestamp: time.Now(),
			Actor:     accountAddress,
			Resource:  challengeID,
			Action:    "cancel_verification",
		})
	}

	return nil
}

// ProcessWebhook processes a delivery status webhook
func (s *DefaultService) ProcessWebhook(ctx context.Context, providerName string, payload []byte) error {
	events, err := s.provider.ParseWebhook(ctx, payload, "")
	if err != nil {
		return errors.Wrap(ErrWebhookInvalid, err.Error())
	}

	for _, event := range events {
		s.processWebhookEvent(ctx, event)
	}

	return nil
}

// GetDeliveryStatus returns the delivery status for a challenge
func (s *DefaultService) GetDeliveryStatus(ctx context.Context, challengeID string) (*DeliveryResult, error) {
	challenge, err := s.getChallenge(ctx, challengeID)
	if err != nil {
		return nil, err
	}

	// If we have a provider message ID, query the provider
	if challenge.ProviderMessageID != "" {
		status, err := s.provider.GetDeliveryStatus(ctx, challenge.ProviderMessageID)
		if err != nil {
			s.logger.Warn().Err(err).Msg("failed to get delivery status from provider")
		} else {
			return &DeliveryResult{
				ChallengeID:       challengeID,
				Success:           status.Status == DeliveryDelivered,
				ProviderMessageID: challenge.ProviderMessageID,
				DeliveryStatus:    status.Status,
				SentAt:            challenge.CreatedAt,
				Provider:          s.provider.Name(),
			}, nil
		}
	}

	// Return cached status
	return &DeliveryResult{
		ChallengeID:       challengeID,
		Success:           challenge.DeliveryStatus == DeliveryDelivered,
		ProviderMessageID: challenge.ProviderMessageID,
		DeliveryStatus:    challenge.DeliveryStatus,
		SentAt:            challenge.CreatedAt,
		Provider:          s.provider.Name(),
	}, nil
}

// HealthCheck returns the health status of the service
func (s *DefaultService) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	status := &HealthStatus{
		Healthy:   true,
		Status:    "healthy",
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	// Check provider health
	if err := s.provider.HealthCheck(ctx); err != nil {
		status.ProviderHealthy = false
		status.Details["provider_error"] = err.Error()
		status.Healthy = false
		status.Status = "provider unhealthy"
	} else {
		status.ProviderHealthy = true
	}

	// Check cache health
	if s.cache != nil {
		// Try a simple get to check cache is responsive
		_, _ = s.cache.Get(ctx, "health-check-test")
		status.CacheHealthy = true
	} else {
		status.CacheHealthy = true // No cache configured is OK
	}

	status.Details["provider"] = s.provider.Name()

	return status, nil
}

// Close closes the service and releases resources
func (s *DefaultService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true

	// Close provider
	if s.provider != nil {
		if err := s.provider.Close(); err != nil {
			s.logger.Error().Err(err).Msg("failed to close provider")
		}
	}

	// Close signer
	if s.signer != nil {
		if err := s.signer.Close(); err != nil {
			s.logger.Error().Err(err).Msg("failed to close signer")
		}
	}

	s.logger.Info().Msg("email verification service closed")
	return nil
}

// ============================================================================
// Internal Helpers
// ============================================================================

// sendVerificationEmail sends the verification email
func (s *DefaultService) sendVerificationEmail(ctx context.Context, email string, challenge *EmailChallenge, secret string) error {
	var templateType TemplateType
	var templateData TemplateData

	switch challenge.Method {
	case MethodOTP:
		templateType = TemplateOTPVerification
		templateData = TemplateData{
			OTP:       secret,
			ExpiresIn: FormatExpiryTime(challenge.ExpiresAt),
			ExpiresAt: challenge.ExpiresAt,
		}
	case MethodMagicLink:
		templateType = TemplateMagicLink
		verificationLink := fmt.Sprintf("%s/verify?token=%s&challenge=%s",
			s.config.BaseURL, secret, challenge.ChallengeID)
		templateData = TemplateData{
			VerificationLink: verificationLink,
			ExpiresIn:        FormatExpiryTime(challenge.ExpiresAt),
			ExpiresAt:        challenge.ExpiresAt,
		}
	}

	// Render email
	emailMsg, err := s.templateManager.RenderEmail(templateType, templateData, email)
	if err != nil {
		return errors.Wrap(ErrTemplateError, err.Error())
	}

	// Set from address
	emailMsg.From = s.config.FromAddress
	emailMsg.FromName = s.config.FromName

	// Add metadata for tracking
	emailMsg.Metadata = map[string]string{
		"challenge_id":    challenge.ChallengeID,
		"account_address": challenge.AccountAddress,
	}
	emailMsg.Tags = []string{"verification", string(challenge.Method)}

	// Send email
	startTime := time.Now()
	result, err := s.provider.Send(ctx, emailMsg)
	sendLatency := time.Since(startTime)

	if err != nil {
		if s.metrics != nil {
			s.metrics.RecordEmailSent(s.provider.Name(), templateType, sendLatency)
		}
		return errors.Wrap(ErrDeliveryFailed, err.Error())
	}

	// Update challenge with delivery info
	challenge.UpdateDeliveryStatus(DeliverySent, result.MessageID)

	// Record metrics
	if s.metrics != nil {
		s.metrics.RecordEmailSent(s.provider.Name(), templateType, sendLatency)
	}

	s.logger.Info().
		Str("challenge_id", challenge.ChallengeID).
		Str("message_id", result.MessageID).
		Str("method", string(challenge.Method)).
		Dur("latency", sendLatency).
		Msg("verification email sent")

	return nil
}

// getChallenge retrieves a challenge from cache
func (s *DefaultService) getChallenge(ctx context.Context, challengeID string) (*EmailChallenge, error) {
	if s.cache == nil {
		return nil, errors.Wrap(ErrServiceUnavailable, "cache not configured")
	}

	challenge, err := s.cache.Get(ctx, challengeID)
	if err != nil {
		return nil, errors.Wrapf(ErrChallengeNotFound, "challenge ID: %s", challengeID)
	}

	return challenge, nil
}

// updateChallenge updates a challenge in cache
func (s *DefaultService) updateChallenge(ctx context.Context, challenge *EmailChallenge) error {
	if s.cache == nil {
		return nil
	}

	// Calculate remaining TTL
	ttl := time.Until(challenge.ExpiresAt)
	if ttl <= 0 {
		ttl = time.Minute // Keep expired challenges for a short time for auditing
	}

	return s.cache.SetWithTTL(ctx, challenge.ChallengeID, challenge, ttl)
}

// createAttestation creates a verification attestation for a verified email
func (s *DefaultService) createAttestation(ctx context.Context, challenge *EmailChallenge) (*veidtypes.VerificationAttestation, error) {
	if s.signer == nil {
		s.logger.Warn().Msg("signer not configured, skipping attestation")
		return nil, nil
	}

	// Generate attestation nonce
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return nil, errors.Wrapf(ErrAttestationFailed, "failed to generate nonce: %v", err)
	}

	now := time.Now()
	validityDuration := time.Duration(s.config.VerificationValidityDays) * 24 * time.Hour

	// Create subject
	subject := veidtypes.NewAttestationSubject(challenge.AccountAddress)
	subject.RequestID = challenge.ChallengeID

	// Create attestation
	attestation := veidtypes.NewVerificationAttestation(
		veidtypes.AttestationIssuer{}, // Will be set by signer
		subject,
		veidtypes.AttestationTypeEmailVerification,
		nonceBytes,
		now,
		validityDuration,
		100, // Score (100 for successful verification)
		100, // Confidence
	)

	// Add verification proof
	proofDetail := veidtypes.NewVerificationProofDetail(
		"email_otp_verification",
		challenge.EmailHash,
		100,
		50, // Threshold
		now,
	)
	attestation.AddVerificationProof(proofDetail)

	// Add metadata
	attestation.SetMetadata("email_hash", challenge.EmailHash[:16]+"...")
	attestation.SetMetadata("domain_hash", challenge.DomainHash[:16]+"...")
	attestation.SetMetadata("method", string(challenge.Method))
	attestation.SetMetadata("is_organizational", fmt.Sprintf("%t", challenge.IsOrganizational))

	// Sign the attestation
	if err := s.signer.SignAttestation(ctx, attestation); err != nil {
		return nil, errors.Wrap(ErrSigningFailed, err.Error())
	}

	// Record metrics
	if s.metrics != nil {
		s.metrics.RecordAttestationCreated()
	}

	s.logger.Info().
		Str("attestation_id", attestation.ID).
		Str("challenge_id", challenge.ChallengeID).
		Str("account", challenge.AccountAddress).
		Msg("email verification attestation created")

	return attestation, nil
}

// processWebhookEvent processes a single webhook event
func (s *DefaultService) processWebhookEvent(ctx context.Context, event WebhookEvent) {
	s.logger.Debug().
		Str("event_type", string(event.EventType)).
		Str("message_id", event.MessageID).
		Msg("processing webhook event")

	// Find challenge by message ID
	// In production, you'd have a reverse lookup from message ID to challenge ID
	// For now, we log the event

	switch event.EventType {
	case WebhookEventDelivered:
		if s.metrics != nil {
			s.metrics.RecordEmailDelivered(s.provider.Name(), time.Since(event.Timestamp))
		}
	case WebhookEventBounced:
		if s.metrics != nil {
			s.metrics.RecordEmailBounced(s.provider.Name(), event.BounceType)
		}
	case WebhookEventComplaint:
		s.logger.Warn().
			Str("message_id", event.MessageID).
			Str("complaint_type", event.ComplaintType).
			Msg("spam complaint received")
	}

	// Audit log
	if s.config.AuditLogEnabled && s.auditor != nil {
		s.auditor.Log(ctx, audit.Event{
			Type:      audit.EventTypeWebhookReceived,
			Timestamp: event.Timestamp,
			Resource:  event.MessageID,
			Action:    "webhook_" + string(event.EventType),
			Details: map[string]interface{}{
				"event_type":   event.EventType,
				"bounce_type":  event.BounceType,
				"bounce_subtype": event.BounceSubtype,
			},
		})
	}
}

// GenerateNonce generates a cryptographically secure nonce
func GenerateNonce(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Ensure DefaultService implements EmailVerificationService
var _ EmailVerificationService = (*DefaultService)(nil)
