#!/usr/bin/env python3

from __future__ import annotations

import argparse
import re
import sys
import tomllib
from dataclasses import dataclass
from datetime import date, datetime
from pathlib import Path
from typing import Callable

import yaml


ROOT = Path(__file__).resolve().parents[2]


@dataclass(frozen=True)
class WorkflowSpec:
    required_jobs: tuple[str, ...]
    required_snippets: tuple[str, ...]
    forbidden_snippets: tuple[str, ...] = ()
    # Groups of interchangeable spellings of the same required invocation: at least one
    # member of each group must appear. Only used where the GitHub-expression form
    # (${{ env.X }}) and the shell-env form (${X}) resolve to the same pinned value.
    required_snippet_aliases: tuple[tuple[str, ...], ...] = ()


WORKFLOW_SPECS: dict[str, WorkflowSpec] = {
    "security.yaml": WorkflowSpec(
        required_jobs=(
            "policy-validation",
            "codeql-analysis",
            "go-vuln-scan",
            "python-vuln-scan",
            "node-vuln-scan",
            "container-scan",
            "sbom-generation",
            "secret-scan",
            "gosec-scan",
            "security-summary",
        ),
        required_snippets=(
            "golang.org/x/vuln/cmd/govulncheck@${{ env.GOVULNCHECK_VERSION }}",
            "github.com/securego/gosec/v2/cmd/gosec@${{ env.GOSEC_VERSION }}",
            "pip-audit==2.10.0",
            "pnpm audit --prod --audit-level=high",
            "npm --prefix sdk/ts audit --audit-level=high",
            "scripts/supply-chain/generate-sbom.sh --format all",
            "gitleaks dir .",
            "trivy image",
        ),
    ),
    "supply-chain.yaml": WorkflowSpec(
        required_jobs=(
            "policy-validation",
            "dependency-verification",
            "lockfile-integrity",
            "attack-detection",
            "risk-assessment",
            "sbom-generation",
            "build-release-subjects",
            "sign-release-artifacts",
            "verify-release-signatures",
            "provenance",
            "verify-provenance",
            "summary",
        ),
        required_snippets=(
            "scripts/supply-chain/detect-supply-chain-attacks.sh --all --json",
            "go run ./scripts/supply-chain/assess-dependencies.go --report --json",
            "scripts/supply-chain/generate-sbom.sh --format all",
            "cosign sign-blob",
            "cosign verify-blob",
            "slsa-verifier@${{ env.SLSA_VERIFIER_VERSION }}",
            'slsa-verifier" verify-artifact',
            "generator_generic_slsa3.yml@v2.1.0",
        ),
    ),
    "license-compliance.yaml": WorkflowSpec(
        required_jobs=(
            "policy-validation",
            "go-licenses",
            "javascript-licenses",
            "python-licenses",
            "spdx-sbom",
            "license-summary",
        ),
        required_snippets=(
            "github.com/google/go-licenses@${{ env.GO_LICENSES_VERSION }}",
            "license-checker-rseidelsohn@${{ env.LICENSE_CHECKER_VERSION }}",
            "scripts/supply-chain/generate-sbom.sh --format spdx",
        ),
        # #888 installs the pinned scanner inside each requirement venv, where the shell-env
        # form of the same workflow-level pin is the correct spelling. Accept either, but
        # keep requiring an explicitly pinned pip-licenses install.
        required_snippet_aliases=(
            (
                "pip-licenses==${{ env.PIP_LICENSES_VERSION }}",
                "pip-licenses==${PIP_LICENSES_VERSION}",
            ),
        ),
    ),
    "pr-security-check.yaml": WorkflowSpec(
        required_jobs=(
            "policy-validation",
            "analyze-changes",
            "go-security",
            "gitleaks-scan",
            "dependency-review",
            "security-lint",
            "security-summary",
        ),
        required_snippets=(
            "golang.org/x/vuln/cmd/govulncheck@${{ env.GOVULNCHECK_VERSION }}",
            "github.com/securego/gosec/v2/cmd/gosec@${{ env.GOSEC_VERSION }}",
            "gitleaks git .",
            "validate_inference_deployment_policy.py",
        ),
        forbidden_snippets=("coverage-gate",),
    ),
}


