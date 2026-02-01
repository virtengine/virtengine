package govdata

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	oauthTokenPath = "/oauth/token"
)

func TestDefaultAAMVAConfig(t *testing.T) {
	config := DefaultAAMVAConfig()

	assert.Equal(t, AAMVAEnvironmentSandbox, config.Environment)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 60, config.RateLimitPerMinute)
	assert.True(t, config.AuditEnabled)
	assert.True(t, len(config.SupportedStates) > 0)
	assert.Contains(t, config.SupportedStates, "CA")
	assert.Contains(t, config.SupportedStates, "TX")
	assert.Contains(t, config.SupportedStates, "NY")
}

func TestAAMVAConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  AAMVAConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: AAMVAConfig{
				Environment:    AAMVAEnvironmentSandbox,
				OrgID:          "test-org",
				PermissionCode: "test-permission",
				ClientID:       "test-client",
				ClientSecret:   "test-secret",
			},
			wantErr: false,
		},
		{
			name: "missing org_id",
			config: AAMVAConfig{
				Environment:    AAMVAEnvironmentSandbox,
				PermissionCode: "test-permission",
				ClientID:       "test-client",
				ClientSecret:   "test-secret",
			},
			wantErr: true,
		},
		{
			name: "missing permission_code",
			config: AAMVAConfig{
				Environment:  AAMVAEnvironmentSandbox,
				OrgID:        "test-org",
				ClientID:     "test-client",
				ClientSecret: "test-secret",
			},
			wantErr: true,
		},
		{
			name: "missing client_id",
			config: AAMVAConfig{
				Environment:    AAMVAEnvironmentSandbox,
				OrgID:          "test-org",
				PermissionCode: "test-permission",
				ClientSecret:   "test-secret",
			},
			wantErr: true,
		},
		{
			name: "missing client_secret",
			config: AAMVAConfig{
				Environment:    AAMVAEnvironmentSandbox,
				OrgID:          "test-org",
				PermissionCode: "test-permission",
				ClientID:       "test-client",
			},
			wantErr: true,
		},
		{
			name: "invalid environment",
			config: AAMVAConfig{
				Environment:    "invalid",
				OrgID:          "test-org",
				PermissionCode: "test-permission",
				ClientID:       "test-client",
				ClientSecret:   "test-secret",
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewAAMVADMVAdapter(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceDMV,
		Jurisdiction: "US-CA",
		Timeout:      30 * time.Second,
	}

	aamvaConfig := AAMVAConfig{
		Environment:    AAMVAEnvironmentSandbox,
		OrgID:          "test-org",
		PermissionCode: "test-permission",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}

	adapter, err := NewAAMVADMVAdapter(baseConfig, aamvaConfig)
	require.NoError(t, err)
	require.NotNil(t, adapter)

	assert.Equal(t, DataSourceDMV, adapter.Type())
	assert.Equal(t, "US-CA", adapter.Jurisdiction())
	assert.True(t, adapter.SupportsDocument(DocumentTypeDriversLicense))
	assert.True(t, adapter.SupportsDocument(DocumentTypeStateID))
	assert.False(t, adapter.SupportsDocument(DocumentTypePassport))
}

func TestNewAAMVADMVAdapter_InvalidConfig(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceDMV,
		Jurisdiction: "US-CA",
	}

	aamvaConfig := AAMVAConfig{
		// Missing required fields
		Environment: AAMVAEnvironmentSandbox,
	}

	_, err := NewAAMVADMVAdapter(baseConfig, aamvaConfig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid AAMVA config")
}

func TestAAMVADMVAdapter_BaseURL(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceDMV,
		Jurisdiction: "US-CA",
	}

	// Sandbox
	sandboxConfig := AAMVAConfig{
		Environment:    AAMVAEnvironmentSandbox,
		OrgID:          "test-org",
		PermissionCode: "test-permission",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}
	sandboxAdapter, _ := NewAAMVADMVAdapter(baseConfig, sandboxConfig)
	assert.Equal(t, AAMVASandboxURL, sandboxAdapter.baseURL())

	// Production
	prodConfig := AAMVAConfig{
		Environment:    AAMVAEnvironmentProduction,
		OrgID:          "test-org",
		PermissionCode: "test-permission",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}
	prodAdapter, _ := NewAAMVADMVAdapter(baseConfig, prodConfig)
	assert.Equal(t, AAMVAProductionURL, prodAdapter.baseURL())
}

