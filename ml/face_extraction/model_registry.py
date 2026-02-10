"""
Model registry for face extraction models.

Tracks model versions, hashes, and provides loading utilities
with determinism controls for blockchain consensus.

Task Reference: VE-3040 - Extract RMIT U-Net ResNet34 Weights
"""

import hashlib
import logging
import os
from pathlib import Path
from typing import Any, Dict, Optional

import numpy as np

logger = logging.getLogger(__name__)

# Set deterministic environment variables
os.environ["PYTHONHASHSEED"] = "42"

# Model registry with version tracking
MODEL_REGISTRY: Dict[str, Dict[str, Any]] = {
    "unet_resnet34_v1": {
        "filename": "unet_resnet34.pt",
        "sha256": "0ea89b9d7b249d04ebe767cc38e78d04545067eabdda0f9fc22a1ae2c19bca57",
        "source": "RMIT OCR_Document_Scan intern project (2023)",
        "description": "ResNet34 backbone U-Net for document face extraction",
        "input_size": (512, 512),
        "output_classes": 4,  # Background + 3 document fields
        "trained_on": "Turkish ID documents",
        "framework": "pytorch",
        "architecture": "UNet with ResNet34 encoder",
    }
}


def get_weights_dir() -> Path:
    """Get the weights directory path."""
    return Path(__file__).parent / "weights"


def get_model_path(model_name: str) -> Path:
    """
    Get the path to a registered model.
    
    Args:
        model_name: Name of the model in the registry
        
    Returns:
        Path to the model file
        
    Raises:
        ValueError: If model is not in registry
    """
    if model_name not in MODEL_REGISTRY:
        raise ValueError(f"Unknown model: {model_name}. Available: {list(MODEL_REGISTRY.keys())}")
    
    return get_weights_dir() / MODEL_REGISTRY[model_name]["filename"]


def get_model_info(model_name: str) -> Dict[str, Any]:
    """
    Get metadata for a registered model.
    
    Args:
        model_name: Name of the model in the registry
        
    Returns:
        Dictionary with model metadata
        
    Raises:
        ValueError: If model is not in registry
    """
    if model_name not in MODEL_REGISTRY:
        raise ValueError(f"Unknown model: {model_name}")
    
    return MODEL_REGISTRY[model_name].copy()


def calculate_file_hash(filepath: Path) -> str:
    """
    Calculate SHA256 hash of a file.
    
    Args:
        filepath: Path to the file
        
    Returns:
        Lowercase hex string of SHA256 hash
    """
    sha256_hash = hashlib.sha256()
    with open(filepath, "rb") as f:
        for byte_block in iter(lambda: f.read(65536), b""):
            sha256_hash.update(byte_block)
    return sha256_hash.hexdigest().lower()


def verify_model_hash(model_name: str) -> bool:
    """
    Verify model file matches registered hash.
    
    Args:
        model_name: Name of the model in the registry
        
    Returns:
        True if hash matches or no hash registered, False otherwise
    """
    model_path = get_model_path(model_name)
    expected_hash = MODEL_REGISTRY[model_name]["sha256"]
    
    if not expected_hash:
        logger.warning(f"No hash registered for model {model_name}")
        return True  # No hash registered yet
    
    if not model_path.exists():
        logger.error(f"Model file not found: {model_path}")
        return False
    
    actual_hash = calculate_file_hash(model_path)
    matches = actual_hash == expected_hash.lower()
    
    if not matches:
        logger.error(
            f"Model hash mismatch for {model_name}! "
            f"Expected: {expected_hash}, Got: {actual_hash}"
        )
    
    return matches


def ensure_determinism(seed: int = 42) -> None:
    """
    Set deterministic execution mode for ML frameworks.
    
    This is critical for blockchain consensus - all validators must
    produce identical results.
    
    Args:
        seed: Random seed to use (default: 42)
    """
    import random
    
    # Python random
    random.seed(seed)
    os.environ["PYTHONHASHSEED"] = str(seed)
    
    # NumPy
    np.random.seed(seed)
    
    # PyTorch
    try:
        import torch
        torch.manual_seed(seed)
        torch.cuda.manual_seed_all(seed)
        torch.backends.cudnn.deterministic = True
        torch.backends.cudnn.benchmark = False
        os.environ["CUBLAS_WORKSPACE_CONFIG"] = ":4096:8"
    except ImportError:
        pass
    
    # TensorFlow
    try:
        import tensorflow as tf
        tf.random.set_seed(seed)
        os.environ["TF_DETERMINISTIC_OPS"] = "1"
        os.environ["TF_CUDNN_DETERMINISTIC"] = "1"
    except ImportError:
        pass


def load_model_deterministic(
    model_name: str,
    device: str = "cpu",
    verify_hash: bool = True,
    seed: int = 42,
) -> Any:
    """
    Load a model with determinism controls for consensus.
    
    IMPORTANT: For blockchain consensus, always use device="cpu" to ensure
    identical results across different hardware.
    
    Args:
        model_name: Name of the model in the registry
        device: Device to load model on (forced to "cpu" for consensus)
        verify_hash: Whether to verify model hash before loading
        seed: Random seed for determinism
        
    Returns:
        Model state dict
        
    Raises:
        ValueError: If model hash verification fails
        FileNotFoundError: If model file doesn't exist
    """
    try:
        import torch
    except ImportError:
        raise RuntimeError("PyTorch not available. Install with: pip install torch")
    
    # Force CPU for consensus determinism
    if device != "cpu":
        import warnings
        warnings.warn(
            "Forcing CPU device for consensus determinism. "
            "GPU results may vary across hardware.",
            UserWarning
        )
        device = "cpu"
    
    # Set deterministic mode
    ensure_determinism(seed)
    
    model_path = get_model_path(model_name)
    
    if not model_path.exists():
        raise FileNotFoundError(
            f"Model file not found: {model_path}. "
            f"Run scripts/extract_rmit_weights.py to extract weights."
        )
    
    # Verify hash before loading
    if verify_hash and not verify_model_hash(model_name):
        raise ValueError(
            f"Model hash mismatch for {model_name} - consensus violation! "
            f"Model may be corrupted or tampered with."
        )
    
    logger.info(f"Loading model {model_name} from {model_path}")
    state_dict = torch.load(model_path, map_location=device, weights_only=True)
    
    return state_dict


def list_available_models() -> Dict[str, Dict[str, Any]]:
    """
    List all available models and their status.
    
    Returns:
        Dictionary mapping model names to their info including availability
    """
    result = {}
    for name, info in MODEL_REGISTRY.items():
        model_info = info.copy()
        model_path = get_weights_dir() / info["filename"]
        model_info["available"] = model_path.exists()
        model_info["path"] = str(model_path)
        result[name] = model_info
    return result
