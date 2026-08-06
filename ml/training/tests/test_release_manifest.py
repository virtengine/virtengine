import json
import os
import shlex
import shutil
import subprocess
from pathlib import Path

import pytest

import ml.training.model.release_manifest as release_manifest_module
from ml.training.model.release_manifest import (
    ReleaseManifestError,
    _sha256_model_dir,
    _validate_local_evidence,
    build_release_manifest,
    write_release_manifest,
)


def create_bundle(tmp_path: Path, *, version: str = "v1.0.0") -> Path:
    bundle_dir = tmp_path / version
    model_dir = bundle_dir / "model" / "variables"
    model_dir.mkdir(parents=True)

    (bundle_dir / "model" / "saved_model.pb").write_bytes(b"saved-model")
    (model_dir / "variables.data-00000-of-00001").write_bytes(b"weights")
    (model_dir / "variables.index").write_bytes(b"index")
    frozen_model_bytes = subprocess.run(
        ["git", "-C", str(repo_root()), "cat-file", "blob", "HEAD:SECURITY.md"],
        check=True,
        capture_output=True,
    ).stdout
    (bundle_dir / "model_frozen.pb").write_bytes(frozen_model_bytes)

    runtime_hash = subprocess_hash_model_dir(bundle_dir / "model")
    frozen_hash = sha256_file(bundle_dir / "model_frozen.pb")

    training_manifest = {
        "model_version": version,
        "model_hash": runtime_hash,
        "input_signature": {"name": "features", "shape": [None, 768], "dtype": "float32"},
        "output_signature": {"name": "trust_score", "shape": [None, 1], "dtype": "float32"},
        "tensorflow_version": "2.15.0",
    }
    (bundle_dir / "manifest.json").write_text(json.dumps(training_manifest, indent=2), encoding="utf-8")

    export_metadata = {
        "version": version,
        "model_hash": frozen_hash,
        "signature_name": "serving_default",
        "input_signature": training_manifest["input_signature"],
        "output_signature": training_manifest["output_signature"],
        "tensorflow_version": "2.15.0",
    }
    (bundle_dir / "export_metadata.json").write_text(json.dumps(export_metadata, indent=2), encoding="utf-8")

    (bundle_dir / "MODEL_HASH.txt").write_text(
        "\n".join(
            [
                "# Trust score model release hash",
                f"SHA256={runtime_hash}",
                f"VERSION={version}",
                "",
            ]
        ),
        encoding="utf-8",
    )

    return bundle_dir


def create_model_provenance(
    tmp_path: Path,
    *,
    status: str = "production_approved",
    schema_version: str = "virtengine.model-provenance/v1",
) -> Path:
    evidence_paths = [
        "README.md",
        "LICENSE",
        "go.mod",
        "Makefile",
        "SECURITY.md",
        "PRIVACY_POLICY.md",
        "TERMS_OF_SERVICE.md",
        "VERIFICATION.md",
        "ml/training/configs/trust_score_v1.yaml",
        "CONSENT_FRAMEWORK.md",
        "SUPPLY_CHAIN_SECURITY.md",
        "CHANGELOG.md",
        "models/trust_score/MODEL_CARD.md",
        "DATA_INVENTORY.md",
    ]
    revision = source_revision()

    def evidence(index: int, evidence_id: str) -> dict:
        relative_path = evidence_paths[index]
        committed_bytes = subprocess.run(
            ["git", "-C", str(repo_root()), "cat-file", "blob", f"{revision}:{relative_path}"],
            check=True,
            capture_output=True,
        ).stdout
        import hashlib

        digest = hashlib.sha256(committed_bytes).hexdigest()
        return {
            "id": evidence_id,
            "state": "present",
            "path": relative_path,
            "sha256": digest,
            "size": len(committed_bytes),
            "source": {
                "type": "repository_file",
                "uri": f"repo:{relative_path}@{revision}",
                "revision": revision,
                "digest": digest,
            },
        }

    payload = {
        "schema_version": schema_version,
        "manifest_id": "trust-score-v1.0.0",
        "status": status,
        "source": {"commit": revision, "tree": source_tree()},
        "stages": [
            {"id": "preprocessing", "state": "present", "evidence": evidence(0, "stage.preprocessing")},
            {"id": "training", "state": "present", "evidence": evidence(1, "stage.training")},
            {"id": "evaluation", "state": "present", "evidence": evidence(2, "stage.evaluation")},
            {"id": "packaging", "state": "present", "evidence": evidence(3, "stage.packaging")},
        ],
        "artifacts": [evidence(4, "artifact.model")],
        "datasets": [evidence(5, "dataset.training")],
        "licenses": [evidence(6, "license.model")],
        "redistribution": {"approved": True, "evidence": evidence(7, "redistribution.approval")},
        "bindings": {
            "preprocessing": evidence(8, "binding.preprocessing"),
            "schema": evidence(9, "binding.schema"),
            "runtime": evidence(10, "binding.runtime"),
        },
        "sbom": evidence(11, "sbom.release"),
        "model_card": evidence(12, "model-card.release"),
        "evaluation_report": evidence(13, "evaluation.release"),
        "blockers": [],
    }
    provenance_path = tmp_path / "source-model-provenance.json"
    provenance_path.write_text(
        json.dumps(payload, indent=2, sort_keys=True) + "\n",
        encoding="utf-8",
    )
    return provenance_path


