"""
Document face embedding extraction module.

This module provides embedding extraction from document faces with
data minimization - face images are not retained after embedding
extraction for privacy compliance.
"""

import logging
import hashlib
from dataclasses import dataclass
from typing import Optional, Tuple, Any

import numpy as np

from ml.face_extraction.config import FaceExtractionConfig, EmbeddingConfig
from ml.face_extraction.unet_model import UNetFaceSegmentor
from ml.face_extraction.mask_processing import MaskProcessor
from ml.face_extraction.face_cropper import FaceCropper, FaceExtractionResult

logger = logging.getLogger(__name__)


@dataclass
class EmbeddingResult:
    """Result of embedding extraction."""
    
    embedding: Optional[np.ndarray]
    embedding_hash: str
    confidence: float
    model_version: str
    success: bool = True
    error_message: Optional[str] = None
    segmentation_confidence: float = 0.0
    extraction_confidence: float = 0.0
    
    def to_dict(self) -> dict:
        """Convert to dictionary for serialization."""
        return {
            "success": self.success,
            "embedding_hash": self.embedding_hash,
            "confidence": self.confidence,
            "model_version": self.model_version,
            "embedding_dimension": len(self.embedding) if self.embedding is not None else 0,
            "segmentation_confidence": self.segmentation_confidence,
            "extraction_confidence": self.extraction_confidence,
            "error_message": self.error_message,
        }
    
    def get_embedding_bytes(self) -> bytes:
        """Get embedding as bytes for hashing/comparison."""
        if self.embedding is None:
            return b''
        return self.embedding.tobytes()


def compute_embedding_hash(embedding: np.ndarray, algorithm: str = "sha256") -> str:
    """
    Compute hash of embedding vector.
    
    Args:
        embedding: Embedding vector
        algorithm: Hash algorithm (sha256, sha512, etc.)
        
    Returns:
        Hex string hash
    """
    if embedding is None:
        return ""
    
    # Round to reduce floating-point precision issues
    rounded = np.round(embedding, decimals=6)
    
    hasher = hashlib.new(algorithm)
    hasher.update(rounded.tobytes())
    return hasher.hexdigest()


