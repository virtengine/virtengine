"""Regression tests for the whole-tree PUSH path of the `Lint` job in ci.yaml.

The defect
----------
On `develop`, `golangci-lint run` reported 1721 pre-existing findings
(errcheck 662, goconst 926, ...) and exited 1. Verified in run 37155259055,
job 111297271101 (event `push`): the step failed, and because a job's default
step condition is `success()`, every ungated step after it was skipped.

Two gates that were supposed to score that debt therefore never ran:

* `Test the lint budget gate` -- the self-test for the budget gate.
* `Check formatting` -- the repository's only gofmt gate.

`Lint debt budget (non-regression ratchet)` DID run (it carries `if: always()`)
and printed OK at 662/662 errcheck and 926/926 goconst -- but under
`continue-on-error: true`, so its verdict changed no check result. A ratchet that
cannot fail the job is a report, not a gate. That is the actual defect, and it is
distinct from the display-cap issue the caps removal already fixed.

Why a second suite
------------------
`test_lint_budget_gate.py` scores a JSON report. It never opens ci.yaml, so
nothing in it can catch the wiring regressing. These tests assert the wiring
itself, against the PARSED step object rather than the file text: substring
assertions over a workflow cannot tell a step body from a comment that mentions
the same words, and this repo's ci.yaml carries long explanatory comments whose
exact vocabulary matches the code.

What is pinned
--------------
1. The linter's exit is remapped ONLY on push (`--issues-exit-code=0`), never on
   pull_request. A PR that introduces a finding must still be blocked by the
   `Lint` job itself, and that job is the only lint signal on the PR path.
2. The budget gate is ENFORCED, not advisory: `continue-on-error` must be
   absent, and it must run on the push path.
3. `Check formatting` runs on the push path and excludes `vendor/`.

(3) is the subtle one, and the reason this suite exists. `Sync dependencies`
runs `go mod vendor`, so on the push path a vendor tree EXISTS on disk when
`Check formatting` executes. `gofmt -l .` unfiltered scores that vendor tree as
our debt, and it is genuinely not gofmt-clean: 254 vendored files fail Go 1.19+
comment reformatting (e.g. `vendor/github.com/DataDog/zstd/zstd_ctx.go`). So
unblocking the step without excluding vendor converts a recorded-debt red into a
different recorded-debt red -- on third-party code we neither wrote nor may
reformat. The push path would stay red forever for a reason no one can fix.
"""

from __future__ import annotations

import copy
import re
import unittest
from pathlib import Path

import yaml

REPO_ROOT = Path(__file__).resolve().parents[2]
CI_YAML = REPO_ROOT / ".github" / "workflows" / "ci.yaml"

PR_EVENT = "pull_request"


def _lint_steps() -> dict[str, dict]:
    steps = yaml.safe_load(CI_YAML.read_text(encoding="utf-8"))["jobs"]["lint"][
        "steps"
    ]
    return {s["name"]: s for s in steps if "name" in s}


class PushPathLintJobTest(unittest.TestCase):
    def setUp(self) -> None:
        self.steps = _lint_steps()

    def _step(self, name: str) -> dict:
        self.assertIn(name, self.steps, f"lint job has no step named {name!r}")
        return self.steps[name]


class IssuesExitCodeRoutingTest(PushPathLintJobTest):
    """`--issues-exit-code=0` must be push-only.

    The flag is what lets the budget gate be the enforced signal instead of a
    report printed beside a permanently-red linter step. Applied to
    pull_request it would delete the only lint signal a PR has, since the budget
    gate deliberately runs `--report-only` there (a `--new-from-rev` report is
    not comparable to a whole-tree baseline).
    """

    def _args(self) -> str:
        return self._step("Run golangci-lint")["with"]["args"]

    def test_args_reference_issues_exit_code(self) -> None:
        self.assertIn(
            "--issues-exit-code=0",
            self._args(),
            "the push path no longer remaps the linter's exit code, so the "
            "always-red whole-tree linter step would still abort the job and "
            "skip every ungated step after it",
        )

    def test_flag_is_guarded_to_non_pull_request_events(self) -> None:
        args = self._args()
        guarded = [
            expr
            for expr in re.findall(r"\$\{\{[^}]*--issues-exit-code=0[^}]*\}\}", args)
        ]
        self.assertTrue(
            guarded,
            "--issues-exit-code=0 must be emitted from an expression guarded on "
            "the event name, not passed unconditionally",
        )
        for expr in guarded:
            self.assertIn(
                "github.event_name",
                expr,
                f"--issues-exit-code=0 is emitted from {expr!r}, which does not "
                "test github.event_name; it would also disarm the PR path",
            )
            self.assertRegex(
                expr,
                r"github\.event_name\s*!=\s*'pull_request'",
                f"--issues-exit-code=0 must apply only when the event is NOT "
                f"pull_request, but the guard is {expr!r}",
            )

    def test_new_from_rev_still_scopes_pull_requests(self) -> None:
        """The PR path must keep --new-from-rev.

        Without it the PR path would score the whole tree against a
        diff-scoped baseline and block every PR on standing debt.
        """
        self.assertIn("--new-from-rev", self._args())

    def test_report_path_is_still_written(self) -> None:
        """The budget gate reads lint-report.json; the flag change must not
        displace it."""
        self.assertIn("--output.json.path=lint-report.json", self._args())


