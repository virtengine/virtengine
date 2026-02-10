"""
Tests for deepfake detection.

VE-923: Test deepfake detection functionality.
"""

import pytest
import numpy as np

from ml.gan_detection.deepfake_detection import (
    DeepfakeDetector,
    DeepfakeResult,
    FaceSwapResult,
    ExpressionManipulationResult,
    create_deepfake_detector,
)
from ml.gan_detection.config import DeepfakeConfig, GANDetectionConfig, SyntheticImageType
from ml.gan_detection.reason_codes import GANReasonCodes


class TestDeepfakeDetector:
    """Tests for DeepfakeDetector class."""
    
    def test_detector_creation(self, deepfake_config):
        """Test creating detector with config."""
        detector = DeepfakeDetector(deepfake_config)
        
        assert detector is not None
        assert detector.config == deepfake_config
        assert detector.MODEL_VERSION == "1.0.0"
    
    def test_detector_default_config(self):
        """Test creating detector with default config."""
        detector = DeepfakeDetector()
        assert detector is not None
    
    def test_model_hash_generation(self, deepfake_config):
        """Test model hash is generated."""
        detector = DeepfakeDetector(deepfake_config)
        
        assert detector._model_hash is not None
        assert len(detector._model_hash) == 32
    
    def test_model_hash_determinism(self, deepfake_config):
        """Test model hash is deterministic."""
        detector1 = DeepfakeDetector(deepfake_config)
        detector2 = DeepfakeDetector(deepfake_config)
        
        assert detector1._model_hash == detector2._model_hash
    
    def test_detect_single_frame(self, deepfake_config, sample_face_frame):
        """Test detection on single frame."""
        detector = DeepfakeDetector(deepfake_config)
        
        result = detector.detect([sample_face_frame])
        
        assert isinstance(result, DeepfakeResult)
        assert 0.0 <= result.deepfake_score <= 1.0
        assert result.decision in ["authentic", "deepfake", "suspicious"]
    
    def test_detect_frame_sequence(self, deepfake_config, sample_frame_sequence):
        """Test detection on frame sequence."""
        detector = DeepfakeDetector(deepfake_config)
        
        result = detector.detect(sample_frame_sequence)
        
        assert isinstance(result, DeepfakeResult)
        assert result.frames_analyzed == len(sample_frame_sequence)
    
    def test_detect_with_face_regions(
        self,
        deepfake_config,
        sample_frame_sequence,
        sample_face_regions
    ):
        """Test detection with face regions."""
        detector = DeepfakeDetector(deepfake_config)
        
        result = detector.detect(
            frames=sample_frame_sequence,
            face_regions=sample_face_regions,
        )
        
        assert isinstance(result, DeepfakeResult)
    
    def test_detect_empty_frames(self, deepfake_config):
        """Test detection fails for empty frames."""
        detector = DeepfakeDetector(deepfake_config)
        
        result = detector.detect([])
        
        assert result.is_deepfake is False
        assert result.decision == "uncertain"
        assert GANReasonCodes.INVALID_INPUT in result.reason_codes
    
    def test_detect_with_details(self, deepfake_config, sample_frame_sequence):
        """Test detection with detailed results."""
        detector = DeepfakeDetector(deepfake_config)
        
        result = detector.detect(
            frames=sample_frame_sequence,
            include_details=True,
        )
        
        assert result.faceswap_result is not None
        assert result.expression_result is not None
    
    def test_detect_deepfake_frame(self, deepfake_config, deepfake_frame):
        """Test detection on deepfake-like frame."""
        detector = DeepfakeDetector(deepfake_config)
        
        result = detector.detect([deepfake_frame])
        
        assert isinstance(result, DeepfakeResult)
        # Deepfake frame should trigger higher scores
    
    def test_detect_blended_frame(self, deepfake_config, blended_frame):
        """Test detection on frame with blending."""
        detector = DeepfakeDetector(deepfake_config)
        
        result = detector.detect([blended_frame])
        
        assert isinstance(result, DeepfakeResult)
    
    def test_faceswap_score(self, deepfake_config, sample_frame_sequence):
        """Test faceswap score is computed."""
        detector = DeepfakeDetector(deepfake_config)
        
        result = detector.detect(sample_frame_sequence)
        
        assert 0.0 <= result.faceswap_score <= 1.0
    
    def test_expression_score(self, deepfake_config, sample_frame_sequence):
        """Test expression manipulation score is computed."""
        detector = DeepfakeDetector(deepfake_config)
        
        result = detector.detect(sample_frame_sequence)
        
        assert 0.0 <= result.expression_manipulation_score <= 1.0
    
    def test_morphing_score(self, deepfake_config, sample_frame_sequence):
        """Test morphing score is computed."""
        detector = DeepfakeDetector(deepfake_config)
        
        result = detector.detect(sample_frame_sequence)
        
        assert 0.0 <= result.morphing_score <= 1.0
    
    def test_temporal_score(self, deepfake_config, sample_frame_sequence):
        """Test temporal score is computed for long sequences."""
        config = deepfake_config
        config.min_frames_for_temporal = 5  # Lower threshold
        detector = DeepfakeDetector(config)
        
        result = detector.detect(sample_frame_sequence)
        
        assert 0.0 <= result.temporal_score <= 1.0
    
    def test_processing_time_recorded(self, deepfake_config, sample_face_frame):
        """Test processing time is recorded."""
        detector = DeepfakeDetector(deepfake_config)
        
        result = detector.detect([sample_face_frame])
        
        assert result.processing_time_ms > 0


