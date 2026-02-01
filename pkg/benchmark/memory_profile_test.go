package benchmark

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMemoryProfiler(t *testing.T) {
	config := MemoryProfilerConfig{
		Interval: 100 * time.Millisecond,
		MaxSize:  100,
	}

	profiler := NewMemoryProfiler(config)
	require.NotNil(t, profiler)

	// Start profiler
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := profiler.Start(ctx)
	require.NoError(t, err)

	// Cannot start twice
	err = profiler.Start(ctx)
	require.Error(t, err)

	// Wait for some snapshots
	time.Sleep(500 * time.Millisecond)

	// Check snapshots collected
	snapshots := profiler.GetSnapshots()
	require.NotEmpty(t, snapshots)
	require.Greater(t, len(snapshots), 1)

	// Check latest snapshot
	latest := profiler.GetLatestSnapshot()
	require.NotNil(t, latest)
	require.Greater(t, latest.HeapAlloc, uint64(0))
	require.Greater(t, latest.Goroutines, 0)

	// Stop profiler
	profiler.Stop()

	// Check stats
	stats := profiler.GetStats()
	require.NotNil(t, stats)
	require.Greater(t, stats.SampleCount, 1)
}

func TestMemorySnapshot(t *testing.T) {
	config := DefaultMemoryProfilerConfig()
	profiler := NewMemoryProfiler(config)

	// Take manual snapshot
	snapshot := profiler.TakeSnapshot()

	require.False(t, snapshot.Timestamp.IsZero())
	require.Greater(t, snapshot.HeapAlloc, uint64(0))
	require.Greater(t, snapshot.HeapSys, uint64(0))
	require.Greater(t, snapshot.Goroutines, 0)
}

func TestForceGC(t *testing.T) {
	// Allocate some memory
	data := make([]byte, 10*1024*1024) // 10MB
	for i := range data {
		data[i] = byte(i)
	}

	before, after := ForceGC()

	require.Greater(t, before.NumGC, uint32(0))
	require.GreaterOrEqual(t, after.NumGC, before.NumGC)

	// Clear reference
	data = nil
	_ = data

	// Force another GC
	_, afterClear := ForceGC()
	require.GreaterOrEqual(t, afterClear.NumGC, after.NumGC)
}

func TestMemoryStats(t *testing.T) {
	config := MemoryProfilerConfig{
		Interval: 50 * time.Millisecond,
		MaxSize:  100,
	}

	profiler := NewMemoryProfiler(config)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_ = profiler.Start(ctx)
	defer profiler.Stop()

	// Wait for snapshots
	time.Sleep(300 * time.Millisecond)

	stats := profiler.GetStats()
	require.NotNil(t, stats)
	require.GreaterOrEqual(t, stats.MaxHeapAlloc, stats.MinHeapAlloc)
	require.GreaterOrEqual(t, stats.MaxHeapObjects, stats.MinHeapObjects)
	require.Greater(t, stats.AvgHeapAlloc, uint64(0))
}

func TestMemoryBudget(t *testing.T) {
	config := DefaultMemoryProfilerConfig()
	profiler := NewMemoryProfiler(config)

	// Take snapshot
	profiler.TakeSnapshot()

	// Check with default budget
	budget := DefaultMemoryBudget()
	violations := profiler.CheckBudget(budget)
	require.Empty(t, violations, "Should not violate default budget")

	// Check with very restrictive budget
	strictBudget := MemoryBudget{
		MaxHeapAlloc:   1,        // 1 byte
		MaxHeapObjects: 1,        // 1 object
		MaxGoroutines:  1,        // 1 goroutine
		MaxGCPauseNs:   1,        // 1 nanosecond
	}

	violations = profiler.CheckBudget(strictBudget)
	require.NotEmpty(t, violations, "Should violate strict budget")
}

func TestDetectLeaks(t *testing.T) {
	config := MemoryProfilerConfig{
		Interval: 50 * time.Millisecond,
		MaxSize:  100,
	}

	profiler := NewMemoryProfiler(config)

	// Initial snapshot
	profiler.TakeSnapshot()

	// Allocate some memory
	var leaky [][]byte
	for i := 0; i < 100; i++ {
		leaky = append(leaky, make([]byte, 100*1024)) // 100KB each
		time.Sleep(10 * time.Millisecond)
		profiler.TakeSnapshot()
	}

	// Check for leaks
	leak := profiler.DetectLeaks(10.0) // 10% threshold
	require.NotNil(t, leak)
	require.Greater(t, leak.Growth, uint64(0))

	// Clear reference
	leaky = nil
	_ = leaky
	runtime.GC()
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    uint64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1048576, "1.00 MB"},
		{1572864, "1.50 MB"},
		{1073741824, "1.00 GB"},
		{1610612736, "1.50 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestProfilerClear(t *testing.T) {
	config := DefaultMemoryProfilerConfig()
	profiler := NewMemoryProfiler(config)

	// Take some snapshots
	for i := 0; i < 5; i++ {
		profiler.TakeSnapshot()
	}

	require.Len(t, profiler.GetSnapshots(), 5)

	// Clear
	profiler.Clear()
	require.Empty(t, profiler.GetSnapshots())
}

func TestProfilerMaxSize(t *testing.T) {
	config := MemoryProfilerConfig{
		Interval: time.Millisecond,
		MaxSize:  5,
	}

	profiler := NewMemoryProfiler(config)

	// Take more snapshots than max size
	for i := 0; i < 10; i++ {
		profiler.TakeSnapshot()
	}

	// Should only keep maxSize snapshots
	snapshots := profiler.GetSnapshots()
	require.LessOrEqual(t, len(snapshots), config.MaxSize)
}

// BenchmarkTakeSnapshot benchmarks snapshot collection
func BenchmarkTakeSnapshot(b *testing.B) {
	profiler := NewMemoryProfiler(DefaultMemoryProfilerConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		profiler.TakeSnapshot()
	}
}

// BenchmarkForceGC benchmarks forced garbage collection
func BenchmarkForceGC(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ForceGC()
	}
}

// BenchmarkGetStats benchmarks stats calculation
func BenchmarkGetStats(b *testing.B) {
	profiler := NewMemoryProfiler(DefaultMemoryProfilerConfig())

	// Pre-populate snapshots
	for i := 0; i < 100; i++ {
		profiler.TakeSnapshot()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = profiler.GetStats()
	}
}

