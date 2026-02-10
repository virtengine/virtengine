"""
Dataset ingestion for trust score training.

This module handles loading and organizing training data from multiple sources,
supporting various document types and face samples with proper splitting.
"""

import os
import json
import logging
import hashlib
import secrets
from dataclasses import dataclass, field
from typing import List, Dict, Optional, Tuple, Iterator, Any
from pathlib import Path
from enum import Enum

import numpy as np

from ml.training.config import DatasetConfig, DocumentType

logger = logging.getLogger(__name__)


class SplitType(str, Enum):
    """Dataset split types."""
    TRAIN = "train"
    VALIDATION = "validation"
    TEST = "test"


@dataclass
class ImageData:
    """Container for image data."""
    data: np.ndarray
    path: Optional[str] = None
    format: str = "RGB"
    size: Tuple[int, int] = (0, 0)  # (width, height)
    
    def __post_init__(self):
        if self.data is not None and self.size == (0, 0):
            self.size = (self.data.shape[1], self.data.shape[0])


@dataclass
class DocumentInfo:
    """Document-specific information."""
    doc_type: str
    doc_id_hash: str  # Hashed document ID
    country_code: Optional[str] = None
    expiry_valid: Optional[bool] = None
    mrz_present: Optional[bool] = None


@dataclass
class CaptureMetadata:
    """Metadata about the capture session."""
    device_type: Optional[str] = None
    device_model: Optional[str] = None
    os_version: Optional[str] = None
    app_version: Optional[str] = None
    capture_timestamp: Optional[float] = None
    gps_available: Optional[bool] = None
    camera_facing: Optional[str] = None  # "front" or "back"
    light_level: Optional[float] = None
    motion_detected: Optional[bool] = None


