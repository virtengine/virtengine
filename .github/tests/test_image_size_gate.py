"""Assert the INFRA-003 container image size gate enforces a real ratchet.

DONE WHEN for kanban t_85756c54 requires the 100 MiB budget to be re-baselined
as a measured ratchet (or an accepted size above the real image) with the gate
comparing against that recorded number. These tests pin that, and -- more
importantly -- pin that the ratchet is NOT a disabled check.

The regression they guard: `ci.yaml` compared the image against a fixed
`MAX_BYTES=$((100 * 1024 * 1024))` that the shipped image has never met. The
measured image is 216,548,169 bytes and the step has zero successful evaluations
in the run history, so the gate could only ever say "fail" -- which hides real
size regressions behind a target nobody is working toward.

What makes this an honest test rather than a rubber stamp:

* `test_the_recorded_measurement_is_the_real_one` pins the constant to the byte
  value four real CI runs actually reported, so a future edit cannot quietly
  re-baseline the gate to something convenient.
* `test_ratchet_still_fails_a_real_regression` proves growth past the ceiling
  fails, at the boundary.
* `test_gate_fails_closed_when_it_cannot_measure` proves an absent or junk input
  fails instead of passing or falling back to the recorded number.
* `test_expired_review_window_fails_the_gate` proves the budget has an expiry,
  so a recorded number cannot silently become a permanent waiver.
* `test_workflow_no_longer_compares_against_the_unreachable_literal` proves the
  dead inline comparison cannot come back.

The suite is falsified against the pre-change fixed-100MiB behaviour; see
`test_prechange_fixed_cap_would_fail` for the comparison that proves the tests
are not vacuous.
"""

from __future__ import annotations

import importlib.util
import os
import subprocess
import sys
import unittest
from datetime import date, timedelta
from pathlib import Path
from unittest import mock

REPO_ROOT = Path(__file__).resolve().parents[2]
SCRIPT_PATH = REPO_ROOT / "scripts" / "ci" / "image_size_gate.py"
CI_YAML = REPO_ROOT / ".github" / "workflows" / "ci.yaml"

# The byte value the `Enforce virtengine image size` step reported in four real
# `Container Security` jobs across three different head SHAs. Deterministic, so
# it doubles as the gate's regression test fixture.
MEASURED_IN_CI = 216548169

# What the pre-change inline step compared against.
OLD_MAX_BYTES = 100 * 1024 * 1024


def load_module():
    spec = importlib.util.spec_from_file_location("image_size_gate", SCRIPT_PATH)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def run_gate(image_bytes: str | None, extra_env: dict | None = None):
    """Run the gate as a process with the real production call shape."""
    env = {**os.environ}
    if image_bytes is None:
        env.pop("IMAGE_SIZE_BYTES", None)
    else:
        env["IMAGE_SIZE_BYTES"] = image_bytes
    env.update(extra_env or {})
    return subprocess.run(
        [sys.executable, str(SCRIPT_PATH), "virtengine:ci"],
        cwd=REPO_ROOT,
        env=env,
        capture_output=True,
        text=True,
    )


