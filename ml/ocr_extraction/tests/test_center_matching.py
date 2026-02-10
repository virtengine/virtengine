"""
Tests for center-based matching algorithm.

Task Reference: VE-3041 - Port Center-Matching Algorithm
"""

import pytest
import numpy as np
from typing import List, Tuple

from ml.ocr_extraction.center_matching import (
    BoundingBox,
    MaskRegion,
    MatchResult,
    CenterMatcher,
    CenterMatchingConfig,
    extract_mask_regions_from_unet,
    convert_ocr_boxes,
)


class TestBoundingBox:
    """Tests for BoundingBox dataclass."""
    
    def test_center_calculation(self):
        """Test center point calculation."""
        box = BoundingBox(x=100, y=50, width=200, height=100)
        cx, cy = box.center
        assert cx == 200.0  # 100 + 200/2
        assert cy == 100.0  # 50 + 100/2
    
    def test_center_ratio(self):
        """Test center ratio calculation."""
        box = BoundingBox(x=100, y=100, width=100, height=100)
        # Center at (150, 150) on 300x300 image
        rx, ry = box.get_center_ratio(300, 300)
        assert rx == 0.5
        assert ry == 0.5
    
    def test_right_left_centers(self):
        """Test edge center calculations."""
        box = BoundingBox(x=100, y=100, width=100, height=100)
        
        # Right center at (200, 150)
        rx, ry = box.right_center
        assert rx == 200.0
        assert ry == 150.0
        
        # Left center at (100, 150)
        lx, ly = box.left_center
        assert lx == 100.0
        assert ly == 150.0
    
    def test_from_xyxy(self):
        """Test creation from corner coordinates."""
        box = BoundingBox.from_xyxy(100, 50, 300, 150)
        assert box.x == 100
        assert box.y == 50
        assert box.width == 200
        assert box.height == 100
    
    def test_from_xyxy_swapped_coords(self):
        """Test from_xyxy handles swapped coordinates."""
        box = BoundingBox.from_xyxy(300, 150, 100, 50)
        assert box.x == 100
        assert box.y == 50
        assert box.width == 200
        assert box.height == 100
    
    def test_from_polygon(self):
        """Test creation from polygon points."""
        # Rectangle as 4-point polygon
        points = [(100, 50), (300, 50), (300, 150), (100, 150)]
        box = BoundingBox.from_polygon(points)
        assert box.x == 100
        assert box.y == 50
        assert box.width == 200
        assert box.height == 100
    
    def test_area(self):
        """Test area calculation."""
        box = BoundingBox(x=0, y=0, width=100, height=50)
        assert box.area() == 5000
    
    def test_to_tuple(self):
        """Test tuple conversion."""
        box = BoundingBox(x=10, y=20, width=30, height=40)
        assert box.to_tuple() == (10, 20, 30, 40)


class TestMaskRegion:
    """Tests for MaskRegion dataclass."""
    
    def test_center(self):
        """Test mask region center from bounding box."""
        bbox = BoundingBox(x=100, y=100, width=100, height=100)
        region = MaskRegion(
            class_id=0,
            class_name="surname",
            mask=None,
            bounding_box=bbox,
            area=10000
        )
        cx, cy = region.center
        assert cx == 150.0
        assert cy == 150.0
    
    def test_center_ratio(self):
        """Test mask region center ratio."""
        bbox = BoundingBox(x=0, y=0, width=100, height=100)
        region = MaskRegion(
            class_id=0,
            class_name="test",
            mask=None,
            bounding_box=bbox,
            area=10000
        )
        rx, ry = region.get_center_ratio(200, 200)
        assert rx == 0.25
        assert ry == 0.25


