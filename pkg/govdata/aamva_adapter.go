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
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
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

	// RateLimitPerMinute is the max requests per minute
	RateLimitPerMinute int `json:"rate_limit_per_minute"`

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
	XMLName      xml.Name `xml:"dldvRequest"`
	OrgID        string   `xml:"orgId"`
	Permission   string   `xml:"permissionCode"`
	Transaction  string   `xml:"transactionType"`
	MessageID    string   `xml:"messageId"`
	Timestamp    string   `xml:"timestamp"`
	
	// Driver Information
	State        string   `xml:"state"`
	LicenseNumber string  `xml:"licenseNumber"`
	FirstName    string   `xml:"firstName,omitempty"`
	MiddleName   string   `xml:"middleName,omitempty"`
	LastName     string   `xml:"lastName,omitempty"`
	DateOfBirth  string   `xml:"dateOfBirth,omitempty"`
	
	// Optional Address Verification
	AddressLine1 string   `xml:"addressLine1,omitempty"`
	AddressLine2 string   `xml:"addressLine2,omitempty"`
	City         string   `xml:"city,omitempty"`
	ZipCode      string   `xml:"zipCode,omitempty"`
	
	// Optional Photo Request (DLDV Plus)
	RequestPhoto bool     `xml:"requestPhoto,omitempty"`
}

// AAMVADLDVResponse represents a DLDV verification response
type AAMVADLDVResponse struct {
	XMLName       xml.Name `xml:"dldvResponse"`
	MessageID     string   `xml:"messageId"`
	ResponseCode  string   `xml:"responseCode"`
	ResponseText  string   `xml:"responseText"`
	Timestamp     string   `xml:"timestamp"`
	
	// Match Results
	OverallMatch     string `xml:"overallMatch"`
	LicenseMatch     string `xml:"licenseMatch"`
	FirstNameMatch   string `xml:"firstNameMatch"`
	MiddleNameMatch  string `xml:"middleNameMatch"`
	LastNameMatch    string `xml:"lastNameMatch"`
	DOBMatch         string `xml:"dobMatch"`
	AddressMatch     string `xml:"addressMatch"`
	
	// License Status
	LicenseStatus    string `xml:"licenseStatus"`
	LicenseClass     string `xml:"licenseClass"`
	ExpirationDate   string `xml:"expirationDate"`
	IssuedDate       string `xml:"issuedDate"`
	
	// Restrictions/Endorsements
	Restrictions     string `xml:"restrictions,omitempty"`
	Endorsements     string `xml:"endorsements,omitempty"`
	
	// Photo (DLDV Plus only)
	PhotoBase64      string `xml:"photo,omitempty"`
	
	// Error Information
	ErrorCode        string `xml:"errorCode,omitempty"`
	ErrorMessage     string `xml:"errorMessage,omitempty"`
}

// AAMVAMatchResult represents match outcome values
type AAMVAMatchResult string

const (
	AAMVAMatchYes       AAMVAMatchResult = "Y"  // Exact match
	AAMVAMatchNo        AAMVAMatchResult = "N"  // No match
	AAMVAMatchPartial   AAMVAMatchResult = "P"  // Partial match
	AAMVAMatchNoData    AAMVAMatchResult = "U"  // Unable to verify (no data)
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
	requestCount   int
	windowStart    time.Time
}

// NewAAMVADMVAdapter creates a new AAMVA DMV adapter
func NewAAMVADMVAdapter(baseConfig AdapterConfig, aamvaConfig AAMVAConfig) (*AAMVADMVAdapter, error) {
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
		httpClient: &http.Client{
			Timeout: aamvaConfig.Timeout,
		},
		windowStart: time.Now(),
	}

	return adapter, nil
}

