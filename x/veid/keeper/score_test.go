package keeper_test

import (
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/veid/keeper"
	"github.com/virtengine/virtengine/x/veid/types"
)

// Test address constants
const (
	testScoreAddress1 = "virtengine1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5z5tpwx"
	testScoreAddress2 = "virtengine1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5z5tpwy"
	testScoreAddress3 = "virtengine1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5z5tpwz"
)

type ScoreKeeperTestSuite struct {
	suite.Suite
	ctx    sdk.Context
	keeper keeper.Keeper
	cdc    codec.Codec
}

func TestScoreKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(ScoreKeeperTestSuite))
}

func (s *ScoreKeeperTestSuite) SetupTest() {
	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	s.cdc = codec.NewProtoCodec(interfaceRegistry)

	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create context with store
	s.ctx = s.createContextWithStore(storeKey)

	// Create keeper
	s.keeper = keeper.NewKeeper(s.cdc, storeKey, "authority")

	// Set default params
	err := s.keeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)
}

func (s *ScoreKeeperTestSuite) createContextWithStore(storeKey *storetypes.KVStoreKey) sdk.Context {
	db := runtime.NewKVStoreService(storeKey)
	_ = db
	ctx := sdk.Context{}.WithBlockTime(time.Now()).WithBlockHeight(100)
	return ctx
}

// ============================================================================
// Score Set/Get Roundtrip Tests
// ============================================================================

func (s *ScoreKeeperTestSuite) TestSetScoreBasic() {
	// Set a score for an account
	err := s.keeper.SetScore(s.ctx, testScoreAddress1, 75, "v1.0.0")
	s.Require().NoError(err)

	// Get the score back
	score, status, found := s.keeper.GetScore(s.ctx, testScoreAddress1)
	s.Require().True(found)
	s.Require().Equal(uint32(75), score)
	s.Require().Equal(types.AccountStatusVerified, status)
}

func (s *ScoreKeeperTestSuite) TestSetScoreWithDetails() {
	expiresAt := time.Now().Add(365 * 24 * time.Hour)
	verificationHash := []byte("test-verification-hash")

	err := s.keeper.SetScoreWithDetails(s.ctx, testScoreAddress1, 85, keeper.ScoreDetails{
		Status:           types.AccountStatusVerified,
		ModelVersion:     "v2.0.0",
		VerificationHash: verificationHash,
		ExpiresAt:        &expiresAt,
		Reason:           "initial score computation",
	})
	s.Require().NoError(err)

	// Get full identity score
	identityScore, found := s.keeper.GetIdentityScore(s.ctx, testScoreAddress1)
	s.Require().True(found)
	s.Require().Equal(uint32(85), identityScore.Score)
	s.Require().Equal(types.AccountStatusVerified, identityScore.Status)
	s.Require().Equal("v2.0.0", identityScore.ModelVersion)
	s.Require().NotNil(identityScore.ExpiresAt)
}

func (s *ScoreKeeperTestSuite) TestSetScoreExceedsMaximum() {
	// Try to set a score above maximum
	err := s.keeper.SetScore(s.ctx, testScoreAddress1, 101, "v1.0.0")
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "exceeds maximum")
}

func (s *ScoreKeeperTestSuite) TestSetScoreInvalidAddress() {
	// Try to set a score with invalid address
	err := s.keeper.SetScore(s.ctx, "invalid-address", 50, "v1.0.0")
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "invalid address")
}

func (s *ScoreKeeperTestSuite) TestGetScoreNotFound() {
	// Try to get score for non-existent account
	score, status, found := s.keeper.GetScore(s.ctx, testScoreAddress1)
	s.Require().False(found)
	s.Require().Equal(uint32(0), score)
	s.Require().Equal(types.AccountStatusUnknown, status)
}

// ============================================================================
// Score History Tests
// ============================================================================

