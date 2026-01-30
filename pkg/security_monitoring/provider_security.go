package security_monitoring

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// ProviderSecurityConfig configures provider security monitoring
type ProviderSecurityConfig struct {
	// Key compromise detection
	MaxKeyUsagePerMinute        int `json:"max_key_usage_per_minute"`
	MaxFailedSignaturesPerHour  int `json:"max_failed_signatures_per_hour"`
	AnomalousTimeWindowStart    int `json:"anomalous_time_window_start"` // hour
	AnomalousTimeWindowEnd      int `json:"anomalous_time_window_end"`   // hour

	// Activity thresholds
	MaxBidsPerMinute           int `json:"max_bids_per_minute"`
	MaxLeasesPerHour           int `json:"max_leases_per_hour"`
	MaxDeploymentsPerHour      int `json:"max_deployments_per_hour"`

	// Location tracking
	EnableLocationTracking     bool     `json:"enable_location_tracking"`
	AllowedIPRanges            []string `json:"allowed_ip_ranges,omitempty"`

	// Resource anomaly detection
	EnableResourceAnomalyDetection bool    `json:"enable_resource_anomaly_detection"`
	ResourceUsageVarianceThreshold float64 `json:"resource_usage_variance_threshold"`
}

// DefaultProviderSecurityConfig returns default configuration
func DefaultProviderSecurityConfig() *ProviderSecurityConfig {
	return &ProviderSecurityConfig{
		MaxKeyUsagePerMinute:       20,
		MaxFailedSignaturesPerHour: 10,
		AnomalousTimeWindowStart:   6,
		AnomalousTimeWindowEnd:     22,

		MaxBidsPerMinute:      50,
		MaxLeasesPerHour:      100,
		MaxDeploymentsPerHour: 50,

		EnableLocationTracking: true,

		EnableResourceAnomalyDetection: true,
		ResourceUsageVarianceThreshold: 0.5,
	}
}

