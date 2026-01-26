"""
Passive liveness analysis module.

VE-901: Liveness detection - passive analysis implementation

This module provides passive liveness detection through:
- Texture analysis (LBP patterns)
- Depth cue analysis
- Motion pattern analysis
- Reflection analysis
- Moire pattern detection
- Frequency analysis
"""

import logging
from dataclasses import dataclass, field
from typing import List, Optional, Tuple, Dict, Any

import numpy as np

from ml.liveness_detection.config import (
    LivenessConfig,
    PassiveAnalysisConfig,
    PassiveFeatureType,
)
from ml.liveness_detection.reason_codes import (
    LivenessReasonCodes,
    ReasonCodeDetails,
)

logger = logging.getLogger(__name__)


@dataclass
class PassiveAnalysisResult:
    """Result of passive liveness analysis."""
    
    is_live: bool
    confidence: float  # 0.0 - 1.0
    
    # Component scores
    texture_score: float = 0.0
    depth_score: float = 0.0
    motion_score: float = 0.0
    reflection_score: float = 0.0
    moire_score: float = 0.0
    frequency_score: float = 0.0
    
    # Combined score
    combined_score: float = 0.0
    
    # Reason codes
    reason_codes: List[LivenessReasonCodes] = field(default_factory=list)
    
    # Details for debugging
    details: Dict[str, Any] = field(default_factory=dict)
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "is_live": self.is_live,
            "confidence": self.confidence,
            "texture_score": self.texture_score,
            "depth_score": self.depth_score,
            "motion_score": self.motion_score,
            "reflection_score": self.reflection_score,
            "moire_score": self.moire_score,
            "frequency_score": self.frequency_score,
            "combined_score": self.combined_score,
            "reason_codes": [rc.value for rc in self.reason_codes],
            "details": self.details,
        }


