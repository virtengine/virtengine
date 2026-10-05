package types_test

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Fixture
// ============================================================================

// distinguishableInputs builds a composite scoring fixture whose factors are
// DELIBERATELY different from one another, so a test that asserts "the vector
// matches each factor" cannot pass by accident on uniform values.
func distinguishableInputs(t *testing.T) types.CompositeScoringInputs {
	t.Helper()

	return types.CompositeScoringInputs{
		AccountAddress: testVectorAccount,
		// Document: four sub-signals, deliberately not equal to any other factor.
		DocumentAuthenticity: types.DocumentAuthenticityInput{
			Present:               true,
			TamperScore:           9000,
			FormatValidityScore:   8000,
			TemplateMatchScore:    7000,
			SecurityFeaturesScore: 6000,
		},
		// Face match: distinct values again.
		FaceMatch: types.FaceMatchInput{
			Present:         true,
			SimilarityScore: 9600,
			Confidence:      9200,
			QualityScore:    8800,
		},
		LivenessDetection: types.LivenessDetectionInput{
			Present:              true,
			LivenessScore:        7500,
			BlinkDetected:        true,
			HeadMovementDetected: true,
			DepthCheckPassed:     true,
			AntiSpoofScore:       7200,
		},
		DataConsistency: types.DataConsistencyInput{
			Present:                 true,
			NameMatchScore:          10000,
			DOBConsistencyScore:     10000,
			AgeVerificationPassed:   true,
			AddressConsistencyScore: 10000,
			DocumentExpiryValid:     true,
			CrossFieldValidation:    10000,
		},
		HistoricalSignals: types.HistoricalSignalsInput{
			Present:                    true,
			PriorVerificationScore:     4000,
			AccountAgeScore:            3000,
			VerificationHistoryCount:   3,
			SuccessfulVerificationRate: 5000,
		},
		RiskIndicators: types.RiskIndicatorsInput{
			Present:                 true,
			FraudPatternScore:       2000,
			DeviceFingerprintScore:  8500,
			DeviceIntegrityScore:    8100,
			IPReputationScore:       7800,
			VelocityCheckPassed:     true,
			DeviceAttestationPassed: true,
			GeoConsistencyScore:     7600,
		},
	}
}

const testVectorAccount = "virtengine1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq"

// deriveFixtureVector computes a composite score from the fixture and derives
// the vector from it, returning both so a test can cross-check them.
func deriveFixtureVector(t *testing.T) (*types.CompositeScoreResult, *types.AssuranceVector) {
	t.Helper()

	inputs := distinguishableInputs(t)
	result, err := types.ComputeCompositeScore(
		inputs,
		types.DefaultCompositeScoringWeights(),
		types.DefaultCompositeScoringThresholds(),
	)
	require.NoError(t, err)

	vector, err := types.DeriveAssuranceVector(result, types.DerivationInput{
		Account:       testVectorAccount,
		Epoch:         0,
		ModelVersion:  "test-model-v1",
		AccountAgeBps: 4200,
		Device:        types.DeviceEvidence{FingerprintBps: 8500, IntegrityBps: 8100, AttestationBps: 9000, Present: true},
		VerifiedAt:    time.Unix(1_700_000_000, 0).UTC(),
	})
	require.NoError(t, err)

	return result, vector
}

// ============================================================================
// Derivation matches the composite result factor by factor
// ============================================================================

