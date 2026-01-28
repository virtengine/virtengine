// Package load contains load and performance tests for VirtEngine.
// These tests verify system behavior under high load conditions.
//
// Task Reference: VE-801 - Load & performance testing
package load

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	mathrand "math/rand/v2"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// =============================================================================
// Scenario A: Identity Scope Upload Burst
// =============================================================================

// BenchmarkIdentityUploadBurst benchmarks burst identity scope uploads.
func BenchmarkIdentityUploadBurst(b *testing.B) {
	client := NewMockChainClient()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			payload := generateIdentityPayload()
			_, err := client.SubmitIdentityUpload(payload)
			if err != nil {
				b.Fatalf("identity upload failed: %v", err)
			}
		}
	})
}

// TestIdentityBurstLoad tests identity upload throughput under burst load.
func TestIdentityBurstLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	t.Log("=== Load Test: Identity Scope Upload Burst ===")

	config := LoadTestConfig{
		Concurrency:     50,
		Duration:        30 * time.Second,
		RampUpDuration:  5 * time.Second,
		TargetTPS:       100,
	}

	client := NewMockChainClient()
	results := runIdentityBurstTest(t, client, config)

	// Validate performance baselines
	t.Logf("Total requests: %d", results.TotalRequests)
	t.Logf("Successful: %d", results.SuccessfulRequests)
	t.Logf("Failed: %d", results.FailedRequests)
	t.Logf("P50 latency: %v", results.P50Latency)
	t.Logf("P95 latency: %v", results.P95Latency)
	t.Logf("P99 latency: %v", results.P99Latency)
	t.Logf("Throughput: %.2f TPS", results.Throughput)

	// Check baseline targets
	require.Less(t, results.P95Latency, 5*time.Second,
		"P95 latency should be under 5 seconds")
	require.Greater(t, results.Throughput, float64(50),
		"Throughput should be at least 50 TPS")
	require.Less(t, float64(results.FailedRequests)/float64(results.TotalRequests), 0.01,
		"Error rate should be under 1%")
}

func runIdentityBurstTest(t *testing.T, client *MockChainClient, config LoadTestConfig) *LoadTestResults {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), config.Duration+config.RampUpDuration)
	defer cancel()

	results := &LoadTestResults{
		Latencies: make([]time.Duration, 0),
	}
	var mu sync.Mutex

	var activeWorkers int32 = 0
	var wg sync.WaitGroup

	// Ramp up workers
	workersPerStep := config.Concurrency / 5
	rampInterval := config.RampUpDuration / 5

	for step := 0; step < 5; step++ {
		for i := 0; i < workersPerStep; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				atomic.AddInt32(&activeWorkers, 1)
				defer atomic.AddInt32(&activeWorkers, -1)

				for {
					select {
					case <-ctx.Done():
						return
					default:
						start := time.Now()
						payload := generateIdentityPayload()
						_, err := client.SubmitIdentityUpload(payload)
						elapsed := time.Since(start)

						mu.Lock()
						results.TotalRequests++
						if err != nil {
							results.FailedRequests++
						} else {
							results.SuccessfulRequests++
							results.Latencies = append(results.Latencies, elapsed)
						}
						mu.Unlock()

						// Throttle to target TPS
						if config.TargetTPS > 0 {
							workers := atomic.LoadInt32(&activeWorkers)
							if workers > 0 {
								delay := time.Second / time.Duration(config.TargetTPS/int(workers))
								time.Sleep(delay)
							}
						}
					}
				}
			}()
		}
		time.Sleep(rampInterval)
	}

	wg.Wait()

	// Calculate percentiles
	results.calculatePercentiles()
	results.Throughput = float64(results.SuccessfulRequests) / config.Duration.Seconds()

	return results
}

// =============================================================================
// Scenario B: Marketplace Order Burst
// =============================================================================

// BenchmarkMarketplaceOrderBurst benchmarks burst order submissions.
func BenchmarkMarketplaceOrderBurst(b *testing.B) {
	client := NewMockChainClient()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			order := generateMarketplaceOrder()
			_, err := client.SubmitOrder(order)
			if err != nil {
				b.Fatalf("order submission failed: %v", err)
			}
		}
	})
}

