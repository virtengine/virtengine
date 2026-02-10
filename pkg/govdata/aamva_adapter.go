// Package govdata provides government data source integration for identity verification.
//
// VE-2006: Real AAMVA DMV API adapter implementation
package govdata

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/security"
)

// ============================================================================
// AAMVA Constants
// ============================================================================

// AAMVA Environments
const (
	AAMVAEnvironmentProduction = "production"
	AAMVAEnvironmentSandbox    = "sandbox"
)

// AAMVA API Endpoints
const (
	AAMVAProductionURL = "https://api.aamva.org/dldv/v1"
	AAMVASandboxURL    = "https://sandbox.aamva.org/dldv/v1"
)

// AAMVA DLDV Transaction Types
const (
	// AAMVATransactionDLDV is the standard Driver License Data Verification
	AAMVATransactionDLDV = "DLDV"

	// AAMVATransactionDLDVP is Driver License Data Verification Plus (with photo)
	AAMVATransactionDLDVP = "DLDVP"
)

// ============================================================================
// AAMVA Errors
// ============================================================================

var (
	// ErrAAMVANotConfigured is returned when AAMVA is not configured
	ErrAAMVANotConfigured = errors.New("AAMVA API not configured")

	// ErrAAMVAAuthentication is returned on auth failure
	ErrAAMVAAuthentication = errors.New("AAMVA authentication failed")

	// ErrAAMVARateLimit is returned when rate limited
	ErrAAMVARateLimit = errors.New("AAMVA API rate limit exceeded")

	// ErrAAMVATimeout is returned on timeout
	ErrAAMVATimeout = errors.New("AAMVA API request timed out")

	// ErrAAMVAStateNotSupported is returned for unsupported jurisdictions
	ErrAAMVAStateNotSupported = errors.New("jurisdiction not supported by AAMVA DLDV")

	// ErrAAMVAInvalidResponse is returned for malformed responses
	ErrAAMVAInvalidResponse = errors.New("invalid AAMVA API response")

	// ErrAAMVARequestFailed is returned for general request failures
	ErrAAMVARequestFailed = errors.New("AAMVA API request failed")
)

// ============================================================================
// AAMVA Configuration
// ============================================================================

// AAMVAConfig contains AAMVA API configuration
type AAMVAConfig struct {
	// Environment is "production" or "sandbox"
	Environment string `json:"environment"`

	// OrgID is the AAMVA organization ID
	OrgID string `json:"org_id"`

	// PermissionCode is the AAMVA permission code
	PermissionCode string `json:"permission_code"`

	// ClientID is the OAuth client ID
	ClientID string `json:"client_id"`

	// ClientSecret is the OAuth client secret (NEVER log this)
	ClientSecret string `json:"-"`

	// APIKey is the API key (NEVER log this)
	APIKey string `json:"-"`

	// Timeout is the request timeout
	Timeout time.Duration `json:"timeout"`

	// MaxRetries is the maximum number of retries
	MaxRetries int `json:"max_retries"`

	// RetryBackoff is the base backoff duration for retries
	RetryBackoff time.Duration `json:"retry_backoff"`

	// RateLimitPerMinute is the max requests per minute
	RateLimitPerMinute int `json:"rate_limit_per_minute"`

	// BaseURL overrides the default AAMVA API URL (used for proxies/tests)
	BaseURL string `json:"base_url"`

	// SupportedStates lists the states enabled for DLDV
	SupportedStates []string `json:"supported_states"`

	// EnableDLDVPlus enables photo verification
	EnableDLDVPlus bool `json:"enable_dldv_plus"`

	// AuditEnabled enables detailed audit logging
	AuditEnabled bool `json:"audit_enabled"`
}

// DefaultAAMVAConfig returns default AAMVA configuration
func DefaultAAMVAConfig() AAMVAConfig {
	return AAMVAConfig{
		Environment:        AAMVAEnvironmentSandbox,
		Timeout:            30 * time.Second,
		MaxRetries:         3,
		RetryBackoff:       500 * time.Millisecond,
		RateLimitPerMinute: 60,
		AuditEnabled:       true,
		// All US states and territories
		SupportedStates: []string{
			"AL", "AK", "AZ", "AR", "CA", "CO", "CT", "DE", "FL", "GA",
			"HI", "ID", "IL", "IN", "IA", "KS", "KY", "LA", "ME", "MD",
			"MA", "MI", "MN", "MS", "MO", "MT", "NE", "NV", "NH", "NJ",
			"NM", "NY", "NC", "ND", "OH", "OK", "OR", "PA", "RI", "SC",
			"SD", "TN", "TX", "UT", "VT", "VA", "WA", "WV", "WI", "WY",
			"DC", "PR", "VI", "GU", "AS", "MP",
		},
	}
}

