"""
Tests for the OCR Evaluation Framework.

VE-3045: Create OCR Evaluation Framework
"""

import pytest
import json
import tempfile
from pathlib import Path
from typing import Dict, List

from ml.ocr_extraction.evaluation import (
    # Data structures
    GroundTruth,
    PredictedResult,
    EvaluationSample,
    FieldResult,
    FieldMetrics,
    TextMetrics,
    SampleReport,
    BatchReport,
    MatchType,
    # Evaluator
    OCREvaluator,
    # Utilities
    GroundTruthDataset,
    create_evaluation_sample,
    quick_evaluate,
    # Distance functions
    levenshtein_distance,
    word_levenshtein_distance,
)


# =============================================================================
# Fixtures
# =============================================================================


@pytest.fixture
def evaluator() -> OCREvaluator:
    """Create a default evaluator."""
    return OCREvaluator(
        case_sensitive=False,
        partial_match_threshold=0.8,
        normalize_whitespace=True,
    )


@pytest.fixture
def case_sensitive_evaluator() -> OCREvaluator:
    """Create a case-sensitive evaluator."""
    return OCREvaluator(
        case_sensitive=True,
        partial_match_threshold=0.8,
        normalize_whitespace=True,
    )


@pytest.fixture
def sample_ground_truth() -> GroundTruth:
    """Create a sample ground truth."""
    return GroundTruth(
        image_path="/data/images/id_001.jpg",
        document_type="id_card",
        fields={
            "surname": "SMITH",
            "given_names": "JOHN MICHAEL",
            "date_of_birth": "1990-01-15",
            "id_number": "AB123456",
        },
        full_text="SURNAME: SMITH GIVEN NAMES: JOHN MICHAEL DOB: 1990-01-15 ID: AB123456",
    )


@pytest.fixture
def sample_predicted() -> PredictedResult:
    """Create a sample predicted result."""
    return PredictedResult(
        fields={
            "surname": "SMITH",
            "given_names": "JOHN MICHEAL",  # Typo: MICHEAL instead of MICHAEL
            "date_of_birth": "1990-01-15",
            # Missing id_number
        },
        full_text="SURNAME: SMITH GIVEN NAMES: JOHN MICHEAL DOB: 1990-01-15",
        confidence=0.85,
        processing_time_ms=150.0,
    )


@pytest.fixture
def sample_evaluation_sample(
    sample_ground_truth: GroundTruth,
    sample_predicted: PredictedResult,
) -> EvaluationSample:
    """Create a sample evaluation sample."""
    return EvaluationSample(
        ground_truth=sample_ground_truth,
        predicted=sample_predicted,
        sample_id="sample_001",
    )


@pytest.fixture
def batch_samples() -> List[EvaluationSample]:
    """Create a batch of evaluation samples."""
    samples = []
    
    # Sample 1: Perfect match
    samples.append(EvaluationSample(
        ground_truth=GroundTruth(
            image_path="/data/images/id_001.jpg",
            document_type="id_card",
            fields={"name": "JOHN SMITH", "dob": "1990-01-01"},
            full_text="JOHN SMITH 1990-01-01",
        ),
        predicted=PredictedResult(
            fields={"name": "JOHN SMITH", "dob": "1990-01-01"},
            full_text="JOHN SMITH 1990-01-01",
            confidence=0.95,
            processing_time_ms=100.0,
        ),
        sample_id="perfect_match",
    ))
    
    # Sample 2: Partial match
    samples.append(EvaluationSample(
        ground_truth=GroundTruth(
            image_path="/data/images/id_002.jpg",
            document_type="passport",
            fields={"name": "JANE DOE", "passport_no": "X12345678"},
            full_text="JANE DOE X12345678",
        ),
        predicted=PredictedResult(
            fields={"name": "JANE D0E", "passport_no": "X12345678"},  # 0 instead of O
            full_text="JANE D0E X12345678",
            confidence=0.88,
            processing_time_ms=120.0,
        ),
        sample_id="partial_match",
    ))
    
    # Sample 3: Poor match
    samples.append(EvaluationSample(
        ground_truth=GroundTruth(
            image_path="/data/images/id_003.jpg",
            document_type="id_card",
            fields={"name": "ROBERT JOHNSON", "id": "999888777"},
            full_text="ROBERT JOHNSON 999888777",
        ),
        predicted=PredictedResult(
            fields={"name": "R0BERT J0HNS0N", "id": "998888777"},  # Multiple errors
            full_text="R0BERT J0HNS0N 998888777",
            confidence=0.72,
            processing_time_ms=180.0,
        ),
        sample_id="poor_match",
    ))
    
    return samples