func (s *ScoreKeeperTestSuite) TestScoreHistoryAccumulation() {
	// Set multiple scores for the same account
	err := s.keeper.SetScoreWithDetails(s.ctx, testScoreAddress1, 50, keeper.ScoreDetails{
		Status: types.AccountStatusVerified, ModelVersion: "v1.0.0", Reason: "first score",
	})
	s.Require().NoError(err)

	// Advance block time
	s.ctx = s.ctx.WithBlockTime(s.ctx.BlockTime().Add(time.Hour)).WithBlockHeight(101)

	err = s.keeper.SetScoreWithDetails(s.ctx, testScoreAddress1, 65, keeper.ScoreDetails{
		Status: types.AccountStatusVerified, ModelVersion: "v1.0.0", Reason: "second score",
	})
	s.Require().NoError(err)

	// Advance block time again
	s.ctx = s.ctx.WithBlockTime(s.ctx.BlockTime().Add(time.Hour)).WithBlockHeight(102)

	err = s.keeper.SetScoreWithDetails(s.ctx, testScoreAddress1, 80, keeper.ScoreDetails{
		Status: types.AccountStatusVerified, ModelVersion: "v1.1.0", Reason: "third score",
	})
	s.Require().NoError(err)

	// Get score history
	history := s.keeper.GetScoreHistory(s.ctx, testScoreAddress1)
	s.Require().Len(history, 3)

	// Verify newest first ordering (reverse chronological)
	s.Require().Equal(uint32(80), history[0].Score)
	s.Require().Equal("third score", history[0].Reason)
	s.Require().Equal(uint32(65), history[1].Score)
	s.Require().Equal(uint32(50), history[2].Score)
}

func (s *ScoreKeeperTestSuite) TestScoreHistoryPaginated() {
	// Set multiple scores
	for i := 0; i < 10; i++ {
		s.ctx = s.ctx.WithBlockTime(s.ctx.BlockTime().Add(time.Minute)).WithBlockHeight(int64(100 + i))
		err := s.keeper.SetScoreWithDetails(s.ctx, testScoreAddress1, uint32(50+i*5), keeper.ScoreDetails{
			Status: types.AccountStatusVerified, ModelVersion: "v1.0.0",
		})
		s.Require().NoError(err)
	}

	// Get first page
	page1 := s.keeper.GetScoreHistoryPaginated(s.ctx, testScoreAddress1, 5, 0)
	s.Require().Len(page1, 5)


	// Get second page
	page2 := s.keeper.GetScoreHistoryPaginated(s.ctx, testScoreAddress1, 5, 5)
	s.Require().Len(page2, 5)

	// Verify no overlap
	for _, p1Entry := range page1 {
		for _, p2Entry := range page2 {
			s.Require().NotEqual(p1Entry.BlockHeight, p2Entry.BlockHeight)
		}
	}
}

// ============================================================================
// Tier Calculation Tests
// ============================================================================

func (s *ScoreKeeperTestSuite) TestTierCalculation() {
	testCases := []struct {
		score        uint32
		status       types.AccountStatus
		expectedTier int
	}{
		{0, types.AccountStatusVerified, types.TierUnverified},
		{49, types.AccountStatusVerified, types.TierUnverified},
		{50, types.AccountStatusVerified, types.TierBasic},
		{69, types.AccountStatusVerified, types.TierBasic},
		{70, types.AccountStatusVerified, types.TierStandard},
		{84, types.AccountStatusVerified, types.TierStandard},
		{85, types.AccountStatusVerified, types.TierPremium},
		{100, types.AccountStatusVerified, types.TierPremium},
		// Non-verified accounts are always tier 0
		{100, types.AccountStatusPending, types.TierUnverified},
		{100, types.AccountStatusRejected, types.TierUnverified},
		{100, types.AccountStatusExpired, types.TierUnverified},
	}

	for _, tc := range testCases {
		tier := types.ComputeTierFromScoreValue(tc.score, tc.status)
		s.Require().Equal(tc.expectedTier, tier, "Score: %d, Status: %s", tc.score, tc.status)
	}
}

