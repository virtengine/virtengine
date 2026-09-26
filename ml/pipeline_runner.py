"""VirtEngine VEID pipeline runner (VE-219).

Container entry point for the deterministic identity-verification pipeline
image (``_build/Dockerfile.veid-pipeline``). Invoked as::

    python3 -m ml.pipeline_runner --mode=sidecar --port=50051 --deterministic=true
    python3 -m ml.pipeline_runner --mode=check

Design rules for this module:

* Importing it MUST be side-effect free (no weight loading, no network).
  The image HEALTHCHECK is ``python3 -c "import ml.pipeline_runner"``.
* It MUST fail closed: missing model bundle or non-deterministic
  environment exits non-zero with an actionable message on stderr. It MUST
  never emit a fake score, hash, or health claim.
* It is a code-only image entry point for now: model weights are produced
  by ``ml/training`` (see ``models/trust_score``) and validated by
  ``_build/build-veid-pipeline.sh``, then mounted at ``/app/models`` at
  deploy time. ``--mode=sidecar`` validates startup preconditions; the
  TensorFlow ComputeScore serving path is ML-owner work and exits non-zero
  until it lands (see module docstring reference below).

Standard library only, so the module imports even in the distroless final
stage and on hosts without the pinned ML dependencies installed.
"""

from __future__ import annotations

import argparse
import hashlib
import os
import random
import sys

PIPELINE_VERSION = "1.0.0"

#: Environment contract enforced for deterministic execution. Mirrors the
#: ENV block in _build/Dockerfile.veid-pipeline (Stage 1).
REQUIRED_DETERMINISM_ENV = {
    "TF_DETERMINISTIC_OPS": "1",
    "TF_CUDNN_DETERMINISTIC": "1",
    "TF_ENABLE_ONEDNN_OPTS": "0",
    "PYTHONHASHSEED": "42",
    "OMP_NUM_THREADS": "1",
    "CUDA_VISIBLE_DEVICES": "-1",
}

#: Artifacts that identify a validated model bundle. Mirrors the
#: ``require_bundle`` checks in _build/build-veid-pipeline.sh.
REQUIRED_BUNDLE_FILES = (
    "manifest.json",
    "export_metadata.json",
    "MODEL_HASH.txt",
    "model_frozen.pb",
)

DEFAULT_MODEL_DIR = "/app/models"
DEFAULT_PORT = 50051


def check_determinism_env() -> list[str]:
    """Return a list of human-readable determinism violations (empty = OK)."""
    problems = []
    for key, want in sorted(REQUIRED_DETERMINISM_ENV.items()):
        got = os.environ.get(key)
        if got != want:
            problems.append(f"{key} must be {want!r} (got {got!r})")
    return problems


def apply_determinism_seeds() -> None:
    """Seed stdlib RNGs. Numpy/torch/TF seeds are applied lazily by callers."""
    random.seed(42)


def find_model_bundle(model_dir: str) -> tuple[bool, list[str]]:
    """Check for a validated trust_score model bundle under *model_dir*.

    Returns (ok, missing): ok is True only when every file in
    REQUIRED_BUNDLE_FILES exists somewhere under model_dir.
    """
    missing = []
    for name in REQUIRED_BUNDLE_FILES:
        found = False
        for root, _dirs, files in os.walk(model_dir):
            if name in files:
                found = True
                break
        if not found:
            missing.append(name)
    return (not missing, missing)


def bundle_fingerprint(model_dir: str) -> str:
    """Best-effort sha256 over bundle file paths (presence proof, not a hash claim)."""
    digest = hashlib.sha256()
    for root, _dirs, files in os.walk(model_dir):
        for name in sorted(files):
            digest.update(os.path.join(root, name).encode("utf-8"))
    return digest.hexdigest()


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="ml.pipeline_runner",
        description="VirtEngine VEID deterministic pipeline runner (VE-219).",
    )
    parser.add_argument(
        "--mode",
        choices=("sidecar", "check"),
        default="sidecar",
        help="'sidecar': serve inference preconditions then start serving; "
        "'check': validate environment + bundle and exit (0 = ready).",
    )
    parser.add_argument("--port", type=int, default=DEFAULT_PORT)
    parser.add_argument(
        "--deterministic",
        default="true",
        choices=("true", "false"),
        help="Refuse to start unless the determinism contract holds.",
    )
    parser.add_argument(
        "--model-dir",
        default=os.environ.get("VEID_MODEL_DIR", DEFAULT_MODEL_DIR),
        help="Directory the validated model bundle is mounted at.",
    )
    parser.add_argument("--version", action="version", version=PIPELINE_VERSION)
    return parser


def run_check(model_dir: str, enforce_determinism: bool) -> int:
    problems = check_determinism_env() if enforce_determinism else []
    if problems:
        sys.stderr.write("non-deterministic environment:\n")
        for problem in problems:
            sys.stderr.write(f"  - {problem}\n")
        return 2
    ok, missing = find_model_bundle(model_dir)
    if not ok:
        sys.stderr.write(
            f"model bundle not present under {model_dir} "
            f"(missing: {', '.join(missing)}). "
            "Produce it with ml/training + _build/build-veid-pipeline.sh "
            "and mount it at --model-dir.\n"
        )
        return 2
    apply_determinism_seeds()
    sys.stdout.write(
        f"READY pipeline={PIPELINE_VERSION} model_dir={model_dir} "
        f"bundle={bundle_fingerprint(model_dir)[:12]}\n"
    )
    return 0


def run_sidecar(port: int, model_dir: str, enforce_determinism: bool) -> int:
    """Validate startup preconditions, then serve.

    The TensorFlow ComputeScore serving path is not implemented yet: this
    refuses to start (exit 2) rather than serve unvalidated responses, which
    would break validator consensus. ML owners: implement the serving path
    against pkg/inference/proto/inference.proto and relax this gate.
    """
    if enforce_determinism:
        problems = check_determinism_env()
        if problems:
            sys.stderr.write("refusing to start in non-deterministic environment:\n")
            for problem in problems:
                sys.stderr.write(f"  - {problem}\n")
            return 2
    ok, missing = find_model_bundle(model_dir)
    if not ok:
        sys.stderr.write(
            f"refusing to start: model bundle not present under {model_dir} "
            f"(missing: {', '.join(missing)}).\n"
        )
        return 2
    apply_determinism_seeds()
    sys.stderr.write(
        "refusing to start: TensorFlow ComputeScore serving path is not "
        "implemented in this image (consensus safety: no unvalidated "
        "responses). Pipeline code ships for scanning; serve blocked until "
        "ML owners implement InferenceService.ComputeScore against "
        "pkg/inference/proto/inference.proto.\n"
    )
    return 2


def main(argv: list[str] | None = None) -> int:
    args = build_parser().parse_args(argv)
    enforce = args.deterministic == "true"
    if args.mode == "check":
        return run_check(args.model_dir, enforce)
    return run_sidecar(args.port, args.model_dir, enforce)


if __name__ == "__main__":
    raise SystemExit(main())
