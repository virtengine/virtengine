"""
Tests for U-Net Factory.

Task Reference: VE-3044 - Create U-Net Factory and Training Script
"""

import os
import pytest
import torch
import torch.nn as nn
from typing import List
from unittest import mock

# Ensure deterministic behavior in tests
os.environ["CUDA_VISIBLE_DEVICES"] = "-1"
os.environ["CUBLAS_WORKSPACE_CONFIG"] = ":4096:8"
os.environ["PYTHONHASHSEED"] = "42"


# Import after setting environment
from ml.face_extraction.unet_factory import (
    AttentionGate,
    DecoderBlock,
    DoubleConv,
    EfficientNetEncoder,
    ENCODER_CONFIGS,
    ResNetEncoder,
    UNet,
    UNetDecoder,
    UNetFactory,
    count_parameters,
)


# ============================================================================
# Fixtures
# ============================================================================


@pytest.fixture
def seed():
    """Fixed random seed for reproducibility."""
    return 42


@pytest.fixture
def set_deterministic(seed):
    """Set deterministic mode for tests."""
    torch.manual_seed(seed)
    if torch.cuda.is_available():
        torch.cuda.manual_seed_all(seed)
    torch.backends.cudnn.deterministic = True
    torch.backends.cudnn.benchmark = False


@pytest.fixture
def sample_input():
    """Sample input tensor for testing."""
    return torch.randn(2, 3, 256, 256)


@pytest.fixture
def sample_input_512():
    """Sample 512x512 input tensor for testing."""
    return torch.randn(2, 3, 512, 512)


@pytest.fixture
def sample_mask():
    """Sample binary mask for testing."""
    return torch.randint(0, 2, (2, 1, 256, 256)).float()


@pytest.fixture
def sample_multiclass_mask():
    """Sample multi-class mask for testing."""
    return torch.randint(0, 4, (2, 256, 256))


# ============================================================================
# Test DoubleConv Block
# ============================================================================


class TestDoubleConv:
    """Tests for DoubleConv block."""

    def test_output_shape(self, set_deterministic):
        """Test output shape is correct."""
        block = DoubleConv(64, 128)
        x = torch.randn(2, 64, 32, 32)
        out = block(x)
        assert out.shape == (2, 128, 32, 32)

    def test_with_batch_norm(self, set_deterministic):
        """Test block with batch normalization."""
        block = DoubleConv(64, 128, use_batch_norm=True)
        x = torch.randn(2, 64, 32, 32)
        out = block(x)
        assert out.shape == (2, 128, 32, 32)

    def test_without_batch_norm(self, set_deterministic):
        """Test block without batch normalization."""
        block = DoubleConv(64, 128, use_batch_norm=False)
        x = torch.randn(2, 64, 32, 32)
        out = block(x)
        assert out.shape == (2, 128, 32, 32)

    def test_custom_mid_channels(self, set_deterministic):
        """Test block with custom mid channels."""
        block = DoubleConv(64, 128, mid_channels=96)
        x = torch.randn(2, 64, 32, 32)
        out = block(x)
        assert out.shape == (2, 128, 32, 32)


# ============================================================================
# Test DecoderBlock
# ============================================================================


class TestDecoderBlock:
    """Tests for DecoderBlock."""

    def test_output_shape_with_skip(self, set_deterministic):
        """Test output shape with skip connection."""
        block = DecoderBlock(256, 128, 64)
        x = torch.randn(2, 256, 16, 16)
        skip = torch.randn(2, 128, 32, 32)
        out = block(x, skip)
        assert out.shape == (2, 64, 32, 32)

    def test_output_shape_without_skip(self, set_deterministic):
        """Test output shape without skip connection."""
        block = DecoderBlock(256, 0, 64)
        x = torch.randn(2, 256, 16, 16)
        out = block(x, None)
        assert out.shape == (2, 64, 32, 32)

    def test_handles_size_mismatch(self, set_deterministic):
        """Test that size mismatch is handled."""
        block = DecoderBlock(256, 128, 64)
        x = torch.randn(2, 256, 16, 16)
        skip = torch.randn(2, 128, 31, 31)  # Odd size
        out = block(x, skip)
        assert out.shape == (2, 64, 31, 31)

    def test_with_attention(self, set_deterministic):
        """Test decoder block with attention gate."""
        block = DecoderBlock(256, 128, 64, attention=True)
        x = torch.randn(2, 256, 16, 16)
        skip = torch.randn(2, 128, 32, 32)
        out = block(x, skip)
        assert out.shape == (2, 64, 32, 32)


