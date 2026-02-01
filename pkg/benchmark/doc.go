// Package benchmark provides performance benchmarking and memory profiling utilities
// for VirtEngine. It includes tools for measuring transaction throughput, VEID
// verification latency, provider daemon bidding performance, and detecting memory
// leaks in long-running processes.
//
// # Overview
//
// This package is designed to support PERF-001 requirements:
//   - Transaction throughput benchmarks (target: 10k TPS)
//   - VEID verification latency benchmarks (target: <100ms)
//   - Provider daemon bidding latency benchmarks
//   - Memory profiling for long-running processes
//   - Baseline metrics for regression detection
//
// # Memory Profiling
//
// The MemoryProfiler collects memory snapshots at configurable intervals:
//
//	config := DefaultMemoryProfilerConfig()
//	profiler := NewMemoryProfiler(config)
//	profiler.Start(ctx)
//	defer profiler.Stop()
//
//	// Run your workload...
//
//	stats := profiler.GetStats()
//	leak := profiler.DetectLeaks(10.0) // 10% threshold
//
// # Baseline Comparison
//
// Use BenchmarkComparator to compare results against established baselines:
//
//	comparator := DefaultBenchmarkComparator()
//	result := comparator.CompareTransaction("hash_compute", nsPerOp)
//	if result.Status == "fail" {
//	    // Performance regression detected
//	}
//
// # Running Benchmarks
//
// Run all performance benchmarks:
//
//	go test -bench=. -benchmem ./tests/benchmark/...
//	go test -bench=. -benchmem ./x/veid/keeper/...
//	go test -bench=. -benchmem ./pkg/provider_daemon/...
//
// Run specific benchmark categories:
//
//	go test -bench=BenchmarkTransaction ./tests/benchmark/...
//	go test -bench=BenchmarkVerification ./x/veid/keeper/...
//	go test -bench=BenchmarkBidding ./pkg/provider_daemon/...
//	go test -bench=BenchmarkZK ./x/veid/keeper/...
//
// # CI Integration
//
// For CI pipelines, use the baseline comparison utilities to detect regressions:
//
//	baselines, _ := LoadBaselines("BASELINE_METRICS.json")
//	comparator := NewBenchmarkComparator(*baselines, 10.0, 20.0)
//	report := comparator.GenerateReport(results)
//	if report.Summary.OverallStatus == "fail" {
//	    os.Exit(1)
//	}
package benchmark

