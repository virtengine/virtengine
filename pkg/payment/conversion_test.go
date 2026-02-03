// Package payment provides payment gateway integration for Visa/Mastercard.
//
// PAY-002: Tests for conversion execution and treasury transfer
package payment

import (
	"context"
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
)

// ============================================================================
// Conversion Executor Tests
// ============================================================================

func TestConversionExecutor_ExecuteConversion_Success(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryLedgerStore()
	treasury := NewMockTreasuryTransfer("virtengine1treasury", map[string]int64{"uve": 1000000000})
	config := DefaultConversionExecutorConfig()
	config.RequirePaymentVerification = false // Disable for this test

	executor := NewConversionExecutor(config, store, treasury, nil)

	quote := CreateTestQuote(10000, 10000000, "virtengine1abc123def456ghi789")
	req := ConversionExecutionRequest{
		Quote:           quote,
		PaymentIntentID: "pi_test123",
		IdempotencyKey:  "test-key-001",
	}

	result, err := executor.ExecuteConversion(ctx, req)
	if err != nil {
		t.Fatalf("ExecuteConversion failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got failure: %v", result.ErrorCode)
	}

	if result.TxHash == "" {
		t.Error("Expected TxHash to be set")
	}

	if result.LedgerEntry == nil {
		t.Fatal("Expected LedgerEntry to be set")
	}

	if result.LedgerEntry.Status != ConversionStatusCompleted {
		t.Errorf("Expected status Completed, got %s", result.LedgerEntry.Status)
	}
}

func TestConversionExecutor_Idempotency(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryLedgerStore()
	treasury := NewMockTreasuryTransfer("virtengine1treasury", map[string]int64{"uve": 1000000000})
	config := DefaultConversionExecutorConfig()
	config.RequirePaymentVerification = false

	executor := NewConversionExecutor(config, store, treasury, nil)

	quote := CreateTestQuote(10000, 10000000, "virtengine1abc123def456ghi789")
	req := ConversionExecutionRequest{
		Quote:           quote,
		PaymentIntentID: "pi_test123",
		IdempotencyKey:  "idempotent-key-001",
	}

	// First execution
	result1, err := executor.ExecuteConversion(ctx, req)
	if err != nil {
		t.Fatalf("First execution failed: %v", err)
	}
	if !result1.Success {
		t.Fatalf("First execution should succeed")
	}

	// Second execution with same idempotency key
	result2, err := executor.ExecuteConversion(ctx, req)
	if err != nil {
		t.Fatalf("Second execution failed: %v", err)
	}
	if !result2.Success {
		t.Fatalf("Second execution should succeed (idempotent)")
	}
	if !result2.AlreadyCompleted {
		t.Error("Second execution should be marked as AlreadyCompleted")
	}
	if result1.TxHash != result2.TxHash {
		t.Error("TxHash should be the same for idempotent requests")
	}
	if result1.LedgerEntry.ID != result2.LedgerEntry.ID {
		t.Error("Ledger entry ID should be the same for idempotent requests")
	}
}

func TestConversionExecutor_InsufficientTreasury(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryLedgerStore()
	treasury := NewMockTreasuryTransfer("virtengine1treasury", map[string]int64{"uve": 100}) // Low balance
	config := DefaultConversionExecutorConfig()
	config.RequirePaymentVerification = false

	executor := NewConversionExecutor(config, store, treasury, nil)

	quote := CreateTestQuote(10000, 10000000, "virtengine1abc123def456ghi789")
	req := ConversionExecutionRequest{
		Quote:           quote,
		PaymentIntentID: "pi_test123",
		IdempotencyKey:  "insufficient-key-001",
	}

	result, _ := executor.ExecuteConversion(ctx, req)

	if result.Success {
		t.Fatal("Expected failure due to insufficient treasury")
	}

	if result.ErrorCode != ConversionErrorInsufficientTreasury {
		t.Errorf("Expected error code InsufficientTreasury, got %s", result.ErrorCode)
	}

	// Should be marked for retry since insufficient treasury is retryable
	entry, _ := store.GetByID(ctx, result.LedgerEntry.ID)
	if entry.RetryCount == 0 {
		t.Error("Expected retry count to be incremented")
	}
}

