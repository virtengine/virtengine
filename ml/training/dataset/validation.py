"""
Dataset validation for schema and label quality.

This module provides:
- Schema validation against data contract
- Label anomaly detection
- Data quality checks
- Validation reporting
"""

import hashlib
import json
import logging
import re
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from pathlib import Path
from typing import Any, Callable, Dict, List, Optional, Set, Tuple

import numpy as np

from ml.training.config import DocumentType
from ml.training.dataset.ingestion import (
    Dataset,
    DatasetSplit,
    IdentitySample,
)

logger = logging.getLogger(__name__)


class ValidationLevel(str, Enum):
    """Validation severity levels."""
    ERROR = "error"
    WARNING = "warning"
    INFO = "info"


class ValidationType(str, Enum):
    """Types of validation checks."""
    SCHEMA = "schema"
    LABEL = "label"
    DATA_QUALITY = "data_quality"
    SPLIT = "split"
    INTEGRITY = "integrity"


@dataclass
class ValidationIssue:
    """A single validation issue."""
    
    level: ValidationLevel
    validation_type: ValidationType
    message: str
    sample_id: Optional[str] = None
    field: Optional[str] = None
    value: Optional[Any] = None
    expected: Optional[Any] = None
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "level": self.level.value,
            "type": self.validation_type.value,
            "message": self.message,
            "sample_id": self.sample_id,
            "field": self.field,
            "value": str(self.value) if self.value is not None else None,
            "expected": str(self.expected) if self.expected is not None else None,
        }


@dataclass
class ValidationReport:
    """Complete validation report."""
    
    # Summary
    valid: bool = True
    total_issues: int = 0
    error_count: int = 0
    warning_count: int = 0
    info_count: int = 0
    
    # Details
    issues: List[ValidationIssue] = field(default_factory=list)
    
    # Statistics
    samples_validated: int = 0
    validation_time_seconds: float = 0.0
    
    # Metadata
    schema_version: str = ""
    validated_at: float = field(default_factory=lambda: datetime.now().timestamp())
    
    def add_issue(self, issue: ValidationIssue) -> None:
        """Add a validation issue."""
        self.issues.append(issue)
        self.total_issues += 1
        
        if issue.level == ValidationLevel.ERROR:
            self.error_count += 1
            self.valid = False
        elif issue.level == ValidationLevel.WARNING:
            self.warning_count += 1
        else:
            self.info_count += 1
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "valid": self.valid,
            "total_issues": self.total_issues,
            "error_count": self.error_count,
            "warning_count": self.warning_count,
            "info_count": self.info_count,
            "samples_validated": self.samples_validated,
            "validation_time_seconds": self.validation_time_seconds,
            "schema_version": self.schema_version,
            "validated_at": self.validated_at,
            "validated_at_iso": datetime.fromtimestamp(self.validated_at).isoformat(),
            "issues": [i.to_dict() for i in self.issues],
        }
    
    def save(self, path: str) -> None:
        """Save report to file."""
        file_path = Path(path)
        file_path.parent.mkdir(parents=True, exist_ok=True)
        
        with open(file_path, "w") as f:
            json.dump(self.to_dict(), f, indent=2)
    
    def print_summary(self) -> str:
        """Generate a printable summary."""
        status = "✓ VALID" if self.valid else "✗ INVALID"
        lines = [
            f"Validation Report: {status}",
            f"  Samples validated: {self.samples_validated}",
            f"  Total issues: {self.total_issues}",
            f"  Errors: {self.error_count}",
            f"  Warnings: {self.warning_count}",
            f"  Info: {self.info_count}",
        ]
        
        if self.issues:
            lines.append("\nTop issues:")
            for issue in self.issues[:10]:
                lines.append(f"  [{issue.level.value.upper()}] {issue.message}")
        
        return "\n".join(lines)


@dataclass
class SchemaField:
    """Definition of a schema field."""
    
    name: str
    field_type: str  # "float", "int", "bool", "str", "list", "dict"
    required: bool = True
    nullable: bool = True
    
    # Constraints
    min_value: Optional[float] = None
    max_value: Optional[float] = None
    allowed_values: Optional[Set[Any]] = None
    pattern: Optional[str] = None
    
    # Nested schema
    nested_schema: Optional[List["SchemaField"]] = None


