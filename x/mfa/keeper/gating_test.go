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

type GatingTestSuite struct {
	suite.Suite
	ctx    sdk.Context
	keeper keeper.Keeper
	hooks  keeper.MFAGatingHooks
	cdc    codec.Codec
}

func TestGatingTestSuite(t *testing.T) {
	suite.Run(t, new(GatingTestSuite))
}

func (s *GatingTestSuite) SetupTest() {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	s.cdc = codec.NewProtoCodec(interfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	s.ctx = s.createContextWithStore(storeKey)
	s.keeper = keeper.NewKeeper(s.cdc, storeKey, "authority")
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
	requires := s.hooks.RequiresMFA(s.ctx, address, types.SensitiveTransactionTypeProviderWithdrawal)
	s.Require().False(requires)
}

// Test: RequiresMFA - with policy enabled
func (s *GatingTestSuite) TestRequiresMFA_PolicyEnabled() {
	address := sdk.AccAddress([]byte("test-mfa-enabled"))

	// Enable MFA policy
	policy := &types.MFAPolicy{
		AccountAddress: address.String(),
		Enabled:        true,
		CreatedAt:      time.Now().Unix(),
		UpdatedAt:      time.Now().Unix(),
	}
	err := s.keeper.SetMFAPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Configure sensitive transaction
	txConfig := &types.SensitiveTxConfig{
		TransactionType: types.SensitiveTransactionTypeProviderWithdrawal,
		Enabled:         true,
	}
	err = s.keeper.SetSensitiveTxConfig(s.ctx, txConfig)
	s.Require().NoError(err)

	requires := s.hooks.RequiresMFA(s.ctx, address, types.SensitiveTransactionTypeProviderWithdrawal)
	s.Require().True(requires)
}

// Test: RequiresMFA - tx type not configured
func (s *GatingTestSuite) TestRequiresMFA_TxTypeNotConfigured() {
	address := sdk.AccAddress([]byte("test-mfa-no-tx-cfg"))

	// Enable MFA policy but don't configure tx type
	policy := &types.MFAPolicy{
		AccountAddress: address.String(),
		Enabled:        true,
		CreatedAt:      time.Now().Unix(),
		UpdatedAt:      time.Now().Unix(),
	}
	err := s.keeper.SetMFAPolicy(s.ctx, policy)
	s.Require().NoError(err)

	requires := s.hooks.RequiresMFA(s.ctx, address, types.SensitiveTransactionTypeProviderWithdrawal)
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
		TransactionType: types.SensitiveTransactionTypeProviderWithdrawal,
		CreatedAt:       s.ctx.BlockTime().Unix(),
		ExpiresAt:       s.ctx.BlockTime().Unix() + 3600,
		Verified:        true,
	}
	err := s.keeper.CreateAuthorizationSession(s.ctx, session)
	s.Require().NoError(err)

	proof := &types.MFAProof{
		SessionID: "valid-session-id",
	}

	valid := s.hooks.ValidateMFAProof(s.ctx, address, types.SensitiveTransactionTypeProviderWithdrawal, proof)
	s.Require().True(valid)
}

// Test: ValidateMFAProof - expired session
func (s *GatingTestSuite) TestValidateMFAProof_ExpiredSession() {
	address := sdk.AccAddress([]byte("test-proof-expired"))

	// Create an expired session
	session := &types.AuthorizationSession{
		SessionID:       "expired-session-id",
		AccountAddress:  address.String(),
		TransactionType: types.SensitiveTransactionTypeProviderWithdrawal,
		CreatedAt:       s.ctx.BlockTime().Unix() - 7200,
		ExpiresAt:       s.ctx.BlockTime().Unix() - 3600, // Already expired
		Verified:        true,
	}
	err := s.keeper.CreateAuthorizationSession(s.ctx, session)
	s.Require().NoError(err)

	proof := &types.MFAProof{
		SessionID: "expired-session-id",
	}

	valid := s.hooks.ValidateMFAProof(s.ctx, address, types.SensitiveTransactionTypeProviderWithdrawal, proof)
	s.Require().False(valid)
}

// Test: ValidateMFAProof - wrong account
func (s *GatingTestSuite) TestValidateMFAProof_WrongAccount() {
	address := sdk.AccAddress([]byte("test-proof-wrong"))
	otherAddress := sdk.AccAddress([]byte("other-address"))

	// Create a session for a different address
	session := &types.AuthorizationSession{
		SessionID:       "wrong-account-session",
		AccountAddress:  otherAddress.String(),
		TransactionType: types.SensitiveTransactionTypeProviderWithdrawal,
		CreatedAt:       s.ctx.BlockTime().Unix(),
		ExpiresAt:       s.ctx.BlockTime().Unix() + 3600,
		Verified:        true,
	}
	err := s.keeper.CreateAuthorizationSession(s.ctx, session)
	s.Require().NoError(err)

	proof := &types.MFAProof{
		SessionID: "wrong-account-session",
	}

	valid := s.hooks.ValidateMFAProof(s.ctx, address, types.SensitiveTransactionTypeProviderWithdrawal, proof)
	s.Require().False(valid)
}