def sha256_file(path: Path) -> str:
    import hashlib

    hasher = hashlib.sha256()
    hasher.update(path.read_bytes())
    return hasher.hexdigest()


def subprocess_hash_model_dir(model_dir: Path) -> str:
    import hashlib

    hasher = hashlib.sha256()
    paths = sorted(
        (
            (candidate.relative_to(model_dir).as_posix(), candidate)
            for candidate in model_dir.rglob("*")
            if candidate.is_file() and candidate.name != "export_metadata.json"
        ),
        key=lambda item: item[0],
    )
    for _, path in paths:
        hasher.update(path.read_bytes())
    return hasher.hexdigest()


def repo_root() -> Path:
    return Path(__file__).resolve().parents[3]


def git_revision(expression: str) -> str:
    return subprocess.run(
        ["git", "-C", str(repo_root()), "rev-parse", expression],
        check=True,
        capture_output=True,
        text=True,
    ).stdout.strip()


def source_revision() -> str:
    return git_revision("HEAD")


def source_tree() -> str:
    return git_revision("HEAD^{tree}")


def create_evidence_repository(tmp_path: Path) -> tuple[Path, str]:
    repository = tmp_path / "evidence-repository"
    repository.mkdir()
    subprocess.run(["git", "-C", str(repository), "init", "--quiet"], check=True)
    (repository / "evidence.txt").write_bytes(b"committed evidence\n")
    subprocess.run(["git", "-C", str(repository), "add", "evidence.txt"], check=True)
    subprocess.run(
        [
            "git",
            "-C",
            str(repository),
            "-c",
            "user.name=Release Manifest Tests",
            "-c",
            "user.email=release-manifest@example.invalid",
            "commit",
            "--quiet",
            "-m",
            "test evidence",
        ],
        check=True,
    )
    revision = subprocess.run(
        ["git", "-C", str(repository), "rev-parse", "HEAD"],
        check=True,
        capture_output=True,
        text=True,
    ).stdout.strip()
    return repository, revision


def repository_evidence(repository: Path, revision: str, path: str = "evidence.txt") -> dict:
    evidence_path = repository / path
    digest = sha256_file(evidence_path)
    return {
        "id": "artifact.repository-evidence",
        "state": "present",
        "path": path,
        "sha256": digest,
        "size": evidence_path.stat().st_size,
        "source": {
            "type": "repository_file",
            "uri": f"repo:{path}@{revision}",
            "revision": revision,
            "digest": digest,
        },
    }


def oci_evidence(repository: Path, path: str = "evidence.txt") -> dict:
    evidence_path = repository / path
    digest = sha256_file(evidence_path)
    return {
        "id": "artifact.oci-evidence",
        "state": "present",
        "path": path,
        "sha256": digest,
        "size": evidence_path.stat().st_size,
        "source": {
            "type": "oci",
            "uri": f"registry.example.invalid/models/evidence@sha256:{digest}",
            "revision": f"sha256:{digest}",
            "digest": digest,
        },
    }


def config_path() -> Path:
    return repo_root() / "ml" / "training" / "configs" / "trust_score_v1.yaml"


def model_card_path() -> Path:
    return repo_root() / "models" / "trust_score" / "MODEL_CARD.md"


def shell_path_arg(path: Path) -> str:
    try:
        return path.relative_to(repo_root()).as_posix()
    except ValueError:
        return path.as_posix()


def find_bash() -> str | None:
    candidates = [
        Path(r"C:\Program Files\Git\bin\bash.exe"),
        Path(r"C:\Program Files\Git\usr\bin\bash.exe"),
    ]
    for candidate in candidates:
        if candidate.is_file():
            return str(candidate)
    return None


BASH = find_bash()


def test_release_manifest_is_deterministic(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path)

    manifest_a = build_release_manifest(
        bundle_dir,
        model_name="trust_score",
        version="v1.0.0",
        profile="production",
        config_path=config_path(),
        model_card_path=model_card_path(),
        provenance_path=provenance_path,
        source_revision=source_revision(),
    )
    manifest_b = build_release_manifest(
        bundle_dir,
        model_name="trust_score",
        version="v1.0.0",
        profile="production",
        config_path=config_path(),
        model_card_path=model_card_path(),
        provenance_path=provenance_path,
        source_revision=source_revision(),
    )

    assert manifest_a == manifest_b


def test_release_manifest_uses_portable_source_paths(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path)
    revision = source_revision()

    manifest = build_release_manifest(
        bundle_dir,
        model_name="trust_score",
        version="v1.0.0",
        profile="production",
        config_path=config_path(),
        model_card_path=model_card_path(),
        provenance_path=provenance_path,
        source_revision=revision,
        source_describe="v1.0.0-1-gdeadbeef",
        bundle_display_path="models/trust_score/v1.0.0",
        config_display_path="ml/training/configs/trust_score_v1.yaml",
        model_card_display_path="models/trust_score/MODEL_CARD.md",
    )

    source = manifest["source"]
    assert source["bundle_path"] == "models/trust_score/v1.0.0"
    assert source["config_path"] == "ml/training/configs/trust_score_v1.yaml"
    assert source["model_card_path"] == "models/trust_score/MODEL_CARD.md"
    assert source["revision"] == revision
    assert source["revision_short"] == revision[:12]
    assert source["describe"] == "v1.0.0-1-gdeadbeef"
    assert source["bundle_artifact_index_sha256"]


