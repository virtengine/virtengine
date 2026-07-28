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

	mfatypes "github.com/virtengine/virtengine/x/mfa/types"
	"github.com/virtengine/virtengine/x/veid/keeper"
	"github.com/virtengine/virtengine/x/veid/types"
)

// Test address variables - valid bech32 addresses
var (
	testBorderlineAddress = sdk.AccAddress([]byte("borderline_address__")).String()
)

// MockMFAKeeper is a mock implementation of MFAKeeper for testing
type MockMFAKeeper struct {
	challenges   map[string]*mfatypes.Challenge
	enrollments  map[string][]mfatypes.FactorEnrollment
	params       mfatypes.Params
	verifyResult bool
	verifyErr    error
}

func NewMockMFAKeeper() *MockMFAKeeper {
	return &MockMFAKeeper{
		challenges:  make(map[string]*mfatypes.Challenge),
		enrollments: make(map[string][]mfatypes.FactorEnrollment),
		params:      mfatypes.DefaultParams(),
	}
}

func (m *MockMFAKeeper) CreateChallenge(ctx sdk.Context, challenge *mfatypes.Challenge) error {
	m.challenges[challenge.ChallengeID] = challenge
	return nil
}

func (m *MockMFAKeeper) GetChallenge(ctx sdk.Context, challengeID string) (*mfatypes.Challenge, bool) {
	c, found := m.challenges[challengeID]
	return c, found
}

func (m *MockMFAKeeper) GetFactorEnrollments(ctx sdk.Context, address sdk.AccAddress) []mfatypes.FactorEnrollment {
	return m.enrollments[address.String()]
}

func (m *MockMFAKeeper) HasActiveFactorOfType(ctx sdk.Context, address sdk.AccAddress, factorType mfatypes.FactorType) bool {
	enrollments := m.enrollments[address.String()]
	for _, e := range enrollments {
		if e.FactorType == factorType && e.IsActive() {
			return true
		}
	}
	return false
}

func (m *MockMFAKeeper) VerifyMFAChallenge(ctx sdk.Context, challengeID string, response *mfatypes.ChallengeResponse) (bool, error) {
	return m.verifyResult, m.verifyErr
}

func (m *MockMFAKeeper) GetParams(ctx sdk.Context) mfatypes.Params {
	return m.params
}

// AddEnrollment adds a factor enrollment for testing
func (m *MockMFAKeeper) AddEnrollment(address string, enrollment mfatypes.FactorEnrollment) {
	m.enrollments[address] = append(m.enrollments[address], enrollment)
}

// SetChallengeVerified marks a challenge as verified for testing
func (m *MockMFAKeeper) SetChallengeVerified(challengeID string) {
	if c, found := m.challenges[challengeID]; found {
		c.Status = mfatypes.ChallengeStatusVerified
		c.VerifiedAt = time.Now().Unix()
	}
}

// BorderlineFallbackTestSuite is the test suite for borderline fallback
type BorderlineFallbackTestSuite struct {
	suite.Suite
	ctx        sdk.Context
	keeper     keeper.Keeper
	mfaKeeper  *MockMFAKeeper
	cdc        codec.Codec
	stateStore store.CommitMultiStore
}

func TestBorderlineFallbackTestSuite(t *testing.T) {
	suite.Run(t, new(BorderlineFallbackTestSuite))
}

func (s *BorderlineFallbackTestSuite) SetupTest() {
	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	mfatypes.RegisterInterfaces(interfaceRegistry)
	s.cdc = codec.NewProtoCodec(interfaceRegistry)

	// Create store keys
	veidStoreKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create context with store
	s.ctx = s.createContextWithStore(veidStoreKey)

	// Create veid keeper
	s.keeper = keeper.NewKeeper(s.cdc, veidStoreKey, "authority")

	// Create and set mock MFA keeper
	s.mfaKeeper = NewMockMFAKeeper()
	s.keeper.SetMFAKeeper(s.mfaKeeper)

	// Set default params
	err := s.keeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)

	// Set default borderline params
	err = s.keeper.SetBorderlineParams(s.ctx, types.DefaultBorderlineParams())
	s.Require().NoError(err)
}

func (s *BorderlineFallbackTestSuite) createContextWithStore(storeKey *storetypes.KVStoreKey) sdk.Context {
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	if err != nil {
		s.T().Fatalf("failed to load latest version: %v", err)
	}
	s.stateStore = stateStore

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())
	return ctx
}