func TestConversionExecutor_ExpiredQuote(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryLedgerStore()
	treasury := NewMockTreasuryTransfer("virtengine1treasury", map[string]int64{"uve": 1000000000})
	config := DefaultConversionExecutorConfig()
	config.RequirePaymentVerification = false

	executor := NewConversionExecutor(config, store, treasury, nil)

	quote := CreateTestQuote(10000, 10000000, "virtengine1abc123def456ghi789")
	quote.ExpiresAt = time.Now().Add(-1 * time.Hour) // Already expired

	req := ConversionExecutionRequest{
		Quote:           quote,
		PaymentIntentID: "pi_test123",
		IdempotencyKey:  "expired-key-001",
	}

	result, err := executor.ExecuteConversion(ctx, req)

	if result.Success {
		t.Fatal("Expected failure due to expired quote")
	}

	if err == nil {
		t.Error("Expected error to be returned")
	}
}

func TestConversionExecutor_InvalidAddress(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryLedgerStore()
	treasury := NewMockTreasuryTransfer("virtengine1treasury", map[string]int64{"uve": 1000000000})
	config := DefaultConversionExecutorConfig()
	config.RequirePaymentVerification = false

	executor := NewConversionExecutor(config, store, treasury, nil)

	quote := CreateTestQuote(10000, 10000000, "cosmos1invalid") // Wrong prefix

	req := ConversionExecutionRequest{
		Quote:           quote,
		PaymentIntentID: "pi_test123",
		IdempotencyKey:  "invalid-addr-001",
	}

	result, err := executor.ExecuteConversion(ctx, req)

	if result.Success {
		t.Fatal("Expected failure due to invalid address")
	}

	if err == nil {
		t.Error("Expected error to be returned")
	}
}

func TestConversionExecutor_TransferFailure(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryLedgerStore()
	treasury := NewMockTreasuryTransfer("virtengine1treasury", map[string]int64{"uve": 1000000000})
	treasury.SimulateFailure = true
	config := DefaultConversionExecutorConfig()
	config.RequirePaymentVerification = false

	executor := NewConversionExecutor(config, store, treasury, nil)

	quote := CreateTestQuote(10000, 10000000, "virtengine1abc123def456ghi789")
	req := ConversionExecutionRequest{
		Quote:           quote,
		PaymentIntentID: "pi_test123",
		IdempotencyKey:  "transfer-fail-001",
	}

	result, _ := executor.ExecuteConversion(ctx, req)

	if result.Success {
		t.Fatal("Expected failure due to transfer failure")
	}

	if result.ErrorCode != ConversionErrorTransferFailed {
		t.Errorf("Expected error code TransferFailed, got %s", result.ErrorCode)
	}

	// Transfer failures are retryable
	entry, _ := store.GetByID(ctx, result.LedgerEntry.ID)
	if entry.RetryCount == 0 {
		t.Error("Expected retry count to be incremented")
	}
}

func TestConversionExecutor_RetryFailedConversion(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryLedgerStore()
	treasury := NewMockTreasuryTransfer("virtengine1treasury", map[string]int64{"uve": 1000000000})
	treasury.SimulateFailure = true
	config := DefaultConversionExecutorConfig()
	config.RequirePaymentVerification = false

	executor := NewConversionExecutor(config, store, treasury, nil)

	quote := CreateTestQuote(10000, 10000000, "virtengine1abc123def456ghi789")
	req := ConversionExecutionRequest{
		Quote:           quote,
		PaymentIntentID: "pi_test123",
		IdempotencyKey:  "retry-test-001",
	}

	// First attempt fails
	result1, _ := executor.ExecuteConversion(ctx, req)
	if result1.Success {
		t.Fatal("Expected first attempt to fail")
	}

	// Fix the treasury
	treasury.SimulateFailure = false

	// Retry
	result2, err := executor.RetryFailedConversion(ctx, result1.LedgerEntry.ID)
	if err != nil {
		t.Fatalf("Retry failed unexpectedly: %v", err)
	}

	if !result2.Success {
		t.Errorf("Expected retry to succeed, got error: %v", result2.ErrorCode)
	}
}

