package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/mfa/types"
)

type KeeperTestSuite struct {
	suite.Suite
	ctx    sdk.Context
	keeper Keeper
	cdc    codec.BinaryCodec
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	// Create store key
	key := storetypes.NewKVStoreKey(types.StoreKey)

	// Create context with store
	testCtx := sdktestutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	s.ctx = testCtx.Ctx.WithBlockTime(time.Now())

	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	s.cdc = codec.NewProtoCodec(interfaceRegistry)

	// Create keeper
	storeService := runtime.NewKVStoreService(key)
	s.keeper = NewKeeper(s.cdc, storeService, log.NewNopLogger(), nil)
}

func (s *KeeperTestSuite) TestFactorEnrollment() {
	addr := sdk.AccAddress([]byte("test_address_123456"))
	addrStr := addr.String()

	// Test initial state - no factors enrolled
	factors, err := s.keeper.GetFactorEnrollments(s.ctx, addrStr)
	s.Require().NoError(err)
	s.Require().Empty(factors)

	// Enroll a TOTP factor
	enrollment := types.FactorEnrollment{
		Address:      addrStr,
		FactorType:   types.FactorTypeTOTP,
		FactorID:     "totp-001",
		Status:       types.EnrollmentStatusActive,
		CreatedAt:    s.ctx.BlockTime(),
		LastUsedAt:   time.Time{},
		FactorInfo:   types.FactorInfo{Label: "My Authenticator"},
		Fingerprint:  types.ComputeFactorFingerprint(types.FactorTypeTOTP, []byte("totp-001")),
		SecurityLevel: types.FactorSecurityLevelMedium,
	}

	err = s.keeper.SetFactorEnrollment(s.ctx, enrollment)
	s.Require().NoError(err)

	// Retrieve and verify
	factors, err = s.keeper.GetFactorEnrollments(s.ctx, addrStr)
	s.Require().NoError(err)
	s.Require().Len(factors, 1)
	s.Require().Equal(types.FactorTypeTOTP, factors[0].FactorType)
	s.Require().Equal("totp-001", factors[0].FactorID)

	// Get specific factor
	retrieved, found := s.keeper.GetFactorEnrollment(s.ctx, addrStr, types.FactorTypeTOTP, "totp-001")
	s.Require().True(found)
	s.Require().Equal(enrollment.FactorID, retrieved.FactorID)

	// Delete factor
	err = s.keeper.DeleteFactorEnrollment(s.ctx, addrStr, types.FactorTypeTOTP, "totp-001")
	s.Require().NoError(err)

	// Verify deletion
	factors, err = s.keeper.GetFactorEnrollments(s.ctx, addrStr)
	s.Require().NoError(err)
	s.Require().Empty(factors)
}

func (s *KeeperTestSuite) TestMFAPolicy() {
	addr := sdk.AccAddress([]byte("test_address_policy_"))
	addrStr := addr.String()

	// Test initial state - no policy
	_, found := s.keeper.GetMFAPolicy(s.ctx, addrStr)
	s.Require().False(found)

	// Set a policy
	policy := types.MFAPolicy{
		Address:   addrStr,
		IsEnabled: true,
		RequiredFactors: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP}},
		},
		CreatedAt: s.ctx.BlockTime(),
		UpdatedAt: s.ctx.BlockTime(),
	}

	err := s.keeper.SetMFAPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Retrieve and verify
	retrieved, found := s.keeper.GetMFAPolicy(s.ctx, addrStr)
	s.Require().True(found)
	s.Require().True(retrieved.IsEnabled)
	s.Require().Len(retrieved.RequiredFactors, 1)

	// Delete policy
	err = s.keeper.DeleteMFAPolicy(s.ctx, addrStr)
	s.Require().NoError(err)

	// Verify deletion
	_, found = s.keeper.GetMFAPolicy(s.ctx, addrStr)
	s.Require().False(found)
}

