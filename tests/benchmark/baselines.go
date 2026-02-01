// Package benchmark provides baseline metrics for regression detection.
// Task Reference: PERF-001 - Baseline Metrics for Regression Detection
package benchmark

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	resultPass = "pass"
	resultFail = "fail"
	resultWarn = "warn"
)

// BaselineMetrics contains all baseline performance metrics
type BaselineMetrics struct {
	Version     string    `json:"version"`
	GeneratedAt time.Time `json:"generated_at"`
	Platform    string    `json:"platform"`
	GoVersion   string    `json:"go_version"`

	// Transaction Throughput Baselines
	Transaction TransactionBaselines `json:"transaction"`

	// VEID Verification Baselines
	Verification VerificationBaselines `json:"verification"`

	// Provider Daemon Bidding Baselines
	Bidding BiddingBaselines `json:"bidding"`

	// ZK Proof Generation Baselines
	ZKProof ZKProofBaselines `json:"zk_proof"`

	// Memory Baselines
	Memory MemoryBaselines `json:"memory"`
}

// TransactionBaselines defines transaction throughput baselines
type TransactionBaselines struct {
	TargetTPS         int64   `json:"target_tps"`
	MinAcceptableTPS  int64   `json:"min_acceptable_tps"`
	MaxP95LatencyMs   int64   `json:"max_p95_latency_ms"`
	MaxP99LatencyMs   int64   `json:"max_p99_latency_ms"`
	MaxErrorRatePct   float64 `json:"max_error_rate_pct"`
	HashComputeNs     int64   `json:"hash_compute_ns"`
	ValidationNs      int64   `json:"validation_ns"`
	ExecutionNs       int64   `json:"execution_ns"`
	StateWriteNs      int64   `json:"state_write_ns"`
}

// VerificationBaselines defines VEID verification baselines
type VerificationBaselines struct {
	TargetLatencyMs        int64   `json:"target_latency_ms"`
	MaxP95LatencyMs        int64   `json:"max_p95_latency_ms"`
	MaxP99LatencyMs        int64   `json:"max_p99_latency_ms"`
	MinThroughput          float64 `json:"min_throughput_per_sec"`
	MaxDecryptionMs        int64   `json:"max_decryption_ms"`
	MaxScoringMs           int64   `json:"max_scoring_ms"`
	MaxStateUpdateMs       int64   `json:"max_state_update_ms"`
	IdentityRecordCreateNs int64   `json:"identity_record_create_ns"`
	IdentityRecordGetNs    int64   `json:"identity_record_get_ns"`
	ScoreUpdateNs          int64   `json:"score_update_ns"`
}

// BiddingBaselines defines provider daemon bidding baselines
type BiddingBaselines struct {
	TargetLatencyMs     int64   `json:"target_latency_ms"`
	MaxP95LatencyMs     int64   `json:"max_p95_latency_ms"`
	MaxP99LatencyMs     int64   `json:"max_p99_latency_ms"`
	MinBidsPerSecond    float64 `json:"min_bids_per_sec"`
	MaxOrderMatchingMs  int64   `json:"max_order_matching_ms"`
	MaxPriceCalcMs      int64   `json:"max_price_calc_ms"`
	MaxSigningMs        int64   `json:"max_signing_ms"`
	RateLimiterCheckNs  int64   `json:"rate_limiter_check_ns"`
}

// ZKProofBaselines defines ZK proof generation baselines
type ZKProofBaselines struct {
	AgeProofGenerationMs       int64 `json:"age_proof_generation_ms"`
	ResidencyProofGenerationMs int64 `json:"residency_proof_generation_ms"`
	ScoreThresholdProofMs      int64 `json:"score_threshold_proof_ms"`
	SelectiveDisclosureProofMs int64 `json:"selective_disclosure_proof_ms"`
	ProofVerificationMs        int64 `json:"proof_verification_ms"`
	CommitmentGenerationNs     int64 `json:"commitment_generation_ns"`
	NonceGenerationNs          int64 `json:"nonce_generation_ns"`
	ProofIDGenerationNs        int64 `json:"proof_id_generation_ns"`
	CircuitCompilationMs       int64 `json:"circuit_compilation_ms"`
}

