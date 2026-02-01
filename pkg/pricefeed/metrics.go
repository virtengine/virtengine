package pricefeed

import (
	"sync"
	"sync/atomic"
)

// ============================================================================
// Prometheus Metrics Implementation
// ============================================================================

// PrometheusMetrics implements the Metrics interface for Prometheus
type PrometheusMetrics struct {
	// Fetch metrics
	fetchTotal   map[string]*atomic.Int64 // source -> count
	fetchSuccess map[string]*atomic.Int64 // source -> count
	fetchErrors  map[string]*atomic.Int64 // source -> count

	// Cache metrics
	cacheHits   atomic.Int64
	cacheMisses atomic.Int64

	// Deviation metrics
	deviations map[string]float64 // pair -> last deviation

	// Health metrics
	sourceHealthy map[string]bool    // source -> healthy
	sourceLatency map[string]float64 // source -> latency seconds

	mu sync.RWMutex
}

// NewPrometheusMetrics creates a new Prometheus metrics collector
func NewPrometheusMetrics() *PrometheusMetrics {
	return &PrometheusMetrics{
		fetchTotal:    make(map[string]*atomic.Int64),
		fetchSuccess:  make(map[string]*atomic.Int64),
		fetchErrors:   make(map[string]*atomic.Int64),
		deviations:    make(map[string]float64),
		sourceHealthy: make(map[string]bool),
		sourceLatency: make(map[string]float64),
	}
}

// RecordFetch records a price fetch operation
func (m *PrometheusMetrics) RecordFetch(source string, baseAsset, quoteAsset string, latency float64, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.fetchTotal[source]; !ok {
		m.fetchTotal[source] = &atomic.Int64{}
		m.fetchSuccess[source] = &atomic.Int64{}
		m.fetchErrors[source] = &atomic.Int64{}
	}

	m.fetchTotal[source].Add(1)
	if success {
		m.fetchSuccess[source].Add(1)
	} else {
		m.fetchErrors[source].Add(1)
	}

	m.sourceLatency[source] = latency
}

// RecordCacheHit records a cache hit
func (m *PrometheusMetrics) RecordCacheHit(baseAsset, quoteAsset string) {
	m.cacheHits.Add(1)
}

// RecordCacheMiss records a cache miss
func (m *PrometheusMetrics) RecordCacheMiss(baseAsset, quoteAsset string) {
	m.cacheMisses.Add(1)
}

// RecordPriceDeviation records price deviation between sources
func (m *PrometheusMetrics) RecordPriceDeviation(baseAsset, quoteAsset string, deviation float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := baseAsset + "/" + quoteAsset
	m.deviations[key] = deviation
}

// RecordSourceHealth records source health status
func (m *PrometheusMetrics) RecordSourceHealth(source string, healthy bool, latency float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sourceHealthy[source] = healthy
	m.sourceLatency[source] = latency
}

// ============================================================================
// Metrics Export Methods
// ============================================================================

// MetricsSummary contains a summary of all metrics
type MetricsSummary struct {
	// Per-source metrics
	Sources map[string]SourceMetrics `json:"sources"`

	// Cache metrics
	CacheHits   int64   `json:"cache_hits"`
	CacheMisses int64   `json:"cache_misses"`
	CacheHitRate float64 `json:"cache_hit_rate"`

	// Deviation metrics
	Deviations map[string]float64 `json:"deviations"`
}

// SourceMetrics contains metrics for a single source
type SourceMetrics struct {
	TotalRequests   int64   `json:"total_requests"`
	SuccessRequests int64   `json:"success_requests"`
	ErrorRequests   int64   `json:"error_requests"`
	ErrorRate       float64 `json:"error_rate"`
	Healthy         bool    `json:"healthy"`
	LatencySeconds  float64 `json:"latency_seconds"`
}

// Summary returns a summary of all collected metrics
func (m *PrometheusMetrics) Summary() MetricsSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summary := MetricsSummary{
		Sources:     make(map[string]SourceMetrics),
		CacheHits:   m.cacheHits.Load(),
		CacheMisses: m.cacheMisses.Load(),
		Deviations:  make(map[string]float64),
	}

	totalCache := summary.CacheHits + summary.CacheMisses
	if totalCache > 0 {
		summary.CacheHitRate = float64(summary.CacheHits) / float64(totalCache) * 100
	}

	for source := range m.fetchTotal {
		total := m.fetchTotal[source].Load()
		success := m.fetchSuccess[source].Load()
		errors := m.fetchErrors[source].Load()

		var errorRate float64
		if total > 0 {
			errorRate = float64(errors) / float64(total) * 100
		}

		summary.Sources[source] = SourceMetrics{
			TotalRequests:   total,
			SuccessRequests: success,
			ErrorRequests:   errors,
			ErrorRate:       errorRate,
			Healthy:         m.sourceHealthy[source],
			LatencySeconds:  m.sourceLatency[source],
		}
	}

	for pair, dev := range m.deviations {
		summary.Deviations[pair] = dev
	}

	return summary
}

// Reset resets all metrics
func (m *PrometheusMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.fetchTotal = make(map[string]*atomic.Int64)
	m.fetchSuccess = make(map[string]*atomic.Int64)
	m.fetchErrors = make(map[string]*atomic.Int64)
	m.deviations = make(map[string]float64)
	m.sourceHealthy = make(map[string]bool)
	m.sourceLatency = make(map[string]float64)
	m.cacheHits.Store(0)
	m.cacheMisses.Store(0)
}

// ============================================================================
// No-Op Metrics (for testing or when metrics are disabled)
// ============================================================================

// NoOpMetrics is a no-op implementation of Metrics
type NoOpMetrics struct{}

func (m *NoOpMetrics) RecordFetch(source string, baseAsset, quoteAsset string, latency float64, success bool) {}
func (m *NoOpMetrics) RecordCacheHit(baseAsset, quoteAsset string)                                           {}
func (m *NoOpMetrics) RecordCacheMiss(baseAsset, quoteAsset string)                                          {}
func (m *NoOpMetrics) RecordPriceDeviation(baseAsset, quoteAsset string, deviation float64)                  {}
func (m *NoOpMetrics) RecordSourceHealth(source string, healthy bool, latency float64)                       {}