// ProviderActivityData represents provider daemon activity
type ProviderActivityData struct {
	ProviderID      string                 `json:"provider_id"`
	ActivityType    string                 `json:"activity_type"` // bid, lease, deploy, key_usage, etc.
	Timestamp       time.Time              `json:"timestamp"`
	SourceIP        string                 `json:"source_ip,omitempty"`
	KeyID           string                 `json:"key_id,omitempty"`
	KeyFingerprint  string                 `json:"key_fingerprint,omitempty"`
	Success         bool                   `json:"success"`
	FailureReason   string                 `json:"failure_reason,omitempty"`

	// Activity-specific data
	LeaseID         string                 `json:"lease_id,omitempty"`
	OrderID         string                 `json:"order_id,omitempty"`
	DeploymentID    string                 `json:"deployment_id,omitempty"`
	BidAmount       float64                `json:"bid_amount,omitempty"`

	// Resource metrics
	CPUUsage        float64                `json:"cpu_usage,omitempty"`
	MemoryUsage     float64                `json:"memory_usage,omitempty"`
	StorageUsage    float64                `json:"storage_usage,omitempty"`
	NetworkUsage    float64                `json:"network_usage,omitempty"`

	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ProviderSecurityIndicator represents types of provider security indicators
type ProviderSecurityIndicator string

const (
	ProviderIndicatorKeyCompromise      ProviderSecurityIndicator = "key_compromise"
	ProviderIndicatorUnauthorizedAccess ProviderSecurityIndicator = "unauthorized_access"
	ProviderIndicatorAnomalousLocation  ProviderSecurityIndicator = "anomalous_location"
	ProviderIndicatorAnomalousTime      ProviderSecurityIndicator = "anomalous_time"
	ProviderIndicatorRapidActivity      ProviderSecurityIndicator = "rapid_activity"
	ProviderIndicatorResourceAnomaly    ProviderSecurityIndicator = "resource_anomaly"
	ProviderIndicatorBidManipulation    ProviderSecurityIndicator = "bid_manipulation"
	ProviderIndicatorLeaseAbuse         ProviderSecurityIndicator = "lease_abuse"
)

// ProviderSecurityMonitor monitors provider daemon security
type ProviderSecurityMonitor struct {
	config  *ProviderSecurityConfig
	logger  zerolog.Logger
	metrics *SecurityMetrics

	// State tracking
	providerActivity   map[string][]providerActivityRecord
	providerKeyUsage   map[string][]keyUsageRecord
	providerLocations  map[string][]locationRecord
	resourceBaselines  map[string]resourceBaseline
	mu                 sync.RWMutex

	// Event channel
	eventChan chan<- *SecurityEvent
	ctx       context.Context
}

type providerActivityRecord struct {
	activityType string
	timestamp    time.Time
	success      bool
	sourceIP     string
}

type keyUsageRecord struct {
	keyID     string
	timestamp time.Time
	success   bool
	sourceIP  string
}

type locationRecord struct {
	sourceIP  string
	timestamp time.Time
}

type resourceBaseline struct {
	avgCPU      float64
	avgMemory   float64
	avgStorage  float64
	sampleCount int
}

// NewProviderSecurityMonitor creates a new provider security monitor
func NewProviderSecurityMonitor(config *ProviderSecurityConfig, logger zerolog.Logger) *ProviderSecurityMonitor {
	if config == nil {
		config = DefaultProviderSecurityConfig()
	}

	return &ProviderSecurityMonitor{
		config:            config,
		logger:            logger.With().Str("monitor", "provider-security").Logger(),
		metrics:           GetSecurityMetrics(),
		providerActivity:  make(map[string][]providerActivityRecord),
		providerKeyUsage:  make(map[string][]keyUsageRecord),
		providerLocations: make(map[string][]locationRecord),
		resourceBaselines: make(map[string]resourceBaseline),
	}
}

// Start starts the monitor
func (m *ProviderSecurityMonitor) Start(ctx context.Context, eventChan chan<- *SecurityEvent) {
	m.ctx = ctx
	m.eventChan = eventChan

	// Start cleanup goroutine
	go m.cleanup(ctx)
}

// Analyze analyzes provider activity for security issues
func (m *ProviderSecurityMonitor) Analyze(activity *ProviderActivityData) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Record activity
	record := providerActivityRecord{
		activityType: activity.ActivityType,
		timestamp:    activity.Timestamp,
		success:      activity.Success,
		sourceIP:     activity.SourceIP,
	}

	if _, exists := m.providerActivity[activity.ProviderID]; !exists {
		m.providerActivity[activity.ProviderID] = make([]providerActivityRecord, 0)
	}
	m.providerActivity[activity.ProviderID] = append(
		m.providerActivity[activity.ProviderID], record)

	// Record key usage if applicable
	if activity.KeyID != "" {
		if _, exists := m.providerKeyUsage[activity.ProviderID]; !exists {
			m.providerKeyUsage[activity.ProviderID] = make([]keyUsageRecord, 0)
		}
		m.providerKeyUsage[activity.ProviderID] = append(
			m.providerKeyUsage[activity.ProviderID], keyUsageRecord{
				keyID:     activity.KeyID,
				timestamp: activity.Timestamp,
				success:   activity.Success,
				sourceIP:  activity.SourceIP,
			})
	}

	// Run all checks
	m.checkKeyUsageVelocity(activity)
	m.checkKeyFailures(activity)
	m.checkActivityVelocity(activity)
	m.checkLocationAnomaly(activity)
	m.checkTimeAnomaly(activity)
	m.checkResourceAnomaly(activity)
	m.checkBidManipulation(activity)
}