# VEID Dataset Schema
VEID_SCHEMA_V1 = [
    SchemaField("sample_id", "str", required=True, nullable=False),
    SchemaField("trust_score", "float", required=True, min_value=0.0, max_value=100.0),
    SchemaField("is_genuine", "bool", required=True),
    SchemaField("fraud_type", "str", required=False, nullable=True, allowed_values={
        None, "printed_photo", "screen_replay", "document_tampering",
        "identity_theft", "synthetic_identity", "deepfake", "mask_attack",
        "missing_face",
    }),
    SchemaField("face_detected", "bool", required=True),
    SchemaField("face_confidence", "float", required=True, min_value=0.0, max_value=1.0),
    SchemaField("document_quality_score", "float", required=True, min_value=0.0, max_value=1.0),
    SchemaField("ocr_success", "bool", required=True),
    SchemaField("ocr_confidence", "float", required=True, min_value=0.0, max_value=1.0),
]


class SchemaValidator:
    """Validates samples against a defined schema."""
    
    def __init__(
        self,
        schema: Optional[List[SchemaField]] = None,
        schema_version: str = "1.0.0",
    ):
        """
        Initialize schema validator.
        
        Args:
            schema: List of schema field definitions
            schema_version: Version string for the schema
        """
        self.schema = schema or VEID_SCHEMA_V1
        self.schema_version = schema_version
        self._field_map = {f.name: f for f in self.schema}
    
    def validate_sample(self, sample: IdentitySample) -> List[ValidationIssue]:
        """Validate a single sample against the schema."""
        issues = []
        
        for field_def in self.schema:
            value = self._get_field_value(sample, field_def.name)
            field_issues = self._validate_field(sample.sample_id, field_def, value)
            issues.extend(field_issues)
        
        return issues
    
    def _get_field_value(self, sample: IdentitySample, field_name: str) -> Any:
        """Get field value from sample."""
        if hasattr(sample, field_name):
            return getattr(sample, field_name)
        
        # Check nested objects
        if sample.document_info and hasattr(sample.document_info, field_name):
            return getattr(sample.document_info, field_name)
        
        if sample.capture_metadata and hasattr(sample.capture_metadata, field_name):
            return getattr(sample.capture_metadata, field_name)
        
        if field_name in sample.annotations:
            return sample.annotations[field_name]
        
        return None
    
    def _validate_field(
        self,
        sample_id: str,
        field_def: SchemaField,
        value: Any,
    ) -> List[ValidationIssue]:
        """Validate a single field value."""
        issues = []
        
        # Check required
        if field_def.required and value is None:
            if not field_def.nullable:
                issues.append(ValidationIssue(
                    level=ValidationLevel.ERROR,
                    validation_type=ValidationType.SCHEMA,
                    message=f"Required field '{field_def.name}' is missing",
                    sample_id=sample_id,
                    field=field_def.name,
                ))
            return issues
        
        if value is None:
            return issues
        
        # Check type
        type_valid = self._check_type(value, field_def.field_type)
        if not type_valid:
            issues.append(ValidationIssue(
                level=ValidationLevel.ERROR,
                validation_type=ValidationType.SCHEMA,
                message=f"Field '{field_def.name}' has wrong type, expected {field_def.field_type}",
                sample_id=sample_id,
                field=field_def.name,
                value=value,
                expected=field_def.field_type,
            ))
            return issues
        
        # Check range
        if field_def.min_value is not None and value < field_def.min_value:
            issues.append(ValidationIssue(
                level=ValidationLevel.ERROR,
                validation_type=ValidationType.SCHEMA,
                message=f"Field '{field_def.name}' value {value} is below minimum {field_def.min_value}",
                sample_id=sample_id,
                field=field_def.name,
                value=value,
                expected=f">= {field_def.min_value}",
            ))
        
        if field_def.max_value is not None and value > field_def.max_value:
            issues.append(ValidationIssue(
                level=ValidationLevel.ERROR,
                validation_type=ValidationType.SCHEMA,
                message=f"Field '{field_def.name}' value {value} is above maximum {field_def.max_value}",
                sample_id=sample_id,
                field=field_def.name,
                value=value,
                expected=f"<= {field_def.max_value}",
            ))
        
        # Check allowed values
        if field_def.allowed_values is not None and value not in field_def.allowed_values:
            issues.append(ValidationIssue(
                level=ValidationLevel.ERROR,
                validation_type=ValidationType.SCHEMA,
                message=f"Field '{field_def.name}' has invalid value",
                sample_id=sample_id,
                field=field_def.name,
                value=value,
                expected=f"one of {field_def.allowed_values}",
            ))
        
        # Check pattern
        if field_def.pattern is not None and isinstance(value, str):
            if not re.match(field_def.pattern, value):
                issues.append(ValidationIssue(
                    level=ValidationLevel.ERROR,
                    validation_type=ValidationType.SCHEMA,
                    message=f"Field '{field_def.name}' does not match pattern",
                    sample_id=sample_id,
                    field=field_def.name,
                    value=value,
                    expected=field_def.pattern,
                ))
        
        return issues
    
    def _check_type(self, value: Any, expected_type: str) -> bool:
        """Check if value matches expected type."""
        type_map = {
            "float": (int, float),
            "int": int,
            "bool": bool,
            "str": str,
            "list": list,
            "dict": dict,
        }
        expected = type_map.get(expected_type, object)
        return isinstance(value, expected)


