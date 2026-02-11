package keeper_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
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
		Version:          1,
		AlgorithmId:      "X25519-XSALSA20-POLY1305",
		AlgorithmVersion: 1,
		RecipientKeyIds:  []string{"recipient1"},
		Nonce:            nonce,
		Ciphertext:       ciphertext,
		SenderSignature:  sig,
		SenderPubKey:     pubKey,
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

	resp, err := s.msgServer.UploadScope(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Equal("test-scope-msg-1", resp.ScopeId)
	s.Require().Equal(veidv1.VerificationStatusPending, resp.Status)

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

	_, err := s.msgServer.UploadScope(s.ctx, msg)
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

	_, err = s.msgServer.UploadScope(s.ctx, msg)
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

	resp, err := s.msgServer.RevokeScope(s.ctx, msg)
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

	_, err := s.msgServer.RevokeScope(s.ctx, msg)
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

	resp, err := s.msgServer.RequestVerification(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Equal("scope-verify-req", resp.ScopeId)
	s.Require().Equal(veidv1.VerificationStatusInProgress, resp.Status)

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

	_, err := s.msgServer.RequestVerification(s.ctx, msg)
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

	_, err = s.msgServer.RequestVerification(s.ctx, msg)
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

	_, err := s.msgServer.UpdateScore(s.ctx, msg)
	s.Require().Error(err)
	// Should fail due to not being a validator
	s.Require().Contains(err.Error(), "validator")
}

// Test: MsgCreateIdentityWallet - success
func (s *MsgServerTestSuite) TestMsgCreateIdentityWallet_Success() {
	address := sdk.AccAddress([]byte("test-wallet-create"))
	kp := generateTestKeyPair()
	walletID := keeper.GenerateWalletID(address.String())
	bindingSignature := kp.signWalletBinding(walletID, address.String())

	msg := &types.MsgCreateIdentityWallet{
		Sender:           address.String(),
		BindingPubKey:    kp.pub,
		BindingSignature: bindingSignature,
	}

	resp, err := s.msgServer.CreateIdentityWallet(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotEmpty(resp.WalletId)
	s.Require().Equal(walletID, resp.WalletId)

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

	_, err := s.msgServer.CreateIdentityWallet(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "invalid")
}

// testKeyPair holds an Ed25519 key pair for test signing
type testKeyPair struct {
	pub  ed25519.PublicKey
	priv ed25519.PrivateKey
}

// generateTestKeyPair creates a deterministic Ed25519 key pair for testing
func generateTestKeyPair() testKeyPair {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic("failed to generate ed25519 key pair: " + err.Error())
	}
	return testKeyPair{pub: pub, priv: priv}
}

// signMessage signs a message using the same logic as verifySignature in wallet.go:
// if message is not 32 bytes, SHA256-hash it first, then sign with Ed25519.
func (kp testKeyPair) signMessage(message []byte) []byte {
	var msgToSign []byte
	if len(message) == 32 {
		msgToSign = message
	} else {
		hash := sha256.Sum256(message)
		msgToSign = hash[:]
	}
	return ed25519.Sign(kp.priv, msgToSign)
}

// signWalletBinding signs the wallet binding message: SHA256("VEID_WALLET_BINDING:" + walletID + ":" + accountAddress)
func (kp testKeyPair) signWalletBinding(walletID, accountAddress string) []byte {
	msg := types.GetWalletBindingMessage(walletID, accountAddress)
	return kp.signMessage(msg)
}

// signAddScope signs the add-scope message: "VEID_ADD_SCOPE:" + sender + ":" + scopeID
func (kp testKeyPair) signAddScope(sender, scopeID string) []byte {
	msg := types.GetAddScopeSigningMessage(sender, scopeID)
	return kp.signMessage(msg)
}

// signRevokeScope signs the revoke-scope message: "VEID_REVOKE_SCOPE:" + sender + ":" + scopeID
func (kp testKeyPair) signRevokeScope(sender, scopeID string) []byte {
	msg := types.GetRevokeScopeSigningMessage(sender, scopeID)
	return kp.signMessage(msg)
}

func (s *MsgServerTestSuite) grantScopeConsent(address sdk.AccAddress, scopeID, purpose string, kp testKeyPair) {
	update := types.ConsentUpdateRequest{
		ScopeID:      scopeID,
		GrantConsent: true,
		Purpose:      purpose,
	}
	consentMsg := []byte("VEID_CONSENT_UPDATE:" + address.String() + ":" + scopeID + ":grant")
	consentSig := kp.signMessage(consentMsg)
	err := s.keeper.UpdateConsent(s.ctx, address, update, consentSig)
	s.Require().NoError(err)
}

// signConsentUpdate signs the consent update message: "VEID_CONSENT_UPDATE:" + sender + ":" + scopeID + ":" + grant/revoke
func (kp testKeyPair) signConsentUpdate(sender, scopeID string, grant bool) []byte {
	grantStr := "revoke"
	if grant {
		grantStr = "grant"
	}
	msg := []byte("VEID_CONSENT_UPDATE:" + sender + ":" + scopeID + ":" + grantStr)
	return kp.signMessage(msg)
}

// signRebind signs the new public key with the old private key to authorize a rebind.
// The rebind verification expects verifySignature(oldPubKey, newPubKey, signature),
// where newPubKey is 32 bytes (used directly, no hashing).
func (kp testKeyPair) signRebind(newPubKey []byte) []byte {
	return kp.signMessage(newPubKey)
}

// createWalletWithKey creates an identity wallet bound to the given Ed25519 key pair
func (s *MsgServerTestSuite) createWalletWithKey(address sdk.AccAddress, kp testKeyPair) {
	walletID := keeper.GenerateWalletID(address.String())
	bindingSig := kp.signWalletBinding(walletID, address.String())

	msg := &types.MsgCreateIdentityWallet{
		Sender:           address.String(),
		BindingPubKey:    kp.pub,
		BindingSignature: bindingSig,
	}

	resp, err := s.msgServer.CreateIdentityWallet(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotEmpty(resp.WalletId)
}

// Test: MsgAddScopeToWallet - success
func (s *MsgServerTestSuite) TestMsgAddScopeToWallet_Success() {
	address := sdk.AccAddress([]byte("test-addr-add-scope1"))
	kp := generateTestKeyPair()

	// Disable upload signature requirements for testing
	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	// Create wallet with valid binding signature
	s.createWalletWithKey(address, kp)

	scopeID := "scope-wallet-add"
	envelopeHash := sha256.Sum256([]byte("test-envelope"))
	userSig := kp.signAddScope(address.String(), scopeID)

	s.grantScopeConsent(address, scopeID, "Identity verification", kp)

	msg := &types.MsgAddScopeToWallet{
		Sender:        address.String(),
		ScopeId:       scopeID,
		ScopeType:     veidv1.ScopeTypeIDDocument,
		EnvelopeHash:  envelopeHash[:],
		UserSignature: userSig,
	}

	resp, err := s.msgServer.AddScopeToWallet(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Equal(scopeID, resp.ScopeId)

	// Verify scope was added to wallet
	wallet, found := s.keeper.GetWallet(s.ctx, address)
	s.Require().True(found)
	_, exists := wallet.GetScopeReference(scopeID)
	s.Require().True(exists)
}

// Test: MsgRevokeScopeFromWallet - success
func (s *MsgServerTestSuite) TestMsgRevokeScopeFromWallet_Success() {
	address := sdk.AccAddress([]byte("test-addr-rev-scope"))
	kp := generateTestKeyPair()

	// Disable upload signature requirements for testing
	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	// Create wallet and add a scope
	s.createWalletWithKey(address, kp)

	scopeID := "scope-wallet-revoke"
	envelopeHash := sha256.Sum256([]byte("test-envelope-revoke"))
	addSig := kp.signAddScope(address.String(), scopeID)

	s.grantScopeConsent(address, scopeID, "Identity verification", kp)

	addMsg := &types.MsgAddScopeToWallet{
		Sender:        address.String(),
		ScopeId:       scopeID,
		ScopeType:     veidv1.ScopeTypeSelfie,
		EnvelopeHash:  envelopeHash[:],
		UserSignature: addSig,
	}

	_, err = s.msgServer.AddScopeToWallet(s.ctx, addMsg)
	s.Require().NoError(err)

	// Revoke the scope from wallet
	revokeSig := kp.signRevokeScope(address.String(), scopeID)

	revokeMsg := &types.MsgRevokeScopeFromWallet{
		Sender:        address.String(),
		ScopeId:       scopeID,
		Reason:        "user requested removal",
		UserSignature: revokeSig,
	}

	resp, err := s.msgServer.RevokeScopeFromWallet(s.ctx, revokeMsg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Equal(scopeID, resp.ScopeId)

	// Verify scope was revoked in wallet
	wallet, found := s.keeper.GetWallet(s.ctx, address)
	s.Require().True(found)
	ref, exists := wallet.GetScopeReference(scopeID)
	s.Require().True(exists)
	s.Require().Equal(types.ScopeRefStatusRevoked, ref.Status)
}

// Test: MsgUpdateConsentSettings - success
func (s *MsgServerTestSuite) TestMsgUpdateConsentSettings_Success() {
	address := sdk.AccAddress([]byte("test-addr-consent01"))
	kp := generateTestKeyPair()

	// Disable upload signature requirements for testing
	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	// Create wallet
	s.createWalletWithKey(address, kp)

	scopeID := "scope-consent-update"
	consentSig := kp.signConsentUpdate(address.String(), scopeID, true)

	msg := &types.MsgUpdateConsentSettings{
		Sender:        address.String(),
		ScopeId:       scopeID,
		GrantConsent:  true,
		Purpose:       "identity verification",
		UserSignature: consentSig,
	}

	resp, err := s.msgServer.UpdateConsentSettings(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().True(resp.ConsentVersion > 0)

	// Verify consent was updated in wallet
	wallet, found := s.keeper.GetWallet(s.ctx, address)
	s.Require().True(found)
	consent, hasConsent := wallet.ConsentSettings.GetScopeConsent(scopeID)
	s.Require().True(hasConsent)
	s.Require().True(consent.Granted)
}

// Test: MsgRebindWallet - success
func (s *MsgServerTestSuite) TestMsgRebindWallet_Success() {
	address := sdk.AccAddress([]byte("test-addr-rebind001"))
	oldKP := generateTestKeyPair()
	newKP := generateTestKeyPair()

	// Create wallet with old key pair
	s.createWalletWithKey(address, oldKP)

	// Old key signs the new public key to authorize the rebind
	oldSig := oldKP.signRebind(newKP.pub)

	// New key signs the wallet binding message
	walletID := keeper.GenerateWalletID(address.String())
	newBindingSig := newKP.signWalletBinding(walletID, address.String())

	msg := &types.MsgRebindWallet{
		Sender:              address.String(),
		NewBindingPubKey:    newKP.pub,
		NewBindingSignature: newBindingSig,
		OldSignature:        oldSig,
	}

	resp, err := s.msgServer.RebindWallet(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().True(resp.ReboundAt > 0)

	// Verify wallet was rebound with the new key
	wallet, found := s.keeper.GetWallet(s.ctx, address)
	s.Require().True(found)
	s.Require().Equal([]byte(newKP.pub), wallet.BindingPubKey)
}

// Test: MsgUpdateBorderlineParams - unauthorized
func (s *MsgServerTestSuite) TestMsgUpdateBorderlineParams_Unauthorized() {
	msg := &types.MsgUpdateBorderlineParams{
		Authority: "unauthorized-authority",
		Params:    types.DefaultBorderlineParams(),
	}

	_, err := s.msgServer.UpdateBorderlineParams(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "invalid authority")
}

// Test: MsgUpdateBorderlineParams - success (with correct authority)
func (s *MsgServerTestSuite) TestMsgUpdateBorderlineParams_Success() {
	msg := &types.MsgUpdateBorderlineParams{
		Authority: "authority", // matches the keeper's authority
		Params:    types.DefaultBorderlineParams(),
	}

	resp, err := s.msgServer.UpdateBorderlineParams(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
}
