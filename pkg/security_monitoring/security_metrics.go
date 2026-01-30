package security_monitoring

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	namespace = "virtengine"
	subsystem = "security"
)

// SecurityMetrics contains all Prometheus metrics for security monitoring
type SecurityMetrics struct {
	// Transaction security metrics
	TxAnomaliesDetected    *prometheus.CounterVec
	TxVelocityRate         *prometheus.GaugeVec
	TxSuspiciousPatterns   *prometheus.CounterVec
	TxValueAnomalies       prometheus.Counter
	TxReplayAttempts       prometheus.Counter

	// Identity verification fraud metrics
	VEIDFraudIndicators     *prometheus.CounterVec
	VEIDVerificationFailures *prometheus.CounterVec
	VEIDTamperingAttempts   prometheus.Counter
	VEIDReplayAttempts      prometheus.Counter
	VEIDScoreAnomalies      prometheus.Counter
	VEIDBiometricMismatches prometheus.Counter
	VEIDDocumentForgery     prometheus.Counter

	// Rate limiting security metrics
	RateLimitBreaches       *prometheus.CounterVec
	RateLimitBans           *prometheus.CounterVec
	DDoSIndicators          *prometheus.CounterVec
	BruteForceAttempts      *prometheus.CounterVec
	RateLimitBypassAttempts prometheus.Counter

	// Cryptographic security metrics
	CryptoOperationFailures *prometheus.CounterVec
	CryptoWeakEntropy       prometheus.Counter
	CryptoSignatureFailures *prometheus.CounterVec
	CryptoKeyMisuse         *prometheus.CounterVec
	CryptoAlgorithmMisuse   prometheus.Counter

	// Provider daemon security metrics
	ProviderCompromiseIndicators *prometheus.CounterVec
	ProviderKeyCompromise        *prometheus.CounterVec
	ProviderUnauthorizedAccess   prometheus.Counter
	ProviderAnomalousActivity    *prometheus.CounterVec

	// Authentication & authorization metrics
	AuthFailures           *prometheus.CounterVec
	AuthzViolations        *prometheus.CounterVec
	SessionAnomalies       prometheus.Counter
	PrivilegeEscalation    prometheus.Counter

	// Audit metrics
	AuditEventsTotal       *prometheus.CounterVec
	AuditEventsBySeverity  *prometheus.CounterVec
	AlertsTriggered        *prometheus.CounterVec
	IncidentResponseActions *prometheus.CounterVec

	// Overall security health
	SecurityIncidentsActive prometheus.Gauge
	ThreatLevel             prometheus.Gauge
	SecurityScore           prometheus.Gauge
}

var (
	globalMetrics     *SecurityMetrics
	globalMetricsOnce sync.Once
)

// GetSecurityMetrics returns the global security metrics instance
func GetSecurityMetrics() *SecurityMetrics {
	globalMetricsOnce.Do(func() {
		globalMetrics = newSecurityMetrics()
	})
	return globalMetrics
}

