"""
VirtEngine GAN Fraud Detection Module

VE-923: GAN fraud detection for identity fraud prevention

This module provides GAN-based synthetic image detection for the VEID
identity verification system. It implements:
- CNN discriminator for real vs synthetic classification
- Deepfake detection (face swap, expression manipulation)
- Artifact analysis (frequency, checkerboard, blending patterns)
- Integration with VEID scoring

Usage:
    from ml.gan_detection import GANDetector, GANDetectionConfig
    
    config = GANDetectionConfig()
    detector = GANDetector(config)
    result = detector.detect(image, face_region)
    
    # Check if synthetic
    if result.is_synthetic:
        print(f"Synthetic image detected: {result.decision}")
        print(f"Detection type: {result.detected_type}")
    
    # Get VEID scoring record
    veid_record = result.to_veid_record()
"""

from ml.gan_detection.config import (
    GANDetectionConfig,
    DiscriminatorConfig,
    DeepfakeConfig,
    ArtifactAnalysisConfig,
    VEIDIntegrationConfig,
    DetectionMode,
    SyntheticImageType,
    ArtifactType,
)
from ml.gan_detection.detector import (
    GANDetector,
    GANDetectionResult,
    create_detector,
)
from ml.gan_detection.discriminator import (
    CNNDiscriminator,
    DiscriminatorResult,
    DiscriminatorFeatures,
    create_discriminator,
)
from ml.gan_detection.deepfake_detection import (
    DeepfakeDetector,
    DeepfakeResult,
    FaceSwapResult,
    ExpressionManipulationResult,
    create_deepfake_detector,
)
from ml.gan_detection.artifact_analysis import (
    ArtifactAnalyzer,
    ArtifactResult,
    create_artifact_analyzer,
)
from ml.gan_detection.reason_codes import (
    GANReasonCodes,
    ReasonCodeDetails,
    REASON_CODE_DESCRIPTIONS,
    aggregate_reason_codes,
)

__version__ = "1.0.0"
__all__ = [
    # Config
    "GANDetectionConfig",
    "DiscriminatorConfig",
    "DeepfakeConfig",
    "ArtifactAnalysisConfig",
    "VEIDIntegrationConfig",
    "DetectionMode",
    "SyntheticImageType",
    "ArtifactType",
    # Main detector
    "GANDetector",
    "GANDetectionResult",
    "create_detector",
    # Discriminator
    "CNNDiscriminator",
    "DiscriminatorResult",
    "DiscriminatorFeatures",
    "create_discriminator",
    # Deepfake detection
    "DeepfakeDetector",
    "DeepfakeResult",
    "FaceSwapResult",
    "ExpressionManipulationResult",
    "create_deepfake_detector",
    # Artifact analysis
    "ArtifactAnalyzer",
    "ArtifactResult",
    "create_artifact_analyzer",
    # Reason codes
    "GANReasonCodes",
    "ReasonCodeDetails",
    "REASON_CODE_DESCRIPTIONS",
    "aggregate_reason_codes",
]
