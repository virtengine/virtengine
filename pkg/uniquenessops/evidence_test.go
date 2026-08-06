// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package uniquenessops

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComputeMetricsExactFormulaTruncationAndSuppression(t *testing.T) {
	counts := validCounts()
	counts.CandidateSearches = 20
	counts.CompletedSearches = 30
	counts.ReviewRequired = 20
	counts.ReviewsCompleted = 20
	metrics, err := ComputeMetrics(counts)
	require.NoError(t, err)
	require.Equal(t, Ratio{Published: true, Numerator: 20, Denominator: 30, ValuePPM: 666_666}, metrics.CandidateSearchRate)

	counts.FalseMatchEvaluations = MinimumCell - 1
	counts.FalseMatches = 7
	metrics, err = ComputeMetrics(counts)
	require.NoError(t, err)
	require.Equal(t, Ratio{}, metrics.FalseMatchRate)
}

func TestCountsLimitsAndContradictions(t *testing.T) {
	tests := []func(*Counts){
		func(counts *Counts) { counts.CompletedSearches = MaxCount + 1 },
		func(counts *Counts) { counts.CandidateSearches = counts.CompletedSearches + 1 },
		func(counts *Counts) { counts.ReviewsCompleted = counts.ReviewRequired + 1 },
		func(counts *Counts) { counts.ConfirmedMatches = counts.ReviewsCompleted + 1 },
		func(counts *Counts) { counts.FalseMatches = counts.FalseMatchEvaluations + 1 },
		func(counts *Counts) { counts.FalseNonMatches = counts.FalseNonMatchEvaluations + 1 },
		func(counts *Counts) { counts.AtomicEnrollmentConflicts = counts.EnrollmentAttempts + 1 },
		func(counts *Counts) { counts.ThresholdQuorumMet = counts.ThresholdObservations + 1 },
	}
	for _, mutate := range tests {
		counts := validCounts()
		mutate(&counts)
		_, err := ComputeMetrics(counts)
		require.Error(t, err)
	}
}

func TestUnavailableStaleAndInvalidSourcesOmitOperationalValues(t *testing.T) {
	for _, status := range []SourceStatus{SourceStatusUnavailable, SourceStatusStale, SourceStatusInvalid} {
		sources := fixtureSources()
		sources[3].Status = status
		evidence, err := BuildUnavailableEvidence(validMetadata(), sources)
		require.NoError(t, err)
		require.Equal(t, StatusSourceUnavailable, evidence.Status)
		require.Nil(t, evidence.Metrics)
		require.Nil(t, evidence.Subgroups)
		require.Nil(t, evidence.AppealBacklog)

		evidence.Metrics = &Metrics{}
		require.Error(t, evidence.Validate())
	}

	sources := fixtureSources()
	sources[0].Name = SourceName("unknown-private-source")
	_, err := BuildUnavailableEvidence(validMetadata(), sources)
	require.Error(t, err)
	require.NotContains(t, err.Error(), "unknown-private-source")
}

func TestSuppressionStatusAndAppealBands(t *testing.T) {
	input := validFixtureInput()
	evidence, err := BuildFixtureEvidence(input)
	require.NoError(t, err)
	require.Equal(t, StatusFixtureComplete, evidence.Status)

	input.Counts.ThresholdObservations = 19
	input.Counts.ThresholdQuorumMet = 19
	evidence, err = BuildFixtureEvidence(input)
	require.NoError(t, err)
	require.Equal(t, StatusFixtureSuppressed, evidence.Status)
	require.Equal(t, Ratio{}, evidence.Metrics.ThresholdAvailabilityRate)
	serialized, err := json.Marshal(evidence)
	require.NoError(t, err)
	require.NotContains(t, string(serialized), "threshold_observations")
	require.NotContains(t, string(serialized), `"denominator":19`)

	for count, expected := range map[uint64]AppealBacklogBand{
		0: AppealBacklogNone, 1: AppealBacklogSuppressed1To19, 19: AppealBacklogSuppressed1To19,
		20: AppealBacklog20To99, 99: AppealBacklog20To99, 100: AppealBacklog100To999,
		999: AppealBacklog100To999, 1_000: AppealBacklog1000Plus,
	} {
		input := validFixtureInput()
		input.AppealBacklogCount = count
		evidence, err := BuildFixtureEvidence(input)
		require.NoError(t, err)
		require.Equal(t, expected, evidence.AppealBacklog.Band)
		if expected == AppealBacklogSuppressed1To19 {
			require.Equal(t, StatusFixtureSuppressed, evidence.Status)
		}
		serialized, err := json.Marshal(evidence)
		require.NoError(t, err)
		require.NotContains(t, string(serialized), "appeal_backlog_count")
	}
}

