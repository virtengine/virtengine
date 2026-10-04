#!/usr/bin/env python3
"""Per-linter lint debt budget gate (ratchet), the errcheck analogue of #1166.

Why this exists
---------------
`.golangci.yaml` previously set no issue caps, so golangci-lint v2 applied its
defaults: `max-issues-per-linter=50`, `max-same-issues=3`. Those caps are not a
statement about the tree, they are a display limit. On `develop` at 1cd9af8e3f
the capped CI run printed golangci-lint's own tally as errcheck=50 and
goconst=50 -- both pinned exactly at the cap -- while the uncapped count on the
same tree is errcheck 662 and goconst 926. Anyone reading the CI log believed
the lint debt was 50 per linter; it was 662, and the cap hid ~87% of it.

`.golangci.yaml` now sets both limits to 0 so the reported count is the real
count. That makes the debt honest but leaves it unbounded in the other
direction: nothing stops it growing. This gate is the other half.

Design
------
Absolute zero is unreachable here (the debt is 662 findings over 213 files) and
an absolute cap set above the current count is just a slower form of the same
lie, so the gate scores the same question the coverage gate scores on the branch
path: "is this getting worse?" -- a non-regression ratchet against a recorded
per-linter baseline.

What this gate does NOT do, deliberately:

* It does not disable, silence, or scope any linter. Every finding golangci-lint
  reports is still reported; `.golangci.yaml` still enables all 13.
* It does not lower a number to go green. Raising the baseline to accommodate
  growth is exactly the coverage gate's forbidden move.

What it DOES do: on the push/whole-tree path it now FAILS the `Lint` job when a
budget is exceeded or has expired. It previously ran with `continue-on-error:
true`, which made it advisory-only -- in run 37155259055 (job 111297271101) it
printed OK at 662/662 errcheck while the debt sat behind a permanently-failing
linter step, so its verdict changed no check result and it could not have
caught growth even had the debt increased. The linter's own exit is remapped to
0 on that path (see the `--issues-exit-code` note in ci.yaml), so this gate is
now the enforced signal rather than a report printed next to a red.

The enforced limit is the measured value plus SLACK=1, so ordinary run-to-run
float noise cannot flip the gate, while any real growth fails. `--report-only`
prints
the verdict and exits 0 for measurement runs that must not gate.

Every baseline below carries the issue that tracks paying it down and the date it
expires. ESTATE.md requires a linked issue and an expiry for anything that
tolerates existing debt; an expired budget is a hard failure here, so a budget
cannot outlive its own paperwork.

Event routing (mirrors scripts/ci/coverage_gate.py)
--------------------------------------------------
The `Lint` job scores two different populations on two paths:

* push / whole-tree -- the report holds every finding in the repository, so the
  measured counts are comparable to the baselines and the gate is ENFORCED.
* pull_request -- the job runs golangci-lint with `--new-from-rev`, so the report
  holds only findings NEW to the PR. Comparing that against a whole-tree
  baseline would be meaningless (a PR adding one finding would look like the
  repository has one errcheck finding), and enforcing it would double-count the
  same debt the whole-tree path already scores. On that path the gate runs in
  --report-only and PRs are still blocked for *introducing* findings by the
  `Lint` job itself.

The routing lives here, in the decision point, rather than in the workflow YAML,
so it can be asserted by a test against real event payloads.
"""

from __future__ import annotations

import argparse
import json
import os
import sys
from collections import Counter
from pathlib import Path

EXIT_OK = 0
EXIT_FAILED = 1
EXIT_USAGE = 2

PR_EVENT = "pull_request"

# Per-linter debt baselines.
#
# Measured uncapped with golangci-lint v2.13.2 on `develop` at 1cd9af8e3f
# (whole tree, --max-issues-per-linter=0 --max-same-issues=0):
#
#     goconst 926  errcheck 662  staticcheck 61  gosec 51
#     prealloc  13  govet       6  gocritic  2
#
# Only the two linters with real backlog are budgeted. The rest are recorded
# for context and enforced at their measured count, so the gate also catches
# them growing.
#
# Each entry: (measured findings, tracked issue, expiry).
#   - Raise a baseline ONLY by fixing findings in the tree.
#   - Lowering a baseline is a lint regression and must not be done to go green.
#   - A past expiry is a hard failure: renew it deliberately in a follow-up that
#     states why, or fix the debt. Silence is not an option.
BASELINES: dict[str, tuple[int, str, str]] = {
    "goconst": (926, "#1122", "2027-01-15"),
    "errcheck": (662, "#1122", "2027-01-15"),
    "staticcheck": (61, "#1122", "2027-01-15"),
    "gosec": (51, "#1122", "2027-01-15"),
    "prealloc": (13, "#1122", "2027-01-15"),
    "govet": (6, "#1122", "2027-01-15"),
    "gocritic": (2, "#1122", "2027-01-15"),
}

# Slack between the enforced budget and the measured value, so float noise in a
# re-run (an added test file, a refactor) cannot flip a gate on its own.
SLACK = 1

