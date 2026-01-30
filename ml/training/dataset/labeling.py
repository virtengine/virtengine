"""
Labeling pipeline for VEID training data.

This module provides:
- Human review CSV import/export
- Heuristic auto-labeling based on signals
- Label quality validation
- Consensus labeling for multi-annotator workflows
"""

import csv
import hashlib
import json
import logging
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from pathlib import Path
from typing import Any, Callable, Dict, List, Optional, Tuple

import numpy as np

from ml.training.dataset.ingestion import Dataset, IdentitySample, DatasetSplit

logger = logging.getLogger(__name__)


class LabelSource(str, Enum):
    """Source of labels."""
    HUMAN = "human"  # Human annotator
    HEURISTIC = "heuristic"  # Rule-based auto-labeling
    MODEL = "model"  # Model-predicted
    CONSENSUS = "consensus"  # Aggregated from multiple sources


class LabelStatus(str, Enum):
    """Status of a label."""
    PENDING = "pending"
    REVIEWED = "reviewed"
    APPROVED = "approved"
    REJECTED = "rejected"
    NEEDS_REVIEW = "needs_review"


@dataclass
class Label:
    """A single label for a sample."""
    
    sample_id: str
    trust_score: float
    is_genuine: bool
    fraud_type: Optional[str] = None
    
    # Metadata
    source: LabelSource = LabelSource.HEURISTIC
    status: LabelStatus = LabelStatus.PENDING
    confidence: float = 1.0
    
    # Annotator info
    annotator_id: Optional[str] = None
    annotation_timestamp: Optional[float] = None
    
    # Notes and flags
    notes: str = ""
    flags: List[str] = field(default_factory=list)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "sample_id": self.sample_id,
            "trust_score": self.trust_score,
            "is_genuine": self.is_genuine,
            "fraud_type": self.fraud_type,
            "source": self.source.value,
            "status": self.status.value,
            "confidence": self.confidence,
            "annotator_id": self.annotator_id,
            "annotation_timestamp": self.annotation_timestamp,
            "notes": self.notes,
            "flags": self.flags,
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "Label":
        """Create from dictionary."""
        return cls(
            sample_id=data["sample_id"],
            trust_score=float(data.get("trust_score", 0)),
            is_genuine=bool(data.get("is_genuine", True)),
            fraud_type=data.get("fraud_type"),
            source=LabelSource(data.get("source", "heuristic")),
            status=LabelStatus(data.get("status", "pending")),
            confidence=float(data.get("confidence", 1.0)),
            annotator_id=data.get("annotator_id"),
            annotation_timestamp=data.get("annotation_timestamp"),
            notes=data.get("notes", ""),
            flags=data.get("flags", []),
        )


@dataclass
class LabelBatch:
    """A batch of labels."""
    
    labels: List[Label]
    batch_id: str = ""
    created_at: float = field(default_factory=lambda: datetime.now().timestamp())
    metadata: Dict[str, Any] = field(default_factory=dict)
    
    def __len__(self) -> int:
        return len(self.labels)
    
    def __iter__(self):
        return iter(self.labels)
    
    def get_label(self, sample_id: str) -> Optional[Label]:
        """Get label by sample ID."""
        for label in self.labels:
            if label.sample_id == sample_id:
                return label
        return None
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "batch_id": self.batch_id,
            "created_at": self.created_at,
            "labels": [l.to_dict() for l in self.labels],
            "metadata": self.metadata,
        }


@dataclass
class HeuristicRule:
    """A heuristic rule for auto-labeling."""
    
    name: str
    description: str
    condition: Callable[[IdentitySample], bool]
    score_adjustment: float = 0.0  # Adjustment to trust score
    is_fraud_indicator: bool = False
    fraud_type: Optional[str] = None
    confidence: float = 0.8
    priority: int = 0  # Higher = applied later


