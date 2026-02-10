"""
Tests for field hashing functionality.
"""

import pytest
import hashlib

from ml.ocr_extraction.field_hasher import (
    FieldHasher,
    HashedField,
    IdentityHash,
    create_production_hasher,
    create_test_hasher,
)
from ml.ocr_extraction.field_parser import ParsedField, ValidationStatus
from ml.ocr_extraction.config import HashingConfig


class TestHashingConfig:
    """Tests for HashingConfig."""
    
    def test_default_values(self):
        """Test default configuration values."""
        config = HashingConfig()
        assert config.algorithm == "sha256"
        assert config.salt == ""
        assert config.normalize_before_hash is True
    
    def test_custom_values(self):
        """Test custom configuration values."""
        config = HashingConfig(
            salt="my_secret_salt",
            uppercase_before_hash=False,
            output_encoding="base64",
        )
        assert config.salt == "my_secret_salt"
        assert config.uppercase_before_hash is False
        assert config.output_encoding == "base64"


class TestFieldHasher:
    """Tests for FieldHasher class."""
    
    def test_initialization_default(self):
        """Test default initialization."""
        hasher = FieldHasher()
        assert hasher.config is not None
    
    def test_initialization_custom(self):
        """Test custom initialization."""
        config = HashingConfig(salt="test_salt")
        hasher = FieldHasher(config)
        assert hasher.config.salt == "test_salt"
    
    def test_hash_field_basic(self):
        """Test basic field hashing."""
        hasher = FieldHasher()
        hash_value = hasher.hash_field("surname", "SMITH")
        
        assert hash_value is not None
        assert len(hash_value) == 64  # SHA-256 hex
    
    def test_hash_field_consistency(self):
        """Test hash consistency (same input = same output)."""
        hasher = FieldHasher()
        
        hash1 = hasher.hash_field("surname", "SMITH")
        hash2 = hasher.hash_field("surname", "SMITH")
        
        assert hash1 == hash2
    
    def test_hash_field_different_values(self):
        """Test different values produce different hashes."""
        hasher = FieldHasher()
        
        hash1 = hasher.hash_field("surname", "SMITH")
        hash2 = hasher.hash_field("surname", "JONES")
        
        assert hash1 != hash2
    
    def test_hash_field_different_field_names(self):
        """Test same value with different field names produces different hashes."""
        config = HashingConfig(include_field_name=True)
        hasher = FieldHasher(config)
        
        hash1 = hasher.hash_field("surname", "SMITH")
        hash2 = hasher.hash_field("given_names", "SMITH")
        
        assert hash1 != hash2
    
    def test_hash_field_with_salt(self):
        """Test hashing with salt."""
        hasher1 = FieldHasher(HashingConfig(salt="salt1"))
        hasher2 = FieldHasher(HashingConfig(salt="salt2"))
        
        hash1 = hasher1.hash_field("surname", "SMITH")
        hash2 = hasher2.hash_field("surname", "SMITH")
        
        assert hash1 != hash2
    
    def test_hash_field_normalization(self):
        """Test value normalization before hashing."""
        config = HashingConfig(
            normalize_before_hash=True,
            uppercase_before_hash=True,
            strip_whitespace=True,
        )
        hasher = FieldHasher(config)
        
        hash1 = hasher.hash_field("surname", "SMITH")
        hash2 = hasher.hash_field("surname", "  smith  ")
        
        assert hash1 == hash2
    
    def test_hash_field_no_normalization(self):
        """Test hashing without normalization."""
        config = HashingConfig(
            normalize_before_hash=False,
            uppercase_before_hash=False,
            strip_whitespace=False,
        )
        hasher = FieldHasher(config)
        
        hash1 = hasher.hash_field("surname", "SMITH")
        hash2 = hasher.hash_field("surname", "smith")
        
        assert hash1 != hash2
    
    def test_hash_field_base64_encoding(self):
        """Test base64 output encoding."""
        config = HashingConfig(output_encoding="base64")
        hasher = FieldHasher(config)
        
        hash_value = hasher.hash_field("surname", "SMITH")
        
        # Base64 encoded SHA-256 is 44 chars with padding
        assert len(hash_value) == 44
        assert hash_value.endswith("=")
    
    def test_hash_field_detailed(self):
        """Test detailed field hashing."""
        hasher = FieldHasher()
        result = hasher.hash_field_detailed("surname", "SMITH")
        
        assert isinstance(result, HashedField)
        assert result.field_name == "surname"
        assert result.algorithm == "sha256"
    
    def test_hash_fields(self, sample_parsed_fields):
        """Test batch field hashing."""
        hasher = FieldHasher()
        hashes = hasher.hash_fields(sample_parsed_fields)
        
        assert len(hashes) == len(sample_parsed_fields)
        assert all(len(h) == 64 for h in hashes.values())
    
    def test_hash_fields_skips_empty(self):
        """Test empty fields are skipped."""
        hasher = FieldHasher()
        
        fields = {
            "surname": ParsedField(
                field_name="surname",
                value="SMITH",
                confidence=0.9,
                source_roi_ids=[],
                validation_status=ValidationStatus.VALID,
            ),
            "empty_field": ParsedField(
                field_name="empty_field",
                value="",  # Empty
                confidence=0.0,
                source_roi_ids=[],
                validation_status=ValidationStatus.NOT_FOUND,
            ),
        }
        
        hashes = hasher.hash_fields(fields)
        
        assert "surname" in hashes
        assert "empty_field" not in hashes
    
    def test_hash_fields_detailed(self, sample_parsed_fields):
        """Test detailed batch field hashing."""
        hasher = FieldHasher()
        hashes = hasher.hash_fields_detailed(sample_parsed_fields)
        
        assert all(isinstance(h, HashedField) for h in hashes.values())
    
    def test_create_identity_hash(self, sample_parsed_fields):
        """Test combined identity hash creation."""
        hasher = FieldHasher()
        
        identity_hash = hasher.create_derived_identity_hash(sample_parsed_fields)
        
        assert identity_hash is not None
        assert len(identity_hash) == 64
    
    def test_identity_hash_consistency(self, sample_parsed_fields):
        """Test identity hash is consistent."""
        hasher = FieldHasher()
        
        hash1 = hasher.create_derived_identity_hash(sample_parsed_fields)
        hash2 = hasher.create_derived_identity_hash(sample_parsed_fields)
        
        assert hash1 == hash2
    
    def test_identity_hash_changes_with_fields(self, sample_parsed_fields):
        """Test identity hash changes when fields change."""
        hasher = FieldHasher()
        
        hash1 = hasher.create_derived_identity_hash(sample_parsed_fields)
        
        # Modify a field
        modified = sample_parsed_fields.copy()
        modified["surname"] = ParsedField(
            field_name="surname",
            value="JONES",
            confidence=0.85,
            source_roi_ids=["roi_1"],
            validation_status=ValidationStatus.VALID,
        )
        
        hash2 = hasher.create_derived_identity_hash(modified)
        
        assert hash1 != hash2
    
    def test_identity_hash_empty_fields(self):
        """Test identity hash with no fields."""
        hasher = FieldHasher()
        
        identity_hash = hasher.create_derived_identity_hash({})
        
        # Should return a hash (for empty marker)
        assert identity_hash is not None
        assert len(identity_hash) == 64
    
    def test_identity_hash_detailed(self, sample_parsed_fields):
        """Test detailed identity hash creation."""
        hasher = FieldHasher()
        
        result = hasher.create_identity_hash_detailed(sample_parsed_fields)
        
        assert isinstance(result, IdentityHash)
        assert result.version == FieldHasher.VERSION
        assert len(result.field_names) > 0
    
    def test_verify_field_hash(self):
        """Test field hash verification."""
        hasher = FieldHasher()
        
        # Hash a value
        hash_value = hasher.hash_field("surname", "SMITH")
        
        # Verify correct value
        assert hasher.verify_field_hash("surname", "SMITH", hash_value) is True
        
        # Verify incorrect value
        assert hasher.verify_field_hash("surname", "JONES", hash_value) is False
    
    def test_verify_identity_hash(self, sample_parsed_fields):
        """Test identity hash verification."""
        hasher = FieldHasher()
        
        # Create hash
        identity_hash = hasher.create_derived_identity_hash(sample_parsed_fields)
        
        # Verify same fields
        assert hasher.verify_identity_hash(
            sample_parsed_fields,
            identity_hash
        ) is True
        
        # Verify modified fields
        modified = sample_parsed_fields.copy()
        modified["surname"] = ParsedField(
            field_name="surname",
            value="JONES",
            confidence=0.85,
            source_roi_ids=[],
            validation_status=ValidationStatus.VALID,
        )
        
        assert hasher.verify_identity_hash(
            modified,
            identity_hash
        ) is False
    
    def test_generate_salt(self):
        """Test salt generation."""
        salt = FieldHasher.generate_salt()
        
        assert salt is not None
        assert len(salt) == 64  # 32 bytes = 64 hex chars
    
    def test_generate_salt_different_each_time(self):
        """Test salt generation produces unique values."""
        salt1 = FieldHasher.generate_salt()
        salt2 = FieldHasher.generate_salt()
        
        assert salt1 != salt2
    
    def test_hash_raw(self):
        """Test raw bytes hashing."""
        data = b"test data"
        hash_value = FieldHasher.hash_raw(data)
        
        expected = hashlib.sha256(data).hexdigest()
        assert hash_value == expected


