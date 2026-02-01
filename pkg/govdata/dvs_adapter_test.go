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

// ============================================================================
// DVS Adapter Tests
// ============================================================================

func TestDefaultDVSConfig(t *testing.T) {
	config := DefaultDVSConfig()

	assert.Equal(t, DVSEnvironmentSandbox, config.Environment)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 60, config.RateLimitPerMinute)
	assert.True(t, config.AuditEnabled)
	assert.True(t, len(config.SupportedStates) > 0)
	assert.Contains(t, config.SupportedStates, "NSW")
	assert.Contains(t, config.SupportedStates, "VIC")
	assert.Contains(t, config.SupportedStates, "QLD")
}

func TestDVSConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  DVSConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: DVSConfig{
				Environment:    DVSEnvironmentSandbox,
				OrganisationID: "test-org",
				ClientID:       "test-client",
				ClientSecret:   "test-secret",
			},
			wantErr: false,
		},
		{
			name: "missing organisation_id",
			config: DVSConfig{
				Environment:  DVSEnvironmentSandbox,
				ClientID:     "test-client",
				ClientSecret: "test-secret",
			},
			wantErr: true,
		},
		{
			name: "missing client_id",
			config: DVSConfig{
				Environment:    DVSEnvironmentSandbox,
				OrganisationID: "test-org",
				ClientSecret:   "test-secret",
			},
			wantErr: true,
		},
		{
			name: "missing client_secret",
			config: DVSConfig{
				Environment:    DVSEnvironmentSandbox,
				OrganisationID: "test-org",
				ClientID:       "test-client",
			},
			wantErr: true,
		},
		{
			name: "invalid environment",
			config: DVSConfig{
				Environment:    "invalid",
				OrganisationID: "test-org",
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

func TestNewDVSDMVAdapter(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceDVS,
		Jurisdiction: "AU-NSW",
		Timeout:      30 * time.Second,
	}

	dvsConfig := DVSConfig{
		Environment:    DVSEnvironmentSandbox,
		OrganisationID: "test-org",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}

	adapter, err := NewDVSDMVAdapter(baseConfig, dvsConfig)
	require.NoError(t, err)
	require.NotNil(t, adapter)

	assert.Equal(t, DataSourceDVS, adapter.Type())
	assert.Equal(t, "AU-NSW", adapter.Jurisdiction())
	assert.True(t, adapter.SupportsDocument(DocumentTypeDriversLicense))
	assert.True(t, adapter.SupportsDocument(DocumentTypePassport))
	assert.True(t, adapter.SupportsDocument(DocumentTypeBirthCertificate))
}

func TestNewDVSDMVAdapter_InvalidConfig(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceDVS,
		Jurisdiction: "AU-NSW",
	}

	dvsConfig := DVSConfig{
		Environment: DVSEnvironmentSandbox,
	}

	_, err := NewDVSDMVAdapter(baseConfig, dvsConfig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid DVS config")
}

func TestDVSDMVAdapter_BaseURL(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceDVS,
		Jurisdiction: "AU-NSW",
	}

	// Sandbox
	sandboxConfig := DVSConfig{
		Environment:    DVSEnvironmentSandbox,
		OrganisationID: "test-org",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}
	sandboxAdapter, _ := NewDVSDMVAdapter(baseConfig, sandboxConfig)
	assert.Equal(t, DVSSandboxURL, sandboxAdapter.baseURL())

	// Production
	prodConfig := DVSConfig{
		Environment:    DVSEnvironmentProduction,
		OrganisationID: "test-org",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}
	prodAdapter, _ := NewDVSDMVAdapter(baseConfig, prodConfig)
	assert.Equal(t, DVSProductionURL, prodAdapter.baseURL())
}

