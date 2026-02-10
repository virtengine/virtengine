"""
ML Conformance Testing Package.

This package provides tools for testing Go-Python inference conformance:
- generate_test_vectors.py: Generate deterministic test vectors
- verify_go_output.py: Verify Go outputs match Python expectations

Task Reference: VE-3006 - Go-Python Conformance Testing
"""

from ml.conformance.generate_test_vectors import (
    generate_test_vectors,
    generate_deterministic_embedding,
    extract_features,
    compute_score,
    TestVectorEntry,
    TestVectorInput,
    DocQualityFeatures,
)

from ml.conformance.verify_go_output import (
    verify_outputs,
    verify_single_vector,
    generate_report,
    VerificationResult,
    DEFAULT_TOLERANCE,
)

__all__ = [
    # Generation
    "generate_test_vectors",
    "generate_deterministic_embedding",
    "extract_features",
    "compute_score",
    "TestVectorEntry",
    "TestVectorInput",
    "DocQualityFeatures",
    # Verification
    "verify_outputs",
    "verify_single_vector",
    "generate_report",
    "VerificationResult",
    "DEFAULT_TOLERANCE",
]