// TestMarketplaceBurstLoad tests marketplace throughput under burst load.
func TestMarketplaceBurstLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	t.Log("=== Load Test: Marketplace Order Burst ===")

	config := LoadTestConfig{
		Concurrency:     30,
		Duration:        30 * time.Second,
		RampUpDuration:  5 * time.Second,
		TargetTPS:       50,
	}

	client := NewMockChainClient()
	results := runMarketplaceBurstTest(t, client, config)

	t.Logf("Total orders: %d", results.TotalRequests)
	t.Logf("Successful: %d", results.SuccessfulRequests)
	t.Logf("Failed: %d", results.FailedRequests)
	t.Logf("P50 latency: %v", results.P50Latency)
	t.Logf("P95 latency: %v", results.P95Latency)
	t.Logf("Throughput: %.2f orders/sec", results.Throughput)

	// Check baseline targets
	require.Less(t, results.P95Latency, 30*time.Second,
		"P95 latency should be under 30 seconds")
}

func runMarketplaceBurstTest(t *testing.T, client *MockChainClient, config LoadTestConfig) *LoadTestResults {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), config.Duration)
	defer cancel()

	results := &LoadTestResults{
		Latencies: make([]time.Duration, 0),
	}
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()

					// Full order lifecycle: create -> bid -> allocate
					order := generateMarketplaceOrder()
					orderID, err := client.SubmitOrder(order)

					if err == nil {
						// Simulate bid
						_, err = client.SubmitBid(orderID, generateBid())
					}

					if err == nil {
						// Simulate allocation
						_, err = client.AllocateOrder(orderID)
					}

					elapsed := time.Since(start)

					mu.Lock()
					results.TotalRequests++
					if err != nil {
						results.FailedRequests++
					} else {
						results.SuccessfulRequests++
						results.Latencies = append(results.Latencies, elapsed)
					}
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()
	results.calculatePercentiles()
	results.Throughput = float64(results.SuccessfulRequests) / config.Duration.Seconds()

	return results
}

// =============================================================================
// Scenario C: HPC Job Submission Burst
// =============================================================================

// BenchmarkHPCJobSubmission benchmarks HPC job submissions.
func BenchmarkHPCJobSubmission(b *testing.B) {
	client := NewMockChainClient()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			job := generateHPCJob()
			_, err := client.SubmitHPCJob(job)
			if err != nil {
				b.Fatalf("job submission failed: %v", err)
			}
		}
	})
}

// TestHPCBurstLoad tests HPC scheduling throughput under burst load.
func TestHPCBurstLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	t.Log("=== Load Test: HPC Job Submission Burst ===")

	config := LoadTestConfig{
		Concurrency:     20,
		Duration:        30 * time.Second,
		RampUpDuration:  5 * time.Second,
		TargetTPS:       30,
	}

	client := NewMockChainClient()
	results := runHPCBurstTest(t, client, config)

	t.Logf("Total jobs: %d", results.TotalRequests)
	t.Logf("Successful: %d", results.SuccessfulRequests)
	t.Logf("Failed: %d", results.FailedRequests)
	t.Logf("P50 latency: %v", results.P50Latency)
	t.Logf("P95 latency: %v", results.P95Latency)
	t.Logf("Throughput: %.2f jobs/sec", results.Throughput)

	// Check baseline targets
	require.Less(t, results.P95Latency, 10*time.Second,
		"P95 latency should be under 10 seconds")
}

func runHPCBurstTest(t *testing.T, client *MockChainClient, config LoadTestConfig) *LoadTestResults {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), config.Duration)
	defer cancel()

	results := &LoadTestResults{
		Latencies: make([]time.Duration, 0),
	}
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()

					job := generateHPCJob()
					jobID, err := client.SubmitHPCJob(job)

					if err == nil {
						// Wait for scheduling
						_, err = client.WaitForJobScheduled(jobID, 5*time.Second)
					}

					elapsed := time.Since(start)

					mu.Lock()
					results.TotalRequests++
					if err != nil {
						results.FailedRequests++
					} else {
						results.SuccessfulRequests++
						results.Latencies = append(results.Latencies, elapsed)
					}
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()
	results.calculatePercentiles()
	results.Throughput = float64(results.SuccessfulRequests) / config.Duration.Seconds()

	return results
}

// =============================================================================
// Backpressure Tests
// =============================================================================

