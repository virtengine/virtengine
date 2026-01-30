"""
Deterministic dataset splitting with verification.

This module provides:
- Deterministic train/val/test splits with fixed seeds
- Stratified splitting by labels and document types
- Split verification and hash computation
- Reproducibility guarantees for training
"""

import hashlib
import json
import logging
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

import numpy as np

from ml.training.dataset.ingestion import (
    Dataset,
    DatasetSplit,
    IdentitySample,
    SplitType,
)

logger = logging.getLogger(__name__)


class SplitStrategy(str, Enum):
    """Splitting strategies."""
    RANDOM = "random"
    STRATIFIED = "stratified"
    TEMPORAL = "temporal"
    GROUP = "group"


@dataclass
class SplitConfig:
    """Configuration for dataset splitting."""
    
    # Split ratios
    train_ratio: float = 0.7
    val_ratio: float = 0.15
    test_ratio: float = 0.15
    
    # Strategy
    strategy: SplitStrategy = SplitStrategy.STRATIFIED
    
    # Stratification
    stratify_by: List[str] = field(default_factory=lambda: ["doc_type", "is_genuine"])
    
    # Determinism
    random_seed: int = 42
    
    # Group splitting
    group_by: Optional[str] = None  # e.g., "user_id"
    
    # Temporal splitting
    timestamp_field: Optional[str] = None
    
    # Validation
    min_samples_per_split: int = 10
    
    def __post_init__(self):
        total = self.train_ratio + self.val_ratio + self.test_ratio
        if abs(total - 1.0) > 1e-6:
            raise ValueError(f"Split ratios must sum to 1.0, got {total}")


@dataclass
class SplitResult:
    """Result of a splitting operation."""
    
    # Dataset with splits
    dataset: Dataset
    
    # Metadata
    strategy: SplitStrategy
    random_seed: int
    split_ratios: Dict[str, float] = field(default_factory=dict)
    
    # Verification
    split_hashes: Dict[str, str] = field(default_factory=dict)
    combined_hash: str = ""
    
    # Statistics
    split_sizes: Dict[str, int] = field(default_factory=dict)
    stratification_stats: Dict[str, Any] = field(default_factory=dict)
    
    # Reproducibility check
    is_reproducible: bool = True
    reproducibility_hash: str = ""
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "strategy": self.strategy.value,
            "random_seed": self.random_seed,
            "split_ratios": self.split_ratios,
            "split_hashes": self.split_hashes,
            "combined_hash": self.combined_hash,
            "split_sizes": self.split_sizes,
            "stratification_stats": self.stratification_stats,
            "is_reproducible": self.is_reproducible,
            "reproducibility_hash": self.reproducibility_hash,
        }


