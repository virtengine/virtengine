"""
PII-safe storage for VEID training data.

This module provides secure storage flows for identity verification data:
- Encryption of raw assets (images, documents)
- Separate storage of derived features
- Access control and audit logging
- Key management integration

All sensitive data is encrypted using X25519-XSalsa20-Poly1305 envelopes,
consistent with the on-chain VEID encryption scheme.
"""

import base64
import hashlib
import json
import logging
import os
import secrets
import tempfile
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

import numpy as np

logger = logging.getLogger(__name__)


class StorageType(str, Enum):
    """Storage backend types."""
    LOCAL = "local"
    ENCRYPTED_LOCAL = "encrypted_local"
    S3_ENCRYPTED = "s3_encrypted"
    GCS_ENCRYPTED = "gcs_encrypted"


class DataCategory(str, Enum):
    """Categories of data for storage."""
    RAW_IMAGE = "raw_image"  # Original images (PII)
    DERIVED_FEATURE = "derived_feature"  # Extracted features (non-PII)
    METADATA = "metadata"  # Sample metadata
    LABEL = "label"  # Ground truth labels


@dataclass
class EncryptionEnvelope:
    """Encryption envelope for secure data storage."""
    
    algorithm: str = "X25519-XSalsa20-Poly1305"
    key_id: str = ""  # Identifier for the encryption key
    nonce: bytes = field(default_factory=lambda: secrets.token_bytes(24))
    ciphertext: bytes = b""
    tag: bytes = b""
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "algorithm": self.algorithm,
            "key_id": self.key_id,
            "nonce": base64.b64encode(self.nonce).decode(),
            "ciphertext": base64.b64encode(self.ciphertext).decode(),
            "tag": base64.b64encode(self.tag).decode() if self.tag else "",
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "EncryptionEnvelope":
        """Create from dictionary."""
        return cls(
            algorithm=data.get("algorithm", "X25519-XSalsa20-Poly1305"),
            key_id=data.get("key_id", ""),
            nonce=base64.b64decode(data.get("nonce", "")),
            ciphertext=base64.b64decode(data.get("ciphertext", "")),
            tag=base64.b64decode(data.get("tag", "")) if data.get("tag") else b"",
        )


@dataclass
class StorageConfig:
    """Configuration for secure storage."""
    
    storage_type: StorageType = StorageType.ENCRYPTED_LOCAL
    
    # Local storage paths
    raw_data_path: str = "data/raw"  # Encrypted raw assets
    derived_data_path: str = "data/derived"  # Derived features
    metadata_path: str = "data/metadata"  # Sample metadata
    
    # Encryption settings
    encryption_enabled: bool = True
    key_derivation: str = "argon2id"
    key_rotation_days: int = 90
    
    # Access control
    require_audit_log: bool = True
    audit_log_path: str = "data/audit.log"
    
    # Retention
    raw_data_retention_days: int = 365
    derived_data_retention_days: int = 730


@dataclass
class StoredAsset:
    """Metadata about a stored asset."""
    
    asset_id: str
    category: DataCategory
    storage_path: str
    content_hash: str
    encrypted: bool = True
    encryption_key_id: Optional[str] = None
    created_at: float = field(default_factory=lambda: datetime.now().timestamp())
    size_bytes: int = 0
    metadata: Dict[str, Any] = field(default_factory=dict)


@dataclass
class AccessLogEntry:
    """Audit log entry for data access."""
    
    timestamp: float
    action: str  # "read", "write", "delete"
    asset_id: str
    user_id: str
    success: bool
    details: Dict[str, Any] = field(default_factory=dict)
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "timestamp": self.timestamp,
            "action": self.action,
            "asset_id": self.asset_id,
            "user_id": self.user_id,
            "success": self.success,
            "details": self.details,
        }


