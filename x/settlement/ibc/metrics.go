// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type TransferMetrics interface {
	PendingAdded(PacketType)
	TerminalCommitted(PacketType, TransferState, CompensationReason)
	CompensationFailed(PacketType, string, CompensationReason)
	SetPending(map[PacketType]float64, bool)
}

type noOpTransferMetrics struct{}

func (noOpTransferMetrics) PendingAdded(PacketType)                                         {}
func (noOpTransferMetrics) TerminalCommitted(PacketType, TransferState, CompensationReason) {}
func (noOpTransferMetrics) CompensationFailed(PacketType, string, CompensationReason)       {}
func (noOpTransferMetrics) SetPending(map[PacketType]float64, bool)                         {}

// SettlementIBCMetrics exports bounded transfer lifecycle telemetry.
type SettlementIBCMetrics struct {
	pending              *pendingMetricsCollector
	TerminalOutcomes     *prometheus.CounterVec
	CompensationFailures *prometheus.CounterVec
}

func NewSettlementIBCMetrics(registerer prometheus.Registerer) (*SettlementIBCMetrics, error) {
	if registerer == nil {
		return nil, fmt.Errorf("settlement IBC metrics registerer is required")
	}
	registered := make([]prometheus.Collector, 0, 3)
	rollback := func() {
		for index := len(registered) - 1; index >= 0; index-- {
			registerer.Unregister(registered[index])
		}
	}
	pending, added, err := registerIBCCollector(registerer, newPendingMetricsCollector())
	if err != nil {
		return nil, err
	}
	if added {
		registered = append(registered, pending)
	}
	outcomes, added, err := registerIBCCollector(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "virtengine", Subsystem: "settlement_ibc", Name: "terminal_outcomes_total",
		Help: "Committed terminal transfer outcomes by bounded packet type, outcome, and reason.",
	}, []string{"packet_type", "outcome", "reason"}))
	if err != nil {
		rollback()
		return nil, err
	}
	if added {
		registered = append(registered, outcomes)
	}
	failures, added, err := registerIBCCollector(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "virtengine", Subsystem: "settlement_ibc", Name: "compensation_failures_total",
		Help: "Failed compensation stages by bounded packet type, stage, and reason.",
	}, []string{"packet_type", "stage", "reason"}))
	if err != nil {
		rollback()
		return nil, err
	}
	return &SettlementIBCMetrics{pending: pending, TerminalOutcomes: outcomes, CompensationFailures: failures}, nil
}

func (m *SettlementIBCMetrics) PendingAdded(packetType PacketType) {
	m.pending.Increment(packetType)
}

func (m *SettlementIBCMetrics) TerminalCommitted(packetType PacketType, state TransferState, reason CompensationReason) {
	reasonLabel := string(reason)
	if reasonLabel == "" {
		reasonLabel = "none"
	}
	m.TerminalOutcomes.WithLabelValues(string(packetType), string(state), reasonLabel).Inc()
}

func (m *SettlementIBCMetrics) CompensationFailed(packetType PacketType, stage string, reason CompensationReason) {
	m.CompensationFailures.WithLabelValues(string(packetType), stage, string(reason)).Inc()
}

func (m *SettlementIBCMetrics) SetPending(counts map[PacketType]float64, available bool) {
	m.pending.Set(counts, available)
}

type pendingMetricsCollector struct {
	mu        sync.RWMutex
	available bool
	counts    map[PacketType]float64
	exposure  *prometheus.Desc
	status    *prometheus.Desc
}

func newPendingMetricsCollector() *pendingMetricsCollector {
	return &pendingMetricsCollector{
		counts: make(map[PacketType]float64),
		exposure: prometheus.NewDesc(
			"virtengine_settlement_ibc_pending_exposure",
			"Current committed pending transfers by bounded packet type.",
			[]string{"packet_type"}, nil,
		),
		status: prometheus.NewDesc(
			"virtengine_settlement_ibc_pending_projection_available",
			"Whether the complete committed pending-transfer projection is valid and available.",
			nil, nil,
		),
	}
}

func (collector *pendingMetricsCollector) Increment(packetType PacketType) {
	collector.mu.Lock()
	defer collector.mu.Unlock()
	if collector.available && validPacketType(packetType) {
		collector.counts[packetType]++
	}
}

func (collector *pendingMetricsCollector) Set(counts map[PacketType]float64, available bool) {
	collector.mu.Lock()
	defer collector.mu.Unlock()
	collector.available = available
	collector.counts = make(map[PacketType]float64, len(counts))
	if available {
		for packetType, count := range counts {
			if validPacketType(packetType) {
				collector.counts[packetType] = count
			}
		}
	}
}

func (collector *pendingMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.exposure
	ch <- collector.status
}

func (collector *pendingMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	collector.mu.RLock()
	defer collector.mu.RUnlock()
	ch <- prometheus.MustNewConstMetric(collector.status, prometheus.GaugeValue, boolMetricValue(collector.available))
	if !collector.available {
		return
	}
	for _, packetType := range []PacketType{PacketTypeEscrowDeposit, PacketTypeEscrowRelease, PacketTypeSettlementRecord} {
		ch <- prometheus.MustNewConstMetric(collector.exposure, prometheus.GaugeValue, collector.counts[packetType], string(packetType))
	}
}

func validPacketType(packetType PacketType) bool {
	switch packetType {
	case PacketTypeEscrowDeposit, PacketTypeEscrowRelease, PacketTypeSettlementRecord:
		return true
	default:
		return false
	}
}

func boolMetricValue(value bool) float64 {
	if value {
		return 1
	}
	return 0
}

func registerIBCCollector[T prometheus.Collector](registerer prometheus.Registerer, collector T) (T, bool, error) {
	if err := registerer.Register(collector); err != nil {
		alreadyRegistered, ok := err.(prometheus.AlreadyRegisteredError)
		if !ok {
			var zero T
			return zero, false, err
		}
		existing, ok := alreadyRegistered.ExistingCollector.(T)
		if !ok {
			var zero T
			return zero, false, fmt.Errorf("registered settlement IBC collector has unexpected type")
		}
		return existing, false, nil
	}
	return collector, true, nil
}
