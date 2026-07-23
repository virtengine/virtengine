// Package dex provides DEX integration for crypto-to-fiat conversions.
//
// VE-2007: Real Osmosis DEX adapter using gRPC queries
package dex

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/virtengine/virtengine/pkg/security"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// ============================================================================
// Osmosis Constants
// ============================================================================

const (
	// OsmosisAdapterVersion is the only pool implementation whose math is
	// currently implemented and covered by golden fixtures.
	OsmosisAdapterVersion = "gamm-v1beta1-cp-equal-weight-v1"
	osmosisGAMMPoolType   = "/osmosis.gamm.v1beta1.Pool"

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

	maxOsmosisResponseBytes int64 = 2 << 20
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

	// RouteProfile is the exact support-matrix row used for production safety.
	RouteProfile *DEXRouteProfile `json:"-"`

	// ValidationMode controls the explicit engineering-only opt-in.
	ValidationMode RouteValidationMode `json:"-"`

	// Process-boundary dependencies are injected for deterministic tests.
	PoolState         OsmosisPoolStateProvider  `json:"-"`
	ChainEvidence     ChainEvidenceProvider     `json:"-"`
	Oracle            OraclePriceProvider       `json:"-"`
	ExecutionVerifier ExecutionEnvelopeVerifier `json:"-"`
	RouteAuthorizer   RouteProfileAuthorizer    `json:"-"`
	HTTPClient        *http.Client              `json:"-"`
	Now               func() time.Time          `json:"-"`
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

// OsmosisEstimateSwapResponse is retained for compatibility with external
// response fixtures. Real quote math does not trust or consume this endpoint.
type OsmosisEstimateSwapResponse struct {
	TokenOutAmount string `json:"token_out_amount"`
}

// ============================================================================
// Real Osmosis Adapter Implementation
// ============================================================================

// RealOsmosisAdapter implements real Osmosis DEX integration using gRPC/REST
type RealOsmosisAdapter struct {
	*BaseAdapter
	config            OsmosisConfig
	grpcConn          *grpc.ClientConn
	httpClient        *http.Client
	pools             map[string]LiquidityPool
	poolsMu           sync.RWMutex
	connected         bool
	lastRefresh       time.Time
	lifecycleMu       sync.Mutex
	profile           DEXRouteProfile
	mode              RouteValidationMode
	poolState         OsmosisPoolStateProvider
	evidence          ChainEvidenceProvider
	oracle            OraclePriceProvider
	executionVerifier ExecutionEnvelopeVerifier
	now               func() time.Time
}

// NewRealOsmosisAdapter creates a new real Osmosis adapter
func NewRealOsmosisAdapter(cfg AdapterConfig, osmosisConfig OsmosisConfig) (*RealOsmosisAdapter, error) {
	mode := osmosisConfig.ValidationMode
	if mode == "" {
		mode = RouteValidationRuntime
	}
	if osmosisConfig.RouteProfile == nil {
		return nil, fmt.Errorf("osmosis route profile is required")
	}
	if err := osmosisConfig.RouteProfile.Validate(mode); err != nil {
		return nil, err
	}
	if mode == RouteValidationRuntime {
		if osmosisConfig.RouteAuthorizer == nil {
			return nil, fmt.Errorf("trusted DEX route authorizer is required")
		}
		if err := osmosisConfig.RouteAuthorizer.AuthorizeDEXRoute(*osmosisConfig.RouteProfile); err != nil {
			return nil, fmt.Errorf("authorize DEX route: %w", err)
		}
		if err := validateOsmosisGRPCEndpoint(osmosisConfig.GetGRPCEndpoint()); err != nil {
			return nil, err
		}
	}
	if osmosisConfig.RouteProfile.ChainID != osmosisConfig.GetChainID() {
		return nil, fmt.Errorf("%w: profile %s, adapter %s", ErrWrongChain, osmosisConfig.RouteProfile.ChainID, osmosisConfig.GetChainID())
	}
	if osmosisConfig.ChainEvidence == nil {
		return nil, fmt.Errorf("chain evidence provider is required")
	}
	if strings.TrimSpace(osmosisConfig.ChainEvidence.SourceID()) == "" {
		return nil, fmt.Errorf("chain evidence source identity is required")
	}
	if osmosisConfig.PoolState == nil {
		return nil, fmt.Errorf("bound Osmosis pool state provider is required")
	}
	if osmosisConfig.Oracle == nil {
		return nil, fmt.Errorf("oracle price provider is required")
	}
	if osmosisConfig.ExecutionVerifier == nil {
		return nil, fmt.Errorf("execution envelope verifier is required")
	}
	restEndpoint, err := validateOsmosisRESTEndpoint(osmosisConfig.GetRESTEndpoint(), mode)
	if err != nil {
		return nil, err
	}
	osmosisConfig.RESTEndpoint = restEndpoint.String()
	now := osmosisConfig.Now
	if now == nil {
		now = time.Now
	}
	httpClient := osmosisConfig.HTTPClient
	if httpClient == nil {
		httpClient = security.NewSecureHTTPClient(security.WithTimeout(osmosisConfig.Timeout))
		if mode == RouteValidationRuntime {
			if transport, ok := httpClient.Transport.(*http.Transport); ok {
				clonedTransport := transport.Clone()
				clonedTransport.Proxy = nil
				httpClient.Transport = clonedTransport
			}
		}
	} else {
		cloned := *httpClient
		httpClient = &cloned
		if mode == RouteValidationRuntime {
			transport, ok := httpClient.Transport.(*http.Transport)
			if !ok || transport.Proxy != nil || transport.TLSClientConfig != nil && transport.TLSClientConfig.InsecureSkipVerify {
				return nil, fmt.Errorf("unsafe production Osmosis HTTP transport")
			}
			httpClient.Transport = transport.Clone()
		}
	}
	httpClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
	adapter := &RealOsmosisAdapter{
		BaseAdapter:       NewBaseAdapter(cfg),
		config:            osmosisConfig,
		httpClient:        httpClient,
		pools:             make(map[string]LiquidityPool),
		profile:           *osmosisConfig.RouteProfile,
		mode:              mode,
		poolState:         osmosisConfig.PoolState,
		evidence:          osmosisConfig.ChainEvidence,
		oracle:            osmosisConfig.Oracle,
		executionVerifier: osmosisConfig.ExecutionVerifier,
		now:               now,
	}

	adapter.chainID = osmosisConfig.RouteProfile.ChainID

	return adapter, nil
}

// Connect establishes connection to Osmosis
func (a *RealOsmosisAdapter) Connect(ctx context.Context) error {
	a.lifecycleMu.Lock()
	defer a.lifecycleMu.Unlock()
	endpoint := a.config.GetGRPCEndpoint()

	// Create gRPC connection
	conn, err := grpc.NewClient(endpoint,
		grpc.WithTransportCredentials(credentials.NewTLS(security.SecureTLSConfig())),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to Osmosis gRPC: %w", err)
	}

	if a.grpcConn != nil {
		_ = a.grpcConn.Close()
	}
	a.grpcConn = conn
	a.connected = true
	a.healthy.Store(true)

	// Initial pool refresh
	if err := a.refreshPools(ctx); err != nil {
		// Log but don't fail - we can still use REST API
		_ = err
	}

	return nil
}

// Disconnect closes the connection
func (a *RealOsmosisAdapter) Disconnect() error {
	a.lifecycleMu.Lock()
	defer a.lifecycleMu.Unlock()
	a.connected = false
	a.healthy.Store(false)
	if a.grpcConn != nil {
		err := a.grpcConn.Close()
		a.grpcConn = nil
		return err
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
	a.lastRefresh = a.now()

	return nil
}

// queryPools acquires pools and their authenticated block observation through
// one mandatory boundary. Unheighted REST pool responses are not executable.
func (a *RealOsmosisAdapter) queryPools(ctx context.Context, limit int) ([]LiquidityPool, error) {
	bound, err := a.poolState.PoolStates(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("%w: bound pool discovery: %v", ErrOsmosisQueryFailed, err)
	}
	observation, err := a.validateBoundPoolResponse(bound)
	if err != nil {
		return nil, err
	}
	var poolsResp OsmosisPoolsResponse
	if err := decodeBoundedJSON(bytes.NewReader(bound.Payload), &poolsResp); err != nil {
		return nil, fmt.Errorf("failed to decode pools response: %w", err)
	}
	pools, err := a.convertPools(poolsResp.Pools, observation)
	if err != nil {
		return nil, err
	}
	return pools, nil
}

// convertPools converts Osmosis pool data to our LiquidityPool type
func (a *RealOsmosisAdapter) convertPools(osmPools []OsmosisPoolData, observation ChainObservation) ([]LiquidityPool, error) {
	pools := make([]LiquidityPool, 0, len(osmPools))

	for _, op := range osmPools {
		if !a.profile.poolAllowed(op.ID) {
			continue
		}
		pool, err := a.convertPool(op, observation)
		if err != nil {
			return nil, err
		}
		pools = append(pools, pool)
	}

	return pools, nil
}

// convertPool converts a single Osmosis pool to our LiquidityPool type
func (a *RealOsmosisAdapter) convertPool(op OsmosisPoolData, observation ChainObservation) (LiquidityPool, error) {
	if !a.profile.poolAllowed(op.ID) {
		return LiquidityPool{}, ErrPoolNotAllowed
	}
	if !a.supportedConstantProductPool(op) {
		return LiquidityPool{}, ErrOsmosisInvalidPool
	}
	pool := LiquidityPool{
		ID:        op.ID,
		DEX:       a.name,
		Type:      PoolTypeConstantProduct, // Osmosis pools are x*y=k AMMs
		Tokens:    make([]Token, 0, 2),
		Reserves:  make(map[string]sdkmath.Int),
		UpdatedAt: observation.ObservedAt.UTC(),
		ChainID:   observation.ChainID,
		SourceID:  observation.SourceID,
		Height:    observation.Height,
		BlockHash: observation.BlockHash,
	}

	// Parse swap fee
	fee, err := sdkmath.LegacyNewDecFromStr(op.PoolParams.SwapFee)
	if err != nil || fee.IsNegative() || !fee.LT(sdkmath.LegacyOneDec()) {
		return LiquidityPool{}, ErrOsmosisInvalidPool
	}
	pool.Fee = fee

	// Parse pool assets
	for _, asset := range op.PoolAssets {
		token := a.parseToken(asset.Token)
		pool.Tokens = append(pool.Tokens, token)

		// Parse reserves by exact denom. Symbols are display-only and are never
		// used for production reserve direction.
		amount, ok := sdkmath.NewIntFromString(asset.Token.Amount)
		if !ok || !amount.IsPositive() {
			return LiquidityPool{}, ErrOsmosisInvalidPool
		}
		pool.Reserves[token.Denom] = amount
	}

	return pool, nil
}

func (a *RealOsmosisAdapter) supportedConstantProductPool(pool OsmosisPoolData) bool {
	typeSupported := pool.Type == osmosisGAMMPoolType || a.mode == RouteValidationEngineering && pool.Type == ""
	if !typeSupported || len(pool.PoolAssets) != 2 {
		return false
	}
	firstWeight, err := sdkmath.LegacyNewDecFromStr(pool.PoolAssets[0].Weight)
	if err != nil || !firstWeight.IsPositive() {
		return false
	}
	secondWeight, err := sdkmath.LegacyNewDecFromStr(pool.PoolAssets[1].Weight)
	return err == nil && secondWeight.IsPositive() && firstWeight.Equal(secondWeight)
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

	decimals := uint8(0)
	if configured, ok := a.profile.token(coin.Denom); ok {
		decimals = configured.Decimals
	}
	return Token{
		Symbol:   symbol,
		Denom:    coin.Denom,
		Decimals: decimals,
		ChainID:  a.chainID,
		IsNative: !strings.HasPrefix(coin.Denom, "ibc/"),
	}
}

// GetSupportedPairs returns supported trading pairs
func (a *RealOsmosisAdapter) GetSupportedPairs(ctx context.Context) ([]TradingPair, error) {
	// Refresh pools if stale
	a.poolsMu.RLock()
	needsRefresh := a.now().Sub(a.lastRefresh) > a.config.PoolRefreshInterval
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
		Timestamp:  a.now().UTC(),
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
	needsRefresh := len(a.pools) == 0 || a.now().Sub(a.lastRefresh) > a.config.PoolRefreshInterval
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
			if !a.profile.poolAllowed(pool.ID) {
				return LiquidityPool{}, ErrPoolNotAllowed
			}
			return pool, nil
		}
	}

	return LiquidityPool{}, ErrOsmosisPoolNotFound
}

