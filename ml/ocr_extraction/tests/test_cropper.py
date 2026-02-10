"""
Tests for ROI cropping functionality.
"""

import pytest
import numpy as np
import cv2

from ml.text_detection import TextROI, BoundingBox, Point, TextType
from ml.ocr_extraction.roi_cropper import ROICropper, CropResult
from ml.ocr_extraction.config import CropperConfig, ThresholdingMethod


class TestROICropper:
    """Tests for ROICropper class."""
    
    def test_initialization_default(self):
        """Test default initialization."""
        cropper = ROICropper()
        assert cropper.config is not None
        assert cropper.config.margin_pixels == 5
    
    def test_initialization_custom_config(self):
        """Test initialization with custom config."""
        config = CropperConfig(margin_pixels=10, margin_fraction=0.1)
        cropper = ROICropper(config)
        assert cropper.config.margin_pixels == 10
        assert cropper.config.margin_fraction == 0.1
    
    def test_crop_roi_basic(self, sample_image, sample_roi):
        """Test basic ROI cropping."""
        cropper = ROICropper()
        crop = cropper.crop_roi(sample_image, sample_roi)
        
        assert crop is not None
        assert len(crop.shape) == 2  # Grayscale
        assert crop.shape[0] > 0
        assert crop.shape[1] > 0
    
    def test_crop_roi_with_margin(self, sample_image, sample_roi):
        """Test cropping respects margin settings."""
        config = CropperConfig(margin_pixels=10)
        cropper = ROICropper(config)
        crop = cropper.crop_roi(sample_image, sample_roi)
        
        # Crop should be larger than ROI due to margin
        expected_min_height = int(sample_roi.bounding_box.height)
        assert crop.shape[0] >= expected_min_height
    
    def test_crop_roi_color_image(self, sample_color_image, sample_roi):
        """Test cropping from color image."""
        cropper = ROICropper()
        crop = cropper.crop_roi(sample_color_image, sample_roi)
        
        assert crop is not None
        assert len(crop.shape) == 3  # Color preserved in crop
    
    def test_crop_roi_bounds_checking(self, sample_image):
        """Test cropping handles boundary ROIs."""
        # ROI at edge of image
        edge_roi = TextROI.create(
            bounding_box=BoundingBox(x=0, y=0, width=50, height=20),
            confidence=0.9,
            text_type=TextType.LINE,
        )
        
        cropper = ROICropper()
        crop = cropper.crop_roi(sample_image, edge_roi)
        
        assert crop is not None
        assert crop.shape[0] > 0
        assert crop.shape[1] > 0
    
    def test_crop_rotated_roi(self, sample_image, rotated_roi):
        """Test cropping rotated ROI with perspective transform."""
        config = CropperConfig(use_subpixel_crop=True)
        cropper = ROICropper(config)
        crop = cropper.crop_roi(sample_image, rotated_roi)
        
        assert crop is not None
        assert crop.shape[0] > 0
    
    def test_prepare_for_ocr_grayscale(self, sample_color_image, sample_roi):
        """Test OCR preparation converts to grayscale."""
        config = CropperConfig(convert_grayscale=True)
        cropper = ROICropper(config)
        
        crop = cropper.crop_roi(sample_color_image, sample_roi)
        prepared = cropper.prepare_for_ocr(crop)
        
        # Should be grayscale after preparation
        assert len(prepared.shape) == 2
    
    def test_prepare_for_ocr_thresholding_otsu(self, sample_image, sample_roi):
        """Test Otsu thresholding in OCR preparation."""
        config = CropperConfig(thresholding=ThresholdingMethod.OTSU)
        cropper = ROICropper(config)
        
        crop = cropper.crop_roi(sample_image, sample_roi)
        prepared = cropper.prepare_for_ocr(crop)
        
        # Should only have black and white pixels after thresholding
        unique_values = np.unique(prepared)
        assert len(unique_values) <= 2
    
    def test_prepare_for_ocr_thresholding_adaptive(self, sample_image, sample_roi):
        """Test adaptive thresholding."""
        config = CropperConfig(thresholding=ThresholdingMethod.ADAPTIVE_GAUSSIAN)
        cropper = ROICropper(config)
        
        crop = cropper.crop_roi(sample_image, sample_roi)
        prepared = cropper.prepare_for_ocr(crop)
        
        assert prepared is not None
        unique_values = np.unique(prepared)
        assert len(unique_values) <= 2
    
    def test_prepare_for_ocr_scaling(self, sample_image, sample_roi):
        """Test scaling to target height."""
        target_height = 64
        config = CropperConfig(scale_to_height=target_height)
        cropper = ROICropper(config)
        
        crop = cropper.crop_roi(sample_image, sample_roi)
        prepared = cropper.prepare_for_ocr(crop)
        
        assert prepared.shape[0] == target_height
    
    def test_prepare_for_ocr_no_scaling(self, sample_image, sample_roi):
        """Test no scaling when disabled."""
        config = CropperConfig(scale_to_height=None)
        cropper = ROICropper(config)
        
        crop = cropper.crop_roi(sample_image, sample_roi)
        original_height = crop.shape[0]
        prepared = cropper.prepare_for_ocr(crop)
        
        # Height should be preserved (modulo thresholding effects)
        assert prepared.shape[0] == original_height
    
    def test_crop_and_prepare_success(self, sample_image, sample_roi):
        """Test combined crop and prepare operation."""
        cropper = ROICropper()
        result = cropper.crop_and_prepare(sample_image, sample_roi)
        
        assert isinstance(result, CropResult)
        assert result.success is True
        assert result.roi_id == sample_roi.roi_id
        assert result.original_crop is not None
        assert result.processed_crop is not None
    
    def test_crop_and_prepare_failure_small_roi(self, sample_image):
        """Test handling of too-small ROIs."""
        tiny_roi = TextROI.create(
            bounding_box=BoundingBox(x=50, y=50, width=5, height=3),
            confidence=0.9,
            text_type=TextType.CHARACTER,
        )
        
        config = CropperConfig(min_crop_width=20, min_crop_height=10)
        cropper = ROICropper(config)
        result = cropper.crop_and_prepare(sample_image, tiny_roi)
        
        assert result.success is False
        assert "too small" in result.error_message.lower()
    
    def test_crop_all_rois(self, sample_document_image, sample_rois):
        """Test batch cropping of multiple ROIs."""
        cropper = ROICropper()
        results = cropper.crop_all_rois(sample_document_image, sample_rois)
        
        assert len(results) == len(sample_rois)
        assert all(isinstance(r, CropResult) for r in results)
    
    def test_empty_crop_handling(self):
        """Test handling of empty/invalid crops."""
        cropper = ROICropper()
        
        with pytest.raises(ValueError, match="Empty crop"):
            cropper.prepare_for_ocr(np.array([]))
    
    def test_order_points(self, sample_image):
        """Test polygon point ordering."""
        cropper = ROICropper()
        
        # Unordered points
        pts = np.array([
            [100, 50],  # top-right
            [10, 40],   # top-left
            [110, 70],  # bottom-right
            [20, 60],   # bottom-left
        ], dtype=np.float32)
        
        ordered = cropper._order_points(pts)
        
        # Should be ordered: TL, TR, BR, BL
        assert ordered[0][0] < ordered[1][0]  # TL.x < TR.x
        assert ordered[0][1] < ordered[3][1]  # TL.y < BL.y
    
    def test_deskew_straight_image(self, sample_image, sample_roi):
        """Test deskewing doesn't affect straight images significantly."""
        config = CropperConfig(deskew_enabled=True, deskew_max_angle=5.0)
        cropper = ROICropper(config)
        
        crop = cropper.crop_roi(sample_image, sample_roi)
        deskewed = cropper._deskew(crop)
        
        assert deskewed.shape == crop.shape


class TestCropperConfig:
    """Tests for CropperConfig."""
    
    def test_default_values(self):
        """Test default configuration values."""
        config = CropperConfig()
        assert config.margin_pixels == 5
        assert config.use_subpixel_crop is True
        assert config.convert_grayscale is True
        assert config.thresholding == ThresholdingMethod.OTSU
    
    def test_custom_values(self):
        """Test custom configuration values."""
        config = CropperConfig(
            margin_pixels=20,
            margin_fraction=0.15,
            thresholding=ThresholdingMethod.ADAPTIVE_MEAN,
            scale_to_height=48,
        )
        assert config.margin_pixels == 20
        assert config.margin_fraction == 0.15
        assert config.scale_to_height == 48