func TestConversionExecutor_ReconcileConversion(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryLedgerStore()
	treasury := NewMockTreasuryTransfer("virtengine1treasury", map[string]int64{"uve": 1000000000})
	config := DefaultConversionExecutorConfig()

	executor := NewConversionExecutor(config, store, treasury, nil)

	// Manually create an entry in reconciling state
	entry := NewConversionLedgerEntry(
		"conv_test_reconcile",
		"reconcile-key-001",
		CreateTestQuote(10000, 10000000, "virtengine1abc123def456ghi789"),
		"pi_test123",
		3,
	)
	entry.MarkReconciling(ConversionErrorInternal, "test reconciliation")
	_ = store.Save(ctx, entry)

	// Reconcile
	err := executor.ReconcileConversion(ctx, entry.ID, "0x123abc", 12345)
	if err != nil {
		t.Fatalf("Reconciliation failed: %v", err)
	}

	// Verify entry is now completed
	updated, _ := executor.GetLedgerEntry(ctx, entry.ID)
	if updated.Status != ConversionStatusCompleted {
		t.Errorf("Expected status Completed, got %s", updated.Status)
	}
	if updated.TxHash != "0x123abc" {
		t.Error("TxHash not set correctly")
	}
	if updated.BlockHeight != 12345 {
		t.Error("BlockHeight not set correctly")
	}
}

func TestConversionExecutor_RefundConversion(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryLedgerStore()
	treasury := NewMockTreasuryTransfer("virtengine1treasury", map[string]int64{"uve": 1000000000})
	treasury.SimulateFailure = true
	config := DefaultConversionExecutorConfig()
	config.RequirePaymentVerification = false
	config.MaxRetries = 0 // No retries for this test

	executor := NewConversionExecutor(config, store, treasury, nil)

	quote := CreateTestQuote(10000, 10000000, "virtengine1abc123def456ghi789")
	req := ConversionExecutionRequest{
		Quote:           quote,
		PaymentIntentID: "pi_test123",
		IdempotencyKey:  "refund-test-001",
	}

	// Execute and fail
	result, _ := executor.ExecuteConversion(ctx, req)
	if result.Success {
		t.Fatal("Expected execution to fail")
	}

	// Refund
	err := executor.RefundConversion(ctx, result.LedgerEntry.ID, "customer requested refund")
	if err != nil {
		t.Fatalf("Refund failed: %v", err)
	}

	// Verify entry is refunded
	updated, _ := executor.GetLedgerEntry(ctx, result.LedgerEntry.ID)
	if updated.Status != ConversionStatusRefunded {
		t.Errorf("Expected status Refunded, got %s", updated.Status)
	}
	if updated.Metadata["refund_reason"] != "customer requested refund" {
		t.Error("Refund reason not set correctly")
	}
}

func TestConversionExecutor_ListPendingConversions(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryLedgerStore()
	treasury := NewMockTreasuryTransfer("virtengine1treasury", map[string]int64{"uve": 1000000000})
	config := DefaultConversionExecutorConfig()
	config.RequirePaymentVerification = false
	config.MaxRetries = 5     // High retry count so entries stay pending after first failure
	config.BaseRetryDelay = 0 // No delay for test

	// Simulate failure initially
	treasury.SimulateFailure = true

	executor := NewConversionExecutor(config, store, treasury, nil)

	// Create multiple failed conversions that should be pending (scheduled for retry)
	for i := 0; i < 3; i++ {
		quote := CreateTestQuote(10000, 10000000, "virtengine1abc123def456ghi789")
		req := ConversionExecutionRequest{
			Quote:           quote,
			PaymentIntentID: fmt.Sprintf("pi_test%d", i),
			IdempotencyKey:  fmt.Sprintf("pending-test-%d", i),
		}
		_, _ = executor.ExecuteConversion(ctx, req)
	}

	// List pending - entries should be pending with retry scheduled
	pending, err := executor.ListPendingConversions(ctx)
	if err != nil {
		t.Fatalf("ListPendingConversions failed: %v", err)
	}

	if len(pending) != 3 {
		t.Errorf("Expected 3 pending conversions, got %d", len(pending))
	}

	// Verify they are all in pending status
	for _, entry := range pending {
		if entry.Status != ConversionStatusPending {
			t.Errorf("Expected status Pending, got %s", entry.Status)
		}
		if entry.RetryCount == 0 {
			t.Error("Expected retry count to be > 0")
		}
	}
}

