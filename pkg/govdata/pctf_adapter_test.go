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
// PCTF Adapter Tests
// ============================================================================

func TestDefaultPCTFConfig(t *testing.T) {
	config := DefaultPCTFConfig()

	assert.Equal(t, PCTFEnvironmentSandbox, config.Environment)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 60, config.RateLimitPerMinute)
	assert.Equal(t, PCTFAssuranceLevelStandard, config.RequiredAssuranceLevel)
	assert.Equal(t, "en", config.PreferredLanguage)
	assert.True(t, config.AuditEnabled)
	assert.True(t, config.PIPEDACompliant)
	assert.True(t, len(config.SupportedProvinces) > 0)
	assert.Contains(t, config.SupportedProvinces, "ON")
	assert.Contains(t, config.SupportedProvinces, "BC")
	assert.Contains(t, config.SupportedProvinces, "QC")
}

func TestPCTFConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  PCTFConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: PCTFConfig{
				Environment:    PCTFEnvironmentSandbox,
				OrganizationID: "test-org",
				ClientID:       "test-client",
				ClientSecret:   "test-secret",
			},
			wantErr: false,
		},
		{
			name: "missing organization_id",
			config: PCTFConfig{
				Environment:  PCTFEnvironmentSandbox,
				ClientID:     "test-client",
				ClientSecret: "test-secret",
			},
			wantErr: true,
		},
		{
			name: "missing client_id",
			config: PCTFConfig{
				Environment:    PCTFEnvironmentSandbox,
				OrganizationID: "test-org",
				ClientSecret:   "test-secret",
			},
			wantErr: true,
		},
		{
			name: "missing client_secret",
			config: PCTFConfig{
				Environment:    PCTFEnvironmentSandbox,
				OrganizationID: "test-org",
				ClientID:       "test-client",
			},
			wantErr: true,
		},
		{
			name: "invalid environment",
			config: PCTFConfig{
				Environment:    "invalid",
				OrganizationID: "test-org",
				ClientID:       "test-client",
				ClientSecret:   "test-secret",
			},
			wantErr: true,
		},
		{
			name: "test environment",
			config: PCTFConfig{
				Environment:    PCTFEnvironmentTest,
				OrganizationID: "test-org",
				ClientID:       "test-client",
				ClientSecret:   "test-secret",
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

func TestNewPCTFAdapter(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourcePCTF,
		Jurisdiction: "CA-ON",
		Timeout:      30 * time.Second,
	}

	pctfConfig := PCTFConfig{
		Environment:    PCTFEnvironmentSandbox,
		OrganizationID: "test-org",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}

	adapter, err := NewPCTFAdapter(baseConfig, pctfConfig)
	require.NoError(t, err)
	require.NotNil(t, adapter)

	assert.Equal(t, DataSourcePCTF, adapter.Type())
	assert.Equal(t, "CA-ON", adapter.Jurisdiction())
	assert.True(t, adapter.SupportsDocument(DocumentTypeDriversLicense))
	assert.True(t, adapter.SupportsDocument(DocumentTypePassport))
	assert.True(t, adapter.SupportsDocument(DocumentTypeBirthCertificate))
	assert.True(t, adapter.SupportsDocument(DocumentTypeNationalID))
	assert.True(t, adapter.SupportsDocument(DocumentTypeResidencePermit))
}

func TestNewPCTFAdapter_InvalidConfig(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourcePCTF,
		Jurisdiction: "CA-ON",
	}

	pctfConfig := PCTFConfig{
		Environment: PCTFEnvironmentSandbox,
	}

	_, err := NewPCTFAdapter(baseConfig, pctfConfig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid PCTF config")
}

func TestPCTFAdapter_BaseURL(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourcePCTF,
		Jurisdiction: "CA-ON",
	}

	tests := []struct {
		name        string
		environment string
		expectedURL string
	}{
		{"sandbox", PCTFEnvironmentSandbox, PCTFSandboxURL},
		{"test", PCTFEnvironmentTest, PCTFTestURL},
		{"production", PCTFEnvironmentProduction, PCTFProductionURL},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pctfConfig := PCTFConfig{
				Environment:    tc.environment,
				OrganizationID: "test-org",
				ClientID:       "test-client",
				ClientSecret:   "test-secret",
			}
			adapter, _ := NewPCTFAdapter(baseConfig, pctfConfig)
			assert.Equal(t, tc.expectedURL, adapter.baseURL())
		})
	}
}

