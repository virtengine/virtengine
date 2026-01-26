"""
Artifact analysis module for GAN-generated image detection.

VE-923: GAN fraud detection - artifact analysis

This module provides detection of GAN-specific artifacts:
- Frequency domain anomalies
- Checkerboard patterns
- Blending boundaries
- Texture inconsistencies
- Upsampling artifacts
"""

import logging
import hashlib
from dataclasses import dataclass, field
from typing import List, Optional, Tuple, Dict, Any

import numpy as np

from ml.gan_detection.config import (
    GANDetectionConfig,
    ArtifactAnalysisConfig,
    ArtifactType,
)
from ml.gan_detection.reason_codes import (
    GANReasonCodes,
    ReasonCodeDetails,
)

logger = logging.getLogger(__name__)


@dataclass
class ArtifactResult:
    """Result of artifact analysis."""
    
    # Overall detection
    has_artifacts: bool
    artifact_score: float  # 0.0 = clean, 1.0 = heavy artifacts
    confidence: float
    
    # Individual artifact scores
    frequency_score: float = 0.0
    checkerboard_score: float = 0.0
    blending_score: float = 0.0
    texture_score: float = 0.0
    color_score: float = 0.0
    compression_score: float = 0.0
    upsampling_score: float = 0.0
    
    # Detected artifact types
    detected_artifacts: List[ArtifactType] = field(default_factory=list)
    
    # Reason codes
    reason_codes: List[GANReasonCodes] = field(default_factory=list)
    
    # Processing info
    processing_time_ms: float = 0.0
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "has_artifacts": self.has_artifacts,
            "artifact_score": self.artifact_score,
            "confidence": self.confidence,
            "frequency_score": self.frequency_score,
            "checkerboard_score": self.checkerboard_score,
            "blending_score": self.blending_score,
            "texture_score": self.texture_score,
            "color_score": self.color_score,
            "compression_score": self.compression_score,
            "upsampling_score": self.upsampling_score,
            "detected_artifacts": [a.value for a in self.detected_artifacts],
            "reason_codes": [rc.value for rc in self.reason_codes],
            "processing_time_ms": self.processing_time_ms,
        }


