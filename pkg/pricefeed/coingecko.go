package pricefeed

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/virtengine/virtengine/pkg/security"
)

// ============================================================================
// CoinGecko Provider Implementation
// ============================================================================

// CoinGeckoProvider implements the Provider interface for CoinGecko API
type CoinGeckoProvider struct {
	name        string
	config      CoinGeckoConfig
	client      *http.Client
	rateLimiter *rateLimiter
	health      SourceHealth
	healthMu    sync.RWMutex
	circuit     *CircuitBreaker
	retryer     *Retryer
	closed      atomic.Bool
}

// rateLimiter implements simple rate limiting
type rateLimiter struct {
	mu        sync.Mutex
	tokens    int
	maxTokens int
	lastFill  time.Time
	fillRate  time.Duration
}

func newRateLimiter(maxPerMinute int) *rateLimiter {
	return &rateLimiter{
		tokens:    maxPerMinute,
		maxTokens: maxPerMinute,
		lastFill:  time.Now(),
		fillRate:  time.Minute / time.Duration(maxPerMinute),
	}
}

func (r *rateLimiter) acquire() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Refill tokens based on time passed
	now := time.Now()
	elapsed := now.Sub(r.lastFill)
	tokensToAdd := int(elapsed / r.fillRate)
	if tokensToAdd > 0 {
		r.tokens += tokensToAdd
		if r.tokens > r.maxTokens {
			r.tokens = r.maxTokens
		}
		r.lastFill = now
	}

	if r.tokens > 0 {
		r.tokens--
		return true
	}
	return false
}

// NewCoinGeckoProvider creates a new CoinGecko provider
func NewCoinGeckoProvider(name string, cfg CoinGeckoConfig, retryCfg RetryConfig) (*CoinGeckoProvider, error) {
	if cfg.APIURL == "" {
		cfg.APIURL = "https://api.coingecko.com/api/v3"
	}
	if cfg.RateLimitPerMinute <= 0 {
		cfg.RateLimitPerMinute = 10 // Default free tier limit
	}

	provider := &CoinGeckoProvider{
		name:        name,
		config:      cfg,
		client:      security.NewSecureHTTPClient(security.WithTimeout(30 * time.Second)),
		rateLimiter: newRateLimiter(cfg.RateLimitPerMinute),
		health: SourceHealth{
			Source:    name,
			Type:      SourceTypeCoinGecko,
			Healthy:   true,
			LastCheck: time.Now(),
		},
		circuit: NewCircuitBreaker(5, 2, 30*time.Second),
		retryer: NewRetryer(retryCfg),
	}

	return provider, nil
}

// Name returns the provider name
func (p *CoinGeckoProvider) Name() string {
	return p.name
}

// Type returns the provider type
func (p *CoinGeckoProvider) Type() SourceType {
	return SourceTypeCoinGecko
}

