// Package govdata provides government data source integration for identity verification.
//
// GOVDATA-002: Canada Pan-Canadian Trust Framework (PCTF) adapter implementation
package govdata

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
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
// PCTF Constants
// ============================================================================

// PCTF Environments
const (
	PCTFEnvironmentProduction = "production"
	PCTFEnvironmentTest       = "test"
	PCTFEnvironmentSandbox    = "sandbox"
)

// PCTF API Endpoints
const (
	PCTFProductionURL = "https://api.pctf.canada.ca/v1"
	PCTFTestURL       = "https://test.pctf.canada.ca/v1"
	PCTFSandboxURL    = "https://sandbox.pctf.canada.ca/v1"
)

// PCTF Assurance Levels (based on PCTF Verified Person component)
const (
	PCTFAssuranceLevelBasic     = "BASIC"
	PCTFAssuranceLevelStandard  = "STANDARD"
	PCTFAssuranceLevelEnhanced  = "ENHANCED"
	PCTFAssuranceLevelHigh      = "HIGH"
)

// PCTF Document Types
const (
	PCTFDocDriverLicence    = "DRIVER_LICENCE"
	PCTFDocPassport         = "PASSPORT"
	PCTFDocBirthCertificate = "BIRTH_CERTIFICATE"
	PCTFDocPRCard           = "PERMANENT_RESIDENT_CARD"
	PCTFDocSIN              = "SOCIAL_INSURANCE_NUMBER"
	PCTFDocHealthCard       = "HEALTH_CARD"
	PCTFDocCitizenshipCard  = "CITIZENSHIP_CARD"
	PCTFDocSecureStatusCard = "SECURE_CERTIFICATE_OF_INDIAN_STATUS"
)

// Canadian Provinces and Territories
var PCTFProvinceTerritories = []string{
	"AB", "BC", "MB", "NB", "NL", "NS", "NT", "NU", "ON", "PE", "QC", "SK", "YT",
}

// ============================================================================
// PCTF Errors
// ============================================================================

var (
	// ErrPCTFNotConfigured is returned when PCTF is not configured
	ErrPCTFNotConfigured = errors.New("PCTF API not configured")

	// ErrPCTFAuthentication is returned on auth failure
	ErrPCTFAuthentication = errors.New("PCTF authentication failed")

	// ErrPCTFRateLimit is returned when rate limited
	ErrPCTFRateLimit = errors.New("PCTF API rate limit exceeded")

	// ErrPCTFTimeout is returned on timeout
	ErrPCTFTimeout = errors.New("PCTF API request timed out")

	// ErrPCTFProvinceNotSupported is returned for unsupported provinces
	ErrPCTFProvinceNotSupported = errors.New("province/territory not supported by PCTF")

	// ErrPCTFInvalidResponse is returned for malformed responses
	ErrPCTFInvalidResponse = errors.New("invalid PCTF API response")

	// ErrPCTFRequestFailed is returned for general request failures
	ErrPCTFRequestFailed = errors.New("PCTF API request failed")

	// ErrPCTFConsentRequired is returned when user consent is required
	ErrPCTFConsentRequired = errors.New("PCTF user consent required")

	// ErrPCTFAssuranceNotMet is returned when assurance requirements are not met
	ErrPCTFAssuranceNotMet = errors.New("PCTF assurance level not met")
)

// ============================================================================
// PCTF Configuration
// ============================================================================

// PCTFConfig contains PCTF API configuration
type PCTFConfig struct {
	// Environment is "production", "test", or "sandbox"
	Environment string `json:"environment"`

	// OrganizationID is the relying party organization ID
	OrganizationID string `json:"organization_id"`

	// OrganizationName is the relying party organization name
	OrganizationName string `json:"organization_name"`

	// ServiceProviderID is the SP entity ID
	ServiceProviderID string `json:"service_provider_id"`

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

	// BaseURL overrides the default API URL (used for proxies/tests)
	BaseURL string `json:"base_url"`

	// RequiredAssuranceLevel is the minimum required assurance level
	RequiredAssuranceLevel string `json:"required_assurance_level"`

	// SupportedProvinces lists the provinces enabled for queries
	SupportedProvinces []string `json:"supported_provinces"`

	// PreferredLanguage is the preferred language (en/fr)
	PreferredLanguage string `json:"preferred_language"`

	// AuditEnabled enables detailed audit logging
	AuditEnabled bool `json:"audit_enabled"`

	// PIPEDACompliant ensures PIPEDA compliance mode
	PIPEDACompliant bool `json:"pipeda_compliant"`
}

