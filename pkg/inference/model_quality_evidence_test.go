// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModelQualityOCRExactFormulaAndInsertionsOverOneHundredPercent(t *testing.T) {
	metrics, err := ComputeOCRMetrics(OCRCounts{
		CharacterReferences: 100, CharacterSubstitutions: 10, CharacterDeletions: 10, CharacterInsertions: 100,
		WordReferences: 10, WordSubstitutions: 1, WordDeletions: 1, WordInsertions: 1,
	})
	require.NoError(t, err)
	require.Equal(t, OCRMetrics{CERPPM: 1_200_000, WERPPM: 300_000}, metrics)
}

func TestModelQualityDocumentAndPADExactFormulas(t *testing.T) {
	document, err := ComputeDocumentConfusionMetrics(DocumentConfusionCounts{
		TruePositive: 60, TrueNegative: 20, FalsePositive: 10, FalseNegative: 10,
	})
	require.NoError(t, err)
	require.Equal(t, DocumentConfusionMetrics{
		AccuracyPPM: 800_000, PrecisionPPM: 857_142, RecallPPM: 857_142, FPRPPM: 333_333, FNRPPM: 142_857,
	}, document)

	pad, err := ComputePADMetrics(PADCounts{
		AttackPresentations: 10, AttackPresentationsAccepted: 2,
		BonaFidePresentations: 20, BonaFidePresentationsRejected: 1,
	})
	require.NoError(t, err)
	require.Equal(t, PADMetrics{APCERPPM: 200_000, BPCERPPM: 50_000}, pad)
}

func TestModelQualityFaceOperatingPointsEERAndTie(t *testing.T) {
	counts := []FaceOperatingPointCounts{
		{ThresholdPPM: 100_000, FalseAccepts: 4, ImpostorTrials: 10, FalseRejects: 1, GenuineTrials: 10},
		{ThresholdPPM: 200_000, FalseAccepts: 3, ImpostorTrials: 10, FalseRejects: 2, GenuineTrials: 10},
		{ThresholdPPM: 300_000, FalseAccepts: 1, ImpostorTrials: 10, FalseRejects: 4, GenuineTrials: 10},
	}
	points, eer, err := ComputeFaceOperatingPoints(counts)
	require.NoError(t, err)
	require.Equal(t, uint64(200_000), eer.ThresholdPPM)
	require.Equal(t, uint64(250_000), eer.AverageErrorPPM)
	require.Len(t, points, 3)

	tied, tiedEER, err := ComputeFaceOperatingPoints([]FaceOperatingPointCounts{
		{ThresholdPPM: 100_000, FalseAccepts: 4, ImpostorTrials: 10, FalseRejects: 3, GenuineTrials: 10},
		{ThresholdPPM: 200_000, FalseAccepts: 2, ImpostorTrials: 10, FalseRejects: 1, GenuineTrials: 10},
	})
	require.NoError(t, err)
	require.Equal(t, tied[0], tiedEER, "lower threshold wins equal-distance EER ties")
}