// Validate validates the AAMVA configuration
func (c *AAMVAConfig) Validate() error {
	if c.OrgID == "" {
		return errors.New("AAMVA org_id is required")
	}
	if c.PermissionCode == "" {
		return errors.New("AAMVA permission_code is required")
	}
	if c.ClientID == "" {
		return errors.New("AAMVA client_id is required")
	}
	if c.ClientSecret == "" {
		return errors.New("AAMVA client_secret is required")
	}
	if c.Environment != AAMVAEnvironmentProduction && c.Environment != AAMVAEnvironmentSandbox {
		return fmt.Errorf("invalid AAMVA environment: %s", c.Environment)
	}
	return nil
}

// ============================================================================
// AAMVA Request/Response Types
// ============================================================================

// AAMVADLDVRequest represents a DLDV verification request
type AAMVADLDVRequest struct {
	XMLName     xml.Name `xml:"dldvRequest"`
	OrgID       string   `xml:"orgId"`
	Permission  string   `xml:"permissionCode"`
	Transaction string   `xml:"transactionType"`
	MessageID   string   `xml:"messageId"`
	Timestamp   string   `xml:"timestamp"`

	// Driver Information
	State         string `xml:"state"`
	LicenseNumber string `xml:"licenseNumber"`
	FirstName     string `xml:"firstName,omitempty"`
	MiddleName    string `xml:"middleName,omitempty"`
	LastName      string `xml:"lastName,omitempty"`
	DateOfBirth   string `xml:"dateOfBirth,omitempty"`

	// Optional Address Verification
	AddressLine1 string `xml:"addressLine1,omitempty"`
	AddressLine2 string `xml:"addressLine2,omitempty"`
	City         string `xml:"city,omitempty"`
	ZipCode      string `xml:"zipCode,omitempty"`

	// Optional Photo Request (DLDV Plus)
	RequestPhoto bool `xml:"requestPhoto,omitempty"`
}

// AAMVADLDVResponse represents a DLDV verification response
type AAMVADLDVResponse struct {
	XMLName      xml.Name `xml:"dldvResponse"`
	MessageID    string   `xml:"messageId"`
	ResponseCode string   `xml:"responseCode"`
	ResponseText string   `xml:"responseText"`
	Timestamp    string   `xml:"timestamp"`

	// Match Results
	OverallMatch    string `xml:"overallMatch"`
	LicenseMatch    string `xml:"licenseMatch"`
	FirstNameMatch  string `xml:"firstNameMatch"`
	MiddleNameMatch string `xml:"middleNameMatch"`
	LastNameMatch   string `xml:"lastNameMatch"`
	DOBMatch        string `xml:"dobMatch"`
	AddressMatch    string `xml:"addressMatch"`

	// License Status
	LicenseStatus  string `xml:"licenseStatus"`
	LicenseClass   string `xml:"licenseClass"`
	ExpirationDate string `xml:"expirationDate"`
	IssuedDate     string `xml:"issuedDate"`

	// Restrictions/Endorsements
	Restrictions string `xml:"restrictions,omitempty"`
	Endorsements string `xml:"endorsements,omitempty"`

	// Photo (DLDV Plus only)
	PhotoBase64 string `xml:"photo,omitempty"`

	// Error Information
	ErrorCode    string `xml:"errorCode,omitempty"`
	ErrorMessage string `xml:"errorMessage,omitempty"`
}

// AAMVAMatchResult represents match outcome values
type AAMVAMatchResult string

const (
	AAMVAMatchYes     AAMVAMatchResult = "Y" // Exact match
	AAMVAMatchNo      AAMVAMatchResult = "N" // No match
	AAMVAMatchPartial AAMVAMatchResult = "P" // Partial match
	AAMVAMatchNoData  AAMVAMatchResult = "U" // Unable to verify (no data)
)

// AAMVALicenseStatus represents license status values
type AAMVALicenseStatus string

const (
	AAMVALicenseValid     AAMVALicenseStatus = "VALID"
	AAMVALicenseExpired   AAMVALicenseStatus = "EXPIRED"
	AAMVALicenseSuspended AAMVALicenseStatus = "SUSPENDED"
	AAMVALicenseRevoked   AAMVALicenseStatus = "REVOKED"
	AAMVALicenseCanceled  AAMVALicenseStatus = "CANCELED"
)

