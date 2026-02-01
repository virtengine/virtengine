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
// GOV.UK Verify Adapter Tests
// ============================================================================

func TestDefaultGovUKConfig(t *testing.T) {
	config := DefaultGovUKConfig()

	assert.Equal(t, GovUKEnvironmentSandbox, config.Environment)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 60, config.RateLimitPerMinute)
	assert.Equal(t, GovUKLoA2, config.LevelOfAssurance)
	assert.True(t, config.AuditEnabled)
}

func TestGovUKConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  GovUKConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: GovUKConfig{
				Environment:     GovUKEnvironmentSandbox,
				ServiceEntityID: "test-service",
				APIKey:          "test-key",
			},
			wantErr: false,
		},
		{
			name: "missing service_entity_id",
			config: GovUKConfig{
				Environment: GovUKEnvironmentSandbox,
				APIKey:      "test-key",
			},
			wantErr: true,
		},
		{
			name: "missing api_key",
			config: GovUKConfig{
				Environment:     GovUKEnvironmentSandbox,
				ServiceEntityID: "test-service",
			},
			wantErr: true,
		},
		{
			name: "invalid environment",
			config: GovUKConfig{
				Environment:     "invalid",
				ServiceEntityID: "test-service",
				APIKey:          "test-key",
			},
			wantErr: true,
		},
		{
			name: "integration environment",
			config: GovUKConfig{
				Environment:     GovUKEnvironmentIntegration,
				ServiceEntityID: "test-service",
				APIKey:          "test-key",
			},
			wantErr: false,
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

func TestNewGovUKAdapter(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceGovUK,
		Jurisdiction: "GB",
		Timeout:      30 * time.Second,
	}

	govUKConfig := GovUKConfig{
		Environment:     GovUKEnvironmentSandbox,
		ServiceEntityID: "test-service",
		APIKey:          "test-key",
	}

	adapter, err := NewGovUKAdapter(baseConfig, govUKConfig)
	require.NoError(t, err)
	require.NotNil(t, adapter)

	assert.Equal(t, DataSourceGovUK, adapter.Type())
	assert.Equal(t, "GB", adapter.Jurisdiction())
	assert.True(t, adapter.SupportsDocument(DocumentTypePassport))
	assert.True(t, adapter.SupportsDocument(DocumentTypeDriversLicense))
	assert.True(t, adapter.SupportsDocument(DocumentTypeNationalID))
}

func TestNewGovUKAdapter_InvalidConfig(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceGovUK,
		Jurisdiction: "GB",
	}

	govUKConfig := GovUKConfig{
		Environment: GovUKEnvironmentSandbox,
	}

	_, err := NewGovUKAdapter(baseConfig, govUKConfig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid GOV.UK Verify config")
}

func TestGovUKAdapter_BaseURL(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceGovUK,
		Jurisdiction: "GB",
	}

	tests := []struct {
		name        string
		environment string
		expectedURL string
	}{
		{"sandbox", GovUKEnvironmentSandbox, GovUKSandboxURL},
		{"integration", GovUKEnvironmentIntegration, GovUKIntegrationURL},
		{"production", GovUKEnvironmentProduction, GovUKProductionURL},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			govUKConfig := GovUKConfig{
				Environment:     tc.environment,
				ServiceEntityID: "test-service",
				APIKey:          "test-key",
			}
			adapter, _ := NewGovUKAdapter(baseConfig, govUKConfig)
			assert.Equal(t, tc.expectedURL, adapter.baseURL())
		})
	}
}