# Budgets past this date fail the gate regardless of the counts.
EXPIRY_GRACE_DAYS = 0


def resolve_baseline(linter: str) -> int:
    """Budget for a linter, overridable so a ratchet bump needs no code edit."""
    raw = os.environ.get(f"LINT_BUDGET_{linter.upper()}", "").strip()
    if not raw:
        return BASELINES[linter][0]
    try:
        return int(raw)
    except ValueError:
        print(
            f"::warning::LINT_BUDGET_{linter.upper()}={raw!r} is not an integer; "
            f"using {BASELINES[linter][0]}",
            file=sys.stderr,
        )
        return BASELINES[linter][0]


def load_findings(report_path: Path) -> Counter:
    """Count findings per linter out of a golangci-lint JSON report.

    Counting by the linter each issue reports as (`FromLinter`), never by what
    the run was invoked with: golangci-lint's `--enable` flag is ADDITIVE to the
    config's `linters.enable`, so a run invoked `--default=none --enable=errcheck`
    still reports every linter `.golangci.yaml` enables. Filtering by flag would
    silently compare the wrong population against the baseline.
    """
    if not report_path.is_file():
        raise FileNotFoundError(report_path)
    with report_path.open(encoding="utf-8") as handle:
        report = json.load(handle)
    issues = report.get("Issues")
    if issues is None:
        raise ValueError(f"no 'Issues' key in {report_path}")
    return Counter(issue.get("FromLinter") for issue in issues)


def evaluate(counts: Counter, report_only: bool, today: str) -> int:
    print("Lint debt budget (per-linter non-regression ratchet)")
    print(f"Total findings reported: {sum(counts.values())}")
    print()

    failures: list[str] = []

    for linter in sorted(BASELINES):
        measured, issue, expiry = BASELINES[linter]
        budget = resolve_baseline(linter)
        allowed = budget + SLACK
        actual = counts.get(linter, 0)

        if today > expiry:
            failures.append(
                f"{linter}: budget EXPIRED on {expiry} (tracked in {issue}). "
                "Renew it deliberately or fix the debt; an expired budget fails."
            )
            status = "EXPIRED"
        elif actual > allowed:
            failures.append(
                f"{linter}: {actual} findings exceeds the {allowed} budget "
                f"(measured {measured}, tracked in {issue}, expires {expiry}). "
                "Fix the new findings or do not raise the baseline to absorb them."
            )
            status = "OVER"
        elif actual <= budget:
            status = "OK"
        else:
            # Inside the slack band: at or below measured, but the budget can be
            # tightened by SLACK. Report it so the ratchet is actionable.
            status = "TIGHTEN"

        note = ""
        if status == "TIGHTEN":
            note = f"  <- budget can drop to {actual}"
        print(
            f"  {linter:<12} {actual:>5} / {allowed:<5} allowed "
            f"(measured {measured}, {issue}, expires {expiry})  {status}{note}"
        )

    # Linters reporting findings that are not budgeted at all.
    unbudgeted = {name: n for name, n in counts.items() if name not in BASELINES}
    for name, n in sorted(unbudgeted.items()):
        print()
        print(
            f"::warning::{name} reported {n} findings but has no recorded "
            "baseline; add one in scripts/ci/lint_budget_gate.py so its debt is "
            "tracked rather than appearing for the first time in a red build."
        )

    print()
    if failures:
        for failure in failures:
            print(f"::error::{failure}")
        if report_only:
            print("\n(--report-only: failures above would fail the gate)")
        return EXIT_OK if report_only else EXIT_FAILED

    print("OK: every budgeted linter is within its non-regression budget")
    return EXIT_OK


def main(argv: list[str]) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("report", help="golangci-lint JSON report to score")
    parser.add_argument(
        "--report-only",
        action="store_true",
        help="print the verdict but always exit 0; forces the reporting path "
        "regardless of GITHUB_EVENT_NAME",
    )
    parser.add_argument(
        "--today",
        default=None,
        help="date to evaluate budget expiry against (YYYY-MM-DD); "
        "defaults to the current UTC date",
    )
    args = parser.parse_args(argv[1:])

    event = os.environ.get("GITHUB_EVENT_NAME", "").strip()
    # A diff-scoped report is not comparable to a whole-tree baseline, so the
    # pull_request path reports rather than gates.
    report_only = args.report_only or event == PR_EVENT
    if report_only and event == PR_EVENT:
        print(f"Event: {event} -- diff-scoped report, not gating")

    if args.today:
        today = args.today
    else:
        import datetime

        today = datetime.datetime.now(datetime.UTC).date().isoformat()

    try:
        counts = load_findings(Path(args.report))
    except (FileNotFoundError, ValueError, json.JSONDecodeError) as exc:
        # Refuse to pass unmeasured. A budget gate that cannot read its input
        # must not report success.
        print(f"::error::cannot read lint report: {exc}", file=sys.stderr)
        return EXIT_FAILED

    return evaluate(counts, report_only, today)


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))