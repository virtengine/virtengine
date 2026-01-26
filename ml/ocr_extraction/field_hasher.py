"""
Field hasher for secure on-chain storage.

This module provides cryptographic hashing of parsed fields
to enable on-chain identity verification without storing
plaintext PII (Personally Identifiable Information).

SECURITY NOTE: Never store plaintext PII on-chain. Only hashed
or encrypted values should be persisted to the blockchain.
"""

import hashlib
import hmac
import base64
from typing import Dict, Optional, List, Tuple
from dataclasses import dataclass

from ml.ocr_extraction.config import HashingConfig
from ml.ocr_extraction.field_parser import ParsedField


@dataclass
class HashedField:
    """A hashed field value with metadata."""
    field_name: str
    hash_value: str
    algorithm: str
    includes_field_name: bool
    normalized: bool
    
    def to_dict(self) -> Dict[str, str]:
        """Convert to dictionary."""
        return {
            "field_name": self.field_name,
            "hash": self.hash_value,
            "algorithm": self.algorithm,
        }


@dataclass
class IdentityHash:
    """Combined identity hash from multiple fields."""
    hash_value: str
    algorithm: str
    field_names: List[str]
    version: str
    
    def to_dict(self) -> Dict[str, str]:
        """Convert to dictionary."""
        return {
            "identity_hash": self.hash_value,
            "algorithm": self.algorithm,
            "fields_included": self.field_names,
            "version": self.version,
        }


