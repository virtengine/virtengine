// Package sso provides the SSO/OIDC verification service for VEID.
package sso

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/virtengine/virtengine/pkg/verification/audit"
	"github.com/virtengine/virtengine/pkg/verification/oidc"
	"github.com/virtengine/virtengine/pkg/verification/ratelimit"
	"github.com/virtengine/virtengine/pkg/verification/signer"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Default Verification Service
// ============================================================================

// DefaultService implements VerificationService.
type DefaultService struct {
	config      Config
	oidcVerifier oidc.OIDCVerifier
	signer      signer.SignerService
	rateLimiter ratelimit.VerificationLimiter
	auditor     audit.AuditLogger
	logger      zerolog.Logger

	// Challenge storage (in-memory for now, should use Redis in production)
	mu          sync.RWMutex
	challenges  map[string]*Challenge
	byAccount   map[string][]string // accountAddress -> challengeIDs
}

// NewDefaultService creates a new DefaultService.
func NewDefaultService(
	ctx context.Context,
	config Config,
	oidcVerifier oidc.OIDCVerifier,
	signerSvc signer.SignerService,
	rateLimiter ratelimit.VerificationLimiter,
	auditor audit.AuditLogger,
	logger zerolog.Logger,
) (*DefaultService, error) {
	s := &DefaultService{
		config:      config,
		oidcVerifier: oidcVerifier,
		signer:      signerSvc,
		rateLimiter: rateLimiter,
		auditor:     auditor,
		logger:      logger.With().Str("component", "sso_service").Logger(),
		challenges:  make(map[string]*Challenge),
		byAccount:   make(map[string][]string),
	}

	// Start background cleanup
	go s.cleanupExpiredChallenges(ctx)

	s.logger.Info().Msg("SSO verification service initialized")
	return s, nil
}

// InitiateVerification starts an SSO verification flow.
func (s *DefaultService) InitiateVerification(ctx context.Context, req *InitiateRequest) (*InitiateResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Rate limiting
	if s.config.RateLimitEnabled && s.rateLimiter != nil {
		allowed, result, err := s.rateLimiter.AllowVerification(ctx, req.AccountAddress, ratelimit.LimitTypeSSOVerification)
		if err != nil {
			s.logger.Warn().Err(err).Msg("rate limit check failed")
		} else if !allowed {
			s.logAudit(ctx, audit.EventTypeRateLimitExceeded, req.AccountAddress, "initiate_verification", map[string]interface{}{
				"reason":   result.Reason,
				"retry_after": result.RetryAfter,
			})
			return nil, fmt.Errorf("%w: retry after %d seconds", ErrRateLimitExceeded, result.RetryAfter)
		}
	}

	// Check max pending challenges
	s.mu.RLock()
	pendingCount := len(s.byAccount[req.AccountAddress])
	s.mu.RUnlock()

	if pendingCount >= s.config.MaxChallengesPerAccount {
		return nil, fmt.Errorf("%w: account has %d pending challenges", ErrMaxChallengesExceeded, pendingCount)
	}

	// Generate challenge parameters
	challengeID := uuid.New().String()
	state := generateSecureToken(32)
	nonce := generateSecureToken(32)
	now := time.Now()
	expiresAt := now.Add(time.Duration(s.config.ChallengeTTLSeconds) * time.Second)

	// Determine OIDC issuer
	oidcIssuer := req.OIDCIssuer
	if oidcIssuer == "" {
		if knownIssuer, ok := oidc.WellKnownIssuers[req.ProviderType]; ok {
			oidcIssuer = knownIssuer
		} else {
			return nil, fmt.Errorf("%w: provider: %s", ErrProviderNotConfigured, req.ProviderType)
		}
	}

	// Generate linkage message
	linkageMessage := req.LinkageMessage
	if linkageMessage == "" {
		linkageMessage = fmt.Sprintf(s.config.LinkageMessageTemplate, req.AccountAddress, now.UTC().Format(time.RFC3339), nonce)
	}

	// Get authorization URL from OIDC verifier
	policy, err := s.oidcVerifier.GetIssuerPolicy(ctx, oidcIssuer)
	if err != nil {
		return nil, fmt.Errorf("%w: issuer %s: %v", ErrProviderNotConfigured, oidcIssuer, err)
	}

	scopes := req.RequestedScopes
	if len(scopes) == 0 {
		scopes = policy.RequiredScopes
	}

	authURL, err := s.oidcVerifier.GetAuthorizationURL(ctx, &oidc.AuthorizationRequest{
		ProviderType: req.ProviderType,
		Issuer:       oidcIssuer,
		ClientID:     policy.ClientID,
		RedirectURI:  req.RedirectURI,
		State:        state,
		Nonce:        nonce,
		Scopes:       scopes,
	})
	if err != nil {
		return nil, err
	}

	// Create challenge
	challenge := &Challenge{
		ChallengeID:    challengeID,
		AccountAddress: req.AccountAddress,
		ProviderType:   req.ProviderType,
		OIDCIssuer:     oidcIssuer,
		State:          state,
		Nonce:          nonce,
		LinkageMessage: linkageMessage,
		RedirectURI:    req.RedirectURI,
		Status:         ChallengeStatusPending,
		CreatedAt:      now,
		ExpiresAt:      expiresAt,
		ClientIP:       req.ClientIP,
	}

	// Store challenge
	s.mu.Lock()
	s.challenges[challengeID] = challenge
	s.byAccount[req.AccountAddress] = append(s.byAccount[req.AccountAddress], challengeID)
	s.mu.Unlock()

	s.logAudit(ctx, audit.EventTypeVerificationInitiated, req.AccountAddress, "initiate_verification", map[string]interface{}{
		"challenge_id": challengeID,
		"provider":     req.ProviderType,
		"issuer":       oidcIssuer,
	})

	return &InitiateResponse{
		ChallengeID:      challengeID,
		AuthorizationURL: authURL,
		State:            state,
		Nonce:            nonce,
		LinkageMessage:   linkageMessage,
		ExpiresAt:        expiresAt,
	}, nil
}

