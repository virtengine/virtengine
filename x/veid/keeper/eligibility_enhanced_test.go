package keeper

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
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Enhanced Eligibility Test Suite (VE-3033)
// ============================================================================

type EligibilityEnhancedTestSuite struct {
	suite.Suite
	ctx               sdk.Context
	keeper            Keeper
	stateStore        store.CommitMultiStore
	premiumAddress    sdk.AccAddress
	standardAddress   sdk.AccAddress
	basicAddress      sdk.AccAddress
	unverifiedAddress sdk.AccAddress
	expiredAddress    sdk.AccAddress
	lockedAddress     sdk.AccAddress
}

func TestEligibilityEnhancedTestSuite(t *testing.T) {
	suite.Run(t, new(EligibilityEnhancedTestSuite))
}

func (s *EligibilityEnhancedTestSuite) SetupTest() {
	// Create codec
	interfaceRegistry := cdctypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create in-memory store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	s.Require().NoError(err)

	s.stateStore = stateStore

	// Create context with store
	s.ctx = sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())

	// Create keeper
	s.keeper = NewKeeper(cdc, storeKey, "authority")

	// Set default params
	_ = s.keeper.SetParams(s.ctx, types.Params{
		MaxScopesPerAccount:    10,
		MaxScopesPerType:       3,
		SaltMinBytes:           32,
		SaltMaxBytes:           64,
		RequireClientSignature: false,
		RequireUserSignature:   false,
		VerificationExpiryDays: 365,
	})

	// Generate test addresses
	s.premiumAddress = sdk.AccAddress([]byte("premium_address_____"))
	s.standardAddress = sdk.AccAddress([]byte("standard_address____"))
	s.basicAddress = sdk.AccAddress([]byte("basic_address_______"))
	s.unverifiedAddress = sdk.AccAddress([]byte("unverified_address__"))
	s.expiredAddress = sdk.AccAddress([]byte("expired_address_____"))
	s.lockedAddress = sdk.AccAddress([]byte("locked_address______"))

	// Set up premium identity (score 90, all scopes)
	s.setupVerifiedIdentity(s.premiumAddress, 90, []types.ScopeType{
		types.ScopeTypeEmailProof,
		types.ScopeTypeSelfie,
		types.ScopeTypeIDDocument,
		types.ScopeTypeFaceVideo,
	}, false)

	// Set up standard identity (score 75, standard scopes)
	s.setupVerifiedIdentity(s.standardAddress, 75, []types.ScopeType{
		types.ScopeTypeEmailProof,
		types.ScopeTypeSelfie,
		types.ScopeTypeIDDocument,
	}, false)

	// Set up basic identity (score 55, basic scopes)
	s.setupVerifiedIdentity(s.basicAddress, 55, []types.ScopeType{
		types.ScopeTypeEmailProof,
	}, false)

	// Set up expired identity
	s.setupExpiredIdentity(s.expiredAddress)

	// Set up locked identity
	s.setupLockedIdentity(s.lockedAddress)
}

// TearDownTest closes the IAVL store to stop background pruning goroutines
func (s *EligibilityEnhancedTestSuite) TearDownTest() {
	closeStoreIfNeeded(s.stateStore)
}

func (s *EligibilityEnhancedTestSuite) setupVerifiedIdentity(address sdk.AccAddress, score uint32, scopes []types.ScopeType, locked bool) {
	now := s.ctx.BlockTime()

	record := types.NewIdentityRecord(address.String(), now)
	record.CurrentScore = score
	record.LastVerifiedAt = &now
	record.Tier = types.ComputeTierFromScore(score)
	record.Locked = locked

	for _, scopeType := range scopes {
		scopeID := string(scopeType) + "_scope"
		record.AddScopeRef(types.ScopeRef{
			ScopeID:   scopeID,
			ScopeType: scopeType,
		})

		scope := &types.IdentityScope{
			ScopeID:    scopeID,
			ScopeType:  scopeType,
			Version:    types.ScopeSchemaVersion,
			Status:     types.VerificationStatusVerified,
			UploadedAt: now,
		}
		err := s.keeper.setScope(s.ctx, address, scope)
		s.Require().NoError(err)
	}

	err := s.keeper.SetIdentityRecord(s.ctx, *record)
	s.Require().NoError(err)
}

