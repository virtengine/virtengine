"""
Data preprocessing for training samples.

This module handles preprocessing of images and metadata before
feature extraction, including document preprocessing and image
normalization.
"""

import logging
import time
from dataclasses import dataclass, field
from typing import List, Optional, Tuple, Dict, Any

import numpy as np

from ml.training.config import PreprocessingConfig
from ml.training.dataset.ingestion import (
    IdentitySample,
    ImageData,
    Dataset,
    DatasetSplit,
    SplitType,
)

logger = logging.getLogger(__name__)


@dataclass
class PreprocessedSample:
    """A preprocessed sample ready for feature extraction."""
    
    # Original sample reference
    sample_id: str
    
    # Preprocessed images
    document_image: Optional[np.ndarray] = None  # Normalized
    selfie_image: Optional[np.ndarray] = None    # Normalized
    
    # Original sample data
    original_sample: Optional[IdentitySample] = None
    
    # Preprocessing metadata
    preprocessing_applied: List[str] = field(default_factory=list)
    processing_time_ms: float = 0.0
    
    # Quality indicators from preprocessing
    document_orientation_corrected: bool = False
    document_perspectiVIRTENGINE_corrected: bool = False
    
    # Status
    success: bool = True
    error_message: Optional[str] = None


@dataclass
class PreprocessedDataset:
    """Dataset with preprocessed samples."""
    
    train: List[PreprocessedSample]
    validation: List[PreprocessedSample]
    test: List[PreprocessedSample]
    
    config: Optional[PreprocessingConfig] = None
    total_processing_time_ms: float = 0.0


