// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
)

const (
	ModelQualityPPMScale        uint64 = 1_000_000
	ModelQualityMaxCount        uint64 = 1_000_000_000_000
	ModelQualityEnvelopeDomain         = "virtengine/model-quality-evidence/v1"
	ModelQualityEnvelopeVersion uint32 = 1
	ModelQualitySubgroupFloor   uint64 = 20
)

type ModelQualityCorpusClass string
type ModelQualityEvidenceUse string
type ModelQualityCertificationState string
type ModelQualitySubgroupMetricKind string

const (
	ModelQualityCorpusSyntheticFixture   ModelQualityCorpusClass = "synthetic_fixture"
	ModelQualityCorpusDeclaredEvaluation ModelQualityCorpusClass = "declared_evaluation"

	ModelQualityUseComputationTest  ModelQualityEvidenceUse = "computation_test"
	ModelQualityUseEvaluationReport ModelQualityEvidenceUse = "evaluation_report"

	ModelQualityNotCertified ModelQualityCertificationState = "not_certified"

	SubgroupMetricOCRCER            ModelQualitySubgroupMetricKind = "ocr_cer"
	SubgroupMetricOCRWER            ModelQualitySubgroupMetricKind = "ocr_wer"
	SubgroupMetricDocumentAccuracy  ModelQualitySubgroupMetricKind = "document_accuracy"
	SubgroupMetricDocumentPrecision ModelQualitySubgroupMetricKind = "document_precision"
	SubgroupMetricDocumentRecall    ModelQualitySubgroupMetricKind = "document_recall"
	SubgroupMetricDocumentFPR       ModelQualitySubgroupMetricKind = "document_fpr"
	SubgroupMetricDocumentFNR       ModelQualitySubgroupMetricKind = "document_fnr"
	SubgroupMetricFaceFMR           ModelQualitySubgroupMetricKind = "face_fmr"
	SubgroupMetricFaceFNMR          ModelQualitySubgroupMetricKind = "face_fnmr"
	SubgroupMetricFaceEER           ModelQualitySubgroupMetricKind = "face_eer"
	SubgroupMetricPADAPCER          ModelQualitySubgroupMetricKind = "pad_apcer"
	SubgroupMetricPADBPCER          ModelQualitySubgroupMetricKind = "pad_bpcer"
	SubgroupMetricFraudECE          ModelQualitySubgroupMetricKind = "fraud_ece"
)

var allowedModelQualitySubgroupMetrics = map[ModelQualitySubgroupMetricKind]struct{}{
	SubgroupMetricOCRCER: {}, SubgroupMetricOCRWER: {},
	SubgroupMetricDocumentAccuracy: {}, SubgroupMetricDocumentPrecision: {}, SubgroupMetricDocumentRecall: {},
	SubgroupMetricDocumentFPR: {}, SubgroupMetricDocumentFNR: {}, SubgroupMetricFaceFMR: {},
	SubgroupMetricFaceFNMR: {}, SubgroupMetricFaceEER: {}, SubgroupMetricPADAPCER: {},
	SubgroupMetricPADBPCER: {}, SubgroupMetricFraudECE: {},
}

type OCRCounts struct {
	CharacterReferences    uint64 `json:"character_references"`
	CharacterSubstitutions uint64 `json:"character_substitutions"`
	CharacterDeletions     uint64 `json:"character_deletions"`
	CharacterInsertions    uint64 `json:"character_insertions"`
	WordReferences         uint64 `json:"word_references"`
	WordSubstitutions      uint64 `json:"word_substitutions"`
	WordDeletions          uint64 `json:"word_deletions"`
	WordInsertions         uint64 `json:"word_insertions"`
}

type OCRMetrics struct {
	CERPPM uint64 `json:"cer_ppm"`
	WERPPM uint64 `json:"wer_ppm"`
}

type OCREvidence struct {
	Counts  OCRCounts  `json:"counts"`
	Metrics OCRMetrics `json:"metrics"`
}

type DocumentConfusionCounts struct {
	TruePositive  uint64 `json:"true_positive"`
	TrueNegative  uint64 `json:"true_negative"`
	FalsePositive uint64 `json:"false_positive"`
	FalseNegative uint64 `json:"false_negative"`
}

