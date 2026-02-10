// Package keeper_test provides tests for VEID module keeper.
//
// This file tests the appeal system for contesting verification decisions.
//
// Task Reference: VE-3020 - Appeal and Dispute System
package keeper_test

import (
	"crypto/rand"
	"crypto/sha256"
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

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	"github.com/virtengine/virtengine/x/veid/keeper"
	"github.com/virtengine/virtengine/x/veid/types"
)

// AppealTestSuite is the test suite for appeal functionality
type AppealTestSuite struct {
	suite.Suite
	ctx        sdk.Context
	keeper     keeper.Keeper
	cdc        codec.Codec
	authority  string
	stateStore store.CommitMultiStore
}

func TestAppealTestSuite(t *testing.T) {
	suite.Run(t, new(AppealTestSuite))
}

func (s *AppealTestSuite) SetupTest() {
	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	s.cdc = codec.NewProtoCodec(interfaceRegistry)

	// Create store keys
	veidStoreKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Set authority
	s.authority = "cosmos1authority_account______"

	// Create context with store
	s.ctx = s.createContextWithStore(veidStoreKey)

	// Create veid keeper
	s.keeper = keeper.NewKeeper(s.cdc, veidStoreKey, s.authority)

	// Set default params
	err := s.keeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)

	// Set default appeal params
	err = s.keeper.SetAppealParams(s.ctx, types.DefaultAppealParams())
	s.Require().NoError(err)
}

