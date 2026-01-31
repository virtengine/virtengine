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
// eIDAS Adapter Tests
// ============================================================================

func TestDefaultEIDASConfig(t *testing.T) {
	config := DefaultEIDASConfig()

	assert.Equal(t, EIDASEnvironmentSandbox, config.Environment)
	assert.Equal(t, 60*time.Second, config.Timeout)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 30, config.RateLimitPerMinute)
	assert.Equal(t, EIDASLoASubstantial, config.RequiredLoA)
	assert.True(t, config.AuditEnabled)
	assert.True(t, config.GDPRCompliant)
	assert.True(t, len(config.SupportedCountries) > 0)
	assert.Contains(t, config.SupportedCountries, "DE")
	assert.Contains(t, config.SupportedCountries, "FR")
}

func TestEIDASConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  EIDASConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: EIDASConfig{
				Environment:            EIDASEnvironmentSandbox,
				ServiceProviderID:      "test-sp",
				ServiceProviderCountry: "DE",
				APIKey:                 "test-key",
			},
			wantErr: false,
		},
		{
			name: "missing service_provider_id",
			config: EIDASConfig{
				Environment:            EIDASEnvironmentSandbox,
				ServiceProviderCountry: "DE",
				APIKey:                 "test-key",
			},
			wantErr: true,
		},
		{
			name: "missing service_provider_country",
			config: EIDASConfig{
				Environment:       EIDASEnvironmentSandbox,
				ServiceProviderID: "test-sp",
				APIKey:            "test-key",
			},
			wantErr: true,
		},
		{
			name: "missing api_key",
			config: EIDASConfig{
				Environment:            EIDASEnvironmentSandbox,
				ServiceProviderID:      "test-sp",
				ServiceProviderCountry: "DE",
			},
			wantErr: true,
		},
		{
			name: "invalid environment",
			config: EIDASConfig{
				Environment:            "invalid",
				ServiceProviderID:      "test-sp",
				ServiceProviderCountry: "DE",
				APIKey:                 "test-key",
			},
			wantErr: true,
		},
		{
			name: "test environment",
			config: EIDASConfig{
				Environment:            EIDASEnvironmentTest,
				ServiceProviderID:      "test-sp",
				ServiceProviderCountry: "DE",
				APIKey:                 "test-key",
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

func TestNewEIDASAdapter(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceEIDAS,
		Jurisdiction: "EU-DE",
		Timeout:      60 * time.Second,
	}

	eidasConfig := EIDASConfig{
		Environment:            EIDASEnvironmentSandbox,
		ServiceProviderID:      "test-sp",
		ServiceProviderCountry: "DE",
		APIKey:                 "test-key",
	}

	adapter, err := NewEIDASAdapter(baseConfig, eidasConfig)
	require.NoError(t, err)
	require.NotNil(t, adapter)

	assert.Equal(t, DataSourceEIDAS, adapter.Type())
	assert.Equal(t, "EU-DE", adapter.Jurisdiction())
	assert.True(t, adapter.SupportsDocument(DocumentTypeNationalID))
	assert.True(t, adapter.SupportsDocument(DocumentTypePassport))
	assert.True(t, adapter.SupportsDocument(DocumentTypeResidencePermit))
}

func TestNewEIDASAdapter_InvalidConfig(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceEIDAS,
		Jurisdiction: "EU-DE",
	}

	eidasConfig := EIDASConfig{
		Environment: EIDASEnvironmentSandbox,
	}

	_, err := NewEIDASAdapter(baseConfig, eidasConfig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid eIDAS config")
}

func TestEIDASAdapter_BaseURL(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceEIDAS,
		Jurisdiction: "EU-DE",
	}

	tests := []struct {
		name        string
		environment string
		expectedURL string
	}{
		{"sandbox", EIDASEnvironmentSandbox, EIDASSandboxURL},
		{"test", EIDASEnvironmentTest, EIDASTestURL},
		{"production", EIDASEnvironmentProduction, EIDASProductionURL},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			eidasConfig := EIDASConfig{
				Environment:            tc.environment,
				ServiceProviderID:      "test-sp",
				ServiceProviderCountry: "DE",
				APIKey:                 "test-key",
			}
			adapter, _ := NewEIDASAdapter(baseConfig, eidasConfig)
			assert.Equal(t, tc.expectedURL, adapter.baseURL())
		})
	}
}