# ============================================================================
# Test AttentionGate
# ============================================================================


class TestAttentionGate:
    """Tests for AttentionGate."""

    def test_output_shape(self, set_deterministic):
        """Test attention gate preserves input shape."""
        gate = AttentionGate(128, 64, 32)
        g = torch.randn(2, 128, 32, 32)
        x = torch.randn(2, 64, 32, 32)
        out = gate(g, x)
        assert out.shape == x.shape

    def test_attention_range(self, set_deterministic):
        """Test attention weights are in [0, 1] range."""
        gate = AttentionGate(128, 64, 32)
        g = torch.randn(2, 128, 32, 32)
        x = torch.randn(2, 64, 32, 32)
        
        # Access internal attention computation
        g1 = gate.W_g(g)
        x1 = gate.W_x(x)
        psi = gate.relu(g1 + x1)
        attention = gate.psi(psi)
        
        assert attention.min() >= 0.0
        assert attention.max() <= 1.0


# ============================================================================
# Test ResNetEncoder
# ============================================================================


class TestResNetEncoder:
    """Tests for ResNetEncoder."""

    @pytest.mark.parametrize("encoder_name", ["resnet18", "resnet34", "resnet50"])
    def test_encoder_creation(self, encoder_name, set_deterministic):
        """Test encoder creation with different backbones."""
        encoder = ResNetEncoder(encoder_name=encoder_name, pretrained=False)
        assert encoder.encoder_name == encoder_name

    @pytest.mark.parametrize("encoder_name", ["resnet18", "resnet34", "resnet50"])
    def test_encoder_output_shapes(self, encoder_name, sample_input, set_deterministic):
        """Test encoder produces correct number of feature maps."""
        encoder = ResNetEncoder(encoder_name=encoder_name, pretrained=False)
        features = encoder(sample_input)
        
        assert len(features) == 5  # 5 stages
        
        # Check expected channels
        expected_channels = ENCODER_CONFIGS[encoder_name]["out_channels"]
        for i, (feat, expected_ch) in enumerate(zip(features, expected_channels)):
            assert feat.shape[1] == expected_ch, f"Stage {i}: expected {expected_ch}, got {feat.shape[1]}"

    def test_encoder_custom_in_channels(self, set_deterministic):
        """Test encoder with custom input channels."""
        encoder = ResNetEncoder(encoder_name="resnet34", pretrained=False, in_channels=1)
        x = torch.randn(2, 1, 256, 256)
        features = encoder(x)
        assert len(features) == 5

    def test_invalid_encoder_raises(self, set_deterministic):
        """Test that invalid encoder name raises ValueError."""
        with pytest.raises(ValueError, match="Unknown encoder"):
            ResNetEncoder(encoder_name="invalid_encoder")


# ============================================================================
# Test EfficientNetEncoder
# ============================================================================


class TestEfficientNetEncoder:
    """Tests for EfficientNetEncoder."""

    def test_encoder_creation(self, set_deterministic):
        """Test encoder creation."""
        encoder = EfficientNetEncoder(pretrained=False)
        assert encoder.encoder_name == "efficientnet-b0"

    def test_encoder_output_count(self, sample_input, set_deterministic):
        """Test encoder produces correct number of feature maps."""
        encoder = EfficientNetEncoder(pretrained=False)
        features = encoder(sample_input)
        assert len(features) == 5

    def test_encoder_custom_in_channels(self, set_deterministic):
        """Test encoder with custom input channels."""
        encoder = EfficientNetEncoder(pretrained=False, in_channels=1)
        x = torch.randn(2, 1, 256, 256)
        features = encoder(x)
        assert len(features) == 5


# ============================================================================
# Test UNet Model
# ============================================================================


