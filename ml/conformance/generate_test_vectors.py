"""
Generate test vectors for Go conformance testing.

Outputs JSON file with input/output pairs that Go must match exactly.
Uses the same feature extraction and scoring logic as the training pipeline
to ensure Go and Python produce identical outputs.

Task Reference: VE-3006 - Go-Python Conformance Testing

Usage:
    python -m ml.conformance.generate_test_vectors
    python ml/conformance/generate_test_vectors.py
"""

import json
import hashlib
import time
from dataclasses import dataclass, field, asdict
from pathlib import Path
from typing import List, Dict, Any, Optional
import numpy as np

# Set deterministic seeds immediately
np.random.seed(42)

# Constants matching Go inference package
FACE_EMBEDDING_DIM = 512
DOC_QUALITY_DIM = 5
OCR_FIELDS_DIM = 10  # 5 fields * 2
METADATA_DIM = 16
TOTAL_FEATURE_DIM = 768
PADDING_DIM = TOTAL_FEATURE_DIM - FACE_EMBEDDING_DIM - DOC_QUALITY_DIM - OCR_FIELDS_DIM - METADATA_DIM

# OCR field names matching Go
OCR_FIELD_NAMES = [
    "name",
    "date_of_birth",
    "document_number",
    "expiry_date",
    "nationality",
]

# Reason codes matching Go
REASON_CODE_SUCCESS = "SUCCESS"
REASON_CODE_HIGH_CONFIDENCE = "HIGH_CONFIDENCE"
REASON_CODE_LOW_CONFIDENCE = "LOW_CONFIDENCE"
REASON_CODE_FACE_MISMATCH = "FACE_MISMATCH"
REASON_CODE_LOW_DOC_QUALITY = "LOW_DOC_QUALITY"
REASON_CODE_LOW_OCR_CONFIDENCE = "LOW_OCR_CONFIDENCE"
REASON_CODE_INSUFFICIENT_SCOPES = "INSUFFICIENT_SCOPES"
REASON_CODE_MISSING_FACE = "MISSING_FACE"
REASON_CODE_MISSING_DOCUMENT = "MISSING_DOCUMENT"


@dataclass
class DocQualityFeatures:
    """Document quality features matching Go struct."""
    sharpness: float = 0.0
    brightness: float = 0.0
    contrast: float = 0.0
    noise_level: float = 0.0
    blur_score: float = 0.0


@dataclass
class TestVectorInput:
    """Test vector input matching Go struct."""
    face_embedding: List[float] = field(default_factory=list)
    face_confidence: float = 0.0
    doc_quality_score: float = 0.0
    doc_quality_features: DocQualityFeatures = field(default_factory=DocQualityFeatures)
    ocr_confidences: Dict[str, float] = field(default_factory=dict)
    ocr_field_validation: Dict[str, bool] = field(default_factory=dict)
    scope_types: List[str] = field(default_factory=list)
    scope_count: int = 0
    block_height: int = 0


@dataclass 
class TestVectorEntry:
    """Complete test vector entry matching Go struct."""
    name: str = ""
    description: str = ""
    input: TestVectorInput = field(default_factory=TestVectorInput)
    expected_score: float = 0.0
    expected_tier: int = 1
    expected_codes: List[str] = field(default_factory=list)
    expected_confidence: float = 0.0
    tolerance: float = 1e-6


def generate_deterministic_embedding(dim: int, seed: int, scale: float) -> List[float]:
    """
    Generate deterministic embedding using LCG matching Go implementation.
    
    This uses the same linear congruential generator as Go to ensure
    identical embeddings are generated.
    """
    # LCG parameters (must match Go)
    a = 1664525
    c = 1013904223
    m = 4294967296  # 2^32
    
    embedding = []
    state = seed
    
    for _ in range(dim):
        state = (a * state + c) % m
        # Normalize to [-scale, scale]
        normalized = (state / m) * 2 * scale - scale
        embedding.append(float(normalized))
    
    return embedding


def normalize_embedding(embedding: List[float]) -> List[float]:
    """Normalize embedding to unit length matching Go implementation."""
    embedding_array = np.array(embedding, dtype=np.float64)
    norm = np.sqrt(np.sum(embedding_array ** 2))
    
    if norm > 1e-10:
        normalized = embedding_array / norm
        return [float(x) for x in normalized]
    
    return embedding


