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
from collections.abc import Mapping
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
            from ml.text_detection.third_party.craft_text_detector import Craft
            
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
                "Install the text detection requirements: "
                "pip install -r ml/text_detection/requirements.txt"
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
            Both are 2D float arrays with values 0.0 - 1.0, shaped to match
            the input image (H, W) so that bounding boxes derived from them
            are already in input-image coordinates.
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
        
        return self._extract_score_maps(image.shape[:2], prediction_result)

    @staticmethod
    def _prediction_field(prediction_result: Any, name: str) -> Any:
        """
        Read a field from a CRAFT prediction result.
        
        ``craft_text_detector.Craft.detect_text`` returns a plain ``dict``, but
        other/older backends expose attribute-style result objects. Support both.
        """
        if isinstance(prediction_result, Mapping):
            return prediction_result.get(name)
        return getattr(prediction_result, name, None)

    # Container/key tuples accepted for the raw region and affinity score maps,
    # in priority order. The first is the contract of the vendored CRAFT build
    # (``third_party/craft_text_detector``); the second is the attribute/heatmap
    # style result older backends and test doubles produce.
    _SCORE_MAP_KEYS = (
        ("score_maps", "text_score_map", "link_score_map"),
        ("heatmaps", "text_score_heatmap", "link_score_heatmap"),
    )

    def _extract_score_maps(
        self,
        image_shape: Tuple[int, int],
        prediction_result: Any,
    ) -> Tuple[np.ndarray, np.ndarray]:
        """
        Extract raw region/affinity score maps from a CRAFT prediction result.

        CRAFT produces its score maps at the network's native heatmap resolution
        (half the resized inference canvas), so they are resampled here to the
        input image size. That keeps the coordinates of the bounding boxes
        derived from them in input-image space.

        Only *scalar* (2-D) maps qualify. ``prediction_result["heatmaps"]`` from
        the vendored build holds colourised ``uint8`` visualisations
        (``cv2.applyColorMap`` output, shape ``(H, W, 3)``), which are rejected by
        the dimensionality check and never treated as score maps.

        Args:
            image_shape: (height, width) of the input image
            prediction_result: Result of ``Craft.detect_text``

        Returns:
            Tuple of (region_scores, affinity_scores) shaped (height, width)
        """
        for container_key, region_key, affinity_key in self._SCORE_MAP_KEYS:
            container = self._prediction_field(prediction_result, container_key)
            if not isinstance(container, Mapping):
                continue

            raw_region = container.get(region_key)
            raw_affinity = container.get(affinity_key)
            if raw_region is None or raw_affinity is None:
                continue

            try:
                region_scores = np.asarray(raw_region, dtype=np.float32)
                affinity_scores = np.asarray(raw_affinity, dtype=np.float32)
            except (TypeError, ValueError):
                continue

            if region_scores.ndim != 2 or affinity_scores.ndim != 2:
                # A colourised heatmap, not a thresholds-able score map.
                continue

            return self._resample_score_maps(
                image_shape, region_scores, affinity_scores
            )

        # Backend exposes neither raw nor scalar maps: derive masks from boxes.
        logger.debug(
            "CRAFT result carried no scalar score maps; deriving masks from boxes"
        )
        return self._generate_score_maps_from_boxes(image_shape, prediction_result)

    @staticmethod
    def _resample_score_maps(
        image_shape: Tuple[int, int],
        region_scores: np.ndarray,
        affinity_scores: np.ndarray,
    ) -> Tuple[np.ndarray, np.ndarray]:
        """Resample score maps to the input image size and clamp to 0.0 - 1.0."""
        height, width = image_shape

        if region_scores.shape != (height, width) or affinity_scores.shape != (
            height,
            width,
        ):
            import cv2

            region_scores = cv2.resize(
                region_scores, (width, height), interpolation=cv2.INTER_LINEAR
            )
            affinity_scores = cv2.resize(
                affinity_scores, (width, height), interpolation=cv2.INTER_LINEAR
            )

        # Keep the documented 0.0 - 1.0 contract.
        region_scores = np.clip(region_scores, 0.0, 1.0).astype(np.float32, copy=False)
        affinity_scores = np.clip(affinity_scores, 0.0, 1.0).astype(
            np.float32, copy=False
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
        
        # Get boxes from prediction (handles both dict and attribute results)
        boxes = self._prediction_field(prediction_result, "boxes")
        if boxes is None or not isinstance(boxes, (list, tuple, np.ndarray)):
            boxes = []
        
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

        # The cached hash was derived from the unloaded weights, so it is stale
        # whether or not a model had been loaded.
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