// CompleteVerification completes an SSO verification with OIDC token.
func (s *DefaultService) CompleteVerification(ctx context.Context, req *CompleteRequest) (*CompleteResponse, error) {
	if err := req.Validate(); err != nil {
		return NewCompleteError("validation_error", err.Error()), nil
	}

	// Get challenge
	challenge, err := s.GetChallenge(ctx, req.ChallengeID)
	if err != nil {
		return NewCompleteError("challenge_error", err.Error()), nil
	}

	// Validate state
	if challenge.State != req.State {
		s.logAudit(ctx, audit.EventTypeSecurityAlert, challenge.AccountAddress, "state_mismatch", map[string]interface{}{
			"challenge_id": req.ChallengeID,
		})
		return NewCompleteError("state_mismatch", "OAuth state parameter mismatch"), nil
	}

	// Check if challenge can be completed
	now := time.Now()
	if !challenge.CanComplete(now) {
		if challenge.IsExpired(now) {
			s.updateChallengeStatus(req.ChallengeID, ChallengeStatusExpired)
			return NewCompleteError("challenge_expired", "SSO challenge has expired"), nil
		}
		return NewCompleteError("challenge_invalid", fmt.Sprintf("challenge status: %s", challenge.Status)), nil
	}

	// Get issuer policy for verification
	policy, err := s.oidcVerifier.GetIssuerPolicy(ctx, challenge.OIDCIssuer)
	if err != nil {
		return NewCompleteError("config_error", err.Error()), nil
	}

	// Verify the ID token
	claims, err := s.oidcVerifier.VerifyToken(ctx, req.IDToken, &oidc.VerificationRequest{
		ExpectedAudience:     policy.ClientID,
		ExpectedIssuer:       challenge.OIDCIssuer,
		ExpectedNonce:        challenge.Nonce,
		RequiredClaims:       policy.RequiredClaims,
		MaxAge:               policy.MaxAuthAgeSeconds,
		AllowedACRValues:     policy.AllowedACRValues,
		RequireEmailVerified: policy.RequireEmailVerified,
		ProviderType:         challenge.ProviderType,
	})
	if err != nil {
		s.updateChallengeStatus(req.ChallengeID, ChallengeStatusFailed)
		s.logAudit(ctx, audit.EventTypeVerificationFailed, challenge.AccountAddress, "token_verification_failed", map[string]interface{}{
			"challenge_id": req.ChallengeID,
			"error":        err.Error(),
		})
		return NewCompleteError("token_verification_failed", err.Error()), nil
	}

	// TODO: Verify linkage signature (requires crypto key verification)
	// For now, we just check that a signature was provided

	// Create attestation
	attestation, err := s.createAttestation(ctx, challenge, claims, req.LinkageSignature)
	if err != nil {
		s.updateChallengeStatus(req.ChallengeID, ChallengeStatusFailed)
		return NewCompleteError("attestation_failed", err.Error()), nil
	}

	// Sign attestation
	if err := s.signer.SignAttestation(ctx, &attestation.VerificationAttestation); err != nil {
		s.updateChallengeStatus(req.ChallengeID, ChallengeStatusFailed)
		return NewCompleteError("signing_failed", err.Error()), nil
	}

	// Generate linkage ID
	linkageID := fmt.Sprintf("sso:%s:%s", challenge.ProviderType, uuid.New().String()[:8])

	// Update challenge status
	s.updateChallengeStatus(req.ChallengeID, ChallengeStatusCompleted)

	s.logAudit(ctx, audit.EventTypeVerificationCompleted, challenge.AccountAddress, "complete_verification", map[string]interface{}{
		"challenge_id":  req.ChallengeID,
		"linkage_id":    linkageID,
		"provider":      challenge.ProviderType,
		"subject_hash":  attestation.SubjectHash,
		"email_domain":  claims.GetEmailDomain(),
	})

	return NewCompleteSuccess(attestation, linkageID), nil
}