class TestUNet:
    """Tests for UNet model."""

    @pytest.mark.parametrize("encoder_name", ["resnet18", "resnet34", "resnet50", "efficientnet-b0"])
    def test_model_creation(self, encoder_name, set_deterministic):
        """Test model creation with different backbones."""
        model = UNet(encoder_name=encoder_name, pretrained=False)
        assert model.encoder_name == encoder_name

    @pytest.mark.parametrize("encoder_name", ["resnet18", "resnet34", "efficientnet-b0"])
    def test_forward_pass_shapes(self, encoder_name, sample_input, set_deterministic):
        """Test forward pass produces correct output shape."""
        model = UNet(encoder_name=encoder_name, pretrained=False, out_channels=1)
        output = model(sample_input)
        
        # Output should match input spatial dimensions
        assert output.shape == (2, 1, 256, 256)

    def test_multiclass_output(self, sample_input, set_deterministic):
        """Test multi-class output."""
        model = UNet(encoder_name="resnet34", pretrained=False, out_channels=4)
        output = model(sample_input)
        assert output.shape == (2, 4, 256, 256)

    def test_512_input_size(self, sample_input_512, set_deterministic):
        """Test with 512x512 input size."""
        model = UNet(encoder_name="resnet34", pretrained=False, out_channels=4)
        output = model(sample_input_512)
        assert output.shape == (2, 4, 512, 512)

    def test_freeze_encoder(self, set_deterministic):
        """Test freezing encoder weights."""
        model = UNet(encoder_name="resnet34", pretrained=False)
        
        # Initially all trainable
        trainable_before = count_parameters(model, trainable_only=True)
        
        model.freeze_encoder()
        
        # Encoder params should be frozen
        for param in model.encoder.parameters():
            assert not param.requires_grad
        
        trainable_after = count_parameters(model, trainable_only=True)
        assert trainable_after < trainable_before

    def test_unfreeze_encoder(self, set_deterministic):
        """Test unfreezing encoder weights."""
        model = UNet(encoder_name="resnet34", pretrained=False)
        model.freeze_encoder()
        model.unfreeze_encoder()
        
        for param in model.encoder.parameters():
            assert param.requires_grad

    def test_get_config(self, set_deterministic):
        """Test getting model configuration."""
        model = UNet(
            encoder_name="resnet34",
            pretrained=False,
            out_channels=4,
            attention=True,
        )
        config = model.get_config()
        
        assert config["encoder_name"] == "resnet34"
        assert config["out_channels"] == 4
        assert config["attention"] is True

    def test_custom_decoder_channels(self, sample_input, set_deterministic):
        """Test model with custom decoder channels."""
        model = UNet(
            encoder_name="resnet34",
            pretrained=False,
            decoder_channels=[128, 64, 32, 16, 8],
        )
        output = model(sample_input)
        assert output.shape == (2, 1, 256, 256)

    def test_with_dropout(self, sample_input, set_deterministic):
        """Test model with dropout."""
        model = UNet(encoder_name="resnet34", pretrained=False, dropout=0.5)
        model.train()
        
        # Run forward pass - dropout should be active
        _ = model(sample_input)  # Just verify it runs
        
        # In eval mode outputs should be deterministic
        model.eval()
        output1 = model(sample_input)
        output2 = model(sample_input)
        assert torch.allclose(output1, output2)  # Same in eval mode


# ============================================================================
# Test UNetFactory
# ============================================================================