func (s *ScoreKeeperTestSuite) TestGetAccountTier() {
	// Set score for account
	err := s.keeper.SetScoreWithDetails(s.ctx, testScoreAddress1, 75, keeper.ScoreDetails{
		Status: types.AccountStatusVerified, ModelVersion: "v1.0.0",
	})
	s.Require().NoError(err)

	tier, err := s.keeper.GetAccountTier(s.ctx, testScoreAddress1)
	s.Require().NoError(err)
	s.Require().Equal(types.TierStandard, tier)
}

func (s *ScoreKeeperTestSuite) TestGetAccountTierNotFound() {
	tier, err := s.keeper.GetAccountTier(s.ctx, testScoreAddress1)
	s.Require().Error(err)
	s.Require().Equal(types.TierUnverified, tier)
}

// ============================================================================
// Threshold Check Tests
// ============================================================================

func (s *ScoreKeeperTestSuite) TestIsScoreAboveThreshold() {
	// Set a verified account with score 75
	err := s.keeper.SetScoreWithDetails(s.ctx, testScoreAddress1, 75, keeper.ScoreDetails{
		Status: types.AccountStatusVerified, ModelVersion: "v1.0.0",
	})
	s.Require().NoError(err)

	// Test various thresholds
	s.Require().True(s.keeper.IsScoreAboveThreshold(s.ctx, testScoreAddress1, 50))
	s.Require().True(s.keeper.IsScoreAboveThreshold(s.ctx, testScoreAddress1, 70))
	s.Require().True(s.keeper.IsScoreAboveThreshold(s.ctx, testScoreAddress1, 75))
	s.Require().False(s.keeper.IsScoreAboveThreshold(s.ctx, testScoreAddress1, 76))
	s.Require().False(s.keeper.IsScoreAboveThreshold(s.ctx, testScoreAddress1, 85))
}

func (s *ScoreKeeperTestSuite) TestIsScoreAboveThresholdUnverified() {
	// Set an unverified account with high score
	err := s.keeper.SetScoreWithDetails(s.ctx, testScoreAddress1, 90, keeper.ScoreDetails{
		Status: types.AccountStatusPending, ModelVersion: "v1.0.0",
	})
	s.Require().NoError(err)

	// Should fail all threshold checks because not verified
	s.Require().False(s.keeper.IsScoreAboveThreshold(s.ctx, testScoreAddress1, 0))
	s.Require().False(s.keeper.IsScoreAboveThreshold(s.ctx, testScoreAddress1, 50))
}

func (s *ScoreKeeperTestSuite) TestIsScoreAboveThresholdNotFound() {
	// Check threshold for non-existent account
	s.Require().False(s.keeper.IsScoreAboveThreshold(s.ctx, testScoreAddress1, 0))
}

// ============================================================================
// Eligibility Check Tests
// ============================================================================

func (s *ScoreKeeperTestSuite) TestCheckEligibilityBasic() {
	// Set up a verified account with basic tier score
	err := s.keeper.SetScoreWithDetails(s.ctx, testScoreAddress1, 55, keeper.ScoreDetails{
		Status: types.AccountStatusVerified, ModelVersion: "v1.0.0",
	})
	s.Require().NoError(err)

	// Check eligibility for basic offerings
	result := s.keeper.CheckEligibility(s.ctx, testScoreAddress1, types.OfferingTypeBasic)
	s.Require().Equal(uint32(55), result.CurrentScore)
	s.Require().Equal(types.AccountStatusVerified, result.CurrentStatus)
	s.Require().Equal(uint32(50), result.RequiredScore)
}

func (s *ScoreKeeperTestSuite) TestCheckEligibilityInsufficientScore() {
	// Set up a verified account with low score
	err := s.keeper.SetScoreWithDetails(s.ctx, testScoreAddress1, 45, keeper.ScoreDetails{
		Status: types.AccountStatusVerified, ModelVersion: "v1.0.0",
	})
	s.Require().NoError(err)

	// Check eligibility for basic offerings (requires 50)
	result := s.keeper.CheckEligibility(s.ctx, testScoreAddress1, types.OfferingTypeBasic)
	s.Require().False(result.Eligible)
	s.Require().Contains(result.Reason, "minimum requirement")
}

