package pricefeed

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	sdkmath "cosmossdk.io/math"
)

// ============================================================================
// Chainlink Provider Implementation
// ============================================================================

// ChainlinkProvider implements the Provider interface for Chainlink price feeds
type ChainlinkProvider struct {
	name      string
	config    ChainlinkConfig
	health    SourceHealth
	healthMu  sync.RWMutex
	circuit   *CircuitBreaker
	retryer   *Retryer
	closed    atomic.Bool
	ethClient ChainlinkEthClient
}

// ChainlinkEthClient is an interface for Ethereum RPC calls
// In production, this would use go-ethereum's ethclient
type ChainlinkEthClient interface {
	// CallContract calls a smart contract method
	CallContract(ctx context.Context, address string, data []byte) ([]byte, error)
	// Close closes the client
	Close() error
}

// defaultEthClient is a stub implementation that can be replaced in production
type defaultEthClient struct {
	rpcURL string
}

func newDefaultEthClient(rpcURL string) *defaultEthClient {
	return &defaultEthClient{rpcURL: rpcURL}
}

func (c *defaultEthClient) CallContract(ctx context.Context, address string, data []byte) ([]byte, error) {
	// In production, this would use go-ethereum's ethclient to make RPC calls
	// For now, return an error indicating the client needs to be configured
	return nil, fmt.Errorf("chainlink ETH client not configured - set RPCURL in ChainlinkConfig")
}

func (c *defaultEthClient) Close() error {
	return nil
}

// NewChainlinkProvider creates a new Chainlink provider
func NewChainlinkProvider(name string, cfg ChainlinkConfig, retryCfg RetryConfig) (*ChainlinkProvider, error) {
	if cfg.RPCURL == "" {
		return nil, fmt.Errorf("chainlink provider requires RPC URL")
	}

	provider := &ChainlinkProvider{
		name:      name,
		config:    cfg,
		ethClient: newDefaultEthClient(cfg.RPCURL),
		health: SourceHealth{
			Source:    name,
			Type:      SourceTypeChainlink,
			Healthy:   true,
			LastCheck: time.Now(),
		},
		circuit: NewCircuitBreaker(5, 2, 30*time.Second),
		retryer: NewRetryer(retryCfg),
	}

	return provider, nil
}

// SetEthClient sets a custom Ethereum client (for production use)
func (p *ChainlinkProvider) SetEthClient(client ChainlinkEthClient) {
	p.ethClient = client
}

// Name returns the provider name
func (p *ChainlinkProvider) Name() string {
	return p.name
}

// Type returns the provider type
func (p *ChainlinkProvider) Type() SourceType {
	return SourceTypeChainlink
}

// GetPrice fetches the current price for an asset pair
func (p *ChainlinkProvider) GetPrice(ctx context.Context, baseAsset, quoteAsset string) (PriceData, error) {
	if p.closed.Load() {
		return PriceData{}, ErrSourceUnhealthy
	}

	// Check circuit breaker
	if !p.circuit.Allow() {
		return PriceData{}, ErrSourceUnhealthy
	}

	startTime := time.Now()

	price, err := DoWithResult(ctx, p.retryer, func(ctx context.Context) (PriceData, error) {
		return p.fetchPrice(ctx, baseAsset, quoteAsset)
	})

	latency := time.Since(startTime)

	p.healthMu.Lock()
	p.health.LastCheck = time.Now()
	p.health.RequestCount++
	p.health.Latency = latency
	if err != nil {
		p.health.ErrorCount++
		p.health.LastError = err.Error()
		p.circuit.RecordFailure()
	} else {
		p.health.Healthy = true
		p.health.LastSuccess = time.Now()
		p.health.LastError = ""
		p.circuit.RecordSuccess()
	}
	p.healthMu.Unlock()

	return price, err
}

