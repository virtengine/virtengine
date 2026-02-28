// Package offramp provides fiat off-ramp integration for token-to-fiat payouts.
//
// VE-5E: Fiat off-ramp integration for PayPal/ACH payouts
package offramp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/virtengine/virtengine/pkg/security"
)

// statusFailed is the VEID verification status for failed/rejected verifications
const statusFailed = "failed"

// ============================================================================
// KYC Gate Implementation
// ============================================================================

// VEIDKYCGate implements KYCGate using the VEID verification system.
type VEIDKYCGate struct {
	config      KYCConfig
	veidChecker VEIDChecker
}

// VEIDChecker is an interface for checking VEID verification status.
// This abstracts the actual VEID module interaction.
type VEIDChecker interface {
	// GetVerificationStatus retrieves the verification status for an account
	GetVerificationStatus(ctx context.Context, accountAddress string) (*VEIDVerificationStatus, error)

	// GetVerificationByID retrieves verification by VEID ID
	GetVerificationByID(ctx context.Context, veidID string) (*VEIDVerificationStatus, error)
}

// VEIDVerificationStatus represents the verification status from VEID module.
type VEIDVerificationStatus struct {
	// VEIDID is the verified identity ID
	VEIDID string `json:"veid_id"`

	// AccountAddress is the blockchain account
	AccountAddress string `json:"account_address"`

	// VerificationLevel is the current verification level
	VerificationLevel KYCVerificationLevel `json:"verification_level"`

	// Status is the verification status
	Status string `json:"status"`

	// VerifiedAt is when verification completed
	VerifiedAt *time.Time `json:"verified_at,omitempty"`

	// ExpiresAt is when verification expires
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// VerificationTypes lists completed verification types
	VerificationTypes []string `json:"verification_types,omitempty"`

	// Score is the aggregate verification score
	Score int `json:"score"`
}

// NewVEIDKYCGate creates a new VEID-based KYC gate.
func NewVEIDKYCGate(config KYCConfig, checker VEIDChecker) *VEIDKYCGate {
	return &VEIDKYCGate{
		config:      config,
		veidChecker: checker,
	}
}

// CheckKYCStatus checks the KYC status for an account.
func (g *VEIDKYCGate) CheckKYCStatus(ctx context.Context, accountAddress string, veidID string) (KYCCheckResult, error) {
	if !g.config.Enabled {
		// KYC disabled, return verified
		return KYCCheckResult{
			Status: KYCStatusVerified,
			Level:  KYCLevelFull,
		}, nil
	}

	var status *VEIDVerificationStatus
	var err error

	if veidID != "" {
		status, err = g.veidChecker.GetVerificationByID(ctx, veidID)
	} else {
		status, err = g.veidChecker.GetVerificationStatus(ctx, accountAddress)
	}

	if err != nil {
		return KYCCheckResult{
			Status:  KYCStatusPending,
			Message: fmt.Sprintf("Failed to check verification status: %v", err),
		}, nil
	}

	if status == nil {
		return KYCCheckResult{
			Status:  KYCStatusPending,
			Message: "No verification found for account",
		}, nil
	}

	result := KYCCheckResult{
		VEIDID: status.VEIDID,
		Level:  status.VerificationLevel,
	}

	// Map VEID status to KYC status
	switch status.Status {
	case "verified", "active":
		result.Status = KYCStatusVerified
	case "pending", "in_progress":
		result.Status = KYCStatusInProgress
	case statusFailed, "rejected":
		result.Status = KYCStatusFailed
	case "expired":
		result.Status = KYCStatusExpired
	default:
		result.Status = KYCStatusPending
	}

	if status.VerifiedAt != nil {
		result.VerifiedAt = status.VerifiedAt.Format(time.RFC3339)
	}

	if status.ExpiresAt != nil {
		result.ExpiresAt = status.ExpiresAt.Format(time.RFC3339)

		// Check if revalidation is needed
		if g.config.RevalidationDays > 0 && status.VerifiedAt != nil {
			revalidationDate := status.VerifiedAt.AddDate(0, 0, g.config.RevalidationDays)
			if time.Now().After(revalidationDate) {
				result.RequiresRevalidation = true
			}
		}
	}

	return result, nil
}

