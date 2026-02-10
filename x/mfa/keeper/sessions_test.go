package keeper

import (
	"encoding/json"
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/mfa/types"
)

// mockVEIDKeeper implements VEIDKeeper for testing
type mockVEIDKeeper struct {
	scores map[string]uint32
}

func newMockVEIDKeeper() *mockVEIDKeeper {
	return &mockVEIDKeeper{scores: make(map[string]uint32)}
}

func (m *mockVEIDKeeper) GetVEIDScore(ctx sdk.Context, address sdk.AccAddress) (uint32, bool) {
	score, ok := m.scores[address.String()]
	return score, ok
}

// mockRolesKeeper implements RolesKeeper for testing
type mockRolesKeeper struct {
	operational map[string]bool
}

func newMockRolesKeeper() *mockRolesKeeper {
	return &mockRolesKeeper{operational: make(map[string]bool)}
}

func (m *mockRolesKeeper) IsAccountOperational(ctx sdk.Context, address sdk.AccAddress) bool {
	op, ok := m.operational[address.String()]
	return ok && op
}

func setupTestKeeper(t *testing.T) (sdk.Context, Keeper) {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := sdktestutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockTime(time.Now())

	cdc := codec.NewProtoCodec(nil)
	keeper := NewKeeper(cdc, key, "authority", newMockVEIDKeeper(), newMockRolesKeeper())

	return ctx, keeper
}

func TestHasValidAuthSession(t *testing.T) {
	ctx, keeper := setupTestKeeper(t)
	addr := sdk.AccAddress([]byte("test_address_1234567"))

	// Test: No session exists
	hasSession := keeper.HasValidAuthSession(ctx, addr, types.SensitiveTxProviderRegistration)
	require.False(t, hasSession, "should return false when no session exists")

	// Create a session
	now := ctx.BlockTime().Unix()
	session := &types.AuthorizationSession{
		SessionID:       "test-session-1",
		AccountAddress:  addr.String(),
		TransactionType: types.SensitiveTxProviderRegistration,
		VerifiedFactors: []types.FactorType{types.FactorTypeFIDO2},
		CreatedAt:       now,
		ExpiresAt:       now + 15*60, // 15 minutes
		IsSingleUse:     false,
	}
	err := keeper.CreateAuthorizationSession(ctx, session)
	require.NoError(t, err)

	// Test: Valid session exists
	hasSession = keeper.HasValidAuthSession(ctx, addr, types.SensitiveTxProviderRegistration)
	require.True(t, hasSession, "should return true when valid session exists")

	// Test: Wrong action type
	hasSession = keeper.HasValidAuthSession(ctx, addr, types.SensitiveTxKeyRotation)
	require.False(t, hasSession, "should return false for different action type")
}

func TestHasValidAuthSessionWithDevice(t *testing.T) {
	ctx, keeper := setupTestKeeper(t)
	addr := sdk.AccAddress([]byte("test_address_1234567"))

	now := ctx.BlockTime().Unix()
	session := &types.AuthorizationSession{
		SessionID:         "test-session-device",
		AccountAddress:    addr.String(),
		TransactionType:   types.SensitiveTxLargeWithdrawal,
		VerifiedFactors:   []types.FactorType{types.FactorTypeFIDO2, types.FactorTypeVEID},
		CreatedAt:         now,
		ExpiresAt:         now + 15*60,
		IsSingleUse:       false,
		DeviceFingerprint: "device-fingerprint-abc123",
	}
	err := keeper.CreateAuthorizationSession(ctx, session)
	require.NoError(t, err)

	// Test: Correct device fingerprint
	hasSession := keeper.HasValidAuthSessionWithDevice(ctx, addr, types.SensitiveTxLargeWithdrawal, "device-fingerprint-abc123")
	require.True(t, hasSession, "should return true for matching device fingerprint")

	// Test: Wrong device fingerprint
	hasSession = keeper.HasValidAuthSessionWithDevice(ctx, addr, types.SensitiveTxLargeWithdrawal, "wrong-device")
	require.False(t, hasSession, "should return false for non-matching device fingerprint")

	// Test: Empty device fingerprint (should match any)
	hasSession = keeper.HasValidAuthSession(ctx, addr, types.SensitiveTxLargeWithdrawal)
	require.True(t, hasSession, "should return true when not checking device fingerprint")
}