func (s *EligibilityEnhancedTestSuite) setupExpiredIdentity(address sdk.AccAddress) {
	// Set verification 2 years ago (expired)
	verifiedAt := s.ctx.BlockTime().Add(-2 * 365 * 24 * time.Hour)

	record := types.NewIdentityRecord(address.String(), verifiedAt)
	record.CurrentScore = 70
	record.LastVerifiedAt = &verifiedAt
	record.Tier = types.IdentityTierStandard

	err := s.keeper.SetIdentityRecord(s.ctx, *record)
	s.Require().NoError(err)
}

func (s *EligibilityEnhancedTestSuite) setupLockedIdentity(address sdk.AccAddress) {
	now := s.ctx.BlockTime()

	record := types.NewIdentityRecord(address.String(), now)
	record.CurrentScore = 80
	record.LastVerifiedAt = &now
	record.Tier = types.IdentityTierStandard
	record.Locked = true

	err := s.keeper.SetIdentityRecord(s.ctx, *record)
	s.Require().NoError(err)
}

// ============================================================================
// Enhanced Eligibility Check Tests
// ============================================================================

func (s *EligibilityEnhancedTestSuite) TestEnhancedCheckEligibility_PremiumForPremium() {
	result, err := s.keeper.EnhancedCheckEligibility(
		s.ctx,
		s.premiumAddress,
		types.IdentityTierPremium,
		false,
	)

	s.Require().NoError(err)
	s.Assert().True(result.IsEligible)
	s.Assert().Equal(types.IdentityTierPremium, result.RequiredTier)
	s.Assert().Equal(uint32(90), result.CurrentScore)
	s.Assert().Empty(result.FailedChecks)
	s.Assert().Empty(result.Remediation)
	s.Assert().Contains(result.Summary, "eligible")
}

func (s *EligibilityEnhancedTestSuite) TestEnhancedCheckEligibility_BasicForPremium() {
	result, err := s.keeper.EnhancedCheckEligibility(
		s.ctx,
		s.basicAddress,
		types.IdentityTierPremium,
		false,
	)

	s.Require().NoError(err)
	s.Assert().False(result.IsEligible)
	s.Assert().Equal(types.IdentityTierPremium, result.RequiredTier)
	s.Assert().NotEmpty(result.FailedChecks)
	s.Assert().NotEmpty(result.Remediation)
	s.Assert().NotEmpty(result.NextSteps)

	// Should have tier and score failures
	failedTypes := make([]types.EligibilityCheckType, 0)
	for _, check := range result.FailedChecks {
		failedTypes = append(failedTypes, check.CheckType)
	}
	s.Assert().Contains(failedTypes, types.EligibilityCheckTypeTierRequirement)
	s.Assert().Contains(failedTypes, types.EligibilityCheckTypeScoreThreshold)
}

func (s *EligibilityEnhancedTestSuite) TestEnhancedCheckEligibility_UnverifiedAddress() {
	result, err := s.keeper.EnhancedCheckEligibility(
		s.ctx,
		s.unverifiedAddress,
		types.IdentityTierBasic,
		false,
	)

	s.Require().NoError(err)
	s.Assert().False(result.IsEligible)
	s.Assert().Equal(types.IdentityTierUnverified, result.CurrentTier)
	s.Assert().Equal(uint32(0), result.CurrentScore)

	// Should have account status check failure
	s.Assert().Len(result.FailedChecks, 1)
	s.Assert().Equal(types.EligibilityCheckTypeAccountStatus, result.FailedChecks[0].CheckType)
}

