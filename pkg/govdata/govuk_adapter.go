// Package govdata provides government data source integration for identity verification.
//
// GOVDATA-002: UK GOV.UK Verify adapter implementation
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
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/security"
)

// ============================================================================
// GOV.UK Verify Constants
// ============================================================================

// GOV.UK Verify Environments
const (
	GovUKEnvironmentProduction  = "production"
	GovUKEnvironmentIntegration = "integration"
	GovUKEnvironmentSandbox     = "sandbox"
)

// GOV.UK Verify API Endpoints
const (
	GovUKProductionURL  = "https://api.verify.service.gov.uk/v2"
	GovUKIntegrationURL = "https://integration.verify.service.gov.uk/v2"
	GovUKSandboxURL     = "https://sandbox.verify.service.gov.uk/v2"
)

// GOV.UK Verify Levels of Assurance
const (
	GovUKLoA1 = "LOA_1" // Basic identity verification
	GovUKLoA2 = "LOA_2" // Standard identity verification
)

// ============================================================================
// GOV.UK Verify Errors
// ============================================================================

var (
	// ErrGovUKNotConfigured is returned when GOV.UK Verify is not configured
	ErrGovUKNotConfigured = errors.New("GOV.UK Verify API not configured")

	// ErrGovUKAuthentication is returned on auth failure
	ErrGovUKAuthentication = errors.New("GOV.UK Verify authentication failed")

	// ErrGovUKRateLimit is returned when rate limited
	ErrGovUKRateLimit = errors.New("GOV.UK Verify API rate limit exceeded")

	// ErrGovUKTimeout is returned on timeout
	ErrGovUKTimeout = errors.New("GOV.UK Verify API request timed out")

	// ErrGovUKInvalidResponse is returned for malformed responses
	ErrGovUKInvalidResponse = errors.New("invalid GOV.UK Verify API response")

	// ErrGovUKRequestFailed is returned for general request failures
	ErrGovUKRequestFailed = errors.New("GOV.UK Verify API request failed")

	// ErrGovUKSessionExpired is returned when session has expired
	ErrGovUKSessionExpired = errors.New("GOV.UK Verify session expired")
)

// ============================================================================
// GOV.UK Verify Configuration
// ============================================================================

// GovUKConfig contains GOV.UK Verify API configuration
type GovUKConfig struct {
	// Environment is "production", "integration", or "sandbox"
	Environment string `json:"environment"`

	// ServiceEntityID is the service provider entity ID
	ServiceEntityID string `json:"service_entity_id"`

	// SigningCertPath is the path to the signing certificate
	SigningCertPath string `json:"signing_cert_path"`

	// EncryptionCertPath is the path to the encryption certificate
	EncryptionCertPath string `json:"encryption_cert_path"`

	// PrivateKeyPath is the path to the private key
	PrivateKeyPath string `json:"private_key_path"`

	// HubEntityID is the Verify Hub entity ID
	HubEntityID string `json:"hub_entity_id"`

	// MatchingServiceEntityID is the matching service entity ID
	MatchingServiceEntityID string `json:"matching_service_entity_id"`

	// AssertionConsumerServiceURL is the callback URL
	AssertionConsumerServiceURL string `json:"assertion_consumer_service_url"`

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

	// BaseURL overrides the default API URL (used for proxies/tests)
	BaseURL string `json:"base_url"`

	// LevelOfAssurance is the required level of assurance
	LevelOfAssurance string `json:"level_of_assurance"`

	// AuditEnabled enables detailed audit logging
	AuditEnabled bool `json:"audit_enabled"`
}

// DefaultGovUKConfig returns default GOV.UK Verify configuration
func DefaultGovUKConfig() GovUKConfig {
	return GovUKConfig{
		Environment:        GovUKEnvironmentSandbox,
		Timeout:            30 * time.Second,
		MaxRetries:         3,
		RetryBackoff:       500 * time.Millisecond,
		RateLimitPerMinute: 60,
		LevelOfAssurance:   GovUKLoA2,
		AuditEnabled:       true,
	}
}

// Validate validates the GOV.UK Verify configuration
func (c *GovUKConfig) Validate() error {
	validEnvs := map[string]bool{
		GovUKEnvironmentProduction:  true,
		GovUKEnvironmentIntegration: true,
		GovUKEnvironmentSandbox:     true,
	}
	if !validEnvs[c.Environment] {
		return fmt.Errorf("invalid GOV.UK Verify environment: %s", c.Environment)
	}
	if c.ServiceEntityID == "" {
		return errors.New("GOV.UK Verify service_entity_id is required")
	}
	if c.APIKey == "" {
		return errors.New("GOV.UK Verify api_key is required")
	}
	return nil
}