class LabelAnomalyDetector:
    """Detects anomalies in labels."""
    
    def __init__(
        self,
        score_std_threshold: float = 2.0,
        min_genuine_ratio: float = 0.3,
        max_genuine_ratio: float = 0.95,
    ):
        """
        Initialize anomaly detector.
        
        Args:
            score_std_threshold: Z-score threshold for outliers
            min_genuine_ratio: Minimum expected genuine ratio
            max_genuine_ratio: Maximum expected genuine ratio
        """
        self.score_std_threshold = score_std_threshold
        self.min_genuine_ratio = min_genuine_ratio
        self.max_genuine_ratio = max_genuine_ratio
    
    def detect_anomalies(
        self,
        samples: List[IdentitySample],
    ) -> List[ValidationIssue]:
        """Detect label anomalies in samples."""
        issues = []
        
        if len(samples) < 10:
            return issues
        
        # Collect scores
        scores = np.array([s.trust_score for s in samples])
        genuine_flags = np.array([s.is_genuine for s in samples])
        
        # Check score distribution
        mean_score = np.mean(scores)
        std_score = np.std(scores)
        
        if std_score > 0:
            z_scores = (scores - mean_score) / std_score
            
            for i, (sample, z) in enumerate(zip(samples, z_scores)):
                if abs(z) > self.score_std_threshold:
                    issues.append(ValidationIssue(
                        level=ValidationLevel.WARNING,
                        validation_type=ValidationType.LABEL,
                        message=f"Trust score outlier (z={z:.2f})",
                        sample_id=sample.sample_id,
                        field="trust_score",
                        value=sample.trust_score,
                        expected=f"within {self.score_std_threshold} std of mean ({mean_score:.2f})",
                    ))
        
        # Check genuine ratio
        genuine_ratio = np.mean(genuine_flags)
        
        if genuine_ratio < self.min_genuine_ratio:
            issues.append(ValidationIssue(
                level=ValidationLevel.WARNING,
                validation_type=ValidationType.LABEL,
                message=f"Genuine ratio ({genuine_ratio:.2%}) is suspiciously low",
                expected=f">= {self.min_genuine_ratio:.2%}",
            ))
        
        if genuine_ratio > self.max_genuine_ratio:
            issues.append(ValidationIssue(
                level=ValidationLevel.INFO,
                validation_type=ValidationType.LABEL,
                message=f"Genuine ratio ({genuine_ratio:.2%}) is very high",
                expected=f"<= {self.max_genuine_ratio:.2%}",
            ))
        
        # Check score-label consistency
        genuine_scores = scores[genuine_flags]
        fraud_scores = scores[~genuine_flags]
        
        if len(genuine_scores) > 0 and len(fraud_scores) > 0:
            genuine_mean = np.mean(genuine_scores)
            fraud_mean = np.mean(fraud_scores)
            
            if genuine_mean < fraud_mean:
                issues.append(ValidationIssue(
                    level=ValidationLevel.ERROR,
                    validation_type=ValidationType.LABEL,
                    message=(
                        f"Genuine samples have lower mean score ({genuine_mean:.2f}) "
                        f"than fraud samples ({fraud_mean:.2f})"
                    ),
                ))
        
        # Check individual sample consistency
        for sample in samples:
            sample_issues = self._check_sample_consistency(sample)
            issues.extend(sample_issues)
        
        return issues
    
    def _check_sample_consistency(self, sample: IdentitySample) -> List[ValidationIssue]:
        """Check consistency of a single sample's labels."""
        issues = []
        
        # Low score for genuine
        if sample.is_genuine and sample.trust_score < 30:
            issues.append(ValidationIssue(
                level=ValidationLevel.WARNING,
                validation_type=ValidationType.LABEL,
                message="Genuine sample has very low trust score",
                sample_id=sample.sample_id,
                value=sample.trust_score,
            ))
        
        # High score for fraud
        if not sample.is_genuine and sample.trust_score > 70:
            issues.append(ValidationIssue(
                level=ValidationLevel.WARNING,
                validation_type=ValidationType.LABEL,
                message="Fraud sample has very high trust score",
                sample_id=sample.sample_id,
                value=sample.trust_score,
            ))
        
        # Missing fraud type for fraud sample
        if not sample.is_genuine and not sample.fraud_type:
            issues.append(ValidationIssue(
                level=ValidationLevel.WARNING,
                validation_type=ValidationType.LABEL,
                message="Fraud sample is missing fraud_type",
                sample_id=sample.sample_id,
            ))
        
        # Fraud type for genuine sample
        if sample.is_genuine and sample.fraud_type:
            issues.append(ValidationIssue(
                level=ValidationLevel.ERROR,
                validation_type=ValidationType.LABEL,
                message="Genuine sample has fraud_type set",
                sample_id=sample.sample_id,
                value=sample.fraud_type,
            ))
        
        return issues


