"""
VirtEngine Facial Verification Pipeline

This module provides deterministic facial verification for the VEID identity
verification system. It supports:
- Face detection and alignment
- Face embedding extraction
- Verification against ID documents and enrolled references
- Configurable decision thresholds
- Deterministic execution for blockchain consensus

Usage:
    from ml.facial_verification import FaceVerifier, VerificationConfig
    
    config = VerificationConfig(match_threshold=0.90)
    verifier = FaceVerifier(config)
    result = verifier.verify(selfie_image, id_document_face)
"""

from ml.facial_verification.config import (
    VerificationConfig,
    PreprocessingConfig,
    DetectionConfig,
    DeterminismConfig,
)
from ml.facial_verification.verification import (
    FaceVerifier,
    VerificationResult,
)
from ml.facial_verification.preprocessing import FacePreprocessor
from ml.facial_verification.face_detection import FaceDetector, FaceDetection
from ml.facial_verification.embeddings import FaceEmbedder
from ml.facial_verification.determinism import DeterminismController
from ml.facial_verification.reason_codes import ReasonCodes
from ml.facial_verification.mtcnn_detector import (
    MTCNNDetector,
    MTCNNConfig,
    MTCNNDetection,
    MTCNNDetectionResult,
    FiveLandmarks,
    FaceAligner,
)

__version__ = "1.0.0"
__all__ = [
    "VerificationConfig",
    "PreprocessingConfig",
    "DetectionConfig",
    "DeterminismConfig",
    "FaceVerifier",
    "VerificationResult",
    "FacePreprocessor",
    "FaceDetector",
    "FaceDetection",
    "FaceEmbedder",
    "DeterminismController",
    "ReasonCodes",
    # MTCNN Integration (VE-903)
    "MTCNNDetector",
    "MTCNNConfig",
    "MTCNNDetection",
    "MTCNNDetectionResult",
    "FiveLandmarks",
    "FaceAligner",
]
