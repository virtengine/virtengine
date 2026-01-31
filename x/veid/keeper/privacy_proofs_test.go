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
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/veid/keeper"
	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Test Suite Setup
// ============================================================================

type PrivacyProofsTestSuite struct {
	suite.Suite
	ctx           sdk.Context
	keeper        keeper.Keeper
	cdc           codec.Codec
	stateStore    store.CommitMultiStore
	subjectAddr   sdk.AccAddress
	requesterAddr sdk.AccAddress
	verifierAddr  sdk.AccAddress
}

func TestPrivacyProofsTestSuite(t *testing.T) {
	suite.Run(t, new(PrivacyProofsTestSuite))
}

func (s *PrivacyProofsTestSuite) SetupTest() {
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

	// Create test addresses
	s.subjectAddr = sdk.AccAddress([]byte("test-subject-address"))
	s.requesterAddr = sdk.AccAddress([]byte("test-requester-addr"))
	s.verifierAddr = sdk.AccAddress([]byte("test-verifier-addrs"))
}

func (s *PrivacyProofsTestSuite) createContextWithStore(storeKey *storetypes.KVStoreKey) sdk.Context {
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
func (s *PrivacyProofsTestSuite) TearDownTest() {
	CloseStoreIfNeeded(s.stateStore)
}

func (s *PrivacyProofsTestSuite) setupVerifiedIdentity(level int) {
	now := s.ctx.BlockTime()
	record := types.NewIdentityRecord(s.subjectAddr.String(), now)
	record.UpdatedAt = now

	// Set tier based on verification level
	switch level {
	case 1:
		record.Tier = types.IdentityTierBasic
		record.CurrentScore = 55
	case 2:
		record.Tier = types.IdentityTierStandard
		record.CurrentScore = 75
	default:
		record.Tier = types.IdentityTierUnverified
		record.CurrentScore = 0
	}

	err := s.keeper.SetIdentityRecord(s.ctx, *record)
	s.Require().NoError(err)
}

func (s *PrivacyProofsTestSuite) setupVerifiedIdentityWithScore(level int, score uint32) {
	s.setupVerifiedIdentity(level)
	// Use SetScore to store score in the score store (separate from IdentityRecord)
	err := s.keeper.SetScore(s.ctx, s.subjectAddr.String(), score, "v1.0.0")
	s.Require().NoError(err)
}

// ============================================================================
// Selective Disclosure Request Tests
// ============================================================================

func (s *PrivacyProofsTestSuite) TestCreateSelectiveDisclosureRequest_Success() {
	requestedClaims := []types.ClaimType{
		types.ClaimTypeAgeOver18,
		types.ClaimTypeHumanVerified,
	}

	request, err := s.keeper.CreateSelectiveDisclosureRequest(
		s.ctx,
		s.requesterAddr,
		s.subjectAddr,
		requestedClaims,
		nil,
		"age verification for marketplace",
		24*time.Hour,
		1*time.Hour,
	)

	s.Require().NoError(err)
	s.Require().NotNil(request)
	s.Assert().NotEmpty(request.RequestID)
	s.Assert().Equal(s.requesterAddr.String(), request.RequesterAddress)
	s.Assert().Equal(s.subjectAddr.String(), request.SubjectAddress)
	s.Assert().Equal(requestedClaims, request.RequestedClaims)
	s.Assert().Equal("age verification for marketplace", request.Purpose)
	s.Assert().NotNil(request.Nonce)
	s.Assert().Len(request.Nonce, 32)
}

func (s *PrivacyProofsTestSuite) TestCreateSelectiveDisclosureRequest_EmptyClaims() {
	_, err := s.keeper.CreateSelectiveDisclosureRequest(
		s.ctx,
		s.requesterAddr,
		s.subjectAddr,
		[]types.ClaimType{},
		nil,
		"test purpose",
		24*time.Hour,
		1*time.Hour,
	)

	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "requested_claims cannot be empty")
}

func (s *PrivacyProofsTestSuite) TestCreateSelectiveDisclosureRequest_EmptyPurpose() {
	requestedClaims := []types.ClaimType{types.ClaimTypeAgeOver18}

	_, err := s.keeper.CreateSelectiveDisclosureRequest(
		s.ctx,
		s.requesterAddr,
		s.subjectAddr,
		requestedClaims,
		nil,
		"",
		24*time.Hour,
		1*time.Hour,
	)

	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "purpose cannot be empty")
}

