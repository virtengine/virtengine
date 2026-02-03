// Package dex provides DEX integration for crypto-to-fiat conversions.
//
// VE-905: DEX integration for crypto-to-fiat conversions
package dex

import (
	"context"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
)

// BaseAdapter provides a base implementation for DEX adapters
type BaseAdapter struct {
	name      string
	adpType   string
	chainID   string
	cfg       AdapterConfig
	healthy   bool
	pools     map[string]LiquidityPool
	pairCache map[string]TradingPair
}

// NewBaseAdapter creates a new base adapter
func NewBaseAdapter(cfg AdapterConfig) *BaseAdapter {
	return &BaseAdapter{
		name:      cfg.Name,
		adpType:   cfg.Type,
		chainID:   cfg.ChainID,
		cfg:       cfg,
		healthy:   true,
		pools:     make(map[string]LiquidityPool),
		pairCache: make(map[string]TradingPair),
	}
}

func (a *BaseAdapter) Name() string    { return a.name }
func (a *BaseAdapter) Type() string    { return a.adpType }
func (a *BaseAdapter) ChainID() string { return a.chainID }

func (a *BaseAdapter) IsHealthy(ctx context.Context) bool {
	return a.healthy && a.cfg.Enabled
}

func (a *BaseAdapter) Close() error {
	a.healthy = false
	return nil
}

// ============================================================================
// Uniswap V2 Adapter
// ============================================================================

// UniswapV2Adapter implements the Adapter interface for Uniswap V2
type UniswapV2Adapter struct {
	*BaseAdapter
	factoryAddress string
	routerAddress  string
}

// NewUniswapV2Adapter creates a new Uniswap V2 adapter
func NewUniswapV2Adapter(cfg AdapterConfig) (*UniswapV2Adapter, error) {
	factoryAddr, ok := cfg.ContractAddresses["factory"]
	if !ok {
		return nil, fmt.Errorf("factory address required for Uniswap V2 adapter")
	}
	routerAddr, ok := cfg.ContractAddresses["router"]
	if !ok {
		return nil, fmt.Errorf("router address required for Uniswap V2 adapter")
	}

	return &UniswapV2Adapter{
		BaseAdapter:    NewBaseAdapter(cfg),
		factoryAddress: factoryAddr,
		routerAddress:  routerAddr,
	}, nil
}

func (a *UniswapV2Adapter) GetSupportedPairs(ctx context.Context) ([]TradingPair, error) {
	// In production, this would query the factory contract for all pairs
	pairs := make([]TradingPair, 0, len(a.pairCache))
	for _, pair := range a.pairCache {
		pairs = append(pairs, pair)
	}
	return pairs, nil
}

func (a *UniswapV2Adapter) GetPrice(ctx context.Context, baseSymbol, quoteSymbol string) (Price, error) {
	// In production, this would query the pair contract reserves
	// and calculate price from the constant product formula
	return Price{
		Pair: TradingPair{
			BaseToken:  Token{Symbol: baseSymbol},
			QuoteToken: Token{Symbol: quoteSymbol},
		},
		Rate:       sdkmath.LegacyOneDec(), // Placeholder
		Timestamp:  time.Now().UTC(),
		Source:     a.name,
		Confidence: 1.0,
	}, nil
}

func (a *UniswapV2Adapter) GetPool(ctx context.Context, poolID string) (LiquidityPool, error) {
	pool, ok := a.pools[poolID]
	if !ok {
		return LiquidityPool{}, fmt.Errorf("pool %s not found", poolID)
	}
	return pool, nil
}

func (a *UniswapV2Adapter) ListPools(ctx context.Context, query PoolQuery) ([]LiquidityPool, error) {
	var result []LiquidityPool
	for _, pool := range a.pools {
		if matchesPoolQuery(pool, query) {
			result = append(result, pool)
		}
	}
	return result, nil
}

func (a *UniswapV2Adapter) GetSwapQuote(ctx context.Context, request SwapRequest) (SwapQuote, error) {
	// In production, this would:
	// 1. Find the pair contract
	// 2. Get reserves
	// 3. Calculate output using getAmountOut
	// 4. Build the swap route

	route := SwapRoute{
		Hops: []SwapHop{
			{
				PoolID:    fmt.Sprintf("%s-%s", request.FromToken.Symbol, request.ToToken.Symbol),
				DEX:       a.name,
				FromToken: request.FromToken,
				ToToken:   request.ToToken,
				AmountIn:  request.Amount,
				AmountOut: request.Amount,                     // Placeholder
				Fee:       sdkmath.LegacyNewDecWithPrec(3, 3), // 0.3%
			},
		},
		TotalGas:    150000,
		PriceImpact: 0.001,
	}

	return SwapQuote{
		Request:         request,
		Route:           route,
		InputAmount:     request.Amount,
		OutputAmount:    request.Amount,
		MinOutputAmount: request.Amount,
		Rate:            sdkmath.LegacyOneDec(),
		ExpiresAt:       time.Now().Add(30 * time.Second),
		CreatedAt:       time.Now().UTC(),
	}, nil
}

