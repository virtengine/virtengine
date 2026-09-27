from __future__ import annotations

import importlib.util
import sys
import tempfile
import unittest
from datetime import date, timedelta
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SCRIPT_PATH = REPO_ROOT / ".github" / "scripts" / "validate_security_policies.py"


def load_validator_module():
    spec = importlib.util.spec_from_file_location("validate_security_policies", SCRIPT_PATH)
    module = importlib.util.module_from_spec(spec)
    assert spec is not None and spec.loader is not None
    sys.modules[spec.name] = module
    spec.loader.exec_module(module)
    return module


class SecurityPolicyValidatorUnitTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls.validator = load_validator_module()

    def write_file(self, relative_path: str, content: str) -> Path:
        temp_dir = tempfile.TemporaryDirectory()
        self.addCleanup(temp_dir.cleanup)
        base = Path(temp_dir.name)
        target = base / relative_path
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_text(content, encoding="utf-8")
        return target

    def test_workflow_rejects_latest_continue_on_error_and_extra_secret(self) -> None:
        workflow = self.write_file(
            "security.yaml",
            """
name: Test
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
    continue-on-error: true
    steps:
      - run: go install golang.org/x/vuln/cmd/govulncheck@latest
      - run: echo ${{ secrets.SLACK_WEBHOOK }}
      - run: scripts/supply-chain/generate-sbom.sh --format all
      - run: gitleaks detect
      - run: trivy image
      - run: pnpm audit --prod --audit-level=high
      - run: npm --prefix sdk/ts audit --audit-level=high
      - run: pip-audit==2.10.0
      - run: github.com/securego/gosec/v2/cmd/gosec@${{ env.GOSEC_VERSION }}
""".strip(),
        )

        errors = self.validator.validate_workflow(workflow)

        self.assertTrue(any("@latest" in error for error in errors))
        self.assertTrue(any("continue-on-error: true" in error for error in errors))
        self.assertTrue(any("secrets.SLACK_WEBHOOK" in error for error in errors))

    def test_gitleaks_rejects_broad_allowlists(self) -> None:
        config = self.write_file(
            ".gitleaks.toml",
            """
title = "Test"
[extend]
useDefault = true
[allowlist]
paths = [
  '''.*\\.md$''',
  '''(^|/)vendor/''',
]
""".strip(),
        )

        errors = self.validator.validate_gitleaks(config)

        self.assertTrue(any("forbidden broad path exemption" in error for error in errors))

    def test_allowlist_rejects_expired_or_placeholder_entries(self) -> None:
        allowlist = self.write_file(
            ".vulnerability-allowlist.yaml",
            """
version: 2
policy:
  block_on: [CRITICAL, HIGH]
  max_allowlist_age_days: 30
  require_issue_reference: true
  require_compensating_controls: true
  active_exception_count: 1
exceptions:
  go:
    - id: CVE-2026-0001
      package: example.com/pkg
      reason: Example entry
      reviewed_by: security-team
      reviewed_date: "2026-01-01"
      expires: "2026-02-15"
      references:
        - https://example.test/advisory
      compensating_controls:
        - blocked by firewall
  python: []
  npm: []
  containers: []
audit_log:
  - date: "2024-01-01"
    actor: security-team
    action: created
    note: Initial vulnerability allowlist created
""".strip(),
        )

        errors = self.validator.validate_allowlist(allowlist)

        self.assertTrue(any("stale placeholder text" in error for error in errors))
        self.assertTrue(any("expired" in error or "exceeds the maximum allowlist age" in error for error in errors))

    def allowlist_with_exception(self, reviewed: date, expires: date) -> Path:
        """Build an allowlist whose only entry spans exactly reviewed..expires.

        Dates are computed relative to today by the caller so the suite cannot rot, and the
        placeholder-triggering literals are deliberately avoided so the expiry and age rules
        are the only rules that can fire.
        """
        return self.write_file(
            ".vulnerability-allowlist.yaml",
            f"""
version: 2
policy:
  block_on: [CRITICAL, HIGH]
  max_allowlist_age_days: 30
  require_issue_reference: true
  require_compensating_controls: true
  active_exception_count: 1
exceptions:
  go:
    - id: GO-2026-4740
      package: github.com/shamaton/msgpack/v2
      reason: DoS with no patched release; decode path unreachable from shipped binaries
      reviewed_by: secops
      reviewed_date: "{reviewed.isoformat()}"
      expires: "{expires.isoformat()}"
      references:
        - https://pkg.go.dev/vuln/GO-2026-4740
      compensating_controls:
        - no VirtEngine code imports the affected package
  python: []
  npm: []
  containers: []
audit_log:
  - date: "{reviewed.isoformat()}"
    actor: secops
    action: reviewed
    note: expiry coverage fixture
""".strip(),
        )

    def test_expiry_branch_fires_in_isolation(self) -> None:
        """A lapsed entry must be reported even when the age rule is satisfied.

        Regression guard: the pre-existing expired-entry test also carried placeholder text and
        a span over `max_allowlist_age_days`, so it passed via the *age* rule alone — deleting
        the `expires < today` branch outright left the suite fully green.
        """
        today = date.today()
        allowlist = self.allowlist_with_exception(reviewed=today - timedelta(days=16), expires=today - timedelta(days=1))

        errors = self.validator.validate_allowlist(allowlist)

        self.assertTrue(any("is expired" in error for error in errors), errors)
        self.assertFalse(any("exceeds the maximum allowlist age" in error for error in errors), errors)

    def test_age_branch_fires_in_isolation(self) -> None:
        """An over-long time-box must be reported even when the entry is not yet expired."""
        today = date.today()
        allowlist = self.allowlist_with_exception(reviewed=today, expires=today + timedelta(days=31))

        errors = self.validator.validate_allowlist(allowlist)

        self.assertTrue(any("exceeds the maximum allowlist age" in error for error in errors), errors)
        self.assertFalse(any("is expired" in error for error in errors), errors)

    def test_exception_at_exactly_max_age_is_accepted(self) -> None:
        """The 30-day boundary itself must stay valid, so reviews are not forced shorter."""
        today = date.today()
        allowlist = self.allowlist_with_exception(reviewed=today, expires=today + timedelta(days=30))

        errors = self.validator.validate_allowlist(allowlist)

        self.assertFalse(any("is expired" in error for error in errors), errors)
        self.assertFalse(any("exceeds the maximum allowlist age" in error for error in errors), errors)

    def test_doc_rejects_stale_claims(self) -> None:
        document = self.write_file(
            "SUPPLY_CHAIN_SECURITY.md",
            """
# Supply Chain

SLSA Level 3
In Progress
_Last updated: 2024_
""".strip(),
        )

        errors = self.validator.validate_supply_chain_doc(document)

        self.assertGreaterEqual(len(errors), 3)

    def license_compliance_workflow(self, pip_licenses_step: str) -> Path:
        """A license-compliance.yaml that satisfies every requirement except the pip-licenses pin."""
        return self.write_file(
            "license-compliance.yaml",
            """
name: License Compliance
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
      - run: python -m unittest discover -s .github/tests -p "test_security_policy*.py"
      - run: go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12 .github/workflows/license-compliance.yaml
  go-licenses:
    runs-on: ubuntu-latest
    steps:
      - run: go run github.com/google/go-licenses@${{ env.GO_LICENSES_VERSION }} csv ./... > licenses.csv
  javascript-licenses:
    runs-on: ubuntu-latest
    steps:
      - run: npx license-checker-rseidelsohn@${{ env.LICENSE_CHECKER_VERSION }} --json
  python-licenses:
    runs-on: ubuntu-latest
    steps:
      - @@PIP_LICENSES_STEP@@
  spdx-sbom:
    runs-on: ubuntu-latest
    steps:
      - run: scripts/supply-chain/generate-sbom.sh --format spdx
  license-summary:
    runs-on: ubuntu-latest
""".replace("@@PIP_LICENSES_STEP@@", pip_licenses_step).strip(),
        )

    def test_pip_licenses_pin_accepts_expression_and_shell_forms(self) -> None:
        """#888 installs the pinned scanner inside a venv, where the shell form of the same
        workflow-level pin is correct. The policy guards the *pin*, not the spelling, so both
        forms must validate - the earlier snippet-only check produced a false red on develop."""
        workflows = {
            "expression": self.license_compliance_workflow(
                'run: python -m pip install --quiet "pip-licenses==${{ env.PIP_LICENSES_VERSION }}"'
            ),
            "shell": self.license_compliance_workflow(
                'run: python -m pip install --quiet "pip-licenses==${PIP_LICENSES_VERSION}"'
            ),
        }

        for form, workflow in workflows.items():
            with self.subTest(form=form):
                errors = self.validator.validate_workflow(workflow)
                self.assertFalse([error for error in errors if "pip-licenses" in error], errors)

    def test_unpinned_pip_licenses_install_is_rejected(self) -> None:
        """Accepting both spellings must not accept an unpinned scanner."""
        workflow = self.license_compliance_workflow("run: python -m pip install --quiet pip-licenses")

        errors = self.validator.validate_workflow(workflow)

        self.assertTrue(any("pip-licenses" in error for error in errors), errors)


if __name__ == "__main__":
    unittest.main()