// Test: ValidateMFAProof - nil proof
func (s *GatingTestSuite) TestValidateMFAProof_NilProof() {
	address := sdk.AccAddress([]byte("test-proof-nil"))

	valid := s.hooks.ValidateMFAProof(s.ctx, address, types.SensitiveTransactionTypeProviderWithdrawal, nil)
	s.Require().False(valid)
}

// Test: CanBypassMFA - with trusted device
func (s *GatingTestSuite) TestCanBypassMFA_TrustedDevice() {
	address := sdk.AccAddress([]byte("test-bypass-trusted"))

	// Add a trusted device with bypass enabled
	deviceInfo := &types.TrustedDeviceInfo{
		DeviceFingerprint: "trusted-device-fp",
		DeviceName:        "My Trusted Device",
		TrustExpiresAt:    s.ctx.BlockTime().Unix() + 86400, // Expires in 24 hours
		CanBypassMFA:      true,
	}
	err := s.keeper.AddTrustedDevice(s.ctx, address, deviceInfo)
	s.Require().NoError(err)

	// Set policy to allow device bypass
	policy := &types.MFAPolicy{
		AccountAddress:        address.String(),
		Enabled:               true,
		AllowTrustedDevices:   true,
		AllowTrustedDevBypass: true,
	}
	err = s.keeper.SetMFAPolicy(s.ctx, policy)
	s.Require().NoError(err)

	canBypass := s.hooks.CanBypassMFA(s.ctx, address, "trusted-device-fp")
	s.Require().True(canBypass)
}

// Test: CanBypassMFA - expired trusted device
func (s *GatingTestSuite) TestCanBypassMFA_ExpiredDevice() {
	address := sdk.AccAddress([]byte("test-bypass-expired"))

	// Add an expired trusted device
	deviceInfo := &types.TrustedDeviceInfo{
		DeviceFingerprint: "expired-device-fp",
		DeviceName:        "Expired Device",
		TrustExpiresAt:    s.ctx.BlockTime().Unix() - 3600, // Already expired
		CanBypassMFA:      true,
	}
	err := s.keeper.AddTrustedDevice(s.ctx, address, deviceInfo)
	s.Require().NoError(err)

	canBypass := s.hooks.CanBypassMFA(s.ctx, address, "expired-device-fp")
	s.Require().False(canBypass)
}

// Test: CanBypassMFA - device not in list
func (s *GatingTestSuite) TestCanBypassMFA_UnknownDevice() {
	address := sdk.AccAddress([]byte("test-bypass-unknown"))

	canBypass := s.hooks.CanBypassMFA(s.ctx, address, "unknown-device-fp")
	s.Require().False(canBypass)
}

// Test: CanBypassMFA - policy disallows bypass
func (s *GatingTestSuite) TestCanBypassMFA_PolicyDisallows() {
	address := sdk.AccAddress([]byte("test-bypass-disallowed"))

	// Add a trusted device
	deviceInfo := &types.TrustedDeviceInfo{
		DeviceFingerprint: "device-fp",
		DeviceName:        "Device",
		TrustExpiresAt:    s.ctx.BlockTime().Unix() + 86400,
		CanBypassMFA:      true,
	}
	err := s.keeper.AddTrustedDevice(s.ctx, address, deviceInfo)
	s.Require().NoError(err)

	// Set policy that disallows bypass
	policy := &types.MFAPolicy{
		AccountAddress:        address.String(),
		Enabled:               true,
		AllowTrustedDevices:   true,
		AllowTrustedDevBypass: false, // Explicitly disable
	}
	err = s.keeper.SetMFAPolicy(s.ctx, policy)
	s.Require().NoError(err)

	canBypass := s.hooks.CanBypassMFA(s.ctx, address, "device-fp")
	s.Require().False(canBypass)
}

// Test: CheckMFARequired - comprehensive flow
func (s *GatingTestSuite) TestCheckMFARequired_FullFlow() {
	address := sdk.AccAddress([]byte("test-check-full"))
	txType := types.SensitiveTransactionTypeProviderWithdrawal

	// Step 1: No policy - should not require MFA
	err := s.hooks.CheckMFARequired(s.ctx, address, txType, nil, "")
	s.Require().NoError(err)

	// Step 2: Enable policy
	policy := &types.MFAPolicy{
		AccountAddress:      address.String(),
		Enabled:             true,
		AllowTrustedDevices: true,
	}
	err = s.keeper.SetMFAPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Configure tx type as sensitive
	txConfig := &types.SensitiveTxConfig{
		TransactionType: txType,
		Enabled:         true,
	}
	err = s.keeper.SetSensitiveTxConfig(s.ctx, txConfig)
	s.Require().NoError(err)

	// Step 3: Try without proof - should fail
	err = s.hooks.CheckMFARequired(s.ctx, address, txType, nil, "")
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "MFA")

	// Step 4: Create valid session and provide proof
	session := &types.AuthorizationSession{
		SessionID:       "full-flow-session",
		AccountAddress:  address.String(),
		TransactionType: txType,
		CreatedAt:       s.ctx.BlockTime().Unix(),
		ExpiresAt:       s.ctx.BlockTime().Unix() + 3600,
		Verified:        true,
	}
	err = s.keeper.CreateAuthorizationSession(s.ctx, session)
	s.Require().NoError(err)

	proof := &types.MFAProof{
		SessionID: "full-flow-session",
	}
	err = s.hooks.CheckMFARequired(s.ctx, address, txType, proof, "")
	s.Require().NoError(err)
}