func (s *EligibilityEnhancedTestSuite) TestEnhancedCheckEligibility_LockedAccount() {
	result, err := s.keeper.EnhancedCheckEligibility(
		s.ctx,
		s.lockedAddress,
		types.IdentityTierBasic,
		false,
	)

	s.Require().NoError(err)
	s.Assert().False(result.IsEligible)

	// Should have account status failure for locked account
	hasAccountStatusFailure := false
	for _, check := range result.FailedChecks {
		if check.CheckType == types.EligibilityCheckTypeAccountStatus {
			hasAccountStatusFailure = true
			s.Assert().Contains(check.Reason, "locked")
		}
	}
	s.Assert().True(hasAccountStatusFailure)
}

func (s *EligibilityEnhancedTestSuite) TestEnhancedCheckEligibility_ExpiredVerification() {
	result, err := s.keeper.EnhancedCheckEligibility(
		s.ctx,
		s.expiredAddress,
		types.IdentityTierBasic,
		false,
	)

	s.Require().NoError(err)
	s.Assert().False(result.IsEligible)

	// Should have verification age failure
	hasAgeFailure := false
	for _, check := range result.FailedChecks {
		if check.CheckType == types.EligibilityCheckTypeVerificationAge {
			hasAgeFailure = true
			s.Assert().Contains(check.Reason, "expired")
		}
	}
	s.Assert().True(hasAgeFailure)
}

func (s *EligibilityEnhancedTestSuite) TestEnhancedCheckEligibility_MissingScopes() {
	// Standard address doesn't have FaceVideo scope required for premium
	result, err := s.keeper.EnhancedCheckEligibility(
		s.ctx,
		s.standardAddress,
		types.IdentityTierPremium,
		false,
	)

	s.Require().NoError(err)
	s.Assert().False(result.IsEligible)
	s.Assert().Contains(result.MissingScopes, types.ScopeTypeFaceVideo)
}

func (s *EligibilityEnhancedTestSuite) TestEnhancedCheckEligibility_StandardForBasic() {
	// Standard address should be eligible for basic tier
	result, err := s.keeper.EnhancedCheckEligibility(
		s.ctx,
		s.standardAddress,
		types.IdentityTierBasic,
		false,
	)

	s.Require().NoError(err)
	s.Assert().True(result.IsEligible)
	s.Assert().Empty(result.FailedChecks)
}

// ============================================================================
// Remediation Hints Tests
// ============================================================================

func (s *EligibilityEnhancedTestSuite) TestGetRemediationHints_BasicForPremium() {
	hints, err := s.keeper.GetRemediationHints(
		s.ctx,
		s.basicAddress,
		types.IdentityTierPremium,
	)

	s.Require().NoError(err)
	s.Assert().NotEmpty(hints)

	// Should include hints for missing scopes
	categories := make([]string, 0)
	for _, hint := range hints {
		categories = append(categories, hint.Category)
		s.Assert().NotEmpty(hint.Issue)
		s.Assert().NotEmpty(hint.Action)
	}

	// Should include biometric and document verification hints
	s.Assert().Contains(categories, "biometric_verification")
	s.Assert().Contains(categories, "document_verification")
}

func (s *EligibilityEnhancedTestSuite) TestGetRemediationHints_UnverifiedAddress() {
	hints, err := s.keeper.GetRemediationHints(
		s.ctx,
		s.unverifiedAddress,
		types.IdentityTierBasic,
	)

	s.Require().NoError(err)
	s.Assert().NotEmpty(hints)

	// First hint should be about registration
	s.Assert().Equal("registration", hints[0].Category)
}

func (s *EligibilityEnhancedTestSuite) TestGetRemediationHints_SortedByPriority() {
	hints, err := s.keeper.GetRemediationHints(
		s.ctx,
		s.basicAddress,
		types.IdentityTierPremium,
	)

	s.Require().NoError(err)
	s.Assert().NotEmpty(hints)

	// Verify hints are sorted by priority (ascending)
	for i := 0; i < len(hints)-1; i++ {
		s.Assert().LessOrEqual(hints[i].Priority, hints[i+1].Priority,
			"Hints should be sorted by priority")
	}
}

