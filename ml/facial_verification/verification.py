"""
Face verification module.

This module provides the main verification logic, combining preprocessing,
detection, embedding extraction, and decision making.
"""

import logging
import hashlib
import time
from dataclasses import dataclass, field
from typing import List, Optional, Tuple

import numpy as np

from ml.facial_verification.config import VerificationConfig
from ml.facial_verification.preprocessing import FacePreprocessor, PreprocessingResult
from ml.facial_verification.face_detection import FaceDetector, DetectionResult
from ml.facial_verification.embeddings import FaceEmbedder, EmbeddingResult
from ml.facial_verification.determinism import DeterminismController
from ml.facial_verification.reason_codes import ReasonCodes

logger = logging.getLogger(__name__)


@dataclass
class VerificationResult:
    """Complete result of face verification."""
    
    # Decision
    match: bool
    decision: str  # "match", "no_match", "borderline"
    
    # Scores
    similarity_score: float  # 0.0 - 1.0
    confidence_percent: float  # 0 - 100
    
    # Model info
    model_name: str
    model_version: str
    model_hash: str
    
    # Reason codes
    reason_codes: List[str] = field(default_factory=list)
    
    # Hashes for consensus
    embeddings_hash: str = ""
    result_hash: str = ""
    
    # Timing
    processing_time_ms: float = 0.0
    
    # Component results (optional)
    preprocessing_result: Optional[dict] = None
    detection_result: Optional[dict] = None
    embedding_result: Optional[dict] = None
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "match": self.match,
            "decision": self.decision,
            "similarity_score": self.similarity_score,
            "confidence_percent": self.confidence_percent,
            "model_name": self.model_name,
            "model_version": self.model_version,
            "model_hash": self.model_hash,
            "reason_codes": self.reason_codes,
            "embeddings_hash": self.embeddings_hash,
            "result_hash": self.result_hash,
            "processing_time_ms": self.processing_time_ms,
        }
    
    def to_verification_record(self) -> dict:
        """Convert to verification record for on-chain storage."""
        return {
            "match": self.match,
            "decision": self.decision,
            "similarity_score": round(self.similarity_score, 4),
            "confidence_percent": round(self.confidence_percent, 2),
            "model_name": self.model_name,
            "model_version": self.model_version,
            "model_hash": self.model_hash[:16],  # Truncated for storage
            "reason_codes": self.reason_codes[:5],  # Limit reason codes
            "embeddings_hash": self.embeddings_hash[:16],  # Truncated
            "result_hash": self.result_hash,
        }


