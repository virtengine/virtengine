"""
Face feature extraction for trust score training.

Extracts face embeddings from document and selfie images using
the facial verification pipeline.
"""

import logging
from dataclasses import dataclass, field
from typing import List, Optional, Tuple

import numpy as np

from ml.training.config import FeatureConfig

logger = logging.getLogger(__name__)


@dataclass
class FaceFeatures:
    """Face-related features for a sample."""
    
    # Face embeddings
    document_face_embedding: Optional[np.ndarray] = None  # From ID document
    selfie_face_embedding: Optional[np.ndarray] = None    # From selfie
    
    # Similarity between document and selfie faces
    face_similarity: float = 0.0
    
    # Confidence scores
    document_face_confidence: float = 0.0
    selfie_face_confidence: float = 0.0
    
    # Detection flags
    document_face_detected: bool = False
    selfie_face_detected: bool = False
    
    # Quality indicators
    document_face_quality: float = 0.0
    selfie_face_quality: float = 0.0
    
    # Combined embedding (for model input)
    combined_embedding: Optional[np.ndarray] = None
    
    def to_vector(self, embedding_dim: int = 512) -> np.ndarray:
        """Convert to feature vector."""
        features = []
        
        # Document face embedding
        if self.document_face_embedding is not None:
            features.extend(self.document_face_embedding.flatten()[:embedding_dim])
        else:
            features.extend([0.0] * embedding_dim)
        
        # Selfie face embedding
        if self.selfie_face_embedding is not None:
            features.extend(self.selfie_face_embedding.flatten()[:embedding_dim])
        else:
            features.extend([0.0] * embedding_dim)
        
        # Scalar features
        features.extend([
            self.face_similarity,
            self.document_face_confidence,
            self.selfie_face_confidence,
            float(self.document_face_detected),
            float(self.selfie_face_detected),
            self.document_face_quality,
            self.selfie_face_quality,
        ])
        
        return np.array(features, dtype=np.float32)


class FaceFeatureExtractor:
    """
    Extracts face features from samples.
    
    Uses the facial verification pipeline to:
    - Detect faces in document and selfie images
    - Extract face embeddings
    - Compute similarity scores
    - Assess face quality
    """
    
    def __init__(self, config: Optional[FeatureConfig] = None):
        """
        Initialize the face feature extractor.
        
        Args:
            config: Feature configuration
        """
        self.config = config or FeatureConfig()
        self._embedding_dim = self.config.face_embedding_dim
        
        # Initialize face verification pipeline
        self._verifier = None
        self._embedder = None
        try:
            from ml.facial_verification import FaceVerifier, FaceEmbedder
            self._verifier = FaceVerifier()
            self._embedder = FaceEmbedder()
        except ImportError:
            logger.warning(
                "Facial verification pipeline not available, "
                "face features will be empty"
            )
        
        # Initialize face extraction pipeline for documents
        self._face_extractor = None
        try:
            from ml.face_extraction import FaceExtractionPipeline
            self._face_extractor = FaceExtractionPipeline()
        except ImportError:
            logger.warning(
                "Face extraction pipeline not available, "
                "document face features will be empty"
            )
    
    def extract(
        self,
        document_image: Optional[np.ndarray],
        selfie_image: Optional[np.ndarray],
    ) -> FaceFeatures:
        """
        Extract face features from document and selfie images.
        
        Args:
            document_image: Preprocessed document image
            selfie_image: Preprocessed selfie image
            
        Returns:
            FaceFeatures containing embeddings and scores
        """
        features = FaceFeatures()
        
        # Extract document face embedding
        if document_image is not None and self._face_extractor is not None:
            try:
                result = self._face_extractor.extract(document_image)
                if result.success:
                    features.document_face_embedding = result.embedding
                    features.document_face_confidence = result.confidence
                    features.document_face_detected = True
                    features.document_face_quality = result.quality_score
            except Exception as e:
                logger.debug(f"Document face extraction failed: {e}")
        
        # Extract selfie face embedding
        if selfie_image is not None and self._embedder is not None:
            try:
                result = self._embedder.extract(selfie_image)
                if result.success:
                    features.selfie_face_embedding = result.embedding
                    features.selfie_face_confidence = result.confidence
                    features.selfie_face_detected = True
                    features.selfie_face_quality = result.quality_score
            except Exception as e:
                logger.debug(f"Selfie face extraction failed: {e}")
        
        # Compute similarity if both embeddings available
        if (features.document_face_embedding is not None and 
            features.selfie_face_embedding is not None):
            features.face_similarity = self._compute_similarity(
                features.document_face_embedding,
                features.selfie_face_embedding
            )
        
        # Create combined embedding
        features.combined_embedding = self._create_combined_embedding(features)
        
        return features
    
    def _compute_similarity(
        self,
        embedding1: np.ndarray,
        embedding2: np.ndarray
    ) -> float:
        """Compute cosine similarity between two embeddings."""
        # Normalize embeddings
        norm1 = np.linalg.norm(embedding1)
        norm2 = np.linalg.norm(embedding2)
        
        if norm1 == 0 or norm2 == 0:
            return 0.0
        
        embedding1_normalized = embedding1 / norm1
        embedding2_normalized = embedding2 / norm2
        
        # Compute cosine similarity
        similarity = np.dot(embedding1_normalized, embedding2_normalized)
        
        return float(np.clip(similarity, 0.0, 1.0))
    
    def _create_combined_embedding(
        self,
        features: FaceFeatures
    ) -> np.ndarray:
        """Create a combined embedding from document and selfie embeddings."""
        dim = self._embedding_dim
        
        # Initialize with zeros
        combined = np.zeros(dim, dtype=np.float32)
        
        if features.document_face_embedding is not None:
            doc_emb = features.document_face_embedding.flatten()[:dim]
            combined[:len(doc_emb)] = doc_emb
        
        if features.selfie_face_embedding is not None:
            selfie_emb = features.selfie_face_embedding.flatten()[:dim]
            # Average with document embedding if both present
            if features.document_face_embedding is not None:
                combined = (combined + selfie_emb) / 2
            else:
                combined[:len(selfie_emb)] = selfie_emb
        
        return combined
    
    def get_feature_dim(self) -> int:
        """Get the dimension of the output feature vector."""
        # Two embeddings + 7 scalar features
        return self._embedding_dim * 2 + 7