// ExchangeCodeAndComplete exchanges an authorization code and completes verification.
func (s *DefaultService) ExchangeCodeAndComplete(ctx context.Context, req *CodeExchangeCompleteRequest) (*CompleteResponse, error) {
	if err := req.Validate(); err != nil {
		return NewCompleteError("validation_error", err.Error()), nil
	}

	// Get challenge
	challenge, err := s.GetChallenge(ctx, req.ChallengeID)
	if err != nil {
		return NewCompleteError("challenge_error", err.Error()), nil
	}

	// Validate state
	if challenge.State != req.State {
		return NewCompleteError("state_mismatch", "OAuth state parameter mismatch"), nil
	}

	// Check if challenge can be completed
	now := time.Now()
	if !challenge.CanComplete(now) {
		if challenge.IsExpired(now) {
			s.updateChallengeStatus(req.ChallengeID, ChallengeStatusExpired)
			return NewCompleteError("challenge_expired", "SSO challenge has expired"), nil
		}
		return NewCompleteError("challenge_invalid", fmt.Sprintf("challenge status: %s", challenge.Status)), nil
	}

	// Get issuer policy
	policy, err := s.oidcVerifier.GetIssuerPolicy(ctx, challenge.OIDCIssuer)
	if err != nil {
		return NewCompleteError("config_error", err.Error()), nil
	}

	// Exchange code for tokens
	tokenResp, err := s.oidcVerifier.ExchangeCode(ctx, req.AuthorizationCode, &oidc.CodeExchangeRequest{
		ProviderType:  challenge.ProviderType,
		Issuer:        challenge.OIDCIssuer,
		ClientID:      policy.ClientID,
		ClientSecret:  "", // Would need to get from secure storage
		RedirectURI:   challenge.RedirectURI,
		Code:          req.AuthorizationCode,
		ExpectedNonce: challenge.Nonce,
	})
	if err != nil {
		s.updateChallengeStatus(req.ChallengeID, ChallengeStatusFailed)
		return NewCompleteError("code_exchange_failed", err.Error()), nil
	}

	// Create attestation from verified claims
	attestation, err := s.createAttestation(ctx, challenge, tokenResp.VerifiedClaims, req.LinkageSignature)
	if err != nil {
		s.updateChallengeStatus(req.ChallengeID, ChallengeStatusFailed)
		return NewCompleteError("attestation_failed", err.Error()), nil
	}

	// Sign attestation
	if err := s.signer.SignAttestation(ctx, &attestation.VerificationAttestation); err != nil {
		s.updateChallengeStatus(req.ChallengeID, ChallengeStatusFailed)
		return NewCompleteError("signing_failed", err.Error()), nil
	}

	// Generate linkage ID
	linkageID := fmt.Sprintf("sso:%s:%s", challenge.ProviderType, uuid.New().String()[:8])

	// Update challenge status
	s.updateChallengeStatus(req.ChallengeID, ChallengeStatusCompleted)

	s.logAudit(ctx, audit.EventTypeVerificationCompleted, challenge.AccountAddress, "complete_verification", map[string]interface{}{
		"challenge_id": req.ChallengeID,
		"linkage_id":   linkageID,
		"provider":     challenge.ProviderType,
	})

	return NewCompleteSuccess(attestation, linkageID), nil
}

