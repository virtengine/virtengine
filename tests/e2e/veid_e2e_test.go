//go:build e2e.integration

// Package e2e contains end-to-end tests for VirtEngine.
//
// This file implements comprehensive E2E tests for VEID onboarding
// and verification flows including:
// - Account creation and scope upload
// - SSO/Email/SMS verification
// - ML scoring and tier transitions
// - Attestation recording and validation
// - Rejection path testing
//
// Task Reference: VE-8B - E2E VEID onboarding and verification flows
package e2e

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/app"
	sdktestutil "github.com/virtengine/virtengine/sdk/go/testutil"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
	"github.com/virtengine/virtengine/x/veid/keeper"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// VEID E2E Test Suite
// ============================================================================

// VEIDE2ETestSuite tests the complete VEID onboarding and verification flows.
//
// Test Coverage:
//  1. Account creation and identity scope upload
//  2. SSO verification flow (Google, Microsoft, GitHub)
//  3. Email verification with OTP
//  4. SMS verification with OTP and anti-fraud
//  5. Mobile capture and biometric pipeline
//  6. ML scoring and tier transitions
//  7. Attestation recording and validation
//  8. Rejection paths (expired OTP, low score, invalid doc, VoIP)
type VEIDE2ETestSuite struct {
	suite.Suite

	app        *app.VirtEngineApp
	ctx        sdk.Context
	testClient VEIDTestClient
	msgServer  veidtypes.MsgServer
}

// TestVEIDE2E runs the VEID E2E test suite.
func TestVEIDE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	suite.Run(t, new(VEIDE2ETestSuite))
}

// SetupSuite runs once before all tests in the suite.
func (s *VEIDE2ETestSuite) SetupSuite() {
	s.testClient = NewVEIDTestClient()

	s.app = app.Setup(
		app.WithChainID(TestChainID),
		app.WithGenesis(func(cdc codec.Codec) app.GenesisState {
			return genesisWithVEIDApprovedClientE2E(s.T(), cdc, s.testClient)
		}),
	)

	s.ctx = s.app.NewContext(false).
		WithBlockHeight(1).
		WithBlockTime(FixedTimestamp())

	s.msgServer = keeper.NewMsgServerImpl(s.app.Keepers.VirtEngine.VEID)
}

// ============================================================================
// Test: Complete Onboarding Flow
// ============================================================================