func (s *KeeperTestSuite) TestChallenge() {
	addr := sdk.AccAddress([]byte("test_addr_challenge"))
	addrStr := addr.String()
	challengeID := "challenge-001"

	// Create challenge
	challenge := types.Challenge{
		ChallengeID:  challengeID,
		Address:      addrStr,
		FactorType:   types.FactorTypeFIDO2,
		ChallengeData: []byte("random-challenge-data"),
		CreatedAt:    s.ctx.BlockTime(),
		ExpiresAt:    s.ctx.BlockTime().Add(5 * time.Minute),
		Status:       types.ChallengeStatusPending,
	}

	err := s.keeper.SetChallenge(s.ctx, challenge)
	s.Require().NoError(err)

	// Retrieve
	retrieved, found := s.keeper.GetChallenge(s.ctx, challengeID)
	s.Require().True(found)
	s.Require().Equal(addrStr, retrieved.Address)
	s.Require().Equal(types.FactorTypeFIDO2, retrieved.FactorType)

	// Get pending challenges
	pending, err := s.keeper.GetPendingChallenges(s.ctx, addrStr)
	s.Require().NoError(err)
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
	addrStr := addr.String()
	sessionID := "session-001"

	// Create session
	session := types.AuthorizationSession{
		SessionID:    sessionID,
		Address:      addrStr,
		CreatedAt:    s.ctx.BlockTime(),
		ExpiresAt:    s.ctx.BlockTime().Add(15 * time.Minute),
		VerifiedFactors: []types.FactorType{types.FactorTypeTOTP, types.FactorTypeFIDO2},
		SecurityLevel: types.FactorSecurityLevelHigh,
		AllowedOperations: []string{
			types.SensitiveTxKeyRotation.String(),
		},
	}

	err := s.keeper.SetAuthorizationSession(s.ctx, session)
	s.Require().NoError(err)

	// Retrieve
	retrieved, found := s.keeper.GetAuthorizationSession(s.ctx, sessionID)
	s.Require().True(found)
	s.Require().Equal(addrStr, retrieved.Address)
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
	addrStr := addr.String()
	deviceID := "device-001"

	// Create trusted device
	device := types.TrustedDevice{
		DeviceID:     deviceID,
		Address:      addrStr,
		DeviceName:   "My Laptop",
		DeviceType:   "laptop",
		PublicKey:    []byte("device-public-key"),
		Fingerprint:  types.ComputeDeviceFingerprint("My Laptop", []byte("device-public-key")),
		RegisteredAt: s.ctx.BlockTime(),
		LastUsedAt:   s.ctx.BlockTime(),
		IsActive:     true,
		AllowedFactorReduction: true,
	}

	err := s.keeper.SetTrustedDevice(s.ctx, device)
	s.Require().NoError(err)

	// Retrieve all devices
	devices, err := s.keeper.GetTrustedDevices(s.ctx, addrStr)
	s.Require().NoError(err)
	s.Require().Len(devices, 1)

	// Get specific device
	retrieved, found := s.keeper.GetTrustedDevice(s.ctx, addrStr, deviceID)
	s.Require().True(found)
	s.Require().Equal("My Laptop", retrieved.DeviceName)

	// Delete device
	err = s.keeper.DeleteTrustedDevice(s.ctx, addrStr, deviceID)
	s.Require().NoError(err)

	// Verify deletion
	devices, err = s.keeper.GetTrustedDevices(s.ctx, addrStr)
	s.Require().NoError(err)
	s.Require().Empty(devices)
}

func (s *KeeperTestSuite) TestSensitiveTxConfig() {
	// Get default config
	config := s.keeper.GetSensitiveTxConfig(s.ctx)
	s.Require().NotNil(config.Configs)

	// Set custom config
	customConfig := types.SensitiveTxConfig{
		Configs: map[string]types.SensitiveTxSettings{
			types.SensitiveTxKeyRotation.String(): {
				RequiresMFA:       true,
				MinSecurityLevel:  types.FactorSecurityLevelHigh,
				MinFactorCount:    2,
				CooldownPeriod:    24 * time.Hour,
			},
		},
		GlobalMinSecurityLevel: types.FactorSecurityLevelMedium,
		IsEnabled:              true,
	}

	err := s.keeper.SetSensitiveTxConfig(s.ctx, customConfig)
	s.Require().NoError(err)

	// Retrieve
	retrieved := s.keeper.GetSensitiveTxConfig(s.ctx)
	s.Require().True(retrieved.IsEnabled)
	s.Require().Equal(types.FactorSecurityLevelMedium, retrieved.GlobalMinSecurityLevel)
}

func (s *KeeperTestSuite) TestParams() {
	// Get default params
	params := s.keeper.GetParams(s.ctx)
	s.Require().Equal(types.DefaultParams().MaxFactorsPerAccount, params.MaxFactorsPerAccount)

	// Set custom params
	customParams := types.Params{
		MaxFactorsPerAccount:     10,
		ChallengeExpiryDuration:  10 * time.Minute,
		SessionExpiryDuration:    30 * time.Minute,
		MaxTrustedDevices:        5,
		RequireFactorVerification: true,
		MinFactorsForRecovery:    2,
	}

	err := s.keeper.SetParams(s.ctx, customParams)
	s.Require().NoError(err)

	// Retrieve
	retrieved := s.keeper.GetParams(s.ctx)
	s.Require().Equal(10, retrieved.MaxFactorsPerAccount)
	s.Require().Equal(10*time.Minute, retrieved.ChallengeExpiryDuration)
}

