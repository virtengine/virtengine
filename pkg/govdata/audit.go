// Package govdata provides government data source integration for identity verification.
//
// VE-909: Government data integration for authoritative identity verification
package govdata

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// Audit Logger Implementation
// ============================================================================

// auditLogger implements the AuditLogger interface
type auditLogger struct {
	config  AuditConfig
	entries map[string]*AuditLogEntry
	mu      sync.RWMutex
}

// newAuditLogger creates a new audit logger
func newAuditLogger(config AuditConfig) AuditLogger {
	return &auditLogger{
		config:  config,
		entries: make(map[string]*AuditLogEntry),
	}
}

// Log logs an audit entry
func (a *auditLogger) Log(ctx context.Context, entry *AuditLogEntry) error {
	if !a.config.Enabled {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if entry.ID == "" {
		entry.ID = generateAuditID()
	}

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	if entry.RetentionExpiresAt.IsZero() {
		entry.RetentionExpiresAt = time.Now().AddDate(0, 0, a.config.RetentionDays)
	}

	// In production, this would:
	// 1. Encrypt the entry if config.EncryptLogs is true
	// 2. Write to secure audit log file/system
	// 3. Send alerts if config.AlertOnFailure is true and entry indicates failure

	a.entries[entry.ID] = entry

	return nil
}

// Get retrieves an audit log entry
func (a *auditLogger) Get(ctx context.Context, auditID string) (*AuditLogEntry, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	entry, ok := a.entries[auditID]
	if !ok {
		return nil, fmt.Errorf("audit log entry not found: %s", auditID)
	}

	return entry, nil
}

// List lists audit log entries with filtering
func (a *auditLogger) List(ctx context.Context, filter AuditLogFilter) ([]AuditLogEntry, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var results []AuditLogEntry

	for _, entry := range a.entries {
		if a.matchesFilter(entry, filter) {
			results = append(results, *entry)
		}
	}

	// Apply pagination
	if filter.Offset > 0 && filter.Offset < len(results) {
		results = results[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < len(results) {
		results = results[:filter.Limit]
	}

	return results, nil
}

// matchesFilter checks if an entry matches the filter
func (a *auditLogger) matchesFilter(entry *AuditLogEntry, filter AuditLogFilter) bool {
	if filter.WalletAddress != "" && entry.WalletAddress != filter.WalletAddress {
		return false
	}
	if filter.Jurisdiction != "" && entry.Jurisdiction != filter.Jurisdiction {
		return false
	}
	if filter.Action != nil && entry.Action != *filter.Action {
		return false
	}
	if filter.Status != nil && entry.Status != *filter.Status {
		return false
	}
	if !filter.Since.IsZero() && entry.Timestamp.Before(filter.Since) {
		return false
	}
	if !filter.Until.IsZero() && entry.Timestamp.After(filter.Until) {
		return false
	}
	return true
}

// Export exports audit logs
func (a *auditLogger) Export(ctx context.Context, filter AuditLogFilter, format string) ([]byte, error) {
	entries, err := a.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	switch format {
	case "json":
		return json.MarshalIndent(entries, "", "  ")
	case "csv":
		return a.exportCSV(entries)
	default:
		return json.MarshalIndent(entries, "", "  ")
	}
}

// exportCSV exports entries as CSV
func (a *auditLogger) exportCSV(entries []AuditLogEntry) ([]byte, error) {
	var result string
	result = "id,request_id,action,wallet_address,jurisdiction,document_type,data_source,status,timestamp,duration\n"

	for _, e := range entries {
		result += fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			e.ID,
			e.RequestID,
			e.Action,
			e.WalletAddress,
			e.Jurisdiction,
			e.DocumentType,
			e.DataSource,
			e.Status,
			e.Timestamp.Format(time.RFC3339),
			e.Duration.String(),
		)
	}

	return []byte(result), nil
}

// Purge purges old audit logs
func (a *auditLogger) Purge(ctx context.Context, before time.Time) (int, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	count := 0
	for id, entry := range a.entries {
		if entry.RetentionExpiresAt.Before(before) || entry.Timestamp.Before(before) {
			delete(a.entries, id)
			count++
		}
	}

	return count, nil
}

// Count counts audit log entries
func (a *auditLogger) Count(ctx context.Context, filter AuditLogFilter) (int64, error) {
	entries, err := a.List(ctx, filter)
	if err != nil {
		return 0, err
	}
	return int64(len(entries)), nil
}

// ============================================================================
// Consent Manager Implementation
// ============================================================================

// consentManager implements the ConsentManager interface
type consentManager struct {
	config   Config
	consents map[string]*Consent
	mu       sync.RWMutex
}

// newConsentManager creates a new consent manager
func newConsentManager(config Config) ConsentManager {
	return &consentManager{
		config:   config,
		consents: make(map[string]*Consent),
	}
}

// Grant grants consent
func (m *consentManager) Grant(ctx context.Context, consent *Consent) (*Consent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if consent.ID == "" {
		consent.ID = generateConsentID()
	}

	now := time.Now()
	consent.GrantedAt = now
	consent.Active = true

	if consent.ExpiresAt.IsZero() {
		consent.ExpiresAt = now.Add(m.config.ConsentDuration)
	}

	m.consents[consent.ID] = consent

	return consent, nil
}

// Revoke revokes consent
func (m *consentManager) Revoke(ctx context.Context, consentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	consent, ok := m.consents[consentID]
	if !ok {
		return fmt.Errorf("consent not found: %s", consentID)
	}

	now := time.Now()
	consent.RevokedAt = &now
	consent.Active = false

	return nil
}

// Get retrieves consent by ID
func (m *consentManager) Get(ctx context.Context, consentID string) (*Consent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	consent, ok := m.consents[consentID]
	if !ok {
		return nil, fmt.Errorf("consent not found: %s", consentID)
	}

	return consent, nil
}

// List lists consents for a wallet
func (m *consentManager) List(ctx context.Context, walletAddress string) ([]Consent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []Consent
	for _, consent := range m.consents {
		if consent.WalletAddress == walletAddress {
			results = append(results, *consent)
		}
	}

	return results, nil
}

// Validate validates consent for a request
func (m *consentManager) Validate(ctx context.Context, consentID string, req *VerificationRequest) error {
	consent, err := m.Get(ctx, consentID)
	if err != nil {
		return ErrConsentRequired
	}

	if !consent.IsValid() {
		if consent.RevokedAt != nil {
			return ErrConsentRequired
		}
		if time.Now().After(consent.ExpiresAt) {
			return ErrConsentExpired
		}
		return ErrConsentRequired
	}

	// Check if consent covers this wallet
	if consent.WalletAddress != req.WalletAddress {
		return ErrConsentRequired
	}

	// Check if consent covers this document type
	docTypeAllowed := false
	for _, dt := range consent.DocumentTypes {
		if dt == req.DocumentType {
			docTypeAllowed = true
			break
		}
	}
	if !docTypeAllowed && len(consent.DocumentTypes) > 0 {
		return ErrConsentRequired
	}

	// Check if consent covers this jurisdiction
	jurisdictionAllowed := false
	for _, j := range consent.Jurisdictions {
		if j == req.Jurisdiction || j == req.Jurisdiction[:2] {
			jurisdictionAllowed = true
			break
		}
	}
	if !jurisdictionAllowed && len(consent.Jurisdictions) > 0 {
		return ErrConsentRequired
	}

	return nil
}

// IsConsentValid checks if consent is valid
func (m *consentManager) IsConsentValid(ctx context.Context, consentID string) bool {
	consent, err := m.Get(ctx, consentID)
	if err != nil {
		return false
	}
	return consent.IsValid()
}

// CleanupExpired removes expired consent records
func (m *consentManager) CleanupExpired(ctx context.Context) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	count := 0
	retentionDuration := time.Duration(m.config.DefaultRetention.ConsentRetentionDays) * 24 * time.Hour

	for id, consent := range m.consents {
		// Keep revoked consents for retention period
		if consent.RevokedAt != nil {
			if consent.RevokedAt.Add(retentionDuration).Before(now) {
				delete(m.consents, id)
				count++
			}
			continue
		}

		// Keep expired consents for retention period
		if consent.ExpiresAt.Add(retentionDuration).Before(now) {
			delete(m.consents, id)
			count++
		}
	}

	return count, nil
}

// ============================================================================
// Verification Store Implementation
// ============================================================================

// verificationStore implements the VerificationStore interface
type verificationStore struct {
	retention     RetentionPolicy
	verifications map[string]*VerificationResponse
	walletIndex   map[string][]string // walletAddress -> requestIDs
	mu            sync.RWMutex
}

// newVerificationStore creates a new verification store
func newVerificationStore(retention RetentionPolicy) VerificationStore {
	return &verificationStore{
		retention:     retention,
		verifications: make(map[string]*VerificationResponse),
		walletIndex:   make(map[string][]string),
	}
}

// Store stores a verification result
func (s *verificationStore) Store(ctx context.Context, result *VerificationResponse) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.verifications[result.RequestID] = result

	// We need wallet address from the request, which isn't stored in response
	// In production, this would be handled by a proper database schema

	return nil
}

