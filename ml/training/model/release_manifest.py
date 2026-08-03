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
import subprocess
from pathlib import Path
from typing import Any, Dict, Iterable, List, Optional, Tuple


PLACEHOLDER_PATTERN = re.compile(
    r"(?i)(placeholder|pending|tbd|not published yet|<path>|sha256:placeholder)"
)

REQUIRED_BUNDLE_FILES = (
    "manifest.json",
    "export_metadata.json",
    "MODEL_HASH.txt",
    "model_provenance.json",
    "model_frozen.pb",
)

MODEL_PROVENANCE_SCHEMA = "virtengine.model-provenance/v1"
PRODUCTION_PROVENANCE_STATUS = "production_approved"
BUILD_PROFILES = ("production", "fixture_only")
PROVENANCE_ROOT_KEYS = {
    "schema_version",
    "manifest_id",
    "status",
    "source",
    "stages",
    "artifacts",
    "datasets",
    "licenses",
    "redistribution",
    "bindings",
    "sbom",
    "model_card",
    "evaluation_report",
    "blockers",
}
STAGE_IDS = {"preprocessing", "training", "evaluation", "packaging"}
LOCAL_EVIDENCE_KEYS = {"id", "state", "path", "sha256", "size", "source"}
IMMUTABLE_SOURCE_KEYS = {"type", "uri", "revision", "digest"}
ID_PATTERN = re.compile(r"[a-z0-9.-]+")
SHA1_PATTERN = re.compile(r"[a-f0-9]{40}")
SHA256_PATTERN = re.compile(r"[a-f0-9]{64}")
SHA256_REVISION_PATTERN = re.compile(r"sha256:[a-f0-9]{64}")
REPO_ROOT = Path(__file__).resolve().parents[3]


class ReleaseManifestError(ValueError):
    """Raised when a model bundle cannot be released safely."""


def _read_text(path: Path) -> str:
    return path.read_text(encoding="utf-8")


def _load_json(path: Path) -> Dict[str, Any]:
    return _load_json_bytes(path.read_bytes(), path)


def _load_json_bytes(data: bytes, path: Path) -> Dict[str, Any]:
    try:
        payload = json.loads(data)
    except (json.JSONDecodeError, UnicodeDecodeError) as exc:
        raise ReleaseManifestError(f"invalid JSON in {path}: {exc}") from exc
    if not isinstance(payload, dict):
        raise ReleaseManifestError(f"JSON document must be an object: {path}")
    return payload


def _require_exact_keys(payload: Any, expected: set[str], field_name: str) -> Dict[str, Any]:
    if not isinstance(payload, dict):
        raise ReleaseManifestError(f"{field_name} must be an object")
    actual = set(payload)
    if actual != expected:
        missing = sorted(expected - actual)
        unknown = sorted(actual - expected)
        raise ReleaseManifestError(
            f"{field_name} keys must match schema; missing={missing}, unknown={unknown}"
        )
    return payload


def _require_id(value: Any, field_name: str) -> str:
    if not isinstance(value, str) or ID_PATTERN.fullmatch(value) is None:
        raise ReleaseManifestError(f"{field_name} must match ^[a-z0-9.-]+$")
    return value


def _require_sha1(value: Any, field_name: str) -> str:
    if not isinstance(value, str) or SHA1_PATTERN.fullmatch(value) is None:
        raise ReleaseManifestError(f"{field_name} must be a lowercase SHA-1 digest")
    return value


def _require_sha256(value: Any, field_name: str) -> str:
    if not isinstance(value, str) or SHA256_PATTERN.fullmatch(value) is None:
        raise ReleaseManifestError(f"{field_name} must be a lowercase SHA-256 digest")
    return value


def _validate_immutable_uri(source_type: str, uri: str, field_name: str) -> None:
    lowered = uri.lower()
    mutable_tokens = re.compile(r"(^|[/:?&=#-])(latest|main|master|branches|tags)(?=$|[/:?&=#-])")
    if mutable_tokens.search(lowered):
        raise ReleaseManifestError(f"{field_name} must not reference a mutable source")
    if source_type == "oci" and "@sha256:" not in lowered:
        raise ReleaseManifestError(f"{field_name} OCI reference must be pinned by digest")


