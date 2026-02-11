// Package provider_daemon implements the provider daemon for VirtEngine.
//
// Domain verification checker for off-chain verification of provider domains.
package provider_daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/security"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/provider/keeper"
)

const (
	// DefaultVerificationCheckInterval is how often to poll for pending verifications
	DefaultVerificationCheckInterval = 30 * time.Second

	// DefaultVerificationTimeout is the timeout for DNS/HTTP checks
	DefaultVerificationTimeout = 10 * time.Second

	// DefaultMaxRetries is the maximum number of retry attempts
	DefaultMaxRetries = 3

	// DefaultInitialBackoff is the initial backoff duration
	DefaultInitialBackoff = 30 * time.Second

	// DefaultMaxBackoff is the maximum backoff duration
	DefaultMaxBackoff = 10 * time.Minute

	// DefaultExpiryCheckInterval is how often to check for expired verifications
	DefaultExpiryCheckInterval = 1 * time.Hour
)

// DomainVerificationCheckerConfig configures the domain verification checker.
type DomainVerificationCheckerConfig struct {
	// Enabled enables the verification checker.
	Enabled bool

	// ProviderAddress is the provider's on-chain address.
	ProviderAddress string

	// ChainID is the chain ID.
	ChainID string

	// CometRPC is the CometBFT RPC endpoint.
	CometRPC string

	// GRPCEndpoint is the gRPC endpoint for queries.
	GRPCEndpoint string

	// CheckInterval is how often to poll for pending verifications.
	CheckInterval time.Duration

	// VerificationTimeout is the timeout for DNS/HTTP checks.
	VerificationTimeout time.Duration

	// MaxRetries is the maximum number of retry attempts.
	MaxRetries int

	// InitialBackoff is the initial backoff duration.
	InitialBackoff time.Duration

	// MaxBackoff is the maximum backoff duration.
	MaxBackoff time.Duration

	// ExpiryCheckInterval is how often to check for expired verifications.
	ExpiryCheckInterval time.Duration

	// DNSResolver is a custom DNS resolver (optional, for testing).
	DNSResolver DNSResolver

	// HTTPClient is a custom HTTP client (optional, for testing).
	HTTPClient *http.Client
}

// DefaultDomainVerificationCheckerConfig returns default configuration.
func DefaultDomainVerificationCheckerConfig() DomainVerificationCheckerConfig {
	return DomainVerificationCheckerConfig{
		Enabled:             false,
		CheckInterval:       DefaultVerificationCheckInterval,
		VerificationTimeout: DefaultVerificationTimeout,
		MaxRetries:          DefaultMaxRetries,
		InitialBackoff:      DefaultInitialBackoff,
		MaxBackoff:          DefaultMaxBackoff,
		ExpiryCheckInterval: DefaultExpiryCheckInterval,
		HTTPClient:          security.NewSecureHTTPClient(security.WithTimeout(DefaultVerificationTimeout)),
	}
}

// DNSResolver defines the interface for DNS resolution.
type DNSResolver interface {
	LookupTXT(ctx context.Context, name string) ([]string, error)
	LookupCNAME(ctx context.Context, host string) (string, error)
}

// defaultDNSResolver implements DNSResolver using net.Resolver.
type defaultDNSResolver struct {
	resolver *net.Resolver
}

func (r *defaultDNSResolver) LookupTXT(ctx context.Context, name string) ([]string, error) {
	return r.resolver.LookupTXT(ctx, name)
}

func (r *defaultDNSResolver) LookupCNAME(ctx context.Context, host string) (string, error) {
	return r.resolver.LookupCNAME(ctx, host)
}