# Default heuristic rules
DEFAULT_RULES: List[HeuristicRule] = [
    HeuristicRule(
        name="low_face_confidence",
        description="Face detection confidence too low",
        condition=lambda s: s.face_confidence < 0.5,
        score_adjustment=-20.0,
        confidence=0.7,
        priority=1,
    ),
    HeuristicRule(
        name="no_face_detected",
        description="No face detected in image",
        condition=lambda s: not s.face_detected,
        score_adjustment=-40.0,
        is_fraud_indicator=True,
        fraud_type="missing_face",
        confidence=0.9,
        priority=2,
    ),
    HeuristicRule(
        name="poor_document_quality",
        description="Document quality score below threshold",
        condition=lambda s: s.document_quality_score < 0.4,
        score_adjustment=-15.0,
        confidence=0.6,
        priority=1,
    ),
    HeuristicRule(
        name="ocr_failure",
        description="OCR extraction failed",
        condition=lambda s: not s.ocr_success,
        score_adjustment=-25.0,
        confidence=0.7,
        priority=1,
    ),
    HeuristicRule(
        name="low_ocr_confidence",
        description="OCR confidence too low",
        condition=lambda s: s.ocr_success and s.ocr_confidence < 0.4,
        score_adjustment=-10.0,
        confidence=0.6,
        priority=1,
    ),
    HeuristicRule(
        name="high_quality_signals",
        description="All quality signals above threshold",
        condition=lambda s: (
            s.face_confidence > 0.9 and
            s.document_quality_score > 0.8 and
            s.ocr_confidence > 0.85
        ),
        score_adjustment=10.0,
        confidence=0.85,
        priority=1,
    ),
    HeuristicRule(
        name="expired_document",
        description="Document is expired",
        condition=lambda s: s.document_info and s.document_info.expiry_valid is False,
        score_adjustment=-15.0,
        confidence=0.9,
        priority=2,
    ),
]


class HeuristicLabeler:
    """
    Auto-labels samples using heuristic rules.
    
    Rules are applied in priority order to compute trust scores
    and detect fraud indicators.
    """
    
    def __init__(
        self,
        rules: Optional[List[HeuristicRule]] = None,
        base_score: float = 75.0,
    ):
        """
        Initialize the heuristic labeler.
        
        Args:
            rules: List of heuristic rules to apply
            base_score: Base trust score before adjustments
        """
        self.rules = rules or DEFAULT_RULES
        self.base_score = base_score
        
        # Sort rules by priority
        self.rules = sorted(self.rules, key=lambda r: r.priority)
    
    def label_sample(self, sample: IdentitySample) -> Label:
        """
        Generate a label for a single sample.
        
        Args:
            sample: Sample to label
            
        Returns:
            Generated label
        """
        trust_score = self.base_score
        is_fraud = False
        fraud_type = None
        applied_rules = []
        min_confidence = 1.0
        flags = []
        
        for rule in self.rules:
            try:
                if rule.condition(sample):
                    trust_score += rule.score_adjustment
                    applied_rules.append(rule.name)
                    min_confidence = min(min_confidence, rule.confidence)
                    
                    if rule.is_fraud_indicator:
                        is_fraud = True
                        fraud_type = rule.fraud_type
                    
                    flags.append(rule.name)
            except Exception as e:
                logger.warning(f"Rule {rule.name} failed for {sample.sample_id}: {e}")
        
        # Clamp score
        trust_score = max(0.0, min(100.0, trust_score))
        
        return Label(
            sample_id=sample.sample_id,
            trust_score=trust_score,
            is_genuine=not is_fraud,
            fraud_type=fraud_type,
            source=LabelSource.HEURISTIC,
            status=LabelStatus.PENDING,
            confidence=min_confidence,
            notes=f"Applied rules: {', '.join(applied_rules)}" if applied_rules else "",
            flags=flags,
        )
    
    def label_batch(
        self,
        samples: List[IdentitySample],
        batch_id: Optional[str] = None,
    ) -> LabelBatch:
        """
        Label a batch of samples.
        
        Args:
            samples: Samples to label
            batch_id: Optional batch identifier
            
        Returns:
            LabelBatch with generated labels
        """
        labels = [self.label_sample(s) for s in samples]
        
        batch_id = batch_id or f"heuristic_{datetime.now().strftime('%Y%m%d_%H%M%S')}"
        
        return LabelBatch(
            labels=labels,
            batch_id=batch_id,
            metadata={
                "source": "heuristic",
                "num_rules": len(self.rules),
                "base_score": self.base_score,
            },
        )


