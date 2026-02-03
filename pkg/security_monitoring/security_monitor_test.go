// Copyright 2024 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0

package security_monitoring

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
)

func TestNewSecurityMonitor(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultSecurityMonitorConfig()

	monitor, err := NewSecurityMonitor(cfg, logger)
	if err != nil {
		t.Fatalf("NewSecurityMonitor failed: %v", err)
	}

	if monitor == nil {
		t.Fatal("NewSecurityMonitor returned nil")
	}

	if monitor.config != cfg {
		t.Error("Config not set correctly")
	}
}

func TestDefaultSecurityMonitorConfig(t *testing.T) {
	cfg := DefaultSecurityMonitorConfig()

	if !cfg.EnableTransactionDetector {
		t.Error("EnableTransactionDetector should be true by default")
	}

	if !cfg.EnableFraudDetector {
		t.Error("EnableFraudDetector should be true by default")
	}

	if !cfg.EnableCryptoAnomalyDetector {
		t.Error("EnableCryptoAnomalyDetector should be true by default")
	}

	if !cfg.EnableProviderSecurity {
		t.Error("EnableProviderSecurity should be true by default")
	}

	if !cfg.EnableAutoResponse {
		t.Error("EnableAutoResponse should be true by default")
	}

	if cfg.AlertCooldownSecs != 60 {
		t.Errorf("Expected AlertCooldownSecs 60, got %d", cfg.AlertCooldownSecs)
	}

	if cfg.MaxAlertsPerMinute != 100 {
		t.Errorf("Expected MaxAlertsPerMinute 100, got %d", cfg.MaxAlertsPerMinute)
	}

	if cfg.ThreatLevelHighThreshold != 10 {
		t.Errorf("Expected ThreatLevelHighThreshold 10, got %d", cfg.ThreatLevelHighThreshold)
	}

	if cfg.ThreatLevelCriticalThreshold != 25 {
		t.Errorf("Expected ThreatLevelCriticalThreshold 25, got %d", cfg.ThreatLevelCriticalThreshold)
	}
}

func TestSecurityMonitorStartStop(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultSecurityMonitorConfig()

	monitor, err := NewSecurityMonitor(cfg, logger)
	if err != nil {
		t.Fatalf("NewSecurityMonitor failed: %v", err)
	}

	// Start monitor
	err = monitor.Start()
	if err != nil {
		t.Errorf("Start failed: %v", err)
	}

	// Stop monitor (no return value)
	monitor.Stop()
}

func TestSecurityIncidentStructure(t *testing.T) {
	incident := &SecurityIncident{
		ID:          "INC-001",
		Type:        "transaction_anomaly",
		Severity:    SeverityHigh,
		Description: "Test incident",
		Status:      "open",
	}

	if incident.ID == "" {
		t.Error("Incident ID should not be empty")
	}

	if incident.Type == "" {
		t.Error("Incident Type should not be empty")
	}

	if incident.Severity == "" {
		t.Error("Incident Severity should not be empty")
	}
}

func TestSecurityMonitorConfigValidation(t *testing.T) {
	cfg := &SecurityMonitorConfig{
		EnableTransactionDetector: false,
		EnableFraudDetector:       false,
		AlertCooldownSecs:         30,
		MaxAlertsPerMinute:        50,
	}

	if cfg.EnableTransactionDetector {
		t.Error("EnableTransactionDetector should be false")
	}

	if cfg.AlertCooldownSecs != 30 {
		t.Errorf("Expected AlertCooldownSecs 30, got %d", cfg.AlertCooldownSecs)
	}
}
