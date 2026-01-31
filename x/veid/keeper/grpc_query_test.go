package keeper

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/virtengine/virtengine/x/veid/types"
)

// grpcQueryTestSuite provides test setup for gRPC query tests
type grpcQueryTestSuite struct {
	ctx        sdk.Context
	keeper     Keeper
	querier    GRPCQuerier
	pubKey     ed25519.PublicKey
	privKey    ed25519.PrivateKey
	address    sdk.AccAddress
	stateStore store.CommitMultiStore
}

func setupGRPCQueryTest(t *testing.T) *grpcQueryTestSuite {
	t.Helper()

	// Create codec
	interfaceRegistry := cdctypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create in-memory store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	require.NoError(t, err)

	// Create context with store
	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())

	// Create keeper
	keeper := NewKeeper(cdc, storeKey, "authority")

	// Generate test key pair
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// Generate test address
	address := sdk.AccAddress(pubKey[:20])

	ts := &grpcQueryTestSuite{
		ctx:        ctx,
		keeper:     keeper,
		querier:    GRPCQuerier{Keeper: keeper},
		pubKey:     pubKey,
		privKey:    privKey,
		address:    address,
		stateStore: stateStore,
	}

	t.Cleanup(func() {
		if closer, ok := ts.stateStore.(io.Closer); ok {
			closer.Close()
		}
	})

	return ts
}

// signWalletBinding signs the wallet binding message
func (ts *grpcQueryTestSuite) signWalletBinding(walletID string) []byte {
	msg := types.GetWalletBindingMessage(walletID, ts.address.String())
	return ed25519.Sign(ts.privKey, msg)
}

// createTestWallet creates a wallet for testing
func (ts *grpcQueryTestSuite) createTestWallet(t *testing.T) *types.IdentityWallet {
	t.Helper()

	walletID := GenerateWalletID(ts.address.String())
	bindingSignature := ts.signWalletBinding(walletID)

	wallet, err := ts.keeper.CreateWallet(ts.ctx, ts.address, bindingSignature, ts.pubKey)
	require.NoError(t, err)
	require.NotNil(t, wallet)

	return wallet
}

// createTestWalletWithScopes creates a wallet with scope references
func (ts *grpcQueryTestSuite) createTestWalletWithScopes(t *testing.T) *types.IdentityWallet {
	t.Helper()

	wallet := ts.createTestWallet(t)

	// Add scope references
	scopeRef1 := types.ScopeReference{
		ScopeID:        "scope-1",
		ScopeType:      types.ScopeTypeSelfie,
		EnvelopeHash:   make([]byte, 32),
		AddedAt:        ts.ctx.BlockTime(),
		Status:         types.ScopeRefStatusActive,
		ConsentGranted: true,
	}
	scopeRef2 := types.ScopeReference{
		ScopeID:        "scope-2",
		ScopeType:      types.ScopeTypeIDDocument,
		EnvelopeHash:   make([]byte, 32),
		AddedAt:        ts.ctx.BlockTime(),
		Status:         types.ScopeRefStatusActive,
		ConsentGranted: false,
	}
	scopeRef3 := types.ScopeReference{
		ScopeID:        "scope-3",
		ScopeType:      types.ScopeTypeSelfie,
		EnvelopeHash:   make([]byte, 32),
		AddedAt:        ts.ctx.BlockTime(),
		Status:         types.ScopeRefStatusRevoked,
		ConsentGranted: false,
	}

	wallet.ScopeRefs = []types.ScopeReference{scopeRef1, scopeRef2, scopeRef3}
	err := ts.keeper.SetWallet(ts.ctx, wallet)
	require.NoError(t, err)

	return wallet
}

// createTestWalletWithConsent creates a wallet with consent settings
func (ts *grpcQueryTestSuite) createTestWalletWithConsent(t *testing.T) *types.IdentityWallet {
	t.Helper()

	wallet := ts.createTestWallet(t)

	// Set consent settings
	wallet.ConsentSettings.ShareWithProviders = true
	wallet.ConsentSettings.ShareForVerification = true
	wallet.ConsentSettings.AllowDerivedFeatureSharing = true
	wallet.ConsentSettings.GrantScopeConsentAt("scope-1", "verification", nil, ts.ctx.BlockTime())

	err := ts.keeper.SetWallet(ts.ctx, wallet)
	require.NoError(t, err)

	return wallet
}

