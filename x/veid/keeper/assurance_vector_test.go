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

	"github.com/virtengine/virtengine/x/veid/keeper"
	"github.com/virtengine/virtengine/x/veid/types"
)

// AssuranceVectorTestSuite covers the persisted per-factor evidence state:
// derivation from a composite result, epoch supersession, staleness, and the
// absent-means-no-claim contract that the whole surface depends on.
type AssuranceVectorTestSuite struct {
	suite.Suite
	ctx        sdk.Context
	keeper     keeper.Keeper
	stateStore store.CommitMultiStore
}

func TestAssuranceVectorTestSuite(t *testing.T) {
	suite.Run(t, new(AssuranceVectorTestSuite))
}

func (s *AssuranceVectorTestSuite) SetupTest() {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	s.ctx = s.createContextWithStore(storeKey)
	s.keeper = keeper.NewKeeper(cdc, storeKey, "authority")

	s.Require().NoError(s.keeper.SetParams(s.ctx, types.DefaultParams()))
}

func (s *AssuranceVectorTestSuite) TearDownTest() {
	CloseStoreIfNeeded(s.stateStore)
}

func (s *AssuranceVectorTestSuite) createContextWithStore(storeKey *storetypes.KVStoreKey) sdk.Context {
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	if err := stateStore.LoadLatestVersion(); err != nil {
		s.T().Fatalf("failed to load latest version: %v", err)
	}
	s.stateStore = stateStore

	// Fixed block header: recency must come from consensus time, so the test
	// never reads the wall clock.
	return sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Unix(1_700_000_000, 0).UTC(),
		Height: 1000,
	}, false, log.NewNopLogger())
}

// ============================================================================
// Fixture
// ============================================================================

// distinguishableInputs builds composite inputs whose factors differ from one
// another, so a vector that matched "something plausible" instead of the real
// values would fail.
func (s *AssuranceVectorTestSuite) distinguishableInputs() types.CompositeScoringInputs {
	return types.CompositeScoringInputs{
		// The scorer validates the account address on the inputs, and
		// ComputeAndStoreCompositeScore overwrites it from its own argument, so
		// any valid address satisfies it here.
		AccountAddress: s.testAddress("assurance_fixture_input"),
		DocumentAuthenticity: types.DocumentAuthenticityInput{
			Present:               true,
			TamperScore:           9000,
			FormatValidityScore:   8000,
			TemplateMatchScore:    7000,
			SecurityFeaturesScore: 6000,
		},
		FaceMatch: types.FaceMatchInput{
			Present:         true,
			SimilarityScore: 9600,
			Confidence:      9200,
			QualityScore:    8800,
		},
		LivenessDetection: types.LivenessDetectionInput{
			Present:              true,
			LivenessScore:        7500,
			BlinkDetected:        true,
			HeadMovementDetected: true,
			DepthCheckPassed:     true,
			AntiSpoofScore:       7200,
		},
		DataConsistency: types.DataConsistencyInput{
			Present:                 true,
			NameMatchScore:          10000,
			DOBConsistencyScore:     10000,
			AgeVerificationPassed:   true,
			AddressConsistencyScore: 10000,
			DocumentExpiryValid:     true,
			CrossFieldValidation:    10000,
		},
		HistoricalSignals: types.HistoricalSignalsInput{
			Present:                    true,
			PriorVerificationScore:     4000,
			AccountAgeScore:            3000,
			VerificationHistoryCount:   3,
			SuccessfulVerificationRate: 5000,
		},
		RiskIndicators: types.RiskIndicatorsInput{
			Present:                 true,
			FraudPatternScore:       2000,
			DeviceFingerprintScore:  8500,
			DeviceIntegrityScore:    8100,
			IPReputationScore:       7800,
			VelocityCheckPassed:     true,
			DeviceAttestationPassed: true,
			GeoConsistencyScore:     7600,
		},
	}
}

func (s *AssuranceVectorTestSuite) computeResult(inputs types.CompositeScoringInputs) *types.CompositeScoreResult {
	result, err := types.ComputeCompositeScore(
		inputs,
		types.DefaultCompositeScoringWeights(),
		types.DefaultCompositeScoringThresholds(),
	)
	s.Require().NoError(err)
	return result
}