class TestCenterMatchingConfig:
    """Tests for CenterMatchingConfig."""
    
    def test_default_config(self):
        """Test default configuration values."""
        config = CenterMatchingConfig()
        assert config.distance_threshold == 0.15
        assert config.merge_threshold == 0.02
        assert config.min_confidence == 0.5
        assert config.use_weighted_center is True
        assert config.neighbor_distance_pixels == 50.0
    
    def test_custom_config(self):
        """Test custom configuration."""
        config = CenterMatchingConfig(
            distance_threshold=0.2,
            merge_threshold=0.05,
            min_confidence=0.3
        )
        assert config.distance_threshold == 0.2
        assert config.merge_threshold == 0.05
        assert config.min_confidence == 0.3
    
    def test_invalid_distance_threshold(self):
        """Test validation of distance threshold."""
        with pytest.raises(ValueError):
            CenterMatchingConfig(distance_threshold=0)
        with pytest.raises(ValueError):
            CenterMatchingConfig(distance_threshold=-0.1)
    
    def test_invalid_merge_threshold(self):
        """Test validation of merge threshold."""
        with pytest.raises(ValueError):
            CenterMatchingConfig(merge_threshold=-0.01)
    
    def test_invalid_min_confidence(self):
        """Test validation of min confidence."""
        with pytest.raises(ValueError):
            CenterMatchingConfig(min_confidence=-0.1)
        with pytest.raises(ValueError):
            CenterMatchingConfig(min_confidence=1.5)