FORBIDDEN_GLOBAL_WORKFLOW_SNIPPETS = (
    "@latest",
    "continue-on-error: true",
    "|| true",
)

FORBIDDEN_SECRET_REFERENCES = {"GITHUB_TOKEN"}

FORBIDDEN_GITLEAKS_PATHS = {
    ".*\\.md$",
    "_docs/.*",
    "docs/.*",
    "portal/src/.*",
    "lib/portal/.*",
    "lib/admin/.*",
    "lib/capture/.*",
    ".*mock.*",
    ".*sample.*",
    ".*\\.env$",
    ".*\\.env\\..*",
    "tests?/.*",
}

REQUIRED_ALLOWLIST_FIELDS = {
    "id",
    "package",
    "reason",
    "reviewed_by",
    "reviewed_date",
    "expires",
    "references",
    "compensating_controls",
}

# The control behind policy.require_issue_reference: an exception must point at a REAL issue or
# pull request in this repository, by positive integer id. Advisory pages and vendor links are
# evidence, not ownership - and a placeholder (".../issues/NEW") is neither. Two shipped
# exceptions carried `issues/NEW` (HTTP 404) while the policy advertised that every entry was
# tracked, and nothing failed: the key was declared and never read. Regression case:
# .github/tests/test_security_policy_validator.py
TRACKING_REFERENCE_PATTERN = re.compile(r"^https://github\.com/virtengine/virtengine/(?:issues|pull)/[1-9][0-9]*$")

# References are evidence, and a reference into this repository has to resolve. The same shipped
# allowlist that pointed at a placeholder issue also cited
# docs/security/GO-2022-06*-S3CRYPTO-ASSESSMENT.md and docs/security/GO-2026-4513-MSGPACK-ASSESSMENT.md
# - files that did not exist. The tracking-reference rule cannot see that, because a link to an
# advisory page and a link to a missing blob are both "a reference": this rule is what makes the
# evidence real.
REPO_BLOB_REFERENCE_PATTERN = re.compile(
    r"^https://github\.com/virtengine/virtengine/blob/[^/]+/(?P<path>.+)$"
)

# ...and so is a doc path written as BARE PROSE. Every shipped assessment citation lives in the
# `reason:` field, not in `references:`:
#     reason: '... Full assessment: docs/security/GO-2026-4740-MSGPACK-ASSESSMENT.md'
# The blob rule walks `references` only, so that prose citation is invisible to it. Measured on
# develop @ 6c5fb13e: repointing only the prose to a file that does not exist left
# validate_security_policies.py at exit 0, with the entry still passing - the exact failure mode
# the blob rule was written to close, reopened through the other field. The natural way to DRAFT
# an entry is to write its assessment citation in the reason, so the unchecked field is the one
# new entries reach for. This pattern extracts repo-relative doc paths out of free text.
#
# Narrow on purpose: it matches a path that starts at a known top-level entry and carries a
# documentation extension, so a prose mention of `go list -deps ./...`, a package path or a
# command flag cannot be mistaken for a cited document. Regression case:
# .github/tests/test_security_policy_validator.py
PROSE_REPO_PATH_PATTERN = re.compile(
    r"(?<![\w/.-])(?P<path>(?:docs|scripts|_docs|\.github)/[\w./-]*"
    r"\.(?:md|rst|txt|json|ya?ml|sh|py))"
)

# A cited path is only evidence if it resolves to a FILE inside the repository. `ROOT / path`
# escapes upward for an absolute or `..`-bearing candidate, so normalise and confine.


def _resolve_prose_repo_path(path: str) -> str | None:
    """Return a validation error message, or None when the cited path resolves to a file."""
    try:
        candidate = (ROOT / path).resolve()
        candidate.relative_to(ROOT.resolve())
    except (OSError, ValueError):
        # Escapes ROOT (absolute path, or `..` traversal) - not a plain repo citation.
        return None
    if not candidate.is_file():
        return (
            f"cites {path} in its reason but no such file exists in the repository; a "
            "prose citation of a missing document is a 404, not evidence"
        )
    return None