// createTestWalletWithDerivedFeatures creates a wallet with derived features
func (ts *grpcQueryTestSuite) createTestWalletWithDerivedFeatures(t *testing.T) *types.IdentityWallet {
	t.Helper()

	wallet := ts.createTestWallet(t)

	// Set derived features
	hash := sha256.Sum256([]byte("test-face-embedding"))
	wallet.DerivedFeatures = types.DerivedFeatures{
		FaceEmbeddingHash: hash[:],
		DocFieldHashes: map[string][]byte{
			types.DocFieldNameHash: hash[:],
			types.DocFieldDOBHash:  hash[:],
		},
		BiometricHash:  hash[:],
		ModelVersion:   "v1.0.0",
		LastComputedAt: ts.ctx.BlockTime(),
		FeatureVersion: 1,
	}
	wallet.ConsentSettings.AllowDerivedFeatureSharing = true

	err := ts.keeper.SetWallet(ts.ctx, wallet)
	require.NoError(t, err)

	return wallet
}

// createTestWalletWithHistory creates a wallet with verification history
func (ts *grpcQueryTestSuite) createTestWalletWithHistory(t *testing.T) *types.IdentityWallet {
	t.Helper()

	wallet := ts.createTestWallet(t)

	// Add verification history
	wallet.VerificationHistory = []types.VerificationHistoryEntry{
		{
			EntryID:          "entry-1",
			Timestamp:        ts.ctx.BlockTime().Add(-2 * time.Hour),
			BlockHeight:      90,
			PreviousScore:    0,
			NewScore:         50,
			PreviousStatus:   types.AccountStatusUnknown,
			NewStatus:        types.AccountStatusPending,
			ScopesEvaluated:  []string{"scope-1"},
			ModelVersion:     "v1.0.0",
			ValidatorAddress: "validator1",
		},
		{
			EntryID:          "entry-2",
			Timestamp:        ts.ctx.BlockTime().Add(-1 * time.Hour),
			BlockHeight:      95,
			PreviousScore:    50,
			NewScore:         80,
			PreviousStatus:   types.AccountStatusPending,
			NewStatus:        types.AccountStatusVerified,
			ScopesEvaluated:  []string{"scope-1", "scope-2"},
			ModelVersion:     "v1.0.0",
			ValidatorAddress: "validator1",
		},
		{
			EntryID:          "entry-3",
			Timestamp:        ts.ctx.BlockTime(),
			BlockHeight:      100,
			PreviousScore:    80,
			NewScore:         90,
			PreviousStatus:   types.AccountStatusVerified,
			NewStatus:        types.AccountStatusVerified,
			ScopesEvaluated:  []string{"scope-1", "scope-2", "scope-3"},
			ModelVersion:     "v1.1.0",
			ValidatorAddress: "validator2",
		},
	}

	err := ts.keeper.SetWallet(ts.ctx, wallet)
	require.NoError(t, err)

	return wallet
}

// ============================================================================
// IdentityWallet Query Tests
// ============================================================================

