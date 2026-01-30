"""
Signed label manifests for VEID training data.

This module provides:
- Creation of cryptographically signed manifests
- Hash verification of dataset contents
- Timestamp attestation
- Provenance tracking

Manifests ensure dataset integrity and enable audit trails.
"""

import base64
import hashlib
import hmac
import json
import logging
import secrets
import time
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

logger = logging.getLogger(__name__)


class SignatureAlgorithm(str, Enum):
    """Supported signature algorithms."""
    HMAC_SHA256 = "HMAC-SHA256"
    ED25519 = "Ed25519"
    RSA_SHA256 = "RSA-SHA256"


class ManifestVersion(str, Enum):
    """Manifest format versions."""
    V1 = "1.0.0"
    V2 = "2.0.0"


@dataclass
class ContentHash:
    """Hash of a content item."""
    
    item_id: str
    item_type: str  # "sample", "label", "feature"
    hash_algorithm: str = "SHA256"
    hash_value: str = ""
    size_bytes: int = 0
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "item_id": self.item_id,
            "item_type": self.item_type,
            "hash_algorithm": self.hash_algorithm,
            "hash_value": self.hash_value,
            "size_bytes": self.size_bytes,
        }


@dataclass
class ManifestSignature:
    """Cryptographic signature of a manifest."""
    
    algorithm: SignatureAlgorithm
    signer_id: str
    public_key: Optional[str] = None
    signature: str = ""
    timestamp: float = field(default_factory=lambda: datetime.now().timestamp())
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "algorithm": self.algorithm.value,
            "signer_id": self.signer_id,
            "public_key": self.public_key,
            "signature": self.signature,
            "timestamp": self.timestamp,
            "timestamp_iso": datetime.fromtimestamp(self.timestamp).isoformat(),
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "ManifestSignature":
        return cls(
            algorithm=SignatureAlgorithm(data["algorithm"]),
            signer_id=data["signer_id"],
            public_key=data.get("public_key"),
            signature=data["signature"],
            timestamp=data["timestamp"],
        )


@dataclass
class DatasetManifest:
    """
    Complete manifest for a dataset version.
    
    Contains:
    - Dataset metadata (version, splits, counts)
    - Content hashes for all items
    - Cryptographic signature
    - Build information
    """
    
    # Identification
    manifest_id: str = ""
    dataset_name: str = ""
    dataset_version: str = ""
    manifest_version: str = ManifestVersion.V2.value
    
    # Content
    content_hashes: List[ContentHash] = field(default_factory=list)
    total_samples: int = 0
    split_counts: Dict[str, int] = field(default_factory=dict)
    
    # Schema
    schema_version: str = "1.0.0"
    schema_hash: str = ""
    
    # Build info
    build_timestamp: float = field(default_factory=lambda: datetime.now().timestamp())
    build_tool_version: str = "1.0.0"
    build_config_hash: str = ""
    
    # Lineage
    source_manifests: List[str] = field(default_factory=list)
    transform_chain: List[str] = field(default_factory=list)
    
    # Signature
    signature: Optional[ManifestSignature] = None
    
    # Computed hash
    manifest_hash: str = ""
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "manifest_id": self.manifest_id,
            "dataset_name": self.dataset_name,
            "dataset_version": self.dataset_version,
            "manifest_version": self.manifest_version,
            "content_hashes": [h.to_dict() for h in self.content_hashes],
            "total_samples": self.total_samples,
            "split_counts": self.split_counts,
            "schema_version": self.schema_version,
            "schema_hash": self.schema_hash,
            "build_timestamp": self.build_timestamp,
            "build_timestamp_iso": datetime.fromtimestamp(self.build_timestamp).isoformat(),
            "build_tool_version": self.build_tool_version,
            "build_config_hash": self.build_config_hash,
            "source_manifests": self.source_manifests,
            "transform_chain": self.transform_chain,
            "signature": self.signature.to_dict() if self.signature else None,
            "manifest_hash": self.manifest_hash,
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "DatasetManifest":
        """Create from dictionary."""
        manifest = cls(
            manifest_id=data["manifest_id"],
            dataset_name=data["dataset_name"],
            dataset_version=data["dataset_version"],
            manifest_version=data.get("manifest_version", ManifestVersion.V1.value),
            content_hashes=[
                ContentHash(**h) for h in data.get("content_hashes", [])
            ],
            total_samples=data.get("total_samples", 0),
            split_counts=data.get("split_counts", {}),
            schema_version=data.get("schema_version", "1.0.0"),
            schema_hash=data.get("schema_hash", ""),
            build_timestamp=data.get("build_timestamp", 0),
            build_tool_version=data.get("build_tool_version", "1.0.0"),
            build_config_hash=data.get("build_config_hash", ""),
            source_manifests=data.get("source_manifests", []),
            transform_chain=data.get("transform_chain", []),
            manifest_hash=data.get("manifest_hash", ""),
        )
        
        if data.get("signature"):
            manifest.signature = ManifestSignature.from_dict(data["signature"])
        
        return manifest
    
    def save(self, path: str) -> None:
        """Save manifest to file."""
        file_path = Path(path)
        file_path.parent.mkdir(parents=True, exist_ok=True)
        
        with open(file_path, "w") as f:
            json.dump(self.to_dict(), f, indent=2)
        
        logger.info(f"Saved manifest to {path}")
    
    @classmethod
    def load(cls, path: str) -> "DatasetManifest":
        """Load manifest from file."""
        with open(path) as f:
            data = json.load(f)
        return cls.from_dict(data)


