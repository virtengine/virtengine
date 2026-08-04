"""Canonical trust-score feature assembly shared with Go inference."""

from __future__ import annotations

import hashlib
import json
from dataclasses import dataclass
from pathlib import Path
from typing import Mapping, Optional, Sequence

import numpy as np


REPO_ROOT = Path(__file__).resolve().parents[3]
SCHEMA_PATH = REPO_ROOT / "pkg/inference/schema/trust_score_features_v1.json"
SCHEMA_BYTES = SCHEMA_PATH.read_bytes()
SCHEMA_SHA256 = hashlib.sha256(SCHEMA_BYTES).hexdigest()
SCHEMA = json.loads(SCHEMA_BYTES)

OCR_FIELD_NAMES = (
    "name",
    "date_of_birth",
    "document_number",
    "expiry_date",
    "nationality",
)
METADATA_SCOPE_TYPES = (
    "id_document",
    "selfie",
    "face_video",
    "biometric",
    "sso_metadata",
    "email_proof",
    "sms_proof",
    "domain_verify",
)

_EXPECTED_SEGMENTS = [
    {"name": "selfie_embedding", "offset": 0, "dimensions": 512, "required": True, "transform": "l2_normalize"},
    {"name": "face_confidence", "offset": 512, "dimensions": 1, "required": True},
    {"name": "document_quality", "offset": 513, "dimensions": 1, "required": True},
    {"name": "document_sharpness", "offset": 514, "dimensions": 1, "required": True},
    {"name": "document_brightness", "offset": 515, "dimensions": 1, "required": True},
    {"name": "document_contrast", "offset": 516, "dimensions": 1, "required": True},
    {"name": "document_noise_quality", "offset": 517, "dimensions": 1, "required": True, "transform": "one_minus"},
    {"name": "document_blur_quality", "offset": 518, "dimensions": 1, "required": True, "transform": "one_minus"},
    {"name": "ocr_name", "offset": 519, "dimensions": 2, "required": True, "values": ["confidence", "validated"]},
    {"name": "ocr_date_of_birth", "offset": 521, "dimensions": 2, "required": True, "values": ["confidence", "validated"]},
    {"name": "ocr_document_number", "offset": 523, "dimensions": 2, "required": True, "values": ["confidence", "validated"]},
    {"name": "ocr_expiry_date", "offset": 525, "dimensions": 2, "required": True, "values": ["confidence", "validated"]},
    {"name": "ocr_nationality", "offset": 527, "dimensions": 2, "required": True, "values": ["confidence", "validated"]},
    {
        "name": "metadata",
        "offset": 529,
        "dimensions": 16,
        "required": True,
        "values": [
            "scope_count", "scope_id_document", "scope_selfie", "scope_face_video",
            "scope_biometric", "scope_sso_metadata", "scope_email_proof",
            "scope_sms_proof", "scope_domain_verify", "block_height_mod_1000000",
            "reserved_0", "reserved_1", "reserved_2", "reserved_3", "reserved_4",
            "reserved_5",
        ],
    },
    {"name": "reserved", "offset": 545, "dimensions": 223, "required": False, "value": 0},
]


def _validate_schema() -> None:
    expected = {
        "schema_version": "virtengine.trust-score-features/v1",
        "dtype": "float32",
        "byte_order": "big-endian",
        "total_dimensions": 768,
        "segments": _EXPECTED_SEGMENTS,
    }
    if SCHEMA != expected:
        raise ValueError(f"canonical feature schema drift: {SCHEMA_PATH}")


_validate_schema()
_SEGMENTS = {segment["name"]: segment for segment in SCHEMA["segments"]}

FACE_EMBEDDING_DIM = _SEGMENTS["selfie_embedding"]["dimensions"]
FACE_OFFSET = _SEGMENTS["selfie_embedding"]["offset"]
FACE_CONFIDENCE_OFFSET = _SEGMENTS["face_confidence"]["offset"]
DOC_QUALITY_OFFSET = _SEGMENTS["document_quality"]["offset"]
DOC_QUALITY_DIM = 6
OCR_OFFSET = _SEGMENTS["ocr_name"]["offset"]
OCR_FIELDS_DIM = len(OCR_FIELD_NAMES) * 2
METADATA_OFFSET = _SEGMENTS["metadata"]["offset"]
METADATA_DIM = _SEGMENTS["metadata"]["dimensions"]
PADDING_OFFSET = _SEGMENTS["reserved"]["offset"]
PADDING_DIM = _SEGMENTS["reserved"]["dimensions"]
TOTAL_FEATURE_DIM = SCHEMA["total_dimensions"]


@dataclass(frozen=True)
class CanonicalInferenceInputs:
    """Source values consumed by the canonical Go and Python assemblers."""

    face_embedding: Optional[Sequence[float]] = None
    face_confidence: Optional[float] = None
    doc_quality_score: Optional[float] = None
    sharpness: Optional[float] = None
    brightness: Optional[float] = None
    contrast: Optional[float] = None
    noise_level: Optional[float] = None
    blur_score: Optional[float] = None
    ocr_confidences: Optional[Mapping[str, float]] = None
    ocr_field_validation: Optional[Mapping[str, bool]] = None
    scope_types: Optional[Sequence[str]] = None
    scope_count: Optional[int] = None
    block_height: Optional[int] = None