func (s *AppealTestSuite) createContextWithStore(storeKey *storetypes.KVStoreKey) sdk.Context {
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	s.stateStore = stateStore
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

// Helper to create a valid test address
// TearDownTest closes the IAVL store to stop background pruning goroutines
func (s *AppealTestSuite) TearDownTest() {
	CloseStoreIfNeeded(s.stateStore)
}

func (s *AppealTestSuite) testAddress(seed string) sdk.AccAddress {
	return sdk.AccAddress([]byte(seed))
}

func (s *AppealTestSuite) generateSalt() []byte {
	salt := make([]byte, 32)
	_, _ = rand.Read(salt)
	return salt
}

func (s *AppealTestSuite) generateSignature() []byte {
	sig := make([]byte, 64)
	_, _ = rand.Read(sig)
	return sig
}

func (s *AppealTestSuite) createTestPayload() encryptiontypes.EncryptedPayloadEnvelope {
	nonce := make([]byte, 24)
	_, _ = rand.Read(nonce)

	ciphertext := make([]byte, 128)
	_, _ = rand.Read(ciphertext)

	pubKey := make([]byte, 32)
	_, _ = rand.Read(pubKey)

	senderSignature := make([]byte, 64)
	_, _ = rand.Read(senderSignature)

	return encryptiontypes.EncryptedPayloadEnvelope{
		Version:          1,
		AlgorithmID:      "X25519-XSALSA20-POLY1305",
		AlgorithmVersion: encryptiontypes.AlgorithmVersionV1,
		RecipientKeyIDs:  []string{"recipient1"},
		Nonce:            nonce,
		Ciphertext:       ciphertext,
		SenderSignature:  senderSignature,
		SenderPubKey:     pubKey,
	}
}

func (s *AppealTestSuite) createTestUploadMetadata() types.UploadMetadata {
	salt := s.generateSalt()
	payload := s.createTestPayload()
	payloadHash := sha256.Sum256(payload.Ciphertext)

	return types.UploadMetadata{
		Salt:              salt,
		SaltHash:          types.ComputeSaltHash(salt),
		DeviceFingerprint: "test-device-fp",
		ClientID:          "test-client",
		ClientSignature:   s.generateSignature(),
		UserSignature:     s.generateSignature(),
		PayloadHash:       payloadHash[:],
	}
}

// Helper to create a rejected scope for testing appeals
func (s *AppealTestSuite) createRejectedScope(address sdk.AccAddress, scopeID string) {
	// Use relaxed params for testing
	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	metadata := s.createTestUploadMetadata()
	scope := types.NewIdentityScope(
		scopeID,
		types.ScopeTypeSelfie,
		s.createTestPayload(),
		metadata,
		time.Now(),
	)
	err = s.keeper.UploadScope(s.ctx, address, scope)
	s.Require().NoError(err)

	// Mark scope as rejected
	err = s.keeper.UpdateVerificationStatus(s.ctx, address, scopeID, types.VerificationStatusRejected, "test rejection", "")
	s.Require().NoError(err)
}

// Helper to create an identity record with score
func (s *AppealTestSuite) createIdentityWithScore(address sdk.AccAddress, score uint32) {
	_, err := s.keeper.CreateIdentityRecord(s.ctx, address)
	s.Require().NoError(err)
	// Use SetScore to store in the score store that GetIdentityScore reads from
	err = s.keeper.SetScore(s.ctx, address.String(), score, "v1.0.0")
	s.Require().NoError(err)
}

// ============================================================================
// Test Submit Appeal
// ============================================================================

func (s *AppealTestSuite) TestSubmitAppeal_Success() {
	address := s.testAddress("appeal_submit_success____")
	scopeID := "scope_001"

	// Setup: Create identity with score and rejected scope
	s.createIdentityWithScore(address, 75)
	s.createRejectedScope(address, scopeID)

	// Submit appeal
	msg := types.NewMsgSubmitAppeal(
		address.String(),
		scopeID,
		"I believe my verification was incorrectly rejected due to poor lighting conditions during the capture process. The attached evidence shows the correct conditions.",
		[]string{"a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"},
	)

	appeal, err := s.keeper.SubmitAppeal(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(appeal)
	s.Require().NotEmpty(appeal.AppealID)
	s.Require().Equal(address.String(), appeal.AccountAddress)
	s.Require().Equal(scopeID, appeal.ScopeID)
	s.Require().Equal(types.AppealStatusPending, appeal.Status)
	s.Require().Equal(uint32(1), appeal.AppealNumber)
	s.Require().Equal(uint32(75), appeal.OriginalScore)
}

func (s *AppealTestSuite) TestSubmitAppeal_ScopeNotRejected() {
	address := s.testAddress("appeal_not_rejected______")
	scopeID := "scope_002"

	// Use relaxed params for testing
	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	// Setup: Create identity and verified scope (not rejected)
	s.createIdentityWithScore(address, 90)
	metadata := s.createTestUploadMetadata()
	scope := types.NewIdentityScope(
		scopeID,
		types.ScopeTypeSelfie,
		s.createTestPayload(),
		metadata,
		time.Now(),
	)
	err = s.keeper.UploadScope(s.ctx, address, scope)
	s.Require().NoError(err)

	// Try to submit appeal - should fail
	msg := types.NewMsgSubmitAppeal(
		address.String(),
		scopeID,
		"This is a test appeal reason that is long enough to meet the minimum character requirement for appeals.",
		nil,
	)

	_, err = s.keeper.SubmitAppeal(s.ctx, msg)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrScopeNotRejected)
}

func (s *AppealTestSuite) TestSubmitAppeal_ReasonTooShort() {
	// Note: MsgSubmitAppeal.ValidateBasic() only checks for empty reason and max length (2000)
	// Minimum length validation is handled at the keeper level via Appeal.Validate()
	msg := &types.MsgSubmitAppeal{
		Submitter: s.testAddress("appeal_short_reason______").String(),
		ScopeId:   "scope_003",
		Reason:    "", // Empty reason should fail
	}

	err := msg.ValidateBasic()
	s.Require().Error(err)
}

func (s *AppealTestSuite) TestSubmitAppeal_DuplicateAppeal() {
	address := s.testAddress("appeal_duplicate_________")
	scopeID := "scope_004"

	// Setup
	s.createIdentityWithScore(address, 70)
	s.createRejectedScope(address, scopeID)

	// Submit first appeal
	msg := types.NewMsgSubmitAppeal(
		address.String(),
		scopeID,
		"First appeal with sufficient reason length to meet the minimum character requirement for the appeal system.",
		nil,
	)

	_, err := s.keeper.SubmitAppeal(s.ctx, msg)
	s.Require().NoError(err)

	// Try to submit second appeal while first is pending
	_, err = s.keeper.SubmitAppeal(s.ctx, msg)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrAppealAlreadyExists)
}

// ============================================================================
// Test Resolve Appeal - Approved
// ============================================================================

