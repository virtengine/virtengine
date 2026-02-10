// Package sev provides the AMD Key Distribution Server (KDS) client.
//
// The AMD KDS provides VCEK (Versioned Chip Endorsement Key) certificates
// and the certificate chain (ASK, ARK) required to verify SEV-SNP attestation
// reports.
//
// # KDS Endpoints
//
// AMD provides different endpoints for different processor families:
// - Milan (EPYC 7003): https://kdsintf.amd.com/vcek/v1/Milan
// - Genoa (EPYC 9004): https://kdsintf.amd.com/vcek/v1/Genoa
//
// # Certificate Types
//
// - VCEK: Versioned Chip Endorsement Key - unique per chip and TCB version
// - ASK: AMD SEV Signing Key - signs VCEKs
// - ARK: AMD Root Key - signs ASK, root of trust
//
// # Rate Limiting
//
// AMD KDS has rate limits. This client implements:
// - Local caching with configurable TTL
// - Automatic retry with exponential backoff
// - Concurrent request deduplication
//
// Task Reference: VE-2029 - Hardware TEE Integration Layer
package sev

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/security"
)

// =============================================================================
// Constants
// =============================================================================

const (
	// Default KDS base URLs
	DefaultKDSBaseURL  = "https://kdsintf.amd.com"
	DefaultKDSVCEKPath = "/vcek/v1"
	DefaultKDSCertPath = "/cek/v1"

	// Product names for URL construction
	ProductMilan   = "Milan"
	ProductGenoa   = "Genoa"
	ProductBergamo = "Bergamo"
	ProductSiena   = "Siena"

	// Certificate endpoints
	VCEKEndpoint  = "vcek"
	ASKEndpoint   = "ask"
	ARKEndpoint   = "ark"
	ChainEndpoint = "cert_chain"

	// Default timeouts
	DefaultRequestTimeout = 30 * time.Second
	DefaultCacheTTL       = 24 * time.Hour

	// Retry configuration
	DefaultMaxRetries     = 3
	DefaultRetryBaseDelay = 1 * time.Second
	DefaultRetryMaxDelay  = 30 * time.Second
)

// =============================================================================
// Errors
// =============================================================================

var (
	// ErrKDSUnavailable indicates the KDS service is unavailable
	ErrKDSUnavailable = errors.New("kds: AMD KDS service unavailable")

	// ErrCertNotFound indicates the requested certificate was not found
	ErrCertNotFound = errors.New("kds: certificate not found")

	// ErrInvalidCert indicates the certificate is malformed
	ErrInvalidCert = errors.New("kds: invalid certificate")

	// ErrRateLimited indicates we've been rate limited
	ErrRateLimited = errors.New("kds: rate limited by AMD KDS")

	// ErrInvalidChipID indicates an invalid chip ID was provided
	ErrInvalidChipID = errors.New("kds: invalid chip ID")

	// ErrInvalidTCB indicates an invalid TCB version was provided
	ErrInvalidTCB = errors.New("kds: invalid TCB version")

	// ErrCacheExpired indicates the cached certificate has expired
	ErrCacheExpired = errors.New("kds: cached certificate expired")

	// ErrChainVerification indicates certificate chain verification failed
	ErrChainVerification = errors.New("kds: certificate chain verification failed")
)

// KDSError provides detailed error information
type KDSError struct {
	Op         string
	Product    string
	StatusCode int
	Err        error
}

func (e *KDSError) Error() string {
	if e.StatusCode != 0 {
		return fmt.Sprintf("kds: %s for %s failed: HTTP %d: %v", e.Op, e.Product, e.StatusCode, e.Err)
	}
	return fmt.Sprintf("kds: %s for %s failed: %v", e.Op, e.Product, e.Err)
}

func (e *KDSError) Unwrap() error {
	return e.Err
}

// =============================================================================
// Configuration
// =============================================================================

