"""
Reason codes for GAN fraud detection outcomes.

VE-923: GAN fraud detection - reason codes for audit trail

These codes provide detailed explanations for GAN detection decisions
and are included in the verification record for audit purposes.
"""

from enum import Enum
from typing import List, Optional, Dict, Any
from dataclasses import dataclass


class GANReasonCodes(str, Enum):
    """Standard reason codes for GAN detection outcomes."""
    
    # Success codes - Image verified as authentic
    IMAGE_AUTHENTIC = "IMAGE_AUTHENTIC"
    HIGH_CONFIDENCE_REAL = "HIGH_CONFIDENCE_REAL"
    ALL_CHECKS_PASSED = "ALL_CHECKS_PASSED"
    
    # GAN detection codes
    GAN_DETECTED = "GAN_DETECTED"
    GAN_HIGH_CONFIDENCE = "GAN_HIGH_CONFIDENCE"
    GAN_LOW_CONFIDENCE = "GAN_LOW_CONFIDENCE"
    SYNTHETIC_IMAGE_DETECTED = "SYNTHETIC_IMAGE_DETECTED"
    AI_GENERATED_DETECTED = "AI_GENERATED_DETECTED"
    
    # Deepfake detection codes
    DEEPFAKE_DETECTED = "DEEPFAKE_DETECTED"
    FACESWAP_DETECTED = "FACESWAP_DETECTED"
    EXPRESSION_MANIPULATION_DETECTED = "EXPRESSION_MANIPULATION_DETECTED"
    MORPHED_FACE_DETECTED = "MORPHED_FACE_DETECTED"
    
    # Artifact detection codes
    FREQUENCY_ANOMALY = "FREQUENCY_ANOMALY"
    CHECKERBOARD_ARTIFACT = "CHECKERBOARD_ARTIFACT"
    BLENDING_BOUNDARY_DETECTED = "BLENDING_BOUNDARY_DETECTED"
    TEXTURE_INCONSISTENCY = "TEXTURE_INCONSISTENCY"
    COLOR_MISMATCH = "COLOR_MISMATCH"
    COMPRESSION_ANOMALY = "COMPRESSION_ANOMALY"
    UPSAMPLING_ARTIFACT = "UPSAMPLING_ARTIFACT"
    
    # Facial analysis codes
    EYE_REFLECTION_ANOMALY = "EYE_REFLECTION_ANOMALY"
    EYE_ASYMMETRY_DETECTED = "EYE_ASYMMETRY_DETECTED"
    UNNATURAL_SKIN_TEXTURE = "UNNATURAL_SKIN_TEXTURE"
    HAIR_BOUNDARY_ANOMALY = "HAIR_BOUNDARY_ANOMALY"
    BACKGROUND_INCONSISTENCY = "BACKGROUND_INCONSISTENCY"
    
    # Temporal analysis codes (video)
    TEMPORAL_INCONSISTENCY = "TEMPORAL_INCONSISTENCY"
    UNNATURAL_BLINK_PATTERN = "UNNATURAL_BLINK_PATTERN"
    EXPRESSION_DISCONTINUITY = "EXPRESSION_DISCONTINUITY"
    LIP_SYNC_MISMATCH = "LIP_SYNC_MISMATCH"
    
    # Quality issues
    IMAGE_TOO_SMALL = "IMAGE_TOO_SMALL"
    IMAGE_TOO_LARGE = "IMAGE_TOO_LARGE"
    LOW_QUALITY_INPUT = "LOW_QUALITY_INPUT"
    FACE_NOT_DETECTED = "FACE_NOT_DETECTED"
    MULTIPLE_FACES = "MULTIPLE_FACES"
    INSUFFICIENT_FRAMES = "INSUFFICIENT_FRAMES"
    
    # System issues
    PROCESSING_ERROR = "PROCESSING_ERROR"
    MODEL_ERROR = "MODEL_ERROR"
    TIMEOUT_ERROR = "TIMEOUT_ERROR"
    INVALID_INPUT = "INVALID_INPUT"
    
    @classmethod
    def get_category(cls, code: "GANReasonCodes") -> str:
        """Get the category of a reason code."""
        success_codes = {
            cls.IMAGE_AUTHENTIC, cls.HIGH_CONFIDENCE_REAL,
            cls.ALL_CHECKS_PASSED
        }
        gan_codes = {
            cls.GAN_DETECTED, cls.GAN_HIGH_CONFIDENCE, cls.GAN_LOW_CONFIDENCE,
            cls.SYNTHETIC_IMAGE_DETECTED, cls.AI_GENERATED_DETECTED
        }
        deepfake_codes = {
            cls.DEEPFAKE_DETECTED, cls.FACESWAP_DETECTED,
            cls.EXPRESSION_MANIPULATION_DETECTED, cls.MORPHED_FACE_DETECTED
        }
        artifact_codes = {
            cls.FREQUENCY_ANOMALY, cls.CHECKERBOARD_ARTIFACT,
            cls.BLENDING_BOUNDARY_DETECTED, cls.TEXTURE_INCONSISTENCY,
            cls.COLOR_MISMATCH, cls.COMPRESSION_ANOMALY, cls.UPSAMPLING_ARTIFACT
        }
        facial_codes = {
            cls.EYE_REFLECTION_ANOMALY, cls.EYE_ASYMMETRY_DETECTED,
            cls.UNNATURAL_SKIN_TEXTURE, cls.HAIR_BOUNDARY_ANOMALY,
            cls.BACKGROUND_INCONSISTENCY
        }
        temporal_codes = {
            cls.TEMPORAL_INCONSISTENCY, cls.UNNATURAL_BLINK_PATTERN,
            cls.EXPRESSION_DISCONTINUITY, cls.LIP_SYNC_MISMATCH
        }
        quality_codes = {
            cls.IMAGE_TOO_SMALL, cls.IMAGE_TOO_LARGE, cls.LOW_QUALITY_INPUT,
            cls.FACE_NOT_DETECTED, cls.MULTIPLE_FACES, cls.INSUFFICIENT_FRAMES
        }
        system_codes = {
            cls.PROCESSING_ERROR, cls.MODEL_ERROR, cls.TIMEOUT_ERROR,
            cls.INVALID_INPUT
        }
        
        if code in success_codes:
            return "success"
        elif code in gan_codes:
            return "gan_detection"
        elif code in deepfake_codes:
            return "deepfake_detection"
        elif code in artifact_codes:
            return "artifact_detection"
        elif code in facial_codes:
            return "facial_analysis"
        elif code in temporal_codes:
            return "temporal_analysis"
        elif code in quality_codes:
            return "quality"
        elif code in system_codes:
            return "system"
        return "unknown"
    
    @classmethod
    def is_rejection_code(cls, code: "GANReasonCodes") -> bool:
        """Check if a code indicates rejection of the image."""
        rejection_codes = {
            cls.GAN_DETECTED, cls.GAN_HIGH_CONFIDENCE,
            cls.SYNTHETIC_IMAGE_DETECTED, cls.AI_GENERATED_DETECTED,
            cls.DEEPFAKE_DETECTED, cls.FACESWAP_DETECTED,
            cls.EXPRESSION_MANIPULATION_DETECTED, cls.MORPHED_FACE_DETECTED
        }
        return code in rejection_codes
    
    @classmethod
    def get_severity(cls, code: "GANReasonCodes") -> int:
        """
        Get severity level for a reason code.
        
        Returns:
            0: Informational
            1: Warning
            2: Error (soft rejection)
            3: Critical (hard rejection)
        """
        critical_codes = {
            cls.GAN_HIGH_CONFIDENCE, cls.DEEPFAKE_DETECTED,
            cls.FACESWAP_DETECTED
        }
        error_codes = {
            cls.GAN_DETECTED, cls.SYNTHETIC_IMAGE_DETECTED,
            cls.AI_GENERATED_DETECTED, cls.EXPRESSION_MANIPULATION_DETECTED,
            cls.MORPHED_FACE_DETECTED
        }
        warning_codes = {
            cls.GAN_LOW_CONFIDENCE, cls.FREQUENCY_ANOMALY,
            cls.CHECKERBOARD_ARTIFACT, cls.BLENDING_BOUNDARY_DETECTED,
            cls.TEXTURE_INCONSISTENCY, cls.COLOR_MISMATCH,
            cls.EYE_REFLECTION_ANOMALY, cls.TEMPORAL_INCONSISTENCY
        }
        
        if code in critical_codes:
            return 3
        elif code in error_codes:
            return 2
        elif code in warning_codes:
            return 1
        return 0