func (a *UniswapV2Adapter) ExecuteSwap(ctx context.Context, quote SwapQuote, signedTx []byte) (SwapResult, error) {
	// In production, this would broadcast the signed transaction
	return SwapResult{
		QuoteID:      quote.ID,
		TxHash:       "0x...",
		InputAmount:  quote.InputAmount,
		OutputAmount: quote.OutputAmount,
		Fee:          quote.TotalFee,
		GasUsed:      quote.GasEstimate,
		ExecutedAt:   time.Now().UTC(),
		Route:        quote.Route,
	}, nil
}

func (a *UniswapV2Adapter) EstimateGas(ctx context.Context, request SwapRequest) (uint64, error) {
	return 150000, nil
}

// ============================================================================
// Osmosis Adapter
// ============================================================================

// OsmosisAdapter implements the Adapter interface for Osmosis DEX
type OsmosisAdapter struct {
	*BaseAdapter
	grpcEndpoint string
}

// NewOsmosisAdapter creates a new Osmosis adapter
func NewOsmosisAdapter(cfg AdapterConfig) (*OsmosisAdapter, error) {
	return &OsmosisAdapter{
		BaseAdapter:  NewBaseAdapter(cfg),
		grpcEndpoint: cfg.RPCEndpoint,
	}, nil
}

func (a *OsmosisAdapter) GetSupportedPairs(ctx context.Context) ([]TradingPair, error) {
	// In production, this would query Osmosis gamm module for pools
	pairs := make([]TradingPair, 0, len(a.pairCache))
	for _, pair := range a.pairCache {
		pairs = append(pairs, pair)
	}
	return pairs, nil
}

func (a *OsmosisAdapter) GetPrice(ctx context.Context, baseSymbol, quoteSymbol string) (Price, error) {
	// In production, this would query pool spot price
	return Price{
		Pair: TradingPair{
			BaseToken:  Token{Symbol: baseSymbol},
			QuoteToken: Token{Symbol: quoteSymbol},
		},
		Rate:       sdkmath.LegacyOneDec(),
		Timestamp:  time.Now().UTC(),
		Source:     a.name,
		Confidence: 1.0,
	}, nil
}

func (a *OsmosisAdapter) GetPool(ctx context.Context, poolID string) (LiquidityPool, error) {
	pool, ok := a.pools[poolID]
	if !ok {
		return LiquidityPool{}, fmt.Errorf("pool %s not found", poolID)
	}
	return pool, nil
}

func (a *OsmosisAdapter) ListPools(ctx context.Context, query PoolQuery) ([]LiquidityPool, error) {
	var result []LiquidityPool
	for _, pool := range a.pools {
		if matchesPoolQuery(pool, query) {
			result = append(result, pool)
		}
	}
	return result, nil
}

func (a *OsmosisAdapter) GetSwapQuote(ctx context.Context, request SwapRequest) (SwapQuote, error) {
	route := SwapRoute{
		Hops: []SwapHop{
			{
				PoolID:    "1", // Osmosis pool ID
				DEX:       a.name,
				FromToken: request.FromToken,
				ToToken:   request.ToToken,
				AmountIn:  request.Amount,
				AmountOut: request.Amount,
				Fee:       sdkmath.LegacyNewDecWithPrec(2, 3), // 0.2%
			},
		},
		TotalGas:    300000,
		PriceImpact: 0.001,
	}

	return SwapQuote{
		Request:         request,
		Route:           route,
		InputAmount:     request.Amount,
		OutputAmount:    request.Amount,
		MinOutputAmount: request.Amount,
		Rate:            sdkmath.LegacyOneDec(),
		ExpiresAt:       time.Now().Add(30 * time.Second),
		CreatedAt:       time.Now().UTC(),
	}, nil
}

func (a *OsmosisAdapter) ExecuteSwap(ctx context.Context, quote SwapQuote, signedTx []byte) (SwapResult, error) {
	return SwapResult{
		QuoteID:      quote.ID,
		TxHash:       "osmosis...",
		InputAmount:  quote.InputAmount,
		OutputAmount: quote.OutputAmount,
		Fee:          quote.TotalFee,
		GasUsed:      quote.GasEstimate,
		ExecutedAt:   time.Now().UTC(),
		Route:        quote.Route,
	}, nil
}

func (a *OsmosisAdapter) EstimateGas(ctx context.Context, request SwapRequest) (uint64, error) {
	return 300000, nil
}

// ============================================================================
// Curve Adapter
// ============================================================================

