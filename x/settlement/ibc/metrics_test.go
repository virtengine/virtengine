// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

func TestSettlementIBCMetricsReuseRegisteredCollectors(t *testing.T) {
	registry := prometheus.NewRegistry()
	first, err := NewSettlementIBCMetrics(registry)
	require.NoError(t, err)
	second, err := NewSettlementIBCMetrics(registry)
	require.NoError(t, err)

	second.PendingAdded(PacketTypeEscrowDeposit)
	require.Equal(t, float64(1), testutil.ToFloat64(first.PendingExposure.WithLabelValues(string(PacketTypeEscrowDeposit))))
	second.TerminalOutcomes.WithLabelValues(string(PacketTypeEscrowDeposit), string(TransferStateFinalized), "none").Inc()
	require.Equal(t, float64(1), testutil.ToFloat64(first.TerminalOutcomes.WithLabelValues(string(PacketTypeEscrowDeposit), string(TransferStateFinalized), "none")))
}

func TestSettlementIBCMetricsFollowCommittedTransitions(t *testing.T) {
	env := setupIBCTestEnv(t)
	metrics, err := NewSettlementIBCMetrics(prometheus.NewRegistry())
	require.NoError(t, err)
	env.keeper.SetMetrics(metrics)

	sequence, err := env.keeper.SendEscrowDepositPacket(env.ctx, "channel-0", clienttypes.Height{}, 0, validDepositPacket(env.ctx.BlockTime()))
	require.NoError(t, err)
	require.Equal(t, float64(1), testutil.ToFloat64(metrics.PendingExposure.WithLabelValues(string(PacketTypeEscrowDeposit))))
	packet := channeltypes.NewPacket(env.channel.sent[0].data, sequence, PortID, "channel-0", PortID, "channel-1", clienttypes.NewHeight(0, 20), 0)
	relayer := types.AccAddress([]byte("relayer_addr________"))

	env.lifecycle.failAt = "compensate"
	require.Error(t, env.keeper.OnTimeoutPacket(env.ctx, packet, relayer))
	require.Equal(t, float64(1), testutil.ToFloat64(metrics.PendingExposure.WithLabelValues(string(PacketTypeEscrowDeposit))))
	require.Equal(t, float64(1), testutil.ToFloat64(metrics.CompensationFailures.WithLabelValues(string(PacketTypeEscrowDeposit), "compensate", string(CompensationReasonTimeout))))
	require.Equal(t, float64(0), testutil.ToFloat64(metrics.TerminalOutcomes.WithLabelValues(string(PacketTypeEscrowDeposit), string(TransferStateCompensated), string(CompensationReasonTimeout))))

	env.lifecycle.failAt = ""
	require.NoError(t, env.keeper.OnTimeoutPacket(env.ctx, packet, relayer))
	require.NoError(t, env.keeper.OnTimeoutPacket(env.ctx, packet, relayer))
	require.Equal(t, float64(0), testutil.ToFloat64(metrics.PendingExposure.WithLabelValues(string(PacketTypeEscrowDeposit))))
	require.Equal(t, float64(1), testutil.ToFloat64(metrics.TerminalOutcomes.WithLabelValues(string(PacketTypeEscrowDeposit), string(TransferStateCompensated), string(CompensationReasonTimeout))))
}

func TestSettlementIBCMetricsRefreshPendingFromStore(t *testing.T) {
	env := setupIBCTestEnv(t)
	metrics, err := NewSettlementIBCMetrics(prometheus.NewRegistry())
	require.NoError(t, err)
	env.keeper.SetMetrics(metrics)
	_, err = env.keeper.SendEscrowDepositPacket(env.ctx, "channel-0", clienttypes.Height{}, 0, validDepositPacket(env.ctx.BlockTime()))
	require.NoError(t, err)
	metrics.PendingExposure.WithLabelValues(string(PacketTypeEscrowDeposit)).Set(99)

	env.keeper.RefreshMetrics(env.ctx)
	require.Equal(t, float64(1), testutil.ToFloat64(metrics.PendingExposure.WithLabelValues(string(PacketTypeEscrowDeposit))))
	require.Equal(t, float64(0), testutil.ToFloat64(metrics.PendingExposure.WithLabelValues(string(PacketTypeEscrowRelease))))
}