// baseURL returns the AAMVA API base URL
func (a *AAMVADMVAdapter) baseURL() string {
	if a.config.Environment == AAMVAEnvironmentProduction {
		return AAMVAProductionURL
	}
	return AAMVASandboxURL
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
	
	reqBody := fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s",
		a.config.ClientID, "[REDACTED]") // Never log actual secret

	req, err := http.NewRequestWithContext(ctx, "POST", authURL, strings.NewReader(
		fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s",
			a.config.ClientID, a.config.ClientSecret)))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAAMVAAuthentication, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: status %d: %s", ErrAAMVAAuthentication, resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to parse auth response: %w", err)
	}

	a.accessToken = tokenResp.AccessToken
	// Set expiry with buffer for safety
	a.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

	_ = reqBody // silence unused variable warning

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
		"CA": regexp.MustCompile(`^[A-Z]\d{7}$`),                    // California: 1 letter + 7 digits
		"TX": regexp.MustCompile(`^\d{8}$`),                          // Texas: 8 digits
		"FL": regexp.MustCompile(`^[A-Z]\d{12}$`),                   // Florida: 1 letter + 12 digits
		"NY": regexp.MustCompile(`^\d{9}$`),                          // New York: 9 digits
		"PA": regexp.MustCompile(`^\d{8}$`),                          // Pennsylvania: 8 digits
		"IL": regexp.MustCompile(`^[A-Z]\d{11,12}$`),                // Illinois: 1 letter + 11-12 digits
		"OH": regexp.MustCompile(`^[A-Z]{2}\d{6}$`),                  // Ohio: 2 letters + 6 digits
		"GA": regexp.MustCompile(`^\d{9}$`),                          // Georgia: 9 digits
		"NC": regexp.MustCompile(`^\d{1,12}$`),                       // North Carolina: 1-12 digits
		"MI": regexp.MustCompile(`^[A-Z]\d{10,12}$`),                 // Michigan: 1 letter + 10-12 digits
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

	// Check rate limit
	if err := a.checkRateLimit(); err != nil {
		return nil, err
	}

	// Authenticate
	if err := a.authenticate(ctx); err != nil {
		a.recordError(err)
		return nil, err
	}

	// Build AAMVA request
	transType := AAMVATransactionDLDV
	if a.config.EnableDLDVPlus {
		transType = AAMVATransactionDLDVP
	}

	dldvReq := &AAMVADLDVRequest{
		OrgID:         a.config.OrgID,
		Permission:    a.config.PermissionCode,
		Transaction:   transType,
		MessageID:     a.generateMessageID(req.RequestID),
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		State:         state,
		LicenseNumber: req.Fields.DocumentNumber,
		FirstName:     req.Fields.FirstName,
		MiddleName:    req.Fields.MiddleName,
		LastName:      req.Fields.LastName,
		RequestPhoto:  a.config.EnableDLDVPlus,
	}

	if !req.Fields.DateOfBirth.IsZero() {
		dldvReq.DateOfBirth = req.Fields.DateOfBirth.Format("2006-01-02")
	}

	// Address verification if provided
	if req.Fields.Address != nil && req.Fields.Address.Street != "" {
		dldvReq.AddressLine1 = req.Fields.Address.Street
		dldvReq.City = req.Fields.Address.City
		dldvReq.ZipCode = req.Fields.Address.PostalCode
	}

	// Serialize request
	xmlData, err := xml.MarshalIndent(dldvReq, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DLDV request: %w", err)
	}

	// Make API request
	apiURL := a.baseURL() + "/verify"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(xmlData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	a.mu.RLock()
	token := a.accessToken
	a.mu.RUnlock()

	httpReq.Header.Set("Content-Type", "application/xml")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("X-AAMVA-MessageID", dldvReq.MessageID)

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		a.recordError(err)
		if ctx.Err() == context.DeadlineExceeded {
			return nil, ErrAAMVATimeout
		}
		return nil, fmt.Errorf("%w: %v", ErrAAMVARequestFailed, err)
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, ErrAAMVARateLimit
	}
	if resp.StatusCode == http.StatusUnauthorized {
		// Clear token and retry once
		a.mu.Lock()
		a.accessToken = ""
		a.mu.Unlock()
		return nil, ErrAAMVAAuthentication
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status %d: %s", ErrAAMVARequestFailed, resp.StatusCode, string(body))
	}

	// Parse response
	var dldvResp AAMVADLDVResponse
	if err := xml.NewDecoder(resp.Body).Decode(&dldvResp); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAAMVAInvalidResponse, err)
	}

	// Convert AAMVA response to our format
	verificationResp := a.convertResponse(req, &dldvResp)
	
	latency := time.Since(startTime)
	a.recordSuccess(latency)

	return verificationResp, nil
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

// ============================================================================
// Compile-time interface check
// ============================================================================

var _ DataSourceAdapter = (*AAMVADMVAdapter)(nil)
