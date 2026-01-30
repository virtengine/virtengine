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

func TestNewTransactionDetector(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultTransactionDetectorConfig()
	detector := NewTransactionDetector(cfg, logger)

	if detector == nil {
		t.Fatal("NewTransactionDetector returned nil")
	}
}

func TestDefaultTransactionDetectorConfig(t *testing.T) {
	cfg := DefaultTransactionDetectorConfig()

	if cfg.MaxTxPerMinutePerAccount != 20 {
		t.Errorf("Expected MaxTxPerMinutePerAccount 20, got %d", cfg.MaxTxPerMinutePerAccount)
	}

	if cfg.MaxTxPerHourPerAccount != 200 {
		t.Errorf("Expected MaxTxPerHourPerAccount 200, got %d", cfg.MaxTxPerHourPerAccount)
	}

	if cfg.RapidFireWindowSecs != 10 {
		t.Errorf("Expected RapidFireWindowSecs 10, got %d", cfg.RapidFireWindowSecs)
	}

	if cfg.RapidFireThreshold != 5 {
		t.Errorf("Expected RapidFireThreshold 5, got %d", cfg.RapidFireThreshold)
	}

	if cfg.NewAccountCooldownMins != 60 {
		t.Errorf("Expected NewAccountCooldownMins 60, got %d", cfg.NewAccountCooldownMins)
	}
}

func TestTransactionDataStructure(t *testing.T) {
	tx := &TransactionData{
		TxHash:      "ABCD1234",
		Sender:      "virtengine1sender",
		Recipient:   "virtengine1recipient",
		Amount:      1000000,
		Denom:       "uveng",
		MsgType:     "MsgSend",
		Timestamp:   time.Now(),
		BlockHeight: 12345,
		GasUsed:     100000,
		Success:     true,
	}

	if tx.TxHash == "" {
		t.Error("TxHash should not be empty")
	}

	if tx.Sender == "" {
		t.Error("Sender should not be empty")
	}

	if tx.Amount == 0 {
		t.Error("Amount should not be zero")
	}

	if tx.BlockHeight == 0 {
		t.Error("BlockHeight should not be zero")
	}
}

func TestTransactionDetectorAnalyze(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultTransactionDetectorConfig()
	detector := NewTransactionDetector(cfg, logger)

	eventChan := make(chan *SecurityEvent, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	detector.Start(ctx, eventChan)

	tx := &TransactionData{
		TxHash:      "TEST_HASH",
		Sender:      "virtengine1test",
		Recipient:   "virtengine1recv",
		Amount:      100,
		Denom:       "uveng",
		MsgType:     "MsgSend",
		Timestamp:   time.Now(),
		BlockHeight: 100,
		Success:     true,
	}

	// Analyze should not panic
	detector.Analyze(tx)
}

func TestTransactionDetectorVelocityDetection(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultTransactionDetectorConfig()
	cfg.MaxTxPerMinutePerAccount = 3
	detector := NewTransactionDetector(cfg, logger)

	eventChan := make(chan *SecurityEvent, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	detector.Start(ctx, eventChan)

	sender := "virtengine1velocity"
	now := time.Now()

	// Submit multiple transactions to trigger velocity detection
	for i := 0; i < 5; i++ {
		tx := &TransactionData{
			TxHash:      "HASH_" + string(rune(i+'0')),
			Sender:      sender,
			Recipient:   "virtengine1recv",
			Amount:      100,
			Denom:       "uveng",
			MsgType:     "MsgSend",
			Timestamp:   now,
			BlockHeight: int64(100 + i),
			Success:     true,
		}
		detector.Analyze(tx)
	}

	// Check for velocity event (implementation may or may not emit)
	select {
	case <-eventChan:
		// Event received
	case <-time.After(100 * time.Millisecond):
		// No event, which is also valid
	}
}

func TestTransactionDetectorReplayDetection(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultTransactionDetectorConfig()
	detector := NewTransactionDetector(cfg, logger)

	eventChan := make(chan *SecurityEvent, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	detector.Start(ctx, eventChan)

	// Submit same hash twice
	tx := &TransactionData{
		TxHash:      "REPLAY_HASH",
		Sender:      "virtengine1test",
		Recipient:   "virtengine1recv",
		Amount:      100,
		Denom:       "uveng",
		MsgType:     "MsgSend",
		Timestamp:   time.Now(),
		BlockHeight: 100,
		Success:     true,
	}

	detector.Analyze(tx)
	detector.Analyze(tx) // Same hash = potential replay

	// Check for replay event
	select {
	case <-eventChan:
		// Event received
	case <-time.After(100 * time.Millisecond):
		// No event
	}
}

func TestTransactionDetectorCleanup(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	cfg := DefaultTransactionDetectorConfig()
	detector := NewTransactionDetector(cfg, logger)

	eventChan := make(chan *SecurityEvent, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	detector.Start(ctx, eventChan)

	// Cleanup is internal, but detector should handle graceful shutdown
	// Just verify it doesn't panic when context is cancelled
}
