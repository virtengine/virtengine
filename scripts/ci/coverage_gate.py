#!/usr/bin/env python3
"""INFRA-003 coverage gate decision point.

Why this exists
---------------
`ci.yaml` enforced a repo-wide `MIN_COVERAGE=80` floor on the branch (push)
path while the pull_request path used diff-scoped coverage. The repo sits at
46.5% overall, so the branch path was unconditionally red and the required
`CI Quality Gates` context (which inherits `test-go`) reported a permanent
failure that consumed the tier-2 develop->main gate's fix-attempt budget on
every run. There were zero test failures: the only error was the coverage floor
itself.

Scope note, so this is not overclaimed: no PR was ever blocked *by GitHub*.
`develop` has no branch protection (`404 Branch not protected`), `main`'s
protection carries no `required_status_checks`, both rulesets are
`enforcement: disabled`, and tier-1 merges evaluate no checks at all. The defect
was a permanently red signal, not a merge block.

Design
------
The two paths answer two different questions and must not share one number:

* pull_request -- "is the code this PR adds covered?" That is a diff-scoped
  question, so it is scored against the 80% target on changed executable lines
  (delegated to `check_pr_diff_coverage.py`).
* push / branch -- "is the repository as a whole getting worse?" That cannot be
  answered by an absolute target the repository has never met, so it is scored
  as a non-regression ratchet against a recorded baseline. The ratchet is
  strictly stronger than "no gate": coverage below the baseline fails.

The 80% number is NOT lowered or waived. It stays the enforced target for
changed lines, the reported goal on the branch path, and Codecov's project
target (`codecov.yml`). Convergence is by ratchet: raising
`COVERAGE_BASELINE` as coverage improves. That is the mechanism that makes the
floor reachable, where lowering the number would only have hidden the gap.

This module is the single decision point so the routing can be asserted by a
test against real event payloads, rather than re-derived by reading YAML.
"""

from __future__ import annotations

import math
import os
import subprocess
import sys
from pathlib import Path

# INFRA-003: the enforced target for code a change adds.
TARGET_COVERAGE = 80.0

# INFRA-003: the whole-repo target, reported (not enforced) on the branch path.
# Kept identical to the constant in check_pr_diff_coverage.py so the two paths
# cannot drift apart silently.
REPO_TARGET_COVERAGE = 80.0

# Measured repo-wide coverage on `develop` at 9804a5010b, from the `Go Tests`
# job log of run 37031738225: `Total coverage: 46.5%`. The branch floor is
# pinned just below the measurement so ordinary run-to-run float noise cannot
# flip the gate, while any real regression past 0.5 points fails.
#
# Ratchet policy: raise this as coverage improves. Lowering it is a coverage
# regression and must not be done to make a run green.
DEFAULT_BASELINE_COVERAGE = 46.0

PR_EVENT = "pull_request"

# GitHub renders a `::error::` workflow command wherever it appears in a step's
# stdout, independently of that step's exit status. The routing suite
# (`.github/tests/test_coverage_gate.py`) must prove the floor BITES, so it calls
# the scorer directly with below-floor values and asserts the return code. That
# call used to print the real `::error::` line into the `Go Tests` log, so a
# perfectly GREEN job carried an error annotation reading
#
#   ::error::Repository coverage regressed to 45.0%, below the 46.0% floor
#
# Nothing had regressed -- 45.0 was the suite's own fixture -- but the log said
# otherwise, and that line was read as a real coverage failure on
# virtengine/virtengine (kanban t_93bf2133, run 37268464612) and sent two agents
# chasing a ratchet regression that does not exist. A gate whose negative
# controls announce themselves as live failures is a gate whose log cannot be
# trusted to tell you whether it is enforcing.
#
# `ANNOTATION_ERROR` is the real enforcement prefix; the suite redirects its
# fixture verdicts to a throwaway buffer via `evaluate_branch(..., stream=)`, so
# the only thing that reaches a CI log is a verdict that actually happened.
# Keeping the two paths on one prefix constant is what stops this drifting back:
# a second hardcoded literal in the suite is what let it happen.
ANNOTATION_ERROR = "::error::"
ANNOTATION_PREFIX = ANNOTATION_ERROR

