package offramp

import (
	"context"
	"fmt"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"
)

type testClock struct {
	mu  sync.Mutex
	now time.Time
}

func newTestClock(start time.Time) *testClock {
	return &testClock{now: start.UTC()}
}

func (c *testClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *testClock) Advance(delta time.Duration) time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(delta)
	return c.now
}

type providerStep struct {
	result PayoutResult
	err    error
}

type contractProvider struct {
	name        string
	currencies  map[string]bool
	methods     map[string]bool
	healthy     bool
	quote       Quote
	quoteErr    error
	initSteps   []providerStep
	statusSteps map[string][]PayoutResult
	statusErr   error
	lookup      map[string]PayoutResult
	payouts     map[string]PayoutResult
	normalize   func(operation string, err error) error
	cancelErr   error
	now         func() time.Time
	mu          sync.Mutex

	quoteCalls  int
	initCalls   int
	statusCalls int
	lookupCalls int
	cancelCalls int
}

func newContractProvider(name string, now func() time.Time, quote Quote) *contractProvider {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &contractProvider{
		name:        name,
		currencies:  map[string]bool{"USD": true},
		methods:     map[string]bool{"bank_transfer": true},
		healthy:     true,
		quote:       quote,
		statusSteps: make(map[string][]PayoutResult),
		lookup:      make(map[string]PayoutResult),
		payouts:     make(map[string]PayoutResult),
		now:         now,
	}
}

func (p *contractProvider) Name() string { return p.name }

func (p *contractProvider) GetQuote(ctx context.Context, req QuoteRequest) (Quote, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.quoteCalls++
	if p.quoteErr != nil {
		return Quote{}, p.quoteErr
	}

	quote := cloneQuote(p.quote)
	quote.Request = req
	if quote.Provider == "" {
		quote.Provider = p.name
	}
	if quote.ID == "" {
		quote.ID = fmt.Sprintf("%s-contract-quote", p.name)
	}
	now := p.now().UTC()
	if quote.CreatedAt.IsZero() {
		quote.CreatedAt = now
	}
	if quote.ExpiresAt.IsZero() {
		quote.ExpiresAt = now.Add(defaultQuoteValidity)
	}
	if quote.FiatAmount.IsNil() || !quote.FiatAmount.IsPositive() {
		quote.FiatAmount = sdkmath.LegacyNewDec(100)
	}
	if quote.ExchangeRate.IsNil() || !quote.ExchangeRate.IsPositive() {
		quote.ExchangeRate = sdkmath.LegacyOneDec()
	}
	quote.AuditFields = mergeStringMaps(quote.AuditFields, map[string]string{
		"provider_type": "contract",
	})
	return quote, nil
}

func (p *contractProvider) InitiatePayout(ctx context.Context, req PayoutRequest) (PayoutResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.initCalls++

	if len(p.initSteps) > 0 {
		step := p.initSteps[0]
		p.initSteps = p.initSteps[1:]
		if step.err != nil {
			if step.result.ID != "" {
				result := p.normalizeResult(step.result, req)
				p.storeResultLocked(result)
			}
			return PayoutResult{}, step.err
		}
		result := p.normalizeResult(step.result, req)
		p.storeResultLocked(result)
		return clonePayoutResult(result), nil
	}

	result := p.normalizeResult(PayoutResult{
		ID:         fmt.Sprintf("%s-contract-payout", p.name),
		Status:     StatusProcessing,
		Reference:  fmt.Sprintf("%s-ref", p.name),
		FiatAmount: req.Quote.FiatAmount,
	}, req)
	p.storeResultLocked(result)
	return clonePayoutResult(result), nil
}