class RecordedMeasurementTest(unittest.TestCase):
    """The recorded numbers must be the real measured ones, not convenient ones."""

    def setUp(self):
        self.gate = load_module()

    def test_the_recorded_measurement_is_the_real_one(self):
        """Pinned to what CI actually reported, byte for byte.

        Four `Container Security` jobs reported this exact value on four runs
        across three head SHAs (9804a5010, db0a6b20f, 9f1bcd6d0, 1cd9af8e3), so
        a real regression in the constant is visible here as a failing test.
        """
        self.assertEqual(self.gate.MEASURED_IMAGE_BYTES, MEASURED_IN_CI)

    def test_the_recorded_binary_is_the_measured_one(self):
        """193,294,498 B: linux/amd64, Dockerfile flags, develop @ 1cd9af8e3.

        HOW this was measured -- reproduce it this way, or you will record a
        binary that never ships:

            git worktree add --detach <path> <sha>
            cd <path>
            CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \\
              go build -trimpath -buildvcs=false -ldflags "-s -w" -o out ./cmd/virtengine
            wc -c < out

        Do NOT add `GOWORK=off`. `_build/Dockerfile.virtengine` sets only
        CGO_ENABLED/GOOS/GOARCH and never sets GOWORK, `go.work` is not excluded
        by `.dockerignore`, and `COPY . .` puts it in the build context -- so the
        container builds in workspace mode. At c2cd1baf the two modes differ:

            workspace mode (what the image ships)  193,294,498 B
            GOWORK=off      (a different binary)    193,314,978 B  (+20,480)

        Re-pinning this assertion to whatever a re-measure happens to print is
        exactly the mutation class the falsification harness exists to catch, so
        if this test fails, confirm the build MODE first and then update the
        constant, the docstring and the ci.yaml comment together.
        """
        self.assertEqual(self.gate.MEASURED_BINARY_BYTES, 193294498)

    def test_recorded_binary_is_the_workspace_mode_build(self):
        """The recorded binary must be the workspace-mode (shipped) build.

        The gate never fails on this constant -- `evaluate()` decides on
        `image_bytes` against `MAX_IMAGE_BYTES` -- so a regression here would be
        invisible to CI. Without this test the diagnostic print can attribute the
        image gap to a binary the container never copies, which is how the
        GOWORK=off figure got in originally.
        """
        # The size of the GOWORK=off build of the same tree, same toolchain and
        # same flags. Kept as a distinct literal rather than a recomputation so
        # that swapping the recorded constant back to it fails here.
        gowork_off_build = 193314978
        self.assertNotEqual(
            self.gate.MEASURED_BINARY_BYTES,
            gowork_off_build,
            "recorded binary is the GOWORK=off build, which the image does not ship",
        )
        # Workspace mode is the smaller build here; record that as the
        # direction, so a future toolchain change that flips it is visible.
        self.assertLess(
            self.gate.MEASURED_BINARY_BYTES,
            gowork_off_build,
            "workspace-mode build stopped being smaller than GOWORK=off",
        )
        # Unstripped must exceed stripped, or the stripping figure is nonsense.
        self.assertGreater(
            self.gate.MEASURED_UNSTRIPPED_BINARY_BYTES,
            self.gate.MEASURED_BINARY_BYTES,
        )
        # The binary must remain the dominant layer, and must still be over the
        # reported 100 MiB target on its own -- the premise of the whole design.
        self.assertLess(
            self.gate.MEASURED_BINARY_BYTES,
            self.gate.MEASURED_IMAGE_BYTES,
            "recorded binary exceeds the recorded image; attribution is wrong",
        )
        self.assertGreater(
            self.gate.MEASURED_BINARY_BYTES,
            self.gate.TARGET_IMAGE_BYTES,
            "recorded binary is now under the INFRA-003 target; re-measure the image",
        )

    def test_ratchet_is_tight_enough_to_notice_growth(self):
        """~2.2% of headroom. A ceiling with huge headroom is a disabled check."""
        headroom = self.gate.MAX_IMAGE_BYTES - self.gate.MEASURED_IMAGE_BYTES
        self.assertGreater(headroom, 0, "ceiling must admit the measured image")
        self.assertLess(
            headroom,
            self.gate.MEASURED_IMAGE_BYTES * 0.05,
            "ratchet headroom must stay under 5% of the measurement",
        )

    def test_target_is_reported_not_deleted(self):
        """The 100 MiB INFRA-003 target stays in the module as the aspiration."""
        self.assertEqual(self.gate.TARGET_IMAGE_BYTES, OLD_MAX_BYTES)

    def test_budget_names_a_tracking_issue(self):
        """An unowned number is exactly what this change exists to prevent."""
        self.assertRegex(self.gate.BUDGET_ISSUE, r"virtengine/virtengine#\d+")


