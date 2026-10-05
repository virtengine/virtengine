package app

import (
	"io"
	"testing"

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

	veidkeeper "github.com/virtengine/virtengine/x/veid/keeper"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Assurance-vector regression guard for the VEID ante gate
// ============================================================================
//
// Introducing the per-factor assurance vector added new persisted state to
// x/veid. This file proves the pre-existing GATING behaviour is untouched by it:
//
//   - checkVEIDRequirements still reads only the scalar score, so an account that
//     has a vector is gated exactly as an account with the same scalar and no
//     vector is;
//   - an account with NO vector is gated exactly as before, since absence is not
//     a score and must not change any verdict.
//
// Together these assert the invariant the vector work must not break: the vector
// is additive information for relying parties, never an input to consensus
// gating.

// newVEIDGateTestKeeper builds a real x/veid keeper over an in-memory store.
func newVEIDGateTestKeeper(t *testing.T) (veidkeeper.Keeper, sdk.Context, func()) {
	t.Helper()

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	veidtypes.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(veidtypes.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Height: 1000,
	}, false, log.NewNopLogger())

	k := veidkeeper.NewKeeper(cdc, storeKey, "authority")
	require.NoError(t, k.SetParams(ctx, veidtypes.DefaultParams()))

	// The IAVL store starts background pruning goroutines; stop them if the
	// underlying store supports closing. A checked assertion, because
	// rootmulti.Store is not itself an io.Closer.
	return k, ctx, func() {
		if closer, ok := stateStore.(io.Closer); ok {
			_ = closer.Close()
		}
	}
}

// compositeInputs builds distinguishable evidence for the composite scorer.
func compositeInputs() veidtypes.CompositeScoringInputs {
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

// TestAnteGateIgnoresAssuranceVector proves that whether an account holds a
// vector has no effect on the scalar gating decision.
func TestAnteGateIgnoresAssuranceVector(t *testing.T) {
	k, ctx, cleanup := newVEIDGateTestKeeper(t)
	defer cleanup()

	withVector := sdk.AccAddress([]byte("ante_with_vector______")).String()
	withoutVector := sdk.AccAddress([]byte("ante_without_vector___")).String()

	// Score ONLY the first account, and let it also acquire a vector through the
	// normal composite path.
	result, err := k.ComputeAndStoreCompositeScore(ctx, withVector, compositeInputs())
	require.NoError(t, err)
	require.True(t, result.Passed)

	vector, found := k.GetAssuranceVector(ctx, withVector)
	require.True(t, found, "the scored account must hold a vector")
	require.True(t, vector.VerifyCommitment())

	// The second account has the same score requirement but no vector and no
	// score at all; it is the "before the vector existed" shape.
	_, hasVector := k.GetAssuranceVector(ctx, withoutVector)
	require.False(t, hasVector)

	t.Run("threshold at the score passes for a vector holder", func(t *testing.T) {
		require.True(t, k.IsScoreAboveThreshold(ctx, withVector, result.FinalScore),
			"gating at exactly the recorded score must pass")
	})

	t.Run("threshold above the score fails even with a strong vector", func(t *testing.T) {
		require.False(t, k.IsScoreAboveThreshold(ctx, withVector, result.FinalScore+1),
			"a vector must not let an account past a threshold its scalar misses")
	})

	t.Run("threshold above 100 always fails", func(t *testing.T) {
		require.False(t, k.IsScoreAboveThreshold(ctx, withVector, 101))
	})

	t.Run("an account with no vector and no score fails every threshold", func(t *testing.T) {
		require.False(t, k.IsScoreAboveThreshold(ctx, withoutVector, 0),
			"absence is not a passing score")
		require.False(t, k.IsScoreAboveThreshold(ctx, withoutVector, 1))
	})

	// The decisive pair: identical scalar, opposite vector presence, identical
	// verdict. If the ante path ever started reading the vector, these would
	// diverge.
	t.Run("vector presence does not change the verdict", func(t *testing.T) {
		// Give the second account the SAME scalar score as the first, but no
		// vector, by writing the score directly.
		require.NoError(t, k.SetScoreWithDetails(ctx, withoutVector, result.FinalScore, veidkeeper.ScoreDetails{
			Status:           veidtypes.AccountStatusVerified,
			ModelVersion:     result.ScoreVersion,
			VerificationHash: result.InputHash,
		}))

		_, secondHasVector := k.GetAssuranceVector(ctx, withoutVector)
		require.False(t, secondHasVector, "this account is scored without a vector")

		require.True(t, k.IsScoreAboveThreshold(ctx, withVector, result.FinalScore))
		require.True(t, k.IsScoreAboveThreshold(ctx, withoutVector, result.FinalScore),
			"the same scalar must gate identically whether or not a vector exists")
	})
}

// TestAnteGateRejectsUnverifiedStatusRegardlessOfVector proves the status check is
// still authoritative: a vector does not substitute for verified status.
func TestAnteGateRejectsUnverifiedStatusRegardlessOfVector(t *testing.T) {
	k, ctx, cleanup := newVEIDGateTestKeeper(t)
	defer cleanup()

	account := sdk.AccAddress([]byte("ante_rejected_status____")).String()

	result, err := k.ComputeAndStoreCompositeScore(ctx, account, compositeInputs())
	require.NoError(t, err)
	require.True(t, k.IsScoreAboveThreshold(ctx, account, 1))

	// Force the status back to rejected while leaving the vector in place.
	require.NoError(t, k.SetScoreWithDetails(ctx, account, 90, veidkeeper.ScoreDetails{
		Status:           veidtypes.AccountStatusRejected,
		ModelVersion:     result.ScoreVersion,
		VerificationHash: result.InputHash,
	}))

	_, hasVector := k.GetAssuranceVector(ctx, account)
	require.True(t, hasVector, "the vector survives the status change")

	require.False(t, k.IsScoreAboveThreshold(ctx, account, 1),
		"a present vector must not rescue an unverified status")
}
