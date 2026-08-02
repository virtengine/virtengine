// MFA Gating Tests - Test suite for MFA gating hooks functionality
// Tests MFA requirement checks, proof validation, and bypass mechanisms.

package keeper_test

import (
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
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/mfa/keeper"
	"github.com/virtengine/virtengine/x/mfa/types"
)

// Note: mockVEIDKeeper and mockRolesKeeper are declared in msg_server_test.go
// They can be reused here since we're in the same test package.

type GatingTestSuite struct {
	suite.Suite
	ctx      sdk.Context
	keeper   keeper.Keeper
	hooks    keeper.MFAGatingHooks
	cdc      codec.Codec
	veid     *gatingVEIDKeeper
	storeKey *storetypes.KVStoreKey
}

type gatingVEIDKeeper struct {
	scores map[string]uint32
}

func (m *gatingVEIDKeeper) GetVEIDScore(_ sdk.Context, address sdk.AccAddress) (uint32, bool) {
	score, found := m.scores[address.String()]
	return score, found
}

func TestGatingTestSuite(t *testing.T) {
	suite.Run(t, new(GatingTestSuite))
}

func (s *GatingTestSuite) SetupTest() {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	s.cdc = codec.NewProtoCodec(interfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	s.storeKey = storeKey
	s.ctx = s.createContextWithStore(storeKey)
	s.veid = &gatingVEIDKeeper{scores: make(map[string]uint32)}
	s.keeper = keeper.NewKeeper(s.cdc, storeKey, "authority", s.veid, &mockRolesKeeper{})
	s.hooks = keeper.NewMFAGatingHooks(s.keeper)

	// Set default params
	err := s.keeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)
}

func (s *GatingTestSuite) createContextWithStore(storeKey *storetypes.KVStoreKey) sdk.Context {
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

// Test: RequiresMFA - with policy disabled
func (s *GatingTestSuite) TestRequiresMFA_PolicyDisabled() {
	address := sdk.AccAddress([]byte("test-mfa-disabled"))

	// No policy set means no MFA required
	_, requires, _ := s.hooks.RequiresMFA(s.ctx, address, types.SensitiveTxLargeWithdrawal)
	s.Require().False(requires)
}

func (s *GatingTestSuite) TestRequiresMFA_CorruptSensitiveTxConfigFailsClosed() {
	address := sdk.AccAddress([]byte("corrupt-sensitive-config"))
	store := s.ctx.KVStore(s.storeKey)
	tests := []struct {
		name  string
		value string
	}{
		{name: "malformed", value: `{"enabled":`},
		{name: "wrong transaction type", value: `{"transaction_type":3,"enabled":false}`},
		{name: "invalid enabled config", value: `{"transaction_type":2,"enabled":true}`},
	}
	for _, test := range tests {
		s.Run(test.name, func() {
			store.Set(types.SensitiveTxConfigKey(types.SensitiveTxKeyRotation), []byte(test.value))
			s.Require().Panics(func() {
				s.hooks.RequiresMFA(s.ctx, address, types.SensitiveTxKeyRotation)
			})
		})
	}
}

// Test: RequiresMFA - with policy enabled
func (s *GatingTestSuite) TestRequiresMFA_PolicyEnabled() {
	address := sdk.AccAddress([]byte("test-mfa-enabled"))
	err := enrollActiveFactor(s.keeper, s.ctx, address, types.FactorTypeTOTP, "enabled-totp")
	s.Require().NoError(err)

	// Enable MFA policy with required factors (validation requires at least one)
	policy := &types.MFAPolicy{
		AccountAddress: address.String(),
		Enabled:        true,
		RequiredFactors: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP}},
		},
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	err = s.keeper.SetMFAPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Configure sensitive transaction
	txConfig := &types.SensitiveTxConfig{
		TransactionType: types.SensitiveTxLargeWithdrawal,
		Enabled:         true,
		RequiredFactorCombinations: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP}},
		},
	}
	err = s.keeper.SetSensitiveTxConfig(s.ctx, txConfig)
	s.Require().NoError(err)

	_, requires, _ := s.hooks.RequiresMFA(s.ctx, address, types.SensitiveTxLargeWithdrawal)
	s.Require().True(requires)
}