// DefaultPCTFConfig returns default PCTF configuration
func DefaultPCTFConfig() PCTFConfig {
	return PCTFConfig{
		Environment:            PCTFEnvironmentSandbox,
		Timeout:                30 * time.Second,
		MaxRetries:             3,
		RetryBackoff:           500 * time.Millisecond,
		RateLimitPerMinute:     60,
		RequiredAssuranceLevel: PCTFAssuranceLevelStandard,
		SupportedProvinces:     PCTFProvinceTerritories,
		PreferredLanguage:      "en",
		AuditEnabled:           true,
		PIPEDACompliant:        true,
	}
}

// Validate validates the PCTF configuration
func (c *PCTFConfig) Validate() error {
	validEnvs := map[string]bool{
		PCTFEnvironmentProduction: true,
		PCTFEnvironmentTest:       true,
		PCTFEnvironmentSandbox:    true,
	}
	if !validEnvs[c.Environment] {
		return fmt.Errorf("invalid PCTF environment: %s", c.Environment)
	}
	if c.OrganizationID == "" {
		return errors.New("PCTF organization_id is required")
	}
	if c.ClientID == "" {
		return errors.New("PCTF client_id is required")
	}
	if c.ClientSecret == "" {
		return errors.New("PCTF client_secret is required")
	}
	return nil
}

// ============================================================================
// PCTF Request/Response Types
// ============================================================================

// PCTFVerifyRequest represents a PCTF verification request
type PCTFVerifyRequest struct {
	RequestID          string            `json:"requestId"`
	AssuranceLevel     string            `json:"assuranceLevel"`
	DocumentType       string            `json:"documentType"`
	Province           string            `json:"province,omitempty"`
	PersonData         PCTFPersonData    `json:"personData"`
	DocumentData       PCTFDocumentData  `json:"documentData"`
	ConsentReference   string            `json:"consentReference"`
	PreferredLanguage  string            `json:"preferredLanguage,omitempty"`
}

// PCTFPersonData contains person identity data
type PCTFPersonData struct {
	GivenName       string `json:"givenName,omitempty"`
	MiddleNames     string `json:"middleNames,omitempty"`
	Surname         string `json:"surname,omitempty"`
	DateOfBirth     string `json:"dateOfBirth,omitempty"` // Format: YYYY-MM-DD
	Gender          string `json:"gender,omitempty"`
	PlaceOfBirth    string `json:"placeOfBirth,omitempty"`
}

// PCTFDocumentData contains document-specific data
type PCTFDocumentData struct {
	DocumentNumber  string `json:"documentNumber,omitempty"`
	ExpiryDate      string `json:"expiryDate,omitempty"`
	IssueDate       string `json:"issueDate,omitempty"`
	IssuingAuthority string `json:"issuingAuthority,omitempty"`
	DocumentClass   string `json:"documentClass,omitempty"`
}

// PCTFVerifyResponse represents a PCTF verification response
type PCTFVerifyResponse struct {
	RequestID         string              `json:"requestId"`
	TransactionID     string              `json:"transactionId"`
	Status            string              `json:"status"`
	AssuranceLevel    string              `json:"assuranceLevel,omitempty"`
	VerificationResult PCTFVerificationResult `json:"verificationResult"`
	Timestamp         string              `json:"timestamp"`
	ErrorCode         string              `json:"errorCode,omitempty"`
	ErrorMessage      string              `json:"errorMessage,omitempty"`
}

// PCTFVerificationResult contains the verification result
type PCTFVerificationResult struct {
	OverallMatch      string                   `json:"overallMatch"`
	IdentityVerified  bool                     `json:"identityVerified"`
	DocumentVerified  bool                     `json:"documentVerified"`
	FieldResults      map[string]PCTFFieldResult `json:"fieldResults,omitempty"`
	Confidence        float64                  `json:"confidence"`
	AssuranceLevelMet bool                     `json:"assuranceLevelMet"`
}

// PCTFFieldResult contains a field-level verification result
type PCTFFieldResult struct {
	FieldName   string  `json:"fieldName"`
	Match       string  `json:"match"`
	Confidence  float64 `json:"confidence"`
	Note        string  `json:"note,omitempty"`
}

// PCTF Status codes
const (
	PCTFStatusSuccess      = "SUCCESS"
	PCTFStatusPartialMatch = "PARTIAL_MATCH"
	PCTFStatusNoMatch      = "NO_MATCH"
	PCTFStatusNotFound     = "NOT_FOUND"
	PCTFStatusError        = "ERROR"
	PCTFStatusExpired      = "DOCUMENT_EXPIRED"
	PCTFStatusRevoked      = "DOCUMENT_REVOKED"
)

