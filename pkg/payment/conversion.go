// Package payment provides payment gateway integration for Visa/Mastercard.
//
// PAY-002: Conversion execution and treasury transfer implementation
package payment

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"
)

// ============================================================================
// Conversion Executor Errors
// ============================================================================

var (
	// ErrConversionNotFound indicates ledger entry not found
	ErrConversionNotFound = errors.New("conversion ledger entry not found")

	// ErrConversionAlreadyCompleted indicates conversion already finished
	ErrConversionAlreadyCompleted = errors.New("conversion already completed")

	// ErrConversionNotRetryable indicates conversion cannot be retried
	ErrConversionNotRetryable = errors.New("conversion is not retryable")

	// ErrConversionInProgress indicates conversion is already executing
	ErrConversionInProgress = errors.New("conversion is already in progress")

	// ErrTreasuryTransferFailed indicates treasury transfer failed
	ErrTreasuryTransferFailed = errors.New("treasury transfer failed")

	// ErrInvalidReconciliation indicates invalid reconciliation attempt
	ErrInvalidReconciliation = errors.New("invalid reconciliation")
)

// ============================================================================
// Conversion Executor Configuration
// ============================================================================

// ConversionExecutorConfig configures the conversion executor
type ConversionExecutorConfig struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int `json:"max_retries"`

	// BaseRetryDelay is the base delay for exponential backoff
	BaseRetryDelay time.Duration `json:"base_retry_delay"`

	// TreasuryAddress is the treasury source address
	TreasuryAddress string `json:"treasury_address"`

	// DefaultDenom is the default crypto denomination
	DefaultDenom string `json:"default_denom"`

	// RequirePaymentVerification requires payment intent to be verified
	RequirePaymentVerification bool `json:"require_payment_verification"`
}

// DefaultConversionExecutorConfig returns default configuration
func DefaultConversionExecutorConfig() ConversionExecutorConfig {
	return ConversionExecutorConfig{
		MaxRetries:                 3,
		BaseRetryDelay:             5 * time.Second,
		TreasuryAddress:            "",
		DefaultDenom:               "uve",
		RequirePaymentVerification: true,
	}
}

// ============================================================================
// Conversion Executor Implementation
// ============================================================================

// conversionExecutor implements ConversionExecutor
type conversionExecutor struct {
	config   ConversionExecutorConfig
	store    ConversionLedgerStore
	treasury TreasuryTransfer
	payment  Service

	// Mutex for concurrent access control
	executionMu sync.Mutex
	// Track in-flight executions by idempotency key
	inFlight map[string]struct{}
}

// NewConversionExecutor creates a new conversion executor
func NewConversionExecutor(
	config ConversionExecutorConfig,
	store ConversionLedgerStore,
	treasury TreasuryTransfer,
	payment Service,
) ConversionExecutor {
	return &conversionExecutor{
		config:   config,
		store:    store,
		treasury: treasury,
		payment:  payment,
		inFlight: make(map[string]struct{}),
	}
}

// ExecuteConversion executes a conversion with idempotency guarantees
func (e *conversionExecutor) ExecuteConversion(ctx context.Context, req ConversionExecutionRequest) (*ConversionExecutionResult, error) {
	// Validate request
	if err := e.validateRequest(ctx, req); err != nil {
		return &ConversionExecutionResult{
			Success:   false,
			ErrorCode: ConversionErrorInvalidAddress,
			Error:     err,
		}, err
	}

	// Check for existing execution by idempotency key
	existing, err := e.store.GetByIdempotencyKey(ctx, req.IdempotencyKey)
	if err == nil && existing != nil {
		// Idempotent check: return existing result if completed
		if existing.Status == ConversionStatusCompleted {
			return &ConversionExecutionResult{
				LedgerEntry:      existing,
				Success:          true,
				TxHash:           existing.TxHash,
				AlreadyCompleted: true,
			}, nil
		}

		// If failed but not retryable, return the failure
		if existing.Status.IsFinal() {
			return &ConversionExecutionResult{
				LedgerEntry: existing,
				Success:     false,
				ErrorCode:   existing.ErrorCode,
				Error:       fmt.Errorf("%s: %s", existing.ErrorCode, existing.ErrorMessage),
			}, nil
		}

		// If in progress, return error
		if existing.Status == ConversionStatusExecuting {
			return &ConversionExecutionResult{
				LedgerEntry: existing,
				Success:     false,
				ErrorCode:   ConversionErrorDuplicate,
				Error:       ErrConversionInProgress,
			}, ErrConversionInProgress
		}

		// Otherwise, this is a retry of a pending entry
		return e.executeEntry(ctx, existing)
	}

	// Create new ledger entry
	entryID := e.generateEntryID(req)
	entry := NewConversionLedgerEntry(
		entryID,
		req.IdempotencyKey,
		req.Quote,
		req.PaymentIntentID,
		e.config.MaxRetries,
	)

	if req.Metadata != nil {
		entry.Metadata = req.Metadata
	}

	// Save initial entry
	if err := e.store.Save(ctx, entry); err != nil {
		return &ConversionExecutionResult{
			Success:   false,
			ErrorCode: ConversionErrorInternal,
			Error:     fmt.Errorf("failed to save ledger entry: %w", err),
		}, err
	}

	return e.executeEntry(ctx, entry)
}

