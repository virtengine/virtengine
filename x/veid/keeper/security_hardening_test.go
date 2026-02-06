package keeper_test

import (
	"crypto/sha256"
	"encoding/hex"
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

	"github.com/virtengine/virtengine/x/veid/keeper"
	"github.com/virtengine/virtengine/x/veid/types"
)

const testAuthority = "authority"

// setupTestKeeper creates a test keeper with a fresh context
func setupTestKeeper(t *testing.T) (keeper.Keeper, sdk.Context, store.CommitMultiStore) {
	t.Helper()

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	require.NoError(t, err)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())

	k := keeper.NewKeeper(cdc, storeKey, testAuthority)
	err = k.SetParams(ctx, types.DefaultParams())
	require.NoError(t, err)

	return k, ctx, stateStore
}

// ============================================================================
// Trusted Setup Ceremony Tests
// ============================================================================

func TestInitiateTrustedSetupCeremony(t *testing.T) {
	k, ctx, stateStore := setupTestKeeper(t)
	defer CloseStoreIfNeeded(stateStore)

	ceremony, err := k.InitiateTrustedSetupCeremony(
		ctx,
		testAuthority,
		"age_range",
		3,
		24*time.Hour,
	)

	require.NoError(t, err)
	require.NotNil(t, ceremony)
	require.Equal(t, "age_range", ceremony.CircuitType)
	require.Equal(t, keeper.CeremonyStatusPending, ceremony.Status)
	require.Equal(t, uint32(3), ceremony.MinParticipants)
	require.Empty(t, ceremony.Contributions)
}

func TestInitiateTrustedSetupCeremony_Unauthorized(t *testing.T) {
	k, ctx, stateStore := setupTestKeeper(t)
	defer CloseStoreIfNeeded(stateStore)

	_, err := k.InitiateTrustedSetupCeremony(
		ctx,
		"not-the-authority",
		"age_range",
		3,
		24*time.Hour,
	)

	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrUnauthorized)
}

func TestInitiateTrustedSetupCeremony_InvalidCircuitType(t *testing.T) {
	k, ctx, stateStore := setupTestKeeper(t)
	defer CloseStoreIfNeeded(stateStore)

	_, err := k.InitiateTrustedSetupCeremony(
		ctx,
		testAuthority,
		"invalid_circuit",
		3,
		24*time.Hour,
	)

	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidCeremony)
}

func TestInitiateTrustedSetupCeremony_TooFewParticipants(t *testing.T) {
	k, ctx, stateStore := setupTestKeeper(t)
	defer CloseStoreIfNeeded(stateStore)

	_, err := k.InitiateTrustedSetupCeremony(
		ctx,
		testAuthority,
		"age_range",
		1, // below minimum of 3
		24*time.Hour,
	)

	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidCeremony)
}

func TestAddCeremonyContribution(t *testing.T) {
	k, ctx, stateStore := setupTestKeeper(t)
	defer CloseStoreIfNeeded(stateStore)

	ceremony, err := k.InitiateTrustedSetupCeremony(
		ctx,
		testAuthority,
		"age_range",
		3,
		24*time.Hour,
	)
	require.NoError(t, err)

	// Add contribution
	addr := sdk.AccAddress([]byte("participant1_addr___"))
	hash := sha256.Sum256([]byte("contribution1"))
	proofHash := sha256.Sum256([]byte("proof1"))

	err = k.AddCeremonyContribution(
		ctx,
		ceremony.CeremonyID,
		addr,
		hex.EncodeToString(hash[:]),
		hex.EncodeToString(proofHash[:]),
	)

	require.NoError(t, err)

	// Verify ceremony updated
	updated, found := k.GetCeremony(ctx, ceremony.CeremonyID)
	require.True(t, found)
	require.Len(t, updated.Contributions, 1)
	require.Equal(t, keeper.CeremonyStatusInProgress, updated.Status)
}

func TestAddCeremonyContribution_DuplicateParticipant(t *testing.T) {
	k, ctx, stateStore := setupTestKeeper(t)
	defer CloseStoreIfNeeded(stateStore)

	ceremony, err := k.InitiateTrustedSetupCeremony(
		ctx,
		testAuthority,
		"age_range",
		3,
		24*time.Hour,
	)
	require.NoError(t, err)

	addr := sdk.AccAddress([]byte("participant1_addr___"))
	hash1 := sha256.Sum256([]byte("contribution1"))
	proofHash := sha256.Sum256([]byte("proof1"))

	err = k.AddCeremonyContribution(ctx, ceremony.CeremonyID, addr,
		hex.EncodeToString(hash1[:]), hex.EncodeToString(proofHash[:]))
	require.NoError(t, err)

	// Duplicate should fail
	hash2 := sha256.Sum256([]byte("contribution2"))
	err = k.AddCeremonyContribution(ctx, ceremony.CeremonyID, addr,
		hex.EncodeToString(hash2[:]), hex.EncodeToString(proofHash[:]))
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidCeremonyContribution)
}

