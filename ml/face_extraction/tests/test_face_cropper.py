"""
Tests for face cropping module.

This module tests face extraction and cropping including:
- Bounding box extraction from masks
- Face cropping with margins
- Validation of face regions
- Edge cases and error handling
"""

import pytest
import numpy as np


class TestBoundingBox:
    """Tests for BoundingBox dataclass."""
    
    def test_bounding_box_creation(self):
        """Test BoundingBox creation."""
        from ml.face_extraction.face_cropper import BoundingBox
        
        bbox = BoundingBox(x=10, y=20, width=100, height=150)
        
        assert bbox.x == 10
        assert bbox.y == 20
        assert bbox.width == 100
        assert bbox.height == 150
    
    def test_bounding_box_properties(self):
        """Test BoundingBox computed properties."""
        from ml.face_extraction.face_cropper import BoundingBox
        
        bbox = BoundingBox(x=10, y=20, width=100, height=150)
        
        assert bbox.x1 == 10
        assert bbox.y1 == 20
        assert bbox.x2 == 110
        assert bbox.y2 == 170
        assert bbox.center == (60, 95)
        assert bbox.area == 15000
        assert abs(bbox.aspect_ratio - 100/150) < 0.01
    
    def test_bounding_box_expand(self):
        """Test BoundingBox expansion."""
        from ml.face_extraction.face_cropper import BoundingBox
        
        bbox = BoundingBox(x=100, y=100, width=100, height=100)
        expanded = bbox.expand(0.1)
        
        assert expanded.x == 90
        assert expanded.y == 90
        assert expanded.width == 120
        assert expanded.height == 120
    
    def test_bounding_box_clip(self):
        """Test BoundingBox clipping to image boundaries."""
        from ml.face_extraction.face_cropper import BoundingBox
        
        # Box that extends beyond image
        bbox = BoundingBox(x=-10, y=-10, width=200, height=200)
        clipped = bbox.clip_to_image(100, 100)
        
        assert clipped.x == 0
        assert clipped.y == 0
        assert clipped.x2 == 100
        assert clipped.y2 == 100
    
    def test_bounding_box_to_dict(self):
        """Test BoundingBox serialization."""
        from ml.face_extraction.face_cropper import BoundingBox
        
        bbox = BoundingBox(x=10, y=20, width=100, height=150)
        bbox_dict = bbox.to_dict()
        
        assert isinstance(bbox_dict, dict)
        assert bbox_dict["x"] == 10
        assert bbox_dict["width"] == 100
        assert "center" in bbox_dict
        assert "area" in bbox_dict


class TestFaceCropper:
    """Tests for FaceCropper class."""
    
    def test_initialization(self, cropper_config):
        """Test cropper initialization."""
        from ml.face_extraction.face_cropper import FaceCropper
        
        cropper = FaceCropper(cropper_config)
        
        assert cropper is not None
        assert cropper.config == cropper_config
    
    def test_default_initialization(self):
        """Test cropper with default config."""
        from ml.face_extraction.face_cropper import FaceCropper
        
        cropper = FaceCropper()
        
        assert cropper is not None
        assert cropper.config is not None


class TestBoundingBoxExtraction:
    """Tests for bounding box extraction from masks."""
    
    def test_extract_bounding_box_simple(self, cropper_config):
        """Test bounding box extraction from simple mask."""
        from ml.face_extraction.face_cropper import FaceCropper
        
        cropper = FaceCropper(cropper_config)
        
        # Create simple rectangular mask
        mask = np.zeros((100, 100), dtype=np.uint8)
        mask[20:60, 30:80] = 255
        
        bbox = cropper.extract_bounding_box(mask)
        
        assert bbox is not None
        assert bbox.x == 30
        assert bbox.y == 20
        assert bbox.width == 50
        assert bbox.height == 40
    
    def test_extract_bounding_box_ellipse(self, cropper_config, sample_face_mask):
        """Test bounding box extraction from elliptical mask."""
        from ml.face_extraction.face_cropper import FaceCropper
        
        cropper = FaceCropper(cropper_config)
        
        # Threshold mask to binary
        binary = (sample_face_mask > 0.5).astype(np.uint8) * 255
        
        bbox = cropper.extract_bounding_box(binary)
        
        assert bbox is not None
        assert bbox.width > 0
        assert bbox.height > 0
    
    def test_extract_bounding_box_empty_mask(self, cropper_config):
        """Test bounding box extraction from empty mask."""
        from ml.face_extraction.face_cropper import FaceCropper
        
        cropper = FaceCropper(cropper_config)
        
        empty_mask = np.zeros((100, 100), dtype=np.uint8)
        bbox = cropper.extract_bounding_box(empty_mask)
        
        assert bbox is None
    
    def test_extract_bounding_box_multiple_components(self, cropper_config):
        """Test bounding box extraction from mask with multiple components."""
        from ml.face_extraction.face_cropper import FaceCropper
        
        cropper = FaceCropper(cropper_config)
        
        # Create mask with multiple regions
        mask = np.zeros((100, 100), dtype=np.uint8)
        mask[10:30, 10:30] = 255  # Region 1
        mask[60:80, 60:80] = 255  # Region 2
        
        bbox = cropper.extract_bounding_box(mask)
        
        # Should encompass all regions
        assert bbox is not None
        assert bbox.x1 <= 10
        assert bbox.y1 <= 10
        assert bbox.x2 >= 80
        assert bbox.y2 >= 80