func (s *AppealTestSuite) TestResolveAppealApproved() {
	address := s.testAddress("appeal_approve___________")
	scopeID := "scope_005"

	// Setup
	s.createIdentityWithScore(address, 70)
	s.createRejectedScope(address, scopeID)

	// Submit appeal
	submitMsg := types.NewMsgSubmitAppeal(
		address.String(),
		scopeID,
		"I believe my verification was incorrectly rejected. Please review the evidence and reconsider my case.",
		nil,
	)
	appeal, err := s.keeper.SubmitAppeal(s.ctx, submitMsg)
	s.Require().NoError(err)

	// Resolve appeal as approved with score adjustment
	resolveMsg := types.NewMsgResolveAppeal(
		s.authority, // Authority can resolve
		appeal.AppealID,
		types.AppealStatusApproved,
		"After review, the appeal is valid. The original rejection was in error.",
		15, // Increase score by 15
	)

	err = s.keeper.ResolveAppeal(s.ctx, resolveMsg)
	s.Require().NoError(err)

	// Verify appeal was resolved
	resolved, found := s.keeper.GetAppeal(s.ctx, appeal.AppealID)
	s.Require().True(found)
	s.Require().Equal(types.AppealStatusApproved, resolved.Status)
	s.Require().Equal(int32(15), resolved.ScoreAdjustment)
	s.Require().NotZero(resolved.ReviewedAt)

	// Verify score was adjusted
	score, found := s.keeper.GetIdentityScore(s.ctx, address.String())
	s.Require().True(found)
	s.Require().Equal(uint32(85), score.Score) // 70 + 15 = 85
}

// ============================================================================
// Test Resolve Appeal - Rejected
// ============================================================================

func (s *AppealTestSuite) TestResolveAppealRejected() {
	address := s.testAddress("appeal_reject____________")
	scopeID := "scope_006"

	// Setup
	s.createIdentityWithScore(address, 65)
	s.createRejectedScope(address, scopeID)

	// Submit appeal
	submitMsg := types.NewMsgSubmitAppeal(
		address.String(),
		scopeID,
		"I believe my verification was incorrectly rejected. Please review the evidence and reconsider my case.",
		nil,
	)
	appeal, err := s.keeper.SubmitAppeal(s.ctx, submitMsg)
	s.Require().NoError(err)

	// Resolve appeal as rejected
	resolveMsg := types.NewMsgResolveAppeal(
		s.authority,
		appeal.AppealID,
		types.AppealStatusRejected,
		"After careful review, the original rejection was correct. The submitted evidence does not support the appeal.",
		0, // No score adjustment
	)

	err = s.keeper.ResolveAppeal(s.ctx, resolveMsg)
	s.Require().NoError(err)

	// Verify appeal was rejected
	resolved, found := s.keeper.GetAppeal(s.ctx, appeal.AppealID)
	s.Require().True(found)
	s.Require().Equal(types.AppealStatusRejected, resolved.Status)

	// Verify score was NOT adjusted
	score, found := s.keeper.GetIdentityScore(s.ctx, address.String())
	s.Require().True(found)
	s.Require().Equal(uint32(65), score.Score) // Unchanged
}

// ============================================================================
// Test Withdraw Appeal
// ============================================================================

func (s *AppealTestSuite) TestWithdrawAppeal() {
	address := s.testAddress("appeal_withdraw__________")
	scopeID := "scope_007"

	// Setup
	s.createIdentityWithScore(address, 75)
	s.createRejectedScope(address, scopeID)

	// Submit appeal
	submitMsg := types.NewMsgSubmitAppeal(
		address.String(),
		scopeID,
		"I believe my verification was incorrectly rejected. Please review the evidence and reconsider my case.",
		nil,
	)
	appeal, err := s.keeper.SubmitAppeal(s.ctx, submitMsg)
	s.Require().NoError(err)

	// Withdraw appeal
	withdrawMsg := types.NewMsgWithdrawAppeal(address.String(), appeal.AppealID)
	err = s.keeper.WithdrawAppeal(s.ctx, withdrawMsg)
	s.Require().NoError(err)

	// Verify appeal was withdrawn
	withdrawn, found := s.keeper.GetAppeal(s.ctx, appeal.AppealID)
	s.Require().True(found)
	s.Require().Equal(types.AppealStatusWithdrawn, withdrawn.Status)

	// Verify it's no longer in pending queue
	pending := s.keeper.GetPendingAppeals(s.ctx)
	for _, p := range pending {
		s.Require().NotEqual(appeal.AppealID, p.AppealID)
	}
}

