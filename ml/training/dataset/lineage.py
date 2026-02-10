"""
Dataset lineage tracking for VEID training.

This module provides:
- Source tracking for all data
- Build version metadata
- Schema version tracking
- Transform chain recording
- Reproducibility guarantees
"""

import hashlib
import json
import logging
import os
import subprocess
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from pathlib import Path
from typing import Any, Dict, List, Optional

logger = logging.getLogger(__name__)


class SourceType(str, Enum):
    """Types of data sources."""
    LOCAL_FILE = "local_file"
    OBJECT_STORE = "object_store"
    API = "api"
    DATABASE = "database"
    SYNTHETIC = "synthetic"
    DERIVED = "derived"


class TransformType(str, Enum):
    """Types of data transforms."""
    INGESTION = "ingestion"
    PREPROCESSING = "preprocessing"
    AUGMENTATION = "augmentation"
    ANONYMIZATION = "anonymization"
    LABELING = "labeling"
    SPLITTING = "splitting"
    FEATURE_EXTRACTION = "feature_extraction"
    FILTERING = "filtering"
    MERGING = "merging"


@dataclass
class SourceRecord:
    """Record of a data source."""
    
    source_id: str
    source_type: SourceType
    location: str  # URI or path
    
    # Identification
    content_hash: Optional[str] = None
    record_count: int = 0
    
    # Timing
    accessed_at: float = field(default_factory=lambda: datetime.now().timestamp())
    
    # Metadata
    schema_version: Optional[str] = None
    format: str = ""  # "json", "csv", "tfrecord", etc.
    
    # Additional info
    metadata: Dict[str, Any] = field(default_factory=dict)
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "source_id": self.source_id,
            "source_type": self.source_type.value,
            "location": self.location,
            "content_hash": self.content_hash,
            "record_count": self.record_count,
            "accessed_at": self.accessed_at,
            "accessed_at_iso": datetime.fromtimestamp(self.accessed_at).isoformat(),
            "schema_version": self.schema_version,
            "format": self.format,
            "metadata": self.metadata,
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "SourceRecord":
        return cls(
            source_id=data["source_id"],
            source_type=SourceType(data["source_type"]),
            location=data["location"],
            content_hash=data.get("content_hash"),
            record_count=data.get("record_count", 0),
            accessed_at=data.get("accessed_at", 0),
            schema_version=data.get("schema_version"),
            format=data.get("format", ""),
            metadata=data.get("metadata", {}),
        )


@dataclass
class TransformRecord:
    """Record of a data transformation."""
    
    transform_id: str
    transform_type: TransformType
    description: str
    
    # Versioning
    tool_version: str = "1.0.0"
    config_hash: str = ""
    
    # Input/output
    input_hash: str = ""
    output_hash: str = ""
    input_count: int = 0
    output_count: int = 0
    
    # Timing
    started_at: float = field(default_factory=lambda: datetime.now().timestamp())
    completed_at: Optional[float] = None
    duration_seconds: float = 0.0
    
    # Configuration
    config: Dict[str, Any] = field(default_factory=dict)
    
    # Status
    success: bool = True
    error_message: Optional[str] = None
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "transform_id": self.transform_id,
            "transform_type": self.transform_type.value,
            "description": self.description,
            "tool_version": self.tool_version,
            "config_hash": self.config_hash,
            "input_hash": self.input_hash,
            "output_hash": self.output_hash,
            "input_count": self.input_count,
            "output_count": self.output_count,
            "started_at": self.started_at,
            "started_at_iso": datetime.fromtimestamp(self.started_at).isoformat(),
            "completed_at": self.completed_at,
            "duration_seconds": self.duration_seconds,
            "config": self.config,
            "success": self.success,
            "error_message": self.error_message,
        }


