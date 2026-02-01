package cache

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics provides cache performance metrics and monitoring.
type Metrics struct {
	// Per-cache statistics
	cachesMu sync.RWMutex
	caches   map[string]MetricsProvider

	// Global counters
	totalHits        atomic.Uint64
	totalMisses      atomic.Uint64
	totalEvictions   atomic.Uint64
	totalExpirations atomic.Uint64

	// Configuration
	namespace      string
	reportInterval time.Duration

	// Reporting
	stopCh    chan struct{}
	wg        sync.WaitGroup
	reporters []MetricsReporter
}

// MetricsProvider is implemented by caches that provide statistics.
type MetricsProvider interface {
	Stats() CacheStats
}

// MetricsReporter is a callback for reporting metrics periodically.
type MetricsReporter func(report MetricsReport)

// MetricsReport contains aggregated metrics for all registered caches.
type MetricsReport struct {
	// Timestamp when the report was generated.
	Timestamp time.Time `json:"timestamp"`

	// Caches contains per-cache statistics.
	Caches map[string]CacheStats `json:"caches"`

	// Totals contains aggregate statistics.
	Totals AggregateCacheStats `json:"totals"`
}

// AggregateCacheStats contains aggregate statistics across all caches.
type AggregateCacheStats struct {
	TotalHits        uint64  `json:"total_hits"`
	TotalMisses      uint64  `json:"total_misses"`
	TotalEvictions   uint64  `json:"total_evictions"`
	TotalExpirations uint64  `json:"total_expirations"`
	TotalSize        int     `json:"total_size"`
	TotalMaxSize     int     `json:"total_max_size"`
	OverallHitRate   float64 `json:"overall_hit_rate"`
}

// NewMetrics creates a new Metrics instance.
func NewMetrics(namespace string, reportInterval time.Duration) *Metrics {
	if reportInterval <= 0 {
		reportInterval = 1 * time.Minute
	}

	m := &Metrics{
		caches:         make(map[string]MetricsProvider),
		namespace:      namespace,
		reportInterval: reportInterval,
		stopCh:         make(chan struct{}),
	}

	return m
}

// RegisterCache registers a cache for metrics collection.
func (m *Metrics) RegisterCache(name string, cache MetricsProvider) {
	m.cachesMu.Lock()
	defer m.cachesMu.Unlock()
	m.caches[name] = cache
}

// UnregisterCache removes a cache from metrics collection.
func (m *Metrics) UnregisterCache(name string) {
	m.cachesMu.Lock()
	defer m.cachesMu.Unlock()
	delete(m.caches, name)
}

// AddReporter adds a metrics reporter callback.
func (m *Metrics) AddReporter(reporter MetricsReporter) {
	m.cachesMu.Lock()
	defer m.cachesMu.Unlock()
	m.reporters = append(m.reporters, reporter)
}

// Start begins periodic metrics reporting.
func (m *Metrics) Start() {
	m.wg.Add(1)
	go m.reportLoop()
}

// Stop stops periodic metrics reporting.
func (m *Metrics) Stop() {
	close(m.stopCh)
	m.wg.Wait()
}

// GetReport generates a current metrics report.
func (m *Metrics) GetReport() MetricsReport {
	m.cachesMu.RLock()
	defer m.cachesMu.RUnlock()

	report := MetricsReport{
		Timestamp: time.Now(),
		Caches:    make(map[string]CacheStats),
	}

	var (
		totalHits        uint64
		totalMisses      uint64
		totalEvictions   uint64
		totalExpirations uint64
		totalSize        int
		totalMaxSize     int
	)

	for name, cache := range m.caches {
		stats := cache.Stats()
		report.Caches[name] = stats

		totalHits += stats.Hits
		totalMisses += stats.Misses
		totalEvictions += stats.Evictions
		totalExpirations += stats.Expirations
		totalSize += stats.Size
		totalMaxSize += stats.MaxSize
	}

	overallHitRate := float64(0)
	if total := totalHits + totalMisses; total > 0 {
		overallHitRate = float64(totalHits) / float64(total) * 100
	}

	report.Totals = AggregateCacheStats{
		TotalHits:        totalHits,
		TotalMisses:      totalMisses,
		TotalEvictions:   totalEvictions,
		TotalExpirations: totalExpirations,
		TotalSize:        totalSize,
		TotalMaxSize:     totalMaxSize,
		OverallHitRate:   overallHitRate,
	}

	// Update global counters
	m.totalHits.Store(totalHits)
	m.totalMisses.Store(totalMisses)
	m.totalEvictions.Store(totalEvictions)
	m.totalExpirations.Store(totalExpirations)

	return report
}

