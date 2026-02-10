"""
ROI type definitions for text detection.

This module defines the data structures used to represent:
- Bounding boxes and polygons
- Text regions of interest (ROI)
- Detection results with versioning
"""

from dataclasses import dataclass, field
from typing import List, Dict, Any, Optional, Tuple
from enum import Enum
import uuid


class TextType(str, Enum):
    """Types of text regions detected."""
    CHARACTER = "character"
    WORD = "word"
    LINE = "line"
    PARAGRAPH = "paragraph"


@dataclass
class Point:
    """A 2D point coordinate."""
    x: float
    y: float
    
    def to_tuple(self) -> Tuple[float, float]:
        """Convert to tuple."""
        return (self.x, self.y)
    
    def to_int_tuple(self) -> Tuple[int, int]:
        """Convert to integer tuple."""
        return (int(round(self.x)), int(round(self.y)))
    
    def to_dict(self) -> Dict[str, float]:
        """Convert to dictionary."""
        return {"x": self.x, "y": self.y}


@dataclass
class BoundingBox:
    """Axis-aligned bounding box representation."""
    x: float  # Top-left x coordinate
    y: float  # Top-left y coordinate
    width: float
    height: float
    
    @property
    def x2(self) -> float:
        """Get right x coordinate."""
        return self.x + self.width
    
    @property
    def y2(self) -> float:
        """Get bottom y coordinate."""
        return self.y + self.height
    
    @property
    def center(self) -> Point:
        """Get center point."""
        return Point(self.x + self.width / 2, self.y + self.height / 2)
    
    @property
    def area(self) -> float:
        """Get box area."""
        return self.width * self.height
    
    def to_tuple(self) -> Tuple[float, float, float, float]:
        """Convert to tuple (x, y, width, height)."""
        return (self.x, self.y, self.width, self.height)
    
    def to_xyxy(self) -> Tuple[float, float, float, float]:
        """Convert to (x1, y1, x2, y2) format."""
        return (self.x, self.y, self.x2, self.y2)
    
    def to_int_xyxy(self) -> Tuple[int, int, int, int]:
        """Convert to integer (x1, y1, x2, y2) format."""
        return (
            int(round(self.x)),
            int(round(self.y)),
            int(round(self.x2)),
            int(round(self.y2))
        )
    
    def to_dict(self) -> Dict[str, float]:
        """Convert to dictionary."""
        return {
            "x": self.x,
            "y": self.y,
            "width": self.width,
            "height": self.height,
        }
    
    @classmethod
    def from_xyxy(cls, x1: float, y1: float, x2: float, y2: float) -> "BoundingBox":
        """Create from (x1, y1, x2, y2) format."""
        return cls(x=x1, y=y1, width=x2 - x1, height=y2 - y1)
    
    @classmethod
    def from_points(cls, points: List[Point]) -> "BoundingBox":
        """Create bounding box that encloses all points."""
        if not points:
            raise ValueError("Cannot create bounding box from empty points list")
        
        xs = [p.x for p in points]
        ys = [p.y for p in points]
        x1, x2 = min(xs), max(xs)
        y1, y2 = min(ys), max(ys)
        return cls.from_xyxy(x1, y1, x2, y2)
    
    def intersection(self, other: "BoundingBox") -> Optional["BoundingBox"]:
        """Compute intersection with another box."""
        x1 = max(self.x, other.x)
        y1 = max(self.y, other.y)
        x2 = min(self.x2, other.x2)
        y2 = min(self.y2, other.y2)
        
        if x2 <= x1 or y2 <= y1:
            return None
        return BoundingBox.from_xyxy(x1, y1, x2, y2)
    
    def union(self, other: "BoundingBox") -> "BoundingBox":
        """Compute union (enclosing box) with another box."""
        x1 = min(self.x, other.x)
        y1 = min(self.y, other.y)
        x2 = max(self.x2, other.x2)
        y2 = max(self.y2, other.y2)
        return BoundingBox.from_xyxy(x1, y1, x2, y2)
    
    def iou(self, other: "BoundingBox") -> float:
        """Compute Intersection over Union with another box."""
        intersection = self.intersection(other)
        if intersection is None:
            return 0.0
        
        union_area = self.area + other.area - intersection.area
        if union_area <= 0:
            return 0.0
        
        return intersection.area / union_area
    
    def expand(self, margin: float) -> "BoundingBox":
        """Expand box by margin on all sides."""
        return BoundingBox(
            x=self.x - margin,
            y=self.y - margin,
            width=self.width + 2 * margin,
            height=self.height + 2 * margin
        )