// TearDownTest closes the IAVL store to stop background pruning goroutines
func (s *BorderlineFallbackTestSuite) TearDownTest() {
	CloseStoreIfNeeded(s.stateStore)
}

// ============================================================================
// Borderline Detection Tests
// ============================================================================

func (s *BorderlineFallbackTestSuite) TestCheckBorderlineAndTriggerFallback_AboveUpperThreshold() {
	// Setup: Add TOTP enrollment for the account
	s.addTOTPEnrollment(testBorderlineAddress)

	// Score above upper threshold (90) should return Verified directly
	status, err := s.keeper.CheckBorderlineAndTriggerFallback(s.ctx, testBorderlineAddress, 95)
	s.Require().NoError(err)
	s.Require().Equal(types.VerificationStatusVerified, status)
}

func (s *BorderlineFallbackTestSuite) TestCheckBorderlineAndTriggerFallback_InBorderlineBand() {
	// Setup: Add TOTP enrollment for the account
	s.addTOTPEnrollment(testBorderlineAddress)

	// Score in borderline band (85-90) should trigger fallback
	status, err := s.keeper.CheckBorderlineAndTriggerFallback(s.ctx, testBorderlineAddress, 87)
	s.Require().NoError(err)
	s.Require().Equal(types.VerificationStatusNeedsAdditionalFactor, status)

	// Verify a fallback record was created
	fallbacks := s.keeper.GetPendingFallbacksForAccount(s.ctx, testBorderlineAddress)
	s.Require().Len(fallbacks, 1)
	s.Require().Equal(uint32(87), fallbacks[0].BorderlineScore)
	s.Require().Equal(types.BorderlineFallbackStatusPending, fallbacks[0].Status)
}

func (s *BorderlineFallbackTestSuite) TestCheckBorderlineAndTriggerFallback_BelowLowerThreshold() {
	// Score below lower threshold (85) should return Rejected
	status, err := s.keeper.CheckBorderlineAndTriggerFallback(s.ctx, testBorderlineAddress, 70)
	s.Require().NoError(err)
	s.Require().Equal(types.VerificationStatusRejected, status)
}

func (s *BorderlineFallbackTestSuite) TestCheckBorderlineAndTriggerFallback_NoEnrolledFactors() {
	// No MFA factors enrolled should return Rejected with error
	status, err := s.keeper.CheckBorderlineAndTriggerFallback(s.ctx, testBorderlineAddress, 87)
	s.Require().Error(err)
	s.Require().Equal(types.VerificationStatusRejected, status)
	s.Require().ErrorIs(err, types.ErrNoEnrolledFactors)
}

func (s *BorderlineFallbackTestSuite) TestCheckBorderlineAndTriggerFallback_BorderlineDisabled() {
	// Disable borderline by setting thresholds to zero width (lower == upper)
	params := types.BorderlineParams{
		LowerThreshold:   90,
		UpperThreshold:   90, // No borderline band when lower == upper
		MfaTimeoutBlocks: 100,
		RequiredFactors:  1,
	}
	err := s.keeper.SetBorderlineParams(s.ctx, params)
	s.Require().NoError(err)

	// Score in borderline band should return Rejected when disabled
	status, err := s.keeper.CheckBorderlineAndTriggerFallback(s.ctx, testBorderlineAddress, 87)
	s.Require().NoError(err)
	s.Require().Equal(types.VerificationStatusRejected, status)
}

// ============================================================================
// Borderline Fallback Flow Integration Test
// ============================================================================

