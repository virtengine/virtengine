// Package govdata provides government data source integration for identity verification.
//
// GOVDATA-002: EU eIDAS (Electronic Identification and Trust Services) adapter implementation
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
// eIDAS Constants
// ============================================================================

// eIDAS Environments
const (
	EIDASEnvironmentProduction = "production"
	EIDASEnvironmentTest       = "test"
	EIDASEnvironmentSandbox    = "sandbox"
)

// eIDAS Node API Endpoints (EU Member State nodes)
const (
	EIDASProductionURL = "https://eidas.ec.europa.eu/EidasNode/v2"
	EIDASTestURL       = "https://test.eidas.ec.europa.eu/EidasNode/v2"
	EIDASSandboxURL    = "https://sandbox.eidas.ec.europa.eu/EidasNode/v2"
)

// eIDAS Levels of Assurance
const (
	EIDASLoALow         = "http://eidas.europa.eu/LoA/low"
	EIDASLoASubstantial = "http://eidas.europa.eu/LoA/substantial"
	EIDASLoAHigh        = "http://eidas.europa.eu/LoA/high"
)

// eIDAS Attribute URIs
const (
	EIDASAttrPersonIdentifier = "http://eidas.europa.eu/attributes/naturalperson/PersonIdentifier"
	EIDASAttrFamilyName       = "http://eidas.europa.eu/attributes/naturalperson/CurrentFamilyName"
	EIDASAttrFirstName        = "http://eidas.europa.eu/attributes/naturalperson/CurrentGivenName"
	EIDASAttrDateOfBirth      = "http://eidas.europa.eu/attributes/naturalperson/DateOfBirth"
	EIDASAttrPlaceOfBirth     = "http://eidas.europa.eu/attributes/naturalperson/PlaceOfBirth"
	EIDASAttrCurrentAddress   = "http://eidas.europa.eu/attributes/naturalperson/CurrentAddress"
	EIDASAttrGender           = "http://eidas.europa.eu/attributes/naturalperson/Gender"
	EIDASAttrBirthName        = "http://eidas.europa.eu/attributes/naturalperson/BirthName"
	EIDASAttrNationality      = "http://eidas.europa.eu/attributes/naturalperson/Nationality"
)

// EU Member States with eIDAS nodes
var EIDASMemberStates = []string{
	"AT", "BE", "BG", "HR", "CY", "CZ", "DK", "EE", "FI", "FR",
	"DE", "GR", "HU", "IE", "IT", "LV", "LT", "LU", "MT", "NL",
	"PL", "PT", "RO", "SK", "SI", "ES", "SE",
	// EEA countries
	"IS", "LI", "NO",
}

// ============================================================================
// eIDAS Errors
// ============================================================================

var (
	// ErrEIDASNotConfigured is returned when eIDAS is not configured
	ErrEIDASNotConfigured = errors.New("eIDAS API not configured")

	// ErrEIDASAuthentication is returned on auth failure
	ErrEIDASAuthentication = errors.New("eIDAS authentication failed")

	// ErrEIDASRateLimit is returned when rate limited
	ErrEIDASRateLimit = errors.New("eIDAS API rate limit exceeded")

	// ErrEIDASTimeout is returned on timeout
	ErrEIDASTimeout = errors.New("eIDAS API request timed out")

	// ErrEIDASCountryNotSupported is returned for unsupported member states
	ErrEIDASCountryNotSupported = errors.New("member state not supported by eIDAS")

	// ErrEIDASInvalidResponse is returned for malformed responses
	ErrEIDASInvalidResponse = errors.New("invalid eIDAS API response")

	// ErrEIDASRequestFailed is returned for general request failures
	ErrEIDASRequestFailed = errors.New("eIDAS API request failed")

	// ErrEIDASConsentRequired is returned when user consent is required
	ErrEIDASConsentRequired = errors.New("eIDAS user consent required")

	// ErrEIDASLoANotMet is returned when LoA requirements are not met
	ErrEIDASLoANotMet = errors.New("eIDAS level of assurance not met")
)

// ============================================================================
// eIDAS Configuration
// ============================================================================

