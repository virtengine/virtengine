"""Tests for the source-driven inference receipt parity fixture."""

import copy
import json

import pytest

from ml.training.tools.receipt_parity_fixture import (
    FIXTURE_PATH,
    build_fixture,
    ed25519_public_key,
    quantize_score,
    verify_fixture,
)


def test_receipt_parity_fixture_is_current():
    fixture = json.loads(FIXTURE_PATH.read_text(encoding="utf-8"))
    verify_fixture(fixture)
    assert fixture["cases"]
    assert {case["name"] for case in fixture["cases"]} == {
        "success",
        "partial_canonicalizes_duplicates",
        "failed",
        "fractional_raw_score_boundary",
        "nan_raw_score_clamps_min",
        "positive_infinity_clamps_max",
        "negative_infinity_clamps_min",
    }


def test_receipt_parity_rejects_zero_cases():
    with pytest.raises(ValueError, match="at least one case"):
        build_fixture([])

    fixture = build_fixture()
    fixture["cases"] = []
    with pytest.raises(ValueError, match="at least one case"):
        verify_fixture(fixture)


@pytest.mark.parametrize(
    ("value", "expected"),
    [("nan", 0), ("positive_infinity", 100), ("negative_infinity", 0)],
)
def test_receipt_parity_matches_nonfinite_production_clamping(value, expected):
    assert quantize_score(value) == expected


def test_receipt_parity_rejects_invalid_seed_and_inputs():
    with pytest.raises(ValueError, match="32 bytes"):
        ed25519_public_key(b"short")

    cases = build_fixture()["cases"]
    source_cases = [{"name": case["name"], "source": case["source"]} for case in cases]
    source_cases[0]["source"]["scope_ids"] = [""]
    with pytest.raises(ValueError, match="scope_ids"):
        build_fixture(source_cases)


@pytest.mark.parametrize(
    ("path", "replacement"),
    [
        (("signing", "fingerprint_hex"), "00" * 32),
        (("cases", 0, "source", "nonce"), "mutated"),
        (("cases", 0, "expected", "digest_hex"), "00" * 32),
        (("cases", 1, "expected", "canonical_scope_ids"), ["scope-z"]),
    ],
)
def test_receipt_parity_detects_mutations(path, replacement):
    fixture = build_fixture()
    mutated = copy.deepcopy(fixture)
    target = mutated
    for component in path[:-1]:
        target = target[component]
    target[path[-1]] = replacement
    with pytest.raises(ValueError, match="drift"):
        verify_fixture(mutated)


def test_fractional_score_uses_float32_truncation_only():
    fixture = build_fixture()
    case = next(case for case in fixture["cases"] if case["name"] == "fractional_raw_score_boundary")
    assert quantize_score(case["source"]["raw_output_float32"]) == 49
    assert case["expected"]["quantized_score"] == case["source"]["score"]
    assert any("not model outputs" in claim for claim in fixture["nonclaims"])
