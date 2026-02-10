package pricefeed

import (
	"context"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
)

// pairUVEUSD is the cache key for the UVE/USD price pair
const pairUVEUSD = "uve/usd"

func TestMockProvider_GetPrice(t *testing.T) {
	provider := NewMockProvider("test-mock")
	ctx := context.Background()

	// Set a price
	price := sdkmath.LegacyMustNewDecFromStr("1.5")
	provider.SetPrice("uve", "usd", price)

	// Get the price
	result, err := provider.GetPrice(ctx, "uve", "usd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Price.Equal(price) {
		t.Errorf("expected price %s, got %s", price, result.Price)
	}

	if result.Source != "test-mock" {
		t.Errorf("expected source test-mock, got %s", result.Source)
	}
}

func TestMockProvider_NotFound(t *testing.T) {
	provider := NewMockProvider("test-mock")
	ctx := context.Background()

	_, err := provider.GetPrice(ctx, "unknown", "usd")
	if err != ErrPriceNotFound {
		t.Errorf("expected ErrPriceNotFound, got %v", err)
	}
}

func TestMockProvider_Error(t *testing.T) {
	provider := NewMockProvider("test-mock")
	ctx := context.Background()

	provider.SetError(ErrRateLimitExceeded)

	_, err := provider.GetPrice(ctx, "uve", "usd")
	if err != ErrRateLimitExceeded {
		t.Errorf("expected ErrRateLimitExceeded, got %v", err)
	}
}

func TestMockProvider_Unhealthy(t *testing.T) {
	provider := NewMockProvider("test-mock")
	ctx := context.Background()

	provider.SetHealthy(false)

	if provider.IsHealthy(ctx) {
		t.Error("expected provider to be unhealthy")
	}
}

func TestInMemoryCache_Basic(t *testing.T) {
	cfg := CacheConfig{
		Enabled:     true,
		TTL:         1 * time.Second,
		MaxSize:     100,
		AllowStale:  true,
		StaleMaxAge: 5 * time.Second,
	}

	cache := NewInMemoryCache(cfg)

	// Set a price
	price := PriceData{
		BaseAsset:  "uve",
		QuoteAsset: "usd",
		Price:      sdkmath.LegacyOneDec(),
		Timestamp:  time.Now(),
		Source:     "test",
	}

	key := pairUVEUSD
	cache.Set(key, price)

	// Get the price
	result, ok := cache.Get(key)
	if !ok {
		t.Fatal("expected cache hit")
	}

	if !result.Price.Equal(price.Price) {
		t.Errorf("expected price %s, got %s", price.Price, result.Price)
	}

	// Check stats
	stats := cache.Stats()
	if stats.Hits != 1 {
		t.Errorf("expected 1 hit, got %d", stats.Hits)
	}
}

func TestInMemoryCache_Expiry(t *testing.T) {
	cfg := CacheConfig{
		Enabled:     true,
		TTL:         50 * time.Millisecond,
		MaxSize:     100,
		AllowStale:  false,
		StaleMaxAge: 100 * time.Millisecond,
	}

	cache := NewInMemoryCache(cfg)

	price := PriceData{
		BaseAsset:  "uve",
		QuoteAsset: "usd",
		Price:      sdkmath.LegacyOneDec(),
		Timestamp:  time.Now(),
		Source:     "test",
	}

	key := pairUVEUSD
	cache.Set(key, price)

	// Wait for expiry
	time.Sleep(100 * time.Millisecond)

	// Should not find expired entry
	_, ok := cache.Get(key)
	if ok {
		t.Error("expected cache miss for expired entry")
	}
}

func TestInMemoryCache_StaleAllowed(t *testing.T) {
	cfg := CacheConfig{
		Enabled:     true,
		TTL:         50 * time.Millisecond,
		MaxSize:     100,
		AllowStale:  true,
		StaleMaxAge: 200 * time.Millisecond,
	}

	cache := NewInMemoryCache(cfg)

	price := PriceData{
		BaseAsset:  "uve",
		QuoteAsset: "usd",
		Price:      sdkmath.LegacyOneDec(),
		Timestamp:  time.Now(),
		Source:     "test",
	}

	key := pairUVEUSD
	cache.Set(key, price)

	// Wait for TTL expiry but within stale max age
	time.Sleep(100 * time.Millisecond)

	// Should still find stale entry
	_, ok := cache.Get(key)
	if !ok {
		t.Error("expected stale cache hit")
	}
}