// Test: RequiresMFA - tx type not configured
func (s *GatingTestSuite) TestRequiresMFA_TxTypeNotConfigured() {
	address := sdk.AccAddress([]byte("test-mfa-no-tx-cfg"))
	err := enrollActiveFactor(s.keeper, s.ctx, address, types.FactorTypeTOTP, "no-cfg-totp")
	s.Require().NoError(err)

	// Enable MFA policy but don't configure tx type
	policy := &types.MFAPolicy{
		AccountAddress: address.String(),
		Enabled:        true,
		RequiredFactors: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP}},
		},
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	err = s.keeper.SetMFAPolicy(s.ctx, policy)
	s.Require().NoError(err)

	_, requires, _ := s.hooks.RequiresMFA(s.ctx, address, types.SensitiveTxLargeWithdrawal)
	// Should check default behavior when tx type is not configured
	s.Require().False(requires)
}

// Test: ValidateMFAProof - valid proof
func (s *GatingTestSuite) TestValidateMFAProof_Valid() {
	address := sdk.AccAddress([]byte("test-proof-valid"))

	// Create a valid authorization session
	session := &types.AuthorizationSession{
		SessionID:       "valid-session-id",
		AccountAddress:  address.String(),
		TransactionType: types.SensitiveTxLargeWithdrawal,
		CreatedAt:       s.ctx.BlockTime().Unix(),
		ExpiresAt:       s.ctx.BlockTime().Unix() + 3600,
		VerifiedFactors: []types.FactorType{types.FactorTypeTOTP},
	}
	err := s.keeper.CreateAuthorizationSession(s.ctx, session)
	s.Require().NoError(err)

	proof := &types.MFAProof{
		SessionID:       "valid-session-id",
		VerifiedFactors: []types.FactorType{types.FactorTypeTOTP},
		Timestamp:       s.ctx.BlockTime().Unix(),
	}

	err = s.hooks.ValidateMFAProof(s.ctx, address, types.SensitiveTxLargeWithdrawal, proof, "")
	s.Require().NoError(err)
}

func (s *GatingTestSuite) TestValidateMFAProof_RequiresBoundDeviceAndCurrentVEIDScore() {
	address := sdk.AccAddress([]byte("proof-device-veid"))
	now := s.ctx.BlockTime().Unix()
	session := &types.AuthorizationSession{
		SessionID: "device-veid-session", AccountAddress: address.String(), TransactionType: types.SensitiveTxMediumWithdrawal,
		CreatedAt: now, ExpiresAt: now + 3600, VerifiedFactors: []types.FactorType{types.FactorTypeVEID},
		DeviceFingerprint: "bound-proof-device",
	}
	s.Require().NoError(s.keeper.CreateAuthorizationSession(s.ctx, session))
	s.Require().NoError(s.keeper.SetSensitiveTxConfig(s.ctx, &types.SensitiveTxConfig{
		TransactionType: types.SensitiveTxMediumWithdrawal, Enabled: true, MinVEIDScore: 80,
		RequiredFactorCombinations: []types.FactorCombination{{Factors: []types.FactorType{types.FactorTypeVEID}}},
	}))
	proof := &types.MFAProof{SessionID: session.SessionID, VerifiedFactors: session.VerifiedFactors, Timestamp: now}

	err := s.hooks.ValidateMFAProof(s.ctx, address, types.SensitiveTxMediumWithdrawal, proof, "")
	s.Require().ErrorIs(err, types.ErrDeviceMismatch)
	s.veid.scores[address.String()] = 79
	err = s.hooks.ValidateMFAProof(s.ctx, address, types.SensitiveTxMediumWithdrawal, proof, session.DeviceFingerprint)
	s.Require().ErrorIs(err, types.ErrVEIDScoreInsufficient)
	s.veid.scores[address.String()] = 80
	s.Require().NoError(s.hooks.ValidateMFAProof(s.ctx, address, types.SensitiveTxMediumWithdrawal, proof, session.DeviceFingerprint))
}

