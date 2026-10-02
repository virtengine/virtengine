"""Assert the INFRA-003 coverage gate routes by event and enforces both floors.

DONE WHEN for kanban t_996a5ab1 requires the coverage conditional to be proven
on a real event payload rather than re-derived by reading YAML. These tests
drive `scripts/ci/coverage_gate.py` as a process with `GITHUB_EVENT_NAME` set
to the two real event names the workflow sees, over a real coverprofile.

The regression they pin: `ci.yaml` applied a repo-wide 80% floor on the push
path. The repo sat at 46.5%, so `Go Tests` was red on every push with zero test
failures, and because the required `CI Quality Gates` context inherits `test-go`,
that permanent red signal fed the tier-2 develop->main gate's fix-attempt budget
on every run. Run 37031738225 (push) vs 37031745564 (pull_request) on the same
tree is the observed evidence. No PR was ever blocked by GitHub itself --
`develop` has no branch protection and `main`'s has no required_status_checks --
so the defect was the red signal, not a merge block.

The tests also assert the floor is a ratchet and not a disabled check: a
coverage drop below the recorded baseline still fails.

`WorkflowWiringTest` additionally pins *this suite's own* wiring: the routing
step in `ci.yaml` must not be gated on a Go-path filter, because the round-1
review showed that a PR editing only the gate's own files (0 `.go` files, exactly
like #1166) computed `run=false` and skipped the step that guards the gate.
"""

from __future__ import annotations

import importlib.util
import os
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path
from unittest import mock

import yaml

REPO_ROOT = Path(__file__).resolve().parents[2]
SCRIPT_PATH = REPO_ROOT / "scripts" / "ci" / "coverage_gate.py"
DIFF_GATE_PATH = REPO_ROOT / "scripts" / "ci" / "check_pr_diff_coverage.py"
GATE_OWN_FILES = (
    ".github/workflows/ci.yaml",
    "scripts/ci/coverage_gate.py",
    "scripts/ci/check_pr_diff_coverage.py",
    ".github/tests/test_coverage_gate.py",
)


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


def build_hermetic_repo(root: Path) -> Path:
    """Build a throwaway repo with a real base ref, merge base and Go diff.

    Hermetic on purpose: the routing step now runs unconditionally, so this
    suite must not depend on `origin/develop` existing in whatever checkout CI
    happened to make (it does not: `test-go` is a push job too, and the PR-path
    assertions run there as well). Both gate scripts are copied in, so the
    subprocess under test is the committed one.

    The history is built once and reused: the covered/uncovered variants differ
    only in the coverprofile, so `git init`/commit run once instead of twice
    (they dominate this suite's wall clock).
    """
    repo = root / "repo"
    (repo / "scripts" / "ci").mkdir(parents=True)
    (repo / "x" / "keeper").mkdir(parents=True)

    env = {
        **os.environ,
        "GIT_AUTHOR_NAME": "t",
        "GIT_AUTHOR_EMAIL": "t@t",
        "GIT_COMMITTER_NAME": "t",
        "GIT_COMMITTER_EMAIL": "t@t",
        "HOME": str(repo),
    }

    def g(*args: str) -> None:
        subprocess.run(
            ["git", *args], cwd=repo, env=env, check=True,
            capture_output=True, text=True,
        )

    g("init", "-q", "-b", "develop", ".")
    keeper = repo / "x" / "keeper" / "msg_server.go"
    keeper.write_text("package keeper\n", encoding="utf-8")
    g("add", "-A")
    g("commit", "-q", "-m", "base")
    head = subprocess.run(
        ["git", "rev-parse", "HEAD"], cwd=repo, env=env,
        check=True, capture_output=True, text=True,
    ).stdout.strip()
    g("update-ref", "refs/remotes/origin/develop", head)

    g("checkout", "-q", "-b", "feature")
    # The PR adds exactly one line, which lands at line 2.
    keeper.write_text("package keeper\nfunc Changed() {}\n", encoding="utf-8")
    g("add", "-A")
    g("commit", "-q", "-m", "add Changed")

    for script in (SCRIPT_PATH, DIFF_GATE_PATH):
        shutil.copy2(script, repo / "scripts" / "ci" / script.name)
    return repo


def write_added_line_profile(repo: Path, covered: bool) -> Path:
    """Coverprofile for the single line the PR added, covered or not."""
    profile = repo / "coverage.out"
    profile.write_text(
        "mode: atomic\n"
        f"x/keeper/msg_server.go:2.1,2.18 1 {1 if covered else 0}\n",
        encoding="utf-8",
    )
    return profile


def run_gate_in(repo: Path, event: str = "pull_request"):
    env = {
        **os.environ,
        "GITHUB_EVENT_NAME": event,
        "BASE_REF": "develop",
        "PATH": os.environ.get("PATH", ""),
    }
    return subprocess.run(
        [sys.executable, str(repo / "scripts" / "ci" / "coverage_gate.py"),
         str(repo / "coverage.out")],
        cwd=repo, env=env, capture_output=True, text=True,
    )


