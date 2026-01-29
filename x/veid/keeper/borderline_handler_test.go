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

// Test addresses
var (
	testHandlerAddress       = sdk.AccAddress([]byte("handler_test_addr___")).String()
	testHandlerAddress2      = sdk.AccAddress([]byte("handler_test_addr2__")).String()
	testReviewerAddress      = sdk.AccAddress([]byte("reviewer_address____")).String()
)

// BorderlineHandlerTestSuite is the test suite for borderline handler functionality
type BorderlineHandlerTestSuite struct {
	suite.Suite
	ctx       sdk.Context
	keeper    keeper.Keeper
	mfaKeeper *MockMFAKeeper
	cdc       codec.Codec
}

func TestBorderlineHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(BorderlineHandlerTestSuite))
}

func (s *BorderlineHandlerTestSuite) SetupTest() {
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

func (s *BorderlineHandlerTestSuite) createContextWithStore(storeKey *storetypes.KVStoreKey) sdk.Context {
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

// ============================================================================
// DetectBorderlineCase Tests
// ============================================================================

func (s *BorderlineHandlerTestSuite) TestDetectBorderlineCase_InBorderlineBand() {
	// Score in borderline band should be detected
	borderlineCase, detected := s.keeper.DetectBorderlineCase(
		s.ctx,
		testHandlerAddress,
		types.ScopeTypeBiometric,
		87, // In borderline band (85-90)
	)

	s.Require().True(detected)
	s.Require().NotNil(borderlineCase)
	s.Require().Equal(testHandlerAddress, borderlineCase.Address)
	s.Require().Equal(types.ScopeTypeBiometric, borderlineCase.ScopeType)
	s.Require().Equal(uint32(87), borderlineCase.Score)
	s.Require().Equal(keeper.CaseStatusPending, borderlineCase.Status)
	s.Require().NotEmpty(borderlineCase.CaseID)
}

func (s *BorderlineHandlerTestSuite) TestDetectBorderlineCase_AboveUpperThreshold() {
	// Score above upper threshold should NOT be detected as borderline
	borderlineCase, detected := s.keeper.DetectBorderlineCase(
		s.ctx,
		testHandlerAddress,
		types.ScopeTypeBiometric,
		95, // Above upper threshold (90)
	)

	s.Require().False(detected)
	s.Require().Nil(borderlineCase)
}

func (s *BorderlineHandlerTestSuite) TestDetectBorderlineCase_BelowLowerThreshold() {
	// Score below lower threshold should NOT be detected as borderline
	borderlineCase, detected := s.keeper.DetectBorderlineCase(
		s.ctx,
		testHandlerAddress,
		types.ScopeTypeBiometric,
		70, // Below lower threshold (85)
	)

	s.Require().False(detected)
	s.Require().Nil(borderlineCase)
}

func (s *BorderlineHandlerTestSuite) TestDetectBorderlineCase_AtLowerThreshold() {
	// Score exactly at lower threshold should be detected
	borderlineCase, detected := s.keeper.DetectBorderlineCase(
		s.ctx,
		testHandlerAddress,
		types.ScopeTypeBiometric,
		85, // At lower threshold
	)

	s.Require().True(detected)
	s.Require().NotNil(borderlineCase)
	s.Require().Equal(uint32(85), borderlineCase.Score)
}

func (s *BorderlineHandlerTestSuite) TestDetectBorderlineCase_AtUpperThresholdMinus1() {
	// Score at upper threshold - 1 should be detected
	borderlineCase, detected := s.keeper.DetectBorderlineCase(
		s.ctx,
		testHandlerAddress,
		types.ScopeTypeBiometric,
		89, // Just below upper threshold (90)
	)

	s.Require().True(detected)
	s.Require().NotNil(borderlineCase)
	s.Require().Equal(uint32(89), borderlineCase.Score)
}

func (s *BorderlineHandlerTestSuite) TestDetectBorderlineCase_MarginCalculation() {
	// Test margin calculation - closer to upper threshold
	borderlineCase, detected := s.keeper.DetectBorderlineCase(
		s.ctx,
		testHandlerAddress,
		types.ScopeTypeBiometric,
		89, // 1 point from upper (90), 4 points from lower (85)
	)

	s.Require().True(detected)
	s.Require().Equal(uint32(1), borderlineCase.Margin)  // Margin from nearest threshold
	s.Require().Equal(uint32(90), borderlineCase.Threshold) // Nearest is upper
}

func (s *BorderlineHandlerTestSuite) TestDetectBorderlineCase_MarginCalculation_CloserToLower() {
	// Test margin calculation - closer to lower threshold
	borderlineCase, detected := s.keeper.DetectBorderlineCase(
		s.ctx,
		testHandlerAddress,
		types.ScopeTypeBiometric,
		86, // 4 points from upper (90), 1 point from lower (85)
	)

	s.Require().True(detected)
	s.Require().Equal(uint32(1), borderlineCase.Margin)    // Margin from nearest threshold
	s.Require().Equal(uint32(85), borderlineCase.Threshold) // Nearest is lower
}

// ============================================================================
// HandleBorderlineCase Tests - Each Fallback Action
// ============================================================================

func (s *BorderlineHandlerTestSuite) TestHandleBorderlineCase_ManualReview() {
	// Create a borderline case with manual review action
	borderlineCase := &keeper.BorderlineCase{
		CaseID:         "test-case-001",
		Address:        testHandlerAddress,
		ScopeType:      types.ScopeTypeBiometric,
		Score:          88,
		Threshold:      90,
		Margin:         2,
		FallbackAction: keeper.ActionManualReview,
		Status:         keeper.CaseStatusPending,
		CreatedAt:      s.ctx.BlockTime().Unix(),
		ExpiresAt:      s.ctx.BlockTime().Unix() + 300,
		BlockHeight:    s.ctx.BlockHeight(),
	}

	err := s.keeper.HandleBorderlineCase(s.ctx, borderlineCase)
	s.Require().NoError(err)

	// Verify case is now in review
	retrievedCase, found := s.keeper.GetBorderlineCase(s.ctx, "test-case-001")
	s.Require().True(found)
	s.Require().Equal(keeper.CaseStatusInReview, retrievedCase.Status)

	// Verify it's in the manual review queue
	reviewQueue := s.keeper.GetManualReviewQueue(s.ctx)
	s.Require().Len(reviewQueue, 1)
	s.Require().Equal("test-case-001", reviewQueue[0].CaseID)
}

func (s *BorderlineHandlerTestSuite) TestHandleBorderlineCase_RequestAdditionalData() {
	borderlineCase := &keeper.BorderlineCase{
		CaseID:         "test-case-002",
		Address:        testHandlerAddress,
		ScopeType:      types.ScopeTypeBiometric,
		Score:          86,
		Threshold:      85,
		Margin:         4,
		FallbackAction: keeper.ActionRequestAdditionalData,
		Status:         keeper.CaseStatusPending,
		CreatedAt:      s.ctx.BlockTime().Unix(),
		ExpiresAt:      s.ctx.BlockTime().Unix() + 300,
		BlockHeight:    s.ctx.BlockHeight(),
	}

	err := s.keeper.HandleBorderlineCase(s.ctx, borderlineCase)
	s.Require().NoError(err)

	// Verify case status
	retrievedCase, found := s.keeper.GetBorderlineCase(s.ctx, "test-case-002")
	s.Require().True(found)
	s.Require().Equal(keeper.CaseStatusAwaitingData, retrievedCase.Status)
}

func (s *BorderlineHandlerTestSuite) TestHandleBorderlineCase_GrantProvisional() {
	borderlineCase := &keeper.BorderlineCase{
		CaseID:         "test-case-003",
		Address:        testHandlerAddress,
		ScopeType:      types.ScopeTypeBiometric,
		Score:          88,
		Threshold:      90,
		Margin:         2,
		FallbackAction: keeper.ActionGrantProvisional,
		Status:         keeper.CaseStatusPending,
		CreatedAt:      s.ctx.BlockTime().Unix(),
		ExpiresAt:      s.ctx.BlockTime().Unix() + 300,
		BlockHeight:    s.ctx.BlockHeight(),
	}

	err := s.keeper.HandleBorderlineCase(s.ctx, borderlineCase)
	s.Require().NoError(err)

	// Verify case status
	retrievedCase, found := s.keeper.GetBorderlineCase(s.ctx, "test-case-003")
	s.Require().True(found)
	s.Require().Equal(keeper.CaseStatusProvisional, retrievedCase.Status)
	s.Require().NotZero(retrievedCase.ProvisionalExpiresAt)

	// Verify provisional approval record exists
	provisional, found := s.keeper.GetProvisionalApproval(s.ctx, "test-case-003")
	s.Require().True(found)
	s.Require().Equal(testHandlerAddress, provisional.Address)
	s.Require().Equal(keeper.ProvisionalStatusActive, provisional.Status)
}

func (s *BorderlineHandlerTestSuite) TestHandleBorderlineCase_ApplyPenalty() {
	borderlineCase := &keeper.BorderlineCase{
		CaseID:         "test-case-004",
		Address:        testHandlerAddress,
		ScopeType:      types.ScopeTypeBiometric,
		Score:          87,
		Threshold:      90,
		Margin:         3,
		FallbackAction: keeper.ActionApplyPenalty,
		Status:         keeper.CaseStatusPending,
		CreatedAt:      s.ctx.BlockTime().Unix(),
		ExpiresAt:      s.ctx.BlockTime().Unix() + 300,
		BlockHeight:    s.ctx.BlockHeight(),
	}

	err := s.keeper.HandleBorderlineCase(s.ctx, borderlineCase)
	s.Require().NoError(err)

	// Verify case is resolved
	retrievedCase, found := s.keeper.GetBorderlineCase(s.ctx, "test-case-004")
	s.Require().True(found)
	s.Require().Equal(keeper.CaseStatusResolved, retrievedCase.Status)
	s.Require().Contains(retrievedCase.Resolution, "penalty applied")
}

func (s *BorderlineHandlerTestSuite) TestHandleBorderlineCase_Refer() {
	borderlineCase := &keeper.BorderlineCase{
		CaseID:         "test-case-005",
		Address:        testHandlerAddress,
		ScopeType:      types.ScopeTypeBiometric,
		Score:          86,
		Threshold:      85,
		Margin:         1,
		FallbackAction: keeper.ActionRefer,
		Status:         keeper.CaseStatusPending,
		CreatedAt:      s.ctx.BlockTime().Unix(),
		ExpiresAt:      s.ctx.BlockTime().Unix() + 300,
		BlockHeight:    s.ctx.BlockHeight(),
	}

	err := s.keeper.HandleBorderlineCase(s.ctx, borderlineCase)
	s.Require().NoError(err)

	// Verify case is referred
	retrievedCase, found := s.keeper.GetBorderlineCase(s.ctx, "test-case-005")
	s.Require().True(found)
	s.Require().Equal(keeper.CaseStatusReferred, retrievedCase.Status)
}

// ============================================================================
// SubmitForManualReview Tests
// ============================================================================

func (s *BorderlineHandlerTestSuite) TestSubmitForManualReview() {
	borderlineCase := &keeper.BorderlineCase{
		CaseID:         "manual-review-001",
		Address:        testHandlerAddress,
		ScopeType:      types.ScopeTypeBiometric,
		Score:          88,
		Threshold:      90,
		Margin:         2,
		FallbackAction: keeper.ActionManualReview,
		Status:         keeper.CaseStatusPending,
		CreatedAt:      s.ctx.BlockTime().Unix(),
		ExpiresAt:      s.ctx.BlockTime().Unix() + 300,
		BlockHeight:    s.ctx.BlockHeight(),
	}

	err := s.keeper.SubmitForManualReview(s.ctx, borderlineCase, "Borderline score needs review")
	s.Require().NoError(err)

	// Verify case status
	retrievedCase, found := s.keeper.GetBorderlineCase(s.ctx, "manual-review-001")
	s.Require().True(found)
	s.Require().Equal(keeper.CaseStatusInReview, retrievedCase.Status)
	s.Require().Equal("Borderline score needs review", retrievedCase.Notes)

	// Verify review queue
	queue := s.keeper.GetManualReviewQueue(s.ctx)
	s.Require().Len(queue, 1)
	s.Require().Equal("Borderline score needs review", queue[0].Reason)
}

func (s *BorderlineHandlerTestSuite) TestSubmitForManualReview_PriorityCalculation() {
	// Test priority 1 - margin <= 1
	case1 := &keeper.BorderlineCase{
		CaseID:      "priority-1",
		Address:     testHandlerAddress,
		Score:       89,
		Threshold:   90,
		Margin:      1,
		Status:      keeper.CaseStatusPending,
		CreatedAt:   s.ctx.BlockTime().Unix(),
		ExpiresAt:   s.ctx.BlockTime().Unix() + 300,
		BlockHeight: s.ctx.BlockHeight(),
	}

	err := s.keeper.SubmitForManualReview(s.ctx, case1, "Test priority 1")
	s.Require().NoError(err)

	// Test priority 2 - margin <= 3
	case2 := &keeper.BorderlineCase{
		CaseID:      "priority-2",
		Address:     testHandlerAddress2,
		Score:       87,
		Threshold:   90,
		Margin:      3,
		Status:      keeper.CaseStatusPending,
		CreatedAt:   s.ctx.BlockTime().Unix(),
		ExpiresAt:   s.ctx.BlockTime().Unix() + 300,
		BlockHeight: s.ctx.BlockHeight(),
	}

	err = s.keeper.SubmitForManualReview(s.ctx, case2, "Test priority 2")
	s.Require().NoError(err)

	queue := s.keeper.GetManualReviewQueue(s.ctx)
	s.Require().Len(queue, 2)

	// Queue should be ordered by priority (lower number = higher priority)
	s.Require().Equal(1, queue[0].Priority)
	s.Require().Equal(2, queue[1].Priority)
}

// ============================================================================
// GrantProvisionalApproval Tests
// ============================================================================

func (s *BorderlineHandlerTestSuite) TestGrantProvisionalApproval() {
	borderlineCase := &keeper.BorderlineCase{
		CaseID:      "provisional-001",
		Address:     testHandlerAddress,
		Score:       88,
		Threshold:   90,
		Margin:      2,
		Status:      keeper.CaseStatusPending,
		CreatedAt:   s.ctx.BlockTime().Unix(),
		ExpiresAt:   s.ctx.BlockTime().Unix() + 300,
		BlockHeight: s.ctx.BlockHeight(),
	}

	duration := 24 * time.Hour
	err := s.keeper.GrantProvisionalApproval(s.ctx, borderlineCase, duration)
	s.Require().NoError(err)

	// Verify provisional approval
	provisional, found := s.keeper.GetProvisionalApproval(s.ctx, "provisional-001")
	s.Require().True(found)
	s.Require().Equal(testHandlerAddress, provisional.Address)
	s.Require().Equal(keeper.ProvisionalStatusActive, provisional.Status)
	s.Require().Len(provisional.Conditions, 1)
	s.Require().Len(provisional.RequiredActions, 2)
	s.Require().Equal(uint32(88), provisional.TemporaryScore)
	s.Require().Equal(uint32(88), provisional.OriginalScore)

	// Verify expiry is approximately 24 hours from now
	expectedExpiry := s.ctx.BlockTime().Add(duration).Unix()
	s.Require().Equal(expectedExpiry, provisional.ExpiresAt)
}

func (s *BorderlineHandlerTestSuite) TestProvisionalApproval_Expiration() {
	borderlineCase := &keeper.BorderlineCase{
		CaseID:      "provisional-expire-001",
		Address:     testHandlerAddress,
		Score:       88,
		Threshold:   90,
		Margin:      2,
		Status:      keeper.CaseStatusPending,
		CreatedAt:   s.ctx.BlockTime().Unix(),
		ExpiresAt:   s.ctx.BlockTime().Unix() + 300,
		BlockHeight: s.ctx.BlockHeight(),
	}

	// Grant with short duration for testing
	duration := 1 * time.Second
	err := s.keeper.GrantProvisionalApproval(s.ctx, borderlineCase, duration)
	s.Require().NoError(err)

	// Verify it's active
	provisional, found := s.keeper.GetProvisionalApproval(s.ctx, "provisional-expire-001")
	s.Require().True(found)
	s.Require().Equal(keeper.ProvisionalStatusActive, provisional.Status)

	// Advance time past expiration
	futureCtx := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(2 * time.Second))

	// Process expired provisional approvals
	expiredCount := s.keeper.ProcessExpiredProvisionalApprovals(futureCtx)
	s.Require().Equal(1, expiredCount)

	// Verify it's now expired
	provisional, found = s.keeper.GetProvisionalApproval(futureCtx, "provisional-expire-001")
	s.Require().True(found)
	s.Require().Equal(keeper.ProvisionalStatusExpired, provisional.Status)
}

// ============================================================================
// GetBorderlineCases Tests
// ============================================================================

func (s *BorderlineHandlerTestSuite) TestGetBorderlineCases() {
	// Create multiple cases
	cases := []*keeper.BorderlineCase{
		{
			CaseID:      "list-case-001",
			Address:     testHandlerAddress,
			Score:       87,
			Status:      keeper.CaseStatusPending,
			CreatedAt:   s.ctx.BlockTime().Unix(),
			ExpiresAt:   s.ctx.BlockTime().Unix() + 300,
			BlockHeight: s.ctx.BlockHeight(),
		},
		{
			CaseID:      "list-case-002",
			Address:     testHandlerAddress2,
			Score:       88,
			Status:      keeper.CaseStatusPending,
			CreatedAt:   s.ctx.BlockTime().Unix(),
			ExpiresAt:   s.ctx.BlockTime().Unix() + 300,
			BlockHeight: s.ctx.BlockHeight(),
		},
	}

	for _, c := range cases {
		err := s.keeper.SubmitForManualReview(s.ctx, c, "Test case")
		s.Require().NoError(err)
	}

	// Get all cases
	allCases := s.keeper.GetBorderlineCases(s.ctx)
	s.Require().Len(allCases, 2)
}

func (s *BorderlineHandlerTestSuite) TestGetPendingBorderlineCases() {
	// Create cases with different statuses
	pendingCase := &keeper.BorderlineCase{
		CaseID:      "pending-filter-001",
		Address:     testHandlerAddress,
		Score:       87,
		Status:      keeper.CaseStatusPending,
		CreatedAt:   s.ctx.BlockTime().Unix(),
		ExpiresAt:   s.ctx.BlockTime().Unix() + 300,
		BlockHeight: s.ctx.BlockHeight(),
	}

	resolvedCase := &keeper.BorderlineCase{
		CaseID:      "pending-filter-002",
		Address:     testHandlerAddress2,
		Score:       88,
		Status:      keeper.CaseStatusResolved,
		CreatedAt:   s.ctx.BlockTime().Unix(),
		ExpiresAt:   s.ctx.BlockTime().Unix() + 300,
		BlockHeight: s.ctx.BlockHeight(),
	}

	// Store both cases
	err := s.keeper.SubmitForManualReview(s.ctx, pendingCase, "Pending case")
	s.Require().NoError(err)

	// Manually store resolved case
	err = s.keeper.ResolveBorderlineCase(s.ctx, "pending-filter-001", "Test resolution", testReviewerAddress, true, nil)
	s.Require().NoError(err)

	err = s.keeper.SubmitForManualReview(s.ctx, resolvedCase, "Will be resolved")
	s.Require().NoError(err)

	// Get pending cases
	pendingCases := s.keeper.GetPendingBorderlineCases(s.ctx)
	s.Require().Len(pendingCases, 1)
	s.Require().Equal("pending-filter-002", pendingCases[0].CaseID)
}

func (s *BorderlineHandlerTestSuite) TestGetBorderlineCasesForAccount() {
	// Create cases for different accounts
	cases := []*keeper.BorderlineCase{
		{
			CaseID:      "account-filter-001",
			Address:     testHandlerAddress,
			Score:       87,
			Status:      keeper.CaseStatusPending,
			CreatedAt:   s.ctx.BlockTime().Unix(),
			ExpiresAt:   s.ctx.BlockTime().Unix() + 300,
			BlockHeight: s.ctx.BlockHeight(),
		},
		{
			CaseID:      "account-filter-002",
			Address:     testHandlerAddress,
			Score:       88,
			Status:      keeper.CaseStatusPending,
			CreatedAt:   s.ctx.BlockTime().Unix(),
			ExpiresAt:   s.ctx.BlockTime().Unix() + 300,
			BlockHeight: s.ctx.BlockHeight(),
		},
		{
			CaseID:      "account-filter-003",
			Address:     testHandlerAddress2,
			Score:       86,
			Status:      keeper.CaseStatusPending,
			CreatedAt:   s.ctx.BlockTime().Unix(),
			ExpiresAt:   s.ctx.BlockTime().Unix() + 300,
			BlockHeight: s.ctx.BlockHeight(),
		},
	}

	for _, c := range cases {
		err := s.keeper.SubmitForManualReview(s.ctx, c, "Test case")
		s.Require().NoError(err)
	}

	// Get cases for first account
	account1Cases := s.keeper.GetBorderlineCasesForAccount(s.ctx, testHandlerAddress)
	s.Require().Len(account1Cases, 2)

	// Get cases for second account
	account2Cases := s.keeper.GetBorderlineCasesForAccount(s.ctx, testHandlerAddress2)
	s.Require().Len(account2Cases, 1)
}

// ============================================================================
// ResolveBorderlineCase Tests
// ============================================================================

func (s *BorderlineHandlerTestSuite) TestResolveBorderlineCase_Approved() {
	borderlineCase := &keeper.BorderlineCase{
		CaseID:      "resolve-approved-001",
		Address:     testHandlerAddress,
		Score:       88,
		Threshold:   90,
		Margin:      2,
		Status:      keeper.CaseStatusPending,
		CreatedAt:   s.ctx.BlockTime().Unix(),
		ExpiresAt:   s.ctx.BlockTime().Unix() + 300,
		BlockHeight: s.ctx.BlockHeight(),
	}

	err := s.keeper.SubmitForManualReview(s.ctx, borderlineCase, "Needs review")
	s.Require().NoError(err)

	// Resolve as approved
	err = s.keeper.ResolveBorderlineCase(
		s.ctx,
		"resolve-approved-001",
		"Approved after manual review",
		testReviewerAddress,
		true,
		nil,
	)
	s.Require().NoError(err)

	// Verify case is resolved
	resolved, found := s.keeper.GetBorderlineCase(s.ctx, "resolve-approved-001")
	s.Require().True(found)
	s.Require().Equal(keeper.CaseStatusResolved, resolved.Status)
	s.Require().Equal("Approved after manual review", resolved.Resolution)
	s.Require().Equal(testReviewerAddress, resolved.ReviewerAddress)
	s.Require().NotZero(resolved.ResolvedAt)
}

func (s *BorderlineHandlerTestSuite) TestResolveBorderlineCase_Rejected() {
	borderlineCase := &keeper.BorderlineCase{
		CaseID:      "resolve-rejected-001",
		Address:     testHandlerAddress,
		Score:       86,
		Threshold:   85,
		Margin:      1,
		Status:      keeper.CaseStatusPending,
		CreatedAt:   s.ctx.BlockTime().Unix(),
		ExpiresAt:   s.ctx.BlockTime().Unix() + 300,
		BlockHeight: s.ctx.BlockHeight(),
	}

	err := s.keeper.SubmitForManualReview(s.ctx, borderlineCase, "Needs review")
	s.Require().NoError(err)

	// Resolve as rejected
	err = s.keeper.ResolveBorderlineCase(
		s.ctx,
		"resolve-rejected-001",
		"Rejected after manual review - suspicious patterns",
		testReviewerAddress,
		false,
		nil,
	)
	s.Require().NoError(err)

	// Verify case is resolved
	resolved, found := s.keeper.GetBorderlineCase(s.ctx, "resolve-rejected-001")
	s.Require().True(found)
	s.Require().Equal(keeper.CaseStatusResolved, resolved.Status)
	s.Require().Contains(resolved.Resolution, "Rejected")
}

func (s *BorderlineHandlerTestSuite) TestResolveBorderlineCase_WithScoreOverride() {
	borderlineCase := &keeper.BorderlineCase{
		CaseID:      "resolve-override-001",
		Address:     testHandlerAddress,
		Score:       88,
		Threshold:   90,
		Margin:      2,
		Status:      keeper.CaseStatusPending,
		CreatedAt:   s.ctx.BlockTime().Unix(),
		ExpiresAt:   s.ctx.BlockTime().Unix() + 300,
		BlockHeight: s.ctx.BlockHeight(),
	}

	err := s.keeper.SubmitForManualReview(s.ctx, borderlineCase, "Needs review")
	s.Require().NoError(err)

	// Resolve with score override
	overrideScore := uint32(92)
	err = s.keeper.ResolveBorderlineCase(
		s.ctx,
		"resolve-override-001",
		"Approved with score adjustment based on additional evidence",
		testReviewerAddress,
		true,
		&overrideScore,
	)
	s.Require().NoError(err)

	// Verify case is resolved
	resolved, found := s.keeper.GetBorderlineCase(s.ctx, "resolve-override-001")
	s.Require().True(found)
	s.Require().Equal(keeper.CaseStatusResolved, resolved.Status)
}

func (s *BorderlineHandlerTestSuite) TestResolveBorderlineCase_NotFound() {
	err := s.keeper.ResolveBorderlineCase(
		s.ctx,
		"nonexistent-case",
		"Test resolution",
		testReviewerAddress,
		true,
		nil,
	)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrBorderlineFallbackNotFound)
}

