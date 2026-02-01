// Package scale contains performance degradation analysis.
// These tests identify bottlenecks and measure performance at various scales.
//
// Task Reference: SCALE-001 - Load Testing - 1M Nodes Simulation
package scale

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ============================================================================
// Performance Analysis Constants
// ============================================================================

const (
	// Scale levels for analysis
	ScaleSmall    = 1000
	ScaleMedium   = 10000
	ScaleLarge    = 100000
	ScaleXLarge   = 1000000

	// Degradation thresholds
	AcceptableDegradation   = 2.0  // 2x slower is acceptable
	WarningDegradation      = 5.0  // 5x slower is warning
	CriticalDegradation     = 10.0 // 10x slower is critical

	// Severity levels
	severityLow = "low"
)

// ============================================================================
// Analysis Types
// ============================================================================

// PerformanceMetric represents a single performance measurement
type PerformanceMetric struct {
	Operation     string        `json:"operation"`
	Scale         int           `json:"scale"`
	Duration      time.Duration `json:"duration"`
	Throughput    float64       `json:"throughput"`
	MemoryUsedMB  int64         `json:"memory_used_mb"`
	GoroutineCount int          `json:"goroutine_count"`
	Timestamp     time.Time     `json:"timestamp"`
}

// ScalingAnalysis represents analysis of how an operation scales
type ScalingAnalysis struct {
	Operation      string             `json:"operation"`
	Metrics        []PerformanceMetric `json:"metrics"`
	ScalingFactor  float64            `json:"scaling_factor"` // O(n) = 1.0, O(n^2) = 2.0
	Bottleneck     string             `json:"bottleneck"`
	Recommendation string             `json:"recommendation"`
}

// DegradationReport summarizes performance degradation findings
type DegradationReport struct {
	GeneratedAt    time.Time          `json:"generated_at"`
	Environment    EnvironmentInfo    `json:"environment"`
	Analyses       []ScalingAnalysis  `json:"analyses"`
	Bottlenecks    []Bottleneck       `json:"bottlenecks"`
	Recommendations []Recommendation  `json:"recommendations"`
}

// EnvironmentInfo captures test environment details
type EnvironmentInfo struct {
	GOOS        string `json:"goos"`
	GOARCH      string `json:"goarch"`
	NumCPU      int    `json:"num_cpu"`
	GoVersion   string `json:"go_version"`
	MaxProcs    int    `json:"max_procs"`
}

// Bottleneck identifies a performance bottleneck
type Bottleneck struct {
	Component    string  `json:"component"`
	Severity     string  `json:"severity"` // low, medium, high, critical
	Description  string  `json:"description"`
	Impact       string  `json:"impact"`
	ScaleFactor  float64 `json:"scale_factor"`
}

// Recommendation provides scaling guidance
type Recommendation struct {
	Priority    int    `json:"priority"` // 1=highest
	Category    string `json:"category"` // performance, architecture, resource
	Title       string `json:"title"`
	Description string `json:"description"`
	Effort      string `json:"effort"` // low, medium, high
}

// PerformanceAnalyzer runs performance analysis tests
type PerformanceAnalyzer struct {
	mu       sync.Mutex
	metrics  []PerformanceMetric
	analyses []ScalingAnalysis
}

// NewPerformanceAnalyzer creates a new analyzer
func NewPerformanceAnalyzer() *PerformanceAnalyzer {
	return &PerformanceAnalyzer{
		metrics:  make([]PerformanceMetric, 0),
		analyses: make([]ScalingAnalysis, 0),
	}
}

// RecordMetric records a performance metric
func (a *PerformanceAnalyzer) RecordMetric(metric PerformanceMetric) {
	a.mu.Lock()
	defer a.mu.Unlock()
	metric.Timestamp = time.Now()
	// Sanitize throughput to avoid NaN/Inf
	if math.IsNaN(metric.Throughput) || math.IsInf(metric.Throughput, 0) {
		metric.Throughput = 0
	}
	a.metrics = append(a.metrics, metric)
}

