package pricefeed

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	sdkmath "cosmossdk.io/math"
)

// ============================================================================
// Pyth Provider Implementation
// ============================================================================

// PythProvider implements the Provider interface for Pyth Network price feeds
type PythProvider struct {
	name     string
	config   PythConfig
	client   *http.Client
	health   SourceHealth
	healthMu sync.RWMutex
	circuit  *CircuitBreaker
	retryer  *Retryer
	closed   atomic.Bool
}

// NewPythProvider creates a new Pyth provider
func NewPythProvider(name string, cfg PythConfig, retryCfg RetryConfig) (*PythProvider, error) {
	if cfg.HermesURL == "" {
		cfg.HermesURL = "https://hermes.pyth.network"
	}

	provider := &PythProvider{
		name:   name,
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		health: SourceHealth{
			Source:    name,
			Type:      SourceTypePyth,
			Healthy:   true,
			LastCheck: time.Now(),
		},
		circuit: NewCircuitBreaker(5, 2, 30*time.Second),
		retryer: NewRetryer(retryCfg),
	}

	return provider, nil
}

// Name returns the provider name
func (p *PythProvider) Name() string {
	return p.name
}

// Type returns the provider type
func (p *PythProvider) Type() SourceType {
	return SourceTypePyth
}

// GetPrice fetches the current price for an asset pair
func (p *PythProvider) GetPrice(ctx context.Context, baseAsset, quoteAsset string) (PriceData, error) {
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

// pythPriceResponse represents the Pyth API response
type pythPriceResponse struct {
	ID    string `json:"id"`
	Price struct {
		Price       string `json:"price"`
		Conf        string `json:"conf"`
		Expo        int    `json:"expo"`
		PublishTime int64  `json:"publish_time"`
	} `json:"price"`
	EMAPrice struct {
		Price       string `json:"price"`
		Conf        string `json:"conf"`
		Expo        int    `json:"expo"`
		PublishTime int64  `json:"publish_time"`
	} `json:"ema_price"`
}

// fetchPrice fetches the price from Pyth Network
func (p *PythProvider) fetchPrice(ctx context.Context, baseAsset, quoteAsset string) (PriceData, error) {
	// Look up the price feed ID
	pairKey := fmt.Sprintf("%s/%s", strings.ToUpper(baseAsset), strings.ToUpper(quoteAsset))
	priceID, ok := p.config.PriceIDs[pairKey]
	if !ok {
		// Try alternative mapping
		priceID = p.mapToPythPriceID(baseAsset, quoteAsset)
		if priceID == "" {
			return PriceData{}, fmt.Errorf("%w: no Pyth price feed for %s", ErrPriceNotFound, pairKey)
		}
	}

	// Remove 0x prefix if present
	priceID = strings.TrimPrefix(priceID, "0x")

	// Build URL for Hermes API
	endpoint := fmt.Sprintf("%s/api/latest_price_feeds?ids[]=%s", p.config.HermesURL, priceID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return PriceData{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return PriceData{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return PriceData{}, fmt.Errorf("Pyth API error %d: %s", resp.StatusCode, string(body))
	}

	var responses []pythPriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&responses); err != nil {
		return PriceData{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(responses) == 0 {
		return PriceData{}, ErrPriceNotFound
	}

	priceResp := responses[0]

	// Parse price
	priceInt, ok := sdkmath.NewIntFromString(priceResp.Price.Price)
	if !ok {
		return PriceData{}, fmt.Errorf("failed to parse price: invalid format")
	}

	// Apply exponent (Pyth uses negative exponents)
	price := sdkmath.LegacyNewDecFromBigIntWithPrec(priceInt.BigInt(), int64(-priceResp.Price.Expo))

	// Parse confidence interval
	confInt, _ := sdkmath.NewIntFromString(priceResp.Price.Conf)

	// Calculate confidence as percentage (lower conf = higher reliability)
	var confidence float64 = 0.95
	if !confInt.IsZero() && !priceInt.IsZero() {
		confPct := float64(confInt.Int64()) / float64(priceInt.Int64())
		confidence = 1 - confPct // Lower conf% = higher confidence
		if confidence < 0.5 {
			confidence = 0.5
		}
	}

	publishTime := time.Unix(priceResp.Price.PublishTime, 0).UTC()

	return PriceData{
		BaseAsset:     baseAsset,
		QuoteAsset:    quoteAsset,
		Price:         price,
		Timestamp:     time.Now().UTC(),
		Source:        p.name,
		Confidence:    confidence,
		LastUpdatedAt: publishTime,
	}, nil
}

// mapToPythPriceID maps common pairs to Pyth price feed IDs
func (p *PythProvider) mapToPythPriceID(baseAsset, quoteAsset string) string {
	base := strings.ToUpper(baseAsset)
	quote := strings.ToUpper(quoteAsset)

	// Common Pyth price feed IDs (mainnet)
	feeds := map[string]string{
		"BTC/USD":  "e62df6c8b4a85fe1a67db44dc12de5db330f7ac66b72dc658afedf0f4a415b43",
		"ETH/USD":  "ff61491a931112ddf1bd8147cd1b641375f79f5825126d665480874634fd0ace",
		"ATOM/USD": "b00b60f88b03a6a625a8d1c048c3f66653edf217439983d037e7222c4e612819",
		"USDC/USD": "eaa020c61cc479712813461ce153894a96a6c00b21ed0cfc2798d1f9a9e9c94a",
		"SOL/USD":  "ef0d8b6fda2ceba41da15d4095d1da392a0d2f8ed0c6c7bc0f4cfac8c280b56d",
		"LINK/USD": "8ac0c70fff57e9aefdf5edf44b51d62c2d433653cbb2cf5cc06bb115af04d221",
	}

	key := fmt.Sprintf("%s/%s", base, quote)
	if id, ok := feeds[key]; ok {
		return id
	}

	return ""
}

// GetPrices fetches prices for multiple asset pairs
func (p *PythProvider) GetPrices(ctx context.Context, pairs []AssetPair) (map[string]PriceData, error) {
	if p.closed.Load() {
		return nil, ErrSourceUnhealthy
	}

	if !p.circuit.Allow() {
		return nil, ErrSourceUnhealthy
	}

	// Collect all price IDs
	priceIDs := make([]string, 0, len(pairs))
	pairMap := make(map[string]AssetPair) // priceID -> pair

	for _, pair := range pairs {
		pairKey := fmt.Sprintf("%s/%s", strings.ToUpper(pair.Base), strings.ToUpper(pair.Quote))
		priceID, ok := p.config.PriceIDs[pairKey]
		if !ok {
			priceID = p.mapToPythPriceID(pair.Base, pair.Quote)
		}
		if priceID != "" {
			priceID = strings.TrimPrefix(priceID, "0x")
			priceIDs = append(priceIDs, priceID)
			pairMap[priceID] = pair
		}
	}

	if len(priceIDs) == 0 {
		return nil, ErrPriceNotFound
	}

	// Build URL with multiple IDs
	endpoint := p.config.HermesURL + "/api/latest_price_feeds?"
	for i, id := range priceIDs {
		if i > 0 {
			endpoint += "&"
		}
		endpoint += "ids[]=" + id
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Pyth API error %d: %s", resp.StatusCode, string(body))
	}

	var responses []pythPriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&responses); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	results := make(map[string]PriceData)

	for _, priceResp := range responses {
		pair, ok := pairMap[priceResp.ID]
		if !ok {
			continue
		}

		priceInt, ok := sdkmath.NewIntFromString(priceResp.Price.Price)
		if !ok {
			continue
		}

		price := sdkmath.LegacyNewDecFromBigIntWithPrec(priceInt.BigInt(), int64(-priceResp.Price.Expo))
		publishTime := time.Unix(priceResp.Price.PublishTime, 0).UTC()

		results[pair.String()] = PriceData{
			BaseAsset:     pair.Base,
			QuoteAsset:    pair.Quote,
			Price:         price,
			Timestamp:     time.Now().UTC(),
			Source:        p.name,
			Confidence:    0.95,
			LastUpdatedAt: publishTime,
		}
	}

	p.healthMu.Lock()
	p.health.LastCheck = time.Now()
	p.health.RequestCount++
	if len(results) > 0 {
		p.health.Healthy = true
		p.health.LastSuccess = time.Now()
		p.circuit.RecordSuccess()
	}
	p.healthMu.Unlock()

	return results, nil
}

// IsHealthy checks if the provider is responding
func (p *PythProvider) IsHealthy(ctx context.Context) bool {
	if p.closed.Load() {
		return false
	}

	p.healthMu.RLock()
	health := p.health.Healthy
	lastSuccess := p.health.LastSuccess
	p.healthMu.RUnlock()

	if time.Since(lastSuccess) > 5*time.Minute && !lastSuccess.IsZero() {
		return false
	}

	return health && p.circuit.State() != CircuitOpen
}

// Health returns detailed health information
func (p *PythProvider) Health() SourceHealth {
	p.healthMu.RLock()
	defer p.healthMu.RUnlock()
	health := p.health
	health.Healthy = p.circuit.State() != CircuitOpen && p.health.Healthy
	return health
}

// Close closes the provider
func (p *PythProvider) Close() error {
	p.closed.Store(true)
	return nil
}

// ============================================================================
// Common Pyth Price Feed IDs
// ============================================================================

// PythMainnetPriceIDs contains common Pyth price feed IDs
var PythMainnetPriceIDs = map[string]string{
	"BTC/USD":   "0xe62df6c8b4a85fe1a67db44dc12de5db330f7ac66b72dc658afedf0f4a415b43",
	"ETH/USD":   "0xff61491a931112ddf1bd8147cd1b641375f79f5825126d665480874634fd0ace",
	"ATOM/USD":  "0xb00b60f88b03a6a625a8d1c048c3f66653edf217439983d037e7222c4e612819",
	"USDC/USD":  "0xeaa020c61cc479712813461ce153894a96a6c00b21ed0cfc2798d1f9a9e9c94a",
	"USDT/USD":  "0x2b89b9dc8fdf9f34709a5b106b472f0f39bb6ca9ce04b0fd7f2e971688e2e53b",
	"SOL/USD":   "0xef0d8b6fda2ceba41da15d4095d1da392a0d2f8ed0c6c7bc0f4cfac8c280b56d",
	"LINK/USD":  "0x8ac0c70fff57e9aefdf5edf44b51d62c2d433653cbb2cf5cc06bb115af04d221",
	"AVAX/USD":  "0x93da3352f9f1d105fdfe4971cfa80e9dd777bfc5d0f683ebb6e1294b92137bb7",
	"DOGE/USD":  "0xdcef50dd0a4cd2dcc17e45df1676dcb336a11a61c69df7a0299b0150c672d25c",
	"MATIC/USD": "0x5de33440f6c50b5d2b4e8f65c4d2e03d7f2a8dd8d3b9d0e2d7a8b7c8d9e0f1a2",
}

