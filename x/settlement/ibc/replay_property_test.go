// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"
)

type replayCallbackKind uint8

const (
	replaySuccessAck replayCallbackKind = iota
	replayErrorAckA
	replayErrorAckB
	replayTimeout
)

type replayDurableSnapshot struct {
	Pending []byte                 `json:"pending,omitempty"`
	Marker  []byte                 `json:"marker,omitempty"`
	Ack     []byte                 `json:"ack,omitempty"`
	Timeout []byte                 `json:"timeout,omitempty"`
	Ledger  testConservationLedger `json:"ledger"`
}

func TestIBCKeeperRandomizedTerminalReplayConservation(t *testing.T) {
	rng := rand.New(rand.NewSource(84)) //nolint:gosec // deterministic property-test input
	for caseIndex := 0; caseIndex < 256; caseIndex++ {
		trace := make([]byte, 1+rng.Intn(32))
		_, err := rng.Read(trace)
		require.NoError(t, err)
		runTerminalReplayCase(t, trace, uint16(rng.Intn(1<<16)), uint8(rng.Intn(5)), caseIndex)
	}
}

func FuzzIBCKeeperTerminalCallbackReplay(f *testing.F) {
	f.Add([]byte{0}, uint16(0), uint8(0))
	f.Add([]byte{1, 2, 3, 0}, uint16(999), uint8(1))
	f.Add([]byte{2}, uint16(65535), uint8(3))
	f.Add([]byte{}, uint16(41), uint8(4))

	f.Fuzz(func(t *testing.T, trace []byte, rawAmount uint16, rawFailure uint8) {
		if len(trace) == 0 {
			trace = []byte{0}
		}
		if len(trace) > 32 {
			trace = trace[:32]
		}
		runTerminalReplayCase(t, trace, rawAmount, rawFailure, int(rawAmount))
	})
}

func runTerminalReplayCase(t *testing.T, trace []byte, rawAmount uint16, rawFailure uint8, caseIndex int) {
	t.Helper()
	env := setupIBCTestEnv(t)
	amount := int64(rawAmount%1000) + 1
	deposit := validDepositPacket(env.ctx.BlockTime())
	deposit.DepositID = fmt.Sprintf("deposit-replay-%d", caseIndex)
	deposit.OrderID = fmt.Sprintf("order-replay-%d", caseIndex)
	deposit.Amount = sdk.NewCoins(sdk.NewInt64Coin("uve", amount))

	sequence, err := env.keeper.SendEscrowDepositPacket(env.ctx, "channel-0", clienttypes.Height{}, 0, deposit)
	require.NoError(t, err)
	require.Len(t, env.channel.sent, 1)
	packet := channeltypes.NewPacket(
		env.channel.sent[0].data,
		sequence,
		PortID,
		"channel-0",
		PortID,
		"channel-1",
		clienttypes.NewHeight(0, 20),
		0,
	)
	relayer := sdk.AccAddress([]byte("relayer_addr________"))
	pending, found := env.keeper.getPendingPacket(env.ctx, "channel-0", sequence)
	require.True(t, found)
	require.Equal(t, deposit.DepositID, pending.Identity.LogicalPayoutID)
	require.Equal(t, payloadDigest(packet.GetData()), pending.Identity.PayloadDigest)
	assertReplayConservation(t, env)

	firstKind := replayKind(trace[0])
	env.lifecycle.failAt = replayFailurePoint(rawFailure, firstKind)
	err = runReplayCallback(env, packet, relayer, firstKind)
	nextTraceIndex := 1
	winner := firstKind
	if env.lifecycle.failAt != "" {
		require.Error(t, err)
		assertReplayFailedCallback(t, env, packet, amount)
		env.lifecycle.failAt = ""
		if len(trace) > 1 {
			winner = replayKind(trace[1])
			nextTraceIndex = 2
		}
		require.NoError(t, runReplayCallback(env, packet, relayer, winner))
	} else {
		require.NoError(t, err)
	}
	assertReplayTerminal(t, env, packet, pending.Identity, amount, winner)

	for _, rawKind := range trace[nextTraceIndex:] {
		kind := replayKind(rawKind)
		before := replaySnapshot(t, env, packet)
		err = runReplayCallback(env, packet, relayer, kind)
		if kind == winner {
			require.NoError(t, err)
		} else {
			require.ErrorIs(t, err, ErrTerminalConflict)
		}
		require.Equal(t, before, replaySnapshot(t, env, packet))
		assertReplayTerminal(t, env, packet, pending.Identity, amount, winner)
	}

	before := replaySnapshot(t, env, packet)
	require.NoError(t, runReplayCallback(env, packet, relayer, winner))
	require.Equal(t, before, replaySnapshot(t, env, packet))
	assertReplayTerminal(t, env, packet, pending.Identity, amount, winner)

	for kind := replaySuccessAck; kind <= replayTimeout; kind++ {
		if kind == winner {
			continue
		}
		before = replaySnapshot(t, env, packet)
		require.ErrorIs(t, runReplayCallback(env, packet, relayer, kind), ErrTerminalConflict)
		require.Equal(t, before, replaySnapshot(t, env, packet))
		assertReplayTerminal(t, env, packet, pending.Identity, amount, winner)
	}
}