def unresolved_repo_reference_errors(entry_id: str, references: list) -> list[str]:
    """Report references that point at a repository path which does not exist."""
    errors: list[str] = []
    for reference in references:
        if not isinstance(reference, str):
            continue
        match = REPO_BLOB_REFERENCE_PATTERN.match(reference)
        if match is None:
            continue
        path = match.group("path")
        if not (ROOT / path).exists():
            errors.append(
                f"exceptions entry {entry_id} cites {path} but no such file exists in the "
                "repository; a blob link to a missing file is a 404, not evidence"
            )
    return errors


def unresolved_reason_citation_errors(entry_id: str, reason: object) -> list[str]:
    """Report repo-relative document paths cited in an entry's prose that do not exist.

    The `reason` field is the natural place to cite an assessment document, and the
    reference-channel rule cannot see it. Same obligation, second field: if a citation is
    required to be real, it is required to be real wherever it is written.
    """
    if not isinstance(reason, str):
        return []
    errors: list[str] = []
    for match in PROSE_REPO_PATH_PATTERN.finditer(reason):
        path = match.group("path")
        problem = _resolve_prose_repo_path(path)
        if problem is not None:
            errors.append(f"exceptions entry {entry_id} {problem}")
    return errors


# A `run:` step that executes a repository-local path can only work when the job has checked the
# repository out first. The security-summary job shipped
# `bash .github/scripts/check_security_gate_results.sh` with no actions/checkout step: the runner
# workspace is empty, so the only step that can report a failed gate died with exit 127,
# "No such file or directory", while the ten gate results it was handed had already resolved to
# success. `.github/tests/test_security_gate_summary.sh` exercises the script's logic in a local
# workspace and stayed green the whole time - a guard that cannot observe the job it guards. This
# rule checks the job: a job that runs a repo-local path must check the repository out before it.
REPO_LOCAL_PATH_PATTERN = re.compile(
    r"(?:^|(?<=[\s\"'=(]))(?:\./)?(?:\.github/(?:scripts|tests|actions)/|scripts/)[A-Za-z0-9_./-]+"
)
CHECKOUT_STEP_PATTERN = re.compile(r"^actions/checkout@")


def repo_local_path_errors(workflow_name: str, job_name: str, job: dict) -> list[str]:
    """Report steps that run a repository-local path before any actions/checkout step."""
    steps = job.get("steps")
    if not isinstance(steps, list):
        return []

    checked_out = False
    offenders: list[str] = []
    for step in steps:
        if not isinstance(step, dict):
            continue
        uses = str(step.get("uses") or "")
        if CHECKOUT_STEP_PATTERN.match(uses):
            checked_out = True
            continue
        if uses.startswith("./") and not checked_out:
            # A repo-local action is a repository path too, and needs the same checkout.
            offenders.append(uses)
            continue
        run = step.get("run")
        if not isinstance(run, str):
            continue
        for match in REPO_LOCAL_PATH_PATTERN.finditer(run):
            if not checked_out:
                offenders.append(match.group(0).strip())

    if not offenders:
        return []
    listed = ", ".join(sorted(set(offenders)))
    return [
        f"{workflow_name}: job '{job_name}' runs the repo-local path(s) {listed} but no "
        "actions/checkout step precedes it; the runner workspace is empty and the step fails "
        "with 'No such file or directory' instead of reporting a result. Add "
        "`- uses: actions/checkout@v5` before the first such step."
    ]