// GetVerificationLevel returns the verification level for an account.
func (g *VEIDKYCGate) GetVerificationLevel(ctx context.Context, accountAddress string) (KYCVerificationLevel, error) {
	if !g.config.Enabled {
		return KYCLevelFull, nil
	}

	status, err := g.veidChecker.GetVerificationStatus(ctx, accountAddress)
	if err != nil {
		return 0, err
	}

	if status == nil {
		return 0, ErrKYCNotVerified
	}

	return status.VerificationLevel, nil
}

// RequireVerification returns an error if verification is required but not met.
func (g *VEIDKYCGate) RequireVerification(ctx context.Context, accountAddress string, requiredLevel KYCVerificationLevel) error {
	if !g.config.Enabled {
		return nil
	}

	level, err := g.GetVerificationLevel(ctx, accountAddress)
	if err != nil {
		return err
	}

	if level < requiredLevel {
		return fmt.Errorf("%w: required level %d, current level %d", ErrKYCNotVerified, requiredLevel, level)
	}

	return nil
}

// ============================================================================
// AML Screener Implementation
// ============================================================================

// DefaultAMLScreener implements AMLScreener with configurable backend.
type DefaultAMLScreener struct {
	config AMLConfig
	client AMLClient
}

// AMLClient is an interface for AML screening providers.
type AMLClient interface {
	// Screen performs AML screening
	Screen(ctx context.Context, req *AMLClientRequest) (*AMLClientResponse, error)

	// GetScreeningStatus retrieves the status of a screening
	GetScreeningStatus(ctx context.Context, screeningID string) (*AMLClientResponse, error)
}

// AMLClientRequest is a request to the AML client.
type AMLClientRequest struct {
	FullName    string `json:"full_name"`
	DateOfBirth string `json:"date_of_birth,omitempty"`
	Country     string `json:"country"`
	EntityType  string `json:"entity_type"` // "individual" or "company"
}

// AMLClientResponse is a response from the AML client.
type AMLClientResponse struct {
	ScreeningID string     `json:"screening_id"`
	RiskScore   int        `json:"risk_score"`
	Status      string     `json:"status"`
	Matches     []AMLMatch `json:"matches,omitempty"`
	ScreenedAt  time.Time  `json:"screened_at"`
}

// NewDefaultAMLScreener creates a new AML screener.
func NewDefaultAMLScreener(config AMLConfig, client AMLClient) *DefaultAMLScreener {
	return &DefaultAMLScreener{
		config: config,
		client: client,
	}
}

// Screen performs AML screening on a user.
func (s *DefaultAMLScreener) Screen(ctx context.Context, req AMLScreenRequest) (*AMLScreenResult, error) {
	if !s.config.Enabled {
		// AML disabled, return cleared
		return &AMLScreenResult{
			ScreeningID: fmt.Sprintf("skip_%d", time.Now().UnixNano()),
			Status:      AMLStatusCleared,
			RiskScore:   0,
			ScreenedAt:  time.Now().Format(time.RFC3339),
		}, nil
	}

	// Call AML provider
	clientReq := &AMLClientRequest{
		FullName:    req.FullName,
		DateOfBirth: req.DateOfBirth,
		Country:     req.Country,
		EntityType:  "individual",
	}

	clientResp, err := s.client.Screen(ctx, clientReq)
	if err != nil {
		return nil, fmt.Errorf("AML screening failed: %w", err)
	}

	return s.mapScreeningResponse(clientResp), nil
}

// GetScreeningStatus retrieves the status of a screening.
func (s *DefaultAMLScreener) GetScreeningStatus(ctx context.Context, screeningID string) (*AMLScreenResult, error) {
	if !s.config.Enabled {
		return &AMLScreenResult{
			ScreeningID: screeningID,
			Status:      AMLStatusCleared,
			RiskScore:   0,
			ScreenedAt:  time.Now().Format(time.RFC3339),
			ExpiresAt:   time.Now().AddDate(0, 0, 30).Format(time.RFC3339),
		}, nil
	}
	if strings.TrimSpace(screeningID) == "" {
		return nil, fmt.Errorf("screening ID is required")
	}

	clientResp, err := s.client.GetScreeningStatus(ctx, screeningID)
	if err != nil {
		return nil, fmt.Errorf("AML screening status lookup failed: %w", err)
	}

	return s.mapScreeningResponse(clientResp), nil
}

