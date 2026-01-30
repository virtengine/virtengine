// Package edugain provides EduGAIN federation integration.
//
// VE-2005: Certificate caching for EduGAIN federation metadata
package edugain

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// Certificate Cache
// ============================================================================

// CertificateCacheConfig configures the certificate cache
type CertificateCacheConfig struct {
	// MaxSize is the maximum number of certificates to cache
	MaxSize int

	// TTL is how long to cache certificates
	TTL time.Duration

	// RefreshInterval is how often to check for expired entries
	RefreshInterval time.Duration
}

// DefaultCertificateCacheConfig returns sensible defaults
func DefaultCertificateCacheConfig() CertificateCacheConfig {
	return CertificateCacheConfig{
		MaxSize:         1000,
		TTL:             24 * time.Hour,
		RefreshInterval: 1 * time.Hour,
	}
}

// CachedCertificate represents a cached certificate with metadata
type CachedCertificate struct {
	// Certificate is the parsed X.509 certificate
	Certificate *x509.Certificate

	// EntityID is the IdP entity ID that owns this certificate
	EntityID string

	// Fingerprint is the SHA-256 fingerprint of the certificate
	Fingerprint string

	// CachedAt is when this certificate was cached
	CachedAt time.Time

	// ExpiresAt is when this cache entry expires
	ExpiresAt time.Time

	// Source indicates where this certificate came from
	Source CertificateSource

	// Validated indicates if the certificate has been validated against trust anchors
	Validated bool

	// ValidationError contains any error from validation
	ValidationError error
}

// CertificateSource indicates where a certificate came from
type CertificateSource string

const (
	// CertificateSourceMetadata indicates the certificate came from federation metadata
	CertificateSourceMetadata CertificateSource = "metadata"

	// CertificateSourceSAML indicates the certificate came from a SAML response
	CertificateSourceSAML CertificateSource = "saml"

	// CertificateSourceManual indicates the certificate was manually configured
	CertificateSourceManual CertificateSource = "manual"
)

// CertificateCache provides a thread-safe cache for IdP certificates
type CertificateCache struct {
	config       CertificateCacheConfig
	certificates map[string]*CachedCertificate            // key: fingerprint
	byEntityID   map[string]map[string]*CachedCertificate // key: entityID -> fingerprint -> cert
	trustAnchors []*x509.Certificate
	mu           sync.RWMutex
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

// NewCertificateCache creates a new certificate cache
func NewCertificateCache(config CertificateCacheConfig) *CertificateCache {
	if config.MaxSize <= 0 {
		config.MaxSize = 1000
	}
	if config.TTL <= 0 {
		config.TTL = 24 * time.Hour
	}
	if config.RefreshInterval <= 0 {
		config.RefreshInterval = 1 * time.Hour
	}

	return &CertificateCache{
		config:       config,
		certificates: make(map[string]*CachedCertificate),
		byEntityID:   make(map[string]map[string]*CachedCertificate),
		stopCh:       make(chan struct{}),
	}
}

// Start starts the background cleanup routine
func (c *CertificateCache) Start() {
	c.wg.Add(1)
	go c.cleanupLoop()
}

// Stop stops the background cleanup routine
func (c *CertificateCache) Stop() {
	close(c.stopCh)
	c.wg.Wait()
}

// cleanupLoop periodically removes expired entries
func (c *CertificateCache) cleanupLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.config.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopCh:
			return
		}
	}
}

// cleanup removes expired entries
func (c *CertificateCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for fingerprint, cached := range c.certificates {
		if now.After(cached.ExpiresAt) {
			delete(c.certificates, fingerprint)
			if entityCerts, ok := c.byEntityID[cached.EntityID]; ok {
				delete(entityCerts, fingerprint)
				if len(entityCerts) == 0 {
					delete(c.byEntityID, cached.EntityID)
				}
			}
		}
	}
}

// SetTrustAnchors sets the trust anchors for certificate validation
func (c *CertificateCache) SetTrustAnchors(anchors []*x509.Certificate) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.trustAnchors = anchors
}

