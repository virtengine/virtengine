package keeper

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veid/types"
)

// testWalletSetup creates a test environment for wallet tests
type testWalletSetup struct {
	ctx     sdk.Context
	keeper  Keeper
	pubKey  ed25519.PublicKey
	privKey ed25519.PrivateKey
	address sdk.AccAddress
}

func setupWalletTest(t *testing.T) *testWalletSetup {
	t.Helper()

	// Create codec
	interfaceRegistry := cdctypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create store service using runtime
	storeService := runtime.NewKVStoreService(storeKey)

	// Create context with store
	ctx := sdk.Context{}.WithKVStore(storeService.OpenKVStore(sdk.Context{}))

	// For testing, we need a proper context with block time
	ctx = ctx.WithBlockTime(time.Now()).WithBlockHeight(100)

	// Create keeper
	keeper := NewKeeper(cdc, storeKey, "authority")

	// Generate test key pair
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// Generate test address
	address := sdk.AccAddress(pubKey[:20])

	return &testWalletSetup{
		ctx:     ctx,
		keeper:  keeper,
		pubKey:  pubKey,
		privKey: privKey,
		address: address,
	}
}

// signMessage signs a message with the test private key
func (ts *testWalletSetup) signMessage(message []byte) []byte {
	var msgToSign []byte
	if len(message) == 32 {
		msgToSign = message
	} else {
		hash := sha256.Sum256(message)
		msgToSign = hash[:]
	}
	return ed25519.Sign(ts.privKey, msgToSign)
}

// signWalletBinding signs the wallet binding message
func (ts *testWalletSetup) signWalletBinding(walletID string) []byte {
	msg := types.GetWalletBindingMessage(walletID, ts.address.String())
	return ed25519.Sign(ts.privKey, msg)
}

func TestCreateWallet(t *testing.T) {
	ts := setupWalletTest(t)

	// Generate wallet ID
	walletID := GenerateWalletID(ts.address.String())

	// Create binding signature
	bindingSignature := ts.signWalletBinding(walletID)

	// Create wallet
	wallet, err := ts.keeper.CreateWallet(ts.ctx, ts.address, bindingSignature, ts.pubKey)
	require.NoError(t, err)
	require.NotNil(t, wallet)

	// Verify wallet properties
	require.Equal(t, walletID, wallet.WalletID)
	require.Equal(t, ts.address.String(), wallet.AccountAddress)
	require.Equal(t, types.WalletStatusActive, wallet.Status)
	require.Equal(t, uint32(0), wallet.CurrentScore)
	require.Equal(t, types.AccountStatusUnknown, wallet.ScoreStatus)
	require.Equal(t, types.IdentityTierUnverified, wallet.Tier)
	require.NotEmpty(t, wallet.BindingSignature)
	require.NotEmpty(t, wallet.BindingPubKey)

	// Verify wallet can be retrieved
	retrieved, found := ts.keeper.GetWallet(ts.ctx, ts.address)
	require.True(t, found)
	require.Equal(t, wallet.WalletID, retrieved.WalletID)

	// Verify wallet can be retrieved by ID
	retrievedByID, found := ts.keeper.GetWalletByID(ts.ctx, walletID)
	require.True(t, found)
	require.Equal(t, wallet.WalletID, retrievedByID.WalletID)
}