@dataclass
class BuildInfo:
    """Information about the build environment."""
    
    # Git info
    git_commit: str = ""
    git_branch: str = ""
    git_dirty: bool = False
    git_tags: List[str] = field(default_factory=list)
    
    # Tool versions
    python_version: str = ""
    numpy_version: str = ""
    tensorflow_version: str = ""
    tool_versions: Dict[str, str] = field(default_factory=dict)
    
    # Environment
    hostname: str = ""
    username: str = ""
    platform: str = ""
    
    # Build time
    build_timestamp: float = field(default_factory=lambda: datetime.now().timestamp())
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "git_commit": self.git_commit,
            "git_branch": self.git_branch,
            "git_dirty": self.git_dirty,
            "git_tags": self.git_tags,
            "python_version": self.python_version,
            "numpy_version": self.numpy_version,
            "tensorflow_version": self.tensorflow_version,
            "tool_versions": self.tool_versions,
            "hostname": self.hostname,
            "username": self.username,
            "platform": self.platform,
            "build_timestamp": self.build_timestamp,
            "build_timestamp_iso": datetime.fromtimestamp(self.build_timestamp).isoformat(),
        }
    
    @classmethod
    def capture(cls) -> "BuildInfo":
        """Capture current build environment info."""
        import platform
        import sys
        
        info = cls()
        
        # Python version
        info.python_version = sys.version.split()[0]
        info.platform = platform.platform()
        info.hostname = platform.node()
        
        try:
            info.username = os.getlogin()
        except Exception:
            info.username = os.environ.get("USER", os.environ.get("USERNAME", "unknown"))
        
        # Numpy version
        try:
            import numpy
            info.numpy_version = numpy.__version__
        except ImportError:
            pass
        
        # TensorFlow version
        try:
            import tensorflow
            info.tensorflow_version = tensorflow.__version__
        except ImportError:
            pass
        
        # Git info
        try:
            info.git_commit = subprocess.check_output(
                ["git", "rev-parse", "HEAD"],
                stderr=subprocess.DEVNULL,
            ).decode().strip()
            
            info.git_branch = subprocess.check_output(
                ["git", "rev-parse", "--abbrev-ref", "HEAD"],
                stderr=subprocess.DEVNULL,
            ).decode().strip()
            
            # Check if dirty
            status = subprocess.check_output(
                ["git", "status", "--porcelain"],
                stderr=subprocess.DEVNULL,
            ).decode().strip()
            info.git_dirty = len(status) > 0
            
            # Get tags
            try:
                tags = subprocess.check_output(
                    ["git", "tag", "--points-at", "HEAD"],
                    stderr=subprocess.DEVNULL,
                ).decode().strip()
                info.git_tags = [t for t in tags.split("\n") if t]
            except Exception:
                pass
                
        except Exception:
            pass
        
        return info