class DatasetPreprocessor:
    """
    Preprocesses dataset samples for feature extraction.
    
    Applies:
    - Image normalization (mean/std normalization)
    - Image resizing to target dimensions
    - Document preprocessing (orientation, perspective)
    - Color space conversion
    """
    
    def __init__(self, config: Optional[PreprocessingConfig] = None):
        """
        Initialize the preprocessor.
        
        Args:
            config: Preprocessing configuration
        """
        self.config = config or PreprocessingConfig()
        
        # Initialize document preprocessing pipeline if available
        self._doc_pipeline = None
        try:
            from ml.document_preprocessing import DocumentPreprocessingPipeline
            self._doc_pipeline = DocumentPreprocessingPipeline()
        except ImportError:
            logger.warning("Document preprocessing pipeline not available")
    
    def preprocess_dataset(self, dataset: Dataset) -> PreprocessedDataset:
        """
        Preprocess all samples in a dataset.
        
        Args:
            dataset: Dataset to preprocess
            
        Returns:
            PreprocessedDataset with preprocessed samples
        """
        start_time = time.perf_counter()
        
        train_preprocessed = self._preprocess_split(dataset.train)
        val_preprocessed = self._preprocess_split(dataset.validation)
        test_preprocessed = self._preprocess_split(dataset.test)
        
        total_time = (time.perf_counter() - start_time) * 1000
        
        logger.info(
            f"Preprocessed dataset: {len(train_preprocessed)} train, "
            f"{len(val_preprocessed)} val, {len(test_preprocessed)} test "
            f"in {total_time:.2f}ms"
        )
        
        return PreprocessedDataset(
            train=train_preprocessed,
            validation=val_preprocessed,
            test=test_preprocessed,
            config=self.config,
            total_processing_time_ms=total_time,
        )
    
    def _preprocess_split(
        self,
        split: DatasetSplit
    ) -> List[PreprocessedSample]:
        """Preprocess a dataset split."""
        preprocessed = []
        
        for sample in split.samples:
            result = self.preprocess_sample(sample)
            if result.success:
                preprocessed.append(result)
            else:
                logger.warning(
                    f"Failed to preprocess sample {sample.sample_id}: "
                    f"{result.error_message}"
                )
        
        return preprocessed
    
    def preprocess_sample(
        self,
        sample: IdentitySample
    ) -> PreprocessedSample:
        """
        Preprocess a single sample.
        
        Args:
            sample: Sample to preprocess
            
        Returns:
            PreprocessedSample with normalized images
        """
        start_time = time.perf_counter()
        preprocessing_applied = []
        
        try:
            # Preprocess document image
            document_image = None
            doc_orientation_corrected = False
            doc_perspectiVIRTENGINE_corrected = False
            
            if sample.document_image is not None:
                doc_result = self._preprocess_document_image(
                    sample.document_image.data
                )
                document_image = doc_result['image']
                doc_orientation_corrected = doc_result.get('orientation_corrected', False)
                doc_perspectiVIRTENGINE_corrected = doc_result.get('perspectiVIRTENGINE_corrected', False)
                preprocessing_applied.extend(doc_result.get('applied', []))
            
            # Preprocess selfie image
            selfie_image = None
            if sample.selfie_image is not None:
                selfie_image = self._preprocess_selfie_image(
                    sample.selfie_image.data
                )
                preprocessing_applied.append("selfie_normalized")
            
            processing_time = (time.perf_counter() - start_time) * 1000
            
            return PreprocessedSample(
                sample_id=sample.sample_id,
                document_image=document_image,
                selfie_image=selfie_image,
                original_sample=sample,
                preprocessing_applied=preprocessing_applied,
                processing_time_ms=processing_time,
                document_orientation_corrected=doc_orientation_corrected,
                document_perspectiVIRTENGINE_corrected=doc_perspectiVIRTENGINE_corrected,
                success=True,
            )
            
        except Exception as e:
            processing_time = (time.perf_counter() - start_time) * 1000
            logger.error(f"Preprocessing error for sample {sample.sample_id}: {e}")
            
            return PreprocessedSample(
                sample_id=sample.sample_id,
                original_sample=sample,
                processing_time_ms=processing_time,
                success=False,
                error_message=str(e),
            )
    
    def _preprocess_document_image(
        self,
        image: np.ndarray
    ) -> Dict[str, Any]:
        """Preprocess a document image."""
        applied = []
        result_image = image.copy()
        orientation_corrected = False
        perspectiVIRTENGINE_corrected = False
        
        # Apply document preprocessing pipeline if available
        if self._doc_pipeline and self.config.apply_orientation_correction:
            try:
                doc_result = self._doc_pipeline.process(result_image)
                if doc_result.success:
                    result_image = doc_result.normalized_image
                    orientation_corrected = doc_result.rotation_applied != 0
                    perspectiVIRTENGINE_corrected = doc_result.perspectiVIRTENGINE_corrected
                    applied.extend(doc_result.enhancements_applied)
            except Exception as e:
                logger.warning(f"Document preprocessing failed: {e}")
        
        # Resize to target size
        result_image = self._resize_image(
            result_image,
            self.config.image_size
        )
        applied.append("resized")
        
        # Normalize
        if self.config.normalize_images:
            result_image = self._normalize_image(result_image)
            applied.append("normalized")
        
        return {
            'image': result_image,
            'orientation_corrected': orientation_corrected,
            'perspectiVIRTENGINE_corrected': perspectiVIRTENGINE_corrected,
            'applied': applied,
        }
    
    def _preprocess_selfie_image(
        self,
        image: np.ndarray
    ) -> np.ndarray:
        """Preprocess a selfie image."""
        result_image = image.copy()
        
        # Resize to target size
        result_image = self._resize_image(
            result_image,
            self.config.image_size
        )
        
        # Normalize
        if self.config.normalize_images:
            result_image = self._normalize_image(result_image)
        
        return result_image
    
    def _resize_image(
        self,
        image: np.ndarray,
        target_size: Tuple[int, int]
    ) -> np.ndarray:
        """Resize an image to target size."""
        try:
            from PIL import Image
            
            pil_image = Image.fromarray(image)
            pil_image = pil_image.resize(target_size, Image.LANCZOS)
            return np.array(pil_image)
        except ImportError:
            # Fallback to basic resize
            import cv2
            return cv2.resize(image, target_size, interpolation=cv2.INTER_LANCZOS4)
    
    def _normalize_image(self, image: np.ndarray) -> np.ndarray:
        """Normalize image with mean and std."""
        # Convert to float
        image_float = image.astype(np.float32) / 255.0
        
        # Apply mean/std normalization
        mean = np.array(self.config.normalize_mean)
        std = np.array(self.config.normalize_std)
        
        normalized = (image_float - mean) / std
        
        return normalized.astype(np.float32)