class TestFaceCropping:
    """Tests for face cropping operations."""
    
    def test_crop_face_basic(self, cropper_config, sample_document_image):
        """Test basic face cropping."""
        from ml.face_extraction.face_cropper import FaceCropper, BoundingBox
        
        cropper = FaceCropper(cropper_config)
        
        bbox = BoundingBox(x=50, y=100, width=200, height=250)
        cropped = cropper.crop_face(sample_document_image, bbox)
        
        assert cropped is not None
        assert cropped.shape[0] > 0
        assert cropped.shape[1] > 0
    
    def test_crop_face_with_margin(self, cropper_config, sample_document_image):
        """Test face cropping with margin."""
        from ml.face_extraction.face_cropper import FaceCropper, BoundingBox
        
        cropper = FaceCropper(cropper_config)
        
        bbox = BoundingBox(x=100, y=100, width=100, height=100)
        
        # Crop with different margins
        crop_no_margin = cropper.crop_face(sample_document_image, bbox, margin=0.0)
        crop_with_margin = cropper.crop_face(sample_document_image, bbox, margin=0.2)
        
        # With margin should have more content before resize
        # (output size is fixed by config, but source area is larger)
        assert crop_no_margin is not None
        assert crop_with_margin is not None
    
    def test_crop_face_output_size(self, cropper_config, sample_document_image):
        """Test that cropped face has correct output size."""
        from ml.face_extraction.face_cropper import FaceCropper, BoundingBox
        
        cropper = FaceCropper(cropper_config)
        
        bbox = BoundingBox(x=50, y=100, width=200, height=250)
        cropped = cropper.crop_face(sample_document_image, bbox)
        
        expected_h, expected_w = cropper_config.output_size
        assert cropped.shape[0] == expected_h
        assert cropped.shape[1] == expected_w
    
    def test_crop_face_boundary_conditions(self, cropper_config, sample_document_image):
        """Test cropping when bbox extends beyond image."""
        from ml.face_extraction.face_cropper import FaceCropper, BoundingBox
        
        cropper = FaceCropper(cropper_config)
        
        # Bbox at image edge
        h, w = sample_document_image.shape[:2]
        bbox = BoundingBox(x=w-50, y=h-50, width=100, height=100)
        
        cropped = cropper.crop_face(sample_document_image, bbox)
        
        assert cropped is not None
    
    def test_crop_face_maintains_aspect_ratio(self, sample_document_image):
        """Test that aspect ratio is maintained when configured."""
        from ml.face_extraction.face_cropper import FaceCropper, BoundingBox
        from ml.face_extraction.config import CropperConfig
        
        config = CropperConfig(
            output_size=(224, 224),
            maintain_aspect_ratio=True
        )
        cropper = FaceCropper(config)
        
        # Non-square bbox
        bbox = BoundingBox(x=50, y=100, width=100, height=200)
        cropped = cropper.crop_face(sample_document_image, bbox)
        
        # Output should be square with padding
        assert cropped.shape[0] == 224
        assert cropped.shape[1] == 224


class TestFaceValidation:
    """Tests for face region validation."""
    
    def test_validate_valid_face(self, cropper_config, sample_document_image):
        """Test validation of valid face region."""
        from ml.face_extraction.face_cropper import FaceCropper, BoundingBox
        
        cropper = FaceCropper(cropper_config)
        
        bbox = BoundingBox(x=50, y=100, width=200, height=250)
        is_valid, reason = cropper.validate_face_region(bbox, sample_document_image.shape[:2])
        
        assert is_valid
        assert reason == "Valid"
    
    def test_validate_face_too_small_width(self, cropper_config, sample_document_image):
        """Test validation of face that's too narrow."""
        from ml.face_extraction.face_cropper import FaceCropper, BoundingBox
        
        cropper = FaceCropper(cropper_config)
        
        bbox = BoundingBox(x=50, y=100, width=10, height=250)  # Too narrow
        is_valid, reason = cropper.validate_face_region(bbox, sample_document_image.shape[:2])
        
        assert not is_valid
        assert "width" in reason.lower()
    
    def test_validate_face_too_small_height(self, cropper_config, sample_document_image):
        """Test validation of face that's too short."""
        from ml.face_extraction.face_cropper import FaceCropper, BoundingBox
        
        cropper = FaceCropper(cropper_config)
        
        bbox = BoundingBox(x=50, y=100, width=200, height=10)  # Too short
        is_valid, reason = cropper.validate_face_region(bbox, sample_document_image.shape[:2])
        
        assert not is_valid
        assert "height" in reason.lower()
    
    def test_validate_face_too_large(self, sample_document_image):
        """Test validation of face that's too large (percentage of image)."""
        from ml.face_extraction.face_cropper import FaceCropper, BoundingBox
        from ml.face_extraction.config import CropperConfig
        
        config = CropperConfig(max_face_percentage=0.5)
        cropper = FaceCropper(config)
        
        h, w = sample_document_image.shape[:2]
        # Face covering 90% of image
        bbox = BoundingBox(x=int(w*0.05), y=int(h*0.05), 
                          width=int(w*0.9), height=int(h*0.9))
        
        is_valid, reason = cropper.validate_face_region(bbox, (h, w))
        
        assert not is_valid
        assert "percentage" in reason.lower()


