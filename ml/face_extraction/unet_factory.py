"""
U-Net Factory for Face Extraction Models.

This module provides a factory class for creating U-Net segmentation models
with various encoder backbones. Supports deterministic initialization for
blockchain consensus requirements.

Task Reference: VE-3044 - Create U-Net Factory and Training Script
"""

import logging
import os
from typing import Any, Dict, List, Optional, Tuple

import torch
import torch.nn as nn
import torch.nn.functional as F
from torchvision import models
from torchvision.models import (
    ResNet18_Weights,
    ResNet34_Weights,
    ResNet50_Weights,
    EfficientNet_B0_Weights,
)

logger = logging.getLogger(__name__)

# Set deterministic environment variables
os.environ["CUBLAS_WORKSPACE_CONFIG"] = ":4096:8"


# Encoder configurations
ENCODER_CONFIGS = {
    "resnet18": {
        "model_fn": models.resnet18,
        "weights": ResNet18_Weights.IMAGENET1K_V1,
        "out_channels": [64, 64, 128, 256, 512],
    },
    "resnet34": {
        "model_fn": models.resnet34,
        "weights": ResNet34_Weights.IMAGENET1K_V1,
        "out_channels": [64, 64, 128, 256, 512],
    },
    "resnet50": {
        "model_fn": models.resnet50,
        "weights": ResNet50_Weights.IMAGENET1K_V1,
        "out_channels": [64, 256, 512, 1024, 2048],
    },
    "efficientnet-b0": {
        "model_fn": models.efficientnet_b0,
        "weights": EfficientNet_B0_Weights.IMAGENET1K_V1,
        "out_channels": [16, 24, 40, 112, 320],
    },
}


class DoubleConv(nn.Module):
    """Double convolution block: (Conv2d -> BN -> ReLU) * 2"""
    
    def __init__(
        self,
        in_channels: int,
        out_channels: int,
        mid_channels: Optional[int] = None,
        use_batch_norm: bool = True,
    ):
        super().__init__()
        mid_channels = mid_channels or out_channels
        
        if use_batch_norm:
            self.double_conv = nn.Sequential(
                nn.Conv2d(in_channels, mid_channels, kernel_size=3, padding=1, bias=False),
                nn.BatchNorm2d(mid_channels),
                nn.ReLU(inplace=True),
                nn.Conv2d(mid_channels, out_channels, kernel_size=3, padding=1, bias=False),
                nn.BatchNorm2d(out_channels),
                nn.ReLU(inplace=True),
            )
        else:
            self.double_conv = nn.Sequential(
                nn.Conv2d(in_channels, mid_channels, kernel_size=3, padding=1),
                nn.ReLU(inplace=True),
                nn.Conv2d(mid_channels, out_channels, kernel_size=3, padding=1),
                nn.ReLU(inplace=True),
            )
    
    def forward(self, x: torch.Tensor) -> torch.Tensor:
        return self.double_conv(x)


class DecoderBlock(nn.Module):
    """U-Net decoder block with skip connection."""
    
    def __init__(
        self,
        in_channels: int,
        skip_channels: int,
        out_channels: int,
        use_batch_norm: bool = True,
        attention: bool = False,
    ):
        super().__init__()
        self.upsample = nn.ConvTranspose2d(
            in_channels, in_channels // 2, kernel_size=2, stride=2
        )
        self.conv = DoubleConv(
            in_channels // 2 + skip_channels,
            out_channels,
            use_batch_norm=use_batch_norm,
        )
        self.attention = attention
        
        if attention:
            self.attention_gate = AttentionGate(
                gate_channels=in_channels // 2,
                skip_channels=skip_channels,
                inter_channels=skip_channels // 2,
            )
    
    def forward(
        self,
        x: torch.Tensor,
        skip: Optional[torch.Tensor] = None,
    ) -> torch.Tensor:
        x = self.upsample(x)
        
        if skip is not None:
            # Handle size mismatch
            if x.shape[2:] != skip.shape[2:]:
                x = F.interpolate(x, size=skip.shape[2:], mode="bilinear", align_corners=True)
            
            if self.attention:
                skip = self.attention_gate(x, skip)
            
            x = torch.cat([x, skip], dim=1)
        
        return self.conv(x)


