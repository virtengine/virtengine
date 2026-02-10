// Copyright (c) VirtEngine Inc.
// SPDX-License-Identifier: BUSL-1.1

package keeper

import (
	"context"
	"testing"
	"time"

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
	"github.com/stretchr/testify/require"

	types "github.com/virtengine/virtengine/sdk/go/node/oracle/v1"
	otypes "github.com/virtengine/virtengine/x/oracle/types"
)

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
		return ErrInsufficientFunds
	}

	m.balances[from] = m.balances[from].Sub(amt...)
	m.balances[to] = m.balances[to].Add(amt...)
	return nil
}

func (m *MockBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	if !m.modules[senderModule].IsAllGTE(amt) {
		return ErrInsufficientFunds
	}

	m.modules[senderModule] = m.modules[senderModule].Sub(amt...)
	to := recipientAddr.String()
	m.balances[to] = m.balances[to].Add(amt...)
	return nil
}

func (m *MockBankKeeper) SendCoinsFromAccountToModule(_ context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	from := senderAddr.String()
	if !m.balances[from].IsAllGTE(amt) {
		return ErrInsufficientFunds
	}

	m.balances[from] = m.balances[from].Sub(amt...)
	m.modules[recipientModule] = m.modules[recipientModule].Add(amt...)
	return nil
}

func (m *MockBankKeeper) SendCoinsFromModuleToModule(_ context.Context, senderModule, recipientModule string, amt sdk.Coins) error {
	if !m.modules[senderModule].IsAllGTE(amt) {
		return ErrInsufficientFunds
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
		return ErrInsufficientFunds
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
var _ BankKeeper = (*MockBankKeeper)(nil)

func setupKeeperWithBankMock(t testing.TB) (IKeeper, sdk.Context, *MockBankKeeper) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(otypes.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	bankKeeper := NewMockBankKeeper()
	k := NewKeeper(cdc, storeKey, "test-authority", bankKeeper)
	ctx := sdk.NewContext(stateStore, cmtproto.Header{Height: 1, Time: time.Now()}, false, log.NewNopLogger())

	// Initialize with default params
	err := k.SetParams(ctx, types.DefaultParams())
	require.NoError(t, err)

	return k, ctx, bankKeeper
}

func TestDelegateStake(t *testing.T) {
	k, ctx, bankKeeper := setupKeeperWithBankMock(t)

	delegator := sdk.AccAddress([]byte("delegator123456789"))
	oracleAddr := sdk.AccAddress([]byte("oracle123456789012"))
	amount := sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(1000)))

	// Set delegator balance
	bankKeeper.SetBalance(delegator, amount)

	// Delegate stake
	err := k.DelegateStake(ctx, delegator, oracleAddr, amount)
	require.NoError(t, err)

	// Verify delegator balance is zero
	delegatorBalance := bankKeeper.SpendableCoins(ctx, delegator)
	require.True(t, delegatorBalance.IsZero())

	// Verify oracle stake
	stake := k.GetOracleStake(ctx, oracleAddr)
	require.True(t, stake.Equal(amount))

	// Verify module received the coins
	moduleBalance := bankKeeper.GetModuleBalance(types.ModuleName)
	require.True(t, moduleBalance.Equal(amount))
}

func TestUndelegateStake(t *testing.T) {
	k, ctx, bankKeeper := setupKeeperWithBankMock(t)

	delegator := sdk.AccAddress([]byte("delegator123456789"))
	oracleAddr := sdk.AccAddress([]byte("oracle123456789012"))
	amount := sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(1000)))
	undelegateAmount := sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(500)))

	// Set delegator balance and delegate
	bankKeeper.SetBalance(delegator, amount)
	err := k.DelegateStake(ctx, delegator, oracleAddr, amount)
	require.NoError(t, err)

	// Undelegate partial amount
	err = k.UndelegateStake(ctx, delegator, oracleAddr, undelegateAmount)
	require.NoError(t, err)

	// Verify delegator received coins back
	delegatorBalance := bankKeeper.SpendableCoins(ctx, delegator)
	require.True(t, delegatorBalance.Equal(undelegateAmount))

	// Verify oracle stake is reduced
	stake := k.GetOracleStake(ctx, oracleAddr)
	require.True(t, stake.Equal(undelegateAmount))
}

func TestSlashDeposit(t *testing.T) {
	k, ctx, bankKeeper := setupKeeperWithBankMock(t)

	delegator := sdk.AccAddress([]byte("delegator123456789"))
	oracleAddr := sdk.AccAddress([]byte("oracle123456789012"))
	stakeAmount := sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(1000)))
	slashAmount := sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(100)))

	// Set up oracle stake
	bankKeeper.SetBalance(delegator, stakeAmount)
	err := k.DelegateStake(ctx, delegator, oracleAddr, stakeAmount)
	require.NoError(t, err)

	// Slash deposit
	err = k.SlashDeposit(ctx, oracleAddr, slashAmount, "test misbehavior")
	require.NoError(t, err)

	// Verify stake is reduced
	stake := k.GetOracleStake(ctx, oracleAddr)
	expectedStake := stakeAmount.Sub(slashAmount...)
	require.True(t, stake.Equal(expectedStake))
}