class RatchetEvaluationTest(unittest.TestCase):
    """The enforced decision, asserted against the real measured input."""

    def setUp(self):
        self.gate = load_module()

    def test_the_measured_image_passes_the_ratchet(self):
        """THE REGRESSION.

        Before this change the measured image was compared against 100 MiB and
        failed on every run, forever, with no path to green.
        """
        self.assertEqual(self.gate.evaluate(MEASURED_IN_CI, self.gate.MAX_IMAGE_BYTES), 0)

    def test_prechange_fixed_cap_would_fail(self):
        """Falsification: the OLD comparison rejects the exact same input.

        This is what makes the suite non-vacuous. If the measured image ever fit
        under the old fixed cap, the tests above would prove nothing, because
        the gate would not have needed changing at all.
        """
        self.assertGreater(MEASURED_IN_CI, OLD_MAX_BYTES)
        self.assertEqual(self.gate.evaluate(MEASURED_IN_CI, OLD_MAX_BYTES), 1)

    def test_ratchet_still_fails_a_real_regression(self):
        """The floor is not a waiver: growth past the ceiling fails."""
        self.assertEqual(self.gate.evaluate(self.gate.MAX_IMAGE_BYTES + 1, self.gate.MAX_IMAGE_BYTES), 1)
        self.assertEqual(self.gate.evaluate(self.gate.MAX_IMAGE_BYTES, self.gate.MAX_IMAGE_BYTES), 0)
        # 250 MiB is above the 211 MiB ceiling. (200 MiB would NOT fail -- it is
        # under the ceiling, and that is the gate working, not a gap in it.)
        self.assertEqual(self.gate.evaluate(250 * 1024 * 1024, self.gate.MAX_IMAGE_BYTES), 1)
        # A large reduction still passes: the ratchet must not punish shrinkage.
        self.assertEqual(self.gate.evaluate(150 * 1024 * 1024, self.gate.MAX_IMAGE_BYTES), 0)

    def test_over_target_image_still_reports_the_gap(self):
        """While the gap persists, the aspiration must stay visible in the output."""
        import io
        import contextlib

        buffer = io.StringIO()
        with contextlib.redirect_stdout(buffer):
            self.gate.evaluate(MEASURED_IN_CI, self.gate.MAX_IMAGE_BYTES)
        output = buffer.getvalue()
        self.assertIn("::warning::", output)
        self.assertIn("100 MiB", output)
        self.assertIn(self.gate.BUDGET_ISSUE, output)

    def test_expired_review_window_fails_the_gate(self):
        """A recorded budget with no expiry is a waiver nobody has to renew."""
        day_after_expiry = self.gate.RATCHET_EXPIRES + timedelta(days=1)
        self.assertEqual(
            self.gate.evaluate(
                MEASURED_IN_CI, self.gate.MAX_IMAGE_BYTES, today=day_after_expiry
            ),
            1,
        )
        self.assertEqual(
            self.gate.evaluate(
                MEASURED_IN_CI, self.gate.MAX_IMAGE_BYTES, today=self.gate.RATCHET_EXPIRES
            ),
            0,
            "the expiry date itself must still be in window",
        )

    def test_review_window_is_a_real_window(self):
        """Guards the window against being widened into meaninglessness."""
        self.assertLessEqual(self.gate.RATCHET_EXPIRES, date(2027, 1, 1))
        self.assertLess(self.gate.RATCHET_REVIEWED, self.gate.RATCHET_EXPIRES)


class CeilingResolutionTest(unittest.TestCase):
    """The ceiling is tightenable, but the override can never RAISE it."""

    def setUp(self):
        self.gate = load_module()

    def test_override_can_tighten_the_ceiling(self):
        """Lowering the bar is always safe, so it is allowed."""
        tighter = self.gate.MEASURED_IMAGE_BYTES
        with mock.patch.dict(os.environ, {"IMAGE_SIZE_MAX": str(tighter)}, clear=False):
            self.assertEqual(self.gate.resolve_max_bytes(), tighter)

    def test_override_cannot_raise_the_ceiling(self):
        """GOVERNANCE: a raisable override launders growth as "expected weight".

        Anyone able to set a repo/org variable (the workflow wires
        `IMAGE_SIZE_MAX: ${{ vars.IMAGE_SIZE_MAX }}`) could otherwise raise the
        budget with no diff in the recorded gate, defeating the whole point of
        recording it. ESTATE.md rejected exactly this move on the pages perf
        gate. Raising must require editing MAX_IMAGE_BYTES in the script.
        """
        raised = self.gate.MAX_IMAGE_BYTES + 1
        with mock.patch.dict(os.environ, {"IMAGE_SIZE_MAX": str(raised)}, clear=False):
            self.assertEqual(self.gate.resolve_max_bytes(), self.gate.MAX_IMAGE_BYTES)

    def test_raising_override_does_not_actually_raise_the_verdict(self):
        """End to end: a 954 MiB budget must not wave a growing image through."""
        recorded = self.gate.MAX_IMAGE_BYTES
        with mock.patch.dict(os.environ, {"IMAGE_SIZE_MAX": "999999999"}, clear=False):
            self.assertEqual(
                self.gate.resolve_max_bytes(), recorded, "override must be refused"
            )
            # 250 MiB image would pass a 999999999-byte ceiling.
            self.assertEqual(self.gate.evaluate(250 * 1024 * 1024, self.gate.resolve_max_bytes()), 1)

    def test_junk_and_nonpositive_overrides_fall_back(self):
        for junk in ("not-a-number", "0", "-1", ""):
            with mock.patch.dict(os.environ, {"IMAGE_SIZE_MAX": junk}, clear=False):
                self.assertEqual(self.gate.resolve_max_bytes(), self.gate.MAX_IMAGE_BYTES)


