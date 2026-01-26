"""
Metadata feature extraction for trust score training.

Extracts features from device, session, and capture metadata
to detect anomalies and assess trust.
"""

import logging
from dataclasses import dataclass, field
from typing import List, Optional, Dict, Any

import numpy as np

from ml.training.config import FeatureConfig
from ml.training.dataset.ingestion import CaptureMetadata

logger = logging.getLogger(__name__)


# Known device types for encoding
DEVICE_TYPES = ["mobile", "tablet", "desktop", "camera", "scanner", "unknown"]

# Known camera types
CAMERA_FACING = ["front", "back", "unknown"]

# Known operating systems
OS_TYPES = ["android", "ios", "windows", "macos", "linux", "unknown"]


@dataclass
class MetadataFeatures:
    """Metadata-based features for a sample."""
    
    # Device features
    device_type_encoding: np.ndarray = field(
        default_factory=lambda: np.zeros(len(DEVICE_TYPES))
    )
    os_type_encoding: np.ndarray = field(
        default_factory=lambda: np.zeros(len(OS_TYPES))
    )
    camera_facing_encoding: np.ndarray = field(
        default_factory=lambda: np.zeros(len(CAMERA_FACING))
    )
    
    # Capture quality indicators
    light_level_score: float = 0.5  # 0-1, 0.5 = optimal
    motion_detected: bool = False
    
    # Session features
    gps_available: bool = False
    app_version_known: bool = False
    
    # Temporal features
    capture_hour: float = 0.0       # 0-1 (normalized hour)
    capture_day_of_week: float = 0.0  # 0-1 (normalized day)
    
    # Anomaly indicators
    unusual_device: bool = False
    unusual_time: bool = False
    metadata_complete: bool = False
    
    def to_vector(self) -> np.ndarray:
        """Convert to feature vector."""
        features = []
        
        # Device type one-hot encoding
        features.extend(self.device_type_encoding)
        
        # OS type one-hot encoding
        features.extend(self.os_type_encoding)
        
        # Camera facing one-hot encoding
        features.extend(self.camera_facing_encoding)
        
        # Scalar features
        features.extend([
            self.light_level_score,
            float(self.motion_detected),
            float(self.gps_available),
            float(self.app_version_known),
            self.capture_hour,
            self.capture_day_of_week,
            float(not self.unusual_device),  # Normal device = 1
            float(not self.unusual_time),     # Normal time = 1
            float(self.metadata_complete),
        ])
        
        return np.array(features, dtype=np.float32)


