package keeper_test

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/json"
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

// mockVEIDKeeper implements keeper.VEIDKeeper for testing
type mockVEIDKeeper struct{}

func (m *mockVEIDKeeper) GetVEIDScore(_ sdk.Context, _ sdk.AccAddress) (uint32, bool) {
	return 0, false
}

// mockRolesKeeper implements keeper.RolesKeeper for testing
type mockRolesKeeper struct{}

func (m *mockRolesKeeper) IsAccountOperational(_ sdk.Context, _ sdk.AccAddress) bool {
	return true
}

type MsgServerTestSuite struct {
	suite.Suite
	ctx       sdk.Context
	keeper    keeper.Keeper
	msgServer types.MsgServer
	cdc       codec.Codec
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

func (s *MsgServerTestSuite) SetupTest() {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	s.cdc = codec.NewProtoCodec(interfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	s.ctx = s.createContextWithStore(storeKey)
	s.keeper = keeper.NewKeeper(s.cdc, storeKey, "authority", &mockVEIDKeeper{}, &mockRolesKeeper{})
	s.msgServer = keeper.NewMsgServerWithContext(s.keeper)

	// Set default params
	err := s.keeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)
}

func (s *MsgServerTestSuite) createContextWithStore(storeKey *storetypes.KVStoreKey) sdk.Context {
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	if err != nil {
		s.T().Fatalf("failed to load latest version: %v", err)
	}

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		ChainID: "virtengine-test-1",
		Time:    time.Date(2026, time.August, 1, 12, 0, 0, 0, time.UTC),
		Height:  100,
	}, false, log.NewNopLogger())
	return ctx
}