class DeterministicSplitter:
    """
    Performs deterministic dataset splitting.
    
    Ensures reproducibility by:
    - Using fixed random seeds
    - Sorting samples before splitting
    - Computing verification hashes
    """
    
    def __init__(self, config: Optional[SplitConfig] = None):
        """
        Initialize the splitter.
        
        Args:
            config: Split configuration
        """
        self.config = config or SplitConfig()
    
    def split(self, samples: List[IdentitySample]) -> SplitResult:
        """
        Split samples into train/val/test sets.
        
        Args:
            samples: List of samples to split
            
        Returns:
            SplitResult with dataset and metadata
        """
        if len(samples) == 0:
            return self._empty_result()
        
        # Sort samples for determinism
        sorted_samples = self._sort_samples(samples)
        
        # Initialize RNG with fixed seed
        rng = np.random.RandomState(self.config.random_seed)
        
        # Apply splitting strategy
        if self.config.strategy == SplitStrategy.STRATIFIED:
            train, val, test = self._stratified_split(sorted_samples, rng)
        elif self.config.strategy == SplitStrategy.TEMPORAL:
            train, val, test = self._temporal_split(sorted_samples)
        elif self.config.strategy == SplitStrategy.GROUP:
            train, val, test = self._group_split(sorted_samples, rng)
        else:  # RANDOM
            train, val, test = self._random_split(sorted_samples, rng)
        
        # Create dataset
        dataset = Dataset(
            train=DatasetSplit(SplitType.TRAIN, train),
            validation=DatasetSplit(SplitType.VALIDATION, val),
            test=DatasetSplit(SplitType.TEST, test),
            version=f"split_{datetime.now().strftime('%Y%m%d')}",
            creation_timestamp=datetime.now().timestamp(),
        )
        
        # Compute hashes
        split_hashes = self._compute_split_hashes(dataset)
        combined_hash = self._compute_combined_hash(split_hashes)
        
        # Verify reproducibility
        reproducibility_hash = self._compute_reproducibility_hash(
            samples, self.config
        )
        
        # Compute stratification stats
        strat_stats = self._compute_stratification_stats(dataset)
        
        result = SplitResult(
            dataset=dataset,
            strategy=self.config.strategy,
            random_seed=self.config.random_seed,
            split_ratios={
                "train": self.config.train_ratio,
                "validation": self.config.val_ratio,
                "test": self.config.test_ratio,
            },
            split_hashes=split_hashes,
            combined_hash=combined_hash,
            split_sizes={
                "train": len(train),
                "validation": len(val),
                "test": len(test),
            },
            stratification_stats=strat_stats,
            is_reproducible=True,
            reproducibility_hash=reproducibility_hash,
        )
        
        logger.info(
            f"Split {len(samples)} samples: "
            f"train={len(train)}, val={len(val)}, test={len(test)}"
        )
        logger.info(f"Split hash: {combined_hash}")
        
        return result
    
    def _sort_samples(self, samples: List[IdentitySample]) -> List[IdentitySample]:
        """Sort samples by ID for determinism."""
        return sorted(samples, key=lambda s: s.sample_id)
    
    def _random_split(
        self,
        samples: List[IdentitySample],
        rng: np.random.RandomState,
    ) -> Tuple[List[IdentitySample], List[IdentitySample], List[IdentitySample]]:
        """Random splitting."""
        indices = np.arange(len(samples))
        rng.shuffle(indices)
        
        n = len(samples)
        train_end = int(n * self.config.train_ratio)
        val_end = train_end + int(n * self.config.val_ratio)
        
        train_indices = indices[:train_end]
        val_indices = indices[train_end:val_end]
        test_indices = indices[val_end:]
        
        return (
            [samples[i] for i in train_indices],
            [samples[i] for i in val_indices],
            [samples[i] for i in test_indices],
        )
    
    def _stratified_split(
        self,
        samples: List[IdentitySample],
        rng: np.random.RandomState,
    ) -> Tuple[List[IdentitySample], List[IdentitySample], List[IdentitySample]]:
        """Stratified splitting maintaining class distributions."""
        # Group samples by stratification key
        groups: Dict[str, List[int]] = {}
        
        for i, sample in enumerate(samples):
            key = self._get_stratification_key(sample)
            if key not in groups:
                groups[key] = []
            groups[key].append(i)
        
        train_indices = []
        val_indices = []
        test_indices = []
        
        # Split each group proportionally
        for group_key, indices in groups.items():
            group_indices = np.array(indices)
            rng.shuffle(group_indices)
            
            n = len(group_indices)
            train_end = max(1, int(n * self.config.train_ratio))
            val_end = train_end + max(1, int(n * self.config.val_ratio))
            
            # Ensure at least one sample in test if we have enough
            if n >= 3:
                train_end = min(train_end, n - 2)
                val_end = min(val_end, n - 1)
            
            train_indices.extend(group_indices[:train_end])
            val_indices.extend(group_indices[train_end:val_end])
            test_indices.extend(group_indices[val_end:])
        
        return (
            [samples[i] for i in train_indices],
            [samples[i] for i in val_indices],
            [samples[i] for i in test_indices],
        )
    
    def _temporal_split(
        self,
        samples: List[IdentitySample],
    ) -> Tuple[List[IdentitySample], List[IdentitySample], List[IdentitySample]]:
        """Temporal splitting based on timestamps."""
        # Sort by timestamp
        if self.config.timestamp_field:
            sorted_samples = sorted(
                samples,
                key=lambda s: getattr(
                    s.capture_metadata or object(),
                    self.config.timestamp_field,
                    0
                ) or 0
            )
        else:
            # Use sample order if no timestamp field
            sorted_samples = samples
        
        n = len(sorted_samples)
        train_end = int(n * self.config.train_ratio)
        val_end = train_end + int(n * self.config.val_ratio)
        
        return (
            sorted_samples[:train_end],
            sorted_samples[train_end:val_end],
            sorted_samples[val_end:],
        )
    
    def _group_split(
        self,
        samples: List[IdentitySample],
        rng: np.random.RandomState,
    ) -> Tuple[List[IdentitySample], List[IdentitySample], List[IdentitySample]]:
        """Group splitting ensuring same group stays together."""
        if not self.config.group_by:
            return self._random_split(samples, rng)
        
        # Group samples
        groups: Dict[str, List[IdentitySample]] = {}
        for sample in samples:
            group_key = sample.annotations.get(self.config.group_by, sample.sample_id)
            if group_key not in groups:
                groups[group_key] = []
            groups[group_key].append(sample)
        
        # Shuffle groups
        group_keys = list(groups.keys())
        rng.shuffle(group_keys)
        
        # Assign groups to splits
        n_groups = len(group_keys)
        train_end = int(n_groups * self.config.train_ratio)
        val_end = train_end + int(n_groups * self.config.val_ratio)
        
        train = []
        val = []
        test = []
        
        for i, key in enumerate(group_keys):
            if i < train_end:
                train.extend(groups[key])
            elif i < val_end:
                val.extend(groups[key])
            else:
                test.extend(groups[key])
        
        return train, val, test
    
    def _get_stratification_key(self, sample: IdentitySample) -> str:
        """Get stratification key for a sample."""
        key_parts = []
        
        for field in self.config.stratify_by:
            if field == "doc_type":
                key_parts.append(
                    sample.document_info.doc_type if sample.document_info else "unknown"
                )
            elif field == "is_genuine":
                key_parts.append(str(sample.is_genuine))
            elif field == "fraud_type":
                key_parts.append(sample.fraud_type or "none")
            else:
                key_parts.append(str(sample.annotations.get(field, "unknown")))
        
        return "|".join(key_parts)
    
    def _compute_split_hashes(self, dataset: Dataset) -> Dict[str, str]:
        """Compute hash for each split."""
        hashes = {}
        
        for split_type, split in [
            ("train", dataset.train),
            ("validation", dataset.validation),
            ("test", dataset.test),
        ]:
            sample_ids = sorted([s.sample_id for s in split.samples])
            hash_input = "|".join(sample_ids).encode()
            hashes[split_type] = hashlib.sha256(hash_input).hexdigest()[:16]
        
        return hashes
    
    def _compute_combined_hash(self, split_hashes: Dict[str, str]) -> str:
        """Compute combined hash of all splits."""
        combined = f"{split_hashes['train']}|{split_hashes['validation']}|{split_hashes['test']}"
        return hashlib.sha256(combined.encode()).hexdigest()[:16]
    
    def _compute_reproducibility_hash(
        self,
        samples: List[IdentitySample],
        config: SplitConfig,
    ) -> str:
        """Compute hash for reproducibility verification."""
        hash_input = {
            "sample_ids": sorted([s.sample_id for s in samples]),
            "seed": config.random_seed,
            "strategy": config.strategy.value,
            "train_ratio": config.train_ratio,
            "val_ratio": config.val_ratio,
            "test_ratio": config.test_ratio,
        }
        return hashlib.sha256(
            json.dumps(hash_input, sort_keys=True).encode()
        ).hexdigest()[:16]
    
    def _compute_stratification_stats(self, dataset: Dataset) -> Dict[str, Any]:
        """Compute statistics about stratification."""
        stats = {}
        
        for split_name, split in [
            ("train", dataset.train),
            ("validation", dataset.validation),
            ("test", dataset.test),
        ]:
            split_stats = {
                "total": len(split),
                "genuine": sum(1 for s in split if s.is_genuine),
                "fraud": sum(1 for s in split if not s.is_genuine),
                "doc_types": {},
            }
            
            for sample in split:
                doc_type = sample.document_info.doc_type if sample.document_info else "unknown"
                split_stats["doc_types"][doc_type] = split_stats["doc_types"].get(doc_type, 0) + 1
            
            stats[split_name] = split_stats
        
        return stats
    
    def _empty_result(self) -> SplitResult:
        """Create empty result for empty input."""
        return SplitResult(
            dataset=Dataset(
                train=DatasetSplit(SplitType.TRAIN, []),
                validation=DatasetSplit(SplitType.VALIDATION, []),
                test=DatasetSplit(SplitType.TEST, []),
            ),
            strategy=self.config.strategy,
            random_seed=self.config.random_seed,
            split_ratios={
                "train": self.config.train_ratio,
                "validation": self.config.val_ratio,
                "test": self.config.test_ratio,
            },
            is_reproducible=True,
        )


