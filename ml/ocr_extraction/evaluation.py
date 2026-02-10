"""
OCR Evaluation Framework for measuring extraction accuracy.

This module provides:
- Character Error Rate (CER) and Word Error Rate (WER) calculations
- Field-level accuracy metrics (precision, recall, F1)
- Batch evaluation with aggregated statistics
- Ground truth annotation support (JSON format)

VE-3045: Create OCR Evaluation Framework
"""

import json
import logging
from dataclasses import dataclass, field
from typing import Dict, List, Optional, Any, Tuple, Union
from pathlib import Path
from enum import Enum
from datetime import datetime, timezone
import statistics

import numpy as np

logger = logging.getLogger(__name__)


# =============================================================================
# Ground Truth Data Structures
# =============================================================================


@dataclass
class GroundTruth:
    """
    Ground truth annotation for a document image.
    
    Attributes:
        image_path: Path to the source image file
        document_type: Type of document (e.g., "id_card", "passport")
        fields: Mapping of field names to expected values
        full_text: Complete text content (optional)
        metadata: Additional annotation metadata
    """
    image_path: str
    document_type: str
    fields: Dict[str, str]
    full_text: Optional[str] = None
    metadata: Dict[str, Any] = field(default_factory=dict)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "image_path": self.image_path,
            "document_type": self.document_type,
            "fields": self.fields,
            "full_text": self.full_text,
            "metadata": self.metadata,
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "GroundTruth":
        """Create from dictionary."""
        return cls(
            image_path=data["image_path"],
            document_type=data["document_type"],
            fields=data.get("fields", {}),
            full_text=data.get("full_text"),
            metadata=data.get("metadata", {}),
        )
    
    @classmethod
    def from_json_file(cls, path: Union[str, Path]) -> "GroundTruth":
        """Load ground truth from JSON file."""
        with open(path, "r", encoding="utf-8") as f:
            data = json.load(f)
        return cls.from_dict(data)
    
    def to_json_file(self, path: Union[str, Path]) -> None:
        """Save ground truth to JSON file."""
        with open(path, "w", encoding="utf-8") as f:
            json.dump(self.to_dict(), f, indent=2, ensure_ascii=False)


@dataclass
class PredictedResult:
    """
    Predicted OCR result for evaluation.
    
    Attributes:
        fields: Mapping of field names to extracted values
        full_text: Complete extracted text (optional)
        confidence: Overall confidence score
        processing_time_ms: Time taken for extraction
    """
    fields: Dict[str, str]
    full_text: Optional[str] = None
    confidence: float = 0.0
    processing_time_ms: float = 0.0
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "fields": self.fields,
            "full_text": self.full_text,
            "confidence": self.confidence,
            "processing_time_ms": self.processing_time_ms,
        }


@dataclass
class EvaluationSample:
    """
    A single evaluation sample pairing ground truth with prediction.
    
    Attributes:
        ground_truth: The annotated ground truth
        predicted: The OCR prediction result
        sample_id: Optional identifier for the sample
    """
    ground_truth: GroundTruth
    predicted: PredictedResult
    sample_id: Optional[str] = None


# =============================================================================
# Metrics Data Structures
# =============================================================================


class MatchType(str, Enum):
    """Field matching result type."""
    EXACT = "exact"           # Exact string match
    PARTIAL = "partial"       # Partial match (substring or similar)
    MISMATCH = "mismatch"     # Values don't match
    MISSING = "missing"       # Field not found in prediction
    EXTRA = "extra"           # Field in prediction but not in ground truth


@dataclass
class FieldResult:
    """Result for a single field comparison."""
    field_name: str
    ground_truth_value: Optional[str]
    predicted_value: Optional[str]
    match_type: MatchType
    character_error_rate: float = 0.0
    word_error_rate: float = 0.0
    similarity_score: float = 0.0
    
    @property
    def is_correct(self) -> bool:
        """Field is considered correct if exact match."""
        return self.match_type == MatchType.EXACT
    
    @property
    def is_usable(self) -> bool:
        """Field is usable if exact or partial match."""
        return self.match_type in (MatchType.EXACT, MatchType.PARTIAL)


