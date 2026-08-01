// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package uniquenessops

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
)

const (
	PPMScale       uint64 = 1_000_000
	MaxCount       uint64 = 1_000_000_000_000
	MinimumCell    uint64 = 20
	MaximumCohorts        = 64
	WindowSeconds  uint64 = 86_400

	Domain         = "virtengine/uniqueness-operations-evidence/v1"
	Version uint32 = 1
)

type EvidenceClass string
type Certification string
type EvidenceStatus string
type SourceName string
type SourceStatus string
type AppealBacklogBand string
type SubgroupMetricKind string

const (
	EvidenceClassSyntheticFixture EvidenceClass = "synthetic_fixture"
	CertificationNotCertified     Certification = "not_certified"

	StatusFixtureComplete   EvidenceStatus = "fixture_complete"
	StatusFixtureSuppressed EvidenceStatus = "fixture_suppressed"
	StatusSourceUnavailable EvidenceStatus = "source_unavailable"

	SourceAppeal          SourceName = "appeal"
	SourceAdjudication    SourceName = "adjudication"
	SourceEnrollment      SourceName = "enrollment"
	SourceSearch          SourceName = "search"
	SourceSubgroup        SourceName = "subgroup"
	SourceThresholdQuorum SourceName = "threshold_quorum"

	SourceStatusFixture     SourceStatus = "fixture"
	SourceStatusUnavailable SourceStatus = "unavailable"
	SourceStatusStale       SourceStatus = "stale"
	SourceStatusInvalid     SourceStatus = "invalid"

	AppealBacklogNone            AppealBacklogBand = "none"
	AppealBacklogSuppressed1To19 AppealBacklogBand = "suppressed_1_19"
	AppealBacklog20To99          AppealBacklogBand = "20_99"
	AppealBacklog100To999        AppealBacklogBand = "100_999"
	AppealBacklog1000Plus        AppealBacklogBand = "1000_plus"

	SubgroupMetricCandidateRate      SubgroupMetricKind = "candidate_rate"
	SubgroupMetricConfirmedMatchRate SubgroupMetricKind = "confirmed_match_rate"
	SubgroupMetricFalseMatchRate     SubgroupMetricKind = "false_match_rate"
	SubgroupMetricFalseNonMatchRate  SubgroupMetricKind = "false_non_match_rate"
	SubgroupMetricConflictRate       SubgroupMetricKind = "conflict_rate"
)

var requiredSourceNames = [...]SourceName{
	SourceAppeal,
	SourceAdjudication,
	SourceEnrollment,
	SourceSearch,
	SourceSubgroup,
	SourceThresholdQuorum,
}

type Metadata struct {
	WindowStartUnixSeconds uint64 `json:"window_start_unix_seconds"`
	WindowEndUnixSeconds   uint64 `json:"window_end_unix_seconds"`
	PolicyDigest           string `json:"policy_digest"`
	ThresholdProfileDigest string `json:"threshold_profile_digest"`
	FixtureInputDigest     string `json:"fixture_input_digest"`
}

type Source struct {
	Name   SourceName   `json:"name"`
	Status SourceStatus `json:"status"`
}

type Counts struct {
	CompletedSearches         uint64 `json:"completed_searches"`
	CandidateSearches         uint64 `json:"candidate_searches"`
	ReviewRequired            uint64 `json:"review_required"`
	ReviewsCompleted          uint64 `json:"reviews_completed"`
	ConfirmedMatches          uint64 `json:"confirmed_matches"`
	FalseMatches              uint64 `json:"false_matches"`
	FalseMatchEvaluations     uint64 `json:"false_match_evaluations"`
	FalseNonMatches           uint64 `json:"false_non_matches"`
	FalseNonMatchEvaluations  uint64 `json:"false_non_match_evaluations"`
	EnrollmentAttempts        uint64 `json:"enrollment_attempts"`
	AtomicEnrollmentConflicts uint64 `json:"atomic_enrollment_conflicts"`
	ThresholdObservations     uint64 `json:"threshold_observations"`
	ThresholdQuorumMet        uint64 `json:"threshold_quorum_met"`
}