func TestSuppressedAndPublishedRatioValidation(t *testing.T) {
	evidence, err := BuildFixtureEvidence(validFixtureInput())
	require.NoError(t, err)
	evidence.Metrics.FalseMatchRate = Ratio{Published: false, Numerator: 1, Denominator: 19, ValuePPM: 52_631}
	require.Error(t, evidence.Validate())

	evidence, err = BuildFixtureEvidence(validFixtureInput())
	require.NoError(t, err)
	evidence.Metrics.FalseMatchRate.ValuePPM++
	require.Error(t, evidence.Validate())
}

func TestSubgroupFloorOrderDuplicatesCardinalityAndToken(t *testing.T) {
	mutations := []func(*FixtureInput){
		func(input *FixtureInput) { input.Subgroups.MinimumSampleCount = MinimumCell - 1 },
		func(input *FixtureInput) { input.Subgroups.Cohorts[0].SampleCount = MinimumCell - 1 },
		func(input *FixtureInput) { input.Subgroups.Cohorts[0].CohortToken = [32]byte{} },
		func(input *FixtureInput) {
			input.Subgroups.Cohorts[0].Metrics = append(input.Subgroups.Cohorts[0].Metrics, input.Subgroups.Cohorts[0].Metrics[0])
		},
		func(input *FixtureInput) {
			second := input.Subgroups.Cohorts[0]
			second.CohortToken[0] = 2
			input.Subgroups.Cohorts = []SubgroupCohort{second, input.Subgroups.Cohorts[0]}
		},
		func(input *FixtureInput) {
			cohorts := make([]SubgroupCohort, MaximumCohorts+1)
			for index := range cohorts {
				cohorts[index] = input.Subgroups.Cohorts[0]
				cohorts[index].CohortToken[31] = byte(index + 1)
			}
			input.Subgroups.Cohorts = cohorts
		},
	}
	for _, mutate := range mutations {
		input := validFixtureInput()
		mutate(&input)
		_, err := BuildFixtureEvidence(input)
		require.Error(t, err)
	}

	input := validFixtureInput()
	input.Subgroups.Cohorts = nil
	input.Subgroups.SuppressedSubgroupCount = 3
	evidence, err := BuildFixtureEvidence(input)
	require.NoError(t, err)
	require.Equal(t, StatusFixtureSuppressed, evidence.Status)
}

func TestWindowIsExactUTCDay(t *testing.T) {
	for _, mutate := range []func(*Metadata){
		func(metadata *Metadata) { metadata.WindowStartUnixSeconds++ },
		func(metadata *Metadata) { metadata.WindowEndUnixSeconds-- },
		func(metadata *Metadata) { metadata.WindowEndUnixSeconds += WindowSeconds },
	} {
		input := validFixtureInput()
		mutate(&input.Metadata)
		_, err := BuildFixtureEvidence(input)
		require.Error(t, err)
	}
}