// EIDASConfig contains eIDAS API configuration
type EIDASConfig struct {
	// Environment is "production", "test", or "sandbox"
	Environment string `json:"environment"`

	// ServiceProviderID is the SP entity ID
	ServiceProviderID string `json:"service_provider_id"`

	// ServiceProviderCountry is the SP country code
	ServiceProviderCountry string `json:"service_provider_country"`

	// MetadataURL is the SP metadata URL
	MetadataURL string `json:"metadata_url"`

	// SigningCertPath is the path to the signing certificate
	SigningCertPath string `json:"signing_cert_path"`

	// EncryptionCertPath is the path to the encryption certificate
	EncryptionCertPath string `json:"encryption_cert_path"`

	// PrivateKeyPath is the path to the private key
	PrivateKeyPath string `json:"private_key_path"`

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

	// RequiredLoA is the minimum required Level of Assurance
	RequiredLoA string `json:"required_loa"`

	// SupportedCountries lists the member states enabled for queries
	SupportedCountries []string `json:"supported_countries"`

	// RequestedAttributes lists the attributes to request
	RequestedAttributes []string `json:"requested_attributes"`

	// AuditEnabled enables detailed audit logging
	AuditEnabled bool `json:"audit_enabled"`

	// GDPRCompliant ensures GDPR compliance mode
	GDPRCompliant bool `json:"gdpr_compliant"`
}

// DefaultEIDASConfig returns default eIDAS configuration
func DefaultEIDASConfig() EIDASConfig {
	return EIDASConfig{
		Environment:        EIDASEnvironmentSandbox,
		Timeout:            60 * time.Second, // eIDAS can be slow due to cross-border
		MaxRetries:         3,
		RetryBackoff:       1 * time.Second,
		RateLimitPerMinute: 30,
		RequiredLoA:        EIDASLoASubstantial,
		SupportedCountries: EIDASMemberStates,
		RequestedAttributes: []string{
			EIDASAttrPersonIdentifier,
			EIDASAttrFamilyName,
			EIDASAttrFirstName,
			EIDASAttrDateOfBirth,
		},
		AuditEnabled:  true,
		GDPRCompliant: true,
	}
}

// Validate validates the eIDAS configuration
func (c *EIDASConfig) Validate() error {
	validEnvs := map[string]bool{
		EIDASEnvironmentProduction: true,
		EIDASEnvironmentTest:       true,
		EIDASEnvironmentSandbox:    true,
	}
	if !validEnvs[c.Environment] {
		return fmt.Errorf("invalid eIDAS environment: %s", c.Environment)
	}
	if c.ServiceProviderID == "" {
		return errors.New("eIDAS service_provider_id is required")
	}
	if c.ServiceProviderCountry == "" {
		return errors.New("eIDAS service_provider_country is required")
	}
	if c.APIKey == "" {
		return errors.New("eIDAS api_key is required")
	}
	return nil
}

// ============================================================================
// eIDAS Request/Response Types
// ============================================================================

// EIDASAuthRequest represents an eIDAS authentication request
type EIDASAuthRequest struct {
	RequestID           string   `json:"requestId"`
	DestinationCountry  string   `json:"destinationCountry"`
	LevelOfAssurance    string   `json:"levelOfAssurance"`
	RequestedAttributes []string `json:"requestedAttributes"`
	SPCountry           string   `json:"spCountry"`
	CitizenCountry      string   `json:"citizenCountry,omitempty"`
	NameIDFormat        string   `json:"nameIdFormat,omitempty"`
	ForceAuth           bool     `json:"forceAuth,omitempty"`
}

// EIDASAuthResponse represents an eIDAS authentication response
type EIDASAuthResponse struct {
	RequestID           string                    `json:"requestId"`
	Status              string                    `json:"status"`
	StatusCode          string                    `json:"statusCode"`
	SubStatusCode       string                    `json:"subStatusCode,omitempty"`
	StatusMessage       string                    `json:"statusMessage,omitempty"`
	LevelOfAssurance    string                    `json:"levelOfAssurance,omitempty"`
	Attributes          []EIDASAttribute          `json:"attributes,omitempty"`
	Issuer              string                    `json:"issuer,omitempty"`
	IssueInstant        string                    `json:"issueInstant,omitempty"`
	SubjectConfirmation *EIDASSubjectConfirmation `json:"subjectConfirmation,omitempty"`
}

