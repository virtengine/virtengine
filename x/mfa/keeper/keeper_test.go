// MFA Keeper Tests - Test suite for MFA keeper functionality
// Tests factor enrollment, policies, challenges, sessions, and trusted devices.

package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/mfa/keeper"
	"github.com/virtengine/virtengine/x/mfa/types"
)

// testVEIDKeeper implements keeper.VEIDKeeper for testing
type testVEIDKeeper struct{}

func (m *testVEIDKeeper) GetVEIDScore(_ sdk.Context, _ sdk.AccAddress) (uint32, bool) {
	return 75, true
}

// testRolesKeeper implements keeper.RolesKeeper for testing
type testRolesKeeper struct{}

func (m *testRolesKeeper) IsAccountOperational(_ sdk.Context, _ sdk.AccAddress) bool {
	return true
}

type KeeperTestSuite struct {
	suite.Suite
	ctx    sdk.Context
	keeper keeper.Keeper
	cdc    codec.Codec
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) createContextWithStore(storeKey *storetypes.KVStoreKey) sdk.Context {
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	if err != nil {
		s.T().Fatalf("failed to load latest version: %v", err)
	}

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())
	return ctx
}

func (s *KeeperTestSuite) SetupTest() {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	s.cdc = codec.NewProtoCodec(interfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	s.ctx = s.createContextWithStore(storeKey)
	s.keeper = keeper.NewKeeper(s.cdc, storeKey, "authority", &testVEIDKeeper{}, &testRolesKeeper{})

	// Set default params
	err := s.keeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestFactorEnrollment() {
	addr := sdk.AccAddress([]byte("test_address_123456"))

	// Test initial state - no factors enrolled
	factors := s.keeper.GetFactorEnrollments(s.ctx, addr)
	s.Require().Empty(factors)

	// Enroll a TOTP factor
	enrollment := &types.FactorEnrollment{
		AccountAddress:   addr.String(),
		FactorType:       types.FactorTypeTOTP,
		FactorID:         "totp-001",
		PublicIdentifier: []byte("totp-key"),
		Label:            "My Authenticator",
		Status:           types.EnrollmentStatusActive,
		EnrolledAt:       s.ctx.BlockTime().Unix(),
	}

	err := s.keeper.EnrollFactor(s.ctx, enrollment)
	s.Require().NoError(err)

	// Retrieve and verify
	factors = s.keeper.GetFactorEnrollments(s.ctx, addr)
	s.Require().Len(factors, 1)
	s.Require().Equal(types.FactorTypeTOTP, factors[0].FactorType)
	s.Require().Equal("totp-001", factors[0].FactorID)

	// Get specific factor
	retrieved, found := s.keeper.GetFactorEnrollment(s.ctx, addr, types.FactorTypeTOTP, "totp-001")
	s.Require().True(found)
	s.Require().Equal(enrollment.FactorID, retrieved.FactorID)

	// Revoke factor (replaces delete)
	err = s.keeper.RevokeFactor(s.ctx, addr, types.FactorTypeTOTP, "totp-001")
	s.Require().NoError(err)

	// Verify factor is revoked (not deleted, status changed)
	retrieved, found = s.keeper.GetFactorEnrollment(s.ctx, addr, types.FactorTypeTOTP, "totp-001")
	s.Require().True(found)
	s.Require().Equal(types.EnrollmentStatusRevoked, retrieved.Status)
}

func (s *KeeperTestSuite) TestMFAPolicy() {
	addr := sdk.AccAddress([]byte("test_address_policy_"))

	// Test initial state - no policy
	_, found := s.keeper.GetMFAPolicy(s.ctx, addr)
	s.Require().False(found)

	// Set a policy
	policy := &types.MFAPolicy{
		AccountAddress: addr.String(),
		Enabled:        true,
		RequiredFactors: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP}},
		},
		CreatedAt: s.ctx.BlockTime().Unix(),
		UpdatedAt: s.ctx.BlockTime().Unix(),
	}

	err := s.keeper.SetMFAPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Retrieve and verify
	retrieved, found := s.keeper.GetMFAPolicy(s.ctx, addr)
	s.Require().True(found)
	s.Require().True(retrieved.Enabled)
	s.Require().Len(retrieved.RequiredFactors, 1)

	// Delete policy
	err = s.keeper.DeleteMFAPolicy(s.ctx, addr)
	s.Require().NoError(err)

	// Verify deletion
	_, found = s.keeper.GetMFAPolicy(s.ctx, addr)
	s.Require().False(found)
}