func (s *ScoreKeeperTestSuite) TestCheckEligibilityNotVerified() {
	// Set up an unverified account with high score
	err := s.keeper.SetScoreWithDetails(s.ctx, testScoreAddress1, 95, keeper.ScoreDetails{
		Status: types.AccountStatusPending, ModelVersion: "v1.0.0",
	})
	s.Require().NoError(err)

	// Check eligibility - should fail because not verified
	result := s.keeper.CheckEligibility(s.ctx, testScoreAddress1, types.OfferingTypeBasic)
	s.Require().False(result.Eligible)
	s.Require().Contains(result.Reason, "not verified")
}

func (s *ScoreKeeperTestSuite) TestCheckEligibilityAccountNotFound() {
	// Check eligibility for non-existent account
	result := s.keeper.CheckEligibility(s.ctx, testScoreAddress1, types.OfferingTypeBasic)
	s.Require().False(result.Eligible)
	s.Require().Contains(result.Reason, "not found")
}

// ============================================================================
// Required Scopes Tests
// ============================================================================

func (s *ScoreKeeperTestSuite) TestGetRequiredScopesForOffering() {
	testCases := []struct {
		offeringType   types.OfferingType
		expectedScore  uint32
		expectedMFA    bool
	}{
		{types.OfferingTypeBasic, 50, false},
		{types.OfferingTypeStandard, 70, false},
		{types.OfferingTypePremium, 85, true},
		{types.OfferingTypeProvider, 70, true},
		{types.OfferingTypeValidator, 85, true},
	}

	for _, tc := range testCases {
		requirements := types.GetRequiredScopesForOffering(tc.offeringType)
		s.Require().Equal(tc.expectedScore, requirements.MinimumScore, "OfferingType: %s", tc.offeringType)
		s.Require().Equal(tc.expectedMFA, requirements.RequiresMFA, "OfferingType: %s", tc.offeringType)
	}
}

// ============================================================================
// Score Status Tests
// ============================================================================

func (s *ScoreKeeperTestSuite) TestGetVerificationStatus() {
	// Set score with pending status
	err := s.keeper.SetScoreWithDetails(s.ctx, testScoreAddress1, 60, keeper.ScoreDetails{
		Status: types.AccountStatusPending, ModelVersion: "v1.0.0",
	})
	s.Require().NoError(err)

	status := s.keeper.GetVerificationStatus(s.ctx, testScoreAddress1)
	s.Require().Equal(types.AccountStatusPending, status)
}

func (s *ScoreKeeperTestSuite) TestGetVerificationStatusNotFound() {
	status := s.keeper.GetVerificationStatus(s.ctx, testScoreAddress1)
	s.Require().Equal(types.AccountStatusUnknown, status)
}

// ============================================================================
// Tier-based Account Retrieval Tests
// ============================================================================

func (s *ScoreKeeperTestSuite) TestGetAccountsByTier() {
	// Create identity records first, then set scores
	address1, _ := sdk.AccAddressFromBech32(testScoreAddress1)
	address2, _ := sdk.AccAddressFromBech32(testScoreAddress2)
	address3, _ := sdk.AccAddressFromBech32(testScoreAddress3)

	_, _ = s.keeper.CreateIdentityRecord(s.ctx, address1)
	_, _ = s.keeper.CreateIdentityRecord(s.ctx, address2)
	_, _ = s.keeper.CreateIdentityRecord(s.ctx, address3)

	// Set different scores for each
	_ = s.keeper.SetScoreWithDetails(s.ctx, testScoreAddress1, 55, keeper.ScoreDetails{
		Status: types.AccountStatusVerified, ModelVersion: "v1.0.0",
	})
	_ = s.keeper.SetScoreWithDetails(s.ctx, testScoreAddress2, 75, keeper.ScoreDetails{
		Status: types.AccountStatusVerified, ModelVersion: "v1.0.0",
	})
	_ = s.keeper.SetScoreWithDetails(s.ctx, testScoreAddress3, 90, keeper.ScoreDetails{
		Status: types.AccountStatusVerified, ModelVersion: "v1.0.0",
	})

	// Get Basic tier accounts
	basicAccounts := s.keeper.GetAccountsByTier(s.ctx, types.TierBasic)
	s.Require().Equal(1, len(basicAccounts))

	// Get Standard tier accounts
	standardAccounts := s.keeper.GetAccountsByTier(s.ctx, types.TierStandard)
	s.Require().Equal(1, len(standardAccounts))

	// Get Premium tier accounts
	premiumAccounts := s.keeper.GetAccountsByTier(s.ctx, types.TierPremium)
	s.Require().Equal(1, len(premiumAccounts))
}