// EIDASAttribute represents a verified eIDAS attribute
type EIDASAttribute struct {
	Name         string   `json:"name"`
	FriendlyName string   `json:"friendlyName,omitempty"`
	Values       []string `json:"values"`
	Status       string   `json:"status,omitempty"`
}

// EIDASSubjectConfirmation contains subject confirmation data
type EIDASSubjectConfirmation struct {
	Method       string `json:"method"`
	NotOnOrAfter string `json:"notOnOrAfter,omitempty"`
	NotBefore    string `json:"notBefore,omitempty"`
}

// eIDAS Status codes
const (
	EIDASStatusSuccess         = "urn:oasis:names:tc:SAML:2.0:status:Success"
	EIDASStatusRequester       = "urn:oasis:names:tc:SAML:2.0:status:Requester"
	EIDASStatusResponder       = "urn:oasis:names:tc:SAML:2.0:status:Responder"
	EIDASStatusVersionMismatch = "urn:oasis:names:tc:SAML:2.0:status:VersionMismatch"
	EIDASStatusAuthnFailed     = "urn:oasis:names:tc:SAML:2.0:status:AuthnFailed"
	//nolint:gosec // G101: This is a SAML URN constant, not a credential
	EIDASStatusNoPassive = "urn:oasis:names:tc:SAML:2.0:status:NoPassive"
)

// ============================================================================
// eIDAS Adapter Implementation
// ============================================================================

// eidasAdapter implements the eIDAS adapter for EU identity verification
type eidasAdapter struct {
	*baseAdapter
	eidasConfig  EIDASConfig
	accessToken  string    //nolint:unused // Reserved for OAuth token caching
	tokenExpiry  time.Time //nolint:unused // Reserved for OAuth token caching
	windowStart  time.Time
	requestCount int
	mu           sync.RWMutex
	httpClient   *http.Client
}

// NewEIDASAdapter creates a new eIDAS adapter
func NewEIDASAdapter(config AdapterConfig, eidasConfig EIDASConfig) (*eidasAdapter, error) {
	if err := eidasConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid eIDAS config: %w", err)
	}

	if len(config.SupportedDocuments) == 0 {
		config.SupportedDocuments = []DocumentType{
			DocumentTypeNationalID,
			DocumentTypePassport,
			DocumentTypeResidencePermit,
		}
	}

	timeout := eidasConfig.Timeout
	if timeout == 0 {
		timeout = config.Timeout
	}
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	return &eidasAdapter{
		baseAdapter: newBaseAdapter(config),
		eidasConfig: eidasConfig,
		windowStart: time.Now(),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// baseURL returns the eIDAS API base URL
func (a *eidasAdapter) baseURL() string {
	if a.eidasConfig.BaseURL != "" {
		return a.eidasConfig.BaseURL
	}
	switch a.eidasConfig.Environment {
	case EIDASEnvironmentProduction:
		return EIDASProductionURL
	case EIDASEnvironmentTest:
		return EIDASTestURL
	default:
		return EIDASSandboxURL
	}
}

// Verify performs eIDAS verification
func (a *eidasAdapter) Verify(ctx context.Context, req *VerificationRequest) (*VerificationResponse, error) {
	startTime := time.Now()

	// Check document type support
	if !a.SupportsDocument(req.DocumentType) {
		return nil, ErrDocumentTypeNotSupported
	}

	// Extract country from jurisdiction (e.g., "EU-DE" -> "DE", or just "DE")
	country := a.extractCountry(req.Jurisdiction)
	if !a.isCountrySupported(country) {
		return nil, ErrEIDASCountryNotSupported
	}

	// Check rate limit
	if err := a.checkRateLimit(); err != nil {
		return nil, err
	}

	// Build eIDAS request
	eidasReq := a.buildRequest(req, country)

	// Make API call
	eidasResp, err := a.callEIDASAPI(ctx, eidasReq)
	if err != nil {
		a.recordError(err)
		return nil, err
	}

	// Convert response
	response := a.convertResponse(req, eidasResp)

	latency := time.Since(startTime)
	a.recordSuccess(latency)

	return response, nil
}

// extractCountry extracts the country code from jurisdiction
func (a *eidasAdapter) extractCountry(jurisdiction string) string {
	// Handle formats: "EU-DE", "DE", "EU/DE"
	jurisdiction = strings.ToUpper(jurisdiction)

	// Remove EU prefix
	jurisdiction = strings.TrimPrefix(jurisdiction, "EU-")
	jurisdiction = strings.TrimPrefix(jurisdiction, "EU/")

	// Take first 2 characters as country code
	if len(jurisdiction) >= 2 {
		return jurisdiction[:2]
	}
	return jurisdiction
}

// isCountrySupported checks if a country is supported
func (a *eidasAdapter) isCountrySupported(country string) bool {
	country = strings.ToUpper(country)
	for _, c := range a.eidasConfig.SupportedCountries {
		if strings.EqualFold(c, country) {
			return true
		}
	}
	return false
}

// checkRateLimit checks and enforces rate limiting
func (a *eidasAdapter) checkRateLimit() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	if now.Sub(a.windowStart) >= time.Minute {
		a.windowStart = now
		a.requestCount = 0
	}

	if a.requestCount >= a.eidasConfig.RateLimitPerMinute {
		return ErrEIDASRateLimit
	}

	a.requestCount++
	return nil
}

