// Package govdata provides government data source integration for identity verification.
//
// VE-909: Government data integration for authoritative identity verification
package govdata

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================================
// Service Implementation
// ============================================================================

// service implements the Service interface
type service struct {
	config        Config
	adapters      map[string]DataSourceAdapter // key: jurisdiction-type
	jurisdictions JurisdictionRegistry
	consents      ConsentManager
	auditLogger   AuditLogger
	verifications VerificationStore
	veid          VEIDIntegrator
	rateLimiter   RateLimiter

	status   ServiceStatus
	running  bool
	stopChan chan struct{}
	mu       sync.RWMutex

	// Metrics
	totalVerifications      int64
	successfulVerifications int64
	failedVerifications     int64
}

// NewService creates a new government data service
func NewService(config Config) (Service, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	s := &service{
		config:   config,
		adapters: make(map[string]DataSourceAdapter),
		stopChan: make(chan struct{}),
		status: ServiceStatus{
			Version:       GovDataVersion,
			AdapterStatus: make(map[string]AdapterStatus),
		},
	}

	// Initialize jurisdiction registry
	s.jurisdictions = newJurisdictionRegistry(DefaultJurisdictions())

	// Initialize consent manager
	s.consents = newConsentManager(config)

	// Initialize audit logger
	s.auditLogger = newAuditLogger(config.Audit)

	// Initialize verification store
	s.verifications = newVerificationStore(config.DefaultRetention)

	// Initialize rate limiter
	s.rateLimiter = newRateLimiter(config.RateLimits)

	// Initialize VEID integrator if enabled
	if config.VEIDIntegration.Enabled {
		s.veid = newVEIDIntegrator(config.VEIDIntegration)
	}

	// Initialize adapters
	for name, adapterConfig := range config.Adapters {
		if !adapterConfig.Enabled {
			continue
		}

		adapter, err := createAdapter(adapterConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create adapter %s: %w", name, err)
		}

		key := fmt.Sprintf("%s-%s", adapterConfig.Jurisdiction, adapterConfig.Type)
		s.adapters[key] = adapter
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

	// Start health check loop
	if s.config.HealthCheck.Enabled {
		go s.healthCheckLoop(ctx)
	}

	// Start cleanup loop
	go s.cleanupLoop(ctx)

	s.running = true
	s.status.StartedAt = time.Now()
	s.status.Healthy = true
	s.status.ActiveAdapters = len(s.adapters)

	return nil
}

// Stop stops the service
func (s *service) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	close(s.stopChan)
	s.running = false
	s.status.Healthy = false

	return nil
}

// IsHealthy returns true if the service is healthy
func (s *service) IsHealthy() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status.Healthy
}

// GetStatus returns the service status
func (s *service) GetStatus() ServiceStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := s.status
	status.Uptime = time.Since(s.status.StartedAt)
	status.TotalVerifications = atomic.LoadInt64(&s.totalVerifications)
	status.SuccessfulVerifications = atomic.LoadInt64(&s.successfulVerifications)
	status.FailedVerifications = atomic.LoadInt64(&s.failedVerifications)

	// Get adapter statuses
	for key, adapter := range s.adapters {
		status.AdapterStatus[key] = adapter.GetStats()
	}

	return status
}

// ============================================================================
// Verification Methods
// ============================================================================