# =============================================================================
# Levenshtein Distance Tests
# =============================================================================


class TestLevenshteinDistance:
    """Tests for Levenshtein distance calculation."""
    
    def test_identical_strings(self):
        """Identical strings have distance 0."""
        assert levenshtein_distance("hello", "hello") == 0
        assert levenshtein_distance("", "") == 0
        assert levenshtein_distance("abc123", "abc123") == 0
    
    def test_empty_strings(self):
        """Distance to empty string equals length."""
        assert levenshtein_distance("hello", "") == 5
        assert levenshtein_distance("", "world") == 5
    
    def test_single_insertion(self):
        """Single insertion has distance 1."""
        assert levenshtein_distance("helo", "hello") == 1
        assert levenshtein_distance("cat", "cats") == 1
    
    def test_single_deletion(self):
        """Single deletion has distance 1."""
        assert levenshtein_distance("hello", "helo") == 1
        assert levenshtein_distance("cats", "cat") == 1
    
    def test_single_substitution(self):
        """Single substitution has distance 1."""
        assert levenshtein_distance("hello", "hallo") == 1
        assert levenshtein_distance("cat", "bat") == 1
    
    def test_multiple_operations(self):
        """Multiple operations are counted correctly."""
        # "kitten" -> "sitting": k->s, e->i, +g = 3
        assert levenshtein_distance("kitten", "sitting") == 3
        
        # "saturday" -> "sunday": at->un, delete r = 3
        assert levenshtein_distance("saturday", "sunday") == 3
    
    def test_symmetry(self):
        """Distance is symmetric."""
        assert levenshtein_distance("abc", "xyz") == levenshtein_distance("xyz", "abc")
        assert levenshtein_distance("hello", "helo") == levenshtein_distance("helo", "hello")


class TestWordLevenshteinDistance:
    """Tests for word-level Levenshtein distance."""
    
    def test_identical_word_lists(self):
        """Identical word lists have distance 0."""
        assert word_levenshtein_distance(["hello", "world"], ["hello", "world"]) == 0
        assert word_levenshtein_distance([], []) == 0
    
    def test_empty_word_lists(self):
        """Distance to empty list equals length."""
        assert word_levenshtein_distance(["hello", "world"], []) == 2
        assert word_levenshtein_distance([], ["one", "two", "three"]) == 3
    
    def test_single_word_difference(self):
        """Single word difference has distance 1."""
        assert word_levenshtein_distance(["hello"], ["hello", "world"]) == 1
        assert word_levenshtein_distance(["hello", "world"], ["hello"]) == 1
        assert word_levenshtein_distance(["hello", "world"], ["hello", "there"]) == 1
    
    def test_multiple_word_differences(self):
        """Multiple word differences are counted."""
        assert word_levenshtein_distance(
            ["the", "quick", "brown", "fox"],
            ["a", "slow", "red", "cat"]
        ) == 4


# =============================================================================
# Character Error Rate Tests
# =============================================================================


