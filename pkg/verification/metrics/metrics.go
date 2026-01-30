// Package metrics provides Prometheus metrics for verification services.
//
// This package implements metrics collection and exposure for monitoring
// verification service health and performance.
//
// Task Reference: VE-2B - Verification Shared Infrastructure
package metrics

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ============================================================================
// Metrics Collector
// ============================================================================

// Collector collects and exposes metrics for verification services.
type Collector struct {
	mu       sync.RWMutex
	registry *prometheus.Registry
	config   Config

	// Signer metrics
	signerSignRequests       *prometheus.CounterVec
	signerVerifyRequests     *prometheus.CounterVec
	signerSignLatency        *prometheus.HistogramVec
	signerVerifyLatency      *prometheus.HistogramVec
	signerActiveKeys         prometheus.Gauge
	signerKeyRotations       prometheus.Counter
	signerKeyAge             prometheus.Gauge
	signerErrors             *prometheus.CounterVec

	// Nonce metrics
	nonceCreated      prometheus.Counter
	nonceUsed         prometheus.Counter
	nonceRejected     *prometheus.CounterVec
	nonceExpired      prometheus.Counter
	nonceStoreSize    prometheus.Gauge
	nonceValidateTime prometheus.Histogram

	// Rate limiting metrics
	rateLimitChecks    *prometheus.CounterVec
	rateLimitBlocked   *prometheus.CounterVec
	rateLimitBans      *prometheus.CounterVec
	abuseScoreGauge    *prometheus.GaugeVec

	// Verification metrics
	verificationRequests  *prometheus.CounterVec
	verificationSuccess   *prometheus.CounterVec
	verificationFailures  *prometheus.CounterVec
	verificationLatency   *prometheus.HistogramVec
	verificationScores    *prometheus.HistogramVec

	// Audit metrics
	auditEventsLogged   *prometheus.CounterVec
	auditBufferSize     prometheus.Gauge
	auditFlushLatency   prometheus.Histogram

	// Health metrics
	serviceHealth       *prometheus.GaugeVec
	lastHealthCheck     prometheus.Gauge
}

// Config contains metrics configuration.
type Config struct {
	// Namespace is the metrics namespace
	Namespace string `json:"namespace"`

	// Subsystem is the metrics subsystem
	Subsystem string `json:"subsystem"`

	// HTTPPath is the path for the metrics endpoint
	HTTPPath string `json:"http_path"`

	// HTTPPort is the port for the metrics endpoint
	HTTPPort int `json:"http_port"`

	// EnableGoMetrics enables Go runtime metrics
	EnableGoMetrics bool `json:"enable_go_metrics"`

	// EnableProcessMetrics enables process metrics
	EnableProcessMetrics bool `json:"enable_process_metrics"`

	// LatencyBuckets are the histogram buckets for latency metrics
	LatencyBuckets []float64 `json:"latency_buckets"`

	// ScoreBuckets are the histogram buckets for score metrics
	ScoreBuckets []float64 `json:"score_buckets"`
}

// DefaultConfig returns the default metrics configuration.
func DefaultConfig() Config {
	return Config{
		Namespace:            "virtengine",
		Subsystem:            "verification",
		HTTPPath:             "/metrics",
		HTTPPort:             9090,
		EnableGoMetrics:      true,
		EnableProcessMetrics: true,
		LatencyBuckets:       []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		ScoreBuckets:         []float64{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
	}
}

// NewCollector creates a new metrics collector.
func NewCollector(config Config) (*Collector, error) {
	registry := prometheus.NewRegistry()

	if config.EnableGoMetrics {
		registry.MustRegister(prometheus.NewGoCollector())
	}
	if config.EnableProcessMetrics {
		registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	}

	if config.LatencyBuckets == nil {
		config.LatencyBuckets = DefaultConfig().LatencyBuckets
	}
	if config.ScoreBuckets == nil {
		config.ScoreBuckets = DefaultConfig().ScoreBuckets
	}

	c := &Collector{
		registry: registry,
		config:   config,
	}

	c.initSignerMetrics()
	c.initNonceMetrics()
	c.initRateLimitMetrics()
	c.initVerificationMetrics()
	c.initAuditMetrics()
	c.initHealthMetrics()

	return c, nil
}

func (c *Collector) initSignerMetrics() {
	c.signerSignRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "signer_sign_requests_total",
			Help:      "Total number of attestation signing requests",
		},
		[]string{"signer_id", "attestation_type", "status"},
	)

	c.signerVerifyRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "signer_verify_requests_total",
			Help:      "Total number of attestation verification requests",
		},
		[]string{"signer_id", "status"},
	)

	c.signerSignLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "signer_sign_latency_seconds",
			Help:      "Latency of attestation signing operations",
			Buckets:   c.config.LatencyBuckets,
		},
		[]string{"signer_id", "attestation_type"},
	)

	c.signerVerifyLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "signer_verify_latency_seconds",
			Help:      "Latency of attestation verification operations",
			Buckets:   c.config.LatencyBuckets,
		},
		[]string{"signer_id"},
	)

	c.signerActiveKeys = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "signer_active_keys",
			Help:      "Number of active signing keys",
		},
	)

	c.signerKeyRotations = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "signer_key_rotations_total",
			Help:      "Total number of key rotations",
		},
	)

	c.signerKeyAge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "signer_key_age_seconds",
			Help:      "Age of the active signing key in seconds",
		},
	)

	c.signerErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "signer_errors_total",
			Help:      "Total number of signer errors",
		},
		[]string{"signer_id", "error_type"},
	)

	c.registry.MustRegister(
		c.signerSignRequests,
		c.signerVerifyRequests,
		c.signerSignLatency,
		c.signerVerifyLatency,
		c.signerActiveKeys,
		c.signerKeyRotations,
		c.signerKeyAge,
		c.signerErrors,
	)
}