@dataclass
class FieldMetrics:
    """
    Aggregated metrics for field-level extraction.
    
    Attributes:
        precision: TP / (TP + FP) - proportion of correct among extracted
        recall: TP / (TP + FN) - proportion of correct among expected
        f1_score: Harmonic mean of precision and recall
        exact_match_rate: Proportion of exact matches
        field_results: Individual field comparison results
    """
    precision: float
    recall: float
    f1_score: float
    exact_match_rate: float
    partial_match_rate: float
    field_results: List[FieldResult] = field(default_factory=list)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "precision": self.precision,
            "recall": self.recall,
            "f1_score": self.f1_score,
            "exact_match_rate": self.exact_match_rate,
            "partial_match_rate": self.partial_match_rate,
            "field_count": len(self.field_results),
            "field_results": [
                {
                    "field_name": fr.field_name,
                    "ground_truth": fr.ground_truth_value,
                    "predicted": fr.predicted_value,
                    "match_type": fr.match_type.value,
                    "cer": fr.character_error_rate,
                    "wer": fr.word_error_rate,
                    "similarity": fr.similarity_score,
                }
                for fr in self.field_results
            ],
        }


@dataclass
class TextMetrics:
    """
    Metrics for full text extraction.
    
    Attributes:
        character_error_rate: CER for complete text
        word_error_rate: WER for complete text
        character_accuracy: 1 - CER
        word_accuracy: 1 - WER
    """
    character_error_rate: float
    word_error_rate: float
    
    @property
    def character_accuracy(self) -> float:
        """Character-level accuracy."""
        return max(0.0, 1.0 - self.character_error_rate)
    
    @property
    def word_accuracy(self) -> float:
        """Word-level accuracy."""
        return max(0.0, 1.0 - self.word_error_rate)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "character_error_rate": self.character_error_rate,
            "word_error_rate": self.word_error_rate,
            "character_accuracy": self.character_accuracy,
            "word_accuracy": self.word_accuracy,
        }


@dataclass
class SampleReport:
    """
    Evaluation report for a single sample.
    
    Combines text-level and field-level metrics.
    """
    sample_id: str
    document_type: str
    text_metrics: Optional[TextMetrics]
    field_metrics: FieldMetrics
    processing_time_ms: float
    confidence: float
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "sample_id": self.sample_id,
            "document_type": self.document_type,
            "text_metrics": self.text_metrics.to_dict() if self.text_metrics else None,
            "field_metrics": self.field_metrics.to_dict(),
            "processing_time_ms": self.processing_time_ms,
            "confidence": self.confidence,
        }


