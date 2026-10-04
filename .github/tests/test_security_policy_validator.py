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
        - https://github.com/virtengine/virtengine/issues/872
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

    def allowlist_entry_fixture(
        self,
        references: str,
        compensating_controls: str = "        - no VirtEngine code imports the affected package",
        require_issue_reference: str = "true",
        require_compensating_controls: str = "true",
        reason: str = "DoS with no patched release; decode path unreachable from shipped binaries",
    ) -> Path:
        """A current, in-window allowlist whose only variables are the fields under test.

        Dates are relative to today and the placeholder-triggering literals are avoided, so the
        reference and compensating-control rules are the only ones that can fire.
        """
        today = date.today()
        return self.write_file(
            ".vulnerability-allowlist.yaml",
            f"""
version: 2
policy:
  block_on: [CRITICAL, HIGH]
  max_allowlist_age_days: 30
  require_issue_reference: {require_issue_reference}
  require_compensating_controls: {require_compensating_controls}
  active_exception_count: 1
exceptions:
  go:
    - id: GO-2026-4740
      package: github.com/shamaton/msgpack/v2
      reason: '{reason}'
      reviewed_by: secops
      reviewed_date: "{today.isoformat()}"
      expires: "{(today + timedelta(days=5)).isoformat()}"
      references:
{references}
      compensating_controls:
{compensating_controls}
  python: []
  npm: []
  containers: []
audit_log:
  - date: "{today.isoformat()}"
    actor: secops
    action: review-exception
    note: reference-validation fixture
""".strip(),
        )

    def test_allowlist_rejects_placeholder_issue_reference(self) -> None:
        """`.../issues/NEW` shipped in two live exceptions while the policy advertised that every
        entry was tracked. It is an unresolvable 404, and no rule read the reference list at all,
        so the control was decorative. Regression case for the aws-sdk-go / s3crypto entries."""
        path = self.allowlist_entry_fixture(
            references=(
                "        - https://github.com/virtengine/virtengine/issues/NEW\n"
                "        - https://pkg.go.dev/vuln/GO-2026-4740"
            )
        )

        errors = self.validator.validate_allowlist(path)

        self.assertTrue(any("needs a tracking reference" in error for error in errors), errors)

    def test_allowlist_rejects_advisory_only_references(self) -> None:
        """An advisory page is evidence, not ownership: it is not a tracked issue in this repo."""
        path = self.allowlist_entry_fixture(references="        - https://pkg.go.dev/vuln/GO-2026-4740")

        errors = self.validator.validate_allowlist(path)

        self.assertTrue(any("needs a tracking reference" in error for error in errors), errors)

    def test_allowlist_accepts_a_real_repository_issue_reference(self) -> None:
        """The positive control: the rule must not reject a correctly tracked exception."""
        path = self.allowlist_entry_fixture(
            references="        - https://github.com/virtengine/virtengine/issues/1079"
        )

        errors = self.validator.validate_allowlist(path)

        self.assertFalse([error for error in errors if "reference" in error], errors)

    def test_allowlist_rejects_empty_compensating_controls(self) -> None:
        """policy.require_compensating_controls was likewise declared and never enforced."""
        path = self.allowlist_entry_fixture(
            references="        - https://github.com/virtengine/virtengine/issues/1079",
            compensating_controls="        []",
        )

        errors = self.validator.validate_allowlist(path)

        self.assertTrue(any("compensating control" in error for error in errors), errors)

    def test_allowlist_rejects_a_disabled_issue_reference_policy(self) -> None:
        """Flipping the switch off must fail closed rather than silently dropping the control."""
        path = self.allowlist_entry_fixture(
            references="        - https://github.com/virtengine/virtengine/issues/1079",
            require_issue_reference="false",
        )

        errors = self.validator.validate_allowlist(path)

        self.assertTrue(any("policy.require_issue_reference must be enabled" in error for error in errors), errors)

    def test_allowlist_rejects_a_reference_to_a_missing_repository_file(self) -> None:
        """Three shipped exceptions cited docs/security/... assessment files that did not exist, in
        the same reference list the tracking rule had just made honest. A blob link to a missing
        path is a 404, not evidence, and nothing read the reference list for resolvability."""
        path = self.allowlist_entry_fixture(
            references=(
                "        - https://github.com/virtengine/virtengine/issues/1079\n"
                "        - https://github.com/virtengine/virtengine/blob/develop/docs/security/DOES-NOT-EXIST-ASSESSMENT.md"
            )
        )

        errors = self.validator.validate_allowlist(path)

        self.assertTrue(
            any("docs/security/DOES-NOT-EXIST-ASSESSMENT.md" in error for error in errors), errors
        )

    def test_allowlist_accepts_a_reference_to_an_existing_repository_file(self) -> None:
        """The positive control: a blob link to a file that is really in the tree stays valid."""
        path = self.allowlist_entry_fixture(
            references=(
                "        - https://github.com/virtengine/virtengine/issues/1079\n"
                "        - https://github.com/virtengine/virtengine/blob/develop/SUPPLY_CHAIN_SECURITY.md"
            )
        )

        errors = self.validator.validate_allowlist(path)

        self.assertEqual([error for error in errors if "no such file exists" in error], [], errors)

    def test_allowlist_rejects_a_missing_doc_cited_in_the_reason(self) -> None:
        """The prose channel. Every shipped assessment citation lives in `reason:`, NOT in
        `references:`, and the blob rule walks only the reference list - so repointing the prose
        alone left the entry passing (measured on develop @ 6c5fb13e: exit 0, no error). The
        unresolved class re-opened through the other field, and the reason is the field a new
        entry naturally writes its assessment citation in. The reference list here is entirely
        valid, which is the point: the prose must carry its own obligation."""
        path = self.allowlist_entry_fixture(
            references="        - https://github.com/virtengine/virtengine/issues/1079",
            reason=(
                "DoS with no patched release. Full assessment: "
                "docs/security/DOES-NOT-EXIST-ASSESSMENT.md"
            ),
        )

        errors = self.validator.validate_allowlist(path)

        self.assertTrue(
            any("docs/security/DOES-NOT-EXIST-ASSESSMENT.md" in error for error in errors), errors
        )

    def test_allowlist_accepts_a_reason_citing_a_document_that_exists(self) -> None:
        """The positive control, and the shape every real exception in the ledger uses: a reason
        that cites a repo document which is really in the tree must still validate."""
        path = self.allowlist_entry_fixture(
            references="        - https://github.com/virtengine/virtengine/issues/1079",
            reason=(
                "DoS with no patched release. Full assessment: "
                "docs/security/GO-2026-4740-MSGPACK-ASSESSMENT.md"
            ),
        )

        errors = self.validator.validate_allowlist(path)

        self.assertEqual([error for error in errors if "no such file exists" in error], [], errors)

    def test_reason_citation_rule_ignores_non_citation_repo_syntax(self) -> None:
        """The rule must not fire on a shell command, a Go import path or a package module path
        that merely starts with a repo-looking prefix. A gate that cries wolf on `go list -deps
        ./...` gets muted, which is the same end state as no gate at all."""
        path = self.allowlist_entry_fixture(
            references="        - https://github.com/virtengine/virtengine/issues/1079",
            reason=(
                "Verified with `go list -deps ./...`; the module github.com/shamaton/msgpack/v2 "
                "and ./.../scripts/supply-chain/go-module-policy.json aside, nothing links it"
            ),
        )

        errors = self.validator.validate_allowlist(path)

        self.assertEqual([error for error in errors if "no such file exists" in error], [], errors)

    def test_reason_citation_rule_ignores_paths_that_escape_the_repository(self) -> None:
        """A `..` traversal that climbs OUT of the repository, or an absolute path, is not a plain
        repository citation: the resolver must refuse to follow it out of the tree, and must not
        manufacture a failure for text that was never a citation. Note the boundary is the
        resolved location, not the spelling - `docs/../x.md` normalises to a path still inside the
        repository, so it stays a citation and is judged on whether the file exists."""
        path = self.allowlist_entry_fixture(
            references="        - https://github.com/virtengine/virtengine/issues/1079",
            reason="Out-of-tree artifact at /etc/docs/outside.md and docs/../../outside-repo.md",
        )

        errors = self.validator.validate_allowlist(path)

        self.assertEqual([error for error in errors if "no such file exists" in error], [], errors)

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

    def write_workflows(self, files: dict[str, str]) -> Path:
        """A directory of workflow files, so the repo-wide sweep can be exercised in one tree."""
        temp_dir = tempfile.TemporaryDirectory()
        self.addCleanup(temp_dir.cleanup)
        base = Path(temp_dir.name)
        for name, content in files.items():
            (base / name).write_text(content, encoding="utf-8")
        return base

    def one_job_workflow(self, steps: str) -> str:
        """A one-job workflow whose only variable is the step list under test."""
        return f"name: Fixture\njobs:\n  security-summary:\n    runs-on: ubuntu-latest\n    steps:\n{steps}"

    def test_repo_local_path_without_checkout_is_rejected(self) -> None:
        """The shipped shape. `security-summary` ran `bash .github/scripts/check_security_gate_results.sh`
        with no actions/checkout step, so the runner workspace was empty and the only job whose
        purpose is to report a failed gate died with exit 127 - while
        `.github/tests/test_security_gate_summary.sh`, which exercises the script in a local
        workspace, stayed green. The guard has to observe the job, not the script."""
        directory = self.write_workflows(
            {
                "security.yaml": self.one_job_workflow(
                    "      - name: Fail if any security gate failed\n"
                    "        run: bash .github/scripts/check_security_gate_results.sh success\n"
                )
            }
        )

        errors = self.validator.validate_workflows_checkout_coverage(directory)

        self.assertTrue(any("'security-summary'" in error for error in errors), errors)
        self.assertTrue(
            any(".github/scripts/check_security_gate_results.sh" in error for error in errors), errors
        )
        self.assertTrue(any("actions/checkout@v5" in error for error in errors), errors)

    def test_repo_local_path_with_checkout_is_accepted(self) -> None:
        """The positive control: the rule must not reject a job that checks the repository out."""
        directory = self.write_workflows(
            {
                "security.yaml": self.one_job_workflow(
                    "      - uses: actions/checkout@v5\n"
                    "      - name: Fail if any security gate failed\n"
                    "        run: bash .github/scripts/check_security_gate_results.sh success\n"
                )
            }
        )

        self.assertEqual(self.validator.validate_workflows_checkout_coverage(directory), [])

    def test_checkout_after_the_repo_local_step_is_still_rejected(self) -> None:
        """Too late is the same as absent: the first step still starts on an empty workspace."""
        directory = self.write_workflows(
            {
                "security.yaml": self.one_job_workflow(
                    "      - run: python .github/scripts/validate_security_policies.py\n"
                    "      - uses: actions/checkout@v5\n"
                )
            }
        )

        errors = self.validator.validate_workflows_checkout_coverage(directory)

        self.assertTrue(any("'security-summary'" in error for error in errors), errors)

    def test_repo_local_composite_action_without_checkout_is_rejected(self) -> None:
        """`uses: ./…` is a repository path too, and needs the same checkout."""
        directory = self.write_workflows(
            {"security.yaml": self.one_job_workflow("      - uses: ./.github/actions/lint\n")}
        )

        errors = self.validator.validate_workflows_checkout_coverage(directory)

        self.assertTrue(any(".github/actions/lint" in error for error in errors), errors)

    def test_paths_outside_the_repository_are_not_repo_local(self) -> None:
        """A `scripts/` directory that is not the repository's must not demand a checkout."""
        directory = self.write_workflows(
            {
                "security.yaml": self.one_job_workflow(
                    "      - run: /usr/local/bin/scripts/helper.sh\n"
                    "      - run: docker run --rm image /opt/app/scripts/entrypoint.sh\n"
                )
            }
        )

        self.assertEqual(self.validator.validate_workflows_checkout_coverage(directory), [])

    def test_sweep_covers_every_workflow_in_the_directory(self) -> None:
        """The rule is repo-wide: a defect in a workflow outside the four policy specs still fails,
        so the next script-ification cannot repeat the security-summary mistake elsewhere."""
        directory = self.write_workflows(
            {
                "unrelated.yaml": self.one_job_workflow(
                    "      - run: bash scripts/supply-chain/verify-pinned-images.sh\n"
                ),
                "security.yaml": self.one_job_workflow("      - uses: actions/checkout@v5\n"),
            }
        )

        errors = self.validator.validate_workflows_checkout_coverage(directory)

        self.assertEqual(len(errors), 1, errors)
        self.assertTrue(any("unrelated.yaml" in error for error in errors), errors)

    def test_sweep_fails_closed_on_an_empty_workflow_directory(self) -> None:
        """No workflow files is not 'nothing to check' - it is a surface that stopped being read."""
        temp_dir = tempfile.TemporaryDirectory()
        self.addCleanup(temp_dir.cleanup)

        errors = self.validator.validate_workflows_checkout_coverage(Path(temp_dir.name))

        self.assertTrue(any("no workflow files found" in error for error in errors), errors)


if __name__ == "__main__":
    unittest.main()
