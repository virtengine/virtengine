// Package benchmark provides memory profiling utilities for VirtEngine.
// Task Reference: PERF-001 - Memory Profiling for Long-Running Processes
package benchmark

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

// MemorySnapshot represents a point-in-time memory snapshot
type MemorySnapshot struct {
	Timestamp      time.Time `json:"timestamp"`
	HeapAlloc      uint64    `json:"heap_alloc"`       // bytes allocated and still in use
	HeapSys        uint64    `json:"heap_sys"`         // bytes obtained from system
	HeapInuse      uint64    `json:"heap_inuse"`       // bytes in in-use spans
	HeapIdle       uint64    `json:"heap_idle"`        // bytes in idle spans
	HeapReleased   uint64    `json:"heap_released"`    // bytes released to OS
	HeapObjects    uint64    `json:"heap_objects"`     // number of allocated objects
	StackInuse     uint64    `json:"stack_inuse"`      // bytes in stack spans
	StackSys       uint64    `json:"stack_sys"`        // bytes obtained from system for stacks
	MSpanInuse     uint64    `json:"mspan_inuse"`      // bytes in mspan structures
	MCacheInuse    uint64    `json:"mcache_inuse"`     // bytes in mcache structures
	BuckHashSys    uint64    `json:"buckhash_sys"`     // bytes in profiling bucket hash table
	GCSys          uint64    `json:"gc_sys"`           // bytes in GC metadata
	OtherSys       uint64    `json:"other_sys"`        // other system allocations
	NextGC         uint64    `json:"next_gc"`          // target heap size for next GC
	LastGC         uint64    `json:"last_gc"`          // time of last GC (nanoseconds since epoch)
	PauseTotalNs   uint64    `json:"pause_total_ns"`   // total GC pause time
	NumGC          uint32    `json:"num_gc"`           // number of completed GC cycles
	NumForcedGC    uint32    `json:"num_forced_gc"`    // number of forced GC cycles
	GCCPUFraction  float64   `json:"gc_cpu_fraction"`  // fraction of CPU time used by GC
	Goroutines     int       `json:"goroutines"`       // number of goroutines
}

// MemoryProfiler tracks memory usage over time
type MemoryProfiler struct {
	mu        sync.RWMutex
	snapshots []MemorySnapshot
	interval  time.Duration
	maxSize   int
	running   bool
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// MemoryProfilerConfig configures the memory profiler
type MemoryProfilerConfig struct {
	Interval time.Duration // Sampling interval
	MaxSize  int           // Maximum number of snapshots to keep
}

// DefaultMemoryProfilerConfig returns default configuration
func DefaultMemoryProfilerConfig() MemoryProfilerConfig {
	return MemoryProfilerConfig{
		Interval: 5 * time.Second,
		MaxSize:  1000,
	}
}

// NewMemoryProfiler creates a new memory profiler
func NewMemoryProfiler(config MemoryProfilerConfig) *MemoryProfiler {
	if config.Interval == 0 {
		config.Interval = 5 * time.Second
	}
	if config.MaxSize == 0 {
		config.MaxSize = 1000
	}

	return &MemoryProfiler{
		snapshots: make([]MemorySnapshot, 0, config.MaxSize),
		interval:  config.Interval,
		maxSize:   config.MaxSize,
	}
}

// Start begins memory profiling
func (p *MemoryProfiler) Start(ctx context.Context) error {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return fmt.Errorf("profiler already running")
	}

	p.ctx, p.cancel = context.WithCancel(ctx)
	p.running = true
	p.mu.Unlock()

	p.wg.Add(1)
	go p.collectLoop()

	return nil
}

// Stop stops memory profiling
func (p *MemoryProfiler) Stop() {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return
	}
	p.running = false
	p.mu.Unlock()

	if p.cancel != nil {
		p.cancel()
	}
	p.wg.Wait()
}

func (p *MemoryProfiler) collectLoop() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	// Take initial snapshot
	p.takeSnapshot()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.takeSnapshot()
		}
	}
}

// TakeSnapshot takes a manual memory snapshot
func (p *MemoryProfiler) TakeSnapshot() MemorySnapshot {
	return p.takeSnapshot()
}

func (p *MemoryProfiler) takeSnapshot() MemorySnapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	snapshot := MemorySnapshot{
		Timestamp:     time.Now().UTC(),
		HeapAlloc:     m.HeapAlloc,
		HeapSys:       m.HeapSys,
		HeapInuse:     m.HeapInuse,
		HeapIdle:      m.HeapIdle,
		HeapReleased:  m.HeapReleased,
		HeapObjects:   m.HeapObjects,
		StackInuse:    m.StackInuse,
		StackSys:      m.StackSys,
		MSpanInuse:    m.MSpanInuse,
		MCacheInuse:   m.MCacheInuse,
		BuckHashSys:   m.BuckHashSys,
		GCSys:         m.GCSys,
		OtherSys:      m.OtherSys,
		NextGC:        m.NextGC,
		LastGC:        m.LastGC,
		PauseTotalNs:  m.PauseTotalNs,
		NumGC:         m.NumGC,
		NumForcedGC:   m.NumForcedGC,
		GCCPUFraction: m.GCCPUFraction,
		Goroutines:    runtime.NumGoroutine(),
	}

	p.mu.Lock()
	p.snapshots = append(p.snapshots, snapshot)

	// Trim if exceeds max size
	if len(p.snapshots) > p.maxSize {
		p.snapshots = p.snapshots[len(p.snapshots)-p.maxSize:]
	}
	p.mu.Unlock()

	return snapshot
}