// MemoryBaselines defines memory usage baselines
type MemoryBaselines struct {
	MaxHeapAllocMB       int64 `json:"max_heap_alloc_mb"`
	MaxHeapObjectsM      int64 `json:"max_heap_objects_millions"`
	MaxGoroutines        int64 `json:"max_goroutines"`
	MaxGCPauseMs         int64 `json:"max_gc_pause_ms"`
	MaxLeakGrowthPct     int64 `json:"max_leak_growth_pct"`
	IdleHeapMB           int64 `json:"idle_heap_mb"`
	PerVerificationKB    int64 `json:"per_verification_kb"`
	PerTransactionKB     int64 `json:"per_transaction_kb"`
}

// DefaultBaselineMetrics returns the default baseline metrics
func DefaultBaselineMetrics() BaselineMetrics {
	return BaselineMetrics{
		Version:     "1.0.0",
		GeneratedAt: time.Now().UTC(),
		Platform:    "go-benchmark",
		GoVersion:   "1.21+",

		Transaction: TransactionBaselines{
			TargetTPS:         10000,
			MinAcceptableTPS:  5000,
			MaxP95LatencyMs:   100,
			MaxP99LatencyMs:   250,
			MaxErrorRatePct:   0.1,
			HashComputeNs:     1000,     // 1µs
			ValidationNs:      10000,    // 10µs
			ExecutionNs:       50000,    // 50µs
			StateWriteNs:      5000,     // 5µs
		},

		Verification: VerificationBaselines{
			TargetLatencyMs:        100,
			MaxP95LatencyMs:        150,
			MaxP99LatencyMs:        250,
			MinThroughput:          10.0,
			MaxDecryptionMs:        30,
			MaxScoringMs:           50,
			MaxStateUpdateMs:       10,
			IdentityRecordCreateNs: 100000,  // 100µs
			IdentityRecordGetNs:    10000,   // 10µs
			ScoreUpdateNs:          50000,   // 50µs
		},

		Bidding: BiddingBaselines{
			TargetLatencyMs:    50,
			MaxP95LatencyMs:    100,
			MaxP99LatencyMs:    200,
			MinBidsPerSecond:   20.0,
			MaxOrderMatchingMs: 10,
			MaxPriceCalcMs:     5,
			MaxSigningMs:       20,
			RateLimiterCheckNs: 1000, // 1µs
		},

		ZKProof: ZKProofBaselines{
			AgeProofGenerationMs:       500,
			ResidencyProofGenerationMs: 500,
			ScoreThresholdProofMs:      500,
			SelectiveDisclosureProofMs: 1000,
			ProofVerificationMs:        100,
			CommitmentGenerationNs:     5000,   // 5µs
			NonceGenerationNs:          1000,   // 1µs
			ProofIDGenerationNs:        2000,   // 2µs
			CircuitCompilationMs:       5000,   // 5s (one-time)
		},

		Memory: MemoryBaselines{
			MaxHeapAllocMB:       4096,   // 4GB
			MaxHeapObjectsM:      10,     // 10M
			MaxGoroutines:        100000, // 100K
			MaxGCPauseMs:         100,
			MaxLeakGrowthPct:     10,
			IdleHeapMB:           256,
			PerVerificationKB:    64,
			PerTransactionKB:     8,
		},
	}
}

// BenchmarkResult represents a single benchmark result
type BenchmarkResult struct {
	Name           string        `json:"name"`
	Category       string        `json:"category"`
	Iterations     int64         `json:"iterations"`
	NsPerOp        int64         `json:"ns_per_op"`
	BytesPerOp     int64         `json:"bytes_per_op"`
	AllocsPerOp    int64         `json:"allocs_per_op"`
	Throughput     float64       `json:"throughput,omitempty"`
	P50LatencyNs   int64         `json:"p50_latency_ns,omitempty"`
	P95LatencyNs   int64         `json:"p95_latency_ns,omitempty"`
	P99LatencyNs   int64         `json:"p99_latency_ns,omitempty"`
	ErrorRate      float64       `json:"error_rate,omitempty"`
	Timestamp      time.Time     `json:"timestamp"`
}

// ComparisonResult represents the result of comparing against baseline
type ComparisonResult struct {
	Name           string  `json:"name"`
	Category       string  `json:"category"`
	Current        int64   `json:"current"`
	Baseline       int64   `json:"baseline"`
	DiffPercent    float64 `json:"diff_percent"`
	DiffAbs        int64   `json:"diff_abs"`
	Status         string  `json:"status"` // "pass", "warn", "fail"
	Threshold      float64 `json:"threshold"`
}

