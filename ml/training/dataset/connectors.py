"""
Data source connectors for dataset ingestion.

This module provides a registry of connectors for loading data from:
- Local filesystem
- Object storage (S3, GCS, Azure Blob)
- API feeds
- Database sources

Each connector implements the DataConnector protocol for consistent interface.
"""

import abc
import json
import hashlib
import logging
import os
import tempfile
import urllib.request
import urllib.error
from dataclasses import dataclass, field
from enum import Enum
from pathlib import Path
from typing import (
    Any,
    Callable,
    Dict,
    Iterator,
    List,
    Optional,
    Protocol,
    Type,
    runtime_checkable,
)

logger = logging.getLogger(__name__)


class ConnectorType(str, Enum):
    """Supported connector types."""
    LOCAL = "local"
    S3 = "s3"
    GCS = "gcs"
    AZURE_BLOB = "azure_blob"
    HTTP_API = "http_api"
    DATABASE = "database"


@dataclass
class ConnectorConfig:
    """Configuration for a data connector."""
    
    connector_type: ConnectorType
    
    # Connection settings
    endpoint: Optional[str] = None
    bucket: Optional[str] = None
    prefix: Optional[str] = None
    
    # Authentication
    auth_type: Optional[str] = None  # "key", "iam", "oauth", "env"
    credentials_path: Optional[str] = None
    access_key: Optional[str] = None
    secret_key: Optional[str] = None
    
    # API settings
    api_url: Optional[str] = None
    api_key: Optional[str] = None
    api_headers: Dict[str, str] = field(default_factory=dict)
    
    # Database settings
    connection_string: Optional[str] = None
    query: Optional[str] = None
    
    # General settings
    timeout_seconds: float = 30.0
    retry_count: int = 3
    cache_dir: Optional[str] = None
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary (excluding secrets)."""
        return {
            "connector_type": self.connector_type.value,
            "endpoint": self.endpoint,
            "bucket": self.bucket,
            "prefix": self.prefix,
            "auth_type": self.auth_type,
            "timeout_seconds": self.timeout_seconds,
            "retry_count": self.retry_count,
        }


@dataclass
class DataRecord:
    """A single data record from a connector."""
    
    record_id: str
    data: Dict[str, Any]
    source_path: Optional[str] = None
    source_hash: Optional[str] = None
    metadata: Dict[str, Any] = field(default_factory=dict)


@dataclass
class ConnectionResult:
    """Result of a connection attempt."""
    
    success: bool
    connector_type: ConnectorType
    message: str = ""
    record_count: int = 0
    metadata: Dict[str, Any] = field(default_factory=dict)


@runtime_checkable
class DataConnector(Protocol):
    """Protocol for data connectors."""
    
    @property
    def connector_type(self) -> ConnectorType:
        """Get the connector type."""
        ...
    
    def connect(self) -> ConnectionResult:
        """Establish connection to the data source."""
        ...
    
    def list_records(self) -> List[str]:
        """List available record IDs."""
        ...
    
    def get_record(self, record_id: str) -> Optional[DataRecord]:
        """Get a single record by ID."""
        ...
    
    def iter_records(self) -> Iterator[DataRecord]:
        """Iterate over all records."""
        ...
    
    def close(self) -> None:
        """Close the connection."""
        ...


class BaseConnector(abc.ABC):
    """Base class for data connectors."""
    
    def __init__(self, config: ConnectorConfig):
        self.config = config
        self._connected = False
    
    @property
    @abc.abstractmethod
    def connector_type(self) -> ConnectorType:
        """Get the connector type."""
        pass
    
    @abc.abstractmethod
    def connect(self) -> ConnectionResult:
        """Establish connection to the data source."""
        pass
    
    @abc.abstractmethod
    def list_records(self) -> List[str]:
        """List available record IDs."""
        pass
    
    @abc.abstractmethod
    def get_record(self, record_id: str) -> Optional[DataRecord]:
        """Get a single record by ID."""
        pass
    
    def iter_records(self) -> Iterator[DataRecord]:
        """Iterate over all records."""
        for record_id in self.list_records():
            record = self.get_record(record_id)
            if record is not None:
                yield record
    
    def close(self) -> None:
        """Close the connection."""
        self._connected = False
    
    def _compute_hash(self, data: bytes) -> str:
        """Compute SHA256 hash of data."""
        return hashlib.sha256(data).hexdigest()[:16]
    
    def __enter__(self) -> "BaseConnector":
        self.connect()
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb) -> None:
        self.close()


class LocalFileConnector(BaseConnector):
    """Connector for local filesystem data."""
    
    @property
    def connector_type(self) -> ConnectorType:
        return ConnectorType.LOCAL
    
    def connect(self) -> ConnectionResult:
        """Check that the local path exists."""
        base_path = self.config.endpoint or self.config.prefix or "."
        path = Path(base_path)
        
        if not path.exists():
            return ConnectionResult(
                success=False,
                connector_type=self.connector_type,
                message=f"Path does not exist: {path}",
            )
        
        self._base_path = path
        self._connected = True
        
        # Count records
        record_count = len(list(self._find_records()))
        
        return ConnectionResult(
            success=True,
            connector_type=self.connector_type,
            message=f"Connected to local path: {path}",
            record_count=record_count,
            metadata={"base_path": str(path)},
        )
    
    def _find_records(self) -> Iterator[Path]:
        """Find all record directories or files."""
        if not hasattr(self, "_base_path"):
            return
        
        # Look for manifest.json
        manifest = self._base_path / "manifest.json"
        if manifest.exists():
            with open(manifest) as f:
                data = json.load(f)
                for sample in data.get("samples", []):
                    yield Path(sample.get("sample_id", ""))
                return
        
        # Otherwise scan for subdirectories with metadata.json
        for subdir in self._base_path.iterdir():
            if subdir.is_dir():
                metadata_file = subdir / "metadata.json"
                if metadata_file.exists():
                    yield subdir
    
    def list_records(self) -> List[str]:
        """List all record IDs."""
        return [str(p.name) for p in self._find_records()]
    
    def get_record(self, record_id: str) -> Optional[DataRecord]:
        """Get a record by ID."""
        if not hasattr(self, "_base_path"):
            return None
        
        # Check manifest first
        manifest = self._base_path / "manifest.json"
        if manifest.exists():
            with open(manifest) as f:
                data = json.load(f)
                for sample in data.get("samples", []):
                    if sample.get("sample_id") == record_id:
                        return DataRecord(
                            record_id=record_id,
                            data=sample,
                            source_path=str(manifest),
                            source_hash=self._compute_hash(
                                json.dumps(sample).encode()
                            ),
                        )
            return None
        
        # Check subdirectory
        record_dir = self._base_path / record_id
        if not record_dir.is_dir():
            return None
        
        metadata_file = record_dir / "metadata.json"
        if not metadata_file.exists():
            return None
        
        with open(metadata_file) as f:
            data = json.load(f)
        
        return DataRecord(
            record_id=record_id,
            data=data,
            source_path=str(record_dir),
            source_hash=self._compute_hash(json.dumps(data).encode()),
            metadata={
                "has_document": (record_dir / "document.png").exists()
                or (record_dir / "document.jpg").exists(),
                "has_selfie": (record_dir / "selfie.png").exists()
                or (record_dir / "selfie.jpg").exists(),
            },
        )


class S3Connector(BaseConnector):
    """Connector for AWS S3 object storage."""
    
    @property
    def connector_type(self) -> ConnectorType:
        return ConnectorType.S3
    
    def connect(self) -> ConnectionResult:
        """Connect to S3 bucket."""
        try:
            import boto3
            from botocore.exceptions import ClientError
        except ImportError:
            return ConnectionResult(
                success=False,
                connector_type=self.connector_type,
                message="boto3 not installed. Install with: pip install boto3",
            )
        
        try:
            # Create S3 client
            session_kwargs = {}
            if self.config.access_key and self.config.secret_key:
                session_kwargs["aws_access_key_id"] = self.config.access_key
                session_kwargs["aws_secret_access_key"] = self.config.secret_key
            
            if self.config.endpoint:
                self._s3 = boto3.client(
                    "s3",
                    endpoint_url=self.config.endpoint,
                    **session_kwargs,
                )
            else:
                self._s3 = boto3.client("s3", **session_kwargs)
            
            # Test connection
            self._s3.head_bucket(Bucket=self.config.bucket)
            
            self._connected = True
            return ConnectionResult(
                success=True,
                connector_type=self.connector_type,
                message=f"Connected to S3 bucket: {self.config.bucket}",
                metadata={"bucket": self.config.bucket},
            )
            
        except ClientError as e:
            return ConnectionResult(
                success=False,
                connector_type=self.connector_type,
                message=f"S3 connection failed: {e}",
            )
    
    def list_records(self) -> List[str]:
        """List all record IDs in the bucket."""
        if not self._connected:
            return []
        
        try:
            paginator = self._s3.get_paginator("list_objects_v2")
            prefix = self.config.prefix or ""
            
            record_ids = set()
            for page in paginator.paginate(
                Bucket=self.config.bucket,
                Prefix=prefix,
            ):
                for obj in page.get("Contents", []):
                    key = obj["Key"]
                    # Extract record ID from path
                    if key.endswith("metadata.json"):
                        parts = key.split("/")
                        if len(parts) >= 2:
                            record_ids.add(parts[-2])
            
            return list(record_ids)
            
        except Exception as e:
            logger.error(f"Error listing S3 records: {e}")
            return []
    
    def get_record(self, record_id: str) -> Optional[DataRecord]:
        """Get a record by ID from S3."""
        if not self._connected:
            return None
        
        try:
            prefix = self.config.prefix or ""
            key = f"{prefix}/{record_id}/metadata.json".lstrip("/")
            
            response = self._s3.get_object(
                Bucket=self.config.bucket,
                Key=key,
            )
            
            content = response["Body"].read()
            data = json.loads(content)
            
            return DataRecord(
                record_id=record_id,
                data=data,
                source_path=f"s3://{self.config.bucket}/{key}",
                source_hash=self._compute_hash(content),
            )
            
        except Exception as e:
            logger.error(f"Error getting S3 record {record_id}: {e}")
            return None


class GCSConnector(BaseConnector):
    """Connector for Google Cloud Storage."""
    
    @property
    def connector_type(self) -> ConnectorType:
        return ConnectorType.GCS
    
    def connect(self) -> ConnectionResult:
        """Connect to GCS bucket."""
        try:
            from google.cloud import storage
        except ImportError:
            return ConnectionResult(
                success=False,
                connector_type=self.connector_type,
                message="google-cloud-storage not installed",
            )
        
        try:
            if self.config.credentials_path:
                self._client = storage.Client.from_service_account_json(
                    self.config.credentials_path
                )
            else:
                self._client = storage.Client()
            
            self._bucket = self._client.bucket(self.config.bucket)
            
            # Test connection
            self._bucket.exists()
            
            self._connected = True
            return ConnectionResult(
                success=True,
                connector_type=self.connector_type,
                message=f"Connected to GCS bucket: {self.config.bucket}",
                metadata={"bucket": self.config.bucket},
            )
            
        except Exception as e:
            return ConnectionResult(
                success=False,
                connector_type=self.connector_type,
                message=f"GCS connection failed: {e}",
            )
    
    def list_records(self) -> List[str]:
        """List all record IDs in the bucket."""
        if not self._connected:
            return []
        
        try:
            prefix = self.config.prefix or ""
            blobs = self._client.list_blobs(
                self._bucket,
                prefix=prefix,
            )
            
            record_ids = set()
            for blob in blobs:
                if blob.name.endswith("metadata.json"):
                    parts = blob.name.split("/")
                    if len(parts) >= 2:
                        record_ids.add(parts[-2])
            
            return list(record_ids)
            
        except Exception as e:
            logger.error(f"Error listing GCS records: {e}")
            return []
    
    def get_record(self, record_id: str) -> Optional[DataRecord]:
        """Get a record by ID from GCS."""
        if not self._connected:
            return None
        
        try:
            prefix = self.config.prefix or ""
            key = f"{prefix}/{record_id}/metadata.json".lstrip("/")
            
            blob = self._bucket.blob(key)
            content = blob.download_as_bytes()
            data = json.loads(content)
            
            return DataRecord(
                record_id=record_id,
                data=data,
                source_path=f"gs://{self.config.bucket}/{key}",
                source_hash=self._compute_hash(content),
            )
            
        except Exception as e:
            logger.error(f"Error getting GCS record {record_id}: {e}")
            return None


class HTTPAPIConnector(BaseConnector):
    """Connector for HTTP API data feeds."""
    
    @property
    def connector_type(self) -> ConnectorType:
        return ConnectorType.HTTP_API
    
    def connect(self) -> ConnectionResult:
        """Test connection to API endpoint."""
        try:
            url = self.config.api_url
            if not url:
                return ConnectionResult(
                    success=False,
                    connector_type=self.connector_type,
                    message="api_url not configured",
                )
            
            # Build headers
            headers = dict(self.config.api_headers)
            if self.config.api_key:
                headers["Authorization"] = f"Bearer {self.config.api_key}"
            
            # Test endpoint
            req = urllib.request.Request(
                f"{url}/health",
                headers=headers,
                method="GET",
            )
            
            try:
                with urllib.request.urlopen(
                    req, timeout=self.config.timeout_seconds
                ) as response:
                    if response.status == 200:
                        self._connected = True
                        self._api_url = url
                        self._headers = headers
                        return ConnectionResult(
                            success=True,
                            connector_type=self.connector_type,
                            message=f"Connected to API: {url}",
                        )
            except urllib.error.HTTPError as e:
                # Health endpoint might not exist, try list endpoint
                if e.code == 404:
                    self._connected = True
                    self._api_url = url
                    self._headers = headers
                    return ConnectionResult(
                        success=True,
                        connector_type=self.connector_type,
                        message=f"Connected to API: {url} (health check skipped)",
                    )
                raise
            
            return ConnectionResult(
                success=False,
                connector_type=self.connector_type,
                message="API connection failed",
            )
            
        except Exception as e:
            return ConnectionResult(
                success=False,
                connector_type=self.connector_type,
                message=f"API connection error: {e}",
            )
    
    def list_records(self) -> List[str]:
        """List all record IDs from API."""
        if not self._connected:
            return []
        
        try:
            req = urllib.request.Request(
                f"{self._api_url}/records",
                headers=self._headers,
                method="GET",
            )
            
            with urllib.request.urlopen(
                req, timeout=self.config.timeout_seconds
            ) as response:
                data = json.loads(response.read())
                return [r.get("id", r.get("sample_id", "")) for r in data.get("records", [])]
                
        except Exception as e:
            logger.error(f"Error listing API records: {e}")
            return []
    
    def get_record(self, record_id: str) -> Optional[DataRecord]:
        """Get a record by ID from API."""
        if not self._connected:
            return None
        
        try:
            req = urllib.request.Request(
                f"{self._api_url}/records/{record_id}",
                headers=self._headers,
                method="GET",
            )
            
            with urllib.request.urlopen(
                req, timeout=self.config.timeout_seconds
            ) as response:
                content = response.read()
                data = json.loads(content)
                
                return DataRecord(
                    record_id=record_id,
                    data=data,
                    source_path=f"{self._api_url}/records/{record_id}",
                    source_hash=self._compute_hash(content),
                )
                
        except Exception as e:
            logger.error(f"Error getting API record {record_id}: {e}")
            return None


class ConnectorRegistry:
    """Registry for data connectors."""
    
    _connectors: Dict[ConnectorType, Type[BaseConnector]] = {
        ConnectorType.LOCAL: LocalFileConnector,
        ConnectorType.S3: S3Connector,
        ConnectorType.GCS: GCSConnector,
        ConnectorType.HTTP_API: HTTPAPIConnector,
    }
    
    @classmethod
    def register(
        cls,
        connector_type: ConnectorType,
        connector_class: Type[BaseConnector],
    ) -> None:
        """Register a custom connector."""
        cls._connectors[connector_type] = connector_class
    
    @classmethod
    def get_connector(cls, config: ConnectorConfig) -> BaseConnector:
        """Get a connector instance for the given configuration."""
        connector_class = cls._connectors.get(config.connector_type)
        
        if connector_class is None:
            raise ValueError(f"Unknown connector type: {config.connector_type}")
        
        return connector_class(config)
    
    @classmethod
    def list_connectors(cls) -> List[ConnectorType]:
        """List all registered connector types."""
        return list(cls._connectors.keys())
    
    @classmethod
    def from_uri(cls, uri: str, **kwargs) -> BaseConnector:
        """
        Create a connector from a URI string.
        
        Supported URI formats:
        - file:///path/to/data or /path/to/data
        - s3://bucket/prefix
        - gs://bucket/prefix
        - https://api.example.com/v1
        """
        if uri.startswith("s3://"):
            parts = uri[5:].split("/", 1)
            bucket = parts[0]
            prefix = parts[1] if len(parts) > 1 else ""
            config = ConnectorConfig(
                connector_type=ConnectorType.S3,
                bucket=bucket,
                prefix=prefix,
                **kwargs,
            )
        elif uri.startswith("gs://"):
            parts = uri[5:].split("/", 1)
            bucket = parts[0]
            prefix = parts[1] if len(parts) > 1 else ""
            config = ConnectorConfig(
                connector_type=ConnectorType.GCS,
                bucket=bucket,
                prefix=prefix,
                **kwargs,
            )
        elif uri.startswith("http://") or uri.startswith("https://"):
            config = ConnectorConfig(
                connector_type=ConnectorType.HTTP_API,
                api_url=uri,
                **kwargs,
            )
        elif uri.startswith("file://"):
            config = ConnectorConfig(
                connector_type=ConnectorType.LOCAL,
                endpoint=uri[7:],
                **kwargs,
            )
        else:
            # Assume local path
            config = ConnectorConfig(
                connector_type=ConnectorType.LOCAL,
                endpoint=uri,
                **kwargs,
            )
        
        return cls.get_connector(config)