def _read_repository_blob(
    repo_root: Path, source_commit: str, evidence_path: str, field_name: str
) -> bytes:
    try:
        result = subprocess.run(
            ["git", "-C", str(repo_root), "cat-file", "blob", f"{source_commit}:{evidence_path}"],
            check=False,
            capture_output=True,
        )
    except (OSError, ValueError, subprocess.SubprocessError) as exc:
        raise ReleaseManifestError(
            f"{field_name}.path cannot be resolved at model provenance source.commit"
        ) from exc
    if result.returncode != 0:
        raise ReleaseManifestError(
            f"{field_name}.path cannot be resolved at model provenance source.commit"
        )
    return result.stdout


def _repository_blob_id(
    repo_root: Path, revision: str, evidence_path: str, field_name: str
) -> str:
    try:
        result = subprocess.run(
            ["git", "-C", str(repo_root), "rev-parse", f"{revision}:{evidence_path}"],
            check=False,
            capture_output=True,
            text=True,
        )
    except (OSError, ValueError, subprocess.SubprocessError) as exc:
        raise ReleaseManifestError(
            f"{field_name}.path cannot be resolved at model provenance source.commit"
        ) from exc
    if result.returncode != 0:
        raise ReleaseManifestError(
            f"{field_name}.path cannot be resolved at model provenance source.commit"
        )
    return result.stdout.strip()


def _working_tree_blob_id(repo_root: Path, evidence_path: str, field_name: str) -> str:
    try:
        result = subprocess.run(
            [
                "git",
                "-C",
                str(repo_root),
                "hash-object",
                f"--path={evidence_path}",
                evidence_path,
            ],
            check=False,
            capture_output=True,
            text=True,
        )
    except (OSError, ValueError, subprocess.SubprocessError) as exc:
        raise ReleaseManifestError(
            f"{field_name}.path cannot be normalized through Git attributes"
        ) from exc
    if result.returncode != 0:
        raise ReleaseManifestError(
            f"{field_name}.path cannot be normalized through Git attributes"
        )
    return result.stdout.strip()


def _validate_local_evidence(
    payload: Any,
    field_name: str,
    repo_root: Path,
    evidence_ids: set[str],
    evidence_paths: set[str],
    source_commit: str,
) -> Dict[str, Any]:
    evidence = _require_exact_keys(payload, LOCAL_EVIDENCE_KEYS, field_name)
    evidence_id = _require_id(evidence["id"], f"{field_name}.id")
    if evidence_id in evidence_ids:
        raise ReleaseManifestError(f"duplicate production evidence id: {evidence_id}")
    evidence_ids.add(evidence_id)

    if evidence["state"] != "present":
        raise ReleaseManifestError(f"{field_name}.state must be present for production")
    if not isinstance(evidence["path"], str) or not evidence["path"].strip():
        raise ReleaseManifestError(f"{field_name}.path must be a nonempty string")
    evidence_path = evidence["path"]
    relative_path = Path(evidence_path)
    if (
        relative_path.is_absolute()
        or relative_path.drive
        or ".." in relative_path.parts
        or "\\" in evidence_path
        or "\x00" in evidence_path
        or relative_path.as_posix() != evidence_path
    ):
        raise ReleaseManifestError(f"{field_name}.path must be a normalized relative slash path")
    if evidence_path in evidence_paths:
        raise ReleaseManifestError(f"duplicate production evidence path: {evidence_path}")
    evidence_paths.add(evidence_path)

    evidence_sha = _require_sha256(evidence["sha256"], f"{field_name}.sha256")
    if isinstance(evidence["size"], bool) or not isinstance(evidence["size"], int) or evidence["size"] < 1:
        raise ReleaseManifestError(f"{field_name}.size must be an integer greater than zero")

    source = _require_exact_keys(evidence["source"], IMMUTABLE_SOURCE_KEYS, f"{field_name}.source")
    if source["type"] not in {"repository_file", "oci"}:
        raise ReleaseManifestError(f"{field_name}.source.type must be repository_file or oci")
    if not isinstance(source["uri"], str) or not source["uri"].strip():
        raise ReleaseManifestError(f"{field_name}.source.uri must be a nonempty string")
    revision = source["revision"]
    if not isinstance(revision, str) or not (
        SHA1_PATTERN.fullmatch(revision) or SHA256_REVISION_PATTERN.fullmatch(revision)
    ):
        raise ReleaseManifestError(f"{field_name}.source.revision must be an immutable digest")
    source_digest = _require_sha256(source["digest"], f"{field_name}.source.digest")
    if source_digest != evidence_sha:
        raise ReleaseManifestError(f"{field_name}.source.digest must equal evidence sha256")
    _validate_immutable_uri(source["type"], source["uri"], f"{field_name}.source.uri")
    source_uri = source["uri"]
    if source["type"] == "oci":
        if not source_uri.endswith(f"@sha256:{evidence_sha}"):
            raise ReleaseManifestError(f"{field_name}.source.uri digest must equal evidence sha256")
        if revision != f"sha256:{evidence_sha}":
            raise ReleaseManifestError(
                f"{field_name}.source.revision must equal sha256:evidence sha256"
            )
    if source["type"] == "repository_file" and revision not in source_uri and source_digest not in source_uri:
        raise ReleaseManifestError(f"{field_name}.source.uri must contain its immutable revision or digest")

    local_path = (repo_root / relative_path).resolve()
    try:
        local_path.relative_to(repo_root.resolve())
    except ValueError as exc:
        raise ReleaseManifestError(f"{field_name}.path escapes the repository root") from exc
    if not local_path.is_file():
        raise ReleaseManifestError(f"{field_name}.path is not a local evidence file: {evidence_path}")
    local_bytes = local_path.read_bytes()

    if source["type"] == "repository_file":
        if revision != source_commit:
            raise ReleaseManifestError(
                f"{field_name}.source.revision must equal model provenance source.commit"
            )
        committed_bytes = _read_repository_blob(repo_root, source_commit, evidence_path, field_name)
        committed_sha = hashlib.sha256(committed_bytes).hexdigest()
        if committed_sha != evidence_sha:
            raise ReleaseManifestError(f"{field_name}.sha256 does not match committed repository blob")
        if len(committed_bytes) != evidence["size"]:
            raise ReleaseManifestError(f"{field_name}.size does not match committed repository blob")

        committed_blob_id = _repository_blob_id(
            repo_root, source_commit, evidence_path, field_name
        )
        working_tree_blob_id = _working_tree_blob_id(repo_root, evidence_path, field_name)
        if working_tree_blob_id != committed_blob_id:
            raise ReleaseManifestError(
                f"{field_name}.path working-tree content does not match committed repository blob"
            )
    else:
        if len(local_bytes) != evidence["size"]:
            raise ReleaseManifestError(f"{field_name}.size does not match local evidence file")
        if hashlib.sha256(local_bytes).hexdigest() != evidence_sha:
            raise ReleaseManifestError(f"{field_name}.sha256 does not match local evidence file")

    return evidence