// ============================================================================
// Tier Requirements Tests
// ============================================================================

func (s *EligibilityEnhancedTestSuite) TestGetTierRequirements_AllTiers() {
	tiers := []types.IdentityTier{
		types.IdentityTierUnverified,
		types.IdentityTierBasic,
		types.IdentityTierStandard,
		types.IdentityTierPremium,
	}

	for _, tier := range tiers {
		reqs, err := s.keeper.GetTierRequirements(s.ctx, tier)
		s.Require().NoError(err)
		s.Assert().NotNil(reqs)
		s.Assert().Equal(tier, reqs.Tier)
		s.Assert().NotEmpty(reqs.Description)
	}
}

func (s *EligibilityEnhancedTestSuite) TestGetTierRequirements_Basic() {
	reqs, err := s.keeper.GetTierRequirements(s.ctx, types.IdentityTierBasic)

	s.Require().NoError(err)
	s.Assert().Equal(types.ThresholdBasic, reqs.MinScore)
	s.Assert().Equal(types.ThresholdStandard, reqs.MaxScore)
	s.Assert().Contains(reqs.RequiredScopes, types.ScopeTypeEmailProof)
	s.Assert().False(reqs.RequiresMFA)
}

func (s *EligibilityEnhancedTestSuite) TestGetTierRequirements_Standard() {
	reqs, err := s.keeper.GetTierRequirements(s.ctx, types.IdentityTierStandard)

	s.Require().NoError(err)
	s.Assert().Equal(types.ThresholdStandard, reqs.MinScore)
	s.Assert().Equal(types.ThresholdPremium, reqs.MaxScore)
	s.Assert().Contains(reqs.RequiredScopes, types.ScopeTypeEmailProof)
	s.Assert().Contains(reqs.RequiredScopes, types.ScopeTypeSelfie)
	s.Assert().Contains(reqs.RequiredScopes, types.ScopeTypeIDDocument)
	s.Assert().True(reqs.RequiresMFA)
}

func (s *EligibilityEnhancedTestSuite) TestGetTierRequirements_Premium() {
	reqs, err := s.keeper.GetTierRequirements(s.ctx, types.IdentityTierPremium)

	s.Require().NoError(err)
	s.Assert().Equal(types.ThresholdPremium, reqs.MinScore)
	s.Assert().Contains(reqs.RequiredScopes, types.ScopeTypeFaceVideo)
	s.Assert().True(reqs.RequiresMFA)
	s.Assert().NotEmpty(reqs.Benefits)
}

func (s *EligibilityEnhancedTestSuite) TestGetAllTierRequirements() {
	allReqs := s.keeper.GetAllTierRequirements(s.ctx)

	s.Assert().Len(allReqs, 4) // Unverified, Basic, Standard, Premium

	// Verify order
	s.Assert().Equal(types.IdentityTierUnverified, allReqs[0].Tier)
	s.Assert().Equal(types.IdentityTierBasic, allReqs[1].Tier)
	s.Assert().Equal(types.IdentityTierStandard, allReqs[2].Tier)
	s.Assert().Equal(types.IdentityTierPremium, allReqs[3].Tier)
}

// ============================================================================
// Failure Explanation Tests
// ============================================================================

func (s *EligibilityEnhancedTestSuite) TestExplainFailure_NotEligible() {
	explanation, err := s.keeper.ExplainFailure(
		s.ctx,
		s.basicAddress,
		types.IdentityTierPremium,
	)

	s.Require().NoError(err)
	s.Assert().NotNil(explanation)
	s.Assert().NotEmpty(explanation.Title)
	s.Assert().NotEmpty(explanation.Description)
	s.Assert().NotEmpty(explanation.FailedChecks)
	s.Assert().NotEmpty(explanation.Steps)
}

