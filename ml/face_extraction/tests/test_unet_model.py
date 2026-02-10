"""
Tests for U-Net face segmentation model.

This module tests the U-Net model for face segmentation including:
- Model initialization and loading
- Segmentation inference
- Confidence mapping
- Deterministic execution
- Model versioning and hashing
"""

import pytest
import numpy as np
import os

# Ensure deterministic execution
os.environ["CUDA_VISIBLE_DEVICES"] = "-1"
os.environ["TF_DETERMINISTIC_OPS"] = "1"


class TestUNetFaceSegmentor:
    """Tests for UNetFaceSegmentor class."""
    
    def test_initialization(self, unet_config):
        """Test model initialization."""
        from ml.face_extraction.unet_model import UNetFaceSegmentor
        
        segmentor = UNetFaceSegmentor(unet_config)
        
        assert segmentor is not None
        assert segmentor.model_version == unet_config.model_version
        assert not segmentor._initialized  # Lazy init
    
    def test_lazy_initialization(self, unet_config, sample_document_image):
        """Test that model initializes lazily on first use."""
        from ml.face_extraction.unet_model import UNetFaceSegmentor
        
        segmentor = UNetFaceSegmentor(unet_config)
        
        assert not segmentor._initialized
        
        # Trigger initialization via segmentation
        result = segmentor.segment_with_details(sample_document_image)
        
        assert segmentor._initialized
        assert segmentor._model is not None
    
    def test_segment_returns_mask(self, unet_config, sample_document_image):
        """Test that segment returns a mask of correct shape."""
        from ml.face_extraction.unet_model import UNetFaceSegmentor
        
        segmentor = UNetFaceSegmentor(unet_config)
        mask = segmentor.segment(sample_document_image)
        
        assert mask is not None
        assert mask.shape == sample_document_image.shape[:2]
        assert mask.dtype == np.float32 or mask.dtype == np.float64
        assert np.all(mask >= 0) and np.all(mask <= 1)
    
    def test_segment_with_details(self, unet_config, sample_document_image):
        """Test segmentation with detailed results."""
        from ml.face_extraction.unet_model import UNetFaceSegmentor, SegmentationResult
        
        segmentor = UNetFaceSegmentor(unet_config)
        result = segmentor.segment_with_details(sample_document_image)
        
        assert isinstance(result, SegmentationResult)
        assert result.success
        assert result.mask is not None
        assert result.confidence_map is not None
        assert result.model_version == unet_config.model_version
        assert len(result.model_hash) > 0
    
    def test_confidence_map(self, unet_config, sample_document_image):
        """Test confidence map extraction."""
        from ml.face_extraction.unet_model import UNetFaceSegmentor
        
        segmentor = UNetFaceSegmentor(unet_config)
        confidence_map = segmentor.get_confidence_map(sample_document_image)
        
        assert confidence_map is not None
        assert confidence_map.shape == sample_document_image.shape[:2]
        assert np.all(confidence_map >= 0) and np.all(confidence_map <= 1)
    
    def test_model_hash_consistency(self, unet_config, sample_document_image):
        """Test that model hash is consistent across calls."""
        from ml.face_extraction.unet_model import UNetFaceSegmentor
        
        segmentor = UNetFaceSegmentor(unet_config)
        
        # First call
        result1 = segmentor.segment_with_details(sample_document_image)
        hash1 = segmentor.model_hash
        
        # Second call
        result2 = segmentor.segment_with_details(sample_document_image)
        hash2 = segmentor.model_hash
        
        assert hash1 == hash2
        assert len(hash1) == 64  # SHA256 hex length
    
    def test_deterministic_output(self, unet_config, sample_document_image):
        """Test that segmentation is deterministic."""
        from ml.face_extraction.unet_model import UNetFaceSegmentor
        
        # Create two separate instances
        segmentor1 = UNetFaceSegmentor(unet_config)
        segmentor2 = UNetFaceSegmentor(unet_config)
        
        mask1 = segmentor1.segment(sample_document_image)
        mask2 = segmentor2.segment(sample_document_image)
        
        # Masks should be identical (deterministic)
        np.testing.assert_array_almost_equal(mask1, mask2, decimal=5)
    
    def test_invalid_image_handling(self, unet_config):
        """Test handling of invalid input images."""
        from ml.face_extraction.unet_model import UNetFaceSegmentor
        
        segmentor = UNetFaceSegmentor(unet_config)
        
        # None input
        result = segmentor.segment_with_details(None)
        assert not result.success
        assert result.error_message is not None
        
        # Empty array
        result = segmentor.segment_with_details(np.array([]))
        assert not result.success
    
    def test_grayscale_image_handling(self, unet_config, np_random):
        """Test handling of grayscale images."""
        from ml.face_extraction.unet_model import UNetFaceSegmentor
        
        segmentor = UNetFaceSegmentor(unet_config)
        
        # Create grayscale image
        gray_image = np_random.randint(0, 256, (768, 1024), dtype=np.uint8)
        
        # Should handle gracefully (either succeed or fail gracefully)
        result = segmentor.segment_with_details(gray_image)
        
        # The result should be valid (success or clean failure)
        assert result.mask is not None or result.error_message is not None
    
    def test_different_image_sizes(self, unet_config, np_random):
        """Test segmentation on different image sizes."""
        from ml.face_extraction.unet_model import UNetFaceSegmentor
        
        segmentor = UNetFaceSegmentor(unet_config)
        
        sizes = [
            (256, 256),
            (480, 640),
            (768, 1024),
            (1080, 1920),
        ]
        
        for h, w in sizes:
            image = np_random.randint(0, 256, (h, w, 3), dtype=np.uint8)
            result = segmentor.segment_with_details(image)
            
            assert result.mask.shape == (h, w), f"Failed for size {h}x{w}"
    
    def test_batch_segment(self, unet_config, sample_document_image, np_random):
        """Test batch segmentation."""
        from ml.face_extraction.unet_model import UNetFaceSegmentor
        
        segmentor = UNetFaceSegmentor(unet_config)
        
        # Create batch of images
        images = [
            sample_document_image,
            sample_document_image.copy(),
            np_random.randint(0, 256, (768, 1024, 3), dtype=np.uint8),
        ]
        
        results = segmentor.batch_segment(images)
        
        assert len(results) == len(images)
        for result in results:
            assert result.mask is not None
    
    def test_to_dict_serialization(self, unet_config, sample_document_image):
        """Test result serialization to dictionary."""
        from ml.face_extraction.unet_model import UNetFaceSegmentor
        
        segmentor = UNetFaceSegmentor(unet_config)
        result = segmentor.segment_with_details(sample_document_image)
        
        result_dict = result.to_dict()
        
        assert isinstance(result_dict, dict)
        assert "success" in result_dict
        assert "mean_confidence" in result_dict
        assert "max_confidence" in result_dict
        assert "model_version" in result_dict


