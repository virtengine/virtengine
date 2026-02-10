"""
Center-based matching for field-to-text correspondence.

This module provides algorithms for matching U-Net mask regions to OCR text boxes
by comparing their center positions. This allows field identification without
pre-defined layouts.

Ported from RMIT OCR_Document_Scan intern project (2023) with improvements:
- Configurable distance thresholds
- Determinism controls for blockchain consensus
- Proper error handling
- Type annotations
- Generalized to arbitrary number of mask regions

Task Reference: VE-3041 - Port Center-Matching Algorithm

Original Algorithm Source:
- getCenterRatios: Converts box centers to normalized ratios (0-1) of image dimensions
- matchCenters: Maps mask region centers to nearest OCR text box centers
- NearestBox: Extends matched boxes by finding horizontally adjacent text boxes
"""

import numpy as np
from typing import List, Tuple, Dict, Optional, Any, Sequence
from dataclasses import dataclass
from enum import Enum

# Set deterministic seed for blockchain consensus
np.random.seed(42)


@dataclass
class BoundingBox:
    """
    Represents a bounding box with position and size.
    
    Coordinates use top-left origin (x, y) with width and height.
    This matches the (x, w, y, h) format from the intern code's getBoxRegions.
    
    Attributes:
        x: Left edge x coordinate
        y: Top edge y coordinate
        width: Box width
        height: Box height
        label: Optional text label (OCR result)
        confidence: OCR confidence score (0.0-1.0)
    """
    x: int
    y: int
    width: int
    height: int
    label: Optional[str] = None
    confidence: float = 1.0
    
    @property
    def center(self) -> Tuple[float, float]:
        """Return the center point of the box as (cx, cy)."""
        return (self.x + self.width / 2.0, self.y + self.height / 2.0)
    
    @property
    def right_center(self) -> Tuple[float, float]:
        """Return the right edge center point (for adjacent box detection)."""
        return (self.x + self.width, self.y + self.height / 2.0)
    
    @property
    def left_center(self) -> Tuple[float, float]:
        """Return the left edge center point (for adjacent box detection)."""
        return (self.x, self.y + self.height / 2.0)
    
    def get_center_ratio(self, img_width: int, img_height: int) -> Tuple[float, float]:
        """
        Return center position as ratio (0-1) of image dimensions.
        
        This mirrors getCenterRatios from the intern code.
        
        Args:
            img_width: Image width in pixels
            img_height: Image height in pixels
            
        Returns:
            Tuple of (x_ratio, y_ratio) where each is in range [0, 1]
        """
        cx, cy = self.center
        return (cx / img_width, cy / img_height)
    
    def to_tuple(self) -> Tuple[int, int, int, int]:
        """Return as (x, y, width, height) tuple."""
        return (self.x, self.y, self.width, self.height)
    
    def area(self) -> int:
        """Return the area of the box."""
        return self.width * self.height
    
    @classmethod
    def from_xyxy(cls, x1: int, y1: int, x2: int, y2: int, 
                  label: Optional[str] = None, confidence: float = 1.0) -> "BoundingBox":
        """
        Create from corner coordinates (x1, y1) to (x2, y2).
        
        Args:
            x1, y1: Top-left corner
            x2, y2: Bottom-right corner
            label: Optional text label
            confidence: Confidence score
        """
        return cls(
            x=min(x1, x2),
            y=min(y1, y2),
            width=abs(x2 - x1),
            height=abs(y2 - y1),
            label=label,
            confidence=confidence
        )
    
    @classmethod
    def from_polygon(cls, points: Sequence[Tuple[int, int]], 
                     label: Optional[str] = None, confidence: float = 1.0) -> "BoundingBox":
        """
        Create from polygon points (e.g., CRAFT detector output).
        
        This matches the intern code's getBoxRegions which converts
        8-point polygons to (x, w, y, h) format.
        
        Args:
            points: Sequence of (x, y) coordinate tuples
            label: Optional text label
            confidence: Confidence score
        """
        xs = [p[0] for p in points]
        ys = [p[1] for p in points]
        x = min(xs)
        y = min(ys)
        width = max(xs) - x
        height = max(ys) - y
        return cls(x=int(x), y=int(y), width=int(width), height=int(height),
                   label=label, confidence=confidence)