func TestPCTFAdapter_ExtractProvince(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourcePCTF,
		Jurisdiction: "CA-ON",
	}
	pctfConfig := PCTFConfig{
		Environment:    PCTFEnvironmentSandbox,
		OrganizationID: "test-org",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}
	adapter, _ := NewPCTFAdapter(baseConfig, pctfConfig)

	tests := []struct {
		jurisdiction string
		expected     string
	}{
		{"CA-ON", "ON"},
		{"CA-BC", "BC"},
		{"CA/QC", "QC"},
		{"ON", "ON"},
		{"ca-ab", "AB"},
	}

	for _, tc := range tests {
		t.Run(tc.jurisdiction, func(t *testing.T) {
			result := adapter.extractProvince(tc.jurisdiction)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestPCTFAdapter_IsProvinceSupported(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourcePCTF,
		Jurisdiction: "CA-ON",
	}
	pctfConfig := PCTFConfig{
		Environment:        PCTFEnvironmentSandbox,
		OrganizationID:     "test-org",
		ClientID:           "test-client",
		ClientSecret:       "test-secret",
		SupportedProvinces: []string{"ON", "BC", "QC", "AB"},
	}
	adapter, _ := NewPCTFAdapter(baseConfig, pctfConfig)

	assert.True(t, adapter.isProvinceSupported("ON"))
	assert.True(t, adapter.isProvinceSupported("on")) // Case insensitive
	assert.True(t, adapter.isProvinceSupported("BC"))
	assert.True(t, adapter.isProvinceSupported("QC"))
	assert.True(t, adapter.isProvinceSupported("AB"))
	assert.False(t, adapter.isProvinceSupported("MB"))
	assert.False(t, adapter.isProvinceSupported("XX"))
}

func TestPCTFAdapter_RateLimit(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourcePCTF,
		Jurisdiction: "CA-ON",
	}
	pctfConfig := PCTFConfig{
		Environment:        PCTFEnvironmentSandbox,
		OrganizationID:     "test-org",
		ClientID:           "test-client",
		ClientSecret:       "test-secret",
		RateLimitPerMinute: 3,
	}
	adapter, _ := NewPCTFAdapter(baseConfig, pctfConfig)

	// First 3 requests should succeed
	assert.NoError(t, adapter.checkRateLimit())
	assert.NoError(t, adapter.checkRateLimit())
	assert.NoError(t, adapter.checkRateLimit())

	// 4th request should be rate limited
	assert.ErrorIs(t, adapter.checkRateLimit(), ErrPCTFRateLimit)

	// Reset window
	adapter.mu.Lock()
	adapter.windowStart = time.Now().Add(-2 * time.Minute)
	adapter.mu.Unlock()

	// Should work again
	assert.NoError(t, adapter.checkRateLimit())
}

func TestPCTFAdapter_Verify_UnsupportedDocument(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourcePCTF,
		Jurisdiction: "CA-ON",
	}
	pctfConfig := PCTFConfig{
		Environment:    PCTFEnvironmentSandbox,
		OrganizationID: "test-org",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}
	adapter, _ := NewPCTFAdapter(baseConfig, pctfConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "CA-ON",
		DocumentType: DocumentTypeTaxID, // Not supported
	}

	_, err := adapter.Verify(context.Background(), req)
	assert.ErrorIs(t, err, ErrDocumentTypeNotSupported)
}