// GetCacheStats returns statistics for a specific cache.
func (m *Metrics) GetCacheStats(name string) (CacheStats, bool) {
	m.cachesMu.RLock()
	defer m.cachesMu.RUnlock()

	cache, ok := m.caches[name]
	if !ok {
		return CacheStats{}, false
	}

	return cache.Stats(), true
}

// reportLoop periodically generates and sends metrics reports.
func (m *Metrics) reportLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.reportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.sendReport()
		case <-m.stopCh:
			return
		}
	}
}

// sendReport generates and sends a metrics report to all reporters.
func (m *Metrics) sendReport() {
	report := m.GetReport()

	m.cachesMu.RLock()
	reporters := make([]MetricsReporter, len(m.reporters))
	copy(reporters, m.reporters)
	m.cachesMu.RUnlock()

	for _, reporter := range reporters {
		go reporter(report)
	}
}

// CacheMetricsCollector collects metrics from a cache and exposes them
// in a format suitable for Prometheus or other monitoring systems.
type CacheMetricsCollector struct {
	cache     MetricsProvider
	namespace string
	name      string
}

// NewCacheMetricsCollector creates a new metrics collector for a cache.
func NewCacheMetricsCollector(cache MetricsProvider, namespace, name string) *CacheMetricsCollector {
	return &CacheMetricsCollector{
		cache:     cache,
		namespace: namespace,
		name:      name,
	}
}

// Describe returns metric descriptions (for Prometheus compatibility).
func (c *CacheMetricsCollector) Describe() []MetricDesc {
	return []MetricDesc{
		{Name: "cache_hits_total", Help: "Total number of cache hits", Type: "counter"},
		{Name: "cache_misses_total", Help: "Total number of cache misses", Type: "counter"},
		{Name: "cache_evictions_total", Help: "Total number of cache evictions", Type: "counter"},
		{Name: "cache_expirations_total", Help: "Total number of cache expirations", Type: "counter"},
		{Name: "cache_size", Help: "Current number of items in cache", Type: "gauge"},
		{Name: "cache_max_size", Help: "Maximum cache capacity", Type: "gauge"},
		{Name: "cache_hit_rate", Help: "Cache hit rate percentage", Type: "gauge"},
	}
}

// Collect returns current metric values.
func (c *CacheMetricsCollector) Collect() []MetricValue {
	stats := c.cache.Stats()

	return []MetricValue{
		{Name: c.fqName("cache_hits_total"), Value: float64(stats.Hits), Labels: map[string]string{"cache": c.name}},
		{Name: c.fqName("cache_misses_total"), Value: float64(stats.Misses), Labels: map[string]string{"cache": c.name}},
		{Name: c.fqName("cache_evictions_total"), Value: float64(stats.Evictions), Labels: map[string]string{"cache": c.name}},
		{Name: c.fqName("cache_expirations_total"), Value: float64(stats.Expirations), Labels: map[string]string{"cache": c.name}},
		{Name: c.fqName("cache_size"), Value: float64(stats.Size), Labels: map[string]string{"cache": c.name}},
		{Name: c.fqName("cache_max_size"), Value: float64(stats.MaxSize), Labels: map[string]string{"cache": c.name}},
		{Name: c.fqName("cache_hit_rate"), Value: stats.HitRate(), Labels: map[string]string{"cache": c.name}},
	}
}

func (c *CacheMetricsCollector) fqName(name string) string {
	if c.namespace == "" {
		return name
	}
	return c.namespace + "_" + name
}

// MetricDesc describes a metric.
type MetricDesc struct {
	Name string
	Help string
	Type string // "counter", "gauge", "histogram"
}

// MetricValue represents a metric value with labels.
type MetricValue struct {
	Name   string
	Value  float64
	Labels map[string]string
}

