"""
Tests for text detection post-processing module.
"""

import pytest
import numpy as np

from ml.text_detection.config import PostProcessingConfig
from ml.text_detection.postprocessing import TextPostProcessor
from ml.text_detection.roi_types import TextROI, BoundingBox, TextType, Point


class TestTextPostProcessorInit:
    """Tests for TextPostProcessor initialization."""
    
    def test_init_default_config(self):
        """Test initialization with default config."""
        processor = TextPostProcessor()
        assert processor.config is not None
        assert processor.config.nms_iou_threshold == 0.5
    
    def test_init_custom_config(self, postprocessing_config):
        """Test initialization with custom config."""
        processor = TextPostProcessor(postprocessing_config)
        assert processor.config == postprocessing_config


class TestThresholdScores:
    """Tests for score thresholding."""
    
    def test_threshold_basic(self, postprocessor):
        """Test basic thresholding."""
        scores = np.array([[0.3, 0.5, 0.7], [0.8, 0.2, 0.9]], dtype=np.float32)
        
        result = postprocessor.threshold_scores(scores, 0.5)
        
        expected = np.array([[0, 1, 1], [1, 0, 1]], dtype=np.float32)
        np.testing.assert_array_equal(result, expected)
    
    def test_threshold_empty(self, postprocessor):
        """Test thresholding empty array."""
        result = postprocessor.threshold_scores(np.array([]), 0.5)
        assert len(result) == 0
    
    def test_threshold_all_below(self, postprocessor):
        """Test when all scores are below threshold."""
        scores = np.array([[0.1, 0.2], [0.3, 0.4]], dtype=np.float32)
        result = postprocessor.threshold_scores(scores, 0.5)
        
        assert np.sum(result) == 0
    
    def test_threshold_all_above(self, postprocessor):
        """Test when all scores are above threshold."""
        scores = np.array([[0.6, 0.7], [0.8, 0.9]], dtype=np.float32)
        result = postprocessor.threshold_scores(scores, 0.5)
        
        assert np.sum(result) == 4


class TestFilterByConfidence:
    """Tests for confidence filtering."""
    
    def test_filter_removes_low_confidence(self, postprocessor, sample_text_rois):
        """Test that low confidence ROIs are removed."""
        result = postprocessor.filter_by_confidence(sample_text_rois, min_confidence=0.85)
        
        for roi in result:
            assert roi.confidence >= 0.85
    
    def test_filter_keeps_all_aboVIRTENGINE_threshold(self, postprocessor, sample_text_rois):
        """Test that all ROIs above threshold are kept."""
        result = postprocessor.filter_by_confidence(sample_text_rois, min_confidence=0.5)
        
        assert len(result) == len(sample_text_rois)
    
    def test_filter_empty_list(self, postprocessor):
        """Test filtering empty list."""
        result = postprocessor.filter_by_confidence([], min_confidence=0.5)
        assert result == []


class TestFilterBySize:
    """Tests for size filtering."""
    
    def test_filter_by_minimum_size(self, postprocessor):
        """Test that small ROIs are filtered out."""
        small_roi = TextROI.create(
            bounding_box=BoundingBox(x=0, y=0, width=3, height=3),
            confidence=0.9,
            text_type=TextType.CHARACTER,
        )
        large_roi = TextROI.create(
            bounding_box=BoundingBox(x=10, y=10, width=20, height=25),
            confidence=0.9,
            text_type=TextType.CHARACTER,
        )
        
        result = postprocessor.filter_by_size([small_roi, large_roi])
        
        assert len(result) == 1
        assert result[0].bounding_box.width == 20


class TestNonMaxSuppression:
    """Tests for non-maximum suppression."""
    
    def test_nms_removes_overlapping(self, postprocessor, overlapping_rois):
        """Test that NMS removes overlapping boxes."""
        result = postprocessor.non_max_suppression(overlapping_rois, iou_threshold=0.3)
        
        # Should keep highest confidence and non-overlapping
        assert len(result) < len(overlapping_rois)
    
    def test_nms_keeps_highest_confidence(self, postprocessor, overlapping_rois):
        """Test that NMS keeps highest confidence box."""
        result = postprocessor.non_max_suppression(overlapping_rois, iou_threshold=0.3)
        
        # First box has highest confidence (0.95)
        confidences = [r.confidence for r in result]
        assert 0.95 in confidences
    
    def test_nms_no_overlap(self, postprocessor):
        """Test NMS with non-overlapping boxes."""
        rois = [
            TextROI.create(
                bounding_box=BoundingBox(x=0, y=0, width=20, height=20),
                confidence=0.9,
                text_type=TextType.CHARACTER,
            ),
            TextROI.create(
                bounding_box=BoundingBox(x=50, y=50, width=20, height=20),
                confidence=0.85,
                text_type=TextType.CHARACTER,
            ),
        ]
        
        result = postprocessor.non_max_suppression(rois)
        
        assert len(result) == 2
    
    def test_nms_empty_list(self, postprocessor):
        """Test NMS with empty list."""
        result = postprocessor.non_max_suppression([])
        assert result == []
    
    def test_nms_single_box(self, postprocessor):
        """Test NMS with single box."""
        roi = TextROI.create(
            bounding_box=BoundingBox(x=10, y=10, width=20, height=20),
            confidence=0.9,
            text_type=TextType.CHARACTER,
        )
        
        result = postprocessor.non_max_suppression([roi])
        assert len(result) == 1


