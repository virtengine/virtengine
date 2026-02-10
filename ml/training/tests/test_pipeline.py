"""
Tests for dataset connectors and pipeline components.
"""

import json
import os
import tempfile
from pathlib import Path

import numpy as np
import pytest

from ml.training.config import DatasetConfig
from ml.training.dataset.ingestion import (
    Dataset,
    DatasetSplit,
    DocumentInfo,
    IdentitySample,
    SplitType,
)


class TestConnectors:
    """Tests for data connectors."""
    
    def test_local_file_connector(self):
        """Test local file connector."""
        from ml.training.dataset.connectors import (
            ConnectorConfig,
            ConnectorType,
            LocalFileConnector,
        )
        
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create test data
            sample_dir = Path(tmpdir) / "sample_001"
            sample_dir.mkdir()
            
            metadata = {
                "sample_id": "sample_001",
                "document_type": "id_card",
                "trust_score": 75.0,
            }
            
            with open(sample_dir / "metadata.json", "w") as f:
                json.dump(metadata, f)
            
            # Test connector
            config = ConnectorConfig(
                connector_type=ConnectorType.LOCAL,
                endpoint=tmpdir,
            )
            
            connector = LocalFileConnector(config)
            result = connector.connect()
            
            assert result.success
            assert result.record_count == 1
            
            records = connector.list_records()
            assert len(records) == 1
            assert "sample_001" in records
            
            record = connector.get_record("sample_001")
            assert record is not None
            assert record.data["trust_score"] == 75.0
    
    def test_connector_registry(self):
        """Test connector registry."""
        from ml.training.dataset.connectors import (
            ConnectorRegistry,
            ConnectorType,
        )
        
        # List connectors
        connectors = ConnectorRegistry.list_connectors()
        assert ConnectorType.LOCAL in connectors
        assert ConnectorType.S3 in connectors
        assert ConnectorType.GCS in connectors
        assert ConnectorType.HTTP_API in connectors
    
    def test_connector_from_uri(self):
        """Test creating connector from URI."""
        from ml.training.dataset.connectors import (
            ConnectorRegistry,
            ConnectorType,
        )
        
        with tempfile.TemporaryDirectory() as tmpdir:
            connector = ConnectorRegistry.from_uri(tmpdir)
            assert connector.connector_type == ConnectorType.LOCAL
        
        # S3 URI
        connector = ConnectorRegistry.from_uri("s3://my-bucket/prefix")
        assert connector.connector_type == ConnectorType.S3
        
        # GCS URI
        connector = ConnectorRegistry.from_uri("gs://my-bucket/prefix")
        assert connector.connector_type == ConnectorType.GCS
        
        # HTTP URI
        connector = ConnectorRegistry.from_uri("https://api.example.com/v1")
        assert connector.connector_type == ConnectorType.HTTP_API


class TestSyntheticGenerator:
    """Tests for synthetic data generator."""
    
    def test_generate_dataset(self):
        """Test generating synthetic dataset."""
        from ml.training.dataset.synthetic import (
            SyntheticConfig,
            SyntheticDataGenerator,
        )
        
        config = SyntheticConfig(
            num_samples=50,
            random_seed=42,
            generate_images=False,
        )
        
        generator = SyntheticDataGenerator(config)
        dataset = generator.generate_dataset()
        
        assert len(dataset) == 50
        assert len(dataset.train) > 0
        assert len(dataset.validation) > 0
        assert len(dataset.test) > 0
        assert dataset.dataset_hash is not None
    
    def test_deterministic_generation(self):
        """Test that generation is deterministic."""
        from ml.training.dataset.synthetic import (
            SyntheticConfig,
            SyntheticDataGenerator,
        )
        
        config1 = SyntheticConfig(num_samples=30, random_seed=42, generate_images=False)
        config2 = SyntheticConfig(num_samples=30, random_seed=42, generate_images=False)
        
        gen1 = SyntheticDataGenerator(config1)
        gen2 = SyntheticDataGenerator(config2)
        
        dataset1 = gen1.generate_dataset()
        dataset2 = gen2.generate_dataset()
        
        assert dataset1.dataset_hash == dataset2.dataset_hash
    
    def test_synthetic_profile(self):
        """Test using synthetic profiles."""
        from ml.training.dataset.synthetic import (
            SyntheticConfig,
            SyntheticProfile,
        )
        
        config = SyntheticConfig.from_profile(SyntheticProfile.CI_MINIMAL)
        assert config.num_samples == 30
        assert config.generate_images is False