class FailClosedTest(unittest.TestCase):
    """An unmeasurable input must fail, never pass and never fall back."""

    def setUp(self):
        self.gate = load_module()

    def test_missing_size_fails_closed(self):
        self.assertIsNone(self.gate.resolve_image_size())
        with mock.patch.dict(os.environ, {}, clear=False):
            os.environ.pop("IMAGE_SIZE_BYTES", None)
            self.assertEqual(self.gate.main(["image_size_gate.py", "virtengine:ci"]), 1)

    def test_unparseable_size_fails_closed(self):
        with mock.patch.dict(os.environ, {"IMAGE_SIZE_BYTES": "not-a-number"}, clear=False):
            self.assertIsNone(self.gate.resolve_image_size())

    def test_nonpositive_size_is_rejected(self):
        proc = run_gate("0")
        self.assertEqual(proc.returncode, 1, proc.stdout + proc.stderr)

    def test_usage_error(self):
        self.assertEqual(self.gate.main(["image_size_gate.py"]), 2)


class ProductionShapeTest(unittest.TestCase):
    """Drive the gate exactly as the workflow does."""

    def test_measured_image_passes_as_a_process(self):
        proc = run_gate(str(MEASURED_IN_CI))
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        self.assertIn(f"Image size: {MEASURED_IN_CI} bytes", proc.stdout)
        self.assertIn("OK: Image size", proc.stdout)

    def test_growth_fails_as_a_process(self):
        proc = run_gate(str(MEASURED_IN_CI + 50_000_000))
        self.assertEqual(proc.returncode, 1, proc.stdout + proc.stderr)
        self.assertIn("::error::", proc.stdout)

    def test_env_override_raises_the_ceiling(self):
        """A RAISING override must be refused end to end, as a process too."""
        proc = run_gate(str(MEASURED_IN_CI), extra_env={"IMAGE_SIZE_MAX": "300000000"})
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        self.assertIn("REFUSED", proc.stderr)
        # The verdict must be the recorded ceiling's, not the override's.
        self.assertIn(f"holds the {load_module().MAX_IMAGE_BYTES}-byte ratchet ceiling", proc.stdout)

    def test_env_override_may_tighten(self):
        """Tightening below the measured image makes the gate fail -- as intended."""
        proc = run_gate(str(MEASURED_IN_CI), extra_env={"IMAGE_SIZE_MAX": "200000000"})
        self.assertEqual(proc.returncode, 1, proc.stdout + proc.stderr)
        self.assertIn("::error::", proc.stdout)