// AnalyzeOperation analyzes scaling behavior of an operation
func (a *PerformanceAnalyzer) AnalyzeOperation(operation string) *ScalingAnalysis {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	// Find metrics for this operation
	var opMetrics []PerformanceMetric
	for _, m := range a.metrics {
		if m.Operation == operation {
			opMetrics = append(opMetrics, m)
		}
	}
	
	if len(opMetrics) < 2 {
		return nil
	}
	
	// Sort by scale
	sort.Slice(opMetrics, func(i, j int) bool {
		return opMetrics[i].Scale < opMetrics[j].Scale
	})
	
	analysis := &ScalingAnalysis{
		Operation: operation,
		Metrics:   opMetrics,
	}
	
	// Calculate scaling factor
	if len(opMetrics) >= 2 {
		first := opMetrics[0]
		last := opMetrics[len(opMetrics)-1]
		
		scaleRatio := float64(last.Scale) / float64(first.Scale)
		timeRatio := float64(last.Duration) / float64(first.Duration)
		
		if scaleRatio > 0 && first.Duration > 0 && timeRatio > 0 {
			// log(timeRatio) / log(scaleRatio) gives the exponent
			// O(n) -> 1, O(n^2) -> 2, O(log n) -> ~0.3
			analysis.ScalingFactor = timeRatio / scaleRatio
		} else {
			// Default to linear if we can't compute (including when first.Duration == 0)
			analysis.ScalingFactor = 1.0
		}
		
		// Handle NaN/Inf
		if math.IsNaN(analysis.ScalingFactor) || math.IsInf(analysis.ScalingFactor, 0) {
			analysis.ScalingFactor = 1.0
		}
	}
	
	// Determine bottleneck and recommendation
	if analysis.ScalingFactor > 2.0 {
		analysis.Bottleneck = "super-linear scaling detected"
		analysis.Recommendation = "Consider algorithmic improvements or caching"
	} else if analysis.ScalingFactor > 1.0 {
		analysis.Bottleneck = "linear scaling"
		analysis.Recommendation = "Scaling is acceptable, consider horizontal scaling for larger loads"
	} else {
		analysis.Bottleneck = "sub-linear scaling (optimal)"
		analysis.Recommendation = "Scaling is efficient"
	}
	
	a.analyses = append(a.analyses, *analysis)
	return analysis
}

// GenerateReport generates a comprehensive degradation report
func (a *PerformanceAnalyzer) GenerateReport() *DegradationReport {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	report := &DegradationReport{
		GeneratedAt: time.Now(),
		Environment: EnvironmentInfo{
			GOOS:      runtime.GOOS,
			GOARCH:    runtime.GOARCH,
			NumCPU:    runtime.NumCPU(),
			GoVersion: runtime.Version(),
			MaxProcs:  runtime.GOMAXPROCS(0),
		},
		Analyses: a.analyses,
	}
	
	// Identify bottlenecks
	for _, analysis := range a.analyses {
		severity := severityLow
		if analysis.ScalingFactor > CriticalDegradation {
			severity = "critical"
		} else if analysis.ScalingFactor > WarningDegradation {
			severity = "high"
		} else if analysis.ScalingFactor > AcceptableDegradation {
			severity = "medium"
		}
		
		if severity != severityLow {
			report.Bottlenecks = append(report.Bottlenecks, Bottleneck{
				Component:   analysis.Operation,
				Severity:    severity,
				Description: analysis.Bottleneck,
				ScaleFactor: analysis.ScalingFactor,
			})
		}
	}
	
	// Generate recommendations
	report.Recommendations = generateRecommendations(a.analyses)
	
	return report
}

func generateRecommendations(analyses []ScalingAnalysis) []Recommendation {
	var recs []Recommendation
	priority := 1
	
	// Sort analyses by scaling factor (worst first)
	sorted := make([]ScalingAnalysis, len(analyses))
	copy(sorted, analyses)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ScalingFactor > sorted[j].ScalingFactor
	})
	
	for _, a := range sorted {
		if a.ScalingFactor > AcceptableDegradation {
			rec := Recommendation{
				Priority: priority,
				Category: "performance",
				Title:    fmt.Sprintf("Optimize %s", a.Operation),
				Description: fmt.Sprintf("%s shows %.1fx degradation at scale. %s", 
					a.Operation, a.ScalingFactor, a.Recommendation),
			}
			
			if a.ScalingFactor > CriticalDegradation {
				rec.Effort = "high"
			} else if a.ScalingFactor > WarningDegradation {
				rec.Effort = "medium"
			} else {
				rec.Effort = "low"
			}
			
			recs = append(recs, rec)
			priority++
		}
	}
	
	// Add general recommendations
	recs = append(recs, Recommendation{
		Priority:    priority,
		Category:    "architecture",
		Title:       "Implement horizontal scaling",
		Description: "For 1M+ validators, consider sharding validator set across multiple nodes",
		Effort:      "high",
	})
	
	recs = append(recs, Recommendation{
		Priority:    priority + 1,
		Category:    "resource",
		Title:       "Increase memory allocation",
		Description: "Large-scale operations require significant memory. Plan for 16GB+ per node at 1M scale",
		Effort:      "low",
	})
	
	return recs
}