class MetadataFeatureExtractor:
    """
    Extracts features from capture and device metadata.
    
    Analyzes:
    - Device type and platform
    - Capture conditions (lighting, motion)
    - Session metadata (GPS, app version)
    - Temporal patterns
    - Anomaly detection
    """
    
    def __init__(self, config: Optional[FeatureConfig] = None):
        """
        Initialize the metadata feature extractor.
        
        Args:
            config: Feature configuration
        """
        self.config = config or FeatureConfig()
        
        # Statistics for anomaly detection (would be computed from training data)
        self._normal_capture_hours = (7, 22)  # 7 AM to 10 PM
        self._known_devices = set()
        self._known_app_versions = set()
    
    def extract(
        self,
        capture_metadata: Optional[CaptureMetadata],
    ) -> MetadataFeatures:
        """
        Extract metadata features from capture metadata.
        
        Args:
            capture_metadata: Capture session metadata
            
        Returns:
            MetadataFeatures containing encoded metadata
        """
        features = MetadataFeatures()
        
        if capture_metadata is None:
            return features
        
        # Device type encoding
        device_type = (capture_metadata.device_type or "unknown").lower()
        features.device_type_encoding = self._encode_device_type(device_type)
        
        # OS type encoding
        os_version = (capture_metadata.os_version or "unknown").lower()
        features.os_type_encoding = self._encode_os_type(os_version)
        
        # Camera facing encoding
        camera_facing = (capture_metadata.camera_facing or "unknown").lower()
        features.camera_facing_encoding = self._encode_camera_facing(camera_facing)
        
        # Light level
        if capture_metadata.light_level is not None:
            features.light_level_score = self._normalize_light_level(
                capture_metadata.light_level
            )
        
        # Motion detection
        features.motion_detected = capture_metadata.motion_detected or False
        
        # GPS availability
        features.gps_available = capture_metadata.gps_available or False
        
        # App version
        features.app_version_known = capture_metadata.app_version is not None
        if capture_metadata.app_version:
            self._known_app_versions.add(capture_metadata.app_version)
        
        # Temporal features
        if capture_metadata.capture_timestamp is not None:
            features.capture_hour, features.capture_day_of_week = \
                self._extract_temporal_features(capture_metadata.capture_timestamp)
        
        # Anomaly detection
        features.unusual_device = self._is_unusual_device(capture_metadata)
        features.unusual_time = self._is_unusual_time(capture_metadata)
        
        # Metadata completeness
        features.metadata_complete = self._check_completeness(capture_metadata)
        
        return features
    
    def _encode_device_type(self, device_type: str) -> np.ndarray:
        """One-hot encode device type."""
        encoding = np.zeros(len(DEVICE_TYPES), dtype=np.float32)
        
        for i, dtype in enumerate(DEVICE_TYPES):
            if dtype in device_type:
                encoding[i] = 1.0
                return encoding
        
        # Unknown
        encoding[-1] = 1.0
        return encoding
    
    def _encode_os_type(self, os_version: str) -> np.ndarray:
        """One-hot encode OS type."""
        encoding = np.zeros(len(OS_TYPES), dtype=np.float32)
        
        for i, os_type in enumerate(OS_TYPES):
            if os_type in os_version:
                encoding[i] = 1.0
                return encoding
        
        # Unknown
        encoding[-1] = 1.0
        return encoding
    
    def _encode_camera_facing(self, camera_facing: str) -> np.ndarray:
        """One-hot encode camera facing direction."""
        encoding = np.zeros(len(CAMERA_FACING), dtype=np.float32)
        
        for i, facing in enumerate(CAMERA_FACING):
            if facing in camera_facing:
                encoding[i] = 1.0
                return encoding
        
        # Unknown
        encoding[-1] = 1.0
        return encoding
    
    def _normalize_light_level(self, light_level: float) -> float:
        """
        Normalize light level to 0-1 score.
        
        Score of 0.5 is optimal (good lighting).
        Score near 0 is too dark, near 1 is too bright.
        """
        # Assume light_level is in lux (0-100000+)
        # Optimal range is roughly 100-1000 lux
        if light_level < 100:
            return light_level / 200.0  # Too dark
        elif light_level > 1000:
            return 1.0 - min(0.5, (light_level - 1000) / 2000.0)  # Too bright
        else:
            return 0.5  # Optimal
    
    def _extract_temporal_features(
        self,
        timestamp: float
    ) -> tuple:
        """Extract normalized hour and day of week from timestamp."""
        import datetime
        
        try:
            dt = datetime.datetime.fromtimestamp(timestamp)
            hour_normalized = dt.hour / 24.0
            day_normalized = dt.weekday() / 7.0
            return hour_normalized, day_normalized
        except Exception:
            return 0.5, 0.5  # Default to midpoint
    
    def _is_unusual_device(self, metadata: CaptureMetadata) -> bool:
        """Check if device seems unusual."""
        # No device info at all is suspicious
        if not metadata.device_type and not metadata.device_model:
            return True
        
        # Desktop/scanner for selfie is unusual
        device_type = (metadata.device_type or "").lower()
        if device_type in ["desktop", "scanner"]:
            return True
        
        return False
    
    def _is_unusual_time(self, metadata: CaptureMetadata) -> bool:
        """Check if capture time is unusual."""
        if metadata.capture_timestamp is None:
            return False
        
        try:
            import datetime
            dt = datetime.datetime.fromtimestamp(metadata.capture_timestamp)
            hour = dt.hour
            
            # Late night/early morning captures might be suspicious
            min_hour, max_hour = self._normal_capture_hours
            return hour < min_hour or hour > max_hour
        except Exception:
            return False
    
    def _check_completeness(self, metadata: CaptureMetadata) -> bool:
        """Check if metadata is reasonably complete."""
        required_fields = [
            metadata.device_type,
            metadata.os_version,
            metadata.capture_timestamp,
        ]
        
        return all(field is not None for field in required_fields)
    
    def get_feature_dim(self) -> int:
        """Get the dimension of the output feature vector."""
        # Device types + OS types + Camera facing + 9 scalar features
        return len(DEVICE_TYPES) + len(OS_TYPES) + len(CAMERA_FACING) + 9
