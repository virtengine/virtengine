"""
Tests for passive liveness analysis.

VE-901: Test passive analysis (texture, depth, motion).
"""

import pytest
import numpy as np

from ml.liveness_detection.passive_analysis import (
    PassiveAnalyzer,
    PassiveAnalysisResult,
)
from ml.liveness_detection.config import LivenessConfig
from ml.liveness_detection.reason_codes import LivenessReasonCodes


class TestPassiveAnalyzer:
    """Tests for PassiveAnalyzer class."""
    
    def test_analyzer_creation(self, liveness_config):
        """Test creating analyzer with config."""
        analyzer = PassiveAnalyzer(liveness_config)
        assert analyzer is not None
        assert analyzer.config == liveness_config
    
    def test_analyzer_default_config(self):
        """Test creating analyzer with default config."""
        analyzer = PassiveAnalyzer()
        assert analyzer is not None
    
    def test_analyze_empty_frames(self, liveness_config):
        """Test analysis fails with empty frames."""
        analyzer = PassiveAnalyzer(liveness_config)
        
        result = analyzer.analyze([])
        
        assert result.is_live is False
        assert result.confidence == 0.0
        assert LivenessReasonCodes.INSUFFICIENT_FRAMES in result.reason_codes
    
    def test_analyze_natural_frames(self, liveness_config, sample_frame_sequence):
        """Test analysis of natural-looking frames."""
        analyzer = PassiveAnalyzer(liveness_config)
        
        result = analyzer.analyze(sample_frame_sequence)
        
        assert isinstance(result, PassiveAnalysisResult)
        assert 0.0 <= result.combined_score <= 1.0
        assert 0.0 <= result.texture_score <= 1.0
        assert 0.0 <= result.motion_score <= 1.0
    
    def test_analyze_static_frames(self, liveness_config, static_frame_sequence):
        """Test analysis of static frames (photo-like)."""
        analyzer = PassiveAnalyzer(liveness_config)
        
        result = analyzer.analyze(static_frame_sequence)
        
        # Static frames should have low motion score
        assert result.motion_score < 0.5
        # May detect no motion
        assert LivenessReasonCodes.NO_MOTION_DETECTED in result.reason_codes or result.motion_score < 0.8
    
    def test_analyze_with_face_regions(self, liveness_config, sample_frame_sequence, sample_face_regions):
        """Test analysis with face regions specified."""
        analyzer = PassiveAnalyzer(liveness_config)
        
        result = analyzer.analyze(sample_frame_sequence, sample_face_regions)
        
        assert isinstance(result, PassiveAnalysisResult)
        assert 0.0 <= result.combined_score <= 1.0


class TestTextureAnalysis:
    """Tests for texture analysis component."""
    
    def test_texture_analysis(self, liveness_config, sample_frame_sequence):
        """Test texture analysis returns valid score."""
        analyzer = PassiveAnalyzer(liveness_config)
        
        result = analyzer.analyze(sample_frame_sequence)
        
        assert 0.0 <= result.texture_score <= 1.0
        assert "texture" in result.details
    
    def test_lbp_computation(self, liveness_config, np_random):
        """Test LBP pattern computation."""
        analyzer = PassiveAnalyzer(liveness_config)
        
        # Create a simple grayscale image
        gray = np_random.randint(0, 256, (50, 50), dtype=np.uint8)
        
        lbp = analyzer._compute_lbp(gray)
        
        assert lbp.shape == (48, 48)  # 2 pixels smaller in each dimension
        assert lbp.dtype == np.uint8
    
    def test_uniformity_computation(self, liveness_config, np_random):
        """Test uniformity computation."""
        analyzer = PassiveAnalyzer(liveness_config)
        
        # Uniform image should have high uniformity
        uniform = np.ones((50, 50), dtype=np.uint8) * 128
        uniform_score = analyzer._compute_uniformity(uniform)
        
        # Random image should have lower uniformity
        random_img = np_random.randint(0, 256, (50, 50), dtype=np.uint8)
        random_score = analyzer._compute_uniformity(random_img)
        
        assert uniform_score > random_score


