"""
Document feature extraction for trust score training.

Extracts quality and structural features from document images.
"""

import logging
from dataclasses import dataclass, field
from typing import List, Optional, Dict, Any

import numpy as np

from ml.training.config import FeatureConfig

logger = logging.getLogger(__name__)


@dataclass
class DocumentFeatures:
    """Document quality and structural features."""
    
    # Image quality metrics
    sharpness_score: float = 0.0
    brightness_score: float = 0.0
    contrast_score: float = 0.0
    noise_level: float = 0.0
    blur_score: float = 0.0
    
    # Color analysis
    saturation_score: float = 0.0
    color_uniformity: float = 0.0
    
    # Structural features
    edge_density: float = 0.0
    corner_count: int = 0
    text_region_ratio: float = 0.0
    
    # Document-specific features
    has_photo_region: bool = False
    has_mrz_region: bool = False
    has_barcode: bool = False
    document_bounds_confidence: float = 0.0
    
    # Preprocessing indicators
    orientation_corrected: bool = False
    perspective_corrected: bool = False
    
    # Overall quality
    overall_quality_score: float = 0.0
    
    def to_vector(self) -> np.ndarray:
        """Convert to feature vector."""
        return np.array([
            self.sharpness_score,
            self.brightness_score,
            self.contrast_score,
            self.noise_level,
            self.blur_score,
            self.saturation_score,
            self.color_uniformity,
            self.edge_density,
            float(self.corner_count) / 100.0,  # Normalize
            self.text_region_ratio,
            float(self.has_photo_region),
            float(self.has_mrz_region),
            float(self.has_barcode),
            self.document_bounds_confidence,
            float(self.orientation_corrected),
            float(self.perspective_corrected),
            self.overall_quality_score,
        ], dtype=np.float32)