// ============================================================================
// Ledger Entry Tests
// ============================================================================

func TestConversionLedgerEntry_StatusTransitions(t *testing.T) {
	entry := NewConversionLedgerEntry(
		"test-id",
		"test-key",
		CreateTestQuote(10000, 10000000, "virtengine1abc123def456ghi789"),
		"pi_test",
		3,
	)

	// Initial state
	if entry.Status != ConversionStatusPending {
		t.Errorf("Expected initial status Pending, got %s", entry.Status)
	}

	// Pending -> Executing
	if err := entry.MarkExecuting(); err != nil {
		t.Errorf("MarkExecuting failed: %v", err)
	}
	if entry.Status != ConversionStatusExecuting {
		t.Errorf("Expected status Executing, got %s", entry.Status)
	}

	// Executing -> Completed
	entry.MarkCompleted("0xabc", 1234, "virtengine1treasury")
	if entry.Status != ConversionStatusCompleted {
		t.Errorf("Expected status Completed, got %s", entry.Status)
	}
	if entry.TxHash != "0xabc" {
		t.Error("TxHash not set")
	}
	if entry.BlockHeight != 1234 {
		t.Error("BlockHeight not set")
	}
	if entry.CompletedAt == nil {
		t.Error("CompletedAt should be set")
	}
}

func TestConversionLedgerEntry_RetryWithBackoff(t *testing.T) {
	entry := NewConversionLedgerEntry(
		"test-id",
		"test-key",
		CreateTestQuote(10000, 10000000, "virtengine1abc123def456ghi789"),
		"pi_test",
		3,
	)

	baseDelay := 1 * time.Second

	// First retry
	retried := entry.MarkForRetry(ConversionErrorTransferFailed, "transfer failed", baseDelay)
	if !retried {
		t.Error("First retry should succeed")
	}
	if entry.RetryCount != 1 {
		t.Errorf("Expected retry count 1, got %d", entry.RetryCount)
	}
	if entry.NextRetryAt == nil {
		t.Error("NextRetryAt should be set")
	}

	// Second retry
	retried = entry.MarkForRetry(ConversionErrorTransferFailed, "transfer failed", baseDelay)
	if !retried {
		t.Error("Second retry should succeed")
	}
	if entry.RetryCount != 2 {
		t.Errorf("Expected retry count 2, got %d", entry.RetryCount)
	}

	// Third retry
	retried = entry.MarkForRetry(ConversionErrorTransferFailed, "transfer failed", baseDelay)
	if !retried {
		t.Error("Third retry should succeed")
	}
	if entry.RetryCount != 3 {
		t.Errorf("Expected retry count 3, got %d", entry.RetryCount)
	}

	// Fourth retry should fail (max retries = 3)
	retried = entry.MarkForRetry(ConversionErrorTransferFailed, "transfer failed", baseDelay)
	if retried {
		t.Error("Fourth retry should not succeed (max retries)")
	}
	if entry.Status != ConversionStatusFailed {
		t.Errorf("Expected status Failed, got %s", entry.Status)
	}
}