func (s *PrivacyProofsTestSuite) TestCreateSelectiveDisclosureRequest_InvalidValidityDuration() {
	requestedClaims := []types.ClaimType{types.ClaimTypeAgeOver18}

	_, err := s.keeper.CreateSelectiveDisclosureRequest(
		s.ctx,
		s.requesterAddr,
		s.subjectAddr,
		requestedClaims,
		nil,
		"test purpose",
		-1*time.Hour, // negative duration
		1*time.Hour,
	)

	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "validity_duration must be positive")
}

func (s *PrivacyProofsTestSuite) TestCreateSelectiveDisclosureRequest_WithParameters() {
	requestedClaims := []types.ClaimType{types.ClaimTypeTrustScoreAbove}
	params := map[string]interface{}{
		"threshold": 75,
	}

	request, err := s.keeper.CreateSelectiveDisclosureRequest(
		s.ctx,
		s.requesterAddr,
		s.subjectAddr,
		requestedClaims,
		params,
		"trust score verification",
		24*time.Hour,
		1*time.Hour,
	)

	s.Require().NoError(err)
	s.Require().NotNil(request)
	s.Assert().Equal(params, request.ClaimParameters)
}

// ============================================================================
// Selective Disclosure Proof Generation Tests
// ============================================================================

func (s *PrivacyProofsTestSuite) TestGenerateSelectiveDisclosureProof_Success() {
	// Setup verified identity
	s.setupVerifiedIdentity(2)

	// Create request
	requestedClaims := []types.ClaimType{types.ClaimTypeHumanVerified}
	request, err := s.keeper.CreateSelectiveDisclosureRequest(
		s.ctx,
		s.requesterAddr,
		s.subjectAddr,
		requestedClaims,
		nil,
		"human verification",
		24*time.Hour,
		1*time.Hour,
	)
	s.Require().NoError(err)

	// Generate proof
	disclosedClaims := map[string]interface{}{
		"human_verified": true,
	}
	proof, err := s.keeper.GenerateSelectiveDisclosureProof(
		s.ctx,
		s.subjectAddr,
		request,
		disclosedClaims,
		types.ProofSchemeSNARK,
	)

	s.Require().NoError(err)
	s.Require().NotNil(proof)
	s.Assert().NotEmpty(proof.ProofID)
	s.Assert().Equal(s.subjectAddr.String(), proof.SubjectAddress)
	s.Assert().Equal(requestedClaims, proof.ClaimTypes)
	s.Assert().Equal(disclosedClaims, proof.DisclosedClaims)
	s.Assert().NotNil(proof.CommitmentHash)
	s.Assert().NotNil(proof.ProofValue)
	s.Assert().Equal(types.ProofSchemeSNARK, proof.ProofScheme)
	s.Assert().NotNil(proof.Nonce)
}

func (s *PrivacyProofsTestSuite) TestGenerateSelectiveDisclosureProof_NoIdentityRecord() {
	// Don't setup identity record

	requestedClaims := []types.ClaimType{types.ClaimTypeHumanVerified}
	request, err := s.keeper.CreateSelectiveDisclosureRequest(
		s.ctx,
		s.requesterAddr,
		s.subjectAddr,
		requestedClaims,
		nil,
		"human verification",
		24*time.Hour,
		1*time.Hour,
	)
	s.Require().NoError(err)

	_, err = s.keeper.GenerateSelectiveDisclosureProof(
		s.ctx,
		s.subjectAddr,
		request,
		nil,
		types.ProofSchemeSNARK,
	)

	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "identity record not found")
}

func (s *PrivacyProofsTestSuite) TestGenerateSelectiveDisclosureProof_InsufficientVerification() {
	// Setup identity with low verification level
	s.setupVerifiedIdentity(1)

	requestedClaims := []types.ClaimType{types.ClaimTypeAgeOver18}
	request, err := s.keeper.CreateSelectiveDisclosureRequest(
		s.ctx,
		s.requesterAddr,
		s.subjectAddr,
		requestedClaims,
		nil,
		"age verification",
		24*time.Hour,
		1*time.Hour,
	)
	s.Require().NoError(err)

	_, err = s.keeper.GenerateSelectiveDisclosureProof(
		s.ctx,
		s.subjectAddr,
		request,
		nil,
		types.ProofSchemeSNARK,
	)

	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "insufficient verification level")
}