def test_release_manifest_rejects_partial_bundle(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path)
    os.remove(bundle_dir / "model_frozen.pb")

    with pytest.raises(ReleaseManifestError, match="required bundle artifact is missing"):
        build_release_manifest(
            bundle_dir,
            model_name="trust_score",
            version="v1.0.0",
            profile="production",
            config_path=config_path(),
            model_card_path=model_card_path(),
            provenance_path=provenance_path,
            source_revision=source_revision(),
        )


def test_release_manifest_rejects_placeholder_hashes(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path)
    (bundle_dir / "MODEL_HASH.txt").write_text(
        "SHA256=placeholder\nVERSION=v1.0.0\n",
        encoding="utf-8",
    )

    with pytest.raises(ReleaseManifestError, match="placeholder content detected"):
        build_release_manifest(
            bundle_dir,
            model_name="trust_score",
            version="v1.0.0",
            profile="production",
            config_path=config_path(),
            model_card_path=model_card_path(),
            provenance_path=provenance_path,
            source_revision=source_revision(),
        )


def test_release_manifest_includes_and_binds_production_provenance(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path)
    source_bytes = provenance_path.read_bytes()

    manifest = build_release_manifest(
        bundle_dir,
        model_name="trust_score",
        version="v1.0.0",
        profile="production",
        config_path=config_path(),
        model_card_path=model_card_path(),
        provenance_path=provenance_path,
        source_revision=source_revision(),
    )

    bundled_provenance = bundle_dir / "model_provenance.json"
    provenance_artifact = next(
        artifact for artifact in manifest["artifacts"] if artifact["path"] == "model_provenance.json"
    )
    assert provenance_path.read_bytes() == source_bytes
    assert bundled_provenance.read_bytes() == source_bytes
    assert manifest["provenance"] == {
        "path": "model_provenance.json",
        "sha256": sha256_file(bundled_provenance),
    }
    assert provenance_artifact["sha256"] == manifest["provenance"]["sha256"]
    assert provenance_artifact["size_bytes"] == bundled_provenance.stat().st_size


def test_release_manifest_writes_captured_provenance_bytes(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch
):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path)
    captured_bytes = provenance_path.read_bytes()
    captured_provenance = json.loads(captured_bytes)

    def validate_then_mutate(*args, **kwargs):
        provenance_path.write_bytes(b"mutated after validation\n")
        return {
            "config_sha256": captured_provenance["bindings"]["preprocessing"]["sha256"],
            "model_card_sha256": captured_provenance["model_card"]["sha256"],
        }

    monkeypatch.setattr(
        release_manifest_module,
        "_validate_production_provenance",
        validate_then_mutate,
    )

    build_release_manifest(
        bundle_dir,
        model_name="trust_score",
        version="v1.0.0",
        profile="production",
        config_path=config_path(),
        model_card_path=model_card_path(),
        provenance_path=provenance_path,
        source_revision=source_revision(),
    )

    assert provenance_path.read_bytes() == b"mutated after validation\n"
    assert (bundle_dir / "model_provenance.json").read_bytes() == captured_bytes


def test_release_manifest_rejects_missing_provenance(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)

    with pytest.raises(ReleaseManifestError, match="model provenance not found"):
        build_release_manifest(
            bundle_dir,
            model_name="trust_score",
            version="v1.0.0",
            profile="production",
            config_path=config_path(),
            model_card_path=model_card_path(),
            provenance_path=tmp_path / "missing-provenance.json",
            source_revision=source_revision(),
        )


def test_release_manifest_rejects_dependency_blocked_production_provenance(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path, status="dependency_blocked")

    with pytest.raises(ReleaseManifestError, match="must be production_approved"):
        build_release_manifest(
            bundle_dir,
            model_name="trust_score",
            version="v1.0.0",
            profile="production",
            config_path=config_path(),
            model_card_path=model_card_path(),
            provenance_path=provenance_path,
            source_revision=source_revision(),
        )


def test_release_manifest_rejects_wrong_provenance_schema(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path, schema_version="virtengine.model-provenance/v2")

    with pytest.raises(ReleaseManifestError, match="schema_version must be"):
        build_release_manifest(
            bundle_dir,
            model_name="trust_score",
            version="v1.0.0",
            profile="production",
            config_path=config_path(),
            model_card_path=model_card_path(),
            provenance_path=provenance_path,
            source_revision=source_revision(),
        )