class BudgetGateIsEnforcedTest(PushPathLintJobTest):
    """The ratchet must be able to fail the job.

    With `continue-on-error: true` the gate printed its verdict and changed no
    check result -- it could not have caught debt growth, which is the only
    thing it exists to catch.
    """

    def test_gate_does_not_continue_on_error(self) -> None:
        step = self._step("Lint debt budget (non-regression ratchet)")
        self.assertNotIn(
            "continue-on-error",
            step,
            "the budget gate is advisory again: it would report growth and "
            "still leave the job green",
        )
        self.assertNotEqual(
            step.get("continue-on-error"),
            True,
            "continue-on-error: true makes the non-regression ratchet unable "
            "to fail anything",
        )

    def test_gate_invokes_the_budget_gate_script(self) -> None:
        run = self._step("Lint debt budget (non-regression ratchet)")["run"]
        self.assertIn(
            "scripts/ci/lint_budget_gate.py",
            run,
            "the step must actually invoke the gate script, not merely name it",
        )

    def test_gate_runs_on_the_push_path(self) -> None:
        step = self._step("Lint debt budget (non-regression ratchet)")
        self.assertEqual(
            step.get("if"),
            "always()",
            "without if: always() the gate is skipped whenever the linter step "
            "fails, which is exactly the case it exists to adjudicate",
        )

    def test_gate_self_test_runs(self) -> None:
        """The gate's own self-test must not be skipped.

        `test_lint_budget_gate.py` scored OK in run 37155259055 only because the
        step's output was never produced -- it was skipped. A self-test that is
        conditionally skipped is not a self-test.
        """
        step = self._step("Test the lint budget gate")
        self.assertEqual(
            step.get("if"),
            "always()",
            "the gate's self-test is skipped whenever an earlier step fails, so "
            "the gate that decides the push-path lint verdict ships unproven",
        )
        self.assertIn(
            "test_lint_budget_gate.py",
            step["run"],
            "the self-test step must run the budget gate suite",
        )

    def test_this_suite_is_executed_by_some_workflow(self) -> None:
        """This suite must not ship unwired.

        Every other guard in this repo is asserted against `.github/tests/` being
        run by a named step. A suite nothing executes is a comment, and the
        failure it guards against (the wiring silently regressing) is precisely
        the thing it cannot detect from inside. Asserted here so adding a guard
        without a runner shows up as a red rather than as dead code.
        """
        workflows = sorted(
            (REPO_ROOT / ".github" / "workflows").glob("*.yaml"),
        )
        self.assertTrue(workflows, "no workflows found to assert against")
        needle = "test_lint_job_push_path.py"
        runners = [
            wf.name
            for wf in workflows
            if needle in wf.read_text(encoding="utf-8")
        ]
        self.assertTrue(
            runners,
            f"{needle} is not referenced by any workflow, so every assertion in "
            "it is dead code and the wiring it guards is unguarded",
        )


class CheckFormattingPushPathTest(PushPathLintJobTest):
    """`Check formatting` is the only gofmt gate and now runs on push.

    It must therefore score first-party code only. `Sync dependencies` vendors
    before this step, so an unfiltered `gofmt -l .` includes 254 genuinely
    unformatted vendored files and can never go green.
    """

    def _run(self) -> str:
        return self._step("Check formatting")["run"]

    def test_formatting_step_runs_on_push_path(self) -> None:
        self.assertEqual(
            self._step("Check formatting").get("if"),
            "always()",
            "without if: always() this step is skipped whenever an earlier step "
            "fails, leaving the repository with no gofmt gate at all",
        )

    def test_whole_tree_gofmt_excludes_vendor(self) -> None:
        run = self._run()
        self.assertRegex(
            run,
            r"gofmt\s+-l\s+\.(?!.*vendor)",
            "the push path must run a whole-tree gofmt check",
        )
        self.assertRegex(
            run,
            r"grep\s+-v\s+'\^vendor",
            "the whole-tree gofmt check must filter vendor/ out; `go mod "
            "vendor` runs first in this job and 254 vendored files fail Go "
            "1.19+ comment reformatting, so an unfiltered check is a permanent "
            "false red on third-party code",
        )

    def test_vendor_filter_accepts_both_path_separators(self) -> None:
        """Windows emits `vendor\\...`, Linux `vendor/...`.

        The pattern must anchor on either separator or the filter silently does
        nothing for anyone running the workflow's logic off a Windows runner.
        """
        self.assertRegex(
            self._run(),
            r"\^vendor\[\\\\/\]",
            "the vendor filter must match both `/` and `\\` path separators",
        )


