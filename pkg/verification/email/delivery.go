// Package email provides email OTP delivery service with retry logic.
//
// This file implements the complete email delivery pipeline with:
// - Retry with exponential backoff for transient failures
// - Delivery status tracking and webhooks
// - Provider failover support
// - Rate limiting integration
//
// Task Reference: VE-3F - Email Verification Delivery + Attestation
package email

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/virtengine/virtengine/pkg/errors"
	"github.com/virtengine/virtengine/pkg/verification/audit"
	"github.com/virtengine/virtengine/pkg/verification/ratelimit"
)

// ============================================================================
// Delivery Configuration
// ============================================================================

const (
	// DefaultMaxRetries is the default maximum number of delivery retries
	DefaultMaxRetries = 3

	// DefaultRetryBaseDelay is the base delay for exponential backoff
	DefaultRetryBaseDelay = time.Second * 2

	// DefaultRetryMaxDelay is the maximum delay between retries
	DefaultRetryMaxDelay = time.Minute

	// DefaultDeliveryTimeout is the timeout for a single delivery attempt
	DefaultDeliveryTimeout = time.Second * 30
)

// DeliveryConfig contains configuration for the delivery service.
type DeliveryConfig struct {
	// MaxRetries is the maximum number of delivery retries
	MaxRetries int `json:"max_retries"`

	// RetryBaseDelay is the base delay for exponential backoff
	RetryBaseDelay time.Duration `json:"retry_base_delay"`

	// RetryMaxDelay is the maximum delay between retries
	RetryMaxDelay time.Duration `json:"retry_max_delay"`

	// DeliveryTimeout is the timeout for a single delivery attempt
	DeliveryTimeout time.Duration `json:"delivery_timeout"`

	// EnableFailover enables provider failover on repeated failures
	EnableFailover bool `json:"enable_failover"`

	// FailoverProviders are backup providers for failover
	FailoverProviders []string `json:"failover_providers,omitempty"`
}

// DefaultDeliveryConfig returns the default delivery configuration.
func DefaultDeliveryConfig() DeliveryConfig {
	return DeliveryConfig{
		MaxRetries:      DefaultMaxRetries,
		RetryBaseDelay:  DefaultRetryBaseDelay,
		RetryMaxDelay:   DefaultRetryMaxDelay,
		DeliveryTimeout: DefaultDeliveryTimeout,
		EnableFailover:  false,
	}
}

// ============================================================================
// Delivery Service
// ============================================================================

// DeliveryService handles email delivery with retry and failover support.
type DeliveryService struct {
	config          DeliveryConfig
	primaryProvider Provider
	failoverProvs   []Provider
	rateLimiter     ratelimit.VerificationLimiter
	auditor         audit.AuditLogger
	templateMgr     *TemplateManager
	metrics         *Metrics
	logger          zerolog.Logger

	// State
	mu                  sync.RWMutex
	providerFailures    map[string]int
	lastProviderFailure map[string]time.Time
	closed              bool
}

// DeliveryServiceOption is a functional option for configuring the delivery service.
type DeliveryServiceOption func(*DeliveryService)

// WithDeliveryRateLimiter sets the rate limiter.
func WithDeliveryRateLimiter(limiter ratelimit.VerificationLimiter) DeliveryServiceOption {
	return func(s *DeliveryService) {
		s.rateLimiter = limiter
	}
}

// WithDeliveryAuditor sets the audit logger.
func WithDeliveryAuditor(auditor audit.AuditLogger) DeliveryServiceOption {
	return func(s *DeliveryService) {
		s.auditor = auditor
	}
}

// WithDeliveryMetrics sets the metrics collector.
func WithDeliveryMetrics(m *Metrics) DeliveryServiceOption {
	return func(s *DeliveryService) {
		s.metrics = m
	}
}

// WithFailoverProviders sets the failover providers.
func WithFailoverProviders(providers []Provider) DeliveryServiceOption {
	return func(s *DeliveryService) {
		s.failoverProvs = providers
	}
}