class TestFaceExtractionResult:
    """Tests for FaceExtractionResult."""
    
    def test_extraction_result_creation(self):
        """Test FaceExtractionResult creation."""
        from ml.face_extraction.face_cropper import FaceExtractionResult, BoundingBox
        
        face_image = np.zeros((224, 224, 3), dtype=np.uint8)
        bbox = BoundingBox(x=10, y=20, width=100, height=150)
        mask = np.zeros((768, 1024), dtype=np.uint8)
        
        result = FaceExtractionResult(
            face_image=face_image,
            bounding_box=bbox,
            mask=mask,
            confidence=0.95,
            model_version="1.0.0",
            model_hash="abc123"
        )
        
        assert result.success
        assert result.face_image is not None
        assert result.confidence == 0.95
    
    def test_extraction_result_to_dict(self):
        """Test FaceExtractionResult serialization."""
        from ml.face_extraction.face_cropper import FaceExtractionResult, BoundingBox
        
        result = FaceExtractionResult(
            face_image=np.zeros((224, 224, 3), dtype=np.uint8),
            bounding_box=BoundingBox(10, 20, 100, 150),
            mask=np.zeros((768, 1024), dtype=np.uint8),
            confidence=0.95,
            model_version="1.0.0",
            model_hash="abc123def456",
            face_percentage=0.05
        )
        
        result_dict = result.to_dict()
        
        assert isinstance(result_dict, dict)
        assert result_dict["success"]
        assert result_dict["confidence"] == 0.95
        assert "bounding_box" in result_dict
        assert result_dict["has_face_image"]


class TestExtractMethod:
    """Tests for the full extract method."""
    
    def test_extract_success(self, cropper_config, sample_document_image, sample_face_mask):
        """Test successful face extraction."""
        from ml.face_extraction.face_cropper import FaceCropper
        from ml.face_extraction.mask_processing import MaskProcessor
        
        cropper = FaceCropper(cropper_config)
        mask_processor = MaskProcessor()
        
        # Process mask
        processed = mask_processor.process(sample_face_mask)
        
        result = cropper.extract(
            image=sample_document_image,
            mask=processed.processed_mask,
            confidence=0.9,
            model_version="1.0.0",
            model_hash="testhash",
            return_face_image=True
        )
        
        assert result.success
        assert result.face_image is not None
        assert result.bounding_box is not None
    
    def test_extract_data_minimization(self, cropper_config, sample_document_image, sample_face_mask):
        """Test extraction with data minimization (no face image returned)."""
        from ml.face_extraction.face_cropper import FaceCropper
        from ml.face_extraction.mask_processing import MaskProcessor
        
        cropper = FaceCropper(cropper_config)
        mask_processor = MaskProcessor()
        
        processed = mask_processor.process(sample_face_mask)
        
        result = cropper.extract(
            image=sample_document_image,
            mask=processed.processed_mask,
            confidence=0.9,
            model_version="1.0.0",
            model_hash="testhash",
            return_face_image=False  # Data minimization
        )
        
        assert result.success
        assert result.face_image is None  # Not returned
        assert result.bounding_box is not None  # Metadata still available
    
    def test_extract_empty_mask(self, cropper_config, sample_document_image):
        """Test extraction with empty mask."""
        from ml.face_extraction.face_cropper import FaceCropper
        
        cropper = FaceCropper(cropper_config)
        
        empty_mask = np.zeros((768, 1024), dtype=np.uint8)
        
        result = cropper.extract(
            image=sample_document_image,
            mask=empty_mask,
            confidence=0.9,
            model_version="1.0.0",
            model_hash="testhash",
            return_face_image=True
        )
        
        assert not result.success
        assert "no face region" in result.error_message.lower()
    
    def test_extract_invalid_image(self, cropper_config, sample_face_mask):
        """Test extraction with invalid image."""
        from ml.face_extraction.face_cropper import FaceCropper
        
        cropper = FaceCropper(cropper_config)
        
        result = cropper.extract(
            image=None,
            mask=sample_face_mask,
            confidence=0.9,
            model_version="1.0.0",
            model_hash="testhash"
        )
        
        assert not result.success
        assert "invalid" in result.error_message.lower()
