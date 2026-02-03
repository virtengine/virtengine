// Package govdata provides government data source integration for identity verification.
//
// GOVDATA-002: Australia Document Verification Service (DVS) adapter
package govdata

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// DVS Constants
// ============================================================================

// DVS Environments
const (
	DVSEnvironmentProduction = "production"
	DVSEnvironmentSandbox    = "sandbox"
)

// DVS API Endpoints
const (
	DVSProductionURL = "https://api.dvs.gov.au/v1"
	DVSSandboxURL    = "https://sandbox.dvs.gov.au/v1"
)

// DVS Document Types
const (
	DVSDocTypeDriverLicence    = "DRIVER_LICENCE"
	DVSDocTypePassport         = "PASSPORT"
	DVSDocTypeBirthCertificate = "BIRTH_CERTIFICATE"
	DVSDocTypeCitizenshipCert  = "CITIZENSHIP_CERTIFICATE"
	DVSDocTypeImmiCard         = "IMMI_CARD"
	DVSDocTypeMedicareCard     = "MEDICARE_CARD"
	DVSDocTypeVisaDocument     = "VISA"
)

// ============================================================================
// DVS Errors
// ============================================================================

var (
	// ErrDVSNotConfigured is returned when DVS is not configured
	ErrDVSNotConfigured = errors.New("DVS API not configured")

	// ErrDVSAuthentication is returned on auth failure
	ErrDVSAuthentication = errors.New("DVS authentication failed")

	// ErrDVSRateLimit is returned when rate limited
	ErrDVSRateLimit = errors.New("DVS API rate limit exceeded")

	// ErrDVSTimeout is returned on timeout
	ErrDVSTimeout = errors.New("DVS API request timed out")

	// ErrDVSStateNotSupported is returned for unsupported states/territories
	ErrDVSStateNotSupported = errors.New("state/territory not supported by DVS")

	// ErrDVSInvalidResponse is returned for malformed responses
	ErrDVSInvalidResponse = errors.New("invalid DVS API response")

	// ErrDVSRequestFailed is returned for general request failures
	ErrDVSRequestFailed = errors.New("DVS API request failed")
)

// ============================================================================
// DVS Configuration
// ============================================================================

// DVSConfig contains DVS API configuration
type DVSConfig struct {
	// Environment is "production" or "sandbox"
	Environment string `json:"environment"`

	// OrganisationID is the DVS organisation ID
	OrganisationID string `json:"organisation_id"`

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

	// BaseURL overrides the default DVS API URL (used for proxies/tests)
	BaseURL string `json:"base_url"`

	// SupportedStates lists the Australian states/territories enabled
	SupportedStates []string `json:"supported_states"`

	// AuditEnabled enables detailed audit logging
	AuditEnabled bool `json:"audit_enabled"`
}

// DefaultDVSConfig returns default DVS configuration
func DefaultDVSConfig() DVSConfig {
	return DVSConfig{
		Environment:        DVSEnvironmentSandbox,
		Timeout:            30 * time.Second,
		MaxRetries:         3,
		RetryBackoff:       500 * time.Millisecond,
		RateLimitPerMinute: 60,
		AuditEnabled:       true,
		// All Australian states and territories
		SupportedStates: []string{
			"NSW", "VIC", "QLD", "SA", "WA", "TAS", "NT", "ACT",
		},
	}
}

// Validate validates the DVS configuration
func (c *DVSConfig) Validate() error {
	if c.Environment != DVSEnvironmentProduction && c.Environment != DVSEnvironmentSandbox {
		return fmt.Errorf("invalid DVS environment: %s", c.Environment)
	}
	if c.OrganisationID == "" {
		return errors.New("DVS organisation_id is required")
	}
	if c.ClientID == "" {
		return errors.New("DVS client_id is required")
	}
	if c.ClientSecret == "" {
		return errors.New("DVS client_secret is required")
	}
	return nil
}

// ============================================================================
// DVS Request/Response Types
// ============================================================================

// DVSVerifyRequest represents a DVS verification request
type DVSVerifyRequest struct {
	DocumentType   string       `json:"documentType"`
	DocumentFields DVSDocFields `json:"documentFields"`
	ConsentGiven   bool         `json:"consentGiven"`
	RequestID      string       `json:"requestId"`
}