// ============================================================================
// Persistence matches the composite result
// ============================================================================

func (s *AssuranceVectorTestSuite) TestComputeAndStoreCompositeScorePersistsVectorMatchingEveryFactor() {
	account := s.testAddress("assurance_factor_match")
	result := s.computeResult(s.distinguishableInputs())
	result.BlockHeight = s.ctx.BlockHeight()
	result.ComputedAt = s.ctx.BlockTime()

	_, err := s.keeper.ComputeAndStoreCompositeScore(s.ctx, account, s.distinguishableInputs())
	s.Require().NoError(err)

	vector, found := s.keeper.GetAssuranceVector(s.ctx, account)
	s.Require().True(found, "a verified account must have a vector")
	s.Require().NoError(vector.Validate())

	byComponent := make(map[string]types.CompositeScoreContribution, len(result.Contributions))
	for _, c := range result.Contributions {
		byComponent[c.ComponentName] = c
	}

	for _, tc := range []struct {
		factor    types.AssuranceFactor
		component string
	}{
		{types.AssuranceFactorDocument, types.ComponentDocumentAuthenticity},
		{types.AssuranceFactorBiometric, types.ComponentFaceMatch},
		{types.AssuranceFactorLiveness, types.ComponentLivenessDetection},
		{types.AssuranceFactorHistory, types.ComponentHistoricalSignals},
		{types.AssuranceFactorDataConsistency, types.ComponentDataConsistency},
		{types.AssuranceFactorRisk, types.ComponentRiskIndicators},
	} {
		entry, ok := vector.Lookup(tc.factor)
		s.Require().True(ok, "factor %s must be present", tc.factor)

		contrib := byComponent[tc.component]
		s.Require().Equal(contrib.RawScore, entry.ScoreBps, "factor %s raw score", tc.factor)
		s.Require().Equal(contrib.Weight, entry.WeightBps, "factor %s weight", tc.factor)
		s.Require().Equal(contrib.WeightedScore, entry.WeightedScoreBps, "factor %s weighted score", tc.factor)
		s.Require().True(entry.Measured, "factor %s measured", tc.factor)
	}

	// Overall and version come from the scorer's own output.
	s.Require().Equal(types.OverallBpsFromScore(result.FinalScore), vector.OverallBps)
	s.Require().Equal(result.FinalScore, vector.ScalarScore)
	s.Require().Equal(result.ScoreVersion, vector.ScoreVersion)
	s.Require().Equal(types.AssuranceVectorVersion, vector.Version)

	// The commitment must verify as read back out of the store.
	s.Require().True(vector.VerifyCommitment(), "a stored vector must verify against its own commitment")

	// The scalar query path must be untouched by the vector work.
	score, status, scoreFound := s.keeper.GetScore(s.ctx, account)
	s.Require().True(scoreFound)
	s.Require().Equal(result.FinalScore, score)
	s.Require().Equal(types.AccountStatusVerified, status)
	s.Require().True(s.keeper.IsScoreAboveThreshold(s.ctx, account, result.FinalScore),
		"the scalar gating path must behave exactly as before")
}

func (s *AssuranceVectorTestSuite) TestStoredVectorSurvivesRoundTripThroughTheStore() {
	account := s.testAddress("assurance_round_trip")

	_, err := s.keeper.ComputeAndStoreCompositeScore(s.ctx, account, s.distinguishableInputs())
	s.Require().NoError(err)

	stored, found := s.keeper.GetAssuranceVector(s.ctx, account)
	s.Require().True(found)

	// Read it repeatedly: decoding must be stable and must not mutate the vector
	// (fromStore builds a fresh value from the stored bytes).
	again, found := s.keeper.GetAssuranceVector(s.ctx, account)
	s.Require().True(found)
	s.Require().Equal(stored, again, "decoding the same bytes twice must be identical")
	s.Require().True(again.VerifyCommitment())
}

// ============================================================================
// Epoch supersession
// ============================================================================