func TestDeriveAssuranceVectorMatchesEachFactorAndWeight(t *testing.T) {
	result, vector := deriveFixtureVector(t)

	require.NoError(t, vector.Validate())

	// The vector must carry every canonical factor, in canonical order.
	require.Len(t, vector.Factors, len(types.AssuranceFactors))
	for i, entry := range vector.Factors {
		require.Equal(t, types.AssuranceFactors[i], entry.Factor, "factors must be in canonical order")
	}

	// Build the expected per-factor readings straight off the composite
	// contributions, so the assertion is against the scorer's own output rather
	// than a hand-written copy of the same numbers.
	byComponent := make(map[string]types.CompositeScoreContribution, len(result.Contributions))
	for _, c := range result.Contributions {
		byComponent[c.ComponentName] = c
	}

	for _, tc := range []struct {
		factor    types.AssuranceFactor
		component string
	}{
		{types.AssuranceFactorDocument, types.ComponentDocumentAuthenticity},
		{types.AssuranceFactorBiometric, types.ComponentFaceMatch},
		{types.AssuranceFactorLiveness, types.ComponentLivenessDetection},
		{types.AssuranceFactorHistory, types.ComponentHistoricalSignals},
		{types.AssuranceFactorDataConsistency, types.ComponentDataConsistency},
		{types.AssuranceFactorRisk, types.ComponentRiskIndicators},
	} {
		entry, ok := vector.Lookup(tc.factor)
		require.True(t, ok, "factor %s must be present", tc.factor)

		contrib, ok := byComponent[tc.component]
		require.True(t, ok, "component %s must exist on the result", tc.component)

		require.Equal(t, contrib.RawScore, entry.ScoreBps, "factor %s score must equal its component raw score", tc.factor)
		require.Equal(t, contrib.Weight, entry.WeightBps, "factor %s weight must equal its component weight", tc.factor)
		require.Equal(t, contrib.WeightedScore, entry.WeightedScoreBps, "factor %s weighted score must match", tc.factor)
		require.True(t, entry.Measured, "factor %s must be measured for a present component", tc.factor)
	}

	// Overall must be the same measurement as the scalar, re-expressed in bps.
	require.Equal(t, result.FinalScore, vector.ScalarScore, "scalar score must be carried verbatim")
	require.Equal(t, result.FinalScore*100, vector.OverallBps, "overall bps must be the scalar in basis points")
	require.Equal(t, result.Passed, vector.Passed)

	// Versions are carried, not invented.
	require.Equal(t, result.ScoreVersion, vector.ScoreVersion)
	require.Equal(t, "test-model-v1", vector.ModelVersion)

	// The device factor is derived from the device sub-signals:
	// (8500 + 8100 + 9000) / 3 = 8533.
	device, ok := vector.Lookup(types.AssuranceFactorDevice)
	require.True(t, ok)
	require.True(t, device.Measured, "device evidence was supplied, so the factor is measured")
	require.Equal(t, uint32(8533), device.ScoreBps, "device score is the truncating mean of its sub-signals")
	require.Equal(t, byComponent[types.ComponentRiskIndicators].Weight, device.WeightBps,
		"device rides on the risk component's weight")

	// Account age and recency are carried from the derivation input.
	require.Equal(t, uint32(4200), vector.AccountAgeBps)
	require.Equal(t, result.BlockHeight, vector.VerifiedHeight)
	require.Equal(t, time.Unix(1_700_000_000, 0).UTC().Unix(), vector.VerifiedAtUnix)
	require.Equal(t, result.InputHash, vector.InputHash)

	// A fully measured vector's distinct component weights must total 10000.
	require.Equal(t, uint32(10000), vector.TotalComponentWeight(),
		"distinct component weights must sum to the full weight, with the shared risk weight counted once")
}

func TestDeriveAssuranceVectorIsDeterministic(t *testing.T) {
	_, first := deriveFixtureVector(t)
	_, second := deriveFixtureVector(t)

	require.Equal(t, first.Commitment, second.Commitment, "the same inputs must produce the same commitment")
	require.Equal(t, first.Factors, second.Factors)
}

func TestDeriveAssuranceVectorCommitmentDetectsTampering(t *testing.T) {
	_, vector := deriveFixtureVector(t)
	require.True(t, vector.VerifyCommitment())

	// Flip a factor's score: the commitment must no longer verify, so a relying
	// party cannot be handed a vector whose hash disagrees with its contents.
	tampered := *vector
	tampered.Factors = append([]types.AssuranceFactorEntry(nil), vector.Factors...)
	tampered.Factors[0].ScoreBps++

	require.False(t, tampered.VerifyCommitment(), "a tampered factor must break the commitment")
	require.Error(t, tampered.Validate(), "a tampered vector must fail validation")
}