class TestUNetModelBuilding:
    """Tests for U-Net model architecture building."""
    
    def test_unet_architecture_layers(self, unet_config):
        """Test that U-Net has expected layer structure."""
        from ml.face_extraction.unet_model import UNetFaceSegmentor
        
        segmentor = UNetFaceSegmentor(unet_config)
        segmentor._lazy_init()
        
        model = segmentor._model
        
        # Check model exists and has layers
        assert model is not None
        assert len(model.layers) > 0
        
        # Check input shape
        input_shape = model.input_shape
        assert input_shape[1:3] == unet_config.input_size
        
        # Check output shape
        output_shape = model.output_shape
        assert output_shape[1:3] == unet_config.input_size
        assert output_shape[3] == 1  # Single channel output
    
    def test_model_can_be_saved_loaded(self, unet_config, tmp_path):
        """Test that model can be saved and loaded."""
        from ml.face_extraction.unet_model import UNetFaceSegmentor
        from ml.face_extraction.config import UNetConfig
        
        # Create and initialize model
        segmentor = UNetFaceSegmentor(unet_config)
        segmentor._lazy_init()
        
        # Save model
        model_path = tmp_path / "unet_test.keras"
        segmentor._model.save(str(model_path))
        
        # Load in new instance
        load_config = UNetConfig(model_path=str(model_path))
        loaded_segmentor = UNetFaceSegmentor(load_config)
        loaded_segmentor._lazy_init()
        
        assert loaded_segmentor._model is not None


class TestUNetVaryingConditions:
    """Tests for U-Net under varying image conditions."""
    
    def test_low_quality_image(self, unet_config, low_quality_document):
        """Test segmentation on low quality image."""
        from ml.face_extraction.unet_model import UNetFaceSegmentor
        
        segmentor = UNetFaceSegmentor(unet_config)
        result = segmentor.segment_with_details(low_quality_document)
        
        # Should still produce a result
        assert result.mask is not None
        assert result.mask.shape == low_quality_document.shape[:2]
    
    def test_bright_image(self, unet_config, bright_document):
        """Test segmentation on overexposed image."""
        from ml.face_extraction.unet_model import UNetFaceSegmentor
        
        segmentor = UNetFaceSegmentor(unet_config)
        result = segmentor.segment_with_details(bright_document)
        
        assert result.mask is not None
    
    def test_dark_image(self, unet_config, dark_document):
        """Test segmentation on underexposed image."""
        from ml.face_extraction.unet_model import UNetFaceSegmentor
        
        segmentor = UNetFaceSegmentor(unet_config)
        result = segmentor.segment_with_details(dark_document)
        
        assert result.mask is not None
    
    def test_small_face_document(self, unet_config, small_face_document):
        """Test segmentation with very small face."""
        from ml.face_extraction.unet_model import UNetFaceSegmentor
        
        segmentor = UNetFaceSegmentor(unet_config)
        result = segmentor.segment_with_details(small_face_document)
        
        assert result.mask is not None
        # Confidence might be lower for small faces
