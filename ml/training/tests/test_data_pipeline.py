"""
Tests for production data pipeline components.

Tests cover:
- ProductionIngestion with PII redaction
- ProductionLabelingPipeline with workflows
- DatasetVersionManager with provenance
- CI synthetic dataset loading
"""

import json
import os
import tempfile
from pathlib import Path

import numpy as np
import pytest

from ml.training.dataset.ingestion import (
    Dataset,
    DatasetSplit,
    DocumentInfo,
    IdentitySample,
    SplitType,
)


class TestProductionIngestion:
    """Tests for production data ingestion."""
    
    def test_ingestion_config(self):
        """Test ingestion configuration."""
        from ml.training.data.ingestion import (
            DataSource,
            DataSourceType,
            IngestionConfig,
            PIIRedactionLevel,
        )
        
        config = IngestionConfig(
            sources=[
                DataSource(
                    source_id="test",
                    source_type=DataSourceType.SYNTHETIC,
                    location="synthetic://test",
                    max_samples=10,
                ),
            ],
            pii_redaction_level=PIIRedactionLevel.STANDARD,
            random_seed=42,
        )
        
        assert len(config.sources) == 1
        assert config.pii_redaction_level == PIIRedactionLevel.STANDARD
        assert config.anonymization_salt is not None
    
    def test_synthetic_ingestion(self):
        """Test ingestion with synthetic data source."""
        from ml.training.data.ingestion import (
            DataSource,
            DataSourceType,
            IngestionConfig,
            PIIRedactionLevel,
            ProductionIngestion,
        )
        
        config = IngestionConfig(
            sources=[
                DataSource(
                    source_id="synthetic_ci",
                    source_type=DataSourceType.SYNTHETIC,
                    location="synthetic://ci",
                    max_samples=30,
                ),
            ],
            pii_redaction_level=PIIRedactionLevel.MINIMAL,
            random_seed=42,
            require_audit_log=False,
        )
        
        ingestion = ProductionIngestion(config)
        dataset, result = ingestion.run()
        
        assert result.success
        assert result.samples_loaded == 30
        assert len(dataset) == 30
        assert result.content_hash != ""
    
    def test_pii_redaction_levels(self):
        """Test different PII redaction levels."""
        from ml.training.data.ingestion import (
            PIIRedactionLevel,
            PII_FIELDS_BY_LEVEL,
        )
        
        # Check increasing strictness
        none_fields = PII_FIELDS_BY_LEVEL[PIIRedactionLevel.NONE]
        minimal_fields = PII_FIELDS_BY_LEVEL[PIIRedactionLevel.MINIMAL]
        standard_fields = PII_FIELDS_BY_LEVEL[PIIRedactionLevel.STANDARD]
        strict_fields = PII_FIELDS_BY_LEVEL[PIIRedactionLevel.STRICT]
        production_fields = PII_FIELDS_BY_LEVEL[PIIRedactionLevel.PRODUCTION]
        
        assert len(none_fields) == 0
        assert len(minimal_fields) < len(standard_fields)
        assert len(standard_fields) < len(strict_fields)
        assert len(strict_fields) < len(production_fields)
        
        # Production should include all sensitive fields
        assert "sample_id" in production_fields
        assert "doc_id_hash" in production_fields
        assert "face_embedding" in production_fields


