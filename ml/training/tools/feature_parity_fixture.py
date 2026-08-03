"""Generate and verify the canonical Python/Go feature parity fixture."""

from __future__ import annotations

import argparse
import hashlib
import json
import math
import struct
import sys
from pathlib import Path
from typing import Any, Dict

import numpy as np

from ml.training.features.canonical_vector import (
    CanonicalInferenceInputs,
    DOC_QUALITY_DIM,
    DOC_QUALITY_OFFSET,
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
    TOTAL_FEATURE_DIM,
    assemble_canonical_inference_vector,
)


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "pkg/inference/conformance/testdata/feature_parity_v1.json"
)


def expand_face_embedding(spec: Dict[str, Any]) -> np.ndarray:
    """Expand a deterministic fixture source into a float32 embedding."""
    kind = spec["kind"]
    if kind == "missing":
        return np.array([], dtype=np.float32)
    dimension = int(spec["dimension"])
    if kind == "zeros":
        return np.zeros(dimension, dtype=np.float32)
    if kind == "single_index":
        result = np.zeros(dimension, dtype=np.float32)
        result[int(spec["index"])] = np.float32(spec["value"])
        return result
    if kind == "alternating_bounds":
        low = np.float32(spec["low"])
        high = np.float32(spec["high"])
        return np.array(
            [low if index % 2 == 0 else high for index in range(dimension)],
            dtype=np.float32,
        )
    if kind == "asymmetric_sequence":
        start = np.float32(spec["start"])
        step = np.float32(spec["step"])
        return np.array(
            [np.float32(start + np.float32(index) * step) for index in range(dimension)],
            dtype=np.float32,
        )
    raise ValueError(f"unsupported face embedding source kind: {kind}")


def inputs_from_source(source: Dict[str, Any]) -> CanonicalInferenceInputs:
    """Convert JSON source inputs without consuming expected vector data."""
    document = source.get("document_quality", {})
    return CanonicalInferenceInputs(
        face_embedding=expand_face_embedding(source["face_embedding"]),
        face_confidence=source.get("face_confidence", 0.0),
        doc_quality_score=source.get("doc_quality_score", 0.0),
        sharpness=document.get("sharpness", 0.0),
        brightness=document.get("brightness", 0.0),
        contrast=document.get("contrast", 0.0),
        noise_level=document.get("noise_level", 0.0),
        ocr_confidences=source.get("ocr_confidences", {}),
        ocr_field_validation=source.get("ocr_field_validation", {}),
        scope_types=source.get("scope_types", []),
        scope_count=source.get("scope_count", 0),
        block_height=source.get("block_height", 0),
    )


def _round_go(value: float, places: int) -> float:
    multiplier = 10 ** places
    scaled = value * multiplier
    rounded = math.floor(scaled + 0.5) if scaled >= 0 else math.ceil(scaled - 0.5)
    return rounded / multiplier


def feature_hash(vector: np.ndarray, places: int = 6) -> str:
    """Match DeterminismController.ComputeFeatureHash byte-for-byte."""
    digest = hashlib.sha256()
    for value in vector:
        rounded = np.float32(_round_go(float(np.float32(value)), places))
        digest.update(struct.pack(">f", rounded))
    return digest.hexdigest()


def _cases() -> list[Dict[str, Any]]:
    return [
        {
            "name": "zero_missing",
            "source": {
                "face_embedding": {"kind": "missing"},
            },
        },
        {
            "name": "unique_position_sentinel",
            "source": {
                "face_embedding": {
                    "kind": "single_index",
                    "dimension": FACE_EMBEDDING_DIM,
                    "index": 511,
                    "value": -2.0,
                },
                "face_confidence": 0.91,
                "doc_quality_score": 0.11,
                "document_quality": {
                    "sharpness": 0.22,
                    "brightness": 0.33,
                    "contrast": 0.44,
                    "noise_level": 0.45,
                },
                "ocr_confidences": {
                    "name": 0.61,
                    "date_of_birth": 0.62,
                    "document_number": 0.63,
                    "expiry_date": 0.64,
                    "nationality": 0.65,
                },
                "ocr_field_validation": {
                    "name": True,
                    "date_of_birth": False,
                    "document_number": True,
                    "expiry_date": False,
                    "nationality": True,
                },
                "scope_types": [
                    "id_document",
                    "face_video",
                    "sso_metadata",
                    "sms_proof",
                ],
                "scope_count": 7,
                "block_height": 1234567,
            },
        },
        {
            "name": "boundary_values",
            "source": {
                "face_embedding": {
                    "kind": "alternating_bounds",
                    "dimension": FACE_EMBEDDING_DIM,
                    "low": -1.0,
                    "high": 1.0,
                },
                "face_confidence": 1.0,
                "doc_quality_score": 1.0,
                "document_quality": {
                    "sharpness": 0.0,
                    "brightness": 1.0,
                    "contrast": 0.0,
                    "noise_level": 1.0,
                },
                "ocr_confidences": {
                    "name": 0.0,
                    "date_of_birth": 1.0,
                    "document_number": 0.0,
                    "expiry_date": 1.0,
                    "nationality": 0.0,
                },
                "ocr_field_validation": {
                    "name": False,
                    "date_of_birth": True,
                    "document_number": False,
                    "expiry_date": True,
                    "nationality": False,
                },
                "scope_types": list(METADATA_SCOPE_TYPES),
                "scope_count": 10,
                "block_height": 999999,
            },
        },
        {
            "name": "asymmetric_negative_height",
            "source": {
                "face_embedding": {
                    "kind": "asymmetric_sequence",
                    "dimension": FACE_EMBEDDING_DIM,
                    "start": -0.73,
                    "step": 0.0031,
                },
                "face_confidence": 0.37,
                "scope_count": 3,
                "block_height": -1234567,
            },
        },
    ]


