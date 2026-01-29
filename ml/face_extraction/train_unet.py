"""
U-Net Training Script for Face Extraction.

This module provides training utilities for U-Net face segmentation models
using PyTorch Lightning. Supports configurable loss functions, learning rate
scheduling, and integration with the model registry.

Task Reference: VE-3044 - Create U-Net Factory and Training Script
"""

import argparse
import hashlib
import logging
import os
from dataclasses import dataclass, field
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Callable, Dict, List, Optional, Tuple, Union

import numpy as np
import torch
import torch.nn as nn
import torch.nn.functional as F
from torch.optim import Adam, AdamW, SGD
from torch.optim.lr_scheduler import (
    CosineAnnealingLR,
    OneCycleLR,
    ReduceLROnPlateau,
    StepLR,
)
from torch.utils.data import DataLoader, Dataset

try:
    import pytorch_lightning as pl
    from pytorch_lightning.callbacks import (
        EarlyStopping,
        LearningRateMonitor,
        ModelCheckpoint,
    )
    from pytorch_lightning.loggers import TensorBoardLogger

    HAS_LIGHTNING = True
except ImportError:
    HAS_LIGHTNING = False
    pl = None

from ml.face_extraction.model_registry import (
    MODEL_REGISTRY,
    calculate_file_hash,
    ensure_determinism,
    get_weights_dir,
)
from ml.face_extraction.unet_factory import UNetFactory, count_parameters

logger = logging.getLogger(__name__)

# Set deterministic environment
os.environ["CUBLAS_WORKSPACE_CONFIG"] = ":4096:8"

# Metric key constants
METRIC_TRAIN_LOSS = "train/loss"
METRIC_TRAIN_IOU = "train/iou"
METRIC_TRAIN_DICE = "train/dice"
METRIC_VAL_LOSS = "val/loss"
METRIC_VAL_IOU = "val/iou"
METRIC_VAL_DICE = "val/dice"


# ============================================================================
# Loss Functions
# ============================================================================


class DiceLoss(nn.Module):
    """Dice loss for segmentation."""

    def __init__(self, smooth: float = 1.0, multiclass: bool = False):
        super().__init__()
        self.smooth = smooth
        self.multiclass = multiclass

    def forward(
        self,
        pred: torch.Tensor,
        target: torch.Tensor,
    ) -> torch.Tensor:
        """
        Compute Dice loss.

        Args:
            pred: Predictions of shape (B, C, H, W) - raw logits
            target: Targets of shape (B, H, W) for multiclass or (B, 1, H, W) for binary

        Returns:
            Dice loss value
        """
        if self.multiclass:
            pred = F.softmax(pred, dim=1)
            num_classes = pred.shape[1]

            # Convert target to one-hot
            target_one_hot = F.one_hot(target.long(), num_classes)
            target_one_hot = target_one_hot.permute(0, 3, 1, 2).float()

            # Compute per-class dice
            intersection = (pred * target_one_hot).sum(dim=(2, 3))
            union = pred.sum(dim=(2, 3)) + target_one_hot.sum(dim=(2, 3))

            dice = (2.0 * intersection + self.smooth) / (union + self.smooth)
            return 1.0 - dice.mean()
        else:
            pred = torch.sigmoid(pred)
            pred = pred.view(-1)
            target = target.view(-1)

            intersection = (pred * target).sum()
            union = pred.sum() + target.sum()

            dice = (2.0 * intersection + self.smooth) / (union + self.smooth)
            return 1.0 - dice


