//go:build e2e.integration

// Package helpers provides shared test helper functions for E2E tests.
//
// This file implements VEID-specific helpers for:
// - Account creation with identity records
// - Scope upload with encryption
// - Score updates and tier transitions
// - Market order creation with VEID gating
//
// Task Reference: VE-15B - E2E VEID onboarding flow (account → order)
package helpers

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/json"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/app"
	sdktestutil "github.com/virtengine/virtengine/sdk/go/testutil"
	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
	"github.com/virtengine/virtengine/x/veid/keeper"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Constants
// ============================================================================

const (
	// TestChainID is the chain ID for E2E tests
	TestChainID = "virtengine-e2e-onboarding"

	// TestModelVersion is the ML model version for scoring tests
	TestModelVersion = "veid-score-v1.0.0-e2e-onboarding"

	// TestClientID is the approved client ID for test uploads
	TestClientID = "ve-e2e-onboarding-app"

	// TestDeviceFingerprint is a deterministic device fingerprint
	TestDeviceFingerprint = "e2e-device-onboarding-001"

	// TestBlockTimeUnix is a fixed block time for deterministic tests
	TestBlockTimeUnix = 1700000000

	// DeterministicSeed for reproducible key generation
	DeterministicSeed = 42

	// Secp256k1SignatureSize is the expected size of a secp256k1 signature
	Secp256k1SignatureSize = 65
)

// ============================================================================
// Test Client (Approved Capture App)
// ============================================================================

// OnboardingTestClient represents an approved capture client for onboarding tests
type OnboardingTestClient struct {
	ClientID   string
	Name       string
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
}

// NewOnboardingTestClient creates a deterministic test client using fixed seed
func NewOnboardingTestClient() OnboardingTestClient {
	// Use deterministic seed for reproducible key generation
	seed := bytes.Repeat([]byte{byte(DeterministicSeed + 1)}, ed25519.SeedSize)
	privateKey := ed25519.NewKeyFromSeed(seed)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	return OnboardingTestClient{
		ClientID:   TestClientID,
		Name:       "E2E Onboarding Test Capture App",
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}
}

// Sign signs data with the test client's private key
func (c OnboardingTestClient) Sign(data []byte) []byte {
	return ed25519.Sign(c.PrivateKey, data)
}

// ToApprovedClient converts to an ApprovedClient type for genesis
func (c OnboardingTestClient) ToApprovedClient() veidtypes.ApprovedClient {
	return veidtypes.ApprovedClient{
		ClientID:     c.ClientID,
		Name:         c.Name,
		PublicKey:    c.PublicKey,
		Algorithm:    "Ed25519",
		Active:       true,
		RegisteredAt: TestBlockTimeUnix,
	}
}

// ============================================================================
// App Setup Helpers
// ============================================================================

// SetupOnboardingTestApp creates a test app with VEID approved client in genesis
func SetupOnboardingTestApp(t *testing.T, client OnboardingTestClient) *app.VirtEngineApp {
	t.Helper()

	return app.Setup(
		app.WithChainID(TestChainID),
		app.WithGenesis(func(cdc codec.Codec) app.GenesisState {
			return genesisWithApprovedClient(t, cdc, client)
		}),
	)
}

// genesisWithApprovedClient creates genesis state with an approved test client
func genesisWithApprovedClient(t *testing.T, cdc codec.Codec, client OnboardingTestClient) app.GenesisState {
	t.Helper()

	genesis := app.NewDefaultGenesisState(cdc)

	var veidGenesis veidtypes.GenesisState
	require.NoError(t, json.Unmarshal(genesis[veidtypes.ModuleName], &veidGenesis))

	veidGenesis.ApprovedClients = append(veidGenesis.ApprovedClients, client.ToApprovedClient())

	veidGenesisBz, err := json.Marshal(&veidGenesis)
	require.NoError(t, err)
	genesis[veidtypes.ModuleName] = veidGenesisBz

	return genesis
}

// NewTestContext creates a new test context with the given block height and time
func NewTestContext(a *app.VirtEngineApp, height int64, timestamp time.Time) sdk.Context {
	return a.NewContext(false).
		WithBlockHeight(height).
		WithBlockTime(timestamp)
}

// ============================================================================
// Account Creation Helpers
// ============================================================================

// CreateTestAccount creates a new test account address
func CreateTestAccount(t *testing.T) sdk.AccAddress {
	t.Helper()
	return sdktestutil.AccAddress(t)
}

// CreateIdentityRecordForAccount creates an identity record for a test account
func CreateIdentityRecordForAccount(
	t *testing.T,
	a *app.VirtEngineApp,
	ctx sdk.Context,
	account sdk.AccAddress,
) *veidtypes.IdentityRecord {
	t.Helper()

	record, err := a.Keepers.VirtEngine.VEID.CreateIdentityRecord(ctx, account)
	require.NoError(t, err)
	require.Equal(t, veidtypes.IdentityTierUnverified, record.Tier)
	require.Equal(t, uint32(0), record.CurrentScore)

	return record
}