// ============================================================================
// GOV.UK Verify Request/Response Types
// ============================================================================

// GovUKVerifyRequest represents a GOV.UK Verify request
type GovUKVerifyRequest struct {
	RequestID        string                `json:"requestId"`
	LevelOfAssurance string                `json:"levelOfAssurance"`
	Attributes       GovUKAttributeRequest `json:"attributes"`
}

// GovUKAttributeRequest specifies which attributes to verify
type GovUKAttributeRequest struct {
	FirstName    bool `json:"firstName"`
	MiddleNames  bool `json:"middleNames,omitempty"`
	Surname      bool `json:"surname"`
	DateOfBirth  bool `json:"dateOfBirth"`
	Address      bool `json:"address,omitempty"`
	Gender       bool `json:"gender,omitempty"`
	PlaceOfBirth bool `json:"placeOfBirth,omitempty"`
}

// GovUKVerifyResponse represents a GOV.UK Verify response
type GovUKVerifyResponse struct {
	RequestID        string           `json:"requestId"`
	Scenario         string           `json:"scenario"`
	PID              string           `json:"pid,omitempty"`
	Attributes       *GovUKAttributes `json:"attributes,omitempty"`
	LevelOfAssurance string           `json:"levelOfAssurance,omitempty"`
	ErrorCode        string           `json:"errorCode,omitempty"`
	ErrorMessage     string           `json:"errorMessage,omitempty"`
}

// GovUKAttributes contains verified identity attributes
type GovUKAttributes struct {
	FirstName   *GovUKVerifiedValue `json:"firstName,omitempty"`
	MiddleNames *GovUKVerifiedValue `json:"middleNames,omitempty"`
	Surname     *GovUKVerifiedValue `json:"surname,omitempty"`
	DateOfBirth *GovUKVerifiedValue `json:"dateOfBirth,omitempty"`
	Address     *GovUKAddress       `json:"address,omitempty"`
	Gender      *GovUKVerifiedValue `json:"gender,omitempty"`
}

// GovUKVerifiedValue represents a verified attribute value
type GovUKVerifiedValue struct {
	Value    string `json:"value"`
	Verified bool   `json:"verified"`
	From     string `json:"from,omitempty"`
	To       string `json:"to,omitempty"`
}

// GovUKAddress represents a verified address
type GovUKAddress struct {
	Lines                 []string `json:"lines"`
	PostCode              string   `json:"postCode"`
	InternationalPostCode string   `json:"internationalPostCode,omitempty"`
	UPRN                  string   `json:"uprn,omitempty"`
	Verified              bool     `json:"verified"`
}

// GOV.UK Verify Scenarios
const (
	GovUKScenarioSuccessMatch   = "SUCCESS_MATCH"
	GovUKScenarioAccountCreated = "ACCOUNT_CREATION"
	GovUKScenarioNoMatch        = "NO_MATCH"
	GovUKScenarioCancellation   = "CANCELLATION"
	GovUKScenarioError          = "ERROR"
	GovUKScenarioAuthFailed     = "AUTHENTICATION_FAILED"
)

// ============================================================================
// GOV.UK Verify Adapter Implementation
// ============================================================================

// govUKAdapter implements the GOV.UK Verify adapter
type govUKAdapter struct {
	*baseAdapter
	govUKConfig  GovUKConfig
	accessToken  string    //nolint:unused // Reserved for OAuth token caching
	tokenExpiry  time.Time //nolint:unused // Reserved for OAuth token caching
	windowStart  time.Time
	requestCount int
	mu           sync.RWMutex
	httpClient   *http.Client
}

// NewGovUKAdapter creates a new GOV.UK Verify adapter
func NewGovUKAdapter(config AdapterConfig, govUKConfig GovUKConfig) (*govUKAdapter, error) {
	if err := govUKConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid GOV.UK Verify config: %w", err)
	}

	if len(config.SupportedDocuments) == 0 {
		config.SupportedDocuments = []DocumentType{
			DocumentTypePassport,
			DocumentTypeDriversLicense,
			DocumentTypeNationalID,
		}
	}

	timeout := govUKConfig.Timeout
	if timeout == 0 {
		timeout = config.Timeout
	}
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &govUKAdapter{
		baseAdapter: newBaseAdapter(config),
		govUKConfig: govUKConfig,
		windowStart: time.Now(),
		httpClient:  security.NewSecureHTTPClient(security.WithTimeout(timeout)),
	}, nil
}