// checkKeyUsageVelocity checks for rapid key usage
func (m *ProviderSecurityMonitor) checkKeyUsageVelocity(activity *ProviderActivityData) {
	if activity.KeyID == "" {
		return
	}

	keyUsage := m.providerKeyUsage[activity.ProviderID]
	minuteAgo := activity.Timestamp.Add(-1 * time.Minute)

	var recentUsage int
	for _, rec := range keyUsage {
		if rec.timestamp.After(minuteAgo) {
			recentUsage++
		}
	}

	if recentUsage > m.config.MaxKeyUsagePerMinute {
		m.metrics.ProviderCompromiseIndicators.WithLabelValues(
			string(ProviderIndicatorKeyCompromise), "high").Inc()
		m.emitEvent(&SecurityEvent{
			ID:          generateEventID(),
			Type:        string(ProviderIndicatorKeyCompromise),
			Severity:    SeverityHigh,
			Timestamp:   activity.Timestamp,
			Source:      activity.ProviderID,
			Description: "Rapid key usage detected - potential key compromise",
			Metadata: map[string]interface{}{
				"provider_id":  activity.ProviderID,
				"key_id":       activity.KeyID,
				"usage_count":  recentUsage,
				"threshold":    m.config.MaxKeyUsagePerMinute,
				"window":       "1m",
			},
		})
	}
}

// checkKeyFailures checks for signature verification failures
func (m *ProviderSecurityMonitor) checkKeyFailures(activity *ProviderActivityData) {
	if activity.Success {
		return
	}

	keyUsage := m.providerKeyUsage[activity.ProviderID]
	hourAgo := activity.Timestamp.Add(-1 * time.Hour)

	var failureCount int
	for _, rec := range keyUsage {
		if !rec.success && rec.timestamp.After(hourAgo) {
			failureCount++
		}
	}

	if failureCount >= m.config.MaxFailedSignaturesPerHour {
		m.metrics.ProviderKeyCompromise.WithLabelValues("signature").Inc()
		m.emitEvent(&SecurityEvent{
			ID:          generateEventID(),
			Type:        string(ProviderIndicatorKeyCompromise),
			Severity:    SeverityHigh,
			Timestamp:   activity.Timestamp,
			Source:      activity.ProviderID,
			Description: "Multiple key operation failures - potential compromise",
			Metadata: map[string]interface{}{
				"provider_id":    activity.ProviderID,
				"failure_count":  failureCount,
				"threshold":      m.config.MaxFailedSignaturesPerHour,
				"window":         "1h",
				"last_reason":    activity.FailureReason,
			},
		})
	}
}

// checkActivityVelocity checks for rapid activity
func (m *ProviderSecurityMonitor) checkActivityVelocity(activity *ProviderActivityData) {
	history := m.providerActivity[activity.ProviderID]
	minuteAgo := activity.Timestamp.Add(-1 * time.Minute)
	hourAgo := activity.Timestamp.Add(-1 * time.Hour)

	var bidsPerMinute, leasesPerHour, deploymentsPerHour int

	for _, rec := range history {
		if rec.timestamp.After(minuteAgo) && rec.activityType == "bid" {
			bidsPerMinute++
		}
		if rec.timestamp.After(hourAgo) {
			switch rec.activityType {
			case "lease":
				leasesPerHour++
			case "deploy":
				deploymentsPerHour++
			}
		}
	}

	// Check bid velocity
	if bidsPerMinute > m.config.MaxBidsPerMinute {
		m.metrics.ProviderAnomalousActivity.WithLabelValues("rapid_bidding").Inc()
		m.emitEvent(&SecurityEvent{
			ID:          generateEventID(),
			Type:        string(ProviderIndicatorRapidActivity),
			Severity:    SeverityMedium,
			Timestamp:   activity.Timestamp,
			Source:      activity.ProviderID,
			Description: "Rapid bidding activity detected",
			Metadata: map[string]interface{}{
				"provider_id": activity.ProviderID,
				"bid_count":   bidsPerMinute,
				"threshold":   m.config.MaxBidsPerMinute,
				"window":      "1m",
			},
		})
	}

	// Check lease velocity
	if leasesPerHour > m.config.MaxLeasesPerHour {
		m.metrics.ProviderAnomalousActivity.WithLabelValues("rapid_leasing").Inc()
		m.emitEvent(&SecurityEvent{
			ID:          generateEventID(),
			Type:        string(ProviderIndicatorLeaseAbuse),
			Severity:    SeverityMedium,
			Timestamp:   activity.Timestamp,
			Source:      activity.ProviderID,
			Description: "Excessive lease activity detected",
			Metadata: map[string]interface{}{
				"provider_id": activity.ProviderID,
				"lease_count": leasesPerHour,
				"threshold":   m.config.MaxLeasesPerHour,
				"window":      "1h",
			},
		})
	}

	// Check deployment velocity
	if deploymentsPerHour > m.config.MaxDeploymentsPerHour {
		m.metrics.ProviderAnomalousActivity.WithLabelValues("rapid_deployment").Inc()
		m.emitEvent(&SecurityEvent{
			ID:          generateEventID(),
			Type:        string(ProviderIndicatorRapidActivity),
			Severity:    SeverityMedium,
			Timestamp:   activity.Timestamp,
			Source:      activity.ProviderID,
			Description: "Excessive deployment activity detected",
			Metadata: map[string]interface{}{
				"provider_id":      activity.ProviderID,
				"deployment_count": deploymentsPerHour,
				"threshold":        m.config.MaxDeploymentsPerHour,
				"window":           "1h",
			},
		})
	}
}