// VerifyDocument verifies a government document
func (s *service) VerifyDocument(ctx context.Context, req *VerificationRequest) (*VerificationResponse, error) {
	if !s.running {
		return nil, ErrServiceNotInitialized
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Generate request ID if not provided
	if req.RequestID == "" {
		req.RequestID = generateRequestID()
	}
	req.RequestedAt = time.Now()

	// Check rate limits
	if s.config.RateLimits.Enabled {
		allowed, err := s.rateLimiter.Allow(ctx, req.WalletAddress)
		if err != nil {
			return nil, fmt.Errorf("rate limit check failed: %w", err)
		}
		if !allowed {
			return s.buildErrorResponse(req, VerificationStatusRateLimited, "rate limit exceeded"), nil
		}
	}

	// Validate consent
	if s.config.RequireConsent {
		if err := s.consents.Validate(ctx, req.ConsentID, req); err != nil {
			return nil, err
		}
	}

	// Check jurisdiction support
	if !s.jurisdictions.IsSupported(ctx, req.Jurisdiction) {
		return nil, ErrJurisdictionNotSupported
	}

	// Find appropriate adapter
	adapter, err := s.findAdapter(req.Jurisdiction, req.DocumentType)
	if err != nil {
		return nil, err
	}

	// Check adapter availability
	if !adapter.IsAvailable(ctx) {
		return nil, ErrAdapterUnavailable
	}

	// Start timing
	startTime := time.Now()

	// Perform verification
	response, err := adapter.Verify(ctx, req)
	duration := time.Since(startTime)

	// Update metrics
	atomic.AddInt64(&s.totalVerifications, 1)
	if err != nil || (response != nil && !response.Status.IsSuccess()) {
		atomic.AddInt64(&s.failedVerifications, 1)
	} else {
		atomic.AddInt64(&s.successfulVerifications, 1)
	}

	// Create audit log entry
	auditEntry := &AuditLogEntry{
		ID:            generateAuditID(),
		RequestID:     req.RequestID,
		Action:        AuditActionVerify,
		WalletAddress: req.WalletAddress,
		Jurisdiction:  req.Jurisdiction,
		DocumentType:  req.DocumentType,
		DataSource:    adapter.Type(),
		ConsentID:     req.ConsentID,
		Timestamp:     time.Now(),
		Duration:      duration,
		RetentionExpiresAt: time.Now().AddDate(0, 0, s.config.DefaultRetention.AuditLogRetentionDays),
	}

	if err != nil {
		auditEntry.Status = VerificationStatusError
		auditEntry.ErrorCode = "VERIFICATION_ERROR"
		if logErr := s.auditLogger.Log(ctx, auditEntry); logErr != nil {
			s.status.LastError = fmt.Sprintf("audit log failed: %v", logErr)
		}
		return nil, err
	}

	auditEntry.Status = response.Status
	response.AuditLogID = auditEntry.ID

	// Log audit entry
	if err := s.auditLogger.Log(ctx, auditEntry); err != nil {
		s.status.LastError = fmt.Sprintf("audit log failed: %v", err)
		// Don't fail the verification for audit log errors
	}

	// Store verification result
	if err := s.verifications.Store(ctx, response); err != nil {
		s.status.LastError = fmt.Sprintf("verification store failed: %v", err)
	}

	return response, nil
}

// VerifyDocumentBatch verifies multiple documents in batch
func (s *service) VerifyDocumentBatch(ctx context.Context, req *BatchVerificationRequest) (*BatchVerificationResponse, error) {
	if !s.running {
		return nil, ErrServiceNotInitialized
	}

	if len(req.Requests) == 0 {
		return nil, ErrInvalidRequest
	}

	startTime := time.Now()
	response := &BatchVerificationResponse{
		BatchID:       req.BatchID,
		Responses:     make([]VerificationResponse, 0, len(req.Requests)),
		TotalRequests: len(req.Requests),
	}

	// Process requests (sequentially for simplicity, could be parallelized)
	for _, verifyReq := range req.Requests {
		reqCopy := verifyReq
		result, err := s.VerifyDocument(ctx, &reqCopy)
		if err != nil {
			if req.FailFast {
				response.FailureCount++
				break
			}
			// Create error response
			errorResult := s.buildErrorResponse(&reqCopy, VerificationStatusError, err.Error())
			response.Responses = append(response.Responses, *errorResult)
			response.FailureCount++
			continue
		}

		response.Responses = append(response.Responses, *result)
		if result.Status.IsSuccess() {
			response.SuccessCount++
		} else {
			response.FailureCount++
		}
	}

	response.Duration = time.Since(startTime)
	return response, nil
}

// GetVerification retrieves a previous verification result
func (s *service) GetVerification(ctx context.Context, requestID string) (*VerificationResponse, error) {
	return s.verifications.Get(ctx, requestID)
}

// ListVerifications lists verifications for a wallet
func (s *service) ListVerifications(ctx context.Context, walletAddress string, opts ListOptions) ([]VerificationResponse, error) {
	return s.verifications.List(ctx, walletAddress, opts)
}

// ============================================================================
// Consent Methods
// ============================================================================

// GrantConsent records user consent
func (s *service) GrantConsent(ctx context.Context, consent *Consent) (*Consent, error) {
	result, err := s.consents.Grant(ctx, consent)
	if err != nil {
		return nil, err
	}

	// Log consent grant
	auditEntry := &AuditLogEntry{
		ID:            generateAuditID(),
		RequestID:     consent.ID,
		Action:        AuditActionConsentGrant,
		WalletAddress: consent.WalletAddress,
		ConsentID:     consent.ID,
		Timestamp:     time.Now(),
		Status:        VerificationStatusVerified,
		RetentionExpiresAt: time.Now().AddDate(0, 0, s.config.DefaultRetention.AuditLogRetentionDays),
	}
	if err := s.auditLogger.Log(ctx, auditEntry); err != nil {
		s.status.LastError = fmt.Sprintf("audit log failed: %v", err)
	}

	return result, nil
}

// RevokeConsent revokes user consent
func (s *service) RevokeConsent(ctx context.Context, consentID string) error {
	consent, err := s.consents.Get(ctx, consentID)
	if err != nil {
		return err
	}

	if err := s.consents.Revoke(ctx, consentID); err != nil {
		return err
	}

	// Log consent revocation
	auditEntry := &AuditLogEntry{
		ID:            generateAuditID(),
		RequestID:     consentID,
		Action:        AuditActionConsentRevoke,
		WalletAddress: consent.WalletAddress,
		ConsentID:     consentID,
		Timestamp:     time.Now(),
		Status:        VerificationStatusVerified,
		RetentionExpiresAt: time.Now().AddDate(0, 0, s.config.DefaultRetention.AuditLogRetentionDays),
	}
	if err := s.auditLogger.Log(ctx, auditEntry); err != nil {
		s.status.LastError = fmt.Sprintf("audit log failed: %v", err)
	}

	return nil
}

// GetConsent retrieves consent by ID
func (s *service) GetConsent(ctx context.Context, consentID string) (*Consent, error) {
	return s.consents.Get(ctx, consentID)
}

// ListConsents lists consents for a wallet
func (s *service) ListConsents(ctx context.Context, walletAddress string) ([]Consent, error) {
	return s.consents.List(ctx, walletAddress)
}

// ValidateConsent validates consent for a request
func (s *service) ValidateConsent(ctx context.Context, consentID string, req *VerificationRequest) error {
	return s.consents.Validate(ctx, consentID, req)
}

// ============================================================================
// Jurisdiction Methods
// ============================================================================

// GetJurisdiction returns a jurisdiction by code
func (s *service) GetJurisdiction(ctx context.Context, code string) (*Jurisdiction, error) {
	return s.jurisdictions.Get(ctx, code)
}

// ListJurisdictions lists all supported jurisdictions
func (s *service) ListJurisdictions(ctx context.Context) ([]Jurisdiction, error) {
	return s.jurisdictions.List(ctx)
}

// IsJurisdictionSupported checks if a jurisdiction is supported
func (s *service) IsJurisdictionSupported(ctx context.Context, code string) bool {
	return s.jurisdictions.IsSupported(ctx, code)
}

// GetSupportedDocuments returns supported documents for a jurisdiction
func (s *service) GetSupportedDocuments(ctx context.Context, jurisdiction string) ([]DocumentType, error) {
	return s.jurisdictions.GetSupportedDocuments(ctx, jurisdiction)
}

// ============================================================================
// Audit Methods
// ============================================================================

// GetAuditLog retrieves an audit log entry
func (s *service) GetAuditLog(ctx context.Context, auditID string) (*AuditLogEntry, error) {
	return s.auditLogger.Get(ctx, auditID)
}

// ListAuditLogs lists audit logs with filtering
func (s *service) ListAuditLogs(ctx context.Context, filter AuditLogFilter) ([]AuditLogEntry, error) {
	return s.auditLogger.List(ctx, filter)
}

// ExportAuditLogs exports audit logs
func (s *service) ExportAuditLogs(ctx context.Context, filter AuditLogFilter) ([]byte, error) {
	return s.auditLogger.Export(ctx, filter, s.config.Audit.ExportFormat)
}

// ============================================================================
// VEID Integration Methods
// ============================================================================

// CreateVEIDScope creates a VEID scope from verification
func (s *service) CreateVEIDScope(ctx context.Context, verification *VerificationResponse) (*VEIDScope, error) {
	if s.veid == nil {
		return nil, ErrVEIDIntegrationFailed
	}
	return s.veid.CreateScope(ctx, verification)
}

// EnrichVEID enriches an existing VEID
func (s *service) EnrichVEID(ctx context.Context, verification *VerificationResponse, veidID string) error {
	if s.veid == nil {
		return ErrVEIDIntegrationFailed
	}
	return s.veid.EnrichIdentity(ctx, verification, veidID)
}

// GetVEIDScopes retrieves VEID scopes for a wallet
func (s *service) GetVEIDScopes(ctx context.Context, walletAddress string) ([]VEIDScope, error) {
	if s.veid == nil {
		return nil, ErrVEIDIntegrationFailed
	}
	return s.veid.GetScopes(ctx, walletAddress)
}

// ============================================================================
// Helper Methods
// ============================================================================

// findAdapter finds an appropriate adapter for the request
func (s *service) findAdapter(jurisdiction string, docType DocumentType) (DataSourceAdapter, error) {
	// Try exact match first
	for _, adapter := range s.adapters {
		if adapter.Jurisdiction() == jurisdiction && adapter.SupportsDocument(docType) {
			return adapter, nil
		}
	}

	// Try parent jurisdiction (e.g., US-CA -> US)
	if len(jurisdiction) > 2 {
		parentJurisdiction := jurisdiction[:2]
		for _, adapter := range s.adapters {
			if adapter.Jurisdiction() == parentJurisdiction && adapter.SupportsDocument(docType) {
				return adapter, nil
			}
		}
	}

	return nil, ErrAdapterNotFound
}

// buildErrorResponse builds an error response
func (s *service) buildErrorResponse(req *VerificationRequest, status VerificationStatus, message string) *VerificationResponse {
	return &VerificationResponse{
		RequestID:    req.RequestID,
		Status:       status,
		Confidence:   0,
		Jurisdiction: req.Jurisdiction,
		VerifiedAt:   time.Now(),
		ExpiresAt:    time.Now(),
		ErrorMessage: message,
	}
}

// healthCheckLoop performs periodic health checks
func (s *service) healthCheckLoop(ctx context.Context) {
	ticker := time.NewTicker(s.config.HealthCheck.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.performHealthChecks(ctx)
		}
	}
}

