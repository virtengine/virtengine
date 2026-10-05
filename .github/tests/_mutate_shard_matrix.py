#!/usr/bin/env python3
"""Mutate the go-vuln-scan shard matrix IN the security workflow.

This is the mutation engine for
`.github/tests/falsify_govulncheck_shard_partition.sh`. It edits
`.github/workflows/security.yaml` in place - preserving its line endings byte for
byte outside the mutation - and prints `changed=True|False` so the falsifier can
refuse to credit an inert mutation.

The workflow is the SUBJECT of the guard, so mutating it is the only way to
prove the guard observes the matrix. An earlier version of the falsifier mutated
the guard's own private copy of the matrix and was green on a guard that could
not see the matrix at all.

Usage: python .github/tests/_mutate_shard_matrix.py <case>
  noop             rewrite the file unchanged (control; prints changed=False)
  drop-shard       delete the tests-and-sim shard entry
  drop-util        remove ./util/... from tests-and-sim (uncovered packages)
  repoint-platform point the platform shard at ./infra/... (zero packages)
  duplicate-util   give ./util/... to tooling as well (overlap)
  empty-matrix     remove every shard entry (no shards at all)
Exit 3 if a case is unknown or its anchor is missing.
"""
from __future__ import annotations

import re
import sys

WORKFLOW = ".github/workflows/security.yaml"


def _eol_of(text: str) -> str:
    """Match the file's existing newline so the mutation is byte-minimal."""
    return "\r\n" if "\r\n" in text else "\n"


def _drop(text: str, pattern: str) -> str:
    """Remove the first regex match, asserting exactly one was present."""
    new, n = re.subn(pattern, "", text, count=1)
    if n != 1:
        raise SystemExit(f"anchor did not match: {pattern!r}")
    return new


def mutate(case: str, text: str) -> str:
    eol = _eol_of(text)
    if case == "noop":
        return text
    if case == "drop-shard":
        return _drop(
            text,
            r"[ \t]*- shard: tests-and-sim\r?\n[ \t]*patterns: \""
            r"\./tests/\.\.\. [^\"]*\"(?:\r?\n)",
        )
    if case == "drop-util":
        # Remove the trailing "./util/..." from the tests-and-sim pattern.
        return _drop(text, r" \./util/\.\.\.")
    if case == "repoint-platform":
        return re.sub(
            r'(- shard: platform\r?\n[ \t]*patterns: )"\./pkg/\.\.\."',
            r'\1"./infra/..."',
            text,
            count=1,
        )
    if case == "duplicate-util":
        # tooling also claims ./util/... , which tests-and-sim already owns.
        return re.sub(
            r'(- shard: tooling\r?\n[ \t]*patterns: "\./cmd/\.\.\. )',
            r'\1./util/... ./scripts/...',
            text,
            count=1,
        )
    if case == "empty-matrix":
        # Comment out every shard entry, leaving the include: list empty.
        out = re.sub(
            r"^([ \t]*)- shard: ([A-Za-z0-9_-]+)(\r?\n)"
            r"([ \t]*patterns: \"[^\"]*\")",
            lambda m: f"{m.group(1)}# - shard: {m.group(2)}{m.group(3)}"
            f"{m.group(4)}",
            text,
            flags=re.MULTILINE,
        )
        if out == text:
            raise SystemExit("empty-matrix: no shard entries matched")
        return out
    raise SystemExit(f"unknown case: {case}")


def main() -> int:
    if len(sys.argv) != 2:
        print(__doc__, file=sys.stderr)
        return 3
    case = sys.argv[1]
    # newline="" so universal-newline translation cannot rewrite the file.
    with open(WORKFLOW, encoding="utf-8", newline="") as fh:
        original = fh.read()

    try:
        out = mutate(case, original)
    except SystemExit as exc:
        print(f"ERROR: {exc}", file=sys.stderr)
        return 3

    if out == original:
        print(f"CASE={case} changed=False")
        return 0

    with open(WORKFLOW, "w", encoding="utf-8", newline="") as fh:
        fh.write(out)
    print(f"CASE={case} changed=True")
    return 0


if __name__ == "__main__":
    sys.exit(main())