// Config configures the KDS client
type Config struct {
	// BaseURL is the KDS base URL (default: https://kdsintf.amd.com)
	BaseURL string

	// Product is the processor product name (Milan, Genoa, etc.)
	Product string

	// HTTPClient is the HTTP client to use (default: http.DefaultClient with timeout)
	HTTPClient *http.Client

	// RequestTimeout is the timeout for individual requests
	RequestTimeout time.Duration

	// CacheTTL is how long to cache certificates
	CacheTTL time.Duration

	// MaxRetries is the maximum number of retry attempts
	MaxRetries int

	// RetryBaseDelay is the base delay for exponential backoff
	RetryBaseDelay time.Duration

	// RetryMaxDelay is the maximum delay between retries
	RetryMaxDelay time.Duration

	// DisableCache disables certificate caching
	DisableCache bool
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		BaseURL:        DefaultKDSBaseURL,
		Product:        ProductMilan,
		RequestTimeout: DefaultRequestTimeout,
		CacheTTL:       DefaultCacheTTL,
		MaxRetries:     DefaultMaxRetries,
		RetryBaseDelay: DefaultRetryBaseDelay,
		RetryMaxDelay:  DefaultRetryMaxDelay,
	}
}

// MilanConfig returns configuration for AMD EPYC Milan (7003 series)
func MilanConfig() *Config {
	cfg := DefaultConfig()
	cfg.Product = ProductMilan
	return cfg
}

// GenoaConfig returns configuration for AMD EPYC Genoa (9004 series)
func GenoaConfig() *Config {
	cfg := DefaultConfig()
	cfg.Product = ProductGenoa
	return cfg
}

// =============================================================================
// Cache Entry
// =============================================================================

type cacheEntry struct {
	cert      *x509.Certificate
	raw       []byte
	fetchedAt time.Time
	expiresAt time.Time
}

func (e *cacheEntry) isExpired() bool {
	return time.Now().After(e.expiresAt)
}

// =============================================================================
// Certificate Chain
// =============================================================================

// CertChain represents the AMD SEV-SNP certificate chain
type CertChain struct {
	// VCEK is the Versioned Chip Endorsement Key certificate
	VCEK *x509.Certificate

	// ASK is the AMD SEV Signing Key certificate
	ASK *x509.Certificate

	// ARK is the AMD Root Key certificate
	ARK *x509.Certificate

	// Raw DER-encoded certificates
	VCEKRaw []byte
	ASKRaw  []byte
	ARKRaw  []byte
}

// Verify verifies the certificate chain
func (c *CertChain) Verify() error {
	if c.VCEK == nil || c.ASK == nil || c.ARK == nil {
		return errors.New("incomplete certificate chain")
	}

	// Verify ARK is self-signed
	if err := c.ARK.CheckSignatureFrom(c.ARK); err != nil {
		return fmt.Errorf("ARK not self-signed: %w", err)
	}

	// Verify ASK is signed by ARK
	if err := c.ASK.CheckSignatureFrom(c.ARK); err != nil {
		return fmt.Errorf("ASK not signed by ARK: %w", err)
	}

	// Verify VCEK is signed by ASK
	if err := c.VCEK.CheckSignatureFrom(c.ASK); err != nil {
		return fmt.Errorf("VCEK not signed by ASK: %w", err)
	}

	return nil
}

// =============================================================================
// KDS Client
// =============================================================================

// KDSClient is a client for the AMD Key Distribution Server
type KDSClient struct {
	config *Config
	client *http.Client

	// Cache
	mu        sync.RWMutex
	vcekCache map[string]*cacheEntry // key: chipID-tcb
	askCache  *cacheEntry
	arkCache  *cacheEntry

	// Request deduplication
	inFlight   map[string]*sync.WaitGroup
	inFlightMu sync.Mutex
}

// NewKDSClient creates a new KDS client
func NewKDSClient(config *Config) *KDSClient {
	if config == nil {
		config = DefaultConfig()
	}

	client := config.HTTPClient
	if client == nil {
		client = security.NewSecureHTTPClient(security.WithTimeout(config.RequestTimeout))
	}

	return &KDSClient{
		config:    config,
		client:    client,
		vcekCache: make(map[string]*cacheEntry),
		inFlight:  make(map[string]*sync.WaitGroup),
	}
}

