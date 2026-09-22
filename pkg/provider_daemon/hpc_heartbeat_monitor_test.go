package provider_daemon

import (
	"sync/atomic"
	"testing"
	"time"
)

// alertHandlerDeadline bounds how long a heartbeat call may take before the test
// declares a deadlock. A self-deadlock never returns, so without this guard the
// whole package would hang until the go-test timeout instead of failing here.
const (
	alertHandlerDeadline = 10 * time.Second
	probeClusterID       = "cluster-1"
)

// newDeadlockProbeMonitor builds a monitor whose AlertHandler reads monitor state.
// That read is exactly the shape that deadlocked while the handler ran under
// m.nodesMu (Go mutexes are not reentrant).
func newDeadlockProbeMonitor(alerts *int32) *HPCHeartbeatMonitor {
	var m *HPCHeartbeatMonitor

	config := DefaultHPCHeartbeatMonitorConfig()
	config.AlertHandler = func(alert HeartbeatAlert) {
		m.GetNodeStatus(alert.NodeID)
		m.ClusterHealthSummary(alert.ClusterID)
		m.GetMetrics()
		atomic.AddInt32(alerts, 1)
	}

	// Assign m before returning so the handler above closes over the live monitor.
	m = NewHPCHeartbeatMonitor(config)
	return m
}

// TestHeartbeatMonitorSequenceGapAlertDoesNotDeadlock reproduces the VE-500
// self-deadlock: RecordHeartbeat held m.nodesMu for its entire body and then
// invoked the AlertHandler, so any handler that read monitor state re-entered the
// non-reentrant mutex and blocked forever.
//
// The sequence-gap path is reached deterministically on the second heartbeat, so
// this needs no timers.
func TestHeartbeatMonitorSequenceGapAlertDoesNotDeadlock(t *testing.T) {
	var alerts int32
	m := newDeadlockProbeMonitor(&alerts)

	done := make(chan struct{})
	go func() {
		defer close(done)
		m.RecordHeartbeat("node-gap", probeClusterID, 1)
		m.RecordHeartbeat("node-gap", probeClusterID, 5) // 3-beat gap -> sequence_gap alert
	}()

	select {
	case <-done:
	case <-time.After(alertHandlerDeadline):
		t.Fatal("RecordHeartbeat deadlocked: AlertHandler ran while m.nodesMu was held")
	}

	if got := atomic.LoadInt32(&alerts); got < 1 {
		t.Fatalf("expected the sequence-gap alert to reach the handler, got %d alerts", got)
	}
}

// TestHeartbeatMonitorRecoveryAlertDoesNotDeadlock covers the specific stack that
// hung CI (node_agent_test.go:133 -> RecordHeartbeat -> raiseAlert): a node that
// had gone stale/offline receives a heartbeat, which raises the "recovered" alert.
func TestHeartbeatMonitorRecoveryAlertDoesNotDeadlock(t *testing.T) {
	var alerts int32
	m := newDeadlockProbeMonitor(&alerts)

	nodeID := "node-recover"
	clusterID := probeClusterID

	// Seed the node as offline with a fresh heartbeat, so the next
	// RecordHeartbeat raises only the recovery alert (no gap, no anomaly).
	m.nodesMu.Lock()
	m.nodes[nodeID] = &NodeHeartbeatState{
		NodeID:           nodeID,
		ClusterID:        clusterID,
		LastHeartbeat:    time.Now(),
		LastSequence:     1,
		ExpectedInterval: 30 * time.Second,
		Status:           statusOffline,
		AlertSent:        make(map[string]time.Time),
		HeartbeatHistory: make([]time.Time, 0, 100),
		IntervalHistory:  make([]time.Duration, 0, 100),
	}
	m.nodesMu.Unlock()

	done := make(chan struct{})
	go func() {
		defer close(done)
		m.RecordHeartbeat(nodeID, clusterID, 2)
	}()

	select {
	case <-done:
	case <-time.After(alertHandlerDeadline):
		t.Fatal("RecordHeartbeat deadlocked on the recovery alert: AlertHandler ran while m.nodesMu was held")
	}

	if got := atomic.LoadInt32(&alerts); got != 1 {
		t.Fatalf("expected 1 recovered alert, got %d", got)
	}

	// The heartbeat must still be applied even though an alert fired.
	status, ok := m.GetNodeStatus(nodeID)
	if !ok {
		t.Fatal("node disappeared from the monitor")
	}
	if status != statusHealthy {
		t.Fatalf("expected status %q after recovery, got %q", statusHealthy, status)
	}
}