EXIT_OK = 0
EXIT_FAILED = 1
EXIT_USAGE = 2


def resolve_baseline() -> float:
    """Branch floor. Overridable so a ratchet bump needs no code edit.

    The override exists to let the floor be RAISED as coverage improves. It
    must never be able to remove it, and the rule below is the invariant that
    makes that true rather than a promise.

    The obvious rule -- refuse `value <= 0.0` -- is necessary and NOT
    sufficient, and this module shipped that version. Two measured holes:

      * `0.0001` and `1e-300` are POSITIVE, so they passed the positivity check
        and then made `coverage < baseline` false for every real measurement.
        The ratchet was off, at a floor of 0.0001%.
      * `nan` is worse, because every comparison against it is false. `nan`
        passes `<= 0.0` and would pass a bare positivity check too, and then
        `coverage < nan` is false for EVERY coverage value -- including 0.0 --
        so the gate cannot fire at all. Its log line reads
        `OK: Repository coverage 45.0% holds the nan% non-regression floor`.

    Both were measured on `origin/develop` (8aa3adc578) against a profile
    measuring 45.0%, i.e. below the 46.0% floor, driven as a process exactly
    as `ci.yaml` drives it: `rc=0`, no `::error::`, ratchet off.

    So the accepted rule is the invariant itself -- the override may only RAISE
    the floor above the recorded default -- checked as `isfinite` first (NaN and
    +/-inf are refused there, and no comparison can be trusted to catch them)
    and the strict raise second. One rule; every degenerate value refused;
    `inf` refused too, which the raise-only rule alone would not do, since an
    infinite floor makes the gate permanently red and bricks the pipeline.

    Rejecting a LOWER floor is a behaviour change from the raw `float(raw)`
    that predate this. Nothing sets it: `grep -rn COVERAGE_BASELINE` finds one
    comment in `ci.yaml` and no assignment, so no consumer is relying on being
    able to lower it -- and the recorded policy has always been
    "raise `COVERAGE_BASELINE` as coverage improves. Lowering it is a coverage
    regression and must not be done to make a run green."
    """
    raw = os.environ.get("COVERAGE_BASELINE", "").strip()
    if not raw:
        return DEFAULT_BASELINE_COVERAGE
    try:
        value = float(raw)
    except ValueError:
        print(
            f"::warning::COVERAGE_BASELINE={raw!r} is not a number; "
            f"using {DEFAULT_BASELINE_COVERAGE}",
            file=sys.stderr,
        )
        return DEFAULT_BASELINE_COVERAGE
    if not math.isfinite(value) or value <= DEFAULT_BASELINE_COVERAGE:
        print(
            f"::warning::COVERAGE_BASELINE={raw!r} does not raise the "
            f"non-regression floor above the recorded default "
            f"{DEFAULT_BASELINE_COVERAGE}; using {DEFAULT_BASELINE_COVERAGE}. "
            "The override may only RAISE the floor as coverage improves -- "
            "never lower it, switch it off, or saturate it.",
            file=sys.stderr,
        )
        return DEFAULT_BASELINE_COVERAGE
    return value


def resolve_event_name() -> str:
    return os.environ.get("GITHUB_EVENT_NAME", "").strip()


