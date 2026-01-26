"""
Tests for dataset ingestion and preprocessing.
"""

import pytest
import numpy as np
from pathlib import Path
import tempfile
import json

from ml.training.config import DatasetConfig
from ml.training.dataset.ingestion import (
    DatasetIngestion,
    IdentitySample,
    ImageData,
    DocumentInfo,
    Dataset,
    DatasetSplit,
    SplitType,
)
from ml.training.dataset.preprocessing import DatasetPreprocessor
from ml.training.dataset.augmentation import DataAugmentation
from ml.training.dataset.anonymization import PIIAnonymizer


class TestDatasetConfig:
    """Tests for DatasetConfig."""
    
    def test_default_config(self):
        """Test default configuration values."""
        config = DatasetConfig()
        
        assert config.train_split == 0.8
        assert config.val_split == 0.1
        assert config.test_split == 0.1
        assert config.anonymize is True
        assert config.random_seed == 42
    
    def test_split_validation(self):
        """Test that splits must sum to 1."""
        with pytest.raises(AssertionError):
            DatasetConfig(
                train_split=0.5,
                val_split=0.1,
                test_split=0.1,  # Sum = 0.7, not 1.0
            )
    
    def test_valid_splits(self):
        """Test valid split configuration."""
        config = DatasetConfig(
            train_split=0.7,
            val_split=0.2,
            test_split=0.1,
        )
        
        total = config.train_split + config.val_split + config.test_split
        assert abs(total - 1.0) < 1e-6


class TestIdentitySample:
    """Tests for IdentitySample."""
    
    def test_sample_creation(self, sample_identity_sample):
        """Test creating an identity sample."""
        sample = sample_identity_sample
        
        assert sample.sample_id == "test_sample_001"
        assert sample.trust_score == 85.0
        assert sample.is_genuine is True
        assert sample.face_detected is True
    
    def test_sample_to_dict(self, sample_identity_sample):
        """Test converting sample to dictionary."""
        sample = sample_identity_sample
        data = sample.to_dict()
        
        assert "sample_id" in data
        assert "trust_score" in data
        assert data["trust_score"] == 85.0


class TestDatasetIngestion:
    """Tests for DatasetIngestion."""
    
    def test_empty_dataset(self):
        """Test loading with no data paths."""
        config = DatasetConfig(data_paths=[])
        ingestion = DatasetIngestion(config)
        
        dataset = ingestion.load_dataset()
        
        assert len(dataset) == 0
        assert len(dataset.train) == 0
        assert len(dataset.validation) == 0
        assert len(dataset.test) == 0
    
    def test_load_from_manifest(self):
        """Test loading from a manifest file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create manifest
            manifest = {
                "samples": [
                    {
                        "sample_id": "sample_001",
                        "document_type": "id_card",
                        "trust_score": 75.0,
                        "is_genuine": True,
                    },
                    {
                        "sample_id": "sample_002",
                        "document_type": "passport",
                        "trust_score": 82.0,
                        "is_genuine": True,
                    },
                ]
            }
            
            manifest_path = Path(tmpdir) / "manifest.json"
            with open(manifest_path, 'w') as f:
                json.dump(manifest, f)
            
            config = DatasetConfig(
                data_paths=[tmpdir],
                min_face_confidence=0.0,
                min_doc_quality=0.0,
                min_ocr_confidence=0.0,
            )
            ingestion = DatasetIngestion(config)
            
            dataset = ingestion.load_dataset()
            
            # Check that samples were loaded
            assert len(dataset) == 2
    
    def test_dataset_splitting(self, sample_dataset):
        """Test that dataset is split correctly."""
        dataset = sample_dataset
        
        assert len(dataset.train) == 20
        assert len(dataset.validation) == 5
        assert len(dataset.test) == 5
    
    def test_dataset_statistics(self, sample_dataset):
        """Test dataset statistics computation."""
        stats = sample_dataset.get_statistics()
        
        assert "total_samples" in stats
        assert "train" in stats
        assert "validation" in stats
        assert "test" in stats
        assert stats["total_samples"] == 30


class TestDatasetPreprocessor:
    """Tests for DatasetPreprocessor."""
    
    def test_preprocessor_creation(self):
        """Test creating a preprocessor."""
        preprocessor = DatasetPreprocessor()
        assert preprocessor is not None
    
    def test_preprocess_sample(self, sample_identity_sample):
        """Test preprocessing a single sample."""
        preprocessor = DatasetPreprocessor()
        
        result = preprocessor.preprocess_sample(sample_identity_sample)
        
        assert result.sample_id == sample_identity_sample.sample_id
        assert result.success is True


class TestDataAugmentation:
    """Tests for DataAugmentation."""
    
    def test_augmentation_disabled(self):
        """Test augmentation when disabled."""
        from ml.training.config import AugmentationConfig
        
        config = AugmentationConfig(enabled=False)
        augmentor = DataAugmentation(config)
        
        # Create mock preprocessed samples
        from ml.training.dataset.preprocessing import PreprocessedSample
        
        samples = [
            PreprocessedSample(
                sample_id="test_001",
                document_image=np.random.randn(224, 224, 3).astype(np.float32),
            )
        ]
        
        augmented = augmentor.augment_batch(samples)
        
        # Should only return original samples
        assert len(augmented) == 1
    
    def test_augmentation_enabled(self):
        """Test augmentation when enabled."""
        from ml.training.config import AugmentationConfig
        from ml.training.dataset.preprocessing import PreprocessedSample
        
        config = AugmentationConfig(
            enabled=True,
            num_augmented_copies=2,
        )
        augmentor = DataAugmentation(config)
        
        samples = [
            PreprocessedSample(
                sample_id="test_001",
                document_image=np.random.randn(224, 224, 3).astype(np.float32),
            )
        ]
        
        augmented = augmentor.augment_batch(samples, seed=42)
        
        # Should return original + 2 augmented
        assert len(augmented) == 3


class TestPIIAnonymizer:
    """Tests for PIIAnonymizer."""
    
    def test_anonymizer_creation(self):
        """Test creating an anonymizer."""
        anonymizer = PIIAnonymizer()
        assert anonymizer is not None
    
    def test_hash_anonymization(self):
        """Test SHA256 hash anonymization."""
        anonymizer = PIIAnonymizer(salt="test_salt")
        
        value1 = anonymizer.anonymize_value("test_value")
        value2 = anonymizer.anonymize_value("test_value")
        
        # Same input should give same output
        assert value1 == value2
        # Should be different from input
        assert value1 != "test_value"
    
    def test_dataset_anonymization(self, sample_dataset):
        """Test anonymizing a dataset."""
        anonymizer = PIIAnonymizer()
        
        original_ids = [s.sample_id for s in sample_dataset.train.samples]
        
        result = anonymizer.anonymize_dataset(sample_dataset)
        
        assert result.success
        assert result.samples_processed == 30
        
        # IDs should be changed
        new_ids = [s.sample_id for s in sample_dataset.train.samples]
        assert new_ids != original_ids
    
    def test_anonymization_report(self):
        """Test getting anonymization report."""
        anonymizer = PIIAnonymizer()
        
        report = anonymizer.get_anonymization_report()
        
        assert "method" in report
        assert "salt_hash" in report
        assert "total_pii_fields" in report
