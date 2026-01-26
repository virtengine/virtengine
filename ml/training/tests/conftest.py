"""
Pytest fixtures for training pipeline tests.
"""

import pytest
import numpy as np
from typing import List

from ml.training.config import (
    TrainingConfig,
    DatasetConfig,
    FeatureConfig,
    ModelConfig,
    ExportConfig,
)
from ml.training.dataset.ingestion import (
    IdentitySample,
    ImageData,
    DocumentInfo,
    CaptureMetadata,
    Dataset,
    DatasetSplit,
    SplitType,
)
from ml.training.features.feature_combiner import FeatureVector, FeatureDataset


@pytest.fixture
def sample_config() -> TrainingConfig:
    """Create a sample training configuration for testing."""
    return TrainingConfig(
        dataset=DatasetConfig(
            data_paths=[],
            train_split=0.7,
            val_split=0.15,
            test_split=0.15,
            anonymize=True,
            random_seed=42,
        ),
        model=ModelConfig(
            input_dim=128,  # Smaller for testing
            hidden_layers=[64, 32],
            dropout_rate=0.2,
            epochs=5,
            batch_size=8,
        ),
        random_seed=42,
    )


@pytest.fixture
def sample_image() -> np.ndarray:
    """Create a sample image array for testing."""
    return np.random.randint(0, 255, size=(224, 224, 3), dtype=np.uint8)


@pytest.fixture
def sample_identity_sample(sample_image: np.ndarray) -> IdentitySample:
    """Create a sample identity sample for testing."""
    return IdentitySample(
        sample_id="test_sample_001",
        document_image=ImageData(
            data=sample_image,
            path="/test/document.jpg",
            format="RGB",
        ),
        selfie_image=ImageData(
            data=sample_image,
            path="/test/selfie.jpg",
            format="RGB",
        ),
        document_info=DocumentInfo(
            doc_type="id_card",
            doc_id_hash="abc123",
            country_code="US",
            expiry_valid=True,
            mrz_present=True,
        ),
        capture_metadata=CaptureMetadata(
            device_type="mobile",
            device_model="iPhone 14",
            os_version="iOS 17.0",
            app_version="1.0.0",
            capture_timestamp=1705000000.0,
            gps_available=True,
            camera_facing="front",
            light_level=500.0,
            motion_detected=False,
        ),
        trust_score=85.0,
        is_genuine=True,
        face_detected=True,
        face_confidence=0.95,
        document_quality_score=0.88,
        ocr_success=True,
        ocr_confidence=0.92,
    )


@pytest.fixture
def sample_dataset(sample_identity_sample: IdentitySample) -> Dataset:
    """Create a sample dataset for testing."""
    # Create multiple samples with variations
    samples = []
    for i in range(30):
        sample = IdentitySample(
            sample_id=f"sample_{i:03d}",
            document_info=DocumentInfo(
                doc_type=["id_card", "passport", "drivers_license"][i % 3],
                doc_id_hash=f"hash_{i}",
            ),
            trust_score=50.0 + (i % 50),
            is_genuine=i % 5 != 0,  # 20% fraud
            face_confidence=0.7 + (i % 30) / 100.0,
            document_quality_score=0.6 + (i % 40) / 100.0,
            ocr_confidence=0.5 + (i % 50) / 100.0,
        )
        samples.append(sample)
    
    # Split into train/val/test
    train_samples = samples[:20]
    val_samples = samples[20:25]
    test_samples = samples[25:30]
    
    return Dataset(
        train=DatasetSplit(SplitType.TRAIN, train_samples),
        validation=DatasetSplit(SplitType.VALIDATION, val_samples),
        test=DatasetSplit(SplitType.TEST, test_samples),
    )


@pytest.fixture
def sample_feature_vector() -> FeatureVector:
    """Create a sample feature vector for testing."""
    return FeatureVector(
        sample_id="test_001",
        face_embedding=np.random.randn(512).astype(np.float32),
        face_confidence=0.95,
        doc_quality_score=0.88,
        ocr_field_scores={"name": 0.92, "dob": 0.85},
        combined_vector=np.random.randn(768).astype(np.float32),
        trust_score=85.0,
    )


@pytest.fixture
def sample_feature_dataset(sample_feature_vector: FeatureVector) -> FeatureDataset:
    """Create a sample feature dataset for testing."""
    # Create multiple feature vectors
    train_features = []
    val_features = []
    test_features = []
    
    for i in range(50):
        fv = FeatureVector(
            sample_id=f"sample_{i:03d}",
            combined_vector=np.random.randn(768).astype(np.float32),
            trust_score=30.0 + np.random.rand() * 60,  # 30-90 range
        )
        if i < 35:
            train_features.append(fv)
        elif i < 42:
            val_features.append(fv)
        else:
            test_features.append(fv)
    
    return FeatureDataset(
        train=train_features,
        validation=val_features,
        test=test_features,
        feature_dim=768,
    )


@pytest.fixture
def sample_train_arrays() -> tuple:
    """Create sample training arrays for testing."""
    n_train = 100
    n_val = 20
    n_test = 20
    feature_dim = 128
    
    # Generate random features
    train_X = np.random.randn(n_train, feature_dim).astype(np.float32)
    train_y = 30 + np.random.rand(n_train) * 60  # 30-90 range
    
    val_X = np.random.randn(n_val, feature_dim).astype(np.float32)
    val_y = 30 + np.random.rand(n_val) * 60
    
    test_X = np.random.randn(n_test, feature_dim).astype(np.float32)
    test_y = 30 + np.random.rand(n_test) * 60
    
    return train_X, train_y, val_X, val_y, test_X, test_y
