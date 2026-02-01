// Package edugain provides EduGAIN federation integration.
//
// VE-908: EduGAIN federation support for academic/research institution SSO
package edugain

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// Service Implementation
// ============================================================================

// service implements the Service interface
type service struct {
	config           Config
	metadata         MetadataService
	saml             SAMLProvider
	sessions         SessionManager
	discovery        DiscoveryService
	veid             VEIDIntegrator
	attributeMapper  AttributeMapper
	status           ServiceStatus
	refreshTicker    *time.Ticker
	stopChan         chan struct{}
	running          bool
	mu sync.RWMutex
	//nolint:unused // Reserved for metadata refresh callback functionality
	metadataCallback MetadataRefreshCallback
}

// NewService creates a new EduGAIN service
func NewService(config Config) (Service, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	s := &service{
		config:   config,
		stopChan: make(chan struct{}),
		status: ServiceStatus{
			Version: EduGAINVersion,
		},
	}

	// Initialize metadata service
	s.metadata = newMetadataService(config)

	// Initialize SAML provider
	s.saml = newSAMLProvider(config)

	// Initialize session manager
	s.sessions = newSessionManager(config)

	// Initialize discovery service
	s.discovery = newDiscoveryService(s.metadata)

	// Initialize attribute mapper
	s.attributeMapper = newAttributeMapper()

	// Initialize VEID integrator (if enabled)
	if config.VEIDIntegration.Enabled {
		s.veid = newVEIDIntegrator(config)
	}

	return s, nil
}

// Start starts the service
func (s *service) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	// Load initial metadata
	if err := s.metadata.Load(ctx); err != nil {
		// Log warning but don't fail - metadata can be loaded later
		s.status.LastError = fmt.Sprintf("failed to load initial metadata: %v", err)
	}

	// Start metadata refresh ticker
	s.refreshTicker = time.NewTicker(s.config.MetadataRefreshInterval)
	go s.metadataRefreshLoop(ctx)

	// Start session cleanup ticker
	go s.sessionCleanupLoop(ctx)

	s.running = true
	s.status.StartedAt = time.Now()
	s.status.Healthy = true

	return nil
}

// Stop stops the service
func (s *service) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	// Stop tickers
	if s.refreshTicker != nil {
		s.refreshTicker.Stop()
	}

	// Signal goroutines to stop
	close(s.stopChan)

	s.running = false
	s.status.Healthy = false

	return nil
}

// IsHealthy returns true if the service is healthy
func (s *service) IsHealthy() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status.Healthy && s.metadata.IsValid()
}

// GetStatus returns the service status
func (s *service) GetStatus() ServiceStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := s.status
	status.Uptime = time.Since(s.status.StartedAt)
	status.Metadata = s.metadata.GetStatus()

	if stats, err := s.sessions.GetStats(context.Background()); err == nil {
		status.Sessions = stats
	}

	return status
}

// RefreshMetadata forces a metadata refresh
func (s *service) RefreshMetadata(ctx context.Context) error {
	return s.metadata.ForceRefresh(ctx)
}

// GetMetadata returns the current federation metadata
func (s *service) GetMetadata() (*FederationMetadata, error) {
	return s.metadata.Get()
}

// GetMetadataStatus returns the metadata status
func (s *service) GetMetadataStatus() MetadataStatus {
	return s.metadata.GetStatus()
}

// DiscoverInstitutions searches for institutions
func (s *service) DiscoverInstitutions(ctx context.Context, query InstitutionSearchQuery) (*InstitutionSearchResult, error) {
	return s.discovery.Search(ctx, query)
}

// GetInstitution returns a specific institution by entity ID
func (s *service) GetInstitution(ctx context.Context, entityID string) (*Institution, error) {
	return s.discovery.GetByEntityID(ctx, entityID)
}

// GetFederationStats returns federation statistics
func (s *service) GetFederationStats(ctx context.Context) (*FederationStats, error) {
	return s.metadata.GetStats(), nil
}

// CreateAuthnRequest creates a SAML AuthnRequest
func (s *service) CreateAuthnRequest(ctx context.Context, params AuthnRequestParams) (*AuthnRequestResult, error) {
	// Get institution
	idp, err := s.metadata.FindInstitution(params.InstitutionID)
	if err != nil {
		return nil, err
	}

	// Check if institution is allowed
	if !s.config.IsInstitutionAllowed(params.InstitutionID) {
		return nil, fmt.Errorf("institution not allowed: %s", params.InstitutionID)
	}

	// Check if MFA is required
	if s.config.RequireMFA {
		params.RequireMFA = true
	}

	// Create AuthnRequest
	return s.saml.CreateAuthnRequest(ctx, idp, params)
}

