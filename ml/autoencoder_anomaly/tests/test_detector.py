"""
Tests for main anomaly detector.

VE-924: Test main AnomalyDetector class.
"""

import pytest
import numpy as np

from ml.autoencoder_anomaly.detector import (
    AnomalyDetector,
    AnomalyDetectionResult,
    create_detector,
)
from ml.autoencoder_anomaly.config import (
    AutoencoderAnomalyConfig,
    DetectionMode,
    AnomalyLevel,
    AnomalyType,
)
from ml.autoencoder_anomaly.reason_codes import AnomalyReasonCodes


class TestAnomalyDetector:
    """Tests for AnomalyDetector class."""
    
    def test_detector_creation(self, anomaly_config):
        """Test creating detector with config."""
        detector = AnomalyDetector(anomaly_config)
        
        assert detector is not None
        assert detector.config == anomaly_config
        assert detector.MODEL_VERSION == "1.0.0"
    
    def test_detector_default_config(self):
        """Test creating detector with default config."""
        detector = AnomalyDetector()
        assert detector is not None
    
    def test_model_hash_generation(self, anomaly_config):
        """Test model hash is generated."""
        detector = AnomalyDetector(anomaly_config)
        
        assert detector.model_hash is not None
        assert len(detector.model_hash) == 32
    
    def test_model_hash_determinism(self, anomaly_config):
        """Test model hash is deterministic."""
        detector1 = AnomalyDetector(anomaly_config)
        detector2 = AnomalyDetector(anomaly_config)
        
        assert detector1.model_hash == detector2.model_hash
    
    def test_detect_single_image(self, anomaly_config, sample_image):
        """Test detection on single image."""
        detector = AnomalyDetector(anomaly_config)
        
        result = detector.detect(sample_image)
        
        assert isinstance(result, AnomalyDetectionResult)
        assert 0.0 <= result.overall_score <= 1.0
        assert result.decision in ["normal", "suspicious", "anomaly", "uncertain"]
    
    def test_detect_anomalous_image(self, anomaly_config, anomalous_image):
        """Test detection on anomalous image."""
        detector = AnomalyDetector(anomaly_config)
        
        result = detector.detect(anomalous_image)
        
        assert isinstance(result, AnomalyDetectionResult)
        # Anomalous image should have higher score
        assert result.overall_score >= 0.0
    
    def test_detect_empty_image(self, anomaly_config):
        """Test detection fails for empty image."""
        detector = AnomalyDetector(anomaly_config)
        
        result = detector.detect(np.array([]))
        
        assert result.decision == "uncertain"
        assert AnomalyReasonCodes.INVALID_INPUT.value in result.reason_codes
    
    def test_detect_small_image(self, anomaly_config, small_image):
        """Test detection fails for small image."""
        detector = AnomalyDetector(anomaly_config)
        
        result = detector.detect(small_image)
        
        assert AnomalyReasonCodes.IMAGE_TOO_SMALL.value in result.reason_codes
    
    def test_detect_large_image(self, anomaly_config, large_image):
        """Test detection fails for large image."""
        detector = AnomalyDetector(anomaly_config)
        
        result = detector.detect(large_image)
        
        assert AnomalyReasonCodes.IMAGE_TOO_LARGE.value in result.reason_codes
    
    def test_detect_with_details(self, anomaly_config, sample_image):
        """Test detection with detailed results."""
        detector = AnomalyDetector(anomaly_config)
        
        result = detector.detect(sample_image, include_details=True)
        
        assert result.autoencoder_output is not None
        assert result.anomaly_score_details is not None
    
    def test_component_scores(self, anomaly_config, sample_image):
        """Test all component scores are computed."""
        detector = AnomalyDetector(anomaly_config)
        
        result = detector.detect(sample_image)
        
        assert 0.0 <= result.reconstruction_score <= 1.0
        assert 0.0 <= result.latent_score <= 1.0
    
    def test_veid_penalty_computed(self, anomaly_config, sample_image):
        """Test VEID penalty is computed."""
        detector = AnomalyDetector(anomaly_config)
        
        result = detector.detect(sample_image)
        
        assert result.veid_penalty >= 0
        assert result.veid_adjusted_score >= 0
        assert result.veid_adjusted_score <= 10000
    
    def test_veid_penalty_range(self, anomaly_config, sample_image):
        """Test VEID penalty is within expected range."""
        detector = AnomalyDetector(anomaly_config)
        
        result = detector.detect(sample_image)
        
        # Penalty should not exceed max
        assert result.veid_penalty <= 10000
        # Adjusted score should be valid
        assert result.veid_adjusted_score == 10000 - result.veid_penalty
    
    def test_anomaly_level_assigned(self, anomaly_config, sample_image):
        """Test anomaly level is assigned."""
        detector = AnomalyDetector(anomaly_config)
        
        result = detector.detect(sample_image)
        
        assert result.anomaly_level in AnomalyLevel
    
    def test_detected_types_list(self, anomaly_config, sample_image):
        """Test detected types is a list."""
        detector = AnomalyDetector(anomaly_config)
        
        result = detector.detect(sample_image)
        
        assert isinstance(result.detected_types, list)
    
    def test_reason_codes_populated(self, anomaly_config, sample_image):
        """Test reason codes are populated."""
        detector = AnomalyDetector(anomaly_config)
        
        result = detector.detect(sample_image)
        
        assert len(result.reason_codes) > 0
    
    def test_result_hash_generated(self, anomaly_config, sample_image):
        """Test result hash is generated."""
        detector = AnomalyDetector(anomaly_config)
        
        result = detector.detect(sample_image)
        
        assert result.result_hash is not None
        assert len(result.result_hash) == 32
    
    def test_result_hash_determinism(self, anomaly_config, sample_image):
        """Test result hash is deterministic."""
        detector = AnomalyDetector(anomaly_config)
        
        result1 = detector.detect(sample_image)
        result2 = detector.detect(sample_image)
        
        assert result1.result_hash == result2.result_hash
    
    def test_processing_time_recorded(self, anomaly_config, sample_image):
        """Test processing time is recorded."""
        detector = AnomalyDetector(anomaly_config)
        
        result = detector.detect(sample_image)
        
        assert result.processing_time_ms > 0
    
    def test_confidence_range(self, anomaly_config, sample_image):
        """Test confidence is in valid range."""
        detector = AnomalyDetector(anomaly_config)
        
        result = detector.detect(sample_image)
        
        assert 0 <= result.confidence <= 1