class PassiveAnalyzer:
    """
    Passive liveness analyzer.
    
    This class performs passive analysis on image frames to detect
    liveness without requiring user interaction.
    """
    
    def __init__(self, config: Optional[LivenessConfig] = None):
        """
        Initialize the passive analyzer.
        
        Args:
            config: Liveness configuration. Uses defaults if not provided.
        """
        self.config = config or LivenessConfig()
        self.passiVIRTENGINE_config = self.config.passive
    
    def analyze(
        self,
        frames: List[np.ndarray],
        face_regions: Optional[List[Tuple[int, int, int, int]]] = None
    ) -> PassiveAnalysisResult:
        """
        Perform passive liveness analysis on a sequence of frames.
        
        Args:
            frames: List of image frames (BGR format).
            face_regions: Optional list of face bounding boxes (x, y, w, h).
        
        Returns:
            PassiveAnalysisResult with detection outcome.
        """
        if not frames:
            return PassiveAnalysisResult(
                is_live=False,
                confidence=0.0,
                reason_codes=[LivenessReasonCodes.INSUFFICIENT_FRAMES],
            )
        
        reason_codes = []
        details = {}
        
        # Analyze texture
        texture_score, texture_codes, texture_details = self._analyze_texture(frames, face_regions)
        reason_codes.extend(texture_codes)
        details["texture"] = texture_details
        
        # Analyze depth cues
        depth_score, depth_codes, depth_details = self._analyze_depth(frames, face_regions)
        reason_codes.extend(depth_codes)
        details["depth"] = depth_details
        
        # Analyze motion
        motion_score, motion_codes, motion_details = self._analyze_motion(frames, face_regions)
        reason_codes.extend(motion_codes)
        details["motion"] = motion_details
        
        # Analyze reflections
        reflection_score, reflection_codes, reflection_details = self._analyze_reflection(frames, face_regions)
        reason_codes.extend(reflection_codes)
        details["reflection"] = reflection_details
        
        # Detect moire patterns
        moire_score, moire_codes, moire_details = self._detect_moire(frames, face_regions)
        reason_codes.extend(moire_codes)
        details["moire"] = moire_details
        
        # Frequency analysis
        frequency_score, freq_codes, freq_details = self._analyze_frequency(frames, face_regions)
        reason_codes.extend(freq_codes)
        details["frequency"] = freq_details
        
        # Combine scores with weights
        combined_score = (
            texture_score * self.passiVIRTENGINE_config.texture_weight +
            depth_score * self.passiVIRTENGINE_config.depth_weight +
            motion_score * self.passiVIRTENGINE_config.motion_weight +
            reflection_score * self.passiVIRTENGINE_config.reflection_weight +
            (1.0 - moire_score) * self.passiVIRTENGINE_config.moire_weight  # Invert moire (high = bad)
        )
        
        # Determine if live
        is_live = combined_score >= 0.5 and moire_score < 0.5
        
        # Compute confidence
        confidence = combined_score if is_live else 1.0 - combined_score
        
        return PassiveAnalysisResult(
            is_live=is_live,
            confidence=confidence,
            texture_score=texture_score,
            depth_score=depth_score,
            motion_score=motion_score,
            reflection_score=reflection_score,
            moire_score=moire_score,
            frequency_score=frequency_score,
            combined_score=combined_score,
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
    
    def _analyze_texture(
        self,
        frames: List[np.ndarray],
        face_regions: Optional[List[Tuple[int, int, int, int]]]
    ) -> Tuple[float, List[LivenessReasonCodes], Dict[str, Any]]:
        """
        Analyze texture using Local Binary Patterns (LBP).
        
        Real faces have natural texture variations, while photos/screens
        have more uniform or artificial texture patterns.
        """
        reason_codes = []
        details = {}
        
        lbp_variances = []
        lbp_uniformities = []
        
        for i, frame in enumerate(frames):
            face_box = face_regions[i] if face_regions and i < len(face_regions) else None
            face = self._get_face_region(frame, face_box)
            
            if face.size == 0:
                continue
            
            # Convert to grayscale
            if len(face.shape) == 3:
                gray = np.mean(face, axis=2).astype(np.uint8)
            else:
                gray = face
            
            # Compute simplified LBP
            lbp = self._compute_lbp(gray)
            
            # Compute statistics
            lbp_variance = np.var(lbp)
            lbp_uniformity = self._compute_uniformity(lbp)
            
            lbp_variances.append(lbp_variance)
            lbp_uniformities.append(lbp_uniformity)
        
        if not lbp_variances:
            return 0.0, [LivenessReasonCodes.PROCESSING_ERROR], {"error": "no_frames"}
        
        avg_variance = np.mean(lbp_variances)
        avg_uniformity = np.mean(lbp_uniformities)
        
        # Score based on texture properties
        # Real faces have moderate variance and low uniformity
        variance_score = min(1.0, avg_variance / (self.passiVIRTENGINE_config.texture_variance_threshold * 100))
        uniformity_penalty = max(0.0, (avg_uniformity - self.passiVIRTENGINE_config.texture_uniformity_threshold) / 0.6)
        
        texture_score = max(0.0, variance_score - uniformity_penalty)
        
        if texture_score < 0.5:
            reason_codes.append(LivenessReasonCodes.UNNATURAL_TEXTURE)
        
        details = {
            "avg_variance": float(avg_variance),
            "avg_uniformity": float(avg_uniformity),
            "variance_score": float(variance_score),
        }
        
        return texture_score, reason_codes, details
    
    def _compute_lbp(self, gray: np.ndarray) -> np.ndarray:
        """Compute simplified Local Binary Pattern."""
        rows, cols = gray.shape
        lbp = np.zeros((rows - 2, cols - 2), dtype=np.uint8)
        
        for i in range(1, rows - 1):
            for j in range(1, cols - 1):
                center = gray[i, j]
                code = 0
                
                # 8 neighbors
                neighbors = [
                    gray[i-1, j-1], gray[i-1, j], gray[i-1, j+1],
                    gray[i, j+1], gray[i+1, j+1], gray[i+1, j],
                    gray[i+1, j-1], gray[i, j-1]
                ]
                
                for k, neighbor in enumerate(neighbors):
                    if neighbor >= center:
                        code |= (1 << k)
                
                lbp[i-1, j-1] = code
        
        return lbp
    
    def _compute_uniformity(self, lbp: np.ndarray) -> float:
        """Compute LBP uniformity (ratio of uniform patterns)."""
        if lbp.size == 0:
            return 0.0
        
        hist, _ = np.histogram(lbp.flatten(), bins=256, range=(0, 256))
        hist = hist.astype(float) / hist.sum()
        
        # Uniformity is the sum of squared probabilities
        uniformity = np.sum(hist ** 2)
        return float(uniformity)
    
    def _analyze_depth(
        self,
        frames: List[np.ndarray],
        face_regions: Optional[List[Tuple[int, int, int, int]]]
    ) -> Tuple[float, List[LivenessReasonCodes], Dict[str, Any]]:
        """
        Analyze depth cues from monocular images.
        
        Uses gradient analysis and shading patterns to estimate
        3D structure. Photos appear flatter than real faces.
        """
        reason_codes = []
        details = {}
        
        gradient_variances = []
        depth_variations = []
        
        for i, frame in enumerate(frames):
            face_box = face_regions[i] if face_regions and i < len(face_regions) else None
            face = self._get_face_region(frame, face_box)
            
            if face.size == 0:
                continue
            
            # Convert to grayscale
            if len(face.shape) == 3:
                gray = np.mean(face, axis=2).astype(np.float32)
            else:
                gray = face.astype(np.float32)
            
            # Compute gradients
            grad_x = np.diff(gray, axis=1)
            grad_y = np.diff(gray, axis=0)
            
            # Gradient magnitude
            min_shape = (min(grad_x.shape[0], grad_y.shape[0]),
                        min(grad_x.shape[1], grad_y.shape[1]))
            grad_x = grad_x[:min_shape[0], :min_shape[1]]
            grad_y = grad_y[:min_shape[0], :min_shape[1]]
            
            grad_mag = np.sqrt(grad_x**2 + grad_y**2)
            
            gradient_variances.append(np.var(grad_mag))
            
            # Estimate depth variation from shading
            depth_var = np.std(gray) / (np.mean(gray) + 1e-6)
            depth_variations.append(depth_var)
        
        if not gradient_variances:
            return 0.0, [LivenessReasonCodes.PROCESSING_ERROR], {"error": "no_frames"}
        
        avg_gradient_var = np.mean(gradient_variances)
        avg_depth_var = np.mean(depth_variations)
        
        # Score based on depth properties
        # Real faces have more gradient variation
        gradient_score = min(1.0, avg_gradient_var / (self.passiVIRTENGINE_config.depth_gradient_threshold * 1000))
        
        # Check for flatness
        if avg_depth_var < self.passiVIRTENGINE_config.depth_min_variation:
            flatness_penalty = 0.3
            reason_codes.append(LivenessReasonCodes.FLAT_DEPTH_DETECTED)
        else:
            flatness_penalty = 0.0
        
        depth_score = max(0.0, gradient_score - flatness_penalty)
        
        details = {
            "avg_gradient_variance": float(avg_gradient_var),
            "avg_depth_variation": float(avg_depth_var),
        }
        
        return depth_score, reason_codes, details
    
    def _analyze_motion(
        self,
        frames: List[np.ndarray],
        face_regions: Optional[List[Tuple[int, int, int, int]]]
    ) -> Tuple[float, List[LivenessReasonCodes], Dict[str, Any]]:
        """
        Analyze motion patterns in the frame sequence.
        
        Real faces exhibit natural micro-movements, while photos
        are static and video replays may have unnatural motion.
        """
        reason_codes = []
        details = {}
        
        if len(frames) < self.passiVIRTENGINE_config.motion_min_frames:
            return 0.5, [], {"insufficient_frames": True}
        
        # Compute frame differences
        frame_diffs = []
        for i in range(1, len(frames)):
            prev = frames[i-1].astype(np.float32)
            curr = frames[i].astype(np.float32)
            
            # Use face regions if available
            if face_regions:
                prev_box = face_regions[i-1] if i-1 < len(face_regions) else None
                curr_box = face_regions[i] if i < len(face_regions) else None
                
                if prev_box and curr_box:
                    prev = self._get_face_region(prev, prev_box)
                    curr = self._get_face_region(curr, curr_box)
            
            # Resize to same shape if needed
            if prev.shape != curr.shape:
                min_h = min(prev.shape[0], curr.shape[0])
                min_w = min(prev.shape[1], curr.shape[1])
                prev = prev[:min_h, :min_w]
                curr = curr[:min_h, :min_w]
            
            diff = np.abs(curr - prev)
            frame_diffs.append(np.mean(diff))
        
        if not frame_diffs:
            return 0.5, [], {"no_diffs": True}
        
        avg_motion = np.mean(frame_diffs)
        motion_std = np.std(frame_diffs)
        
        # Check for motion presence
        if avg_motion < self.passiVIRTENGINE_config.motion_natural_variation_min:
            reason_codes.append(LivenessReasonCodes.NO_MOTION_DETECTED)
            motion_score = 0.2
        elif avg_motion > self.passiVIRTENGINE_config.motion_natural_variation_max:
            reason_codes.append(LivenessReasonCodes.UNNATURAL_MOTION)
            motion_score = 0.3
        else:
            # Natural motion range
            motion_score = 0.8
        
        # Check motion consistency (natural motion has variation)
        if motion_std > 0:
            consistency = min(1.0, motion_std / avg_motion)
            if consistency < self.passiVIRTENGINE_config.motion_consistency_threshold:
                motion_score = min(motion_score, 0.6)
        
        details = {
            "avg_motion": float(avg_motion),
            "motion_std": float(motion_std),
            "num_frames": len(frames),
        }
        
        return motion_score, reason_codes, details
    
    def _analyze_reflection(
        self,
        frames: List[np.ndarray],
        face_regions: Optional[List[Tuple[int, int, int, int]]]
    ) -> Tuple[float, List[LivenessReasonCodes], Dict[str, Any]]:
        """
        Analyze reflection patterns.
        
        Screens and glossy photos have different reflection properties
        than real skin.
        """
        reason_codes = []
        details = {}
        
        specular_scores = []
        
        for i, frame in enumerate(frames):
            face_box = face_regions[i] if face_regions and i < len(face_regions) else None
            face = self._get_face_region(frame, face_box)
            
            if face.size == 0:
                continue
            
            # Convert to grayscale
            if len(face.shape) == 3:
                gray = np.mean(face, axis=2)
            else:
                gray = face
            
            # Detect specular highlights (very bright spots)
            threshold = np.percentile(gray, 99)
            bright_ratio = np.sum(gray > threshold) / gray.size
            
            specular_scores.append(bright_ratio)
        
        if not specular_scores:
            return 0.5, [], {"error": "no_frames"}
        
        avg_specular = np.mean(specular_scores)
        
        # Real faces have some natural specular reflection
        # Too much or too little is suspicious
        if avg_specular > self.passiVIRTENGINE_config.reflection_specular_threshold:
            reason_codes.append(LivenessReasonCodes.REFLECTION_ANOMALY)
            reflection_score = 0.4
        elif avg_specular < 0.001:
            # Too matte (possibly printed photo)
            reflection_score = 0.5
        else:
            reflection_score = 0.8
        
        details = {
            "avg_specular": float(avg_specular),
        }
        
        return reflection_score, reason_codes, details
    
    def _detect_moire(
        self,
        frames: List[np.ndarray],
        face_regions: Optional[List[Tuple[int, int, int, int]]]
    ) -> Tuple[float, List[LivenessReasonCodes], Dict[str, Any]]:
        """
        Detect moire patterns typical of screen displays.
        
        Screens captured by cameras often show interference patterns.
        """
        reason_codes = []
        details = {}
        
        moire_scores = []
        
        for i, frame in enumerate(frames):
            face_box = face_regions[i] if face_regions and i < len(face_regions) else None
            face = self._get_face_region(frame, face_box)
            
            if face.size == 0 or min(face.shape[:2]) < 32:
                continue
            
            # Convert to grayscale
            if len(face.shape) == 3:
                gray = np.mean(face, axis=2).astype(np.float32)
            else:
                gray = face.astype(np.float32)
            
            # Apply FFT
            fft = np.fft.fft2(gray)
            fft_shift = np.fft.fftshift(fft)
            magnitude = np.abs(fft_shift)
            
            # Analyze frequency bands for moire patterns
            rows, cols = gray.shape
            center_row, center_col = rows // 2, cols // 2
            
            # Create mask for moire frequency band
            low, high = self.passiVIRTENGINE_config.moire_frequency_bands
            low = min(low, min(rows, cols) // 4)
            high = min(high, min(rows, cols) // 2)
            
            y, x = np.ogrid[:rows, :cols]
            r = np.sqrt((x - center_col)**2 + (y - center_row)**2)
            
            band_mask = (r >= low) & (r <= high)
            
            # Energy in moire band vs total
            band_energy = np.sum(magnitude[band_mask])
            total_energy = np.sum(magnitude) + 1e-6
            
            moire_ratio = band_energy / total_energy
            moire_scores.append(moire_ratio)
        
        if not moire_scores:
            return 0.0, [], {"error": "no_frames"}
        
        avg_moire = np.mean(moire_scores)
        
        # High moire score indicates screen display
        if avg_moire > self.passiVIRTENGINE_config.moire_energy_threshold:
            reason_codes.append(LivenessReasonCodes.MOIRE_PATTERN_DETECTED)
        
        details = {
            "avg_moire_score": float(avg_moire),
            "threshold": self.passiVIRTENGINE_config.moire_energy_threshold,
        }
        
        return avg_moire, reason_codes, details
    
    def _analyze_frequency(
        self,
        frames: List[np.ndarray],
        face_regions: Optional[List[Tuple[int, int, int, int]]]
    ) -> Tuple[float, List[LivenessReasonCodes], Dict[str, Any]]:
        """
        Analyze frequency distribution for anomalies.
        
        Photos and screens have different frequency characteristics
        than real faces captured by camera.
        """
        reason_codes = []
        details = {}
        
        high_freq_ratios = []
        
        for i, frame in enumerate(frames):
            face_box = face_regions[i] if face_regions and i < len(face_regions) else None
            face = self._get_face_region(frame, face_box)
            
            if face.size == 0 or min(face.shape[:2]) < 16:
                continue
            
            # Convert to grayscale
            if len(face.shape) == 3:
                gray = np.mean(face, axis=2).astype(np.float32)
            else:
                gray = face.astype(np.float32)
            
            # Apply FFT
            fft = np.fft.fft2(gray)
            fft_shift = np.fft.fftshift(fft)
            magnitude = np.abs(fft_shift)
            
            # Separate low and high frequencies
            rows, cols = gray.shape
            center_row, center_col = rows // 2, cols // 2
            
            y, x = np.ogrid[:rows, :cols]
            r = np.sqrt((x - center_col)**2 + (y - center_row)**2)
            
            low_freq_mask = r < min(rows, cols) // 8
            high_freq_mask = r > min(rows, cols) // 4
            
            low_energy = np.sum(magnitude[low_freq_mask])
            high_energy = np.sum(magnitude[high_freq_mask])
            
            if low_energy > 0:
                ratio = high_energy / (low_energy + high_energy)
            else:
                ratio = 0.5
            
            high_freq_ratios.append(ratio)
        
        if not high_freq_ratios:
            return 0.5, [], {"error": "no_frames"}
        
        avg_ratio = np.mean(high_freq_ratios)
        
        # Check for anomalies
        if avg_ratio > self.passiVIRTENGINE_config.frequency_high_threshold:
            reason_codes.append(LivenessReasonCodes.FREQUENCY_ANOMALY)
            freq_score = 0.4
        elif avg_ratio < self.passiVIRTENGINE_config.frequency_low_threshold:
            # Too smooth (possibly blurred or compressed)
            freq_score = 0.5
        else:
            freq_score = 0.8
        
        details = {
            "avg_high_freq_ratio": float(avg_ratio),
        }
        
        return freq_score, reason_codes, details