// VerifyResponse verifies a SAML response and returns the assertion
func (s *service) VerifyResponse(ctx context.Context, samlResponseBase64 string) (*SAMLAssertion, error) {
	// Parse response to get issuer
	issuerEntityID, err := s.extractIssuerFromResponse(samlResponseBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to extract issuer: %w", err)
	}

	// Get IdP metadata
	idp, err := s.metadata.FindInstitution(issuerEntityID)
	if err != nil {
		return nil, err
	}

	// Check if institution is allowed
	if !s.config.IsInstitutionAllowed(issuerEntityID) {
		return nil, fmt.Errorf("institution not allowed: %s", issuerEntityID)
	}

	// Verify response
	assertion, err := s.saml.VerifyResponse(ctx, samlResponseBase64, idp)
	if err != nil {
		return nil, err
	}

	// Check for replay
	replayed, err := s.sessions.IsAssertionReplayed(ctx, assertion.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check replay: %w", err)
	}
	if replayed {
		return nil, ErrReplayDetected
	}

	// Track assertion ID
	if err := s.sessions.TrackAssertionID(ctx, assertion.ID, assertion.NotOnOrAfter.Add(s.config.ReplayWindowDuration)); err != nil {
		return nil, fmt.Errorf("failed to track assertion: %w", err)
	}

	// Validate assertion timing
	if err := assertion.Validate(s.config.ClockSkew); err != nil {
		return nil, err
	}

	// Check MFA requirement
	if s.config.RequireMFA && !assertion.IsMFA {
		return nil, ErrMFARequired
	}

	// Map and validate attributes
	attrs, err := s.attributeMapper.MapAttributes(assertion.Attributes.Raw)
	if err != nil {
		return nil, fmt.Errorf("failed to map attributes: %w", err)
	}

	if err := s.attributeMapper.ValidateAttributes(attrs); err != nil {
		return nil, err
	}

	// Check affiliation restrictions
	if len(s.config.AllowedAffiliations) > 0 {
		hasAllowedAffiliation := false
		for _, aff := range attrs.EduPerson.Affiliation {
			if s.config.IsAffiliationAllowed(aff) {
				hasAllowedAffiliation = true
				break
			}
		}
		if !hasAllowedAffiliation {
			return nil, fmt.Errorf("no allowed affiliation found")
		}
	}

	// Update assertion with mapped attributes
	assertion.Attributes = *attrs

	// Hash sensitive data if configured
	if s.config.VEIDIntegration.HashSensitiveData {
		assertion.Attributes = *s.attributeMapper.HashSensitiveData(&assertion.Attributes)
	}

	return assertion, nil
}

// CreateLogoutRequest creates a SAML LogoutRequest
func (s *service) CreateLogoutRequest(ctx context.Context, sessionID string) (*LogoutRequestResult, error) {
	// Get session
	session, err := s.sessions.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Get IdP metadata
	idp, err := s.metadata.FindInstitution(session.InstitutionID)
	if err != nil {
		return nil, err
	}

	return s.saml.CreateLogoutRequest(ctx, session, idp)
}

// CreateSession creates a session from a verified assertion
func (s *service) CreateSession(ctx context.Context, assertion *SAMLAssertion, walletAddress string) (*Session, error) {
	session, err := s.sessions.Create(ctx, assertion, walletAddress)
	if err != nil {
		return nil, err
	}

	// Record institution usage for discovery
	if err := s.discovery.RecordUsage(ctx, assertion.IssuerEntityID, walletAddress); err != nil {
		// Log but don't fail
		s.status.LastError = fmt.Sprintf("failed to record usage: %v", err)
	}

	return session, nil
}

// GetSession returns a session by ID
func (s *service) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	return s.sessions.Get(ctx, sessionID)
}

// ValidateSession validates a session token
func (s *service) ValidateSession(ctx context.Context, sessionToken string) (*Session, error) {
	return s.sessions.ValidateToken(ctx, sessionToken)
}

// RevokeSession revokes a session
func (s *service) RevokeSession(ctx context.Context, sessionID string) error {
	return s.sessions.Revoke(ctx, sessionID)
}

// ListSessions lists sessions for a wallet
func (s *service) ListSessions(ctx context.Context, walletAddress string) ([]Session, error) {
	return s.sessions.List(ctx, walletAddress)
}

// CreateVEIDScope creates a VEID scope from a session
func (s *service) CreateVEIDScope(ctx context.Context, session *Session) (*VEIDScope, error) {
	if s.veid == nil {
		return nil, ErrVEIDIntegrationFailed
	}
	return s.veid.CreateScope(ctx, session)
}

// EnrichVEID enriches an existing VEID with EduGAIN attributes
func (s *service) EnrichVEID(ctx context.Context, session *Session, veidID string) error {
	if s.veid == nil {
		return ErrVEIDIntegrationFailed
	}
	return s.veid.EnrichIdentity(ctx, session, veidID)
}

// extractIssuerFromResponse extracts the issuer entity ID from a SAML response
func (s *service) extractIssuerFromResponse(samlResponseBase64 string) (string, error) {
	// This would normally decode and parse the SAML response XML
	// For now, return a placeholder - the actual implementation would use XML parsing
	return extractIssuerEntityID(samlResponseBase64)
}

// metadataRefreshLoop periodically refreshes federation metadata
func (s *service) metadataRefreshLoop(ctx context.Context) {
	for {
		select {
		case <-s.refreshTicker.C:
			if err := s.metadata.Refresh(ctx); err != nil {
				s.mu.Lock()
				s.status.LastError = fmt.Sprintf("metadata refresh failed: %v", err)
				s.mu.Unlock()
			}
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// sessionCleanupLoop periodically cleans up expired sessions
func (s *service) sessionCleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(s.config.SessionStorage.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if _, err := s.sessions.Cleanup(ctx); err != nil {
				s.mu.Lock()
				s.status.LastError = fmt.Sprintf("session cleanup failed: %v", err)
				s.mu.Unlock()
			}
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// extractIssuerEntityID extracts issuer from base64 encoded SAML response
func extractIssuerEntityID(samlResponseBase64 string) (string, error) {
	// Decode base64
	data, err := decodeBase64(samlResponseBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Parse XML to find Issuer element
	// This is a simplified implementation - production would use proper XML parsing
	issuer, err := parseIssuerFromXML(data)
	if err != nil {
		return "", fmt.Errorf("failed to parse issuer: %w", err)
	}

	return issuer, nil
}