func TestConsumeAuthSession_SingleUse(t *testing.T) {
	ctx, keeper := setupTestKeeper(t)
	addr := sdk.AccAddress([]byte("test_address_1234567"))

	now := ctx.BlockTime().Unix()
	session := &types.AuthorizationSession{
		SessionID:       "test-single-use",
		AccountAddress:  addr.String(),
		TransactionType: types.SensitiveTxAccountRecovery,
		VerifiedFactors: []types.FactorType{types.FactorTypeVEID, types.FactorTypeFIDO2, types.FactorTypeSMS},
		CreatedAt:       now,
		ExpiresAt:       now + 5*60, // 5 minutes for single-use
		IsSingleUse:     true,
	}
	err := keeper.CreateAuthorizationSession(ctx, session)
	require.NoError(t, err)

	// Verify session exists
	hasSession := keeper.HasValidAuthSession(ctx, addr, types.SensitiveTxAccountRecovery)
	require.True(t, hasSession)

	// Consume the session
	err = keeper.ConsumeAuthSession(ctx, addr, types.SensitiveTxAccountRecovery)
	require.NoError(t, err)

	// Session should no longer be valid (consumed)
	hasSession = keeper.HasValidAuthSession(ctx, addr, types.SensitiveTxAccountRecovery)
	require.False(t, hasSession, "single-use session should be consumed after use")

	// Trying to consume again should fail
	err = keeper.ConsumeAuthSession(ctx, addr, types.SensitiveTxAccountRecovery)
	require.Error(t, err)
}

func TestConsumeAuthSession_MultiUse(t *testing.T) {
	ctx, keeper := setupTestKeeper(t)
	addr := sdk.AccAddress([]byte("test_address_1234567"))

	now := ctx.BlockTime().Unix()
	session := &types.AuthorizationSession{
		SessionID:       "test-multi-use",
		AccountAddress:  addr.String(),
		TransactionType: types.SensitiveTxHighValueOrder,
		VerifiedFactors: []types.FactorType{types.FactorTypeVEID, types.FactorTypeFIDO2},
		CreatedAt:       now,
		ExpiresAt:       now + 30*60, // 30 minutes
		IsSingleUse:     false,
	}
	err := keeper.CreateAuthorizationSession(ctx, session)
	require.NoError(t, err)

	// Consume the session
	err = keeper.ConsumeAuthSession(ctx, addr, types.SensitiveTxHighValueOrder)
	require.NoError(t, err)

	// Session should still be valid (multi-use)
	hasSession := keeper.HasValidAuthSession(ctx, addr, types.SensitiveTxHighValueOrder)
	require.True(t, hasSession, "multi-use session should remain valid after use")

	// Should be able to consume again
	err = keeper.ConsumeAuthSession(ctx, addr, types.SensitiveTxHighValueOrder)
	require.NoError(t, err)
}

func TestConsumeAuthSession_DeviceMismatch(t *testing.T) {
	ctx, keeper := setupTestKeeper(t)
	addr := sdk.AccAddress([]byte("test_address_1234567"))

	now := ctx.BlockTime().Unix()
	session := &types.AuthorizationSession{
		SessionID:         "test-device-bound",
		AccountAddress:    addr.String(),
		TransactionType:   types.SensitiveTxProviderRegistration,
		VerifiedFactors:   []types.FactorType{types.FactorTypeVEID, types.FactorTypeFIDO2},
		CreatedAt:         now,
		ExpiresAt:         now + 15*60,
		IsSingleUse:       false,
		DeviceFingerprint: "original-device",
	}
	err := keeper.CreateAuthorizationSession(ctx, session)
	require.NoError(t, err)

	// Try to consume with wrong device
	err = keeper.ConsumeAuthSessionWithDevice(ctx, addr, types.SensitiveTxProviderRegistration, "different-device")
	require.Error(t, err)
	require.Contains(t, err.Error(), "device fingerprint")

	// Consume with correct device should work
	err = keeper.ConsumeAuthSessionWithDevice(ctx, addr, types.SensitiveTxProviderRegistration, "original-device")
	require.NoError(t, err)
}