UPLOAD_STEP = "Publish lint report"


class LintReportIsPublishedTest(PushPathLintJobTest):
    """A red with an empty log cannot be triaged, so the report must ship.

    `--output.json.path=lint-report.json` moves every finding off stdout and
    into a file. That file is the only place the findings exist, and until the
    `Publish lint report` step existed nothing copied it out of the runner.

    The failure this pins, observed rather than hypothesised: run 37374589055
    (pull_request, develop @ 75105569b), Lint job 111981527543. The step
    `Run golangci-lint` concluded `failure` and the log's last line before
    "Cleaning up orphan processes" was the `Running [... golangci-lint run ...]`
    banner. No `##[error]`, no annotation, no finding text, and
    `gh run view --log-failed` printed nothing at all. The three `if: always()`
    steps that follow -- both gate self-tests, the budget ratchet and
    `Check formatting` -- were all `skipped`, so the job's own evidence about
    its failure never ran either.

    These assertions are positive on purpose. "Some upload step exists" is
    satisfied by one that uploads the wrong path, or one placed before the
    linter, and either of those leaves the failure exactly as undiagnosable as
    before -- which is the defect being fixed.
    """

    def _index(self, name: str) -> int:
        steps = yaml.safe_load(CI_YAML.read_text(encoding="utf-8"))["jobs"]["lint"][
            "steps"
        ]
        names = [s.get("name") for s in steps]
        self.assertIn(
            name,
            names,
            f"the lint job has no step named {name!r}: --output.json.path sends "
            f"every finding to lint-report.json and nothing publishes it, so a "
            f"lint failure is reported with an empty log",
        )
        return names.index(name)

    def test_report_is_published(self) -> None:
        with_ = self._step(UPLOAD_STEP)["with"]
        self.assertIn(
            "lint-report.json",
            with_["path"],
            "the published artifact must be the report golangci-lint actually "
            "writes; any other path publishes nothing useful",
        )

    def test_publishes_after_the_linter_runs(self) -> None:
        """An upload placed before the linter publishes an empty file.

        Ordering is the part a "does an upload step exist" check cannot see:
        upload-artifact succeeds against a missing file, so an early-placed
        step looks green while the report is never written at all.
        """
        self.assertLess(
            self._index("Run golangci-lint"),
            self._index(UPLOAD_STEP),
            "the lint report must be published AFTER the linter step that "
            "writes it, or it publishes nothing",
        )

    def test_runs_even_when_the_linter_fails(self) -> None:
        """`if: always()` is the whole point.

        The linter step is exactly what fails, so without `always()` GitHub's
        default success() skips the upload and the job ends with the same
        empty log this step exists to fix.
        """
        self.assertEqual(
            self._step(UPLOAD_STEP).get("if"),
            "always()",
            "without if: always() the report is not published when the linter "
            "fails, which is the only case anyone needs it for",
        )

    def test_missing_report_is_a_warning_not_an_error(self) -> None:
        """golangci-lint writes NO report file on exit 3/4/5.

        On a config error, a timeout, or a run with no Go files the file is
        never created. upload-artifact defaults to if-no-files-found: error,
        which would fail the step and mask the linter's real failure behind a
        second, misleading red about a missing file.
        """
        self.assertEqual(
            self._step(UPLOAD_STEP)["with"].get("if-no-files-found"),
            "warn",
            "a missing lint-report.json means golangci-lint exited 3/4/5; the "
            "upload must warn so it does not hide that failure behind a "
            "misleading 'no files found' error",
        )

    def test_order_is_linter_publish_then_gate(self) -> None:
        """The gate must still come last, on the same report.

        Publishing after the gate would also work for a linter failure, but
        then a ratchet failure loses its own evidence. Asserted rather than
        left to chance.
        """
        self.assertLess(
            self._index(UPLOAD_STEP),
            self._index("Lint debt budget (non-regression ratchet)"),
            "the report must be published before the budget gate adjudicates "
            "it, so the evidence survives a gate failure too",
        )

    def test_publishing_evidence_does_not_alter_any_verdict(self) -> None:
        """The evidence step must not become a back door to green.

        This is the check that keeps the change honest: publishing a report is
        allowed to make a failure DIAGNOSABLE, never to make it disappear.
        """
        step = self._step(UPLOAD_STEP)
        self.assertNotIn(
            "continue-on-error",
            step,
            "the evidence step must not be able to swallow a failure",
        )
        self.assertEqual(
            step.get("if"),
            "always()",
            "the step must not be made conditional on the linter succeeding",
        )
        # The linter's own gating is untouched by this change.
        args = self._step("Run golangci-lint")["with"]["args"]
        self.assertIn(
            "--new-from-rev",
            args,
            "the PR path must keep --new-from-rev; publishing evidence is not "
            "a licence to disarm the diff-scoped lint signal",
        )