def test_release_manifest_provenance_tamper_changes_digest_deterministically(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path)

    def generate() -> dict:
        return build_release_manifest(
            bundle_dir,
            model_name="trust_score",
            version="v1.0.0",
            profile="production",
            config_path=config_path(),
            model_card_path=model_card_path(),
            provenance_path=provenance_path,
            source_revision=source_revision(),
        )

    original = generate()
    payload = json.loads(provenance_path.read_text(encoding="utf-8"))
    payload["manifest_id"] = "trust-score-v1.0.1"
    provenance_path.write_text(json.dumps(payload, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    changed = generate()

    assert changed == generate()
    assert original["provenance"]["sha256"] != changed["provenance"]["sha256"]


def test_release_manifest_accepts_provenance_already_in_bundle(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path)
    bundled_provenance = bundle_dir / "model_provenance.json"
    shutil.copyfile(provenance_path, bundled_provenance)

    manifest = build_release_manifest(
        bundle_dir,
        model_name="trust_score",
        version="v1.0.0",
        profile="production",
        config_path=config_path(),
        model_card_path=model_card_path(),
        provenance_path=bundled_provenance,
        source_revision=source_revision(),
    )

    assert manifest["provenance"]["sha256"] == sha256_file(bundled_provenance)


def test_release_manifest_rejects_changed_provenance_already_in_bundle(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch
):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path)
    bundled_provenance = bundle_dir / "model_provenance.json"
    shutil.copyfile(provenance_path, bundled_provenance)

    def validate_then_mutate(*args, **kwargs):
        bundled_provenance.write_bytes(b"mutated after validation\n")

    monkeypatch.setattr(
        release_manifest_module,
        "_validate_production_provenance",
        validate_then_mutate,
    )

    with pytest.raises(ReleaseManifestError, match="changed during validation"):
        build_release_manifest(
            bundle_dir,
            model_name="trust_score",
            version="v1.0.0",
            profile="production",
            config_path=config_path(),
            model_card_path=model_card_path(),
            provenance_path=bundled_provenance,
            source_revision=source_revision(),
        )


def test_sha256_model_dir_sorts_relative_slash_paths(tmp_path: Path):
    model_dir = tmp_path / "model"
    (model_dir / "a").mkdir(parents=True)
    (model_dir / "a" / "x").write_bytes(b"first")
    (model_dir / "a0").write_bytes(b"second")

    import hashlib

    assert _sha256_model_dir(model_dir) == hashlib.sha256(b"firstsecond").hexdigest()


@pytest.mark.parametrize("profile", ["Production", "prod"])
def test_release_manifest_rejects_profile_aliases(tmp_path: Path, profile: str):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path)

    with pytest.raises(ReleaseManifestError, match="profile must be one of"):
        build_release_manifest(
            bundle_dir,
            model_name="trust_score",
            version="v1.0.0",
            profile=profile,
            config_path=config_path(),
            model_card_path=model_card_path(),
            provenance_path=provenance_path,
            source_revision=source_revision(),
        )