// newSecurityMetrics creates and registers all security metrics
func newSecurityMetrics() *SecurityMetrics {
	return &SecurityMetrics{
		// Transaction security metrics
		TxAnomaliesDetected: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "tx_anomalies_detected_total",
				Help:      "Total transaction anomalies detected by type",
			},
			[]string{"type", "severity"},
		),
		TxVelocityRate: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "tx_velocity_rate",
				Help:      "Transaction velocity rate by account type",
			},
			[]string{"account_type"},
		),
		TxSuspiciousPatterns: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "tx_suspicious_patterns_total",
				Help:      "Total suspicious transaction patterns detected",
			},
			[]string{"pattern_type"},
		),
		TxValueAnomalies: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "tx_value_anomalies_total",
				Help:      "Total transaction value anomalies detected",
			},
		),
		TxReplayAttempts: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "tx_replay_attempts_total",
				Help:      "Total transaction replay attempts detected",
			},
		),

		// Identity verification fraud metrics
		VEIDFraudIndicators: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "veid_fraud_indicators_total",
				Help:      "Total VEID fraud indicators detected by type",
			},
			[]string{"indicator_type", "severity"},
		),
		VEIDVerificationFailures: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "veid_verification_failures_total",
				Help:      "Total VEID verification failures by reason",
			},
			[]string{"reason"},
		),
		VEIDTamperingAttempts: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "veid_tampering_attempts_total",
				Help:      "Total VEID tampering attempts detected",
			},
		),
		VEIDReplayAttempts: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "veid_replay_attempts_total",
				Help:      "Total VEID replay attempts detected",
			},
		),
		VEIDScoreAnomalies: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "veid_score_anomalies_total",
				Help:      "Total VEID score anomalies detected",
			},
		),
		VEIDBiometricMismatches: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "veid_biometric_mismatches_total",
				Help:      "Total VEID biometric mismatches detected",
			},
		),
		VEIDDocumentForgery: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "veid_document_forgery_total",
				Help:      "Total VEID document forgery attempts detected",
			},
		),

		// Rate limiting security metrics
		RateLimitBreaches: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "ratelimit_breaches_total",
				Help:      "Total rate limit breaches by type and severity",
			},
			[]string{"limit_type", "severity"},
		),
		RateLimitBans: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "ratelimit_bans_total",
				Help:      "Total bans issued due to rate limit violations",
			},
			[]string{"ban_type"},
		),
		DDoSIndicators: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "ddos_indicators_total",
				Help:      "Total DDoS indicators detected by type",
			},
			[]string{"attack_type"},
		),
		BruteForceAttempts: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "brute_force_attempts_total",
				Help:      "Total brute force attempts detected by target",
			},
			[]string{"target"},
		),
		RateLimitBypassAttempts: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "ratelimit_bypass_attempts_total",
				Help:      "Total rate limit bypass attempts detected",
			},
		),

		// Cryptographic security metrics
		CryptoOperationFailures: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "crypto_operation_failures_total",
				Help:      "Total cryptographic operation failures by operation type",
			},
			[]string{"operation", "reason"},
		),
		CryptoWeakEntropy: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "crypto_weak_entropy_total",
				Help:      "Total weak entropy detections in cryptographic operations",
			},
		),
		CryptoSignatureFailures: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "crypto_signature_failures_total",
				Help:      "Total signature verification failures by type",
			},
			[]string{"signature_type", "reason"},
		),
		CryptoKeyMisuse: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "crypto_key_misuse_total",
				Help:      "Total key misuse incidents by key type",
			},
			[]string{"key_type", "misuse_type"},
		),
		CryptoAlgorithmMisuse: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "crypto_algorithm_misuse_total",
				Help:      "Total algorithm misuse incidents",
			},
		),

		// Provider daemon security metrics
		ProviderCompromiseIndicators: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "provider_compromise_indicators_total",
				Help:      "Total provider compromise indicators by type",
			},
			[]string{"indicator_type", "severity"},
		),
		ProviderKeyCompromise: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "provider_key_compromise_total",
				Help:      "Total provider key compromise events by key type",
			},
			[]string{"key_type"},
		),
		ProviderUnauthorizedAccess: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "provider_unauthorized_access_total",
				Help:      "Total unauthorized access attempts to provider",
			},
		),
		ProviderAnomalousActivity: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "provider_anomalous_activity_total",
				Help:      "Total anomalous provider activity by type",
			},
			[]string{"activity_type"},
		),

		// Authentication & authorization metrics
		AuthFailures: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "auth_failures_total",
				Help:      "Total authentication failures by reason",
			},
			[]string{"reason", "source"},
		),
		AuthzViolations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "authz_violations_total",
				Help:      "Total authorization violations by resource type",
			},
			[]string{"resource_type", "action"},
		),
		SessionAnomalies: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "session_anomalies_total",
				Help:      "Total session anomalies detected",
			},
		),
		PrivilegeEscalation: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "privilege_escalation_total",
				Help:      "Total privilege escalation attempts detected",
			},
		),

		// Audit metrics
		AuditEventsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "audit_events_total",
				Help:      "Total audit events logged by type",
			},
			[]string{"event_type"},
		),
		AuditEventsBySeverity: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "audit_events_by_severity_total",
				Help:      "Total audit events by severity level",
			},
			[]string{"severity"},
		),
		AlertsTriggered: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "alerts_triggered_total",
				Help:      "Total security alerts triggered by type",
			},
			[]string{"alert_type", "severity"},
		),
		IncidentResponseActions: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "incident_response_actions_total",
				Help:      "Total incident response actions taken",
			},
			[]string{"action_type", "playbook"},
		),

		// Overall security health
		SecurityIncidentsActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "incidents_active",
				Help:      "Number of active security incidents",
			},
		),
		ThreatLevel: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "threat_level",
				Help:      "Current threat level (0=low, 1=medium, 2=high, 3=critical)",
			},
		),
		SecurityScore: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "security_score",
				Help:      "Overall security score (0-100, higher is better)",
			},
		),
	}
}

// SecurityEventSeverity represents the severity of a security event
type SecurityEventSeverity string

const (
	SeverityInfo     SecurityEventSeverity = "info"
	SeverityLow      SecurityEventSeverity = "low"
	SeverityMedium   SecurityEventSeverity = "medium"
	SeverityHigh     SecurityEventSeverity = "high"
	SeverityCritical SecurityEventSeverity = "critical"
)

// ThreatLevelValue maps threat level to numeric value
func ThreatLevelValue(severity SecurityEventSeverity) float64 {
	switch severity {
	case SeverityLow:
		return 0
	case SeverityMedium:
		return 1
	case SeverityHigh:
		return 2
	case SeverityCritical:
		return 3
	default:
		return 0
	}
}