func (s *BorderlineHandlerTestSuite) TestResolveBorderlineCase_AlreadyResolved() {
	borderlineCase := &keeper.BorderlineCase{
		CaseID:      "already-resolved-001",
		Address:     testHandlerAddress,
		Score:       88,
		Status:      keeper.CaseStatusPending,
		CreatedAt:   s.ctx.BlockTime().Unix(),
		ExpiresAt:   s.ctx.BlockTime().Unix() + 300,
		BlockHeight: s.ctx.BlockHeight(),
	}

	err := s.keeper.SubmitForManualReview(s.ctx, borderlineCase, "Needs review")
	s.Require().NoError(err)

	// Resolve once
	err = s.keeper.ResolveBorderlineCase(s.ctx, "already-resolved-001", "First resolution", testReviewerAddress, true, nil)
	s.Require().NoError(err)

	// Try to resolve again
	err = s.keeper.ResolveBorderlineCase(s.ctx, "already-resolved-001", "Second resolution", testReviewerAddress, false, nil)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrBorderlineFallbackAlreadyCompleted)
}

// ============================================================================
// BorderlineAction String Tests
// ============================================================================

func (s *BorderlineHandlerTestSuite) TestBorderlineAction_String() {
	testCases := []struct {
		action   keeper.BorderlineAction
		expected string
	}{
		{keeper.ActionManualReview, "manual_review"},
		{keeper.ActionRequestAdditionalData, "request_additional_data"},
		{keeper.ActionApplyPenalty, "apply_penalty"},
		{keeper.ActionGrantProvisional, "grant_provisional"},
		{keeper.ActionRefer, "refer"},
		{keeper.BorderlineAction(99), "unknown"}, // Unknown action
	}

	for _, tc := range testCases {
		s.Require().Equal(tc.expected, tc.action.String())
	}
}

