"""
PII anonymization for training data.

This module provides secure anonymization of personally identifiable
information (PII) in training data to comply with privacy requirements.

Anonymization methods:
- SHA256 hashing with salt
- BLAKE2 hashing with salt
- Redaction (complete removal)
- Tokenization (replacement with tokens)
"""

import hashlib
import secrets
import logging
from dataclasses import dataclass, field
from typing import List, Dict, Optional, Set, Any
from enum import Enum

from ml.training.config import AnonymizationMethod
from ml.training.dataset.ingestion import (
    Dataset,
    IdentitySample,
    DocumentInfo,
)

logger = logging.getLogger(__name__)


# PII field categories
class PIICategory(str, Enum):
    """Categories of PII fields."""
    DIRECT_IDENTIFIER = "direct_identifier"  # Names, IDs
    QUASI_IDENTIFIER = "quasi_identifier"    # DOB, location
    SENSITIVE = "sensitive"                  # Biometric data
    DEVICE = "device"                        # Device identifiers


# Default PII fields to anonymize
DEFAULT_PII_FIELDS: Dict[PIICategory, List[str]] = {
    PIICategory.DIRECT_IDENTIFIER: [
        "sample_id",
        "doc_id_hash",
        "document_number",
        "name",
        "full_name",
        "first_name",
        "last_name",
    ],
    PIICategory.QUASI_IDENTIFIER: [
        "date_of_birth",
        "nationality",
        "address",
        "postal_code",
    ],
    PIICategory.SENSITIVE: [
        "face_embedding",
        "biometric_hash",
    ],
    PIICategory.DEVICE: [
        "device_id",
        "device_model",
        "ip_address",
        "session_id",
    ],
}


