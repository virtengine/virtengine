"""
Reason codes for verification failures and borderline cases.

These codes are used to provide detailed explanations for verification
decisions and are included in the verification record for audit purposes.
"""

from enum import Enum
from typing import List, Optional
from dataclasses import dataclass


class ReasonCodes(str, Enum):
    """Standard reason codes for verification outcomes."""
    
    # Success codes
    MATCH_CONFIRMED = "MATCH_CONFIRMED"
    HIGH_CONFIDENCE_MATCH = "HIGH_CONFIDENCE_MATCH"
    
    # Face detection issues
    NO_FACE_DETECTED = "NO_FACE_DETECTED"
    MULTIPLE_FACES = "MULTIPLE_FACES"
    FACE_TOO_SMALL = "FACE_TOO_SMALL"
    FACE_PARTIALLY_VISIBLE = "FACE_PARTIALLY_VISIBLE"
    FACE_OCCLUDED = "FACE_OCCLUDED"
    
    # Image quality issues
    LOW_QUALITY_IMAGE = "LOW_QUALITY_IMAGE"
    IMAGE_TOO_DARK = "IMAGE_TOO_DARK"
    IMAGE_TOO_BRIGHT = "IMAGE_TOO_BRIGHT"
    IMAGE_BLURRY = "IMAGE_BLURRY"
    LOW_RESOLUTION = "LOW_RESOLUTION"
    INVALID_IMAGE_FORMAT = "INVALID_IMAGE_FORMAT"
    CORRUPT_IMAGE_DATA = "CORRUPT_IMAGE_DATA"
    
    # Verification issues
    BORDERLINE_MATCH = "BORDERLINE_MATCH"
    LOW_SIMILARITY_SCORE = "LOW_SIMILARITY_SCORE"
    EMBEDDING_MISMATCH = "EMBEDDING_MISMATCH"
    POSE_MISMATCH = "POSE_MISMATCH"
    
    # Model/system issues
    EMBEDDING_ERROR = "EMBEDDING_ERROR"
    MODEL_MISMATCH = "MODEL_MISMATCH"
    MODEL_HASH_MISMATCH = "MODEL_HASH_MISMATCH"
    PREPROCESSING_ERROR = "PREPROCESSING_ERROR"
    DETECTION_ERROR = "DETECTION_ERROR"
    ALIGNMENT_ERROR = "ALIGNMENT_ERROR"
    
    # Determinism issues
    DETERMINISM_VIOLATION = "DETERMINISM_VIOLATION"
    HASH_MISMATCH = "HASH_MISMATCH"
    SEED_NOT_SET = "SEED_NOT_SET"
    
    # Configuration issues
    INVALID_CONFIGURATION = "INVALID_CONFIGURATION"
    THRESHOLD_NOT_MET = "THRESHOLD_NOT_MET"
    
    # Timeout/resource issues
    PROCESSING_TIMEOUT = "PROCESSING_TIMEOUT"
    RESOURCE_EXHAUSTED = "RESOURCE_EXHAUSTED"
    
    @classmethod
    def get_category(cls, code: "ReasonCodes") -> str:
        """Get the category of a reason code."""
        detection_codes = {
            cls.NO_FACE_DETECTED, cls.MULTIPLE_FACES, cls.FACE_TOO_SMALL,
            cls.FACE_PARTIALLY_VISIBLE, cls.FACE_OCCLUDED
        }
        quality_codes = {
            cls.LOW_QUALITY_IMAGE, cls.IMAGE_TOO_DARK, cls.IMAGE_TOO_BRIGHT,
            cls.IMAGE_BLURRY, cls.LOW_RESOLUTION, cls.INVALID_IMAGE_FORMAT,
            cls.CORRUPT_IMAGE_DATA
        }
        verification_codes = {
            cls.BORDERLINE_MATCH, cls.LOW_SIMILARITY_SCORE, 
            cls.EMBEDDING_MISMATCH, cls.POSE_MISMATCH
        }
        model_codes = {
            cls.EMBEDDING_ERROR, cls.MODEL_MISMATCH, cls.MODEL_HASH_MISMATCH,
            cls.PREPROCESSING_ERROR, cls.DETECTION_ERROR, cls.ALIGNMENT_ERROR
        }
        determinism_codes = {
            cls.DETERMINISM_VIOLATION, cls.HASH_MISMATCH, cls.SEED_NOT_SET
        }
        success_codes = {cls.MATCH_CONFIRMED, cls.HIGH_CONFIDENCE_MATCH}
        
        if code in detection_codes:
            return "detection"
        elif code in quality_codes:
            return "quality"
        elif code in verification_codes:
            return "verification"
        elif code in model_codes:
            return "model"
        elif code in determinism_codes:
            return "determinism"
        elif code in success_codes:
            return "success"
        else:
            return "other"
    
    @classmethod
    def is_retriable(cls, code: "ReasonCodes") -> bool:
        """Check if a failure with this code should be retried."""
        non_retriable = {
            cls.INVALID_IMAGE_FORMAT, cls.CORRUPT_IMAGE_DATA,
            cls.MODEL_MISMATCH, cls.MODEL_HASH_MISMATCH,
            cls.INVALID_CONFIGURATION
        }
        return code not in non_retriable
    
    @classmethod
    def get_user_message(cls, code: "ReasonCodes") -> str:
        """Get a user-friendly message for a reason code."""
        messages = {
            cls.MATCH_CONFIRMED: "Identity verified successfully.",
            cls.HIGH_CONFIDENCE_MATCH: "High-confidence identity match.",
            cls.NO_FACE_DETECTED: "No face was detected in the image. Please ensure your face is clearly visible.",
            cls.MULTIPLE_FACES: "Multiple faces detected. Please ensure only one face is in the frame.",
            cls.FACE_TOO_SMALL: "Face is too small. Please move closer to the camera.",
            cls.FACE_PARTIALLY_VISIBLE: "Face is not fully visible. Please center your face in the frame.",
            cls.FACE_OCCLUDED: "Face appears to be covered. Please remove any obstructions.",
            cls.LOW_QUALITY_IMAGE: "Image quality is too low. Please retake the photo in better conditions.",
            cls.IMAGE_TOO_DARK: "Image is too dark. Please improve the lighting.",
            cls.IMAGE_TOO_BRIGHT: "Image is too bright. Please reduce the lighting or avoid direct light.",
            cls.IMAGE_BLURRY: "Image is blurry. Please hold the camera steady.",
            cls.LOW_RESOLUTION: "Image resolution is too low. Please use a higher quality camera.",
            cls.INVALID_IMAGE_FORMAT: "Invalid image format. Please use JPEG or PNG.",
            cls.CORRUPT_IMAGE_DATA: "Image data is corrupted. Please retake the photo.",
            cls.BORDERLINE_MATCH: "Verification result is inconclusive. Additional verification may be required.",
            cls.LOW_SIMILARITY_SCORE: "Face does not match the reference. Please try again or contact support.",
            cls.EMBEDDING_MISMATCH: "Unable to verify identity. Please try again.",
            cls.POSE_MISMATCH: "Face angle is different from the reference. Please face the camera directly.",
            cls.EMBEDDING_ERROR: "Technical error during processing. Please try again.",
            cls.MODEL_MISMATCH: "System configuration error. Please contact support.",
            cls.MODEL_HASH_MISMATCH: "System verification failed. Please contact support.",
            cls.PREPROCESSING_ERROR: "Error processing image. Please try again.",
            cls.DETECTION_ERROR: "Error detecting face. Please try again.",
            cls.ALIGNMENT_ERROR: "Error aligning face. Please try again.",
            cls.DETERMINISM_VIOLATION: "System verification failed. Please contact support.",
            cls.HASH_MISMATCH: "Verification consistency check failed. Please try again.",
            cls.SEED_NOT_SET: "System configuration error. Please contact support.",
            cls.INVALID_CONFIGURATION: "System configuration error. Please contact support.",
            cls.THRESHOLD_NOT_MET: "Verification threshold not met. Please try again.",
            cls.PROCESSING_TIMEOUT: "Processing took too long. Please try again.",
            cls.RESOURCE_EXHAUSTED: "System is busy. Please try again later.",
        }
        return messages.get(code, "An unknown error occurred. Please try again.")


@dataclass
class ReasonCodeDetail:
    """Detailed information about a reason code occurrence."""
    
    code: ReasonCodes
    message: str
    severity: str  # "info", "warning", "error"
    context: Optional[dict] = None
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "code": self.code.value,
            "message": self.message,
            "severity": self.severity,
            "context": self.context or {},
        }


def collect_reason_codes(
    codes: List[ReasonCodes],
    include_messages: bool = True
) -> List[dict]:
    """Collect reason codes with their details."""
    result = []
    for code in codes:
        entry = {"code": code.value}
        if include_messages:
            entry["message"] = ReasonCodes.get_user_message(code)
            entry["category"] = ReasonCodes.get_category(code)
            entry["retriable"] = ReasonCodes.is_retriable(code)
        result.append(entry)
    return result