func (s *KeeperTestSuite) TestChallenge() {
	addr := sdk.AccAddress([]byte("test_addr_challenge"))
	challengeID := "challenge-001"

	// Create challenge
	challenge := &types.Challenge{
		ChallengeID:     challengeID,
		AccountAddress:  addr.String(),
		FactorType:      types.FactorTypeFIDO2,
		FactorID:        "fido2-001",
		TransactionType: types.SensitiveTxKeyRotation,
		ChallengeData:   []byte("random-challenge-data"),
		CreatedAt:       s.ctx.BlockTime().Unix(),
		ExpiresAt:       s.ctx.BlockTime().Unix() + 300, // 5 minutes
		Status:          types.ChallengeStatusPending,
		Nonce:           "test-nonce-001",
		MaxAttempts:     3,
	}

	err := s.keeper.CreateChallenge(s.ctx, challenge)
	s.Require().NoError(err)

	// Retrieve
	retrieved, found := s.keeper.GetChallenge(s.ctx, challengeID)
	s.Require().True(found)
	s.Require().Equal(addr.String(), retrieved.AccountAddress)
	s.Require().Equal(types.FactorTypeFIDO2, retrieved.FactorType)

	// Get pending challenges
	pending := s.keeper.GetPendingChallenges(s.ctx, addr)
	s.Require().Len(pending, 1)

	// Delete challenge
	err = s.keeper.DeleteChallenge(s.ctx, challengeID)
	s.Require().NoError(err)

	// Verify deletion
	_, found = s.keeper.GetChallenge(s.ctx, challengeID)
	s.Require().False(found)
}

func (s *KeeperTestSuite) TestAuthorizationSession() {
	addr := sdk.AccAddress([]byte("test_addr_session__"))
	sessionID := "session-001"

	// Create session
	session := &types.AuthorizationSession{
		SessionID:       sessionID,
		AccountAddress:  addr.String(),
		TransactionType: types.SensitiveTxKeyRotation,
		CreatedAt:       s.ctx.BlockTime().Unix(),
		ExpiresAt:       s.ctx.BlockTime().Unix() + 900, // 15 minutes
		VerifiedFactors: []types.FactorType{types.FactorTypeTOTP, types.FactorTypeFIDO2},
		IsSingleUse:     true,
	}

	err := s.keeper.CreateAuthorizationSession(s.ctx, session)
	s.Require().NoError(err)

	// Retrieve
	retrieved, found := s.keeper.GetAuthorizationSession(s.ctx, sessionID)
	s.Require().True(found)
	s.Require().Equal(addr.String(), retrieved.AccountAddress)
	s.Require().Len(retrieved.VerifiedFactors, 2)

	// Delete session
	err = s.keeper.DeleteAuthorizationSession(s.ctx, sessionID)
	s.Require().NoError(err)

	// Verify deletion
	_, found = s.keeper.GetAuthorizationSession(s.ctx, sessionID)
	s.Require().False(found)
}

func (s *KeeperTestSuite) TestTrustedDevice() {
	addr := sdk.AccAddress([]byte("test_addr_device___"))

	// Create trusted device
	device := &types.DeviceInfo{
		Fingerprint:    "device-001",
		UserAgent:      "My Laptop",
		TrustExpiresAt: s.ctx.BlockTime().Unix() + 86400, // 24 hours
	}

	_, err := s.keeper.AddTrustedDevice(s.ctx, addr, device)
	s.Require().NoError(err)

	// Retrieve all devices
	devices := s.keeper.GetTrustedDevices(s.ctx, addr)
	s.Require().Len(devices, 1)

	// Get specific device
	retrieved, found := s.keeper.GetTrustedDevice(s.ctx, addr, "device-001")
	s.Require().True(found)
	s.Require().Equal("My Laptop", retrieved.DeviceInfo.UserAgent)

	// Check trusted device status
	isTrusted := s.keeper.IsTrustedDevice(s.ctx, addr, "device-001")
	s.Require().True(isTrusted)

	// Delete device
	err = s.keeper.RemoveTrustedDevice(s.ctx, addr, "device-001")
	s.Require().NoError(err)

	// Verify deletion
	devices = s.keeper.GetTrustedDevices(s.ctx, addr)
	s.Require().Empty(devices)
}