// checkLocationAnomaly checks for activity from unexpected locations
func (m *ProviderSecurityMonitor) checkLocationAnomaly(activity *ProviderActivityData) {
	if !m.config.EnableLocationTracking || activity.SourceIP == "" {
		return
	}

	// Record location
	if _, exists := m.providerLocations[activity.ProviderID]; !exists {
		m.providerLocations[activity.ProviderID] = make([]locationRecord, 0)
	}
	m.providerLocations[activity.ProviderID] = append(
		m.providerLocations[activity.ProviderID], locationRecord{
			sourceIP:  activity.SourceIP,
			timestamp: activity.Timestamp,
		})

	// Check if IP is in allowed ranges
	if len(m.config.AllowedIPRanges) > 0 {
		allowed := false
		for _, ipRange := range m.config.AllowedIPRanges {
			if matchesIPRange(activity.SourceIP, ipRange) {
				allowed = true
				break
			}
		}
		if !allowed {
			m.metrics.ProviderCompromiseIndicators.WithLabelValues(
				string(ProviderIndicatorAnomalousLocation), "high").Inc()
			m.emitEvent(&SecurityEvent{
				ID:          generateEventID(),
				Type:        string(ProviderIndicatorAnomalousLocation),
				Severity:    SeverityHigh,
				Timestamp:   activity.Timestamp,
				Source:      activity.ProviderID,
				Description: "Provider activity from unauthorized location",
				Metadata: map[string]interface{}{
					"provider_id": activity.ProviderID,
					"source_ip":   activity.SourceIP,
					"allowed_ips": m.config.AllowedIPRanges,
				},
			})
		}
	}

	// Check for rapid location changes
	locations := m.providerLocations[activity.ProviderID]
	if len(locations) >= 2 {
		prevLocation := locations[len(locations)-2]
		timeDiff := activity.Timestamp.Sub(prevLocation.timestamp)
		if timeDiff < 5*time.Minute && prevLocation.sourceIP != activity.SourceIP {
			m.metrics.ProviderCompromiseIndicators.WithLabelValues(
				string(ProviderIndicatorAnomalousLocation), "medium").Inc()
			m.emitEvent(&SecurityEvent{
				ID:          generateEventID(),
				Type:        string(ProviderIndicatorAnomalousLocation),
				Severity:    SeverityMedium,
				Timestamp:   activity.Timestamp,
				Source:      activity.ProviderID,
				Description: "Rapid location change detected",
				Metadata: map[string]interface{}{
					"provider_id":    activity.ProviderID,
					"previous_ip":    prevLocation.sourceIP,
					"current_ip":     activity.SourceIP,
					"time_diff_secs": timeDiff.Seconds(),
				},
			})
		}
	}
}

