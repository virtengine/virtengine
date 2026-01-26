"""
Tests for main GAN detector.

VE-923: Test main GANDetector class.
"""

import pytest
import numpy as np

from ml.gan_detection.detector import (
    GANDetector,
    GANDetectionResult,
    create_detector,
)
from ml.gan_detection.config import GANDetectionConfig, DetectionMode, SyntheticImageType
from ml.gan_detection.reason_codes import GANReasonCodes


class TestGANDetector:
    """Tests for GANDetector class."""
    
    def test_detector_creation(self, gan_config):
        """Test creating detector with config."""
        detector = GANDetector(gan_config)
        
        assert detector is not None
        assert detector.config == gan_config
        assert detector.MODEL_VERSION == "1.0.0"
    
    def test_detector_default_config(self):
        """Test creating detector with default config."""
        detector = GANDetector()
        assert detector is not None
    
    def test_model_hash_generation(self, gan_config):
        """Test model hash is generated."""
        detector = GANDetector(gan_config)
        
        assert detector.model_hash is not None
        assert len(detector.model_hash) == 32
    
    def test_model_hash_determinism(self, gan_config):
        """Test model hash is deterministic."""
        detector1 = GANDetector(gan_config)
        detector2 = GANDetector(gan_config)
        
        assert detector1.model_hash == detector2.model_hash
    
    def test_detect_single_image(self, gan_config, sample_face_frame):
        """Test detection on single image."""
        detector = GANDetector(gan_config)
        
        result = detector.detect(sample_face_frame)
        
        assert isinstance(result, GANDetectionResult)
        assert 0.0 <= result.overall_score <= 1.0
        assert result.decision in ["authentic", "synthetic", "suspicious", "uncertain"]
    
    def test_detect_with_face_region(
        self,
        gan_config,
        sample_face_frame,
        sample_face_region
    ):
        """Test detection with face region."""
        detector = GANDetector(gan_config)
        
        result = detector.detect(sample_face_frame, face_region=sample_face_region)
        
        assert isinstance(result, GANDetectionResult)
    
    def test_detect_sequence(self, gan_config, sample_frame_sequence):
        """Test detection on frame sequence."""
        detector = GANDetector(gan_config)
        
        result = detector.detect_sequence(sample_frame_sequence)
        
        assert isinstance(result, GANDetectionResult)
        assert result.frames_analyzed == len(sample_frame_sequence)
    
    def test_detect_sequence_with_regions(
        self,
        gan_config,
        sample_frame_sequence,
        sample_face_regions
    ):
        """Test sequence detection with face regions."""
        detector = GANDetector(gan_config)
        
        result = detector.detect_sequence(
            frames=sample_frame_sequence,
            face_regions=sample_face_regions,
        )
        
        assert isinstance(result, GANDetectionResult)
    
    def test_detect_empty_frames(self, gan_config):
        """Test detection fails for empty frames."""
        detector = GANDetector(gan_config)
        
        result = detector.detect_sequence([])
        
        assert result.decision == "uncertain"
        assert GANReasonCodes.INVALID_INPUT.value in result.reason_codes
    
    def test_detect_small_image(self, gan_config, small_frame):
        """Test detection fails for small image."""
        detector = GANDetector(gan_config)
        
        result = detector.detect(small_frame)
        
        assert GANReasonCodes.IMAGE_TOO_SMALL.value in result.reason_codes
    
    def test_detect_large_image(self, gan_config, large_frame):
        """Test detection fails for large image."""
        detector = GANDetector(gan_config)
        
        result = detector.detect(large_frame)
        
        assert GANReasonCodes.IMAGE_TOO_LARGE.value in result.reason_codes
    
    def test_detect_with_details(self, gan_config, sample_face_frame):
        """Test detection with detailed results."""
        detector = GANDetector(gan_config)
        
        result = detector.detect(sample_face_frame, include_details=True)
        
        assert result.discriminator_result is not None
        assert result.deepfake_result is not None
        assert result.artifact_result is not None
    
    def test_component_scores(self, gan_config, sample_face_frame):
        """Test all component scores are computed."""
        detector = GANDetector(gan_config)
        
        result = detector.detect(sample_face_frame)
        
        assert 0.0 <= result.discriminator_score <= 1.0
        assert 0.0 <= result.deepfake_score <= 1.0
        assert 0.0 <= result.artifact_score <= 1.0
    
    def test_veid_penalty_computed(self, gan_config, sample_face_frame):
        """Test VEID penalty is computed."""
        detector = GANDetector(gan_config)
        
        result = detector.detect(sample_face_frame)
        
        assert result.veid_penalty >= 0
        assert result.veid_adjusted_score >= 0
        assert result.veid_adjusted_score <= 10000
    
    def test_result_hash_computed(self, gan_config, sample_face_frame):
        """Test result hash is computed."""
        detector = GANDetector(gan_config)
        
        result = detector.detect(sample_face_frame)
        
        assert result.result_hash is not None
        assert len(result.result_hash) == 32
    
    def test_result_hash_determinism(self, gan_config, sample_face_frame):
        """Test result hash is deterministic."""
        detector = GANDetector(gan_config)
        
        result1 = detector.detect(sample_face_frame)
        result2 = detector.detect(sample_face_frame)
        
        assert result1.result_hash == result2.result_hash
    
    def test_processing_time_recorded(self, gan_config, sample_face_frame):
        """Test processing time is recorded."""
        detector = GANDetector(gan_config)
        
        result = detector.detect(sample_face_frame)
        
        assert result.processing_time_ms > 0
    
    def test_reason_codes_included(self, gan_config, sample_face_frame):
        """Test reason codes are included."""
        detector = GANDetector(gan_config)
        
        result = detector.detect(sample_face_frame)
        
        assert isinstance(result.reason_codes, list)
        assert len(result.reason_codes) > 0
    
    def test_detect_synthetic_image(self, gan_config, synthetic_frame):
        """Test detection of synthetic image."""
        detector = GANDetector(gan_config)
        
        result = detector.detect(synthetic_frame)
        
        assert isinstance(result, GANDetectionResult)
        # Synthetic frame may have higher scores
    
    def test_detect_deepfake_image(self, gan_config, deepfake_frame):
        """Test detection of deepfake image."""
        detector = GANDetector(gan_config)
        
        result = detector.detect(deepfake_frame)
        
        assert isinstance(result, GANDetectionResult)
    
    def test_sequence_truncation(self, gan_config, sample_face_frame):
        """Test long sequences are truncated."""
        detector = GANDetector(gan_config)
        
        # Create very long sequence
        frames = [sample_face_frame.copy() for _ in range(200)]
        
        result = detector.detect_sequence(frames)
        
        # Should be truncated to max_frames_per_sequence
        assert result.frames_analyzed <= gan_config.max_frames_per_sequence
    
    def test_detect_single_frame_method(self, gan_config, sample_face_frame):
        """Test quick single frame detection method."""
        detector = GANDetector(gan_config)
        
        score, reason_codes = detector.detect_single_frame(sample_face_frame)
        
        assert 0.0 <= score <= 1.0
        assert isinstance(reason_codes, list)