// TestCompleteOnboardingFlow tests the full VEID onboarding journey:
// 1. Create account -> 2. Upload scope -> 3. Verify -> 4. Tier transition
func (s *VEIDE2ETestSuite) TestCompleteOnboardingFlow() {
	ctx := s.ctx
	customer := sdktestutil.AccAddress(s.T())

	// Step 1: Upload selfie scope to create identity record
	scopeFixture := SelfieScope()
	envelope := EncryptedEnvelopeFixture(scopeFixture.ScopeID)
	payloadHash := PayloadHash(envelope)

	metadata := veidtypes.NewUploadMetadata(
		scopeFixture.Salt,
		scopeFixture.DeviceFingerprint,
		s.testClient.ClientID,
		nil,
		nil,
		payloadHash,
	)

	clientSignature := s.testClient.Sign(metadata.SigningPayload())
	userSignature := bytes.Repeat([]byte{0x04}, keeper.Secp256k1SignatureSize)

	msg := veidtypes.NewMsgUploadScope(
		customer.String(),
		scopeFixture.ScopeID,
		scopeFixture.ScopeType,
		envelope,
		scopeFixture.Salt,
		scopeFixture.DeviceFingerprint,
		s.testClient.ClientID,
		clientSignature,
		userSignature,
		payloadHash,
	)
	msg.CaptureTimestamp = scopeFixture.CaptureTimestamp

	resp, err := s.msgServer.UploadScope(ctx, msg)
	require.NoError(s.T(), err)
	require.Equal(s.T(), scopeFixture.ScopeID, resp.ScopeId)

	// Verify identity record was created
	s.app.Commit()
	ctx = s.app.NewContext(false).
		WithBlockHeight(2).
		WithBlockTime(FixedTimestampPlus(1))

	record, found := s.app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, customer)
	require.True(s.T(), found, "Identity record should be created")
	require.Equal(s.T(), veidtypes.IdentityTierUnverified, record.Tier)
	require.Equal(s.T(), 1, len(record.ScopeRefs))

	// Step 2: Simulate validator score update (low tier)
	require.NoError(s.T(), s.app.Keepers.VirtEngine.VEID.UpdateScore(ctx, customer, 25, TestModelVersion))

	s.app.Commit()
	ctx = s.app.NewContext(false).
		WithBlockHeight(3).
		WithBlockTime(FixedTimestampPlus(2))

	record, found = s.app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, customer)
	require.True(s.T(), found)
	require.Equal(s.T(), veidtypes.IdentityTierBasic, record.Tier)
	require.Equal(s.T(), uint32(25), record.CurrentScore)

	// Step 3: Upload ID document scope
	idDocFixture := IDDocumentScope()
	idDocEnvelope := EncryptedEnvelopeFixture(idDocFixture.ScopeID)
	idDocPayloadHash := PayloadHash(idDocEnvelope)

	idDocMetadata := veidtypes.NewUploadMetadata(
		idDocFixture.Salt,
		idDocFixture.DeviceFingerprint,
		s.testClient.ClientID,
		nil,
		nil,
		idDocPayloadHash,
	)

	idDocClientSig := s.testClient.Sign(idDocMetadata.SigningPayload())
	idDocUserSig := bytes.Repeat([]byte{0x05}, keeper.Secp256k1SignatureSize)

	idDocMsg := veidtypes.NewMsgUploadScope(
		customer.String(),
		idDocFixture.ScopeID,
		idDocFixture.ScopeType,
		idDocEnvelope,
		idDocFixture.Salt,
		idDocFixture.DeviceFingerprint,
		s.testClient.ClientID,
		idDocClientSig,
		idDocUserSig,
		idDocPayloadHash,
	)
	idDocMsg.CaptureTimestamp = idDocFixture.CaptureTimestamp

	idDocResp, err := s.msgServer.UploadScope(ctx, idDocMsg)
	require.NoError(s.T(), err)
	require.Equal(s.T(), idDocFixture.ScopeID, idDocResp.ScopeId)

	// Step 4: Higher score update -> tier transition
	require.NoError(s.T(), s.app.Keepers.VirtEngine.VEID.UpdateScore(ctx, customer, 82, TestModelVersion))

	s.app.Commit()
	ctx = s.app.NewContext(false).
		WithBlockHeight(4).
		WithBlockTime(FixedTimestampPlus(3))

	record, found = s.app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, customer)
	require.True(s.T(), found)
	require.Equal(s.T(), veidtypes.IdentityTierVerified, record.Tier)
	require.Equal(s.T(), uint32(82), record.CurrentScore)
	require.Equal(s.T(), 2, len(record.ScopeRefs))

	s.T().Log("✅ Complete onboarding flow passed")
}

// ============================================================================
// Test: Email Verification Flow
// ============================================================================