// Get retrieves a verification by request ID
func (s *verificationStore) Get(ctx context.Context, requestID string) (*VerificationResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result, ok := s.verifications[requestID]
	if !ok {
		return nil, fmt.Errorf("verification not found: %s", requestID)
	}

	return result, nil
}

// List lists verifications for a wallet
func (s *verificationStore) List(ctx context.Context, walletAddress string, opts ListOptions) ([]VerificationResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	requestIDs, ok := s.walletIndex[walletAddress]
	if !ok {
		return []VerificationResponse{}, nil
	}

	var results []VerificationResponse
	for _, id := range requestIDs {
		if result, ok := s.verifications[id]; ok {
			// Apply filters
			if opts.Status != nil && result.Status != *opts.Status {
				continue
			}
			if opts.Since != nil && result.VerifiedAt.Before(*opts.Since) {
				continue
			}
			if opts.Until != nil && result.VerifiedAt.After(*opts.Until) {
				continue
			}
			results = append(results, *result)
		}
	}

	// Apply pagination
	if opts.Offset > 0 && opts.Offset < len(results) {
		results = results[opts.Offset:]
	}
	if opts.Limit > 0 && opts.Limit < len(results) {
		results = results[:opts.Limit]
	}

	return results, nil
}

// Delete deletes a verification result
func (s *verificationStore) Delete(ctx context.Context, requestID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.verifications, requestID)
	return nil
}