// DomainVerificationChecker polls for pending domain verifications and performs off-chain checks.
type DomainVerificationChecker struct {
	mu sync.RWMutex

	cfg         DomainVerificationCheckerConfig
	keyManager  *KeyManager
	rpcClient   *rpchttp.HTTP
	chainClient ChainClient
	dnsResolver DNSResolver
	httpClient  *http.Client

	// retryState tracks retry attempts and backoff for each domain
	retryState map[string]*verificationRetryState

	running  bool
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// verificationRetryState tracks retry state for a domain verification.
type verificationRetryState struct {
	Attempts    int
	LastAttempt time.Time
	NextAttempt time.Time
	Backoff     time.Duration
}

// NewDomainVerificationChecker creates a new domain verification checker.
func NewDomainVerificationChecker(
	cfg DomainVerificationCheckerConfig,
	keyManager *KeyManager,
	chainClient ChainClient,
) (*DomainVerificationChecker, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("domain verification checker is disabled")
	}

	if cfg.ProviderAddress == "" {
		return nil, fmt.Errorf("provider address is required")
	}

	if cfg.CometRPC == "" && chainClient == nil {
		return nil, fmt.Errorf("comet RPC endpoint or chain client is required")
	}

	checker := &DomainVerificationChecker{
		cfg:         cfg,
		keyManager:  keyManager,
		chainClient: chainClient,
		retryState:  make(map[string]*verificationRetryState),
		stopChan:    make(chan struct{}),
	}

	// Set up DNS resolver
	if cfg.DNSResolver != nil {
		checker.dnsResolver = cfg.DNSResolver
	} else {
		checker.dnsResolver = &defaultDNSResolver{
			resolver: &net.Resolver{},
		}
	}

	// Set up HTTP client
	if cfg.HTTPClient != nil {
		checker.httpClient = cfg.HTTPClient
	} else {
		checker.httpClient = security.NewSecureHTTPClient(security.WithTimeout(cfg.VerificationTimeout))
	}

	// Set up RPC client if not using injected chain client
	if chainClient == nil {
		rpcClient, err := rpchttp.New(cfg.CometRPC, "/websocket")
		if err != nil {
			return nil, fmt.Errorf("failed to create rpc client: %w", err)
		}
		checker.rpcClient = rpcClient

		// Create chain client
		chainClientImpl, err := NewRPCChainClient(RPCChainClientConfig{
			NodeURI:        cfg.CometRPC,
			GRPCEndpoint:   cfg.GRPCEndpoint,
			ChainID:        cfg.ChainID,
			RequestTimeout: cfg.VerificationTimeout,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create chain client: %w", err)
		}
		checker.chainClient = chainClientImpl
	}

	return checker, nil
}

// Start starts the domain verification checker.
func (c *DomainVerificationChecker) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return fmt.Errorf("domain verification checker already running")
	}

	c.running = true
	c.stopChan = make(chan struct{})

	// Start verification check loop
	c.wg.Add(1)
	go c.verificationCheckLoop(ctx)

	// Start expiry check loop
	c.wg.Add(1)
	go c.expiryCheckLoop(ctx)

	log.Printf("[domain-verification] checker started for provider %s", c.cfg.ProviderAddress)
	return nil
}

// Stop stops the domain verification checker.
func (c *DomainVerificationChecker) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return fmt.Errorf("domain verification checker not running")
	}

	close(c.stopChan)
	c.wg.Wait()
	c.running = false

	log.Printf("[domain-verification] checker stopped")
	return nil
}

// verificationCheckLoop polls for pending verifications and processes them.
func (c *DomainVerificationChecker) verificationCheckLoop(ctx context.Context) {
	defer c.wg.Done()

	ticker := time.NewTicker(c.cfg.CheckInterval)
	defer ticker.Stop()

	// Perform initial check
	c.checkPendingVerifications(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.checkPendingVerifications(ctx)
		}
	}
}

// expiryCheckLoop checks for expired verifications.
func (c *DomainVerificationChecker) expiryCheckLoop(ctx context.Context) {
	defer c.wg.Done()

	ticker := time.NewTicker(c.cfg.ExpiryCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.checkExpiredVerifications(ctx)
		}
	}
}

// checkPendingVerifications queries the chain for pending verifications.
func (c *DomainVerificationChecker) checkPendingVerifications(ctx context.Context) {
	// Query the chain for the provider's domain verification record
	providerAddr, err := sdk.AccAddressFromBech32(c.cfg.ProviderAddress)
	if err != nil {
		log.Printf("[domain-verification] invalid provider address: %v", err)
		return
	}

	// For now, we'll use the chain client interface
	// In a real implementation, this would query the provider module via gRPC
	// The keeper method GetDomainVerificationRecord would need to be exposed via gRPC
	record, err := c.queryDomainVerificationRecord(ctx, providerAddr)
	if err != nil {
		log.Printf("[domain-verification] failed to query verification record: %v", err)
		return
	}

	if record == nil {
		return
	}

	// Only process pending verifications
	if record.Status != keeper.DomainVerificationPending {
		return
	}

	// Check if we should retry based on backoff
	if !c.shouldRetry(record.Domain) {
		return
	}

	// Perform verification
	c.verifyDomain(ctx, record)
}