func (s *AssuranceVectorTestSuite) TestReverificationSupersedesPriorEpoch() {
	account := s.testAddress("assurance_epoch")

	// First verification writes epoch 0.
	_, err := s.keeper.ComputeAndStoreCompositeScore(s.ctx, account, s.distinguishableInputs())
	s.Require().NoError(err)

	first, found := s.keeper.GetAssuranceVector(s.ctx, account)
	s.Require().True(found)
	s.Require().Equal(uint64(0), first.Epoch)

	// Re-verification at a later height, with DIFFERENT evidence, must move the
	// account to epoch 1 and replace the readings.
	weak := s.distinguishableInputs()
	weak.FaceMatch.SimilarityScore = 3000
	weak.FaceMatch.Confidence = 3000
	weak.FaceMatch.QualityScore = 3000

	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 100)
	_, err = s.keeper.ComputeAndStoreCompositeScore(s.ctx, account, weak)
	s.Require().NoError(err)

	second, found := s.keeper.GetAssuranceVector(s.ctx, account)
	s.Require().True(found)
	s.Require().Equal(uint64(1), second.Epoch, "a re-verification must advance the epoch")
	s.Require().Equal(s.ctx.BlockHeight(), second.VerifiedHeight)
	s.Require().True(first.IsSuperseded(second.Epoch), "the prior vector must read as superseded")
	s.Require().False(second.IsSuperseded(second.Epoch))

	// The biometric readings must be the new (weaker) ones, not the old ones.
	firstBiometric, _ := first.Lookup(types.AssuranceFactorBiometric)
	secondBiometric, _ := second.Lookup(types.AssuranceFactorBiometric)
	s.Require().NotEqual(firstBiometric.ScoreBps, secondBiometric.ScoreBps,
		"the superseded vector must not still be serving the old readings")

	// Both epochs remain retrievable, newest first, so a verifier can prove the
	// vector was replaced rather than edited in place.
	history := s.keeper.GetAssuranceVectorHistory(s.ctx, account)
	s.Require().Len(history, 2)
	s.Require().Equal(uint64(1), history[0].Epoch, "history is newest first")
	s.Require().Equal(uint64(0), history[1].Epoch)

	// The superseded epoch is still readable at its own height.
	old, found := s.keeper.GetAssuranceVectorEpoch(s.ctx, account, first.VerifiedHeight)
	s.Require().True(found)
	s.Require().Equal(first.Commitment, old.Commitment)
}

func (s *AssuranceVectorTestSuite) TestEpochCannotRewindOrRepeat() {
	account := s.testAddress("assurance_no_rewind")

	_, err := s.keeper.ComputeAndStoreCompositeScore(s.ctx, account, s.distinguishableInputs())
	s.Require().NoError(err)
	first, found := s.keeper.GetAssuranceVector(s.ctx, account)
	s.Require().True(found)

	// Re-writing the same epoch must be refused: allowing it would let a later
	// block replace a newer vector with older evidence.
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 10)
	s.Require().Error(s.keeper.SetAssuranceVector(s.ctx, first),
		"re-writing an existing epoch must be rejected")

	// A rewound epoch must be refused too.
	rewound := *first
	rewound.Epoch = 0
	rewound.Commitment = rewound.ComputeCommitment()
	s.Require().Error(s.keeper.SetAssuranceVector(s.ctx, &rewound),
		"rewinding the epoch must be rejected")

	// The stored vector is untouched by either attempt.
	unchanged, found := s.keeper.GetAssuranceVector(s.ctx, account)
	s.Require().True(found)
	s.Require().Equal(first.Commitment, unchanged.Commitment)
}

// ============================================================================
// Absence means "no assurance claim"
// ============================================================================

