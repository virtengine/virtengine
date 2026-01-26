"""
Main face extraction pipeline.

This module orchestrates the complete face extraction workflow:
1. U-Net segmentation of face region
2. Mask post-processing
3. Face bounding box extraction and cropping
4. Embedding extraction (with data minimization)
"""

import logging
import time
import hashlib
from dataclasses import dataclass, field
from typing import Union, Optional, List

import numpy as np

from ml.face_extraction.config import FaceExtractionConfig
from ml.face_extraction.unet_model import UNetFaceSegmentor, SegmentationResult
from ml.face_extraction.mask_processing import MaskProcessor, MaskProcessingResult
from ml.face_extraction.face_cropper import FaceCropper, FaceExtractionResult, BoundingBox
from ml.face_extraction.embedding_extractor import (
    DocumentFaceEmbeddingExtractor,
    EmbeddingResult,
)

logger = logging.getLogger(__name__)


@dataclass
class PipelineResult:
    """Complete result of face extraction pipeline."""
    
    # Primary outputs
    face_extraction: Optional[FaceExtractionResult] = None
    embedding: Optional[EmbeddingResult] = None
    
    # Intermediate results (if requested)
    segmentation: Optional[SegmentationResult] = None
    mask_processing: Optional[MaskProcessingResult] = None
    
    # Metadata
    success: bool = True
    error_message: Optional[str] = None
    processing_time_ms: float = 0.0
    stages_completed: List[str] = field(default_factory=list)
    
    def to_dict(self) -> dict:
        """Convert to dictionary for serialization."""
        return {
            "success": self.success,
            "error_message": self.error_message,
            "processing_time_ms": self.processing_time_ms,
            "stages_completed": self.stages_completed,
            "face_extraction": self.face_extraction.to_dict() if self.face_extraction else None,
            "embedding": self.embedding.to_dict() if self.embedding else None,
        }