// GetPrice fetches the current price for an asset pair
func (p *CoinGeckoProvider) GetPrice(ctx context.Context, baseAsset, quoteAsset string) (PriceData, error) {
	if p.closed.Load() {
		return PriceData{}, ErrSourceUnhealthy
	}

	// Check circuit breaker
	if !p.circuit.Allow() {
		return PriceData{}, ErrSourceUnhealthy
	}

	// Check rate limit
	if !p.rateLimiter.acquire() {
		return PriceData{}, ErrRateLimitExceeded
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

// fetchPrice makes the actual HTTP request to CoinGecko
func (p *CoinGeckoProvider) fetchPrice(ctx context.Context, baseAsset, quoteAsset string) (PriceData, error) {
	// Map internal asset names to CoinGecko IDs
	coinID := p.mapToCoinGeckoID(baseAsset)
	currency := strings.ToLower(quoteAsset)

	// Build URL
	endpoint := fmt.Sprintf("%s/simple/price", p.config.APIURL)
	params := url.Values{}
	params.Set("ids", coinID)
	params.Set("vs_currencies", currency)
	params.Set("include_24hr_vol", "true")
	params.Set("include_last_updated_at", "true")
	params.Set("include_market_cap", "true")

	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return PriceData{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Accept", "application/json")
	if p.config.APIKey != "" {
		if p.config.UsePro {
			req.Header.Set("x-cg-pro-api-key", p.config.APIKey)
		} else {
			req.Header.Set("x-cg-demo-api-key", p.config.APIKey)
		}
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return PriceData{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return PriceData{}, ErrRateLimitExceeded
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return PriceData{}, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result map[string]map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return PriceData{}, fmt.Errorf("failed to decode response: %w", err)
	}

	coinData, ok := result[coinID]
	if !ok {
		return PriceData{}, ErrPriceNotFound
	}

	priceVal, ok := coinData[currency]
	if !ok {
		return PriceData{}, ErrPriceNotFound
	}

	price, err := p.parseFloat(priceVal)
	if err != nil {
		return PriceData{}, fmt.Errorf("failed to parse price: %w", err)
	}

	// Parse optional fields
	var volume24h, marketCap sdkmath.LegacyDec
	var lastUpdated time.Time

	if v, ok := coinData[currency+"_24h_vol"]; ok {
		if vol, err := p.parseFloat(v); err == nil {
			volume24h = sdkmath.LegacyNewDecFromBigInt(vol.BigInt())
		}
	}

	if v, ok := coinData[currency+"_market_cap"]; ok {
		if mc, err := p.parseFloat(v); err == nil {
			marketCap = sdkmath.LegacyNewDecFromBigInt(mc.BigInt())
		}
	}

	if v, ok := coinData["last_updated_at"]; ok {
		if ts, err := p.parseInt(v); err == nil {
			lastUpdated = time.Unix(ts, 0).UTC()
		}
	}

	return PriceData{
		BaseAsset:     baseAsset,
		QuoteAsset:    quoteAsset,
		Price:         price,
		Timestamp:     time.Now().UTC(),
		Source:        p.name,
		Confidence:    0.95, // CoinGecko aggregates from multiple exchanges
		Volume24h:     volume24h,
		MarketCap:     marketCap,
		LastUpdatedAt: lastUpdated,
	}, nil
}

// parseFloat parses a float value from JSON
func (p *CoinGeckoProvider) parseFloat(v interface{}) (sdkmath.LegacyDec, error) {
	switch val := v.(type) {
	case float64:
		return sdkmath.LegacyMustNewDecFromStr(fmt.Sprintf("%.18f", val)), nil
	case int64:
		return sdkmath.LegacyNewDec(val), nil
	case int:
		return sdkmath.LegacyNewDec(int64(val)), nil
	default:
		return sdkmath.LegacyDec{}, fmt.Errorf("unexpected type %T", v)
	}
}

// parseInt parses an int value from JSON
func (p *CoinGeckoProvider) parseInt(v interface{}) (int64, error) {
	switch val := v.(type) {
	case float64:
		return int64(val), nil
	case int64:
		return val, nil
	case int:
		return int64(val), nil
	default:
		return 0, fmt.Errorf("unexpected type %T", v)
	}
}

// mapToCoinGeckoID maps internal asset IDs to CoinGecko IDs
func (p *CoinGeckoProvider) mapToCoinGeckoID(asset string) string {
	// Common mappings
	mappings := map[string]string{
		"uve":        "virtengine",
		"virtengine": "virtengine",
		"atom":       "cosmos",
		"cosmos":     "cosmos",
		"usdc":       "usd-coin",
		"usdt":       "tether",
		"eth":        "ethereum",
		"btc":        "bitcoin",
	}

	if id, ok := mappings[strings.ToLower(asset)]; ok {
		return id
	}
	return strings.ToLower(asset)
}

// GetPrices fetches prices for multiple asset pairs
func (p *CoinGeckoProvider) GetPrices(ctx context.Context, pairs []AssetPair) (map[string]PriceData, error) {
	if p.closed.Load() {
		return nil, ErrSourceUnhealthy
	}

	if !p.circuit.Allow() {
		return nil, ErrSourceUnhealthy
	}

	if !p.rateLimiter.acquire() {
		return nil, ErrRateLimitExceeded
	}

	// Group by quote currency to minimize API calls
	byQuote := make(map[string][]string)
	for _, pair := range pairs {
		quote := strings.ToLower(pair.Quote)
		coinID := p.mapToCoinGeckoID(pair.Base)
		byQuote[quote] = append(byQuote[quote], coinID)
	}

	results := make(map[string]PriceData)

	for quote, coinIDs := range byQuote {
		// Build URL
		endpoint := fmt.Sprintf("%s/simple/price", p.config.APIURL)
		params := url.Values{}
		params.Set("ids", strings.Join(unique(coinIDs), ","))
		params.Set("vs_currencies", quote)
		params.Set("include_last_updated_at", "true")

		reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			continue
		}

		req.Header.Set("Accept", "application/json")
		if p.config.APIKey != "" {
			if p.config.UsePro {
				req.Header.Set("x-cg-pro-api-key", p.config.APIKey)
			} else {
				req.Header.Set("x-cg-demo-api-key", p.config.APIKey)
			}
		}

		resp, err := p.client.Do(req)
		if err != nil {
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			continue
		}

		var result map[string]map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		for coinID, data := range result {
			if priceVal, ok := data[quote]; ok {
				if price, err := p.parseFloat(priceVal); err == nil {
					baseAsset := p.reverseMapCoinGeckoID(coinID)
					key := baseAsset + "/" + strings.ToUpper(quote)
					results[key] = PriceData{
						BaseAsset:  baseAsset,
						QuoteAsset: strings.ToUpper(quote),
						Price:      price,
						Timestamp:  time.Now().UTC(),
						Source:     p.name,
						Confidence: 0.95,
					}
				}
			}
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

// reverseMapCoinGeckoID maps CoinGecko IDs back to internal IDs
func (p *CoinGeckoProvider) reverseMapCoinGeckoID(coinID string) string {
	mappings := map[string]string{
		"virtengine": "uve",
		"cosmos":     "atom",
		"usd-coin":   "usdc",
		"tether":     "usdt",
		"ethereum":   "eth",
		"bitcoin":    "btc",
	}

	if id, ok := mappings[coinID]; ok {
		return id
	}
	return coinID
}

// IsHealthy checks if the provider is responding
func (p *CoinGeckoProvider) IsHealthy(ctx context.Context) bool {
	if p.closed.Load() {
		return false
	}

	p.healthMu.RLock()
	health := p.health.Healthy
	lastSuccess := p.health.LastSuccess
	p.healthMu.RUnlock()

	// Consider unhealthy if no successful request in 5 minutes
	if time.Since(lastSuccess) > 5*time.Minute {
		return false
	}

	return health && p.circuit.State() != CircuitOpen
}

// Health returns detailed health information
func (p *CoinGeckoProvider) Health() SourceHealth {
	p.healthMu.RLock()
	defer p.healthMu.RUnlock()
	health := p.health
	health.Healthy = p.circuit.State() != CircuitOpen && p.health.Healthy
	return health
}

// Close closes the provider
func (p *CoinGeckoProvider) Close() error {
	p.closed.Store(true)
	return nil
}

// unique returns unique strings from a slice
func unique(strs []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range strs {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
