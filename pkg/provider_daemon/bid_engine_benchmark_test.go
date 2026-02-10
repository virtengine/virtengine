// Package provider_daemon contains bidding latency benchmarks.
// Task Reference: PERF-001 - Provider Daemon Bidding Latency Benchmarks
package provider_daemon

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ============================================================================
// Provider Daemon Bidding Latency Benchmarks
// ============================================================================

// BiddingLatencyBaseline defines baseline metrics for bidding latency
type BiddingLatencyBaseline struct {
	TargetLatency              time.Duration `json:"target_latency"`
	MaxLatencyP95              time.Duration `json:"max_latency_p95"`
	MaxLatencyP99              time.Duration `json:"max_latency_p99"`
	MinBidsPerSecond           float64       `json:"min_bids_per_second"`
	MaxOrderMatchingLatency    time.Duration `json:"max_order_matching_latency"`
	MaxPriceCalculationLatency time.Duration `json:"max_price_calculation_latency"`
	MaxSigningLatency          time.Duration `json:"max_signing_latency"`
}

// DefaultBiddingBaseline returns baseline metrics for bidding
func DefaultBiddingBaseline() BiddingLatencyBaseline {
	return BiddingLatencyBaseline{
		TargetLatency:              50 * time.Millisecond,
		MaxLatencyP95:              100 * time.Millisecond,
		MaxLatencyP99:              200 * time.Millisecond,
		MinBidsPerSecond:           20.0,
		MaxOrderMatchingLatency:    10 * time.Millisecond,
		MaxPriceCalculationLatency: 5 * time.Millisecond,
		MaxSigningLatency:          20 * time.Millisecond,
	}
}

// MockBidEngine provides a benchmarking version of the bid engine
type MockBidEngine struct {
	config      BidEngineConfig
	provConfig  *ProviderConfig
	rateLimiter *RateLimiter

	mu             sync.Mutex
	latencies      []time.Duration
	matchLatencies []time.Duration
	priceLatencies []time.Duration
	signLatencies  []time.Duration
	bidsPlaced     int64
	bidsFailed     int64
}

// NewMockBidEngine creates a new mock bid engine for benchmarking
func NewMockBidEngine() *MockBidEngine {
	config := DefaultBidEngineConfig()
	config.MaxBidsPerMinute = 10000 // High limit for benchmarking
	config.MaxBidsPerHour = 100000

	return &MockBidEngine{
		config:         config,
		provConfig:     createBenchmarkProviderConfig(),
		rateLimiter:    NewRateLimiter(config.MaxBidsPerMinute, config.MaxBidsPerHour),
		latencies:      make([]time.Duration, 0, 10000),
		matchLatencies: make([]time.Duration, 0, 10000),
		priceLatencies: make([]time.Duration, 0, 10000),
		signLatencies:  make([]time.Duration, 0, 10000),
	}
}

func createBenchmarkProviderConfig() *ProviderConfig {
	return &ProviderConfig{
		ProviderAddress:    "cosmos1provider123",
		SupportedOfferings: []string{"compute", "storage", "gpu"},
		Regions:            []string{"us-east", "us-west", "eu-west"},
		Active:             true,
		Pricing: PricingConfig{
			CPUPricePerCore:   "0.1",
			MemoryPricePerGB:  "0.05",
			StoragePricePerGB: "0.01",
			GPUPricePerHour:   "1.0",
			MinBidPrice:       "0.001",
			BidMarkupPercent:  10.0,
			Currency:          "uvirt",
		},
		Capacity: CapacityConfig{
			TotalCPUCores:    128,
			TotalMemoryGB:    512,
			TotalStorageGB:   10000,
			TotalGPUs:        8,
			ReservedCPUCores: 8,
			ReservedMemoryGB: 32,
		},
	}
}

// BidResult tracks individual bid result
type BidResultBenchmark struct {
	OrderID      string
	BidID        string
	Success      bool
	Error        error
	TotalLatency time.Duration
	MatchLatency time.Duration
	PriceLatency time.Duration
	SignLatency  time.Duration
}

