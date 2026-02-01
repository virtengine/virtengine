// Package oidc provides OIDC token verification for the VEID SSO verification service.
package oidc

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// ============================================================================
// JWKS Manager
// ============================================================================

// JWKSManager manages JWKS caching and rotation for OIDC issuers.
type JWKSManager struct {
	httpClient *http.Client
	logger     zerolog.Logger

	mu     sync.RWMutex
	caches map[string]*jwksCache // issuer -> cache
	config *JWKSManagerConfig
}

// JWKSManagerConfig contains configuration for the JWKS manager.
type JWKSManagerConfig struct {
	// DefaultCacheTTL is the default cache TTL
	DefaultCacheTTL time.Duration `json:"default_cache_ttl"`

	// RefreshInterval is how often to proactively refresh JWKS
	RefreshInterval time.Duration `json:"refresh_interval"`

	// HTTPTimeout is the HTTP client timeout
	HTTPTimeout time.Duration `json:"http_timeout"`

	// MaxRetries is the maximum number of retries for JWKS fetch
	MaxRetries int `json:"max_retries"`

	// RetryBackoff is the base retry backoff duration
	RetryBackoff time.Duration `json:"retry_backoff"`
}

// DefaultJWKSManagerConfig returns the default configuration.
func DefaultJWKSManagerConfig() *JWKSManagerConfig {
	return &JWKSManagerConfig{
		DefaultCacheTTL: time.Hour,
		RefreshInterval: 30 * time.Minute,
		HTTPTimeout:     30 * time.Second,
		MaxRetries:      3,
		RetryBackoff:    time.Second,
	}
}

// jwksCache contains cached JWKS data for an issuer.
type jwksCache struct {
	jwksURL     string
	keys        map[string]*JSONWebKey // kid -> key
	fetchedAt   time.Time
	expiresAt   time.Time
	lastError   error
	lastErrorAt *time.Time
	refreshing  bool
}

// JSONWebKey represents a JSON Web Key.
type JSONWebKey struct {
	// Standard JWK fields
	KTY string `json:"kty"` // Key Type (RSA, EC, etc.)
	USE string `json:"use"` // Key Usage (sig, enc)
	KID string `json:"kid"` // Key ID
	ALG string `json:"alg"` // Algorithm

	// RSA key fields
	N string `json:"n,omitempty"` // RSA modulus
	E string `json:"e,omitempty"` // RSA exponent

	// EC key fields
	CRV string `json:"crv,omitempty"` // Curve
	X   string `json:"x,omitempty"`   // X coordinate
	Y   string `json:"y,omitempty"`   // Y coordinate

	// Parsed key (cached)
	parsedKey interface{} `json:"-"`
}

// JWKS represents a JSON Web Key Set.
type JWKS struct {
	Keys []JSONWebKey `json:"keys"`
}

// NewJWKSManager creates a new JWKS manager.
func NewJWKSManager(config *JWKSManagerConfig, logger zerolog.Logger) *JWKSManager {
	if config == nil {
		config = DefaultJWKSManagerConfig()
	}

	return &JWKSManager{
		httpClient: &http.Client{
			Timeout: config.HTTPTimeout,
		},
		logger: logger.With().Str("component", "jwks_manager").Logger(),
		caches: make(map[string]*jwksCache),
		config: config,
	}
}

// GetKey retrieves a key by issuer and key ID.
func (m *JWKSManager) GetKey(ctx context.Context, issuer, jwksURL, kid string) (*JSONWebKey, error) {
	m.mu.RLock()
	cache, ok := m.caches[issuer]
	m.mu.RUnlock()

	// Check if we have a valid cached key
	if ok && time.Now().Before(cache.expiresAt) {
		if key, found := cache.keys[kid]; found {
			return key, nil
		}
		// Key not found in cache, but cache is still valid
		// This might indicate a key rotation, so refresh
	}

	// Fetch/refresh JWKS
	if err := m.fetchJWKS(ctx, issuer, jwksURL); err != nil {
		// If we have stale cache, use it
		if ok && cache.keys != nil {
			if key, found := cache.keys[kid]; found {
				m.logger.Warn().
					Str("issuer", issuer).
					Str("kid", kid).
					Err(err).
					Msg("using stale JWKS cache after fetch failure")
				return key, nil
			}
		}
		return nil, err
	}

	// Try to get the key from refreshed cache
	m.mu.RLock()
	cache = m.caches[issuer]
	m.mu.RUnlock()

	if cache == nil {
		return nil, fmt.Errorf("%w: cache is nil after fetch", ErrJWKSFetchFailed)
	}

	key, found := cache.keys[kid]
	if !found {
		return nil, fmt.Errorf("%w: key ID: %s, issuer: %s", ErrKeyNotFound, kid, issuer)
	}

	return key, nil
}