class FaceVerifier:
    """
    Face verification pipeline.
    
    This class orchestrates the complete face verification process:
    1. Preprocessing of probe and reference images
    2. Face detection and alignment
    3. Embedding extraction
    4. Similarity computation
    5. Decision making with reason codes
    """
    
    def __init__(self, config: Optional[VerificationConfig] = None):
        """
        Initialize the face verifier.
        
        Args:
            config: Verification configuration. Uses defaults if not provided.
        """
        self.config = config or VerificationConfig()
        self._preprocessor = FacePreprocessor(self.config)
        self._detector = FaceDetector(self.config)
        self._embedder = FaceEmbedder(self.config)
        self._determinism = DeterminismController(self.config)
    
    def verify(
        self,
        probe_image: np.ndarray,
        reference_image: np.ndarray,
        include_details: bool = False
    ) -> VerificationResult:
        """
        Verify that a probe image matches a reference image.
        
        Args:
            probe_image: Selfie or live capture image
            reference_image: ID document face or enrolled reference
            include_details: Whether to include detailed component results
            
        Returns:
            VerificationResult with match decision and scores
        """
        start_time = time.time()
        reason_codes = []
        
        # Ensure deterministic execution
        self._determinism.ensure_deterministic()
        
        # Process probe image
        probe_result = self._process_image(probe_image, "probe")
        if probe_result is None:
            return self._create_failure_result(
                reason_codes=[ReasonCodes.NO_FACE_DETECTED.value],
                processing_time_ms=(time.time() - start_time) * 1000
            )
        
        probe_face, probe_embedding, probe_codes = probe_result
        reason_codes.extend(probe_codes)
        
        # Process reference image
        ref_result = self._process_image(reference_image, "reference")
        if ref_result is None:
            return self._create_failure_result(
                reason_codes=[ReasonCodes.NO_FACE_DETECTED.value],
                processing_time_ms=(time.time() - start_time) * 1000
            )
        
        ref_face, ref_embedding, ref_codes = ref_result
        reason_codes.extend(ref_codes)
        
        # Compute similarity
        similarity = self._embedder.compute_similarity(
            probe_embedding,
            ref_embedding
        )
        
        # Make decision
        decision, decision_codes = self._make_decision(similarity)
        reason_codes.extend(decision_codes)
        
        # Compute hashes for consensus
        embeddings_hash = self._compute_embeddings_hash(probe_embedding, ref_embedding)
        
        processing_time_ms = (time.time() - start_time) * 1000
        
        result = VerificationResult(
            match=(decision == "match"),
            decision=decision,
            similarity_score=similarity,
            confidence_percent=similarity * 100,
            model_name=self.config.model_name.value,
            model_version=self.config.model_version,
            model_hash=self._embedder.model_hash,
            reason_codes=[c.value if isinstance(c, ReasonCodes) else c for c in reason_codes],
            embeddings_hash=embeddings_hash,
            processing_time_ms=processing_time_ms
        )
        
        # Compute result hash
        result.result_hash = self._determinism.compute_result_hash(result)
        
        return result
    
    def verify_batch(
        self,
        probe_image: np.ndarray,
        reference_images: List[np.ndarray]
    ) -> VerificationResult:
        """
        Verify a probe image against multiple reference images.
        
        Uses the best match among all references.
        
        Args:
            probe_image: Selfie or live capture image
            reference_images: List of reference images to match against
            
        Returns:
            VerificationResult with best match decision
        """
        if not reference_images:
            return self._create_failure_result(
                reason_codes=[ReasonCodes.INVALID_IMAGE_FORMAT.value],
                processing_time_ms=0.0
            )
        
        start_time = time.time()
        
        # Ensure deterministic execution
        self._determinism.ensure_deterministic()
        
        # Process probe image once
        probe_result = self._process_image(probe_image, "probe")
        if probe_result is None:
            return self._create_failure_result(
                reason_codes=[ReasonCodes.NO_FACE_DETECTED.value],
                processing_time_ms=(time.time() - start_time) * 1000
            )
        
        probe_face, probe_embedding, probe_codes = probe_result
        
        # Process all reference images and find best match
        best_similarity = 0.0
        best_ref_embedding = None
        all_reason_codes = list(probe_codes)
        
        for i, ref_image in enumerate(reference_images):
            ref_result = self._process_image(ref_image, f"reference_{i}")
            if ref_result is None:
                continue
            
            ref_face, ref_embedding, ref_codes = ref_result
            
            similarity = self._embedder.compute_similarity(
                probe_embedding,
                ref_embedding
            )
            
            if similarity > best_similarity:
                best_similarity = similarity
                best_ref_embedding = ref_embedding
        
        if best_ref_embedding is None:
            return self._create_failure_result(
                reason_codes=[ReasonCodes.NO_FACE_DETECTED.value],
                processing_time_ms=(time.time() - start_time) * 1000
            )
        
        # Make decision based on best match
        decision, decision_codes = self._make_decision(best_similarity)
        all_reason_codes.extend(decision_codes)
        
        # Compute hashes
        embeddings_hash = self._compute_embeddings_hash(probe_embedding, best_ref_embedding)
        
        processing_time_ms = (time.time() - start_time) * 1000
        
        result = VerificationResult(
            match=(decision == "match"),
            decision=decision,
            similarity_score=best_similarity,
            confidence_percent=best_similarity * 100,
            model_name=self.config.model_name.value,
            model_version=self.config.model_version,
            model_hash=self._embedder.model_hash,
            reason_codes=[c.value if isinstance(c, ReasonCodes) else c for c in all_reason_codes],
            embeddings_hash=embeddings_hash,
            processing_time_ms=processing_time_ms
        )
        
        result.result_hash = self._determinism.compute_result_hash(result)
        
        return result
    
    def _process_image(
        self, 
        image: np.ndarray, 
        image_type: str
    ) -> Optional[Tuple[np.ndarray, np.ndarray, List[ReasonCodes]]]:
        """
        Process an image through preprocessing, detection, and embedding.
        
        Args:
            image: Input image
            image_type: Type of image for logging
            
        Returns:
            Tuple of (face image, embedding, reason codes) or None on failure
        """
        reason_codes = []
        
        # Check image quality
        quality_score, quality_issues = self._preprocessor.check_image_quality(image)
        reason_codes.extend(quality_issues)
        
        if quality_score < self.config.min_image_quality_score:
            logger.warning(f"{image_type}: Low quality score {quality_score:.2f}")
            # Continue anyway, but record the issue
        
        # Detect and align face
        face_image, detection_result = self._detector.detect_and_align(
            image,
            target_size=self.config.preprocessing.target_resolution
        )
        
        if face_image is None:
            logger.warning(f"{image_type}: Face detection failed")
            if detection_result.error_code:
                reason_codes.append(detection_result.error_code)
            return None
        
        # Extract embedding
        embedding_result = self._embedder.embed(face_image)
        
        if not embedding_result.success:
            logger.warning(f"{image_type}: Embedding extraction failed")
            if embedding_result.error_code:
                reason_codes.append(embedding_result.error_code)
            return None
        
        return face_image, embedding_result.embedding, reason_codes
    
    def _make_decision(
        self, 
        similarity: float
    ) -> Tuple[str, List[ReasonCodes]]:
        """
        Make verification decision based on similarity score.
        
        Args:
            similarity: Similarity score (0-1)
            
        Returns:
            Tuple of (decision string, reason codes)
        """
        reason_codes = []
        
        if similarity >= self.config.match_threshold:
            decision = "match"
            reason_codes.append(ReasonCodes.MATCH_CONFIRMED)
            
            if similarity >= 0.95:
                reason_codes.append(ReasonCodes.HIGH_CONFIDENCE_MATCH)
        
        elif similarity >= self.config.borderline_lower:
            decision = "borderline"
            reason_codes.append(ReasonCodes.BORDERLINE_MATCH)
        
        elif similarity >= self.config.reject_threshold:
            decision = "no_match"
            reason_codes.append(ReasonCodes.LOW_SIMILARITY_SCORE)
        
        else:
            decision = "no_match"
            reason_codes.append(ReasonCodes.EMBEDDING_MISMATCH)
        
        return decision, reason_codes
    
    def _compute_embeddings_hash(
        self, 
        embedding1: np.ndarray, 
        embedding2: np.ndarray
    ) -> str:
        """
        Compute combined hash of both embeddings.
        
        Args:
            embedding1: First embedding
            embedding2: Second embedding
            
        Returns:
            SHA256 hash of concatenated embeddings
        """
        # Round to reduce floating-point precision issues
        e1_rounded = np.round(embedding1, decimals=6)
        e2_rounded = np.round(embedding2, decimals=6)
        
        combined = np.concatenate([e1_rounded, e2_rounded])
        return hashlib.sha256(combined.tobytes()).hexdigest()
    
    def _create_failure_result(
        self,
        reason_codes: List[str],
        processing_time_ms: float
    ) -> VerificationResult:
        """Create a failure verification result."""
        return VerificationResult(
            match=False,
            decision="no_match",
            similarity_score=0.0,
            confidence_percent=0.0,
            model_name=self.config.model_name.value,
            model_version=self.config.model_version,
            model_hash=self._embedder.model_hash,
            reason_codes=reason_codes,
            embeddings_hash="",
            result_hash="",
            processing_time_ms=processing_time_ms
        )
    
    def get_model_info(self) -> dict:
        """Get information about the loaded model."""
        return {
            "model_name": self.config.model_name.value,
            "model_version": self.config.model_version,
            "model_hash": self._embedder.model_hash,
            "embedding_dimension": self._embedder.embedding_dimension,
            "distance_metric": self.config.distance_metric.value,
            "thresholds": {
                "match": self.config.match_threshold,
                "borderline": self.config.borderline_lower,
                "reject": self.config.reject_threshold,
            },
        }