// ============================================================================
// Analysis Tests
// ============================================================================

// TestValidatorScaleDegradation analyzes validator operations scaling
func TestValidatorScaleDegradation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping degradation analysis in short mode")
	}
	
	t.Logf("=== Validator Scale Degradation Analysis ===")
	
	analyzer := NewPerformanceAnalyzer()
	scales := []int{1000, 5000, 10000, 50000}
	
	for _, scale := range scales {
		t.Run(fmt.Sprintf("scale_%d", scale), func(t *testing.T) {
			// Measure store population
			var m runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&m)
			memBefore := m.HeapAlloc
			
			start := time.Now()
			store := populateValidatorStore(scale)
			populateTime := time.Since(start)
			
			runtime.GC()
			runtime.ReadMemStats(&m)
			memAfter := m.HeapAlloc
			
			analyzer.RecordMetric(PerformanceMetric{
				Operation:      "validator_store_population",
				Scale:          scale,
				Duration:       populateTime,
				Throughput:     float64(scale) / populateTime.Seconds(),
				MemoryUsedMB:   int64((memAfter - memBefore) / 1024 / 1024),
				GoroutineCount: runtime.NumGoroutine(),
			})
			
			t.Logf("Scale %d: population=%v, throughput=%.0f/sec, memory=%dMB",
				scale, populateTime, float64(scale)/populateTime.Seconds(),
				(memAfter-memBefore)/1024/1024)
			
			// Measure lookup
			addresses := make([][20]byte, 100)
			for i := range addresses {
				addresses[i] = generateMockValidator(i * (scale / 100)).Address
			}
			
			start = time.Now()
			for _, addr := range addresses {
				store.GetValidator(addr)
			}
			lookupTime := time.Since(start)
			
			analyzer.RecordMetric(PerformanceMetric{
				Operation:  "validator_lookup",
				Scale:      scale,
				Duration:   lookupTime,
				Throughput: float64(len(addresses)) / lookupTime.Seconds(),
			})
			
			// Measure iteration
			start = time.Now()
			store.IterateValidators(func(v *MockValidator) bool {
				return false
			})
			iterTime := time.Since(start)
			
			analyzer.RecordMetric(PerformanceMetric{
				Operation:  "validator_iteration",
				Scale:      scale,
				Duration:   iterTime,
				Throughput: float64(scale) / iterTime.Seconds(),
			})
		})
	}
	
	// Analyze results
	popAnalysis := analyzer.AnalyzeOperation("validator_store_population")
	if popAnalysis != nil {
		t.Logf("Population scaling factor: %.2f (%s)", popAnalysis.ScalingFactor, popAnalysis.Bottleneck)
	}
	
	lookupAnalysis := analyzer.AnalyzeOperation("validator_lookup")
	if lookupAnalysis != nil {
		t.Logf("Lookup scaling factor: %.2f (%s)", lookupAnalysis.ScalingFactor, lookupAnalysis.Bottleneck)
	}
	
	iterAnalysis := analyzer.AnalyzeOperation("validator_iteration")
	if iterAnalysis != nil {
		t.Logf("Iteration scaling factor: %.2f (%s)", iterAnalysis.ScalingFactor, iterAnalysis.Bottleneck)
	}
}