func TestGRPCQuerier_IdentityWallet_NilRequest(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	resp, err := ts.querier.IdentityWallet(ts.ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGRPCQuerier_IdentityWallet_EmptyAddress(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	resp, err := ts.querier.IdentityWallet(ts.ctx, &types.QueryIdentityWalletRequest{AccountAddress: ""})
	require.Error(t, err)
	require.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGRPCQuerier_IdentityWallet_InvalidAddress(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	resp, err := ts.querier.IdentityWallet(ts.ctx, &types.QueryIdentityWalletRequest{AccountAddress: "invalid"})
	require.Error(t, err)
	require.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGRPCQuerier_IdentityWallet_NotFound(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	resp, err := ts.querier.IdentityWallet(ts.ctx, &types.QueryIdentityWalletRequest{
		AccountAddress: ts.address.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.False(t, resp.Found)
}

func TestGRPCQuerier_IdentityWallet_Found(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	ts.createTestWallet(t)

	resp, err := ts.querier.IdentityWallet(ts.ctx, &types.QueryIdentityWalletRequest{
		AccountAddress: ts.address.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.True(t, resp.Found)
	require.NotNil(t, resp.Wallet)
	require.Equal(t, ts.address.String(), resp.Wallet.AccountAddress)
}

// ============================================================================
// WalletScopes Query Tests
// ============================================================================

func TestGRPCQuerier_WalletScopes_NilRequest(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	resp, err := ts.querier.WalletScopes(ts.ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGRPCQuerier_WalletScopes_EmptyAddress(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	resp, err := ts.querier.WalletScopes(ts.ctx, &types.QueryWalletScopesRequest{AccountAddress: ""})
	require.Error(t, err)
	require.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGRPCQuerier_WalletScopes_InvalidAddress(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	resp, err := ts.querier.WalletScopes(ts.ctx, &types.QueryWalletScopesRequest{AccountAddress: "invalid"})
	require.Error(t, err)
	require.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGRPCQuerier_WalletScopes_NotFound(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	resp, err := ts.querier.WalletScopes(ts.ctx, &types.QueryWalletScopesRequest{
		AccountAddress: ts.address.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 0, resp.TotalCount)
	require.Equal(t, 0, resp.ActiveCount)
	require.Empty(t, resp.Scopes)
}

func TestGRPCQuerier_WalletScopes_AllScopes(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	ts.createTestWalletWithScopes(t)

	resp, err := ts.querier.WalletScopes(ts.ctx, &types.QueryWalletScopesRequest{
		AccountAddress: ts.address.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 3, resp.TotalCount)
	require.Equal(t, 2, resp.ActiveCount)
	require.Len(t, resp.Scopes, 3)
}

func TestGRPCQuerier_WalletScopes_ActiveOnly(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	ts.createTestWalletWithScopes(t)

	resp, err := ts.querier.WalletScopes(ts.ctx, &types.QueryWalletScopesRequest{
		AccountAddress: ts.address.String(),
		ActiveOnly:     true,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 3, resp.TotalCount)
	require.Equal(t, 2, resp.ActiveCount)
	require.Len(t, resp.Scopes, 2)
}

func TestGRPCQuerier_WalletScopes_FilterByType(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	ts.createTestWalletWithScopes(t)

	resp, err := ts.querier.WalletScopes(ts.ctx, &types.QueryWalletScopesRequest{
		AccountAddress: ts.address.String(),
		ScopeType:      string(types.ScopeTypeSelfie),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 3, resp.TotalCount)
	require.Len(t, resp.Scopes, 2)
}

func TestGRPCQuerier_WalletScopes_FilterByStatus(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	ts.createTestWalletWithScopes(t)

	resp, err := ts.querier.WalletScopes(ts.ctx, &types.QueryWalletScopesRequest{
		AccountAddress: ts.address.String(),
		StatusFilter:   string(types.ScopeRefStatusRevoked),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 3, resp.TotalCount)
	require.Len(t, resp.Scopes, 1)
}

// ============================================================================
// ConsentSettings Query Tests
// ============================================================================

func TestGRPCQuerier_ConsentSettings_NilRequest(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	resp, err := ts.querier.ConsentSettings(ts.ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGRPCQuerier_ConsentSettings_EmptyAddress(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	resp, err := ts.querier.ConsentSettings(ts.ctx, &types.QueryConsentSettingsRequest{AccountAddress: ""})
	require.Error(t, err)
	require.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGRPCQuerier_ConsentSettings_InvalidAddress(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	resp, err := ts.querier.ConsentSettings(ts.ctx, &types.QueryConsentSettingsRequest{AccountAddress: "invalid"})
	require.Error(t, err)
	require.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGRPCQuerier_ConsentSettings_NotFound(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	resp, err := ts.querier.ConsentSettings(ts.ctx, &types.QueryConsentSettingsRequest{
		AccountAddress: ts.address.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.False(t, resp.GlobalSettings.ShareWithProviders)
	require.Empty(t, resp.ScopeConsents)
}

func TestGRPCQuerier_ConsentSettings_Found(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	ts.createTestWalletWithConsent(t)

	resp, err := ts.querier.ConsentSettings(ts.ctx, &types.QueryConsentSettingsRequest{
		AccountAddress: ts.address.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.True(t, resp.GlobalSettings.ShareWithProviders)
	require.True(t, resp.GlobalSettings.ShareForVerification)
	require.Len(t, resp.ScopeConsents, 1)
}

func TestGRPCQuerier_ConsentSettings_FilterByScope(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	ts.createTestWalletWithConsent(t)

	resp, err := ts.querier.ConsentSettings(ts.ctx, &types.QueryConsentSettingsRequest{
		AccountAddress: ts.address.String(),
		ScopeID:        "scope-1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.ScopeConsents, 1)
	require.Equal(t, "scope-1", resp.ScopeConsents[0].ScopeID)
}

func TestGRPCQuerier_ConsentSettings_FilterByScopeNotFound(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	ts.createTestWalletWithConsent(t)

	resp, err := ts.querier.ConsentSettings(ts.ctx, &types.QueryConsentSettingsRequest{
		AccountAddress: ts.address.String(),
		ScopeID:        "non-existent",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Empty(t, resp.ScopeConsents)
}

// ============================================================================
// DerivedFeatures Query Tests
// ============================================================================

func TestGRPCQuerier_DerivedFeatures_NilRequest(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	resp, err := ts.querier.DerivedFeatures(ts.ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGRPCQuerier_DerivedFeatures_EmptyAddress(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	resp, err := ts.querier.DerivedFeatures(ts.ctx, &types.QueryDerivedFeaturesRequest{AccountAddress: ""})
	require.Error(t, err)
	require.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGRPCQuerier_DerivedFeatures_InvalidAddress(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	resp, err := ts.querier.DerivedFeatures(ts.ctx, &types.QueryDerivedFeaturesRequest{AccountAddress: "invalid"})
	require.Error(t, err)
	require.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGRPCQuerier_DerivedFeatures_NotFound(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	resp, err := ts.querier.DerivedFeatures(ts.ctx, &types.QueryDerivedFeaturesRequest{
		AccountAddress: ts.address.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.False(t, resp.Found)
}

func TestGRPCQuerier_DerivedFeatures_EmptyFeatures(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	ts.createTestWallet(t)

	resp, err := ts.querier.DerivedFeatures(ts.ctx, &types.QueryDerivedFeaturesRequest{
		AccountAddress: ts.address.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.True(t, resp.Found)
	require.NotNil(t, resp.Features)
	require.False(t, resp.Features.HasFaceEmbedding)
	require.False(t, resp.Features.HasBiometric)
}

func TestGRPCQuerier_DerivedFeatures_WithFeatures(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	ts.createTestWalletWithDerivedFeatures(t)

	resp, err := ts.querier.DerivedFeatures(ts.ctx, &types.QueryDerivedFeaturesRequest{
		AccountAddress: ts.address.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.True(t, resp.Found)
	require.NotNil(t, resp.Features)
	require.True(t, resp.Features.HasFaceEmbedding)
	require.True(t, resp.Features.HasBiometric)
	require.Len(t, resp.Features.DocFieldKeys, 2)
	require.Equal(t, "v1.0.0", resp.Features.ModelVersion)
}

// ============================================================================
// DerivedFeatureHashes Query Tests
// ============================================================================

func TestGRPCQuerier_DerivedFeatureHashes_NilRequest(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	resp, err := ts.querier.DerivedFeatureHashes(ts.ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGRPCQuerier_DerivedFeatureHashes_EmptyAddress(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	resp, err := ts.querier.DerivedFeatureHashes(ts.ctx, &types.QueryDerivedFeatureHashesRequest{AccountAddress: ""})
	require.Error(t, err)
	require.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGRPCQuerier_DerivedFeatureHashes_InvalidAddress(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	resp, err := ts.querier.DerivedFeatureHashes(ts.ctx, &types.QueryDerivedFeatureHashesRequest{AccountAddress: "invalid"})
	require.Error(t, err)
	require.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGRPCQuerier_DerivedFeatureHashes_NotFound(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	resp, err := ts.querier.DerivedFeatureHashes(ts.ctx, &types.QueryDerivedFeatureHashesRequest{
		AccountAddress: ts.address.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.False(t, resp.Allowed)
	require.Equal(t, "wallet not found", resp.DenialReason)
}

func TestGRPCQuerier_DerivedFeatureHashes_ConsentDenied(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	wallet := ts.createTestWalletWithDerivedFeatures(t)
	wallet.ConsentSettings.AllowDerivedFeatureSharing = false
	err := ts.keeper.SetWallet(ts.ctx, wallet)
	require.NoError(t, err)

	resp, err := ts.querier.DerivedFeatureHashes(ts.ctx, &types.QueryDerivedFeatureHashesRequest{
		AccountAddress: ts.address.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.False(t, resp.Allowed)
	require.Equal(t, "derived feature sharing not allowed", resp.DenialReason)
}

func TestGRPCQuerier_DerivedFeatureHashes_ConsentGranted(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	ts.createTestWalletWithDerivedFeatures(t)

	resp, err := ts.querier.DerivedFeatureHashes(ts.ctx, &types.QueryDerivedFeatureHashesRequest{
		AccountAddress: ts.address.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.True(t, resp.Allowed)
	require.Empty(t, resp.DenialReason)
	require.NotEmpty(t, resp.FaceEmbeddingHash)
	require.NotEmpty(t, resp.BiometricHash)
	require.NotEmpty(t, resp.DocFieldHashes)
	require.Equal(t, "v1.0.0", resp.ModelVersion)
}

// ============================================================================
// VerificationHistory Query Tests
// ============================================================================

func TestGRPCQuerier_VerificationHistory_NilRequest(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	resp, err := ts.querier.VerificationHistory(ts.ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGRPCQuerier_VerificationHistory_EmptyAddress(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	resp, err := ts.querier.VerificationHistory(ts.ctx, &types.QueryVerificationHistoryRequest{AccountAddress: ""})
	require.Error(t, err)
	require.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGRPCQuerier_VerificationHistory_InvalidAddress(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	resp, err := ts.querier.VerificationHistory(ts.ctx, &types.QueryVerificationHistoryRequest{AccountAddress: "invalid"})
	require.Error(t, err)
	require.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGRPCQuerier_VerificationHistory_NotFound(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	resp, err := ts.querier.VerificationHistory(ts.ctx, &types.QueryVerificationHistoryRequest{
		AccountAddress: ts.address.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 0, resp.TotalCount)
	require.Empty(t, resp.Entries)
}

func TestGRPCQuerier_VerificationHistory_EmptyHistory(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	ts.createTestWallet(t)

	resp, err := ts.querier.VerificationHistory(ts.ctx, &types.QueryVerificationHistoryRequest{
		AccountAddress: ts.address.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 0, resp.TotalCount)
	require.Empty(t, resp.Entries)
}

func TestGRPCQuerier_VerificationHistory_AllEntries(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	ts.createTestWalletWithHistory(t)

	resp, err := ts.querier.VerificationHistory(ts.ctx, &types.QueryVerificationHistoryRequest{
		AccountAddress: ts.address.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 3, resp.TotalCount)
	require.Len(t, resp.Entries, 3)
}

func TestGRPCQuerier_VerificationHistory_WithLimit(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	ts.createTestWalletWithHistory(t)

	resp, err := ts.querier.VerificationHistory(ts.ctx, &types.QueryVerificationHistoryRequest{
		AccountAddress: ts.address.String(),
		Limit:          2,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 3, resp.TotalCount)
	require.Len(t, resp.Entries, 2)
}

func TestGRPCQuerier_VerificationHistory_WithOffset(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	ts.createTestWalletWithHistory(t)

	resp, err := ts.querier.VerificationHistory(ts.ctx, &types.QueryVerificationHistoryRequest{
		AccountAddress: ts.address.String(),
		Offset:         1,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 3, resp.TotalCount)
	require.Len(t, resp.Entries, 2)
}

func TestGRPCQuerier_VerificationHistory_WithLimitAndOffset(t *testing.T) {
	ts := setupGRPCQueryTest(t)
	ts.createTestWalletWithHistory(t)

	resp, err := ts.querier.VerificationHistory(ts.ctx, &types.QueryVerificationHistoryRequest{
		AccountAddress: ts.address.String(),
		Limit:          1,
		Offset:         1,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 3, resp.TotalCount)
	require.Len(t, resp.Entries, 1)
}

// ============================================================================
// buildPublicVerificationEntry Tests
// ============================================================================

func TestBuildPublicVerificationEntry(t *testing.T) {
	timestamp := time.Now().UTC()

	entry := types.VerificationHistoryEntry{
		EntryID:          "test-entry-123",
		Timestamp:        timestamp,
		BlockHeight:      500,
		PreviousScore:    60,
		NewScore:         85,
		PreviousStatus:   types.AccountStatusPending,
		NewStatus:        types.AccountStatusVerified,
		ScopesEvaluated:  []string{"scope-1", "scope-2", "scope-3"},
		ModelVersion:     "v2.0.0",
		ValidatorAddress: "validator123",
		Reason:           "verification complete",
	}

	publicEntry := buildPublicVerificationEntry(entry)

	require.Equal(t, entry.EntryID, publicEntry.EntryID)
	require.Equal(t, entry.Timestamp.Unix(), publicEntry.Timestamp)
	require.Equal(t, entry.BlockHeight, publicEntry.BlockHeight)
	require.Equal(t, entry.PreviousScore, publicEntry.PreviousScore)
	require.Equal(t, entry.NewScore, publicEntry.NewScore)
	require.Equal(t, string(entry.PreviousStatus), publicEntry.PreviousStatus)
	require.Equal(t, string(entry.NewStatus), publicEntry.NewStatus)
	require.Equal(t, len(entry.ScopesEvaluated), publicEntry.ScopeCount)
	require.Equal(t, entry.ModelVersion, publicEntry.ModelVersion)
}

// ============================================================================
// buildPublicConsentInfo Tests
// ============================================================================

func TestBuildPublicConsentInfo(t *testing.T) {
	now := time.Now().UTC()
	expiresAt := now.Add(24 * time.Hour)

	consent := types.ScopeConsent{
		ScopeID:   "test-scope",
		Granted:   true,
		Purpose:   "verification",
		ExpiresAt: &expiresAt,
	}
	grantedAt := now.Add(-time.Hour)
	consent.GrantedAt = &grantedAt

	publicInfo := buildPublicConsentInfo(consent, now)

	require.Equal(t, consent.ScopeID, publicInfo.ScopeID)
	require.Equal(t, consent.Granted, publicInfo.Granted)
	require.True(t, publicInfo.IsActive)
	require.Equal(t, consent.Purpose, publicInfo.Purpose)
	require.NotNil(t, publicInfo.ExpiresAt)
	require.Equal(t, expiresAt.Unix(), *publicInfo.ExpiresAt)
}

func TestBuildPublicConsentInfo_Expired(t *testing.T) {
	now := time.Now().UTC()
	expiresAt := now.Add(-1 * time.Hour) // Expired

	consent := types.ScopeConsent{
		ScopeID:   "test-scope",
		Granted:   true,
		Purpose:   "verification",
		ExpiresAt: &expiresAt,
	}
	grantedAt := now.Add(-2 * time.Hour)
	consent.GrantedAt = &grantedAt

	publicInfo := buildPublicConsentInfo(consent, now)

	require.Equal(t, consent.ScopeID, publicInfo.ScopeID)
	require.True(t, consent.Granted)
	require.False(t, publicInfo.IsActive) // Expired, so not active
}

func TestBuildPublicConsentInfo_NoExpiry(t *testing.T) {
	now := time.Now().UTC()

	consent := types.ScopeConsent{
		ScopeID:   "test-scope",
		Granted:   true,
		Purpose:   "verification",
		ExpiresAt: nil,
	}
	grantedAt := now.Add(-time.Hour)
	consent.GrantedAt = &grantedAt

	publicInfo := buildPublicConsentInfo(consent, now)

	require.Equal(t, consent.ScopeID, publicInfo.ScopeID)
	require.True(t, publicInfo.Granted)
	require.True(t, publicInfo.IsActive)
	require.Nil(t, publicInfo.ExpiresAt)
}