class ManifestSigner:
    """
    Signs manifests using cryptographic keys.
    
    Supports:
    - HMAC-SHA256 for simple signing
    - Ed25519 for asymmetric signing
    """
    
    def __init__(
        self,
        signer_id: str,
        secret_key: Optional[bytes] = None,
        algorithm: SignatureAlgorithm = SignatureAlgorithm.HMAC_SHA256,
    ):
        """
        Initialize the signer.
        
        Args:
            signer_id: Identifier for the signer
            secret_key: Secret key for signing (generated if not provided)
            algorithm: Signature algorithm to use
        """
        self.signer_id = signer_id
        self.algorithm = algorithm
        
        if secret_key:
            self._secret_key = secret_key
        else:
            self._secret_key = secrets.token_bytes(32)
        
        self._public_key: Optional[str] = None
        
        if algorithm == SignatureAlgorithm.ED25519:
            self._init_ed25519()
    
    def _init_ed25519(self) -> None:
        """Initialize Ed25519 keys."""
        try:
            from nacl.signing import SigningKey
            
            self._signing_key = SigningKey(self._secret_key[:32])
            self._public_key = base64.b64encode(
                self._signing_key.verify_key.encode()
            ).decode()
        except ImportError:
            logger.warning("nacl not available, falling back to HMAC")
            self.algorithm = SignatureAlgorithm.HMAC_SHA256
    
    def sign(self, manifest: DatasetManifest) -> DatasetManifest:
        """
        Sign a manifest.
        
        Args:
            manifest: Manifest to sign
            
        Returns:
            Signed manifest (modified in place)
        """
        # Compute manifest hash (excluding signature)
        manifest.manifest_hash = self._compute_manifest_hash(manifest)
        
        # Create signature
        timestamp = datetime.now().timestamp()
        message = self._create_signing_message(manifest, timestamp)
        
        if self.algorithm == SignatureAlgorithm.HMAC_SHA256:
            signature_bytes = hmac.new(
                self._secret_key,
                message.encode(),
                hashlib.sha256,
            ).digest()
        elif self.algorithm == SignatureAlgorithm.ED25519:
            try:
                from nacl.signing import SigningKey
                signature_bytes = self._signing_key.sign(message.encode()).signature
            except Exception as e:
                logger.error(f"Ed25519 signing failed: {e}")
                raise
        else:
            raise ValueError(f"Unsupported algorithm: {self.algorithm}")
        
        manifest.signature = ManifestSignature(
            algorithm=self.algorithm,
            signer_id=self.signer_id,
            public_key=self._public_key,
            signature=base64.b64encode(signature_bytes).decode(),
            timestamp=timestamp,
        )
        
        logger.info(f"Signed manifest {manifest.manifest_id} with {self.algorithm.value}")
        
        return manifest
    
    def _compute_manifest_hash(self, manifest: DatasetManifest) -> str:
        """Compute hash of manifest content."""
        # Create deterministic representation
        content = {
            "manifest_id": manifest.manifest_id,
            "dataset_name": manifest.dataset_name,
            "dataset_version": manifest.dataset_version,
            "content_hashes": [h.to_dict() for h in manifest.content_hashes],
            "total_samples": manifest.total_samples,
            "split_counts": manifest.split_counts,
            "schema_version": manifest.schema_version,
            "build_timestamp": manifest.build_timestamp,
        }
        
        content_str = json.dumps(content, sort_keys=True)
        return hashlib.sha256(content_str.encode()).hexdigest()
    
    def _create_signing_message(
        self,
        manifest: DatasetManifest,
        timestamp: float,
    ) -> str:
        """Create the message to sign."""
        return f"{manifest.manifest_hash}|{self.signer_id}|{timestamp}"