def _repo_relative_path(path: Path, repo_root: Path, field_name: str) -> str:
    try:
        return path.resolve().relative_to(repo_root.resolve()).as_posix()
    except ValueError as exc:
        raise ReleaseManifestError(f"{field_name} must be a repository file") from exc


def _validate_production_provenance(
    provenance: Dict[str, Any],
    repo_root: Path,
    *,
    source_revision: str,
    config_path: Path,
    model_card_path: Path,
    runtime_model_hash: str,
    frozen_graph_hash: str,
) -> Dict[str, str]:
    root = _require_exact_keys(provenance, PROVENANCE_ROOT_KEYS, "model provenance")
    if root["schema_version"] != MODEL_PROVENANCE_SCHEMA:
        raise ReleaseManifestError(
            f"model provenance schema_version must be {MODEL_PROVENANCE_SCHEMA}"
        )
    _require_id(root["manifest_id"], "model provenance manifest_id")
    if root["status"] != PRODUCTION_PROVENANCE_STATUS:
        raise ReleaseManifestError("production model provenance status must be production_approved")

    source = _require_exact_keys(root["source"], {"commit", "tree"}, "model provenance source")
    source_commit = _require_sha1(source["commit"], "model provenance source.commit")
    _require_sha1(source["tree"], "model provenance source.tree")
    if source_commit != source_revision:
        raise ReleaseManifestError("model provenance source.commit must equal source_revision")

    resolved: Dict[str, str] = {}
    for object_type in ("commit", "tree"):
        try:
            result = subprocess.run(
                ["git", "-C", str(repo_root), "rev-parse", f"{source_commit}^{{{object_type}}}"],
                check=False,
                capture_output=True,
                text=True,
            )
        except (OSError, subprocess.SubprocessError) as exc:
            raise ReleaseManifestError(
                f"model provenance source.commit cannot be resolved as a {object_type}"
            ) from exc
        if result.returncode != 0:
            raise ReleaseManifestError(
                f"model provenance source.commit cannot be resolved as a {object_type}"
            )
        resolved[object_type] = result.stdout.strip()

    if resolved["commit"] != source_commit:
        raise ReleaseManifestError("model provenance source.commit does not resolve exactly")
    if resolved["tree"] != source["tree"]:
        raise ReleaseManifestError("model provenance source.tree does not match source.commit")

    evidence_ids: set[str] = set()
    evidence_paths: set[str] = set()

    stages = root["stages"]
    if not isinstance(stages, list) or len(stages) != len(STAGE_IDS):
        raise ReleaseManifestError("model provenance stages must contain all production stages exactly once")
    stage_ids: set[str] = set()
    for index, payload in enumerate(stages):
        stage = _require_exact_keys(payload, {"id", "state", "evidence"}, f"stages[{index}]")
        stage_id = stage["id"]
        if stage_id not in STAGE_IDS or stage_id in stage_ids:
            raise ReleaseManifestError(f"stages[{index}].id must be a unique production stage id")
        stage_ids.add(stage_id)
        evidence = _validate_local_evidence(
            stage["evidence"],
            f"stages[{index}].evidence",
            repo_root,
            evidence_ids,
            evidence_paths,
            source_commit,
        )
        if stage["state"] != evidence["state"]:
            raise ReleaseManifestError(f"stages[{index}].state must match its evidence state")
    if stage_ids != STAGE_IDS:
        raise ReleaseManifestError("model provenance stages must contain all production stages exactly once")

    validated_arrays: Dict[str, List[Dict[str, Any]]] = {}
    for array_name in ("artifacts", "datasets", "licenses"):
        values = root[array_name]
        if not isinstance(values, list) or not values:
            raise ReleaseManifestError(f"model provenance {array_name} must be a nonempty array")
        validated_arrays[array_name] = []
        for index, evidence in enumerate(values):
            validated_arrays[array_name].append(
                _validate_local_evidence(
                    evidence,
                    f"{array_name}[{index}]",
                    repo_root,
                    evidence_ids,
                    evidence_paths,
                    source_commit,
                )
            )

    redistribution = _require_exact_keys(
        root["redistribution"], {"approved", "evidence"}, "model provenance redistribution"
    )
    if redistribution["approved"] is not True:
        raise ReleaseManifestError("model provenance redistribution.approved must be true")
    _validate_local_evidence(
        redistribution["evidence"],
        "redistribution.evidence",
        repo_root,
        evidence_ids,
        evidence_paths,
        source_commit,
    )

    bindings = _require_exact_keys(
        root["bindings"], {"preprocessing", "schema", "runtime"}, "model provenance bindings"
    )
    validated_bindings = {}
    for binding_name in ("preprocessing", "schema", "runtime"):
        validated_bindings[binding_name] = _validate_local_evidence(
            bindings[binding_name],
            f"bindings.{binding_name}",
            repo_root,
            evidence_ids,
            evidence_paths,
            source_commit,
        )
    config_relative_path = _repo_relative_path(config_path, repo_root, "config_path")
    preprocessing = validated_bindings["preprocessing"]
    if preprocessing["path"] != config_relative_path:
        raise ReleaseManifestError("bindings.preprocessing must match config_path")

    validated_fields = {}
    for field_name in ("sbom", "model_card", "evaluation_report"):
        validated_fields[field_name] = _validate_local_evidence(
            root[field_name],
            field_name,
            repo_root,
            evidence_ids,
            evidence_paths,
            source_commit,
        )
    model_card_relative_path = _repo_relative_path(model_card_path, repo_root, "model_card_path")
    model_card = validated_fields["model_card"]
    if model_card["path"] != model_card_relative_path:
        raise ReleaseManifestError("model_card evidence must match model_card_path")

    accepted_model_hashes = {runtime_model_hash, frozen_graph_hash}
    if not any(evidence["sha256"] in accepted_model_hashes for evidence in validated_arrays["artifacts"]):
        raise ReleaseManifestError(
            "model provenance artifacts must include the runtime model hash or frozen graph hash"
        )
    if not isinstance(root["blockers"], list) or root["blockers"]:
        raise ReleaseManifestError("production model provenance blockers must be an empty array")

    return {
        "config_sha256": preprocessing["sha256"],
        "model_card_sha256": model_card["sha256"],
    }