func TestModelQualityFraudCalibrationExactECEAndCoverage(t *testing.T) {
	bins := []FraudCalibrationBin{
		{LowerBoundPPM: 0, UpperBoundPPM: 500_000, Count: 2, Positives: 1, PredictionSum: 800_000},
		{LowerBoundPPM: 500_000, UpperBoundPPM: 1_000_001, Count: 2, Positives: 2, PredictionSum: 1_800_000},
	}
	metrics, err := ComputeFraudCalibrationMetrics(bins)
	require.NoError(t, err)
	require.Equal(t, uint64(100_000), metrics.ECEPPM)
	require.Equal(t, FraudCalibrationBinMetrics{
		LowerBoundPPM: 0, UpperBoundPPM: 500_000, Count: 2,
		ObservedPositivePPM: 500_000, MeanPredictionPPM: 400_000, AbsoluteGapPPM: 100_000,
	}, metrics.Bins[0])

	for _, invalid := range [][]FraudCalibrationBin{
		{{LowerBoundPPM: 1, UpperBoundPPM: 500_000, Count: 1}, {LowerBoundPPM: 500_000, UpperBoundPPM: 1_000_001, Count: 1}},
		{{LowerBoundPPM: 0, UpperBoundPPM: 400_000, Count: 1}, {LowerBoundPPM: 500_000, UpperBoundPPM: 1_000_001, Count: 1}},
		{{LowerBoundPPM: 0, UpperBoundPPM: 600_000, Count: 1}, {LowerBoundPPM: 500_000, UpperBoundPPM: 1_000_001, Count: 1}},
		{{LowerBoundPPM: 0, UpperBoundPPM: 500_000, Count: 2, PredictionSum: 1_200_000}, {LowerBoundPPM: 500_000, UpperBoundPPM: 1_000_001, Count: 1, PredictionSum: 500_000}},
	} {
		_, err := ComputeFraudCalibrationMetrics(invalid)
		require.Error(t, err)
	}
}

func TestModelQualityEnvelopeValidationCanonicalDigestAndTamper(t *testing.T) {
	envelope := validModelQualityEnvelope(t)
	require.NoError(t, envelope.Validate())
	firstBytes, err := envelope.CanonicalBytes()
	require.NoError(t, err)
	secondBytes, err := envelope.CanonicalBytes()
	require.NoError(t, err)
	require.Equal(t, firstBytes, secondBytes)
	require.True(t, bytes.HasPrefix(firstBytes, append([]byte(ModelQualityEnvelopeDomain), 0)))
	firstDigest, err := envelope.Digest()
	require.NoError(t, err)
	require.Equal(t, "90ace46d805a7e11b6ccefe9aa7b14f95ba8480484e931270bd5a804efa931e8", hex.EncodeToString(firstDigest[:]))

	envelope.EvaluationUnixSeconds++
	secondDigest, err := envelope.Digest()
	require.NoError(t, err)
	require.NotEqual(t, firstDigest, secondDigest)

	envelope.OCR.Metrics.CERPPM++
	require.Error(t, envelope.Validate(), "claimed metric tampering must be rejected")
}

func TestModelQualityMalformedOverflowDenominatorOrderAndDuplicateRejection(t *testing.T) {
	_, err := ComputeOCRMetrics(OCRCounts{CharacterReferences: 0, WordReferences: 1})
	require.Error(t, err)
	_, err = ComputeOCRMetrics(OCRCounts{CharacterReferences: 1, CharacterInsertions: ModelQualityMaxCount + 1, WordReferences: 1})
	require.Error(t, err)
	_, err = ratioPPM(math.MaxUint64, 1)
	require.Error(t, err)
	_, _, err = ComputeFaceOperatingPoints([]FaceOperatingPointCounts{
		{ThresholdPPM: 2, ImpostorTrials: 1, GenuineTrials: 1},
		{ThresholdPPM: 1, ImpostorTrials: 1, GenuineTrials: 1},
	})
	require.Error(t, err)
	_, _, err = ComputeFaceOperatingPoints([]FaceOperatingPointCounts{
		{ThresholdPPM: 1, ImpostorTrials: 1, GenuineTrials: 1},
		{ThresholdPPM: 1, ImpostorTrials: 1, GenuineTrials: 1},
	})
	require.Error(t, err)

	envelope := validModelQualityEnvelope(t)
	envelope.ModelDigest = "ABC-private-model-identifier"
	err = envelope.Validate()
	require.Error(t, err)
	require.NotContains(t, err.Error(), envelope.ModelDigest)

	envelope = validModelQualityEnvelope(t)
	envelope.PAD = nil
	require.Error(t, envelope.Validate())
}