// DVSDocFields contains document-specific fields for verification
type DVSDocFields struct {
	// Common fields
	FamilyName  string `json:"familyName,omitempty"`
	GivenNames  string `json:"givenNames,omitempty"`
	DateOfBirth string `json:"dateOfBirth,omitempty"` // Format: YYYY-MM-DD

	// Driver's licence fields
	LicenceNumber string `json:"licenceNumber,omitempty"`
	StateOfIssue  string `json:"stateOfIssue,omitempty"`
	CardNumber    string `json:"cardNumber,omitempty"`

	// Passport fields
	PassportNumber string `json:"passportNumber,omitempty"`
	Nationality    string `json:"nationality,omitempty"`
	Gender         string `json:"gender,omitempty"`

	// Birth certificate fields
	RegistrationNumber string `json:"registrationNumber,omitempty"`
	RegistrationYear   string `json:"registrationYear,omitempty"`
	RegistrationState  string `json:"registrationState,omitempty"`
	CertificateNumber  string `json:"certificateNumber,omitempty"`

	// Medicare fields
	MedicareNumber    string `json:"medicareNumber,omitempty"`
	MedicareReference string `json:"medicareReference,omitempty"`
	MedicareExpiry    string `json:"medicareExpiry,omitempty"`
}

// DVSVerifyResponse represents a DVS verification response
type DVSVerifyResponse struct {
	RequestID      string          `json:"requestId"`
	VerificationID string          `json:"verificationId"`
	Status         string          `json:"status"`
	VerifyResult   DVSVerifyResult `json:"verifyResult"`
	Timestamp      string          `json:"timestamp"`
	ErrorCode      string          `json:"errorCode,omitempty"`
	ErrorMessage   string          `json:"errorMessage,omitempty"`
}

// DVSVerifyResult contains the verification result
type DVSVerifyResult struct {
	OverallResult  string            `json:"overallResult"`
	FieldResults   map[string]string `json:"fieldResults"`
	DocumentValid  bool              `json:"documentValid"`
	DocumentStatus string            `json:"documentStatus,omitempty"`
}

// DVS Result codes
const (
	DVSResultMatch       = "MATCH"
	DVSResultNoMatch     = "NO_MATCH"
	DVSResultPartial     = "PARTIAL_MATCH"
	DVSResultUnavailable = "UNAVAILABLE"
	DVSResultError       = "ERROR"
)

// ============================================================================
// DVS Adapter Implementation
// ============================================================================

// dvsDMVAdapter implements the DVS adapter for Australian documents
type dvsDMVAdapter struct {
	*baseAdapter
	dvsConfig    DVSConfig
	accessToken  string
	tokenExpiry  time.Time
	windowStart  time.Time
	requestCount int
	mu           sync.RWMutex
	httpClient   *http.Client
}

// NewDVSDMVAdapter creates a new DVS adapter
func NewDVSDMVAdapter(config AdapterConfig, dvsConfig DVSConfig) (*dvsDMVAdapter, error) {
	if err := dvsConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid DVS config: %w", err)
	}

	if len(config.SupportedDocuments) == 0 {
		config.SupportedDocuments = []DocumentType{
			DocumentTypeDriversLicense,
			DocumentTypePassport,
			DocumentTypeBirthCertificate,
			DocumentTypeNationalID,
		}
	}

	timeout := dvsConfig.Timeout
	if timeout == 0 {
		timeout = config.Timeout
	}
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &dvsDMVAdapter{
		baseAdapter: newBaseAdapter(config),
		dvsConfig:   dvsConfig,
		windowStart: time.Now(),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// baseURL returns the DVS API base URL
func (a *dvsDMVAdapter) baseURL() string {
	if a.dvsConfig.BaseURL != "" {
		return a.dvsConfig.BaseURL
	}
	if a.dvsConfig.Environment == DVSEnvironmentProduction {
		return DVSProductionURL
	}
	return DVSSandboxURL
}

// Verify performs DVS verification
func (a *dvsDMVAdapter) Verify(ctx context.Context, req *VerificationRequest) (*VerificationResponse, error) {
	startTime := time.Now()

	// Check document type support
	if !a.SupportsDocument(req.DocumentType) {
		return nil, ErrDocumentTypeNotSupported
	}

	// Extract state from jurisdiction (e.g., "AU-NSW" -> "NSW")
	state := a.extractState(req.Jurisdiction)
	if state != "" && !a.isStateSupported(state) {
		return nil, ErrDVSStateNotSupported
	}

	// Check rate limit
	if err := a.checkRateLimit(); err != nil {
		return nil, err
	}

	// Build DVS request
	dvsReq, err := a.buildRequest(req, state)
	if err != nil {
		return nil, err
	}

	// Make API call
	dvsResp, err := a.callDVSAPI(ctx, dvsReq)
	if err != nil {
		a.recordError(err)
		return nil, err
	}

	// Convert response
	response := a.convertResponse(req, dvsResp)

	latency := time.Since(startTime)
	a.recordSuccess(latency)

	return response, nil
}

// extractState extracts the state code from jurisdiction
func (a *dvsDMVAdapter) extractState(jurisdiction string) string {
	parts := strings.Split(jurisdiction, "-")
	if len(parts) >= 2 {
		return strings.ToUpper(parts[1])
	}
	return ""
}

// isStateSupported checks if a state is supported
func (a *dvsDMVAdapter) isStateSupported(state string) bool {
	state = strings.ToUpper(state)
	for _, s := range a.dvsConfig.SupportedStates {
		if strings.EqualFold(s, state) {
			return true
		}
	}
	return false
}

// checkRateLimit checks and enforces rate limiting
func (a *dvsDMVAdapter) checkRateLimit() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	if now.Sub(a.windowStart) >= time.Minute {
		a.windowStart = now
		a.requestCount = 0
	}

	if a.requestCount >= a.dvsConfig.RateLimitPerMinute {
		return ErrDVSRateLimit
	}

	a.requestCount++
	return nil
}