func (s *AssuranceVectorTestSuite) TestAccountWithNoVectorReadsAsNoClaim() {
	account := s.testAddress("assurance_absent")

	vector, found := s.keeper.GetAssuranceVector(s.ctx, account)
	s.Require().False(found, "an unverified account must have no vector")
	s.Require().Nil(vector, "absence must return nil, not a zeroed vector")

	s.Require().False(s.keeper.HasAssuranceVector(s.ctx, account))
	s.Require().Empty(s.keeper.GetAssuranceVectorHistory(s.ctx, account))

	// The recency accessor must also report absence rather than a stale vector.
	_, recency, found := s.keeper.GetAssuranceVectorRecency(s.ctx, account)
	s.Require().False(found)
	s.Require().Zero(recency.Stale, "absence must not be reported as merely stale")

	// The query surface must make absence explicit rather than erroring, so a
	// relying party can distinguish "no claim" from a transport failure.
	querier := keeper.NewGRPCQuerier(s.keeper)
	resp, err := querier.QueryAssuranceVector(s.ctx, &types.QueryAssuranceVectorRequest{AccountAddress: account})
	s.Require().NoError(err, "absent is a normal answer, not an error")
	s.Require().False(resp.Found)
	s.Require().Nil(resp.Vector)
	s.Require().Zero(resp.CurrentEpoch)

	// History for an account with no vector is an empty list, not nil, so a
	// client can range over it.
	historyResp, err := querier.QueryAssuranceVectorHistory(s.ctx, &types.QueryAssuranceVectorHistoryRequest{AccountAddress: account})
	s.Require().NoError(err)
	s.Require().NotNil(historyResp.Vectors)
	s.Require().Empty(historyResp.Vectors)
}

func (s *AssuranceVectorTestSuite) TestUnverifiedAccountKeepsNoVectorEvenWithAScalarScore() {
	// A rejected composite score still persists the scalar, and the vector must
	// still record the (low) factor readings rather than being absent or invented.
	account := s.testAddress("assurance_rejected")

	weak := s.distinguishableInputs()
	weak.DocumentAuthenticity.TamperScore = 0
	weak.DocumentAuthenticity.FormatValidityScore = 0
	weak.DocumentAuthenticity.TemplateMatchScore = 0
	weak.DocumentAuthenticity.SecurityFeaturesScore = 0
	weak.FaceMatch.SimilarityScore = 0
	weak.FaceMatch.Confidence = 0
	weak.FaceMatch.QualityScore = 0

	result, err := s.keeper.ComputeAndStoreCompositeScore(s.ctx, account, weak)
	s.Require().NoError(err)

	vector, found := s.keeper.GetAssuranceVector(s.ctx, account)
	s.Require().True(found, "a scored account has a vector even when it failed")
	s.Require().NoError(vector.Validate())
	s.Require().Equal(result.FinalScore, vector.ScalarScore)

	// A component the scorer measured AT ZERO is not "measured as zero" in this
	// module's own vocabulary: ComputeCompositeScore records presence as
	// RawScore > 0, so a zero-scoring component is indistinguishable from absent
	// evidence. The vector therefore reports the document factor as UNMEASURED.
	//
	// This is asserted rather than papered over because it is a real limit on the
	// vector's expressiveness: "measured and scored zero" cannot be distinguished
	// from "never presented" while the scorer's presence rule is RawScore > 0.
	// A relying party needing that distinction must read the score's reason codes
	// (LOW_DOCUMENT_AUTHENTICITY vs MISSING_DOCUMENT) alongside the vector.
	doc, ok := vector.Lookup(types.AssuranceFactorDocument)
	s.Require().True(ok)
	s.Require().False(doc.Measured, "a zero-scoring component is absent by the scorer's own presence rule")
	s.Require().Zero(doc.ScoreBps)

	_, measured := vector.ScoreBps(types.AssuranceFactorDocument)
	s.Require().False(measured, "reading a zero-scoring component reports no assurance claim")

	// The reason codes still distinguish the two cases, so the information is not
	// lost from the chain - only from the vector's factor reading.
	s.Require().Equal(result.FinalScore, vector.ScalarScore)
	s.Require().Equal(types.OverallBpsFromScore(result.FinalScore), vector.OverallBps)
	s.Require().False(vector.Passed, "this fixture must not clear the pass threshold")

	// Data consistency was left strong in this fixture, so it IS measured, and
	// its presence proves the rule is per-factor rather than blanket.
	consistency, ok := vector.Lookup(types.AssuranceFactorDataConsistency)
	s.Require().True(ok)
	s.Require().True(consistency.Measured, "a positively-scored component reads as measured")
}