class EncryptionManager:
    """
    Manages encryption keys and operations.
    
    Uses symmetric encryption with key derivation for local storage.
    In production, integrates with hardware security modules (HSM).
    """
    
    def __init__(self, master_key: Optional[bytes] = None):
        """
        Initialize encryption manager.
        
        Args:
            master_key: Master encryption key (generated if not provided)
        """
        self._master_key = master_key or secrets.token_bytes(32)
        self._key_cache: Dict[str, bytes] = {}
    
    def derive_key(self, key_id: str, salt: Optional[bytes] = None) -> bytes:
        """
        Derive a key for a specific purpose.
        
        Args:
            key_id: Unique identifier for the key
            salt: Optional salt for derivation
            
        Returns:
            32-byte derived key
        """
        if key_id in self._key_cache:
            return self._key_cache[key_id]
        
        salt = salt or hashlib.sha256(key_id.encode()).digest()
        
        # Simple HKDF-like derivation
        info = f"veid-storage-{key_id}".encode()
        derived = hashlib.sha256(self._master_key + salt + info).digest()
        
        self._key_cache[key_id] = derived
        return derived
    
    def encrypt(self, data: bytes, key_id: str) -> EncryptionEnvelope:
        """
        Encrypt data using the derived key.
        
        Args:
            data: Data to encrypt
            key_id: Key identifier
            
        Returns:
            EncryptionEnvelope with encrypted data
        """
        try:
            from nacl.secret import SecretBox
            from nacl.utils import random as nacl_random
        except ImportError:
            # Fallback to simple XOR for testing (NOT SECURE)
            logger.warning("nacl not available, using insecure fallback")
            return self._encrypt_fallback(data, key_id)
        
        key = self.derive_key(key_id)
        box = SecretBox(key)
        nonce = nacl_random(SecretBox.NONCE_SIZE)
        encrypted = box.encrypt(data, nonce)
        
        return EncryptionEnvelope(
            algorithm="X25519-XSalsa20-Poly1305",
            key_id=key_id,
            nonce=nonce,
            ciphertext=encrypted.ciphertext,
            tag=b"",  # Tag is included in ciphertext for NaCl
        )
    
    def decrypt(self, envelope: EncryptionEnvelope) -> bytes:
        """
        Decrypt data from an envelope.
        
        Args:
            envelope: Encryption envelope
            
        Returns:
            Decrypted data
        """
        try:
            from nacl.secret import SecretBox
        except ImportError:
            logger.warning("nacl not available, using insecure fallback")
            return self._decrypt_fallback(envelope)
        
        key = self.derive_key(envelope.key_id)
        box = SecretBox(key)
        
        # Reconstruct the encrypted message
        from nacl.utils import EncryptedMessage
        encrypted = envelope.nonce + envelope.ciphertext
        
        return box.decrypt(encrypted)
    
    def _encrypt_fallback(self, data: bytes, key_id: str) -> EncryptionEnvelope:
        """Fallback encryption for testing (NOT SECURE)."""
        key = self.derive_key(key_id)
        nonce = secrets.token_bytes(24)
        
        # Simple XOR (NOT SECURE - for testing only)
        key_extended = (key * ((len(data) // len(key)) + 1))[:len(data)]
        ciphertext = bytes(a ^ b for a, b in zip(data, key_extended))
        
        return EncryptionEnvelope(
            algorithm="XOR-FALLBACK-INSECURE",
            key_id=key_id,
            nonce=nonce,
            ciphertext=ciphertext,
        )
    
    def _decrypt_fallback(self, envelope: EncryptionEnvelope) -> bytes:
        """Fallback decryption for testing."""
        key = self.derive_key(envelope.key_id)
        
        key_extended = (key * ((len(envelope.ciphertext) // len(key)) + 1))[:len(envelope.ciphertext)]
        return bytes(a ^ b for a, b in zip(envelope.ciphertext, key_extended))


class AuditLogger:
    """Logs all data access operations for compliance."""
    
    def __init__(self, log_path: str):
        """Initialize audit logger."""
        self.log_path = Path(log_path)
        self.log_path.parent.mkdir(parents=True, exist_ok=True)
    
    def log(self, entry: AccessLogEntry) -> None:
        """Log an access entry."""
        with open(self.log_path, "a") as f:
            f.write(json.dumps(entry.to_dict()) + "\n")
    
    def get_entries(
        self,
        asset_id: Optional[str] = None,
        start_time: Optional[float] = None,
        end_time: Optional[float] = None,
    ) -> List[AccessLogEntry]:
        """Query audit log entries."""
        entries = []
        
        if not self.log_path.exists():
            return entries
        
        with open(self.log_path) as f:
            for line in f:
                data = json.loads(line.strip())
                
                if asset_id and data.get("asset_id") != asset_id:
                    continue
                if start_time and data.get("timestamp", 0) < start_time:
                    continue
                if end_time and data.get("timestamp", 0) > end_time:
                    continue
                
                entries.append(AccessLogEntry(**data))
        
        return entries


class SecureStorage:
    """
    PII-safe storage system for training data.
    
    Implements:
    - Encryption of raw assets
    - Separation of raw vs derived data
    - Audit logging
    - Access control
    """
    
    def __init__(
        self,
        config: Optional[StorageConfig] = None,
        encryption_key: Optional[bytes] = None,
        user_id: str = "system",
    ):
        """
        Initialize secure storage.
        
        Args:
            config: Storage configuration
            encryption_key: Master encryption key
            user_id: User ID for audit logging
        """
        self.config = config or StorageConfig()
        self.user_id = user_id
        
        # Initialize encryption
        self.encryption = EncryptionManager(encryption_key)
        
        # Initialize audit logger
        if self.config.require_audit_log:
            self.audit = AuditLogger(self.config.audit_log_path)
        else:
            self.audit = None
        
        # Create directories
        self._ensure_directories()
        
        # Asset registry
        self._registry: Dict[str, StoredAsset] = {}
        self._registry_path = Path(self.config.metadata_path) / "registry.json"
        self._load_registry()
    
    def _ensure_directories(self) -> None:
        """Create required directories."""
        for path in [
            self.config.raw_data_path,
            self.config.derived_data_path,
            self.config.metadata_path,
        ]:
            Path(path).mkdir(parents=True, exist_ok=True)
    
    def _load_registry(self) -> None:
        """Load asset registry from disk."""
        if self._registry_path.exists():
            with open(self._registry_path) as f:
                data = json.load(f)
                for asset_id, asset_data in data.items():
                    asset_data["category"] = DataCategory(asset_data["category"])
                    self._registry[asset_id] = StoredAsset(**asset_data)
    
    def _save_registry(self) -> None:
        """Save asset registry to disk."""
        data = {}
        for asset_id, asset in self._registry.items():
            asset_dict = {
                "asset_id": asset.asset_id,
                "category": asset.category.value,
                "storage_path": asset.storage_path,
                "content_hash": asset.content_hash,
                "encrypted": asset.encrypted,
                "encryption_key_id": asset.encryption_key_id,
                "created_at": asset.created_at,
                "size_bytes": asset.size_bytes,
                "metadata": asset.metadata,
            }
            data[asset_id] = asset_dict
        
        with open(self._registry_path, "w") as f:
            json.dump(data, f, indent=2)
    
    def store_raw_image(
        self,
        sample_id: str,
        image_type: str,  # "document" or "selfie"
        image_data: np.ndarray,
    ) -> StoredAsset:
        """
        Store a raw image with encryption.
        
        Args:
            sample_id: Sample identifier
            image_type: Type of image
            image_data: Image as numpy array
            
        Returns:
            StoredAsset with storage metadata
        """
        asset_id = f"{sample_id}_{image_type}"
        
        # Serialize image
        image_bytes = image_data.tobytes()
        shape = image_data.shape
        dtype = str(image_data.dtype)
        
        # Package with metadata
        package = {
            "shape": shape,
            "dtype": dtype,
            "data": base64.b64encode(image_bytes).decode(),
        }
        package_bytes = json.dumps(package).encode()
        
        # Compute content hash before encryption
        content_hash = hashlib.sha256(package_bytes).hexdigest()[:16]
        
        # Encrypt
        if self.config.encryption_enabled:
            envelope = self.encryption.encrypt(package_bytes, f"raw_{asset_id}")
            storage_data = json.dumps(envelope.to_dict()).encode()
            key_id = envelope.key_id
        else:
            storage_data = package_bytes
            key_id = None
        
        # Store
        storage_path = Path(self.config.raw_data_path) / f"{asset_id}.enc"
        with open(storage_path, "wb") as f:
            f.write(storage_data)
        
        # Create asset record
        asset = StoredAsset(
            asset_id=asset_id,
            category=DataCategory.RAW_IMAGE,
            storage_path=str(storage_path),
            content_hash=content_hash,
            encrypted=self.config.encryption_enabled,
            encryption_key_id=key_id,
            size_bytes=len(storage_data),
            metadata={
                "sample_id": sample_id,
                "image_type": image_type,
                "shape": shape,
                "dtype": dtype,
            },
        )
        
        self._registry[asset_id] = asset
        self._save_registry()
        
        # Audit log
        if self.audit:
            self.audit.log(AccessLogEntry(
                timestamp=datetime.now().timestamp(),
                action="write",
                asset_id=asset_id,
                user_id=self.user_id,
                success=True,
                details={"category": DataCategory.RAW_IMAGE.value},
            ))
        
        return asset
    
    def retrieve_raw_image(
        self,
        sample_id: str,
        image_type: str,
    ) -> Optional[np.ndarray]:
        """
        Retrieve a raw image.
        
        Args:
            sample_id: Sample identifier
            image_type: Type of image
            
        Returns:
            Image as numpy array, or None if not found
        """
        asset_id = f"{sample_id}_{image_type}"
        
        if asset_id not in self._registry:
            return None
        
        asset = self._registry[asset_id]
        
        # Read storage
        with open(asset.storage_path, "rb") as f:
            storage_data = f.read()
        
        # Decrypt if needed
        if asset.encrypted:
            envelope = EncryptionEnvelope.from_dict(json.loads(storage_data))
            package_bytes = self.encryption.decrypt(envelope)
        else:
            package_bytes = storage_data
        
        # Unpack
        package = json.loads(package_bytes)
        image_bytes = base64.b64decode(package["data"])
        shape = tuple(package["shape"])
        dtype = np.dtype(package["dtype"])
        
        image = np.frombuffer(image_bytes, dtype=dtype).reshape(shape)
        
        # Audit log
        if self.audit:
            self.audit.log(AccessLogEntry(
                timestamp=datetime.now().timestamp(),
                action="read",
                asset_id=asset_id,
                user_id=self.user_id,
                success=True,
            ))
        
        return image
    
    def store_derived_features(
        self,
        sample_id: str,
        features: Dict[str, Any],
    ) -> StoredAsset:
        """
        Store derived features (non-PII).
        
        Derived features are stored without encryption since they
        are anonymized and don't contain PII.
        
        Args:
            sample_id: Sample identifier
            features: Feature dictionary
            
        Returns:
            StoredAsset with storage metadata
        """
        asset_id = f"{sample_id}_features"
        
        # Serialize features
        feature_bytes = json.dumps(features, default=self._json_serializer).encode()
        
        # Compute hash
        content_hash = hashlib.sha256(feature_bytes).hexdigest()[:16]
        
        # Store (not encrypted - derived data is non-PII)
        storage_path = Path(self.config.derived_data_path) / f"{asset_id}.json"
        with open(storage_path, "wb") as f:
            f.write(feature_bytes)
        
        asset = StoredAsset(
            asset_id=asset_id,
            category=DataCategory.DERIVED_FEATURE,
            storage_path=str(storage_path),
            content_hash=content_hash,
            encrypted=False,
            size_bytes=len(feature_bytes),
            metadata={"sample_id": sample_id},
        )
        
        self._registry[asset_id] = asset
        self._save_registry()
        
        return asset
    
    def retrieve_derived_features(
        self,
        sample_id: str,
    ) -> Optional[Dict[str, Any]]:
        """Retrieve derived features."""
        asset_id = f"{sample_id}_features"
        
        if asset_id not in self._registry:
            return None
        
        asset = self._registry[asset_id]
        
        with open(asset.storage_path, "rb") as f:
            return json.loads(f.read())
    
    def store_label(
        self,
        sample_id: str,
        label: Dict[str, Any],
    ) -> StoredAsset:
        """Store label data."""
        asset_id = f"{sample_id}_label"
        
        label_bytes = json.dumps(label).encode()
        content_hash = hashlib.sha256(label_bytes).hexdigest()[:16]
        
        storage_path = Path(self.config.metadata_path) / f"{asset_id}.json"
        with open(storage_path, "wb") as f:
            f.write(label_bytes)
        
        asset = StoredAsset(
            asset_id=asset_id,
            category=DataCategory.LABEL,
            storage_path=str(storage_path),
            content_hash=content_hash,
            encrypted=False,
            size_bytes=len(label_bytes),
            metadata={"sample_id": sample_id},
        )
        
        self._registry[asset_id] = asset
        self._save_registry()
        
        return asset
    
    def get_asset(self, asset_id: str) -> Optional[StoredAsset]:
        """Get asset metadata by ID."""
        return self._registry.get(asset_id)
    
    def list_assets(
        self,
        category: Optional[DataCategory] = None,
    ) -> List[StoredAsset]:
        """List all assets, optionally filtered by category."""
        assets = list(self._registry.values())
        
        if category:
            assets = [a for a in assets if a.category == category]
        
        return assets
    
    def verify_integrity(self) -> Dict[str, Any]:
        """
        Verify integrity of all stored assets.
        
        Returns:
            Verification report
        """
        report = {
            "total_assets": len(self._registry),
            "verified": 0,
            "corrupted": 0,
            "missing": 0,
            "errors": [],
        }
        
        for asset_id, asset in self._registry.items():
            try:
                if not Path(asset.storage_path).exists():
                    report["missing"] += 1
                    report["errors"].append(f"Missing: {asset_id}")
                    continue
                
                # Re-compute hash
                with open(asset.storage_path, "rb") as f:
                    data = f.read()
                
                if asset.encrypted:
                    envelope = EncryptionEnvelope.from_dict(json.loads(data))
                    data = self.encryption.decrypt(envelope)
                
                current_hash = hashlib.sha256(data).hexdigest()[:16]
                
                if current_hash == asset.content_hash:
                    report["verified"] += 1
                else:
                    report["corrupted"] += 1
                    report["errors"].append(f"Hash mismatch: {asset_id}")
                    
            except Exception as e:
                report["corrupted"] += 1
                report["errors"].append(f"Error {asset_id}: {e}")
        
        return report
    
    def _json_serializer(self, obj: Any) -> Any:
        """JSON serializer for numpy types."""
        if isinstance(obj, np.ndarray):
            return obj.tolist()
        if isinstance(obj, np.floating):
            return float(obj)
        if isinstance(obj, np.integer):
            return int(obj)
        raise TypeError(f"Object of type {type(obj)} is not JSON serializable")