func TestConversionLedgerEntry_CanRetry(t *testing.T) {
	tests := []struct {
		name       string
		status     ConversionStatus
		errorCode  ConversionErrorCode
		retryCount int
		maxRetries int
		expected   bool
	}{
		{
			name:       "Pending with retryable error",
			status:     ConversionStatusPending,
			errorCode:  ConversionErrorTransferFailed,
			retryCount: 0,
			maxRetries: 3,
			expected:   true,
		},
		{
			name:       "Completed cannot retry",
			status:     ConversionStatusCompleted,
			errorCode:  ConversionErrorNone,
			retryCount: 0,
			maxRetries: 3,
			expected:   false,
		},
		{
			name:       "Max retries exceeded",
			status:     ConversionStatusPending,
			errorCode:  ConversionErrorTransferFailed,
			retryCount: 3,
			maxRetries: 3,
			expected:   false,
		},
		{
			name:       "Non-retryable error",
			status:     ConversionStatusPending,
			errorCode:  ConversionErrorQuoteExpired,
			retryCount: 0,
			maxRetries: 3,
			expected:   false,
		},
		{
			name:       "Failed with retryable error",
			status:     ConversionStatusFailed,
			errorCode:  ConversionErrorInsufficientTreasury,
			retryCount: 1,
			maxRetries: 3,
			expected:   true,
		},
		{
			name:       "Reconciling can retry",
			status:     ConversionStatusReconciling,
			errorCode:  ConversionErrorInternal,
			retryCount: 0,
			maxRetries: 3,
			expected:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entry := &ConversionLedgerEntry{
				Status:     tc.status,
				ErrorCode:  tc.errorCode,
				RetryCount: tc.retryCount,
				MaxRetries: tc.maxRetries,
			}
			result := entry.CanRetry()
			if result != tc.expected {
				t.Errorf("CanRetry() = %v, expected %v", result, tc.expected)
			}
		})
	}
}

// ============================================================================
// In-Memory Store Tests
// ============================================================================