// GetSnapshots returns all collected snapshots
func (p *MemoryProfiler) GetSnapshots() []MemorySnapshot {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]MemorySnapshot, len(p.snapshots))
	copy(result, p.snapshots)
	return result
}

// GetLatestSnapshot returns the most recent snapshot
func (p *MemoryProfiler) GetLatestSnapshot() *MemorySnapshot {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.snapshots) == 0 {
		return nil
	}
	s := p.snapshots[len(p.snapshots)-1]
	return &s
}

// Clear removes all snapshots
func (p *MemoryProfiler) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.snapshots = make([]MemorySnapshot, 0, p.maxSize)
}

// MemoryStats provides analysis of memory snapshots
type MemoryStats struct {
	MinHeapAlloc    uint64        `json:"min_heap_alloc"`
	MaxHeapAlloc    uint64        `json:"max_heap_alloc"`
	AvgHeapAlloc    uint64        `json:"avg_heap_alloc"`
	MinHeapObjects  uint64        `json:"min_heap_objects"`
	MaxHeapObjects  uint64        `json:"max_heap_objects"`
	AvgHeapObjects  uint64        `json:"avg_heap_objects"`
	MinGoroutines   int           `json:"min_goroutines"`
	MaxGoroutines   int           `json:"max_goroutines"`
	AvgGoroutines   int           `json:"avg_goroutines"`
	TotalGCPauses   uint64        `json:"total_gc_pauses_ns"`
	AvgGCPause      time.Duration `json:"avg_gc_pause"`
	GCCycles        uint32        `json:"gc_cycles"`
	SampleCount     int           `json:"sample_count"`
	Duration        time.Duration `json:"duration"`
}

// GetStats returns statistics for collected snapshots
func (p *MemoryProfiler) GetStats() *MemoryStats {
	snapshots := p.GetSnapshots()
	if len(snapshots) == 0 {
		return nil
	}

	stats := &MemoryStats{
		MinHeapAlloc:   ^uint64(0),
		MinHeapObjects: ^uint64(0),
		MinGoroutines:  int(^uint(0) >> 1),
		SampleCount:    len(snapshots),
	}

	var totalHeapAlloc, totalHeapObjects uint64
	var totalGoroutines int64

	for _, s := range snapshots {
		if s.HeapAlloc < stats.MinHeapAlloc {
			stats.MinHeapAlloc = s.HeapAlloc
		}
		if s.HeapAlloc > stats.MaxHeapAlloc {
			stats.MaxHeapAlloc = s.HeapAlloc
		}
		totalHeapAlloc += s.HeapAlloc

		if s.HeapObjects < stats.MinHeapObjects {
			stats.MinHeapObjects = s.HeapObjects
		}
		if s.HeapObjects > stats.MaxHeapObjects {
			stats.MaxHeapObjects = s.HeapObjects
		}
		totalHeapObjects += s.HeapObjects

		if s.Goroutines < stats.MinGoroutines {
			stats.MinGoroutines = s.Goroutines
		}
		if s.Goroutines > stats.MaxGoroutines {
			stats.MaxGoroutines = s.Goroutines
		}
		totalGoroutines += int64(s.Goroutines)
	}

	stats.AvgHeapAlloc = totalHeapAlloc / uint64(len(snapshots))
	stats.AvgHeapObjects = totalHeapObjects / uint64(len(snapshots))
	stats.AvgGoroutines = int(totalGoroutines / int64(len(snapshots)))

	// GC stats from last snapshot
	last := snapshots[len(snapshots)-1]
	first := snapshots[0]
	stats.TotalGCPauses = last.PauseTotalNs - first.PauseTotalNs
	stats.GCCycles = last.NumGC - first.NumGC
	if stats.GCCycles > 0 {
		//nolint:gosec // G115: GCCycles is positive uint32
		stats.AvgGCPause = time.Duration(stats.TotalGCPauses / uint64(stats.GCCycles))
	}
	stats.Duration = last.Timestamp.Sub(first.Timestamp)

	return stats
}

