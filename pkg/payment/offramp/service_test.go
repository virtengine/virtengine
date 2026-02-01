// Package offramp provides fiat off-ramp integration for token-to-fiat payouts.
//
// VE-5E: Fiat off-ramp integration tests
package offramp

import (
	"context"
	"testing"
)

// ============================================================================
// Monitoring Tests (these work with types defined in monitoring.go)
// ============================================================================

func TestMetricsCollection(t *testing.T) {
	metrics := NewMetricsCollector()
	
	// Record some payouts
	metrics.RecordPayout(ProviderPayPal, "USD", 10000, 100)
	metrics.RecordPayout(ProviderACH, "USD", 20000, 50)
	metrics.RecordPayoutResult(PayoutStatusSucceeded, 500)
	metrics.RecordPayoutResult(PayoutStatusFailed, 1000)
	
	// Record compliance events
	metrics.RecordKYCRejection()
	metrics.RecordAMLFlagged()
	
	// Get snapshot
	snapshot := metrics.GetSnapshot()
	
	if snapshot.PayoutsTotal != 2 {
		t.Errorf("expected 2 total payouts, got %d", snapshot.PayoutsTotal)
	}
	
	if snapshot.PayoutsSucceeded != 1 {
		t.Errorf("expected 1 succeeded payout, got %d", snapshot.PayoutsSucceeded)
	}
	
	if snapshot.PayoutsFailed != 1 {
		t.Errorf("expected 1 failed payout, got %d", snapshot.PayoutsFailed)
	}
	
	if snapshot.KYCRejections != 1 {
		t.Errorf("expected 1 KYC rejection, got %d", snapshot.KYCRejections)
	}
	
	if snapshot.TotalAmountPaidOut != 30000 {
		t.Errorf("expected total amount 30000, got %d", snapshot.TotalAmountPaidOut)
	}
}

func TestAlertManager(t *testing.T) {
	cfg := DefaultAlertConfig()
	alertMgr := NewAlertManager(cfg)
	
	ctx := context.Background()
	
	// Create some alerts
	alertMgr.CreateAlert(ctx, AlertLevelInfo, "test", "Test Alert", "This is a test")
	alertMgr.CreateAlert(ctx, AlertLevelWarning, "test", "Warning Alert", "This is a warning")
	alertMgr.CreateAlert(ctx, AlertLevelError, "test", "Error Alert", "This is an error")
	
	// Get alerts
	alerts := alertMgr.GetAlerts(10)
	if len(alerts) != 3 {
		t.Errorf("expected 3 alerts, got %d", len(alerts))
	}
	
	// Check unacknowledged
	unack := alertMgr.GetUnacknowledgedAlerts()
	if len(unack) != 3 {
		t.Errorf("expected 3 unacknowledged alerts, got %d", len(unack))
	}
	
	// Acknowledge one
	err := alertMgr.AcknowledgeAlert(alerts[0].ID, "admin")
	if err != nil {
		t.Errorf("failed to acknowledge alert: %v", err)
	}
	
	unack = alertMgr.GetUnacknowledgedAlerts()
	if len(unack) != 2 {
		t.Errorf("expected 2 unacknowledged alerts, got %d", len(unack))
	}
}

func TestAlertLevels(t *testing.T) {
	// Verify alert level constants
	levels := []AlertLevel{AlertLevelInfo, AlertLevelWarning, AlertLevelError, AlertLevelCritical}
	expectedStrings := []string{"info", "warning", "error", "critical"}
	
	for i, level := range levels {
		if string(level) != expectedStrings[i] {
			t.Errorf("expected level %s, got %s", expectedStrings[i], level)
		}
	}
}