@dataclass
class DatasetLineage:
    """
    Complete lineage record for a dataset.
    
    Tracks:
    - All data sources
    - All transformations applied
    - Build environment
    - Schema versions
    """
    
    # Identification
    lineage_id: str = ""
    dataset_name: str = ""
    dataset_version: str = ""
    
    # Schema
    schema_version: str = "1.0.0"
    
    # Sources
    sources: List[SourceRecord] = field(default_factory=list)
    
    # Transforms
    transforms: List[TransformRecord] = field(default_factory=list)
    
    # Build info
    build_info: Optional[BuildInfo] = None
    
    # Final state
    final_hash: str = ""
    sample_count: int = 0
    
    # Metadata
    created_at: float = field(default_factory=lambda: datetime.now().timestamp())
    metadata: Dict[str, Any] = field(default_factory=dict)
    
    def add_source(self, source: SourceRecord) -> None:
        """Add a source record."""
        self.sources.append(source)
    
    def add_transform(self, transform: TransformRecord) -> None:
        """Add a transform record."""
        self.transforms.append(transform)
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "lineage_id": self.lineage_id,
            "dataset_name": self.dataset_name,
            "dataset_version": self.dataset_version,
            "schema_version": self.schema_version,
            "sources": [s.to_dict() for s in self.sources],
            "transforms": [t.to_dict() for t in self.transforms],
            "build_info": self.build_info.to_dict() if self.build_info else None,
            "final_hash": self.final_hash,
            "sample_count": self.sample_count,
            "created_at": self.created_at,
            "created_at_iso": datetime.fromtimestamp(self.created_at).isoformat(),
            "metadata": self.metadata,
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "DatasetLineage":
        lineage = cls(
            lineage_id=data.get("lineage_id", ""),
            dataset_name=data.get("dataset_name", ""),
            dataset_version=data.get("dataset_version", ""),
            schema_version=data.get("schema_version", "1.0.0"),
            final_hash=data.get("final_hash", ""),
            sample_count=data.get("sample_count", 0),
            created_at=data.get("created_at", 0),
            metadata=data.get("metadata", {}),
        )
        
        lineage.sources = [
            SourceRecord.from_dict(s) for s in data.get("sources", [])
        ]
        
        for t in data.get("transforms", []):
            t["transform_type"] = TransformType(t["transform_type"])
            lineage.transforms.append(TransformRecord(**{
                k: v for k, v in t.items()
                if k not in ["started_at_iso"]
            }))
        
        if data.get("build_info"):
            bi = data["build_info"]
            lineage.build_info = BuildInfo(
                git_commit=bi.get("git_commit", ""),
                git_branch=bi.get("git_branch", ""),
                git_dirty=bi.get("git_dirty", False),
                git_tags=bi.get("git_tags", []),
                python_version=bi.get("python_version", ""),
                numpy_version=bi.get("numpy_version", ""),
                tensorflow_version=bi.get("tensorflow_version", ""),
                tool_versions=bi.get("tool_versions", {}),
                hostname=bi.get("hostname", ""),
                username=bi.get("username", ""),
                platform=bi.get("platform", ""),
                build_timestamp=bi.get("build_timestamp", 0),
            )
        
        return lineage
    
    def save(self, path: str) -> None:
        """Save lineage to file."""
        file_path = Path(path)
        file_path.parent.mkdir(parents=True, exist_ok=True)
        
        with open(file_path, "w") as f:
            json.dump(self.to_dict(), f, indent=2)
        
        logger.info(f"Saved lineage to {path}")
    
    @classmethod
    def load(cls, path: str) -> "DatasetLineage":
        """Load lineage from file."""
        with open(path) as f:
            return cls.from_dict(json.load(f))
    
    def get_transform_chain(self) -> List[str]:
        """Get ordered list of transform descriptions."""
        return [t.description for t in self.transforms]
    
    def compute_hash(self) -> str:
        """Compute hash of the lineage."""
        hash_input = {
            "sources": [s.content_hash for s in self.sources if s.content_hash],
            "transforms": [t.config_hash for t in self.transforms],
            "schema_version": self.schema_version,
        }
        return hashlib.sha256(
            json.dumps(hash_input, sort_keys=True).encode()
        ).hexdigest()[:16]


