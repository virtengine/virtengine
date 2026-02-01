// Package govdata provides government data source integration for identity verification.
//
// SECURITY-004: Tests for privacy and cost management components
package govdata

import (
	"context"
	"testing"
	"time"
)

// ============================================================================
// Privacy Manager Tests
// ============================================================================

func TestPrivacyManager_HashPII(t *testing.T) {
	pm := newPrivacyManager(DefaultPrivacyConfig())

	// Test hashing
	hash1 := pm.HashPII("test@example.com", []byte("salt1234567890ab"))
	hash2 := pm.HashPII("test@example.com", []byte("salt1234567890ab"))
	hash3 := pm.HashPII("different@example.com", []byte("salt1234567890ab"))

	// Same input + salt = same hash
	if hash1 != hash2 {
		t.Errorf("Same input should produce same hash: %s != %s", hash1, hash2)
	}

	// Different input = different hash
	if hash1 == hash3 {
		t.Error("Different input should produce different hash")
	}
}

func TestPrivacyManager_AnonymizePII(t *testing.T) {
	pm := newPrivacyManager(DefaultPrivacyConfig())

	anon1 := pm.AnonymizePII("user123")
	anon2 := pm.AnonymizePII("user123")
	anon3 := pm.AnonymizePII("user456")

	// Should start with ANON- prefix
	if len(anon1) < 5 || anon1[:5] != "ANON-" {
		t.Errorf("Expected ANON- prefix, got: %s", anon1)
	}

	// Same input = same pseudonym
	if anon1 != anon2 {
		t.Errorf("Same input should produce same pseudonym: %s != %s", anon1, anon2)
	}

	// Different input = different pseudonym
	if anon1 == anon3 {
		t.Error("Different input should produce different pseudonym")
	}
}