class FaceExtractionPipeline:
    """
    Complete face extraction pipeline.
    
    This class orchestrates all stages of face extraction from identity
    documents, including segmentation, mask processing, cropping, and
    optional embedding extraction.
    
    Data Minimization:
    By default, this pipeline returns only the embedding (if configured),
    not the face image. This follows privacy-by-design principles.
    To get the face image, explicitly set return_face_image=True.
    
    Processing Steps:
    1. U-Net segmentation - Identify face region in document
    2. Mask processing - Clean up segmentation mask
    3. Face cropping - Extract face region
    4. Embedding extraction (optional) - Compute face embedding
    """
    
    def __init__(self, config: Optional[FaceExtractionConfig] = None):
        """
        Initialize the face extraction pipeline.
        
        Args:
            config: Pipeline configuration. Uses defaults if not provided.
        """
        self.config = config or FaceExtractionConfig()
        
        # Initialize components
        self.segmentor = UNetFaceSegmentor(self.config.unet)
        self.mask_processor = MaskProcessor(self.config.mask)
        self.cropper = FaceCropper(self.config.cropper)
        self.embedder = DocumentFaceEmbeddingExtractor(
            config=self.config.embedding
        )
        
        logger.info("Face extraction pipeline initialized")
    
    def _segment(self, image: np.ndarray) -> SegmentationResult:
        """Run segmentation stage."""
        return self.segmentor.segment_with_details(image)
    
    def _process_mask(self, mask: np.ndarray) -> MaskProcessingResult:
        """Run mask processing stage."""
        return self.mask_processor.process(mask)
    
    def _extract_face(
        self, 
        image: np.ndarray, 
        mask: np.ndarray,
        confidence: float,
        model_version: str,
        model_hash: str,
        return_face_image: bool
    ) -> FaceExtractionResult:
        """Run face extraction stage."""
        return self.cropper.extract(
            image=image,
            mask=mask,
            confidence=confidence,
            model_version=model_version,
            model_hash=model_hash,
            return_face_image=return_face_image
        )
    
    def _compute_embedding(
        self, 
        face_image: np.ndarray,
        segmentation_confidence: float
    ) -> EmbeddingResult:
        """Run embedding extraction stage."""
        return self.embedder.extract_embedding_only(
            face_image,
            segmentation_confidence=segmentation_confidence
        )
    
    def _fallback_detection(self, image: np.ndarray) -> Optional[FaceExtractionResult]:
        """
        Fallback to face detection if segmentation fails.
        
        Uses the facial verification face detector if available.
        """
        if not self.config.fallback_to_detection:
            return None
        
        try:
            from ml.facial_verification import FaceDetector, DetectionConfig
            
            detector = FaceDetector(DetectionConfig())
            detections = detector.detect(image)
            
            if not detections:
                return None
            
            # Use largest detection
            largest = max(detections, key=lambda d: d.bbox.area if hasattr(d.bbox, 'area') else 0)
            
            # Convert to our BoundingBox format
            bbox = BoundingBox(
                x=largest.bbox.x,
                y=largest.bbox.y,
                width=largest.bbox.width,
                height=largest.bbox.height
            )
            
            # Crop face
            face_image = self.cropper.crop_face(image, bbox)
            
            return FaceExtractionResult(
                face_image=face_image,
                bounding_box=bbox,
                mask=np.zeros(image.shape[:2], dtype=np.uint8),
                confidence=largest.confidence,
                model_version="fallback_detection",
                model_hash="",
                success=True,
            )
            
        except Exception as e:
            logger.warning(f"Fallback detection failed: {e}")
            return None
    
    def extract(
        self, 
        document_image: np.ndarray,
        return_face_image: bool = False
    ) -> Union[FaceExtractionResult, EmbeddingResult]:
        """
        Extract face or embedding from document image.
        
        By default, returns embedding only (data minimization).
        Set return_face_image=True to get the face image.
        
        Args:
            document_image: Input document image (H, W, C) in BGR format
            return_face_image: Whether to return face image (default: False)
            
        Returns:
            EmbeddingResult if return_face_image=False
            FaceExtractionResult if return_face_image=True
        """
        result = self.extract_full(document_image, return_face_image)
        
        if return_face_image:
            return result.face_extraction or FaceExtractionResult(
                face_image=None,
                bounding_box=BoundingBox(0, 0, 0, 0),
                mask=np.zeros((1, 1), dtype=np.uint8),
                confidence=0.0,
                model_version="",
                model_hash="",
                success=False,
                error_message=result.error_message
            )
        else:
            return result.embedding or EmbeddingResult(
                embedding=None,
                embedding_hash="",
                confidence=0.0,
                model_version="",
                success=False,
                error_message=result.error_message
            )
    
    def extract_full(
        self, 
        document_image: np.ndarray,
        return_face_image: bool = False
    ) -> PipelineResult:
        """
        Run full extraction pipeline with detailed results.
        
        Args:
            document_image: Input document image (H, W, C) in BGR format
            return_face_image: Whether to include face image in result
            
        Returns:
            PipelineResult with all intermediate results
        """
        start_time = time.time()
        stages_completed = []
        
        if document_image is None or document_image.size == 0:
            return PipelineResult(
                success=False,
                error_message="Invalid input image",
                processing_time_ms=0.0,
            )
        
        try:
            # Stage 1: Segmentation
            logger.debug("Running segmentation...")
            seg_result = self._segment(document_image)
            stages_completed.append("segmentation")
            
            if not seg_result.success:
                # Try fallback detection
                if self.config.fallback_to_detection:
                    logger.info("Segmentation failed, trying fallback detection...")
                    fallback_result = self._fallback_detection(document_image)
                    
                    if fallback_result and fallback_result.success:
                        stages_completed.append("fallback_detection")
                        
                        # Compute embedding if face image available
                        embedding_result = None
                        if fallback_result.face_image is not None:
                            embedding_result = self._compute_embedding(
                                fallback_result.face_image,
                                segmentation_confidence=fallback_result.confidence
                            )
                            stages_completed.append("embedding")
                            
                            # Data minimization - discard face if not needed
                            if not return_face_image:
                                fallback_result.face_image = None
                        
                        elapsed_ms = (time.time() - start_time) * 1000
                        
                        return PipelineResult(
                            face_extraction=fallback_result if return_face_image else None,
                            embedding=embedding_result,
                            success=True,
                            processing_time_ms=elapsed_ms,
                            stages_completed=stages_completed,
                        )
                
                elapsed_ms = (time.time() - start_time) * 1000
                return PipelineResult(
                    segmentation=seg_result if self.config.return_intermediate_results else None,
                    success=False,
                    error_message=f"Segmentation failed: {seg_result.error_message}",
                    processing_time_ms=elapsed_ms,
                    stages_completed=stages_completed,
                )
            
            # Check confidence threshold
            if seg_result.mean_confidence < self.config.min_extraction_confidence:
                logger.warning(f"Low segmentation confidence: {seg_result.mean_confidence:.2f}")
            
            # Stage 2: Mask processing
            logger.debug("Processing mask...")
            mask_result = self._process_mask(seg_result.mask)
            stages_completed.append("mask_processing")
            
            if not mask_result.success:
                elapsed_ms = (time.time() - start_time) * 1000
                return PipelineResult(
                    segmentation=seg_result if self.config.return_intermediate_results else None,
                    mask_processing=mask_result if self.config.return_intermediate_results else None,
                    success=False,
                    error_message=f"Mask processing failed: {mask_result.error_message}",
                    processing_time_ms=elapsed_ms,
                    stages_completed=stages_completed,
                )
            
            # Stage 3: Face extraction
            logger.debug("Extracting face region...")
            # We need face image temporarily for embedding
            need_face_for_embedding = not return_face_image and self.config.embedding.use_facial_verification
            
            extraction_result = self._extract_face(
                image=document_image,
                mask=mask_result.processed_mask,
                confidence=seg_result.mean_confidence,
                model_version=seg_result.model_version,
                model_hash=seg_result.model_hash,
                return_face_image=return_face_image or need_face_for_embedding
            )
            stages_completed.append("face_extraction")
            
            if not extraction_result.success:
                elapsed_ms = (time.time() - start_time) * 1000
                return PipelineResult(
                    segmentation=seg_result if self.config.return_intermediate_results else None,
                    mask_processing=mask_result if self.config.return_intermediate_results else None,
                    face_extraction=extraction_result,
                    success=False,
                    error_message=f"Face extraction failed: {extraction_result.error_message}",
                    processing_time_ms=elapsed_ms,
                    stages_completed=stages_completed,
                )
            
            # Stage 4: Embedding (optional)
            embedding_result = None
            if extraction_result.face_image is not None:
                logger.debug("Extracting embedding...")
                embedding_result = self._compute_embedding(
                    extraction_result.face_image,
                    segmentation_confidence=seg_result.mean_confidence
                )
                stages_completed.append("embedding")
                
                # DATA MINIMIZATION: Discard face image if not explicitly requested
                if not return_face_image:
                    extraction_result.face_image = None
            
            elapsed_ms = (time.time() - start_time) * 1000
            
            logger.info(f"Face extraction completed in {elapsed_ms:.1f}ms. "
                       f"Stages: {stages_completed}")
            
            return PipelineResult(
                face_extraction=extraction_result if return_face_image else None,
                embedding=embedding_result,
                segmentation=seg_result if self.config.return_intermediate_results else None,
                mask_processing=mask_result if self.config.return_intermediate_results else None,
                success=True,
                processing_time_ms=elapsed_ms,
                stages_completed=stages_completed,
            )
            
        except Exception as e:
            logger.error(f"Pipeline failed: {e}")
            elapsed_ms = (time.time() - start_time) * 1000
            return PipelineResult(
                success=False,
                error_message=str(e),
                processing_time_ms=elapsed_ms,
                stages_completed=stages_completed,
            )
    
    def extract_embedding(
        self, 
        document_image: np.ndarray
    ) -> EmbeddingResult:
        """
        Convenience method to extract only embedding.
        
        Face image is never returned - strict data minimization.
        
        Args:
            document_image: Input document image
            
        Returns:
            EmbeddingResult
        """
        return self.extract(document_image, return_face_image=False)
    
    def extract_face(
        self, 
        document_image: np.ndarray
    ) -> FaceExtractionResult:
        """
        Convenience method to extract face image.
        
        Use only when face image is explicitly needed.
        
        Args:
            document_image: Input document image
            
        Returns:
            FaceExtractionResult with face image
        """
        return self.extract(document_image, return_face_image=True)
    
    def get_model_info(self) -> dict:
        """Get information about loaded models."""
        return {
            "segmentor": {
                "model_version": self.segmentor.model_version,
                "model_hash": self.segmentor.model_hash,
            },
            "config": {
                "min_extraction_confidence": self.config.min_extraction_confidence,
                "fallback_to_detection": self.config.fallback_to_detection,
                "data_minimization": self.config.embedding.discard_face_image,
            }
        }