func TestDeriveAssuranceVectorRejectsInvalidInput(t *testing.T) {
	inputs := distinguishableInputs(t)
	result, err := types.ComputeCompositeScore(
		inputs,
		types.DefaultCompositeScoringWeights(),
		types.DefaultCompositeScoringThresholds(),
	)
	require.NoError(t, err)

	t.Run("nil result", func(t *testing.T) {
		_, err := types.DeriveAssuranceVector(nil, types.DerivationInput{Account: testVectorAccount})
		require.Error(t, err)
	})

	t.Run("empty account", func(t *testing.T) {
		_, err := types.DeriveAssuranceVector(result, types.DerivationInput{Account: ""})
		require.Error(t, err)
	})

	t.Run("account age out of range", func(t *testing.T) {
		_, err := types.DeriveAssuranceVector(result, types.DerivationInput{
			Account:       testVectorAccount,
			AccountAgeBps: uint32(types.MaxBasisPoints) + 1,
		})
		require.Error(t, err, "an out-of-range basis point must be rejected, not clamped")
	})

	t.Run("duplicate contribution", func(t *testing.T) {
		dup := *result
		dup.Contributions = append(append([]types.CompositeScoreContribution(nil), result.Contributions...),
			result.Contributions[0])
		_, err := types.DeriveAssuranceVector(&dup, types.DerivationInput{Account: testVectorAccount})
		require.Error(t, err, "a duplicate component must be rejected rather than silently collapsed")
	})
}

// ============================================================================
// Absence of a factor is NOT zero
// ============================================================================

func TestUnmeasuredFactorCarriesNoAssurance(t *testing.T) {
	// A result with only document evidence: every other component is absent.
	result := types.NewCompositeScoreResult(500, time.Unix(1_700_000_000, 0).UTC())
	result.AddContribution(types.CompositeScoreContribution{
		ComponentName:   types.ComponentDocumentAuthenticity,
		RawScore:        9500,
		Weight:          types.WeightDocumentAuthenticity,
		WeightedScore:   2375,
		PassedThreshold: true,
	})

	vector, err := types.DeriveAssuranceVector(result, types.DerivationInput{
		Account:    testVectorAccount,
		VerifiedAt: time.Unix(1_700_000_000, 0).UTC(),
		// Device evidence deliberately absent.
	})
	require.NoError(t, err)
	require.NoError(t, vector.Validate())

	// Document is measured.
	score, ok := vector.ScoreBps(types.AssuranceFactorDocument)
	require.True(t, ok)
	require.Equal(t, uint32(9500), score)

	// Everything else is UNMEASURED, and reading it must report "no claim"
	// rather than a score of zero.
	for _, factor := range []types.AssuranceFactor{
		types.AssuranceFactorBiometric,
		types.AssuranceFactorDataConsistency,
		types.AssuranceFactorDevice,
		types.AssuranceFactorLiveness,
		types.AssuranceFactorHistory,
		types.AssuranceFactorRisk,
	} {
		entry, found := vector.Lookup(factor)
		require.True(t, found, "factor %s must still be present in the vector", factor)
		require.False(t, entry.Measured, "factor %s must be unmeasured", factor)
		require.Zero(t, entry.ScoreBps, "an unmeasured factor must not carry a score")

		score, ok := vector.ScoreBps(factor)
		require.False(t, ok, "an unmeasured factor must read as no assurance claim")
		require.Zero(t, score)
	}

	// Only the document factor counts toward a distinct-factor count.
	require.Equal(t, []types.AssuranceFactor{types.AssuranceFactorDocument}, vector.MeasuredFactors())
}