// querySpotPrice queries the spot price from Osmosis
func (a *RealOsmosisAdapter) querySpotPrice(ctx context.Context, poolID, baseDenom, quoteDenom string) (sdkmath.LegacyDec, error) {
	if !a.profile.poolAllowed(poolID) {
		return sdkmath.LegacyDec{}, ErrPoolNotAllowed
	}
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
	if err := requireJSONResponse(resp); err != nil {
		return sdkmath.LegacyDec{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return sdkmath.LegacyZeroDec(), fmt.Errorf("%w: status %d", ErrOsmosisQueryFailed, resp.StatusCode)
	}

	var priceResp OsmosisSpotPriceResponse
	if err := decodeBoundedJSON(resp.Body, &priceResp); err != nil {
		return sdkmath.LegacyZeroDec(), fmt.Errorf("failed to decode price response: %w", err)
	}

	price, err := sdkmath.LegacyNewDecFromStr(priceResp.SpotPrice)
	if err != nil {
		return sdkmath.LegacyZeroDec(), fmt.Errorf("failed to parse spot price: %w", err)
	}

	return price, nil
}

// GetPool fetches pool information and authenticated state evidence by ID.
func (a *RealOsmosisAdapter) GetPool(ctx context.Context, poolID string) (LiquidityPool, error) {
	if !a.profile.poolAllowed(poolID) {
		return LiquidityPool{}, ErrPoolNotAllowed
	}
	// Check cache first
	a.poolsMu.RLock()
	pool, ok := a.pools[poolID]
	a.poolsMu.RUnlock()

	if ok {
		return pool, nil
	}

	bound, err := a.poolState.PoolState(ctx, poolID)
	if err != nil {
		return LiquidityPool{}, fmt.Errorf("%w: bound pool %s: %v", ErrOsmosisQueryFailed, poolID, err)
	}
	observation, err := a.validateBoundPoolResponse(bound)
	if err != nil {
		return LiquidityPool{}, err
	}
	var poolResp OsmosisPoolResponse
	if err := decodeBoundedJSON(bytes.NewReader(bound.Payload), &poolResp); err != nil {
		return LiquidityPool{}, fmt.Errorf("failed to decode pool response: %w", err)
	}
	if poolResp.Pool.ID != poolID {
		return LiquidityPool{}, fmt.Errorf("%w: requested pool %s, received %s", ErrPoolStateEvidence, poolID, poolResp.Pool.ID)
	}
	return a.convertPool(poolResp.Pool, observation)
}

// ListPools lists pools matching the query
func (a *RealOsmosisAdapter) ListPools(ctx context.Context, query PoolQuery) ([]LiquidityPool, error) {
	// Ensure pools are loaded
	a.poolsMu.RLock()
	empty := len(a.pools) == 0
	a.poolsMu.RUnlock()
	if empty {
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
	if err := request.Validate(); err != nil {
		return SwapQuote{}, err
	}
	if request.Type != "" && request.Type != SwapTypeExactIn {
		return SwapQuote{}, fmt.Errorf("only exact-in Osmosis quotes are supported")
	}
	if request.FromToken.ChainID == "" || request.ToToken.ChainID == "" || request.FromToken.ChainID != a.chainID || request.ToToken.ChainID != a.chainID {
		return SwapQuote{}, ErrWrongChain
	}
	fromProfile, ok := a.profile.token(request.FromToken.Denom)
	if !ok {
		return SwapQuote{}, ErrUnsupportedPair
	}
	toProfile, ok := a.profile.token(request.ToToken.Denom)
	if !ok {
		return SwapQuote{}, ErrUnsupportedPair
	}
	if request.FromToken.Decimals != fromProfile.Decimals || request.ToToken.Decimals != toProfile.Decimals {
		return SwapQuote{}, ErrTokenDecimals
	}
	if request.FromToken.Symbol != fromProfile.Symbol || request.ToToken.Symbol != toProfile.Symbol {
		return SwapQuote{}, ErrUnsupportedPair
	}
	if request.Amount.GT(a.profile.MaxAmount) {
		return SwapQuote{}, ErrAmountTooLarge
	}
	if !request.Deadline.IsZero() && !request.Deadline.After(a.now()) {
		return SwapQuote{}, ErrQuoteExpired
	}
	// Find pool for the swap
	pool, err := a.findPoolForPair(ctx, request.FromToken.Symbol, request.ToToken.Symbol)
	if err != nil {
		return SwapQuote{}, err
	}
	if !a.profile.poolAllowed(pool.ID) {
		return SwapQuote{}, ErrPoolNotAllowed
	}
	if pool.ChainID != a.chainID || pool.SourceID == "" || pool.Height == 0 || pool.BlockHash == "" {
		return SwapQuote{}, ErrPoolStateEvidence
	}
	reserveIn := pool.Reserves[request.FromToken.Denom]
	reserveOut := pool.Reserves[request.ToToken.Denom]
	if reserveIn.IsNil() || reserveOut.IsNil() || reserveIn.LT(a.profile.MinReserve) || reserveOut.LT(a.profile.MinReserve) {
		return SwapQuote{}, ErrInsufficientLiquidity
	}
	if pool.TotalLiquidity.IsNil() || pool.TotalLiquidity.TruncateInt().LT(a.profile.MinLiquidity) {
		// When the upstream response does not provide a trusted common-value TVL,
		// require each exact reserve to satisfy the declared liquidity floor.
		if reserveIn.LT(a.profile.MinLiquidity) || reserveOut.LT(a.profile.MinLiquidity) {
			return SwapQuote{}, ErrInsufficientLiquidity
		}
	}
	outputAmount, err := ConstantProductExactIn(reserveIn, reserveOut, request.Amount, pool.Fee)
	if err != nil {
		return SwapQuote{}, err
	}
	priceImpact, err := exactPriceImpact(reserveIn, reserveOut, request.Amount, outputAmount)
	if err != nil {
		return SwapQuote{}, err
	}
	if priceImpact.GT(a.profile.MaxPriceImpact) {
		return SwapQuote{}, ErrPriceImpactExceeded
	}
	slippage := request.SlippageToleranceExact
	if slippage.IsNil() || slippage.IsNegative() || slippage.GT(a.profile.MaxPriceImpact) {
		return SwapQuote{}, ErrSlippageExceeded
	}
	minOutputInt := sdkmath.LegacyOneDec().Sub(slippage).MulInt(outputAmount).TruncateInt()
	if !minOutputInt.IsPositive() || minOutputInt.GT(outputAmount) {
		return SwapQuote{}, ErrMinimumOutput
	}
	if request.Amount.GTE(reserveIn) {
		return SwapQuote{}, ErrInsufficientLiquidity
	}
	observation, err := a.validatePoolObservation(pool)
	if err != nil {
		return SwapQuote{}, err
	}
	rate, err := exactExchangeRate(request.Amount, outputAmount, request.FromToken.Decimals, request.ToToken.Decimals)
	if err != nil {
		return SwapQuote{}, err
	}
	oraclePrice, err := a.oracle.Price(request.FromToken.Denom, request.ToToken.Denom, observation.Height)
	if err != nil {
		return SwapQuote{}, fmt.Errorf("oracle price: %w", err)
	}
	oracleDeviation, err := exactDeviation(rate, oraclePrice)
	if err != nil {
		return SwapQuote{}, err
	}
	if oracleDeviation.GT(a.profile.MaxOracleDeviation) {
		return SwapQuote{}, ErrOracleDeviation
	}
	evidence := PoolStateEvidence{
		ChainID: a.chainID, SourceID: pool.SourceID, ProfileID: a.profile.ID, PoolID: pool.ID, Height: pool.Height,
		BlockHash: pool.BlockHash, ObservedAt: pool.UpdatedAt, FromDenom: request.FromToken.Denom,
		ToDenom: request.ToToken.Denom, FromDecimals: request.FromToken.Decimals, ToDecimals: request.ToToken.Decimals,
		ReserveIn: reserveIn, ReserveOut: reserveOut, SwapFee: pool.Fee,
	}
	evidence.StateDigest, err = canonicalPoolStateDigest(evidence)
	if err != nil {
		return SwapQuote{}, err
	}

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
		PriceImpactExact: priceImpact,
	}
	now := a.now().UTC()
	expires := now.Add(a.profile.QuoteTTL)
	if !request.Deadline.IsZero() && request.Deadline.Before(expires) {
		expires = request.Deadline.UTC()
	}
	quote := SwapQuote{
		Request:              request,
		Route:                route,
		InputAmount:          request.Amount,
		OutputAmount:         outputAmount,
		MinOutputAmount:      minOutputInt,
		Rate:                 rate,
		TotalFee:             pool.Fee.Mul(sdkmath.LegacyNewDecFromInt(request.Amount)).TruncateInt(),
		PriceImpactExact:     priceImpact,
		ExpiresAt:            expires,
		CreatedAt:            now,
		ProfileID:            a.profile.ID,
		ChainID:              a.chainID,
		DEX:                  "osmosis",
		DEXVersion:           a.profile.Version,
		PoolStateEvidence:    []PoolStateEvidence{evidence},
		OraclePrice:          oraclePrice,
		OracleDeviation:      oracleDeviation,
		ObservationHeight:    pool.Height,
		ObservationBlockHash: pool.BlockHash,
	}
	quote.ID, err = QuoteDigest(quote)
	if err != nil {
		return SwapQuote{}, err
	}
	quote.QuoteDigest = quote.ID
	return quote, nil
}

// ExecuteSwap executes a swap on Osmosis after verifying the signed envelope is
// bound to the canonical quote and unsigned payload.
func (a *RealOsmosisAdapter) ExecuteSwap(ctx context.Context, quote SwapQuote, signedTx []byte) (SwapResult, error) {
	if err := a.validateQuote(quote); err != nil {
		return SwapResult{}, err
	}
	rawSignedTx, err := verifySignedExecutionEnvelope(quote, signedTx)
	if err != nil {
		return SwapResult{}, err
	}
	payload, err := BuildExecutionPayload(quote)
	if err != nil {
		return SwapResult{}, err
	}
	if err := a.executionVerifier.VerifySignedExecution(ctx, payload, rawSignedTx); err != nil {
		return SwapResult{}, fmt.Errorf("%w: signed transaction is detached from quote payload: %v", ErrExecutionPayload, err)
	}

	endpoint := a.config.GetRESTEndpoint()
	url := fmt.Sprintf("%s/cosmos/tx/v1beta1/txs", endpoint)

	var txRaw tx.TxRaw
	if err := txRaw.Unmarshal(rawSignedTx); err != nil || len(txRaw.BodyBytes) == 0 || len(txRaw.AuthInfoBytes) == 0 || len(txRaw.Signatures) == 0 {
		return SwapResult{}, fmt.Errorf("%w: signed transaction is not canonical TxRaw bytes", ErrExecutionPayload)
	}

	// Cosmos REST JSON requires protobuf bytes to be canonical base64 strings.
	broadcastReq := struct {
		TxBytes string `json:"tx_bytes"`
		Mode    string `json:"mode"`
	}{TxBytes: base64.StdEncoding.EncodeToString(rawSignedTx), Mode: "BROADCAST_MODE_SYNC"}

	reqBody, err := json.Marshal(broadcastReq)
	if err != nil {
		return SwapResult{}, fmt.Errorf("failed to marshal broadcast request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return SwapResult{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return SwapResult{}, fmt.Errorf("%w: %v", ErrOsmosisSwapFailed, err)
	}
	defer resp.Body.Close()
	if err := requireJSONResponse(resp); err != nil {
		return SwapResult{}, err
	}
	if resp.StatusCode != http.StatusOK {
		body, readErr := readBoundedResponse(resp.Body)
		if readErr != nil {
			return SwapResult{}, readErr
		}
		return SwapResult{}, fmt.Errorf("%w: status %d: %s", ErrOsmosisSwapFailed, resp.StatusCode, string(body))
	}

	// Parse broadcast response
	var broadcastResp struct {
		TxResponse struct {
			TxHash string `json:"txhash"`
			Code   int    `json:"code"`
			Log    string `json:"raw_log"`
		} `json:"tx_response"`
	}

	if err := decodeBoundedJSON(resp.Body, &broadcastResp); err != nil {
		return SwapResult{}, fmt.Errorf("failed to decode broadcast response: %w", err)
	}

	if broadcastResp.TxResponse.Code != 0 {
		return SwapResult{}, fmt.Errorf("%w: code %d: %s",
			ErrOsmosisSwapFailed, broadcastResp.TxResponse.Code, broadcastResp.TxResponse.Log)
	}
	if strings.TrimSpace(broadcastResp.TxResponse.TxHash) == "" {
		return SwapResult{}, fmt.Errorf("%w: missing transaction hash", ErrOsmosisSwapFailed)
	}

	confirmed, err := a.queryConfirmedSwap(ctx, broadcastResp.TxResponse.TxHash, quote)
	if err != nil {
		return SwapResult{}, err
	}
	return confirmed, nil
}

type osmosisTxResponse struct {
	TxResponse struct {
		TxHash  string `json:"txhash"`
		Code    int    `json:"code"`
		Log     string `json:"raw_log"`
		Height  string `json:"height"`
		GasUsed string `json:"gas_used"`
		Events  []struct {
			Type       string `json:"type"`
			Attributes []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			} `json:"attributes"`
		} `json:"events"`
	} `json:"tx_response"`
}

func (a *RealOsmosisAdapter) queryConfirmedSwap(ctx context.Context, txHash string, quote SwapQuote) (SwapResult, error) {
	endpoint := a.config.GetRESTEndpoint()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"/cosmos/tx/v1beta1/txs/"+url.PathEscape(txHash), nil)
	if err != nil {
		return SwapResult{}, err
	}
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return SwapResult{}, fmt.Errorf("%w: confirmation query: %v", ErrOsmosisSwapFailed, err)
	}
	defer resp.Body.Close()
	if err := requireJSONResponse(resp); err != nil {
		return SwapResult{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return SwapResult{}, fmt.Errorf("%w: confirmation status %d", ErrOsmosisSwapFailed, resp.StatusCode)
	}
	var confirmed osmosisTxResponse
	if err := decodeBoundedJSON(resp.Body, &confirmed); err != nil {
		return SwapResult{}, err
	}
	if confirmed.TxResponse.Code != 0 || !strings.EqualFold(confirmed.TxResponse.TxHash, txHash) {
		return SwapResult{}, fmt.Errorf("%w: transaction not confirmed successfully", ErrOsmosisSwapFailed)
	}
	confirmedHeight, err := strconv.ParseUint(confirmed.TxResponse.Height, 10, 64)
	if err != nil || confirmedHeight == 0 {
		return SwapResult{}, fmt.Errorf("%w: invalid confirmation height", ErrOsmosisSwapFailed)
	}
	latest, err := a.currentObservation()
	if err != nil {
		return SwapResult{}, err
	}
	if confirmedHeight > latest.Height || latest.Height-confirmedHeight < a.profile.FinalityBlocks {
		return SwapResult{}, fmt.Errorf("%w: transaction lacks required finality", ErrOsmosisSwapFailed)
	}
	confirmedBlockHash, err := a.evidence.BlockHash(confirmedHeight)
	if err != nil || strings.TrimSpace(confirmedBlockHash) == "" {
		return SwapResult{}, fmt.Errorf("%w: confirmation block identity unavailable", ErrOsmosisSwapFailed)
	}
	output, ok := confirmedSwapOutput(confirmed, quote.Request.ToToken.Denom)
	if !ok || output.LT(quote.MinOutputAmount) {
		return SwapResult{}, ErrMinimumOutput
	}
	gasUsed, err := strconv.ParseUint(confirmed.TxResponse.GasUsed, 10, 64)
	if err != nil {
		return SwapResult{}, fmt.Errorf("%w: invalid confirmed gas", ErrOsmosisSwapFailed)
	}
	return SwapResult{QuoteID: quote.ID, TxHash: txHash, InputAmount: quote.InputAmount, OutputAmount: output,
		Fee: quote.TotalFee, GasUsed: gasUsed, ExecutedAt: a.now().UTC(), Route: quote.Route}, nil
}