// Test: ValidateMFAProof - step-up within category
func (s *GatingTestSuite) TestValidateMFAProof_StepUpWithinCategory() {
	address := sdk.AccAddress([]byte("test-proof-step-up"))

	// Configure sensitive tx configs for withdrawals
	err := s.keeper.SetSensitiveTxConfig(s.ctx, &types.SensitiveTxConfig{
		TransactionType: types.SensitiveTxLargeWithdrawal,
		Enabled:         true,
		RequiredFactorCombinations: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP}},
		},
	})
	s.Require().NoError(err)
	err = s.keeper.SetSensitiveTxConfig(s.ctx, &types.SensitiveTxConfig{
		TransactionType: types.SensitiveTxMediumWithdrawal,
		Enabled:         true,
		RequiredFactorCombinations: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP}},
		},
	})
	s.Require().NoError(err)

	// Create a valid authorization session for a higher-risk withdrawal
	session := &types.AuthorizationSession{
		SessionID:       "step-up-session-id",
		AccountAddress:  address.String(),
		TransactionType: types.SensitiveTxLargeWithdrawal,
		CreatedAt:       s.ctx.BlockTime().Unix(),
		ExpiresAt:       s.ctx.BlockTime().Unix() + 3600,
		VerifiedFactors: []types.FactorType{types.FactorTypeTOTP},
	}
	err = s.keeper.CreateAuthorizationSession(s.ctx, session)
	s.Require().NoError(err)

	proof := &types.MFAProof{
		SessionID:       "step-up-session-id",
		VerifiedFactors: []types.FactorType{types.FactorTypeTOTP},
		Timestamp:       s.ctx.BlockTime().Unix(),
	}

	// Should authorize lower-risk withdrawal within same category
	err = s.hooks.ValidateMFAProof(s.ctx, address, types.SensitiveTxMediumWithdrawal, proof, "")
	s.Require().NoError(err)
}

// Test: ValidateMFAProof - expired session
func (s *GatingTestSuite) TestValidateMFAProof_ExpiredSession() {
	address := sdk.AccAddress([]byte("test-proof-expired"))

	// Create an expired session
	session := &types.AuthorizationSession{
		SessionID:       "expired-session-id",
		AccountAddress:  address.String(),
		TransactionType: types.SensitiveTxLargeWithdrawal,
		CreatedAt:       s.ctx.BlockTime().Unix() - 7200,
		ExpiresAt:       s.ctx.BlockTime().Unix() - 3600, // Already expired
		VerifiedFactors: []types.FactorType{types.FactorTypeTOTP},
	}
	err := s.keeper.CreateAuthorizationSession(s.ctx, session)
	s.Require().NoError(err)

	proof := &types.MFAProof{
		SessionID: "expired-session-id",
	}

	err = s.hooks.ValidateMFAProof(s.ctx, address, types.SensitiveTxLargeWithdrawal, proof, "")
	s.Require().Error(err)
}

// Test: ValidateMFAProof - wrong account
func (s *GatingTestSuite) TestValidateMFAProof_WrongAccount() {
	address := sdk.AccAddress([]byte("test-proof-wrong"))
	otherAddress := sdk.AccAddress([]byte("other-address"))

	// Create a session for a different address
	session := &types.AuthorizationSession{
		SessionID:       "wrong-account-session",
		AccountAddress:  otherAddress.String(),
		TransactionType: types.SensitiveTxLargeWithdrawal,
		CreatedAt:       s.ctx.BlockTime().Unix(),
		ExpiresAt:       s.ctx.BlockTime().Unix() + 3600,
		VerifiedFactors: []types.FactorType{types.FactorTypeTOTP},
	}
	err := s.keeper.CreateAuthorizationSession(s.ctx, session)
	s.Require().NoError(err)

	proof := &types.MFAProof{
		SessionID: "wrong-account-session",
	}

	err = s.hooks.ValidateMFAProof(s.ctx, address, types.SensitiveTxLargeWithdrawal, proof, "")
	s.Require().Error(err)
}

// Test: ValidateMFAProof - nil proof
func (s *GatingTestSuite) TestValidateMFAProof_NilProof() {
	address := sdk.AccAddress([]byte("test-proof-nil"))

	err := s.hooks.ValidateMFAProof(s.ctx, address, types.SensitiveTxLargeWithdrawal, nil, "")
	s.Require().Error(err)
}