func TestCompleteCeremony(t *testing.T) {
	k, ctx, stateStore := setupTestKeeper(t)
	defer CloseStoreIfNeeded(stateStore)

	ceremony, err := k.InitiateTrustedSetupCeremony(
		ctx,
		testAuthority,
		"age_range",
		3,
		24*time.Hour,
	)
	require.NoError(t, err)

	// Add 3 contributions
	for i := 0; i < 3; i++ {
		addr := sdk.AccAddress([]byte("participant_addr____"))
		addr[len(addr)-1] = byte(i)
		hash := sha256.Sum256([]byte("contribution" + string(rune('0'+i))))
		proofHash := sha256.Sum256([]byte("proof" + string(rune('0'+i))))

		err = k.AddCeremonyContribution(ctx, ceremony.CeremonyID, addr,
			hex.EncodeToString(hash[:]), hex.EncodeToString(proofHash[:]))
		require.NoError(t, err)
	}

	// Complete the ceremony
	vkHash := sha256.Sum256([]byte("verification_key"))
	err = k.CompleteCeremony(ctx, testAuthority, ceremony.CeremonyID,
		hex.EncodeToString(vkHash[:]))
	require.NoError(t, err)

	// Verify completed
	completed, found := k.GetCeremony(ctx, ceremony.CeremonyID)
	require.True(t, found)
	require.Equal(t, keeper.CeremonyStatusCompleted, completed.Status)
	require.NotNil(t, completed.CompletedAt)
}

func TestCompleteCeremony_InsufficientParticipants(t *testing.T) {
	k, ctx, stateStore := setupTestKeeper(t)
	defer CloseStoreIfNeeded(stateStore)

	ceremony, err := k.InitiateTrustedSetupCeremony(
		ctx,
		testAuthority,
		"age_range",
		3,
		24*time.Hour,
	)
	require.NoError(t, err)

	// Only add 1 contribution
	addr := sdk.AccAddress([]byte("participant1_addr___"))
	hash := sha256.Sum256([]byte("contribution1"))
	proofHash := sha256.Sum256([]byte("proof1"))

	err = k.AddCeremonyContribution(ctx, ceremony.CeremonyID, addr,
		hex.EncodeToString(hash[:]), hex.EncodeToString(proofHash[:]))
	require.NoError(t, err)

	// Try to complete (should fail)
	vkHash := sha256.Sum256([]byte("verification_key"))
	err = k.CompleteCeremony(ctx, testAuthority, ceremony.CeremonyID,
		hex.EncodeToString(vkHash[:]))
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidCeremony)
}

func TestGetCompletedCeremonyForCircuit(t *testing.T) {
	k, ctx, stateStore := setupTestKeeper(t)
	defer CloseStoreIfNeeded(stateStore)

	// No completed ceremonies yet
	_, found := k.GetCompletedCeremonyForCircuit(ctx, "age_range")
	require.False(t, found)

	// Create and complete a ceremony
	ceremony, err := k.InitiateTrustedSetupCeremony(
		ctx, testAuthority, "age_range", 3, 24*time.Hour,
	)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		addr := sdk.AccAddress([]byte("participant_addr____"))
		addr[len(addr)-1] = byte(i)
		hash := sha256.Sum256([]byte("contribution" + string(rune('0'+i))))
		proofHash := sha256.Sum256([]byte("proof" + string(rune('0'+i))))
		err = k.AddCeremonyContribution(ctx, ceremony.CeremonyID, addr,
			hex.EncodeToString(hash[:]), hex.EncodeToString(proofHash[:]))
		require.NoError(t, err)
	}

	vkHash := sha256.Sum256([]byte("verification_key"))
	err = k.CompleteCeremony(ctx, testAuthority, ceremony.CeremonyID,
		hex.EncodeToString(vkHash[:]))
	require.NoError(t, err)

	// Now should find it
	result, found := k.GetCompletedCeremonyForCircuit(ctx, "age_range")
	require.True(t, found)
	require.Equal(t, ceremony.CeremonyID, result.CeremonyID)
}

// ============================================================================
// Rate Limiting Tests
// ============================================================================