func (s *PrivacyProofsTestSuite) TestGenerateSelectiveDisclosureProof_SubjectMismatch() {
	// Setup verified identity
	s.setupVerifiedIdentity(2)

	// Create request for different subject
	otherSubject := sdk.AccAddress([]byte("other-subject-addrs"))
	requestedClaims := []types.ClaimType{types.ClaimTypeHumanVerified}
	request, err := s.keeper.CreateSelectiveDisclosureRequest(
		s.ctx,
		s.requesterAddr,
		otherSubject,
		requestedClaims,
		nil,
		"human verification",
		24*time.Hour,
		1*time.Hour,
	)
	s.Require().NoError(err)

	// Try to generate proof with different subject
	_, err = s.keeper.GenerateSelectiveDisclosureProof(
		s.ctx,
		s.subjectAddr,
		request,
		nil,
		types.ProofSchemeSNARK,
	)

	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "subject address does not match")
}

func (s *PrivacyProofsTestSuite) TestGenerateSelectiveDisclosureProof_InvalidScheme() {
	// Setup verified identity
	s.setupVerifiedIdentity(2)

	requestedClaims := []types.ClaimType{types.ClaimTypeHumanVerified}
	request, err := s.keeper.CreateSelectiveDisclosureRequest(
		s.ctx,
		s.requesterAddr,
		s.subjectAddr,
		requestedClaims,
		nil,
		"human verification",
		24*time.Hour,
		1*time.Hour,
	)
	s.Require().NoError(err)

	_, err = s.keeper.GenerateSelectiveDisclosureProof(
		s.ctx,
		s.subjectAddr,
		request,
		nil,
		types.ProofScheme(999), // Invalid scheme
	)

	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "invalid proof scheme")
}

// ============================================================================
// Proof Verification Tests
// ============================================================================

func (s *PrivacyProofsTestSuite) TestVerifySelectiveDisclosureProof_Success() {
	// Setup verified identity
	s.setupVerifiedIdentity(2)

	// Create request and generate proof
	requestedClaims := []types.ClaimType{types.ClaimTypeHumanVerified}
	request, err := s.keeper.CreateSelectiveDisclosureRequest(
		s.ctx,
		s.requesterAddr,
		s.subjectAddr,
		requestedClaims,
		nil,
		"human verification",
		24*time.Hour,
		1*time.Hour,
	)
	s.Require().NoError(err)

	disclosedClaims := map[string]interface{}{"human_verified": true}
	proof, err := s.keeper.GenerateSelectiveDisclosureProof(
		s.ctx,
		s.subjectAddr,
		request,
		disclosedClaims,
		types.ProofSchemeSNARK,
	)
	s.Require().NoError(err)

	// Verify proof
	result, err := s.keeper.VerifySelectiveDisclosureProof(
		s.ctx,
		proof,
		s.verifierAddr,
	)

	s.Require().NoError(err)
	s.Require().NotNil(result)
	s.Assert().True(result.IsValid)
	s.Assert().Equal(requestedClaims, result.ClaimsVerified)
	s.Assert().Equal(s.verifierAddr.String(), result.VerifierAddress)
	s.Assert().Empty(result.Error)
	s.Assert().NotEmpty(result.ProofHash)
}

func (s *PrivacyProofsTestSuite) TestVerifySelectiveDisclosureProof_ExpiredProof() {
	// Setup verified identity
	s.setupVerifiedIdentity(2)

	// Create request and generate proof with short validity
	requestedClaims := []types.ClaimType{types.ClaimTypeHumanVerified}
	request, err := s.keeper.CreateSelectiveDisclosureRequest(
		s.ctx,
		s.requesterAddr,
		s.subjectAddr,
		requestedClaims,
		nil,
		"human verification",
		1*time.Millisecond, // Very short validity
		1*time.Hour,
	)
	s.Require().NoError(err)

	disclosedClaims := map[string]interface{}{"human_verified": true}
	proof, err := s.keeper.GenerateSelectiveDisclosureProof(
		s.ctx,
		s.subjectAddr,
		request,
		disclosedClaims,
		types.ProofSchemeSNARK,
	)
	s.Require().NoError(err)

	// Advance block time past expiration
	expiredCtx := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(1 * time.Hour))

	// Verify proof
	result, err := s.keeper.VerifySelectiveDisclosureProof(
		expiredCtx,
		proof,
		s.verifierAddr,
	)

	s.Require().NoError(err)
	s.Require().NotNil(result)
	s.Assert().False(result.IsValid)
	s.Assert().Contains(result.Error, "expired")
}