class TestCenterMatcher:
    """Tests for CenterMatcher class."""
    
    @pytest.fixture
    def matcher(self):
        """Create default matcher."""
        return CenterMatcher()
    
    @pytest.fixture
    def sample_mask_regions(self) -> List[MaskRegion]:
        """Create sample mask regions for testing."""
        return [
            MaskRegion(
                class_id=0,
                class_name="surname",
                mask=None,
                bounding_box=BoundingBox(x=50, y=50, width=100, height=30),
                area=3000
            ),
            MaskRegion(
                class_id=1,
                class_name="given_names",
                mask=None,
                bounding_box=BoundingBox(x=50, y=100, width=150, height=30),
                area=4500
            ),
            MaskRegion(
                class_id=2,
                class_name="date_of_birth",
                mask=None,
                bounding_box=BoundingBox(x=50, y=150, width=80, height=30),
                area=2400
            ),
        ]
    
    @pytest.fixture
    def sample_text_boxes(self) -> List[BoundingBox]:
        """Create sample text boxes for testing."""
        return [
            BoundingBox(x=55, y=52, width=95, height=28, label="SMITH", confidence=0.95),
            BoundingBox(x=52, y=102, width=145, height=28, label="JOHN MICHAEL", confidence=0.92),
            BoundingBox(x=48, y=148, width=78, height=28, label="15/03/1990", confidence=0.88),
            BoundingBox(x=300, y=50, width=50, height=30, label="OTHER", confidence=0.90),
        ]
    
    def test_get_center_ratios(self, matcher, sample_mask_regions):
        """Test center ratio extraction from mask regions."""
        ratios = matcher.get_center_ratios(sample_mask_regions, 400, 300)
        
        assert len(ratios) == 3
        # First region center at (100, 65) on 400x300 image
        assert ratios[0] == pytest.approx((0.25, 0.2166), rel=0.01)
    
    def test_get_box_center_ratios(self, matcher, sample_text_boxes):
        """Test center ratio extraction from bounding boxes."""
        ratios = matcher.get_box_center_ratios(sample_text_boxes, 400, 300)
        
        assert len(ratios) == 4
    
    def test_invalid_dimensions(self, matcher, sample_mask_regions):
        """Test error handling for invalid dimensions."""
        with pytest.raises(ValueError):
            matcher.get_center_ratios(sample_mask_regions, 0, 300)
        with pytest.raises(ValueError):
            matcher.get_center_ratios(sample_mask_regions, 400, -1)
    
    def test_match_centers_basic(self, matcher):
        """Test basic center matching."""
        # Mask ratios at (0.25, 0.2) and (0.75, 0.8)
        mask_ratios = [(0.25, 0.2), (0.75, 0.8)]
        # Box ratios: one close to first mask, one close to second
        box_ratios = [(0.26, 0.21), (0.74, 0.79), (0.5, 0.5)]
        
        matches = matcher.match_centers(mask_ratios, box_ratios)
        
        assert len(matches) == 2
        assert matches[0][0] == 0  # First mask matched to first box
        assert matches[1][0] == 1  # Second mask matched to second box
    
    def test_match_centers_threshold(self):
        """Test that distant boxes are not matched."""
        config = CenterMatchingConfig(distance_threshold=0.05)
        matcher = CenterMatcher(config)
        
        # Mask at (0.25, 0.25)
        mask_ratios = [(0.25, 0.25)]
        # Box at (0.5, 0.5) - too far away
        box_ratios = [(0.5, 0.5)]
        
        matches = matcher.match_centers(mask_ratios, box_ratios)
        
        assert len(matches) == 0
    
    def test_match_centers_empty_boxes(self, matcher):
        """Test matching with no boxes."""
        mask_ratios = [(0.25, 0.25)]
        box_ratios = []
        
        matches = matcher.match_centers(mask_ratios, box_ratios)
        
        assert len(matches) == 0
    
    def test_match_masks_to_boxes(self, matcher, sample_mask_regions, sample_text_boxes):
        """Test full matching pipeline."""
        results = matcher.match_masks_to_boxes(
            sample_mask_regions,
            sample_text_boxes,
            img_width=400,
            img_height=300
        )
        
        assert len(results) == 3
        # All should be matched (boxes are close to masks)
        assert all(r.is_matched for r in results)
        
        # Check matched labels
        assert results[0].matched_box.label == "SMITH"
        assert results[1].matched_box.label == "JOHN MICHAEL"
        assert results[2].matched_box.label == "15/03/1990"
    
    def test_match_masks_to_boxes_confidence_filter(self, sample_mask_regions):
        """Test confidence filtering."""
        config = CenterMatchingConfig(min_confidence=0.95)
        matcher = CenterMatcher(config)
        
        # Only one box above 0.95 confidence
        boxes = [
            BoundingBox(x=55, y=52, width=95, height=28, label="HIGH", confidence=0.96),
            BoundingBox(x=52, y=102, width=145, height=28, label="LOW", confidence=0.5),
        ]
        
        results = matcher.match_masks_to_boxes(
            sample_mask_regions[:2],
            boxes,
            img_width=400,
            img_height=300
        )
        
        # Only first should match (second box filtered out)
        matched_count = sum(1 for r in results if r.is_matched)
        assert matched_count == 1
        assert results[0].matched_box.label == "HIGH"
    
    def test_merge_adjacent_boxes(self, matcher):
        """Test horizontal box merging."""
        boxes = [
            BoundingBox(x=100, y=100, width=50, height=30, label="JOHN"),
            BoundingBox(x=155, y=100, width=80, height=30, label="MICHAEL"),  # 5px gap
        ]
        
        merged = matcher.merge_adjacent_boxes(boxes, img_width=400, img_height=300)
        
        assert len(merged) == 1
        assert merged[0].label == "JOHN MICHAEL"
        assert merged[0].x == 100
        assert merged[0].width == 135  # 155 + 80 - 100
    
    def test_merge_non_adjacent_boxes(self, matcher):
        """Test that distant boxes are not merged."""
        boxes = [
            BoundingBox(x=100, y=100, width=50, height=30, label="FIRST"),
            BoundingBox(x=250, y=100, width=50, height=30, label="SECOND"),  # 100px gap
        ]
        
        merged = matcher.merge_adjacent_boxes(boxes, img_width=400, img_height=300)
        
        assert len(merged) == 2
    
    def test_merge_vertically_misaligned_boxes(self, matcher):
        """Test that vertically misaligned boxes are not merged."""
        boxes = [
            BoundingBox(x=100, y=100, width=50, height=30, label="TOP"),
            BoundingBox(x=155, y=200, width=50, height=30, label="BOTTOM"),  # Different row
        ]
        
        merged = matcher.merge_adjacent_boxes(boxes, img_width=400, img_height=300)
        
        assert len(merged) == 2
    
    def test_find_adjacent_boxes_right(self, matcher):
        """Test finding right adjacent boxes."""
        target = BoundingBox(x=100, y=100, width=50, height=30, label="TARGET")
        all_boxes = [
            target,
            BoundingBox(x=160, y=100, width=50, height=30, label="RIGHT1"),  # 10px gap
            BoundingBox(x=220, y=100, width=50, height=30, label="FAR"),  # Too far
        ]
        
        neighbors = matcher.find_adjacent_boxes(target, all_boxes, 400, direction="right")
        
        assert len(neighbors) == 1
        assert neighbors[0].label == "RIGHT1"
    
    def test_find_adjacent_boxes_left(self, matcher):
        """Test finding left adjacent boxes."""
        target = BoundingBox(x=200, y=100, width=50, height=30, label="TARGET")
        all_boxes = [
            BoundingBox(x=140, y=100, width=50, height=30, label="LEFT1"),  # 10px gap
            target,
        ]
        
        neighbors = matcher.find_adjacent_boxes(target, all_boxes, 400, direction="left")
        
        assert len(neighbors) == 1
        assert neighbors[0].label == "LEFT1"
    
    def test_extend_box_with_neighbors(self, matcher):
        """Test box extension with neighbors."""
        target = BoundingBox(x=200, y=100, width=50, height=30, label="MIDDLE")
        all_boxes = [
            BoundingBox(x=140, y=100, width=50, height=30, label="LEFT"),
            target,
            BoundingBox(x=260, y=100, width=50, height=30, label="RIGHT"),
        ]
        
        extended = matcher.extend_box_with_neighbors(target, all_boxes, 400)
        
        assert extended.x == 140
        assert extended.width == 170  # 140 to 310
        assert extended.label == "LEFT MIDDLE RIGHT"
    
    def test_determinism(self, matcher, sample_mask_regions, sample_text_boxes):
        """Verify same input produces same output (determinism for consensus)."""
        results1 = matcher.match_masks_to_boxes(
            sample_mask_regions, sample_text_boxes, 400, 300
        )
        results2 = matcher.match_masks_to_boxes(
            sample_mask_regions, sample_text_boxes, 400, 300
        )
        
        assert len(results1) == len(results2)
        for r1, r2 in zip(results1, results2):
            assert r1.distance == r2.distance
            assert r1.confidence == r2.confidence
            if r1.matched_box and r2.matched_box:
                assert r1.matched_box.label == r2.matched_box.label
    
    def test_weighted_centroid(self, matcher):
        """Test mask-weighted centroid calculation."""
        # Create a mask with more pixels in bottom-right
        mask = np.zeros((100, 100), dtype=np.uint8)
        mask[60:80, 60:80] = 1  # Bottom-right quadrant
        
        cx, cy = matcher._weighted_centroid(mask)
        
        # Centroid should be in bottom-right area
        assert cx > 50
        assert cy > 50
    
    def test_weighted_centroid_empty_mask(self, matcher):
        """Test weighted centroid with empty mask."""
        mask = np.zeros((100, 100), dtype=np.uint8)
        
        cx, cy = matcher._weighted_centroid(mask)
        
        # Should return center of image
        assert cx == 50.0
        assert cy == 50.0
    
    def test_euclidean_distance(self, matcher):
        """Test Euclidean distance calculation."""
        d = matcher._euclidean_distance((0, 0), (3, 4))
        assert d == 5.0
    
    def test_vertical_overlap_ratio(self, matcher):
        """Test vertical overlap ratio calculation."""
        box1 = BoundingBox(x=0, y=100, width=100, height=50)
        box2 = BoundingBox(x=110, y=110, width=100, height=50)  # 40px overlap
        
        ratio = matcher._vertical_overlap_ratio(box1, box2)
        
        assert ratio == pytest.approx(0.8)  # 40/50
    
    def test_vertical_overlap_no_overlap(self, matcher):
        """Test vertical overlap with no overlap."""
        box1 = BoundingBox(x=0, y=100, width=100, height=50)
        box2 = BoundingBox(x=110, y=200, width=100, height=50)  # No overlap
        
        ratio = matcher._vertical_overlap_ratio(box1, box2)
        
        assert ratio == 0.0