@dataclass
class BatchReport:
    """
    Aggregated evaluation report for a batch of samples.
    
    Contains summary statistics and per-sample details.
    """
    total_samples: int
    samples_evaluated: int
    samples_failed: int
    
    # Aggregate text metrics
    mean_cer: float
    mean_wer: float
    std_cer: float
    std_wer: float
    
    # Aggregate field metrics
    mean_precision: float
    mean_recall: float
    mean_f1: float
    mean_exact_match_rate: float
    
    # Per-field breakdown
    field_statistics: Dict[str, Dict[str, float]]
    
    # Per-document-type breakdown
    document_type_statistics: Dict[str, Dict[str, float]]
    
    # Individual sample reports
    sample_reports: List[SampleReport] = field(default_factory=list)
    
    # Processing statistics
    total_processing_time_ms: float = 0.0
    mean_processing_time_ms: float = 0.0
    
    # Evaluation metadata
    evaluation_timestamp: str = field(
        default_factory=lambda: datetime.now(timezone.utc).isoformat()
    )
    evaluator_version: str = "1.0.0"
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "summary": {
                "total_samples": self.total_samples,
                "samples_evaluated": self.samples_evaluated,
                "samples_failed": self.samples_failed,
                "evaluation_timestamp": self.evaluation_timestamp,
                "evaluator_version": self.evaluator_version,
            },
            "text_metrics": {
                "mean_cer": self.mean_cer,
                "mean_wer": self.mean_wer,
                "std_cer": self.std_cer,
                "std_wer": self.std_wer,
            },
            "field_metrics": {
                "mean_precision": self.mean_precision,
                "mean_recall": self.mean_recall,
                "mean_f1": self.mean_f1,
                "mean_exact_match_rate": self.mean_exact_match_rate,
            },
            "processing_stats": {
                "total_time_ms": self.total_processing_time_ms,
                "mean_time_ms": self.mean_processing_time_ms,
            },
            "field_statistics": self.field_statistics,
            "document_type_statistics": self.document_type_statistics,
            "sample_reports": [sr.to_dict() for sr in self.sample_reports],
        }
    
    def to_json_file(self, path: Union[str, Path]) -> None:
        """Save report to JSON file."""
        with open(path, "w", encoding="utf-8") as f:
            json.dump(self.to_dict(), f, indent=2, ensure_ascii=False)
    
    def summary_string(self) -> str:
        """Generate a human-readable summary."""
        lines = [
            "=" * 60,
            "OCR EVALUATION REPORT",
            "=" * 60,
            f"Samples: {self.samples_evaluated}/{self.total_samples} evaluated "
            f"({self.samples_failed} failed)",
            "",
            "TEXT METRICS:",
            f"  Character Error Rate (CER): {self.mean_cer:.4f} ± {self.std_cer:.4f}",
            f"  Word Error Rate (WER):      {self.mean_wer:.4f} ± {self.std_wer:.4f}",
            "",
            "FIELD METRICS:",
            f"  Precision:        {self.mean_precision:.4f}",
            f"  Recall:           {self.mean_recall:.4f}",
            f"  F1 Score:         {self.mean_f1:.4f}",
            f"  Exact Match Rate: {self.mean_exact_match_rate:.4f}",
            "",
            "PROCESSING:",
            f"  Total Time: {self.total_processing_time_ms:.2f} ms",
            f"  Mean Time:  {self.mean_processing_time_ms:.2f} ms/sample",
            "=" * 60,
        ]
        return "\n".join(lines)


# =============================================================================
# Edit Distance Calculation
# =============================================================================


def levenshtein_distance(s1: str, s2: str) -> int:
    """
    Calculate Levenshtein (edit) distance between two strings.
    
    Uses dynamic programming with O(min(len(s1), len(s2))) space.
    
    Args:
        s1: First string
        s2: Second string
        
    Returns:
        Minimum number of single-character edits to transform s1 to s2
    """
    if len(s1) < len(s2):
        return levenshtein_distance(s2, s1)
    
    if len(s2) == 0:
        return len(s1)
    
    # Use only two rows for space efficiency
    previous_row = list(range(len(s2) + 1))
    current_row = [0] * (len(s2) + 1)
    
    for i, c1 in enumerate(s1):
        current_row[0] = i + 1
        for j, c2 in enumerate(s2):
            # Costs for operations
            insertions = previous_row[j + 1] + 1
            deletions = current_row[j] + 1
            substitutions = previous_row[j] + (0 if c1 == c2 else 1)
            current_row[j + 1] = min(insertions, deletions, substitutions)
        
        previous_row, current_row = current_row, previous_row
    
    return previous_row[len(s2)]


def word_levenshtein_distance(words1: List[str], words2: List[str]) -> int:
    """
    Calculate Levenshtein distance at the word level.
    
    Args:
        words1: First list of words
        words2: Second list of words
        
    Returns:
        Minimum number of word-level edits
    """
    if len(words1) < len(words2):
        return word_levenshtein_distance(words2, words1)
    
    if len(words2) == 0:
        return len(words1)
    
    previous_row = list(range(len(words2) + 1))
    current_row = [0] * (len(words2) + 1)
    
    for i, w1 in enumerate(words1):
        current_row[0] = i + 1
        for j, w2 in enumerate(words2):
            insertions = previous_row[j + 1] + 1
            deletions = current_row[j] + 1
            substitutions = previous_row[j] + (0 if w1 == w2 else 1)
            current_row[j + 1] = min(insertions, deletions, substitutions)
        
        previous_row, current_row = current_row, previous_row
    
    return previous_row[len(words2)]


# =============================================================================
# OCR Evaluator
# =============================================================================