// checkTimeAnomaly checks for activity outside normal hours
func (m *ProviderSecurityMonitor) checkTimeAnomaly(activity *ProviderActivityData) {
	hour := activity.Timestamp.Hour()

	if hour < m.config.AnomalousTimeWindowStart || hour > m.config.AnomalousTimeWindowEnd {
		m.metrics.ProviderCompromiseIndicators.WithLabelValues(
			string(ProviderIndicatorAnomalousTime), "low").Inc()
		m.emitEvent(&SecurityEvent{
			ID:          generateEventID(),
			Type:        string(ProviderIndicatorAnomalousTime),
			Severity:    SeverityLow,
			Timestamp:   activity.Timestamp,
			Source:      activity.ProviderID,
			Description: "Provider activity outside normal hours",
			Metadata: map[string]interface{}{
				"provider_id":    activity.ProviderID,
				"activity_hour":  hour,
				"expected_start": m.config.AnomalousTimeWindowStart,
				"expected_end":   m.config.AnomalousTimeWindowEnd,
			},
		})
	}
}

// checkResourceAnomaly checks for resource usage anomalies
func (m *ProviderSecurityMonitor) checkResourceAnomaly(activity *ProviderActivityData) {
	if !m.config.EnableResourceAnomalyDetection {
		return
	}

	// Skip if no resource data
	if activity.CPUUsage == 0 && activity.MemoryUsage == 0 {
		return
	}

	baseline, exists := m.resourceBaselines[activity.ProviderID]
	if !exists {
		// Initialize baseline
		m.resourceBaselines[activity.ProviderID] = resourceBaseline{
			avgCPU:      activity.CPUUsage,
			avgMemory:   activity.MemoryUsage,
			avgStorage:  activity.StorageUsage,
			sampleCount: 1,
		}
		return
	}

	// Update baseline with exponential moving average
	alpha := 0.1
	baseline.avgCPU = alpha*activity.CPUUsage + (1-alpha)*baseline.avgCPU
	baseline.avgMemory = alpha*activity.MemoryUsage + (1-alpha)*baseline.avgMemory
	baseline.avgStorage = alpha*activity.StorageUsage + (1-alpha)*baseline.avgStorage
	baseline.sampleCount++
	m.resourceBaselines[activity.ProviderID] = baseline

	// Check for anomalies after establishing baseline
	if baseline.sampleCount < 10 {
		return
	}

	// Check CPU anomaly
	if baseline.avgCPU > 0 {
		cpuVariance := (activity.CPUUsage - baseline.avgCPU) / baseline.avgCPU
		if cpuVariance > m.config.ResourceUsageVarianceThreshold {
			m.metrics.ProviderAnomalousActivity.WithLabelValues("resource_cpu").Inc()
			m.emitEvent(&SecurityEvent{
				ID:          generateEventID(),
				Type:        string(ProviderIndicatorResourceAnomaly),
				Severity:    SeverityMedium,
				Timestamp:   activity.Timestamp,
				Source:      activity.ProviderID,
				Description: "Anomalous CPU usage detected",
				Metadata: map[string]interface{}{
					"provider_id":    activity.ProviderID,
					"current_cpu":    activity.CPUUsage,
					"baseline_cpu":   baseline.avgCPU,
					"variance":       cpuVariance,
					"threshold":      m.config.ResourceUsageVarianceThreshold,
				},
			})
		}
	}

	// Check memory anomaly
	if baseline.avgMemory > 0 {
		memVariance := (activity.MemoryUsage - baseline.avgMemory) / baseline.avgMemory
		if memVariance > m.config.ResourceUsageVarianceThreshold {
			m.metrics.ProviderAnomalousActivity.WithLabelValues("resource_memory").Inc()
			m.emitEvent(&SecurityEvent{
				ID:          generateEventID(),
				Type:        string(ProviderIndicatorResourceAnomaly),
				Severity:    SeverityMedium,
				Timestamp:   activity.Timestamp,
				Source:      activity.ProviderID,
				Description: "Anomalous memory usage detected",
				Metadata: map[string]interface{}{
					"provider_id":      activity.ProviderID,
					"current_memory":   activity.MemoryUsage,
					"baseline_memory":  baseline.avgMemory,
					"variance":         memVariance,
					"threshold":        m.config.ResourceUsageVarianceThreshold,
				},
			})
		}
	}
}