class TestCharacterErrorRate:
    """Tests for CER calculation."""
    
    def test_perfect_match(self, evaluator: OCREvaluator):
        """Perfect match has CER of 0."""
        assert evaluator.character_error_rate("hello world", "hello world") == 0.0
    
    def test_empty_ground_truth(self, evaluator: OCREvaluator):
        """Empty ground truth with prediction returns prediction length."""
        assert evaluator.character_error_rate("hello", "") == 5.0
    
    def test_empty_prediction(self, evaluator: OCREvaluator):
        """Empty prediction with ground truth returns 1.0."""
        assert evaluator.character_error_rate("", "hello") == 1.0
    
    def test_both_empty(self, evaluator: OCREvaluator):
        """Both empty returns 0."""
        assert evaluator.character_error_rate("", "") == 0.0
    
    def test_single_error(self, evaluator: OCREvaluator):
        """Single character error is counted correctly."""
        # "helo" vs "hello" - 1 error, 5 chars = 0.2
        cer = evaluator.character_error_rate("helo", "hello")
        assert abs(cer - 0.2) < 0.001
    
    def test_case_insensitive(self, evaluator: OCREvaluator):
        """Default evaluator is case insensitive."""
        assert evaluator.character_error_rate("HELLO", "hello") == 0.0
        assert evaluator.character_error_rate("HeLLo WoRLd", "hello world") == 0.0
    
    def test_case_sensitive(self, case_sensitive_evaluator: OCREvaluator):
        """Case-sensitive evaluator counts case differences."""
        cer = case_sensitive_evaluator.character_error_rate("HELLO", "hello")
        assert cer > 0.0  # All characters differ
    
    def test_whitespace_normalization(self, evaluator: OCREvaluator):
        """Whitespace is normalized."""
        assert evaluator.character_error_rate("hello   world", "hello world") == 0.0
        assert evaluator.character_error_rate("hello\nworld", "hello world") == 0.0
        assert evaluator.character_error_rate("  hello  world  ", "hello world") == 0.0
    
    def test_known_cer_values(self, evaluator: OCREvaluator):
        """Test with known CER values."""
        # 3 errors in 10 character string = 0.3
        cer = evaluator.character_error_rate("abcdefghij", "xbcdefghyz")
        assert abs(cer - 0.3) < 0.001
        
        # Complete mismatch - 5 substitutions in 5 chars = 1.0
        cer = evaluator.character_error_rate("abcde", "vwxyz")
        assert cer == 1.0


# =============================================================================
# Word Error Rate Tests
# =============================================================================


class TestWordErrorRate:
    """Tests for WER calculation."""
    
    def test_perfect_match(self, evaluator: OCREvaluator):
        """Perfect match has WER of 0."""
        assert evaluator.word_error_rate("hello world", "hello world") == 0.0
    
    def test_empty_ground_truth(self, evaluator: OCREvaluator):
        """Empty ground truth with prediction returns word count."""
        assert evaluator.word_error_rate("hello world", "") == 2.0
    
    def test_empty_prediction(self, evaluator: OCREvaluator):
        """Empty prediction returns 1.0."""
        assert evaluator.word_error_rate("", "hello world") == 1.0
    
    def test_both_empty(self, evaluator: OCREvaluator):
        """Both empty returns 0."""
        assert evaluator.word_error_rate("", "") == 0.0
    
    def test_single_word_error(self, evaluator: OCREvaluator):
        """Single word error is counted correctly."""
        # "hello word" vs "hello world" - 1 error, 2 words = 0.5
        wer = evaluator.word_error_rate("hello word", "hello world")
        assert wer == 0.5
    
    def test_case_insensitive(self, evaluator: OCREvaluator):
        """Default evaluator is case insensitive."""
        assert evaluator.word_error_rate("HELLO WORLD", "hello world") == 0.0
    
    def test_known_wer_values(self, evaluator: OCREvaluator):
        """Test with known WER values."""
        # 2 word errors in 4 words = 0.5
        wer = evaluator.word_error_rate("the quick red dog", "the quick brown fox")
        assert wer == 0.5
        
        # All words wrong - 3 errors in 3 words = 1.0
        wer = evaluator.word_error_rate("one two three", "four five six")
        assert wer == 1.0


# =============================================================================
# Similarity Score Tests
# =============================================================================


class TestSimilarityScore:
    """Tests for similarity score calculation."""
    
    def test_identical_strings(self, evaluator: OCREvaluator):
        """Identical strings have similarity of 1.0."""
        assert evaluator.similarity_score("hello", "hello") == 1.0
    
    def test_completely_different(self, evaluator: OCREvaluator):
        """Completely different strings have low similarity."""
        similarity = evaluator.similarity_score("abc", "xyz")
        assert similarity == 0.0  # All 3 chars different
    
    def test_empty_strings(self, evaluator: OCREvaluator):
        """Both empty strings have similarity 1.0."""
        assert evaluator.similarity_score("", "") == 1.0
    
    def test_one_empty(self, evaluator: OCREvaluator):
        """One empty string has similarity 0.0."""
        assert evaluator.similarity_score("hello", "") == 0.0
        assert evaluator.similarity_score("", "hello") == 0.0
    
    def test_partial_similarity(self, evaluator: OCREvaluator):
        """Partial similarity is calculated correctly."""
        # "helo" vs "hello" - 1 error in 5 chars = 0.8 similarity
        similarity = evaluator.similarity_score("helo", "hello")
        assert abs(similarity - 0.8) < 0.001