// ProcessBid processes a bid for benchmarking
func (e *MockBidEngine) ProcessBid(order Order) *BidResultBenchmark {
	start := time.Now()
	result := &BidResultBenchmark{
		OrderID: order.OrderID,
	}

	// Step 1: Order Matching
	matchStart := time.Now()
	matched := e.matchOrder(order)
	result.MatchLatency = time.Since(matchStart)

	if !matched {
		result.Error = ErrOrderNotMatchable
		result.TotalLatency = time.Since(start)
		atomic.AddInt64(&e.bidsFailed, 1)
		return result
	}

	// Step 2: Price Calculation
	priceStart := time.Now()
	price, err := e.calculateBidPrice(order, e.provConfig)
	result.PriceLatency = time.Since(priceStart)

	if err != nil {
		result.Error = err
		result.TotalLatency = time.Since(start)
		atomic.AddInt64(&e.bidsFailed, 1)
		return result
	}

	// Step 3: Bid Signing
	signStart := time.Now()
	bidID, err := e.signBid(order.OrderID, price)
	result.SignLatency = time.Since(signStart)

	if err != nil {
		result.Error = err
		result.TotalLatency = time.Since(start)
		atomic.AddInt64(&e.bidsFailed, 1)
		return result
	}

	result.BidID = bidID
	result.Success = true
	result.TotalLatency = time.Since(start)

	// Record latencies
	e.mu.Lock()
	e.latencies = append(e.latencies, result.TotalLatency)
	e.matchLatencies = append(e.matchLatencies, result.MatchLatency)
	e.priceLatencies = append(e.priceLatencies, result.PriceLatency)
	e.signLatencies = append(e.signLatencies, result.SignLatency)
	e.mu.Unlock()

	atomic.AddInt64(&e.bidsPlaced, 1)
	return result
}

func (e *MockBidEngine) matchOrder(order Order) bool {
	// Simulate order matching logic
	if e.provConfig == nil || !e.provConfig.Active {
		return false
	}

	// Check offering type
	supported := false
	for _, t := range e.provConfig.SupportedOfferings {
		if t == order.OfferingType {
			supported = true
			break
		}
	}
	if !supported {
		return false
	}

	// Check region
	if order.Region != "" {
		regionMatch := false
		for _, r := range e.provConfig.Regions {
			if r == order.Region {
				regionMatch = true
				break
			}
		}
		if !regionMatch {
			return false
		}
	}

	// Check capacity
	capacity := e.provConfig.Capacity
	if order.Requirements.CPUCores > capacity.AvailableCPU() {
		return false
	}
	if order.Requirements.MemoryGB > capacity.AvailableMemory() {
		return false
	}
	if order.Requirements.StorageGB > capacity.AvailableStorage() {
		return false
	}

	return true
}

//nolint:unparam // result 1 (error) reserved for future pricing failures
func (e *MockBidEngine) calculateBidPrice(order Order, config *ProviderConfig) (string, error) {
	// Simulate price calculation based on order requirements
	time.Sleep(2 * time.Microsecond) // Minimal delay for calculation

	// Calculate base price from requirements
	req := order.Requirements
	cpuPrice := float64(req.CPUCores) * 0.1       // per core
	memPrice := float64(req.MemoryGB) * 0.05      // per GB
	storagePrice := float64(req.StorageGB) * 0.01 // per GB
	gpuPrice := float64(req.GPUs) * 1.0           // per GPU

	basePrice := cpuPrice + memPrice + storagePrice + gpuPrice

	// Apply markup
	finalPrice := basePrice * (1 + config.Pricing.BidMarkupPercent/100)

	return fmt.Sprintf("%.6fuvirt", finalPrice), nil
}

//nolint:unparam // price kept for future signature verification
func (e *MockBidEngine) signBid(orderID, _ string) (string, error) {
	// Simulate cryptographic signing
	time.Sleep(5 * time.Microsecond) // Minimal delay for signing

	// Generate bid ID
	idBytes := make([]byte, 8)
	_, _ = rand.Read(idBytes)

	providerPrefix := "provider"
	if len(e.config.ProviderAddress) >= 8 {
		providerPrefix = e.config.ProviderAddress[:8]
	} else if len(e.config.ProviderAddress) > 0 {
		providerPrefix = e.config.ProviderAddress
	}
	bidID := fmt.Sprintf("bid-%s-%s-%s", orderID, providerPrefix, hex.EncodeToString(idBytes))

	return bidID, nil
}

// GetStats returns bidding statistics
func (e *MockBidEngine) GetStats() (placed, failed int64, latencies, matchLat, priceLat, signLat []time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.bidsPlaced, e.bidsFailed, e.latencies, e.matchLatencies, e.priceLatencies, e.signLatencies
}