class DocumentFaceEmbeddingExtractor:
    """
    Face embedding extraction from identity documents.
    
    This class provides end-to-end embedding extraction with data
    minimization - face images are not stored or returned.
    
    Data Minimization Policy:
    - Face images are extracted temporarily for embedding computation
    - After embedding is computed, face image is immediately discarded
    - Only the embedding vector is retained/returned
    - This complies with privacy-by-design principles
    """
    
    def __init__(
        self, 
        config: Optional[EmbeddingConfig] = None,
        face_embedder: Any = None
    ):
        """
        Initialize the embedding extractor.
        
        Args:
            config: Embedding configuration
            face_embedder: FaceEmbedder instance from facial_verification module
        """
        self.config = config or EmbeddingConfig()
        self._embedder = face_embedder
        self._initialized = False
    
    def _lazy_init(self) -> None:
        """Lazy initialization of the embedder."""
        if self._initialized:
            return
        
        if self._embedder is None and self.config.use_facial_verification:
            try:
                from ml.facial_verification import FaceEmbedder
                self._embedder = FaceEmbedder()
                logger.info("Initialized FaceEmbedder from facial_verification module")
            except ImportError as e:
                logger.warning(f"Could not import FaceEmbedder: {e}")
                raise RuntimeError("FaceEmbedder not available")
        
        self._initialized = True
    
    def _compute_embedding(self, face_image: np.ndarray) -> Tuple[Optional[np.ndarray], str, float]:
        """
        Compute embedding from face image.
        
        Args:
            face_image: Cropped face image
            
        Returns:
            Tuple of (embedding, model_version, confidence)
        """
        self._lazy_init()
        
        if self._embedder is None:
            raise RuntimeError("No embedder available")
        
        # Use facial verification embedder
        result = self._embedder.embed(face_image)
        
        if result.success and result.embedding is not None:
            embedding = result.embedding
            
            # Normalize if configured
            if self.config.normalize_embedding:
                norm = np.linalg.norm(embedding)
                if norm > 0:
                    embedding = embedding / norm
            
            return embedding, result.model_version, 1.0
        else:
            return None, "", 0.0
    
    def extract_and_embed(
        self,
        face_image: np.ndarray
    ) -> Tuple[np.ndarray, str]:
        """
        Extract embedding from face image.
        
        The face image is NOT returned - data minimization policy.
        
        Args:
            face_image: Cropped face image
            
        Returns:
            Tuple of (embedding_vector, embedding_hash)
        """
        embedding, _, _ = self._compute_embedding(face_image)
        
        if embedding is None:
            raise ValueError("Failed to compute embedding")
        
        embedding_hash = compute_embedding_hash(embedding, self.config.hash_algorithm)
        
        # Face image is discarded after this function returns
        return embedding, embedding_hash
    
    def extract_embedding_only(
        self,
        face_image: np.ndarray,
        segmentation_confidence: float = 0.0
    ) -> EmbeddingResult:
        """
        Extract embedding only, discarding face image.
        
        For consensus verification - only the embedding is needed.
        
        Args:
            face_image: Cropped face image
            segmentation_confidence: Confidence from segmentation
            
        Returns:
            EmbeddingResult with embedding and metadata
        """
        try:
            embedding, model_version, extraction_conf = self._compute_embedding(face_image)
            
            if embedding is None:
                return EmbeddingResult(
                    embedding=None,
                    embedding_hash="",
                    confidence=0.0,
                    model_version="",
                    success=False,
                    error_message="Failed to compute embedding",
                    segmentation_confidence=segmentation_confidence,
                    extraction_confidence=0.0,
                )
            
            # Compute hash
            embedding_hash = ""
            if self.config.compute_hash:
                embedding_hash = compute_embedding_hash(
                    embedding, 
                    self.config.hash_algorithm
                )
            
            # Combined confidence
            combined_confidence = (segmentation_confidence + extraction_conf) / 2
            
            return EmbeddingResult(
                embedding=embedding,
                embedding_hash=embedding_hash,
                confidence=combined_confidence,
                model_version=model_version,
                success=True,
                segmentation_confidence=segmentation_confidence,
                extraction_confidence=extraction_conf,
            )
            
        except Exception as e:
            logger.error(f"Embedding extraction failed: {e}")
            return EmbeddingResult(
                embedding=None,
                embedding_hash="",
                confidence=0.0,
                model_version="",
                success=False,
                error_message=str(e),
                segmentation_confidence=segmentation_confidence,
                extraction_confidence=0.0,
            )
    
    def extract_from_document(
        self,
        document_image: np.ndarray,
        segmentor: UNetFaceSegmentor,
        mask_processor: MaskProcessor,
        cropper: FaceCropper
    ) -> EmbeddingResult:
        """
        Full pipeline: segment, crop, and extract embedding from document.
        
        Face image is discarded after embedding extraction.
        
        Args:
            document_image: Full document image
            segmentor: U-Net segmentor instance
            mask_processor: Mask processor instance
            cropper: Face cropper instance
            
        Returns:
            EmbeddingResult
        """
        # Step 1: Segment face region
        seg_result = segmentor.segment_with_details(document_image)
        
        if not seg_result.success:
            return EmbeddingResult(
                embedding=None,
                embedding_hash="",
                confidence=0.0,
                model_version=seg_result.model_version,
                success=False,
                error_message=f"Segmentation failed: {seg_result.error_message}",
                segmentation_confidence=0.0,
                extraction_confidence=0.0,
            )
        
        # Step 2: Process mask
        mask_result = mask_processor.process(seg_result.mask)
        
        if not mask_result.success:
            return EmbeddingResult(
                embedding=None,
                embedding_hash="",
                confidence=0.0,
                model_version=seg_result.model_version,
                success=False,
                error_message=f"Mask processing failed: {mask_result.error_message}",
                segmentation_confidence=seg_result.mean_confidence,
                extraction_confidence=0.0,
            )
        
        # Step 3: Extract and crop face
        extraction_result = cropper.extract(
            image=document_image,
            mask=mask_result.processed_mask,
            confidence=seg_result.mean_confidence,
            model_version=seg_result.model_version,
            model_hash=seg_result.model_hash,
            return_face_image=True  # Need face for embedding
        )
        
        if not extraction_result.success or extraction_result.face_image is None:
            return EmbeddingResult(
                embedding=None,
                embedding_hash="",
                confidence=0.0,
                model_version=seg_result.model_version,
                success=False,
                error_message=f"Face extraction failed: {extraction_result.error_message}",
                segmentation_confidence=seg_result.mean_confidence,
                extraction_confidence=0.0,
            )
        
        # Step 4: Compute embedding (face image is used here)
        face_image = extraction_result.face_image
        
        result = self.extract_embedding_only(
            face_image,
            segmentation_confidence=seg_result.mean_confidence
        )
        
        # Step 5: DATA MINIMIZATION - face image is discarded
        # (it goes out of scope here and is garbage collected)
        # We explicitly set it to None to emphasize this
        del face_image
        
        return result
