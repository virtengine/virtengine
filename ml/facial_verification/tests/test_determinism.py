"""
Tests for the determinism module.
"""

import pytest
import numpy as np
import os

from ml.facial_verification.determinism import (
    DeterminismController,
    DeterminismVerifier,
)
from ml.facial_verification.verification import VerificationResult
from ml.facial_verification.config import VerificationConfig, DeterminismConfig


class TestDeterminismController:
    """Tests for DeterminismController class."""
    
    def test_init_default_config(self):
        """Test initialization with default config."""
        controller = DeterminismController()
        assert controller.config is not None
    
    def test_init_custom_config(self, verification_config):
        """Test initialization with custom config."""
        controller = DeterminismController(verification_config)
        assert controller.config == verification_config
    
    def test_ensure_deterministic_sets_seed(self, determinism_controller):
        """Test that ensure_deterministic sets random seed."""
        determinism_controller.ensure_deterministic()
        
        # Generate random numbers
        rand1 = np.random.rand(10)
        
        # Reset and generate again
        determinism_controller._initialized = False
        determinism_controller.ensure_deterministic()
        rand2 = np.random.rand(10)
        
        np.testing.assert_array_equal(rand1, rand2)
    
    def test_ensure_deterministic_sets_numpy_seed(self, determinism_controller):
        """Test that NumPy seed is set."""
        determinism_controller.ensure_deterministic()
        
        result1 = np.random.randint(0, 1000, 10)
        
        np.random.seed(determinism_controller._determinism_config.seed)
        result2 = np.random.randint(0, 1000, 10)
        
        np.testing.assert_array_equal(result1, result2)
    
    def test_force_cpu_mode(self):
        """Test CPU-only mode configuration."""
        config = VerificationConfig()
        config.determinism.force_cpu = True
        controller = DeterminismController(config)
        
        controller.ensure_deterministic()
        
        # Check environment variables
        assert os.environ.get("CUDA_VISIBLE_DEVICES") == "-1"
    
    def test_compute_result_hash(self, determinism_controller):
        """Test computing result hash."""
        result = VerificationResult(
            match=True,
            decision="match",
            similarity_score=0.95,
            confidence_percent=95.0,
            model_name="VGG-Face",
            model_version="1.0.0",
            model_hash="abc123",
            embeddings_hash="xyz789",
            reason_codes=["MATCH_CONFIRMED"],
        )
        
        hash1 = determinism_controller.compute_result_hash(result)
        hash2 = determinism_controller.compute_result_hash(result)
        
        assert hash1 == hash2
        assert len(hash1) == 64  # SHA256
    
    def test_compute_result_hash_different_results(self, determinism_controller):
        """Test that different results produce different hashes."""
        result1 = VerificationResult(
            match=True,
            decision="match",
            similarity_score=0.95,
            confidence_percent=95.0,
            model_name="VGG-Face",
            model_version="1.0.0",
            model_hash="abc123",
            embeddings_hash="xyz789",
        )
        
        result2 = VerificationResult(
            match=False,
            decision="no_match",
            similarity_score=0.50,
            confidence_percent=50.0,
            model_name="VGG-Face",
            model_version="1.0.0",
            model_hash="abc123",
            embeddings_hash="xyz789",
        )
        
        hash1 = determinism_controller.compute_result_hash(result1)
        hash2 = determinism_controller.compute_result_hash(result2)
        
        assert hash1 != hash2
    
    def test_verify_result_hash(self, determinism_controller):
        """Test verifying result hash."""
        result = VerificationResult(
            match=True,
            decision="match",
            similarity_score=0.95,
            confidence_percent=95.0,
            model_name="VGG-Face",
            model_version="1.0.0",
            model_hash="abc123",
            embeddings_hash="xyz789",
        )
        
        expected_hash = determinism_controller.compute_result_hash(result)
        
        assert determinism_controller.verify_result_hash(result, expected_hash) is True
        assert determinism_controller.verify_result_hash(result, "wrong_hash") is False
    
    def test_compute_embedding_hash(self, determinism_controller):
        """Test computing embedding hash."""
        embedding = np.array([0.1, 0.2, 0.3, 0.4, 0.5])
        
        hash1 = determinism_controller.compute_embedding_hash(embedding)
        hash2 = determinism_controller.compute_embedding_hash(embedding)
        
        assert hash1 == hash2
        assert len(hash1) == 64
    
    def test_compute_embedding_hash_different_embeddings(self, determinism_controller):
        """Test that different embeddings produce different hashes."""
        embedding1 = np.array([0.1, 0.2, 0.3, 0.4, 0.5])
        embedding2 = np.array([0.5, 0.4, 0.3, 0.2, 0.1])
        
        hash1 = determinism_controller.compute_embedding_hash(embedding1)
        hash2 = determinism_controller.compute_embedding_hash(embedding2)
        
        assert hash1 != hash2
    
    def test_compute_image_hash(self, determinism_controller, sample_face_image):
        """Test computing image hash."""
        hash1 = determinism_controller.compute_image_hash(sample_face_image)
        hash2 = determinism_controller.compute_image_hash(sample_face_image)
        
        assert hash1 == hash2
        assert len(hash1) == 64
    
    def test_get_environment_info(self, determinism_controller):
        """Test getting environment info."""
        info = determinism_controller.get_environment_info()
        
        assert "seed" in info
        assert "force_cpu" in info
        assert "deterministic_ops" in info
        assert "numpy_version" in info


