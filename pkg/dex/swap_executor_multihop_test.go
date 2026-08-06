package dex

import (
	"context"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
)

type mockPairQuote struct {
	poolID      string
	numerator   int64
	denominator int64
}

type mockMultiHopAdapter struct {
	name  string
	pools []LiquidityPool
	pairs map[string]mockPairQuote
}

func (m *mockMultiHopAdapter) Name() string { return m.name }

func (m *mockMultiHopAdapter) Type() string { return "mock" }

func (m *mockMultiHopAdapter) ChainID() string { return "virtengine-test" }

func (m *mockMultiHopAdapter) IsHealthy(context.Context) bool { return true }

func (m *mockMultiHopAdapter) GetSupportedPairs(context.Context) ([]TradingPair, error) {
	pairs := make([]TradingPair, 0, len(m.pairs))
	for key := range m.pairs {
		from, to, ok := strings.Cut(key, "->")
		if !ok {
			continue
		}
		pairs = append(pairs, TradingPair{
			BaseToken:  Token{Symbol: from},
			QuoteToken: Token{Symbol: to},
		})
	}
	return pairs, nil
}

func (m *mockMultiHopAdapter) GetPrice(context.Context, string, string) (Price, error) {
	return Price{}, ErrUnsupportedPair
}

func (m *mockMultiHopAdapter) GetPool(context.Context, string) (LiquidityPool, error) {
	return LiquidityPool{}, ErrAdapterNotFound
}

func (m *mockMultiHopAdapter) ListPools(_ context.Context, query PoolQuery) ([]LiquidityPool, error) {
	if query.DEX != "" && query.DEX != m.name {
		return nil, nil
	}

	filtered := make([]LiquidityPool, 0, len(m.pools))
	for _, pool := range m.pools {
		if len(query.TokenSymbols) > 0 {
			matchesAll := true
			for _, symbol := range query.TokenSymbols {
				if !poolContainsToken(pool, symbol) {
					matchesAll = false
					break
				}
			}
			if !matchesAll {
				continue
			}
		}
		filtered = append(filtered, pool)
	}

	return filtered, nil
}

func (m *mockMultiHopAdapter) GetSwapQuote(_ context.Context, request SwapRequest) (SwapQuote, error) {
	quote, ok := m.pairs[testPairKey(request.FromToken.Symbol, request.ToToken.Symbol)]
	if !ok {
		return SwapQuote{}, ErrUnsupportedPair
	}

	var amountIn, amountOut sdkmath.Int
	switch request.Type {
	case SwapTypeExactOut:
		amountOut = request.Amount
		amountIn = ceilMulDiv(request.Amount, quote.denominator, quote.numerator)
	default:
		amountIn = request.Amount
		amountOut = request.Amount.MulRaw(quote.numerator).QuoRaw(quote.denominator)
	}

	route := SwapRoute{
		Hops: []SwapHop{
			{
				PoolID:    quote.poolID,
				DEX:       m.name,
				FromToken: request.FromToken,
				ToToken:   request.ToToken,
				AmountIn:  amountIn,
				AmountOut: amountOut,
				Fee:       sdkmath.LegacyMustNewDecFromStr("0.003"),
			},
		},
		TotalGas:    100_000,
		PriceImpact: 0.01,
	}

	return SwapQuote{
		Request:      request,
		Route:        route,
		InputAmount:  amountIn,
		OutputAmount: amountOut,
		ExpiresAt:    time.Now().Add(time.Minute),
		CreatedAt:    time.Now(),
	}, nil
}

func (m *mockMultiHopAdapter) ExecuteSwap(context.Context, SwapQuote, []byte) (SwapResult, error) {
	return SwapResult{}, nil
}

func (m *mockMultiHopAdapter) EstimateGas(context.Context, SwapRequest) (uint64, error) {
	return 100_000, nil
}

func (m *mockMultiHopAdapter) Close() error { return nil }