class TestLabelingPipeline:
    """Tests for labeling pipeline."""
    
    def test_heuristic_labeler(self):
        """Test heuristic labeling."""
        from ml.training.dataset.labeling import HeuristicLabeler, LabelSource
        
        labeler = HeuristicLabeler()
        
        sample = IdentitySample(
            sample_id="test_001",
            face_confidence=0.95,
            document_quality_score=0.85,
            ocr_confidence=0.90,
            ocr_success=True,
            face_detected=True,
        )
        
        label = labeler.label_sample(sample)
        
        assert label.sample_id == "test_001"
        assert label.source == LabelSource.HEURISTIC
        assert 0 <= label.trust_score <= 100
    
    def test_csv_reader_writer(self, tmp_path):
        """Test CSV label import/export."""
        from ml.training.dataset.labeling import (
            CSVLabelReader,
            CSVLabelWriter,
            Label,
            LabelBatch,
            LabelSource,
            LabelStatus,
        )
        
        # Create labels
        labels = LabelBatch(
            labels=[
                Label(
                    sample_id="sample_001",
                    trust_score=75.0,
                    is_genuine=True,
                    source=LabelSource.HUMAN,
                    status=LabelStatus.REVIEWED,
                ),
                Label(
                    sample_id="sample_002",
                    trust_score=25.0,
                    is_genuine=False,
                    fraud_type="printed_photo",
                    source=LabelSource.HUMAN,
                    status=LabelStatus.REVIEWED,
                ),
            ],
            batch_id="test_batch",
        )
        
        # Write
        csv_path = tmp_path / "labels.csv"
        writer = CSVLabelWriter()
        writer.write(labels, str(csv_path))
        
        assert csv_path.exists()
        
        # Read back
        reader = CSVLabelReader()
        loaded = reader.read(str(csv_path))
        
        assert len(loaded) == 2
        assert loaded.labels[0].sample_id == "sample_001"
        assert loaded.labels[1].fraud_type == "printed_photo"
    
    def test_label_validator(self):
        """Test label validation."""
        from ml.training.dataset.labeling import (
            Label,
            LabelBatch,
            LabelValidator,
            LabelSource,
        )
        
        labels = LabelBatch(
            labels=[
                Label(sample_id="s1", trust_score=80.0, is_genuine=True),
                Label(sample_id="s2", trust_score=25.0, is_genuine=False, fraud_type="fake"),
                Label(sample_id="s3", trust_score=150.0, is_genuine=True),  # Invalid score
            ],
        )
        
        validator = LabelValidator()
        report = validator.validate(labels)
        
        assert report["total"] == 3
        assert report["invalid"] == 1  # s3 has invalid score


class TestManifest:
    """Tests for manifest signing and verification."""
    
    def test_manifest_builder(self):
        """Test building a manifest."""
        from ml.training.dataset.manifest import ManifestBuilder
        
        builder = ManifestBuilder(
            dataset_name="test_dataset",
            schema_version="1.0.0",
        )
        
        builder.add_sample("s1", b"sample1_data", "train")
        builder.add_sample("s2", b"sample2_data", "train")
        builder.add_sample("s3", b"sample3_data", "test")
        
        manifest = builder.build("v1.0.0")
        
        assert manifest.dataset_name == "test_dataset"
        assert manifest.total_samples == 3
        assert manifest.split_counts["train"] == 2
        assert manifest.split_counts["test"] == 1
    
    def test_manifest_signing(self):
        """Test signing a manifest."""
        from ml.training.dataset.manifest import (
            ManifestBuilder,
            ManifestSigner,
            ManifestVerifier,
        )
        
        builder = ManifestBuilder(dataset_name="test")
        builder.add_sample("s1", b"data", "train")
        manifest = builder.build("v1")
        
        # Sign
        signer = ManifestSigner(signer_id="test_signer")
        signed = signer.sign(manifest)
        
        assert signed.signature is not None
        assert signed.signature.signer_id == "test_signer"
        assert signed.manifest_hash != ""
    
    def test_manifest_save_load(self, tmp_path):
        """Test saving and loading manifest."""
        from ml.training.dataset.manifest import (
            DatasetManifest,
            ManifestBuilder,
            ManifestSigner,
        )
        
        builder = ManifestBuilder(dataset_name="test")
        builder.add_sample("s1", b"data", "train")
        manifest = builder.build("v1")
        
        signer = ManifestSigner(signer_id="test")
        manifest = signer.sign(manifest)
        
        # Save
        manifest_path = tmp_path / "manifest.json"
        manifest.save(str(manifest_path))
        
        assert manifest_path.exists()
        
        # Load
        loaded = DatasetManifest.load(str(manifest_path))
        
        assert loaded.manifest_id == manifest.manifest_id
        assert loaded.manifest_hash == manifest.manifest_hash


