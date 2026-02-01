// Package dex provides DEX integration for crypto-to-fiat conversions.
//
// VE-905: DEX integration for crypto-to-fiat conversions
package dex

import (
	"context"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
)

// ============================================================================
// Types Tests
// ============================================================================

func TestTradingPair_PairID(t *testing.T) {
	pair := TradingPair{
		BaseToken:  Token{Symbol: "UVE"},
		QuoteToken: Token{Symbol: "USDC"},
	}

	expected := "UVE/USDC"
	if got := pair.PairID(); got != expected {
		t.Errorf("PairID() = %q, want %q", got, expected)
	}
}

func TestTradingPair_Reverse(t *testing.T) {
	pair := TradingPair{
		BaseToken:  Token{Symbol: "UVE"},
		QuoteToken: Token{Symbol: "USDC"},
	}

	reversed := pair.Reverse()
	if reversed.BaseToken.Symbol != "USDC" {
		t.Errorf("Reverse().BaseToken.Symbol = %q, want %q", reversed.BaseToken.Symbol, "USDC")
	}
	if reversed.QuoteToken.Symbol != "UVE" {
		t.Errorf("Reverse().QuoteToken.Symbol = %q, want %q", reversed.QuoteToken.Symbol, "UVE")
	}
}

func TestPrice_IsStale(t *testing.T) {
	tests := []struct {
		name     string
		age      time.Duration
		maxAge   time.Duration
		expected bool
	}{
		{"fresh price", 1 * time.Second, 5 * time.Minute, false},
		{"stale price", 10 * time.Minute, 5 * time.Minute, true},
		{"exactly at limit", 5 * time.Minute, 5 * time.Minute, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price := Price{
				Timestamp: time.Now().Add(-tt.age),
			}
			if got := price.IsStale(tt.maxAge); got != tt.expected {
				t.Errorf("IsStale() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSwapRequest_Validate(t *testing.T) {
	validRequest := SwapRequest{
		FromToken:         Token{Symbol: "UVE"},
		ToToken:           Token{Symbol: "USDC"},
		Amount:            sdkmath.NewInt(1000000),
		Type:              SwapTypeExactIn,
		SlippageTolerance: 0.01,
		Sender:            "virtengine1abc...",
	}

	tests := []struct {
		name    string
		modify  func(*SwapRequest)
		wantErr bool
	}{
		{"valid request", func(r *SwapRequest) {}, false},
		{"missing from token", func(r *SwapRequest) { r.FromToken.Symbol = "" }, true},
		{"missing to token", func(r *SwapRequest) { r.ToToken.Symbol = "" }, true},
		{"same tokens", func(r *SwapRequest) { r.ToToken.Symbol = r.FromToken.Symbol }, true},
		{"zero amount", func(r *SwapRequest) { r.Amount = sdkmath.ZeroInt() }, true},
		{"negative amount", func(r *SwapRequest) { r.Amount = sdkmath.NewInt(-100) }, true},
		{"slippage too high", func(r *SwapRequest) { r.SlippageTolerance = 1.5 }, true},
		{"slippage negative", func(r *SwapRequest) { r.SlippageTolerance = -0.1 }, true},
		{"missing sender", func(r *SwapRequest) { r.Sender = "" }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validRequest
			tt.modify(&req)
			err := req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSwapQuote_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{"not expired", time.Now().Add(5 * time.Minute), false},
		{"expired", time.Now().Add(-5 * time.Minute), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quote := SwapQuote{ExpiresAt: tt.expiresAt}
			if got := quote.IsExpired(); got != tt.expected {
				t.Errorf("IsExpired() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestOffRampRequest_Validate(t *testing.T) {
	validRequest := OffRampRequest{
		CryptoToken:   Token{Symbol: "UVE"},
		CryptoAmount:  sdkmath.NewInt(1000000),
		FiatCurrency:  FiatUSD,
		PaymentMethod: PaymentMethodBankTransfer,
		Sender:        "virtengine1abc...",
	}

	tests := []struct {
		name    string
		modify  func(*OffRampRequest)
		wantErr bool
	}{
		{"valid request", func(r *OffRampRequest) {}, false},
		{"missing crypto token", func(r *OffRampRequest) { r.CryptoToken.Symbol = "" }, true},
		{"zero amount", func(r *OffRampRequest) { r.CryptoAmount = sdkmath.ZeroInt() }, true},
		{"missing fiat currency", func(r *OffRampRequest) { r.FiatCurrency = "" }, true},
		{"missing payment method", func(r *OffRampRequest) { r.PaymentMethod = "" }, true},
		{"missing sender", func(r *OffRampRequest) { r.Sender = "" }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validRequest
			tt.modify(&req)
			err := req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ============================================================================
// Config Tests
// ============================================================================

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.PriceFeed.MinSources < 1 {
		t.Error("DefaultConfig().PriceFeed.MinSources should be at least 1")
	}
	if cfg.Swap.DefaultSlippage <= 0 {
		t.Error("DefaultConfig().Swap.DefaultSlippage should be positive")
	}
	if cfg.Swap.MaxSlippage < cfg.Swap.DefaultSlippage {
		t.Error("DefaultConfig().Swap.MaxSlippage should be >= DefaultSlippage")
	}
	if !cfg.CircuitBreaker.Enabled {
		t.Error("DefaultConfig().CircuitBreaker.Enabled should be true")
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr bool
	}{
		{"valid default config", func(c *Config) {}, false},
		{"min sources zero", func(c *Config) { c.PriceFeed.MinSources = 0 }, true},
		{"max deviation zero", func(c *Config) { c.PriceFeed.MaxDeviation = 0 }, true},
		{"max deviation too high", func(c *Config) { c.PriceFeed.MaxDeviation = 1.0 }, true},
		{"default slippage zero", func(c *Config) { c.Swap.DefaultSlippage = 0 }, true},
		{"default slippage > max", func(c *Config) {
			c.Swap.DefaultSlippage = 0.2
			c.Swap.MaxSlippage = 0.1
		}, true},
		{"max hops zero", func(c *Config) { c.Swap.MaxHops = 0 }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.modify(&cfg)
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ============================================================================
// Service Tests
// ============================================================================

func TestNewService(t *testing.T) {
	cfg := DefaultConfig()
	svc, err := NewService(cfg)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	if svc == nil {
		t.Fatal("NewService() returned nil")
	}
}

func TestNewService_InvalidConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.PriceFeed.MinSources = 0 // Invalid

	_, err := NewService(cfg)
	if err == nil {
		t.Error("NewService() should fail with invalid config")
	}
}

func TestService_AdapterManagement(t *testing.T) {
	cfg := DefaultConfig()
	svc, err := NewService(cfg)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	// Create mock adapter
	adapterCfg := AdapterConfig{
		Name:    "test-uniswap",
		Type:    "uniswap_v2",
		Enabled: true,
		ContractAddresses: map[string]string{
			"factory": "0x1234...",
			"router":  "0x5678...",
		},
	}
	adapter, err := NewUniswapV2Adapter(adapterCfg)
	if err != nil {
		t.Fatalf("NewUniswapV2Adapter() error = %v", err)
	}

	// Register adapter
	err = svc.RegisterAdapter(adapter)
	if err != nil {
		t.Errorf("RegisterAdapter() error = %v", err)
	}

	// List adapters
	adapters := svc.ListAdapters()
	if len(adapters) != 1 || adapters[0] != "test-uniswap" {
		t.Errorf("ListAdapters() = %v, want [test-uniswap]", adapters)
	}

	// Get adapter
	retrieved, err := svc.GetAdapter("test-uniswap")
	if err != nil {
		t.Errorf("GetAdapter() error = %v", err)
	}
	if retrieved.Name() != "test-uniswap" {
		t.Errorf("GetAdapter().Name() = %q, want %q", retrieved.Name(), "test-uniswap")
	}

	// Unregister adapter
	err = svc.UnregisterAdapter("test-uniswap")
	if err != nil {
		t.Errorf("UnregisterAdapter() error = %v", err)
	}

	// Verify removed
	adapters = svc.ListAdapters()
	if len(adapters) != 0 {
		t.Errorf("ListAdapters() after unregister = %v, want []", adapters)
	}
}

func TestService_GetAdapter_NotFound(t *testing.T) {
	cfg := DefaultConfig()
	svc, _ := NewService(cfg)

	_, err := svc.GetAdapter("nonexistent")
	if err != ErrAdapterNotFound {
		t.Errorf("GetAdapter() error = %v, want ErrAdapterNotFound", err)
	}
}

func TestService_Lifecycle(t *testing.T) {
	cfg := DefaultConfig()
	svc, _ := NewService(cfg)

	ctx := context.Background()

	// Start
	err := svc.Start(ctx)
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}

	// Double start should fail
	err = svc.Start(ctx)
	if err == nil {
		t.Error("Start() should fail when already started")
	}

	// Stop
	err = svc.Stop()
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}

// ============================================================================
// Circuit Breaker Tests
// ============================================================================

func TestCircuitBreaker_NotTrippedByDefault(t *testing.T) {
	cfg := CircuitBreakerConfig{Enabled: true, MaxFailuresPerMinute: 5}
	cb := newCircuitBreaker(cfg)

	if cb.IsTripped() {
		t.Error("Circuit breaker should not be tripped by default")
	}
}

func TestCircuitBreaker_DisabledNeverTrips(t *testing.T) {
	cfg := CircuitBreakerConfig{Enabled: false, MaxFailuresPerMinute: 1}
	cb := newCircuitBreaker(cfg)

	for i := 0; i < 10; i++ {
		cb.RecordFailure()
	}

	if cb.IsTripped() {
		t.Error("Disabled circuit breaker should never trip")
	}
}

func TestCircuitBreaker_TripsOnFailures(t *testing.T) {
	cfg := CircuitBreakerConfig{
		Enabled:              true,
		MaxFailuresPerMinute: 3,
		CooldownPeriod:       100 * time.Millisecond,
	}
	cb := newCircuitBreaker(cfg)

	// Record failures
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	if !cb.IsTripped() {
		t.Error("Circuit breaker should trip after max failures")
	}
}

func TestCircuitBreaker_RecoveryOnSuccess(t *testing.T) {
	cfg := CircuitBreakerConfig{
		Enabled:              true,
		MaxFailuresPerMinute: 2,
		RecoveryThreshold:    2,
		CooldownPeriod:       1 * time.Hour, // Long cooldown to test recovery via success
	}
	cb := newCircuitBreaker(cfg)

	// Trip the breaker
	cb.RecordFailure()
	cb.RecordFailure()

	if !cb.IsTripped() {
		t.Fatal("Circuit breaker should be tripped")
	}

	// Record successes to recover
	cb.RecordSuccess()
	cb.RecordSuccess()

	if cb.IsTripped() {
		t.Error("Circuit breaker should recover after success threshold")
	}
}

func TestCircuitBreaker_PriceDeviation(t *testing.T) {
	cfg := CircuitBreakerConfig{
		Enabled:                 true,
		PriceDeviationThreshold: 0.10, // 10%
	}
	cb := newCircuitBreaker(cfg)

	// Normal deviation
	tripped := cb.CheckPriceDeviation(100, 95)
	if tripped {
		t.Error("5% deviation should not trip")
	}

	// High deviation
	tripped = cb.CheckPriceDeviation(100, 80)
	if !tripped {
		t.Error("25% deviation should trip")
	}

	if !cb.IsTripped() {
		t.Error("Circuit breaker should be tripped after high deviation")
	}
}

// ============================================================================
// Adapter Tests
// ============================================================================

func TestCreateAdapter_UniswapV2(t *testing.T) {
	cfg := AdapterConfig{
		Name:    "test",
		Type:    "uniswap_v2",
		Enabled: true,
		ContractAddresses: map[string]string{
			"factory": "0x1234...",
			"router":  "0x5678...",
		},
	}

	adapter, err := CreateAdapter(cfg)
	if err != nil {
		t.Fatalf("CreateAdapter() error = %v", err)
	}
	if adapter.Name() != "test" {
		t.Errorf("Adapter.Name() = %q, want %q", adapter.Name(), "test")
	}
	if adapter.Type() != "uniswap_v2" {
		t.Errorf("Adapter.Type() = %q, want %q", adapter.Type(), "uniswap_v2")
	}
}

func TestCreateAdapter_Osmosis(t *testing.T) {
	cfg := AdapterConfig{
		Name:        "test-osmosis",
		Type:        "osmosis",
		Enabled:     true,
		RPCEndpoint: "grpc://localhost:9090",
	}

	adapter, err := CreateAdapter(cfg)
	if err != nil {
		t.Fatalf("CreateAdapter() error = %v", err)
	}
	if adapter.Type() != "osmosis" {
		t.Errorf("Adapter.Type() = %q, want %q", adapter.Type(), "osmosis")
	}
}

func TestCreateAdapter_Curve(t *testing.T) {
	cfg := AdapterConfig{
		Name:    "test-curve",
		Type:    "curve",
		Enabled: true,
		ContractAddresses: map[string]string{
			"registry": "0xabcd...",
		},
	}

	adapter, err := CreateAdapter(cfg)
	if err != nil {
		t.Fatalf("CreateAdapter() error = %v", err)
	}
	if adapter.Type() != "curve" {
		t.Errorf("Adapter.Type() = %q, want %q", adapter.Type(), "curve")
	}
}

func TestCreateAdapter_Unsupported(t *testing.T) {
	cfg := AdapterConfig{
		Name:    "test",
		Type:    "unsupported_dex",
		Enabled: true,
	}

	_, err := CreateAdapter(cfg)
	if err == nil {
		t.Error("CreateAdapter() should fail for unsupported type")
	}
}

func TestAdapter_IsHealthy(t *testing.T) {
	cfg := AdapterConfig{
		Name:    "test",
		Type:    "osmosis",
		Enabled: true,
	}

	adapter, _ := NewOsmosisAdapter(cfg)
	ctx := context.Background()

	if !adapter.IsHealthy(ctx) {
		t.Error("New adapter should be healthy")
	}

	_ = adapter.Close()

	if adapter.IsHealthy(ctx) {
		t.Error("Closed adapter should not be healthy")
	}
}

// ============================================================================
// Off-Ramp Tests
// ============================================================================

func TestMockOffRampProvider(t *testing.T) {
	provider := NewMockOffRampProvider(
		"test-provider",
		[]FiatCurrency{FiatUSD, FiatEUR},
		[]PaymentMethod{PaymentMethodBankTransfer},
	)

	ctx := context.Background()

	// Test health
	if !provider.IsHealthy(ctx) {
		t.Error("Mock provider should be healthy")
	}

	// Test currency support
	if !provider.SupportsCurrency(FiatUSD) {
		t.Error("Provider should support USD")
	}
	if provider.SupportsCurrency(FiatJPY) {
		t.Error("Provider should not support JPY")
	}

	// Test method support
	if !provider.SupportsMethod(PaymentMethodBankTransfer) {
		t.Error("Provider should support bank transfer")
	}
	if provider.SupportsMethod(PaymentMethodCard) {
		t.Error("Provider should not support card")
	}

	// Test quote
	request := OffRampRequest{
		CryptoToken:   Token{Symbol: "UVE"},
		CryptoAmount:  sdkmath.NewInt(1000000000), // 1000 tokens
		FiatCurrency:  FiatUSD,
		PaymentMethod: PaymentMethodBankTransfer,
		Sender:        "virtengine1abc...",
	}

	quote, err := provider.GetQuote(ctx, request)
	if err != nil {
		t.Fatalf("GetQuote() error = %v", err)
	}

	if quote.Provider != "test-provider" {
		t.Errorf("Quote.Provider = %q, want %q", quote.Provider, "test-provider")
	}
	if !quote.CryptoAmount.Equal(request.CryptoAmount) {
		t.Errorf("Quote.CryptoAmount = %s, want %s", quote.CryptoAmount, request.CryptoAmount)
	}
}

func TestOffRampBridge_ValidateKYC(t *testing.T) {
	cfg := OffRampConfig{
		MinVEIDScore: 500,
	}
	bridge := newOffRampBridge(cfg)

	tests := []struct {
		name    string
		score   int64
		wantErr bool
	}{
		{"sufficient score", 600, false},
		{"exactly minimum", 500, false},
		{"insufficient score", 400, true},
		{"zero score", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bridge.ValidateKYC(context.Background(), "virtengine1...", tt.score)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateKYC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ============================================================================
// Price Feed Tests
// ============================================================================

func TestPriceFeed_PairKey(t *testing.T) {
	key := pairKey("UVE", "USDC")
	expected := "UVE/USDC"
	if key != expected {
		t.Errorf("pairKey() = %q, want %q", key, expected)
	}
}

func TestPriceCache_GetSet(t *testing.T) {
	cache := &priceCache{
		prices: make(map[string]Price),
		ttl:    1 * time.Minute,
	}

	price := Price{
		Rate:      sdkmath.LegacyOneDec(),
		Timestamp: time.Now().UTC(),
	}

	cache.set("UVE/USDC", price)

	retrieved, ok := cache.get("UVE/USDC")
	if !ok {
		t.Fatal("Cache get failed")
	}
	if !retrieved.Rate.Equal(price.Rate) {
		t.Errorf("Retrieved rate = %s, want %s", retrieved.Rate, price.Rate)
	}
}

func TestPriceCache_Expiry(t *testing.T) {
	cache := &priceCache{
		prices: make(map[string]Price),
		ttl:    1 * time.Millisecond, // Very short TTL
	}

	price := Price{
		Rate:      sdkmath.LegacyOneDec(),
		Timestamp: time.Now().Add(-1 * time.Hour), // Old timestamp
	}

	cache.set("UVE/USDC", price)

	// Wait for TTL
	time.Sleep(2 * time.Millisecond)

	_, ok := cache.get("UVE/USDC")
	if ok {
		t.Error("Expired cache entry should not be returned")
	}
}

func TestPriceHistory_Add(t *testing.T) {
	history := &priceHistory{
		entries: make(map[string][]priceEntry),
		maxAge:  1 * time.Hour,
	}

	price := sdkmath.LegacyNewDec(100)
	volume := sdkmath.NewInt(1000)

	history.add("UVE/USDC", price, volume)

	entries := history.entries["UVE/USDC"]
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}
	if !entries[0].price.Equal(price) {
		t.Errorf("Entry price = %s, want %s", entries[0].price, price)
	}
}

// ============================================================================
// Pool Query Tests
// ============================================================================

func TestMatchesPoolQuery(t *testing.T) {
	pool := LiquidityPool{
		ID:             "pool-1",
		DEX:            "uniswap",
		Type:           PoolTypeConstantProduct,
		Tokens:         []Token{{Symbol: "UVE"}, {Symbol: "USDC"}},
		TotalLiquidity: sdkmath.LegacyNewDec(1000000),
	}

	tests := []struct {
		name    string
		query   PoolQuery
		matches bool
	}{
		{"empty query matches all", PoolQuery{}, true},
		{"matching DEX", PoolQuery{DEX: "uniswap"}, true},
		{"non-matching DEX", PoolQuery{DEX: "osmosis"}, false},
		{"matching type", PoolQuery{PoolType: PoolTypeConstantProduct}, true},
		{"non-matching type", PoolQuery{PoolType: PoolTypeStableSwap}, false},
		{"matching tokens", PoolQuery{TokenSymbols: []string{"UVE"}}, true},
		{"non-matching tokens", PoolQuery{TokenSymbols: []string{"BTC"}}, false},
		{"min liquidity met", PoolQuery{MinLiquidity: sdkmath.LegacyNewDec(500000)}, true},
		{"min liquidity not met", PoolQuery{MinLiquidity: sdkmath.LegacyNewDec(2000000)}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesPoolQuery(pool, tt.query); got != tt.matches {
				t.Errorf("matchesPoolQuery() = %v, want %v", got, tt.matches)
			}
		})
	}
}