def test_release_manifest_rejects_two_field_production_assertion(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = tmp_path / "provenance.json"
    provenance_path.write_text(
        json.dumps({"schema_version": "virtengine.model-provenance/v1", "status": "production_approved"}),
        encoding="utf-8",
    )

    with pytest.raises(ReleaseManifestError, match="keys must match schema"):
        build_release_manifest(
            bundle_dir,
            model_name="trust_score",
            version="v1.0.0",
            profile="production",
            config_path=config_path(),
            model_card_path=model_card_path(),
            provenance_path=provenance_path,
            source_revision=source_revision(),
        )


def test_release_manifest_rejects_two_field_production_assertion_in_fixture_profile(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = tmp_path / "provenance.json"
    provenance_path.write_text(
        json.dumps({"schema_version": "virtengine.model-provenance/v1", "status": "production_approved"}),
        encoding="utf-8",
    )

    with pytest.raises(ReleaseManifestError, match="keys must match schema"):
        build_release_manifest(
            bundle_dir,
            model_name="trust_score",
            version="v1.0.0",
            profile="fixture_only",
            config_path=config_path(),
            model_card_path=model_card_path(),
            provenance_path=provenance_path,
            source_revision=source_revision(),
        )


def test_release_manifest_rejects_fixture_evidence_in_production(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path)
    payload = json.loads(provenance_path.read_text(encoding="utf-8"))
    payload["artifacts"][0]["state"] = "fixture_only"
    provenance_path.write_text(json.dumps(payload), encoding="utf-8")

    with pytest.raises(ReleaseManifestError, match="state must be present"):
        build_release_manifest(
            bundle_dir,
            model_name="trust_score",
            version="v1.0.0",
            profile="production",
            config_path=config_path(),
            model_card_path=model_card_path(),
            provenance_path=provenance_path,
            source_revision=source_revision(),
        )


def test_release_manifest_rejects_unknown_nested_key(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path)
    payload = json.loads(provenance_path.read_text(encoding="utf-8"))
    payload["bindings"]["runtime"]["source"]["tofu"] = True
    provenance_path.write_text(json.dumps(payload), encoding="utf-8")

    with pytest.raises(ReleaseManifestError, match="unknown=\\['tofu'\\]"):
        build_release_manifest(
            bundle_dir,
            model_name="trust_score",
            version="v1.0.0",
            profile="production",
            config_path=config_path(),
            model_card_path=model_card_path(),
            provenance_path=provenance_path,
            source_revision=source_revision(),
        )


def test_release_manifest_rejects_source_hash_mismatch(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path)
    payload = json.loads(provenance_path.read_text(encoding="utf-8"))
    payload["datasets"][0]["sha256"] = "0" * 64
    payload["datasets"][0]["source"]["digest"] = "0" * 64
    provenance_path.write_text(json.dumps(payload), encoding="utf-8")

    with pytest.raises(ReleaseManifestError, match="sha256 does not match committed repository blob"):
        build_release_manifest(
            bundle_dir,
            model_name="trust_score",
            version="v1.0.0",
            profile="production",
            config_path=config_path(),
            model_card_path=model_card_path(),
            provenance_path=provenance_path,
            source_revision=source_revision(),
        )


def test_repository_evidence_rejects_dirty_worktree_with_updated_manifest_hash(tmp_path: Path):
    repository, revision = create_evidence_repository(tmp_path)
    (repository / "evidence.txt").write_bytes(b"dirty evidence\n")
    evidence = repository_evidence(repository, revision)

    with pytest.raises(ReleaseManifestError, match="sha256 does not match committed repository blob"):
        _validate_local_evidence(
            evidence,
            "artifacts[0]",
            repository,
            set(),
            set(),
            revision,
        )


def test_repository_evidence_rejects_path_absent_at_source_commit(tmp_path: Path):
    repository, revision = create_evidence_repository(tmp_path)
    (repository / "later.txt").write_bytes(b"not committed\n")
    evidence = repository_evidence(repository, revision, "later.txt")

    with pytest.raises(ReleaseManifestError, match="path cannot be resolved at model provenance source.commit"):
        _validate_local_evidence(
            evidence,
            "artifacts[0]",
            repository,
            set(),
            set(),
            revision,
        )


def test_repository_evidence_rejects_committed_blob_hash_mismatch(tmp_path: Path):
    repository, revision = create_evidence_repository(tmp_path)
    evidence = repository_evidence(repository, revision)
    evidence["sha256"] = "0" * 64
    evidence["source"]["digest"] = "0" * 64

    with pytest.raises(ReleaseManifestError, match="sha256 does not match committed repository blob"):
        _validate_local_evidence(
            evidence,
            "artifacts[0]",
            repository,
            set(),
            set(),
            revision,
        )


@pytest.mark.parametrize(
    ("mutation", "message"),
    [
        (lambda evidence: evidence.update(path="missing.txt"), "not a local evidence file"),
        (lambda evidence: evidence.update(size=evidence["size"] + 1), "size does not match local"),
        (
            lambda evidence: evidence["source"].update(revision="sha256:" + "0" * 64),
            "source.revision must equal",
        ),
        (
            lambda evidence: evidence["source"].update(
                uri="registry.example.invalid/models/unrelated@sha256:" + "0" * 64
            ),
            "source.uri digest must equal evidence sha256",
        ),
        (
            lambda evidence: evidence["source"].update(digest="0" * 64),
            "source.digest must equal evidence sha256",
        ),
    ],
)
def test_oci_evidence_rejects_unbound_or_missing_local_evidence(
    tmp_path: Path, mutation, message: str
):
    (tmp_path / "evidence.txt").write_bytes(b"local OCI evidence\n")
    evidence = oci_evidence(tmp_path)
    mutation(evidence)

    with pytest.raises(ReleaseManifestError, match=message):
        _validate_local_evidence(
            evidence,
            "artifacts[0]",
            tmp_path,
            set(),
            set(),
            "0" * 40,
        )


def test_oci_evidence_rejects_local_byte_hash_mismatch(tmp_path: Path):
    evidence_path = tmp_path / "evidence.txt"
    evidence_path.write_bytes(b"original OCI evidence\n")
    evidence = oci_evidence(tmp_path)
    evidence_path.write_bytes(b"tampered OCI evidence\n")

    with pytest.raises(ReleaseManifestError, match="sha256 does not match local evidence file"):
        _validate_local_evidence(
            evidence,
            "artifacts[0]",
            tmp_path,
            set(),
            set(),
            "0" * 40,
        )


def test_release_manifest_rejects_wrong_source_revision(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path)

    with pytest.raises(ReleaseManifestError, match="source.commit must equal source_revision"):
        build_release_manifest(
            bundle_dir,
            model_name="trust_score",
            version="v1.0.0",
            profile="production",
            config_path=config_path(),
            model_card_path=model_card_path(),
            provenance_path=provenance_path,
            source_revision="0" * 40,
        )


def test_release_manifest_rejects_wrong_source_tree(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path)
    payload = json.loads(provenance_path.read_text(encoding="utf-8"))
    payload["source"]["tree"] = "0" * 40
    provenance_path.write_text(json.dumps(payload), encoding="utf-8")

    with pytest.raises(ReleaseManifestError, match="source.tree does not match"):
        build_release_manifest(
            bundle_dir,
            model_name="trust_score",
            version="v1.0.0",
            profile="production",
            config_path=config_path(),
            model_card_path=model_card_path(),
            provenance_path=provenance_path,
            source_revision=source_revision(),
        )


def test_release_manifest_rejects_nonexistent_source_commit(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path)
    nonexistent_commit = "0" * 40
    payload = json.loads(provenance_path.read_text(encoding="utf-8"))
    payload["source"]["commit"] = nonexistent_commit
    provenance_path.write_text(json.dumps(payload), encoding="utf-8")

    with pytest.raises(ReleaseManifestError, match="source.commit cannot be resolved as a commit"):
        build_release_manifest(
            bundle_dir,
            model_name="trust_score",
            version="v1.0.0",
            profile="production",
            config_path=config_path(),
            model_card_path=model_card_path(),
            provenance_path=provenance_path,
            source_revision=nonexistent_commit,
        )


def test_release_manifest_rejects_evidence_revision_drift(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path)
    payload = json.loads(provenance_path.read_text(encoding="utf-8"))
    payload["datasets"][0]["source"]["revision"] = "0" * 40
    payload["datasets"][0]["source"]["uri"] = (
        f"repo:{payload['datasets'][0]['path']}@{'0' * 40}"
    )
    provenance_path.write_text(json.dumps(payload), encoding="utf-8")

    with pytest.raises(ReleaseManifestError, match="source.revision must equal"):
        build_release_manifest(
            bundle_dir,
            model_name="trust_score",
            version="v1.0.0",
            profile="production",
            config_path=config_path(),
            model_card_path=model_card_path(),
            provenance_path=provenance_path,
            source_revision=source_revision(),
        )


@pytest.mark.parametrize(
    ("left", "right", "message"),
    [
        (("bindings", "preprocessing"), ("bindings", "schema"), "bindings.preprocessing"),
        (("model_card",), ("evaluation_report",), "model_card evidence"),
    ],
)
def test_release_manifest_rejects_supplied_file_binding_mismatch(
    tmp_path: Path, left: tuple[str, ...], right: tuple[str, ...], message: str
):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path)
    payload = json.loads(provenance_path.read_text(encoding="utf-8"))

    def parent(keys: tuple[str, ...]) -> tuple[dict, str]:
        value = payload
        for key in keys[:-1]:
            value = value[key]
        return value, keys[-1]

    left_parent, left_key = parent(left)
    right_parent, right_key = parent(right)
    left_parent[left_key], right_parent[right_key] = right_parent[right_key], left_parent[left_key]
    provenance_path.write_text(json.dumps(payload), encoding="utf-8")

    with pytest.raises(ReleaseManifestError, match=message):
        build_release_manifest(
            bundle_dir,
            model_name="trust_score",
            version="v1.0.0",
            profile="production",
            config_path=config_path(),
            model_card_path=model_card_path(),
            provenance_path=provenance_path,
            source_revision=source_revision(),
        )


def test_release_manifest_rejects_provenance_without_matching_model_artifact(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path)
    payload = json.loads(provenance_path.read_text(encoding="utf-8"))
    payload["artifacts"], payload["datasets"] = payload["datasets"], payload["artifacts"]
    provenance_path.write_text(json.dumps(payload), encoding="utf-8")

    with pytest.raises(ReleaseManifestError, match="artifacts must include"):
        build_release_manifest(
            bundle_dir,
            model_name="trust_score",
            version="v1.0.0",
            profile="production",
            config_path=config_path(),
            model_card_path=model_card_path(),
            provenance_path=provenance_path,
            source_revision=source_revision(),
        )


@pytest.mark.parametrize("existing", [False, True])
def test_failed_validation_does_not_mutate_bundled_provenance(tmp_path: Path, existing: bool):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path)
    bundled_provenance = bundle_dir / "model_provenance.json"
    if existing:
        bundled_provenance.write_bytes(b"existing provenance sentinel")

    with pytest.raises(ReleaseManifestError, match="source.commit must equal source_revision"):
        build_release_manifest(
            bundle_dir,
            model_name="trust_score",
            version="v1.0.0",
            profile="production",
            config_path=config_path(),
            model_card_path=model_card_path(),
            provenance_path=provenance_path,
            source_revision="0" * 40,
        )

    if existing:
        assert bundled_provenance.read_bytes() == b"existing provenance sentinel"
    else:
        assert not bundled_provenance.exists()


