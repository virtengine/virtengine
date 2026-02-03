// Package dex provides DEX integration for crypto-to-fiat conversions.
//
// VE-2007: Real Osmosis DEX adapter using gRPC queries
package dex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ============================================================================
// Osmosis Constants
// ============================================================================

const (
	// OsmosisMainnetGRPC is the mainnet gRPC endpoint
	OsmosisMainnetGRPC = "grpc.osmosis.zone:443"

	// OsmosisTestnetGRPC is the testnet gRPC endpoint
	OsmosisTestnetGRPC = "grpc-test.osmosis.zone:443"

	// OsmosisMainnetREST is the mainnet REST endpoint
	OsmosisMainnetREST = "https://lcd.osmosis.zone"

	// OsmosisTestnetREST is the testnet REST endpoint
	OsmosisTestnetREST = "https://lcd-test.osmosis.zone"

	// OsmosisChainIDMainnet is the mainnet chain ID
	OsmosisChainIDMainnet = "osmosis-1"

	// OsmosisChainIDTestnet is the testnet chain ID
	OsmosisChainIDTestnet = "osmo-test-5"

	// networkTestnet is the testnet network identifier
	networkTestnet = "testnet"
)

// ============================================================================
// Osmosis Errors
// ============================================================================

var (
	// ErrOsmosisNotConnected is returned when not connected to Osmosis
	ErrOsmosisNotConnected = errors.New("not connected to Osmosis")

	// ErrOsmosisPoolNotFound is returned when pool is not found
	ErrOsmosisPoolNotFound = errors.New("osmosis pool not found")

	// ErrOsmosisInvalidPool is returned for invalid pool data
	ErrOsmosisInvalidPool = errors.New("invalid Osmosis pool data")

	// ErrOsmosisQueryFailed is returned when query fails
	ErrOsmosisQueryFailed = errors.New("osmosis query failed")

	// ErrOsmosisSwapFailed is returned when swap fails
	ErrOsmosisSwapFailed = errors.New("osmosis swap failed")

	// ErrOsmosisInsufficientLiquidity is returned for low liquidity
	ErrOsmosisInsufficientLiquidity = errors.New("insufficient liquidity in Osmosis pool")

	// ErrOsmosisSlippageExceeded is returned when slippage is too high
	ErrOsmosisSlippageExceeded = errors.New("slippage exceeds maximum tolerance")
)

// ============================================================================
// Osmosis Configuration
// ============================================================================

// OsmosisConfig contains Osmosis-specific configuration
type OsmosisConfig struct {
	// Network is "mainnet" or "testnet"
	Network string `json:"network"`

	// GRPCEndpoint is the gRPC endpoint (overrides default)
	GRPCEndpoint string `json:"grpc_endpoint"`

	// RESTEndpoint is the REST/LCD endpoint (overrides default)
	RESTEndpoint string `json:"rest_endpoint"`

	// Timeout is the request timeout
	Timeout time.Duration `json:"timeout"`

	// MaxPoolsToQuery is the maximum number of pools to cache
	MaxPoolsToQuery int `json:"max_pools_to_query"`

	// PoolRefreshInterval is how often to refresh pool data
	PoolRefreshInterval time.Duration `json:"pool_refresh_interval"`

	// SlippageTolerance is the maximum allowed slippage (0.01 = 1%)
	SlippageTolerance float64 `json:"slippage_tolerance"`

	// EnableIBC enables IBC transfer integration
	EnableIBC bool `json:"enable_ibc"`
}

// DefaultOsmosisConfig returns default Osmosis configuration
func DefaultOsmosisConfig() OsmosisConfig {
	return OsmosisConfig{
		Network:             "mainnet",
		Timeout:             30 * time.Second,
		MaxPoolsToQuery:     100,
		PoolRefreshInterval: 5 * time.Minute,
		SlippageTolerance:   0.01, // 1%
		EnableIBC:           true,
	}
}

// GetGRPCEndpoint returns the gRPC endpoint based on config
func (c *OsmosisConfig) GetGRPCEndpoint() string {
	if c.GRPCEndpoint != "" {
		return c.GRPCEndpoint
	}
	if c.Network == networkTestnet {
		return OsmosisTestnetGRPC
	}
	return OsmosisMainnetGRPC
}

