"""Generate and verify the canonical Python/Go feature parity fixture."""

from __future__ import annotations

import argparse
import hashlib
import json
import math
import struct
import sys
from pathlib import Path
from typing import Any

import numpy as np

from ml.training.features.canonical_vector import (
    CanonicalInferenceInputs,
    DOC_QUALITY_DIM,
    DOC_QUALITY_OFFSET,
    FACE_CONFIDENCE_OFFSET,
    FACE_EMBEDDING_DIM,
    FACE_OFFSET,
    METADATA_DIM,
    METADATA_OFFSET,
    METADATA_SCOPE_TYPES,
    OCR_FIELD_NAMES,
    OCR_FIELDS_DIM,
    OCR_OFFSET,
    PADDING_DIM,
    PADDING_OFFSET,
    SCHEMA_SHA256,
    TOTAL_FEATURE_DIM,
    assemble_canonical_inference_vector,
)


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "pkg/inference/conformance/testdata/feature_parity_v1.json"


def expand_face_embedding(spec: dict[str, Any]) -> np.ndarray:
    """Expand a compact deterministic embedding source."""
    kind = spec["kind"]
    if kind == "missing":
        return np.array([], dtype=np.float32)
    dimension = int(spec["dimension"])
    if kind == "single_index":
        result = np.zeros(dimension, dtype=np.float32)
        result[int(spec["index"])] = np.float32(spec["value"])
        return result
    if kind == "alternating_bounds":
        low, high = np.float32(spec["low"]), np.float32(spec["high"])
        return np.array(
            [low if index % 2 == 0 else high for index in range(dimension)],
            dtype=np.float32,
        )
    if kind == "asymmetric_sequence":
        start, step = np.float32(spec["start"]), np.float32(spec["step"])
        return np.array(
            [np.float32(start + np.float32(index) * step) for index in range(dimension)],
            dtype=np.float32,
        )
    raise ValueError(f"unsupported face embedding source kind: {kind}")


def inputs_from_source(source: dict[str, Any]) -> CanonicalInferenceInputs:
    """Convert fixture source values without consuming expected output data."""
    document = source.get("document_quality", {})
    return CanonicalInferenceInputs(
        face_embedding=expand_face_embedding(source["face_embedding"]),
        face_confidence=source.get("face_confidence"),
        doc_quality_score=source.get("doc_quality_score"),
        sharpness=document.get("sharpness"),
        brightness=document.get("brightness"),
        contrast=document.get("contrast"),
        noise_level=document.get("noise_level"),
        blur_score=document.get("blur_score"),
        ocr_confidences=source.get("ocr_confidences"),
        ocr_field_validation=source.get("ocr_field_validation"),
        scope_types=source.get("scope_types"),
        scope_count=source.get("scope_count"),
        block_height=source.get("block_height"),
    )


def _round_half_away_from_zero(value: float, places: int) -> float:
    multiplier = 10**places
    scaled = value * multiplier
    rounded = math.floor(scaled + 0.5) if scaled >= 0 else math.ceil(scaled - 0.5)
    return rounded / multiplier


def feature_hash(vector: np.ndarray, places: int = 6) -> str:
    """Match DeterminismController.ComputeFeatureHash byte-for-byte."""
    digest = hashlib.sha256()
    for value in vector:
        rounded = np.float32(_round_half_away_from_zero(float(np.float32(value)), places))
        digest.update(struct.pack(">f", rounded))
    return digest.hexdigest()


def raw_feature_hash(vector: np.ndarray) -> str:
    """Hash exact big-endian IEEE-754 float32 bytes without rounding."""
    digest = hashlib.sha256()
    for value in vector:
        digest.update(struct.pack(">f", np.float32(value)))
    return digest.hexdigest()


def _complete_source() -> dict[str, Any]:
    return {
        "face_embedding": {"kind": "single_index", "dimension": 512, "index": 0, "value": 1.0},
        "face_confidence": 0.0,
        "doc_quality_score": 0.0,
        "document_quality": {"sharpness": 0.0, "brightness": 0.0, "contrast": 0.0, "noise_level": 0.0, "blur_score": 0.0},
        "ocr_confidences": {name: 0.0 for name in OCR_FIELD_NAMES},
        "ocr_field_validation": {name: False for name in OCR_FIELD_NAMES},
        "scope_types": [],
        "scope_count": 0,
        "block_height": 0,
    }


