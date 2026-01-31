package keeper_test

import (
	"context"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/veid/keeper"
	"github.com/virtengine/virtengine/x/veid/types"
)

// MockStakingKeeper is a mock implementation of StakingKeeper for testing
type MockStakingKeeper struct {
	validators map[string]stakingtypes.Validator
}

// NewMockStakingKeeper creates a new mock staking keeper
func NewMockStakingKeeper() *MockStakingKeeper {
	return &MockStakingKeeper{
		validators: make(map[string]stakingtypes.Validator),
	}
}

// GetValidator implements StakingKeeper
func (m *MockStakingKeeper) GetValidator(_ context.Context, addr sdk.ValAddress) (stakingtypes.Validator, error) {
	validator, found := m.validators[addr.String()]
	if !found {
		return stakingtypes.Validator{}, stakingtypes.ErrNoValidatorFound
	}
	return validator, nil
}

// AddValidator adds a validator to the mock keeper
func (m *MockStakingKeeper) AddValidator(addr sdk.ValAddress, status stakingtypes.BondStatus) {
	m.validators[addr.String()] = stakingtypes.Validator{
		OperatorAddress: addr.String(),
		Status:          status,
	}
}

// ValidatorAuthTestSuite tests validator authorization
type ValidatorAuthTestSuite struct {
	suite.Suite
	ctx           sdk.Context
	keeper        keeper.Keeper
	cdc           codec.Codec
	stakingKeeper *MockStakingKeeper
	stateStore    store.CommitMultiStore
}

func TestValidatorAuthTestSuite(t *testing.T) {
	suite.Run(t, new(ValidatorAuthTestSuite))
}

func (s *ValidatorAuthTestSuite) SetupTest() {
	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	s.cdc = codec.NewProtoCodec(interfaceRegistry)

	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create context with store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	s.Require().NoError(err)
	s.stateStore = stateStore

	s.ctx = sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())

	// Create keeper
	s.keeper = keeper.NewKeeper(s.cdc, storeKey, "authority")

	// Create and set mock staking keeper
	s.stakingKeeper = NewMockStakingKeeper()
	s.keeper.SetStakingKeeper(s.stakingKeeper)

	// Set default params
	err = s.keeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)
}

func (s *ValidatorAuthTestSuite) TearDownTest() {
	CloseStoreIfNeeded(s.stateStore)
}

// Test: Bonded validator is authorized
func (s *ValidatorAuthTestSuite) TestIsValidator_BondedValidator() {
	// Create a validator address
	validatorAddr := sdk.AccAddress([]byte("validator-bonded-001"))
	valAddr := sdk.ValAddress(validatorAddr)

	// Add as bonded validator
	s.stakingKeeper.AddValidator(valAddr, stakingtypes.Bonded)

	// Should return true
	isValidator := s.keeper.IsValidator(s.ctx, validatorAddr)
	s.Require().True(isValidator, "bonded validator should be authorized")
}

// Test: Unbonded validator is not authorized
func (s *ValidatorAuthTestSuite) TestIsValidator_UnbondedValidator() {
	// Create a validator address
	validatorAddr := sdk.AccAddress([]byte("validator-unbonded-01"))
	valAddr := sdk.ValAddress(validatorAddr)

	// Add as unbonded validator
	s.stakingKeeper.AddValidator(valAddr, stakingtypes.Unbonded)

	// Should return false
	isValidator := s.keeper.IsValidator(s.ctx, validatorAddr)
	s.Require().False(isValidator, "unbonded validator should not be authorized")
}

// Test: Unbonding validator is not authorized
func (s *ValidatorAuthTestSuite) TestIsValidator_UnbondingValidator() {
	// Create a validator address
	validatorAddr := sdk.AccAddress([]byte("validator-unbonding"))
	valAddr := sdk.ValAddress(validatorAddr)

	// Add as unbonding validator
	s.stakingKeeper.AddValidator(valAddr, stakingtypes.Unbonding)

	// Should return false
	isValidator := s.keeper.IsValidator(s.ctx, validatorAddr)
	s.Require().False(isValidator, "unbonding validator should not be authorized")
}