func (s *EligibilityEnhancedTestSuite) TestExplainFailure_IsEligible() {
	explanation, err := s.keeper.ExplainFailure(
		s.ctx,
		s.premiumAddress,
		types.IdentityTierPremium,
	)

	s.Require().NoError(err)
	s.Assert().Nil(explanation) // No failure to explain
}

func (s *EligibilityEnhancedTestSuite) TestExplainFailure_Severity() {
	// Test critical severity for locked account
	explanation, err := s.keeper.ExplainFailure(
		s.ctx,
		s.lockedAddress,
		types.IdentityTierBasic,
	)

	s.Require().NoError(err)
	s.Assert().NotNil(explanation)
	s.Assert().Equal("critical", explanation.Severity)
}

// ============================================================================
// Progress Tracking Tests
// ============================================================================

func (s *EligibilityEnhancedTestSuite) TestGetProgressToTier_BasicToPremium() {
	progress, err := s.keeper.GetProgressToTier(
		s.ctx,
		s.basicAddress,
		types.IdentityTierPremium,
	)

	s.Require().NoError(err)
	s.Assert().NotNil(progress)
	s.Assert().Equal(types.IdentityTierPremium, progress.TargetTier)
	s.Assert().Less(progress.PercentComplete, float64(100))
	s.Assert().NotEmpty(progress.MissingScopes)
	s.Assert().NotEmpty(progress.RemainingSteps)
}

func (s *EligibilityEnhancedTestSuite) TestGetProgressToTier_PremiumToPremium() {
	progress, err := s.keeper.GetProgressToTier(
		s.ctx,
		s.premiumAddress,
		types.IdentityTierPremium,
	)

	s.Require().NoError(err)
	s.Assert().NotNil(progress)
	s.Assert().Equal(float64(100), progress.ScoreProgress)
	s.Assert().Empty(progress.MissingScopes)
}

func (s *EligibilityEnhancedTestSuite) TestGetProgressToTier_UnverifiedAddress() {
	progress, err := s.keeper.GetProgressToTier(
		s.ctx,
		s.unverifiedAddress,
		types.IdentityTierBasic,
	)

	s.Require().NoError(err)
	s.Assert().NotNil(progress)
	s.Assert().Equal(types.IdentityTierUnverified, progress.CurrentTier)
	s.Assert().Equal(float64(0), progress.PercentComplete)
	s.Assert().NotEmpty(progress.RemainingSteps)
}

// ============================================================================
// Market Integration Tests
// ============================================================================

func (s *EligibilityEnhancedTestSuite) TestEnhancedCheckEligibilityForMarket_Compute() {
	result, err := s.keeper.EnhancedCheckEligibilityForMarket(
		s.ctx,
		s.basicAddress,
		types.MarketTypeCompute,
	)

	s.Require().NoError(err)
	s.Assert().NotNil(result)
	s.Assert().Equal(types.MarketTypeCompute, result.Context.MarketType)
}

func (s *EligibilityEnhancedTestSuite) TestEnhancedCheckEligibilityForMarket_TEE() {
	// Basic address should not be eligible for TEE
	result, err := s.keeper.EnhancedCheckEligibilityForMarket(
		s.ctx,
		s.basicAddress,
		types.MarketTypeTEE,
	)

	s.Require().NoError(err)
	s.Assert().NotNil(result)
	s.Assert().False(result.IsEligible)
	s.Assert().Equal(types.MarketTypeTEE, result.Context.MarketType)
}

func (s *EligibilityEnhancedTestSuite) TestEnhancedCheckEligibilityForMarket_PremiumForTEE() {
	result, err := s.keeper.EnhancedCheckEligibilityForMarket(
		s.ctx,
		s.premiumAddress,
		types.MarketTypeTEE,
	)

	s.Require().NoError(err)
	s.Assert().NotNil(result)
	// Premium address should meet TEE requirements (score-wise)
	// MFA check might fail if MFA keeper not set, but score/tier should pass
}

// ============================================================================
// Types Tests
// ============================================================================

