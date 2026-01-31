package keeper_test

import (
	"encoding/json"
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

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	"github.com/virtengine/virtengine/x/veid/keeper"
	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Tier Transitions Tests
// ============================================================================

func TestUpdateAccountTier_NotFound(t *testing.T) {
	ctx, k := setupTierTestKeeper(t)
	addr := sdk.AccAddress([]byte("nonexistent"))

	result, err := k.UpdateAccountTier(ctx, addr)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "identity record not found")
}

func TestUpdateAccountTier_Tier0_NoScopes(t *testing.T) {
	ctx, k := setupTierTestKeeper(t)
	addr := sdk.AccAddress([]byte("testaddr0000000001"))

	// Create identity record with no scopes
	record := types.NewIdentityRecord(addr.String(), ctx.BlockTime())
	err := k.SetIdentityRecord(ctx, *record)
	require.NoError(t, err)

	result, err := k.UpdateAccountTier(ctx, addr)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, types.TierUnverified, result.NewTier)
	require.Equal(t, uint32(0), result.CompositeScore)
	require.Equal(t, 0, result.VerifiedScopeCount)
}

func TestUpdateAccountTier_Tier1_BasicAccess(t *testing.T) {
	ctx, k := setupTierTestKeeper(t)
	addr := sdk.AccAddress([]byte("testaddr0000000002"))

	// Create identity record
	record := types.NewIdentityRecord(addr.String(), ctx.BlockTime())
	err := k.SetIdentityRecord(ctx, *record)
	require.NoError(t, err)

	// Add scopes that total 50-69 points (Tier 1 Basic)
	// ID Document = 30, Selfie = 20 = 50 points
	addVerifiedScope(t, ctx, k, addr, "scope1", types.ScopeTypeIDDocument)
	addVerifiedScope(t, ctx, k, addr, "scope2", types.ScopeTypeSelfie)

	result, err := k.UpdateAccountTier(ctx, addr)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, types.TierBasic, result.NewTier)
	require.Equal(t, uint32(50), result.CompositeScore)
	require.Equal(t, 2, result.VerifiedScopeCount)
}

func TestUpdateAccountTier_Tier2_StandardAccess(t *testing.T) {
	ctx, k := setupTierTestKeeper(t)
	addr := sdk.AccAddress([]byte("testaddr0000000003"))

	// Create identity record
	record := types.NewIdentityRecord(addr.String(), ctx.BlockTime())
	err := k.SetIdentityRecord(ctx, *record)
	require.NoError(t, err)

	// Add scopes that total 70-84 points (Tier 2 Standard)
	// ID Document = 30, Selfie = 20, FaceVideo = 25 = 75 points
	addVerifiedScope(t, ctx, k, addr, "scope1", types.ScopeTypeIDDocument)
	addVerifiedScope(t, ctx, k, addr, "scope2", types.ScopeTypeSelfie)
	addVerifiedScope(t, ctx, k, addr, "scope3", types.ScopeTypeFaceVideo)

	result, err := k.UpdateAccountTier(ctx, addr)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, types.TierStandard, result.NewTier)
	require.Equal(t, uint32(75), result.CompositeScore)
	require.Equal(t, 3, result.VerifiedScopeCount)
}

func TestUpdateAccountTier_Tier3_PremiumAccess(t *testing.T) {
	ctx, k := setupTierTestKeeper(t)
	addr := sdk.AccAddress([]byte("testaddr0000000004"))

	// Create identity record
	record := types.NewIdentityRecord(addr.String(), ctx.BlockTime())
	err := k.SetIdentityRecord(ctx, *record)
	require.NoError(t, err)

	// Add scopes that total 85+ points (Tier 3 Premium)
	// ID Document = 30, Selfie = 20, FaceVideo = 25, DomainVerify = 15 = 90 points
	addVerifiedScope(t, ctx, k, addr, "scope1", types.ScopeTypeIDDocument)
	addVerifiedScope(t, ctx, k, addr, "scope2", types.ScopeTypeSelfie)
	addVerifiedScope(t, ctx, k, addr, "scope3", types.ScopeTypeFaceVideo)
	addVerifiedScope(t, ctx, k, addr, "scope4", types.ScopeTypeDomainVerify)

	result, err := k.UpdateAccountTier(ctx, addr)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, types.TierPremium, result.NewTier)
	require.Equal(t, uint32(90), result.CompositeScore)
	require.Equal(t, 4, result.VerifiedScopeCount)
}