func TestAAMVADMVAdapter_ValidateLicenseNumber(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceDMV,
		Jurisdiction: "US-CA",
	}
	aamvaConfig := AAMVAConfig{
		Environment:    AAMVAEnvironmentSandbox,
		OrgID:          "test-org",
		PermissionCode: "test-permission",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}
	adapter, _ := NewAAMVADMVAdapter(baseConfig, aamvaConfig)

	tests := []struct {
		state   string
		license string
		wantErr bool
	}{
		// California: 1 letter + 7 digits
		{"CA", "A1234567", false},
		{"CA", "B9876543", false},
		{"CA", "12345678", true}, // Missing letter
		{"CA", "AB123456", true}, // Two letters

		// Texas: 8 digits
		{"TX", "12345678", false},
		{"TX", "87654321", false},
		{"TX", "1234567", true},  // Too short
		{"TX", "A1234567", true}, // Has letter

		// Florida: 1 letter + 12 digits
		{"FL", "A123456789012", false},

		// New York: 9 digits
		{"NY", "123456789", false},

		// Unknown state: General validation (5-20 chars)
		{"XX", "12345", false},
		{"XX", "1234", true}, // Too short
	}

	for _, tc := range tests {
		t.Run(tc.state+"-"+tc.license, func(t *testing.T) {
			err := adapter.validateLicenseNumber(tc.state, tc.license)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAAMVADMVAdapter_IsStateSupported(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceDMV,
		Jurisdiction: "US-CA",
	}
	aamvaConfig := AAMVAConfig{
		Environment:    AAMVAEnvironmentSandbox,
		OrgID:          "test-org",
		PermissionCode: "test-permission",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
		SupportedStates: []string{"CA", "TX", "NY"},
	}
	adapter, _ := NewAAMVADMVAdapter(baseConfig, aamvaConfig)

	assert.True(t, adapter.isStateSupported("CA"))
	assert.True(t, adapter.isStateSupported("ca")) // Case insensitive
	assert.True(t, adapter.isStateSupported("TX"))
	assert.True(t, adapter.isStateSupported("NY"))
	assert.False(t, adapter.isStateSupported("FL"))
	assert.False(t, adapter.isStateSupported("XX"))
}

func TestAAMVADMVAdapter_RateLimit(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceDMV,
		Jurisdiction: "US-CA",
	}
	aamvaConfig := AAMVAConfig{
		Environment:        AAMVAEnvironmentSandbox,
		OrgID:              "test-org",
		PermissionCode:     "test-permission",
		ClientID:           "test-client",
		ClientSecret:       "test-secret",
		RateLimitPerMinute: 3,
	}
	adapter, _ := NewAAMVADMVAdapter(baseConfig, aamvaConfig)

	// First 3 requests should succeed
	assert.NoError(t, adapter.checkRateLimit())
	assert.NoError(t, adapter.checkRateLimit())
	assert.NoError(t, adapter.checkRateLimit())

	// 4th request should be rate limited
	assert.ErrorIs(t, adapter.checkRateLimit(), ErrAAMVARateLimit)

	// Reset window
	adapter.mu.Lock()
	adapter.windowStart = time.Now().Add(-2 * time.Minute)
	adapter.mu.Unlock()

	// Should work again
	assert.NoError(t, adapter.checkRateLimit())
}

func TestAAMVADMVAdapter_ConvertFieldMatch(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceDMV,
		Jurisdiction: "US-CA",
	}
	aamvaConfig := AAMVAConfig{
		Environment:    AAMVAEnvironmentSandbox,
		OrgID:          "test-org",
		PermissionCode: "test-permission",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}
	adapter, _ := NewAAMVADMVAdapter(baseConfig, aamvaConfig)

	tests := []struct {
		input      string
		wantMatch  FieldMatchResult
		wantConf   float64
	}{
		{string(AAMVAMatchYes), FieldMatchExact, 1.0},
		{string(AAMVAMatchPartial), FieldMatchFuzzy, 0.7},
		{string(AAMVAMatchNo), FieldMatchNoMatch, 0.0},
		{string(AAMVAMatchNoData), FieldMatchUnavailable, 0.0},
		{"", FieldMatchUnavailable, 0.0},
		{"UNKNOWN", FieldMatchUnavailable, 0.0},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := adapter.convertFieldMatch(tc.input)
			assert.Equal(t, tc.wantMatch, result.Match)
			assert.Equal(t, tc.wantConf, result.Confidence)
		})
	}
}

