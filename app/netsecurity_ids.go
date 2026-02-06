package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/virtengine/virtengine/pkg/security"
)

// IDSIntegration provides integration with Intrusion Detection Systems.
type IDSIntegration struct {
	config IDSConfig
	logger log.Logger

	// Alert channels
	alertChan  chan IDSAlert
	httpClient *http.Client

	// Metrics
	alertsSent    int64
	alertsDropped int64

	// Log file
	logFile *os.File

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu sync.RWMutex
}

// IDSAlert represents an IDS alert.
type IDSAlert struct {
	Timestamp   time.Time              `json:"timestamp"`
	AlertID     string                 `json:"alert_id"`
	Type        string                 `json:"type"`
	Severity    IDSSeverity            `json:"severity"`
	Source      IDSAlertSource         `json:"source"`
	Destination IDSAlertDestination    `json:"destination,omitempty"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Protocol    string                 `json:"protocol,omitempty"`
	Action      string                 `json:"action"` // "allowed", "blocked", "alerted"

	// Signature-based detection
	SignatureID   int    `json:"signature_id,omitempty"`
	SignatureName string `json:"signature_name,omitempty"`

	// Classification
	Classification string `json:"classification,omitempty"`
	Priority       int    `json:"priority,omitempty"`
}

// IDSSeverity represents alert severity levels.
type IDSSeverity string

const (
	IDSSeverityLow      IDSSeverity = "low"
	IDSSeverityMedium   IDSSeverity = "medium"
	IDSSeverityHigh     IDSSeverity = "high"
	IDSSeverityCritical IDSSeverity = "critical"
)

// IDSAlertSource represents the source of an alert.
type IDSAlertSource struct {
	IP   string `json:"ip"`
	Port int    `json:"port,omitempty"`
	ID   string `json:"id,omitempty"` // Peer ID if available
}

// IDSAlertDestination represents the destination of an alert.
type IDSAlertDestination struct {
	IP   string `json:"ip,omitempty"`
	Port int    `json:"port,omitempty"`
}

// Predefined alert types
const (
	AlertTypeDDoS              = "ddos_attack"
	AlertTypeSybil             = "sybil_attack"
	AlertTypeEclipse           = "eclipse_attack"
	AlertTypeRateLimit         = "rate_limit_exceeded"
	AlertTypeMalformedMessage  = "malformed_message"
	AlertTypeUnauthorizedPeer  = "unauthorized_peer"
	AlertTypeProtocolViolation = "protocol_violation"
	AlertTypeAuthFailure       = "auth_failure"
	AlertTypeBruteForce        = "brute_force"
	AlertTypeSuspiciousPattern = "suspicious_pattern"
)

// Signature IDs for common attacks
const (
	SigDDoSConnectionFlood = 1001
	SigDDoSMessageFlood    = 1002
	SigDDoSSYNFlood        = 1003
	SigSybilSubnetAbuse    = 2001
	SigSybilASNAbuse       = 2002
	SigEclipseAttempt      = 3001
	SigRateLimitConnection = 4001
	SigRateLimitMessage    = 4002
	SigRateLimitBandwidth  = 4003
	SigAuthFailure         = 5001
	SigMalformedHandshake  = 6001
	SigMalformedMessage    = 6002
)

// NewIDSIntegration creates a new IDS integration.
func NewIDSIntegration(config IDSConfig, logger log.Logger) (*IDSIntegration, error) {
	if logger == nil {
		logger = log.NewNopLogger()
	}

	ctx, cancel := context.WithCancel(context.Background())

	ids := &IDSIntegration{
		config:     config,
		logger:     logger.With("module", "ids"),
		alertChan:  make(chan IDSAlert, 1000),
		httpClient: security.NewSecureHTTPClient(security.WithTimeout(10 * time.Second)),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Open log file if configured
	if config.LogPath != "" {
		var err error
		ids.logFile, err = os.OpenFile(config.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to open IDS log file: %w", err)
		}
	}

	return ids, nil
}

// Start starts the IDS integration background workers.
func (ids *IDSIntegration) Start() {
	ids.wg.Add(1)
	go ids.alertProcessor()
}

// Stop stops the IDS integration.
func (ids *IDSIntegration) Stop() {
	ids.cancel()
	ids.wg.Wait()

	if ids.logFile != nil {
		ids.logFile.Close()
	}
}

// alertProcessor processes alerts in the background.
func (ids *IDSIntegration) alertProcessor() {
	defer ids.wg.Done()

	for {
		select {
		case <-ids.ctx.Done():
			return
		case alert := <-ids.alertChan:
			ids.processAlert(alert)
		}
	}
}

// processAlert processes a single alert.
func (ids *IDSIntegration) processAlert(alert IDSAlert) {
	// Check severity threshold
	if !ids.meetsThreshold(alert.Severity) {
		return
	}

	// Log to file
	if ids.logFile != nil {
		ids.writeLogEntry(alert)
	}

	// Send to alert endpoint
	if ids.config.AlertEndpoint != "" {
		ids.sendAlertHTTP(alert)
	}

	// Emit metrics
	if ids.config.EnableMetrics {
		telemetry.IncrCounter(1, "ids", "alerts", string(alert.Severity))
		telemetry.IncrCounter(1, "ids", "alerts", "type", alert.Type)
	}

	ids.mu.Lock()
	ids.alertsSent++
	ids.mu.Unlock()

	ids.logger.Info("IDS alert processed",
		"type", alert.Type,
		"severity", alert.Severity,
		"source", alert.Source.IP)
}

// meetsThreshold checks if alert severity meets the configured threshold.
func (ids *IDSIntegration) meetsThreshold(severity IDSSeverity) bool {
	severityOrder := map[IDSSeverity]int{
		IDSSeverityLow:      1,
		IDSSeverityMedium:   2,
		IDSSeverityHigh:     3,
		IDSSeverityCritical: 4,
	}

	threshold := IDSSeverity(ids.config.AlertLevel)
	return severityOrder[severity] >= severityOrder[threshold]
}

// writeLogEntry writes an alert to the log file in Suricata EVE JSON format.
func (ids *IDSIntegration) writeLogEntry(alert IDSAlert) {
	// EVE JSON format compatible with Suricata/ELK
	entry := map[string]interface{}{
		"timestamp":  alert.Timestamp.Format(time.RFC3339Nano),
		"event_type": "alert",
		"src_ip":     alert.Source.IP,
		"src_port":   alert.Source.Port,
		"alert": map[string]interface{}{
			"action":       alert.Action,
			"severity":     severityToInt(alert.Severity),
			"signature":    alert.SignatureName,
			"signature_id": alert.SignatureID,
			"category":     alert.Classification,
			"metadata": map[string]interface{}{
				"type":    alert.Type,
				"message": alert.Message,
			},
		},
		"app_proto": "virtengine",
		"flow_id":   alert.AlertID,
	}

	if alert.Destination.IP != "" {
		entry["dest_ip"] = alert.Destination.IP
		entry["dest_port"] = alert.Destination.Port
	}

	if alert.Protocol != "" {
		entry["proto"] = alert.Protocol
	}

	// Add custom details
	if len(alert.Details) > 0 {
		entry["extra"] = alert.Details
	}

	data, err := json.Marshal(entry)
	if err != nil {
		ids.logger.Error("failed to marshal IDS log entry", "error", err)
		return
	}

	_, _ = ids.logFile.Write(data)
	_, _ = ids.logFile.WriteString("\n")
}

// sendAlertHTTP sends an alert to the configured HTTP endpoint.
func (ids *IDSIntegration) sendAlertHTTP(alert IDSAlert) {
	data, err := json.Marshal(alert)
	if err != nil {
		ids.logger.Error("failed to marshal alert", "error", err)
		return
	}

	req, err := http.NewRequestWithContext(ids.ctx, "POST", ids.config.AlertEndpoint, strings.NewReader(string(data)))
	if err != nil {
		ids.logger.Error("failed to create HTTP request", "error", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "VirtEngine-IDS/1.0")

	resp, err := ids.httpClient.Do(req)
	if err != nil {
		ids.logger.Error("failed to send alert to endpoint", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		ids.logger.Warn("alert endpoint returned error", "status", resp.StatusCode)
	}
}

// SendAlert queues an alert for processing.
func (ids *IDSIntegration) SendAlert(alert IDSAlert) {
	if !ids.config.Enabled {
		return
	}

	// Set defaults
	if alert.Timestamp.IsZero() {
		alert.Timestamp = time.Now()
	}
	if alert.AlertID == "" {
		alert.AlertID = fmt.Sprintf("ve-%d", time.Now().UnixNano())
	}

	select {
	case ids.alertChan <- alert:
		// Alert queued
	default:
		// Queue full, drop alert
		ids.mu.Lock()
		ids.alertsDropped++
		ids.mu.Unlock()
		ids.logger.Warn("IDS alert queue full, dropping alert", "type", alert.Type)
	}
}

// Helper methods for common alert types

// AlertDDoS sends a DDoS attack alert.
func (ids *IDSIntegration) AlertDDoS(attackType string, sourceIP string, count int64, threshold int) {
	ids.SendAlert(IDSAlert{
		Type:           AlertTypeDDoS,
		Severity:       IDSSeverityHigh,
		Source:         IDSAlertSource{IP: sourceIP},
		Message:        fmt.Sprintf("%s detected: %d events (threshold: %d)", attackType, count, threshold),
		Action:         "blocked",
		SignatureID:    getSignatureID(attackType),
		SignatureName:  fmt.Sprintf("VirtEngine %s Detection", attackType),
		Classification: "Network Attack",
		Priority:       1,
		Details: map[string]interface{}{
			"attack_type": attackType,
			"count":       count,
			"threshold":   threshold,
		},
	})
}

// AlertSybil sends a Sybil attack alert.
func (ids *IDSIntegration) AlertSybil(reason string, sourceIP string, subnet string, count int) {
	ids.SendAlert(IDSAlert{
		Type:           AlertTypeSybil,
		Severity:       IDSSeverityMedium,
		Source:         IDSAlertSource{IP: sourceIP},
		Message:        fmt.Sprintf("Sybil attack indicator: %s (%d peers from %s)", reason, count, subnet),
		Action:         "blocked",
		SignatureID:    SigSybilSubnetAbuse,
		SignatureName:  "VirtEngine Sybil Attack Detection",
		Classification: "Attempted Network Manipulation",
		Priority:       2,
		Details: map[string]interface{}{
			"reason":     reason,
			"subnet":     subnet,
			"peer_count": count,
		},
	})
}

// AlertRateLimit sends a rate limit alert.
func (ids *IDSIntegration) AlertRateLimit(limitType string, sourceIP string, current, limit int64) {
	ids.SendAlert(IDSAlert{
		Type:           AlertTypeRateLimit,
		Severity:       IDSSeverityLow,
		Source:         IDSAlertSource{IP: sourceIP},
		Message:        fmt.Sprintf("Rate limit exceeded: %s (%d > %d)", limitType, current, limit),
		Action:         "blocked",
		SignatureID:    SigRateLimitConnection,
		SignatureName:  "VirtEngine Rate Limit Violation",
		Classification: "Potential DoS",
		Priority:       3,
		Details: map[string]interface{}{
			"limit_type": limitType,
			"current":    current,
			"limit":      limit,
		},
	})
}

// AlertAuthFailure sends an authentication failure alert.
func (ids *IDSIntegration) AlertAuthFailure(reason string, sourceIP string, peerID string) {
	ids.SendAlert(IDSAlert{
		Type:           AlertTypeAuthFailure,
		Severity:       IDSSeverityMedium,
		Source:         IDSAlertSource{IP: sourceIP, ID: peerID},
		Message:        fmt.Sprintf("Authentication failed: %s", reason),
		Action:         "blocked",
		SignatureID:    SigAuthFailure,
		SignatureName:  "VirtEngine Auth Failure",
		Classification: "Attempted Access",
		Priority:       2,
		Details: map[string]interface{}{
			"reason":  reason,
			"peer_id": peerID,
		},
	})
}

// AlertMalformedMessage sends a malformed message alert.
func (ids *IDSIntegration) AlertMalformedMessage(messageType string, sourceIP string, details string) {
	ids.SendAlert(IDSAlert{
		Type:           AlertTypeMalformedMessage,
		Severity:       IDSSeverityLow,
		Source:         IDSAlertSource{IP: sourceIP},
		Message:        fmt.Sprintf("Malformed %s message: %s", messageType, details),
		Action:         "blocked",
		SignatureID:    SigMalformedMessage,
		SignatureName:  "VirtEngine Malformed Message",
		Classification: "Protocol Anomaly",
		Priority:       3,
		Details: map[string]interface{}{
			"message_type": messageType,
			"details":      details,
		},
	})
}

// AlertFromEvent converts an AlertEvent to an IDSAlert and sends it.
func (ids *IDSIntegration) AlertFromEvent(event AlertEvent) {
	severity := IDSSeverityMedium
	switch event.Severity {
	case "low":
		severity = IDSSeverityLow
	case "medium":
		severity = IDSSeverityMedium
	case "high":
		severity = IDSSeverityHigh
	case "critical":
		severity = IDSSeverityCritical
	}

	ids.SendAlert(IDSAlert{
		Type:           event.Type,
		Severity:       severity,
		Source:         IDSAlertSource{IP: event.Source},
		Message:        event.Message,
		Timestamp:      event.Timestamp,
		Details:        event.Details,
		Action:         "alerted",
		Classification: "Security Event",
		Priority:       severityToPriority(severity),
	})
}

// GetStats returns IDS statistics.
func (ids *IDSIntegration) GetStats() IDSStats {
	ids.mu.RLock()
	defer ids.mu.RUnlock()

	return IDSStats{
		AlertsSent:    ids.alertsSent,
		AlertsDropped: ids.alertsDropped,
		QueueSize:     len(ids.alertChan),
		QueueCapacity: cap(ids.alertChan),
	}
}

// IDSStats contains IDS statistics.
type IDSStats struct {
	AlertsSent    int64
	AlertsDropped int64
	QueueSize     int
	QueueCapacity int
}

// Helper functions

func severityToInt(severity IDSSeverity) int {
	switch severity {
	case IDSSeverityCritical:
		return 1
	case IDSSeverityHigh:
		return 2
	case IDSSeverityMedium:
		return 3
	case IDSSeverityLow:
		return 4
	default:
		return 3
	}
}

func severityToPriority(severity IDSSeverity) int {
	switch severity {
	case IDSSeverityCritical:
		return 1
	case IDSSeverityHigh:
		return 2
	case IDSSeverityMedium:
		return 3
	case IDSSeverityLow:
		return 4
	default:
		return 3
	}
}

func getSignatureID(attackType string) int {
	switch attackType {
	case "connection_flood":
		return SigDDoSConnectionFlood
	case "message_flood":
		return SigDDoSMessageFlood
	case "syn_flood":
		return SigDDoSSYNFlood
	default:
		return SigDDoSConnectionFlood
	}
}

// CreateSourceFromAddr creates an IDSAlertSource from a net.Addr.
func CreateSourceFromAddr(addr net.Addr) IDSAlertSource {
	if addr == nil {
		return IDSAlertSource{}
	}

	ip := extractIP(addr)
	port := 0

	switch v := addr.(type) {
	case *net.TCPAddr:
		port = v.Port
	case *net.UDPAddr:
		port = v.Port
	}

	return IDSAlertSource{IP: ip, Port: port}
}