// TestMarketplaceScaleDegradation analyzes marketplace operations scaling
func TestMarketplaceScaleDegradation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping marketplace degradation analysis in short mode")
	}
	
	t.Logf("=== Marketplace Scale Degradation Analysis ===")
	
	analyzer := NewPerformanceAnalyzer()
	scales := []int{1000, 5000, 10000}
	
	for _, scale := range scales {
		t.Run(fmt.Sprintf("scale_%d", scale), func(t *testing.T) {
			// Measure marketplace population
			start := time.Now()
			store := populateMarketplace(scale, 10, 100)
			populateTime := time.Since(start)
			
			orders, bids, _, _, _, _ := store.GetStats()
			
			analyzer.RecordMetric(PerformanceMetric{
				Operation:  "marketplace_population",
				Scale:      scale,
				Duration:   populateTime,
				Throughput: float64(orders+bids) / populateTime.Seconds(),
			})
			
			t.Logf("Scale %d: population=%v, orders=%d, bids=%d",
				scale, populateTime, orders, bids)
			
			// Measure order creation
			start = time.Now()
			for i := 0; i < 100; i++ {
				store.CreateOrder(generateRandomAddress(), generateOrderSpecs(), 1000)
			}
			createTime := time.Since(start)
			
			analyzer.RecordMetric(PerformanceMetric{
				Operation:  "order_creation",
				Scale:      scale,
				Duration:   createTime,
				Throughput: 100.0 / createTime.Seconds(),
			})
			
			// Measure bid submission
			start = time.Now()
			for i := 0; i < 100; i++ {
				_, _ = store.SubmitBid(uint64(i%scale+1), generateRandomAddress(), 500)
			}
			bidTime := time.Since(start)
			
			analyzer.RecordMetric(PerformanceMetric{
				Operation:  "bid_submission",
				Scale:      scale,
				Duration:   bidTime,
				Throughput: 100.0 / bidTime.Seconds(),
			})
		})
	}
	
	// Analyze
	popAnalysis := analyzer.AnalyzeOperation("marketplace_population")
	if popAnalysis != nil {
		t.Logf("Population scaling: %.2f", popAnalysis.ScalingFactor)
	}
	
	createAnalysis := analyzer.AnalyzeOperation("order_creation")
	if createAnalysis != nil {
		t.Logf("Order creation scaling: %.2f", createAnalysis.ScalingFactor)
	}
	
	bidAnalysis := analyzer.AnalyzeOperation("bid_submission")
	if bidAnalysis != nil {
		t.Logf("Bid submission scaling: %.2f", bidAnalysis.ScalingFactor)
	}
}

// TestProviderScaleDegradation analyzes provider operations scaling
func TestProviderScaleDegradation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping provider degradation analysis in short mode")
	}
	
	t.Logf("=== Provider Scale Degradation Analysis ===")
	
	analyzer := NewPerformanceAnalyzer()
	scales := []int{50, 100, 200, 500}
	
	for _, scale := range scales {
		t.Run(fmt.Sprintf("providers_%d", scale), func(t *testing.T) {
			// Measure pool creation
			start := time.Now()
			pool := NewProviderPool(scale)
			createTime := time.Since(start)
			
			analyzer.RecordMetric(PerformanceMetric{
				Operation:      "provider_pool_creation",
				Scale:          scale,
				Duration:       createTime,
				Throughput:     float64(scale) / createTime.Seconds(),
				GoroutineCount: runtime.NumGoroutine(),
			})
			
			// Measure startup
			start = time.Now()
			pool.Start()
			startTime := time.Since(start)
			
			analyzer.RecordMetric(PerformanceMetric{
				Operation:      "provider_pool_startup",
				Scale:          scale,
				Duration:       startTime,
				Throughput:     float64(scale) / startTime.Seconds(),
				GoroutineCount: runtime.NumGoroutine(),
			})
			
			// Measure event broadcast
			start = time.Now()
			for i := 0; i < 100; i++ {
				pool.BroadcastEvent(&ProviderEvent{
					Type:      "order_created",
					OrderID:   uint64(i),
					Timestamp: time.Now(),
				})
			}
			broadcastTime := time.Since(start)
			
			analyzer.RecordMetric(PerformanceMetric{
				Operation:  "event_broadcast",
				Scale:      scale,
				Duration:   broadcastTime,
				Throughput: 100.0 / broadcastTime.Seconds(),
			})
			
			pool.Stop()
			
			t.Logf("Scale %d: create=%v, start=%v, broadcast=%v",
				scale, createTime, startTime, broadcastTime)
		})
	}
	
	// Analyze
	for _, op := range []string{"provider_pool_creation", "provider_pool_startup", "event_broadcast"} {
		analysis := analyzer.AnalyzeOperation(op)
		if analysis != nil {
			t.Logf("%s scaling: %.2f (%s)", op, analysis.ScalingFactor, analysis.Bottleneck)
		}
	}
}