class TestUNetFactory:
    """Tests for UNetFactory."""

    def test_list_encoders(self):
        """Test listing available encoders."""
        encoders = UNetFactory.list_encoders()
        
        assert "resnet18" in encoders
        assert "resnet34" in encoders
        assert "resnet50" in encoders
        assert "efficientnet-b0" in encoders

    @pytest.mark.parametrize("encoder_name", UNetFactory.SUPPORTED_ENCODERS)
    def test_create_unet_all_encoders(self, encoder_name, set_deterministic):
        """Test creating U-Net with all supported encoders."""
        model = UNetFactory.create_unet(
            encoder_name=encoder_name,
            pretrained=False,
            deterministic=True,
        )
        assert isinstance(model, UNet)

    def test_create_unet_default_params(self, set_deterministic):
        """Test creating U-Net with default parameters."""
        model = UNetFactory.create_unet(pretrained=False, deterministic=True)
        
        assert model.encoder_name == "resnet34"
        assert model.out_channels == 1

    def test_create_unet_custom_params(self, set_deterministic):
        """Test creating U-Net with custom parameters."""
        model = UNetFactory.create_unet(
            encoder_name="resnet50",
            pretrained=False,
            in_channels=1,
            out_channels=4,
            decoder_channels=[512, 256, 128, 64, 32],
            deterministic=True,
        )
        
        assert model.encoder_name == "resnet50"
        assert model.in_channels == 1
        assert model.out_channels == 4

    def test_create_unet_for_face_extraction(self, set_deterministic):
        """Test creating U-Net specifically for face extraction."""
        model = UNetFactory.create_unet_for_face_extraction(
            pretrained=False,
            deterministic=True,
        )
        
        assert model.encoder_name == "resnet34"
        assert model.out_channels == 4

    def test_invalid_encoder_raises(self, set_deterministic):
        """Test that invalid encoder raises ValueError."""
        with pytest.raises(ValueError, match="Unsupported encoder"):
            UNetFactory.create_unet(encoder_name="invalid")

    def test_get_encoder_info(self):
        """Test getting encoder info."""
        info = UNetFactory.get_encoder_info("resnet34")
        
        assert info["name"] == "resnet34"
        assert "out_channels" in info
        assert "model_fn" not in info  # Should be removed

    def test_get_encoder_info_invalid_raises(self):
        """Test that invalid encoder raises ValueError."""
        with pytest.raises(ValueError, match="Unknown encoder"):
            UNetFactory.get_encoder_info("invalid")

    def test_deterministic_mode(self, set_deterministic):
        """Test deterministic mode produces consistent results."""
        # Create two models with same seed
        model1 = UNetFactory.create_unet(
            encoder_name="resnet34",
            pretrained=False,
            deterministic=True,
            seed=42,
        )
        model2 = UNetFactory.create_unet(
            encoder_name="resnet34",
            pretrained=False,
            deterministic=True,
            seed=42,
        )
        
        # Compare weights
        for (n1, p1), (n2, p2) in zip(
            model1.named_parameters(), model2.named_parameters()
        ):
            assert n1 == n2
            assert torch.allclose(p1, p2), f"Mismatch in {n1}"


# ============================================================================
# Test Training Step (Mock)
# ============================================================================


class TestTrainingStep:
    """Tests for training step with mock data."""

    def test_forward_backward_pass(self, sample_input, sample_mask, set_deterministic):
        """Test complete forward and backward pass."""
        model = UNetFactory.create_unet(
            encoder_name="resnet34",
            pretrained=False,
            out_channels=1,
            deterministic=True,
        )
        
        # Forward pass
        output = model(sample_input)
        
        # Compute loss
        loss = torch.nn.functional.binary_cross_entropy_with_logits(
            output, sample_mask
        )
        
        # Backward pass
        loss.backward()
        
        # Check gradients exist
        for param in model.parameters():
            if param.requires_grad:
                assert param.grad is not None

    def test_multiclass_training_step(
        self, sample_input, sample_multiclass_mask, set_deterministic
    ):
        """Test training step with multi-class output."""
        model = UNetFactory.create_unet(
            encoder_name="resnet34",
            pretrained=False,
            out_channels=4,
            deterministic=True,
        )
        
        # Forward pass
        output = model(sample_input)
        
        # Compute loss
        loss = torch.nn.functional.cross_entropy(output, sample_multiclass_mask)
        
        # Backward pass
        loss.backward()
        
        # Check loss is valid
        assert not torch.isnan(loss)
        assert not torch.isinf(loss)

    def test_training_with_frozen_encoder(
        self, sample_input, sample_mask, set_deterministic
    ):
        """Test training with frozen encoder."""
        model = UNetFactory.create_unet(
            encoder_name="resnet34",
            pretrained=False,
            out_channels=1,
            deterministic=True,
        )
        model.freeze_encoder()
        
        # Forward and backward
        output = model(sample_input)
        loss = torch.nn.functional.binary_cross_entropy_with_logits(
            output, sample_mask
        )
        loss.backward()
        
        # Encoder gradients should be None
        for param in model.encoder.parameters():
            assert param.grad is None

        # Decoder gradients should exist
        for param in model.decoder.parameters():
            if param.requires_grad:
                assert param.grad is not None