# =============================================================================
# Field Accuracy Tests
# =============================================================================


class TestFieldAccuracy:
    """Tests for field-level accuracy metrics."""
    
    def test_perfect_field_match(self, evaluator: OCREvaluator):
        """Perfect field matches have precision/recall of 1.0."""
        predicted = {"name": "JOHN", "dob": "1990-01-01"}
        ground_truth = {"name": "JOHN", "dob": "1990-01-01"}
        
        metrics = evaluator.field_accuracy(predicted, ground_truth)
        
        assert metrics.precision == 1.0
        assert metrics.recall == 1.0
        assert metrics.f1_score == 1.0
        assert metrics.exact_match_rate == 1.0
    
    def test_missing_field(self, evaluator: OCREvaluator):
        """Missing field reduces recall."""
        predicted = {"name": "JOHN"}
        ground_truth = {"name": "JOHN", "dob": "1990-01-01"}
        
        metrics = evaluator.field_accuracy(predicted, ground_truth)
        
        assert metrics.precision == 1.0  # All predicted are correct
        assert metrics.recall == 0.5     # Only half of expected found
        assert metrics.exact_match_rate == 0.5
    
    def test_extra_field(self, evaluator: OCREvaluator):
        """Extra field doesn't affect recall but affects precision."""
        predicted = {"name": "JOHN", "dob": "1990-01-01", "extra": "value"}
        ground_truth = {"name": "JOHN", "dob": "1990-01-01"}
        
        metrics = evaluator.field_accuracy(predicted, ground_truth)
        
        # Precision: 2 correct / 3 predicted = 0.667
        assert abs(metrics.precision - 2/3) < 0.001
        assert metrics.recall == 1.0  # All expected found
    
    def test_partial_match(self, evaluator: OCREvaluator):
        """Partial matches are detected with threshold."""
        predicted = {"name": "JOHN SMIT"}  # Missing 'H'
        ground_truth = {"name": "JOHN SMITH"}
        
        metrics = evaluator.field_accuracy(predicted, ground_truth)
        
        # Should be partial match (similarity ~0.9)
        assert len(metrics.field_results) == 1
        result = metrics.field_results[0]
        assert result.match_type == MatchType.PARTIAL
        assert result.similarity_score > 0.8
    
    def test_mismatch(self, evaluator: OCREvaluator):
        """Complete mismatch is detected."""
        predicted = {"name": "ALICE JOHNSON"}
        ground_truth = {"name": "BOB SMITH"}
        
        metrics = evaluator.field_accuracy(predicted, ground_truth)
        
        assert metrics.exact_match_rate == 0.0
        result = metrics.field_results[0]
        assert result.match_type == MatchType.MISMATCH
    
    def test_empty_fields(self, evaluator: OCREvaluator):
        """Empty field dictionaries are handled."""
        metrics = evaluator.field_accuracy({}, {})
        
        assert metrics.precision == 0.0
        assert metrics.recall == 0.0
        assert metrics.f1_score == 0.0
    
    def test_case_insensitive_fields(self, evaluator: OCREvaluator):
        """Field values are compared case-insensitively by default."""
        predicted = {"name": "john smith"}
        ground_truth = {"name": "JOHN SMITH"}
        
        metrics = evaluator.field_accuracy(predicted, ground_truth)
        
        assert metrics.exact_match_rate == 1.0


# =============================================================================
# Sample Evaluation Tests
# =============================================================================