// PCTF Match results
const (
	PCTFMatchYes     = "Y"
	PCTFMatchNo      = "N"
	PCTFMatchPartial = "P"
	PCTFMatchNA      = "NA"
)

// ============================================================================
// PCTF Adapter Implementation
// ============================================================================

// pctfAdapter implements the PCTF adapter for Canadian identity verification
type pctfAdapter struct {
	*baseAdapter
	pctfConfig   PCTFConfig
	accessToken  string
	tokenExpiry  time.Time
	windowStart  time.Time
	requestCount int
	mu           sync.RWMutex
	httpClient   *http.Client
}

// NewPCTFAdapter creates a new PCTF adapter
func NewPCTFAdapter(config AdapterConfig, pctfConfig PCTFConfig) (*pctfAdapter, error) {
	if err := pctfConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid PCTF config: %w", err)
	}

	if len(config.SupportedDocuments) == 0 {
		config.SupportedDocuments = []DocumentType{
			DocumentTypeDriversLicense,
			DocumentTypePassport,
			DocumentTypeBirthCertificate,
			DocumentTypeNationalID,
			DocumentTypeResidencePermit,
		}
	}

	timeout := pctfConfig.Timeout
	if timeout == 0 {
		timeout = config.Timeout
	}
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &pctfAdapter{
		baseAdapter: newBaseAdapter(config),
		pctfConfig:  pctfConfig,
		windowStart: time.Now(),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// baseURL returns the PCTF API base URL
func (a *pctfAdapter) baseURL() string {
	if a.pctfConfig.BaseURL != "" {
		return a.pctfConfig.BaseURL
	}
	switch a.pctfConfig.Environment {
	case PCTFEnvironmentProduction:
		return PCTFProductionURL
	case PCTFEnvironmentTest:
		return PCTFTestURL
	default:
		return PCTFSandboxURL
	}
}

// Verify performs PCTF verification
func (a *pctfAdapter) Verify(ctx context.Context, req *VerificationRequest) (*VerificationResponse, error) {
	startTime := time.Now()

	// Check document type support
	if !a.SupportsDocument(req.DocumentType) {
		return nil, ErrDocumentTypeNotSupported
	}

	// Extract province from jurisdiction (e.g., "CA-ON" -> "ON")
	province := a.extractProvince(req.Jurisdiction)
	if province != "" && !a.isProvinceSupported(province) {
		return nil, ErrPCTFProvinceNotSupported
	}

	// Check rate limit
	if err := a.checkRateLimit(); err != nil {
		return nil, err
	}

	// Ensure access token
	if err := a.ensureAccessToken(ctx); err != nil {
		return nil, err
	}

	// Build PCTF request
	pctfReq, err := a.buildRequest(req, province)
	if err != nil {
		return nil, err
	}

	// Make API call
	pctfResp, err := a.callPCTFAPI(ctx, pctfReq)
	if err != nil {
		a.recordError(err)
		return nil, err
	}

	// Convert response
	response := a.convertResponse(req, pctfResp)

	latency := time.Since(startTime)
	a.recordSuccess(latency)

	return response, nil
}

// extractProvince extracts the province code from jurisdiction
func (a *pctfAdapter) extractProvince(jurisdiction string) string {
	// Handle formats: "CA-ON", "ON", "CA/ON"
	jurisdiction = strings.ToUpper(jurisdiction)
	
	// Remove CA prefix
	jurisdiction = strings.TrimPrefix(jurisdiction, "CA-")
	jurisdiction = strings.TrimPrefix(jurisdiction, "CA/")
	
	// Take first 2 characters as province code
	if len(jurisdiction) >= 2 {
		return jurisdiction[:2]
	}
	return jurisdiction
}

// isProvinceSupported checks if a province is supported
func (a *pctfAdapter) isProvinceSupported(province string) bool {
	province = strings.ToUpper(province)
	for _, p := range a.pctfConfig.SupportedProvinces {
		if strings.EqualFold(p, province) {
			return true
		}
	}
	return false
}

// checkRateLimit checks and enforces rate limiting
func (a *pctfAdapter) checkRateLimit() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	if now.Sub(a.windowStart) >= time.Minute {
		a.windowStart = now
		a.requestCount = 0
	}

	if a.requestCount >= a.pctfConfig.RateLimitPerMinute {
		return ErrPCTFRateLimit
	}

	a.requestCount++
	return nil
}