type Ratio struct {
	Published   bool   `json:"published"`
	Numerator   uint64 `json:"numerator"`
	Denominator uint64 `json:"denominator"`
	ValuePPM    uint64 `json:"value_ppm"`
}

type Metrics struct {
	CandidateSearchRate          Ratio `json:"candidate_search_rate"`
	ReviewCompletionRate         Ratio `json:"review_completion_rate"`
	ConfirmedMatchRate           Ratio `json:"confirmed_match_rate"`
	FalseMatchRate               Ratio `json:"false_match_rate"`
	FalseNonMatchRate            Ratio `json:"false_non_match_rate"`
	AtomicEnrollmentConflictRate Ratio `json:"atomic_enrollment_conflict_rate"`
	ThresholdAvailabilityRate    Ratio `json:"threshold_availability_rate"`
}

type AppealBacklog struct {
	Band AppealBacklogBand `json:"band"`
}

type SubgroupMetric struct {
	Kind        SubgroupMetricKind `json:"kind"`
	Numerator   uint64             `json:"numerator"`
	Denominator uint64             `json:"denominator"`
	ValuePPM    uint64             `json:"value_ppm"`
}

type SubgroupCohort struct {
	CohortToken       [32]byte         `json:"cohort_token"`
	IntersectionDepth uint32           `json:"intersection_depth"`
	SampleCount       uint64           `json:"sample_count"`
	Metrics           []SubgroupMetric `json:"metrics"`
}

type SubgroupEvidence struct {
	TaxonomyDigest          string           `json:"taxonomy_digest"`
	MinimumSampleCount      uint64           `json:"minimum_sample_count"`
	Cohorts                 []SubgroupCohort `json:"cohorts"`
	SuppressedSubgroupCount uint64           `json:"suppressed_subgroup_count"`
}

type FixtureInput struct {
	Metadata           Metadata
	Sources            []Source
	Counts             Counts
	AppealBacklogCount uint64
	Subgroups          SubgroupEvidence
}

type Evidence struct {
	Domain        string            `json:"domain"`
	Version       uint32            `json:"version"`
	Class         EvidenceClass     `json:"class"`
	Certification Certification     `json:"certification"`
	Status        EvidenceStatus    `json:"status"`
	Metadata      Metadata          `json:"metadata"`
	Sources       []Source          `json:"sources"`
	Metrics       *Metrics          `json:"metrics"`
	AppealBacklog *AppealBacklog    `json:"appeal_backlog"`
	Subgroups     *SubgroupEvidence `json:"subgroups"`
}

func BuildFixtureEvidence(input FixtureInput) (*Evidence, error) {
	metrics, err := ComputeMetrics(input.Counts)
	if err != nil {
		return nil, err
	}
	band, err := appealBand(input.AppealBacklogCount)
	if err != nil {
		return nil, err
	}
	status := StatusFixtureComplete
	if metrics.suppressed() || input.Subgroups.SuppressedSubgroupCount > 0 || band == AppealBacklogSuppressed1To19 {
		status = StatusFixtureSuppressed
	}
	envelope := &Evidence{
		Domain: Domain, Version: Version, Class: EvidenceClassSyntheticFixture,
		Certification: CertificationNotCertified, Status: status,
		Metadata: input.Metadata, Sources: input.Sources,
		Metrics: &metrics, AppealBacklog: &AppealBacklog{Band: band}, Subgroups: &input.Subgroups,
	}
	if err := envelope.Validate(); err != nil {
		return nil, err
	}
	return envelope, nil
}

func BuildUnavailableEvidence(metadata Metadata, sources []Source) (*Evidence, error) {
	envelope := &Evidence{
		Domain: Domain, Version: Version, Class: EvidenceClassSyntheticFixture,
		Certification: CertificationNotCertified, Status: StatusSourceUnavailable,
		Metadata: metadata, Sources: sources,
	}
	if err := envelope.Validate(); err != nil {
		return nil, err
	}
	return envelope, nil
}