class TestSampleEvaluation:
    """Tests for single sample evaluation."""
    
    def test_evaluate_sample(
        self,
        evaluator: OCREvaluator,
        sample_evaluation_sample: EvaluationSample,
    ):
        """Evaluate a single sample."""
        report = evaluator.evaluate_sample(sample_evaluation_sample)
        
        assert report.sample_id == "sample_001"
        assert report.document_type == "id_card"
        assert report.text_metrics is not None
        assert report.field_metrics is not None
    
    def test_evaluate_sample_text_metrics(
        self,
        evaluator: OCREvaluator,
        sample_evaluation_sample: EvaluationSample,
    ):
        """Text metrics are calculated correctly."""
        report = evaluator.evaluate_sample(sample_evaluation_sample)
        
        # Text has one character difference (MICHAEL vs MICHEAL)
        assert report.text_metrics.character_error_rate > 0
        assert report.text_metrics.word_error_rate > 0
    
    def test_evaluate_sample_field_metrics(
        self,
        evaluator: OCREvaluator,
        sample_evaluation_sample: EvaluationSample,
    ):
        """Field metrics are calculated correctly."""
        report = evaluator.evaluate_sample(sample_evaluation_sample)
        
        # 2 exact matches (surname, dob), 1 partial (given_names), 1 missing (id_number)
        assert report.field_metrics.exact_match_rate < 1.0
        assert report.field_metrics.recall < 1.0  # Missing field
    
    def test_evaluate_sample_without_text(self, evaluator: OCREvaluator):
        """Sample without full_text has no text metrics."""
        sample = EvaluationSample(
            ground_truth=GroundTruth(
                image_path="/test.jpg",
                document_type="id_card",
                fields={"name": "JOHN"},
                full_text=None,
            ),
            predicted=PredictedResult(
                fields={"name": "JOHN"},
                full_text=None,
            ),
        )
        
        report = evaluator.evaluate_sample(sample)
        
        assert report.text_metrics is None
        assert report.field_metrics is not None


# =============================================================================
# Batch Evaluation Tests
# =============================================================================


class TestBatchEvaluation:
    """Tests for batch evaluation."""
    
    def test_evaluate_batch(
        self,
        evaluator: OCREvaluator,
        batch_samples: List[EvaluationSample],
    ):
        """Evaluate a batch of samples."""
        report = evaluator.evaluate_batch(batch_samples)
        
        assert report.total_samples == 3
        assert report.samples_evaluated == 3
        assert report.samples_failed == 0
        assert len(report.sample_reports) == 3
    
    def test_batch_aggregate_metrics(
        self,
        evaluator: OCREvaluator,
        batch_samples: List[EvaluationSample],
    ):
        """Aggregate metrics are calculated."""
        report = evaluator.evaluate_batch(batch_samples)
        
        # Should have aggregated metrics
        assert 0.0 <= report.mean_cer <= 1.0
        assert 0.0 <= report.mean_wer <= 1.0
        assert 0.0 <= report.mean_precision <= 1.0
        assert 0.0 <= report.mean_recall <= 1.0
        assert 0.0 <= report.mean_f1 <= 1.0
    
    def test_batch_field_statistics(
        self,
        evaluator: OCREvaluator,
        batch_samples: List[EvaluationSample],
    ):
        """Per-field statistics are computed."""
        report = evaluator.evaluate_batch(batch_samples)
        
        assert len(report.field_statistics) > 0
        for field_name, stats in report.field_statistics.items():
            assert "exact_match_rate" in stats
            assert "mean_cer" in stats
            assert "sample_count" in stats
    
    def test_batch_document_type_statistics(
        self,
        evaluator: OCREvaluator,
        batch_samples: List[EvaluationSample],
    ):
        """Per-document-type statistics are computed."""
        report = evaluator.evaluate_batch(batch_samples)
        
        # Should have id_card and passport types
        assert "id_card" in report.document_type_statistics
        assert "passport" in report.document_type_statistics
        
        # id_card has 2 samples, passport has 1
        assert report.document_type_statistics["id_card"]["sample_count"] == 2
        assert report.document_type_statistics["passport"]["sample_count"] == 1
    
    def test_batch_processing_time(
        self,
        evaluator: OCREvaluator,
        batch_samples: List[EvaluationSample],
    ):
        """Processing time is aggregated."""
        report = evaluator.evaluate_batch(batch_samples)
        
        assert report.total_processing_time_ms == 400.0  # 100 + 120 + 180
        assert abs(report.mean_processing_time_ms - 400.0/3) < 0.001
    
    def test_empty_batch(self, evaluator: OCREvaluator):
        """Empty batch returns zero metrics."""
        report = evaluator.evaluate_batch([])
        
        assert report.total_samples == 0
        assert report.samples_evaluated == 0
        assert report.mean_cer == 0.0
        assert report.mean_f1 == 0.0
    
    def test_batch_report_to_dict(
        self,
        evaluator: OCREvaluator,
        batch_samples: List[EvaluationSample],
    ):
        """Batch report converts to dictionary."""
        report = evaluator.evaluate_batch(batch_samples)
        data = report.to_dict()
        
        assert "summary" in data
        assert "text_metrics" in data
        assert "field_metrics" in data
        assert "field_statistics" in data
        assert "sample_reports" in data
    
    def test_batch_report_summary(
        self,
        evaluator: OCREvaluator,
        batch_samples: List[EvaluationSample],
    ):
        """Batch report generates human-readable summary."""
        report = evaluator.evaluate_batch(batch_samples)
        summary = report.summary_string()
        
        assert "OCR EVALUATION REPORT" in summary
        assert "Character Error Rate" in summary
        assert "Word Error Rate" in summary
        assert "F1 Score" in summary


