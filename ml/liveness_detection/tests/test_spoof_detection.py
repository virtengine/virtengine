"""
Tests for spoof detection.

VE-901: Test spoof detection (photo, screen, mask, deepfake).
"""

import pytest
import numpy as np

from ml.liveness_detection.spoof_detection import (
    SpoofDetector,
    SpoofDetectionResult,
    SpoofType,
)
from ml.liveness_detection.config import LivenessConfig
from ml.liveness_detection.reason_codes import LivenessReasonCodes


class TestSpoofDetector:
    """Tests for SpoofDetector class."""
    
    def test_detector_creation(self, liveness_config):
        """Test creating detector with config."""
        detector = SpoofDetector(liveness_config)
        assert detector is not None
        assert detector.config == liveness_config
    
    def test_detector_default_config(self):
        """Test creating detector with default config."""
        detector = SpoofDetector()
        assert detector is not None
    
    def test_detect_empty_frames(self, liveness_config):
        """Test detection fails with empty frames."""
        detector = SpoofDetector(liveness_config)
        
        result = detector.detect([])
        
        assert result.is_spoof is False
        assert LivenessReasonCodes.INSUFFICIENT_FRAMES in result.reason_codes
    
    def test_detect_natural_frames(self, liveness_config, sample_frame_sequence):
        """Test detection on natural-looking frames."""
        detector = SpoofDetector(liveness_config)
        
        result = detector.detect(sample_frame_sequence)
        
        assert isinstance(result, SpoofDetectionResult)
        assert 0.0 <= result.overall_spoof_score <= 1.0
        # Natural frames should generally not be detected as spoof
        assert result.confidence > 0.0


class TestPhotoPrintDetection:
    """Tests for photo print detection."""
    
    def test_photo_print_detection(self, liveness_config, photo_attack_frame):
        """Test detection of printed photo attacks."""
        detector = SpoofDetector(liveness_config)
        
        frames = [photo_attack_frame.copy() for _ in range(10)]
        result = detector.detect(frames)
        
        assert 0.0 <= result.photo_print_score <= 1.0
        # Photo attack should have some detection
        assert "photo_print" in result.details
    
    def test_print_texture_analysis(self, liveness_config, np_random):
        """Test print texture analysis."""
        detector = SpoofDetector(liveness_config)
        
        # Create print-like texture (high-frequency patterns)
        face = np.zeros((100, 100, 3), dtype=np.uint8)
        face[:, :] = 150
        # Add periodic noise (like print dots)
        for i in range(0, 100, 2):
            for j in range(0, 100, 2):
                face[i, j] = 155
        
        score = detector._analyze_print_texture(face)
        assert 0.0 <= score <= 1.0
    
    def test_saturation_analysis(self, liveness_config):
        """Test color saturation analysis."""
        detector = SpoofDetector(liveness_config)
        
        # High saturation image
        high_sat = np.zeros((100, 100, 3), dtype=np.uint8)
        high_sat[:, :, 0] = 50   # Low blue
        high_sat[:, :, 1] = 100  # Medium green
        high_sat[:, :, 2] = 250  # High red
        
        high_score = detector._analyze_saturation(high_sat)
        
        # Low saturation image (almost grayscale)
        low_sat = np.zeros((100, 100, 3), dtype=np.uint8)
        low_sat[:, :] = 128
        
        low_score = detector._analyze_saturation(low_sat)
        
        assert high_score > low_score


class TestScreenDisplayDetection:
    """Tests for screen display detection."""
    
    def test_screen_display_detection(self, liveness_config, screen_attack_frame):
        """Test detection of screen display attacks."""
        detector = SpoofDetector(liveness_config)
        
        frames = [screen_attack_frame.copy() for _ in range(10)]
        result = detector.detect(frames)
        
        assert 0.0 <= result.screen_display_score <= 1.0
        assert "screen_display" in result.details
    
    def test_moire_pattern_detection(self, liveness_config, screen_attack_frame):
        """Test moire pattern detection."""
        detector = SpoofDetector(liveness_config)
        
        score = detector._detect_moire_pattern(screen_attack_frame)
        
        assert 0.0 <= score <= 1.0
    
    def test_color_banding_detection(self, liveness_config, np_random):
        """Test color banding detection."""
        detector = SpoofDetector(liveness_config)
        
        # Create image with color banding
        banded = np.zeros((100, 100, 3), dtype=np.uint8)
        for i in range(0, 100, 10):
            color = (i * 2 % 256, i * 2 % 256, i * 2 % 256)
            banded[i:i+10, :] = color
        
        score = detector._detect_color_banding(banded)
        
        assert 0.0 <= score <= 1.0


class TestVideoReplayDetection:
    """Tests for video replay detection."""
    
    def test_video_replay_detection(self, liveness_config, sample_frame_sequence):
        """Test video replay detection."""
        detector = SpoofDetector(liveness_config)
        
        result = detector.detect(sample_frame_sequence)
        
        assert 0.0 <= result.video_replay_score <= 1.0
        assert "video_replay" in result.details
    
    def test_temporal_consistency_analysis(self, liveness_config, sample_frame_sequence):
        """Test temporal consistency analysis."""
        detector = SpoofDetector(liveness_config)
        
        consistency = detector._analyze_temporal_consistency(sample_frame_sequence, None)
        
        assert 0.0 <= consistency <= 1.0
    
    def test_compression_artifacts_analysis(self, liveness_config, sample_frame_sequence):
        """Test compression artifacts analysis."""
        detector = SpoofDetector(liveness_config)
        
        score = detector._analyze_compression_artifacts(sample_frame_sequence, None)
        
        assert 0.0 <= score <= 1.0