func TestInMemoryLedgerStore_BasicOperations(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryLedgerStore()

	entry := NewConversionLedgerEntry(
		"test-id-001",
		"test-key-001",
		CreateTestQuote(10000, 10000000, "virtengine1abc123def456ghi789"),
		"pi_test",
		3,
	)

	// Save
	if err := store.Save(ctx, entry); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Get by ID
	retrieved, err := store.GetByID(ctx, entry.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.ID != entry.ID {
		t.Error("Retrieved entry ID mismatch")
	}

	// Get by idempotency key
	retrieved2, err := store.GetByIdempotencyKey(ctx, entry.IdempotencyKey)
	if err != nil {
		t.Fatalf("GetByIdempotencyKey failed: %v", err)
	}
	if retrieved2.ID != entry.ID {
		t.Error("Retrieved entry ID mismatch")
	}

	// Not found
	_, err = store.GetByID(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent entry")
	}
}

func TestInMemoryLedgerStore_ListByStatus(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryLedgerStore()

	// Create entries with different statuses
	pending := NewConversionLedgerEntry("id1", "key1", CreateTestQuote(10000, 10000000, "virtengine1abc"), "pi1", 3)
	pending.Status = ConversionStatusPending

	completed := NewConversionLedgerEntry("id2", "key2", CreateTestQuote(10000, 10000000, "virtengine1def"), "pi2", 3)
	completed.Status = ConversionStatusCompleted

	failed := NewConversionLedgerEntry("id3", "key3", CreateTestQuote(10000, 10000000, "virtengine1ghi"), "pi3", 3)
	failed.Status = ConversionStatusFailed

	_ = store.Save(ctx, pending)
	_ = store.Save(ctx, completed)
	_ = store.Save(ctx, failed)

	// List pending
	pendingList, _ := store.ListByStatus(ctx, ConversionStatusPending)
	if len(pendingList) != 1 {
		t.Errorf("Expected 1 pending entry, got %d", len(pendingList))
	}

	// List completed
	completedList, _ := store.ListByStatus(ctx, ConversionStatusCompleted)
	if len(completedList) != 1 {
		t.Errorf("Expected 1 completed entry, got %d", len(completedList))
	}
}

// ============================================================================
// Mock Treasury Tests
// ============================================================================

func TestMockTreasuryTransfer(t *testing.T) {
	ctx := context.Background()
	treasury := NewMockTreasuryTransfer("virtengine1treasury", map[string]int64{"uve": 1000000})

	// Check balance
	balance, err := treasury.GetTreasuryBalance(ctx, "uve")
	if err != nil {
		t.Fatalf("GetTreasuryBalance failed: %v", err)
	}
	if balance.Available != 1000000 {
		t.Errorf("Expected balance 1000000, got %d", balance.Available)
	}

	// Send transfer
	req := TreasuryTransferRequest{
		DestinationAddress: "virtengine1abc123def456ghi789",
		Amount:             100000,
		Denom:              "uve",
		IdempotencyKey:     "test-transfer",
	}
	result, err := treasury.SendFromTreasury(ctx, req)
	if err != nil {
		t.Fatalf("SendFromTreasury failed: %v", err)
	}
	if !result.Success {
		t.Error("Expected transfer to succeed")
	}
	if result.TxHash == "" {
		t.Error("Expected TxHash to be set")
	}

	// Check balance after transfer
	balance, _ = treasury.GetTreasuryBalance(ctx, "uve")
	if balance.Available != 900000 {
		t.Errorf("Expected balance 900000 after transfer, got %d", balance.Available)
	}

	// Validate address
	if err := treasury.ValidateAddress(ctx, "virtengine1validaddress123456789"); err != nil {
		t.Errorf("Valid address rejected: %v", err)
	}
	if err := treasury.ValidateAddress(ctx, "cosmos1invalid"); err == nil {
		t.Error("Invalid address should be rejected")
	}
}

// ============================================================================
// Conversion Types Tests
// ============================================================================

func TestConversionStatus_IsFinal(t *testing.T) {
	tests := []struct {
		status   ConversionStatus
		expected bool
	}{
		{ConversionStatusPending, false},
		{ConversionStatusExecuting, false},
		{ConversionStatusCompleted, true},
		{ConversionStatusFailed, true},
		{ConversionStatusRefunded, true},
		{ConversionStatusReconciling, false},
	}

	for _, tc := range tests {
		t.Run(string(tc.status), func(t *testing.T) {
			if tc.status.IsFinal() != tc.expected {
				t.Errorf("IsFinal() = %v, expected %v", tc.status.IsFinal(), tc.expected)
			}
		})
	}
}

func TestConversionErrorCode_IsRetryable(t *testing.T) {
	tests := []struct {
		code     ConversionErrorCode
		expected bool
	}{
		{ConversionErrorNone, false},
		{ConversionErrorQuoteExpired, false},
		{ConversionErrorPaymentNotSucceeded, false},
		{ConversionErrorInsufficientTreasury, true},
		{ConversionErrorTransferFailed, true},
		{ConversionErrorInvalidAddress, false},
		{ConversionErrorDuplicate, false},
		{ConversionErrorInternal, true},
	}

	for _, tc := range tests {
		t.Run(string(tc.code), func(t *testing.T) {
			if tc.code.IsRetryable() != tc.expected {
				t.Errorf("IsRetryable() = %v, expected %v", tc.code.IsRetryable(), tc.expected)
			}
		})
	}
}

func TestNewConversionLedgerEntry(t *testing.T) {
	quote := ConversionQuote{
		ID: "quote-123",
		FiatAmount: Amount{
			Value:    10000,
			Currency: CurrencyUSD,
		},
		CryptoAmount:       sdkmath.NewInt(10000000),
		CryptoDenom:        "uve",
		DestinationAddress: "virtengine1abc123def456ghi789",
		Rate: ConversionRate{
			FromCurrency: CurrencyUSD,
			ToCrypto:     "uve",
			Rate:         sdkmath.LegacyNewDec(1000),
		},
		Fee: Amount{
			Value:    100,
			Currency: CurrencyUSD,
		},
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	entry := NewConversionLedgerEntry("entry-001", "key-001", quote, "pi_123", 5)

	if entry.ID != "entry-001" {
		t.Error("ID not set correctly")
	}
	if entry.IdempotencyKey != "key-001" {
		t.Error("IdempotencyKey not set correctly")
	}
	if entry.QuoteID != "quote-123" {
		t.Error("QuoteID not set correctly")
	}
	if entry.PaymentIntentID != "pi_123" {
		t.Error("PaymentIntentID not set correctly")
	}
	if entry.MaxRetries != 5 {
		t.Error("MaxRetries not set correctly")
	}
	if entry.Status != ConversionStatusPending {
		t.Error("Initial status should be Pending")
	}
	if entry.CryptoAmount.Int64() != 10000000 {
		t.Error("CryptoAmount not set correctly")
	}
	if entry.DestinationAddress != "virtengine1abc123def456ghi789" {
		t.Error("DestinationAddress not set correctly")
	}
}