class TestGANDetectionResult:
    """Tests for GANDetectionResult dataclass."""
    
    def test_result_to_dict(self):
        """Test converting result to dict."""
        result = GANDetectionResult(
            is_synthetic=True,
            decision="synthetic",
            overall_score=0.85,
            confidence=0.9,
            discriminator_score=0.8,
            deepfake_score=0.7,
            artifact_score=0.6,
            detected_type=SyntheticImageType.GAN_GENERATED,
            veid_penalty=5000,
            veid_adjusted_score=5000,
            model_version="1.0.0",
            model_hash="abc123",
            reason_codes=["GAN_DETECTED", "SYNTHETIC_IMAGE_DETECTED"],
            result_hash="def456",
            processing_time_ms=150.0,
            frames_analyzed=10,
        )
        
        d = result.to_dict()
        
        assert d["is_synthetic"] is True
        assert d["decision"] == "synthetic"
        assert d["overall_score"] == 0.85
        assert d["detected_type"] == "gan_generated"
        assert d["veid_penalty"] == 5000
        assert len(d["reason_codes"]) == 2
    
    def test_result_to_veid_record(self):
        """Test converting result to VEID record format."""
        result = GANDetectionResult(
            is_synthetic=True,
            decision="synthetic",
            overall_score=0.85,
            confidence=0.9,
            detected_type=SyntheticImageType.GAN_GENERATED,
            veid_penalty=5000,
            veid_adjusted_score=5000,
            model_version="1.0.0",
            model_hash="abc123def456ghi789",
            reason_codes=["GAN_DETECTED"],
            result_hash="def456",
        )
        
        record = result.to_veid_record()
        
        # Score is inverted: higher = more authentic
        assert record["gan_score"] == 1500  # (1 - 0.85) * 10000
        assert record["is_synthetic"] is True
        assert record["confidence"] == 90
        assert record["veid_penalty"] == 5000
        assert len(record["model_hash"]) <= 16