class ManifestVerifier:
    """Verifies manifest signatures and content integrity."""
    
    def __init__(self, trusted_signers: Optional[Dict[str, bytes]] = None):
        """
        Initialize verifier.
        
        Args:
            trusted_signers: Map of signer_id to public key or secret key
        """
        self.trusted_signers = trusted_signers or {}
    
    def verify(self, manifest: DatasetManifest) -> Dict[str, Any]:
        """
        Verify a manifest's signature and integrity.
        
        Args:
            manifest: Manifest to verify
            
        Returns:
            Verification result
        """
        result = {
            "valid": True,
            "signature_valid": False,
            "hash_valid": False,
            "content_valid": True,
            "errors": [],
            "warnings": [],
        }
        
        if not manifest.signature:
            result["valid"] = False
            result["errors"].append("Manifest is not signed")
            return result
        
        # Verify signature
        sig_result = self._verify_signature(manifest)
        result["signature_valid"] = sig_result["valid"]
        if not sig_result["valid"]:
            result["valid"] = False
            result["errors"].extend(sig_result.get("errors", []))
        
        # Verify hash
        computed_hash = self._compute_hash(manifest)
        result["hash_valid"] = computed_hash == manifest.manifest_hash
        if not result["hash_valid"]:
            result["valid"] = False
            result["errors"].append("Manifest hash mismatch")
        
        # Check timestamp
        sig_age = time.time() - manifest.signature.timestamp
        if sig_age > 86400 * 365:  # 1 year
            result["warnings"].append(f"Signature is {sig_age / 86400:.0f} days old")
        
        return result
    
    def _verify_signature(self, manifest: DatasetManifest) -> Dict[str, Any]:
        """Verify the cryptographic signature."""
        sig = manifest.signature
        result = {"valid": False, "errors": []}
        
        # Check if signer is trusted
        signer_key = self.trusted_signers.get(sig.signer_id)
        if not signer_key and sig.public_key:
            # Use embedded public key (less secure)
            signer_key = base64.b64decode(sig.public_key)
        
        if not signer_key:
            result["errors"].append(f"Unknown signer: {sig.signer_id}")
            return result
        
        message = f"{manifest.manifest_hash}|{sig.signer_id}|{sig.timestamp}"
        signature_bytes = base64.b64decode(sig.signature)
        
        try:
            if sig.algorithm == SignatureAlgorithm.HMAC_SHA256:
                expected = hmac.new(
                    signer_key,
                    message.encode(),
                    hashlib.sha256,
                ).digest()
                result["valid"] = hmac.compare_digest(expected, signature_bytes)
                
            elif sig.algorithm == SignatureAlgorithm.ED25519:
                from nacl.signing import VerifyKey
                verify_key = VerifyKey(signer_key)
                verify_key.verify(message.encode(), signature_bytes)
                result["valid"] = True
                
        except Exception as e:
            result["errors"].append(f"Signature verification failed: {e}")
        
        return result
    
    def _compute_hash(self, manifest: DatasetManifest) -> str:
        """Recompute manifest hash."""
        content = {
            "manifest_id": manifest.manifest_id,
            "dataset_name": manifest.dataset_name,
            "dataset_version": manifest.dataset_version,
            "content_hashes": [h.to_dict() for h in manifest.content_hashes],
            "total_samples": manifest.total_samples,
            "split_counts": manifest.split_counts,
            "schema_version": manifest.schema_version,
            "build_timestamp": manifest.build_timestamp,
        }
        
        content_str = json.dumps(content, sort_keys=True)
        return hashlib.sha256(content_str.encode()).hexdigest()