class FocalLoss(nn.Module):
    """Focal loss for handling class imbalance."""

    def __init__(
        self,
        alpha: float = 0.25,
        gamma: float = 2.0,
        multiclass: bool = False,
    ):
        super().__init__()
        self.alpha = alpha
        self.gamma = gamma
        self.multiclass = multiclass

    def forward(
        self,
        pred: torch.Tensor,
        target: torch.Tensor,
    ) -> torch.Tensor:
        if self.multiclass:
            ce_loss = F.cross_entropy(pred, target.long(), reduction="none")
            pt = torch.exp(-ce_loss)
            focal_loss = self.alpha * (1 - pt) ** self.gamma * ce_loss
            return focal_loss.mean()
        else:
            bce_loss = F.binary_cross_entropy_with_logits(
                pred, target.float(), reduction="none"
            )
            pt = torch.exp(-bce_loss)
            focal_loss = self.alpha * (1 - pt) ** self.gamma * bce_loss
            return focal_loss.mean()


class CombinedLoss(nn.Module):
    """Combined loss function (BCE/CE + Dice)."""

    def __init__(
        self,
        bce_weight: float = 0.5,
        dice_weight: float = 0.5,
        multiclass: bool = False,
    ):
        super().__init__()
        self.bce_weight = bce_weight
        self.dice_weight = dice_weight
        self.multiclass = multiclass
        self.dice_loss = DiceLoss(multiclass=multiclass)

    def forward(
        self,
        pred: torch.Tensor,
        target: torch.Tensor,
    ) -> torch.Tensor:
        if self.multiclass:
            ce_loss = F.cross_entropy(pred, target.long())
        else:
            ce_loss = F.binary_cross_entropy_with_logits(pred, target.float())

        dice_loss = self.dice_loss(pred, target)
        return self.bce_weight * ce_loss + self.dice_weight * dice_loss


def get_loss_function(
    loss_name: str,
    multiclass: bool = False,
    **kwargs,
) -> nn.Module:
    """
    Get loss function by name.

    Args:
        loss_name: Name of loss function
            - "bce": Binary Cross Entropy (binary only)
            - "ce": Cross Entropy (multiclass only)
            - "dice": Dice loss
            - "focal": Focal loss
            - "bce_dice" / "combined": BCE/CE + Dice combined
        multiclass: Whether this is multi-class segmentation
        **kwargs: Additional arguments for loss function

    Returns:
        Loss function module
    """
    loss_name = loss_name.lower().replace("-", "_")

    if loss_name == "bce":
        if multiclass:
            return nn.CrossEntropyLoss()
        return nn.BCEWithLogitsLoss()

    elif loss_name == "ce":
        return nn.CrossEntropyLoss()

    elif loss_name == "dice":
        return DiceLoss(multiclass=multiclass)

    elif loss_name == "focal":
        return FocalLoss(multiclass=multiclass, **kwargs)

    elif loss_name in ("bce_dice", "combined"):
        return CombinedLoss(multiclass=multiclass, **kwargs)

    else:
        raise ValueError(f"Unknown loss function: {loss_name}")


# ============================================================================
# Metrics
# ============================================================================


def compute_iou(
    pred: torch.Tensor,
    target: torch.Tensor,
    threshold: float = 0.5,
    smooth: float = 1e-6,
) -> torch.Tensor:
    """
    Compute Intersection over Union (IoU / Jaccard index).

    Args:
        pred: Predictions (logits)
        target: Ground truth
        threshold: Threshold for binarizing predictions
        smooth: Smoothing factor to avoid division by zero

    Returns:
        IoU score
    """
    if pred.shape[1] > 1:  # Multiclass
        pred = torch.argmax(pred, dim=1)
    else:
        pred = (torch.sigmoid(pred) > threshold).float().squeeze(1)

    target = target.float()
    pred = pred.float()

    intersection = (pred * target).sum()
    union = pred.sum() + target.sum() - intersection

    return (intersection + smooth) / (union + smooth)


