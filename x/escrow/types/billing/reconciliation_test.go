// Package billing provides tests for reconciliation and dispute workflow types.
package billing

import (
	"bytes"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// testAddressForReconciliation generates a valid test bech32 address from a seed number
func testAddressForReconciliation(seed int) string {
	var buffer bytes.Buffer
	buffer.WriteString("A58856F0FD53BF058B4909A21AEC019107BA6")
	buffer.WriteString(string(rune('0' + (seed/100)%10)))
	buffer.WriteString(string(rune('0' + (seed/10)%10)))
	buffer.WriteString(string(rune('0' + seed%10)))
	res, _ := sdk.AccAddressFromHexUnsafe(buffer.String())
	return res.String()
}

func TestCorrectionValidation(t *testing.T) {
	now := time.Now().UTC()
	blockHeight := int64(12345)
	provider := testAddressForReconciliation(100)

	tests := []struct {
		name        string
		correction  *Correction
		expectError bool
		errContains string
	}{
		{
			name: "valid correction",
			correction: NewCorrection(
				"corr-001",
				"invoice-001",
				"settlement-001",
				CorrectionTypeInvoiceAdjustment,
				sdk.NewCoins(sdk.NewInt64Coin("uvirt", 900)),
				sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
				"billing_error",
				"Correction for billing error",
				provider,
				blockHeight,
				now,
			),
			expectError: false,
		},
		{
			name: "missing correction_id",
			correction: &Correction{
				InvoiceID:       "invoice-001",
				Type:            CorrectionTypeInvoiceAdjustment,
				Status:          CorrectionStatusPending,
				OriginalAmount:  sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
				CorrectedAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 900)),
				Reason:          "billing_error",
				RequestedBy:     provider,
			},
			expectError: true,
			errContains: "correction_id is required",
		},
		{
			name: "correction_id too long",
			correction: &Correction{
				CorrectionID:    "this-correction-id-is-way-too-long-and-exceeds-the-maximum-length-of-64-characters-so-it-should-fail-validation",
				InvoiceID:       "invoice-001",
				Type:            CorrectionTypeInvoiceAdjustment,
				Status:          CorrectionStatusPending,
				OriginalAmount:  sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
				CorrectedAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 900)),
				Reason:          "billing_error",
				RequestedBy:     provider,
			},
			expectError: true,
			errContains: "exceeds maximum length",
		},
		{
			name: "missing invoice_id",
			correction: &Correction{
				CorrectionID:    "corr-002",
				Type:            CorrectionTypeInvoiceAdjustment,
				Status:          CorrectionStatusPending,
				OriginalAmount:  sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
				CorrectedAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 900)),
				Reason:          "billing_error",
				RequestedBy:     provider,
			},
			expectError: true,
			errContains: "invoice_id is required",
		},
		{
			name: "invalid correction type",
			correction: &Correction{
				CorrectionID:    "corr-003",
				InvoiceID:       "invoice-001",
				Type:            CorrectionType(99),
				Status:          CorrectionStatusPending,
				OriginalAmount:  sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
				CorrectedAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 900)),
				Reason:          "billing_error",
				RequestedBy:     provider,
			},
			expectError: true,
			errContains: "invalid correction type",
		},
		{
			name: "missing reason",
			correction: &Correction{
				CorrectionID:    "corr-004",
				InvoiceID:       "invoice-001",
				Type:            CorrectionTypeInvoiceAdjustment,
				Status:          CorrectionStatusPending,
				OriginalAmount:  sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
				CorrectedAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 900)),
				Reason:          "",
				RequestedBy:     provider,
			},
			expectError: true,
			errContains: "reason is required",
		},
		{
			name: "invalid requested_by address",
			correction: &Correction{
				CorrectionID:    "corr-005",
				InvoiceID:       "invoice-001",
				Type:            CorrectionTypeInvoiceAdjustment,
				Status:          CorrectionStatusPending,
				OriginalAmount:  sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
				CorrectedAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 900)),
				Reason:          "billing_error",
				RequestedBy:     "invalid-address",
			},
			expectError: true,
			errContains: "invalid requested_by address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.correction.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				} else if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCorrectionLimit_DefaultValues(t *testing.T) {
	limit := DefaultCorrectionLimit()

	if limit.MaxCorrectionAmount.Empty() {
		t.Error("MaxCorrectionAmount should not be empty")
	}

	if limit.MaxCorrectionsPerPeriod == 0 {
		t.Error("MaxCorrectionsPerPeriod should be positive")
	}

	if limit.CorrectionWindowSeconds <= 0 {
		t.Error("CorrectionWindowSeconds should be positive")
	}

	if limit.RequireApprovalThreshold.Empty() {
		t.Error("RequireApprovalThreshold should not be empty")
	}

	if len(limit.AllowedCorrectionTypes) == 0 {
		t.Error("AllowedCorrectionTypes should have values")
	}

	// Verify all correction types are allowed by default
	expectedTypes := []CorrectionType{
		CorrectionTypeInvoiceAdjustment,
		CorrectionTypeUsageMismatch,
		CorrectionTypePayoutCorrection,
		CorrectionTypeRefundAdjustment,
		CorrectionTypeTaxCorrection,
		CorrectionTypeFeeAdjustment,
	}

	for _, ct := range expectedTypes {
		if !limit.IsCorrectionTypeAllowed(ct) {
			t.Errorf("expected correction type %s to be allowed", ct)
		}
	}

	// Test validation
	if err := limit.Validate(); err != nil {
		t.Errorf("DefaultCorrectionLimit should be valid: %v", err)
	}
}

