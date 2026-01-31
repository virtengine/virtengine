"""
Dataset versioning and provenance tracking for VEID ML training.

This module provides production-ready versioning with:
- Semantic versioning for datasets
- Provenance tracking (sources, transforms, build info)
- Content hash verification
- Version comparison and diff
- Reproducibility guarantees

All versioned datasets include full provenance for audit and reproducibility.
"""

import hashlib
import json
import logging
import os
import shutil
import subprocess
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
from ml.training.dataset.lineage import (
    BuildInfo,
    DatasetLineage,
    LineageTracker,
    SourceRecord,
    SourceType,
    TransformRecord,
    TransformType,
)
from ml.training.dataset.manifest import (
    DatasetManifest,
    ManifestBuilder,
    ManifestSigner,
)

logger = logging.getLogger(__name__)


class VersionStatus(str, Enum):
    """Status of a dataset version."""
    DRAFT = "draft"  # Being built
    PENDING = "pending"  # Awaiting validation
    VALIDATED = "validated"  # Passed validation
    RELEASED = "released"  # Production-ready
    DEPRECATED = "deprecated"  # No longer recommended
    ARCHIVED = "archived"  # Historical only


@dataclass
class ProvenanceRecord:
    """Complete provenance record for a dataset version."""
    
    # Version info
    version: str
    created_at: float = field(default_factory=lambda: datetime.now().timestamp())
    
    # Lineage
    lineage: Optional[DatasetLineage] = None
    
    # Build environment
    build_info: Optional[BuildInfo] = None
    
    # Parent version (if derived)
    parent_version: Optional[str] = None
    
    # Content integrity
    content_hash: str = ""
    manifest_hash: str = ""
    
    # Status
    status: VersionStatus = VersionStatus.DRAFT
    
    # Metadata
    description: str = ""
    tags: List[str] = field(default_factory=list)
    metadata: Dict[str, Any] = field(default_factory=dict)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "version": self.version,
            "created_at": self.created_at,
            "created_at_iso": datetime.fromtimestamp(self.created_at).isoformat(),
            "lineage": self.lineage.to_dict() if self.lineage else None,
            "build_info": self.build_info.to_dict() if self.build_info else None,
            "parent_version": self.parent_version,
            "content_hash": self.content_hash,
            "manifest_hash": self.manifest_hash,
            "status": self.status.value,
            "description": self.description,
            "tags": self.tags,
            "metadata": self.metadata,
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "ProvenanceRecord":
        """Create from dictionary."""
        record = cls(
            version=data["version"],
            created_at=data.get("created_at", 0),
            parent_version=data.get("parent_version"),
            content_hash=data.get("content_hash", ""),
            manifest_hash=data.get("manifest_hash", ""),
            status=VersionStatus(data.get("status", "draft")),
            description=data.get("description", ""),
            tags=data.get("tags", []),
            metadata=data.get("metadata", {}),
        )
        
        if data.get("lineage"):
            record.lineage = DatasetLineage.from_dict(data["lineage"])
        
        if data.get("build_info"):
            bi = data["build_info"]
            record.build_info = BuildInfo(
                git_commit=bi.get("git_commit", ""),
                git_branch=bi.get("git_branch", ""),
                git_dirty=bi.get("git_dirty", False),
                python_version=bi.get("python_version", ""),
                numpy_version=bi.get("numpy_version", ""),
                tensorflow_version=bi.get("tensorflow_version", ""),
                hostname=bi.get("hostname", ""),
                platform=bi.get("platform", ""),
            )
        
        return record