func (s *KeeperTestSuite) TestGenesis() {
	addr := sdk.AccAddress([]byte("genesis_test_addr__"))
	addrStr := addr.String()

	// Set up some state
	policy := types.MFAPolicy{
		Address:   addrStr,
		IsEnabled: true,
		RequiredFactors: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP}},
		},
		CreatedAt: s.ctx.BlockTime(),
		UpdatedAt: s.ctx.BlockTime(),
	}
	err := s.keeper.SetMFAPolicy(s.ctx, policy)
	s.Require().NoError(err)

	enrollment := types.FactorEnrollment{
		Address:      addrStr,
		FactorType:   types.FactorTypeTOTP,
		FactorID:     "genesis-totp",
		Status:       types.EnrollmentStatusActive,
		CreatedAt:    s.ctx.BlockTime(),
	}
	err = s.keeper.SetFactorEnrollment(s.ctx, enrollment)
	s.Require().NoError(err)

	// Export genesis
	gs := s.keeper.ExportGenesis(s.ctx)
	s.Require().NotNil(gs)
	s.Require().Len(gs.Policies, 1)
	s.Require().Len(gs.Enrollments, 1)

	// Clear state by creating new keeper
	key := storetypes.NewKVStoreKey(types.StoreKey + "_new")
	testCtx := sdktestutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test_new"))
	newCtx := testCtx.Ctx.WithBlockTime(time.Now())
	storeService := runtime.NewKVStoreService(key)
	newKeeper := NewKeeper(s.cdc, storeService, log.NewNopLogger(), nil)

	// Init genesis
	newKeeper.InitGenesis(newCtx, gs)

	// Verify state was restored
	restoredPolicy, found := newKeeper.GetMFAPolicy(newCtx, addrStr)
	s.Require().True(found)
	s.Require().True(restoredPolicy.IsEnabled)

	enrollments, err := newKeeper.GetFactorEnrollments(newCtx, addrStr)
	s.Require().NoError(err)
	s.Require().Len(enrollments, 1)
}

// TestIsMFARequired tests the MFA requirement logic
func TestIsMFARequired(t *testing.T) {
	// Setup
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := sdktestutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockTime(time.Now())

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeService := runtime.NewKVStoreService(key)
	k := NewKeeper(cdc, storeService, log.NewNopLogger(), nil)

	addr := sdk.AccAddress([]byte("test_mfa_required__"))
	addrStr := addr.String()

	// Without policy, MFA is not required
	hooks := NewMFAGatingHooks(k)
	required, err := hooks.RequiresMFA(ctx, addrStr, types.SensitiveTxKeyRotation)
	require.NoError(t, err)
	require.False(t, required)

	// Set a policy
	policy := types.MFAPolicy{
		Address:   addrStr,
		IsEnabled: true,
		RequiredFactors: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP}},
		},
		CreatedAt: ctx.BlockTime(),
		UpdatedAt: ctx.BlockTime(),
	}
	err = k.SetMFAPolicy(ctx, policy)
	require.NoError(t, err)

	// Now MFA should be required for sensitive transactions
	required, err = hooks.RequiresMFA(ctx, addrStr, types.SensitiveTxKeyRotation)
	require.NoError(t, err)
	require.True(t, required)
}

// TestSensitiveTransactionDetection tests detecting sensitive transactions
func TestSensitiveTransactionDetection(t *testing.T) {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := sdktestutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockTime(time.Now())

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeService := runtime.NewKVStoreService(key)
	k := NewKeeper(cdc, storeService, log.NewNopLogger(), nil)

	hooks := NewMFAGatingHooks(k)

	// Test known sensitive transaction types
	testCases := []struct {
		msgTypeURL   string
		expectedType types.SensitiveTransactionType
		isSensitive  bool
	}{
		{"/cosmos.staking.v1beta1.MsgCreateValidator", types.SensitiveTxValidatorRegistration, true},
		{"/cosmos.staking.v1beta1.MsgEditValidator", types.SensitiveTxValidatorUpdate, true},
		{"/cosmos.gov.v1beta1.MsgSubmitProposal", types.SensitiveTxGovernanceProposal, true},
		{"/virtengine.provider.v1.MsgRegisterProvider", types.SensitiveTxProviderRegistration, true},
		{"/cosmos.bank.v1beta1.MsgSend", types.SensitiveTxNone, false}, // Regular transfer
	}

	for _, tc := range testCases {
		t.Run(tc.msgTypeURL, func(t *testing.T) {
			isSensitive, txType := hooks.IsSensitiveTransaction(ctx, tc.msgTypeURL)
			require.Equal(t, tc.isSensitive, isSensitive)
			if tc.isSensitive {
				require.Equal(t, tc.expectedType, txType)
			}
		})
	}
}