// createAttestation creates an SSO attestation from verified claims.
func (s *DefaultService) createAttestation(
	ctx context.Context,
	challenge *Challenge,
	claims *oidc.VerifiedClaims,
	linkageSignature []byte,
) (*veidtypes.SSOAttestation, error) {
	// Generate attestation nonce
	nonceBytes := make([]byte, 32)
	if _, err := rand.Read(nonceBytes); err != nil {
		return nil, fmt.Errorf("%w: failed to generate nonce: %v", ErrAttestationFailed, err)
	}

	// Get signing key info
	keyInfo, err := s.signer.GetActiveKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get signing key: %v", ErrSigningFailed, err)
	}

	now := time.Now()
	validityDuration := time.Duration(s.config.AttestationValidityDays) * 24 * time.Hour

	issuer := veidtypes.NewAttestationIssuer(keyInfo.Fingerprint, "")
	subject := veidtypes.NewAttestationSubject(challenge.AccountAddress)

	attestation := veidtypes.NewSSOAttestation(
		issuer,
		subject,
		challenge.OIDCIssuer,
		claims.Subject,
		challenge.ProviderType,
		challenge.Nonce,
		nonceBytes,
		now,
		validityDuration,
	)

	// Set email info
	if claims.Email != "" {
		attestation.SetEmail(claims.Email, claims.GetEmailDomain(), claims.EmailVerified)
	}

	// Set tenant ID (for Microsoft)
	if claims.TenantID != "" {
		attestation.SetTenantID(claims.TenantID)
	}

	// Set auth context
	attestation.SetAuthContext(claims.AuthTime, []string{claims.ACR}, claims.AMR)

	// Set linkage signature
	attestation.SetLinkageSignature(linkageSignature)

	// Add verification proof
	proof := veidtypes.NewVerificationProofDetail(
		"oidc_token_verified",
		veidtypes.HashSubjectID(claims.Subject),
		100, // Full score for valid token
		100,
		now,
	)
	attestation.AddVerificationProof(proof)

	// Set metadata
	attestation.SetMetadata("provider_type", string(challenge.ProviderType))
	attestation.SetMetadata("email_domain_hash", attestation.EmailDomainHash)
	if claims.TenantID != "" {
		attestation.SetMetadata("tenant_id_hash", attestation.TenantIDHash)
	}

	return attestation, nil
}

// GetChallenge retrieves a pending verification challenge.
func (s *DefaultService) GetChallenge(ctx context.Context, challengeID string) (*Challenge, error) {
	s.mu.RLock()
	challenge, ok := s.challenges[challengeID]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%w: challenge ID: %s", ErrChallengeNotFound, challengeID)
	}

	// Make a copy
	challengeCopy := *challenge
	return &challengeCopy, nil
}

// updateChallengeStatus updates the status of a challenge.
func (s *DefaultService) updateChallengeStatus(challengeID string, status ChallengeStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if challenge, ok := s.challenges[challengeID]; ok {
		challenge.Status = status
		if status == ChallengeStatusCompleted || status == ChallengeStatusFailed {
			now := time.Now()
			challenge.CompletedAt = &now
		}
	}
}

// RevokeVerification revokes an existing SSO verification.
func (s *DefaultService) RevokeVerification(ctx context.Context, req *RevokeRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	// TODO: Implement revocation logic with on-chain submission
	// For now, just log the revocation request

	s.logAudit(ctx, audit.EventTypeKeyRevoked, req.AccountAddress, "revoke_verification", map[string]interface{}{
		"linkage_id": req.LinkageID,
		"reason":     req.Reason,
	})

	return nil
}