func TestUnmeasuredFactorWithNonZeroScoreIsRejected(t *testing.T) {
	_, vector := deriveFixtureVector(t)

	// Hand-build a vector where an unmeasured factor claims a real score. This is
	// the state a relying party must never be able to read.
	bad := *vector
	bad.Factors = append([]types.AssuranceFactorEntry(nil), vector.Factors...)
	for i := range bad.Factors {
		if bad.Factors[i].Factor == types.AssuranceFactorDevice {
			bad.Factors[i].Measured = false
			bad.Factors[i].ScoreBps = 9999
			bad.Factors[i].WeightedScoreBps = 499
		}
	}
	bad.Commitment = bad.ComputeCommitment() // even a self-consistent hash must not help

	require.Error(t, bad.Validate(),
		"an unmeasured factor carrying a non-zero score must be rejected even with a matching commitment")
}

func TestValidateRejectsNonCanonicalFactorOrder(t *testing.T) {
	_, vector := deriveFixtureVector(t)

	bad := *vector
	bad.Factors = append([]types.AssuranceFactorEntry(nil), vector.Factors...)
	bad.Factors[0], bad.Factors[1] = bad.Factors[1], bad.Factors[0]
	bad.Commitment = bad.ComputeCommitment()

	require.Error(t, bad.Validate(), "factors out of canonical order must be rejected")
}

func TestValidateRejectsUnknownFactor(t *testing.T) {
	_, vector := deriveFixtureVector(t)

	bad := *vector
	bad.Factors = append([]types.AssuranceFactorEntry(nil), vector.Factors...)
	bad.Factors[0].Factor = types.AssuranceFactor("favourite_colour")
	bad.Commitment = bad.ComputeCommitment()

	require.Error(t, bad.Validate(), "an unknown factor name must be rejected")
}

func TestValidateRejectsOutOfRangeScore(t *testing.T) {
	_, vector := deriveFixtureVector(t)

	bad := *vector
	bad.Factors = append([]types.AssuranceFactorEntry(nil), vector.Factors...)
	bad.Factors[0].ScoreBps = uint32(types.MaxBasisPoints) + 1
	bad.Commitment = bad.ComputeCommitment()

	require.Error(t, bad.Validate(), "a score above basis points must be rejected")
}

// ============================================================================
// Recency / staleness
// ============================================================================

func TestRecencyUsesConsensusTimeAndFlagsStale(t *testing.T) {
	_, vector := deriveFixtureVector(t)

	verifiedAt := time.Unix(1_700_000_000, 0).UTC()
	const maxAge = int64(30 * 24 * 60 * 60)

	t.Run("fresh", func(t *testing.T) {
		recency := vector.Recency(verifiedAt.Add(time.Duration(10*24)*time.Hour), 700, maxAge)
		require.False(t, recency.Stale)
		require.Equal(t, int64(10*24*60*60), recency.AgeSeconds)
		require.Equal(t, maxAge, recency.MaxAgeSeconds)
	})

	t.Run("exactly at the boundary is not stale", func(t *testing.T) {
		recency := vector.Recency(verifiedAt.Add(time.Duration(maxAge)*time.Second), 700, maxAge)
		require.False(t, recency.Stale, "staleness is strictly greater than the window")
		require.Equal(t, maxAge, recency.AgeSeconds)
	})

	t.Run("stale", func(t *testing.T) {
		recency := vector.Recency(verifiedAt.Add(time.Duration(maxAge+1)*time.Second), 700, maxAge)
		require.True(t, recency.Stale)
	})

	t.Run("a vector dated in the future is not treated as fresh", func(t *testing.T) {
		recency := vector.Recency(verifiedAt.Add(-time.Hour), 700, maxAge)
		require.False(t, recency.Stale)
		require.Equal(t, int64(0), recency.AgeSeconds, "a negative age must clamp to zero, never read as fresh evidence")
	})

	t.Run("a non-positive window falls back to the default", func(t *testing.T) {
		recency := vector.Recency(verifiedAt, 700, 0)
		require.Equal(t, types.DefaultAssuranceVectorMaxAgeSeconds, recency.MaxAgeSeconds)
	})
}

