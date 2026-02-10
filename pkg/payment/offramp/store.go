// Package offramp provides fiat off-ramp integration for token-to-fiat payouts.
//
// VE-5E: Fiat off-ramp integration for PayPal/ACH payouts
package offramp

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// In-Memory Payout Store
// ============================================================================

// InMemoryPayoutStore implements PayoutStore with in-memory storage.
// Suitable for testing and development. Use a persistent store for production.
type InMemoryPayoutStore struct {
	mu               sync.RWMutex
	payouts          map[string]*PayoutIntent
	byIdempotencyKey map[string]string   // idempotency_key -> payout_id
	byProviderID     map[string]string   // provider_payout_id -> payout_id
	byAccount        map[string][]string // account_address -> []payout_id
}

// NewInMemoryPayoutStore creates a new in-memory payout store.
func NewInMemoryPayoutStore() *InMemoryPayoutStore {
	return &InMemoryPayoutStore{
		payouts:          make(map[string]*PayoutIntent),
		byIdempotencyKey: make(map[string]string),
		byProviderID:     make(map[string]string),
		byAccount:        make(map[string][]string),
	}
}

// Save saves or updates a payout intent.
func (s *InMemoryPayoutStore) Save(ctx context.Context, payout *PayoutIntent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for idempotency conflict
	if payout.IdempotencyKey != "" {
		if existingID, ok := s.byIdempotencyKey[payout.IdempotencyKey]; ok && existingID != payout.ID {
			return fmt.Errorf("idempotency key already used by payout %s", existingID)
		}
	}

	// Make a copy to store
	stored := *payout
	stored.UpdatedAt = time.Now()

	// Update indexes
	if payout.IdempotencyKey != "" {
		s.byIdempotencyKey[payout.IdempotencyKey] = payout.ID
	}
	if payout.ProviderPayoutID != "" {
		s.byProviderID[payout.ProviderPayoutID] = payout.ID
	}

	// Update account index
	if _, exists := s.payouts[payout.ID]; !exists {
		// New payout, add to account index
		s.byAccount[payout.AccountAddress] = append(s.byAccount[payout.AccountAddress], payout.ID)
	}

	s.payouts[payout.ID] = &stored

	return nil
}

// GetByID retrieves a payout by ID.
func (s *InMemoryPayoutStore) GetByID(ctx context.Context, id string) (*PayoutIntent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	payout, ok := s.payouts[id]
	if !ok {
		return nil, ErrPayoutNotFound
	}

	// Return a copy
	result := *payout
	return &result, nil
}

// GetByIdempotencyKey retrieves a payout by idempotency key.
func (s *InMemoryPayoutStore) GetByIdempotencyKey(ctx context.Context, key string) (*PayoutIntent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.byIdempotencyKey[key]
	if !ok {
		return nil, ErrPayoutNotFound
	}

	payout, ok := s.payouts[id]
	if !ok {
		return nil, ErrPayoutNotFound
	}

	result := *payout
	return &result, nil
}

// GetByProviderPayoutID retrieves a payout by provider payout ID.
func (s *InMemoryPayoutStore) GetByProviderPayoutID(ctx context.Context, providerPayoutID string) (*PayoutIntent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.byProviderID[providerPayoutID]
	if !ok {
		return nil, ErrPayoutNotFound
	}

	payout, ok := s.payouts[id]
	if !ok {
		return nil, ErrPayoutNotFound
	}

	result := *payout
	return &result, nil
}

// ListByAccount lists payouts for an account.
func (s *InMemoryPayoutStore) ListByAccount(ctx context.Context, accountAddress string, limit int) ([]*PayoutIntent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids, ok := s.byAccount[accountAddress]
	if !ok {
		return []*PayoutIntent{}, nil
	}

	result := make([]*PayoutIntent, 0, len(ids))
	for i := len(ids) - 1; i >= 0 && len(result) < limit; i-- {
		if payout, ok := s.payouts[ids[i]]; ok {
			copied := *payout
			result = append(result, &copied)
		}
	}

	return result, nil
}

// ListByStatus lists payouts by status.
func (s *InMemoryPayoutStore) ListByStatus(ctx context.Context, status PayoutStatus, limit int) ([]*PayoutIntent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*PayoutIntent, 0)
	for _, payout := range s.payouts {
		if payout.Status == status {
			copied := *payout
			result = append(result, &copied)
			if len(result) >= limit {
				break
			}
		}
	}

	return result, nil
}