class ManifestBuilder:
    """Builds dataset manifests from dataset contents."""
    
    def __init__(
        self,
        dataset_name: str,
        schema_version: str = "1.0.0",
        build_tool_version: str = "1.0.0",
    ):
        """
        Initialize manifest builder.
        
        Args:
            dataset_name: Name of the dataset
            schema_version: Schema version string
            build_tool_version: Version of the build tool
        """
        self.dataset_name = dataset_name
        self.schema_version = schema_version
        self.build_tool_version = build_tool_version
        
        self._content_hashes: List[ContentHash] = []
        self._split_counts: Dict[str, int] = {}
    
    def add_sample(
        self,
        sample_id: str,
        sample_data: bytes,
        split: str = "train",
    ) -> None:
        """Add a sample to the manifest."""
        hash_value = hashlib.sha256(sample_data).hexdigest()
        
        self._content_hashes.append(ContentHash(
            item_id=sample_id,
            item_type="sample",
            hash_value=hash_value,
            size_bytes=len(sample_data),
        ))
        
        self._split_counts[split] = self._split_counts.get(split, 0) + 1
    
    def add_label(
        self,
        sample_id: str,
        label_data: bytes,
    ) -> None:
        """Add a label to the manifest."""
        hash_value = hashlib.sha256(label_data).hexdigest()
        
        self._content_hashes.append(ContentHash(
            item_id=f"{sample_id}_label",
            item_type="label",
            hash_value=hash_value,
            size_bytes=len(label_data),
        ))
    
    def add_feature(
        self,
        sample_id: str,
        feature_data: bytes,
    ) -> None:
        """Add features to the manifest."""
        hash_value = hashlib.sha256(feature_data).hexdigest()
        
        self._content_hashes.append(ContentHash(
            item_id=f"{sample_id}_features",
            item_type="feature",
            hash_value=hash_value,
            size_bytes=len(feature_data),
        ))
    
    def build(
        self,
        version: str,
        config_hash: Optional[str] = None,
        source_manifests: Optional[List[str]] = None,
    ) -> DatasetManifest:
        """
        Build the manifest.
        
        Args:
            version: Dataset version string
            config_hash: Hash of build configuration
            source_manifests: IDs of source manifests
            
        Returns:
            DatasetManifest
        """
        manifest_id = self._generate_manifest_id(version)
        
        # Compute schema hash
        schema_hash = hashlib.sha256(
            f"{self.schema_version}_{self.dataset_name}".encode()
        ).hexdigest()[:16]
        
        manifest = DatasetManifest(
            manifest_id=manifest_id,
            dataset_name=self.dataset_name,
            dataset_version=version,
            content_hashes=self._content_hashes,
            total_samples=sum(self._split_counts.values()),
            split_counts=self._split_counts,
            schema_version=self.schema_version,
            schema_hash=schema_hash,
            build_tool_version=self.build_tool_version,
            build_config_hash=config_hash or "",
            source_manifests=source_manifests or [],
        )
        
        return manifest
    
    def _generate_manifest_id(self, version: str) -> str:
        """Generate a unique manifest ID."""
        timestamp = datetime.now().strftime("%Y%m%d%H%M%S")
        random_suffix = secrets.token_hex(4)
        return f"{self.dataset_name}_{version}_{timestamp}_{random_suffix}"


def create_signed_manifest(
    dataset_name: str,
    version: str,
    samples: List[Tuple[str, bytes, str]],  # (sample_id, data, split)
    signer_id: str,
    secret_key: Optional[bytes] = None,
) -> DatasetManifest:
    """
    Convenience function to create and sign a manifest.
    
    Args:
        dataset_name: Name of the dataset
        version: Version string
        samples: List of (sample_id, data, split) tuples
        signer_id: ID of the signer
        secret_key: Secret key for signing
        
    Returns:
        Signed DatasetManifest
    """
    builder = ManifestBuilder(dataset_name)
    
    for sample_id, data, split in samples:
        builder.add_sample(sample_id, data, split)
    
    manifest = builder.build(version)
    
    signer = ManifestSigner(signer_id, secret_key)
    return signer.sign(manifest)