// BenchmarkOrderMatching benchmarks order matching logic
func BenchmarkOrderMatching(b *testing.B) {
	engine := NewMockBidEngine()
	order := createBenchmarkOrder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.matchOrder(order)
	}
}

// BenchmarkPriceCalculation benchmarks price calculation
func BenchmarkPriceCalculation(b *testing.B) {
	engine := NewMockBidEngine()
	order := createBenchmarkOrder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.calculateBidPrice(order, engine.provConfig)
	}
}

// BenchmarkBidSigning benchmarks bid signing
func BenchmarkBidSigning(b *testing.B) {
	engine := NewMockBidEngine()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.signBid("order_123", "1.5")
	}
}

// BenchmarkFullBidProcessing benchmarks full bid processing pipeline
func BenchmarkFullBidProcessing(b *testing.B) {
	engine := NewMockBidEngine()
	order := createBenchmarkOrder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := engine.ProcessBid(order)
		if !result.Success {
			b.Fatalf("bid processing failed: %v", result.Error)
		}
	}

	b.ReportMetric(float64(time.Second)/float64(b.Elapsed()/time.Duration(b.N)), "bids/sec")
}

// BenchmarkBidProcessingParallel benchmarks parallel bid processing
func BenchmarkBidProcessingParallel(b *testing.B) {
	engine := NewMockBidEngine()
	var counter atomic.Int64

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := counter.Add(1)
			order := Order{
				OrderID:      fmt.Sprintf("order_%d", i),
				OfferingType: "compute",
				Region:       "us-east",
				Requirements: ResourceRequirements{
					CPUCores:  4,
					MemoryGB:  8,
					StorageGB: 100,
				},
				MaxPrice: "10.0",
				Currency: "uvirt",
			}
			result := engine.ProcessBid(order)
			if !result.Success {
				b.Fatalf("bid processing failed: %v", result.Error)
			}
		}
	})

	placed, failed, _, _, _, _ := engine.GetStats()
	b.ReportMetric(float64(placed)/b.Elapsed().Seconds(), "bids/sec")
	b.ReportMetric(float64(failed)/float64(placed+failed)*100, "error_rate_%")
}

// BenchmarkBidProcessingWithVaryingResources benchmarks bids with different resource requirements
func BenchmarkBidProcessingWithVaryingResources(b *testing.B) {
	resourceProfiles := []struct {
		name    string
		cpu     int64
		memory  int64
		storage int64
		gpus    int64
	}{
		{"small", 1, 2, 10, 0},
		{"medium", 4, 8, 100, 0},
		{"large", 16, 64, 500, 0},
		{"gpu_small", 4, 16, 100, 1},
		{"gpu_large", 32, 128, 1000, 4},
	}

	for _, profile := range resourceProfiles {
		b.Run(profile.name, func(b *testing.B) {
			engine := NewMockBidEngine()
			order := Order{
				OrderID:      "order_bench",
				OfferingType: "compute",
				Region:       "us-east",
				Requirements: ResourceRequirements{
					CPUCores:  profile.cpu,
					MemoryGB:  profile.memory,
					StorageGB: profile.storage,
					GPUs:      profile.gpus,
				},
				MaxPrice: "100.0",
				Currency: "uvirt",
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result := engine.ProcessBid(order)
				if !result.Success && result.Error != ErrOrderNotMatchable {
					b.Fatalf("bid processing failed: %v", result.Error)
				}
			}
		})
	}
}

// BenchmarkRateLimiter benchmarks rate limiter performance
func BenchmarkRateLimiter(b *testing.B) {
	limiter := NewRateLimiter(1000000, 10000000) // Very high limits for benchmarking

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = limiter.Allow()
	}
}

// BenchmarkRateLimiterParallel benchmarks rate limiter under concurrent access
func BenchmarkRateLimiterParallel(b *testing.B) {
	limiter := NewRateLimiter(1000000, 10000000)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = limiter.Allow()
		}
	})
}