class CSVLabelReader:
    """
    Reads labels from CSV files exported from labeling tools.
    
    Expected CSV columns:
    - sample_id (required)
    - trust_score (required)
    - is_genuine (optional, default True)
    - fraud_type (optional)
    - annotator_id (optional)
    - notes (optional)
    """
    
    REQUIRED_COLUMNS = ["sample_id", "trust_score"]
    
    def read(self, csv_path: str) -> LabelBatch:
        """
        Read labels from a CSV file.
        
        Args:
            csv_path: Path to CSV file
            
        Returns:
            LabelBatch with imported labels
        """
        labels = []
        path = Path(csv_path)
        
        with open(path, "r", newline="", encoding="utf-8") as f:
            reader = csv.DictReader(f)
            
            # Validate columns
            if not all(col in reader.fieldnames for col in self.REQUIRED_COLUMNS):
                raise ValueError(
                    f"CSV missing required columns: {self.REQUIRED_COLUMNS}"
                )
            
            for row in reader:
                label = self._parse_row(row)
                labels.append(label)
        
        batch_id = f"csv_{path.stem}_{datetime.now().strftime('%Y%m%d')}"
        
        return LabelBatch(
            labels=labels,
            batch_id=batch_id,
            metadata={
                "source": "csv",
                "file": str(path),
            },
        )
    
    def _parse_row(self, row: Dict[str, str]) -> Label:
        """Parse a CSV row into a Label."""
        return Label(
            sample_id=row["sample_id"].strip(),
            trust_score=float(row["trust_score"]),
            is_genuine=self._parse_bool(row.get("is_genuine", "true")),
            fraud_type=row.get("fraud_type", "").strip() or None,
            source=LabelSource.HUMAN,
            status=LabelStatus.REVIEWED,
            annotator_id=row.get("annotator_id", "").strip() or None,
            annotation_timestamp=datetime.now().timestamp(),
            notes=row.get("notes", "").strip(),
        )
    
    def _parse_bool(self, value: str) -> bool:
        """Parse a string to boolean."""
        return value.lower() in ("true", "1", "yes", "y")


class CSVLabelWriter:
    """Writes labels to CSV format for human review."""
    
    COLUMNS = [
        "sample_id",
        "trust_score",
        "is_genuine",
        "fraud_type",
        "annotator_id",
        "status",
        "confidence",
        "notes",
    ]
    
    def write(
        self,
        labels: LabelBatch,
        output_path: str,
        include_metadata: bool = True,
    ) -> None:
        """
        Write labels to a CSV file.
        
        Args:
            labels: Labels to export
            output_path: Output file path
            include_metadata: Include additional columns
        """
        path = Path(output_path)
        path.parent.mkdir(parents=True, exist_ok=True)
        
        with open(path, "w", newline="", encoding="utf-8") as f:
            writer = csv.DictWriter(f, fieldnames=self.COLUMNS)
            writer.writeheader()
            
            for label in labels:
                row = {
                    "sample_id": label.sample_id,
                    "trust_score": f"{label.trust_score:.2f}",
                    "is_genuine": str(label.is_genuine).lower(),
                    "fraud_type": label.fraud_type or "",
                    "annotator_id": label.annotator_id or "",
                    "status": label.status.value,
                    "confidence": f"{label.confidence:.2f}",
                    "notes": label.notes,
                }
                writer.writerow(row)
        
        logger.info(f"Wrote {len(labels)} labels to {path}")