// TestDaemonBackpressure tests provider daemon backpressure handling.
func TestDaemonBackpressure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	t.Log("=== Load Test: Daemon Backpressure ===")

	daemon := NewMockProviderDaemon(DaemonConfig{
		MaxConcurrentJobs:    10,
		EventBufferSize:      100,
		ProcessingTimeout:    5 * time.Second,
	})

	// Generate events faster than daemon can process
	eventCount := 1000
	events := make([]*ChainEvent, eventCount)
	for i := 0; i < eventCount; i++ {
		events[i] = &ChainEvent{
			Type:      "order_created",
			OrderID:   fmt.Sprintf("order_%d", i),
			Timestamp: time.Now().UTC(),
		}
	}

	// Submit all events rapidly
	start := time.Now()
	for _, event := range events {
		err := daemon.EnqueueEvent(event)
		if err != nil {
			// Backpressure should cause some rejections
			t.Logf("Event rejected (backpressure): %v", err)
		}
	}
	submitDuration := time.Since(start)

	// Wait for processing
	time.Sleep(2 * time.Second)

	stats := daemon.GetStats()
	t.Logf("Submit duration: %v", submitDuration)
	t.Logf("Events queued: %d", stats.EventsQueued)
	t.Logf("Events processed: %d", stats.EventsProcessed)
	t.Logf("Events dropped: %d", stats.EventsDropped)
	t.Logf("Processing backlog: %d", stats.ProcessingBacklog)

	// Verify backpressure behavior
	require.GreaterOrEqual(t, stats.EventsProcessed+stats.EventsDropped, eventCount,
		"All events should be either processed or dropped")
}

// =============================================================================
// Types and Helpers
// =============================================================================

type LoadTestConfig struct {
	Concurrency     int
	Duration        time.Duration
	RampUpDuration  time.Duration
	TargetTPS       int
}

type LoadTestResults struct {
	TotalRequests      int
	SuccessfulRequests int
	FailedRequests     int
	Latencies          []time.Duration
	P50Latency         time.Duration
	P95Latency         time.Duration
	P99Latency         time.Duration
	Throughput         float64
}

func (r *LoadTestResults) calculatePercentiles() {
	if len(r.Latencies) == 0 {
		return
	}

	// Sort latencies (simple insertion sort for testing)
	sorted := make([]time.Duration, len(r.Latencies))
	copy(sorted, r.Latencies)
	for i := 1; i < len(sorted); i++ {
		key := sorted[i]
		j := i - 1
		for j >= 0 && sorted[j] > key {
			sorted[j+1] = sorted[j]
			j--
		}
		sorted[j+1] = key
	}

	r.P50Latency = sorted[len(sorted)*50/100]
	r.P95Latency = sorted[len(sorted)*95/100]
	r.P99Latency = sorted[len(sorted)*99/100]
}

// MockChainClient simulates chain interactions for load testing
type MockChainClient struct {
	mu              sync.Mutex
	identityCounter int64
	orderCounter    int64
	jobCounter      int64
}

func NewMockChainClient() *MockChainClient {
	return &MockChainClient{}
}

func (c *MockChainClient) SubmitIdentityUpload(payload *IdentityPayload) (string, error) {
	// Simulate processing delay
	time.Sleep(time.Duration(50+mathrand.IntN(100)) * time.Millisecond)
	id := atomic.AddInt64(&c.identityCounter, 1)
	return fmt.Sprintf("identity_%d", id), nil
}

func (c *MockChainClient) SubmitOrder(order *MarketplaceOrder) (string, error) {
	time.Sleep(time.Duration(100+mathrand.IntN(200)) * time.Millisecond)
	id := atomic.AddInt64(&c.orderCounter, 1)
	return fmt.Sprintf("order_%d", id), nil
}

func (c *MockChainClient) SubmitBid(orderID string, bid *ProviderBid) (string, error) {
	time.Sleep(time.Duration(50+mathrand.IntN(100)) * time.Millisecond)
	return "bid_" + orderID, nil
}

func (c *MockChainClient) AllocateOrder(orderID string) (string, error) {
	time.Sleep(time.Duration(100+mathrand.IntN(200)) * time.Millisecond)
	return "alloc_" + orderID, nil
}