def _unit_scalar(value: Optional[float], name: str, strict: bool) -> np.float32:
    if value is None:
        if strict:
            raise ValueError(f"missing required input: {name}")
        value = 0.0
    converted = np.float32(value)
    if not np.isfinite(converted):
        raise ValueError(f"{name} must be finite")
    if converted < 0 or converted > 1:
        raise ValueError(f"{name} must be in [0,1]")
    return converted


def _integer(value: Optional[int], name: str, strict: bool) -> int:
    if value is None:
        if strict:
            raise ValueError(f"missing required input: {name}")
        value = 0
    converted = int(value)
    if converted != value:
        raise ValueError(f"{name} must be an integer")
    return converted


def assemble_canonical_inference_vector(
    inputs: CanonicalInferenceInputs,
    *,
    strict_production: bool = False,
) -> np.ndarray:
    """Build the exact 768-position canonical vector."""
    if inputs is None:
        raise ValueError("canonical inputs are required")

    vector = np.zeros(TOTAL_FEATURE_DIM, dtype=np.float32)
    if inputs.face_embedding is None:
        if strict_production:
            raise ValueError("missing required input: face_embedding")
        embedding = np.array([], dtype=np.float32)
    else:
        embedding = np.asarray(inputs.face_embedding, dtype=np.float32).reshape(-1)
    if embedding.size not in (0, FACE_EMBEDDING_DIM):
        raise ValueError(
            f"face embedding dimension mismatch: expected {FACE_EMBEDDING_DIM}, got {embedding.size}"
        )
    if strict_production and embedding.size == 0:
        raise ValueError("missing required input: face_embedding")
    if not np.all(np.isfinite(embedding)):
        raise ValueError("face_embedding must contain only finite values")
    if embedding.size:
        sum_squares = 0.0
        for value in embedding:
            converted = float(value)
            sum_squares += converted * converted
        norm = np.sqrt(sum_squares)
        if strict_production and norm <= 1e-10:
            raise ValueError("face_embedding must have nonzero norm")
        if norm > 1e-10:
            embedding = np.array(
                [np.float32(float(value) / norm) for value in embedding],
                dtype=np.float32,
            )
        vector[FACE_OFFSET:FACE_CONFIDENCE_OFFSET] = embedding

    vector[FACE_CONFIDENCE_OFFSET] = _unit_scalar(
        inputs.face_confidence, "face_confidence", strict_production
    )
    document_values = (
        (inputs.doc_quality_score, "doc_quality_score", False),
        (inputs.sharpness, "sharpness", False),
        (inputs.brightness, "brightness", False),
        (inputs.contrast, "contrast", False),
        (inputs.noise_level, "noise_level", True),
        (inputs.blur_score, "blur_score", True),
    )
    for index, (value, name, invert) in enumerate(document_values):
        converted = _unit_scalar(value, name, strict_production)
        vector[DOC_QUALITY_OFFSET + index] = np.float32(1.0) - converted if invert else converted

    confidences = inputs.ocr_confidences
    validations = inputs.ocr_field_validation
    if strict_production and confidences is None:
        raise ValueError("missing required input: ocr_confidences")
    if strict_production and validations is None:
        raise ValueError("missing required input: ocr_field_validation")
    confidences = confidences or {}
    validations = validations or {}
    for index, field_name in enumerate(OCR_FIELD_NAMES):
        if strict_production and field_name not in confidences:
            raise ValueError(f"missing OCR confidence for {field_name}")
        if strict_production and field_name not in validations:
            raise ValueError(f"missing OCR validation for {field_name}")
        base = OCR_OFFSET + index * 2
        vector[base] = _unit_scalar(
            confidences.get(field_name), f"ocr_confidences[{field_name}]", False
        )
        vector[base + 1] = np.float32(bool(validations.get(field_name, False)))

    if strict_production and inputs.scope_types is None:
        raise ValueError("missing required input: scope_types")
    scope_count = _integer(inputs.scope_count, "scope_count", strict_production)
    if scope_count < 0 or (strict_production and scope_count > 10):
        raise ValueError("scope_count must be in [0,10]")
    vector[METADATA_OFFSET] = min(np.float32(1.0), np.float32(scope_count) / np.float32(10.0))
    scope_type_set = set(inputs.scope_types or ())
    for index, scope_type in enumerate(METADATA_SCOPE_TYPES):
        vector[METADATA_OFFSET + 1 + index] = np.float32(scope_type in scope_type_set)
    block_height = _integer(inputs.block_height, "block_height", strict_production)
    if strict_production and block_height < 0:
        raise ValueError("block_height must be nonnegative")
    height_remainder = block_height % 1_000_000 if block_height >= 0 else -((-block_height) % 1_000_000)
    vector[METADATA_OFFSET + 9] = np.float32(height_remainder) / np.float32(1_000_000)

    if vector.shape != (TOTAL_FEATURE_DIM,) or not np.all(np.isfinite(vector)):
        raise ValueError("canonical inference vector is invalid")
    return vector