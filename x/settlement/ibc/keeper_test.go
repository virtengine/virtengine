// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
	"github.com/stretchr/testify/require"

	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
)

type mockSettlementKeeper struct {
	nextEscrowID int
	escrows      map[string]settlementtypes.EscrowAccount
	escrowsByOrd map[string]string
	settlements  map[string]settlementtypes.SettlementRecord
	releases     []releaseCall
	refunds      []releaseCall
}

type releaseCall struct {
	escrowID string
	reason   string
}

func newMockSettlementKeeper() *mockSettlementKeeper {
	return &mockSettlementKeeper{
		nextEscrowID: 1,
		escrows:      make(map[string]settlementtypes.EscrowAccount),
		escrowsByOrd: make(map[string]string),
		settlements:  make(map[string]settlementtypes.SettlementRecord),
		releases:     make([]releaseCall, 0),
		refunds:      make([]releaseCall, 0),
	}
}

func (m *mockSettlementKeeper) CreateEscrow(ctx sdk.Context, orderID string, depositor sdk.AccAddress, amount sdk.Coins, expiresIn time.Duration, conditions []settlementtypes.ReleaseCondition) (string, error) {
	escrowID := fmt.Sprintf("escrow-%d", m.nextEscrowID)
	m.nextEscrowID++

	escrow := settlementtypes.NewEscrowAccount(
		escrowID,
		orderID,
		depositor.String(),
		amount,
		ctx.BlockTime().Add(expiresIn),
		conditions,
		ctx.BlockTime(),
		ctx.BlockHeight(),
	)

	m.escrows[escrowID] = *escrow
	m.escrowsByOrd[orderID] = escrowID
	return escrowID, nil
}

func (m *mockSettlementKeeper) ReleaseEscrow(ctx sdk.Context, escrowID string, reason string) error {
	if _, found := m.escrows[escrowID]; !found {
		return fmt.Errorf("escrow %s not found", escrowID)
	}
	m.releases = append(m.releases, releaseCall{escrowID: escrowID, reason: reason})
	return nil
}

func (m *mockSettlementKeeper) RefundEscrow(ctx sdk.Context, escrowID string, reason string) error {
	if _, found := m.escrows[escrowID]; !found {
		return fmt.Errorf("escrow %s not found", escrowID)
	}
	m.refunds = append(m.refunds, releaseCall{escrowID: escrowID, reason: reason})
	return nil
}

func (m *mockSettlementKeeper) GetEscrow(ctx sdk.Context, escrowID string) (settlementtypes.EscrowAccount, bool) {
	escrow, found := m.escrows[escrowID]
	return escrow, found
}

func (m *mockSettlementKeeper) GetEscrowByOrder(ctx sdk.Context, orderID string) (settlementtypes.EscrowAccount, bool) {
	escrowID, found := m.escrowsByOrd[orderID]
	if !found {
		return settlementtypes.EscrowAccount{}, false
	}
	return m.GetEscrow(ctx, escrowID)
}

func (m *mockSettlementKeeper) SetSettlement(ctx sdk.Context, settlement settlementtypes.SettlementRecord) error {
	m.settlements[settlement.SettlementID] = settlement
	return nil
}

func (m *mockSettlementKeeper) GetSettlement(ctx sdk.Context, settlementID string) (settlementtypes.SettlementRecord, bool) {
	settlement, found := m.settlements[settlementID]
	return settlement, found
}

type mockChannelKeeper struct {
	sent    []sentPacket
	channel channeltypes.Channel
}

type sentPacket struct {
	sourcePort       string
	sourceChannel    string
	timeoutHeight    clienttypes.Height
	timeoutTimestamp uint64
	data             []byte
}

func (m *mockChannelKeeper) GetChannel(ctx sdk.Context, portID, channelID string) (channeltypes.Channel, bool) {
	return m.channel, true
}

func (m *mockChannelKeeper) SendPacket(ctx sdk.Context, sourcePort, sourceChannel string, timeoutHeight clienttypes.Height, timeoutTimestamp uint64, data []byte) (uint64, error) {
	sequence := uint64(len(m.sent) + 1) //nolint:gosec // test helper sequence
	m.sent = append(m.sent, sentPacket{
		sourcePort:       sourcePort,
		sourceChannel:    sourceChannel,
		timeoutHeight:    timeoutHeight,
		timeoutTimestamp: timeoutTimestamp,
		data:             data,
	})
	return sequence, nil
}