@dataclass
class MaskRegion:
    """
    Represents a segmented mask region from U-Net.
    
    This corresponds to the contour regions extracted in getCenterOfMasks.
    
    Attributes:
        class_id: Numeric class identifier
        class_name: Human-readable class name (e.g., "surname", "dob")
        mask: Binary mask as numpy array (optional, for weighted centroid)
        bounding_box: Bounding box around the mask region
        area: Mask area in pixels
    """
    class_id: int
    class_name: str
    mask: Optional[np.ndarray]
    bounding_box: BoundingBox
    area: int
    
    @property
    def center(self) -> Tuple[float, float]:
        """Return the centroid of the mask region."""
        return self.bounding_box.center
    
    def get_center_ratio(self, img_width: int, img_height: int) -> Tuple[float, float]:
        """Return center as ratio of image dimensions."""
        return self.bounding_box.get_center_ratio(img_width, img_height)
    
    @classmethod
    def from_contour(cls, contour: np.ndarray, class_id: int, 
                     class_name: str, mask: Optional[np.ndarray] = None) -> "MaskRegion":
        """
        Create from OpenCV contour.
        
        Args:
            contour: OpenCV contour array
            class_id: Class identifier
            class_name: Class name
            mask: Optional binary mask
        """
        import cv2
        x, y, w, h = cv2.boundingRect(contour)
        area = cv2.contourArea(contour)
        return cls(
            class_id=class_id,
            class_name=class_name,
            mask=mask,
            bounding_box=BoundingBox(x=x, y=y, width=w, height=h),
            area=int(area)
        )


@dataclass
class MatchResult:
    """
    Result of matching a mask region to a text box.
    
    Attributes:
        mask_region: The source mask region
        matched_box: The matched text box (None if no match)
        distance: Euclidean distance between centers (in ratio space)
        confidence: Match confidence (inverse of distance, clamped)
    """
    mask_region: MaskRegion
    matched_box: Optional[BoundingBox]
    distance: float
    confidence: float
    
    @property
    def is_matched(self) -> bool:
        """Return True if a match was found."""
        return self.matched_box is not None


class CenterMatchingConfig:
    """
    Configuration for center matching algorithm.
    
    Attributes:
        distance_threshold: Maximum normalized distance for valid match (default 0.15)
        merge_threshold: Distance threshold for merging adjacent boxes (default 0.02)
        min_confidence: Minimum OCR confidence to consider a box (default 0.5)
        use_weighted_center: Use mask-weighted centroid instead of bbox center
        neighbor_distance_pixels: Pixel distance for finding adjacent boxes (from intern code)
    """
    
    def __init__(
        self,
        distance_threshold: float = 0.15,
        merge_threshold: float = 0.02,
        min_confidence: float = 0.5,
        use_weighted_center: bool = True,
        neighbor_distance_pixels: float = 50.0,
    ):
        if distance_threshold <= 0:
            raise ValueError("distance_threshold must be positive")
        if merge_threshold < 0:
            raise ValueError("merge_threshold must be non-negative")
        if not 0 <= min_confidence <= 1:
            raise ValueError("min_confidence must be between 0 and 1")
        
        self.distance_threshold = distance_threshold
        self.merge_threshold = merge_threshold
        self.min_confidence = min_confidence
        self.use_weighted_center = use_weighted_center
        self.neighbor_distance_pixels = neighbor_distance_pixels