class TestHelperFunctions:
    """Tests for module-level helper functions."""
    
    def test_convert_ocr_boxes(self):
        """Test conversion of OCR detector output."""
        # Simulated CRAFT output: 4-point polygons
        regions = np.array([
            [[100, 50], [200, 50], [200, 100], [100, 100]],
            [[100, 150], [250, 150], [250, 200], [100, 200]],
        ])
        labels = ["TEXT1", "TEXT2"]
        confidences = [0.95, 0.88]
        
        boxes, centers = convert_ocr_boxes(regions, labels, confidences)
        
        assert len(boxes) == 2
        assert len(centers) == 2
        
        assert boxes[0].x == 100
        assert boxes[0].width == 100
        assert boxes[0].label == "TEXT1"
        assert boxes[0].confidence == 0.95
        
        assert centers[0] == (150.0, 75.0)
    
    def test_convert_ocr_boxes_no_labels(self):
        """Test conversion without labels."""
        regions = np.array([
            [[100, 50], [200, 50], [200, 100], [100, 100]],
        ])
        
        boxes, centers = convert_ocr_boxes(regions)
        
        assert len(boxes) == 1
        assert boxes[0].label is None
        assert boxes[0].confidence == 1.0


class TestExtractMaskRegions:
    """Tests for mask region extraction (requires cv2)."""
    
    @pytest.fixture
    def sample_mask(self) -> np.ndarray:
        """Create sample binary mask with 3 regions."""
        mask = np.zeros((200, 300), dtype=np.uint8)
        # Three horizontal bars
        mask[20:40, 50:250] = 255
        mask[70:90, 50:280] = 255
        mask[120:140, 50:200] = 255
        return mask
    
    def test_extract_regions_basic(self, sample_mask):
        """Test basic mask region extraction."""
        try:
            import cv2
        except ImportError:
            pytest.skip("OpenCV not available")
        
        regions = extract_mask_regions_from_unet(
            sample_mask,
            class_names=["surname", "given_names", "dob"]
        )
        
        assert len(regions) == 3
        # Sorted by y-position
        assert regions[0].class_name == "surname"
        assert regions[1].class_name == "given_names"
        assert regions[2].class_name == "dob"
    
    def test_extract_regions_max_limit(self, sample_mask):
        """Test max_regions limit."""
        try:
            import cv2
        except ImportError:
            pytest.skip("OpenCV not available")
        
        regions = extract_mask_regions_from_unet(sample_mask, max_regions=2)
        
        assert len(regions) == 2
    
    def test_extract_regions_min_area(self):
        """Test min_area filtering."""
        try:
            import cv2
        except ImportError:
            pytest.skip("OpenCV not available")
        
        # Create mask with one small and one large region
        mask = np.zeros((100, 100), dtype=np.uint8)
        mask[10:12, 10:12] = 255  # 4 pixels
        mask[50:70, 50:90] = 255  # 800 pixels
        
        regions = extract_mask_regions_from_unet(mask, min_area=100)
        
        assert len(regions) == 1
        assert regions[0].area >= 100