@pytest.mark.parametrize("output_kind", ["alias", "parent", "nested"])
def test_write_release_manifest_rejects_output_aliases_and_overlap(tmp_path: Path, output_kind: str):
    bundle_dir = create_bundle(tmp_path / "bundle-root")
    provenance_path = create_model_provenance(tmp_path)
    if output_kind == "alias":
        output_path = provenance_path
    elif output_kind == "parent":
        output_path = tmp_path / "bundle-root" / "release_manifest.json"
    else:
        output_path = bundle_dir / "output" / "release_manifest.json"

    with pytest.raises(ReleaseManifestError, match="must not overwrite|must be disjoint"):
        write_release_manifest(
            output_path,
            bundle_dir=bundle_dir,
            model_name="trust_score",
            version="v1.0.0",
            profile="production",
            config_path=config_path(),
            model_card_path=model_card_path(),
            provenance_path=provenance_path,
            source_revision=source_revision(),
        )


@pytest.mark.skipif(BASH is None, reason="Git Bash is required for build script verification")
def test_build_script_manifest_only_is_repeatable(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path)
    output_a = tmp_path / "out-a"
    output_b = tmp_path / "out-b"
    script_path = "./_build/build-veid-pipeline.sh"

    for output_dir in (output_a, output_b):
        bundle_arg = shell_path_arg(bundle_dir)
        output_arg = shell_path_arg(output_dir)
        config_arg = config_path().relative_to(repo_root()).as_posix()
        model_card_arg = model_card_path().relative_to(repo_root()).as_posix()
        completed = subprocess.run(
            [
                BASH,
                "-lc",
                " ".join(
                    [
                        "VEID_SKIP_DOCKER_BUILD=true",
                        f"VEID_MODEL_BUNDLE_DIR={shlex.quote(bundle_arg)}",
                        f"VEID_OUTPUT_DIR={shlex.quote(output_arg)}",
                        f"VEID_MODEL_CONFIG={shlex.quote(config_arg)}",
                        f"VEID_MODEL_CARD={shlex.quote(model_card_arg)}",
                        f"VEID_MODEL_PROVENANCE={shlex.quote(shell_path_arg(provenance_path))}",
                        f"{shlex.quote(script_path)}",
                        "1.0.0",
                    ]
                ),
            ],
            cwd=repo_root(),
            check=False,
            capture_output=True,
            text=True,
            env=os.environ.copy(),
        )
        assert completed.returncode == 0, completed.stderr or completed.stdout

    assert (output_a / "release_manifest.json").read_text(encoding="utf-8") == (
        output_b / "release_manifest.json"
    ).read_text(encoding="utf-8")
    assert (output_a / "bundle_checksums.txt").read_text(encoding="utf-8") == (
        output_b / "bundle_checksums.txt"
    ).read_text(encoding="utf-8")
    assert (output_a / "bundle_checksums.txt.sha256").read_text(encoding="utf-8") == (
        output_b / "bundle_checksums.txt.sha256"
    ).read_text(encoding="utf-8")
    assert (output_a / "source_provenance.json").read_text(encoding="utf-8") == (
        output_b / "source_provenance.json"
    ).read_text(encoding="utf-8")
    assert (output_a / "model_signature.json").read_text(encoding="utf-8") == (
        output_b / "model_signature.json"
    ).read_text(encoding="utf-8")
    assert (output_a / "pipeline_version.json").read_text(encoding="utf-8") == (
        output_b / "pipeline_version.json"
    ).read_text(encoding="utf-8")


