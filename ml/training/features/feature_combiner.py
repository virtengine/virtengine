"""
Feature combiner for trust score training.

Combines features from all extractors into a unified feature vector
suitable for model training.
"""

import logging
import time
from dataclasses import dataclass, field
from typing import List, Optional, Dict, Any, Tuple

import numpy as np

from ml.training.config import FeatureConfig
from ml.training.dataset.ingestion import IdentitySample, Dataset, DatasetSplit
from ml.training.dataset.preprocessing import PreprocessedSample
from ml.training.dataset.augmentation import AugmentedSample
from ml.training.features.face_features import FaceFeatureExtractor, FaceFeatures
from ml.training.features.doc_features import DocumentFeatureExtractor, DocumentFeatures
from ml.training.features.ocr_features import OCRFeatureExtractor, OCRFeatures
from ml.training.features.metadata_features import MetadataFeatureExtractor, MetadataFeatures

logger = logging.getLogger(__name__)


@dataclass
class FeatureVector:
    """Combined feature vector for a single sample."""
    
    # Sample identification
    sample_id: str
    
    # Component features
    face_embedding: np.ndarray = field(default_factory=lambda: np.zeros(512))
    face_confidence: float = 0.0
    face_features: Optional[FaceFeatures] = None
    
    doc_quality_score: float = 0.0
    doc_features: Optional[DocumentFeatures] = None
    
    ocr_field_scores: Dict[str, float] = field(default_factory=dict)
    ocr_features: Optional[OCRFeatures] = None
    
    metadata_features: Optional[MetadataFeatures] = None
    
    # Combined vector for model input
    combined_vector: np.ndarray = field(default_factory=lambda: np.zeros(768))
    
    # Label
    trust_score: float = 0.0
    
    # Metadata
    extraction_time_ms: float = 0.0
    
    def to_model_input(self) -> np.ndarray:
        """Get the combined vector as model input."""
        return self.combined_vector.astype(np.float32)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary (excluding large arrays)."""
        return {
            "sample_id": self.sample_id,
            "face_confidence": self.face_confidence,
            "doc_quality_score": self.doc_quality_score,
            "ocr_field_scores": self.ocr_field_scores,
            "trust_score": self.trust_score,
            "combined_vector_dim": len(self.combined_vector),
            "extraction_time_ms": self.extraction_time_ms,
        }


@dataclass
class FeatureDataset:
    """Dataset of feature vectors ready for training."""
    
    # Feature vectors by split
    train: List[FeatureVector]
    validation: List[FeatureVector]
    test: List[FeatureVector]
    
    # Feature dimensions
    feature_dim: int = 768
    
    # Normalization parameters (computed from training set)
    feature_mean: Optional[np.ndarray] = None
    feature_std: Optional[np.ndarray] = None
    
    # Statistics
    total_extraction_time_ms: float = 0.0
    
    def get_arrays(
        self,
        split: str = "train"
    ) -> Tuple[np.ndarray, np.ndarray]:
        """
        Get feature and label arrays for a split.
        
        Args:
            split: "train", "validation", or "test"
            
        Returns:
            Tuple of (features, labels) arrays
        """
        if split == "train":
            vectors = self.train
        elif split == "validation":
            vectors = self.validation
        elif split == "test":
            vectors = self.test
        else:
            raise ValueError(f"Unknown split: {split}")
        
        if len(vectors) == 0:
            return np.array([]), np.array([])
        
        features = np.stack([v.combined_vector for v in vectors])
        labels = np.array([v.trust_score for v in vectors])
        
        return features, labels
    
    def normalize(self) -> None:
        """Normalize features using training set statistics."""
        if len(self.train) == 0:
            return
        
        # Compute mean and std from training set
        train_features = np.stack([v.combined_vector for v in self.train])
        self.feature_mean = np.mean(train_features, axis=0)
        self.feature_std = np.std(train_features, axis=0) + 1e-8
        
        # Apply normalization to all splits
        for vectors in [self.train, self.validation, self.test]:
            for v in vectors:
                v.combined_vector = (
                    v.combined_vector - self.feature_mean
                ) / self.feature_std


class FeatureExtractor:
    """
    Unified feature extractor combining all feature types.
    
    Extracts and combines:
    - Face embeddings and confidence scores
    - Document quality features
    - OCR field scores and validation
    - Capture metadata features
    """
    
    def __init__(self, config: Optional[FeatureConfig] = None):
        """
        Initialize the feature extractor.
        
        Args:
            config: Feature configuration
        """
        self.config = config or FeatureConfig()
        
        # Initialize component extractors
        self.face_extractor = FaceFeatureExtractor(config)
        self.doc_extractor = DocumentFeatureExtractor(config)
        self.ocr_extractor = OCRFeatureExtractor(config)
        self.meta_extractor = MetadataFeatureExtractor(config)
        
        # Compute total feature dimension
        self._compute_feature_dim()
    
    def _compute_feature_dim(self) -> None:
        """Compute the total feature dimension."""
        # Use configured dimension or compute from components
        self._feature_dim = self.config.combined_feature_dim
    
    def extract_all(
        self,
        sample: IdentitySample,
        document_image: Optional[np.ndarray] = None,
        selfie_image: Optional[np.ndarray] = None,
    ) -> FeatureVector:
        """
        Extract all features from a sample.
        
        Args:
            sample: Identity sample with metadata
            document_image: Preprocessed document image
            selfie_image: Preprocessed selfie image
            
        Returns:
            FeatureVector with combined features
        """
        start_time = time.perf_counter()
        
        # Use images from sample if not provided
        if document_image is None and sample.document_image is not None:
            document_image = sample.document_image.data
        if selfie_image is None and sample.selfie_image is not None:
            selfie_image = sample.selfie_image.data
        
        # Extract component features
        face_features = self.face_extractor.extract(document_image, selfie_image)
        doc_features = self.doc_extractor.extract(document_image)
        
        doc_type = sample.document_info.doc_type if sample.document_info else "id_card"
        ocr_features = self.ocr_extractor.extract(document_image, doc_type)
        
        meta_features = self.meta_extractor.extract(sample.capture_metadata)
        
        # Combine into unified vector
        combined_vector = self._combine_features(
            face_features, doc_features, ocr_features, meta_features
        )
        
        # Build OCR field scores dict
        ocr_field_scores = {
            name: score.confidence
            for name, score in ocr_features.field_scores.items()
        }
        
        extraction_time = (time.perf_counter() - start_time) * 1000
        
        return FeatureVector(
            sample_id=sample.sample_id,
            face_embedding=face_features.combined_embedding if face_features.combined_embedding is not None else np.zeros(512),
            face_confidence=face_features.selfie_face_confidence,
            face_features=face_features,
            doc_quality_score=doc_features.overall_quality_score,
            doc_features=doc_features,
            ocr_field_scores=ocr_field_scores,
            ocr_features=ocr_features,
            metadata_features=meta_features,
            combined_vector=combined_vector,
            trust_score=sample.trust_score,
            extraction_time_ms=extraction_time,
        )
    
    def _combine_features(
        self,
        face_features: FaceFeatures,
        doc_features: DocumentFeatures,
        ocr_features: OCRFeatures,
        meta_features: MetadataFeatures,
    ) -> np.ndarray:
        """Combine all features into a single vector."""
        components = []
        
        # Face embedding (primary feature)
        if face_features.combined_embedding is not None:
            face_emb = face_features.combined_embedding
            # Pad or truncate to target size
            target_face_dim = self.config.face_embedding_dim
            if len(face_emb) < target_face_dim:
                face_emb = np.pad(face_emb, (0, target_face_dim - len(face_emb)))
            else:
                face_emb = face_emb[:target_face_dim]
            components.append(face_emb)
        else:
            components.append(np.zeros(self.config.face_embedding_dim))
        
        # Face scalar features
        face_scalars = np.array([
            face_features.face_similarity,
            face_features.document_face_confidence,
            face_features.selfie_face_confidence,
            face_features.document_face_quality,
            face_features.selfie_face_quality,
            float(face_features.document_face_detected),
            float(face_features.selfie_face_detected),
        ], dtype=np.float32)
        components.append(face_scalars)
        
        # Document features
        components.append(doc_features.to_vector())
        
        # OCR features (simplified)
        ocr_scalars = np.array([
            ocr_features.overall_ocr_confidence,
            ocr_features.fields_extracted_ratio,
            ocr_features.fields_validated_ratio,
            float(ocr_features.ocr_success),
            float(ocr_features.has_name),
            float(ocr_features.has_dob),
            float(ocr_features.has_doc_number),
            float(ocr_features.has_expiry),
            float(ocr_features.has_nationality),
        ], dtype=np.float32)
        components.append(ocr_scalars)
        
        # Metadata features
        components.append(meta_features.to_vector())
        
        # Concatenate all components
        combined = np.concatenate(components)
        
        # Pad or truncate to target dimension
        target_dim = self._feature_dim
        if len(combined) < target_dim:
            combined = np.pad(combined, (0, target_dim - len(combined)))
        else:
            combined = combined[:target_dim]
        
        return combined.astype(np.float32)
    
    def extract_batch(
        self,
        dataset: Dataset,
        normalize: bool = True,
    ) -> FeatureDataset:
        """
        Extract features for an entire dataset.
        
        Args:
            dataset: Dataset with samples
            normalize: Whether to normalize features
            
        Returns:
            FeatureDataset with extracted features
        """
        start_time = time.perf_counter()
        
        train_features = self._extract_split(dataset.train)
        val_features = self._extract_split(dataset.validation)
        test_features = self._extract_split(dataset.test)
        
        total_time = (time.perf_counter() - start_time) * 1000
        
        feature_dataset = FeatureDataset(
            train=train_features,
            validation=val_features,
            test=test_features,
            feature_dim=self._feature_dim,
            total_extraction_time_ms=total_time,
        )
        
        if normalize:
            feature_dataset.normalize()
        
        logger.info(
            f"Extracted features: {len(train_features)} train, "
            f"{len(val_features)} val, {len(test_features)} test "
            f"in {total_time:.2f}ms"
        )
        
        return feature_dataset
    
    def _extract_split(
        self,
        split: DatasetSplit
    ) -> List[FeatureVector]:
        """Extract features for a dataset split."""
        features = []
        
        for sample in split.samples:
            try:
                feature_vector = self.extract_all(sample)
                features.append(feature_vector)
            except Exception as e:
                logger.error(f"Feature extraction failed for {sample.sample_id}: {e}")
        
        return features
    
    def extract_from_preprocessed(
        self,
        samples: List[PreprocessedSample],
    ) -> List[FeatureVector]:
        """Extract features from preprocessed samples."""
        features = []
        
        for sample in samples:
            if not sample.success or sample.original_sample is None:
                continue
            
            try:
                feature_vector = self.extract_all(
                    sample.original_sample,
                    document_image=sample.document_image,
                    selfie_image=sample.selfie_image,
                )
                features.append(feature_vector)
            except Exception as e:
                logger.error(f"Feature extraction failed for {sample.sample_id}: {e}")
        
        return features
    
    def extract_from_augmented(
        self,
        samples: List[AugmentedSample],
    ) -> List[FeatureVector]:
        """Extract features from augmented samples."""
        features = []
        
        for sample in samples:
            if sample.original_sample is None:
                continue
            
            original = sample.original_sample.original_sample
            if original is None:
                continue
            
            try:
                feature_vector = self.extract_all(
                    original,
                    document_image=sample.document_image,
                    selfie_image=sample.selfie_image,
                )
                # Use augmentation ID
                feature_vector.sample_id = sample.augmentation_id
                feature_vector.trust_score = sample.trust_score
                features.append(feature_vector)
            except Exception as e:
                logger.error(
                    f"Feature extraction failed for {sample.augmentation_id}: {e}"
                )
        
        return features
    
    def get_feature_dim(self) -> int:
        """Get the total feature dimension."""
        return self._feature_dim