func TestGovUKAdapter_RateLimit(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceGovUK,
		Jurisdiction: "GB",
	}
	govUKConfig := GovUKConfig{
		Environment:        GovUKEnvironmentSandbox,
		ServiceEntityID:    "test-service",
		APIKey:             "test-key",
		RateLimitPerMinute: 3,
	}
	adapter, _ := NewGovUKAdapter(baseConfig, govUKConfig)

	// First 3 requests should succeed
	assert.NoError(t, adapter.checkRateLimit())
	assert.NoError(t, adapter.checkRateLimit())
	assert.NoError(t, adapter.checkRateLimit())

	// 4th request should be rate limited
	assert.ErrorIs(t, adapter.checkRateLimit(), ErrGovUKRateLimit)

	// Reset window
	adapter.mu.Lock()
	adapter.windowStart = time.Now().Add(-2 * time.Minute)
	adapter.mu.Unlock()

	// Should work again
	assert.NoError(t, adapter.checkRateLimit())
}

func TestGovUKAdapter_Verify_UnsupportedDocument(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceGovUK,
		Jurisdiction: "GB",
	}
	govUKConfig := GovUKConfig{
		Environment:     GovUKEnvironmentSandbox,
		ServiceEntityID: "test-service",
		APIKey:          "test-key",
	}
	adapter, _ := NewGovUKAdapter(baseConfig, govUKConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "GB",
		DocumentType: DocumentTypeTaxID, // Not supported
	}

	_, err := adapter.Verify(context.Background(), req)
	assert.ErrorIs(t, err, ErrDocumentTypeNotSupported)
}

func TestGovUKAdapter_BuildRequest(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceGovUK,
		Jurisdiction: "GB",
	}
	govUKConfig := GovUKConfig{
		Environment:      GovUKEnvironmentSandbox,
		ServiceEntityID:  "test-service",
		APIKey:           "test-key",
		LevelOfAssurance: GovUKLoA2,
	}
	adapter, _ := NewGovUKAdapter(baseConfig, govUKConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "GB",
		DocumentType: DocumentTypePassport,
		Fields: VerificationFields{
			FirstName:   "John",
			LastName:    "Doe",
			DateOfBirth: time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	govUKReq := adapter.buildRequest(req)

	assert.NotEmpty(t, govUKReq.RequestID)
	assert.Equal(t, GovUKLoA2, govUKReq.LevelOfAssurance)
	assert.True(t, govUKReq.Attributes.FirstName)
	assert.True(t, govUKReq.Attributes.Surname)
	assert.True(t, govUKReq.Attributes.DateOfBirth)
}

func TestGovUKAdapter_ConvertResponse_SuccessMatch(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceGovUK,
		Jurisdiction: "GB",
	}
	govUKConfig := GovUKConfig{
		Environment:     GovUKEnvironmentSandbox,
		ServiceEntityID: "test-service",
		APIKey:          "test-key",
	}
	adapter, _ := NewGovUKAdapter(baseConfig, govUKConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "GB",
	}

	govUKResp := &GovUKVerifyResponse{
		RequestID:        "test-req-123",
		Scenario:         GovUKScenarioSuccessMatch,
		LevelOfAssurance: GovUKLoA2,
		Attributes: &GovUKAttributes{
			FirstName: &GovUKVerifiedValue{Value: "John", Verified: true},
			Surname:   &GovUKVerifiedValue{Value: "Doe", Verified: true},
		},
	}

	resp := adapter.convertResponse(req, govUKResp)

	assert.Equal(t, "test-req-123", resp.RequestID)
	assert.Equal(t, VerificationStatusVerified, resp.Status)
	assert.True(t, resp.DocumentValid)
	assert.Equal(t, 1.0, resp.Confidence)
	assert.Equal(t, FieldMatchExact, resp.FieldResults["first_name"].Match)
	assert.Equal(t, FieldMatchExact, resp.FieldResults["last_name"].Match)
}

func TestGovUKAdapter_ConvertResponse_NoMatch(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceGovUK,
		Jurisdiction: "GB",
	}
	govUKConfig := GovUKConfig{
		Environment:     GovUKEnvironmentSandbox,
		ServiceEntityID: "test-service",
		APIKey:          "test-key",
	}
	adapter, _ := NewGovUKAdapter(baseConfig, govUKConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "GB",
	}

	govUKResp := &GovUKVerifyResponse{
		RequestID: "test-req-123",
		Scenario:  GovUKScenarioNoMatch,
	}

	resp := adapter.convertResponse(req, govUKResp)

	assert.Equal(t, VerificationStatusNotFound, resp.Status)
	assert.False(t, resp.DocumentValid)
	assert.Equal(t, 0.0, resp.Confidence)
}