type mockPortKeeper struct {
	bound map[string]bool
}

func newMockPortKeeper() *mockPortKeeper {
	return &mockPortKeeper{bound: make(map[string]bool)}
}

func (m *mockPortKeeper) BindPort(ctx sdk.Context, portID string) error {
	m.bound[portID] = true
	return nil
}

func (m *mockPortKeeper) Route(module string) (porttypes.IBCModule, bool) {
	if m.bound[module] {
		return nil, true
	}
	return nil, false
}

type ibcTestEnv struct {
	ctx       sdk.Context
	keeper    IBCKeeper
	lifecycle *testLifecycleHooks
	settle    *mockSettlementKeeper
	channel   *mockChannelKeeper
	port      *mockPortKeeper
	storeKey  storetypes.StoreKey
	codec     codec.BinaryCodec
}

type testConservationLedger struct {
	Available int64    `json:"available"`
	Held      int64    `json:"held"`
	Committed int64    `json:"committed"`
	Refunded  int64    `json:"refunded"`
	Calls     []string `json:"calls"`
}

type testLifecycleHooks struct {
	storeKey storetypes.StoreKey
	failAt   string
}

var testLifecycleLedgerKey = []byte{0xf0}

func (h *testLifecycleHooks) PrepareTransfer(ctx sdk.Context, packet SettlementPacketData) (string, error) {
	if h.failAt == "prepare" {
		return "", fmt.Errorf("injected prepare failure")
	}
	ledger := h.getLedger(ctx)
	if ledger.Available == 0 && ledger.Held == 0 && ledger.Committed == 0 && ledger.Refunded == 0 {
		ledger.Available = 1000
	}
	amount := packetAmount(packet)
	if amount <= 0 || ledger.Available < amount {
		return "", fmt.Errorf("insufficient test custody")
	}
	ledger.Available -= amount
	ledger.Held += amount
	ledger.Calls = append(ledger.Calls, "prepare")
	h.setLedger(ctx, ledger)
	return "custody/deposit-1", nil
}

func (h *testLifecycleHooks) FinalizeTransfer(ctx sdk.Context, _ PendingPacket, _ TransferTransition) error {
	if h.failAt == "finalize" {
		return fmt.Errorf("injected finalize failure")
	}
	ledger := h.getLedger(ctx)
	ledger.Committed += ledger.Held
	ledger.Held = 0
	ledger.Calls = append(ledger.Calls, "finalize")
	h.setLedger(ctx, ledger)
	return nil
}

func (h *testLifecycleHooks) CompensateTransfer(ctx sdk.Context, _ PendingPacket, _ TransferTransition) error {
	if h.failAt == "compensate" {
		return fmt.Errorf("injected compensate failure")
	}
	ledger := h.getLedger(ctx)
	ledger.Refunded += ledger.Held
	ledger.Held = 0
	ledger.Calls = append(ledger.Calls, "compensate")
	h.setLedger(ctx, ledger)
	return nil
}

func (h *testLifecycleHooks) RecordAccounting(ctx sdk.Context, _ PendingPacket, _ TransferTransition) error {
	if h.failAt == "accounting" {
		return fmt.Errorf("injected accounting failure")
	}
	ledger := h.getLedger(ctx)
	ledger.Calls = append(ledger.Calls, "accounting")
	h.setLedger(ctx, ledger)
	return nil
}

func (h *testLifecycleHooks) RecordAudit(ctx sdk.Context, _ PendingPacket, _ TransferTransition) error {
	if h.failAt == "audit" {
		return fmt.Errorf("injected audit failure")
	}
	ledger := h.getLedger(ctx)
	ledger.Calls = append(ledger.Calls, "audit")
	h.setLedger(ctx, ledger)
	return nil
}

func (h *testLifecycleHooks) getLedger(ctx sdk.Context) testConservationLedger {
	bz := ctx.KVStore(h.storeKey).Get(testLifecycleLedgerKey)
	if bz == nil {
		return testConservationLedger{}
	}
	var ledger testConservationLedger
	if err := json.Unmarshal(bz, &ledger); err != nil {
		panic(err)
	}
	return ledger
}