// ExportJSON exports snapshots to JSON file
func (p *MemoryProfiler) ExportJSON(filename string) error {
	snapshots := p.GetSnapshots()

	data, err := json.MarshalIndent(snapshots, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshots: %w", err)
	}

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	//nolint:gosec // G306: profile export file, 0644 permissions acceptable
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// MemoryLeak represents a potential memory leak detection
type MemoryLeak struct {
	StartHeap       uint64        `json:"start_heap"`
	EndHeap         uint64        `json:"end_heap"`
	Growth          uint64        `json:"growth"`
	GrowthPercent   float64       `json:"growth_percent"`
	Duration        time.Duration `json:"duration"`
	GrowthRate      uint64        `json:"growth_rate_bytes_per_sec"`
	SuspectedLeak   bool          `json:"suspected_leak"`
	ObjectGrowth    int64         `json:"object_growth"`
	GoroutineGrowth int           `json:"goroutine_growth"`
}

// DetectLeaks analyzes snapshots for potential memory leaks
func (p *MemoryProfiler) DetectLeaks(threshold float64) *MemoryLeak {
	snapshots := p.GetSnapshots()
	if len(snapshots) < 2 {
		return nil
	}

	first := snapshots[0]
	last := snapshots[len(snapshots)-1]
	duration := last.Timestamp.Sub(first.Timestamp)

	if duration < time.Second {
		return nil
	}

	leak := &MemoryLeak{
		StartHeap:       first.HeapAlloc,
		EndHeap:         last.HeapAlloc,
		Duration:        duration,
		//nolint:gosec // G115: HeapObjects fits in int64
		ObjectGrowth:    int64(last.HeapObjects) - int64(first.HeapObjects),
		GoroutineGrowth: last.Goroutines - first.Goroutines,
	}

	if last.HeapAlloc > first.HeapAlloc {
		leak.Growth = last.HeapAlloc - first.HeapAlloc
		leak.GrowthPercent = float64(leak.Growth) / float64(first.HeapAlloc) * 100
		leak.GrowthRate = leak.Growth / uint64(duration.Seconds())
	}

	// Suspect leak if growth exceeds threshold percentage
	leak.SuspectedLeak = leak.GrowthPercent > threshold

	return leak
}

// ForceGC forces garbage collection and returns memory stats before and after
func ForceGC() (before, after MemorySnapshot) {
	var m runtime.MemStats

	// Before
	runtime.ReadMemStats(&m)
	before = MemorySnapshot{
		Timestamp:    time.Now().UTC(),
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
		HeapInuse:    m.HeapInuse,
		HeapObjects:  m.HeapObjects,
		NumGC:        m.NumGC,
		Goroutines:   runtime.NumGoroutine(),
	}

	// Force GC
	runtime.GC()
	debug.FreeOSMemory()

	// After
	runtime.ReadMemStats(&m)
	after = MemorySnapshot{
		Timestamp:    time.Now().UTC(),
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
		HeapInuse:    m.HeapInuse,
		HeapObjects:  m.HeapObjects,
		NumGC:        m.NumGC,
		Goroutines:   runtime.NumGoroutine(),
	}

	return
}

// MemoryBudget defines memory constraints
type MemoryBudget struct {
	MaxHeapAlloc   uint64 `json:"max_heap_alloc"`
	MaxHeapObjects uint64 `json:"max_heap_objects"`
	MaxGoroutines  int    `json:"max_goroutines"`
	MaxGCPauseNs   uint64 `json:"max_gc_pause_ns"`
}

// DefaultMemoryBudget returns default memory budget
func DefaultMemoryBudget() MemoryBudget {
	return MemoryBudget{
		MaxHeapAlloc:   4 * 1024 * 1024 * 1024, // 4GB
		MaxHeapObjects: 10_000_000,              // 10M objects
		MaxGoroutines:  100_000,                 // 100K goroutines
		MaxGCPauseNs:   100_000_000,             // 100ms
	}
}

// BudgetViolation represents a memory budget violation
type BudgetViolation struct {
	Type      string `json:"type"`
	Current   uint64 `json:"current"`
	Limit     uint64 `json:"limit"`
	Exceeded  bool   `json:"exceeded"`
}

// CheckBudget checks memory against budget constraints
func (p *MemoryProfiler) CheckBudget(budget MemoryBudget) []BudgetViolation {
	latest := p.GetLatestSnapshot()
	if latest == nil {
		return nil
	}

	var violations []BudgetViolation

	// Check heap allocation
	if latest.HeapAlloc > budget.MaxHeapAlloc {
		violations = append(violations, BudgetViolation{
			Type:     "heap_alloc",
			Current:  latest.HeapAlloc,
			Limit:    budget.MaxHeapAlloc,
			Exceeded: true,
		})
	}

	// Check heap objects
	if latest.HeapObjects > budget.MaxHeapObjects {
		violations = append(violations, BudgetViolation{
			Type:     "heap_objects",
			Current:  latest.HeapObjects,
			Limit:    budget.MaxHeapObjects,
			Exceeded: true,
		})
	}

	// Check goroutines
	if latest.Goroutines > budget.MaxGoroutines {
		violations = append(violations, BudgetViolation{
			Type: "goroutines",
			//nolint:gosec // G115: Goroutine count is positive int
			Current: uint64(latest.Goroutines),
			//nolint:gosec // G115: MaxGoroutines is positive int
			Limit:    uint64(budget.MaxGoroutines),
			Exceeded: true,
		})
	}

	return violations
}

// FormatBytes formats bytes as human-readable string
func FormatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