func TestEligibilityCheck_WithValues(t *testing.T) {
	check := types.NewEligibilityCheck(
		types.EligibilityCheckTypeScoreThreshold,
		false,
		"Score too low",
	).WithValues(50, 85)

	assert.Equal(t, 50, check.CurrentValue)
	assert.Equal(t, 85, check.RequiredValue)
}

func TestEligibilityCheck_WithWeight(t *testing.T) {
	check := types.NewEligibilityCheck(
		types.EligibilityCheckTypeTierRequirement,
		true,
		"Tier met",
	).WithWeight(100)

	assert.Equal(t, 100, check.Weight)
}

func TestRemediationHint_Builder(t *testing.T) {
	hint := types.NewRemediationHint("Issue", "Action").
		WithDocuments("Doc1", "Doc2").
		WithEstimatedTime(15 * time.Minute).
		WithPriority(1).
		WithCategory("verification")

	assert.Equal(t, "Issue", hint.Issue)
	assert.Equal(t, "Action", hint.Action)
	assert.Len(t, hint.DocumentsNeeded, 2)
	assert.Equal(t, 15*time.Minute, hint.EstimatedTime)
	assert.Equal(t, 1, hint.Priority)
	assert.Equal(t, "verification", hint.Category)
}

func TestEnhancedEligibilityResult_AddCheck(t *testing.T) {
	result := types.NewEnhancedEligibilityResult("addr", time.Now())

	result.AddCheck(types.NewEligibilityCheck(
		types.EligibilityCheckTypeAccountStatus,
		true,
		"Active",
	))
	result.AddCheck(types.NewEligibilityCheck(
		types.EligibilityCheckTypeScoreThreshold,
		false,
		"Low score",
	))

	assert.Len(t, result.Checks, 2)
	assert.Len(t, result.FailedChecks, 1)
	assert.Equal(t, 1, result.PassedCheckCount())
	assert.Equal(t, 2, result.TotalCheckCount())
}

func TestEnhancedEligibilityResult_Finalize(t *testing.T) {
	result := types.NewEnhancedEligibilityResult("addr123", time.Now())

	result.AddCheck(types.NewEligibilityCheck(
		types.EligibilityCheckTypeAccountStatus,
		true,
		"Active",
	))
	result.CurrentTier = types.IdentityTierBasic
	result.CurrentScore = 55

	result.Finalize()

	assert.True(t, result.IsEligible)
	assert.Contains(t, result.Summary, "eligible")
	assert.Contains(t, result.Summary, "addr123")
}

func TestEnhancedEligibilityResult_GetFailureReasons(t *testing.T) {
	result := types.NewEnhancedEligibilityResult("addr", time.Now())

	result.AddCheck(types.NewEligibilityCheck(
		types.EligibilityCheckTypeScoreThreshold,
		false,
		"Low score",
	))
	result.AddCheck(types.NewEligibilityCheck(
		types.EligibilityCheckTypeTierRequirement,
		false,
		"Wrong tier",
	))

	reasons := result.GetFailureReasons()
	assert.Len(t, reasons, 2)
	assert.Contains(t, reasons, "Low score")
	assert.Contains(t, reasons, "Wrong tier")
}

func TestGetTierRequirements_AllTiers(t *testing.T) {
	tiers := []types.IdentityTier{
		types.IdentityTierUnverified,
		types.IdentityTierBasic,
		types.IdentityTierStandard,
		types.IdentityTierPremium,
	}

	for _, tier := range tiers {
		reqs := types.GetTierRequirements(tier)
		assert.NotNil(t, reqs, "Requirements should exist for tier: %s", tier)
		assert.Equal(t, tier, reqs.Tier)
		assert.NotEmpty(t, reqs.Description)
	}
}

func TestGetTierRequirements_Unknown(t *testing.T) {
	reqs := types.GetTierRequirements("unknown_tier")
	assert.Nil(t, reqs)
}