func ComputeMetrics(counts Counts) (Metrics, error) {
	if err := validateCounts(counts); err != nil {
		return Metrics{}, err
	}
	return Metrics{
		CandidateSearchRate:          publishRatio(counts.CandidateSearches, counts.CompletedSearches),
		ReviewCompletionRate:         publishRatio(counts.ReviewsCompleted, counts.ReviewRequired),
		ConfirmedMatchRate:           publishRatio(counts.ConfirmedMatches, counts.ReviewsCompleted),
		FalseMatchRate:               publishRatio(counts.FalseMatches, counts.FalseMatchEvaluations),
		FalseNonMatchRate:            publishRatio(counts.FalseNonMatches, counts.FalseNonMatchEvaluations),
		AtomicEnrollmentConflictRate: publishRatio(counts.AtomicEnrollmentConflicts, counts.EnrollmentAttempts),
		ThresholdAvailabilityRate:    publishRatio(counts.ThresholdQuorumMet, counts.ThresholdObservations),
	}, nil
}

func (envelope *Evidence) Validate() error {
	if envelope == nil {
		return errors.New("uniqueness operations evidence is required")
	}
	if envelope.Domain != Domain || envelope.Version != Version {
		return errors.New("unsupported uniqueness operations evidence domain or version")
	}
	if envelope.Class != EvidenceClassSyntheticFixture {
		return errors.New("unsupported evidence class")
	}
	if envelope.Certification != CertificationNotCertified {
		return errors.New("unsupported certification state")
	}
	if err := envelope.Metadata.validate(); err != nil {
		return err
	}
	allFixture, err := validateSources(envelope.Sources)
	if err != nil {
		return err
	}
	if !allFixture {
		if envelope.Status != StatusSourceUnavailable {
			return errors.New("nonfixture sources require unavailable status")
		}
		if envelope.Metrics != nil || envelope.AppealBacklog != nil || envelope.Subgroups != nil {
			return errors.New("unavailable evidence must omit operational values")
		}
		return nil
	}
	if envelope.Metrics == nil || envelope.AppealBacklog == nil || envelope.Subgroups == nil {
		return errors.New("fixture evidence requires all aggregate sections")
	}
	if err := envelope.Metrics.validate(); err != nil {
		return errors.New("aggregate metrics are invalid")
	}
	if !validAppealBand(envelope.AppealBacklog.Band) {
		return errors.New("appeal backlog band is invalid")
	}
	if err := envelope.Subgroups.validate(); err != nil {
		return err
	}
	expectedStatus := StatusFixtureComplete
	if envelope.Metrics.suppressed() || envelope.Subgroups.SuppressedSubgroupCount > 0 || envelope.AppealBacklog.Band == AppealBacklogSuppressed1To19 {
		expectedStatus = StatusFixtureSuppressed
	}
	if envelope.Status != expectedStatus {
		return errors.New("fixture evidence status is inconsistent")
	}
	return nil
}

func (metrics Metrics) validate() error {
	for _, ratio := range []Ratio{
		metrics.CandidateSearchRate, metrics.ReviewCompletionRate, metrics.ConfirmedMatchRate,
		metrics.FalseMatchRate, metrics.FalseNonMatchRate, metrics.AtomicEnrollmentConflictRate,
		metrics.ThresholdAvailabilityRate,
	} {
		if !ratio.Published {
			if ratio.Numerator != 0 || ratio.Denominator != 0 || ratio.ValuePPM != 0 {
				return errors.New("suppressed aggregate ratio must omit numeric values")
			}
			continue
		}
		if ratio.Numerator > MaxCount || ratio.Denominator < MinimumCell || ratio.Denominator > MaxCount || ratio.Numerator > ratio.Denominator {
			return errors.New("published aggregate ratio counts are invalid")
		}
		if ratio.ValuePPM != ratio.Numerator*PPMScale/ratio.Denominator {
			return errors.New("published aggregate ratio value is invalid")
		}
	}
	return nil
}

