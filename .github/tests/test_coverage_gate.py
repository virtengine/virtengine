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
* Round 5: the allow-list above hardened *how the command is written* and never
  asked *whether the step runs at all*. `step_problems` read `run` and
  `continue-on-error` but not `step["if"]`, and the committed gate step carries
  one -- so `if: false`, an inverted `if:`, an `if:` narrowed to `pull_request`,
  and `if: always()` all disabled the INFRA-003 gate with 19 green tests. The
  same gap existed on the suite's own inputs (`env: COVERAGE_BASELINE: "0"`
  silently zeroes the ratchet) and on the routing step's name. See
  `gate_input_problems` and `job_problems`.
* Round 6: round 5's pin fixed the gate step's `if:` but asked nothing of the
  step that PRODUCES its input. Narrowing the producer's `if:` (with
  `&& github.event_name == 'push'`, or with a `hashFiles(...)` pathspec) or
  giving it `continue-on-error: true` left all 25 tests green while the gate
  scored a stale profile on the runs where the producer was skipped. The pin is
  now two-sided: the producer must run at least as often as the gate and its
  failure must be fatal. Measured, with the mutations named in the harness.

So the wiring tests here locate the real step object in
`doc["jobs"]["test-go"]["steps"]` and assert it is *structurally* an execution
of its command: exactly one command, and that command BE the invocation,
whole-line (see `step_problems`). Matching the shape rather than a substring is
what makes the INFRA-003 gate, and this suite's own verdict, undeletable and
un-discardable without these tests going red.
"""

from __future__ import annotations

import contextlib
import importlib.util
import io
import os
import re
import shlex
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
# Round-7 review defect: this suite made the gate's own log lie.
#
# Several tests here call the scorer DIRECTLY with below-floor values and assert
# its return code -- `evaluate_branch(45.0, 46.0) == 1`. That is the right way
# to prove the floor bites. But `evaluate_branch` PRINTED its verdict, so a
# GREEN `Go Tests` job shipped this into its log:
#
#   ::error::Repository coverage regressed to 45.0%, below the 46.0% floor
#
# `::error::` is a GitHub workflow command: the runner renders it as a failure
# annotation regardless of the step's exit status. So on run 37268464612 a
# passing job displayed a coverage regression that had not happened, and it was
# taken at face value -- kanban t_93bf2133 was raised to "ci.yaml RED: coverage
# ratchet regressed to 45.0% below the 46.0% floor" and spent chasing a
# regression that does not exist. `main()` has the same problem via its
# fail-closed paths, on stderr.
#
# This is a gate-integrity defect in the same class the rest of this file
# exists to close: an assertion whose own output is indistinguishable from the
# failure it is testing for. A guard must be able to tell you, from the log,
# whether the thing it guards is actually enforcing.
#
# The fix is two-sided, because either half alone is insufficient:
#   1. every direct-call test now runs under `no_gate_annotations()`, so a
#      fixture verdict cannot reach a CI log; and
#   2. `test_the_suite_itself_emits_no_error_annotations` runs THIS SUITE as a
#      subprocess and fails if a single `::error::` reaches its stdout -- so the
#      leak cannot come back through a new call site, in this file or in a
#      future edit to the scorer.
# --------------------------------------------------------------------------

ANNOTATION_ERROR = "::error::"

# Set on the child spawned by `test_the_suite_itself_emits_no_error_annotations`
# so that run does not spawn another. See that test for why it is needed.
SUITE_CHILD_ENV = "VE_COVERAGE_GATE_SUITE_CHILD"


@contextlib.contextmanager
def no_gate_annotations():
    """Capture the gate's stdout AND stderr for the duration of a fixture call.

    Both streams: GitHub renders workflow commands emitted on either, and the
    fail-closed paths in `main()` report to stderr. Capturing stdout alone
    would leave `test_missing_coverage_file_fails_closed` still annotating.
    """
    buf = io.StringIO()
    with contextlib.redirect_stdout(buf), contextlib.redirect_stderr(buf):
        yield buf


# --------------------------------------------------------------------------
# Reading the suite's OWN output the way GitHub reads it.
#
# Round-8 review defect: the subprocess guard rebuilt this suite's argv by
# hand, so it drifted from `ci.yaml`, which runs it with `-v`. Under `-v`
# unittest echoes each test's METHOD docstring, and a docstring whose first
# line is a workflow command therefore reaches the log line-initial -- exactly
# the shape GitHub renders as a failure annotation, and exactly the phantom
# this file exists to kill. The guard was blind to it because the child it
# spawned was not the child CI runs. A guard that does not reproduce the real
# invocation is a guard with a hole in exactly the place it was written to
# close; this is the same defect class as round 7, one level of indirection up.
#
# Two rules follow, and both are enforced below:
#   1. the argv comes from the WORKFLOW, parsed -- never hand-typed, so the
#      guard cannot fall behind `ci.yaml` the way a copy did; and
#   2. the assertion matches LINE-INITIAL commands only. That is the exact set
#      GitHub renders (a command must open the line; it is echoed mid-sentence
#      otherwise), so substring matching would flag this file's own prose --
#      which necessarily quotes the literal it is discussing -- and a guard
#      that false-positives on its own documentation is a guard that gets
#      deleted rather than satisfied.
# --------------------------------------------------------------------------


def workflow_suite_argv(workflow_step_run: str) -> list[str]:
    """The argv `ci.yaml` uses to run this suite, from the step's own text.

    Parsed with `shlex` so quoted arguments survive, then `python`/`python3` is
    replaced with `sys.executable` -- the interpreter identity is irrelevant to
    what the suite PRINTS, and using the running interpreter keeps the child on
    the same environment. A parse failure is raised rather than defaulted: a
    silent fallback would reintroduce the very drift this exists to remove.
    """
    argv = shlex.split(workflow_step_run, posix=True)
    if not argv or Path(argv[0]).name not in ("python", "python3", "python.exe"):
        raise AssertionError(
            f"cannot read the suite's argv out of {workflow_step_run!r}: it does "
            "not start with a python interpreter"
        )
    return [sys.executable, *argv[1:]]


def leaked_annotations(output: str) -> list[str]:
    """Lines GitHub would render as failure annotations.

    Line-initial only, and only the `::error::` command: that is the exact set
    of lines this suite is forbidden to emit. Substring matching would also
    catch `::error::` quoted mid-sentence (this file's docstrings, and
    `test_this_file_declares_no_bare_error_annotation` itself), which GitHub
    does not render and which the suite must be able to discuss.
    """
    return [
        line for line in output.splitlines()
        if line.lstrip().startswith(ANNOTATION_ERROR)
    ]


# The routing step's name, shared by `WorkflowWiringTest` (which parses
# `ci.yaml`) and `GateAnnotationHygieneTest` (which needs the argv that step
# carries). Duplicated here rather than reaching across classes so the two
# cannot disagree about which step they mean.
ROUTING_STEP_NAME = "Test coverage gate routing"


def routing_step_run() -> str:
    """The `run` of the `ci.yaml` step that executes this suite.

    Parsed from the live workflow every call rather than cached, so an edit to
    `ci.yaml` is reflected immediately -- and, more importantly, so a test that
    asserts on the workflow cannot be satisfied by a stale copy of it.
    """
    doc = yaml.safe_load(
        (REPO_ROOT / ".github" / "workflows" / "ci.yaml").read_text(encoding="utf-8")
    )
    matches = [
        step for step in doc["jobs"]["test-go"]["steps"]
        if isinstance(step, dict) and step.get("name") == ROUTING_STEP_NAME
    ]
    if len(matches) != 1:
        raise AssertionError(
            f"expected exactly one step named {ROUTING_STEP_NAME!r} in "
            f"jobs.test-go, found {len(matches)}: the suite's own runner must "
            "exist under that exact name, or the guard below is not inspecting "
            "the invocation CI really uses"
        )
    return str(matches[0]["run"])

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

# --------------------------------------------------------------------------
# Round-5 review defect 6: the allow-list above says nothing about WHETHER the
# step runs, so it reads `run` and `continue-on-error` but never `step["if"]` --
# and the committed gate step carries one. Reproduced on this tree at
# cebdb3d33; all four left the suite green with the INFRA-003 gate
# total-loss disabled:
#
#   if: false                                    -> the gate never runs
#   if: <detect> && github.event_name == 'pull_request'  -> the branch
#                                                   ratchet is silently dead
#   if: steps.detect.outputs.run != 'true'       -> inverted; runs only when
#                                                   the Go tests did NOT
#   if: always()                                  -> runs even with no profile
#
# The two bypasses are opposite in sign, which is what makes this unambiguous:
# the first and last make the gate never run, the second and third make it run
# in the wrong circumstances. All are total loss of enforcement.
#
# The honest form of the assertion is POSITIVE -- pin the exact condition the
# gate is allowed to carry -- not "has no `if:` key", because `if: always()`
# satisfies "no restriction is written" while being weaker than the step that
# produces its input. Any change to the pin must be made here, in the diff,
# where it is reviewable.
# --------------------------------------------------------------------------

# The one condition under which the gate may run: exactly when the step that
# produced `coverage.out` ran. Same expression, same truth value, same event.
#
# The review offered two acceptable forms: make the gate step unconditional, or
# require its `if:` to be no weaker than the producer of its input. This is the
# second, and the first would be actively harmful: on a pull_request that
# touches no Go files, `detect` reports run=false, `go test` never runs, and
# `coverage.out` does not exist -- so an unconditional gate would call
# `coverage_gate.py` with no profile, and `main()` exits 1 on
# "Coverage file not found", turning every non-Go PR red. That is the same
# failure class this gate exists to end. So the pin is equality with the
# producer's condition, which is the strongest honest form.
ALLOWED_GATE_IF = "steps.detect.outputs.run == 'true'"

# The complete set of variables the gate step is allowed to receive, and the
# exact value each must carry. An allow-list of KEYS alone is not enough: a
# hardcoded `GITHUB_EVENT_NAME: push` keeps both keys present and makes every PR
# take the branch path, which is the routing defect this whole file exists to
# prevent. (My own meta-test caught that while I was adding it -- the first
# draft of `gate_input_problems` checked key sets only.)
GATE_ENV_KEYS = ("GITHUB_EVENT_NAME", "BASE_REF")
GATE_ENV_VALUES = {
    "GITHUB_EVENT_NAME": "${{ github.event_name }}",
    "BASE_REF": "${{ github.base_ref }}",
}


def gate_input_problems(
    step: dict,
    producer_run: str,
    producer_condition: str | None,
    producer_fatal: bool,
) -> list[str]:
    """Every reason `step` would not enforce the gate it is wired to enforce.

    `step_problems` answers "does this command run?"; this answers "is it even
    reached, and is it measuring what it claims to?". Five properties:

    1. `if:` is exactly `ALLOWED_GATE_IF` -- absent is NOT accepted, because the
       gate's input (`coverage.out`) is produced by a step carrying that same
       condition, and a gate that runs when its input does not exist fails
       closed on a missing file but silently passes on a STALE one;
    2. `env:` is exactly the two event-derived routing inputs, each carrying its
       exact value, and nothing else. An extra key is an input the gate can
       consume without anyone reasoning about it -- round-5 measurement:
       `env: COVERAGE_BASELINE: "0"` zeroed the enforced non-regression floor
       with all 19 tests green, because `resolve_baseline` prefers the variable
       over the recorded baseline;
    3. the profile the producer writes is the profile the gate is passed;
    4. neither value is hardcoded, which would re-route the gate silently;
    5. the PRODUCER runs at least as often as the gate, and its failure is
       fatal. This is round-6 measurement, and it is the direction the pin in
       (1) alone does not cover: pinning the gate's `if:` to the honest
       condition while leaving the producer free to carry a narrower one means
       the gate scores a stale or absent profile on the runs where the producer
       was skipped. Measured at 25 green tests on this tree, all three of:
           producer `if:` narrowed with `&& event == 'push'`
           producer `if:` narrowed with `&& hashFiles(...) != ''`
           `continue-on-error: true` on the producer
       The first two leave the gate reading last run's `coverage.out`; the last
       lets a failed test run still hand the gate a truncated profile.

    `producer_condition`/`producer_fatal` are required rather than defaulted: a
    checker that can be called without the input it needs is how rounds 3-5
    happened, one assertion at a time, each still green on the real workflow.
    An ABSENT producer `if:` is accepted and means "unconditional", which is
    strictly *wider* than the gate and therefore safe -- the risk runs one way.
    """
    problems = []

    condition = step.get("if")
    if condition != ALLOWED_GATE_IF:
        problems.append(
            f"the gate step's `if:` must be exactly {ALLOWED_GATE_IF!r}, but it is "
            f"{condition!r}. This is round-5 review defect 6: the step that "
            "produced coverage.out carries that condition, so the gate must too. "
            "`if: false`/`always()` disables the gate outright, an inverted `if:` "
            "runs it only when the Go tests did not, and an extra `&& "
            "github.event_name == 'pull_request'` kills the branch ratchet -- "
            "all four were accepted by the suite before this check existed."
        )

    if producer_condition is not None and producer_condition != ALLOWED_GATE_IF:
        problems.append(
            f"the step producing coverage.out carries `if: {producer_condition!r}`, "
            f"but the gate runs under {ALLOWED_GATE_IF!r}. The producer must run "
            "at least as often as the gate, never less: a narrower `if:` on the "
            "producer means the gate scores a STALE profile from an earlier run "
            "while reporting success (round-6 measurement: narrowing the producer "
            "to `&& github.event_name == 'push'`, or to a `hashFiles(...)` "
            "pathspec, left the whole suite green)."
        )

    if not producer_fatal:
        problems.append(
            "the step producing coverage.out sets continue-on-error, so a failing "
            "`go test` still lets the gate score whatever partial profile was "
            "written (round-6 measurement: accepted with the whole suite green)."
        )

    env = step.get("env") or {}
    if not isinstance(env, dict):
        problems.append(f"the gate step's `env:` must be a mapping, got {env!r}")
        env = {}
    extra = sorted(set(env) - set(GATE_ENV_KEYS))
    missing = sorted(set(GATE_ENV_KEYS) - set(env))
    if extra:
        problems.append(
            f"the gate step's `env:` carries unexpected keys {extra!r}. Every extra "
            "key is an input the gate does not reason about but can still consume "
            "-- `COVERAGE_BASELINE: \"0\"` was measured to zero the enforced "
            "non-regression floor with the whole suite green. The gate reads "
            f"{sorted(GATE_ENV_KEYS)} and nothing else."
        )
    if missing:
        problems.append(
            f"the gate step's `env:` is missing {missing!r}; the gate's event "
            "routing is driven entirely by these variables"
        )
    for key, expected in GATE_ENV_VALUES.items():
        if key in env and env[key] != expected:
            problems.append(
                f"the gate step's `env: {key}` is {env[key]!r}, but it must be "
                f"{expected!r}. A hardcoded value silently re-routes the gate: "
                "`GITHUB_EVENT_NAME: push` sends every pull_request down the "
                "branch path, which is the exact defect this gate exists to "
                "prevent."
            )

    consumed = re.search(r"(\S*coverage\.out)\s*$", str(step.get("run", "")))
    if not consumed:
        problems.append(
            f"the gate step does not consume a coverage profile: {step.get('run')!r}"
        )
    elif f"-coverprofile={consumed.group(1)}" not in producer_run:
        problems.append(
            f"the gate scores {consumed.group(1)!r} but the producer step writes "
            f"{producer_run!r}: no `-coverprofile={consumed.group(1)}` was found, "
            "so the gate reads a file that step never wrote (round-5 measurement: "
            "renaming the producer's profile was accepted with 19 green tests)"
        )
    return problems


def job_problems(job: dict, expected_steps: tuple[str, ...]) -> list[str]:
    """Every reason `job` could report success without running its steps.

    A step-level guard cannot see its own job: `if: false` on the job skips all
    of it, and so does a `needs:` that never completes. Measured at cebdb3d33:
    `if: false` on `jobs.test-go` left the suite green, so "Go Tests" could be
    skipped wholesale with nothing in this file objecting.
    """
    problems = []
    condition = job.get("if")
    if condition is not None:
        problems.append(
            f"the test-go job carries `if: {condition!r}`, which can skip every "
            "step in it. The job that owns the INFRA-003 gate must be "
            "unconditional: a job-level condition is how the whole gate -- "
            "including the routing suite that guards it -- disappears (round-5 "
            "review)."
        )

    needs = job.get("needs")
    if needs:
        listed = [needs] if isinstance(needs, str) else list(needs)
        problems.append(
            f"the test-go job declares `needs: {listed!r}`, so its steps can be "
            "skipped wholesale whenever a listed job does not succeed. The "
            "INFRA-003 gate must not be behind another job's outcome."
        )
    names = [s.get("name") for s in job.get("steps", []) if isinstance(s, dict)]
    for step_name in expected_steps:
        if step_name not in names:
            problems.append(f"the test-go job has no step named {step_name!r}")
    return problems


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
            with no_gate_annotations():
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
        """The floor is not a waiver: dropping below the baseline fails.

        These are the calls that used to announce a fake regression in the CI
        log. The verdicts are still asserted -- that is the point of the test --
        but they are asserted against a captured buffer, so proving the gate
        bites no longer costs a reader of the log a phantom incident.
        """
        with no_gate_annotations():
            self.assertEqual(self.gate.evaluate_branch(45.0, 46.0), 1)
            self.assertEqual(self.gate.evaluate_branch(46.0, 46.0), 0)
            self.assertEqual(self.gate.evaluate_branch(60.0, 46.0), 0)

    def test_a_real_regression_is_ANNOUNCED_not_merely_failed(self):
        """Round-8: the return code is not the only thing a floor must do.

        Measured bypass on this tree, on BOTH `origin/develop` and the round-8
        branch, so it predates this change and was never reported:

          M5  replace `{ANNOTATION_PREFIX}Repository coverage regressed...`
              with the same sentence and no prefix
              -> 32 tests green. The gate still returns 1, so every existing
              assertion passed.

        That is a silent-failure guard, and this file exists because a silent
        signal was read as a real incident (round 7) and because a gate that
        quietly stops ratcheting is this card's actual subject. A floor that
        exits 1 while printing nothing produces a red job whose log does not say
        why -- and the next reader cannot tell a real regression from a broken
        scorer.

        So the breach is asserted as OUTPUT, driven through the gate exactly as
        `ci.yaml` drives it, and `leaked_annotations` is used deliberately here:
        this is the one place in the suite where a line-initial `::error::` is
        the required behaviour, so it is captured and asserted, not tolerated.
        """
        with tempfile.TemporaryDirectory() as tmp:
            profile = Path(tmp) / "coverage.out"
            # 45.0% covered: below the recorded floor, so this must be reported.
            profile.write_text(
                "mode: atomic\n"
                "example.com/m/covered.go:1.1,1.10 450 1\n"
                "example.com/m/uncovered.go:1.1,1.10 550 0\n",
                encoding="utf-8",
            )
            proc = subprocess.run(
                [sys.executable, str(SCRIPT_PATH), str(profile)],
                cwd=REPO_ROOT,
                capture_output=True,
                text=True,
                env={**os.environ, "GITHUB_EVENT_NAME": "push"},
            )
            combined = proc.stdout + proc.stderr
            self.assertEqual(
                proc.returncode, 1,
                f"a profile below the floor must fail, not pass:\n{combined}",
            )
            announced = leaked_annotations(combined)
            self.assertTrue(
                announced,
                "the gate failed but ANNOUNCED nothing. A red job that does "
                "not say why is indistinguishable from a broken scorer -- the "
                "same unreadable-log defect as round 7, and it is how a real "
                f"regression gets missed.\nGot:\n{combined}",
            )
            self.assertTrue(
                any("45.0%" in line for line in announced),
                f"the annotation must name the coverage that breached the "
                f"floor:\n{combined}",
            )

    def test_resolve_baseline_cannot_be_silenced_to_zero(self):
        """Round-8, M6: the ratchet itself must not be switchable off.

        Measured bypass: replacing the `return DEFAULT_BASELINE_COVERAGE`
        fallback in `resolve_baseline()` with `return 0.0` left every test
        green, on this tree and on `origin/develop`. `evaluate_branch` takes
        `baseline` as a parameter, so the suite's negative controls
        (`evaluate_branch(45.0, 46.0)`) never call `resolve_baseline()` at all
        -- the ratchet's real entry point was untested.

        Consequence: deleting the floor entirely would not turn any test red.
        This card's DONE WHEN says the floor must not be lowered to make CI
        pass; this is the assertion that makes that claim enforceable rather
        than a promise.

        Note the env var is only an *override for a ratchet bump* -- it may
        raise the floor, never remove it, which is why both directions are
        pinned: an unset, empty, malformed, negative or zero value all resolve
        to the recorded default rather than to "no floor".
        """
        self.assertEqual(self.gate.resolve_baseline(), 46.0)
        for hostile in ("", "   ", "not-a-number", "0", "0.0", "-1", "-99.5"):
            with self.subTest(baseline=hostile):
                with mock.patch.dict(
                    os.environ, {"COVERAGE_BASELINE": hostile}, clear=False
                ):
                    resolved = self.gate.resolve_baseline()
                self.assertGreater(
                    resolved, 0.0,
                    f"COVERAGE_BASELINE={hostile!r} resolved to {resolved}: "
                    "a zero or negative floor is not a ratchet",
                )
                self.assertEqual(
                    resolved, self.gate.DEFAULT_BASELINE_COVERAGE,
                    f"COVERAGE_BASELINE={hostile!r} must fall back to the "
                    f"recorded default {self.gate.DEFAULT_BASELINE_COVERAGE}, "
                    f"not to {resolved}",
                )
        # And the override may still RAISE the floor, which is its purpose.
        with mock.patch.dict(os.environ, {"COVERAGE_BASELINE": "52.5"}, clear=False):
            self.assertEqual(self.gate.resolve_baseline(), 52.5)

    def test_baseline_is_overridable_and_rejects_junk(self):
        with mock.patch.dict(os.environ, {"COVERAGE_BASELINE": "52.5"}, clear=False):
            self.assertEqual(self.gate.resolve_baseline(), 52.5)
        with mock.patch.dict(os.environ, {"COVERAGE_BASELINE": "not-a-number"}, clear=False):
            self.assertEqual(self.gate.resolve_baseline(), self.gate.DEFAULT_BASELINE_COVERAGE)

    def test_missing_coverage_file_fails_closed(self):
        """An absent gate input must fail, not pass by default."""
        self.assertIsNone(self.gate.resolve_coverage(self.tmp / "nope.out"))
        with no_gate_annotations():
            self.assertEqual(self.gate.main(["coverage_gate.py", "nope.out"]), 1)

    def test_branch_path_fails_closed_on_an_unmeasurable_profile(self):
        """A profile with no statements cannot be scored -- fail, never pass."""
        empty = self.tmp / "empty.out"
        empty.write_text("mode: atomic\n", encoding="utf-8")
        with mock.patch.dict(os.environ, {"GITHUB_EVENT_NAME": "push"}, clear=False):
            with no_gate_annotations():
                self.assertEqual(self.gate.main(["coverage_gate.py", str(empty)]), 1)

    def test_usage_error(self):
        with no_gate_annotations():
            self.assertEqual(self.gate.main(["coverage_gate.py"]), 2)


class WhitespaceOnlyChangeTest(unittest.TestCase):
    """A `gofmt -w` fix must not score as newly uncovered code (and vice versa).

    The defect this pins: `load_changed_lines` ran `git diff --unified=0`
    WITHOUT `--ignore-all-space`, so reformatting counted as added lines. A
    gofmt run on a struct literal realigns the field columns, which touches many
    pre-existing lines and adds no statement. Those lines match coverprofile
    blocks recorded against the previous layout, whose counts are 0, so the
    branch scored 0/34 = 0.0% against an 80% floor.

    Real instance: virtengine#1222, run 37177320130 job 111362592944. The
    branch's entire real payload was scripts/ci/lint_budget_gate.py -- the three
    Go files it also touched are invisible to `git diff -w`.

    The load-bearing half is the NEGATIVE CONTROL below. `test_real_added_line
    without_a_matching_covered_block_still_fails` is what stops this class from
    being satisfied by a gate that simply stopped counting things: if the
    `--ignore-all-space` flag were ever widened into something that also drops
    real content, or if the exclusion were moved into `main()` and applied to
    every line, the control goes red. A green pair is the only evidence that the
    exclusion is whitespace-shaped rather than coverage-shaped.
    """

    @classmethod
    def setUpClass(cls):
        cls._tmp = tempfile.TemporaryDirectory()
        root = Path(cls._tmp.name)
        # TWO separate repos. One repo cannot express both cases: the whitespace
        # case must have a diff that is EMPTY under --ignore-all-space, and the
        # control must have a diff that is not. Sharing one fixture would let the
        # fixture's own content decide which case runs, and the control would
        # pass for the wrong reason.
        cls.ws_repo = _git_repo_with_change(
            root / "ws",
            before="package keeper\n\ntype s struct {\n\ta int\n\tbcd int\n}\n",
            after="package keeper\n\ntype s struct {\n\ta   int\n\tbcd int\n}\n",
        )
        cls.real_repo = _git_repo_with_change(
            root / "real",
            before="package keeper\n\nfunc Changed() {}\n",
            after="package keeper\n\nfunc Changed() {}\n\nfunc Untested() int {\n\treturn 1\n}\n",
        )

    @classmethod
    def tearDownClass(cls):
        cls._tmp.cleanup()

    def _gate(self, repo: Path, profile: Path) -> subprocess.CompletedProcess:
        return subprocess.run(
            [sys.executable, str(repo / "scripts" / "ci" / "check_pr_diff_coverage.py"),
             str(profile)],
            cwd=repo,
            env={**os.environ, "BASE_REF": "develop"},
            capture_output=True,
            text=True,
        )

    def test_alignment_realignment_alone_does_not_fail_the_floor(self):
        """The #1222 shape: realignment only -> no changed lines -> green."""
        profile = self.ws_repo / "coverage.out"
        # Count 0 on every block, so the gate would score 0% if it counted them.
        profile.write_text(
            "mode: atomic\n"
            "x/keeper/query.go:4.1,4.10 1 0\n"
            "x/keeper/query.go:5.1,5.10 1 0\n",
            encoding="utf-8",
        )
        result = self._gate(self.ws_repo, profile)
        self.assertEqual(
            result.returncode, 0,
            "whitespace-only change must not fail the coverage floor:\n"
            f"{result.stdout}\n{result.stderr}",
        )
        self.assertIn("No changed Go lines detected", result.stdout)

    def test_real_added_line_without_a_matching_covered_block_still_fails(self):
        """NEGATIVE CONTROL: real content is still scored, and still fails."""
        profile = self.real_repo / "coverage.out"
        # A block covering the added `func Untested() int {` line, count 0.
        profile.write_text(
            "mode: atomic\n"
            "x/keeper/query.go:5.1,5.22 1 0\n",
            encoding="utf-8",
        )
        result = self._gate(self.real_repo, profile)
        self.assertEqual(
            result.returncode, 1,
            "a genuinely uncovered added line MUST still fail the 80% floor -- "
            "if this is green the whitespace exclusion has become a blanket "
            "exclusion and the gate is decorative:\n"
            f"{result.stdout}\n{result.stderr}",
        )
        self.assertIn("below minimum", result.stdout)

    def test_a_covered_real_added_line_still_passes(self):
        """The other half of the control: covered real code is not failed."""
        profile = self.real_repo / "coverage_covered.out"
        profile.write_text(
            "mode: atomic\n"
            "x/keeper/query.go:5.1,5.22 1 1\n",
            encoding="utf-8",
        )
        result = self._gate(self.real_repo, profile)
        self.assertEqual(result.returncode, 0, result.stdout)

    def test_the_flag_is_present_in_the_invocation(self):
        """Pin the flag itself, so it cannot be dropped without a test going red."""
        source = DIFF_GATE_PATH.read_text(encoding="utf-8")
        self.assertIn("--ignore-all-space", source)