class VerificationDecisionEngine:
    """
    Decision engine with configurable thresholds and reason codes.
    
    This class provides advanced decision-making capabilities with
    support for multiple threshold levels and detailed reason codes.
    """
    
    def __init__(self, config: Optional[VerificationConfig] = None):
        """Initialize the decision engine."""
        self.config = config or VerificationConfig()
    
    def evaluate(
        self,
        similarity: float,
        probe_quality: float = 1.0,
        reference_quality: float = 1.0
    ) -> Tuple[str, float, List[ReasonCodes]]:
        """
        Evaluate similarity score and return decision.
        
        Args:
            similarity: Raw similarity score
            probe_quality: Quality score of probe image (0-1)
            reference_quality: Quality score of reference image (0-1)
            
        Returns:
            Tuple of (decision, adjusted_confidence, reason_codes)
        """
        reason_codes = []
        
        # Adjust confidence based on image quality
        quality_factor = (probe_quality + reference_quality) / 2
        adjusted_confidence = similarity * quality_factor
        
        # Make decision
        if similarity >= self.config.match_threshold:
            if quality_factor < 0.5:
                # Low quality might affect reliability
                decision = "borderline"
                reason_codes.append(ReasonCodes.LOW_QUALITY_IMAGE)
                reason_codes.append(ReasonCodes.BORDERLINE_MATCH)
            else:
                decision = "match"
                reason_codes.append(ReasonCodes.MATCH_CONFIRMED)
        
        elif similarity >= self.config.borderline_lower:
            decision = "borderline"
            reason_codes.append(ReasonCodes.BORDERLINE_MATCH)
        
        else:
            decision = "no_match"
            if similarity < self.config.reject_threshold:
                reason_codes.append(ReasonCodes.EMBEDDING_MISMATCH)
            else:
                reason_codes.append(ReasonCodes.LOW_SIMILARITY_SCORE)
        
        return decision, adjusted_confidence, reason_codes
    
    def get_threshold_info(self) -> dict:
        """Get information about configured thresholds."""
        return {
            "match_threshold": self.config.match_threshold,
            "match_description": f"Similarity >= {self.config.match_threshold:.0%} is a match",
            "borderline_lower": self.config.borderline_lower,
            "borderline_description": f"Similarity {self.config.borderline_lower:.0%}-{self.config.match_threshold:.0%} is borderline",
            "reject_threshold": self.config.reject_threshold,
            "reject_description": f"Similarity < {self.config.reject_threshold:.0%} is rejected",
        }
