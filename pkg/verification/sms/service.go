// Package sms provides the main SMS verification service implementation.
//
// Task Reference: VE-4C - SMS Verification Delivery + Anti-Fraud
package sms

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

// Note: SMSVerificationService interface and HealthStatus are defined in types.go

// ============================================================================
// Default SMS Verification Service
// ============================================================================

// DefaultService implements SMSVerificationService
type DefaultService struct {
	config          Config
	provider        Provider
	signer          signer.SignerService
	auditor         audit.AuditLogger
	rateLimiter     ratelimit.VerificationLimiter
	antiFraud       AntiFraudEngine
	cache           cache.Cache[string, *SMSChallenge]
	templateManager *TemplateManager
	regionLimits    *RegionRateLimits
	metrics         *Metrics
	logger          zerolog.Logger

	// State
	mu     sync.RWMutex
	closed bool
}

// ServiceOption is a functional option for configuring the service
type ServiceOption func(*DefaultService)

// WithProvider sets the SMS provider
func WithProvider(provider Provider) ServiceOption {
	return func(s *DefaultService) {
		s.provider = provider
	}
}

// WithSigner sets the attestation signer
func WithSigner(signerSvc signer.SignerService) ServiceOption {
	return func(s *DefaultService) {
		s.signer = signerSvc
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

// WithAntiFraud sets the anti-fraud engine
func WithAntiFraud(engine AntiFraudEngine) ServiceOption {
	return func(s *DefaultService) {
		s.antiFraud = engine
	}
}

// WithCache sets the challenge cache
func WithCache(c cache.Cache[string, *SMSChallenge]) ServiceOption {
	return func(s *DefaultService) {
		s.cache = c
	}
}

// WithMetrics sets the metrics collector
func WithSMSMetrics(m *Metrics) ServiceOption {
	return func(s *DefaultService) {
		s.metrics = m
	}
}

// NewService creates a new SMS verification service
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
		config:       config,
		logger:       logger.With().Str("component", "sms_verification").Logger(),
		metrics:      DefaultMetrics,
		regionLimits: NewRegionRateLimits(),
	}

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	// Create provider if not set
	if s.provider == nil {
		provider, err := createSMSProvider(config, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create SMS provider: %w", err)
		}
		s.provider = provider
	}

	// Create template manager
	s.templateManager = NewTemplateManager(TemplateDefaults{
		ProductName:    "VirtEngine",
		ExpiresMinutes: int(config.OTPTTLSeconds / 60),
	})

	// Create in-memory anti-fraud engine if not set
	if s.antiFraud == nil && config.EnableVelocityChecks {
		s.antiFraud = NewInMemoryAntiFraudEngine(DefaultAntiFraudConfig(), logger)
	}

	// Register metrics if enabled
	if config.MetricsEnabled && s.metrics != nil {
		if err := s.metrics.Register(); err != nil {
			s.logger.Warn().Err(err).Msg("failed to register metrics")
		}
	}

	s.logger.Info().
		Str("primary_provider", config.PrimaryProvider).
		Bool("voip_blocking", config.EnableVoIPBlocking).
		Bool("velocity_checks", config.EnableVelocityChecks).
		Bool("rate_limit_enabled", config.RateLimitEnabled).
		Bool("audit_log_enabled", config.AuditLogEnabled).
		Msg("SMS verification service initialized")

	return s, nil
}

// createSMSProvider creates a new SMS provider based on configuration
func createSMSProvider(config Config, logger zerolog.Logger) (Provider, error) {
	providerConfig, ok := config.ProviderConfigs[config.PrimaryProvider]
	if !ok {
		// Use mock provider for testing
		logger.Warn().Str("provider", config.PrimaryProvider).Msg("provider config not found, using mock")
		return NewMockProvider(logger), nil
	}

	var primary Provider
	var secondary Provider
	var err error

	// Create primary provider using existing factory
	primary, err = NewProvider(config.PrimaryProvider, providerConfig, logger)
	if err != nil {
		return nil, err
	}

	// Create secondary provider if configured
	if config.FailoverEnabled && config.SecondaryProvider != "" {
		secondaryConfig, ok := config.ProviderConfigs[config.SecondaryProvider]
		if ok {
			secondary, err = NewProvider(config.SecondaryProvider, secondaryConfig, logger)
			if err != nil {
				logger.Warn().Err(err).Str("provider", config.SecondaryProvider).Msg("failed to create secondary provider")
			}
		}
	}

	// Wrap in failover provider if secondary is available
	if secondary != nil {
		return NewFailoverProvider(primary, secondary, logger), nil
	}

	return primary, nil
}