def validate_workflows_checkout_coverage(workflows_dir: Path) -> list[str]:
    """Every job in `.github/workflows` that runs a repo-local path must check the repo out."""
    if not workflows_dir.is_dir():
        return [f"workflows directory does not exist: {repo_path(workflows_dir)}"]

    workflow_paths = sorted(workflows_dir.glob("*.yaml"))
    if not workflow_paths:
        return [f"no workflow files found in {repo_path(workflows_dir)}"]

    errors: list[str] = []
    for path in workflow_paths:
        try:
            workflow = load_yaml(path)
        except Exception as exc:  # pragma: no cover - defensive, load_yaml is exercised elsewhere
            errors.append(f"{path.name}: could not be parsed ({exc.__class__.__name__}: {exc})")
            continue
        jobs = workflow.get("jobs", {})
        if not isinstance(jobs, dict):
            continue
        for job_name, job in jobs.items():
            if isinstance(job, dict):
                errors.extend(repo_local_path_errors(path.name, str(job_name), job))

    return errors


def repo_path(path: Path) -> str:
    try:
        return str(path.relative_to(ROOT)).replace("\\", "/")
    except ValueError:
        return str(path)


def load_yaml(path: Path) -> dict:
    data = yaml.safe_load(path.read_text(encoding="utf-8"))
    if not isinstance(data, dict):
        raise ValueError("YAML document must be a mapping")
    return data


def validate_defaults_shell(workflow: dict) -> list[str]:
    defaults = workflow.get("defaults", {})
    run_defaults = defaults.get("run", {}) if isinstance(defaults, dict) else {}
    if run_defaults.get("shell") != "bash":
        return ["defaults.run.shell must be bash"]
    return []


def validate_top_permissions(workflow: dict) -> list[str]:
    errors: list[str] = []
    permissions = workflow.get("permissions", {})
    if not isinstance(permissions, dict):
        return ["top-level permissions must be an explicit mapping"]
    for forbidden_permission in ("id-token", "packages", "attestations"):
        if forbidden_permission in permissions:
            errors.append(f"top-level permissions must not include {forbidden_permission}")
    return errors


def validate_secret_references(raw: str) -> list[str]:
    errors: list[str] = []
    for token in set(part.split("}}", 1)[0].strip() for part in raw.split("secrets.")[1:]):
        secret_name = token.split()[0].split("|")[0].split(")")[0].strip(" ,}")
        if secret_name and secret_name not in FORBIDDEN_SECRET_REFERENCES:
            errors.append(f"unexpected secret reference secrets.{secret_name}")
    return errors


def validate_workflow(path: Path) -> list[str]:
    raw = path.read_text(encoding="utf-8")
    workflow = load_yaml(path)
    errors: list[str] = []

    errors.extend(validate_defaults_shell(workflow))
    errors.extend(validate_top_permissions(workflow))
    errors.extend(validate_secret_references(raw))

    for snippet in FORBIDDEN_GLOBAL_WORKFLOW_SNIPPETS:
        if snippet in raw:
            errors.append(f"workflow contains forbidden snippet: {snippet}")

    spec = WORKFLOW_SPECS[path.name]
    jobs = workflow.get("jobs", {})
    if not isinstance(jobs, dict):
        errors.append("jobs must be a mapping")
        return errors

    for job_name in spec.required_jobs:
        if job_name not in jobs:
            errors.append(f"missing required job: {job_name}")

    if "validate_security_policies.py" not in raw:
        errors.append("policy-validation must run validate_security_policies.py")
    accepted_test_invocations = (
        'python -m unittest discover -s .github/tests -p "test_security_policy*.py"',
        'python -m unittest discover -s .github/tests -p "test_*policy*.py"',
    )
    if not any(invocation in raw for invocation in accepted_test_invocations):
        errors.append("policy-validation must run policy validator tests")
    if "actionlint@v1.7.12" not in raw:
        errors.append("policy-validation must run actionlint at v1.7.12")

    for snippet in spec.required_snippets:
        if snippet not in raw:
            errors.append(f"missing required snippet: {snippet}")

    for alias_group in spec.required_snippet_aliases:
        if not any(alias in raw for alias in alias_group):
            errors.append(f"missing required snippet: {alias_group[0]}")

    for snippet in spec.forbidden_snippets:
        if snippet in raw:
            errors.append(f"workflow contains forbidden snippet: {snippet}")

    return errors