// GetTrustAnchors returns the current trust anchors
func (c *CertificateCache) GetTrustAnchors() []*x509.Certificate {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.trustAnchors
}

// Add adds a certificate to the cache
func (c *CertificateCache) Add(cert *x509.Certificate, entityID string, source CertificateSource) (*CachedCertificate, error) {
	if cert == nil {
		return nil, fmt.Errorf("certificate is nil")
	}

	fingerprint := computeCertFingerprint(cert)
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already cached
	if existing, ok := c.certificates[fingerprint]; ok {
		return existing, nil
	}

	// Check capacity
	if len(c.certificates) >= c.config.MaxSize {
		c.evictOldest()
	}

	// Validate against trust anchors if available
	var validationErr error
	validated := false
	if len(c.trustAnchors) > 0 {
		roots := x509.NewCertPool()
		for _, ta := range c.trustAnchors {
			roots.AddCert(ta)
		}
		opts := x509.VerifyOptions{
			Roots:       roots,
			CurrentTime: now,
		}
		if _, err := cert.Verify(opts); err != nil {
			validationErr = err
		} else {
			validated = true
		}
	}

	cached := &CachedCertificate{
		Certificate:     cert,
		EntityID:        entityID,
		Fingerprint:     fingerprint,
		CachedAt:        now,
		ExpiresAt:       now.Add(c.config.TTL),
		Source:          source,
		Validated:       validated,
		ValidationError: validationErr,
	}

	c.certificates[fingerprint] = cached

	if c.byEntityID[entityID] == nil {
		c.byEntityID[entityID] = make(map[string]*CachedCertificate)
	}
	c.byEntityID[entityID][fingerprint] = cached

	return cached, nil
}

// AddFromMetadata adds certificates from IdP metadata (base64 encoded)
func (c *CertificateCache) AddFromMetadata(entityID string, certStrings []string) ([]*CachedCertificate, error) {
	certs, err := ParseCertificatesFromMetadata(certStrings)
	if err != nil {
		return nil, err
	}

	result := make([]*CachedCertificate, 0, len(certs))
	for _, cert := range certs {
		cached, err := c.Add(cert, entityID, CertificateSourceMetadata)
		if err != nil {
			continue
		}
		result = append(result, cached)
	}

	return result, nil
}

// Get retrieves a certificate by fingerprint
func (c *CertificateCache) Get(fingerprint string) (*CachedCertificate, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, ok := c.certificates[fingerprint]
	if !ok {
		return nil, false
	}

	// Check if expired
	if time.Now().After(cached.ExpiresAt) {
		return nil, false
	}

	return cached, true
}

// GetByEntityID retrieves all certificates for an entity ID
func (c *CertificateCache) GetByEntityID(entityID string) []*CachedCertificate {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entityCerts, ok := c.byEntityID[entityID]
	if !ok {
		return nil
	}

	now := time.Now()
	result := make([]*CachedCertificate, 0, len(entityCerts))
	for _, cached := range entityCerts {
		if now.Before(cached.ExpiresAt) {
			result = append(result, cached)
		}
	}

	return result
}

// GetCertificatesByEntityID retrieves parsed certificates for an entity ID
func (c *CertificateCache) GetCertificatesByEntityID(entityID string) []*x509.Certificate {
	cached := c.GetByEntityID(entityID)
	result := make([]*x509.Certificate, len(cached))
	for i, cc := range cached {
		result[i] = cc.Certificate
	}
	return result
}

// GetValidatedByEntityID retrieves only validated certificates for an entity ID
func (c *CertificateCache) GetValidatedByEntityID(entityID string) []*CachedCertificate {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entityCerts, ok := c.byEntityID[entityID]
	if !ok {
		return nil
	}

	now := time.Now()
	result := make([]*CachedCertificate, 0, len(entityCerts))
	for _, cached := range entityCerts {
		if now.Before(cached.ExpiresAt) && cached.Validated {
			result = append(result, cached)
		}
	}

	return result
}