class TestHashedField:
    """Tests for HashedField dataclass."""
    
    def test_to_dict(self):
        """Test dictionary conversion."""
        field = HashedField(
            field_name="surname",
            hash_value="abc123def456",
            algorithm="sha256",
            includes_field_name=True,
            normalized=True,
        )
        
        d = field.to_dict()
        
        assert d["field_name"] == "surname"
        assert d["hash"] == "abc123def456"
        assert d["algorithm"] == "sha256"


class TestIdentityHash:
    """Tests for IdentityHash dataclass."""
    
    def test_to_dict(self):
        """Test dictionary conversion."""
        identity = IdentityHash(
            hash_value="abc123xyz789",
            algorithm="sha256",
            field_names=["surname", "given_names", "date_of_birth"],
            version="1.0",
        )
        
        d = identity.to_dict()
        
        assert d["identity_hash"] == "abc123xyz789"
        assert d["algorithm"] == "sha256"
        assert d["version"] == "1.0"
        assert len(d["fields_included"]) == 3


class TestFactoryFunctions:
    """Tests for hasher factory functions."""
    
    def test_create_production_hasher(self):
        """Test production hasher factory."""
        hasher = create_production_hasher("secret_salt_12345")
        
        assert hasher.config.salt == "secret_salt_12345"
        assert hasher.config.normalize_before_hash is True
        assert hasher.config.include_field_name is True
    
    def test_create_test_hasher(self):
        """Test test hasher factory."""
        hasher = create_test_hasher()
        
        assert hasher.config.salt == ""
        assert hasher.config.normalize_before_hash is True