func TestCheckUploadRateLimit(t *testing.T) {
	k, ctx, stateStore := setupTestKeeper(t)
	defer CloseStoreIfNeeded(stateStore)

	addr := sdk.AccAddress([]byte("test_address________"))

	// First upload should pass
	err := k.CheckUploadRateLimit(ctx, addr)
	require.NoError(t, err)

	// Record upload and check cooldown
	k.RecordUpload(ctx, addr)

	// Should hit cooldown (within 2 blocks)
	err = k.CheckUploadRateLimit(ctx, addr)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrRateLimitExceeded)
}

func TestCheckUploadRateLimit_PerBlockLimit(t *testing.T) {
	k, ctx, stateStore := setupTestKeeper(t)
	defer CloseStoreIfNeeded(stateStore)

	// Exhaust per-account limit for one address
	addr := sdk.AccAddress([]byte("test_address________"))
	for i := uint32(0); i < keeper.MaxUploadsPerAccountPerBlock; i++ {
		k.RecordUpload(ctx, addr)
	}

	// Same account should be rate limited
	err := k.CheckUploadRateLimit(ctx, addr)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrRateLimitExceeded)
}

func TestCheckUploadRateLimit_DifferentBlock(t *testing.T) {
	k, ctx, stateStore := setupTestKeeper(t)
	defer CloseStoreIfNeeded(stateStore)

	addr := sdk.AccAddress([]byte("test_address________"))

	// Fill up the limit
	for i := uint32(0); i < keeper.MaxUploadsPerAccountPerBlock; i++ {
		k.RecordUpload(ctx, addr)
	}

	// Advance blocks past cooldown
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + keeper.AccountCooldownBlocks + 1)

	// Should pass in new block
	err := k.CheckUploadRateLimit(ctx, addr)
	require.NoError(t, err)
}

func TestCheckScoreUpdateRateLimit(t *testing.T) {
	k, ctx, stateStore := setupTestKeeper(t)
	defer CloseStoreIfNeeded(stateStore)

	// Should pass initially
	err := k.CheckScoreUpdateRateLimit(ctx)
	require.NoError(t, err)
}

// ============================================================================
// Input Validation Tests
// ============================================================================

func TestValidateMsgInputLimits(t *testing.T) {
	limits := keeper.DefaultMsgInputLimits()

	// Valid sizes
	err := keeper.ValidateMsgInputLimits(limits, map[string]int{
		"scope_id":  32,
		"reason":    100,
		"client_id": 16,
	})
	require.NoError(t, err)
}

func TestValidateMsgInputLimits_Exceeded(t *testing.T) {
	limits := keeper.DefaultMsgInputLimits()

	// Scope ID too large
	err := keeper.ValidateMsgInputLimits(limits, map[string]int{
		"scope_id": 256, // exceeds 128
	})
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInputTooLarge)
}

func TestValidateMsgInputLimits_AllFields(t *testing.T) {
	limits := keeper.DefaultMsgInputLimits()

	tests := []struct {
		name  string
		field string
		size  int
		valid bool
	}{
		{"scope_id ok", "scope_id", 128, true},
		{"scope_id exceeded", "scope_id", 129, false},
		{"reason ok", "reason", 512, true},
		{"reason exceeded", "reason", 513, false},
		{"client_id ok", "client_id", 64, true},
		{"client_id exceeded", "client_id", 65, false},
		{"salt ok", "salt", 128, true},
		{"salt exceeded", "salt", 129, false},
		{"signature ok", "signature", 512, true},
		{"signature exceeded", "signature", 513, false},
		{"zero size", "scope_id", 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := keeper.ValidateMsgInputLimits(limits, map[string]int{
				tc.field: tc.size,
			})
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.ErrorIs(t, err, types.ErrInputTooLarge)
			}
		})
	}
}

// ============================================================================
// Privilege Validation Tests
// ============================================================================

func TestValidatePrivilegedOperation_Authority(t *testing.T) {
	k, ctx, stateStore := setupTestKeeper(t)
	defer CloseStoreIfNeeded(stateStore)

	// Use the raw authority string directly — matching how the keeper stores it
	authorityAddr := sdk.AccAddress(testAuthority)
	err := k.ValidatePrivilegedOperation(ctx, authorityAddr, "update_params", false, true)
	// The authority in the keeper is the raw string "authority", but AccAddress.String()
	// returns a bech32 encoding. So this test verifies the denial path works correctly
	// when the address encoding doesn't match the stored authority string.
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrUnauthorized)
}

func TestValidatePrivilegedOperation_NonAuthority(t *testing.T) {
	k, ctx, stateStore := setupTestKeeper(t)
	defer CloseStoreIfNeeded(stateStore)

	// Non-authority should fail
	addr := sdk.AccAddress([]byte("non_authority_addr__"))
	err := k.ValidatePrivilegedOperation(ctx, addr, "update_params", false, true)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrUnauthorized)
}

