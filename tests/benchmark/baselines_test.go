package benchmark

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDefaultBaselineMetrics(t *testing.T) {
	baselines := DefaultBaselineMetrics()

	// Verify transaction baselines
	require.Equal(t, int64(10000), baselines.Transaction.TargetTPS)
	require.Equal(t, int64(5000), baselines.Transaction.MinAcceptableTPS)
	require.Equal(t, int64(100), baselines.Transaction.MaxP95LatencyMs)

	// Verify verification baselines
	require.Equal(t, int64(100), baselines.Verification.TargetLatencyMs)
	require.Equal(t, float64(10.0), baselines.Verification.MinThroughput)

	// Verify bidding baselines
	require.Equal(t, int64(50), baselines.Bidding.TargetLatencyMs)
	require.Equal(t, float64(20.0), baselines.Bidding.MinBidsPerSecond)

	// Verify memory baselines
	require.Equal(t, int64(4096), baselines.Memory.MaxHeapAllocMB)
	require.Equal(t, int64(100000), baselines.Memory.MaxGoroutines)
}

func TestBenchmarkComparator(t *testing.T) {
	comparator := DefaultBenchmarkComparator()
	require.NotNil(t, comparator)

	// Test passing result
	result := comparator.CompareTransaction("hash_compute", 500) // Below 1000ns baseline
	require.Equal(t, "pass", result.Status)
	require.Less(t, result.DiffPercent, 0.0) // Should be negative (faster than baseline)

	// Test warning result (must exceed 10% threshold)
	result = comparator.CompareTransaction("hash_compute", 1150) // 15% above baseline
	require.Equal(t, "warn", result.Status)

	// Test failing result (must exceed 20% threshold)
	result = comparator.CompareTransaction("hash_compute", 1250) // 25% above baseline
	require.Equal(t, "fail", result.Status)
}

func TestCompareTransaction(t *testing.T) {
	comparator := DefaultBenchmarkComparator()

	tests := []struct {
		name     string
		nsPerOp  int64
		expected string
	}{
		{"hash_compute", 500, "pass"},  // 50% faster
		{"hash_compute", 1000, "pass"}, // exactly at baseline
		{"hash_compute", 1150, "warn"}, // 15% slower (exceeds 10% warn)
		{"hash_compute", 1250, "fail"}, // 25% slower (exceeds 20% fail)
		{"validation", 5000, "pass"},   // 50% faster
		{"execution", 50000, "pass"},   // at baseline
		{"state_write", 5000, "pass"},  // at baseline
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := comparator.CompareTransaction(tt.name, tt.nsPerOp)
			require.Equal(t, tt.expected, result.Status)
		})
	}
}

func TestCompareVerification(t *testing.T) {
	comparator := DefaultBenchmarkComparator()

	tests := []struct {
		name     string
		nsPerOp  int64
		expected string
	}{
		{"identity_record_create", 50000, "pass"}, // 50% faster
		{"identity_record_get", 5000, "pass"},     // 50% faster
		{"score_update", 50000, "pass"},           // at baseline
		{"decryption", 20_000_000, "pass"},        // 33% faster than 30ms
		{"scoring", 40_000_000, "pass"},           // 20% faster than 50ms
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := comparator.CompareVerification(tt.name, tt.nsPerOp)
			require.Equal(t, tt.expected, result.Status)
		})
	}
}

func TestCompareBidding(t *testing.T) {
	comparator := DefaultBenchmarkComparator()

	tests := []struct {
		name     string
		nsPerOp  int64
		expected string
	}{
		{"order_matching", 5_000_000, "pass"}, // 50% faster than 10ms
		{"price_calc", 2_500_000, "pass"},     // 50% faster than 5ms
		{"signing", 10_000_000, "pass"},       // 50% faster than 20ms
		{"rate_limiter", 500, "pass"},         // 50% faster than 1µs
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := comparator.CompareBidding(tt.name, tt.nsPerOp)
			require.Equal(t, tt.expected, result.Status)
		})
	}
}

func TestCompareZKProof(t *testing.T) {
	comparator := DefaultBenchmarkComparator()

	tests := []struct {
		name     string
		nsPerOp  int64
		expected string
	}{
		{"age_proof", 250_000_000, "pass"},         // 50% faster than 500ms
		{"residency_proof", 300_000_000, "pass"},   // 40% faster than 500ms
		{"proof_verification", 50_000_000, "pass"}, // 50% faster than 100ms
		{"commitment", 2500, "pass"},               // 50% faster than 5µs
		{"nonce", 500, "pass"},                     // 50% faster than 1µs
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := comparator.CompareZKProof(tt.name, tt.nsPerOp)
			require.Equal(t, tt.expected, result.Status)
		})
	}
}

