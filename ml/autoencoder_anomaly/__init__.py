"""
VirtEngine Autoencoder Anomaly Detection Module

VE-924: Autoencoder anomaly detection for identity verification

This module provides autoencoder-based anomaly detection for the VEID
identity verification system. It implements:
- Convolutional autoencoder architecture
- Reconstruction error analysis
- Latent space outlier detection
- Integration with VEID scoring

Usage:
    from ml.autoencoder_anomaly import AnomalyDetector, AutoencoderAnomalyConfig
    
    config = AutoencoderAnomalyConfig()
    detector = AnomalyDetector(config)
    result = detector.detect(image)
    
    # Check for anomaly
    if result.is_anomaly:
        print(f"Anomaly detected: {result.decision}")
        print(f"Anomaly level: {result.anomaly_level}")
    
    # Get VEID scoring record
    veid_record = result.to_veid_record()
"""

from ml.autoencoder_anomaly.config import (
    AutoencoderAnomalyConfig,
    EncoderConfig,
    DecoderConfig,
    ReconstructionConfig,
    AnomalyScoringConfig,
    VEIDIntegrationConfig,
    DetectionMode,
    AnomalyType,
    AnomalyLevel,
)
from ml.autoencoder_anomaly.detector import (
    AnomalyDetector,
    AnomalyDetectionResult,
    create_detector,
)
from ml.autoencoder_anomaly.autoencoder import (
    Autoencoder,
    AutoencoderOutput,
    ConvolutionalEncoder,
    ConvolutionalDecoder,
    EncoderOutput,
    DecoderOutput,
    create_autoencoder,
)
from ml.autoencoder_anomaly.anomaly_scorer import (
    AnomalyScorer,
    AnomalyScore,
    ReconstructionMetrics,
    LatentAnalysisResult,
    ReconstructionErrorCalculator,
    LatentSpaceAnalyzer,
    create_anomaly_scorer,
)
from ml.autoencoder_anomaly.reason_codes import (
    AnomalyReasonCodes,
    ReasonCodeDetails,
    REASON_CODE_DESCRIPTIONS,
    aggregate_reason_codes,
    get_total_score_impact,
)

__version__ = "1.0.0"
__all__ = [
    # Config
    "AutoencoderAnomalyConfig",
    "EncoderConfig",
    "DecoderConfig",
    "ReconstructionConfig",
    "AnomalyScoringConfig",
    "VEIDIntegrationConfig",
    "DetectionMode",
    "AnomalyType",
    "AnomalyLevel",
    # Main detector
    "AnomalyDetector",
    "AnomalyDetectionResult",
    "create_detector",
    # Autoencoder
    "Autoencoder",
    "AutoencoderOutput",
    "ConvolutionalEncoder",
    "ConvolutionalDecoder",
    "EncoderOutput",
    "DecoderOutput",
    "create_autoencoder",
    # Anomaly scoring
    "AnomalyScorer",
    "AnomalyScore",
    "ReconstructionMetrics",
    "LatentAnalysisResult",
    "ReconstructionErrorCalculator",
    "LatentSpaceAnalyzer",
    "create_anomaly_scorer",
    # Reason codes
    "AnomalyReasonCodes",
    "ReasonCodeDetails",
    "REASON_CODE_DESCRIPTIONS",
    "aggregate_reason_codes",
    "get_total_score_impact",
]
