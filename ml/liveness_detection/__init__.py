"""
VirtEngine Liveness Detection Module

This module provides anti-spoofing liveness detection for the VEID identity
verification system. It implements both active and passive liveness detection
to prevent photo, video, screen, and mask attacks.

VE-901: Liveness detection - anti-spoofing implementation

Features:
- Active liveness with challenges (blink, smile, head turn)
- Passive liveness analyzing texture/depth/motion
- Detection of photos/screens/masks
- Liveness score integration into VEID scoring

Usage:
    from ml.liveness_detection import LivenessDetector, LivenessConfig
    
    config = LivenessConfig()
    detector = LivenessDetector(config)
    result = detector.detect(frame_sequence, challenges)
"""

from ml.liveness_detection.config import (
    LivenessConfig,
    ActiveChallengeConfig,
    PassiveAnalysisConfig,
    SpoofDetectionConfig,
    ChallengeType,
)
from ml.liveness_detection.detector import (
    LivenessDetector,
    LivenessResult,
)
from ml.liveness_detection.actiVIRTENGINE_challenges import (
    ActiveChallengeDetector,
    ChallengeResult,
)
from ml.liveness_detection.passiVIRTENGINE_analysis import (
    PassiveAnalyzer,
    PassiveAnalysisResult,
)
from ml.liveness_detection.spoof_detection import (
    SpoofDetector,
    SpoofDetectionResult,
    SpoofType,
)
from ml.liveness_detection.reason_codes import LivenessReasonCodes

__version__ = "1.0.0"
__all__ = [
    "LivenessConfig",
    "ActiveChallengeConfig",
    "PassiveAnalysisConfig",
    "SpoofDetectionConfig",
    "ChallengeType",
    "LivenessDetector",
    "LivenessResult",
    "ActiveChallengeDetector",
    "ChallengeResult",
    "PassiveAnalyzer",
    "PassiveAnalysisResult",
    "SpoofDetector",
    "SpoofDetectionResult",
    "SpoofType",
    "LivenessReasonCodes",
]