class AttentionGate(nn.Module):
    """Attention gate for skip connections."""
    
    def __init__(
        self,
        gate_channels: int,
        skip_channels: int,
        inter_channels: int,
    ):
        super().__init__()
        self.W_g = nn.Conv2d(gate_channels, inter_channels, kernel_size=1, bias=False)
        self.W_x = nn.Conv2d(skip_channels, inter_channels, kernel_size=1, bias=False)
        self.psi = nn.Sequential(
            nn.Conv2d(inter_channels, 1, kernel_size=1, bias=False),
            nn.Sigmoid(),
        )
        self.relu = nn.ReLU(inplace=True)
    
    def forward(
        self,
        g: torch.Tensor,
        x: torch.Tensor,
    ) -> torch.Tensor:
        g1 = self.W_g(g)
        x1 = self.W_x(x)
        
        # Handle size mismatch
        if g1.shape[2:] != x1.shape[2:]:
            g1 = F.interpolate(g1, size=x1.shape[2:], mode="bilinear", align_corners=True)
        
        psi = self.relu(g1 + x1)
        psi = self.psi(psi)
        return x * psi


class ResNetEncoder(nn.Module):
    """ResNet-based encoder for U-Net."""
    
    def __init__(
        self,
        encoder_name: str = "resnet34",
        pretrained: bool = True,
        in_channels: int = 3,
    ):
        super().__init__()
        
        if encoder_name not in ENCODER_CONFIGS:
            raise ValueError(
                f"Unknown encoder: {encoder_name}. "
                f"Available: {list(ENCODER_CONFIGS.keys())}"
            )
        
        config = ENCODER_CONFIGS[encoder_name]
        weights = config["weights"] if pretrained else None
        resnet = config["model_fn"](weights=weights)
        
        self.out_channels = config["out_channels"]
        self.encoder_name = encoder_name
        
        # Modify first conv if in_channels != 3
        if in_channels != 3:
            self.conv1 = nn.Conv2d(
                in_channels, 64, kernel_size=7, stride=2, padding=3, bias=False
            )
        else:
            self.conv1 = resnet.conv1
        
        self.bn1 = resnet.bn1
        self.relu = resnet.relu
        self.maxpool = resnet.maxpool
        
        self.layer1 = resnet.layer1
        self.layer2 = resnet.layer2
        self.layer3 = resnet.layer3
        self.layer4 = resnet.layer4
    
    def forward(self, x: torch.Tensor) -> List[torch.Tensor]:
        """
        Forward pass returning feature maps at different scales.
        
        Returns:
            List of feature maps from shallow to deep
        """
        features = []
        
        # Initial block
        x = self.conv1(x)
        x = self.bn1(x)
        x = self.relu(x)
        features.append(x)  # 1/2
        
        x = self.maxpool(x)
        x = self.layer1(x)
        features.append(x)  # 1/4
        
        x = self.layer2(x)
        features.append(x)  # 1/8
        
        x = self.layer3(x)
        features.append(x)  # 1/16
        
        x = self.layer4(x)
        features.append(x)  # 1/32
        
        return features


class EfficientNetEncoder(nn.Module):
    """EfficientNet-based encoder for U-Net."""
    
    def __init__(
        self,
        pretrained: bool = True,
        in_channels: int = 3,
    ):
        super().__init__()
        
        config = ENCODER_CONFIGS["efficientnet-b0"]
        weights = config["weights"] if pretrained else None
        efficientnet = config["model_fn"](weights=weights)
        
        self.out_channels = config["out_channels"]
        self.encoder_name = "efficientnet-b0"
        
        # Extract feature layers from EfficientNet
        self.features = efficientnet.features
        
        # Indices for skip connections (after each MBConv block)
        self.stage_indices = [1, 2, 3, 5, 7]
        
        # Modify first conv if in_channels != 3
        if in_channels != 3:
            old_conv = self.features[0][0]
            self.features[0][0] = nn.Conv2d(
                in_channels,
                old_conv.out_channels,
                kernel_size=old_conv.kernel_size,
                stride=old_conv.stride,
                padding=old_conv.padding,
                bias=False,
            )
    
    def forward(self, x: torch.Tensor) -> List[torch.Tensor]:
        """Forward pass returning feature maps at different scales."""
        features = []
        
        for i, layer in enumerate(self.features):
            x = layer(x)
            if i in self.stage_indices:
                features.append(x)
        
        return features