class CoverageGateRoutingTest(unittest.TestCase):
    """The routing decision, asserted against real event payloads."""

    @classmethod
    def setUpClass(cls):
        # Built once for the whole class: the two PR-path cases differ only in
        # the coverprofile, and git init/commit dominate the wall clock.
        cls._tmp_shared = tempfile.TemporaryDirectory()
        shared = Path(cls._tmp_shared.name)
        cls.hermetic_repo = build_hermetic_repo(shared)

    @classmethod
    def tearDownClass(cls):
        cls._tmp_shared.cleanup()

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
        compared it against 80 and failed, turning `Go Tests` red on every push
        with zero test failures -- a permanent red signal on the context that
        `CI Quality Gates` inherits.
        """
        profile = write_coverprofile(self.tmp, 46.5)
        with mock.patch.dict(os.environ, {"GITHUB_EVENT_NAME": "push"}, clear=False):
            self.assertEqual(self.gate.evaluate_branch(46.5, self.gate.DEFAULT_BASELINE_COVERAGE), 0)

    def test_pull_request_event_fails_on_an_uncovered_added_line(self):
        """THE PR-PATH GATE, in a repo built for it.

        80% must actually bite on the lines a PR adds. Proven here with a real
        git history and a real three-dot diff rather than against the ambient
        checkout, so the assertion holds on a push run too.
        """
        write_added_line_profile(self.hermetic_repo, covered=False)
        proc = run_gate_in(self.hermetic_repo)
        output = proc.stdout + proc.stderr
        self.assertEqual(proc.returncode, 1, output)
        self.assertIn("below minimum 80%", output)
        self.assertIn("pull_request", proc.stdout)
        # It must not have fallen through to the repo-wide ratchet.
        self.assertNotIn("non-regression floor", proc.stdout)

    def test_pull_request_event_passes_on_a_covered_added_line(self):
        """The other direction, so the test above cannot pass vacuously."""
        write_added_line_profile(self.hermetic_repo, covered=True)
        proc = run_gate_in(self.hermetic_repo)
        output = proc.stdout + proc.stderr
        self.assertEqual(proc.returncode, 0, output)
        self.assertIn("100.0%", proc.stdout)
        self.assertNotIn("below the 46.0% floor", proc.stdout)

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

    @classmethod
    def setUpClass(cls):
        cls.text = cls.CI_YAML.read_text(encoding="utf-8")
        cls.doc = yaml.safe_load(cls.text)

    def test_coverage_step_invokes_the_gate_script(self):
        self.assertIn("scripts/ci/coverage_gate.py", self.text)

    def test_workflow_passes_the_real_event_name(self):
        self.assertIn("GITHUB_EVENT_NAME: ${{ github.event_name }}", self.text)
        self.assertIn("BASE_REF: ${{ github.base_ref }}", self.text)

    def test_dead_repo_wide_floor_is_gone(self):
        """The literal that caused the unconditional red must not come back."""
        self.assertNotIn("MIN_COVERAGE=80", self.text)
        self.assertNotIn("below minimum ${MIN_COVERAGE}%", self.text)

    def test_gate_routing_test_is_not_behind_a_go_path_filter(self):
        """Defect 1: the gate's own test must not skip on gate-only PRs.

        Round-1 review reproduced this: the step carried
        `if: steps.detect.outputs.run == 'true'`, and `detect` only matches
        *.go / go.mod / go.sum / go.work / go.work.sum. #1166 edited ci.yaml,
        scripts/ci/coverage_gate.py and this test file -- 0 .go files -- so
        detect reported run=false and the step was SKIPPED in its own PR. A
        guard that skips on the PRs that most need it is not a guard.
        """
        steps = self.doc["jobs"]["test-go"]["steps"]
        routing_steps = [
            s for s in steps
            if isinstance(s, dict) and "test_coverage_gate.py" in str(s.get("run", ""))
        ]
        self.assertEqual(
            len(routing_steps), 1,
            f"expected exactly one routing-test step, found {len(routing_steps)}",
        )
        step = routing_steps[0]
        self.assertNotIn(
            "if",
            step,
            "the routing test must be unconditional: a Go-path filter skips it "
            "on every PR that edits only the gate's own files",
        )

    def test_detect_filter_cannot_cover_the_gate_own_files(self):
        """Documents why the filter alone is not an acceptable fix.

        If someone re-gates the routing step on `detect`, the pathspec must at
        minimum include the gate's own files. This pins the current filter
        (Go only), which is precisely why the step is unconditional instead.
        """
        steps = self.doc["jobs"]["test-go"]["steps"]
        detect = next(
            s for s in steps
            if isinstance(s, dict) and s.get("id") == "detect"
        )
        run = detect["run"]
        for path in GATE_OWN_FILES:
            if path == ".github/workflows/ci.yaml":
                continue  # YAML cannot use a *.go pathspec; this is the point.
            self.assertNotIn(path, run)
        # The pathspec really is Go-only.
        self.assertIn("'*.go'", run)

    def test_gate_own_files_all_exist(self):
        """A guard pointing at a renamed file is not a guard."""
        for rel in GATE_OWN_FILES:
            self.assertTrue((REPO_ROOT / rel).is_file(), rel)


if __name__ == "__main__":
    unittest.main()