func TestCreateAuthSessionForAction(t *testing.T) {
	ctx, keeper := setupTestKeeper(t)
	addr := sdk.AccAddress([]byte("test_address_1234567"))

	testCases := []struct {
		name         string
		action       types.SensitiveTransactionType
		expectSingle bool
		minDuration  int64
		maxDuration  int64
	}{
		{
			name:         "Critical - AccountRecovery (single-use)",
			action:       types.SensitiveTxAccountRecovery,
			expectSingle: true,
			minDuration:  0,
			maxDuration:  5 * 60, // 5 minute window for single-use
		},
		{
			name:         "Critical - KeyRotation (single-use)",
			action:       types.SensitiveTxKeyRotation,
			expectSingle: true,
			minDuration:  0,
			maxDuration:  5 * 60,
		},
		{
			name:         "Medium - ProviderRegistration (30 min)",
			action:       types.SensitiveTxProviderRegistration,
			expectSingle: false,
			minDuration:  29 * 60,
			maxDuration:  31 * 60,
		},
		{
			name:         "High - LargeWithdrawal (15 min)",
			action:       types.SensitiveTxLargeWithdrawal,
			expectSingle: false,
			minDuration:  14 * 60,
			maxDuration:  16 * 60,
		},
		{
			name:         "Medium - HighValueOrder (30 min)",
			action:       types.SensitiveTxHighValueOrder,
			expectSingle: false,
			minDuration:  29 * 60,
			maxDuration:  31 * 60,
		},
		{
			name:         "Low - MediumWithdrawal (60 min)",
			action:       types.SensitiveTxMediumWithdrawal,
			expectSingle: false,
			minDuration:  59 * 60,
			maxDuration:  61 * 60,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			session, err := keeper.CreateAuthSessionForAction(
				ctx,
				addr,
				tc.action,
				[]types.FactorType{types.FactorTypeFIDO2},
				"test-device",
			)
			require.NoError(t, err)
			require.NotNil(t, session)

			require.Equal(t, tc.expectSingle, session.IsSingleUse, "single-use mismatch")

			duration := session.ExpiresAt - session.CreatedAt
			require.GreaterOrEqual(t, duration, tc.minDuration, "duration too short")
			require.LessOrEqual(t, duration, tc.maxDuration, "duration too long")

			// Cleanup
			_ = keeper.DeleteAuthorizationSession(ctx, session.SessionID)
		})
	}
}

func TestGetValidSessionsForAccount(t *testing.T) {
	ctx, keeper := setupTestKeeper(t)
	addr := sdk.AccAddress([]byte("test_address_1234567"))

	now := ctx.BlockTime().Unix()

	// Create mix of valid and expired/used sessions
	validSession := &types.AuthorizationSession{
		SessionID:       "valid-session",
		AccountAddress:  addr.String(),
		TransactionType: types.SensitiveTxProviderRegistration,
		VerifiedFactors: []types.FactorType{types.FactorTypeFIDO2},
		CreatedAt:       now,
		ExpiresAt:       now + 15*60,
		IsSingleUse:     false,
	}
	_ = keeper.CreateAuthorizationSession(ctx, validSession)

	expiredSession := &types.AuthorizationSession{
		SessionID:       "expired-session",
		AccountAddress:  addr.String(),
		TransactionType: types.SensitiveTxHighValueOrder,
		VerifiedFactors: []types.FactorType{types.FactorTypeTOTP},
		CreatedAt:       now - 60*60,
		ExpiresAt:       now - 30*60, // Expired 30 min ago
		IsSingleUse:     false,
	}
	_ = keeper.CreateAuthorizationSession(ctx, expiredSession)

	usedSingleUseSession := &types.AuthorizationSession{
		SessionID:       "used-single-use",
		AccountAddress:  addr.String(),
		TransactionType: types.SensitiveTxAccountRecovery,
		VerifiedFactors: []types.FactorType{types.FactorTypeVEID, types.FactorTypeFIDO2},
		CreatedAt:       now,
		ExpiresAt:       now + 5*60,
		IsSingleUse:     true,
		UsedAt:          now - 60, // Used 1 minute ago
	}
	_ = keeper.CreateAuthorizationSession(ctx, usedSingleUseSession)

	// Get valid sessions
	validSessions := keeper.GetValidSessionsForAccount(ctx, addr)
	require.Len(t, validSessions, 1, "should only return 1 valid session")
	require.Equal(t, "valid-session", validSessions[0].SessionID)
}