def extract_features(input_data: TestVectorInput) -> np.ndarray:
    """
    Extract features from input matching Go FeatureExtractor.
    
    Returns 768-dimensional feature vector.
    """
    features = np.zeros(TOTAL_FEATURE_DIM, dtype=np.float32)
    offset = 0
    
    # 1. Face embedding (512 dimensions)
    if len(input_data.face_embedding) == FACE_EMBEDDING_DIM:
        face_emb = np.array(input_data.face_embedding, dtype=np.float32)
        # Normalize to unit length
        norm = np.sqrt(np.sum(face_emb ** 2))
        if norm > 1e-10:
            face_emb = face_emb / norm
        features[offset:offset + FACE_EMBEDDING_DIM] = face_emb
    offset += FACE_EMBEDDING_DIM
    
    # 2. Document quality features (5 dimensions)
    features[offset] = input_data.doc_quality_score
    features[offset + 1] = input_data.doc_quality_features.sharpness
    features[offset + 2] = input_data.doc_quality_features.brightness
    features[offset + 3] = input_data.doc_quality_features.contrast
    features[offset + 4] = 1.0 - input_data.doc_quality_features.noise_level  # Inverted
    offset += DOC_QUALITY_DIM
    
    # 3. OCR features (10 dimensions)
    for i, field_name in enumerate(OCR_FIELD_NAMES):
        base_idx = offset + (i * 2)
        features[base_idx] = input_data.ocr_confidences.get(field_name, 0.0)
        features[base_idx + 1] = 1.0 if input_data.ocr_field_validation.get(field_name, False) else 0.0
    offset += OCR_FIELDS_DIM
    
    # 4. Metadata features (16 dimensions)
    scope_count_norm = min(input_data.scope_count / 10.0, 1.0)
    features[offset] = scope_count_norm
    
    # Scope type indicators
    scope_types = [
        "id_document", "selfie", "face_video", "biometric",
        "sso_metadata", "email_proof", "sms_proof", "domain_verify",
    ]
    scope_set = set(input_data.scope_types)
    for i, st in enumerate(scope_types):
        features[offset + 1 + i] = 1.0 if st in scope_set else 0.0
    
    features[offset + 9] = input_data.face_confidence
    features[offset + 10] = (input_data.block_height % 1000000) / 1000000.0
    
    return features


def compute_score(input_data: TestVectorInput, features: np.ndarray) -> Dict[str, Any]:
    """
    Compute trust score matching Go scorer logic.
    
    This is a simplified scoring function for test vector generation.
    The actual model inference would produce these values.
    """
    # Base score from feature contributions
    face_contrib = 0.0
    if len(input_data.face_embedding) == FACE_EMBEDDING_DIM:
        face_contrib = input_data.face_confidence * 40.0  # Up to 40 points
    
    doc_contrib = input_data.doc_quality_score * 25.0  # Up to 25 points
    
    # OCR contribution
    ocr_sum = sum(input_data.ocr_confidences.values())
    ocr_count = len(input_data.ocr_confidences)
    ocr_avg = ocr_sum / ocr_count if ocr_count > 0 else 0.0
    
    # Validation bonus
    valid_count = sum(1 for v in input_data.ocr_field_validation.values() if v)
    valid_ratio = valid_count / len(input_data.ocr_field_validation) if input_data.ocr_field_validation else 0.0
    
    ocr_contrib = (ocr_avg * 0.6 + valid_ratio * 0.4) * 20.0  # Up to 20 points
    
    # Scope contribution
    scope_contrib = min(input_data.scope_count * 5.0, 15.0)  # Up to 15 points
    
    raw_score = face_contrib + doc_contrib + ocr_contrib + scope_contrib
    raw_score = max(0.0, min(100.0, raw_score))
    
    # Compute confidence
    distance_from_middle = abs(raw_score - 50.0) / 50.0
    confidence = 0.5 + (distance_from_middle * 0.4)
    confidence = max(0.3, min(0.95, confidence))
    
    # Determine tier
    if raw_score >= 80:
        tier = 4
    elif raw_score >= 60:
        tier = 3
    elif raw_score >= 40:
        tier = 2
    else:
        tier = 1
    
    # Determine reason codes
    codes = []
    
    if raw_score >= 50:
        codes.append(REASON_CODE_SUCCESS)
    
    if confidence >= 0.8:
        codes.append(REASON_CODE_HIGH_CONFIDENCE)
    elif confidence < 0.5:
        codes.append(REASON_CODE_LOW_CONFIDENCE)
    
    if len(input_data.face_embedding) == 0:
        codes.append(REASON_CODE_MISSING_FACE)
    
    if input_data.doc_quality_score < 0.6:
        codes.append(REASON_CODE_LOW_DOC_QUALITY)
    
    if ocr_avg < 0.5:
        codes.append(REASON_CODE_LOW_OCR_CONFIDENCE)
    
    if input_data.scope_count < 2:
        codes.append(REASON_CODE_INSUFFICIENT_SCOPES)
    
    return {
        "score": round(raw_score, 1),
        "confidence": round(confidence, 3),
        "tier": tier,
        "codes": codes,
    }