// ensureAccessToken ensures we have a valid access token
func (a *pctfAdapter) ensureAccessToken(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.accessToken != "" && time.Now().Before(a.tokenExpiry) {
		return nil
	}

	// Request new token
	url := a.baseURL() + "/oauth/token"
	data := fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s",
		a.pctfConfig.ClientID, a.pctfConfig.ClientSecret)

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
		return ErrPCTFAuthentication
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

// buildRequest builds a PCTF verification request
func (a *pctfAdapter) buildRequest(req *VerificationRequest, province string) (*PCTFVerifyRequest, error) {
	pctfReq := &PCTFVerifyRequest{
		RequestID:         a.generateRequestID(req.RequestID),
		AssuranceLevel:    a.pctfConfig.RequiredAssuranceLevel,
		Province:          province,
		ConsentReference:  req.ConsentID,
		PreferredLanguage: a.pctfConfig.PreferredLanguage,
	}

	// Map document type
	switch req.DocumentType {
	case DocumentTypeDriversLicense:
		pctfReq.DocumentType = PCTFDocDriverLicence
	case DocumentTypePassport:
		pctfReq.DocumentType = PCTFDocPassport
	case DocumentTypeBirthCertificate:
		pctfReq.DocumentType = PCTFDocBirthCertificate
	case DocumentTypeResidencePermit:
		pctfReq.DocumentType = PCTFDocPRCard
	case DocumentTypeNationalID:
		pctfReq.DocumentType = PCTFDocCitizenshipCard
	default:
		return nil, ErrDocumentTypeNotSupported
	}

	// Person data
	pctfReq.PersonData = PCTFPersonData{
		GivenName:   req.Fields.FirstName,
		MiddleNames: req.Fields.MiddleName,
		Surname:     req.Fields.LastName,
		Gender:      req.Fields.Gender,
	}
	if !req.Fields.DateOfBirth.IsZero() {
		pctfReq.PersonData.DateOfBirth = req.Fields.DateOfBirth.Format("2006-01-02")
	}

	// Document data
	pctfReq.DocumentData = PCTFDocumentData{
		DocumentNumber:   req.Fields.DocumentNumber,
		IssuingAuthority: req.Fields.IssuingAuthority,
		DocumentClass:    req.Fields.DocumentClass,
	}
	if !req.Fields.ExpirationDate.IsZero() {
		pctfReq.DocumentData.ExpiryDate = req.Fields.ExpirationDate.Format("2006-01-02")
	}
	if !req.Fields.IssueDate.IsZero() {
		pctfReq.DocumentData.IssueDate = req.Fields.IssueDate.Format("2006-01-02")
	}

	return pctfReq, nil
}

// generateRequestID generates a unique request ID
func (a *pctfAdapter) generateRequestID(reqID string) string {
	h := hmac.New(sha256.New, []byte(a.pctfConfig.ClientSecret))
	h.Write([]byte(reqID + time.Now().String()))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))[:32]
}

// callPCTFAPI makes the API call to PCTF
func (a *pctfAdapter) callPCTFAPI(ctx context.Context, pctfReq *PCTFVerifyRequest) (*PCTFVerifyResponse, error) {
	// Marshal request
	body, err := json.Marshal(pctfReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PCTF request: %w", err)
	}

	// Create HTTP request
	url := a.baseURL() + "/verify/identity"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create PCTF request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.accessToken)
	httpReq.Header.Set("X-Request-ID", pctfReq.RequestID)
	httpReq.Header.Set("Accept-Language", a.pctfConfig.PreferredLanguage)

	// Execute request with retries
	var resp *http.Response
	var lastErr error
	for i := 0; i <= a.pctfConfig.MaxRetries; i++ {
		resp, lastErr = a.httpClient.Do(httpReq)
		if lastErr == nil && resp.StatusCode < 500 {
			break
		}
		if i < a.pctfConfig.MaxRetries {
			time.Sleep(a.pctfConfig.RetryBackoff * time.Duration(i+1))
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("%w: %v", ErrPCTFRequestFailed, lastErr)
	}
	defer resp.Body.Close()

	// Handle error responses
	if resp.StatusCode == http.StatusUnauthorized {
		// Clear token and retry once
		a.mu.Lock()
		a.accessToken = ""
		a.mu.Unlock()
		return nil, ErrPCTFAuthentication
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, ErrPCTFRateLimit
	}
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status %d: %s", ErrPCTFRequestFailed, resp.StatusCode, string(respBody))
	}

	// Parse response
	var pctfResp PCTFVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&pctfResp); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPCTFInvalidResponse, err)
	}

	return &pctfResp, nil
}

