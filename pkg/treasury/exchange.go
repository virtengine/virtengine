package treasury

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/virtengine/virtengine/pkg/dex"
)

// ExchangeAdapter abstracts DEX/CEX quote + execution.
type ExchangeAdapter interface {
	Name() string
	Type() AdapterType
	IsHealthy(ctx context.Context) bool
	GetQuote(ctx context.Context, req ExchangeRequest) (ExchangeQuote, error)
	Execute(ctx context.Context, quote ExchangeQuote) (ExchangeExecution, error)
}

// ExchangeRouter selects best quotes across adapters.
type ExchangeRouter struct {
	mu       sync.RWMutex
	adapters map[string]ExchangeAdapter
	policy   BestExecutionPolicy
}

// NewExchangeRouter creates a new router.
func NewExchangeRouter(policy BestExecutionPolicy) *ExchangeRouter {
	return &ExchangeRouter{
		adapters: make(map[string]ExchangeAdapter),
		policy:   policy,
	}
}

func (r *ExchangeRouter) RegisterAdapter(adapter ExchangeAdapter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[adapter.Name()] = adapter
}

func (r *ExchangeRouter) ListAdapters() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.adapters))
	for name := range r.adapters {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (r *ExchangeRouter) SelectBestQuote(ctx context.Context, req ExchangeRequest) (ExchangeQuote, error) {
	r.mu.RLock()
	adapters := make([]ExchangeAdapter, 0, len(r.adapters))
	for _, adapter := range r.adapters {
		adapters = append(adapters, adapter)
	}
	r.mu.RUnlock()

	if len(adapters) == 0 {
		return ExchangeQuote{}, ErrNoAdapters
	}

	quotes := make([]ExchangeQuote, 0)
	for _, adapter := range adapters {
		if r.policy.RequireHealthy && !adapter.IsHealthy(ctx) {
			continue
		}
		quote, err := adapter.GetQuote(ctx, req)
		if err != nil {
			continue
		}
		quote = ensureQuoteID(quote, adapter)
		if !r.policy.AllowExpired && time.Now().After(quote.ExpiresAt) {
			continue
		}
		if r.policy.MaxSlippageBps > 0 && quote.SlippageBps > r.policy.MaxSlippageBps {
			continue
		}
		if r.policy.MaxFeeBps > 0 && feeBps(quote) > r.policy.MaxFeeBps {
			continue
		}
		if !r.policy.MinOutputAmount.IsNil() && quote.OutputAmount.LT(r.policy.MinOutputAmount) {
			continue
		}
		quotes = append(quotes, quote)
	}

	if len(quotes) == 0 {
		return ExchangeQuote{}, ErrQuoteUnavailable
	}

	sort.SliceStable(quotes, func(i, j int) bool {
		left := netOutput(quotes[i])
		right := netOutput(quotes[j])
		if !left.Equal(right) {
			return left.GT(right)
		}
		return preferType(quotes[i].AdapterType, quotes[j].AdapterType, r.policy.PreferTypeOrder)
	})

	return quotes[0], nil
}

func (r *ExchangeRouter) ExecuteBestQuote(ctx context.Context, req ExchangeRequest) (ExchangeExecution, error) {
	quote, err := r.SelectBestQuote(ctx, req)
	if err != nil {
		return ExchangeExecution{}, err
	}
	r.mu.RLock()
	adapter := r.adapters[quote.AdapterName]
	r.mu.RUnlock()
	if adapter == nil {
		return ExchangeExecution{}, ErrExecutionUnavailable
	}
	return adapter.Execute(ctx, quote)
}

func netOutput(quote ExchangeQuote) sdkmath.Int {
	if quote.OutputAmount.IsNil() {
		return sdkmath.ZeroInt()
	}
	if quote.FeeAmount.IsNil() || quote.FeeAmount.IsZero() {
		return quote.OutputAmount
	}
	if quote.FeeAsset == "" || quote.FeeAsset == quote.ToAsset {
		return quote.OutputAmount.Sub(quote.FeeAmount)
	}
	return quote.OutputAmount
}

func feeBps(quote ExchangeQuote) int64 {
	if quote.InputAmount.IsNil() || quote.FeeAmount.IsNil() {
		return 0
	}
	if quote.InputAmount.IsZero() || quote.FeeAmount.IsZero() {
		return 0
	}
	feeDec := sdkmath.LegacyNewDecFromInt(quote.FeeAmount)
	inputDec := sdkmath.LegacyNewDecFromInt(quote.InputAmount)
	bps := feeDec.Quo(inputDec).Mul(sdkmath.LegacyNewDec(10000))
	return bps.TruncateInt64()
}

func preferType(left AdapterType, right AdapterType, order []AdapterType) bool {
	if len(order) == 0 || left == right {
		return false
	}
	rank := func(t AdapterType) int {
		for i, v := range order {
			if v == t {
				return i
			}
		}
		return len(order) + 1
	}
	return rank(left) < rank(right)
}

func ensureQuoteID(quote ExchangeQuote, adapter ExchangeAdapter) ExchangeQuote {
	if quote.ID == "" {
		quote.ID = fmt.Sprintf("%s-%d", adapter.Name(), time.Now().UnixNano())
	}
	return quote
}

// DexAdapterWrapper adapts pkg/dex adapters to ExchangeAdapter.
type DexAdapterWrapper struct {
	adapter dex.Adapter
}

func NewDexAdapterWrapper(adapter dex.Adapter) *DexAdapterWrapper {
	return &DexAdapterWrapper{adapter: adapter}
}

func (d *DexAdapterWrapper) Name() string {
	return d.adapter.Name()
}

func (d *DexAdapterWrapper) Type() AdapterType {
	return AdapterTypeDEX
}

func (d *DexAdapterWrapper) IsHealthy(ctx context.Context) bool {
	return d.adapter.IsHealthy(ctx)
}

func (d *DexAdapterWrapper) GetQuote(ctx context.Context, req ExchangeRequest) (ExchangeQuote, error) {
	swapReq := dex.SwapRequest{
		FromToken:         dex.Token{Symbol: req.FromAsset},
		ToToken:           dex.Token{Symbol: req.ToAsset},
		Amount:            req.Amount,
		Type:              dex.SwapTypeExactIn,
		SlippageTolerance: float64(req.SlippageBps) / 10000,
		Deadline:          deadlineOrDefault(req.Deadline),
		Sender:            "treasury-router",
		Recipient:         "",
	}

	quote, err := d.adapter.GetSwapQuote(ctx, swapReq)
	if err != nil {
		return ExchangeQuote{}, err
	}

	return ExchangeQuote{
		ID:           quote.ID,
		AdapterName:  d.adapter.Name(),
		AdapterType:  AdapterTypeDEX,
		FromAsset:    req.FromAsset,
		ToAsset:      req.ToAsset,
		InputAmount:  quote.InputAmount,
		OutputAmount: quote.OutputAmount,
		FeeAmount:    quote.TotalFee,
		FeeAsset:     req.FromAsset,
		SlippageBps:  int64(quote.PriceImpact * 10000),
		ExpiresAt:    quote.ExpiresAt,
		QuotedAt:     quote.CreatedAt,
	}, nil
}

func (d *DexAdapterWrapper) Execute(ctx context.Context, quote ExchangeQuote) (ExchangeExecution, error) {
	swapQuote := dex.SwapQuote{
		ID:           quote.ID,
		InputAmount:  quote.InputAmount,
		OutputAmount: quote.OutputAmount,
		TotalFee:     quote.FeeAmount,
		ExpiresAt:    quote.ExpiresAt,
		CreatedAt:    quote.QuotedAt,
	}

	result, err := d.adapter.ExecuteSwap(ctx, swapQuote, nil)
	if err != nil {
		return ExchangeExecution{}, err
	}

	return ExchangeExecution{
		Quote:       quote,
		TxID:        result.TxHash,
		FilledInput: result.InputAmount,
		FilledOut:   result.OutputAmount,
		ExecutedAt:  result.ExecutedAt,
	}, nil
}

// TestCEXAdapter is a deterministic CEX adapter for simulation/testing.
type TestCEXAdapter struct {
	name    string
	prices  map[string]sdkmath.LegacyDec
	feeBps  int64
	healthy bool
	latency time.Duration
}

func NewTestCEXAdapter(name string, prices map[string]sdkmath.LegacyDec, feeBps int64) *TestCEXAdapter {
	return &TestCEXAdapter{
		name:    name,
		prices:  prices,
		feeBps:  feeBps,
		healthy: true,
		latency: 50 * time.Millisecond,
	}
}

func (c *TestCEXAdapter) Name() string {
	return c.name
}

func (c *TestCEXAdapter) Type() AdapterType {
	return AdapterTypeCEX
}

func (c *TestCEXAdapter) IsHealthy(ctx context.Context) bool {
	return c.healthy
}

func (c *TestCEXAdapter) GetQuote(ctx context.Context, req ExchangeRequest) (ExchangeQuote, error) {
	price, ok := c.prices[req.FromAsset+"/"+req.ToAsset]
	if !ok {
		return ExchangeQuote{}, fmt.Errorf("pair not supported")
	}

	output := sdkmath.LegacyNewDecFromInt(req.Amount).Mul(price).TruncateInt()
	fee := sdkmath.LegacyNewDecFromInt(output).
		Mul(sdkmath.LegacyNewDecWithPrec(c.feeBps, 4)).
		TruncateInt()

	quotedAt := time.Now().UTC()

	return ExchangeQuote{
		ID:           fmt.Sprintf("%s-%d", c.name, quotedAt.UnixNano()),
		AdapterName:  c.name,
		AdapterType:  AdapterTypeCEX,
		FromAsset:    req.FromAsset,
		ToAsset:      req.ToAsset,
		InputAmount:  req.Amount,
		OutputAmount: output,
		FeeAmount:    fee,
		FeeAsset:     req.ToAsset,
		SlippageBps:  5,
		ExpiresAt:    quotedAt.Add(2 * time.Minute),
		QuotedAt:     quotedAt,
	}, nil
}

func (c *TestCEXAdapter) Execute(ctx context.Context, quote ExchangeQuote) (ExchangeExecution, error) {
	if c.latency > 0 {
		select {
		case <-time.After(c.latency):
		case <-ctx.Done():
			return ExchangeExecution{}, ctx.Err()
		}
	}

	return ExchangeExecution{
		Quote:       quote,
		TxID:        fmt.Sprintf("cex-%s", quote.ID),
		FilledInput: quote.InputAmount,
		FilledOut:   quote.OutputAmount,
		ExecutedAt:  time.Now().UTC(),
	}, nil
}

func deadlineOrDefault(deadline time.Time) time.Time {
	if deadline.IsZero() {
		return time.Now().Add(30 * time.Second)
	}
	return deadline
}