type DocumentConfusionMetrics struct {
	AccuracyPPM  uint64 `json:"accuracy_ppm"`
	PrecisionPPM uint64 `json:"precision_ppm"`
	RecallPPM    uint64 `json:"recall_ppm"`
	FPRPPM       uint64 `json:"fpr_ppm"`
	FNRPPM       uint64 `json:"fnr_ppm"`
}

type DocumentConfusionEvidence struct {
	Counts  DocumentConfusionCounts  `json:"counts"`
	Metrics DocumentConfusionMetrics `json:"metrics"`
}

type FaceOperatingPointCounts struct {
	ThresholdPPM   uint64 `json:"threshold_ppm"`
	FalseAccepts   uint64 `json:"false_accepts"`
	ImpostorTrials uint64 `json:"impostor_trials"`
	FalseRejects   uint64 `json:"false_rejects"`
	GenuineTrials  uint64 `json:"genuine_trials"`
}

type FaceOperatingPoint struct {
	ThresholdPPM    uint64 `json:"threshold_ppm"`
	FMRPPM          uint64 `json:"fmr_ppm"`
	FNMRPPM         uint64 `json:"fnmr_ppm"`
	AverageErrorPPM uint64 `json:"average_error_ppm"`
}

type FaceEvidence struct {
	Counts          []FaceOperatingPointCounts `json:"counts"`
	OperatingPoints []FaceOperatingPoint       `json:"operating_points"`
	EERPoint        FaceOperatingPoint         `json:"eer_point"`
}

type PADCounts struct {
	AttackPresentations           uint64 `json:"attack_presentations"`
	AttackPresentationsAccepted   uint64 `json:"attack_presentations_accepted"`
	BonaFidePresentations         uint64 `json:"bona_fide_presentations"`
	BonaFidePresentationsRejected uint64 `json:"bona_fide_presentations_rejected"`
}

type PADMetrics struct {
	APCERPPM uint64 `json:"apcer_ppm"`
	BPCERPPM uint64 `json:"bpcer_ppm"`
}

type PADEvidence struct {
	Counts  PADCounts  `json:"counts"`
	Metrics PADMetrics `json:"metrics"`
}

type FraudCalibrationBin struct {
	LowerBoundPPM uint64 `json:"lower_bound_ppm"`
	UpperBoundPPM uint64 `json:"upper_bound_ppm"`
	Count         uint64 `json:"count"`
	Positives     uint64 `json:"positives"`
	PredictionSum uint64 `json:"prediction_sum"`
}

type FraudCalibrationBinMetrics struct {
	LowerBoundPPM       uint64 `json:"lower_bound_ppm"`
	UpperBoundPPM       uint64 `json:"upper_bound_ppm"`
	Count               uint64 `json:"count"`
	ObservedPositivePPM uint64 `json:"observed_positive_ppm"`
	MeanPredictionPPM   uint64 `json:"mean_prediction_ppm"`
	AbsoluteGapPPM      uint64 `json:"absolute_gap_ppm"`
}

type FraudCalibrationMetrics struct {
	Bins   []FraudCalibrationBinMetrics `json:"bins"`
	ECEPPM uint64                       `json:"ece_ppm"`
}

type FraudCalibrationEvidence struct {
	Bins    []FraudCalibrationBin   `json:"bins"`
	Metrics FraudCalibrationMetrics `json:"metrics"`
}

type ModelQualitySubgroupMetric struct {
	Kind        ModelQualitySubgroupMetricKind `json:"kind"`
	Numerator   uint64                         `json:"numerator"`
	Denominator uint64                         `json:"denominator"`
	ValuePPM    uint64                         `json:"value_ppm"`
}

type ModelQualitySubgroupCohort struct {
	CohortToken       [32]byte                     `json:"cohort_token"`
	IntersectionDepth uint32                       `json:"intersection_depth"`
	SampleCount       uint64                       `json:"sample_count"`
	Metrics           []ModelQualitySubgroupMetric `json:"metrics"`
}