func TestIsSupersededAndNextEpoch(t *testing.T) {
	_, vector := deriveFixtureVector(t)
	require.Equal(t, uint64(0), vector.Epoch)

	require.False(t, vector.IsSuperseded(0), "a vector is not superseded by itself")
	require.False(t, vector.IsSuperseded(0))
	require.True(t, vector.IsSuperseded(1), "a later epoch supersedes this vector")
	require.False(t, vector.IsSuperseded(0))

	require.Equal(t, uint64(1), types.NextEpoch(0),
		"a re-verification must advance past the first vector's epoch 0")
	require.Equal(t, uint64(2), types.NextEpoch(1))
	require.Equal(t, uint64(3), types.NextEpoch(2))
}

func TestOverallBpsFromScoreClampsAndIsIntegerExact(t *testing.T) {
	require.Equal(t, uint32(0), types.OverallBpsFromScore(0))
	require.Equal(t, uint32(5000), types.OverallBpsFromScore(50))
	require.Equal(t, uint32(10000), types.OverallBpsFromScore(100))
	require.Equal(t, uint32(10000), types.OverallBpsFromScore(1000), "a score above the max clamps")
}

// ============================================================================
// Privacy: no personal attribute may appear in the vector
// ============================================================================

// personalAttributeSubstrings are substrings that must never appear in an
// assurance vector's field names or JSON tags. An assurance vector carries scores
// and provenance references only.
//
// The list covers the categories named in the module's privacy rules (date of
// birth, document number, nationality, raw biometrics) plus the obvious adjacent
// PII a future field might introduce.
var personalAttributeSubstrings = []string{
	"dob",
	"date_of_birth",
	"birth",
	"birthdate",
	"passport",
	"nationality",
	"citizen",
	"ssn",
	"tax_id",
	"document_number",
	"doc_number",
	"id_number",
	"licence",
	"license_number",
	"first_name",
	"last_name",
	"full_name",
	"surname",
	"given_name",
	"address_line",
	"street",
	"postal",
	"postcode",
	"zip",
	"email",
	"phone",
	"biometric_template",
	"face_template",
	"fingerprint_template",
	"voiceprint",
	"iris",
	"raw_biometric",
	"embedding",
	"gender",
	"ethnicity",
	"religion",
	"height_cm",
	"weight_kg",
	"age_years",
}

// TestAssuranceVectorHasNoPersonalAttributes fails if any field of the assurance
// vector looks like it carries a personal attribute.
//
// This is a structural test over the Go type, so it breaks the build the moment
// someone adds a PII-bearing field, whatever the JSON tags say.
func TestAssuranceVectorHasNoPersonalAttributes(t *testing.T) {
	t.Run("AssuranceVector fields", func(t *testing.T) {
		assertNoPersonalAttributes(t, reflect.TypeOf(types.AssuranceVector{}))
	})

	t.Run("AssuranceFactorEntry fields", func(t *testing.T) {
		assertNoPersonalAttributes(t, reflect.TypeOf(types.AssuranceFactorEntry{}))
	})

	t.Run("AssuranceVectorRecency fields", func(t *testing.T) {
		assertNoPersonalAttributes(t, reflect.TypeOf(types.AssuranceVectorRecency{}))
	})

	t.Run("DerivationInput fields", func(t *testing.T) {
		assertNoPersonalAttributes(t, reflect.TypeOf(types.DerivationInput{}))
	})

	t.Run("DeviceEvidence fields", func(t *testing.T) {
		assertNoPersonalAttributes(t, reflect.TypeOf(types.DeviceEvidence{}))
	})

	t.Run("factor names", func(t *testing.T) {
		for _, factor := range types.AssuranceFactors {
			name := string(factor)
			for _, banned := range personalAttributeSubstrings {
				require.NotContains(t, name, banned,
					"factor name %q looks like it carries a personal attribute (%q)", name, banned)
			}
		}
	})

	// A derived vector's serialised form must not leak the inputs it came from
	// beyond the scores themselves: the DOB consistency sub-score is a score, but
	// a raw value would not be. Assert the vector's own rendering carries only
	// factor names, scores and provenance.
	t.Run("vector rendering carries no attributes", func(t *testing.T) {
		_, vector := deriveFixtureVector(t)
		rendered := vector.String()
		for _, banned := range personalAttributeSubstrings {
			require.NotContains(t, rendered, banned,
				"vector rendering must not mention a personal attribute (%q)", banned)
		}
	})
}

