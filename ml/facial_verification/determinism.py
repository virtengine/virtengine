"""
Determinism controls for face verification.

This module ensures that face verification produces identical results
across different runs and validators for blockchain consensus.
"""

import logging
import os
import hashlib
import random
from typing import Optional, Any

import numpy as np

from ml.facial_verification.config import VerificationConfig

logger = logging.getLogger(__name__)


class DeterminismController:
    """
    Controller for ensuring deterministic execution.
    
    This class manages all aspects of determinism required for
    blockchain consensus:
    - Random seed management
    - CPU-only execution
    - Disabling non-deterministic operations
    - Result hashing
    """
    
    def __init__(self, config: Optional[VerificationConfig] = None):
        """
        Initialize the determinism controller.
        
        Args:
            config: Verification configuration. Uses defaults if not provided.
        """
        self.config = config or VerificationConfig()
        self._determinism_config = self.config.determinism
        self._initialized = False
    
    def ensure_deterministic(self) -> None:
        """
        Set up deterministic execution environment.
        
        This method should be called before any ML operations to ensure
        consistent results across validators.
        """
        if self._initialized and not self._should_reinitialize():
            return
        
        # Set Python random seed
        random.seed(self._determinism_config.seed)
        
        # Set NumPy random seed
        np.random.seed(self._determinism_config.seed)
        
        # Set environment variables before importing TensorFlow
        if self._determinism_config.force_cpu:
            self._force_cpu_mode()
        
        if self._determinism_config.deterministic_ops:
            self._enable_deterministic_ops()
        
        # Set OpenCV-specific settings
        self._configure_opencv()

        # Set TensorFlow-specific settings
        self._configure_tensorflow()
        
        self._initialized = True
        logger.info("Deterministic execution environment configured")
    
    def _should_reinitialize(self) -> bool:
        """Check if re-initialization is needed."""
        # Re-initialize if seed has changed
        return False
    
    def _force_cpu_mode(self) -> None:
        """Force CPU-only execution."""
        # Disable CUDA
        os.environ["CUDA_VISIBLE_DEVICES"] = "-1"
        os.environ["TF_FORCE_CPU"] = "1"
        
        # Disable GPU for various frameworks
        os.environ["USE_GPU"] = "0"
        os.environ["CUDA_DEVICE_ORDER"] = ""
        
        logger.debug("CPU-only mode enabled")
    
    def _enable_deterministic_ops(self) -> None:
        """Enable deterministic operations."""
        # TensorFlow determinism
        os.environ["TF_DETERMINISTIC_OPS"] = "1"
        os.environ["TF_CUDNN_DETERMINISTIC"] = "1"
        
        # Disable cuDNN auto-tuning
        os.environ["TF_CUDNN_USE_AUTOTUNE"] = "0"
        
        # PyTorch determinism (if used)
        os.environ["CUBLAS_WORKSPACE_CONFIG"] = ":4096:8"
        
        logger.debug("Deterministic operations enabled")
    
    def _configure_tensorflow(self) -> None:
        """Configure TensorFlow for deterministic execution."""
        try:
            import tensorflow as tf
            
            # Set random seed
            tf.random.set_seed(self._determinism_config.seed)
            
            # Disable GPU if required
            if self._determinism_config.force_cpu:
                tf.config.set_visible_devices([], 'GPU')
            
            # Enable deterministic ops
            if self._determinism_config.deterministic_ops:
                try:
                    tf.config.experimental.enable_op_determinism()
                except AttributeError:
                    # Older TensorFlow versions
                    pass
            
            # Set inter/intra op parallelism for reproducibility
            tf.config.threading.set_inter_op_parallelism_threads(1)
            tf.config.threading.set_intra_op_parallelism_threads(1)
            
            logger.debug("TensorFlow configured for deterministic execution")
            
        except ImportError:
            logger.warning("TensorFlow not available for configuration")
        except Exception as e:
            logger.warning(f"Error configuring TensorFlow: {e}")

    def _configure_opencv(self) -> None:
        """Configure OpenCV for deterministic execution."""
        try:
            import cv2

            # Ensure single-threaded deterministic execution
            cv2.setNumThreads(1)
            cv2.setUseOptimized(False)

            logger.debug("OpenCV configured for deterministic execution")
        except ImportError:
            logger.warning("OpenCV not available for configuration")
        except Exception as e:
            logger.warning(f"Error configuring OpenCV: {e}")
    
    def compute_result_hash(self, result: Any) -> str:
        """
        Compute deterministic hash of verification result.
        
        This hash is used for consensus verification to ensure
        all validators produce the same result.
        
        Args:
            result: VerificationResult object
            
        Returns:
            SHA256 hash of result
        """
        # Create canonical representation
        canonical = self._create_canonical_representation(result)
        
        # Compute hash
        hash_obj = hashlib.sha256()
        hash_obj.update(canonical.encode('utf-8'))
        
        return hash_obj.hexdigest()
    
    def _create_canonical_representation(self, result: Any) -> str:
        """
        Create a canonical string representation of a result.
        
        Args:
            result: VerificationResult object
            
        Returns:
            Canonical string representation
        """
        # Extract relevant fields in deterministic order
        fields = [
            ("match", result.match),
            ("decision", result.decision),
            ("similarity_score", round(result.similarity_score, 6)),
            ("model_name", result.model_name),
            ("model_version", result.model_version),
            ("model_hash", result.model_hash),
            ("embeddings_hash", result.embeddings_hash),
            ("reason_codes", sorted(result.reason_codes)),
        ]
        
        # Create canonical string
        parts = []
        for key, value in fields:
            if isinstance(value, list):
                value = ",".join(str(v) for v in value)
            parts.append(f"{key}={value}")
        
        return "|".join(parts)
    
    def verify_result_hash(self, result: Any, expected_hash: str) -> bool:
        """
        Verify that a result matches an expected hash.
        
        Args:
            result: VerificationResult object
            expected_hash: Expected hash value
            
        Returns:
            True if hash matches
        """
        computed_hash = self.compute_result_hash(result)
        return computed_hash == expected_hash
    
    def compute_model_hash(self, model: Any) -> str:
        """
        Compute hash of model weights.
        
        Args:
            model: Keras/TensorFlow model
            
        Returns:
            SHA256 hash of weights
        """
        try:
            weights = model.get_weights()
            
            hash_obj = hashlib.sha256()
            for w in weights:
                # Round to reduce floating-point precision issues
                w_rounded = np.round(w, decimals=6)
                hash_obj.update(w_rounded.tobytes())
            
            return hash_obj.hexdigest()
            
        except Exception as e:
            logger.warning(f"Could not compute model hash: {e}")
            return ""
    
    def compute_embedding_hash(self, embedding: np.ndarray) -> str:
        """
        Compute hash of embedding vector.
        
        Args:
            embedding: Face embedding vector
            
        Returns:
            SHA256 hash
        """
        # Round to reduce floating-point precision issues
        embedding_rounded = np.round(embedding, decimals=6)
        return hashlib.sha256(embedding_rounded.tobytes()).hexdigest()
    
    def compute_image_hash(self, image: np.ndarray) -> str:
        """
        Compute hash of image data.
        
        Args:
            image: Image as numpy array
            
        Returns:
            SHA256 hash
        """
        if image.dtype != np.float32:
            image = image.astype(np.float32)
        
        # Round to reduce floating-point precision issues
        image_rounded = np.round(image, decimals=4)
        return hashlib.sha256(image_rounded.tobytes()).hexdigest()
    
    def get_environment_info(self) -> dict:
        """Get information about the current determinism environment."""
        info = {
            "seed": self._determinism_config.seed,
            "force_cpu": self._determinism_config.force_cpu,
            "deterministic_ops": self._determinism_config.deterministic_ops,
            "initialized": self._initialized,
        }
        
        # Add TensorFlow info
        try:
            import tensorflow as tf
            info["tensorflow_version"] = tf.__version__
            info["tensorflow_gpu_available"] = len(tf.config.list_physical_devices('GPU')) > 0
        except ImportError:
            info["tensorflow_version"] = "not installed"
        
        # Add NumPy info
        info["numpy_version"] = np.__version__
        
        return info