// NewDeliveryService creates a new delivery service.
func NewDeliveryService(
	config DeliveryConfig,
	primaryProvider Provider,
	templateMgr *TemplateManager,
	logger zerolog.Logger,
	opts ...DeliveryServiceOption,
) *DeliveryService {
	s := &DeliveryService{
		config:              config,
		primaryProvider:     primaryProvider,
		templateMgr:         templateMgr,
		logger:              logger.With().Str("component", "email_delivery").Logger(),
		failoverProvs:       make([]Provider, 0),
		providerFailures:    make(map[string]int),
		lastProviderFailure: make(map[string]time.Time),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// ============================================================================
// Delivery Operations
// ============================================================================

// DeliveryRequest represents a request to deliver an email.
type DeliveryRequest struct {
	// ChallengeID is the verification challenge ID
	ChallengeID string `json:"challenge_id"`

	// AccountAddress is the requesting account
	AccountAddress string `json:"account_address"`

	// Email is the recipient email address
	Email string `json:"email"`

	// Method is the verification method (OTP or magic link)
	Method VerificationMethod `json:"method"`

	// Secret is the OTP or magic link token (to be sent)
	Secret string `json:"secret"`

	// ExpiresAt is when the challenge expires
	ExpiresAt time.Time `json:"expires_at"`

	// Locale is the preferred locale for the email
	Locale string `json:"locale,omitempty"`

	// RequestID is an optional request identifier
	RequestID string `json:"request_id,omitempty"`

	// IPAddress is the client IP address
	IPAddress string `json:"ip_address,omitempty"`
}

// Validate validates the delivery request.
func (r *DeliveryRequest) Validate() error {
	if r.ChallengeID == "" {
		return errors.Wrap(ErrInvalidRequest, "challenge_id is required")
	}
	if r.AccountAddress == "" {
		return errors.Wrap(ErrInvalidRequest, "account_address is required")
	}
	if r.Email == "" {
		return errors.Wrap(ErrInvalidEmail, "email is required")
	}
	if r.Secret == "" {
		return errors.Wrap(ErrInvalidRequest, "secret is required")
	}
	if r.Method != MethodOTP && r.Method != MethodMagicLink {
		return errors.Wrapf(ErrInvalidRequest, "invalid method: %s", r.Method)
	}
	return nil
}

// Deliver sends a verification email with retry support.
func (s *DeliveryService) Deliver(ctx context.Context, req *DeliveryRequest) (*DeliveryResult, error) {
	if s.closed {
		return nil, ErrServiceUnavailable
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check rate limits if configured
	if s.rateLimiter != nil {
		allowed, result, err := s.rateLimiter.AllowVerification(
			ctx,
			req.AccountAddress,
			ratelimit.LimitTypeEmailVerification,
		)
		if err != nil {
			s.logger.Warn().Err(err).Msg("rate limiter error, allowing request")
		} else if !allowed {
			if s.metrics != nil {
				s.metrics.RecordRateLimitHit()
			}
			return nil, errors.Wrapf(ErrRateLimited, "retry after %d seconds", result.RetryAfter)
		}
	}

	// Build the email
	email, err := s.buildEmail(req)
	if err != nil {
		return nil, err
	}

	// Attempt delivery with retries
	startTime := time.Now()
	result, err := s.deliverWithRetry(ctx, email, req.ChallengeID)
	deliveryDuration := time.Since(startTime)

	// Record metrics
	if s.metrics != nil {
		if err != nil {
			s.metrics.RecordChallengeFailed(req.Method, "delivery_failed")
		} else {
			s.metrics.RecordEmailSent(result.Provider, TemplateOTPVerification, deliveryDuration)
		}
	}

	// Audit log
	if s.auditor != nil {
		outcome := audit.OutcomeSuccess
		if err != nil {
			outcome = audit.OutcomeFailure
		}
		s.auditor.Log(ctx, audit.Event{
			Type:      audit.EventTypeVerificationInitiated,
			Timestamp: time.Now(),
			Actor:     req.AccountAddress,
			Resource:  req.ChallengeID,
			Action:    "email_delivery",
			Outcome:   outcome,
			Duration:  deliveryDuration,
			Details: map[string]interface{}{
				"method":     req.Method,
				"provider":   result.Provider,
				"message_id": result.ProviderMessageID,
				"ip_address": req.IPAddress,
			},
		})
	}

	if err != nil {
		return nil, err
	}

	s.logger.Info().
		Str("challenge_id", req.ChallengeID).
		Str("message_id", result.ProviderMessageID).
		Str("provider", result.Provider).
		Dur("duration", deliveryDuration).
		Msg("verification email delivered successfully")

	return result, nil
}

// buildEmail constructs the email from the delivery request.
func (s *DeliveryService) buildEmail(req *DeliveryRequest) (*Email, error) {
	var templateType TemplateType
	var templateData TemplateData

	switch req.Method {
	case MethodOTP:
		templateType = TemplateOTPVerification
		templateData = TemplateData{
			OTP:       req.Secret,
			ExpiresIn: FormatExpiryTime(req.ExpiresAt),
			ExpiresAt: req.ExpiresAt,
		}
	case MethodMagicLink:
		templateType = TemplateMagicLink
		templateData = TemplateData{
			VerificationLink: req.Secret, // Secret is the full link for magic links
			ExpiresIn:        FormatExpiryTime(req.ExpiresAt),
			ExpiresAt:        req.ExpiresAt,
		}
	default:
		return nil, errors.Wrapf(ErrInvalidRequest, "unsupported method: %s", req.Method)
	}

	email, err := s.templateMgr.RenderEmail(templateType, templateData, req.Email)
	if err != nil {
		return nil, errors.Wrap(ErrTemplateError, err.Error())
	}

	// Add tracking metadata
	email.Metadata = map[string]string{
		"challenge_id":    req.ChallengeID,
		"account_address": req.AccountAddress,
		"request_id":      req.RequestID,
	}
	email.Tags = []string{"verification", string(req.Method)}

	return email, nil
}

// deliverWithRetry attempts delivery with exponential backoff.
func (s *DeliveryService) deliverWithRetry(ctx context.Context, email *Email, challengeID string) (*DeliveryResult, error) {
	var lastErr error
	providers := s.getProviders()

	for providerIdx, provider := range providers {
		for attempt := 0; attempt <= s.config.MaxRetries; attempt++ {
			// Calculate delay for retry (exponential backoff)
			if attempt > 0 {
				delay := s.calculateRetryDelay(attempt)
				s.logger.Debug().
					Int("attempt", attempt).
					Dur("delay", delay).
					Str("provider", provider.Name()).
					Msg("retrying email delivery")

				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(delay):
				}
			}

			// Create timeout context for this attempt
			attemptCtx, cancel := context.WithTimeout(ctx, s.config.DeliveryTimeout)
			result, err := provider.Send(attemptCtx, email)
			cancel()

			if err == nil && result.Success {
				// Success - reset failure counter for this provider
				s.recordProviderSuccess(provider.Name())
				return &DeliveryResult{
					ChallengeID:       challengeID,
					Success:           true,
					ProviderMessageID: result.MessageID,
					DeliveryStatus:    DeliverySent,
					SentAt:            result.Timestamp,
					Provider:          provider.Name(),
				}, nil
			}

			// Record failure
			lastErr = err
			s.recordProviderFailure(provider.Name())
			s.logger.Warn().
				Err(err).
				Int("attempt", attempt+1).
				Int("max_retries", s.config.MaxRetries).
				Str("provider", provider.Name()).
				Msg("email delivery attempt failed")

			// Check if we should try failover
			if attempt == s.config.MaxRetries && s.config.EnableFailover && providerIdx < len(providers)-1 {
				s.logger.Info().
					Str("from_provider", provider.Name()).
					Str("to_provider", providers[providerIdx+1].Name()).
					Msg("failing over to backup provider")
				break // Try next provider
			}
		}
	}

	// All attempts failed
	errMsg := "delivery failed after all retries"
	if lastErr != nil {
		errMsg = fmt.Sprintf("%s: %v", errMsg, lastErr)
	}

	return &DeliveryResult{
		ChallengeID:    challengeID,
		Success:        false,
		DeliveryStatus: DeliveryFailed,
		SentAt:         time.Now(),
		Provider:       s.primaryProvider.Name(),
		ErrorCode:      "DELIVERY_FAILED",
		ErrorMessage:   errMsg,
	}, errors.Wrap(ErrDeliveryFailed, errMsg)
}

// getProviders returns the list of providers to try (primary + failover).
func (s *DeliveryService) getProviders() []Provider {
	providers := []Provider{s.primaryProvider}
	if s.config.EnableFailover {
		providers = append(providers, s.failoverProvs...)
	}
	return providers
}

// calculateRetryDelay calculates the delay for a retry attempt using exponential backoff.
func (s *DeliveryService) calculateRetryDelay(attempt int) time.Duration {
	shift := attempt - 1
	if shift < 0 {
		shift = 0
	}
	if shift > 30 {
		shift = 30
	}
	delay := s.config.RetryBaseDelay * time.Duration(1<<shift)
	if delay > s.config.RetryMaxDelay {
		delay = s.config.RetryMaxDelay
	}
	return delay
}

// recordProviderSuccess resets the failure counter for a provider.
func (s *DeliveryService) recordProviderSuccess(providerName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.providerFailures[providerName] = 0
}

// recordProviderFailure increments the failure counter for a provider.
func (s *DeliveryService) recordProviderFailure(providerName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.providerFailures[providerName]++
	s.lastProviderFailure[providerName] = time.Now()
}

// GetProviderHealth returns the health status of a provider.
func (s *DeliveryService) GetProviderHealth(providerName string) (failures int, lastFailure time.Time) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.providerFailures[providerName], s.lastProviderFailure[providerName]
}

// ============================================================================
// Batch Delivery
// ============================================================================

// BatchDeliveryResult contains the result of a batch delivery operation.
type BatchDeliveryResult struct {
	// TotalRequested is the number of deliveries requested
	TotalRequested int `json:"total_requested"`

	// Successful is the number of successful deliveries
	Successful int `json:"successful"`

	// Failed is the number of failed deliveries
	Failed int `json:"failed"`

	// Results contains individual delivery results
	Results []*DeliveryResult `json:"results"`
}

// DeliverBatch sends multiple verification emails.
func (s *DeliveryService) DeliverBatch(ctx context.Context, requests []*DeliveryRequest) (*BatchDeliveryResult, error) {
	if s.closed {
		return nil, ErrServiceUnavailable
	}

	if len(requests) == 0 {
		return &BatchDeliveryResult{}, nil
	}

	result := &BatchDeliveryResult{
		TotalRequested: len(requests),
		Results:        make([]*DeliveryResult, len(requests)),
	}

	// Process each request (could be parallelized with worker pool)
	for i, req := range requests {
		deliveryResult, err := s.Deliver(ctx, req)
		if err != nil {
			result.Failed++
			result.Results[i] = &DeliveryResult{
				ChallengeID:    req.ChallengeID,
				Success:        false,
				DeliveryStatus: DeliveryFailed,
				SentAt:         time.Now(),
				ErrorMessage:   err.Error(),
			}
		} else {
			result.Successful++
			result.Results[i] = deliveryResult
		}
	}

	return result, nil
}

// ============================================================================
// Delivery Status Tracking
// ============================================================================

// TrackDeliveryStatus polls for delivery status updates.
func (s *DeliveryService) TrackDeliveryStatus(ctx context.Context, messageID string, challengeID string) (*DeliveryResult, error) {
	status, err := s.primaryProvider.GetDeliveryStatus(ctx, messageID)
	if err != nil {
		return nil, errors.Wrap(ErrProviderError, err.Error())
	}

	return &DeliveryResult{
		ChallengeID:       challengeID,
		Success:           status.Status == DeliveryDelivered,
		ProviderMessageID: messageID,
		DeliveryStatus:    status.Status,
		SentAt:            status.Timestamp,
		Provider:          s.primaryProvider.Name(),
	}, nil
}

// ============================================================================
// Lifecycle
// ============================================================================

// HealthCheck returns the health status of the delivery service.
func (s *DeliveryService) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	status := &HealthStatus{
		Healthy:   true,
		Status:    "healthy",
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	// Check primary provider
	if err := s.primaryProvider.HealthCheck(ctx); err != nil {
		status.ProviderHealthy = false
		status.Details["primary_provider_error"] = err.Error()
		status.Healthy = false
		status.Status = "primary provider unhealthy"
	} else {
		status.ProviderHealthy = true
	}

	// Check failover providers
	for i, provider := range s.failoverProvs {
		if err := provider.HealthCheck(ctx); err != nil {
			status.Details[fmt.Sprintf("failover_%d_error", i)] = err.Error()
		}
	}

	status.Details["primary_provider"] = s.primaryProvider.Name()
	status.Details["failover_count"] = len(s.failoverProvs)

	return status, nil
}

// Close closes the delivery service.
func (s *DeliveryService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true

	// Close providers
	if err := s.primaryProvider.Close(); err != nil {
		s.logger.Error().Err(err).Msg("failed to close primary provider")
	}

	for _, provider := range s.failoverProvs {
		if err := provider.Close(); err != nil {
			s.logger.Error().Err(err).Str("provider", provider.Name()).Msg("failed to close failover provider")
		}
	}

	s.logger.Info().Msg("delivery service closed")
	return nil
}