func TestEIDASAdapter_ExtractCountry(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceEIDAS,
		Jurisdiction: "EU-DE",
	}
	eidasConfig := EIDASConfig{
		Environment:            EIDASEnvironmentSandbox,
		ServiceProviderID:      "test-sp",
		ServiceProviderCountry: "DE",
		APIKey:                 "test-key",
	}
	adapter, _ := NewEIDASAdapter(baseConfig, eidasConfig)

	tests := []struct {
		jurisdiction string
		expected     string
	}{
		{"EU-DE", "DE"},
		{"EU-FR", "FR"},
		{"EU/IT", "IT"},
		{"DE", "DE"},
		{"fr", "FR"},
		{"eu-es", "ES"},
	}

	for _, tc := range tests {
		t.Run(tc.jurisdiction, func(t *testing.T) {
			result := adapter.extractCountry(tc.jurisdiction)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestEIDASAdapter_IsCountrySupported(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceEIDAS,
		Jurisdiction: "EU-DE",
	}
	eidasConfig := EIDASConfig{
		Environment:            EIDASEnvironmentSandbox,
		ServiceProviderID:      "test-sp",
		ServiceProviderCountry: "DE",
		APIKey:                 "test-key",
		SupportedCountries:     []string{"DE", "FR", "IT", "ES"},
	}
	adapter, _ := NewEIDASAdapter(baseConfig, eidasConfig)

	assert.True(t, adapter.isCountrySupported("DE"))
	assert.True(t, adapter.isCountrySupported("de")) // Case insensitive
	assert.True(t, adapter.isCountrySupported("FR"))
	assert.True(t, adapter.isCountrySupported("IT"))
	assert.True(t, adapter.isCountrySupported("ES"))
	assert.False(t, adapter.isCountrySupported("GB"))
	assert.False(t, adapter.isCountrySupported("XX"))
}

func TestEIDASAdapter_RateLimit(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceEIDAS,
		Jurisdiction: "EU-DE",
	}
	eidasConfig := EIDASConfig{
		Environment:            EIDASEnvironmentSandbox,
		ServiceProviderID:      "test-sp",
		ServiceProviderCountry: "DE",
		APIKey:                 "test-key",
		RateLimitPerMinute:     3,
	}
	adapter, _ := NewEIDASAdapter(baseConfig, eidasConfig)

	// First 3 requests should succeed
	assert.NoError(t, adapter.checkRateLimit())
	assert.NoError(t, adapter.checkRateLimit())
	assert.NoError(t, adapter.checkRateLimit())

	// 4th request should be rate limited
	assert.ErrorIs(t, adapter.checkRateLimit(), ErrEIDASRateLimit)

	// Reset window
	adapter.mu.Lock()
	adapter.windowStart = time.Now().Add(-2 * time.Minute)
	adapter.mu.Unlock()

	// Should work again
	assert.NoError(t, adapter.checkRateLimit())
}

func TestEIDASAdapter_Verify_UnsupportedDocument(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceEIDAS,
		Jurisdiction: "EU-DE",
	}
	eidasConfig := EIDASConfig{
		Environment:            EIDASEnvironmentSandbox,
		ServiceProviderID:      "test-sp",
		ServiceProviderCountry: "DE",
		APIKey:                 "test-key",
	}
	adapter, _ := NewEIDASAdapter(baseConfig, eidasConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "EU-DE",
		DocumentType: DocumentTypeTaxID, // Not supported
	}

	_, err := adapter.Verify(context.Background(), req)
	assert.ErrorIs(t, err, ErrDocumentTypeNotSupported)
}

func TestEIDASAdapter_Verify_UnsupportedCountry(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceEIDAS,
		Jurisdiction: "EU-DE",
	}
	eidasConfig := EIDASConfig{
		Environment:            EIDASEnvironmentSandbox,
		ServiceProviderID:      "test-sp",
		ServiceProviderCountry: "DE",
		APIKey:                 "test-key",
		SupportedCountries:     []string{"DE", "FR"},
	}
	adapter, _ := NewEIDASAdapter(baseConfig, eidasConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "EU-IT", // Not in supported list
		DocumentType: DocumentTypeNationalID,
		Fields: VerificationFields{
			DocumentNumber: "12345678",
		},
	}

	_, err := adapter.Verify(context.Background(), req)
	assert.ErrorIs(t, err, ErrEIDASCountryNotSupported)
}

