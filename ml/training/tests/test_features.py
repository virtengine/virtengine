"""
Tests for feature extraction.
"""

import pytest
import numpy as np

from ml.training.config import FeatureConfig
from ml.training.features.face_features import FaceFeatureExtractor, FaceFeatures
from ml.training.features.doc_features import DocumentFeatureExtractor, DocumentFeatures
from ml.training.features.ocr_features import OCRFeatureExtractor, OCRFeatures
from ml.training.features.metadata_features import MetadataFeatureExtractor, MetadataFeatures
from ml.training.features.feature_combiner import FeatureExtractor, FeatureVector


class TestFaceFeatures:
    """Tests for face feature extraction."""
    
    def test_face_features_creation(self):
        """Test creating face features."""
        features = FaceFeatures(
            face_similarity=0.95,
            document_face_confidence=0.92,
            selfie_face_confidence=0.88,
        )
        
        assert features.face_similarity == 0.95
        assert features.document_face_confidence == 0.92
    
    def test_face_features_to_vector(self):
        """Test converting face features to vector."""
        features = FaceFeatures(
            document_face_embedding=np.random.randn(512).astype(np.float32),
            selfie_face_embedding=np.random.randn(512).astype(np.float32),
            face_similarity=0.95,
        )
        
        vector = features.to_vector(embedding_dim=512)
        
        # 2 * 512 + 7 scalar features
        assert len(vector) == 512 * 2 + 7
    
    def test_face_extractor_creation(self):
        """Test creating face feature extractor."""
        config = FeatureConfig()
        extractor = FaceFeatureExtractor(config)
        
        assert extractor is not None
        assert extractor.get_feature_dim() == 512 * 2 + 7
    
    def test_face_extraction_empty_input(self):
        """Test face extraction with no images."""
        extractor = FaceFeatureExtractor()
        
        features = extractor.extract(None, None)
        
        assert features.document_face_detected is False
        assert features.selfie_face_detected is False


class TestDocumentFeatures:
    """Tests for document feature extraction."""
    
    def test_doc_features_creation(self):
        """Test creating document features."""
        features = DocumentFeatures(
            sharpness_score=0.85,
            brightness_score=0.78,
            overall_quality_score=0.80,
        )
        
        assert features.sharpness_score == 0.85
        assert features.overall_quality_score == 0.80
    
    def test_doc_features_to_vector(self):
        """Test converting document features to vector."""
        features = DocumentFeatures(
            sharpness_score=0.85,
            brightness_score=0.78,
            contrast_score=0.82,
            overall_quality_score=0.80,
        )
        
        vector = features.to_vector()
        
        assert len(vector) == 17
        assert vector[0] == pytest.approx(0.85)  # sharpness_score
    
    def test_doc_extractor_creation(self):
        """Test creating document feature extractor."""
        extractor = DocumentFeatureExtractor()
        
        assert extractor is not None
        assert extractor.get_feature_dim() == 17
    
    def test_doc_extraction_with_image(self, sample_image):
        """Test document feature extraction."""
        extractor = DocumentFeatureExtractor()
        
        features = extractor.extract(sample_image)
        
        assert 0 <= features.sharpness_score <= 1
        assert 0 <= features.brightness_score <= 1
        assert 0 <= features.overall_quality_score <= 1


class TestOCRFeatures:
    """Tests for OCR feature extraction."""
    
    def test_ocr_features_creation(self):
        """Test creating OCR features."""
        features = OCRFeatures(
            overall_ocr_confidence=0.88,
            fields_extracted_ratio=0.8,
            ocr_success=True,
        )
        
        assert features.overall_ocr_confidence == 0.88
        assert features.ocr_success is True
    
    def test_ocr_features_to_vector(self):
        """Test converting OCR features to vector."""
        features = OCRFeatures(
            overall_ocr_confidence=0.88,
            fields_extracted_ratio=0.8,
            has_name=True,
            has_dob=True,
        )
        
        expected_fields = ["name", "date_of_birth", "document_number", "expiry_date", "nationality"]
        vector = features.to_vector(expected_fields)
        
        # 5 fields * 5 features + 12 aggregate
        assert len(vector) == 5 * 5 + 12
    
    def test_ocr_extractor_creation(self):
        """Test creating OCR feature extractor."""
        extractor = OCRFeatureExtractor()
        
        assert extractor is not None


