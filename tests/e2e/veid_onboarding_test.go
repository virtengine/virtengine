//go:build e2e.integration

// Package e2e contains end-to-end tests for VirtEngine.
//
// This file implements the comprehensive E2E VEID onboarding flow test:
// 1. Create new account
// 2. Upload identity scope (encrypted)
// 3. Validator receives and processes scope
// 4. ML scoring computes trust score
// 5. Tier transition occurs
// 6. Account attempts marketplace order
// 7. Order succeeds if tier/score sufficient
//
// Task Reference: VE-15B - E2E VEID onboarding flow (account → order)
package e2e

import (
	"bytes"
	"testing"

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
// VEID Onboarding E2E Test Suite
// ============================================================================

// VEIDOnboardingTestSuite tests the complete VEID onboarding flow from
// account creation to successful marketplace order placement.
//
// Test Coverage:
//  1. Account creation with initial identity record
//  2. Encrypted scope upload with signature validation
//  3. ML scoring simulation and tier transitions
//  4. Marketplace order gating by VEID score/tier
//  5. Order success after meeting VEID requirements
//  6. Multiple tier transition scenarios
//  7. Rejection paths (insufficient score)
type VEIDOnboardingTestSuite struct {
	suite.Suite

	app        *app.VirtEngineApp
	ctx        sdk.Context
	testClient VEIDTestClient
	msgServer  veidtypes.MsgServer
}

// TestVEIDOnboarding runs the VEID onboarding E2E test suite.
func TestVEIDOnboarding(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	suite.Run(t, new(VEIDOnboardingTestSuite))
}

// SetupSuite runs once before all tests in the suite.
func (s *VEIDOnboardingTestSuite) SetupSuite() {
	s.testClient = NewVEIDTestClient()

	s.app = app.Setup(
		app.WithChainID("virtengine-onboarding-e2e"),
		app.WithGenesis(func(cdc codec.Codec) app.GenesisState {
			return genesisWithVEIDApprovedClientOnboarding(s.T(), cdc, s.testClient)
		}),
	)

	s.ctx = s.app.NewContext(false).
		WithBlockHeight(1).
		WithBlockTime(FixedTimestamp())

	s.msgServer = keeper.NewMsgServerImpl(s.app.Keepers.VirtEngine.VEID)
}

// ============================================================================
// Test: Full Onboarding Flow (Account → Order)
// ============================================================================

// TestFullOnboardingToMarketplaceOrder tests the complete flow:
// 1. Create account → 2. Upload scope → 3. Score update → 4. Tier transition → 5. Order
func (s *VEIDOnboardingTestSuite) TestFullOnboardingToMarketplaceOrder() {
	ctx := s.ctx
	t := s.T()

	// Step 1: Create new account
	customer := sdktestutil.AccAddress(t)
	provider := sdktestutil.AccAddress(t)
	t.Logf("Step 1: Created customer account: %s", customer)
	t.Logf("        Created provider account: %s", provider)

	// Step 2: Upload selfie scope to create identity record
	selfieScopeFixture := SelfieScope()
	envelope := EncryptedEnvelopeFixture(selfieScopeFixture.ScopeID)
	payloadHash := PayloadHash(envelope)

	metadata := veidtypes.NewUploadMetadata(
		selfieScopeFixture.Salt,
		selfieScopeFixture.DeviceFingerprint,
		s.testClient.ClientID,
		nil,
		nil,
		payloadHash,
	)

	clientSignature := s.testClient.Sign(metadata.SigningPayload())
	userSignature := bytes.Repeat([]byte{0x04}, keeper.Secp256k1SignatureSize)

	msg := veidtypes.NewMsgUploadScope(
		customer.String(),
		selfieScopeFixture.ScopeID,
		selfieScopeFixture.ScopeType,
		envelope,
		selfieScopeFixture.Salt,
		selfieScopeFixture.DeviceFingerprint,
		s.testClient.ClientID,
		clientSignature,
		userSignature,
		payloadHash,
	)
	msg.CaptureTimestamp = selfieScopeFixture.CaptureTimestamp

	resp, err := s.msgServer.UploadScope(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, selfieScopeFixture.ScopeID, resp.ScopeId)
	t.Logf("Step 2: Uploaded encrypted selfie scope: %s", resp.ScopeId)

	// Verify identity record was created
	s.app.Commit()
	ctx = s.app.NewContext(false).
		WithBlockHeight(2).
		WithBlockTime(FixedTimestampPlus(1))

	record, found := s.app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, customer)
	require.True(t, found, "Identity record should be created after scope upload")
	require.Equal(t, veidtypes.IdentityTierUnverified, record.Tier)
	require.Equal(t, 1, len(record.ScopeRefs))
	t.Logf("        Identity record created with tier: %s", record.Tier)

	// Step 3: Simulate validator processing and ML scoring (low initial score)
	initialScore := uint32(25)
	require.NoError(t, s.app.Keepers.VirtEngine.VEID.UpdateScore(ctx, customer, initialScore, TestModelVersion))

	s.app.Commit()
	ctx = s.app.NewContext(false).
		WithBlockHeight(3).
		WithBlockTime(FixedTimestampPlus(2))

	record, found = s.app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, customer)
	require.True(t, found)
	require.Equal(t, veidtypes.IdentityTierBasic, record.Tier)
	require.Equal(t, initialScore, record.CurrentScore)
	t.Logf("Step 3: ML scoring computed trust score: %d (tier: %s)", record.CurrentScore, record.Tier)

	// Step 4: Create marketplace offering with VEID requirements
	minRequiredScore := uint32(70)
	requiredStatus := string(veidtypes.AccountStatusVerified)

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
		"Premium VEID-Gated Compute",
		marketplace.OfferingCategoryCompute,
		pricing,
		ctx.BlockTime(),
	)
	offering.IdentityRequirement = marketplace.IdentityRequirement{
		MinScore:              minRequiredScore,
		RequiredStatus:        requiredStatus,
		RequireVerifiedEmail:  false,
		RequireVerifiedDomain: false,
		RequireMFA:            false,
	}

	require.NoError(t, s.app.Keepers.VirtEngine.Marketplace.CreateOffering(ctx, offering))
	t.Logf("Step 4: Created offering with VEID requirement (min score: %d)", minRequiredScore)

	// Step 5: Attempt order with insufficient score - should fail
	orderIDLow := marketplace.OrderID{
		CustomerAddress: customer.String(),
		Sequence:        1,
	}
	orderLow := marketplace.NewOrderAt(orderIDLow, offering.ID, 5000, 1, ctx.BlockTime())

	err = s.app.Keepers.VirtEngine.Marketplace.CreateOrder(ctx, orderLow)
	require.Error(t, err, "Order should fail with insufficient VEID score")

	var gatingErr *marketplace.IdentityGatingError
	require.ErrorAs(t, err, &gatingErr)
	require.NotEmpty(t, gatingErr.Reasons)
	t.Logf("Step 5: Order rejected due to insufficient score: %v", gatingErr.Reasons[0].Message)

	// Step 6: Upload additional ID document scope for higher score
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
	require.NoError(t, err)
	require.Equal(t, idDocFixture.ScopeID, idDocResp.ScopeId)
	t.Logf("Step 6: Uploaded ID document scope: %s", idDocResp.ScopeId)

	// Step 7: Update score to meet marketplace requirements
	verifiedScore := uint32(82)
	require.NoError(t, s.app.Keepers.VirtEngine.VEID.UpdateScore(ctx, customer, verifiedScore, TestModelVersion))

	s.app.Commit()
	ctx = s.app.NewContext(false).
		WithBlockHeight(4).
		WithBlockTime(FixedTimestampPlus(3))

	record, found = s.app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, customer)
	require.True(t, found)
	require.Equal(t, veidtypes.IdentityTierVerified, record.Tier)
	require.Equal(t, verifiedScore, record.CurrentScore)
	require.Equal(t, 2, len(record.ScopeRefs))
	t.Logf("Step 7: Tier transition complete - score: %d, tier: %s, scopes: %d",
		record.CurrentScore, record.Tier, len(record.ScopeRefs))

	// Step 8: Create order - should now succeed
	orderIDHigh := marketplace.OrderID{
		CustomerAddress: customer.String(),
		Sequence:        2,
	}
	orderHigh := marketplace.NewOrderAt(orderIDHigh, offering.ID, 5000, 1, ctx.BlockTime())

	require.NoError(t, s.app.Keepers.VirtEngine.Marketplace.CreateOrder(ctx, orderHigh))

	stored, found := s.app.Keepers.VirtEngine.Marketplace.GetOrder(ctx, orderIDHigh)
	require.True(t, found)
	require.Equal(t, orderIDHigh, stored.ID)
	t.Logf("Step 8: Order successfully created after meeting VEID requirements!")

	t.Log("✅ Full onboarding flow test passed (account → scope → score → tier → order)")
}

