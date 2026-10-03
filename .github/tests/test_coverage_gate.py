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

`WorkflowWiringTest` additionally pins *this suite's own* wiring, and pins it
against the *parsed workflow step* rather than the file text. Three review
rounds found the difference matters:

* Round 1: the routing step was gated on a Go-path filter, so a PR editing only
  the gate's own files (0 `.go` files, exactly like #1166) computed `run=false`
  and skipped the step that guards the gate.
* Round 2: the "the workflow calls the script" assertion matched the literal
  `scripts/ci/coverage_gate.py` against the *whole file*, and that literal also
  lives in two comments -- so deleting the gate outright left the suite green.
  A guard asserted on file text does not fire when the thing it guards is gone.
* Round 3/4: locating the step is necessary but not sufficient. Every assertion
  was still a *substring* match on its `run` string, and a substring cannot
  tell an invocation from a mention -- `echo "<the whole command>"`, a `#`
  comment, and `if false; then <invocation>; fi` all satisfy it, as does
  `continue-on-error` on the step that runs this suite. The gate was still
  deletable with the suite reporting OK.

So the wiring tests here locate the real step object in
`doc["jobs"]["test-go"]["steps"]` and assert it is *structurally* an execution
of its command: exactly one command, and that command BE the invocation,
whole-line (see `step_problems`). Matching the shape rather than a substring is
what makes the INFRA-003 gate, and this suite's own verdict, undeletable and
un-discardable without these tests going red.
"""

from __future__ import annotations

import importlib.util
import os
import re
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

# --------------------------------------------------------------------------
# What a gate step is allowed to be.
#
# The load-bearing rule (round-4 review defect 5): a gate step must be EXACTLY
# ONE command, and that command must BE the invocation. Substring assertions
# over `run` cannot tell an invocation from a mention, so these match the whole
# command line. The reviewer's four mutations all carried the literal
# `scripts/ci/coverage_gate.py coverage.out` and all left the suite green
# while the gate stopped running:
#
#   echo "python3 scripts/ci/coverage_gate.py coverage.out"   -> first word echo
#   if false; then <invocation>; fi                            -> first word if
#   # <invocation>                                             -> no command at all
#   continue-on-error: true (on the suite's own runner)        -> verdict discarded
#
# An allow-list is the only shape that closes that: there is nothing left for a
# bypass to hide in. The cost is deliberate -- any future change to either
# step's body must edit this file in the same PR, where it is visible in the
# diff instead of being a silent, unreviewed weakening of the gate.
# --------------------------------------------------------------------------

# The gate step: interpreter, then the script, then the measured profile.
GATE_COMMAND_RE = re.compile(
    r"^python3?\s+\S*scripts/ci/coverage_gate\.py\s+\S*coverage\.out$"
)
# The routing step: interpreter, then unittest, then this suite as an argument
# anywhere in the tail. The suite name sits behind `-p` in the real command
# (`unittest discover -s .github/tests -p "test_coverage_gate.py" -v`), so
# requiring it adjacent to `unittest` would reject the workflow as it stands --
# which is exactly how the first draft of this regex failed on its own control.
ROUTING_COMMAND_RE = re.compile(
    r"^python3?\s+-m\s+unittest\b.*test_coverage_gate\.py.*$"
)

# Shell constructs that throw away a command's exit status. Anchored at the end
# of the command so a legitimate continuation is not mistaken for a bypass.
_EXIT_OVERRIDE_TAIL_RE = re.compile(r"(?:;|&&|\|\|)\s*(?:exit\s+0\b|true\b|:)\s*$")
_BARE_EXIT_RE = re.compile(r"^exit\s+[01]$")


def step_commands(run: str) -> list[str]:
    """`run` reduced to the shell commands it actually executes, in order.

    A line whose first non-whitespace character is `#` is a shell comment and a
    blank line is not a command; everything else counts as a command. Line
    continuations are deliberately NOT joined: a command split across several
    lines is exactly the extra surface this function exists to refuse.
    """
    commands = []
    for raw in run.splitlines():
        line = raw.strip()
        if not line or line.startswith("#"):
            continue
        commands.append(line)
    return commands


def step_problems(step: dict, command_re: re.Pattern[str]) -> list[str]:
    """Every reason `step` does not actually run its command. Empty means it does.

    Three structural properties, in order of how much they have caught:

    1. the step is exactly ONE command once comments are stripped;
    2. that one command matches `command_re` whole-line, so the script is
       executed rather than printed, quoted, commented, or parked in an
       unreachable branch;
    3. nothing discards the command's exit status and the step does not set
       `continue-on-error`, so the command's verdict decides the job's.

    Property 3 is defence in depth: with exactly one command, `; exit 0` on the
    same line is the only way to keep (2) and lose the verdict, and it is
    caught. Property 1 is what closed round-4 defect 5.
    """
    run = step.get("run")
    if not isinstance(run, str):
        return [f"the step has no `run:` body ({run!r}); it cannot execute anything"]

    commands = step_commands(run)
    if len(commands) != 1:
        return [
            f"the step must be exactly ONE command, but it has {len(commands)} "
            f"({commands!r}). Every additional line is surface a bypass can "
            "hide in, and an invocation inside a comment or an unreachable "
            "branch enforces nothing (round-4 review defect 5)."
        ]

    command = commands[0]
    problems = []
    if not command_re.match(command):
        problems.append(
            f"the command must BE the invocation, whole-line. Got {command!r}. "
            "A command that merely mentions the gate -- `echo \"...\"`, a quoted "
            "string, `if false; then ...; fi`, or a shell comment -- leaves the "
            "gate unenforced while every substring assertion still passes "
            "(round-4 review defect 5, mutations A/B/C)."
        )
    if _EXIT_OVERRIDE_TAIL_RE.search(command):
        problems.append(
            f"{command!r} forces a successful exit after the gate, discarding "
            "its verdict"
        )
    if "|| true" in run or "set +e" in run:
        problems.append(
            f"{run!r} tolerates or ignores a failing gate (`|| true` / `set +e`)"
        )
    for line in commands:
        if _BARE_EXIT_RE.match(line.strip()):
            problems.append(
                f"a bare `exit` in a gate step overrides the gate's own non-zero "
                f"verdict: {line.strip()!r}"
            )
    if step.get("continue-on-error"):
        problems.append(
            "the step sets continue-on-error, so the command's verdict is "
            "recorded but does not decide the job"
        )
    return problems


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
    """The workflow must actually call the script, not re-inline the routing.

    Round-2 review defect, reproduced on this tree at 2191979db: the old
    `test_coverage_step_invokes_the_gate_script` asserted
    `assertIn("scripts/ci/coverage_gate.py", self.text)` where `self.text` is
    the ENTIRE `ci.yaml`. The literal also appears in two comments (ci.yaml:13
    and the INFRA-003 comment above the step), so replacing the step body with
    `echo ...; exit 0` left all 16 tests green -- the gate was deletable from CI
    with the suite reporting OK, and `actionlint` does not catch it either.

    Everything below asserts on the PARSED step object, so the comment can no
    longer stand in for the code. Each assertion names the mutation it catches.
    """

    CI_YAML = REPO_ROOT / ".github" / "workflows" / "ci.yaml"

    # The coverage gate step, identified by name. A step that is renamed, or
    # whose `run` no longer invokes the script, fails the lookup below rather
    # than silently matching a different step.
    COVERAGE_STEP_NAME = "Check coverage threshold"

    @classmethod
    def setUpClass(cls):
        cls.text = cls.CI_YAML.read_text(encoding="utf-8")
        cls.doc = yaml.safe_load(cls.text)
        cls.steps = cls.doc["jobs"]["test-go"]["steps"]

    def find_coverage_step(self):
        """The single parsed step that runs the gate. Fails if it is gone.

        Located by name so the assertions below cannot be satisfied by some
        other step's text, and required to be exactly one so the gate cannot be
        duplicated into a second, unenforced copy.
        """
        matches = [
            s for s in self.steps
            if isinstance(s, dict) and s.get("name") == self.COVERAGE_STEP_NAME
        ]
        self.assertEqual(
            len(matches), 1,
            f"expected exactly one step named {self.COVERAGE_STEP_NAME!r} in "
            f"jobs.test-go, found {len(matches)}: the INFRA-003 gate step must "
            "exist and must not be duplicated",
        )
        return matches[0]

    def assert_step_runs_its_command(self, step, command_re, what):
        problems = step_problems(step, command_re)
        self.assertEqual(
            problems, [],
            f"{what} must be a single command that executes it:\n  - "
            + "\n  - ".join(problems),
        )

    def test_coverage_step_actually_executes_the_gate(self):
        """DEFECT 5, the load-bearing assertion.

        Round-3 review proved the parsed-step assertions real, then round-4
        review showed the residual: every one of them was a substring match on
        `run`, so an invocation that was *visible but dead* satisfied them all.
        Reproduced on this tree at e9439fc0a7 -- all four left the suite green
        with the gate no longer running:

            echo "python3 scripts/ci/coverage_gate.py coverage.out"
            if false; then python3 scripts/ci/coverage_gate.py ...; fi
            # python3 scripts/ci/coverage_gate.py coverage.out
            continue-on-error: true   # on the routing (suite) step

        The fix is an allow-list rather than a longer denylist: the step must be
        exactly one command, and that command must BE the invocation, matched
        whole-line. There is no surface left for a mention to hide in.
        """
        step = self.find_coverage_step()
        self.assert_step_runs_its_command(
            step, GATE_COMMAND_RE,
            "the coverage gate step (it decides whether the job passes)",
        )

    def test_routing_step_actually_executes_the_suite_and_is_fatal(self):
        """DEFECT 5's mutation E: the suite's own verdict must reach CI.

        `test_coverage_step_cannot_bypass_the_gate` (round 3) checked
        `continue-on-error` on the *coverage* step only. Adding it to the
        *routing* step makes the 19 tests' verdict non-fatal: CI records the
        failure and carries on, and nothing in the repo fails while the gate it
        was written to protect is disabled. Same allow-list shape, applied to
        the step that runs this file.
        """
        routing = [
            s for s in self.steps
            if isinstance(s, dict) and "test_coverage_gate.py" in str(s.get("run", ""))
        ]
        self.assertEqual(
            len(routing), 1,
            f"expected exactly one routing-test step, found {len(routing)}",
        )
        self.assert_step_runs_its_command(
            routing[0], ROUTING_COMMAND_RE,
            "the coverage gate routing step (it runs this suite)",
        )

    def test_the_allow_list_itself_rejects_the_round_4_mutations(self):
        """Pins defect 5's regression *in the suite*, not only in my sweep.

        Without this, `step_problems` could be weakened the same way the old
        assertions were -- one relaxed regex at a time, each still green on the
        real workflow -- and the mutation sweep would only catch it on the next
        review round. Here each bypass shape is fed to the checker directly, so
        the checker's own contract is under test.
        """
        real_gate = self.find_coverage_step()
        real_routing = next(
            s for s in self.steps
            if isinstance(s, dict) and "test_coverage_gate.py" in str(s.get("run", ""))
        )
        # The control: the real steps are clean, so the checks below cannot pass
        # vacuously by flagging everything.
        self.assertEqual(step_problems(real_gate, GATE_COMMAND_RE), [])
        self.assertEqual(step_problems(real_routing, ROUTING_COMMAND_RE), [])

        invocation = "python3 scripts/ci/coverage_gate.py coverage.out"
        bypasses = {
            "echoed": {"run": f'echo "{invocation}"'},
            "commented": {"run": f"# {invocation}"},
            "false branch": {"run": f"if false; then {invocation}; fi"},
            "quoted": {"run": f'echo "{invocation}" > /dev/null; true'},
            "trailing exit": {"run": f"{invocation}; exit 0"},
            "or true": {"run": f"{invocation} || true"},
            "errexit off": {"run": f"set +e\n{invocation}"},
            "bare exit": {"run": f"{invocation}\nexit 0"},
            "preceded by setup": {"run": f"echo setting up\n{invocation}"},
            "no run body": {},
            "continue-on-error": {"run": invocation, "continue-on-error": True},
        }
        for label, step in bypasses.items():
            with self.subTest(bypass=label):
                self.assertNotEqual(
                    step_problems(step, GATE_COMMAND_RE), [],
                    f"the allow-list must reject the {label} bypass",
                )

    def test_coverage_step_env_carries_the_real_event_and_base(self):
        """The routing inputs must come from the event, on the step itself.

        Parsed `env` rather than file text: an `env:` block on some other step,
        or one that stopped being wired, must not satisfy the gate.
        """
        step = self.find_coverage_step()
        env = step.get("env") or {}
        self.assertEqual(
            env.get("GITHUB_EVENT_NAME"), "${{ github.event_name }}",
            "the coverage step must set GITHUB_EVENT_NAME from github.event_name",
        )
        self.assertEqual(
            env.get("BASE_REF"), "${{ github.base_ref }}",
            "the coverage step must set BASE_REF from github.base_ref",
        )

    def test_workflow_passes_the_real_event_name(self):
        """Same guarantee, pinned on the text so a rename of the step cannot
        silently drop the env wiring while the parsed test still passes."""
        self.assertIn("GITHUB_EVENT_NAME: ${{ github.event_name }}", self.text)
        self.assertIn("BASE_REF: ${{ github.base_ref }}", self.text)

    def test_dead_repo_wide_floor_is_gone(self):
        """The literal that caused the unconditional red must not come back.

        Deliberately a *text* assertion over the whole file: a repo-wide floor
        re-inlined anywhere in the workflow (not only in the coverage step) is
        the regression, and catching it anywhere is the point. The parsed-step
        test above covers the opposite failure -- a gate that is not called --
        which text matching cannot see.
        """
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