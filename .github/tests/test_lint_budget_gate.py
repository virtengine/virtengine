#!/usr/bin/env python3
"""Tests for the per-linter lint budget gate.

These assert the gate's DECISION LOGIC against real golangci-lint JSON, and are
falsified against the actual failure modes the gate exists to prevent:

* a linter that grows past its budget must fail (the default 50-cap allowed
  662 findings to accumulate unnoticed, so silent growth is the real risk);
* an EXPIRED budget must fail on its own, independent of counts, so debt cannot
  outlive its own paperwork (ESTATE.md requires an expiry);
* a report that cannot be read must NOT pass -- a budget gate that reports
  success without measuring is worse than no gate;
* raising the budget must be possible without a code edit (the ratchet has to be
  payable), and lowering it must not be silently accepted;
* a linter with no recorded baseline must be surfaced, not ignored.

Run: python -m unittest discover -s .github/tests -p "test_lint_budget_gate.py" -v
"""

from __future__ import annotations

import json
import os
import subprocess
import sys
import tempfile
import unittest
import unittest.mock
from collections import Counter
from pathlib import Path

import yaml

REPO_ROOT = Path(__file__).resolve().parents[2]
GATE = REPO_ROOT / "scripts" / "ci" / "lint_budget_gate.py"

sys.path.insert(0, str(REPO_ROOT / "scripts" / "ci"))

import lint_budget_gate as gate  # noqa: E402


def make_report(counts: dict[str, int]) -> str:
    """A golangci-lint JSON report with N findings per linter."""
    issues = [
        {
            "FromLinter": linter,
            "Text": "Error return value of `f` is not checked",
            "Pos": {"Filename": "pkg/x/y.go", "Line": 1},
        }
        for linter, n in counts.items()
        for _ in range(n)
    ]
    path = Path(tempfile.mkdtemp()) / "report.json"
    path.write_text(json.dumps({"Issues": issues}), encoding="utf-8")
    return str(path)


def run_gate(*args: str, env: dict[str, str] | None = None) -> tuple[int, str]:
    proc = subprocess.run(
        [sys.executable, str(GATE), *args],
        capture_output=True,
        text=True,
        env={**os.environ, **(env or {})},
    )
    return proc.returncode, proc.stdout + proc.stderr


def load_enabled_linters() -> list[str]:
    """The linters `.golangci.yaml` enables, read from the config itself.

    Parsed rather than hardcoded: the test is meant to fail when a linter is
    ADDED to the config, so a literal copy of the list would pass forever and
    stop guarding anything.
    """
    config = yaml.safe_load((REPO_ROOT / ".golangci.yaml").read_text(encoding="utf-8"))
    return list(config["linters"]["enable"])


class LoadFindingsTest(unittest.TestCase):
    def test_counts_by_reported_linter_not_by_flags(self):
        # The bug this guards: `--enable` is additive to the config, so a run
        # invoked for one linter reports all 13. Counting by flag would compare
        # the wrong population to the baseline.
        report = make_report({"errcheck": 662, "goconst": 926})
        counts = gate.load_findings(Path(report))
        self.assertEqual(counts["errcheck"], 662)
        self.assertEqual(counts["goconst"], 926)

    def test_missing_file_raises(self):
        with self.assertRaises(FileNotFoundError):
            gate.load_findings(Path("/nonexistent/report.json"))

    def test_report_without_issues_key_raises(self):
        path = Path(tempfile.mkdtemp()) / "bad.json"
        path.write_text(json.dumps({"nope": []}), encoding="utf-8")
        with self.assertRaises(ValueError):
            gate.load_findings(path)


class BaselineTests(unittest.TestCase):
    def test_every_enabled_linter_has_a_baseline(self):
        # The ratchet must cover every linter `.golangci.yaml` enables.
        #
        # An UNBUDGETED linter falls into the `unbudgeted` branch, which emits a
        # `::warning::` and returns EXIT_OK. On the push path the linter's own
        # exit is remapped to 0 by `--issues-exit-code=0`, so nothing else can
        # catch growth there: findings in an unbudgeted linter are accepted
        # silently. That is how six enabled linters (copyloopvar, errchkjson,
        # ineffassign, misspell, unparam, unused) came to be unbounded while the
        # ratchet read as complete. Asserting the config and the gate agree is
        # the check that makes a linter added to the config later fail HERE
        # rather than silently appearing untracked in a build.
        enabled = load_enabled_linters()
        self.assertTrue(enabled, "no linters parsed from .golangci.yaml")
        unbudgeted = sorted(set(enabled) - set(gate.BASELINES))
        self.assertEqual(
            unbudgeted, [],
            f"linters enabled in .golangci.yaml with no baseline in "
            f"lint_budget_gate.py: {unbudgeted}. On the push path an unbudgeted "
            f"linter only warns, so its debt can grow unbounded. Budget it at its "
            f"measured count.",
        )

    def test_every_baseline_has_issue_and_expiry(self):
        # ESTATE.md: a tolerated debt needs a linked issue and an expiry.
        for linter, (_measured, issue, expiry) in gate.BASELINES.items():
            with self.subTest(linter=linter):
                self.assertTrue(issue.startswith("#"), f"{linter} has no issue")
                self.assertRegex(expiry, r"^\d{4}-\d{2}-\d{2}$")

    def test_env_override_changes_budget_without_code_edit(self):
        self.assertEqual(gate.resolve_baseline("errcheck"), gate.BASELINES["errcheck"][0])
        with unittest.mock.patch.dict(os.environ, {"LINT_BUDGET_ERRCHECK": "10"}):
            self.assertEqual(gate.resolve_baseline("errcheck"), 10)

    def test_non_numeric_override_warns_and_falls_back(self):
        with unittest.mock.patch.dict(os.environ, {"LINT_BUDGET_ERRCHECK": "abc"}):
            self.assertEqual(
                gate.resolve_baseline("errcheck"), gate.BASELINES["errcheck"][0]
            )