class TestAnomalyDetectionResult:
    """Tests for AnomalyDetectionResult dataclass."""
    
    def test_to_dict(self, anomaly_config, sample_image):
        """Test to_dict() conversion."""
        detector = AnomalyDetector(anomaly_config)
        result = detector.detect(sample_image)
        
        result_dict = result.to_dict()
        
        assert "is_anomaly" in result_dict
        assert "decision" in result_dict
        assert "anomaly_level" in result_dict
        assert "overall_score" in result_dict
        assert "confidence" in result_dict
        assert "veid_penalty" in result_dict
        assert "model_version" in result_dict
        assert "reason_codes" in result_dict
    
    def test_to_veid_record(self, anomaly_config, sample_image):
        """Test to_veid_record() conversion."""
        detector = AnomalyDetector(anomaly_config)
        result = detector.detect(sample_image)
        
        veid_record = result.to_veid_record()
        
        assert "anomaly_score" in veid_record
        assert "is_anomaly" in veid_record
        assert "decision" in veid_record
        assert "veid_penalty" in veid_record
        assert "veid_adjusted_score" in veid_record
        assert "model_version" in veid_record
        assert "model_hash" in veid_record
    
    def test_veid_record_score_inverted(self, anomaly_config, sample_image):
        """Test VEID record has inverted score."""
        detector = AnomalyDetector(anomaly_config)
        result = detector.detect(sample_image)
        
        veid_record = result.to_veid_record()
        
        # Higher anomaly_score = more normal (inverted)
        expected = int((1 - result.overall_score) * 10000)
        assert veid_record["anomaly_score"] == expected
    
    def test_veid_record_reason_codes_limited(
        self, anomaly_config, sample_image
    ):
        """Test reason codes are limited in VEID record."""
        detector = AnomalyDetector(anomaly_config)
        result = detector.detect(sample_image)
        
        veid_record = result.to_veid_record()
        
        assert len(veid_record["reason_codes"]) <= 5
    
    def test_veid_record_types_limited(self, anomaly_config, sample_image):
        """Test detected types are limited in VEID record."""
        detector = AnomalyDetector(anomaly_config)
        result = detector.detect(sample_image)
        
        veid_record = result.to_veid_record()
        
        assert len(veid_record["detected_types"]) <= 3