func (s *DefaultAMLScreener) mapScreeningResponse(clientResp *AMLClientResponse) *AMLScreenResult {
	result := &AMLScreenResult{
		ScreeningID: clientResp.ScreeningID,
		RiskScore:   clientResp.RiskScore,
		Matches:     clientResp.Matches,
		ScreenedAt:  clientResp.ScreenedAt.Format(time.RFC3339),
	}

	screenedAt := clientResp.ScreenedAt
	if screenedAt.IsZero() {
		screenedAt = time.Now()
		result.ScreenedAt = screenedAt.Format(time.RFC3339)
	}

	status := normalizeAMLProviderStatus(clientResp.Status)
	switch status {
	case AMLStatusPending, AMLStatusScreening:
		result.Status = status
	case AMLStatusRejected:
		result.Status = AMLStatusRejected
		result.ReviewRequired = true
	case AMLStatusFlagged:
		result.Status = AMLStatusFlagged
		result.ReviewRequired = true
	default:
		if clientResp.RiskScore < s.config.AutoApproveBelow && len(clientResp.Matches) == 0 {
			result.Status = AMLStatusCleared
		} else if clientResp.RiskScore >= s.config.RiskThreshold || len(clientResp.Matches) > 0 {
			result.Status = AMLStatusFlagged
			result.ReviewRequired = true
		} else {
			result.Status = AMLStatusCleared
		}
	}

	expiresAt := screenedAt.AddDate(0, 0, 30)
	result.ExpiresAt = expiresAt.Format(time.RFC3339)
	return result
}

func normalizeAMLProviderStatus(status string) AMLStatus {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", "cleared", "approved", "complete", "completed":
		return AMLStatusCleared
	case "pending", "queued":
		return AMLStatusPending
	case "screening", "in_progress", "processing", "reviewing":
		return AMLStatusScreening
	case "flagged", "review_required", "manual_review":
		return AMLStatusFlagged
	case "rejected", "denied", "blocked", "failed":
		return AMLStatusRejected
	default:
		return AMLStatusPending
	}
}

// HTTPAMLClient implements AML provider calls over HTTP.
type HTTPAMLClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewHTTPAMLClient creates a new HTTP AML client from config.
func NewHTTPAMLClient(config AMLConfig) (*HTTPAMLClient, error) {
	if strings.TrimSpace(config.APIURL) == "" || strings.TrimSpace(config.APIKey) == "" {
		return nil, ErrProviderNotConfigured
	}

	return &HTTPAMLClient{
		baseURL:    strings.TrimRight(config.APIURL, "/"),
		apiKey:     config.APIKey,
		httpClient: security.NewSecureHTTPClient(security.WithTimeout(30 * time.Second)),
	}, nil
}