func (s *PrivacyProofsTestSuite) TestVerifySelectiveDisclosureProof_InvalidProof() {
	// Create an invalid proof
	proof := &types.SelectiveDisclosureProof{
		ProofID:        "", // Invalid: empty ID
		SubjectAddress: s.subjectAddr.String(),
	}

	result, err := s.keeper.VerifySelectiveDisclosureProof(
		s.ctx,
		proof,
		s.verifierAddr,
	)

	s.Require().NoError(err) // Should not error, but return invalid result
	s.Require().NotNil(result)
	s.Assert().False(result.IsValid)
	s.Assert().NotEmpty(result.Error)
}

// ============================================================================
// Age Proof Tests
// ============================================================================

func (s *PrivacyProofsTestSuite) TestCreateAgeProof_Success() {
	// Setup verified identity with level 2
	s.setupVerifiedIdentity(2)

	proof, err := s.keeper.CreateAgeProof(
		s.ctx,
		s.subjectAddr,
		18,
		24*time.Hour,
	)

	s.Require().NoError(err)
	s.Require().NotNil(proof)
	s.Assert().NotEmpty(proof.ProofID)
	s.Assert().Equal(s.subjectAddr.String(), proof.SubjectAddress)
	s.Assert().Equal(uint32(18), proof.AgeThreshold)
	s.Assert().True(proof.SatisfiesThreshold)
	s.Assert().NotNil(proof.CommitmentHash)
	s.Assert().NotNil(proof.ProofValue)
	s.Assert().NotNil(proof.Nonce)
}

func (s *PrivacyProofsTestSuite) TestCreateAgeProof_InsufficientVerification() {
	// Setup identity with low verification level
	s.setupVerifiedIdentity(1)

	_, err := s.keeper.CreateAgeProof(
		s.ctx,
		s.subjectAddr,
		18,
		24*time.Hour,
	)

	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "verification level")
}

func (s *PrivacyProofsTestSuite) TestCreateAgeProof_NoIdentityRecord() {
	_, err := s.keeper.CreateAgeProof(
		s.ctx,
		s.subjectAddr,
		18,
		24*time.Hour,
	)

	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "identity record not found")
}

func (s *PrivacyProofsTestSuite) TestCreateAgeProof_InvalidThreshold() {
	s.setupVerifiedIdentity(2)

	// Test zero threshold
	_, err := s.keeper.CreateAgeProof(
		s.ctx,
		s.subjectAddr,
		0,
		24*time.Hour,
	)
	s.Require().Error(err)

	// Test threshold > 150
	_, err = s.keeper.CreateAgeProof(
		s.ctx,
		s.subjectAddr,
		151,
		24*time.Hour,
	)
	s.Require().Error(err)
}

// ============================================================================
// Residency Proof Tests
// ============================================================================

func (s *PrivacyProofsTestSuite) TestCreateResidencyProof_Success() {
	// Setup verified identity with level 2
	s.setupVerifiedIdentity(2)

	proof, err := s.keeper.CreateResidencyProof(
		s.ctx,
		s.subjectAddr,
		"US",
		24*time.Hour,
	)

	s.Require().NoError(err)
	s.Require().NotNil(proof)
	s.Assert().NotEmpty(proof.ProofID)
	s.Assert().Equal(s.subjectAddr.String(), proof.SubjectAddress)
	s.Assert().Equal("US", proof.CountryCode)
	s.Assert().True(proof.IsResident)
	s.Assert().NotNil(proof.CommitmentHash)
	s.Assert().NotNil(proof.ProofValue)
	s.Assert().NotNil(proof.Nonce)
}

func (s *PrivacyProofsTestSuite) TestCreateResidencyProof_InvalidCountryCode() {
	s.setupVerifiedIdentity(2)

	// Test 3-letter code
	_, err := s.keeper.CreateResidencyProof(
		s.ctx,
		s.subjectAddr,
		"USA",
		24*time.Hour,
	)
	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "ISO 3166-1 alpha-2")

	// Test 1-letter code
	_, err = s.keeper.CreateResidencyProof(
		s.ctx,
		s.subjectAddr,
		"U",
		24*time.Hour,
	)
	s.Require().Error(err)
}

func (s *PrivacyProofsTestSuite) TestCreateResidencyProof_InsufficientVerification() {
	s.setupVerifiedIdentity(1)

	_, err := s.keeper.CreateResidencyProof(
		s.ctx,
		s.subjectAddr,
		"US",
		24*time.Hour,
	)

	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "verification level")
}