func TestUpdateAccountTier_ScoreCappedAt100(t *testing.T) {
	ctx, k := setupTierTestKeeper(t)
	addr := sdk.AccAddress([]byte("testaddr0000000005"))

	// Create identity record
	record := types.NewIdentityRecord(addr.String(), ctx.BlockTime())
	err := k.SetIdentityRecord(ctx, *record)
	require.NoError(t, err)

	// Add all possible scopes (way more than 100 points)
	addVerifiedScope(t, ctx, k, addr, "scope1", types.ScopeTypeIDDocument)   // 30
	addVerifiedScope(t, ctx, k, addr, "scope2", types.ScopeTypeSelfie)       // 20
	addVerifiedScope(t, ctx, k, addr, "scope3", types.ScopeTypeFaceVideo)    // 25
	addVerifiedScope(t, ctx, k, addr, "scope4", types.ScopeTypeBiometric)    // 20
	addVerifiedScope(t, ctx, k, addr, "scope5", types.ScopeTypeDomainVerify) // 15
	addVerifiedScope(t, ctx, k, addr, "scope6", types.ScopeTypeEmailProof)   // 10
	addVerifiedScope(t, ctx, k, addr, "scope7", types.ScopeTypeSMSProof)     // 10
	addVerifiedScope(t, ctx, k, addr, "scope8", types.ScopeTypeSSOMetadata)  // 5
	addVerifiedScope(t, ctx, k, addr, "scope9", types.ScopeTypeADSSO)        // 12
	// Total: 147 points, but should cap at 100

	result, err := k.UpdateAccountTier(ctx, addr)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, types.TierPremium, result.NewTier)
	require.Equal(t, uint32(100), result.CompositeScore) // Capped at 100
}

func TestUpdateAccountTier_LockedAccount(t *testing.T) {
	ctx, k := setupTierTestKeeper(t)
	addr := sdk.AccAddress([]byte("testaddr0000000006"))

	// Create locked identity record
	record := types.NewIdentityRecord(addr.String(), ctx.BlockTime())
	record.Locked = true
	record.LockedReason = "fraud detected"
	err := k.SetIdentityRecord(ctx, *record)
	require.NoError(t, err)

	// Add high-value scopes
	addVerifiedScope(t, ctx, k, addr, "scope1", types.ScopeTypeIDDocument)
	addVerifiedScope(t, ctx, k, addr, "scope2", types.ScopeTypeSelfie)
	addVerifiedScope(t, ctx, k, addr, "scope3", types.ScopeTypeFaceVideo)

	result, err := k.UpdateAccountTier(ctx, addr)
	require.NoError(t, err)
	require.NotNil(t, result)
	// Even with high score, locked accounts are Tier 0
	require.Equal(t, types.TierUnverified, result.NewTier)
}

func TestUpdateAccountTier_TierChange_EmitsEvent(t *testing.T) {
	ctx, k := setupTierTestKeeper(t)
	addr := sdk.AccAddress([]byte("testaddr0000000007"))

	// Create identity record at Tier 0
	record := types.NewIdentityRecord(addr.String(), ctx.BlockTime())
	record.Tier = types.IdentityTierUnverified
	err := k.SetIdentityRecord(ctx, *record)
	require.NoError(t, err)

	// Add scopes to bring to Tier 1
	addVerifiedScope(t, ctx, k, addr, "scope1", types.ScopeTypeIDDocument)
	addVerifiedScope(t, ctx, k, addr, "scope2", types.ScopeTypeSelfie)

	result, err := k.UpdateAccountTier(ctx, addr)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Changed)
	require.Equal(t, types.TierUnverified, result.OldTier)
	require.Equal(t, types.TierBasic, result.NewTier)

	// Check event was emitted
	events := ctx.EventManager().Events()
	found := false
	for _, event := range events {
		if event.Type == "virtengine.veid.v1.EventTierChanged" {
			found = true
			break
		}
	}
	require.True(t, found, "EventTierChanged should be emitted")
}

