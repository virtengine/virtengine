"""Scoped smoke/release orchestration tests for VirtEngine CI."""

from __future__ import annotations

import os
import shlex
import stat
import subprocess
from pathlib import Path
from textwrap import dedent


REPO_ROOT = Path(__file__).resolve().parents[2]
SMOKE_SCRIPT = REPO_ROOT / "scripts" / "smoke-test.sh"
CI_WORKFLOW = REPO_ROOT / ".github" / "workflows" / "ci.yaml"
RELEASE_WORKFLOW = REPO_ROOT / ".github" / "workflows" / "release.yaml"
SMOKE_WORKFLOW = REPO_ROOT / ".github" / "workflows" / "smoke-test.yaml"


def _write_executable(path: Path, content: str) -> None:
    path.write_text(dedent(content), encoding="utf-8", newline="\n")
    path.chmod(path.stat().st_mode | stat.S_IEXEC)


def _to_bash_path(path: Path) -> str:
    resolved = path.resolve()
    if not resolved.drive:
        return resolved.as_posix()

    drive = resolved.drive.rstrip(":").lower()
    tail = resolved.as_posix().split(":", 1)[1]
    return f"/mnt/{drive}{tail}"


def _prepare_fake_tooling(tmp_path: Path, *, heights: tuple[int, int], require_bootstrap: bool = False) -> dict[str, str]:
    bin_dir = tmp_path / "bin"
    bin_dir.mkdir()
    bootstrap_flag = tmp_path / "kubeconfig.ready"
    bootstrap_flag_bash = _to_bash_path(bootstrap_flag)
    curl_counter = tmp_path / "curl.counter"
    curl_counter.write_text("0", encoding="utf-8")
    curl_counter_bash = _to_bash_path(curl_counter)

    _write_executable(
        bin_dir / "kubectl",
        f"""
        #!/usr/bin/env bash
        set -euo pipefail
        if [[ "$*" == *"cluster-info"* ]]; then
          if [[ "{'1' if require_bootstrap else '0'}" == "1" ]] && [[ ! -f "{bootstrap_flag_bash}" ]]; then
            exit 1
          fi
          echo "cluster-info"
          exit 0
        fi
        if [[ "$*" == *"get nodes"* ]]; then
          printf '%s\\n' "node-a Ready" "node-b Ready"
          exit 0
        fi
        if [[ "$*" == *"-n virtengine"* ]] && [[ "$*" == *"-l app=virtengine"* ]]; then
          printf '%s\\n' "virtengine-0 1/1 Running 0 1m"
          exit 0
        fi
        if [[ "$*" == *"-n cockroachdb"* ]] && [[ "$*" == *"-l app.kubernetes.io/name=cockroachdb"* ]]; then
          printf '%s\\n' "cockroachdb-0 1/1 Running 0 1m"
          exit 0
        fi
        if [[ "$*" == *"-n monitoring"* ]] && [[ "$*" == *"-l app.kubernetes.io/name=prometheus"* ]]; then
          printf '%s\\n' "prometheus-0 1/1 Running 0 1m"
          exit 0
        fi
        echo "unexpected kubectl invocation: $*" >&2
        exit 1
        """,
    )

    _write_executable(
        bin_dir / "aws",
        f"""
        #!/usr/bin/env bash
        set -euo pipefail
        if [[ "$*" == *"eks update-kubeconfig"* ]]; then
          mkdir -p "$(dirname "{bootstrap_flag_bash}")"
          touch "{bootstrap_flag_bash}"
          exit 0
        fi
        echo "unexpected aws invocation: $*" >&2
        exit 1
        """,
    )

    _write_executable(
        bin_dir / "curl",
        f"""
        #!/usr/bin/env bash
        set -euo pipefail
        counter_file="{curl_counter_bash}"
        count="$(cat "$counter_file")"
        if [[ "$*" == *"/health"* ]]; then
          printf '{{}}'
          exit 0
        fi
        if [[ "$*" == *"/status"* ]]; then
          if [[ "$count" == "0" ]]; then
            printf '{{"result":{{"sync_info":{{"latest_block_height":"{heights[0]}"}}}}}}'
            echo "1" > "$counter_file"
          else
            printf '{{"result":{{"sync_info":{{"latest_block_height":"{heights[1]}"}}}}}}'
          fi
          exit 0
        fi
        echo "unexpected curl invocation: $*" >&2
        exit 1
        """,
    )

    env = os.environ.copy()
    env["PATH"] = f"{bin_dir}{os.pathsep}{env['PATH']}"
    env["VE_TEST_BIN_DIR"] = _to_bash_path(bin_dir)
    env["VE_BLOCK_PROGRESS_WAIT_SECONDS"] = "0"
    return env