// baseURL returns the GOV.UK Verify API base URL
func (a *govUKAdapter) baseURL() string {
	if a.govUKConfig.BaseURL != "" {
		return a.govUKConfig.BaseURL
	}
	switch a.govUKConfig.Environment {
	case GovUKEnvironmentProduction:
		return GovUKProductionURL
	case GovUKEnvironmentIntegration:
		return GovUKIntegrationURL
	default:
		return GovUKSandboxURL
	}
}

// Verify performs GOV.UK Verify verification
func (a *govUKAdapter) Verify(ctx context.Context, req *VerificationRequest) (*VerificationResponse, error) {
	startTime := time.Now()

	// Check document type support
	if !a.SupportsDocument(req.DocumentType) {
		return nil, ErrDocumentTypeNotSupported
	}

	// Check rate limit
	if err := a.checkRateLimit(); err != nil {
		return nil, err
	}

	// Build GOV.UK request
	govUKReq := a.buildRequest(req)

	// Make API call
	govUKResp, err := a.callGovUKAPI(ctx, govUKReq)
	if err != nil {
		a.recordError(err)
		return nil, err
	}

	// Convert response
	response := a.convertResponse(req, govUKResp)

	latency := time.Since(startTime)
	a.recordSuccess(latency)

	return response, nil
}

// checkRateLimit checks and enforces rate limiting
func (a *govUKAdapter) checkRateLimit() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	if now.Sub(a.windowStart) >= time.Minute {
		a.windowStart = now
		a.requestCount = 0
	}

	if a.requestCount >= a.govUKConfig.RateLimitPerMinute {
		return ErrGovUKRateLimit
	}

	a.requestCount++
	return nil
}

// buildRequest builds a GOV.UK Verify request
func (a *govUKAdapter) buildRequest(req *VerificationRequest) *GovUKVerifyRequest {
	govUKReq := &GovUKVerifyRequest{
		RequestID:        a.generateRequestID(req.RequestID),
		LevelOfAssurance: a.govUKConfig.LevelOfAssurance,
		Attributes: GovUKAttributeRequest{
			FirstName:   req.Fields.FirstName != "",
			Surname:     req.Fields.LastName != "",
			DateOfBirth: !req.Fields.DateOfBirth.IsZero(),
		},
	}

	if req.Fields.MiddleName != "" {
		govUKReq.Attributes.MiddleNames = true
	}
	if req.Fields.Address != nil {
		govUKReq.Attributes.Address = true
	}
	if req.Fields.Gender != "" {
		govUKReq.Attributes.Gender = true
	}

	return govUKReq
}

// generateRequestID generates a unique request ID
func (a *govUKAdapter) generateRequestID(reqID string) string {
	h := hmac.New(sha256.New, []byte(a.govUKConfig.APIKey))
	h.Write([]byte(reqID + time.Now().String()))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))[:32]
}

// callGovUKAPI makes the API call to GOV.UK Verify
func (a *govUKAdapter) callGovUKAPI(ctx context.Context, govUKReq *GovUKVerifyRequest) (*GovUKVerifyResponse, error) {
	// Marshal request
	body, err := json.Marshal(govUKReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GOV.UK request: %w", err)
	}

	// Create HTTP request
	url := a.baseURL() + "/verify/match"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create GOV.UK request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", a.govUKConfig.APIKey)
	httpReq.Header.Set("X-Request-ID", govUKReq.RequestID)

	// Execute request with retries
	var resp *http.Response
	var lastErr error
	for i := 0; i <= a.govUKConfig.MaxRetries; i++ {
		resp, lastErr = a.httpClient.Do(httpReq)
		if lastErr == nil && resp.StatusCode < 500 {
			break
		}
		if i < a.govUKConfig.MaxRetries {
			time.Sleep(a.govUKConfig.RetryBackoff * time.Duration(i+1))
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("%w: %v", ErrGovUKRequestFailed, lastErr)
	}
	defer resp.Body.Close()

	// Handle error responses
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrGovUKAuthentication
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, ErrGovUKRateLimit
	}
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status %d: %s", ErrGovUKRequestFailed, resp.StatusCode, string(respBody))
	}

	// Parse response
	var govUKResp GovUKVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&govUKResp); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGovUKInvalidResponse, err)
	}

	return &govUKResp, nil
}