func TestPCTFAdapter_Verify_UnsupportedProvince(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourcePCTF,
		Jurisdiction: "CA-ON",
	}
	pctfConfig := PCTFConfig{
		Environment:        PCTFEnvironmentSandbox,
		OrganizationID:     "test-org",
		ClientID:           "test-client",
		ClientSecret:       "test-secret",
		SupportedProvinces: []string{"ON", "BC"},
	}
	adapter, _ := NewPCTFAdapter(baseConfig, pctfConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "CA-QC", // Not in supported list
		DocumentType: DocumentTypeDriversLicense,
		Fields: VerificationFields{
			DocumentNumber: "12345678",
		},
	}

	_, err := adapter.Verify(context.Background(), req)
	assert.ErrorIs(t, err, ErrPCTFProvinceNotSupported)
}

func TestPCTFAdapter_BuildRequest(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourcePCTF,
		Jurisdiction: "CA-ON",
	}
	pctfConfig := PCTFConfig{
		Environment:            PCTFEnvironmentSandbox,
		OrganizationID:         "test-org",
		ClientID:               "test-client",
		ClientSecret:           "test-secret",
		RequiredAssuranceLevel: PCTFAssuranceLevelEnhanced,
		PreferredLanguage:      "fr",
	}
	adapter, _ := NewPCTFAdapter(baseConfig, pctfConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "CA-ON",
		DocumentType: DocumentTypeDriversLicense,
		ConsentID:    "consent-456",
		Fields: VerificationFields{
			DocumentNumber: "A1234567",
			FirstName:      "John",
			LastName:       "Doe",
			DateOfBirth:    time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	pctfReq, err := adapter.buildRequest(req, "ON")
	require.NoError(t, err)

	assert.NotEmpty(t, pctfReq.RequestID)
	assert.Equal(t, PCTFDocDriverLicence, pctfReq.DocumentType)
	assert.Equal(t, "ON", pctfReq.Province)
	assert.Equal(t, PCTFAssuranceLevelEnhanced, pctfReq.AssuranceLevel)
	assert.Equal(t, "fr", pctfReq.PreferredLanguage)
	assert.Equal(t, "consent-456", pctfReq.ConsentReference)
	assert.Equal(t, "John", pctfReq.PersonData.GivenName)
	assert.Equal(t, "Doe", pctfReq.PersonData.Surname)
	assert.Equal(t, "1990-01-15", pctfReq.PersonData.DateOfBirth)
	assert.Equal(t, "A1234567", pctfReq.DocumentData.DocumentNumber)
}

func TestPCTFAdapter_BuildRequest_AllDocTypes(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourcePCTF,
		Jurisdiction: "CA-ON",
	}
	pctfConfig := PCTFConfig{
		Environment:    PCTFEnvironmentSandbox,
		OrganizationID: "test-org",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}
	adapter, _ := NewPCTFAdapter(baseConfig, pctfConfig)

	tests := []struct {
		docType         DocumentType
		expectedPCTFDoc string
	}{
		{DocumentTypeDriversLicense, PCTFDocDriverLicence},
		{DocumentTypePassport, PCTFDocPassport},
		{DocumentTypeBirthCertificate, PCTFDocBirthCertificate},
		{DocumentTypeResidencePermit, PCTFDocPRCard},
		{DocumentTypeNationalID, PCTFDocCitizenshipCard},
	}

	for _, tc := range tests {
		t.Run(string(tc.docType), func(t *testing.T) {
			req := &VerificationRequest{
				RequestID:    "test-req-123",
				DocumentType: tc.docType,
				Fields: VerificationFields{
					DocumentNumber: "12345678",
				},
			}
			pctfReq, err := adapter.buildRequest(req, "ON")
			require.NoError(t, err)
			assert.Equal(t, tc.expectedPCTFDoc, pctfReq.DocumentType)
		})
	}
}