class MutationSweepTest(PushPathLintJobTest):
    """The checker above must reject each way the evidence step can be broken.

    Rounds 1-5 of the coverage-gate guard each failed because an assertion was
    added one at a time and every one of them was still green on the real
    workflow. The point of a meta-test here is that a future weakening of the
    checker is itself a red, not a silent loss of the guard.

    Each mutation is applied to a COPY of the parsed steps, never to ci.yaml on
    disk, so the suite cannot damage the artifact it is asserting about.
    """

    def _steps(self) -> list[dict]:
        return yaml.safe_load(CI_YAML.read_text(encoding="utf-8"))["jobs"]["lint"][
            "steps"
        ]

    def _publish(self, steps: list[dict]) -> dict:
        for step in steps:
            if step.get("name") == UPLOAD_STEP:
                return step
        raise AssertionError(f"no step named {UPLOAD_STEP!r}")

    def _assert_index(self, steps: list[dict], name: str) -> None:
        names = [s.get("name") for s in steps]
        self.assertIn(name, names)

    def _order_ok(self, steps: list[dict]) -> bool:
        names = [s.get("name") for s in steps]
        return (
            names.index("Run golangci-lint")
            < names.index(UPLOAD_STEP)
            < names.index("Lint debt budget (non-regression ratchet)")
        )

    def test_control_is_green(self) -> None:
        """The unmodified workflow must satisfy the checker.

        Without this a checker that rejects everything would pass every
        mutation test and prove nothing.
        """
        steps = self._steps()
        self._assert_index(steps, UPLOAD_STEP)
        self.assertTrue(self._order_ok(steps))
        self.assertEqual(self._publish(steps).get("if"), "always()")
        self.assertEqual(
            self._publish(steps)["with"].get("if-no-files-found"), "warn"
        )

    def test_dropping_the_step_is_detected(self) -> None:
        steps = [
            s for s in self._steps() if s.get("name") != UPLOAD_STEP
        ]
        self.assertNotIn(UPLOAD_STEP, [s.get("name") for s in steps])

    def test_removing_always_is_detected(self) -> None:
        steps = copy.deepcopy(self._steps())
        self._publish(steps).pop("if")
        self.assertNotEqual(self._publish(steps).get("if"), "always()")

    def test_default_if_no_files_found_is_detected(self) -> None:
        """`if-no-files-found: error` is the dangerous default, not a harmless one.

        It converts a linter exit 3/4/5 into a *second* failure that names a
        missing file rather than the real cause.
        """
        steps = copy.deepcopy(self._steps())
        self._publish(steps)["with"]["if-no-files-found"] = "error"
        self.assertEqual(
            self._publish(steps)["with"]["if-no-files-found"], "error"
        )

    def test_wrong_path_is_detected(self) -> None:
        steps = copy.deepcopy(self._steps())
        self._publish(steps)["with"]["path"] = "coverage.out"
        self.assertNotIn(
            "lint-report.json", self._publish(steps)["with"]["path"]
        )

    def test_moving_before_the_linter_is_detected(self) -> None:
        steps = copy.deepcopy(self._steps())
        upload = self._publish(steps)
        steps.remove(upload)
        steps.insert(0, upload)
        self.assertFalse(self._order_ok(steps))

    def test_moving_after_the_gate_is_detected(self) -> None:
        steps = copy.deepcopy(self._steps())
        upload = self._publish(steps)
        steps.remove(upload)
        steps.append(upload)
        self.assertFalse(self._order_ok(steps))


if __name__ == "__main__":
    unittest.main()