def compute_dice(
    pred: torch.Tensor,
    target: torch.Tensor,
    threshold: float = 0.5,
    smooth: float = 1.0,
) -> torch.Tensor:
    """
    Compute Dice coefficient.

    Args:
        pred: Predictions (logits)
        target: Ground truth
        threshold: Threshold for binarizing predictions
        smooth: Smoothing factor

    Returns:
        Dice score
    """
    if pred.shape[1] > 1:  # Multiclass
        pred = torch.argmax(pred, dim=1)
    else:
        pred = (torch.sigmoid(pred) > threshold).float().squeeze(1)

    target = target.float()
    pred = pred.float()

    intersection = (pred * target).sum()
    union = pred.sum() + target.sum()

    return (2.0 * intersection + smooth) / (union + smooth)


# ============================================================================
# Training Configuration
# ============================================================================


@dataclass
class TrainingConfig:
    """Configuration for U-Net training."""

    # Model settings
    encoder_name: str = "resnet34"
    pretrained: bool = True
    in_channels: int = 3
    out_channels: int = 1
    decoder_channels: List[int] = field(
        default_factory=lambda: [256, 128, 64, 32, 16]
    )

    # Training settings
    batch_size: int = 8
    num_epochs: int = 100
    learning_rate: float = 1e-4
    weight_decay: float = 1e-5
    optimizer: str = "adamw"  # "adam", "adamw", "sgd"
    scheduler: str = "cosine"  # "cosine", "step", "plateau", "onecycle"

    # Loss function
    loss_function: str = "bce_dice"  # "bce", "dice", "focal", "bce_dice"

    # Early stopping
    early_stopping: bool = True
    patience: int = 10
    min_delta: float = 1e-4

    # Checkpointing
    save_top_k: int = 3
    checkpoint_dir: str = "checkpoints"

    # Data
    train_data_dir: Optional[str] = None
    val_data_dir: Optional[str] = None
    num_workers: int = 4
    input_size: Tuple[int, int] = (512, 512)

    # Augmentation
    augment: bool = True

    # Determinism
    deterministic: bool = True
    seed: int = 42

    # Logging
    log_dir: str = "logs"
    experiment_name: str = "unet_training"

    # Hardware
    accelerator: str = "auto"  # "cpu", "gpu", "auto"
    devices: int = 1

    def __post_init__(self):
        """Validate configuration."""
        if self.out_channels > 1:
            self.multiclass = True
        else:
            self.multiclass = False


# ============================================================================
# PyTorch Lightning Module
# ============================================================================