func (h *testLifecycleHooks) setLedger(ctx sdk.Context, ledger testConservationLedger) {
	bz, err := json.Marshal(ledger)
	if err != nil {
		panic(err)
	}
	ctx.KVStore(h.storeKey).Set(testLifecycleLedgerKey, bz)
}

func packetAmount(packet SettlementPacketData) int64 {
	if packet.Type != PacketTypeEscrowDeposit {
		return 1
	}
	var deposit EscrowDepositPacket
	if err := json.Unmarshal(packet.Data, &deposit); err != nil || len(deposit.Amount) != 1 {
		return 0
	}
	return deposit.Amount[0].Amount.Int64()
}

func setupIBCTestEnv(t *testing.T) ibcTestEnv {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey("settlement")
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Height: 10,
		Time:   time.Unix(1700000000, 0).UTC(),
	}, false, log.NewNopLogger())

	settle := newMockSettlementKeeper()
	channel := &mockChannelKeeper{
		channel: channeltypes.Channel{Version: Version},
	}
	port := newMockPortKeeper()

	keeper := NewIBCKeeper(cdc, storeKey, settle, channel, port)
	lifecycle := &testLifecycleHooks{storeKey: storeKey}
	keeper.SetTransferLifecycleHooks(lifecycle)

	return ibcTestEnv{
		ctx:       ctx,
		keeper:    keeper,
		lifecycle: lifecycle,
		settle:    settle,
		channel:   channel,
		port:      port,
		storeKey:  storeKey,
		codec:     cdc,
	}
}

func TestIBCKeeperSendEscrowDepositDefaults(t *testing.T) {
	env := setupIBCTestEnv(t)

	depositor := sdk.AccAddress([]byte("depositor_addr______"))
	deposit := EscrowDepositPacket{
		DepositID:        "deposit-1",
		OrderID:          "order-1",
		Depositor:        depositor.String(),
		Amount:           sdk.NewCoins(sdk.NewInt64Coin("uve", 1000)),
		ExpiresInSeconds: 3600,
		SourceChainID:    "chain-a",
		SourceChannel:    "channel-0",
		RequestedAt:      env.ctx.BlockTime(),
	}

	sequence, err := env.keeper.SendEscrowDepositPacket(env.ctx, "channel-0", clienttypes.Height{}, 0, deposit)
	require.NoError(t, err)
	require.Equal(t, uint64(1), sequence)
	require.Len(t, env.channel.sent, 1)

	sent := env.channel.sent[0]
	expectedHeight := uint64(env.ctx.BlockHeight()) + DefaultTimeoutHeightDelta                //nolint:gosec // test helper
	expectedTimestamp := uint64(env.ctx.BlockTime().UnixNano()) + DefaultTimeoutTimestampDelta //nolint:gosec // test helper
	require.Equal(t, expectedHeight, sent.timeoutHeight.RevisionHeight)
	require.Equal(t, expectedTimestamp, sent.timeoutTimestamp)

	pending, found := env.keeper.getPendingPacket(env.ctx, "channel-0", sequence)
	require.True(t, found)
	require.NoError(t, pending.Identity.Validate())
	require.Equal(t, "deposit-1", pending.Identity.LogicalPayoutID)
	require.Equal(t, payloadDigest(sent.data), pending.Identity.PayloadDigest)
	require.Equal(t, TransferStatePending, pending.State)
}

