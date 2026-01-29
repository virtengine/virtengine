"""
Tests for the face extraction model registry.

Task Reference: VE-3040 - Extract RMIT U-Net ResNet34 Weights
"""

import hashlib
import tempfile
from pathlib import Path
from unittest import mock

import pytest

from ml.face_extraction.model_registry import (
    MODEL_REGISTRY,
    calculate_file_hash,
    ensure_determinism,
    get_model_info,
    get_model_path,
    get_weights_dir,
    list_available_models,
    verify_model_hash,
)


def _has_torch() -> bool:
    """Check if PyTorch is available."""
    try:
        import torch
        return True
    except ImportError:
        return False


class TestModelRegistry:
    """Tests for MODEL_REGISTRY structure."""

    def test_registry_contains_unet(self):
        """Verify unet_resnet34_v1 is in the registry."""
        assert "unet_resnet34_v1" in MODEL_REGISTRY

    def test_unet_has_required_fields(self):
        """Verify unet_resnet34_v1 has all required fields."""
        model = MODEL_REGISTRY["unet_resnet34_v1"]
        required_fields = [
            "filename",
            "sha256",
            "source",
            "description",
            "input_size",
            "output_classes",
            "framework",
        ]
        for field in required_fields:
            assert field in model, f"Missing field: {field}"

    def test_unet_sha256_format(self):
        """Verify SHA256 hash is valid hex string."""
        sha256 = MODEL_REGISTRY["unet_resnet34_v1"]["sha256"]
        assert len(sha256) == 64, "SHA256 should be 64 characters"
        assert all(c in "0123456789abcdef" for c in sha256.lower())


class TestGetModelPath:
    """Tests for get_model_path function."""

    def test_get_model_path_valid(self):
        """Test getting path for valid model."""
        path = get_model_path("unet_resnet34_v1")
        assert path.name == "unet_resnet34.pt"
        assert "weights" in str(path)

    def test_get_model_path_invalid(self):
        """Test that invalid model raises ValueError."""
        with pytest.raises(ValueError, match="Unknown model"):
            get_model_path("nonexistent_model")


class TestGetModelInfo:
    """Tests for get_model_info function."""

    def test_get_model_info_returns_copy(self):
        """Verify get_model_info returns a copy, not the original."""
        info = get_model_info("unet_resnet34_v1")
        info["sha256"] = "modified"
        # Original should be unchanged
        assert MODEL_REGISTRY["unet_resnet34_v1"]["sha256"] != "modified"

    def test_get_model_info_invalid(self):
        """Test that invalid model raises ValueError."""
        with pytest.raises(ValueError):
            get_model_info("nonexistent_model")


class TestCalculateFileHash:
    """Tests for calculate_file_hash function."""

    def test_calculate_hash_known_content(self):
        """Test hash calculation with known content."""
        with tempfile.NamedTemporaryFile(delete=False) as f:
            content = b"test content for hashing"
            f.write(content)
            f.flush()
            
            # Calculate expected hash
            expected = hashlib.sha256(content).hexdigest().lower()
            actual = calculate_file_hash(Path(f.name))
            
            assert actual == expected

    def test_calculate_hash_empty_file(self):
        """Test hash of empty file."""
        with tempfile.NamedTemporaryFile(delete=False) as f:
            f.flush()
            
            expected = hashlib.sha256(b"").hexdigest()
            actual = calculate_file_hash(Path(f.name))
            
            assert actual == expected


class TestVerifyModelHash:
    """Tests for verify_model_hash function."""

    def test_verify_hash_no_hash_registered(self):
        """Test that missing hash returns True with warning."""
        with mock.patch.dict(MODEL_REGISTRY, {
            "test_model": {"filename": "test.pt", "sha256": ""}
        }):
            # Should return True when no hash registered
            assert verify_model_hash("test_model") is True

    def test_verify_hash_file_not_found(self):
        """Test that missing file returns False."""
        with mock.patch.dict(MODEL_REGISTRY, {
            "test_model": {
                "filename": "nonexistent.pt",
                "sha256": "abc123"
            }
        }):
            assert verify_model_hash("test_model") is False


class TestEnsureDeterminism:
    """Tests for ensure_determinism function."""

    def test_ensure_determinism_sets_seeds(self):
        """Test that determinism function sets random seeds."""
        import os
        import random
        
        ensure_determinism(seed=12345)
        
        # Check Python hash seed
        assert os.environ.get("PYTHONHASHSEED") == "12345"
        
        # Check random produces deterministic results
        val1 = random.random()
        ensure_determinism(seed=12345)
        val2 = random.random()
        assert val1 == val2


class TestListAvailableModels:
    """Tests for list_available_models function."""

    def test_list_includes_availability(self):
        """Test that list includes availability status."""
        models = list_available_models()
        
        assert "unet_resnet34_v1" in models
        assert "available" in models["unet_resnet34_v1"]
        assert "path" in models["unet_resnet34_v1"]


class TestGetWeightsDir:
    """Tests for get_weights_dir function."""

    def test_weights_dir_path(self):
        """Test weights directory path is correct."""
        weights_dir = get_weights_dir()
        assert weights_dir.name == "weights"
        assert "face_extraction" in str(weights_dir)


# Integration tests - skipped if weights not present
@pytest.mark.skipif(
    not get_model_path("unet_resnet34_v1").exists(),
    reason="Model weights not extracted yet - run scripts/extract_rmit_weights.py"
)
class TestWithWeights:
    """Integration tests that require actual model weights."""

    def test_verify_hash_with_actual_weights(self):
        """Verify hash of actual weights file."""
        assert verify_model_hash("unet_resnet34_v1")

    @pytest.mark.skipif(
        not _has_torch(),
        reason="PyTorch not installed"
    )
    def test_load_model_deterministic(self):
        """Test loading model with determinism controls."""
        from ml.face_extraction.model_registry import load_model_deterministic
        
        state_dict = load_model_deterministic("unet_resnet34_v1")
        assert state_dict is not None
        assert isinstance(state_dict, dict) or hasattr(state_dict, "keys")
