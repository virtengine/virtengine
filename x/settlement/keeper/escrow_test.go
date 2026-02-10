package keeper_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	escrowid "github.com/virtengine/virtengine/sdk/go/node/escrow/id/v1"
	escrowmodule "github.com/virtengine/virtengine/sdk/go/node/escrow/module"
	etypes "github.com/virtengine/virtengine/sdk/go/node/escrow/types/v1"
	"github.com/virtengine/virtengine/x/settlement/keeper"
	"github.com/virtengine/virtengine/x/settlement/types"
)

// ptrTime returns a pointer to a time.Time value
func ptrTime(t time.Time) *time.Time {
	return &t
}

// MockBankKeeper is a mock implementation of the bank keeper
type MockBankKeeper struct {
	balances map[string]sdk.Coins
}

func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{
		balances: make(map[string]sdk.Coins),
	}
}

func (m *MockBankKeeper) SendCoins(_ context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	from := fromAddr.String()
	to := toAddr.String()

	fromBalance := m.balances[from]
	if !fromBalance.IsAllGTE(amt) {
		return types.ErrInsufficientFunds
	}

	m.balances[from] = fromBalance.Sub(amt...)
	m.balances[to] = m.balances[to].Add(amt...)
	return nil
}

func (m *MockBankKeeper) SendCoinsFromModuleToModule(_ context.Context, senderModule string, recipientModule string, amt sdk.Coins) error {
	if senderBalance, ok := m.balances[senderModule]; ok {
		if !senderBalance.IsAllGTE(amt) {
			return types.ErrInsufficientFunds
		}
		m.balances[senderModule] = senderBalance.Sub(amt...)
	}
	m.balances[recipientModule] = m.balances[recipientModule].Add(amt...)
	return nil
}

func (m *MockBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	if senderBalance, ok := m.balances[senderModule]; ok {
		if !senderBalance.IsAllGTE(amt) {
			return types.ErrInsufficientFunds
		}
		m.balances[senderModule] = senderBalance.Sub(amt...)
	}
	m.balances[recipientAddr.String()] = m.balances[recipientAddr.String()].Add(amt...)
	return nil
}

func (m *MockBankKeeper) SendCoinsFromAccountToModule(_ context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	sender := senderAddr.String()
	senderBalance := m.balances[sender]
	if !senderBalance.IsAllGTE(amt) {
		return types.ErrInsufficientFunds
	}
	m.balances[sender] = senderBalance.Sub(amt...)
	m.balances[recipientModule] = m.balances[recipientModule].Add(amt...)
	return nil
}

func (m *MockBankKeeper) SpendableCoins(_ context.Context, addr sdk.AccAddress) sdk.Coins {
	return m.balances[addr.String()]
}

func (m *MockBankKeeper) GetBalance(_ context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	balance := m.balances[addr.String()]
	return sdk.NewCoin(denom, balance.AmountOf(denom))
}

func (m *MockBankKeeper) SetBalance(addr sdk.AccAddress, coins sdk.Coins) {
	m.balances[addr.String()] = coins
}

// MockEscrowKeeper is a minimal escrow keeper implementation for tests.
type MockEscrowKeeper struct {
	accounts map[string]etypes.Account
	bank     *MockBankKeeper
}

func NewMockEscrowKeeper(bank *MockBankKeeper) *MockEscrowKeeper {
	return &MockEscrowKeeper{
		accounts: make(map[string]etypes.Account),
		bank:     bank,
	}
}

func (m *MockEscrowKeeper) AccountCreate(ctx sdk.Context, id escrowid.Account, owner sdk.AccAddress, deposits []etypes.Depositor) error {
	key := id.Key()
	if _, exists := m.accounts[key]; exists {
		return types.ErrEscrowExists
	}

	denoms := make(map[string]struct{})
	for _, dep := range deposits {
		denoms[dep.Balance.Denom] = struct{}{}
	}

	funds := make([]etypes.Balance, 0, len(denoms))
	transferred := make([]sdk.DecCoin, 0, len(denoms))
	for denom := range denoms {
		funds = append(funds, etypes.Balance{Denom: denom, Amount: sdkmath.LegacyZeroDec()})
		transferred = append(transferred, sdk.NewDecCoin(denom, sdkmath.ZeroInt()))
	}

	account := etypes.Account{
		ID: id,
		State: etypes.AccountState{
			Owner:       owner.String(),
			State:       etypes.StateOpen,
			Transferred: transferred,
			SettledAt:   ctx.BlockHeight(),
			Funds:       funds,
			Deposits:    make([]etypes.Depositor, 0),
		},
	}

	coins := sdk.NewCoins()
	for _, dep := range deposits {
		account.State.Deposits = append(account.State.Deposits, dep)
		for i := range account.State.Funds {
			if account.State.Funds[i].Denom == dep.Balance.Denom {
				account.State.Funds[i].Amount.AddMut(dep.Balance.Amount)
			}
		}
		coins = coins.Add(sdk.NewCoin(dep.Balance.Denom, dep.Balance.Amount.TruncateInt()))
	}

	if m.bank != nil && !coins.IsZero() {
		if err := m.bank.SendCoinsFromAccountToModule(ctx, owner, escrowmodule.ModuleName, coins); err != nil {
			return err
		}
	}

	m.accounts[key] = account
	return nil
}