func (s *BorderlineFallbackTestSuite) TestBorderlineFallbackFlow() {
	// Step 1: Setup - Add TOTP enrollment for the account
	s.addTOTPEnrollment(testBorderlineAddress)

	// Step 2: Submit identity with borderline score (87)
	status, err := s.keeper.CheckBorderlineAndTriggerFallback(s.ctx, testBorderlineAddress, 87)
	s.Require().NoError(err)
	s.Require().Equal(types.VerificationStatusNeedsAdditionalFactor, status)

	// Verify fallback was created
	fallbacks := s.keeper.GetPendingFallbacksForAccount(s.ctx, testBorderlineAddress)
	s.Require().Len(fallbacks, 1)
	fallback := fallbacks[0]
	s.Require().Equal(types.BorderlineFallbackStatusPending, fallback.Status)
	s.Require().Equal(uint32(87), fallback.BorderlineScore)

	// Step 3: Simulate MFA challenge completion
	s.mfaKeeper.SetChallengeVerified(fallback.ChallengeID)

	// Step 4: Complete the borderline fallback
	err = s.keeper.HandleBorderlineFallbackCompleted(
		s.ctx,
		testBorderlineAddress,
		fallback.ChallengeID,
		[]string{"totp"},
	)
	s.Require().NoError(err)

	// Step 5: Verify status is now Verified
	updatedFallback, found := s.keeper.GetBorderlineFallbackRecord(s.ctx, fallback.FallbackID)
	s.Require().True(found)
	s.Require().Equal(types.BorderlineFallbackStatusCompleted, updatedFallback.Status)
	s.Require().Equal(types.VerificationStatusVerified, updatedFallback.FinalVerificationStatus)
	s.Require().Contains(updatedFallback.SatisfiedFactors, "totp")

	// Verify score was updated
	score, accStatus, found := s.keeper.GetScore(s.ctx, testBorderlineAddress)
	s.Require().True(found)
	s.Require().Equal(uint32(87), score)
	s.Require().Equal(types.AccountStatusVerified, accStatus)
}

func (s *BorderlineFallbackTestSuite) TestBorderlineFallbackFlow_MFANotSatisfied() {
	// Setup
	s.addTOTPEnrollment(testBorderlineAddress)

	// Trigger borderline fallback
	status, err := s.keeper.CheckBorderlineAndTriggerFallback(s.ctx, testBorderlineAddress, 87)
	s.Require().NoError(err)
	s.Require().Equal(types.VerificationStatusNeedsAdditionalFactor, status)

	fallbacks := s.keeper.GetPendingFallbacksForAccount(s.ctx, testBorderlineAddress)
	s.Require().Len(fallbacks, 1)
	fallback := fallbacks[0]

	// Don't mark challenge as verified - it's still pending

	// Attempt to complete fallback should fail
	err = s.keeper.HandleBorderlineFallbackCompleted(
		s.ctx,
		testBorderlineAddress,
		fallback.ChallengeID,
		[]string{"totp"},
	)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrMFAChallengeNotSatisfied)

	// Fallback should remain pending; proof failures must not mutate fallback state.
	updatedFallback, found := s.keeper.GetBorderlineFallbackRecord(s.ctx, fallback.FallbackID)
	s.Require().True(found)
	s.Require().Equal(types.BorderlineFallbackStatusPending, updatedFallback.Status)
	s.Require().Empty(updatedFallback.SatisfiedFactors)
	s.Require().Zero(updatedFallback.CompletedAt)
	pending := s.keeper.GetPendingFallbacksForAccount(s.ctx, testBorderlineAddress)
	s.Require().Len(pending, 1)
	s.Require().Equal(fallback.FallbackID, pending[0].FallbackID)
}

func (s *BorderlineFallbackTestSuite) TestBorderlineFallbackFlow_HighSecurityFactor() {
	// Setup: Add FIDO2 enrollment (high security)
	s.addFIDO2Enrollment(testBorderlineAddress)

	// Trigger borderline fallback
	status, err := s.keeper.CheckBorderlineAndTriggerFallback(s.ctx, testBorderlineAddress, 88)
	s.Require().NoError(err)
	s.Require().Equal(types.VerificationStatusNeedsAdditionalFactor, status)

	fallbacks := s.keeper.GetPendingFallbacksForAccount(s.ctx, testBorderlineAddress)
	s.Require().Len(fallbacks, 1)
	fallback := fallbacks[0]

	// Complete with FIDO2
	s.mfaKeeper.SetChallengeVerified(fallback.ChallengeID)
	err = s.keeper.HandleBorderlineFallbackCompleted(
		s.ctx,
		testBorderlineAddress,
		fallback.ChallengeID,
		[]string{"fido2"},
	)
	s.Require().NoError(err)

	// Verify completion
	updatedFallback, _ := s.keeper.GetBorderlineFallbackRecord(s.ctx, fallback.FallbackID)
	s.Require().Equal(types.BorderlineFallbackStatusCompleted, updatedFallback.Status)
}