func TestCreateWallet_AlreadyExists(t *testing.T) {
	ts := setupWalletTest(t)

	walletID := GenerateWalletID(ts.address.String())
	bindingSignature := ts.signWalletBinding(walletID)

	// Create first wallet
	_, err := ts.keeper.CreateWallet(ts.ctx, ts.address, bindingSignature, ts.pubKey)
	require.NoError(t, err)

	// Try to create second wallet for same address
	_, err = ts.keeper.CreateWallet(ts.ctx, ts.address, bindingSignature, ts.pubKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already exists")
}

func TestCreateWallet_InvalidSignature(t *testing.T) {
	ts := setupWalletTest(t)

	// Create invalid signature
	invalidSignature := make([]byte, 64)
	rand.Read(invalidSignature)

	// Try to create wallet with invalid signature
	_, err := ts.keeper.CreateWallet(ts.ctx, ts.address, invalidSignature, ts.pubKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "signature")
}

func TestAddScopeToWallet(t *testing.T) {
	ts := setupWalletTest(t)

	// Create wallet first
	walletID := GenerateWalletID(ts.address.String())
	bindingSignature := ts.signWalletBinding(walletID)
	_, err := ts.keeper.CreateWallet(ts.ctx, ts.address, bindingSignature, ts.pubKey)
	require.NoError(t, err)

	// Create scope reference
	scopeID := "scope_123"
	envelopeHash := sha256.Sum256([]byte("test envelope"))
	scopeRef := types.ScopeReference{
		ScopeID:        scopeID,
		ScopeType:      types.ScopeTypeIDDocument,
		EnvelopeHash:   envelopeHash[:],
		AddedAt:        time.Now(),
		Status:         types.ScopeRefStatusPending,
		ConsentGranted: false,
	}

	// Create signature for adding scope
	addScopeMsg := types.GetAddScopeSigningMessage(ts.address.String(), scopeID)
	addScopeSig := ts.signMessage(addScopeMsg)

	// Add scope to wallet
	err = ts.keeper.AddScopeToWallet(ts.ctx, ts.address, scopeRef, addScopeSig)
	require.NoError(t, err)

	// Verify scope was added
	wallet, found := ts.keeper.GetWallet(ts.ctx, ts.address)
	require.True(t, found)
	require.Len(t, wallet.ScopeRefs, 1)
	require.Equal(t, scopeID, wallet.ScopeRefs[0].ScopeID)
}

func TestAddScopeToWallet_WalletNotFound(t *testing.T) {
	ts := setupWalletTest(t)

	scopeRef := types.ScopeReference{
		ScopeID:      "scope_123",
		ScopeType:    types.ScopeTypeIDDocument,
		EnvelopeHash: make([]byte, 32),
		AddedAt:      time.Now(),
		Status:       types.ScopeRefStatusPending,
	}

	err := ts.keeper.AddScopeToWallet(ts.ctx, ts.address, scopeRef, []byte("signature"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "wallet not found")
}

func TestAddScopeToWallet_InvalidSignature(t *testing.T) {
	ts := setupWalletTest(t)

	// Create wallet
	walletID := GenerateWalletID(ts.address.String())
	bindingSignature := ts.signWalletBinding(walletID)
	_, err := ts.keeper.CreateWallet(ts.ctx, ts.address, bindingSignature, ts.pubKey)
	require.NoError(t, err)

	scopeRef := types.ScopeReference{
		ScopeID:      "scope_123",
		ScopeType:    types.ScopeTypeIDDocument,
		EnvelopeHash: make([]byte, 32),
		AddedAt:      time.Now(),
		Status:       types.ScopeRefStatusPending,
	}

	// Invalid signature
	invalidSig := make([]byte, 64)
	rand.Read(invalidSig)

	err = ts.keeper.AddScopeToWallet(ts.ctx, ts.address, scopeRef, invalidSig)
	require.Error(t, err)
	require.Contains(t, err.Error(), "signature")
}

func TestAddScopeToWallet_DuplicateScope(t *testing.T) {
	ts := setupWalletTest(t)

	// Create wallet
	walletID := GenerateWalletID(ts.address.String())
	bindingSignature := ts.signWalletBinding(walletID)
	_, err := ts.keeper.CreateWallet(ts.ctx, ts.address, bindingSignature, ts.pubKey)
	require.NoError(t, err)

	scopeID := "scope_123"
	envelopeHash := sha256.Sum256([]byte("test envelope"))
	scopeRef := types.ScopeReference{
		ScopeID:      scopeID,
		ScopeType:    types.ScopeTypeIDDocument,
		EnvelopeHash: envelopeHash[:],
		AddedAt:      time.Now(),
		Status:       types.ScopeRefStatusPending,
	}

	addScopeMsg := types.GetAddScopeSigningMessage(ts.address.String(), scopeID)
	addScopeSig := ts.signMessage(addScopeMsg)

	// Add scope first time
	err = ts.keeper.AddScopeToWallet(ts.ctx, ts.address, scopeRef, addScopeSig)
	require.NoError(t, err)

	// Try to add same scope again
	err = ts.keeper.AddScopeToWallet(ts.ctx, ts.address, scopeRef, addScopeSig)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already in wallet")
}

func TestRevokeScopeFromWallet(t *testing.T) {
	ts := setupWalletTest(t)

	// Create wallet and add scope
	walletID := GenerateWalletID(ts.address.String())
	bindingSignature := ts.signWalletBinding(walletID)
	_, err := ts.keeper.CreateWallet(ts.ctx, ts.address, bindingSignature, ts.pubKey)
	require.NoError(t, err)

	scopeID := "scope_123"
	envelopeHash := sha256.Sum256([]byte("test envelope"))
	scopeRef := types.ScopeReference{
		ScopeID:      scopeID,
		ScopeType:    types.ScopeTypeIDDocument,
		EnvelopeHash: envelopeHash[:],
		AddedAt:      time.Now(),
		Status:       types.ScopeRefStatusActive,
	}

	addScopeMsg := types.GetAddScopeSigningMessage(ts.address.String(), scopeID)
	addScopeSig := ts.signMessage(addScopeMsg)
	err = ts.keeper.AddScopeToWallet(ts.ctx, ts.address, scopeRef, addScopeSig)
	require.NoError(t, err)

	// Revoke the scope
	revokeScopeMsg := types.GetRevokeScopeSigningMessage(ts.address.String(), scopeID)
	revokeScopeSig := ts.signMessage(revokeScopeMsg)
	err = ts.keeper.RevokeScopeFromWallet(ts.ctx, ts.address, scopeID, "user requested", revokeScopeSig)
	require.NoError(t, err)

	// Verify scope was revoked
	wallet, found := ts.keeper.GetWallet(ts.ctx, ts.address)
	require.True(t, found)
	require.Len(t, wallet.ScopeRefs, 1)
	require.Equal(t, types.ScopeRefStatusRevoked, wallet.ScopeRefs[0].Status)
	require.Equal(t, "user requested", wallet.ScopeRefs[0].RevocationReason)
	require.NotNil(t, wallet.ScopeRefs[0].RevokedAt)
}

func TestRevokeScopeFromWallet_NotFound(t *testing.T) {
	ts := setupWalletTest(t)

	// Create wallet
	walletID := GenerateWalletID(ts.address.String())
	bindingSignature := ts.signWalletBinding(walletID)
	_, err := ts.keeper.CreateWallet(ts.ctx, ts.address, bindingSignature, ts.pubKey)
	require.NoError(t, err)

	// Try to revoke non-existent scope
	revokeScopeMsg := types.GetRevokeScopeSigningMessage(ts.address.String(), "nonexistent")
	revokeScopeSig := ts.signMessage(revokeScopeMsg)
	err = ts.keeper.RevokeScopeFromWallet(ts.ctx, ts.address, "nonexistent", "reason", revokeScopeSig)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not in wallet")
}

func TestRevokeScopeFromWallet_InvalidSignature(t *testing.T) {
	ts := setupWalletTest(t)

	// Create wallet and add scope
	walletID := GenerateWalletID(ts.address.String())
	bindingSignature := ts.signWalletBinding(walletID)
	_, err := ts.keeper.CreateWallet(ts.ctx, ts.address, bindingSignature, ts.pubKey)
	require.NoError(t, err)

	scopeID := "scope_123"
	envelopeHash := sha256.Sum256([]byte("test envelope"))
	scopeRef := types.ScopeReference{
		ScopeID:      scopeID,
		ScopeType:    types.ScopeTypeIDDocument,
		EnvelopeHash: envelopeHash[:],
		AddedAt:      time.Now(),
		Status:       types.ScopeRefStatusActive,
	}

	addScopeMsg := types.GetAddScopeSigningMessage(ts.address.String(), scopeID)
	addScopeSig := ts.signMessage(addScopeMsg)
	err = ts.keeper.AddScopeToWallet(ts.ctx, ts.address, scopeRef, addScopeSig)
	require.NoError(t, err)

	// Try to revoke with invalid signature
	invalidSig := make([]byte, 64)
	rand.Read(invalidSig)
	err = ts.keeper.RevokeScopeFromWallet(ts.ctx, ts.address, scopeID, "reason", invalidSig)
	require.Error(t, err)
	require.Contains(t, err.Error(), "signature")
}

func TestUpdateConsent(t *testing.T) {
	ts := setupWalletTest(t)

	// Create wallet
	walletID := GenerateWalletID(ts.address.String())
	bindingSignature := ts.signWalletBinding(walletID)
	_, err := ts.keeper.CreateWallet(ts.ctx, ts.address, bindingSignature, ts.pubKey)
	require.NoError(t, err)

	// Update consent for a scope
	scopeID := "scope_123"
	update := types.ConsentUpdateRequest{
		ScopeID:      scopeID,
		GrantConsent: true,
		Purpose:      "identity verification",
	}

	// Sign consent update
	consentMsg := []byte("VEID_CONSENT_UPDATE:" + ts.address.String() + ":" + scopeID + ":grant")
	consentSig := ts.signMessage(consentMsg)

	err = ts.keeper.UpdateConsent(ts.ctx, ts.address, update, consentSig)
	require.NoError(t, err)

	// Verify consent was updated
	wallet, found := ts.keeper.GetWallet(ts.ctx, ts.address)
	require.True(t, found)
	consent, found := wallet.ConsentSettings.GetScopeConsent(scopeID)
	require.True(t, found)
	require.True(t, consent.Granted)
	require.Equal(t, "identity verification", consent.Purpose)
}

func TestUpdateConsent_GlobalSettings(t *testing.T) {
	ts := setupWalletTest(t)

	// Create wallet
	walletID := GenerateWalletID(ts.address.String())
	bindingSignature := ts.signWalletBinding(walletID)
	_, err := ts.keeper.CreateWallet(ts.ctx, ts.address, bindingSignature, ts.pubKey)
	require.NoError(t, err)

	// Update global consent settings
	shareWithProviders := true
	shareForVerification := true
	update := types.ConsentUpdateRequest{
		GlobalSettings: &types.GlobalConsentUpdate{
			ShareWithProviders:   &shareWithProviders,
			ShareForVerification: &shareForVerification,
		},
	}

	// Sign consent update
	consentMsg := []byte("VEID_CONSENT_UPDATE:" + ts.address.String() + "::grant")
	consentSig := ts.signMessage(consentMsg)

	err = ts.keeper.UpdateConsent(ts.ctx, ts.address, update, consentSig)
	require.NoError(t, err)

	// Verify global settings were updated
	wallet, found := ts.keeper.GetWallet(ts.ctx, ts.address)
	require.True(t, found)
	require.True(t, wallet.ConsentSettings.ShareWithProviders)
	require.True(t, wallet.ConsentSettings.ShareForVerification)
}

func TestUpdateDerivedFeatures(t *testing.T) {
	ts := setupWalletTest(t)

	// Create wallet
	walletID := GenerateWalletID(ts.address.String())
	bindingSignature := ts.signWalletBinding(walletID)
	_, err := ts.keeper.CreateWallet(ts.ctx, ts.address, bindingSignature, ts.pubKey)
	require.NoError(t, err)

	// Create derived features update
	faceHash := sha256.Sum256([]byte("face embedding data"))
	nameHash := sha256.Sum256([]byte("John Doe"))

	update := &types.DerivedFeaturesUpdate{
		AccountAddress:    ts.address.String(),
		FaceEmbeddingHash: faceHash[:],
		DocFieldHashes: map[string][]byte{
			types.DocFieldNameHash: nameHash[:],
		},
		ModelVersion:     "v1.0.0",
		ValidatorAddress: "validator123",
	}

	err = ts.keeper.UpdateDerivedFeatures(ts.ctx, ts.address, update)
	require.NoError(t, err)

	// Verify features were updated
	wallet, found := ts.keeper.GetWallet(ts.ctx, ts.address)
	require.True(t, found)
	require.Equal(t, faceHash[:], wallet.DerivedFeatures.FaceEmbeddingHash)
	require.Equal(t, nameHash[:], wallet.DerivedFeatures.DocFieldHashes[types.DocFieldNameHash])
	require.Equal(t, "v1.0.0", wallet.DerivedFeatures.ModelVersion)
}

func TestUpdateDerivedFeatures_InvalidUpdate(t *testing.T) {
	ts := setupWalletTest(t)

	// Create wallet
	walletID := GenerateWalletID(ts.address.String())
	bindingSignature := ts.signWalletBinding(walletID)
	_, err := ts.keeper.CreateWallet(ts.ctx, ts.address, bindingSignature, ts.pubKey)
	require.NoError(t, err)

	// Invalid update - wrong hash size
	update := &types.DerivedFeaturesUpdate{
		AccountAddress:    ts.address.String(),
		FaceEmbeddingHash: []byte("too short"),
		ModelVersion:      "v1.0.0",
		ValidatorAddress:  "validator123",
	}

	err = ts.keeper.UpdateDerivedFeatures(ts.ctx, ts.address, update)
	require.Error(t, err)
	require.Contains(t, err.Error(), "32 bytes")
}

func TestRebindWallet(t *testing.T) {
	ts := setupWalletTest(t)

	// Create wallet
	walletID := GenerateWalletID(ts.address.String())
	bindingSignature := ts.signWalletBinding(walletID)
	_, err := ts.keeper.CreateWallet(ts.ctx, ts.address, bindingSignature, ts.pubKey)
	require.NoError(t, err)

	// Generate new key pair for rebind
	newPubKey, newPrivKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// Old key signs the new public key
	oldSignature := ed25519.Sign(ts.privKey, newPubKey)

	// New key creates new binding signature
	newBindingMsg := types.GetWalletBindingMessage(walletID, ts.address.String())
	newBindingSignature := ed25519.Sign(newPrivKey, newBindingMsg)

	// Rebind wallet
	err = ts.keeper.RebindWallet(ts.ctx, ts.address, newBindingSignature, newPubKey, oldSignature)
	require.NoError(t, err)

	// Verify wallet was rebound
	wallet, found := ts.keeper.GetWallet(ts.ctx, ts.address)
	require.True(t, found)
	require.Equal(t, []byte(newPubKey), wallet.BindingPubKey)
	require.Equal(t, newBindingSignature, wallet.BindingSignature)
}

func TestRebindWallet_InvalidOldSignature(t *testing.T) {
	ts := setupWalletTest(t)

	// Create wallet
	walletID := GenerateWalletID(ts.address.String())
	bindingSignature := ts.signWalletBinding(walletID)
	_, err := ts.keeper.CreateWallet(ts.ctx, ts.address, bindingSignature, ts.pubKey)
	require.NoError(t, err)

	// Generate new key pair
	newPubKey, newPrivKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// Invalid old signature
	invalidOldSig := make([]byte, 64)
	rand.Read(invalidOldSig)

	newBindingMsg := types.GetWalletBindingMessage(walletID, ts.address.String())
	newBindingSignature := ed25519.Sign(newPrivKey, newBindingMsg)

	err = ts.keeper.RebindWallet(ts.ctx, ts.address, newBindingSignature, newPubKey, invalidOldSig)
	require.Error(t, err)
	require.Contains(t, err.Error(), "signature")
}

func TestGetWalletPublicMetadata(t *testing.T) {
	ts := setupWalletTest(t)

	// Create wallet
	walletID := GenerateWalletID(ts.address.String())
	bindingSignature := ts.signWalletBinding(walletID)
	_, err := ts.keeper.CreateWallet(ts.ctx, ts.address, bindingSignature, ts.pubKey)
	require.NoError(t, err)

	// Get public metadata
	info, found := ts.keeper.GetWalletPublicMetadata(ts.ctx, ts.address)
	require.True(t, found)
	require.Equal(t, walletID, info.WalletID)
	require.Equal(t, ts.address.String(), info.AccountAddress)
	require.Equal(t, types.WalletStatusActive, info.Status)
	require.Equal(t, uint32(0), info.CurrentScore)
	require.Equal(t, 0, info.ScopeCount)
}

func TestGetWalletPublicMetadata_NotFound(t *testing.T) {
	ts := setupWalletTest(t)

	_, found := ts.keeper.GetWalletPublicMetadata(ts.ctx, ts.address)
	require.False(t, found)
}

func TestUpdateWalletScore(t *testing.T) {
	ts := setupWalletTest(t)

	// Create wallet
	walletID := GenerateWalletID(ts.address.String())
	bindingSignature := ts.signWalletBinding(walletID)
	_, err := ts.keeper.CreateWallet(ts.ctx, ts.address, bindingSignature, ts.pubKey)
	require.NoError(t, err)

	// Update score
	err = ts.keeper.UpdateWalletScore(
		ts.ctx,
		ts.address,
		75,
		types.AccountStatusVerified,
		"v1.0.0",
		"validator123",
		[]string{"scope_1", "scope_2"},
		"initial verification",
	)
	require.NoError(t, err)

	// Verify score was updated
	wallet, found := ts.keeper.GetWallet(ts.ctx, ts.address)
	require.True(t, found)
	require.Equal(t, uint32(75), wallet.CurrentScore)
	require.Equal(t, types.AccountStatusVerified, wallet.ScoreStatus)
	require.Equal(t, types.IdentityTierStandard, wallet.Tier)

	// Verify history entry was added
	require.Len(t, wallet.VerificationHistory, 1)
	require.Equal(t, uint32(0), wallet.VerificationHistory[0].PreviousScore)
	require.Equal(t, uint32(75), wallet.VerificationHistory[0].NewScore)
	require.Equal(t, "v1.0.0", wallet.VerificationHistory[0].ModelVersion)
	require.Len(t, wallet.VerificationHistory[0].ScopesEvaluated, 2)
}

func TestWalletConsentFlow(t *testing.T) {
	ts := setupWalletTest(t)

	// Create wallet
	walletID := GenerateWalletID(ts.address.String())
	bindingSignature := ts.signWalletBinding(walletID)
	_, err := ts.keeper.CreateWallet(ts.ctx, ts.address, bindingSignature, ts.pubKey)
	require.NoError(t, err)

	// Add a scope
	scopeID := "scope_kyc"
	envelopeHash := sha256.Sum256([]byte("kyc data"))
	scopeRef := types.ScopeReference{
		ScopeID:      scopeID,
		ScopeType:    types.ScopeTypeIDDocument,
		EnvelopeHash: envelopeHash[:],
		AddedAt:      time.Now(),
		Status:       types.ScopeRefStatusActive,
	}

	addScopeMsg := types.GetAddScopeSigningMessage(ts.address.String(), scopeID)
	addScopeSig := ts.signMessage(addScopeMsg)
	err = ts.keeper.AddScopeToWallet(ts.ctx, ts.address, scopeRef, addScopeSig)
	require.NoError(t, err)

	// Grant consent
	grantUpdate := types.ConsentUpdateRequest{
		ScopeID:      scopeID,
		GrantConsent: true,
		Purpose:      "KYC verification",
	}
	grantMsg := []byte("VEID_CONSENT_UPDATE:" + ts.address.String() + ":" + scopeID + ":grant")
	grantSig := ts.signMessage(grantMsg)
	err = ts.keeper.UpdateConsent(ts.ctx, ts.address, grantUpdate, grantSig)
	require.NoError(t, err)

	// Verify consent is active
	wallet, _ := ts.keeper.GetWallet(ts.ctx, ts.address)
	require.True(t, wallet.ConsentSettings.IsScopeConsentActive(scopeID))

	// Revoke consent
	revokeUpdate := types.ConsentUpdateRequest{
		ScopeID:      scopeID,
		GrantConsent: false,
	}
	revokeMsg := []byte("VEID_CONSENT_UPDATE:" + ts.address.String() + ":" + scopeID + ":revoke")
	revokeSig := ts.signMessage(revokeMsg)
	err = ts.keeper.UpdateConsent(ts.ctx, ts.address, revokeUpdate, revokeSig)
	require.NoError(t, err)

	// Verify consent is revoked
	wallet, _ = ts.keeper.GetWallet(ts.ctx, ts.address)
	require.False(t, wallet.ConsentSettings.IsScopeConsentActive(scopeID))
}

func TestWalletTierCalculation(t *testing.T) {
	tests := []struct {
		score    uint32
		status   types.AccountStatus
		expected types.IdentityTier
	}{
		{0, types.AccountStatusUnknown, types.IdentityTierUnverified},
		{50, types.AccountStatusPending, types.IdentityTierUnverified},
		{50, types.AccountStatusVerified, types.IdentityTierBasic},
		{69, types.AccountStatusVerified, types.IdentityTierBasic},
		{70, types.AccountStatusVerified, types.IdentityTierStandard},
		{84, types.AccountStatusVerified, types.IdentityTierStandard},
		{85, types.AccountStatusVerified, types.IdentityTierPremium},
		{100, types.AccountStatusVerified, types.IdentityTierPremium},
		{100, types.AccountStatusRejected, types.IdentityTierUnverified},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			tier := types.TierFromScore(tt.score, tt.status)
			require.Equal(t, tt.expected, tier)
		})
	}
}

func TestCountWallets(t *testing.T) {
	ts := setupWalletTest(t)

	// Initially no wallets
	count := ts.keeper.CountWallets(ts.ctx)
	require.Equal(t, 0, count)

	// Create first wallet
	walletID := GenerateWalletID(ts.address.String())
	bindingSignature := ts.signWalletBinding(walletID)
	_, err := ts.keeper.CreateWallet(ts.ctx, ts.address, bindingSignature, ts.pubKey)
	require.NoError(t, err)

	count = ts.keeper.CountWallets(ts.ctx)
	require.Equal(t, 1, count)

	// Create second wallet with different address
	pubKey2, privKey2, _ := ed25519.GenerateKey(rand.Reader)
	address2 := sdk.AccAddress(pubKey2[:20])
	walletID2 := GenerateWalletID(address2.String())
	bindingMsg2 := types.GetWalletBindingMessage(walletID2, address2.String())
	bindingSig2 := ed25519.Sign(privKey2, bindingMsg2)
	_, err = ts.keeper.CreateWallet(ts.ctx, address2, bindingSig2, pubKey2)
	require.NoError(t, err)

	count = ts.keeper.CountWallets(ts.ctx)
	require.Equal(t, 2, count)
}