// executeEntry performs the actual conversion execution
func (e *conversionExecutor) executeEntry(ctx context.Context, entry *ConversionLedgerEntry) (*ConversionExecutionResult, error) {
	// Acquire execution lock for this idempotency key
	if !e.acquireExecution(entry.IdempotencyKey) {
		return &ConversionExecutionResult{
			LedgerEntry: entry,
			Success:     false,
			ErrorCode:   ConversionErrorDuplicate,
			Error:       ErrConversionInProgress,
		}, ErrConversionInProgress
	}
	defer e.releaseExecution(entry.IdempotencyKey)

	// Check if entry is ready for execution
	if !entry.IsReadyForExecution() {
		if entry.Status.IsFinal() {
			return &ConversionExecutionResult{
				LedgerEntry:      entry,
				Success:          entry.Status == ConversionStatusCompleted,
				TxHash:           entry.TxHash,
				ErrorCode:        entry.ErrorCode,
				AlreadyCompleted: entry.Status == ConversionStatusCompleted,
			}, nil
		}
		return &ConversionExecutionResult{
			LedgerEntry: entry,
			Success:     false,
			ErrorCode:   ConversionErrorDuplicate,
			Error:       errors.New("entry not ready for execution"),
		}, nil
	}

	// Mark as executing
	if err := entry.MarkExecuting(); err != nil {
		return &ConversionExecutionResult{
			LedgerEntry: entry,
			Success:     false,
			ErrorCode:   ConversionErrorInternal,
			Error:       err,
		}, err
	}
	if err := e.store.Save(ctx, entry); err != nil {
		return &ConversionExecutionResult{
			Success:   false,
			ErrorCode: ConversionErrorInternal,
			Error:     fmt.Errorf("failed to update entry status: %w", err),
		}, err
	}

	// Verify payment if required
	if e.config.RequirePaymentVerification && e.payment != nil {
		intent, err := e.payment.GetPaymentIntent(ctx, entry.PaymentIntentID)
		if err != nil {
			return e.handleExecutionError(ctx, entry, ConversionErrorPaymentNotSucceeded,
				fmt.Sprintf("failed to verify payment: %v", err))
		}
		if !intent.Status.IsSuccessful() {
			return e.handleExecutionError(ctx, entry, ConversionErrorPaymentNotSucceeded,
				fmt.Sprintf("payment status is %s, not succeeded", intent.Status))
		}
	}

	// Validate treasury has sufficient funds
	balance, err := e.treasury.GetTreasuryBalance(ctx, entry.CryptoDenom)
	if err != nil {
		return e.handleExecutionError(ctx, entry, ConversionErrorInsufficientTreasury,
			fmt.Sprintf("failed to check treasury balance: %v", err))
	}
	if balance.Available < entry.CryptoAmount.Int64() {
		return e.handleExecutionError(ctx, entry, ConversionErrorInsufficientTreasury,
			fmt.Sprintf("treasury has %d available, need %d", balance.Available, entry.CryptoAmount.Int64()))
	}

	// Execute treasury transfer
	transferReq := TreasuryTransferRequest{
		DestinationAddress: entry.DestinationAddress,
		Amount:             entry.CryptoAmount.Int64(),
		Denom:              entry.CryptoDenom,
		Memo:               fmt.Sprintf("conversion:%s", entry.ID),
		IdempotencyKey:     entry.IdempotencyKey,
	}

	result, err := e.treasury.SendFromTreasury(ctx, transferReq)
	if err != nil || result == nil || !result.Success {
		errMsg := "transfer failed"
		if err != nil {
			errMsg = err.Error()
		} else if result != nil && result.Error != nil {
			errMsg = result.Error.Error()
		}
		return e.handleExecutionError(ctx, entry, ConversionErrorTransferFailed, errMsg)
	}

	// Mark as completed
	entry.MarkCompleted(result.TxHash, result.BlockHeight, result.TreasuryAddress)
	if err := e.store.Save(ctx, entry); err != nil {
		// Transfer succeeded but save failed - mark for reconciliation
		entry.MarkReconciling(ConversionErrorInternal, fmt.Sprintf("transfer succeeded but save failed: %v", err))
		_ = e.store.Save(ctx, entry)
		return &ConversionExecutionResult{
			LedgerEntry: entry,
			Success:     true,
			TxHash:      result.TxHash,
		}, nil
	}

	return &ConversionExecutionResult{
		LedgerEntry: entry,
		Success:     true,
		TxHash:      result.TxHash,
	}, nil
}