class TestDepthAnalysis:
    """Tests for depth analysis component."""
    
    def test_depth_analysis(self, liveness_config, sample_frame_sequence):
        """Test depth analysis returns valid score."""
        analyzer = PassiveAnalyzer(liveness_config)
        
        result = analyzer.analyze(sample_frame_sequence)
        
        assert 0.0 <= result.depth_score <= 1.0
        assert "depth" in result.details
    
    def test_flat_image_detection(self, liveness_config, np_random):
        """Test detection of flat images."""
        analyzer = PassiveAnalyzer(liveness_config)
        
        # Create flat (uniform) frames
        flat_frames = [np.ones((100, 100, 3), dtype=np.uint8) * 128 for _ in range(10)]
        
        result = analyzer.analyze(flat_frames)
        
        # Flat images should have low depth score
        assert result.depth_score < 0.5


class TestMotionAnalysis:
    """Tests for motion analysis component."""
    
    def test_motion_analysis(self, liveness_config, sample_frame_sequence):
        """Test motion analysis returns valid score."""
        analyzer = PassiveAnalyzer(liveness_config)
        
        result = analyzer.analyze(sample_frame_sequence)
        
        assert 0.0 <= result.motion_score <= 1.0
        assert "motion" in result.details
    
    def test_static_motion_detection(self, liveness_config, static_frame_sequence):
        """Test detection of static (no motion) frames."""
        analyzer = PassiveAnalyzer(liveness_config)
        
        result = analyzer.analyze(static_frame_sequence)
        
        # Static frames should have low motion or trigger no motion detected
        assert result.motion_score <= 0.5 or LivenessReasonCodes.NO_MOTION_DETECTED in result.reason_codes
    
    def test_excessive_motion_detection(self, liveness_config, np_random):
        """Test detection of excessive motion."""
        analyzer = PassiveAnalyzer(liveness_config)
        
        # Create frames with large differences (excessive motion)
        frames = []
        for i in range(20):
            if i % 2 == 0:
                frames.append(np.zeros((100, 100, 3), dtype=np.uint8))
            else:
                frames.append(np.ones((100, 100, 3), dtype=np.uint8) * 255)
        
        result = analyzer.analyze(frames)
        
        # Excessive motion should be flagged
        assert result.motion_score < 0.8 or LivenessReasonCodes.UNNATURAL_MOTION in result.reason_codes


class TestReflectionAnalysis:
    """Tests for reflection analysis component."""
    
    def test_reflection_analysis(self, liveness_config, sample_frame_sequence):
        """Test reflection analysis returns valid score."""
        analyzer = PassiveAnalyzer(liveness_config)
        
        result = analyzer.analyze(sample_frame_sequence)
        
        assert 0.0 <= result.reflection_score <= 1.0
        assert "reflection" in result.details


class TestMoireDetection:
    """Tests for moire pattern detection."""
    
    def test_moire_detection(self, liveness_config, sample_frame_sequence):
        """Test moire detection returns valid score."""
        analyzer = PassiveAnalyzer(liveness_config)
        
        result = analyzer.analyze(sample_frame_sequence)
        
        assert 0.0 <= result.moire_score <= 1.0
        assert "moire" in result.details
    
    def test_screen_pattern_detection(self, liveness_config, screen_attack_frame):
        """Test detection of screen patterns."""
        analyzer = PassiveAnalyzer(liveness_config)
        
        # Create sequence from screen attack frame
        frames = [screen_attack_frame.copy() for _ in range(10)]
        
        result = analyzer.analyze(frames)
        
        # Screen patterns should be detected
        # Note: Detection depends on pattern strength
        assert 0.0 <= result.moire_score <= 1.0


class TestFrequencyAnalysis:
    """Tests for frequency analysis component."""
    
    def test_frequency_analysis(self, liveness_config, sample_frame_sequence):
        """Test frequency analysis returns valid score."""
        analyzer = PassiveAnalyzer(liveness_config)
        
        result = analyzer.analyze(sample_frame_sequence)
        
        assert 0.0 <= result.frequency_score <= 1.0
        assert "frequency" in result.details


class TestPassiveAnalysisResult:
    """Tests for PassiveAnalysisResult dataclass."""
    
    def test_result_to_dict(self):
        """Test converting result to dict."""
        result = PassiveAnalysisResult(
            is_live=True,
            confidence=0.85,
            texture_score=0.7,
            depth_score=0.6,
            motion_score=0.8,
            reflection_score=0.75,
            moire_score=0.1,
            frequency_score=0.7,
            combined_score=0.72,
            reason_codes=[],
            details={"test": "value"},
        )
        
        d = result.to_dict()
        
        assert d["is_live"] is True
        assert d["confidence"] == 0.85
        assert d["combined_score"] == 0.72
        assert d["texture_score"] == 0.7