// ============================================================================
// Integration Test - Full Workflow
// ============================================================================

func (s *BorderlineHandlerTestSuite) TestBorderlineHandler_FullWorkflow() {
	// Setup MFA enrollment
	s.addTOTPEnrollment(testHandlerAddress)

	// Step 1: Detect borderline case from ML score
	borderlineCase, detected := s.keeper.DetectBorderlineCase(
		s.ctx,
		testHandlerAddress,
		types.ScopeTypeBiometric,
		88,
	)
	s.Require().True(detected)
	s.Require().NotNil(borderlineCase)

	// Step 2: Handle the borderline case (submit for manual review)
	borderlineCase.FallbackAction = keeper.ActionManualReview
	err := s.keeper.HandleBorderlineCase(s.ctx, borderlineCase)
	s.Require().NoError(err)

	// Step 3: Verify it's in the review queue
	queue := s.keeper.GetManualReviewQueue(s.ctx)
	s.Require().Len(queue, 1)

	// Step 4: Reviewer resolves the case
	err = s.keeper.ResolveBorderlineCase(
		s.ctx,
		borderlineCase.CaseID,
		"Approved after reviewing additional context",
		testReviewerAddress,
		true,
		nil,
	)
	s.Require().NoError(err)

	// Step 5: Verify final state
	finalCase, found := s.keeper.GetBorderlineCase(s.ctx, borderlineCase.CaseID)
	s.Require().True(found)
	s.Require().Equal(keeper.CaseStatusResolved, finalCase.Status)
	s.Require().Equal(testReviewerAddress, finalCase.ReviewerAddress)
}

// ============================================================================
// Helper Functions
// ============================================================================

func (s *BorderlineHandlerTestSuite) addTOTPEnrollment(address string) {
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