class TestVEIDIntegration:
    """Tests for VEID score integration."""
    
    def test_authentic_image_no_penalty(self, gan_config, high_quality_frame):
        """Test authentic image has no VEID penalty."""
        detector = GANDetector(gan_config)
        
        result = detector.detect(high_quality_frame)
        
        if not result.is_synthetic:
            assert result.veid_penalty == 0
            assert result.veid_adjusted_score == 10000
    
    def test_veid_record_format(self, gan_config, sample_face_frame):
        """Test VEID record has correct format."""
        detector = GANDetector(gan_config)
        
        result = detector.detect(sample_face_frame)
        record = result.to_veid_record()
        
        # Check required fields
        assert "gan_score" in record
        assert "is_synthetic" in record
        assert "decision" in record
        assert "confidence" in record
        assert "veid_penalty" in record
        assert "veid_adjusted_score" in record
        assert "model_version" in record
        assert "model_hash" in record
        assert "result_hash" in record
        assert "reason_codes" in record
        
        # Check value ranges
        assert 0 <= record["gan_score"] <= 10000
        assert 0 <= record["confidence"] <= 100
        assert len(record["reason_codes"]) <= 5


class TestDetectionModes:
    """Tests for different detection modes."""
    
    def test_fast_mode(self, sample_face_frame):
        """Test fast detection mode."""
        config = GANDetectionConfig.fast_mode()
        detector = GANDetector(config)
        
        result = detector.detect(sample_face_frame)
        
        assert isinstance(result, GANDetectionResult)
    
    def test_accurate_mode(self, sample_face_frame):
        """Test accurate detection mode."""
        config = GANDetectionConfig.accurate_mode()
        detector = GANDetector(config)
        
        result = detector.detect(sample_face_frame)
        
        assert isinstance(result, GANDetectionResult)
    
    def test_full_mode(self, sample_face_frame):
        """Test full detection mode."""
        config = GANDetectionConfig()
        assert config.mode == DetectionMode.FULL
        
        detector = GANDetector(config)
        result = detector.detect(sample_face_frame)
        
        assert isinstance(result, GANDetectionResult)


class TestDeterminism:
    """Tests for deterministic behavior."""
    
    def test_deterministic_detection(self, gan_config, sample_face_frame):
        """Test detection is deterministic."""
        detector = GANDetector(gan_config)
        
        result1 = detector.detect(sample_face_frame)
        result2 = detector.detect(sample_face_frame)
        
        assert result1.overall_score == result2.overall_score
        assert result1.discriminator_score == result2.discriminator_score
        assert result1.deepfake_score == result2.deepfake_score
        assert result1.artifact_score == result2.artifact_score
    
    def test_deterministic_across_instances(self, gan_config, sample_face_frame):
        """Test detection is deterministic across instances."""
        detector1 = GANDetector(gan_config)
        detector2 = GANDetector(gan_config)
        
        result1 = detector1.detect(sample_face_frame)
        result2 = detector2.detect(sample_face_frame)
        
        assert result1.overall_score == result2.overall_score


class TestCreateDetector:
    """Tests for factory function."""
    
    def test_create_without_config(self):
        """Test creating detector without config."""
        detector = create_detector()
        
        assert isinstance(detector, GANDetector)
    
    def test_create_with_config(self, gan_config):
        """Test creating detector with config."""
        detector = create_detector(gan_config)
        
        assert isinstance(detector, GANDetector)
        assert detector.config == gan_config