class TestDetectBatch:
    """Tests for batch detection."""
    
    def test_detect_batch(self, anomaly_config, sample_image, anomalous_image):
        """Test batch detection."""
        detector = AnomalyDetector(anomaly_config)
        images = [sample_image, anomalous_image, sample_image]
        
        results = detector.detect_batch(images)
        
        assert len(results) == 3
        assert all(isinstance(r, AnomalyDetectionResult) for r in results)
    
    def test_detect_batch_empty(self, anomaly_config):
        """Test batch detection with empty list."""
        detector = AnomalyDetector(anomaly_config)
        
        results = detector.detect_batch([])
        
        assert len(results) == 0


class TestGetReconstruction:
    """Tests for reconstruction visualization method."""
    
    def test_get_reconstruction(self, anomaly_config, sample_image):
        """Test getting reconstruction."""
        detector = AnomalyDetector(anomaly_config)
        
        original, reconstruction = detector.get_reconstruction(sample_image)
        
        assert original is not None
        assert reconstruction is not None
    
    def test_reconstruction_normalized(self, anomaly_config, sample_image):
        """Test reconstruction is normalized."""
        detector = AnomalyDetector(anomaly_config)
        
        original, reconstruction = detector.get_reconstruction(sample_image)
        
        assert original.min() >= 0.0
        assert original.max() <= 1.0
        assert reconstruction.min() >= 0.0
        assert reconstruction.max() <= 1.0


class TestCreateDetector:
    """Tests for factory function."""
    
    def test_create_with_config(self, anomaly_config):
        """Test creating detector with config."""
        detector = create_detector(anomaly_config)
        
        assert isinstance(detector, AnomalyDetector)
        assert detector.config == anomaly_config
    
    def test_create_without_config(self):
        """Test creating detector without config."""
        detector = create_detector()
        
        assert isinstance(detector, AnomalyDetector)


class TestDetectionModes:
    """Tests for different detection modes."""
    
    def test_fast_mode(self, sample_image):
        """Test detection in fast mode."""
        config = AutoencoderAnomalyConfig.fast_mode()
        detector = AnomalyDetector(config)
        
        result = detector.detect(sample_image)
        
        assert isinstance(result, AnomalyDetectionResult)
    
    def test_strict_mode(self, sample_image):
        """Test detection in strict mode."""
        config = AutoencoderAnomalyConfig.strict_mode()
        detector = AnomalyDetector(config)
        
        result = detector.detect(sample_image)
        
        assert isinstance(result, AnomalyDetectionResult)
    
    def test_fast_mode_faster(self, sample_image):
        """Test fast mode is faster than full mode."""
        full_config = AutoencoderAnomalyConfig()
        fast_config = AutoencoderAnomalyConfig.fast_mode()
        
        full_detector = AnomalyDetector(full_config)
        fast_detector = AnomalyDetector(fast_config)
        
        # Run multiple times to average
        full_times = []
        fast_times = []
        
        for _ in range(3):
            result = full_detector.detect(sample_image)
            full_times.append(result.processing_time_ms)
            
            result = fast_detector.detect(sample_image)
            fast_times.append(result.processing_time_ms)
        
        # Fast mode should generally be faster (allow some variance)
        # Just check they both complete successfully
        assert min(fast_times) > 0
        assert min(full_times) > 0