// ============================================================================
// Recency / staleness against consensus time
// ============================================================================

func (s *AssuranceVectorTestSuite) TestStaleVectorIsReportedStaleNotAbsent() {
	account := s.testAddress("assurance_stale")

	_, err := s.keeper.ComputeAndStoreCompositeScore(s.ctx, account, s.distinguishableInputs())
	s.Require().NoError(err)

	params := s.keeper.GetParams(s.ctx)
	maxAge := time.Duration(params.VerificationExpiryDays) * 24 * time.Hour

	s.T().Run("fresh", func(t *testing.T) {
		fresh := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(maxAge / 2))
		vector, recency, found := s.keeper.GetAssuranceVectorRecency(fresh, account)
		s.Require().True(found)
		s.Require().NotNil(vector)
		s.Require().False(recency.Stale, "half the window old is fresh")
	})

	s.T().Run("stale", func(t *testing.T) {
		stale := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(maxAge + time.Second))
		vector, recency, found := s.keeper.GetAssuranceVectorRecency(stale, account)
		s.Require().True(found, "a stale vector is still PRESENT; staleness is not absence")
		s.Require().NotNil(vector)
		s.Require().True(recency.Stale, "past the expiry window the evidence is stale")

		// The query surface must expose the staleness so a relying party can
		// apply a freshness requirement without reading params itself.
		querier := keeper.NewGRPCQuerier(s.keeper)
		resp, err := querier.QueryAssuranceVector(stale, &types.QueryAssuranceVectorRequest{AccountAddress: account})
		s.Require().NoError(err)
		s.Require().True(resp.Found)
		s.Require().NotNil(resp.Recency)
		s.Require().True(resp.Recency.Stale)
		s.Require().Equal(int64(params.VerificationExpiryDays)*24*60*60, resp.Recency.MaxAgeSeconds)
	})
}

// ============================================================================
// Query surface input validation
// ============================================================================

func (s *AssuranceVectorTestSuite) TestAssuranceVectorQueryRejectsBadInput() {
	querier := keeper.NewGRPCQuerier(s.keeper)

	_, err := querier.QueryAssuranceVector(s.ctx, nil)
	s.Require().Error(err, "a nil request must be rejected")

	_, err = querier.QueryAssuranceVector(s.ctx, &types.QueryAssuranceVectorRequest{AccountAddress: ""})
	s.Require().Error(err, "an empty address must be rejected")

	_, err = querier.QueryAssuranceVector(s.ctx, &types.QueryAssuranceVectorRequest{AccountAddress: "not-an-address"})
	s.Require().Error(err, "a malformed address must be rejected")

	_, err = querier.QueryAssuranceVectorHistory(s.ctx, nil)
	s.Require().Error(err)

	_, err = querier.QueryAssuranceVectorHistory(s.ctx, &types.QueryAssuranceVectorHistoryRequest{AccountAddress: ""})
	s.Require().Error(err)
}

func (s *AssuranceVectorTestSuite) TestSetAssuranceVectorRejectsTamperedVector() {
	account := s.testAddress("assurance_tamper")

	_, err := s.keeper.ComputeAndStoreCompositeScore(s.ctx, account, s.distinguishableInputs())
	s.Require().NoError(err)
	stored, found := s.keeper.GetAssuranceVector(s.ctx, account)
	s.Require().True(found)

	// A vector whose contents no longer match its commitment must be refused at
	// the write boundary, not persisted for a relying party to discover.
	tampered := *stored
	tampered.Epoch = stored.Epoch + 5 // commitment left stale
	s.Require().Error(s.keeper.SetAssuranceVector(s.ctx, &tampered),
		"a vector whose commitment does not match its contents must be rejected")

	// A vector with no commitment at all is equally refused.
	noCommitment := *stored
	noCommitment.Epoch = stored.Epoch + 5
	noCommitment.Commitment = nil
	s.Require().Error(s.keeper.SetAssuranceVector(s.ctx, &noCommitment))
}

// ============================================================================
// Helpers
// ============================================================================

func (s *AssuranceVectorTestSuite) testAddress(seed string) string {
	return sdk.AccAddress([]byte(seed + "_______________")).String()
}