def validate_gitleaks(path: Path) -> list[str]:
    raw = path.read_text(encoding="utf-8")
    data = tomllib.loads(raw)
    errors: list[str] = []

    extend = data.get("extend", {})
    if extend.get("useDefault") is not True:
        errors.append("gitleaks config must extend the default rule set")

    allowlist = data.get("allowlist", {})
    paths = set(allowlist.get("paths", []))
    for forbidden in FORBIDDEN_GITLEAKS_PATHS:
        if forbidden in paths:
            errors.append(f"gitleaks allowlist contains forbidden broad path exemption: {forbidden}")

    required_paths = {
        "(^|/)vendor/",
        "(^|/)node_modules/",
        "(^|/)(testdata|fixtures|__snapshots__)/",
    }
    for required in required_paths:
        if required not in paths:
            errors.append(f"gitleaks allowlist is missing required scoped exemption: {required}")

    return errors


def validate_allowlist(path: Path) -> list[str]:
    raw = path.read_text(encoding="utf-8")
    data = load_yaml(path)
    errors: list[str] = []

    for forbidden_text in ("Example entry", "Initial vulnerability allowlist created", "2024-01-01", "security-team"):
        if forbidden_text in raw:
            errors.append(f"allowlist contains stale placeholder text: {forbidden_text}")

    policy = data.get("policy", {})
    exceptions = data.get("exceptions", {})
    if not isinstance(policy, dict):
        errors.append("allowlist policy must be a mapping")
        return errors
    if not isinstance(exceptions, dict):
        errors.append("allowlist exceptions must be a mapping")
        return errors

    required_exception_keys = {"go", "python", "npm", "containers"}
    if set(exceptions.keys()) != required_exception_keys:
        errors.append("allowlist exceptions must include exactly go, python, npm, and containers")

    active_exception_count = sum(len(entries or []) for entries in exceptions.values())
    if policy.get("active_exception_count") != active_exception_count:
        errors.append("policy.active_exception_count does not match the exception list contents")

    max_age = policy.get("max_allowlist_age_days")
    if not isinstance(max_age, int) or max_age > 30:
        errors.append("policy.max_allowlist_age_days must be an integer no greater than 30")

    # Both switches were declared in the policy and never read, so "every exception is tracked
    # and compensated" was an assertion no run could falsify. Fail closed if either is turned off.
    if policy.get("require_issue_reference") is not True:
        errors.append("policy.require_issue_reference must be enabled")
    if policy.get("require_compensating_controls") is not True:
        errors.append("policy.require_compensating_controls must be enabled")

    today = date.today()
    for ecosystem, entries in exceptions.items():
        if not isinstance(entries, list):
            errors.append(f"exceptions.{ecosystem} must be a list")
            continue
        for entry in entries:
            if not isinstance(entry, dict):
                errors.append(f"exceptions.{ecosystem} entries must be mappings")
                continue
            missing = REQUIRED_ALLOWLIST_FIELDS - set(entry.keys())
            if missing:
                errors.append(f"exceptions.{ecosystem} entry missing fields: {sorted(missing)}")
                continue
            reviewed_date = datetime.strptime(str(entry["reviewed_date"]), "%Y-%m-%d").date()
            expires = datetime.strptime(str(entry["expires"]), "%Y-%m-%d").date()
            if expires < today:
                errors.append(f"exceptions.{ecosystem} entry {entry['id']} is expired")
            if (expires - reviewed_date).days > max_age:
                errors.append(f"exceptions.{ecosystem} entry {entry['id']} exceeds the maximum allowlist age")

            references = entry["references"]
            if not isinstance(references, list) or not references:
                errors.append(f"exceptions.{ecosystem} entry {entry['id']} must list at least one reference")
            elif not any(
                isinstance(reference, str) and TRACKING_REFERENCE_PATTERN.match(reference) for reference in references
            ):
                errors.append(
                    f"exceptions.{ecosystem} entry {entry['id']} needs a tracking reference to a "
                    "virtengine/virtengine issue or pull request (policy.require_issue_reference); "
                    "advisory URLs are evidence and a placeholder such as .../issues/NEW is not a reference"
                )

            controls = entry["compensating_controls"]
            if not isinstance(controls, list) or not controls or not all(
                isinstance(control, str) and control.strip() for control in controls
            ):
                errors.append(
                    f"exceptions.{ecosystem} entry {entry['id']} must list at least one compensating control "
                    "(policy.require_compensating_controls)"
                )

            if isinstance(references, list):
                errors.extend(unresolved_repo_reference_errors(str(entry["id"]), references))

            # The prose channel carries the same obligation as the reference list: a doc path
            # cited in `reason` is a citation, and a missing one is a 404, not evidence.
            errors.extend(unresolved_reason_citation_errors(str(entry["id"]), entry.get("reason")))

    return errors