func TestUpdateAccountTier_NoChange_NoEvent(t *testing.T) {
	ctx, k := setupTierTestKeeper(t)
	addr := sdk.AccAddress([]byte("testaddr0000000008"))

	// Create identity record already at Tier 1
	record := types.NewIdentityRecord(addr.String(), ctx.BlockTime())
	record.Tier = types.IdentityTierBasic
	record.CurrentScore = 50
	err := k.SetIdentityRecord(ctx, *record)
	require.NoError(t, err)

	// Add scopes for exactly Tier 1
	addVerifiedScope(t, ctx, k, addr, "scope1", types.ScopeTypeIDDocument) // 30
	addVerifiedScope(t, ctx, k, addr, "scope2", types.ScopeTypeSelfie)     // 20 = 50

	// Clear events
	ctx = ctx.WithEventManager(sdk.NewEventManager())

	result, err := k.UpdateAccountTier(ctx, addr)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.Changed)
	require.Equal(t, types.TierBasic, result.OldTier)
	require.Equal(t, types.TierBasic, result.NewTier)

	// Check no tier change event was emitted
	events := ctx.EventManager().Events()
	for _, event := range events {
		require.NotEqual(t, "virtengine.veid.v1.EventTierChanged", event.Type)
	}
}

// ============================================================================
// MeetsScoreThreshold Tests
// ============================================================================

func TestMeetsScoreThreshold_Success(t *testing.T) {
	ctx, k := setupTierTestKeeper(t)
	addr := sdk.AccAddress([]byte("testaddr0000000009"))

	// Create verified identity record with score
	record := types.NewIdentityRecord(addr.String(), ctx.BlockTime())
	err := k.SetIdentityRecord(ctx, *record)
	require.NoError(t, err)

	// Set score with verified status
	err = k.SetScore(ctx, addr.String(), 75, "v1.0")
	require.NoError(t, err)

	require.True(t, k.MeetsScoreThreshold(ctx, addr, 50))
	require.True(t, k.MeetsScoreThreshold(ctx, addr, 70))
	require.True(t, k.MeetsScoreThreshold(ctx, addr, 75))
	require.False(t, k.MeetsScoreThreshold(ctx, addr, 76))
	require.False(t, k.MeetsScoreThreshold(ctx, addr, 85))
}

func TestMeetsScoreThreshold_NotFound(t *testing.T) {
	ctx, k := setupTierTestKeeper(t)
	addr := sdk.AccAddress([]byte("nonexistent"))

	require.False(t, k.MeetsScoreThreshold(ctx, addr, 50))
}

func TestMeetsScoreThreshold_LockedAccount(t *testing.T) {
	ctx, k := setupTierTestKeeper(t)
	addr := sdk.AccAddress([]byte("testaddr0000000010"))

	// Create locked identity record with high score
	record := types.NewIdentityRecord(addr.String(), ctx.BlockTime())
	record.Locked = true
	record.CurrentScore = 100
	err := k.SetIdentityRecord(ctx, *record)
	require.NoError(t, err)

	// Even with score 100, locked accounts don't meet thresholds
	require.False(t, k.MeetsScoreThreshold(ctx, addr, 0))
}

// ============================================================================
// MeetsTierRequirement Tests
// ============================================================================

func TestMeetsTierRequirement(t *testing.T) {
	ctx, k := setupTierTestKeeper(t)
	addr := sdk.AccAddress([]byte("testaddr0000000011"))

	// Create verified identity record
	record := types.NewIdentityRecord(addr.String(), ctx.BlockTime())
	err := k.SetIdentityRecord(ctx, *record)
	require.NoError(t, err)

	// Set score for Tier 2 (70-84)
	err = k.SetScore(ctx, addr.String(), 75, "v1.0")
	require.NoError(t, err)

	require.True(t, k.MeetsTierRequirement(ctx, addr, types.TierUnverified))
	require.True(t, k.MeetsTierRequirement(ctx, addr, types.TierBasic))
	require.True(t, k.MeetsTierRequirement(ctx, addr, types.TierStandard))
	require.False(t, k.MeetsTierRequirement(ctx, addr, types.TierPremium))
}

// ============================================================================
// GetAccountTierDetails Tests
// ============================================================================