// TestFactorCombinationValidation tests OR of ANDs logic
func TestFactorCombinationValidation(t *testing.T) {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := sdktestutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockTime(time.Now())

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeService := runtime.NewKVStoreService(key)
	k := NewKeeper(cdc, storeService, log.NewNopLogger(), nil)

	addr := sdk.AccAddress([]byte("factor_combo_test__"))
	addrStr := addr.String()

	// Set policy: (TOTP AND FIDO2) OR (VEID)
	policy := types.MFAPolicy{
		Address:   addrStr,
		IsEnabled: true,
		RequiredFactors: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP, types.FactorTypeFIDO2}},
			{Factors: []types.FactorType{types.FactorTypeVEID}},
		},
		CreatedAt: ctx.BlockTime(),
		UpdatedAt: ctx.BlockTime(),
	}
	err := k.SetMFAPolicy(ctx, policy)
	require.NoError(t, err)

	hooks := NewMFAGatingHooks(k)

	// Test case 1: Only TOTP verified - should fail
	proof := types.MFAProof{
		VerifiedFactors: []types.FactorType{types.FactorTypeTOTP},
		SessionID:       "test-session",
	}
	valid := hooks.ValidateMFAProof(ctx, addrStr, proof)
	require.False(t, valid)

	// Test case 2: TOTP and FIDO2 verified - should pass
	proof = types.MFAProof{
		VerifiedFactors: []types.FactorType{types.FactorTypeTOTP, types.FactorTypeFIDO2},
		SessionID:       "test-session",
	}
	valid = hooks.ValidateMFAProof(ctx, addrStr, proof)
	require.True(t, valid)

	// Test case 3: Only VEID verified - should pass (alternative combination)
	proof = types.MFAProof{
		VerifiedFactors: []types.FactorType{types.FactorTypeVEID},
		SessionID:       "test-session",
	}
	valid = hooks.ValidateMFAProof(ctx, addrStr, proof)
	require.True(t, valid)
}

// TestTrustedDeviceReduction tests trusted device factor reduction
func TestTrustedDeviceReduction(t *testing.T) {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := sdktestutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockTime(time.Now())

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeService := runtime.NewKVStoreService(key)
	k := NewKeeper(cdc, storeService, log.NewNopLogger(), nil)

	addr := sdk.AccAddress([]byte("trusted_device_test"))
	addrStr := addr.String()

	// Set policy with trusted device policy
	policy := types.MFAPolicy{
		Address:   addrStr,
		IsEnabled: true,
		RequiredFactors: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP, types.FactorTypeFIDO2}},
		},
		TrustedDevicePolicy: types.TrustedDevicePolicy{
			AllowFactorReduction: true,
			ReduceToFactors:      1,
			TrustDuration:        30 * 24 * time.Hour,
		},
		CreatedAt: ctx.BlockTime(),
		UpdatedAt: ctx.BlockTime(),
	}
	err := k.SetMFAPolicy(ctx, policy)
	require.NoError(t, err)

	// Add a trusted device
	device := types.TrustedDevice{
		DeviceID:     "trusted-001",
		Address:      addrStr,
		DeviceName:   "My Trusted Laptop",
		RegisteredAt: ctx.BlockTime(),
		LastUsedAt:   ctx.BlockTime(),
		IsActive:     true,
		AllowedFactorReduction: true,
	}
	err = k.SetTrustedDevice(ctx, device)
	require.NoError(t, err)

	hooks := NewMFAGatingHooks(k)

	// With trusted device, can bypass with reduced factors
	canBypass := hooks.CanBypassMFA(ctx, addrStr, "trusted-001", types.SensitiveTxNone)
	require.True(t, canBypass)

	// For high-security operations, trusted device should not allow bypass
	canBypass = hooks.CanBypassMFA(ctx, addrStr, "trusted-001", types.SensitiveTxKeyRotation)
	require.False(t, canBypass)
}