func replayKind(raw byte) replayCallbackKind {
	return replayCallbackKind(raw % 4)
}

func replayFailurePoint(raw uint8, first replayCallbackKind) string {
	switch raw % 5 {
	case 1:
		if first == replaySuccessAck {
			return "finalize"
		}
		return "compensate"
	case 2:
		return "accounting"
	case 3:
		return "audit"
	default:
		return ""
	}
}

func replayAcknowledgement(kind replayCallbackKind) []byte {
	switch kind {
	case replaySuccessAck:
		return NewResultAcknowledgement(EscrowDepositAck{EscrowID: "escrow-replay", Status: "created"}).GetBytes()
	case replayErrorAckA:
		return NewErrorAcknowledgement(fmt.Errorf("replay rejected A")).GetBytes()
	case replayErrorAckB:
		return NewErrorAcknowledgement(fmt.Errorf("replay rejected B")).GetBytes()
	default:
		return nil
	}
}

func runReplayCallback(env ibcTestEnv, packet channeltypes.Packet, relayer sdk.AccAddress, kind replayCallbackKind) error {
	if kind == replayTimeout {
		return env.keeper.OnTimeoutPacket(env.ctx, packet, relayer)
	}
	return env.keeper.OnAcknowledgementPacket(env.ctx, packet, replayAcknowledgement(kind), relayer)
}

func assertReplayConservation(t *testing.T, env ibcTestEnv) {
	t.Helper()
	ledger := env.lifecycle.getLedger(env.ctx)
	require.GreaterOrEqual(t, ledger.Available, int64(0))
	require.GreaterOrEqual(t, ledger.Held, int64(0))
	require.GreaterOrEqual(t, ledger.Committed, int64(0))
	require.GreaterOrEqual(t, ledger.Refunded, int64(0))
	require.Equal(t, int64(1000), ledger.Available+ledger.Held+ledger.Committed+ledger.Refunded)
}

func assertReplayFailedCallback(t *testing.T, env ibcTestEnv, packet channeltypes.Packet, amount int64) {
	t.Helper()
	pending, found := env.keeper.getPendingPacket(env.ctx, packet.GetSourceChannel(), packet.GetSequence())
	require.True(t, found)
	require.Equal(t, TransferStatePending, pending.State)
	_, found = env.keeper.getTerminalMarker(env.ctx, packet.GetSourceChannel(), packet.GetSequence())
	require.False(t, found)
	store := env.ctx.KVStore(env.storeKey)
	require.Nil(t, store.Get(AckPacketKey(packet.GetSourceChannel(), packet.GetSequence())))
	require.Nil(t, store.Get(TimeoutPacketKey(packet.GetSourceChannel(), packet.GetSequence())))
	require.Equal(t, testConservationLedger{
		Available: 1000 - amount,
		Held:      amount,
		Calls:     []string{"prepare"},
	}, env.lifecycle.getLedger(env.ctx))
	assertReplayConservation(t, env)
}

