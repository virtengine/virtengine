//go:build e2e.integration

// Package integration contains integration tests for VirtEngine.
// These tests verify end-to-end flows against a running localnet.
package integration

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/app"
	sdktestutil "github.com/virtengine/virtengine/sdk/go/testutil"
	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
	"github.com/virtengine/virtengine/x/veid/keeper"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// VEIDOnboardingIntegrationTestSuite tests the VEID onboarding flow with marketplace gating.
//
// Flow:
//  1. Create account + upload scope
//  2. Update validator score (simulated) to low tier
//  3. Verify tier change
//  4. Attempt order placement (should fail VEID gating)
//  5. Update validator score to higher tier
//  6. Place order successfully
type VEIDOnboardingIntegrationTestSuite struct {
	suite.Suite

	app       *app.VirtEngineApp
	ctx       sdk.Context
	client    veidTestClient
	msgServer veidtypes.MsgServer
}

// TestVEIDOnboardingIntegration runs the VEID onboarding integration test suite.
func TestVEIDOnboardingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(VEIDOnboardingIntegrationTestSuite))
}

// SetupSuite runs once before all tests in the suite.
func (s *VEIDOnboardingIntegrationTestSuite) SetupSuite() {
	s.client = newVEIDTestClient()

	s.app = app.Setup(
		app.WithChainID("virtengine-integration-1"),
		app.WithGenesis(func(cdc codec.Codec) app.GenesisState {
			return genesisWithVEIDApprovedClient(s.T(), cdc, s.client)
		}),
	)

	s.ctx = s.app.NewContext(false).
		WithBlockHeight(1).
		WithBlockTime(time.Unix(1_700_000_000, 0).UTC())

	s.msgServer = keeper.NewMsgServerImpl(s.app.Keepers.VirtEngine.VEID)
}

func (s *VEIDOnboardingIntegrationTestSuite) TestVEIDOnboardingFlow() {
	ctx := s.ctx

	customer := sdktestutil.AccAddress(s.T())
	provider := sdktestutil.AccAddress(s.T())

	// Step 1: Upload scope to create identity record
	scopeID := "scope-selfie-001"
	deviceFingerprint := "device-fingerprint-test"
	salt := bytes.Repeat([]byte{0x1a}, 16)

	envelope := encryptiontypes.NewEncryptedPayloadEnvelope()
	envelope.RecipientKeyIDs = []string{"validator-recipient"}
	envelope.Nonce = bytes.Repeat([]byte{0x02}, encryptiontypes.XSalsa20NonceSize)
	envelope.Ciphertext = []byte("encrypted-identity-payload")
	envelope.SenderPubKey = bytes.Repeat([]byte{0x03}, encryptiontypes.X25519PublicKeySize)

	payloadHash := sha256.Sum256(envelope.Ciphertext)

	metadata := veidtypes.NewUploadMetadata(
		salt,
		deviceFingerprint,
		s.client.ClientID,
		nil,
		nil,
		payloadHash[:],
	)

	clientSignature := ed25519.Sign(s.client.PrivateKey, metadata.SigningPayload())
	userSignature := bytes.Repeat([]byte{0x04}, keeper.Secp256k1SignatureSize)

	msg := veidtypes.NewMsgUploadScope(
		customer.String(),
		scopeID,
		veidtypes.ScopeTypeSelfie,
		*envelope,
		salt,
		deviceFingerprint,
		s.client.ClientID,
		clientSignature,
		userSignature,
		payloadHash[:],
	)
	msg.CaptureTimestamp = ctx.BlockTime().Unix()

	resp, err := s.msgServer.UploadScope(ctx, msg)
	require.NoError(s.T(), err)
	require.Equal(s.T(), scopeID, resp.ScopeID)

	s.app.Commit()
	ctx = s.app.NewContext(false).
		WithBlockHeight(2).
		WithBlockTime(ctx.BlockTime().Add(time.Minute))

	record, found := s.app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, customer)
	require.True(s.T(), found)
	require.Equal(s.T(), veidtypes.IdentityTierUnverified, record.Tier)

	// Step 2: Simulate validator score update to a low tier
	require.NoError(s.T(), s.app.Keepers.VirtEngine.VEID.UpdateScore(ctx, customer, 25, "score-model-v1"))

	s.app.Commit()
	ctx = s.app.NewContext(false).
		WithBlockHeight(3).
		WithBlockTime(ctx.BlockTime().Add(time.Minute))

	record, found = s.app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, customer)
	require.True(s.T(), found)
	require.Equal(s.T(), veidtypes.IdentityTierBasic, record.Tier)

	// Step 3: Create offering that requires verified identity + minimum score
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
		"Standard Compute",
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
	offering.RequireMFAForOrders = false

	require.NoError(s.T(), s.app.Keepers.VirtEngine.Marketplace.CreateOffering(ctx, offering))

	// Step 4: Order placement should fail due to insufficient score
	orderIDLow := marketplace.OrderID{
		CustomerAddress: customer.String(),
		Sequence:        1,
	}
	orderLow := marketplace.NewOrderAt(orderIDLow, offering.ID, 5000, 1, ctx.BlockTime())

	err = s.app.Keepers.VirtEngine.Marketplace.CreateOrder(ctx, orderLow)
	require.Error(s.T(), err)

	var gatingErr *marketplace.IdentityGatingError
	require.ErrorAs(s.T(), err, &gatingErr)
	require.NotEmpty(s.T(), gatingErr.Reasons)

	// Step 5: Simulate validator score update to higher tier
	require.NoError(s.T(), s.app.Keepers.VirtEngine.VEID.UpdateScore(ctx, customer, 82, "score-model-v1"))

	s.app.Commit()
	ctx = s.app.NewContext(false).
		WithBlockHeight(4).
		WithBlockTime(ctx.BlockTime().Add(time.Minute))

	record, found = s.app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, customer)
	require.True(s.T(), found)
	require.Equal(s.T(), veidtypes.IdentityTierVerified, record.Tier)

	// Step 6: Order placement succeeds after tier change
	orderIDHigh := marketplace.OrderID{
		CustomerAddress: customer.String(),
		Sequence:        2,
	}
	orderHigh := marketplace.NewOrderAt(orderIDHigh, offering.ID, 5000, 1, ctx.BlockTime())

	require.NoError(s.T(), s.app.Keepers.VirtEngine.Marketplace.CreateOrder(ctx, orderHigh))

	stored, found := s.app.Keepers.VirtEngine.Marketplace.GetOrder(ctx, orderIDHigh)
	require.True(s.T(), found)
	require.Equal(s.T(), orderIDHigh, stored.ID)
	require.Equal(s.T(), marketplace.OrderStatePendingPayment, stored.State)
}