func TestPrivacyManager_ValidateDataMinimization(t *testing.T) {
	pm := newPrivacyManager(DefaultPrivacyConfig())

	tests := []struct {
		name           string
		fields         map[string]string
		requiredFields []string
		wantErr        bool
	}{
		{
			name: "valid - only required fields",
			fields: map[string]string{
				"name": "John Doe",
				"dob":  "1990-01-01",
			},
			requiredFields: []string{"name", "dob"},
			wantErr:        false,
		},
		{
			name: "invalid - extra fields provided",
			fields: map[string]string{
				"name":         "John Doe",
				"dob":          "1990-01-01",
				"ssn":          "123-45-6789",
				"bank_account": "12345678",
			},
			requiredFields: []string{"name", "dob"},
			wantErr:        true,
		},
		{
			name: "valid - subset of required fields",
			fields: map[string]string{
				"name": "John Doe",
			},
			requiredFields: []string{"name", "dob"},
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pm.ValidateDataMinimization(tt.fields, tt.requiredFields)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDataMinimization() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPrivacyManager_ProcessErasureRequest(t *testing.T) {
	pm := newPrivacyManager(DefaultPrivacyConfig())
	ctx := context.Background()

	// Test normal erasure request
	req := &ErasureRequest{
		ID:             "erase-001",
		SubjectID:      "user-123",
		RequestedAt:    time.Now(),
		Reason:         ErasureReasonWithdrawConsent,
		DataCategories: []PIICategory{PIICategoryBasic, PIICategoryContact},
		Status:         ErasureStatusPending,
	}

	err := pm.ProcessErasureRequest(ctx, req)
	if err != nil {
		t.Errorf("ProcessErasureRequest() error = %v", err)
	}

	if req.Status != ErasureStatusCompleted {
		t.Errorf("Expected status %s, got %s", ErasureStatusCompleted, req.Status)
	}

	if req.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}
}

func TestPrivacyManager_ProcessErasureRequest_LegalHold(t *testing.T) {
	pm := newPrivacyManager(DefaultPrivacyConfig())
	ctx := context.Background()

	// Set legal hold
	pm.SetLegalHold("user-with-hold", time.Now().Add(24*time.Hour))

	req := &ErasureRequest{
		ID:             "erase-002",
		SubjectID:      "user-with-hold",
		RequestedAt:    time.Now(),
		Reason:         ErasureReasonNoLongerNecessary,
		DataCategories: []PIICategory{PIICategoryBasic},
		Status:         ErasureStatusPending,
	}

	err := pm.ProcessErasureRequest(ctx, req)
	if err != ErrErasureNotAllowed {
		t.Errorf("Expected ErrErasureNotAllowed, got: %v", err)
	}

	if req.Status != ErasureStatusRejected {
		t.Errorf("Expected status %s, got %s", ErasureStatusRejected, req.Status)
	}
}

func TestPrivacyManager_GetRetentionPolicy(t *testing.T) {
	pm := newPrivacyManager(DefaultPrivacyConfig())

	tests := []struct {
		jurisdiction       string
		expectedResultDays int
	}{
		{"EU", 30},
		{"GB", 30},
		{"US-CA", 90},
		{"US", 90},
		{"DE", 90}, // DE uses default since it's not exact match for EU/GB
	}

	for _, tt := range tests {
		t.Run(tt.jurisdiction, func(t *testing.T) {
			policy, err := pm.GetRetentionPolicy(tt.jurisdiction)
			if err != nil {
				t.Errorf("GetRetentionPolicy() error = %v", err)
				return
			}
			if policy.ResultRetentionDays != tt.expectedResultDays {
				t.Errorf("ResultRetentionDays = %d, want %d", policy.ResultRetentionDays, tt.expectedResultDays)
			}
		})
	}
}

// ============================================================================
// Utility Function Tests
// ============================================================================

func TestRedactPII(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"abc", "***"},
		{"abcd", "****"},
		{"hello", "he*lo"},
		{"test@example.com", "te************om"},
	}

	for _, tt := range tests {
		result := RedactPII(tt.input)
		if result != tt.expected {
			t.Errorf("RedactPII(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestRedactDocumentNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"AB123456", "****3456"},
		{"1234", "****"},       // 4 chars or less = all masked
		{"12345678901234", "**********1234"},
	}

	for _, tt := range tests {
		result := RedactDocumentNumber(tt.input)
		if result != tt.expected {
			t.Errorf("RedactDocumentNumber(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestIsGDPRJurisdiction(t *testing.T) {
	gdprJurisdictions := []string{
		"EU", "DE", "FR", "IT", "ES", "NL", "PL", "GB",
		"AT", "BE", "CZ", "DK", "FI", "GR", "HU", "IE",
		"LU", "PT", "RO", "SE", "IS", "NO", "LI",
	}

	nonGDPRJurisdictions := []string{
		"US", "CA", "JP", "AU", "CN", "BR", "MX",
	}

	for _, j := range gdprJurisdictions {
		if !IsGDPRJurisdiction(j) {
			t.Errorf("IsGDPRJurisdiction(%q) should be true", j)
		}
	}

	for _, j := range nonGDPRJurisdictions {
		if IsGDPRJurisdiction(j) {
			t.Errorf("IsGDPRJurisdiction(%q) should be false", j)
		}
	}
}

func TestIsCCPAJurisdiction(t *testing.T) {
	if !IsCCPAJurisdiction("US-CA") {
		t.Error("US-CA should be CCPA jurisdiction")
	}
	if !IsCCPAJurisdiction("CA") {
		t.Error("CA should be CCPA jurisdiction")
	}
	if IsCCPAJurisdiction("US-NY") {
		t.Error("US-NY should not be CCPA jurisdiction")
	}
}

// ============================================================================
// Cost Manager Tests
// ============================================================================

func TestCostManager_RecordCost(t *testing.T) {
	cm := newCostManager(DefaultCostConfig())
	ctx := context.Background()

	record := &CostRecord{
		ID:           "cost-001",
		AdapterName:  "aamva",
		Timestamp:    time.Now(),
		AmountCents:  50,
		Currency:     "USD",
		Operation:    "verify",
		Success:      true,
		RequestID:    "req-001",
		Jurisdiction: "US",
	}

	err := cm.RecordCost(ctx, record)
	if err != nil {
		t.Errorf("RecordCost() error = %v", err)
	}

	// Check daily total
	daily := cm.GetDailyTotal()
	if daily != 50 {
		t.Errorf("GetDailyTotal() = %d, want 50", daily)
	}

	// Check monthly total
	monthly := cm.GetMonthlyTotal()
	if monthly != 50 {
		t.Errorf("GetMonthlyTotal() = %d, want 50", monthly)
	}
}

func TestCostManager_CheckBudget(t *testing.T) {
	config := DefaultCostConfig()
	config.AdapterCosts["test"] = &AdapterCostConfig{
		DailyBudgetCents:   100,
		MonthlyBudgetCents: 1000,
		CostPerCallCents:   10,
	}
	cm := newCostManager(config)
	ctx := context.Background()

	// Should be within budget
	err := cm.CheckBudget(ctx, "test")
	if err != nil {
		t.Errorf("CheckBudget() should succeed, got error: %v", err)
	}

	// Add costs to exceed daily budget
	for i := 0; i < 10; i++ {
		_ = cm.RecordCost(ctx, &CostRecord{
			ID:          "cost-" + string(rune(i)),
			AdapterName: "test",
			Timestamp:   time.Now(),
			AmountCents: 10,
		})
	}

	// Should exceed daily budget
	err = cm.CheckBudget(ctx, "test")
	if err == nil {
		t.Error("CheckBudget() should fail after exceeding daily budget")
	}
}

func TestCostManager_GetSummary(t *testing.T) {
	cm := newCostManager(DefaultCostConfig())
	ctx := context.Background()

	now := time.Now()
	startTime := now.Add(-24 * time.Hour)
	endTime := now.Add(24 * time.Hour)

	// Add some costs
	adapters := []string{"aamva", "dvs", "eidas"}
	for i, adapter := range adapters {
		_ = cm.RecordCost(ctx, &CostRecord{
			ID:           "cost-" + adapter,
			AdapterName:  adapter,
			Timestamp:    now,
			AmountCents:  int64((i + 1) * 100),
			Currency:     "USD",
			Operation:    "verify",
			Success:      i%2 == 0,
			Jurisdiction: "US",
		})
	}

	summary, err := cm.GetSummary(ctx, "daily", startTime, endTime)
	if err != nil {
		t.Errorf("GetSummary() error = %v", err)
		return
	}

	// 100 + 200 + 300 = 600
	if summary.TotalCostCents != 600 {
		t.Errorf("TotalCostCents = %d, want 600", summary.TotalCostCents)
	}

	if summary.CallCount != 3 {
		t.Errorf("CallCount = %d, want 3", summary.CallCount)
	}

	// aamva (i=0) and eidas (i=2) are successful
	if summary.SuccessCount != 2 {
		t.Errorf("SuccessCount = %d, want 2", summary.SuccessCount)
	}

	if len(summary.ByAdapter) != 3 {
		t.Errorf("ByAdapter count = %d, want 3", len(summary.ByAdapter))
	}
}

func TestCostManager_CalculateCost(t *testing.T) {
	cm := newCostManager(DefaultCostConfig())

	// AAMVA - per_call billing
	cost := cm.CalculateCost("aamva", false)
	if cost != 50 { // CostPerCallCents
		t.Errorf("CalculateCost(aamva, false) = %d, want 50", cost)
	}

	// AAMVA - successful verification costs more
	cost = cm.CalculateCost("aamva", true)
	if cost != 75 { // CostPerSuccessfulVerificationCents
		t.Errorf("CalculateCost(aamva, true) = %d, want 75", cost)
	}
}

func TestCostManager_EstimateCost(t *testing.T) {
	cm := newCostManager(DefaultCostConfig())

	estimate, err := cm.EstimateCost("aamva")
	if err != nil {
		t.Errorf("EstimateCost() error = %v", err)
		return
	}

	if estimate.AdapterName != "aamva" {
		t.Errorf("AdapterName = %s, want aamva", estimate.AdapterName)
	}

	if estimate.EstimatedCostCents != 50 {
		t.Errorf("EstimatedCostCents = %d, want 50", estimate.EstimatedCostCents)
	}

	if estimate.Currency != "USD" {
		t.Errorf("Currency = %s, want USD", estimate.Currency)
	}
}

func TestCostManager_SetBudget(t *testing.T) {
	cm := newCostManager(DefaultCostConfig())
	ctx := context.Background()

	// Set budget for new adapter
	err := cm.SetBudget(ctx, "new_adapter", 50000, 2000)
	if err != nil {
		t.Errorf("SetBudget() error = %v", err)
		return
	}

	config, err := cm.GetAdapterCost("new_adapter")
	if err != nil {
		t.Errorf("GetAdapterCost() error = %v", err)
		return
	}

	if config.MonthlyBudgetCents != 50000 {
		t.Errorf("MonthlyBudgetCents = %d, want 50000", config.MonthlyBudgetCents)
	}

	if config.DailyBudgetCents != 2000 {
		t.Errorf("DailyBudgetCents = %d, want 2000", config.DailyBudgetCents)
	}
}

func TestCostManager_GetAlerts(t *testing.T) {
	config := DefaultCostConfig()
	config.AlertEnabled = true
	config.AlertThresholdPercent = 80
	config.AdapterCosts["test"] = &AdapterCostConfig{
		DailyBudgetCents: 100,
		CostPerCallCents: 10,
	}
	cm := newCostManager(config)
	ctx := context.Background()

	since := time.Now().Add(-1 * time.Hour)

	// Add costs to trigger warning (80% = 80 cents)
	for i := 0; i < 9; i++ { // 90 cents = 90%
		_ = cm.RecordCost(ctx, &CostRecord{
			ID:          "cost-" + string(rune(i)),
			AdapterName: "test",
			Timestamp:   time.Now(),
			AmountCents: 10,
		})
	}

	alerts, err := cm.GetAlerts(ctx, since)
	if err != nil {
		t.Errorf("GetAlerts() error = %v", err)
		return
	}

	// Should have at least one warning alert
	if len(alerts) == 0 {
		t.Error("Expected at least one alert")
	}
}