// ============================================================================
// Score Expiration Tests
// ============================================================================

func (s *ScoreKeeperTestSuite) TestExpireScore() {
	// Set initial score
	err := s.keeper.SetScoreWithDetails(s.ctx, testScoreAddress1, 80, keeper.ScoreDetails{
		Status: types.AccountStatusVerified, ModelVersion: "v1.0.0",
	})
	s.Require().NoError(err)

	// Expire the score
	err = s.keeper.ExpireScore(s.ctx, testScoreAddress1, "expired due to inactivity")
	s.Require().NoError(err)

	// Verify score is expired
	_, status, found := s.keeper.GetScore(s.ctx, testScoreAddress1)
	s.Require().True(found)
	s.Require().Equal(types.AccountStatusExpired, status)
}

func (s *ScoreKeeperTestSuite) TestExpireScoreNotFound() {
	err := s.keeper.ExpireScore(s.ctx, testScoreAddress1, "test")
	s.Require().Error(err)
}

func (s *ScoreKeeperTestSuite) TestRefreshScoreExpiration() {
	// Set initial score with expiration
	initialExpiry := time.Now().Add(24 * time.Hour)
	err := s.keeper.SetScoreWithDetails(s.ctx, testScoreAddress1, 80, keeper.ScoreDetails{
		Status: types.AccountStatusVerified, ModelVersion: "v1.0.0", ExpiresAt: &initialExpiry,
	})
	s.Require().NoError(err)

	// Refresh expiration
	newExpiry := time.Now().Add(365 * 24 * time.Hour)
	err = s.keeper.RefreshScoreExpiration(s.ctx, testScoreAddress1, newExpiry)
	s.Require().NoError(err)

	// Verify new expiration
	identityScore, found := s.keeper.GetIdentityScore(s.ctx, testScoreAddress1)
	s.Require().True(found)
	s.Require().NotNil(identityScore.ExpiresAt)
	s.Require().True(identityScore.ExpiresAt.After(initialExpiry))
}

// ============================================================================
// Score Statistics Tests
// ============================================================================

func (s *ScoreKeeperTestSuite) TestGetScoreStatistics() {
	// Create accounts with various scores
	addresses := []string{testScoreAddress1, testScoreAddress2, testScoreAddress3}
	scores := []uint32{55, 75, 90}

	for i, addr := range addresses {
		address, _ := sdk.AccAddressFromBech32(addr)
		_, _ = s.keeper.CreateIdentityRecord(s.ctx, address)
		_ = s.keeper.SetScoreWithDetails(s.ctx, addr, scores[i], keeper.ScoreDetails{
			Status: types.AccountStatusVerified, ModelVersion: "v1.0.0",
		})
	}

	stats := s.keeper.GetScoreStatistics(s.ctx)

	s.Require().Equal(uint64(3), stats.TotalAccounts)
	s.Require().Equal(uint64(3), stats.StatusCounts[types.AccountStatusVerified])
}

// ============================================================================
// Interface Implementation Tests
// ============================================================================