class DataQualityChecker:
    """Checks data quality of samples."""
    
    def __init__(
        self,
        min_face_confidence: float = 0.3,
        min_doc_quality: float = 0.2,
        min_ocr_confidence: float = 0.2,
    ):
        """
        Initialize quality checker.
        
        Args:
            min_face_confidence: Minimum acceptable face confidence
            min_doc_quality: Minimum acceptable document quality
            min_ocr_confidence: Minimum acceptable OCR confidence
        """
        self.min_face_confidence = min_face_confidence
        self.min_doc_quality = min_doc_quality
        self.min_ocr_confidence = min_ocr_confidence
    
    def check_quality(
        self,
        samples: List[IdentitySample],
    ) -> List[ValidationIssue]:
        """Check data quality of samples."""
        issues = []
        
        for sample in samples:
            # Check face confidence
            if sample.face_detected and sample.face_confidence < self.min_face_confidence:
                issues.append(ValidationIssue(
                    level=ValidationLevel.INFO,
                    validation_type=ValidationType.DATA_QUALITY,
                    message="Low face confidence",
                    sample_id=sample.sample_id,
                    field="face_confidence",
                    value=sample.face_confidence,
                    expected=f">= {self.min_face_confidence}",
                ))
            
            # Check document quality
            if sample.document_quality_score < self.min_doc_quality:
                issues.append(ValidationIssue(
                    level=ValidationLevel.INFO,
                    validation_type=ValidationType.DATA_QUALITY,
                    message="Low document quality score",
                    sample_id=sample.sample_id,
                    field="document_quality_score",
                    value=sample.document_quality_score,
                    expected=f">= {self.min_doc_quality}",
                ))
            
            # Check OCR confidence
            if sample.ocr_success and sample.ocr_confidence < self.min_ocr_confidence:
                issues.append(ValidationIssue(
                    level=ValidationLevel.INFO,
                    validation_type=ValidationType.DATA_QUALITY,
                    message="Low OCR confidence",
                    sample_id=sample.sample_id,
                    field="ocr_confidence",
                    value=sample.ocr_confidence,
                    expected=f">= {self.min_ocr_confidence}",
                ))
            
            # Check for missing images
            if sample.document_image is None:
                issues.append(ValidationIssue(
                    level=ValidationLevel.WARNING,
                    validation_type=ValidationType.DATA_QUALITY,
                    message="Missing document image",
                    sample_id=sample.sample_id,
                ))
            
            if sample.selfie_image is None:
                issues.append(ValidationIssue(
                    level=ValidationLevel.WARNING,
                    validation_type=ValidationType.DATA_QUALITY,
                    message="Missing selfie image",
                    sample_id=sample.sample_id,
                ))
        
        return issues