class DocumentFeatureExtractor:
    """
    Extracts document quality features from images.
    
    Analyzes:
    - Image quality (sharpness, brightness, contrast, noise)
    - Color properties
    - Structural elements (edges, corners, text regions)
    - Document-specific regions (photo, MRZ, barcode)
    """
    
    def __init__(self, config: Optional[FeatureConfig] = None):
        """
        Initialize the document feature extractor.
        
        Args:
            config: Feature configuration
        """
        self.config = config or FeatureConfig()
        self._quality_features = self.config.doc_quality_features
    
    def extract(
        self,
        document_image: Optional[np.ndarray],
        orientation_corrected: bool = False,
        perspective_corrected: bool = False,
    ) -> DocumentFeatures:
        """
        Extract document features from an image.
        
        Args:
            document_image: Preprocessed document image
            orientation_corrected: Whether orientation was corrected
            perspective_corrected: Whether perspective was corrected
            
        Returns:
            DocumentFeatures containing quality metrics
        """
        features = DocumentFeatures(
            orientation_corrected=orientation_corrected,
            perspective_corrected=perspective_corrected,
        )
        
        if document_image is None:
            return features
        
        # Denormalize if needed for quality analysis
        image = self._denormalize_image(document_image)
        
        # Extract quality features
        if "sharpness" in self._quality_features:
            features.sharpness_score = self._compute_sharpness(image)
        
        if "brightness" in self._quality_features:
            features.brightness_score = self._compute_brightness(image)
        
        if "contrast" in self._quality_features:
            features.contrast_score = self._compute_contrast(image)
        
        if "noise_level" in self._quality_features:
            features.noise_level = self._compute_noise_level(image)
        
        if "blur_score" in self._quality_features:
            features.blur_score = self._compute_blur_score(image)
        
        # Color analysis
        features.saturation_score = self._compute_saturation(image)
        features.color_uniformity = self._compute_color_uniformity(image)
        
        # Structural features
        features.edge_density = self._compute_edge_density(image)
        features.corner_count = self._detect_corners(image)
        features.text_region_ratio = self._compute_text_region_ratio(image)
        
        # Document-specific features
        features.has_photo_region = self._detect_photo_region(image)
        features.has_mrz_region = self._detect_mrz_region(image)
        features.has_barcode = self._detect_barcode(image)
        features.document_bounds_confidence = self._compute_bounds_confidence(image)
        
        # Compute overall quality
        features.overall_quality_score = self._compute_overall_quality(features)
        
        return features
    
    def _denormalize_image(self, image: np.ndarray) -> np.ndarray:
        """Convert normalized image back to uint8 for analysis."""
        if image.dtype == np.float32:
            # Assume standard normalization
            mean = np.array([0.485, 0.456, 0.406])
            std = np.array([0.229, 0.224, 0.225])
            denorm = (image * std + mean) * 255
            return np.clip(denorm, 0, 255).astype(np.uint8)
        return image
    
    def _compute_sharpness(self, image: np.ndarray) -> float:
        """Compute image sharpness using Laplacian variance."""
        try:
            # Convert to grayscale if needed
            if len(image.shape) == 3:
                gray = np.mean(image, axis=2).astype(np.uint8)
            else:
                gray = image
            
            # Compute Laplacian
            from scipy.ndimage import laplace
            lap = laplace(gray.astype(np.float32))
            variance = np.var(lap)
            
            # Normalize to 0-1 range (empirical)
            return min(1.0, variance / 1000.0)
        except Exception:
            return 0.5
    
    def _compute_brightness(self, image: np.ndarray) -> float:
        """Compute average brightness normalized to 0-1."""
        if len(image.shape) == 3:
            gray = np.mean(image, axis=2)
        else:
            gray = image
        
        avg_brightness = np.mean(gray) / 255.0
        
        # Score based on deviation from ideal (0.5)
        deviation = abs(avg_brightness - 0.5)
        return 1.0 - min(1.0, deviation * 2)
    
    def _compute_contrast(self, image: np.ndarray) -> float:
        """Compute image contrast normalized to 0-1."""
        if len(image.shape) == 3:
            gray = np.mean(image, axis=2)
        else:
            gray = image
        
        contrast = np.std(gray) / 128.0
        return min(1.0, contrast)
    
    def _compute_noise_level(self, image: np.ndarray) -> float:
        """Estimate noise level in the image."""
        try:
            if len(image.shape) == 3:
                gray = np.mean(image, axis=2).astype(np.float32)
            else:
                gray = image.astype(np.float32)
            
            # High-pass filter to isolate noise
            from scipy.ndimage import gaussian_filter
            smoothed = gaussian_filter(gray, sigma=2)
            noise = gray - smoothed
            noise_level = np.std(noise) / 255.0
            
            return min(1.0, noise_level * 10)
        except Exception:
            return 0.5
    
    def _compute_blur_score(self, image: np.ndarray) -> float:
        """Compute blur score (higher = more blur)."""
        sharpness = self._compute_sharpness(image)
        return 1.0 - sharpness
    
    def _compute_saturation(self, image: np.ndarray) -> float:
        """Compute average saturation."""
        if len(image.shape) != 3:
            return 0.0
        
        # Simple saturation: (max - min) / max for each pixel
        max_channel = np.max(image, axis=2).astype(np.float32)
        min_channel = np.min(image, axis=2).astype(np.float32)
        
        with np.errstate(divide='ignore', invalid='ignore'):
            saturation = np.where(
                max_channel > 0,
                (max_channel - min_channel) / max_channel,
                0
            )
        
        return float(np.mean(saturation))
    
    def _compute_color_uniformity(self, image: np.ndarray) -> float:
        """Compute color uniformity (lower = more uniform)."""
        if len(image.shape) != 3:
            return 1.0
        
        # Compute variance in each channel
        variances = [np.var(image[:, :, c]) for c in range(3)]
        avg_variance = np.mean(variances) / (255 * 255)
        
        return 1.0 - min(1.0, avg_variance * 10)
    
    def _compute_edge_density(self, image: np.ndarray) -> float:
        """Compute edge density in the image."""
        try:
            if len(image.shape) == 3:
                gray = np.mean(image, axis=2).astype(np.uint8)
            else:
                gray = image
            
            # Sobel edge detection
            from scipy.ndimage import sobel
            edges_x = sobel(gray.astype(np.float32), axis=0)
            edges_y = sobel(gray.astype(np.float32), axis=1)
            edge_magnitude = np.sqrt(edges_x**2 + edges_y**2)
            
            # Threshold to get edge pixels
            threshold = np.percentile(edge_magnitude, 90)
            edge_pixels = np.sum(edge_magnitude > threshold)
            total_pixels = image.shape[0] * image.shape[1]
            
            return edge_pixels / total_pixels
        except Exception:
            return 0.1
    
    def _detect_corners(self, image: np.ndarray) -> int:
        """Detect number of corners/keypoints in the image."""
        try:
            if len(image.shape) == 3:
                gray = np.mean(image, axis=2).astype(np.uint8)
            else:
                gray = image
            
            # Simple corner detection using structure tensor
            from scipy.ndimage import sobel, gaussian_filter
            
            Ix = sobel(gray.astype(np.float32), axis=1)
            Iy = sobel(gray.astype(np.float32), axis=0)
            
            Ixx = gaussian_filter(Ix * Ix, sigma=2)
            Iyy = gaussian_filter(Iy * Iy, sigma=2)
            Ixy = gaussian_filter(Ix * Iy, sigma=2)
            
            # Harris corner response
            det = Ixx * Iyy - Ixy * Ixy
            trace = Ixx + Iyy
            k = 0.04
            response = det - k * trace * trace
            
            # Count corners above threshold
            threshold = np.percentile(response, 99)
            corners = np.sum(response > threshold)
            
            return int(corners)
        except Exception:
            return 0
    
    def _compute_text_region_ratio(self, image: np.ndarray) -> float:
        """Estimate ratio of text regions in the image."""
        try:
            edge_density = self._compute_edge_density(image)
            # Text regions tend to have high edge density
            # This is a rough approximation
            return min(1.0, edge_density * 3)
        except Exception:
            return 0.3
    
    def _detect_photo_region(self, image: np.ndarray) -> bool:
        """Detect if image contains a photo region (face area)."""
        # Simplified detection - check for smooth region in expected location
        try:
            h, w = image.shape[:2]
            # Photo typically in left portion of ID
            region = image[:, :w//3]
            
            if len(region.shape) == 3:
                gray = np.mean(region, axis=2)
            else:
                gray = region
            
            # Photo regions tend to have moderate variance
            variance = np.var(gray)
            return 500 < variance < 5000
        except Exception:
            return False
    
    def _detect_mrz_region(self, image: np.ndarray) -> bool:
        """Detect if image contains MRZ (machine readable zone)."""
        try:
            h, w = image.shape[:2]
            # MRZ typically at bottom of document
            bottom_region = image[int(h * 0.75):, :]
            
            if len(bottom_region.shape) == 3:
                gray = np.mean(bottom_region, axis=2)
            else:
                gray = bottom_region
            
            # MRZ has high contrast (black text on light background)
            contrast = np.std(gray)
            edge_density = self._compute_edge_density(bottom_region)
            
            return contrast > 50 and edge_density > 0.1
        except Exception:
            return False
    
    def _detect_barcode(self, image: np.ndarray) -> bool:
        """Detect if image contains a barcode."""
        try:
            # Barcodes have very regular edge patterns
            edge_density = self._compute_edge_density(image)
            
            # Check for vertical line patterns
            if len(image.shape) == 3:
                gray = np.mean(image, axis=2).astype(np.uint8)
            else:
                gray = image
            
            # Simple vertical edge detection
            from scipy.ndimage import sobel
            v_edges = np.abs(sobel(gray.astype(np.float32), axis=1))
            h_edges = np.abs(sobel(gray.astype(np.float32), axis=0))
            
            # Barcode: more vertical edges than horizontal
            v_ratio = np.mean(v_edges) / (np.mean(h_edges) + 1e-6)
            
            return v_ratio > 2.0 and edge_density > 0.2
        except Exception:
            return False
    
    def _compute_bounds_confidence(self, image: np.ndarray) -> float:
        """Compute confidence that document bounds are properly detected."""
        # Check for clean edges around document
        corner_count = self._detect_corners(image)
        edge_density = self._compute_edge_density(image)
        
        # Well-bounded documents tend to have clear corners
        if corner_count >= 4 and edge_density > 0.05:
            return min(1.0, corner_count / 10.0)
        return 0.5
    
    def _compute_overall_quality(self, features: DocumentFeatures) -> float:
        """Compute overall quality score from individual features."""
        scores = [
            features.sharpness_score,
            features.brightness_score,
            features.contrast_score,
            1.0 - features.noise_level,  # Lower noise is better
            1.0 - features.blur_score,    # Lower blur is better
            features.document_bounds_confidence,
        ]
        
        # Add bonuses for detected regions
        if features.has_photo_region:
            scores.append(1.0)
        if features.has_mrz_region:
            scores.append(1.0)
        
        return np.mean(scores)
    
    def get_feature_dim(self) -> int:
        """Get the dimension of the output feature vector."""
        return 17  # Number of features in to_vector()