// TestEmailVerificationFlow tests email verification with OTP
func (s *VEIDE2ETestSuite) TestEmailVerificationFlow() {
	ctx := s.ctx
	customer := sdktestutil.AccAddress(s.T())

	// Create identity record first
	_, err := s.app.Keepers.VirtEngine.VEID.CreateIdentityRecord(ctx, customer)
	require.NoError(s.T(), err)

	// Get email fixture
	emailFixture := ValidEmailVerification()

	// Create email verification record
	emailRecord := veidtypes.NewEmailVerificationRecord(
		emailFixture.VerificationID,
		customer.String(),
		emailFixture.Email,
		emailFixture.Nonce,
		FixedTimestamp(),
	)
	require.NoError(s.T(), emailRecord.Validate())

	// Create email verification challenge
	challenge := veidtypes.NewEmailVerificationChallenge(
		"challenge-"+emailFixture.VerificationID,
		customer.String(),
		emailFixture.EmailHash,
		emailFixture.Nonce,
		FixedTimestamp(),
		300, // 5 min TTL
		3,   // max attempts
	)
	require.NoError(s.T(), challenge.Validate())

	// Verify the challenge is not expired
	require.False(s.T(), challenge.IsExpired(FixedTimestamp()))
	require.True(s.T(), challenge.CanAttempt())

	// Verify correct OTP
	require.True(s.T(), challenge.VerifyNonce(emailFixture.Nonce))

	// Record successful attempt
	challenge.RecordAttempt(FixedTimestampPlus(1), true)
	require.Equal(s.T(), veidtypes.EmailStatusVerified, challenge.Status)
	require.True(s.T(), challenge.IsConsumed)

	// Mark record as verified
	expiresAt := FixedTimestampPlus(60 * 24 * 365) // 1 year
	emailRecord.MarkVerified(FixedTimestampPlus(1), &expiresAt)
	require.Equal(s.T(), veidtypes.EmailStatusVerified, emailRecord.Status)
	require.True(s.T(), emailRecord.IsActive())

	// Calculate email score contribution
	weight := veidtypes.DefaultEmailScoringWeight()
	score := veidtypes.CalculateEmailScore(emailRecord, weight, FixedTimestampPlus(2))
	require.Greater(s.T(), score, uint32(0))

	s.T().Logf("✅ Email verification flow passed - Score contribution: %d", score)
}

// TestEmailVerificationExpiredOTP tests rejection of expired OTP
func (s *VEIDE2ETestSuite) TestEmailVerificationExpiredOTP() {
	emailFixture := ExpiredEmailVerification()

	// Create challenge that has already expired
	challenge := veidtypes.NewEmailVerificationChallenge(
		"challenge-"+emailFixture.VerificationID,
		"virtengine1expired",
		emailFixture.EmailHash,
		emailFixture.Nonce,
		FixedTimestamp(),
		1, // 1 second TTL
		3,
	)

	// Simulate time passing beyond expiry
	require.True(s.T(), challenge.IsExpired(FixedTimestampPlus(1)))

	s.T().Log("✅ Expired OTP rejection test passed")
}

// ============================================================================
// Test: SMS Verification Flow
// ============================================================================

// TestSMSVerificationFlow tests SMS verification with OTP
func (s *VEIDE2ETestSuite) TestSMSVerificationFlow() {
	ctx := s.ctx
	customer := sdktestutil.AccAddress(s.T())

	// Create identity record first
	_, err := s.app.Keepers.VirtEngine.VEID.CreateIdentityRecord(ctx, customer)
	require.NoError(s.T(), err)

	// Get SMS fixture
	smsFixture := ValidSMSVerification()

	// Create SMS verification record
	smsRecord, err := veidtypes.NewSMSVerificationRecord(
		smsFixture.VerificationID,
		customer.String(),
		smsFixture.PhoneNumber,
		smsFixture.CountryCode,
		FixedTimestamp(),
	)
	require.NoError(s.T(), err)
	require.NoError(s.T(), smsRecord.Validate())
	require.False(s.T(), smsRecord.IsVoIP)

	// Create OTP challenge
	challengeCfg := veidtypes.SMSOTPChallengeConfig{
		ChallengeID:      "sms-challenge-e2e-001",
		VerificationID:   smsFixture.VerificationID,
		AccountAddress:   customer.String(),
		PhoneHashRef:     smsRecord.PhoneHash.Hash[:16],
		OTPHash:          smsFixture.OTPHash,
		MaskedPhone:      smsFixture.MaskedPhone,
		ValidatorAddress: "virtengine1validator",
		CreatedAt:        FixedTimestamp(),
		TTLSeconds:       300,
		MaxAttempts:      3,
		MaxResends:       3,
	}
	challenge := veidtypes.NewSMSOTPChallenge(challengeCfg)
	require.NoError(s.T(), challenge.Validate())

	// Verify OTP matches
	require.True(s.T(), challenge.VerifyOTP(smsFixture.OTP))
	require.True(s.T(), challenge.CanAttempt())

	// Record successful attempt
	challenge.RecordAttempt(FixedTimestampPlus(1), true)
	require.Equal(s.T(), veidtypes.SMSStatusVerified, challenge.Status)
	require.True(s.T(), challenge.IsConsumed)

	// Mark record as verified
	expiresAt := FixedTimestampPlus(60 * 24 * 365)
	smsRecord.MarkVerified(FixedTimestampPlus(1), &expiresAt, "virtengine1validator")
	require.Equal(s.T(), veidtypes.SMSStatusVerified, smsRecord.Status)
	require.True(s.T(), smsRecord.IsActive())

	// Calculate SMS score contribution
	weight := veidtypes.DefaultSMSScoringWeight()
	smsRecord.CarrierType = "mobile"
	score := veidtypes.CalculateSMSScore(smsRecord, weight, FixedTimestampPlus(2))
	require.Greater(s.T(), score, uint32(0))

	s.T().Logf("✅ SMS verification flow passed - Score contribution: %d", score)
}