class OCREvaluator:
    """
    OCR evaluation framework for measuring extraction accuracy.
    
    Supports:
    - Character Error Rate (CER) calculation
    - Word Error Rate (WER) calculation
    - Field-level accuracy (precision, recall, F1)
    - Batch evaluation with statistics
    - Configurable matching thresholds
    
    Example:
        evaluator = OCREvaluator()
        
        # Single text comparison
        cer = evaluator.character_error_rate("hello", "helo")
        wer = evaluator.word_error_rate("hello world", "hello word")
        
        # Field comparison
        predicted = {"name": "John Smith", "dob": "1990-01-01"}
        ground_truth = {"name": "JOHN SMITH", "dob": "1990-01-01", "id": "123"}
        metrics = evaluator.field_accuracy(predicted, ground_truth)
        
        # Batch evaluation
        samples = [EvaluationSample(...), ...]
        report = evaluator.evaluate_batch(samples)
    """
    
    def __init__(
        self,
        case_sensitive: bool = False,
        partial_match_threshold: float = 0.8,
        normalize_whitespace: bool = True,
    ):
        """
        Initialize the evaluator.
        
        Args:
            case_sensitive: Whether to use case-sensitive comparison
            partial_match_threshold: Similarity threshold for partial matches (0-1)
            normalize_whitespace: Whether to normalize whitespace in text
        """
        self.case_sensitive = case_sensitive
        self.partial_match_threshold = partial_match_threshold
        self.normalize_whitespace = normalize_whitespace
    
    def _normalize_text(self, text: str) -> str:
        """Normalize text for comparison."""
        if text is None:
            return ""
        
        result = text
        
        if not self.case_sensitive:
            result = result.lower()
        
        if self.normalize_whitespace:
            # Normalize all whitespace to single spaces
            result = " ".join(result.split())
        
        return result
    
    def _tokenize(self, text: str) -> List[str]:
        """Tokenize text into words."""
        normalized = self._normalize_text(text)
        return normalized.split()
    
    def character_error_rate(self, predicted: str, ground_truth: str) -> float:
        """
        Calculate Character Error Rate (CER).
        
        CER = (Substitutions + Insertions + Deletions) / Reference Length
        
        Args:
            predicted: OCR output text
            ground_truth: Expected text
            
        Returns:
            CER value (0.0 = perfect, can be > 1.0 for very poor predictions)
        """
        pred_normalized = self._normalize_text(predicted)
        gt_normalized = self._normalize_text(ground_truth)
        
        if len(gt_normalized) == 0:
            # If ground truth is empty, CER is 0 if prediction is also empty
            return 0.0 if len(pred_normalized) == 0 else float(len(pred_normalized))
        
        distance = levenshtein_distance(pred_normalized, gt_normalized)
        return distance / len(gt_normalized)
    
    def word_error_rate(self, predicted: str, ground_truth: str) -> float:
        """
        Calculate Word Error Rate (WER).
        
        WER = (Substitutions + Insertions + Deletions) / Reference Word Count
        
        Args:
            predicted: OCR output text
            ground_truth: Expected text
            
        Returns:
            WER value (0.0 = perfect, can be > 1.0 for very poor predictions)
        """
        pred_words = self._tokenize(predicted)
        gt_words = self._tokenize(ground_truth)
        
        if len(gt_words) == 0:
            return 0.0 if len(pred_words) == 0 else float(len(pred_words))
        
        distance = word_levenshtein_distance(pred_words, gt_words)
        return distance / len(gt_words)
    
    def similarity_score(self, predicted: str, ground_truth: str) -> float:
        """
        Calculate similarity score between two strings.
        
        Returns 1.0 - normalized_edit_distance, clamped to [0, 1].
        
        Args:
            predicted: OCR output text
            ground_truth: Expected text
            
        Returns:
            Similarity score (0.0 = completely different, 1.0 = identical)
        """
        pred_normalized = self._normalize_text(predicted)
        gt_normalized = self._normalize_text(ground_truth)
        
        if len(pred_normalized) == 0 and len(gt_normalized) == 0:
            return 1.0
        
        max_len = max(len(pred_normalized), len(gt_normalized))
        if max_len == 0:
            return 1.0
        
        distance = levenshtein_distance(pred_normalized, gt_normalized)
        return max(0.0, 1.0 - distance / max_len)
    
    def _compare_field(
        self,
        field_name: str,
        predicted_value: Optional[str],
        ground_truth_value: Optional[str],
    ) -> FieldResult:
        """Compare a single field."""
        # Handle missing cases
        if ground_truth_value is None and predicted_value is None:
            return FieldResult(
                field_name=field_name,
                ground_truth_value=None,
                predicted_value=None,
                match_type=MatchType.EXACT,
                similarity_score=1.0,
            )
        
        if ground_truth_value is None:
            return FieldResult(
                field_name=field_name,
                ground_truth_value=None,
                predicted_value=predicted_value,
                match_type=MatchType.EXTRA,
                similarity_score=0.0,
            )
        
        if predicted_value is None:
            return FieldResult(
                field_name=field_name,
                ground_truth_value=ground_truth_value,
                predicted_value=None,
                match_type=MatchType.MISSING,
                similarity_score=0.0,
            )
        
        # Calculate metrics
        cer = self.character_error_rate(predicted_value, ground_truth_value)
        wer = self.word_error_rate(predicted_value, ground_truth_value)
        similarity = self.similarity_score(predicted_value, ground_truth_value)
        
        # Determine match type
        if similarity >= 1.0:
            match_type = MatchType.EXACT
        elif similarity >= self.partial_match_threshold:
            match_type = MatchType.PARTIAL
        else:
            match_type = MatchType.MISMATCH
        
        return FieldResult(
            field_name=field_name,
            ground_truth_value=ground_truth_value,
            predicted_value=predicted_value,
            match_type=match_type,
            character_error_rate=cer,
            word_error_rate=wer,
            similarity_score=similarity,
        )
    
    def field_accuracy(
        self,
        predicted: Dict[str, str],
        ground_truth: Dict[str, str],
    ) -> FieldMetrics:
        """
        Calculate field-level accuracy metrics.
        
        Args:
            predicted: Dictionary of extracted field values
            ground_truth: Dictionary of expected field values
            
        Returns:
            FieldMetrics with precision, recall, F1, and per-field results
        """
        field_results: List[FieldResult] = []
        
        # Get all field names from both
        all_fields = set(ground_truth.keys()) | set(predicted.keys())
        
        for field_name in all_fields:
            gt_value = ground_truth.get(field_name)
            pred_value = predicted.get(field_name)
            
            result = self._compare_field(field_name, pred_value, gt_value)
            field_results.append(result)
        
        # Calculate precision and recall
        # TP = exact matches on expected fields
        # FP = extra fields or mismatches
        # FN = missing fields
        
        expected_fields = [fr for fr in field_results if fr.ground_truth_value is not None]
        
        true_positives = sum(1 for fr in expected_fields if fr.match_type == MatchType.EXACT)
        partial_matches = sum(1 for fr in expected_fields if fr.match_type == MatchType.PARTIAL)
        false_negatives = sum(
            1 for fr in expected_fields 
            if fr.match_type in (MatchType.MISSING, MatchType.MISMATCH)
        )
        false_positives = sum(1 for fr in field_results if fr.match_type == MatchType.EXTRA)
        
        # Calculate metrics
        total_expected = len(expected_fields)
        total_predicted = sum(1 for fr in field_results if fr.predicted_value is not None)
        
        precision = true_positives / total_predicted if total_predicted > 0 else 0.0
        recall = true_positives / total_expected if total_expected > 0 else 0.0
        f1_score = (
            2 * precision * recall / (precision + recall)
            if (precision + recall) > 0
            else 0.0
        )
        
        exact_match_rate = true_positives / total_expected if total_expected > 0 else 0.0
        partial_match_rate = (
            (true_positives + partial_matches) / total_expected
            if total_expected > 0
            else 0.0
        )
        
        return FieldMetrics(
            precision=precision,
            recall=recall,
            f1_score=f1_score,
            exact_match_rate=exact_match_rate,
            partial_match_rate=partial_match_rate,
            field_results=field_results,
        )
    
    def evaluate_sample(
        self,
        sample: EvaluationSample,
    ) -> SampleReport:
        """
        Evaluate a single sample.
        
        Args:
            sample: EvaluationSample with ground truth and prediction
            
        Returns:
            SampleReport with all metrics
        """
        gt = sample.ground_truth
        pred = sample.predicted
        
        # Calculate text metrics if full text is available
        text_metrics = None
        if gt.full_text is not None and pred.full_text is not None:
            cer = self.character_error_rate(pred.full_text, gt.full_text)
            wer = self.word_error_rate(pred.full_text, gt.full_text)
            text_metrics = TextMetrics(
                character_error_rate=cer,
                word_error_rate=wer,
            )
        
        # Calculate field metrics
        field_metrics = self.field_accuracy(pred.fields, gt.fields)
        
        sample_id = sample.sample_id or gt.image_path
        
        return SampleReport(
            sample_id=sample_id,
            document_type=gt.document_type,
            text_metrics=text_metrics,
            field_metrics=field_metrics,
            processing_time_ms=pred.processing_time_ms,
            confidence=pred.confidence,
        )
    
    def evaluate_batch(
        self,
        samples: List[EvaluationSample],
    ) -> BatchReport:
        """
        Evaluate a batch of samples and generate aggregated statistics.
        
        Args:
            samples: List of evaluation samples
            
        Returns:
            BatchReport with aggregate metrics and per-sample details
        """
        if not samples:
            return BatchReport(
                total_samples=0,
                samples_evaluated=0,
                samples_failed=0,
                mean_cer=0.0,
                mean_wer=0.0,
                std_cer=0.0,
                std_wer=0.0,
                mean_precision=0.0,
                mean_recall=0.0,
                mean_f1=0.0,
                mean_exact_match_rate=0.0,
                field_statistics={},
                document_type_statistics={},
            )
        
        sample_reports: List[SampleReport] = []
        failed_count = 0
        
        # Evaluate each sample
        for sample in samples:
            try:
                report = self.evaluate_sample(sample)
                sample_reports.append(report)
            except Exception as e:
                logger.error(f"Failed to evaluate sample: {e}")
                failed_count += 1
        
        if not sample_reports:
            return BatchReport(
                total_samples=len(samples),
                samples_evaluated=0,
                samples_failed=failed_count,
                mean_cer=0.0,
                mean_wer=0.0,
                std_cer=0.0,
                std_wer=0.0,
                mean_precision=0.0,
                mean_recall=0.0,
                mean_f1=0.0,
                mean_exact_match_rate=0.0,
                field_statistics={},
                document_type_statistics={},
            )
        
        # Aggregate text metrics
        cers = [
            sr.text_metrics.character_error_rate
            for sr in sample_reports
            if sr.text_metrics is not None
        ]
        wers = [
            sr.text_metrics.word_error_rate
            for sr in sample_reports
            if sr.text_metrics is not None
        ]
        
        mean_cer = statistics.mean(cers) if cers else 0.0
        mean_wer = statistics.mean(wers) if wers else 0.0
        std_cer = statistics.stdev(cers) if len(cers) > 1 else 0.0
        std_wer = statistics.stdev(wers) if len(wers) > 1 else 0.0
        
        # Aggregate field metrics
        precisions = [sr.field_metrics.precision for sr in sample_reports]
        recalls = [sr.field_metrics.recall for sr in sample_reports]
        f1s = [sr.field_metrics.f1_score for sr in sample_reports]
        exact_rates = [sr.field_metrics.exact_match_rate for sr in sample_reports]
        
        mean_precision = statistics.mean(precisions)
        mean_recall = statistics.mean(recalls)
        mean_f1 = statistics.mean(f1s)
        mean_exact_match_rate = statistics.mean(exact_rates)
        
        # Per-field statistics
        field_statistics = self._compute_field_statistics(sample_reports)
        
        # Per-document-type statistics
        document_type_statistics = self._compute_document_type_statistics(sample_reports)
        
        # Processing time statistics
        total_time = sum(sr.processing_time_ms for sr in sample_reports)
        mean_time = total_time / len(sample_reports)
        
        return BatchReport(
            total_samples=len(samples),
            samples_evaluated=len(sample_reports),
            samples_failed=failed_count,
            mean_cer=mean_cer,
            mean_wer=mean_wer,
            std_cer=std_cer,
            std_wer=std_wer,
            mean_precision=mean_precision,
            mean_recall=mean_recall,
            mean_f1=mean_f1,
            mean_exact_match_rate=mean_exact_match_rate,
            field_statistics=field_statistics,
            document_type_statistics=document_type_statistics,
            sample_reports=sample_reports,
            total_processing_time_ms=total_time,
            mean_processing_time_ms=mean_time,
        )
    
    def _compute_field_statistics(
        self,
        sample_reports: List[SampleReport],
    ) -> Dict[str, Dict[str, float]]:
        """Compute per-field statistics across all samples."""
        field_data: Dict[str, Dict[str, List[float]]] = {}
        
        for sr in sample_reports:
            for fr in sr.field_metrics.field_results:
                if fr.field_name not in field_data:
                    field_data[fr.field_name] = {
                        "exact_matches": [],
                        "partial_matches": [],
                        "cer": [],
                        "wer": [],
                        "similarity": [],
                    }
                
                data = field_data[fr.field_name]
                data["exact_matches"].append(1.0 if fr.match_type == MatchType.EXACT else 0.0)
                data["partial_matches"].append(1.0 if fr.is_usable else 0.0)
                
                if fr.ground_truth_value is not None and fr.predicted_value is not None:
                    data["cer"].append(fr.character_error_rate)
                    data["wer"].append(fr.word_error_rate)
                    data["similarity"].append(fr.similarity_score)
        
        # Aggregate
        field_statistics: Dict[str, Dict[str, float]] = {}
        for field_name, data in field_data.items():
            field_statistics[field_name] = {
                "exact_match_rate": statistics.mean(data["exact_matches"]) if data["exact_matches"] else 0.0,
                "partial_match_rate": statistics.mean(data["partial_matches"]) if data["partial_matches"] else 0.0,
                "mean_cer": statistics.mean(data["cer"]) if data["cer"] else 0.0,
                "mean_wer": statistics.mean(data["wer"]) if data["wer"] else 0.0,
                "mean_similarity": statistics.mean(data["similarity"]) if data["similarity"] else 0.0,
                "sample_count": len(data["exact_matches"]),
            }
        
        return field_statistics
    
    def _compute_document_type_statistics(
        self,
        sample_reports: List[SampleReport],
    ) -> Dict[str, Dict[str, float]]:
        """Compute per-document-type statistics."""
        type_data: Dict[str, List[SampleReport]] = {}
        
        for sr in sample_reports:
            doc_type = sr.document_type
            if doc_type not in type_data:
                type_data[doc_type] = []
            type_data[doc_type].append(sr)
        
        statistics_by_type: Dict[str, Dict[str, float]] = {}
        for doc_type, reports in type_data.items():
            cers = [
                r.text_metrics.character_error_rate
                for r in reports
                if r.text_metrics is not None
            ]
            wers = [
                r.text_metrics.word_error_rate
                for r in reports
                if r.text_metrics is not None
            ]
            
            statistics_by_type[doc_type] = {
                "sample_count": len(reports),
                "mean_cer": statistics.mean(cers) if cers else 0.0,
                "mean_wer": statistics.mean(wers) if wers else 0.0,
                "mean_precision": statistics.mean([r.field_metrics.precision for r in reports]),
                "mean_recall": statistics.mean([r.field_metrics.recall for r in reports]),
                "mean_f1": statistics.mean([r.field_metrics.f1_score for r in reports]),
                "mean_exact_match_rate": statistics.mean(
                    [r.field_metrics.exact_match_rate for r in reports]
                ),
            }
        
        return statistics_by_type