func (m *MockEscrowKeeper) AccountDeposit(ctx sdk.Context, id escrowid.Account, deposits []etypes.Depositor) error {
	key := id.Key()
	account, exists := m.accounts[key]
	if !exists {
		return types.ErrEscrowNotFound
	}

	coins := sdk.NewCoins()
	for _, dep := range deposits {
		account.State.Deposits = append(account.State.Deposits, dep)
		for i := range account.State.Funds {
			if account.State.Funds[i].Denom == dep.Balance.Denom {
				account.State.Funds[i].Amount.AddMut(dep.Balance.Amount)
			}
		}
		coins = coins.Add(sdk.NewCoin(dep.Balance.Denom, dep.Balance.Amount.TruncateInt()))
	}

	if m.bank != nil && !coins.IsZero() {
		owner, err := sdk.AccAddressFromBech32(account.State.Owner)
		if err != nil {
			return err
		}
		if err := m.bank.SendCoinsFromAccountToModule(ctx, owner, escrowmodule.ModuleName, coins); err != nil {
			return err
		}
	}

	m.accounts[key] = account
	return nil
}

func (m *MockEscrowKeeper) GetAccount(_ sdk.Context, id escrowid.Account) (etypes.Account, error) {
	account, exists := m.accounts[id.Key()]
	if !exists {
		return etypes.Account{}, errors.New("account not found")
	}
	return account, nil
}

func (m *MockEscrowKeeper) SaveAccount(_ sdk.Context, account etypes.Account) error {
	m.accounts[account.ID.Key()] = account
	return nil
}

// KeeperTestSuite is the test suite for the settlement keeper
type KeeperTestSuite struct {
	suite.Suite

	ctx        sdk.Context
	keeper     keeper.Keeper
	bankKeeper *MockBankKeeper
	escrow     *MockEscrowKeeper
	cdc        codec.Codec
	storeKey   storetypes.StoreKey

	// Test addresses
	depositor sdk.AccAddress
	provider  sdk.AccAddress
	validator sdk.AccAddress
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	s.storeKey = storeKey

	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	s.cdc = codec.NewProtoCodec(interfaceRegistry)

	// Create proper multi-store with backing database
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(s.T(), stateStore.LoadLatestVersion())

	// Create proper context with multi-store
	s.ctx = sdk.NewContext(stateStore, cmtproto.Header{
		Height: 1,
		Time:   time.Now().UTC(),
	}, false, log.NewNopLogger())

	// Create mock bank keeper
	s.bankKeeper = NewMockBankKeeper()
	s.escrow = NewMockEscrowKeeper(s.bankKeeper)

	// Create keeper
	s.keeper = keeper.NewKeeper(s.cdc, storeKey, s.bankKeeper, s.escrow, "authority", mockEncryptionKeeper{})

	// Create test addresses
	s.depositor = sdk.AccAddress([]byte("depositor___________"))
	s.provider = sdk.AccAddress([]byte("provider____________"))
	s.validator = sdk.AccAddress([]byte("validator___________"))

	// Fund depositor
	s.bankKeeper.SetBalance(s.depositor, sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000000))))
}

func (s *KeeperTestSuite) TestCreateEscrow() {
	testCases := []struct {
		name        string
		orderID     string
		depositor   sdk.AccAddress
		amount      sdk.Coins
		duration    time.Duration
		conditions  []types.ReleaseCondition
		expectError bool
	}{
		{
			name:      "valid escrow creation",
			orderID:   "order-1",
			depositor: s.depositor,
			amount:    sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
			duration:  time.Hour * 24,
			conditions: []types.ReleaseCondition{
				{
					Type:        types.ConditionTypeTimelock,
					Satisfied:   false,
					UnlockAfter: ptrTime(time.Now().Add(time.Hour)),
				},
			},
			expectError: false,
		},
		{
			name:        "empty order ID",
			orderID:     "",
			depositor:   s.depositor,
			amount:      sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
			duration:    time.Hour * 24,
			conditions:  []types.ReleaseCondition{},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			escrowID, err := s.keeper.CreateEscrow(s.ctx, tc.orderID, tc.depositor, tc.amount, tc.duration, tc.conditions)

			if tc.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(escrowID)

				// Verify escrow was stored
				escrow, found := s.keeper.GetEscrow(s.ctx, escrowID)
				s.Require().True(found)
				s.Require().Equal(tc.orderID, escrow.OrderID)
				s.Require().Equal(tc.depositor.String(), escrow.Depositor)
				s.Require().Equal(types.EscrowStatePending, escrow.State)
			}
		})
	}
}