func TestRetryer_Success(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries:    3,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
		RetryableErrors: []string{
			"timeout",
			"connection refused",
		},
	}

	retryer := NewRetryer(cfg)

	callCount := 0
	err := retryer.Do(context.Background(), func(ctx context.Context) error {
		callCount++
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

func TestRetryer_RetryOnError(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries:      2,
		InitialDelay:    10 * time.Millisecond,
		MaxDelay:        100 * time.Millisecond,
		BackoffFactor:   2.0,
		RetryableErrors: []string{"timeout"},
	}

	retryer := NewRetryer(cfg)

	callCount := 0
	err := retryer.Do(context.Background(), func(ctx context.Context) error {
		callCount++
		if callCount < 3 {
			return &retryableError{msg: "timeout occurred"}
		}
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

type retryableError struct {
	msg string
}

func (e *retryableError) Error() string {
	return e.msg
}

func TestCircuitBreaker_Open(t *testing.T) {
	cb := NewCircuitBreaker(3, 2, 100*time.Millisecond)

	// Record failures to open circuit
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	if cb.State() != CircuitOpen {
		t.Errorf("expected circuit open, got %s", cb.State())
	}

	if cb.Allow() {
		t.Error("expected request to be blocked when circuit is open")
	}
}

func TestCircuitBreaker_HalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(3, 2, 50*time.Millisecond)

	// Open the circuit
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Should be half-open now
	if cb.State() != CircuitHalfOpen {
		t.Errorf("expected circuit half-open, got %s", cb.State())
	}

	// Should allow requests
	if !cb.Allow() {
		t.Error("expected request to be allowed in half-open state")
	}
}

func TestCircuitBreaker_Close(t *testing.T) {
	cb := NewCircuitBreaker(3, 2, 50*time.Millisecond)

	// Open the circuit
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	// Wait for timeout to transition to half-open
	time.Sleep(60 * time.Millisecond)

	// Check state to trigger transition to half-open (Allow() does this)
	if !cb.Allow() {
		t.Error("expected circuit to allow in half-open")
	}

	// Record successes to close circuit
	cb.RecordSuccess()
	cb.RecordSuccess()

	if cb.State() != CircuitClosed {
		t.Errorf("expected circuit closed, got %s", cb.State())
	}
}

func TestPriceData_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		price PriceData
		valid bool
	}{
		{
			name: "valid price",
			price: PriceData{
				BaseAsset:  "uve",
				QuoteAsset: "usd",
				Price:      sdkmath.LegacyOneDec(),
				Timestamp:  time.Now(),
			},
			valid: true,
		},
		{
			name: "missing base asset",
			price: PriceData{
				QuoteAsset: "usd",
				Price:      sdkmath.LegacyOneDec(),
				Timestamp:  time.Now(),
			},
			valid: false,
		},
		{
			name: "zero price",
			price: PriceData{
				BaseAsset:  "uve",
				QuoteAsset: "usd",
				Price:      sdkmath.LegacyZeroDec(),
				Timestamp:  time.Now(),
			},
			valid: false,
		},
		{
			name: "negative price",
			price: PriceData{
				BaseAsset:  "uve",
				QuoteAsset: "usd",
				Price:      sdkmath.LegacyNewDec(-1),
				Timestamp:  time.Now(),
			},
			valid: false,
		},
		{
			name: "zero timestamp",
			price: PriceData{
				BaseAsset:  "uve",
				QuoteAsset: "usd",
				Price:      sdkmath.LegacyOneDec(),
			},
			valid: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.price.IsValid() != tc.valid {
				t.Errorf("expected valid=%v, got %v", tc.valid, tc.price.IsValid())
			}
		})
	}
}