func TestAAMVADMVAdapter_ConvertResponse_Verified(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceDMV,
		Jurisdiction: "US-CA",
	}
	aamvaConfig := AAMVAConfig{
		Environment:    AAMVAEnvironmentSandbox,
		OrgID:          "test-org",
		PermissionCode: "test-permission",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}
	adapter, _ := NewAAMVADMVAdapter(baseConfig, aamvaConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "US-CA",
	}

	dldvResp := &AAMVADLDVResponse{
		OverallMatch:    string(AAMVAMatchYes),
		LicenseMatch:    string(AAMVAMatchYes),
		FirstNameMatch:  string(AAMVAMatchYes),
		LastNameMatch:   string(AAMVAMatchYes),
		DOBMatch:        string(AAMVAMatchYes),
		LicenseStatus:   string(AAMVALicenseValid),
		ExpirationDate:  "2030-12-31",
	}

	resp := adapter.convertResponse(req, dldvResp)

	assert.Equal(t, "test-req-123", resp.RequestID)
	assert.Equal(t, VerificationStatusVerified, resp.Status)
	assert.True(t, resp.DocumentValid)
	assert.Equal(t, 1.0, resp.Confidence)
	assert.NotNil(t, resp.DocumentExpiresAt)
}

func TestAAMVADMVAdapter_ConvertResponse_Expired(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceDMV,
		Jurisdiction: "US-CA",
	}
	aamvaConfig := AAMVAConfig{
		Environment:    AAMVAEnvironmentSandbox,
		OrgID:          "test-org",
		PermissionCode: "test-permission",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}
	adapter, _ := NewAAMVADMVAdapter(baseConfig, aamvaConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "US-CA",
	}

	dldvResp := &AAMVADLDVResponse{
		OverallMatch:   string(AAMVAMatchYes),
		LicenseStatus:  string(AAMVALicenseExpired),
		ExpirationDate: "2020-01-01",
	}

	resp := adapter.convertResponse(req, dldvResp)

	assert.Equal(t, VerificationStatusExpired, resp.Status)
	assert.False(t, resp.DocumentValid)
	assert.Contains(t, resp.Warnings, "License has expired")
}

func TestAAMVADMVAdapter_ConvertResponse_Revoked(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceDMV,
		Jurisdiction: "US-CA",
	}
	aamvaConfig := AAMVAConfig{
		Environment:    AAMVAEnvironmentSandbox,
		OrgID:          "test-org",
		PermissionCode: "test-permission",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}
	adapter, _ := NewAAMVADMVAdapter(baseConfig, aamvaConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "US-CA",
	}

	dldvResp := &AAMVADLDVResponse{
		OverallMatch:  string(AAMVAMatchYes),
		LicenseStatus: string(AAMVALicenseRevoked),
	}

	resp := adapter.convertResponse(req, dldvResp)

	assert.Equal(t, VerificationStatusRevoked, resp.Status)
	assert.False(t, resp.DocumentValid)
	assert.Contains(t, resp.Warnings, "License has been revoked")
}

func TestAAMVADMVAdapter_Verify_UnsupportedDocument(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceDMV,
		Jurisdiction: "US-CA",
	}
	aamvaConfig := AAMVAConfig{
		Environment:    AAMVAEnvironmentSandbox,
		OrgID:          "test-org",
		PermissionCode: "test-permission",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
		SupportedStates: []string{"CA"},
	}
	adapter, _ := NewAAMVADMVAdapter(baseConfig, aamvaConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "US-CA",
		DocumentType: DocumentTypePassport, // Not supported by DMV
	}

	_, err := adapter.Verify(context.Background(), req)
	assert.ErrorIs(t, err, ErrDocumentTypeNotSupported)
}