// convertResponse converts GOV.UK Verify response to verification response
func (a *govUKAdapter) convertResponse(req *VerificationRequest, govUKResp *GovUKVerifyResponse) *VerificationResponse {
	response := &VerificationResponse{
		RequestID:      req.RequestID,
		DataSourceType: DataSourceNationalRegistry,
		Jurisdiction:   req.Jurisdiction,
		VerifiedAt:     time.Now(),
		ExpiresAt:      time.Now().Add(365 * 24 * time.Hour),
		FieldResults:   make(map[string]FieldVerificationResult),
	}

	// Map scenario to status
	switch govUKResp.Scenario {
	case GovUKScenarioSuccessMatch:
		response.Status = VerificationStatusVerified
		response.Confidence = 1.0
		response.DocumentValid = true
	case GovUKScenarioAccountCreated:
		response.Status = VerificationStatusVerified
		response.Confidence = 0.95
		response.DocumentValid = true
	case GovUKScenarioNoMatch:
		response.Status = VerificationStatusNotFound
		response.Confidence = 0.0
		response.DocumentValid = false
	case GovUKScenarioCancellation:
		response.Status = VerificationStatusNotVerified
		response.Confidence = 0.0
		response.DocumentValid = false
		response.Warnings = append(response.Warnings, "User cancelled verification")
	case GovUKScenarioAuthFailed:
		response.Status = VerificationStatusNotVerified
		response.Confidence = 0.0
		response.DocumentValid = false
		response.Warnings = append(response.Warnings, "Authentication failed")
	default:
		response.Status = VerificationStatusError
		response.Confidence = 0.0
		response.DocumentValid = false
		response.ErrorCode = govUKResp.ErrorCode
		response.ErrorMessage = govUKResp.ErrorMessage
	}

	// Map verified attributes to field results
	if govUKResp.Attributes != nil {
		if govUKResp.Attributes.FirstName != nil {
			response.FieldResults["first_name"] = FieldVerificationResult{
				FieldName:  "first_name",
				Match:      a.boolToMatch(govUKResp.Attributes.FirstName.Verified),
				Confidence: a.boolToConfidence(govUKResp.Attributes.FirstName.Verified),
			}
		}
		if govUKResp.Attributes.Surname != nil {
			response.FieldResults["last_name"] = FieldVerificationResult{
				FieldName:  "last_name",
				Match:      a.boolToMatch(govUKResp.Attributes.Surname.Verified),
				Confidence: a.boolToConfidence(govUKResp.Attributes.Surname.Verified),
			}
		}
		if govUKResp.Attributes.DateOfBirth != nil {
			response.FieldResults["date_of_birth"] = FieldVerificationResult{
				FieldName:  "date_of_birth",
				Match:      a.boolToMatch(govUKResp.Attributes.DateOfBirth.Verified),
				Confidence: a.boolToConfidence(govUKResp.Attributes.DateOfBirth.Verified),
			}
		}
		if govUKResp.Attributes.Address != nil {
			response.FieldResults["address"] = FieldVerificationResult{
				FieldName:  "address",
				Match:      a.boolToMatch(govUKResp.Attributes.Address.Verified),
				Confidence: a.boolToConfidence(govUKResp.Attributes.Address.Verified),
			}
		}
	}

	return response
}

// boolToMatch converts a boolean to FieldMatchResult
func (a *govUKAdapter) boolToMatch(verified bool) FieldMatchResult {
	if verified {
		return FieldMatchExact
	}
	return FieldMatchNoMatch
}

// boolToConfidence converts a boolean to confidence score
func (a *govUKAdapter) boolToConfidence(verified bool) float64 {
	if verified {
		return 1.0
	}
	return 0.0
}

// loadGovUKConfigFromEnv loads GOV.UK Verify configuration from environment variables
//
//nolint:unparam // result 2 (error) reserved for future validation failures
func loadGovUKConfigFromEnv(_ AdapterConfig) (GovUKConfig, bool, error) {
	serviceEntityID := os.Getenv("GOVUK_SERVICE_ENTITY_ID")
	apiKey := os.Getenv("GOVUK_API_KEY")

	if serviceEntityID == "" || apiKey == "" {
		return GovUKConfig{}, false, nil
	}

	govUKConfig := DefaultGovUKConfig()
	govUKConfig.ServiceEntityID = serviceEntityID
	govUKConfig.APIKey = apiKey

	if env := os.Getenv("GOVUK_ENVIRONMENT"); env != "" {
		govUKConfig.Environment = env
	}

	if baseURL := os.Getenv("GOVUK_BASE_URL"); baseURL != "" {
		govUKConfig.BaseURL = baseURL
	}

	if loa := os.Getenv("GOVUK_LEVEL_OF_ASSURANCE"); loa != "" {
		govUKConfig.LevelOfAssurance = loa
	}

	if govUKConfig.AuditEnabled {
		log.Printf("[GOV.UK Verify] Loaded configuration for service %s in %s environment",
			govUKConfig.ServiceEntityID, govUKConfig.Environment)
	}

	return govUKConfig, true, nil
}