// ============================================================================
// AAMVA DMV Adapter Implementation
// ============================================================================

// AAMVADMVAdapter implements DataSourceAdapter for AAMVA DLDV
type AAMVADMVAdapter struct {
	*baseAdapter
	config      AAMVAConfig
	httpClient  *http.Client
	accessToken string
	tokenExpiry time.Time
	mu          sync.RWMutex

	// Rate limiting
	requestCount int
	windowStart  time.Time
}

// NewAAMVADMVAdapter creates a new AAMVA DMV adapter
func NewAAMVADMVAdapter(baseConfig AdapterConfig, aamvaConfig AAMVAConfig) (*AAMVADMVAdapter, error) {
	if aamvaConfig.Timeout <= 0 {
		if baseConfig.Timeout > 0 {
			aamvaConfig.Timeout = baseConfig.Timeout
		} else {
			aamvaConfig.Timeout = 30 * time.Second
		}
	}
	if aamvaConfig.RetryBackoff <= 0 {
		aamvaConfig.RetryBackoff = 500 * time.Millisecond
	}
	if aamvaConfig.RateLimitPerMinute <= 0 {
		aamvaConfig.RateLimitPerMinute = 60
	}

	if err := aamvaConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid AAMVA config: %w", err)
	}

	if len(baseConfig.SupportedDocuments) == 0 {
		baseConfig.SupportedDocuments = []DocumentType{
			DocumentTypeDriversLicense,
			DocumentTypeStateID,
		}
	}

	adapter := &AAMVADMVAdapter{
		baseAdapter: newBaseAdapter(baseConfig),
		config:      aamvaConfig,
		httpClient:  security.NewSecureHTTPClient(security.WithTimeout(aamvaConfig.Timeout)),
		windowStart: time.Now(),
	}

	return adapter, nil
}

// baseURL returns the AAMVA API base URL
func (a *AAMVADMVAdapter) baseURL() string {
	if a.config.BaseURL != "" {
		return strings.TrimRight(a.config.BaseURL, "/")
	}
	if a.config.Environment == AAMVAEnvironmentProduction {
		return AAMVAProductionURL
	}
	return AAMVASandboxURL
}

func (a *AAMVADMVAdapter) logf(format string, args ...interface{}) {
	if !a.config.AuditEnabled {
		return
	}
	log.Printf("[govdata/aamva] "+format, args...)
}

// authenticate obtains an OAuth access token
func (a *AAMVADMVAdapter) authenticate(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Check if token is still valid
	if a.accessToken != "" && time.Now().Before(a.tokenExpiry) {
		return nil
	}

	// Request new token
	authURL := a.baseURL() + "/oauth/token"

	req, err := http.NewRequestWithContext(ctx, "POST", authURL, strings.NewReader(
		fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s",
			a.config.ClientID, a.config.ClientSecret)))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if a.config.APIKey != "" {
		req.Header.Set("X-API-Key", a.config.APIKey)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAAMVAAuthentication, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		httpErr := &aamvaHTTPError{statusCode: resp.StatusCode}
		return fmt.Errorf("%w: %v", ErrAAMVAAuthentication, httpErr)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to parse auth response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return fmt.Errorf("%w: empty access token", ErrAAMVAAuthentication)
	}

	a.accessToken = tokenResp.AccessToken
	// Set expiry with buffer for safety
	bufferedExpiry := tokenResp.ExpiresIn - 60
	if bufferedExpiry < 0 {
		bufferedExpiry = 0
	}
	a.tokenExpiry = time.Now().Add(time.Duration(bufferedExpiry) * time.Second)
	a.logf("auth token refreshed expires_in=%ds", tokenResp.ExpiresIn)

	return nil
}

// checkRateLimit checks and updates rate limiting
func (a *AAMVADMVAdapter) checkRateLimit() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()

	// Reset window if expired
	if now.Sub(a.windowStart) > time.Minute {
		a.windowStart = now
		a.requestCount = 0
	}

	if a.requestCount >= a.config.RateLimitPerMinute {
		return ErrAAMVARateLimit
	}

	a.requestCount++
	return nil
}

// isStateSupported checks if a state is supported
func (a *AAMVADMVAdapter) isStateSupported(state string) bool {
	state = strings.ToUpper(state)
	for _, s := range a.config.SupportedStates {
		if s == state {
			return true
		}
	}
	return false
}

