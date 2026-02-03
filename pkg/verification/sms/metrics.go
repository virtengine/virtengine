// Package sms provides Prometheus metrics for the SMS verification service.
//
// Task Reference: VE-4C - SMS Verification Delivery + Anti-Fraud
package sms

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// ============================================================================
// Metrics
// ============================================================================

// Metrics contains Prometheus metrics for SMS verification
type Metrics struct {
	mu sync.Once

	// Challenge metrics
	challengesCreated  *prometheus.CounterVec
	challengesVerified *prometheus.CounterVec
	challengesFailed   *prometheus.CounterVec
	challengesExpired  *prometheus.CounterVec

	// SMS delivery metrics
	smsSent      *prometheus.CounterVec
	smsDelivered *prometheus.CounterVec
	smsFailed    *prometheus.CounterVec
	smsLatency   *prometheus.HistogramVec

	// OTP metrics
	otpAttempts         *prometheus.CounterVec
	otpSuccessful       *prometheus.CounterVec
	otpFailed           *prometheus.CounterVec
	otpResends          *prometheus.CounterVec
	verificationLatency *prometheus.HistogramVec

	// Anti-fraud metrics
	voipDetected       *prometheus.CounterVec
	phoneBlocked       *prometheus.CounterVec
	ipBlocked          *prometheus.CounterVec
	velocityExceeded   *prometheus.CounterVec
	riskScoreHistogram *prometheus.HistogramVec
	fraudDetected      *prometheus.CounterVec

	// Rate limiting metrics
	rateLimitHit *prometheus.CounterVec

	// Attestation metrics
	attestationsCreated *prometheus.CounterVec
	attestationsFailed  *prometheus.CounterVec

	// Provider metrics
	providerFailover *prometheus.CounterVec
	providerHealth   *prometheus.GaugeVec

	// Carrier lookup metrics
	carrierLookups       *prometheus.CounterVec
	carrierLookupLatency *prometheus.HistogramVec
}

// DefaultMetrics is the default metrics instance
var DefaultMetrics = NewMetrics()

// NewMetrics creates a new Metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		challengesCreated: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "challenges_created_total",
				Help:      "Total number of SMS verification challenges created",
			},
			[]string{"country_code"},
		),
		challengesVerified: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "challenges_verified_total",
				Help:      "Total number of SMS verification challenges verified successfully",
			},
			[]string{"country_code"},
		),
		challengesFailed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "challenges_failed_total",
				Help:      "Total number of SMS verification challenges that failed",
			},
			[]string{"country_code", "reason"},
		),
		challengesExpired: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "challenges_expired_total",
				Help:      "Total number of SMS verification challenges that expired",
			},
			[]string{"country_code"},
		),
		smsSent: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "sent_total",
				Help:      "Total number of SMS messages sent",
			},
			[]string{"provider", "country_code"},
		),
		smsDelivered: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "delivered_total",
				Help:      "Total number of SMS messages delivered",
			},
			[]string{"provider", "country_code"},
		),
		smsFailed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "failed_total",
				Help:      "Total number of SMS messages that failed to send",
			},
			[]string{"provider", "error_code"},
		),
		smsLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "send_latency_seconds",
				Help:      "Latency of SMS sending in seconds",
				Buckets:   []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"provider"},
		),
		otpAttempts: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "otp_attempts_total",
				Help:      "Total number of OTP verification attempts",
			},
			[]string{"country_code"},
		),
		otpSuccessful: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "otp_successful_total",
				Help:      "Total number of successful OTP verifications",
			},
			[]string{"country_code"},
		),
		otpFailed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "otp_failed_total",
				Help:      "Total number of failed OTP verifications",
			},
			[]string{"country_code", "reason"},
		),
		otpResends: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "otp_resends_total",
				Help:      "Total number of OTP resends",
			},
			[]string{"country_code"},
		),
		verificationLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "verification_latency_seconds",
				Help:      "Time from challenge creation to successful verification",
				Buckets:   []float64{10, 30, 60, 120, 180, 300},
			},
			[]string{"country_code"},
		),
		voipDetected: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "voip_detected_total",
				Help:      "Total number of VoIP numbers detected",
			},
			[]string{"country_code", "carrier"},
		),
		phoneBlocked: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "phone_blocked_total",
				Help:      "Total number of phone numbers blocked",
			},
			[]string{"reason"},
		),
		ipBlocked: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "ip_blocked_total",
				Help:      "Total number of IP addresses blocked",
			},
			[]string{"reason"},
		),
		velocityExceeded: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "velocity_exceeded_total",
				Help:      "Total number of velocity limit exceeded events",
			},
			[]string{"type"},
		),
		riskScoreHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "risk_score",
				Help:      "Distribution of anti-fraud risk scores",
				Buckets:   []float64{10, 25, 50, 70, 80, 90, 100},
			},
			[]string{"country_code"},
		),
		fraudDetected: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "fraud_detected_total",
				Help:      "Total number of fraud attempts detected",
			},
			[]string{"type"},
		),
		rateLimitHit: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "rate_limit_hit_total",
				Help:      "Total number of rate limit hits",
			},
			[]string{"type"},
		),
		attestationsCreated: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "attestations_created_total",
				Help:      "Total number of SMS verification attestations created",
			},
			[]string{"country_code"},
		),
		attestationsFailed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "attestations_failed_total",
				Help:      "Total number of attestation creation failures",
			},
			[]string{"reason"},
		),
		providerFailover: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "provider_failover_total",
				Help:      "Total number of provider failovers",
			},
			[]string{"from_provider", "to_provider"},
		),
		providerHealth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "provider_health",
				Help:      "Health status of SMS providers (1=healthy, 0=unhealthy)",
			},
			[]string{"provider"},
		),
		carrierLookups: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "carrier_lookups_total",
				Help:      "Total number of carrier lookups performed",
			},
			[]string{"provider", "result"},
		),
		carrierLookupLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "veid",
				Subsystem: "sms",
				Name:      "carrier_lookup_latency_seconds",
				Help:      "Latency of carrier lookups in seconds",
				Buckets:   []float64{0.1, 0.25, 0.5, 1, 2.5, 5},
			},
			[]string{"provider"},
		),
	}
}