func (c *Collector) initNonceMetrics() {
	c.nonceCreated = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "nonce_created_total",
			Help:      "Total number of nonces created",
		},
	)

	c.nonceUsed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "nonce_used_total",
			Help:      "Total number of nonces used",
		},
	)

	c.nonceRejected = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "nonce_rejected_total",
			Help:      "Total number of nonces rejected",
		},
		[]string{"reason"},
	)

	c.nonceExpired = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "nonce_expired_total",
			Help:      "Total number of nonces expired",
		},
	)

	c.nonceStoreSize = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "nonce_store_size",
			Help:      "Current size of the nonce store",
		},
	)

	c.nonceValidateTime = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "nonce_validate_latency_seconds",
			Help:      "Latency of nonce validation operations",
			Buckets:   c.config.LatencyBuckets,
		},
	)

	c.registry.MustRegister(
		c.nonceCreated,
		c.nonceUsed,
		c.nonceRejected,
		c.nonceExpired,
		c.nonceStoreSize,
		c.nonceValidateTime,
	)
}

func (c *Collector) initRateLimitMetrics() {
	c.rateLimitChecks = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "ratelimit_checks_total",
			Help:      "Total number of rate limit checks",
		},
		[]string{"limit_type", "result"},
	)

	c.rateLimitBlocked = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "ratelimit_blocked_total",
			Help:      "Total number of blocked requests",
		},
		[]string{"limit_type", "reason"},
	)

	c.rateLimitBans = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "ratelimit_bans_total",
			Help:      "Total number of bans issued",
		},
		[]string{"reason"},
	)

	c.abuseScoreGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "abuse_score",
			Help:      "Current abuse score by identifier",
		},
		[]string{"identifier_type"},
	)

	c.registry.MustRegister(
		c.rateLimitChecks,
		c.rateLimitBlocked,
		c.rateLimitBans,
		c.abuseScoreGauge,
	)
}

func (c *Collector) initVerificationMetrics() {
	c.verificationRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "verification_requests_total",
			Help:      "Total number of verification requests",
		},
		[]string{"type"},
	)

	c.verificationSuccess = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "verification_success_total",
			Help:      "Total number of successful verifications",
		},
		[]string{"type"},
	)

	c.verificationFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "verification_failures_total",
			Help:      "Total number of failed verifications",
		},
		[]string{"type", "reason"},
	)

	c.verificationLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "verification_latency_seconds",
			Help:      "Latency of verification operations",
			Buckets:   c.config.LatencyBuckets,
		},
		[]string{"type"},
	)

	c.verificationScores = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "verification_scores",
			Help:      "Distribution of verification scores",
			Buckets:   c.config.ScoreBuckets,
		},
		[]string{"type"},
	)

	c.registry.MustRegister(
		c.verificationRequests,
		c.verificationSuccess,
		c.verificationFailures,
		c.verificationLatency,
		c.verificationScores,
	)
}

func (c *Collector) initAuditMetrics() {
	c.auditEventsLogged = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "audit_events_total",
			Help:      "Total number of audit events logged",
		},
		[]string{"event_type", "severity"},
	)

	c.auditBufferSize = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "audit_buffer_size",
			Help:      "Current size of the audit event buffer",
		},
	)

	c.auditFlushLatency = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "audit_flush_latency_seconds",
			Help:      "Latency of audit buffer flush operations",
			Buckets:   c.config.LatencyBuckets,
		},
	)

	c.registry.MustRegister(
		c.auditEventsLogged,
		c.auditBufferSize,
		c.auditFlushLatency,
	)
}

func (c *Collector) initHealthMetrics() {
	c.serviceHealth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "service_health",
			Help:      "Health status of verification services (1=healthy, 0=unhealthy)",
		},
		[]string{"service"},
	)

	c.lastHealthCheck = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: c.config.Namespace,
			Subsystem: c.config.Subsystem,
			Name:      "last_health_check_timestamp",
			Help:      "Timestamp of the last health check",
		},
	)

	c.registry.MustRegister(
		c.serviceHealth,
		c.lastHealthCheck,
	)
}

// ============================================================================
// Recording Methods
// ============================================================================

// RecordSignRequest records a signing request.
func (c *Collector) RecordSignRequest(signerID, attestationType, status string, duration time.Duration) {
	c.signerSignRequests.WithLabelValues(signerID, attestationType, status).Inc()
	c.signerSignLatency.WithLabelValues(signerID, attestationType).Observe(duration.Seconds())
}