// Purge purges expired verification results
func (s *verificationStore) Purge(ctx context.Context, before time.Time) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for id, result := range s.verifications {
		if result.ExpiresAt.Before(before) || result.VerifiedAt.Before(before) {
			delete(s.verifications, id)
			count++
		}
	}

	return count, nil
}

// ============================================================================
// Jurisdiction Registry Implementation
// ============================================================================

// jurisdictionRegistry implements the JurisdictionRegistry interface
type jurisdictionRegistry struct {
	jurisdictions map[string]*Jurisdiction
	mu            sync.RWMutex
}

// newJurisdictionRegistry creates a new jurisdiction registry
func newJurisdictionRegistry(config JurisdictionConfig) JurisdictionRegistry {
	r := &jurisdictionRegistry{
		jurisdictions: make(map[string]*Jurisdiction),
	}

	for code, j := range config.Jurisdictions {
		jCopy := j
		r.jurisdictions[code] = &jCopy
	}

	return r
}

// Get retrieves a jurisdiction by code
func (r *jurisdictionRegistry) Get(ctx context.Context, code string) (*Jurisdiction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	j, ok := r.jurisdictions[code]
	if !ok {
		return nil, ErrJurisdictionNotSupported
	}

	return j, nil
}

// List lists all jurisdictions
func (r *jurisdictionRegistry) List(ctx context.Context) ([]Jurisdiction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []Jurisdiction
	for _, j := range r.jurisdictions {
		results = append(results, *j)
	}

	return results, nil
}

// IsSupported checks if a jurisdiction is supported
func (r *jurisdictionRegistry) IsSupported(ctx context.Context, code string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	j, ok := r.jurisdictions[code]
	if !ok {
		// Try parent jurisdiction
		if len(code) > 2 {
			j, ok = r.jurisdictions[code[:2]]
		}
	}

	return ok && j.Active
}

// Register registers a new jurisdiction
func (r *jurisdictionRegistry) Register(ctx context.Context, jurisdiction *Jurisdiction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.jurisdictions[jurisdiction.Code] = jurisdiction
	return nil
}