func TestDVSDMVAdapter_IsStateSupported(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceDVS,
		Jurisdiction: "AU-NSW",
	}
	dvsConfig := DVSConfig{
		Environment:     DVSEnvironmentSandbox,
		OrganisationID:  "test-org",
		ClientID:        "test-client",
		ClientSecret:    "test-secret",
		SupportedStates: []string{"NSW", "VIC", "QLD"},
	}
	adapter, _ := NewDVSDMVAdapter(baseConfig, dvsConfig)

	assert.True(t, adapter.isStateSupported("NSW"))
	assert.True(t, adapter.isStateSupported("nsw")) // Case insensitive
	assert.True(t, adapter.isStateSupported("VIC"))
	assert.True(t, adapter.isStateSupported("QLD"))
	assert.False(t, adapter.isStateSupported("SA"))
	assert.False(t, adapter.isStateSupported("XX"))
}

func TestDVSDMVAdapter_ExtractState(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceDVS,
		Jurisdiction: "AU-NSW",
	}
	dvsConfig := DVSConfig{
		Environment:    DVSEnvironmentSandbox,
		OrganisationID: "test-org",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}
	adapter, _ := NewDVSDMVAdapter(baseConfig, dvsConfig)

	tests := []struct {
		jurisdiction string
		expected     string
	}{
		{"AU-NSW", "NSW"},
		{"AU-VIC", "VIC"},
		{"NSW", ""},
		{"AU-qld", "QLD"},
	}

	for _, tc := range tests {
		result := adapter.extractState(tc.jurisdiction)
		assert.Equal(t, tc.expected, result)
	}
}

func TestDVSDMVAdapter_RateLimit(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceDVS,
		Jurisdiction: "AU-NSW",
	}
	dvsConfig := DVSConfig{
		Environment:        DVSEnvironmentSandbox,
		OrganisationID:     "test-org",
		ClientID:           "test-client",
		ClientSecret:       "test-secret",
		RateLimitPerMinute: 3,
	}
	adapter, _ := NewDVSDMVAdapter(baseConfig, dvsConfig)

	// First 3 requests should succeed
	assert.NoError(t, adapter.checkRateLimit())
	assert.NoError(t, adapter.checkRateLimit())
	assert.NoError(t, adapter.checkRateLimit())

	// 4th request should be rate limited
	assert.ErrorIs(t, adapter.checkRateLimit(), ErrDVSRateLimit)

	// Reset window
	adapter.mu.Lock()
	adapter.windowStart = time.Now().Add(-2 * time.Minute)
	adapter.mu.Unlock()

	// Should work again
	assert.NoError(t, adapter.checkRateLimit())
}

func TestDVSDMVAdapter_Verify_UnsupportedDocument(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceDVS,
		Jurisdiction: "AU-NSW",
	}
	dvsConfig := DVSConfig{
		Environment:     DVSEnvironmentSandbox,
		OrganisationID:  "test-org",
		ClientID:        "test-client",
		ClientSecret:    "test-secret",
		SupportedStates: []string{"NSW"},
	}
	adapter, _ := NewDVSDMVAdapter(baseConfig, dvsConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "AU-NSW",
		DocumentType: DocumentTypeTaxID, // Not supported by DVS
	}

	_, err := adapter.Verify(context.Background(), req)
	assert.ErrorIs(t, err, ErrDocumentTypeNotSupported)
}

func TestDVSDMVAdapter_Verify_UnsupportedState(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceDVS,
		Jurisdiction: "AU-NSW",
	}
	dvsConfig := DVSConfig{
		Environment:     DVSEnvironmentSandbox,
		OrganisationID:  "test-org",
		ClientID:        "test-client",
		ClientSecret:    "test-secret",
		SupportedStates: []string{"NSW", "VIC"},
	}
	adapter, _ := NewDVSDMVAdapter(baseConfig, dvsConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "AU-SA", // Not in supported list
		DocumentType: DocumentTypeDriversLicense,
		Fields: VerificationFields{
			DocumentNumber: "12345678",
		},
	}

	_, err := adapter.Verify(context.Background(), req)
	assert.ErrorIs(t, err, ErrDVSStateNotSupported)
}

