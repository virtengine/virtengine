// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type TransferMetrics interface {
	PendingAdded(PacketType)
	TerminalCommitted(PacketType, TransferState, CompensationReason)
	CompensationFailed(PacketType, string, CompensationReason)
	SetPending(map[PacketType]float64)
}

type noOpTransferMetrics struct{}

func (noOpTransferMetrics) PendingAdded(PacketType)                                         {}
func (noOpTransferMetrics) TerminalCommitted(PacketType, TransferState, CompensationReason) {}
func (noOpTransferMetrics) CompensationFailed(PacketType, string, CompensationReason)       {}
func (noOpTransferMetrics) SetPending(map[PacketType]float64)                               {}

// SettlementIBCMetrics exports bounded transfer lifecycle telemetry.
type SettlementIBCMetrics struct {
	PendingExposure      *prometheus.GaugeVec
	TerminalOutcomes     *prometheus.CounterVec
	CompensationFailures *prometheus.CounterVec
}

func NewSettlementIBCMetrics(registerer prometheus.Registerer) (*SettlementIBCMetrics, error) {
	if registerer == nil {
		return nil, fmt.Errorf("settlement IBC metrics registerer is required")
	}
	pending, err := registerIBCGaugeVec(registerer, prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "virtengine", Subsystem: "settlement_ibc", Name: "pending_exposure",
		Help: "Current committed pending transfers by bounded packet type.",
	}, []string{"packet_type"}))
	if err != nil {
		return nil, err
	}
	outcomes, err := registerIBCCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "virtengine", Subsystem: "settlement_ibc", Name: "terminal_outcomes_total",
		Help: "Committed terminal transfer outcomes by bounded packet type, outcome, and reason.",
	}, []string{"packet_type", "outcome", "reason"}))
	if err != nil {
		return nil, err
	}
	failures, err := registerIBCCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "virtengine", Subsystem: "settlement_ibc", Name: "compensation_failures_total",
		Help: "Failed compensation stages by bounded packet type, stage, and reason.",
	}, []string{"packet_type", "stage", "reason"}))
	if err != nil {
		return nil, err
	}
	return &SettlementIBCMetrics{PendingExposure: pending, TerminalOutcomes: outcomes, CompensationFailures: failures}, nil
}

func (m *SettlementIBCMetrics) PendingAdded(packetType PacketType) {
	m.PendingExposure.WithLabelValues(string(packetType)).Inc()
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

func (m *SettlementIBCMetrics) SetPending(counts map[PacketType]float64) {
	m.PendingExposure.Reset()
	for _, packetType := range []PacketType{PacketTypeEscrowDeposit, PacketTypeEscrowRelease, PacketTypeSettlementRecord} {
		m.PendingExposure.WithLabelValues(string(packetType)).Set(counts[packetType])
	}
}

func registerIBCGaugeVec(registerer prometheus.Registerer, collector *prometheus.GaugeVec) (*prometheus.GaugeVec, error) {
	if err := registerer.Register(collector); err != nil {
		alreadyRegistered, ok := err.(prometheus.AlreadyRegisteredError)
		if !ok {
			return nil, err
		}
		existing, ok := alreadyRegistered.ExistingCollector.(*prometheus.GaugeVec)
		if !ok {
			return nil, fmt.Errorf("registered settlement IBC gauge has unexpected type")
		}
		return existing, nil
	}
	return collector, nil
}

func registerIBCCounterVec(registerer prometheus.Registerer, collector *prometheus.CounterVec) (*prometheus.CounterVec, error) {
	if err := registerer.Register(collector); err != nil {
		alreadyRegistered, ok := err.(prometheus.AlreadyRegisteredError)
		if !ok {
			return nil, err
		}
		existing, ok := alreadyRegistered.ExistingCollector.(*prometheus.CounterVec)
		if !ok {
			return nil, fmt.Errorf("registered settlement IBC counter has unexpected type")
		}
		return existing, nil
	}
	return collector, nil
}