type ModelQualitySubgroupEvidence struct {
	TaxonomyDigest          string                       `json:"taxonomy_digest"`
	MinimumSampleCount      uint64                       `json:"minimum_sample_count"`
	Cohorts                 []ModelQualitySubgroupCohort `json:"cohorts"`
	SuppressedSubgroupCount uint64                       `json:"suppressed_subgroup_count"`
}

type ModelQualityEvidenceEnvelopeV1 struct {
	Domain                 string                         `json:"domain"`
	Version                uint32                         `json:"version"`
	ModelDigest            string                         `json:"model_digest"`
	ProfileDigest          string                         `json:"profile_digest"`
	CorpusDigest           string                         `json:"corpus_digest"`
	SplitDigest            string                         `json:"split_digest"`
	LabelSchemaDigest      string                         `json:"label_schema_digest"`
	EvaluationConfigDigest string                         `json:"evaluation_config_digest"`
	ThresholdProfileDigest string                         `json:"threshold_profile_digest"`
	EvaluationUnixSeconds  uint64                         `json:"evaluation_unix_seconds"`
	CorpusClass            ModelQualityCorpusClass        `json:"corpus_class"`
	EvidenceUse            ModelQualityEvidenceUse        `json:"evidence_use"`
	CertificationState     ModelQualityCertificationState `json:"certification_state"`
	Representative         bool                           `json:"representative"`
	OCR                    *OCREvidence                   `json:"ocr"`
	Document               *DocumentConfusionEvidence     `json:"document"`
	Face                   *FaceEvidence                  `json:"face"`
	PAD                    *PADEvidence                   `json:"pad"`
	FraudCalibration       *FraudCalibrationEvidence      `json:"fraud_calibration"`
	Subgroups              *ModelQualitySubgroupEvidence  `json:"subgroups"`
}

func ComputeOCRMetrics(counts OCRCounts) (OCRMetrics, error) {
	values := []uint64{counts.CharacterReferences, counts.CharacterSubstitutions, counts.CharacterDeletions, counts.CharacterInsertions, counts.WordReferences, counts.WordSubstitutions, counts.WordDeletions, counts.WordInsertions}
	if err := validateModelQualityCounts(values...); err != nil {
		return OCRMetrics{}, err
	}
	characterErrors, err := checkedSum(counts.CharacterSubstitutions, counts.CharacterDeletions, counts.CharacterInsertions)
	if err != nil {
		return OCRMetrics{}, err
	}
	wordErrors, err := checkedSum(counts.WordSubstitutions, counts.WordDeletions, counts.WordInsertions)
	if err != nil {
		return OCRMetrics{}, err
	}
	cer, err := ratioPPM(characterErrors, counts.CharacterReferences)
	if err != nil {
		return OCRMetrics{}, errors.New("OCR character denominator must be nonzero")
	}
	wer, err := ratioPPM(wordErrors, counts.WordReferences)
	if err != nil {
		return OCRMetrics{}, errors.New("OCR word denominator must be nonzero")
	}
	return OCRMetrics{CERPPM: cer, WERPPM: wer}, nil
}

func ComputeDocumentConfusionMetrics(counts DocumentConfusionCounts) (DocumentConfusionMetrics, error) {
	if err := validateModelQualityCounts(counts.TruePositive, counts.TrueNegative, counts.FalsePositive, counts.FalseNegative); err != nil {
		return DocumentConfusionMetrics{}, err
	}
	total, err := checkedSum(counts.TruePositive, counts.TrueNegative, counts.FalsePositive, counts.FalseNegative)
	if err != nil || total > ModelQualityMaxCount {
		return DocumentConfusionMetrics{}, errors.New("document confusion total is invalid")
	}
	accuracyNumerator, _ := checkedSum(counts.TruePositive, counts.TrueNegative)
	positivePredictions, _ := checkedSum(counts.TruePositive, counts.FalsePositive)
	actualPositives, _ := checkedSum(counts.TruePositive, counts.FalseNegative)
	actualNegatives, _ := checkedSum(counts.TrueNegative, counts.FalsePositive)
	accuracy, err := ratioPPM(accuracyNumerator, total)
	if err != nil {
		return DocumentConfusionMetrics{}, errors.New("document confusion total must be nonzero")
	}
	precision, err := ratioPPM(counts.TruePositive, positivePredictions)
	if err != nil {
		return DocumentConfusionMetrics{}, errors.New("document precision denominator must be nonzero")
	}
	recall, err := ratioPPM(counts.TruePositive, actualPositives)
	if err != nil {
		return DocumentConfusionMetrics{}, errors.New("document recall denominator must be nonzero")
	}
	fpr, err := ratioPPM(counts.FalsePositive, actualNegatives)
	if err != nil {
		return DocumentConfusionMetrics{}, errors.New("document FPR denominator must be nonzero")
	}
	fnr, err := ratioPPM(counts.FalseNegative, actualPositives)
	if err != nil {
		return DocumentConfusionMetrics{}, errors.New("document FNR denominator must be nonzero")
	}
	return DocumentConfusionMetrics{AccuracyPPM: accuracy, PrecisionPPM: precision, RecallPPM: recall, FPRPPM: fpr, FNRPPM: fnr}, nil
}