class TestDeterministicSplits:
    """Tests for deterministic dataset splitting."""
    
    def test_deterministic_split(self):
        """Test that splits are deterministic."""
        from ml.training.dataset.splits import (
            DeterministicSplitter,
            SplitConfig,
        )
        
        samples = [
            IdentitySample(
                sample_id=f"sample_{i:03d}",
                trust_score=50 + i % 50,
                is_genuine=i % 3 != 0,
            )
            for i in range(100)
        ]
        
        config = SplitConfig(random_seed=42)
        
        splitter1 = DeterministicSplitter(config)
        splitter2 = DeterministicSplitter(config)
        
        result1 = splitter1.split(samples)
        result2 = splitter2.split(samples)
        
        assert result1.combined_hash == result2.combined_hash
    
    def test_stratified_split(self):
        """Test stratified splitting."""
        from ml.training.dataset.splits import (
            DeterministicSplitter,
            SplitConfig,
            SplitStrategy,
        )
        
        samples = [
            IdentitySample(
                sample_id=f"sample_{i:03d}",
                document_info=DocumentInfo(
                    doc_type="id_card" if i % 2 == 0 else "passport",
                    doc_id_hash=f"hash_{i}",
                ),
                trust_score=50 + i % 50,
                is_genuine=i % 3 != 0,
            )
            for i in range(100)
        ]
        
        config = SplitConfig(
            strategy=SplitStrategy.STRATIFIED,
            stratify_by=["doc_type"],
            random_seed=42,
        )
        
        splitter = DeterministicSplitter(config)
        result = splitter.split(samples)
        
        # Check that doc types are distributed across splits
        train_types = set(
            s.document_info.doc_type for s in result.dataset.train if s.document_info
        )
        test_types = set(
            s.document_info.doc_type for s in result.dataset.test if s.document_info
        )
        
        assert "id_card" in train_types
        assert "passport" in train_types
        assert "id_card" in test_types
        assert "passport" in test_types
    
    def test_split_verification(self):
        """Test split verification."""
        from ml.training.dataset.splits import (
            DeterministicSplitter,
            SplitConfig,
            SplitVerifier,
        )
        
        samples = [
            IdentitySample(sample_id=f"s{i}", trust_score=50)
            for i in range(50)
        ]
        
        config = SplitConfig(random_seed=42)
        splitter = DeterministicSplitter(config)
        result = splitter.split(samples)
        
        verifier = SplitVerifier()
        report = verifier.verify(result)
        
        assert report["valid"]
        assert report["checks"]["no_train_val_overlap"]
        assert report["checks"]["no_train_test_overlap"]
        assert report["checks"]["no_val_test_overlap"]


class TestLineage:
    """Tests for lineage tracking."""
    
    def test_lineage_tracker(self):
        """Test lineage tracking."""
        from ml.training.dataset.lineage import (
            LineageTracker,
            SourceType,
            TransformType,
        )
        
        tracker = LineageTracker(
            dataset_name="test",
            dataset_version="1.0.0",
        )
        
        tracker.add_source(
            location="/data/input",
            source_type=SourceType.LOCAL_FILE,
            record_count=100,
        )
        
        tracker.add_transform(
            transform_type=TransformType.PREPROCESSING,
            description="Normalize images",
            input_count=100,
            output_count=100,
        )
        
        lineage = tracker.finalize(sample_count=100)
        
        assert lineage.dataset_name == "test"
        assert len(lineage.sources) == 1
        assert len(lineage.transforms) == 1
        assert lineage.sample_count == 100
    
    def test_lineage_save_load(self, tmp_path):
        """Test saving and loading lineage."""
        from ml.training.dataset.lineage import (
            DatasetLineage,
            LineageTracker,
            SourceType,
        )
        
        tracker = LineageTracker("test", "1.0")
        tracker.add_source("/data", SourceType.LOCAL_FILE, record_count=50)
        lineage = tracker.finalize(sample_count=50)
        
        # Save
        lineage_path = tmp_path / "lineage.json"
        lineage.save(str(lineage_path))
        
        assert lineage_path.exists()
        
        # Load
        loaded = DatasetLineage.load(str(lineage_path))
        
        assert loaded.dataset_name == "test"
        assert len(loaded.sources) == 1