// GetVCEK retrieves the VCEK certificate for a specific chip and TCB version
func (c *KDSClient) GetVCEK(chipID []byte, tcb TCBVersion) (*x509.Certificate, error) {
	return c.GetVCEKWithContext(context.Background(), chipID, tcb)
}

// GetVCEKWithContext retrieves the VCEK certificate with context
func (c *KDSClient) GetVCEKWithContext(ctx context.Context, chipID []byte, tcb TCBVersion) (*x509.Certificate, error) {
	if len(chipID) != ChipIDSize {
		return nil, ErrInvalidChipID
	}

	cacheKey := c.vcekCacheKey(chipID, tcb)

	// Check cache
	if !c.config.DisableCache {
		if entry := c.getCachedVCEK(cacheKey); entry != nil {
			return entry.cert, nil
		}
	}

	// Deduplicate concurrent requests
	cert, err := c.fetchWithDedup(ctx, cacheKey, func(ctx context.Context) (*x509.Certificate, []byte, error) {
		return c.fetchVCEK(ctx, chipID, tcb)
	})
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// GetASK retrieves the AMD SEV Signing Key certificate
func (c *KDSClient) GetASK() (*x509.Certificate, error) {
	return c.GetASKWithContext(context.Background())
}

// GetASKWithContext retrieves the ASK certificate with context
func (c *KDSClient) GetASKWithContext(ctx context.Context) (*x509.Certificate, error) {
	// Check cache
	if !c.config.DisableCache {
		c.mu.RLock()
		if c.askCache != nil && !c.askCache.isExpired() {
			cert := c.askCache.cert
			c.mu.RUnlock()
			return cert, nil
		}
		c.mu.RUnlock()
	}

	// Fetch
	cert, raw, err := c.fetchCertificate(ctx, ASKEndpoint)
	if err != nil {
		return nil, err
	}

	// Cache
	if !c.config.DisableCache {
		c.mu.Lock()
		c.askCache = &cacheEntry{
			cert:      cert,
			raw:       raw,
			fetchedAt: time.Now(),
			expiresAt: time.Now().Add(c.config.CacheTTL),
		}
		c.mu.Unlock()
	}

	return cert, nil
}

// GetARK retrieves the AMD Root Key certificate
func (c *KDSClient) GetARK() (*x509.Certificate, error) {
	return c.GetARKWithContext(context.Background())
}

// GetARKWithContext retrieves the ARK certificate with context
func (c *KDSClient) GetARKWithContext(ctx context.Context) (*x509.Certificate, error) {
	// Check cache
	if !c.config.DisableCache {
		c.mu.RLock()
		if c.arkCache != nil && !c.arkCache.isExpired() {
			cert := c.arkCache.cert
			c.mu.RUnlock()
			return cert, nil
		}
		c.mu.RUnlock()
	}

	// Fetch
	cert, raw, err := c.fetchCertificate(ctx, ARKEndpoint)
	if err != nil {
		return nil, err
	}

	// Cache
	if !c.config.DisableCache {
		c.mu.Lock()
		c.arkCache = &cacheEntry{
			cert:      cert,
			raw:       raw,
			fetchedAt: time.Now(),
			expiresAt: time.Now().Add(c.config.CacheTTL),
		}
		c.mu.Unlock()
	}

	return cert, nil
}

// GetCertificateChain retrieves the complete certificate chain for a chip
func (c *KDSClient) GetCertificateChain(chipID []byte, tcb TCBVersion) (*CertChain, error) {
	return c.GetCertificateChainWithContext(context.Background(), chipID, tcb)
}

// GetCertificateChainWithContext retrieves the certificate chain with context
func (c *KDSClient) GetCertificateChainWithContext(ctx context.Context, chipID []byte, tcb TCBVersion) (*CertChain, error) {
	// Fetch all certificates concurrently
	var (
		wg      sync.WaitGroup
		vcek    *x509.Certificate
		ask     *x509.Certificate
		ark     *x509.Certificate
		vcekRaw []byte
		askRaw  []byte
		arkRaw  []byte
		vcekErr error
		askErr  error
		arkErr  error
	)

	wg.Add(3)

	go func() {
		defer wg.Done()
		var raw []byte
		vcek, raw, vcekErr = c.fetchVCEK(ctx, chipID, tcb)
		if vcekErr == nil {
			vcekRaw = raw
		}
	}()

	go func() {
		defer wg.Done()
		ask, askErr = c.GetASKWithContext(ctx)
		if askErr == nil {
			c.mu.RLock()
			if c.askCache != nil {
				askRaw = c.askCache.raw
			}
			c.mu.RUnlock()
		}
	}()

	go func() {
		defer wg.Done()
		ark, arkErr = c.GetARKWithContext(ctx)
		if arkErr == nil {
			c.mu.RLock()
			if c.arkCache != nil {
				arkRaw = c.arkCache.raw
			}
			c.mu.RUnlock()
		}
	}()

	wg.Wait()

	// Check for errors
	if vcekErr != nil {
		return nil, fmt.Errorf("failed to get VCEK: %w", vcekErr)
	}
	if askErr != nil {
		return nil, fmt.Errorf("failed to get ASK: %w", askErr)
	}
	if arkErr != nil {
		return nil, fmt.Errorf("failed to get ARK: %w", arkErr)
	}

	chain := &CertChain{
		VCEK:    vcek,
		ASK:     ask,
		ARK:     ark,
		VCEKRaw: vcekRaw,
		ASKRaw:  askRaw,
		ARKRaw:  arkRaw,
	}

	// Verify chain
	if err := chain.Verify(); err != nil {
		return nil, &KDSError{
			Op:      "verify_chain",
			Product: c.config.Product,
			Err:     err,
		}
	}

	return chain, nil
}

// =============================================================================
// Internal Methods
// =============================================================================

// vcekCacheKey generates a cache key for VCEK
func (c *KDSClient) vcekCacheKey(chipID []byte, tcb TCBVersion) string {
	h := sha256.New()
	h.Write(chipID)
	h.Write([]byte{tcb.BootLoader, tcb.TEE, tcb.SNP, tcb.Microcode})
	return hex.EncodeToString(h.Sum(nil)[:16])
}

// getCachedVCEK returns cached VCEK if valid
func (c *KDSClient) getCachedVCEK(key string) *cacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.vcekCache[key]
	if !ok || entry.isExpired() {
		return nil
	}
	return entry
}