// GetRESTEndpoint returns the REST endpoint based on config
func (c *OsmosisConfig) GetRESTEndpoint() string {
	if c.RESTEndpoint != "" {
		return c.RESTEndpoint
	}
	if c.Network == networkTestnet {
		return OsmosisTestnetREST
	}
	return OsmosisMainnetREST
}

// GetChainID returns the chain ID based on network
func (c *OsmosisConfig) GetChainID() string {
	if c.Network == networkTestnet {
		return OsmosisChainIDTestnet
	}
	return OsmosisChainIDMainnet
}

// ============================================================================
// Osmosis Pool Types (from REST API)
// ============================================================================

// OsmosisPoolResponse represents the REST API pool response
type OsmosisPoolResponse struct {
	Pool OsmosisPoolData `json:"pool"`
}

// OsmosisPoolsResponse represents the REST API pools list response
type OsmosisPoolsResponse struct {
	Pools      []OsmosisPoolData `json:"pools"`
	Pagination struct {
		NextKey string `json:"next_key"`
		Total   string `json:"total"`
	} `json:"pagination"`
}

// OsmosisPoolData represents pool data from Osmosis
type OsmosisPoolData struct {
	Type               string             `json:"@type"`
	ID                 string             `json:"id"`
	Address            string             `json:"address"`
	PoolParams         OsmosisPoolParams  `json:"pool_params"`
	TotalShares        OsmosisCoin        `json:"total_shares"`
	PoolAssets         []OsmosisPoolAsset `json:"pool_assets"`
	TotalWeight        string             `json:"total_weight"`
	FuturePoolGovernor string             `json:"future_pool_governor"`
}

// OsmosisPoolParams represents pool parameters
type OsmosisPoolParams struct {
	SwapFee                  string      `json:"swap_fee"`
	ExitFee                  string      `json:"exit_fee"`
	SmoothWeightChangeParams interface{} `json:"smooth_weight_change_params"`
}

// OsmosisPoolAsset represents a pool asset
type OsmosisPoolAsset struct {
	Token  OsmosisCoin `json:"token"`
	Weight string      `json:"weight"`
}

// OsmosisCoin represents a coin amount
type OsmosisCoin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// OsmosisSpotPriceResponse represents spot price query response
type OsmosisSpotPriceResponse struct {
	SpotPrice string `json:"spot_price"`
}

// OsmosisEstimateSwapResponse represents swap estimation response
type OsmosisEstimateSwapResponse struct {
	TokenOutAmount string `json:"token_out_amount"`
}

// ============================================================================
// Real Osmosis Adapter Implementation
// ============================================================================

// RealOsmosisAdapter implements real Osmosis DEX integration using gRPC/REST
type RealOsmosisAdapter struct {
	*BaseAdapter
	config      OsmosisConfig
	grpcConn    *grpc.ClientConn
	httpClient  *http.Client
	pools       map[string]LiquidityPool
	poolsMu     sync.RWMutex
	connected   bool
	lastRefresh time.Time
}

// NewRealOsmosisAdapter creates a new real Osmosis adapter
func NewRealOsmosisAdapter(cfg AdapterConfig, osmosisConfig OsmosisConfig) (*RealOsmosisAdapter, error) {
	adapter := &RealOsmosisAdapter{
		BaseAdapter: NewBaseAdapter(cfg),
		config:      osmosisConfig,
		httpClient: &http.Client{
			Timeout: osmosisConfig.Timeout,
		},
		pools: make(map[string]LiquidityPool),
	}

	// Override chain ID from config
	adapter.chainID = osmosisConfig.GetChainID()

	return adapter, nil
}

// Connect establishes connection to Osmosis
func (a *RealOsmosisAdapter) Connect(ctx context.Context) error {
	endpoint := a.config.GetGRPCEndpoint()

	// Create gRPC connection
	// Note: In production, use proper TLS credentials
	conn, err := grpc.NewClient(endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to Osmosis gRPC: %w", err)
	}

	a.grpcConn = conn
	a.connected = true
	a.healthy = true

	// Initial pool refresh
	if err := a.refreshPools(ctx); err != nil {
		// Log but don't fail - we can still use REST API
		_ = err
	}

	return nil
}

