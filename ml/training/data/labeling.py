"""
Production labeling pipeline for VEID ML training.

This module provides production-ready labeling workflow with:
- Synthetic and real data path support
- Human review queue management
- Auto-labeling with confidence thresholds
- Label quality validation
- Consensus labeling for multi-annotator workflows

Designed for both CI/testing (synthetic) and production (real) data paths.
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

from ml.training.dataset.ingestion import (
    Dataset,
    DatasetSplit,
    IdentitySample,
    SplitType,
)
from ml.training.dataset.labeling import (
    CSVLabelReader,
    CSVLabelWriter,
    HeuristicLabeler,
    Label,
    LabelBatch,
    LabelSource,
    LabelStatus,
    LabelValidator,
    ConsensusLabeler,
)

logger = logging.getLogger(__name__)


class LabelingWorkflow(str, Enum):
    """Labeling workflow types."""
    SYNTHETIC = "synthetic"  # Auto-label synthetic data (CI path)
    HEURISTIC = "heuristic"  # Heuristic-only labeling
    HUMAN_REVIEW = "human_review"  # Full human review workflow
    HYBRID = "hybrid"  # Heuristic + human review for uncertain cases
    CONSENSUS = "consensus"  # Multi-annotator consensus


class ReviewPriority(str, Enum):
    """Priority levels for review queue."""
    CRITICAL = "critical"  # Immediate review needed
    HIGH = "high"  # High priority
    NORMAL = "normal"  # Standard priority
    LOW = "low"  # Low priority


@dataclass
class ReviewItem:
    """An item in the review queue."""
    
    sample_id: str
    sample: IdentitySample
    priority: ReviewPriority = ReviewPriority.NORMAL
    heuristic_label: Optional[Label] = None
    reason: str = ""
    created_at: float = field(default_factory=lambda: datetime.now().timestamp())
    
    # Review status
    assigned_to: Optional[str] = None
    reviewed: bool = False
    final_label: Optional[Label] = None
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "sample_id": self.sample_id,
            "priority": self.priority.value,
            "heuristic_score": self.heuristic_label.trust_score if self.heuristic_label else None,
            "heuristic_confidence": self.heuristic_label.confidence if self.heuristic_label else None,
            "reason": self.reason,
            "created_at": self.created_at,
            "assigned_to": self.assigned_to,
            "reviewed": self.reviewed,
        }


@dataclass
class LabelingConfig:
    """Configuration for labeling pipeline."""
    
    # Workflow
    workflow: LabelingWorkflow = LabelingWorkflow.HYBRID
    
    # Auto-labeling thresholds
    auto_label_confidence_threshold: float = 0.85  # Auto-accept above this
    review_confidence_threshold: float = 0.6  # Review below this
    
    # Quality control
    sample_for_qa: float = 0.1  # Fraction to sample for QA review
    min_qa_samples: int = 10
    max_qa_samples: int = 100
    
    # Consensus settings
    min_annotators: int = 2
    consensus_strategy: str = "average"  # "average", "majority", "weighted"
    
    # Output
    export_dir: Optional[str] = None
    export_format: str = "csv"  # "csv", "json"
    
    # Heuristic labeler settings
    heuristic_base_score: float = 75.0
    
    # Random seed
    random_seed: int = 42


@dataclass
class LabelingResult:
    """Result of a labeling operation."""
    
    # Statistics
    total_samples: int = 0
    auto_labeled: int = 0
    sent_for_review: int = 0
    human_labeled: int = 0
    
    # Quality metrics
    mean_confidence: float = 0.0
    low_confidence_count: int = 0
    
    # Timing
    started_at: float = field(default_factory=lambda: datetime.now().timestamp())
    completed_at: Optional[float] = None
    
    # Status
    success: bool = True
    warnings: List[str] = field(default_factory=list)
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "total_samples": self.total_samples,
            "auto_labeled": self.auto_labeled,
            "sent_for_review": self.sent_for_review,
            "human_labeled": self.human_labeled,
            "mean_confidence": self.mean_confidence,
            "low_confidence_count": self.low_confidence_count,
            "success": self.success,
            "warnings": self.warnings,
        }


class ReviewQueue:
    """
    Queue for managing human review of samples.
    
    Provides priority-based assignment and tracking of review items.
    """
    
    def __init__(self, queue_id: str = "default"):
        """
        Initialize review queue.
        
        Args:
            queue_id: Unique identifier for this queue
        """
        self.queue_id = queue_id
        self._items: Dict[str, ReviewItem] = {}
        self._created_at = datetime.now().timestamp()
    
    def add(
        self,
        sample: IdentitySample,
        heuristic_label: Optional[Label] = None,
        priority: ReviewPriority = ReviewPriority.NORMAL,
        reason: str = "",
    ) -> ReviewItem:
        """Add a sample to the review queue."""
        item = ReviewItem(
            sample_id=sample.sample_id,
            sample=sample,
            priority=priority,
            heuristic_label=heuristic_label,
            reason=reason,
        )
        self._items[sample.sample_id] = item
        return item
    
    def get(self, sample_id: str) -> Optional[ReviewItem]:
        """Get a specific review item."""
        return self._items.get(sample_id)
    
    def list_pending(
        self,
        priority: Optional[ReviewPriority] = None,
        limit: Optional[int] = None,
    ) -> List[ReviewItem]:
        """List pending review items."""
        items = [i for i in self._items.values() if not i.reviewed]
        
        if priority:
            items = [i for i in items if i.priority == priority]
        
        # Sort by priority and creation time
        priority_order = {
            ReviewPriority.CRITICAL: 0,
            ReviewPriority.HIGH: 1,
            ReviewPriority.NORMAL: 2,
            ReviewPriority.LOW: 3,
        }
        items.sort(key=lambda i: (priority_order[i.priority], i.created_at))
        
        if limit:
            items = items[:limit]
        
        return items
    
    def assign(self, sample_id: str, annotator_id: str) -> bool:
        """Assign a sample to an annotator."""
        item = self._items.get(sample_id)
        if item and not item.reviewed:
            item.assigned_to = annotator_id
            return True
        return False
    
    def submit_review(self, sample_id: str, label: Label) -> bool:
        """Submit a review for a sample."""
        item = self._items.get(sample_id)
        if item:
            item.final_label = label
            item.reviewed = True
            return True
        return False
    
    def get_completed(self) -> List[ReviewItem]:
        """Get all completed review items."""
        return [i for i in self._items.values() if i.reviewed]
    
    def export_for_review(self, output_path: str) -> int:
        """Export pending items for external review."""
        pending = self.list_pending()
        
        rows = []
        for item in pending:
            row = {
                "sample_id": item.sample_id,
                "priority": item.priority.value,
                "reason": item.reason,
                "heuristic_score": item.heuristic_label.trust_score if item.heuristic_label else "",
                "is_genuine": "",  # To be filled by reviewer
                "trust_score": "",  # To be filled by reviewer
                "fraud_type": "",  # To be filled by reviewer
                "notes": "",  # To be filled by reviewer
            }
            rows.append(row)
        
        path = Path(output_path)
        path.parent.mkdir(parents=True, exist_ok=True)
        
        with open(path, "w", newline="") as f:
            writer = csv.DictWriter(f, fieldnames=list(rows[0].keys()) if rows else [])
            writer.writeheader()
            writer.writerows(rows)
        
        logger.info(f"Exported {len(rows)} samples for review to {output_path}")
        return len(rows)
    
    def import_reviews(self, input_path: str) -> int:
        """Import reviews from external file."""
        reader = CSVLabelReader()
        batch = reader.read(input_path)
        
        imported = 0
        for label in batch.labels:
            if self.submit_review(label.sample_id, label):
                imported += 1
        
        logger.info(f"Imported {imported} reviews from {input_path}")
        return imported
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert queue to dictionary."""
        return {
            "queue_id": self.queue_id,
            "created_at": self._created_at,
            "total_items": len(self._items),
            "pending": len([i for i in self._items.values() if not i.reviewed]),
            "completed": len([i for i in self._items.values() if i.reviewed]),
            "items": [i.to_dict() for i in self._items.values()],
        }