// ============================================================================
// Test: Encrypted Scope Handling
// ============================================================================

// TestEncryptedScopeHandling verifies that encrypted scopes are properly handled
func (s *VEIDOnboardingTestSuite) TestEncryptedScopeHandling() {
	ctx := s.ctx
	t := s.T()
	customer := sdktestutil.AccAddress(t)

	// Create scope with encryption envelope
	scopeFixture := SelfieScope()
	scopeFixture.ScopeID = "scope-encrypted-handling-001"
	envelope := EncryptedEnvelopeFixture(scopeFixture.ScopeID)
	payloadHash := PayloadHash(envelope)

	// Verify envelope structure
	require.NotEmpty(t, envelope.RecipientKeyIDs, "Envelope should have recipient key IDs")
	require.NotEmpty(t, envelope.Nonce, "Envelope should have nonce")
	require.NotEmpty(t, envelope.Ciphertext, "Envelope should have ciphertext")
	require.NotEmpty(t, envelope.SenderPubKey, "Envelope should have sender public key")

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
	require.NoError(t, err)
	require.Equal(t, scopeFixture.ScopeID, resp.ScopeId)

	// Verify scope was stored with encryption metadata
	s.app.Commit()
	ctx = s.app.NewContext(false).
		WithBlockHeight(2).
		WithBlockTime(FixedTimestampPlus(1))

	scope, found := s.app.Keepers.VirtEngine.VEID.GetScope(ctx, customer, scopeFixture.ScopeID)
	require.True(t, found, "Scope should be stored")
	require.NotEmpty(t, scope.UploadMetadata.PayloadHash, "Scope should have payload hash stored")

	t.Log("✅ Encrypted scope handling test passed")
}