@dataclass
class TextROI:
    """
    Text Region of Interest.
    
    Represents a detected text region with its bounding box,
    confidence scores, and type information.
    """
    roi_id: str  # Unique identifier for this ROI
    bounding_box: BoundingBox  # Axis-aligned bounding box
    confidence: float  # Overall confidence score (0.0 - 1.0)
    text_type: TextType  # Type of text region
    polygon: List[Point]  # For rotated/non-rectangular text
    affinity_score: float  # CRAFT affinity score (link between characters)
    region_score: float  # CRAFT region score (text likelihood)
    
    # Optional metadata
    parent_roi_id: Optional[str] = None  # Parent ROI (e.g., word's parent line)
    child_roi_ids: List[str] = field(default_factory=list)  # Child ROIs
    
    @classmethod
    def create(
        cls,
        bounding_box: BoundingBox,
        confidence: float,
        text_type: TextType,
        polygon: Optional[List[Point]] = None,
        affinity_score: float = 0.0,
        region_score: float = 0.0,
        parent_roi_id: Optional[str] = None,
    ) -> "TextROI":
        """Factory method to create a TextROI with auto-generated ID."""
        roi_id = f"roi_{uuid.uuid4().hex[:12]}"
        
        # If no polygon provided, create from bounding box corners
        if polygon is None:
            polygon = [
                Point(bounding_box.x, bounding_box.y),
                Point(bounding_box.x2, bounding_box.y),
                Point(bounding_box.x2, bounding_box.y2),
                Point(bounding_box.x, bounding_box.y2),
            ]
        
        return cls(
            roi_id=roi_id,
            bounding_box=bounding_box,
            confidence=confidence,
            text_type=text_type,
            polygon=polygon,
            affinity_score=affinity_score,
            region_score=region_score,
            parent_roi_id=parent_roi_id,
        )
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "roi_id": self.roi_id,
            "bounding_box": self.bounding_box.to_dict(),
            "confidence": self.confidence,
            "text_type": self.text_type.value,
            "polygon": [p.to_dict() for p in self.polygon],
            "affinity_score": self.affinity_score,
            "region_score": self.region_score,
            "parent_roi_id": self.parent_roi_id,
            "child_roi_ids": self.child_roi_ids,
        }


@dataclass
class TextDetectionResult:
    """
    Complete result from text detection pipeline.
    
    Contains all detected ROIs with versioning and metadata
    for reproducibility and audit purposes.
    """
    image_hash: str  # SHA256 hash of input image
    rois: List[TextROI]  # All detected text ROIs
    model_version: str  # CRAFT model version
    processing_time_ms: float  # Total processing time
    thresholds_used: Dict[str, float]  # Thresholds applied
    
    # Metadata
    success: bool = True
    error_message: Optional[str] = None
    image_size: Tuple[int, int] = (0, 0)  # (width, height)
    suite_version: str = "1.0.0"  # Text detection suite version
    
    # Score maps (optional, for debugging)
    region_score_map: Optional[Any] = None  # np.ndarray, not stored in dict
    affinity_score_map: Optional[Any] = None  # np.ndarray, not stored in dict
    
    @property
    def character_rois(self) -> List[TextROI]:
        """Get only character-level ROIs."""
        return [r for r in self.rois if r.text_type == TextType.CHARACTER]
    
    @property
    def word_rois(self) -> List[TextROI]:
        """Get only word-level ROIs."""
        return [r for r in self.rois if r.text_type == TextType.WORD]
    
    @property
    def line_rois(self) -> List[TextROI]:
        """Get only line-level ROIs."""
        return [r for r in self.rois if r.text_type == TextType.LINE]
    
    @property
    def paragraph_rois(self) -> List[TextROI]:
        """Get only paragraph-level ROIs."""
        return [r for r in self.rois if r.text_type == TextType.PARAGRAPH]
    
    def get_rois_by_type(self, text_type: TextType) -> List[TextROI]:
        """Get ROIs filtered by type."""
        return [r for r in self.rois if r.text_type == text_type]
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "success": self.success,
            "image_hash": self.image_hash,
            "image_size": self.image_size,
            "model_version": self.model_version,
            "suite_version": self.suite_version,
            "processing_time_ms": self.processing_time_ms,
            "thresholds_used": self.thresholds_used,
            "error_message": self.error_message,
            "roi_count": len(self.rois),
            "roi_counts_by_type": {
                "character": len(self.character_rois),
                "word": len(self.word_rois),
                "line": len(self.line_rois),
                "paragraph": len(self.paragraph_rois),
            },
            "rois": [roi.to_dict() for roi in self.rois],
        }