func TestCanonicalDigestTamperAndCertification(t *testing.T) {
	evidence, err := BuildFixtureEvidence(validFixtureInput())
	require.NoError(t, err)
	require.Equal(t, CertificationNotCertified, evidence.Certification)
	canonical, err := evidence.CanonicalBytes()
	require.NoError(t, err)
	require.True(t, bytes.HasPrefix(canonical, append([]byte(Domain), 0)))
	digest, err := evidence.Digest()
	require.NoError(t, err)
	require.Equal(t, "a1a021c97e5a3c6508e529785190766746eee3e739dbeb65fb3db3c35d929e5a", hex.EncodeToString(digest[:]))

	evidence.Metadata.WindowStartUnixSeconds += WindowSeconds
	evidence.Metadata.WindowEndUnixSeconds += WindowSeconds
	tampered, err := evidence.Digest()
	require.NoError(t, err)
	require.NotEqual(t, digest, tampered)

	evidence.Certification = Certification("certified")
	require.Error(t, evidence.Validate())
}

func TestNoPrivateCanaryInSerializationOrErrors(t *testing.T) {
	canary := "alice-biometric-nullifier-node-appeal-id"
	input := validFixtureInput()
	input.Metadata.PolicyDigest = canary
	_, err := BuildFixtureEvidence(input)
	require.Error(t, err)
	require.NotContains(t, err.Error(), canary)

	evidence, err := BuildFixtureEvidence(validFixtureInput())
	require.NoError(t, err)
	canonical, err := evidence.CanonicalBytes()
	require.NoError(t, err)
	require.NotContains(t, string(canonical), canary)
	for _, forbidden := range []string{"\"raw_id\"", "\"person_name\"", "\"biometric_data\"", "\"nullifier\"", "\"node_id\"", "\"appeal_id\""} {
		require.False(t, strings.Contains(string(canonical), forbidden))
	}
}

func FuzzEvidenceValidationPanicFree(f *testing.F) {
	f.Add(uint64(20), uint64(20), uint64(0))
	f.Fuzz(func(t *testing.T, denominator, numerator, appeal uint64) {
		input := validFixtureInput()
		input.Counts.FalseMatchEvaluations = denominator
		input.Counts.FalseMatches = numerator
		input.AppealBacklogCount = appeal
		_, _ = BuildFixtureEvidence(input)
	})
}

func validFixtureInput() FixtureInput {
	var token [32]byte
	token[0] = 1
	return FixtureInput{
		Metadata: validMetadata(), Sources: fixtureSources(), Counts: validCounts(), AppealBacklogCount: 20,
		Subgroups: SubgroupEvidence{
			TaxonomyDigest: testDigest("taxonomy"), MinimumSampleCount: MinimumCell,
			Cohorts: []SubgroupCohort{{
				CohortToken: token, IntersectionDepth: 1, SampleCount: 30,
				Metrics: []SubgroupMetric{
					{Kind: SubgroupMetricCandidateRate, Numerator: 20, Denominator: 30, ValuePPM: 666_666},
					{Kind: SubgroupMetricConflictRate, Numerator: 1, Denominator: 20, ValuePPM: 50_000},
				},
			}},
		},
	}
}

func validCounts() Counts {
	return Counts{
		CompletedSearches: 100, CandidateSearches: 80, ReviewRequired: 60,
		ReviewsCompleted: 50, ConfirmedMatches: 10,
		FalseMatches: 1, FalseMatchEvaluations: 25,
		FalseNonMatches: 2, FalseNonMatchEvaluations: 40,
		EnrollmentAttempts: 50, AtomicEnrollmentConflicts: 2,
		ThresholdObservations: 50, ThresholdQuorumMet: 45,
	}
}

func validMetadata() Metadata {
	return Metadata{
		WindowStartUnixSeconds: 1_785_628_800,
		WindowEndUnixSeconds:   1_785_715_200,
		PolicyDigest:           testDigest("policy"), ThresholdProfileDigest: testDigest("threshold"),
		FixtureInputDigest: testDigest("fixture"),
	}
}

func fixtureSources() []Source {
	return []Source{
		{Name: SourceAppeal, Status: SourceStatusFixture},
		{Name: SourceAdjudication, Status: SourceStatusFixture},
		{Name: SourceEnrollment, Status: SourceStatusFixture},
		{Name: SourceSearch, Status: SourceStatusFixture},
		{Name: SourceSubgroup, Status: SourceStatusFixture},
		{Name: SourceThresholdQuorum, Status: SourceStatusFixture},
	}
}

func testDigest(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}