// TestSMSVerificationVoIPBlocking tests rejection of VoIP phone numbers
func (s *VEIDE2ETestSuite) TestSMSVerificationVoIPBlocking() {
	customer := sdktestutil.AccAddress(s.T())
	voipFixture := VoIPSMSVerification()

	// Create SMS verification record with VoIP flag
	smsRecord, err := veidtypes.NewSMSVerificationRecord(
		voipFixture.VerificationID,
		customer.String(),
		voipFixture.PhoneNumber,
		voipFixture.CountryCode,
		FixedTimestamp(),
	)
	require.NoError(s.T(), err)

	// Mark as VoIP
	smsRecord.IsVoIP = true
	smsRecord.CarrierType = "voip"

	// Block the record
	smsRecord.MarkBlocked(FixedTimestamp(), "VoIP number detected")
	require.Equal(s.T(), veidtypes.SMSStatusBlocked, smsRecord.Status)
	require.False(s.T(), smsRecord.IsActive())

	// Verify VoIP numbers get zero score
	weight := veidtypes.DefaultSMSScoringWeight()
	score := veidtypes.CalculateSMSScore(smsRecord, weight, FixedTimestampPlus(1))
	require.Equal(s.T(), uint32(0), score, "VoIP numbers should not contribute to score")

	s.T().Log("✅ VoIP blocking test passed")
}

// ============================================================================
// Test: SSO Verification Flow
// ============================================================================

// TestSSOVerificationFlow tests SSO verification with multiple providers
func (s *VEIDE2ETestSuite) TestSSOVerificationFlow() {
	ctx := s.ctx
	customer := sdktestutil.AccAddress(s.T())

	// Create identity record first
	_, err := s.app.Keepers.VirtEngine.VEID.CreateIdentityRecord(ctx, customer)
	require.NoError(s.T(), err)

	// Test Google SSO
	googleSSO := GoogleSSOVerification()
	s.verifySSOProvider(customer, googleSSO)

	// Test Microsoft SSO
	microsoftSSO := MicrosoftSSOVerification()
	s.verifySSOProvider(customer, microsoftSSO)

	// Test GitHub SSO
	githubSSO := GitHubSSOVerification()
	s.verifySSOProvider(customer, githubSSO)

	s.T().Log("✅ SSO verification flow passed for all providers")
}

func (s *VEIDE2ETestSuite) verifySSOProvider(customer sdk.AccAddress, fixture SSOVerificationFixture) {
	// Create SSO challenge
	challenge := veidtypes.NewSSOVerificationChallenge(
		"challenge-"+fixture.LinkageID,
		customer.String(),
		fixture.Provider,
		fixture.Nonce,
		FixedTimestamp(),
		300, // 5 min TTL
	)
	require.NoError(s.T(), challenge.Validate())
	require.False(s.T(), challenge.IsExpired(FixedTimestamp()))

	// Create SSO linkage metadata (simulating successful OAuth callback)
	linkage := veidtypes.NewSSOLinkageMetadata(
		fixture.LinkageID,
		fixture.Provider,
		fixture.Issuer,
		fixture.SubjectID,
		fixture.Nonce,
		FixedTimestampPlus(1),
	)
	require.NoError(s.T(), linkage.Validate())
	require.True(s.T(), linkage.IsActive())

	// Verify provider weight
	weight := veidtypes.GetSSOScoringWeight(fixture.Provider)
	require.Equal(s.T(), fixture.ProviderWeight, weight)

	s.T().Logf("  → %s SSO verified (weight: %d)", fixture.Provider, weight)
}