// validateLicenseNumber validates license number format for a state
func (a *AAMVADMVAdapter) validateLicenseNumber(state, licenseNumber string) error {
	// State-specific license number validation patterns
	patterns := map[string]*regexp.Regexp{
		"CA": regexp.MustCompile(`^[A-Z]\d{7}$`),     // California: 1 letter + 7 digits
		"TX": regexp.MustCompile(`^\d{8}$`),          // Texas: 8 digits
		"FL": regexp.MustCompile(`^[A-Z]\d{12}$`),    // Florida: 1 letter + 12 digits
		"NY": regexp.MustCompile(`^\d{9}$`),          // New York: 9 digits
		"PA": regexp.MustCompile(`^\d{8}$`),          // Pennsylvania: 8 digits
		"IL": regexp.MustCompile(`^[A-Z]\d{11,12}$`), // Illinois: 1 letter + 11-12 digits
		"OH": regexp.MustCompile(`^[A-Z]{2}\d{6}$`),  // Ohio: 2 letters + 6 digits
		"GA": regexp.MustCompile(`^\d{9}$`),          // Georgia: 9 digits
		"NC": regexp.MustCompile(`^\d{1,12}$`),       // North Carolina: 1-12 digits
		"MI": regexp.MustCompile(`^[A-Z]\d{10,12}$`), // Michigan: 1 letter + 10-12 digits
	}

	// If no specific pattern, use general validation
	if pattern, ok := patterns[strings.ToUpper(state)]; ok {
		if !pattern.MatchString(strings.ToUpper(licenseNumber)) {
			return fmt.Errorf("invalid license number format for %s", state)
		}
	} else {
		// General pattern: alphanumeric, 5-20 characters
		if len(licenseNumber) < 5 || len(licenseNumber) > 20 {
			return ErrInvalidDocumentNumber
		}
	}

	return nil
}

// generateMessageID generates a unique message ID for AAMVA
func (a *AAMVADMVAdapter) generateMessageID(requestID string) string {
	// Create HMAC of request ID + timestamp for unique, reproducible ID
	h := hmac.New(sha256.New, []byte(a.config.OrgID))
	h.Write([]byte(requestID + time.Now().Format(time.RFC3339Nano)))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))[:32]
}

type aamvaHTTPError struct {
	statusCode int
	retryAfter time.Duration
}

func (e *aamvaHTTPError) Error() string {
	if e.retryAfter > 0 {
		return fmt.Sprintf("AAMVA HTTP error: status %d (retry-after %s)", e.statusCode, e.retryAfter)
	}
	return fmt.Sprintf("AAMVA HTTP error: status %d", e.statusCode)
}