// convertResponse converts PCTF response to verification response
func (a *pctfAdapter) convertResponse(req *VerificationRequest, pctfResp *PCTFVerifyResponse) *VerificationResponse {
	response := &VerificationResponse{
		RequestID:      req.RequestID,
		DataSourceType: DataSourceDMV, // PCTF uses DMV-like data sources
		Jurisdiction:   req.Jurisdiction,
		VerifiedAt:     time.Now(),
		ExpiresAt:      time.Now().Add(365 * 24 * time.Hour),
		FieldResults:   make(map[string]FieldVerificationResult),
	}

	// Map status to verification status
	switch pctfResp.Status {
	case PCTFStatusSuccess:
		response.Status = VerificationStatusVerified
		response.Confidence = pctfResp.VerificationResult.Confidence
		response.DocumentValid = pctfResp.VerificationResult.DocumentVerified
	case PCTFStatusPartialMatch:
		response.Status = VerificationStatusPartialMatch
		response.Confidence = pctfResp.VerificationResult.Confidence
		response.DocumentValid = pctfResp.VerificationResult.DocumentVerified
	case PCTFStatusNoMatch:
		response.Status = VerificationStatusNotVerified
		response.Confidence = 0.0
		response.DocumentValid = false
	case PCTFStatusNotFound:
		response.Status = VerificationStatusNotFound
		response.Confidence = 0.0
		response.DocumentValid = false
	case PCTFStatusExpired:
		response.Status = VerificationStatusExpired
		response.Confidence = pctfResp.VerificationResult.Confidence
		response.DocumentValid = false
		response.Warnings = append(response.Warnings, "Document has expired")
	case PCTFStatusRevoked:
		response.Status = VerificationStatusRevoked
		response.Confidence = 0.0
		response.DocumentValid = false
		response.Warnings = append(response.Warnings, "Document has been revoked")
	default:
		response.Status = VerificationStatusError
		response.Confidence = 0.0
		response.DocumentValid = false
		response.ErrorCode = pctfResp.ErrorCode
		response.ErrorMessage = pctfResp.ErrorMessage
	}

	// Check assurance level
	if !pctfResp.VerificationResult.AssuranceLevelMet {
		response.Warnings = append(response.Warnings, "Required assurance level not met")
	}

	// Map field results
	for fieldName, fieldResult := range pctfResp.VerificationResult.FieldResults {
		response.FieldResults[fieldName] = a.convertFieldResult(fieldResult)
	}

	return response
}

// convertFieldResult converts a PCTF field result to our format
func (a *pctfAdapter) convertFieldResult(pctfResult PCTFFieldResult) FieldVerificationResult {
	result := FieldVerificationResult{
		FieldName:  pctfResult.FieldName,
		Confidence: pctfResult.Confidence,
		Note:       pctfResult.Note,
	}

	switch pctfResult.Match {
	case PCTFMatchYes:
		result.Match = FieldMatchExact
	case PCTFMatchPartial:
		result.Match = FieldMatchFuzzy
	case PCTFMatchNo:
		result.Match = FieldMatchNoMatch
	default:
		result.Match = FieldMatchUnavailable
	}

	return result
}

// loadPCTFConfigFromEnv loads PCTF configuration from environment variables
func loadPCTFConfigFromEnv(_ AdapterConfig) (PCTFConfig, bool, error) {
	orgID := os.Getenv("PCTF_ORGANIZATION_ID")
	clientID := os.Getenv("PCTF_CLIENT_ID")
	clientSecret := os.Getenv("PCTF_CLIENT_SECRET")

	if orgID == "" || clientID == "" || clientSecret == "" {
		return PCTFConfig{}, false, nil
	}

	pctfConfig := DefaultPCTFConfig()
	pctfConfig.OrganizationID = orgID
	pctfConfig.ClientID = clientID
	pctfConfig.ClientSecret = clientSecret

	if env := os.Getenv("PCTF_ENVIRONMENT"); env != "" {
		pctfConfig.Environment = env
	}

	if baseURL := os.Getenv("PCTF_BASE_URL"); baseURL != "" {
		pctfConfig.BaseURL = baseURL
	}

	if level := os.Getenv("PCTF_ASSURANCE_LEVEL"); level != "" {
		pctfConfig.RequiredAssuranceLevel = level
	}

	if lang := os.Getenv("PCTF_LANGUAGE"); lang != "" {
		pctfConfig.PreferredLanguage = lang
	}

	if pctfConfig.AuditEnabled {
		log.Printf("[PCTF] Loaded configuration for organization %s in %s environment",
			pctfConfig.OrganizationID, pctfConfig.Environment)
	}

	return pctfConfig, true, nil
}
