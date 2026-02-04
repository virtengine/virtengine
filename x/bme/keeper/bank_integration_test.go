// Copyright (c) VirtEngine Inc.
// SPDX-License-Identifier: BUSL-1.1

package keeper_test

import (
	"context"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	types "github.com/virtengine/virtengine/sdk/go/node/bme/v1"
	"github.com/virtengine/virtengine/x/bme/keeper"
)

const testOrderID = "order-1"

// MockBankKeeper implements the BankKeeper interface for testing.
type MockBankKeeper struct {
	balances map[string]sdk.Coins
	modules  map[string]sdk.Coins
}

func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{
		balances: make(map[string]sdk.Coins),
		modules:  make(map[string]sdk.Coins),
	}
}

func (m *MockBankKeeper) SpendableCoins(_ context.Context, addr sdk.AccAddress) sdk.Coins {
	coins := m.balances[addr.String()]
	if coins == nil {
		return sdk.Coins{}
	}
	return coins
}

func (m *MockBankKeeper) SendCoins(_ context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error {
	from := fromAddr.String()
	to := toAddr.String()

	if !m.balances[from].IsAllGTE(amt) {
		return keeper.ErrInsufficientFunds
	}

	m.balances[from] = m.balances[from].Sub(amt...)
	m.balances[to] = m.balances[to].Add(amt...)
	return nil
}

func (m *MockBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	if !m.modules[senderModule].IsAllGTE(amt) {
		return keeper.ErrInsufficientFunds
	}

	m.modules[senderModule] = m.modules[senderModule].Sub(amt...)
	to := recipientAddr.String()
	m.balances[to] = m.balances[to].Add(amt...)
	return nil
}

func (m *MockBankKeeper) SendCoinsFromAccountToModule(_ context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	from := senderAddr.String()
	if !m.balances[from].IsAllGTE(amt) {
		return keeper.ErrInsufficientFunds
	}

	m.balances[from] = m.balances[from].Sub(amt...)
	m.modules[recipientModule] = m.modules[recipientModule].Add(amt...)
	return nil
}

func (m *MockBankKeeper) SendCoinsFromModuleToModule(_ context.Context, senderModule, recipientModule string, amt sdk.Coins) error {
	if !m.modules[senderModule].IsAllGTE(amt) {
		return keeper.ErrInsufficientFunds
	}

	m.modules[senderModule] = m.modules[senderModule].Sub(amt...)
	m.modules[recipientModule] = m.modules[recipientModule].Add(amt...)
	return nil
}

func (m *MockBankKeeper) MintCoins(_ context.Context, moduleName string, amt sdk.Coins) error {
	m.modules[moduleName] = m.modules[moduleName].Add(amt...)
	return nil
}

func (m *MockBankKeeper) BurnCoins(_ context.Context, moduleName string, amt sdk.Coins) error {
	if !m.modules[moduleName].IsAllGTE(amt) {
		return keeper.ErrInsufficientFunds
	}
	m.modules[moduleName] = m.modules[moduleName].Sub(amt...)
	return nil
}

func (m *MockBankKeeper) GetBalance(_ context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	coins := m.balances[addr.String()]
	return sdk.NewCoin(denom, coins.AmountOf(denom))
}

// SetBalance sets the balance for an address (test helper).
func (m *MockBankKeeper) SetBalance(addr sdk.AccAddress, coins sdk.Coins) {
	m.balances[addr.String()] = coins
}

// SetModuleBalance sets the balance for a module (test helper).
func (m *MockBankKeeper) SetModuleBalance(module string, coins sdk.Coins) {
	m.modules[module] = coins
}

// GetModuleBalance returns the module balance (test helper).
func (m *MockBankKeeper) GetModuleBalance(module string) sdk.Coins {
	return m.modules[module]
}

// Ensure MockBankKeeper implements the interface
var _ keeper.BankKeeper = (*MockBankKeeper)(nil)

func setupKeeperWithBankMock(t testing.TB) (keeper.IKeeper, sdk.Context, *MockBankKeeper) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	bankKeeper := NewMockBankKeeper()
	k := keeper.NewKeeper(cdc, storeKey, authority, bankKeeper, nil)
	ctx := sdk.NewContext(stateStore, cmtproto.Header{Height: 1}, false, log.NewNopLogger())

	// Initialize with default params
	err := k.SetParams(ctx, types.DefaultParams())
	require.NoError(t, err)

	return k, ctx, bankKeeper
}

func TestHoldEscrow(t *testing.T) {
	k, ctx, bankKeeper := setupKeeperWithBankMock(t)

	depositor := sdk.AccAddress([]byte("depositor1234567890"))
	orderID := testOrderID
	amount := sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(1000)))

	// Set depositor balance
	bankKeeper.SetBalance(depositor, amount)

	// Hold escrow
	err := k.HoldEscrow(ctx, orderID, depositor, amount)
	require.NoError(t, err)

	// Verify escrow balance
	escrowBalance := k.GetEscrowBalance(ctx, orderID)
	require.True(t, escrowBalance.Equal(amount))

	// Verify depositor balance is zero
	depositorBalance := bankKeeper.SpendableCoins(ctx, depositor)
	require.True(t, depositorBalance.IsZero())

	// Verify module received the coins
	moduleBalance := bankKeeper.GetModuleBalance(types.ModuleName)
	require.True(t, moduleBalance.Equal(amount))
}