class TestProductionLabeling:
    """Tests for production labeling pipeline."""
    
    def test_labeling_config(self):
        """Test labeling configuration."""
        from ml.training.data.labeling import (
            LabelingConfig,
            LabelingWorkflow,
        )
        
        config = LabelingConfig(
            workflow=LabelingWorkflow.HYBRID,
            auto_label_confidence_threshold=0.85,
            random_seed=42,
        )
        
        assert config.workflow == LabelingWorkflow.HYBRID
        assert config.auto_label_confidence_threshold == 0.85
    
    def test_synthetic_workflow(self):
        """Test synthetic data labeling workflow."""
        from ml.training.data.labeling import (
            LabelingConfig,
            LabelingWorkflow,
            ProductionLabelingPipeline,
        )
        
        # Create test samples
        samples = [
            IdentitySample(
                sample_id=f"test_{i:03d}",
                document_info=DocumentInfo(
                    doc_type="id_card",
                    doc_id_hash=f"hash_{i}",
                ),
                face_confidence=0.9,
                document_quality_score=0.85,
                ocr_confidence=0.88,
                ocr_success=True,
                face_detected=True,
            )
            for i in range(30)
        ]
        
        config = LabelingConfig(
            workflow=LabelingWorkflow.SYNTHETIC,
            random_seed=42,
        )
        
        pipeline = ProductionLabelingPipeline(config)
        dataset, result = pipeline.label_dataset(samples)
        
        assert result.total_samples == 30
        assert result.auto_labeled == 30
        assert result.sent_for_review == 0
    
    def test_hybrid_workflow(self):
        """Test hybrid labeling workflow."""
        from ml.training.data.labeling import (
            LabelingConfig,
            LabelingWorkflow,
            ProductionLabelingPipeline,
        )
        
        # Create samples with varying confidence
        samples = []
        for i in range(30):
            confidence = 0.5 + (i % 10) / 20.0  # Range 0.5-0.95
            samples.append(IdentitySample(
                sample_id=f"hybrid_{i:03d}",
                document_info=DocumentInfo(
                    doc_type="passport",
                    doc_id_hash=f"hash_{i}",
                ),
                face_confidence=confidence,
                document_quality_score=confidence,
                ocr_confidence=confidence,
                ocr_success=True,
                face_detected=True,
            ))
        
        config = LabelingConfig(
            workflow=LabelingWorkflow.HYBRID,
            auto_label_confidence_threshold=0.85,
            random_seed=42,
        )
        
        pipeline = ProductionLabelingPipeline(config)
        dataset, result = pipeline.label_dataset(samples)
        
        assert result.total_samples == 30
        assert result.auto_labeled + result.sent_for_review == 30
    
    def test_review_queue(self):
        """Test review queue management."""
        from ml.training.data.labeling import (
            ReviewQueue,
            ReviewPriority,
        )
        from ml.training.dataset.labeling import Label
        
        queue = ReviewQueue("test_queue")
        
        # Add samples
        for i in range(5):
            sample = IdentitySample(
                sample_id=f"review_{i}",
                face_confidence=0.7,
            )
            priority = ReviewPriority.HIGH if i % 2 == 0 else ReviewPriority.NORMAL
            queue.add(sample, priority=priority, reason=f"Test {i}")
        
        # Check queue
        pending = queue.list_pending()
        assert len(pending) == 5
        
        # High priority should come first
        high_priority = queue.list_pending(priority=ReviewPriority.HIGH)
        assert len(high_priority) == 3
        
        # Submit review
        label = Label(sample_id="review_0", trust_score=75.0, is_genuine=True)
        assert queue.submit_review("review_0", label)
        
        # Check completion
        completed = queue.get_completed()
        assert len(completed) == 1