class DeterminismVerifier:
    """
    Verifier for checking determinism of verification results.
    
    This class provides utilities for validators to verify that
    their results match those of other validators.
    """
    
    def __init__(self):
        """Initialize the verifier."""
        self._controller = DeterminismController()
    
    def verify_consistency(
        self,
        result1: Any,
        result2: Any,
        tolerance: float = 1e-6
    ) -> bool:
        """
        Verify that two results are consistent.
        
        Args:
            result1: First verification result
            result2: Second verification result
            tolerance: Tolerance for floating-point comparison
            
        Returns:
            True if results are consistent
        """
        # Check decision match
        if result1.match != result2.match:
            return False
        
        if result1.decision != result2.decision:
            return False
        
        # Check similarity score within tolerance
        if abs(result1.similarity_score - result2.similarity_score) > tolerance:
            return False
        
        # Check model info
        if result1.model_name != result2.model_name:
            return False
        
        if result1.model_version != result2.model_version:
            return False
        
        if result1.model_hash != result2.model_hash:
            return False
        
        # Check embeddings hash
        if result1.embeddings_hash != result2.embeddings_hash:
            return False
        
        return True
    
    def run_determinism_test(
        self,
        verifier: Any,
        probe_image: np.ndarray,
        reference_image: np.ndarray,
        num_runs: int = 5
    ) -> bool:
        """
        Run multiple verification passes to test determinism.
        
        Args:
            verifier: FaceVerifier instance
            probe_image: Probe image
            reference_image: Reference image
            num_runs: Number of runs to test
            
        Returns:
            True if all runs produce identical results
        """
        results = []
        hashes = []
        
        for i in range(num_runs):
            result = verifier.verify(probe_image, reference_image)
            results.append(result)
            hashes.append(result.result_hash)
        
        # All hashes should be identical
        return len(set(hashes)) == 1
