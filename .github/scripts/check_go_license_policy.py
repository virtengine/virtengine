#!/usr/bin/env python3
"""Enforce the Go license policy, honouring the policy's own exceptions.

`license-compliance.yaml`'s Go License Check used to inline this check and key
only on `license.deniedTokens`. That meant the reviewed `license.exceptions`
list in `scripts/supply-chain/go-module-policy.json` was inert: the policy
declared `github.com/ethereum/go-ethereum` (GPL-3.0) and
`github.com/zondax/hid/libusb/libusb` (LGPL-2.1) as reviewed exceptions, while
the gate rejected both on every run. A policy that states an exception and a
gate that ignores it is worse than no exception list at all, because the
declaration makes reviewers believe the case is closed.

Exit codes: 0 clean, 1 policy violations found, 2 usage/input error.

Usage: check_go_license_policy.py <licenses.csv> [policy.json]
"""

from __future__ import annotations

import argparse
import csv
import json
import sys
from pathlib import Path

DEFAULT_POLICY = "scripts/supply-chain/go-module-policy.json"


def normalise(path: str) -> str:
    """Strip a trailing slash and the conventional ./ or / prefix."""
    p = path.strip().rstrip("/")
    while p.startswith("./"):
        p = p[2:]
    return p.lstrip("/")


def is_excepted(package: str, exceptions: list[str]) -> bool:
    """True when `package` is covered by a reviewed exception entry.

    An exception names a module or package path. It covers that exact path and
    anything beneath it, but a plain string prefix is not enough: the entry
    `github.com/ethereum/go-ethereum` must not silently except the unrelated
    `github.com/ethereum/go-ethereum-evil`, so the boundary is a path segment.
    """
    pkg = normalise(package)
    for entry in exceptions:
        exc = normalise(str(entry))
        if not exc:
            continue
        if pkg == exc or pkg.startswith(exc + "/"):
            return True
    return False


def find_violations(
    rows: list[list[str]], policy: dict
) -> tuple[list[str], list[str]]:
    """Return (violations, excepted) for the given go-licenses CSV rows.

    `violations` are packages whose license carries a denied token and which no
    exception covers; `excepted` are the ones an exception did cover, so the run
    can say out loud what it allowed through rather than hiding it.
    """
    license_policy = policy["license"]
    denied = tuple(str(t).upper() for t in license_policy["deniedTokens"])
    exceptions = list(license_policy.get("exceptions") or [])

    violations: list[str] = []
    excepted: list[str] = []

    for row in rows:
        if not row:
            continue
        package = normalise(row[0])
        license_name = (row[1] if len(row) > 1 else "UNKNOWN") or "UNKNOWN"
        if not any(token in license_name.upper() for token in denied):
            continue
        entry = f"{package}: {license_name}"
        if is_excepted(package, exceptions):
            excepted.append(entry)
        else:
            violations.append(entry)

    return violations, excepted


def load_policy(path: Path) -> dict:
    data = json.loads(path.read_text(encoding="utf-8"))
    license_policy = data.get("license")
    if not isinstance(license_policy, dict):
        raise ValueError(f"{path}: missing license policy block")
    if not license_policy.get("deniedTokens"):
        raise ValueError(f"{path}: license.deniedTokens must be a non-empty array")
    exceptions = license_policy.get("exceptions")
    if exceptions is not None and not isinstance(exceptions, list):
        raise ValueError(f"{path}: license.exceptions must be an array when present")
    return data


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("report", help="go-licenses CSV report path")
    parser.add_argument("policy", nargs="?", default=DEFAULT_POLICY,
                       help=f"policy JSON (default: {DEFAULT_POLICY})")
    args = parser.parse_args(argv)

    report_path = Path(args.report)
    policy_path = Path(args.policy)

    if not report_path.is_file():
        print(f"::error::license report not found: {report_path}", file=sys.stderr)
        return 2
    try:
        policy = load_policy(policy_path)
    except (OSError, ValueError) as exc:
        print(f"::error::cannot read license policy: {exc}", file=sys.stderr)
        return 2

    with report_path.open(newline="", encoding="utf-8") as fh:
        rows = list(csv.reader(fh))

    violations, excepted = find_violations(rows, policy)

    for entry in excepted:
        print(f"excepted (reviewed in {policy_path}): {entry}")

    if violations:
        print("Disallowed Go licenses detected:", file=sys.stderr)
        for entry in violations:
            print(entry, file=sys.stderr)
        return 1

    total = len(rows)
    print(
        f"Go license policy passed: {total} package(s) checked, "
        f"{len(excepted)} covered by a reviewed exception."
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())