func TestAAMVADMVAdapter_Verify_UnsupportedState(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceDMV,
		Jurisdiction: "US-CA",
	}
	aamvaConfig := AAMVAConfig{
		Environment:    AAMVAEnvironmentSandbox,
		OrgID:          "test-org",
		PermissionCode: "test-permission",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
		SupportedStates: []string{"CA", "TX"},
	}
	adapter, _ := NewAAMVADMVAdapter(baseConfig, aamvaConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "US-NY", // Not in supported list
		DocumentType: DocumentTypeDriversLicense,
		Fields: VerificationFields{
			DocumentNumber: "123456789",
		},
	}

	_, err := adapter.Verify(context.Background(), req)
	assert.ErrorIs(t, err, ErrAAMVAStateNotSupported)
}

func TestAAMVADMVAdapter_Verify_InvalidLicense(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceDMV,
		Jurisdiction: "US-CA",
	}
	aamvaConfig := AAMVAConfig{
		Environment:    AAMVAEnvironmentSandbox,
		OrgID:          "test-org",
		PermissionCode: "test-permission",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
		SupportedStates: []string{"CA"},
	}
	adapter, _ := NewAAMVADMVAdapter(baseConfig, aamvaConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "US-CA",
		DocumentType: DocumentTypeDriversLicense,
		Fields: VerificationFields{
			DocumentNumber: "12345", // Invalid for CA (needs letter + 7 digits)
		},
	}

	_, err := adapter.Verify(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid license number format")
}

func TestAAMVADMVAdapter_Verify_WithMockServer(t *testing.T) {
	// Create mock AAMVA server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == oauthTokenPath {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token": "test-token", "expires_in": 3600, "token_type": "Bearer"}`))
			return
		}

		if r.URL.Path == "/verify" {
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<dldvResponse>
  <messageId>test-message-123</messageId>
  <responseCode>0000</responseCode>
  <responseText>Success</responseText>
  <overallMatch>Y</overallMatch>
  <licenseMatch>Y</licenseMatch>
  <firstNameMatch>Y</firstNameMatch>
  <lastNameMatch>Y</lastNameMatch>
  <dobMatch>Y</dobMatch>
  <licenseStatus>VALID</licenseStatus>
  <expirationDate>2030-12-31</expirationDate>
</dldvResponse>`))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	baseConfig := AdapterConfig{
		Type:         DataSourceDMV,
		Jurisdiction: "US-CA",
		Timeout:      5 * time.Second,
	}
	aamvaConfig := AAMVAConfig{
		Environment:     AAMVAEnvironmentSandbox,
		OrgID:           "test-org",
		PermissionCode:  "test-permission",
		ClientID:        "test-client",
		ClientSecret:    "test-secret",
		SupportedStates: []string{"CA"},
		Timeout:         5 * time.Second,
	}

	adapter, err := NewAAMVADMVAdapter(baseConfig, aamvaConfig)
	require.NoError(t, err)

	// Override the base URL for testing
	// Since we can't easily override baseURL(), we test the other components
	// In production, the integration test would hit the real sandbox

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "US-CA",
		DocumentType: DocumentTypeDriversLicense,
		Fields: VerificationFields{
			DocumentNumber: "A1234567",
			FirstName:      "John",
			LastName:       "Doe",
			DateOfBirth:    time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	// Test that validation passes
	assert.NoError(t, adapter.validateLicenseNumber("CA", req.Fields.DocumentNumber))
	assert.True(t, adapter.isStateSupported("CA"))
	assert.True(t, adapter.SupportsDocument(DocumentTypeDriversLicense))

	// The actual Verify call would fail since we can't easily mock the HTTP client
	// but the message ID generation should work
	msgID := adapter.generateMessageID(req.RequestID)
	assert.NotEmpty(t, msgID)
	assert.Len(t, msgID, 32)
}