def _cases() -> list[dict[str, Any]]:
    sentinel = _complete_source()
    sentinel.update({
        "face_embedding": {"kind": "single_index", "dimension": 512, "index": 511, "value": -2.0},
        "face_confidence": 0.91,
        "doc_quality_score": 0.11,
        "document_quality": {"sharpness": 0.22, "brightness": 0.33, "contrast": 0.44, "noise_level": 0.45, "blur_score": 0.27},
        "ocr_confidences": {"name": 0.61, "date_of_birth": 0.62, "document_number": 0.63, "expiry_date": 0.64, "nationality": 0.65},
        "ocr_field_validation": {"name": True, "date_of_birth": False, "document_number": True, "expiry_date": False, "nationality": True},
        "scope_types": ["id_document", "face_video", "sso_metadata", "sms_proof"],
        "scope_count": 7,
        "block_height": 1234567,
    })
    boundary = _complete_source()
    boundary.update({
        "face_embedding": {"kind": "alternating_bounds", "dimension": 512, "low": -1.0, "high": 1.0},
        "face_confidence": 1.0,
        "doc_quality_score": 1.0,
        "document_quality": {"sharpness": 0.0, "brightness": 1.0, "contrast": 0.0, "noise_level": 1.0, "blur_score": 0.0},
        "ocr_confidences": {"name": 0.0, "date_of_birth": 1.0, "document_number": 0.0, "expiry_date": 1.0, "nationality": 0.0},
        "ocr_field_validation": {"name": False, "date_of_birth": True, "document_number": False, "expiry_date": True, "nationality": False},
        "scope_types": list(METADATA_SCOPE_TYPES),
        "scope_count": 10,
        "block_height": 999999,
    })
    negative = _complete_source()
    negative.update({
        "face_embedding": {"kind": "asymmetric_sequence", "dimension": 512, "start": -0.73, "step": 0.0031},
        "face_confidence": 0.37,
        "document_quality": {"sharpness": 0.0, "brightness": 0.0, "contrast": 0.0, "noise_level": 1.0, "blur_score": 1.0},
        "scope_count": 3,
        "block_height": -1234567,
    })
    return [
        {"name": "explicit_development_missing", "profile": "development", "source": {"face_embedding": {"kind": "missing"}}},
        {"name": "unique_position_sentinel", "profile": "production", "source": sentinel},
        {"name": "boundary_values", "profile": "production", "source": boundary},
        {"name": "asymmetric_negative_height", "profile": "development", "source": negative},
    ]


def build_fixture() -> dict[str, Any]:
    """Build the complete fixture deterministically."""
    fixture = {
        "schema": "virtengine.inference.feature_parity",
        "version": 1,
        "contract": "fixture_only",
        "feature_schema_sha256": SCHEMA_SHA256,
        "layout": {
            "total_dimension": TOTAL_FEATURE_DIM,
            "components": [
                {"name": "selfie_embedding", "offset": FACE_OFFSET, "dimension": FACE_EMBEDDING_DIM},
                {"name": "face_confidence", "offset": FACE_CONFIDENCE_OFFSET, "dimension": 1},
                {"name": "document_quality", "offset": DOC_QUALITY_OFFSET, "dimension": DOC_QUALITY_DIM},
                {"name": "ocr", "offset": OCR_OFFSET, "dimension": OCR_FIELDS_DIM, "fields": list(OCR_FIELD_NAMES)},
                {"name": "metadata", "offset": METADATA_OFFSET, "dimension": METADATA_DIM},
                {"name": "reserved", "offset": PADDING_OFFSET, "dimension": PADDING_DIM},
            ],
            "encoding": {"value_type": "ieee754-float32", "byte_order": "big-endian", "pre_hash_decimal_places": 6, "rounding": "half_away_from_zero", "hash_algorithm": "sha256"},
        },
        "cases": _cases(),
    }
    for case in fixture["cases"]:
        vector = assemble_canonical_inference_vector(
            inputs_from_source(case["source"]),
            strict_production=case["profile"] == "production",
        )
        case["expected_vector_hash"] = feature_hash(vector)
        case["expected_raw_vector_hash"] = raw_feature_hash(vector)
        case["expected_nonzero_positions"] = {
            str(index): float(value) for index, value in enumerate(vector) if value != 0.0
        }
    return fixture


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    expected = build_fixture()
    if args.check:
        if not FIXTURE_PATH.is_file() or json.loads(FIXTURE_PATH.read_text(encoding="utf-8")) != expected:
            print("feature parity fixture drift", file=sys.stderr)
            return 1
        print(f"feature parity fixture is current: {FIXTURE_PATH}")
        return 0
    FIXTURE_PATH.parent.mkdir(parents=True, exist_ok=True)
    FIXTURE_PATH.write_text(json.dumps(expected, indent=2) + "\n", encoding="utf-8")
    print(f"wrote {FIXTURE_PATH}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())