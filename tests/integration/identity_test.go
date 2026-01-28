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
	"github.com/virtengine/virtengine/x/veid/keeper"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// IdentityIntegrationTestSuite tests identity-related flows.
// This suite verifies:
//   - Identity scope upload transactions
//   - Identity score computation and commitment
//   - Identity verification pipeline
//
// Acceptance Criteria (VE-002):
//   - Integration test suite can submit identity scope upload
//   - Observe identity score committed to chain state
type IdentityIntegrationTestSuite struct {
	suite.Suite

	app        *app.VirtEngineApp
	ctx        sdk.Context
	client     veidTestClient
	msgServer  veidtypes.MsgServer
	validator  sdk.AccAddress
}

// TestIdentityIntegration runs the identity integration test suite.
func TestIdentityIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(IdentityIntegrationTestSuite))
}

// SetupSuite runs once before all tests in the suite.
func (s *IdentityIntegrationTestSuite) SetupSuite() {
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
	// Use a deterministic validator address for score updates
	s.validator = sdktestutil.AccAddress(s.T())
}

// TestIdentityScopeUpload tests the identity scope upload flow.
//
// Flow:
//  1. Create test identity with document data
//  2. Encrypt identity scopes (selfie, document, liveness)
//  3. Submit identity upload transaction
//  4. Verify transaction is included in block
//  5. Query identity state from chain
func (s *IdentityIntegrationTestSuite) TestIdentityScopeUploadAndScoreCommit() {
	ctx := s.ctx

	owner := sdktestutil.AccAddress(s.T())
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
	userSignature := []byte("user-signature")

	msg := veidtypes.NewMsgUploadScope(
		owner.String(),
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
	ctx = s.app.NewContext(false)

	record, found := s.app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, owner)
	require.True(s.T(), found)
	require.Len(s.T(), record.ScopeRefs, 1)
	require.Equal(s.T(), scopeID, record.ScopeRefs[0].ScopeID)

	updateScore := veidtypes.NewMsgUpdateScore(
		s.validator.String(),
		owner.String(),
		82,
		"score-model-v1",
	)

	_, err = s.msgServer.UpdateScore(ctx, updateScore)
	require.NoError(s.T(), err)

	s.app.Commit()
	ctx = s.app.NewContext(false)

	updated, found := s.app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, owner)
	require.True(s.T(), found)
	require.Equal(s.T(), uint32(82), updated.CurrentScore)
	require.Equal(s.T(), "score-model-v1", updated.ScoreVersion)
}

// =============================================================================
// Test Data Structures
// =============================================================================