func confirmedSwapOutput(response osmosisTxResponse, denom string) (sdkmath.Int, bool) {
	for _, event := range response.TxResponse.Events {
		if event.Type != "token_swapped" {
			continue
		}
		for _, attribute := range event.Attributes {
			if attribute.Key != "tokens_out" || !strings.HasSuffix(attribute.Value, denom) {
				continue
			}
			amount := strings.TrimSuffix(attribute.Value, denom)
			parsed, ok := sdkmath.NewIntFromString(amount)
			return parsed, ok && parsed.IsPositive()
		}
	}
	return sdkmath.Int{}, false
}

func requireJSONResponse(response *http.Response) error {
	contentType := strings.ToLower(strings.TrimSpace(strings.Split(response.Header.Get("Content-Type"), ";")[0]))
	if contentType != "application/json" && !strings.HasSuffix(contentType, "+json") {
		return fmt.Errorf("%w: remote response is not JSON", ErrOsmosisQueryFailed)
	}
	return nil
}

func validateOsmosisRESTEndpoint(raw string, mode RouteValidationMode) (*url.URL, error) {
	parsed, err := url.Parse(raw)
	if err != nil || !parsed.IsAbs() || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Path != "" {
		return nil, fmt.Errorf("invalid Osmosis REST endpoint")
	}
	if mode == RouteValidationRuntime && parsed.Scheme != "https" {
		return nil, fmt.Errorf("production Osmosis REST endpoint must use HTTPS")
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return nil, fmt.Errorf("unsupported Osmosis REST endpoint scheme")
	}
	return parsed, nil
}