class DatasetValidator:
    """
    Complete dataset validation.
    
    Combines schema validation, label anomaly detection,
    and data quality checks.
    """
    
    def __init__(
        self,
        schema_validator: Optional[SchemaValidator] = None,
        anomaly_detector: Optional[LabelAnomalyDetector] = None,
        quality_checker: Optional[DataQualityChecker] = None,
        fail_on_error: bool = True,
    ):
        """
        Initialize dataset validator.
        
        Args:
            schema_validator: Schema validator instance
            anomaly_detector: Label anomaly detector instance
            quality_checker: Data quality checker instance
            fail_on_error: Raise exception on validation errors
        """
        self.schema_validator = schema_validator or SchemaValidator()
        self.anomaly_detector = anomaly_detector or LabelAnomalyDetector()
        self.quality_checker = quality_checker or DataQualityChecker()
        self.fail_on_error = fail_on_error
    
    def validate(self, dataset: Dataset) -> ValidationReport:
        """
        Validate a complete dataset.
        
        Args:
            dataset: Dataset to validate
            
        Returns:
            ValidationReport
            
        Raises:
            ValueError: If fail_on_error=True and validation fails
        """
        import time
        start_time = time.time()
        
        report = ValidationReport(
            schema_version=self.schema_validator.schema_version,
        )
        
        # Collect all samples
        all_samples = (
            list(dataset.train) +
            list(dataset.validation) +
            list(dataset.test)
        )
        
        report.samples_validated = len(all_samples)
        
        # Schema validation
        for sample in all_samples:
            issues = self.schema_validator.validate_sample(sample)
            for issue in issues:
                report.add_issue(issue)
        
        # Label anomaly detection
        anomaly_issues = self.anomaly_detector.detect_anomalies(all_samples)
        for issue in anomaly_issues:
            report.add_issue(issue)
        
        # Data quality checks
        quality_issues = self.quality_checker.check_quality(all_samples)
        for issue in quality_issues:
            report.add_issue(issue)
        
        # Split validation
        split_issues = self._validate_splits(dataset)
        for issue in split_issues:
            report.add_issue(issue)
        
        report.validation_time_seconds = time.time() - start_time
        
        logger.info(report.print_summary())
        
        if self.fail_on_error and not report.valid:
            raise ValueError(
                f"Dataset validation failed with {report.error_count} errors. "
                f"See report for details."
            )
        
        return report
    
    def _validate_splits(self, dataset: Dataset) -> List[ValidationIssue]:
        """Validate dataset splits."""
        issues = []
        
        # Check for empty splits
        for split_name, split in [
            ("train", dataset.train),
            ("validation", dataset.validation),
            ("test", dataset.test),
        ]:
            if len(split) == 0:
                issues.append(ValidationIssue(
                    level=ValidationLevel.ERROR,
                    validation_type=ValidationType.SPLIT,
                    message=f"{split_name} split is empty",
                ))
        
        # Check for overlap
        train_ids = {s.sample_id for s in dataset.train}
        val_ids = {s.sample_id for s in dataset.validation}
        test_ids = {s.sample_id for s in dataset.test}
        
        train_val_overlap = train_ids & val_ids
        train_test_overlap = train_ids & test_ids
        val_test_overlap = val_ids & test_ids
        
        if train_val_overlap:
            issues.append(ValidationIssue(
                level=ValidationLevel.ERROR,
                validation_type=ValidationType.SPLIT,
                message=f"Train-validation split overlap: {len(train_val_overlap)} samples",
            ))
        
        if train_test_overlap:
            issues.append(ValidationIssue(
                level=ValidationLevel.ERROR,
                validation_type=ValidationType.SPLIT,
                message=f"Train-test split overlap: {len(train_test_overlap)} samples",
            ))
        
        if val_test_overlap:
            issues.append(ValidationIssue(
                level=ValidationLevel.ERROR,
                validation_type=ValidationType.SPLIT,
                message=f"Validation-test split overlap: {len(val_test_overlap)} samples",
            ))
        
        # Check for duplicate IDs within splits
        for split_name, split in [
            ("train", dataset.train),
            ("validation", dataset.validation),
            ("test", dataset.test),
        ]:
            ids = [s.sample_id for s in split]
            if len(ids) != len(set(ids)):
                issues.append(ValidationIssue(
                    level=ValidationLevel.ERROR,
                    validation_type=ValidationType.SPLIT,
                    message=f"Duplicate sample IDs in {split_name} split",
                ))
        
        return issues


def validate_dataset(
    dataset: Dataset,
    fail_on_error: bool = False,
) -> ValidationReport:
    """
    Convenience function to validate a dataset.
    
    Args:
        dataset: Dataset to validate
        fail_on_error: Raise exception on validation errors
        
    Returns:
        ValidationReport
    """
    validator = DatasetValidator(fail_on_error=fail_on_error)
    return validator.validate(dataset)