// ============================================================================
// Score Threshold Proof Tests
// ============================================================================

func (s *PrivacyProofsTestSuite) TestCreateScoreThresholdProof_Success() {
	// Setup verified identity with score
	s.setupVerifiedIdentityWithScore(2, 85)

	proof, err := s.keeper.CreateScoreThresholdProof(
		s.ctx,
		s.subjectAddr,
		75,
		24*time.Hour,
	)

	s.Require().NoError(err)
	s.Require().NotNil(proof)
	s.Assert().NotEmpty(proof.ProofID)
	s.Assert().Equal(s.subjectAddr.String(), proof.SubjectAddress)
	s.Assert().Equal(uint32(75), proof.ScoreThreshold)
	s.Assert().True(proof.ExceedsThreshold)
	s.Assert().NotNil(proof.CommitmentHash)
	s.Assert().NotNil(proof.ProofValue)
	s.Assert().NotNil(proof.Nonce)
	s.Assert().Equal("v1.0.0", proof.ScoreVersion)
}

func (s *PrivacyProofsTestSuite) TestCreateScoreThresholdProof_BelowThreshold() {
	// Setup verified identity with lower score
	s.setupVerifiedIdentityWithScore(2, 50)

	proof, err := s.keeper.CreateScoreThresholdProof(
		s.ctx,
		s.subjectAddr,
		75,
		24*time.Hour,
	)

	s.Require().NoError(err)
	s.Require().NotNil(proof)
	s.Assert().False(proof.ExceedsThreshold)
}

func (s *PrivacyProofsTestSuite) TestCreateScoreThresholdProof_NoScore() {
	// Setup identity without score
	s.setupVerifiedIdentity(2)

	_, err := s.keeper.CreateScoreThresholdProof(
		s.ctx,
		s.subjectAddr,
		75,
		24*time.Hour,
	)

	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "no verified score")
}

func (s *PrivacyProofsTestSuite) TestCreateScoreThresholdProof_InvalidThreshold() {
	s.setupVerifiedIdentityWithScore(2, 85)

	// Test zero threshold
	_, err := s.keeper.CreateScoreThresholdProof(
		s.ctx,
		s.subjectAddr,
		0,
		24*time.Hour,
	)
	s.Require().Error(err)

	// Test threshold > 100
	_, err = s.keeper.CreateScoreThresholdProof(
		s.ctx,
		s.subjectAddr,
		101,
		24*time.Hour,
	)
	s.Require().Error(err)
}

// ============================================================================
// Claim Type Tests
// ============================================================================

func (s *PrivacyProofsTestSuite) TestClaimType_Validation() {
	tests := []struct {
		claimType types.ClaimType
		valid     bool
	}{
		{types.ClaimTypeAgeOver18, true},
		{types.ClaimTypeAgeOver21, true},
		{types.ClaimTypeAgeOver25, true},
		{types.ClaimTypeCountryResident, true},
		{types.ClaimTypeHumanVerified, true},
		{types.ClaimTypeTrustScoreAbove, true},
		{types.ClaimTypeEmailVerified, true},
		{types.ClaimTypeSMSVerified, true},
		{types.ClaimTypeDomainVerified, true},
		{types.ClaimTypeBiometricVerified, true},
		{types.ClaimType(100), false},
		{types.ClaimType(-1), false},
	}

	for _, tc := range tests {
		s.Assert().Equal(tc.valid, tc.claimType.IsValid(), "claim type %d", tc.claimType)
	}
}

func (s *PrivacyProofsTestSuite) TestClaimType_String() {
	tests := []struct {
		claimType types.ClaimType
		expected  string
	}{
		{types.ClaimTypeAgeOver18, "age_over_18"},
		{types.ClaimTypeAgeOver21, "age_over_21"},
		{types.ClaimTypeAgeOver25, "age_over_25"},
		{types.ClaimTypeCountryResident, "country_resident"},
		{types.ClaimTypeHumanVerified, "human_verified"},
		{types.ClaimTypeTrustScoreAbove, "trust_score_above"},
		{types.ClaimTypeEmailVerified, "email_verified"},
		{types.ClaimTypeSMSVerified, "sms_verified"},
		{types.ClaimTypeDomainVerified, "domain_verified"},
		{types.ClaimTypeBiometricVerified, "biometric_verified"},
		{types.ClaimType(100), "unknown"},
	}

	for _, tc := range tests {
		s.Assert().Equal(tc.expected, tc.claimType.String())
	}
}