func (s *BorderlineFallbackTestSuite) TestBorderlineFallbackCompletionRejectsNonAuthoritativeFactorsWithoutMutation() {
	testCases := []struct {
		name            string
		callerFactors   []string
		requiredFactors uint32
		mutateChallenge func(*mfatypes.Challenge)
	}{
		{
			name:          "spoofed high",
			callerFactors: []string{"fido2"},
		},
		{
			name:          "extra",
			callerFactors: []string{"totp", "fido2"},
		},
		{
			name:          "duplicate",
			callerFactors: []string{"totp", "totp"},
		},
		{
			name:          "invalid",
			callerFactors: []string{"bogus"},
		},
		{
			name:          "noncanonical",
			callerFactors: []string{"TOTP"},
		},
		{
			name:          "empty",
			callerFactors: []string{},
		},
		{
			name:          "account mismatch",
			callerFactors: []string{"totp"},
			mutateChallenge: func(challenge *mfatypes.Challenge) {
				challenge.AccountAddress = sdk.AccAddress([]byte("borderline_other____")).String()
			},
		},
		{
			name:          "expiry exact",
			callerFactors: []string{"totp"},
			mutateChallenge: func(challenge *mfatypes.Challenge) {
				challenge.ExpiresAt = s.ctx.BlockTime().Unix()
			},
		},
		{
			name:          "expiry past",
			callerFactors: []string{"totp"},
			mutateChallenge: func(challenge *mfatypes.Challenge) {
				challenge.ExpiresAt = s.ctx.BlockTime().Add(-time.Second).Unix()
			},
		},
		{
			name:          "invalid challenge factor",
			callerFactors: []string{"totp"},
			mutateChallenge: func(challenge *mfatypes.Challenge) {
				challenge.FactorType = mfatypes.FactorTypeUnspecified
			},
		},
		{
			name:            "required factors 2",
			callerFactors:   []string{"totp"},
			requiredFactors: 2,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			fallback := s.preparePremiumTOTPBorderlineFallback(tc.requiredFactors)
			s.mfaKeeper.SetChallengeVerified(fallback.ChallengeID)
			if tc.mutateChallenge != nil {
				challenge, found := s.mfaKeeper.GetChallenge(s.ctx, fallback.ChallengeID)
				s.Require().True(found)
				tc.mutateChallenge(challenge)
			}

			err := s.keeper.HandleBorderlineFallbackCompleted(
				s.ctx,
				testBorderlineAddress,
				fallback.ChallengeID,
				tc.callerFactors,
			)
			s.Require().Error(err)
			s.assertFallbackProofFailureMutationFree(fallback)
		})
	}
}

func (s *BorderlineFallbackTestSuite) TestCompleteBorderlineFallbackMsgServerUsesDerivedTOTPFactors() {
	fallback := s.preparePremiumTOTPBorderlineFallback(1)
	s.mfaKeeper.SetChallengeVerified(fallback.ChallengeID)
	s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())

	msgServer := keeper.NewMsgServerImpl(s.keeper)
	resp, err := msgServer.CompleteBorderlineFallback(s.ctx, &types.MsgCompleteBorderlineFallback{
		Sender:           testBorderlineAddress,
		ChallengeId:      fallback.ChallengeID,
		FactorsSatisfied: []string{"totp"},
	})
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Equal(fallback.FallbackID, resp.FallbackId)
	s.Require().Equal(types.VerificationStatusToProto(types.VerificationStatusVerified), resp.FinalStatus)
	s.Require().Equal("medium", resp.FactorClass)

	updated, found := s.keeper.GetBorderlineFallbackRecord(s.ctx, fallback.FallbackID)
	s.Require().True(found)
	s.Require().Equal(types.BorderlineFallbackStatusCompleted, updated.Status)
	s.Require().Equal([]string{"totp"}, updated.SatisfiedFactors)

	var completionEvent sdk.Event
	completionEventCount := 0
	for _, event := range s.ctx.EventManager().Events() {
		if event.Type == types.EventTypeBorderlineFallbackCompleted {
			completionEvent = event
			completionEventCount++
		}
	}
	s.Require().Equal(1, completionEventCount)
	s.Require().Equal(types.EventTypeBorderlineFallbackCompleted, completionEvent.Type)
	s.Require().Equal("medium", eventAttributeValue(completionEvent, types.AttributeKeyFactorClass))
	s.Require().Equal("totp", eventAttributeValue(completionEvent, types.AttributeKeySatisfiedFactors))
}