func TestGetAccountTierDetails_Success(t *testing.T) {
	ctx, k := setupTierTestKeeper(t)
	addr := sdk.AccAddress([]byte("testaddr0000000012"))

	// Create identity record
	record := types.NewIdentityRecord(addr.String(), ctx.BlockTime())
	err := k.SetIdentityRecord(ctx, *record)
	require.NoError(t, err)

	// Set score for Tier 2
	err = k.SetScore(ctx, addr.String(), 75, "v1.0")
	require.NoError(t, err)

	details, err := k.GetAccountTierDetails(ctx, addr)
	require.NoError(t, err)
	require.NotNil(t, details)
	require.Equal(t, addr.String(), details.Address)
	require.Equal(t, types.TierStandard, details.Tier)
	require.Equal(t, "standard", details.TierName)
	require.Equal(t, uint32(75), details.Score)
	require.Equal(t, types.ThresholdPremium, details.NextTierThreshold)
	require.Equal(t, uint32(10), details.PointsToNextTier) // 85 - 75 = 10
	require.False(t, details.Locked)
}

func TestGetAccountTierDetails_AtMaxTier(t *testing.T) {
	ctx, k := setupTierTestKeeper(t)
	addr := sdk.AccAddress([]byte("testaddr0000000013"))

	// Create identity record
	record := types.NewIdentityRecord(addr.String(), ctx.BlockTime())
	err := k.SetIdentityRecord(ctx, *record)
	require.NoError(t, err)

	// Set score for Tier 3 (Premium)
	err = k.SetScore(ctx, addr.String(), 95, "v1.0")
	require.NoError(t, err)

	details, err := k.GetAccountTierDetails(ctx, addr)
	require.NoError(t, err)
	require.NotNil(t, details)
	require.Equal(t, types.TierPremium, details.Tier)
	require.Equal(t, "premium", details.TierName)
	require.Equal(t, uint32(0), details.NextTierThreshold) // Already at max
	require.Equal(t, uint32(0), details.PointsToNextTier)
}

func TestGetAccountTierDetails_NotFound(t *testing.T) {
	ctx, k := setupTierTestKeeper(t)
	addr := sdk.AccAddress([]byte("nonexistent"))

	details, err := k.GetAccountTierDetails(ctx, addr)
	require.Error(t, err)
	require.Nil(t, details)
	require.Contains(t, err.Error(), "identity record not found")
}

// ============================================================================
// RecalculateAllAccountTiers Tests
// ============================================================================

func TestRecalculateAllAccountTiers(t *testing.T) {
	ctx, k := setupTierTestKeeper(t)

	// Create multiple accounts with different scores
	for i := 0; i < 5; i++ {
		addr := sdk.AccAddress([]byte{byte(i + 100), byte(i + 100), byte(i + 100), 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, byte(i)})
		record := types.NewIdentityRecord(addr.String(), ctx.BlockTime())
		err := k.SetIdentityRecord(ctx, *record)
		require.NoError(t, err)

		// Add varying scopes
		if i >= 1 {
			addVerifiedScope(t, ctx, k, addr, "scope1", types.ScopeTypeIDDocument)
		}
		if i >= 2 {
			addVerifiedScope(t, ctx, k, addr, "scope2", types.ScopeTypeSelfie)
		}
		if i >= 3 {
			addVerifiedScope(t, ctx, k, addr, "scope3", types.ScopeTypeFaceVideo)
		}
	}

	processed, changed := k.RecalculateAllAccountTiers(ctx)
	require.Equal(t, 5, processed)
	require.GreaterOrEqual(t, changed, 0)
}

// ============================================================================
// Tier Threshold Edge Cases
// ============================================================================

func TestTierThresholds_EdgeCases(t *testing.T) {
	testCases := []struct {
		name         string
		score        uint32
		expectedTier int
	}{
		{"Score 0", 0, types.TierUnverified},
		{"Score 49 (just below Basic)", 49, types.TierUnverified},
		{"Score 50 (exactly Basic)", 50, types.TierBasic},
		{"Score 51", 51, types.TierBasic},
		{"Score 69 (top of Basic)", 69, types.TierBasic},
		{"Score 70 (exactly Standard)", 70, types.TierStandard},
		{"Score 84 (top of Standard)", 84, types.TierStandard},
		{"Score 85 (exactly Premium)", 85, types.TierPremium},
		{"Score 100 (max)", 100, types.TierPremium},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, k := setupTierTestKeeper(t)
			addr := sdk.AccAddress([]byte("testaddr" + tc.name))

			record := types.NewIdentityRecord(addr.String(), ctx.BlockTime())
			err := k.SetIdentityRecord(ctx, *record)
			require.NoError(t, err)

			// Set score directly
			err = k.SetScore(ctx, addr.String(), tc.score, "v1.0")
			require.NoError(t, err)

			tier, err := k.GetAccountTier(ctx, addr.String())
			require.NoError(t, err)
			require.Equal(t, tc.expectedTier, tier, "Score %d should map to tier %d", tc.score, tc.expectedTier)
		})
	}
}

