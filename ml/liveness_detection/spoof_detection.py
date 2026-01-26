"""
Spoof detection module.

VE-901: Liveness detection - spoof attack detection

This module provides detection for various spoofing attacks:
- Photo print attacks
- Screen display attacks (phone, tablet, monitor)
- Video replay attacks
- 2D mask attacks
- 3D mask attacks
- Deepfake attacks
"""

import logging
from dataclasses import dataclass, field
from typing import List, Optional, Tuple, Dict, Any
from enum import Enum

import numpy as np

from ml.liveness_detection.config import (
    LivenessConfig,
    SpoofDetectionConfig,
    SpoofType,
)
from ml.liveness_detection.reason_codes import (
    LivenessReasonCodes,
    ReasonCodeDetails,
)

logger = logging.getLogger(__name__)


@dataclass
class SpoofDetectionResult:
    """Result of spoof detection analysis."""
    
    is_spoof: bool
    confidence: float  # 0.0 - 1.0
    
    # Detected spoof type (if any)
    spoof_type: Optional[SpoofType] = None
    
    # Individual attack scores (0 = not detected, 1 = definitely detected)
    photo_print_score: float = 0.0
    screen_display_score: float = 0.0
    video_replay_score: float = 0.0
    mask_2d_score: float = 0.0
    mask_3d_score: float = 0.0
    deepfake_score: float = 0.0
    
    # Overall spoof score
    overall_spoof_score: float = 0.0
    
    # Reason codes
    reason_codes: List[LivenessReasonCodes] = field(default_factory=list)
    
    # Details for debugging
    details: Dict[str, Any] = field(default_factory=dict)
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "is_spoof": self.is_spoof,
            "confidence": self.confidence,
            "spoof_type": self.spoof_type.value if self.spoof_type else None,
            "photo_print_score": self.photo_print_score,
            "screen_display_score": self.screen_display_score,
            "video_replay_score": self.video_replay_score,
            "mask_2d_score": self.mask_2d_score,
            "mask_3d_score": self.mask_3d_score,
            "deepfake_score": self.deepfake_score,
            "overall_spoof_score": self.overall_spoof_score,
            "reason_codes": [rc.value for rc in self.reason_codes],
            "details": self.details,
        }