class TestDeterminismVerifier:
    """Tests for DeterminismVerifier class."""
    
    def test_verify_consistency_identical_results(self):
        """Test consistency verification with identical results."""
        verifier = DeterminismVerifier()
        
        result1 = VerificationResult(
            match=True,
            decision="match",
            similarity_score=0.95,
            confidence_percent=95.0,
            model_name="VGG-Face",
            model_version="1.0.0",
            model_hash="abc123",
            embeddings_hash="xyz789",
        )
        
        result2 = VerificationResult(
            match=True,
            decision="match",
            similarity_score=0.95,
            confidence_percent=95.0,
            model_name="VGG-Face",
            model_version="1.0.0",
            model_hash="abc123",
            embeddings_hash="xyz789",
        )
        
        assert verifier.verify_consistency(result1, result2) is True
    
    def test_verify_consistency_different_match(self):
        """Test consistency verification with different match values."""
        verifier = DeterminismVerifier()
        
        result1 = VerificationResult(
            match=True,
            decision="match",
            similarity_score=0.95,
            confidence_percent=95.0,
            model_name="VGG-Face",
            model_version="1.0.0",
            model_hash="abc123",
            embeddings_hash="xyz789",
        )
        
        result2 = VerificationResult(
            match=False,
            decision="match",
            similarity_score=0.95,
            confidence_percent=95.0,
            model_name="VGG-Face",
            model_version="1.0.0",
            model_hash="abc123",
            embeddings_hash="xyz789",
        )
        
        assert verifier.verify_consistency(result1, result2) is False
    
    def test_verify_consistency_different_decision(self):
        """Test consistency verification with different decisions."""
        verifier = DeterminismVerifier()
        
        result1 = VerificationResult(
            match=True,
            decision="match",
            similarity_score=0.95,
            confidence_percent=95.0,
            model_name="VGG-Face",
            model_version="1.0.0",
            model_hash="abc123",
            embeddings_hash="xyz789",
        )
        
        result2 = VerificationResult(
            match=True,
            decision="borderline",
            similarity_score=0.95,
            confidence_percent=95.0,
            model_name="VGG-Face",
            model_version="1.0.0",
            model_hash="abc123",
            embeddings_hash="xyz789",
        )
        
        assert verifier.verify_consistency(result1, result2) is False
    
    def test_verify_consistency_different_similarity(self):
        """Test consistency verification with different similarity scores."""
        verifier = DeterminismVerifier()
        
        result1 = VerificationResult(
            match=True,
            decision="match",
            similarity_score=0.95,
            confidence_percent=95.0,
            model_name="VGG-Face",
            model_version="1.0.0",
            model_hash="abc123",
            embeddings_hash="xyz789",
        )
        
        result2 = VerificationResult(
            match=True,
            decision="match",
            similarity_score=0.90,
            confidence_percent=95.0,
            model_name="VGG-Face",
            model_version="1.0.0",
            model_hash="abc123",
            embeddings_hash="xyz789",
        )
        
        assert verifier.verify_consistency(result1, result2) is False
    
    def test_verify_consistency_within_tolerance(self):
        """Test consistency verification within tolerance."""
        verifier = DeterminismVerifier()
        
        result1 = VerificationResult(
            match=True,
            decision="match",
            similarity_score=0.950000001,
            confidence_percent=95.0,
            model_name="VGG-Face",
            model_version="1.0.0",
            model_hash="abc123",
            embeddings_hash="xyz789",
        )
        
        result2 = VerificationResult(
            match=True,
            decision="match",
            similarity_score=0.950000002,
            confidence_percent=95.0,
            model_name="VGG-Face",
            model_version="1.0.0",
            model_hash="abc123",
            embeddings_hash="xyz789",
        )
        
        # Within default tolerance (1e-6)
        assert verifier.verify_consistency(result1, result2, tolerance=1e-6) is True
    
    def test_verify_consistency_different_model_hash(self):
        """Test consistency verification with different model hashes."""
        verifier = DeterminismVerifier()
        
        result1 = VerificationResult(
            match=True,
            decision="match",
            similarity_score=0.95,
            confidence_percent=95.0,
            model_name="VGG-Face",
            model_version="1.0.0",
            model_hash="abc123",
            embeddings_hash="xyz789",
        )
        
        result2 = VerificationResult(
            match=True,
            decision="match",
            similarity_score=0.95,
            confidence_percent=95.0,
            model_name="VGG-Face",
            model_version="1.0.0",
            model_hash="def456",  # Different model hash
            embeddings_hash="xyz789",
        )
        
        assert verifier.verify_consistency(result1, result2) is False
    
    def test_verify_consistency_different_embeddings_hash(self):
        """Test consistency verification with different embeddings hashes."""
        verifier = DeterminismVerifier()
        
        result1 = VerificationResult(
            match=True,
            decision="match",
            similarity_score=0.95,
            confidence_percent=95.0,
            model_name="VGG-Face",
            model_version="1.0.0",
            model_hash="abc123",
            embeddings_hash="xyz789",
        )
        
        result2 = VerificationResult(
            match=True,
            decision="match",
            similarity_score=0.95,
            confidence_percent=95.0,
            model_name="VGG-Face",
            model_version="1.0.0",
            model_hash="abc123",
            embeddings_hash="uvw456",  # Different embeddings hash
        )
        
        assert verifier.verify_consistency(result1, result2) is False


