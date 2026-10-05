#!/usr/bin/env python3
"""Fail closed unless every declared govulncheck shard produced a verdict.

Why this exists
---------------
`t_76147140` asked for a verdict-less shard to be distinguishable from a clean
one. Measurement first, and the premise turned out to be PARTLY false:
`check_security_gate_results.sh` already rejects a `cancelled` matrix
aggregate, so a cancelled shard is NOT reported clean. But that aggregate is a
single GitHub-level verdict for a four-way matrix, and it cannot say WHICH shard
has no verdict, or that four shards were expected and only three reported.

GitHub derives a matrix job's result from its step results, so two very different
outcomes collapse onto the same conclusion:

  * a shard whose scan completed and found nothing   -> `success`, artifact present
  * a shard cancelled before it produced any verdict  -> `cancelled`, and the
    `Upload govulncheck report` step is SKIPPED, so the artifact is ABSENT

Measured on Security run 37174317397 / job 111353646594: steps were
`success x5, cancelled, skipped, skipped, success` - the shard produced NO
govulncheck output at all, and the run's own artifacts still listed
`govulncheck-report-tests-and-sim` because the runner had uploaded it on the
first attempt. Artifact presence is therefore evidence of an attempt, not of a
verdict, and must not be read as one.

So this gate does not ask "was the matrix green?" - the summary already does
that. It asks a question the summary cannot: DID EVERY DECLARED SHARD REPORT A
VERDICT? Each shard writes a one-line ledger entry naming its verdict; a shard
that dies mid-scan writes none, and a missing entry is an UNKNOWN that fails
closed rather than being read as clean.

Deliberately NOT a weakening. No `continue-on-error`, no `|| true`, no
reclassification of an unfinished scan as clean: an absent verdict is red.

Usage: govulncheck_verdict_ledger.py <ledger-dir> [--expect N] [--json]
  Reads `<ledger-dir>/<shard>.verdict`, one `verdict=<value>` line per shard.
  Exit 0 only when every expected shard reports a recognised verdict.
Exit 1 = a shard has no verdict, or an unknown verdict; exit 2 = usage/IO error.
"""
from __future__ import annotations

import argparse
import json
import os
import re
import sys

# The verdict vocabulary is closed. `unknown` is deliberately NOT acceptable:
# a value this gate does not recognise must fail closed, because a future
# workflow that invents a new verdict state has not proved it is safe.
KNOWN = {"clean", "vulns", "error", "incomplete"}
VERDICT_RE = re.compile(r"^verdict=([A-Za-z0-9_-]+)\s*$", re.MULTILINE)


def read_shard_ledger(directory: str, shard: str) -> tuple[str | None, str]:
    """Return (verdict, note) for one shard. verdict is None when absent."""
    path = os.path.join(directory, f"{shard}.verdict")
    if not os.path.isfile(path):
        return None, f"{shard}: no verdict file was written"
    try:
        with open(path, encoding="utf-8", errors="replace") as fh:
            body = fh.read()
    except OSError as exc:
        return None, f"{shard}: verdict file unreadable ({exc})"
    # A shard writes exactly one verdict line. Zero matches means the file
    # exists but carries no verdict, which is the same failure as no file.
    found = VERDICT_RE.findall(body)
    if not found:
        return None, f"{shard}: verdict file carries no 'verdict=' line"
    if len(found) > 1:
        return None, f"{shard}: {len(found)} verdict lines (expected exactly 1)"
    verdict = found[0]
    if verdict not in KNOWN:
        return None, f"{shard}: unknown verdict '{verdict}' (not in {sorted(KNOWN)})"
    return verdict, f"{shard}: verdict={verdict}"


def expected_shards(explicit: int | None, directory: str) -> list[str]:
    """Shard names to require. `--expect N` needs the names to be discoverable.

    Without a manifest of names, an expected COUNT cannot be attributed to a
    shard, so the count alone would let a renamed shard look covered. The
    workflow writes `shards.manifest` listing the matrix names it deployed;
    that file is the authority for which shards must report.
    """
    manifest = os.path.join(directory, "shards.manifest")
    names: list[str] = []
    if os.path.isfile(manifest):
        with open(manifest, encoding="utf-8", errors="replace") as fh:
            for line in fh:
                line = line.strip()
                if line and not line.startswith("#"):
                    names.append(line)
    if names:
        return names
    if explicit is not None:
        raise ValueError(
            f"--expect {explicit} given but no shards.manifest in {directory}; "
            "a count cannot be attributed to a shard, so nothing is checked"
        )
    return []


def main(argv: list[str] | None = None) -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("ledger_dir", help="directory holding <shard>.verdict files")
    ap.add_argument(
        "--expect",
        type=int,
        default=None,
        help="require this many shards to report (needs shards.manifest)",
    )
    ap.add_argument("--json", action="store_true", help="emit a JSON report")
    args = ap.parse_args(argv)

    if not os.path.isdir(args.ledger_dir):
        print(f"ERROR: no ledger directory: {args.ledger_dir}", file=sys.stderr)
        return 2

    try:
        shards = expected_shards(args.expect, args.ledger_dir)
    except ValueError as exc:
        print(f"ERROR: {exc}", file=sys.stderr)
        return 2

    if not shards:
        print(
            f"ERROR: no shard manifest in {args.ledger_dir} and no --expect: "
            "there is nothing to verify, which is not a pass",
            file=sys.stderr,
        )
        return 2

    results = []
    unknown = 0
    for shard in shards:
        verdict, note = read_shard_ledger(args.ledger_dir, shard)
        if verdict is None:
            unknown += 1
            results.append({"shard": shard, "verdict": None, "note": note})
        else:
            results.append({"shard": shard, "verdict": verdict, "note": note})

    if args.expect is not None and len(shards) != args.expect:
        print(
            f"ERROR: manifest lists {len(shards)} shards but the matrix "
            f"declares {args.expect}",
            file=sys.stderr,
        )
        return 1

    if args.json:
        print(json.dumps(results, indent=2))
    else:
        for row in results:
            print(f"  {row['note']}")
        print(f"  shards declared: {len(shards)}; without a verdict: {unknown}")

    if unknown:
        print(
            f"::error::{unknown} of {len(shards)} govulncheck shard(s) produced NO "
            "verdict. A shard that did not finish is NOT clean: it is unknown. "
            "Re-run the gate; do not treat this as a passing scan.",
            file=sys.stderr,
        )
        return 1

    print(f"All {len(shards)} govulncheck shards reported a verdict")
    return 0


if __name__ == "__main__":
    sys.exit(main())