// TestComprehensiveDegradationReport generates a full degradation report
func TestComprehensiveDegradationReport(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comprehensive report in short mode")
	}
	
	t.Logf("=== Comprehensive Degradation Report ===")
	
	analyzer := NewPerformanceAnalyzer()
	
	// Run validator tests
	t.Log("Analyzing validator operations...")
	for _, scale := range []int{1000, 5000, 10000} {
		start := time.Now()
		store := populateValidatorStore(scale)
		populateTime := time.Since(start)
		
		analyzer.RecordMetric(PerformanceMetric{
			Operation:  "validator_population",
			Scale:      scale,
			Duration:   populateTime,
			Throughput: float64(scale) / populateTime.Seconds(),
		})
		
		start = time.Now()
		store.CalculateTotalVotingPower()
		powerTime := time.Since(start)
		
		analyzer.RecordMetric(PerformanceMetric{
			Operation:  "voting_power_calculation",
			Scale:      scale,
			Duration:   powerTime,
			Throughput: float64(scale) / powerTime.Seconds(),
		})
	}
	
	// Run marketplace tests
	t.Log("Analyzing marketplace operations...")
	for _, scale := range []int{1000, 5000, 10000} {
		start := time.Now()
		store := populateMarketplace(scale, 5, 50)
		populateTime := time.Since(start)
		
		orders, bids, _, _, _, _ := store.GetStats()
		
		analyzer.RecordMetric(PerformanceMetric{
			Operation:  "marketplace_setup",
			Scale:      scale,
			Duration:   populateTime,
			Throughput: float64(orders+bids) / populateTime.Seconds(),
		})
	}
	
	// Run state sync tests
	t.Log("Analyzing state sync operations...")
	for _, scale := range []int{1000, 5000, 10000} {
		start := time.Now()
		store := populateStateStore(scale, 256)
		populateTime := time.Since(start)
		
		analyzer.RecordMetric(PerformanceMetric{
			Operation:  "state_population",
			Scale:      scale,
			Duration:   populateTime,
			Throughput: float64(scale) / populateTime.Seconds(),
		})
		
		manager := NewSnapshotManager(store, ChunkSize)
		start = time.Now()
		snapshot, _ := manager.CreateSnapshot()
		snapshotTime := time.Since(start)
		
		analyzer.RecordMetric(PerformanceMetric{
			Operation:  "snapshot_creation",
			Scale:      scale,
			Duration:   snapshotTime,
			Throughput: float64(snapshot.TotalSize) / snapshotTime.Seconds(),
		})
	}
	
	// Analyze all operations
	operations := []string{
		"validator_population",
		"voting_power_calculation",
		"marketplace_setup",
		"state_population",
		"snapshot_creation",
	}
	
	for _, op := range operations {
		analyzer.AnalyzeOperation(op)
	}
	
	// Generate report
	report := analyzer.GenerateReport()
	
	// Print summary
	t.Log("\n=== DEGRADATION REPORT SUMMARY ===")
	t.Logf("Environment: %s/%s, %d CPUs, %s", 
		report.Environment.GOOS, report.Environment.GOARCH,
		report.Environment.NumCPU, report.Environment.GoVersion)
	
	t.Log("\nScaling Analysis:")
	for _, a := range report.Analyses {
		status := "✓"
		if a.ScalingFactor > WarningDegradation {
			status = "⚠"
		} else if a.ScalingFactor > CriticalDegradation {
			status = "✗"
		}
		t.Logf("  %s %s: %.2fx (%s)", status, a.Operation, a.ScalingFactor, a.Bottleneck)
	}
	
	if len(report.Bottlenecks) > 0 {
		t.Log("\nBottlenecks Identified:")
		for _, b := range report.Bottlenecks {
			t.Logf("  [%s] %s: %s (%.1fx)", b.Severity, b.Component, b.Description, b.ScaleFactor)
		}
	}
	
	t.Log("\nRecommendations:")
	for _, r := range report.Recommendations {
		t.Logf("  %d. [%s] %s - %s", r.Priority, r.Effort, r.Title, r.Description)
	}
	
	// Save report to JSON
	t.Run("save_report", func(t *testing.T) {
		reportJSON, err := json.MarshalIndent(report, "", "  ")
		require.NoError(t, err)
		
		// Print to test output
		t.Logf("\nFull Report JSON:\n%s", string(reportJSON))
	})
}