// fetchJWKS fetches the JWKS from the issuer.
func (m *JWKSManager) fetchJWKS(ctx context.Context, issuer, jwksURL string) error {
	m.mu.Lock()
	cache, ok := m.caches[issuer]
	if ok && cache.refreshing {
		m.mu.Unlock()
		// Wait a bit for the other goroutine to finish
		time.Sleep(100 * time.Millisecond)
		return nil
	}
	if !ok {
		cache = &jwksCache{
			jwksURL: jwksURL,
			keys:    make(map[string]*JSONWebKey),
		}
		m.caches[issuer] = cache
	}
	cache.refreshing = true
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		cache.refreshing = false
		m.mu.Unlock()
	}()

	var lastErr error
	for attempt := 0; attempt <= m.config.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := m.config.RetryBackoff * time.Duration(1<<(attempt-1))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, jwksURL, nil)
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Accept", "application/json")

		resp, err := m.httpClient.Do(req)
		if err != nil {
			lastErr = err
			m.logger.Debug().
				Str("issuer", issuer).
				Str("jwks_url", jwksURL).
				Int("attempt", attempt+1).
				Err(err).
				Msg("JWKS fetch attempt failed")
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			continue
		}

		var jwks JWKS
		if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
			lastErr = err
			continue
		}

		// Parse and cache keys
		keys := make(map[string]*JSONWebKey, len(jwks.Keys))
		for i := range jwks.Keys {
			key := &jwks.Keys[i]
			if err := m.parseKey(key); err != nil {
				m.logger.Warn().
					Str("issuer", issuer).
					Str("kid", key.KID).
					Err(err).
					Msg("failed to parse JWK")
				continue
			}
			keys[key.KID] = key
		}

		m.mu.Lock()
		cache.keys = keys
		cache.fetchedAt = time.Now()
		cache.expiresAt = time.Now().Add(m.config.DefaultCacheTTL)
		cache.lastError = nil
		cache.lastErrorAt = nil
		m.mu.Unlock()

		m.logger.Debug().
			Str("issuer", issuer).
			Int("key_count", len(keys)).
			Msg("JWKS refreshed successfully")

		return nil
	}

	// Update error state
	now := time.Now()
	m.mu.Lock()
	cache.lastError = lastErr
	cache.lastErrorAt = &now
	m.mu.Unlock()

	return fmt.Errorf("%w: after %d retries: %v", ErrJWKSFetchFailed, m.config.MaxRetries, lastErr)
}

// parseKey parses the JWK into a usable crypto key.
func (m *JWKSManager) parseKey(key *JSONWebKey) error {
	switch key.KTY {
	case "RSA":
		return m.parseRSAKey(key)
	default:
		return fmt.Errorf("unsupported key type: %s", key.KTY)
	}
}

// parseRSAKey parses an RSA JWK.
func (m *JWKSManager) parseRSAKey(key *JSONWebKey) error {
	nBytes, err := base64URLDecode(key.N)
	if err != nil {
		return fmt.Errorf("failed to decode modulus: %w", err)
	}

	eBytes, err := base64URLDecode(key.E)
	if err != nil {
		return fmt.Errorf("failed to decode exponent: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	key.parsedKey = &rsa.PublicKey{
		N: n,
		E: e,
	}

	return nil
}

// GetPublicKey returns the parsed public key.
func (key *JSONWebKey) GetPublicKey() interface{} {
	return key.parsedKey
}

// Refresh forces a refresh of the JWKS for an issuer.
func (m *JWKSManager) Refresh(ctx context.Context, issuer, jwksURL string) error {
	// Invalidate cache first
	m.mu.Lock()
	if cache, ok := m.caches[issuer]; ok {
		cache.expiresAt = time.Time{} // Force expiry
	}
	m.mu.Unlock()

	return m.fetchJWKS(ctx, issuer, jwksURL)
}

// GetCacheStatus returns the cache status for an issuer.
func (m *JWKSManager) GetCacheStatus(issuer string) *IssuerHealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cache, ok := m.caches[issuer]
	if !ok {
		return &IssuerHealthStatus{
			Issuer:     issuer,
			Healthy:    false,
			JWKSCached: false,
		}
	}

	status := &IssuerHealthStatus{
		Issuer:          issuer,
		Healthy:         cache.lastError == nil && time.Now().Before(cache.expiresAt),
		JWKSCached:      len(cache.keys) > 0,
		JWKSLastRefresh: &cache.fetchedAt,
		JWKSExpiresAt:   &cache.expiresAt,
		KeyCount:        len(cache.keys),
	}

	if cache.lastError != nil {
		status.LastError = cache.lastError.Error()
		status.LastErrorAt = cache.lastErrorAt
	}

	return status
}

// Close closes the JWKS manager.
func (m *JWKSManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.caches = make(map[string]*jwksCache)
	return nil
}

// base64URLDecode decodes base64url encoded data.
func base64URLDecode(s string) ([]byte, error) {
	// Add padding if necessary
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}