func TestDistributeRewards(t *testing.T) {
	k, ctx, bankKeeper := setupKeeperWithBankMock(t)

	// Get keeper to access internal methods
	kp := k.(*Keeper)

	// Create oracle addresses using raw bytes (like BME tests)
	oracle1 := sdk.AccAddress([]byte("oracle1234567890123"))
	oracle2 := sdk.AccAddress([]byte("oracle2234567890123"))

	// Get the bech32 strings for params
	oracle1Str := oracle1.String()
	oracle2Str := oracle2.String()

	// Set params with oracles
	params := types.DefaultParams()
	params.Sources = []string{oracle1Str, oracle2Str}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	// Add to reward pool
	rewardAmount := sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(1000)))
	bankKeeper.SetModuleBalance(types.ModuleName, rewardAmount)
	kp.setRewardPool(ctx, rewardAmount)

	// Distribute rewards
	err = k.DistributeRewards(ctx, 1)
	require.NoError(t, err)

	// Verify oracles received rewards
	oracle1Balance := bankKeeper.SpendableCoins(ctx, oracle1)
	oracle2Balance := bankKeeper.SpendableCoins(ctx, oracle2)

	// Each oracle should get 500 (1000 / 2 oracles)
	expectedPerOracle := sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(500)))
	require.True(t, oracle1Balance.Equal(expectedPerOracle))
	require.True(t, oracle2Balance.Equal(expectedPerOracle))

	// Verify reward pool is now empty
	pool := kp.GetRewardPool(ctx)
	require.True(t, pool.IsZero())
}

func TestAddToRewardPool(t *testing.T) {
	k, ctx, bankKeeper := setupKeeperWithBankMock(t)

	sender := sdk.AccAddress([]byte("sender123456789012"))
	amount := sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(500)))

	// Set sender balance
	bankKeeper.SetBalance(sender, amount)

	// Get keeper to access internal methods
	kp := k.(*Keeper)

	// Add to reward pool
	err := kp.AddToRewardPool(ctx, sender, amount)
	require.NoError(t, err)

	// Verify pool has the coins
	pool := kp.GetRewardPool(ctx)
	require.True(t, pool.Equal(amount))

	// Verify sender balance is zero
	senderBalance := bankKeeper.SpendableCoins(ctx, sender)
	require.True(t, senderBalance.IsZero())
}

func TestDelegateStakeInsufficientFunds(t *testing.T) {
	k, ctx, bankKeeper := setupKeeperWithBankMock(t)

	delegator := sdk.AccAddress([]byte("delegator123456789"))
	oracleAddr := sdk.AccAddress([]byte("oracle123456789012"))
	amount := sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(1000)))

	// Set delegator balance to less than required
	bankKeeper.SetBalance(delegator, sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(500))))

	// Delegate stake should fail
	err := k.DelegateStake(ctx, delegator, oracleAddr, amount)
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient funds")
}

func TestSlashDepositNoStake(t *testing.T) {
	k, ctx, _ := setupKeeperWithBankMock(t)

	oracleAddr := sdk.AccAddress([]byte("oracle123456789012"))
	slashAmount := sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(100)))

	// Slash deposit without stake should fail
	err := k.SlashDeposit(ctx, oracleAddr, slashAmount, "test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "has no stake")
}

func TestGetDelegatorDelegations(t *testing.T) {
	k, ctx, bankKeeper := setupKeeperWithBankMock(t)

	delegator := sdk.AccAddress([]byte("delegator123456789"))
	oracle1 := sdk.AccAddress([]byte("oracle111111111111"))
	oracle2 := sdk.AccAddress([]byte("oracle222222222222"))
	totalAmount := sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(2000)))
	amountPerOracle := sdk.NewCoins(sdk.NewCoin("uve", math.NewInt(1000)))

	// Set delegator balance
	bankKeeper.SetBalance(delegator, totalAmount)

	// Delegate to multiple oracles
	err := k.DelegateStake(ctx, delegator, oracle1, amountPerOracle)
	require.NoError(t, err)
	err = k.DelegateStake(ctx, delegator, oracle2, amountPerOracle)
	require.NoError(t, err)

	// Get keeper to access internal methods
	kp := k.(*Keeper)

	// Get delegator delegations
	delegations := kp.GetDelegatorDelegations(ctx, delegator)
	require.Len(t, delegations, 2)
}