class AutoLabeler:
    """
    Automatic labeling with confidence-based routing.
    
    High-confidence samples are auto-labeled, low-confidence
    samples are routed to human review.
    """
    
    def __init__(
        self,
        confidence_threshold: float = 0.85,
        base_score: float = 75.0,
    ):
        """
        Initialize auto-labeler.
        
        Args:
            confidence_threshold: Minimum confidence for auto-labeling
            base_score: Base trust score for heuristic labeling
        """
        self.confidence_threshold = confidence_threshold
        self.heuristic_labeler = HeuristicLabeler(base_score=base_score)
    
    def label(
        self,
        sample: IdentitySample
    ) -> Tuple[Label, bool]:
        """
        Generate label for a sample.
        
        Args:
            sample: Sample to label
            
        Returns:
            Tuple of (Label, auto_accepted: bool)
        """
        label = self.heuristic_labeler.label_sample(sample)
        auto_accepted = label.confidence >= self.confidence_threshold
        
        if auto_accepted:
            label.status = LabelStatus.APPROVED
        else:
            label.status = LabelStatus.NEEDS_REVIEW
        
        return label, auto_accepted
    
    def label_batch(
        self,
        samples: List[IdentitySample]
    ) -> Tuple[List[Label], List[IdentitySample]]:
        """
        Label a batch of samples.
        
        Args:
            samples: Samples to label
            
        Returns:
            Tuple of (accepted labels, samples needing review)
        """
        accepted_labels = []
        needs_review = []
        
        for sample in samples:
            label, auto_accepted = self.label(sample)
            
            if auto_accepted:
                accepted_labels.append(label)
            else:
                # Attach label for reference during review
                sample.annotations["heuristic_label"] = label.to_dict()
                needs_review.append(sample)
        
        return accepted_labels, needs_review