func (s *KeeperTestSuite) TestSensitiveTxConfig() {
	// Set custom config
	customConfig := &types.SensitiveTxConfig{
		TransactionType: types.SensitiveTxKeyRotation,
		Enabled:         true,
		MinVEIDScore:    50,
		RequiredFactorCombinations: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeFIDO2}},
		},
		SessionDuration: 900, // 15 minutes
		IsSingleUse:     true,
		Description:     "Key rotation requires FIDO2",
	}

	err := s.keeper.SetSensitiveTxConfig(s.ctx, customConfig)
	s.Require().NoError(err)

	// Retrieve
	retrieved, found := s.keeper.GetSensitiveTxConfig(s.ctx, types.SensitiveTxKeyRotation)
	s.Require().True(found)
	s.Require().True(retrieved.Enabled)
	s.Require().Equal(uint32(50), retrieved.MinVEIDScore)
}

func (s *KeeperTestSuite) TestParams() {
	// Get default params
	params := s.keeper.GetParams(s.ctx)
	s.Require().Equal(types.DefaultParams().MaxFactorsPerAccount, params.MaxFactorsPerAccount)

	// Set custom params
	customParams := types.Params{
		DefaultSessionDuration:  1800, // 30 minutes
		MaxFactorsPerAccount:    10,
		MaxChallengeAttempts:    5,
		ChallengeTTL:            600, // 10 minutes
		MaxTrustedDevices:       5,
		TrustedDeviceTTL:        86400, // 24 hours
		MinVEIDScoreForMFA:      50,
		RequireAtLeastOneFactor: true,
		AllowedFactorTypes:      []types.FactorType{types.FactorTypeTOTP, types.FactorTypeFIDO2},
	}

	err := s.keeper.SetParams(s.ctx, customParams)
	s.Require().NoError(err)

	// Retrieve
	retrieved := s.keeper.GetParams(s.ctx)
	s.Require().Equal(uint32(10), retrieved.MaxFactorsPerAccount)
	s.Require().Equal(int64(1800), retrieved.DefaultSessionDuration)
}

func (s *KeeperTestSuite) TestGenesis() {
	addr := sdk.AccAddress([]byte("genesis_test_addr__"))

	// Set up some state
	policy := &types.MFAPolicy{
		AccountAddress: addr.String(),
		Enabled:        true,
		RequiredFactors: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP}},
		},
		CreatedAt: s.ctx.BlockTime().Unix(),
		UpdatedAt: s.ctx.BlockTime().Unix(),
	}
	err := s.keeper.SetMFAPolicy(s.ctx, policy)
	s.Require().NoError(err)

	enrollment := &types.FactorEnrollment{
		AccountAddress:   addr.String(),
		FactorType:       types.FactorTypeTOTP,
		FactorID:         "genesis-totp",
		PublicIdentifier: []byte("totp-key"),
		Status:           types.EnrollmentStatusActive,
		EnrolledAt:       s.ctx.BlockTime().Unix(),
	}
	err = s.keeper.EnrollFactor(s.ctx, enrollment)
	s.Require().NoError(err)

	// Export genesis
	gs := s.keeper.ExportGenesis(s.ctx)
	s.Require().NotNil(gs)
	s.Require().Len(gs.MFAPolicies, 1)
	s.Require().Len(gs.FactorEnrollments, 1)

	// Create new keeper with fresh store
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	newStoreKey := storetypes.NewKVStoreKey(types.StoreKey + "_new")
	newCtx := s.createContextWithStore(newStoreKey)
	newKeeper := keeper.NewKeeper(cdc, newStoreKey, "authority", &testVEIDKeeper{}, &testRolesKeeper{})

	// Init genesis
	newKeeper.InitGenesis(newCtx, gs)

	// Verify state was restored
	restoredPolicy, found := newKeeper.GetMFAPolicy(newCtx, addr)
	s.Require().True(found)
	s.Require().True(restoredPolicy.Enabled)

	enrollments := newKeeper.GetFactorEnrollments(newCtx, addr)
	s.Require().Len(enrollments, 1)
}