class CenterMatcher:
    """
    Matches U-Net mask regions to OCR text boxes by center positions.
    
    This implements the core matching algorithm from the intern code:
    1. Extract center positions from mask regions (getCenterOfMasks)
    2. Extract center positions from OCR text boxes (getBoxRegions)
    3. Convert to normalized ratios (getCenterRatios)
    4. For each mask, find nearest text box within threshold (matchCenters)
    5. Optionally extend boxes with adjacent neighbors (NearestBox)
    
    Example usage:
        matcher = CenterMatcher()
        results = matcher.match_masks_to_boxes(
            mask_regions=unet_masks,
            text_boxes=ocr_boxes,
            img_width=640,
            img_height=480
        )
        for result in results:
            if result.is_matched:
                print(f"{result.mask_region.class_name}: {result.matched_box.label}")
    """
    
    def __init__(self, config: Optional[CenterMatchingConfig] = None):
        """
        Initialize the center matcher.
        
        Args:
            config: Configuration options. Uses defaults if None.
        """
        self.config = config or CenterMatchingConfig()
    
    def get_center_ratios(
        self,
        regions: List[MaskRegion],
        img_width: int,
        img_height: int
    ) -> List[Tuple[float, float]]:
        """
        Get center positions as ratios for mask regions.
        
        This mirrors getCenterRatios from the intern code.
        
        Args:
            regions: List of mask regions from U-Net
            img_width: Image width in pixels
            img_height: Image height in pixels
            
        Returns:
            List of (x_ratio, y_ratio) tuples, each in range [0, 1]
        """
        if img_width <= 0 or img_height <= 0:
            raise ValueError("Image dimensions must be positive")
        
        ratios = []
        for region in regions:
            if self.config.use_weighted_center and region.mask is not None:
                # Use mask-weighted centroid
                cx, cy = self._weighted_centroid(region.mask)
            else:
                cx, cy = region.center
            
            ratios.append((cx / img_width, cy / img_height))
        
        return ratios
    
    def get_box_center_ratios(
        self,
        boxes: List[BoundingBox],
        img_width: int,
        img_height: int
    ) -> List[Tuple[float, float]]:
        """
        Get center positions as ratios for bounding boxes.
        
        Args:
            boxes: List of OCR bounding boxes
            img_width: Image width in pixels
            img_height: Image height in pixels
            
        Returns:
            List of (x_ratio, y_ratio) tuples
        """
        if img_width <= 0 or img_height <= 0:
            raise ValueError("Image dimensions must be positive")
        
        return [box.get_center_ratio(img_width, img_height) for box in boxes]
    
    def match_centers(
        self,
        mask_ratios: List[Tuple[float, float]],
        box_ratios: List[Tuple[float, float]],
    ) -> Dict[int, Tuple[int, float]]:
        """
        Match mask centers to nearest box centers.
        
        This implements the core matching logic from matchCenters in the intern code,
        generalized to handle arbitrary number of mask regions.
        
        The intern code computed absolute differences and summed them:
            bbb0[i] = abs(ratios1[0] - r2)
            sum_b0 = np.sum(bbb0, axis=1)
            
        This is equivalent to Manhattan distance. We use Euclidean for better accuracy.
        
        Args:
            mask_ratios: Center ratios from mask regions
            box_ratios: Center ratios from text boxes
            
        Returns:
            Dictionary mapping mask index to (box_index, distance)
        """
        if not box_ratios:
            return {}
        
        matches: Dict[int, Tuple[int, float]] = {}
        
        for i, mask_center in enumerate(mask_ratios):
            best_distance = float('inf')
            best_box_idx = -1
            
            for j, box_center in enumerate(box_ratios):
                distance = self._euclidean_distance(mask_center, box_center)
                
                if distance < best_distance and distance < self.config.distance_threshold:
                    best_distance = distance
                    best_box_idx = j
            
            if best_box_idx >= 0:
                matches[i] = (best_box_idx, best_distance)
        
        return matches
    
    def match_masks_to_boxes(
        self,
        mask_regions: List[MaskRegion],
        text_boxes: List[BoundingBox],
        img_width: int,
        img_height: int
    ) -> List[MatchResult]:
        """
        Full matching pipeline: mask regions to text boxes.
        
        Args:
            mask_regions: U-Net segmentation regions
            text_boxes: OCR detected text boxes
            img_width: Image width in pixels
            img_height: Image height in pixels
            
        Returns:
            List of MatchResult for each mask region
        """
        # Filter boxes by confidence
        valid_boxes = [
            box for box in text_boxes 
            if box.confidence >= self.config.min_confidence
        ]
        
        mask_ratios = self.get_center_ratios(mask_regions, img_width, img_height)
        box_ratios = self.get_box_center_ratios(valid_boxes, img_width, img_height)
        
        matches = self.match_centers(mask_ratios, box_ratios)
        
        results = []
        for i, region in enumerate(mask_regions):
            if i in matches:
                box_idx, distance = matches[i]
                # Confidence is inverse of distance, clamped to [0, 1]
                confidence = max(0.0, min(1.0, 1.0 - distance / self.config.distance_threshold))
                results.append(MatchResult(
                    mask_region=region,
                    matched_box=valid_boxes[box_idx],
                    distance=distance,
                    confidence=confidence
                ))
            else:
                results.append(MatchResult(
                    mask_region=region,
                    matched_box=None,
                    distance=float('inf'),
                    confidence=0.0
                ))
        
        return results
    
    def find_adjacent_boxes(
        self,
        target_box: BoundingBox,
        all_boxes: List[BoundingBox],
        img_width: int,
        direction: str = "right"
    ) -> List[BoundingBox]:
        """
        Find boxes adjacent to the target box.
        
        This implements the neighbor-finding logic from NearestBox class.
        The intern code finds boxes whose left center is within DISTANCE_THRESH
        of the target box's right center (or vice versa for left neighbors).
        
        Args:
            target_box: The reference box
            all_boxes: All detected text boxes
            img_width: Image width for threshold normalization
            direction: "right" or "left"
            
        Returns:
            List of adjacent boxes, sorted by x position
        """
        threshold = self.config.neighbor_distance_pixels
        neighbors = []
        
        if direction == "right":
            ref_point = target_box.right_center
            for box in all_boxes:
                if box is target_box:
                    continue
                compare_point = box.left_center
                distance = self._euclidean_distance_pixels(ref_point, compare_point)
                if 0 < distance < threshold:
                    neighbors.append(box)
        else:  # left
            ref_point = target_box.left_center
            for box in all_boxes:
                if box is target_box:
                    continue
                compare_point = box.right_center
                distance = self._euclidean_distance_pixels(ref_point, compare_point)
                if 0 < distance < threshold:
                    neighbors.append(box)
        
        # Sort by x position
        return sorted(neighbors, key=lambda b: b.x)
    
    def extend_box_with_neighbors(
        self,
        target_box: BoundingBox,
        all_boxes: List[BoundingBox],
        img_width: int
    ) -> BoundingBox:
        """
        Extend a box to include horizontally adjacent boxes.
        
        This implements getExtendedBoxCoordinates from NearestBox.
        
        Args:
            target_box: The reference box
            all_boxes: All detected text boxes
            img_width: Image width
            
        Returns:
            Extended bounding box encompassing all adjacent boxes
        """
        right_neighbors = self.find_adjacent_boxes(target_box, all_boxes, img_width, "right")
        left_neighbors = self.find_adjacent_boxes(target_box, all_boxes, img_width, "left")
        
        if not right_neighbors and not left_neighbors:
            return target_box
        
        # Start with target box bounds
        min_x = target_box.x
        max_x = target_box.x + target_box.width
        min_y = target_box.y
        max_y = target_box.y + target_box.height
        
        # Combine labels
        labels = [target_box.label] if target_box.label else []
        confidences = [target_box.confidence]
        
        for box in left_neighbors + right_neighbors:
            min_x = min(min_x, box.x)
            max_x = max(max_x, box.x + box.width)
            min_y = min(min_y, box.y)
            max_y = max(max_y, box.y + box.height)
            if box.label:
                labels.append(box.label)
            confidences.append(box.confidence)
        
        # Combined label (in reading order: left to right)
        all_boxes_sorted = sorted(left_neighbors + [target_box] + right_neighbors, key=lambda b: b.x)
        combined_label = " ".join(b.label for b in all_boxes_sorted if b.label)
        
        return BoundingBox(
            x=min_x,
            y=min_y,
            width=max_x - min_x,
            height=max_y - min_y,
            label=combined_label or None,
            confidence=min(confidences)  # Use minimum confidence
        )
    
    def merge_adjacent_boxes(
        self,
        boxes: List[BoundingBox],
        img_width: int,
        img_height: int
    ) -> List[BoundingBox]:
        """
        Merge horizontally adjacent boxes (e.g., first name + middle name).
        
        Uses normalized distance threshold (merge_threshold in config).
        
        Args:
            boxes: List of text boxes
            img_width: Image width in pixels
            img_height: Image height in pixels
            
        Returns:
            List of merged boxes
        """
        if not boxes:
            return []
        
        # Sort by x position
        sorted_boxes = sorted(boxes, key=lambda b: b.x)
        merged = [sorted_boxes[0]]
        
        for box in sorted_boxes[1:]:
            last = merged[-1]
            
            # Check if boxes are adjacent horizontally and aligned vertically
            horizontal_gap = (box.x - (last.x + last.width)) / img_width
            vertical_overlap = self._vertical_overlap_ratio(last, box)
            
            if horizontal_gap < self.config.merge_threshold and vertical_overlap > 0.5:
                # Merge boxes
                new_x = last.x
                new_y = min(last.y, box.y)
                new_width = (box.x + box.width) - last.x
                new_height = max(last.y + last.height, box.y + box.height) - new_y
                
                # Combine labels
                combined_label = f"{last.label or ''} {box.label or ''}".strip() or None
                
                merged[-1] = BoundingBox(
                    x=new_x, y=new_y, width=new_width, height=new_height,
                    label=combined_label,
                    confidence=min(last.confidence, box.confidence)
                )
            else:
                merged.append(box)
        
        return merged
    
    def _euclidean_distance(
        self,
        p1: Tuple[float, float],
        p2: Tuple[float, float]
    ) -> float:
        """Calculate Euclidean distance between two points."""
        return float(np.sqrt((p1[0] - p2[0])**2 + (p1[1] - p2[1])**2))
    
    def _euclidean_distance_pixels(
        self,
        p1: Tuple[float, float],
        p2: Tuple[float, float]
    ) -> float:
        """Calculate Euclidean distance in pixel coordinates."""
        return float(np.linalg.norm(np.array(p1) - np.array(p2)))
    
    def _weighted_centroid(self, mask: np.ndarray) -> Tuple[float, float]:
        """
        Calculate mask-weighted centroid.
        
        Args:
            mask: Binary mask array
            
        Returns:
            (cx, cy) centroid coordinates
        """
        if mask.sum() == 0:
            return (mask.shape[1] / 2.0, mask.shape[0] / 2.0)
        
        y_coords, x_coords = np.nonzero(mask)
        return (float(x_coords.mean()), float(y_coords.mean()))
    
    def _vertical_overlap_ratio(self, box1: BoundingBox, box2: BoundingBox) -> float:
        """
        Calculate vertical overlap ratio between boxes.
        
        Args:
            box1, box2: Bounding boxes to compare
            
        Returns:
            Overlap ratio in range [0, 1]
        """
        top = max(box1.y, box2.y)
        bottom = min(box1.y + box1.height, box2.y + box2.height)
        
        if bottom <= top:
            return 0.0
        
        overlap = bottom - top
        min_height = min(box1.height, box2.height)
        
        return float(overlap / min_height) if min_height > 0 else 0.0


