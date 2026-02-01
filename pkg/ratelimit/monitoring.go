package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
)

// Monitor provides monitoring and alerting for rate limiting
type Monitor struct {
	limiter RateLimiter
	logger  zerolog.Logger
	config  MonitorConfig

	// Prometheus metrics
	metrics *monitorMetrics

	// Alert channels
	alertChan chan Alert
}

// MonitorConfig configures the monitor
type MonitorConfig struct {
	// MetricsInterval is how often to collect metrics
	MetricsInterval time.Duration

	// AlertThresholds defines alert thresholds
	AlertThresholds AlertThresholds

	// EnableAlerts enables alerting
	EnableAlerts bool

	// AlertWebhookURL is the webhook URL for alerts (optional)
	AlertWebhookURL string
}

// AlertThresholds defines alert thresholds
type AlertThresholds struct {
	// BlockedRequestsPerMinute triggers an alert
	BlockedRequestsPerMinute uint64

	// BypassAttemptsPerMinute triggers an alert
	BypassAttemptsPerMinute uint64

	// BannedIdentifiersCount triggers an alert
	BannedIdentifiersCount int

	// LoadPercentage triggers an alert
	LoadPercentage float64

	// TopBlockedIPRequests triggers an alert for a single IP
	TopBlockedIPRequests uint64
}