// CurveAdapter implements the Adapter interface for Curve Finance
type CurveAdapter struct {
	*BaseAdapter
	registryAddress string
}

// NewCurveAdapter creates a new Curve adapter
func NewCurveAdapter(cfg AdapterConfig) (*CurveAdapter, error) {
	registryAddr, ok := cfg.ContractAddresses["registry"]
	if !ok {
		return nil, fmt.Errorf("registry address required for Curve adapter")
	}

	return &CurveAdapter{
		BaseAdapter:     NewBaseAdapter(cfg),
		registryAddress: registryAddr,
	}, nil
}

func (a *CurveAdapter) GetSupportedPairs(ctx context.Context) ([]TradingPair, error) {
	pairs := make([]TradingPair, 0, len(a.pairCache))
	for _, pair := range a.pairCache {
		pairs = append(pairs, pair)
	}
	return pairs, nil
}

func (a *CurveAdapter) GetPrice(ctx context.Context, baseSymbol, quoteSymbol string) (Price, error) {
	return Price{
		Pair: TradingPair{
			BaseToken:  Token{Symbol: baseSymbol},
			QuoteToken: Token{Symbol: quoteSymbol},
		},
		Rate:       sdkmath.LegacyOneDec(),
		Timestamp:  time.Now().UTC(),
		Source:     a.name,
		Confidence: 1.0,
	}, nil
}

func (a *CurveAdapter) GetPool(ctx context.Context, poolID string) (LiquidityPool, error) {
	pool, ok := a.pools[poolID]
	if !ok {
		return LiquidityPool{}, fmt.Errorf("pool %s not found", poolID)
	}
	return pool, nil
}

func (a *CurveAdapter) ListPools(ctx context.Context, query PoolQuery) ([]LiquidityPool, error) {
	var result []LiquidityPool
	for _, pool := range a.pools {
		if matchesPoolQuery(pool, query) {
			result = append(result, pool)
		}
	}
	return result, nil
}

func (a *CurveAdapter) GetSwapQuote(ctx context.Context, request SwapRequest) (SwapQuote, error) {
	route := SwapRoute{
		Hops: []SwapHop{
			{
				PoolID:    "3pool",
				DEX:       a.name,
				FromToken: request.FromToken,
				ToToken:   request.ToToken,
				AmountIn:  request.Amount,
				AmountOut: request.Amount,
				Fee:       sdkmath.LegacyNewDecWithPrec(4, 4), // 0.04%
			},
		},
		TotalGas:    250000,
		PriceImpact: 0.0001,
	}

	return SwapQuote{
		Request:         request,
		Route:           route,
		InputAmount:     request.Amount,
		OutputAmount:    request.Amount,
		MinOutputAmount: request.Amount,
		Rate:            sdkmath.LegacyOneDec(),
		ExpiresAt:       time.Now().Add(30 * time.Second),
		CreatedAt:       time.Now().UTC(),
	}, nil
}

func (a *CurveAdapter) ExecuteSwap(ctx context.Context, quote SwapQuote, signedTx []byte) (SwapResult, error) {
	return SwapResult{
		QuoteID:      quote.ID,
		TxHash:       "0x...",
		InputAmount:  quote.InputAmount,
		OutputAmount: quote.OutputAmount,
		Fee:          quote.TotalFee,
		GasUsed:      quote.GasEstimate,
		ExecutedAt:   time.Now().UTC(),
		Route:        quote.Route,
	}, nil
}

func (a *CurveAdapter) EstimateGas(ctx context.Context, request SwapRequest) (uint64, error) {
	return 250000, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// matchesPoolQuery checks if a pool matches a query
func matchesPoolQuery(pool LiquidityPool, query PoolQuery) bool {
	if query.DEX != "" && pool.DEX != query.DEX {
		return false
	}
	if query.PoolType != "" && pool.Type != query.PoolType {
		return false
	}
	if !query.MinLiquidity.IsNil() && pool.TotalLiquidity.LT(query.MinLiquidity) {
		return false
	}
	if len(query.TokenSymbols) > 0 {
		tokenSet := make(map[string]bool)
		for _, t := range pool.Tokens {
			tokenSet[t.Symbol] = true
		}
		for _, sym := range query.TokenSymbols {
			if !tokenSet[sym] {
				return false
			}
		}
	}
	return true
}

// CreateAdapter creates an adapter based on the config type
func CreateAdapter(cfg AdapterConfig) (Adapter, error) {
	switch cfg.Type {
	case "uniswap_v2":
		return NewUniswapV2Adapter(cfg)
	case "osmosis":
		return NewOsmosisAdapter(cfg)
	case "curve":
		return NewCurveAdapter(cfg)
	default:
		return nil, fmt.Errorf("unsupported adapter type: %s", cfg.Type)
	}
}