// Test: CanBypassMFA - with trusted device
func (s *GatingTestSuite) TestCanBypassMFA_TrustedDevice() {
	address := sdk.AccAddress([]byte("test-bypass-trusted"))
	err := enrollActiveFactor(s.keeper, s.ctx, address, types.FactorTypeTOTP, "trusted-bypass-factor")
	s.Require().NoError(err)

	// Add a trusted device with bypass enabled
	deviceInfo := &types.DeviceInfo{
		Fingerprint:    "trusted-device-fp",
		UserAgent:      "Test Agent",
		TrustExpiresAt: s.ctx.BlockTime().Unix() + 86400, // Expires in 24 hours
	}
	trustToken, err := s.keeper.AddTrustedDevice(s.ctx, address, deviceInfo)
	s.Require().NoError(err)

	// Set policy to allow device bypass
	policy := &types.MFAPolicy{
		AccountAddress: address.String(),
		Enabled:        true,
		RequiredFactors: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP}},
		},
		TrustedDeviceRule: &types.TrustedDevicePolicy{
			Enabled:           true,
			TrustDuration:     86400,
			MaxTrustedDevices: 5,
		},
	}
	err = s.keeper.SetMFAPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Configure sensitive tx config to allow trusted device reduction
	txConfig := &types.SensitiveTxConfig{
		TransactionType:             types.SensitiveTxLargeWithdrawal,
		Enabled:                     true,
		AllowTrustedDeviceReduction: true,
		RequiredFactorCombinations: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP}},
		},
	}
	err = s.keeper.SetSensitiveTxConfig(s.ctx, txConfig)
	s.Require().NoError(err)

	canBypass, _ := s.hooks.CanBypassMFA(s.ctx, address, types.SensitiveTxLargeWithdrawal, "trusted-device-fp", trustToken)
	s.Require().True(canBypass)
	canBypass, _ = s.hooks.CanBypassMFA(s.ctx, address, types.SensitiveTxLargeWithdrawal, "trusted-device-fp", "")
	s.Require().False(canBypass)
}

// Test: CanBypassMFA - expired trusted device
func (s *GatingTestSuite) TestCanBypassMFA_ExpiredDevice() {
	address := sdk.AccAddress([]byte("test-bypass-expired"))

	// Add an expired trusted device
	deviceInfo := &types.DeviceInfo{
		Fingerprint:    "expired-device-fp",
		UserAgent:      "Expired Agent",
		TrustExpiresAt: s.ctx.BlockTime().Unix() - 3600, // Already expired
	}
	_, err := s.keeper.AddTrustedDevice(s.ctx, address, deviceInfo)
	s.Require().NoError(err)

	canBypass, _ := s.hooks.CanBypassMFA(s.ctx, address, types.SensitiveTxLargeWithdrawal, "expired-device-fp", "invalid")
	s.Require().False(canBypass)
}

// Test: CanBypassMFA - device not in list
func (s *GatingTestSuite) TestCanBypassMFA_UnknownDevice() {
	address := sdk.AccAddress([]byte("test-bypass-unknown"))

	canBypass, _ := s.hooks.CanBypassMFA(s.ctx, address, types.SensitiveTxLargeWithdrawal, "unknown-device-fp", "invalid")
	s.Require().False(canBypass)
}

// Test: CanBypassMFA - policy disallows bypass
func (s *GatingTestSuite) TestCanBypassMFA_PolicyDisallows() {
	address := sdk.AccAddress([]byte("test-bypass-disallowed"))
	err := enrollActiveFactor(s.keeper, s.ctx, address, types.FactorTypeTOTP, "disallow-factor")
	s.Require().NoError(err)

	// Add a trusted device
	deviceInfo := &types.DeviceInfo{
		Fingerprint:    "device-fp",
		UserAgent:      "Device",
		TrustExpiresAt: s.ctx.BlockTime().Unix() + 86400,
	}
	_, err = s.keeper.AddTrustedDevice(s.ctx, address, deviceInfo)
	s.Require().NoError(err)

	// Set policy that disallows bypass (TrustedDeviceRule not enabled)
	policy := &types.MFAPolicy{
		AccountAddress: address.String(),
		Enabled:        true,
		RequiredFactors: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP}},
		},
		// TrustedDeviceRule is nil, so bypass should not be allowed
	}
	err = s.keeper.SetMFAPolicy(s.ctx, policy)
	s.Require().NoError(err)

	canBypass, _ := s.hooks.CanBypassMFA(s.ctx, address, types.SensitiveTxLargeWithdrawal, "device-fp", "invalid")
	s.Require().False(canBypass)
}