def generate_test_vectors() -> List[TestVectorEntry]:
    """Generate a comprehensive set of test vectors with known outputs."""
    vectors = []
    
    # Vector 1: High-quality verification
    input1 = TestVectorInput(
        face_embedding=generate_deterministic_embedding(512, 42, 0.1),
        face_confidence=0.95,
        doc_quality_score=0.92,
        doc_quality_features=DocQualityFeatures(
            sharpness=0.88,
            brightness=0.75,
            contrast=0.82,
            noise_level=0.08,
            blur_score=0.05,
        ),
        ocr_confidences={
            "name": 0.95,
            "date_of_birth": 0.92,
            "document_number": 0.98,
            "expiry_date": 0.89,
            "nationality": 0.91,
        },
        ocr_field_validation={
            "name": True,
            "date_of_birth": True,
            "document_number": True,
            "expiry_date": True,
            "nationality": True,
        },
        scope_types=["id_document", "selfie", "face_video"],
        scope_count=3,
        block_height=1000000,
    )
    features1 = extract_features(input1)
    result1 = compute_score(input1, features1)
    
    vectors.append(TestVectorEntry(
        name="high_quality_verification",
        description="High-quality face, document, and OCR - should produce high score",
        input=input1,
        expected_score=result1["score"],
        expected_tier=result1["tier"],
        expected_codes=result1["codes"],
        expected_confidence=result1["confidence"],
        tolerance=1e-6,
    ))
    
    # Vector 2: Medium-quality verification
    input2 = TestVectorInput(
        face_embedding=generate_deterministic_embedding(512, 43, 0.15),
        face_confidence=0.82,
        doc_quality_score=0.75,
        doc_quality_features=DocQualityFeatures(
            sharpness=0.70,
            brightness=0.68,
            contrast=0.72,
            noise_level=0.15,
            blur_score=0.12,
        ),
        ocr_confidences={
            "name": 0.78,
            "date_of_birth": 0.72,
            "document_number": 0.85,
            "expiry_date": 0.70,
            "nationality": 0.75,
        },
        ocr_field_validation={
            "name": True,
            "date_of_birth": True,
            "document_number": True,
            "expiry_date": False,
            "nationality": True,
        },
        scope_types=["id_document", "selfie"],
        scope_count=2,
        block_height=1500000,
    )
    features2 = extract_features(input2)
    result2 = compute_score(input2, features2)
    
    vectors.append(TestVectorEntry(
        name="medium_quality_verification",
        description="Medium quality inputs - should produce moderate score",
        input=input2,
        expected_score=result2["score"],
        expected_tier=result2["tier"],
        expected_codes=result2["codes"],
        expected_confidence=result2["confidence"],
        tolerance=1e-6,
    ))
    
    # Vector 3: Low-quality document
    input3 = TestVectorInput(
        face_embedding=generate_deterministic_embedding(512, 44, 0.12),
        face_confidence=0.88,
        doc_quality_score=0.45,
        doc_quality_features=DocQualityFeatures(
            sharpness=0.35,
            brightness=0.50,
            contrast=0.40,
            noise_level=0.35,
            blur_score=0.40,
        ),
        ocr_confidences={
            "name": 0.55,
            "date_of_birth": 0.48,
            "document_number": 0.60,
            "expiry_date": 0.42,
            "nationality": 0.50,
        },
        ocr_field_validation={
            "name": True,
            "date_of_birth": False,
            "document_number": True,
            "expiry_date": False,
            "nationality": True,
        },
        scope_types=["id_document"],
        scope_count=1,
        block_height=2000000,
    )
    features3 = extract_features(input3)
    result3 = compute_score(input3, features3)
    
    vectors.append(TestVectorEntry(
        name="low_quality_document",
        description="Low document quality - should produce lower score with reason codes",
        input=input3,
        expected_score=result3["score"],
        expected_tier=result3["tier"],
        expected_codes=result3["codes"],
        expected_confidence=result3["confidence"],
        tolerance=1e-6,
    ))
    
    # Vector 4: Missing face embedding
    input4 = TestVectorInput(
        face_embedding=[],  # Empty
        face_confidence=0.0,
        doc_quality_score=0.85,
        doc_quality_features=DocQualityFeatures(
            sharpness=0.82,
            brightness=0.78,
            contrast=0.80,
            noise_level=0.10,
            blur_score=0.08,
        ),
        ocr_confidences={
            "name": 0.90,
            "date_of_birth": 0.88,
            "document_number": 0.92,
            "expiry_date": 0.85,
            "nationality": 0.87,
        },
        ocr_field_validation={
            "name": True,
            "date_of_birth": True,
            "document_number": True,
            "expiry_date": True,
            "nationality": True,
        },
        scope_types=["id_document"],
        scope_count=1,
        block_height=2500000,
    )
    features4 = extract_features(input4)
    result4 = compute_score(input4, features4)
    
    vectors.append(TestVectorEntry(
        name="missing_face_embedding",
        description="No face embedding provided - should handle gracefully",
        input=input4,
        expected_score=result4["score"],
        expected_tier=result4["tier"],
        expected_codes=result4["codes"],
        expected_confidence=result4["confidence"],
        tolerance=1e-6,
    ))
    
    # Vector 5: Perfect score scenario
    input5 = TestVectorInput(
        face_embedding=generate_deterministic_embedding(512, 45, 0.08),
        face_confidence=0.99,
        doc_quality_score=0.98,
        doc_quality_features=DocQualityFeatures(
            sharpness=0.95,
            brightness=0.85,
            contrast=0.92,
            noise_level=0.02,
            blur_score=0.01,
        ),
        ocr_confidences={
            "name": 0.99,
            "date_of_birth": 0.98,
            "document_number": 0.99,
            "expiry_date": 0.97,
            "nationality": 0.98,
        },
        ocr_field_validation={
            "name": True,
            "date_of_birth": True,
            "document_number": True,
            "expiry_date": True,
            "nationality": True,
        },
        scope_types=["id_document", "selfie", "face_video", "biometric"],
        scope_count=4,
        block_height=3000000,
    )
    features5 = extract_features(input5)
    result5 = compute_score(input5, features5)
    
    vectors.append(TestVectorEntry(
        name="perfect_verification",
        description="All inputs at maximum quality - should produce near-perfect score",
        input=input5,
        expected_score=result5["score"],
        expected_tier=result5["tier"],
        expected_codes=result5["codes"],
        expected_confidence=result5["confidence"],
        tolerance=1e-6,
    ))
    
    # Vector 6: Edge case - minimal valid input
    input6 = TestVectorInput(
        face_embedding=generate_deterministic_embedding(512, 46, 0.2),
        face_confidence=0.50,
        doc_quality_score=0.50,
        doc_quality_features=DocQualityFeatures(
            sharpness=0.50,
            brightness=0.50,
            contrast=0.50,
            noise_level=0.50,
            blur_score=0.50,
        ),
        ocr_confidences={
            "name": 0.50,
            "date_of_birth": 0.50,
            "document_number": 0.50,
            "expiry_date": 0.50,
            "nationality": 0.50,
        },
        ocr_field_validation={
            "name": True,
            "date_of_birth": False,
            "document_number": True,
            "expiry_date": False,
            "nationality": False,
        },
        scope_types=["id_document", "selfie"],
        scope_count=2,
        block_height=500000,
    )
    features6 = extract_features(input6)
    result6 = compute_score(input6, features6)
    
    vectors.append(TestVectorEntry(
        name="minimal_valid_input",
        description="Bare minimum valid inputs",
        input=input6,
        expected_score=result6["score"],
        expected_tier=result6["tier"],
        expected_codes=result6["codes"],
        expected_confidence=result6["confidence"],
        tolerance=1e-6,
    ))
    
    # Vector 7: Passport verification
    input7 = TestVectorInput(
        face_embedding=generate_deterministic_embedding(512, 47, 0.11),
        face_confidence=0.91,
        doc_quality_score=0.88,
        doc_quality_features=DocQualityFeatures(
            sharpness=0.85,
            brightness=0.80,
            contrast=0.82,
            noise_level=0.08,
            blur_score=0.06,
        ),
        ocr_confidences={
            "name": 0.92,
            "date_of_birth": 0.89,
            "document_number": 0.95,
            "expiry_date": 0.88,
            "nationality": 0.93,
        },
        ocr_field_validation={
            "name": True,
            "date_of_birth": True,
            "document_number": True,
            "expiry_date": True,
            "nationality": True,
        },
        scope_types=["id_document", "selfie", "face_video"],
        scope_count=3,
        block_height=1800000,
    )
    features7 = extract_features(input7)
    result7 = compute_score(input7, features7)
    
    vectors.append(TestVectorEntry(
        name="passport_verification",
        description="Passport verification with good quality",
        input=input7,
        expected_score=result7["score"],
        expected_tier=result7["tier"],
        expected_codes=result7["codes"],
        expected_confidence=result7["confidence"],
        tolerance=1e-6,
    ))
    
    # Vector 8: Low face confidence
    input8 = TestVectorInput(
        face_embedding=generate_deterministic_embedding(512, 48, 0.25),
        face_confidence=0.55,
        doc_quality_score=0.85,
        doc_quality_features=DocQualityFeatures(
            sharpness=0.82,
            brightness=0.78,
            contrast=0.80,
            noise_level=0.10,
            blur_score=0.08,
        ),
        ocr_confidences={
            "name": 0.88,
            "date_of_birth": 0.85,
            "document_number": 0.90,
            "expiry_date": 0.82,
            "nationality": 0.86,
        },
        ocr_field_validation={
            "name": True,
            "date_of_birth": True,
            "document_number": True,
            "expiry_date": True,
            "nationality": True,
        },
        scope_types=["id_document", "selfie"],
        scope_count=2,
        block_height=2200000,
    )
    features8 = extract_features(input8)
    result8 = compute_score(input8, features8)
    
    vectors.append(TestVectorEntry(
        name="low_face_confidence",
        description="Good document but low face confidence",
        input=input8,
        expected_score=result8["score"],
        expected_tier=result8["tier"],
        expected_codes=result8["codes"],
        expected_confidence=result8["confidence"],
        tolerance=1e-6,
    ))
    
    # Vector 9: All OCR validations failed
    input9 = TestVectorInput(
        face_embedding=generate_deterministic_embedding(512, 49, 0.10),
        face_confidence=0.90,
        doc_quality_score=0.82,
        doc_quality_features=DocQualityFeatures(
            sharpness=0.80,
            brightness=0.75,
            contrast=0.78,
            noise_level=0.12,
            blur_score=0.10,
        ),
        ocr_confidences={
            "name": 0.85,
            "date_of_birth": 0.82,
            "document_number": 0.88,
            "expiry_date": 0.80,
            "nationality": 0.83,
        },
        ocr_field_validation={
            "name": False,
            "date_of_birth": False,
            "document_number": False,
            "expiry_date": False,
            "nationality": False,
        },
        scope_types=["id_document", "selfie", "face_video"],
        scope_count=3,
        block_height=2800000,
    )
    features9 = extract_features(input9)
    result9 = compute_score(input9, features9)
    
    vectors.append(TestVectorEntry(
        name="ocr_validation_failure",
        description="High confidence OCR but all validations fail",
        input=input9,
        expected_score=result9["score"],
        expected_tier=result9["tier"],
        expected_codes=result9["codes"],
        expected_confidence=result9["confidence"],
        tolerance=1e-6,
    ))
    
    # Vector 10: Biometric-only verification
    input10 = TestVectorInput(
        face_embedding=generate_deterministic_embedding(512, 50, 0.09),
        face_confidence=0.97,
        doc_quality_score=0.60,
        doc_quality_features=DocQualityFeatures(
            sharpness=0.55,
            brightness=0.60,
            contrast=0.58,
            noise_level=0.20,
            blur_score=0.18,
        ),
        ocr_confidences={
            "name": 0.60,
            "date_of_birth": 0.55,
            "document_number": 0.65,
            "expiry_date": 0.50,
            "nationality": 0.58,
        },
        ocr_field_validation={
            "name": True,
            "date_of_birth": True,
            "document_number": True,
            "expiry_date": False,
            "nationality": True,
        },
        scope_types=["biometric", "face_video"],
        scope_count=2,
        block_height=3200000,
    )
    features10 = extract_features(input10)
    result10 = compute_score(input10, features10)
    
    vectors.append(TestVectorEntry(
        name="biometric_only",
        description="Strong biometric with minimal document",
        input=input10,
        expected_score=result10["score"],
        expected_tier=result10["tier"],
        expected_codes=result10["codes"],
        expected_confidence=result10["confidence"],
        tolerance=1e-6,
    ))
    
    return vectors