class TestDatasetVersioning:
    """Tests for dataset versioning and provenance."""
    
    def test_version_config(self):
        """Test version configuration."""
        from ml.training.data.versioning import VersionConfig
        
        config = VersionConfig(
            base_path="data/test_versions",
            sign_manifests=True,
        )
        
        assert config.base_path == "data/test_versions"
        assert config.sign_manifests is True
    
    def test_create_version(self, tmp_path):
        """Test creating a versioned dataset."""
        from ml.training.data.versioning import (
            DatasetVersionManager,
            VersionConfig,
            VersionStatus,
        )
        
        # Create test dataset
        samples = [
            IdentitySample(
                sample_id=f"ver_{i:03d}",
                trust_score=75.0 + i,
                is_genuine=True,
            )
            for i in range(30)
        ]
        
        dataset = Dataset(
            train=DatasetSplit(SplitType.TRAIN, samples[:20]),
            validation=DatasetSplit(SplitType.VALIDATION, samples[20:25]),
            test=DatasetSplit(SplitType.TEST, samples[25:30]),
        )
        
        config = VersionConfig(
            base_path=str(tmp_path / "versions"),
            sign_manifests=False,
        )
        
        manager = DatasetVersionManager(config)
        version = manager.create_version(
            dataset=dataset,
            version="v1.0.0",
            description="Test version",
            tags=["test", "ci"],
        )
        
        assert version.version == "v1.0.0"
        assert version.provenance.status == VersionStatus.DRAFT
        assert len(version.dataset) == 30
    
    def test_version_save_load(self, tmp_path):
        """Test saving and loading versions."""
        from ml.training.data.versioning import (
            DatasetVersion,
            DatasetVersionManager,
            VersionConfig,
        )
        
        # Create test dataset
        samples = [
            IdentitySample(
                sample_id=f"save_{i:03d}",
                trust_score=80.0,
                is_genuine=True,
            )
            for i in range(20)
        ]
        
        dataset = Dataset(
            train=DatasetSplit(SplitType.TRAIN, samples[:14]),
            validation=DatasetSplit(SplitType.VALIDATION, samples[14:17]),
            test=DatasetSplit(SplitType.TEST, samples[17:20]),
        )
        
        config = VersionConfig(
            base_path=str(tmp_path / "versions"),
            sign_manifests=False,
        )
        
        manager = DatasetVersionManager(config)
        
        # Create and save
        version = manager.create_version(
            dataset=dataset,
            version="v1.0.0",
        )
        
        # Load back
        loaded = manager.load_version("v1.0.0")
        
        assert loaded.version == "v1.0.0"
        assert len(loaded.dataset) == 20
        assert loaded.content_hash == version.content_hash
    
    def test_version_verification(self, tmp_path):
        """Test version integrity verification."""
        from ml.training.data.versioning import (
            DatasetVersionManager,
            VersionConfig,
        )
        
        samples = [
            IdentitySample(
                sample_id=f"verify_{i:03d}",
                trust_score=75.0,
                is_genuine=True,
            )
            for i in range(15)
        ]
        
        dataset = Dataset(
            train=DatasetSplit(SplitType.TRAIN, samples[:10]),
            validation=DatasetSplit(SplitType.VALIDATION, samples[10:12]),
            test=DatasetSplit(SplitType.TEST, samples[12:15]),
        )
        
        config = VersionConfig(
            base_path=str(tmp_path / "versions"),
            sign_manifests=False,
        )
        
        manager = DatasetVersionManager(config)
        manager.create_version(dataset=dataset, version="v1.0.0")
        
        # Verify
        report = manager.verify_version("v1.0.0")
        
        assert report["valid"]
        assert report["checks"]["content_hash"]
    
    def test_provenance_record(self):
        """Test provenance record creation."""
        from ml.training.data.versioning import (
            ProvenanceRecord,
            VersionStatus,
        )
        from ml.training.dataset.lineage import BuildInfo
        
        build_info = BuildInfo.capture()
        
        provenance = ProvenanceRecord(
            version="v1.0.0",
            build_info=build_info,
            content_hash="abc123",
            description="Test provenance",
            tags=["test"],
        )
        
        assert provenance.version == "v1.0.0"
        assert provenance.status == VersionStatus.DRAFT
        
        # Test serialization
        data = provenance.to_dict()
        assert "version" in data
        assert "build_info" in data
        
        # Test deserialization
        loaded = ProvenanceRecord.from_dict(data)
        assert loaded.version == "v1.0.0"