def resolve_coverage(coverage_file: Path) -> float | None:
    """Read the total coverage percentage out of a coverprofile.

    Computed directly from the profile's `(statements, count)` columns: the
    percentage of statements whose count is greater than zero. That is the
    number `go tool cover -func` prints in its `total:` line, derived the same
    way `codecov` and the previous CI `awk` pipeline derived it.

    `go tool cover -func` is deliberately not used: its per-function numbers
    come from re-parsing the source file, so it needs the covered sources to
    be present at the profile's recorded paths and silently reports 0.0% when
    they are not. A gate must not depend on that.
    """
    if not coverage_file.is_file():
        return None

    total = 0
    covered = 0
    try:
        with coverage_file.open("r", encoding="utf-8") as handle:
            next(handle, None)  # "mode: atomic"
            for raw in handle:
                line = raw.strip()
                if not line:
                    continue
                try:
                    _location, num_statements, count = line.rsplit(" ", 2)
                    statements = int(num_statements)
                    hits = int(count)
                except ValueError:
                    continue
                total += statements
                if hits > 0:
                    covered += statements
    except OSError:
        return None

    if total == 0:
        return None
    return covered * 100.0 / total


def run_pr_diff_gate(coverage_file: Path, repo_root: Path) -> int:
    """Score only the lines this PR added, against the 80% target."""
    script = repo_root / "scripts" / "ci" / "check_pr_diff_coverage.py"
    proc = subprocess.run(
        [sys.executable, str(script), str(coverage_file)],
        cwd=repo_root,
    )
    return proc.returncode


def evaluate_branch(coverage: float, baseline: float, stream=None) -> int:
    """Non-regression ratchet for the whole repository.

    `stream` defaults to stdout. It exists so a caller that is exercising the
    gate's decision logic rather than *enforcing* it -- the routing suite's
    negative controls, which assert `evaluate_branch(45.0, 46.0) == 1` -- can
    send the verdict somewhere GitHub does not render. See `ANNOTATION_PREFIX`.
    """
    out = stream if stream is not None else sys.stdout
    print(f"Total coverage: {coverage:.1f}%", file=out)
    print(f"Repository target (INFRA-003, informational): {REPO_TARGET_COVERAGE:.0f}%", file=out)
    print(f"Enforced non-regression floor: {baseline:.1f}%", file=out)

    if coverage < REPO_TARGET_COVERAGE:
        gap = REPO_TARGET_COVERAGE - coverage
        print(
            f"::warning::Repository coverage {coverage:.1f}% is {gap:.1f} points "
            f"below the {REPO_TARGET_COVERAGE:.0f}% INFRA-003 target. The floor is "
            "enforced as a non-regression ratchet; raise COVERAGE_BASELINE as "
            "coverage improves.",
            file=out,
        )

    if coverage < baseline:
        print(
            f"{ANNOTATION_PREFIX}Repository coverage regressed to {coverage:.1f}%, "
            f"below the {baseline:.1f}% floor",
            file=out,
        )
        return EXIT_FAILED

    print(
        f"OK: Repository coverage {coverage:.1f}% holds the {baseline:.1f}% "
        "non-regression floor"
    )
    return EXIT_OK


def main(argv: list[str], repo_root: Path | None = None) -> int:
    if len(argv) != 2:
        print("Usage: coverage_gate.py <coverage-file>", file=sys.stderr)
        return EXIT_USAGE

    root = repo_root or Path(__file__).resolve().parents[2]
    coverage_file = Path(argv[1])

    if not coverage_file.is_file():
        print(
            f"{ANNOTATION_PREFIX}Coverage file not found: {coverage_file}",
            file=sys.stderr,
        )
        return EXIT_FAILED

    event = resolve_event_name()

    if event == PR_EVENT:
        # The diff-scoped gate parses the profile itself; it does not need a
        # repo-wide total, so do not fail here on a measurement this path
        # never consults.
        print(f"Event: {PR_EVENT} -- scoring changed Go lines at {TARGET_COVERAGE:.0f}%")
        return run_pr_diff_gate(coverage_file, root)

    coverage = resolve_coverage(coverage_file)
    if coverage is None:
        print(
            f"{ANNOTATION_PREFIX}Could not measure total coverage from "
            f"{coverage_file}; refusing to pass the gate unmeasured",
            file=sys.stderr,
        )
        return EXIT_FAILED

    print(f"Event: {event or '(unset)'} -- scoring repository coverage")
    return evaluate_branch(coverage, resolve_baseline())


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))