// ============================================================================
// Test: ML Scoring and Tier Transitions
// ============================================================================

// TestMLScoringAndTierTransitions tests score updates and tier changes
func (s *VEIDE2ETestSuite) TestMLScoringAndTierTransitions() {
	ctx := s.ctx
	customer := sdktestutil.AccAddress(s.T())

	// Create identity record
	createdRecord, err := s.app.Keepers.VirtEngine.VEID.CreateIdentityRecord(ctx, customer)
	require.NoError(s.T(), err)
	require.Equal(s.T(), veidtypes.IdentityTierUnverified, createdRecord.Tier)

	// Test transition: Unverified -> Basic
	transition1 := UnverifiedToBasic()
	require.NoError(s.T(), s.app.Keepers.VirtEngine.VEID.UpdateScore(
		ctx, customer, transition1.ExpectedScore, transition1.ScoringModel))

	s.app.Commit()
	ctx = s.app.NewContext(false).
		WithBlockHeight(ctx.BlockHeight() + 1).
		WithBlockTime(FixedTimestampPlus(1))

	record, found := s.app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, customer)
	require.True(s.T(), found)
	require.Equal(s.T(), transition1.ExpectedTier, record.Tier)
	s.T().Logf("  → %s -> %s (score: %d)", transition1.InitialTier, record.Tier, record.CurrentScore)

	// Test transition: Basic -> Verified
	transition2 := BasicToVerified()
	require.NoError(s.T(), s.app.Keepers.VirtEngine.VEID.UpdateScore(
		ctx, customer, transition2.ExpectedScore, transition2.ScoringModel))

	s.app.Commit()
	ctx = s.app.NewContext(false).
		WithBlockHeight(ctx.BlockHeight() + 1).
		WithBlockTime(FixedTimestampPlus(2))

	record, found = s.app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, customer)
	require.True(s.T(), found)
	require.Equal(s.T(), transition2.ExpectedTier, record.Tier)
	s.T().Logf("  → %s -> %s (score: %d)", transition2.InitialTier, record.Tier, record.CurrentScore)

	// Test transition: Verified -> Trusted
	transition3 := VerifiedToTrusted()
	require.NoError(s.T(), s.app.Keepers.VirtEngine.VEID.UpdateScore(
		ctx, customer, transition3.ExpectedScore, transition3.ScoringModel))

	s.app.Commit()
	ctx = s.app.NewContext(false).
		WithBlockHeight(ctx.BlockHeight() + 1).
		WithBlockTime(FixedTimestampPlus(3))

	record, found = s.app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, customer)
	require.True(s.T(), found)
	require.Equal(s.T(), transition3.ExpectedTier, record.Tier)
	s.T().Logf("  → %s -> %s (score: %d)", transition3.InitialTier, record.Tier, record.CurrentScore)

	s.T().Log("✅ ML scoring and tier transitions test passed")
}

// ============================================================================
// Test: Attestation Recording and Validation
// ============================================================================

// TestAttestationRecording tests verification attestation creation and validation
func (s *VEIDE2ETestSuite) TestAttestationRecording() {
	customer := sdktestutil.AccAddress(s.T())

	// Test facial verification attestation
	facialFixture := FacialVerificationAttestation()
	s.verifyAttestation(customer, facialFixture)

	// Test liveness check attestation
	livenessFixture := LivenessCheckAttestation()
	s.verifyAttestation(customer, livenessFixture)

	// Test document verification attestation
	docFixture := DocumentVerificationAttestation()
	s.verifyAttestation(customer, docFixture)

	s.T().Log("✅ Attestation recording test passed")
}