// buildRequest builds a DVS request from verification request
func (a *dvsDMVAdapter) buildRequest(req *VerificationRequest, state string) (*DVSVerifyRequest, error) {
	dvsReq := &DVSVerifyRequest{
		ConsentGiven: true,
		RequestID:    a.generateRequestID(req.RequestID),
	}

	// Map document type
	switch req.DocumentType {
	case DocumentTypeDriversLicense:
		dvsReq.DocumentType = DVSDocTypeDriverLicence
		dvsReq.DocumentFields = DVSDocFields{
			LicenceNumber: req.Fields.DocumentNumber,
			StateOfIssue:  state,
			FamilyName:    req.Fields.LastName,
			GivenNames:    req.Fields.FirstName,
		}
		if !req.Fields.DateOfBirth.IsZero() {
			dvsReq.DocumentFields.DateOfBirth = req.Fields.DateOfBirth.Format("2006-01-02")
		}

	case DocumentTypePassport:
		dvsReq.DocumentType = DVSDocTypePassport
		dvsReq.DocumentFields = DVSDocFields{
			PassportNumber: req.Fields.DocumentNumber,
			FamilyName:     req.Fields.LastName,
			GivenNames:     req.Fields.FirstName,
			Gender:         req.Fields.Gender,
		}
		if !req.Fields.DateOfBirth.IsZero() {
			dvsReq.DocumentFields.DateOfBirth = req.Fields.DateOfBirth.Format("2006-01-02")
		}

	case DocumentTypeBirthCertificate:
		dvsReq.DocumentType = DVSDocTypeBirthCertificate
		dvsReq.DocumentFields = DVSDocFields{
			RegistrationNumber: req.Fields.DocumentNumber,
			RegistrationState:  state,
			FamilyName:         req.Fields.LastName,
			GivenNames:         req.Fields.FirstName,
		}
		if !req.Fields.DateOfBirth.IsZero() {
			dvsReq.DocumentFields.DateOfBirth = req.Fields.DateOfBirth.Format("2006-01-02")
		}

	default:
		return nil, ErrDocumentTypeNotSupported
	}

	return dvsReq, nil
}

// generateRequestID generates a unique request ID for DVS
func (a *dvsDMVAdapter) generateRequestID(reqID string) string {
	h := hmac.New(sha256.New, []byte(a.dvsConfig.ClientSecret))
	h.Write([]byte(reqID + time.Now().String()))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))[:32]
}

// callDVSAPI makes the API call to DVS
func (a *dvsDMVAdapter) callDVSAPI(ctx context.Context, dvsReq *DVSVerifyRequest) (*DVSVerifyResponse, error) {
	// Get access token
	if err := a.ensureAccessToken(ctx); err != nil {
		return nil, err
	}

	// Marshal request
	body, err := json.Marshal(dvsReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DVS request: %w", err)
	}

	// Create HTTP request
	url := a.baseURL() + "/verify"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create DVS request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.accessToken)
	httpReq.Header.Set("X-Request-ID", dvsReq.RequestID)

	// Execute request with retries
	var resp *http.Response
	var lastErr error
	for i := 0; i <= a.dvsConfig.MaxRetries; i++ {
		resp, lastErr = a.httpClient.Do(httpReq)
		if lastErr == nil && resp.StatusCode < 500 {
			break
		}
		if i < a.dvsConfig.MaxRetries {
			time.Sleep(a.dvsConfig.RetryBackoff * time.Duration(i+1))
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("%w: %v", ErrDVSRequestFailed, lastErr)
	}
	defer resp.Body.Close()

	// Handle error responses
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrDVSAuthentication
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, ErrDVSRateLimit
	}
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status %d: %s", ErrDVSRequestFailed, resp.StatusCode, string(respBody))
	}

	// Parse response
	var dvsResp DVSVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&dvsResp); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDVSInvalidResponse, err)
	}

	return &dvsResp, nil
}