func TestFindBestRoute_FindsCrossDEXMultiHopForExactIn(t *testing.T) {
	cfg := DefaultConfig()
	svc := &service{
		cfg:      cfg,
		adapters: map[string]Adapter{},
		started:  true,
	}

	svc.adapters["direct"] = &mockMultiHopAdapter{
		name: "direct",
		pools: []LiquidityPool{
			{
				ID:     "pool-direct",
				DEX:    "direct",
				Tokens: []Token{{Symbol: "UVE"}, {Symbol: "USDC"}},
			},
		},
		pairs: map[string]mockPairQuote{
			testPairKey("UVE", "USDC"): {poolID: "pool-direct", numerator: 1, denominator: 1},
		},
	}
	svc.adapters["dex-a"] = &mockMultiHopAdapter{
		name: "dex-a",
		pools: []LiquidityPool{
			{
				ID:     "pool-uve-atom",
				DEX:    "dex-a",
				Tokens: []Token{{Symbol: "UVE"}, {Symbol: "ATOM"}},
			},
		},
		pairs: map[string]mockPairQuote{
			testPairKey("UVE", "ATOM"): {poolID: "pool-uve-atom", numerator: 2, denominator: 1},
		},
	}
	svc.adapters["dex-b"] = &mockMultiHopAdapter{
		name: "dex-b",
		pools: []LiquidityPool{
			{
				ID:     "pool-atom-usdc",
				DEX:    "dex-b",
				Tokens: []Token{{Symbol: "ATOM"}, {Symbol: "USDC"}},
			},
		},
		pairs: map[string]mockPairQuote{
			testPairKey("ATOM", "USDC"): {poolID: "pool-atom-usdc", numerator: 1, denominator: 1},
		},
	}

	executor := newSwapExecutor(cfg.Swap, svc)
	route, err := executor.FindBestRoute(context.Background(), SwapRequest{
		FromToken: Token{Symbol: "UVE"},
		ToToken:   Token{Symbol: "USDC"},
		Amount:    sdkmath.NewInt(100),
		Type:      SwapTypeExactIn,
		Sender:    "ve1sender",
	})
	if err != nil {
		t.Fatalf("FindBestRoute() error = %v", err)
	}

	if len(route.Hops) != 2 {
		t.Fatalf("expected 2-hop route, got %d", len(route.Hops))
	}
	if route.Hops[0].DEX != "dex-a" || route.Hops[1].DEX != "dex-b" {
		t.Fatalf("expected cross-DEX route dex-a -> dex-b, got %s -> %s", route.Hops[0].DEX, route.Hops[1].DEX)
	}
	if got := executor.calculateRouteOutput(route); !got.Equal(sdkmath.NewInt(200)) {
		t.Fatalf("calculateRouteOutput() = %s, want 200", got)
	}
}

func TestFindBestRoute_FindsCrossDEXMultiHopForExactOut(t *testing.T) {
	cfg := DefaultConfig()
	svc := &service{
		cfg:      cfg,
		adapters: map[string]Adapter{},
		started:  true,
	}

	svc.adapters["direct"] = &mockMultiHopAdapter{
		name: "direct",
		pools: []LiquidityPool{
			{
				ID:     "pool-direct",
				DEX:    "direct",
				Tokens: []Token{{Symbol: "UVE"}, {Symbol: "USDC"}},
			},
		},
		pairs: map[string]mockPairQuote{
			testPairKey("UVE", "USDC"): {poolID: "pool-direct", numerator: 1, denominator: 1},
		},
	}
	svc.adapters["dex-a"] = &mockMultiHopAdapter{
		name: "dex-a",
		pools: []LiquidityPool{
			{
				ID:     "pool-uve-atom",
				DEX:    "dex-a",
				Tokens: []Token{{Symbol: "UVE"}, {Symbol: "ATOM"}},
			},
		},
		pairs: map[string]mockPairQuote{
			testPairKey("UVE", "ATOM"): {poolID: "pool-uve-atom", numerator: 2, denominator: 1},
		},
	}
	svc.adapters["dex-b"] = &mockMultiHopAdapter{
		name: "dex-b",
		pools: []LiquidityPool{
			{
				ID:     "pool-atom-usdc",
				DEX:    "dex-b",
				Tokens: []Token{{Symbol: "ATOM"}, {Symbol: "USDC"}},
			},
		},
		pairs: map[string]mockPairQuote{
			testPairKey("ATOM", "USDC"): {poolID: "pool-atom-usdc", numerator: 1, denominator: 1},
		},
	}

	executor := newSwapExecutor(cfg.Swap, svc)
	route, err := executor.FindBestRoute(context.Background(), SwapRequest{
		FromToken: Token{Symbol: "UVE"},
		ToToken:   Token{Symbol: "USDC"},
		Amount:    sdkmath.NewInt(150),
		Type:      SwapTypeExactOut,
		Sender:    "ve1sender",
	})
	if err != nil {
		t.Fatalf("FindBestRoute() error = %v", err)
	}

	if len(route.Hops) != 2 {
		t.Fatalf("expected 2-hop route, got %d", len(route.Hops))
	}
	if got := executor.calculateRouteInput(route); !got.Equal(sdkmath.NewInt(75)) {
		t.Fatalf("calculateRouteInput() = %s, want 75", got)
	}
}

func testPairKey(from, to string) string {
	return from + "->" + to
}

func ceilMulDiv(value sdkmath.Int, numerator, denominator int64) sdkmath.Int {
	product := value.MulRaw(numerator)
	quotient := product.QuoRaw(denominator)
	if product.ModRaw(denominator).IsZero() {
		return quotient
	}
	return quotient.AddRaw(1)
}