def extract_mask_regions_from_unet(
    predicted_mask: np.ndarray,
    class_names: Optional[List[str]] = None,
    min_area: int = 100,
    max_regions: int = 10
) -> List[MaskRegion]:
    """
    Extract mask regions from U-Net prediction.
    
    This implements getCenterOfMasks logic from the intern code.
    
    Args:
        predicted_mask: Binary or multi-class mask from U-Net
        class_names: List of class names indexed by class_id
        min_area: Minimum contour area to consider
        max_regions: Maximum number of regions to return
        
    Returns:
        List of MaskRegion objects, sorted by y-position (top to bottom)
    """
    try:
        import cv2
    except ImportError:
        raise ImportError("OpenCV (cv2) is required for mask extraction")
    
    # Ensure binary mask
    if predicted_mask.dtype != np.uint8:
        thresh = (predicted_mask > 0.5).astype(np.uint8) * 255
    else:
        thresh = predicted_mask
    
    contours, _ = cv2.findContours(thresh, cv2.RETR_TREE, cv2.CHAIN_APPROX_SIMPLE)
    
    # Filter by area and sort by size (largest first, then take top N)
    valid_contours = [c for c in contours if cv2.contourArea(c) >= min_area]
    valid_contours = sorted(valid_contours, key=cv2.contourArea, reverse=True)[:max_regions]
    
    # Sort by y-position (top to bottom) as in intern code
    valid_contours = sorted(valid_contours, key=lambda c: cv2.boundingRect(c)[1])
    
    regions = []
    for i, contour in enumerate(valid_contours):
        class_name = class_names[i] if class_names and i < len(class_names) else f"field_{i}"
        region = MaskRegion.from_contour(contour, class_id=i, class_name=class_name, mask=None)
        regions.append(region)
    
    return regions