@dataclass
class AnonymizationResult:
    """Result of anonymization operation."""
    
    # Statistics
    samples_processed: int = 0
    fields_anonymized: int = 0
    
    # Anonymization details
    method_used: str = ""
    salt_hash: str = ""  # Hash of salt for audit (not the salt itself)
    
    # Field statistics
    fields_by_category: Dict[str, int] = field(default_factory=dict)
    
    # Status
    success: bool = True
    warnings: List[str] = field(default_factory=list)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for logging/storage."""
        return {
            "samples_processed": self.samples_processed,
            "fields_anonymized": self.fields_anonymized,
            "method_used": self.method_used,
            "salt_hash": self.salt_hash,
            "fields_by_category": self.fields_by_category,
            "success": self.success,
            "warnings": self.warnings,
        }


class PIIAnonymizer:
    """
    Anonymizes PII in training datasets.
    
    Supports multiple anonymization methods:
    - Hashing: One-way transformation using SHA256 or BLAKE2
    - Redaction: Complete removal of field values
    - Tokenization: Replacement with random tokens
    
    Anonymization is deterministic within a session (same salt)
    to maintain consistency for training.
    """
    
    def __init__(
        self,
        method: str = AnonymizationMethod.HASH_SHA256.value,
        salt: Optional[str] = None,
        pii_fields: Optional[Dict[PIICategory, List[str]]] = None,
    ):
        """
        Initialize the anonymizer.
        
        Args:
            method: Anonymization method to use
            salt: Salt for hashing (generated if not provided)
            pii_fields: Custom PII field definitions
        """
        self.method = method
        self.salt = salt or secrets.token_hex(32)
        self.pii_fields = pii_fields or DEFAULT_PII_FIELDS
        
        # Token mapping for tokenization method
        self._token_map: Dict[str, str] = {}
        self._next_token_id = 0
        
        # Track all PII field names
        self._all_pii_fields: Set[str] = set()
        for fields in self.pii_fields.values():
            self._all_pii_fields.update(fields)
    
    def anonymize_dataset(self, dataset: Dataset) -> AnonymizationResult:
        """
        Anonymize all PII in a dataset.
        
        Args:
            dataset: Dataset to anonymize (modified in place)
            
        Returns:
            AnonymizationResult with statistics
        """
        result = AnonymizationResult(
            method_used=self.method,
            salt_hash=self._hash_salt(),
        )
        
        # Process all splits
        for split in [dataset.train, dataset.validation, dataset.test]:
            for sample in split.samples:
                fields_count = self._anonymize_sample(sample)
                result.samples_processed += 1
                result.fields_anonymized += fields_count
        
        # Count by category
        for category in PIICategory:
            result.fields_by_category[category.value] = len(
                self.pii_fields.get(category, [])
            )
        
        logger.info(
            f"Anonymized {result.fields_anonymized} fields in "
            f"{result.samples_processed} samples using {self.method}"
        )
        
        return result
    
    def _anonymize_sample(self, sample: IdentitySample) -> int:
        """Anonymize a single sample. Returns count of fields anonymized."""
        fields_anonymized = 0
        
        # Anonymize sample ID
        if "sample_id" in self._all_pii_fields:
            sample.sample_id = self._anonymize_value(sample.sample_id)
            fields_anonymized += 1
        
        # Anonymize document info
        if sample.document_info:
            fields_anonymized += self._anonymize_document_info(
                sample.document_info
            )
        
        # Anonymize annotations
        if sample.annotations:
            for key in list(sample.annotations.keys()):
                if key in self._all_pii_fields:
                    sample.annotations[key] = self._anonymize_value(
                        str(sample.annotations[key])
                    )
                    fields_anonymized += 1
        
        return fields_anonymized
    
    def _anonymize_document_info(self, doc_info: DocumentInfo) -> int:
        """Anonymize document info fields. Returns count of fields anonymized."""
        fields_anonymized = 0
        
        if "doc_id_hash" in self._all_pii_fields:
            doc_info.doc_id_hash = self._anonymize_value(doc_info.doc_id_hash)
            fields_anonymized += 1
        
        return fields_anonymized
    
    def _anonymize_value(self, value: str) -> str:
        """Anonymize a single value using the configured method."""
        if not value:
            return ""
        
        if self.method == AnonymizationMethod.HASH_SHA256.value:
            return self._hash_sha256(value)
        elif self.method == AnonymizationMethod.HASH_BLAKE2.value:
            return self._hash_blake2(value)
        elif self.method == AnonymizationMethod.REDACT.value:
            return self._redact(value)
        elif self.method == AnonymizationMethod.TOKENIZE.value:
            return self._tokenize(value)
        else:
            logger.warning(f"Unknown anonymization method: {self.method}")
            return self._hash_sha256(value)
    
    def _hash_sha256(self, value: str) -> str:
        """Hash value with SHA256 and salt."""
        data = f"{self.salt}{value}".encode('utf-8')
        return hashlib.sha256(data).hexdigest()[:16]
    
    def _hash_blake2(self, value: str) -> str:
        """Hash value with BLAKE2b and salt."""
        data = f"{self.salt}{value}".encode('utf-8')
        return hashlib.blake2b(data, digest_size=8).hexdigest()
    
    def _redact(self, value: str) -> str:
        """Redact value completely."""
        return "[REDACTED]"
    
    def _tokenize(self, value: str) -> str:
        """Replace value with a consistent token."""
        if value not in self._token_map:
            self._token_map[value] = f"TOKEN_{self._next_token_id:08d}"
            self._next_token_id += 1
        return self._token_map[value]
    
    def _hash_salt(self) -> str:
        """Hash the salt for audit logging (don't expose actual salt)."""
        return hashlib.sha256(self.salt.encode()).hexdigest()[:16]
    
    def anonymize_value(self, value: str) -> str:
        """
        Public method to anonymize a single value.
        
        Useful for anonymizing values outside of datasets.
        
        Args:
            value: Value to anonymize
            
        Returns:
            Anonymized value
        """
        return self._anonymize_value(value)
    
    def get_anonymization_report(self) -> Dict[str, Any]:
        """
        Generate a report of anonymization settings.
        
        Returns:
            Dictionary with anonymization configuration
            (does not include actual salt)
        """
        return {
            "method": self.method,
            "salt_hash": self._hash_salt(),
            "pii_categories": list(self.pii_fields.keys()),
            "total_pii_fields": len(self._all_pii_fields),
            "pii_fields_by_category": {
                cat.value: len(fields)
                for cat, fields in self.pii_fields.items()
            },
        }