func TestMetricsRecordWebhook(t *testing.T) {
	metrics := NewMetricsCollector()
	
	// Record successful and failed webhooks
	metrics.RecordWebhook(true)
	metrics.RecordWebhook(true)
	metrics.RecordWebhook(false)
	
	snapshot := metrics.GetSnapshot()
	
	if snapshot.WebhooksReceived != 3 {
		t.Errorf("expected 3 webhooks received, got %d", snapshot.WebhooksReceived)
	}
	
	if snapshot.WebhooksProcessed != 2 {
		t.Errorf("expected 2 webhooks processed, got %d", snapshot.WebhooksProcessed)
	}
	
	if snapshot.WebhooksFailed != 1 {
		t.Errorf("expected 1 webhook failed, got %d", snapshot.WebhooksFailed)
	}
}

func TestMetricsRecordProviderError(t *testing.T) {
	metrics := NewMetricsCollector()
	
	// Record some provider errors
	metrics.RecordProviderError(ProviderPayPal, ErrProviderUnavailable)
	metrics.RecordProviderError(ProviderPayPal, ErrProviderUnavailable)
	metrics.RecordProviderError(ProviderACH, ErrProviderUnavailable)
	
	snapshot := metrics.GetSnapshot()
	
	if snapshot.ProviderErrors[ProviderPayPal] != 2 {
		t.Errorf("expected 2 PayPal errors, got %d", snapshot.ProviderErrors[ProviderPayPal])
	}
	
	if snapshot.ProviderErrors[ProviderACH] != 1 {
		t.Errorf("expected 1 ACH error, got %d", snapshot.ProviderErrors[ProviderACH])
	}
	
	if snapshot.LastError != ErrProviderUnavailable.Error() {
		t.Errorf("expected last error %s, got %s", ErrProviderUnavailable.Error(), snapshot.LastError)
	}
}

// ============================================================================
// Config Tests
// ============================================================================

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	if !cfg.Enabled {
		t.Error("expected default config to be enabled")
	}
	
	if cfg.QuoteValiditySeconds <= 0 {
		t.Error("expected positive quote validity")
	}
	
	if cfg.LimitsConfig.DefaultDailyLimit <= 0 {
		t.Error("expected positive daily limit")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "disabled config is valid",
			cfg: Config{
				Enabled: false,
			},
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ============================================================================
// Type Tests
// ============================================================================

func TestProviderTypeValidity(t *testing.T) {
	tests := []struct {
		provider ProviderType
		valid    bool
	}{
		{ProviderPayPal, true},
		{ProviderACH, true},
		{ProviderWire, true},
		{"invalid", false},
		{"", false},
	}
	
	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			if tt.provider.IsValid() != tt.valid {
				t.Errorf("expected IsValid()=%v for %s", tt.valid, tt.provider)
			}
		})
	}
}

func TestPayoutStatusTerminal(t *testing.T) {
	terminalStatuses := []PayoutStatus{
		PayoutStatusSucceeded,
		PayoutStatusFailed,
		PayoutStatusCanceled,
		PayoutStatusReversed,
	}
	
	nonTerminalStatuses := []PayoutStatus{
		PayoutStatusPending,
		PayoutStatusProcessing,
		PayoutStatusApproved,
		PayoutStatusOnHold,
	}
	
	for _, status := range terminalStatuses {
		if !status.IsTerminal() {
			t.Errorf("expected %s to be terminal", status)
		}
	}
	
	for _, status := range nonTerminalStatuses {
		if status.IsTerminal() {
			t.Errorf("expected %s to be non-terminal", status)
		}
	}
}

func TestKYCStatusValues(t *testing.T) {
	// Verify status constants are set correctly
	if KYCStatusVerified != "verified" {
		t.Error("unexpected KYCStatusVerified value")
	}
	if KYCStatusPending != "pending" {
		t.Error("unexpected KYCStatusPending value")
	}
	if KYCStatusFailed != "failed" {
		t.Error("unexpected KYCStatusFailed value")
	}
}