class TestMask2DDetection:
    """Tests for 2D mask detection."""
    
    def test_2d_mask_detection(self, liveness_config, sample_frame_sequence):
        """Test 2D mask detection."""
        detector = SpoofDetector(liveness_config)
        
        result = detector.detect(sample_frame_sequence)
        
        assert 0.0 <= result.mask_2d_score <= 1.0
        assert "mask_2d" in result.details
    
    def test_face_boundary_analysis(self, liveness_config, sample_face_frame):
        """Test face boundary analysis."""
        detector = SpoofDetector(liveness_config)
        
        score = detector._analyze_face_boundary(sample_face_frame)
        
        assert 0.0 <= score <= 1.0
    
    def test_skin_texture_analysis(self, liveness_config, sample_face_frame):
        """Test skin texture analysis."""
        detector = SpoofDetector(liveness_config)
        
        score = detector._analyze_skin_texture(sample_face_frame)
        
        assert 0.0 <= score <= 1.0


class TestMask3DDetection:
    """Tests for 3D mask detection."""
    
    def test_3d_mask_detection(self, liveness_config, sample_frame_sequence):
        """Test 3D mask detection."""
        detector = SpoofDetector(liveness_config)
        
        result = detector.detect(sample_frame_sequence)
        
        assert 0.0 <= result.mask_3d_score <= 1.0
        assert "mask_3d" in result.details
    
    def test_depth_uniformity_analysis(self, liveness_config, sample_face_frame):
        """Test depth uniformity analysis."""
        detector = SpoofDetector(liveness_config)
        
        score = detector._analyze_depth_uniformity(sample_face_frame)
        
        assert 0.0 <= score <= 1.0
    
    def test_material_properties_analysis(self, liveness_config, sample_face_frame):
        """Test material properties analysis."""
        detector = SpoofDetector(liveness_config)
        
        score = detector._analyze_material_properties(sample_face_frame)
        
        assert 0.0 <= score <= 1.0


class TestDeepfakeDetection:
    """Tests for deepfake detection."""
    
    def test_deepfake_detection(self, liveness_config, sample_frame_sequence):
        """Test deepfake detection."""
        detector = SpoofDetector(liveness_config)
        
        result = detector.detect(sample_frame_sequence)
        
        assert 0.0 <= result.deepfake_score <= 1.0
        assert "deepfake" in result.details
    
    def test_deepfake_boundary_analysis(self, liveness_config, sample_face_frame):
        """Test deepfake boundary analysis."""
        detector = SpoofDetector(liveness_config)
        
        score = detector._analyze_deepfake_boundary(sample_face_frame)
        
        assert 0.0 <= score <= 1.0
    
    def test_temporal_coherence_analysis(self, liveness_config, sample_frame_sequence):
        """Test temporal coherence analysis."""
        detector = SpoofDetector(liveness_config)
        
        score = detector._analyze_temporal_coherence(sample_frame_sequence, None)
        
        assert 0.0 <= score <= 1.0


class TestSpoofDetectionResult:
    """Tests for SpoofDetectionResult dataclass."""
    
    def test_result_to_dict(self):
        """Test converting result to dict."""
        result = SpoofDetectionResult(
            is_spoof=True,
            confidence=0.85,
            spoof_type=SpoofType.PHOTO_PRINT,
            photo_print_score=0.8,
            screen_display_score=0.2,
            video_replay_score=0.1,
            mask_2d_score=0.15,
            mask_3d_score=0.1,
            deepfake_score=0.05,
            overall_spoof_score=0.8,
            reason_codes=[LivenessReasonCodes.PHOTO_PRINT_DETECTED],
            details={"test": "value"},
        )
        
        d = result.to_dict()
        
        assert d["is_spoof"] is True
        assert d["confidence"] == 0.85
        assert d["spoof_type"] == "photo_print"
        assert d["photo_print_score"] == 0.8
    
    def test_result_no_spoof_type_when_not_detected(self):
        """Test spoof type is None when not detected."""
        result = SpoofDetectionResult(
            is_spoof=False,
            confidence=0.9,
            spoof_type=None,
            overall_spoof_score=0.1,
        )
        
        d = result.to_dict()
        
        assert d["spoof_type"] is None


class TestSpoofType:
    """Tests for SpoofType enum."""
    
    def test_all_spoof_types(self):
        """Test all spoof types exist."""
        assert SpoofType.PHOTO_PRINT.value == "photo_print"
        assert SpoofType.PHOTO_SCREEN.value == "photo_screen"
        assert SpoofType.VIDEO_REPLAY.value == "video_replay"
        assert SpoofType.MASK_2D.value == "mask_2d"
        assert SpoofType.MASK_3D.value == "mask_3d"
        assert SpoofType.DEEPFAKE.value == "deepfake"
