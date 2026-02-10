// Package email provides Prometheus metrics for email verification service.
package email

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// ============================================================================
// Metrics Collector
// ============================================================================

// Metrics contains Prometheus metrics for the email verification service
type Metrics struct {
	// Counters
	ChallengesCreated    *prometheus.CounterVec
	ChallengesVerified   *prometheus.CounterVec
	ChallengesFailed     *prometheus.CounterVec
	ChallengesExpired    *prometheus.CounterVec
	EmailsSent           *prometheus.CounterVec
	EmailsDelivered      *prometheus.CounterVec
	EmailsBounced        *prometheus.CounterVec
	VerificationAttempts *prometheus.CounterVec
	ResendsTotal         *prometheus.CounterVec
	RateLimitHits        prometheus.Counter
	AttestationsCreated  prometheus.Counter

	// Gauges
	ActiveChallenges  prometheus.Gauge
	PendingDeliveries prometheus.Gauge

	// Histograms
	VerificationLatency *prometheus.HistogramVec
	DeliveryLatency     *prometheus.HistogramVec
	EmailSendLatency    prometheus.Histogram

	// registry tracks if metrics are registered
	registered bool
	mu         sync.Mutex
}

// DefaultMetrics is the default metrics instance
var DefaultMetrics = NewMetrics()

// NewMetrics creates a new Metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		ChallengesCreated: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "email_verification",
				Name:      "challenges_created_total",
				Help:      "Total number of email verification challenges created",
			},
			[]string{"method"}, // otp, magic_link
		),

		ChallengesVerified: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "email_verification",
				Name:      "challenges_verified_total",
				Help:      "Total number of email verification challenges successfully verified",
			},
			[]string{"method"},
		),

		ChallengesFailed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "email_verification",
				Name:      "challenges_failed_total",
				Help:      "Total number of email verification challenges that failed",
			},
			[]string{"method", "reason"}, // max_attempts, expired, bounced
		),

		ChallengesExpired: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "email_verification",
				Name:      "challenges_expired_total",
				Help:      "Total number of email verification challenges that expired",
			},
			[]string{"method"},
		),

		EmailsSent: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "email_verification",
				Name:      "emails_sent_total",
				Help:      "Total number of verification emails sent",
			},
			[]string{"provider", "template"}, // ses/sendgrid, otp/magic_link
		),

		EmailsDelivered: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "email_verification",
				Name:      "emails_delivered_total",
				Help:      "Total number of verification emails successfully delivered",
			},
			[]string{"provider"},
		),

		EmailsBounced: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "email_verification",
				Name:      "emails_bounced_total",
				Help:      "Total number of verification emails that bounced",
			},
			[]string{"provider", "bounce_type"}, // hard, soft
		),

		VerificationAttempts: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "email_verification",
				Name:      "verification_attempts_total",
				Help:      "Total number of verification attempts",
			},
			[]string{"method", "success"}, // true/false
		),

		ResendsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "email_verification",
				Name:      "resends_total",
				Help:      "Total number of verification email resends",
			},
			[]string{"method"},
		),

		RateLimitHits: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "email_verification",
				Name:      "rate_limit_hits_total",
				Help:      "Total number of rate limit hits",
			},
		),

		AttestationsCreated: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "email_verification",
				Name:      "attestations_created_total",
				Help:      "Total number of email verification attestations created",
			},
		),

		ActiveChallenges: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "veid",
				Subsystem: "email_verification",
				Name:      "active_challenges",
				Help:      "Number of currently active (non-expired) verification challenges",
			},
		),

		PendingDeliveries: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "veid",
				Subsystem: "email_verification",
				Name:      "pending_deliveries",
				Help:      "Number of emails pending delivery confirmation",
			},
		),

		VerificationLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "veid",
				Subsystem: "email_verification",
				Name:      "verification_latency_seconds",
				Help:      "Time from challenge creation to successful verification",
				Buckets:   []float64{10, 30, 60, 120, 300, 600, 1800, 3600}, // 10s to 1h
			},
			[]string{"method"},
		),

		DeliveryLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "veid",
				Subsystem: "email_verification",
				Name:      "delivery_latency_seconds",
				Help:      "Time from email send to delivery confirmation",
				Buckets:   []float64{1, 2, 5, 10, 30, 60, 120, 300}, // 1s to 5m
			},
			[]string{"provider"},
		),

		EmailSendLatency: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: "veid",
				Subsystem: "email_verification",
				Name:      "email_send_latency_seconds",
				Help:      "Time to send email to provider",
				Buckets:   []float64{0.1, 0.25, 0.5, 1, 2, 5, 10}, // 100ms to 10s
			},
		),
	}
}

// Register registers all metrics with the default Prometheus registry
func (m *Metrics) Register() error {
	return m.RegisterWith(prometheus.DefaultRegisterer)
}

// RegisterWith registers all metrics with the provided registerer
func (m *Metrics) RegisterWith(registerer prometheus.Registerer) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.registered {
		return nil
	}

	collectors := []prometheus.Collector{
		m.ChallengesCreated,
		m.ChallengesVerified,
		m.ChallengesFailed,
		m.ChallengesExpired,
		m.EmailsSent,
		m.EmailsDelivered,
		m.EmailsBounced,
		m.VerificationAttempts,
		m.ResendsTotal,
		m.RateLimitHits,
		m.AttestationsCreated,
		m.ActiveChallenges,
		m.PendingDeliveries,
		m.VerificationLatency,
		m.DeliveryLatency,
		m.EmailSendLatency,
	}

	for _, c := range collectors {
		if err := registerer.Register(c); err != nil {
			// Check if already registered
			if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
				return err
			}
		}
	}

	m.registered = true
	return nil
}

