"""
Model registry for face recognition models.

This module maintains a registry of approved model versions with their
expected hashes for deterministic verification.
"""

import logging
from dataclasses import dataclass
from typing import Dict, Optional, List
from enum import Enum

logger = logging.getLogger(__name__)


@dataclass
class ModelInfo:
    """Information about a registered model."""
    
    # Model identification
    name: str
    version: str
    
    # Model properties
    embedding_dimension: int
    input_size: tuple
    
    # Hashes for verification
    weights_hash: str
    config_hash: str
    
    # Metadata
    description: str
    source_url: str
    license: str
    
    # Performance metrics (optional)
    lfw_accuracy: Optional[float] = None
    inference_time_ms: Optional[float] = None
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "name": self.name,
            "version": self.version,
            "embedding_dimension": self.embedding_dimension,
            "input_size": self.input_size,
            "weights_hash": self.weights_hash,
            "config_hash": self.config_hash,
            "description": self.description,
            "source_url": self.source_url,
            "license": self.license,
            "lfw_accuracy": self.lfw_accuracy,
            "inference_time_ms": self.inference_time_ms,
        }


class ModelRegistry:
    """
    Registry of approved face recognition models.
    
    This registry maintains a list of approved models with their
    expected hashes for deterministic verification during consensus.
    """
    
    # Registered models with their expected hashes
    # Note: Actual hashes should be computed and verified during model validation
    _MODELS: Dict[str, Dict[str, ModelInfo]] = {
        "VGG-Face": {
            "1.0.0": ModelInfo(
                name="VGG-Face",
                version="1.0.0",
                embedding_dimension=2622,
                input_size=(224, 224),
                weights_hash="",  # To be computed on first run
                config_hash="",
                description="VGG-Face model trained on VGGFace dataset",
                source_url="https://github.com/serengil/deepface",
                license="MIT",
                lfw_accuracy=0.9762,
                inference_time_ms=250.0,
            ),
        },
        "Facenet": {
            "1.0.0": ModelInfo(
                name="Facenet",
                version="1.0.0",
                embedding_dimension=128,
                input_size=(160, 160),
                weights_hash="",
                config_hash="",
                description="Facenet model trained on MS-Celeb-1M",
                source_url="https://github.com/serengil/deepface",
                license="MIT",
                lfw_accuracy=0.9963,
                inference_time_ms=180.0,
            ),
        },
        "Facenet512": {
            "1.0.0": ModelInfo(
                name="Facenet512",
                version="1.0.0",
                embedding_dimension=512,
                input_size=(160, 160),
                weights_hash="",
                config_hash="",
                description="Facenet512 model with 512-dim embeddings",
                source_url="https://github.com/serengil/deepface",
                license="MIT",
                lfw_accuracy=0.9965,
                inference_time_ms=200.0,
            ),
        },
        "ArcFace": {
            "1.0.0": ModelInfo(
                name="ArcFace",
                version="1.0.0",
                embedding_dimension=512,
                input_size=(112, 112),
                weights_hash="",
                config_hash="",
                description="ArcFace model with additive angular margin loss",
                source_url="https://github.com/serengil/deepface",
                license="MIT",
                lfw_accuracy=0.9982,
                inference_time_ms=150.0,
            ),
        },
        "Dlib": {
            "1.0.0": ModelInfo(
                name="Dlib",
                version="1.0.0",
                embedding_dimension=128,
                input_size=(150, 150),
                weights_hash="",
                config_hash="",
                description="Dlib face recognition model",
                source_url="http://dlib.net/",
                license="BSL-1.0",
                lfw_accuracy=0.9938,
                inference_time_ms=80.0,
            ),
        },
        "SFace": {
            "1.0.0": ModelInfo(
                name="SFace",
                version="1.0.0",
                embedding_dimension=128,
                input_size=(112, 112),
                weights_hash="",
                config_hash="",
                description="SFace optimized for mobile deployment",
                source_url="https://github.com/serengil/deepface",
                license="MIT",
                lfw_accuracy=0.9940,
                inference_time_ms=60.0,
            ),
        },
    }
    
    # Current approved versions for production use
    _APPROVED_VERSIONS: Dict[str, str] = {
        "VGG-Face": "1.0.0",
        "Facenet": "1.0.0",
        "Facenet512": "1.0.0",
        "ArcFace": "1.0.0",
        "Dlib": "1.0.0",
        "SFace": "1.0.0",
    }
    
    @classmethod
    def get_model_info(
        cls, 
        model_name: str, 
        version: Optional[str] = None
    ) -> Optional[ModelInfo]:
        """
        Get information about a registered model.
        
        Args:
            model_name: Name of the model
            version: Specific version (defaults to approved version)
            
        Returns:
            ModelInfo or None if not found
        """
        if model_name not in cls._MODELS:
            logger.warning(f"Model not registered: {model_name}")
            return None
        
        version = version or cls._APPROVED_VERSIONS.get(model_name)
        if version not in cls._MODELS[model_name]:
            logger.warning(f"Version not registered: {model_name} v{version}")
            return None
        
        return cls._MODELS[model_name][version]
    
    @classmethod
    def is_approved(cls, model_name: str, version: str) -> bool:
        """
        Check if a model version is approved for production use.
        
        Args:
            model_name: Name of the model
            version: Version to check
            
        Returns:
            True if approved
        """
        approved_version = cls._APPROVED_VERSIONS.get(model_name)
        return approved_version == version
    
    @classmethod
    def get_approved_version(cls, model_name: str) -> Optional[str]:
        """
        Get the approved version for a model.
        
        Args:
            model_name: Name of the model
            
        Returns:
            Approved version string or None
        """
        return cls._APPROVED_VERSIONS.get(model_name)
    
    @classmethod
    def list_models(cls) -> List[str]:
        """Get list of all registered model names."""
        return list(cls._MODELS.keys())
    
    @classmethod
    def list_versions(cls, model_name: str) -> List[str]:
        """Get list of all versions for a model."""
        if model_name not in cls._MODELS:
            return []
        return list(cls._MODELS[model_name].keys())
    
    @classmethod
    def register_model(cls, info: ModelInfo) -> None:
        """
        Register a new model or version.
        
        Args:
            info: ModelInfo for the model to register
        """
        if info.name not in cls._MODELS:
            cls._MODELS[info.name] = {}
        
        cls._MODELS[info.name][info.version] = info
        logger.info(f"Registered model: {info.name} v{info.version}")
    
    @classmethod
    def update_weights_hash(
        cls, 
        model_name: str, 
        version: str, 
        weights_hash: str
    ) -> bool:
        """
        Update the weights hash for a registered model.
        
        Args:
            model_name: Name of the model
            version: Version to update
            weights_hash: New weights hash
            
        Returns:
            True if updated successfully
        """
        if model_name not in cls._MODELS:
            return False
        
        if version not in cls._MODELS[model_name]:
            return False
        
        cls._MODELS[model_name][version].weights_hash = weights_hash
        logger.info(f"Updated weights hash for {model_name} v{version}")
        return True
    
    @classmethod
    def verify_model_hash(
        cls, 
        model_name: str, 
        version: str, 
        computed_hash: str
    ) -> bool:
        """
        Verify a model's computed hash against the registered hash.
        
        Args:
            model_name: Name of the model
            version: Version to verify
            computed_hash: Hash computed from model weights
            
        Returns:
            True if hash matches
        """
        info = cls.get_model_info(model_name, version)
        if info is None:
            return False
        
        # If no hash registered yet, accept and register
        if not info.weights_hash:
            cls.update_weights_hash(model_name, version, computed_hash)
            return True
        
        return info.weights_hash == computed_hash


# Convenience functions
def get_model_info(model_name: str, version: Optional[str] = None) -> Optional[ModelInfo]:
    """Get information about a registered model."""
    return ModelRegistry.get_model_info(model_name, version)


def verify_model_hash(model_name: str, version: str, computed_hash: str) -> bool:
    """Verify a model's computed hash against the registered hash."""
    return ModelRegistry.verify_model_hash(model_name, version, computed_hash)
