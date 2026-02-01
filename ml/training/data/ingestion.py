"""
Production data ingestion pipeline for VEID ML training.

This module provides production-ready data ingestion with:
- Multi-source ingestion (local, S3, GCS, API)
- PII-safe processing with redaction controls
- Dataset versioning and lineage tracking
- Deterministic processing for reproducibility

Supports both synthetic (for CI) and real data paths.
"""

import hashlib
import json
import logging
import os
import secrets
import tempfile
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from pathlib import Path
from typing import Any, Callable, Dict, Iterator, List, Optional, Tuple, Union

import numpy as np

from ml.training.config import DatasetConfig, AnonymizationMethod
from ml.training.dataset.ingestion import (
    CaptureMetadata,
    Dataset,
    DatasetSplit,
    DocumentInfo,
    IdentitySample,
    ImageData,
    SplitType,
)
from ml.training.dataset.anonymization import PIIAnonymizer, AnonymizationResult
from ml.training.dataset.lineage import (
    LineageTracker,
    SourceType,
    TransformType,
)

logger = logging.getLogger(__name__)


class DataSourceType(str, Enum):
    """Types of production data sources."""
    LOCAL_DIRECTORY = "local_directory"
    S3_BUCKET = "s3_bucket"
    GCS_BUCKET = "gcs_bucket"
    API_ENDPOINT = "api_endpoint"
    SYNTHETIC = "synthetic"
    DATABASE = "database"


class PIIRedactionLevel(str, Enum):
    """PII redaction levels for data processing."""
    NONE = "none"  # No redaction (testing only)
    MINIMAL = "minimal"  # Hash identifiers only
    STANDARD = "standard"  # Hash identifiers + redact names
    STRICT = "strict"  # Full redaction of all PII fields
    PRODUCTION = "production"  # Maximum protection


# PII fields by redaction level
PII_FIELDS_BY_LEVEL: Dict[PIIRedactionLevel, List[str]] = {
    PIIRedactionLevel.NONE: [],
    PIIRedactionLevel.MINIMAL: [
        "sample_id",
        "doc_id_hash",
    ],
    PIIRedactionLevel.STANDARD: [
        "sample_id",
        "doc_id_hash",
        "name",
        "full_name",
        "first_name",
        "last_name",
    ],
    PIIRedactionLevel.STRICT: [
        "sample_id",
        "doc_id_hash",
        "name",
        "full_name",
        "first_name",
        "last_name",
        "date_of_birth",
        "address",
        "postal_code",
        "document_number",
        "nationality",
    ],
    PIIRedactionLevel.PRODUCTION: [
        "sample_id",
        "doc_id_hash",
        "name",
        "full_name",
        "first_name",
        "last_name",
        "date_of_birth",
        "address",
        "postal_code",
        "document_number",
        "nationality",
        "face_embedding",
        "biometric_hash",
        "device_id",
        "session_id",
        "ip_address",
    ],
}


@dataclass
class DataSource:
    """Configuration for a data source."""
    
    source_id: str
    source_type: DataSourceType
    location: str  # URI or path
    
    # Credentials (never logged)
    credentials: Optional[Dict[str, str]] = None
    
    # Filters
    document_types: Optional[List[str]] = None
    date_range: Optional[Tuple[str, str]] = None
    max_samples: Optional[int] = None
    
    # Quality thresholds
    min_quality_score: float = 0.0
    
    # Processing
    preprocess: bool = True
    validate_schema: bool = True
    
    # Metadata
    metadata: Dict[str, Any] = field(default_factory=dict)


@dataclass
class IngestionConfig:
    """Configuration for production data ingestion."""
    
    # Sources
    sources: List[DataSource] = field(default_factory=list)
    
    # PII controls
    pii_redaction_level: PIIRedactionLevel = PIIRedactionLevel.PRODUCTION
    anonymization_salt: Optional[str] = None  # Auto-generated if not provided
    anonymization_method: str = AnonymizationMethod.HASH_SHA256.value
    
    # Access controls
    require_audit_log: bool = True
    audit_log_path: str = "logs/ingestion_audit.log"
    
    # Processing
    validate_samples: bool = True
    skip_invalid: bool = True
    max_parallel_sources: int = 4
    
    # Output
    output_dir: Optional[str] = None
    save_intermediate: bool = False
    
    # Determinism
    random_seed: int = 42
    
    # Versioning
    dataset_name: str = "veid_trust"
    schema_version: str = "1.0.0"
    
    def __post_init__(self):
        """Initialize derived values."""
        if not self.anonymization_salt:
            # Derive salt deterministically from random_seed for reproducibility
            # Using SHA256 hash of seed ensures consistent salt across runs
            seed_bytes = f"veid_pii_salt_{self.random_seed}".encode()
            self.anonymization_salt = hashlib.sha256(seed_bytes).hexdigest()