func (s *AppealTestSuite) TestWithdrawAppeal_NotSubmitter() {
	address := s.testAddress("appeal_withdraw_wrong____")
	otherAddress := s.testAddress("appeal_other_user________")
	scopeID := "scope_008"

	// Setup
	s.createIdentityWithScore(address, 75)
	s.createRejectedScope(address, scopeID)

	// Submit appeal
	submitMsg := types.NewMsgSubmitAppeal(
		address.String(),
		scopeID,
		"I believe my verification was incorrectly rejected. Please review the evidence and reconsider my case.",
		nil,
	)
	appeal, err := s.keeper.SubmitAppeal(s.ctx, submitMsg)
	s.Require().NoError(err)

	// Try to withdraw as different user
	withdrawMsg := types.NewMsgWithdrawAppeal(otherAddress.String(), appeal.AppealID)
	err = s.keeper.WithdrawAppeal(s.ctx, withdrawMsg)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrNotAppealSubmitter)
}

// ============================================================================
// Test Appeal Window Expired
// ============================================================================

func (s *AppealTestSuite) TestAppealWindowExpired() {
	address := s.testAddress("appeal_window_expired____")
	scopeID := "scope_009"

	// Use relaxed params for testing
	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	// Setup: Create identity
	s.createIdentityWithScore(address, 70)

	// Create scope rejected long ago (beyond appeal window)
	metadata := s.createTestUploadMetadata()
	scope := types.NewIdentityScope(
		scopeID,
		types.ScopeTypeSelfie,
		s.createTestPayload(),
		metadata,
		time.Unix(1, 0), // Very early time
	)
	err = s.keeper.UploadScope(s.ctx, address, scope)
	s.Require().NoError(err)

	// Mark as rejected
	err = s.keeper.UpdateVerificationStatus(s.ctx, address, scopeID, types.VerificationStatusRejected, "test rejection", "")
	s.Require().NoError(err)

	// Update context to be way past appeal window
	s.ctx = s.ctx.WithBlockHeight(types.DefaultAppealWindowBlocks + 100)

	// Try to submit appeal - should fail due to expired window
	msg := types.NewMsgSubmitAppeal(
		address.String(),
		scopeID,
		"I believe my verification was incorrectly rejected. Please review the evidence and reconsider my case.",
		nil,
	)

	_, err = s.keeper.SubmitAppeal(s.ctx, msg)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrAppealWindowExpired)
}

// ============================================================================
// Test Max Appeals Exceeded
// ============================================================================

func (s *AppealTestSuite) TestMaxAppealsExceeded() {
	address := s.testAddress("appeal_max_exceeded______")
	scopeID := "scope_010"

	// Setup
	s.createIdentityWithScore(address, 70)
	s.createRejectedScope(address, scopeID)

	// Set max appeals to 2 for testing
	params := types.DefaultAppealParams()
	params.MaxAppealsPerScope = 2
	err := s.keeper.SetAppealParams(s.ctx, params)
	s.Require().NoError(err)

	// Submit first appeal
	msg1 := types.NewMsgSubmitAppeal(
		address.String(),
		scopeID,
		"First appeal with sufficient reason length to meet the minimum character requirement for the appeal system.",
		nil,
	)
	appeal1, err := s.keeper.SubmitAppeal(s.ctx, msg1)
	s.Require().NoError(err)

	// Withdraw first appeal so we can submit another
	err = s.keeper.WithdrawAppeal(s.ctx, types.NewMsgWithdrawAppeal(address.String(), appeal1.AppealID))
	s.Require().NoError(err)

	// Submit second appeal
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1)
	appeal2, err := s.keeper.SubmitAppeal(s.ctx, msg1)
	s.Require().NoError(err)

	// Withdraw second
	err = s.keeper.WithdrawAppeal(s.ctx, types.NewMsgWithdrawAppeal(address.String(), appeal2.AppealID))
	s.Require().NoError(err)

	// Submit third appeal - should fail
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1)
	_, err = s.keeper.SubmitAppeal(s.ctx, msg1)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrMaxAppealsExceeded)
}

// ============================================================================
// Test Unauthorized Resolver
// ============================================================================