func TestCorrectionLedgerEntry(t *testing.T) {
	now := time.Now().UTC()
	blockHeight := int64(12345)
	provider := testAddressForReconciliation(100)

	entry := NewCorrectionLedgerEntry(
		"entry-001",
		"corr-001",
		"invoice-001",
		CorrectionLedgerEntryTypeCreated,
		CorrectionStatusPending,
		CorrectionStatusPending,
		sdk.NewCoins(sdk.NewInt64Coin("uvirt", 100)),
		"correction created",
		provider,
		"txhash123",
		blockHeight,
		now,
	)

	if entry.EntryID != "entry-001" {
		t.Errorf("expected EntryID entry-001, got %s", entry.EntryID)
	}

	if entry.CorrectionID != "corr-001" {
		t.Errorf("expected CorrectionID corr-001, got %s", entry.CorrectionID)
	}

	if entry.InvoiceID != "invoice-001" {
		t.Errorf("expected InvoiceID invoice-001, got %s", entry.InvoiceID)
	}

	if entry.EntryType != CorrectionLedgerEntryTypeCreated {
		t.Errorf("expected EntryType created, got %s", entry.EntryType)
	}

	if err := entry.Validate(); err != nil {
		t.Errorf("Validate failed: %v", err)
	}

	// Test invalid entry
	invalidEntry := &CorrectionLedgerEntry{
		EntryID:      "",
		CorrectionID: "corr-001",
		InvoiceID:    "invoice-001",
		Initiator:    provider,
	}

	if err := invalidEntry.Validate(); err == nil {
		t.Error("expected validation error for empty EntryID")
	}
}

func TestCorrectionStatusTransitions(t *testing.T) {
	tests := []struct {
		name       string
		status     CorrectionStatus
		isTerminal bool
	}{
		{"pending is not terminal", CorrectionStatusPending, false},
		{"approved is not terminal", CorrectionStatusApproved, false},
		{"applied is terminal", CorrectionStatusApplied, true},
		{"rejected is terminal", CorrectionStatusRejected, true},
		{"cancelled is terminal", CorrectionStatusCancelled, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.status.IsTerminal() != tt.isTerminal {
				t.Errorf("status %s: IsTerminal() = %v, want %v", tt.status, tt.status.IsTerminal(), tt.isTerminal)
			}
		})
	}

	// Test status string representations
	for status, name := range CorrectionStatusNames {
		if status.String() != name {
			t.Errorf("expected status string %s, got %s", name, status.String())
		}
	}
}

func TestAuditEntry_NewAndValidate(t *testing.T) {
	now := time.Now().UTC()
	blockHeight := int64(12345)
	actor := testAddressForReconciliation(100)

	entry := NewAuditEntry(
		"audit-001",
		AuditActionDisputeInitiated,
		AuditEntityTypeDispute,
		"dispute-001",
		actor,
		AuditActorRoleCustomer,
		AuditOutcomeSuccess,
		blockHeight,
		now,
	)

	if entry.EntryID != "audit-001" {
		t.Errorf("expected EntryID audit-001, got %s", entry.EntryID)
	}

	if entry.Action != AuditActionDisputeInitiated {
		t.Errorf("expected Action %s, got %s", AuditActionDisputeInitiated, entry.Action)
	}

	if entry.EntityType != AuditEntityTypeDispute {
		t.Errorf("expected EntityType %s, got %s", AuditEntityTypeDispute, entry.EntityType)
	}

	if err := entry.Validate(); err != nil {
		t.Errorf("Validate failed: %v", err)
	}

	// Test chained methods
	entry.AddDetail("key1", "value1").WithClientInfo("192.168.1.1", "TestAgent/1.0")

	if entry.Details["key1"] != "value1" {
		t.Error("AddDetail did not set detail")
	}

	if entry.IPAddress != "192.168.1.1" {
		t.Error("WithClientInfo did not set IP address")
	}

	// Test WithError
	entry.WithError("test error")
	if entry.Outcome != AuditOutcomeFailure {
		t.Error("WithError should set outcome to failure")
	}

	// Test validation failures
	invalidTests := []struct {
		name   string
		entry  *AuditEntry
		errMsg string
	}{
		{
			name:   "missing entry_id",
			entry:  &AuditEntry{Action: AuditActionDisputeInitiated, EntityType: AuditEntityTypeDispute, EntityID: "disp-001", Actor: actor, ActorRole: "customer", Outcome: AuditOutcomeSuccess, BlockHeight: 1, Timestamp: now},
			errMsg: "entry_id is required",
		},
		{
			name:   "missing entity_id",
			entry:  &AuditEntry{EntryID: "audit-001", Action: AuditActionDisputeInitiated, EntityType: AuditEntityTypeDispute, Actor: actor, ActorRole: "customer", Outcome: AuditOutcomeSuccess, BlockHeight: 1, Timestamp: now},
			errMsg: "entity_id is required",
		},
		{
			name:   "missing actor",
			entry:  &AuditEntry{EntryID: "audit-001", Action: AuditActionDisputeInitiated, EntityType: AuditEntityTypeDispute, EntityID: "disp-001", ActorRole: "customer", Outcome: AuditOutcomeSuccess, BlockHeight: 1, Timestamp: now},
			errMsg: "actor is required",
		},
		{
			name:   "negative block_height",
			entry:  &AuditEntry{EntryID: "audit-001", Action: AuditActionDisputeInitiated, EntityType: AuditEntityTypeDispute, EntityID: "disp-001", Actor: actor, ActorRole: "customer", Outcome: AuditOutcomeSuccess, BlockHeight: -1, Timestamp: now},
			errMsg: "block_height must be non-negative",
		},
	}

	for _, tt := range invalidTests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.entry.Validate()
			if err == nil {
				t.Errorf("expected error containing %q", tt.errMsg)
			} else if !containsString(err.Error(), tt.errMsg) {
				t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
			}
		})
	}
}