def _load_provenance(
    provenance_path: Path,
    profile: str,
) -> Tuple[Dict[str, Any], bytes]:
    if not provenance_path.is_file():
        raise ReleaseManifestError(f"model provenance not found: {provenance_path}")

    provenance_bytes = provenance_path.read_bytes()
    try:
        provenance_text = provenance_bytes.decode("utf-8")
    except UnicodeDecodeError as exc:
        raise ReleaseManifestError(f"invalid JSON in {provenance_path}: {exc}") from exc
    if PLACEHOLDER_PATTERN.search(provenance_text):
        raise ReleaseManifestError(f"placeholder content detected in {provenance_path}")
    provenance = _load_json_bytes(provenance_bytes, provenance_path)
    schema_version = _require_value(
        provenance.get("schema_version"),
        "model provenance schema_version",
    )
    status = _require_value(provenance.get("status"), "model provenance status")
    if schema_version != MODEL_PROVENANCE_SCHEMA:
        raise ReleaseManifestError(
            f"model provenance schema_version must be {MODEL_PROVENANCE_SCHEMA}"
        )
    if profile == "production" and status != PRODUCTION_PROVENANCE_STATUS:
        raise ReleaseManifestError(
            "production model provenance status must be production_approved"
        )
    return provenance, provenance_bytes