class ArtifactAnalyzer:
    """
    Analyzer for GAN-specific artifacts.
    
    This class detects various artifacts commonly found in GAN-generated images:
    - Frequency domain anomalies (unusual spectral patterns)
    - Checkerboard patterns (from transposed convolutions)
    - Blending boundaries (from inpainting/editing)
    - Texture inconsistencies
    - Upsampling artifacts
    
    Usage:
        config = ArtifactAnalysisConfig()
        analyzer = ArtifactAnalyzer(config)
        result = analyzer.analyze(image, face_region)
    """
    
    def __init__(self, config: Optional[ArtifactAnalysisConfig] = None):
        """
        Initialize the artifact analyzer.
        
        Args:
            config: Artifact analysis configuration.
        """
        self.config = config or ArtifactAnalysisConfig()
    
    def analyze(
        self,
        image: np.ndarray,
        face_region: Optional[Tuple[int, int, int, int]] = None
    ) -> ArtifactResult:
        """
        Analyze image for GAN artifacts.
        
        Args:
            image: Input image (BGR format, uint8).
            face_region: Optional face bounding box for focused analysis.
        
        Returns:
            ArtifactResult with analysis outcome.
        """
        import time
        start_time = time.time()
        
        reason_codes = []
        detected_artifacts = []
        
        # Validate input
        if image is None or image.size == 0:
            return ArtifactResult(
                has_artifacts=False,
                artifact_score=0.0,
                confidence=0.0,
                reason_codes=[GANReasonCodes.INVALID_INPUT],
            )
        
        # Frequency analysis
        frequency_score = 0.0
        if self.config.use_frequency_analysis:
            frequency_score = self._analyze_frequency(image)
            if frequency_score > self.config.frequency_threshold:
                detected_artifacts.append(ArtifactType.FREQUENCY_ANOMALY)
                reason_codes.append(GANReasonCodes.FREQUENCY_ANOMALY)
        
        # Checkerboard detection
        checkerboard_score = 0.0
        if self.config.detect_checkerboard:
            checkerboard_score = self._detect_checkerboard(image)
            if checkerboard_score > self.config.checkerboard_threshold:
                detected_artifacts.append(ArtifactType.CHECKERBOARD)
                reason_codes.append(GANReasonCodes.CHECKERBOARD_ARTIFACT)
        
        # Blending boundary detection
        blending_score = 0.0
        if self.config.detect_blending:
            blending_score = self._detect_blending_boundary(image, face_region)
            if blending_score > self.config.blending_threshold:
                detected_artifacts.append(ArtifactType.BLENDING_BOUNDARY)
                reason_codes.append(GANReasonCodes.BLENDING_BOUNDARY_DETECTED)
        
        # Texture consistency analysis
        texture_score = 0.0
        if self.config.analyze_texture_consistency:
            texture_score = self._analyze_texture_consistency(image)
            if texture_score > self.config.texture_threshold:
                detected_artifacts.append(ArtifactType.TEXTURE_INCONSISTENCY)
                reason_codes.append(GANReasonCodes.TEXTURE_INCONSISTENCY)
        
        # Color consistency analysis
        color_score = 0.0
        if self.config.analyze_color_consistency:
            color_score = self._analyze_color_consistency(image, face_region)
            if color_score > self.config.color_threshold:
                detected_artifacts.append(ArtifactType.COLOR_MISMATCH)
                reason_codes.append(GANReasonCodes.COLOR_MISMATCH)
        
        # Compression artifact detection
        compression_score = 0.0
        if self.config.detect_compression_artifacts:
            compression_score = self._detect_compression_artifacts(image)
            if compression_score > 0.5:
                detected_artifacts.append(ArtifactType.COMPRESSION_ARTIFACT)
                reason_codes.append(GANReasonCodes.COMPRESSION_ANOMALY)
        
        # Upsampling artifact detection
        upsampling_score = 0.0
        if self.config.detect_upsampling:
            upsampling_score = self._detect_upsampling_artifacts(image)
            if upsampling_score > self.config.upsampling_threshold:
                detected_artifacts.append(ArtifactType.UPSAMPLING_ARTIFACT)
                reason_codes.append(GANReasonCodes.UPSAMPLING_ARTIFACT)
        
        # Compute overall score
        artifact_score = self._compute_overall_score(
            frequency_score, checkerboard_score, blending_score,
            texture_score, color_score, compression_score, upsampling_score
        )
        
        has_artifacts = len(detected_artifacts) > 0 or artifact_score > 0.5
        confidence = abs(artifact_score - 0.5) * 2
        
        if not has_artifacts:
            reason_codes.append(GANReasonCodes.ALL_CHECKS_PASSED)
        
        processing_time = (time.time() - start_time) * 1000
        
        return ArtifactResult(
            has_artifacts=has_artifacts,
            artifact_score=artifact_score,
            confidence=confidence,
            frequency_score=frequency_score,
            checkerboard_score=checkerboard_score,
            blending_score=blending_score,
            texture_score=texture_score,
            color_score=color_score,
            compression_score=compression_score,
            upsampling_score=upsampling_score,
            detected_artifacts=detected_artifacts,
            reason_codes=list(set(reason_codes)),
            processing_time_ms=processing_time,
        )
    
    def _analyze_frequency(self, image: np.ndarray) -> float:
        """
        Analyze frequency domain for GAN artifacts.
        
        GANs often leave characteristic patterns in the frequency domain,
        especially high-frequency artifacts from upsampling layers.
        """
        # Convert to grayscale
        if len(image.shape) == 3:
            gray = np.mean(image, axis=2)
        else:
            gray = image.astype(np.float32)
        
        # Resize to FFT size
        h, w = gray.shape
        fft_size = self.config.fft_size
        
        # Crop or pad to FFT size
        if h >= fft_size and w >= fft_size:
            # Center crop
            start_h = (h - fft_size) // 2
            start_w = (w - fft_size) // 2
            gray = gray[start_h:start_h+fft_size, start_w:start_w+fft_size]
        else:
            # Pad
            padded = np.zeros((fft_size, fft_size), dtype=np.float32)
            start_h = (fft_size - h) // 2
            start_w = (fft_size - w) // 2
            padded[start_h:start_h+h, start_w:start_w+w] = gray[:fft_size, :fft_size]
            gray = padded
        
        # Compute FFT
        fft = np.fft.fft2(gray)
        fft_shift = np.fft.fftshift(fft)
        magnitude = np.abs(fft_shift)
        
        # Log transform for better visualization
        magnitude_log = np.log1p(magnitude)
        
        # Analyze radial power spectrum
        center = fft_size // 2
        
        # Create radial bins
        y, x = np.ogrid[:fft_size, :fft_size]
        r = np.sqrt((x - center)**2 + (y - center)**2)
        r = r.astype(int)
        
        # Compute mean power at each radius
        max_r = min(center, fft_size - center)
        radial_power = np.zeros(max_r)
        for i in range(max_r):
            mask = (r == i)
            if np.any(mask):
                radial_power[i] = np.mean(magnitude_log[mask])
        
        # Check for anomalous high-frequency content
        # GANs often have unusual patterns in mid-to-high frequencies
        mid_freq = radial_power[max_r//4:3*max_r//4]
        high_freq = radial_power[3*max_r//4:]
        
        if len(mid_freq) > 0 and len(high_freq) > 0:
            # Check for unexpected peaks or patterns
            mid_var = np.var(mid_freq)
            high_mean = np.mean(high_freq)
            
            # Anomaly score based on high-frequency characteristics
            anomaly_score = (mid_var / 100.0) + (high_mean / np.mean(radial_power) - 0.1)
            anomaly_score = min(1.0, max(0.0, anomaly_score))
        else:
            anomaly_score = 0.0
        
        return float(anomaly_score)
    
    def _detect_checkerboard(self, image: np.ndarray) -> float:
        """
        Detect checkerboard artifacts from transposed convolutions.
        
        These appear as regular grid-like patterns in the image.
        """
        # Convert to grayscale
        if len(image.shape) == 3:
            gray = np.mean(image, axis=2).astype(np.float32)
        else:
            gray = image.astype(np.float32)
        
        h, w = gray.shape
        
        if h < 16 or w < 16:
            return 0.0
        
        # Compute horizontal and vertical gradients
        grad_x = np.abs(np.diff(gray, axis=1))
        grad_y = np.abs(np.diff(gray, axis=0))
        
        # Look for alternating pattern (checkerboard)
        # Compute second derivative of gradients
        grad_xx = np.diff(grad_x, axis=1)
        grad_yy = np.diff(grad_y, axis=0)
        
        # Checkerboard pattern has sign changes at regular intervals
        sign_changes_x = np.sum(np.abs(np.diff(np.sign(grad_xx), axis=1)))
        sign_changes_y = np.sum(np.abs(np.diff(np.sign(grad_yy), axis=0)))
        
        # Normalize by image size
        total_pixels = (h - 2) * (w - 3) + (h - 3) * (w - 2)
        if total_pixels > 0:
            change_ratio = (sign_changes_x + sign_changes_y) / total_pixels
        else:
            change_ratio = 0.0
        
        # High change ratio indicates checkerboard pattern
        score = min(1.0, change_ratio * 10)
        
        return float(score)
    
    def _detect_blending_boundary(
        self,
        image: np.ndarray,
        face_region: Optional[Tuple[int, int, int, int]]
    ) -> float:
        """Detect blending boundaries around face region."""
        if face_region is None:
            h, w = image.shape[:2]
            face_region = (w // 4, h // 4, 3 * w // 4, 3 * h // 4)
        
        x1, y1, x2, y2 = face_region
        h, w = image.shape[:2]
        
        # Ensure valid region
        x1 = max(0, min(x1, w - 1))
        x2 = max(x1 + 1, min(x2, w))
        y1 = max(0, min(y1, h - 1))
        y2 = max(y1 + 1, min(y2, h))
        
        boundary_width = 10
        
        scores = []
        
        # Analyze each boundary
        for kernel_size in self.config.blending_kernel_sizes:
            # Top boundary
            if y1 > boundary_width:
                inner = image[y1:y1+boundary_width, x1:x2]
                outer = image[y1-boundary_width:y1, x1:x2]
                if inner.size > 0 and outer.size > 0:
                    diff = np.mean(np.abs(
                        np.mean(inner, axis=0).astype(np.float32) -
                        np.mean(outer, axis=0).astype(np.float32)
                    ))
                    scores.append(diff)
            
            # Bottom boundary
            if y2 + boundary_width < h:
                inner = image[y2-boundary_width:y2, x1:x2]
                outer = image[y2:y2+boundary_width, x1:x2]
                if inner.size > 0 and outer.size > 0:
                    diff = np.mean(np.abs(
                        np.mean(inner, axis=0).astype(np.float32) -
                        np.mean(outer, axis=0).astype(np.float32)
                    ))
                    scores.append(diff)
            
            # Left boundary
            if x1 > boundary_width:
                inner = image[y1:y2, x1:x1+boundary_width]
                outer = image[y1:y2, x1-boundary_width:x1]
                if inner.size > 0 and outer.size > 0:
                    diff = np.mean(np.abs(
                        np.mean(inner, axis=1).astype(np.float32) -
                        np.mean(outer, axis=1).astype(np.float32)
                    ))
                    scores.append(diff)
            
            # Right boundary
            if x2 + boundary_width < w:
                inner = image[y1:y2, x2-boundary_width:x2]
                outer = image[y1:y2, x2:x2+boundary_width]
                if inner.size > 0 and outer.size > 0:
                    diff = np.mean(np.abs(
                        np.mean(inner, axis=1).astype(np.float32) -
                        np.mean(outer, axis=1).astype(np.float32)
                    ))
                    scores.append(diff)
        
        if not scores:
            return 0.0
        
        # High boundary difference = suspicious
        avg_score = np.mean(scores)
        normalized_score = min(1.0, avg_score / 50.0)
        
        return float(normalized_score)
    
    def _analyze_texture_consistency(self, image: np.ndarray) -> float:
        """Analyze texture consistency across the image."""
        if len(image.shape) == 3:
            gray = np.mean(image, axis=2).astype(np.float32)
        else:
            gray = image.astype(np.float32)
        
        h, w = gray.shape
        patch_size = self.config.texture_patch_size
        
        if h < patch_size * 2 or w < patch_size * 2:
            return 0.0
        
        # Compute texture features for patches
        patch_variances = []
        
        for y in range(0, h - patch_size, patch_size):
            for x in range(0, w - patch_size, patch_size):
                patch = gray[y:y+patch_size, x:x+patch_size]
                
                # Compute local variance as texture measure
                variance = np.var(patch)
                patch_variances.append(variance)
        
        if len(patch_variances) < 2:
            return 0.0
        
        # Check for inconsistent texture patterns
        variance_array = np.array(patch_variances)
        
        # Coefficient of variation
        mean_var = np.mean(variance_array)
        if mean_var > 0:
            cv = np.std(variance_array) / mean_var
        else:
            cv = 0.0
        
        # Very high CV = inconsistent textures (suspicious)
        # Very low CV = too uniform (also suspicious)
        if cv > 2.0:
            score = min(1.0, (cv - 2.0) / 2.0)
        elif cv < 0.2:
            score = min(1.0, (0.2 - cv) / 0.2 * 0.5)
        else:
            score = 0.0
        
        return float(score)
    
    def _analyze_color_consistency(
        self,
        image: np.ndarray,
        face_region: Optional[Tuple[int, int, int, int]]
    ) -> float:
        """Analyze color consistency between regions."""
        if len(image.shape) != 3:
            return 0.0
        
        h, w, c = image.shape
        
        if face_region is None:
            face_region = (w // 4, h // 4, 3 * w // 4, 3 * h // 4)
        
        x1, y1, x2, y2 = face_region
        x1 = max(0, min(x1, w - 1))
        x2 = max(x1 + 1, min(x2, w))
        y1 = max(0, min(y1, h - 1))
        y2 = max(y1 + 1, min(y2, h))
        
        face = image[y1:y2, x1:x2]
        
        if face.size == 0:
            return 0.0
        
        # Compare face histogram to background
        bins = self.config.color_histogram_bins
        
        face_hists = []
        bg_hists = []
        
        for channel in range(c):
            face_hist, _ = np.histogram(
                face[:, :, channel].flatten(),
                bins=bins,
                range=(0, 256)
            )
            face_hist = face_hist / (face_hist.sum() + 1e-7)
            face_hists.append(face_hist)
            
            # Background (exclude face region)
            bg_mask = np.ones((h, w), dtype=bool)
            bg_mask[y1:y2, x1:x2] = False
            
            bg_pixels = image[:, :, channel][bg_mask]
            if len(bg_pixels) > 0:
                bg_hist, _ = np.histogram(bg_pixels, bins=bins, range=(0, 256))
                bg_hist = bg_hist / (bg_hist.sum() + 1e-7)
                bg_hists.append(bg_hist)
        
        if not bg_hists:
            return 0.0
        
        # Compute histogram intersection
        intersections = []
        for face_hist, bg_hist in zip(face_hists, bg_hists):
            intersection = np.sum(np.minimum(face_hist, bg_hist))
            intersections.append(intersection)
        
        avg_intersection = np.mean(intersections)
        
        # Low intersection = color mismatch (suspicious)
        score = 1.0 - avg_intersection
        
        return float(score)
    
    def _detect_compression_artifacts(self, image: np.ndarray) -> float:
        """Detect JPEG compression artifacts that may indicate manipulation."""
        if len(image.shape) == 3:
            gray = np.mean(image, axis=2).astype(np.float32)
        else:
            gray = image.astype(np.float32)
        
        h, w = gray.shape
        
        # Look for 8x8 block artifacts (JPEG)
        block_size = 8
        
        if h < block_size * 4 or w < block_size * 4:
            return 0.0
        
        # Compute block boundary differences
        boundary_diffs = []
        
        for y in range(block_size, h - block_size, block_size):
            for x in range(block_size, w - block_size, block_size):
                # Horizontal boundary
                left = gray[y-1:y+1, x-block_size:x]
                right = gray[y-1:y+1, x:x+block_size]
                if left.size > 0 and right.size > 0:
                    h_diff = np.abs(np.mean(left) - np.mean(right))
                    boundary_diffs.append(h_diff)
                
                # Vertical boundary
                top = gray[y-block_size:y, x-1:x+1]
                bottom = gray[y:y+block_size, x-1:x+1]
                if top.size > 0 and bottom.size > 0:
                    v_diff = np.abs(np.mean(top) - np.mean(bottom))
                    boundary_diffs.append(v_diff)
        
        if not boundary_diffs:
            return 0.0
        
        # High boundary differences at 8x8 grid = compression artifacts
        avg_diff = np.mean(boundary_diffs)
        score = min(1.0, avg_diff / 20.0)
        
        return float(score)
    
    def _detect_upsampling_artifacts(self, image: np.ndarray) -> float:
        """Detect upsampling artifacts from image resizing."""
        if len(image.shape) == 3:
            gray = np.mean(image, axis=2).astype(np.float32)
        else:
            gray = image.astype(np.float32)
        
        h, w = gray.shape
        
        if h < 32 or w < 32:
            return 0.0
        
        # Compute autocorrelation to detect periodic patterns
        # (common in naive upsampling)
        
        # Use a small region for efficiency
        region_size = min(64, min(h, w))
        start_h = (h - region_size) // 2
        start_w = (w - region_size) // 2
        region = gray[start_h:start_h+region_size, start_w:start_w+region_size]
        
        # Compute 1D autocorrelation along rows
        row_autocorr = []
        for row in region:
            row_normalized = row - np.mean(row)
            if np.std(row_normalized) > 1e-7:
                row_normalized /= np.std(row_normalized)
                autocorr = np.correlate(row_normalized, row_normalized, mode='full')
                autocorr = autocorr[len(autocorr)//2:]
                row_autocorr.append(autocorr)
        
        if not row_autocorr:
            return 0.0
        
        # Average autocorrelation
        avg_autocorr = np.mean(row_autocorr, axis=0)
        
        # Look for peaks at regular intervals (upsampling pattern)
        # Skip lag 0 (always 1)
        if len(avg_autocorr) > 10:
            peaks = avg_autocorr[2:10]
            peak_strength = np.max(np.abs(peaks))
            
            # Strong periodic pattern = upsampling artifacts
            score = min(1.0, peak_strength)
        else:
            score = 0.0
        
        return float(score)
    
    def _compute_overall_score(
        self,
        frequency: float,
        checkerboard: float,
        blending: float,
        texture: float,
        color: float,
        compression: float,
        upsampling: float
    ) -> float:
        """Compute weighted overall artifact score."""
        weights = {
            "frequency": 0.20,
            "checkerboard": 0.15,
            "blending": 0.20,
            "texture": 0.15,
            "color": 0.10,
            "compression": 0.10,
            "upsampling": 0.10,
        }
        
        score = (
            weights["frequency"] * frequency +
            weights["checkerboard"] * checkerboard +
            weights["blending"] * blending +
            weights["texture"] * texture +
            weights["color"] * color +
            weights["compression"] * compression +
            weights["upsampling"] * upsampling
        )
        
        return min(1.0, max(0.0, score))


def create_artifact_analyzer(
    config: Optional[GANDetectionConfig] = None
) -> ArtifactAnalyzer:
    """
    Factory function to create an artifact analyzer.
    
    Args:
        config: Full GAN detection config.
    
    Returns:
        Configured ArtifactAnalyzer instance.
    """
    if config is None:
        return ArtifactAnalyzer()
    return ArtifactAnalyzer(config.artifact)
