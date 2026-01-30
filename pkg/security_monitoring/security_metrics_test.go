// Copyright 2024 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0

package security_monitoring

import (
	"testing"
)

func TestGetSecurityMetrics(t *testing.T) {
	metrics := GetSecurityMetrics()
	if metrics == nil {
		t.Fatal("GetSecurityMetrics returned nil")
	}

	// Test singleton pattern
	metrics2 := GetSecurityMetrics()
	if metrics != metrics2 {
		t.Error("GetSecurityMetrics should return the same instance")
	}
}

func TestSecurityMetricsFields(t *testing.T) {
	metrics := GetSecurityMetrics()

	// Check that critical metric fields are not nil
	if metrics.TxAnomaliesDetected == nil {
		t.Error("TxAnomaliesDetected should not be nil")
	}

	if metrics.VEIDFraudIndicators == nil {
		t.Error("VEIDFraudIndicators should not be nil")
	}

	if metrics.RateLimitBreaches == nil {
		t.Error("RateLimitBreaches should not be nil")
	}

	if metrics.CryptoSignatureFailures == nil {
		t.Error("CryptoSignatureFailures should not be nil")
	}

	if metrics.ProviderCompromiseIndicators == nil {
		t.Error("ProviderCompromiseIndicators should not be nil")
	}

	if metrics.SecurityIncidentsActive == nil {
		t.Error("SecurityIncidentsActive should not be nil")
	}

	if metrics.ThreatLevel == nil {
		t.Error("ThreatLevel should not be nil")
	}

	if metrics.SecurityScore == nil {
		t.Error("SecurityScore should not be nil")
	}
}

func TestSecurityEventSeverities(t *testing.T) {
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

func TestSecurityEventSeverityValues(t *testing.T) {
	if SeverityCritical != "critical" {
		t.Error("SeverityCritical should be 'critical'")
	}
	if SeverityHigh != "high" {
		t.Error("SeverityHigh should be 'high'")
	}
	if SeverityMedium != "medium" {
		t.Error("SeverityMedium should be 'medium'")
	}
	if SeverityLow != "low" {
		t.Error("SeverityLow should be 'low'")
	}
}

func TestMetricsLabels(t *testing.T) {
	metrics := GetSecurityMetrics()

	// Test that we can use the counter vectors with labels
	// This should not panic
	metrics.TxAnomaliesDetected.WithLabelValues("velocity", "high").Add(0)
	metrics.VEIDFraudIndicators.WithLabelValues("document_tampering", "high").Add(0)
	metrics.RateLimitBreaches.WithLabelValues("per_ip", "high").Add(0)
}

func TestMetricsGauges(t *testing.T) {
	metrics := GetSecurityMetrics()

	// Test that we can set gauge values
	// This should not panic
	metrics.SecurityIncidentsActive.Set(0)
	metrics.ThreatLevel.Set(0)
	metrics.SecurityScore.Set(100)
}