// buildRequest builds an eIDAS authentication request
func (a *eidasAdapter) buildRequest(req *VerificationRequest, country string) *EIDASAuthRequest {
	eidasReq := &EIDASAuthRequest{
		RequestID:           a.generateRequestID(req.RequestID),
		DestinationCountry:  country,
		LevelOfAssurance:    a.eidasConfig.RequiredLoA,
		RequestedAttributes: a.eidasConfig.RequestedAttributes,
		SPCountry:           a.eidasConfig.ServiceProviderCountry,
		CitizenCountry:      country,
	}

	return eidasReq
}

// generateRequestID generates a unique request ID
func (a *eidasAdapter) generateRequestID(reqID string) string {
	h := hmac.New(sha256.New, []byte(a.eidasConfig.APIKey))
	h.Write([]byte(reqID + time.Now().String()))
	return "_" + base64.RawURLEncoding.EncodeToString(h.Sum(nil))[:31]
}

// callEIDASAPI makes the API call to eIDAS node
func (a *eidasAdapter) callEIDASAPI(ctx context.Context, eidasReq *EIDASAuthRequest) (*EIDASAuthResponse, error) {
	// Marshal request
	body, err := json.Marshal(eidasReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal eIDAS request: %w", err)
	}

	// Create HTTP request
	url := a.baseURL() + "/ServiceProvider/authenticate"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create eIDAS request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", a.eidasConfig.APIKey)
	httpReq.Header.Set("X-Request-ID", eidasReq.RequestID)
	httpReq.Header.Set("X-SP-Country", a.eidasConfig.ServiceProviderCountry)

	// Execute request with retries
	var resp *http.Response
	var lastErr error
	for i := 0; i <= a.eidasConfig.MaxRetries; i++ {
		resp, lastErr = a.httpClient.Do(httpReq)
		if lastErr == nil && resp.StatusCode < 500 {
			break
		}
		if i < a.eidasConfig.MaxRetries {
			time.Sleep(a.eidasConfig.RetryBackoff * time.Duration(i+1))
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("%w: %v", ErrEIDASRequestFailed, lastErr)
	}
	defer resp.Body.Close()

	// Handle error responses
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrEIDASAuthentication
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, ErrEIDASRateLimit
	}
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status %d: %s", ErrEIDASRequestFailed, resp.StatusCode, string(respBody))
	}

	// Parse response
	var eidasResp EIDASAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&eidasResp); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEIDASInvalidResponse, err)
	}

	return &eidasResp, nil
}