def _sha256_file(path: Path) -> str:
    hasher = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            hasher.update(chunk)
    return hasher.hexdigest()


def _sha256_model_dir(model_dir: Path) -> str:
    hasher = hashlib.sha256()
    files = sorted(
        (
            (path.relative_to(model_dir).as_posix(), path)
            for path in model_dir.rglob("*")
            if path.is_file() and path.name != "export_metadata.json"
        ),
        key=lambda item: item[0],
    )
    for _, path in files:
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


def _paths_overlap(left: Path, right: Path) -> bool:
    left = left.resolve()
    right = right.resolve()
    return left == right or left in right.parents or right in left.parents


def _validate_output_path(
    output_path: Path,
    bundle_dir: Path,
    config_path: Path,
    model_card_path: Path,
    provenance_path: Path,
) -> None:
    resolved_output = output_path.resolve()
    protected_paths = {config_path.resolve(), model_card_path.resolve(), provenance_path.resolve()}
    protected_paths.update((bundle_dir / relative).resolve() for relative in REQUIRED_BUNDLE_FILES)
    if resolved_output in protected_paths:
        raise ReleaseManifestError("output path must not overwrite a required input or bundle artifact")
    if _paths_overlap(resolved_output.parent, bundle_dir):
        raise ReleaseManifestError("output directory and bundle directory must be disjoint")