// ============================================================================
// Test Helpers
// ============================================================================

func setupTierTestKeeper(t *testing.T) (sdk.Context, keeper.Keeper) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	ctx := createTierTestContext(t, storeKey)
	ctx = ctx.WithBlockTime(time.Now().UTC())

	k := keeper.NewKeeper(cdc, storeKey, "gov")

	// Initialize default params
	err := k.SetParams(ctx, types.DefaultParams())
	require.NoError(t, err)

	return ctx, k
}

func createTierTestContext(t *testing.T, storeKey *storetypes.KVStoreKey) sdk.Context {
	t.Helper()

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	require.NoError(t, err)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())
	return ctx
}

// scopeStoreTest mirrors the internal scopeStore for test marshaling
type scopeStoreTest struct {
	ScopeID          string                   `json:"scope_id"`
	ScopeType        types.ScopeType          `json:"scope_type"`
	Version          uint32                   `json:"version"`
	EncryptedPayload json.RawMessage          `json:"encrypted_payload"`
	UploadMetadata   types.UploadMetadata     `json:"upload_metadata"`
	Status           types.VerificationStatus `json:"status"`
	UploadedAt       int64                    `json:"uploaded_at"`
	VerifiedAt       *int64                   `json:"verified_at,omitempty"`
	ExpiresAt        *int64                   `json:"expires_at,omitempty"`
	Revoked          bool                     `json:"revoked"`
	RevokedAt        *int64                   `json:"revoked_at,omitempty"`
	RevokedReason    string                   `json:"revoked_reason,omitempty"`
}

func addVerifiedScope(t *testing.T, ctx sdk.Context, k keeper.Keeper, addr sdk.AccAddress, scopeID string, scopeType types.ScopeType) {
	t.Helper()

	// Get or create identity record
	record, found := k.GetIdentityRecord(ctx, addr)
	if !found {
		newRecord := types.NewIdentityRecord(addr.String(), ctx.BlockTime())
		err := k.SetIdentityRecord(ctx, *newRecord)
		require.NoError(t, err)
		record = *newRecord
	}

	now := ctx.BlockTime()

	scope := &types.IdentityScope{
		ScopeID:   scopeID,
		ScopeType: scopeType,
		Version:   types.ScopeSchemaVersion,
		EncryptedPayload: encryptiontypes.EncryptedPayloadEnvelope{
			Ciphertext: []byte("testciphertext"),
			Nonce:      []byte("testnonce"),
		},
		UploadMetadata: types.UploadMetadata{
			Salt:              []byte("testsalt12345678"),
			ClientSignature:   []byte("sig"),
			UserSignature:     []byte("usersig"),
			DeviceFingerprint: "test-device",
			ClientID:          "test-client",
			CaptureTimestamp:  now.Unix(),
		},
		Status:     types.VerificationStatusVerified,
		UploadedAt: now,
		VerifiedAt: &now,
	}

	// Marshal the scope for storage
	payloadBz, err := json.Marshal(scope.EncryptedPayload)
	require.NoError(t, err)

	ts := scope.VerifiedAt.Unix()
	ss := scopeStoreTest{
		ScopeID:          scope.ScopeID,
		ScopeType:        scope.ScopeType,
		Version:          scope.Version,
		EncryptedPayload: payloadBz,
		UploadMetadata:   scope.UploadMetadata,
		Status:           scope.Status,
		UploadedAt:       scope.UploadedAt.Unix(),
		VerifiedAt:       &ts,
		Revoked:          scope.Revoked,
	}

	scopeBytes, err := json.Marshal(&ss)
	require.NoError(t, err)

	// Store scope directly (bypass salt validation for testing)
	store := ctx.KVStore(k.StoreKey())
	store.Set(types.ScopeKey(addr.Bytes(), scopeID), scopeBytes)

	// Update identity record with scope reference
	record.AddScopeRef(types.NewScopeRef(scope))
	record.UpdatedAt = now
	err = k.SetIdentityRecord(ctx, record)
	require.NoError(t, err)
}
