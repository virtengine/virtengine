// Package govdata provides government data source integration for identity verification.
//
// VE-909: Government data integration for authoritative identity verification
package govdata

import (
	"context"
	"testing"
	"time"
)

// ============================================================================
// Type Tests
// ============================================================================

func TestDocumentType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		docType  DocumentType
		expected bool
	}{
		{"drivers_license", DocumentTypeDriversLicense, true},
		{"state_id", DocumentTypeStateID, true},
		{"passport", DocumentTypePassport, true},
		{"birth_certificate", DocumentTypeBirthCertificate, true},
		{"national_id", DocumentTypeNationalID, true},
		{"tax_id", DocumentTypeTaxID, true},
		{"voter_id", DocumentTypeVoterID, true},
		{"military_id", DocumentTypeMilitaryID, true},
		{"residence_permit", DocumentTypeResidencePermit, true},
		{"visa_document", DocumentTypeVisaDocument, true},
		{"invalid_type", DocumentType("invalid"), false},
		{"empty_type", DocumentType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.docType.IsValid(); got != tt.expected {
				t.Errorf("DocumentType.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDataSourceType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		dsType   DataSourceType
		expected bool
	}{
		{"dmv", DataSourceDMV, true},
		{"passport", DataSourcePassport, true},
		{"vital_records", DataSourceVitalRecords, true},
		{"national_registry", DataSourceNationalRegistry, true},
		{"tax_authority", DataSourceTaxAuthority, true},
		{"immigration", DataSourceImmigration, true},
		{"invalid_source", DataSourceType("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dsType.IsValid(); got != tt.expected {
				t.Errorf("DataSourceType.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestVerificationStatus_IsSuccess(t *testing.T) {
	tests := []struct {
		name     string
		status   VerificationStatus
		expected bool
	}{
		{"verified", VerificationStatusVerified, true},
		{"partial_match", VerificationStatusPartialMatch, true},
		{"not_verified", VerificationStatusNotVerified, false},
		{"not_found", VerificationStatusNotFound, false},
		{"expired", VerificationStatusExpired, false},
		{"revoked", VerificationStatusRevoked, false},
		{"error", VerificationStatusError, false},
		{"timeout", VerificationStatusTimeout, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsSuccess(); got != tt.expected {
				t.Errorf("VerificationStatus.IsSuccess() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestVerificationRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request VerificationRequest
		wantErr bool
	}{
		{
			name: "valid_request",
			request: VerificationRequest{
				WalletAddress: "ve1abc123",
				DocumentType:  DocumentTypeDriversLicense,
				Jurisdiction:  "US-CA",
				Fields: VerificationFields{
					DocumentNumber: "D1234567",
					LastName:       "Doe",
				},
			},
			wantErr: false,
		},
		{
			name: "missing_wallet_address",
			request: VerificationRequest{
				DocumentType: DocumentTypeDriversLicense,
				Jurisdiction: "US-CA",
				Fields: VerificationFields{
					DocumentNumber: "D1234567",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid_document_type",
			request: VerificationRequest{
				WalletAddress: "ve1abc123",
				DocumentType:  DocumentType("invalid"),
				Jurisdiction:  "US-CA",
				Fields: VerificationFields{
					DocumentNumber: "D1234567",
				},
			},
			wantErr: true,
		},
		{
			name: "missing_jurisdiction",
			request: VerificationRequest{
				WalletAddress: "ve1abc123",
				DocumentType:  DocumentTypeDriversLicense,
				Fields: VerificationFields{
					DocumentNumber: "D1234567",
				},
			},
			wantErr: true,
		},
		{
			name: "missing_document_number",
			request: VerificationRequest{
				WalletAddress: "ve1abc123",
				DocumentType:  DocumentTypeDriversLicense,
				Jurisdiction:  "US-CA",
				Fields:        VerificationFields{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("VerificationRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConsent_IsValid(t *testing.T) {
	now := time.Now()
	pastTime := now.Add(-24 * time.Hour)
	futureTime := now.Add(24 * time.Hour)

	tests := []struct {
		name     string
		consent  Consent
		expected bool
	}{
		{
			name: "valid_active_consent",
			consent: Consent{
				Active:    true,
				ExpiresAt: futureTime,
			},
			expected: true,
		},
		{
			name: "inactive_consent",
			consent: Consent{
				Active:    false,
				ExpiresAt: futureTime,
			},
			expected: false,
		},
		{
			name: "expired_consent",
			consent: Consent{
				Active:    true,
				ExpiresAt: pastTime,
			},
			expected: false,
		},
		{
			name: "revoked_consent",
			consent: Consent{
				Active:    true,
				ExpiresAt: futureTime,
				RevokedAt: &pastTime,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.consent.IsValid(); got != tt.expected {
				t.Errorf("Consent.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// ============================================================================
// Config Tests
// ============================================================================

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Enabled {
		t.Error("Default config should have Enabled=false")
	}

	if cfg.ServiceID != "govdata-service" {
		t.Errorf("ServiceID = %s, want govdata-service", cfg.ServiceID)
	}

	if cfg.DefaultTimeout != 30*time.Second {
		t.Errorf("DefaultTimeout = %v, want 30s", cfg.DefaultTimeout)
	}

	if !cfg.RequireConsent {
		t.Error("RequireConsent should be true by default")
	}

	if !cfg.Audit.Enabled {
		t.Error("Audit should be enabled by default")
	}

	if !cfg.VEIDIntegration.Enabled {
		t.Error("VEIDIntegration should be enabled by default")
	}

	if cfg.VEIDIntegration.MinConfidenceThreshold != 0.7 {
		t.Errorf("MinConfidenceThreshold = %f, want 0.7", cfg.VEIDIntegration.MinConfidenceThreshold)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid_default_config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "missing_service_id",
			config: Config{
				ServiceID:      "",
				DefaultTimeout: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid_timeout",
			config: Config{
				ServiceID:      "test",
				DefaultTimeout: 0,
			},
			wantErr: true,
		},
		{
			name: "negative_retries",
			config: Config{
				ServiceID:      "test",
				DefaultTimeout: 30 * time.Second,
				MaxRetries:     -1,
			},
			wantErr: true,
		},
		{
			name: "invalid_rate_limit",
			config: Config{
				ServiceID:      "test",
				DefaultTimeout: 30 * time.Second,
				RateLimits: RateLimitConfig{
					Enabled:           true,
					RequestsPerMinute: 0,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid_confidence_threshold",
			config: Config{
				ServiceID:      "test",
				DefaultTimeout: 30 * time.Second,
				VEIDIntegration: VEIDIntegrationConfig{
					Enabled:                true,
					MinConfidenceThreshold: 1.5, // Invalid: > 1.0
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultJurisdictions(t *testing.T) {
	jConfig := DefaultJurisdictions()

	expectedJurisdictions := []string{"US", "US-CA", "EU", "GB", "AU"}

	for _, code := range expectedJurisdictions {
		j, ok := jConfig.Jurisdictions[code]
		if !ok {
			t.Errorf("Missing jurisdiction: %s", code)
			continue
		}

		if !j.Active {
			t.Errorf("Jurisdiction %s should be active", code)
		}

		if len(j.SupportedDocuments) == 0 {
			t.Errorf("Jurisdiction %s should have supported documents", code)
		}

		if len(j.DataSources) == 0 {
			t.Errorf("Jurisdiction %s should have data sources", code)
		}
	}

	// Check GDPR/CCPA compliance flags
	if !jConfig.Jurisdictions["EU"].GDPRApplicable {
		t.Error("EU should have GDPR applicable")
	}
	if !jConfig.Jurisdictions["US-CA"].CCPAApplicable {
		t.Error("US-CA should have CCPA applicable")
	}
}

// ============================================================================
// Service Tests
// ============================================================================

func TestNewService(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true

	svc, err := NewService(cfg)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	if svc == nil {
		t.Fatal("NewService() returned nil")
	}
}

func TestNewService_InvalidConfig(t *testing.T) {
	cfg := Config{
		ServiceID:      "",
		DefaultTimeout: 30 * time.Second,
	}

	_, err := NewService(cfg)
	if err == nil {
		t.Error("NewService() should fail with invalid config")
	}
}

func TestService_Lifecycle(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true

	svc, err := NewService(cfg)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	ctx := context.Background()

	// Start service
	if err := svc.Start(ctx); err != nil {
		t.Errorf("Start() error = %v", err)
	}

	// Check health
	if !svc.IsHealthy() {
		t.Error("Service should be healthy after start")
	}

	// Get status
	status := svc.GetStatus()
	if status.Version != GovDataVersion {
		t.Errorf("Version = %s, want %s", status.Version, GovDataVersion)
	}

	// Stop service
	if err := svc.Stop(ctx); err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	if svc.IsHealthy() {
		t.Error("Service should not be healthy after stop")
	}
}

func TestService_VerifyDocument(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.RequireConsent = false // Disable for testing
	cfg.RateLimits.Enabled = false

	// Add a test adapter
	cfg.Adapters = map[string]AdapterConfig{
		"us-dmv": {
			Type:         DataSourceDMV,
			Jurisdiction: "US",
			Enabled:      true,
			Endpoint:     "https://test.example.com",
			Timeout:      30 * time.Second,
			SupportedDocuments: []DocumentType{
				DocumentTypeDriversLicense,
				DocumentTypeStateID,
			},
		},
	}

	svc, err := NewService(cfg)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	ctx := context.Background()
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() { _ = svc.Stop(ctx) }()

	req := &VerificationRequest{
		WalletAddress: "ve1abc123",
		DocumentType:  DocumentTypeDriversLicense,
		Jurisdiction:  "US",
		Fields: VerificationFields{
			DocumentNumber: "D1234567",
			FirstName:      "John",
			LastName:       "Doe",
			DateOfBirth:    time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	resp, err := svc.VerifyDocument(ctx, req)
	if err != nil {
		t.Fatalf("VerifyDocument() error = %v", err)
	}

	if resp == nil {
		t.Fatal("VerifyDocument() returned nil response")
	}

	if resp.Status != VerificationStatusVerified {
		t.Errorf("Status = %s, want verified", resp.Status)
	}

	if resp.Confidence < 0.9 {
		t.Errorf("Confidence = %f, want >= 0.9", resp.Confidence)
	}

	if resp.DataSourceType != DataSourceDMV {
		t.Errorf("DataSourceType = %s, want dmv", resp.DataSourceType)
	}

	if resp.AuditLogID == "" {
		t.Error("AuditLogID should not be empty")
	}
}

func TestService_VerifyDocument_RateLimited(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.RequireConsent = false
	cfg.RateLimits = RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 1,
		RequestsPerHour:   10,
		RequestsPerDay:    100,
	}

	cfg.Adapters = map[string]AdapterConfig{
		"us-dmv": {
			Type:         DataSourceDMV,
			Jurisdiction: "US",
			Enabled:      true,
			Endpoint:     "https://test.example.com",
			Timeout:      30 * time.Second,
		},
	}

	svc, err := NewService(cfg)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	ctx := context.Background()
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() { _ = svc.Stop(ctx) }()

	req := &VerificationRequest{
		WalletAddress: "ve1ratelimit",
		DocumentType:  DocumentTypeDriversLicense,
		Jurisdiction:  "US",
		Fields: VerificationFields{
			DocumentNumber: "D1234567",
		},
	}

	// First request should succeed
	resp1, err := svc.VerifyDocument(ctx, req)
	if err != nil {
		t.Fatalf("First request error = %v", err)
	}
	if resp1.Status != VerificationStatusVerified {
		t.Errorf("First request status = %s, want verified", resp1.Status)
	}

	// Second request should be rate limited
	resp2, err := svc.VerifyDocument(ctx, req)
	if err != nil {
		t.Fatalf("Second request error = %v", err)
	}
	if resp2.Status != VerificationStatusRateLimited {
		t.Errorf("Second request status = %s, want rate_limited", resp2.Status)
	}
}

// ============================================================================
// Consent Tests
// ============================================================================

func TestService_ConsentWorkflow(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true

	svc, err := NewService(cfg)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	ctx := context.Background()
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() { _ = svc.Stop(ctx) }()

	// Grant consent
	consent := &Consent{
		WalletAddress: "ve1consent123",
		DocumentTypes: []DocumentType{DocumentTypeDriversLicense, DocumentTypePassport},
		Jurisdictions: []string{"US", "EU"},
		Purpose:       "Identity verification",
	}

	grantedConsent, err := svc.GrantConsent(ctx, consent)
	if err != nil {
		t.Fatalf("GrantConsent() error = %v", err)
	}

	if grantedConsent.ID == "" {
		t.Error("Consent ID should be generated")
	}

	if !grantedConsent.Active {
		t.Error("Consent should be active")
	}

	// Get consent
	retrieved, err := svc.GetConsent(ctx, grantedConsent.ID)
	if err != nil {
		t.Fatalf("GetConsent() error = %v", err)
	}

	if retrieved.WalletAddress != consent.WalletAddress {
		t.Errorf("WalletAddress = %s, want %s", retrieved.WalletAddress, consent.WalletAddress)
	}

	// List consents
	consents, err := svc.ListConsents(ctx, consent.WalletAddress)
	if err != nil {
		t.Fatalf("ListConsents() error = %v", err)
	}

	if len(consents) != 1 {
		t.Errorf("ListConsents() count = %d, want 1", len(consents))
	}

	// Revoke consent
	if err := svc.RevokeConsent(ctx, grantedConsent.ID); err != nil {
		t.Fatalf("RevokeConsent() error = %v", err)
	}

	// Verify consent is revoked
	retrieved2, err := svc.GetConsent(ctx, grantedConsent.ID)
	if err != nil {
		t.Fatalf("GetConsent() after revoke error = %v", err)
	}

	if retrieved2.Active {
		t.Error("Consent should not be active after revoke")
	}
}

// ============================================================================
// Jurisdiction Tests
// ============================================================================

func TestService_JurisdictionMethods(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true

	svc, err := NewService(cfg)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	ctx := context.Background()
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer svc.Stop(ctx)

	// List jurisdictions
	jurisdictions, err := svc.ListJurisdictions(ctx)
	if err != nil {
		t.Fatalf("ListJurisdictions() error = %v", err)
	}

	if len(jurisdictions) == 0 {
		t.Error("Should have jurisdictions")
	}

	// Get specific jurisdiction
	us, err := svc.GetJurisdiction(ctx, "US")
	if err != nil {
		t.Fatalf("GetJurisdiction(US) error = %v", err)
	}

	if us.Code != "US" {
		t.Errorf("Code = %s, want US", us.Code)
	}

	// Check supported
	if !svc.IsJurisdictionSupported(ctx, "US") {
		t.Error("US should be supported")
	}

	if svc.IsJurisdictionSupported(ctx, "XX") {
		t.Error("XX should not be supported")
	}

	// Get supported documents
	docs, err := svc.GetSupportedDocuments(ctx, "US")
	if err != nil {
		t.Fatalf("GetSupportedDocuments(US) error = %v", err)
	}

	if len(docs) == 0 {
		t.Error("US should have supported documents")
	}
}

// ============================================================================
// Adapter Tests
// ============================================================================

func TestDMVAdapter_Verify(t *testing.T) {
	config := AdapterConfig{
		Type:         DataSourceDMV,
		Jurisdiction: "US",
		Enabled:      true,
		Endpoint:     "https://test.example.com",
		Timeout:      30 * time.Second,
	}

	adapter := newDMVAdapter(config)

	ctx := context.Background()
	req := &VerificationRequest{
		WalletAddress: "ve1test",
		DocumentType:  DocumentTypeDriversLicense,
		Jurisdiction:  "US",
		Fields: VerificationFields{
			DocumentNumber: "D1234567",
			FirstName:      "John",
			LastName:       "Doe",
			DateOfBirth:    time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	resp, err := adapter.Verify(ctx, req)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if resp.Status != VerificationStatusVerified {
		t.Errorf("Status = %s, want verified", resp.Status)
	}

	if resp.DataSourceType != DataSourceDMV {
		t.Errorf("DataSourceType = %s, want dmv", resp.DataSourceType)
	}

	// Check field results
	if resp.FieldResults["document_number"].Match != FieldMatchExact {
		t.Error("document_number should have exact match")
	}
}

func TestDMVAdapter_VerifyExpiredDocument(t *testing.T) {
	config := AdapterConfig{
		Type:         DataSourceDMV,
		Jurisdiction: "US",
		Enabled:      true,
		Endpoint:     "https://test.example.com",
		Timeout:      30 * time.Second,
	}

	adapter := newDMVAdapter(config)

	ctx := context.Background()
	req := &VerificationRequest{
		WalletAddress: "ve1test",
		DocumentType:  DocumentTypeDriversLicense,
		Jurisdiction:  "US",
		Fields: VerificationFields{
			DocumentNumber: "D1234567",
			ExpirationDate: time.Now().Add(-365 * 24 * time.Hour), // Expired 1 year ago
		},
	}

	resp, err := adapter.Verify(ctx, req)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if resp.Status != VerificationStatusExpired {
		t.Errorf("Status = %s, want expired", resp.Status)
	}

	if resp.DocumentValid {
		t.Error("DocumentValid should be false for expired document")
	}
}

func TestPassportAdapter_Verify(t *testing.T) {
	config := AdapterConfig{
		Type:         DataSourcePassport,
		Jurisdiction: "US",
		Enabled:      true,
		Endpoint:     "https://test.example.com",
		Timeout:      30 * time.Second,
	}

	adapter := newPassportAdapter(config)

	ctx := context.Background()
	req := &VerificationRequest{
		WalletAddress: "ve1test",
		DocumentType:  DocumentTypePassport,
		Jurisdiction:  "US",
		Fields: VerificationFields{
			DocumentNumber: "123456789", // Valid passport number format
			Nationality:    "US",
		},
	}

	resp, err := adapter.Verify(ctx, req)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if resp.Status != VerificationStatusVerified {
		t.Errorf("Status = %s, want verified", resp.Status)
	}
}

func TestTaxAuthorityAdapter_Verify(t *testing.T) {
	config := AdapterConfig{
		Type:         DataSourceTaxAuthority,
		Jurisdiction: "US",
		Enabled:      true,
		Endpoint:     "https://test.example.com",
		Timeout:      30 * time.Second,
	}

	adapter := newTaxAuthorityAdapter(config)

	ctx := context.Background()
	req := &VerificationRequest{
		WalletAddress: "ve1test",
		DocumentType:  DocumentTypeTaxID,
		Jurisdiction:  "US",
		Fields: VerificationFields{
			DocumentNumber: "123-45-6789",
			LastName:       "Doe",
		},
	}

	resp, err := adapter.Verify(ctx, req)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if resp.Status != VerificationStatusVerified {
		t.Errorf("Status = %s, want verified", resp.Status)
	}

	// Verify privacy-preserving response
	if _, hasRawData := resp.FieldResults["ssn"]; hasRawData {
		t.Error("Should not expose raw SSN in response")
	}
}

// ============================================================================
// VEID Integration Tests
// ============================================================================

func TestVEIDIntegrator_CreateScope(t *testing.T) {
	config := VEIDIntegrationConfig{
		Enabled:                    true,
		BaseScoreContribution:      0.25,
		GovernmentSourceWeight:     1.5,
		MinConfidenceThreshold:     0.7,
		ScopeExpiryDuration:        365 * 24 * time.Hour,
		VerificationFreshnessDecay: 180 * 24 * time.Hour,
	}

	integrator := newVEIDIntegrator(config)

	ctx := context.Background()
	verification := &VerificationResponse{
		RequestID:      "test-123",
		Status:         VerificationStatusVerified,
		Confidence:     0.95,
		DataSourceType: DataSourcePassport,
		Jurisdiction:   "US",
		DocumentValid:  true,
		VerifiedAt:     time.Now(),
		FieldResults: map[string]FieldVerificationResult{
			"document_number": {Match: FieldMatchExact, Confidence: 1.0},
			"first_name":      {Match: FieldMatchExact, Confidence: 1.0},
			"last_name":       {Match: FieldMatchExact, Confidence: 1.0},
		},
	}

	scope, err := integrator.CreateScope(ctx, verification)
	if err != nil {
		t.Fatalf("CreateScope() error = %v", err)
	}

	if scope == nil {
		t.Fatal("CreateScope() returned nil")
	}

	if scope.ID == "" {
		t.Error("Scope ID should be generated")
	}

	if scope.Status != "active" {
		t.Errorf("Status = %s, want active", scope.Status)
	}

	if scope.Confidence != verification.Confidence {
		t.Errorf("Confidence = %f, want %f", scope.Confidence, verification.Confidence)
	}

	if scope.ScoreContribution <= 0 {
		t.Error("ScoreContribution should be positive")
	}

	if len(scope.FieldsVerified) != 3 {
		t.Errorf("FieldsVerified count = %d, want 3", len(scope.FieldsVerified))
	}
}

func TestVEIDIntegrator_CreateScope_LowConfidence(t *testing.T) {
	config := VEIDIntegrationConfig{
		Enabled:                true,
		MinConfidenceThreshold: 0.7,
	}

	integrator := newVEIDIntegrator(config)

	ctx := context.Background()
	verification := &VerificationResponse{
		RequestID:      "test-123",
		Status:         VerificationStatusVerified,
		Confidence:     0.5, // Below threshold
		DataSourceType: DataSourceDMV,
	}

	_, err := integrator.CreateScope(ctx, verification)
	if err == nil {
		t.Error("CreateScope() should fail for low confidence")
	}
}

func TestVEIDIntegrator_ComputeScoreContribution(t *testing.T) {
	config := VEIDIntegrationConfig{
		Enabled:                true,
		BaseScoreContribution:  0.25,
		GovernmentSourceWeight: 1.5,
	}

	integrator := newVEIDIntegrator(config).(*veidIntegrator)

	tests := []struct {
		name         string
		verification *VerificationResponse
		minExpected  float64
		maxExpected  float64
	}{
		{
			name: "passport_high_confidence",
			verification: &VerificationResponse{
				Status:         VerificationStatusVerified,
				Confidence:     0.98,
				DataSourceType: DataSourcePassport,
				DocumentValid:  true,
				FieldResults:   map[string]FieldVerificationResult{"a": {}, "b": {}, "c": {}, "d": {}},
			},
			minExpected: 0.4,
			maxExpected: 0.5,
		},
		{
			name: "dmv_medium_confidence",
			verification: &VerificationResponse{
				Status:         VerificationStatusVerified,
				Confidence:     0.8,
				DataSourceType: DataSourceDMV,
				DocumentValid:  true,
			},
			minExpected: 0.25,
			maxExpected: 0.45,
		},
		{
			name:         "nil_verification",
			verification: nil,
			minExpected:  0,
			maxExpected:  0,
		},
		{
			name: "failed_verification",
			verification: &VerificationResponse{
				Status:     VerificationStatusNotVerified,
				Confidence: 0.3,
			},
			minExpected: 0,
			maxExpected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := integrator.ComputeScoreContribution(tt.verification)
			if score < tt.minExpected || score > tt.maxExpected {
				t.Errorf("ComputeScoreContribution() = %f, want between %f and %f",
					score, tt.minExpected, tt.maxExpected)
			}
		})
	}
}

func TestMapToVEIDVerificationLevel(t *testing.T) {
	tests := []struct {
		name         string
		verification *VerificationResponse
		expected     string
	}{
		{
			name: "high_confidence_verified",
			verification: &VerificationResponse{
				Status:     VerificationStatusVerified,
				Confidence: 0.98,
			},
			expected: "government_verified_high",
		},
		{
			name: "medium_confidence_verified",
			verification: &VerificationResponse{
				Status:     VerificationStatusVerified,
				Confidence: 0.85,
			},
			expected: "government_verified_medium",
		},
		{
			name: "low_confidence_verified",
			verification: &VerificationResponse{
				Status:     VerificationStatusVerified,
				Confidence: 0.75,
			},
			expected: "government_verified_low",
		},
		{
			name: "partial_match",
			verification: &VerificationResponse{
				Status:     VerificationStatusPartialMatch,
				Confidence: 0.6,
			},
			expected: "government_partial",
		},
		{
			name: "expired",
			verification: &VerificationResponse{
				Status: VerificationStatusExpired,
			},
			expected: "government_expired",
		},
		{
			name:         "nil_verification",
			verification: nil,
			expected:     "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := MapToVEIDVerificationLevel(tt.verification)
			if level != tt.expected {
				t.Errorf("MapToVEIDVerificationLevel() = %s, want %s", level, tt.expected)
			}
		})
	}
}

func TestGetVerificationTrustLevel(t *testing.T) {
	tests := []struct {
		name         string
		verification *VerificationResponse
		minExpected  int
		maxExpected  int
	}{
		{
			name: "passport_verified",
			verification: &VerificationResponse{
				Status:         VerificationStatusVerified,
				Confidence:     0.95,
				DataSourceType: DataSourcePassport,
				DocumentValid:  true,
			},
			minExpected: 100,
			maxExpected: 100,
		},
		{
			name: "dmv_verified",
			verification: &VerificationResponse{
				Status:         VerificationStatusVerified,
				Confidence:     0.8,
				DataSourceType: DataSourceDMV,
				DocumentValid:  true,
			},
			minExpected: 80,
			maxExpected: 100,
		},
		{
			name:         "nil_verification",
			verification: nil,
			minExpected:  0,
			maxExpected:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := GetVerificationTrustLevel(tt.verification)
			if level < tt.minExpected || level > tt.maxExpected {
				t.Errorf("GetVerificationTrustLevel() = %d, want between %d and %d",
					level, tt.minExpected, tt.maxExpected)
			}
		})
	}
}

// ============================================================================
// Audit Tests
// ============================================================================

func TestAuditLogger_LogAndRetrieve(t *testing.T) {
	config := AuditConfig{
		Enabled:       true,
		RetentionDays: 365,
	}

	logger := newAuditLogger(config)
	ctx := context.Background()

	entry := &AuditLogEntry{
		RequestID:     "test-123",
		Action:        AuditActionVerify,
		WalletAddress: "ve1audit",
		Jurisdiction:  "US",
		DocumentType:  DocumentTypeDriversLicense,
		DataSource:    DataSourceDMV,
		Status:        VerificationStatusVerified,
		Timestamp:     time.Now(),
		Duration:      100 * time.Millisecond,
	}

	// Log entry
	if err := logger.Log(ctx, entry); err != nil {
		t.Fatalf("Log() error = %v", err)
	}

	if entry.ID == "" {
		t.Error("Entry ID should be generated")
	}

	// Retrieve entry
	retrieved, err := logger.Get(ctx, entry.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if retrieved.RequestID != entry.RequestID {
		t.Errorf("RequestID = %s, want %s", retrieved.RequestID, entry.RequestID)
	}
}

func TestAuditLogger_ListWithFilter(t *testing.T) {
	config := AuditConfig{
		Enabled:       true,
		RetentionDays: 365,
	}

	logger := newAuditLogger(config)
	ctx := context.Background()

	// Log multiple entries
	for i := 0; i < 5; i++ {
		entry := &AuditLogEntry{
			RequestID:     "test-" + string(rune('a'+i)),
			Action:        AuditActionVerify,
			WalletAddress: "ve1audit",
			Jurisdiction:  "US",
			Status:        VerificationStatusVerified,
		}
		if err := logger.Log(ctx, entry); err != nil {
			t.Fatalf("Log() error = %v", err)
		}
	}

	// List with filter
	filter := AuditLogFilter{
		WalletAddress: "ve1audit",
		Limit:         3,
	}

	entries, err := logger.List(ctx, filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("List() count = %d, want 3", len(entries))
	}
}

func TestAuditLogger_Export(t *testing.T) {
	config := AuditConfig{
		Enabled:       true,
		RetentionDays: 365,
	}

	logger := newAuditLogger(config)
	ctx := context.Background()

	entry := &AuditLogEntry{
		RequestID:     "export-test",
		Action:        AuditActionVerify,
		WalletAddress: "ve1export",
		Jurisdiction:  "US",
		DocumentType:  DocumentTypePassport,
		DataSource:    DataSourcePassport,
		Status:        VerificationStatusVerified,
	}

	if err := logger.Log(ctx, entry); err != nil {
		t.Fatalf("Log() error = %v", err)
	}

	// Export as JSON
	jsonData, err := logger.Export(ctx, AuditLogFilter{}, "json")
	if err != nil {
		t.Fatalf("Export(json) error = %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("JSON export should not be empty")
	}

	// Export as CSV
	csvData, err := logger.Export(ctx, AuditLogFilter{}, "csv")
	if err != nil {
		t.Fatalf("Export(csv) error = %v", err)
	}

	if len(csvData) == 0 {
		t.Error("CSV export should not be empty")
	}
}

// ============================================================================
// Rate Limiter Tests
// ============================================================================

func TestRateLimiter_Allow(t *testing.T) {
	config := RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 3,
		RequestsPerHour:   10,
		RequestsPerDay:    100,
	}

	limiter := newRateLimiter(config)
	ctx := context.Background()
	wallet := "ve1ratelimittest"

	// First 3 requests should succeed
	for i := 0; i < 3; i++ {
		allowed, err := limiter.Allow(ctx, wallet)
		if err != nil {
			t.Fatalf("Allow() error = %v", err)
		}
		if !allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 4th request should be rate limited
	allowed, err := limiter.Allow(ctx, wallet)
	if err != nil {
		t.Fatalf("Allow() error = %v", err)
	}
	if allowed {
		t.Error("4th request should be rate limited")
	}
}

func TestRateLimiter_GetRemaining(t *testing.T) {
	config := RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 10,
		RequestsPerHour:   100,
		RequestsPerDay:    1000,
	}

	limiter := newRateLimiter(config)
	ctx := context.Background()
	wallet := "ve1remaining"

	// Check initial remaining
	info, err := limiter.GetRemaining(ctx, wallet)
	if err != nil {
		t.Fatalf("GetRemaining() error = %v", err)
	}

	if info.RemainingMinute != 10 {
		t.Errorf("RemainingMinute = %d, want 10", info.RemainingMinute)
	}

	// Use one request
	_, _ = limiter.Allow(ctx, wallet)

	info, _ = limiter.GetRemaining(ctx, wallet)
	if info.RemainingMinute != 9 {
		t.Errorf("RemainingMinute after 1 request = %d, want 9", info.RemainingMinute)
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	config := RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 2,
		RequestsPerHour:   10,
		RequestsPerDay:    100,
	}

	limiter := newRateLimiter(config)
	ctx := context.Background()
	wallet := "ve1reset"

	// Use all requests
	_, _ = limiter.Allow(ctx, wallet)
	_, _ = limiter.Allow(ctx, wallet)

	// Should be rate limited
	allowed, _ := limiter.Allow(ctx, wallet)
	if allowed {
		t.Error("Should be rate limited")
	}

	// Reset
	if err := limiter.Reset(ctx, wallet); err != nil {
		t.Fatalf("Reset() error = %v", err)
	}

	// Should be allowed again
	allowed, _ = limiter.Allow(ctx, wallet)
	if !allowed {
		t.Error("Should be allowed after reset")
	}
}

// ============================================================================
// Batch Verification Tests
// ============================================================================

func TestService_VerifyDocumentBatch(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.RequireConsent = false
	cfg.RateLimits.Enabled = false

	cfg.Adapters = map[string]AdapterConfig{
		"us-dmv": {
			Type:         DataSourceDMV,
			Jurisdiction: "US",
			Enabled:      true,
			Endpoint:     "https://test.example.com",
			Timeout:      30 * time.Second,
		},
	}

	svc, err := NewService(cfg)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	ctx := context.Background()
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer svc.Stop(ctx)

	batchReq := &BatchVerificationRequest{
		BatchID: "batch-test-1",
		Requests: []VerificationRequest{
			{
				WalletAddress: "ve1batch1",
				DocumentType:  DocumentTypeDriversLicense,
				Jurisdiction:  "US",
				Fields: VerificationFields{
					DocumentNumber: "D1111111",
				},
			},
			{
				WalletAddress: "ve1batch2",
				DocumentType:  DocumentTypeDriversLicense,
				Jurisdiction:  "US",
				Fields: VerificationFields{
					DocumentNumber: "D2222222",
				},
			},
		},
	}

	resp, err := svc.VerifyDocumentBatch(ctx, batchReq)
	if err != nil {
		t.Fatalf("VerifyDocumentBatch() error = %v", err)
	}

	if resp.BatchID != batchReq.BatchID {
		t.Errorf("BatchID = %s, want %s", resp.BatchID, batchReq.BatchID)
	}

	if resp.TotalRequests != 2 {
		t.Errorf("TotalRequests = %d, want 2", resp.TotalRequests)
	}

	if resp.SuccessCount != 2 {
		t.Errorf("SuccessCount = %d, want 2", resp.SuccessCount)
	}

	if len(resp.Responses) != 2 {
		t.Errorf("Responses count = %d, want 2", len(resp.Responses))
	}
}