// ============================================================================
// Test: ML Scoring Integration
// ============================================================================

// TestMLScoringIntegration verifies ML scoring updates identity records correctly
func (s *VEIDOnboardingTestSuite) TestMLScoringIntegration() {
	ctx := s.ctx
	t := s.T()
	customer := sdktestutil.AccAddress(t)

	// Create identity record
	record, err := s.app.Keepers.VirtEngine.VEID.CreateIdentityRecord(ctx, customer)
	require.NoError(t, err)
	require.Equal(t, uint32(0), record.CurrentScore)
	require.Equal(t, veidtypes.IdentityTierUnverified, record.Tier)

	// Test incremental score updates
	testScores := []struct {
		score        uint32
		expectedTier veidtypes.IdentityTier
	}{
		{10, veidtypes.IdentityTierBasic},
		{35, veidtypes.IdentityTierStandard},
		{65, veidtypes.IdentityTierVerified},
		{90, veidtypes.IdentityTierTrusted},
	}

	for i, tc := range testScores {
		require.NoError(t, s.app.Keepers.VirtEngine.VEID.UpdateScore(ctx, customer, tc.score, TestModelVersion))

		s.app.Commit()
		ctx = s.app.NewContext(false).
			WithBlockHeight(int64(i + 2)).
			WithBlockTime(FixedTimestampPlus(i + 1))

		record, found := s.app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, customer)
		require.True(t, found)
		require.Equal(t, tc.score, record.CurrentScore, "Score should match")
		require.Equal(t, tc.expectedTier, record.Tier, "Tier should match for score %d", tc.score)

		t.Logf("  Score %d → Tier %s ✓", tc.score, record.Tier)
	}

	t.Log("✅ ML scoring integration test passed")
}

// ============================================================================
// Test: Tier Transitions
// ============================================================================

// TestTierTransitions verifies all tier transitions work correctly
func (s *VEIDOnboardingTestSuite) TestTierTransitions() {
	ctx := s.ctx
	t := s.T()
	customer := sdktestutil.AccAddress(t)

	// Create identity record
	_, err := s.app.Keepers.VirtEngine.VEID.CreateIdentityRecord(ctx, customer)
	require.NoError(t, err)

	transitions := []ScoreTransitionFixture{
		UnverifiedToBasic(),
		BasicToVerified(),
		VerifiedToTrusted(),
	}

	for i, transition := range transitions {
		require.NoError(t, s.app.Keepers.VirtEngine.VEID.UpdateScore(
			ctx, customer, transition.ExpectedScore, transition.ScoringModel))

		s.app.Commit()
		ctx = s.app.NewContext(false).
			WithBlockHeight(int64(i + 2)).
			WithBlockTime(FixedTimestampPlus(i + 1))

		record, found := s.app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, customer)
		require.True(t, found)
		require.Equal(t, transition.ExpectedTier, record.Tier,
			"Transition %s → %s failed", transition.InitialTier, transition.ExpectedTier)

		t.Logf("  %s → %s (score: %d) ✓", transition.InitialTier, record.Tier, record.CurrentScore)
	}

	t.Log("✅ Tier transitions test passed")
}

