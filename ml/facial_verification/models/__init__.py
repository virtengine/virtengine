"""
Model management for facial verification.

This module provides model loading, versioning, and hash verification
for deterministic face recognition.
"""

from ml.facial_verification.models.model_registry import (
    ModelRegistry,
    ModelInfo,
    get_model_info,
    verify_model_hash,
)

__all__ = [
    "ModelRegistry",
    "ModelInfo",
    "get_model_info",
    "verify_model_hash",
]