func TestIBCKeeperTerminalCallbacks(t *testing.T) {
	tests := []struct {
		name       string
		first      string
		second     string
		secondAck  Acknowledgement
		wantErr    error
		wantState  TransferState
		wantReason CompensationReason
	}{
		{name: "exact acknowledgement retry", first: "success", second: "success", wantState: TransferStateFinalized},
		{name: "conflicting acknowledgement", first: "success", second: "ack", secondAck: NewErrorAcknowledgement(fmt.Errorf("rejected")), wantErr: ErrTerminalConflict, wantState: TransferStateFinalized},
		{name: "late acknowledgement after timeout", first: "timeout", second: "success", wantErr: ErrTerminalConflict, wantState: TransferStateCompensated, wantReason: CompensationReasonTimeout},
		{name: "reordered timeout after acknowledgement", first: "success", second: "timeout", wantErr: ErrTerminalConflict, wantState: TransferStateFinalized},
		{name: "duplicate timeout", first: "timeout", second: "timeout", wantState: TransferStateCompensated, wantReason: CompensationReasonTimeout},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env := setupIBCTestEnv(t)
			deposit := validDepositPacket(env.ctx.BlockTime())
			sequence, err := env.keeper.SendEscrowDepositPacket(env.ctx, "channel-0", clienttypes.Height{}, 0, deposit)
			require.NoError(t, err)
			packet := channeltypes.NewPacket(env.channel.sent[0].data, sequence, PortID, "channel-0", PortID, "channel-1", clienttypes.NewHeight(0, 20), 0)
			relayer := sdk.AccAddress([]byte("relayer_addr________"))
			successAck := NewResultAcknowledgement(EscrowDepositAck{EscrowID: "escrow-1", Status: "created"})

			runCallback := func(kind string, ack Acknowledgement) error {
				switch kind {
				case "success":
					return env.keeper.OnAcknowledgementPacket(env.ctx, packet, successAck.GetBytes(), relayer)
				case "ack":
					return env.keeper.OnAcknowledgementPacket(env.ctx, packet, ack.GetBytes(), relayer)
				case "timeout":
					return env.keeper.OnTimeoutPacket(env.ctx, packet, relayer)
				default:
					t.Fatalf("unknown callback kind %q", kind)
					return nil
				}
			}

			require.NoError(t, runCallback(test.first, test.secondAck))
			err = runCallback(test.second, test.secondAck)
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
			}

			marker, found := env.keeper.getTerminalMarker(env.ctx, "channel-0", sequence)
			require.True(t, found)
			require.Equal(t, test.wantState, marker.State)
			require.Equal(t, test.wantReason, marker.CompensationReason)
		})
	}
}

func TestIBCKeeperLifecycleHooksFailClosed(t *testing.T) {
	env := setupIBCTestEnv(t)
	env.keeper.SetTransferLifecycleHooks(nil)

	_, err := env.keeper.SendEscrowDepositPacket(env.ctx, "channel-0", clienttypes.Height{}, 0, validDepositPacket(env.ctx.BlockTime()))
	require.ErrorIs(t, err, ErrLifecycleHooksUnavailable)
	require.Empty(t, env.channel.sent)
}

func TestIBCKeeperAtomicTransitionRollbackAndRetry(t *testing.T) {
	tests := []struct {
		name          string
		callback      string
		failAt        string
		wantCommitted int64
		wantRefunded  int64
	}{
		{name: "finalize custody failure", callback: "success", failAt: "finalize", wantCommitted: 1000},
		{name: "compensation custody failure", callback: "timeout", failAt: "compensate", wantRefunded: 1000},
		{name: "accounting failure", callback: "success", failAt: "accounting", wantCommitted: 1000},
		{name: "audit failure", callback: "timeout", failAt: "audit", wantRefunded: 1000},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env := setupIBCTestEnv(t)
			sequence, err := env.keeper.SendEscrowDepositPacket(env.ctx, "channel-0", clienttypes.Height{}, 0, validDepositPacket(env.ctx.BlockTime()))
			require.NoError(t, err)
			packet := channeltypes.NewPacket(env.channel.sent[0].data, sequence, PortID, "channel-0", PortID, "channel-1", clienttypes.NewHeight(0, 20), 0)
			relayer := sdk.AccAddress([]byte("relayer_addr________"))
			ack := NewResultAcknowledgement(EscrowDepositAck{EscrowID: "escrow-1", Status: "created"})
			runCallback := func() error {
				if test.callback == "timeout" {
					return env.keeper.OnTimeoutPacket(env.ctx, packet, relayer)
				}
				return env.keeper.OnAcknowledgementPacket(env.ctx, packet, ack.GetBytes(), relayer)
			}

			env.lifecycle.failAt = test.failAt
			require.Error(t, runCallback())
			pending, found := env.keeper.getPendingPacket(env.ctx, "channel-0", sequence)
			require.True(t, found)
			require.Equal(t, TransferStatePending, pending.State)
			_, found = env.keeper.getTerminalMarker(env.ctx, "channel-0", sequence)
			require.False(t, found)
			store := env.ctx.KVStore(env.storeKey)
			require.Nil(t, store.Get(AckPacketKey("channel-0", sequence)))
			require.Nil(t, store.Get(TimeoutPacketKey("channel-0", sequence)))
			ledger := env.lifecycle.getLedger(env.ctx)
			require.Equal(t, testConservationLedger{Available: 0, Held: 1000, Calls: []string{"prepare"}}, ledger)

			env.lifecycle.failAt = ""
			require.NoError(t, runCallback())
			require.NoError(t, runCallback())
			_, found = env.keeper.getPendingPacket(env.ctx, "channel-0", sequence)
			require.False(t, found)
			_, found = env.keeper.getTerminalMarker(env.ctx, "channel-0", sequence)
			require.True(t, found)
			if test.callback == "timeout" {
				require.NotNil(t, store.Get(TimeoutPacketKey("channel-0", sequence)))
				require.Nil(t, store.Get(AckPacketKey("channel-0", sequence)))
			} else {
				require.NotNil(t, store.Get(AckPacketKey("channel-0", sequence)))
				require.Nil(t, store.Get(TimeoutPacketKey("channel-0", sequence)))
			}
			ledger = env.lifecycle.getLedger(env.ctx)
			require.Equal(t, int64(1000), ledger.Available+ledger.Held+ledger.Committed+ledger.Refunded)
			require.Equal(t, test.wantCommitted, ledger.Committed)
			require.Equal(t, test.wantRefunded, ledger.Refunded)
			custodyCall := "finalize"
			if test.callback == "timeout" {
				custodyCall = "compensate"
			}
			require.Equal(t, []string{"prepare", custodyCall, "accounting", "audit"}, ledger.Calls)
		})
	}
}