class LineageTracker:
    """
    Tracks dataset lineage during build.
    
    Usage:
        tracker = LineageTracker("my_dataset", "1.0.0")
        
        with tracker.track_source("file://data/input") as source:
            data = load_data(...)
            source.record_count = len(data)
        
        with tracker.track_transform(TransformType.PREPROCESSING, "Resize images") as t:
            processed = preprocess(data)
            t.output_count = len(processed)
        
        lineage = tracker.finalize()
        lineage.save("lineage.json")
    """
    
    def __init__(
        self,
        dataset_name: str,
        dataset_version: str,
        schema_version: str = "1.0.0",
    ):
        """
        Initialize lineage tracker.
        
        Args:
            dataset_name: Name of the dataset
            dataset_version: Version of the dataset
            schema_version: Schema version
        """
        self.lineage = DatasetLineage(
            lineage_id=self._generate_id(dataset_name, dataset_version),
            dataset_name=dataset_name,
            dataset_version=dataset_version,
            schema_version=schema_version,
            build_info=BuildInfo.capture(),
        )
        
        self._current_transform: Optional[TransformRecord] = None
    
    def _generate_id(self, name: str, version: str) -> str:
        """Generate unique lineage ID."""
        timestamp = datetime.now().strftime("%Y%m%d%H%M%S")
        hash_input = f"{name}_{version}_{timestamp}"
        short_hash = hashlib.sha256(hash_input.encode()).hexdigest()[:8]
        return f"{name}_{version}_{timestamp}_{short_hash}"
    
    def add_source(
        self,
        location: str,
        source_type: SourceType = SourceType.LOCAL_FILE,
        content_hash: Optional[str] = None,
        record_count: int = 0,
        **metadata
    ) -> SourceRecord:
        """
        Add a data source.
        
        Args:
            location: Source URI or path
            source_type: Type of source
            content_hash: Hash of source content
            record_count: Number of records
            **metadata: Additional metadata
            
        Returns:
            SourceRecord
        """
        source = SourceRecord(
            source_id=f"source_{len(self.lineage.sources)}",
            source_type=source_type,
            location=location,
            content_hash=content_hash,
            record_count=record_count,
            schema_version=self.lineage.schema_version,
            metadata=metadata,
        )
        
        self.lineage.add_source(source)
        logger.debug(f"Added source: {location}")
        
        return source
    
    def add_transform(
        self,
        transform_type: TransformType,
        description: str,
        config: Optional[Dict[str, Any]] = None,
        input_hash: str = "",
        output_hash: str = "",
        input_count: int = 0,
        output_count: int = 0,
        tool_version: str = "1.0.0",
    ) -> TransformRecord:
        """
        Add a transform record.
        
        Args:
            transform_type: Type of transform
            description: Human-readable description
            config: Transform configuration
            input_hash: Hash of input data
            output_hash: Hash of output data
            input_count: Number of input items
            output_count: Number of output items
            tool_version: Version of transform tool
            
        Returns:
            TransformRecord
        """
        config = config or {}
        config_hash = hashlib.sha256(
            json.dumps(config, sort_keys=True).encode()
        ).hexdigest()[:16]
        
        transform = TransformRecord(
            transform_id=f"transform_{len(self.lineage.transforms)}",
            transform_type=transform_type,
            description=description,
            tool_version=tool_version,
            config_hash=config_hash,
            input_hash=input_hash,
            output_hash=output_hash,
            input_count=input_count,
            output_count=output_count,
            config=config,
            completed_at=datetime.now().timestamp(),
        )
        
        transform.duration_seconds = transform.completed_at - transform.started_at
        
        self.lineage.add_transform(transform)
        logger.debug(f"Added transform: {description}")
        
        return transform
    
    class TransformContext:
        """Context manager for tracking transforms."""
        
        def __init__(self, tracker: "LineageTracker", transform: TransformRecord):
            self.tracker = tracker
            self.transform = transform
        
        def __enter__(self) -> TransformRecord:
            return self.transform
        
        def __exit__(self, exc_type, exc_val, exc_tb) -> None:
            self.transform.completed_at = datetime.now().timestamp()
            self.transform.duration_seconds = (
                self.transform.completed_at - self.transform.started_at
            )
            
            if exc_type:
                self.transform.success = False
                self.transform.error_message = str(exc_val)
            
            self.tracker.lineage.add_transform(self.transform)
    
    def track_transform(
        self,
        transform_type: TransformType,
        description: str,
        config: Optional[Dict[str, Any]] = None,
    ) -> TransformContext:
        """
        Create a transform tracking context.
        
        Usage:
            with tracker.track_transform(TransformType.PREPROCESSING, "Normalize") as t:
                result = process(data)
                t.output_count = len(result)
        
        Args:
            transform_type: Type of transform
            description: Human-readable description
            config: Transform configuration
            
        Returns:
            TransformContext
        """
        config = config or {}
        config_hash = hashlib.sha256(
            json.dumps(config, sort_keys=True).encode()
        ).hexdigest()[:16]
        
        transform = TransformRecord(
            transform_id=f"transform_{len(self.lineage.transforms)}",
            transform_type=transform_type,
            description=description,
            config_hash=config_hash,
            config=config,
        )
        
        return self.TransformContext(self, transform)
    
    def finalize(
        self,
        final_hash: str = "",
        sample_count: int = 0,
    ) -> DatasetLineage:
        """
        Finalize the lineage record.
        
        Args:
            final_hash: Hash of final dataset
            sample_count: Number of samples in final dataset
            
        Returns:
            Complete DatasetLineage
        """
        self.lineage.final_hash = final_hash or self.lineage.compute_hash()
        self.lineage.sample_count = sample_count
        
        logger.info(
            f"Finalized lineage for {self.lineage.dataset_name} "
            f"v{self.lineage.dataset_version} with {sample_count} samples"
        )
        
        return self.lineage