// ============================================================================
// Test: Order Gating by Tier
// ============================================================================

// TestOrderGatingByTier verifies marketplace orders are gated by VEID tier
func (s *VEIDOnboardingTestSuite) TestOrderGatingByTier() {
	ctx := s.ctx
	t := s.T()
	customer := sdktestutil.AccAddress(t)
	provider := sdktestutil.AccAddress(t)

	// Create identity with basic tier
	_, err := s.app.Keepers.VirtEngine.VEID.CreateIdentityRecord(ctx, customer)
	require.NoError(t, err)
	require.NoError(t, s.app.Keepers.VirtEngine.VEID.UpdateScore(ctx, customer, 30, TestModelVersion))

	s.app.Commit()
	ctx = s.app.NewContext(false).
		WithBlockHeight(2).
		WithBlockTime(FixedTimestampPlus(1))

	// Create offering requiring high score
	pricing := marketplace.PricingInfo{
		Model:     marketplace.PricingModelHourly,
		BasePrice: 2000,
		Currency:  "uve",
	}

	offeringID := marketplace.OfferingID{
		ProviderAddress: provider.String(),
		Sequence:        100,
	}

	offering := marketplace.NewOfferingAt(
		offeringID,
		"Premium Tier-Gated Service",
		marketplace.OfferingCategoryCompute,
		pricing,
		ctx.BlockTime(),
	)
	offering.IdentityRequirement = marketplace.IdentityRequirement{
		MinScore:       80,
		RequiredStatus: string(veidtypes.AccountStatusVerified),
	}

	require.NoError(t, s.app.Keepers.VirtEngine.Marketplace.CreateOffering(ctx, offering))

	// Attempt order with basic tier - should fail
	orderID := marketplace.OrderID{
		CustomerAddress: customer.String(),
		Sequence:        100,
	}
	order := marketplace.NewOrderAt(orderID, offering.ID, 10000, 1, ctx.BlockTime())

	err = s.app.Keepers.VirtEngine.Marketplace.CreateOrder(ctx, order)
	require.Error(t, err, "Order should fail for basic tier")

	var gatingErr *marketplace.IdentityGatingError
	require.ErrorAs(t, err, &gatingErr)
	t.Logf("  Basic tier order rejected: %s", gatingErr.Reasons[0].Message)

	// Upgrade to verified tier
	require.NoError(t, s.app.Keepers.VirtEngine.VEID.UpdateScore(ctx, customer, 85, TestModelVersion))

	s.app.Commit()
	ctx = s.app.NewContext(false).
		WithBlockHeight(3).
		WithBlockTime(FixedTimestampPlus(2))

	// Retry order - should succeed
	orderID2 := marketplace.OrderID{
		CustomerAddress: customer.String(),
		Sequence:        101,
	}
	order2 := marketplace.NewOrderAt(orderID2, offering.ID, 10000, 1, ctx.BlockTime())

	require.NoError(t, s.app.Keepers.VirtEngine.Marketplace.CreateOrder(ctx, order2))
	t.Log("  Verified tier order accepted ✓")

	t.Log("✅ Order gating by tier test passed")
}

// ============================================================================
// Test: Multiple Scope Types
// ============================================================================