func (p *contractProvider) GetStatus(ctx context.Context, payoutID string) (PayoutResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.statusCalls++

	if p.statusErr != nil {
		return PayoutResult{}, p.statusErr
	}

	steps := p.statusSteps[payoutID]
	if len(steps) > 0 {
		result := clonePayoutResult(steps[0])
		p.statusSteps[payoutID] = steps[1:]
		if result.ID == "" {
			result.ID = payoutID
		}
		if existing, ok := p.payouts[payoutID]; ok {
			if result.QuoteID == "" {
				result.QuoteID = existing.QuoteID
			}
			if len(result.Metadata) == 0 {
				result.Metadata = cloneStringMap(existing.Metadata)
			}
		}
		result.Provider = p.name
		if result.StatusUpdatedAt.IsZero() {
			result.StatusUpdatedAt = p.now().UTC()
		}
		p.storeResultLocked(result)
		return result, nil
	}

	result, ok := p.payouts[payoutID]
	if !ok {
		return PayoutResult{}, ErrPayoutNotFound
	}
	return clonePayoutResult(result), nil
}

func (p *contractProvider) FindPayoutByMetadata(ctx context.Context, metadata map[string]string) (PayoutResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.lookupCalls++

	if key := metadataLookupKey(p.name, metadata); key != "" {
		if result, ok := p.lookup[key]; ok {
			result = p.normalizeLookupResult(result)
			p.storeResultLocked(result)
			return clonePayoutResult(result), nil
		}
	}

	for _, result := range p.payouts {
		if metadataMatches(result.Metadata, metadata) {
			return clonePayoutResult(result), nil
		}
	}
	return PayoutResult{}, ErrPayoutNotFound
}

func (p *contractProvider) Cancel(ctx context.Context, payoutID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cancelCalls++

	if p.cancelErr != nil {
		return p.cancelErr
	}

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
	p.storeResultLocked(result)
	return nil
}

func (p *contractProvider) SupportsCurrency(currency string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.currencies[currency]
}

func (p *contractProvider) SupportsMethod(method string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.methods[method]
}

func (p *contractProvider) IsHealthy(ctx context.Context) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.healthy
}

func (p *contractProvider) NormalizeError(operation string, err error) error {
	if p.normalize != nil {
		return p.normalize(operation, err)
	}
	return err
}

func (p *contractProvider) normalizeResult(result PayoutResult, req PayoutRequest) PayoutResult {
	now := p.now().UTC()
	if result.ID == "" {
		result.ID = fmt.Sprintf("%s-contract-payout", p.name)
	}
	if result.Provider == "" {
		result.Provider = p.name
	}
	if result.QuoteID == "" {
		result.QuoteID = req.Quote.ID
	}
	if result.FiatAmount.IsNil() || !result.FiatAmount.IsPositive() {
		result.FiatAmount = req.Quote.FiatAmount
	}
	if result.CryptoAmount.IsNil() || !result.CryptoAmount.IsPositive() {
		result.CryptoAmount = req.Quote.Request.CryptoAmount
	}
	if result.Fee.IsNil() {
		result.Fee = req.Quote.Fee
	}
	if result.Reference == "" {
		result.Reference = fmt.Sprintf("%s-ref", p.name)
	}
	if result.Metadata == nil {
		result.Metadata = cloneStringMap(req.Metadata)
	}
	result.Metadata = mergeStringMaps(result.Metadata, req.Metadata)
	if result.InitiatedAt.IsZero() {
		result.InitiatedAt = now
	}
	if result.Status == "" {
		result.Status = StatusProcessing
	}
	if result.StatusUpdatedAt.IsZero() {
		result.StatusUpdatedAt = now
	}
	if result.Status == StatusCompleted && result.CompletedAt == nil {
		completedAt := now
		result.CompletedAt = &completedAt
	}
	result.AuditFields = mergeStringMaps(result.AuditFields, map[string]string{
		"provider_type": "contract",
	})
	return result
}

func (p *contractProvider) normalizeLookupResult(result PayoutResult) PayoutResult {
	now := p.now().UTC()
	if result.Provider == "" {
		result.Provider = p.name
	}
	if result.StatusUpdatedAt.IsZero() {
		result.StatusUpdatedAt = now
	}
	result.AuditFields = mergeStringMaps(result.AuditFields, map[string]string{
		"provider_type": "contract",
	})
	return result
}

func (p *contractProvider) storeResultLocked(result PayoutResult) {
	cloned := clonePayoutResult(result)
	p.payouts[cloned.ID] = cloned
	if key := metadataLookupKey(p.name, cloned.Metadata); key != "" {
		p.lookup[key] = cloned
	}
}