func (envelope *Evidence) CanonicalBytes() ([]byte, error) {
	if err := envelope.Validate(); err != nil {
		return nil, err
	}
	payload, err := json.Marshal(envelope)
	if err != nil {
		return nil, errors.New("uniqueness operations evidence serialization failed")
	}
	result := make([]byte, 0, len(Domain)+1+len(payload))
	result = append(result, Domain...)
	result = append(result, 0)
	result = append(result, payload...)
	return result, nil
}

func (envelope *Evidence) Digest() ([32]byte, error) {
	canonical, err := envelope.CanonicalBytes()
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(canonical), nil
}

func (metadata Metadata) validate() error {
	if metadata.WindowStartUnixSeconds%WindowSeconds != 0 ||
		metadata.WindowStartUnixSeconds > math.MaxUint64-WindowSeconds ||
		metadata.WindowEndUnixSeconds != metadata.WindowStartUnixSeconds+WindowSeconds {
		return errors.New("evidence window must be one UTC-aligned day")
	}
	for _, digest := range []string{metadata.PolicyDigest, metadata.ThresholdProfileDigest, metadata.FixtureInputDigest} {
		if !isCanonicalDigest(digest) {
			return errors.New("evidence digest is malformed")
		}
	}
	return nil
}

func validateSources(sources []Source) (bool, error) {
	if len(sources) != len(requiredSourceNames) {
		return false, errors.New("all required sources must be present")
	}
	allFixture := true
	for index, source := range sources {
		if source.Name != requiredSourceNames[index] {
			return false, errors.New("required sources are not in canonical order")
		}
		switch source.Status {
		case SourceStatusFixture:
		case SourceStatusUnavailable, SourceStatusStale, SourceStatusInvalid:
			allFixture = false
		default:
			return false, errors.New("source status is invalid")
		}
	}
	return allFixture, nil
}

func validateCounts(counts Counts) error {
	values := [...]uint64{
		counts.CompletedSearches, counts.CandidateSearches, counts.ReviewRequired,
		counts.ReviewsCompleted, counts.ConfirmedMatches, counts.FalseMatches,
		counts.FalseMatchEvaluations, counts.FalseNonMatches, counts.FalseNonMatchEvaluations,
		counts.EnrollmentAttempts, counts.AtomicEnrollmentConflicts,
		counts.ThresholdObservations, counts.ThresholdQuorumMet,
	}
	for _, value := range values {
		if value > MaxCount {
			return errors.New("aggregate count exceeds limit")
		}
	}
	if counts.ConfirmedMatches > counts.ReviewsCompleted ||
		counts.ReviewsCompleted > counts.ReviewRequired ||
		counts.ReviewRequired > counts.CandidateSearches ||
		counts.CandidateSearches > counts.CompletedSearches {
		return errors.New("search and review counts contradict their logical bounds")
	}
	if counts.FalseMatches > counts.FalseMatchEvaluations ||
		counts.FalseNonMatches > counts.FalseNonMatchEvaluations {
		return errors.New("false outcome count exceeds governed evaluations")
	}
	if counts.AtomicEnrollmentConflicts > counts.EnrollmentAttempts {
		return errors.New("enrollment conflict count exceeds attempts")
	}
	if counts.ThresholdQuorumMet > counts.ThresholdObservations {
		return errors.New("threshold quorum count exceeds observations")
	}
	return nil
}

func publishRatio(numerator, denominator uint64) Ratio {
	if denominator < MinimumCell {
		return Ratio{}
	}
	return Ratio{Published: true, Numerator: numerator, Denominator: denominator, ValuePPM: numerator * PPMScale / denominator}
}

func (metrics Metrics) suppressed() bool {
	return !metrics.CandidateSearchRate.Published || !metrics.ReviewCompletionRate.Published ||
		!metrics.ConfirmedMatchRate.Published || !metrics.FalseMatchRate.Published ||
		!metrics.FalseNonMatchRate.Published || !metrics.AtomicEnrollmentConflictRate.Published ||
		!metrics.ThresholdAvailabilityRate.Published
}

