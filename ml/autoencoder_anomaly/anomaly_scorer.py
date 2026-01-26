"""
Anomaly scoring based on autoencoder reconstruction error.

VE-924: Autoencoder anomaly detection - reconstruction error and scoring

This module provides:
- Reconstruction error calculation (MSE, MAE, SSIM)
- Anomaly scoring based on reconstruction error
- Latent space analysis for outlier detection
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
    ReconstructionConfig,
    AnomalyScoringConfig,
    AnomalyType,
    AnomalyLevel,
)
from ml.autoencoder_anomaly.reason_codes import (
    AnomalyReasonCodes,
    aggregate_reason_codes,
    get_total_score_impact,
)

logger = logging.getLogger(__name__)


@dataclass
class ReconstructionMetrics:
    """Metrics from reconstruction error analysis."""
    
    # Global metrics
    mse: float = 0.0  # Mean Squared Error
    mae: float = 0.0  # Mean Absolute Error
    ssim: float = 1.0  # Structural Similarity (1.0 = identical)
    psnr: float = 0.0  # Peak Signal-to-Noise Ratio
    
    # Per-channel metrics
    mse_per_channel: List[float] = field(default_factory=list)
    mae_per_channel: List[float] = field(default_factory=list)
    
    # Patch-level metrics
    patch_mse_mean: float = 0.0
    patch_mse_max: float = 0.0
    patch_mse_std: float = 0.0
    num_anomalous_patches: int = 0
    
    # Combined error score (0.0 = perfect, 1.0 = maximum error)
    combined_error: float = 0.0
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "mse": self.mse,
            "mae": self.mae,
            "ssim": self.ssim,
            "psnr": self.psnr,
            "mse_per_channel": self.mse_per_channel,
            "mae_per_channel": self.mae_per_channel,
            "patch_mse_mean": self.patch_mse_mean,
            "patch_mse_max": self.patch_mse_max,
            "patch_mse_std": self.patch_mse_std,
            "num_anomalous_patches": self.num_anomalous_patches,
            "combined_error": self.combined_error,
        }


@dataclass
class LatentAnalysisResult:
    """Result from latent space analysis."""
    
    # Distance metrics
    euclidean_distance: float = 0.0
    mahalanobis_distance: float = 0.0
    cosine_distance: float = 0.0
    
    # Distribution analysis
    z_scores: Optional[np.ndarray] = None
    max_z_score: float = 0.0
    mean_z_score: float = 0.0
    
    # Outlier detection
    is_outlier: bool = False
    outlier_dimensions: List[int] = field(default_factory=list)
    
    # Combined latent anomaly score (0.0 = normal, 1.0 = anomaly)
    latent_anomaly_score: float = 0.0
    
    def to_dict(self) -> dict:
        """Convert to dictionary (without raw arrays)."""
        return {
            "euclidean_distance": self.euclidean_distance,
            "mahalanobis_distance": self.mahalanobis_distance,
            "cosine_distance": self.cosine_distance,
            "max_z_score": self.max_z_score,
            "mean_z_score": self.mean_z_score,
            "is_outlier": self.is_outlier,
            "num_outlier_dimensions": len(self.outlier_dimensions),
            "latent_anomaly_score": self.latent_anomaly_score,
        }


@dataclass
class AnomalyScore:
    """Complete anomaly score result."""
    
    # Overall anomaly assessment
    is_anomaly: bool
    anomaly_level: AnomalyLevel
    
    # Scores (0.0 = normal, 1.0 = definite anomaly)
    overall_score: float
    confidence: float
    
    # Component scores
    reconstruction_score: float = 0.0
    latent_score: float = 0.0
    
    # Detected anomaly types
    detected_types: List[AnomalyType] = field(default_factory=list)
    
    # Reason codes
    reason_codes: List[AnomalyReasonCodes] = field(default_factory=list)
    
    # Detailed results
    reconstruction_metrics: Optional[ReconstructionMetrics] = None
    latent_analysis: Optional[LatentAnalysisResult] = None
    
    # Processing info
    processing_time_ms: float = 0.0
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "is_anomaly": self.is_anomaly,
            "anomaly_level": self.anomaly_level.value,
            "overall_score": self.overall_score,
            "confidence": self.confidence,
            "reconstruction_score": self.reconstruction_score,
            "latent_score": self.latent_score,
            "detected_types": [t.value for t in self.detected_types],
            "reason_codes": [rc.value for rc in self.reason_codes],
            "processing_time_ms": self.processing_time_ms,
        }


class ReconstructionErrorCalculator:
    """
    Calculator for reconstruction error metrics.
    
    Computes various metrics comparing original and reconstructed images:
    - MSE (Mean Squared Error)
    - MAE (Mean Absolute Error)
    - SSIM (Structural Similarity Index)
    - Per-channel and patch-level analysis
    """
    
    def __init__(self, config: Optional[ReconstructionConfig] = None):
        """
        Initialize the calculator.
        
        Args:
            config: Reconstruction configuration.
        """
        self.config = config or ReconstructionConfig()
    
    def _compute_mse(
        self,
        original: np.ndarray,
        reconstructed: np.ndarray
    ) -> float:
        """Compute Mean Squared Error."""
        diff = original.astype(np.float32) - reconstructed.astype(np.float32)
        return float(np.mean(diff ** 2))
    
    def _compute_mae(
        self,
        original: np.ndarray,
        reconstructed: np.ndarray
    ) -> float:
        """Compute Mean Absolute Error."""
        diff = original.astype(np.float32) - reconstructed.astype(np.float32)
        return float(np.mean(np.abs(diff)))
    
    def _compute_ssim(
        self,
        original: np.ndarray,
        reconstructed: np.ndarray,
        window_size: int = 11,
        k1: float = 0.01,
        k2: float = 0.03,
        L: float = 1.0
    ) -> float:
        """
        Compute Structural Similarity Index.
        
        Simplified SSIM computation for anomaly detection.
        """
        # Ensure float32
        img1 = original.astype(np.float32)
        img2 = reconstructed.astype(np.float32)
        
        # Flatten for global comparison if multi-dimensional
        if len(img1.shape) > 2:
            # Compute SSIM per channel and average
            ssim_values = []
            for c in range(img1.shape[-1]):
                ssim_c = self._compute_ssim_2d(
                    img1[:, :, c], img2[:, :, c],
                    window_size, k1, k2, L
                )
                ssim_values.append(ssim_c)
            return float(np.mean(ssim_values))
        else:
            return self._compute_ssim_2d(img1, img2, window_size, k1, k2, L)
    
    def _compute_ssim_2d(
        self,
        img1: np.ndarray,
        img2: np.ndarray,
        window_size: int = 11,
        k1: float = 0.01,
        k2: float = 0.03,
        L: float = 1.0
    ) -> float:
        """Compute SSIM for 2D images."""
        c1 = (k1 * L) ** 2
        c2 = (k2 * L) ** 2
        
        # Compute means
        mu1 = np.mean(img1)
        mu2 = np.mean(img2)
        
        # Compute variances
        sigma1_sq = np.var(img1)
        sigma2_sq = np.var(img2)
        
        # Compute covariance
        sigma12 = np.mean((img1 - mu1) * (img2 - mu2))
        
        # Compute SSIM
        numerator = (2 * mu1 * mu2 + c1) * (2 * sigma12 + c2)
        denominator = (mu1 ** 2 + mu2 ** 2 + c1) * (sigma1_sq + sigma2_sq + c2)
        
        ssim = numerator / (denominator + 1e-8)
        return float(np.clip(ssim, -1.0, 1.0))
    
    def _compute_psnr(
        self,
        original: np.ndarray,
        reconstructed: np.ndarray,
        max_val: float = 1.0
    ) -> float:
        """Compute Peak Signal-to-Noise Ratio."""
        mse = self._compute_mse(original, reconstructed)
        if mse < 1e-10:
            return 100.0  # Perfect reconstruction
        psnr = 10 * np.log10(max_val ** 2 / mse)
        return float(np.clip(psnr, 0, 100))
    
    def _compute_per_channel_metrics(
        self,
        original: np.ndarray,
        reconstructed: np.ndarray
    ) -> Tuple[List[float], List[float]]:
        """Compute MSE and MAE per channel."""
        if len(original.shape) < 3:
            return [self._compute_mse(original, reconstructed)], \
                   [self._compute_mae(original, reconstructed)]
        
        mse_per_channel = []
        mae_per_channel = []
        
        for c in range(original.shape[-1]):
            mse_per_channel.append(
                self._compute_mse(original[:, :, c], reconstructed[:, :, c])
            )
            mae_per_channel.append(
                self._compute_mae(original[:, :, c], reconstructed[:, :, c])
            )
        
        return mse_per_channel, mae_per_channel
    
    def _compute_patch_metrics(
        self,
        original: np.ndarray,
        reconstructed: np.ndarray
    ) -> Tuple[float, float, float, int]:
        """
        Compute patch-level reconstruction error.
        
        Returns:
            Tuple of (mean_mse, max_mse, std_mse, num_anomalous)
        """
        h, w = original.shape[:2]
        patch_size = self.config.patch_size
        stride = self.config.patch_stride
        
        patch_mses = []
        
        for i in range(0, h - patch_size + 1, stride):
            for j in range(0, w - patch_size + 1, stride):
                orig_patch = original[i:i+patch_size, j:j+patch_size]
                recon_patch = reconstructed[i:i+patch_size, j:j+patch_size]
                
                patch_mse = self._compute_mse(orig_patch, recon_patch)
                patch_mses.append(patch_mse)
        
        if not patch_mses:
            return 0.0, 0.0, 0.0, 0
        
        patch_mses = np.array(patch_mses)
        mean_mse = float(np.mean(patch_mses))
        max_mse = float(np.max(patch_mses))
        std_mse = float(np.std(patch_mses))
        
        # Count anomalous patches (above threshold)
        threshold = mean_mse + 2 * std_mse
        num_anomalous = int(np.sum(patch_mses > threshold))
        
        return mean_mse, max_mse, std_mse, num_anomalous
    
    def calculate(
        self,
        original: np.ndarray,
        reconstructed: np.ndarray
    ) -> ReconstructionMetrics:
        """
        Calculate reconstruction error metrics.
        
        Args:
            original: Original image (normalized to [0, 1]).
            reconstructed: Reconstructed image (normalized to [0, 1]).
        
        Returns:
            ReconstructionMetrics with all computed metrics.
        """
        # Ensure same shape
        if original.shape != reconstructed.shape:
            # Resize reconstructed to match original
            h, w = original.shape[:2]
            rh, rw = reconstructed.shape[:2]
            
            if h != rh or w != rw:
                # Simple resize
                y_ratio = rh / h
                x_ratio = rw / w
                
                y_indices = (np.arange(h) * y_ratio).astype(int)
                x_indices = (np.arange(w) * x_ratio).astype(int)
                
                y_indices = np.clip(y_indices, 0, rh - 1)
                x_indices = np.clip(x_indices, 0, rw - 1)
                
                reconstructed = reconstructed[y_indices][:, x_indices]
        
        # Compute global metrics
        mse = 0.0
        mae = 0.0
        ssim = 1.0
        
        if self.config.use_mse:
            mse = self._compute_mse(original, reconstructed)
        
        if self.config.use_mae:
            mae = self._compute_mae(original, reconstructed)
        
        if self.config.use_ssim:
            ssim = self._compute_ssim(original, reconstructed)
        
        psnr = self._compute_psnr(original, reconstructed)
        
        # Per-channel analysis
        mse_per_channel = []
        mae_per_channel = []
        
        if self.config.analyze_per_channel:
            mse_per_channel, mae_per_channel = self._compute_per_channel_metrics(
                original, reconstructed
            )
        
        # Patch analysis
        patch_mse_mean = 0.0
        patch_mse_max = 0.0
        patch_mse_std = 0.0
        num_anomalous_patches = 0
        
        if self.config.use_patch_analysis:
            patch_mse_mean, patch_mse_max, patch_mse_std, num_anomalous_patches = \
                self._compute_patch_metrics(original, reconstructed)
        
        # Combined error score
        combined_error = (
            self.config.mse_weight * min(mse / self.config.mse_threshold, 1.0) +
            self.config.mae_weight * min(mae / self.config.mae_threshold, 1.0) +
            self.config.ssim_weight * (1.0 - ssim)
        )
        combined_error = float(np.clip(combined_error, 0.0, 1.0))
        
        return ReconstructionMetrics(
            mse=mse,
            mae=mae,
            ssim=ssim,
            psnr=psnr,
            mse_per_channel=mse_per_channel,
            mae_per_channel=mae_per_channel,
            patch_mse_mean=patch_mse_mean,
            patch_mse_max=patch_mse_max,
            patch_mse_std=patch_mse_std,
            num_anomalous_patches=num_anomalous_patches,
            combined_error=combined_error,
        )


class LatentSpaceAnalyzer:
    """
    Analyzer for latent space representations.
    
    Detects anomalies based on:
    - Distance from learned distribution
    - Mahalanobis distance for outlier detection
    - Z-score analysis per dimension
    """
    
    def __init__(self, config: Optional[AnomalyScoringConfig] = None):
        """
        Initialize the analyzer.
        
        Args:
            config: Scoring configuration.
        """
        self.config = config or AnomalyScoringConfig()
        
        # Reference statistics (would be learned from training data)
        self._reference_mean: Optional[np.ndarray] = None
        self._reference_std: Optional[np.ndarray] = None
        self._reference_cov_inv: Optional[np.ndarray] = None
        
        # Initialize with default statistics
        self._initialize_reference_stats()
    
    def _initialize_reference_stats(self, latent_dim: int = 128) -> None:
        """Initialize reference statistics for latent space."""
        np.random.seed(44)
        
        # Simulated reference statistics
        # In production, these would come from training data
        self._reference_mean = np.zeros(latent_dim, dtype=np.float32)
        self._reference_std = np.ones(latent_dim, dtype=np.float32)
        
        # Identity covariance for Mahalanobis
        self._reference_cov_inv = np.eye(latent_dim, dtype=np.float32)
    
    def _compute_euclidean_distance(
        self,
        latent: np.ndarray
    ) -> float:
        """Compute Euclidean distance from reference mean."""
        if self._reference_mean is None:
            self._initialize_reference_stats(len(latent))
        
        diff = latent - self._reference_mean[:len(latent)]
        return float(np.sqrt(np.sum(diff ** 2)))
    
    def _compute_mahalanobis_distance(
        self,
        latent: np.ndarray
    ) -> float:
        """Compute Mahalanobis distance from reference distribution."""
        if self._reference_mean is None:
            self._initialize_reference_stats(len(latent))
        
        diff = latent - self._reference_mean[:len(latent)]
        cov_inv = self._reference_cov_inv[:len(latent), :len(latent)]
        
        mahal_sq = np.dot(np.dot(diff, cov_inv), diff)
        return float(np.sqrt(max(0, mahal_sq)))
    
    def _compute_cosine_distance(
        self,
        latent: np.ndarray
    ) -> float:
        """Compute cosine distance from reference mean."""
        if self._reference_mean is None:
            self._initialize_reference_stats(len(latent))
        
        ref = self._reference_mean[:len(latent)]
        
        dot_product = np.dot(latent, ref)
        norm_latent = np.linalg.norm(latent) + 1e-8
        norm_ref = np.linalg.norm(ref) + 1e-8
        
        cosine_similarity = dot_product / (norm_latent * norm_ref)
        return float(1.0 - cosine_similarity)
    
    def _compute_z_scores(
        self,
        latent: np.ndarray
    ) -> np.ndarray:
        """Compute Z-scores for each latent dimension."""
        if self._reference_std is None:
            self._initialize_reference_stats(len(latent))
        
        ref_mean = self._reference_mean[:len(latent)]
        ref_std = self._reference_std[:len(latent)]
        
        z_scores = (latent - ref_mean) / (ref_std + 1e-8)
        return z_scores
    
    def analyze(self, latent_vector: np.ndarray) -> LatentAnalysisResult:
        """
        Analyze latent representation for anomalies.
        
        Args:
            latent_vector: Latent representation from encoder.
        
        Returns:
            LatentAnalysisResult with analysis metrics.
        """
        # Compute distances
        euclidean_distance = self._compute_euclidean_distance(latent_vector)
        mahalanobis_distance = self._compute_mahalanobis_distance(latent_vector)
        cosine_distance = self._compute_cosine_distance(latent_vector)
        
        # Compute Z-scores
        z_scores = self._compute_z_scores(latent_vector)
        max_z_score = float(np.max(np.abs(z_scores)))
        mean_z_score = float(np.mean(np.abs(z_scores)))
        
        # Detect outlier dimensions (|z| > 3)
        outlier_dims = np.where(np.abs(z_scores) > 3.0)[0].tolist()
        
        # Determine if overall outlier
        is_outlier = (
            mahalanobis_distance > 3.0 * np.sqrt(len(latent_vector)) or
            len(outlier_dims) > len(latent_vector) * 0.1 or
            max_z_score > 5.0
        )
        
        # Compute latent anomaly score
        # Normalize distances to [0, 1] range
        latent_anomaly_score = min(1.0, (
            0.4 * min(mahalanobis_distance / (3.0 * np.sqrt(len(latent_vector))), 1.0) +
            0.3 * min(mean_z_score / 3.0, 1.0) +
            0.3 * min(len(outlier_dims) / (len(latent_vector) * 0.1), 1.0)
        ))
        
        return LatentAnalysisResult(
            euclidean_distance=euclidean_distance,
            mahalanobis_distance=mahalanobis_distance,
            cosine_distance=cosine_distance,
            z_scores=z_scores,
            max_z_score=max_z_score,
            mean_z_score=mean_z_score,
            is_outlier=is_outlier,
            outlier_dimensions=outlier_dims,
            latent_anomaly_score=latent_anomaly_score,
        )


class AnomalyScorer:
    """
    Complete anomaly scorer combining reconstruction and latent analysis.
    
    Usage:
        config = AutoencoderAnomalyConfig()
        scorer = AnomalyScorer(config)
        
        score = scorer.compute_anomaly_score(
            original=original_image,
            reconstructed=reconstructed_image,
            latent_vector=latent_representation
        )
    """
    
    def __init__(self, config: Optional[AutoencoderAnomalyConfig] = None):
        """
        Initialize the scorer.
        
        Args:
            config: Autoencoder configuration.
        """
        self.config = config or AutoencoderAnomalyConfig()
        
        self._reconstruction_calculator = ReconstructionErrorCalculator(
            self.config.reconstruction
        )
        self._latent_analyzer = LatentSpaceAnalyzer(
            self.config.scoring
        )
    
    def _determine_anomaly_level(
        self,
        score: float
    ) -> AnomalyLevel:
        """Determine anomaly level from score."""
        scoring = self.config.scoring
        
        if score < scoring.normal_threshold:
            return AnomalyLevel.NONE
        elif score < scoring.suspicious_threshold:
            return AnomalyLevel.LOW
        elif score < scoring.anomaly_threshold:
            return AnomalyLevel.MEDIUM
        elif score < 0.85:
            return AnomalyLevel.HIGH
        else:
            return AnomalyLevel.CRITICAL
    
    def _collect_reason_codes(
        self,
        reconstruction_metrics: ReconstructionMetrics,
        latent_analysis: LatentAnalysisResult
    ) -> List[AnomalyReasonCodes]:
        """Collect reason codes based on analysis results."""
        codes = []
        recon_cfg = self.config.reconstruction
        
        # Reconstruction-based codes
        if reconstruction_metrics.mse > recon_cfg.mse_threshold:
            codes.append(AnomalyReasonCodes.MSE_ABOVIRTENGINE_THRESHOLD)
        
        if reconstruction_metrics.mae > recon_cfg.mae_threshold:
            codes.append(AnomalyReasonCodes.MAE_ABOVIRTENGINE_THRESHOLD)
        
        if reconstruction_metrics.ssim < recon_cfg.ssim_threshold:
            codes.append(AnomalyReasonCodes.SSIM_BELOW_THRESHOLD)
        
        if reconstruction_metrics.combined_error > 0.5:
            codes.append(AnomalyReasonCodes.HIGH_RECONSTRUCTION_ERROR)
        
        if reconstruction_metrics.num_anomalous_patches > 5:
            codes.append(AnomalyReasonCodes.MULTIPLE_PATCH_ANOMALIES)
        elif reconstruction_metrics.num_anomalous_patches > 0:
            codes.append(AnomalyReasonCodes.PATCH_ANOMALY_DETECTED)
        
        # Check per-channel anomalies
        for i, mse in enumerate(reconstruction_metrics.mse_per_channel):
            if mse > recon_cfg.mse_threshold * 1.5:
                if i == 0:
                    codes.append(AnomalyReasonCodes.RED_CHANNEL_ANOMALY)
                elif i == 1:
                    codes.append(AnomalyReasonCodes.GREEN_CHANNEL_ANOMALY)
                elif i == 2:
                    codes.append(AnomalyReasonCodes.BLUE_CHANNEL_ANOMALY)
        
        # Latent space codes
        if latent_analysis.is_outlier:
            codes.append(AnomalyReasonCodes.LATENT_OUTLIER)
        
        if latent_analysis.mahalanobis_distance > 50:
            codes.append(AnomalyReasonCodes.MAHALANOBIS_DISTANCE_HIGH)
        elif latent_analysis.latent_anomaly_score > 0.5:
            codes.append(AnomalyReasonCodes.LATENT_DISTANCE_HIGH)
        
        if latent_analysis.max_z_score > 4.0:
            codes.append(AnomalyReasonCodes.LATENT_DISTRIBUTION_ANOMALY)
        
        # Multi-metric anomaly
        num_anomaly_codes = sum(
            1 for c in codes if AnomalyReasonCodes.is_anomaly_code(c)
        )
        if num_anomaly_codes >= 3:
            codes.append(AnomalyReasonCodes.MULTI_METRIC_ANOMALY)
        
        # Success codes if no anomalies
        if not codes:
            codes.append(AnomalyReasonCodes.ALL_CHECKS_PASSED)
            codes.append(AnomalyReasonCodes.RECONSTRUCTION_NORMAL)
            codes.append(AnomalyReasonCodes.LATENT_NORMAL)
        
        return codes
    
    def _detect_anomaly_types(
        self,
        reconstruction_metrics: ReconstructionMetrics,
        latent_analysis: LatentAnalysisResult
    ) -> List[AnomalyType]:
        """Detect types of anomalies present."""
        types = []
        
        if reconstruction_metrics.combined_error > 0.5:
            types.append(AnomalyType.RECONSTRUCTION_HIGH)
        
        if latent_analysis.is_outlier:
            types.append(AnomalyType.LATENT_OUTLIER)
        
        if latent_analysis.latent_anomaly_score > 0.5:
            types.append(AnomalyType.FEATURE_ANOMALY)
        
        if latent_analysis.max_z_score > 4.0:
            types.append(AnomalyType.STATISTICAL_OUTLIER)
        
        return types
    
    def compute_anomaly_score(
        self,
        original: np.ndarray,
        reconstructed: np.ndarray,
        latent_vector: np.ndarray,
        include_details: bool = False
    ) -> AnomalyScore:
        """
        Compute comprehensive anomaly score.
        
        Args:
            original: Original image (normalized to [0, 1]).
            reconstructed: Reconstructed image (normalized to [0, 1]).
            latent_vector: Latent representation from encoder.
            include_details: Include detailed metric results.
        
        Returns:
            AnomalyScore with complete analysis.
        """
        start_time = time.time()
        
        # Calculate reconstruction error
        reconstruction_metrics = self._reconstruction_calculator.calculate(
            original, reconstructed
        )
        
        # Analyze latent space
        latent_analysis = self._latent_analyzer.analyze(latent_vector)
        
        # Compute component scores
        reconstruction_score = reconstruction_metrics.combined_error
        latent_score = latent_analysis.latent_anomaly_score
        
        # Combine scores
        scoring = self.config.scoring
        overall_score = (
            scoring.reconstruction_weight * reconstruction_score +
            scoring.latent_weight * latent_score
        )
        overall_score = float(np.clip(overall_score, 0.0, 1.0))
        
        # Determine if anomaly
        is_anomaly = overall_score >= scoring.normal_threshold
        anomaly_level = self._determine_anomaly_level(overall_score)
        
        # Compute confidence
        # Higher confidence when score is far from threshold
        distance_from_threshold = abs(overall_score - scoring.suspicious_threshold)
        confidence = min(1.0, 0.5 + distance_from_threshold)
        
        # Collect reason codes
        reason_codes = self._collect_reason_codes(
            reconstruction_metrics, latent_analysis
        )
        
        # Detect anomaly types
        detected_types = self._detect_anomaly_types(
            reconstruction_metrics, latent_analysis
        )
        
        processing_time = (time.time() - start_time) * 1000
        
        return AnomalyScore(
            is_anomaly=is_anomaly,
            anomaly_level=anomaly_level,
            overall_score=overall_score,
            confidence=confidence,
            reconstruction_score=reconstruction_score,
            latent_score=latent_score,
            detected_types=detected_types,
            reason_codes=reason_codes,
            reconstruction_metrics=reconstruction_metrics if include_details else None,
            latent_analysis=latent_analysis if include_details else None,
            processing_time_ms=processing_time,
        )


def create_anomaly_scorer(
    config: Optional[AutoencoderAnomalyConfig] = None
) -> AnomalyScorer:
    """
    Factory function to create an anomaly scorer.
    
    Args:
        config: Autoencoder configuration.
    
    Returns:
        Initialized AnomalyScorer instance.
    """
    return AnomalyScorer(config)