func TestAuditActionTypes(t *testing.T) {
	// Verify all action types have string representations
	for action, name := range AuditActionTypeNames {
		if action.String() != name {
			t.Errorf("expected action string %s, got %s", name, action.String())
		}
		if !action.IsValid() {
			t.Errorf("action %s should be valid", action)
		}
	}

	// Test unknown action type
	unknownAction := AuditActionType(255)
	if unknownAction.IsValid() {
		t.Error("unknown action type should not be valid")
	}
	if !containsString(unknownAction.String(), "unknown") {
		t.Error("unknown action should have 'unknown' in string representation")
	}
}

func TestAlertThreshold_Evaluate(t *testing.T) {
	tests := []struct {
		name      string
		operator  ComparisonOperator
		actual    sdkmath.LegacyDec
		threshold sdkmath.LegacyDec
		expected  bool
	}{
		{"GT true", ComparisonOperatorGT, sdkmath.LegacyNewDec(10), sdkmath.LegacyNewDec(5), true},
		{"GT false", ComparisonOperatorGT, sdkmath.LegacyNewDec(5), sdkmath.LegacyNewDec(10), false},
		{"GT equal false", ComparisonOperatorGT, sdkmath.LegacyNewDec(5), sdkmath.LegacyNewDec(5), false},
		{"GTE true greater", ComparisonOperatorGTE, sdkmath.LegacyNewDec(10), sdkmath.LegacyNewDec(5), true},
		{"GTE true equal", ComparisonOperatorGTE, sdkmath.LegacyNewDec(5), sdkmath.LegacyNewDec(5), true},
		{"GTE false", ComparisonOperatorGTE, sdkmath.LegacyNewDec(4), sdkmath.LegacyNewDec(5), false},
		{"LT true", ComparisonOperatorLT, sdkmath.LegacyNewDec(3), sdkmath.LegacyNewDec(5), true},
		{"LT false", ComparisonOperatorLT, sdkmath.LegacyNewDec(10), sdkmath.LegacyNewDec(5), false},
		{"LTE true less", ComparisonOperatorLTE, sdkmath.LegacyNewDec(3), sdkmath.LegacyNewDec(5), true},
		{"LTE true equal", ComparisonOperatorLTE, sdkmath.LegacyNewDec(5), sdkmath.LegacyNewDec(5), true},
		{"LTE false", ComparisonOperatorLTE, sdkmath.LegacyNewDec(10), sdkmath.LegacyNewDec(5), false},
		{"EQ true", ComparisonOperatorEQ, sdkmath.LegacyNewDec(5), sdkmath.LegacyNewDec(5), true},
		{"EQ false", ComparisonOperatorEQ, sdkmath.LegacyNewDec(10), sdkmath.LegacyNewDec(5), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.operator.Evaluate(tt.actual, tt.threshold)
			if result != tt.expected {
				t.Errorf("%s.Evaluate(%s, %s) = %v, want %v",
					tt.operator, tt.actual, tt.threshold, result, tt.expected)
			}
		})
	}

	// Test unknown operator
	unknownOp := ComparisonOperator(99)
	if unknownOp.Evaluate(sdkmath.LegacyNewDec(5), sdkmath.LegacyNewDec(5)) {
		t.Error("unknown operator should return false")
	}
}

func TestAlertConfig_DefaultValues(t *testing.T) {
	config := DefaultAlertConfig()

	if !config.IsEnabled {
		t.Error("alert system should be enabled by default")
	}

	if config.DefaultCooldownSeconds <= 0 {
		t.Error("DefaultCooldownSeconds should be positive")
	}

	if config.DefaultEvaluationWindowSeconds <= 0 {
		t.Error("DefaultEvaluationWindowSeconds should be positive")
	}

	if config.MaxActiveAlerts == 0 {
		t.Error("MaxActiveAlerts should be positive")
	}

	if config.RetentionDays == 0 {
		t.Error("RetentionDays should be positive")
	}

	if len(config.DefaultThresholds) == 0 {
		t.Error("DefaultThresholds should have values")
	}

	if len(config.NotificationChannels) == 0 {
		t.Error("NotificationChannels should have values")
	}

	// Validate all default thresholds
	for i, threshold := range config.DefaultThresholds {
		if err := threshold.Validate(); err != nil {
			t.Errorf("DefaultThresholds[%d] validation failed: %v", i, err)
		}
	}

	// Validate config
	if err := config.Validate(); err != nil {
		t.Errorf("DefaultAlertConfig should be valid: %v", err)
	}
}