// TestConcurrentLoadDegradation tests degradation under concurrent load
func TestConcurrentLoadDegradation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent load test in short mode")
	}
	
	t.Logf("=== Concurrent Load Degradation Analysis ===")
	
	concurrencyLevels := []int{1, 2, 4, 8, 16}
	duration := 5 * time.Second
	
	var results []struct {
		concurrency int
		throughput  float64
		avgLatency  time.Duration
	}
	
	for _, concurrency := range concurrencyLevels {
		t.Run(fmt.Sprintf("concurrency_%d", concurrency), func(t *testing.T) {
			store := populateValidatorStore(10000)
			
			ctx, cancel := context.WithTimeout(context.Background(), duration)
			defer cancel()
			
			var ops atomic.Int64
			var totalLatency atomic.Int64
			
			var wg sync.WaitGroup
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func(workerID int) {
					defer wg.Done()
					
					for {
						select {
						case <-ctx.Done():
							return
						default:
							start := time.Now()
							
							// Mixed operations
							addr := generateMockValidator(workerID * 100).Address
							store.GetValidator(addr)
							store.CalculateTotalVotingPower()
							
							latency := time.Since(start)
							totalLatency.Add(int64(latency))
							ops.Add(1)
						}
					}
				}(i)
			}
			
			wg.Wait()
			
			totalOps := ops.Load()
			throughput := float64(totalOps) / duration.Seconds()
			avgLatency := time.Duration(totalLatency.Load() / totalOps)
			
			results = append(results, struct {
				concurrency int
				throughput  float64
				avgLatency  time.Duration
			}{concurrency, throughput, avgLatency})
			
			t.Logf("Concurrency %d: %.0f ops/sec, avg latency %v", concurrency, throughput, avgLatency)
		})
	}
	
	// Analyze scaling efficiency
	if len(results) >= 2 {
		baselineOps := results[0].throughput
		for i := 1; i < len(results); i++ {
			expectedOps := baselineOps * float64(results[i].concurrency)
			actualOps := results[i].throughput
			efficiency := actualOps / expectedOps * 100
			
			t.Logf("Concurrency %d: %.1f%% scaling efficiency",
				results[i].concurrency, efficiency)
		}
	}
}

// TestMemoryDegradation tests memory usage at various scales
func TestMemoryDegradation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory degradation test in short mode")
	}
	
	t.Logf("=== Memory Degradation Analysis ===")
	
	scales := []int{1000, 5000, 10000, 50000}
	
	type memResult struct {
		scale       int
		heapMB      uint64
		allocsMB    uint64
		perEntryB   uint64
	}
	
	var results []memResult
	
	for _, scale := range scales {
		t.Run(fmt.Sprintf("scale_%d", scale), func(t *testing.T) {
			runtime.GC()
			var before runtime.MemStats
			runtime.ReadMemStats(&before)
			
			// Allocate mixed data structures
			valStore := populateValidatorStore(scale)
			mktStore := populateMarketplace(scale/10, 5, 50)
			stateStore := populateStateStore(scale, 256)
			
			runtime.GC()
			var after runtime.MemStats
			runtime.ReadMemStats(&after)
			
			heapUsed := (after.HeapAlloc - before.HeapAlloc) / 1024 / 1024
			allocsUsed := (after.TotalAlloc - before.TotalAlloc) / 1024 / 1024
			perEntry := (after.HeapAlloc - before.HeapAlloc) / uint64(scale*2)
			
			results = append(results, memResult{
				scale:     scale,
				heapMB:    heapUsed,
				allocsMB:  allocsUsed,
				perEntryB: perEntry,
			})
			
			t.Logf("Scale %d: heap=%dMB, allocs=%dMB, per_entry=%dB",
				scale, heapUsed, allocsUsed, perEntry)
			
			// Keep references alive
			_ = valStore.Count()
			orders, _, _, _, _, _ := mktStore.GetStats()
			_ = orders
			_ = stateStore.Count()
		})
	}
	
	// Analyze memory scaling
	if len(results) >= 2 {
		first := results[0]
		last := results[len(results)-1]
		
		scaleRatio := float64(last.scale) / float64(first.scale)
		memRatio := float64(last.heapMB) / float64(first.heapMB)
		
		t.Logf("\nMemory Scaling Analysis:")
		t.Logf("  Scale increase: %.1fx", scaleRatio)
		t.Logf("  Memory increase: %.1fx", memRatio)
		t.Logf("  Scaling efficiency: %.2f (1.0 = linear)", memRatio/scaleRatio)
		
		if memRatio > scaleRatio*1.5 {
			t.Log("  ⚠ WARNING: Super-linear memory growth detected")
		} else {
			t.Log("  ✓ Memory scaling is acceptable")
		}
	}
}

// SaveReportToFile saves report to a file (for CI integration)
func SaveReportToFile(report *DegradationReport, filename string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}
