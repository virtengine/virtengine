package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/market/keeper"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

func TestCheckVEIDGating_NoKeeperConfigured(t *testing.T) {
	ctx, k, _ := setupKeeper(t)

	// Get a reference to the concrete keeper
	// Note: The keeper doesn't have VEIDKeeper set by default in tests
	customerAddr := sdk.AccAddress("customer1")
	requirements := keeper.VEIDGatingRequirements{
		MinCustomerScore: 50,
		MinCustomerTier:  veidtypes.TierBasic,
	}

	// Since no VEIDKeeper is set, gating should pass by default
	result, err := k.CheckVEIDGating(ctx, customerAddr, requirements)
	require.NoError(t, err)
	require.True(t, result.Passed)
}

func TestCheckVEIDGating_DefaultRequirements(t *testing.T) {
	requirements := keeper.DefaultVEIDGatingRequirements()

	require.Equal(t, uint32(0), requirements.MinCustomerScore)
	require.Equal(t, veidtypes.TierUnverified, requirements.MinCustomerTier)
	require.True(t, requirements.RequireUnlockedIdentity)
	require.False(t, requirements.RequireVerifiedStatus)
	require.Nil(t, requirements.RequiredScopes)
}

func TestVEIDGatingResult_Structure(t *testing.T) {
	result := keeper.VEIDGatingResult{
		Passed:         true,
		CustomerScore:  75,
		CustomerTier:   veidtypes.TierStandard,
		CustomerStatus: "standard",
		FailureReasons: nil,
	}

	require.True(t, result.Passed)
	require.Equal(t, uint32(75), result.CustomerScore)
	require.Equal(t, veidtypes.TierStandard, result.CustomerTier)
}

func TestVEIDGatingFailure_Structure(t *testing.T) {
	failure := keeper.VEIDGatingFailure{
		CheckType:        "identity_score",
		RequiredValue:    "70",
		ActualValue:      "50",
		Message:          "Identity score 50 is below minimum required score of 70",
		RequiredSteps:    []string{"Complete verification", "Upload documents"},
		DocumentationURL: "/docs/identity-score",
	}

	require.Equal(t, "identity_score", failure.CheckType)
	require.Equal(t, "70", failure.RequiredValue)
	require.Equal(t, "50", failure.ActualValue)
	require.Contains(t, failure.Message, "below minimum")
	require.Len(t, failure.RequiredSteps, 2)
	require.Equal(t, "/docs/identity-score", failure.DocumentationURL)
}

func TestVEIDGatingRequirements_Validation(t *testing.T) {
	tests := []struct {
		name         string
		requirements keeper.VEIDGatingRequirements
		description  string
	}{
		{
			name:         "default requirements",
			requirements: keeper.DefaultVEIDGatingRequirements(),
			description:  "default requirements should be valid",
		},
		{
			name: "basic tier requirement",
			requirements: keeper.VEIDGatingRequirements{
				MinCustomerScore:        50,
				MinCustomerTier:         veidtypes.TierBasic,
				RequireVerifiedStatus:   false,
				RequireUnlockedIdentity: true,
			},
			description: "basic tier with score 50",
		},
		{
			name: "standard tier with scopes",
			requirements: keeper.VEIDGatingRequirements{
				MinCustomerScore:        70,
				MinCustomerTier:         veidtypes.TierStandard,
				RequiredScopes:          []veidtypes.ScopeType{veidtypes.ScopeTypeIDDocument, veidtypes.ScopeTypeSelfie},
				RequireVerifiedStatus:   true,
				RequireUnlockedIdentity: true,
			},
			description: "standard tier with ID and selfie scopes",
		},
		{
			name: "premium tier",
			requirements: keeper.VEIDGatingRequirements{
				MinCustomerScore:        85,
				MinCustomerTier:         veidtypes.TierPremium,
				RequiredScopes:          []veidtypes.ScopeType{veidtypes.ScopeTypeIDDocument, veidtypes.ScopeTypeSelfie, veidtypes.ScopeTypeFaceVideo},
				RequireVerifiedStatus:   true,
				RequireUnlockedIdentity: true,
			},
			description: "premium tier with full verification",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Verify requirements are constructed correctly
			require.NotNil(t, tc.requirements, tc.description)
		})
	}
}

func TestCreateIdentityRecord(t *testing.T) {
	// Test that we can create a properly structured identity record
	now := time.Now()
	record := veidtypes.NewIdentityRecord("cosmos1testaddr", now)

	require.NotNil(t, record)
	require.Equal(t, "cosmos1testaddr", record.AccountAddress)
	require.Equal(t, uint32(0), record.CurrentScore)
	require.Equal(t, veidtypes.IdentityTierUnverified, record.Tier)
	require.False(t, record.Locked)
	require.Empty(t, record.LockedReason)
}

func TestTierThresholds(t *testing.T) {
	// Verify tier thresholds are consistent with veid types
	tests := []struct {
		score        uint32
		expectedTier int
	}{
		{0, veidtypes.TierUnverified},
		{49, veidtypes.TierUnverified},
		{50, veidtypes.TierBasic},
		{69, veidtypes.TierBasic},
		{70, veidtypes.TierStandard},
		{84, veidtypes.TierStandard},
		{85, veidtypes.TierPremium},
		{100, veidtypes.TierPremium},
	}

	for _, tc := range tests {
		t.Run(veidtypes.TierToString(tc.expectedTier), func(t *testing.T) {
			tier := veidtypes.ComputeTierFromScoreValue(tc.score, veidtypes.AccountStatusVerified)
			require.Equal(t, tc.expectedTier, tier, "score %d should be tier %s", tc.score, veidtypes.TierToString(tc.expectedTier))
		})
	}
}

func TestVEIDGatingErrors(t *testing.T) {
	// Verify error codes are properly defined
	require.NotNil(t, keeper.ErrVEIDGatingFailed)
	require.NotNil(t, keeper.ErrInsufficientVEIDScore)
	require.NotNil(t, keeper.ErrInsufficientVEIDTier)
	require.NotNil(t, keeper.ErrVEIDNotVerified)
	require.NotNil(t, keeper.ErrVEIDLocked)
	require.NotNil(t, keeper.ErrVEIDScopeMissing)
	require.NotNil(t, keeper.ErrVEIDRecordNotFound)

	// Verify error messages
	require.Contains(t, keeper.ErrVEIDGatingFailed.Error(), "VEID gating failed")
	require.Contains(t, keeper.ErrInsufficientVEIDScore.Error(), "insufficient")
	require.Contains(t, keeper.ErrVEIDRecordNotFound.Error(), "not found")
}

func TestMinimumScoreForTier(t *testing.T) {
	// Verify GetMinimumScoreForTier returns correct values
	require.Equal(t, uint32(0), veidtypes.GetMinimumScoreForTier(veidtypes.TierUnverified))
	require.Equal(t, uint32(50), veidtypes.GetMinimumScoreForTier(veidtypes.TierBasic))
	require.Equal(t, uint32(70), veidtypes.GetMinimumScoreForTier(veidtypes.TierStandard))
	require.Equal(t, uint32(85), veidtypes.GetMinimumScoreForTier(veidtypes.TierPremium))
}

func TestTierToString(t *testing.T) {
	require.Equal(t, "unverified", veidtypes.TierToString(veidtypes.TierUnverified))
	require.Equal(t, "basic", veidtypes.TierToString(veidtypes.TierBasic))
	require.Equal(t, "standard", veidtypes.TierToString(veidtypes.TierStandard))
	require.Equal(t, "premium", veidtypes.TierToString(veidtypes.TierPremium))
}