// RecordVerifyRequest records a verification request.
func (c *Collector) RecordVerifyRequest(signerID, status string, duration time.Duration) {
	c.signerVerifyRequests.WithLabelValues(signerID, status).Inc()
	c.signerVerifyLatency.WithLabelValues(signerID).Observe(duration.Seconds())
}

// RecordKeyRotation records a key rotation.
func (c *Collector) RecordKeyRotation() {
	c.signerKeyRotations.Inc()
}

// SetActiveKeys sets the number of active keys.
func (c *Collector) SetActiveKeys(count int) {
	c.signerActiveKeys.Set(float64(count))
}

// SetKeyAge sets the age of the active key.
func (c *Collector) SetKeyAge(age time.Duration) {
	c.signerKeyAge.Set(age.Seconds())
}

// RecordSignerError records a signer error.
func (c *Collector) RecordSignerError(signerID, errorType string) {
	c.signerErrors.WithLabelValues(signerID, errorType).Inc()
}

// RecordNonceCreated records a nonce creation.
func (c *Collector) RecordNonceCreated() {
	c.nonceCreated.Inc()
}

// RecordNonceUsed records a nonce usage.
func (c *Collector) RecordNonceUsed() {
	c.nonceUsed.Inc()
}

// RecordNonceRejected records a nonce rejection.
func (c *Collector) RecordNonceRejected(reason string) {
	c.nonceRejected.WithLabelValues(reason).Inc()
}

// RecordNonceExpired records a nonce expiration.
func (c *Collector) RecordNonceExpired() {
	c.nonceExpired.Inc()
}

// SetNonceStoreSize sets the nonce store size.
func (c *Collector) SetNonceStoreSize(size int64) {
	c.nonceStoreSize.Set(float64(size))
}

// RecordNonceValidateTime records nonce validation latency.
func (c *Collector) RecordNonceValidateTime(duration time.Duration) {
	c.nonceValidateTime.Observe(duration.Seconds())
}

// RecordRateLimitCheck records a rate limit check.
func (c *Collector) RecordRateLimitCheck(limitType, result string) {
	c.rateLimitChecks.WithLabelValues(limitType, result).Inc()
}

// RecordRateLimitBlocked records a blocked request.
func (c *Collector) RecordRateLimitBlocked(limitType, reason string) {
	c.rateLimitBlocked.WithLabelValues(limitType, reason).Inc()
}

// RecordBan records a ban.
func (c *Collector) RecordBan(reason string) {
	c.rateLimitBans.WithLabelValues(reason).Inc()
}

// SetAbuseScore sets the abuse score for an identifier type.
func (c *Collector) SetAbuseScore(identifierType string, score float64) {
	c.abuseScoreGauge.WithLabelValues(identifierType).Set(score)
}

// RecordVerificationRequest records a verification request.
func (c *Collector) RecordVerificationRequest(verificationType string) {
	c.verificationRequests.WithLabelValues(verificationType).Inc()
}

// RecordVerificationSuccess records a successful verification.
func (c *Collector) RecordVerificationSuccess(verificationType string, duration time.Duration, score float64) {
	c.verificationSuccess.WithLabelValues(verificationType).Inc()
	c.verificationLatency.WithLabelValues(verificationType).Observe(duration.Seconds())
	c.verificationScores.WithLabelValues(verificationType).Observe(score)
}

// RecordVerificationFailure records a failed verification.
func (c *Collector) RecordVerificationFailure(verificationType, reason string, duration time.Duration) {
	c.verificationFailures.WithLabelValues(verificationType, reason).Inc()
	c.verificationLatency.WithLabelValues(verificationType).Observe(duration.Seconds())
}

// RecordAuditEvent records an audit event.
func (c *Collector) RecordAuditEvent(eventType, severity string) {
	c.auditEventsLogged.WithLabelValues(eventType, severity).Inc()
}

// SetAuditBufferSize sets the audit buffer size.
func (c *Collector) SetAuditBufferSize(size int) {
	c.auditBufferSize.Set(float64(size))
}

// RecordAuditFlush records an audit flush operation.
func (c *Collector) RecordAuditFlush(duration time.Duration) {
	c.auditFlushLatency.Observe(duration.Seconds())
}

// SetServiceHealth sets the health status of a service.
func (c *Collector) SetServiceHealth(service string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	c.serviceHealth.WithLabelValues(service).Set(value)
	c.lastHealthCheck.Set(float64(time.Now().Unix()))
}

// ============================================================================
// HTTP Handler
// ============================================================================

// Handler returns the HTTP handler for the metrics endpoint.
func (c *Collector) Handler() http.Handler {
	return promhttp.HandlerFor(c.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// ServeHTTP starts the metrics HTTP server.
func (c *Collector) ServeHTTP(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.Handle(c.config.HTTPPath, c.Handler())

	server := &http.Server{
		Addr:    ":" + string(rune(c.config.HTTPPort)),
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()

	return server.ListenAndServe()
}

// Registry returns the Prometheus registry.
func (c *Collector) Registry() *prometheus.Registry {
	return c.registry
}