@pytest.mark.skipif(BASH is None, reason="Git Bash is required for build script verification")
def test_build_script_rejects_partial_bundle(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path)
    os.remove(bundle_dir / "model_frozen.pb")

    completed = subprocess.run(
        [
            BASH,
            "-lc",
            " ".join(
                    [
                        "VEID_SKIP_DOCKER_BUILD=true",
                        f"VEID_MODEL_BUNDLE_DIR={shlex.quote(shell_path_arg(bundle_dir))}",
                        f"VEID_OUTPUT_DIR={shlex.quote(shell_path_arg(tmp_path / 'out'))}",
                        f"VEID_MODEL_CONFIG={shlex.quote(config_path().relative_to(repo_root()).as_posix())}",
                        f"VEID_MODEL_CARD={shlex.quote(model_card_path().relative_to(repo_root()).as_posix())}",
                        f"VEID_MODEL_PROVENANCE={shlex.quote(shell_path_arg(provenance_path))}",
                    "./_build/build-veid-pipeline.sh",
                    "1.0.0",
                ]
            ),
        ],
        cwd=repo_root(),
        check=False,
        capture_output=True,
        text=True,
        env=os.environ.copy(),
    )

    assert completed.returncode != 0
    assert "missing" in (completed.stderr + completed.stdout).lower()


