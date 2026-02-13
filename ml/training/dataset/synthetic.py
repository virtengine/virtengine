"""
Synthetic data generator for VEID training.

This module generates synthetic identity verification samples for:
- CI/CD pipeline testing
- Development without real PII data
- Model prototyping and experimentation

Synthetic data is deterministically generated with fixed seeds
to ensure reproducibility.
"""

import hashlib
import json
import logging
import os
import random
import string
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from enum import Enum
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

import numpy as np

from ml.training.config import DocumentType
from ml.training.dataset.ingestion import (
    CaptureMetadata,
    Dataset,
    DatasetSplit,
    DocumentInfo,
    IdentitySample,
    ImageData,
    SplitType,
)

logger = logging.getLogger(__name__)


class SyntheticProfile(str, Enum):
    """Predefined synthetic data profiles."""
    CI_MINIMAL = "ci_minimal"  # Small dataset for fast CI tests
    CI_STANDARD = "ci_standard"  # Standard CI dataset
    DEV_SMALL = "dev_small"  # Small development dataset
    DEV_MEDIUM = "dev_medium"  # Medium development dataset
    DEV_LARGE = "dev_large"  # Large development dataset
    BENCHMARK = "benchmark"  # For benchmarking performance


@dataclass
class SyntheticConfig:
    """Configuration for synthetic data generation."""
    
    # Dataset size
    num_samples: int = 100
    train_ratio: float = 0.7
    val_ratio: float = 0.15
    test_ratio: float = 0.15
    
    # Sample distribution
    fraud_ratio: float = 0.1  # Ratio of fraudulent samples
    doc_type_weights: Dict[str, float] = field(default_factory=lambda: {
        DocumentType.ID_CARD.value: 0.4,
        DocumentType.PASSPORT.value: 0.35,
        DocumentType.DRIVERS_LICENSE.value: 0.25,
    })
    
    # Quality distribution
    quality_mean: float = 0.75
    quality_std: float = 0.15
    
    # Trust score distribution
    trust_score_genuine_mean: float = 80.0
    trust_score_genuine_std: float = 10.0
    trust_score_fraud_mean: float = 30.0
    trust_score_fraud_std: float = 15.0
    
    # Image generation
    generate_images: bool = True
    image_size: Tuple[int, int] = (224, 224)
    
    # Determinism
    random_seed: int = 42
    
    # Output
    output_dir: Optional[str] = None
    save_to_disk: bool = False
    
    @classmethod
    def from_profile(cls, profile: SyntheticProfile) -> "SyntheticConfig":
        """Create config from a predefined profile."""
        profiles = {
            SyntheticProfile.CI_MINIMAL: cls(
                num_samples=30,
                generate_images=False,
            ),
            SyntheticProfile.CI_STANDARD: cls(
                num_samples=100,
                generate_images=True,
            ),
            SyntheticProfile.DEV_SMALL: cls(
                num_samples=500,
                generate_images=True,
            ),
            SyntheticProfile.DEV_MEDIUM: cls(
                num_samples=2000,
                generate_images=True,
            ),
            SyntheticProfile.DEV_LARGE: cls(
                num_samples=10000,
                generate_images=True,
            ),
            SyntheticProfile.BENCHMARK: cls(
                num_samples=5000,
                generate_images=True,
                fraud_ratio=0.2,
            ),
        }
        return profiles.get(profile, cls())


# Country codes and their weights
COUNTRY_CODES = {
    "US": 0.30,
    "GB": 0.15,
    "DE": 0.10,
    "FR": 0.08,
    "CA": 0.08,
    "AU": 0.07,
    "JP": 0.05,
    "BR": 0.05,
    "IN": 0.05,
    "MX": 0.04,
    "ES": 0.03,
}

# Device types and models
DEVICE_TYPES = ["mobile", "tablet", "desktop"]
DEVICE_MODELS = {
    "mobile": [
        "iPhone 14 Pro", "iPhone 13", "iPhone 12", "iPhone SE",
        "Samsung Galaxy S23", "Samsung Galaxy S22", "Samsung Galaxy A54",
        "Google Pixel 8", "Google Pixel 7", "Google Pixel 6",
        "OnePlus 11", "Xiaomi 13",
    ],
    "tablet": [
        "iPad Pro 12.9", "iPad Air", "iPad Mini",
        "Samsung Galaxy Tab S8", "Samsung Galaxy Tab A8",
    ],
    "desktop": [
        "MacBook Pro", "MacBook Air", "iMac",
        "Windows PC", "Linux Workstation",
    ],
}