func TestAggregator_PrimaryStrategy(t *testing.T) {
	// Create mock providers
	primary := NewMockProvider("primary")
	secondary := NewMockProvider("secondary")

	primaryPrice := sdkmath.LegacyMustNewDecFromStr("1.5")
	secondaryPrice := sdkmath.LegacyMustNewDecFromStr("1.6")

	primary.SetPrice("uve", "usd", primaryPrice)
	secondary.SetPrice("uve", "usd", secondaryPrice)

	// Create aggregator with primary strategy
	cfg := Config{
		Providers: []ProviderConfig{
			{Name: "primary", Type: SourceTypeMock, Enabled: true, Priority: 1},
			{Name: "secondary", Type: SourceTypeMock, Enabled: true, Priority: 2},
		},
		Strategy: StrategyPrimary,
		CacheConfig: CacheConfig{
			Enabled:     true,
			TTL:         30 * time.Second,
			MaxSize:     100,
			AllowStale:  true,
			StaleMaxAge: 5 * time.Minute,
		},
		RetryConfig: RetryConfig{
			MaxRetries:    1,
			InitialDelay:  10 * time.Millisecond,
			MaxDelay:      100 * time.Millisecond,
			BackoffFactor: 2.0,
		},
		MaxPriceDeviation:   0.1,
		HealthCheckInterval: 0, // Disable background health check
	}

	agg := &PriceFeedAggregator{
		config:          cfg,
		providers:       make(map[string]Provider),
		healthCheckStop: make(chan struct{}),
	}
	agg.cache = NewInMemoryCache(cfg.CacheConfig)
	agg.providers["primary"] = primary
	agg.providers["secondary"] = secondary

	ctx := context.Background()
	result, err := agg.GetPrice(ctx, "uve", "usd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should get price from primary
	if !result.Price.Equal(primaryPrice) {
		t.Errorf("expected primary price %s, got %s", primaryPrice, result.Price)
	}

	// Primary should have been called
	if primary.RequestCount() != 1 {
		t.Errorf("expected 1 primary request, got %d", primary.RequestCount())
	}

	agg.Close()
}

func TestAggregator_FallbackOnError(t *testing.T) {
	primary := NewMockProvider("primary")
	secondary := NewMockProvider("secondary")

	// Primary returns error
	primary.SetError(ErrRateLimitExceeded)

	secondaryPrice := sdkmath.LegacyMustNewDecFromStr("1.6")
	secondary.SetPrice("uve", "usd", secondaryPrice)

	cfg := Config{
		Providers: []ProviderConfig{
			{Name: "primary", Type: SourceTypeMock, Enabled: true, Priority: 1},
			{Name: "secondary", Type: SourceTypeMock, Enabled: true, Priority: 2},
		},
		Strategy: StrategyPrimary,
		CacheConfig: CacheConfig{
			Enabled:     true,
			TTL:         30 * time.Second,
			MaxSize:     100,
			AllowStale:  true,
			StaleMaxAge: 5 * time.Minute,
		},
		RetryConfig: RetryConfig{
			MaxRetries:    0, // No retries for faster test
			InitialDelay:  10 * time.Millisecond,
			MaxDelay:      100 * time.Millisecond,
			BackoffFactor: 2.0,
		},
		MaxPriceDeviation:   0.1,
		HealthCheckInterval: 0,
	}

	agg := &PriceFeedAggregator{
		config:          cfg,
		providers:       make(map[string]Provider),
		healthCheckStop: make(chan struct{}),
	}
	agg.cache = NewInMemoryCache(cfg.CacheConfig)
	agg.providers["primary"] = primary
	agg.providers["secondary"] = secondary

	ctx := context.Background()
	result, err := agg.GetPrice(ctx, "uve", "usd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should fall back to secondary
	if !result.Price.Equal(secondaryPrice) {
		t.Errorf("expected secondary price %s, got %s", secondaryPrice, result.Price)
	}

	agg.Close()
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid default config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "no providers",
			config: Config{
				Providers: []ProviderConfig{},
				Strategy:  StrategyPrimary,
			},
			wantErr: true,
		},
		{
			name: "all providers disabled",
			config: Config{
				Providers: []ProviderConfig{
					{Name: "test", Enabled: false},
				},
				Strategy:          StrategyPrimary,
				MaxPriceDeviation: 0.1,
			},
			wantErr: true,
		},
		{
			name: "invalid max deviation",
			config: Config{
				Providers: []ProviderConfig{
					{Name: "test", Type: SourceTypeCoinGecko, Enabled: true, RequestTimeout: 10 * time.Second},
				},
				Strategy:          StrategyPrimary,
				MaxPriceDeviation: 1.5, // > 1 is invalid
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestSourceType_IsValid(t *testing.T) {
	validTypes := []SourceType{SourceTypeCoinGecko, SourceTypeChainlink, SourceTypePyth, SourceTypeMock}
	for _, st := range validTypes {
		if !st.IsValid() {
			t.Errorf("expected %s to be valid", st)
		}
	}

	invalidType := SourceType("invalid")
	if invalidType.IsValid() {
		t.Error("expected invalid type to be invalid")
	}
}

func TestAggregationStrategy_IsValid(t *testing.T) {
	validStrategies := []AggregationStrategy{StrategyPrimary, StrategyMedian, StrategyWeighted}
	for _, s := range validStrategies {
		if !s.IsValid() {
			t.Errorf("expected %s to be valid", s)
		}
	}

	invalidStrategy := AggregationStrategy("invalid")
	if invalidStrategy.IsValid() {
		t.Error("expected invalid strategy to be invalid")
	}
}
