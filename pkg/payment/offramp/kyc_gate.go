// Package offramp provides fiat off-ramp integration for token-to-fiat payouts.
//
// VE-5E: Fiat off-ramp integration for PayPal/ACH payouts
package offramp

import (
	"context"
	"fmt"
	"time"
)

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
	case "failed", "rejected":
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

	result := &AMLScreenResult{
		ScreeningID: clientResp.ScreeningID,
		RiskScore:   clientResp.RiskScore,
		Matches:     clientResp.Matches,
		ScreenedAt:  clientResp.ScreenedAt.Format(time.RFC3339),
	}

	// Determine status based on risk score and matches
	if clientResp.RiskScore < s.config.AutoApproveBelow && len(clientResp.Matches) == 0 {
		result.Status = AMLStatusCleared
	} else if clientResp.RiskScore >= s.config.RiskThreshold || len(clientResp.Matches) > 0 {
		result.Status = AMLStatusFlagged
		result.ReviewRequired = true
	} else {
		result.Status = AMLStatusCleared
	}

	// Set expiry (typically 30-90 days)
	expiresAt := time.Now().AddDate(0, 0, 30)
	result.ExpiresAt = expiresAt.Format(time.RFC3339)

	return result, nil
}

// GetScreeningStatus retrieves the status of a screening.
func (s *DefaultAMLScreener) GetScreeningStatus(ctx context.Context, screeningID string) (*AMLScreenResult, error) {
	// In a full implementation, this would query the AML provider
	// For now, return a placeholder
	return nil, fmt.Errorf("screening status lookup not implemented")
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
}

// NewMockAMLClient creates a new mock AML client.
func NewMockAMLClient() *MockAMLClient {
	return &MockAMLClient{
		DefaultScore: 0,
		Matches:      nil,
	}
}

// Screen performs mock AML screening.
func (m *MockAMLClient) Screen(ctx context.Context, req *AMLClientRequest) (*AMLClientResponse, error) {
	return &AMLClientResponse{
		ScreeningID: fmt.Sprintf("scr_%d", time.Now().UnixNano()),
		RiskScore:   m.DefaultScore,
		Status:      "completed",
		Matches:     m.Matches,
		ScreenedAt:  time.Now(),
	}, nil
}

// Ensure implementations satisfy interfaces
var (
	_ KYCGate     = (*VEIDKYCGate)(nil)
	_ AMLScreener = (*DefaultAMLScreener)(nil)
	_ VEIDChecker = (*MockVEIDChecker)(nil)
	_ AMLClient   = (*MockAMLClient)(nil)
)