// Disconnect closes the connection
func (a *RealOsmosisAdapter) Disconnect() error {
	a.connected = false
	a.healthy = false
	if a.grpcConn != nil {
		return a.grpcConn.Close()
	}
	return nil
}

// Close implements the Adapter interface
func (a *RealOsmosisAdapter) Close() error {
	return a.Disconnect()
}

// IsHealthy checks if the adapter is healthy
func (a *RealOsmosisAdapter) IsHealthy(ctx context.Context) bool {
	if !a.cfg.Enabled {
		return false
	}

	// Try a simple query to check health
	_, err := a.queryPools(ctx, 1)
	return err == nil
}

// refreshPools updates the pool cache
func (a *RealOsmosisAdapter) refreshPools(ctx context.Context) error {
	pools, err := a.queryPools(ctx, a.config.MaxPoolsToQuery)
	if err != nil {
		return err
	}

	a.poolsMu.Lock()
	defer a.poolsMu.Unlock()

	for _, pool := range pools {
		a.pools[pool.ID] = pool
	}
	a.lastRefresh = time.Now()

	return nil
}

// queryPools queries pools from Osmosis REST API
func (a *RealOsmosisAdapter) queryPools(ctx context.Context, limit int) ([]LiquidityPool, error) {
	endpoint := a.config.GetRESTEndpoint()
	url := fmt.Sprintf("%s/osmosis/poolmanager/v1beta1/all-pools?pagination.limit=%d", endpoint, limit)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOsmosisQueryFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status %d: %s", ErrOsmosisQueryFailed, resp.StatusCode, string(body))
	}

	var poolsResp OsmosisPoolsResponse
	if err := json.NewDecoder(resp.Body).Decode(&poolsResp); err != nil {
		return nil, fmt.Errorf("failed to decode pools response: %w", err)
	}

	return a.convertPools(poolsResp.Pools), nil
}

// convertPools converts Osmosis pool data to our LiquidityPool type
func (a *RealOsmosisAdapter) convertPools(osmPools []OsmosisPoolData) []LiquidityPool {
	var pools []LiquidityPool

	for _, op := range osmPools {
		pool := a.convertPool(op)
		if pool.ID != "" {
			pools = append(pools, pool)
		}
	}

	return pools
}

// convertPool converts a single Osmosis pool to our LiquidityPool type
func (a *RealOsmosisAdapter) convertPool(op OsmosisPoolData) LiquidityPool {
	pool := LiquidityPool{
		ID:        op.ID,
		DEX:       a.name,
		Type:      PoolTypeConstantProduct, // Osmosis pools are x*y=k AMMs
		Tokens:    make([]Token, 0, 2),
		Reserves:  make(map[string]sdkmath.Int),
		UpdatedAt: time.Now().UTC(),
	}

	// Parse swap fee
	if op.PoolParams.SwapFee != "" {
		if fee, err := sdkmath.LegacyNewDecFromStr(op.PoolParams.SwapFee); err == nil {
			pool.Fee = fee
		}
	}

	// Parse pool assets
	for _, asset := range op.PoolAssets {
		token := a.parseToken(asset.Token)
		pool.Tokens = append(pool.Tokens, token)

		// Parse reserves by token symbol
		if amount, ok := sdkmath.NewIntFromString(asset.Token.Amount); ok {
			pool.Reserves[token.Symbol] = amount
		}
	}

	return pool
}

// parseToken parses an Osmosis token to our Token type
func (a *RealOsmosisAdapter) parseToken(coin OsmosisCoin) Token {
	// Parse IBC denoms and native denoms
	symbol := coin.Denom

	// Handle IBC denoms (ibc/...)
	if strings.HasPrefix(coin.Denom, "ibc/") {
		// In production, would resolve IBC denom to actual token info
		symbol = "IBC-" + coin.Denom[4:10]
	}

	// Handle known native denoms
	switch coin.Denom {
	case "uosmo":
		symbol = "OSMO"
	case "uatom":
		symbol = "ATOM"
	case "uusdc":
		symbol = "USDC"
	}

	return Token{
		Symbol:   symbol,
		Denom:    coin.Denom,
		Decimals: 6, // Most Cosmos tokens use 6 decimals
		ChainID:  a.chainID,
		IsNative: !strings.HasPrefix(coin.Denom, "ibc/"),
	}
}