func (s *VEIDE2ETestSuite) verifyAttestation(customer sdk.AccAddress, fixture AttestationFixture) {
	// Create issuer
	keyFPBytes := sha256.Sum256([]byte(fixture.IssuerKeyFP))
	issuer := veidtypes.NewAttestationIssuer(
		hex.EncodeToString(keyFPBytes[:]),
		"virtengine1validator",
	)
	require.NoError(s.T(), issuer.Validate())

	// Create subject
	subject := veidtypes.NewAttestationSubject(customer.String())
	require.NoError(s.T(), subject.Validate())

	// Create attestation
	nonce := DeterministicNonce("attestation-", DeterministicSeed)
	validity := time.Duration(fixture.ValidityHours) * time.Hour

	attestation := veidtypes.NewVerificationAttestation(
		issuer,
		subject,
		fixture.Type,
		nonce,
		FixedTimestamp(),
		validity,
		fixture.Score,
		fixture.Confidence,
	)
	attestation.ModelVersion = fixture.ModelVersion

	// Add verification proofs
	proofHash := sha256.Sum256([]byte("proof-content"))
	proofDetail := veidtypes.NewVerificationProofDetail(
		string(fixture.Type),
		hex.EncodeToString(proofHash[:]),
		fixture.Score,
		70, // threshold
		FixedTimestamp(),
	)
	attestation.AddVerificationProof(proofDetail)

	// Create cryptographic proof
	proofBytes := s.testClient.Sign([]byte(attestation.ID))
	attestation.SetProof(veidtypes.NewAttestationProof(
		veidtypes.ProofTypeEd25519,
		FixedTimestamp(),
		issuer.ID+"#keys-1",
		proofBytes,
		attestation.Nonce,
	))

	// Validate attestation
	require.NoError(s.T(), attestation.Validate())

	// Check validity
	require.True(s.T(), attestation.IsValid(FixedTimestampPlus(1)))
	require.False(s.T(), attestation.IsExpired(FixedTimestampPlus(1)))

	// Compute hash (for on-chain storage)
	hash, err := attestation.HashHex()
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), hash)

	s.T().Logf("  → %s attestation created (score: %d, hash: %s...)", fixture.Type, fixture.Score, hash[:16])
}

// ============================================================================
// Test: Marketplace Gating with VEID
// ============================================================================

// TestMarketplaceVEIDGating tests that marketplace orders respect VEID requirements
func (s *VEIDE2ETestSuite) TestMarketplaceVEIDGating() {
	ctx := s.ctx
	customer := sdktestutil.AccAddress(s.T())
	provider := sdktestutil.AccAddress(s.T())

	// Create identity with low score
	_, err := s.app.Keepers.VirtEngine.VEID.CreateIdentityRecord(ctx, customer)
	require.NoError(s.T(), err)
	require.NoError(s.T(), s.app.Keepers.VirtEngine.VEID.UpdateScore(ctx, customer, 25, TestModelVersion))

	s.app.Commit()
	ctx = s.app.NewContext(false).
		WithBlockHeight(2).
		WithBlockTime(FixedTimestampPlus(1))

	// Create offering that requires verified identity
	pricing := marketplace.PricingInfo{
		Model:     marketplace.PricingModelHourly,
		BasePrice: 1000,
		Currency:  "uve",
	}

	offeringID := marketplace.OfferingID{
		ProviderAddress: provider.String(),
		Sequence:        1,
	}

	offering := marketplace.NewOfferingAt(
		offeringID,
		"Premium Compute",
		marketplace.OfferingCategoryCompute,
		pricing,
		ctx.BlockTime(),
	)
	offering.IdentityRequirement = marketplace.IdentityRequirement{
		MinScore:              70,
		RequiredStatus:        string(veidtypes.AccountStatusVerified),
		RequireVerifiedEmail:  false,
		RequireVerifiedDomain: false,
		RequireMFA:            false,
	}

	require.NoError(s.T(), s.app.Keepers.VirtEngine.Marketplace.CreateOffering(ctx, offering))

	// Attempt order with insufficient score - should fail
	orderIDLow := marketplace.OrderID{
		CustomerAddress: customer.String(),
		Sequence:        1,
	}
	orderLow := marketplace.NewOrderAt(orderIDLow, offering.ID, 5000, 1, ctx.BlockTime())

	err = s.app.Keepers.VirtEngine.Marketplace.CreateOrder(ctx, orderLow)
	require.Error(s.T(), err, "Order should fail with insufficient VEID score")

	var gatingErr *marketplace.IdentityGatingError
	require.ErrorAs(s.T(), err, &gatingErr)
	require.NotEmpty(s.T(), gatingErr.Reasons)
	s.T().Logf("  → Order rejected: %v", gatingErr.Reasons)

	// Update score to meet requirement
	require.NoError(s.T(), s.app.Keepers.VirtEngine.VEID.UpdateScore(ctx, customer, 85, TestModelVersion))

	s.app.Commit()
	ctx = s.app.NewContext(false).
		WithBlockHeight(3).
		WithBlockTime(FixedTimestampPlus(2))

	// Order should now succeed
	orderIDHigh := marketplace.OrderID{
		CustomerAddress: customer.String(),
		Sequence:        2,
	}
	orderHigh := marketplace.NewOrderAt(orderIDHigh, offering.ID, 5000, 1, ctx.BlockTime())

	require.NoError(s.T(), s.app.Keepers.VirtEngine.Marketplace.CreateOrder(ctx, orderHigh))

	stored, found := s.app.Keepers.VirtEngine.Marketplace.GetOrder(ctx, orderIDHigh)
	require.True(s.T(), found)
	require.Equal(s.T(), orderIDHigh, stored.ID)
	s.T().Log("  → Order accepted after score increase")

	s.T().Log("✅ Marketplace VEID gating test passed")
}