func (c *MockChainClient) SubmitHPCJob(job *HPCJob) (string, error) {
	time.Sleep(time.Duration(50+mathrand.IntN(100)) * time.Millisecond)
	id := atomic.AddInt64(&c.jobCounter, 1)
	return fmt.Sprintf("job_%d", id), nil
}

func (c *MockChainClient) WaitForJobScheduled(jobID string, timeout time.Duration) (string, error) {
	time.Sleep(time.Duration(100+mathrand.IntN(200)) * time.Millisecond)
	return "scheduled", nil
}

type IdentityPayload struct {
	Scopes    []byte
	Salt      []byte
	Signature []byte
}

type MarketplaceOrder struct {
	CustomerID string
	OfferingID string
	Quantity   int
	Config     map[string]string
}

type ProviderBid struct {
	ProviderID string
	Price      int64
}

type HPCJob struct {
	UserID     string
	Manifest   []byte
	Resources  HPCResources
}

type HPCResources struct {
	CPUs   int
	Memory int64
	GPUs   int
}

type DaemonConfig struct {
	MaxConcurrentJobs int
	EventBufferSize   int
	ProcessingTimeout time.Duration
}

type MockProviderDaemon struct {
	config          DaemonConfig
	eventQueue      chan *ChainEvent
	stats           DaemonStats
	mu              sync.Mutex
}

type ChainEvent struct {
	Type      string
	OrderID   string
	Timestamp time.Time
}

type DaemonStats struct {
	EventsQueued      int
	EventsProcessed   int
	EventsDropped     int
	ProcessingBacklog int
}

func NewMockProviderDaemon(config DaemonConfig) *MockProviderDaemon {
	d := &MockProviderDaemon{
		config:     config,
		eventQueue: make(chan *ChainEvent, config.EventBufferSize),
	}

	// Start workers
	for i := 0; i < config.MaxConcurrentJobs; i++ {
		go d.processEvents()
	}

	return d
}

func (d *MockProviderDaemon) EnqueueEvent(event *ChainEvent) error {
	select {
	case d.eventQueue <- event:
		d.mu.Lock()
		d.stats.EventsQueued++
		d.mu.Unlock()
		return nil
	default:
		d.mu.Lock()
		d.stats.EventsDropped++
		d.mu.Unlock()
		return fmt.Errorf("event queue full")
	}
}

func (d *MockProviderDaemon) processEvents() {
	for event := range d.eventQueue {
		// Simulate processing
		_ = event
		time.Sleep(100 * time.Millisecond)

		d.mu.Lock()
		d.stats.EventsProcessed++
		d.mu.Unlock()
	}
}

func (d *MockProviderDaemon) GetStats() DaemonStats {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.stats.ProcessingBacklog = len(d.eventQueue)
	return d.stats
}

func generateIdentityPayload() *IdentityPayload {
	scopes := make([]byte, 1024)
	salt := make([]byte, 32)
	sig := make([]byte, 64)
	io.ReadFull(rand.Reader, scopes)
	io.ReadFull(rand.Reader, salt)
	io.ReadFull(rand.Reader, sig)
	return &IdentityPayload{Scopes: scopes, Salt: salt, Signature: sig}
}

func generateMarketplaceOrder() *MarketplaceOrder {
	id := make([]byte, 8)
	io.ReadFull(rand.Reader, id)
	return &MarketplaceOrder{
		CustomerID: "customer_" + hex.EncodeToString(id),
		OfferingID: "offering_001",
		Quantity:   1,
		Config:     map[string]string{"region": "us-east"},
	}
}

func generateBid() *ProviderBid {
	return &ProviderBid{
		ProviderID: "provider_001",
		Price:      1000,
	}
}

func generateHPCJob() *HPCJob {
	id := make([]byte, 8)
	io.ReadFull(rand.Reader, id)
	return &HPCJob{
		UserID:   "user_" + hex.EncodeToString(id),
		Manifest: []byte(`#!/bin/bash\necho "Hello HPC"`),
		Resources: HPCResources{
			CPUs:   4,
			Memory: 8 * 1024 * 1024 * 1024,
			GPUs:   0,
		},
	}
}

// rand.Intn replacement for deterministic testing
func init() {
	// Use crypto/rand for better randomness in load tests
}