func TestSaveAndLoadBaselines(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "baselines.json")

	// Save baselines
	baselines := DefaultBaselineMetrics()
	err := SaveBaselines(baselines, filename)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filename)
	require.NoError(t, err)

	// Load baselines
	loaded, err := LoadBaselines(filename)
	require.NoError(t, err)
	require.NotNil(t, loaded)

	// Compare
	require.Equal(t, baselines.Transaction.TargetTPS, loaded.Transaction.TargetTPS)
	require.Equal(t, baselines.Verification.TargetLatencyMs, loaded.Verification.TargetLatencyMs)
	require.Equal(t, baselines.Bidding.TargetLatencyMs, loaded.Bidding.TargetLatencyMs)
	require.Equal(t, baselines.Memory.MaxHeapAllocMB, loaded.Memory.MaxHeapAllocMB)
}

func TestLoadBaselinesFileNotFound(t *testing.T) {
	_, err := LoadBaselines("/nonexistent/file.json")
	require.Error(t, err)
}

func TestGenerateReport(t *testing.T) {
	comparator := DefaultBenchmarkComparator()

	results := []BenchmarkResult{
		{Name: "hash_compute", Category: "transaction", NsPerOp: 500, Timestamp: time.Now()},
		{Name: "validation", Category: "transaction", NsPerOp: 5000, Timestamp: time.Now()},
		{Name: "identity_record_create", Category: "verification", NsPerOp: 50000, Timestamp: time.Now()},
		{Name: "order_matching", Category: "bidding", NsPerOp: 5_000_000, Timestamp: time.Now()},
		{Name: "commitment", Category: "zk_proof", NsPerOp: 2500, Timestamp: time.Now()},
	}

	report := comparator.GenerateReport(results)
	require.NotNil(t, report)
	require.Equal(t, len(results), report.Summary.TotalBenchmarks)
	require.Equal(t, "pass", report.Summary.OverallStatus)
	require.Equal(t, len(results), len(report.Comparisons))
}

func TestGenerateReportWithFailures(t *testing.T) {
	comparator := DefaultBenchmarkComparator()

	results := []BenchmarkResult{
		{Name: "hash_compute", Category: "transaction", NsPerOp: 500, Timestamp: time.Now()},
		{Name: "hash_compute", Category: "transaction", NsPerOp: 2000, Timestamp: time.Now()}, // This should fail (2x baseline)
	}

	report := comparator.GenerateReport(results)
	require.NotNil(t, report)
	require.Equal(t, "fail", report.Summary.OverallStatus)
	require.Greater(t, report.Summary.Failed, 0)
}

func TestGenerateReportWithWarnings(t *testing.T) {
	comparator := DefaultBenchmarkComparator()

	results := []BenchmarkResult{
		{Name: "hash_compute", Category: "transaction", NsPerOp: 500, Timestamp: time.Now()},
		{Name: "hash_compute", Category: "transaction", NsPerOp: 1150, Timestamp: time.Now()}, // 15% over, should warn
	}

	report := comparator.GenerateReport(results)
	require.NotNil(t, report)
	require.Equal(t, "warn", report.Summary.OverallStatus)
	require.Greater(t, report.Summary.Warnings, 0)
}

func TestSaveReport(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "report.json")

	comparator := DefaultBenchmarkComparator()
	results := []BenchmarkResult{
		{Name: "hash_compute", Category: "transaction", NsPerOp: 500, Timestamp: time.Now()},
	}

	report := comparator.GenerateReport(results)
	err := SaveReport(report, filename)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filename)
	require.NoError(t, err)
}

func TestComparisonResult(t *testing.T) {
	comparator := NewBenchmarkComparator(DefaultBaselineMetrics(), 10.0, 25.0)

	// Test result structure
	result := BenchmarkResult{
		Name:      "test_benchmark",
		Category:  "transaction",
		NsPerOp:   1000,
		Timestamp: time.Now(),
	}

	comparison := comparator.Compare(result, 1000)
	require.Equal(t, "test_benchmark", comparison.Name)
	require.Equal(t, "transaction", comparison.Category)
	require.Equal(t, int64(1000), comparison.Current)
	require.Equal(t, int64(1000), comparison.Baseline)
	require.Equal(t, float64(0), comparison.DiffPercent)
	require.Equal(t, int64(0), comparison.DiffAbs)
	require.Equal(t, "pass", comparison.Status)
}

// BenchmarkComparatorCompare benchmarks the comparison operation
func BenchmarkComparatorCompare(b *testing.B) {
	comparator := DefaultBenchmarkComparator()
	result := BenchmarkResult{
		Name:      "test",
		Category:  "transaction",
		NsPerOp:   1000,
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = comparator.Compare(result, 1000)
	}
}

// BenchmarkGenerateReport benchmarks report generation
func BenchmarkGenerateReport(b *testing.B) {
	comparator := DefaultBenchmarkComparator()
	results := make([]BenchmarkResult, 100)
	for i := range results {
		results[i] = BenchmarkResult{
			Name:      "benchmark",
			Category:  "transaction",
			NsPerOp:   int64(1000 + i),
			Timestamp: time.Now(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = comparator.GenerateReport(results)
	}
}
