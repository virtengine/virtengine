// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestSettlementIBCMetricsReuseRegisteredCollectors(t *testing.T) {
	registry := prometheus.NewRegistry()
	first, err := NewSettlementIBCMetrics(registry)
	require.NoError(t, err)
	second, err := NewSettlementIBCMetrics(registry)
	require.NoError(t, err)

	second.SetPending(map[PacketType]float64{PacketTypeEscrowDeposit: 1}, true)
	require.Same(t, first.pending, second.pending)
	require.Equal(t, float64(1), pendingMetricValue(t, registry, "virtengine_settlement_ibc_pending_exposure", map[string]string{"packet_type": string(PacketTypeEscrowDeposit)}))
	second.TerminalOutcomes.WithLabelValues(string(PacketTypeEscrowDeposit), string(TransferStateFinalized), "none").Inc()
	require.Equal(t, float64(1), testutil.ToFloat64(first.TerminalOutcomes.WithLabelValues(string(PacketTypeEscrowDeposit), string(TransferStateFinalized), "none")))
}

func TestSettlementIBCMetricsFollowCommittedTransitions(t *testing.T) {
	env := setupIBCTestEnv(t)
	registry := prometheus.NewRegistry()
	metrics, err := NewSettlementIBCMetrics(registry)
	require.NoError(t, err)
	env.keeper.SetMetrics(metrics)

	sequence, err := env.keeper.SendEscrowDepositPacket(env.ctx, "channel-0", clienttypes.Height{}, 0, validDepositPacket(env.ctx.BlockTime()))
	require.NoError(t, err)
	require.Equal(t, float64(1), pendingMetricValue(t, registry, "virtengine_settlement_ibc_pending_exposure", map[string]string{"packet_type": string(PacketTypeEscrowDeposit)}))
	packet := channeltypes.NewPacket(env.channel.sent[0].data, sequence, PortID, "channel-0", PortID, "channel-1", clienttypes.NewHeight(0, 20), 0)
	relayer := types.AccAddress([]byte("relayer_addr________"))

	env.lifecycle.failAt = "compensate"
	require.Error(t, env.keeper.OnTimeoutPacket(env.ctx, packet, relayer))
	require.Equal(t, float64(1), pendingMetricValue(t, registry, "virtengine_settlement_ibc_pending_exposure", map[string]string{"packet_type": string(PacketTypeEscrowDeposit)}))
	require.Equal(t, float64(1), testutil.ToFloat64(metrics.CompensationFailures.WithLabelValues(string(PacketTypeEscrowDeposit), "compensate", string(CompensationReasonTimeout))))
	require.Equal(t, float64(0), testutil.ToFloat64(metrics.TerminalOutcomes.WithLabelValues(string(PacketTypeEscrowDeposit), string(TransferStateCompensated), string(CompensationReasonTimeout))))

	env.lifecycle.failAt = ""
	require.NoError(t, env.keeper.OnTimeoutPacket(env.ctx, packet, relayer))
	require.NoError(t, env.keeper.OnTimeoutPacket(env.ctx, packet, relayer))
	require.Equal(t, float64(0), pendingMetricValue(t, registry, "virtengine_settlement_ibc_pending_exposure", map[string]string{"packet_type": string(PacketTypeEscrowDeposit)}))
	require.Equal(t, float64(1), testutil.ToFloat64(metrics.TerminalOutcomes.WithLabelValues(string(PacketTypeEscrowDeposit), string(TransferStateCompensated), string(CompensationReasonTimeout))))
}

