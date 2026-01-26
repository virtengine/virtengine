"""
Reason codes for liveness detection outcomes.

VE-901: Liveness detection - anti-spoofing reason codes

These codes are used to provide detailed explanations for liveness
decisions and are included in the verification record for audit purposes.
"""

from enum import Enum
from typing import List, Optional
from dataclasses import dataclass


class LivenessReasonCodes(str, Enum):
    """Standard reason codes for liveness detection outcomes."""
    
    # Success codes
    LIVENESS_CONFIRMED = "LIVENESS_CONFIRMED"
    HIGH_CONFIDENCE_LIVE = "HIGH_CONFIDENCE_LIVE"
    ALL_CHALLENGES_PASSED = "ALL_CHALLENGES_PASSED"
    
    # Active challenge failures
    BLINK_NOT_DETECTED = "BLINK_NOT_DETECTED"
    BLINK_TOO_FAST = "BLINK_TOO_FAST"
    BLINK_TOO_SLOW = "BLINK_TOO_SLOW"
    SMILE_NOT_DETECTED = "SMILE_NOT_DETECTED"
    SMILE_INSUFFICIENT = "SMILE_INSUFFICIENT"
    HEAD_TURN_NOT_DETECTED = "HEAD_TURN_NOT_DETECTED"
    HEAD_TURN_WRONG_DIRECTION = "HEAD_TURN_WRONG_DIRECTION"
    HEAD_TURN_TOO_EXTREME = "HEAD_TURN_TOO_EXTREME"
    HEAD_NOD_NOT_DETECTED = "HEAD_NOD_NOT_DETECTED"
    EYEBROW_RAISE_NOT_DETECTED = "EYEBROW_RAISE_NOT_DETECTED"
    CHALLENGE_TIMEOUT = "CHALLENGE_TIMEOUT"
    CHALLENGE_INCOMPLETE = "CHALLENGE_INCOMPLETE"
    
    # Passive analysis failures
    UNNATURAL_TEXTURE = "UNNATURAL_TEXTURE"
    FLAT_DEPTH_DETECTED = "FLAT_DEPTH_DETECTED"
    NO_MOTION_DETECTED = "NO_MOTION_DETECTED"
    UNNATURAL_MOTION = "UNNATURAL_MOTION"
    REFLECTION_ANOMALY = "REFLECTION_ANOMALY"
    MOIRE_PATTERN_DETECTED = "MOIRE_PATTERN_DETECTED"
    FREQUENCY_ANOMALY = "FREQUENCY_ANOMALY"
    
    # Spoof detection codes
    PHOTO_PRINT_DETECTED = "PHOTO_PRINT_DETECTED"
    SCREEN_DISPLAY_DETECTED = "SCREEN_DISPLAY_DETECTED"
    VIDEO_REPLAY_DETECTED = "VIDEO_REPLAY_DETECTED"
    MASK_2D_DETECTED = "MASK_2D_DETECTED"
    MASK_3D_DETECTED = "MASK_3D_DETECTED"
    DEEPFAKE_DETECTED = "DEEPFAKE_DETECTED"
    SPOOF_HIGH_CONFIDENCE = "SPOOF_HIGH_CONFIDENCE"
    
    # Quality issues
    FACE_TOO_SMALL = "FACE_TOO_SMALL"
    FACE_TOO_LARGE = "FACE_TOO_LARGE"
    FACE_NOT_CENTERED = "FACE_NOT_CENTERED"
    FACE_OCCLUDED = "FACE_OCCLUDED"
    POOR_LIGHTING = "POOR_LIGHTING"
    MOTION_BLUR = "MOTION_BLUR"
    INSUFFICIENT_FRAMES = "INSUFFICIENT_FRAMES"
    INCONSISTENT_FACE = "INCONSISTENT_FACE"
    
    # System issues
    PROCESSING_ERROR = "PROCESSING_ERROR"
    MODEL_ERROR = "MODEL_ERROR"
    TIMEOUT_ERROR = "TIMEOUT_ERROR"
    INVALID_INPUT = "INVALID_INPUT"
    
    @classmethod
    def get_category(cls, code: "LivenessReasonCodes") -> str:
        """Get the category of a reason code."""
        success_codes = {
            cls.LIVENESS_CONFIRMED, cls.HIGH_CONFIDENCE_LIVE,
            cls.ALL_CHALLENGES_PASSED
        }
        challenge_codes = {
            cls.BLINK_NOT_DETECTED, cls.BLINK_TOO_FAST, cls.BLINK_TOO_SLOW,
            cls.SMILE_NOT_DETECTED, cls.SMILE_INSUFFICIENT,
            cls.HEAD_TURN_NOT_DETECTED, cls.HEAD_TURN_WRONG_DIRECTION,
            cls.HEAD_TURN_TOO_EXTREME, cls.HEAD_NOD_NOT_DETECTED,
            cls.EYEBROW_RAISE_NOT_DETECTED, cls.CHALLENGE_TIMEOUT,
            cls.CHALLENGE_INCOMPLETE
        }
        passiVIRTENGINE_codes = {
            cls.UNNATURAL_TEXTURE, cls.FLAT_DEPTH_DETECTED,
            cls.NO_MOTION_DETECTED, cls.UNNATURAL_MOTION,
            cls.REFLECTION_ANOMALY, cls.MOIRE_PATTERN_DETECTED,
            cls.FREQUENCY_ANOMALY
        }
        spoof_codes = {
            cls.PHOTO_PRINT_DETECTED, cls.SCREEN_DISPLAY_DETECTED,
            cls.VIDEO_REPLAY_DETECTED, cls.MASK_2D_DETECTED,
            cls.MASK_3D_DETECTED, cls.DEEPFAKE_DETECTED,
            cls.SPOOF_HIGH_CONFIDENCE
        }
        quality_codes = {
            cls.FACE_TOO_SMALL, cls.FACE_TOO_LARGE, cls.FACE_NOT_CENTERED,
            cls.FACE_OCCLUDED, cls.POOR_LIGHTING, cls.MOTION_BLUR,
            cls.INSUFFICIENT_FRAMES, cls.INCONSISTENT_FACE
        }
        system_codes = {
            cls.PROCESSING_ERROR, cls.MODEL_ERROR,
            cls.TIMEOUT_ERROR, cls.INVALID_INPUT
        }
        
        if code in success_codes:
            return "success"
        elif code in challenge_codes:
            return "challenge"
        elif code in passiVIRTENGINE_codes:
            return "passive"
        elif code in spoof_codes:
            return "spoof"
        elif code in quality_codes:
            return "quality"
        elif code in system_codes:
            return "system"
        else:
            return "unknown"
    
    @classmethod
    def is_fatal(cls, code: "LivenessReasonCodes") -> bool:
        """Check if this reason code indicates a fatal failure."""
        fatal_codes = {
            cls.PHOTO_PRINT_DETECTED, cls.SCREEN_DISPLAY_DETECTED,
            cls.VIDEO_REPLAY_DETECTED, cls.MASK_2D_DETECTED,
            cls.MASK_3D_DETECTED, cls.DEEPFAKE_DETECTED,
            cls.SPOOF_HIGH_CONFIDENCE, cls.INVALID_INPUT,
            cls.PROCESSING_ERROR, cls.MODEL_ERROR
        }
        return code in fatal_codes
    
    @classmethod
    def get_user_message(cls, code: "LivenessReasonCodes") -> str:
        """Get a user-friendly message for a reason code."""
        messages = {
            cls.LIVENESS_CONFIRMED: "Liveness verified successfully.",
            cls.HIGH_CONFIDENCE_LIVE: "Liveness verified with high confidence.",
            cls.ALL_CHALLENGES_PASSED: "All liveness challenges completed.",
            cls.BLINK_NOT_DETECTED: "Please blink naturally.",
            cls.BLINK_TOO_FAST: "Blink was too quick, please try again.",
            cls.BLINK_TOO_SLOW: "Blink was too slow, please try again.",
            cls.SMILE_NOT_DETECTED: "Please smile for the camera.",
            cls.SMILE_INSUFFICIENT: "Please smile more naturally.",
            cls.HEAD_TURN_NOT_DETECTED: "Please turn your head as directed.",
            cls.HEAD_TURN_WRONG_DIRECTION: "Please turn your head in the other direction.",
            cls.HEAD_TURN_TOO_EXTREME: "Please turn your head less.",
            cls.HEAD_NOD_NOT_DETECTED: "Please nod your head.",
            cls.EYEBROW_RAISE_NOT_DETECTED: "Please raise your eyebrows.",
            cls.CHALLENGE_TIMEOUT: "Challenge timed out. Please try again.",
            cls.CHALLENGE_INCOMPLETE: "Challenge not completed. Please try again.",
            cls.UNNATURAL_TEXTURE: "Image quality issue detected.",
            cls.FLAT_DEPTH_DETECTED: "Please ensure 3D capture.",
            cls.NO_MOTION_DETECTED: "Please ensure natural movement.",
            cls.UNNATURAL_MOTION: "Unusual motion pattern detected.",
            cls.REFLECTION_ANOMALY: "Reflection issue detected.",
            cls.MOIRE_PATTERN_DETECTED: "Screen pattern detected.",
            cls.FREQUENCY_ANOMALY: "Image anomaly detected.",
            cls.PHOTO_PRINT_DETECTED: "Printed photo detected. Please use live camera.",
            cls.SCREEN_DISPLAY_DETECTED: "Screen display detected. Please use live camera.",
            cls.VIDEO_REPLAY_DETECTED: "Video replay detected. Please use live camera.",
            cls.MASK_2D_DETECTED: "Mask detected. Please show your real face.",
            cls.MASK_3D_DETECTED: "3D mask detected. Please show your real face.",
            cls.DEEPFAKE_DETECTED: "Synthetic media detected.",
            cls.SPOOF_HIGH_CONFIDENCE: "Spoofing attempt detected.",
            cls.FACE_TOO_SMALL: "Please move closer to the camera.",
            cls.FACE_TOO_LARGE: "Please move further from the camera.",
            cls.FACE_NOT_CENTERED: "Please center your face in the frame.",
            cls.FACE_OCCLUDED: "Please ensure your face is fully visible.",
            cls.POOR_LIGHTING: "Please improve lighting conditions.",
            cls.MOTION_BLUR: "Please hold still to reduce blur.",
            cls.INSUFFICIENT_FRAMES: "Recording too short. Please try again.",
            cls.INCONSISTENT_FACE: "Face detection inconsistent. Please try again.",
            cls.PROCESSING_ERROR: "Processing error. Please try again.",
            cls.MODEL_ERROR: "System error. Please try again.",
            cls.TIMEOUT_ERROR: "Request timed out. Please try again.",
            cls.INVALID_INPUT: "Invalid input. Please try again.",
        }
        return messages.get(code, "Unknown liveness issue.")


@dataclass
class ReasonCodeDetails:
    """Detailed information about a reason code occurrence."""
    
    code: LivenessReasonCodes
    frame_index: Optional[int] = None
    confidence: float = 0.0
    details: Optional[str] = None
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "code": self.code.value,
            "frame_index": self.frame_index,
            "confidence": self.confidence,
            "details": self.details,
            "category": LivenessReasonCodes.get_category(self.code),
            "is_fatal": LivenessReasonCodes.is_fatal(self.code),
            "user_message": LivenessReasonCodes.get_user_message(self.code),
        }


def aggregate_reason_codes(codes: List[ReasonCodeDetails]) -> List[str]:
    """Aggregate reason codes, removing duplicates and keeping highest confidence."""
    code_map = {}
    for detail in codes:
        key = detail.code.value
        if key not in code_map or detail.confidence > code_map[key].confidence:
            code_map[key] = detail
    
    # Sort by confidence (highest first)
    sorted_codes = sorted(code_map.values(), key=lambda x: -x.confidence)
    return [c.code.value for c in sorted_codes]