func TestIBCKeeperCallbacksFailClosed(t *testing.T) {
	env := setupIBCTestEnv(t)
	relayer := sdk.AccAddress([]byte("relayer_addr________"))
	ack := NewResultAcknowledgement(EscrowDepositAck{EscrowID: "escrow-1", Status: "created"})
	unknown := channeltypes.NewPacket([]byte(`{"unknown":true}`), 99, PortID, "channel-0", PortID, "channel-1", clienttypes.NewHeight(0, 20), 0)
	require.ErrorIs(t, env.keeper.OnAcknowledgementPacket(env.ctx, unknown, ack.GetBytes(), relayer), ErrPendingPacketNotFound)

	sequence, err := env.keeper.SendEscrowDepositPacket(env.ctx, "channel-0", clienttypes.Height{}, 0, validDepositPacket(env.ctx.BlockTime()))
	require.NoError(t, err)
	packet := channeltypes.NewPacket(env.channel.sent[0].data, sequence, PortID, "channel-0", PortID, "channel-1", clienttypes.NewHeight(0, 20), 0)
	require.ErrorIs(t, env.keeper.OnAcknowledgementPacket(env.ctx, packet, Acknowledgement{}.GetBytes(), relayer), ErrInvalidPacket)

	mismatched := channeltypes.NewPacket(packet.GetData(), sequence, "other-port", "channel-0", PortID, "channel-1", clienttypes.NewHeight(0, 20), 0)
	require.ErrorIs(t, env.keeper.OnTimeoutPacket(env.ctx, mismatched, relayer), ErrPacketIdentityMismatch)
	pending, found := env.keeper.getPendingPacket(env.ctx, "channel-0", sequence)
	require.True(t, found)
	require.Equal(t, TransferStatePending, pending.State)
}

func validDepositPacket(now time.Time) EscrowDepositPacket {
	depositor := sdk.AccAddress([]byte("depositor_addr______"))
	return EscrowDepositPacket{
		DepositID:        "deposit-1",
		OrderID:          "order-1",
		Depositor:        depositor.String(),
		Amount:           sdk.NewCoins(sdk.NewInt64Coin("uve", 1000)),
		ExpiresInSeconds: 3600,
		SourceChainID:    "chain-a",
		SourceChannel:    "channel-0",
		RequestedAt:      now,
	}
}

