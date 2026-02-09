package offramp

import (
	"context"
	"fmt"
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
}

// NewMockProvider creates a mock provider.
func NewMockProvider(name string, currencies []string, methods []string) *MockProvider {
	cur := make(map[string]bool)
	for _, c := range currencies {
		cur[c] = true
	}
	met := make(map[string]bool)
	for _, m := range methods {
		met[m] = true
	}

	return &MockProvider{
		name:       name,
		currencies: cur,
		methods:    met,
		feePercent: sdkmath.LegacyNewDecWithPrec(15, 3), // 1.5%
		healthy:    true,
		rateByFiat: map[string]sdkmath.LegacyDec{
			"USD": sdkmath.LegacyNewDec(1),
			"EUR": sdkmath.LegacyNewDecWithPrec(92, 2),
			"GBP": sdkmath.LegacyNewDecWithPrec(78, 2),
		},
	}
}

func (p *MockProvider) Name() string { return p.name }

func (p *MockProvider) GetQuote(ctx context.Context, req QuoteRequest) (Quote, error) {
	if !p.SupportsCurrency(req.FiatCurrency) || !p.SupportsMethod(req.PaymentMethod) {
		return Quote{}, fmt.Errorf("unsupported currency or method")
	}

	rate, ok := p.rateByFiat[req.FiatCurrency]
	if !ok {
		rate = sdkmath.LegacyOneDec()
	}

	fiatAmount := rate.MulInt(req.CryptoAmount)
	fee := p.feePercent.MulInt(req.CryptoAmount).TruncateInt()

	id, err := generateID("quote")
	if err != nil {
		return Quote{}, err
	}

	return Quote{
		ID:           id,
		Request:      req,
		FiatAmount:   fiatAmount,
		ExchangeRate: rate,
		Fee:          fee,
		Provider:     p.name,
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(5 * time.Minute),
	}, nil
}

func (p *MockProvider) InitiatePayout(ctx context.Context, req PayoutRequest) (PayoutResult, error) {
	id, err := generateID("payout")
	if err != nil {
		return PayoutResult{}, err
	}

	result := PayoutResult{
		ID:           id,
		QuoteID:      req.Quote.ID,
		Status:       StatusCompleted,
		Provider:     p.name,
		FiatAmount:   req.Quote.FiatAmount,
		CryptoAmount: req.Quote.Request.CryptoAmount,
		Fee:          req.Quote.Fee,
		Reference:    fmt.Sprintf("mock-%s", id),
		InitiatedAt:  time.Now().UTC(),
	}

	completedAt := time.Now().UTC()
	result.CompletedAt = &completedAt

	return result, nil
}

func (p *MockProvider) GetStatus(ctx context.Context, payoutID string) (PayoutResult, error) {
	return PayoutResult{}, fmt.Errorf("payout %s not found", payoutID)
}

func (p *MockProvider) Cancel(ctx context.Context, payoutID string) error {
	return nil
}

func (p *MockProvider) SupportsCurrency(currency string) bool {
	return p.currencies[currency]
}

func (p *MockProvider) SupportsMethod(method string) bool {
	return p.methods[method]
}

func (p *MockProvider) IsHealthy(ctx context.Context) bool {
	return p.healthy
}

// SetHealthy toggles mock provider health.
func (p *MockProvider) SetHealthy(healthy bool) {
	p.healthy = healthy
}