// TestMultipleScopeTypes verifies handling of different scope types in onboarding
func (s *VEIDOnboardingTestSuite) TestMultipleScopeTypes() {
	ctx := s.ctx
	t := s.T()
	customer := sdktestutil.AccAddress(t)

	scopeTypes := []struct {
		fixture   ScopeFixture
		scopeType veidtypes.ScopeType
	}{
		{SelfieScope(), veidtypes.ScopeTypeSelfie},
		{IDDocumentScope(), veidtypes.ScopeTypeIDDocument},
		{FaceVideoScope(), veidtypes.ScopeTypeFaceVideo},
	}

	for i, tc := range scopeTypes {
		// Create unique scope ID for each
		tc.fixture.ScopeID = tc.fixture.ScopeID + "-multi-" + string(rune('A'+i))
		tc.fixture.Salt = bytes.Repeat([]byte{byte(0x10 + i)}, 16)

		envelope := EncryptedEnvelopeFixture(tc.fixture.ScopeID)
		payloadHash := PayloadHash(envelope)

		metadata := veidtypes.NewUploadMetadata(
			tc.fixture.Salt,
			tc.fixture.DeviceFingerprint,
			s.testClient.ClientID,
			nil,
			nil,
			payloadHash,
		)

		clientSig := s.testClient.Sign(metadata.SigningPayload())
		userSig := bytes.Repeat([]byte{byte(0x04 + i)}, keeper.Secp256k1SignatureSize)

		msg := veidtypes.NewMsgUploadScope(
			customer.String(),
			tc.fixture.ScopeID,
			tc.fixture.ScopeType,
			envelope,
			tc.fixture.Salt,
			tc.fixture.DeviceFingerprint,
			s.testClient.ClientID,
			clientSig,
			userSig,
			payloadHash,
		)
		msg.CaptureTimestamp = tc.fixture.CaptureTimestamp

		resp, err := s.msgServer.UploadScope(ctx, msg)
		require.NoError(t, err)
		require.Equal(t, tc.fixture.ScopeID, resp.ScopeId)

		t.Logf("  Uploaded %s scope: %s ✓", tc.scopeType, resp.ScopeId)
	}

	// Verify all scopes are recorded
	s.app.Commit()
	ctx = s.app.NewContext(false).
		WithBlockHeight(2).
		WithBlockTime(FixedTimestampPlus(1))

	record, found := s.app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, customer)
	require.True(t, found)
	require.Equal(t, len(scopeTypes), len(record.ScopeRefs),
		"All scope types should be recorded")

	t.Log("✅ Multiple scope types test passed")
}

// ============================================================================
// Test: CI Integration (Localnet Compatible)
// ============================================================================

// TestOnboardingFlowCICompatible tests the flow in a CI-compatible manner
func (s *VEIDOnboardingTestSuite) TestOnboardingFlowCICompatible() {
	ctx := s.ctx
	t := s.T()

	// This test is designed to run in CI with localnet
	// It uses deterministic values for reproducibility

	customer := sdktestutil.AccAddress(t)

	// Step 1: Create identity
	_, err := s.app.Keepers.VirtEngine.VEID.CreateIdentityRecord(ctx, customer)
	require.NoError(t, err)

	// Step 2: Simulate onboarding progression
	stages := []struct {
		score       uint32
		tier        veidtypes.IdentityTier
		description string
	}{
		{0, veidtypes.IdentityTierUnverified, "Initial state"},
		{25, veidtypes.IdentityTierBasic, "After selfie upload"},
		{60, veidtypes.IdentityTierVerified, "After ID document"},
		{85, veidtypes.IdentityTierTrusted, "After liveness check"},
	}

	for i, stage := range stages {
		if stage.score > 0 {
			require.NoError(t, s.app.Keepers.VirtEngine.VEID.UpdateScore(
				ctx, customer, stage.score, TestModelVersion))

			s.app.Commit()
			ctx = s.app.NewContext(false).
				WithBlockHeight(int64(i + 2)).
				WithBlockTime(FixedTimestampPlus(i + 1))
		}

		record, found := s.app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, customer)
		require.True(t, found)
		require.Equal(t, stage.tier, record.Tier, "Stage: %s", stage.description)

		t.Logf("  %s: score=%d, tier=%s ✓", stage.description, stage.score, record.Tier)
	}

	t.Log("✅ CI-compatible onboarding flow test passed")
}

// ============================================================================
// Helper Functions
// ============================================================================

// genesisWithVEIDApprovedClientOnboarding creates genesis state with an approved test client
func genesisWithVEIDApprovedClientOnboarding(t *testing.T, cdc codec.Codec, client VEIDTestClient) app.GenesisState {
	t.Helper()

	genesis := app.NewDefaultGenesisState(cdc)

	var veidGenesis veidtypes.GenesisState
	require.NoError(t, cdc.UnmarshalJSON(genesis[veidtypes.ModuleName], &veidGenesis))

	veidGenesis.ApprovedClients = append(veidGenesis.ApprovedClients, client.ToApprovedClient())

	veidGenesisBz, err := cdc.MarshalJSON(&veidGenesis)
	require.NoError(t, err)
	genesis[veidtypes.ModuleName] = veidGenesisBz

	return genesis
}