// Register registers all metrics with the default Prometheus registry
func (m *Metrics) Register() error {
	var err error
	m.mu.Do(func() {
		collectors := []prometheus.Collector{
			m.challengesCreated,
			m.challengesVerified,
			m.challengesFailed,
			m.challengesExpired,
			m.smsSent,
			m.smsDelivered,
			m.smsFailed,
			m.smsLatency,
			m.otpAttempts,
			m.otpSuccessful,
			m.otpFailed,
			m.otpResends,
			m.verificationLatency,
			m.voipDetected,
			m.phoneBlocked,
			m.ipBlocked,
			m.velocityExceeded,
			m.riskScoreHistogram,
			m.fraudDetected,
			m.rateLimitHit,
			m.attestationsCreated,
			m.attestationsFailed,
			m.providerFailover,
			m.providerHealth,
			m.carrierLookups,
			m.carrierLookupLatency,
		}

		for _, c := range collectors {
			if regErr := prometheus.Register(c); regErr != nil {
				// Ignore already registered errors
				if _, ok := regErr.(prometheus.AlreadyRegisteredError); !ok {
					err = regErr
					return
				}
			}
		}
	})
	return err
}

// RecordChallengeCreated records a challenge creation
func (m *Metrics) RecordChallengeCreated(countryCode string) {
	m.challengesCreated.WithLabelValues(countryCode).Inc()
}

// RecordChallengeVerified records a successful verification
func (m *Metrics) RecordChallengeVerified(countryCode string, latency time.Duration) {
	m.challengesVerified.WithLabelValues(countryCode).Inc()
	m.verificationLatency.WithLabelValues(countryCode).Observe(latency.Seconds())
}

// RecordChallengeFailed records a failed challenge
func (m *Metrics) RecordChallengeFailed(countryCode string, reason string) {
	m.challengesFailed.WithLabelValues(countryCode, reason).Inc()
}

// RecordChallengeExpired records an expired challenge
func (m *Metrics) RecordChallengeExpired(countryCode string) {
	m.challengesExpired.WithLabelValues(countryCode).Inc()
}

