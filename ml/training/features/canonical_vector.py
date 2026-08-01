"""Canonical inference feature-vector assembly shared with Go inference."""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Mapping, Sequence

import numpy as np


FACE_EMBEDDING_DIM = 512
DOC_QUALITY_DIM = 5
OCR_FIELD_NAMES = (
    "name",
    "date_of_birth",
    "document_number",
    "expiry_date",
    "nationality",
)
OCR_FIELDS_DIM = 10
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
METADATA_DIM = 16
TOTAL_FEATURE_DIM = 768

FACE_OFFSET = 0
DOC_QUALITY_OFFSET = FACE_OFFSET + FACE_EMBEDDING_DIM
OCR_OFFSET = DOC_QUALITY_OFFSET + DOC_QUALITY_DIM
METADATA_OFFSET = OCR_OFFSET + OCR_FIELDS_DIM
PADDING_OFFSET = METADATA_OFFSET + METADATA_DIM
PADDING_DIM = TOTAL_FEATURE_DIM - PADDING_OFFSET


@dataclass(frozen=True)
class CanonicalInferenceInputs:
    """Source values consumed by the Go canonical feature extractor."""

    face_embedding: Sequence[float] = field(default_factory=tuple)
    face_confidence: float = 0.0
    doc_quality_score: float = 0.0
    sharpness: float = 0.0
    brightness: float = 0.0
    contrast: float = 0.0
    noise_level: float = 0.0
    ocr_confidences: Mapping[str, float] = field(default_factory=dict)
    ocr_field_validation: Mapping[str, bool] = field(default_factory=dict)
    scope_types: Sequence[str] = field(default_factory=tuple)
    scope_count: int = 0
    block_height: int = 0


def _finite_float32(value: float, name: str) -> np.float32:
    converted = np.float32(value)
    if not np.isfinite(converted):
        raise ValueError(f"{name} must be finite")
    return converted


def _integer(value: int, name: str) -> int:
    converted = int(value)
    if converted != value:
        raise ValueError(f"{name} must be an integer")
    return converted


def assemble_canonical_inference_vector(
    inputs: CanonicalInferenceInputs,
) -> np.ndarray:
    """Build the exact 768-position vector emitted by Go FeatureExtractor."""
    vector = np.zeros(TOTAL_FEATURE_DIM, dtype=np.float32)

    embedding = np.asarray(inputs.face_embedding, dtype=np.float32).reshape(-1)
    if embedding.size not in (0, FACE_EMBEDDING_DIM):
        raise ValueError(
            "face embedding dimension mismatch: "
            f"expected {FACE_EMBEDDING_DIM}, got {embedding.size}"
        )
    if not np.all(np.isfinite(embedding)):
        raise ValueError("face embedding must contain only finite values")
    if embedding.size:
        sum_squares = 0.0
        for value in embedding:
            converted = float(value)
            sum_squares += converted * converted
        norm = np.sqrt(sum_squares)
        if norm > 1e-10:
            embedding = np.array(
                [np.float32(float(value) / norm) for value in embedding],
                dtype=np.float32,
            )
        vector[FACE_OFFSET:DOC_QUALITY_OFFSET] = embedding

    vector[DOC_QUALITY_OFFSET:DOC_QUALITY_OFFSET + DOC_QUALITY_DIM] = [
        _finite_float32(inputs.doc_quality_score, "doc_quality_score"),
        _finite_float32(inputs.sharpness, "sharpness"),
        _finite_float32(inputs.brightness, "brightness"),
        _finite_float32(inputs.contrast, "contrast"),
        np.float32(1.0) - _finite_float32(inputs.noise_level, "noise_level"),
    ]

    for index, field_name in enumerate(OCR_FIELD_NAMES):
        base = OCR_OFFSET + index * 2
        vector[base] = _finite_float32(
            inputs.ocr_confidences.get(field_name, 0.0),
            f"ocr_confidences[{field_name}]",
        )
        vector[base + 1] = np.float32(
            bool(inputs.ocr_field_validation.get(field_name, False))
        )

    scope_count = _finite_float32(_integer(inputs.scope_count, "scope_count"), "scope_count")
    vector[METADATA_OFFSET] = min(np.float32(1.0), scope_count / np.float32(10.0))
    scope_type_set = set(inputs.scope_types)
    for index, scope_type in enumerate(METADATA_SCOPE_TYPES):
        vector[METADATA_OFFSET + 1 + index] = np.float32(
            scope_type in scope_type_set
        )
    vector[METADATA_OFFSET + 9] = _finite_float32(
        inputs.face_confidence, "face_confidence"
    )
    block_height = _integer(inputs.block_height, "block_height")
    height_remainder = (
        block_height % 1_000_000
        if block_height >= 0
        else -((-block_height) % 1_000_000)
    )
    vector[METADATA_OFFSET + 10] = (
        np.float32(height_remainder) / np.float32(1_000_000)
    )

    if vector.shape != (TOTAL_FEATURE_DIM,) or not np.all(np.isfinite(vector)):
        raise ValueError("canonical inference vector is invalid")
    return vector