func (s *PrivacyProofsTestSuite) TestClaimType_Parse() {
	tests := []struct {
		input     string
		expected  types.ClaimType
		shouldErr bool
	}{
		{"age_over_18", types.ClaimTypeAgeOver18, false},
		{"age_over_21", types.ClaimTypeAgeOver21, false},
		{"age_over_25", types.ClaimTypeAgeOver25, false},
		{"country_resident", types.ClaimTypeCountryResident, false},
		{"human_verified", types.ClaimTypeHumanVerified, false},
		{"trust_score_above", types.ClaimTypeTrustScoreAbove, false},
		{"email_verified", types.ClaimTypeEmailVerified, false},
		{"sms_verified", types.ClaimTypeSMSVerified, false},
		{"domain_verified", types.ClaimTypeDomainVerified, false},
		{"biometric_verified", types.ClaimTypeBiometricVerified, false},
		{"invalid", 0, true},
		{"", 0, true},
	}

	for _, tc := range tests {
		result, err := types.ParseClaimType(tc.input)
		if tc.shouldErr {
			s.Assert().Error(err, "input: %s", tc.input)
		} else {
			s.Assert().NoError(err, "input: %s", tc.input)
			s.Assert().Equal(tc.expected, result)
		}
	}
}

// ============================================================================
// Proof Scheme Tests
// ============================================================================

func (s *PrivacyProofsTestSuite) TestProofScheme_Validation() {
	tests := []struct {
		scheme types.ProofScheme
		valid  bool
	}{
		{types.ProofSchemeSNARK, true},
		{types.ProofSchemeSTARK, true},
		{types.ProofSchemeBulletproofs, true},
		{types.ProofSchemeRangeProof, true},
		{types.ProofSchemeCommitmentScheme, true},
		{types.ProofScheme(100), false},
		{types.ProofScheme(-1), false},
	}

	for _, tc := range tests {
		s.Assert().Equal(tc.valid, tc.scheme.IsValid(), "scheme %d", tc.scheme)
	}
}

func (s *PrivacyProofsTestSuite) TestProofScheme_String() {
	tests := []struct {
		scheme   types.ProofScheme
		expected string
	}{
		{types.ProofSchemeSNARK, "snark"},
		{types.ProofSchemeSTARK, "stark"},
		{types.ProofSchemeBulletproofs, "bulletproofs"},
		{types.ProofSchemeRangeProof, "range_proof"},
		{types.ProofSchemeCommitmentScheme, "commitment_scheme"},
		{types.ProofScheme(100), "unknown"},
	}

	for _, tc := range tests {
		s.Assert().Equal(tc.expected, tc.scheme.String())
	}
}

func (s *PrivacyProofsTestSuite) TestProofScheme_Parse() {
	tests := []struct {
		input     string
		expected  types.ProofScheme
		shouldErr bool
	}{
		{"snark", types.ProofSchemeSNARK, false},
		{"stark", types.ProofSchemeSTARK, false},
		{"bulletproofs", types.ProofSchemeBulletproofs, false},
		{"range_proof", types.ProofSchemeRangeProof, false},
		{"commitment_scheme", types.ProofSchemeCommitmentScheme, false},
		{"invalid", 0, true},
		{"", 0, true},
	}

	for _, tc := range tests {
		result, err := types.ParseProofScheme(tc.input)
		if tc.shouldErr {
			s.Assert().Error(err, "input: %s", tc.input)
		} else {
			s.Assert().NoError(err, "input: %s", tc.input)
			s.Assert().Equal(tc.expected, result)
		}
	}
}

// ============================================================================
// Expiration Handling Tests
// ============================================================================

func (s *PrivacyProofsTestSuite) TestSelectiveDisclosureRequest_Expiration() {
	request := types.NewSelectiveDisclosureRequest(
		"req_test",
		s.requesterAddr.String(),
		s.subjectAddr.String(),
		[]types.ClaimType{types.ClaimTypeHumanVerified},
		"test",
		24*time.Hour,
		1*time.Hour,
	)

	// Should not be expired initially
	s.Assert().False(request.IsExpired(time.Now()))

	// Should be expired after expiry time
	s.Assert().True(request.IsExpired(time.Now().Add(2 * time.Hour)))
}