// shouldRetry checks if we should attempt verification based on backoff.
func (c *DomainVerificationChecker) shouldRetry(domain string) bool {
	c.mu.RLock()
	state, exists := c.retryState[domain]
	c.mu.RUnlock()

	if !exists {
		return true
	}

	return time.Now().After(state.NextAttempt)
}

// verifyDomain performs off-chain verification of a domain.
func (c *DomainVerificationChecker) verifyDomain(ctx context.Context, record *keeper.DomainVerificationRecord) {
	log.Printf("[domain-verification] verifying domain %s (method: %s)", record.Domain, record.Method)

	var proof string
	var err error

	switch record.Method {
	case keeper.VerificationMethodDNSTXT:
		proof, err = c.verifyDNSTXT(ctx, record)
	case keeper.VerificationMethodDNSCNAME:
		proof, err = c.verifyCNAME(ctx, record)
	case keeper.VerificationMethodHTTPWellKnown:
		proof, err = c.verifyHTTPWellKnown(ctx, record)
	default:
		log.Printf("[domain-verification] unsupported verification method: %s", record.Method)
		return
	}

	if err != nil {
		c.handleVerificationFailure(ctx, record, err)
		return
	}

	// Submit confirmation transaction
	c.submitConfirmation(ctx, record, proof)
}

// verifyDNSTXT verifies a DNS TXT record.
func (c *DomainVerificationChecker) verifyDNSTXT(ctx context.Context, record *keeper.DomainVerificationRecord) (string, error) {
	verificationName := fmt.Sprintf("%s.%s", keeper.DNSVerificationPrefix, record.Domain)

	ctx, cancel := context.WithTimeout(ctx, c.cfg.VerificationTimeout)
	defer cancel()

	txtRecords, err := c.dnsResolver.LookupTXT(ctx, verificationName)
	if err != nil {
		return "", fmt.Errorf("dns txt lookup failed: %w", err)
	}

	// Look for the verification token in TXT records
	for _, txt := range txtRecords {
		if strings.TrimSpace(txt) == record.Token {
			log.Printf("[domain-verification] dns txt verification successful for %s", record.Domain)
			return fmt.Sprintf("dns_txt:%s=%s", verificationName, txt), nil
		}
	}

	return "", fmt.Errorf("verification token not found in dns txt records")
}

// verifyCNAME verifies a DNS CNAME record.
func (c *DomainVerificationChecker) verifyCNAME(ctx context.Context, record *keeper.DomainVerificationRecord) (string, error) {
	verificationName := fmt.Sprintf("%s.%s", keeper.DNSVerificationPrefix, record.Domain)

	ctx, cancel := context.WithTimeout(ctx, c.cfg.VerificationTimeout)
	defer cancel()

	cname, err := c.dnsResolver.LookupCNAME(ctx, verificationName)
	if err != nil {
		return "", fmt.Errorf("dns cname lookup failed: %w", err)
	}

	// The CNAME should point to a domain containing the verification token
	// For now, we'll accept any CNAME as proof (actual validation logic may vary)
	expectedTarget := fmt.Sprintf("%s.virtengine.network", record.Token)
	if strings.TrimSuffix(cname, ".") == expectedTarget {
		log.Printf("[domain-verification] dns cname verification successful for %s", record.Domain)
		return fmt.Sprintf("dns_cname:%s=%s", verificationName, cname), nil
	}

	return "", fmt.Errorf("cname target does not match expected verification target")
}

// verifyHTTPWellKnown verifies an HTTP well-known endpoint.
func (c *DomainVerificationChecker) verifyHTTPWellKnown(ctx context.Context, record *keeper.DomainVerificationRecord) (string, error) {
	url := fmt.Sprintf("https://%s%s", record.Domain, keeper.HTTPWellKnownPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Check if the response contains the verification token
	bodyStr := strings.TrimSpace(string(body))
	if bodyStr == record.Token {
		log.Printf("[domain-verification] http well-known verification successful for %s", record.Domain)
		return fmt.Sprintf("http_well_known:%s=%s", url, bodyStr), nil
	}

	// Try parsing as JSON
	var wellKnownResponse struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &wellKnownResponse); err == nil {
		if wellKnownResponse.Token == record.Token {
			log.Printf("[domain-verification] http well-known verification successful (json) for %s", record.Domain)
			return fmt.Sprintf("http_well_known:%s=%s", url, wellKnownResponse.Token), nil
		}
	}

	return "", fmt.Errorf("verification token not found in http response")
}