# =============================================================================
# Ground Truth Data Structure Tests
# =============================================================================


class TestGroundTruth:
    """Tests for GroundTruth data structure."""
    
    def test_create_ground_truth(self):
        """Create GroundTruth with all fields."""
        gt = GroundTruth(
            image_path="/path/to/image.jpg",
            document_type="id_card",
            fields={"name": "JOHN", "dob": "1990-01-01"},
            full_text="JOHN 1990-01-01",
            metadata={"annotator": "test"},
        )
        
        assert gt.image_path == "/path/to/image.jpg"
        assert gt.document_type == "id_card"
        assert gt.fields["name"] == "JOHN"
        assert gt.full_text == "JOHN 1990-01-01"
        assert gt.metadata["annotator"] == "test"
    
    def test_ground_truth_to_dict(self):
        """GroundTruth converts to dictionary."""
        gt = GroundTruth(
            image_path="/path/to/image.jpg",
            document_type="passport",
            fields={"name": "JANE"},
        )
        
        data = gt.to_dict()
        
        assert data["image_path"] == "/path/to/image.jpg"
        assert data["document_type"] == "passport"
        assert data["fields"]["name"] == "JANE"
    
    def test_ground_truth_from_dict(self):
        """GroundTruth loads from dictionary."""
        data = {
            "image_path": "/path/to/image.jpg",
            "document_type": "id_card",
            "fields": {"name": "BOB"},
            "full_text": "BOB SMITH",
        }
        
        gt = GroundTruth.from_dict(data)
        
        assert gt.image_path == "/path/to/image.jpg"
        assert gt.document_type == "id_card"
        assert gt.fields["name"] == "BOB"
        assert gt.full_text == "BOB SMITH"
    
    def test_ground_truth_json_roundtrip(self):
        """GroundTruth survives JSON roundtrip."""
        original = GroundTruth(
            image_path="/test.jpg",
            document_type="passport",
            fields={"name": "TEST", "id": "123"},
            full_text="TEST 123",
            metadata={"version": "1.0"},
        )
        
        with tempfile.NamedTemporaryFile(suffix=".json", delete=False) as f:
            path = f.name
        
        try:
            original.to_json_file(path)
            loaded = GroundTruth.from_json_file(path)
            
            assert loaded.image_path == original.image_path
            assert loaded.document_type == original.document_type
            assert loaded.fields == original.fields
            assert loaded.full_text == original.full_text
            assert loaded.metadata == original.metadata
        finally:
            Path(path).unlink(missing_ok=True)


# =============================================================================
# Ground Truth Dataset Tests
# =============================================================================