func (s *PrivacyProofsTestSuite) TestAgeProof_Expiration() {
	proof := types.NewAgeProof("proof_test", s.subjectAddr.String(), 18, 1*time.Hour)

	// Should not be expired initially
	s.Assert().False(proof.IsExpired(time.Now()))

	// Should be expired after validity
	s.Assert().True(proof.IsExpired(time.Now().Add(2 * time.Hour)))
}

func (s *PrivacyProofsTestSuite) TestResidencyProof_Expiration() {
	proof := types.NewResidencyProof("proof_test", s.subjectAddr.String(), "US", 1*time.Hour)

	// Should not be expired initially
	s.Assert().False(proof.IsExpired(time.Now()))

	// Should be expired after validity
	s.Assert().True(proof.IsExpired(time.Now().Add(2 * time.Hour)))
}

func (s *PrivacyProofsTestSuite) TestScoreThresholdProof_Expiration() {
	proof := types.NewScoreThresholdProof("proof_test", s.subjectAddr.String(), 75, 1*time.Hour)

	// Should not be expired initially
	s.Assert().False(proof.IsExpired(time.Now()))

	// Should be expired after validity
	s.Assert().True(proof.IsExpired(time.Now().Add(2 * time.Hour)))
}

// ============================================================================
// Helper Function Tests
// ============================================================================

func (s *PrivacyProofsTestSuite) TestGenerateProofID() {
	claimTypes := []types.ClaimType{types.ClaimTypeAgeOver18}
	nonce := []byte("test-nonce-12345")

	proofID := types.GenerateProofID(s.subjectAddr.String(), claimTypes, nonce)

	s.Assert().NotEmpty(proofID)
	s.Assert().True(len(proofID) > 0)
	s.Assert().Contains(proofID, "proof_")

	// Same inputs should generate same ID
	proofID2 := types.GenerateProofID(s.subjectAddr.String(), claimTypes, nonce)
	s.Assert().Equal(proofID, proofID2)

	// Different nonce should generate different ID
	differentNonce := []byte("different-nonce-")
	proofID3 := types.GenerateProofID(s.subjectAddr.String(), claimTypes, differentNonce)
	s.Assert().NotEqual(proofID, proofID3)
}

func (s *PrivacyProofsTestSuite) TestGenerateRequestID() {
	nonce := []byte("test-nonce-12345")

	requestID := types.GenerateRequestID(s.requesterAddr.String(), s.subjectAddr.String(), nonce)

	s.Assert().NotEmpty(requestID)
	s.Assert().Contains(requestID, "req_")

	// Same inputs should generate same ID
	requestID2 := types.GenerateRequestID(s.requesterAddr.String(), s.subjectAddr.String(), nonce)
	s.Assert().Equal(requestID, requestID2)
}

func (s *PrivacyProofsTestSuite) TestComputeCommitmentHash() {
	salt := []byte("test-salt-1234567890123456")

	hash1, err := types.ComputeCommitmentHash("value1", salt)
	s.Require().NoError(err)
	s.Assert().NotNil(hash1)
	s.Assert().Len(hash1, 32)

	// Same inputs should produce same hash
	hash2, err := types.ComputeCommitmentHash("value1", salt)
	s.Require().NoError(err)
	s.Assert().Equal(hash1, hash2)

	// Different value should produce different hash
	hash3, err := types.ComputeCommitmentHash("value2", salt)
	s.Require().NoError(err)
	s.Assert().NotEqual(hash1, hash3)

	// Different salt should produce different hash
	differentSalt := []byte("different-salt-123456")
	hash4, err := types.ComputeCommitmentHash("value1", differentSalt)
	s.Require().NoError(err)
	s.Assert().NotEqual(hash1, hash4)
}

// ============================================================================
// Standalone Unit Tests
// ============================================================================

func TestClaimTypeValidation(t *testing.T) {
	require.True(t, types.ClaimTypeAgeOver18.IsValid())
	require.True(t, types.ClaimTypeBiometricVerified.IsValid())
	require.False(t, types.ClaimType(999).IsValid())
}

func TestProofSchemeValidation(t *testing.T) {
	require.True(t, types.ProofSchemeSNARK.IsValid())
	require.True(t, types.ProofSchemeCommitmentScheme.IsValid())
	require.False(t, types.ProofScheme(999).IsValid())
}

