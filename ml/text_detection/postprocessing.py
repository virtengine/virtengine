"""
Post-processing utilities for text detection.

This module provides post-processing operations for text ROIs:
- Score thresholding
- Non-maximum suppression (NMS)
- Character to word merging
- Word to line grouping
"""

import logging
from typing import List, Optional, Tuple
from dataclasses import dataclass

import numpy as np

from ml.text_detection.config import PostProcessingConfig
from ml.text_detection.roi_types import TextROI, BoundingBox, Point, TextType

logger = logging.getLogger(__name__)


class TextPostProcessor:
    """
    Post-processor for text detection results.
    
    Handles thresholding, NMS, and hierarchical grouping
    of text ROIs from character to word to line level.
    """
    
    def __init__(self, config: Optional[PostProcessingConfig] = None):
        """
        Initialize the post-processor.
        
        Args:
            config: Post-processing configuration. Uses defaults if not provided.
        """
        self.config = config or PostProcessingConfig()
    
    def threshold_scores(
        self,
        scores: np.ndarray,
        threshold: float,
    ) -> np.ndarray:
        """
        Apply thresholding to a score map.
        
        Args:
            scores: Input score map (2D array)
            threshold: Threshold value (0.0 - 1.0)
            
        Returns:
            Binary mask where scores >= threshold
        """
        if scores is None or scores.size == 0:
            return np.array([], dtype=np.float32)
        
        return (scores >= threshold).astype(np.float32)
    
    def filter_by_confidence(
        self,
        boxes: List[TextROI],
        min_confidence: Optional[float] = None,
    ) -> List[TextROI]:
        """
        Filter boxes by minimum confidence.
        
        Args:
            boxes: List of TextROI objects
            min_confidence: Minimum confidence threshold
            
        Returns:
            Filtered list of TextROI objects
        """
        threshold = min_confidence or self.config.min_confidence
        return [box for box in boxes if box.confidence >= threshold]
    
    def filter_by_size(
        self,
        boxes: List[TextROI],
        min_width: Optional[int] = None,
        min_height: Optional[int] = None,
    ) -> List[TextROI]:
        """
        Filter boxes by minimum size.
        
        Args:
            boxes: List of TextROI objects
            min_width: Minimum width in pixels
            min_height: Minimum height in pixels
            
        Returns:
            Filtered list of TextROI objects
        """
        # Determine size thresholds based on text type
        filtered = []
        for box in boxes:
            if box.text_type == TextType.CHARACTER:
                w_thresh = min_width or self.config.min_char_width
                h_thresh = min_height or self.config.min_char_height
            elif box.text_type == TextType.WORD:
                w_thresh = min_width or self.config.min_word_width
                h_thresh = min_height or self.config.min_char_height
            else:
                w_thresh = min_width or self.config.min_line_width
                h_thresh = min_height or self.config.min_char_height
            
            if box.bounding_box.width >= w_thresh and box.bounding_box.height >= h_thresh:
                filtered.append(box)
        
        return filtered
    
    def non_max_suppression(
        self,
        boxes: List[TextROI],
        iou_threshold: Optional[float] = None,
    ) -> List[TextROI]:
        """
        Apply non-maximum suppression to remove overlapping boxes.
        
        Args:
            boxes: List of TextROI objects
            iou_threshold: IoU threshold for suppression
            
        Returns:
            List of TextROI objects after NMS
        """
        if not boxes:
            return []
        
        threshold = iou_threshold or self.config.nms_iou_threshold
        
        # Sort by confidence descending
        sorted_boxes = sorted(boxes, key=lambda x: x.confidence, reverse=True)
        
        keep = []
        suppressed = set()
        
        for i, box in enumerate(sorted_boxes):
            if i in suppressed:
                continue
            
            keep.append(box)
            
            # Check overlap with remaining boxes
            for j in range(i + 1, len(sorted_boxes)):
                if j in suppressed:
                    continue
                
                iou = box.bounding_box.iou(sorted_boxes[j].bounding_box)
                if iou >= threshold:
                    suppressed.add(j)
        
        return keep
    
    def merge_character_boxes(
        self,
        char_boxes: List[TextROI],
        affinity_map: Optional[np.ndarray] = None,
    ) -> List[TextROI]:
        """
        Merge character boxes into word boxes using spatial proximity.
        
        Uses affinity map if available, otherwise uses spatial heuristics.
        
        Args:
            char_boxes: List of character-level TextROI objects
            affinity_map: Optional affinity score map from CRAFT
            
        Returns:
            List of word-level TextROI objects
        """
        if not char_boxes:
            return []
        
        # Sort characters by x position (left to right)
        sorted_chars = sorted(char_boxes, key=lambda x: x.bounding_box.x)
        
        # Group characters into words
        word_groups: List[List[TextROI]] = []
        current_group: List[TextROI] = []
        
        for char in sorted_chars:
            if not current_group:
                current_group.append(char)
                continue
            
            # Check if this character should be merged with current group
            last_char = current_group[-1]
            
            # Calculate distance thresholds
            avg_width = (last_char.bounding_box.width + char.bounding_box.width) / 2
            avg_height = (last_char.bounding_box.height + char.bounding_box.height) / 2
            
            x_threshold = avg_width * self.config.char_merge_x_threshold
            y_threshold = avg_height * self.config.char_merge_y_threshold
            
            # Calculate distances
            x_gap = char.bounding_box.x - last_char.bounding_box.x2
            y_diff = abs(char.bounding_box.center.y - last_char.bounding_box.center.y)
            
            # Check affinity if available
            should_merge = False
            if affinity_map is not None:
                # Check affinity between characters
                mid_x = int((last_char.bounding_box.x2 + char.bounding_box.x) / 2)
                mid_y = int((last_char.bounding_box.center.y + char.bounding_box.center.y) / 2)
                
                if 0 <= mid_y < affinity_map.shape[0] and 0 <= mid_x < affinity_map.shape[1]:
                    affinity = affinity_map[mid_y, mid_x]
                    should_merge = affinity >= self.config.affinity_score_threshold
            
            # Fallback to spatial heuristics
            if not should_merge:
                should_merge = (
                    0 <= x_gap <= x_threshold * 2 and
                    y_diff <= y_threshold
                )
            
            if should_merge:
                current_group.append(char)
            else:
                # Start new word
                if current_group:
                    word_groups.append(current_group)
                current_group = [char]
        
        # Don't forget last group
        if current_group:
            word_groups.append(current_group)
        
        # Convert groups to word ROIs
        word_boxes = []
        for group in word_groups:
            if not group:
                continue
            
            word_box = self._create_merged_roi(group, TextType.WORD)
            word_boxes.append(word_box)
        
        return word_boxes
    
    def group_words_to_lines(
        self,
        word_boxes: List[TextROI],
    ) -> List[TextROI]:
        """
        Group word boxes into line boxes.
        
        Args:
            word_boxes: List of word-level TextROI objects
            
        Returns:
            List of line-level TextROI objects
        """
        if not word_boxes:
            return []
        
        # Sort words by y position (top to bottom), then x
        sorted_words = sorted(
            word_boxes,
            key=lambda x: (x.bounding_box.y, x.bounding_box.x)
        )
        
        # Group words into lines based on vertical overlap
        line_groups: List[List[TextROI]] = []
        current_line: List[TextROI] = []
        
        for word in sorted_words:
            if not current_line:
                current_line.append(word)
                continue
            
            # Check vertical overlap with current line
            line_y_center = sum(w.bounding_box.center.y for w in current_line) / len(current_line)
            line_avg_height = sum(w.bounding_box.height for w in current_line) / len(current_line)
            
            y_diff = abs(word.bounding_box.center.y - line_y_center)
            y_threshold = line_avg_height * self.config.line_y_threshold
            
            # Check horizontal gap
            rightmost = max(current_line, key=lambda w: w.bounding_box.x2)
            x_gap = word.bounding_box.x - rightmost.bounding_box.x2
            x_threshold = line_avg_height * self.config.line_x_gap_threshold
            
            if y_diff <= y_threshold and x_gap <= x_threshold:
                current_line.append(word)
            else:
                # Start new line
                if current_line:
                    line_groups.append(current_line)
                current_line = [word]
        
        # Don't forget last line
        if current_line:
            line_groups.append(current_line)
        
        # Convert groups to line ROIs
        line_boxes = []
        for group in line_groups:
            if not group:
                continue
            
            line_box = self._create_merged_roi(group, TextType.LINE)
            
            # Set parent-child relationships
            for word in group:
                word.parent_roi_id = line_box.roi_id
            line_box.child_roi_ids = [w.roi_id for w in group]
            
            line_boxes.append(line_box)
        
        return line_boxes
    
    def _create_merged_roi(
        self,
        rois: List[TextROI],
        text_type: TextType,
    ) -> TextROI:
        """
        Create a merged ROI from a list of ROIs.
        
        Args:
            rois: List of ROIs to merge
            text_type: Type for the merged ROI
            
        Returns:
            Merged TextROI
        """
        if not rois:
            raise ValueError("Cannot merge empty ROI list")
        
        # Compute bounding box that encloses all ROIs
        all_points = []
        for roi in rois:
            all_points.extend(roi.polygon)
        
        merged_bbox = BoundingBox.from_points(all_points)
        
        # Compute average scores
        avg_confidence = sum(r.confidence for r in rois) / len(rois)
        avg_affinity = sum(r.affinity_score for r in rois) / len(rois)
        avg_region = sum(r.region_score for r in rois) / len(rois)
        
        # Create merged polygon (convex hull would be better, but simplified here)
        polygon = [
            Point(merged_bbox.x, merged_bbox.y),
            Point(merged_bbox.x2, merged_bbox.y),
            Point(merged_bbox.x2, merged_bbox.y2),
            Point(merged_bbox.x, merged_bbox.y2),
        ]
        
        return TextROI.create(
            bounding_box=merged_bbox,
            confidence=avg_confidence,
            text_type=text_type,
            polygon=polygon,
            affinity_score=avg_affinity,
            region_score=avg_region,
        )
    
    def process(
        self,
        char_boxes: List[TextROI],
        affinity_map: Optional[np.ndarray] = None,
        apply_nms: bool = True,
        merge_to_words: bool = True,
        group_to_lines: bool = True,
    ) -> Tuple[List[TextROI], List[TextROI], List[TextROI]]:
        """
        Apply complete post-processing pipeline.
        
        Args:
            char_boxes: Character-level ROIs from CRAFT
            affinity_map: Optional affinity score map
            apply_nms: Whether to apply NMS
            merge_to_words: Whether to merge to word level
            group_to_lines: Whether to group to line level
            
        Returns:
            Tuple of (char_boxes, word_boxes, line_boxes)
        """
        # Filter by confidence and size
        char_boxes = self.filter_by_confidence(char_boxes)
        char_boxes = self.filter_by_size(char_boxes)
        
        # Apply NMS to characters
        if apply_nms:
            char_boxes = self.non_max_suppression(char_boxes)
        
        # Merge to words
        word_boxes = []
        if merge_to_words:
            word_boxes = self.merge_character_boxes(char_boxes, affinity_map)
            if apply_nms:
                word_boxes = self.non_max_suppression(word_boxes)
        
        # Group to lines
        line_boxes = []
        if group_to_lines and word_boxes:
            line_boxes = self.group_words_to_lines(word_boxes)
            if apply_nms:
                line_boxes = self.non_max_suppression(line_boxes)
        
        return char_boxes, word_boxes, line_boxes
