from __future__ import annotations

import importlib.util
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SCRIPT_PATH = REPO_ROOT / ".github" / "scripts" / "validate_inference_deployment_policy.py"


def load_validator_module():
    spec = importlib.util.spec_from_file_location("validate_inference_deployment_policy", SCRIPT_PATH)
    module = importlib.util.module_from_spec(spec)
    assert spec is not None and spec.loader is not None
    sys.modules[spec.name] = module
    spec.loader.exec_module(module)
    return module


class InferenceDeploymentPolicyTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls.validator = load_validator_module()

    def write_file(self, root: Path, relative_path: str, content: str) -> Path:
        target = root / relative_path
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_text(content, encoding="utf-8")
        return target

    def test_validator_rejects_stub_flag_in_deployment_surface(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            root = Path(temp_dir)
            deploy_root = root / "deploy"
            self.write_file(
                deploy_root,
                "kubernetes/base/inference-sidecar.yaml",
                """
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
        - name: inference-sidecar
          args:
            - --grpc-addr=:50051
            - --allow-fallback-to-stub
""".strip(),
            )

            errors = self.validator.validate_surfaces(
                [self.validator.PolicySurface(deploy_root, recursive=True)]
            )

            self.assertTrue(any("--allow-fallback-to-stub" in error for error in errors))

    def test_validator_rejects_local_stub_marker_in_workflow(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            root = Path(temp_dir)
            workflows_root = root / ".github" / "workflows"
            self.write_file(
                workflows_root,
                "deploy.yaml",
                """
name: Deploy
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - run: echo "local_stub accepted"
""".strip(),
            )

            errors = self.validator.validate_surfaces(
                [self.validator.PolicySurface(workflows_root, recursive=True)]
            )

            self.assertTrue(any("local_stub" in error for error in errors))

    def test_validator_ignores_dev_only_docs_outside_production_surfaces(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            root = Path(temp_dir)
            internal_docs = root / "_docs"
            self.write_file(
                internal_docs,
                "inference-fallback-behavior.md",
                "Use --allow-fallback-to-stub only for local testing; local_stub indicates fallback mode.",
            )

            errors = self.validator.validate_surfaces(
                [self.validator.PolicySurface(root / "deploy", recursive=True)]
            )

            self.assertEqual(errors, [])

    def test_validator_cli_passes_on_repo_surfaces(self) -> None:
        result = subprocess.run(
            [sys.executable, str(SCRIPT_PATH)],
            cwd=REPO_ROOT,
            capture_output=True,
            text=True,
            check=False,
        )

        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertIn("PASS inference deployment policy", result.stdout)

    def test_validator_cli_fails_for_invalid_fixture(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            root = Path(temp_dir)
            self.write_file(
                root,
                "docs/inference-sidecar-deployment.md",
                "Production example: --allow-fallback-to-stub",
            )

            result = subprocess.run(
                [sys.executable, str(SCRIPT_PATH), "--paths", str(root / "docs")],
                cwd=REPO_ROOT,
                capture_output=True,
                text=True,
                check=False,
            )

            self.assertNotEqual(result.returncode, 0, result.stdout + result.stderr)
            self.assertIn("--allow-fallback-to-stub", result.stdout)


if __name__ == "__main__":
    unittest.main()
