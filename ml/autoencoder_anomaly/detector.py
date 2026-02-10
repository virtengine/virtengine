"""
Main autoencoder anomaly detection module.

VE-924: Autoencoder anomaly detection - main detector implementation

This module provides the main AnomalyDetector class that combines:
- Autoencoder-based reconstruction
- Reconstruction error analysis
- Latent space anomaly detection
- Integration with VEID scoring
"""

import logging
import hashlib
import time
from dataclasses import dataclass, field
from typing import List, Optional, Tuple, Dict, Any

import numpy as np

from ml.autoencoder_anomaly.config import (
    AutoencoderAnomalyConfig,
    DetectionMode,
    AnomalyType,
    AnomalyLevel,
)
from ml.autoencoder_anomaly.autoencoder import (
    Autoencoder,
    AutoencoderOutput,
    create_autoencoder,
)
from ml.autoencoder_anomaly.anomaly_scorer import (
    AnomalyScorer,
    AnomalyScore,
    ReconstructionMetrics,
    LatentAnalysisResult,
    create_anomaly_scorer,
)
from ml.autoencoder_anomaly.reason_codes import (
    AnomalyReasonCodes,
    aggregate_reason_codes,
    get_total_score_impact,
)

logger = logging.getLogger(__name__)


@dataclass
class AnomalyDetectionResult:
    """Complete result of autoencoder anomaly detection."""
    
    # Overall decision
    is_anomaly: bool
    decision: str  # "normal", "suspicious", "anomaly", "uncertain"
    anomaly_level: AnomalyLevel
    
    # Scores (0.0 = normal, 1.0 = definite anomaly)
    overall_score: float
    confidence: float
    
    # Component scores
    reconstruction_score: float = 0.0
    latent_score: float = 0.0
    
    # Detected anomaly types
    detected_types: List[AnomalyType] = field(default_factory=list)
    
    # VEID integration
    veid_penalty: int = 0  # Basis points penalty for VEID score
    veid_adjusted_score: int = 10000  # Score after penalty (max 10000)
    
    # Model info
    model_version: str = "1.0.0"
    model_hash: str = ""
    
    # Reason codes
    reason_codes: List[str] = field(default_factory=list)
    
    # Hashes for consensus verification
    result_hash: str = ""
    
    # Timing
    processing_time_ms: float = 0.0
    
    # Detailed results (optional)
    autoencoder_output: Optional[Dict[str, Any]] = None
    anomaly_score_details: Optional[Dict[str, Any]] = None
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "is_anomaly": self.is_anomaly,
            "decision": self.decision,
            "anomaly_level": self.anomaly_level.value,
            "overall_score": self.overall_score,
            "confidence": self.confidence,
            "reconstruction_score": self.reconstruction_score,
            "latent_score": self.latent_score,
            "detected_types": [t.value for t in self.detected_types],
            "veid_penalty": self.veid_penalty,
            "veid_adjusted_score": self.veid_adjusted_score,
            "model_version": self.model_version,
            "model_hash": self.model_hash,
            "reason_codes": self.reason_codes,
            "result_hash": self.result_hash,
            "processing_time_ms": self.processing_time_ms,
        }
    
    def to_veid_record(self) -> dict:
        """
        Convert to VEID scoring record format.
        
        This format is used for integration with the on-chain
        VEID scoring model.
        """
        return {
            "anomaly_score": int((1 - self.overall_score) * 10000),  # Inverted: higher = more normal
            "is_anomaly": self.is_anomaly,
            "decision": self.decision,
            "confidence": int(self.confidence * 100),
            "anomaly_level": self.anomaly_level.value,
            "detected_types": [t.value for t in self.detected_types[:3]],  # Limit for storage
            "veid_penalty": self.veid_penalty,
            "veid_adjusted_score": self.veid_adjusted_score,
            "model_version": self.model_version,
            "model_hash": self.model_hash[:16] if self.model_hash else "",
            "result_hash": self.result_hash,
            "reason_codes": self.reason_codes[:5],  # Limit for storage
        }