func TestDVSDMVAdapter_ConvertResponse_Match(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceDVS,
		Jurisdiction: "AU-NSW",
	}
	dvsConfig := DVSConfig{
		Environment:    DVSEnvironmentSandbox,
		OrganisationID: "test-org",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}
	adapter, _ := NewDVSDMVAdapter(baseConfig, dvsConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "AU-NSW",
	}

	dvsResp := &DVSVerifyResponse{
		RequestID: "test-req-123",
		Status:    "SUCCESS",
		VerifyResult: DVSVerifyResult{
			OverallResult: DVSResultMatch,
			DocumentValid: true,
			FieldResults: map[string]string{
				"familyName": DVSResultMatch,
				"givenNames": DVSResultMatch,
			},
		},
	}

	resp := adapter.convertResponse(req, dvsResp)

	assert.Equal(t, "test-req-123", resp.RequestID)
	assert.Equal(t, VerificationStatusVerified, resp.Status)
	assert.True(t, resp.DocumentValid)
	assert.Equal(t, 1.0, resp.Confidence)
}

func TestDVSDMVAdapter_ConvertResponse_NoMatch(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceDVS,
		Jurisdiction: "AU-NSW",
	}
	dvsConfig := DVSConfig{
		Environment:    DVSEnvironmentSandbox,
		OrganisationID: "test-org",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}
	adapter, _ := NewDVSDMVAdapter(baseConfig, dvsConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "AU-NSW",
	}

	dvsResp := &DVSVerifyResponse{
		RequestID: "test-req-123",
		Status:    "FAILED",
		VerifyResult: DVSVerifyResult{
			OverallResult: DVSResultNoMatch,
			DocumentValid: false,
		},
	}

	resp := adapter.convertResponse(req, dvsResp)

	assert.Equal(t, VerificationStatusNotVerified, resp.Status)
	assert.False(t, resp.DocumentValid)
	assert.Equal(t, 0.0, resp.Confidence)
}

func TestDVSDMVAdapter_GenerateRequestID(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceDVS,
		Jurisdiction: "AU-NSW",
	}
	dvsConfig := DVSConfig{
		Environment:    DVSEnvironmentSandbox,
		OrganisationID: "test-org",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}
	adapter, _ := NewDVSDMVAdapter(baseConfig, dvsConfig)

	reqID := adapter.generateRequestID("test-123")
	assert.NotEmpty(t, reqID)
	assert.Len(t, reqID, 32)

	// Different inputs should produce different IDs
	reqID2 := adapter.generateRequestID("test-456")
	assert.NotEqual(t, reqID, reqID2)
}

func TestDVSDMVAdapter_WithMockServer(t *testing.T) {
	// Create mock DVS server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token": "test-token", "expires_in": 3600, "token_type": "Bearer"}`))
			return
		}

		if r.URL.Path == "/verify" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"requestId": "test-123",
				"verificationId": "ver-456",
				"status": "SUCCESS",
				"verifyResult": {
					"overallResult": "MATCH",
					"documentValid": true,
					"fieldResults": {
						"familyName": "MATCH",
						"givenNames": "MATCH"
					}
				},
				"timestamp": "2024-01-01T00:00:00Z"
			}`))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	baseConfig := AdapterConfig{
		Type:         DataSourceDVS,
		Jurisdiction: "AU-NSW",
		Timeout:      5 * time.Second,
	}
	dvsConfig := DVSConfig{
		Environment:        DVSEnvironmentSandbox,
		OrganisationID:     "test-org",
		ClientID:           "test-client",
		ClientSecret:       "test-secret",
		SupportedStates:    []string{"NSW"},
		Timeout:            5 * time.Second,
		BaseURL:            server.URL,
		RateLimitPerMinute: 60,
	}

	adapter, err := NewDVSDMVAdapter(baseConfig, dvsConfig)
	require.NoError(t, err)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "AU-NSW",
		DocumentType: DocumentTypeDriversLicense,
		Fields: VerificationFields{
			DocumentNumber: "12345678",
			FirstName:      "John",
			LastName:       "Doe",
			DateOfBirth:    time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	resp, err := adapter.Verify(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, VerificationStatusVerified, resp.Status)
	assert.True(t, resp.DocumentValid)
	assert.Equal(t, 1.0, resp.Confidence)
}