class TestCISyntheticDataset:
    """Tests for CI synthetic dataset."""
    
    def test_load_ci_dataset(self):
        """Test loading the CI synthetic dataset."""
        ci_data_path = Path(__file__).parent.parent / "ci_data" / "synthetic_dataset.json"
        
        if not ci_data_path.exists():
            pytest.skip("CI dataset not found")
        
        with open(ci_data_path) as f:
            data = json.load(f)
        
        # Verify metadata
        assert data["metadata"]["name"] == "veid_synthetic_ci"
        assert data["metadata"]["pii_safe"] is True
        
        # Verify structure
        assert "samples" in data
        assert len(data["samples"]) == 50
        
        # Verify splits
        train_count = sum(1 for s in data["samples"] if s["split"] == "train")
        val_count = sum(1 for s in data["samples"] if s["split"] == "validation")
        test_count = sum(1 for s in data["samples"] if s["split"] == "test")
        
        assert train_count == 35
        assert val_count == 8
        assert test_count == 7
    
    def test_ci_dataset_no_pii(self):
        """Verify CI dataset contains no PII."""
        ci_data_path = Path(__file__).parent.parent / "ci_data" / "synthetic_dataset.json"
        
        if not ci_data_path.exists():
            pytest.skip("CI dataset not found")
        
        with open(ci_data_path) as f:
            data = json.load(f)
        
        # Check each sample
        pii_fields = ["name", "address", "date_of_birth", "document_number"]
        
        for sample in data["samples"]:
            for field in pii_fields:
                assert field not in sample, f"PII field {field} found in sample"
            
            # Sample IDs should be synthetic
            assert sample["sample_id"].startswith("synthetic_")
    
    def test_ci_dataset_to_samples(self):
        """Test converting CI dataset to IdentitySample objects."""
        ci_data_path = Path(__file__).parent.parent / "ci_data" / "synthetic_dataset.json"
        
        if not ci_data_path.exists():
            pytest.skip("CI dataset not found")
        
        with open(ci_data_path) as f:
            data = json.load(f)
        
        samples = []
        for s in data["samples"]:
            sample = IdentitySample(
                sample_id=s["sample_id"],
                document_info=DocumentInfo(
                    doc_type=s["document_type"],
                    doc_id_hash="",
                    country_code=s.get("country_code"),
                ),
                trust_score=s["trust_score"],
                is_genuine=s["is_genuine"],
                fraud_type=s.get("fraud_type"),
                face_detected=s["face_detected"],
                face_confidence=s["face_confidence"],
                document_quality_score=s["document_quality_score"],
                ocr_success=s["ocr_success"],
                ocr_confidence=s["ocr_confidence"],
            )
            samples.append(sample)
        
        assert len(samples) == 50
        
        # Check genuine/fraud ratio
        genuine = sum(1 for s in samples if s.is_genuine)
        fraud = sum(1 for s in samples if not s.is_genuine)
        
        # CI dataset has ~10% fraud ratio (5/50), but exact count depends on generation
        assert genuine >= 44 and genuine <= 48  # Allow small variance
        assert fraud >= 2 and fraud <= 6  # Allow small variance
        assert genuine + fraud == 50


class TestReproducibility:
    """Tests for reproducibility guarantees."""
    
    def test_deterministic_ingestion(self):
        """Test that ingestion is deterministic with same seed."""
        from ml.training.data.ingestion import (
            DataSource,
            DataSourceType,
            IngestionConfig,
            PIIRedactionLevel,
            ProductionIngestion,
        )
        
        def run_ingestion(seed: int):
            config = IngestionConfig(
                sources=[
                    DataSource(
                        source_id="synthetic",
                        source_type=DataSourceType.SYNTHETIC,
                        location="synthetic://test",
                        max_samples=20,
                    ),
                ],
                pii_redaction_level=PIIRedactionLevel.MINIMAL,
                random_seed=seed,
                require_audit_log=False,
            )
            ingestion = ProductionIngestion(config)
            dataset, result = ingestion.run()
            return result.content_hash
        
        hash1 = run_ingestion(42)
        hash2 = run_ingestion(42)
        hash3 = run_ingestion(123)
        
        assert hash1 == hash2  # Same seed = same result
        assert hash1 != hash3  # Different seed = different result
    
    def test_deterministic_splits(self):
        """Test that splits are deterministic."""
        from ml.training.dataset.splits import (
            DeterministicSplitter,
            SplitConfig,
        )
        
        samples = [
            IdentitySample(sample_id=f"split_{i:03d}", trust_score=75.0)
            for i in range(50)
        ]
        
        config = SplitConfig(random_seed=42)
        
        splitter1 = DeterministicSplitter(config)
        splitter2 = DeterministicSplitter(config)
        
        result1 = splitter1.split(samples)
        result2 = splitter2.split(samples)
        
        assert result1.combined_hash == result2.combined_hash
        assert result1.split_hashes == result2.split_hashes