class AnomalyDetector:
    """
    Main autoencoder-based anomaly detection pipeline.
    
    This class orchestrates the complete anomaly detection process:
    1. Image preprocessing and encoding
    2. Reconstruction via autoencoder
    3. Reconstruction error calculation
    4. Latent space analysis
    5. Anomaly scoring and VEID integration
    
    Usage:
        config = AutoencoderAnomalyConfig()
        detector = AnomalyDetector(config)
        
        # Analyze image
        result = detector.detect(image)
        
        # Check for anomaly
        if result.is_anomaly:
            print(f"Anomaly detected: {result.decision}")
            print(f"Anomaly types: {result.detected_types}")
        
        # Get VEID record
        veid_record = result.to_veid_record()
    """
    
    MODEL_VERSION = "1.0.0"
    
    def __init__(self, config: Optional[AutoencoderAnomalyConfig] = None):
        """
        Initialize the anomaly detector.
        
        Args:
            config: Anomaly detection configuration.
        """
        self.config = config or AutoencoderAnomalyConfig()
        
        # Set determinism
        if self.config.enforce_determinism:
            np.random.seed(self.config.random_seed)
        
        # Initialize components
        self._autoencoder = create_autoencoder(self.config)
        self._scorer = create_anomaly_scorer(self.config)
        
        # Compute model hash
        self._model_hash = self._compute_model_hash()
        
        logger.info(
            f"AnomalyDetector initialized: version={self.MODEL_VERSION}, "
            f"mode={self.config.mode.value}, hash={self._model_hash[:8]}"
        )
    
    def _compute_model_hash(self) -> str:
        """Compute deterministic hash of model configuration."""
        config_str = (
            f"version:{self.MODEL_VERSION},"
            f"mode:{self.config.mode.value},"
            f"autoencoder:{self._autoencoder.model_hash},"
            f"seed:{self.config.random_seed}"
        )
        return hashlib.sha256(config_str.encode()).hexdigest()[:32]
    
    @property
    def model_hash(self) -> str:
        """Get model hash."""
        return self._model_hash
    
    def _compute_result_hash(
        self,
        overall_score: float,
        is_anomaly: bool,
        reason_codes: List[str]
    ) -> str:
        """Compute deterministic hash of detection result."""
        result_str = (
            f"score:{overall_score:.6f},"
            f"anomaly:{is_anomaly},"
            f"codes:{','.join(sorted(reason_codes))},"
            f"model:{self._model_hash}"
        )
        return hashlib.sha256(result_str.encode()).hexdigest()[:32]
    
    def _compute_veid_penalty(
        self,
        anomaly_level: AnomalyLevel,
        confidence: float
    ) -> int:
        """Compute VEID penalty based on anomaly level."""
        veid_config = self.config.veid_integration
        
        base_penalty = {
            AnomalyLevel.NONE: 0,
            AnomalyLevel.LOW: veid_config.low_anomaly_penalty,
            AnomalyLevel.MEDIUM: veid_config.medium_anomaly_penalty,
            AnomalyLevel.HIGH: veid_config.high_anomaly_penalty,
            AnomalyLevel.CRITICAL: veid_config.critical_anomaly_penalty,
        }.get(anomaly_level, 0)
        
        # Scale by confidence if enabled
        if veid_config.apply_confidence_scaling:
            if confidence < veid_config.min_confidence_for_impact:
                # Low confidence reduces penalty
                scale = confidence / veid_config.min_confidence_for_impact
                base_penalty = int(base_penalty * scale)
        
        return min(base_penalty, veid_config.max_veid_impact)
    
    def _determine_decision(
        self,
        overall_score: float,
        anomaly_level: AnomalyLevel
    ) -> str:
        """Determine decision string from score and level."""
        if anomaly_level == AnomalyLevel.NONE:
            return "normal"
        elif anomaly_level == AnomalyLevel.LOW:
            return "suspicious"
        elif anomaly_level in (AnomalyLevel.MEDIUM, AnomalyLevel.HIGH):
            return "anomaly"
        elif anomaly_level == AnomalyLevel.CRITICAL:
            return "anomaly"
        else:
            return "uncertain"
    
    def _preprocess_for_reconstruction(
        self,
        image: np.ndarray
    ) -> np.ndarray:
        """
        Preprocess image for reconstruction comparison.
        
        Args:
            image: Input image (BGR format, uint8).
        
        Returns:
            Normalized image (RGB, float32, [0, 1]).
        """
        # Convert BGR to RGB
        if len(image.shape) == 3 and image.shape[2] == 3:
            image = image[:, :, ::-1]
        
        # Resize to autoencoder input size
        h, w = image.shape[:2]
        target_h, target_w = self.config.encoder.input_size
        
        if h != target_h or w != target_w:
            y_ratio = h / target_h
            x_ratio = w / target_w
            
            y_indices = (np.arange(target_h) * y_ratio).astype(int)
            x_indices = (np.arange(target_w) * x_ratio).astype(int)
            
            y_indices = np.clip(y_indices, 0, h - 1)
            x_indices = np.clip(x_indices, 0, w - 1)
            
            image = image[y_indices][:, x_indices]
        
        # Normalize to [0, 1]
        image = image.astype(np.float32) / 255.0
        
        return image
    
    def detect(
        self,
        image: np.ndarray,
        include_details: bool = False
    ) -> AnomalyDetectionResult:
        """
        Perform anomaly detection on an image.
        
        Args:
            image: Input image (BGR format, uint8).
            include_details: Include detailed component results.
        
        Returns:
            AnomalyDetectionResult with detection outcome.
        """
        start_time = time.time()
        
        # Validate input
        if image is None or image.size == 0:
            return AnomalyDetectionResult(
                is_anomaly=False,
                decision="uncertain",
                anomaly_level=AnomalyLevel.NONE,
                overall_score=0.0,
                confidence=0.0,
                reason_codes=[AnomalyReasonCodes.INVALID_INPUT.value],
                model_version=self.MODEL_VERSION,
                model_hash=self._model_hash,
            )
        
        # Validate image size
        h, w = image.shape[:2]
        min_h, min_w = self.config.min_image_size
        max_h, max_w = self.config.max_image_size
        
        if h < min_h or w < min_w:
            return AnomalyDetectionResult(
                is_anomaly=False,
                decision="uncertain",
                anomaly_level=AnomalyLevel.NONE,
                overall_score=0.0,
                confidence=0.0,
                reason_codes=[AnomalyReasonCodes.IMAGE_TOO_SMALL.value],
                model_version=self.MODEL_VERSION,
                model_hash=self._model_hash,
            )
        
        if h > max_h or w > max_w:
            return AnomalyDetectionResult(
                is_anomaly=False,
                decision="uncertain",
                anomaly_level=AnomalyLevel.NONE,
                overall_score=0.0,
                confidence=0.0,
                reason_codes=[AnomalyReasonCodes.IMAGE_TOO_LARGE.value],
                model_version=self.MODEL_VERSION,
                model_hash=self._model_hash,
            )
        
        try:
            # Run autoencoder forward pass
            autoencoder_output = self._autoencoder.forward(
                image, return_features=include_details
            )
            
            # Preprocess original for comparison
            original_normalized = self._preprocess_for_reconstruction(image)
            
            # Get reconstruction
            reconstruction = autoencoder_output.decoder_output.reconstruction
            
            # Get latent vector
            latent_vector = autoencoder_output.encoder_output.latent_vector
            
            # Compute anomaly score
            anomaly_score = self._scorer.compute_anomaly_score(
                original=original_normalized,
                reconstructed=reconstruction,
                latent_vector=latent_vector,
                include_details=include_details,
            )
            
            # Determine decision and penalty
            decision = self._determine_decision(
                anomaly_score.overall_score,
                anomaly_score.anomaly_level
            )
            
            veid_penalty = self._compute_veid_penalty(
                anomaly_score.anomaly_level,
                anomaly_score.confidence
            )
            
            veid_adjusted_score = 10000 - veid_penalty
            
            # Aggregate reason codes
            reason_codes = aggregate_reason_codes(anomaly_score.reason_codes)
            
            # Compute result hash
            result_hash = self._compute_result_hash(
                anomaly_score.overall_score,
                anomaly_score.is_anomaly,
                reason_codes
            )
            
            processing_time = (time.time() - start_time) * 1000
            
            return AnomalyDetectionResult(
                is_anomaly=anomaly_score.is_anomaly,
                decision=decision,
                anomaly_level=anomaly_score.anomaly_level,
                overall_score=anomaly_score.overall_score,
                confidence=anomaly_score.confidence,
                reconstruction_score=anomaly_score.reconstruction_score,
                latent_score=anomaly_score.latent_score,
                detected_types=anomaly_score.detected_types,
                veid_penalty=veid_penalty,
                veid_adjusted_score=veid_adjusted_score,
                model_version=self.MODEL_VERSION,
                model_hash=self._model_hash,
                reason_codes=reason_codes,
                result_hash=result_hash,
                processing_time_ms=processing_time,
                autoencoder_output=autoencoder_output.to_dict() if include_details else None,
                anomaly_score_details=anomaly_score.to_dict() if include_details else None,
            )
            
        except Exception as e:
            logger.error(f"Anomaly detection error: {e}")
            return AnomalyDetectionResult(
                is_anomaly=False,
                decision="uncertain",
                anomaly_level=AnomalyLevel.NONE,
                overall_score=0.0,
                confidence=0.0,
                reason_codes=[AnomalyReasonCodes.PROCESSING_ERROR.value],
                model_version=self.MODEL_VERSION,
                model_hash=self._model_hash,
            )
    
    def detect_batch(
        self,
        images: List[np.ndarray],
        include_details: bool = False
    ) -> List[AnomalyDetectionResult]:
        """
        Perform anomaly detection on multiple images.
        
        Args:
            images: List of input images.
            include_details: Include detailed component results.
        
        Returns:
            List of AnomalyDetectionResult objects.
        """
        results = []
        for image in images:
            result = self.detect(image, include_details)
            results.append(result)
        return results
    
    def get_reconstruction(
        self,
        image: np.ndarray
    ) -> Tuple[np.ndarray, np.ndarray]:
        """
        Get autoencoder reconstruction for visualization.
        
        Args:
            image: Input image (BGR format, uint8).
        
        Returns:
            Tuple of (original_normalized, reconstruction).
        
        Note: This method is for debugging/visualization only.
              Never log or store actual image data in production.
        """
        # Preprocess original
        original_normalized = self._preprocess_for_reconstruction(image)
        
        # Get reconstruction
        output = self._autoencoder.forward(image)
        reconstruction = output.decoder_output.reconstruction
        
        return original_normalized, reconstruction


def create_detector(
    config: Optional[AutoencoderAnomalyConfig] = None
) -> AnomalyDetector:
    """
    Factory function to create an anomaly detector.
    
    Args:
        config: Anomaly detection configuration.
    
    Returns:
        Initialized AnomalyDetector instance.
    """
    return AnomalyDetector(config)