func TestGetRemediationForScope(t *testing.T) {
	scopeTypes := []types.ScopeType{
		types.ScopeTypeEmailProof,
		types.ScopeTypeIDDocument,
		types.ScopeTypeSelfie,
		types.ScopeTypeFaceVideo,
		types.ScopeTypeSMSProof,
		types.ScopeTypeDomainVerify,
		types.ScopeTypeBiometric,
		types.ScopeTypeADSSO,
	}

	for _, scopeType := range scopeTypes {
		hint := types.GetRemediationForScope(scopeType)
		assert.NotEmpty(t, hint.Issue, "Hint should have issue for scope: %s", scopeType)
		assert.NotEmpty(t, hint.Action, "Hint should have action for scope: %s", scopeType)
		assert.Greater(t, hint.EstimatedTime, time.Duration(0), "Hint should have estimated time for scope: %s", scopeType)
	}
}

func TestNewFailureExplanation(t *testing.T) {
	result := types.NewEnhancedEligibilityResult("addr", time.Now())
	result.AddCheck(types.NewEligibilityCheck(
		types.EligibilityCheckTypeScoreThreshold,
		false,
		"Score below threshold",
	))
	result.AddNextStep("Improve score")
	result.Finalize()

	explanation := types.NewFailureExplanation(result)

	assert.NotNil(t, explanation)
	assert.NotEmpty(t, explanation.Title)
	assert.NotEmpty(t, explanation.Description)
	assert.NotEmpty(t, explanation.FailedChecks)
	assert.NotEmpty(t, explanation.Steps)
}

func TestNewFailureExplanation_Eligible(t *testing.T) {
	result := types.NewEnhancedEligibilityResult("addr", time.Now())
	result.AddCheck(types.NewEligibilityCheck(
		types.EligibilityCheckTypeAccountStatus,
		true,
		"Active",
	))
	result.Finalize()

	explanation := types.NewFailureExplanation(result)
	assert.Nil(t, explanation, "No explanation for eligible result")
}

func TestAllEligibilityCheckTypes(t *testing.T) {
	checkTypes := types.AllEligibilityCheckTypes()
	assert.Len(t, checkTypes, 8)
	assert.Contains(t, checkTypes, types.EligibilityCheckTypeTierRequirement)
	assert.Contains(t, checkTypes, types.EligibilityCheckTypeScoreThreshold)
	assert.Contains(t, checkTypes, types.EligibilityCheckTypeScopeRequired)
	assert.Contains(t, checkTypes, types.EligibilityCheckTypeAccountStatus)
	assert.Contains(t, checkTypes, types.EligibilityCheckTypeMFARequired)
	assert.Contains(t, checkTypes, types.EligibilityCheckTypeVerificationAge)
}

func TestGetAllTierRequirements(t *testing.T) {
	allReqs := types.GetAllTierRequirements()
	assert.Len(t, allReqs, 4)

	// Verify ascending order by MinScore
	for i := 0; i < len(allReqs)-1; i++ {
		assert.Less(t, allReqs[i].MinScore, allReqs[i+1].MinScore,
			"Requirements should be in ascending order by MinScore")
	}
}

func TestGetRequiredScopesForTier(t *testing.T) {
	// Basic tier should require email
	basicScopes := types.GetRequiredScopesForTier(types.IdentityTierBasic)
	assert.Contains(t, basicScopes, types.ScopeTypeEmailProof)

	// Premium tier should require more scopes
	premiumScopes := types.GetRequiredScopesForTier(types.IdentityTierPremium)
	assert.Contains(t, premiumScopes, types.ScopeTypeEmailProof)
	assert.Contains(t, premiumScopes, types.ScopeTypeSelfie)
	assert.Contains(t, premiumScopes, types.ScopeTypeIDDocument)
	assert.Contains(t, premiumScopes, types.ScopeTypeFaceVideo)

	// Unknown tier should return empty
	unknownScopes := types.GetRequiredScopesForTier("unknown")
	assert.Empty(t, unknownScopes)
}