// TestIsMFAEnabled tests the MFA enabled check logic
func TestIsMFAEnabled(t *testing.T) {
	// Setup
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	require.NoError(t, err)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())

	k := keeper.NewKeeper(cdc, storeKey, "authority", &testVEIDKeeper{}, &testRolesKeeper{})

	// Set default params
	err = k.SetParams(ctx, types.DefaultParams())
	require.NoError(t, err)

	addr := sdk.AccAddress([]byte("test_mfa_enabled___"))

	// Without policy, MFA is not enabled
	enabled, err := k.IsMFAEnabled(ctx, addr)
	require.NoError(t, err)
	require.False(t, enabled)

	// Set a policy that's disabled
	policy := &types.MFAPolicy{
		AccountAddress: addr.String(),
		Enabled:        false,
		RequiredFactors: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP}},
		},
		CreatedAt: ctx.BlockTime().Unix(),
		UpdatedAt: ctx.BlockTime().Unix(),
	}
	err = k.SetMFAPolicy(ctx, policy)
	require.NoError(t, err)

	// MFA still not enabled because policy is disabled
	enabled, err = k.IsMFAEnabled(ctx, addr)
	require.NoError(t, err)
	require.False(t, enabled)

	// Enable the policy
	policy.Enabled = true
	err = k.SetMFAPolicy(ctx, policy)
	require.NoError(t, err)

	// Still not enabled because no active factors
	enabled, err = k.IsMFAEnabled(ctx, addr)
	require.NoError(t, err)
	require.False(t, enabled)

	// Enroll a factor
	enrollment := &types.FactorEnrollment{
		AccountAddress:   addr.String(),
		FactorType:       types.FactorTypeTOTP,
		FactorID:         "test-totp",
		PublicIdentifier: []byte("totp-key"),
		Status:           types.EnrollmentStatusActive,
		EnrolledAt:       ctx.BlockTime().Unix(),
	}
	err = k.EnrollFactor(ctx, enrollment)
	require.NoError(t, err)

	// Now MFA should be enabled
	enabled, err = k.IsMFAEnabled(ctx, addr)
	require.NoError(t, err)
	require.True(t, enabled)
}

// TestSensitiveTransactionDetection tests detecting sensitive transactions
func TestSensitiveTransactionDetection(t *testing.T) {
	// Test known sensitive transaction types mapping
	testCases := []struct {
		msgTypeURL   string
		expectedType types.SensitiveTransactionType
		isSensitive  bool
	}{
		{"/cosmos.staking.v1beta1.MsgCreateValidator", types.SensitiveTxValidatorRegistration, true},
		{"/cosmos.gov.v1.MsgSubmitProposal", types.SensitiveTxGovernanceProposal, true},
		{"/cosmos.gov.v1beta1.MsgSubmitProposal", types.SensitiveTxGovernanceProposal, true},
		{"/virtengine.provider.v1.MsgCreateProvider", types.SensitiveTxProviderRegistration, true},
		{"/cosmos.bank.v1beta1.MsgSend", types.SensitiveTxUnspecified, false}, // Regular transfer
		{"/virtengine.mfa.v1.MsgDisableMFA", types.SensitiveTxTwoFactorDisable, true},
	}

	for _, tc := range testCases {
		t.Run(tc.msgTypeURL, func(t *testing.T) {
			txType, isSensitive := types.GetSensitiveTransactionType(tc.msgTypeURL)
			require.Equal(t, tc.isSensitive, isSensitive)
			if tc.isSensitive {
				require.Equal(t, tc.expectedType, txType)
			}
		})
	}
}