// fetchWithDedup fetches with request deduplication
func (c *KDSClient) fetchWithDedup(ctx context.Context, key string, fetch func(context.Context) (*x509.Certificate, []byte, error)) (*x509.Certificate, error) {
	c.inFlightMu.Lock()
	if wg, ok := c.inFlight[key]; ok {
		c.inFlightMu.Unlock()
		wg.Wait()
		// Check cache after waiting
		if entry := c.getCachedVCEK(key); entry != nil {
			return entry.cert, nil
		}
		return nil, errors.New("concurrent request failed")
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	c.inFlight[key] = wg
	c.inFlightMu.Unlock()

	defer func() {
		c.inFlightMu.Lock()
		delete(c.inFlight, key)
		c.inFlightMu.Unlock()
		wg.Done()
	}()

	cert, raw, err := fetch(ctx)
	if err != nil {
		return nil, err
	}

	// Cache result
	if !c.config.DisableCache {
		c.mu.Lock()
		c.vcekCache[key] = &cacheEntry{
			cert:      cert,
			raw:       raw,
			fetchedAt: time.Now(),
			expiresAt: time.Now().Add(c.config.CacheTTL),
		}
		c.mu.Unlock()
	}

	return cert, nil
}

// fetchVCEK fetches VCEK from KDS
func (c *KDSClient) fetchVCEK(ctx context.Context, chipID []byte, tcb TCBVersion) (*x509.Certificate, []byte, error) {
	// Build URL with TCB parameters
	// Format: /vcek/v1/{product}/{chip_id}?blSPL=X&teeSPL=Y&snpSPL=Z&ucodeSPL=W
	chipIDHex := hex.EncodeToString(chipID)

	params := url.Values{}
	params.Set("blSPL", fmt.Sprintf("%d", tcb.BootLoader))
	params.Set("teeSPL", fmt.Sprintf("%d", tcb.TEE))
	params.Set("snpSPL", fmt.Sprintf("%d", tcb.SNP))
	params.Set("ucodeSPL", fmt.Sprintf("%d", tcb.Microcode))

	urlStr := fmt.Sprintf("%s%s/%s/%s?%s",
		c.config.BaseURL,
		DefaultKDSVCEKPath,
		c.config.Product,
		chipIDHex,
		params.Encode(),
	)

	raw, err := c.doRequestWithRetry(ctx, urlStr)
	if err != nil {
		return nil, nil, err
	}

	cert, err := x509.ParseCertificate(raw)
	if err != nil {
		return nil, nil, &KDSError{
			Op:      "parse_vcek",
			Product: c.config.Product,
			Err:     fmt.Errorf("%w: %v", ErrInvalidCert, err),
		}
	}

	return cert, raw, nil
}

// fetchCertificate fetches ASK or ARK
func (c *KDSClient) fetchCertificate(ctx context.Context, certType string) (*x509.Certificate, []byte, error) {
	urlStr := fmt.Sprintf("%s%s/%s/%s",
		c.config.BaseURL,
		DefaultKDSCertPath,
		c.config.Product,
		certType,
	)

	raw, err := c.doRequestWithRetry(ctx, urlStr)
	if err != nil {
		return nil, nil, err
	}

	cert, err := x509.ParseCertificate(raw)
	if err != nil {
		return nil, nil, &KDSError{
			Op:      "parse_" + certType,
			Product: c.config.Product,
			Err:     fmt.Errorf("%w: %v", ErrInvalidCert, err),
		}
	}

	return cert, raw, nil
}

// doRequestWithRetry performs HTTP request with retry logic
func (c *KDSClient) doRequestWithRetry(ctx context.Context, urlStr string) ([]byte, error) {
	var lastErr error
	delay := c.config.RetryBaseDelay

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
			// Exponential backoff
			delay *= 2
			if delay > c.config.RetryMaxDelay {
				delay = c.config.RetryMaxDelay
			}
		}

		body, err := c.doRequest(ctx, urlStr)
		if err == nil {
			return body, nil
		}

		lastErr = err

		// Don't retry on certain errors
		var kdsErr *KDSError
		if errors.As(err, &kdsErr) {
			switch kdsErr.StatusCode {
			case http.StatusNotFound:
				return nil, ErrCertNotFound
			case http.StatusBadRequest:
				return nil, err // Invalid request, don't retry
			}
			if kdsErr.StatusCode == http.StatusTooManyRequests {
				lastErr = ErrRateLimited
			}
		}
	}

	return nil, lastErr
}

// doRequest performs a single HTTP request
func (c *KDSClient) doRequest(ctx context.Context, urlStr string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/x-x509-ca-cert")
	req.Header.Set("User-Agent", "VirtEngine-SEV-Client/1.0")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, &KDSError{
			Op:      "request",
			Product: c.config.Product,
			Err:     err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &KDSError{
			Op:         "request",
			Product:    c.config.Product,
			StatusCode: resp.StatusCode,
			Err:        fmt.Errorf("unexpected status: %s", resp.Status),
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &KDSError{
			Op:      "read_response",
			Product: c.config.Product,
			Err:     err,
		}
	}

	return body, nil
}

// ClearCache clears all cached certificates
func (c *KDSClient) ClearCache() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.vcekCache = make(map[string]*cacheEntry)
	c.askCache = nil
	c.arkCache = nil
}

// GetCacheStats returns cache statistics
func (c *KDSClient) GetCacheStats() (vcekCount int, hasASK, hasARK bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.vcekCache), c.askCache != nil, c.arkCache != nil
}