if HAS_LIGHTNING:

    class UNetLightningModule(pl.LightningModule):
        """PyTorch Lightning module for U-Net training."""

        def __init__(self, config: TrainingConfig):
            super().__init__()
            self.config = config
            self.save_hyperparameters()

            # Create model
            self.model = UNetFactory.create_unet(
                encoder_name=config.encoder_name,
                pretrained=config.pretrained,
                in_channels=config.in_channels,
                out_channels=config.out_channels,
                decoder_channels=config.decoder_channels,
                deterministic=config.deterministic,
                seed=config.seed,
            )

            # Loss function
            self.loss_fn = get_loss_function(
                config.loss_function,
                multiclass=config.multiclass,
            )

            # Metrics tracking
            self.training_step_outputs = []
            self.validation_step_outputs = []

        def forward(self, x: torch.Tensor) -> torch.Tensor:
            return self.model(x)

        def training_step(
            self,
            batch: Tuple[torch.Tensor, torch.Tensor],
            batch_idx: int,
        ) -> torch.Tensor:
            images, masks = batch
            logits = self(images)

            loss = self.loss_fn(logits, masks)

            # Compute metrics
            iou = compute_iou(logits.detach(), masks)
            dice = compute_dice(logits.detach(), masks)

            # Log
            self.log(METRIC_TRAIN_LOSS, loss, prog_bar=True)
            self.log(METRIC_TRAIN_IOU, iou, prog_bar=True)
            self.log(METRIC_TRAIN_DICE, dice)

            self.training_step_outputs.append(
                {"loss": loss.detach(), "iou": iou, "dice": dice}
            )

            return loss

        def on_train_epoch_end(self) -> None:
            if not self.training_step_outputs:
                return

            avg_loss = torch.stack(
                [x["loss"] for x in self.training_step_outputs]
            ).mean()
            avg_iou = torch.stack(
                [x["iou"] for x in self.training_step_outputs]
            ).mean()
            avg_dice = torch.stack(
                [x["dice"] for x in self.training_step_outputs]
            ).mean()

            self.log(f\"{METRIC_TRAIN_LOSS.replace('/', '/epoch_')}\", avg_loss)
            self.log(f\"{METRIC_TRAIN_IOU.replace('/', '/epoch_')}\", avg_iou)
            self.log(f\"{METRIC_TRAIN_DICE.replace('/', '/epoch_')}\", avg_dice)

            self.training_step_outputs.clear()

        def validation_step(
            self,
            batch: Tuple[torch.Tensor, torch.Tensor],
            batch_idx: int,
        ) -> None:
            images, masks = batch
            logits = self(images)

            loss = self.loss_fn(logits, masks)
            iou = compute_iou(logits, masks)
            dice = compute_dice(logits, masks)

            self.log(METRIC_VAL_LOSS, loss, prog_bar=True)
            self.log(METRIC_VAL_IOU, iou, prog_bar=True)
            self.log(METRIC_VAL_DICE, dice)

            self.validation_step_outputs.append(
                {"loss": loss, "iou": iou, "dice": dice}
            )

        def on_validation_epoch_end(self) -> None:
            if not self.validation_step_outputs:
                return

            avg_loss = torch.stack(
                [x["loss"] for x in self.validation_step_outputs]
            ).mean()
            avg_iou = torch.stack(
                [x["iou"] for x in self.validation_step_outputs]
            ).mean()
            avg_dice = torch.stack(
                [x["dice"] for x in self.validation_step_outputs]
            ).mean()

            self.log(f\"{METRIC_VAL_LOSS.replace('/', '/epoch_')}\", avg_loss)
            self.log(f\"{METRIC_VAL_IOU.replace('/', '/epoch_')}\", avg_iou)
            self.log(f\"{METRIC_VAL_DICE.replace('/', '/epoch_')}\", avg_dice)

            self.validation_step_outputs.clear()

        def configure_optimizers(self) -> Dict[str, Any]:
            # Optimizer
            if self.config.optimizer.lower() == "adam":
                optimizer = Adam(
                    self.parameters(),
                    lr=self.config.learning_rate,
                    weight_decay=self.config.weight_decay,
                )
            elif self.config.optimizer.lower() == "adamw":
                optimizer = AdamW(
                    self.parameters(),
                    lr=self.config.learning_rate,
                    weight_decay=self.config.weight_decay,
                )
            elif self.config.optimizer.lower() == "sgd":
                optimizer = SGD(
                    self.parameters(),
                    lr=self.config.learning_rate,
                    weight_decay=self.config.weight_decay,
                    momentum=0.9,
                )
            else:
                raise ValueError(f"Unknown optimizer: {self.config.optimizer}")

            # Scheduler
            if self.config.scheduler.lower() == "cosine":
                scheduler = CosineAnnealingLR(
                    optimizer,
                    T_max=self.config.num_epochs,
                    eta_min=1e-7,
                )
                return {"optimizer": optimizer, "lr_scheduler": scheduler}

            elif self.config.scheduler.lower() == "step":
                scheduler = StepLR(
                    optimizer,
                    step_size=30,
                    gamma=0.1,
                )
                return {"optimizer": optimizer, "lr_scheduler": scheduler}

            elif self.config.scheduler.lower() == "plateau":
                scheduler = ReduceLROnPlateau(
                    optimizer,
                    mode="min",
                    factor=0.5,
                    patience=5,
                    min_lr=1e-7,
                )
                return {
                    "optimizer": optimizer,
                    "lr_scheduler": {
                        "scheduler": scheduler,
                        "monitor": METRIC_VAL_LOSS,
                        "interval": "epoch",
                    },
                }

            elif self.config.scheduler.lower() == "onecycle":
                # OneCycle requires total steps
                scheduler = OneCycleLR(
                    optimizer,
                    max_lr=self.config.learning_rate * 10,
                    total_steps=self.trainer.estimated_stepping_batches,
                    pct_start=0.3,
                )
                return {
                    "optimizer": optimizer,
                    "lr_scheduler": {
                        "scheduler": scheduler,
                        "interval": "step",
                    },
                }

            return {"optimizer": optimizer}


# ============================================================================
# Standard Training Loop (fallback if Lightning not available)
# ============================================================================


class StandardTrainer:
    """
    Standard PyTorch training loop.

    Used when PyTorch Lightning is not available.
    """

    def __init__(
        self,
        model: nn.Module,
        config: TrainingConfig,
        train_loader: DataLoader,
        val_loader: Optional[DataLoader] = None,
    ):
        self.model = model
        self.config = config
        self.train_loader = train_loader
        self.val_loader = val_loader

        # Set device
        self.device = torch.device(
            "cuda" if torch.cuda.is_available() and config.accelerator != "cpu" else "cpu"
        )
        self.model.to(self.device)

        # Loss function
        self.loss_fn = get_loss_function(
            config.loss_function,
            multiclass=config.multiclass,
        )

        # Optimizer
        if config.optimizer.lower() == "adamw":
            self.optimizer = AdamW(
                model.parameters(),
                lr=config.learning_rate,
                weight_decay=config.weight_decay,
            )
        else:
            self.optimizer = Adam(
                model.parameters(),
                lr=config.learning_rate,
                weight_decay=config.weight_decay,
            )

        # Scheduler
        if config.scheduler.lower() == "cosine":
            self.scheduler = CosineAnnealingLR(
                self.optimizer,
                T_max=config.num_epochs,
            )
        else:
            self.scheduler = StepLR(self.optimizer, step_size=30, gamma=0.1)

        # Tracking
        self.best_val_loss = float("inf")
        self.patience_counter = 0
        self.history = {"train_loss": [], "val_loss": [], "train_iou": [], "val_iou": []}

        # Checkpoints
        self.checkpoint_dir = Path(config.checkpoint_dir)
        self.checkpoint_dir.mkdir(parents=True, exist_ok=True)

    def train_epoch(self) -> Dict[str, float]:
        """Train for one epoch."""
        self.model.train()
        total_loss = 0.0
        total_iou = 0.0
        num_batches = 0

        for images, masks in self.train_loader:
            images = images.to(self.device)
            masks = masks.to(self.device)

            self.optimizer.zero_grad()
            logits = self.model(images)
            loss = self.loss_fn(logits, masks)
            loss.backward()
            self.optimizer.step()

            total_loss += loss.item()
            total_iou += compute_iou(logits.detach(), masks).item()
            num_batches += 1

        return {
            "loss": total_loss / num_batches,
            "iou": total_iou / num_batches,
        }

    @torch.no_grad()
    def validate(self) -> Dict[str, float]:
        """Run validation."""
        if self.val_loader is None:
            return {}

        self.model.eval()
        total_loss = 0.0
        total_iou = 0.0
        total_dice = 0.0
        num_batches = 0

        for images, masks in self.val_loader:
            images = images.to(self.device)
            masks = masks.to(self.device)

            logits = self.model(images)
            loss = self.loss_fn(logits, masks)

            total_loss += loss.item()
            total_iou += compute_iou(logits, masks).item()
            total_dice += compute_dice(logits, masks).item()
            num_batches += 1

        return {
            "loss": total_loss / num_batches,
            "iou": total_iou / num_batches,
            "dice": total_dice / num_batches,
        }

    def save_checkpoint(
        self,
        epoch: int,
        metrics: Dict[str, float],
        is_best: bool = False,
    ) -> str:
        """Save model checkpoint."""
        checkpoint = {
            "epoch": epoch,
            "model_state_dict": self.model.state_dict(),
            "optimizer_state_dict": self.optimizer.state_dict(),
            "scheduler_state_dict": self.scheduler.state_dict(),
            "metrics": metrics,
            "config": self.config.__dict__,
        }

        # Regular checkpoint
        filename = f"checkpoint_epoch_{epoch:03d}.pt"
        filepath = self.checkpoint_dir / filename
        torch.save(checkpoint, filepath)

        # Best checkpoint
        if is_best:
            best_path = self.checkpoint_dir / "best_model.pt"
            torch.save(checkpoint, best_path)
            logger.info(f"Saved best model to {best_path}")

        return str(filepath)

    def train(self) -> Dict[str, List[float]]:
        """Run full training loop."""
        logger.info(f"Starting training on {self.device}")
        logger.info(f"Model parameters: {count_parameters(self.model):,}")

        for epoch in range(self.config.num_epochs):
            # Train
            train_metrics = self.train_epoch()
            self.history["train_loss"].append(train_metrics["loss"])
            self.history["train_iou"].append(train_metrics["iou"])

            # Validate
            val_metrics = self.validate()
            if val_metrics:
                self.history["val_loss"].append(val_metrics["loss"])
                self.history["val_iou"].append(val_metrics["iou"])

                # Check for improvement
                is_best = val_metrics["loss"] < self.best_val_loss
                if is_best:
                    self.best_val_loss = val_metrics["loss"]
                    self.patience_counter = 0
                else:
                    self.patience_counter += 1

                # Early stopping
                if (
                    self.config.early_stopping
                    and self.patience_counter >= self.config.patience
                ):
                    logger.info(f"Early stopping at epoch {epoch}")
                    break

                # Save checkpoint
                self.save_checkpoint(epoch, val_metrics, is_best)

            # Update scheduler
            self.scheduler.step()

            # Log progress
            log_msg = (
                f"Epoch {epoch + 1}/{self.config.num_epochs} - "
                f"Train Loss: {train_metrics['loss']:.4f}, "
                f"Train IoU: {train_metrics['iou']:.4f}"
            )
            if val_metrics:
                log_msg += (
                    f", Val Loss: {val_metrics['loss']:.4f}, "
                    f"Val IoU: {val_metrics['iou']:.4f}"
                )
            logger.info(log_msg)

        return self.history


# ============================================================================
# Registry Integration
# ============================================================================


def register_trained_model(
    model_path: Path,
    model_name: str,
    config: TrainingConfig,
    metrics: Dict[str, float],
) -> Dict[str, Any]:
    """
    Register a trained model in the model registry.

    Args:
        model_path: Path to saved model weights
        model_name: Name for the model in registry
        config: Training configuration
        metrics: Final validation metrics

    Returns:
        Registry entry for the model
    """
    # Calculate hash
    sha256_hash = calculate_file_hash(model_path)

    # Create registry entry
    entry = {
        "filename": model_path.name,
        "sha256": sha256_hash,
        "source": "VirtEngine U-Net training",
        "description": f"U-Net {config.encoder_name} trained for face extraction",
        "input_size": config.input_size,
        "output_classes": config.out_channels,
        "trained_on": "VirtEngine face extraction dataset",
        "framework": "pytorch",
        "architecture": f"UNet with {config.encoder_name} encoder",
        "training_config": {
            "encoder": config.encoder_name,
            "pretrained": config.pretrained,
            "decoder_channels": config.decoder_channels,
            "loss_function": config.loss_function,
            "optimizer": config.optimizer,
            "learning_rate": config.learning_rate,
            "epochs_trained": config.num_epochs,
        },
        "metrics": metrics,
        "created_at": datetime.now(timezone.utc).isoformat(),
    }

    # Add to registry (in memory - would need to persist to file)
    MODEL_REGISTRY[model_name] = entry

    logger.info(f"Registered model {model_name} with hash {sha256_hash[:16]}...")

    return entry


# ============================================================================
# Training Entry Points
# ============================================================================


def train_with_lightning(
    config: TrainingConfig,
    train_loader: DataLoader,
    val_loader: Optional[DataLoader] = None,
) -> Tuple[nn.Module, Dict[str, Any]]:
    """
    Train U-Net using PyTorch Lightning.

    Args:
        config: Training configuration
        train_loader: Training data loader
        val_loader: Validation data loader

    Returns:
        Tuple of (trained model, training results)
    """
    if not HAS_LIGHTNING:
        raise RuntimeError(
            "PyTorch Lightning not available. "
            "Install with: pip install pytorch-lightning"
        )

    # Set deterministic mode
    if config.deterministic:
        ensure_determinism(config.seed)
        pl.seed_everything(config.seed, workers=True)

    # Create module
    module = UNetLightningModule(config)

    # Callbacks
    callbacks = [
        LearningRateMonitor(logging_interval="step"),
    ]

    if val_loader is not None:
        callbacks.append(
            ModelCheckpoint(
                dirpath=config.checkpoint_dir,
                filename="unet-{epoch:02d}-{val/loss:.4f}",
                monitor=METRIC_VAL_LOSS,
                mode="min",
                save_top_k=config.save_top_k,
                save_last=True,
            )
        )

        if config.early_stopping:
            callbacks.append(
                EarlyStopping(
                    monitor=METRIC_VAL_LOSS,
                    patience=config.patience,
                    min_delta=config.min_delta,
                    mode="min",
                )
            )

    # Logger
    tb_logger = TensorBoardLogger(
        save_dir=config.log_dir,
        name=config.experiment_name,
    )

    # Trainer
    trainer = pl.Trainer(
        max_epochs=config.num_epochs,
        accelerator=config.accelerator,
        devices=config.devices,
        callbacks=callbacks,
        logger=tb_logger,
        deterministic=config.deterministic,
        log_every_n_steps=10,
    )

    # Train
    trainer.fit(module, train_loader, val_loader)

    # Return trained model
    return module.model, {"trainer": trainer, "module": module}


def train_standard(
    config: TrainingConfig,
    train_loader: DataLoader,
    val_loader: Optional[DataLoader] = None,
) -> Tuple[nn.Module, Dict[str, List[float]]]:
    """
    Train U-Net using standard PyTorch training loop.

    Args:
        config: Training configuration
        train_loader: Training data loader
        val_loader: Validation data loader

    Returns:
        Tuple of (trained model, training history)
    """
    # Set deterministic mode
    if config.deterministic:
        ensure_determinism(config.seed)

    # Create model
    model = UNetFactory.create_unet(
        encoder_name=config.encoder_name,
        pretrained=config.pretrained,
        in_channels=config.in_channels,
        out_channels=config.out_channels,
        decoder_channels=config.decoder_channels,
        deterministic=config.deterministic,
        seed=config.seed,
    )

    # Create trainer and train
    trainer = StandardTrainer(model, config, train_loader, val_loader)
    history = trainer.train()

    return model, history


# ============================================================================
# CLI Entry Point
# ============================================================================


def create_dummy_dataset(
    num_samples: int = 100,
    input_size: Tuple[int, int] = (256, 256),
    num_classes: int = 1,
) -> Dataset:
    """Create dummy dataset for testing."""

    class DummyDataset(Dataset):
        def __init__(self):
            self.num_samples = num_samples
            self.input_size = input_size
            self.num_classes = num_classes

        def __len__(self):
            return self.num_samples

        def __getitem__(self, idx):
            image = torch.randn(3, *self.input_size)
            if self.num_classes > 1:
                mask = torch.randint(0, self.num_classes, self.input_size)
            else:
                mask = torch.randint(0, 2, (1, *self.input_size)).float()
            return image, mask

    return DummyDataset()


def main():
    """Main entry point for training."""
    parser = argparse.ArgumentParser(description="Train U-Net for face extraction")

    # Model arguments
    parser.add_argument(
        "--encoder",
        type=str,
        default="resnet34",
        choices=["resnet18", "resnet34", "resnet50", "efficientnet-b0"],
        help="Encoder backbone",
    )
    parser.add_argument(
        "--pretrained",
        action="store_true",
        default=True,
        help="Use pretrained encoder",
    )
    parser.add_argument(
        "--out-channels",
        type=int,
        default=4,
        help="Number of output classes",
    )

    # Training arguments
    parser.add_argument("--epochs", type=int, default=100, help="Number of epochs")
    parser.add_argument("--batch-size", type=int, default=8, help="Batch size")
    parser.add_argument("--lr", type=float, default=1e-4, help="Learning rate")
    parser.add_argument(
        "--loss",
        type=str,
        default="bce_dice",
        choices=["bce", "dice", "focal", "bce_dice"],
        help="Loss function",
    )
    parser.add_argument(
        "--optimizer",
        type=str,
        default="adamw",
        choices=["adam", "adamw", "sgd"],
        help="Optimizer",
    )

    # Data arguments
    parser.add_argument("--train-dir", type=str, help="Training data directory")
    parser.add_argument("--val-dir", type=str, help="Validation data directory")
    parser.add_argument(
        "--input-size",
        type=int,
        nargs=2,
        default=[512, 512],
        help="Input size (H W)",
    )

    # Other arguments
    parser.add_argument(
        "--checkpoint-dir",
        type=str,
        default="checkpoints",
        help="Checkpoint directory",
    )
    parser.add_argument("--seed", type=int, default=42, help="Random seed")
    parser.add_argument(
        "--use-lightning",
        action="store_true",
        help="Use PyTorch Lightning trainer",
    )
    parser.add_argument("--dummy", action="store_true", help="Use dummy data for testing")

    args = parser.parse_args()

    # Create configuration
    config = TrainingConfig(
        encoder_name=args.encoder,
        pretrained=args.pretrained,
        out_channels=args.out_channels,
        num_epochs=args.epochs,
        batch_size=args.batch_size,
        learning_rate=args.lr,
        loss_function=args.loss,
        optimizer=args.optimizer,
        train_data_dir=args.train_dir,
        val_data_dir=args.val_dir,
        input_size=tuple(args.input_size),
        checkpoint_dir=args.checkpoint_dir,
        seed=args.seed,
    )

    # Setup logging
    logging.basicConfig(level=logging.INFO)

    # Create data loaders
    if args.dummy:
        logger.info("Using dummy dataset for testing")
        train_dataset = create_dummy_dataset(
            num_samples=100,
            input_size=config.input_size,
            num_classes=config.out_channels,
        )
        val_dataset = create_dummy_dataset(
            num_samples=20,
            input_size=config.input_size,
            num_classes=config.out_channels,
        )
        train_loader = DataLoader(
            train_dataset, batch_size=config.batch_size, shuffle=True, num_workers=0
        )
        val_loader = DataLoader(
            val_dataset, batch_size=config.batch_size, num_workers=0
        )
    else:
        # Would load real data here
        raise NotImplementedError("Real data loading not implemented. Use --dummy flag.")

    # Train
    if args.use_lightning and HAS_LIGHTNING:
        trained_model, training_results = train_with_lightning(config, train_loader, val_loader)
        logger.info(f"Training complete! Model: {trained_model.encoder_name}")
    else:
        trained_model, training_results = train_standard(config, train_loader, val_loader)
        logger.info(f"Training complete! Final loss: {training_results.get('train_loss', ['N/A'])[-1]}")


if __name__ == "__main__":
    main()
