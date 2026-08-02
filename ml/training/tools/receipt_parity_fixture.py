"""Generate and verify canonical Python/Go inference receipt fixtures.

This module reproduces receipt serialization and signing only. It never loads or
executes a model and makes no claim that supplied raw scores are model outputs.
"""

from __future__ import annotations

import argparse
import base64
import copy
import hashlib
import json
import math
import struct
import sys
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "pkg/inference/conformance/testdata/receipt_parity_v1.json"

RECEIPT_DOMAIN = "VEID_INFERENCE_RECEIPT"
SIGN_DOMAIN = "VEID_INFERENCE_RECEIPT_SIGN_V1"
DIGEST_DOMAIN = b"VEID_INFERENCE_RECEIPT_DIGEST_V1"
CONTEXT_DOMAIN = "VEID_INFERENCE_RECEIPT_CONTEXT_V1"
CONFIG_DOMAIN = "VEID_INFERENCE_DETERMINISM_CONFIG_V1"

_FIELD_ORDER = (
    "chain_id", "account_address", "request_id", "scope_ids", "nonce",
    "input_digest", "feature_digest", "schema_digest", "evidence_lineage_digest",
    "pipeline_version", "model_manifest_digest", "model_digest",
    "runtime_image_digest", "runtime_digest", "config_digest",
    "determinism_profile", "score", "status", "confidence_millionths",
    "reason_codes", "issued_height", "issued_at_unix", "expires_height",
    "expires_at_unix", "signer_key_id", "signer_fingerprint", "signer_sequence",
)
_CONTEXT_EXCLUDED = {"score", "status", "confidence_millionths", "reason_codes"}
_BYTE_FIELDS = {
    "input_digest", "feature_digest", "schema_digest", "evidence_lineage_digest",
    "model_manifest_digest", "model_digest", "runtime_image_digest",
    "runtime_digest", "config_digest",
}