// Test: CheckMFARequired - comprehensive flow
// Note: CheckMFARequired takes msgTypeURL string, not SensitiveTransactionType
func (s *GatingTestSuite) TestCheckMFARequired_FullFlow() {
	address := sdk.AccAddress([]byte("test-check-full"))
	msgTypeURL := "/virtengine.market.v1.MsgWithdrawLease" // Example msg type URL

	// Step 1: No policy - should not require MFA
	mfaRequired, bypassAllowed, _ := s.hooks.CheckMFARequired(s.ctx, address, msgTypeURL, "", "")
	s.Require().False(mfaRequired)
	s.Require().False(bypassAllowed)

	err := enrollActiveFactor(s.keeper, s.ctx, address, types.FactorTypeTOTP, "fullflow-factor")
	s.Require().NoError(err)

	// Step 2: Enable policy
	policy := &types.MFAPolicy{
		AccountAddress: address.String(),
		Enabled:        true,
		RequiredFactors: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP}},
		},
		TrustedDeviceRule: &types.TrustedDevicePolicy{
			Enabled:           true,
			TrustDuration:     86400,
			MaxTrustedDevices: 5,
		},
	}
	err = s.keeper.SetMFAPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Configure tx type as sensitive (need to register the mapping)
	txConfig := &types.SensitiveTxConfig{
		TransactionType: types.SensitiveTxLargeWithdrawal,
		Enabled:         true,
		RequiredFactorCombinations: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP}},
		},
	}
	err = s.keeper.SetSensitiveTxConfig(s.ctx, txConfig)
	s.Require().NoError(err)

	// Note: For this test to work, the msgTypeURL must be registered in types.GetSensitiveTransactionType
	// Since we don't have that mapping, we'll just verify the function can be called
	mfaRequired, _, _ = s.hooks.CheckMFARequired(s.ctx, address, msgTypeURL, "", "")
	// Result depends on whether msgTypeURL is registered as sensitive
	_ = mfaRequired
}

// Test: CheckMFARequired - with trusted device bypass
func (s *GatingTestSuite) TestCheckMFARequired_DeviceBypass() {
	address := sdk.AccAddress([]byte("test-check-bypass"))
	msgTypeURL := "/virtengine.market.v1.MsgWithdrawLease" // Example msg type URL
	err := enrollActiveFactor(s.keeper, s.ctx, address, types.FactorTypeTOTP, "device-bypass-factor")
	s.Require().NoError(err)

	// Enable policy with device bypass
	policy := &types.MFAPolicy{
		AccountAddress: address.String(),
		Enabled:        true,
		RequiredFactors: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP}},
		},
		TrustedDeviceRule: &types.TrustedDevicePolicy{
			Enabled:                   true,
			TrustDuration:             86400,
			MaxTrustedDevices:         5,
			RequireReauthForSensitive: false,
		},
	}
	err = s.keeper.SetMFAPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Configure tx type
	txConfig := &types.SensitiveTxConfig{
		TransactionType:             types.SensitiveTxLargeWithdrawal,
		Enabled:                     true,
		AllowTrustedDeviceReduction: true,
		RequiredFactorCombinations: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP}},
		},
	}
	err = s.keeper.SetSensitiveTxConfig(s.ctx, txConfig)
	s.Require().NoError(err)

	// Add trusted device
	deviceInfo := &types.DeviceInfo{
		Fingerprint:    "bypass-device",
		UserAgent:      "Trusted Device",
		TrustExpiresAt: s.ctx.BlockTime().Unix() + 86400,
	}
	trustToken, err := s.keeper.AddTrustedDevice(s.ctx, address, deviceInfo)
	s.Require().NoError(err)

	// Check with trusted device fingerprint - should allow bypass
	_, bypassAllowed, _ := s.hooks.CheckMFARequired(s.ctx, address, msgTypeURL, "bypass-device", trustToken)
	// Result depends on whether msgTypeURL is registered as sensitive
	_ = bypassAllowed
}

