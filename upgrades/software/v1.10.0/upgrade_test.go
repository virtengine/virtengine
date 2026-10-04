package v1_10_0_test

import (
	"testing"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/app"
	v1_10_0 "github.com/virtengine/virtengine/upgrades/software/v1.10.0"
	utypes "github.com/virtengine/virtengine/upgrades/types"
	veidkeeper "github.com/virtengine/virtengine/x/veid/keeper"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// TestUpgradeHandlerLeavesExistingAccountsWithoutAVector is the migration
// guarantee: after the upgrade, an account that existed before it holds NO
// assurance vector, and that absence reads as "no assurance claim" rather than as
// a vector of zeros.
//
// This matters because a back-fill from the scalar score would invent per-factor
// values that were never measured, letting a relying party's factor-level policy
// (document > 9500 AND biometric > 9700) pass on fabricated evidence.
func TestUpgradeHandlerLeavesExistingAccountsWithoutAVector(t *testing.T) {
	application := app.Setup(app.WithChainID("veid-assurance-vector-upgrade-absent"))
	ctx := application.NewContext(false)
	from := application.MM.GetVersionMap()
	ctx = ctx.WithBlockHeight(23)

	constructor, found := utypes.GetUpgradesList()[v1_10_0.UpgradeName]
	require.True(t, found, "the assurance vector upgrade must be registered")
	up, err := constructor(application.Logger(), application.App)
	require.NoError(t, err)

	// An account with a pre-existing scalar score and NO vector: the exact shape
	// every account on a live chain has at the upgrade height.
	account := sdk.AccAddress([]byte("upgrade_pre_existing___")).String()
	require.NoError(t, application.Keepers.VirtEngine.VEID.SetScoreWithDetails(ctx, account, 88,
		veidkeeper.DefaultScoreDetails("pre-upgrade-model")))

	_, hadVector := application.Keepers.VirtEngine.VEID.GetAssuranceVector(ctx, account)
	require.False(t, hadVector, "the fixture must start without a vector")

	handler := up.UpgradeHandler()
	to, err := handler(ctx, upgradetypes.Plan{Name: v1_10_0.UpgradeName, Height: 23}, from)
	require.NoError(t, err)
	require.Equal(t, from, to, "the migration requires no module version changes")
	storeUpgrades := up.StoreLoader()
	require.Empty(t, storeUpgrades.Added, "the vector reuses the existing veid store, so no new store is added")
	require.Empty(t, storeUpgrades.Renamed)
	require.Empty(t, storeUpgrades.Deleted)

	// The scalar must be untouched: the migration adds state, it does not
	// reclassify any existing score.
	score, status, foundScore := application.Keepers.VirtEngine.VEID.GetScore(ctx, account)
	require.True(t, foundScore)
	require.Equal(t, uint32(88), score)
	require.Equal(t, "verified", string(status))

	// The vector must still be ABSENT after the upgrade.
	vector, hasVector := application.Keepers.VirtEngine.VEID.GetAssuranceVector(ctx, account)
	require.False(t, hasVector, "the migration must not invent a vector for an existing account")
	require.Nil(t, vector, "absence must read as no assurance claim, not a zeroed vector")

	// The query surface must report the absence explicitly rather than erroring,
	// so a relying party can tell "no claim" from a transport failure.
	queryResp, err := veidkeeper.NewGRPCQuerier(application.Keepers.VirtEngine.VEID).
		QueryAssuranceVector(ctx, &veidtypes.QueryAssuranceVectorRequest{AccountAddress: account})
	require.NoError(t, err)
	require.False(t, queryResp.Found)
	require.Nil(t, queryResp.Vector)
}

// TestUpgradeHandlerIsIdempotent proves re-running the handler changes nothing,
// which the upgrade framework relies on for a safe retry.
func TestUpgradeHandlerIsIdempotent(t *testing.T) {
	application := app.Setup(app.WithChainID("veid-assurance-vector-upgrade-idempotent"))
	ctx := application.NewContext(false)
	from := application.MM.GetVersionMap()
	ctx = ctx.WithBlockHeight(29)

	constructor, found := utypes.GetUpgradesList()[v1_10_0.UpgradeName]
	require.True(t, found)
	up, err := constructor(application.Logger(), application.App)
	require.NoError(t, err)

	handler := up.UpgradeHandler()
	plan := upgradetypes.Plan{Name: v1_10_0.UpgradeName, Height: 29}

	first, err := handler(ctx, plan, from)
	require.NoError(t, err)

	second, err := handler(ctx, plan, first)
	require.NoError(t, err)
	require.Equal(t, first, second, "a second run must be a no-op")

	third, err := handler(ctx, plan, second)
	require.NoError(t, err)
	require.Equal(t, second, third)
}

// TestUpgradeHandlerAllowsANewVectorAfterTheUpgrade proves the migration does not
// block the account's first real verification: once the account is re-verified, a
// vector appears, written at epoch 0 because none existed before.
func TestUpgradeHandlerAllowsANewVectorAfterTheUpgrade(t *testing.T) {
	application := app.Setup(app.WithChainID("veid-assurance-vector-upgrade-new"))
	ctx := application.NewContext(false)
	from := application.MM.GetVersionMap()
	ctx = ctx.WithBlockHeight(31)

	constructor, found := utypes.GetUpgradesList()[v1_10_0.UpgradeName]
	require.True(t, found)
	up, err := constructor(application.Logger(), application.App)
	require.NoError(t, err)

	account := sdk.AccAddress([]byte("upgrade_new_vector_____")).String()

	handler := up.UpgradeHandler()
	_, err = handler(ctx, upgradetypes.Plan{Name: v1_10_0.UpgradeName, Height: 31}, from)
	require.NoError(t, err)

	// No vector yet.
	_, hasVector := application.Keepers.VirtEngine.VEID.GetAssuranceVector(ctx, account)
	require.False(t, hasVector)

	// Re-verification produces one at epoch 0.
	result, err := application.Keepers.VirtEngine.VEID.ComputeAndStoreCompositeScore(ctx, account, upgradeCompositeInputs())
	require.NoError(t, err)

	vector, hasVector := application.Keepers.VirtEngine.VEID.GetAssuranceVector(ctx, account)
	require.True(t, hasVector, "the first post-upgrade verification must produce a vector")
	require.Equal(t, uint64(0), vector.Epoch, "the first vector an account ever holds is epoch 0")
	require.Equal(t, result.FinalScore, vector.ScalarScore)
	require.True(t, vector.VerifyCommitment())

	// A second verification supersedes it.
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 50)
	_, err = application.Keepers.VirtEngine.VEID.ComputeAndStoreCompositeScore(ctx, account, upgradeCompositeInputs())
	require.NoError(t, err)

	second, hasVector := application.Keepers.VirtEngine.VEID.GetAssuranceVector(ctx, account)
	require.True(t, hasVector)
	require.Equal(t, uint64(1), second.Epoch, "a re-verification must advance the epoch")
}

