"""
Face embedding extraction module.

This module provides face embedding extraction using DeepFace with
support for multiple models (VGG-Face, Facenet, ArcFace, etc.).
"""

import logging
import hashlib
import os
from dataclasses import dataclass
from typing import Optional, List, Tuple, Any
from pathlib import Path

import numpy as np

from ml.facial_verification.config import VerificationConfig, ModelName, DistanceMetric
from ml.facial_verification.reason_codes import ReasonCodes
from ml.facial_verification.determinism import DeterminismController

logger = logging.getLogger(__name__)


@dataclass
class EmbeddingResult:
    """Result of embedding extraction."""
    
    embedding: Optional[np.ndarray]
    success: bool
    error_code: Optional[ReasonCodes] = None
    model_name: str = ""
    model_version: str = ""
    embedding_hash: Optional[str] = None
    dimension: int = 0
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "success": self.success,
            "error_code": self.error_code.value if self.error_code else None,
            "model_name": self.model_name,
            "model_version": self.model_version,
            "embedding_hash": self.embedding_hash,
            "dimension": self.dimension,
        }


class FaceEmbedder:
    """
    Face embedding extraction using DeepFace.
    
    This class provides deterministic face embedding extraction with
    support for multiple models and consistent hashing for consensus.
    """
    
    # Model embedding dimensions
    EMBEDDING_DIMS = {
        ModelName.VGG_FACE: 2622,
        ModelName.FACENET: 128,
        ModelName.FACENET512: 512,
        ModelName.ARCFACE: 512,
        ModelName.DLIB: 128,
        ModelName.SFACE: 128,
    }
    
    # DeepFace model name mapping
    DEEPFACE_MODEL_NAMES = {
        ModelName.VGG_FACE: "VGG-Face",
        ModelName.FACENET: "Facenet",
        ModelName.FACENET512: "Facenet512",
        ModelName.ARCFACE: "ArcFace",
        ModelName.DLIB: "Dlib",
        ModelName.SFACE: "SFace",
    }
    
    def __init__(self, config: Optional[VerificationConfig] = None):
        """
        Initialize the face embedder.
        
        Args:
            config: Verification configuration. Uses defaults if not provided.
        """
        self.config = config or VerificationConfig()
        self._model = None
        self._model_hash: Optional[str] = None
        self._initialized = False
        self._determinism = DeterminismController(self.config)
    
    def _lazy_init(self) -> None:
        """Lazy initialization of the model."""
        if self._initialized:
            return
        
        # Ensure deterministic execution
        self._determinism.ensure_deterministic()
        
        try:
            # Import DeepFace
            from deepface import DeepFace
            
            # Pre-load the model by running a dummy prediction
            # This ensures the model is cached
            model_name = self.DEEPFACE_MODEL_NAMES.get(
                self.config.model_name, 
                "VGG-Face"
            )
            
            logger.info(f"Initializing face embedding model: {model_name}")
            
            # Build model to get weights
            from deepface.basemodels import VGGFace, Facenet, Facenet512, ArcFace, Dlib, SFace
            
            model_builders = {
                "VGG-Face": VGGFace,
                "Facenet": Facenet,
                "Facenet512": Facenet512,
                "ArcFace": ArcFace,
                "Dlib": Dlib,
                "SFace": SFace,
            }
            
            builder = model_builders.get(model_name)
            if builder:
                self._model = builder.loadModel()
                self._model_hash = self._compute_weights_hash()
            
            self._initialized = True
            logger.info(f"Model initialized. Hash: {self._model_hash[:16]}...")
            
        except ImportError as e:
            logger.error(f"Failed to import DeepFace: {e}")
            raise RuntimeError(f"DeepFace not available: {e}")
        except Exception as e:
            logger.error(f"Failed to initialize model: {e}")
            raise RuntimeError(f"Model initialization failed: {e}")
    
    def _compute_weights_hash(self) -> str:
        """
        Compute hash of model weights for determinism verification.
        
        Returns:
            SHA256 hash of model weights
        """
        if self._model is None:
            return ""
        
        try:
            # Get model weights
            weights = self._model.get_weights()
            
            # Concatenate all weight arrays and hash
            weight_bytes = b''
            for w in weights:
                # Round to reduce floating-point precision issues
                w_rounded = np.round(w, decimals=6)
                weight_bytes += w_rounded.tobytes()
            
            return hashlib.sha256(weight_bytes).hexdigest()
            
        except Exception as e:
            logger.warning(f"Could not compute weights hash: {e}")
            return ""
    
    @property
    def model_hash(self) -> str:
        """Get the model weights hash."""
        if not self._initialized:
            self._lazy_init()
        return self._model_hash or ""
    
    @property
    def embedding_dimension(self) -> int:
        """Get the embedding dimension for the current model."""
        return self.EMBEDDING_DIMS.get(self.config.model_name, 512)
    
    def embed(self, face_image: np.ndarray) -> EmbeddingResult:
        """
        Extract face embedding from a preprocessed face image.
        
        Args:
            face_image: Preprocessed face image (aligned, cropped)
            
        Returns:
            EmbeddingResult with embedding vector and metadata
        """
        if face_image is None or face_image.size == 0:
            return EmbeddingResult(
                embedding=None,
                success=False,
                error_code=ReasonCodes.INVALID_IMAGE_FORMAT
            )
        
        try:
            self._lazy_init()
            
            # Get model name for DeepFace
            model_name = self.DEEPFACE_MODEL_NAMES.get(
                self.config.model_name, 
                "VGG-Face"
            )
            
            # Import DeepFace
            from deepface import DeepFace
            
            # Extract embedding
            # DeepFace.represent returns list of dicts with 'embedding' key
            result = DeepFace.represent(
                img_path=face_image,
                model_name=model_name,
                enforce_detection=False,  # We already detected the face
                detector_backend='skip',  # Skip detection since face is already cropped
            )
            
            if not result or len(result) == 0:
                return EmbeddingResult(
                    embedding=None,
                    success=False,
                    error_code=ReasonCodes.EMBEDDING_ERROR,
                    model_name=model_name,
                    model_version=self.config.model_version
                )
            
            embedding = np.array(result[0]['embedding'])
            
            # Normalize embedding
            embedding = self._normalize_embedding(embedding)
            
            # Compute embedding hash
            embedding_hash = self._compute_embedding_hash(embedding)
            
            return EmbeddingResult(
                embedding=embedding,
                success=True,
                model_name=model_name,
                model_version=self.config.model_version,
                embedding_hash=embedding_hash,
                dimension=len(embedding)
            )
            
        except Exception as e:
            logger.error(f"Embedding extraction error: {e}")
            return EmbeddingResult(
                embedding=None,
                success=False,
                error_code=ReasonCodes.EMBEDDING_ERROR,
                model_name=self.config.model_name.value,
                model_version=self.config.model_version
            )
    
    def embed_batch(
        self, 
        face_images: List[np.ndarray]
    ) -> List[EmbeddingResult]:
        """
        Extract embeddings from multiple face images.
        
        Args:
            face_images: List of preprocessed face images
            
        Returns:
            List of EmbeddingResult for each image
        """
        results = []
        for image in face_images:
            results.append(self.embed(image))
        return results
    
    def _normalize_embedding(self, embedding: np.ndarray) -> np.ndarray:
        """
        Normalize embedding vector to unit length.
        
        Args:
            embedding: Raw embedding vector
            
        Returns:
            L2-normalized embedding vector
        """
        norm = np.linalg.norm(embedding)
        if norm > 0:
            embedding = embedding / norm
        return embedding
    
    def _compute_embedding_hash(self, embedding: np.ndarray) -> str:
        """
        Compute deterministic hash of embedding.
        
        Args:
            embedding: Embedding vector
            
        Returns:
            SHA256 hash of embedding
        """
        # Round to reduce floating-point precision issues
        embedding_rounded = np.round(embedding, decimals=6)
        return hashlib.sha256(embedding_rounded.tobytes()).hexdigest()
    
    def compute_distance(
        self, 
        embedding1: np.ndarray, 
        embedding2: np.ndarray,
        metric: Optional[DistanceMetric] = None
    ) -> float:
        """
        Compute distance between two embeddings.
        
        Args:
            embedding1: First embedding vector
            embedding2: Second embedding vector
            metric: Distance metric to use (defaults to config)
            
        Returns:
            Distance value (lower = more similar)
        """
        metric = metric or self.config.distance_metric
        
        if metric == DistanceMetric.COSINE:
            return self._cosine_distance(embedding1, embedding2)
        elif metric == DistanceMetric.EUCLIDEAN:
            return self._euclidean_distance(embedding1, embedding2)
        elif metric == DistanceMetric.EUCLIDEAN_L2:
            return self._euclidean_l2_distance(embedding1, embedding2)
        else:
            return self._cosine_distance(embedding1, embedding2)
    
    def compute_similarity(
        self, 
        embedding1: np.ndarray, 
        embedding2: np.ndarray,
        metric: Optional[DistanceMetric] = None
    ) -> float:
        """
        Compute similarity between two embeddings.
        
        Args:
            embedding1: First embedding vector
            embedding2: Second embedding vector
            metric: Distance metric to use (defaults to config)
            
        Returns:
            Similarity value (0-1, higher = more similar)
        """
        metric = metric or self.config.distance_metric
        
        if metric == DistanceMetric.COSINE:
            # Cosine similarity
            return self._cosine_similarity(embedding1, embedding2)
        else:
            # Convert distance to similarity
            distance = self.compute_distance(embedding1, embedding2, metric)
            # Use exponential decay for Euclidean distance
            return np.exp(-distance)
    
    def _cosine_distance(self, a: np.ndarray, b: np.ndarray) -> float:
        """Compute cosine distance."""
        return 1.0 - self._cosine_similarity(a, b)
    
    def _cosine_similarity(self, a: np.ndarray, b: np.ndarray) -> float:
        """Compute cosine similarity."""
        dot = np.dot(a, b)
        norm_a = np.linalg.norm(a)
        norm_b = np.linalg.norm(b)
        
        if norm_a == 0 or norm_b == 0:
            return 0.0
        
        return float(dot / (norm_a * norm_b))
    
    def _euclidean_distance(self, a: np.ndarray, b: np.ndarray) -> float:
        """Compute Euclidean distance."""
        return float(np.linalg.norm(a - b))
    
    def _euclidean_l2_distance(self, a: np.ndarray, b: np.ndarray) -> float:
        """Compute L2-normalized Euclidean distance."""
        a_norm = a / np.linalg.norm(a) if np.linalg.norm(a) > 0 else a
        b_norm = b / np.linalg.norm(b) if np.linalg.norm(b) > 0 else b
        return float(np.linalg.norm(a_norm - b_norm))
    
    def verify_model_hash(self) -> bool:
        """
        Verify that the model hash matches the expected hash.
        
        Returns:
            True if hash matches or no expected hash configured
        """
        if not self.config.determinism.verify_model_hash:
            return True
        
        expected = self.config.determinism.expected_model_hash
        if not expected:
            return True
        
        return self.model_hash == expected