OS_VERSIONS = {
    "iPhone": ["iOS 17.0", "iOS 16.5", "iOS 16.0", "iOS 15.7"],
    "Samsung": ["Android 14", "Android 13", "Android 12"],
    "Google": ["Android 14", "Android 13"],
    "OnePlus": ["Android 14", "Android 13"],
    "Xiaomi": ["Android 13", "Android 12"],
    "iPad": ["iPadOS 17.0", "iPadOS 16.5"],
    "MacBook": ["macOS 14.0", "macOS 13.6"],
    "iMac": ["macOS 14.0", "macOS 13.6"],
    "Windows": ["Windows 11", "Windows 10"],
    "Linux": ["Ubuntu 22.04", "Fedora 39"],
}

# Fraud types
FRAUD_TYPES = [
    "printed_photo",
    "screen_replay",
    "document_tampering",
    "identity_theft",
    "synthetic_identity",
    "deepfake",
    "mask_attack",
]


class SyntheticDataGenerator:
    """
    Generates synthetic identity verification samples.
    
    All generation is deterministic based on the seed provided,
    enabling reproducible datasets for testing.
    """
    
    def __init__(self, config: Optional[SyntheticConfig] = None):
        """
        Initialize the generator.
        
        Args:
            config: Generation configuration
        """
        self.config = config or SyntheticConfig()
        self._rng = random.Random(self.config.random_seed)
        self._np_rng = np.random.RandomState(self.config.random_seed)
        
        # Initialize weighted selectors
        self._country_codes = list(COUNTRY_CODES.keys())
        self._country_weights = list(COUNTRY_CODES.values())
        
        self._doc_types = list(self.config.doc_type_weights.keys())
        self._doc_weights = list(self.config.doc_type_weights.values())
    
    def generate_dataset(self) -> Dataset:
        """
        Generate a complete synthetic dataset.
        
        Returns:
            Dataset with train/val/test splits
        """
        logger.info(
            f"Generating synthetic dataset with {self.config.num_samples} samples"
        )
        
        # Generate all samples
        samples = []
        for i in range(self.config.num_samples):
            sample = self._generate_sample(i)
            samples.append(sample)
        
        # Split into train/val/test
        n = len(samples)
        train_end = int(n * self.config.train_ratio)
        val_end = train_end + int(n * self.config.val_ratio)
        
        # Shuffle with fixed seed
        shuffled = list(samples)
        self._rng.shuffle(shuffled)
        
        train_samples = shuffled[:train_end]
        val_samples = shuffled[train_end:val_end]
        test_samples = shuffled[val_end:]
        
        dataset = Dataset(
            train=DatasetSplit(SplitType.TRAIN, train_samples),
            validation=DatasetSplit(SplitType.VALIDATION, val_samples),
            test=DatasetSplit(SplitType.TEST, test_samples),
            version="synthetic-v1.0.0",
            creation_timestamp=datetime.now().timestamp(),
        )
        
        # Compute hash
        dataset.dataset_hash = self._compute_dataset_hash(dataset)
        
        logger.info(
            f"Generated dataset: {len(train_samples)} train, "
            f"{len(val_samples)} val, {len(test_samples)} test"
        )
        
        # Save to disk if requested
        if self.config.save_to_disk and self.config.output_dir:
            self._save_dataset(dataset)
        
        return dataset
    
    def _generate_sample(self, index: int) -> IdentitySample:
        """Generate a single synthetic sample."""
        sample_id = f"synthetic_{index:06d}"
        
        # Determine if genuine or fraud
        is_genuine = self._rng.random() > self.config.fraud_ratio
        
        # Select document type
        doc_type = self._weighted_choice(self._doc_types, self._doc_weights)
        
        # Select country
        country = self._weighted_choice(self._country_codes, self._country_weights)
        
        # Generate quality scores
        face_confidence = self._clamp(
            self._np_rng.normal(self.config.quality_mean, self.config.quality_std),
            0.0, 1.0
        )
        doc_quality = self._clamp(
            self._np_rng.normal(self.config.quality_mean, self.config.quality_std),
            0.0, 1.0
        )
        ocr_confidence = self._clamp(
            self._np_rng.normal(self.config.quality_mean, self.config.quality_std),
            0.0, 1.0
        )
        
        # Generate trust score based on genuine/fraud
        if is_genuine:
            trust_score = self._clamp(
                self._np_rng.normal(
                    self.config.trust_score_genuine_mean,
                    self.config.trust_score_genuine_std
                ),
                0.0, 100.0
            )
            fraud_type = None
        else:
            trust_score = self._clamp(
                self._np_rng.normal(
                    self.config.trust_score_fraud_mean,
                    self.config.trust_score_fraud_std
                ),
                0.0, 100.0
            )
            fraud_type = self._rng.choice(FRAUD_TYPES)
        
        # Generate images if requested
        document_image = None
        selfie_image = None
        if self.config.generate_images:
            document_image = self._generate_image(sample_id, "document")
            selfie_image = self._generate_image(sample_id, "selfie")
        
        # Generate document info
        doc_info = DocumentInfo(
            doc_type=doc_type,
            doc_id_hash=self._generate_hash(f"{sample_id}_doc"),
            country_code=country,
            expiry_valid=self._rng.random() > 0.05 if is_genuine else self._rng.random() > 0.3,
            mrz_present=doc_type == DocumentType.PASSPORT.value or self._rng.random() > 0.3,
        )
        
        # Generate capture metadata
        capture_metadata = self._generate_capture_metadata()
        
        return IdentitySample(
            sample_id=sample_id,
            document_image=document_image,
            selfie_image=selfie_image,
            document_info=doc_info,
            capture_metadata=capture_metadata,
            trust_score=trust_score,
            is_genuine=is_genuine,
            fraud_type=fraud_type,
            face_detected=face_confidence > 0.3,
            face_confidence=face_confidence,
            document_quality_score=doc_quality,
            ocr_success=ocr_confidence > 0.4,
            ocr_confidence=ocr_confidence,
            annotations={
                "synthetic": True,
                "generator_version": "1.0.0",
                "seed": self.config.random_seed,
            },
        )
    
    def _generate_image(
        self,
        sample_id: str,
        image_type: str
    ) -> ImageData:
        """Generate a synthetic image."""
        h, w = self.config.image_size
        
        # Generate deterministic noise pattern based on sample_id
        seed = int(hashlib.sha256(f"{sample_id}_{image_type}".encode()).hexdigest()[:8], 16)
        rng = np.random.RandomState(seed)
        
        # Create a simple synthetic image with patterns
        image = np.zeros((h, w, 3), dtype=np.uint8)
        
        # Add background color
        bg_color = rng.randint(180, 255, size=3)
        image[:, :] = bg_color
        
        # Add some rectangles (simulating document features)
        for _ in range(5):
            x1, y1 = rng.randint(0, w - 20), rng.randint(0, h - 20)
            x2, y2 = x1 + rng.randint(20, 80), y1 + rng.randint(10, 40)
            x2, y2 = min(x2, w), min(y2, h)
            color = rng.randint(0, 180, size=3)
            image[y1:y2, x1:x2] = color
        
        # Add noise
        noise = rng.randint(-20, 20, size=image.shape)
        image = np.clip(image.astype(np.int16) + noise, 0, 255).astype(np.uint8)
        
        return ImageData(
            data=image,
            path=f"synthetic/{sample_id}/{image_type}.png",
            format="RGB",
            size=(w, h),
        )
    
    def _generate_capture_metadata(self) -> CaptureMetadata:
        """Generate capture session metadata."""
        device_type = self._rng.choice(DEVICE_TYPES)
        device_model = self._rng.choice(DEVICE_MODELS[device_type])
        
        # Find OS version
        os_key = device_model.split()[0]
        os_versions = OS_VERSIONS.get(os_key, ["Unknown"])
        os_version = self._rng.choice(os_versions)
        
        return CaptureMetadata(
            device_type=device_type,
            device_model=device_model,
            os_version=os_version,
            app_version=self._rng.choice(["1.0.0", "1.1.0", "1.2.0", "2.0.0"]),
            capture_timestamp=datetime.now().timestamp() - self._rng.randint(0, 86400 * 30),
            gps_available=self._rng.random() > 0.1,
            camera_facing="front" if self._rng.random() > 0.3 else "back",
            light_level=self._rng.uniform(100, 1000),
            motion_detected=self._rng.random() > 0.8,
        )
    
    def _weighted_choice(self, choices: List[str], weights: List[float]) -> str:
        """Make a weighted random choice."""
        total = sum(weights)
        r = self._rng.random() * total
        cumulative = 0.0
        for choice, weight in zip(choices, weights):
            cumulative += weight
            if r <= cumulative:
                return choice
        return choices[-1]
    
    def _clamp(self, value: float, min_val: float, max_val: float) -> float:
        """Clamp a value to a range."""
        return max(min_val, min(max_val, value))
    
    def _generate_hash(self, value: str) -> str:
        """Generate a deterministic hash."""
        return hashlib.sha256(value.encode()).hexdigest()[:16]
    
    def _compute_dataset_hash(self, dataset: Dataset) -> str:
        """Compute hash of the dataset."""
        hash_data = []
        for split in [dataset.train, dataset.validation, dataset.test]:
            for sample in split.samples:
                hash_data.append(sample.sample_id)
                hash_data.append(str(sample.trust_score))
        
        combined = "|".join(hash_data).encode()
        return hashlib.sha256(combined).hexdigest()[:16]
    
    def _save_dataset(self, dataset: Dataset) -> None:
        """Save dataset to disk."""
        output_dir = Path(self.config.output_dir)
        output_dir.mkdir(parents=True, exist_ok=True)
        
        # Create manifest
        manifest = {
            "version": dataset.version,
            "creation_timestamp": dataset.creation_timestamp,
            "dataset_hash": dataset.dataset_hash,
            "config": {
                "num_samples": self.config.num_samples,
                "random_seed": self.config.random_seed,
                "fraud_ratio": self.config.fraud_ratio,
            },
            "splits": {
                "train": len(dataset.train),
                "validation": len(dataset.validation),
                "test": len(dataset.test),
            },
            "samples": [],
        }
        
        # Save each sample
        for split in [dataset.train, dataset.validation, dataset.test]:
            for sample in split.samples:
                sample_data = sample.to_dict()
                sample_data["split"] = split.split_type.value
                manifest["samples"].append(sample_data)
                
                # Save images if generated
                if sample.document_image is not None:
                    sample_dir = output_dir / sample.sample_id
                    sample_dir.mkdir(exist_ok=True)
                    self._save_image(
                        sample.document_image.data,
                        sample_dir / "document.png"
                    )
                
                if sample.selfie_image is not None:
                    sample_dir = output_dir / sample.sample_id
                    sample_dir.mkdir(exist_ok=True)
                    self._save_image(
                        sample.selfie_image.data,
                        sample_dir / "selfie.png"
                    )
        
        # Save manifest
        with open(output_dir / "manifest.json", "w") as f:
            json.dump(manifest, f, indent=2)
        
        logger.info(f"Saved synthetic dataset to {output_dir}")
    
    def _save_image(self, image: np.ndarray, path: Path) -> None:
        """Save an image to disk."""
        try:
            from PIL import Image
            img = Image.fromarray(image)
            img.save(path)
        except ImportError:
            logger.warning("PIL not available, skipping image save")


def generate_synthetic_dataset(
    num_samples: int = 100,
    profile: Optional[SyntheticProfile] = None,
    seed: int = 42,
    **kwargs
) -> Dataset:
    """
    Convenience function to generate a synthetic dataset.
    
    Args:
        num_samples: Number of samples to generate
        profile: Optional predefined profile
        seed: Random seed for reproducibility
        **kwargs: Additional config options
        
    Returns:
        Generated Dataset
    """
    if profile is not None:
        config = SyntheticConfig.from_profile(profile)
    else:
        config = SyntheticConfig(num_samples=num_samples, random_seed=seed, **kwargs)
    
    generator = SyntheticDataGenerator(config)
    return generator.generate_dataset()