class TestMatchResult:
    """Tests for MatchResult dataclass."""
    
    def test_is_matched_true(self):
        """Test is_matched property when matched."""
        result = MatchResult(
            mask_region=MaskRegion(
                class_id=0, class_name="test", mask=None,
                bounding_box=BoundingBox(0, 0, 100, 100), area=10000
            ),
            matched_box=BoundingBox(0, 0, 100, 100, label="TEST"),
            distance=0.05,
            confidence=0.9
        )
        
        assert result.is_matched is True
    
    def test_is_matched_false(self):
        """Test is_matched property when not matched."""
        result = MatchResult(
            mask_region=MaskRegion(
                class_id=0, class_name="test", mask=None,
                bounding_box=BoundingBox(0, 0, 100, 100), area=10000
            ),
            matched_box=None,
            distance=float('inf'),
            confidence=0.0
        )
        
        assert result.is_matched is False


class TestIntegration:
    """Integration tests simulating the full workflow."""
    
    def test_full_workflow(self):
        """Test complete mask-to-text matching workflow."""
        # Simulate U-Net output: 4 field regions (like ID card)
        mask_regions = [
            MaskRegion(
                class_id=0, class_name="surname",
                mask=None,
                bounding_box=BoundingBox(x=150, y=80, width=200, height=25),
                area=5000
            ),
            MaskRegion(
                class_id=1, class_name="given_names",
                mask=None,
                bounding_box=BoundingBox(x=150, y=120, width=200, height=25),
                area=5000
            ),
            MaskRegion(
                class_id=2, class_name="date_of_birth",
                mask=None,
                bounding_box=BoundingBox(x=150, y=160, width=100, height=25),
                area=2500
            ),
            MaskRegion(
                class_id=3, class_name="id_number",
                mask=None,
                bounding_box=BoundingBox(x=150, y=200, width=150, height=25),
                area=3750
            ),
        ]
        
        # Simulate OCR output: detected text boxes
        text_boxes = [
            BoundingBox(x=155, y=82, width=90, height=22, label="SMITH", confidence=0.95),
            BoundingBox(x=155, y=122, width=50, height=22, label="JOHN", confidence=0.93),
            BoundingBox(x=210, y=122, width=70, height=22, label="MICHAEL", confidence=0.91),
            BoundingBox(x=155, y=162, width=95, height=22, label="1990-03-15", confidence=0.89),
            BoundingBox(x=155, y=202, width=140, height=22, label="ABC123456", confidence=0.97),
            BoundingBox(x=400, y=80, width=50, height=22, label="PHOTO", confidence=0.85),  # Irrelevant
        ]
        
        # Image dimensions (ID card-ish)
        img_width, img_height = 500, 300
        
        # Run matching
        # Use larger threshold since boxes are slightly offset from mask regions
        matcher = CenterMatcher(CenterMatchingConfig(
            distance_threshold=0.15,  # Allow 15% of image dimension distance
            neighbor_distance_pixels=60
        ))
        
        # First, merge adjacent boxes (JOHN + MICHAEL)
        merged_boxes = matcher.merge_adjacent_boxes(text_boxes, img_width, img_height)
        
        # Then match
        results = matcher.match_masks_to_boxes(
            mask_regions, merged_boxes, img_width, img_height
        )
        
        # Verify results
        assert len(results) == 4
        
        # All should be matched
        for i, result in enumerate(results):
            assert result.is_matched, f"Region {i} ({result.mask_region.class_name}) not matched"
        
        # Verify correct field mapping
        assert results[0].mask_region.class_name == "surname"
        assert results[0].matched_box.label == "SMITH"
        
        assert results[2].mask_region.class_name == "date_of_birth"
        assert results[2].matched_box.label == "1990-03-15"
        
        assert results[3].mask_region.class_name == "id_number"
        assert results[3].matched_box.label == "ABC123456"
    
    def test_intern_code_compatibility(self):
        """
        Test that our implementation matches the intern code behavior.
        
        The intern code:
        1. Gets 4 mask regions sorted top-to-bottom
        2. Gets all OCR boxes
        3. Converts both to center ratios
        4. Matches by minimum sum of absolute differences
        """
        # Setup similar to intern code test case
        # Mask centers at fixed positions (like ID card fields)
        mask_ratios = [
            (0.35, 0.2),   # Surname
            (0.35, 0.35),  # Given names
            (0.35, 0.5),   # DOB
            (0.35, 0.65),  # ID number
        ]
        
        # OCR box centers (slightly offset, simulating real detection)
        box_ratios = [
            (0.36, 0.21),  # Should match surname
            (0.37, 0.36),  # Should match given names
            (0.34, 0.49),  # Should match DOB
            (0.35, 0.64),  # Should match ID number
            (0.8, 0.3),    # Noise (far away)
            (0.1, 0.9),    # Noise (far away)
        ]
        
        matcher = CenterMatcher()
        matches = matcher.match_centers(mask_ratios, box_ratios)
        
        # All 4 masks should be matched to boxes 0-3
        assert len(matches) == 4
        assert matches[0][0] == 0
        assert matches[1][0] == 1
        assert matches[2][0] == 2
        assert matches[3][0] == 3
