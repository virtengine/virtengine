#!/usr/bin/env python3

from __future__ import annotations

import argparse
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
            "pip-licenses==${{ env.PIP_LICENSES_VERSION }}",
            "scripts/supply-chain/generate-sbom.sh --format spdx",
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
    if 'python -m unittest discover -s .github/tests -p "test_security_policy*.py"' not in raw:
        errors.append("policy-validation must run security policy tests")
    if "actionlint@v1.7.12" not in raw:
        errors.append("policy-validation must run actionlint at v1.7.12")

    for snippet in spec.required_snippets:
        if snippet not in raw:
            errors.append(f"missing required snippet: {snippet}")

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