// ============================================================================
// Metric Recording Helpers
// ============================================================================

// RecordChallengeCreated records a new challenge creation
func (m *Metrics) RecordChallengeCreated(method VerificationMethod) {
	m.ChallengesCreated.WithLabelValues(string(method)).Inc()
	m.ActiveChallenges.Inc()
}

// RecordChallengeVerified records a successful verification
func (m *Metrics) RecordChallengeVerified(method VerificationMethod, latency time.Duration) {
	m.ChallengesVerified.WithLabelValues(string(method)).Inc()
	m.ActiveChallenges.Dec()
	m.VerificationLatency.WithLabelValues(string(method)).Observe(latency.Seconds())
}

// RecordChallengeFailed records a failed verification
func (m *Metrics) RecordChallengeFailed(method VerificationMethod, reason string) {
	m.ChallengesFailed.WithLabelValues(string(method), reason).Inc()
	m.ActiveChallenges.Dec()
}

// RecordChallengeExpired records an expired challenge
func (m *Metrics) RecordChallengeExpired(method VerificationMethod) {
	m.ChallengesExpired.WithLabelValues(string(method)).Inc()
	m.ActiveChallenges.Dec()
}

// RecordEmailSent records an email being sent
func (m *Metrics) RecordEmailSent(provider string, template TemplateType, latency time.Duration) {
	m.EmailsSent.WithLabelValues(provider, string(template)).Inc()
	m.EmailSendLatency.Observe(latency.Seconds())
	m.PendingDeliveries.Inc()
}

// RecordEmailDelivered records a successful email delivery
func (m *Metrics) RecordEmailDelivered(provider string, latency time.Duration) {
	m.EmailsDelivered.WithLabelValues(provider).Inc()
	m.PendingDeliveries.Dec()
	m.DeliveryLatency.WithLabelValues(provider).Observe(latency.Seconds())
}

// RecordEmailBounced records an email bounce
func (m *Metrics) RecordEmailBounced(provider, bounceType string) {
	m.EmailsBounced.WithLabelValues(provider, bounceType).Inc()
	m.PendingDeliveries.Dec()
}

// RecordVerificationAttempt records a verification attempt
func (m *Metrics) RecordVerificationAttempt(method VerificationMethod, success bool) {
	successStr := "false"
	if success {
		successStr = "true"
	}
	m.VerificationAttempts.WithLabelValues(string(method), successStr).Inc()
}

// RecordResend records a resend
func (m *Metrics) RecordResend(method VerificationMethod) {
	m.ResendsTotal.WithLabelValues(string(method)).Inc()
}

// RecordRateLimitHit records a rate limit hit
func (m *Metrics) RecordRateLimitHit() {
	m.RateLimitHits.Inc()
}

// RecordAttestationCreated records an attestation creation
func (m *Metrics) RecordAttestationCreated() {
	m.AttestationsCreated.Inc()
}

// SetActiveChallenges sets the active challenges gauge
func (m *Metrics) SetActiveChallenges(count float64) {
	m.ActiveChallenges.Set(count)
}

// SetPendingDeliveries sets the pending deliveries gauge
func (m *Metrics) SetPendingDeliveries(count float64) {
	m.PendingDeliveries.Set(count)
}

// ============================================================================
// Dashboard Metrics Summary
// ============================================================================

// MetricsSummary contains a summary of key metrics for dashboards
type MetricsSummary struct {
	// Challenge metrics
	TotalChallengesCreated  int64   `json:"total_challenges_created"`
	TotalChallengesVerified int64   `json:"total_challenges_verified"`
	TotalChallengesFailed   int64   `json:"total_challenges_failed"`
	VerificationSuccessRate float64 `json:"verification_success_rate"`
	ActiveChallenges        int64   `json:"active_challenges"`

	// Email metrics
	TotalEmailsSent          int64   `json:"total_emails_sent"`
	TotalEmailsDelivered     int64   `json:"total_emails_delivered"`
	TotalEmailsBounced       int64   `json:"total_emails_bounced"`
	DeliverySuccessRate      float64 `json:"delivery_success_rate"`
	AverageDeliveryLatencyMs float64 `json:"average_delivery_latency_ms"`

	// Attestation metrics
	TotalAttestationsCreated int64 `json:"total_attestations_created"`

	// Rate limiting
	TotalRateLimitHits int64 `json:"total_rate_limit_hits"`

	// Timestamp
	Timestamp time.Time `json:"timestamp"`
}

// ============================================================================
// Alert Thresholds
// ============================================================================

// AlertThresholds defines thresholds for alerting
type AlertThresholds struct {
	// MinDeliverySuccessRate is the minimum acceptable delivery success rate (0-1)
	MinDeliverySuccessRate float64 `json:"min_delivery_success_rate"`

	// MaxBounceRate is the maximum acceptable bounce rate (0-1)
	MaxBounceRate float64 `json:"max_bounce_rate"`

	// MaxAverageDeliveryLatencySeconds is the maximum acceptable average delivery latency
	MaxAverageDeliveryLatencySeconds float64 `json:"max_average_delivery_latency_seconds"`

	// MaxActiveChallenges is the maximum number of active challenges before alerting
	MaxActiveChallenges int64 `json:"max_active_challenges"`
}

// DefaultAlertThresholds returns the default alert thresholds
func DefaultAlertThresholds() AlertThresholds {
	return AlertThresholds{
		MinDeliverySuccessRate:           0.95, // 95%
		MaxBounceRate:                    0.05, // 5%
		MaxAverageDeliveryLatencySeconds: 60,   // 1 minute
		MaxActiveChallenges:              10000,
	}
}