func TestCleanupExpiredSessions(t *testing.T) {
	ctx, keeper := setupTestKeeper(t)
	addr := sdk.AccAddress([]byte("test_address_1234567"))

	now := ctx.BlockTime().Unix()

	// Create sessions
	_ = keeper.CreateAuthorizationSession(ctx, &types.AuthorizationSession{
		SessionID:       "valid-session",
		AccountAddress:  addr.String(),
		TransactionType: types.SensitiveTxProviderRegistration,
		CreatedAt:       now,
		ExpiresAt:       now + 15*60,
	})

	_ = keeper.CreateAuthorizationSession(ctx, &types.AuthorizationSession{
		SessionID:       "expired-session-1",
		AccountAddress:  addr.String(),
		TransactionType: types.SensitiveTxHighValueOrder,
		CreatedAt:       now - 60*60,
		ExpiresAt:       now - 30*60,
	})

	_ = keeper.CreateAuthorizationSession(ctx, &types.AuthorizationSession{
		SessionID:       "expired-session-2",
		AccountAddress:  addr.String(),
		TransactionType: types.SensitiveTxMediumWithdrawal,
		CreatedAt:       now - 120*60,
		ExpiresAt:       now - 60*60,
	})

	// Verify initial state
	allSessions := keeper.GetAccountSessions(ctx, addr)
	require.Len(t, allSessions, 3)

	// Cleanup expired sessions
	deleted := keeper.CleanupExpiredSessions(ctx, addr)
	require.Equal(t, 2, deleted, "should delete 2 expired sessions")

	// Verify only valid session remains
	remainingSessions := keeper.GetAccountSessions(ctx, addr)
	require.Len(t, remainingSessions, 1)
	require.Equal(t, "valid-session", remainingSessions[0].SessionID)
}