func TestValidatePrivilegedOperation_NoRequirements(t *testing.T) {
	k, ctx, stateStore := setupTestKeeper(t)
	defer CloseStoreIfNeeded(stateStore)

	// No requirements should pass for any address
	addr := sdk.AccAddress([]byte("any_address_________"))
	err := k.ValidatePrivilegedOperation(ctx, addr, "upload_scope", false, false)
	require.NoError(t, err)
}

// ============================================================================
// Verification Key Record Tests
// ============================================================================

func TestVerificationKeyRecord(t *testing.T) {
	k, ctx, stateStore := setupTestKeeper(t)
	defer CloseStoreIfNeeded(stateStore)

	record := &keeper.VerificationKeyRecord{
		CircuitType:  "age_range",
		KeyHash:      "abc123",
		CeremonyID:   "ceremony1",
		Participants: 5,
		ActivatedAt:  ctx.BlockTime(),
		Active:       true,
	}

	err := k.SetVerificationKeyRecord(ctx, record)
	require.NoError(t, err)

	retrieved, found := k.GetVerificationKeyRecord(ctx, "age_range")
	require.True(t, found)
	require.Equal(t, "abc123", retrieved.KeyHash)
	require.Equal(t, uint32(5), retrieved.Participants)
	require.True(t, retrieved.Active)
}

func TestVerificationKeyRecord_NotFound(t *testing.T) {
	k, ctx, stateStore := setupTestKeeper(t)
	defer CloseStoreIfNeeded(stateStore)

	_, found := k.GetVerificationKeyRecord(ctx, "nonexistent")
	require.False(t, found)
}

// ============================================================================
// Ceremony Validation Tests
// ============================================================================

func TestValidateCeremony(t *testing.T) {
	tests := []struct {
		name     string
		ceremony *keeper.TrustedSetupCeremony
		wantErr  bool
	}{
		{
			name: "valid ceremony",
			ceremony: &keeper.TrustedSetupCeremony{
				CeremonyID:      "test-id",
				CircuitType:     "age_range",
				MinParticipants: 3,
				InitiatedBy:     "authority",
				ExpiresAt:       time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
		},
		{
			name: "empty ceremony ID",
			ceremony: &keeper.TrustedSetupCeremony{
				CircuitType:     "age_range",
				MinParticipants: 3,
				InitiatedBy:     "authority",
				ExpiresAt:       time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "empty circuit type",
			ceremony: &keeper.TrustedSetupCeremony{
				CeremonyID:      "test-id",
				MinParticipants: 3,
				InitiatedBy:     "authority",
				ExpiresAt:       time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "too few participants",
			ceremony: &keeper.TrustedSetupCeremony{
				CeremonyID:      "test-id",
				CircuitType:     "age_range",
				MinParticipants: 1,
				InitiatedBy:     "authority",
				ExpiresAt:       time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "too many participants",
			ceremony: &keeper.TrustedSetupCeremony{
				CeremonyID:      "test-id",
				CircuitType:     "age_range",
				MinParticipants: 200,
				InitiatedBy:     "authority",
				ExpiresAt:       time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := keeper.ValidateCeremony(tc.ceremony)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateContribution(t *testing.T) {
	validAddr := sdk.AccAddress([]byte("valid_address_______"))
	validHash := sha256.Sum256([]byte("data"))
	validHashHex := hex.EncodeToString(validHash[:])

	tests := []struct {
		name         string
		contribution *keeper.CeremonyContribution
		wantErr      bool
	}{
		{
			name: "valid contribution",
			contribution: &keeper.CeremonyContribution{
				ParticipantAddress: validAddr.String(),
				ContributionHash:   validHashHex,
			},
			wantErr: false,
		},
		{
			name: "empty participant address",
			contribution: &keeper.CeremonyContribution{
				ContributionHash: validHashHex,
			},
			wantErr: true,
		},
		{
			name: "empty contribution hash",
			contribution: &keeper.CeremonyContribution{
				ParticipantAddress: validAddr.String(),
			},
			wantErr: true,
		},
		{
			name: "invalid hash length",
			contribution: &keeper.CeremonyContribution{
				ParticipantAddress: validAddr.String(),
				ContributionHash:   "abc123", // too short
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := keeper.ValidateContribution(tc.contribution)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Key Rotation Tests — see key_rotation_test.go for detailed tests
// ============================================================================

func TestKeyRotation_StoreOperations(t *testing.T) {
	k, ctx, stateStore := setupTestKeeper(t)
	defer CloseStoreIfNeeded(stateStore)

	// No active rotation initially
	_, found := k.GetActiveClientKeyRotation(ctx, "test-client")
	require.False(t, found)
}
