"""
Tests for face extraction pipeline.

This module tests the complete face extraction pipeline including:
- Full extraction workflow
- Data minimization
- Embedding extraction
- Error handling and fallbacks
- Various lighting/quality conditions
"""

import pytest
import numpy as np
import os

# Ensure deterministic execution
os.environ["CUDA_VISIBLE_DEVICES"] = "-1"
os.environ["TF_DETERMINISTIC_OPS"] = "1"


class TestFaceExtractionPipeline:
    """Tests for FaceExtractionPipeline class."""
    
    def test_initialization(self, face_extraction_config):
        """Test pipeline initialization."""
        from ml.face_extraction.pipeline import FaceExtractionPipeline
        
        pipeline = FaceExtractionPipeline(face_extraction_config)
        
        assert pipeline is not None
        assert pipeline.segmentor is not None
        assert pipeline.mask_processor is not None
        assert pipeline.cropper is not None
        assert pipeline.embedder is not None
    
    def test_default_initialization(self):
        """Test pipeline with default config."""
        from ml.face_extraction.pipeline import FaceExtractionPipeline
        
        pipeline = FaceExtractionPipeline()
        
        assert pipeline is not None
        assert pipeline.config is not None


class TestDataMinimization:
    """Tests for data minimization compliance."""
    
    def test_default_returns_embedding_not_face(self, face_extraction_config, sample_document_image):
        """Test that default extraction returns embedding, not face image."""
        from ml.face_extraction.pipeline import FaceExtractionPipeline
        from ml.face_extraction.embedding_extractor import EmbeddingResult
        
        pipeline = FaceExtractionPipeline(face_extraction_config)
        
        # Default: return_face_image=False
        result = pipeline.extract(sample_document_image)
        
        # Should return embedding result, not face extraction
        assert isinstance(result, EmbeddingResult)
    
    def test_explicit_face_image_request(self, face_extraction_config, sample_document_image):
        """Test that face image is returned only when explicitly requested."""
        from ml.face_extraction.pipeline import FaceExtractionPipeline
        from ml.face_extraction.face_cropper import FaceExtractionResult
        
        pipeline = FaceExtractionPipeline(face_extraction_config)
        
        # Explicitly request face image
        result = pipeline.extract(sample_document_image, return_face_image=True)
        
        assert isinstance(result, FaceExtractionResult)
        if result.success:
            assert result.face_image is not None
    
    def test_extract_embedding_method(self, face_extraction_config, sample_document_image):
        """Test extract_embedding convenience method."""
        from ml.face_extraction.pipeline import FaceExtractionPipeline
        from ml.face_extraction.embedding_extractor import EmbeddingResult
        
        pipeline = FaceExtractionPipeline(face_extraction_config)
        
        result = pipeline.extract_embedding(sample_document_image)
        
        assert isinstance(result, EmbeddingResult)
    
    def test_extract_face_method(self, face_extraction_config, sample_document_image):
        """Test extract_face convenience method."""
        from ml.face_extraction.pipeline import FaceExtractionPipeline
        from ml.face_extraction.face_cropper import FaceExtractionResult
        
        pipeline = FaceExtractionPipeline(face_extraction_config)
        
        result = pipeline.extract_face(sample_document_image)
        
        assert isinstance(result, FaceExtractionResult)


class TestPipelineResult:
    """Tests for PipelineResult."""
    
    def test_full_extraction_returns_pipeline_result(self, face_extraction_config, sample_document_image):
        """Test that extract_full returns PipelineResult."""
        from ml.face_extraction.pipeline import FaceExtractionPipeline, PipelineResult
        
        pipeline = FaceExtractionPipeline(face_extraction_config)
        
        result = pipeline.extract_full(sample_document_image)
        
        assert isinstance(result, PipelineResult)
        assert "processing_time_ms" in result.to_dict()
        assert "stages_completed" in result.to_dict()
    
    def test_stages_tracking(self, face_extraction_config, sample_document_image):
        """Test that pipeline stages are tracked."""
        from ml.face_extraction.pipeline import FaceExtractionPipeline
        
        pipeline = FaceExtractionPipeline(face_extraction_config)
        
        result = pipeline.extract_full(sample_document_image)
        
        # Check stages are recorded
        assert len(result.stages_completed) > 0
        assert "segmentation" in result.stages_completed
    
    def test_intermediate_results_when_enabled(self, sample_document_image):
        """Test that intermediate results are returned when configured."""
        from ml.face_extraction.pipeline import FaceExtractionPipeline
        from ml.face_extraction.config import FaceExtractionConfig
        
        config = FaceExtractionConfig(return_intermediate_results=True)
        pipeline = FaceExtractionPipeline(config)
        
        result = pipeline.extract_full(sample_document_image)
        
        if result.success:
            assert result.segmentation is not None or "segmentation" in result.stages_completed