class FieldHasher:
    """
    Secure hasher for identity document fields.
    
    Implements consistent, reproducible hashing of field values
    for on-chain storage. Uses salted SHA-256 by default.
    
    IMPORTANT: The salt should be kept secret and consistent
    across all nodes for hash verification to work.
    """
    
    # Version for hash format (increment if algorithm changes)
    VERSION = "1.0"
    
    # Standard field order for identity hash
    IDENTITY_HASH_FIELDS = [
        "full_name",
        "surname",
        "given_names",
        "date_of_birth",
        "document_number",
        "nationality",
    ]
    
    def __init__(self, config: Optional[HashingConfig] = None):
        """
        Initialize field hasher.
        
        Args:
            config: Hashing configuration. Uses defaults if None.
        """
        self.config = config or HashingConfig()
    
    def hash_field(
        self,
        field_name: str,
        value: str,
        salt_override: Optional[str] = None
    ) -> str:
        """
        Hash a single field value.
        
        Uses SHA-256 with optional salt and field name inclusion
        for domain separation.
        
        Args:
            field_name: Name of the field
            value: Field value to hash
            salt_override: Override configuration salt
            
        Returns:
            Hex-encoded hash string
        """
        # Normalize value if configured
        normalized_value = self._normalize_value(value)
        
        # Build hash input
        hash_input = self._build_hash_input(field_name, normalized_value)
        
        # Apply salt
        salt = salt_override or self.config.salt
        if salt:
            # Use HMAC for proper key-based hashing
            hash_bytes = hmac.new(
                salt.encode('utf-8'),
                hash_input.encode('utf-8'),
                hashlib.sha256
            ).digest()
        else:
            hash_bytes = hashlib.sha256(hash_input.encode('utf-8')).digest()
        
        # Encode output
        if self.config.output_encoding == "base64":
            return base64.b64encode(hash_bytes).decode('utf-8')
        else:
            return hash_bytes.hex()
    
    def hash_field_detailed(
        self,
        field_name: str,
        value: str
    ) -> HashedField:
        """
        Hash a field with detailed metadata.
        
        Args:
            field_name: Name of the field
            value: Field value to hash
            
        Returns:
            HashedField with hash and metadata
        """
        hash_value = self.hash_field(field_name, value)
        
        return HashedField(
            field_name=field_name,
            hash_value=hash_value,
            algorithm=self.config.algorithm,
            includes_field_name=self.config.include_field_name,
            normalized=self.config.normalize_before_hash,
        )
    
    def hash_fields(
        self,
        fields: Dict[str, ParsedField]
    ) -> Dict[str, str]:
        """
        Hash all fields in a dictionary.
        
        Args:
            fields: Dictionary of parsed fields
            
        Returns:
            Dictionary mapping field names to hash values
        """
        return {
            name: self.hash_field(name, field.value)
            for name, field in fields.items()
            if field.value  # Skip empty values
        }
    
    def hash_fields_detailed(
        self,
        fields: Dict[str, ParsedField]
    ) -> Dict[str, HashedField]:
        """
        Hash all fields with detailed metadata.
        
        Args:
            fields: Dictionary of parsed fields
            
        Returns:
            Dictionary mapping field names to HashedField objects
        """
        return {
            name: self.hash_field_detailed(name, field.value)
            for name, field in fields.items()
            if field.value
        }
    
    def create_derived_identity_hash(
        self,
        fields: Dict[str, ParsedField],
        custom_field_order: Optional[List[str]] = None
    ) -> str:
        """
        Create a combined identity hash from multiple fields.
        
        This creates a single hash representing the identity
        derived from all available fields. The hash is deterministic
        given the same field values.
        
        Args:
            fields: Dictionary of parsed fields
            custom_field_order: Custom order of fields to include
            
        Returns:
            Combined identity hash string
        """
        field_order = custom_field_order or self.IDENTITY_HASH_FIELDS
        
        # Collect values in consistent order
        values = []
        for field_name in field_order:
            if field_name in fields and fields[field_name].value:
                normalized = self._normalize_value(fields[field_name].value)
                values.append(f"{field_name}:{normalized}")
        
        if not values:
            # No fields available - use empty marker
            return self.hash_field("identity", "EMPTY")
        
        # Combine all field values
        combined = "|".join(values)
        
        # Hash the combined value
        return self.hash_field("identity", combined)
    
    def create_identity_hash_detailed(
        self,
        fields: Dict[str, ParsedField],
        custom_field_order: Optional[List[str]] = None
    ) -> IdentityHash:
        """
        Create identity hash with detailed metadata.
        
        Args:
            fields: Dictionary of parsed fields
            custom_field_order: Custom order of fields to include
            
        Returns:
            IdentityHash with hash and metadata
        """
        field_order = custom_field_order or self.IDENTITY_HASH_FIELDS
        
        # Determine which fields are included
        included_fields = [
            name for name in field_order
            if name in fields and fields[name].value
        ]
        
        hash_value = self.create_derived_identity_hash(fields, field_order)
        
        return IdentityHash(
            hash_value=hash_value,
            algorithm=self.config.algorithm,
            field_names=included_fields,
            version=self.VERSION,
        )
    
    def verify_field_hash(
        self,
        field_name: str,
        value: str,
        expected_hash: str
    ) -> bool:
        """
        Verify a field value matches an expected hash.
        
        Args:
            field_name: Name of the field
            value: Field value to verify
            expected_hash: Expected hash value
            
        Returns:
            True if hash matches, False otherwise
        """
        computed_hash = self.hash_field(field_name, value)
        
        # Use constant-time comparison to prevent timing attacks
        return hmac.compare_digest(computed_hash, expected_hash)
    
    def verify_identity_hash(
        self,
        fields: Dict[str, ParsedField],
        expected_hash: str,
        field_order: Optional[List[str]] = None
    ) -> bool:
        """
        Verify an identity hash matches the fields.
        
        Args:
            fields: Dictionary of parsed fields
            expected_hash: Expected identity hash
            field_order: Field order used in original hash
            
        Returns:
            True if hash matches, False otherwise
        """
        computed_hash = self.create_derived_identity_hash(fields, field_order)
        
        return hmac.compare_digest(computed_hash, expected_hash)
    
    def _normalize_value(self, value: str) -> str:
        """Normalize a value for consistent hashing."""
        result = value
        
        if self.config.strip_whitespace:
            result = result.strip()
            # Also normalize internal whitespace
            result = " ".join(result.split())
        
        if self.config.uppercase_before_hash:
            result = result.upper()
        
        return result
    
    def _build_hash_input(self, field_name: str, value: str) -> str:
        """Build the hash input string."""
        if self.config.include_field_name:
            # Domain separation: include field name in hash
            return f"{field_name}:{value}"
        else:
            return value
    
    @staticmethod
    def generate_salt(length: int = 32) -> str:
        """
        Generate a random salt for hashing.
        
        Args:
            length: Salt length in bytes
            
        Returns:
            Hex-encoded random salt
        """
        import secrets
        return secrets.token_hex(length)
    
    @staticmethod
    def hash_raw(data: bytes) -> str:
        """
        Hash raw bytes using SHA-256.
        
        Args:
            data: Raw bytes to hash
            
        Returns:
            Hex-encoded hash
        """
        return hashlib.sha256(data).hexdigest()


def create_production_hasher(salt: str) -> FieldHasher:
    """
    Create a hasher configured for production use.
    
    Args:
        salt: Secret salt for hashing (should be from env/config)
        
    Returns:
        Configured FieldHasher
    """
    config = HashingConfig(
        algorithm="sha256",
        salt=salt,
        normalize_before_hash=True,
        uppercase_before_hash=True,
        strip_whitespace=True,
        include_field_name=True,
        output_encoding="hex",
    )
    return FieldHasher(config)


def create_test_hasher() -> FieldHasher:
    """
    Create a hasher for testing (no salt).
    
    WARNING: Do not use in production.
    
    Returns:
        FieldHasher for testing
    """
    config = HashingConfig(
        algorithm="sha256",
        salt="",  # No salt for testing
        normalize_before_hash=True,
        uppercase_before_hash=True,
        strip_whitespace=True,
        include_field_name=True,
        output_encoding="hex",
    )
    return FieldHasher(config)