func validateOsmosisGRPCEndpoint(raw string) error {
	if strings.ContainsAny(raw, "/?#@\\") {
		return fmt.Errorf("invalid Osmosis gRPC endpoint")
	}
	host, port, err := net.SplitHostPort(raw)
	if err != nil || strings.TrimSpace(host) == "" || port != "443" {
		return fmt.Errorf("production Osmosis gRPC endpoint must be host:443")
	}
	return nil
}

func decodeBoundedJSON(body io.Reader, target interface{}) error {
	limited := &io.LimitedReader{R: body, N: maxOsmosisResponseBytes + 1}
	decoder := json.NewDecoder(limited)
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if limited.N == 0 {
		return ErrRemoteResponseTooBig
	}
	var trailing interface{}
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("remote response contains trailing JSON")
		}
		return err
	}
	return nil
}

func readBoundedResponse(body io.Reader) ([]byte, error) {
	limited := io.LimitReader(body, maxOsmosisResponseBytes+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > maxOsmosisResponseBytes {
		return nil, ErrRemoteResponseTooBig
	}
	return raw, nil
}

func (a *RealOsmosisAdapter) currentObservation() (ChainObservation, error) {
	observation, err := a.evidence.LatestObservation()
	if err != nil {
		return ChainObservation{}, fmt.Errorf("chain observation: %w", err)
	}
	if observation.ChainID != a.chainID {
		return ChainObservation{}, ErrWrongChain
	}
	if observation.Height == 0 || strings.TrimSpace(observation.SourceID) == "" || strings.TrimSpace(observation.BlockHash) == "" || observation.ObservedAt.IsZero() {
		return ChainObservation{}, ErrPoolStateEvidence
	}
	if age := a.now().Sub(observation.ObservedAt); age < 0 || age > a.profile.MaxObservationAge {
		return ChainObservation{}, ErrPoolStateStale
	}
	return observation, nil
}

// validateBoundPoolResponse proves the response metadata and independently
// resolved canonical block identity refer to the same state and source. It
// deliberately does not infer a binding from request order or response time.
func (a *RealOsmosisAdapter) validateBoundPoolResponse(bound BoundOsmosisPoolResponse) (ChainObservation, error) {
	observation := bound.Observation
	if observation.ChainID != a.chainID {
		return ChainObservation{}, ErrWrongChain
	}
	if len(bound.Payload) == 0 || observation.Height == 0 || bound.ResponseHeight == 0 || bound.ResponseHeight != observation.Height ||
		strings.TrimSpace(observation.SourceID) == "" || bound.ResponseSourceID != observation.SourceID ||
		strings.TrimSpace(bound.ResponseBlockHash) == "" || !strings.EqualFold(bound.ResponseBlockHash, observation.BlockHash) ||
		strings.TrimSpace(observation.BlockHash) == "" || observation.ObservedAt.IsZero() {
		return ChainObservation{}, ErrPoolStateEvidence
	}
	if observation.SourceID != a.evidence.SourceID() {
		return ChainObservation{}, ErrPoolStateEvidence
	}
	if age := a.now().Sub(observation.ObservedAt); age < 0 || age > a.profile.MaxObservationAge {
		return ChainObservation{}, ErrPoolStateStale
	}
	canonicalHash, err := a.evidence.BlockHash(observation.Height)
	if err != nil || strings.TrimSpace(canonicalHash) == "" || !strings.EqualFold(canonicalHash, observation.BlockHash) {
		return ChainObservation{}, ErrPoolStateEvidence
	}
	latest, err := a.currentObservation()
	if err != nil {
		return ChainObservation{}, err
	}
	if latest.SourceID != observation.SourceID {
		return ChainObservation{}, ErrPoolStateEvidence
	}
	if observation.Height > latest.Height || latest.Height-observation.Height > a.profile.MaxHeightLag {
		return ChainObservation{}, ErrPoolStateStale
	}
	if observation.Height == latest.Height && !strings.EqualFold(observation.BlockHash, latest.BlockHash) {
		return ChainObservation{}, ErrPoolStateEvidence
	}
	return observation, nil
}

func (a *RealOsmosisAdapter) validatePoolObservation(pool LiquidityPool) (ChainObservation, error) {
	latest, err := a.currentObservation()
	if err != nil {
		return ChainObservation{}, err
	}
	if pool.ChainID != latest.ChainID || pool.SourceID != latest.SourceID {
		return ChainObservation{}, ErrPoolStateEvidence
	}
	if pool.Height == 0 || pool.Height > latest.Height || latest.Height-pool.Height > a.profile.MaxHeightLag {
		return ChainObservation{}, ErrPoolStateStale
	}
	if age := a.now().Sub(pool.UpdatedAt); age < 0 || age > a.profile.MaxObservationAge {
		return ChainObservation{}, ErrPoolStateStale
	}
	canonicalHash, err := a.evidence.BlockHash(pool.Height)
	if err != nil {
		return ChainObservation{}, err
	}
	if strings.TrimSpace(canonicalHash) == "" || !strings.EqualFold(canonicalHash, pool.BlockHash) {
		return ChainObservation{}, ErrPoolStateEvidence
	}
	return latest, nil
}

func (a *RealOsmosisAdapter) validateQuote(quote SwapQuote) error {
	if quote.IsExpiredAt(a.now()) {
		return ErrQuoteExpired
	}
	if quote.ProfileID != a.profile.ID || quote.ChainID != a.chainID || quote.DEX != "osmosis" || quote.DEXVersion != a.profile.Version {
		return ErrRouteNotCertified
	}
	if len(quote.Route.Hops) == 0 {
		return ErrRouteHopsExceeded
	}
	if quote.CreatedAt.IsZero() || quote.ExpiresAt.IsZero() || quote.CreatedAt.After(a.now()) || quote.ExpiresAt.After(quote.CreatedAt.Add(a.profile.QuoteTTL)) ||
		!quote.ExpiresAt.After(quote.CreatedAt) || (!quote.Request.Deadline.IsZero() && quote.ExpiresAt.After(quote.Request.Deadline)) {
		return ErrQuoteExpired
	}
	if quote.Request.Type != SwapTypeExactIn || !quote.Request.Amount.Equal(quote.InputAmount) ||
		quote.Request.FromToken.Denom != quote.Route.Hops[0].FromToken.Denom ||
		quote.Request.ToToken.Denom != quote.Route.Hops[len(quote.Route.Hops)-1].ToToken.Denom {
		return ErrPoolStateEvidence
	}
	if quote.InputAmount.IsNil() || quote.OutputAmount.IsNil() || quote.MinOutputAmount.IsNil() || !quote.InputAmount.IsPositive() || quote.OutputAmount.LT(quote.MinOutputAmount) {
		return ErrMinimumOutput
	}
	if quote.InputAmount.GT(a.profile.MaxAmount) || quote.Rate.IsNil() || !quote.Rate.IsPositive() || quote.OraclePrice.IsNil() || !quote.OraclePrice.IsPositive() ||
		quote.PriceImpactExact.IsNil() || quote.PriceImpactExact.GT(a.profile.MaxPriceImpact) || quote.OracleDeviation.IsNil() || quote.OracleDeviation.GT(a.profile.MaxOracleDeviation) {
		return ErrPriceImpactExceeded
	}
	if err := validateBoundedRoute(quote.Route, a.profile.MaxHops, a.profile); err != nil {
		return err
	}
	digest, err := QuoteDigest(quote)
	if err != nil || digest != quote.ID || digest != quote.QuoteDigest {
		return ErrExecutionPayload
	}
	if len(quote.PoolStateEvidence) != len(quote.Route.Hops) {
		return ErrPoolStateEvidence
	}
	latest, err := a.currentObservation()
	if err != nil {
		return err
	}
	totalFee := sdkmath.ZeroInt()
	totalImpact := sdkmath.LegacyZeroDec()
	newestHeight := uint64(0)
	newestHash := ""
	for i, evidence := range quote.PoolStateEvidence {
		hop := quote.Route.Hops[i]
		if evidence.ChainID != a.chainID || evidence.SourceID != latest.SourceID || evidence.Height == 0 || evidence.Height > latest.Height || latest.Height-evidence.Height > a.profile.MaxHeightLag ||
			evidence.ProfileID != a.profile.ID || evidence.PoolID != hop.PoolID ||
			evidence.FromDenom != hop.FromToken.Denom || evidence.ToDenom != hop.ToToken.Denom ||
			evidence.FromDecimals != hop.FromToken.Decimals || evidence.ToDecimals != hop.ToToken.Decimals {
			return ErrPoolStateEvidence
		}
		expectedDigest, digestErr := canonicalPoolStateDigest(evidence)
		if digestErr != nil || expectedDigest != evidence.StateDigest {
			return ErrPoolStateEvidence
		}
		canonicalHash, hashErr := a.evidence.BlockHash(evidence.Height)
		if hashErr != nil || !strings.EqualFold(canonicalHash, evidence.BlockHash) {
			return ErrPoolStateEvidence
		}
		if age := a.now().Sub(evidence.ObservedAt); age < 0 || age > a.profile.MaxObservationAge {
			return ErrPoolStateStale
		}
		recomputedOutput, outputErr := ConstantProductExactIn(evidence.ReserveIn, evidence.ReserveOut, hop.AmountIn, evidence.SwapFee)
		if outputErr != nil || !recomputedOutput.Equal(hop.AmountOut) || !hop.Fee.Equal(evidence.SwapFee) {
			return ErrPoolStateEvidence
		}
		hopImpact, impactErr := exactPriceImpact(evidence.ReserveIn, evidence.ReserveOut, hop.AmountIn, hop.AmountOut)
		if impactErr != nil {
			return impactErr
		}
		totalImpact = totalImpact.Add(hopImpact)
		totalFee = totalFee.Add(hop.Fee.MulInt(hop.AmountIn).TruncateInt())
		if evidence.Height >= newestHeight {
			newestHeight = evidence.Height
			newestHash = evidence.BlockHash
		}
	}
	expectedMinimum := sdkmath.LegacyOneDec().Sub(quote.Request.SlippageToleranceExact).MulInt(quote.OutputAmount).TruncateInt()
	if !quote.OutputAmount.Equal(quote.Route.Hops[len(quote.Route.Hops)-1].AmountOut) || !quote.MinOutputAmount.Equal(expectedMinimum) ||
		!quote.TotalFee.Equal(totalFee) || !quote.PriceImpactExact.Equal(totalImpact) || quote.ObservationHeight != newestHeight || !strings.EqualFold(quote.ObservationBlockHash, newestHash) {
		return ErrPoolStateEvidence
	}
	expectedRate, err := exactExchangeRate(quote.InputAmount, quote.OutputAmount, quote.Request.FromToken.Decimals, quote.Request.ToToken.Decimals)
	if err != nil {
		return err
	}
	if !quote.Rate.Equal(expectedRate) {
		return ErrPoolStateEvidence
	}
	oraclePrice, err := a.oracle.Price(quote.Request.FromToken.Denom, quote.Request.ToToken.Denom, quote.ObservationHeight)
	if err != nil {
		return err
	}
	if !quote.OraclePrice.Equal(oraclePrice) {
		return ErrOracleDeviation
	}
	deviation, err := exactDeviation(quote.Rate, oraclePrice)
	if err != nil || !quote.OracleDeviation.Equal(deviation) || deviation.GT(a.profile.MaxOracleDeviation) {
		return ErrOracleDeviation
	}
	return nil
}

func validateBoundedRoute(route SwapRoute, maxHops uint32, profile DEXRouteProfile) error {
	if len(route.Hops) == 0 || len(route.Hops) > int(maxHops) {
		return ErrRouteHopsExceeded
	}
	visited := make(map[string]struct{}, len(route.Hops)+1)
	for i, hop := range route.Hops {
		if !profile.poolAllowed(hop.PoolID) {
			return ErrPoolNotAllowed
		}
		if hop.FromToken.Denom == "" || hop.ToToken.Denom == "" || hop.FromToken.Denom == hop.ToToken.Denom {
			return ErrRouteCycle
		}
		if _, seen := visited[hop.FromToken.Denom]; seen {
			return ErrRouteCycle
		}
		visited[hop.FromToken.Denom] = struct{}{}
		if i > 0 {
			previous := route.Hops[i-1]
			if previous.ToToken.Denom != hop.FromToken.Denom || !previous.AmountOut.Equal(hop.AmountIn) {
				return ErrPoolStateEvidence
			}
		}
	}
	if _, seen := visited[route.Hops[len(route.Hops)-1].ToToken.Denom]; seen {
		return ErrRouteCycle
	}
	return nil
}

// EstimateGas estimates gas for a swap
func (a *RealOsmosisAdapter) EstimateGas(ctx context.Context, request SwapRequest) (uint64, error) {
	return 0, fmt.Errorf("real Osmosis gas simulation requires the finalized unsigned transaction payload")
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

	reserve0 := pool.Reserves[pool.Tokens[0].Denom]
	reserve1 := pool.Reserves[pool.Tokens[1].Denom]

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

		reserve := pool.Reserves[token.Denom]
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
	if err := requireJSONResponse(resp); err != nil {
		return sdkmath.LegacyDec{}, err
	}

	var liquidityResp struct {
		Liquidity []OsmosisCoin `json:"liquidity"`
	}
	if resp.StatusCode != http.StatusOK {
		return sdkmath.LegacyZeroDec(), fmt.Errorf("%w: status %d", ErrOsmosisQueryFailed, resp.StatusCode)
	}
	if err := decodeBoundedJSON(resp.Body, &liquidityResp); err != nil {
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

func (a *RealOsmosisAdapter) productionDEXProfile() (string, bool) {
	return a.profile.ID, a.mode == RouteValidationRuntime && a.profile.State == RouteCertifiedEnabled
}