func (s *AppealTestSuite) TestUnauthorizedResolver() {
	address := s.testAddress("appeal_unauth_resolver___")
	scopeID := "scope_011"

	// Setup
	s.createIdentityWithScore(address, 70)
	s.createRejectedScope(address, scopeID)

	// Submit appeal
	submitMsg := types.NewMsgSubmitAppeal(
		address.String(),
		scopeID,
		"I believe my verification was incorrectly rejected. Please review the evidence and reconsider my case.",
		nil,
	)
	appeal, err := s.keeper.SubmitAppeal(s.ctx, submitMsg)
	s.Require().NoError(err)

	// Try to resolve as unauthorized user
	unauthorizedAddr := s.testAddress("unauthorized_resolver____")
	resolveMsg := types.NewMsgResolveAppeal(
		unauthorizedAddr.String(),
		appeal.AppealID,
		types.AppealStatusApproved,
		"Attempting unauthorized resolution",
		10,
	)

	err = s.keeper.ResolveAppeal(s.ctx, resolveMsg)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrNotAuthorizedResolver)
}

// ============================================================================
// Test Score Adjustment On Approval
// ============================================================================

func (s *AppealTestSuite) TestScoreAdjustmentOnApproval() {
	address := s.testAddress("appeal_score_adjust______")
	scopeID := "scope_012"

	// Setup with specific score
	s.createIdentityWithScore(address, 50)
	s.createRejectedScope(address, scopeID)

	// Submit appeal
	submitMsg := types.NewMsgSubmitAppeal(
		address.String(),
		scopeID,
		"I believe my verification was incorrectly rejected. Please review the evidence and reconsider my case.",
		nil,
	)
	appeal, err := s.keeper.SubmitAppeal(s.ctx, submitMsg)
	s.Require().NoError(err)

	// Verify original score was captured
	s.Require().Equal(uint32(50), appeal.OriginalScore)

	// Approve with large score adjustment
	resolveMsg := types.NewMsgResolveAppeal(
		s.authority,
		appeal.AppealID,
		types.AppealStatusApproved,
		"Appeal approved with significant score adjustment",
		40, // +40 points
	)

	err = s.keeper.ResolveAppeal(s.ctx, resolveMsg)
	s.Require().NoError(err)

	// Verify new score
	score, found := s.keeper.GetIdentityScore(s.ctx, address.String())
	s.Require().True(found)
	s.Require().Equal(uint32(90), score.Score) // 50 + 40 = 90
}

func (s *AppealTestSuite) TestScoreAdjustmentClamping() {
	address := s.testAddress("appeal_score_clamp_______")
	scopeID := "scope_013"

	// Setup with high score
	s.createIdentityWithScore(address, 95)
	s.createRejectedScope(address, scopeID)

	// Submit appeal
	submitMsg := types.NewMsgSubmitAppeal(
		address.String(),
		scopeID,
		"I believe my verification was incorrectly rejected. Please review the evidence and reconsider my case.",
		nil,
	)
	appeal, err := s.keeper.SubmitAppeal(s.ctx, submitMsg)
	s.Require().NoError(err)

	// Approve with score adjustment that would exceed max
	resolveMsg := types.NewMsgResolveAppeal(
		s.authority,
		appeal.AppealID,
		types.AppealStatusApproved,
		"Appeal approved",
		20, // Would be 115, should clamp to 100
	)

	err = s.keeper.ResolveAppeal(s.ctx, resolveMsg)
	s.Require().NoError(err)

	// Verify score was clamped to max
	score, found := s.keeper.GetIdentityScore(s.ctx, address.String())
	s.Require().True(found)
	s.Require().Equal(types.MaxScore, score.Score) // Clamped to 100
}

// ============================================================================
// Test Get Appeals By Account
// ============================================================================

func (s *AppealTestSuite) TestGetAppealsByAccount() {
	address := s.testAddress("appeal_get_by_account____")

	// Setup
	s.createIdentityWithScore(address, 70)

	// Create multiple rejected scopes
	for i := 1; i <= 3; i++ {
		scopeID := "scope_multi_" + string(rune('0'+i))
		s.createRejectedScope(address, scopeID)

		// Submit appeal
		msg := types.NewMsgSubmitAppeal(
			address.String(),
			scopeID,
			"Appeal reason that is long enough to meet the minimum character requirement for the system.",
			nil,
		)
		_, err := s.keeper.SubmitAppeal(s.ctx, msg)
		s.Require().NoError(err)

		// Increment block height
		s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1)
	}

	// Get all appeals for account
	appeals := s.keeper.GetAppealsByAccount(s.ctx, address.String())
	s.Require().Len(appeals, 3)
}

