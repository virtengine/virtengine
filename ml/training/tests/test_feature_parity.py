"""Parity tests for the shared canonical feature fixture."""

import json
from pathlib import Path

import numpy as np
import pytest

from ml.training.features.canonical_vector import (
    CanonicalInferenceInputs,
    SCHEMA_BYTES,
    SCHEMA_PATH,
    SCHEMA_SHA256,
    TOTAL_FEATURE_DIM,
    assemble_canonical_inference_vector,
)
from ml.training.tools.feature_parity_fixture import feature_hash, inputs_from_source, raw_feature_hash


FIXTURE_PATH = Path(__file__).resolve().parents[3] / "pkg/inference/conformance/testdata/feature_parity_v1.json"


def test_feature_parity_fixture():
    fixture = json.loads(FIXTURE_PATH.read_text(encoding="utf-8"))
    assert fixture["contract"] == "fixture_only"
    assert fixture["feature_schema_sha256"] == SCHEMA_SHA256
    assert fixture["layout"]["total_dimension"] == TOTAL_FEATURE_DIM
    for case in fixture["cases"]:
        vector = assemble_canonical_inference_vector(
            inputs_from_source(case["source"]),
            strict_production=case["profile"] == "production",
        )
        assert feature_hash(vector) == case["expected_vector_hash"]
        assert raw_feature_hash(vector) == case["expected_raw_vector_hash"]
        actual_nonzero = {str(index): float(value) for index, value in enumerate(vector) if value != 0.0}
        assert actual_nonzero == pytest.approx(case["expected_nonzero_positions"])


def test_schema_bytes_are_loaded_from_go_embedded_artifact():
    assert SCHEMA_BYTES == SCHEMA_PATH.read_bytes()


@pytest.mark.parametrize(
    "change, match",
    [
        (lambda values: values.__setitem__("face_embedding", None), "face_embedding"),
        (lambda values: values["ocr_confidences"].pop("name"), "OCR confidence"),
        (lambda values: values.__setitem__("face_confidence", np.nan), "finite"),
        (lambda values: values.__setitem__("blur_score", 1.1), "\\[0,1\\]"),
        (lambda values: values.__setitem__("scope_count", 11), "scope_count"),
        (lambda values: values.__setitem__("block_height", -1), "block_height"),
        (lambda values: values.__setitem__("scope_types", None), "scope_types"),
    ],
)
def test_strict_production_rejects_invalid_required_inputs(change, match):
    values = {
        "face_embedding": np.ones(512, dtype=np.float32),
        "face_confidence": 0.5,
        "doc_quality_score": 0.5,
        "sharpness": 0.5,
        "brightness": 0.5,
        "contrast": 0.5,
        "noise_level": 0.5,
        "blur_score": 0.5,
        "ocr_confidences": {name: 0.5 for name in ("name", "date_of_birth", "document_number", "expiry_date", "nationality")},
        "ocr_field_validation": {name: True for name in ("name", "date_of_birth", "document_number", "expiry_date", "nationality")},
        "scope_types": [],
        "scope_count": 1,
        "block_height": 1,
    }
    change(values)
    with pytest.raises(ValueError, match=match):
        assemble_canonical_inference_vector(CanonicalInferenceInputs(**values), strict_production=True)