class UNetDecoder(nn.Module):
    """U-Net decoder with configurable channel sizes."""
    
    def __init__(
        self,
        encoder_channels: List[int],
        decoder_channels: List[int],
        use_batch_norm: bool = True,
        attention: bool = False,
    ):
        super().__init__()
        
        # Reverse encoder channels (deep to shallow)
        encoder_channels = encoder_channels[::-1]
        
        # Center block
        self.center = DoubleConv(
            encoder_channels[0],
            encoder_channels[0],
            use_batch_norm=use_batch_norm,
        )
        
        # Decoder blocks
        self.blocks = nn.ModuleList()
        
        in_channels = encoder_channels[0]
        for i, (skip_ch, out_ch) in enumerate(
            zip(encoder_channels[1:], decoder_channels)
        ):
            self.blocks.append(
                DecoderBlock(
                    in_channels=in_channels,
                    skip_channels=skip_ch,
                    out_channels=out_ch,
                    use_batch_norm=use_batch_norm,
                    attention=attention,
                )
            )
            in_channels = out_ch
    
    def forward(
        self,
        features: List[torch.Tensor],
    ) -> torch.Tensor:
        """
        Forward pass with skip connections.
        
        Args:
            features: List of encoder feature maps (shallow to deep)
        """
        # Reverse features (deep to shallow)
        features = features[::-1]
        
        x = self.center(features[0])
        
        for i, block in enumerate(self.blocks):
            skip = features[i + 1] if i + 1 < len(features) else None
            x = block(x, skip)
        
        return x