@dataclass
class IngestionResult:
    """Result of a data ingestion operation."""
    
    # Statistics
    total_samples: int = 0
    samples_loaded: int = 0
    samples_skipped: int = 0
    samples_anonymized: int = 0
    
    # Source breakdown
    samples_per_source: Dict[str, int] = field(default_factory=dict)
    
    # Timing
    started_at: float = field(default_factory=lambda: datetime.now().timestamp())
    completed_at: Optional[float] = None
    duration_seconds: float = 0.0
    
    # Lineage
    lineage_id: str = ""
    content_hash: str = ""
    
    # Status
    success: bool = True
    errors: List[str] = field(default_factory=list)
    warnings: List[str] = field(default_factory=list)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for logging."""
        return {
            "total_samples": self.total_samples,
            "samples_loaded": self.samples_loaded,
            "samples_skipped": self.samples_skipped,
            "samples_anonymized": self.samples_anonymized,
            "samples_per_source": self.samples_per_source,
            "duration_seconds": self.duration_seconds,
            "lineage_id": self.lineage_id,
            "content_hash": self.content_hash,
            "success": self.success,
            "errors": self.errors,
            "warnings": self.warnings,
        }


class ProductionIngestion:
    """
    Production-grade data ingestion pipeline.
    
    Features:
    - Multi-source data loading with validation
    - PII-safe processing with configurable redaction levels
    - Comprehensive lineage tracking
    - Deterministic processing for reproducibility
    - Audit logging for compliance
    
    Usage:
        config = IngestionConfig(
            sources=[
                DataSource("prod_s3", DataSourceType.S3_BUCKET, "s3://bucket/prefix"),
            ],
            pii_redaction_level=PIIRedactionLevel.PRODUCTION,
        )
        
        ingestion = ProductionIngestion(config)
        dataset, result = ingestion.run()
    """
    
    def __init__(self, config: IngestionConfig):
        """
        Initialize production ingestion pipeline.
        
        Args:
            config: Ingestion configuration
        """
        self.config = config
        self._rng = np.random.RandomState(config.random_seed)
        
        # Initialize PII anonymizer
        pii_fields = PII_FIELDS_BY_LEVEL.get(
            config.pii_redaction_level,
            PII_FIELDS_BY_LEVEL[PIIRedactionLevel.PRODUCTION]
        )
        
        from ml.training.dataset.anonymization import PIICategory
        
        # Group fields by category
        pii_fields_dict = {
            PIICategory.DIRECT_IDENTIFIER: [
                f for f in pii_fields if f in [
                    "sample_id", "doc_id_hash", "name", "full_name",
                    "first_name", "last_name", "document_number"
                ]
            ],
            PIICategory.QUASI_IDENTIFIER: [
                f for f in pii_fields if f in [
                    "date_of_birth", "address", "postal_code", "nationality"
                ]
            ],
            PIICategory.SENSITIVE: [
                f for f in pii_fields if f in [
                    "face_embedding", "biometric_hash"
                ]
            ],
            PIICategory.DEVICE: [
                f for f in pii_fields if f in [
                    "device_id", "session_id", "ip_address"
                ]
            ],
        }
        
        self._anonymizer = PIIAnonymizer(
            method=config.anonymization_method,
            salt=config.anonymization_salt,
            pii_fields=pii_fields_dict,
        )
        
        # Initialize lineage tracker
        self._lineage_tracker = LineageTracker(
            dataset_name=config.dataset_name,
            dataset_version=datetime.now().strftime("%Y%m%d-%H%M%S"),
            schema_version=config.schema_version,
        )
        
        # Initialize audit log
        if config.require_audit_log:
            Path(config.audit_log_path).parent.mkdir(parents=True, exist_ok=True)
    
    def run(self) -> Tuple[Dataset, IngestionResult]:
        """
        Execute the ingestion pipeline.
        
        Returns:
            Tuple of (Dataset, IngestionResult)
        """
        result = IngestionResult()
        all_samples: List[IdentitySample] = []
        
        logger.info(f"Starting production ingestion from {len(self.config.sources)} sources")
        
        # Process each source
        for source in self.config.sources:
            try:
                source_samples = self._ingest_source(source)
                all_samples.extend(source_samples)
                
                result.samples_per_source[source.source_id] = len(source_samples)
                result.samples_loaded += len(source_samples)
                
                # Track lineage
                source_type = self._map_source_type(source.source_type)
                self._lineage_tracker.add_source(
                    location=source.location,
                    source_type=source_type,
                    record_count=len(source_samples),
                )
                
                logger.info(f"Loaded {len(source_samples)} samples from {source.source_id}")
                
            except Exception as e:
                error_msg = f"Error ingesting from {source.source_id}: {e}"
                result.errors.append(error_msg)
                logger.error(error_msg)
                
                if not self.config.skip_invalid:
                    raise
        
        if len(all_samples) == 0:
            result.success = False
            result.errors.append("No samples loaded from any source")
            return self._empty_dataset(), result
        
        result.total_samples = len(all_samples)
        
        # Apply PII redaction
        if self.config.pii_redaction_level != PIIRedactionLevel.NONE:
            all_samples = self._apply_pii_redaction(all_samples)
            result.samples_anonymized = len(all_samples)
            
            self._lineage_tracker.add_transform(
                transform_type=TransformType.ANONYMIZATION,
                description=f"PII redaction ({self.config.pii_redaction_level.value})",
                input_count=len(all_samples),
                output_count=len(all_samples),
            )
        
        # Validate samples
        if self.config.validate_samples:
            all_samples, skipped = self._validate_samples(all_samples)
            result.samples_skipped = skipped
        
        # Create dataset
        dataset = self._create_dataset(all_samples)
        
        # Finalize lineage
        content_hash = self._compute_content_hash(dataset)
        lineage = self._lineage_tracker.finalize(
            final_hash=content_hash,
            sample_count=len(all_samples),
        )
        
        result.lineage_id = lineage.lineage_id
        result.content_hash = content_hash
        result.completed_at = datetime.now().timestamp()
        result.duration_seconds = result.completed_at - result.started_at
        
        # Write audit log
        if self.config.require_audit_log:
            self._write_audit_log(result)
        
        logger.info(
            f"Ingestion complete: {result.samples_loaded} samples, "
            f"{result.samples_skipped} skipped, "
            f"hash={content_hash}"
        )
        
        return dataset, result
    
    def _ingest_source(self, source: DataSource) -> List[IdentitySample]:
        """Ingest samples from a single source."""
        if source.source_type == DataSourceType.SYNTHETIC:
            return self._load_synthetic_data(source)
        elif source.source_type == DataSourceType.LOCAL_DIRECTORY:
            return self._load_local_data(source)
        elif source.source_type == DataSourceType.S3_BUCKET:
            return self._load_s3_data(source)
        elif source.source_type == DataSourceType.GCS_BUCKET:
            return self._load_gcs_data(source)
        elif source.source_type == DataSourceType.API_ENDPOINT:
            return self._load_api_data(source)
        else:
            raise ValueError(f"Unsupported source type: {source.source_type}")
    
    def _load_synthetic_data(self, source: DataSource) -> List[IdentitySample]:
        """Load synthetic data for testing."""
        from ml.training.dataset.synthetic import (
            SyntheticDataGenerator,
            SyntheticConfig,
        )
        
        num_samples = source.max_samples or 100
        
        config = SyntheticConfig(
            num_samples=num_samples,
            random_seed=self.config.random_seed,
            generate_images=False,
        )
        
        generator = SyntheticDataGenerator(config)
        dataset = generator.generate_dataset()
        
        return (
            list(dataset.train) +
            list(dataset.validation) +
            list(dataset.test)
        )
    
    def _load_local_data(self, source: DataSource) -> List[IdentitySample]:
        """Load data from local directory."""
        from ml.training.dataset.ingestion import DatasetIngestion
        
        dataset_config = DatasetConfig(
            data_paths=[source.location],
            doc_types=source.document_types or [
                "id_card", "passport", "drivers_license"
            ],
            max_samples=source.max_samples,
            min_doc_quality=source.min_quality_score,
            random_seed=self.config.random_seed,
            anonymize=False,  # We handle anonymization separately
        )
        
        ingestion = DatasetIngestion(dataset_config)
        dataset = ingestion.load_dataset()
        
        return (
            list(dataset.train) +
            list(dataset.validation) +
            list(dataset.test)
        )
    
    def _load_s3_data(self, source: DataSource) -> List[IdentitySample]:
        """Load data from S3 bucket."""
        from ml.training.dataset.connectors import (
            ConnectorRegistry,
            ConnectorConfig,
            ConnectorType,
        )
        
        config = ConnectorConfig(
            connector_type=ConnectorType.S3,
            endpoint=source.location,
            credentials=source.credentials,
        )
        
        connector = ConnectorRegistry.get_connector(config)
        result = connector.connect()
        
        if not result.success:
            raise RuntimeError(f"Failed to connect to S3: {result.message}")
        
        samples = []
        for record_id in connector.list_records():
            record = connector.get_record(record_id)
            if record:
                sample = self._record_to_sample(record)
                samples.append(sample)
                
                if source.max_samples and len(samples) >= source.max_samples:
                    break
        
        return samples
    
    def _load_gcs_data(self, source: DataSource) -> List[IdentitySample]:
        """Load data from GCS bucket."""
        from ml.training.dataset.connectors import (
            ConnectorRegistry,
            ConnectorConfig,
            ConnectorType,
        )
        
        config = ConnectorConfig(
            connector_type=ConnectorType.GCS,
            endpoint=source.location,
            credentials=source.credentials,
        )
        
        connector = ConnectorRegistry.get_connector(config)
        result = connector.connect()
        
        if not result.success:
            raise RuntimeError(f"Failed to connect to GCS: {result.message}")
        
        samples = []
        for record_id in connector.list_records():
            record = connector.get_record(record_id)
            if record:
                sample = self._record_to_sample(record)
                samples.append(sample)
                
                if source.max_samples and len(samples) >= source.max_samples:
                    break
        
        return samples
    
    def _load_api_data(self, source: DataSource) -> List[IdentitySample]:
        """Load data from API endpoint."""
        from ml.training.dataset.connectors import (
            ConnectorRegistry,
            ConnectorConfig,
            ConnectorType,
        )
        
        config = ConnectorConfig(
            connector_type=ConnectorType.HTTP_API,
            endpoint=source.location,
            credentials=source.credentials,
        )
        
        connector = ConnectorRegistry.get_connector(config)
        result = connector.connect()
        
        if not result.success:
            raise RuntimeError(f"Failed to connect to API: {result.message}")
        
        samples = []
        for record_id in connector.list_records():
            record = connector.get_record(record_id)
            if record:
                sample = self._record_to_sample(record)
                samples.append(sample)
                
                if source.max_samples and len(samples) >= source.max_samples:
                    break
        
        return samples
    
    def _record_to_sample(self, record: Any) -> IdentitySample:
        """Convert a connector record to an IdentitySample."""
        data = record.data if hasattr(record, 'data') else record
        
        return IdentitySample(
            sample_id=data.get("sample_id", ""),
            document_info=DocumentInfo(
                doc_type=data.get("document_type", "id_card"),
                doc_id_hash=data.get("doc_id_hash", ""),
                country_code=data.get("country_code"),
                expiry_valid=data.get("expiry_valid"),
                mrz_present=data.get("mrz_present"),
            ),
            trust_score=float(data.get("trust_score", 0)),
            is_genuine=data.get("is_genuine", True),
            fraud_type=data.get("fraud_type"),
            face_detected=data.get("face_detected", True),
            face_confidence=float(data.get("face_confidence", 0)),
            document_quality_score=float(data.get("document_quality_score", 0)),
            ocr_success=data.get("ocr_success", True),
            ocr_confidence=float(data.get("ocr_confidence", 0)),
        )
    
    def _apply_pii_redaction(
        self,
        samples: List[IdentitySample]
    ) -> List[IdentitySample]:
        """Apply PII redaction to all samples."""
        # Create temporary dataset for anonymization
        temp_dataset = Dataset(
            train=DatasetSplit(SplitType.TRAIN, samples),
            validation=DatasetSplit(SplitType.VALIDATION, []),
            test=DatasetSplit(SplitType.TEST, []),
        )
        
        self._anonymizer.anonymize_dataset(temp_dataset)
        
        return list(temp_dataset.train)
    
    def _validate_samples(
        self,
        samples: List[IdentitySample]
    ) -> Tuple[List[IdentitySample], int]:
        """Validate samples and return valid ones."""
        from ml.training.dataset.validation import SchemaValidator
        
        validator = SchemaValidator()
        valid_samples = []
        skipped = 0
        
        for sample in samples:
            issues = validator.validate_sample(sample)
            errors = [i for i in issues if i.level.value == "error"]
            
            if len(errors) == 0:
                valid_samples.append(sample)
            else:
                skipped += 1
                logger.debug(f"Skipped invalid sample {sample.sample_id}: {errors}")
        
        return valid_samples, skipped
    
    def _create_dataset(self, samples: List[IdentitySample]) -> Dataset:
        """Create dataset with proper splits."""
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
        
        self._lineage_tracker.add_transform(
            transform_type=TransformType.SPLITTING,
            description=f"Stratified split (seed={self.config.random_seed})",
            config={"train": 0.7, "val": 0.15, "test": 0.15},
            input_count=len(samples),
            output_count=len(samples),
            output_hash=result.combined_hash,
        )
        
        return result.dataset
    
    def _empty_dataset(self) -> Dataset:
        """Create an empty dataset."""
        return Dataset(
            train=DatasetSplit(SplitType.TRAIN, []),
            validation=DatasetSplit(SplitType.VALIDATION, []),
            test=DatasetSplit(SplitType.TEST, []),
        )
    
    def _compute_content_hash(self, dataset: Dataset) -> str:
        """Compute content hash for the dataset."""
        hash_data = []
        
        for split in [dataset.train, dataset.validation, dataset.test]:
            for sample in split.samples:
                hash_data.append(sample.sample_id)
                hash_data.append(str(sample.trust_score))
        
        combined = "|".join(hash_data).encode()
        return hashlib.sha256(combined).hexdigest()[:16]
    
    def _map_source_type(self, source_type: DataSourceType) -> SourceType:
        """Map DataSourceType to lineage SourceType."""
        mapping = {
            DataSourceType.LOCAL_DIRECTORY: SourceType.LOCAL_FILE,
            DataSourceType.S3_BUCKET: SourceType.OBJECT_STORE,
            DataSourceType.GCS_BUCKET: SourceType.OBJECT_STORE,
            DataSourceType.API_ENDPOINT: SourceType.API,
            DataSourceType.SYNTHETIC: SourceType.SYNTHETIC,
            DataSourceType.DATABASE: SourceType.DATABASE,
        }
        return mapping.get(source_type, SourceType.LOCAL_FILE)
    
    def _write_audit_log(self, result: IngestionResult) -> None:
        """Write ingestion audit log entry."""
        entry = {
            "timestamp": datetime.now().isoformat(),
            "operation": "data_ingestion",
            "lineage_id": result.lineage_id,
            "content_hash": result.content_hash,
            "samples_loaded": result.samples_loaded,
            "samples_skipped": result.samples_skipped,
            "pii_redaction_level": self.config.pii_redaction_level.value,
            "success": result.success,
            "duration_seconds": result.duration_seconds,
        }
        
        with open(self.config.audit_log_path, "a") as f:
            f.write(json.dumps(entry) + "\n")


def load_production_dataset(
    sources: List[Union[str, DataSource]],
    pii_redaction_level: PIIRedactionLevel = PIIRedactionLevel.PRODUCTION,
    seed: int = 42,
) -> Tuple[Dataset, IngestionResult]:
    """
    Convenience function to load a production dataset.
    
    Args:
        sources: List of source URIs or DataSource objects
        pii_redaction_level: Level of PII redaction to apply
        seed: Random seed for reproducibility
        
    Returns:
        Tuple of (Dataset, IngestionResult)
    """
    source_list = []
    
    for i, source in enumerate(sources):
        if isinstance(source, str):
            # Determine source type from URI
            if source.startswith("s3://"):
                source_type = DataSourceType.S3_BUCKET
            elif source.startswith("gs://"):
                source_type = DataSourceType.GCS_BUCKET
            elif source.startswith("http://") or source.startswith("https://"):
                source_type = DataSourceType.API_ENDPOINT
            else:
                source_type = DataSourceType.LOCAL_DIRECTORY
            
            source_list.append(DataSource(
                source_id=f"source_{i}",
                source_type=source_type,
                location=source,
            ))
        else:
            source_list.append(source)
    
    config = IngestionConfig(
        sources=source_list,
        pii_redaction_level=pii_redaction_level,
        random_seed=seed,
    )
    
    ingestion = ProductionIngestion(config)
    return ingestion.run()