// TestFactorCombinationValidation tests OR of ANDs logic
func TestFactorCombinationValidation(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	require.NoError(t, err)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())

	k := keeper.NewKeeper(cdc, storeKey, "authority", &testVEIDKeeper{}, &testRolesKeeper{})

	// Set default params
	err = k.SetParams(ctx, types.DefaultParams())
	require.NoError(t, err)

	addr := sdk.AccAddress([]byte("factor_combo_test__"))

	// Set policy: (TOTP AND FIDO2) OR (VEID)
	policy := &types.MFAPolicy{
		AccountAddress: addr.String(),
		Enabled:        true,
		RequiredFactors: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP, types.FactorTypeFIDO2}},
			{Factors: []types.FactorType{types.FactorTypeVEID}},
		},
		CreatedAt: ctx.BlockTime().Unix(),
		UpdatedAt: ctx.BlockTime().Unix(),
	}
	err = k.SetMFAPolicy(ctx, policy)
	require.NoError(t, err)

	// Retrieve policy and test CheckFactorsMatch
	retrieved, found := k.GetMFAPolicy(ctx, addr)
	require.True(t, found)

	// Test case 1: Only TOTP verified - should fail
	match := retrieved.CheckFactorsMatch([]types.FactorType{types.FactorTypeTOTP})
	require.False(t, match.Matched)

	// Test case 2: TOTP and FIDO2 verified - should pass
	match = retrieved.CheckFactorsMatch([]types.FactorType{types.FactorTypeTOTP, types.FactorTypeFIDO2})
	require.True(t, match.Matched)

	// Test case 3: Only VEID verified - should pass (alternative combination)
	match = retrieved.CheckFactorsMatch([]types.FactorType{types.FactorTypeVEID})
	require.True(t, match.Matched)
}

// TestTrustedDeviceReduction tests trusted device factor reduction
func TestTrustedDeviceReduction(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	require.NoError(t, err)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())

	k := keeper.NewKeeper(cdc, storeKey, "authority", &testVEIDKeeper{}, &testRolesKeeper{})

	// Set default params
	err = k.SetParams(ctx, types.DefaultParams())
	require.NoError(t, err)

	addr := sdk.AccAddress([]byte("trusted_device_test"))

	// Set policy with trusted device policy
	policy := &types.MFAPolicy{
		AccountAddress: addr.String(),
		Enabled:        true,
		RequiredFactors: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP, types.FactorTypeFIDO2}},
		},
		TrustedDeviceRule: &types.TrustedDevicePolicy{
			Enabled:                   true,
			TrustDuration:             30 * 24 * 60 * 60, // 30 days
			MaxTrustedDevices:         5,
			RequireReauthForSensitive: true,
		},
		CreatedAt: ctx.BlockTime().Unix(),
		UpdatedAt: ctx.BlockTime().Unix(),
	}
	err = k.SetMFAPolicy(ctx, policy)
	require.NoError(t, err)

	// Add a trusted device
	device := &types.DeviceInfo{
		Fingerprint:    "trusted-001",
		UserAgent:      "My Trusted Laptop",
		TrustExpiresAt: ctx.BlockTime().Unix() + 86400,
	}
	_, err = k.AddTrustedDevice(ctx, addr, device)
	require.NoError(t, err)

	// Check if device is trusted
	isTrusted := k.IsTrustedDevice(ctx, addr, "trusted-001")
	require.True(t, isTrusted)

	// Retrieve policy and test CanUseTrustedDevice
	retrieved, found := k.GetMFAPolicy(ctx, addr)
	require.True(t, found)

	// For low-risk operations, trusted device can reduce factors
	// (when RequireReauthForSensitive is true, critical actions require full MFA)
	canUse := retrieved.CanUseTrustedDevice(types.SensitiveTxLargeWithdrawal, true)
	require.True(t, canUse)

	// For critical operations, trusted device should not bypass MFA
	canUse = retrieved.CanUseTrustedDevice(types.SensitiveTxKeyRotation, true)
	require.False(t, canUse)
}