class TestMergeCharacterBoxes:
    """Tests for character to word merging."""
    
    def test_merge_adjacent_characters(self, postprocessor, sample_text_rois):
        """Test that adjacent characters are merged into words."""
        # First 3 ROIs are on the same line (y=30)
        char_boxes = sample_text_rois[:3]
        
        result = postprocessor.merge_character_boxes(char_boxes)
        
        # Should produce fewer word boxes than character boxes
        assert len(result) <= len(char_boxes)
        for roi in result:
            assert roi.text_type == TextType.WORD
    
    def test_merge_preserves_confidence(self, postprocessor, sample_text_rois):
        """Test that merged boxes have valid confidence."""
        result = postprocessor.merge_character_boxes(sample_text_rois[:3])
        
        for roi in result:
            assert 0.0 <= roi.confidence <= 1.0
    
    def test_merge_empty_list(self, postprocessor):
        """Test merging empty list."""
        result = postprocessor.merge_character_boxes([])
        assert result == []
    
    def test_merge_with_affinity_map(self, postprocessor, sample_text_rois, sample_affinity_scores):
        """Test merging with affinity map."""
        result = postprocessor.merge_character_boxes(
            sample_text_rois[:3],
            affinity_map=sample_affinity_scores,
        )
        
        assert isinstance(result, list)
        for roi in result:
            assert roi.text_type == TextType.WORD


class TestGroupWordsToLines:
    """Tests for word to line grouping."""
    
    def test_group_same_line_words(self, postprocessor):
        """Test that words on same line are grouped."""
        words = [
            TextROI.create(
                bounding_box=BoundingBox(x=10, y=50, width=40, height=20),
                confidence=0.9,
                text_type=TextType.WORD,
            ),
            TextROI.create(
                bounding_box=BoundingBox(x=60, y=52, width=50, height=20),
                confidence=0.85,
                text_type=TextType.WORD,
            ),
        ]
        
        result = postprocessor.group_words_to_lines(words)
        
        assert len(result) == 1  # One line
        assert result[0].text_type == TextType.LINE
    
    def test_group_different_lines(self, postprocessor):
        """Test that words on different lines are separated."""
        words = [
            TextROI.create(
                bounding_box=BoundingBox(x=10, y=50, width=40, height=20),
                confidence=0.9,
                text_type=TextType.WORD,
            ),
            TextROI.create(
                bounding_box=BoundingBox(x=10, y=100, width=50, height=20),  # Different line
                confidence=0.85,
                text_type=TextType.WORD,
            ),
        ]
        
        result = postprocessor.group_words_to_lines(words)
        
        assert len(result) == 2  # Two lines
    
    def test_group_sets_parent_child(self, postprocessor):
        """Test that parent-child relationships are set."""
        words = [
            TextROI.create(
                bounding_box=BoundingBox(x=10, y=50, width=40, height=20),
                confidence=0.9,
                text_type=TextType.WORD,
            ),
            TextROI.create(
                bounding_box=BoundingBox(x=60, y=52, width=50, height=20),
                confidence=0.85,
                text_type=TextType.WORD,
            ),
        ]
        
        lines = postprocessor.group_words_to_lines(words)
        
        assert len(lines) == 1
        assert len(lines[0].child_roi_ids) == 2
    
    def test_group_empty_list(self, postprocessor):
        """Test grouping empty list."""
        result = postprocessor.group_words_to_lines([])
        assert result == []


class TestProcessPipeline:
    """Tests for complete post-processing pipeline."""
    
    def test_process_returns_three_levels(self, postprocessor, sample_text_rois):
        """Test that process returns char, word, and line boxes."""
        char_boxes, word_boxes, line_boxes = postprocessor.process(
            sample_text_rois,
            apply_nms=True,
            merge_to_words=True,
            group_to_lines=True,
        )
        
        assert isinstance(char_boxes, list)
        assert isinstance(word_boxes, list)
        assert isinstance(line_boxes, list)
    
    def test_process_char_only(self, postprocessor, sample_text_rois):
        """Test process with only character output."""
        char_boxes, word_boxes, line_boxes = postprocessor.process(
            sample_text_rois,
            merge_to_words=False,
            group_to_lines=False,
        )
        
        assert len(char_boxes) > 0
        assert len(word_boxes) == 0
        assert len(line_boxes) == 0
    
    def test_process_applies_nms(self, postprocessor, overlapping_rois):
        """Test that NMS is applied in process."""
        char_boxes, _, _ = postprocessor.process(
            overlapping_rois,
            apply_nms=True,
            merge_to_words=False,
            group_to_lines=False,
        )
        
        assert len(char_boxes) < len(overlapping_rois)


class TestBoundingBoxOperations:
    """Tests for BoundingBox helper methods used in post-processing."""
    
    def test_iou_full_overlap(self):
        """Test IoU with full overlap."""
        box1 = BoundingBox(x=0, y=0, width=10, height=10)
        box2 = BoundingBox(x=0, y=0, width=10, height=10)
        
        assert box1.iou(box2) == 1.0
    
    def test_iou_no_overlap(self):
        """Test IoU with no overlap."""
        box1 = BoundingBox(x=0, y=0, width=10, height=10)
        box2 = BoundingBox(x=20, y=20, width=10, height=10)
        
        assert box1.iou(box2) == 0.0
    
    def test_iou_partial_overlap(self):
        """Test IoU with partial overlap."""
        box1 = BoundingBox(x=0, y=0, width=10, height=10)
        box2 = BoundingBox(x=5, y=5, width=10, height=10)
        
        iou = box1.iou(box2)
        assert 0.0 < iou < 1.0
    
    def test_union(self):
        """Test bounding box union."""
        box1 = BoundingBox(x=0, y=0, width=10, height=10)
        box2 = BoundingBox(x=5, y=5, width=10, height=10)
        
        union = box1.union(box2)
        
        assert union.x == 0
        assert union.y == 0
        assert union.x2 == 15
        assert union.y2 == 15