// ============================================================================
// Borderline Parameters Tests
// ============================================================================

func (s *BorderlineFallbackTestSuite) TestBorderlineParams_SetAndGet() {
	params := types.BorderlineParams{
		LowerThreshold:   80,
		UpperThreshold:   92,
		MfaTimeoutBlocks: 200, // Converted from seconds to blocks
		RequiredFactors:  2,   // Number of factors required
	}

	err := s.keeper.SetBorderlineParams(s.ctx, params)
	s.Require().NoError(err)

	retrieved := s.keeper.GetBorderlineParams(s.ctx)
	s.Require().Equal(params.LowerThreshold, retrieved.LowerThreshold)
	s.Require().Equal(params.UpperThreshold, retrieved.UpperThreshold)
	s.Require().Equal(params.MfaTimeoutBlocks, retrieved.MfaTimeoutBlocks)
	s.Require().Equal(params.RequiredFactors, retrieved.RequiredFactors)
}

func (s *BorderlineFallbackTestSuite) TestBorderlineParams_Validation() {
	// Lower threshold > upper threshold should fail
	invalidParams := types.BorderlineParams{
		LowerThreshold:   95,
		UpperThreshold:   85,
		MfaTimeoutBlocks: 100,
		RequiredFactors:  1,
	}

	err := types.ValidateBorderlineParams(invalidParams)
	s.Require().Error(err)
}

func (s *BorderlineFallbackTestSuite) TestBorderlineParams_IsScoreInBorderlineBand() {
	params := types.BorderlineParams{
		LowerThreshold: 85,
		UpperThreshold: 90,
	}

	s.Require().False(types.IsScoreInBorderlineBand(params, 84)) // Below lower
	s.Require().True(types.IsScoreInBorderlineBand(params, 85))  // At lower bound
	s.Require().True(types.IsScoreInBorderlineBand(params, 87))  // In band
	s.Require().True(types.IsScoreInBorderlineBand(params, 89))  // In band
	s.Require().False(types.IsScoreInBorderlineBand(params, 90)) // At upper bound (excluded)
	s.Require().False(types.IsScoreInBorderlineBand(params, 91)) // Above upper
}

// ============================================================================
// Fallback Cancellation Tests
// ============================================================================

func (s *BorderlineFallbackTestSuite) TestCancelBorderlineFallback() {
	// Setup
	s.addTOTPEnrollment(testBorderlineAddress)

	// Trigger fallback
	_, err := s.keeper.CheckBorderlineAndTriggerFallback(s.ctx, testBorderlineAddress, 87)
	s.Require().NoError(err)

	fallbacks := s.keeper.GetPendingFallbacksForAccount(s.ctx, testBorderlineAddress)
	s.Require().Len(fallbacks, 1)

	// Cancel the fallback
	err = s.keeper.CancelBorderlineFallback(s.ctx, testBorderlineAddress, fallbacks[0].FallbackID)
	s.Require().NoError(err)

	// Verify it's cancelled
	updated, found := s.keeper.GetBorderlineFallbackRecord(s.ctx, fallbacks[0].FallbackID)
	s.Require().True(found)
	s.Require().Equal(types.BorderlineFallbackStatusCancelled, updated.Status)
}

func (s *BorderlineFallbackTestSuite) TestCancelBorderlineFallback_WrongAccount() {
	// Setup
	s.addTOTPEnrollment(testBorderlineAddress)

	// Trigger fallback
	_, err := s.keeper.CheckBorderlineAndTriggerFallback(s.ctx, testBorderlineAddress, 87)
	s.Require().NoError(err)

	fallbacks := s.keeper.GetPendingFallbacksForAccount(s.ctx, testBorderlineAddress)
	s.Require().Len(fallbacks, 1)

	// Try to cancel with wrong account
	wrongAddress := "virtengine1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5z5wrong"
	err = s.keeper.CancelBorderlineFallback(s.ctx, wrongAddress, fallbacks[0].FallbackID)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrUnauthorized)
}

// ============================================================================
// Helper Functions
// ============================================================================