// Remove removes a certificate from the cache
func (c *CertificateCache) Remove(fingerprint string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cached, ok := c.certificates[fingerprint]
	if !ok {
		return
	}

	delete(c.certificates, fingerprint)
	if entityCerts, ok := c.byEntityID[cached.EntityID]; ok {
		delete(entityCerts, fingerprint)
		if len(entityCerts) == 0 {
			delete(c.byEntityID, cached.EntityID)
		}
	}
}

// RemoveByEntityID removes all certificates for an entity ID
func (c *CertificateCache) RemoveByEntityID(entityID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entityCerts, ok := c.byEntityID[entityID]
	if !ok {
		return
	}

	for fingerprint := range entityCerts {
		delete(c.certificates, fingerprint)
	}
	delete(c.byEntityID, entityID)
}

// Clear removes all cached certificates
func (c *CertificateCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.certificates = make(map[string]*CachedCertificate)
	c.byEntityID = make(map[string]map[string]*CachedCertificate)
}

// Size returns the number of cached certificates
func (c *CertificateCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.certificates)
}

// Stats returns cache statistics
func (c *CertificateCache) Stats() CertificateCacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	var expired, validated, fromMetadata, fromSAML int

	for _, cached := range c.certificates {
		if now.After(cached.ExpiresAt) {
			expired++
		}
		if cached.Validated {
			validated++
		}
		switch cached.Source {
		case CertificateSourceMetadata:
			fromMetadata++
		case CertificateSourceSAML:
			fromSAML++
		}
	}

	return CertificateCacheStats{
		TotalCached:    len(c.certificates),
		ExpiredCount:   expired,
		ValidatedCount: validated,
		EntityCount:    len(c.byEntityID),
		FromMetadata:   fromMetadata,
		FromSAML:       fromSAML,
		MaxSize:        c.config.MaxSize,
		TTL:            c.config.TTL,
	}
}

// CertificateCacheStats contains cache statistics
type CertificateCacheStats struct {
	TotalCached    int
	ExpiredCount   int
	ValidatedCount int
	EntityCount    int
	FromMetadata   int
	FromSAML       int
	MaxSize        int
	TTL            time.Duration
}

// evictOldest removes the oldest cache entry
func (c *CertificateCache) evictOldest() {
	var oldestFingerprint string
	var oldestTime time.Time

	for fingerprint, cached := range c.certificates {
		if oldestFingerprint == "" || cached.CachedAt.Before(oldestTime) {
			oldestFingerprint = fingerprint
			oldestTime = cached.CachedAt
		}
	}

	if oldestFingerprint != "" {
		cached := c.certificates[oldestFingerprint]
		delete(c.certificates, oldestFingerprint)
		if entityCerts, ok := c.byEntityID[cached.EntityID]; ok {
			delete(entityCerts, oldestFingerprint)
			if len(entityCerts) == 0 {
				delete(c.byEntityID, cached.EntityID)
			}
		}
	}
}

// ============================================================================
// Certificate Fingerprint Utilities
// ============================================================================

// computeCertFingerprint computes the SHA-256 fingerprint of a certificate
func computeCertFingerprint(cert *x509.Certificate) string {
	hash := sha256.Sum256(cert.Raw)
	return hex.EncodeToString(hash[:])
}

// FormatFingerprint formats a fingerprint for display (colon-separated)
func FormatFingerprint(fingerprint string) string {
	if len(fingerprint) != 64 {
		return fingerprint
	}

	result := make([]byte, 0, 64+31)
	for i := 0; i < len(fingerprint); i += 2 {
		if i > 0 {
			result = append(result, ':')
		}
		result = append(result, fingerprint[i], fingerprint[i+1])
	}
	return string(result)
}

// ============================================================================
// Global Certificate Cache
// ============================================================================

var (
	globalCertCache     *CertificateCache
	globalCertCacheOnce sync.Once
)

// GetGlobalCertificateCache returns the global certificate cache instance
func GetGlobalCertificateCache() *CertificateCache {
	globalCertCacheOnce.Do(func() {
		globalCertCache = NewCertificateCache(DefaultCertificateCacheConfig())
	})
	return globalCertCache
}
