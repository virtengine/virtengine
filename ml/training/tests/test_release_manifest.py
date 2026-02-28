import json
import os
import shlex
import shutil
import subprocess
from pathlib import Path

import pytest

from ml.training.model.release_manifest import (
    ReleaseManifestError,
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
    (bundle_dir / "model_frozen.pb").write_bytes(b"frozen-graph")

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


def sha256_file(path: Path) -> str:
    import hashlib

    hasher = hashlib.sha256()
    hasher.update(path.read_bytes())
    return hasher.hexdigest()


def subprocess_hash_model_dir(model_dir: Path) -> str:
    import hashlib

    hasher = hashlib.sha256()
    for path in sorted(candidate for candidate in model_dir.rglob("*") if candidate.is_file()):
        if path.name == "export_metadata.json":
            continue
        hasher.update(path.read_bytes())
    return hasher.hexdigest()


def repo_root() -> Path:
    return Path(__file__).resolve().parents[3]


def config_path() -> Path:
    return repo_root() / "ml" / "training" / "configs" / "trust_score_v1.yaml"


def model_card_path() -> Path:
    return repo_root() / "models" / "trust_score" / "MODEL_CARD.md"


def test_release_manifest_is_deterministic(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)

    manifest_a = build_release_manifest(
        bundle_dir,
        model_name="trust_score",
        version="v1.0.0",
        profile="production",
        config_path=config_path(),
        model_card_path=model_card_path(),
        source_revision="deadbeef",
    )
    manifest_b = build_release_manifest(
        bundle_dir,
        model_name="trust_score",
        version="v1.0.0",
        profile="production",
        config_path=config_path(),
        model_card_path=model_card_path(),
        source_revision="deadbeef",
    )

    assert manifest_a == manifest_b


def test_release_manifest_uses_portable_source_paths(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    revision = "deadbeefcafebabefeedface1234567890abcdef"

    manifest = build_release_manifest(
        bundle_dir,
        model_name="trust_score",
        version="v1.0.0",
        profile="production",
        config_path=config_path(),
        model_card_path=model_card_path(),
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
    os.remove(bundle_dir / "model_frozen.pb")

    with pytest.raises(ReleaseManifestError, match="required bundle artifact is missing"):
        build_release_manifest(
            bundle_dir,
            model_name="trust_score",
            version="v1.0.0",
            profile="production",
            config_path=config_path(),
            model_card_path=model_card_path(),
            source_revision="deadbeef",
        )


def test_release_manifest_rejects_placeholder_hashes(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
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
            source_revision="deadbeef",
        )


@pytest.mark.skipif(shutil.which("bash") is None, reason="bash is required for build script verification")
def test_build_script_manifest_only_is_repeatable(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    output_a = tmp_path / "out-a"
    output_b = tmp_path / "out-b"
    script_path = "./_build/build-veid-pipeline.sh"

    for output_dir in (output_a, output_b):
        bundle_arg = bundle_dir.relative_to(repo_root()).as_posix()
        output_arg = output_dir.relative_to(repo_root()).as_posix()
        config_arg = config_path().relative_to(repo_root()).as_posix()
        model_card_arg = model_card_path().relative_to(repo_root()).as_posix()
        completed = subprocess.run(
            [
                "bash",
                "-lc",
                " ".join(
                    [
                        "VEID_SKIP_DOCKER_BUILD=true",
                        f"VEID_MODEL_BUNDLE_DIR={shlex.quote(bundle_arg)}",
                        f"VEID_OUTPUT_DIR={shlex.quote(output_arg)}",
                        f"VEID_MODEL_CONFIG={shlex.quote(config_arg)}",
                        f"VEID_MODEL_CARD={shlex.quote(model_card_arg)}",
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


@pytest.mark.skipif(shutil.which("bash") is None, reason="bash is required for build script verification")
def test_build_script_rejects_partial_bundle(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    os.remove(bundle_dir / "model_frozen.pb")

    completed = subprocess.run(
        [
            "bash",
            "-lc",
            " ".join(
                    [
                        "VEID_SKIP_DOCKER_BUILD=true",
                        f"VEID_MODEL_BUNDLE_DIR={shlex.quote(bundle_dir.relative_to(repo_root()).as_posix())}",
                        f"VEID_OUTPUT_DIR={shlex.quote((tmp_path / 'out').relative_to(repo_root()).as_posix())}",
                        f"VEID_MODEL_CONFIG={shlex.quote(config_path().relative_to(repo_root()).as_posix())}",
                        f"VEID_MODEL_CARD={shlex.quote(model_card_path().relative_to(repo_root()).as_posix())}",
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


@pytest.mark.skipif(shutil.which("bash") is None, reason="bash is required for build script verification")
def test_build_script_rejects_placeholder_bundle(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    (bundle_dir / "MODEL_HASH.txt").write_text(
        "SHA256=pending\nVERSION=v1.0.0\n",
        encoding="utf-8",
    )

    completed = subprocess.run(
        [
            "bash",
            "-lc",
            " ".join(
                [
                    "VEID_SKIP_DOCKER_BUILD=true",
                    f"VEID_MODEL_BUNDLE_DIR={shlex.quote(bundle_dir.relative_to(repo_root()).as_posix())}",
                    f"VEID_OUTPUT_DIR={shlex.quote((tmp_path / 'out').relative_to(repo_root()).as_posix())}",
                    f"VEID_MODEL_CONFIG={shlex.quote(config_path().relative_to(repo_root()).as_posix())}",
                    f"VEID_MODEL_CARD={shlex.quote(model_card_path().relative_to(repo_root()).as_posix())}",
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


def test_write_release_manifest_writes_stable_json(tmp_path: Path):
    bundle_dir = create_bundle(tmp_path)
    output_path = tmp_path / "release_manifest.json"

    write_release_manifest(
        output_path,
        bundle_dir=bundle_dir,
        model_name="trust_score",
        version="v1.0.0",
        profile="production",
        config_path=config_path(),
        model_card_path=model_card_path(),
        source_revision="deadbeef",
    )

    written = json.loads(output_path.read_text(encoding="utf-8"))
    assert written["model"]["version"] == "v1.0.0"
    assert written["model"]["runtime_hash"]
