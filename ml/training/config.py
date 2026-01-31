"""
Training configuration for the trust score model.

This module defines all configuration parameters for:
- Dataset ingestion and splitting
- Feature extraction settings
- Model architecture and training hyperparameters
- Export settings for TensorFlow-Go inference
"""

import os
import yaml
import json
from dataclasses import dataclass, field, asdict
from typing import List, Dict, Optional, Tuple, Any
from enum import Enum
from pathlib import Path


class DocumentType(str, Enum):
    """Supported document types for training."""
    ID_CARD = "id_card"
    PASSPORT = "passport"
    DRIVERS_LICENSE = "drivers_license"
    RESIDENCE_PERMIT = "residence_permit"
    NATIONAL_ID = "national_id"


class AugmentationType(str, Enum):
    """Supported data augmentation types."""
    BRIGHTNESS = "brightness"
    CONTRAST = "contrast"
    ROTATION = "rotation"
    BLUR = "blur"
    NOISE = "noise"
    PERSPECTIVE = "perspective"
    JPEG_ARTIFACT = "jpeg_artifact"


class AnonymizationMethod(str, Enum):
    """PII anonymization methods."""
    HASH_SHA256 = "hash_sha256"
    HASH_BLAKE2 = "hash_blake2"
    REDACT = "redact"
    TOKENIZE = "tokenize"


@dataclass
class DatasetConfig:
    """Configuration for dataset ingestion."""
    
    # Data paths
    data_paths: List[str] = field(default_factory=list)
    
    # Document types to include
    doc_types: List[str] = field(default_factory=lambda: [
        DocumentType.ID_CARD.value,
        DocumentType.PASSPORT.value,
        DocumentType.DRIVERS_LICENSE.value,
    ])
    
    # Dataset splits
    train_split: float = 0.8
    val_split: float = 0.1
    test_split: float = 0.1
    
    # Anonymization settings
    anonymize: bool = True
    anonymization_method: str = AnonymizationMethod.HASH_SHA256.value
    anonymization_salt: Optional[str] = None  # Will be generated if not provided
    
    # Sampling settings
    max_samples: Optional[int] = None  # None = use all
    balance_classes: bool = True
    min_samples_per_class: int = 100
    
    # Quality filters
    min_face_confidence: float = 0.8
    min_doc_quality: float = 0.6
    min_ocr_confidence: float = 0.5
    
    # Random seed for reproducibility
    random_seed: int = 42
    
    # Cache settings
    cache_dir: Optional[str] = None
    use_cache: bool = True
    
    def __post_init__(self):
        """Validate configuration."""
        assert 0 < self.train_split <= 1.0, "train_split must be in (0, 1]"
        assert 0 <= self.val_split < 1.0, "val_split must be in [0, 1)"
        assert 0 <= self.test_split < 1.0, "test_split must be in [0, 1)"
        total = self.train_split + self.val_split + self.test_split
        assert abs(total - 1.0) < 1e-6, f"Splits must sum to 1.0, got {total}"


@dataclass
class PreprocessingConfig:
    """Configuration for data preprocessing."""
    
    # Image preprocessing
    image_size: Tuple[int, int] = (224, 224)
    normalize_images: bool = True
    normalize_mean: Tuple[float, float, float] = (0.485, 0.456, 0.406)
    normalize_std: Tuple[float, float, float] = (0.229, 0.224, 0.225)
    
    # Document preprocessing
    apply_orientation_correction: bool = True
    apply_perspective_correction: bool = True
    apply_enhancement: bool = True
    
    # OCR preprocessing
    ocr_language: str = "eng"
    ocr_timeout_seconds: float = 10.0


@dataclass
class AugmentationConfig:
    """Configuration for data augmentation."""
    
    # Enable augmentation
    enabled: bool = True
    
    # Augmentation types to apply
    augmentation_types: List[str] = field(default_factory=lambda: [
        AugmentationType.BRIGHTNESS.value,
        AugmentationType.CONTRAST.value,
        AugmentationType.ROTATION.value,
    ])
    
    # Augmentation parameters
    brightness_range: Tuple[float, float] = (0.8, 1.2)
    contrast_range: Tuple[float, float] = (0.8, 1.2)
    rotation_range: Tuple[float, float] = (-5.0, 5.0)  # Degrees
    blur_kernel_range: Tuple[int, int] = (3, 7)
    noise_std_range: Tuple[float, float] = (0.01, 0.05)
    perspective_strength: float = 0.1
    jpeg_quality_range: Tuple[int, int] = (70, 95)
    
    # Augmentation probability
    augmentation_probability: float = 0.5
    
    # Number of augmented copies per sample
    num_augmented_copies: int = 2


@dataclass
class FeatureConfig:
    """Configuration for feature extraction."""
    
    # Face embedding settings
    face_embedding_dim: int = 512
    face_embedding_model: str = "facenet"
    use_face_confidence: bool = True
    
    # Document feature settings
    doc_quality_features: List[str] = field(default_factory=lambda: [
        "sharpness",
        "brightness",
        "contrast",
        "noise_level",
        "blur_score",
    ])
    
    # OCR feature settings
    ocr_fields: List[str] = field(default_factory=lambda: [
        "name",
        "date_of_birth",
        "document_number",
        "expiry_date",
        "nationality",
    ])
    use_ocr_confidence: bool = True
    use_field_validation: bool = True
    
    # Metadata feature settings
    use_device_metadata: bool = True
    use_session_metadata: bool = True
    use_capture_metadata: bool = True
    
    # Combined feature vector size
    combined_feature_dim: int = 768
    
    # Feature normalization
    normalize_features: bool = True
    feature_scaling: str = "standard"  # "standard", "minmax", "robust"