func TestIBCKeeperOnRecvEscrowDepositFailsClosed(t *testing.T) {
	env := setupIBCTestEnv(t)

	depositor := sdk.AccAddress([]byte("depositor_addr______"))
	deposit := EscrowDepositPacket{
		DepositID:        "deposit-1",
		OrderID:          "order-1",
		Depositor:        depositor.String(),
		Amount:           sdk.NewCoins(sdk.NewInt64Coin("uve", 1000)),
		ExpiresInSeconds: 3600,
		SourceChainID:    "chain-a",
		SourceChannel:    "channel-0",
		RequestedAt:      env.ctx.BlockTime(),
	}

	packetData, err := NewPacketData(PacketTypeEscrowDeposit, deposit)
	require.NoError(t, err)

	packet := channeltypes.NewPacket(
		packetData.GetBytes(),
		1,
		PortID,
		"channel-0",
		PortID,
		"channel-1",
		clienttypes.NewHeight(0, 20),
		0,
	)

	relayer := sdk.AccAddress([]byte("relayer_addr________"))
	ack := env.keeper.OnRecvPacket(env.ctx, packet, relayer)

	ackBz := ack.Acknowledgement()
	var parsed Acknowledgement
	require.NoError(t, json.Unmarshal(ackBz, &parsed))
	require.False(t, parsed.Success())
	require.Contains(t, parsed.Error, ErrInboundDepositUnauthorized.Error())

	_, found := env.settle.GetEscrowByOrder(env.ctx, deposit.OrderID)
	require.False(t, found)
	require.Empty(t, env.ctx.EventManager().Events())
}

func TestIBCKeeperOnRecvEscrowRelease(t *testing.T) {
	env := setupIBCTestEnv(t)

	escrowID, _ := env.settle.CreateEscrow(env.ctx, "order-1", sdk.AccAddress([]byte("depositor_addr______")), sdk.NewCoins(sdk.NewInt64Coin("uve", 1000)), time.Hour, nil)

	release := EscrowReleasePacket{
		EscrowID:      escrowID,
		OrderID:       "order-1",
		ReleaseType:   ReleaseTypeRelease,
		Reason:        "completed",
		SourceChainID: "chain-a",
		SourceChannel: "channel-0",
		RequestedAt:   env.ctx.BlockTime(),
	}

	packetData, err := NewPacketData(PacketTypeEscrowRelease, release)
	require.NoError(t, err)

	packet := channeltypes.NewPacket(
		packetData.GetBytes(),
		2,
		PortID,
		"channel-0",
		PortID,
		"channel-1",
		clienttypes.NewHeight(0, 20),
		0,
	)

	ack := env.keeper.OnRecvPacket(env.ctx, packet, sdk.AccAddress([]byte("relayer_addr________")))
	var parsed Acknowledgement
	require.NoError(t, json.Unmarshal(ack.Acknowledgement(), &parsed))
	require.True(t, parsed.Success())
	require.Len(t, env.settle.releases, 1)
	require.Equal(t, escrowID, env.settle.releases[0].escrowID)
}

func TestIBCKeeperOnRecvSettlementRecord(t *testing.T) {
	env := setupIBCTestEnv(t)

	provider := sdk.AccAddress([]byte("provider_addr_______")).String()
	customer := sdk.AccAddress([]byte("customer_addr_______")).String()
	amount := sdk.NewCoins(sdk.NewInt64Coin("uve", 1000))
	record := settlementtypes.NewSettlementRecord(
		"settle-1",
		"escrow-1",
		"order-1",
		"lease-1",
		provider,
		customer,
		amount,
		amount,
		sdk.NewCoins(),
		sdk.NewCoins(),
		nil,
		0,
		env.ctx.BlockTime().Add(-time.Hour),
		env.ctx.BlockTime(),
		settlementtypes.SettlementTypeFinal,
		true,
		env.ctx.BlockTime(),
		env.ctx.BlockHeight(),
	)

	packetRecord := SettlementRecordPacket{
		Record:        *record,
		SourceChainID: "chain-a",
		SourceChannel: "channel-0",
	}

	packetData, err := NewPacketData(PacketTypeSettlementRecord, packetRecord)
	require.NoError(t, err)

	packet := channeltypes.NewPacket(
		packetData.GetBytes(),
		3,
		PortID,
		"channel-0",
		PortID,
		"channel-1",
		clienttypes.NewHeight(0, 20),
		0,
	)

	ack := env.keeper.OnRecvPacket(env.ctx, packet, sdk.AccAddress([]byte("relayer_addr________")))
	var parsed Acknowledgement
	require.NoError(t, json.Unmarshal(ack.Acknowledgement(), &parsed))
	require.True(t, parsed.Success())

	_, found := env.settle.GetSettlement(env.ctx, record.SettlementID)
	require.True(t, found)
}