@pytest.mark.skipif(BASH is None, reason="Git Bash is required for build script verification")
def test_build_script_rejects_placeholder_bundle(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    provenance_path = create_model_provenance(tmp_path)
    (bundle_dir / "MODEL_HASH.txt").write_text(
        "SHA256=pending\nVERSION=v1.0.0\n",
        encoding="utf-8",
    )

    completed = subprocess.run(
        [
            BASH,
            "-lc",
            " ".join(
                [
                    "VEID_SKIP_DOCKER_BUILD=true",
                    f"VEID_MODEL_BUNDLE_DIR={shlex.quote(shell_path_arg(bundle_dir))}",
                    f"VEID_OUTPUT_DIR={shlex.quote(shell_path_arg(tmp_path / 'out'))}",
                    f"VEID_MODEL_CONFIG={shlex.quote(config_path().relative_to(repo_root()).as_posix())}",
                    f"VEID_MODEL_CARD={shlex.quote(model_card_path().relative_to(repo_root()).as_posix())}",
                    f"VEID_MODEL_PROVENANCE={shlex.quote(shell_path_arg(provenance_path))}",
                    "./_build/build-veid-pipeline.sh",
                    "1.0.0",
                ]
            ),
        ],
        cwd=repo_root(),
        check=False,
        capture_output=True,
        text=True,
        env=os.environ.copy(),
    )

    assert completed.returncode != 0
    assert "placeholder content detected" in (completed.stderr + completed.stdout).lower()


@pytest.mark.skipif(BASH is None, reason="Git Bash is required for build script verification")
@pytest.mark.parametrize("profile", ["Production", "prod"])
def test_build_script_rejects_profile_alias_before_cleanup(tmp_path: Path, profile: str):
    bundle_dir = create_bundle(tmp_path / "bundle-root")
    provenance_path = create_model_provenance(tmp_path)
    output_dir = tmp_path / "output"
    output_dir.mkdir()
    sentinel = output_dir / "keep.txt"
    sentinel.write_text("keep", encoding="utf-8")

    completed = subprocess.run(
        [
            BASH,
            "-lc",
            " ".join(
                [
                    "VEID_SKIP_DOCKER_BUILD=true",
                    f"VEID_BUILD_PROFILE={shlex.quote(profile)}",
                    f"VEID_MODEL_BUNDLE_DIR={shlex.quote(shell_path_arg(bundle_dir))}",
                    f"VEID_OUTPUT_DIR={shlex.quote(shell_path_arg(output_dir))}",
                    f"VEID_MODEL_PROVENANCE={shlex.quote(shell_path_arg(provenance_path))}",
                    "./_build/build-veid-pipeline.sh",
                    "1.0.0",
                ]
            ),
        ],
        cwd=repo_root(),
        check=False,
        capture_output=True,
        text=True,
        env=os.environ.copy(),
    )

    assert completed.returncode != 0
    assert "profile must be production or fixture_only" in (completed.stderr + completed.stdout).lower()
    assert sentinel.read_text(encoding="utf-8") == "keep"


@pytest.mark.skipif(BASH is None, reason="Git Bash is required for build script verification")
def test_build_script_rejects_provenance_beneath_output_before_cleanup(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path / "bundle-root")
    output_dir = tmp_path / "output"
    output_dir.mkdir()
    provenance_path = create_model_provenance(output_dir)
    sentinel = output_dir / "keep.txt"
    sentinel.write_text("keep", encoding="utf-8")

    completed = subprocess.run(
        [
            BASH,
            "-lc",
            " ".join(
                [
                    "VEID_SKIP_DOCKER_BUILD=true",
                    f"VEID_MODEL_BUNDLE_DIR={shlex.quote(shell_path_arg(bundle_dir))}",
                    f"VEID_OUTPUT_DIR={shlex.quote(shell_path_arg(output_dir))}",
                    f"VEID_MODEL_PROVENANCE={shlex.quote(shell_path_arg(provenance_path))}",
                    "./_build/build-veid-pipeline.sh",
                    "1.0.0",
                ]
            ),
        ],
        cwd=repo_root(),
        check=False,
        capture_output=True,
        text=True,
        env=os.environ.copy(),
    )

    assert completed.returncode != 0
    assert "model provenance must not be beneath output directory" in (
        completed.stderr + completed.stdout
    ).lower()
    assert sentinel.read_text(encoding="utf-8") == "keep"


@pytest.mark.skipif(BASH is None, reason="Git Bash is required for build script verification")
@pytest.mark.parametrize("provenance_state", ["missing", "dependency_blocked"])
def test_build_script_provenance_failure_preserves_output_sentinel(
    tmp_path: Path, provenance_state: str
):
    bundle_dir = create_bundle(tmp_path / "bundle-root")
    output_dir = tmp_path / "output"
    output_dir.mkdir()
    sentinel = output_dir / "keep.txt"
    sentinel.write_text("keep", encoding="utf-8")
    if provenance_state == "missing":
        provenance_path = tmp_path / "missing-provenance.json"
    else:
        provenance_path = create_model_provenance(tmp_path, status="dependency_blocked")

    completed = subprocess.run(
        [
            BASH,
            "-lc",
            " ".join(
                [
                    "VEID_SKIP_DOCKER_BUILD=true",
                    f"VEID_MODEL_BUNDLE_DIR={shlex.quote(shell_path_arg(bundle_dir))}",
                    f"VEID_OUTPUT_DIR={shlex.quote(shell_path_arg(output_dir))}",
                    f"VEID_MODEL_PROVENANCE={shlex.quote(shell_path_arg(provenance_path))}",
                    "./_build/build-veid-pipeline.sh",
                    "1.0.0",
                ]
            ),
        ],
        cwd=repo_root(),
        check=False,
        capture_output=True,
        text=True,
        env=os.environ.copy(),
    )

    assert completed.returncode != 0
    assert sentinel.read_text(encoding="utf-8") == "keep"


def test_write_release_manifest_writes_stable_json(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path / "bundle-root")
    provenance_path = create_model_provenance(tmp_path)
    output_path = tmp_path / "output" / "release_manifest.json"

    write_release_manifest(
        output_path,
        bundle_dir=bundle_dir,
        model_name="trust_score",
        version="v1.0.0",
        profile="production",
        config_path=config_path(),
        model_card_path=model_card_path(),
        provenance_path=provenance_path,
        source_revision=source_revision(),
    )

    written = json.loads(output_path.read_text(encoding="utf-8"))
    assert written["model"]["version"] == "v1.0.0"
    assert written["model"]["runtime_hash"]