# =============================================================================
# Ground Truth Dataset Utilities
# =============================================================================


class GroundTruthDataset:
    """
    Collection of ground truth annotations for batch evaluation.
    
    Supports loading from JSON files or directories.
    """
    
    def __init__(self, ground_truths: Optional[List[GroundTruth]] = None):
        """Initialize with optional list of ground truths."""
        self.ground_truths = ground_truths or []
    
    def __len__(self) -> int:
        return len(self.ground_truths)
    
    def __iter__(self):
        return iter(self.ground_truths)
    
    def add(self, gt: GroundTruth) -> None:
        """Add a ground truth annotation."""
        self.ground_truths.append(gt)
    
    @classmethod
    def from_json_file(cls, path: Union[str, Path]) -> "GroundTruthDataset":
        """
        Load dataset from a JSON file.
        
        Expected format:
        {
            "annotations": [
                {"image_path": "...", "document_type": "...", "fields": {...}},
                ...
            ]
        }
        """
        with open(path, "r", encoding="utf-8") as f:
            data = json.load(f)
        
        ground_truths = []
        for annotation in data.get("annotations", []):
            gt = GroundTruth.from_dict(annotation)
            ground_truths.append(gt)
        
        return cls(ground_truths)
    
    @classmethod
    def from_directory(
        cls,
        directory: Union[str, Path],
        pattern: str = "*.json",
    ) -> "GroundTruthDataset":
        """Load dataset from individual JSON files in a directory."""
        directory = Path(directory)
        ground_truths = []
        
        for json_path in directory.glob(pattern):
            try:
                gt = GroundTruth.from_json_file(json_path)
                ground_truths.append(gt)
            except Exception as e:
                logger.warning(f"Failed to load {json_path}: {e}")
        
        return cls(ground_truths)
    
    def to_json_file(self, path: Union[str, Path]) -> None:
        """Save dataset to a JSON file."""
        data = {
            "annotations": [gt.to_dict() for gt in self.ground_truths]
        }
        with open(path, "w", encoding="utf-8") as f:
            json.dump(data, f, indent=2, ensure_ascii=False)
    
    def filter_by_document_type(self, document_type: str) -> "GroundTruthDataset":
        """Return filtered dataset by document type."""
        filtered = [gt for gt in self.ground_truths if gt.document_type == document_type]
        return GroundTruthDataset(filtered)
    
    def split(
        self,
        train_ratio: float = 0.8,
        seed: Optional[int] = None,
    ) -> Tuple["GroundTruthDataset", "GroundTruthDataset"]:
        """Split dataset into train and test sets."""
        import random
        
        if seed is not None:
            random.seed(seed)
        
        indices = list(range(len(self.ground_truths)))
        random.shuffle(indices)
        
        split_idx = int(len(indices) * train_ratio)
        train_indices = indices[:split_idx]
        test_indices = indices[split_idx:]
        
        train_gts = [self.ground_truths[i] for i in train_indices]
        test_gts = [self.ground_truths[i] for i in test_indices]
        
        return GroundTruthDataset(train_gts), GroundTruthDataset(test_gts)