// upgradeCompositeInputs builds strong, fully present evidence so the composite
// scorer produces a high score and every factor is measured.
func upgradeCompositeInputs() veidtypes.CompositeScoringInputs {
	return veidtypes.CompositeScoringInputs{
		AccountAddress: "placeholder",
		DocumentAuthenticity: veidtypes.DocumentAuthenticityInput{
			Present:               true,
			TamperScore:           9500,
			FormatValidityScore:   9500,
			TemplateMatchScore:    9500,
			SecurityFeaturesScore: 9500,
		},
		FaceMatch: veidtypes.FaceMatchInput{
			Present:         true,
			SimilarityScore: 9800,
			Confidence:      9800,
			QualityScore:    9800,
		},
		LivenessDetection: veidtypes.LivenessDetectionInput{
			Present:              true,
			LivenessScore:        9600,
			BlinkDetected:        true,
			HeadMovementDetected: true,
			DepthCheckPassed:     true,
			AntiSpoofScore:       9600,
		},
		DataConsistency: veidtypes.DataConsistencyInput{
			Present:                 true,
			NameMatchScore:          10000,
			DOBConsistencyScore:     10000,
			AgeVerificationPassed:   true,
			AddressConsistencyScore: 10000,
			DocumentExpiryValid:     true,
			CrossFieldValidation:    10000,
		},
		HistoricalSignals: veidtypes.HistoricalSignalsInput{
			Present:                    true,
			PriorVerificationScore:     8000,
			AccountAgeScore:            8000,
			VerificationHistoryCount:   9,
			SuccessfulVerificationRate: 9000,
		},
		RiskIndicators: veidtypes.RiskIndicatorsInput{
			Present:                 true,
			FraudPatternScore:       2000,
			DeviceFingerprintScore:  9000,
			DeviceIntegrityScore:    9000,
			IPReputationScore:       9000,
			VelocityCheckPassed:     true,
			DeviceAttestationPassed: true,
			GeoConsistencyScore:     9000,
		},
	}
}