@dataclass
class ReasonCodeDetails:
    """Detailed information about a reason code."""
    
    code: GANReasonCodes
    description: str
    confidence: float
    severity: int
    details: Dict[str, Any]
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "code": self.code.value,
            "description": self.description,
            "confidence": self.confidence,
            "severity": self.severity,
            "details": self.details,
        }


# Standard descriptions for each reason code
REASON_CODE_DESCRIPTIONS: Dict[GANReasonCodes, str] = {
    GANReasonCodes.IMAGE_AUTHENTIC: "Image verified as authentic with high confidence",
    GANReasonCodes.HIGH_CONFIDENCE_REAL: "All detection methods indicate authentic image",
    GANReasonCodes.ALL_CHECKS_PASSED: "All GAN and deepfake detection checks passed",
    
    GANReasonCodes.GAN_DETECTED: "Image appears to be GAN-generated",
    GANReasonCodes.GAN_HIGH_CONFIDENCE: "High confidence GAN-generated image detected",
    GANReasonCodes.GAN_LOW_CONFIDENCE: "Possible GAN generation detected with low confidence",
    GANReasonCodes.SYNTHETIC_IMAGE_DETECTED: "Image appears to be synthetically generated",
    GANReasonCodes.AI_GENERATED_DETECTED: "Image shows signs of AI generation",
    
    GANReasonCodes.DEEPFAKE_DETECTED: "Deepfake manipulation detected",
    GANReasonCodes.FACESWAP_DETECTED: "Face swap manipulation detected",
    GANReasonCodes.EXPRESSION_MANIPULATION_DETECTED: "Facial expression manipulation detected",
    GANReasonCodes.MORPHED_FACE_DETECTED: "Morphed face detected",
    
    GANReasonCodes.FREQUENCY_ANOMALY: "Unusual frequency patterns detected in image",
    GANReasonCodes.CHECKERBOARD_ARTIFACT: "GAN checkerboard artifact pattern detected",
    GANReasonCodes.BLENDING_BOUNDARY_DETECTED: "Blending boundary detected around face",
    GANReasonCodes.TEXTURE_INCONSISTENCY: "Inconsistent texture patterns in image",
    GANReasonCodes.COLOR_MISMATCH: "Color inconsistency detected between regions",
    GANReasonCodes.COMPRESSION_ANOMALY: "Unusual compression artifacts detected",
    GANReasonCodes.UPSAMPLING_ARTIFACT: "Upsampling artifacts detected",
    
    GANReasonCodes.EYE_REFLECTION_ANOMALY: "Inconsistent eye reflections detected",
    GANReasonCodes.EYE_ASYMMETRY_DETECTED: "Unusual eye asymmetry detected",
    GANReasonCodes.UNNATURAL_SKIN_TEXTURE: "Unnatural skin texture patterns",
    GANReasonCodes.HAIR_BOUNDARY_ANOMALY: "Unusual hair boundary artifacts",
    GANReasonCodes.BACKGROUND_INCONSISTENCY: "Background inconsistency detected",
    
    GANReasonCodes.TEMPORAL_INCONSISTENCY: "Temporal inconsistency in video sequence",
    GANReasonCodes.UNNATURAL_BLINK_PATTERN: "Unnatural blinking pattern detected",
    GANReasonCodes.EXPRESSION_DISCONTINUITY: "Expression discontinuity in video",
    GANReasonCodes.LIP_SYNC_MISMATCH: "Lip sync mismatch detected",
    
    GANReasonCodes.IMAGE_TOO_SMALL: "Input image resolution too low",
    GANReasonCodes.IMAGE_TOO_LARGE: "Input image resolution exceeds limit",
    GANReasonCodes.LOW_QUALITY_INPUT: "Low quality input image",
    GANReasonCodes.FACE_NOT_DETECTED: "No face detected in image",
    GANReasonCodes.MULTIPLE_FACES: "Multiple faces detected where one expected",
    GANReasonCodes.INSUFFICIENT_FRAMES: "Insufficient frames for temporal analysis",
    
    GANReasonCodes.PROCESSING_ERROR: "Error during processing",
    GANReasonCodes.MODEL_ERROR: "Model inference error",
    GANReasonCodes.TIMEOUT_ERROR: "Processing timeout exceeded",
    GANReasonCodes.INVALID_INPUT: "Invalid input format",
}


def aggregate_reason_codes(
    codes: List[ReasonCodeDetails],
    max_codes: int = 5
) -> List[str]:
    """
    Aggregate and prioritize reason codes for storage.
    
    Args:
        codes: List of detailed reason codes.
        max_codes: Maximum number of codes to return.
    
    Returns:
        List of reason code strings, prioritized by severity.
    """
    # Sort by severity (descending), then by confidence (descending)
    sorted_codes = sorted(
        codes,
        key=lambda x: (x.severity, x.confidence),
        reverse=True
    )
    
    # Return code values
    return [c.code.value for c in sorted_codes[:max_codes]]