func TestSettlementIBCMetricsRefreshPendingFromStore(t *testing.T) {
	env := setupIBCTestEnv(t)
	registry := prometheus.NewRegistry()
	metrics, err := NewSettlementIBCMetrics(registry)
	require.NoError(t, err)
	env.keeper.SetMetrics(metrics)
	_, err = env.keeper.SendEscrowDepositPacket(env.ctx, "channel-0", clienttypes.Height{}, 0, validDepositPacket(env.ctx.BlockTime()))
	require.NoError(t, err)
	metrics.pending.Set(map[PacketType]float64{PacketTypeEscrowDeposit: 99}, true)

	env.keeper.RefreshMetrics(env.ctx)
	require.Equal(t, float64(1), pendingMetricValue(t, registry, "virtengine_settlement_ibc_pending_exposure", map[string]string{"packet_type": string(PacketTypeEscrowDeposit)}))
	require.Equal(t, float64(0), pendingMetricValue(t, registry, "virtengine_settlement_ibc_pending_exposure", map[string]string{"packet_type": string(PacketTypeEscrowRelease)}))
	require.Equal(t, float64(1), pendingMetricValue(t, registry, "virtengine_settlement_ibc_pending_projection_available", nil))
}

func TestSettlementIBCMetricsRefreshFailsClosedOnMalformedPendingRecord(t *testing.T) {
	env := setupIBCTestEnv(t)
	registry := prometheus.NewRegistry()
	metrics, err := NewSettlementIBCMetrics(registry)
	require.NoError(t, err)
	env.keeper.SetMetrics(metrics)
	_, err = env.keeper.SendEscrowDepositPacket(env.ctx, "channel-0", clienttypes.Height{}, 0, validDepositPacket(env.ctx.BlockTime()))
	require.NoError(t, err)
	corruptKey := append(append([]byte(nil), PrefixPendingPacket...), []byte("corrupt")...)
	env.ctx.KVStore(env.storeKey).Set(corruptKey, []byte("not-json"))

	env.keeper.RefreshMetrics(env.ctx)
	require.Equal(t, float64(0), pendingMetricValue(t, registry, "virtengine_settlement_ibc_pending_projection_available", nil))
	families, err := registry.Gather()
	require.NoError(t, err)
	for _, family := range families {
		if family.GetName() == "virtengine_settlement_ibc_pending_exposure" {
			t.Fatalf("partial pending exposure must be omitted when projection is unavailable")
		}
	}

	env.ctx.KVStore(env.storeKey).Delete(corruptKey)
	env.keeper.RefreshMetrics(env.ctx)
	require.Equal(t, float64(1), pendingMetricValue(t, registry, "virtengine_settlement_ibc_pending_projection_available", nil))
	require.Equal(t, float64(1), pendingMetricValue(t, registry, "virtengine_settlement_ibc_pending_exposure", map[string]string{"packet_type": string(PacketTypeEscrowDeposit)}))
}

func TestSettlementIBCMetricsStartUnavailableAndRejectMismatchedPendingKey(t *testing.T) {
	env := setupIBCTestEnv(t)
	registry := prometheus.NewRegistry()
	metrics, err := NewSettlementIBCMetrics(registry)
	require.NoError(t, err)
	require.Equal(t, float64(0), pendingMetricValue(t, registry, "virtengine_settlement_ibc_pending_projection_available", nil))
	env.keeper.SetMetrics(metrics)
	sequence, err := env.keeper.SendEscrowDepositPacket(env.ctx, "channel-0", clienttypes.Height{}, 0, validDepositPacket(env.ctx.BlockTime()))
	require.NoError(t, err)
	store := env.ctx.KVStore(env.storeKey)
	valid := store.Get(PendingPacketKey("channel-0", sequence))
	store.Set(PendingPacketKey("channel-other", sequence), valid)

	env.keeper.RefreshMetrics(env.ctx)
	require.Equal(t, float64(0), pendingMetricValue(t, registry, "virtengine_settlement_ibc_pending_projection_available", nil))
}