class ProductionLabelingPipeline:
    """
    Production-grade labeling pipeline.
    
    Supports multiple workflows:
    - SYNTHETIC: Auto-label synthetic data (CI/testing path)
    - HEURISTIC: Heuristic-only labeling
    - HUMAN_REVIEW: Full human review for all samples
    - HYBRID: Heuristic + human review for uncertain cases
    - CONSENSUS: Multi-annotator consensus
    
    Usage:
        config = LabelingConfig(workflow=LabelingWorkflow.HYBRID)
        pipeline = ProductionLabelingPipeline(config)
        
        dataset, result = pipeline.label_dataset(samples)
    """
    
    def __init__(self, config: Optional[LabelingConfig] = None):
        """
        Initialize labeling pipeline.
        
        Args:
            config: Labeling configuration
        """
        self.config = config or LabelingConfig()
        self._rng = np.random.RandomState(self.config.random_seed)
        
        # Initialize components
        self.auto_labeler = AutoLabeler(
            confidence_threshold=self.config.auto_label_confidence_threshold,
            base_score=self.config.heuristic_base_score,
        )
        
        self.review_queue = ReviewQueue()
        self.label_validator = LabelValidator()
        
        if self.config.workflow == LabelingWorkflow.CONSENSUS:
            self.consensus_labeler = ConsensusLabeler(
                strategy=self.config.consensus_strategy,
                min_annotators=self.config.min_annotators,
            )
        else:
            self.consensus_labeler = None
    
    def label_dataset(
        self,
        samples: List[IdentitySample],
        human_labels_path: Optional[str] = None,
    ) -> Tuple[Dataset, LabelingResult]:
        """
        Label a dataset using the configured workflow.
        
        Args:
            samples: Samples to label
            human_labels_path: Path to human labels (for HYBRID/HUMAN_REVIEW)
            
        Returns:
            Tuple of (labeled Dataset, LabelingResult)
        """
        result = LabelingResult(total_samples=len(samples))
        
        logger.info(f"Starting labeling with workflow: {self.config.workflow.value}")
        
        if self.config.workflow == LabelingWorkflow.SYNTHETIC:
            labeled_samples = self._synthetic_workflow(samples, result)
        
        elif self.config.workflow == LabelingWorkflow.HEURISTIC:
            labeled_samples = self._heuristic_workflow(samples, result)
        
        elif self.config.workflow == LabelingWorkflow.HUMAN_REVIEW:
            labeled_samples = self._human_review_workflow(
                samples, human_labels_path, result
            )
        
        elif self.config.workflow == LabelingWorkflow.HYBRID:
            labeled_samples = self._hybrid_workflow(
                samples, human_labels_path, result
            )
        
        elif self.config.workflow == LabelingWorkflow.CONSENSUS:
            labeled_samples = self._consensus_workflow(
                samples, human_labels_path, result
            )
        
        else:
            raise ValueError(f"Unknown workflow: {self.config.workflow}")
        
        # Create dataset
        dataset = self._create_dataset(labeled_samples)
        
        result.completed_at = datetime.now().timestamp()
        
        logger.info(
            f"Labeling complete: {result.auto_labeled} auto, "
            f"{result.human_labeled} human, "
            f"{result.sent_for_review} pending review"
        )
        
        return dataset, result
    
    def _synthetic_workflow(
        self,
        samples: List[IdentitySample],
        result: LabelingResult
    ) -> List[IdentitySample]:
        """Workflow for synthetic data (auto-label all)."""
        batch = self.auto_labeler.heuristic_labeler.label_batch(samples)
        
        label_map = {l.sample_id: l for l in batch}
        
        for sample in samples:
            if sample.sample_id in label_map:
                label = label_map[sample.sample_id]
                sample.trust_score = label.trust_score
                sample.is_genuine = label.is_genuine
                sample.fraud_type = label.fraud_type
                result.auto_labeled += 1
        
        result.mean_confidence = float(np.mean([l.confidence for l in batch]))
        
        return samples
    
    def _heuristic_workflow(
        self,
        samples: List[IdentitySample],
        result: LabelingResult
    ) -> List[IdentitySample]:
        """Workflow using only heuristic labeling."""
        return self._synthetic_workflow(samples, result)
    
    def _human_review_workflow(
        self,
        samples: List[IdentitySample],
        human_labels_path: Optional[str],
        result: LabelingResult
    ) -> List[IdentitySample]:
        """Workflow requiring human review for all samples."""
        # Add all samples to review queue
        for sample in samples:
            heuristic_label = self.auto_labeler.heuristic_labeler.label_sample(sample)
            self.review_queue.add(
                sample,
                heuristic_label=heuristic_label,
                priority=ReviewPriority.NORMAL,
                reason="Full human review workflow",
            )
            result.sent_for_review += 1
        
        # Import human labels if provided
        if human_labels_path:
            imported = self.review_queue.import_reviews(human_labels_path)
            result.human_labeled = imported
            
            # Apply labels to samples
            for item in self.review_queue.get_completed():
                if item.final_label:
                    sample = item.sample
                    sample.trust_score = item.final_label.trust_score
                    sample.is_genuine = item.final_label.is_genuine
                    sample.fraud_type = item.final_label.fraud_type
        
        return samples
    
    def _hybrid_workflow(
        self,
        samples: List[IdentitySample],
        human_labels_path: Optional[str],
        result: LabelingResult
    ) -> List[IdentitySample]:
        """Workflow combining heuristic and human review."""
        accepted_labels, needs_review = self.auto_labeler.label_batch(samples)
        
        # Apply auto-accepted labels
        label_map = {l.sample_id: l for l in accepted_labels}
        
        for sample in samples:
            if sample.sample_id in label_map:
                label = label_map[sample.sample_id]
                sample.trust_score = label.trust_score
                sample.is_genuine = label.is_genuine
                sample.fraud_type = label.fraud_type
                result.auto_labeled += 1
        
        # Add uncertain samples to review queue
        for sample in needs_review:
            heuristic_data = sample.annotations.get("heuristic_label", {})
            heuristic_label = Label.from_dict(heuristic_data) if heuristic_data else None
            
            priority = ReviewPriority.HIGH if (
                heuristic_label and heuristic_label.confidence < 0.5
            ) else ReviewPriority.NORMAL
            
            self.review_queue.add(
                sample,
                heuristic_label=heuristic_label,
                priority=priority,
                reason=f"Confidence below threshold: {heuristic_label.confidence if heuristic_label else 'N/A'}",
            )
            result.sent_for_review += 1
        
        # Import and apply human labels if provided
        if human_labels_path:
            imported = self.review_queue.import_reviews(human_labels_path)
            result.human_labeled = imported
            
            for item in self.review_queue.get_completed():
                if item.final_label:
                    sample = item.sample
                    sample.trust_score = item.final_label.trust_score
                    sample.is_genuine = item.final_label.is_genuine
                    sample.fraud_type = item.final_label.fraud_type
        
        # Compute confidence statistics
        confidences = [l.confidence for l in accepted_labels]
        if confidences:
            result.mean_confidence = float(np.mean(confidences))
            result.low_confidence_count = len(needs_review)
        
        return samples
    
    def _consensus_workflow(
        self,
        samples: List[IdentitySample],
        human_labels_path: Optional[str],
        result: LabelingResult
    ) -> List[IdentitySample]:
        """Workflow using multi-annotator consensus."""
        if not human_labels_path or not self.consensus_labeler:
            # Fall back to hybrid if no consensus data
            return self._hybrid_workflow(samples, human_labels_path, result)
        
        # Load multiple label batches
        label_batches = []
        
        # Check if path is a directory with multiple label files
        path = Path(human_labels_path)
        if path.is_dir():
            reader = CSVLabelReader()
            for label_file in path.glob("*.csv"):
                batch = reader.read(str(label_file))
                label_batches.append(batch)
        else:
            reader = CSVLabelReader()
            batch = reader.read(human_labels_path)
            label_batches.append(batch)
        
        if len(label_batches) < self.config.min_annotators:
            result.warnings.append(
                f"Only {len(label_batches)} annotators, need {self.config.min_annotators}"
            )
            return self._hybrid_workflow(samples, human_labels_path, result)
        
        # Merge to consensus
        consensus_batch = self.consensus_labeler.merge(label_batches)
        
        # Apply consensus labels
        label_map = {l.sample_id: l for l in consensus_batch}
        
        for sample in samples:
            if sample.sample_id in label_map:
                label = label_map[sample.sample_id]
                sample.trust_score = label.trust_score
                sample.is_genuine = label.is_genuine
                sample.fraud_type = label.fraud_type
                result.human_labeled += 1
            else:
                # Fall back to heuristic for missing samples
                heuristic_label = self.auto_labeler.heuristic_labeler.label_sample(sample)
                sample.trust_score = heuristic_label.trust_score
                sample.is_genuine = heuristic_label.is_genuine
                sample.fraud_type = heuristic_label.fraud_type
                result.auto_labeled += 1
        
        return samples
    
    def _create_dataset(self, samples: List[IdentitySample]) -> Dataset:
        """Create dataset with splits."""
        from ml.training.dataset.splits import (
            DeterministicSplitter,
            SplitConfig,
            SplitStrategy,
        )
        
        config = SplitConfig(
            train_ratio=0.7,
            val_ratio=0.15,
            test_ratio=0.15,
            random_seed=self.config.random_seed,
            strategy=SplitStrategy.STRATIFIED,
        )
        
        splitter = DeterministicSplitter(config)
        result = splitter.split(samples)
        
        return result.dataset
    
    def export_for_review(self, output_path: str) -> int:
        """Export pending samples for human review."""
        return self.review_queue.export_for_review(output_path)
    
    def get_review_queue(self) -> ReviewQueue:
        """Get the review queue for inspection."""
        return self.review_queue


def label_synthetic_dataset(
    samples: List[IdentitySample],
    seed: int = 42,
) -> Dataset:
    """
    Convenience function to label synthetic data for CI testing.
    
    Args:
        samples: Synthetic samples to label
        seed: Random seed for reproducibility
        
    Returns:
        Labeled Dataset
    """
    config = LabelingConfig(
        workflow=LabelingWorkflow.SYNTHETIC,
        random_seed=seed,
    )
    
    pipeline = ProductionLabelingPipeline(config)
    dataset, _ = pipeline.label_dataset(samples)
    
    return dataset