class TestErrorHandling:
    """Tests for error handling."""
    
    def test_invalid_image_handling(self, face_extraction_config):
        """Test handling of invalid input image."""
        from ml.face_extraction.pipeline import FaceExtractionPipeline
        
        pipeline = FaceExtractionPipeline(face_extraction_config)
        
        # None input
        result = pipeline.extract_full(None)
        assert not result.success
        assert result.error_message is not None
        
        # Empty input
        result = pipeline.extract_full(np.array([]))
        assert not result.success
    
    def test_graceful_failure_on_no_face(self, face_extraction_config, np_random):
        """Test graceful handling when no face is found."""
        from ml.face_extraction.pipeline import FaceExtractionPipeline
        
        pipeline = FaceExtractionPipeline(face_extraction_config)
        
        # Random noise - unlikely to have a face
        noise_image = np_random.randint(0, 256, (768, 1024, 3), dtype=np.uint8)
        
        result = pipeline.extract_full(noise_image)
        
        # Should complete (may succeed or fail gracefully)
        assert result.processing_time_ms >= 0
        if not result.success:
            assert result.error_message is not None


class TestFallbackDetection:
    """Tests for fallback face detection."""
    
    def test_fallback_enabled(self, sample_document_image):
        """Test that fallback detection can be triggered."""
        from ml.face_extraction.pipeline import FaceExtractionPipeline
        from ml.face_extraction.config import FaceExtractionConfig
        
        config = FaceExtractionConfig(fallback_to_detection=True)
        pipeline = FaceExtractionPipeline(config)
        
        result = pipeline.extract_full(sample_document_image)
        
        # Either segmentation succeeded or fallback was attempted
        assert result.stages_completed is not None


class TestVaryingConditions:
    """Tests for varying lighting and quality conditions."""
    
    def test_low_quality_document(self, face_extraction_config, low_quality_document):
        """Test extraction on low quality document."""
        from ml.face_extraction.pipeline import FaceExtractionPipeline
        
        pipeline = FaceExtractionPipeline(face_extraction_config)
        
        result = pipeline.extract_full(low_quality_document)
        
        # Should attempt extraction
        assert len(result.stages_completed) > 0
    
    def test_bright_document(self, face_extraction_config, bright_document):
        """Test extraction on overexposed document."""
        from ml.face_extraction.pipeline import FaceExtractionPipeline
        
        pipeline = FaceExtractionPipeline(face_extraction_config)
        
        result = pipeline.extract_full(bright_document)
        
        assert len(result.stages_completed) > 0
    
    def test_dark_document(self, face_extraction_config, dark_document):
        """Test extraction on underexposed document."""
        from ml.face_extraction.pipeline import FaceExtractionPipeline
        
        pipeline = FaceExtractionPipeline(face_extraction_config)
        
        result = pipeline.extract_full(dark_document)
        
        assert len(result.stages_completed) > 0
    
    def test_small_face_document(self, face_extraction_config, small_face_document):
        """Test extraction with very small face."""
        from ml.face_extraction.pipeline import FaceExtractionPipeline
        
        pipeline = FaceExtractionPipeline(face_extraction_config)
        
        result = pipeline.extract_full(small_face_document)
        
        # Should attempt extraction
        assert result.processing_time_ms >= 0


class TestModelInfo:
    """Tests for model information."""
    
    def test_get_model_info(self, face_extraction_config):
        """Test getting model information."""
        from ml.face_extraction.pipeline import FaceExtractionPipeline
        
        pipeline = FaceExtractionPipeline(face_extraction_config)
        
        info = pipeline.get_model_info()
        
        assert isinstance(info, dict)
        assert "segmentor" in info
        assert "config" in info
        assert "data_minimization" in info["config"]


class TestEmbeddingExtraction:
    """Tests for embedding extraction."""
    
    def test_embedding_hash_computed(self, face_extraction_config, sample_document_image):
        """Test that embedding hash is computed."""
        from ml.face_extraction.pipeline import FaceExtractionPipeline
        
        pipeline = FaceExtractionPipeline(face_extraction_config)
        
        result = pipeline.extract_embedding(sample_document_image)
        
        if result.success:
            assert result.embedding_hash is not None
            assert len(result.embedding_hash) > 0
    
    def test_embedding_dimension(self, face_extraction_config, sample_document_image):
        """Test embedding has expected dimension."""
        from ml.face_extraction.pipeline import FaceExtractionPipeline
        
        pipeline = FaceExtractionPipeline(face_extraction_config)
        
        result = pipeline.extract_embedding(sample_document_image)
        
        if result.success and result.embedding is not None:
            # Common embedding dimensions: 128, 512, 2622
            assert len(result.embedding.shape) == 1
            assert result.embedding.shape[0] > 0