// ============================================================================
// Core Operations
// ============================================================================

// InitiateVerification starts a new SMS verification challenge
func (s *DefaultService) InitiateVerification(ctx context.Context, req *InitiateRequest) (*InitiateResponse, error) {
	if s.closed {
		return nil, ErrServiceUnavailable
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Normalize phone number
	phoneNumber, err := NormalizePhoneNumber(req.PhoneNumber, req.CountryCode)
	if err != nil {
		return nil, err
	}
	phoneHash := HashPhoneNumber(phoneNumber)

	// Determine country code
	countryCode := req.CountryCode
	if countryCode == "" {
		countryCode = detectCountryCode(phoneNumber)
	}

	// Check if country is allowed
	if !s.config.IsCountryAllowed(countryCode) {
		return nil, errors.Wrapf(ErrCountryBlocked, "country code: %s", countryCode)
	}

	// Get region-specific limits
	regionLimit := s.regionLimits.GetLimit(countryCode)

	// Check rate limits
	if s.config.RateLimitEnabled && s.rateLimiter != nil {
		allowed, result, err := s.rateLimiter.AllowVerification(
			ctx,
			req.AccountAddress,
			ratelimit.LimitTypeSMSVerification,
		)
		if err != nil {
			s.logger.Error().Err(err).Msg("rate limiter error")
		} else if !allowed {
			if s.metrics != nil {
				s.metrics.RecordRateLimitHit("account")
			}
			return nil, errors.Wrapf(ErrRateLimited, "try again in %d seconds", result.RetryAfter)
		}
	}

	// Perform anti-fraud checks
	var phoneInfo *PhoneInfo
	if s.antiFraud != nil {
		afReq := &AntiFraudRequest{
			AccountAddress:    req.AccountAddress,
			PhoneNumber:       phoneNumber,
			PhoneHash:         phoneHash,
			CountryCode:       countryCode,
			IPAddress:         req.IPAddress,
			DeviceFingerprint: req.DeviceFingerprint,
			UserAgent:         req.UserAgent,
			Timestamp:         time.Now(),
		}

		// Carrier lookup if enabled
		if s.config.EnableCarrierLookup && s.provider != nil {
			carrierResult, err := s.provider.LookupCarrier(ctx, phoneNumber)
			if err != nil {
				s.logger.Warn().Err(err).Msg("carrier lookup failed")
			} else {
				afReq.CarrierInfo = carrierResult
				phoneInfo = &PhoneInfo{
					E164:        phoneNumber,
					CountryCode: countryCode,
					CarrierType: carrierResult.CarrierType,
					CarrierName: carrierResult.CarrierName,
					IsVoIP:      carrierResult.IsVoIP,
					IsMobile:    carrierResult.IsMobile,
					IsValid:     carrierResult.IsValid,
					RiskScore:   carrierResult.RiskScore,
					RiskFactors: carrierResult.RiskFactors,
				}

				if s.metrics != nil {
					s.metrics.RecordCarrierLookup(s.provider.Name(), true, time.Since(afReq.Timestamp))
				}
			}
		}

		// Check anti-fraud
		afResult, err := s.antiFraud.CheckPhone(ctx, afReq)
		if err != nil {
			s.logger.Error().Err(err).Msg("anti-fraud check failed")
		} else if !afResult.Allowed {
			if s.metrics != nil {
				s.metrics.RecordFraudDetected(afResult.BlockReason)
			}
			return nil, errors.Wrap(ErrPhoneBlocked, afResult.BlockReason)
		} else if phoneInfo != nil {
			phoneInfo.RiskScore = afResult.RiskScore
		}

		// Block VoIP if enabled and not bypassed
		if s.config.EnableVoIPBlocking && !req.BypassVoIPCheck {
			if afReq.CarrierInfo != nil && afReq.CarrierInfo.IsVoIP {
				if regionLimit.BlockVoIP {
					if s.metrics != nil {
						carrierName := ""
						if afReq.CarrierInfo != nil {
							carrierName = afReq.CarrierInfo.CarrierName
						}
						s.metrics.RecordVoIPDetected(countryCode, carrierName)
					}
					return nil, ErrVoIPNotAllowed
				}
			}
		}

		// Record the request for velocity tracking
		if err := s.antiFraud.RecordRequest(ctx, afReq); err != nil {
			s.logger.Warn().Err(err).Msg("failed to record request for velocity tracking")
		}
	}

	// Generate challenge ID
	challengeID := uuid.New().String()

	// Create challenge
	challenge, otp, err := NewSMSChallenge(ChallengeConfig{
		ChallengeID:       challengeID,
		AccountAddress:    req.AccountAddress,
		PhoneNumber:       phoneNumber, // Will be hashed internally
		CountryCode:       countryCode,
		CreatedAt:         time.Now(),
		TTLSeconds:        s.config.OTPTTLSeconds,
		MaxAttempts:       s.config.MaxAttempts,
		MaxResends:        s.config.MaxResends,
		IPAddress:         req.IPAddress,
		DeviceFingerprint: req.DeviceFingerprint,
		UserAgent:         req.UserAgent,
		Locale:            req.Locale,
		PhoneInfo:         phoneInfo,
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

	// Send verification SMS
	if err := s.sendVerificationSMS(ctx, phoneNumber, challenge, otp); err != nil {
		// Clean up cache on failure
		if s.cache != nil {
			_ = s.cache.Delete(ctx, challengeID)
		}
		return nil, err
	}

	// Record metrics
	if s.metrics != nil {
		s.metrics.RecordChallengeCreated(countryCode)
		if phoneInfo != nil {
			s.metrics.RecordRiskScore(countryCode, phoneInfo.RiskScore)
		}
	}

	// Audit log
	if s.config.AuditLogEnabled && s.auditor != nil {
		s.auditor.Log(ctx, audit.Event{
			Type:      audit.EventTypeVerificationInitiated,
			Timestamp: time.Now(),
			Actor:     req.AccountAddress,
			Resource:  challengeID,
			Action:    "initiate_sms_verification",
			Details: map[string]interface{}{
				"country_code": countryCode,
				"phone_hash":   phoneHash[:16] + "...",
				"masked_phone": challenge.MaskedPhone,
				"ip_address":   req.IPAddress,
				"is_voip":      challenge.IsVoIP,
				"carrier_type": challenge.CarrierType,
			},
		})
	}

	var carrierType CarrierType
	if phoneInfo != nil {
		carrierType = phoneInfo.CarrierType
	}

	return &InitiateResponse{
		ChallengeID:           challengeID,
		MaskedPhone:           challenge.MaskedPhone,
		ExpiresAt:             challenge.ExpiresAt,
		ResendCooldownSeconds: s.config.ResendCooldownSeconds,
		CarrierType:           carrierType,
		IsVoIP:                challenge.IsVoIP,
	}, nil
}

// VerifyChallenge verifies an SMS challenge with the provided OTP
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
			s.metrics.RecordChallengeExpired(challenge.CountryCode)
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

	// Verify the OTP
	success := challenge.VerifyOTP(req.OTP)
	challenge.RecordAttempt(now, success)

	// Record metrics
	if s.metrics != nil {
		s.metrics.RecordOTPAttempt(challenge.CountryCode, success)
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
			if s.metrics != nil {
				s.metrics.RecordAttestationFailed("signing_error")
			}
			// Still return success, attestation can be retried
		}

		// Record verification latency
		if s.metrics != nil {
			latency := now.Sub(challenge.CreatedAt)
			s.metrics.RecordChallengeVerified(challenge.CountryCode, latency)
		}

		// Audit log
		if s.config.AuditLogEnabled && s.auditor != nil {
			var attestationID string
			if attestation != nil {
				attestationID = attestation.ID
			}
			s.auditor.Log(ctx, audit.Event{
				Type:      audit.EventTypeVerificationCompleted,
				Timestamp: now,
				Actor:     req.AccountAddress,
				Resource:  req.ChallengeID,
				Action:    "verify_sms_success",
				Details: map[string]interface{}{
					"country_code":   challenge.CountryCode,
					"attempts":       challenge.Attempts,
					"attestation_id": attestationID,
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
	if s.metrics != nil {
		s.metrics.RecordOTPFailed(challenge.CountryCode, "invalid_otp")
	}

	if s.config.AuditLogEnabled && s.auditor != nil {
		s.auditor.Log(ctx, audit.Event{
			Type:      audit.EventTypeVerificationFailed,
			Timestamp: now,
			Actor:     req.AccountAddress,
			Resource:  req.ChallengeID,
			Action:    "verify_sms_failed",
			Details: map[string]interface{}{
				"country_code": challenge.CountryCode,
				"attempts":     challenge.Attempts,
				"remaining":    challenge.MaxAttempts - challenge.Attempts,
			},
		})
	}

	// Check if max attempts reached
	if challenge.Attempts >= challenge.MaxAttempts {
		if s.metrics != nil {
			s.metrics.RecordChallengeFailed(challenge.CountryCode, "max_attempts")
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
		ErrorCode:         "INVALID_OTP",
		ErrorMessage:      "The verification code is incorrect",
	}, ErrInvalidOTP
}

// ResendVerification resends the verification SMS
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
	if req.PhoneNumber == "" {
		return nil, errors.Wrap(ErrInvalidPhoneNumber, "phone_number is required for resend")
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

	// Verify phone hash matches
	phoneHash := HashPhoneNumber(req.PhoneNumber)
	if phoneHash != challenge.PhoneHash {
		return nil, errors.Wrap(ErrPhoneMismatch, "phone number does not match challenge")
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

	// Generate new OTP
	otp, otpHash, err := GenerateOTP(s.config.OTPLength)
	if err != nil {
		return nil, err
	}

	// Calculate new expiry
	newExpiresAt := now.Add(time.Duration(s.config.OTPTTLSeconds) * time.Second)

	// Update challenge
	if err := challenge.RecordResend(now, otpHash, newExpiresAt); err != nil {
		return nil, err
	}

	// Send new SMS
	if err := s.sendVerificationSMS(ctx, req.PhoneNumber, challenge, otp); err != nil {
		return nil, err
	}

	// Update cache
	if err := s.updateChallenge(ctx, challenge); err != nil {
		s.logger.Error().Err(err).Msg("failed to update challenge")
	}

	// Record metrics
	if s.metrics != nil {
		s.metrics.RecordOTPResend(challenge.CountryCode)
	}

	// Audit log
	if s.config.AuditLogEnabled && s.auditor != nil {
		s.auditor.Log(ctx, audit.Event{
			Type:      audit.EventTypeResendRequested,
			Timestamp: now,
			Actor:     req.AccountAddress,
			Resource:  req.ChallengeID,
			Action:    "resend_sms_verification",
			Details: map[string]interface{}{
				"country_code": challenge.CountryCode,
				"resend_count": challenge.ResendCount,
				"max_resends":  challenge.MaxResends,
			},
		})
	}

	nextResendAt := now.Add(time.Duration(s.config.ResendCooldownSeconds) * time.Second)

	return &ResendResponse{
		Success:          true,
		ExpiresAt:        newExpiresAt,
		ResendsRemaining: challenge.MaxResends - challenge.ResendCount,
		NextResendAt:     nextResendAt,
	}, nil
}

// GetChallenge retrieves a challenge by ID
func (s *DefaultService) GetChallenge(ctx context.Context, challengeID string) (*SMSChallenge, error) {
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
			Action:    "cancel_sms_verification",
		})
	}

	return nil
}

// ProcessWebhook processes a delivery status webhook
func (s *DefaultService) ProcessWebhook(ctx context.Context, providerName string, payload []byte) error {
	events, err := s.provider.ParseWebhook(ctx, payload, s.config.WebhookSecret)
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
				Provider:          challenge.Provider,
				FailoverUsed:      challenge.FailoverUsed,
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
		Provider:          challenge.Provider,
		FailoverUsed:      challenge.FailoverUsed,
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
		status.PrimaryProviderHealthy = false
		status.Details["provider_error"] = err.Error()
		status.Healthy = false
		status.Status = "provider unhealthy"
	} else {
		status.PrimaryProviderHealthy = true
		if s.metrics != nil {
			s.metrics.SetProviderHealth(s.provider.Name(), true)
		}
	}

	// Check cache health
	if s.cache != nil {
		_, _ = s.cache.Get(ctx, "health-check-test")
		status.CacheHealthy = true
	} else {
		status.CacheHealthy = true // No cache configured is OK
	}

	status.Details["provider"] = s.provider.Name()

	return status, nil
}

// LookupPhoneInfo performs carrier/VoIP lookup on a phone number
func (s *DefaultService) LookupPhoneInfo(ctx context.Context, phoneNumber string) (*PhoneInfo, error) {
	if s.closed {
		return nil, ErrServiceUnavailable
	}

	// Normalize phone number
	normalized, err := NormalizePhoneNumber(phoneNumber, "")
	if err != nil {
		return nil, err
	}

	// Perform carrier lookup
	if s.provider == nil {
		return nil, errors.Wrap(ErrServiceUnavailable, "no provider configured")
	}

	result, err := s.provider.LookupCarrier(ctx, normalized)
	if err != nil {
		return nil, errors.Wrap(ErrCarrierLookupFailed, err.Error())
	}

	return &PhoneInfo{
		E164:        result.PhoneNumber,
		CountryCode: result.CountryCode,
		CarrierType: result.CarrierType,
		CarrierName: result.CarrierName,
		IsVoIP:      result.IsVoIP,
		IsMobile:    result.IsMobile,
		IsValid:     result.IsValid,
		RiskScore:   result.RiskScore,
		RiskFactors: result.RiskFactors,
	}, nil
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

	// Close anti-fraud engine
	if s.antiFraud != nil {
		if err := s.antiFraud.Close(); err != nil {
			s.logger.Error().Err(err).Msg("failed to close anti-fraud engine")
		}
	}

	s.logger.Info().Msg("SMS verification service closed")
	return nil
}

// ============================================================================
// Internal Helpers
// ============================================================================

// sendVerificationSMS sends the verification SMS
func (s *DefaultService) sendVerificationSMS(ctx context.Context, phoneNumber string, challenge *SMSChallenge, otp string) error {
	// Render SMS message using template
	locale := challenge.Locale
	if locale == "" {
		locale = s.config.DefaultLocale
	}

	message, err := s.templateManager.RenderMessage(TemplateOTPVerification, TemplateData{
		OTP:            otp,
		ExpiresMinutes: int(s.config.OTPTTLSeconds / 60),
	}, locale)
	if err != nil {
		return errors.Wrap(ErrTemplateError, err.Error())
	}

	// Create SMS message
	smsMsg := &SMSMessage{
		To:   phoneNumber,
		Body: message,
		Metadata: map[string]string{
			"challenge_id":    challenge.ChallengeID,
			"account_address": challenge.AccountAddress,
			"country_code":    challenge.CountryCode,
		},
		Tags: []string{"verification", "otp"},
	}

	// Get from number from config
	if providerConfig, ok := s.config.ProviderConfigs[s.config.PrimaryProvider]; ok {
		smsMsg.From = providerConfig.FromNumber
	}

	// Send SMS
	startTime := time.Now()
	result, err := s.provider.Send(ctx, smsMsg)
	sendLatency := time.Since(startTime)

	if err != nil {
		if s.metrics != nil {
			s.metrics.RecordSMSFailed(s.provider.Name(), "send_error")
		}
		return errors.Wrap(ErrDeliveryFailed, err.Error())
	}

	// Update challenge with delivery info
	challenge.UpdateDeliveryStatus(DeliverySent, result.MessageID, result.Provider)

	// Check if failover was used
	if result.Provider != s.config.PrimaryProvider {
		challenge.FailoverUsed = true
		if s.metrics != nil {
			s.metrics.RecordProviderFailover(s.config.PrimaryProvider, result.Provider)
		}
	}

	// Record metrics
	if s.metrics != nil {
		s.metrics.RecordSMSSent(result.Provider, challenge.CountryCode, sendLatency)
	}

	s.logger.Info().
		Str("challenge_id", challenge.ChallengeID).
		Str("message_id", result.MessageID).
		Str("provider", result.Provider).
		Str("country_code", challenge.CountryCode).
		Dur("latency", sendLatency).
		Bool("failover", challenge.FailoverUsed).
		Msg("verification SMS sent")

	return nil
}

// getChallenge retrieves a challenge from cache
func (s *DefaultService) getChallenge(ctx context.Context, challengeID string) (*SMSChallenge, error) {
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
func (s *DefaultService) updateChallenge(ctx context.Context, challenge *SMSChallenge) error {
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

// createAttestation creates a verification attestation for a verified phone
func (s *DefaultService) createAttestation(ctx context.Context, challenge *SMSChallenge) (*veidtypes.VerificationAttestation, error) {
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

	// Calculate confidence based on risk factors
	confidence := uint32(100)
	if challenge.RiskScore > 0 {
		confidence = 100 - challenge.RiskScore
		if confidence < 50 {
			confidence = 50
		}
	}

	// Create attestation
	attestation := veidtypes.NewVerificationAttestation(
		veidtypes.AttestationIssuer{}, // Will be set by signer
		subject,
		veidtypes.AttestationTypeSMSVerification,
		nonceBytes,
		now,
		validityDuration,
		100, // Score (100 for successful verification)
		confidence,
	)

	// Add verification proof
	proofDetail := veidtypes.NewVerificationProofDetail(
		"sms_otp_verification",
		challenge.PhoneHash,
		100,
		50, // Threshold
		now,
	)
	attestation.AddVerificationProof(proofDetail)

	// Add metadata
	attestation.SetMetadata("phone_hash", challenge.PhoneHash[:16]+"...")
	attestation.SetMetadata("country_code", challenge.CountryCode)
	attestation.SetMetadata("carrier_type", string(challenge.CarrierType))
	attestation.SetMetadata("is_voip", fmt.Sprintf("%t", challenge.IsVoIP))
	attestation.SetMetadata("risk_score", fmt.Sprintf("%d", challenge.RiskScore))

	// Sign the attestation
	if err := s.signer.SignAttestation(ctx, attestation); err != nil {
		return nil, errors.Wrap(ErrSigningFailed, err.Error())
	}

	// Record metrics
	if s.metrics != nil {
		s.metrics.RecordAttestationCreated(challenge.CountryCode)
	}

	s.logger.Info().
		Str("attestation_id", attestation.ID).
		Str("challenge_id", challenge.ChallengeID).
		Str("account", challenge.AccountAddress).
		Str("country_code", challenge.CountryCode).
		Msg("SMS verification attestation created")

	return attestation, nil
}

// processWebhookEvent processes a single webhook event
func (s *DefaultService) processWebhookEvent(ctx context.Context, event WebhookEvent) {
	s.logger.Debug().
		Str("event_type", string(event.EventType)).
		Str("message_id", event.MessageID).
		Msg("processing webhook event")

	// Update metrics based on event type
	switch event.EventType {
	case WebhookEventDelivered:
		if s.metrics != nil {
			s.metrics.RecordSMSDelivered(event.Provider, "")
		}
	case WebhookEventFailed:
		if s.metrics != nil {
			s.metrics.RecordSMSFailed(event.Provider, event.ErrorCode)
		}
	case WebhookEventUndelivered:
		if s.metrics != nil {
			s.metrics.RecordSMSFailed(event.Provider, "undelivered")
		}
	}

	// Audit log
	if s.config.AuditLogEnabled && s.auditor != nil {
		s.auditor.Log(ctx, audit.Event{
			Type:      audit.EventTypeWebhookReceived,
			Timestamp: event.Timestamp,
			Resource:  event.MessageID,
			Action:    "sms_webhook_" + string(event.EventType),
			Details: map[string]interface{}{
				"event_type": event.EventType,
				"error_code": event.ErrorCode,
				"provider":   event.Provider,
			},
		})
	}
}

// detectCountryCode attempts to detect the country code from a phone number
func detectCountryCode(phoneNumber string) string {
	if len(phoneNumber) < 2 {
		return ""
	}

	// Common country calling codes
	if len(phoneNumber) >= 2 && phoneNumber[0] == '+' {
		switch {
		case len(phoneNumber) >= 2 && phoneNumber[1] == '1':
			// North American Numbering Plan (US, Canada, etc.)
			return "US"
		case len(phoneNumber) >= 3 && phoneNumber[1:3] == "44":
			return "GB"
		case len(phoneNumber) >= 3 && phoneNumber[1:3] == "91":
			return "IN"
		case len(phoneNumber) >= 3 && phoneNumber[1:3] == "86":
			return "CN"
		case len(phoneNumber) >= 3 && phoneNumber[1:3] == "81":
			return "JP"
		case len(phoneNumber) >= 3 && phoneNumber[1:3] == "49":
			return "DE"
		case len(phoneNumber) >= 3 && phoneNumber[1:3] == "33":
			return "FR"
		case len(phoneNumber) >= 3 && phoneNumber[1:3] == "55":
			return "BR"
		case len(phoneNumber) >= 3 && phoneNumber[1:3] == "61":
			return "AU"
		case len(phoneNumber) >= 3 && phoneNumber[1:3] == "52":
			return "MX"
		}
	}

	return ""
}

// GenerateNonce generates a cryptographically secure nonce
func GenerateNonce(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Ensure DefaultService implements SMSVerificationService
var _ SMSVerificationService = (*DefaultService)(nil)