@dataclass
class IdentitySample:
    """
    A single identity verification sample for training.
    
    Contains all data needed for feature extraction and training:
    - Document image (preprocessed)
    - Selfie image
    - Document metadata
    - Capture metadata
    - Ground truth label (trust score)
    """
    
    # Unique sample identifier (anonymized)
    sample_id: str
    
    # Images
    document_image: Optional[ImageData] = None
    selfie_image: Optional[ImageData] = None
    
    # Document info
    document_info: Optional[DocumentInfo] = None
    
    # Capture metadata
    capture_metadata: Optional[CaptureMetadata] = None
    
    # Ground truth
    trust_score: float = 0.0  # 0-100
    is_genuine: bool = True
    fraud_type: Optional[str] = None  # If not genuine
    
    # Quality indicators
    face_detected: bool = True
    face_confidence: float = 0.0
    document_quality_score: float = 0.0
    ocr_success: bool = True
    ocr_confidence: float = 0.0
    
    # Additional annotations
    annotations: Dict[str, Any] = field(default_factory=dict)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary (excluding image data)."""
        return {
            "sample_id": self.sample_id,
            "document_type": self.document_info.doc_type if self.document_info else None,
            "trust_score": self.trust_score,
            "is_genuine": self.is_genuine,
            "fraud_type": self.fraud_type,
            "face_detected": self.face_detected,
            "face_confidence": self.face_confidence,
            "document_quality_score": self.document_quality_score,
            "ocr_success": self.ocr_success,
            "ocr_confidence": self.ocr_confidence,
        }


@dataclass
class DatasetSplit:
    """A split of the dataset (train/val/test)."""
    
    split_type: SplitType
    samples: List[IdentitySample] = field(default_factory=list)
    
    def __len__(self) -> int:
        return len(self.samples)
    
    def __iter__(self) -> Iterator[IdentitySample]:
        return iter(self.samples)
    
    def __getitem__(self, idx: int) -> IdentitySample:
        return self.samples[idx]
    
    def get_labels(self) -> np.ndarray:
        """Get all labels (trust scores)."""
        return np.array([s.trust_score for s in self.samples])
    
    def get_statistics(self) -> Dict[str, Any]:
        """Compute statistics for this split."""
        labels = self.get_labels()
        doc_types = [s.document_info.doc_type for s in self.samples if s.document_info]
        
        return {
            "num_samples": len(self.samples),
            "label_mean": float(np.mean(labels)) if len(labels) > 0 else 0,
            "label_std": float(np.std(labels)) if len(labels) > 0 else 0,
            "label_min": float(np.min(labels)) if len(labels) > 0 else 0,
            "label_max": float(np.max(labels)) if len(labels) > 0 else 0,
            "genuine_count": sum(1 for s in self.samples if s.is_genuine),
            "fraud_count": sum(1 for s in self.samples if not s.is_genuine),
            "doc_type_distribution": {
                dt: doc_types.count(dt) for dt in set(doc_types)
            },
        }


@dataclass
class Dataset:
    """Complete dataset with train/val/test splits."""
    
    train: DatasetSplit
    validation: DatasetSplit
    test: DatasetSplit
    
    # Metadata
    config: Optional[DatasetConfig] = None
    version: str = "1.0.0"
    creation_timestamp: Optional[float] = None
    dataset_hash: Optional[str] = None
    
    def __len__(self) -> int:
        return len(self.train) + len(self.validation) + len(self.test)
    
    def get_split(self, split_type: SplitType) -> DatasetSplit:
        """Get a specific split."""
        if split_type == SplitType.TRAIN:
            return self.train
        elif split_type == SplitType.VALIDATION:
            return self.validation
        elif split_type == SplitType.TEST:
            return self.test
        raise ValueError(f"Unknown split type: {split_type}")
    
    def get_statistics(self) -> Dict[str, Any]:
        """Get statistics for all splits."""
        return {
            "total_samples": len(self),
            "train": self.train.get_statistics(),
            "validation": self.validation.get_statistics(),
            "test": self.test.get_statistics(),
            "version": self.version,
            "dataset_hash": self.dataset_hash,
        }


class DatasetIngestion:
    """
    Dataset ingestion pipeline.
    
    Loads identity verification samples from various sources,
    validates them, and splits into train/val/test sets.
    
    Supported data formats:
    - Directory structure with images and JSON metadata
    - CSV/JSON manifest files
    - TFRecord files
    """
    
    def __init__(self, config: Optional[DatasetConfig] = None):
        """
        Initialize the dataset ingestion pipeline.
        
        Args:
            config: Dataset configuration. Uses defaults if not provided.
        """
        self.config = config or DatasetConfig()
        self._rng = np.random.RandomState(self.config.random_seed)
        
        # Initialize anonymization salt
        if self.config.anonymize and not self.config.anonymization_salt:
            self.config.anonymization_salt = secrets.token_hex(32)
    
    def load_dataset(self, config: Optional[DatasetConfig] = None) -> Dataset:
        """
        Load and split the dataset.
        
        Args:
            config: Optional override configuration
            
        Returns:
            Dataset with train/val/test splits
        """
        config = config or self.config
        
        logger.info(f"Loading dataset from {len(config.data_paths)} paths")
        
        # Load all samples
        all_samples = []
        for data_path in config.data_paths:
            samples = self._load_from_path(data_path, config)
            all_samples.extend(samples)
            logger.info(f"Loaded {len(samples)} samples from {data_path}")
        
        if len(all_samples) == 0:
            logger.warning("No samples loaded, creating empty dataset")
            return self._create_empty_dataset(config)
        
        # Filter by quality
        filtered_samples = self._filter_samples(all_samples, config)
        logger.info(f"Filtered to {len(filtered_samples)} samples after quality checks")
        
        # Balance classes if requested
        if config.balance_classes:
            filtered_samples = self._balance_classes(filtered_samples, config)
            logger.info(f"Balanced to {len(filtered_samples)} samples")
        
        # Limit samples if requested
        if config.max_samples and len(filtered_samples) > config.max_samples:
            self._rng.shuffle(filtered_samples)
            filtered_samples = filtered_samples[:config.max_samples]
            logger.info(f"Limited to {len(filtered_samples)} samples")
        
        # Anonymize if requested
        if config.anonymize:
            filtered_samples = self._anonymize_samples(filtered_samples, config)
            logger.info("Applied anonymization to samples")
        
        # Split dataset
        dataset = self._split_dataset(filtered_samples, config)
        
        # Compute dataset hash
        dataset.dataset_hash = self._compute_dataset_hash(dataset)
        dataset.config = config
        
        logger.info(f"Dataset loaded: {dataset.get_statistics()}")
        
        return dataset
    
    def _load_from_path(
        self,
        data_path: str,
        config: DatasetConfig
    ) -> List[IdentitySample]:
        """Load samples from a single data path."""
        path = Path(data_path)
        
        if not path.exists():
            logger.warning(f"Data path does not exist: {data_path}")
            return []
        
        if path.is_file():
            if path.suffix == '.json':
                return self._load_from_json(path, config)
            elif path.suffix == '.csv':
                return self._load_from_csv(path, config)
            elif path.suffix in ['.tfrecord', '.tfrecords']:
                return self._load_from_tfrecord(path, config)
            else:
                logger.warning(f"Unknown file format: {path.suffix}")
                return []
        
        elif path.is_dir():
            return self._load_from_directory(path, config)
        
        return []
    
    def _load_from_directory(
        self,
        dir_path: Path,
        config: DatasetConfig
    ) -> List[IdentitySample]:
        """Load samples from a directory structure."""
        samples = []
        
        # Look for manifest file
        manifest_path = dir_path / "manifest.json"
        if manifest_path.exists():
            return self._load_from_manifest(manifest_path, config)
        
        # Otherwise, scan for sample directories
        for sample_dir in dir_path.iterdir():
            if not sample_dir.is_dir():
                continue
            
            sample = self._load_sample_from_dir(sample_dir, config)
            if sample is not None:
                samples.append(sample)
        
        return samples
    
    def _load_sample_from_dir(
        self,
        sample_dir: Path,
        config: DatasetConfig
    ) -> Optional[IdentitySample]:
        """Load a single sample from a directory."""
        try:
            # Load metadata
            metadata_path = sample_dir / "metadata.json"
            if not metadata_path.exists():
                return None
            
            with open(metadata_path, 'r') as f:
                metadata = json.load(f)
            
            # Check document type filter
            doc_type = metadata.get('document_type', '')
            if doc_type not in config.doc_types:
                return None
            
            # Load images
            document_image = self._load_image(sample_dir / "document.png")
            if document_image is None:
                document_image = self._load_image(sample_dir / "document.jpg")
            
            selfie_image = self._load_image(sample_dir / "selfie.png")
            if selfie_image is None:
                selfie_image = self._load_image(sample_dir / "selfie.jpg")
            
            # Create sample
            sample = IdentitySample(
                sample_id=sample_dir.name,
                document_image=document_image,
                selfie_image=selfie_image,
                document_info=DocumentInfo(
                    doc_type=doc_type,
                    doc_id_hash=metadata.get('doc_id_hash', ''),
                    country_code=metadata.get('country_code'),
                    expiry_valid=metadata.get('expiry_valid'),
                    mrz_present=metadata.get('mrz_present'),
                ),
                capture_metadata=CaptureMetadata(
                    device_type=metadata.get('device_type'),
                    device_model=metadata.get('device_model'),
                    os_version=metadata.get('os_version'),
                    app_version=metadata.get('app_version'),
                    capture_timestamp=metadata.get('capture_timestamp'),
                    gps_available=metadata.get('gps_available'),
                    camera_facing=metadata.get('camera_facing'),
                    light_level=metadata.get('light_level'),
                    motion_detected=metadata.get('motion_detected'),
                ),
                trust_score=float(metadata.get('trust_score', 0)),
                is_genuine=metadata.get('is_genuine', True),
                fraud_type=metadata.get('fraud_type'),
                face_detected=metadata.get('face_detected', True),
                face_confidence=float(metadata.get('face_confidence', 0)),
                document_quality_score=float(metadata.get('document_quality_score', 0)),
                ocr_success=metadata.get('ocr_success', True),
                ocr_confidence=float(metadata.get('ocr_confidence', 0)),
                annotations=metadata.get('annotations', {}),
            )
            
            return sample
            
        except Exception as e:
            logger.error(f"Error loading sample from {sample_dir}: {e}")
            return None
    
    def _load_from_manifest(
        self,
        manifest_path: Path,
        config: DatasetConfig
    ) -> List[IdentitySample]:
        """Load samples from a manifest file."""
        with open(manifest_path, 'r') as f:
            manifest = json.load(f)
        
        base_dir = manifest_path.parent
        samples = []
        
        for entry in manifest.get('samples', []):
            sample = self._load_sample_from_manifest_entry(entry, base_dir, config)
            if sample is not None:
                samples.append(sample)
        
        return samples
    
    def _load_sample_from_manifest_entry(
        self,
        entry: Dict[str, Any],
        base_dir: Path,
        config: DatasetConfig
    ) -> Optional[IdentitySample]:
        """Load a single sample from a manifest entry."""
        try:
            doc_type = entry.get('document_type', '')
            if doc_type not in config.doc_types:
                return None
            
            # Load images
            document_image = None
            if 'document_path' in entry:
                document_image = self._load_image(base_dir / entry['document_path'])
            
            selfie_image = None
            if 'selfie_path' in entry:
                selfie_image = self._load_image(base_dir / entry['selfie_path'])
            
            sample = IdentitySample(
                sample_id=entry.get('sample_id', ''),
                document_image=document_image,
                selfie_image=selfie_image,
                document_info=DocumentInfo(
                    doc_type=doc_type,
                    doc_id_hash=entry.get('doc_id_hash', ''),
                    country_code=entry.get('country_code'),
                    expiry_valid=entry.get('expiry_valid'),
                    mrz_present=entry.get('mrz_present'),
                ),
                trust_score=float(entry.get('trust_score', 0)),
                is_genuine=entry.get('is_genuine', True),
                fraud_type=entry.get('fraud_type'),
            )
            
            return sample
            
        except Exception as e:
            logger.error(f"Error loading sample from manifest entry: {e}")
            return None
    
    def _load_from_json(
        self,
        json_path: Path,
        config: DatasetConfig
    ) -> List[IdentitySample]:
        """Load samples from a JSON file."""
        return self._load_from_manifest(json_path, config)
    
    def _load_from_csv(
        self,
        csv_path: Path,
        config: DatasetConfig
    ) -> List[IdentitySample]:
        """Load samples from a CSV file."""
        import csv
        
        samples = []
        base_dir = csv_path.parent
        
        with open(csv_path, 'r') as f:
            reader = csv.DictReader(f)
            for row in reader:
                sample = self._load_sample_from_manifest_entry(
                    dict(row), base_dir, config
                )
                if sample is not None:
                    samples.append(sample)
        
        return samples
    
    def _load_from_tfrecord(
        self,
        tfrecord_path: Path,
        config: DatasetConfig
    ) -> List[IdentitySample]:
        """Load samples from a TFRecord file."""
        # TFRecord loading would require TensorFlow
        # This is a placeholder for the actual implementation
        logger.warning("TFRecord loading not fully implemented")
        return []
    
    def _load_image(self, image_path: Path) -> Optional[ImageData]:
        """Load an image from disk."""
        if not image_path.exists():
            return None
        
        try:
            # Use PIL for image loading
            from PIL import Image
            
            img = Image.open(image_path)
            img = img.convert('RGB')
            data = np.array(img)
            
            return ImageData(
                data=data,
                path=str(image_path),
                format="RGB",
                size=(img.width, img.height),
            )
        except Exception as e:
            logger.error(f"Error loading image {image_path}: {e}")
            return None
    
    def _filter_samples(
        self,
        samples: List[IdentitySample],
        config: DatasetConfig
    ) -> List[IdentitySample]:
        """Filter samples based on quality thresholds."""
        filtered = []
        
        for sample in samples:
            # Check face confidence
            if sample.face_confidence < config.min_face_confidence:
                continue
            
            # Check document quality
            if sample.document_quality_score < config.min_doc_quality:
                continue
            
            # Check OCR confidence
            if sample.ocr_confidence < config.min_ocr_confidence:
                continue
            
            filtered.append(sample)
        
        return filtered
    
    def _balance_classes(
        self,
        samples: List[IdentitySample],
        config: DatasetConfig
    ) -> List[IdentitySample]:
        """Balance classes by document type and genuine/fraud."""
        # Group by document type and genuine flag
        groups: Dict[str, List[IdentitySample]] = {}
        
        for sample in samples:
            doc_type = sample.document_info.doc_type if sample.document_info else "unknown"
            key = f"{doc_type}_{sample.is_genuine}"
            
            if key not in groups:
                groups[key] = []
            groups[key].append(sample)
        
        # Find minimum count (at least min_samples_per_class)
        min_count = max(
            config.min_samples_per_class,
            min(len(g) for g in groups.values()) if groups else 0
        )
        
        # Sample from each group
        balanced = []
        for key, group in groups.items():
            if len(group) >= min_count:
                self._rng.shuffle(group)
                balanced.extend(group[:min_count])
            else:
                # Use all samples if not enough
                balanced.extend(group)
        
        return balanced
    
    def _anonymize_samples(
        self,
        samples: List[IdentitySample],
        config: DatasetConfig
    ) -> List[IdentitySample]:
        """Anonymize PII in samples."""
        salt = config.anonymization_salt or ""
        
        for sample in samples:
            # Anonymize sample ID
            sample.sample_id = self._hash_value(sample.sample_id, salt)
            
            # Anonymize document info
            if sample.document_info:
                sample.document_info.doc_id_hash = self._hash_value(
                    sample.document_info.doc_id_hash, salt
                )
        
        return samples
    
    def _hash_value(self, value: str, salt: str) -> str:
        """Hash a value with salt for anonymization."""
        if not value:
            return ""
        
        data = f"{salt}{value}".encode('utf-8')
        return hashlib.sha256(data).hexdigest()[:16]
    
    def _split_dataset(
        self,
        samples: List[IdentitySample],
        config: DatasetConfig
    ) -> Dataset:
        """Split samples into train/val/test sets."""
        # Shuffle samples
        samples_shuffled = list(samples)
        self._rng.shuffle(samples_shuffled)
        
        n = len(samples_shuffled)
        train_end = int(n * config.train_split)
        val_end = train_end + int(n * config.val_split)
        
        train_samples = samples_shuffled[:train_end]
        val_samples = samples_shuffled[train_end:val_end]
        test_samples = samples_shuffled[val_end:]
        
        return Dataset(
            train=DatasetSplit(SplitType.TRAIN, train_samples),
            validation=DatasetSplit(SplitType.VALIDATION, val_samples),
            test=DatasetSplit(SplitType.TEST, test_samples),
        )
    
    def _create_empty_dataset(self, config: DatasetConfig) -> Dataset:
        """Create an empty dataset."""
        return Dataset(
            train=DatasetSplit(SplitType.TRAIN, []),
            validation=DatasetSplit(SplitType.VALIDATION, []),
            test=DatasetSplit(SplitType.TEST, []),
            config=config,
        )
    
    def _compute_dataset_hash(self, dataset: Dataset) -> str:
        """Compute a hash of the dataset for versioning."""
        hash_data = []
        
        for split in [dataset.train, dataset.validation, dataset.test]:
            for sample in split.samples:
                hash_data.append(sample.sample_id)
                hash_data.append(str(sample.trust_score))
        
        combined = "|".join(hash_data).encode('utf-8')
        return hashlib.sha256(combined).hexdigest()[:16]
    
    def anonymize_labels(self, dataset: Dataset) -> Dataset:
        """
        Anonymize/hash PII from labels in the dataset.
        
        This method can be called separately to apply additional
        anonymization to an already-loaded dataset.
        
        Args:
            dataset: Dataset to anonymize
            
        Returns:
            Dataset with anonymized labels
        """
        salt = self.config.anonymization_salt or secrets.token_hex(32)
        
        for split in [dataset.train, dataset.validation, dataset.test]:
            for sample in split.samples:
                sample.sample_id = self._hash_value(sample.sample_id, salt)
                if sample.document_info:
                    sample.document_info.doc_id_hash = self._hash_value(
                        sample.document_info.doc_id_hash, salt
                    )
        
        return dataset