// Alert represents a monitoring alert
type Alert struct {
	Severity  string    `json:"severity"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// monitorMetrics holds Prometheus metrics
type monitorMetrics struct {
	totalRequests         prometheus.Gauge
	allowedRequests       prometheus.Gauge
	blockedRequests       prometheus.Gauge
	bypassAttempts        prometheus.Gauge
	bannedIdentifiers     prometheus.Gauge
	currentLoad           prometheus.Gauge
	blockedRequestsPerMin prometheus.Counter
	bypassAttemptsPerMin  prometheus.Counter
	alertsTriggered       *prometheus.CounterVec
}

// NewMonitor creates a new rate limiting monitor
func NewMonitor(limiter RateLimiter, logger zerolog.Logger, config MonitorConfig) *Monitor {
	if config.MetricsInterval == 0 {
		config.MetricsInterval = 30 * time.Second
	}

	metrics := &monitorMetrics{
		totalRequests: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "virtengine_ratelimit_total_requests",
			Help: "Total requests processed by rate limiter",
		}),
		allowedRequests: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "virtengine_ratelimit_allowed_requests",
			Help: "Total allowed requests",
		}),
		blockedRequests: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "virtengine_ratelimit_blocked_requests",
			Help: "Total blocked requests",
		}),
		bypassAttempts: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "virtengine_ratelimit_bypass_attempts",
			Help: "Total bypass attempts detected",
		}),
		bannedIdentifiers: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "virtengine_ratelimit_banned_identifiers",
			Help: "Number of currently banned identifiers",
		}),
		currentLoad: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "virtengine_ratelimit_current_load",
			Help: "Current system load percentage (0-100)",
		}),
		blockedRequestsPerMin: promauto.NewCounter(prometheus.CounterOpts{
			Name: "virtengine_ratelimit_blocked_requests_per_minute",
			Help: "Blocked requests per minute",
		}),
		bypassAttemptsPerMin: promauto.NewCounter(prometheus.CounterOpts{
			Name: "virtengine_ratelimit_bypass_attempts_per_minute",
			Help: "Bypass attempts per minute",
		}),
		alertsTriggered: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "virtengine_ratelimit_alerts_triggered",
				Help: "Number of alerts triggered by severity",
			},
			[]string{"severity", "type"},
		),
	}

	return &Monitor{
		limiter:   limiter,
		logger:    logger.With().Str("component", "rate-limit-monitor").Logger(),
		config:    config,
		metrics:   metrics,
		alertChan: make(chan Alert, 100),
	}
}

// Start starts the monitor
func (m *Monitor) Start(ctx context.Context) error {
	ticker := time.NewTicker(m.config.MetricsInterval)
	defer ticker.Stop()

	m.logger.Info().Dur("interval", m.config.MetricsInterval).Msg("starting rate limit monitor")

	// Start alert processor
	go m.processAlerts(ctx)

	for {
		select {
		case <-ctx.Done():
			m.logger.Info().Msg("stopping rate limit monitor")
			return ctx.Err()
		case <-ticker.C:
			if err := m.collectMetrics(ctx); err != nil {
				m.logger.Error().Err(err).Msg("failed to collect metrics")
			}
		}
	}
}

// collectMetrics collects and updates metrics
func (m *Monitor) collectMetrics(ctx context.Context) error {
	metrics, err := m.limiter.GetMetrics(ctx)
	if err != nil {
		return fmt.Errorf("failed to get metrics: %w", err)
	}

	// Update Prometheus metrics
	m.metrics.totalRequests.Set(float64(metrics.TotalRequests))
	m.metrics.allowedRequests.Set(float64(metrics.AllowedRequests))
	m.metrics.blockedRequests.Set(float64(metrics.BlockedRequests))
	m.metrics.bypassAttempts.Set(float64(metrics.BypassAttempts))
	m.metrics.bannedIdentifiers.Set(float64(metrics.BannedIdentifiers))
	m.metrics.currentLoad.Set(metrics.CurrentLoad)

	// Emit Cosmos SDK telemetry
	telemetry.SetGauge(float32(metrics.TotalRequests), "ratelimit", "total_requests")
	telemetry.SetGauge(float32(metrics.AllowedRequests), "ratelimit", "allowed_requests")
	telemetry.SetGauge(float32(metrics.BlockedRequests), "ratelimit", "blocked_requests")
	telemetry.SetGauge(float32(metrics.BypassAttempts), "ratelimit", "bypass_attempts")
	telemetry.SetGauge(float32(metrics.BannedIdentifiers), "ratelimit", "banned_identifiers")
	telemetry.SetGauge(float32(metrics.CurrentLoad), "ratelimit", "current_load")

	// Check alert thresholds
	if m.config.EnableAlerts {
		m.checkAlertThresholds(metrics)
	}

	m.logger.Debug().
		Uint64("total", metrics.TotalRequests).
		Uint64("allowed", metrics.AllowedRequests).
		Uint64("blocked", metrics.BlockedRequests).
		Uint64("bypass_attempts", metrics.BypassAttempts).
		Int("banned", metrics.BannedIdentifiers).
		Float64("load", metrics.CurrentLoad).
		Msg("metrics collected")

	return nil
}

// checkAlertThresholds checks if any alert thresholds are exceeded
func (m *Monitor) checkAlertThresholds(metrics *Metrics) {
	// Check blocked requests per minute
	if metrics.BlockedRequests > m.config.AlertThresholds.BlockedRequestsPerMinute {
		m.sendAlert(Alert{
			Severity:  "warning",
			Title:     "High Rate Limit Blocks",
			Message:   fmt.Sprintf("Blocked requests: %d (threshold: %d)", metrics.BlockedRequests, m.config.AlertThresholds.BlockedRequestsPerMinute),
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"blocked_requests": metrics.BlockedRequests,
				"threshold":        m.config.AlertThresholds.BlockedRequestsPerMinute,
			},
		})
	}

	// Check bypass attempts
	if metrics.BypassAttempts > m.config.AlertThresholds.BypassAttemptsPerMinute {
		m.sendAlert(Alert{
			Severity:  "critical",
			Title:     "Potential DDoS Attack",
			Message:   fmt.Sprintf("Bypass attempts: %d (threshold: %d)", metrics.BypassAttempts, m.config.AlertThresholds.BypassAttemptsPerMinute),
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"bypass_attempts": metrics.BypassAttempts,
				"threshold":       m.config.AlertThresholds.BypassAttemptsPerMinute,
				"top_blocked_ips": metrics.TopBlockedIPs,
			},
		})
	}

	// Check banned identifiers
	if metrics.BannedIdentifiers > m.config.AlertThresholds.BannedIdentifiersCount {
		m.sendAlert(Alert{
			Severity:  "warning",
			Title:     "High Number of Banned Identifiers",
			Message:   fmt.Sprintf("Banned identifiers: %d (threshold: %d)", metrics.BannedIdentifiers, m.config.AlertThresholds.BannedIdentifiersCount),
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"banned_identifiers": metrics.BannedIdentifiers,
				"threshold":          m.config.AlertThresholds.BannedIdentifiersCount,
			},
		})
	}

	// Check system load
	if metrics.CurrentLoad > m.config.AlertThresholds.LoadPercentage {
		m.sendAlert(Alert{
			Severity:  "warning",
			Title:     "High System Load",
			Message:   fmt.Sprintf("System load: %.2f%% (threshold: %.2f%%)", metrics.CurrentLoad, m.config.AlertThresholds.LoadPercentage),
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"current_load": metrics.CurrentLoad,
				"threshold":    m.config.AlertThresholds.LoadPercentage,
			},
		})
	}

	// Check top blocked IPs
	if len(metrics.TopBlockedIPs) > 0 {
		for _, ip := range metrics.TopBlockedIPs[:min(5, len(metrics.TopBlockedIPs))] {
			if count, ok := metrics.ByLimitType[LimitTypeIP]; ok {
				if count.Blocked > m.config.AlertThresholds.TopBlockedIPRequests {
					m.sendAlert(Alert{
						Severity:  "info",
						Title:     "High Block Rate for IP",
						Message:   fmt.Sprintf("IP %s has been blocked %d times", ip, count.Blocked),
						Timestamp: time.Now(),
						Metadata: map[string]interface{}{
							"ip":      ip,
							"blocked": count.Blocked,
						},
					})
				}
			}
		}
	}
}

// sendAlert sends an alert
func (m *Monitor) sendAlert(alert Alert) {
	select {
	case m.alertChan <- alert:
		m.metrics.alertsTriggered.WithLabelValues(alert.Severity, alert.Title).Inc()
	default:
		m.logger.Warn().Msg("alert channel full, dropping alert")
	}
}

// processAlerts processes alerts from the channel
func (m *Monitor) processAlerts(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case alert := <-m.alertChan:
			m.handleAlert(alert)
		}
	}
}

// handleAlert handles an individual alert
func (m *Monitor) handleAlert(alert Alert) {
	// Log the alert
	logEvent := m.logger.WithLevel(m.getSeverityLevel(alert.Severity))
	logEvent.
		Str("title", alert.Title).
		Str("message", alert.Message).
		Interface("metadata", alert.Metadata).
		Msg("ALERT")

	// Emit telemetry event
	telemetry.IncrCounter(1, "ratelimit", "alerts", alert.Severity)

	// Send to webhook if configured
	if m.config.AlertWebhookURL != "" {
		// In production, you would send this to the webhook
		// For now, just log that we would send it
		m.logger.Debug().
			Str("webhook", m.config.AlertWebhookURL).
			Str("alert", alert.Title).
			Msg("would send alert to webhook")
	}
}

// getSeverityLevel converts alert severity to log level
func (m *Monitor) getSeverityLevel(severity string) zerolog.Level {
	switch severity {
	case "critical":
		return zerolog.ErrorLevel
	case "warning":
		return zerolog.WarnLevel
	case "info":
		return zerolog.InfoLevel
	default:
		return zerolog.DebugLevel
	}
}

// GetAlerts returns the alert channel
func (m *Monitor) GetAlerts() <-chan Alert {
	return m.alertChan
}

// DefaultAlertThresholds returns default alert thresholds
func DefaultAlertThresholds() AlertThresholds {
	return AlertThresholds{
		BlockedRequestsPerMinute: 1000,
		BypassAttemptsPerMinute:  100,
		BannedIdentifiersCount:   50,
		LoadPercentage:           80.0,
		TopBlockedIPRequests:     500,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