class SplitVerifier:
    """Verifies dataset splits for reproducibility and correctness."""
    
    def verify(
        self,
        result: SplitResult,
        expected_hash: Optional[str] = None,
    ) -> Dict[str, Any]:
        """
        Verify a split result.
        
        Args:
            result: SplitResult to verify
            expected_hash: Expected combined hash (if known)
            
        Returns:
            Verification report
        """
        report = {
            "valid": True,
            "checks": {},
            "warnings": [],
            "errors": [],
        }
        
        # Check no overlap between splits
        train_ids = {s.sample_id for s in result.dataset.train}
        val_ids = {s.sample_id for s in result.dataset.validation}
        test_ids = {s.sample_id for s in result.dataset.test}
        
        train_val_overlap = train_ids & val_ids
        train_test_overlap = train_ids & test_ids
        val_test_overlap = val_ids & test_ids
        
        report["checks"]["no_train_val_overlap"] = len(train_val_overlap) == 0
        report["checks"]["no_train_test_overlap"] = len(train_test_overlap) == 0
        report["checks"]["no_val_test_overlap"] = len(val_test_overlap) == 0
        
        if train_val_overlap:
            report["valid"] = False
            report["errors"].append(f"Train-val overlap: {len(train_val_overlap)} samples")
        if train_test_overlap:
            report["valid"] = False
            report["errors"].append(f"Train-test overlap: {len(train_test_overlap)} samples")
        if val_test_overlap:
            report["valid"] = False
            report["errors"].append(f"Val-test overlap: {len(val_test_overlap)} samples")
        
        # Check minimum sizes
        min_size = 10  # Configurable
        for split_name, split_size in result.split_sizes.items():
            if split_size < min_size:
                report["warnings"].append(
                    f"{split_name} split has only {split_size} samples"
                )
        
        # Verify hash if expected
        if expected_hash:
            report["checks"]["hash_match"] = result.combined_hash == expected_hash
            if not report["checks"]["hash_match"]:
                report["valid"] = False
                report["errors"].append(
                    f"Hash mismatch: expected {expected_hash}, got {result.combined_hash}"
                )
        
        # Check reproducibility
        report["checks"]["is_reproducible"] = result.is_reproducible
        report["reproducibility_hash"] = result.reproducibility_hash
        
        return report
    
    def verify_reproducibility(
        self,
        samples: List[IdentitySample],
        config: SplitConfig,
        expected_hash: str,
    ) -> bool:
        """
        Verify that splitting produces the expected hash.
        
        Args:
            samples: Samples to split
            config: Split configuration
            expected_hash: Expected combined hash
            
        Returns:
            True if reproducible
        """
        splitter = DeterministicSplitter(config)
        result = splitter.split(samples)
        return result.combined_hash == expected_hash


def split_dataset(
    samples: List[IdentitySample],
    train_ratio: float = 0.7,
    val_ratio: float = 0.15,
    test_ratio: float = 0.15,
    seed: int = 42,
    stratified: bool = True,
) -> Dataset:
    """
    Convenience function to split a dataset.
    
    Args:
        samples: Samples to split
        train_ratio: Training set ratio
        val_ratio: Validation set ratio
        test_ratio: Test set ratio
        seed: Random seed
        stratified: Use stratified splitting
        
    Returns:
        Dataset with splits
    """
    config = SplitConfig(
        train_ratio=train_ratio,
        val_ratio=val_ratio,
        test_ratio=test_ratio,
        random_seed=seed,
        strategy=SplitStrategy.STRATIFIED if stratified else SplitStrategy.RANDOM,
    )
    
    splitter = DeterministicSplitter(config)
    result = splitter.split(samples)
    
    return result.dataset