@dataclass
class DatasetVersion:
    """A versioned dataset with provenance."""
    
    # Dataset
    dataset: Dataset
    
    # Provenance
    provenance: ProvenanceRecord
    
    # Manifest
    manifest: Optional[DatasetManifest] = None
    
    # Storage path
    storage_path: Optional[str] = None
    
    def __len__(self) -> int:
        return len(self.dataset)
    
    @property
    def version(self) -> str:
        return self.provenance.version
    
    @property
    def content_hash(self) -> str:
        return self.provenance.content_hash
    
    def save(self, base_path: str) -> str:
        """
        Save versioned dataset to disk.
        
        Args:
            base_path: Base directory for storage
            
        Returns:
            Path to saved version
        """
        version_dir = Path(base_path) / self.provenance.version
        version_dir.mkdir(parents=True, exist_ok=True)
        
        # Save dataset samples as JSON
        dataset_data = {
            "version": self.provenance.version,
            "schema_version": self.provenance.lineage.schema_version if self.provenance.lineage else "1.0.0",
            "splits": {
                "train": len(self.dataset.train),
                "validation": len(self.dataset.validation),
                "test": len(self.dataset.test),
            },
            "content_hash": self.provenance.content_hash,
            "samples": [],
        }
        
        for split_name, split in [
            ("train", self.dataset.train),
            ("validation", self.dataset.validation),
            ("test", self.dataset.test),
        ]:
            for sample in split:
                sample_data = sample.to_dict()
                sample_data["split"] = split_name
                dataset_data["samples"].append(sample_data)
        
        with open(version_dir / "dataset.json", "w") as f:
            json.dump(dataset_data, f, indent=2)
        
        # Save provenance
        with open(version_dir / "provenance.json", "w") as f:
            json.dump(self.provenance.to_dict(), f, indent=2)
        
        # Save manifest if present
        if self.manifest:
            self.manifest.save(str(version_dir / "manifest.json"))
        
        # Save lineage if present
        if self.provenance.lineage:
            self.provenance.lineage.save(str(version_dir / "lineage.json"))
        
        self.storage_path = str(version_dir)
        
        logger.info(f"Saved dataset version {self.provenance.version} to {version_dir}")
        
        return str(version_dir)
    
    @classmethod
    def load(cls, version_path: str) -> "DatasetVersion":
        """
        Load versioned dataset from disk.
        
        Args:
            version_path: Path to version directory
            
        Returns:
            DatasetVersion
        """
        version_dir = Path(version_path)
        
        # Load dataset
        with open(version_dir / "dataset.json") as f:
            dataset_data = json.load(f)
        
        # Reconstruct samples
        samples_by_split: Dict[str, List[IdentitySample]] = {
            "train": [],
            "validation": [],
            "test": [],
        }
        
        for sample_data in dataset_data.get("samples", []):
            split_name = sample_data.pop("split", "train")
            sample = cls._sample_from_dict(sample_data)
            samples_by_split[split_name].append(sample)
        
        dataset = Dataset(
            train=DatasetSplit(SplitType.TRAIN, samples_by_split["train"]),
            validation=DatasetSplit(SplitType.VALIDATION, samples_by_split["validation"]),
            test=DatasetSplit(SplitType.TEST, samples_by_split["test"]),
            version=dataset_data.get("version", ""),
            dataset_hash=dataset_data.get("content_hash", ""),
        )
        
        # Load provenance
        with open(version_dir / "provenance.json") as f:
            provenance_data = json.load(f)
        provenance = ProvenanceRecord.from_dict(provenance_data)
        
        # Load manifest if present
        manifest = None
        manifest_path = version_dir / "manifest.json"
        if manifest_path.exists():
            manifest = DatasetManifest.load(str(manifest_path))
        
        return cls(
            dataset=dataset,
            provenance=provenance,
            manifest=manifest,
            storage_path=str(version_dir),
        )
    
    @staticmethod
    def _sample_from_dict(data: Dict[str, Any]) -> IdentitySample:
        """Reconstruct sample from dictionary."""
        from ml.training.dataset.ingestion import DocumentInfo
        
        doc_info = None
        if data.get("document_type"):
            doc_info = DocumentInfo(
                doc_type=data.get("document_type", ""),
                doc_id_hash=data.get("doc_id_hash", ""),
            )
        
        return IdentitySample(
            sample_id=data.get("sample_id", ""),
            document_info=doc_info,
            trust_score=float(data.get("trust_score", 0)),
            is_genuine=data.get("is_genuine", True),
            fraud_type=data.get("fraud_type"),
            face_detected=data.get("face_detected", True),
            face_confidence=float(data.get("face_confidence", 0)),
            document_quality_score=float(data.get("document_quality_score", 0)),
            ocr_success=data.get("ocr_success", True),
            ocr_confidence=float(data.get("ocr_confidence", 0)),
        )


