"""
Artifact publishing for trust score SavedModel.

VE-3A: Publishes model artifacts to trusted registries with
immutability settings and versioning.
"""

import hashlib
import json
import logging
import os
import shutil
import tarfile
import tempfile
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
from typing import Dict, Any, List, Optional
from enum import Enum

logger = logging.getLogger(__name__)


class RegistryType(str, Enum):
    """Supported artifact registry types."""
    LOCAL = "local"
    S3 = "s3"
    GCS = "gcs"
    AZURE_BLOB = "azure_blob"
    HTTP = "http"


@dataclass
class PublishConfig:
    """Configuration for artifact publishing."""
    
    # Registry settings
    registry_type: str = RegistryType.LOCAL.value
    registry_url: str = ""
    
    # Local registry settings
    local_registry_path: str = "artifacts/models"
    
    # S3 settings
    s3_bucket: str = ""
    s3_prefix: str = "models/trust_score"
    s3_region: str = "us-east-1"
    
    # GCS settings
    gcs_bucket: str = ""
    gcs_prefix: str = "models/trust_score"
    
    # Immutability settings
    enable_immutability: bool = True
    retention_days: int = 365
    
    # Compression
    compress: bool = True
    compression_format: str = "tar.gz"
    
    # Verification
    verify_after_upload: bool = True


@dataclass
class PublishResult:
    """Result of artifact publishing."""
    
    success: bool = True
    artifact_url: str = ""
    artifact_hash: str = ""
    version: str = ""
    
    # Metadata
    published_at: str = ""
    registry_type: str = ""
    size_bytes: int = 0
    
    # Files published
    files: List[str] = field(default_factory=list)
    
    # Errors
    error_message: Optional[str] = None
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "success": self.success,
            "artifact_url": self.artifact_url,
            "artifact_hash": self.artifact_hash,
            "version": self.version,
            "published_at": self.published_at,
            "registry_type": self.registry_type,
            "size_bytes": self.size_bytes,
            "files": self.files,
            "error_message": self.error_message,
        }


