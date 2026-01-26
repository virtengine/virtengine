"""
Reason codes for autoencoder anomaly detection outcomes.

VE-924: Autoencoder anomaly detection - reason codes for audit trail

These codes provide detailed explanations for anomaly detection decisions
and are included in the verification record for audit purposes.
"""

from enum import Enum
from typing import List, Optional, Dict, Any
from dataclasses import dataclass


class AnomalyReasonCodes(str, Enum):
    """Standard reason codes for anomaly detection outcomes."""
    
    # Success codes - No anomaly detected
    NO_ANOMALY_DETECTED = "NO_ANOMALY_DETECTED"
    RECONSTRUCTION_NORMAL = "RECONSTRUCTION_NORMAL"
    LATENT_NORMAL = "LATENT_NORMAL"
    ALL_CHECKS_PASSED = "ALL_CHECKS_PASSED"
    
    # Reconstruction error codes
    HIGH_RECONSTRUCTION_ERROR = "HIGH_RECONSTRUCTION_ERROR"
    MSE_ABOVIRTENGINE_THRESHOLD = "MSE_ABOVIRTENGINE_THRESHOLD"
    MAE_ABOVIRTENGINE_THRESHOLD = "MAE_ABOVIRTENGINE_THRESHOLD"
    SSIM_BELOW_THRESHOLD = "SSIM_BELOW_THRESHOLD"
    PERCEPTUAL_LOSS_HIGH = "PERCEPTUAL_LOSS_HIGH"
    
    # Patch-level anomalies
    PATCH_ANOMALY_DETECTED = "PATCH_ANOMALY_DETECTED"
    LOCALIZED_RECONSTRUCTION_FAILURE = "LOCALIZED_RECONSTRUCTION_FAILURE"
    MULTIPLE_PATCH_ANOMALIES = "MULTIPLE_PATCH_ANOMALIES"
    
    # Channel-specific anomalies
    RED_CHANNEL_ANOMALY = "RED_CHANNEL_ANOMALY"
    GREEN_CHANNEL_ANOMALY = "GREEN_CHANNEL_ANOMALY"
    BLUE_CHANNEL_ANOMALY = "BLUE_CHANNEL_ANOMALY"
    CHANNEL_CORRELATION_ANOMALY = "CHANNEL_CORRELATION_ANOMALY"
    
    # Latent space anomalies
    LATENT_OUTLIER = "LATENT_OUTLIER"
    LATENT_DISTANCE_HIGH = "LATENT_DISTANCE_HIGH"
    LATENT_DISTRIBUTION_ANOMALY = "LATENT_DISTRIBUTION_ANOMALY"
    MAHALANOBIS_DISTANCE_HIGH = "MAHALANOBIS_DISTANCE_HIGH"
    
    # Statistical anomalies
    STATISTICAL_OUTLIER = "STATISTICAL_OUTLIER"
    DISTRIBUTION_MISMATCH = "DISTRIBUTION_MISMATCH"
    FEATURE_VARIANCE_ANOMALY = "FEATURE_VARIANCE_ANOMALY"
    
    # Composite anomalies
    MULTI_METRIC_ANOMALY = "MULTI_METRIC_ANOMALY"
    CONSISTENT_ANOMALY_PATTERN = "CONSISTENT_ANOMALY_PATTERN"
    
    # Quality issues
    IMAGE_TOO_SMALL = "IMAGE_TOO_SMALL"
    IMAGE_TOO_LARGE = "IMAGE_TOO_LARGE"
    LOW_QUALITY_INPUT = "LOW_QUALITY_INPUT"
    INVALID_INPUT = "INVALID_INPUT"
    
    # System issues
    PROCESSING_ERROR = "PROCESSING_ERROR"
    MODEL_ERROR = "MODEL_ERROR"
    TIMEOUT_ERROR = "TIMEOUT_ERROR"
    ENCODER_ERROR = "ENCODER_ERROR"
    DECODER_ERROR = "DECODER_ERROR"
    
    @classmethod
    def get_category(cls, code: "AnomalyReasonCodes") -> str:
        """Get the category of a reason code."""
        success_codes = {
            cls.NO_ANOMALY_DETECTED, cls.RECONSTRUCTION_NORMAL,
            cls.LATENT_NORMAL, cls.ALL_CHECKS_PASSED
        }
        reconstruction_codes = {
            cls.HIGH_RECONSTRUCTION_ERROR, cls.MSE_ABOVIRTENGINE_THRESHOLD,
            cls.MAE_ABOVIRTENGINE_THRESHOLD, cls.SSIM_BELOW_THRESHOLD,
            cls.PERCEPTUAL_LOSS_HIGH
        }
        patch_codes = {
            cls.PATCH_ANOMALY_DETECTED, cls.LOCALIZED_RECONSTRUCTION_FAILURE,
            cls.MULTIPLE_PATCH_ANOMALIES
        }
        channel_codes = {
            cls.RED_CHANNEL_ANOMALY, cls.GREEN_CHANNEL_ANOMALY,
            cls.BLUE_CHANNEL_ANOMALY, cls.CHANNEL_CORRELATION_ANOMALY
        }
        latent_codes = {
            cls.LATENT_OUTLIER, cls.LATENT_DISTANCE_HIGH,
            cls.LATENT_DISTRIBUTION_ANOMALY, cls.MAHALANOBIS_DISTANCE_HIGH
        }
        statistical_codes = {
            cls.STATISTICAL_OUTLIER, cls.DISTRIBUTION_MISMATCH,
            cls.FEATURE_VARIANCE_ANOMALY
        }
        quality_codes = {
            cls.IMAGE_TOO_SMALL, cls.IMAGE_TOO_LARGE,
            cls.LOW_QUALITY_INPUT, cls.INVALID_INPUT
        }
        system_codes = {
            cls.PROCESSING_ERROR, cls.MODEL_ERROR, cls.TIMEOUT_ERROR,
            cls.ENCODER_ERROR, cls.DECODER_ERROR
        }
        
        if code in success_codes:
            return "success"
        elif code in reconstruction_codes:
            return "reconstruction"
        elif code in patch_codes:
            return "patch"
        elif code in channel_codes:
            return "channel"
        elif code in latent_codes:
            return "latent"
        elif code in statistical_codes:
            return "statistical"
        elif code in quality_codes:
            return "quality"
        elif code in system_codes:
            return "system"
        else:
            return "unknown"
    
    @classmethod
    def is_anomaly_code(cls, code: "AnomalyReasonCodes") -> bool:
        """Check if a reason code indicates an anomaly."""
        success_codes = {
            cls.NO_ANOMALY_DETECTED, cls.RECONSTRUCTION_NORMAL,
            cls.LATENT_NORMAL, cls.ALL_CHECKS_PASSED
        }
        quality_codes = {
            cls.IMAGE_TOO_SMALL, cls.IMAGE_TOO_LARGE,
            cls.LOW_QUALITY_INPUT, cls.INVALID_INPUT
        }
        system_codes = {
            cls.PROCESSING_ERROR, cls.MODEL_ERROR, cls.TIMEOUT_ERROR,
            cls.ENCODER_ERROR, cls.DECODER_ERROR
        }
        return code not in (success_codes | quality_codes | system_codes)