// GetLinkageStatus returns the status of an SSO linkage.
func (s *DefaultService) GetLinkageStatus(ctx context.Context, accountAddress string) (*LinkageStatus, error) {
	// TODO: Query on-chain linkage status
	// For now, return not found

	return &LinkageStatus{
		Exists:         false,
		AccountAddress: accountAddress,
	}, nil
}

// HealthCheck returns the service health status.
func (s *DefaultService) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	status := &HealthStatus{
		Healthy:    true,
		Status:     "healthy",
		Timestamp:  time.Now(),
		Components: make(map[string]*ComponentHealth),
		Details:    make(map[string]interface{}),
		Warnings:   make([]string, 0),
	}

	// Check OIDC verifier
	oidcHealth, err := s.oidcVerifier.HealthCheck(ctx)
	if err != nil {
		status.Components["oidc_verifier"] = &ComponentHealth{
			Name:      "oidc_verifier",
			Healthy:   false,
			LastError: err.Error(),
		}
		status.Healthy = false
	} else {
		status.Components["oidc_verifier"] = &ComponentHealth{
			Name:    "oidc_verifier",
			Healthy: oidcHealth.Healthy,
			Status:  oidcHealth.Status,
		}
		if !oidcHealth.Healthy {
			status.Healthy = false
		}
	}

	// Check signer
	signerHealth, err := s.signer.HealthCheck(ctx)
	if err != nil {
		status.Components["signer"] = &ComponentHealth{
			Name:      "signer",
			Healthy:   false,
			LastError: err.Error(),
		}
		status.Healthy = false
	} else {
		status.Components["signer"] = &ComponentHealth{
			Name:    "signer",
			Healthy: signerHealth.Healthy,
			Status:  signerHealth.Status,
		}
		if !signerHealth.Healthy {
			status.Healthy = false
		}
	}

	// Add challenge stats
	s.mu.RLock()
	status.Details["pending_challenges"] = len(s.challenges)
	s.mu.RUnlock()

	if !status.Healthy {
		status.Status = "degraded"
	}

	return status, nil
}

// Close releases resources.
func (s *DefaultService) Close() error {
	s.logger.Info().Msg("SSO verification service closing")
	return nil
}

// cleanupExpiredChallenges periodically removes expired challenges.
func (s *DefaultService) cleanupExpiredChallenges(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.doCleanup()
		}
	}
}

func (s *DefaultService) doCleanup() {
	now := time.Now()
	toDelete := make([]string, 0)

	s.mu.Lock()
	defer s.mu.Unlock()

	for id, challenge := range s.challenges {
		// Remove challenges that are expired and older than 1 hour
		if challenge.IsExpired(now) && now.Sub(challenge.ExpiresAt) > time.Hour {
			toDelete = append(toDelete, id)
		}
		// Remove completed challenges older than 24 hours
		if challenge.CompletedAt != nil && now.Sub(*challenge.CompletedAt) > 24*time.Hour {
			toDelete = append(toDelete, id)
		}
	}

	for _, id := range toDelete {
		challenge := s.challenges[id]
		delete(s.challenges, id)

		// Remove from byAccount index
		if ids, ok := s.byAccount[challenge.AccountAddress]; ok {
			newIDs := make([]string, 0, len(ids)-1)
			for _, cid := range ids {
				if cid != id {
					newIDs = append(newIDs, cid)
				}
			}
			if len(newIDs) == 0 {
				delete(s.byAccount, challenge.AccountAddress)
			} else {
				s.byAccount[challenge.AccountAddress] = newIDs
			}
		}
	}

	if len(toDelete) > 0 {
		s.logger.Debug().Int("count", len(toDelete)).Msg("cleaned up expired challenges")
	}
}

// logAudit logs an audit event.
func (s *DefaultService) logAudit(ctx context.Context, eventType audit.EventType, resource, action string, details map[string]interface{}) {
	if s.auditor == nil || !s.config.AuditEnabled {
		return
	}

	s.auditor.Log(ctx, audit.Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Actor:     "sso_service",
		Resource:  resource,
		Action:    action,
		Details:   details,
	})
}

// generateSecureToken generates a secure random token.
func generateSecureToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}
