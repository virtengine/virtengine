from __future__ import annotations

import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SCRIPT_PATH = REPO_ROOT / ".github" / "scripts" / "validate_security_policies.py"


class SecurityPolicyE2ETests(unittest.TestCase):
    def test_validator_cli_passes_on_repo_surface(self) -> None:
        result = subprocess.run(
            [sys.executable, str(SCRIPT_PATH)],
            cwd=REPO_ROOT,
            capture_output=True,
            text=True,
            check=False,
        )

        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertIn("PASS .github/workflows/security.yaml", result.stdout)
        self.assertIn("PASS .vulnerability-allowlist.yaml", result.stdout)

    def test_validator_cli_fails_for_invalid_workflow_fixture(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            target = Path(temp_dir) / "security.yaml"
            target.write_text(
                """
name: Invalid
defaults:
  run:
    shell: bash
permissions:
  contents: read
jobs:
  policy-validation:
    runs-on: ubuntu-latest
    steps:
      - run: python .github/scripts/validate_security_policies.py
      - run: python -m unittest discover -s .github/tests -p "test_*policy*.py"
      - run: python .github/scripts/validate_inference_deployment_policy.py
      - run: go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12 .github/workflows/security.yaml
  codeql-analysis:
    runs-on: ubuntu-latest
  go-vuln-scan:
    runs-on: ubuntu-latest
  python-vuln-scan:
    runs-on: ubuntu-latest
  node-vuln-scan:
    runs-on: ubuntu-latest
  container-scan:
    runs-on: ubuntu-latest
  sbom-generation:
    runs-on: ubuntu-latest
  secret-scan:
    runs-on: ubuntu-latest
  gosec-scan:
    runs-on: ubuntu-latest
  security-summary:
    runs-on: ubuntu-latest
    steps:
      - run: go install golang.org/x/vuln/cmd/govulncheck@latest
      - run: scripts/supply-chain/generate-sbom.sh --format all
      - run: gitleaks detect
      - run: trivy image
      - run: pip-audit==2.10.0
      - run: pnpm audit --prod --audit-level=high
      - run: npm --prefix sdk/ts audit --audit-level=high
      - run: github.com/securego/gosec/v2/cmd/gosec@${{ env.GOSEC_VERSION }}
""".strip(),
                encoding="utf-8",
            )

            result = subprocess.run(
                [sys.executable, str(SCRIPT_PATH), "--files", str(target)],
                cwd=REPO_ROOT,
                capture_output=True,
                text=True,
                check=False,
            )

            self.assertNotEqual(result.returncode, 0, result.stdout + result.stderr)
            self.assertIn("FAIL", result.stdout)
            self.assertIn("@latest", result.stdout)


if __name__ == "__main__":
    unittest.main()