class TestGroundTruthDataset:
    """Tests for GroundTruthDataset."""
    
    def test_create_empty_dataset(self):
        """Create empty dataset."""
        dataset = GroundTruthDataset()
        assert len(dataset) == 0
    
    def test_add_ground_truth(self):
        """Add ground truth to dataset."""
        dataset = GroundTruthDataset()
        gt = GroundTruth(
            image_path="/test.jpg",
            document_type="id_card",
            fields={"name": "TEST"},
        )
        
        dataset.add(gt)
        
        assert len(dataset) == 1
    
    def test_iterate_dataset(self):
        """Iterate over dataset."""
        gts = [
            GroundTruth(image_path=f"/{i}.jpg", document_type="id_card", fields={})
            for i in range(5)
        ]
        dataset = GroundTruthDataset(gts)
        
        count = sum(1 for _ in dataset)
        assert count == 5
    
    def test_filter_by_document_type(self):
        """Filter dataset by document type."""
        gts = [
            GroundTruth(image_path="/1.jpg", document_type="id_card", fields={}),
            GroundTruth(image_path="/2.jpg", document_type="passport", fields={}),
            GroundTruth(image_path="/3.jpg", document_type="id_card", fields={}),
        ]
        dataset = GroundTruthDataset(gts)
        
        filtered = dataset.filter_by_document_type("id_card")
        
        assert len(filtered) == 2
        for gt in filtered:
            assert gt.document_type == "id_card"
    
    def test_split_dataset(self):
        """Split dataset into train/test."""
        gts = [
            GroundTruth(image_path=f"/{i}.jpg", document_type="id_card", fields={})
            for i in range(10)
        ]
        dataset = GroundTruthDataset(gts)
        
        train, test = dataset.split(train_ratio=0.8, seed=42)
        
        assert len(train) == 8
        assert len(test) == 2
    
    def test_dataset_json_roundtrip(self):
        """Dataset survives JSON roundtrip."""
        gts = [
            GroundTruth(
                image_path=f"/{i}.jpg",
                document_type="id_card",
                fields={"id": str(i)},
            )
            for i in range(3)
        ]
        original = GroundTruthDataset(gts)
        
        with tempfile.NamedTemporaryFile(suffix=".json", delete=False) as f:
            path = f.name
        
        try:
            original.to_json_file(path)
            loaded = GroundTruthDataset.from_json_file(path)
            
            assert len(loaded) == len(original)
            for orig, load in zip(original, loaded):
                assert orig.image_path == load.image_path
                assert orig.fields == load.fields
        finally:
            Path(path).unlink(missing_ok=True)


# =============================================================================
# Edge Case Tests
# =============================================================================


class TestEdgeCases:
    """Tests for edge cases and boundary conditions."""
    
    def test_unicode_characters(self, evaluator: OCREvaluator):
        """Unicode characters are handled correctly."""
        cer = evaluator.character_error_rate("café", "cafe")
        assert cer > 0  # é != e
        
        # Exact match with unicode
        assert evaluator.character_error_rate("日本語", "日本語") == 0.0
    
    def test_special_characters(self, evaluator: OCREvaluator):
        """Special characters are handled."""
        assert evaluator.character_error_rate("A-123/456", "A-123/456") == 0.0
        
        # Punctuation differences
        cer = evaluator.character_error_rate("hello!", "hello?")
        assert cer > 0
    
    def test_very_long_strings(self, evaluator: OCREvaluator):
        """Very long strings are handled efficiently."""
        long_str = "a" * 10000
        long_str_with_error = "b" + "a" * 9999
        
        cer = evaluator.character_error_rate(long_str_with_error, long_str)
        assert cer == 0.0001  # 1 error in 10000 chars
    
    def test_single_character_strings(self, evaluator: OCREvaluator):
        """Single character strings work."""
        assert evaluator.character_error_rate("a", "a") == 0.0
        assert evaluator.character_error_rate("a", "b") == 1.0
    
    def test_whitespace_only(self, evaluator: OCREvaluator):
        """Whitespace-only strings are normalized to empty."""
        # Both become empty after normalization
        assert evaluator.character_error_rate("   ", "   ") == 0.0
    
    def test_newlines_and_tabs(self, evaluator: OCREvaluator):
        """Newlines and tabs are normalized."""
        assert evaluator.character_error_rate("hello\tworld", "hello world") == 0.0
        assert evaluator.character_error_rate("hello\nworld", "hello world") == 0.0
    
    def test_field_with_none_values(self, evaluator: OCREvaluator):
        """Fields with None values are handled."""
        predicted: Dict[str, str] = {}  # Empty dict
        ground_truth = {"name": "JOHN"}
        
        metrics = evaluator.field_accuracy(predicted, ground_truth)
        
        assert metrics.precision == 0.0
        assert metrics.recall == 0.0
        
        # Check field result
        assert len(metrics.field_results) == 1
        assert metrics.field_results[0].match_type == MatchType.MISSING


# =============================================================================
# Utility Function Tests
# =============================================================================