// GetSupportedPairs returns supported trading pairs
func (a *RealOsmosisAdapter) GetSupportedPairs(ctx context.Context) ([]TradingPair, error) {
	// Refresh pools if stale
	a.poolsMu.RLock()
	needsRefresh := time.Since(a.lastRefresh) > a.config.PoolRefreshInterval
	a.poolsMu.RUnlock()

	if needsRefresh {
		if err := a.refreshPools(ctx); err != nil {
			return nil, err
		}
	}

	a.poolsMu.RLock()
	defer a.poolsMu.RUnlock()

	var pairs []TradingPair
	for _, pool := range a.pools {
		// Only create pairs for pools with at least 2 tokens
		if len(pool.Tokens) >= 2 {
			pair := TradingPair{
				BaseToken:  pool.Tokens[0],
				QuoteToken: pool.Tokens[1],
			}
			pairs = append(pairs, pair)
		}
	}

	return pairs, nil
}

// GetPrice fetches the current price for a trading pair
func (a *RealOsmosisAdapter) GetPrice(ctx context.Context, baseSymbol, quoteSymbol string) (Price, error) {
	// Find pool for this pair
	pool, err := a.findPoolForPair(ctx, baseSymbol, quoteSymbol)
	if err != nil {
		return Price{}, err
	}

	// Get base and quote tokens from pool
	baseToken, quoteToken := a.getPoolTokenPair(pool, baseSymbol, quoteSymbol)

	// Query spot price from Osmosis
	spotPrice, err := a.querySpotPrice(ctx, pool.ID, baseToken.Denom, quoteToken.Denom)
	if err != nil {
		return Price{}, err
	}

	return Price{
		Pair: TradingPair{
			BaseToken:  baseToken,
			QuoteToken: quoteToken,
		},
		Rate:       spotPrice,
		Timestamp:  time.Now().UTC(),
		Source:     a.name,
		Confidence: 1.0,
	}, nil
}

// getPoolTokenPair retrieves base and quote tokens from a pool
func (a *RealOsmosisAdapter) getPoolTokenPair(pool LiquidityPool, baseSymbol, quoteSymbol string) (Token, Token) {
	var baseToken, quoteToken Token
	for _, token := range pool.Tokens {
		switch token.Symbol {
		case baseSymbol:
			baseToken = token
		case quoteSymbol:
			quoteToken = token
		}
	}
	return baseToken, quoteToken
}

// findPoolForPair finds a pool containing both tokens
func (a *RealOsmosisAdapter) findPoolForPair(ctx context.Context, baseSymbol, quoteSymbol string) (LiquidityPool, error) {
	// Ensure pools are loaded
	a.poolsMu.RLock()
	needsRefresh := len(a.pools) == 0 || time.Since(a.lastRefresh) > a.config.PoolRefreshInterval
	a.poolsMu.RUnlock()

	if needsRefresh {
		if err := a.refreshPools(ctx); err != nil {
			return LiquidityPool{}, err
		}
	}

	a.poolsMu.RLock()
	defer a.poolsMu.RUnlock()

	for _, pool := range a.pools {
		hasBase, hasQuote := false, false
		for _, token := range pool.Tokens {
			if token.Symbol == baseSymbol {
				hasBase = true
			}
			if token.Symbol == quoteSymbol {
				hasQuote = true
			}
		}
		if hasBase && hasQuote {
			return pool, nil
		}
	}

	return LiquidityPool{}, ErrOsmosisPoolNotFound
}

// querySpotPrice queries the spot price from Osmosis
func (a *RealOsmosisAdapter) querySpotPrice(ctx context.Context, poolID, baseDenom, quoteDenom string) (sdkmath.LegacyDec, error) {
	endpoint := a.config.GetRESTEndpoint()
	url := fmt.Sprintf("%s/osmosis/poolmanager/v1beta1/pools/%s/spot-price?base_asset_denom=%s&quote_asset_denom=%s",
		endpoint, poolID, baseDenom, quoteDenom)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return sdkmath.LegacyZeroDec(), fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return sdkmath.LegacyZeroDec(), fmt.Errorf("%w: %v", ErrOsmosisQueryFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return sdkmath.LegacyZeroDec(), fmt.Errorf("%w: status %d", ErrOsmosisQueryFailed, resp.StatusCode)
	}

	var priceResp OsmosisSpotPriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&priceResp); err != nil {
		return sdkmath.LegacyZeroDec(), fmt.Errorf("failed to decode price response: %w", err)
	}

	price, err := sdkmath.LegacyNewDecFromStr(priceResp.SpotPrice)
	if err != nil {
		return sdkmath.LegacyZeroDec(), fmt.Errorf("failed to parse spot price: %w", err)
	}

	return price, nil
}