class TestDeterminismConfig:
    """Tests for DeterminismConfig."""
    
    def test_default_values(self):
        """Test default configuration values."""
        config = DeterminismConfig()
        
        assert config.force_cpu is True
        assert config.seed == 42
        assert config.deterministic_ops is True
    
    def test_custom_seed(self):
        """Test custom seed configuration."""
        config = DeterminismConfig(seed=12345)
        
        assert config.seed == 12345
    
    def test_disable_determinism(self):
        """Test disabling deterministic operations."""
        config = DeterminismConfig(
            force_cpu=False,
            deterministic_ops=False,
        )
        
        assert config.force_cpu is False
        assert config.deterministic_ops is False


class TestDeterministicRepeatedRuns:
    """Tests for deterministic repeated runs."""
    
    def test_numpy_operations_deterministic(self, determinism_controller):
        """Test that NumPy operations are deterministic."""
        determinism_controller.ensure_deterministic()
        
        # First run
        a = np.random.rand(100, 100)
        b = np.random.rand(100, 100)
        result1 = np.dot(a, b)
        
        # Reset and second run
        determinism_controller._initialized = False
        determinism_controller.ensure_deterministic()
        
        a = np.random.rand(100, 100)
        b = np.random.rand(100, 100)
        result2 = np.dot(a, b)
        
        np.testing.assert_array_almost_equal(result1, result2)
    
    def test_embedding_hash_deterministic(self, determinism_controller):
        """Test that embedding hashes are deterministic."""
        determinism_controller.ensure_deterministic()
        
        # Generate random embedding
        embedding1 = np.random.rand(512)
        hash1 = determinism_controller.compute_embedding_hash(embedding1)
        
        # Reset and generate again
        determinism_controller._initialized = False
        determinism_controller.ensure_deterministic()
        
        embedding2 = np.random.rand(512)
        hash2 = determinism_controller.compute_embedding_hash(embedding2)
        
        # Embeddings should be identical due to same seed
        np.testing.assert_array_equal(embedding1, embedding2)
        assert hash1 == hash2