func TestEIDASAdapter_BuildRequest(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceEIDAS,
		Jurisdiction: "EU-DE",
	}
	eidasConfig := EIDASConfig{
		Environment:            EIDASEnvironmentSandbox,
		ServiceProviderID:      "test-sp",
		ServiceProviderCountry: "DE",
		APIKey:                 "test-key",
		RequiredLoA:            EIDASLoAHigh,
		RequestedAttributes:    []string{EIDASAttrPersonIdentifier, EIDASAttrFamilyName},
	}
	adapter, _ := NewEIDASAdapter(baseConfig, eidasConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "EU-FR",
		DocumentType: DocumentTypeNationalID,
	}

	eidasReq := adapter.buildRequest(req, "FR")

	assert.NotEmpty(t, eidasReq.RequestID)
	assert.True(t, len(eidasReq.RequestID) > 0 && eidasReq.RequestID[0] == '_')
	assert.Equal(t, "FR", eidasReq.DestinationCountry)
	assert.Equal(t, "FR", eidasReq.CitizenCountry)
	assert.Equal(t, "DE", eidasReq.SPCountry)
	assert.Equal(t, EIDASLoAHigh, eidasReq.LevelOfAssurance)
}

func TestEIDASAdapter_LoAToConfidence(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceEIDAS,
		Jurisdiction: "EU-DE",
	}
	eidasConfig := EIDASConfig{
		Environment:            EIDASEnvironmentSandbox,
		ServiceProviderID:      "test-sp",
		ServiceProviderCountry: "DE",
		APIKey:                 "test-key",
	}
	adapter, _ := NewEIDASAdapter(baseConfig, eidasConfig)

	assert.Equal(t, 1.0, adapter.loaToConfidence(EIDASLoAHigh))
	assert.Equal(t, 0.9, adapter.loaToConfidence(EIDASLoASubstantial))
	assert.Equal(t, 0.7, adapter.loaToConfidence(EIDASLoALow))
	assert.Equal(t, 0.5, adapter.loaToConfidence("unknown"))
}

func TestEIDASAdapter_AttrToFieldName(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceEIDAS,
		Jurisdiction: "EU-DE",
	}
	eidasConfig := EIDASConfig{
		Environment:            EIDASEnvironmentSandbox,
		ServiceProviderID:      "test-sp",
		ServiceProviderCountry: "DE",
		APIKey:                 "test-key",
	}
	adapter, _ := NewEIDASAdapter(baseConfig, eidasConfig)

	assert.Equal(t, "document_number", adapter.attrToFieldName(EIDASAttrPersonIdentifier))
	assert.Equal(t, "last_name", adapter.attrToFieldName(EIDASAttrFamilyName))
	assert.Equal(t, "first_name", adapter.attrToFieldName(EIDASAttrFirstName))
	assert.Equal(t, "date_of_birth", adapter.attrToFieldName(EIDASAttrDateOfBirth))
	assert.Equal(t, "nationality", adapter.attrToFieldName(EIDASAttrNationality))
	assert.Equal(t, "", adapter.attrToFieldName("unknown"))
}

func TestEIDASAdapter_ConvertResponse_Success(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceEIDAS,
		Jurisdiction: "EU-DE",
	}
	eidasConfig := EIDASConfig{
		Environment:            EIDASEnvironmentSandbox,
		ServiceProviderID:      "test-sp",
		ServiceProviderCountry: "DE",
		APIKey:                 "test-key",
	}
	adapter, _ := NewEIDASAdapter(baseConfig, eidasConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "EU-DE",
	}

	eidasResp := &EIDASAuthResponse{
		RequestID:        "test-req-123",
		Status:           "SUCCESS",
		StatusCode:       EIDASStatusSuccess,
		LevelOfAssurance: EIDASLoAHigh,
		Attributes: []EIDASAttribute{
			{Name: EIDASAttrFamilyName, Values: []string{"Doe"}},
			{Name: EIDASAttrFirstName, Values: []string{"John"}},
		},
	}

	resp := adapter.convertResponse(req, eidasResp)

	assert.Equal(t, "test-req-123", resp.RequestID)
	assert.Equal(t, VerificationStatusVerified, resp.Status)
	assert.True(t, resp.DocumentValid)
	assert.Equal(t, 1.0, resp.Confidence)
	assert.Equal(t, FieldMatchExact, resp.FieldResults["last_name"].Match)
	assert.Equal(t, FieldMatchExact, resp.FieldResults["first_name"].Match)
}