class TestUtilityFunctions:
    """Tests for utility functions."""
    
    def test_create_evaluation_sample(self):
        """create_evaluation_sample creates valid sample."""
        sample = create_evaluation_sample(
            image_path="/test.jpg",
            document_type="id_card",
            ground_truth_fields={"name": "JOHN"},
            predicted_fields={"name": "JOHN"},
            ground_truth_text="JOHN SMITH",
            predicted_text="JOHN SMITH",
            sample_id="test_001",
        )
        
        assert sample.sample_id == "test_001"
        assert sample.ground_truth.image_path == "/test.jpg"
        assert sample.ground_truth.document_type == "id_card"
        assert sample.predicted.fields["name"] == "JOHN"
    
    def test_quick_evaluate(self):
        """quick_evaluate returns expected metrics."""
        result = quick_evaluate("hello world", "hello world")
        
        assert result["cer"] == 0.0
        assert result["wer"] == 0.0
        assert result["similarity"] == 1.0
    
    def test_quick_evaluate_with_errors(self):
        """quick_evaluate detects errors."""
        result = quick_evaluate("helo world", "hello world")
        
        assert result["cer"] > 0.0
        assert result["similarity"] < 1.0
    
    def test_quick_evaluate_case_sensitive(self):
        """quick_evaluate respects case_sensitive flag."""
        result_insensitive = quick_evaluate("HELLO", "hello", case_sensitive=False)
        result_sensitive = quick_evaluate("HELLO", "hello", case_sensitive=True)
        
        assert result_insensitive["cer"] == 0.0
        assert result_sensitive["cer"] > 0.0


# =============================================================================
# Integration Tests
# =============================================================================


class TestIntegration:
    """Integration tests for the complete evaluation workflow."""
    
    def test_complete_evaluation_workflow(self, evaluator: OCREvaluator):
        """Test complete evaluation workflow."""
        # Create ground truth dataset
        ground_truths = [
            GroundTruth(
                image_path=f"/data/doc_{i}.jpg",
                document_type="id_card" if i % 2 == 0 else "passport",
                fields={
                    "name": f"PERSON {i}",
                    "id_number": f"ID{i:04d}",
                },
                full_text=f"PERSON {i} ID{i:04d}",
            )
            for i in range(5)
        ]
        
        # Simulate predictions with varying accuracy
        predictions = [
            PredictedResult(
                fields={
                    "name": f"PERSON {i}" if i % 3 != 0 else f"PERS0N {i}",  # Error every 3rd
                    "id_number": f"ID{i:04d}",
                },
                full_text=f"PERSON {i} ID{i:04d}" if i % 3 != 0 else f"PERS0N {i} ID{i:04d}",
                confidence=0.9 if i % 3 != 0 else 0.7,
                processing_time_ms=100.0 + i * 10,
            )
            for i in range(5)
        ]
        
        # Create samples
        samples = [
            EvaluationSample(
                ground_truth=gt,
                predicted=pred,
                sample_id=f"sample_{i}",
            )
            for i, (gt, pred) in enumerate(zip(ground_truths, predictions))
        ]
        
        # Run batch evaluation
        report = evaluator.evaluate_batch(samples)
        
        # Verify report
        assert report.total_samples == 5
        assert report.samples_evaluated == 5
        assert report.samples_failed == 0
        
        # Should have some errors (every 3rd sample has typos)
        assert report.mean_exact_match_rate < 1.0
        
        # Check field statistics
        assert "name" in report.field_statistics
        assert "id_number" in report.field_statistics
        
        # Check document type statistics
        assert "id_card" in report.document_type_statistics
        assert "passport" in report.document_type_statistics
    
    def test_report_json_serialization(
        self,
        evaluator: OCREvaluator,
        batch_samples: List[EvaluationSample],
    ):
        """Batch report can be saved to JSON."""
        report = evaluator.evaluate_batch(batch_samples)
        
        with tempfile.NamedTemporaryFile(suffix=".json", delete=False) as f:
            path = f.name
        
        try:
            report.to_json_file(path)
            
            # Verify JSON is valid
            with open(path, "r") as f:
                data = json.load(f)
            
            assert "summary" in data
            assert data["summary"]["total_samples"] == 3
        finally:
            Path(path).unlink(missing_ok=True)
