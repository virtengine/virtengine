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
	sequence := uint64(len(m.sent) + 1)
	m.sent = append(m.sent, sentPacket{
		sourcePort:       sourcePort,
		sourceChannel:    sourceChannel,
		timeoutHeight:    timeoutHeight,
		timeoutTimestamp: timeoutTimestamp,
		data:             data,
	})
	return sequence, nil
}

type ibcTestEnv struct {
	ctx      sdk.Context
	keeper   IBCKeeper
	settle   *mockSettlementKeeper
	channel  *mockChannelKeeper
	storeKey storetypes.StoreKey
	codec    codec.BinaryCodec
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
	keeper := NewIBCKeeper(cdc, storeKey, settle, channel, nil)

	return ibcTestEnv{
		ctx:      ctx,
		keeper:   keeper,
		settle:   settle,
		channel:  channel,
		storeKey: storeKey,
		codec:    cdc,
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
	require.Equal(t, uint64(env.ctx.BlockHeight())+DefaultTimeoutHeightDelta, sent.timeoutHeight.RevisionHeight)
	require.Equal(t, uint64(env.ctx.BlockTime().UnixNano())+DefaultTimeoutTimestampDelta, sent.timeoutTimestamp)
}

func TestIBCKeeperOnRecvEscrowDeposit(t *testing.T) {
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
	require.True(t, parsed.Success())

	_, found := env.settle.GetEscrowByOrder(env.ctx, deposit.OrderID)
	require.True(t, found)
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

func TestIBCKeeperBindPort(t *testing.T) {
	env := setupIBCTestEnv(t)

	require.True(t, env.keeper.IsBound(env.ctx))
	require.NoError(t, env.keeper.BindPort(env.ctx))
	require.True(t, env.keeper.IsBound(env.ctx))
}
