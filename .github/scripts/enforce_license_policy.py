#!/usr/bin/env python3

import argparse
import json
import sys
from dataclasses import dataclass
from pathlib import Path


@dataclass(frozen=True)
class PolicyResult:
    violations: list[str]
    warnings: list[str]


# Ecosystems whose licenses default to undetermined when absent/null
UNKNOWN_LICENSES = {"undetermined", "unknown", "none", ""}


def is_undetermined(license_value: str | None) -> bool:
    """Return True when the license value is explicitly undetermined or unknown."""
    if license_value is None:
        return True
    return license_value.strip().lower() in UNKNOWN_LICENSES


def check_package(
    ecosystem: str,
    name: str,
    version: str,
    license_value: str | None,
    allowed_licenses: set[str],
    overrides: dict[str, set[str]] | None = None,
) -> tuple[list[str], list[str]]:
    """
    Check a single package's license against the allowed set.

    Returns (violations, warnings).
    """
    violations: list[str] = []
    warnings: list[str] = []

    # CRITICAL FIX: handle explicit null BEFORE any stringification.
    # Previously, a JSON null became the string "None" which slipped past
    # the is_undetermined(None) guard and passed the check incorrectly.
    if license_value is None or is_undetermined(license_value):
        violations.append(f"{ecosystem}:{name}@{version} license is undetermined/null")
        return violations, warnings

    # Check against disallowed tokens
    upper = license_value.upper()
    disallowed_tokens = {"GPL", "AGPL", "LGPL", "SSPL", "BUSL"}
    for token in disallowed_tokens:
        if token in upper:
            violations.append(f"{ecosystem}:{name}@{version} disallowed license: {license_value}")
            return violations, warnings

    # Check against allowed set if non-empty
    if allowed_licenses and license_value.strip() not in allowed_licenses:
        # Check for override
        pkg_key = f"{ecosystem}/{name}"
        if overrides and pkg_key in overrides:
            if license_value.strip() in overrides[pkg_key]:
                warnings.append(f"{ecosystem}:{name}@{version} license {license_value} is overridden")
                return violations, warnings
        violations.append(f"{ecosystem}:{name}@{version} license {license_value} not in allowed set")

    return violations, warnings


def check_javascript_reports(
    report_path: str | None = None,
    allowed_licenses: set[str] | None = None,
    overrides: dict[str, set[str]] | None = None,
) -> tuple[int, list[str]]:
    """
    Check JavaScript license reports from license-checker JSON output.

    Fix for t_f289c053: When ``licenses`` is an explicit JSON ``null``, it must
    NOT be stringified to ``"None"`` before calling check_package — that bypasses
    the is_undetermined(None) guard.  We inspect the raw value first, and only
    then fall through to the existing string path for non-null entries.
    """
    allowed = allowed_licenses or set()
    overrides = overrides or {}
    errors: list[str] = []

    if report_path:
        raw = Path(report_path).read_text(encoding="utf-8")
    else:
        raw = sys.stdin.read()

    try:
        report = json.loads(raw)
    except json.JSONDecodeError as exc:
        print(f"::error::Malformed JSON in JS license report: {exc}", file=sys.stderr)
        return 2, []

    if not isinstance(report, dict):
        print("::error::JS license report must be a JSON object (dict), got " f"{type(report).__name__}", file=sys.stderr)
        return 2, []

    for pkg, info in report.items():
        if not isinstance(info, dict):
            errors.append(f"javascript:{pkg} has non-object entry")
            continue

        licenses_raw = info.get("licenses")

        # CRITICAL FIX: handle explicit null BEFORE stringification
        if licenses_raw is None:
            errors.append(f"javascript:{pkg} has explicit null license (treat as undetermined)")
            continue

        # Stringify only non-null values
        licenses_str = str(licenses_raw)
        violations, _warnings = check_package("javascript", pkg, info.get("version", "unknown"), licenses_str, allowed, overrides)
        errors.extend(violations)

    return (1 if errors else 0), errors


def check_python_reports(
    report_path: str | None = None,
    allowed_licenses: set[str] | None = None,
    overrides: dict[str, set[str]] | None = None,
) -> tuple[int, list[str]]:
    """
    Check Python license reports from pip-licenses JSON output.

    Malformed input (dict instead of list) now produces a clean ::error::
    line and exits 2, matching the contract established for other ecosystems.
    """
    allowed = allowed_licenses or set()
    overrides = overrides or {}
    errors: list[str] = []

    if report_path:
        raw = Path(report_path).read_text(encoding="utf-8")
    else:
        raw = sys.stdin.read()

    try:
        report = json.loads(raw)
    except json.JSONDecodeError as exc:
        print(f"::error::Malformed JSON in Python license report: {exc}", file=sys.stderr)
        return 2, []

    # Malformed python report: dict-shaped JSON instead of a list → clean error
    if not isinstance(report, list):
        print(
            f"::error::Python license report must be a JSON array, got {type(report).__name__}",
            file=sys.stderr,
        )
        return 2, []

    for entry in report:
        if not isinstance(entry, dict):
            errors.append(f"python entry is not an object: {entry!r}")
            continue

        name = entry.get("Name") or entry.get("name") or "unknown"
        version = entry.get("Version") or entry.get("version") or "unknown"
        license_value = entry.get("License") or entry.get("license")

        violations, _warnings = check_package("python", name, version, license_value, allowed, overrides)
        errors.extend(violations)

    return (1 if errors else 0), errors


def main() -> int:
    parser = argparse.ArgumentParser(description="Enforce license policy for VirtEngine")
    parser.add_argument("--ecosystem", choices=["javascript", "python"], required=True)
    parser.add_argument("--report", type=str, default=None, help="Path to JSON report file")
    parser.add_argument("--allowed", type=str, default="", help="Comma-separated allowed licenses")
    args = parser.parse_args()

    allowed = set(args.allowed.split(",")) if args.allowed else set()

    if args.ecosystem == "javascript":
        rc, errors = check_javascript_reports(args.report, allowed)
    elif args.ecosystem == "python":
        rc, errors = check_python_reports(args.report, allowed)
    else:
        print(f"::error::Unsupported ecosystem: {args.ecosystem}", file=sys.stderr)
        return 2

    for e in errors:
        print(f"::error::{e}", file=sys.stderr)

    return rc


if __name__ == "__main__":
    sys.exit(main())