@dataclass
class ReasonCodeDetails:
    """Detailed information about a reason code."""
    
    code: AnomalyReasonCodes
    description: str
    severity: str  # "info", "warning", "error"
    category: str
    score_impact: int  # Basis points impact on VEID
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "code": self.code.value,
            "description": self.description,
            "severity": self.severity,
            "category": self.category,
            "score_impact": self.score_impact,
        }


# Detailed descriptions for all reason codes
REASON_CODE_DESCRIPTIONS: Dict[AnomalyReasonCodes, ReasonCodeDetails] = {
    AnomalyReasonCodes.NO_ANOMALY_DETECTED: ReasonCodeDetails(
        code=AnomalyReasonCodes.NO_ANOMALY_DETECTED,
        description="No anomaly detected in the input",
        severity="info",
        category="success",
        score_impact=0,
    ),
    AnomalyReasonCodes.RECONSTRUCTION_NORMAL: ReasonCodeDetails(
        code=AnomalyReasonCodes.RECONSTRUCTION_NORMAL,
        description="Reconstruction error within normal range",
        severity="info",
        category="success",
        score_impact=0,
    ),
    AnomalyReasonCodes.LATENT_NORMAL: ReasonCodeDetails(
        code=AnomalyReasonCodes.LATENT_NORMAL,
        description="Latent representation within normal distribution",
        severity="info",
        category="success",
        score_impact=0,
    ),
    AnomalyReasonCodes.ALL_CHECKS_PASSED: ReasonCodeDetails(
        code=AnomalyReasonCodes.ALL_CHECKS_PASSED,
        description="All anomaly detection checks passed",
        severity="info",
        category="success",
        score_impact=0,
    ),
    AnomalyReasonCodes.HIGH_RECONSTRUCTION_ERROR: ReasonCodeDetails(
        code=AnomalyReasonCodes.HIGH_RECONSTRUCTION_ERROR,
        description="Overall reconstruction error exceeds threshold",
        severity="error",
        category="reconstruction",
        score_impact=2000,
    ),
    AnomalyReasonCodes.MSE_ABOVIRTENGINE_THRESHOLD: ReasonCodeDetails(
        code=AnomalyReasonCodes.MSE_ABOVIRTENGINE_THRESHOLD,
        description="Mean squared error above threshold",
        severity="warning",
        category="reconstruction",
        score_impact=1000,
    ),
    AnomalyReasonCodes.MAE_ABOVIRTENGINE_THRESHOLD: ReasonCodeDetails(
        code=AnomalyReasonCodes.MAE_ABOVIRTENGINE_THRESHOLD,
        description="Mean absolute error above threshold",
        severity="warning",
        category="reconstruction",
        score_impact=800,
    ),
    AnomalyReasonCodes.SSIM_BELOW_THRESHOLD: ReasonCodeDetails(
        code=AnomalyReasonCodes.SSIM_BELOW_THRESHOLD,
        description="Structural similarity below threshold",
        severity="warning",
        category="reconstruction",
        score_impact=1200,
    ),
    AnomalyReasonCodes.PERCEPTUAL_LOSS_HIGH: ReasonCodeDetails(
        code=AnomalyReasonCodes.PERCEPTUAL_LOSS_HIGH,
        description="Perceptual loss exceeds threshold",
        severity="warning",
        category="reconstruction",
        score_impact=1500,
    ),
    AnomalyReasonCodes.PATCH_ANOMALY_DETECTED: ReasonCodeDetails(
        code=AnomalyReasonCodes.PATCH_ANOMALY_DETECTED,
        description="Anomaly detected in image patch",
        severity="warning",
        category="patch",
        score_impact=500,
    ),
    AnomalyReasonCodes.LOCALIZED_RECONSTRUCTION_FAILURE: ReasonCodeDetails(
        code=AnomalyReasonCodes.LOCALIZED_RECONSTRUCTION_FAILURE,
        description="Failed to reconstruct localized region",
        severity="error",
        category="patch",
        score_impact=1500,
    ),
    AnomalyReasonCodes.MULTIPLE_PATCH_ANOMALIES: ReasonCodeDetails(
        code=AnomalyReasonCodes.MULTIPLE_PATCH_ANOMALIES,
        description="Multiple patches show reconstruction anomalies",
        severity="error",
        category="patch",
        score_impact=2500,
    ),
    AnomalyReasonCodes.LATENT_OUTLIER: ReasonCodeDetails(
        code=AnomalyReasonCodes.LATENT_OUTLIER,
        description="Latent representation is an outlier",
        severity="error",
        category="latent",
        score_impact=2500,
    ),
    AnomalyReasonCodes.LATENT_DISTANCE_HIGH: ReasonCodeDetails(
        code=AnomalyReasonCodes.LATENT_DISTANCE_HIGH,
        description="Distance from latent centroid exceeds threshold",
        severity="warning",
        category="latent",
        score_impact=1500,
    ),
    AnomalyReasonCodes.LATENT_DISTRIBUTION_ANOMALY: ReasonCodeDetails(
        code=AnomalyReasonCodes.LATENT_DISTRIBUTION_ANOMALY,
        description="Latent vector deviates from expected distribution",
        severity="warning",
        category="latent",
        score_impact=1200,
    ),
    AnomalyReasonCodes.MAHALANOBIS_DISTANCE_HIGH: ReasonCodeDetails(
        code=AnomalyReasonCodes.MAHALANOBIS_DISTANCE_HIGH,
        description="Mahalanobis distance exceeds threshold",
        severity="error",
        category="latent",
        score_impact=2000,
    ),
    AnomalyReasonCodes.STATISTICAL_OUTLIER: ReasonCodeDetails(
        code=AnomalyReasonCodes.STATISTICAL_OUTLIER,
        description="Input classified as statistical outlier",
        severity="warning",
        category="statistical",
        score_impact=1000,
    ),
    AnomalyReasonCodes.MULTI_METRIC_ANOMALY: ReasonCodeDetails(
        code=AnomalyReasonCodes.MULTI_METRIC_ANOMALY,
        description="Multiple metrics indicate anomaly",
        severity="error",
        category="composite",
        score_impact=3000,
    ),
    AnomalyReasonCodes.IMAGE_TOO_SMALL: ReasonCodeDetails(
        code=AnomalyReasonCodes.IMAGE_TOO_SMALL,
        description="Input image below minimum size",
        severity="error",
        category="quality",
        score_impact=0,
    ),
    AnomalyReasonCodes.IMAGE_TOO_LARGE: ReasonCodeDetails(
        code=AnomalyReasonCodes.IMAGE_TOO_LARGE,
        description="Input image exceeds maximum size",
        severity="error",
        category="quality",
        score_impact=0,
    ),
    AnomalyReasonCodes.INVALID_INPUT: ReasonCodeDetails(
        code=AnomalyReasonCodes.INVALID_INPUT,
        description="Invalid input provided",
        severity="error",
        category="quality",
        score_impact=0,
    ),
    AnomalyReasonCodes.PROCESSING_ERROR: ReasonCodeDetails(
        code=AnomalyReasonCodes.PROCESSING_ERROR,
        description="Error during processing",
        severity="error",
        category="system",
        score_impact=0,
    ),
    AnomalyReasonCodes.MODEL_ERROR: ReasonCodeDetails(
        code=AnomalyReasonCodes.MODEL_ERROR,
        description="Model inference error",
        severity="error",
        category="system",
        score_impact=0,
    ),
    AnomalyReasonCodes.TIMEOUT_ERROR: ReasonCodeDetails(
        code=AnomalyReasonCodes.TIMEOUT_ERROR,
        description="Processing timeout exceeded",
        severity="error",
        category="system",
        score_impact=0,
    ),
}