// ============================================================================
// Scope Upload Helpers
// ============================================================================

// ScopeUploadParams contains parameters for scope upload
type ScopeUploadParams struct {
	ScopeID           string
	ScopeType         veidtypes.ScopeType
	Salt              []byte
	DeviceFingerprint string
	CaptureTimestamp  int64
}

// DefaultSelfieUploadParams returns default parameters for selfie scope upload
func DefaultSelfieUploadParams(scopeID string) ScopeUploadParams {
	return ScopeUploadParams{
		ScopeID:           scopeID,
		ScopeType:         veidtypes.ScopeTypeSelfie,
		Salt:              bytes.Repeat([]byte{0x1a}, 16),
		DeviceFingerprint: TestDeviceFingerprint,
		CaptureTimestamp:  TestBlockTimeUnix,
	}
}

// DefaultIDDocUploadParams returns default parameters for ID document scope upload
func DefaultIDDocUploadParams(scopeID string) ScopeUploadParams {
	return ScopeUploadParams{
		ScopeID:           scopeID,
		ScopeType:         veidtypes.ScopeTypeIDDocument,
		Salt:              bytes.Repeat([]byte{0x2b}, 16),
		DeviceFingerprint: TestDeviceFingerprint,
		CaptureTimestamp:  TestBlockTimeUnix,
	}
}

// DefaultFaceVideoUploadParams returns default parameters for face video scope upload
func DefaultFaceVideoUploadParams(scopeID string) ScopeUploadParams {
	return ScopeUploadParams{
		ScopeID:           scopeID,
		ScopeType:         veidtypes.ScopeTypeFaceVideo,
		Salt:              bytes.Repeat([]byte{0x3c}, 16),
		DeviceFingerprint: TestDeviceFingerprint,
		CaptureTimestamp:  TestBlockTimeUnix,
	}
}

// UploadScope uploads an encrypted scope for a customer
func UploadScope(
	t *testing.T,
	msgServer veidtypes.MsgServer,
	ctx sdk.Context,
	customer sdk.AccAddress,
	client OnboardingTestClient,
	params ScopeUploadParams,
) *veidtypes.MsgUploadScopeResponse {
	t.Helper()

	envelope := CreateEncryptedEnvelope(params.ScopeID)
	payloadHash := ComputePayloadHash(envelope)

	metadata := veidtypes.NewUploadMetadata(
		params.Salt,
		params.DeviceFingerprint,
		client.ClientID,
		nil,
		nil,
		payloadHash,
	)

	clientSignature := client.Sign(metadata.SigningPayload())
	userSignature := bytes.Repeat([]byte{0x04}, Secp256k1SignatureSize)

	msg := veidtypes.NewMsgUploadScope(
		customer.String(),
		params.ScopeID,
		params.ScopeType,
		envelope,
		params.Salt,
		params.DeviceFingerprint,
		client.ClientID,
		clientSignature,
		userSignature,
		payloadHash,
	)
	msg.CaptureTimestamp = params.CaptureTimestamp

	resp, err := msgServer.UploadScope(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, params.ScopeID, resp.ScopeId)

	return resp
}

// CreateEncryptedEnvelope creates a deterministic encrypted envelope for testing
func CreateEncryptedEnvelope(scopeID string) encryptiontypes.EncryptedPayloadEnvelope {
	envelope := encryptiontypes.NewEncryptedPayloadEnvelope()
	envelope.RecipientKeyIDs = []string{"e2e-validator-recipient-onboarding"}
	envelope.Nonce = bytes.Repeat([]byte{0x02}, encryptiontypes.XSalsa20NonceSize)
	envelope.Ciphertext = []byte("e2e-encrypted-identity-payload-" + scopeID)
	envelope.SenderPubKey = bytes.Repeat([]byte{0x03}, encryptiontypes.X25519PublicKeySize)
	return *envelope
}

// ComputePayloadHash computes SHA256 hash of envelope ciphertext
func ComputePayloadHash(envelope encryptiontypes.EncryptedPayloadEnvelope) []byte {
	hash := sha256.Sum256(envelope.Ciphertext)
	return hash[:]
}

// ============================================================================
// Scoring and Tier Helpers
// ============================================================================

// UpdateAccountScore updates the VEID score for an account
func UpdateAccountScore(
	t *testing.T,
	a *app.VirtEngineApp,
	ctx sdk.Context,
	account sdk.AccAddress,
	score uint32,
) {
	t.Helper()

	err := a.Keepers.VirtEngine.VEID.UpdateScore(ctx, account, score, TestModelVersion)
	require.NoError(t, err)
}

