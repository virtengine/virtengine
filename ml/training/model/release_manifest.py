"""
Deterministic release manifest generation for VEID model bundles.

This module validates a versioned model bundle, captures source provenance,
and emits a deterministic manifest that the build pipeline and inference
sidecar can both verify.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import re
from pathlib import Path
from typing import Any, Dict, Iterable, List, Optional


PLACEHOLDER_PATTERN = re.compile(
    r"(?i)(placeholder|pending|tbd|not published yet|<path>|sha256:placeholder)"
)

REQUIRED_BUNDLE_FILES = (
    "manifest.json",
    "export_metadata.json",
    "MODEL_HASH.txt",
    "model_frozen.pb",
)


class ReleaseManifestError(ValueError):
    """Raised when a model bundle cannot be released safely."""


def _read_text(path: Path) -> str:
    return path.read_text(encoding="utf-8")


def _load_json(path: Path) -> Dict[str, Any]:
    try:
        return json.loads(_read_text(path))
    except json.JSONDecodeError as exc:
        raise ReleaseManifestError(f"invalid JSON in {path}: {exc}") from exc


def _sha256_file(path: Path) -> str:
    hasher = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            hasher.update(chunk)
    return hasher.hexdigest()


def _sha256_model_dir(model_dir: Path) -> str:
    hasher = hashlib.sha256()
    files = sorted(path for path in model_dir.rglob("*") if path.is_file())
    for path in files:
        if path.name == "export_metadata.json":
            continue
        with path.open("rb") as handle:
            for chunk in iter(lambda: handle.read(1024 * 1024), b""):
                hasher.update(chunk)
    return hasher.hexdigest()


def _ensure_no_placeholders(path: Path) -> None:
    if PLACEHOLDER_PATTERN.search(_read_text(path)):
        raise ReleaseManifestError(f"placeholder content detected in {path}")


def _require_value(value: Any, field_name: str) -> str:
    normalized = str(value or "").strip()
    if normalized == "":
        raise ReleaseManifestError(f"{field_name} is required")
    if PLACEHOLDER_PATTERN.search(normalized):
        raise ReleaseManifestError(f"placeholder content detected in {field_name}")
    return normalized


def _normalize_hash(value: Any, field_name: str) -> str:
    normalized = str(value or "").strip().lower().removeprefix("sha256:")
    if not re.fullmatch(r"[0-9a-f]{64}", normalized):
        raise ReleaseManifestError(f"{field_name} must be a 64 character SHA-256 hex digest")
    return normalized


def _parse_model_hash_file(path: Path) -> Dict[str, str]:
    _ensure_no_placeholders(path)
    entries: Dict[str, str] = {}
    for line in _read_text(path).splitlines():
        if "=" not in line or line.lstrip().startswith("#"):
            continue
        key, value = line.split("=", 1)
        entries[key.strip()] = value.strip()

    if "SHA256" not in entries:
        raise ReleaseManifestError(f"MODEL_HASH.txt missing SHA256 entry: {path}")

    if "VERSION" not in entries:
        raise ReleaseManifestError(f"MODEL_HASH.txt missing VERSION entry: {path}")

    entries["SHA256"] = _normalize_hash(entries["SHA256"], f"{path} SHA256")
    return entries


def _artifact_record(path: Path, bundle_dir: Path) -> Dict[str, Any]:
    return {
        "path": path.relative_to(bundle_dir).as_posix(),
        "sha256": _sha256_file(path),
        "size_bytes": path.stat().st_size,
    }


def _canonical_json_sha256(payload: Any) -> str:
    encoded = json.dumps(payload, sort_keys=True, separators=(",", ":")).encode("utf-8")
    return hashlib.sha256(encoded).hexdigest()


def _display_path(path: Path, display_path: Optional[str], field_name: str) -> str:
    if display_path is None:
        return path.as_posix()
    return _require_value(display_path, field_name)


def _required_artifacts(bundle_dir: Path) -> Iterable[Path]:
    model_dir = bundle_dir / "model"
    if not model_dir.is_dir():
        raise ReleaseManifestError(f"required model directory is missing: {model_dir}")

    for relative in REQUIRED_BUNDLE_FILES:
        path = bundle_dir / relative
        if not path.exists():
            raise ReleaseManifestError(f"required bundle artifact is missing: {path}")

    yield from (bundle_dir / relative for relative in REQUIRED_BUNDLE_FILES)

    for path in sorted(candidate for candidate in model_dir.rglob("*") if candidate.is_file()):
        yield path


def build_release_manifest(
    bundle_dir: Path,
    *,
    model_name: str,
    version: str,
    profile: str,
    config_path: Path,
    model_card_path: Path,
    source_revision: str,
    source_describe: str = "",
    bundle_display_path: Optional[str] = None,
    config_display_path: Optional[str] = None,
    model_card_display_path: Optional[str] = None,
) -> Dict[str, Any]:
    bundle_dir = bundle_dir.resolve()
    model_dir = bundle_dir / "model"

    model_name = _require_value(model_name, "model_name")
    version = _require_value(version, "version")
    profile = _require_value(profile, "profile")
    source_revision = _require_value(source_revision, "source_revision")
    source_describe = str(source_describe or "").strip()
    if source_describe:
        source_describe = _require_value(source_describe, "source_describe")

    if not bundle_dir.is_dir():
        raise ReleaseManifestError(f"bundle directory does not exist: {bundle_dir}")
    if not config_path.is_file():
        raise ReleaseManifestError(f"training config not found: {config_path}")
    if not model_card_path.is_file():
        raise ReleaseManifestError(f"model card not found: {model_card_path}")

    for path in (config_path, model_card_path):
        _ensure_no_placeholders(path)

    artifacts = list(_required_artifacts(bundle_dir))
    for path in artifacts:
        if path.suffix in {".json", ".txt", ".md"}:
            _ensure_no_placeholders(path)

    training_manifest = _load_json(bundle_dir / "manifest.json")
    export_metadata = _load_json(bundle_dir / "export_metadata.json")
    model_hash_file = _parse_model_hash_file(bundle_dir / "MODEL_HASH.txt")

    runtime_model_hash = _sha256_model_dir(model_dir)
    frozen_graph_hash = _sha256_file(bundle_dir / "model_frozen.pb")
    training_manifest_model_hash = _normalize_hash(
        training_manifest.get("model_hash"),
        "manifest.json model_hash",
    )
    export_metadata_model_hash = _normalize_hash(
        export_metadata.get("model_hash"),
        "export_metadata.json model_hash",
    )
    model_hash_file_value = model_hash_file["SHA256"]
    allowed_model_hashes = {runtime_model_hash, frozen_graph_hash}

    for field_name, value in (
        ("manifest.json model_hash", training_manifest_model_hash),
        ("export_metadata.json model_hash", export_metadata_model_hash),
        ("MODEL_HASH.txt SHA256", model_hash_file_value),
    ):
        if value not in allowed_model_hashes:
            raise ReleaseManifestError(
                f"{field_name} does not match the runtime model hash or frozen graph hash"
            )

    manifest_version = str(training_manifest.get("model_version") or "").strip()
    metadata_version = str(export_metadata.get("version") or "").strip()
    hash_file_version = model_hash_file["VERSION"]

    expected_versions = {version}
    observed_versions = {value for value in (manifest_version, metadata_version, hash_file_version) if value}
    if expected_versions != observed_versions:
        raise ReleaseManifestError(
            f"bundle version mismatch: expected {version}, found {sorted(observed_versions)}"
        )

    input_signature = training_manifest.get("input_signature") or export_metadata.get("input_signature") or {}
    output_signature = training_manifest.get("output_signature") or export_metadata.get("output_signature") or {}
    if not input_signature or not output_signature:
        raise ReleaseManifestError("bundle is missing input/output signature metadata")

    artifact_records = sorted(
        (_artifact_record(path, bundle_dir) for path in artifacts),
        key=lambda item: item["path"],
    )
    artifact_index_sha256 = _canonical_json_sha256(artifact_records)

    return {
        "schema_version": "veid.release.manifest/v1",
        "profile": profile,
        "model": {
            "name": model_name,
            "version": version,
            "model_dir": "model",
            "runtime_hash": runtime_model_hash,
            "runtime_hash_algorithm": "sha256-dir-v1",
            "frozen_graph_hash": frozen_graph_hash,
            "frozen_graph_hash_algorithm": "sha256-file",
            "signature_name": str(export_metadata.get("signature_name") or "serving_default"),
            "input_signature": input_signature,
            "output_signature": output_signature,
            "tensorflow_version": str(
                training_manifest.get("tensorflow_version")
                or export_metadata.get("tensorflow_version")
                or ""
            ),
        },
        "source": {
            "revision": source_revision,
            "revision_short": source_revision[:12],
            "describe": source_describe,
            "generator": "ml.training.model.release_manifest",
            "config_path": _display_path(config_path, config_display_path, "config_display_path"),
            "config_sha256": _sha256_file(config_path),
            "model_card_path": _display_path(
                model_card_path,
                model_card_display_path,
                "model_card_display_path",
            ),
            "model_card_sha256": _sha256_file(model_card_path),
            "bundle_path": _display_path(bundle_dir, bundle_display_path, "bundle_display_path"),
            "bundle_artifact_index_sha256": artifact_index_sha256,
        },
        "consistency": {
            "training_manifest_model_hash": training_manifest_model_hash,
            "export_metadata_model_hash": export_metadata_model_hash,
            "model_hash_file_value": model_hash_file_value,
            "accepted_hashes": sorted(allowed_model_hashes),
        },
        "artifacts": artifact_records,
    }


def write_release_manifest(
    output_path: Path,
    *,
    bundle_dir: Path,
    model_name: str,
    version: str,
    profile: str,
    config_path: Path,
    model_card_path: Path,
    source_revision: str,
    source_describe: str = "",
    bundle_display_path: Optional[str] = None,
    config_display_path: Optional[str] = None,
    model_card_display_path: Optional[str] = None,
) -> Dict[str, Any]:
    manifest = build_release_manifest(
        bundle_dir,
        model_name=model_name,
        version=version,
        profile=profile,
        config_path=config_path,
        model_card_path=model_card_path,
        source_revision=source_revision,
        source_describe=source_describe,
        bundle_display_path=bundle_display_path,
        config_display_path=config_display_path,
        model_card_display_path=model_card_display_path,
    )
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(
        json.dumps(manifest, indent=2, sort_keys=True) + "\n",
        encoding="utf-8",
    )
    return manifest


def _parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Generate a deterministic VEID release manifest")
    parser.add_argument("--bundle-dir", required=True, help="Versioned model bundle directory")
    parser.add_argument("--output", required=True, help="Output path for release_manifest.json")
    parser.add_argument("--model-name", default="trust_score", help="Model name")
    parser.add_argument("--version", required=True, help="Expected model version")
    parser.add_argument("--profile", default="production", help="Build profile")
    parser.add_argument("--config-path", required=True, help="Training config path")
    parser.add_argument("--model-card-path", required=True, help="Model card path")
    parser.add_argument("--source-revision", required=True, help="Source revision for provenance capture")
    parser.add_argument("--source-describe", default="", help="Optional source describe/tag value")
    parser.add_argument("--bundle-display-path", default=None, help="Portable path label for the bundle")
    parser.add_argument("--config-display-path", default=None, help="Portable path label for the training config")
    parser.add_argument("--model-card-display-path", default=None, help="Portable path label for the model card")
    return parser.parse_args()


def main() -> int:
    args = _parse_args()
    write_release_manifest(
        Path(args.output),
        bundle_dir=Path(args.bundle_dir),
        model_name=args.model_name,
        version=args.version,
        profile=args.profile,
        config_path=Path(args.config_path),
        model_card_path=Path(args.model_card_path),
        source_revision=args.source_revision,
        source_describe=args.source_describe,
        bundle_display_path=args.bundle_display_path,
        config_display_path=args.config_display_path,
        model_card_display_path=args.model_card_display_path,
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
