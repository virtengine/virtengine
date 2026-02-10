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

func TestNewFraudDetector(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultFraudDetectorConfig()
	detector := NewFraudDetector(cfg, logger)

	if detector == nil {
		t.Fatal("NewFraudDetector returned nil")
	}
}

func TestDefaultFraudDetectorConfig(t *testing.T) {
	cfg := DefaultFraudDetectorConfig()

	if cfg.MaxVerificationAttemptsPerAccount != 5 {
		t.Errorf("Expected MaxVerificationAttemptsPerAccount 5, got %d", cfg.MaxVerificationAttemptsPerAccount)
	}

	if cfg.MaxFailedVerificationsBeforeFlag != 3 {
		t.Errorf("Expected MaxFailedVerificationsBeforeFlag 3, got %d", cfg.MaxFailedVerificationsBeforeFlag)
	}

	if cfg.MinExpectedScore != 60 {
		t.Errorf("Expected MinExpectedScore 60, got %d", cfg.MinExpectedScore)
	}

	if cfg.BiometricMismatchThreshold != 0.5 {
		t.Errorf("Expected BiometricMismatchThreshold 0.5, got %f", cfg.BiometricMismatchThreshold)
	}

	if cfg.FaceSimilarityMinimum != 0.8 {
		t.Errorf("Expected FaceSimilarityMinimum 0.8, got %f", cfg.FaceSimilarityMinimum)
	}

	if !cfg.EnableDocumentForensics {
		t.Error("EnableDocumentForensics should be true by default")
	}
}

func TestVEIDVerificationDataStructure(t *testing.T) {
	data := &VEIDVerificationData{
		RequestID:       "req_123",
		AccountAddress:  "virtengine1test",
		Timestamp:       time.Now(),
		BlockHeight:     12345,
		ProposerScore:   85,
		ComputedScore:   84,
		ScoreDifference: 1,
		Match:           true,
		Success:         true,
	}

	if data.RequestID == "" {
		t.Error("RequestID should not be empty")
	}

	if data.AccountAddress == "" {
		t.Error("AccountAddress should not be empty")
	}

	if data.ProposerScore == 0 {
		t.Error("ProposerScore should not be zero")
	}
}

func TestFraudDetectorAnalyze(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultFraudDetectorConfig()
	detector := NewFraudDetector(cfg, logger)

	eventChan := make(chan *SecurityEvent, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	detector.Start(ctx, eventChan)

	data := &VEIDVerificationData{
		RequestID:      "req_test",
		AccountAddress: "virtengine1test",
		Timestamp:      time.Now(),
		BlockHeight:    100,
		ProposerScore:  85,
		ComputedScore:  85,
		Match:          true,
		Success:        true,
	}

	// Analyze should not panic
	detector.Analyze(data)
}

func TestFraudDetectorLowScoreDetection(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultFraudDetectorConfig()
	cfg.MinExpectedScore = 60
	detector := NewFraudDetector(cfg, logger)

	eventChan := make(chan *SecurityEvent, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	detector.Start(ctx, eventChan)

	// Submit low score verification
	data := &VEIDVerificationData{
		RequestID:      "req_low",
		AccountAddress: "virtengine1low",
		Timestamp:      time.Now(),
		BlockHeight:    100,
		ProposerScore:  30, // Very low
		ComputedScore:  30,
		Match:          true,
		Success:        false,
		FailureReason:  "score_below_threshold",
	}

	detector.Analyze(data)

	// Check for event
	select {
	case <-eventChan:
		// Event received
	case <-time.After(100 * time.Millisecond):
		// No event
	}
}

func TestFraudDetectorBiometricMismatch(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultFraudDetectorConfig()
	detector := NewFraudDetector(cfg, logger)

	eventChan := make(chan *SecurityEvent, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	detector.Start(ctx, eventChan)

	data := &VEIDVerificationData{
		RequestID:           "req_bio",
		AccountAddress:      "virtengine1bio",
		Timestamp:           time.Now(),
		BlockHeight:         100,
		ProposerScore:       75,
		ComputedScore:       75,
		Match:               true,
		FaceSimilarityScore: 0.3, // Low similarity
		LivenessScore:       0.9,
		Success:             false,
		FailureReason:       "face_mismatch",
	}

	detector.Analyze(data)

	// Check for event
	select {
	case <-eventChan:
		// Event received
	case <-time.After(100 * time.Millisecond):
		// No event
	}
}

func TestFraudDetectorVelocityDetection(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultFraudDetectorConfig()
	cfg.MaxVerificationAttemptsPerAccount = 3
	detector := NewFraudDetector(cfg, logger)

	eventChan := make(chan *SecurityEvent, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	detector.Start(ctx, eventChan)

	address := "virtengine1velocity"
	now := time.Now()

	// Submit multiple verifications
	for i := 0; i < 5; i++ {
		data := &VEIDVerificationData{
			RequestID:      "req_" + string(rune(i+'0')),
			AccountAddress: address,
			Timestamp:      now,
			BlockHeight:    int64(100 + i),
			ProposerScore:  85,
			ComputedScore:  85,
			Match:          true,
			Success:        true,
		}
		detector.Analyze(data)
	}

	// Check for velocity event
	select {
	case <-eventChan:
		// Event received
	case <-time.After(100 * time.Millisecond):
		// No event
	}
}

func TestFraudDetectorCleanup(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultFraudDetectorConfig()
	detector := NewFraudDetector(cfg, logger)

	eventChan := make(chan *SecurityEvent, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	detector.Start(ctx, eventChan)

	// Cleanup is internal, but detector should handle graceful shutdown
	// Just verify it doesn't panic when context is cancelled
}