func TestAlertRule_Validate(t *testing.T) {
	now := time.Now().UTC()

	validThreshold := AlertThreshold{
		ThresholdID:             "threshold-001",
		Name:                    "Test Threshold",
		AlertType:               AlertTypeVarianceThresholdExceeded,
		Severity:                AlertSeverityWarning,
		ThresholdValue:          sdkmath.LegacyNewDecWithPrec(5, 2),
		ThresholdUnit:           ThresholdUnitPercentage,
		ComparisonOperator:      ComparisonOperatorGT,
		EvaluationWindowSeconds: 3600,
		CooldownSeconds:         1800,
		IsEnabled:               true,
	}

	tests := []struct {
		name        string
		rule        *AlertRule
		expectError bool
		errContains string
	}{
		{
			name: "valid rule",
			rule: &AlertRule{
				RuleID:      "rule-001",
				Name:        "Test Rule",
				Description: "A test rule",
				Thresholds:  []AlertThreshold{validThreshold},
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			expectError: false,
		},
		{
			name: "missing rule_id",
			rule: &AlertRule{
				Name:       "Test Rule",
				Thresholds: []AlertThreshold{validThreshold},
				CreatedAt:  now,
			},
			expectError: true,
			errContains: "rule_id is required",
		},
		{
			name: "missing name",
			rule: &AlertRule{
				RuleID:     "rule-001",
				Thresholds: []AlertThreshold{validThreshold},
				CreatedAt:  now,
			},
			expectError: true,
			errContains: "name is required",
		},
		{
			name: "no thresholds",
			rule: &AlertRule{
				RuleID:     "rule-001",
				Name:       "Test Rule",
				Thresholds: []AlertThreshold{},
				CreatedAt:  now,
			},
			expectError: true,
			errContains: "at least one threshold is required",
		},
		{
			name: "missing created_at",
			rule: &AlertRule{
				RuleID:     "rule-001",
				Name:       "Test Rule",
				Thresholds: []AlertThreshold{validThreshold},
			},
			expectError: true,
			errContains: "created_at is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				} else if !containsString(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestExportFilter_Validate(t *testing.T) {
	now := time.Now().UTC()
	provider := testAddressForReconciliation(100)
	customer := testAddressForReconciliation(101)

	tests := []struct {
		name        string
		filter      ExportFilter
		expectError bool
		errContains string
	}{
		{
			name: "valid filter",
			filter: ExportFilter{
				StartTime: now.Add(-24 * time.Hour),
				EndTime:   now,
				Provider:  provider,
				Customer:  customer,
			},
			expectError: false,
		},
		{
			name: "missing start_time",
			filter: ExportFilter{
				EndTime: now,
			},
			expectError: true,
			errContains: "start_time is required",
		},
		{
			name: "missing end_time",
			filter: ExportFilter{
				StartTime: now.Add(-24 * time.Hour),
			},
			expectError: true,
			errContains: "end_time is required",
		},
		{
			name: "end before start",
			filter: ExportFilter{
				StartTime: now,
				EndTime:   now.Add(-24 * time.Hour),
			},
			expectError: true,
			errContains: "end_time must be after start_time",
		},
		{
			name: "invalid provider address",
			filter: ExportFilter{
				StartTime: now.Add(-24 * time.Hour),
				EndTime:   now,
				Provider:  "invalid-address",
			},
			expectError: true,
			errContains: "invalid provider address",
		},
		{
			name: "invalid customer address",
			filter: ExportFilter{
				StartTime: now.Add(-24 * time.Hour),
				EndTime:   now,
				Customer:  "invalid-address",
			},
			expectError: true,
			errContains: "invalid customer address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				} else if !containsString(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestExportRequest_Lifecycle(t *testing.T) {
	now := time.Now().UTC()
	requester := testAddressForReconciliation(100)

	filter := ExportFilter{
		StartTime: now.Add(-24 * time.Hour),
		EndTime:   now,
	}

	request := NewExportRequest(
		"export-001",
		requester,
		ExportTypeReconciliationReport,
		ExportFormatJSON,
		filter,
		true,  // includeLineItems
		false, // includeAuditTrail
		now,
	)

	// Verify initial state
	if request.RequestID != "export-001" {
		t.Errorf("expected RequestID export-001, got %s", request.RequestID)
	}

	if request.Status != ExportStatusPending {
		t.Errorf("expected status pending, got %s", request.Status)
	}

	// Mark in progress
	if err := request.MarkInProgress(); err != nil {
		t.Errorf("MarkInProgress failed: %v", err)
	}

	if request.Status != ExportStatusInProgress {
		t.Errorf("expected status in_progress, got %s", request.Status)
	}

	// Cannot mark in progress twice
	if err := request.MarkInProgress(); err == nil {
		t.Error("expected error when marking in-progress export as in-progress")
	}

	// Mark completed
	completedAt := now.Add(time.Minute)
	if err := request.MarkCompleted("bafytest123", 1024, 50, completedAt); err != nil {
		t.Errorf("MarkCompleted failed: %v", err)
	}

	if request.Status != ExportStatusCompleted {
		t.Errorf("expected status completed, got %s", request.Status)
	}

	if request.OutputArtifactCID != "bafytest123" {
		t.Errorf("expected OutputArtifactCID bafytest123, got %s", request.OutputArtifactCID)
	}

	if request.FileSize != 1024 {
		t.Errorf("expected FileSize 1024, got %d", request.FileSize)
	}

	if request.RecordCount != 50 {
		t.Errorf("expected RecordCount 50, got %d", request.RecordCount)
	}

	// Test failure path
	request2 := NewExportRequest(
		"export-002",
		requester,
		ExportTypeDisputeReport,
		ExportFormatCSV,
		filter,
		false,
		false,
		now,
	)

	_ = request2.MarkInProgress()
	failedAt := now.Add(time.Minute)
	if err := request2.MarkFailed("connection timeout", failedAt); err != nil {
		t.Errorf("MarkFailed failed: %v", err)
	}

	if request2.Status != ExportStatusFailed {
		t.Errorf("expected status failed, got %s", request2.Status)
	}

	if request2.ErrorMessage != "connection timeout" {
		t.Errorf("expected ErrorMessage 'connection timeout', got %s", request2.ErrorMessage)
	}
}

func TestExportFormat_ContentType(t *testing.T) {
	tests := []struct {
		format            ExportFormat
		expectedContent   string
		expectedExtension string
	}{
		{ExportFormatCSV, "text/csv", ".csv"},
		{ExportFormatJSON, "application/json", ".json"},
		{ExportFormatExcel, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", ".xlsx"},
	}

	for _, tt := range tests {
		t.Run(tt.format.String(), func(t *testing.T) {
			if tt.format.ContentType() != tt.expectedContent {
				t.Errorf("ContentType() = %s, want %s", tt.format.ContentType(), tt.expectedContent)
			}
			if tt.format.FileExtension() != tt.expectedExtension {
				t.Errorf("FileExtension() = %s, want %s", tt.format.FileExtension(), tt.expectedExtension)
			}
		})
	}

	// Test unknown format
	unknownFormat := ExportFormat(99)
	if unknownFormat.ContentType() != "application/octet-stream" {
		t.Error("unknown format should return octet-stream content type")
	}
	if unknownFormat.FileExtension() != ".bin" {
		t.Error("unknown format should return .bin extension")
	}
}

func TestDisputeWorkflow_Validate(t *testing.T) {
	now := time.Now().UTC()
	blockHeight := int64(12345)
	initiator := testAddressForReconciliation(100)

	tests := []struct {
		name        string
		workflow    *DisputeWorkflow
		expectError bool
		errContains string
	}{
		{
			name: "valid workflow",
			workflow: NewDisputeWorkflow(
				"dispute-001",
				"invoice-001",
				initiator,
				DisputeCategoryBillingError,
				"Overcharged for CPU",
				"Detailed description of the billing error",
				sdk.NewCoins(sdk.NewInt64Coin("uvirt", 100000)),
				nil,
				nil,
				blockHeight,
				now,
			),
			expectError: false,
		},
		{
			name: "missing dispute_id",
			workflow: &DisputeWorkflow{
				InvoiceID:      "invoice-001",
				InitiatedBy:    initiator,
				Category:       DisputeCategoryBillingError,
				Subject:        "Test",
				Description:    "Test description",
				DisputedAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 100)),
				CreatedAt:      now,
				UpdatedAt:      now,
			},
			expectError: true,
			errContains: "dispute_id is required",
		},
		{
			name: "missing invoice_id",
			workflow: &DisputeWorkflow{
				DisputeID:      "dispute-001",
				InitiatedBy:    initiator,
				Category:       DisputeCategoryBillingError,
				Subject:        "Test",
				Description:    "Test description",
				DisputedAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 100)),
				CreatedAt:      now,
				UpdatedAt:      now,
			},
			expectError: true,
			errContains: "invoice_id is required",
		},
		{
			name: "invalid initiator address",
			workflow: &DisputeWorkflow{
				DisputeID:      "dispute-001",
				InvoiceID:      "invoice-001",
				InitiatedBy:    "invalid-address",
				Category:       DisputeCategoryBillingError,
				Subject:        "Test",
				Description:    "Test description",
				DisputedAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 100)),
				CreatedAt:      now,
				UpdatedAt:      now,
			},
			expectError: true,
			errContains: "invalid initiated_by address",
		},
		{
			name: "zero disputed amount",
			workflow: &DisputeWorkflow{
				DisputeID:      "dispute-001",
				InvoiceID:      "invoice-001",
				InitiatedBy:    initiator,
				Category:       DisputeCategoryBillingError,
				Subject:        "Test",
				Description:    "Test description",
				DisputedAmount: sdk.NewCoins(),
				CreatedAt:      now,
				UpdatedAt:      now,
			},
			expectError: true,
			errContains: "disputed_amount cannot be zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.workflow.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				} else if !containsString(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestDisputeEvidence_Validate(t *testing.T) {
	now := time.Now().UTC()
	uploader := testAddressForReconciliation(100)

	tests := []struct {
		name        string
		evidence    *DisputeEvidence
		expectError bool
		errContains string
	}{
		{
			name: "valid evidence",
			evidence: NewDisputeEvidence(
				"evidence-001",
				"dispute-001",
				DocumentEvidence,
				"Contract showing agreed pricing",
				"bafytest123",
				uploader,
				1024,
				"application/pdf",
				now,
			),
			expectError: false,
		},
		{
			name: "missing evidence_id",
			evidence: &DisputeEvidence{
				DisputeID:   "dispute-001",
				Type:        DocumentEvidence,
				Description: "Test",
				ArtifactCID: "bafytest123",
				UploadedBy:  uploader,
				UploadedAt:  now,
				FileSize:    1024,
				ContentType: "application/pdf",
			},
			expectError: true,
			errContains: "evidence_id is required",
		},
		{
			name: "missing dispute_id",
			evidence: &DisputeEvidence{
				EvidenceID:  "evidence-001",
				Type:        DocumentEvidence,
				Description: "Test",
				ArtifactCID: "bafytest123",
				UploadedBy:  uploader,
				UploadedAt:  now,
				FileSize:    1024,
				ContentType: "application/pdf",
			},
			expectError: true,
			errContains: "dispute_id is required",
		},
		{
			name: "missing artifact_cid",
			evidence: &DisputeEvidence{
				EvidenceID:  "evidence-001",
				DisputeID:   "dispute-001",
				Type:        DocumentEvidence,
				Description: "Test",
				UploadedBy:  uploader,
				UploadedAt:  now,
				FileSize:    1024,
				ContentType: "application/pdf",
			},
			expectError: true,
			errContains: "artifact_cid is required",
		},
		{
			name: "invalid file_size",
			evidence: &DisputeEvidence{
				EvidenceID:  "evidence-001",
				DisputeID:   "dispute-001",
				Type:        DocumentEvidence,
				Description: "Test",
				ArtifactCID: "bafytest123",
				UploadedBy:  uploader,
				UploadedAt:  now,
				FileSize:    0,
				ContentType: "application/pdf",
			},
			expectError: true,
			errContains: "file_size must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.evidence.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				} else if !containsString(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestDisputeRules_DefaultValues(t *testing.T) {
	rules := DefaultDisputeRules()

	if rules.MinDisputeAmount.Empty() {
		t.Error("MinDisputeAmount should not be empty")
	}

	if rules.MaxEvidenceCount == 0 {
		t.Error("MaxEvidenceCount should be positive")
	}

	if rules.MaxEvidenceSize <= 0 {
		t.Error("MaxEvidenceSize should be positive")
	}

	if len(rules.AllowedEvidenceTypes) == 0 {
		t.Error("AllowedEvidenceTypes should have values")
	}

	if rules.RequireEvidenceForAmount.Empty() {
		t.Error("RequireEvidenceForAmount should not be empty")
	}

	if rules.AutoEscalateAfterDays == 0 {
		t.Error("AutoEscalateAfterDays should be positive")
	}

	// Test validation
	if err := rules.Validate(); err != nil {
		t.Errorf("DefaultDisputeRules should be valid: %v", err)
	}

	// Test evidence type check
	for _, et := range rules.AllowedEvidenceTypes {
		if !rules.IsEvidenceTypeAllowed(et) {
			t.Errorf("evidence type %s should be allowed", et)
		}
	}
}

func TestReconciliationConfig_DefaultValues(t *testing.T) {
	config := DefaultReconciliationConfig()

	if config.VarianceThreshold.IsNegative() {
		t.Error("VarianceThreshold should not be negative")
	}

	if config.AutoResolveThreshold.IsNegative() {
		t.Error("AutoResolveThreshold should not be negative")
	}

	if config.MaxDiscrepanciesBeforeFail == 0 {
		t.Error("MaxDiscrepanciesBeforeFail should be positive")
	}

	if config.RequireManualReviewAbove.Empty() {
		t.Error("RequireManualReviewAbove should not be empty")
	}

	// AutoResolveThreshold should be less than or equal to VarianceThreshold
	if config.AutoResolveThreshold.GT(config.VarianceThreshold) {
		t.Error("AutoResolveThreshold should not exceed VarianceThreshold")
	}

	// Test validation
	if err := config.Validate(); err != nil {
		t.Errorf("DefaultReconciliationConfig should be valid: %v", err)
	}
}

func TestUsageRecord_Validate(t *testing.T) {
	now := time.Now().UTC()
	provider := testAddressForReconciliation(100)
	customer := testAddressForReconciliation(101)

	tests := []struct {
		name        string
		record      UsageRecord
		expectError bool
		errContains string
	}{
		{
			name: "valid record",
			record: UsageRecord{
				RecordID:    "usage-001",
				LeaseID:     "lease-001",
				Provider:    provider,
				Customer:    customer,
				StartTime:   now.Add(-1 * time.Hour),
				EndTime:     now,
				UsageAmount: sdkmath.LegacyNewDec(100),
				TotalAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 10000)),
			},
			expectError: false,
		},
		{
			name: "missing record_id",
			record: UsageRecord{
				LeaseID:     "lease-001",
				Provider:    provider,
				Customer:    customer,
				StartTime:   now.Add(-1 * time.Hour),
				EndTime:     now,
				UsageAmount: sdkmath.LegacyNewDec(100),
				TotalAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 10000)),
			},
			expectError: true,
			errContains: "record_id is required",
		},
		{
			name: "missing lease_id",
			record: UsageRecord{
				RecordID:    "usage-001",
				Provider:    provider,
				Customer:    customer,
				StartTime:   now.Add(-1 * time.Hour),
				EndTime:     now,
				UsageAmount: sdkmath.LegacyNewDec(100),
				TotalAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 10000)),
			},
			expectError: true,
			errContains: "lease_id is required",
		},
		{
			name: "invalid provider address",
			record: UsageRecord{
				RecordID:    "usage-001",
				LeaseID:     "lease-001",
				Provider:    "invalid-address",
				Customer:    customer,
				StartTime:   now.Add(-1 * time.Hour),
				EndTime:     now,
				UsageAmount: sdkmath.LegacyNewDec(100),
				TotalAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 10000)),
			},
			expectError: true,
			errContains: "invalid provider address",
		},
		{
			name: "end before start",
			record: UsageRecord{
				RecordID:    "usage-001",
				LeaseID:     "lease-001",
				Provider:    provider,
				Customer:    customer,
				StartTime:   now,
				EndTime:     now.Add(-1 * time.Hour),
				UsageAmount: sdkmath.LegacyNewDec(100),
				TotalAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 10000)),
			},
			expectError: true,
			errContains: "end_time must be after start_time",
		},
		{
			name: "negative usage_amount",
			record: UsageRecord{
				RecordID:    "usage-001",
				LeaseID:     "lease-001",
				Provider:    provider,
				Customer:    customer,
				StartTime:   now.Add(-1 * time.Hour),
				EndTime:     now,
				UsageAmount: sdkmath.LegacyNewDec(-100),
				TotalAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 10000)),
			},
			expectError: true,
			errContains: "usage_amount cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.record.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				} else if !containsString(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestPayoutRecord_Validate(t *testing.T) {
	now := time.Now().UTC()
	provider := testAddressForReconciliation(100)

	tests := []struct {
		name        string
		record      PayoutRecord
		expectError bool
		errContains string
	}{
		{
			name: "valid record",
			record: PayoutRecord{
				PayoutID:     "payout-001",
				Provider:     provider,
				PayoutAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 100000)),
				FeeAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
				NetAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 99000)),
				InvoiceIDs:   []string{"invoice-001"},
				PayoutDate:   now,
			},
			expectError: false,
		},
		{
			name: "missing payout_id",
			record: PayoutRecord{
				Provider:     provider,
				PayoutAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 100000)),
				FeeAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
				NetAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 99000)),
				InvoiceIDs:   []string{"invoice-001"},
				PayoutDate:   now,
			},
			expectError: true,
			errContains: "payout_id is required",
		},
		{
			name: "invalid provider address",
			record: PayoutRecord{
				PayoutID:     "payout-001",
				Provider:     "invalid-address",
				PayoutAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 100000)),
				FeeAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
				NetAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 99000)),
				InvoiceIDs:   []string{"invoice-001"},
				PayoutDate:   now,
			},
			expectError: true,
			errContains: "invalid provider address",
		},
		{
			name: "no invoice_ids",
			record: PayoutRecord{
				PayoutID:     "payout-001",
				Provider:     provider,
				PayoutAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 100000)),
				FeeAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
				NetAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 99000)),
				InvoiceIDs:   []string{},
				PayoutDate:   now,
			},
			expectError: true,
			errContains: "at least one invoice_id is required",
		},
		{
			name: "missing payout_date",
			record: PayoutRecord{
				PayoutID:     "payout-001",
				Provider:     provider,
				PayoutAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 100000)),
				FeeAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
				NetAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 99000)),
				InvoiceIDs:   []string{"invoice-001"},
			},
			expectError: true,
			errContains: "payout_date is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.record.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				} else if !containsString(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestVarianceCalculation(t *testing.T) {
	// Test CalculateVariancePercentage
	tests := []struct {
		name             string
		expected         sdk.Coin
		actual           sdk.Coin
		expectedVariance sdkmath.LegacyDec
	}{
		{
			name:             "no variance",
			expected:         sdk.NewInt64Coin("uvirt", 1000),
			actual:           sdk.NewInt64Coin("uvirt", 1000),
			expectedVariance: sdkmath.LegacyZeroDec(),
		},
		{
			name:             "10% undercharge",
			expected:         sdk.NewInt64Coin("uvirt", 1000),
			actual:           sdk.NewInt64Coin("uvirt", 900),
			expectedVariance: sdkmath.LegacyNewDecWithPrec(1, 1), // 0.1 = 10%
		},
		{
			name:             "10% overcharge",
			expected:         sdk.NewInt64Coin("uvirt", 1000),
			actual:           sdk.NewInt64Coin("uvirt", 1100),
			expectedVariance: sdkmath.LegacyNewDecWithPrec(1, 1), // 0.1 = 10%
		},
		{
			name:             "both zero",
			expected:         sdk.NewInt64Coin("uvirt", 0),
			actual:           sdk.NewInt64Coin("uvirt", 0),
			expectedVariance: sdkmath.LegacyZeroDec(),
		},
		{
			name:             "expected zero actual non-zero",
			expected:         sdk.NewInt64Coin("uvirt", 0),
			actual:           sdk.NewInt64Coin("uvirt", 100),
			expectedVariance: sdkmath.LegacyOneDec(), // 100%
		},
		{
			name:             "different denoms",
			expected:         sdk.NewInt64Coin("uvirt", 1000),
			actual:           sdk.NewInt64Coin("uatom", 1000),
			expectedVariance: sdkmath.LegacyOneDec(), // 100% due to incompatible denoms
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			variance := CalculateVariancePercentage(tt.expected, tt.actual)
			if !variance.Equal(tt.expectedVariance) {
				t.Errorf("CalculateVariancePercentage(%s, %s) = %s, want %s",
					tt.expected, tt.actual, variance, tt.expectedVariance)
			}
		})
	}

	// Test IsVarianceWithinThreshold
	threshold := sdkmath.LegacyNewDecWithPrec(5, 2) // 5%

	withinThreshold := sdkmath.LegacyNewDecWithPrec(3, 2) // 3%
	if !IsVarianceWithinThreshold(withinThreshold, threshold) {
		t.Error("3% variance should be within 5% threshold")
	}

	exactThreshold := sdkmath.LegacyNewDecWithPrec(5, 2) // 5%
	if !IsVarianceWithinThreshold(exactThreshold, threshold) {
		t.Error("5% variance should be within 5% threshold (equal)")
	}

	exceedsThreshold := sdkmath.LegacyNewDecWithPrec(7, 2) // 7%
	if IsVarianceWithinThreshold(exceedsThreshold, threshold) {
		t.Error("7% variance should not be within 5% threshold")
	}
}