func TestPCTFAdapter_ConvertResponse_Success(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourcePCTF,
		Jurisdiction: "CA-ON",
	}
	pctfConfig := PCTFConfig{
		Environment:    PCTFEnvironmentSandbox,
		OrganizationID: "test-org",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}
	adapter, _ := NewPCTFAdapter(baseConfig, pctfConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "CA-ON",
	}

	pctfResp := &PCTFVerifyResponse{
		RequestID:     "test-req-123",
		TransactionID: "txn-456",
		Status:        PCTFStatusSuccess,
		VerificationResult: PCTFVerificationResult{
			OverallMatch:      PCTFMatchYes,
			IdentityVerified:  true,
			DocumentVerified:  true,
			Confidence:        0.95,
			AssuranceLevelMet: true,
			FieldResults: map[string]PCTFFieldResult{
				"given_name": {FieldName: "given_name", Match: PCTFMatchYes, Confidence: 1.0},
				"surname":    {FieldName: "surname", Match: PCTFMatchYes, Confidence: 1.0},
			},
		},
	}

	resp := adapter.convertResponse(req, pctfResp)

	assert.Equal(t, "test-req-123", resp.RequestID)
	assert.Equal(t, VerificationStatusVerified, resp.Status)
	assert.True(t, resp.DocumentValid)
	assert.Equal(t, 0.95, resp.Confidence)
	assert.Equal(t, FieldMatchExact, resp.FieldResults["given_name"].Match)
	assert.Equal(t, FieldMatchExact, resp.FieldResults["surname"].Match)
}

func TestPCTFAdapter_ConvertResponse_PartialMatch(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourcePCTF,
		Jurisdiction: "CA-ON",
	}
	pctfConfig := PCTFConfig{
		Environment:    PCTFEnvironmentSandbox,
		OrganizationID: "test-org",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}
	adapter, _ := NewPCTFAdapter(baseConfig, pctfConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "CA-ON",
	}

	pctfResp := &PCTFVerifyResponse{
		RequestID: "test-req-123",
		Status:    PCTFStatusPartialMatch,
		VerificationResult: PCTFVerificationResult{
			OverallMatch:     PCTFMatchPartial,
			IdentityVerified: true,
			DocumentVerified: true,
			Confidence:       0.7,
		},
	}

	resp := adapter.convertResponse(req, pctfResp)

	assert.Equal(t, VerificationStatusPartialMatch, resp.Status)
	assert.True(t, resp.DocumentValid)
	assert.Equal(t, 0.7, resp.Confidence)
}

func TestPCTFAdapter_ConvertResponse_Expired(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourcePCTF,
		Jurisdiction: "CA-ON",
	}
	pctfConfig := PCTFConfig{
		Environment:    PCTFEnvironmentSandbox,
		OrganizationID: "test-org",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}
	adapter, _ := NewPCTFAdapter(baseConfig, pctfConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "CA-ON",
	}

	pctfResp := &PCTFVerifyResponse{
		RequestID: "test-req-123",
		Status:    PCTFStatusExpired,
		VerificationResult: PCTFVerificationResult{
			OverallMatch:     PCTFMatchYes,
			IdentityVerified: true,
			DocumentVerified: false,
			Confidence:       0.9,
		},
	}

	resp := adapter.convertResponse(req, pctfResp)

	assert.Equal(t, VerificationStatusExpired, resp.Status)
	assert.False(t, resp.DocumentValid)
	assert.Contains(t, resp.Warnings, "Document has expired")
}

func TestPCTFAdapter_ConvertResponse_Revoked(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourcePCTF,
		Jurisdiction: "CA-ON",
	}
	pctfConfig := PCTFConfig{
		Environment:    PCTFEnvironmentSandbox,
		OrganizationID: "test-org",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}
	adapter, _ := NewPCTFAdapter(baseConfig, pctfConfig)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "CA-ON",
	}

	pctfResp := &PCTFVerifyResponse{
		RequestID: "test-req-123",
		Status:    PCTFStatusRevoked,
	}

	resp := adapter.convertResponse(req, pctfResp)

	assert.Equal(t, VerificationStatusRevoked, resp.Status)
	assert.False(t, resp.DocumentValid)
	assert.Contains(t, resp.Warnings, "Document has been revoked")
}

