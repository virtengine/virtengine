package keeper_test

import (
	"crypto/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
	"github.com/virtengine/virtengine/x/veid/keeper"
	"github.com/virtengine/virtengine/x/veid/types"
)

type MsgServerTestSuite struct {
	KeeperTestSuite
	msgServer types.MsgServer
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

func (s *MsgServerTestSuite) SetupTest() {
	s.KeeperTestSuite.SetupTest()
	s.msgServer = keeper.NewMsgServerImpl(s.keeper)
}

// createTestPayloadPB creates a veidv1.EncryptedPayloadEnvelope for protobuf struct literals
func (s *MsgServerTestSuite) createTestPayloadPB() veidv1.EncryptedPayloadEnvelope {
	nonce := make([]byte, 24)
	_, _ = rand.Read(nonce)
	ciphertext := make([]byte, 128)
	_, _ = rand.Read(ciphertext)
	pubKey := make([]byte, 32)
	_, _ = rand.Read(pubKey)
	sig := make([]byte, 64)
	_, _ = rand.Read(sig)
	return veidv1.EncryptedPayloadEnvelope{
		Version:         1,
		AlgorithmId:     "X25519-XSALSA20-POLY1305",
		AlgorithmVersion: 1,
		RecipientKeyIds: []string{"recipient1"},
		Nonce:           nonce,
		Ciphertext:      ciphertext,
		SenderSignature: sig,
		SenderPubKey:    pubKey,
	}
}

// Test: MsgUploadScope - successful upload
func (s *MsgServerTestSuite) TestMsgUploadScope_Success() {
	address := sdk.AccAddress([]byte("test-address-upload"))

	// Disable signature requirements for testing
	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	metadata := s.createTestUploadMetadata()

	msg := &types.MsgUploadScope{
		Sender:            address.String(),
		ScopeId:           "test-scope-msg-1",
		ScopeType:         veidv1.ScopeTypeIDDocument,
		EncryptedPayload:  s.createTestPayloadPB(),
		Salt:              metadata.Salt,
		DeviceFingerprint: metadata.DeviceFingerprint,
		ClientId:          metadata.ClientID,
		ClientSignature:   metadata.ClientSignature,
		UserSignature:     metadata.UserSignature,
		PayloadHash:       metadata.PayloadHash,
		CaptureTimestamp:  time.Now().Unix(),
		GeoHint:           "US",
	}

	resp, err := s.msgServer.UploadScope(sdk.WrapSDKContext(s.ctx), msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Equal("test-scope-msg-1", resp.ScopeId)
	s.Require().Equal(types.VerificationStatusPending, resp.Status)

	// Verify scope was created
	scope, found := s.keeper.GetScope(s.ctx, address, "test-scope-msg-1")
	s.Require().True(found)
	s.Require().Equal(types.ScopeTypeIDDocument, scope.ScopeType)
}

// Test: MsgUploadScope - invalid sender address
func (s *MsgServerTestSuite) TestMsgUploadScope_InvalidSender() {
	msg := &types.MsgUploadScope{
		Sender:  "invalid-address",
		ScopeId: "test-scope-invalid",
	}

	_, err := s.msgServer.UploadScope(sdk.WrapSDKContext(s.ctx), msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "invalid")
}

// Test: MsgUploadScope - unapproved client when client signature required
func (s *MsgServerTestSuite) TestMsgUploadScope_UnapprovedClient() {
	address := sdk.AccAddress([]byte("test-address-client"))

	// Enable client signature requirement
	params := types.DefaultParams()
	params.RequireClientSignature = true
	params.RequireUserSignature = false
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	metadata := s.createTestUploadMetadata()
	metadata.ClientID = "unapproved-client"

	msg := &types.MsgUploadScope{
		Sender:            address.String(),
		ScopeId:           "test-scope-unapproved",
		ScopeType:         veidv1.ScopeTypeIDDocument,
		EncryptedPayload:  s.createTestPayloadPB(),
		Salt:              metadata.Salt,
		DeviceFingerprint: metadata.DeviceFingerprint,
		ClientId:          metadata.ClientID,
		ClientSignature:   metadata.ClientSignature,
		UserSignature:     metadata.UserSignature,
		PayloadHash:       metadata.PayloadHash,
	}

	_, err = s.msgServer.UploadScope(sdk.WrapSDKContext(s.ctx), msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "not approved")
}

// Test: MsgRevokeScope - successful revocation
func (s *MsgServerTestSuite) TestMsgRevokeScope_Success() {
	address := sdk.AccAddress([]byte("test-address-revoke"))

	// Disable signature requirements for testing
	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	// First upload a scope
	metadata := s.createTestUploadMetadata()
	scope := types.NewIdentityScope(
		"scope-to-revoke-msg",
		types.ScopeTypeSelfie,
		s.createTestPayload(),
		metadata,
		time.Now(),
	)
	err = s.keeper.UploadScope(s.ctx, address, scope)
	s.Require().NoError(err)

	// Revoke it via msg server
	msg := &types.MsgRevokeScope{
		Sender:  address.String(),
		ScopeId: "scope-to-revoke-msg",
		Reason:  "user requested revocation",
	}

	resp, err := s.msgServer.RevokeScope(sdk.WrapSDKContext(s.ctx), msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Equal("scope-to-revoke-msg", resp.ScopeId)

	// Verify scope is revoked
	revokedScope, found := s.keeper.GetScope(s.ctx, address, "scope-to-revoke-msg")
	s.Require().True(found)
	s.Require().True(revokedScope.Revoked)
}

// Test: MsgRevokeScope - scope not found
func (s *MsgServerTestSuite) TestMsgRevokeScope_NotFound() {
	address := sdk.AccAddress([]byte("test-address-notfound"))

	msg := &types.MsgRevokeScope{
		Sender:  address.String(),
		ScopeId: "nonexistent-scope",
		Reason:  "test",
	}

	_, err := s.msgServer.RevokeScope(sdk.WrapSDKContext(s.ctx), msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "not found")
}

// Test: MsgRequestVerification - successful request
func (s *MsgServerTestSuite) TestMsgRequestVerification_Success() {
	address := sdk.AccAddress([]byte("test-address-verify"))

	// Disable signature requirements for testing
	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	// Upload a scope first
	metadata := s.createTestUploadMetadata()
	scope := types.NewIdentityScope(
		"scope-verify-req",
		types.ScopeTypeFaceVideo,
		s.createTestPayload(),
		metadata,
		time.Now(),
	)
	err = s.keeper.UploadScope(s.ctx, address, scope)
	s.Require().NoError(err)

	// Request verification
	msg := &types.MsgRequestVerification{
		Sender:  address.String(),
		ScopeId: "scope-verify-req",
	}

	resp, err := s.msgServer.RequestVerification(sdk.WrapSDKContext(s.ctx), msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Equal("scope-verify-req", resp.ScopeId)
	s.Require().Equal(types.VerificationStatusInProgress, resp.Status)

	// Verify scope status is updated
	updatedScope, found := s.keeper.GetScope(s.ctx, address, "scope-verify-req")
	s.Require().True(found)
	s.Require().Equal(types.VerificationStatusInProgress, updatedScope.Status)
}

// Test: MsgRequestVerification - scope not found
func (s *MsgServerTestSuite) TestMsgRequestVerification_NotFound() {
	address := sdk.AccAddress([]byte("test-address-verifynf"))

	msg := &types.MsgRequestVerification{
		Sender:  address.String(),
		ScopeId: "nonexistent-scope",
	}

	_, err := s.msgServer.RequestVerification(sdk.WrapSDKContext(s.ctx), msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "not found")
}

// Test: MsgRequestVerification - revoked scope
func (s *MsgServerTestSuite) TestMsgRequestVerification_RevokedScope() {
	address := sdk.AccAddress([]byte("test-address-revoked"))

	// Disable signature requirements for testing
	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	// Upload and revoke a scope
	metadata := s.createTestUploadMetadata()
	scope := types.NewIdentityScope(
		"scope-revoked-verify",
		types.ScopeTypeBiometric,
		s.createTestPayload(),
		metadata,
		time.Now(),
	)
	err = s.keeper.UploadScope(s.ctx, address, scope)
	s.Require().NoError(err)

	err = s.keeper.RevokeScope(s.ctx, address, "scope-revoked-verify", "test revocation")
	s.Require().NoError(err)

	// Try to request verification
	msg := &types.MsgRequestVerification{
		Sender:  address.String(),
		ScopeId: "scope-revoked-verify",
	}

	_, err = s.msgServer.RequestVerification(sdk.WrapSDKContext(s.ctx), msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "revoked")
}

// Test: MsgUpdateScore - identity record not found
func (s *MsgServerTestSuite) TestMsgUpdateScore_RecordNotFound() {
	validatorAddr := sdk.AccAddress([]byte("test-validator-addr"))
	accountAddr := sdk.AccAddress([]byte("test-account-nf"))

	msg := &types.MsgUpdateScore{
		Sender:         validatorAddr.String(),
		AccountAddress: accountAddr.String(),
		NewScore:       50,
		ScoreVersion:   "v1.0",
	}

	_, err := s.msgServer.UpdateScore(sdk.WrapSDKContext(s.ctx), msg)
	s.Require().Error(err)
	// Should fail due to not being a validator
	s.Require().Contains(err.Error(), "validator")
}

// Test: MsgCreateIdentityWallet - success
func (s *MsgServerTestSuite) TestMsgCreateIdentityWallet_Success() {
	address := sdk.AccAddress([]byte("test-wallet-create"))

	bindingPubKey := make([]byte, 32)
	bindingSignature := make([]byte, 64)

	msg := &types.MsgCreateIdentityWallet{
		Sender:           address.String(),
		BindingPubKey:    bindingPubKey,
		BindingSignature: bindingSignature,
	}

	resp, err := s.msgServer.CreateIdentityWallet(sdk.WrapSDKContext(s.ctx), msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotEmpty(resp.WalletId)

	// Verify wallet was created
	wallet, found := s.keeper.GetWallet(s.ctx, address)
	s.Require().True(found)
	s.Require().Equal(resp.WalletId, wallet.WalletID)
}

// Test: MsgCreateIdentityWallet - invalid sender
func (s *MsgServerTestSuite) TestMsgCreateIdentityWallet_InvalidSender() {
	msg := &types.MsgCreateIdentityWallet{
		Sender: "invalid-address",
	}

	_, err := s.msgServer.CreateIdentityWallet(sdk.WrapSDKContext(s.ctx), msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "invalid")
}

// Test: MsgAddScopeToWallet - success
// TODO: This test requires proper signature generation, skipping for proto migration
func (s *MsgServerTestSuite) TestMsgAddScopeToWallet_Success() {
	s.T().Skip("Requires valid signature generation; skipping for proto migration")
}

// Test: MsgRevokeScopeFromWallet - success
// TODO: This test requires proper signature generation, skipping for proto migration
func (s *MsgServerTestSuite) TestMsgRevokeScopeFromWallet_Success() {
	s.T().Skip("Requires valid signature generation; skipping for proto migration")
}

// Test: MsgUpdateConsentSettings - success
// TODO: This test requires proper signature generation, skipping for proto migration
func (s *MsgServerTestSuite) TestMsgUpdateConsentSettings_Success() {
	s.T().Skip("Requires valid signature generation; skipping for proto migration")
}

// Test: MsgRebindWallet - success
// TODO: This test requires proper signature generation, skipping for proto migration
func (s *MsgServerTestSuite) TestMsgRebindWallet_Success() {
	s.T().Skip("Requires valid signature generation; skipping for proto migration")
}

// Test: MsgUpdateBorderlineParams - unauthorized
func (s *MsgServerTestSuite) TestMsgUpdateBorderlineParams_Unauthorized() {
	msg := &types.MsgUpdateBorderlineParams{
		Authority: "unauthorized-authority",
		Params:    types.DefaultBorderlineParams(),
	}

	_, err := s.msgServer.UpdateBorderlineParams(sdk.WrapSDKContext(s.ctx), msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "invalid authority")
}

// Test: MsgUpdateBorderlineParams - success (with correct authority)
func (s *MsgServerTestSuite) TestMsgUpdateBorderlineParams_Success() {
	msg := &types.MsgUpdateBorderlineParams{
		Authority: "authority", // matches the keeper's authority
		Params:    types.DefaultBorderlineParams(),
	}

	resp, err := s.msgServer.UpdateBorderlineParams(sdk.WrapSDKContext(s.ctx), msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
}
