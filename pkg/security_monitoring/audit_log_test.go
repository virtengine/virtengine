// Copyright 2024 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0

package security_monitoring

import (
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestNewAuditLog(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	// Test without file path
	auditLog, err := NewAuditLog("", logger)
	if err != nil {
		t.Errorf("NewAuditLog failed: %v", err)
	}
	if auditLog == nil {
		t.Fatal("NewAuditLog returned nil")
	}
	auditLog.Close()
}

func TestNewAuditLogWithFile(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	// Create temp file
	tmpFile, err := os.CreateTemp("", "audit-test-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	os.Remove(tmpPath)
	defer os.Remove(tmpPath)

	auditLog, err := NewAuditLog(tmpPath, logger)
	if err != nil {
		t.Errorf("NewAuditLog with file failed: %v", err)
	}
	if auditLog == nil {
		t.Fatal("NewAuditLog returned nil")
	}

	err = auditLog.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestAuditLogLogEvent(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	auditLog, err := NewAuditLog("", logger)
	if err != nil {
		t.Fatalf("NewAuditLog failed: %v", err)
	}
	defer auditLog.Close()

	event := &SecurityEvent{
		ID:             generateEventID(),
		Type:           "test-event",
		Severity:       SeverityMedium,
		Timestamp:      time.Now(),
		Source:         "test",
		Description:    "Test security event",
		Metadata:       map[string]interface{}{"key": "value"},
		AccountAddress: "virtengine1test",
		BlockHeight:    12345,
	}

	// Should not panic
	auditLog.LogEvent(event)
}

func TestAuditLogLogAlert(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	auditLog, err := NewAuditLog("", logger)
	if err != nil {
		t.Fatalf("NewAuditLog failed: %v", err)
	}
	defer auditLog.Close()

	alert := &SecurityAlert{
		ID:          generateAlertID(),
		EventID:     generateEventID(),
		Type:        "test-alert",
		Severity:    SeverityHigh,
		Title:       "Test Alert",
		Description: "Test security alert",
		Source:      "test",
		Timestamp:   time.Now(),
	}

	// Should not panic
	auditLog.LogAlert(alert)
}

func TestAuditLogLogIncidentAction(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	auditLog, err := NewAuditLog("", logger)
	if err != nil {
		t.Fatalf("NewAuditLog failed: %v", err)
	}
	defer auditLog.Close()

	// Should not panic
	auditLog.LogIncidentAction("INC-001", "block_ip", "system", "success")
}

func TestAuditLogLogPlaybookExecution(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	auditLog, err := NewAuditLog("", logger)
	if err != nil {
		t.Fatalf("NewAuditLog failed: %v", err)
	}
	defer auditLog.Close()

	steps := []string{"step1", "step2", "step3"}
	duration := 5 * time.Second

	// Should not panic
	auditLog.LogPlaybookExecution("playbook-1", "INC-001", "success", steps, duration)
}

func TestSecurityEventStructure(t *testing.T) {
	event := &SecurityEvent{
		ID:              "evt_123",
		Type:            "suspicious_transaction",
		Severity:        SeverityCritical,
		Timestamp:       time.Now(),
		Source:          "transaction_detector",
		Description:     "High value transaction detected",
		Category:        "transaction",
		Subcategory:     "value_anomaly",
		AccountAddress:  "virtengine1abc",
		ProviderID:      "provider123",
		BlockHeight:     100000,
		TransactionHash: "ABCD1234",
		ActionsTaken:    []string{"alert", "log"},
	}

	if event.ID == "" {
		t.Error("Event ID should not be empty")
	}

	if event.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}

	if event.Severity == "" {
		t.Error("Severity should not be empty")
	}
}

func TestSecurityAlertStructure(t *testing.T) {
	now := time.Now()
	alert := &SecurityAlert{
		ID:             "alt_123",
		EventID:        "evt_456",
		Type:           "fraud_detected",
		Severity:       SeverityHigh,
		Title:          "Fraud Detected",
		Description:    "Identity fraud detected",
		Source:         "fraud_detector",
		Timestamp:      now,
		Acknowledged:   true,
		AcknowledgedAt: &now,
		AcknowledgedBy: "operator1",
	}

	if alert.ID == "" {
		t.Error("Alert ID should not be empty")
	}

	if !alert.Acknowledged {
		t.Error("Alert should be acknowledged")
	}

	if alert.AcknowledgedAt == nil {
		t.Error("AcknowledgedAt should not be nil when acknowledged")
	}
}

func TestRateLimitBreachDataStructure(t *testing.T) {
	breach := &RateLimitBreachData{
		LimitType:     "per_ip",
		SourceIP:      "192.168.1.1",
		CurrentCount:  150,
		Limit:         100,
		BypassAttempt: true,
		Description:   "Rate limit exceeded",
		Severity:      SeverityMedium,
	}

	if breach.LimitType == "" {
		t.Error("LimitType should not be empty")
	}

	if breach.CurrentCount <= breach.Limit {
		t.Error("CurrentCount should exceed Limit for a breach")
	}
}

func TestAuditLogEntryStructure(t *testing.T) {
	entry := AuditLogEntry{
		Timestamp:   time.Now().Format(time.RFC3339Nano),
		Level:       "warn",
		LogType:     "security_audit",
		EventID:     "evt_123",
		EventType:   "suspicious_activity",
		Severity:    "high",
		Source:      "monitor",
		Description: "Suspicious activity detected",
		Account:     "virtengine1test",
		BlockHeight: 12345,
	}

	if entry.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}

	if entry.LogType != "security_audit" {
		t.Error("LogType should be 'security_audit'")
	}
}

func TestSeverityLogLevelMapping(t *testing.T) {
	// Test the severity levels work in events
	severities := []SecurityEventSeverity{
		SeverityCritical,
		SeverityHigh,
		SeverityMedium,
		SeverityLow,
	}

	for _, sev := range severities {
		if sev == "" {
			t.Error("Severity constant should not be empty")
		}
	}
}