// ListPendingReconciliation lists payouts pending reconciliation.
func (s *InMemoryPayoutStore) ListPendingReconciliation(ctx context.Context) ([]*PayoutIntent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*PayoutIntent, 0)
	for _, payout := range s.payouts {
		// Include completed payouts that haven't been reconciled
		if payout.Status == PayoutStatusSucceeded && payout.ProviderPayoutID != "" {
			copied := *payout
			result = append(result, &copied)
		}
	}

	return result, nil
}

// Delete deletes a payout (for testing only).
func (s *InMemoryPayoutStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	payout, ok := s.payouts[id]
	if !ok {
		return ErrPayoutNotFound
	}

	// Clean up indexes
	if payout.IdempotencyKey != "" {
		delete(s.byIdempotencyKey, payout.IdempotencyKey)
	}
	if payout.ProviderPayoutID != "" {
		delete(s.byProviderID, payout.ProviderPayoutID)
	}

	// Remove from account index
	accountPayouts := s.byAccount[payout.AccountAddress]
	for i, pid := range accountPayouts {
		if pid == id {
			s.byAccount[payout.AccountAddress] = append(accountPayouts[:i], accountPayouts[i+1:]...)
			break
		}
	}

	delete(s.payouts, id)

	return nil
}

// ============================================================================
// In-Memory Reconciliation Store
// ============================================================================

// InMemoryReconciliationStore implements ReconciliationStore with in-memory storage.
type InMemoryReconciliationStore struct {
	mu       sync.RWMutex
	records  map[string]*ReconciliationRecord
	byPayout map[string]string // payout_id -> record_id
}

// NewInMemoryReconciliationStore creates a new in-memory reconciliation store.
func NewInMemoryReconciliationStore() *InMemoryReconciliationStore {
	return &InMemoryReconciliationStore{
		records:  make(map[string]*ReconciliationRecord),
		byPayout: make(map[string]string),
	}
}

// Save saves a reconciliation record.
func (s *InMemoryReconciliationStore) Save(ctx context.Context, record *ReconciliationRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored := *record
	stored.UpdatedAt = time.Now()

	s.records[record.ID] = &stored
	s.byPayout[record.PayoutID] = record.ID

	return nil
}

// GetByPayoutID retrieves a reconciliation record by payout ID.
func (s *InMemoryReconciliationStore) GetByPayoutID(ctx context.Context, payoutID string) (*ReconciliationRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.byPayout[payoutID]
	if !ok {
		return nil, fmt.Errorf("reconciliation record not found for payout %s", payoutID)
	}

	record, ok := s.records[id]
	if !ok {
		return nil, fmt.Errorf("reconciliation record not found")
	}

	result := *record
	return &result, nil
}

// ListByStatus lists records by status.
func (s *InMemoryReconciliationStore) ListByStatus(ctx context.Context, status ReconciliationStatus) ([]*ReconciliationRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*ReconciliationRecord, 0)
	for _, record := range s.records {
		if record.Status == status {
			copied := *record
			result = append(result, &copied)
		}
	}

	return result, nil
}

// ListMismatches lists records with mismatches.
func (s *InMemoryReconciliationStore) ListMismatches(ctx context.Context) ([]*ReconciliationRecord, error) {
	return s.ListByStatus(ctx, ReconciliationMismatch)
}

// ============================================================================
// In-Memory Limits Store
// ============================================================================

// InMemoryLimitsStore implements LimitsStore with in-memory storage.
type InMemoryLimitsStore struct {
	mu     sync.RWMutex
	limits map[string]*PayoutLimits
	config LimitsConfig
}

// NewInMemoryLimitsStore creates a new in-memory limits store.
func NewInMemoryLimitsStore(config LimitsConfig) *InMemoryLimitsStore {
	return &InMemoryLimitsStore{
		limits: make(map[string]*PayoutLimits),
		config: config,
	}
}

