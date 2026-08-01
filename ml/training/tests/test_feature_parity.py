"""Parity tests for the shared canonical feature fixture."""

import json
from pathlib import Path

import pytest

from ml.training.features.canonical_vector import TOTAL_FEATURE_DIM, assemble_canonical_inference_vector
from ml.training.tools.feature_parity_fixture import feature_hash, inputs_from_source


FIXTURE_PATH = (
    Path(__file__).resolve().parents[3]
    / "pkg/inference/conformance/testdata/feature_parity_v1.json"
)


def test_feature_parity_fixture():
    fixture = json.loads(FIXTURE_PATH.read_text(encoding="utf-8"))
    assert fixture["schema"] == "virtengine.inference.feature_parity"
    assert fixture["version"] == 1
    assert fixture["layout"]["total_dimension"] == TOTAL_FEATURE_DIM
    assert fixture["cases"]

    for case in fixture["cases"]:
        vector = assemble_canonical_inference_vector(inputs_from_source(case["source"]))
        assert vector.shape == (TOTAL_FEATURE_DIM,)
        assert feature_hash(vector) == case["expected_vector_hash"]
        expected_nonzero = {
            int(index): value
            for index, value in case["expected_nonzero_positions"].items()
        }
        actual_nonzero = {
            index: float(value)
            for index, value in enumerate(vector)
            if value != 0.0
        }
        assert actual_nonzero == pytest.approx(expected_nonzero)