class WorkflowWiringTest(unittest.TestCase):
    """The workflow must call the script, not re-inline the comparison."""

    def setUp(self):
        self.gate_module = load_module()

    def test_image_size_step_invokes_the_gate_script(self):
        text = CI_YAML.read_text(encoding="utf-8")
        self.assertIn("scripts/ci/image_size_gate.py", text)

    def test_workflow_no_longer_compares_against_the_unreachable_literal(self):
        """The comparison that made every run red must not come back."""
        text = CI_YAML.read_text(encoding="utf-8")
        self.assertNotIn("MAX_BYTES=$((100 * 1024 * 1024))", text)
        self.assertNotIn("Image size exceeds 100MB target", text)

    def test_workflow_measures_the_image_and_hands_it_to_the_gate(self):
        text = CI_YAML.read_text(encoding="utf-8")
        self.assertIn("docker image inspect -f '{{.Size}}' virtengine:ci", text)
        self.assertIn("IMAGE_SIZE_BYTES=", text)

    def test_gate_test_runs_in_the_container_security_job(self):
        """The test must run in the job that enforces the gate, not elsewhere."""
        text = CI_YAML.read_text(encoding="utf-8")
        self.assertIn('python -m unittest discover -s .github/tests -p "test_image_size_gate.py"', text)

    def test_env_does_not_self_reference_the_step_output(self):
        """A step's own output is empty in its own env -- it would blank the gate."""
        text = CI_YAML.read_text(encoding="utf-8")
        self.assertNotIn("IMAGE_SIZE_BYTES: ${{ steps.image_size.outputs", text)

    def test_workflow_comment_records_the_same_binary_size_as_the_gate(self):
        """ci.yaml must quote the binary size the gate actually records.

        The figure only appears in a comment here, so nothing fails if it drifts
        away from `MEASURED_BINARY_BYTES` -- and a stale comment is exactly how
        the next engineer ends up re-pinning the constant from the wrong build.
        This is the cross-file tie that makes the comment trustworthy.
        """
        text = CI_YAML.read_text(encoding="utf-8")
        mib = 1024 * 1024
        binary = self.gate_module.MEASURED_BINARY_BYTES
        self.assertIn(
            f"{binary:,} bytes ({binary / mib:.2f} MiB",
            text,
            "ci.yaml no longer states the recorded binary size; the comment has "
            "drifted from scripts/ci/image_size_gate.py",
        )
        # And it must not still be quoting the GOWORK=off build.
        self.assertNotIn("193,314,978", text)

    def test_workflow_comment_records_which_build_mode_was_measured(self):
        """ci.yaml must say the figure is the workspace-mode build.

        The bare number alone is not enough: `193,294,498` is only meaningful
        next to the reason it is the right number. Without the mode note a
        future engineer can read the comment, re-measure with GOWORK=off, see a
        slightly different figure, and "correct" the constant back. Removal of
        the note is treated as a regression, same as removal of the number.
        """
        text = CI_YAML.read_text(encoding="utf-8")
        self.assertIn("workspace-mode", text)
        self.assertIn(".dockerignore", text)

    def test_gate_script_documents_the_remeasure_method_and_the_gowork_trap(self):
        """The re-measure recipe and its GOWORK=off trap must survive.

        These comments are the only record of *how* the constant was measured.
        Without them the cheapest fix for a failing measurement test is to edit
        the number -- which is the exact mutation the harness exists to catch --
        so their removal is treated as a failure rather than a cleanup.
        """
        text = SCRIPT_PATH.read_text(encoding="utf-8")
        self.assertIn("HOW TO RE-MEASURE", text)
        self.assertIn("Do NOT set `GOWORK=off`", text)
        self.assertIn("workspace mode", text)
        self.assertIn("193,294,498", text)
        self.assertIn("MEASURED_UNSTRIPPED_BINARY_BYTES", text)
        # The runtime-layer figure must stay labelled as derived, since no docker
        # daemon was available to stat() the layers directly.
        self.assertIn("derived by subtraction", text)

    def test_reported_target_and_ceiling_are_untouched_by_the_measurement_fix(self):
        """Correcting a recorded measurement must not move a threshold.

        `MAX_IMAGE_BYTES`, `TARGET_IMAGE_BYTES` and the expiry are the enforced
        part of this gate. The workspace-mode correction is diagnostic-only, so
        if any of these move in the same commit the change is not a measurement
        fix and needs a fresh sign-off.
        """
        self.assertEqual(self.gate_module.TARGET_IMAGE_BYTES, 104857600)
        self.assertEqual(self.gate_module.MAX_IMAGE_BYTES, 211 * 1024 * 1024)
        self.assertEqual(self.gate_module.MEASURED_IMAGE_BYTES, MEASURED_IN_CI)


if __name__ == "__main__":
    unittest.main()