# =============================================================================
# Convenience Functions
# =============================================================================


def create_evaluation_sample(
    image_path: str,
    document_type: str,
    ground_truth_fields: Dict[str, str],
    predicted_fields: Dict[str, str],
    ground_truth_text: Optional[str] = None,
    predicted_text: Optional[str] = None,
    sample_id: Optional[str] = None,
) -> EvaluationSample:
    """
    Create an evaluation sample from raw data.
    
    Convenience function for building evaluation samples.
    """
    gt = GroundTruth(
        image_path=image_path,
        document_type=document_type,
        fields=ground_truth_fields,
        full_text=ground_truth_text,
    )
    
    pred = PredictedResult(
        fields=predicted_fields,
        full_text=predicted_text,
    )
    
    return EvaluationSample(
        ground_truth=gt,
        predicted=pred,
        sample_id=sample_id,
    )


def quick_evaluate(
    predicted_text: str,
    ground_truth_text: str,
    case_sensitive: bool = False,
) -> Dict[str, float]:
    """
    Quick text-level evaluation.
    
    Returns CER, WER, and similarity score.
    """
    evaluator = OCREvaluator(case_sensitive=case_sensitive)
    
    return {
        "cer": evaluator.character_error_rate(predicted_text, ground_truth_text),
        "wer": evaluator.word_error_rate(predicted_text, ground_truth_text),
        "similarity": evaluator.similarity_score(predicted_text, ground_truth_text),
    }