// GetLimits retrieves limits for an account.
func (s *InMemoryLimitsStore) GetLimits(ctx context.Context, accountAddress string) (*PayoutLimits, error) {
	s.mu.RLock()
	limits, ok := s.limits[accountAddress]
	s.mu.RUnlock()

	if ok {
		result := *limits
		return &result, nil
	}

	// Create default limits for new account
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double check after acquiring write lock
	if limits, ok := s.limits[accountAddress]; ok {
		result := *limits
		return &result, nil
	}

	newLimits := &PayoutLimits{
		DailyLimit:          s.config.DefaultDailyLimit,
		MonthlyLimit:        s.config.DefaultMonthlyLimit,
		PerTransactionLimit: s.config.DefaultPerTransactionLimit,
		DailyUsed:           0,
		MonthlyUsed:         0,
		DailyRemaining:      s.config.DefaultDailyLimit,
		MonthlyRemaining:    s.config.DefaultMonthlyLimit,
		LastReset:           time.Now(),
	}

	s.limits[accountAddress] = newLimits

	result := *newLimits
	return &result, nil
}

// UpdateUsage updates the usage for an account.
func (s *InMemoryLimitsStore) UpdateUsage(ctx context.Context, accountAddress string, amount int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	limits, ok := s.limits[accountAddress]
	if !ok {
		// Create new limits
		limits = &PayoutLimits{
			DailyLimit:          s.config.DefaultDailyLimit,
			MonthlyLimit:        s.config.DefaultMonthlyLimit,
			PerTransactionLimit: s.config.DefaultPerTransactionLimit,
			LastReset:           time.Now(),
		}
		s.limits[accountAddress] = limits
	}

	limits.DailyUsed += amount
	limits.MonthlyUsed += amount
	limits.DailyRemaining = limits.DailyLimit - limits.DailyUsed
	limits.MonthlyRemaining = limits.MonthlyLimit - limits.MonthlyUsed

	if limits.DailyRemaining < 0 {
		limits.DailyRemaining = 0
	}
	if limits.MonthlyRemaining < 0 {
		limits.MonthlyRemaining = 0
	}

	return nil
}

// ResetDailyUsage resets daily usage for all accounts.
func (s *InMemoryLimitsStore) ResetDailyUsage(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, limits := range s.limits {
		limits.DailyUsed = 0
		limits.DailyRemaining = limits.DailyLimit
		limits.LastReset = time.Now()
	}

	return nil
}

// ResetMonthlyUsage resets monthly usage for all accounts.
func (s *InMemoryLimitsStore) ResetMonthlyUsage(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, limits := range s.limits {
		limits.MonthlyUsed = 0
		limits.MonthlyRemaining = limits.MonthlyLimit
		limits.LastReset = time.Now()
	}

	return nil
}

// ============================================================================
// Quote Store
// ============================================================================

// QuoteStore stores payout quotes.
type QuoteStore struct {
	mu     sync.RWMutex
	quotes map[string]*PayoutQuote
}

// NewQuoteStore creates a new quote store.
func NewQuoteStore() *QuoteStore {
	return &QuoteStore{
		quotes: make(map[string]*PayoutQuote),
	}
}

// Save saves a quote.
func (s *QuoteStore) Save(ctx context.Context, quote *PayoutQuote) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored := *quote
	s.quotes[quote.QuoteID] = &stored

	return nil
}

// Get retrieves a quote by ID.
func (s *QuoteStore) Get(ctx context.Context, quoteID string) (*PayoutQuote, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	quote, ok := s.quotes[quoteID]
	if !ok {
		return nil, fmt.Errorf("quote not found: %s", quoteID)
	}

	result := *quote
	return &result, nil
}

// Delete deletes a quote.
func (s *QuoteStore) Delete(ctx context.Context, quoteID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.quotes, quoteID)
	return nil
}

// CleanExpired removes expired quotes.
func (s *QuoteStore) CleanExpired(ctx context.Context) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	count := 0

	for id, quote := range s.quotes {
		if quote.IsExpired() || now.After(quote.ExpiresAt) {
			delete(s.quotes, id)
			count++
		}
	}

	return count
}

// Ensure implementations satisfy interfaces
var (
	_ PayoutStore         = (*InMemoryPayoutStore)(nil)
	_ ReconciliationStore = (*InMemoryReconciliationStore)(nil)
	_ LimitsStore         = (*InMemoryLimitsStore)(nil)
)