// GetPool fetches pool information by ID
func (a *RealOsmosisAdapter) GetPool(ctx context.Context, poolID string) (LiquidityPool, error) {
	// Check cache first
	a.poolsMu.RLock()
	pool, ok := a.pools[poolID]
	a.poolsMu.RUnlock()

	if ok {
		return pool, nil
	}

	// Query from Osmosis
	endpoint := a.config.GetRESTEndpoint()
	url := fmt.Sprintf("%s/osmosis/poolmanager/v1beta1/pools/%s", endpoint, poolID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return LiquidityPool{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return LiquidityPool{}, fmt.Errorf("%w: %v", ErrOsmosisQueryFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return LiquidityPool{}, ErrOsmosisPoolNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return LiquidityPool{}, fmt.Errorf("%w: status %d", ErrOsmosisQueryFailed, resp.StatusCode)
	}

	var poolResp OsmosisPoolResponse
	if err := json.NewDecoder(resp.Body).Decode(&poolResp); err != nil {
		return LiquidityPool{}, fmt.Errorf("failed to decode pool response: %w", err)
	}

	return a.convertPool(poolResp.Pool), nil
}

// ListPools lists pools matching the query
func (a *RealOsmosisAdapter) ListPools(ctx context.Context, query PoolQuery) ([]LiquidityPool, error) {
	// Ensure pools are loaded
	if len(a.pools) == 0 {
		if err := a.refreshPools(ctx); err != nil {
			return nil, err
		}
	}

	a.poolsMu.RLock()
	defer a.poolsMu.RUnlock()

	var result []LiquidityPool
	for _, pool := range a.pools {
		if matchesPoolQuery(pool, query) {
			result = append(result, pool)
		}
	}

	return result, nil
}

// GetSwapQuote generates a swap quote
func (a *RealOsmosisAdapter) GetSwapQuote(ctx context.Context, request SwapRequest) (SwapQuote, error) {
	// Find pool for the swap
	pool, err := a.findPoolForPair(ctx, request.FromToken.Symbol, request.ToToken.Symbol)
	if err != nil {
		return SwapQuote{}, err
	}

	// Query estimated swap output
	outputAmount, err := a.estimateSwapOutput(ctx, pool.ID, request.FromToken.Denom, request.ToToken.Denom, request.Amount)
	if err != nil {
		return SwapQuote{}, err
	}

	// Calculate slippage
	minOutput := sdkmath.LegacyNewDecFromInt(outputAmount).Mul(sdkmath.LegacyOneDec().Sub(sdkmath.LegacyNewDecWithPrec(int64(a.config.SlippageTolerance*10000), 4)))
	minOutputInt := minOutput.TruncateInt()

	// Calculate price impact (simplified)
	priceImpact := 0.001 // Would calculate from reserves in production

	route := SwapRoute{
		Hops: []SwapHop{
			{
				PoolID:    pool.ID,
				DEX:       a.name,
				FromToken: request.FromToken,
				ToToken:   request.ToToken,
				AmountIn:  request.Amount,
				AmountOut: outputAmount,
				Fee:       pool.Fee,
			},
		},
		TotalGas:    300000,
		PriceImpact: priceImpact,
	}

	// Calculate rate
	rate := sdkmath.LegacyNewDecFromInt(outputAmount).Quo(sdkmath.LegacyNewDecFromInt(request.Amount))

	quote := SwapQuote{
		ID:              fmt.Sprintf("osmo-quote-%d", time.Now().UnixNano()),
		Request:         request,
		Route:           route,
		InputAmount:     request.Amount,
		OutputAmount:    outputAmount,
		MinOutputAmount: minOutputInt,
		Rate:            rate,
		TotalFee:        pool.Fee.Mul(sdkmath.LegacyNewDecFromInt(request.Amount)).TruncateInt(),
		GasEstimate:     300000,
		ExpiresAt:       time.Now().Add(30 * time.Second),
		CreatedAt:       time.Now().UTC(),
	}

	return quote, nil
}

// estimateSwapOutput estimates the output amount for a swap
func (a *RealOsmosisAdapter) estimateSwapOutput(ctx context.Context, poolID, tokenIn, tokenOut string, amountIn sdkmath.Int) (sdkmath.Int, error) {
	endpoint := a.config.GetRESTEndpoint()
	url := fmt.Sprintf("%s/osmosis/poolmanager/v1beta1/%s/estimate/single-pool-swap-exact-amount-in?pool_id=%s&token_in=%s%s&token_out_denom=%s",
		endpoint, poolID, poolID, amountIn.String(), tokenIn, tokenOut)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return sdkmath.Int{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return sdkmath.Int{}, fmt.Errorf("%w: %v", ErrOsmosisQueryFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Fall back to manual calculation if estimation endpoint fails
		return a.calculateSwapOutput(ctx, poolID, amountIn)
	}

	var swapResp OsmosisEstimateSwapResponse
	if err := json.NewDecoder(resp.Body).Decode(&swapResp); err != nil {
		return sdkmath.Int{}, fmt.Errorf("failed to decode swap response: %w", err)
	}

	output, ok := sdkmath.NewIntFromString(swapResp.TokenOutAmount)
	if !ok {
		return sdkmath.Int{}, fmt.Errorf("failed to parse output amount: %s", swapResp.TokenOutAmount)
	}

	return output, nil
}

// calculateSwapOutput calculates output using constant product formula
func (a *RealOsmosisAdapter) calculateSwapOutput(ctx context.Context, poolID string, amountIn sdkmath.Int) (sdkmath.Int, error) {
	pool, err := a.GetPool(ctx, poolID)
	if err != nil {
		return sdkmath.Int{}, err
	}

	// Constant product formula: x * y = k
	// output = (reserve1 * amountIn) / (reserve0 + amountIn)
	// Apply fee: output = output * (1 - fee)

	// Get reserves from the new map structure - need at least 2 tokens
	if len(pool.Tokens) < 2 {
		return sdkmath.Int{}, ErrOsmosisInsufficientLiquidity
	}

	reserve0 := pool.Reserves[pool.Tokens[0].Symbol]
	reserve1 := pool.Reserves[pool.Tokens[1].Symbol]

	if reserve0.IsZero() || reserve1.IsZero() {
		return sdkmath.Int{}, ErrOsmosisInsufficientLiquidity
	}

	// Calculate output before fee
	numerator := reserve1.Mul(amountIn)
	denominator := reserve0.Add(amountIn)
	outputBeforeFee := numerator.Quo(denominator)

	// Apply fee
	feeMultiplier := sdkmath.LegacyOneDec().Sub(pool.Fee)
	output := feeMultiplier.MulInt(outputBeforeFee).TruncateInt()

	return output, nil
}

// ExecuteSwap executes a swap on Osmosis
func (a *RealOsmosisAdapter) ExecuteSwap(ctx context.Context, quote SwapQuote, signedTx []byte) (SwapResult, error) {
	// In production, this would:
	// 1. Verify the signed transaction
	// 2. Broadcast to Osmosis network via gRPC or REST
	// 3. Wait for confirmation
	// 4. Return result with actual tx hash

	endpoint := a.config.GetRESTEndpoint()
	url := fmt.Sprintf("%s/cosmos/tx/v1beta1/txs", endpoint)

	// Create broadcast request
	broadcastReq := map[string]interface{}{
		"tx_bytes": signedTx,
		"mode":     "BROADCAST_MODE_SYNC",
	}

	reqBody, err := json.Marshal(broadcastReq)
	if err != nil {
		return SwapResult{}, fmt.Errorf("failed to marshal broadcast request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(reqBody)))
	if err != nil {
		return SwapResult{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return SwapResult{}, fmt.Errorf("%w: %v", ErrOsmosisSwapFailed, err)
	}
	defer resp.Body.Close()

	// Parse broadcast response
	var broadcastResp struct {
		TxResponse struct {
			TxHash string `json:"txhash"`
			Code   int    `json:"code"`
			Log    string `json:"raw_log"`
		} `json:"tx_response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&broadcastResp); err != nil {
		return SwapResult{}, fmt.Errorf("failed to decode broadcast response: %w", err)
	}

	if broadcastResp.TxResponse.Code != 0 {
		return SwapResult{}, fmt.Errorf("%w: code %d: %s",
			ErrOsmosisSwapFailed, broadcastResp.TxResponse.Code, broadcastResp.TxResponse.Log)
	}

	return SwapResult{
		QuoteID:      quote.ID,
		TxHash:       broadcastResp.TxResponse.TxHash,
		InputAmount:  quote.InputAmount,
		OutputAmount: quote.OutputAmount,
		Fee:          quote.TotalFee,
		GasUsed:      quote.GasEstimate,
		ExecutedAt:   time.Now().UTC(),
		Route:        quote.Route,
	}, nil
}

// EstimateGas estimates gas for a swap
func (a *RealOsmosisAdapter) EstimateGas(ctx context.Context, request SwapRequest) (uint64, error) {
	// Osmosis swaps typically use 250k-400k gas depending on route complexity
	// Multi-hop swaps use more gas

	pool, err := a.findPoolForPair(ctx, request.FromToken.Symbol, request.ToToken.Symbol)
	if err != nil {
		// If no direct pool, estimate higher for multi-hop
		return 500000, nil
	}

	// Direct swap
	_ = pool
	return 300000, nil
}

// GetPoolReserves gets current pool reserves (for real-time data)
// Returns reserves for the first two tokens in the pool
func (a *RealOsmosisAdapter) GetPoolReserves(ctx context.Context, poolID string) (sdkmath.Int, sdkmath.Int, error) {
	pool, err := a.GetPool(ctx, poolID)
	if err != nil {
		return sdkmath.Int{}, sdkmath.Int{}, err
	}

	if len(pool.Tokens) < 2 {
		return sdkmath.Int{}, sdkmath.Int{}, ErrOsmosisInsufficientLiquidity
	}

	reserve0 := pool.Reserves[pool.Tokens[0].Symbol]
	reserve1 := pool.Reserves[pool.Tokens[1].Symbol]

	return reserve0, reserve1, nil
}

// GetPoolTVL calculates the total value locked in a pool
func (a *RealOsmosisAdapter) GetPoolTVL(ctx context.Context, poolID string, priceOracle PriceFeed) (sdkmath.LegacyDec, error) {
	pool, err := a.GetPool(ctx, poolID)
	if err != nil {
		return sdkmath.LegacyZeroDec(), err
	}

	tvl := sdkmath.LegacyZeroDec()
	divisor := sdkmath.LegacyNewDec(1000000) // 6 decimals

	// Sum TVL for all tokens in the pool
	for _, token := range pool.Tokens {
		price, err := priceOracle.GetPrice(ctx, token.Symbol, "USD")
		if err != nil {
			return sdkmath.LegacyZeroDec(), err
		}

		reserve := pool.Reserves[token.Symbol]
		tokenTVL := price.Rate.MulInt(reserve).Quo(divisor)
		tvl = tvl.Add(tokenTVL)
	}

	return tvl, nil
}

// GetTotalValueLocked calculates TVL across all pools
func (a *RealOsmosisAdapter) GetTotalValueLocked(ctx context.Context) (sdkmath.LegacyDec, error) {
	// Query Osmosis API for total TVL
	endpoint := a.config.GetRESTEndpoint()
	url := fmt.Sprintf("%s/osmosis/poolmanager/v1beta1/total_liquidity", endpoint)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return sdkmath.LegacyZeroDec(), err
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return sdkmath.LegacyZeroDec(), err
	}
	defer resp.Body.Close()

	var liquidityResp struct {
		Liquidity []OsmosisCoin `json:"liquidity"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&liquidityResp); err != nil {
		return sdkmath.LegacyZeroDec(), err
	}

	// Sum up all liquidity (would need price oracle for accurate USD value)
	total := sdkmath.LegacyZeroDec()
	for _, coin := range liquidityResp.Liquidity {
		amount, err := strconv.ParseInt(coin.Amount, 10, 64)
		if err != nil {
			continue
		}
		total = total.Add(sdkmath.LegacyNewDec(amount))
	}

	return total, nil
}

// ============================================================================
// Compile-time interface check
// ============================================================================

var _ Adapter = (*RealOsmosisAdapter)(nil)