func ComputeFaceOperatingPoints(counts []FaceOperatingPointCounts) ([]FaceOperatingPoint, FaceOperatingPoint, error) {
	if len(counts) == 0 {
		return nil, FaceOperatingPoint{}, errors.New("face operating points are required")
	}
	points := make([]FaceOperatingPoint, len(counts))
	for index, item := range counts {
		if err := validateModelQualityCounts(item.ThresholdPPM, item.FalseAccepts, item.ImpostorTrials, item.FalseRejects, item.GenuineTrials); err != nil {
			return nil, FaceOperatingPoint{}, err
		}
		if item.ThresholdPPM > ModelQualityPPMScale || (index > 0 && counts[index-1].ThresholdPPM >= item.ThresholdPPM) {
			return nil, FaceOperatingPoint{}, errors.New("face thresholds must be strictly increasing")
		}
		if item.FalseAccepts > item.ImpostorTrials || item.FalseRejects > item.GenuineTrials {
			return nil, FaceOperatingPoint{}, errors.New("face error count exceeds trial count")
		}
		fmr, err := ratioPPM(item.FalseAccepts, item.ImpostorTrials)
		if err != nil {
			return nil, FaceOperatingPoint{}, errors.New("face impostor denominator must be nonzero")
		}
		fnmr, err := ratioPPM(item.FalseRejects, item.GenuineTrials)
		if err != nil {
			return nil, FaceOperatingPoint{}, errors.New("face genuine denominator must be nonzero")
		}
		points[index] = FaceOperatingPoint{ThresholdPPM: item.ThresholdPPM, FMRPPM: fmr, FNMRPPM: fnmr, AverageErrorPPM: (fmr + fnmr) / 2}
	}
	eer := points[0]
	for _, point := range points[1:] {
		pointDifference := absoluteDifference(point.FMRPPM, point.FNMRPPM)
		eerDifference := absoluteDifference(eer.FMRPPM, eer.FNMRPPM)
		if pointDifference < eerDifference || (pointDifference == eerDifference && (point.ThresholdPPM < eer.ThresholdPPM || (point.ThresholdPPM == eer.ThresholdPPM && point.AverageErrorPPM < eer.AverageErrorPPM))) {
			eer = point
		}
	}
	return points, eer, nil
}

func ComputePADMetrics(counts PADCounts) (PADMetrics, error) {
	if err := validateModelQualityCounts(counts.AttackPresentations, counts.AttackPresentationsAccepted, counts.BonaFidePresentations, counts.BonaFidePresentationsRejected); err != nil {
		return PADMetrics{}, err
	}
	if counts.AttackPresentationsAccepted > counts.AttackPresentations || counts.BonaFidePresentationsRejected > counts.BonaFidePresentations {
		return PADMetrics{}, errors.New("PAD error count exceeds presentation count")
	}
	apcer, err := ratioPPM(counts.AttackPresentationsAccepted, counts.AttackPresentations)
	if err != nil {
		return PADMetrics{}, errors.New("PAD attack denominator must be nonzero")
	}
	bpcer, err := ratioPPM(counts.BonaFidePresentationsRejected, counts.BonaFidePresentations)
	if err != nil {
		return PADMetrics{}, errors.New("PAD bona fide denominator must be nonzero")
	}
	return PADMetrics{APCERPPM: apcer, BPCERPPM: bpcer}, nil
}

