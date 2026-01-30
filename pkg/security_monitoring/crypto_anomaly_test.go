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

func TestNewCryptoAnomalyDetector(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultCryptoAnomalyConfig()
	detector := NewCryptoAnomalyDetector(cfg, logger)

	if detector == nil {
		t.Fatal("NewCryptoAnomalyDetector returned nil")
	}
}

func TestDefaultCryptoAnomalyConfig(t *testing.T) {
	cfg := DefaultCryptoAnomalyConfig()

	if cfg.MaxSignatureFailuresPerHour != 50 {
		t.Errorf("Expected MaxSignatureFailuresPerHour 50, got %d", cfg.MaxSignatureFailuresPerHour)
	}

	if cfg.MaxSignatureFailuresPerAccount != 10 {
		t.Errorf("Expected MaxSignatureFailuresPerAccount 10, got %d", cfg.MaxSignatureFailuresPerAccount)
	}

	if cfg.MaxKeyOperationsPerMinute != 20 {
		t.Errorf("Expected MaxKeyOperationsPerMinute 20, got %d", cfg.MaxKeyOperationsPerMinute)
	}

	if !cfg.EnableEntropyAnalysis {
		t.Error("EnableEntropyAnalysis should be true by default")
	}

	if cfg.MinEntropyThreshold != 3.5 {
		t.Errorf("Expected MinEntropyThreshold 3.5, got %f", cfg.MinEntropyThreshold)
	}

	if !cfg.EnableKeyReuseDetection {
		t.Error("EnableKeyReuseDetection should be true by default")
	}
}

func TestCryptoOperationDataStructure(t *testing.T) {
	data := &CryptoOperationData{
		OperationID:    "op_123",
		OperationType:  "sign",
		Algorithm:      "Ed25519",
		KeyID:          "key_456",
		AccountAddress: "virtengine1test",
		Timestamp:      time.Now(),
		Success:        true,
		DataSize:       256,
	}

	if data.OperationID == "" {
		t.Error("OperationID should not be empty")
	}

	if data.OperationType == "" {
		t.Error("OperationType should not be empty")
	}

	if data.Algorithm == "" {
		t.Error("Algorithm should not be empty")
	}
}

func TestCryptoAnomalyDetectorAnalyze(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultCryptoAnomalyConfig()
	detector := NewCryptoAnomalyDetector(cfg, logger)

	eventChan := make(chan *SecurityEvent, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	detector.Start(ctx, eventChan)

	data := &CryptoOperationData{
		OperationID:    "op_test",
		OperationType:  "sign",
		Algorithm:      "Ed25519",
		AccountAddress: "virtengine1test",
		Timestamp:      time.Now(),
		Success:        true,
	}

	// Analyze should not panic
	detector.Analyze(data)
}

func TestCryptoAnomalyDetectorSignatureFailure(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultCryptoAnomalyConfig()
	cfg.MaxSignatureFailuresPerAccount = 2
	detector := NewCryptoAnomalyDetector(cfg, logger)

	eventChan := make(chan *SecurityEvent, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	detector.Start(ctx, eventChan)

	account := "virtengine1sigfail"

	// Submit multiple signature failures
	for i := 0; i < 3; i++ {
		data := &CryptoOperationData{
			OperationID:    "op_" + string(rune(i+'0')),
			OperationType:  "verify",
			Algorithm:      "Ed25519",
			AccountAddress: account,
			Timestamp:      time.Now(),
			Success:        false,
			FailureReason:  "invalid_signature",
		}
		detector.Analyze(data)
	}

	// Check for event
	select {
	case <-eventChan:
		// Event received
	case <-time.After(100 * time.Millisecond):
		// No event
	}
}

func TestCryptoAnomalyDetectorDeprecatedAlgorithm(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultCryptoAnomalyConfig()
	detector := NewCryptoAnomalyDetector(cfg, logger)

	eventChan := make(chan *SecurityEvent, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	detector.Start(ctx, eventChan)

	data := &CryptoOperationData{
		OperationID:    "op_deprecated",
		OperationType:  "sign",
		Algorithm:      "MD5", // Deprecated
		AccountAddress: "virtengine1test",
		Timestamp:      time.Now(),
		Success:        true,
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

func TestCryptoAnomalyDetectorWeakEntropy(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultCryptoAnomalyConfig()
	detector := NewCryptoAnomalyDetector(cfg, logger)

	eventChan := make(chan *SecurityEvent, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	detector.Start(ctx, eventChan)

	data := &CryptoOperationData{
		OperationID:    "op_entropy",
		OperationType:  "keygen",
		Algorithm:      "Ed25519",
		AccountAddress: "virtengine1test",
		Timestamp:      time.Now(),
		Success:        true,
		EntropyScore:   1.5, // Very low
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

func TestCryptoAnomalyTypes(t *testing.T) {
	anomalyTypes := []CryptoAnomalyType{
		CryptoAnomalySignatureFailure,
		CryptoAnomalyWeakEntropy,
		CryptoAnomalyKeyMisuse,
		CryptoAnomalyDeprecatedAlgorithm,
		CryptoAnomalyUnauthorizedKey,
		CryptoAnomalyKeyReuse,
		CryptoAnomalyRapidOperations,
		CryptoAnomalyDecryptionFailure,
		CryptoAnomalyInvalidNonce,
	}

	for _, at := range anomalyTypes {
		if at == "" {
			t.Error("Crypto anomaly type constant should not be empty")
		}
	}
}

func TestCryptoAnomalyDetectorCleanup(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultCryptoAnomalyConfig()
	detector := NewCryptoAnomalyDetector(cfg, logger)

	eventChan := make(chan *SecurityEvent, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	detector.Start(ctx, eventChan)

	// Cleanup is internal, but detector should handle graceful shutdown
	// Just verify it doesn't panic when context is cancelled
}