class EvaluateTests(unittest.TestCase):
    def _counts(self, **overrides):
        base = {name: measured for name, (measured, _i, _e) in gate.BASELINES.items()}
        base.update(overrides)
        return Counter(base)

    def test_at_baseline_passes(self):
        self.assertEqual(gate.evaluate(self._counts(), False, "2026-10-03"), gate.EXIT_OK)

    def test_one_finding_over_budget_fails(self):
        # Slack is exactly SLACK findings, so measured+SLACK+1 must fail.
        counts = self._counts(errcheck=gate.BASELINES["errcheck"][0] + gate.SLACK + 1)
        self.assertEqual(gate.evaluate(counts, False, "2026-10-03"), gate.EXIT_FAILED)

    def test_within_slack_passes_and_asks_to_tighten(self):
        counts = self._counts(errcheck=gate.BASELINES["errcheck"][0] + gate.SLACK)
        self.assertEqual(gate.evaluate(counts, False, "2026-10-03"), gate.EXIT_OK)

    def test_reduced_debt_passes(self):
        counts = self._counts(errcheck=gate.BASELINES["errcheck"][0] - 100)
        self.assertEqual(gate.evaluate(counts, False, "2026-10-03"), gate.EXIT_OK)

    def test_expired_budget_fails_even_when_counts_are_fine(self):
        # Independent of counts: debt must not outlive its paperwork.
        code = gate.evaluate(self._counts(), False, "2030-01-01")
        self.assertEqual(code, gate.EXIT_FAILED)

    def test_report_only_exits_zero_on_failure(self):
        counts = self._counts(errcheck=99999)
        self.assertEqual(gate.evaluate(counts, True, "2026-10-03"), gate.EXIT_OK)

    def test_unbudgeted_linter_is_warned_not_ignored(self):
        counts = self._counts()
        counts["brandnewlinter"] = 7
        code = gate.evaluate(counts, False, "2026-10-03")
        self.assertEqual(code, gate.EXIT_OK)


class CliTests(unittest.TestCase):
    def test_real_report_at_baseline_passes(self):
        # Proves the gate accepts a genuinely-shaped report.
        counts = {name: measured for name, (measured, _i, _e) in gate.BASELINES.items()}
        code, out = run_gate(make_report(counts), "--today", "2026-10-03")
        self.assertEqual(code, 0, out)
        self.assertIn("within its non-regression budget", out)

    def test_unreadable_report_does_not_pass(self):
        code, out = run_gate("/nonexistent/report.json")
        self.assertEqual(code, 1)
        self.assertIn("cannot read lint report", out)

    def test_missing_report_arg_is_usage_error(self):
        code, _out = run_gate()
        self.assertEqual(code, 2)

    def test_expired_budget_fails_through_cli(self):
        counts = {name: measured for name, (measured, _i, _e) in gate.BASELINES.items()}
        # Pin the event explicitly. `run_gate` inherits os.environ, so on a
        # pull_request run GITHUB_EVENT_NAME would force report_only and the
        # gate would exit 0 -- this assertion would then fail on CI only.
        code, out = run_gate(
            make_report(counts), "--today", "2030-01-01",
            env={"GITHUB_EVENT_NAME": "push"},
        )
        self.assertEqual(code, 1, out)
        self.assertIn("EXPIRED", out)


class EventRoutingTests(unittest.TestCase):
    """The routing that mirrors scripts/ci/coverage_gate.py.

    A diff-scoped report (`--new-from-rev`, which is what the pull_request path
    produces) must NOT be compared against a whole-tree baseline: a PR adding a
    single finding would otherwise look like the whole repository has one
    errcheck finding. These assert the routing against real event payloads
    rather than by reading the workflow YAML.
    """

    def _grown_report(self) -> str:
        over = gate.BASELINES["errcheck"][0] + gate.SLACK + 1
        return make_report({**{
            name: measured for name, (measured, _i, _e) in gate.BASELINES.items()
        }, "errcheck": over})

    def test_push_event_enforces(self):
        # Over budget on a whole-tree push must fail.
        code, out = run_gate(
            self._grown_report(), "--today", "2026-10-03",
            env={"GITHUB_EVENT_NAME": "push"},
        )
        self.assertEqual(code, 1, out)
        self.assertNotIn("not gating", out)

    def test_pull_request_event_does_not_gate(self):
        # Same report, PR event: diff-scoped, so it reports instead of failing.
        code, out = run_gate(
            self._grown_report(), "--today", "2026-10-03",
            env={"GITHUB_EVENT_NAME": "pull_request"},
        )
        self.assertEqual(code, 0, out)
        self.assertIn("not gating", out)

    def test_explicit_report_only_overrides_push(self):
        code, out = run_gate(
            self._grown_report(), "--today", "2026-10-03", "--report-only",
            env={"GITHUB_EVENT_NAME": "push"},
        )
        self.assertEqual(code, 0, out)
        self.assertIn("would fail the gate", out)

    def test_pr_event_still_reports_the_overage(self):
        # Reporting must not mean hiding: the over-budget finding is still shown.
        _code, out = run_gate(
            self._grown_report(), "--today", "2026-10-03",
            env={"GITHUB_EVENT_NAME": "pull_request"},
        )
        self.assertIn("errcheck", out)
        self.assertIn("exceeds", out)


if __name__ == "__main__":
    unittest.main()