func TestPCTFAdapter_ConvertFieldResult(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourcePCTF,
		Jurisdiction: "CA-ON",
	}
	pctfConfig := PCTFConfig{
		Environment:    PCTFEnvironmentSandbox,
		OrganizationID: "test-org",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}
	adapter, _ := NewPCTFAdapter(baseConfig, pctfConfig)

	tests := []struct {
		pctfMatch     string
		expectedMatch FieldMatchResult
	}{
		{PCTFMatchYes, FieldMatchExact},
		{PCTFMatchPartial, FieldMatchFuzzy},
		{PCTFMatchNo, FieldMatchNoMatch},
		{PCTFMatchNA, FieldMatchUnavailable},
		{"unknown", FieldMatchUnavailable},
	}

	for _, tc := range tests {
		t.Run(tc.pctfMatch, func(t *testing.T) {
			pctfResult := PCTFFieldResult{
				FieldName:  "test_field",
				Match:      tc.pctfMatch,
				Confidence: 0.9,
			}
			result := adapter.convertFieldResult(pctfResult)
			assert.Equal(t, tc.expectedMatch, result.Match)
		})
	}
}

func TestPCTFAdapter_GenerateRequestID(t *testing.T) {
	baseConfig := AdapterConfig{
		Type:         DataSourcePCTF,
		Jurisdiction: "CA-ON",
	}
	pctfConfig := PCTFConfig{
		Environment:    PCTFEnvironmentSandbox,
		OrganizationID: "test-org",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
	}
	adapter, _ := NewPCTFAdapter(baseConfig, pctfConfig)

	reqID := adapter.generateRequestID("test-123")
	assert.NotEmpty(t, reqID)
	assert.Len(t, reqID, 32)

	// Different inputs should produce different IDs
	reqID2 := adapter.generateRequestID("test-456")
	assert.NotEqual(t, reqID, reqID2)
}

func TestPCTFAdapter_WithMockServer(t *testing.T) {
	// Create mock PCTF server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"access_token": "test-token", "expires_in": 3600, "token_type": "Bearer"}`))
			return
		}

		if r.URL.Path == "/verify/identity" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"requestId": "test-123",
				"transactionId": "txn-456",
				"status": "SUCCESS",
				"verificationResult": {
					"overallMatch": "Y",
					"identityVerified": true,
					"documentVerified": true,
					"confidence": 0.95,
					"assuranceLevelMet": true,
					"fieldResults": {
						"given_name": {"fieldName": "given_name", "match": "Y", "confidence": 1.0},
						"surname": {"fieldName": "surname", "match": "Y", "confidence": 1.0}
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
		Type:         DataSourcePCTF,
		Jurisdiction: "CA-ON",
		Timeout:      5 * time.Second,
	}
	pctfConfig := PCTFConfig{
		Environment:        PCTFEnvironmentSandbox,
		OrganizationID:     "test-org",
		ClientID:           "test-client",
		ClientSecret:       "test-secret",
		Timeout:            5 * time.Second,
		BaseURL:            server.URL,
		SupportedProvinces: []string{"ON"},
		RateLimitPerMinute: 60,
	}

	adapter, err := NewPCTFAdapter(baseConfig, pctfConfig)
	require.NoError(t, err)

	req := &VerificationRequest{
		RequestID:    "test-req-123",
		Jurisdiction: "CA-ON",
		DocumentType: DocumentTypeDriversLicense,
		ConsentID:    "consent-456",
		Fields: VerificationFields{
			DocumentNumber: "A1234567",
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
	assert.Equal(t, 0.95, resp.Confidence)
}

func TestPCTFProvinceTerritories(t *testing.T) {
	// Verify all expected Canadian provinces/territories are included
	expectedPT := []string{
		"AB", "BC", "MB", "NB", "NL", "NS", "NT", "NU", "ON", "PE", "QC", "SK", "YT",
	}

	for _, pt := range expectedPT {
		assert.Contains(t, PCTFProvinceTerritories, pt, "Missing province/territory: %s", pt)
	}

	// Verify count
	assert.Len(t, PCTFProvinceTerritories, 13)
}