// Update updates a jurisdiction
func (r *jurisdictionRegistry) Update(ctx context.Context, jurisdiction *Jurisdiction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.jurisdictions[jurisdiction.Code]; !ok {
		return ErrJurisdictionNotSupported
	}

	r.jurisdictions[jurisdiction.Code] = jurisdiction
	return nil
}

// Disable disables a jurisdiction
func (r *jurisdictionRegistry) Disable(ctx context.Context, code string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	j, ok := r.jurisdictions[code]
	if !ok {
		return ErrJurisdictionNotSupported
	}

	j.Active = false
	return nil
}

// GetSupportedDocuments returns supported documents for jurisdiction
func (r *jurisdictionRegistry) GetSupportedDocuments(ctx context.Context, code string) ([]DocumentType, error) {
	j, err := r.Get(ctx, code)
	if err != nil {
		return nil, err
	}
	return j.SupportedDocuments, nil
}

// GetDataSources returns data sources for jurisdiction
func (r *jurisdictionRegistry) GetDataSources(ctx context.Context, code string) ([]DataSourceType, error) {
	j, err := r.Get(ctx, code)
	if err != nil {
		return nil, err
	}
	return j.DataSources, nil
}

// ============================================================================
// Rate Limiter Implementation
// ============================================================================

// rateLimiter implements the RateLimiter interface
type rateLimiter struct {
	config  RateLimitConfig
	buckets map[string]*rateBucket
	mu      sync.RWMutex
}

type rateBucket struct {
	minuteCount  int
	hourCount    int
	dayCount     int
	minuteReset  time.Time
	hourReset    time.Time
	dayReset     time.Time
}

// newRateLimiter creates a new rate limiter
func newRateLimiter(config RateLimitConfig) RateLimiter {
	return &rateLimiter{
		config:  config,
		buckets: make(map[string]*rateBucket),
	}
}

// Allow checks if a request is allowed
func (r *rateLimiter) Allow(ctx context.Context, walletAddress string) (bool, error) {
	if !r.config.Enabled {
		return true, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	bucket, ok := r.buckets[walletAddress]
	if !ok {
		bucket = &rateBucket{
			minuteReset: time.Now().Add(time.Minute),
			hourReset:   time.Now().Add(time.Hour),
			dayReset:    time.Now().Add(24 * time.Hour),
		}
		r.buckets[walletAddress] = bucket
	}

	now := time.Now()

	// Reset counters if needed
	if now.After(bucket.minuteReset) {
		bucket.minuteCount = 0
		bucket.minuteReset = now.Add(time.Minute)
	}
	if now.After(bucket.hourReset) {
		bucket.hourCount = 0
		bucket.hourReset = now.Add(time.Hour)
	}
	if now.After(bucket.dayReset) {
		bucket.dayCount = 0
		bucket.dayReset = now.Add(24 * time.Hour)
	}

	// Check limits
	if bucket.minuteCount >= r.config.RequestsPerMinute {
		return false, nil
	}
	if bucket.hourCount >= r.config.RequestsPerHour {
		return false, nil
	}
	if bucket.dayCount >= r.config.RequestsPerDay {
		return false, nil
	}

	// Increment counters
	bucket.minuteCount++
	bucket.hourCount++
	bucket.dayCount++

	return true, nil
}

// GetRemaining returns remaining requests
func (r *rateLimiter) GetRemaining(ctx context.Context, walletAddress string) (RateLimitInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	bucket, ok := r.buckets[walletAddress]
	if !ok {
		return RateLimitInfo{
			RemainingMinute: r.config.RequestsPerMinute,
			RemainingHour:   r.config.RequestsPerHour,
			RemainingDay:    r.config.RequestsPerDay,
			ResetsAt:        time.Now().Add(time.Minute),
		}, nil
	}

	return RateLimitInfo{
		RemainingMinute: r.config.RequestsPerMinute - bucket.minuteCount,
		RemainingHour:   r.config.RequestsPerHour - bucket.hourCount,
		RemainingDay:    r.config.RequestsPerDay - bucket.dayCount,
		ResetsAt:        bucket.minuteReset,
	}, nil
}

// Reset resets rate limits for a wallet
func (r *rateLimiter) Reset(ctx context.Context, walletAddress string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.buckets, walletAddress)
	return nil
}