func TestModelQualitySubgroupPrivacyFloorCardinalityAndOrdering(t *testing.T) {
	envelope := validModelQualityEnvelope(t)
	require.NoError(t, envelope.Validate())

	envelope.Subgroups.MinimumSampleCount = ModelQualitySubgroupFloor - 1
	require.Error(t, envelope.Validate())
	envelope = validModelQualityEnvelope(t)
	envelope.Subgroups.Cohorts[0].SampleCount = ModelQualitySubgroupFloor - 1
	require.Error(t, envelope.Validate())

	envelope = validModelQualityEnvelope(t)
	cohorts := make([]ModelQualitySubgroupCohort, 65)
	for index := range cohorts {
		cohorts[index] = envelope.Subgroups.Cohorts[0]
		cohorts[index].CohortToken[31] = byte(index + 1)
	}
	envelope.Subgroups.Cohorts = cohorts
	require.Error(t, envelope.Validate())

	envelope = validModelQualityEnvelope(t)
	second := envelope.Subgroups.Cohorts[0]
	second.CohortToken[0] = 1
	envelope.Subgroups.Cohorts = []ModelQualitySubgroupCohort{second, envelope.Subgroups.Cohorts[0]}
	require.Error(t, envelope.Validate())

	envelope = validModelQualityEnvelope(t)
	duplicate := envelope.Subgroups.Cohorts[0].Metrics[0]
	envelope.Subgroups.Cohorts[0].Metrics = append(envelope.Subgroups.Cohorts[0].Metrics, duplicate)
	require.Error(t, envelope.Validate())

	envelope = validModelQualityEnvelope(t)
	envelope.Subgroups.Cohorts[0].CohortToken = [32]byte{}
	require.Error(t, envelope.Validate())

	envelope = validModelQualityEnvelope(t)
	envelope.Subgroups.Cohorts[0].Metrics[0].Denominator = envelope.Subgroups.Cohorts[0].SampleCount + 1
	require.Error(t, envelope.Validate())

	envelope = validModelQualityEnvelope(t)
	envelope.Subgroups.Cohorts = nil
	envelope.Subgroups.SuppressedSubgroupCount = 2
	require.NoError(t, envelope.Validate(), "undersized cohorts are represented only by suppression count")
	envelope.Subgroups.SuppressedSubgroupCount = 0
	require.Error(t, envelope.Validate())
}

func TestModelQualitySyntheticContradictionAndCertificationRejection(t *testing.T) {
	envelope := validModelQualityEnvelope(t)
	envelope.Representative = true
	require.Error(t, envelope.Validate())
	envelope = validModelQualityEnvelope(t)
	envelope.EvidenceUse = ModelQualityUseEvaluationReport
	require.Error(t, envelope.Validate())
	envelope = validModelQualityEnvelope(t)
	envelope.CertificationState = ModelQualityCertificationState("certified")
	require.Error(t, envelope.Validate())
	envelope = validModelQualityEnvelope(t)
	envelope.CorpusClass = ModelQualityCorpusDeclaredEvaluation
	require.Error(t, envelope.Validate())
	envelope.EvidenceUse = ModelQualityUseEvaluationReport
	require.NoError(t, envelope.Validate())
}

func TestModelQualitySerializationAndErrorsDoNotLeakPrivateIdentifiers(t *testing.T) {
	forbidden := "named-cohort-alice-passport-region"
	envelope := validModelQualityEnvelope(t)
	canonical, err := envelope.CanonicalBytes()
	require.NoError(t, err)
	require.NotContains(t, string(canonical), forbidden)
	serialized, err := json.Marshal(envelope)
	require.NoError(t, err)
	require.NotContains(t, string(serialized), forbidden)

	envelope.CorpusClass = ModelQualityCorpusClass(forbidden)
	err = envelope.Validate()
	require.Error(t, err)
	require.NotContains(t, err.Error(), forbidden)
	envelope = validModelQualityEnvelope(t)
	envelope.Subgroups.Cohorts[0].Metrics[0].Kind = ModelQualitySubgroupMetricKind(forbidden)
	err = envelope.Validate()
	require.Error(t, err)
	require.NotContains(t, err.Error(), forbidden)
}