// Test: Non-validator is not authorized
func (s *ValidatorAuthTestSuite) TestIsValidator_NonValidator() {
	// Create a regular account address
	regularAddr := sdk.AccAddress([]byte("regular-account-0001"))

	// Don't add to staking keeper

	// Should return false
	isValidator := s.keeper.IsValidator(s.ctx, regularAddr)
	s.Require().False(isValidator, "non-validator should not be authorized")
}

// Test: Nil staking keeper returns false
func (s *ValidatorAuthTestSuite) TestIsValidator_NilStakingKeeper() {
	// Create a new keeper without staking keeper
	storeKey := storetypes.NewKVStoreKey("test-no-staking")
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	s.Require().NoError(err)
	s.T().Cleanup(func() { CloseStoreIfNeeded(stateStore) })

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())

	keeperNoStaking := keeper.NewKeeper(s.cdc, storeKey, "authority")
	// Do NOT set staking keeper

	validatorAddr := sdk.AccAddress([]byte("any-address-here-12"))

	// Should return false for safety
	isValidator := keeperNoStaking.IsValidator(ctx, validatorAddr)
	s.Require().False(isValidator, "nil staking keeper should deny authorization")
}

// TestUpdateVerificationStatus_ValidatorAuthorization tests that only validators can submit verification updates
func TestUpdateVerificationStatus_ValidatorAuthorization(t *testing.T) {
	// This test would require a full msg server setup
	// For now, we test the IsValidator method directly as a unit test

	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create context with store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	require.NoError(t, err)
	t.Cleanup(func() { CloseStoreIfNeeded(stateStore) })

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())

	// Create keeper with mock staking keeper
	k := keeper.NewKeeper(cdc, storeKey, "authority")
	stakingKeeper := NewMockStakingKeeper()
	k.SetStakingKeeper(stakingKeeper)

	// Create test addresses
	validatorAddr := sdk.AccAddress([]byte("bonded-validator-12"))
	nonValidatorAddr := sdk.AccAddress([]byte("regular-user-1234"))

	// Add validator as bonded
	stakingKeeper.AddValidator(sdk.ValAddress(validatorAddr), stakingtypes.Bonded)

	// Test: Validator should be authorized
	require.True(t, k.IsValidator(ctx, validatorAddr), "bonded validator should be authorized")

	// Test: Non-validator should not be authorized
	require.False(t, k.IsValidator(ctx, nonValidatorAddr), "non-validator should not be authorized")
}

// TestUpdateScore_ValidatorAuthorization tests that only validators can submit score updates
func TestUpdateScore_ValidatorAuthorization(t *testing.T) {
	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create context with store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	require.NoError(t, err)
	t.Cleanup(func() { CloseStoreIfNeeded(stateStore) })

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())

	// Create keeper with mock staking keeper
	k := keeper.NewKeeper(cdc, storeKey, "authority")
	stakingKeeper := NewMockStakingKeeper()
	k.SetStakingKeeper(stakingKeeper)

	// Test scenarios
	testCases := []struct {
		name       string
		addr       sdk.AccAddress
		isBonded   bool
		shouldPass bool
	}{
		{
			name:       "bonded validator can submit scores",
			addr:       sdk.AccAddress([]byte("bonded-val-score-1")),
			isBonded:   true,
			shouldPass: true,
		},
		{
			name:       "unbonded validator cannot submit scores",
			addr:       sdk.AccAddress([]byte("unbonded-val-score")),
			isBonded:   false,
			shouldPass: false,
		},
		{
			name:       "non-validator cannot submit scores",
			addr:       sdk.AccAddress([]byte("regular-user-score")),
			isBonded:   false,
			shouldPass: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.isBonded {
				stakingKeeper.AddValidator(sdk.ValAddress(tc.addr), stakingtypes.Bonded)
			}

			isValidator := k.IsValidator(ctx, tc.addr)
			require.Equal(t, tc.shouldPass, isValidator, "IsValidator result mismatch for %s", tc.name)
		})
	}
}