class TestMetadataFeatures:
    """Tests for metadata feature extraction."""
    
    def test_metadata_features_creation(self):
        """Test creating metadata features."""
        features = MetadataFeatures(
            light_level_score=0.5,
            gps_available=True,
            metadata_complete=True,
        )
        
        assert features.light_level_score == 0.5
        assert features.gps_available is True
    
    def test_metadata_features_to_vector(self):
        """Test converting metadata features to vector."""
        features = MetadataFeatures()
        
        vector = features.to_vector()
        
        # Device types (6) + OS types (6) + Camera facing (3) + 9 scalars
        assert len(vector) == 6 + 6 + 3 + 9
    
    def test_metadata_extractor_creation(self):
        """Test creating metadata feature extractor."""
        extractor = MetadataFeatureExtractor()
        
        assert extractor is not None
    
    def test_metadata_extraction(self, sample_identity_sample):
        """Test metadata feature extraction."""
        extractor = MetadataFeatureExtractor()
        
        features = extractor.extract(sample_identity_sample.capture_metadata)
        
        assert features.metadata_complete is True
        # Should detect mobile device
        assert features.device_type_encoding[0] == 1.0  # mobile index


class TestFeatureCombiner:
    """Tests for combined feature extraction."""
    
    def test_feature_vector_creation(self, sample_feature_vector):
        """Test creating a feature vector."""
        fv = sample_feature_vector
        
        assert fv.sample_id == "test_001"
        assert fv.trust_score == 85.0
        assert len(fv.combined_vector) == 768
    
    def test_feature_vector_to_model_input(self, sample_feature_vector):
        """Test getting model input from feature vector."""
        fv = sample_feature_vector
        
        model_input = fv.to_model_input()
        
        assert model_input.dtype == np.float32
        assert len(model_input) == 768
    
    def test_feature_extractor_creation(self):
        """Test creating the combined feature extractor."""
        config = FeatureConfig(combined_feature_dim=512)
        extractor = FeatureExtractor(config)
        
        assert extractor is not None
        assert extractor.get_feature_dim() == 512
    
    def test_feature_extraction_from_sample(self, sample_identity_sample):
        """Test extracting features from an identity sample."""
        config = FeatureConfig(combined_feature_dim=256)
        extractor = FeatureExtractor(config)
        
        features = extractor.extract_all(sample_identity_sample)
        
        assert features.sample_id == sample_identity_sample.sample_id
        assert features.trust_score == sample_identity_sample.trust_score
        assert len(features.combined_vector) == 256
    
    def test_feature_dataset_arrays(self, sample_feature_dataset):
        """Test getting arrays from feature dataset."""
        fds = sample_feature_dataset
        
        train_X, train_y = fds.get_arrays("train")
        
        assert len(train_X) == len(fds.train)
        assert len(train_y) == len(fds.train)
        assert train_X.shape[1] == 768
    
    def test_feature_dataset_normalization(self, sample_feature_dataset):
        """Test feature normalization."""
        fds = sample_feature_dataset
        
        # Get pre-normalization mean
        train_X_before, _ = fds.get_arrays("train")
        mean_before = np.mean(train_X_before)
        
        fds.normalize()
        
        # After normalization, mean should be near 0
        train_X_after, _ = fds.get_arrays("train")
        mean_after = np.mean(train_X_after)
        
        assert abs(mean_after) < abs(mean_before) or abs(mean_after) < 0.1