// ============================================================================
// Test Get Pending Appeals
// ============================================================================

func (s *AppealTestSuite) TestGetPendingAppeals() {
	// Create multiple addresses and appeals
	for i := 1; i <= 3; i++ {
		address := s.testAddress("appeal_pending_" + string(rune('A'+i)) + "_______")
		scopeID := "scope_pending_" + string(rune('0'+i))

		s.createIdentityWithScore(address, 70)
		s.createRejectedScope(address, scopeID)

		msg := types.NewMsgSubmitAppeal(
			address.String(),
			scopeID,
			"Appeal reason that is long enough to meet the minimum character requirement for the system.",
			nil,
		)
		_, err := s.keeper.SubmitAppeal(s.ctx, msg)
		s.Require().NoError(err)

		s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1)
	}

	// Get pending appeals
	pending := s.keeper.GetPendingAppeals(s.ctx)
	s.Require().GreaterOrEqual(len(pending), 3)

	// All should be in pending status
	for _, appeal := range pending {
		s.Require().Equal(types.AppealStatusPending, appeal.Status)
	}
}

// ============================================================================
// Test Appeal Summary
// ============================================================================

func (s *AppealTestSuite) TestAppealSummary() {
	address := s.testAddress("appeal_summary___________")

	// Setup
	s.createIdentityWithScore(address, 70)

	// Create appeals with different statuses
	scopes := []string{"scope_sum_1", "scope_sum_2", "scope_sum_3"}
	for _, scopeID := range scopes {
		s.createRejectedScope(address, scopeID)
		msg := types.NewMsgSubmitAppeal(
			address.String(),
			scopeID,
			"Appeal reason that is long enough to meet the minimum character requirement for the system.",
			nil,
		)
		_, err := s.keeper.SubmitAppeal(s.ctx, msg)
		s.Require().NoError(err)
		s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1)
	}

	// Get summary
	summary := s.keeper.GetAppealSummary(s.ctx, address.String())
	s.Require().Equal(uint32(3), summary.TotalAppeals)
	s.Require().Equal(uint32(3), summary.PendingAppeals)
}

// ============================================================================
// Test Authorized Resolver Management
// ============================================================================

func (s *AppealTestSuite) TestAuthorizedResolverManagement() {
	resolver := s.testAddress("test_resolver____________")

	// Initially not authorized
	s.Require().False(s.keeper.IsAuthorizedResolver(s.ctx, resolver.String()))

	// Add as authorized
	s.keeper.SetAuthorizedResolver(s.ctx, resolver, true)
	s.Require().True(s.keeper.IsAuthorizedResolver(s.ctx, resolver.String()))

	// Verify in list
	resolvers := s.keeper.GetAuthorizedResolvers(s.ctx)
	found := false
	for _, r := range resolvers {
		if r == resolver.String() {
			found = true
			break
		}
	}
	s.Require().True(found)

	// Remove authorization
	s.keeper.SetAuthorizedResolver(s.ctx, resolver, false)
	s.Require().False(s.keeper.IsAuthorizedResolver(s.ctx, resolver.String()))
}

// ============================================================================
// Test Appeal Claim Flow
// ============================================================================

func (s *AppealTestSuite) TestAppealClaimFlow() {
	address := s.testAddress("appeal_claim_flow________")
	scopeID := "scope_claim"

	// Setup
	s.createIdentityWithScore(address, 70)
	s.createRejectedScope(address, scopeID)

	// Submit appeal
	submitMsg := types.NewMsgSubmitAppeal(
		address.String(),
		scopeID,
		"I believe my verification was incorrectly rejected. Please review the evidence and reconsider my case.",
		nil,
	)
	appeal, err := s.keeper.SubmitAppeal(s.ctx, submitMsg)
	s.Require().NoError(err)
	s.Require().Equal(types.AppealStatusPending, appeal.Status)

	// Add resolver
	resolver := s.testAddress("claim_resolver___________")
	s.keeper.SetAuthorizedResolver(s.ctx, resolver, true)

	// Claim appeal
	claimMsg := types.NewMsgClaimAppeal(resolver.String(), appeal.AppealID)
	err = s.keeper.ClaimAppeal(s.ctx, claimMsg)
	s.Require().NoError(err)

	// Verify status changed to reviewing
	claimed, found := s.keeper.GetAppeal(s.ctx, appeal.AppealID)
	s.Require().True(found)
	s.Require().Equal(types.AppealStatusReviewing, claimed.Status)
	s.Require().Equal(resolver.String(), claimed.ReviewerAddress)

	// Verify removed from pending queue
	pending := s.keeper.GetPendingAppeals(s.ctx)
	for _, p := range pending {
		s.Require().NotEqual(appeal.AppealID, p.AppealID)
	}

	// Resolve as the claiming resolver
	resolveMsg := types.NewMsgResolveAppeal(
		resolver.String(),
		appeal.AppealID,
		types.AppealStatusApproved,
		"Approved after careful review",
		10,
	)
	err = s.keeper.ResolveAppeal(s.ctx, resolveMsg)
	s.Require().NoError(err)
}