// handleExecutionError processes an execution error with retry logic
func (e *conversionExecutor) handleExecutionError(
	ctx context.Context,
	entry *ConversionLedgerEntry,
	errCode ConversionErrorCode,
	errMsg string,
) (*ConversionExecutionResult, error) {
	// If error is retryable, schedule retry
	if errCode.IsRetryable() {
		retried := entry.MarkForRetry(errCode, errMsg, e.config.BaseRetryDelay)
		if retried {
			_ = e.store.Save(ctx, entry)
			return &ConversionExecutionResult{
				LedgerEntry: entry,
				Success:     false,
				ErrorCode:   errCode,
				Error:       errors.New(errMsg),
			}, errors.New(errMsg)
		}
		// Max retries exhausted - entry already marked as failed
	} else {
		// Non-retryable error - mark as failed immediately
		entry.MarkFailed(errCode, errMsg)
	}

	_ = e.store.Save(ctx, entry)

	return &ConversionExecutionResult{
		LedgerEntry: entry,
		Success:     false,
		ErrorCode:   errCode,
		Error:       errors.New(errMsg),
	}, errors.New(errMsg)
}

// GetLedgerEntry retrieves a ledger entry by ID
func (e *conversionExecutor) GetLedgerEntry(ctx context.Context, id string) (*ConversionLedgerEntry, error) {
	entry, err := e.store.GetByID(ctx, id)
	if err != nil {
		return nil, ErrConversionNotFound
	}
	return entry, nil
}

// GetLedgerEntryByIdempotencyKey retrieves a ledger entry by idempotency key
func (e *conversionExecutor) GetLedgerEntryByIdempotencyKey(ctx context.Context, key string) (*ConversionLedgerEntry, error) {
	entry, err := e.store.GetByIdempotencyKey(ctx, key)
	if err != nil {
		return nil, ErrConversionNotFound
	}
	return entry, nil
}

// RetryFailedConversion retries a failed conversion
func (e *conversionExecutor) RetryFailedConversion(ctx context.Context, id string) (*ConversionExecutionResult, error) {
	entry, err := e.store.GetByID(ctx, id)
	if err != nil {
		return nil, ErrConversionNotFound
	}

	if entry.Status == ConversionStatusCompleted {
		return &ConversionExecutionResult{
			LedgerEntry:      entry,
			Success:          true,
			TxHash:           entry.TxHash,
			AlreadyCompleted: true,
		}, nil
	}

	if !entry.Status.IsRetryable() {
		return nil, ErrConversionNotRetryable
	}

	// Reset entry for retry
	entry.Status = ConversionStatusPending
	entry.NextRetryAt = nil
	entry.UpdatedAt = time.Now()

	if err := e.store.Save(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to reset entry for retry: %w", err)
	}

	return e.executeEntry(ctx, entry)
}

// ReconcileConversion manually reconciles a stuck conversion
func (e *conversionExecutor) ReconcileConversion(ctx context.Context, id string, txHash string, blockHeight int64) error {
	entry, err := e.store.GetByID(ctx, id)
	if err != nil {
		return ErrConversionNotFound
	}

	if entry.Status == ConversionStatusCompleted {
		return ErrConversionAlreadyCompleted
	}

	if entry.Status != ConversionStatusReconciling && entry.Status != ConversionStatusExecuting {
		return ErrInvalidReconciliation
	}

	if txHash == "" || blockHeight <= 0 {
		return fmt.Errorf("txHash and blockHeight are required for reconciliation")
	}

	entry.MarkCompleted(txHash, blockHeight, e.config.TreasuryAddress)

	if err := e.store.Save(ctx, entry); err != nil {
		return fmt.Errorf("failed to save reconciled entry: %w", err)
	}

	return nil
}