// Verify performs AAMVA DLDV verification
func (a *AAMVADMVAdapter) Verify(ctx context.Context, req *VerificationRequest) (*VerificationResponse, error) {
	startTime := time.Now()

	// Validate document type
	if !a.SupportsDocument(req.DocumentType) {
		return nil, ErrDocumentTypeNotSupported
	}

	// Extract state from jurisdiction (e.g., "US-CA" -> "CA")
	state := strings.TrimPrefix(req.Jurisdiction, "US-")
	if len(state) != 2 {
		return nil, fmt.Errorf("invalid jurisdiction format: %s", req.Jurisdiction)
	}

	// Check state support
	if !a.isStateSupported(state) {
		return nil, ErrAAMVAStateNotSupported
	}

	// Validate license number format
	if err := a.validateLicenseNumber(state, req.Fields.DocumentNumber); err != nil {
		return nil, err
	}

	// Build AAMVA request
	transType := AAMVATransactionDLDV
	if a.config.EnableDLDVPlus {
		transType = AAMVATransactionDLDVP
	}

	baseReq := &AAMVADLDVRequest{
		OrgID:         a.config.OrgID,
		Permission:    a.config.PermissionCode,
		Transaction:   transType,
		State:         state,
		LicenseNumber: req.Fields.DocumentNumber,
		FirstName:     req.Fields.FirstName,
		MiddleName:    req.Fields.MiddleName,
		LastName:      req.Fields.LastName,
		RequestPhoto:  a.config.EnableDLDVPlus,
	}

	if !req.Fields.DateOfBirth.IsZero() {
		baseReq.DateOfBirth = req.Fields.DateOfBirth.Format("2006-01-02")
	}

	// Address verification if provided
	if req.Fields.Address != nil && req.Fields.Address.Street != "" {
		baseReq.AddressLine1 = req.Fields.Address.Street
		baseReq.City = req.Fields.Address.City
		baseReq.ZipCode = req.Fields.Address.PostalCode
	}

	var lastErr error
	for attempt := 0; attempt <= a.config.MaxRetries; attempt++ {
		if err := a.checkRateLimit(); err != nil {
			return nil, err
		}

		if err := a.authenticate(ctx); err != nil {
			a.recordError(err)
			if a.shouldRetry(ctx, err, 0) && attempt < a.config.MaxRetries {
				a.logf("auth retrying attempt=%d request_id=%s jurisdiction=%s", attempt+1, req.RequestID, req.Jurisdiction)
				if err := a.sleepWithBackoff(ctx, attempt, 0, 0); err != nil {
					return nil, err
				}
				continue
			}
			return nil, err
		}

		dldvReq := *baseReq
		dldvReq.MessageID = a.generateMessageID(req.RequestID)
		dldvReq.Timestamp = time.Now().UTC().Format(time.RFC3339)

		xmlData, err := xml.MarshalIndent(&dldvReq, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal DLDV request: %w", err)
		}

		a.logf("verify attempt=%d request_id=%s jurisdiction=%s document_type=%s message_id=%s",
			attempt+1, req.RequestID, req.Jurisdiction, req.DocumentType, dldvReq.MessageID)

		dldvResp, statusCode, err := a.doVerifyRequest(ctx, xmlData, dldvReq.MessageID)
		if err != nil {
			lastErr = err
			a.recordError(err)

			if a.shouldRetry(ctx, err, statusCode) && attempt < a.config.MaxRetries {
				a.logf("verify retrying attempt=%d request_id=%s jurisdiction=%s error=%v", attempt+1, req.RequestID, req.Jurisdiction, err)
				retryAfter := a.retryAfterDuration(err)
				if err := a.sleepWithBackoff(ctx, attempt, statusCode, retryAfter); err != nil {
					return nil, err
				}
				continue
			}

			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				return nil, ErrAAMVATimeout
			}
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				return nil, ErrAAMVATimeout
			}
			if statusCode == http.StatusTooManyRequests {
				return nil, ErrAAMVARateLimit
			}
			if statusCode == http.StatusUnauthorized {
				return nil, ErrAAMVAAuthentication
			}
			return nil, fmt.Errorf("%w: %v", ErrAAMVARequestFailed, err)
		}

		if !a.isResponseUsable(dldvResp) {
			err := fmt.Errorf("%w: response_code=%s error_code=%s", ErrAAMVARequestFailed, dldvResp.ResponseCode, dldvResp.ErrorCode)
			a.recordError(err)
			lastErr = err
			if a.shouldRetry(ctx, err, statusCode) && attempt < a.config.MaxRetries {
				if err := a.sleepWithBackoff(ctx, attempt, statusCode, 0); err != nil {
					return nil, err
				}
				continue
			}
			return nil, err
		}

		verificationResp := a.convertResponse(req, dldvResp)
		if dldvResp.ResponseCode != "" && dldvResp.ResponseCode != "0000" {
			if dldvResp.ResponseText != "" {
				verificationResp.Warnings = append(verificationResp.Warnings, "AAMVA response: "+strings.TrimSpace(dldvResp.ResponseText))
			} else {
				verificationResp.Warnings = append(verificationResp.Warnings, "AAMVA response code: "+dldvResp.ResponseCode)
			}
		}

		latency := time.Since(startTime)
		a.recordSuccess(latency)
		a.logf("verify success request_id=%s jurisdiction=%s status=%s response_code=%s latency_ms=%d",
			req.RequestID, req.Jurisdiction, verificationResp.Status, dldvResp.ResponseCode, latency.Milliseconds())

		return verificationResp, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, ErrAAMVARequestFailed
}

func (a *AAMVADMVAdapter) doVerifyRequest(ctx context.Context, xmlData []byte, messageID string) (*AAMVADLDVResponse, int, error) {
	apiURL := a.baseURL() + "/verify"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(xmlData))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	a.mu.RLock()
	token := a.accessToken
	a.mu.RUnlock()

	httpReq.Header.Set("Content-Type", "application/xml")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("X-AAMVA-MessageID", messageID)
	if a.config.APIKey != "" {
		httpReq.Header.Set("X-API-Key", a.config.APIKey)
	}

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		if ctx.Err() != nil {
			return nil, 0, ctx.Err()
		}
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		a.mu.Lock()
		a.accessToken = ""
		a.mu.Unlock()
		return nil, resp.StatusCode, &aamvaHTTPError{statusCode: resp.StatusCode}
	}

	if resp.StatusCode != http.StatusOK {
		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil, resp.StatusCode, &aamvaHTTPError{statusCode: resp.StatusCode, retryAfter: retryAfter}
	}

	var dldvResp AAMVADLDVResponse
	if err := xml.NewDecoder(resp.Body).Decode(&dldvResp); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("%w: %v", ErrAAMVAInvalidResponse, err)
	}

	return &dldvResp, resp.StatusCode, nil
}