# ============================================================================
# Test count_parameters
# ============================================================================


class TestCountParameters:
    """Tests for count_parameters utility."""

    def test_count_all_parameters(self, set_deterministic):
        """Test counting all parameters."""
        model = UNetFactory.create_unet(
            encoder_name="resnet18",
            pretrained=False,
            deterministic=True,
        )
        count = count_parameters(model, trainable_only=False)
        assert count > 0

    def test_count_trainable_parameters(self, set_deterministic):
        """Test counting trainable parameters only."""
        model = UNetFactory.create_unet(
            encoder_name="resnet18",
            pretrained=False,
            deterministic=True,
        )
        
        total = count_parameters(model, trainable_only=False)
        trainable = count_parameters(model, trainable_only=True)
        
        assert trainable == total  # All trainable initially
        
        model.freeze_encoder()
        trainable_after = count_parameters(model, trainable_only=True)
        assert trainable_after < total

    def test_resnet34_param_count(self, set_deterministic):
        """Test ResNet34 U-Net has reasonable parameter count."""
        model = UNetFactory.create_unet(
            encoder_name="resnet34",
            pretrained=False,
            out_channels=4,
            deterministic=True,
        )
        
        count = count_parameters(model)
        # ResNet34 encoder ~21M params + decoder
        assert 20_000_000 < count < 40_000_000


# ============================================================================
# Integration Tests
# ============================================================================


class TestIntegration:
    """Integration tests for the U-Net factory."""

    def test_end_to_end_binary_segmentation(self, set_deterministic):
        """Test end-to-end binary segmentation pipeline."""
        # Create model
        model = UNetFactory.create_unet(
            encoder_name="resnet34",
            pretrained=False,
            out_channels=1,
            deterministic=True,
        )
        model.eval()
        
        # Create test input
        x = torch.randn(1, 3, 256, 256)
        
        # Forward pass
        with torch.no_grad():
            logits = model(x)
            probs = torch.sigmoid(logits)
            mask = (probs > 0.5).float()
        
        # Verify outputs
        assert logits.shape == (1, 1, 256, 256)
        assert probs.min() >= 0.0
        assert probs.max() <= 1.0
        assert mask.unique().numel() <= 2  # Binary mask

    def test_end_to_end_multiclass_segmentation(self, set_deterministic):
        """Test end-to-end multi-class segmentation pipeline."""
        num_classes = 4
        
        # Create model
        model = UNetFactory.create_unet(
            encoder_name="resnet34",
            pretrained=False,
            out_channels=num_classes,
            deterministic=True,
        )
        model.eval()
        
        # Create test input
        x = torch.randn(1, 3, 512, 512)
        
        # Forward pass
        with torch.no_grad():
            logits = model(x)
            probs = torch.softmax(logits, dim=1)
            mask = torch.argmax(logits, dim=1)
        
        # Verify outputs
        assert logits.shape == (1, num_classes, 512, 512)
        assert probs.shape == (1, num_classes, 512, 512)
        assert mask.shape == (1, 512, 512)
        assert mask.max() < num_classes
        assert mask.min() >= 0

    def test_model_save_load(self, tmp_path, set_deterministic):
        """Test saving and loading model weights."""
        # Create and save model
        model = UNetFactory.create_unet(
            encoder_name="resnet34",
            pretrained=False,
            out_channels=4,
            deterministic=True,
            seed=42,
        )
        
        save_path = tmp_path / "test_model.pt"
        torch.save(model.state_dict(), save_path)
        
        # Create new model and load weights
        new_model = UNetFactory.create_unet(
            encoder_name="resnet34",
            pretrained=False,
            out_channels=4,
            deterministic=True,
            seed=42,
        )
        new_model.load_state_dict(torch.load(save_path, weights_only=True))
        
        # Verify weights match
        model.eval()
        new_model.eval()
        
        x = torch.randn(1, 3, 256, 256)
        with torch.no_grad():
            out1 = model(x)
            out2 = new_model(x)
        
        assert torch.allclose(out1, out2)