func ComputeFraudCalibrationMetrics(bins []FraudCalibrationBin) (FraudCalibrationMetrics, error) {
	if len(bins) == 0 {
		return FraudCalibrationMetrics{}, errors.New("fraud calibration bins are required")
	}
	metrics := FraudCalibrationMetrics{Bins: make([]FraudCalibrationBinMetrics, len(bins))}
	var totalCount uint64
	var weightedAbsoluteGap uint64
	for index, bin := range bins {
		if err := validateModelQualityCounts(bin.LowerBoundPPM, bin.UpperBoundPPM, bin.Count, bin.Positives); err != nil {
			return FraudCalibrationMetrics{}, err
		}
		expectedLower := uint64(0)
		if index > 0 {
			expectedLower = bins[index-1].UpperBoundPPM
		}
		if bin.LowerBoundPPM != expectedLower || bin.UpperBoundPPM <= bin.LowerBoundPPM || bin.UpperBoundPPM > ModelQualityPPMScale+1 {
			return FraudCalibrationMetrics{}, errors.New("fraud calibration bins must be sorted and contiguous")
		}
		if bin.Count == 0 || bin.Positives > bin.Count {
			return FraudCalibrationMetrics{}, errors.New("fraud calibration bin counts are invalid")
		}
		maximumPredictionSum, err := checkedMultiply(bin.Count, ModelQualityPPMScale)
		if err != nil || bin.PredictionSum > maximumPredictionSum {
			return FraudCalibrationMetrics{}, errors.New("fraud calibration prediction sum is invalid")
		}
		minimumBinSum, err := checkedMultiply(bin.Count, bin.LowerBoundPPM)
		if err != nil {
			return FraudCalibrationMetrics{}, errors.New("fraud calibration arithmetic overflow")
		}
		maximumBinValue := bin.UpperBoundPPM - 1
		maximumBinSum, err := checkedMultiply(bin.Count, maximumBinValue)
		if err != nil || bin.PredictionSum < minimumBinSum || bin.PredictionSum > maximumBinSum {
			return FraudCalibrationMetrics{}, errors.New("fraud calibration prediction sum is outside its bin")
		}
		totalCount, err = checkedAdd(totalCount, bin.Count)
		if err != nil || totalCount > ModelQualityMaxCount {
			return FraudCalibrationMetrics{}, errors.New("fraud calibration total count is invalid")
		}
		positiveScaled, _ := checkedMultiply(bin.Positives, ModelQualityPPMScale)
		binGap := absoluteDifference(bin.PredictionSum, positiveScaled)
		weightedAbsoluteGap, err = checkedAdd(weightedAbsoluteGap, binGap)
		if err != nil {
			return FraudCalibrationMetrics{}, errors.New("fraud calibration arithmetic overflow")
		}
		observed, _ := ratioPPM(bin.Positives, bin.Count)
		mean := bin.PredictionSum / bin.Count
		metrics.Bins[index] = FraudCalibrationBinMetrics{LowerBoundPPM: bin.LowerBoundPPM, UpperBoundPPM: bin.UpperBoundPPM, Count: bin.Count, ObservedPositivePPM: observed, MeanPredictionPPM: mean, AbsoluteGapPPM: absoluteDifference(observed, mean)}
	}
	if bins[len(bins)-1].UpperBoundPPM != ModelQualityPPMScale+1 {
		return FraudCalibrationMetrics{}, errors.New("fraud calibration bins must span the full interval")
	}
	metrics.ECEPPM = weightedAbsoluteGap / totalCount
	return metrics, nil
}