// ============================================================================
// Test Appeal Status Transitions
// ============================================================================

func (s *AppealTestSuite) TestAppealStatusTransitions() {
	// Test IsTerminal
	s.Require().False(types.AppealStatusPending.IsTerminal())
	s.Require().False(types.AppealStatusReviewing.IsTerminal())
	s.Require().True(types.AppealStatusApproved.IsTerminal())
	s.Require().True(types.AppealStatusRejected.IsTerminal())
	s.Require().True(types.AppealStatusWithdrawn.IsTerminal())
	s.Require().True(types.AppealStatusExpired.IsTerminal())

	// Test IsActive
	s.Require().True(types.AppealStatusPending.IsActive())
	s.Require().True(types.AppealStatusReviewing.IsActive())
	s.Require().False(types.AppealStatusApproved.IsActive())
	s.Require().False(types.AppealStatusRejected.IsActive())

	// Test String
	s.Require().Equal("PENDING", types.AppealStatusPending.String())
	s.Require().Equal("APPROVED", types.AppealStatusApproved.String())
}

// ============================================================================
// Test Appeal Params Validation
// ============================================================================

func (s *AppealTestSuite) TestAppealParamsValidation() {
	// Valid params
	params := types.DefaultAppealParams()
	s.Require().NoError(params.Validate())

	// Invalid: zero appeal window when enabled
	params = types.DefaultAppealParams()
	params.AppealWindowBlocks = 0
	s.Require().Error(params.Validate())

	// Invalid: escrow required but amount zero
	params = types.DefaultAppealParams()
	params.RequireEscrowDeposit = true
	params.EscrowDepositAmount = 0
	s.Require().Error(params.Validate())

	// Valid: disabled system doesn't need valid window
	params = types.DefaultAppealParams()
	params.Enabled = false
	params.AppealWindowBlocks = 0
	s.Require().NoError(params.Validate())
}

// ============================================================================
// Test Message Validation
// ============================================================================

func (s *AppealTestSuite) TestMsgSubmitAppealValidation() {
	// Valid message
	msg := types.NewMsgSubmitAppeal(
		s.testAddress("valid_submitter__________").String(),
		"scope_id",
		"A sufficiently long reason that meets the minimum character requirement for appeal submissions in the system.",
		[]string{"a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"},
	)
	s.Require().NoError(msg.ValidateBasic())

	// Invalid: empty scope ID
	msg.ScopeId = ""
	s.Require().Error(msg.ValidateBasic())

	// Restore valid scope_id for further tests
	msg.ScopeId = "scope_id"
	s.Require().NoError(msg.ValidateBasic())
}

func (s *AppealTestSuite) TestMsgResolveAppealValidation() {
	// Valid message
	msg := types.NewMsgResolveAppeal(
		s.testAddress("valid_resolver___________").String(),
		"appeal_123",
		types.AppealStatusApproved,
		"Valid resolution reason",
		10,
	)
	s.Require().NoError(msg.ValidateBasic())

	// Invalid: wrong resolution status (Pending is not a valid resolution)
	msg.Resolution = types.AppealStatusToProto(types.AppealStatusPending)
	s.Require().Error(msg.ValidateBasic())

	// Reset to valid resolution
	msg.Resolution = types.AppealStatusToProto(types.AppealStatusApproved)
	s.Require().NoError(msg.ValidateBasic())
}