// RefundConversion refunds a failed conversion
func (e *conversionExecutor) RefundConversion(ctx context.Context, id string, reason string) error {
	entry, err := e.store.GetByID(ctx, id)
	if err != nil {
		return ErrConversionNotFound
	}

	if entry.Status == ConversionStatusCompleted {
		return errors.New("cannot refund completed conversion")
	}

	if entry.Status == ConversionStatusRefunded {
		return nil // Already refunded
	}

	// Mark as refunded
	entry.MarkRefunded()
	entry.Metadata["refund_reason"] = reason

	if err := e.store.Save(ctx, entry); err != nil {
		return fmt.Errorf("failed to save refunded entry: %w", err)
	}

	// Note: Actual payment refund would be triggered here via payment service
	// This is out of scope for this implementation

	return nil
}

// ListPendingConversions lists conversions ready for execution
func (e *conversionExecutor) ListPendingConversions(ctx context.Context) ([]*ConversionLedgerEntry, error) {
	return e.store.ListPendingReadyForExecution(ctx)
}

// ListConversionsForReconciliation lists conversions needing manual reconciliation
func (e *conversionExecutor) ListConversionsForReconciliation(ctx context.Context) ([]*ConversionLedgerEntry, error) {
	return e.store.ListByStatus(ctx, ConversionStatusReconciling)
}

// validateRequest validates a conversion execution request
func (e *conversionExecutor) validateRequest(ctx context.Context, req ConversionExecutionRequest) error {
	if req.IdempotencyKey == "" {
		return errors.New("idempotency key is required")
	}
	if req.Quote.ID == "" {
		return errors.New("quote ID is required")
	}
	if req.PaymentIntentID == "" {
		return errors.New("payment intent ID is required")
	}
	if req.Quote.IsExpired() {
		return ErrQuoteExpired
	}
	if req.Quote.DestinationAddress == "" {
		return errors.New("destination address is required")
	}

	// Validate destination address format
	if !strings.HasPrefix(req.Quote.DestinationAddress, "virtengine1") {
		return fmt.Errorf("invalid destination address format: must start with virtengine1")
	}

	// Validate with treasury
	if e.treasury != nil {
		if err := e.treasury.ValidateAddress(ctx, req.Quote.DestinationAddress); err != nil {
			return fmt.Errorf("invalid destination address: %w", err)
		}
	}

	return nil
}

// generateEntryID generates a unique entry ID
func (e *conversionExecutor) generateEntryID(req ConversionExecutionRequest) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%d", req.IdempotencyKey, req.Quote.ID, time.Now().UnixNano())))
	return fmt.Sprintf("conv_%s", hex.EncodeToString(hash[:])[:16])
}

// acquireExecution attempts to acquire an execution lock for the idempotency key
func (e *conversionExecutor) acquireExecution(key string) bool {
	e.executionMu.Lock()
	defer e.executionMu.Unlock()

	if _, exists := e.inFlight[key]; exists {
		return false
	}
	e.inFlight[key] = struct{}{}
	return true
}

// releaseExecution releases the execution lock
func (e *conversionExecutor) releaseExecution(key string) {
	e.executionMu.Lock()
	defer e.executionMu.Unlock()
	delete(e.inFlight, key)
}

// ============================================================================
// In-Memory Ledger Store (for testing/development)
// ============================================================================

// inMemoryLedgerStore is an in-memory implementation of ConversionLedgerStore
type inMemoryLedgerStore struct {
	mu            sync.RWMutex
	entries       map[string]*ConversionLedgerEntry
	byIdempotency map[string]string // idempotency key -> entry ID
}

// NewInMemoryLedgerStore creates a new in-memory ledger store
func NewInMemoryLedgerStore() ConversionLedgerStore {
	return &inMemoryLedgerStore{
		entries:       make(map[string]*ConversionLedgerEntry),
		byIdempotency: make(map[string]string),
	}
}

func (s *inMemoryLedgerStore) Save(ctx context.Context, entry *ConversionLedgerEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a copy to avoid mutation issues
	entryCopy := *entry
	s.entries[entry.ID] = &entryCopy
	s.byIdempotency[entry.IdempotencyKey] = entry.ID
	return nil
}

func (s *inMemoryLedgerStore) GetByID(ctx context.Context, id string) (*ConversionLedgerEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.entries[id]
	if !ok {
		return nil, ErrConversionNotFound
	}
	// Return a copy
	entryCopy := *entry
	return &entryCopy, nil
}

func (s *inMemoryLedgerStore) GetByIdempotencyKey(ctx context.Context, key string) (*ConversionLedgerEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.byIdempotency[key]
	if !ok {
		return nil, ErrConversionNotFound
	}
	entry := s.entries[id]
	if entry == nil {
		return nil, ErrConversionNotFound
	}
	// Return a copy
	entryCopy := *entry
	return &entryCopy, nil
}