def convert_ocr_boxes(
    regions: np.ndarray,
    labels: Optional[List[str]] = None,
    confidences: Optional[List[float]] = None
) -> Tuple[List[BoundingBox], List[Tuple[float, float]]]:
    """
    Convert OCR detector output to bounding boxes.
    
    This implements getBoxRegions from the intern code.
    
    Args:
        regions: Array of shape (N, 4, 2) with polygon vertices
        labels: Optional text labels for each box
        confidences: Optional confidence scores
        
    Returns:
        Tuple of (boxes, centers) where centers are (cx, cy) tuples
    """
    boxes = []
    centers = []
    
    for i, box_region in enumerate(regions):
        # Flatten polygon to coordinates
        if box_region.shape == (4, 2):
            points = [(int(box_region[j, 0]), int(box_region[j, 1])) for j in range(4)]
        else:
            # Assume flattened 8-point format
            coords = box_region.flatten()
            points = [(int(coords[j]), int(coords[j+1])) for j in range(0, len(coords), 2)]
        
        label = labels[i] if labels and i < len(labels) else None
        confidence = confidences[i] if confidences and i < len(confidences) else 1.0
        
        box = BoundingBox.from_polygon(points, label=label, confidence=confidence)
        boxes.append(box)
        centers.append(box.center)
    
    return boxes, centers
