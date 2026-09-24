// Package provider_daemon implements the VirtEngine provider daemon services.
//
// VE-500: HPC heartbeat monitor for gap detection and alerting.
package provider_daemon

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Node status constants
const (
	statusHealthy = "healthy"
	statusStale   = "stale"
	statusOffline = "offline"
)

// HPCHeartbeatMonitorConfig contains configuration for the heartbeat monitor
type HPCHeartbeatMonitorConfig struct {
	// CheckInterval is how often to check for heartbeat gaps
	CheckInterval time.Duration

	// StaleThreshold is when a node is considered stale
	StaleThreshold time.Duration

	// OfflineThreshold is when a node is considered offline
	OfflineThreshold time.Duration

	// AlertHandler is called when alerts are generated
	AlertHandler func(alert HeartbeatAlert)
}

// DefaultHPCHeartbeatMonitorConfig returns the default configuration
func DefaultHPCHeartbeatMonitorConfig() HPCHeartbeatMonitorConfig {
	return HPCHeartbeatMonitorConfig{
		CheckInterval:    30 * time.Second,
		StaleThreshold:   120 * time.Second,
		OfflineThreshold: 300 * time.Second,
		AlertHandler:     defaultAlertHandler,
	}
}

// HeartbeatAlert represents an alert for heartbeat issues
type HeartbeatAlert struct {
	NodeID      string                 `json:"node_id"`
	ClusterID   string                 `json:"cluster_id"`
	AlertType   string                 `json:"alert_type"` // stale, offline, recovered, anomaly
	Severity    string                 `json:"severity"`   // info, warning, critical
	Message     string                 `json:"message"`
	Timestamp   time.Time              `json:"timestamp"`
	LastSeen    time.Time              `json:"last_seen,omitempty"`
	MissedCount int                    `json:"missed_count,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// NodeHeartbeatState tracks heartbeat state for a node
type NodeHeartbeatState struct {
	NodeID           string
	ClusterID        string
	LastHeartbeat    time.Time
	LastSequence     uint64
	ExpectedInterval time.Duration
	Status           string // healthy, stale, offline
	MissedCount      int
	RecoveryCount    int
	AnomalyScore     float64
	HeartbeatHistory []time.Time
	IntervalHistory  []time.Duration
	AlertSent        map[string]time.Time
}

// HPCHeartbeatMonitor monitors node heartbeats for gaps and anomalies
type HPCHeartbeatMonitor struct {
	config  HPCHeartbeatMonitorConfig
	nodes   map[string]*NodeHeartbeatState
	nodesMu sync.RWMutex
	stopCh  chan struct{}
	wg      sync.WaitGroup

	// pendingAlerts holds alerts raised while m.nodesMu is held. They are
	// drained by takePendingAlertsLocked and dispatched by dispatchAlerts once
	// m.nodesMu has been released, so the AlertHandler never runs under the
	// state lock. Guarded by m.nodesMu.
	pendingAlerts []HeartbeatAlert

	// alertMu serializes AlertHandler invocations so a handler is never called
	// concurrently, preserving the ordering guarantee the lock previously gave.
	// It is never held at the same time as m.nodesMu.
	alertMu sync.Mutex

	// Metrics
	totalNodes   int
	healthyNodes int
	staleNodes   int
	offlineNodes int
	alertsRaised int
	metricsMu    sync.RWMutex
}

// NewHPCHeartbeatMonitor creates a new heartbeat monitor
func NewHPCHeartbeatMonitor(config HPCHeartbeatMonitorConfig) *HPCHeartbeatMonitor {
	return &HPCHeartbeatMonitor{
		config: config,
		nodes:  make(map[string]*NodeHeartbeatState),
		stopCh: make(chan struct{}),
	}
}

// Start begins monitoring
func (m *HPCHeartbeatMonitor) Start(ctx context.Context) error {
	m.wg.Add(1)
	go m.monitorLoop(ctx)
	return nil
}

// Stop halts monitoring
func (m *HPCHeartbeatMonitor) Stop() {
	close(m.stopCh)
	m.wg.Wait()
}

// RecordHeartbeat records a heartbeat for a node
func (m *HPCHeartbeatMonitor) RecordHeartbeat(nodeID, clusterID string, sequence uint64) {
	m.nodesMu.Lock()
	m.recordHeartbeatLocked(nodeID, clusterID, sequence)
	m.nodesMu.Unlock()

	// Alerts queued while the state lock was held are dispatched only now.
	// The AlertHandler must never run under m.nodesMu: any handler that reads
	// monitor state would otherwise self-deadlock (Go mutexes are not
	// reentrant).
	m.dispatchPendingAlerts()
}

// recordHeartbeatLocked updates heartbeat state. It must be called with
// m.nodesMu held and must not invoke the AlertHandler directly — it queues
// alerts via raiseAlertLocked for the caller to dispatch after unlocking.
func (m *HPCHeartbeatMonitor) recordHeartbeatLocked(nodeID, clusterID string, sequence uint64) {
	now := time.Now()
	state, exists := m.nodes[nodeID]

	if !exists {
		state = &NodeHeartbeatState{
			NodeID:           nodeID,
			ClusterID:        clusterID,
			ExpectedInterval: 30 * time.Second,
			Status:           "healthy",
			HeartbeatHistory: make([]time.Time, 0, 100),
			IntervalHistory:  make([]time.Duration, 0, 100),
			AlertSent:        make(map[string]time.Time),
		}
		m.nodes[nodeID] = state
	}

	// Calculate interval since last heartbeat
	if !state.LastHeartbeat.IsZero() {
		interval := now.Sub(state.LastHeartbeat)
		state.IntervalHistory = append(state.IntervalHistory, interval)

		// Keep last 100 intervals
		if len(state.IntervalHistory) > 100 {
			state.IntervalHistory = state.IntervalHistory[1:]
		}

		// Check for sequence gap
		if sequence > state.LastSequence+1 {
			gap := sequence - state.LastSequence - 1
			m.raiseAlertLocked(HeartbeatAlert{
				NodeID:      nodeID,
				ClusterID:   clusterID,
				AlertType:   "sequence_gap",
				Severity:    "warning",
				Message:     fmt.Sprintf("Detected sequence gap of %d heartbeats", gap),
				Timestamp:   now,
				MissedCount: safeIntFromUint64(gap),
			})
		}

		// Calculate anomaly score based on interval variance
		state.AnomalyScore = m.calculateAnomalyScore(state.IntervalHistory, interval)
		if state.AnomalyScore > 3.0 {
			m.raiseAlertLocked(HeartbeatAlert{
				NodeID:    nodeID,
				ClusterID: clusterID,
				AlertType: "anomaly",
				Severity:  "info",
				Message:   fmt.Sprintf("Heartbeat interval anomaly detected (score: %.2f)", state.AnomalyScore),
				Timestamp: now,
				Details: map[string]interface{}{
					"actual_interval_ms":   interval.Milliseconds(),
					"expected_interval_ms": state.ExpectedInterval.Milliseconds(),
					"anomaly_score":        state.AnomalyScore,
				},
			})
		}
	}

	// Check if recovering from stale/offline
	if state.Status != statusHealthy {
		state.RecoveryCount++
		m.raiseAlertLocked(HeartbeatAlert{
			NodeID:    nodeID,
			ClusterID: clusterID,
			AlertType: "recovered",
			Severity:  "info",
			Message:   fmt.Sprintf("Node recovered from %s state", state.Status),
			Timestamp: now,
		})
	}

	state.LastHeartbeat = now
	state.LastSequence = sequence
	state.Status = statusHealthy
	state.MissedCount = 0

	// Update history
	state.HeartbeatHistory = append(state.HeartbeatHistory, now)
	if len(state.HeartbeatHistory) > 100 {
		state.HeartbeatHistory = state.HeartbeatHistory[1:]
	}
}

// RemoveNode removes a node from monitoring
func (m *HPCHeartbeatMonitor) RemoveNode(nodeID string) {
	m.nodesMu.Lock()
	defer m.nodesMu.Unlock()
	delete(m.nodes, nodeID)
}

// GetNodeStatus returns the status of a node
func (m *HPCHeartbeatMonitor) GetNodeStatus(nodeID string) (string, bool) {
	m.nodesMu.RLock()
	defer m.nodesMu.RUnlock()

	state, exists := m.nodes[nodeID]
	if !exists {
		return "", false
	}
	return state.Status, true
}

// GetMetrics returns monitoring metrics
func (m *HPCHeartbeatMonitor) GetMetrics() map[string]interface{} {
	m.metricsMu.RLock()
	defer m.metricsMu.RUnlock()

	return map[string]interface{}{
		"total_nodes":   m.totalNodes,
		"healthy_nodes": m.healthyNodes,
		"stale_nodes":   m.staleNodes,
		"offline_nodes": m.offlineNodes,
		"alerts_raised": m.alertsRaised,
	}
}

func (m *HPCHeartbeatMonitor) monitorLoop(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.checkAllNodes()
		}
	}
}

func (m *HPCHeartbeatMonitor) checkAllNodes() {
	m.nodesMu.Lock()

	now := time.Now()
	healthy, stale, offline := 0, 0, 0

	for nodeID, state := range m.nodes {
		timeSinceHB := now.Sub(state.LastHeartbeat)

		switch {
		case timeSinceHB > m.config.OfflineThreshold:
			if state.Status != statusOffline {
				state.Status = statusOffline
				m.raiseAlertLocked(HeartbeatAlert{
					NodeID:      nodeID,
					ClusterID:   state.ClusterID,
					AlertType:   statusOffline,
					Severity:    "critical",
					Message:     "Node is offline - no heartbeat received",
					Timestamp:   now,
					LastSeen:    state.LastHeartbeat,
					MissedCount: state.MissedCount,
				})
			}
			offline++
			state.MissedCount++

		case timeSinceHB > m.config.StaleThreshold:
			if state.Status == statusHealthy {
				state.Status = statusStale
				m.raiseAlertLocked(HeartbeatAlert{
					NodeID:      nodeID,
					ClusterID:   state.ClusterID,
					AlertType:   "stale",
					Severity:    "warning",
					Message:     "Node heartbeat is stale",
					Timestamp:   now,
					LastSeen:    state.LastHeartbeat,
					MissedCount: state.MissedCount,
				})
			}
			stale++
			state.MissedCount++

		default:
			healthy++
		}
	}

	// Update metrics
	m.metricsMu.Lock()
	m.totalNodes = len(m.nodes)
	m.healthyNodes = healthy
	m.staleNodes = stale
	m.offlineNodes = offline
	m.metricsMu.Unlock()

	m.nodesMu.Unlock()

	// Dispatch queued alerts only once the state lock has been released: a
	// handler that reads monitor state would otherwise self-deadlock.
	m.dispatchPendingAlerts()
}

// raiseAlertLocked queues an alert for dispatch after m.nodesMu is released.
// It must be called with m.nodesMu held. It deliberately does NOT invoke
// m.config.AlertHandler — the handler is arbitrary user code and calling it
// while holding m.nodesMu deadlocks any handler that touches monitor state.
func (m *HPCHeartbeatMonitor) raiseAlertLocked(alert HeartbeatAlert) {
	state, exists := m.nodes[alert.NodeID]
	if exists {
		// Deduplicate alerts
		lastSent, ok := state.AlertSent[alert.AlertType]
		if ok && time.Since(lastSent) < 5*time.Minute {
			return // Don't spam alerts
		}
		state.AlertSent[alert.AlertType] = time.Now()
	}

	m.metricsMu.Lock()
	m.alertsRaised++
	m.metricsMu.Unlock()

	m.pendingAlerts = append(m.pendingAlerts, alert)
}

// takePendingAlertsLocked removes and returns the queued alerts.
// Must be called with m.nodesMu held.
func (m *HPCHeartbeatMonitor) takePendingAlertsLocked() []HeartbeatAlert {
	if len(m.pendingAlerts) == 0 {
		return nil
	}
	pending := m.pendingAlerts
	m.pendingAlerts = nil
	return pending
}

// dispatchPendingAlerts drains queued alerts and invokes the AlertHandler with
// no monitor lock held. alertMu serializes invocations so handlers are never
// called concurrently, matching the mutual exclusion the state lock used to
// provide. alertMu is never held at the same time as m.nodesMu.
//
// The AlertHandler must not call back into this monitor synchronously: doing so
// would re-enter dispatchPendingAlerts and deadlock on alertMu. That
// restriction already applied before this queuing scheme — a re-entrant handler
// previously deadlocked on m.nodesMu instead.
func (m *HPCHeartbeatMonitor) dispatchPendingAlerts() {
	m.alertMu.Lock()
	defer m.alertMu.Unlock()

	handler := m.config.AlertHandler
	if handler == nil {
		// Nothing to deliver. Drop queued alerts (never call a nil handler).
		m.nodesMu.Lock()
		m.takePendingAlertsLocked()
		m.nodesMu.Unlock()
		return
	}

	for {
		m.nodesMu.Lock()
		pending := m.takePendingAlertsLocked()
		m.nodesMu.Unlock()

		if len(pending) == 0 {
			return
		}

		for _, alert := range pending {
			handler(alert)
		}
	}
}

func (m *HPCHeartbeatMonitor) calculateAnomalyScore(history []time.Duration, current time.Duration) float64 {
	if len(history) < 5 {
		return 0
	}

	// Calculate mean and standard deviation
	var sum int64
	for _, d := range history {
		sum += d.Milliseconds()
	}
	mean := float64(sum) / float64(len(history))

	var variance float64
	for _, d := range history {
		diff := float64(d.Milliseconds()) - mean
		variance += diff * diff
	}
	variance /= float64(len(history))
	stdDev := variance

	if stdDev == 0 {
		return 0
	}

	// Z-score
	currentMs := float64(current.Milliseconds())
	return (currentMs - mean) / stdDev
}

func defaultAlertHandler(alert HeartbeatAlert) {
	fmt.Printf("[HPC-HEARTBEAT] %s: %s - %s (node=%s, cluster=%s)\n",
		alert.Severity, alert.AlertType, alert.Message, alert.NodeID, alert.ClusterID)
}

// ClusterHealthSummary returns health summary for a cluster
func (m *HPCHeartbeatMonitor) ClusterHealthSummary(clusterID string) map[string]interface{} {
	m.nodesMu.RLock()
	defer m.nodesMu.RUnlock()

	healthy, stale, offline := 0, 0, 0
	var avgAnomalyScore float64

	for _, state := range m.nodes {
		if state.ClusterID != clusterID {
			continue
		}

		switch state.Status {
		case statusHealthy:
			healthy++
		case statusStale:
			stale++
		case statusOffline:
			offline++
		}
		avgAnomalyScore += state.AnomalyScore
	}

	total := healthy + stale + offline
	if total > 0 {
		avgAnomalyScore /= float64(total)
	}

	var healthPercent float64
	if total > 0 {
		healthPercent = float64(healthy) / float64(total) * 100
	}

	return map[string]interface{}{
		"cluster_id":        clusterID,
		"total_nodes":       total,
		"healthy_nodes":     healthy,
		"stale_nodes":       stale,
		"offline_nodes":     offline,
		"health_percent":    healthPercent,
		"avg_anomaly_score": avgAnomalyScore,
	}
}