func TestReleaseEscrow(t *testing.T) {
	k, ctx, bankKeeper := setupKeeperWithBankMock(t)

	depositor := sdk.AccAddress([]byte("depositor1234567890"))
	recipient := sdk.AccAddress([]byte("recipient123456789"))
	orderID := testOrderID
	amount := sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(1000)))

	// Set depositor balance and hold escrow
	bankKeeper.SetBalance(depositor, amount)
	err := k.HoldEscrow(ctx, orderID, depositor, amount)
	require.NoError(t, err)

	// Release escrow to recipient
	err = k.ReleaseEscrow(ctx, orderID, recipient)
	require.NoError(t, err)

	// Verify recipient received the coins
	recipientBalance := bankKeeper.SpendableCoins(ctx, recipient)
	require.True(t, recipientBalance.Equal(amount))

	// Verify escrow is empty
	escrowBalance := k.GetEscrowBalance(ctx, orderID)
	require.True(t, escrowBalance.IsZero())
}

func TestHoldEscrowInsufficientFunds(t *testing.T) {
	k, ctx, bankKeeper := setupKeeperWithBankMock(t)

	depositor := sdk.AccAddress([]byte("depositor1234567890"))
	orderID := testOrderID
	amount := sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(1000)))

	// Set depositor balance to less than required
	bankKeeper.SetBalance(depositor, sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(500))))

	// Hold escrow should fail
	err := k.HoldEscrow(ctx, orderID, depositor, amount)
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient funds")
}

func TestMintTokens(t *testing.T) {
	k, ctx, bankKeeper := setupKeeperWithBankMock(t)

	recipient := sdk.AccAddress([]byte("recipient123456789"))
	amount := sdk.NewCoins(sdk.NewCoin("uvact", math.NewInt(500)))

	// Mint tokens
	err := k.MintTokens(ctx, recipient, amount)
	require.NoError(t, err)

	// Verify recipient received the minted coins
	recipientBalance := bankKeeper.SpendableCoins(ctx, recipient)
	require.True(t, recipientBalance.Equal(amount))

	// Verify vault state updated
	state := k.GetState(ctx)
	require.True(t, state.TotalMinted.Equal(amount))
}

func TestBurnTokens(t *testing.T) {
	k, ctx, bankKeeper := setupKeeperWithBankMock(t)

	from := sdk.AccAddress([]byte("burner1234567890xx"))
	amount := sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(300)))

	// Set balance
	bankKeeper.SetBalance(from, amount)

	// Burn tokens
	err := k.BurnTokens(ctx, from, amount)
	require.NoError(t, err)

	// Verify balance is zero
	balance := bankKeeper.SpendableCoins(ctx, from)
	require.True(t, balance.IsZero())

	// Verify vault state updated
	state := k.GetState(ctx)
	require.True(t, state.TotalBurned.Equal(amount))
}

func TestCollectFees(t *testing.T) {
	k, ctx, bankKeeper := setupKeeperWithBankMock(t)

	payer := sdk.AccAddress([]byte("payer12345678901234"))
	amount := sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(100)))

	// Set payer balance
	bankKeeper.SetBalance(payer, amount)

	// Collect fees
	err := k.CollectFees(ctx, payer, amount)
	require.NoError(t, err)

	// Verify payer balance is zero
	payerBalance := bankKeeper.SpendableCoins(ctx, payer)
	require.True(t, payerBalance.IsZero())

	// Verify fees went to distribution module
	feeBalance := bankKeeper.GetModuleBalance("distribution")
	require.True(t, feeBalance.Equal(amount))
}

func TestSettleBilling(t *testing.T) {
	k, ctx, bankKeeper := setupKeeperWithBankMock(t)

	depositor := sdk.AccAddress([]byte("depositor1234567890"))
	provider := sdk.AccAddress([]byte("provider1234567890x"))
	orderID := testOrderID
	leaseID := "lease-1"
	depositAmount := sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(10000)))
	usageAmount := sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(1000)))

	// Set depositor balance and hold escrow
	bankKeeper.SetBalance(depositor, depositAmount)
	err := k.HoldEscrow(ctx, orderID, depositor, depositAmount)
	require.NoError(t, err)

	// Register lease escrow
	keeperPtr := k.(*keeper.Keeper)
	err = keeperPtr.RegisterLeaseEscrow(ctx, leaseID, orderID, provider, depositor)
	require.NoError(t, err)

	// Settle billing
	err = k.SettleBilling(ctx, leaseID, provider, usageAmount)
	require.NoError(t, err)

	// Verify provider received payment (minus fees)
	providerBalance := bankKeeper.SpendableCoins(ctx, provider)
	require.False(t, providerBalance.IsZero())

	// Verify escrow still has remaining funds
	escrowBalance := k.GetEscrowBalance(ctx, orderID)
	require.False(t, escrowBalance.IsZero())
	require.True(t, escrowBalance.IsAllLT(depositAmount))
}