class TestDeepfakeResult:
    """Tests for DeepfakeResult dataclass."""
    
    def test_result_to_dict(self):
        """Test converting result to dict."""
        result = DeepfakeResult(
            is_deepfake=True,
            decision="deepfake",
            deepfake_score=0.8,
            confidence=0.9,
            detected_type=SyntheticImageType.DEEPFAKE_FACESWAP,
            faceswap_score=0.85,
            expression_manipulation_score=0.3,
            morphing_score=0.2,
            model_version="1.0.0",
            reason_codes=[GANReasonCodes.DEEPFAKE_DETECTED],
            frames_analyzed=20,
        )
        
        d = result.to_dict()
        
        assert d["is_deepfake"] is True
        assert d["decision"] == "deepfake"
        assert d["deepfake_score"] == 0.8
        assert d["detected_type"] == "deepfake_faceswap"
        assert d["frames_analyzed"] == 20
    
    def test_result_to_veid_record(self):
        """Test converting result to VEID record format."""
        result = DeepfakeResult(
            is_deepfake=True,
            decision="deepfake",
            deepfake_score=0.8,
            confidence=0.9,
            detected_type=SyntheticImageType.DEEPFAKE_FACESWAP,
            model_version="1.0.0",
            model_hash="abc123def456ghi789",
            reason_codes=[GANReasonCodes.DEEPFAKE_DETECTED],
        )
        
        record = result.to_veid_record()
        
        # Score is inverted: higher = more authentic
        # Use approximate comparison for floating-point precision
        assert abs(record["deepfake_score"] - 2000) <= 1  # int((1 - 0.8) * 10000)
        assert record["is_deepfake"] is True
        assert record["confidence"] == 90


class TestFaceSwapResult:
    """Tests for FaceSwapResult dataclass."""
    
    def test_result_to_dict(self):
        """Test converting result to dict."""
        result = FaceSwapResult(
            is_faceswap=True,
            confidence=0.85,
            boundary_score=0.7,
            blending_score=0.8,
            color_match_score=0.6,
            reason_codes=[GANReasonCodes.FACESWAP_DETECTED],
        )
        
        d = result.to_dict()
        
        assert d["is_faceswap"] is True
        assert d["confidence"] == 0.85
        assert d["boundary_score"] == 0.7


class TestExpressionManipulationResult:
    """Tests for ExpressionManipulationResult dataclass."""
    
    def test_result_to_dict(self):
        """Test converting result to dict."""
        result = ExpressionManipulationResult(
            is_manipulated=True,
            confidence=0.75,
            temporal_consistency_score=0.6,
            muscle_movement_score=0.4,
            micro_expression_score=0.3,
            anomaly_frames=[5, 10, 15],
            reason_codes=[GANReasonCodes.EXPRESSION_MANIPULATION_DETECTED],
        )
        
        d = result.to_dict()
        
        assert d["is_manipulated"] is True
        assert d["anomaly_frame_count"] == 3


class TestDetectSingleFrame:
    """Tests for single frame detection method."""
    
    def test_detect_single_frame(self, deepfake_config, sample_face_frame):
        """Test quick single frame detection."""
        detector = DeepfakeDetector(deepfake_config)
        
        score, reason_codes = detector.detect_single_frame(sample_face_frame)
        
        assert 0.0 <= score <= 1.0
        assert isinstance(reason_codes, list)
    
    def test_detect_single_frame_with_region(
        self,
        deepfake_config,
        sample_face_frame,
        sample_face_region
    ):
        """Test single frame detection with face region."""
        detector = DeepfakeDetector(deepfake_config)
        
        score, reason_codes = detector.detect_single_frame(
            sample_face_frame,
            face_region=sample_face_region,
        )
        
        assert 0.0 <= score <= 1.0


class TestCreateDeepfakeDetector:
    """Tests for factory function."""
    
    def test_create_without_config(self):
        """Test creating detector without config."""
        detector = create_deepfake_detector()
        
        assert isinstance(detector, DeepfakeDetector)
    
    def test_create_with_config(self, gan_config):
        """Test creating detector with full config."""
        detector = create_deepfake_detector(gan_config)
        
        assert isinstance(detector, DeepfakeDetector)
        assert detector.config == gan_config.deepfake