// Screen performs AML screening via the configured HTTP provider.
func (c *HTTPAMLClient) Screen(ctx context.Context, req *AMLClientRequest) (*AMLClientResponse, error) {
	resp, err := c.doJSON(ctx, http.MethodPost, "/screenings", req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetScreeningStatus retrieves the status of a screening via the configured HTTP provider.
func (c *HTTPAMLClient) GetScreeningStatus(ctx context.Context, screeningID string) (*AMLClientResponse, error) {
	if strings.TrimSpace(screeningID) == "" {
		return nil, fmt.Errorf("screening ID is required")
	}
	return c.doJSON(ctx, http.MethodGet, "/screenings/"+screeningID, nil)
}

func (c *HTTPAMLClient) doJSON(ctx context.Context, method string, path string, payload any) (*AMLClientResponse, error) {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal AML request: %w", err)
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create AML request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AML provider request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read AML response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("AML provider returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var raw struct {
		ID          string     `json:"id"`
		ScreeningID string     `json:"screening_id"`
		RiskScore   int        `json:"risk_score"`
		Status      string     `json:"status"`
		Matches     []AMLMatch `json:"matches"`
		ScreenedAt  string     `json:"screened_at"`
		CreatedAt   string     `json:"created_at"`
	}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return nil, fmt.Errorf("failed to decode AML response: %w", err)
	}

	screenedAt := time.Now()
	for _, candidate := range []string{raw.ScreenedAt, raw.CreatedAt} {
		if strings.TrimSpace(candidate) == "" {
			continue
		}
		parsed, err := time.Parse(time.RFC3339, candidate)
		if err == nil {
			screenedAt = parsed
			break
		}
	}

	screeningID := strings.TrimSpace(raw.ScreeningID)
	if screeningID == "" {
		screeningID = strings.TrimSpace(raw.ID)
	}

	return &AMLClientResponse{
		ScreeningID: screeningID,
		RiskScore:   raw.RiskScore,
		Status:      raw.Status,
		Matches:     raw.Matches,
		ScreenedAt:  screenedAt,
	}, nil
}

// ============================================================================
// Mock Implementations for Testing
// ============================================================================

// MockVEIDChecker is a mock implementation of VEIDChecker for testing.
type MockVEIDChecker struct {
	Statuses map[string]*VEIDVerificationStatus
}

// NewMockVEIDChecker creates a new mock VEID checker.
func NewMockVEIDChecker() *MockVEIDChecker {
	return &MockVEIDChecker{
		Statuses: make(map[string]*VEIDVerificationStatus),
	}
}

// GetVerificationStatus returns mock verification status.
func (m *MockVEIDChecker) GetVerificationStatus(ctx context.Context, accountAddress string) (*VEIDVerificationStatus, error) {
	if status, ok := m.Statuses[accountAddress]; ok {
		return status, nil
	}
	return nil, nil
}

// GetVerificationByID returns mock verification by ID.
func (m *MockVEIDChecker) GetVerificationByID(ctx context.Context, veidID string) (*VEIDVerificationStatus, error) {
	for _, status := range m.Statuses {
		if status.VEIDID == veidID {
			return status, nil
		}
	}
	return nil, nil
}

// SetVerified sets an account as verified for testing.
func (m *MockVEIDChecker) SetVerified(accountAddress string, level KYCVerificationLevel) {
	now := time.Now()
	expires := now.AddDate(1, 0, 0) // 1 year
	m.Statuses[accountAddress] = &VEIDVerificationStatus{
		VEIDID:            fmt.Sprintf("veid_%s", accountAddress[:8]),
		AccountAddress:    accountAddress,
		VerificationLevel: level,
		Status:            "verified",
		VerifiedAt:        &now,
		ExpiresAt:         &expires,
		Score:             85,
	}
}

// MockAMLClient is a mock implementation of AMLClient for testing.
type MockAMLClient struct {
	DefaultScore int
	Matches      []AMLMatch
	Statuses     map[string]*AMLClientResponse
}

// NewMockAMLClient creates a new mock AML client.
func NewMockAMLClient() *MockAMLClient {
	return &MockAMLClient{
		DefaultScore: 0,
		Matches:      nil,
		Statuses:     make(map[string]*AMLClientResponse),
	}
}

// Screen performs mock AML screening.
func (m *MockAMLClient) Screen(ctx context.Context, req *AMLClientRequest) (*AMLClientResponse, error) {
	resp := &AMLClientResponse{
		ScreeningID: fmt.Sprintf("scr_%d", time.Now().UnixNano()),
		RiskScore:   m.DefaultScore,
		Status:      "completed",
		Matches:     m.Matches,
		ScreenedAt:  time.Now(),
	}
	m.Statuses[resp.ScreeningID] = resp
	return resp, nil
}

// GetScreeningStatus returns mock AML status by screening ID.
func (m *MockAMLClient) GetScreeningStatus(ctx context.Context, screeningID string) (*AMLClientResponse, error) {
	if resp, ok := m.Statuses[screeningID]; ok {
		return resp, nil
	}
	return nil, fmt.Errorf("screening %s not found", screeningID)
}

// Ensure implementations satisfy interfaces
var (
	_ KYCGate     = (*VEIDKYCGate)(nil)
	_ AMLScreener = (*DefaultAMLScreener)(nil)
	_ VEIDChecker = (*MockVEIDChecker)(nil)
	_ AMLClient   = (*MockAMLClient)(nil)
)