class LabelValidator:
    """Validates label quality and consistency."""
    
    def __init__(
        self,
        min_score: float = 0.0,
        max_score: float = 100.0,
        require_fraud_type_for_fraud: bool = True,
    ):
        """
        Initialize validator.
        
        Args:
            min_score: Minimum valid trust score
            max_score: Maximum valid trust score
            require_fraud_type_for_fraud: Require fraud_type when is_genuine=False
        """
        self.min_score = min_score
        self.max_score = max_score
        self.require_fraud_type = require_fraud_type_for_fraud
    
    def validate(self, labels: LabelBatch) -> Dict[str, Any]:
        """
        Validate a batch of labels.
        
        Args:
            labels: Labels to validate
            
        Returns:
            Validation report
        """
        report = {
            "total": len(labels),
            "valid": 0,
            "invalid": 0,
            "errors": [],
            "warnings": [],
        }
        
        seen_ids = set()
        scores = []
        
        for label in labels:
            errors = []
            warnings = []
            
            # Check for duplicates
            if label.sample_id in seen_ids:
                errors.append(f"Duplicate sample_id: {label.sample_id}")
            seen_ids.add(label.sample_id)
            
            # Validate score range
            if not (self.min_score <= label.trust_score <= self.max_score):
                errors.append(
                    f"Score {label.trust_score} out of range "
                    f"[{self.min_score}, {self.max_score}]"
                )
            
            # Check fraud type requirement
            if (not label.is_genuine and
                self.require_fraud_type and
                not label.fraud_type):
                warnings.append("Fraud sample missing fraud_type")
            
            # Check for inconsistent scores
            if label.is_genuine and label.trust_score < 30:
                warnings.append(
                    f"Low trust score ({label.trust_score}) for genuine sample"
                )
            if not label.is_genuine and label.trust_score > 70:
                warnings.append(
                    f"High trust score ({label.trust_score}) for fraud sample"
                )
            
            if errors:
                report["invalid"] += 1
                for e in errors:
                    report["errors"].append(f"{label.sample_id}: {e}")
            else:
                report["valid"] += 1
            
            for w in warnings:
                report["warnings"].append(f"{label.sample_id}: {w}")
            
            scores.append(label.trust_score)
        
        # Compute statistics
        if scores:
            report["statistics"] = {
                "mean_score": float(np.mean(scores)),
                "std_score": float(np.std(scores)),
                "min_score": float(np.min(scores)),
                "max_score": float(np.max(scores)),
                "genuine_ratio": sum(1 for l in labels if l.is_genuine) / len(labels),
            }
        
        return report


class ConsensusLabeler:
    """
    Aggregates labels from multiple annotators.
    
    Strategies:
    - majority: Take majority vote
    - average: Average trust scores
    - weighted: Weighted by annotator confidence
    """
    
    def __init__(
        self,
        strategy: str = "average",
        min_annotators: int = 2,
    ):
        """
        Initialize consensus labeler.
        
        Args:
            strategy: Aggregation strategy
            min_annotators: Minimum number of annotators required
        """
        self.strategy = strategy
        self.min_annotators = min_annotators
    
    def merge(
        self,
        label_batches: List[LabelBatch],
    ) -> LabelBatch:
        """
        Merge multiple label batches into consensus labels.
        
        Args:
            label_batches: List of label batches from different annotators
            
        Returns:
            Consensus label batch
        """
        # Group labels by sample_id
        labels_by_sample: Dict[str, List[Label]] = {}
        
        for batch in label_batches:
            for label in batch:
                if label.sample_id not in labels_by_sample:
                    labels_by_sample[label.sample_id] = []
                labels_by_sample[label.sample_id].append(label)
        
        # Generate consensus labels
        consensus_labels = []
        
        for sample_id, sample_labels in labels_by_sample.items():
            if len(sample_labels) < self.min_annotators:
                logger.warning(
                    f"Sample {sample_id} has only {len(sample_labels)} annotators, "
                    f"need {self.min_annotators}"
                )
                continue
            
            consensus = self._compute_consensus(sample_id, sample_labels)
            consensus_labels.append(consensus)
        
        return LabelBatch(
            labels=consensus_labels,
            batch_id=f"consensus_{datetime.now().strftime('%Y%m%d_%H%M%S')}",
            metadata={
                "strategy": self.strategy,
                "num_batches": len(label_batches),
                "min_annotators": self.min_annotators,
            },
        )
    
    def _compute_consensus(
        self,
        sample_id: str,
        labels: List[Label],
    ) -> Label:
        """Compute consensus from multiple labels."""
        if self.strategy == "average":
            trust_score = float(np.mean([l.trust_score for l in labels]))
        elif self.strategy == "weighted":
            total_weight = sum(l.confidence for l in labels)
            trust_score = sum(
                l.trust_score * l.confidence for l in labels
            ) / total_weight
        else:  # majority
            trust_score = float(np.median([l.trust_score for l in labels]))
        
        # Majority vote for is_genuine
        genuine_votes = sum(1 for l in labels if l.is_genuine)
        is_genuine = genuine_votes > len(labels) / 2
        
        # Most common fraud type
        fraud_types = [l.fraud_type for l in labels if l.fraud_type]
        fraud_type = max(set(fraud_types), key=fraud_types.count) if fraud_types else None
        
        # Compute confidence based on agreement
        score_std = float(np.std([l.trust_score for l in labels]))
        agreement_confidence = max(0.5, 1.0 - score_std / 50.0)
        
        return Label(
            sample_id=sample_id,
            trust_score=trust_score,
            is_genuine=is_genuine,
            fraud_type=fraud_type,
            source=LabelSource.CONSENSUS,
            status=LabelStatus.REVIEWED,
            confidence=agreement_confidence,
            notes=f"Consensus from {len(labels)} annotators",
        )