def _git_repo_with_change(root: Path, before: str, after: str) -> Path:
    """A git repo whose only branch commit rewrites query.go from before to after."""
    repo = root
    (repo / "x" / "keeper").mkdir(parents=True)
    (repo / "scripts" / "ci").mkdir(parents=True)

    env = {**os.environ,
           "GIT_CONFIG_GLOBAL": os.devnull,
           "GIT_CONFIG_SYSTEM": os.devnull}

    def g(*args: str) -> str:
        return subprocess.run(["git", *args], cwd=repo, env=env, check=True,
                              capture_output=True, text=True).stdout.strip()

    g("init", "-q", "-b", "develop", ".")
    g("config", "user.email", "test@example.com")
    g("config", "user.name", "test")

    target = repo / "x" / "keeper" / "query.go"
    target.write_text(before, encoding="utf-8")
    g("add", "-A")
    g("commit", "-q", "-m", "base")
    g("update-ref", "refs/remotes/origin/develop", g("rev-parse", "HEAD"))

    target.write_text(after, encoding="utf-8")
    g("add", "-A")
    g("commit", "-q", "-m", "change")

    shutil.copy2(DIFF_GATE_PATH, repo / "scripts" / "ci" / DIFF_GATE_PATH.name)
    return repo


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

    # Both steps located by name. The gate step was always located this way; the
    # routing step was located by a content match on its `run` string
    # (`'test_coverage_gate.py' in step['run']`), which is why renaming it was
    # invisible to three separate assertions -- they followed the rename and
    # passed (round-5 review, mutation R4). A guard that follows its own target
    # is not a guard.
    COVERAGE_STEP_NAME = "Check coverage threshold"
    # `ROUTING_STEP_NAME` is intentionally NOT redefined here: it is a module
    # global shared with `routing_step_run()`, and two copies of a step name is
    # one more place for the two guards to drift apart.

    @classmethod
    def setUpClass(cls):
        cls.text = cls.CI_YAML.read_text(encoding="utf-8")
        cls.doc = yaml.safe_load(cls.text)
        cls.steps = cls.doc["jobs"]["test-go"]["steps"]
        cls.job = cls.doc["jobs"]["test-go"]

    def find_step(self, name, what):
        """The single parsed step called `name`. Fails if it is gone or renamed.

        Required to be exactly one so the gate cannot be duplicated into a
        second, unenforced copy.
        """
        matches = [s for s in self.steps
                   if isinstance(s, dict) and s.get("name") == name]
        self.assertEqual(
            len(matches), 1,
            f"expected exactly one step named {name!r} in jobs.test-go, found "
            f"{len(matches)}: {what} must exist under that exact name and must "
            f"not be duplicated. Found: "
            f"{[s.get('name') for s in self.steps if isinstance(s, dict)]}",
        )
        return matches[0]

    def find_coverage_step(self):
        """The single parsed step that runs the gate. Fails if it is gone."""
        return self.find_step(self.COVERAGE_STEP_NAME, "the INFRA-003 gate step")

    def find_routing_step(self):
        """The single parsed step that runs this suite. Fails if it is gone."""
        return self.find_step(ROUTING_STEP_NAME, "the routing-suite step")

    def find_profile_producer(self):
        """The step whose `run` writes the coverprofile the gate consumes."""
        matches = [s for s in self.steps
                   if isinstance(s, dict) and "-coverprofile=" in str(s.get("run", ""))]
        self.assertEqual(
            len(matches), 1,
            f"expected exactly one step writing a coverprofile, found "
            f"{len(matches)}: the gate's input must have exactly one producer",
        )
        return matches[0]

    def find_detect_step(self):
        matches = [s for s in self.steps
                   if isinstance(s, dict) and s.get("id") == "detect"]
        self.assertEqual(
            len(matches), 1, f"expected one step with id 'detect', found {len(matches)}")
        return matches[0]

    def assert_step_runs_its_command(self, step, command_re, what):
        problems = step_problems(step, command_re)
        self.assertEqual(
            problems, [],
            f"{what} must be a single command that executes it:\n  - "
            + "\n  - ".join(problems),
        )

    def assert_gate_inputs_are_wired(self):
        """`gate_input_problems` over the real gate step and its producer.

        The producer's `if:` and `continue-on-error` are part of the input, not
        decoration: the gate's only input is whatever that step wrote.
        """
        producer = self.find_profile_producer()
        problems = gate_input_problems(
            self.find_coverage_step(),
            producer["run"],
            producer.get("if"),
            not producer.get("continue-on-error", False),
        )
        self.assertEqual(
            problems, [],
            "the coverage gate step must be reached under exactly the condition "
            "that produced its input, and must carry no unhandled inputs:\n  - "
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
        self.assert_step_runs_its_command(
            self.find_routing_step(), ROUTING_COMMAND_RE,
            "the coverage gate routing step (it runs this suite)",
        )

    def test_gate_step_runs_under_exactly_the_condition_that_produced_its_input(self):
        """DEFECT 6, the load-bearing assertion of this round.

        Reproduced on this tree at cebdb3d33 -- each of these disabled the
        INFRA-003 gate with all 19 tests green, because `step_problems` read
        `run` and `continue-on-error` but never `step["if"]`:

            if: false                        the gate never runs
            if: <detect> && event == 'pull_request'   the branch ratchet dies
            if: steps.detect.outputs.run != 'true'   inverted: runs only when
                                                      the Go tests did NOT
            if: always()                     runs with no profile present

        The assertion is a positive pin on the exact condition rather than
        "has no `if:` key", because `always()` satisfies the latter while being
        weaker than the step that writes `coverage.out`.
        """
        self.assert_gate_inputs_are_wired()

    def test_gate_step_carries_no_unhandled_env(self):
        """DEFECT 6's quieter half: the same field, one line over.

        `env: COVERAGE_BASELINE: "0"` on the gate step was measured green at
        cebdb3d33. It does not disable the gate -- it disables the floor the
        gate exists to enforce, since `resolve_baseline` prefers the variable
        over the recorded baseline. The suite's event-routing assertions kept
        passing because they only checked that the two expected keys were
        *present*, never that they were the only ones.
        """
        step = self.find_coverage_step()
        env = step.get("env") or {}
        self.assertEqual(
            sorted(env), sorted(GATE_ENV_KEYS),
            f"the gate step's env must be exactly {sorted(GATE_ENV_KEYS)!r}, got "
            f"{sorted(env)!r}. An extra key is an input the gate can consume "
            "without the suite reasoning about it -- COVERAGE_BASELINE=0 zeroes "
            "the enforced non-regression floor.",
        )

    def test_the_gate_reads_the_profile_its_producer_writes(self):
        """The gate must score the file the test run actually produced.

        Renaming the producer's `-coverprofile` while leaving the gate's
        argument alone was green at cebdb3d33: the gate would then score a
        stale (or absent) artifact while every test in this file passed.
        """
        consumed = self.find_coverage_step()["run"].strip().split()[-1]
        produced = self.find_profile_producer()["run"]
        self.assertIn(
            f"-coverprofile={consumed}", produced,
            f"the gate scores {consumed!r} but the producing step does not write "
            f"it; it writes {produced!r}",
        )

    def test_detect_step_is_unconditional_so_the_gate_pin_is_not_vacuous(self):
        """`ALLOWED_GATE_IF` is evaluated from `detect`; `detect` must run.

        Otherwise the pin can be made vacuous: gating the producer of the flag
        on a second condition the gate does not see means the gate's `if` is
        true or false for reasons the gate cannot observe.
        """
        detect = self.find_detect_step()
        self.assertNotIn(
            "if", detect,
            "the `detect` step must be unconditional: the gate's pinned `if:` is "
            "evaluated from its output, so gating `detect` makes that pin "
            "vacuous",
        )

    def test_test_go_job_cannot_skip_its_own_steps(self):
        """DEFECT 6's outer ring: the job, not just the step.

        `if: false` on `jobs.test-go` was green at cebdb3d33. Every step-level
        guard in this file is satisfied by a job that never runs any of them --
        including the gate and the suite guarding it.
        """
        problems = job_problems(
            self.job, (self.COVERAGE_STEP_NAME, ROUTING_STEP_NAME),
        )
        self.assertEqual(
            problems, [],
            "the job owning the INFRA-003 gate must be unconditional and must "
            "not wait on another job:\n  - " + "\n  - ".join(problems),
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
        real_routing = self.find_routing_step()
        producer = self.find_profile_producer()
        producer_run = producer["run"]
        # The control: the real steps are clean, so the checks below cannot pass
        # vacuously by flagging everything.
        self.assertEqual(step_problems(real_gate, GATE_COMMAND_RE), [])
        self.assertEqual(step_problems(real_routing, ROUTING_COMMAND_RE), [])
        self.assertEqual(
            gate_input_problems(real_gate, producer_run, producer.get("if"),
                                not producer.get("continue-on-error", False)),
            [],
        )
        self.assertEqual(
            job_problems(self.job, (self.COVERAGE_STEP_NAME, ROUTING_STEP_NAME)), [],
        )

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

    def test_the_input_guards_themselves_reject_the_round_5_mutations(self):
        """Pins defect 6's regression *in the suite*, not only in my sweep.

        Without this, `gate_input_problems` and `job_problems` could be relaxed
        the same way `step_problems` was over rounds 3-5 -- one assertion at a
        time, each still green on the real workflow -- and the next review round
        would be the first place it showed up. Feeding each bypass shape to the
        checkers directly puts the checkers' own contract under test.

        The control comes first and is load-bearing in its own right: it is what
        proved the first draft of `gate_input_problems` under-specified. It
        validated the env KEY set but not the values, so a hardcoded
        `GITHUB_EVENT_NAME: push` -- which sends every PR down the branch path
        -- satisfied it. The checker was fixed, not the test.
        """
        producer_run = self.find_profile_producer()["run"]
        producer_if = self.find_profile_producer().get("if")
        good = {
            "if": ALLOWED_GATE_IF,
            "env": dict(GATE_ENV_VALUES),
            "run": "python3 scripts/ci/coverage_gate.py coverage.out",
        }
        # The controls. Without these the subTests below could all pass for the
        # wrong reason -- a checker that flags everything, or one that cannot be
        # called without the input it needs.
        self.assertEqual(gate_input_problems(good, producer_run, producer_if, True), [])
        self.assertEqual(
            gate_input_problems(good, producer_run, None, True), [],
            "an UNCONDITIONAL producer must be accepted: it is strictly wider "
            "than the gate, so the profile is always fresh. The risk runs one "
            "way and this is the safe end of it.",
        )
        with self.assertRaises(TypeError):
            gate_input_problems(good, producer_run)

        bypasses = {
            "if false": {**good, "if": "false"},
            "if always": {**good, "if": "always()"},
            "if inverted": {**good, "if": "steps.detect.outputs.run != 'true'"},
            "if pull-request only": {
                **good,
                "if": "steps.detect.outputs.run == 'true' && github.event_name == 'pull_request'",
            },
            "if undefined matrix": {**good, "if": "matrix.event == 'pull_request'"},
            "if absent": {k: v for k, v in good.items() if k != "if"},
            "baseline override": {
                **good, "env": {**good["env"], "COVERAGE_BASELINE": "0"},
            },
            "pythonpath override": {
                **good, "env": {**good["env"], "PYTHONPATH": "/nonexistent"},
            },
            "event dropped": {**good, "env": {"BASE_REF": GATE_ENV_VALUES["BASE_REF"]}},
            "event hardcoded": {
                **good, "env": {**good["env"], "GITHUB_EVENT_NAME": "push"},
            },
            "base ref hardcoded": {
                **good, "env": {**good["env"], "BASE_REF": "develop"},
            },
        }
        for label, step in bypasses.items():
            with self.subTest(bypass=label):
                self.assertNotEqual(
                    gate_input_problems(step, producer_run, producer_if, True), [],
                    f"gate_input_problems must reject the {label} bypass",
                )

        # Round 6: the PRODUCER's reach. Pinning the gate's own `if:` says
        # nothing about whether the step that writes coverage.out ran, and the
        # three shapes below were all measured green on this tree.
        producer_bypasses = {
            "producer narrowed to push": (
                "steps.detect.outputs.run == 'true' && github.event_name == 'push'"),
            "producer narrowed to a pathspec": (
                "steps.detect.outputs.run == 'true' && hashFiles('x/y.go') != ''"),
            "producer if false": "false",
            "producer if inverted": "steps.detect.outputs.run != 'true'",
        }
        for label, condition in producer_bypasses.items():
            with self.subTest(bypass=label):
                self.assertNotEqual(
                    gate_input_problems(good, producer_run, condition, True), [],
                    f"gate_input_problems must reject a producer that runs less "
                    f"often than the gate: {label}",
                )
        with self.subTest(bypass="producer continue-on-error"):
            self.assertNotEqual(
                gate_input_problems(good, producer_run, producer_if, False), [],
                "gate_input_problems must reject a producer whose failure is not "
                "fatal: it can hand the gate a truncated profile",
            )
        self.assertEqual(
            gate_input_problems(good, producer_run, producer_if, True), [],
            "the control must accept a fatal producer carrying the allowed if:",
        )

        # The producer must agree with what the gate is handed.
        self.assertNotEqual(
            gate_input_problems(
                good,
                producer_run.replace("-coverprofile=coverage.out",
                                     "-coverprofile=other.out"),
                producer_if, True),
            [],
            "gate_input_problems must reject a gate scoring a file nobody writes",
        )

        # And the job-level guards.
        names = (self.COVERAGE_STEP_NAME, ROUTING_STEP_NAME)
        clean_job = {"steps": [{"name": n} for n in names]}
        self.assertEqual(job_problems(clean_job, names), [])
        job_bypasses = {
            "job if false": {**clean_job, "if": "false"},
            "job always": {**clean_job, "if": "always()"},
            "job needs list": {**clean_job, "needs": ["never-runs"]},
            "job needs scalar": {**clean_job, "needs": "never-runs"},
            "gate step gone": {"steps": [{"name": ROUTING_STEP_NAME}]},
            "routing step gone": {"steps": [{"name": self.COVERAGE_STEP_NAME}]},
        }
        for label, job in job_bypasses.items():
            with self.subTest(bypass=label):
                self.assertNotEqual(
                    job_problems(job, names), [],
                    f"job_problems must reject the {label} bypass",
                )

    def test_coverage_step_env_carries_the_real_event_and_base(self):
        """The routing inputs must come from the event, on the step itself.

        Parsed `env` rather than file text: an `env:` block on some other step,
        or one that stopped being wired, must not satisfy the gate.
        """
        step = self.find_coverage_step()
        env = step.get("env") or {}
        for key, expected in GATE_ENV_VALUES.items():
            self.assertEqual(
                env.get(key), expected,
                f"the coverage step must set {key} from the event context",
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

        Located by name, not by the content match that made a rename invisible
        (round-5 review, mutation R4).
        """
        step = self.find_routing_step()
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
        run = self.find_detect_step()["run"]
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


class GateAnnotationHygieneTest(unittest.TestCase):
    """This suite must not print a `::error::` into the log it runs inside.

    Round-7 defect, observed on run 37268464612: a GREEN `Go Tests` job carried
    `::error::Repository coverage regressed to 45.0%, below the 46.0% floor`
    because the suite's own negative control printed the gate's real failure
    annotation while proving the return code. GitHub renders workflow commands
    regardless of exit status, so the phantom was indistinguishable from a live
    incident in the job log -- and it was read as one (kanban t_93bf2133).

    Everything in this class is about the suite's OUTPUT rather than the gate's
    behaviour, which is why it is a separate class: the gate is allowed to
    annotate, the guard is not.
    """

    def test_this_file_declares_no_bare_error_annotation(self):
        """Pin the reason the suite can talk about `::error::` without printing it.

        Every fixture call site goes through `no_gate_annotations()`. A new test
        that calls the scorer directly without that wrapper reintroduces the
        leak, and this catches the most likely shape of it.
        """
        text = Path(__file__).read_text(encoding="utf-8")
        # The only permitted literal is the constant the wrappers compare
        # against; it must never be handed to `print`.
        self.assertIn(f'ANNOTATION_ERROR = "{ANNOTATION_ERROR}"', text)
        for line in text.splitlines():
            if f'print("{ANNOTATION_ERROR}' in line or f"print('{ANNOTATION_ERROR}'" in line:
                self.fail(
                    f"this suite must not print a workflow annotation: {line.strip()!r}. "
                    "Wrap the call in `no_gate_annotations()` and assert on the "
                    "return code instead -- see the round-7 note above."
                )

    def test_the_suite_itself_emits_no_error_annotations(self):
        """THE LOAD-BEARING ONE: run this suite and read its stdout.

        The wrapper above is discipline; this is the check. It executes this
        exact file as a subprocess -- by reading the argv out of the `ci.yaml`
        step that really runs it, `-v` and all -- and fails if a single
        line-initial `::error::` reaches that output. So the leak cannot return
        via a new call site, a renamed helper, or a future edit that makes the
        scorer annotate somewhere else.

        Two round-8 corrections, both from trying to break this:
          * the argv is PARSED FROM `ci.yaml`, not restated here. Restating it
            is how this guard came to run a non-verbose child while CI ran a
            verbose one, and a docstring leak sailed straight through.
          * the assertion is LINE-INITIAL, not substring. Under `-v` this file's
            own prose quotes `::error::` mid-sentence, which GitHub never
            renders; matching substrings turned CI red on a correct tree.

        Recursion, and how it is stopped: this test runs the suite, and the suite
        contains this test, so an unguarded version spawns itself until the
        machine drowns. The child is marked with `SUITE_CHILD_ENV` and skips
        the spawn -- it still runs every other test and still emits whatever it
        would emit, which is exactly the output the parent inspects. One level
        deep, never two.

        Costs one extra run of a sub-second suite, and runs with the gate
        fixture calls still in place, so it fails on the real regression rather
        than on a hypothetical one.
        """
        if os.environ.get(SUITE_CHILD_ENV) == "1":
            self.skipTest(
                f"{SUITE_CHILD_ENV}=1: this IS the subprocess run whose output "
                "the parent inspects; spawning another would recurse forever"
            )

        env = {**os.environ, SUITE_CHILD_ENV: "1"}
        proc = subprocess.run(
            workflow_suite_argv(routing_step_run()),
            cwd=REPO_ROOT,
            env=env,
            capture_output=True,
            text=True,
        )
        leaked = leaked_annotations(proc.stdout + proc.stderr)
        self.assertEqual(
            leaked, [],
            "this suite leaked a GitHub error annotation into the CI log:\n  "
            + "\n  ".join(leaked)
            + "\nEvery fixture verdict must be captured with "
            "`no_gate_annotations()`; only a real gate run may annotate. Note "
            "that `-v` echoes each test method's docstring, so a docstring "
            "beginning with a workflow command leaks exactly this way.",
        )
        # The control: a green suite. Without this the assertion above could
        # pass because the suite failed to run at all, which would make the
        # guard vacuous on the exact tree it exists to protect.
        self.assertEqual(
            proc.returncode, 0,
            f"the suite must still pass while emitting no annotations:\n"
            f"{proc.stdout[-4000:]}\n{proc.stderr[-4000:]}",
        )

    def test_a_passing_gate_run_annotates_nothing(self):
        """Round-7 gap 2, found by falsifying the guard above.

        The subprocess guard above only inspects the SUITE's output, so it says
        nothing about the gate's own. Measured while falsifying this file:
        mutating `evaluate_branch` to prefix its success line with
        `::error::` left that guard completely green -- 31 tests, no failure --
        because every direct call site here is wrapped in
        `no_gate_annotations()`. The wrapper hid it, which is exactly what the
        wrapper is for.

        But that mutation is still a live defect in CI: `ci.yaml` invokes the
        gate directly, unwrapped, so a scorer that annotates on success paints
        an error annotation on every green push. Nothing in the suite would
        notice.

        So this drives the REAL gate as a process, the way `ci.yaml` does, on a
        profile that PASSES, and requires a clean log. No recursion concern:
        this spawns the gate, not the suite.
        """
        with tempfile.TemporaryDirectory() as tmp:
            profile = Path(tmp) / "coverage.out"
            # 46.5% covered: at the recorded floor, so this is a passing run.
            profile.write_text(
                "mode: atomic\n"
                "example.com/m/covered.go:1.1,1.10 465 1\n"
                "example.com/m/uncovered.go:1.1,1.10 535 0\n",
                encoding="utf-8",
            )
            proc = subprocess.run(
                [sys.executable, str(SCRIPT_PATH), str(profile)],
                cwd=REPO_ROOT,
                capture_output=True,
                text=True,
                env={**os.environ, "GITHUB_EVENT_NAME": "push"},
            )
            combined = proc.stdout + proc.stderr
            self.assertEqual(
                proc.returncode, 0,
                f"this fixture is meant to PASS the floor; it did not:\n{combined}",
            )
            self.assertNotIn(
                ANNOTATION_ERROR, combined,
                "the gate annotated ::error:: on a PASSING run. A green push "
                "would carry a failure annotation, which is the same "
                "phantom-incident defect as round 7 -- indistinguishable from "
                "a live one in the log. Only a genuine regression may annotate "
                f"::error::.\nGot:\n{combined}",
            )
            # It must still be doing its job: saying what it measured.
            self.assertIn("Total coverage: 46.5%", combined)


if __name__ == "__main__":
    unittest.main()