func TestIBCKeeperRateLimit(t *testing.T) {
	env := setupIBCTestEnv(t)

	cfg := RateLimitConfig{
		Enabled:                      true,
		MaxPacketsPerBlock:           1,
		MaxPacketsPerRelayerPerBlock: 1,
	}
	require.NoError(t, env.keeper.SetRateLimitConfig(env.ctx, cfg))

	relayer := sdk.AccAddress([]byte("relayer_addr________"))
	require.NoError(t, env.keeper.CheckRateLimit(env.ctx, relayer, PacketTypeEscrowDeposit))
	require.Error(t, env.keeper.CheckRateLimit(env.ctx, relayer, PacketTypeEscrowDeposit))
}

func TestIBCKeeperHandshakeTimeout(t *testing.T) {
	env := setupIBCTestEnv(t)

	env.keeper.StoreHandshakeRecord(env.ctx, "channel-0")
	require.NoError(t, env.keeper.CheckHandshakeTimeout(env.ctx, "channel-0"))

	lateCtx := env.ctx.WithBlockHeight(env.ctx.BlockHeight() + 101)
	require.ErrorIs(t, env.keeper.CheckHandshakeTimeout(lateCtx, "channel-0"), ErrHandshakeTimedOut)

	env.keeper.ClearHandshakeRecord(env.ctx, "channel-0")
	env.keeper.StoreHandshakeRecord(env.ctx, "channel-1")
	lateTime := env.ctx.BlockTime().Add(16 * time.Minute)
	lateCtx = env.ctx.WithBlockTime(lateTime)
	require.ErrorIs(t, env.keeper.CheckHandshakeTimeout(lateCtx, "channel-1"), ErrHandshakeTimedOut)
}

func TestIBCKeeperRejectsSemanticallyCorruptTerminalMarker(t *testing.T) {
	env := setupIBCTestEnv(t)
	sequence, err := env.keeper.SendEscrowDepositPacket(env.ctx, "channel-0", clienttypes.Height{}, 0, validDepositPacket(env.ctx.BlockTime()))
	require.NoError(t, err)
	packet := channeltypes.NewPacket(
		env.channel.sent[0].data, sequence, PortID, "channel-0", PortID, "channel-1",
		clienttypes.NewHeight(0, 20), 0,
	)
	pending, found := env.keeper.getPendingPacket(env.ctx, "channel-0", sequence)
	require.True(t, found)
	ack := NewResultAcknowledgement(EscrowDepositAck{EscrowID: "escrow-1", OrderID: "order-1", Status: "success"})
	marker := TerminalMarker{
		Identity: pending.Identity, State: TransferStateRecoveryRequired,
		CallbackDigest: acknowledgementDigest(ack.GetBytes()), TransitionedAt: env.ctx.BlockTime(),
	}
	env.keeper.setTerminalMarker(env.ctx, marker)
	relayer := sdk.AccAddress([]byte("relayer_addr________"))

	require.ErrorIs(t, env.keeper.OnAcknowledgementPacket(env.ctx, packet, ack.GetBytes(), relayer), ErrTerminalConflict)
	_, found = env.keeper.getPendingPacket(env.ctx, "channel-0", sequence)
	require.True(t, found)
	require.Nil(t, env.ctx.KVStore(env.storeKey).Get(AckPacketKey("channel-0", sequence)))
	require.NotContains(t, env.lifecycle.getLedger(env.ctx).Calls, "finalize")
}

func TestIBCModuleRejectsAndRetainsMalformedHandshakeRecord(t *testing.T) {
	env := setupIBCTestEnv(t)
	key := HandshakeKey("channel-0")
	store := env.ctx.KVStore(env.storeKey)
	store.Set(key, []byte("not-json"))
	module := NewIBCModule(env.keeper)

	err := module.OnChanOpenConfirm(env.ctx, PortID, "channel-0")
	require.ErrorIs(t, err, ErrInvalidPacket)
	require.Equal(t, []byte("not-json"), store.Get(key))
}

func TestIBCKeeperBindPort(t *testing.T) {
	env := setupIBCTestEnv(t)

	require.False(t, env.keeper.IsBound(env.ctx))
	require.NoError(t, env.keeper.BindPort(env.ctx))
	require.True(t, env.keeper.IsBound(env.ctx))
}