class SpoofDetector:
    """
    Detector for various spoofing attacks.
    
    This class analyzes image frames to detect presentation attacks
    including photos, screens, video replays, and masks.
    """
    
    def __init__(self, config: Optional[LivenessConfig] = None):
        """
        Initialize the spoof detector.
        
        Args:
            config: Liveness configuration. Uses defaults if not provided.
        """
        self.config = config or LivenessConfig()
        self.spoof_config = self.config.spoof
    
    def detect(
        self,
        frames: List[np.ndarray],
        face_regions: Optional[List[Tuple[int, int, int, int]]] = None,
        landmarks_sequence: Optional[List[Any]] = None
    ) -> SpoofDetectionResult:
        """
        Perform spoof detection on a sequence of frames.
        
        Args:
            frames: List of image frames (BGR format).
            face_regions: Optional list of face bounding boxes.
            landmarks_sequence: Optional sequence of facial landmarks.
        
        Returns:
            SpoofDetectionResult with detection outcome.
        """
        if not frames:
            return SpoofDetectionResult(
                is_spoof=False,
                confidence=0.0,
                reason_codes=[LivenessReasonCodes.INSUFFICIENT_FRAMES],
            )
        
        reason_codes = []
        details = {}
        
        # Detect each attack type
        photo_score, photo_details = self._detect_photo_print(frames, face_regions)
        details["photo_print"] = photo_details
        
        screen_score, screen_details = self._detect_screen_display(frames, face_regions)
        details["screen_display"] = screen_details
        
        video_score, video_details = self._detect_video_replay(frames, face_regions)
        details["video_replay"] = video_details
        
        mask_2d_score, mask_2d_details = self._detect_2d_mask(frames, face_regions)
        details["mask_2d"] = mask_2d_details
        
        mask_3d_score, mask_3d_details = self._detect_3d_mask(frames, face_regions, landmarks_sequence)
        details["mask_3d"] = mask_3d_details
        
        deepfake_score, deepfake_details = self._detect_deepfake(frames, face_regions, landmarks_sequence)
        details["deepfake"] = deepfake_details
        
        # Determine overall spoof score and type
        scores = {
            SpoofType.PHOTO_PRINT: photo_score,
            SpoofType.PHOTO_SCREEN: screen_score,
            SpoofType.VIDEO_REPLAY: video_score,
            SpoofType.MASK_2D: mask_2d_score,
            SpoofType.MASK_3D: mask_3d_score,
            SpoofType.DEEPFAKE: deepfake_score,
        }
        
        # Get highest scoring attack
        max_score = max(scores.values())
        max_type = max(scores, key=scores.get)
        
        # Overall score is max of all
        overall_score = max_score
        
        # Add reason codes based on detection
        threshold = self.spoof_config.spoof_score_threshold
        high_threshold = self.spoof_config.high_confidence_threshold
        
        if photo_score >= threshold:
            reason_codes.append(LivenessReasonCodes.PHOTO_PRINT_DETECTED)
        if screen_score >= threshold:
            reason_codes.append(LivenessReasonCodes.SCREEN_DISPLAY_DETECTED)
        if video_score >= threshold:
            reason_codes.append(LivenessReasonCodes.VIDEO_REPLAY_DETECTED)
        if mask_2d_score >= threshold:
            reason_codes.append(LivenessReasonCodes.MASK_2D_DETECTED)
        if mask_3d_score >= threshold:
            reason_codes.append(LivenessReasonCodes.MASK_3D_DETECTED)
        if deepfake_score >= threshold:
            reason_codes.append(LivenessReasonCodes.DEEPFAKE_DETECTED)
        
        if overall_score >= high_threshold:
            reason_codes.append(LivenessReasonCodes.SPOOF_HIGH_CONFIDENCE)
        
        # Determine if spoof
        is_spoof = overall_score >= threshold
        
        # Confidence
        if is_spoof:
            confidence = overall_score
        else:
            confidence = 1.0 - overall_score
        
        return SpoofDetectionResult(
            is_spoof=is_spoof,
            confidence=confidence,
            spoof_type=max_type if is_spoof else None,
            photo_print_score=photo_score,
            screen_display_score=screen_score,
            video_replay_score=video_score,
            mask_2d_score=mask_2d_score,
            mask_3d_score=mask_3d_score,
            deepfake_score=deepfake_score,
            overall_spoof_score=overall_score,
            reason_codes=reason_codes,
            details=details,
        )
    
    def _get_face_region(
        self,
        frame: np.ndarray,
        face_box: Optional[Tuple[int, int, int, int]]
    ) -> np.ndarray:
        """Extract face region from frame."""
        if face_box is not None:
            x, y, w, h = face_box
            x, y = max(0, x), max(0, y)
            return frame[y:y+h, x:x+w]
        return frame
    
    def _detect_photo_print(
        self,
        frames: List[np.ndarray],
        face_regions: Optional[List[Tuple[int, int, int, int]]]
    ) -> Tuple[float, Dict[str, Any]]:
        """
        Detect printed photo attacks.
        
        Printed photos have:
        - Unnatural texture (paper grain, ink dots)
        - Reduced color saturation
        - Flat appearance (no depth)
        - Sharp edges from paper boundaries
        """
        details = {}
        
        texture_scores = []
        saturation_scores = []
        edge_scores = []
        
        for i, frame in enumerate(frames):
            face_box = face_regions[i] if face_regions and i < len(face_regions) else None
            face = self._get_face_region(frame, face_box)
            
            if face.size == 0:
                continue
            
            # Analyze texture for paper/print patterns
            texture_score = self._analyze_print_texture(face)
            texture_scores.append(texture_score)
            
            # Analyze color saturation
            saturation_score = self._analyze_saturation(face)
            saturation_scores.append(saturation_score)
            
            # Analyze edge sharpness
            edge_score = self._analyze_edge_sharpness(face)
            edge_scores.append(edge_score)
        
        if not texture_scores:
            return 0.0, {"error": "no_frames"}
        
        avg_texture = np.mean(texture_scores)
        avg_saturation = np.mean(saturation_scores)
        avg_edge = np.mean(edge_scores)
        
        # Combine scores
        # Low saturation + unnatural texture + sharp edges = photo print
        photo_score = 0.0
        
        if avg_texture > self.spoof_config.photo_print_texture_threshold:
            photo_score += 0.4
        
        if avg_saturation < self.spoof_config.photo_print_color_saturation_min:
            photo_score += 0.3
        
        if avg_edge > self.spoof_config.photo_print_edge_sharpness_threshold:
            photo_score += 0.3
        
        details = {
            "avg_texture_score": float(avg_texture),
            "avg_saturation": float(avg_saturation),
            "avg_edge_score": float(avg_edge),
        }
        
        return min(1.0, photo_score), details
    
    def _analyze_print_texture(self, face: np.ndarray) -> float:
        """Analyze for print texture patterns."""
        if len(face.shape) == 3:
            gray = np.mean(face, axis=2).astype(np.float32)
        else:
            gray = face.astype(np.float32)
        
        # High-frequency analysis for print patterns
        # Compute Laplacian
        laplacian = np.zeros_like(gray)
        if gray.shape[0] > 2 and gray.shape[1] > 2:
            laplacian[1:-1, 1:-1] = (
                gray[:-2, 1:-1] + gray[2:, 1:-1] +
                gray[1:-1, :-2] + gray[1:-1, 2:] -
                4 * gray[1:-1, 1:-1]
            )
        
        # High variance in Laplacian suggests print patterns
        laplacian_var = np.var(laplacian)
        
        # Normalize to 0-1
        score = min(1.0, laplacian_var / 100.0)
        return score
    
    def _analyze_saturation(self, face: np.ndarray) -> float:
        """Analyze color saturation."""
        if len(face.shape) != 3 or face.shape[2] != 3:
            return 0.5  # Can't analyze grayscale
        
        # Simple saturation: (max - min) / max for each pixel
        face_float = face.astype(np.float32)
        max_rgb = np.max(face_float, axis=2)
        min_rgb = np.min(face_float, axis=2)
        
        # Avoid division by zero
        saturation = np.where(max_rgb > 0, (max_rgb - min_rgb) / max_rgb, 0)
        
        return float(np.mean(saturation))
    
    def _analyze_edge_sharpness(self, face: np.ndarray) -> float:
        """Analyze edge sharpness."""
        if len(face.shape) == 3:
            gray = np.mean(face, axis=2).astype(np.float32)
        else:
            gray = face.astype(np.float32)
        
        # Sobel-like edge detection
        if gray.shape[0] < 3 or gray.shape[1] < 3:
            return 0.0
        
        grad_x = np.abs(np.diff(gray, axis=1))
        grad_y = np.abs(np.diff(gray, axis=0))
        
        # Edge strength
        edge_strength = (np.mean(grad_x) + np.mean(grad_y)) / 2.0
        
        # Normalize
        return min(1.0, edge_strength / 50.0)
    
    def _detect_screen_display(
        self,
        frames: List[np.ndarray],
        face_regions: Optional[List[Tuple[int, int, int, int]]]
    ) -> Tuple[float, Dict[str, Any]]:
        """
        Detect screen display attacks (phone, tablet, monitor).
        
        Screens have:
        - Moire patterns from pixel grid
        - Color banding from limited color depth
        - Characteristic reflections
        - Specific frequency signatures
        """
        details = {}
        
        moire_scores = []
        banding_scores = []
        
        for i, frame in enumerate(frames):
            face_box = face_regions[i] if face_regions and i < len(face_regions) else None
            face = self._get_face_region(frame, face_box)
            
            if face.size == 0 or min(face.shape[:2]) < 32:
                continue
            
            # Detect moire patterns
            moire_score = self._detect_moire_pattern(face)
            moire_scores.append(moire_score)
            
            # Detect color banding
            banding_score = self._detect_color_banding(face)
            banding_scores.append(banding_score)
        
        if not moire_scores:
            return 0.0, {"error": "no_frames"}
        
        avg_moire = np.mean(moire_scores)
        avg_banding = np.mean(banding_scores)
        
        # Combine scores
        screen_score = 0.0
        
        if avg_moire > self.spoof_config.screen_moire_threshold:
            screen_score += 0.5
        
        if avg_banding > self.spoof_config.screen_color_banding_threshold:
            screen_score += 0.5
        
        details = {
            "avg_moire_score": float(avg_moire),
            "avg_banding_score": float(avg_banding),
        }
        
        return min(1.0, screen_score), details
    
    def _detect_moire_pattern(self, face: np.ndarray) -> float:
        """Detect moire patterns from screen displays."""
        if len(face.shape) == 3:
            gray = np.mean(face, axis=2).astype(np.float32)
        else:
            gray = face.astype(np.float32)
        
        # FFT analysis
        fft = np.fft.fft2(gray)
        fft_shift = np.fft.fftshift(fft)
        magnitude = np.abs(fft_shift)
        
        rows, cols = gray.shape
        center_r, center_c = rows // 2, cols // 2
        
        # Look for periodic peaks (moire)
        # Exclude DC component
        magnitude[center_r-2:center_r+3, center_c-2:center_c+3] = 0
        
        # Find peaks
        threshold = np.percentile(magnitude, 95)
        peak_ratio = np.sum(magnitude > threshold) / magnitude.size
        
        return min(1.0, peak_ratio * 100)
    
    def _detect_color_banding(self, face: np.ndarray) -> float:
        """Detect color banding from limited color depth."""
        if len(face.shape) != 3:
            return 0.0
        
        # Count unique colors in a region
        h, w = face.shape[:2]
        sample_region = face[h//4:3*h//4, w//4:3*w//4]
        
        # Quantize colors
        quantized = (sample_region // 8) * 8
        unique_colors = len(np.unique(quantized.reshape(-1, 3), axis=0))
        
        # Real faces have more color variation
        expected_unique = sample_region.size // 3
        color_ratio = unique_colors / expected_unique
        
        # Low ratio = banding
        if color_ratio < 0.01:
            return 0.8
        elif color_ratio < 0.05:
            return 0.4
        else:
            return 0.1
    
    def _detect_video_replay(
        self,
        frames: List[np.ndarray],
        face_regions: Optional[List[Tuple[int, int, int, int]]]
    ) -> Tuple[float, Dict[str, Any]]:
        """
        Detect video replay attacks.
        
        Video replays have:
        - Consistent motion patterns (looping)
        - Frame rate artifacts
        - Compression artifacts
        - Temporal inconsistencies
        """
        details = {}
        
        if len(frames) < 10:
            return 0.0, {"insufficient_frames": True}
        
        # Analyze temporal consistency
        consistency_score = self._analyze_temporal_consistency(frames, face_regions)
        
        # Analyze compression artifacts
        compression_score = self._analyze_compression_artifacts(frames, face_regions)
        
        # Combine scores
        video_score = 0.0
        
        if consistency_score < self.spoof_config.video_temporal_consistency_threshold:
            video_score += 0.5
        
        if compression_score > self.spoof_config.video_compression_artifact_threshold:
            video_score += 0.5
        
        details = {
            "temporal_consistency": float(consistency_score),
            "compression_score": float(compression_score),
        }
        
        return min(1.0, video_score), details
    
    def _analyze_temporal_consistency(
        self,
        frames: List[np.ndarray],
        face_regions: Optional[List[Tuple[int, int, int, int]]]
    ) -> float:
        """Analyze temporal consistency of motion."""
        if len(frames) < 5:
            return 1.0
        
        # Compute inter-frame differences
        diffs = []
        for i in range(1, len(frames)):
            prev = frames[i-1].astype(np.float32)
            curr = frames[i].astype(np.float32)
            
            if prev.shape == curr.shape:
                diff = np.mean(np.abs(curr - prev))
                diffs.append(diff)
        
        if not diffs:
            return 1.0
        
        # Natural motion has consistent but varying differences
        diff_std = np.std(diffs)
        diff_mean = np.mean(diffs)
        
        if diff_mean > 0:
            consistency = 1.0 - min(1.0, diff_std / diff_mean)
        else:
            consistency = 1.0
        
        return consistency
    
    def _analyze_compression_artifacts(
        self,
        frames: List[np.ndarray],
        face_regions: Optional[List[Tuple[int, int, int, int]]]
    ) -> float:
        """Analyze compression artifacts."""
        artifact_scores = []
        
        for frame in frames[:10]:  # Sample frames
            if len(frame.shape) == 3:
                gray = np.mean(frame, axis=2).astype(np.float32)
            else:
                gray = frame.astype(np.float32)
            
            # Look for blocking artifacts (8x8 blocks from JPEG/H.264)
            h, w = gray.shape
            block_size = 8
            
            # Compute block boundary differences
            boundary_diffs = []
            
            for y in range(block_size, h - block_size, block_size):
                row_diff = np.mean(np.abs(gray[y, :] - gray[y-1, :]))
                boundary_diffs.append(row_diff)
            
            for x in range(block_size, w - block_size, block_size):
                col_diff = np.mean(np.abs(gray[:, x] - gray[:, x-1]))
                boundary_diffs.append(col_diff)
            
            if boundary_diffs:
                artifact_scores.append(np.mean(boundary_diffs))
        
        if not artifact_scores:
            return 0.0
        
        # Normalize
        avg_artifact = np.mean(artifact_scores)
        return min(1.0, avg_artifact / 10.0)
    
    def _detect_2d_mask(
        self,
        frames: List[np.ndarray],
        face_regions: Optional[List[Tuple[int, int, int, int]]]
    ) -> Tuple[float, Dict[str, Any]]:
        """
        Detect 2D mask attacks (printed face cutouts).
        
        2D masks have:
        - Sharp edges around face
        - Unnatural skin texture
        - Lack of facial micro-movements
        - Asymmetric appearance
        """
        details = {}
        
        edge_scores = []
        texture_scores = []
        
        for i, frame in enumerate(frames):
            face_box = face_regions[i] if face_regions and i < len(face_regions) else None
            face = self._get_face_region(frame, face_box)
            
            if face.size == 0:
                continue
            
            # Analyze face boundary edges
            edge_score = self._analyze_face_boundary(face)
            edge_scores.append(edge_score)
            
            # Analyze skin texture
            texture_score = self._analyze_skin_texture(face)
            texture_scores.append(texture_score)
        
        if not edge_scores:
            return 0.0, {"error": "no_frames"}
        
        avg_edge = np.mean(edge_scores)
        avg_texture = np.mean(texture_scores)
        
        # Combine scores
        mask_score = 0.0
        
        if avg_edge > self.spoof_config.mask_2d_edge_threshold:
            mask_score += 0.5
        
        if avg_texture > self.spoof_config.mask_2d_skin_texture_threshold:
            mask_score += 0.5
        
        details = {
            "avg_edge_score": float(avg_edge),
            "avg_texture_score": float(avg_texture),
        }
        
        return min(1.0, mask_score), details
    
    def _analyze_face_boundary(self, face: np.ndarray) -> float:
        """Analyze face boundary for sharp edges (mask indicator)."""
        if len(face.shape) == 3:
            gray = np.mean(face, axis=2).astype(np.float32)
        else:
            gray = face.astype(np.float32)
        
        h, w = gray.shape
        
        # Analyze edges at face boundary
        border_width = max(5, min(h, w) // 10)
        
        # Top, bottom, left, right borders
        borders = [
            gray[:border_width, :],
            gray[-border_width:, :],
            gray[:, :border_width],
            gray[:, -border_width:],
        ]
        
        edge_strengths = []
        for border in borders:
            if border.size > 0:
                grad = np.abs(np.diff(border, axis=0 if border.shape[0] > 1 else 1))
                edge_strengths.append(np.mean(grad))
        
        if not edge_strengths:
            return 0.0
        
        return min(1.0, np.mean(edge_strengths) / 30.0)
    
    def _analyze_skin_texture(self, face: np.ndarray) -> float:
        """Analyze skin texture for naturalness."""
        if len(face.shape) == 3:
            gray = np.mean(face, axis=2).astype(np.float32)
        else:
            gray = face.astype(np.float32)
        
        # Compute local standard deviation
        h, w = gray.shape
        block_size = max(8, min(h, w) // 8)
        
        local_stds = []
        for y in range(0, h - block_size, block_size):
            for x in range(0, w - block_size, block_size):
                block = gray[y:y+block_size, x:x+block_size]
                local_stds.append(np.std(block))
        
        if not local_stds:
            return 0.0
        
        # Very uniform texture is suspicious
        avg_std = np.mean(local_stds)
        std_of_stds = np.std(local_stds)
        
        # Low variation = unnatural
        if avg_std < 5 or std_of_stds < 2:
            return 0.8
        else:
            return 0.2
    
    def _detect_3d_mask(
        self,
        frames: List[np.ndarray],
        face_regions: Optional[List[Tuple[int, int, int, int]]],
        landmarks_sequence: Optional[List[Any]]
    ) -> Tuple[float, Dict[str, Any]]:
        """
        Detect 3D mask attacks (silicone, resin masks).
        
        3D masks have:
        - Uniform depth
        - Unusual skin response to light
        - Lack of micro-expressions
        - Material-specific texture
        """
        details = {}
        
        depth_scores = []
        material_scores = []
        
        for i, frame in enumerate(frames):
            face_box = face_regions[i] if face_regions and i < len(face_regions) else None
            face = self._get_face_region(frame, face_box)
            
            if face.size == 0:
                continue
            
            # Analyze depth uniformity
            depth_score = self._analyze_depth_uniformity(face)
            depth_scores.append(depth_score)
            
            # Analyze material properties
            material_score = self._analyze_material_properties(face)
            material_scores.append(material_score)
        
        if not depth_scores:
            return 0.0, {"error": "no_frames"}
        
        avg_depth = np.mean(depth_scores)
        avg_material = np.mean(material_scores)
        
        # Check for micro-expressions if landmarks available
        micro_expr_score = 0.0
        if landmarks_sequence and len(landmarks_sequence) >= 10:
            micro_expr_score = self._analyze_micro_expressions(landmarks_sequence)
        
        # Combine scores
        mask_score = 0.0
        
        if avg_depth > self.spoof_config.mask_3d_depth_uniformity_threshold:
            mask_score += 0.35
        
        if avg_material > self.spoof_config.mask_3d_skin_response_threshold:
            mask_score += 0.35
        
        if micro_expr_score < self.spoof_config.mask_3d_micro_expression_threshold:
            mask_score += 0.30
        
        details = {
            "avg_depth_score": float(avg_depth),
            "avg_material_score": float(avg_material),
            "micro_expression_score": float(micro_expr_score),
        }
        
        return min(1.0, mask_score), details
    
    def _analyze_depth_uniformity(self, face: np.ndarray) -> float:
        """Analyze depth uniformity from shading."""
        if len(face.shape) == 3:
            gray = np.mean(face, axis=2).astype(np.float32)
        else:
            gray = face.astype(np.float32)
        
        # Compute gradients as depth proxy
        grad_x = np.diff(gray, axis=1)
        grad_y = np.diff(gray, axis=0)
        
        # Uniformity = low gradient variance
        grad_var = np.var(grad_x) + np.var(grad_y)
        
        # Normalize (high uniformity = suspicious)
        return 1.0 - min(1.0, grad_var / 500.0)
    
    def _analyze_material_properties(self, face: np.ndarray) -> float:
        """Analyze material properties from color/reflection."""
        if len(face.shape) != 3:
            return 0.0
        
        # Analyze color distribution
        face_float = face.astype(np.float32)
        
        # Check for unnatural color uniformity
        color_std = np.std(face_float, axis=(0, 1))
        avg_color_std = np.mean(color_std)
        
        # Very uniform color = synthetic material
        if avg_color_std < 20:
            return 0.8
        elif avg_color_std < 40:
            return 0.4
        else:
            return 0.1
    
    def _analyze_micro_expressions(self, landmarks_sequence: List[Any]) -> float:
        """Analyze micro-expressions from landmark movement."""
        # This would analyze subtle landmark movements
        # Real faces have involuntary micro-movements
        # For now, return a neutral score
        return 0.5
    
    def _detect_deepfake(
        self,
        frames: List[np.ndarray],
        face_regions: Optional[List[Tuple[int, int, int, int]]],
        landmarks_sequence: Optional[List[Any]]
    ) -> Tuple[float, Dict[str, Any]]:
        """
        Detect deepfake/synthetic face attacks.
        
        Deepfakes have:
        - Face boundary artifacts
        - Temporal inconsistencies
        - Unnatural blending
        - Specific GAN artifacts
        """
        details = {}
        
        boundary_scores = []
        temporal_scores = []
        
        for i, frame in enumerate(frames):
            face_box = face_regions[i] if face_regions and i < len(face_regions) else None
            face = self._get_face_region(frame, face_box)
            
            if face.size == 0:
                continue
            
            # Analyze face boundary for blending artifacts
            boundary_score = self._analyze_deepfake_boundary(face)
            boundary_scores.append(boundary_score)
        
        # Analyze temporal coherence
        if len(frames) >= 5:
            temporal_score = self._analyze_temporal_coherence(frames, face_regions)
        else:
            temporal_score = 0.0
        
        if not boundary_scores:
            return 0.0, {"error": "no_frames"}
        
        avg_boundary = np.mean(boundary_scores)
        
        # Combine scores
        deepfake_score = 0.0
        
        if avg_boundary > self.spoof_config.deepfake_face_boundary_threshold:
            deepfake_score += 0.5
        
        if temporal_score < self.spoof_config.deepfake_temporal_coherence_threshold:
            deepfake_score += 0.5
        
        details = {
            "avg_boundary_score": float(avg_boundary),
            "temporal_score": float(temporal_score),
        }
        
        return min(1.0, deepfake_score), details
    
    def _analyze_deepfake_boundary(self, face: np.ndarray) -> float:
        """Analyze face boundary for deepfake artifacts."""
        if len(face.shape) == 3:
            gray = np.mean(face, axis=2).astype(np.float32)
        else:
            gray = face.astype(np.float32)
        
        h, w = gray.shape
        
        # Look for blending artifacts at face edge
        # Deepfakes often have smooth transitions at wrong locations
        
        # Create distance from center
        y, x = np.ogrid[:h, :w]
        center_y, center_x = h // 2, w // 2
        dist_from_center = np.sqrt((x - center_x)**2 + (y - center_y)**2)
        
        # Normalize distances
        max_dist = np.sqrt(center_x**2 + center_y**2)
        norm_dist = dist_from_center / max_dist
        
        # Check gradient correlation with distance
        grad = np.abs(np.diff(gray, axis=1))
        grad_padded = np.zeros_like(gray)
        grad_padded[:, :-1] = grad
        
        # Correlation
        outer_mask = norm_dist > 0.7
        if np.sum(outer_mask) > 0:
            outer_grad = np.mean(grad_padded[outer_mask])
            inner_mask = norm_dist < 0.5
            inner_grad = np.mean(grad_padded[inner_mask])
            
            # Deepfakes often have higher gradients at face boundary
            if inner_grad > 0:
                ratio = outer_grad / inner_grad
                return min(1.0, max(0.0, ratio - 1.0))
        
        return 0.0
    
    def _analyze_temporal_coherence(
        self,
        frames: List[np.ndarray],
        face_regions: Optional[List[Tuple[int, int, int, int]]]
    ) -> float:
        """Analyze temporal coherence for deepfake detection."""
        if len(frames) < 5:
            return 1.0
        
        # Compare consecutive frames
        coherence_scores = []
        
        for i in range(1, len(frames)):
            prev = frames[i-1].astype(np.float32)
            curr = frames[i].astype(np.float32)
            
            if prev.shape != curr.shape:
                continue
            
            # Compute structural similarity (simplified)
            diff = np.mean(np.abs(curr - prev))
            
            # Normalize
            score = 1.0 - min(1.0, diff / 50.0)
            coherence_scores.append(score)
        
        if not coherence_scores:
            return 1.0
        
        return np.mean(coherence_scores)