func (s *ScoreKeeperTestSuite) TestIdentityScoreKeeperInterface() {
	// Verify Keeper implements IdentityScoreKeeper interface
	var _ keeper.IdentityScoreKeeper = s.keeper

	// Test interface methods
	err := s.keeper.SetScore(s.ctx, testScoreAddress1, 75, "v1.0.0")
	s.Require().NoError(err)

	// GetScore via interface
	score, status, found := s.keeper.GetScore(s.ctx, testScoreAddress1)
	s.Require().True(found)
	s.Require().Equal(uint32(75), score)
	s.Require().Equal(types.AccountStatusVerified, status)

	// IsScoreAboveThreshold via interface
	s.Require().True(s.keeper.IsScoreAboveThreshold(s.ctx, testScoreAddress1, 70))
	s.Require().False(s.keeper.IsScoreAboveThreshold(s.ctx, testScoreAddress1, 80))

	// GetAccountTier via interface
	tier, err := s.keeper.GetAccountTier(s.ctx, testScoreAddress1)
	s.Require().NoError(err)
	s.Require().Equal(types.TierStandard, tier)

	// GetVerificationStatus via interface
	resultStatus := s.keeper.GetVerificationStatus(s.ctx, testScoreAddress1)
	s.Require().Equal(types.AccountStatusVerified, resultStatus)

	// CheckEligibility via interface
	eligibility := s.keeper.CheckEligibility(s.ctx, testScoreAddress1, types.OfferingTypeStandard)
	s.Require().Equal(uint32(75), eligibility.CurrentScore)
}

// ============================================================================
// Tier String Conversion Tests
// ============================================================================

func TestTierToString(t *testing.T) {
	require.Equal(t, "unverified", types.TierToString(types.TierUnverified))
	require.Equal(t, "basic", types.TierToString(types.TierBasic))
	require.Equal(t, "standard", types.TierToString(types.TierStandard))
	require.Equal(t, "premium", types.TierToString(types.TierPremium))
	require.Equal(t, "unknown", types.TierToString(99))
}

func TestTierFromString(t *testing.T) {
	require.Equal(t, types.TierBasic, types.TierFromString("basic"))
	require.Equal(t, types.TierStandard, types.TierFromString("standard"))
	require.Equal(t, types.TierPremium, types.TierFromString("premium"))
	require.Equal(t, types.TierUnverified, types.TierFromString("unknown"))
}

func TestGetMinimumScoreForTier(t *testing.T) {
	require.Equal(t, uint32(0), types.GetMinimumScoreForTier(types.TierUnverified))
	require.Equal(t, uint32(50), types.GetMinimumScoreForTier(types.TierBasic))
	require.Equal(t, uint32(70), types.GetMinimumScoreForTier(types.TierStandard))
	require.Equal(t, uint32(85), types.GetMinimumScoreForTier(types.TierPremium))
}

// ============================================================================
// Identity Score Type Tests
// ============================================================================

func TestIdentityScoreValidation(t *testing.T) {
	// Valid score
	validScore := types.NewIdentityScore(
		testScoreAddress1,
		75,
		types.AccountStatusVerified,
		"v1.0.0",
		time.Now(),
		100,
		[]byte("test-input"),
	)
	require.NoError(t, validScore.Validate())

	// Invalid - empty address
	invalidScore := &types.IdentityScore{
		AccountAddress: "",
		Score:          75,
		Status:         types.AccountStatusVerified,
		ComputedAt:     time.Now(),
	}
	require.Error(t, invalidScore.Validate())

	// Invalid - score exceeds max
	invalidScore2 := &types.IdentityScore{
		AccountAddress: testScoreAddress1,
		Score:          150,
		Status:         types.AccountStatusVerified,
		ComputedAt:     time.Now(),
	}
	require.Error(t, invalidScore2.Validate())

	// Invalid - zero computed time
	invalidScore3 := &types.IdentityScore{
		AccountAddress: testScoreAddress1,
		Score:          75,
		Status:         types.AccountStatusVerified,
	}
	require.Error(t, invalidScore3.Validate())
}

func TestIdentityScoreGetTier(t *testing.T) {
	score := &types.IdentityScore{
		Score:  75,
		Status: types.AccountStatusVerified,
	}
	require.Equal(t, types.TierStandard, score.GetTier())
}