func (a *AAMVADMVAdapter) isResponseUsable(resp *AAMVADLDVResponse) bool {
	if resp == nil {
		return false
	}
	if resp.OverallMatch != "" || resp.LicenseMatch != "" || resp.FirstNameMatch != "" || resp.LastNameMatch != "" || resp.DOBMatch != "" {
		return true
	}
	if resp.ResponseCode == "" && resp.ErrorCode == "" {
		return false
	}
	if resp.ResponseCode != "" && resp.ResponseCode != "0000" && resp.ErrorCode != "" {
		return false
	}
	return true
}

func (a *AAMVADMVAdapter) shouldRetry(ctx context.Context, err error, statusCode int) bool {
	if ctx.Err() != nil {
		return false
	}
	if statusCode == http.StatusTooManyRequests || statusCode == http.StatusUnauthorized {
		return true
	}
	if statusCode >= 500 && statusCode <= 599 {
		return true
	}
	var httpErr *aamvaHTTPError
	if errors.As(err, &httpErr) {
		if httpErr.statusCode == http.StatusTooManyRequests || httpErr.statusCode == http.StatusUnauthorized {
			return true
		}
		if httpErr.statusCode >= 500 && httpErr.statusCode <= 599 {
			return true
		}
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}
	return false
}

func (a *AAMVADMVAdapter) retryAfterDuration(err error) time.Duration {
	var httpErr *aamvaHTTPError
	if errors.As(err, &httpErr) {
		return httpErr.retryAfter
	}
	return 0
}

func (a *AAMVADMVAdapter) sleepWithBackoff(ctx context.Context, attempt int, statusCode int, retryAfter time.Duration) error {
	if retryAfter > 0 {
		a.logf("retrying after server backoff=%s", retryAfter)
		timer := time.NewTimer(retryAfter)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return nil
		}
	}

	backoff := a.config.RetryBackoff * (1 << attempt)
	jitter := time.Duration(time.Now().UnixNano() % int64(backoff/2+1))
	sleepFor := backoff + jitter
	if sleepFor <= 0 {
		sleepFor = 100 * time.Millisecond
	}
	a.logf("retrying with backoff=%s status=%d", sleepFor, statusCode)

	timer := time.NewTimer(sleepFor)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// convertResponse converts AAMVA response to our VerificationResponse