func (envelope *ModelQualityEvidenceEnvelopeV1) Validate() error {
	if envelope == nil {
		return errors.New("model quality evidence envelope is required")
	}
	if envelope.Domain != ModelQualityEnvelopeDomain || envelope.Version != ModelQualityEnvelopeVersion {
		return errors.New("unsupported model quality evidence domain or version")
	}
	for _, digest := range []string{envelope.ModelDigest, envelope.ProfileDigest, envelope.CorpusDigest, envelope.SplitDigest, envelope.LabelSchemaDigest, envelope.EvaluationConfigDigest, envelope.ThresholdProfileDigest} {
		if !isCanonicalModelQualityDigest(digest) {
			return errors.New("model quality evidence digest is malformed")
		}
	}
	if envelope.EvaluationUnixSeconds == 0 {
		return errors.New("evaluation timestamp is required")
	}
	if envelope.CorpusClass != ModelQualityCorpusSyntheticFixture && envelope.CorpusClass != ModelQualityCorpusDeclaredEvaluation {
		return errors.New("unsupported model quality corpus class")
	}
	if envelope.EvidenceUse != ModelQualityUseComputationTest && envelope.EvidenceUse != ModelQualityUseEvaluationReport {
		return errors.New("unsupported model quality evidence use")
	}
	if envelope.CertificationState != ModelQualityNotCertified {
		return errors.New("unsupported model quality certification state")
	}
	if envelope.CorpusClass == ModelQualityCorpusSyntheticFixture && (envelope.EvidenceUse != ModelQualityUseComputationTest || envelope.Representative) {
		return errors.New("synthetic fixtures are restricted to non-representative computation tests")
	}
	if envelope.CorpusClass == ModelQualityCorpusDeclaredEvaluation && envelope.EvidenceUse != ModelQualityUseEvaluationReport {
		return errors.New("declared evaluation corpus requires evaluation report use")
	}
	if envelope.OCR == nil || envelope.Document == nil || envelope.Face == nil || envelope.PAD == nil || envelope.FraudCalibration == nil || envelope.Subgroups == nil {
		return errors.New("all model quality evidence sections are required")
	}
	ocr, err := ComputeOCRMetrics(envelope.OCR.Counts)
	if err != nil || ocr != envelope.OCR.Metrics {
		return errors.New("OCR metric evidence is invalid")
	}
	document, err := ComputeDocumentConfusionMetrics(envelope.Document.Counts)
	if err != nil || document != envelope.Document.Metrics {
		return errors.New("document metric evidence is invalid")
	}
	points, eer, err := ComputeFaceOperatingPoints(envelope.Face.Counts)
	if err != nil || !equalFacePoints(points, envelope.Face.OperatingPoints) || eer != envelope.Face.EERPoint {
		return errors.New("face metric evidence is invalid")
	}
	pad, err := ComputePADMetrics(envelope.PAD.Counts)
	if err != nil || pad != envelope.PAD.Metrics {
		return errors.New("PAD metric evidence is invalid")
	}
	calibration, err := ComputeFraudCalibrationMetrics(envelope.FraudCalibration.Bins)
	if err != nil || !equalFraudCalibrationMetrics(calibration, envelope.FraudCalibration.Metrics) {
		return errors.New("fraud calibration metric evidence is invalid")
	}
	if err := envelope.Subgroups.validate(); err != nil {
		return err
	}
	return nil
}

func (envelope *ModelQualityEvidenceEnvelopeV1) CanonicalBytes() ([]byte, error) {
	if err := envelope.Validate(); err != nil {
		return nil, err
	}
	payload, err := json.Marshal(envelope)
	if err != nil {
		return nil, errors.New("model quality evidence serialization failed")
	}
	canonical := make([]byte, 0, len(ModelQualityEnvelopeDomain)+1+len(payload))
	canonical = append(canonical, ModelQualityEnvelopeDomain...)
	canonical = append(canonical, 0)
	canonical = append(canonical, payload...)
	return canonical, nil
}

func (envelope *ModelQualityEvidenceEnvelopeV1) Digest() ([32]byte, error) {
	canonical, err := envelope.CanonicalBytes()
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(canonical), nil
}