// performHealthChecks checks all adapter health
func (s *service) performHealthChecks(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	healthyCount := 0
	for key, adapter := range s.adapters {
		if err := adapter.HealthCheck(ctx); err == nil {
			healthyCount++
			s.status.AdapterStatus[key] = adapter.GetStats()
		}
	}

	s.status.Healthy = healthyCount > 0
	s.status.ActiveAdapters = healthyCount
}

// cleanupLoop performs periodic cleanup
func (s *service) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.performCleanup(ctx)
		}
	}
}

// performCleanup cleans up expired data
func (s *service) performCleanup(ctx context.Context) {
	// Cleanup expired consents
	if _, err := s.consents.CleanupExpired(ctx); err != nil {
		s.status.LastError = fmt.Sprintf("consent cleanup failed: %v", err)
	}

	// Cleanup expired verifications
	retentionDays := s.config.DefaultRetention.ResultRetentionDays
	if _, err := s.verifications.Purge(ctx, time.Now().AddDate(0, 0, -retentionDays)); err != nil {
		s.status.LastError = fmt.Sprintf("verification cleanup failed: %v", err)
	}

	// Cleanup expired audit logs
	auditRetentionDays := s.config.DefaultRetention.AuditLogRetentionDays
	if _, err := s.auditLogger.Purge(ctx, time.Now().AddDate(0, 0, -auditRetentionDays)); err != nil {
		s.status.LastError = fmt.Sprintf("audit log cleanup failed: %v", err)
	}
}

// ============================================================================
// ID Generators
// ============================================================================

var requestCounter int64

func generateRequestID() string {
	counter := atomic.AddInt64(&requestCounter, 1)
	return fmt.Sprintf("govdata-req-%d-%d", time.Now().UnixNano(), counter)
}

func generateAuditID() string {
	counter := atomic.AddInt64(&requestCounter, 1)
	return fmt.Sprintf("govdata-audit-%d-%d", time.Now().UnixNano(), counter)
}

func generateConsentID() string {
	counter := atomic.AddInt64(&requestCounter, 1)
	return fmt.Sprintf("govdata-consent-%d-%d", time.Now().UnixNano(), counter)
}

func generateScopeID() string {
	counter := atomic.AddInt64(&requestCounter, 1)
	return fmt.Sprintf("govdata-scope-%d-%d", time.Now().UnixNano(), counter)
}