func TestAMLStatusValues(t *testing.T) {
	// Verify status constants are set correctly
	if AMLStatusCleared != "cleared" {
		t.Error("unexpected AMLStatusCleared value")
	}
	if AMLStatusFlagged != "flagged" {
		t.Error("unexpected AMLStatusFlagged value")
	}
	if AMLStatusRejected != "rejected" {
		t.Error("unexpected AMLStatusRejected value")
	}
}

// ============================================================================
// Store Tests
// ============================================================================

func TestInMemoryPayoutStore(t *testing.T) {
	store := NewInMemoryPayoutStore()
	ctx := context.Background()
	
	// Create a test payout
	payout := &PayoutIntent{
		ID:             "payout_123",
		Provider:       ProviderPayPal,
		Status:         PayoutStatusPending,
		AccountAddress: "cosmos1abc",
	}
	
	// Save
	err := store.Save(ctx, payout)
	if err != nil {
		t.Fatalf("failed to save payout: %v", err)
	}
	
	// Get by ID
	retrieved, err := store.GetByID(ctx, "payout_123")
	if err != nil {
		t.Fatalf("failed to get payout: %v", err)
	}
	
	if retrieved.ID != payout.ID {
		t.Errorf("expected ID %s, got %s", payout.ID, retrieved.ID)
	}
	
	// List by account
	payouts, err := store.ListByAccount(ctx, "cosmos1abc", 10)
	if err != nil {
		t.Fatalf("failed to list payouts: %v", err)
	}
	
	if len(payouts) != 1 {
		t.Errorf("expected 1 payout, got %d", len(payouts))
	}
}

func TestInMemoryReconciliationStore(t *testing.T) {
	store := NewInMemoryReconciliationStore()
	ctx := context.Background()
	
	// Create a test record
	record := &ReconciliationRecord{
		ID:            "rec_123",
		PayoutID:      "payout_123",
		Status:        ReconciliationMatched,
		OnChainAmount: 10000,
	}
	
	// Save
	err := store.Save(ctx, record)
	if err != nil {
		t.Fatalf("failed to save record: %v", err)
	}
	
	// Get by payout ID
	retrieved, err := store.GetByPayoutID(ctx, "payout_123")
	if err != nil {
		t.Fatalf("failed to get record: %v", err)
	}
	
	if retrieved.PayoutID != record.PayoutID {
		t.Errorf("expected payout ID %s, got %s", record.PayoutID, retrieved.PayoutID)
	}
}

func TestQuoteStore(t *testing.T) {
	store := NewQuoteStore()
	ctx := context.Background()
	
	// Create a test quote
	quote := &PayoutQuote{
		QuoteID:  "quote_123",
		Provider: ProviderPayPal,
	}
	
	// Save
	err := store.Save(ctx, quote)
	if err != nil {
		t.Fatalf("failed to save quote: %v", err)
	}
	
	// Get
	retrieved, err := store.Get(ctx, "quote_123")
	if err != nil {
		t.Fatalf("failed to get quote: %v", err)
	}
	
	if retrieved.QuoteID != quote.QuoteID {
		t.Errorf("expected quote ID %s, got %s", quote.QuoteID, retrieved.QuoteID)
	}
	
	// Delete
	err = store.Delete(ctx, "quote_123")
	if err != nil {
		t.Fatalf("failed to delete quote: %v", err)
	}
	
	// Verify deleted
	_, err = store.Get(ctx, "quote_123")
	if err == nil {
		t.Error("expected error getting deleted quote")
	}
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkMetricsRecording(b *testing.B) {
	metrics := NewMetricsCollector()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.RecordPayout(ProviderPayPal, "USD", 10000, 100)
		metrics.RecordPayoutResult(PayoutStatusSucceeded, 500)
	}
}

func BenchmarkAlertCreation(b *testing.B) {
	cfg := DefaultAlertConfig()
	alertMgr := NewAlertManager(cfg)
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		alertMgr.CreateAlert(ctx, AlertLevelInfo, "benchmark", "Test", "Test message")
	}
}