func (evidence *ModelQualitySubgroupEvidence) validate() error {
	if evidence == nil || !isCanonicalModelQualityDigest(evidence.TaxonomyDigest) {
		return errors.New("subgroup taxonomy digest is malformed")
	}
	if evidence.MinimumSampleCount < ModelQualitySubgroupFloor || evidence.MinimumSampleCount > ModelQualityMaxCount {
		return errors.New("subgroup minimum sample count violates privacy policy")
	}
	if len(evidence.Cohorts) > 64 {
		return errors.New("subgroup cohort cardinality exceeds limit")
	}
	if len(evidence.Cohorts) == 0 && evidence.SuppressedSubgroupCount == 0 {
		return errors.New("subgroup evidence requires a cohort or suppressed count")
	}
	if evidence.SuppressedSubgroupCount > ModelQualityMaxCount {
		return errors.New("suppressed subgroup count exceeds limit")
	}
	for cohortIndex, cohort := range evidence.Cohorts {
		if cohortIndex > 0 && bytes.Compare(evidence.Cohorts[cohortIndex-1].CohortToken[:], cohort.CohortToken[:]) >= 0 {
			return errors.New("subgroup cohort tokens must be strictly ordered")
		}
		if cohort.IntersectionDepth < 1 || cohort.IntersectionDepth > 3 {
			return errors.New("subgroup intersection depth is invalid")
		}
		if cohort.CohortToken == ([32]byte{}) {
			return errors.New("subgroup cohort token is required")
		}
		if cohort.SampleCount < evidence.MinimumSampleCount || cohort.SampleCount > ModelQualityMaxCount {
			return errors.New("subgroup cohort sample count violates privacy policy")
		}
		if len(cohort.Metrics) == 0 || len(cohort.Metrics) > 16 {
			return errors.New("subgroup metric cardinality is invalid")
		}
		for metricIndex, metric := range cohort.Metrics {
			if _, ok := allowedModelQualitySubgroupMetrics[metric.Kind]; !ok {
				return errors.New("unsupported subgroup metric kind")
			}
			if metricIndex > 0 && cohort.Metrics[metricIndex-1].Kind >= metric.Kind {
				return errors.New("subgroup metric kinds must be strictly ordered")
			}
			if err := validateModelQualityCounts(metric.Numerator, metric.Denominator); err != nil || metric.Denominator == 0 {
				return errors.New("subgroup metric counts are invalid")
			}
			if metric.Denominator > cohort.SampleCount {
				return errors.New("subgroup metric denominator exceeds cohort sample count")
			}
			if metric.Kind != SubgroupMetricOCRCER && metric.Kind != SubgroupMetricOCRWER && metric.Numerator > metric.Denominator {
				return errors.New("subgroup metric numerator exceeds denominator")
			}
			computed, err := ratioPPM(metric.Numerator, metric.Denominator)
			if err != nil || computed != metric.ValuePPM {
				return errors.New("subgroup metric value is invalid")
			}
		}
	}
	return nil
}

func validateModelQualityCounts(values ...uint64) error {
	for _, value := range values {
		if value > ModelQualityMaxCount {
			return errors.New("model quality count exceeds limit")
		}
	}
	return nil
}

func ratioPPM(numerator, denominator uint64) (uint64, error) {
	if denominator == 0 {
		return 0, errors.New("metric denominator must be nonzero")
	}
	scaled, err := checkedMultiply(numerator, ModelQualityPPMScale)
	if err != nil {
		return 0, errors.New("metric arithmetic overflow")
	}
	return scaled / denominator, nil
}

func checkedSum(values ...uint64) (uint64, error) {
	var total uint64
	var err error
	for _, value := range values {
		total, err = checkedAdd(total, value)
		if err != nil {
			return 0, err
		}
	}
	return total, nil
}

func checkedAdd(left, right uint64) (uint64, error) {
	if right > math.MaxUint64-left {
		return 0, errors.New("model quality arithmetic overflow")
	}
	return left + right, nil
}

func checkedMultiply(left, right uint64) (uint64, error) {
	if left != 0 && right > math.MaxUint64/left {
		return 0, errors.New("model quality arithmetic overflow")
	}
	return left * right, nil
}

func absoluteDifference(left, right uint64) uint64 {
	if left >= right {
		return left - right
	}
	return right - left
}

func isCanonicalModelQualityDigest(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size && hex.EncodeToString(decoded) == value
}

func equalFacePoints(left, right []FaceOperatingPoint) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func equalFraudCalibrationMetrics(left, right FraudCalibrationMetrics) bool {
	if left.ECEPPM != right.ECEPPM || len(left.Bins) != len(right.Bins) {
		return false
	}
	for index := range left.Bins {
		if left.Bins[index] != right.Bins[index] {
			return false
		}
	}
	return true
}