// Test: EnrollFactor - success
func (s *MsgServerTestSuite) TestEnrollFactor_Success() {
	address := sdk.AccAddress([]byte("test-enroll-addr"))

	msg := &types.MsgEnrollFactor{
		Sender:           address.String(),
		FactorType:       types.FactorTypeTOTP,
		PublicIdentifier: []byte("totp-secret-key"),
		Label:            "My Authenticator",
		Metadata: &types.FactorMetadata{
			DeviceInfo: &types.DeviceInfo{
				Fingerprint: "iphone15-fp",
				UserAgent:   "iPhone 15",
			},
		},
	}

	resp, err := s.msgServer.EnrollFactor(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotEmpty(resp.FactorID)
	s.Require().Equal(types.EnrollmentStatusActive, resp.Status)

	// Verify enrollment was stored
	enrollments := s.keeper.GetFactorEnrollments(s.ctx, address)
	s.Require().Len(enrollments, 1)
	s.Require().Equal(types.FactorTypeTOTP, enrollments[0].FactorType)
}

// Test: EnrollFactor - invalid address
func (s *MsgServerTestSuite) TestEnrollFactor_InvalidAddress() {
	msg := &types.MsgEnrollFactor{
		Sender:     "invalid-address",
		FactorType: types.FactorTypeTOTP,
	}

	_, err := s.msgServer.EnrollFactor(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "invalid")
}

// Test: EnrollFactor - invalid factor type
func (s *MsgServerTestSuite) TestEnrollFactor_InvalidFactorType() {
	address := sdk.AccAddress([]byte("test-invalid-factor"))

	// Set params to disallow certain factor types
	params := types.DefaultParams()
	params.AllowedFactorTypes = []types.FactorType{types.FactorTypeTOTP}
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	msg := &types.MsgEnrollFactor{
		Sender:     address.String(),
		FactorType: types.FactorTypeSMS, // Not in allowed list
	}

	_, err = s.msgServer.EnrollFactor(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "not allowed")
}

// Test: RevokeFactor - success (no MFA required)
func (s *MsgServerTestSuite) TestRevokeFactor_Success() {
	address := sdk.AccAddress([]byte("test-revoke-addr"))

	// Enroll a factor first
	enrollment := &types.FactorEnrollment{
		AccountAddress:   address.String(),
		FactorType:       types.FactorTypeTOTP,
		FactorID:         "factor-to-revoke",
		PublicIdentifier: []byte("totp-key"),
		Label:            "Test TOTP",
		Status:           types.EnrollmentStatusActive,
		EnrolledAt:       time.Now().Unix(),
	}
	err := s.keeper.EnrollFactor(s.ctx, enrollment)
	s.Require().NoError(err)

	msg := &types.MsgRevokeFactor{
		Sender:     address.String(),
		FactorType: types.FactorTypeTOTP,
		FactorID:   "factor-to-revoke",
	}

	resp, err := s.msgServer.RevokeFactor(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().True(resp.Success)

	// Verify enrollment was revoked
	enrollments := s.keeper.GetFactorEnrollments(s.ctx, address)
	s.Require().Len(enrollments, 1)
	s.Require().Equal(types.EnrollmentStatusRevoked, enrollments[0].Status)
}

// Test: RevokeFactor - requires MFA when policy enabled
func (s *MsgServerTestSuite) TestRevokeFactor_RequiresMFA() {
	address := sdk.AccAddress([]byte("test-revoke-mfa"))

	// Enroll a factor
	enrollment := &types.FactorEnrollment{
		AccountAddress:   address.String(),
		FactorType:       types.FactorTypeTOTP,
		FactorID:         "factor-mfa-req",
		PublicIdentifier: []byte("totp-key"),
		Status:           types.EnrollmentStatusActive,
		EnrolledAt:       time.Now().Unix(),
	}
	err := s.keeper.EnrollFactor(s.ctx, enrollment)
	s.Require().NoError(err)

	// Enable MFA policy with required factors
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

	// Try to revoke without MFA proof
	msg := &types.MsgRevokeFactor{
		Sender:     address.String(),
		FactorType: types.FactorTypeTOTP,
		FactorID:   "factor-mfa-req",
		MFAProof:   nil,
	}

	_, err = s.msgServer.RevokeFactor(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "MFA proof required")
}

// Test: SetMFAPolicy - success
func (s *MsgServerTestSuite) TestSetMFAPolicy_Success() {
	address := sdk.AccAddress([]byte("test-policy-addr"))

	// First enroll a factor (required by params)
	params := types.DefaultParams()
	params.RequireAtLeastOneFactor = true
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	enrollment := &types.FactorEnrollment{
		AccountAddress:   address.String(),
		FactorType:       types.FactorTypeTOTP,
		FactorID:         "factor-for-policy",
		PublicIdentifier: []byte("totp-key"),
		Status:           types.EnrollmentStatusActive,
		EnrolledAt:       time.Now().Unix(),
	}
	err = s.keeper.EnrollFactor(s.ctx, enrollment)
	s.Require().NoError(err)
	err = enrollActiveFactor(s.keeper, s.ctx, address, types.FactorTypeVEID, "veid-for-policy")
	s.Require().NoError(err)

	msg := &types.MsgSetMFAPolicy{
		Sender: address.String(),
		Policy: types.MFAPolicy{
			AccountAddress: address.String(),
			Enabled:        true,
			RequiredFactors: []types.FactorCombination{
				{Factors: []types.FactorType{types.FactorTypeTOTP, types.FactorTypeVEID}},
			},
			VEIDThreshold: 50,
		},
	}

	resp, err := s.msgServer.SetMFAPolicy(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().True(resp.Success)

	// Verify policy was stored
	policy, found := s.keeper.GetMFAPolicy(s.ctx, address)
	s.Require().True(found)
	s.Require().True(policy.Enabled)
	s.Require().Equal(uint32(50), policy.VEIDThreshold)
}

// Test: SetMFAPolicy - wrong account
func (s *MsgServerTestSuite) TestSetMFAPolicy_WrongAccount() {
	address := sdk.AccAddress([]byte("test-policy-wrong"))
	otherAddress := sdk.AccAddress([]byte("other-account"))

	msg := &types.MsgSetMFAPolicy{
		Sender: address.String(),
		Policy: types.MFAPolicy{
			AccountAddress: otherAddress.String(), // Different from sender
			Enabled:        true,
		},
	}

	_, err := s.msgServer.SetMFAPolicy(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "only set policy for own account")
}

// Test: SetMFAPolicy - enable without factors
func (s *MsgServerTestSuite) TestSetMFAPolicy_NoFactors() {
	address := sdk.AccAddress([]byte("test-policy-nofact"))

	// Require at least one factor
	params := types.DefaultParams()
	params.RequireAtLeastOneFactor = true
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	msg := &types.MsgSetMFAPolicy{
		Sender: address.String(),
		Policy: types.MFAPolicy{
			AccountAddress: address.String(),
			Enabled:        true,
		},
	}

	_, err = s.msgServer.SetMFAPolicy(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "must enroll at least one factor")
}

// Test: generic CreateChallenge rejects factors that require verifier-receipt helpers.
func (s *MsgServerTestSuite) TestCreateChallenge_RejectsOTPFactors() {
	address := sdk.AccAddress([]byte("test-challenge-addr"))

	for _, factorType := range []types.FactorType{types.FactorTypeTOTP, types.FactorTypeSMS, types.FactorTypeEmail} {
		s.Run(factorType.String(), func() {
			_, err := s.msgServer.CreateChallenge(s.ctx, &types.MsgCreateChallenge{
				Sender:          address.String(),
				FactorType:      factorType,
				TransactionType: types.SensitiveTxLargeWithdrawal,
			})
			s.Require().Error(err)
			s.Contains(err.Error(), "factor-specific helpers")
		})
	}
}

// Test: CreateChallenge - no active factor
func (s *MsgServerTestSuite) TestCreateChallenge_NoActiveFactor() {
	address := sdk.AccAddress([]byte("test-challenge-no"))

	msg := &types.MsgCreateChallenge{
		Sender:          address.String(),
		FactorType:      types.FactorTypeHardwareKey,
		TransactionType: types.SensitiveTxLargeWithdrawal,
	}

	_, err := s.msgServer.CreateChallenge(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "no active")
}

// Test: VerifyChallenge - invalid address
func (s *MsgServerTestSuite) TestVerifyChallenge_InvalidAddress() {
	msg := &types.MsgVerifyChallenge{
		Sender:      "invalid-address",
		ChallengeID: "test-challenge",
		Response:    &types.ChallengeResponse{},
	}

	_, err := s.msgServer.VerifyChallenge(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "invalid")
}

func (s *MsgServerTestSuite) TestVerifyChallenge_OTPReceiptAndOwnership() {
	address := sdk.AccAddress([]byte("msg-receipt-owner__"))
	otherAddress := sdk.AccAddress([]byte("msg-receipt-other__"))
	factorID := "msg-totp-verifier"
	publicKey, privateKey := deterministicMsgVerifierKey()
	s.Require().NoError(s.keeper.EnrollFactor(s.ctx, &types.FactorEnrollment{
		AccountAddress:   address.String(),
		FactorType:       types.FactorTypeTOTP,
		FactorID:         factorID,
		PublicIdentifier: publicKey,
		Status:           types.EnrollmentStatusActive,
		EnrolledAt:       s.ctx.BlockTime().Unix(),
		VerifiedAt:       s.ctx.BlockTime().Unix(),
	}))

	challenge, err := s.keeper.CreateTOTPChallenge(s.ctx, address, factorID, types.SensitiveTxKeyRotation)
	s.Require().NoError(err)
	response := signedMsgReceipt(challenge, privateKey)

	_, err = s.msgServer.VerifyChallenge(s.ctx, &types.MsgVerifyChallenge{
		Sender:      otherAddress.String(),
		ChallengeID: challenge.ChallengeID,
		Response:    response,
	})
	s.Require().Error(err)
	stored, found := s.keeper.GetChallenge(s.ctx, challenge.ChallengeID)
	s.Require().True(found)
	s.Zero(stored.AttemptCount)
	s.Empty(s.keeper.GetAccountSessions(s.ctx, otherAddress))

	result, err := s.msgServer.VerifyChallenge(s.ctx, &types.MsgVerifyChallenge{
		Sender:      address.String(),
		ChallengeID: challenge.ChallengeID,
		Response:    response,
	})
	s.Require().NoError(err)
	s.Require().True(result.Verified)
	s.NotEmpty(result.SessionID)
	s.Greater(result.SessionExpiresAt, s.ctx.BlockTime().Unix())
	session, found := s.keeper.GetAuthorizationSession(s.ctx, result.SessionID)
	s.Require().True(found)
	s.Equal(address.String(), session.AccountAddress)
	s.Equal(challenge.TransactionType, session.TransactionType)
}

func (s *MsgServerTestSuite) TestVerifyChallenge_FailedAttemptsCommit() {
	address := sdk.AccAddress([]byte("msg-receipt-lockout"))
	factorID := "msg-lockout-verifier"
	publicKey, privateKey := deterministicMsgVerifierKey()
	s.Require().NoError(s.keeper.EnrollFactor(s.ctx, &types.FactorEnrollment{
		AccountAddress:   address.String(),
		FactorType:       types.FactorTypeTOTP,
		FactorID:         factorID,
		PublicIdentifier: publicKey,
		Status:           types.EnrollmentStatusActive,
		EnrolledAt:       s.ctx.BlockTime().Unix(),
	}))
	challenge, err := s.keeper.CreateTOTPChallenge(s.ctx, address, factorID, types.SensitiveTxLargeWithdrawal)
	s.Require().NoError(err)
	challenge.MaxAttempts = 2
	s.Require().NoError(s.keeper.UpdateChallenge(s.ctx, challenge))
	badResponse := signedMsgReceipt(challenge, privateKey)
	badResponse.ResponseData = []byte(`{"code":"123456"}`)

	for expectedAttempts := uint32(1); expectedAttempts <= 2; expectedAttempts++ {
		result, verifyErr := s.msgServer.VerifyChallenge(s.ctx, &types.MsgVerifyChallenge{
			Sender: address.String(), ChallengeID: challenge.ChallengeID, Response: badResponse,
		})
		s.Require().NoError(verifyErr)
		s.False(result.Verified)
		stored, found := s.keeper.GetChallenge(s.ctx, challenge.ChallengeID)
		s.Require().True(found)
		s.Equal(expectedAttempts, stored.AttemptCount)
		if expectedAttempts == 1 {
			s.Equal(types.ChallengeStatusPending, stored.Status)
		} else {
			s.Equal(types.ChallengeStatusFailed, stored.Status)
		}
	}
	closed, err := s.keeper.VerifyMFAChallenge(s.ctx, challenge.ChallengeID, signedMsgReceipt(challenge, privateKey))
	s.False(closed)
	s.Error(err)
	s.Empty(s.keeper.GetAccountSessions(s.ctx, address))
}

func deterministicMsgVerifierKey() (ed25519.PublicKey, ed25519.PrivateKey) {
	seed := sha256.Sum256([]byte("virtengine-mfa-msg-verifier-test-key"))
	privateKey := ed25519.NewKeyFromSeed(seed[:])
	return privateKey.Public().(ed25519.PublicKey), privateKey
}

func signedMsgReceipt(challenge *types.Challenge, privateKey ed25519.PrivateKey) *types.ChallengeResponse {
	digest := sha256.Sum256(challenge.ChallengeData)
	receipt := keeper.OTPVerifierReceipt{
		Version: 1, ChainID: "virtengine-test-1", AccountAddress: challenge.AccountAddress,
		ChallengeID: challenge.ChallengeID, FactorType: challenge.FactorType, FactorID: challenge.FactorID,
		TransactionType: challenge.TransactionType, ChallengeNonce: challenge.Nonce,
		ChallengeDataDigest: digest[:], IssuedAt: challenge.CreatedAt, ExpiresAt: challenge.ExpiresAt,
		VerifierKeyEpoch: 1, ReceiptNonce: "msg-receipt-" + challenge.ChallengeID,
	}
	signBytes, err := receipt.SignBytes()
	if err != nil {
		panic(err)
	}
	receipt.Signature = ed25519.Sign(privateKey, signBytes)
	data, err := json.Marshal(receipt)
	if err != nil {
		panic(err)
	}
	return &types.ChallengeResponse{
		ChallengeID: challenge.ChallengeID, FactorType: challenge.FactorType,
		ResponseData: data, Timestamp: receipt.IssuedAt,
	}
}

// Test: AddTrustedDevice - success
func (s *MsgServerTestSuite) TestAddTrustedDevice_Success() {
	address := sdk.AccAddress([]byte("test-trusted-dev"))

	// Create a valid session first
	session := &types.AuthorizationSession{
		SessionID:       "session-for-device",
		AccountAddress:  address.String(),
		TransactionType: types.SensitiveTxHighValueOrder,
		CreatedAt:       s.ctx.BlockTime().Unix(),
		ExpiresAt:       s.ctx.BlockTime().Unix() + 3600,
		VerifiedFactors: []types.FactorType{types.FactorTypeTOTP},
	}
	err := s.keeper.CreateAuthorizationSession(s.ctx, session)
	s.Require().NoError(err)

	msg := &types.MsgAddTrustedDevice{
		Sender: address.String(),
		DeviceInfo: types.DeviceInfo{
			Fingerprint: "my-iphone-fp",
			UserAgent:   "iPhone 15",
		},
		MFAProof: &types.MFAProof{
			SessionID:       "session-for-device",
			VerifiedFactors: []types.FactorType{types.FactorTypeTOTP},
			Timestamp:       s.ctx.BlockTime().Unix(),
		},
	}

	resp, err := s.msgServer.AddTrustedDevice(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().True(resp.Success)
}

// Test: RemoveTrustedDevice - success
func (s *MsgServerTestSuite) TestRemoveTrustedDevice_Success() {
	address := sdk.AccAddress([]byte("test-remove-device"))

	// Add a trusted device first
	deviceInfo := &types.DeviceInfo{
		Fingerprint: "device-to-remove",
		UserAgent:   "Old Device",
	}
	_, err := s.keeper.AddTrustedDevice(s.ctx, address, deviceInfo)
	s.Require().NoError(err)

	msg := &types.MsgRemoveTrustedDevice{
		Sender:            address.String(),
		DeviceFingerprint: "device-to-remove",
	}

	resp, err := s.msgServer.RemoveTrustedDevice(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().True(resp.Success)

	// Verify device was removed
	devices := s.keeper.GetTrustedDevices(s.ctx, address)
	found := false
	for _, d := range devices {
		if d.DeviceInfo.Fingerprint == "device-to-remove" {
			found = true
		}
	}
	s.Require().False(found)
}

// Test: UpdateSensitiveTxConfig - success
func (s *MsgServerTestSuite) TestUpdateSensitiveTxConfig_Success() {
	msg := &types.MsgUpdateSensitiveTxConfig{
		Authority: "authority",
		Config: types.SensitiveTxConfig{
			TransactionType: types.SensitiveTxLargeWithdrawal,
			Enabled:         true,
			Description:     "Provider withdrawals require MFA",
			RequiredFactorCombinations: []types.FactorCombination{
				{Factors: []types.FactorType{types.FactorTypeTOTP}},
			},
		},
	}

	resp, err := s.msgServer.UpdateSensitiveTxConfig(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().True(resp.Success)
}

// Test: UpdateSensitiveTxConfig - unauthorized
func (s *MsgServerTestSuite) TestUpdateSensitiveTxConfig_Unauthorized() {
	msg := &types.MsgUpdateSensitiveTxConfig{
		Authority: "wrong-authority",
		Config: types.SensitiveTxConfig{
			TransactionType: types.SensitiveTxLargeWithdrawal,
			Enabled:         true,
		},
	}

	_, err := s.msgServer.UpdateSensitiveTxConfig(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "expected")
}

// Test: IssueSession - success
func (s *MsgServerTestSuite) TestIssueSession_Success() {
	address := sdk.AccAddress([]byte("test-issue-session"))

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

	session := &types.AuthorizationSession{
		SessionID:       "issue-source-session",
		AccountAddress:  address.String(),
		TransactionType: types.SensitiveTxLargeWithdrawal,
		CreatedAt:       s.ctx.BlockTime().Unix(),
		ExpiresAt:       s.ctx.BlockTime().Unix() + 3600,
		VerifiedFactors: []types.FactorType{types.FactorTypeTOTP},
	}
	err = s.keeper.CreateAuthorizationSession(s.ctx, session)
	s.Require().NoError(err)

	msg := &types.MsgIssueSession{
		Sender:          address.String(),
		TransactionType: types.SensitiveTxMediumWithdrawal,
		MFAProof: &types.MFAProof{
			SessionID:       "issue-source-session",
			VerifiedFactors: []types.FactorType{types.FactorTypeTOTP},
			Timestamp:       s.ctx.BlockTime().Unix(),
		},
	}

	resp, err := s.msgServer.IssueSession(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotEmpty(resp.SessionID)
}

// Test: RefreshSession - success
func (s *MsgServerTestSuite) TestRefreshSession_Success() {
	address := sdk.AccAddress([]byte("test-refresh-session"))

	err := s.keeper.SetSensitiveTxConfig(s.ctx, &types.SensitiveTxConfig{
		TransactionType: types.SensitiveTxHighValueOrder,
		Enabled:         true,
		RequiredFactorCombinations: []types.FactorCombination{
			{Factors: []types.FactorType{types.FactorTypeTOTP}},
		},
	})
	s.Require().NoError(err)

	session := &types.AuthorizationSession{
		SessionID:       "refresh-session-id",
		AccountAddress:  address.String(),
		TransactionType: types.SensitiveTxHighValueOrder,
		CreatedAt:       s.ctx.BlockTime().Unix(),
		ExpiresAt:       s.ctx.BlockTime().Unix() + 60,
		VerifiedFactors: []types.FactorType{types.FactorTypeTOTP},
	}
	err = s.keeper.CreateAuthorizationSession(s.ctx, session)
	s.Require().NoError(err)

	s.ctx = s.ctx.WithBlockTime(s.ctx.BlockTime().Add(time.Second))

	msg := &types.MsgRefreshSession{
		Sender:    address.String(),
		SessionID: "refresh-session-id",
		MFAProof: &types.MFAProof{
			SessionID:       "refresh-session-id",
			VerifiedFactors: []types.FactorType{types.FactorTypeTOTP},
			Timestamp:       s.ctx.BlockTime().Unix(),
		},
	}

	resp, err := s.msgServer.RefreshSession(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().Equal("refresh-session-id", resp.SessionID)
	s.Require().Greater(resp.SessionExpiresAt, session.ExpiresAt)
}

// Test: RevokeSession - success
func (s *MsgServerTestSuite) TestRevokeSession_Success() {
	address := sdk.AccAddress([]byte("test-revoke-session"))

	session := &types.AuthorizationSession{
		SessionID:       "revoke-session-id",
		AccountAddress:  address.String(),
		TransactionType: types.SensitiveTxHighValueOrder,
		CreatedAt:       s.ctx.BlockTime().Unix(),
		ExpiresAt:       s.ctx.BlockTime().Unix() + 3600,
		VerifiedFactors: []types.FactorType{types.FactorTypeTOTP},
	}
	err := s.keeper.CreateAuthorizationSession(s.ctx, session)
	s.Require().NoError(err)

	msg := &types.MsgRevokeSession{
		Sender:    address.String(),
		SessionID: "revoke-session-id",
	}

	resp, err := s.msgServer.RevokeSession(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().True(resp.Success)

	_, found := s.keeper.GetAuthorizationSession(s.ctx, "revoke-session-id")
	s.Require().False(found)
}