def validate_supply_chain_doc(path: Path) -> list[str]:
    raw = path.read_text(encoding="utf-8")
    errors: list[str] = []

    required_strings = (
        "validate_security_policies.py",
        'python -m unittest discover -s .github/tests -p "test_security_policy*.py"',
        "actionlint@v1.7.12",
        "OIDC",
        "cosign verify-blob",
        "slsa-verifier verify-artifact",
        ".vulnerability-allowlist.yaml",
    )
    forbidden_strings = (
        "_Last updated: 2024_",
        "SLSA Level 3",
        "In Progress",
    )

    for snippet in required_strings:
        if snippet not in raw:
            errors.append(f"documentation is missing required text: {snippet}")

    for snippet in forbidden_strings:
        if snippet in raw:
            errors.append(f"documentation contains stale or unsupported claim: {snippet}")

    return errors


VALIDATORS: dict[str, Callable[[Path], list[str]]] = {
    "security.yaml": validate_workflow,
    "supply-chain.yaml": validate_workflow,
    "license-compliance.yaml": validate_workflow,
    "pr-security-check.yaml": validate_workflow,
    # Keyed on the directory's own name: the `.github/workflows` surface target is a directory,
    # because the rule it carries is repo-wide (any workflow, not only the four specs above).
    "workflows": validate_workflows_checkout_coverage,
    ".gitleaks.toml": validate_gitleaks,
    ".vulnerability-allowlist.yaml": validate_allowlist,
    "SUPPLY_CHAIN_SECURITY.md": validate_supply_chain_doc,
}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Validate VirtEngine security workflow and policy files.")
    parser.add_argument(
        "--files",
        nargs="+",
        help="Specific files to validate. Defaults to the full A17 security surface.",
    )
    return parser.parse_args()


def resolve_targets(args: argparse.Namespace) -> list[Path]:
    default_files = [
        ".github/workflows/security.yaml",
        ".github/workflows/supply-chain.yaml",
        ".github/workflows/license-compliance.yaml",
        ".github/workflows/pr-security-check.yaml",
        ".github/workflows",
        ".gitleaks.toml",
        ".vulnerability-allowlist.yaml",
        "SUPPLY_CHAIN_SECURITY.md",
    ]
    selected = args.files or default_files
    return [Path(item) if Path(item).is_absolute() else ROOT / item for item in selected]


def main() -> int:
    args = parse_args()
    targets = resolve_targets(args)
    found_errors = False

    for target in targets:
        if not target.exists():
            print(f"FAIL {repo_path(target)}")
            print("  - file does not exist")
            found_errors = True
            continue

        validator = VALIDATORS.get(target.name)
        if validator is None:
            print(f"FAIL {repo_path(target)}")
            print("  - no validator is registered for this file")
            found_errors = True
            continue

        try:
            errors = validator(target)
        except Exception as exc:  # pragma: no cover
            print(f"FAIL {repo_path(target)}")
            print(f"  - validator raised {exc.__class__.__name__}: {exc}")
            found_errors = True
            continue

        if errors:
            print(f"FAIL {repo_path(target)}")
            for error in errors:
                print(f"  - {error}")
            found_errors = True
        else:
            print(f"PASS {repo_path(target)}")

    return 1 if found_errors else 0


if __name__ == "__main__":
    raise SystemExit(main())