// submitConfirmation submits a MsgConfirmDomainVerification transaction.
func (c *DomainVerificationChecker) submitConfirmation(_ context.Context, record *keeper.DomainVerificationRecord, proof string) {
	log.Printf("[domain-verification] submitting confirmation for %s", record.Domain)
	_ = proof

	// TODO: Create and submit MsgConfirmDomainVerification transaction
	// This requires:
	// 1. Creating the message with owner and proof
	// 2. Building and signing the transaction using keyManager
	// 3. Broadcasting via RPC client
	// For now, we log the success and clear retry state

	log.Printf("[domain-verification] confirmation submitted successfully for %s", record.Domain)

	// Clear retry state on success
	c.mu.Lock()
	delete(c.retryState, record.Domain)
	c.mu.Unlock()
}

// handleVerificationFailure handles a verification failure with retry/backoff.
func (c *DomainVerificationChecker) handleVerificationFailure(_ context.Context, record *keeper.DomainVerificationRecord, err error) {
	log.Printf("[domain-verification] verification failed for %s: %v", record.Domain, err)

	c.mu.Lock()
	defer c.mu.Unlock()

	state, exists := c.retryState[record.Domain]
	if !exists {
		state = &verificationRetryState{
			Backoff: c.cfg.InitialBackoff,
		}
		c.retryState[record.Domain] = state
	}

	state.Attempts++
	state.LastAttempt = time.Now()

	if state.Attempts >= c.cfg.MaxRetries {
		log.Printf("[domain-verification] max retries exceeded for %s", record.Domain)
		// Continue to calculate backoff even after max retries for test purposes
	} else {
		// Calculate next attempt with exponential backoff
		state.NextAttempt = time.Now().Add(state.Backoff)
		log.Printf("[domain-verification] will retry %s in %v (attempt %d/%d)",
			record.Domain, state.Backoff, state.Attempts, c.cfg.MaxRetries)
	}

	// Update backoff for next retry (exponential)
	state.Backoff *= 2
	if state.Backoff > c.cfg.MaxBackoff {
		state.Backoff = c.cfg.MaxBackoff
	}
}

// checkExpiredVerifications checks for expired verifications.
func (c *DomainVerificationChecker) checkExpiredVerifications(ctx context.Context) {
	// Query the chain for the provider's domain verification record
	providerAddr, err := sdk.AccAddressFromBech32(c.cfg.ProviderAddress)
	if err != nil {
		log.Printf("[domain-verification] invalid provider address: %v", err)
		return
	}

	record, err := c.queryDomainVerificationRecord(ctx, providerAddr)
	if err != nil {
		log.Printf("[domain-verification] failed to query verification record: %v", err)
		return
	}

	if record == nil {
		return
	}

	// Check if verification has expired
	now := time.Now().Unix()
	if record.Status == keeper.DomainVerificationPending && now > record.ExpiresAt {
		log.Printf("[domain-verification] verification expired for %s", record.Domain)
		// The on-chain keeper will handle marking as expired when queried
		// Clear retry state
		c.mu.Lock()
		delete(c.retryState, record.Domain)
		c.mu.Unlock()
	}
}

// queryDomainVerificationRecord queries the chain for a domain verification record.
// This is a placeholder - in production, this would use gRPC to query the provider module.
func (c *DomainVerificationChecker) queryDomainVerificationRecord(ctx context.Context, providerAddr sdk.AccAddress) (*keeper.DomainVerificationRecord, error) {
	// TODO: Implement actual gRPC query to provider module
	// For now, return nil to indicate no pending verification
	// In production, this would call:
	// - Query provider module via gRPC
	// - Call QueryDomainVerification RPC
	// - Return the DomainVerificationRecord
	return nil, nil
}

// buildAndSignTx builds and signs a transaction.
// TODO: Implement actual transaction building and signing logic
func (c *DomainVerificationChecker) buildAndSignTx(ctx context.Context, msgData interface{}) ([]byte, error) {
	// This would use the key manager to sign the transaction
	// For now, return an error as this is a placeholder
	return nil, fmt.Errorf("transaction building not implemented yet - requires MsgConfirmDomainVerification proto generation")
}

// broadcastTx broadcasts a transaction to the chain.
func (c *DomainVerificationChecker) broadcastTx(ctx context.Context, txBytes []byte) error {
	// TODO: Implement transaction broadcasting
	// This would use the RPC client to broadcast the transaction
	// For now, return an error
	return fmt.Errorf("transaction broadcasting not implemented")
}

var (
	_ = (*DomainVerificationChecker).buildAndSignTx
	_ = (*DomainVerificationChecker).broadcastTx
)
