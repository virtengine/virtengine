"""
CRAFT (Character-Region Awareness for Text detection) detector wrapper.

This module provides a wrapper around the CRAFT model for detecting
text regions in document images. It handles:
- Model loading and initialization
- Deterministic inference
- Score map generation (region and affinity)
- Bounding box extraction from score maps
"""

import os
import logging
import hashlib
from typing import Tuple, List, Optional, Any

import numpy as np

from ml.text_detection.config import CRAFTConfig, DeviceType
from ml.text_detection.roi_types import TextROI, BoundingBox, Point, TextType

logger = logging.getLogger(__name__)

# Ensure deterministic behavior
os.environ.setdefault("CUDA_VISIBLE_DEVICES", "-1")


class CRAFTDetector:
    """
    CRAFT text detector wrapper.
    
    This class wraps the CRAFT model to provide text detection
    with region and affinity score maps.
    """
    
    # Model version for tracking
    MODEL_VERSION = "2.0.0"
    
    def __init__(self, config: Optional[CRAFTConfig] = None):
        """
        Initialize the CRAFT detector.
        
        Args:
            config: CRAFT configuration. Uses defaults if not provided.
        """
        self.config = config or CRAFTConfig()
        self._model = None
        self._refiner = None
        self._model_hash: Optional[str] = None
        self._device: Optional[str] = None
        
        # Set deterministic mode if requested
        if self.config.deterministic:
            self._setup_deterministic()
    
    def _setup_deterministic(self) -> None:
        """Configure deterministic operations for reproducibility."""
        os.environ["CUBLAS_WORKSPACE_CONFIG"] = ":4096:8"
        
        try:
            import torch
            torch.use_deterministic_algorithms(True)
            if torch.cuda.is_available():
                torch.backends.cudnn.deterministic = True
                torch.backends.cudnn.benchmark = False
        except ImportError:
            logger.warning("PyTorch not available, skipping deterministic setup")
    
    @property
    def model(self) -> Any:
        """Lazy load the CRAFT model."""
        if self._model is None:
            self._model = self._load_model()
        return self._model
    
    @property
    def model_version(self) -> str:
        """Get the model version string."""
        return self.MODEL_VERSION
    
    @property
    def model_hash(self) -> str:
        """Get or compute the model weights hash."""
        if self._model_hash is None:
            self._model_hash = self._compute_model_hash()
        return self._model_hash
    
    def _load_model(self) -> Any:
        """
        Load the CRAFT model.
        
        Returns:
            Loaded CRAFT model instance
        """
        try:
            import torch
            from craft_text_detector import Craft
            
            # Determine device
            if self.config.device == DeviceType.CUDA and torch.cuda.is_available():
                self._device = "cuda"
            elif self.config.device == DeviceType.MPS and hasattr(torch.backends, "mps"):
                if torch.backends.mps.is_available():
                    self._device = "mps"
                else:
                    self._device = "cpu"
            else:
                self._device = "cpu"
            
            logger.info(f"Loading CRAFT model on device: {self._device}")
            
            # Initialize CRAFT detector
            craft = Craft(
                output_dir=None,
                crop_type="box",
                cuda=self._device == "cuda",
                text_threshold=self.config.text_threshold,
                link_threshold=self.config.link_threshold,
                low_text=self.config.low_text_threshold,
                long_size=self.config.canvas_size,
            )
            
            logger.info("CRAFT model loaded successfully")
            return craft
            
        except ImportError as e:
            logger.error(f"Failed to import CRAFT dependencies: {e}")
            raise ImportError(
                "CRAFT dependencies not installed. "
                "Please install: pip install craft-text-detector torch torchvision"
            ) from e
    
    def _compute_model_hash(self) -> str:
        """
        Compute a hash of the model weights for versioning.
        
        Returns:
            SHA256 hash of model weights
        """
        try:
            import torch
            
            # Access the underlying model
            if hasattr(self.model, 'net'):
                state_dict = self.model.net.state_dict()
                
                # Create hash from state dict
                hasher = hashlib.sha256()
                for key in sorted(state_dict.keys()):
                    param = state_dict[key].cpu().numpy()
                    hasher.update(key.encode())
                    hasher.update(param.tobytes())
                
                return hasher.hexdigest()[:16]  # Truncate for readability
            else:
                return "unknown"
                
        except Exception as e:
            logger.warning(f"Could not compute model hash: {e}")
            return "unknown"
    
    def detect(
        self,
        image: np.ndarray,
    ) -> Tuple[np.ndarray, np.ndarray]:
        """
        Run CRAFT inference to get region and affinity score maps.
        
        Args:
            image: Input image as BGR numpy array (H, W, C)
            
        Returns:
            Tuple of (region_score_map, affinity_score_map)
            Both are 2D float arrays with values 0.0 - 1.0
        """
        if image is None or image.size == 0:
            raise ValueError("Invalid input image")
        
        # Ensure 3-channel BGR image
        if len(image.shape) == 2:
            import cv2
            image = cv2.cvtColor(image, cv2.COLOR_GRAY2BGR)
        elif image.shape[2] == 4:
            import cv2
            image = cv2.cvtColor(image, cv2.COLOR_BGRA2BGR)
        
        # Run CRAFT detection
        prediction_result = self.model.detect_text(image)
        
        # Extract score maps from prediction
        if hasattr(prediction_result, 'heatmaps'):
            heatmaps = prediction_result.heatmaps
            region_scores = heatmaps.get('text_score_heatmap', np.zeros_like(image[:,:,0], dtype=np.float32))
            affinity_scores = heatmaps.get('link_score_heatmap', np.zeros_like(image[:,:,0], dtype=np.float32))
        else:
            # Fallback: generate score maps from detection results
            region_scores, affinity_scores = self._generate_score_maps_from_boxes(
                image.shape[:2],
                prediction_result
            )
        
        return region_scores, affinity_scores
    
    def _generate_score_maps_from_boxes(
        self,
        image_shape: Tuple[int, int],
        prediction_result: Any
    ) -> Tuple[np.ndarray, np.ndarray]:
        """
        Generate score maps from detected boxes when heatmaps unavailable.
        
        Args:
            image_shape: (height, width) of image
            prediction_result: CRAFT prediction result
            
        Returns:
            Tuple of (region_scores, affinity_scores)
        """
        height, width = image_shape
        region_scores = np.zeros((height, width), dtype=np.float32)
        affinity_scores = np.zeros((height, width), dtype=np.float32)
        
        # Get boxes from prediction
        boxes = []
        if hasattr(prediction_result, 'boxes'):
            boxes = prediction_result.boxes
        
        # Fill region scores based on boxes
        for box in boxes:
            if len(box) >= 4:
                # Convert box points to mask
                pts = np.array(box, dtype=np.int32)
                import cv2
                cv2.fillPoly(region_scores, [pts], 1.0)
        
        return region_scores, affinity_scores
    
    def get_bounding_boxes(
        self,
        region_scores: np.ndarray,
        affinity_scores: np.ndarray,
        text_threshold: Optional[float] = None,
        link_threshold: Optional[float] = None,
        low_text_threshold: Optional[float] = None,
    ) -> List[TextROI]:
        """
        Extract bounding boxes from score maps.
        
        Args:
            region_scores: Region score map from detect()
            affinity_scores: Affinity score map from detect()
            text_threshold: Override config text threshold
            link_threshold: Override config link threshold
            low_text_threshold: Override config low text threshold
            
        Returns:
            List of TextROI objects for detected characters
        """
        import cv2
        
        # Use config values if not overridden
        text_thresh = text_threshold or self.config.text_threshold
        link_thresh = link_threshold or self.config.link_threshold
        low_text = low_text_threshold or self.config.low_text_threshold
        
        # Threshold the score maps
        text_score = region_scores.copy()
        link_score = affinity_scores.copy()
        
        # Combined score for connected components
        text_score_comb = np.clip(text_score + link_score, 0, 1)
        
        # Binary threshold
        _, text_score_bin = cv2.threshold(
            text_score_comb, low_text, 1.0, cv2.THRESH_BINARY
        )
        text_score_bin = (text_score_bin * 255).astype(np.uint8)
        
        # Find connected components
        num_labels, labels, stats, centroids = cv2.connectedComponentsWithStats(
            text_score_bin, connectivity=4
        )
        
        rois = []
        
        # Process each component (skip background label 0)
        for label_id in range(1, num_labels):
            # Get component stats
            x, y, w, h, area = stats[label_id]
            
            # Skip very small components
            if w < 3 or h < 3:
                continue
            
            # Get mask for this component
            mask = (labels == label_id)
            
            # Calculate scores for this region
            region_values = text_score[mask]
            affinity_values = link_score[mask]
            
            if len(region_values) == 0:
                continue
            
            avg_region_score = float(np.mean(region_values))
            max_region_score = float(np.max(region_values))
            avg_affinity_score = float(np.mean(affinity_values)) if len(affinity_values) > 0 else 0.0
            
            # Skip if below text threshold
            if max_region_score < text_thresh:
                continue
            
            # Create bounding box
            bbox = BoundingBox(x=float(x), y=float(y), width=float(w), height=float(h))
            
            # Calculate confidence from region score
            confidence = min(1.0, max_region_score)
            
            # Get polygon points from contours
            mask_uint8 = mask.astype(np.uint8) * 255
            contours, _ = cv2.findContours(
                mask_uint8, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE
            )
            
            polygon = []
            if contours:
                # Get the largest contour
                largest_contour = max(contours, key=cv2.contourArea)
                # Simplify to quadrilateral
                epsilon = 0.02 * cv2.arcLength(largest_contour, True)
                approx = cv2.approxPolyDP(largest_contour, epsilon, True)
                polygon = [Point(float(pt[0][0]), float(pt[0][1])) for pt in approx]
            
            # Create ROI
            roi = TextROI.create(
                bounding_box=bbox,
                confidence=confidence,
                text_type=TextType.CHARACTER,
                polygon=polygon if polygon else None,
                affinity_score=avg_affinity_score,
                region_score=avg_region_score,
            )
            
            rois.append(roi)
        
        return rois
    
    def detect_and_get_boxes(
        self,
        image: np.ndarray,
        text_threshold: Optional[float] = None,
        link_threshold: Optional[float] = None,
    ) -> Tuple[List[TextROI], np.ndarray, np.ndarray]:
        """
        Combined detect and box extraction.
        
        Args:
            image: Input BGR image
            text_threshold: Optional override for text threshold
            link_threshold: Optional override for link threshold
            
        Returns:
            Tuple of (rois, region_scores, affinity_scores)
        """
        region_scores, affinity_scores = self.detect(image)
        rois = self.get_bounding_boxes(
            region_scores,
            affinity_scores,
            text_threshold=text_threshold,
            link_threshold=link_threshold,
        )
        return rois, region_scores, affinity_scores
    
    def unload(self) -> None:
        """Unload model to free memory."""
        if self._model is not None:
            if hasattr(self._model, 'unload_craftnet_model'):
                self._model.unload_craftnet_model()
            if hasattr(self._model, 'unload_refinenet_model'):
                self._model.unload_refinenet_model()
            self._model = None
            self._model_hash = None
        
        # Clear CUDA cache if available
        try:
            import torch
            if torch.cuda.is_available():
                torch.cuda.empty_cache()
        except ImportError:
            pass
    
    def __del__(self):
        """Cleanup on deletion."""
        self.unload()