// TestBiddingLatencyBaseline tests bidding latency against baseline
func TestBiddingLatencyBaseline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping bidding latency baseline test in short mode")
	}

	baseline := DefaultBiddingBaseline()
	engine := NewMockBidEngine()

	const iterations = 1000
	var successCount int

	for i := 0; i < iterations; i++ {
		order := Order{
			OrderID:      fmt.Sprintf("order_%d", i),
			OfferingType: "compute",
			Region:       "us-east",
			Requirements: ResourceRequirements{
				CPUCores:  4,
				MemoryGB:  8,
				StorageGB: 100,
			},
			MaxPrice: "10.0",
			Currency: "uvirt",
		}

		result := engine.ProcessBid(order)
		if result.Success {
			successCount++
		}
	}

	placed, failed, latencies, matchLat, priceLat, signLat := engine.GetStats()

	// Calculate percentiles
	p95, p99 := calculateBiddingPercentiles(latencies)
	matchP95, _ := calculateBiddingPercentiles(matchLat)
	priceP95, _ := calculateBiddingPercentiles(priceLat)
	signP95, _ := calculateBiddingPercentiles(signLat)

	avgLatency := averageDuration(latencies)

	t.Logf("=== Provider Daemon Bidding Latency Baseline Test ===")
	t.Logf("Iterations: %d", iterations)
	t.Logf("Bids Placed: %d", placed)
	t.Logf("Bids Failed: %d", failed)
	t.Logf("Average Latency: %v (target: %v)", avgLatency, baseline.TargetLatency)
	t.Logf("P95 Latency: %v (max: %v)", p95, baseline.MaxLatencyP95)
	t.Logf("P99 Latency: %v (max: %v)", p99, baseline.MaxLatencyP99)
	t.Logf("Order Matching P95: %v (max: %v)", matchP95, baseline.MaxOrderMatchingLatency)
	t.Logf("Price Calculation P95: %v (max: %v)", priceP95, baseline.MaxPriceCalculationLatency)
	t.Logf("Signing P95: %v (max: %v)", signP95, baseline.MaxSigningLatency)

	// Assertions against baseline
	require.LessOrEqual(t, avgLatency, baseline.TargetLatency,
		"Average latency should meet target")
	require.LessOrEqual(t, p95, baseline.MaxLatencyP95,
		"P95 latency should be within acceptable limit")
}

// calculateBiddingPercentiles calculates P95 and P99 latencies
func calculateBiddingPercentiles(latencies []time.Duration) (p95, p99 time.Duration) {
	if len(latencies) == 0 {
		return 0, 0
	}

	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)

	for i := 1; i < len(sorted); i++ {
		key := sorted[i]
		j := i - 1
		for j >= 0 && sorted[j] > key {
			sorted[j+1] = sorted[j]
			j--
		}
		sorted[j+1] = key
	}

	p95Idx := int(float64(len(sorted)) * 0.95)
	p99Idx := int(float64(len(sorted)) * 0.99)

	if p95Idx >= len(sorted) {
		p95Idx = len(sorted) - 1
	}
	if p99Idx >= len(sorted) {
		p99Idx = len(sorted) - 1
	}

	return sorted[p95Idx], sorted[p99Idx]
}

func averageDuration(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	var total time.Duration
	for _, l := range latencies {
		total += l
	}
	return total / time.Duration(len(latencies))
}

func createBenchmarkOrder() Order {
	return Order{
		OrderID:         "order_benchmark",
		CustomerAddress: "cosmos1customer123",
		OfferingType:    "compute",
		Region:          "us-east",
		Requirements: ResourceRequirements{
			CPUCores:  4,
			MemoryGB:  8,
			StorageGB: 100,
		},
		MaxPrice:  "10.0",
		Currency:  "uvirt",
		CreatedAt: time.Now().UTC(),
	}
}

// ============================================================================
// Event Processing Benchmarks
// ============================================================================

// BenchmarkEventEnqueue benchmarks event queue enqueue operations
func BenchmarkEventEnqueue(b *testing.B) {
	queue := make(chan *mockEvent, 10000)
	event := &mockEvent{Type: "order_created", OrderID: "order_1"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		select {
		case queue <- event:
		default:
			// Queue full, drain and retry
			<-queue
			queue <- event
		}
	}
}

// BenchmarkEventProcessing benchmarks event processing throughput
func BenchmarkEventProcessing(b *testing.B) {
	queue := make(chan *mockEvent, 10000)
	var processed atomic.Int64

	// Start processor
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-queue:
				processed.Add(1)
			}
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event := &mockEvent{
			Type:    "order_created",
			OrderID: fmt.Sprintf("order_%d", i),
		}
		queue <- event
	}

	// Wait for processing
	for processed.Load() < int64(b.N) {
		time.Sleep(time.Microsecond)
	}
}

type mockEvent struct {
	Type    string
	OrderID string
}