func TestSettlementIBCMetricsRejectsDigestCorruption(t *testing.T) {
	env := setupIBCTestEnv(t)
	registry := prometheus.NewRegistry()
	metrics, err := NewSettlementIBCMetrics(registry)
	require.NoError(t, err)
	env.keeper.SetMetrics(metrics)
	sequence, err := env.keeper.SendEscrowDepositPacket(env.ctx, "channel-0", clienttypes.Height{}, 0, validDepositPacket(env.ctx.BlockTime()))
	require.NoError(t, err)
	store := env.ctx.KVStore(env.storeKey)
	key := PendingPacketKey("channel-0", sequence)
	var pending PendingPacket
	require.NoError(t, json.Unmarshal(store.Get(key), &pending))
	pending.Identity.PayloadDigest[0] ^= 0xff
	encoded, err := json.Marshal(pending)
	require.NoError(t, err)
	store.Set(key, encoded)

	env.keeper.RefreshMetrics(env.ctx)
	require.Equal(t, float64(0), pendingMetricValue(t, registry, "virtengine_settlement_ibc_pending_projection_available", nil))
}

func TestSettlementIBCMetricsPendingSnapshotIsAtomic(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics, err := NewSettlementIBCMetrics(registry)
	require.NoError(t, err)
	counts := map[PacketType]float64{PacketTypeEscrowDeposit: 1}
	var wait sync.WaitGroup
	for iteration := 0; iteration < 100; iteration++ {
		wait.Add(1)
		go func(available bool) {
			defer wait.Done()
			metrics.SetPending(counts, available)
		}(iteration%2 == 0)
		families, err := registry.Gather()
		require.NoError(t, err)
		available := familyGaugeValue(t, families, "virtengine_settlement_ibc_pending_projection_available")
		exposureCount := familySeriesCount(families, "virtengine_settlement_ibc_pending_exposure")
		if available == 0 {
			require.Zero(t, exposureCount)
		} else {
			require.Equal(t, 3, exposureCount)
		}
	}
	wait.Wait()
}

func TestSettlementIBCMetricsRollsBackLateRegistrationConflict(t *testing.T) {
	registry := prometheus.NewRegistry()
	require.NoError(t, registry.Register(prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "virtengine", Subsystem: "settlement_ibc", Name: "compensation_failures_total",
		Help: "Conflicting compensation failure collector.",
	}, []string{"packet_type", "stage", "reason"})))
	_, err := NewSettlementIBCMetrics(registry)
	require.Error(t, err)

	require.NoError(t, registry.Register(newPendingMetricsCollector()))
	require.NoError(t, registry.Register(prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "virtengine", Subsystem: "settlement_ibc", Name: "terminal_outcomes_total",
		Help: "Committed terminal transfer outcomes by bounded packet type, outcome, and reason.",
	}, []string{"packet_type", "outcome", "reason"})))
}

func pendingMetricValue(t *testing.T, registry *prometheus.Registry, name string, labels map[string]string) float64 {
	t.Helper()
	families, err := registry.Gather()
	require.NoError(t, err)
	for _, family := range families {
		if family.GetName() != name {
			continue
		}
		for _, metric := range family.Metric {
			if metricLabelsMatch(metric, labels) {
				return metric.GetGauge().GetValue()
			}
		}
	}
	t.Fatalf("metric %s with labels %v not found", name, labels)
	return 0
}

func metricLabelsMatch(metric *dto.Metric, expected map[string]string) bool {
	for name, value := range expected {
		matched := false
		for _, label := range metric.Label {
			if label.GetName() == name && label.GetValue() == value {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func familyGaugeValue(t *testing.T, families []*dto.MetricFamily, name string) float64 {
	t.Helper()
	for _, family := range families {
		if family.GetName() == name && len(family.Metric) == 1 {
			return family.Metric[0].GetGauge().GetValue()
		}
	}
	t.Fatalf("metric family %s not found", name)
	return 0
}

func familySeriesCount(families []*dto.MetricFamily, name string) int {
	for _, family := range families {
		if family.GetName() == name {
			return len(family.Metric)
		}
	}
	return 0
}