// VerifyAccountTier verifies an account has the expected tier
func VerifyAccountTier(
	t *testing.T,
	a *app.VirtEngineApp,
	ctx sdk.Context,
	account sdk.AccAddress,
	expectedTier veidtypes.IdentityTier,
) {
	t.Helper()

	record, found := a.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, account)
	require.True(t, found, "Identity record should exist")
	require.Equal(t, expectedTier, record.Tier, "Tier should match expected")
}

// VerifyAccountScore verifies an account has the expected score
func VerifyAccountScore(
	t *testing.T,
	a *app.VirtEngineApp,
	ctx sdk.Context,
	account sdk.AccAddress,
	expectedScore uint32,
) {
	t.Helper()

	record, found := a.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, account)
	require.True(t, found, "Identity record should exist")
	require.Equal(t, expectedScore, record.CurrentScore, "Score should match expected")
}

// GetAccountRecord retrieves the identity record for an account
func GetAccountRecord(
	t *testing.T,
	a *app.VirtEngineApp,
	ctx sdk.Context,
	account sdk.AccAddress,
) *veidtypes.IdentityRecord {
	t.Helper()

	record, found := a.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, account)
	require.True(t, found, "Identity record should exist")
	return &record
}

// ============================================================================
// Market Order Helpers
// ============================================================================

// CreateOfferingWithVEIDRequirement creates an offering that requires VEID verification
func CreateOfferingWithVEIDRequirement(
	t *testing.T,
	a *app.VirtEngineApp,
	ctx sdk.Context,
	provider sdk.AccAddress,
	minScore uint32,
	requiredStatus string,
) marketplace.Offering {
	t.Helper()

	pricing := marketplace.PricingInfo{
		Model:     marketplace.PricingModelHourly,
		BasePrice: 1000,
		Currency:  "uve",
	}

	offeringID := marketplace.OfferingID{
		ProviderAddress: provider.String(),
		Sequence:        uint64(ctx.BlockHeight()),
	}

	offering := marketplace.NewOfferingAt(
		offeringID,
		"E2E VEID Test Compute",
		marketplace.OfferingCategoryCompute,
		pricing,
		ctx.BlockTime(),
	)
	offering.IdentityRequirement = marketplace.IdentityRequirement{
		MinScore:              minScore,
		RequiredStatus:        requiredStatus,
		RequireVerifiedEmail:  false,
		RequireVerifiedDomain: false,
		RequireMFA:            false,
	}

	err := a.Keepers.VirtEngine.Marketplace.CreateOffering(ctx, offering)
	require.NoError(t, err)

	return *offering
}

// AttemptCreateOrder attempts to create an order (may fail due to gating)
func AttemptCreateOrder(
	t *testing.T,
	a *app.VirtEngineApp,
	ctx sdk.Context,
	customer sdk.AccAddress,
	offering marketplace.Offering,
	expectedError bool,
) *marketplace.Order {
	t.Helper()

	orderID := marketplace.OrderID{
		CustomerAddress: customer.String(),
		Sequence:        uint64(ctx.BlockHeight()),
	}

	order := marketplace.NewOrderAt(orderID, offering.ID, 5000, 1, ctx.BlockTime())

	err := a.Keepers.VirtEngine.Marketplace.CreateOrder(ctx, order)

	if expectedError {
		require.Error(t, err, "Order should fail due to VEID gating")
		var gatingErr *marketplace.IdentityGatingError
		require.ErrorAs(t, err, &gatingErr, "Error should be IdentityGatingError")
		return nil
	}

	require.NoError(t, err, "Order should succeed")

	stored, found := a.Keepers.VirtEngine.Marketplace.GetOrder(ctx, orderID)
	require.True(t, found, "Order should be stored")
	return stored
}

// ============================================================================
// Time Helpers
// ============================================================================

// FixedTimestamp returns a deterministic timestamp for tests
func FixedTimestamp() time.Time {
	return time.Unix(TestBlockTimeUnix, 0).UTC()
}

// FixedTimestampPlus returns a timestamp offset from the fixed timestamp
func FixedTimestampPlus(minutes int) time.Time {
	return FixedTimestamp().Add(time.Duration(minutes) * time.Minute)
}

// CommitAndAdvanceBlock commits the current state and advances to the next block
func CommitAndAdvanceBlock(a *app.VirtEngineApp, ctx sdk.Context) sdk.Context {
	a.Commit()
	return a.NewContext(false).
		WithBlockHeight(ctx.BlockHeight() + 1).
		WithBlockTime(ctx.BlockTime().Add(time.Minute))
}

// ============================================================================
// Msg Server Helpers
// ============================================================================

// GetVEIDMsgServer returns the VEID msg server for a test app
func GetVEIDMsgServer(a *app.VirtEngineApp) veidtypes.MsgServer {
	return keeper.NewMsgServerImpl(a.Keepers.VirtEngine.VEID)
}