// RecordSMSSent records an SMS being sent
func (m *Metrics) RecordSMSSent(provider string, countryCode string, latency time.Duration) {
	m.smsSent.WithLabelValues(provider, countryCode).Inc()
	m.smsLatency.WithLabelValues(provider).Observe(latency.Seconds())
}

// RecordSMSDelivered records an SMS delivery
func (m *Metrics) RecordSMSDelivered(provider string, countryCode string) {
	m.smsDelivered.WithLabelValues(provider, countryCode).Inc()
}

// RecordSMSFailed records a failed SMS delivery
func (m *Metrics) RecordSMSFailed(provider string, errorCode string) {
	m.smsFailed.WithLabelValues(provider, errorCode).Inc()
}

// RecordOTPAttempt records an OTP verification attempt
func (m *Metrics) RecordOTPAttempt(countryCode string, success bool) {
	m.otpAttempts.WithLabelValues(countryCode).Inc()
	if success {
		m.otpSuccessful.WithLabelValues(countryCode).Inc()
	}
}

// RecordOTPFailed records a failed OTP attempt
func (m *Metrics) RecordOTPFailed(countryCode string, reason string) {
	m.otpFailed.WithLabelValues(countryCode, reason).Inc()
}

// RecordOTPResend records an OTP resend
func (m *Metrics) RecordOTPResend(countryCode string) {
	m.otpResends.WithLabelValues(countryCode).Inc()
}

// RecordVoIPDetected records a VoIP number detection
func (m *Metrics) RecordVoIPDetected(countryCode string, carrier string) {
	m.voipDetected.WithLabelValues(countryCode, carrier).Inc()
}

// RecordPhoneBlocked records a phone being blocked
func (m *Metrics) RecordPhoneBlocked(reason string) {
	m.phoneBlocked.WithLabelValues(reason).Inc()
}

// RecordIPBlocked records an IP being blocked
func (m *Metrics) RecordIPBlocked(reason string) {
	m.ipBlocked.WithLabelValues(reason).Inc()
}

// RecordVelocityExceeded records a velocity limit being exceeded
func (m *Metrics) RecordVelocityExceeded(limitType string) {
	m.velocityExceeded.WithLabelValues(limitType).Inc()
}

// RecordRiskScore records a risk score
func (m *Metrics) RecordRiskScore(countryCode string, score uint32) {
	m.riskScoreHistogram.WithLabelValues(countryCode).Observe(float64(score))
}

// RecordFraudDetected records fraud detection
func (m *Metrics) RecordFraudDetected(fraudType string) {
	m.fraudDetected.WithLabelValues(fraudType).Inc()
}

// RecordRateLimitHit records a rate limit hit
func (m *Metrics) RecordRateLimitHit(limitType string) {
	m.rateLimitHit.WithLabelValues(limitType).Inc()
}

// RecordAttestationCreated records an attestation creation
func (m *Metrics) RecordAttestationCreated(countryCode string) {
	m.attestationsCreated.WithLabelValues(countryCode).Inc()
}

// RecordAttestationFailed records a failed attestation creation
func (m *Metrics) RecordAttestationFailed(reason string) {
	m.attestationsFailed.WithLabelValues(reason).Inc()
}

// RecordProviderFailover records a provider failover
func (m *Metrics) RecordProviderFailover(fromProvider string, toProvider string) {
	m.providerFailover.WithLabelValues(fromProvider, toProvider).Inc()
}

// SetProviderHealth sets the health status of a provider
func (m *Metrics) SetProviderHealth(provider string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	m.providerHealth.WithLabelValues(provider).Set(value)
}

// RecordCarrierLookup records a carrier lookup
func (m *Metrics) RecordCarrierLookup(provider string, success bool, latency time.Duration) {
	result := "success"
	if !success {
		result = "failure"
	}
	m.carrierLookups.WithLabelValues(provider, result).Inc()
	m.carrierLookupLatency.WithLabelValues(provider).Observe(latency.Seconds())
}

// RecordChainSubmission records an on-chain verification submission
func (m *Metrics) RecordChainSubmission(countryCode string, success bool) {
	// Use attestation metrics for chain submissions
	if success {
		m.attestationsCreated.WithLabelValues(countryCode).Inc()
	} else {
		m.attestationsFailed.WithLabelValues("chain_submission").Inc()
	}
}