func (a *AAMVADMVAdapter) convertResponse(req *VerificationRequest, dldvResp *AAMVADLDVResponse) *VerificationResponse {
	resp := &VerificationResponse{
		RequestID:      req.RequestID,
		DataSourceType: DataSourceDMV,
		Jurisdiction:   req.Jurisdiction,
		VerifiedAt:     time.Now(),
		FieldResults:   make(map[string]FieldVerificationResult),
	}

	if dldvResp.ResponseCode != "" && dldvResp.ResponseCode != "0000" {
		resp.ErrorCode = dldvResp.ResponseCode
		if dldvResp.ResponseText != "" {
			resp.ErrorMessage = strings.TrimSpace(dldvResp.ResponseText)
		}
	}
	if dldvResp.ErrorCode != "" {
		resp.ErrorCode = dldvResp.ErrorCode
		if dldvResp.ErrorMessage != "" {
			resp.ErrorMessage = strings.TrimSpace(dldvResp.ErrorMessage)
		}
	}

	// Determine overall status from AAMVA response
	switch AAMVAMatchResult(dldvResp.OverallMatch) {
	case AAMVAMatchYes:
		resp.Status = VerificationStatusVerified
		resp.DocumentValid = true
		resp.Confidence = 1.0
	case AAMVAMatchPartial:
		resp.Status = VerificationStatusPartialMatch
		resp.DocumentValid = true
		resp.Confidence = 0.7
	case AAMVAMatchNo:
		resp.Status = VerificationStatusNotVerified
		resp.DocumentValid = false
		resp.Confidence = 0.0
	case AAMVAMatchNoData:
		resp.Status = VerificationStatusNotFound
		resp.DocumentValid = false
		resp.Confidence = 0.0
		resp.Warnings = append(resp.Warnings, "Unable to verify - no data available")
	default:
		resp.Status = VerificationStatusNotVerified
		resp.DocumentValid = false
	}

	// Check license status
	switch AAMVALicenseStatus(dldvResp.LicenseStatus) {
	case AAMVALicenseValid:
		// OK
	case AAMVALicenseExpired:
		resp.Status = VerificationStatusExpired
		resp.DocumentValid = false
		resp.Warnings = append(resp.Warnings, "License has expired")
	case AAMVALicenseSuspended:
		resp.Status = VerificationStatusRevoked
		resp.DocumentValid = false
		resp.Warnings = append(resp.Warnings, "License is suspended")
	case AAMVALicenseRevoked:
		resp.Status = VerificationStatusRevoked
		resp.DocumentValid = false
		resp.Warnings = append(resp.Warnings, "License has been revoked")
	case AAMVALicenseCanceled:
		resp.Status = VerificationStatusRevoked
		resp.DocumentValid = false
		resp.Warnings = append(resp.Warnings, "License has been canceled")
	}

	// Parse expiration date
	if dldvResp.ExpirationDate != "" {
		if expDate, err := time.Parse("2006-01-02", dldvResp.ExpirationDate); err == nil {
			resp.DocumentExpiresAt = &expDate
			if expDate.Before(time.Now()) && resp.Status == VerificationStatusVerified {
				resp.Status = VerificationStatusExpired
				resp.DocumentValid = false
			}
		}
	}

	// Set verification expiry (typically 1 year for AAMVA verifications)
	resp.ExpiresAt = time.Now().Add(365 * 24 * time.Hour)

	// Convert field-level results
	resp.FieldResults["license_number"] = a.convertFieldMatch(dldvResp.LicenseMatch)
	resp.FieldResults["first_name"] = a.convertFieldMatch(dldvResp.FirstNameMatch)
	resp.FieldResults["middle_name"] = a.convertFieldMatch(dldvResp.MiddleNameMatch)
	resp.FieldResults["last_name"] = a.convertFieldMatch(dldvResp.LastNameMatch)
	resp.FieldResults["date_of_birth"] = a.convertFieldMatch(dldvResp.DOBMatch)
	resp.FieldResults["address"] = a.convertFieldMatch(dldvResp.AddressMatch)

	// Note: License class, restrictions, and endorsements are not stored
	// as they are additional metadata not needed for VEID verification
	// and would require schema changes to VerificationResponse

	return resp
}

// convertFieldMatch converts AAMVA match result to our format
func (a *AAMVADMVAdapter) convertFieldMatch(match string) FieldVerificationResult {
	result := FieldVerificationResult{}

	switch AAMVAMatchResult(match) {
	case AAMVAMatchYes:
		result.Match = FieldMatchExact
		result.Confidence = 1.0
	case AAMVAMatchPartial:
		result.Match = FieldMatchFuzzy
		result.Confidence = 0.7
	case AAMVAMatchNo:
		result.Match = FieldMatchNoMatch
		result.Confidence = 0.0
	case AAMVAMatchNoData:
		result.Match = FieldMatchUnavailable
		result.Confidence = 0.0
	default:
		result.Match = FieldMatchUnavailable
		result.Confidence = 0.0
	}

	return result
}

func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}
	if secs, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
		if secs <= 0 {
			return 0
		}
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(value); err == nil {
		if time.Until(t) > 0 {
			return time.Until(t)
		}
	}
	return 0
}