func TestReconciliationKeyBuilders(t *testing.T) {
	// Test ReconciliationReportKey
	reportKey := BuildReconciliationReportKey("report-001")
	if len(reportKey) <= len(ReconciliationReportPrefix) {
		t.Error("report key should be longer than prefix")
	}
	if !bytes.HasPrefix(reportKey, ReconciliationReportPrefix) {
		t.Error("report key should have correct prefix")
	}

	// Test ReconciliationReportByProviderKey
	providerKey := BuildReconciliationReportByProviderKey("provider-addr", "report-001")
	if len(providerKey) <= len(ReconciliationReportByProviderPrefix) {
		t.Error("provider key should be longer than prefix")
	}
	if !bytes.HasPrefix(providerKey, ReconciliationReportByProviderPrefix) {
		t.Error("provider key should have correct prefix")
	}

	// Test ReconciliationReportByCustomerKey
	customerKey := BuildReconciliationReportByCustomerKey("customer-addr", "report-001")
	if len(customerKey) <= len(ReconciliationReportByCustomerPrefix) {
		t.Error("customer key should be longer than prefix")
	}
	if !bytes.HasPrefix(customerKey, ReconciliationReportByCustomerPrefix) {
		t.Error("customer key should have correct prefix")
	}

	// Test CorrectionKey
	correctionKey := BuildCorrectionKey("corr-001")
	if len(correctionKey) <= len(CorrectionPrefix) {
		t.Error("correction key should be longer than prefix")
	}
	if !bytes.HasPrefix(correctionKey, CorrectionPrefix) {
		t.Error("correction key should have correct prefix")
	}

	// Test ParseCorrectionKey
	parsedCorrectionID, err := ParseCorrectionKey(correctionKey)
	if err != nil {
		t.Errorf("ParseCorrectionKey failed: %v", err)
	}
	if parsedCorrectionID != "corr-001" {
		t.Errorf("expected correction ID corr-001, got %s", parsedCorrectionID)
	}

	// Test CorrectionByInvoiceKey
	corrInvoiceKey := BuildCorrectionByInvoiceKey("invoice-001", "corr-001")
	if !bytes.HasPrefix(corrInvoiceKey, CorrectionByInvoicePrefix) {
		t.Error("correction by invoice key should have correct prefix")
	}

	// Test CorrectionByStatusKey
	corrStatusKey := BuildCorrectionByStatusKey(CorrectionStatusPending, "corr-001")
	if !bytes.HasPrefix(corrStatusKey, CorrectionByStatusPrefix) {
		t.Error("correction by status key should have correct prefix")
	}

	// Test AuditEntryKey
	auditKey := BuildAuditEntryKey("audit-001")
	if len(auditKey) <= len(AuditEntryPrefix) {
		t.Error("audit key should be longer than prefix")
	}
	if !bytes.HasPrefix(auditKey, AuditEntryPrefix) {
		t.Error("audit key should have correct prefix")
	}

	// Test ParseAuditEntryKey
	parsedAuditID, err := ParseAuditEntryKey(auditKey)
	if err != nil {
		t.Errorf("ParseAuditEntryKey failed: %v", err)
	}
	if parsedAuditID != "audit-001" {
		t.Errorf("expected audit ID audit-001, got %s", parsedAuditID)
	}

	// Test AlertKey
	alertKey := BuildAlertKey("alert-001")
	if !bytes.HasPrefix(alertKey, AlertPrefix) {
		t.Error("alert key should have correct prefix")
	}

	// Test AlertRuleKey
	ruleKey := BuildAlertRuleKey("rule-001")
	if !bytes.HasPrefix(ruleKey, AlertRulePrefix) {
		t.Error("alert rule key should have correct prefix")
	}

	// Test ExportRequestKey
	exportKey := BuildExportRequestKey("export-001")
	if !bytes.HasPrefix(exportKey, ExportRequestPrefix) {
		t.Error("export request key should have correct prefix")
	}

	// Test DisputeWorkflowKey
	disputeKey := BuildDisputeWorkflowKey("dispute-001")
	if !bytes.HasPrefix(disputeKey, DisputeWorkflowPrefix) {
		t.Error("dispute workflow key should have correct prefix")
	}

	// Test ParseDisputeWorkflowKey
	parsedDisputeID, err := ParseDisputeWorkflowKey(disputeKey)
	if err != nil {
		t.Errorf("ParseDisputeWorkflowKey failed: %v", err)
	}
	if parsedDisputeID != "dispute-001" {
		t.Errorf("expected dispute ID dispute-001, got %s", parsedDisputeID)
	}

	// Test UsageRecordKey
	usageKey := BuildUsageRecordKey("usage-001")
	if !bytes.HasPrefix(usageKey, UsageRecordPrefix) {
		t.Error("usage record key should have correct prefix")
	}

	// Test PayoutRecordKey
	payoutKey := BuildPayoutRecordKey("payout-001")
	if !bytes.HasPrefix(payoutKey, PayoutRecordPrefix) {
		t.Error("payout record key should have correct prefix")
	}

	// Test error cases for parse functions
	_, err = ParseCorrectionKey(CorrectionPrefix)
	if err == nil {
		t.Error("ParseCorrectionKey should fail for key equal to prefix")
	}

	_, err = ParseAuditEntryKey(AuditEntryPrefix)
	if err == nil {
		t.Error("ParseAuditEntryKey should fail for key equal to prefix")
	}

	_, err = ParseDisputeWorkflowKey(DisputeWorkflowPrefix)
	if err == nil {
		t.Error("ParseDisputeWorkflowKey should fail for key equal to prefix")
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
