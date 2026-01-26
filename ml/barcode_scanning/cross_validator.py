"""
Cross-Validation Module for Barcode vs OCR Data.

This module provides cross-validation between barcode/MRZ data
and OCR-extracted data to ensure consistency and detect fraud.

The validation compares corresponding fields from both sources
and produces a combined validation score.

Security Note:
    Comparison is done in memory only.
    No PII is logged during validation.
"""

import re
import logging
import hashlib
from dataclasses import dataclass, field
from typing import Dict, List, Optional, Any, Tuple, Union
from enum import Enum
from datetime import datetime

from ml.barcode_scanning.config import CrossValidationConfig
from ml.barcode_scanning.pdf417_parser import PDF417Data
from ml.barcode_scanning.mrz_parser import MRZData

# Intentionally minimal logging - no PII
logger = logging.getLogger(__name__)


class MatchType(str, Enum):
    """Type of field match result."""
    EXACT = "exact"           # Fields match exactly
    FUZZY = "fuzzy"           # Fields match with minor differences
    PARTIAL = "partial"       # Fields partially match
    MISMATCH = "mismatch"     # Fields do not match
    MISSING_BARCODE = "missing_barcode"  # Field not in barcode
    MISSING_OCR = "missing_ocr"          # Field not in OCR
    NOT_COMPARED = "not_compared"        # Field not in comparison set