@dataclass
class ModelConfig:
    """Configuration for model architecture and training."""
    
    # Architecture settings
    input_dim: int = 768
    hidden_layers: List[int] = field(default_factory=lambda: [512, 256, 128, 64])
    dropout_rate: float = 0.3
    activation: str = "relu"
    output_activation: str = "sigmoid"
    
    # L2 regularization
    l2_regularization: float = 0.01
    
    # Batch normalization
    use_batch_norm: bool = True
    
    # Output scaling (for 0-100 score)
    output_scale: float = 100.0
    
    # Training hyperparameters
    learning_rate: float = 0.001
    batch_size: int = 64
    epochs: int = 100
    
    # Optimizer settings
    optimizer: str = "adam"
    adam_beta_1: float = 0.9
    adam_beta_2: float = 0.999
    adam_epsilon: float = 1e-7
    
    # Learning rate schedule
    use_lr_schedule: bool = True
    lr_decay_factor: float = 0.1
    lr_decay_patience: int = 10
    min_lr: float = 1e-6
    
    # Early stopping
    early_stopping: bool = True
    early_stopping_patience: int = 15
    early_stopping_min_delta: float = 0.001
    
    # Checkpointing
    checkpoint_dir: str = "checkpoints"
    save_best_only: bool = True
    
    # Logging
    tensorboard_log_dir: str = "logs/tensorboard"
    log_every_n_steps: int = 100

    def get_config(self) -> Dict[str, Any]:
        """Return config dict for TensorFlow serialization."""
        return asdict(self)


@dataclass
class ExportConfig:
    """Configuration for model export."""
    
    # Export path
    export_dir: str = "exported_models"
    
    # Model versioning
    version_prefix: str = "v"
    include_timestamp: bool = True
    
    # SavedModel settings
    include_optimizer: bool = False
    signature_name: str = "serving_default"
    
    # Input/output specifications for Go inference
    input_name: str = "features"
    output_name: str = "trust_score"
    
    # Quantization (for smaller model size)
    quantize: bool = False
    quantization_type: str = "float16"  # "float16", "int8"
    
    # Hash computation for versioning
    compute_hash: bool = True
    hash_algorithm: str = "sha256"


@dataclass
class TrainingConfig:
    """Complete training configuration."""
    
    # Component configs
    dataset: DatasetConfig = field(default_factory=DatasetConfig)
    preprocessing: PreprocessingConfig = field(default_factory=PreprocessingConfig)
    augmentation: AugmentationConfig = field(default_factory=AugmentationConfig)
    features: FeatureConfig = field(default_factory=FeatureConfig)
    model: ModelConfig = field(default_factory=ModelConfig)
    export: ExportConfig = field(default_factory=ExportConfig)
    
    # Global settings
    experiment_name: str = "trust_score_v1"
    random_seed: int = 42
    deterministic: bool = True
    
    # Logging
    log_level: str = "INFO"
    
    @classmethod
    def from_yaml(cls, yaml_path: str) -> "TrainingConfig":
        """Load configuration from YAML file."""
        with open(yaml_path, 'r') as f:
            config_dict = yaml.safe_load(f)
        return cls.from_dict(config_dict)
    
    @classmethod
    def from_json(cls, json_path: str) -> "TrainingConfig":
        """Load configuration from JSON file."""
        with open(json_path, 'r') as f:
            config_dict = json.load(f)
        return cls.from_dict(config_dict)
    
    @classmethod
    def from_dict(cls, config_dict: Dict[str, Any]) -> "TrainingConfig":
        """Create configuration from dictionary."""
        dataset_cfg = DatasetConfig(**config_dict.get('dataset', {}))
        preprocessing_cfg = PreprocessingConfig(**config_dict.get('preprocessing', {}))
        augmentation_cfg = AugmentationConfig(**config_dict.get('augmentation', {}))
        features_cfg = FeatureConfig(**config_dict.get('features', {}))
        model_cfg = ModelConfig(**config_dict.get('model', {}))
        export_cfg = ExportConfig(**config_dict.get('export', {}))
        
        return cls(
            dataset=dataset_cfg,
            preprocessing=preprocessing_cfg,
            augmentation=augmentation_cfg,
            features=features_cfg,
            model=model_cfg,
            export=export_cfg,
            experiment_name=config_dict.get('experiment_name', "trust_score_v1"),
            random_seed=config_dict.get('random_seed', 42),
            deterministic=config_dict.get('deterministic', True),
            log_level=config_dict.get('log_level', "INFO"),
        )
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert configuration to dictionary."""
        from dataclasses import asdict
        return asdict(self)
    
    def save_yaml(self, yaml_path: str) -> None:
        """Save configuration to YAML file."""
        Path(yaml_path).parent.mkdir(parents=True, exist_ok=True)
        with open(yaml_path, 'w') as f:
            yaml.dump(self.to_dict(), f, default_flow_style=False)
    
    def save_json(self, json_path: str) -> None:
        """Save configuration to JSON file."""
        Path(json_path).parent.mkdir(parents=True, exist_ok=True)
        with open(json_path, 'w') as f:
            json.dump(self.to_dict(), f, indent=2)