func appealBand(count uint64) (AppealBacklogBand, error) {
	if count > MaxCount {
		return "", errors.New("appeal backlog count exceeds limit")
	}
	switch {
	case count == 0:
		return AppealBacklogNone, nil
	case count < MinimumCell:
		return AppealBacklogSuppressed1To19, nil
	case count < 100:
		return AppealBacklog20To99, nil
	case count < 1_000:
		return AppealBacklog100To999, nil
	default:
		return AppealBacklog1000Plus, nil
	}
}

func validAppealBand(band AppealBacklogBand) bool {
	switch band {
	case AppealBacklogNone, AppealBacklogSuppressed1To19, AppealBacklog20To99, AppealBacklog100To999, AppealBacklog1000Plus:
		return true
	default:
		return false
	}
}

func (subgroups *SubgroupEvidence) validate() error {
	if subgroups == nil || !isCanonicalDigest(subgroups.TaxonomyDigest) {
		return errors.New("subgroup taxonomy digest is malformed")
	}
	if subgroups.MinimumSampleCount < MinimumCell || subgroups.MinimumSampleCount > MaxCount {
		return errors.New("subgroup minimum sample violates privacy policy")
	}
	if len(subgroups.Cohorts) > MaximumCohorts {
		return errors.New("subgroup cohort cardinality exceeds limit")
	}
	if len(subgroups.Cohorts) == 0 && subgroups.SuppressedSubgroupCount == 0 {
		return errors.New("subgroup evidence requires a cohort or suppressed count")
	}
	if subgroups.SuppressedSubgroupCount > MaxCount {
		return errors.New("suppressed subgroup count exceeds limit")
	}
	for cohortIndex, cohort := range subgroups.Cohorts {
		if cohortIndex > 0 && bytes.Compare(subgroups.Cohorts[cohortIndex-1].CohortToken[:], cohort.CohortToken[:]) >= 0 {
			return errors.New("subgroup cohort tokens must be strictly ordered")
		}
		if cohort.CohortToken == ([32]byte{}) {
			return errors.New("subgroup cohort token is required")
		}
		if cohort.IntersectionDepth < 1 || cohort.IntersectionDepth > 3 {
			return errors.New("subgroup intersection depth is invalid")
		}
		if cohort.SampleCount < subgroups.MinimumSampleCount || cohort.SampleCount > MaxCount {
			return errors.New("subgroup cohort sample violates privacy policy")
		}
		if len(cohort.Metrics) == 0 || len(cohort.Metrics) > 16 {
			return errors.New("subgroup metric cardinality is invalid")
		}
		for metricIndex, metric := range cohort.Metrics {
			if !validSubgroupMetricKind(metric.Kind) {
				return errors.New("subgroup metric kind is invalid")
			}
			if metricIndex > 0 && cohort.Metrics[metricIndex-1].Kind >= metric.Kind {
				return errors.New("subgroup metric kinds must be strictly ordered")
			}
			if metric.Numerator > MaxCount || metric.Denominator > MaxCount ||
				metric.Numerator > metric.Denominator || metric.Denominator < subgroups.MinimumSampleCount ||
				metric.Denominator > cohort.SampleCount {
				return errors.New("subgroup metric counts are invalid")
			}
			if metric.ValuePPM != metric.Numerator*PPMScale/metric.Denominator {
				return errors.New("subgroup metric value is invalid")
			}
		}
	}
	return nil
}

func validSubgroupMetricKind(kind SubgroupMetricKind) bool {
	switch kind {
	case SubgroupMetricCandidateRate, SubgroupMetricConfirmedMatchRate, SubgroupMetricFalseMatchRate,
		SubgroupMetricFalseNonMatchRate, SubgroupMetricConflictRate:
		return true
	default:
		return false
	}
}

func isCanonicalDigest(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && hex.EncodeToString(decoded) == value
}