def aggregate_reason_codes(
    codes: List[AnomalyReasonCodes],
    max_codes: int = 5
) -> List[str]:
    """
    Aggregate and prioritize reason codes.
    
    Args:
        codes: List of reason codes to aggregate.
        max_codes: Maximum number of codes to return.
    
    Returns:
        List of prioritized reason code values.
    """
    if not codes:
        return [AnomalyReasonCodes.NO_ANOMALY_DETECTED.value]
    
    # Sort by severity (error > warning > info)
    severity_order = {"error": 0, "warning": 1, "info": 2}
    
    def get_severity(code: AnomalyReasonCodes) -> int:
        if code in REASON_CODE_DESCRIPTIONS:
            return severity_order.get(
                REASON_CODE_DESCRIPTIONS[code].severity, 3
            )
        return 3
    
    sorted_codes = sorted(codes, key=get_severity)
    
    # Remove duplicates while preserving order
    seen = set()
    unique_codes = []
    for code in sorted_codes:
        if code not in seen:
            seen.add(code)
            unique_codes.append(code)
    
    return [c.value for c in unique_codes[:max_codes]]


def get_total_score_impact(codes: List[AnomalyReasonCodes]) -> int:
    """
    Calculate total VEID score impact from reason codes.
    
    Args:
        codes: List of reason codes.
    
    Returns:
        Total score impact in basis points.
    """
    total = 0
    for code in codes:
        if code in REASON_CODE_DESCRIPTIONS:
            total += REASON_CODE_DESCRIPTIONS[code].score_impact
    return min(total, 10000)  # Cap at 10000 basis points