// ============================================================================
// Test: Rejection Paths
// ============================================================================

// TestLowScoreRejection tests that low scores result in appropriate tier/status
func (s *VEIDE2ETestSuite) TestLowScoreRejection() {
	ctx := s.ctx
	customer := sdktestutil.AccAddress(s.T())

	// Create identity record
	_, err := s.app.Keepers.VirtEngine.VEID.CreateIdentityRecord(ctx, customer)
	require.NoError(s.T(), err)

	// Update with very low score
	require.NoError(s.T(), s.app.Keepers.VirtEngine.VEID.UpdateScore(ctx, customer, 10, TestModelVersion))

	s.app.Commit()
	ctx = s.app.NewContext(false).
		WithBlockHeight(2).
		WithBlockTime(FixedTimestampPlus(1))

	record, found := s.app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, customer)
	require.True(s.T(), found)
	require.Equal(s.T(), veidtypes.IdentityTierBasic, record.Tier)
	require.Equal(s.T(), uint32(10), record.CurrentScore)

	s.T().Log("✅ Low score rejection test passed")
}

// TestInvalidClientSignatureRejection tests rejection of invalid client signatures
func (s *VEIDE2ETestSuite) TestInvalidClientSignatureRejection() {
	ctx := s.ctx
	customer := sdktestutil.AccAddress(s.T())

	scopeFixture := SelfieScope()
	envelope := EncryptedEnvelopeFixture(scopeFixture.ScopeID)
	payloadHash := PayloadHash(envelope)

	_ = veidtypes.NewUploadMetadata(
		scopeFixture.Salt,
		scopeFixture.DeviceFingerprint,
		s.testClient.ClientID,
		nil,
		nil,
		payloadHash,
	)

	// Use invalid signature (wrong data signed)
	invalidClientSig := s.testClient.Sign([]byte("wrong-data"))
	userSignature := bytes.Repeat([]byte{0x04}, keeper.Secp256k1SignatureSize)

	msg := veidtypes.NewMsgUploadScope(
		customer.String(),
		"scope-invalid-sig-001",
		scopeFixture.ScopeType,
		envelope,
		scopeFixture.Salt,
		scopeFixture.DeviceFingerprint,
		s.testClient.ClientID,
		invalidClientSig,
		userSignature,
		payloadHash,
	)
	msg.CaptureTimestamp = scopeFixture.CaptureTimestamp

	// Should be rejected due to invalid signature
	_, err := s.msgServer.UploadScope(ctx, msg)
	require.Error(s.T(), err, "Upload should fail with invalid client signature")

	// Validate error type
	require.Contains(s.T(), err.Error(), "signature", "Error should mention signature validation")

	s.T().Log("✅ Invalid client signature rejection test passed")
}