// BenchmarkComparator compares benchmark results against baselines
type BenchmarkComparator struct {
	baselines       BaselineMetrics
	warnThreshold   float64 // percentage above baseline to warn (e.g., 10.0 = 10%)
	failThreshold   float64 // percentage above baseline to fail (e.g., 20.0 = 20%)
}

// NewBenchmarkComparator creates a new comparator with given thresholds
func NewBenchmarkComparator(baselines BaselineMetrics, warnThreshold, failThreshold float64) *BenchmarkComparator {
	return &BenchmarkComparator{
		baselines:     baselines,
		warnThreshold: warnThreshold,
		failThreshold: failThreshold,
	}
}

// DefaultBenchmarkComparator creates a comparator with default thresholds
func DefaultBenchmarkComparator() *BenchmarkComparator {
	return NewBenchmarkComparator(DefaultBaselineMetrics(), 10.0, 20.0)
}

// Compare compares a benchmark result against its baseline
func (c *BenchmarkComparator) Compare(result BenchmarkResult, baselineNs int64) *ComparisonResult {
	diffAbs := result.NsPerOp - baselineNs
	diffPercent := 0.0
	if baselineNs > 0 {
		diffPercent = float64(diffAbs) / float64(baselineNs) * 100
	}

	status := resultPass
	if diffPercent > c.failThreshold {
		status = resultFail
	} else if diffPercent > c.warnThreshold {
		status = resultWarn
	}

	return &ComparisonResult{
		Name:        result.Name,
		Category:    result.Category,
		Current:     result.NsPerOp,
		Baseline:    baselineNs,
		DiffPercent: diffPercent,
		DiffAbs:     diffAbs,
		Status:      status,
		Threshold:   c.failThreshold,
	}
}

// CompareTransaction compares transaction benchmark results
func (c *BenchmarkComparator) CompareTransaction(name string, nsPerOp int64) *ComparisonResult {
	var baselineNs int64
	switch name {
	case "hash_compute":
		baselineNs = c.baselines.Transaction.HashComputeNs
	case "validation":
		baselineNs = c.baselines.Transaction.ValidationNs
	case "execution":
		baselineNs = c.baselines.Transaction.ExecutionNs
	case "state_write":
		baselineNs = c.baselines.Transaction.StateWriteNs
	default:
		baselineNs = 0
	}

	result := BenchmarkResult{Name: name, Category: "transaction", NsPerOp: nsPerOp}
	return c.Compare(result, baselineNs)
}

// CompareVerification compares verification benchmark results
func (c *BenchmarkComparator) CompareVerification(name string, nsPerOp int64) *ComparisonResult {
	var baselineNs int64
	switch name {
	case "identity_record_create":
		baselineNs = c.baselines.Verification.IdentityRecordCreateNs
	case "identity_record_get":
		baselineNs = c.baselines.Verification.IdentityRecordGetNs
	case "score_update":
		baselineNs = c.baselines.Verification.ScoreUpdateNs
	case "decryption":
		baselineNs = c.baselines.Verification.MaxDecryptionMs * 1_000_000
	case "scoring":
		baselineNs = c.baselines.Verification.MaxScoringMs * 1_000_000
	default:
		baselineNs = 0
	}

	result := BenchmarkResult{Name: name, Category: "verification", NsPerOp: nsPerOp}
	return c.Compare(result, baselineNs)
}

// CompareBidding compares bidding benchmark results
func (c *BenchmarkComparator) CompareBidding(name string, nsPerOp int64) *ComparisonResult {
	var baselineNs int64
	switch name {
	case "order_matching":
		baselineNs = c.baselines.Bidding.MaxOrderMatchingMs * 1_000_000
	case "price_calc":
		baselineNs = c.baselines.Bidding.MaxPriceCalcMs * 1_000_000
	case "signing":
		baselineNs = c.baselines.Bidding.MaxSigningMs * 1_000_000
	case "rate_limiter":
		baselineNs = c.baselines.Bidding.RateLimiterCheckNs
	default:
		baselineNs = 0
	}

	result := BenchmarkResult{Name: name, Category: "bidding", NsPerOp: nsPerOp}
	return c.Compare(result, baselineNs)
}