class UNet(nn.Module):
    """
    U-Net segmentation model with configurable encoder backbone.
    
    This implementation supports:
    - Multiple encoder backbones (ResNet, EfficientNet)
    - Pretrained encoder initialization
    - Configurable decoder channels
    - Optional attention gates
    - Binary or multi-class segmentation
    
    Example:
        >>> model = UNet(encoder_name="resnet34", out_channels=1)
        >>> x = torch.randn(1, 3, 256, 256)
        >>> output = model(x)
        >>> output.shape
        torch.Size([1, 1, 256, 256])
    """
    
    def __init__(
        self,
        encoder_name: str = "resnet34",
        pretrained: bool = True,
        in_channels: int = 3,
        out_channels: int = 1,
        decoder_channels: List[int] = None,
        use_batch_norm: bool = True,
        attention: bool = False,
        dropout: float = 0.0,
    ):
        """
        Initialize U-Net model.
        
        Args:
            encoder_name: Name of encoder backbone 
                         ("resnet18", "resnet34", "resnet50", "efficientnet-b0")
            pretrained: Use pretrained encoder weights
            in_channels: Number of input channels
            out_channels: Number of output classes
            decoder_channels: Channel sizes for decoder blocks
            use_batch_norm: Use batch normalization in decoder
            attention: Use attention gates in skip connections
            dropout: Dropout rate before final conv
        """
        super().__init__()
        
        self.encoder_name = encoder_name
        self.in_channels = in_channels
        self.out_channels = out_channels
        
        # Create encoder
        if encoder_name.startswith("efficientnet"):
            self.encoder = EfficientNetEncoder(
                pretrained=pretrained,
                in_channels=in_channels,
            )
        else:
            self.encoder = ResNetEncoder(
                encoder_name=encoder_name,
                pretrained=pretrained,
                in_channels=in_channels,
            )
        
        encoder_channels = self.encoder.out_channels
        
        # Default decoder channels
        if decoder_channels is None:
            decoder_channels = [256, 128, 64, 32, 16]
        
        # Adjust decoder depth to match encoder
        decoder_channels = decoder_channels[:len(encoder_channels) - 1]
        
        # Create decoder
        self.decoder = UNetDecoder(
            encoder_channels=encoder_channels,
            decoder_channels=decoder_channels,
            use_batch_norm=use_batch_norm,
            attention=attention,
        )
        
        # Dropout
        self.dropout = nn.Dropout2d(dropout) if dropout > 0 else nn.Identity()
        
        # Final upsampling to input resolution
        self.final_upsample = nn.Sequential(
            nn.ConvTranspose2d(decoder_channels[-1], decoder_channels[-1], kernel_size=2, stride=2),
            DoubleConv(decoder_channels[-1], decoder_channels[-1], use_batch_norm=use_batch_norm),
        )
        
        # Segmentation head
        self.segmentation_head = nn.Conv2d(
            decoder_channels[-1],
            out_channels,
            kernel_size=1,
        )
        
        # Store configuration
        self._config = {
            "encoder_name": encoder_name,
            "pretrained": pretrained,
            "in_channels": in_channels,
            "out_channels": out_channels,
            "decoder_channels": decoder_channels,
            "use_batch_norm": use_batch_norm,
            "attention": attention,
            "dropout": dropout,
        }
    
    def forward(self, x: torch.Tensor) -> torch.Tensor:
        """
        Forward pass.
        
        Args:
            x: Input tensor of shape (B, C, H, W)
            
        Returns:
            Segmentation logits of shape (B, out_channels, H, W)
        """
        input_size = x.shape[2:]
        
        # Encoder
        features = self.encoder(x)
        
        # Decoder
        x = self.decoder(features)
        
        # Final upsampling
        x = self.final_upsample(x)
        
        # Ensure output matches input size
        if x.shape[2:] != input_size:
            x = F.interpolate(x, size=input_size, mode="bilinear", align_corners=True)
        
        # Dropout and segmentation head
        x = self.dropout(x)
        x = self.segmentation_head(x)
        
        return x
    
    def freeze_encoder(self) -> None:
        """Freeze encoder weights for transfer learning."""
        for param in self.encoder.parameters():
            param.requires_grad = False
        logger.info(f"Encoder {self.encoder_name} frozen")
    
    def unfreeze_encoder(self) -> None:
        """Unfreeze encoder weights."""
        for param in self.encoder.parameters():
            param.requires_grad = True
        logger.info(f"Encoder {self.encoder_name} unfrozen")
    
    def get_config(self) -> Dict[str, Any]:
        """Get model configuration dictionary."""
        return self._config.copy()