// fetchPrice fetches the price from Chainlink oracle
func (p *ChainlinkProvider) fetchPrice(ctx context.Context, baseAsset, quoteAsset string) (PriceData, error) {
	// Look up the feed address
	pairKey := fmt.Sprintf("%s/%s", strings.ToUpper(baseAsset), strings.ToUpper(quoteAsset))
	feedAddress, ok := p.config.FeedAddresses[pairKey]
	if !ok {
		return PriceData{}, fmt.Errorf("%w: no Chainlink feed for %s", ErrPriceNotFound, pairKey)
	}

	// Call the Chainlink aggregator contract
	// Function selector for latestRoundData(): 0xfeaf968c
	latestRoundDataSelector := []byte{0xfe, 0xaf, 0x96, 0x8c}

	result, err := p.ethClient.CallContract(ctx, feedAddress, latestRoundDataSelector)
	if err != nil {
		return PriceData{}, fmt.Errorf("failed to call Chainlink feed: %w", err)
	}

	// Parse the result
	// latestRoundData returns: (roundId, answer, startedAt, updatedAt, answeredInRound)
	// answer is at bytes 32-64 (second 32-byte word)
	if len(result) < 64 {
		return PriceData{}, fmt.Errorf("invalid response from Chainlink feed")
	}

	// Extract the answer (price with 8 decimals for most Chainlink feeds)
	answerBytes := result[32:64]
	answer := new(big.Int).SetBytes(answerBytes)

	// Chainlink price feeds typically use 8 decimals
	decimals := int64(8)

	// Convert to sdkmath.LegacyDec
	price := sdkmath.LegacyNewDecFromBigIntWithPrec(answer, decimals)

	// Extract updatedAt timestamp (bytes 96-128)
	var updatedAt time.Time
	if len(result) >= 128 {
		updatedAtBytes := result[96:128]
		updatedAtInt := new(big.Int).SetBytes(updatedAtBytes).Int64()
		updatedAt = time.Unix(updatedAtInt, 0).UTC()
	}

	return PriceData{
		BaseAsset:     baseAsset,
		QuoteAsset:    quoteAsset,
		Price:         price,
		Timestamp:     time.Now().UTC(),
		Source:        p.name,
		Confidence:    0.99, // Chainlink is highly trusted
		LastUpdatedAt: updatedAt,
	}, nil
}

// GetPrices fetches prices for multiple asset pairs
func (p *ChainlinkProvider) GetPrices(ctx context.Context, pairs []AssetPair) (map[string]PriceData, error) {
	if p.closed.Load() {
		return nil, ErrSourceUnhealthy
	}

	if !p.circuit.Allow() {
		return nil, ErrSourceUnhealthy
	}

	results := make(map[string]PriceData)

	for _, pair := range pairs {
		price, err := p.GetPrice(ctx, pair.Base, pair.Quote)
		if err != nil {
			continue
		}
		results[pair.String()] = price
	}

	return results, nil
}

// IsHealthy checks if the provider is responding
func (p *ChainlinkProvider) IsHealthy(ctx context.Context) bool {
	if p.closed.Load() {
		return false
	}

	p.healthMu.RLock()
	health := p.health.Healthy
	lastSuccess := p.health.LastSuccess
	p.healthMu.RUnlock()

	// Consider unhealthy if no successful request in 5 minutes
	if time.Since(lastSuccess) > 5*time.Minute && !lastSuccess.IsZero() {
		return false
	}

	return health && p.circuit.State() != CircuitOpen
}

// Health returns detailed health information
func (p *ChainlinkProvider) Health() SourceHealth {
	p.healthMu.RLock()
	defer p.healthMu.RUnlock()
	health := p.health
	health.Healthy = p.circuit.State() != CircuitOpen && p.health.Healthy
	return health
}

// Close closes the provider
func (p *ChainlinkProvider) Close() error {
	p.closed.Store(true)
	if p.ethClient != nil {
		return p.ethClient.Close()
	}
	return nil
}

// ============================================================================
// Common Chainlink Feed Addresses (Ethereum Mainnet)
// ============================================================================

// ChainlinkMainnetFeeds contains common Chainlink price feed addresses on Ethereum mainnet
var ChainlinkMainnetFeeds = map[string]string{
	"ETH/USD":  "0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419",
	"BTC/USD":  "0xF4030086522a5bEEa4988F8cA5B36dbC97BeE88c",
	"LINK/USD": "0x2c1d072e956AFFC0D435Cb7AC38EF18d24d9127c",
	"USDC/USD": "0x8fFfFfd4AfB6115b954Bd326cbe7B4BA576818f6",
	"USDT/USD": "0x3E7d1eAB13ad0104d2750B8863b489D65364e32D",
	"DAI/USD":  "0xAed0c38402a5d19df6E4c03F4E2DceD6e29c1ee9",
	"ATOM/USD": "0xDC4BDB458C6361093069Ca2aD30D74cc152EdC75",
}

// ChainlinkSepoliaFeeds contains common Chainlink price feed addresses on Sepolia testnet
var ChainlinkSepoliaFeeds = map[string]string{
	"ETH/USD":  "0x694AA1769357215DE4FAC081bf1f309aDC325306",
	"BTC/USD":  "0x1b44F3514812d835EB1BDB0acB33d3fA3351Ee43",
	"LINK/USD": "0xc59E3633BAAC79493d908e63626716e204A45EdF",
}