@dataclass
class VersionConfig:
    """Configuration for dataset versioning."""
    
    # Storage
    base_path: str = "data/versions"
    
    # Versioning
    version_format: str = "v{major}.{minor}.{patch}"
    auto_increment: bool = True
    
    # Signing
    sign_manifests: bool = True
    signer_id: str = "dataset-builder"
    
    # Retention
    keep_versions: int = 10  # Number of versions to retain
    archive_older: bool = True
    
    # Validation
    validate_on_save: bool = True
    require_validation: bool = True


class DatasetVersionManager:
    """
    Manages versioned datasets with provenance tracking.
    
    Features:
    - Semantic versioning
    - Full provenance chain
    - Content integrity verification
    - Version comparison and diff
    - Reproducibility guarantees
    
    Usage:
        manager = DatasetVersionManager(VersionConfig(base_path="data/versions"))
        
        # Create new version
        version = manager.create_version(dataset, "1.0.0", description="Initial release")
        
        # Load existing version
        version = manager.load_version("1.0.0")
        
        # List versions
        versions = manager.list_versions()
    """
    
    def __init__(self, config: Optional[VersionConfig] = None):
        """
        Initialize version manager.
        
        Args:
            config: Version configuration
        """
        self.config = config or VersionConfig()
        self._versions: Dict[str, ProvenanceRecord] = {}
        
        # Ensure base path exists
        Path(self.config.base_path).mkdir(parents=True, exist_ok=True)
        
        # Load existing versions
        self._load_version_index()
    
    def create_version(
        self,
        dataset: Dataset,
        version: Optional[str] = None,
        description: str = "",
        tags: Optional[List[str]] = None,
        parent_version: Optional[str] = None,
        lineage: Optional[DatasetLineage] = None,
        metadata: Optional[Dict[str, Any]] = None,
    ) -> DatasetVersion:
        """
        Create a new versioned dataset.
        
        Args:
            dataset: Dataset to version
            version: Version string (auto-generated if not provided)
            description: Version description
            tags: Version tags
            parent_version: Parent version if derived
            lineage: Dataset lineage
            metadata: Additional metadata
            
        Returns:
            DatasetVersion
        """
        # Generate version if not provided
        if version is None:
            version = self._generate_version(parent_version)
        
        # Validate version format
        if version in self._versions:
            raise ValueError(f"Version {version} already exists")
        
        # Compute content hash
        content_hash = self._compute_content_hash(dataset)
        
        # Capture build info
        build_info = BuildInfo.capture()
        
        # Create provenance record
        provenance = ProvenanceRecord(
            version=version,
            lineage=lineage,
            build_info=build_info,
            parent_version=parent_version,
            content_hash=content_hash,
            status=VersionStatus.DRAFT,
            description=description,
            tags=tags or [],
            metadata=metadata or {},
        )
        
        # Build manifest
        manifest_builder = ManifestBuilder(
            dataset_name="veid_trust",
            schema_version=lineage.schema_version if lineage else "1.0.0",
        )
        
        for split_name, split in [
            ("train", dataset.train),
            ("validation", dataset.validation),
            ("test", dataset.test),
        ]:
            for sample in split:
                manifest_builder.add_sample(
                    sample.sample_id,
                    json.dumps(sample.to_dict()).encode(),
                    split_name,
                )
        
        manifest = manifest_builder.build(version)
        
        # Sign manifest if configured
        if self.config.sign_manifests:
            signer = ManifestSigner(self.config.signer_id)
            manifest = signer.sign(manifest)
        
        provenance.manifest_hash = manifest.manifest_hash
        
        # Create version object
        dataset_version = DatasetVersion(
            dataset=dataset,
            provenance=provenance,
            manifest=manifest,
        )
        
        # Save to disk
        dataset_version.save(self.config.base_path)
        
        # Update index
        self._versions[version] = provenance
        self._save_version_index()
        
        logger.info(f"Created dataset version {version} with hash {content_hash}")
        
        return dataset_version
    
    def load_version(self, version: str) -> DatasetVersion:
        """
        Load a specific version.
        
        Args:
            version: Version string to load
            
        Returns:
            DatasetVersion
        """
        if version not in self._versions:
            raise ValueError(f"Version {version} not found")
        
        version_path = Path(self.config.base_path) / version
        return DatasetVersion.load(str(version_path))
    
    def load_latest(self) -> Optional[DatasetVersion]:
        """Load the latest version."""
        versions = self.list_versions()
        if not versions:
            return None
        
        return self.load_version(versions[-1].version)
    
    def list_versions(
        self,
        status: Optional[VersionStatus] = None,
        tags: Optional[List[str]] = None,
    ) -> List[ProvenanceRecord]:
        """
        List available versions.
        
        Args:
            status: Filter by status
            tags: Filter by tags (any match)
            
        Returns:
            List of ProvenanceRecords
        """
        versions = list(self._versions.values())
        
        if status:
            versions = [v for v in versions if v.status == status]
        
        if tags:
            versions = [
                v for v in versions
                if any(t in v.tags for t in tags)
            ]
        
        # Sort by creation time
        versions.sort(key=lambda v: v.created_at)
        
        return versions
    
    def get_version_info(self, version: str) -> Optional[ProvenanceRecord]:
        """Get provenance record for a version."""
        return self._versions.get(version)
    
    def update_status(self, version: str, status: VersionStatus) -> None:
        """Update the status of a version."""
        if version not in self._versions:
            raise ValueError(f"Version {version} not found")
        
        self._versions[version].status = status
        
        # Update stored provenance
        version_path = Path(self.config.base_path) / version
        provenance_path = version_path / "provenance.json"
        
        with open(provenance_path, "w") as f:
            json.dump(self._versions[version].to_dict(), f, indent=2)
        
        self._save_version_index()
        
        logger.info(f"Updated version {version} status to {status.value}")
    
    def verify_version(self, version: str) -> Dict[str, Any]:
        """
        Verify integrity of a version.
        
        Args:
            version: Version to verify
            
        Returns:
            Verification report
        """
        if version not in self._versions:
            raise ValueError(f"Version {version} not found")
        
        report = {
            "version": version,
            "valid": True,
            "checks": {},
            "errors": [],
        }
        
        # Load version
        dataset_version = self.load_version(version)
        
        # Check content hash
        current_hash = self._compute_content_hash(dataset_version.dataset)
        expected_hash = dataset_version.provenance.content_hash
        
        report["checks"]["content_hash"] = current_hash == expected_hash
        if not report["checks"]["content_hash"]:
            report["valid"] = False
            report["errors"].append(
                f"Content hash mismatch: expected {expected_hash}, got {current_hash}"
            )
        
        # Check manifest hash
        if dataset_version.manifest:
            manifest_hash = dataset_version.manifest.manifest_hash
            expected_manifest = dataset_version.provenance.manifest_hash
            
            report["checks"]["manifest_hash"] = manifest_hash == expected_manifest
            if not report["checks"]["manifest_hash"]:
                report["valid"] = False
                report["errors"].append(
                    f"Manifest hash mismatch: expected {expected_manifest}, got {manifest_hash}"
                )
        
        # Check sample counts
        total_samples = len(dataset_version.dataset)
        if dataset_version.provenance.lineage:
            expected_count = dataset_version.provenance.lineage.sample_count
            report["checks"]["sample_count"] = total_samples == expected_count
            if not report["checks"]["sample_count"]:
                report["errors"].append(
                    f"Sample count mismatch: expected {expected_count}, got {total_samples}"
                )
        
        return report
    
    def compare_versions(
        self,
        version_a: str,
        version_b: str,
    ) -> Dict[str, Any]:
        """
        Compare two versions.
        
        Args:
            version_a: First version
            version_b: Second version
            
        Returns:
            Comparison report
        """
        v_a = self.load_version(version_a)
        v_b = self.load_version(version_b)
        
        # Get sample IDs
        ids_a = set(s.sample_id for s in (
            list(v_a.dataset.train) +
            list(v_a.dataset.validation) +
            list(v_a.dataset.test)
        ))
        ids_b = set(s.sample_id for s in (
            list(v_b.dataset.train) +
            list(v_b.dataset.validation) +
            list(v_b.dataset.test)
        ))
        
        return {
            "version_a": version_a,
            "version_b": version_b,
            "samples_a": len(ids_a),
            "samples_b": len(ids_b),
            "added": len(ids_b - ids_a),
            "removed": len(ids_a - ids_b),
            "unchanged": len(ids_a & ids_b),
            "content_hash_a": v_a.provenance.content_hash,
            "content_hash_b": v_b.provenance.content_hash,
        }
    
    def _generate_version(self, parent: Optional[str] = None) -> str:
        """Generate next version string."""
        if parent and parent in self._versions:
            # Parse parent version
            parts = parent.lstrip("v").split(".")
            if len(parts) == 3:
                major, minor, patch = map(int, parts)
                return f"v{major}.{minor}.{patch + 1}"
        
        # Find highest version
        existing = list(self._versions.keys())
        if not existing:
            return "v1.0.0"
        
        # Sort and get latest
        def version_key(v: str) -> Tuple[int, ...]:
            try:
                parts = v.lstrip("v").split(".")
                return tuple(int(p) for p in parts)
            except (ValueError, IndexError):
                return (0, 0, 0)
        
        existing.sort(key=version_key)
        latest = existing[-1]
        
        parts = latest.lstrip("v").split(".")
        if len(parts) == 3:
            major, minor, patch = map(int, parts)
            return f"v{major}.{minor}.{patch + 1}"
        
        return f"v1.0.{len(existing)}"
    
    def _compute_content_hash(self, dataset: Dataset) -> str:
        """Compute content hash for a dataset."""
        hash_data = []
        
        for split in [dataset.train, dataset.validation, dataset.test]:
            for sample in sorted(split.samples, key=lambda s: s.sample_id):
                hash_data.append(sample.sample_id)
                hash_data.append(str(sample.trust_score))
                hash_data.append(str(sample.is_genuine))
        
        combined = "|".join(hash_data).encode()
        return hashlib.sha256(combined).hexdigest()[:16]
    
    def _load_version_index(self) -> None:
        """Load version index from disk."""
        index_path = Path(self.config.base_path) / "versions.json"
        
        if index_path.exists():
            with open(index_path) as f:
                data = json.load(f)
            
            for version, record_data in data.items():
                self._versions[version] = ProvenanceRecord.from_dict(record_data)
    
    def _save_version_index(self) -> None:
        """Save version index to disk."""
        index_path = Path(self.config.base_path) / "versions.json"
        
        data = {
            version: record.to_dict()
            for version, record in self._versions.items()
        }
        
        with open(index_path, "w") as f:
            json.dump(data, f, indent=2)


def create_versioned_dataset(
    dataset: Dataset,
    version: str,
    base_path: str = "data/versions",
    description: str = "",
    lineage: Optional[DatasetLineage] = None,
) -> DatasetVersion:
    """
    Convenience function to create a versioned dataset.
    
    Args:
        dataset: Dataset to version
        version: Version string
        base_path: Base path for storage
        description: Version description
        lineage: Optional lineage information
        
    Returns:
        DatasetVersion
    """
    config = VersionConfig(base_path=base_path)
    manager = DatasetVersionManager(config)
    
    return manager.create_version(
        dataset=dataset,
        version=version,
        description=description,
        lineage=lineage,
    )
