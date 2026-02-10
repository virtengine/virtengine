"""
OCR feature extraction for trust score training.

Extracts features from OCR results including field confidences,
validation status, and text quality metrics.
"""

import logging
from dataclasses import dataclass, field
from typing import List, Optional, Dict, Any

import numpy as np

from ml.training.config import FeatureConfig

logger = logging.getLogger(__name__)


@dataclass
class FieldScore:
    """Score and metadata for a single OCR field."""
    
    field_name: str
    extracted: bool = False
    confidence: float = 0.0
    validated: bool = False
    validation_score: float = 0.0
    character_count: int = 0
    
    def to_vector(self) -> np.ndarray:
        """Convert to feature vector."""
        return np.array([
            float(self.extracted),
            self.confidence,
            float(self.validated),
            self.validation_score,
            min(1.0, self.character_count / 50.0),  # Normalize
        ], dtype=np.float32)


@dataclass
class OCRFeatures:
    """OCR-based features for a sample."""
    
    # Per-field scores
    field_scores: Dict[str, FieldScore] = field(default_factory=dict)
    
    # Aggregate metrics
    overall_ocr_confidence: float = 0.0
    fields_extracted_ratio: float = 0.0
    fields_validated_ratio: float = 0.0
    
    # Text quality
    average_character_confidence: float = 0.0
    text_density: float = 0.0
    
    # OCR success indicators
    ocr_success: bool = False
    ocr_timeout: bool = False
    
    # Specific field extractions
    has_name: bool = False
    has_dob: bool = False
    has_doc_number: bool = False
    has_expiry: bool = False
    has_nationality: bool = False
    
    def to_vector(self, expected_fields: List[str]) -> np.ndarray:
        """Convert to feature vector."""
        features = []
        
        # Per-field features
        for field_name in expected_fields:
            if field_name in self.field_scores:
                features.extend(self.field_scores[field_name].to_vector())
            else:
                # Missing field
                features.extend([0.0, 0.0, 0.0, 0.0, 0.0])
        
        # Aggregate features
        features.extend([
            self.overall_ocr_confidence,
            self.fields_extracted_ratio,
            self.fields_validated_ratio,
            self.average_character_confidence,
            self.text_density,
            float(self.ocr_success),
            float(not self.ocr_timeout),
            float(self.has_name),
            float(self.has_dob),
            float(self.has_doc_number),
            float(self.has_expiry),
            float(self.has_nationality),
        ])
        
        return np.array(features, dtype=np.float32)


class OCRFeatureExtractor:
    """
    Extracts OCR-based features from document images.
    
    Uses the OCR extraction pipeline to:
    - Extract text fields from documents
    - Compute field-level confidence scores
    - Validate extracted values
    - Compute aggregate OCR quality metrics
    """
    
    def __init__(self, config: Optional[FeatureConfig] = None):
        """
        Initialize the OCR feature extractor.
        
        Args:
            config: Feature configuration
        """
        self.config = config or FeatureConfig()
        self._expected_fields = self.config.ocr_fields
        
        # Initialize OCR pipeline
        self._ocr_pipeline = None
        try:
            from ml.ocr_extraction import OCRExtractionPipeline, OCRExtractionConfig
            ocr_config = OCRExtractionConfig()
            self._ocr_pipeline = OCRExtractionPipeline(ocr_config)
        except ImportError:
            logger.warning(
                "OCR extraction pipeline not available, "
                "OCR features will be empty"
            )
        
        # Initialize text detection for ROI extraction
        self._text_detector = None
        try:
            from ml.text_detection import TextDetectionPipeline
            self._text_detector = TextDetectionPipeline()
        except ImportError:
            logger.warning(
                "Text detection pipeline not available, "
                "using default ROIs"
            )
    
    def extract(
        self,
        document_image: Optional[np.ndarray],
        document_type: str = "id_card",
    ) -> OCRFeatures:
        """
        Extract OCR features from a document image.
        
        Args:
            document_image: Preprocessed document image
            document_type: Type of document for field parsing
            
        Returns:
            OCRFeatures containing field scores and metrics
        """
        features = OCRFeatures()
        
        if document_image is None:
            return features
        
        # Denormalize image if needed
        image = self._denormalize_image(document_image)
        
        # Detect text regions
        rois = []
        if self._text_detector is not None:
            try:
                detection_result = self._text_detector.detect(image)
                rois = detection_result.rois
            except Exception as e:
                logger.debug(f"Text detection failed: {e}")
        
        # Extract OCR
        if self._ocr_pipeline is not None:
            try:
                ocr_result = self._ocr_pipeline.extract(
                    image, rois, document_type=document_type
                )
                features = self._process_ocr_result(ocr_result)
            except Exception as e:
                logger.debug(f"OCR extraction failed: {e}")
                features.ocr_timeout = True
        
        return features
    
    def _process_ocr_result(self, ocr_result) -> OCRFeatures:
        """Process OCR extraction result into features."""
        features = OCRFeatures(ocr_success=ocr_result.success)
        
        # Process each field
        extracted_count = 0
        validated_count = 0
        confidence_sum = 0.0
        
        for field_name, parsed_field in ocr_result.fields.items():
            field_score = FieldScore(
                field_name=field_name,
                extracted=parsed_field.value is not None and len(str(parsed_field.value)) > 0,
                confidence=parsed_field.confidence,
                validated=parsed_field.validation_status.name == "VALID",
                validation_score=1.0 if parsed_field.validation_status.name == "VALID" else 0.0,
                character_count=len(str(parsed_field.value)) if parsed_field.value else 0,
            )
            features.field_scores[field_name] = field_score
            
            if field_score.extracted:
                extracted_count += 1
                confidence_sum += field_score.confidence
                
                if field_score.validated:
                    validated_count += 1
        
        # Compute aggregate metrics
        total_fields = len(self._expected_fields)
        if total_fields > 0:
            features.fields_extracted_ratio = extracted_count / total_fields
            features.fields_validated_ratio = validated_count / total_fields
        
        if extracted_count > 0:
            features.overall_ocr_confidence = confidence_sum / extracted_count
        
        # Check for specific fields
        features.has_name = self._has_field(features, ["name", "full_name", "surname"])
        features.has_dob = self._has_field(features, ["date_of_birth", "dob", "birth_date"])
        features.has_doc_number = self._has_field(features, ["document_number", "doc_number", "id_number"])
        features.has_expiry = self._has_field(features, ["expiry_date", "expiration", "valid_until"])
        features.has_nationality = self._has_field(features, ["nationality", "country", "citizenship"])
        
        # Character confidence and text density
        features.average_character_confidence = features.overall_ocr_confidence
        features.text_density = features.fields_extracted_ratio
        
        return features
    
    def _has_field(
        self,
        features: OCRFeatures,
        field_names: List[str]
    ) -> bool:
        """Check if any of the given fields was extracted."""
        for name in field_names:
            if name in features.field_scores:
                if features.field_scores[name].extracted:
                    return True
        return False
    
    def _denormalize_image(self, image: np.ndarray) -> np.ndarray:
        """Convert normalized image back to uint8."""
        if image.dtype == np.float32:
            mean = np.array([0.485, 0.456, 0.406])
            std = np.array([0.229, 0.224, 0.225])
            denorm = (image * std + mean) * 255
            return np.clip(denorm, 0, 255).astype(np.uint8)
        return image
    
    def get_feature_dim(self) -> int:
        """Get the dimension of the output feature vector."""
        # 5 features per field + 12 aggregate features
        return len(self._expected_fields) * 5 + 12