func assertReplayTerminal(
	t *testing.T,
	env ibcTestEnv,
	packet channeltypes.Packet,
	identity TransferIdentity,
	amount int64,
	winner replayCallbackKind,
) {
	t.Helper()
	_, found := env.keeper.getPendingPacket(env.ctx, packet.GetSourceChannel(), packet.GetSequence())
	require.False(t, found)
	marker, found := env.keeper.getTerminalMarker(env.ctx, packet.GetSourceChannel(), packet.GetSequence())
	require.True(t, found)
	require.Equal(t, identity, marker.Identity)
	require.Equal(t, packet.GetSourcePort(), marker.Identity.SourcePort)
	require.Equal(t, packet.GetSourceChannel(), marker.Identity.SourceChannel)
	require.Equal(t, packet.GetSequence(), marker.Identity.Sequence)
	require.Equal(t, payloadDigest(packet.GetData()), marker.Identity.PayloadDigest)

	store := env.ctx.KVStore(env.storeKey)
	ackEvidence := store.Get(AckPacketKey(packet.GetSourceChannel(), packet.GetSequence()))
	timeoutEvidence := store.Get(TimeoutPacketKey(packet.GetSourceChannel(), packet.GetSequence()))
	ledger := env.lifecycle.getLedger(env.ctx)
	require.Zero(t, ledger.Held)
	require.Equal(t, int64(1000)-amount, ledger.Available)

	if winner == replaySuccessAck {
		require.Equal(t, TransferStateFinalized, marker.State)
		require.Equal(t, CompensationReasonNone, marker.CompensationReason)
		require.Equal(t, acknowledgementDigest(replayAcknowledgement(winner)), marker.CallbackDigest)
		require.Equal(t, replayAcknowledgement(winner), ackEvidence)
		require.Nil(t, timeoutEvidence)
		require.Equal(t, amount, ledger.Committed)
		require.Zero(t, ledger.Refunded)
		require.Equal(t, []string{"prepare", "finalize", "accounting", "audit"}, ledger.Calls)
	} else {
		require.Equal(t, TransferStateCompensated, marker.State)
		require.Zero(t, ledger.Committed)
		require.Equal(t, amount, ledger.Refunded)
		require.Equal(t, []string{"prepare", "compensate", "accounting", "audit"}, ledger.Calls)
		if winner == replayTimeout {
			require.Equal(t, CompensationReasonTimeout, marker.CompensationReason)
			require.Equal(t, timeoutDigest(), marker.CallbackDigest)
			require.Nil(t, ackEvidence)
			require.Equal(t, []byte{1}, timeoutEvidence)
		} else {
			require.Equal(t, CompensationReasonErrorAck, marker.CompensationReason)
			require.Equal(t, acknowledgementDigest(replayAcknowledgement(winner)), marker.CallbackDigest)
			require.Equal(t, replayAcknowledgement(winner), ackEvidence)
			require.Nil(t, timeoutEvidence)
		}
	}
	assertReplayConservation(t, env)
}

func replaySnapshot(t *testing.T, env ibcTestEnv, packet channeltypes.Packet) []byte {
	t.Helper()
	store := env.ctx.KVStore(env.storeKey)
	snapshot := replayDurableSnapshot{
		Pending: append([]byte(nil), store.Get(PendingPacketKey(packet.GetSourceChannel(), packet.GetSequence()))...),
		Marker:  append([]byte(nil), store.Get(TerminalPacketKey(packet.GetSourceChannel(), packet.GetSequence()))...),
		Ack:     append([]byte(nil), store.Get(AckPacketKey(packet.GetSourceChannel(), packet.GetSequence()))...),
		Timeout: append([]byte(nil), store.Get(TimeoutPacketKey(packet.GetSourceChannel(), packet.GetSequence()))...),
		Ledger:  env.lifecycle.getLedger(env.ctx),
	}
	bz, err := json.Marshal(snapshot)
	require.NoError(t, err)
	return bz
}