def build_fixture() -> Dict[str, Any]:
    """Build the complete versioned fixture deterministically."""
    fixture: Dict[str, Any] = {
        "schema": "virtengine.inference.feature_parity",
        "version": 1,
        "layout": {
            "total_dimension": TOTAL_FEATURE_DIM,
            "components": [
                {
                    "name": "face_embedding",
                    "offset": FACE_OFFSET,
                    "dimension": FACE_EMBEDDING_DIM,
                    "position_name": "face_embedding[{index}]",
                },
                {
                    "name": "document_quality",
                    "offset": DOC_QUALITY_OFFSET,
                    "dimension": DOC_QUALITY_DIM,
                    "position_names": [
                        "doc_quality_score",
                        "sharpness",
                        "brightness",
                        "contrast",
                        "one_minus_noise_level",
                    ],
                },
                {
                    "name": "ocr",
                    "offset": OCR_OFFSET,
                    "dimension": OCR_FIELDS_DIM,
                    "fields": list(OCR_FIELD_NAMES),
                    "position_names_per_field": ["confidence", "validated"],
                },
                {
                    "name": "metadata",
                    "offset": METADATA_OFFSET,
                    "dimension": METADATA_DIM,
                    "position_names": [
                        "scope_count_div_10_capped",
                        *[f"scope_type[{name}]" for name in METADATA_SCOPE_TYPES],
                        "face_confidence",
                        "block_height_mod_1000000_div_1000000",
                        "reserved[0]",
                        "reserved[1]",
                        "reserved[2]",
                        "reserved[3]",
                        "reserved[4]",
                    ],
                },
                {
                    "name": "padding",
                    "offset": PADDING_OFFSET,
                    "dimension": PADDING_DIM,
                    "position_name": "reserved_padding[{index}]",
                },
            ],
            "encoding": {
                "value_type": "ieee754-float32",
                "byte_order": "big-endian",
                "pre_hash_decimal_places": 6,
                "hash_algorithm": "sha256",
            },
        },
        "cases": _cases(),
    }
    for case in fixture["cases"]:
        vector = assemble_canonical_inference_vector(inputs_from_source(case["source"]))
        case["expected_vector_hash"] = feature_hash(vector)
        case["expected_nonzero_positions"] = {
            str(index): float(value)
            for index, value in enumerate(vector)
            if value != 0.0
        }
    return fixture


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--check", action="store_true", help="fail if the checked-in fixture drifts"
    )
    args = parser.parse_args()
    expected = build_fixture()

    if args.check:
        if not FIXTURE_PATH.is_file():
            print(f"missing fixture: {FIXTURE_PATH}", file=sys.stderr)
            return 1
        actual = json.loads(FIXTURE_PATH.read_text(encoding="utf-8"))
        if actual != expected:
            print(
                "feature parity fixture drift; regenerate with "
                "python -m ml.training.tools.feature_parity_fixture",
                file=sys.stderr,
            )
            return 1
        print(f"feature parity fixture is current: {FIXTURE_PATH}")
        return 0

    FIXTURE_PATH.parent.mkdir(parents=True, exist_ok=True)
    FIXTURE_PATH.write_text(
        json.dumps(expected, indent=2, sort_keys=False) + "\n", encoding="utf-8"
    )
    print(f"wrote {FIXTURE_PATH}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())