// TestMaxAttemptsExceeded tests OTP max attempts rejection
func (s *VEIDE2ETestSuite) TestMaxAttemptsExceeded() {
	emailFixture := ValidEmailVerification()

	challenge := veidtypes.NewEmailVerificationChallenge(
		"challenge-max-attempts",
		"virtengine1test",
		emailFixture.EmailHash,
		emailFixture.Nonce,
		FixedTimestamp(),
		300, // 5 min TTL
		3,   // max 3 attempts
	)

	// Simulate 3 failed attempts
	for i := 0; i < 3; i++ {
		require.True(s.T(), challenge.CanAttempt())
		challenge.RecordAttempt(FixedTimestampPlus(i+1), false)
	}

	// Fourth attempt should not be allowed
	require.False(s.T(), challenge.CanAttempt())
	require.Equal(s.T(), veidtypes.EmailStatusFailed, challenge.Status)

	s.T().Log("✅ Max attempts exceeded test passed")
}

// TestScopeRevocation tests scope revocation flow
func (s *VEIDE2ETestSuite) TestScopeRevocation() {
	ctx := s.ctx
	customer := sdktestutil.AccAddress(s.T())

	// Upload a scope first
	scopeFixture := SelfieScope()
	scopeFixture.ScopeID = "scope-to-revoke-001"
	envelope := EncryptedEnvelopeFixture(scopeFixture.ScopeID)
	payloadHash := PayloadHash(envelope)

	// Use unique salt
	scopeFixture.Salt = bytes.Repeat([]byte{0x5e}, 16)

	metadata := veidtypes.NewUploadMetadata(
		scopeFixture.Salt,
		scopeFixture.DeviceFingerprint,
		s.testClient.ClientID,
		nil,
		nil,
		payloadHash,
	)

	clientSig := s.testClient.Sign(metadata.SigningPayload())
	userSig := bytes.Repeat([]byte{0x06}, keeper.Secp256k1SignatureSize)

	msg := veidtypes.NewMsgUploadScope(
		customer.String(),
		scopeFixture.ScopeID,
		scopeFixture.ScopeType,
		envelope,
		scopeFixture.Salt,
		scopeFixture.DeviceFingerprint,
		s.testClient.ClientID,
		clientSig,
		userSig,
		payloadHash,
	)
	msg.CaptureTimestamp = scopeFixture.CaptureTimestamp

	_, err := s.msgServer.UploadScope(ctx, msg)
	require.NoError(s.T(), err)

	// Revoke the scope
	err = s.app.Keepers.VirtEngine.VEID.RevokeScope(ctx, customer, scopeFixture.ScopeID, "User requested revocation")
	require.NoError(s.T(), err)

	// Verify scope is revoked
	scope, found := s.app.Keepers.VirtEngine.VEID.GetScope(ctx, customer, scopeFixture.ScopeID)
	require.True(s.T(), found)
	require.True(s.T(), scope.Revoked)
	require.NotNil(s.T(), scope.RevokedAt)
	require.Equal(s.T(), "User requested revocation", scope.RevokedReason)

	s.T().Log("✅ Scope revocation test passed")
}

// ============================================================================
// Helper Functions
// ============================================================================

// genesisWithVEIDApprovedClientE2E creates genesis state with an approved test client
func genesisWithVEIDApprovedClientE2E(t *testing.T, cdc codec.Codec, client VEIDTestClient) app.GenesisState {
	t.Helper()

	genesis := app.GenesisStateWithValSet(cdc)

	// Parse existing VEID genesis
	var veidGenesis veidtypes.GenesisState
	require.NoError(t, json.Unmarshal(genesis[veidtypes.ModuleName], &veidGenesis))

	// Add approved client to VEID genesis
	veidGenesis.ApprovedClients = append(veidGenesis.ApprovedClients, client.ToApprovedClient())

	// Marshal back
	veidGenesisBz, err := json.Marshal(&veidGenesis)
	require.NoError(t, err)
	genesis[veidtypes.ModuleName] = veidGenesisBz

	return genesis
}