def _run_smoke(env: dict[str, str]) -> subprocess.CompletedProcess[str]:
    command = (
        f'export PATH="{env["VE_TEST_BIN_DIR"]}:$PATH"; '
        f"exec bash {shlex.quote(str(SMOKE_SCRIPT.relative_to(REPO_ROOT).as_posix()))} us-east-1"
    )
    return subprocess.run(
        ["bash", "-c", command],
        cwd=REPO_ROOT,
        env=env,
        text=True,
        capture_output=True,
        check=False,
    )


def test_release_workflow_requires_prepublish_gates() -> None:
    release_text = RELEASE_WORKFLOW.read_text(encoding="utf-8")
    ci_text = CI_WORKFLOW.read_text(encoding="utf-8")

    for required_workflow in (
        "compatibility.yaml",
        "smoke-test.yaml",
        "staging-e2e.yaml",
        "veid-e2e.yaml",
        "ml-determinism.yaml",
        "veid-conformance.yaml",
    ):
        assert f"uses: ./.github/workflows/{required_workflow}" in release_text

    assert "Validate Release Context" in release_text
    assert "refs/tags/${release_tag}" in release_text
    assert "draft: true" in ci_text
    assert "uses: ./.github/workflows/smoke-test.yaml" in ci_text
    assert "uses: ./.github/workflows/staging-e2e.yaml" in ci_text
    assert "uses: ./.github/workflows/compatibility.yaml" in ci_text


def test_smoke_workflow_is_reusable_and_uploads_artifacts() -> None:
    smoke_text = SMOKE_WORKFLOW.read_text(encoding="utf-8")

    assert "workflow_call:" in smoke_text
    assert "scripts/ci/post-deploy-smoke-test.sh" in smoke_text
    assert "scripts/ci/run-portal-smoke.sh" in smoke_text
    assert "Upload smoke artifacts" in smoke_text
    assert "smoke-artifacts-${{ env.TARGET_ENV }}" in smoke_text


def test_regional_smoke_script_passes_with_cluster_and_height_progression(tmp_path: Path) -> None:
    env = _prepare_fake_tooling(tmp_path, heights=(120, 121))
    result = _run_smoke(env)

    assert result.returncode == 0, result.stderr + result.stdout
    assert "RESULT: PASSED" in result.stdout
    assert "Block height advanced (120 -> 121)" in result.stdout


def test_regional_smoke_script_bootstraps_kubeconfig_when_context_missing(tmp_path: Path) -> None:
    env = _prepare_fake_tooling(tmp_path, heights=(400, 402), require_bootstrap=True)
    result = _run_smoke(env)

    assert result.returncode == 0, result.stderr + result.stdout
    assert "Bootstrapping kubeconfig" in result.stdout
    assert "RESULT: PASSED" in result.stdout


def test_regional_smoke_script_fails_when_block_height_stalls(tmp_path: Path) -> None:
    env = _prepare_fake_tooling(tmp_path, heights=(220, 220))
    result = _run_smoke(env)

    assert result.returncode != 0
    assert "Block height did not advance (220 -> 220)" in result.stdout