func TestIdentityScoreIsExpired(t *testing.T) {
	// Not expired
	future := time.Now().Add(time.Hour)
	score := &types.IdentityScore{
		ExpiresAt: &future,
	}
	require.False(t, score.IsExpired())

	// Expired
	past := time.Now().Add(-time.Hour)
	score.ExpiresAt = &past
	require.True(t, score.IsExpired())

	// No expiration set
	score.ExpiresAt = nil
	require.False(t, score.IsExpired())
}

func TestIdentityScoreIsAboveThreshold(t *testing.T) {
	score := &types.IdentityScore{
		Score:  75,
		Status: types.AccountStatusVerified,
	}
	require.True(t, score.IsAboveThreshold(70))
	require.True(t, score.IsAboveThreshold(75))
	require.False(t, score.IsAboveThreshold(76))

	// Non-verified doesn't meet threshold
	score.Status = types.AccountStatusPending
	require.False(t, score.IsAboveThreshold(50))
}

// ============================================================================
// Score History Type Tests
// ============================================================================

func TestScoreHistoryAddEntry(t *testing.T) {
	history := types.NewScoreHistory(testScoreAddress1)
	require.Len(t, history.Entries, 0)

	// Add entries
	entry1 := *types.NewScoreHistoryEntry(50, types.AccountStatusVerified, "v1.0.0", time.Now(), 100, "first")
	history.AddEntry(entry1)
	require.Len(t, history.Entries, 1)

	entry2 := *types.NewScoreHistoryEntry(75, types.AccountStatusVerified, "v1.0.0", time.Now(), 101, "second")
	history.AddEntry(entry2)
	require.Len(t, history.Entries, 2)

	// Verify newest first order
	latest, found := history.GetLatest()
	require.True(t, found)
	require.Equal(t, "second", latest.Reason)
}

func TestScoreHistoryGetLatestEmpty(t *testing.T) {
	history := types.NewScoreHistory(testScoreAddress1)
	_, found := history.GetLatest()
	require.False(t, found)
}

// ============================================================================
// Account Status Tests
// ============================================================================

func TestIsValidAccountStatus(t *testing.T) {
	validStatuses := []types.AccountStatus{
		types.AccountStatusUnknown,
		types.AccountStatusPending,
		types.AccountStatusInProgress,
		types.AccountStatusVerified,
		types.AccountStatusRejected,
		types.AccountStatusExpired,
		types.AccountStatusNeedsAdditionalFactor,
	}

	for _, status := range validStatuses {
		require.True(t, types.IsValidAccountStatus(status), "Status: %s", status)
	}

	require.False(t, types.IsValidAccountStatus("invalid_status"))
}

func TestAccountStatusFromVerificationStatus(t *testing.T) {
	testCases := []struct {
		vs       types.VerificationStatus
		expected types.AccountStatus
	}{
		{types.VerificationStatusUnknown, types.AccountStatusUnknown},
		{types.VerificationStatusPending, types.AccountStatusPending},
		{types.VerificationStatusInProgress, types.AccountStatusInProgress},
		{types.VerificationStatusVerified, types.AccountStatusVerified},
		{types.VerificationStatusRejected, types.AccountStatusRejected},
		{types.VerificationStatusExpired, types.AccountStatusExpired},
	}

	for _, tc := range testCases {
		result := types.AccountStatusFromVerificationStatus(tc.vs)
		require.Equal(t, tc.expected, result, "VerificationStatus: %s", tc.vs)
	}
}

// ============================================================================
// Offering Type Tests
// ============================================================================

func TestIsValidOfferingType(t *testing.T) {
	validTypes := []types.OfferingType{
		types.OfferingTypeBasic,
		types.OfferingTypeStandard,
		types.OfferingTypePremium,
		types.OfferingTypeProvider,
		types.OfferingTypeValidator,
	}

	for _, ot := range validTypes {
		require.True(t, types.IsValidOfferingType(ot), "OfferingType: %s", ot)
	}

	require.False(t, types.IsValidOfferingType("invalid_offering"))
}
