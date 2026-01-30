"""
Dataset handling for trust score training.

This module provides:
- Dataset ingestion from multiple sources (local, S3, GCS, API)
- Data preprocessing and validation
- Data augmentation for training
- PII anonymization for security
- Synthetic data generation for CI testing
- Secure PII-safe storage with encryption
- Labeling pipeline (human review + heuristics)
- Signed manifest generation
- Deterministic train/val/test splits
- Dataset lineage tracking
- Schema and label validation
"""

from ml.training.dataset.ingestion import (
    DatasetIngestion,
    IdentitySample,
    Dataset,
    DatasetSplit,
)
from ml.training.dataset.preprocessing import (
    DatasetPreprocessor,
    PreprocessedSample,
)
from ml.training.dataset.augmentation import (
    DataAugmentation,
    AugmentedSample,
)
from ml.training.dataset.anonymization import (
    PIIAnonymizer,
    AnonymizationResult,
)
from ml.training.dataset.connectors import (
    ConnectorRegistry,
    ConnectorConfig,
    ConnectorType,
    DataConnector,
    LocalFileConnector,
    S3Connector,
    GCSConnector,
    HTTPAPIConnector,
)
from ml.training.dataset.synthetic import (
    SyntheticDataGenerator,
    SyntheticConfig,
    SyntheticProfile,
    generate_synthetic_dataset,
)
from ml.training.dataset.storage import (
    SecureStorage,
    StorageConfig,
    EncryptionManager,
    EncryptionEnvelope,
)
from ml.training.dataset.labeling import (
    LabelingPipeline,
    HeuristicLabeler,
    Label,
    LabelBatch,
    LabelSource,
    CSVLabelReader,
    CSVLabelWriter,
    LabelValidator,
    ConsensusLabeler,
)
from ml.training.dataset.manifest import (
    DatasetManifest,
    ManifestBuilder,
    ManifestSigner,
    ManifestVerifier,
    ContentHash,
    ManifestSignature,
    create_signed_manifest,
)
from ml.training.dataset.splits import (
    DeterministicSplitter,
    SplitConfig,
    SplitResult,
    SplitStrategy,
    SplitVerifier,
    split_dataset,
)
from ml.training.dataset.lineage import (
    LineageTracker,
    DatasetLineage,
    SourceRecord,
    TransformRecord,
    BuildInfo,
    SourceType,
    TransformType,
)
from ml.training.dataset.validation import (
    DatasetValidator,
    ValidationReport,
    ValidationIssue,
    SchemaValidator,
    LabelAnomalyDetector,
    DataQualityChecker,
    validate_dataset,
)

__all__ = [
    # Ingestion
    "DatasetIngestion",
    "IdentitySample",
    "Dataset",
    "DatasetSplit",
    # Preprocessing
    "DatasetPreprocessor",
    "PreprocessedSample",
    # Augmentation
    "DataAugmentation",
    "AugmentedSample",
    # Anonymization
    "PIIAnonymizer",
    "AnonymizationResult",
    # Connectors
    "ConnectorRegistry",
    "ConnectorConfig",
    "ConnectorType",
    "DataConnector",
    "LocalFileConnector",
    "S3Connector",
    "GCSConnector",
    "HTTPAPIConnector",
    # Synthetic
    "SyntheticDataGenerator",
    "SyntheticConfig",
    "SyntheticProfile",
    "generate_synthetic_dataset",
    # Storage
    "SecureStorage",
    "StorageConfig",
    "EncryptionManager",
    "EncryptionEnvelope",
    # Labeling
    "LabelingPipeline",
    "HeuristicLabeler",
    "Label",
    "LabelBatch",
    "LabelSource",
    "CSVLabelReader",
    "CSVLabelWriter",
    "LabelValidator",
    "ConsensusLabeler",
    # Manifest
    "DatasetManifest",
    "ManifestBuilder",
    "ManifestSigner",
    "ManifestVerifier",
    "ContentHash",
    "ManifestSignature",
    "create_signed_manifest",
    # Splits
    "DeterministicSplitter",
    "SplitConfig",
    "SplitResult",
    "SplitStrategy",
    "SplitVerifier",
    "split_dataset",
    # Lineage
    "LineageTracker",
    "DatasetLineage",
    "SourceRecord",
    "TransformRecord",
    "BuildInfo",
    "SourceType",
    "TransformType",
    # Validation
    "DatasetValidator",
    "ValidationReport",
    "ValidationIssue",
    "SchemaValidator",
    "LabelAnomalyDetector",
    "DataQualityChecker",
    "validate_dataset",
]
