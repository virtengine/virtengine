#!/usr/bin/env python3
"""Read the go-vuln-scan shard matrix OUT OF the workflow file.

Why this exists
---------------
`.github/tests/test_govulncheck_shard_partition.sh` used to keep its own
hand-copied `SHARDS=(...)` array and verify THAT copy. It never read
`.github/workflows/security.yaml`, so the guard could not observe a change to
the matrix it claims to protect. Measured on 2026-10-05 by deleting the
`tests-and-sim` shard from the real workflow: the partition guard still printed
"RESULT: shard partition is EXACT" and exited 0. Adding a shard whose patterns
match zero packages was likewise invisible, even though the guard contains a
"shard matched 0 packages" check that could never fire for a workflow shard.

A guard that verifies a copy of its subject verifies that the copy is
self-consistent. Only parsing the real file can make the guard's claim true.

Usage: govulncheck_shards.py [--workflow PATH] [--list|--json]
  --list  print `name<TAB>patterns` per shard, one per line
  --json  print a JSON array of {"shard": ..., "patterns": ...}
  Exit 2 on a malformed workflow or a matrix this reader cannot resolve.
"""
from __future__ import annotations

import argparse
import json
import re
import sys

# Locate `jobs: -> go-vuln-scan: -> strategy: -> matrix: -> include:` without
# requiring PyYAML (the shell guard runs before/without the Python deps of the
# policy validator, and actionlint would be a heavier dependency than the
# structure deserves). Anchored on the job key so a lookalike string in a
# comment or a `needs:` list cannot be mistaken for the matrix.
JOB_RE = re.compile(r"^  go-vuln-scan:\s*$", re.MULTILINE)
SHARD_RE = re.compile(r"^\s*-\s*shard:\s*(\S+)\s*$")
PATTERNS_RE = re.compile(r"^\s*patterns:\s*(.+?)\s*$")


def parse_shards(text: str) -> list[dict[str, str]]:
    """Return the shard include-entries of the go-vuln-scan job matrix."""
    match = JOB_RE.search(text)
    if match is None:
        raise ValueError("go-vuln-scan job not found in the workflow")

    # Bound the scan to this job's own block: the next line that starts a new
    # top-level job (`^  <name>:`) ends it. Without this bound a `shard:` key
    # added to a LATER job would be attributed to go-vuln-scan.
    rest = text[match.end():]
    end = re.search(r"^  [A-Za-z][A-Za-z0-9_-]*:\s*$", rest, re.MULTILINE)
    block = rest[: end.start()] if end else rest

    # Only the strategy.matrix.include list defines shards.
    include_at = block.find("include:")
    if include_at == -1:
        raise ValueError("go-vuln-scan has no strategy.matrix.include list")

    shards: list[dict[str, str]] = []
    for line in block[include_at:].splitlines():
        shard_hit = SHARD_RE.match(line)
        if shard_hit:
            shards.append({"shard": shard_hit.group(1), "patterns": ""})
            continue
        pat_hit = PATTERNS_RE.match(line)
        # A `patterns:` line while no shard is open belongs to some other
        # mapping; ignoring it is safer than inventing a shard name.
        if pat_hit and shards:
            # A shard carries exactly one `patterns:` line; the first wins,
            # which keeps a duplicated key from silently widening coverage.
            if not shards[-1]["patterns"]:
                shards[-1]["patterns"] = pat_hit.group(1)
    return shards


def main(argv: list[str] | None = None) -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument(
        "--workflow",
        default=".github/workflows/security.yaml",
        help="path to the security workflow (default: %(default)s)",
    )
    group = ap.add_mutually_exclusive_group()
    group.add_argument("--list", action="store_true", help="name<TAB>patterns")
    group.add_argument("--json", action="store_true", help="JSON array")
    args = ap.parse_args(argv)

    try:
        with open(args.workflow, encoding="utf-8") as fh:
            shards = parse_shards(fh.read())
    except OSError as exc:
        print(f"ERROR: cannot read {args.workflow}: {exc}", file=sys.stderr)
        return 2
    except ValueError as exc:
        print(f"ERROR: {args.workflow}: {exc}", file=sys.stderr)
        return 2

    if not shards:
        print(
            f"ERROR: {args.workflow}: go-vuln-scan matrix resolved to 0 shards",
            file=sys.stderr,
        )
        return 2

    if args.json:
        print(json.dumps(shards, indent=2))
    elif args.list:
        for entry in shards:
            print(f"{entry['shard']}\t{entry['patterns']}")
    else:
        print(f"{len(shards)} shard(s) in {args.workflow}", file=sys.stderr)
        for entry in shards:
            print(f"  {entry['shard']}: {entry['patterns']}")
    return 0


if __name__ == "__main__":
    sys.exit(main())