func TestSelectiveDisclosureProofValidation(t *testing.T) {
	now := time.Now().UTC()
	validProof := &types.SelectiveDisclosureProof{
		ProofID:        "proof_test123",
		SubjectAddress: "cosmos1abc",
		ClaimTypes:     []types.ClaimType{types.ClaimTypeAgeOver18},
		ProofValue:     []byte("proof-value"),
		CommitmentHash: []byte("commitment"),
		Nonce:          []byte("nonce"),
		ProofScheme:    types.ProofSchemeSNARK,
		CreatedAt:      now,
		ValidUntil:     now.Add(24 * time.Hour),
	}

	require.NoError(t, validProof.Validate())

	// Test empty proof ID
	invalidProof := *validProof
	invalidProof.ProofID = ""
	require.Error(t, invalidProof.Validate())

	// Test empty claim types
	invalidProof = *validProof
	invalidProof.ClaimTypes = []types.ClaimType{}
	require.Error(t, invalidProof.Validate())

	// Test empty proof value
	invalidProof = *validProof
	invalidProof.ProofValue = []byte{}
	require.Error(t, invalidProof.Validate())
}

func TestSelectiveDisclosureRequestValidation(t *testing.T) {
	now := time.Now().UTC()
	validRequest := &types.SelectiveDisclosureRequest{
		RequestID:        "req_test123",
		RequesterAddress: "cosmos1requester",
		SubjectAddress:   "cosmos1subject",
		RequestedClaims:  []types.ClaimType{types.ClaimTypeAgeOver18},
		Purpose:          "age verification",
		ValidityDuration: 24 * time.Hour,
		CreatedAt:        now,
		ExpiresAt:        now.Add(1 * time.Hour),
	}

	require.NoError(t, validRequest.Validate())

	// Test empty purpose
	invalidRequest := *validRequest
	invalidRequest.Purpose = ""
	require.Error(t, invalidRequest.Validate())

	// Test invalid validity duration
	invalidRequest = *validRequest
	invalidRequest.ValidityDuration = -1 * time.Hour
	require.Error(t, invalidRequest.Validate())
}

func TestAgeProofValidation(t *testing.T) {
	now := time.Now().UTC()
	validProof := &types.AgeProof{
		ProofID:        "proof_test123",
		SubjectAddress: "cosmos1abc",
		AgeThreshold:   18,
		ProofValue:     []byte("proof-value"),
		CommitmentHash: []byte("commitment"),
		Nonce:          []byte("nonce"),
		ProofScheme:    types.ProofSchemeRangeProof,
		CreatedAt:      now,
		ValidUntil:     now.Add(24 * time.Hour),
	}

	require.NoError(t, validProof.Validate())

	// Test zero age threshold
	invalidProof := *validProof
	invalidProof.AgeThreshold = 0
	require.Error(t, invalidProof.Validate())
}

func TestResidencyProofValidation(t *testing.T) {
	now := time.Now().UTC()
	validProof := &types.ResidencyProof{
		ProofID:        "proof_test123",
		SubjectAddress: "cosmos1abc",
		CountryCode:    "US",
		ProofValue:     []byte("proof-value"),
		CommitmentHash: []byte("commitment"),
		Nonce:          []byte("nonce"),
		ProofScheme:    types.ProofSchemeCommitmentScheme,
		CreatedAt:      now,
		ValidUntil:     now.Add(24 * time.Hour),
	}

	require.NoError(t, validProof.Validate())

	// Test invalid country code length
	invalidProof := *validProof
	invalidProof.CountryCode = "USA"
	require.Error(t, invalidProof.Validate())

	// Test empty country code
	invalidProof = *validProof
	invalidProof.CountryCode = ""
	require.Error(t, invalidProof.Validate())
}

func TestScoreThresholdProofValidation(t *testing.T) {
	now := time.Now().UTC()
	validProof := &types.ScoreThresholdProof{
		ProofID:        "proof_test123",
		SubjectAddress: "cosmos1abc",
		ScoreThreshold: 75,
		ProofValue:     []byte("proof-value"),
		CommitmentHash: []byte("commitment"),
		Nonce:          []byte("nonce"),
		ProofScheme:    types.ProofSchemeBulletproofs,
		CreatedAt:      now,
		ValidUntil:     now.Add(24 * time.Hour),
	}

	require.NoError(t, validProof.Validate())

	// Test zero score threshold
	invalidProof := *validProof
	invalidProof.ScoreThreshold = 0
	require.Error(t, invalidProof.Validate())

	// Test score threshold > 100
	invalidProof = *validProof
	invalidProof.ScoreThreshold = 101
	require.Error(t, invalidProof.Validate())
}