// loadAAMVAConfigFromEnv loads AAMVA configuration from environment variables.
// Returns (config, true, nil) if AAMVA is configured, (zero, false, nil) if not configured.
func loadAAMVAConfigFromEnv(baseConfig AdapterConfig) (AAMVAConfig, bool, error) {
	cfg := DefaultAAMVAConfig()

	get := func(key string) string {
		return strings.TrimSpace(os.Getenv(key))
	}

	configured := false
	enabledFlagSet := false

	if val := get("AAMVA_ENABLED"); val != "" {
		enabledFlagSet = true
		enabled, err := strconv.ParseBool(val)
		if err != nil {
			return AAMVAConfig{}, true, fmt.Errorf("invalid AAMVA_ENABLED: %w", err)
		}
		if !enabled {
			return AAMVAConfig{}, false, nil
		}
		configured = true
	}

	if val := get("AAMVA_ENVIRONMENT"); val != "" {
		cfg.Environment = val
	}
	if val := get("AAMVA_BASE_URL"); val != "" {
		cfg.BaseURL = val
	}
	if val := get("AAMVA_ORG_ID"); val != "" {
		cfg.OrgID = val
		configured = true
	}
	if val := get("AAMVA_PERMISSION_CODE"); val != "" {
		cfg.PermissionCode = val
		configured = true
	}
	if val := get("AAMVA_CLIENT_ID"); val != "" {
		cfg.ClientID = val
		configured = true
	}
	if val := get("AAMVA_CLIENT_SECRET"); val != "" {
		cfg.ClientSecret = val
		configured = true
	}
	if val := get("AAMVA_API_KEY"); val != "" {
		cfg.APIKey = val
	}
	if val := get("AAMVA_MAX_RETRIES"); val != "" {
		parsed, err := strconv.Atoi(val)
		if err != nil {
			return AAMVAConfig{}, true, fmt.Errorf("invalid AAMVA_MAX_RETRIES: %w", err)
		}
		cfg.MaxRetries = parsed
	}
	if val := get("AAMVA_RATE_LIMIT_PER_MINUTE"); val != "" {
		parsed, err := strconv.Atoi(val)
		if err != nil {
			return AAMVAConfig{}, true, fmt.Errorf("invalid AAMVA_RATE_LIMIT_PER_MINUTE: %w", err)
		}
		cfg.RateLimitPerMinute = parsed
	}
	if val := get("AAMVA_TIMEOUT"); val != "" {
		if dur, err := time.ParseDuration(val); err == nil {
			cfg.Timeout = dur
		} else if secs, err := strconv.Atoi(val); err == nil {
			cfg.Timeout = time.Duration(secs) * time.Second
		} else {
			return AAMVAConfig{}, true, fmt.Errorf("invalid AAMVA_TIMEOUT: %w", err)
		}
	}
	if val := get("AAMVA_RETRY_BACKOFF"); val != "" {
		if dur, err := time.ParseDuration(val); err == nil {
			cfg.RetryBackoff = dur
		} else if ms, err := strconv.Atoi(val); err == nil {
			cfg.RetryBackoff = time.Duration(ms) * time.Millisecond
		} else {
			return AAMVAConfig{}, true, fmt.Errorf("invalid AAMVA_RETRY_BACKOFF: %w", err)
		}
	}
	if val := get("AAMVA_ENABLE_DLDV_PLUS"); val != "" {
		parsed, err := strconv.ParseBool(val)
		if err != nil {
			return AAMVAConfig{}, true, fmt.Errorf("invalid AAMVA_ENABLE_DLDV_PLUS: %w", err)
		}
		cfg.EnableDLDVPlus = parsed
	}
	if val := get("AAMVA_AUDIT_ENABLED"); val != "" {
		parsed, err := strconv.ParseBool(val)
		if err != nil {
			return AAMVAConfig{}, true, fmt.Errorf("invalid AAMVA_AUDIT_ENABLED: %w", err)
		}
		cfg.AuditEnabled = parsed
	}
	if val := get("AAMVA_SUPPORTED_STATES"); val != "" {
		rawStates := strings.Split(val, ",")
		var states []string
		for _, s := range rawStates {
			if trimmed := strings.TrimSpace(strings.ToUpper(s)); trimmed != "" {
				states = append(states, trimmed)
			}
		}
		if len(states) > 0 {
			cfg.SupportedStates = states
		}
	}

	// Allow adapter config to provide network settings and optional key material.
	if cfg.BaseURL == "" && baseConfig.Endpoint != "" {
		cfg.BaseURL = baseConfig.Endpoint
	}
	if cfg.Timeout == 0 && baseConfig.Timeout > 0 {
		cfg.Timeout = baseConfig.Timeout
	}
	if cfg.RateLimitPerMinute == 0 && baseConfig.RateLimit > 0 {
		cfg.RateLimitPerMinute = baseConfig.RateLimit
	}
	if cfg.APIKey == "" && baseConfig.APIKey != "" {
		cfg.APIKey = baseConfig.APIKey
	}
	if cfg.ClientSecret == "" && baseConfig.APISecret != "" {
		cfg.ClientSecret = baseConfig.APISecret
	}

	// If no AAMVA-related env vars and no required fields, treat as not configured.
	if !configured && !enabledFlagSet && cfg.OrgID == "" && cfg.PermissionCode == "" && cfg.ClientID == "" && cfg.ClientSecret == "" {
		return AAMVAConfig{}, false, nil
	}

	if err := cfg.Validate(); err != nil {
		return AAMVAConfig{}, true, err
	}

	return cfg, true, nil
}

// ============================================================================
// Compile-time interface check
// ============================================================================

var _ DataSourceAdapter = (*AAMVADMVAdapter)(nil)
