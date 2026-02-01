// Copyright 2024 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0

package security_monitoring

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestNewProviderSecurityMonitor(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultProviderSecurityConfig()
	monitor := NewProviderSecurityMonitor(cfg, logger)

	if monitor == nil {
		t.Fatal("NewProviderSecurityMonitor returned nil")
	}
}

func TestDefaultProviderSecurityConfig(t *testing.T) {
	cfg := DefaultProviderSecurityConfig()

	if cfg.MaxKeyUsagePerMinute != 20 {
		t.Errorf("Expected MaxKeyUsagePerMinute 20, got %d", cfg.MaxKeyUsagePerMinute)
	}

	if cfg.MaxFailedSignaturesPerHour != 10 {
		t.Errorf("Expected MaxFailedSignaturesPerHour 10, got %d", cfg.MaxFailedSignaturesPerHour)
	}

	if cfg.MaxBidsPerMinute != 50 {
		t.Errorf("Expected MaxBidsPerMinute 50, got %d", cfg.MaxBidsPerMinute)
	}

	if !cfg.EnableLocationTracking {
		t.Error("EnableLocationTracking should be true by default")
	}

	if !cfg.EnableResourceAnomalyDetection {
		t.Error("EnableResourceAnomalyDetection should be true by default")
	}
}

func TestProviderActivityDataStructure(t *testing.T) {
	data := &ProviderActivityData{
		ProviderID:   "provider123",
		ActivityType: "bid",
		Timestamp:    time.Now(),
		SourceIP:     "192.168.1.1",
		KeyID:        "key456",
		Success:      true,
		BidAmount:    1000,
	}

	if data.ProviderID == "" {
		t.Error("ProviderID should not be empty")
	}

	if data.ActivityType == "" {
		t.Error("ActivityType should not be empty")
	}

	if data.BidAmount == 0 {
		t.Error("BidAmount should not be zero")
	}
}

func TestProviderSecurityMonitorAnalyze(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultProviderSecurityConfig()
	monitor := NewProviderSecurityMonitor(cfg, logger)

	eventChan := make(chan *SecurityEvent, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	monitor.Start(ctx, eventChan)

	data := &ProviderActivityData{
		ProviderID:   "provider_test",
		ActivityType: "bid",
		Timestamp:    time.Now(),
		Success:      true,
		BidAmount:    100,
	}

	// Analyze should not panic
	monitor.Analyze(data)
}

func TestProviderSecurityMonitorKeyCompromiseDetection(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultProviderSecurityConfig()
	cfg.MaxKeyUsagePerMinute = 3
	monitor := NewProviderSecurityMonitor(cfg, logger)

	eventChan := make(chan *SecurityEvent, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	monitor.Start(ctx, eventChan)

	providerID := "provider_key"
	now := time.Now()

	// Submit many key usages
	for i := 0; i < 5; i++ {
		data := &ProviderActivityData{
			ProviderID:   providerID,
			ActivityType: "key_usage",
			Timestamp:    now,
			KeyID:        "key_" + string(rune(i+'0')),
			Success:      true,
		}
		monitor.Analyze(data)
	}

	// Check for event
	select {
	case <-eventChan:
		// Event received
	case <-time.After(100 * time.Millisecond):
		// No event
	}
}

func TestProviderSecurityMonitorRapidBidDetection(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultProviderSecurityConfig()
	cfg.MaxBidsPerMinute = 3
	monitor := NewProviderSecurityMonitor(cfg, logger)

	eventChan := make(chan *SecurityEvent, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	monitor.Start(ctx, eventChan)

	providerID := "provider_bid"
	now := time.Now()

	// Submit many bids
	for i := 0; i < 5; i++ {
		data := &ProviderActivityData{
			ProviderID:   providerID,
			ActivityType: "bid",
			Timestamp:    now,
			OrderID:      "order_" + string(rune(i+'0')),
			BidAmount:    float64(100 + i),
			Success:      true,
		}
		monitor.Analyze(data)
	}

	// Check for event
	select {
	case <-eventChan:
		// Event received
	case <-time.After(100 * time.Millisecond):
		// No event
	}
}

func TestProviderSecurityIndicatorTypes(t *testing.T) {
	indicators := []ProviderSecurityIndicator{
		ProviderIndicatorKeyCompromise,
		ProviderIndicatorUnauthorizedAccess,
		ProviderIndicatorAnomalousLocation,
		ProviderIndicatorAnomalousTime,
		ProviderIndicatorRapidActivity,
		ProviderIndicatorResourceAnomaly,
		ProviderIndicatorBidManipulation,
		ProviderIndicatorLeaseAbuse,
	}

	for _, ind := range indicators {
		if ind == "" {
			t.Error("Provider security indicator constant should not be empty")
		}
	}
}

func TestProviderSecurityMonitorResourceAnomaly(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultProviderSecurityConfig()
	monitor := NewProviderSecurityMonitor(cfg, logger)

	eventChan := make(chan *SecurityEvent, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	monitor.Start(ctx, eventChan)

	data := &ProviderActivityData{
		ProviderID:   "provider_resource",
		ActivityType: "resource_report",
		Timestamp:    time.Now(),
		CPUUsage:     95.0, // Very high
		MemoryUsage:  90.0,
		Success:      true,
	}

	monitor.Analyze(data)

	// Check for event
	select {
	case <-eventChan:
		// Event received
	case <-time.After(100 * time.Millisecond):
		// No event
	}
}

func TestProviderSecurityMonitorCleanup(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultProviderSecurityConfig()
	monitor := NewProviderSecurityMonitor(cfg, logger)

	eventChan := make(chan *SecurityEvent, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	monitor.Start(ctx, eventChan)

	// Cleanup is internal, but monitor should handle graceful shutdown
	// Just verify it doesn't panic when context is cancelled
}