// checkBidManipulation checks for bid manipulation patterns
func (m *ProviderSecurityMonitor) checkBidManipulation(activity *ProviderActivityData) {
	if activity.ActivityType != "bid" {
		return
	}

	history := m.providerActivity[activity.ProviderID]

	// Check for bid sniping (rapid bid at last moment)
	// This would need order deadline info which we don't have here
	// For now, just check for bid patterns

	// Check for unusually low bids
	if activity.BidAmount > 0 && activity.BidAmount < 0.01 {
		m.metrics.ProviderAnomalousActivity.WithLabelValues("low_bid").Inc()
		m.emitEvent(&SecurityEvent{
			ID:          generateEventID(),
			Type:        string(ProviderIndicatorBidManipulation),
			Severity:    SeverityLow,
			Timestamp:   activity.Timestamp,
			Source:      activity.ProviderID,
			Description: "Unusually low bid amount detected",
			Metadata: map[string]interface{}{
				"provider_id": activity.ProviderID,
				"bid_amount":  activity.BidAmount,
				"order_id":    activity.OrderID,
			},
		})
	}

	_ = history // Placeholder for more sophisticated analysis
}

// emitEvent sends an event to the security monitor
func (m *ProviderSecurityMonitor) emitEvent(event *SecurityEvent) {
	if m.eventChan == nil {
		return
	}

	select {
	case m.eventChan <- event:
	default:
		m.logger.Warn().Str("event_id", event.ID).Msg("event channel full, dropping event")
	}
}

// matchesIPRange checks if an IP matches a range (simplified)
func matchesIPRange(ip, ipRange string) bool {
	// Simple exact match - production would use proper CIDR matching
	return ip == ipRange
}

// cleanup periodically cleans up old history
func (m *ProviderSecurityMonitor) cleanup(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.mu.Lock()
			cutoff := time.Now().Add(-2 * time.Hour)

			// Clean up activity history
			for provider, history := range m.providerActivity {
				newHistory := make([]providerActivityRecord, 0)
				for _, rec := range history {
					if rec.timestamp.After(cutoff) {
						newHistory = append(newHistory, rec)
					}
				}
				if len(newHistory) == 0 {
					delete(m.providerActivity, provider)
				} else {
					m.providerActivity[provider] = newHistory
				}
			}

			// Clean up key usage history
			for provider, history := range m.providerKeyUsage {
				newHistory := make([]keyUsageRecord, 0)
				for _, rec := range history {
					if rec.timestamp.After(cutoff) {
						newHistory = append(newHistory, rec)
					}
				}
				if len(newHistory) == 0 {
					delete(m.providerKeyUsage, provider)
				} else {
					m.providerKeyUsage[provider] = newHistory
				}
			}

			// Clean up location history
			for provider, history := range m.providerLocations {
				newHistory := make([]locationRecord, 0)
				for _, rec := range history {
					if rec.timestamp.After(cutoff) {
						newHistory = append(newHistory, rec)
					}
				}
				if len(newHistory) == 0 {
					delete(m.providerLocations, provider)
				} else {
					m.providerLocations[provider] = newHistory
				}
			}

			m.mu.Unlock()
		}
	}
}