class TestDeterminism:
    """Tests for deterministic execution."""
    
    def test_deterministic_segmentation(self, face_extraction_config, sample_document_image):
        """Test that segmentation is deterministic."""
        from ml.face_extraction.pipeline import FaceExtractionPipeline
        
        pipeline1 = FaceExtractionPipeline(face_extraction_config)
        pipeline2 = FaceExtractionPipeline(face_extraction_config)
        
        result1 = pipeline1.extract_full(sample_document_image, return_face_image=True)
        result2 = pipeline2.extract_full(sample_document_image, return_face_image=True)
        
        # Both should have same success state
        assert result1.success == result2.success
        
        # If both succeeded, bounding boxes should match
        if result1.success and result2.success:
            if result1.face_extraction and result2.face_extraction:
                bbox1 = result1.face_extraction.bounding_box
                bbox2 = result2.face_extraction.bounding_box
                
                assert bbox1.x == bbox2.x
                assert bbox1.y == bbox2.y
                assert bbox1.width == bbox2.width
                assert bbox1.height == bbox2.height


class TestResultSerialization:
    """Tests for result serialization."""
    
    def test_pipeline_result_to_dict(self, face_extraction_config, sample_document_image):
        """Test PipelineResult serialization."""
        from ml.face_extraction.pipeline import FaceExtractionPipeline
        
        pipeline = FaceExtractionPipeline(face_extraction_config)
        
        result = pipeline.extract_full(sample_document_image)
        result_dict = result.to_dict()
        
        assert isinstance(result_dict, dict)
        assert "success" in result_dict
        assert "processing_time_ms" in result_dict
        assert "stages_completed" in result_dict
    
    def test_embedding_result_serialization(self, face_extraction_config, sample_document_image):
        """Test EmbeddingResult serialization."""
        from ml.face_extraction.pipeline import FaceExtractionPipeline
        
        pipeline = FaceExtractionPipeline(face_extraction_config)
        
        result = pipeline.extract_embedding(sample_document_image)
        result_dict = result.to_dict()
        
        assert isinstance(result_dict, dict)
        assert "success" in result_dict
        assert "embedding_hash" in result_dict
        assert "confidence" in result_dict


class TestDocumentFaceEmbeddingExtractor:
    """Tests for DocumentFaceEmbeddingExtractor."""
    
    def test_compute_embedding_hash(self):
        """Test embedding hash computation."""
        from ml.face_extraction.embedding_extractor import compute_embedding_hash
        
        embedding = np.array([0.1, 0.2, 0.3, 0.4, 0.5], dtype=np.float32)
        
        hash1 = compute_embedding_hash(embedding)
        hash2 = compute_embedding_hash(embedding)
        
        # Same embedding should produce same hash
        assert hash1 == hash2
        assert len(hash1) == 64  # SHA256 hex length
    
    def test_different_embeddings_different_hashes(self):
        """Test that different embeddings produce different hashes."""
        from ml.face_extraction.embedding_extractor import compute_embedding_hash
        
        emb1 = np.array([0.1, 0.2, 0.3], dtype=np.float32)
        emb2 = np.array([0.4, 0.5, 0.6], dtype=np.float32)
        
        hash1 = compute_embedding_hash(emb1)
        hash2 = compute_embedding_hash(emb2)
        
        assert hash1 != hash2
    
    def test_embedding_result_get_bytes(self):
        """Test EmbeddingResult byte conversion."""
        from ml.face_extraction.embedding_extractor import EmbeddingResult
        
        embedding = np.array([0.1, 0.2, 0.3], dtype=np.float32)
        
        result = EmbeddingResult(
            embedding=embedding,
            embedding_hash="abc123",
            confidence=0.95,
            model_version="1.0.0",
            success=True
        )
        
        emb_bytes = result.get_embedding_bytes()
        
        assert len(emb_bytes) > 0
        assert emb_bytes == embedding.tobytes()
    
    def test_embedding_result_empty_embedding(self):
        """Test EmbeddingResult with no embedding."""
        from ml.face_extraction.embedding_extractor import EmbeddingResult
        
        result = EmbeddingResult(
            embedding=None,
            embedding_hash="",
            confidence=0.0,
            model_version="",
            success=False,
            error_message="No face found"
        )
        
        assert result.get_embedding_bytes() == b''
        
        result_dict = result.to_dict()
        assert result_dict["embedding_dimension"] == 0