@dataclass
class FieldMatch:
    """Result of comparing a single field."""
    field_name: str
    barcode_value: str
    ocr_value: str
    match_type: MatchType
    similarity_score: float  # 0.0 - 1.0
    weight: float           # Field importance weight
    contribution: float     # Weighted contribution to total score
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary (without raw values for privacy)."""
        return {
            "field_name": self.field_name,
            "match_type": self.match_type.value,
            "similarity_score": self.similarity_score,
            "weight": self.weight,
            "contribution": self.contribution,
            # Note: barcode_value and ocr_value intentionally omitted
        }


@dataclass
class CrossValidationResult:
    """
    Result of cross-validation between barcode and OCR data.
    """
    # Overall results
    score: float = 0.0                # 0.0 - 1.0 overall score
    is_valid: bool = False            # Whether validation passed
    confidence: float = 0.0           # Confidence in the result
    
    # Field-level results
    field_matches: List[FieldMatch] = field(default_factory=list)
    
    # Statistics
    total_fields_compared: int = 0
    exact_matches: int = 0
    fuzzy_matches: int = 0
    mismatches: int = 0
    missing_fields: int = 0
    
    # Required field status
    required_fields_matched: bool = True
    required_fields_missing: List[str] = field(default_factory=list)
    
    # Error handling
    success: bool = True
    error_message: Optional[str] = None
    
    # VEID scoring contribution
    veid_score_contribution: float = 0.0
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "score": self.score,
            "is_valid": self.is_valid,
            "confidence": self.confidence,
            "total_fields_compared": self.total_fields_compared,
            "exact_matches": self.exact_matches,
            "fuzzy_matches": self.fuzzy_matches,
            "mismatches": self.mismatches,
            "missing_fields": self.missing_fields,
            "required_fields_matched": self.required_fields_matched,
            "success": self.success,
            "veid_score_contribution": self.veid_score_contribution,
            "field_matches": [m.to_dict() for m in self.field_matches],
        }
    
    def to_chain_data(self) -> Dict[str, Any]:
        """Get data safe for on-chain storage."""
        return {
            "cross_validation_score": self.score,
            "cross_validation_valid": self.is_valid,
            "fields_matched": self.exact_matches + self.fuzzy_matches,
            "fields_compared": self.total_fields_compared,
            "veid_contribution": self.veid_score_contribution,
        }


class CrossValidator:
    """
    Cross-validator for barcode/MRZ data against OCR data.
    
    Compares fields from machine-readable sources against
    visually-extracted OCR data to ensure consistency.
    """
    
    # Field name mapping between different sources
    FIELD_MAPPINGS = {
        # Barcode field -> OCR field alternatives
        "full_name": ["full_name", "name"],
        "surname": ["surname", "last_name", "family_name"],
        "given_names": ["given_names", "first_name", "given_name", "forename"],
        "date_of_birth": ["date_of_birth", "dob", "birth_date"],
        "document_number": ["document_number", "id_number", "passport_number", "license_number"],
        "expiry_date": ["expiry_date", "expiration_date", "valid_until"],
        "sex": ["sex", "gender"],
        "nationality": ["nationality", "citizenship"],
        "address": ["address", "residence"],
    }
    
    def __init__(self, config: Optional[CrossValidationConfig] = None):
        """
        Initialize cross-validator.
        
        Args:
            config: Validation configuration. Uses defaults if None.
        """
        self.config = config or CrossValidationConfig()
    
    def validate(
        self,
        barcode_data: Union[PDF417Data, MRZData, Dict[str, str]],
        ocr_data: Dict[str, Any]
    ) -> CrossValidationResult:
        """
        Validate barcode/MRZ data against OCR data.
        
        Args:
            barcode_data: Parsed barcode or MRZ data, or dict of fields.
            ocr_data: OCR extraction result or dict of fields.
            
        Returns:
            CrossValidationResult with detailed matching info.
        """
        result = CrossValidationResult()
        
        try:
            # Normalize inputs to dictionaries
            barcode_fields = self._normalize_barcode_data(barcode_data)
            ocr_fields = self._normalize_ocr_data(ocr_data)
            
            if not barcode_fields:
                result.success = False
                result.error_message = "No barcode fields to validate"
                return result
            
            if not ocr_fields:
                result.success = False
                result.error_message = "No OCR fields to validate"
                return result
            
            # Compare fields
            matches = self._compare_fields(barcode_fields, ocr_fields)
            result.field_matches = matches
            
            # Calculate statistics
            result.total_fields_compared = len([m for m in matches if m.match_type not in (
                MatchType.MISSING_BARCODE, MatchType.MISSING_OCR, MatchType.NOT_COMPARED
            )])
            result.exact_matches = len([m for m in matches if m.match_type == MatchType.EXACT])
            result.fuzzy_matches = len([m for m in matches if m.match_type == MatchType.FUZZY])
            result.mismatches = len([m for m in matches if m.match_type == MatchType.MISMATCH])
            result.missing_fields = len([m for m in matches if m.match_type in (
                MatchType.MISSING_BARCODE, MatchType.MISSING_OCR
            )])
            
            # Check required fields
            result.required_fields_matched = True
            for req_field in self.config.required_fields:
                match = next((m for m in matches if m.field_name == req_field), None)
                if match is None or match.match_type in (
                    MatchType.MISMATCH, MatchType.MISSING_BARCODE, MatchType.MISSING_OCR
                ):
                    result.required_fields_matched = False
                    result.required_fields_missing.append(req_field)
            
            # Calculate overall score
            result.score = self._calculate_score(matches)
            
            # Determine validity
            result.is_valid = self._is_valid(result)
            
            # Calculate confidence
            result.confidence = self._calculate_confidence(result)
            
            # Calculate VEID score contribution
            result.veid_score_contribution = self._calculate_veid_contribution(result)
            
            result.success = True
            
        except Exception as e:
            logger.error(f"Cross-validation failed: {type(e).__name__}")
            result.success = False
            result.error_message = f"Validation error: {type(e).__name__}"
            result.score = 0.0
            result.is_valid = False
        
        return result
    
    def _normalize_barcode_data(
        self,
        data: Union[PDF417Data, MRZData, Dict[str, str]]
    ) -> Dict[str, str]:
        """Convert barcode data to normalized dictionary."""
        if isinstance(data, dict):
            return {k: str(v) for k, v in data.items() if v}
        elif hasattr(data, 'get_identity_fields'):
            return {k: str(v) for k, v in data.get_identity_fields().items() if v}
        else:
            return {}
    
    def _normalize_ocr_data(self, data: Dict[str, Any]) -> Dict[str, str]:
        """Convert OCR data to normalized dictionary."""
        result = {}
        
        for key, value in data.items():
            if isinstance(value, str) and value:
                result[key.lower()] = value
            elif hasattr(value, 'value'):  # ParsedField
                if value.value:
                    result[key.lower()] = str(value.value)
            elif isinstance(value, dict) and 'value' in value:
                if value['value']:
                    result[key.lower()] = str(value['value'])
        
        return result
    
    def _compare_fields(
        self,
        barcode_fields: Dict[str, str],
        ocr_fields: Dict[str, str]
    ) -> List[FieldMatch]:
        """Compare all fields between barcode and OCR data."""
        matches = []
        
        # Compare each barcode field
        for barcode_key, barcode_value in barcode_fields.items():
            # Find corresponding OCR field
            ocr_value = self._find_ocr_field(barcode_key, ocr_fields)
            
            # Get field weight
            weight = self.config.field_weights.get(barcode_key, 0.5)
            
            if ocr_value is None:
                matches.append(FieldMatch(
                    field_name=barcode_key,
                    barcode_value=barcode_value,
                    ocr_value="",
                    match_type=MatchType.MISSING_OCR,
                    similarity_score=0.0,
                    weight=weight,
                    contribution=0.0,
                ))
                continue
            
            # Compare values
            match_type, similarity = self._compare_values(
                barcode_key, barcode_value, ocr_value
            )
            
            contribution = similarity * weight
            if match_type == MatchType.EXACT:
                contribution += self.config.exact_match_bonus * weight
            
            matches.append(FieldMatch(
                field_name=barcode_key,
                barcode_value=barcode_value,
                ocr_value=ocr_value,
                match_type=match_type,
                similarity_score=similarity,
                weight=weight,
                contribution=contribution,
            ))
        
        return matches
    
    def _find_ocr_field(
        self,
        barcode_key: str,
        ocr_fields: Dict[str, str]
    ) -> Optional[str]:
        """Find corresponding OCR field for a barcode field."""
        # Direct match
        if barcode_key.lower() in ocr_fields:
            return ocr_fields[barcode_key.lower()]
        
        # Check alternatives
        alternatives = self.FIELD_MAPPINGS.get(barcode_key, [])
        for alt in alternatives:
            if alt.lower() in ocr_fields:
                return ocr_fields[alt.lower()]
        
        return None
    
    def _compare_values(
        self,
        field_name: str,
        barcode_value: str,
        ocr_value: str
    ) -> Tuple[MatchType, float]:
        """
        Compare two field values.
        
        Returns:
            Tuple of (match_type, similarity_score).
        """
        # Normalize values
        if not self.config.case_sensitive:
            barcode_value = barcode_value.upper()
            ocr_value = ocr_value.upper()
        
        barcode_norm = self._normalize_value(field_name, barcode_value)
        ocr_norm = self._normalize_value(field_name, ocr_value)
        
        # Exact match
        if barcode_norm == ocr_norm:
            return MatchType.EXACT, 1.0
        
        # Date field special handling
        if 'date' in field_name:
            if self.config.flexible_date_matching:
                if self._dates_match(barcode_norm, ocr_norm):
                    return MatchType.EXACT, 1.0
        
        # Fuzzy matching for names
        if self.config.enable_fuzzy_matching and field_name in ('full_name', 'surname', 'given_names'):
            similarity = self._calculate_similarity(barcode_norm, ocr_norm)
            
            if similarity >= self.config.string_similarity_threshold:
                return MatchType.FUZZY, similarity
            
            edit_distance = self._levenshtein_distance(barcode_norm, ocr_norm)
            if edit_distance <= self.config.max_edit_distance:
                # Calculate similarity from edit distance
                max_len = max(len(barcode_norm), len(ocr_norm))
                similarity = 1.0 - (edit_distance / max_len) if max_len > 0 else 0.0
                return MatchType.FUZZY, similarity
        
        # General similarity check
        similarity = self._calculate_similarity(barcode_norm, ocr_norm)
        
        if similarity >= self.config.string_similarity_threshold:
            return MatchType.FUZZY, similarity
        elif similarity >= 0.5:
            return MatchType.PARTIAL, similarity
        else:
            return MatchType.MISMATCH, similarity
    
    def _normalize_value(self, field_name: str, value: str) -> str:
        """Normalize value for comparison."""
        # Remove extra whitespace
        value = ' '.join(value.split())
        
        # Field-specific normalization
        if 'date' in field_name:
            # Remove separators for date comparison
            value = re.sub(r'[-/.]', '', value)
        elif field_name in ('sex', 'gender'):
            # Normalize sex/gender
            value = value.upper()
            if value in ('MALE', 'M', '1'):
                value = 'M'
            elif value in ('FEMALE', 'F', '2'):
                value = 'F'
        elif 'address' in field_name:
            # Normalize common address abbreviations
            value = value.upper()
            value = re.sub(r'\bST\b', 'STREET', value)
            value = re.sub(r'\bAVE\b', 'AVENUE', value)
            value = re.sub(r'\bRD\b', 'ROAD', value)
            value = re.sub(r'\bDR\b', 'DRIVE', value)
            value = re.sub(r'\bAPT\b', 'APARTMENT', value)
        
        return value.strip()
    
    def _dates_match(self, date1: str, date2: str) -> bool:
        """Check if two dates represent the same date."""
        # Try to extract date components
        patterns = [
            r'(\d{4})(\d{2})(\d{2})',  # YYYYMMDD
            r'(\d{2})(\d{2})(\d{4})',  # MMDDYYYY or DDMMYYYY
            r'(\d{4})-(\d{2})-(\d{2})',  # YYYY-MM-DD
        ]
        
        def extract_components(date_str: str) -> Optional[Tuple[str, str, str]]:
            for pattern in patterns:
                match = re.match(pattern, date_str)
                if match:
                    return match.groups()
            return None
        
        comp1 = extract_components(date1)
        comp2 = extract_components(date2)
        
        if comp1 and comp2:
            # Compare sorted components (handles different formats)
            return sorted(comp1) == sorted(comp2)
        
        return False
    
    def _calculate_similarity(self, str1: str, str2: str) -> float:
        """Calculate Jaccard similarity between two strings."""
        if not str1 or not str2:
            return 0.0
        
        # Character n-grams (bigrams)
        def get_ngrams(s: str, n: int = 2) -> set:
            s = s.lower()
            return set(s[i:i+n] for i in range(len(s) - n + 1))
        
        ngrams1 = get_ngrams(str1)
        ngrams2 = get_ngrams(str2)
        
        if not ngrams1 or not ngrams2:
            return 1.0 if str1.lower() == str2.lower() else 0.0
        
        intersection = len(ngrams1 & ngrams2)
        union = len(ngrams1 | ngrams2)
        
        return intersection / union if union > 0 else 0.0
    
    def _levenshtein_distance(self, str1: str, str2: str) -> int:
        """Calculate Levenshtein edit distance."""
        if len(str1) < len(str2):
            str1, str2 = str2, str1
        
        if len(str2) == 0:
            return len(str1)
        
        previous_row = range(len(str2) + 1)
        
        for i, c1 in enumerate(str1):
            current_row = [i + 1]
            
            for j, c2 in enumerate(str2):
                insertions = previous_row[j + 1] + 1
                deletions = current_row[j] + 1
                substitutions = previous_row[j] + (c1 != c2)
                current_row.append(min(insertions, deletions, substitutions))
            
            previous_row = current_row
        
        return previous_row[-1]
    
    def _calculate_score(self, matches: List[FieldMatch]) -> float:
        """Calculate overall validation score."""
        if not matches:
            return 0.0
        
        total_weight = sum(m.weight for m in matches if m.match_type not in (
            MatchType.MISSING_BARCODE, MatchType.MISSING_OCR, MatchType.NOT_COMPARED
        ))
        
        if total_weight == 0:
            return 0.0
        
        total_contribution = sum(m.contribution for m in matches)
        
        return min(1.0, total_contribution / total_weight)
    
    def _is_valid(self, result: CrossValidationResult) -> bool:
        """Determine if validation result is valid."""
        # Must have minimum matching fields
        good_matches = result.exact_matches + result.fuzzy_matches
        if good_matches < self.config.min_matching_fields:
            return False
        
        # Required fields must match
        if not result.required_fields_matched:
            return False
        
        # Score must meet threshold
        if result.score < self.config.string_similarity_threshold:
            return False
        
        return True
    
    def _calculate_confidence(self, result: CrossValidationResult) -> float:
        """Calculate confidence in the validation result."""
        if result.total_fields_compared == 0:
            return 0.0
        
        # Base confidence from match ratio
        good_matches = result.exact_matches + result.fuzzy_matches
        match_ratio = good_matches / result.total_fields_compared
        
        # Bonus for exact matches
        exact_ratio = result.exact_matches / result.total_fields_compared
        
        # Penalty for mismatches
        mismatch_ratio = result.mismatches / result.total_fields_compared
        
        confidence = (
            match_ratio * 0.6 +
            exact_ratio * 0.3 +
            (1.0 - mismatch_ratio) * 0.1
        )
        
        return max(0.0, min(1.0, confidence))
    
    def _calculate_veid_contribution(self, result: CrossValidationResult) -> float:
        """Calculate contribution to VEID score."""
        from ml.barcode_scanning.config import ScoringConfig
        
        scoring = ScoringConfig()
        
        if not result.is_valid:
            return scoring.cross_validation_penalty
        
        # Scale by validation score
        contribution = result.score * scoring.cross_validation_weight
        
        # Bonus for high confidence
        if result.confidence > 0.9:
            contribution += 0.02
        
        return min(contribution, scoring.max_score_contribution)


def calculate_combined_identity_hash(
    barcode_data: Union[PDF417Data, MRZData],
    ocr_data: Dict[str, str],
    salt: str
) -> str:
    """
    Calculate a combined identity hash from barcode and OCR data.
    
    This creates a single hash representing the identity claim
    for on-chain storage.
    
    Args:
        barcode_data: Parsed barcode or MRZ data.
        ocr_data: OCR-extracted fields.
        salt: Salt for hashing.
        
    Returns:
        SHA-256 hash of combined identity data.
    """
    # Get identity fields from barcode
    if hasattr(barcode_data, 'get_identity_fields'):
        barcode_fields = barcode_data.get_identity_fields()
    else:
        barcode_fields = {}
    
    # Combine with OCR fields (prefer barcode data)
    combined = {**ocr_data, **barcode_fields}
    
    # Create deterministic string representation
    sorted_items = sorted(combined.items())
    data_string = '|'.join(f"{k}:{v}" for k, v in sorted_items if v)
    
    # Hash with salt
    salted = f"{salt}:{data_string}".encode('utf-8')
    return hashlib.sha256(salted).hexdigest()