# RFC 8032 Ed25519, expressed here with standard-library integer arithmetic so
# fixture generation does not depend on a second crypto package in training CI.
_Q = 2**255 - 19
_L = 2**252 + 27742317777372353535851937790883648493
_D = (-121665 * pow(121666, _Q - 2, _Q)) % _Q
_I = pow(2, (_Q - 1) // 4, _Q)


def _xrecover(y: int) -> int:
    xx = (y * y - 1) * pow(_D * y * y + 1, _Q - 2, _Q) % _Q
    x = pow(xx, (_Q + 3) // 8, _Q)
    if (x * x - xx) % _Q != 0:
        x = x * _I % _Q
    return _Q - x if x & 1 else x


_BY = 4 * pow(5, _Q - 2, _Q) % _Q
_BASE = (_xrecover(_BY), _BY, 1, (_xrecover(_BY) * _BY) % _Q)


def _point_add(first: tuple[int, ...], second: tuple[int, ...]) -> tuple[int, ...]:
    x1, y1, z1, t1 = first
    x2, y2, z2, t2 = second
    a = (y1 - x1) * (y2 - x2) % _Q
    b = (y1 + x1) * (y2 + x2) % _Q
    c = 2 * _D * t1 * t2 % _Q
    d = 2 * z1 * z2 % _Q
    e, f, g, h = b - a, d - c, d + c, b + a
    return e * f % _Q, g * h % _Q, f * g % _Q, e * h % _Q


def _scalar_mult(point: tuple[int, ...], scalar: int) -> tuple[int, ...]:
    result = (0, 1, 1, 0)
    while scalar:
        if scalar & 1:
            result = _point_add(result, point)
        point = _point_add(point, point)
        scalar >>= 1
    return result


def _encode_point(point: tuple[int, ...]) -> bytes:
    x, y, z, _ = point
    inverse = pow(z, _Q - 2, _Q)
    x, y = x * inverse % _Q, y * inverse % _Q
    return (y | ((x & 1) << 255)).to_bytes(32, "little")


def ed25519_public_key(seed: bytes) -> bytes:
    if len(seed) != 32:
        raise ValueError("Ed25519 seed must be 32 bytes")
    hashed = hashlib.sha512(seed).digest()
    scalar = int.from_bytes(hashed[:32], "little")
    scalar &= (1 << 254) - 8
    scalar |= 1 << 254
    return _encode_point(_scalar_mult(_BASE, scalar))


def ed25519_sign(seed: bytes, message: bytes) -> bytes:
    public_key = ed25519_public_key(seed)
    hashed = hashlib.sha512(seed).digest()
    scalar = int.from_bytes(hashed[:32], "little")
    scalar &= (1 << 254) - 8
    scalar |= 1 << 254
    nonce = int.from_bytes(hashlib.sha512(hashed[32:] + message).digest(), "little") % _L
    encoded_nonce = _encode_point(_scalar_mult(_BASE, nonce))
    challenge = int.from_bytes(
        hashlib.sha512(encoded_nonce + public_key + message).digest(), "little"
    ) % _L
    return encoded_nonce + ((nonce + challenge * scalar) % _L).to_bytes(32, "little")


def _compact_json(value: Any) -> bytes:
    return json.dumps(value, ensure_ascii=True, separators=(",", ":")).encode("utf-8")


def _sha256_source(label: str) -> bytes:
    if not isinstance(label, str) or not label:
        raise ValueError("digest source labels must be non-empty strings")
    return hashlib.sha256(label.encode("utf-8")).digest()


def _canonical_strings(values: list[str], field: str) -> list[str]:
    if not isinstance(values, list) or any(not isinstance(value, str) for value in values):
        raise ValueError(f"{field} must be a list of strings")
    canonical = sorted({value for value in values if value})
    if not canonical:
        raise ValueError(f"{field} must contain at least one non-empty value")
    return canonical


def _float32(value: Any) -> float:
    special = {
        "nan": float("nan"),
        "positive_infinity": float("inf"),
        "negative_infinity": float("-inf"),
    }
    if isinstance(value, str) and value in special:
        return special[value]
    try:
        converted = float(value)
    except (TypeError, ValueError) as exc:
        raise ValueError("raw_output_float32 must be numeric") from exc
    return struct.unpack(">f", struct.pack(">f", converted))[0]


def quantize_score(raw_output: Any) -> int:
    """Match security.SafeFloat32ToUint32(raw, 0, 100)."""
    value = _float32(raw_output)
    if math.isnan(value) or value < 0:
        return 0
    if value > 100:
        return 100
    return int(value)


def _config_digest() -> bytes:
    envelope = {
        "domain": CONFIG_DOMAIN,
        "random_seed": 42,
        "force_cpu": True,
        "single_thread": True,
        "float_precision": 6,
        "tensorflow_deterministic": True,
        "disable_cudnn": True,
        "onnx_deterministic": True,
    }
    return hashlib.sha256(_compact_json(envelope)).digest()


def _source_cases() -> list[dict[str, Any]]:
    common = {
        "chain_id": "virtengine-parity-1",
        "account_address": "virt1receiptparity",
        "pipeline_version": "receipt-fixture-v1",
        "digest_sources": {
            "input_digest": "receipt-v1/input",
            "feature_digest": "receipt-v1/canonical-feature-bytes-not-vector-hash",
            "schema_digest": "receipt-v1/schema",
            "evidence_lineage_digest": "receipt-v1/evidence-lineage",
            "model_manifest_digest": "receipt-v1/model-manifest",
            "model_digest": "receipt-v1/model",
            "runtime_image_digest": "receipt-v1/runtime-image",
            "runtime_digest": "receipt-v1/runtime-config",
        },
        "issued_height": 1_234_500,
        "issued_at_unix": 1_700_000_000,
        "expires_height": 1_234_502,
        "expires_at_unix": 1_700_000_120,
        "signer_key_id": "did:virtengine:receipt-parity:v1",
    }
    cases = [
        {
            "name": "success",
            "source": {"request_id": "request-success", "scope_ids": ["scope-b", "scope-a"],
                       "nonce": "nonce-success", "score": 91, "status": "success",
                       "confidence_millionths": 910_000, "reason_codes": ["SUCCESS"],
                       "signer_sequence": 1},
        },
        {
            "name": "partial_canonicalizes_duplicates",
            "source": {"request_id": "request-partial", "scope_ids": ["scope-c", "scope-a", "scope-c", ""],
                       "nonce": "nonce-partial", "score": 49, "status": "partial",
                       "confidence_millionths": 499_999,
                       "reason_codes": ["LOW_CONFIDENCE", "FACE_MISMATCH", "LOW_CONFIDENCE", ""],
                       "signer_sequence": 2},
        },
        {
            "name": "failed",
            "source": {"request_id": "request-failed", "scope_ids": ["scope-failed"],
                       "nonce": "nonce-failed", "score": 0, "status": "failed",
                       "confidence_millionths": 125_000,
                       "reason_codes": ["ML_INFERENCE_ERROR"], "signer_sequence": 3},
        },
        {
            "name": "fractional_raw_score_boundary",
            "source": {"request_id": "request-fractional", "scope_ids": ["scope-score"],
                       "nonce": "nonce-fractional", "raw_output_float32": 49.999996185302734,
                       "score": 49, "status": "partial", "confidence_millionths": 500_001,
                       "reason_codes": ["LOW_CONFIDENCE"], "signer_sequence": 4},
        },
        {
            "name": "nan_raw_score_clamps_min",
            "source": {"request_id": "request-nan", "scope_ids": ["scope-nan"],
                       "nonce": "nonce-nan", "raw_output_float32": "nan",
                       "score": 0, "status": "failed", "confidence_millionths": 0,
                       "reason_codes": ["ML_INFERENCE_ERROR"], "signer_sequence": 5},
        },
        {
            "name": "positive_infinity_clamps_max",
            "source": {"request_id": "request-positive-infinity", "scope_ids": ["scope-positive-infinity"],
                       "nonce": "nonce-positive-infinity", "raw_output_float32": "positive_infinity",
                       "score": 100, "status": "success", "confidence_millionths": 1_000_000,
                       "reason_codes": ["SUCCESS"], "signer_sequence": 6},
        },
        {
            "name": "negative_infinity_clamps_min",
            "source": {"request_id": "request-negative-infinity", "scope_ids": ["scope-negative-infinity"],
                       "nonce": "nonce-negative-infinity", "raw_output_float32": "negative_infinity",
                       "score": 0, "status": "failed", "confidence_millionths": 0,
                       "reason_codes": ["ML_INFERENCE_ERROR"], "signer_sequence": 7},
        },
    ]
    for case in cases:
        case["source"] = copy.deepcopy(common) | case["source"]
    return cases


def _receipt_from_source(source: dict[str, Any], fingerprint: str) -> dict[str, Any]:
    if "raw_output_float32" in source and quantize_score(source["raw_output_float32"]) != source["score"]:
        raise ValueError("explicit score does not match production float32 truncation contract")
    receipt: dict[str, Any] = {
        "chain_id": source["chain_id"],
        "account_address": source["account_address"],
        "request_id": source["request_id"],
        "scope_ids": _canonical_strings(source["scope_ids"], "scope_ids"),
        "nonce": source["nonce"],
    }
    digest_sources = source["digest_sources"]
    for field in _BYTE_FIELDS - {"config_digest"}:
        receipt[field] = _sha256_source(digest_sources[field])
    receipt.update({
        "pipeline_version": source["pipeline_version"],
        "config_digest": _config_digest(),
        "determinism_profile": {"force_cpu": True, "random_seed": 42,
                                "deterministic_ops": True, "inter_op_threads": 1,
                                "intra_op_threads": 1, "disable_gpu": True},
        "score": source["score"], "status": source["status"],
        "confidence_millionths": source["confidence_millionths"],
        "reason_codes": _canonical_strings(source["reason_codes"], "reason_codes"),
        "issued_height": source["issued_height"], "issued_at_unix": source["issued_at_unix"],
        "expires_height": source["expires_height"], "expires_at_unix": source["expires_at_unix"],
        "signer_key_id": source["signer_key_id"], "signer_fingerprint": fingerprint.lower(),
        "signer_sequence": source["signer_sequence"],
    })
    return receipt


def _envelope(receipt: dict[str, Any], context: bool = False) -> dict[str, Any]:
    domain = CONTEXT_DOMAIN if context else SIGN_DOMAIN
    envelope: dict[str, Any] = {"domain": domain, "version": 1}
    for field in _FIELD_ORDER:
        if context and field in _CONTEXT_EXCLUDED:
            continue
        value = receipt[field]
        envelope[field] = base64.b64encode(value).decode("ascii") if field in _BYTE_FIELDS else value
    return envelope


def build_fixture(cases: list[dict[str, Any]] | None = None) -> dict[str, Any]:
    source_cases = copy.deepcopy(_source_cases() if cases is None else cases)
    if not source_cases:
        raise ValueError("receipt parity fixture must contain at least one case")
    seed = hashlib.sha256(b"virtengine/inference/receipt/parity-fixture/v1").digest()
    public_key = ed25519_public_key(seed)
    fingerprint = hashlib.sha256(public_key).hexdigest()
    fixture: dict[str, Any] = {
        "schema": "virtengine.inference.receipt_parity", "version": 1,
        "nonclaims": [
            "No production model or runtime is loaded or executed.",
            "Supplied raw_output_float32 values test score truncation only and are not model outputs.",
            "FeatureDigest is a receipt input digest and is not the canonical feature vector hash.",
            "ConfidenceMillionths values are explicit receipt fields; no confidence postprocessing parity is claimed.",
        ],
        "signing": {"seed_hex": seed.hex(), "public_key_base64": base64.b64encode(public_key).decode("ascii"),
                    "fingerprint_hex": fingerprint},
        "cases": [],
    }
    for case in source_cases:
        receipt = _receipt_from_source(case["source"], fingerprint)
        sign_bytes = _compact_json(_envelope(receipt))
        context_bytes = _compact_json(_envelope(receipt, context=True))
        expected = {
            "canonical_scope_ids": receipt["scope_ids"],
            "canonical_reason_codes": receipt["reason_codes"],
            "config_digest_hex": receipt["config_digest"].hex(),
            "sign_bytes_base64": base64.b64encode(sign_bytes).decode("ascii"),
            "digest_hex": hashlib.sha256(DIGEST_DOMAIN + sign_bytes).hexdigest(),
            "context_digest_hex": hashlib.sha256(CONTEXT_DOMAIN.encode("ascii") + context_bytes).hexdigest(),
            "signature_base64": base64.b64encode(ed25519_sign(seed, sign_bytes)).decode("ascii"),
        }
        if "raw_output_float32" in case["source"]:
            expected["quantized_score"] = quantize_score(case["source"]["raw_output_float32"])
        fixture["cases"].append({"name": case["name"], "source": case["source"], "expected": expected})
    return fixture


def verify_fixture(actual: Any) -> None:
    if not isinstance(actual, dict):
        raise ValueError("fixture must be a JSON object")
    if actual.get("schema") != "virtengine.inference.receipt_parity" or actual.get("version") != 1:
        raise ValueError("unsupported receipt parity fixture schema or version")
    cases = actual.get("cases")
    if not isinstance(cases, list) or not cases:
        raise ValueError("receipt parity fixture must contain at least one case")
    expected = build_fixture([{"name": case.get("name"), "source": case.get("source")} for case in cases])
    if actual != expected:
        raise ValueError("receipt parity fixture drift")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--check", action="store_true", help="fail if the checked-in fixture drifts")
    args = parser.parse_args()
    if args.check:
        if not FIXTURE_PATH.is_file():
            print(f"missing fixture: {FIXTURE_PATH}", file=sys.stderr)
            return 1
        try:
            verify_fixture(json.loads(FIXTURE_PATH.read_text(encoding="utf-8")))
        except (json.JSONDecodeError, KeyError, TypeError, ValueError) as exc:
            print(f"receipt parity fixture check failed: {exc}", file=sys.stderr)
            return 1
        print(f"receipt parity fixture is current: {FIXTURE_PATH}")
        return 0
    fixture = build_fixture()
    FIXTURE_PATH.parent.mkdir(parents=True, exist_ok=True)
    FIXTURE_PATH.write_text(json.dumps(fixture, indent=2) + "\n", encoding="utf-8")
    print(f"wrote {FIXTURE_PATH}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())