func TestValidateSessionForTransaction(t *testing.T) {
	ctx, keeper := setupTestKeeper(t)
	addr := sdk.AccAddress([]byte("test_address_1234567"))
	otherAddr := sdk.AccAddress([]byte("other_address_123456"))

	now := ctx.BlockTime().Unix()
	session := &types.AuthorizationSession{
		SessionID:         "validate-test-session",
		AccountAddress:    addr.String(),
		TransactionType:   types.SensitiveTxProviderRegistration,
		VerifiedFactors:   []types.FactorType{types.FactorTypeFIDO2},
		CreatedAt:         now,
		ExpiresAt:         now + 15*60,
		IsSingleUse:       false,
		DeviceFingerprint: "bound-device",
	}
	_ = keeper.CreateAuthorizationSession(ctx, session)

	// Test: Valid session
	err := keeper.ValidateSessionForTransaction(ctx, "validate-test-session", addr, types.SensitiveTxProviderRegistration, "bound-device")
	require.NoError(t, err)

	// Test: Session not found
	err = keeper.ValidateSessionForTransaction(ctx, "non-existent", addr, types.SensitiveTxProviderRegistration, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")

	// Test: Wrong account
	err = keeper.ValidateSessionForTransaction(ctx, "validate-test-session", otherAddr, types.SensitiveTxProviderRegistration, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not belong")

	// Test: Wrong action type
	err = keeper.ValidateSessionForTransaction(ctx, "validate-test-session", addr, types.SensitiveTxKeyRotation, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "action")

	// Test: Device mismatch
	err = keeper.ValidateSessionForTransaction(ctx, "validate-test-session", addr, types.SensitiveTxProviderRegistration, "wrong-device")
	require.Error(t, err)
	require.Contains(t, err.Error(), "device")
}

func TestGetSessionDurationForAction(t *testing.T) {
	ctx, keeper := setupTestKeeper(t)

	testCases := []struct {
		action           types.SensitiveTransactionType
		expectedDuration int64
	}{
		{types.SensitiveTxAccountRecovery, 0},            // Critical - single use
		{types.SensitiveTxKeyRotation, 0},                // Critical - single use
		{types.SensitiveTxProviderRegistration, 30 * 60}, // Medium - 30 min
		{types.SensitiveTxLargeWithdrawal, 15 * 60},      // High - 15 min
		{types.SensitiveTxHighValueOrder, 30 * 60},       // Medium - 30 min
		{types.SensitiveTxMediumWithdrawal, 60 * 60},     // Low - 60 min
	}

	for _, tc := range testCases {
		t.Run(tc.action.String(), func(t *testing.T) {
			duration := keeper.GetSessionDurationForAction(ctx, tc.action)
			require.Equal(t, tc.expectedDuration, duration)
		})
	}
}

func TestIsActionSingleUse(t *testing.T) {
	ctx, keeper := setupTestKeeper(t)

	singleUseActions := []types.SensitiveTransactionType{
		types.SensitiveTxAccountRecovery,
		types.SensitiveTxKeyRotation,
		types.SensitiveTxAccountDeletion,
		types.SensitiveTxTwoFactorDisable,
		types.SensitiveTxValidatorRegistration,
		types.SensitiveTxRoleAssignment,
	}

	multiUseActions := []types.SensitiveTransactionType{
		types.SensitiveTxProviderRegistration,
		types.SensitiveTxLargeWithdrawal,
		types.SensitiveTxHighValueOrder,
		types.SensitiveTxMediumWithdrawal,
		types.SensitiveTxGovernanceProposal,
	}

	for _, action := range singleUseActions {
		t.Run(action.String()+"_single", func(t *testing.T) {
			isSingle := keeper.IsActionSingleUse(ctx, action)
			require.True(t, isSingle, "%s should be single-use", action.String())
		})
	}

	for _, action := range multiUseActions {
		t.Run(action.String()+"_multi", func(t *testing.T) {
			isSingle := keeper.IsActionSingleUse(ctx, action)
			require.False(t, isSingle, "%s should be multi-use", action.String())
		})
	}
}

func TestCustomSensitiveTxConfig(t *testing.T) {
	ctx, keeper := setupTestKeeper(t)
	addr := sdk.AccAddress([]byte("test_address_1234567"))

	// Set custom config for HighValueOrder with different duration
	customConfig := &types.SensitiveTxConfig{
		TransactionType: types.SensitiveTxHighValueOrder,
		Enabled:         true,
		MinVEIDScore:    80,
		SessionDuration: 10 * 60, // 10 minutes instead of default 30
		IsSingleUse:     true,    // Override to single-use
		Description:     "Custom high-value order config",
	}
	err := keeper.SetSensitiveTxConfig(ctx, customConfig)
	require.NoError(t, err)

	// Create session should use custom config
	session, err := keeper.CreateAuthSessionForAction(
		ctx,
		addr,
		types.SensitiveTxHighValueOrder,
		[]types.FactorType{types.FactorTypeFIDO2},
		"",
	)
	require.NoError(t, err)

	// Should be single-use (overridden)
	require.True(t, session.IsSingleUse)

	// Duration check helpers
	duration := keeper.GetSessionDurationForAction(ctx, types.SensitiveTxHighValueOrder)
	require.Equal(t, int64(10*60), duration)

	isSingle := keeper.IsActionSingleUse(ctx, types.SensitiveTxHighValueOrder)
	require.True(t, isSingle)
}

// Verify sessionStore JSON serialization works correctly
func TestSessionStoreSerialization(t *testing.T) {
	ss := sessionStore{
		SessionID:         "test-id",
		AccountAddress:    "cosmos1abc123",
		TransactionType:   types.SensitiveTxProviderRegistration,
		VerifiedFactors:   []types.FactorType{types.FactorTypeFIDO2, types.FactorTypeVEID},
		CreatedAt:         1700000000,
		ExpiresAt:         1700000900,
		UsedAt:            0,
		IsSingleUse:       false,
		DeviceFingerprint: "device-123",
	}

	bz, err := json.Marshal(&ss)
	require.NoError(t, err)

	var decoded sessionStore
	err = json.Unmarshal(bz, &decoded)
	require.NoError(t, err)

	require.Equal(t, ss.SessionID, decoded.SessionID)
	require.Equal(t, ss.TransactionType, decoded.TransactionType)
	require.Equal(t, ss.VerifiedFactors, decoded.VerifiedFactors)
	require.Equal(t, ss.DeviceFingerprint, decoded.DeviceFingerprint)
}
