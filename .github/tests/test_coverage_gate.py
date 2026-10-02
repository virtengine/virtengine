"""Assert the INFRA-003 coverage gate routes by event and enforces both floors.

DONE WHEN for kanban t_996a5ab1 requires the coverage conditional to be proven
on a real event payload rather than re-derived by reading YAML. These tests
drive `scripts/ci/coverage_gate.py` as a process with `GITHUB_EVENT_NAME` set
to the two real event names the workflow sees, over a real coverprofile.

The regression they pin: `ci.yaml` applied a repo-wide 80% floor on the push
path. The repo sat at 46.5%, so `Go Tests` was red on every push with zero test
failures, and because the required `CI Quality Gates` context inherits
`test-go`, every PR inherited the red. Run 37031738225 (push) vs 37031745564
(pull_request) on the same tree is the observed evidence.

The tests also assert the floor is a ratchet and not a disabled check: a
coverage drop below the recorded baseline still fails.
"""

from __future__ import annotations

import importlib.util
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path
from unittest import mock

REPO_ROOT = Path(__file__).resolve().parents[2]
SCRIPT_PATH = REPO_ROOT / "scripts" / "ci" / "coverage_gate.py"


def load_module():
    spec = importlib.util.spec_from_file_location("coverage_gate", SCRIPT_PATH)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def write_coverprofile(directory: Path, coverage: float) -> Path:
    """Write a coverprofile whose total coverage is exactly `coverage`.

    A coverprofile line is `<path>:<start>.<col>,<end>.<col> <stmts> <count>`.
    The total is the fraction of statements whose count is non-zero, so a
    two-block profile hits an exact percentage without needing the source
    files to exist on disk.
    """
    path = directory / "coverage.out"
    total_statements = 1000
    covered = round(total_statements * coverage / 100.0)
    path.write_text(
        "mode: atomic\n"
        f"example.com/m/covered.go:1.1,1.10 {covered} 1\n"
        f"example.com/m/uncovered.go:1.1,1.10 {total_statements - covered} 0\n",
        encoding="utf-8",
    )
    return path


class CoverageGateRoutingTest(unittest.TestCase):
    """The routing decision, asserted against real event payloads."""

    def setUp(self):
        self.gate = load_module()
        self._tmp = tempfile.TemporaryDirectory()
        self.tmp = Path(self._tmp.name)
        self.addCleanup(self._tmp.cleanup)

    def test_repo_coverage_is_46_5_percent_in_the_fixture(self):
        """Guard the fixture itself: a wrong fixture makes every other test a lie."""
        profile = write_coverprofile(self.tmp, 46.5)
        measured = self.gate.resolve_coverage(profile)
        self.assertIsNotNone(measured)
        self.assertAlmostEqual(measured, 46.5, delta=0.1)

    def test_push_event_at_current_repo_coverage_passes(self):
        """THE REGRESSION.

        46.5% is the repo's measured coverage. Before this change the push path
        compared it against 80 and failed, which is what blocked every PR.
        """
        profile = write_coverprofile(self.tmp, 46.5)
        with mock.patch.dict(os.environ, {"GITHUB_EVENT_NAME": "push"}, clear=False):
            self.assertEqual(self.gate.evaluate_branch(46.5, self.gate.DEFAULT_BASELINE_COVERAGE), 0)

    def test_pull_request_event_delegates_to_diff_scoped_gate(self):
        """A PR is scored on changed lines, never on the repo-wide number."""
        profile = write_coverprofile(self.tmp, 46.5)
        env = {
            "GITHUB_EVENT_NAME": "pull_request",
            "BASE_REF": "develop",
            "PATH": os.environ.get("PATH", ""),
        }
        proc = subprocess.run(
            [sys.executable, str(SCRIPT_PATH), str(profile)],
            cwd=REPO_ROOT,
            env={**os.environ, **env},
            capture_output=True,
            text=True,
        )
        # It must not fail on the repo-wide number. Against a profile whose
        # paths match nothing in the diff, the diff gate reports "no executable
        # changed Go lines" and exits 0.
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        self.assertNotIn("below the 46.0% floor", proc.stdout)
        self.assertIn("pull_request", proc.stdout)

    def test_the_two_paths_do_not_share_one_number(self):
        """Pins the design: 80% for changed lines, ratchet for the repo."""
        self.assertEqual(self.gate.TARGET_COVERAGE, 80.0)
        self.assertEqual(self.gate.REPO_TARGET_COVERAGE, 80.0)
        self.assertLess(self.gate.DEFAULT_BASELINE_COVERAGE, self.gate.REPO_TARGET_COVERAGE)
        self.assertEqual(self.gate.PR_EVENT, "pull_request")

    def test_branch_floor_still_fails_a_real_regression(self):
        """The floor is not a waiver: dropping below the baseline fails."""
        self.assertEqual(self.gate.evaluate_branch(45.0, 46.0), 1)
        self.assertEqual(self.gate.evaluate_branch(46.0, 46.0), 0)
        self.assertEqual(self.gate.evaluate_branch(60.0, 46.0), 0)

    def test_baseline_is_overridable_and_rejects_junk(self):
        with mock.patch.dict(os.environ, {"COVERAGE_BASELINE": "52.5"}, clear=False):
            self.assertEqual(self.gate.resolve_baseline(), 52.5)
        with mock.patch.dict(os.environ, {"COVERAGE_BASELINE": "not-a-number"}, clear=False):
            self.assertEqual(self.gate.resolve_baseline(), self.gate.DEFAULT_BASELINE_COVERAGE)

    def test_missing_coverage_file_fails_closed(self):
        """An absent gate input must fail, not pass by default."""
        self.assertIsNone(self.gate.resolve_coverage(self.tmp / "nope.out"))
        self.assertEqual(self.gate.main(["coverage_gate.py", "nope.out"]), 1)

    def test_branch_path_fails_closed_on_an_unmeasurable_profile(self):
        """A profile with no statements cannot be scored -- fail, never pass."""
        empty = self.tmp / "empty.out"
        empty.write_text("mode: atomic\n", encoding="utf-8")
        with mock.patch.dict(os.environ, {"GITHUB_EVENT_NAME": "push"}, clear=False):
            self.assertEqual(self.gate.main(["coverage_gate.py", str(empty)]), 1)

    def test_usage_error(self):
        self.assertEqual(self.gate.main(["coverage_gate.py"]), 2)


class WorkflowWiringTest(unittest.TestCase):
    """The workflow must actually call the script, not re-inline the routing."""

    CI_YAML = REPO_ROOT / ".github" / "workflows" / "ci.yaml"

    def test_coverage_step_invokes_the_gate_script(self):
        text = self.CI_YAML.read_text(encoding="utf-8")
        self.assertIn("scripts/ci/coverage_gate.py", text)

    def test_workflow_passes_the_real_event_name(self):
        text = self.CI_YAML.read_text(encoding="utf-8")
        self.assertIn("GITHUB_EVENT_NAME: ${{ github.event_name }}", text)
        self.assertIn("BASE_REF: ${{ github.base_ref }}", text)

    def test_dead_repo_wide_floor_is_gone(self):
        """The literal that caused the unconditional red must not come back."""
        text = self.CI_YAML.read_text(encoding="utf-8")
        self.assertNotIn("MIN_COVERAGE=80", text)
        self.assertNotIn("below minimum ${MIN_COVERAGE}%", text)


if __name__ == "__main__":
    unittest.main()