// ensureAccessToken ensures we have a valid access token
func (a *dvsDMVAdapter) ensureAccessToken(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.accessToken != "" && time.Now().Before(a.tokenExpiry) {
		return nil
	}

	// Request new token
	url := a.baseURL() + "/oauth/token"
	data := fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s",
		a.dvsConfig.ClientID, a.dvsConfig.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ErrDVSAuthentication
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	a.accessToken = tokenResp.AccessToken
	a.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

	return nil
}

// convertResponse converts DVS response to verification response
func (a *dvsDMVAdapter) convertResponse(req *VerificationRequest, dvsResp *DVSVerifyResponse) *VerificationResponse {
	response := &VerificationResponse{
		RequestID:      req.RequestID,
		DataSourceType: DataSourceDMV,
		Jurisdiction:   req.Jurisdiction,
		VerifiedAt:     time.Now(),
		ExpiresAt:      time.Now().Add(365 * 24 * time.Hour),
		FieldResults:   make(map[string]FieldVerificationResult),
	}

	// Map overall result
	switch dvsResp.VerifyResult.OverallResult {
	case DVSResultMatch:
		response.Status = VerificationStatusVerified
		response.Confidence = 1.0
		response.DocumentValid = true
	case DVSResultPartial:
		response.Status = VerificationStatusPartialMatch
		response.Confidence = 0.7
		response.DocumentValid = true
	case DVSResultNoMatch:
		response.Status = VerificationStatusNotVerified
		response.Confidence = 0.0
		response.DocumentValid = false
	default:
		response.Status = VerificationStatusError
		response.Confidence = 0.0
		response.DocumentValid = false
		response.ErrorCode = dvsResp.ErrorCode
		response.ErrorMessage = dvsResp.ErrorMessage
	}

	// Map field results
	for fieldName, result := range dvsResp.VerifyResult.FieldResults {
		response.FieldResults[fieldName] = a.convertFieldResult(fieldName, result)
	}

	// Check document status
	if !dvsResp.VerifyResult.DocumentValid {
		response.DocumentValid = false
		switch dvsResp.VerifyResult.DocumentStatus {
		case "EXPIRED":
			response.Status = VerificationStatusExpired
			response.Warnings = append(response.Warnings, "Document has expired")
		case "CANCELLED":
			response.Status = VerificationStatusRevoked
			response.Warnings = append(response.Warnings, "Document has been cancelled")
		}
	}

	return response
}

// convertFieldResult converts a DVS field result to our format
func (a *dvsDMVAdapter) convertFieldResult(fieldName, result string) FieldVerificationResult {
	fvr := FieldVerificationResult{
		FieldName: fieldName,
	}

	switch result {
	case DVSResultMatch:
		fvr.Match = FieldMatchExact
		fvr.Confidence = 1.0
	case DVSResultPartial:
		fvr.Match = FieldMatchFuzzy
		fvr.Confidence = 0.7
	case DVSResultNoMatch:
		fvr.Match = FieldMatchNoMatch
		fvr.Confidence = 0.0
	default:
		fvr.Match = FieldMatchUnavailable
		fvr.Confidence = 0.0
	}

	return fvr
}

// loadDVSConfigFromEnv loads DVS configuration from environment variables
//
//nolint:unparam // result 2 (error) reserved for future validation failures
func loadDVSConfigFromEnv(_ AdapterConfig) (DVSConfig, bool, error) {
	orgID := os.Getenv("DVS_ORGANISATION_ID")
	clientID := os.Getenv("DVS_CLIENT_ID")
	clientSecret := os.Getenv("DVS_CLIENT_SECRET")

	if orgID == "" || clientID == "" || clientSecret == "" {
		return DVSConfig{}, false, nil
	}

	dvsConfig := DefaultDVSConfig()
	dvsConfig.OrganisationID = orgID
	dvsConfig.ClientID = clientID
	dvsConfig.ClientSecret = clientSecret

	if env := os.Getenv("DVS_ENVIRONMENT"); env != "" {
		dvsConfig.Environment = env
	}

	if baseURL := os.Getenv("DVS_BASE_URL"); baseURL != "" {
		dvsConfig.BaseURL = baseURL
	}

	if dvsConfig.AuditEnabled {
		log.Printf("[DVS] Loaded configuration for organisation %s in %s environment",
			dvsConfig.OrganisationID, dvsConfig.Environment)
	}

	return dvsConfig, true, nil
}
