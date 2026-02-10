"""
Production data pipeline for VEID ML training.

This module provides production-ready components for:
- Data ingestion with versioning
- PII-safe storage with redaction and access controls
- Dataset versioning and provenance tracking
- Reproducible train/validation/test splits
- Synthetic dataset path for CI testing

All components are designed for compliance with privacy requirements
and deterministic model training.
"""

from ml.training.data.ingestion import (
    ProductionIngestion,
    IngestionConfig,
    IngestionResult,
    DataSource,
    DataSourceType,
)

from ml.training.data.labeling import (
    ProductionLabelingPipeline,
    LabelingConfig,
    LabelingWorkflow,
    ReviewQueue,
    AutoLabeler,
)

from ml.training.data.versioning import (
    DatasetVersionManager,
    DatasetVersion,
    ProvenanceRecord,
    VersionConfig,
    create_versioned_dataset,
)

__all__ = [
    # Ingestion
    "ProductionIngestion",
    "IngestionConfig",
    "IngestionResult",
    "DataSource",
    "DataSourceType",
    # Labeling
    "ProductionLabelingPipeline",
    "LabelingConfig",
    "LabelingWorkflow",
    "ReviewQueue",
    "AutoLabeler",
    # Versioning
    "DatasetVersionManager",
    "DatasetVersion",
    "ProvenanceRecord",
    "VersionConfig",
    "create_versioned_dataset",
]