func TestEIDASAdapter_ConvertResponse_AuthnFailed(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceEIDAS,
		Jurisdiction: "EU-DE",
	}
	eidasConfig := EIDASConfig{
		Environment:            EIDASEnvironmentSandbox,
		ServiceProviderID:      "test-sp",
		ServiceProviderCountry: "DE",
		APIKey:                 "test-key",
	}
	adapter, _ := NewEIDASAdapter(baseConfig, eidasConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "EU-DE",
	}

	eidasResp := &EIDASAuthResponse{
		RequestID:  "test-req-123",
		Status:     "FAILED",
		StatusCode: EIDASStatusAuthnFailed,
	}

	resp := adapter.convertResponse(req, eidasResp)

	assert.Equal(t, VerificationStatusNotVerified, resp.Status)
	assert.False(t, resp.DocumentValid)
	assert.Contains(t, resp.Warnings, "Authentication failed")
}

func TestEIDASAdapter_GenerateRequestID(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourceEIDAS,
		Jurisdiction: "EU-DE",
	}
	eidasConfig := EIDASConfig{
		Environment:            EIDASEnvironmentSandbox,
		ServiceProviderID:      "test-sp",
		ServiceProviderCountry: "DE",
		APIKey:                 "test-key",
	}
	adapter, _ := NewEIDASAdapter(baseConfig, eidasConfig)

	reqID := adapter.generateRequestID("test-123")
	assert.NotEmpty(t, reqID)
	assert.Equal(t, '_', rune(reqID[0])) // eIDAS IDs must start with _
	assert.Len(t, reqID, 32)

	// Different inputs should produce different IDs
	reqID2 := adapter.generateRequestID("test-456")
	assert.NotEqual(t, reqID, reqID2)
}

func TestEIDASAdapter_WithMockServer(t *testing.T) {
	// Create mock eIDAS server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ServiceProvider/authenticate" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"requestId": "test-123",
				"status": "SUCCESS",
				"statusCode": "urn:oasis:names:tc:SAML:2.0:status:Success",
				"levelOfAssurance": "http://eidas.europa.eu/LoA/high",
				"attributes": [
					{"name": "http://eidas.europa.eu/attributes/naturalperson/CurrentFamilyName", "values": ["Doe"]},
					{"name": "http://eidas.europa.eu/attributes/naturalperson/CurrentGivenName", "values": ["John"]}
				]
			}`))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	baseConfig := AdapterConfig{
		Type:         DataSourceEIDAS,
		Jurisdiction: "EU-DE",
		Timeout:      5 * time.Second,
	}
	eidasConfig := EIDASConfig{
		Environment:            EIDASEnvironmentSandbox,
		ServiceProviderID:      "test-sp",
		ServiceProviderCountry: "DE",
		APIKey:                 "test-key",
		Timeout:                5 * time.Second,
		BaseURL:                server.URL,
		SupportedCountries:     []string{"DE"},
		RateLimitPerMinute:     30,
	}

	adapter, err := NewEIDASAdapter(baseConfig, eidasConfig)
	require.NoError(t, err)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "EU-DE",
		DocumentType: DocumentTypeNationalID,
		Fields: VerificationFields{
			DocumentNumber: "12345678",
			FirstName:      "John",
			LastName:       "Doe",
		},
	}

	resp, err := adapter.Verify(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, VerificationStatusVerified, resp.Status)
	assert.True(t, resp.DocumentValid)
	assert.Equal(t, 1.0, resp.Confidence)
}

func TestEIDASMemberStates(t *testing.T) {
	// Verify all expected EU member states are included
	expectedStates := []string{
		"AT", "BE", "BG", "HR", "CY", "CZ", "DK", "EE", "FI", "FR",
		"DE", "GR", "HU", "IE", "IT", "LV", "LT", "LU", "MT", "NL",
		"PL", "PT", "RO", "SK", "SI", "ES", "SE",
	}

	for _, state := range expectedStates {
		assert.Contains(t, EIDASMemberStates, state, "Missing EU member state: %s", state)
	}

	// EEA countries
	assert.Contains(t, EIDASMemberStates, "IS") // Iceland
	assert.Contains(t, EIDASMemberStates, "LI") // Liechtenstein
	assert.Contains(t, EIDASMemberStates, "NO") // Norway
}