func (s *BorderlineFallbackTestSuite) preparePremiumTOTPBorderlineFallback(requiredFactors uint32) types.BorderlineFallbackRecord {
	if requiredFactors > 0 {
		params := types.DefaultBorderlineParams()
		params.RequiredFactors = requiredFactors
		s.Require().NoError(s.keeper.SetBorderlineParams(s.ctx, params))
	}

	s.addTOTPEnrollment(testBorderlineAddress)
	record := types.NewIdentityRecord(testBorderlineAddress, s.ctx.BlockTime())
	record.Tier = types.IdentityTierPremium
	s.Require().NoError(s.keeper.SetIdentityRecord(s.ctx, *record))

	status, err := s.keeper.CheckBorderlineAndTriggerFallback(s.ctx, testBorderlineAddress, 87)
	s.Require().NoError(err)
	s.Require().Equal(types.VerificationStatusNeedsAdditionalFactor, status)

	fallbacks := s.keeper.GetPendingFallbacksForAccount(s.ctx, testBorderlineAddress)
	s.Require().NotEmpty(fallbacks)
	return fallbacks[len(fallbacks)-1]
}

func (s *BorderlineFallbackTestSuite) assertFallbackProofFailureMutationFree(fallback types.BorderlineFallbackRecord) {
	updated, found := s.keeper.GetBorderlineFallbackRecord(s.ctx, fallback.FallbackID)
	s.Require().True(found)
	s.Require().Equal(types.BorderlineFallbackStatusPending, updated.Status)
	s.Require().Empty(updated.SatisfiedFactors)
	s.Require().Zero(updated.CompletedAt)
	s.Require().Equal(types.VerificationStatus(""), updated.FinalVerificationStatus)

	pending := s.keeper.GetPendingFallbacksForAccount(s.ctx, testBorderlineAddress)
	queueHasFallback := false
	for _, queued := range pending {
		if queued.FallbackID == fallback.FallbackID {
			queueHasFallback = true
			break
		}
	}
	s.Require().True(queueHasFallback, "fallback should remain in pending queue")

	_, _, scoreFound := s.keeper.GetScore(s.ctx, testBorderlineAddress)
	s.Require().False(scoreFound)

	address, err := sdk.AccAddressFromBech32(testBorderlineAddress)
	s.Require().NoError(err)
	record, found := s.keeper.GetIdentityRecord(s.ctx, address)
	s.Require().True(found)
	s.Require().Equal(types.IdentityTierPremium, record.Tier)
}

func eventAttributeValue(event sdk.Event, key string) string {
	for _, attr := range event.Attributes {
		if string(attr.Key) == key {
			return string(attr.Value)
		}
	}
	return ""
}

func (s *BorderlineFallbackTestSuite) addTOTPEnrollment(address string) {
	enrollment := mfatypes.FactorEnrollment{
		AccountAddress: address,
		FactorType:     mfatypes.FactorTypeTOTP,
		FactorID:       "totp-001",
		Label:          "Test TOTP",
		Status:         mfatypes.EnrollmentStatusActive,
		EnrolledAt:     time.Now().Unix(),
		VerifiedAt:     time.Now().Unix(),
	}
	s.mfaKeeper.AddEnrollment(address, enrollment)
}

func (s *BorderlineFallbackTestSuite) addFIDO2Enrollment(address string) {
	enrollment := mfatypes.FactorEnrollment{
		AccountAddress: address,
		FactorType:     mfatypes.FactorTypeFIDO2,
		FactorID:       "fido2-001",
		Label:          "Test FIDO2 Key",
		Status:         mfatypes.EnrollmentStatusActive,
		EnrolledAt:     time.Now().Unix(),
		VerifiedAt:     time.Now().Unix(),
	}
	s.mfaKeeper.AddEnrollment(address, enrollment)
}

//nolint:unused // helper reserved for future fallback scenarios
func (s *BorderlineFallbackTestSuite) addEmailEnrollment(address string) {
	enrollment := mfatypes.FactorEnrollment{
		AccountAddress: address,
		FactorType:     mfatypes.FactorTypeEmail,
		FactorID:       "email-001",
		Label:          "Test Email",
		Status:         mfatypes.EnrollmentStatusActive,
		EnrolledAt:     time.Now().Unix(),
		VerifiedAt:     time.Now().Unix(),
	}
	s.mfaKeeper.AddEnrollment(address, enrollment)
}