func TestGovUKAdapter_ConvertResponse_Cancellation(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceGovUK,
		Jurisdiction: "GB",
	}
	govUKConfig := GovUKConfig{
		Environment:     GovUKEnvironmentSandbox,
		ServiceEntityID: "test-service",
		APIKey:          "test-key",
	}
	adapter, _ := NewGovUKAdapter(baseConfig, govUKConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "GB",
	}

	govUKResp := &GovUKVerifyResponse{
		RequestID: "test-req-123",
		Scenario:  GovUKScenarioCancellation,
	}

	resp := adapter.convertResponse(req, govUKResp)

	assert.Equal(t, VerificationStatusNotVerified, resp.Status)
	assert.Contains(t, resp.Warnings, "User cancelled verification")
}

func TestGovUKAdapter_BoolToMatch(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceGovUK,
		Jurisdiction: "GB",
	}
	govUKConfig := GovUKConfig{
		Environment:     GovUKEnvironmentSandbox,
		ServiceEntityID: "test-service",
		APIKey:          "test-key",
	}
	adapter, _ := NewGovUKAdapter(baseConfig, govUKConfig)

	assert.Equal(t, FieldMatchExact, adapter.boolToMatch(true))
	assert.Equal(t, FieldMatchNoMatch, adapter.boolToMatch(false))
	assert.Equal(t, 1.0, adapter.boolToConfidence(true))
	assert.Equal(t, 0.0, adapter.boolToConfidence(false))
}

func TestGovUKAdapter_GenerateRequestID(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceGovUK,
		Jurisdiction: "GB",
	}
	govUKConfig := GovUKConfig{
		Environment:     GovUKEnvironmentSandbox,
		ServiceEntityID: "test-service",
		APIKey:          "test-key",
	}
	adapter, _ := NewGovUKAdapter(baseConfig, govUKConfig)

	reqID := adapter.generateRequestID("test-123")
	assert.NotEmpty(t, reqID)
	assert.Len(t, reqID, 32)

	// Different inputs should produce different IDs
	reqID2 := adapter.generateRequestID("test-456")
	assert.NotEqual(t, reqID, reqID2)
}

func TestGovUKAdapter_WithMockServer(t *testing.T) {
	// Create mock GOV.UK server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/verify/match" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"requestId": "test-123",
				"scenario": "SUCCESS_MATCH",
				"pid": "pid-12345",
				"levelOfAssurance": "LOA_2",
				"attributes": {
					"firstName": {"value": "John", "verified": true},
					"surname": {"value": "Doe", "verified": true},
					"dateOfBirth": {"value": "1990-01-15", "verified": true}
				}
			}`))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	baseConfig := AdapterConfig{
		Type:         DataSourceGovUK,
		Jurisdiction: "GB",
		Timeout:      5 * time.Second,
	}
	govUKConfig := GovUKConfig{
		Environment:        GovUKEnvironmentSandbox,
		ServiceEntityID:    "test-service",
		APIKey:             "test-key",
		Timeout:            5 * time.Second,
		BaseURL:            server.URL,
		RateLimitPerMinute: 60,
	}

	adapter, err := NewGovUKAdapter(baseConfig, govUKConfig)
	require.NoError(t, err)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "GB",
		DocumentType: DocumentTypePassport,
		Fields: VerificationFields{
			FirstName:   "John",
			LastName:    "Doe",
			DateOfBirth: time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	resp, err := adapter.Verify(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, VerificationStatusVerified, resp.Status)
	assert.True(t, resp.DocumentValid)
	assert.Equal(t, 1.0, resp.Confidence)
}