func validModelQualityEnvelope(t *testing.T) *ModelQualityEvidenceEnvelopeV1 {
	t.Helper()
	ocrCounts := OCRCounts{
		CharacterReferences: 100, CharacterSubstitutions: 5, CharacterDeletions: 3, CharacterInsertions: 2,
		WordReferences: 20, WordSubstitutions: 1, WordDeletions: 1, WordInsertions: 1,
	}
	ocr, err := ComputeOCRMetrics(ocrCounts)
	require.NoError(t, err)
	documentCounts := DocumentConfusionCounts{TruePositive: 60, TrueNegative: 20, FalsePositive: 10, FalseNegative: 10}
	document, err := ComputeDocumentConfusionMetrics(documentCounts)
	require.NoError(t, err)
	faceCounts := []FaceOperatingPointCounts{
		{ThresholdPPM: 100_000, FalseAccepts: 4, ImpostorTrials: 10, FalseRejects: 1, GenuineTrials: 10},
		{ThresholdPPM: 200_000, FalseAccepts: 2, ImpostorTrials: 10, FalseRejects: 2, GenuineTrials: 10},
	}
	facePoints, eer, err := ComputeFaceOperatingPoints(faceCounts)
	require.NoError(t, err)
	padCounts := PADCounts{AttackPresentations: 10, AttackPresentationsAccepted: 1, BonaFidePresentations: 20, BonaFidePresentationsRejected: 2}
	pad, err := ComputePADMetrics(padCounts)
	require.NoError(t, err)
	bins := []FraudCalibrationBin{
		{LowerBoundPPM: 0, UpperBoundPPM: 500_000, Count: 2, Positives: 1, PredictionSum: 800_000},
		{LowerBoundPPM: 500_000, UpperBoundPPM: 1_000_001, Count: 2, Positives: 2, PredictionSum: 1_800_000},
	}
	calibration, err := ComputeFraudCalibrationMetrics(bins)
	require.NoError(t, err)
	metric := ModelQualitySubgroupMetric{Kind: SubgroupMetricFaceFMR, Numerator: 1, Denominator: 20, ValuePPM: 50_000}
	var token [32]byte
	token[0] = 1
	return &ModelQualityEvidenceEnvelopeV1{
		Domain: ModelQualityEnvelopeDomain, Version: ModelQualityEnvelopeVersion,
		ModelDigest: modelQualityTestDigest("model"), ProfileDigest: modelQualityTestDigest("profile"),
		CorpusDigest: modelQualityTestDigest("corpus"), SplitDigest: modelQualityTestDigest("split"),
		LabelSchemaDigest: modelQualityTestDigest("labels"), EvaluationConfigDigest: modelQualityTestDigest("evaluation"),
		ThresholdProfileDigest: modelQualityTestDigest("thresholds"), EvaluationUnixSeconds: 1_785_628_800,
		CorpusClass: ModelQualityCorpusSyntheticFixture, EvidenceUse: ModelQualityUseComputationTest,
		CertificationState: ModelQualityNotCertified, Representative: false,
		OCR:              &OCREvidence{Counts: ocrCounts, Metrics: ocr},
		Document:         &DocumentConfusionEvidence{Counts: documentCounts, Metrics: document},
		Face:             &FaceEvidence{Counts: faceCounts, OperatingPoints: facePoints, EERPoint: eer},
		PAD:              &PADEvidence{Counts: padCounts, Metrics: pad},
		FraudCalibration: &FraudCalibrationEvidence{Bins: bins, Metrics: calibration},
		Subgroups: &ModelQualitySubgroupEvidence{
			TaxonomyDigest: modelQualityTestDigest("taxonomy"), MinimumSampleCount: 20,
			Cohorts: []ModelQualitySubgroupCohort{{CohortToken: token, IntersectionDepth: 1, SampleCount: 20, Metrics: []ModelQualitySubgroupMetric{metric}}},
		},
	}
}

func modelQualityTestDigest(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}