def dataclass_to_dict(obj) -> Dict[str, Any]:
    """Convert dataclass to dictionary with proper nesting."""
    if hasattr(obj, '__dataclass_fields__'):
        result = {}
        for field_name in obj.__dataclass_fields__:
            value = getattr(obj, field_name)
            result[field_name] = dataclass_to_dict(value)
        return result
    elif isinstance(obj, list):
        return [dataclass_to_dict(item) for item in obj]
    elif isinstance(obj, dict):
        return {k: dataclass_to_dict(v) for k, v in obj.items()}
    else:
        return obj


def compute_vectors_hash(vectors: List[TestVectorEntry]) -> str:
    """Compute SHA256 hash of all test vectors for version tracking."""
    data = json.dumps([dataclass_to_dict(v) for v in vectors], sort_keys=True)
    return hashlib.sha256(data.encode()).hexdigest()


def main():
    """Main entry point for test vector generation."""
    print("=" * 60)
    print("VE-3006: Generating Go-Python Conformance Test Vectors")
    print("=" * 60)
    
    start_time = time.time()
    
    # Generate vectors
    vectors = generate_test_vectors()
    
    # Convert to serializable format
    vectors_dict = [dataclass_to_dict(v) for v in vectors]
    
    # Compute hash
    vectors_hash = compute_vectors_hash(vectors)
    
    # Create output with metadata
    output = {
        "version": "1.0.0",
        "generated_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        "random_seed": 42,
        "vectors_hash": vectors_hash,
        "vector_count": len(vectors),
        "feature_dimensions": {
            "face_embedding": FACE_EMBEDDING_DIM,
            "doc_quality": DOC_QUALITY_DIM,
            "ocr_fields": OCR_FIELDS_DIM,
            "metadata": METADATA_DIM,
            "total": TOTAL_FEATURE_DIM,
        },
        "vectors": vectors_dict,
    }
    
    # Write to file
    output_path = Path(__file__).parent / "test_vectors.json"
    with open(output_path, "w") as f:
        json.dump(output, f, indent=2)
    
    elapsed_time = time.time() - start_time
    
    print(f"\nGeneration Summary:")
    print(f"  - Generated {len(vectors)} test vectors")
    print(f"  - Random seed: 42")
    print(f"  - Vectors hash: {vectors_hash[:16]}...")
    print(f"  - Output: {output_path}")
    print(f"  - Time: {elapsed_time:.2f}s")
    print()
    
    # Print vector summary
    print("Test Vectors:")
    for i, v in enumerate(vectors, 1):
        print(f"  {i:2d}. {v.name}: score={v.expected_score:.1f}, tier={v.expected_tier}")
    
    print("\n" + "=" * 60)
    print("Test vector generation complete!")
    print("=" * 60)


if __name__ == "__main__":
    main()