// convertResponse converts eIDAS response to verification response
func (a *eidasAdapter) convertResponse(req *VerificationRequest, eidasResp *EIDASAuthResponse) *VerificationResponse {
	response := &VerificationResponse{
		RequestID:      req.RequestID,
		DataSourceType: DataSourceNationalRegistry,
		Jurisdiction:   req.Jurisdiction,
		VerifiedAt:     time.Now(),
		ExpiresAt:      time.Now().Add(365 * 24 * time.Hour),
		FieldResults:   make(map[string]FieldVerificationResult),
	}

	// Map status to verification status
	switch eidasResp.StatusCode {
	case EIDASStatusSuccess:
		response.Status = VerificationStatusVerified
		response.DocumentValid = true
		// Determine confidence based on LoA
		response.Confidence = a.loaToConfidence(eidasResp.LevelOfAssurance)
	case EIDASStatusAuthnFailed:
		response.Status = VerificationStatusNotVerified
		response.Confidence = 0.0
		response.DocumentValid = false
		response.Warnings = append(response.Warnings, "Authentication failed")
	case EIDASStatusRequester:
		response.Status = VerificationStatusError
		response.Confidence = 0.0
		response.DocumentValid = false
		response.ErrorCode = "REQUESTER_ERROR"
		response.ErrorMessage = eidasResp.StatusMessage
	case EIDASStatusResponder:
		response.Status = VerificationStatusError
		response.Confidence = 0.0
		response.DocumentValid = false
		response.ErrorCode = "RESPONDER_ERROR"
		response.ErrorMessage = eidasResp.StatusMessage
	default:
		response.Status = VerificationStatusError
		response.Confidence = 0.0
		response.DocumentValid = false
		response.ErrorCode = eidasResp.StatusCode
		response.ErrorMessage = eidasResp.StatusMessage
	}

	// Map attributes to field results
	for _, attr := range eidasResp.Attributes {
		fieldName := a.attrToFieldName(attr.Name)
		if fieldName != "" {
			response.FieldResults[fieldName] = FieldVerificationResult{
				FieldName:  fieldName,
				Match:      FieldMatchExact,
				Confidence: 1.0,
			}
		}
	}

	return response
}

// loaToConfidence converts Level of Assurance to confidence score
func (a *eidasAdapter) loaToConfidence(loa string) float64 {
	switch loa {
	case EIDASLoAHigh:
		return 1.0
	case EIDASLoASubstantial:
		return 0.9
	case EIDASLoALow:
		return 0.7
	default:
		return 0.5
	}
}

// attrToFieldName converts eIDAS attribute URI to field name
func (a *eidasAdapter) attrToFieldName(attrURI string) string {
	mapping := map[string]string{
		EIDASAttrPersonIdentifier: "document_number",
		EIDASAttrFamilyName:       "last_name",
		EIDASAttrFirstName:        "first_name",
		EIDASAttrDateOfBirth:      "date_of_birth",
		EIDASAttrPlaceOfBirth:     "place_of_birth",
		EIDASAttrCurrentAddress:   "address",
		EIDASAttrGender:           "gender",
		EIDASAttrNationality:      "nationality",
	}
	return mapping[attrURI]
}

// loadEIDASConfigFromEnv loads eIDAS configuration from environment variables
//
//nolint:unparam // result 2 (error) reserved for future validation failures
func loadEIDASConfigFromEnv(_ AdapterConfig) (EIDASConfig, bool, error) {
	spID := os.Getenv("EIDAS_SERVICE_PROVIDER_ID")
	spCountry := os.Getenv("EIDAS_SERVICE_PROVIDER_COUNTRY")
	apiKey := os.Getenv("EIDAS_API_KEY")

	if spID == "" || spCountry == "" || apiKey == "" {
		return EIDASConfig{}, false, nil
	}

	eidasConfig := DefaultEIDASConfig()
	eidasConfig.ServiceProviderID = spID
	eidasConfig.ServiceProviderCountry = spCountry
	eidasConfig.APIKey = apiKey

	if env := os.Getenv("EIDAS_ENVIRONMENT"); env != "" {
		eidasConfig.Environment = env
	}

	if baseURL := os.Getenv("EIDAS_BASE_URL"); baseURL != "" {
		eidasConfig.BaseURL = baseURL
	}

	if loa := os.Getenv("EIDAS_REQUIRED_LOA"); loa != "" {
		eidasConfig.RequiredLoA = loa
	}

	if eidasConfig.AuditEnabled {
		log.Printf("[eIDAS] Loaded configuration for SP %s (%s) in %s environment",
			eidasConfig.ServiceProviderID, eidasConfig.ServiceProviderCountry, eidasConfig.Environment)
	}

	return eidasConfig, true, nil
}