// assertNoPersonalAttributes walks a struct type and fails on any field whose
// name or JSON tag looks like a personal attribute.
func assertNoPersonalAttributes(t *testing.T, typ reflect.Type) {
	t.Helper()

	require.Equal(t, reflect.Struct, typ.Kind(), "expected a struct, got %s", typ.Kind())

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		jsonTag := field.Tag.Get("json")
		tagName := strings.Split(jsonTag, ",")[0]

		for _, candidate := range []string{field.Name, tagName} {
			if candidate == "" {
				continue
			}
			lowered := strings.ToLower(candidate)
			for _, banned := range personalAttributeSubstrings {
				require.NotContains(t, lowered, banned,
					"%s.%s (json %q) looks like it carries a personal attribute (%q)",
					typ.Name(), field.Name, jsonTag, banned)
			}
		}
	}
}

// ============================================================================
// Factor direction semantics
// ============================================================================

func TestIsHigherBetterDocumentsEachFactorsDirection(t *testing.T) {
	for _, factor := range types.AssuranceFactors {
		higherBetter, known := types.IsHigherBetter(factor)
		require.True(t, known, "factor %s must document its direction", factor)

		if factor == types.AssuranceFactorRisk {
			require.False(t, higherBetter, "risk is a lower-is-better reading and must say so")
		} else {
			require.True(t, higherBetter, "factor %s is higher-is-better", factor)
		}
	}

	_, known := types.IsHigherBetter(types.AssuranceFactor("unknown"))
	require.False(t, known, "an unknown factor must not claim a direction")
}

func TestIsValidAssuranceFactor(t *testing.T) {
	for _, factor := range types.AssuranceFactors {
		require.True(t, types.IsValidAssuranceFactor(factor))
	}
	require.False(t, types.IsValidAssuranceFactor(types.AssuranceFactor("nope")))
}

func TestAssuranceFactorsAreCanonicallyOrderedAndUnique(t *testing.T) {
	seen := make(map[types.AssuranceFactor]struct{}, len(types.AssuranceFactors))
	for i, factor := range types.AssuranceFactors {
		_, dup := seen[factor]
		require.False(t, dup, "factor %s is listed twice", factor)
		seen[factor] = struct{}{}

		if i > 0 {
			require.Less(t, types.AssuranceFactors[i-1], factor,
				"factors must be in ascending lexicographic order for a stable commitment")
		}
	}
	require.Len(t, seen, 7, "the vector exposes one factor per weighted component, plus device")
}

func TestSortAssuranceFactorEntriesIsDeterministic(t *testing.T) {
	entries := []types.AssuranceFactorEntry{
		{Factor: types.AssuranceFactorRisk},
		{Factor: types.AssuranceFactorBiometric},
		{Factor: types.AssuranceFactorDataConsistency},
		{Factor: types.AssuranceFactorDevice},
	}

	types.SortAssuranceFactorEntries(entries)
	require.Equal(t, types.AssuranceFactors[0], entries[0].Factor)
	require.Equal(t, types.AssuranceFactorBiometric, entries[0].Factor)
	require.Equal(t, types.AssuranceFactorDataConsistency, entries[1].Factor)
	require.Equal(t, types.AssuranceFactorDevice, entries[2].Factor)
	require.Equal(t, types.AssuranceFactorRisk, entries[3].Factor)
}