class TestValidation:
    """Tests for dataset validation."""
    
    def test_schema_validation(self):
        """Test schema validation."""
        from ml.training.dataset.validation import SchemaValidator
        
        validator = SchemaValidator()
        
        # Valid sample
        valid_sample = IdentitySample(
            sample_id="test_001",
            trust_score=75.0,
            is_genuine=True,
            face_detected=True,
            face_confidence=0.95,
            document_quality_score=0.85,
            ocr_success=True,
            ocr_confidence=0.90,
        )
        
        issues = validator.validate_sample(valid_sample)
        errors = [i for i in issues if i.level.value == "error"]
        assert len(errors) == 0
        
        # Invalid sample (score out of range)
        invalid_sample = IdentitySample(
            sample_id="test_002",
            trust_score=150.0,  # Invalid
            is_genuine=True,
            face_detected=True,
            face_confidence=0.95,
            document_quality_score=0.85,
            ocr_success=True,
            ocr_confidence=0.90,
        )
        
        issues = validator.validate_sample(invalid_sample)
        errors = [i for i in issues if i.level.value == "error"]
        assert len(errors) > 0
    
    def test_label_anomaly_detection(self):
        """Test label anomaly detection."""
        from ml.training.dataset.validation import LabelAnomalyDetector
        
        detector = LabelAnomalyDetector()
        
        samples = [
            IdentitySample(
                sample_id=f"s{i}",
                trust_score=75 + np.random.randn() * 10,
                is_genuine=True,
                face_detected=True,
                face_confidence=0.9,
                document_quality_score=0.8,
                ocr_success=True,
                ocr_confidence=0.85,
            )
            for i in range(50)
        ]
        
        # Add an outlier
        samples.append(IdentitySample(
            sample_id="outlier",
            trust_score=5.0,  # Very low for genuine
            is_genuine=True,
            face_detected=True,
            face_confidence=0.9,
            document_quality_score=0.8,
            ocr_success=True,
            ocr_confidence=0.85,
        ))
        
        issues = detector.detect_anomalies(samples)
        
        # Should detect the inconsistent genuine/score
        warnings = [i for i in issues if "outlier" in str(i.sample_id)]
        assert len(warnings) > 0
    
    def test_full_validation(self):
        """Test full dataset validation."""
        from ml.training.dataset.validation import DatasetValidator, validate_dataset
        from ml.training.dataset.synthetic import generate_synthetic_dataset
        
        dataset = generate_synthetic_dataset(num_samples=30, seed=42)
        
        report = validate_dataset(dataset, fail_on_error=False)
        
        assert report.samples_validated == 30
        assert "valid" in report.to_dict()


class TestSecureStorage:
    """Tests for secure storage."""
    
    def test_encryption_manager(self):
        """Test encryption manager."""
        from ml.training.dataset.storage import EncryptionManager
        
        manager = EncryptionManager()
        
        data = b"test data to encrypt"
        envelope = manager.encrypt(data, "test_key")
        
        assert envelope.ciphertext != data
        
        decrypted = manager.decrypt(envelope)
        assert decrypted == data
    
    def test_secure_storage_images(self, tmp_path):
        """Test storing and retrieving encrypted images."""
        from ml.training.dataset.storage import SecureStorage, StorageConfig
        
        config = StorageConfig(
            raw_data_path=str(tmp_path / "raw"),
            derived_data_path=str(tmp_path / "derived"),
            metadata_path=str(tmp_path / "metadata"),
            audit_log_path=str(tmp_path / "audit.log"),
        )
        
        storage = SecureStorage(config)
        
        # Store image
        image = np.random.randint(0, 255, (100, 100, 3), dtype=np.uint8)
        asset = storage.store_raw_image("sample_001", "document", image)
        
        assert asset.encrypted
        assert asset.asset_id == "sample_001_document"
        
        # Retrieve
        retrieved = storage.retrieve_raw_image("sample_001", "document")
        
        assert retrieved is not None
        assert np.array_equal(retrieved, image)
    
    def test_integrity_verification(self, tmp_path):
        """Test integrity verification."""
        from ml.training.dataset.storage import SecureStorage, StorageConfig
        
        config = StorageConfig(
            raw_data_path=str(tmp_path / "raw"),
            derived_data_path=str(tmp_path / "derived"),
            metadata_path=str(tmp_path / "metadata"),
            require_audit_log=False,
        )
        
        storage = SecureStorage(config)
        
        # Store some data
        image = np.random.randint(0, 255, (50, 50, 3), dtype=np.uint8)
        storage.store_raw_image("test_001", "selfie", image)
        
        # Verify
        report = storage.verify_integrity()
        
        assert report["verified"] == 1
        assert report["corrupted"] == 0