class LabelingPipeline:
    """
    Complete labeling pipeline combining human and heuristic labels.
    """
    
    def __init__(
        self,
        heuristic_labeler: Optional[HeuristicLabeler] = None,
        validator: Optional[LabelValidator] = None,
    ):
        """
        Initialize the pipeline.
        
        Args:
            heuristic_labeler: Heuristic labeler for auto-labeling
            validator: Label validator
        """
        self.heuristic_labeler = heuristic_labeler or HeuristicLabeler()
        self.validator = validator or LabelValidator()
        self.csv_reader = CSVLabelReader()
        self.csv_writer = CSVLabelWriter()
    
    def label_dataset(
        self,
        dataset: Dataset,
        human_labels_path: Optional[str] = None,
        export_path: Optional[str] = None,
    ) -> Tuple[Dataset, Dict[str, Any]]:
        """
        Label a dataset with combined heuristic and human labels.
        
        Args:
            dataset: Dataset to label
            human_labels_path: Path to human labels CSV
            export_path: Path to export labels for review
            
        Returns:
            Tuple of (labeled dataset, validation report)
        """
        # Generate heuristic labels for all samples
        all_samples = (
            list(dataset.train) +
            list(dataset.validation) +
            list(dataset.test)
        )
        
        heuristic_labels = self.heuristic_labeler.label_batch(all_samples)
        
        # Import human labels if provided
        if human_labels_path:
            human_labels = self.csv_reader.read(human_labels_path)
            # Override heuristic labels with human labels
            human_label_map = {l.sample_id: l for l in human_labels}
            
            merged_labels = []
            for label in heuristic_labels:
                if label.sample_id in human_label_map:
                    merged_labels.append(human_label_map[label.sample_id])
                else:
                    merged_labels.append(label)
            
            final_labels = LabelBatch(
                labels=merged_labels,
                batch_id=f"merged_{datetime.now().strftime('%Y%m%d')}",
            )
        else:
            final_labels = heuristic_labels
        
        # Validate labels
        validation_report = self.validator.validate(final_labels)
        
        # Apply labels to dataset
        label_map = {l.sample_id: l for l in final_labels}
        
        for split in [dataset.train, dataset.validation, dataset.test]:
            for sample in split.samples:
                if sample.sample_id in label_map:
                    label = label_map[sample.sample_id]
                    sample.trust_score = label.trust_score
                    sample.is_genuine = label.is_genuine
                    sample.fraud_type = label.fraud_type
        
        # Export for review if requested
        if export_path:
            self.csv_writer.write(final_labels, export_path)
        
        return dataset, validation_report