// CompareZKProof compares ZK proof benchmark results
func (c *BenchmarkComparator) CompareZKProof(name string, nsPerOp int64) *ComparisonResult {
	var baselineNs int64
	switch name {
	case "age_proof":
		baselineNs = c.baselines.ZKProof.AgeProofGenerationMs * 1_000_000
	case "residency_proof":
		baselineNs = c.baselines.ZKProof.ResidencyProofGenerationMs * 1_000_000
	case "score_threshold":
		baselineNs = c.baselines.ZKProof.ScoreThresholdProofMs * 1_000_000
	case "selective_disclosure":
		baselineNs = c.baselines.ZKProof.SelectiveDisclosureProofMs * 1_000_000
	case "proof_verification":
		baselineNs = c.baselines.ZKProof.ProofVerificationMs * 1_000_000
	case "commitment":
		baselineNs = c.baselines.ZKProof.CommitmentGenerationNs
	case "nonce":
		baselineNs = c.baselines.ZKProof.NonceGenerationNs
	case "proof_id":
		baselineNs = c.baselines.ZKProof.ProofIDGenerationNs
	default:
		baselineNs = 0
	}

	result := BenchmarkResult{Name: name, Category: "zk_proof", NsPerOp: nsPerOp}
	return c.Compare(result, baselineNs)
}

// SaveBaselines saves baselines to a JSON file
func SaveBaselines(baselines BaselineMetrics, filename string) error {
	data, err := json.MarshalIndent(baselines, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal baselines: %w", err)
	}

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// LoadBaselines loads baselines from a JSON file
func LoadBaselines(filename string) (*BaselineMetrics, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var baselines BaselineMetrics
	if err := json.Unmarshal(data, &baselines); err != nil {
		return nil, fmt.Errorf("failed to unmarshal baselines: %w", err)
	}

	return &baselines, nil
}

// BenchmarkReport represents a full benchmark report
type BenchmarkReport struct {
	GeneratedAt    time.Time          `json:"generated_at"`
	Platform       string             `json:"platform"`
	GoVersion      string             `json:"go_version"`
	BaselineVersion string            `json:"baseline_version"`
	Results        []BenchmarkResult  `json:"results"`
	Comparisons    []ComparisonResult `json:"comparisons"`
	Summary        ReportSummary      `json:"summary"`
}

// ReportSummary summarizes benchmark comparison results
type ReportSummary struct {
	TotalBenchmarks int `json:"total_benchmarks"`
	Passed          int `json:"passed"`
	Warnings        int `json:"warnings"`
	Failed          int `json:"failed"`
	OverallStatus   string `json:"overall_status"`
}

// GenerateReport generates a benchmark report from results
func (c *BenchmarkComparator) GenerateReport(results []BenchmarkResult) *BenchmarkReport {
	report := &BenchmarkReport{
		GeneratedAt:     time.Now().UTC(),
		Platform:        "go-benchmark",
		GoVersion:       "1.21+",
		BaselineVersion: c.baselines.Version,
		Results:         results,
		Comparisons:     make([]ComparisonResult, 0, len(results)),
	}

	summary := ReportSummary{
		TotalBenchmarks: len(results),
	}

	for _, r := range results {
		var comparison *ComparisonResult
		switch r.Category {
		case "transaction":
			comparison = c.CompareTransaction(r.Name, r.NsPerOp)
		case "verification":
			comparison = c.CompareVerification(r.Name, r.NsPerOp)
		case "bidding":
			comparison = c.CompareBidding(r.Name, r.NsPerOp)
		case "zk_proof":
			comparison = c.CompareZKProof(r.Name, r.NsPerOp)
		}

		if comparison != nil {
			report.Comparisons = append(report.Comparisons, *comparison)

			switch comparison.Status {
			case resultPass:
				summary.Passed++
			case resultWarn:
				summary.Warnings++
			case resultFail:
				summary.Failed++
			}
		}
	}

	// Determine overall status
	if summary.Failed > 0 {
		summary.OverallStatus = resultFail
	} else if summary.Warnings > 0 {
		summary.OverallStatus = resultWarn
	} else {
		summary.OverallStatus = resultPass
	}

	report.Summary = summary
	return report
}

// SaveReport saves a benchmark report to a JSON file
func SaveReport(report *BenchmarkReport, filename string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
