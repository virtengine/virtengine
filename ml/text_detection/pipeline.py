"""
Main text detection pipeline.

This module orchestrates the complete text detection workflow:
1. Run CRAFT inference to get region and affinity maps
2. Extract character bounding boxes from score maps
3. Apply thresholding and NMS
4. Merge characters to words
5. Group words to lines
6. Return versioned, reproducible results
"""

import logging
import time
import hashlib
from typing import Optional, List

import numpy as np

from ml.text_detection.config import TextDetectionConfig
from ml.text_detection.roi_types import TextROI, TextDetectionResult, TextType
from ml.text_detection.craft_detector import CRAFTDetector
from ml.text_detection.postprocessing import TextPostProcessor

logger = logging.getLogger(__name__)

# Suite version for tracking
SUITE_VERSION = "1.0.0"


class TextDetectionPipeline:
    """
    Complete text detection pipeline.
    
    This class orchestrates CRAFT detection and post-processing
    to produce structured, versioned text ROI outputs suitable
    for OCR and document analysis.
    
    Processing Steps:
    1. Run CRAFT inference - Get region and affinity score maps
    2. Extract character boxes - Connected component analysis
    3. Apply thresholds - Filter low-confidence detections
    4. Character NMS - Remove overlapping character boxes
    5. Merge to words - Group characters by affinity
    6. Group to lines - Vertical alignment grouping
    7. Version output - Record model version and thresholds
    """
    
    def __init__(self, config: Optional[TextDetectionConfig] = None):
        """
        Initialize the text detection pipeline.
        
        Args:
            config: Pipeline configuration. Uses defaults if not provided.
        """
        self.config = config or TextDetectionConfig()
        
        # Initialize components
        self.detector = CRAFTDetector(self.config.craft)
        self.postprocessor = TextPostProcessor(self.config.postprocessing)
    
    def detect(
        self,
        image: np.ndarray,
        compute_hash: Optional[bool] = None,
    ) -> TextDetectionResult:
        """
        Run complete text detection pipeline on an image.
        
        This is the main entry point for text detection.
        
        Args:
            image: Input image as numpy array (BGR or RGB format)
            compute_hash: Whether to compute image hash (overrides config)
            
        Returns:
            TextDetectionResult with all detected ROIs and metadata
        """
        start_time = time.perf_counter()
        
        # Validate input
        if image is None:
            return self._error_result(
                "Input image is None",
                start_time
            )
        
        if not isinstance(image, np.ndarray):
            return self._error_result(
                f"Expected numpy array, got {type(image).__name__}",
                start_time
            )
        
        if image.size == 0:
            return self._error_result(
                "Input image is empty",
                start_time
            )
        
        # Get image info
        if len(image.shape) == 2:
            height, width = image.shape
        else:
            height, width = image.shape[:2]
        image_size = (width, height)
        
        # Compute image hash
        should_hash = compute_hash if compute_hash is not None else self.config.compute_hash
        image_hash = self._compute_image_hash(image) if should_hash else ""
        
        try:
            # Step 1: Run CRAFT detection
            logger.debug("Running CRAFT inference...")
            region_scores, affinity_scores = self.detector.detect(image)
            
            # Step 2: Extract character boxes
            logger.debug("Extracting character boxes...")
            char_boxes = self.detector.get_bounding_boxes(
                region_scores,
                affinity_scores,
            )
            logger.debug(f"Found {len(char_boxes)} character boxes")
            
            # Step 3-6: Post-processing
            logger.debug("Applying post-processing...")
            char_boxes, word_boxes, line_boxes = self.postprocessor.process(
                char_boxes,
                affinity_map=affinity_scores,
                apply_nms=True,
                merge_to_words=self.config.return_word_boxes,
                group_to_lines=self.config.return_line_boxes,
            )
            
            logger.debug(
                f"After post-processing: {len(char_boxes)} chars, "
                f"{len(word_boxes)} words, {len(line_boxes)} lines"
            )
            
            # Combine all ROIs
            all_rois: List[TextROI] = []
            if self.config.return_character_boxes:
                all_rois.extend(char_boxes)
            if self.config.return_word_boxes:
                all_rois.extend(word_boxes)
            if self.config.return_line_boxes:
                all_rois.extend(line_boxes)
            
            # Build thresholds record
            thresholds_used = {
                "text_threshold": self.config.craft.text_threshold,
                "link_threshold": self.config.craft.link_threshold,
                "low_text_threshold": self.config.craft.low_text_threshold,
                "nms_iou_threshold": self.config.postprocessing.nms_iou_threshold,
                "min_confidence": self.config.postprocessing.min_confidence,
            }
            
            # Calculate processing time
            processing_time = (time.perf_counter() - start_time) * 1000
            
            return TextDetectionResult(
                image_hash=image_hash,
                rois=all_rois,
                model_version=self.detector.model_version,
                processing_time_ms=processing_time,
                thresholds_used=thresholds_used,
                success=True,
                image_size=image_size,
                suite_version=SUITE_VERSION,
                region_score_map=region_scores,
                affinity_score_map=affinity_scores,
            )
            
        except Exception as e:
            logger.exception(f"Text detection failed: {e}")
            return self._error_result(
                f"Detection failed: {str(e)}",
                start_time,
                image_hash=image_hash,
                image_size=image_size,
            )
    
    def detect_characters_only(
        self,
        image: np.ndarray,
    ) -> TextDetectionResult:
        """
        Detect only character-level ROIs (skip word/line grouping).
        
        Args:
            image: Input image
            
        Returns:
            TextDetectionResult with character ROIs only
        """
        # Temporarily modify config
        original_word = self.config.return_word_boxes
        original_line = self.config.return_line_boxes
        
        self.config.return_word_boxes = False
        self.config.return_line_boxes = False
        
        try:
            result = self.detect(image)
        finally:
            self.config.return_word_boxes = original_word
            self.config.return_line_boxes = original_line
        
        return result
    
    def detect_words_only(
        self,
        image: np.ndarray,
    ) -> TextDetectionResult:
        """
        Detect only word-level ROIs.
        
        Args:
            image: Input image
            
        Returns:
            TextDetectionResult with word ROIs only
        """
        original_char = self.config.return_character_boxes
        original_line = self.config.return_line_boxes
        
        self.config.return_character_boxes = False
        self.config.return_line_boxes = False
        
        try:
            result = self.detect(image)
        finally:
            self.config.return_character_boxes = original_char
            self.config.return_line_boxes = original_line
        
        return result
    
    def detect_lines_only(
        self,
        image: np.ndarray,
    ) -> TextDetectionResult:
        """
        Detect only line-level ROIs.
        
        Args:
            image: Input image
            
        Returns:
            TextDetectionResult with line ROIs only
        """
        original_char = self.config.return_character_boxes
        original_word = self.config.return_word_boxes
        
        self.config.return_character_boxes = False
        self.config.return_word_boxes = False
        
        try:
            result = self.detect(image)
        finally:
            self.config.return_character_boxes = original_char
            self.config.return_word_boxes = original_word
        
        return result
    
    def _compute_image_hash(self, image: np.ndarray) -> str:
        """Compute SHA256 hash of image data."""
        return hashlib.sha256(image.tobytes()).hexdigest()
    
    def _error_result(
        self,
        message: str,
        start_time: float,
        image_hash: str = "",
        image_size: tuple = (0, 0),
    ) -> TextDetectionResult:
        """Create an error result."""
        processing_time = (time.perf_counter() - start_time) * 1000
        
        return TextDetectionResult(
            image_hash=image_hash,
            rois=[],
            model_version=self.detector.model_version,
            processing_time_ms=processing_time,
            thresholds_used={},
            success=False,
            error_message=message,
            image_size=image_size,
            suite_version=SUITE_VERSION,
        )
    
    def get_score_maps(
        self,
        image: np.ndarray,
    ) -> tuple:
        """
        Get raw score maps without post-processing.
        
        Useful for visualization and debugging.
        
        Args:
            image: Input image
            
        Returns:
            Tuple of (region_scores, affinity_scores)
        """
        return self.detector.detect(image)
    
    def unload(self) -> None:
        """Unload model to free memory."""
        self.detector.unload()
    
    def __del__(self):
        """Cleanup on deletion."""
        try:
            self.unload()
        except Exception:
            pass
