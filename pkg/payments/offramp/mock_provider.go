package offramp

import (
	"context"
	"fmt"
	"maps"
	"sort"
	"strconv"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"
)

// MockProvider is a deterministic off-ramp adapter for testing.
type MockProvider struct {
	name       string
	currencies map[string]bool
	methods    map[string]bool
	feePercent sdkmath.LegacyDec
	healthy    bool
	rateByFiat map[string]sdkmath.LegacyDec
	payouts    map[string]PayoutResult
	lookupKeys map[string]string
	quoteSeq   int
	payoutSeq  int
	now        func() time.Time
	mu         sync.RWMutex
}

// NewMockProvider creates a mock provider.
func NewMockProvider(name string, currencies []string, methods []string) *MockProvider {
	return newMockProviderWithClock(name, currencies, methods, func() time.Time {
		return time.Now().UTC()
	})
}

func newMockProviderWithClock(name string, currencies []string, methods []string, now func() time.Time) *MockProvider {
	cur := make(map[string]bool)
	for _, c := range currencies {
		cur[c] = true
	}
	met := make(map[string]bool)
	for _, m := range methods {
		met[m] = true
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}

	return &MockProvider{
		name:       name,
		currencies: cur,
		methods:    met,
		feePercent: sdkmath.LegacyNewDecWithPrec(15, 3),
		healthy:    true,
		rateByFiat: map[string]sdkmath.LegacyDec{
			"USD": sdkmath.LegacyNewDec(1),
			"EUR": sdkmath.LegacyNewDecWithPrec(92, 2),
			"GBP": sdkmath.LegacyNewDecWithPrec(78, 2),
		},
		payouts:    make(map[string]PayoutResult),
		lookupKeys: make(map[string]string),
		now:        now,
	}
}

func (p *MockProvider) Name() string { return p.name }

func (p *MockProvider) GetQuote(ctx context.Context, req QuoteRequest) (Quote, error) {
	if err := validateQuoteRequest(req); err != nil {
		return Quote{}, err
	}
	if !p.SupportsCurrency(req.FiatCurrency) || !p.SupportsMethod(req.PaymentMethod) {
		return Quote{}, ErrInvalidRequest
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	rate, ok := p.rateByFiat[req.FiatCurrency]
	if !ok {
		rate = sdkmath.LegacyOneDec()
	}

	fiatAmount := rate.MulInt(req.CryptoAmount)
	fee := p.feePercent.MulInt(req.CryptoAmount).TruncateInt()
	p.quoteSeq++
	now := p.now().UTC()

	return Quote{
		ID:           fmt.Sprintf("%s-quote-%03d", p.name, p.quoteSeq),
		Request:      req,
		FiatAmount:   fiatAmount,
		ExchangeRate: rate,
		Fee:          fee,
		Provider:     p.name,
		CreatedAt:    now,
		ExpiresAt:    now.Add(defaultQuoteValidity),
		AuditFields: map[string]string{
			"provider_type":  "mock",
			"quote_sequence": strconv.Itoa(p.quoteSeq),
		},
	}, nil
}

func (p *MockProvider) InitiatePayout(ctx context.Context, req PayoutRequest) (PayoutResult, error) {
	now := p.now().UTC()
	if err := validatePayoutInputs(req.Quote, req.CryptoTxRef, req.Destination, now); err != nil {
		return PayoutResult{}, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if key := metadataLookupKey(p.name, req.Metadata); key != "" {
		if payoutID, ok := p.lookupKeys[key]; ok {
			if existing, exists := p.payouts[payoutID]; exists {
				return clonePayoutResult(existing), nil
			}
		}
	}

	p.payoutSeq++
	id := fmt.Sprintf("%s-payout-%03d", p.name, p.payoutSeq)
	completedAt := now
	result := PayoutResult{
		ID:              id,
		QuoteID:         req.Quote.ID,
		Status:          StatusCompleted,
		Provider:        p.name,
		FiatAmount:      req.Quote.FiatAmount,
		CryptoAmount:    req.Quote.Request.CryptoAmount,
		Fee:             req.Quote.Fee,
		Reference:       fmt.Sprintf("mock-%s", id),
		Metadata:        maps.Clone(req.Metadata),
		InitiatedAt:     now,
		CompletedAt:     &completedAt,
		StatusUpdatedAt: now,
		Retryable:       false,
		AuditFields: map[string]string{
			"provider_type":   "mock",
			"payout_sequence": strconv.Itoa(p.payoutSeq),
		},
	}

	p.payouts[result.ID] = clonePayoutResult(result)
	cacheMetadataKeys(p.lookupKeys, result)
	return clonePayoutResult(result), nil
}

func (p *MockProvider) GetStatus(ctx context.Context, payoutID string) (PayoutResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result, ok := p.payouts[payoutID]
	if !ok {
		return PayoutResult{}, ErrPayoutNotFound
	}
	return clonePayoutResult(result), nil
}

func (p *MockProvider) FindPayoutByMetadata(ctx context.Context, metadata map[string]string) (PayoutResult, error) {
	if len(metadata) == 0 {
		return PayoutResult{}, ErrInvalidRequest
	}

	key := metadataLookupKey(p.name, metadata)
	p.mu.RLock()
	if payoutID, ok := p.lookupKeys[key]; ok {
		if result, exists := p.payouts[payoutID]; exists {
			p.mu.RUnlock()
			return clonePayoutResult(result), nil
		}
	}

	ids := make([]string, 0, len(p.payouts))
	for id := range p.payouts {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		result := p.payouts[id]
		if metadataMatches(result.Metadata, metadata) {
			p.mu.RUnlock()
			return clonePayoutResult(result), nil
		}
	}
	p.mu.RUnlock()

	return PayoutResult{}, ErrPayoutNotFound
}

func (p *MockProvider) Cancel(ctx context.Context, payoutID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	result, ok := p.payouts[payoutID]
	if !ok {
		return ErrPayoutNotFound
	}
	if result.Status == StatusCancelled {
		return nil
	}
	if result.Status.IsTerminal() {
		return ErrPayoutNotCancellable
	}

	result.Status = StatusCancelled
	result.StatusUpdatedAt = p.now().UTC()
	result.Retryable = false
	result.AuditFields = mergeStringMaps(result.AuditFields, map[string]string{
		"provider_type":     "mock",
		"cancelled_by_mock": "true",
	})
	p.payouts[payoutID] = clonePayoutResult(result)
	cacheMetadataKeys(p.lookupKeys, result)
	return nil
}

func (p *MockProvider) SupportsCurrency(currency string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.currencies[currency]
}

func (p *MockProvider) SupportsMethod(method string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.methods[method]
}

func (p *MockProvider) IsHealthy(ctx context.Context) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.healthy
}

// SetHealthy toggles mock provider health.
func (p *MockProvider) SetHealthy(healthy bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.healthy = healthy
}