func (s *KeeperTestSuite) TestEscrowStateTransitions() {
	// Create an escrow
	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000)))
	escrowID, err := s.keeper.CreateEscrow(s.ctx, "order-state-test", s.depositor, amount, time.Hour*24, nil)
	s.Require().NoError(err)

	// Verify initial state
	escrow, found := s.keeper.GetEscrow(s.ctx, escrowID)
	s.Require().True(found)
	s.Require().Equal(types.EscrowStatePending, escrow.State)

	// Activate escrow
	err = s.keeper.ActivateEscrow(s.ctx, escrowID, "lease-1", s.provider)
	s.Require().NoError(err)

	escrow, found = s.keeper.GetEscrow(s.ctx, escrowID)
	s.Require().True(found)
	s.Require().Equal(types.EscrowStateActive, escrow.State)

	// Release escrow
	err = s.keeper.ReleaseEscrow(s.ctx, escrowID, "service completed")
	s.Require().NoError(err)

	escrow, found = s.keeper.GetEscrow(s.ctx, escrowID)
	s.Require().True(found)
	s.Require().Equal(types.EscrowStateReleased, escrow.State)
}

func (s *KeeperTestSuite) TestInvalidStateTransitions() {
	// Create an escrow
	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000)))
	escrowID, err := s.keeper.CreateEscrow(s.ctx, "order-invalid-transition", s.depositor, amount, time.Hour*24, nil)
	s.Require().NoError(err)

	// Try to release a pending escrow (should fail)
	err = s.keeper.ReleaseEscrow(s.ctx, escrowID, "premature release")
	s.Require().Error(err)

	// Try to refund an active escrow (should fail after activation)
	err = s.keeper.ActivateEscrow(s.ctx, escrowID, "lease-1", s.provider)
	s.Require().NoError(err)

	// Double activation should fail
	err = s.keeper.ActivateEscrow(s.ctx, escrowID, "lease-2", s.provider)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestDispute() {
	// Create and activate an escrow
	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000)))
	escrowID, err := s.keeper.CreateEscrow(s.ctx, "order-dispute", s.depositor, amount, time.Hour*24, nil)
	s.Require().NoError(err)

	err = s.keeper.ActivateEscrow(s.ctx, escrowID, "lease-1", s.provider)
	s.Require().NoError(err)

	// Dispute the escrow
	err = s.keeper.DisputeEscrow(s.ctx, escrowID, "service not as described")
	s.Require().NoError(err)

	escrow, found := s.keeper.GetEscrow(s.ctx, escrowID)
	s.Require().True(found)
	s.Require().Equal(types.EscrowStateDisputed, escrow.State)
}

func TestEscrowValidation(t *testing.T) {
	validAddr := sdk.AccAddress([]byte("test_address________")).String()

	testCases := []struct {
		name        string
		escrow      types.EscrowAccount
		expectError bool
	}{
		{
			name: "valid escrow",
			escrow: types.EscrowAccount{
				EscrowID:  "escrow-1",
				OrderID:   "order-1",
				Depositor: validAddr,
				Amount:    sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
				Balance:   sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
				State:     types.EscrowStatePending,
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(time.Hour * 24),
			},
			expectError: false,
		},
		{
			name: "empty escrow ID",
			escrow: types.EscrowAccount{
				EscrowID:  "",
				OrderID:   "order-1",
				Depositor: validAddr,
				Amount:    sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
				State:     types.EscrowStatePending,
			},
			expectError: true,
		},
		{
			name: "empty order ID",
			escrow: types.EscrowAccount{
				EscrowID:  "escrow-1",
				OrderID:   "",
				Depositor: validAddr,
				Amount:    sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
				State:     types.EscrowStatePending,
			},
			expectError: true,
		},
		{
			name: "zero amount",
			escrow: types.EscrowAccount{
				EscrowID:  "escrow-1",
				OrderID:   "order-1",
				Depositor: validAddr,
				Amount:    sdk.NewCoins(),
				State:     types.EscrowStatePending,
			},
			expectError: true,
		},
		{
			name: "invalid state",
			escrow: types.EscrowAccount{
				EscrowID:  "escrow-1",
				OrderID:   "order-1",
				Depositor: validAddr,
				Amount:    sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
				State:     types.EscrowState("invalid"),
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.escrow.Validate()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