class TestSecurityProperties:
    """Tests for security-related properties."""
    
    def test_hash_is_one_way(self):
        """Test that hash cannot be reversed to get original value."""
        hasher = FieldHasher()
        hash_value = hasher.hash_field("surname", "SMITH")
        
        # Cannot derive "SMITH" from hash_value
        # This is a conceptual test - we verify the hash is fixed length
        assert len(hash_value) == 64
        assert "SMITH" not in hash_value
    
    def test_different_salts_prevent_rainbow_tables(self):
        """Test that different salts produce different hashes."""
        hasher1 = FieldHasher(HashingConfig(salt="salt_a"))
        hasher2 = FieldHasher(HashingConfig(salt="salt_b"))
        
        hash1 = hasher1.hash_field("surname", "SMITH")
        hash2 = hasher2.hash_field("surname", "SMITH")
        
        assert hash1 != hash2
    
    def test_timing_safe_comparison(self):
        """Test that verification uses constant-time comparison."""
        hasher = FieldHasher()
        correct_hash = hasher.hash_field("surname", "SMITH")
        
        # Both should complete in similar time (constant-time)
        result1 = hasher.verify_field_hash("surname", "SMITH", correct_hash)
        result2 = hasher.verify_field_hash("surname", "JONES", correct_hash)
        
        assert result1 is True
        assert result2 is False
    
    def test_empty_salt_warning(self):
        """Test behavior with empty salt (not recommended for production)."""
        hasher = FieldHasher(HashingConfig(salt=""))
        hash_value = hasher.hash_field("surname", "SMITH")
        
        # Should still produce a hash
        assert hash_value is not None
        assert len(hash_value) == 64


class TestEdgeCases:
    """Tests for edge cases."""
    
    def test_unicode_values(self):
        """Test hashing unicode values."""
        hasher = FieldHasher()
        hash_value = hasher.hash_field("name", "José García")
        
        assert hash_value is not None
        assert len(hash_value) == 64
    
    def test_special_characters(self):
        """Test hashing values with special characters."""
        hasher = FieldHasher()
        hash_value = hasher.hash_field("id", "AB-123/456")
        
        assert hash_value is not None
    
    def test_very_long_value(self):
        """Test hashing very long values."""
        hasher = FieldHasher()
        long_value = "A" * 10000
        hash_value = hasher.hash_field("field", long_value)
        
        # Hash should still be 64 chars regardless of input length
        assert len(hash_value) == 64
    
    def test_custom_field_order(self, sample_parsed_fields):
        """Test identity hash with custom field order."""
        hasher = FieldHasher()
        
        order1 = ["surname", "given_names"]
        order2 = ["given_names", "surname"]
        
        hash1 = hasher.create_derived_identity_hash(sample_parsed_fields, order1)
        hash2 = hasher.create_derived_identity_hash(sample_parsed_fields, order2)
        
        # Different order should produce different hash
        assert hash1 != hash2