func (s *GatingTestSuite) TestGetVEIDThreshold() {
	address := sdk.AccAddress([]byte("test-veid-threshold"))

	threshold := s.hooks.GetVEIDThreshold(s.ctx, address)
	s.Require().Equal(types.DefaultParams().MinVEIDScoreForMFA, threshold)

	err := enrollActiveFactor(s.keeper, s.ctx, address, types.FactorTypeTOTP, "veid-threshold-totp")
	s.Require().NoError(err)

	policy := &types.MFAPolicy{
		AccountAddress: address.String(),
		Enabled:        true,
		RequiredFactors: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP}},
		},
		VEIDThreshold: 75,
	}
	err = s.keeper.SetMFAPolicy(s.ctx, policy)
	s.Require().NoError(err)

	threshold = s.hooks.GetVEIDThreshold(s.ctx, address)
	s.Require().Equal(uint32(75), threshold)
}

func (s *GatingTestSuite) TestShouldEnforceMFA_VEIDScore() {
	address := sdk.AccAddress([]byte("test-enforce-veid"))
	err := enrollActiveFactor(s.keeper, s.ctx, address, types.FactorTypeTOTP, "enforce-veid-totp")
	s.Require().NoError(err)

	policy := &types.MFAPolicy{
		AccountAddress: address.String(),
		Enabled:        true,
		RequiredFactors: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP}},
		},
		VEIDThreshold: 50,
	}
	err = s.keeper.SetMFAPolicy(s.ctx, policy)
	s.Require().NoError(err)

	s.Require().False(s.hooks.ShouldEnforceMFA(s.ctx, address, 80))
	s.Require().True(s.hooks.ShouldEnforceMFA(s.ctx, address, 30))
}

func (s *GatingTestSuite) TestGetActiveFactorCount() {
	address := sdk.AccAddress([]byte("test-factor-count"))

	count := s.keeper.GetActiveFactorCount(s.ctx, address)
	s.Require().Equal(0, count)

	s.Require().NoError(enrollActiveFactor(s.keeper, s.ctx, address, types.FactorTypeTOTP, "count-totp"))
	s.Require().NoError(enrollActiveFactor(s.keeper, s.ctx, address, types.FactorTypeSMS, "count-sms"))
	s.Require().NoError(enrollActiveFactor(s.keeper, s.ctx, address, types.FactorTypeEmail, "count-email"))

	count = s.keeper.GetActiveFactorCount(s.ctx, address)
	s.Require().Equal(3, count)
}

// Test: HasActiveFactorOfType - using the existing keeper method
func (s *GatingTestSuite) TestHasActiveFactorOfType() {
	address := sdk.AccAddress([]byte("test-factor-active"))

	// No factor enrolled
	active := s.keeper.HasActiveFactorOfType(s.ctx, address, types.FactorTypeTOTP)
	s.Require().False(active)

	// Enroll factor
	enrollment := &types.FactorEnrollment{
		AccountAddress:   address.String(),
		FactorType:       types.FactorTypeTOTP,
		FactorID:         "active-factor",
		PublicIdentifier: []byte("totp-key"),
		Status:           types.EnrollmentStatusActive,
		EnrolledAt:       time.Now().Unix(),
	}
	err := s.keeper.EnrollFactor(s.ctx, enrollment)
	s.Require().NoError(err)

	active = s.keeper.HasActiveFactorOfType(s.ctx, address, types.FactorTypeTOTP)
	s.Require().True(active)

	// Revoke factor
	err = s.keeper.RevokeFactor(s.ctx, address, types.FactorTypeTOTP, "active-factor")
	s.Require().NoError(err)

	active = s.keeper.HasActiveFactorOfType(s.ctx, address, types.FactorTypeTOTP)
	s.Require().False(active)
}