// Test: CheckMFARequired - with trusted device bypass
func (s *GatingTestSuite) TestCheckMFARequired_DeviceBypass() {
	address := sdk.AccAddress([]byte("test-check-bypass"))
	txType := types.SensitiveTransactionTypeProviderWithdrawal

	// Enable policy with device bypass
	policy := &types.MFAPolicy{
		AccountAddress:        address.String(),
		Enabled:               true,
		AllowTrustedDevices:   true,
		AllowTrustedDevBypass: true,
	}
	err := s.keeper.SetMFAPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Configure tx type
	txConfig := &types.SensitiveTxConfig{
		TransactionType: txType,
		Enabled:         true,
	}
	err = s.keeper.SetSensitiveTxConfig(s.ctx, txConfig)
	s.Require().NoError(err)

	// Add trusted device
	deviceInfo := &types.TrustedDeviceInfo{
		DeviceFingerprint: "bypass-device",
		DeviceName:        "Trusted Device",
		TrustExpiresAt:    s.ctx.BlockTime().Unix() + 86400,
		CanBypassMFA:      true,
	}
	err = s.keeper.AddTrustedDevice(s.ctx, address, deviceInfo)
	s.Require().NoError(err)

	// Should pass with trusted device fingerprint
	err = s.hooks.CheckMFARequired(s.ctx, address, txType, nil, "bypass-device")
	s.Require().NoError(err)
}

// Test: GetVEIDThreshold
func (s *GatingTestSuite) TestGetVEIDThreshold() {
	address := sdk.AccAddress([]byte("test-veid-threshold"))

	// No policy - should return 0
	threshold := s.hooks.GetVEIDThreshold(s.ctx, address)
	s.Require().Equal(uint32(0), threshold)

	// Set policy with threshold
	policy := &types.MFAPolicy{
		AccountAddress: address.String(),
		Enabled:        true,
		VEIDThreshold:  75,
	}
	err := s.keeper.SetMFAPolicy(s.ctx, policy)
	s.Require().NoError(err)

	threshold = s.hooks.GetVEIDThreshold(s.ctx, address)
	s.Require().Equal(uint32(75), threshold)
}

// Test: ShouldEnforceMFA - based on VEID score
func (s *GatingTestSuite) TestShouldEnforceMFA_VEIDScore() {
	address := sdk.AccAddress([]byte("test-enforce-veid"))

	// Set policy with threshold
	policy := &types.MFAPolicy{
		AccountAddress:   address.String(),
		Enabled:          true,
		VEIDThreshold:    50,
		EnforceForLowVEID: true,
	}
	err := s.keeper.SetMFAPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// High score - may not need MFA (depends on implementation)
	enforce := s.hooks.ShouldEnforceMFA(s.ctx, address, 80)
	s.Require().False(enforce)

	// Low score - should enforce MFA
	enforce = s.hooks.ShouldEnforceMFA(s.ctx, address, 30)
	s.Require().True(enforce)
}

// Test: IsFactorActive
func (s *GatingTestSuite) TestIsFactorActive() {
	address := sdk.AccAddress([]byte("test-factor-active"))

	// No factor enrolled
	active := s.keeper.IsFactorActive(s.ctx, address, types.FactorTypeTOTP)
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

	active = s.keeper.IsFactorActive(s.ctx, address, types.FactorTypeTOTP)
	s.Require().True(active)

	// Revoke factor
	err = s.keeper.RevokeFactor(s.ctx, address, types.FactorTypeTOTP, "active-factor")
	s.Require().NoError(err)

	active = s.keeper.IsFactorActive(s.ctx, address, types.FactorTypeTOTP)
	s.Require().False(active)
}

// Test: GetActiveFactorCount
func (s *GatingTestSuite) TestGetActiveFactorCount() {
	address := sdk.AccAddress([]byte("test-factor-count"))

	// No factors
	count := s.keeper.GetActiveFactorCount(s.ctx, address)
	s.Require().Equal(0, count)

	// Add multiple factors
	factors := []types.FactorType{types.FactorTypeTOTP, types.FactorTypeSMS, types.FactorTypeEmail}
	for i, ft := range factors {
		enrollment := &types.FactorEnrollment{
			AccountAddress:   address.String(),
			FactorType:       ft,
			FactorID:         string(rune('a' + i)),
			PublicIdentifier: []byte("key"),
			Status:           types.EnrollmentStatusActive,
			EnrolledAt:       time.Now().Unix(),
		}
		err := s.keeper.EnrollFactor(s.ctx, enrollment)
		s.Require().NoError(err)
	}

	count = s.keeper.GetActiveFactorCount(s.ctx, address)
	s.Require().Equal(3, count)
}
