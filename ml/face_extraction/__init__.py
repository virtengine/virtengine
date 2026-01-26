"""
VirtEngine Face Extraction Pipeline

This module provides face extraction from identity documents using U-Net
segmentation for the VEID identity verification system. It supports:
- U-Net-based face region segmentation
- Mask post-processing (thresholding, morphological operations)
- Face bounding box extraction and cropping
- Direct embedding extraction (data minimization policy)
- Deterministic execution for blockchain consensus

Usage:
    from ml.face_extraction import FaceExtractionPipeline, FaceExtractionConfig
    
    config = FaceExtractionConfig()
    pipeline = FaceExtractionPipeline(config)
    
    # Extract embedding only (data minimization)
    result = pipeline.extract(document_image)
    
    # Extract face image (when explicitly needed)
    result = pipeline.extract(document_image, return_face_image=True)
"""

from ml.face_extraction.config import (
    FaceExtractionConfig,
    UNetConfig,
    MaskProcessingConfig,
    CropperConfig,
)
from ml.face_extraction.unet_model import UNetFaceSegmentor
from ml.face_extraction.mask_processing import MaskProcessor
from ml.face_extraction.face_cropper import (
    FaceCropper,
    BoundingBox,
    FaceExtractionResult,
)
from ml.face_extraction.embedding_extractor import (
    DocumentFaceEmbeddingExtractor,
    EmbeddingResult,
)
from ml.face_extraction.pipeline import FaceExtractionPipeline

__version__ = "1.0.0"
__all__ = [
    "FaceExtractionConfig",
    "UNetConfig",
    "MaskProcessingConfig",
    "CropperConfig",
    "UNetFaceSegmentor",
    "MaskProcessor",
    "FaceCropper",
    "BoundingBox",
    "FaceExtractionResult",
    "DocumentFaceEmbeddingExtractor",
    "EmbeddingResult",
    "FaceExtractionPipeline",
]
