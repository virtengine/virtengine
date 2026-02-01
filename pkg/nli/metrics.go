package nli

import (
	"context"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

var (
	metricsOnce     sync.Once
	defaultMetrics  *NLIMetrics
	metricsRegistry = prometheus.NewRegistry()
)

// NLIMetrics provides metrics collection for the NLI service
type NLIMetrics struct {
	// Session metrics
	activeSessions       prometheus.Gauge
	sessionOperations    *prometheus.CounterVec
	sessionOperationTime *prometheus.HistogramVec

	// Rate limit metrics
	rateLimitHits   prometheus.Counter
	rateLimitMisses prometheus.Counter

	// Request metrics
	requestsTotal    *prometheus.CounterVec
	requestDuration  *prometheus.HistogramVec
	intentClassified *prometheus.CounterVec

	// Error metrics
	errors *prometheus.CounterVec

	logger zerolog.Logger
}

// NewNLIMetrics creates a new metrics collector (singleton for tests)
func NewNLIMetrics(namespace string, logger zerolog.Logger) *NLIMetrics {
	if namespace == "" {
		namespace = "virtengine"
	}

	metricsOnce.Do(func() {
		defaultMetrics = createMetrics(namespace, logger)
	})

	// Update logger for this instance
	defaultMetrics.logger = logger
	return defaultMetrics
}

// createMetrics creates the actual metric instances
func createMetrics(namespace string, logger zerolog.Logger) *NLIMetrics {
	activeSessions := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "nli",
		Name:      "active_sessions",
		Help:      "Number of active NLI sessions",
	})

	sessionOperations := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "nli",
			Name:      "session_operations_total",
			Help:      "Total number of session operations",
		},
		[]string{"operation"},
	)

	sessionOperationTime := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "nli",
			Name:      "session_operation_duration_seconds",
			Help:      "Duration of session operations in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	rateLimitHits := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "nli",
		Name:      "ratelimit_hits_total",
		Help:      "Total number of rate limit hits (blocked requests)",
	})

	rateLimitMisses := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "nli",
		Name:      "ratelimit_passes_total",
		Help:      "Total number of requests that passed rate limiting",
	})

	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "nli",
			Name:      "requests_total",
			Help:      "Total number of NLI requests",
		},
		[]string{"status"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "nli",
			Name:      "request_duration_seconds",
			Help:      "Duration of NLI requests in seconds",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"intent"},
	)

	intentClassified := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "nli",
			Name:      "intents_classified_total",
			Help:      "Total number of intents classified",
		},
		[]string{"intent"},
	)

	errors := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "nli",
			Name:      "errors_total",
			Help:      "Total number of errors",
		},
		[]string{"type"},
	)

	// Register with custom registry (won't panic on re-registration)
	metricsRegistry.MustRegister(
		activeSessions,
		sessionOperations,
		sessionOperationTime,
		rateLimitHits,
		rateLimitMisses,
		requestsTotal,
		requestDuration,
		intentClassified,
		errors,
	)

	return &NLIMetrics{
		activeSessions:       activeSessions,
		sessionOperations:    sessionOperations,
		sessionOperationTime: sessionOperationTime,
		rateLimitHits:        rateLimitHits,
		rateLimitMisses:      rateLimitMisses,
		requestsTotal:        requestsTotal,
		requestDuration:      requestDuration,
		intentClassified:     intentClassified,
		errors:               errors,
		logger:               logger,
	}
}

// RecordSessionCount updates the active session count
func (m *NLIMetrics) RecordSessionCount(count int64) {
	m.activeSessions.Set(float64(count))
	telemetry.SetGauge(float32(count), "nli", "active_sessions")
}

// RecordSessionOperation records a session store operation
func (m *NLIMetrics) RecordSessionOperation(operation string, duration time.Duration) {
	m.sessionOperations.WithLabelValues(operation).Inc()
	m.sessionOperationTime.WithLabelValues(operation).Observe(duration.Seconds())
	telemetry.IncrCounter(1, "nli", "session_operations", operation)
}

// RecordRateLimitHit records a rate limit hit (blocked request)
func (m *NLIMetrics) RecordRateLimitHit() {
	m.rateLimitHits.Inc()
	telemetry.IncrCounter(1, "nli", "ratelimit_hits")
}

// RecordRateLimitPass records a request that passed rate limiting
func (m *NLIMetrics) RecordRateLimitPass() {
	m.rateLimitMisses.Inc()
	telemetry.IncrCounter(1, "nli", "ratelimit_passes")
}

// RecordRequest records a completed request
func (m *NLIMetrics) RecordRequest(status string, intent string, duration time.Duration) {
	m.requestsTotal.WithLabelValues(status).Inc()
	m.requestDuration.WithLabelValues(intent).Observe(duration.Seconds())
	telemetry.IncrCounter(1, "nli", "requests", status)
	telemetry.MeasureSince(time.Now().Add(-duration), "nli", "request_duration")
}

// RecordIntentClassified records an intent classification
func (m *NLIMetrics) RecordIntentClassified(intent string) {
	m.intentClassified.WithLabelValues(intent).Inc()
	telemetry.IncrCounter(1, "nli", "intents_classified", intent)
}

// RecordError records an error
func (m *NLIMetrics) RecordError(errorType string) {
	m.errors.WithLabelValues(errorType).Inc()
	telemetry.IncrCounter(1, "nli", "errors", errorType)
}

// MetricsCollector collects and reports metrics periodically
type MetricsCollector struct {
	metrics      *NLIMetrics
	sessionStore SessionStore
	logger       zerolog.Logger
	stopCh       chan struct{}
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(metrics *NLIMetrics, sessionStore SessionStore, logger zerolog.Logger) *MetricsCollector {
	return &MetricsCollector{
		metrics:      metrics,
		sessionStore: sessionStore,
		logger:       logger,
		stopCh:       make(chan struct{}),
	}
}

// Start starts the metrics collection loop
func (c *MetricsCollector) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	c.logger.Info().Dur("interval", interval).Msg("starting NLI metrics collector")

	for {
		select {
		case <-ctx.Done():
			c.logger.Info().Msg("stopping NLI metrics collector")
			return
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.collectMetrics(ctx)
		}
	}
}

// Stop stops the metrics collector
func (c *MetricsCollector) Stop() {
	close(c.stopCh)
}

// collectMetrics collects metrics from the session store
func (c *MetricsCollector) collectMetrics(ctx context.Context) {
	// Get session count
	count, err := c.sessionStore.Count(ctx)
	if err != nil {
		c.logger.Warn().Err(err).Msg("failed to get session count")
		return
	}

	c.metrics.RecordSessionCount(count)

	c.logger.Debug().Int64("session_count", count).Msg("metrics collected")
}