def build_release_manifest(
    bundle_dir: Path,
    *,
    model_name: str,
    version: str,
    profile: str,
    config_path: Path,
    model_card_path: Path,
    provenance_path: Path,
    source_revision: str,
    source_describe: str = "",
    bundle_display_path: Optional[str] = None,
    config_display_path: Optional[str] = None,
    model_card_display_path: Optional[str] = None,
    repo_root: Path = REPO_ROOT,
    validate_only: bool = False,
) -> Dict[str, Any]:
    bundle_dir = bundle_dir.resolve()
    model_dir = bundle_dir / "model"

    model_name = _require_value(model_name, "model_name")
    version = _require_value(version, "version")
    profile = _require_value(profile, "profile")
    if profile not in BUILD_PROFILES:
        raise ReleaseManifestError(f"profile must be one of: {', '.join(BUILD_PROFILES)}")
    source_revision = _require_value(source_revision, "source_revision")
    if profile == "production":
        _require_sha1(source_revision, "source_revision")
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

    repo_root = repo_root.resolve()
    provenance, provenance_bytes = _load_provenance(provenance_path, profile)
    if not model_dir.is_dir():
        raise ReleaseManifestError(f"required model directory is missing: {model_dir}")
    pre_copy_artifacts = [
        bundle_dir / relative for relative in REQUIRED_BUNDLE_FILES if relative != "model_provenance.json"
    ]
    for path in pre_copy_artifacts:
        if not path.exists():
            raise ReleaseManifestError(f"required bundle artifact is missing: {path}")
    pre_copy_artifacts.extend(
        sorted(candidate for candidate in model_dir.rglob("*") if candidate.is_file())
    )
    for path in pre_copy_artifacts:
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

    source_hashes = {
        "config_sha256": _sha256_file(config_path),
        "model_card_sha256": _sha256_file(model_card_path),
    }
    if provenance.get("status") == PRODUCTION_PROVENANCE_STATUS:
        source_hashes = _validate_production_provenance(
            provenance,
            repo_root,
            source_revision=source_revision,
            config_path=config_path,
            model_card_path=model_card_path,
            runtime_model_hash=runtime_model_hash,
            frozen_graph_hash=frozen_graph_hash,
        )

    if validate_only:
        return {}

    bundle_provenance_path = bundle_dir / "model_provenance.json"
    if provenance_path.resolve() == bundle_provenance_path.resolve():
        if bundle_provenance_path.read_bytes() != provenance_bytes:
            raise ReleaseManifestError("model provenance changed during validation")
    else:
        bundle_provenance_path.write_bytes(provenance_bytes)
    artifacts = list(_required_artifacts(bundle_dir))
    artifact_records = sorted(
        (_artifact_record(path, bundle_dir) for path in artifacts),
        key=lambda item: item["path"],
    )
    artifact_index_sha256 = _canonical_json_sha256(artifact_records)

    return {
        "schema_version": "veid.release.manifest/v1",
        "profile": profile,
        "provenance": {
            "path": "model_provenance.json",
            "sha256": _sha256_file(bundle_provenance_path),
        },
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
            "config_sha256": source_hashes["config_sha256"],
            "model_card_path": _display_path(
                model_card_path,
                model_card_display_path,
                "model_card_display_path",
            ),
            "model_card_sha256": source_hashes["model_card_sha256"],
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
    provenance_path: Path,
    source_revision: str,
    source_describe: str = "",
    bundle_display_path: Optional[str] = None,
    config_display_path: Optional[str] = None,
    model_card_display_path: Optional[str] = None,
    repo_root: Path = REPO_ROOT,
) -> Dict[str, Any]:
    _validate_output_path(output_path, bundle_dir, config_path, model_card_path, provenance_path)
    manifest = build_release_manifest(
        bundle_dir,
        model_name=model_name,
        version=version,
        profile=profile,
        config_path=config_path,
        model_card_path=model_card_path,
        provenance_path=provenance_path,
        source_revision=source_revision,
        source_describe=source_describe,
        bundle_display_path=bundle_display_path,
        config_display_path=config_display_path,
        model_card_display_path=model_card_display_path,
        repo_root=repo_root,
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
    parser.add_argument("--profile", choices=BUILD_PROFILES, default="production", help="Build profile")
    parser.add_argument("--config-path", required=True, help="Training config path")
    parser.add_argument("--model-card-path", required=True, help="Model card path")
    parser.add_argument("--provenance-path", required=True, help="Model provenance JSON path")
    parser.add_argument("--source-revision", required=True, help="Source revision for provenance capture")
    parser.add_argument("--source-describe", default="", help="Optional source describe/tag value")
    parser.add_argument("--bundle-display-path", default=None, help="Portable path label for the bundle")
    parser.add_argument("--config-display-path", default=None, help="Portable path label for the training config")
    parser.add_argument("--model-card-display-path", default=None, help="Portable path label for the model card")
    parser.add_argument(
        "--validate-only",
        action="store_true",
        help="Validate all inputs without copying provenance or writing a manifest",
    )
    return parser.parse_args()


def main() -> int:
    args = _parse_args()
    if args.validate_only:
        build_release_manifest(
            Path(args.bundle_dir),
            model_name=args.model_name,
            version=args.version,
            profile=args.profile,
            config_path=Path(args.config_path),
            model_card_path=Path(args.model_card_path),
            provenance_path=Path(args.provenance_path),
            source_revision=args.source_revision,
            source_describe=args.source_describe,
            bundle_display_path=args.bundle_display_path,
            config_display_path=args.config_display_path,
            model_card_display_path=args.model_card_display_path,
            validate_only=True,
        )
        return 0
    write_release_manifest(
        Path(args.output),
        bundle_dir=Path(args.bundle_dir),
        model_name=args.model_name,
        version=args.version,
        profile=args.profile,
        config_path=Path(args.config_path),
        model_card_path=Path(args.model_card_path),
        provenance_path=Path(args.provenance_path),
        source_revision=args.source_revision,
        source_describe=args.source_describe,
        bundle_display_path=args.bundle_display_path,
        config_display_path=args.config_display_path,
        model_card_display_path=args.model_card_display_path,
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