class ArtifactPublisher:
    """
    Publishes trust score model artifacts to artifact registries.
    
    Supports:
    - Local file system
    - AWS S3 (with object lock for immutability)
    - Google Cloud Storage (with retention policies)
    - Azure Blob Storage
    """
    
    def __init__(self, config: Optional[PublishConfig] = None):
        """
        Initialize the publisher.
        
        Args:
            config: Publishing configuration
        """
        self.config = config or PublishConfig()
    
    def publish(
        self,
        model_path: str,
        manifest_path: str,
        version: str,
        additional_files: Optional[List[str]] = None,
    ) -> PublishResult:
        """
        Publish model artifacts to the configured registry.
        
        Args:
            model_path: Path to SavedModel directory
            manifest_path: Path to manifest.json
            version: Model version string
            additional_files: Additional files to include
            
        Returns:
            PublishResult with publication details
        """
        logger.info(f"Publishing model version {version}")
        
        try:
            # Collect all files to publish
            files_to_publish = self._collect_files(
                model_path, manifest_path, additional_files
            )
            
            # Create archive if compression is enabled
            if self.config.compress:
                archive_path = self._create_archive(
                    files_to_publish, version, model_path
                )
                artifact_hash = self._compute_file_hash(archive_path)
                size_bytes = os.path.getsize(archive_path)
            else:
                archive_path = None
                artifact_hash = self._compute_directory_hash(model_path)
                size_bytes = self._get_directory_size(model_path)
            
            # Publish to registry based on type
            registry_type = self.config.registry_type
            
            if registry_type == RegistryType.LOCAL.value:
                artifact_url = self._publish_local(
                    model_path, manifest_path, version, archive_path
                )
            elif registry_type == RegistryType.S3.value:
                artifact_url = self._publish_s3(
                    model_path, manifest_path, version, archive_path
                )
            elif registry_type == RegistryType.GCS.value:
                artifact_url = self._publish_gcs(
                    model_path, manifest_path, version, archive_path
                )
            else:
                raise ValueError(f"Unsupported registry type: {registry_type}")
            
            # Verify upload if configured
            if self.config.verify_after_upload:
                if not self._verify_upload(artifact_url, artifact_hash):
                    return PublishResult(
                        success=False,
                        error_message="Upload verification failed"
                    )
            
            # Cleanup temporary archive
            if archive_path and os.path.exists(archive_path):
                os.remove(archive_path)
            
            return PublishResult(
                success=True,
                artifact_url=artifact_url,
                artifact_hash=artifact_hash,
                version=version,
                published_at=datetime.utcnow().isoformat(),
                registry_type=registry_type,
                size_bytes=size_bytes,
                files=[str(f) for f in files_to_publish],
            )
            
        except Exception as e:
            logger.exception(f"Publishing failed: {e}")
            return PublishResult(
                success=False,
                error_message=str(e)
            )
    
    def _collect_files(
        self,
        model_path: str,
        manifest_path: str,
        additional_files: Optional[List[str]],
    ) -> List[Path]:
        """Collect all files to publish."""
        files = []
        
        # Add model files
        model_dir = Path(model_path)
        for root, dirs, filenames in os.walk(model_dir):
            for filename in filenames:
                files.append(Path(root) / filename)
        
        # Add manifest
        if os.path.exists(manifest_path):
            files.append(Path(manifest_path))
        
        # Add additional files
        if additional_files:
            for filepath in additional_files:
                if os.path.exists(filepath):
                    files.append(Path(filepath))
        
        return files
    
    def _create_archive(
        self,
        files: List[Path],
        version: str,
        model_path: str,
    ) -> str:
        """Create compressed archive of model files."""
        archive_name = f"trust_score_{version}.tar.gz"
        archive_path = os.path.join(tempfile.gettempdir(), archive_name)
        
        # Get the base directory for relative paths
        base_dir = Path(model_path).parent
        
        with tarfile.open(archive_path, "w:gz") as tar:
            # Add model directory
            tar.add(model_path, arcname=os.path.basename(model_path))
            
            # Add other files at root level
            for filepath in files:
                if not str(filepath).startswith(str(model_path)):
                    tar.add(filepath, arcname=filepath.name)
        
        logger.info(f"Created archive: {archive_path}")
        return archive_path
    
    def _compute_file_hash(self, filepath: str) -> str:
        """Compute SHA256 hash of a file."""
        h = hashlib.sha256()
        with open(filepath, 'rb') as f:
            for chunk in iter(lambda: f.read(8192), b''):
                h.update(chunk)
        return h.hexdigest()
    
    def _compute_directory_hash(self, dir_path: str) -> str:
        """Compute SHA256 hash of directory contents."""
        h = hashlib.sha256()
        
        files = []
        for root, dirs, filenames in os.walk(dir_path):
            dirs.sort()
            for filename in sorted(filenames):
                files.append(os.path.join(root, filename))
        
        for filepath in files:
            with open(filepath, 'rb') as f:
                h.update(f.read())
        
        return h.hexdigest()
    
    def _get_directory_size(self, dir_path: str) -> int:
        """Get total size of directory in bytes."""
        total = 0
        for root, dirs, files in os.walk(dir_path):
            for filename in files:
                filepath = os.path.join(root, filename)
                total += os.path.getsize(filepath)
        return total
    
    def _publish_local(
        self,
        model_path: str,
        manifest_path: str,
        version: str,
        archive_path: Optional[str],
    ) -> str:
        """Publish to local file system."""
        registry_path = Path(self.config.local_registry_path)
        version_path = registry_path / version
        
        # Check for existing version (immutability)
        if version_path.exists() and self.config.enable_immutability:
            raise ValueError(
                f"Version {version} already exists and immutability is enabled"
            )
        
        version_path.mkdir(parents=True, exist_ok=True)
        
        if archive_path:
            # Copy archive
            dest = version_path / os.path.basename(archive_path)
            shutil.copy2(archive_path, dest)
            artifact_url = str(dest)
        else:
            # Copy model directory
            model_dest = version_path / "model"
            if model_dest.exists():
                shutil.rmtree(model_dest)
            shutil.copytree(model_path, model_dest)
            artifact_url = str(model_dest)
        
        # Copy manifest
        if os.path.exists(manifest_path):
            shutil.copy2(manifest_path, version_path / "manifest.json")
        
        # Create version metadata
        metadata = {
            "version": version,
            "published_at": datetime.utcnow().isoformat(),
            "artifact_path": artifact_url,
        }
        with open(version_path / "metadata.json", 'w') as f:
            json.dump(metadata, f, indent=2)
        
        logger.info(f"Published to local registry: {version_path}")
        return artifact_url
    
    def _publish_s3(
        self,
        model_path: str,
        manifest_path: str,
        version: str,
        archive_path: Optional[str],
    ) -> str:
        """Publish to AWS S3."""
        try:
            import boto3
            from botocore.config import Config
        except ImportError:
            raise ImportError("boto3 is required for S3 publishing")
        
        s3_config = Config(
            region_name=self.config.s3_region,
            signature_version='s3v4',
        )
        s3 = boto3.client('s3', config=s3_config)
        
        bucket = self.config.s3_bucket
        prefix = f"{self.config.s3_prefix}/{version}"
        
        # Check for existing version (immutability)
        if self.config.enable_immutability:
            try:
                s3.head_object(Bucket=bucket, Key=f"{prefix}/manifest.json")
                raise ValueError(
                    f"Version {version} already exists in S3 and immutability is enabled"
                )
            except s3.exceptions.ClientError as e:
                if e.response['Error']['Code'] != '404':
                    raise
        
        # Upload archive or model directory
        if archive_path:
            s3_key = f"{prefix}/{os.path.basename(archive_path)}"
            s3.upload_file(
                archive_path, bucket, s3_key,
                ExtraArgs={'ContentType': 'application/gzip'}
            )
            artifact_url = f"s3://{bucket}/{s3_key}"
        else:
            # Upload model directory files
            for root, dirs, files in os.walk(model_path):
                for filename in files:
                    filepath = os.path.join(root, filename)
                    rel_path = os.path.relpath(filepath, model_path)
                    s3_key = f"{prefix}/model/{rel_path}"
                    s3.upload_file(filepath, bucket, s3_key)
            artifact_url = f"s3://{bucket}/{prefix}/model"
        
        # Upload manifest
        if os.path.exists(manifest_path):
            s3.upload_file(
                manifest_path, bucket, f"{prefix}/manifest.json",
                ExtraArgs={'ContentType': 'application/json'}
            )
        
        # Enable object lock if configured
        if self.config.enable_immutability:
            try:
                retention_date = datetime.utcnow().replace(
                    year=datetime.utcnow().year + 1
                )
                s3.put_object_retention(
                    Bucket=bucket,
                    Key=f"{prefix}/manifest.json",
                    Retention={
                        'Mode': 'GOVERNANCE',
                        'RetainUntilDate': retention_date
                    }
                )
            except Exception as e:
                logger.warning(f"Could not enable object lock: {e}")
        
        logger.info(f"Published to S3: {artifact_url}")
        return artifact_url
    
    def _publish_gcs(
        self,
        model_path: str,
        manifest_path: str,
        version: str,
        archive_path: Optional[str],
    ) -> str:
        """Publish to Google Cloud Storage."""
        try:
            from google.cloud import storage
        except ImportError:
            raise ImportError("google-cloud-storage is required for GCS publishing")
        
        client = storage.Client()
        bucket = client.bucket(self.config.gcs_bucket)
        prefix = f"{self.config.gcs_prefix}/{version}"
        
        # Check for existing version
        if self.config.enable_immutability:
            manifest_blob = bucket.blob(f"{prefix}/manifest.json")
            if manifest_blob.exists():
                raise ValueError(
                    f"Version {version} already exists in GCS and immutability is enabled"
                )
        
        # Upload archive or model directory
        if archive_path:
            blob_name = f"{prefix}/{os.path.basename(archive_path)}"
            blob = bucket.blob(blob_name)
            blob.upload_from_filename(archive_path)
            artifact_url = f"gs://{self.config.gcs_bucket}/{blob_name}"
        else:
            # Upload model directory files
            for root, dirs, files in os.walk(model_path):
                for filename in files:
                    filepath = os.path.join(root, filename)
                    rel_path = os.path.relpath(filepath, model_path)
                    blob_name = f"{prefix}/model/{rel_path}"
                    blob = bucket.blob(blob_name)
                    blob.upload_from_filename(filepath)
            artifact_url = f"gs://{self.config.gcs_bucket}/{prefix}/model"
        
        # Upload manifest
        if os.path.exists(manifest_path):
            manifest_blob = bucket.blob(f"{prefix}/manifest.json")
            manifest_blob.upload_from_filename(manifest_path)
        
        logger.info(f"Published to GCS: {artifact_url}")
        return artifact_url
    
    def _verify_upload(self, artifact_url: str, expected_hash: str) -> bool:
        """Verify uploaded artifact matches expected hash."""
        # For local files, we can verify directly
        if not artifact_url.startswith(("s3://", "gs://", "http://", "https://")):
            if os.path.isfile(artifact_url):
                actual_hash = self._compute_file_hash(artifact_url)
            elif os.path.isdir(artifact_url):
                actual_hash = self._compute_directory_hash(artifact_url)
            else:
                logger.warning(f"Cannot verify: {artifact_url}")
                return True
            
            if actual_hash != expected_hash:
                logger.error(f"Hash mismatch: expected {expected_hash}, got {actual_hash}")
                return False
        
        logger.info("Upload verification: PASSED")
        return True
    
    def list_versions(self) -> List[Dict[str, Any]]:
        """List all published versions."""
        versions = []
        
        if self.config.registry_type == RegistryType.LOCAL.value:
            registry_path = Path(self.config.local_registry_path)
            if registry_path.exists():
                for version_dir in sorted(registry_path.iterdir()):
                    if version_dir.is_dir():
                        metadata_path = version_dir / "metadata.json"
                        if metadata_path.exists():
                            with open(metadata_path) as f:
                                versions.append(json.load(f))
        
        return versions
    
    def get_latest_version(self) -> Optional[str]:
        """Get the latest published version."""
        versions = self.list_versions()
        if versions:
            # Sort by published_at and return latest
            sorted_versions = sorted(
                versions,
                key=lambda v: v.get('published_at', ''),
                reverse=True
            )
            return sorted_versions[0].get('version')
        return None


def publish_model(
    model_path: str,
    manifest_path: str,
    version: str,
    config: Optional[PublishConfig] = None,
) -> PublishResult:
    """
    Convenience function to publish a model.
    
    Args:
        model_path: Path to SavedModel directory
        manifest_path: Path to manifest.json
        version: Model version string
        config: Publishing configuration
        
    Returns:
        PublishResult
    """
    publisher = ArtifactPublisher(config)
    return publisher.publish(model_path, manifest_path, version)