class UNetFactory:
    """
    Factory class for creating U-Net segmentation models.
    
    Provides a standardized interface for creating U-Net models with
    various configurations and determinism controls for blockchain consensus.
    
    Example:
        >>> model = UNetFactory.create_unet(
        ...     encoder_name="resnet34",
        ...     pretrained=True,
        ...     out_channels=4,
        ... )
    """
    
    SUPPORTED_ENCODERS = list(ENCODER_CONFIGS.keys())
    
    @staticmethod
    def create_unet(
        encoder_name: str = "resnet34",
        pretrained: bool = True,
        in_channels: int = 3,
        out_channels: int = 1,
        decoder_channels: Optional[List[int]] = None,
        use_batch_norm: bool = True,
        attention: bool = False,
        dropout: float = 0.0,
        deterministic: bool = True,
        seed: int = 42,
        **kwargs,
    ) -> nn.Module:
        """
        Create a U-Net model with specified configuration.
        
        Args:
            encoder_name: Encoder backbone name. Supported:
                - "resnet18": ResNet-18 encoder
                - "resnet34": ResNet-34 encoder (default, matches RMIT model)
                - "resnet50": ResNet-50 encoder
                - "efficientnet-b0": EfficientNet-B0 encoder
            pretrained: Use ImageNet pretrained encoder weights
            in_channels: Number of input channels (default: 3 for RGB)
            out_channels: Number of output classes (1 for binary, N for multi-class)
            decoder_channels: List of channel sizes for decoder blocks
                             Default: [256, 128, 64, 32, 16]
            use_batch_norm: Use batch normalization in decoder
            attention: Use attention gates in skip connections
            dropout: Dropout rate before final convolution
            deterministic: Enable deterministic mode for consensus
            seed: Random seed for initialization
            **kwargs: Additional arguments (ignored for compatibility)
            
        Returns:
            Configured UNet model
            
        Raises:
            ValueError: If encoder_name is not supported
        """
        if encoder_name not in UNetFactory.SUPPORTED_ENCODERS:
            raise ValueError(
                f"Unsupported encoder: {encoder_name}. "
                f"Supported encoders: {UNetFactory.SUPPORTED_ENCODERS}"
            )
        
        if deterministic:
            UNetFactory._set_deterministic_mode(seed)
        
        if decoder_channels is None:
            decoder_channels = [256, 128, 64, 32, 16]
        
        logger.info(
            f"Creating U-Net with encoder={encoder_name}, "
            f"pretrained={pretrained}, out_channels={out_channels}"
        )
        
        model = UNet(
            encoder_name=encoder_name,
            pretrained=pretrained,
            in_channels=in_channels,
            out_channels=out_channels,
            decoder_channels=decoder_channels,
            use_batch_norm=use_batch_norm,
            attention=attention,
            dropout=dropout,
        )
        
        return model
    
    @staticmethod
    def create_unet_for_face_extraction(
        pretrained: bool = True,
        out_channels: int = 4,
        **kwargs,
    ) -> nn.Module:
        """
        Create U-Net configured for face extraction from documents.
        
        Uses ResNet34 backbone to match RMIT U-Net weights.
        Output channels: 4 (background + 3 document fields)
        
        Args:
            pretrained: Use pretrained encoder
            out_channels: Number of output classes (default: 4)
            **kwargs: Additional arguments passed to create_unet
            
        Returns:
            UNet model configured for face extraction
        """
        return UNetFactory.create_unet(
            encoder_name="resnet34",
            pretrained=pretrained,
            in_channels=3,
            out_channels=out_channels,
            decoder_channels=[256, 128, 64, 32, 16],
            use_batch_norm=True,
            attention=False,
            **kwargs,
        )
    
    @staticmethod
    def _set_deterministic_mode(seed: int = 42) -> None:
        """
        Set deterministic mode for reproducibility.
        
        CRITICAL: Required for blockchain consensus - all validators
        must produce identical results.
        """
        import random
        
        random.seed(seed)
        os.environ["PYTHONHASHSEED"] = str(seed)
        
        torch.manual_seed(seed)
        torch.cuda.manual_seed_all(seed)
        torch.backends.cudnn.deterministic = True
        torch.backends.cudnn.benchmark = False
        
        # Enable deterministic algorithms
        try:
            torch.use_deterministic_algorithms(True)
        except Exception:
            pass  # May not be available in older PyTorch versions
    
    @staticmethod
    def get_encoder_info(encoder_name: str) -> Dict[str, Any]:
        """
        Get information about an encoder.
        
        Args:
            encoder_name: Name of the encoder
            
        Returns:
            Dictionary with encoder configuration
        """
        if encoder_name not in ENCODER_CONFIGS:
            raise ValueError(f"Unknown encoder: {encoder_name}")
        
        config = ENCODER_CONFIGS[encoder_name].copy()
        # Remove non-serializable items
        config.pop("model_fn", None)
        config.pop("weights", None)
        config["name"] = encoder_name
        return config
    
    @staticmethod
    def list_encoders() -> List[str]:
        """List all available encoder backbones."""
        return list(ENCODER_CONFIGS.keys())


def count_parameters(model: nn.Module, trainable_only: bool = False) -> int:
    """
    Count model parameters.
    
    Args:
        model: PyTorch model
        trainable_only: Only count trainable parameters
        
    Returns:
        Number of parameters
    """
    if trainable_only:
        return sum(p.numel() for p in model.parameters() if p.requires_grad)
    return sum(p.numel() for p in model.parameters())