func (s *inMemoryLedgerStore) ListByStatus(ctx context.Context, status ConversionStatus) ([]*ConversionLedgerEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*ConversionLedgerEntry
	for _, entry := range s.entries {
		if entry.Status == status {
			entryCopy := *entry
			result = append(result, &entryCopy)
		}
	}
	return result, nil
}

func (s *inMemoryLedgerStore) ListPendingReadyForExecution(ctx context.Context) ([]*ConversionLedgerEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	var result []*ConversionLedgerEntry
	for _, entry := range s.entries {
		if entry.Status == ConversionStatusPending {
			// Entry is ready if NextRetryAt is nil or the retry time has passed (or equals now)
			if entry.NextRetryAt == nil || !now.Before(*entry.NextRetryAt) {
				entryCopy := *entry
				result = append(result, &entryCopy)
			}
		}
	}
	return result, nil
}

// ============================================================================
// Mock Treasury Transfer (for testing/development)
// ============================================================================

// mockTreasuryTransfer is a mock implementation of TreasuryTransfer
type mockTreasuryTransfer struct {
	mu       sync.Mutex
	address  string
	balances map[string]int64
	txSeq    int64

	// For testing: simulate failures
	SimulateFailure   bool
	FailureError      error
	SimulateNoBalance bool
}

// NewMockTreasuryTransfer creates a new mock treasury transfer
func NewMockTreasuryTransfer(address string, initialBalances map[string]int64) *mockTreasuryTransfer {
	balances := make(map[string]int64)
	for k, v := range initialBalances {
		balances[k] = v
	}
	return &mockTreasuryTransfer{
		address:  address,
		balances: balances,
	}
}

func (m *mockTreasuryTransfer) SendFromTreasury(ctx context.Context, req TreasuryTransferRequest) (*TreasuryTransferResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.SimulateFailure {
		err := m.FailureError
		if err == nil {
			err = errors.New("simulated treasury failure")
		}
		return &TreasuryTransferResult{
			Success: false,
			Error:   err,
		}, err
	}

	balance := m.balances[req.Denom]
	if balance < req.Amount || m.SimulateNoBalance {
		return &TreasuryTransferResult{
			Success: false,
			Error:   fmt.Errorf("insufficient treasury balance: have %d, need %d", balance, req.Amount),
		}, fmt.Errorf("insufficient treasury balance")
	}

	// Deduct balance
	m.balances[req.Denom] -= req.Amount
	m.txSeq++

	return &TreasuryTransferResult{
		TxHash:          fmt.Sprintf("0x%064x", m.txSeq),
		BlockHeight:     1000 + m.txSeq,
		TreasuryAddress: m.address,
		Success:         true,
	}, nil
}

func (m *mockTreasuryTransfer) GetTreasuryBalance(ctx context.Context, denom string) (TreasuryBalance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.SimulateNoBalance {
		return TreasuryBalance{
			Denom:     denom,
			Available: 0,
			Reserved:  0,
			Total:     0,
		}, nil
	}

	balance := m.balances[denom]
	return TreasuryBalance{
		Denom:     denom,
		Available: balance,
		Reserved:  0,
		Total:     balance,
	}, nil
}

func (m *mockTreasuryTransfer) ValidateAddress(ctx context.Context, address string) error {
	if !strings.HasPrefix(address, "virtengine1") {
		return fmt.Errorf("invalid address format")
	}
	if len(address) < 20 {
		return fmt.Errorf("address too short")
	}
	return nil
}

// SetBalance sets a balance for testing
func (m *mockTreasuryTransfer) SetBalance(denom string, amount int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.balances[denom] = amount
}

// GetBalance gets a balance for testing
func (m *mockTreasuryTransfer) GetBalance(denom string) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.balances[denom]
}

// ============================================================================
// Helper Functions
// ============================================================================

// CreateTestQuote creates a test conversion quote
func CreateTestQuote(fiatValue int64, cryptoAmount int64, destAddr string) ConversionQuote {
	return ConversionQuote{
		ID: fmt.Sprintf("quote_%d", time.Now().UnixNano()),
		FiatAmount: Amount{
			Value:    fiatValue,
			Currency: CurrencyUSD,
		},
		CryptoAmount:       sdkmath.NewInt(cryptoAmount),
		CryptoDenom:        "uve",
		DestinationAddress: destAddr,
		Rate: ConversionRate{
			FromCurrency: CurrencyUSD,
			ToCrypto:     "uve",
			Rate:         sdkmath.LegacyNewDec(1),
			Timestamp:    time.Now(),
			Source:       "test",
		},
		Fee: Amount{
			Value:    fiatValue / 100, // 1% fee
			Currency: CurrencyUSD,
		},
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
}
