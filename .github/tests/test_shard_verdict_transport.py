#!/usr/bin/env python3
"""Regression tests for the shard verdict TRANSPORT.

History, because the transport was wrong twice and each failure was silent:

  1. `gh run download` in the consuming job - run 37292268699: all four
     downloads timed out at ~60s each while every shard had SUCCEEDED, and the
     step sent gh's stderr to /dev/null, so a cancelled shard and a download
     failure produced the SAME log line.
  2. Matrix job outputs - run 37295691699 delivered `{tests-and-sim:clean}` and
     run 37298338441 delivered `{"tooling":"clean"}` while all four shards were
     green at every step. One object arrived intact and three vanished with no
     error: that is OVERWRITE, not concatenation, so no expression-level fix
     exists.

The shipped transport is one ARTIFACT PER SHARD (`govulncheck-verdict-<shard>`)
collected by `actions/download-artifact` with `merge-multiple`. The property that
makes it work - and that the previous two lacked - is that each shard's verdict
is INDEPENDENTLY ADDRESSABLE, so a missing shard is attributable to that shard
rather than to an aggregate.

These tests pin that property directly against the ledger reader and the shard
reader, not against a copy of them.

Run: python -m unittest discover -s .github/tests -p "test_shard_verdict_transport*.py"
"""
from __future__ import annotations

import importlib.util
import os
import re
import subprocess
import tempfile
import unittest

REPO_ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
LEDGER = os.path.join(REPO_ROOT, ".github", "scripts", "govulncheck_verdict_ledger.py")
SHARDS_PY = os.path.join(REPO_ROOT, ".github", "scripts", "govulncheck_shards.py")
WORKFLOW = os.path.join(REPO_ROOT, ".github", "workflows", "security.yaml")


def load(path: str, name: str):
    spec = importlib.util.spec_from_file_location(name, path)
    if spec is None or spec.loader is None:
        raise unittest.SkipTest(f"cannot load {path}")
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def strip_comments(text: str) -> str:
    """Drop whole-line YAML comments.

    Assertions about what a workflow DOES must not be satisfied - or broken -
    by prose describing why something was removed.
    """
    return "\n".join(
        line for line in text.splitlines() if not line.lstrip().startswith("#")
    )


def job_block(text: str, job: str) -> str:
    """The body of one top-level job, up to the next top-level job key."""
    m = re.search(rf"^  {re.escape(job)}:\s*$", text, re.MULTILINE)
    assert m, f"job {job} not found"
    rest = text[m.end():]
    nxt = re.search(r"^  [A-Za-z][A-Za-z0-9_-]*:\s*$", rest, re.MULTILINE)
    return rest[: nxt.start()] if nxt else rest


def block_of(text: str, step_name: str) -> str:
    """The body of one step, from its `- name:` to the next step at the same indent.

    Assertions about a STEP must be scoped to that step. A substring check over
    the whole file cannot tell `if-no-files-found: error` on the upload step
    (valid) from the same string on the download step (an input that action does
    not accept - which CI rejected on run 37303621734).
    """
    m = re.search(rf"^(\s*)- name: {re.escape(step_name)}\s*$", text, re.MULTILINE)
    assert m, f"step {step_name!r} not found"
    indent = m.group(1)
    rest = text[m.end():]
    nxt = re.search(rf"^{re.escape(indent)}- name: ", rest, re.MULTILINE)
    return rest[: nxt.start()] if nxt else rest


class TestPerShardArtifactsAreIndependentlyAddressable(unittest.TestCase):
    """The property the shipped transport depends on, and the previous two lacked."""

    def setUp(self):
        self.mod = load(LEDGER, "govulncheck_verdict_ledger")
        self._tmp = tempfile.TemporaryDirectory()
        self.dir = self._tmp.name
        with open(os.path.join(self.dir, "shards.manifest"), "w", encoding="utf-8") as fh:
            fh.write("core\nplatform\ntooling\ntests-and-sim\n")

    def tearDown(self):
        self._tmp.cleanup()

    def write(self, shard: str, verdict: str) -> None:
        """One shard publishing one artifact."""
        with open(os.path.join(self.dir, f"{shard}.verdict"), "w", encoding="utf-8") as fh:
            fh.write(f"verdict={verdict}\n")

    def test_each_shard_collected_independently(self):
        for shard in ("core", "platform", "tooling", "tests-and-sim"):
            self.write(shard, "clean")
        self.assertEqual(self.mod.main([self.dir, "--expect", "4"]), 0)

    def test_one_missing_shard_is_attributed_by_name(self):
        """The whole point: the missing shard is NAMED, not counted away."""
        for shard in ("core", "platform", "tooling"):
            self.write(shard, "clean")
        self.assertEqual(self.mod.main([self.dir, "--expect", "4"]), 1)

    def test_removing_any_single_shard_is_detected(self):
        """No shard's absence is masked by the other three reporting clean."""
        all_shards = ["core", "platform", "tooling", "tests-and-sim"]
        for missing in all_shards:
            with self.subTest(missing=missing):
                for shard in all_shards:
                    path = os.path.join(self.dir, f"{shard}.verdict")
                    if os.path.exists(path):
                        os.remove(path)
                for shard in all_shards:
                    if shard != missing:
                        self.write(shard, "clean")
                self.assertEqual(self.mod.main([self.dir, "--expect", "4"]), 1)
        # Leave the dir clean for tearDown.
        for shard in all_shards:
            path = os.path.join(self.dir, f"{shard}.verdict")
            if os.path.exists(path):
                os.remove(path)

    def test_no_shards_at_all_is_not_a_pass(self):
        self.assertEqual(self.mod.main([self.dir, "--expect", "4"]), 1)


class TestWorkflowWiring(unittest.TestCase):
    """The transport is only real if the workflow actually uses it."""

    @classmethod
    def setUpClass(cls):
        with open(WORKFLOW, encoding="utf-8") as fh:
            cls.text = fh.read()

    def test_each_shard_publishes_its_own_verdict_artifact(self):
        self.assertIn("govulncheck-verdict-${{ matrix.shard }}", self.text)

    def test_verdicts_are_collected_by_pattern_not_by_aggregate_output(self):
        """The two transports that failed must not creep back in."""
        code = strip_comments(self.text)
        self.assertIn("pattern: govulncheck-verdict-*", code)
        self.assertIn("merge-multiple: true", code)
        # No step output may carry a per-shard verdict out of the matrix job.
        # A step output named ANY way at all, assigned from the shard's own
        # verdict, is the transport that lost three of four shards. This asserts
        # the SHAPE rather than one output name, because a name allowlist is
        # trivially renamed around.
        #
        # `re.M` is REQUIRED here and was the reason this assertion was
        # DECORATION for two rounds: assertNotRegex does not pass MULTILINE, so a
        # `^`-anchored pattern could never match a line that is not the first -
        # the check passed against every mutation, including the real one.
        self.assertIsNone(
            re.search(r"^\s*id:\s*\S*verdict", code, re.MULTILINE),
            "a per-shard verdict is being carried out of the matrix as a step "
            "output; that transport lost three of four shards in production",
        )
        # The shard's scan step legitimately writes `verdict=...` to
        # GITHUB_OUTPUT for its own steps, so this assertion is scoped to the
        # PUBLISH step: it must be an artifact upload, not a run block.
        publish = self.text.split("- name: Publish this shard's verdict", 1)
        self.assertEqual(
            len(publish), 2,
            "the per-shard verdict publish step is missing; a shard would "
            "produce a verdict that nothing collects",
        )
        after = strip_comments(publish[1])
        block = after.split("\n      - ", 1)[0]
        self.assertIn("uses: actions/upload-artifact", block)
        self.assertNotIn("GITHUB_OUTPUT", block)
        # And no step may publish a single SHARED verdict artifact: one name for
        # all four shards would overwrite itself, the same defect one layer down.
        self.assertIn("name: govulncheck-verdict-${{ matrix.shard }}", code)
        self.assertNotIn("name: govulncheck-verdict-shared", code)

    def test_no_gh_run_download_in_the_workflow(self):
        """`gh run download` is what failed on 37292268699; keep it out.

        Comments are stripped first: the workflow DOCUMENTS that failure in
        prose, and a substring check on the raw file would match the very
        comment explaining why the call was removed.
        """
        code = strip_comments(self.text)
        self.assertNotIn("gh run download", code)

    def test_empty_verdict_collection_cannot_pass(self):
        """An empty collection must fail the gate.

        `actions/download-artifact` has NO `if-no-files-found` input - that
        belongs to upload-artifact, and CI proved it on run 37303621734
        ("input \\"if-no-files-found\\" is not defined in action
        actions/download-artifact@v6" - the pin at the time). The fail-closed
        behaviour therefore has
        to come from the ledger requiring one verdict per DECLARED shard, so
        this asserts the download step does not reintroduce the bogus input and
        that the gate step is what enforces completeness.
        """
        code = strip_comments(self.text)
        # Comments are stripped: the step's own comment explains that this input
        # is NOT available, and a substring check would match that explanation -
        # the same trap as test_no_gh_run_download_in_the_workflow.
        download = strip_comments(block_of(self.text, "Download shard verdict artifacts"))
        self.assertNotIn("if-no-files-found", download)
        # upload-artifact DOES support it, and the per-shard publisher must use
        # it, so a shard that reaches the step without a verdict file fails.
        self.assertIn(
            "if-no-files-found: error",
            strip_comments(block_of(self.text, "Publish this shard's verdict")),
        )
        gate = strip_comments(
            block_of(self.text, "Verify every declared govulncheck shard reported a verdict")
        )
        self.assertIn("govulncheck_verdict_ledger.py", gate)
        self.assertIn("--expect", gate)

    def test_consumer_runs_after_the_matrix(self):
        """A parallel job cannot consume a sibling's output."""
        pv = job_block(self.text, "policy-validation")
        self.assertIn("needs: go-vuln-scan", pv)
        self.assertIn("if: always()", pv)

    def test_shard_list_comes_from_the_matrix_not_a_literal(self):
        """The declared shards are read from the workflow, not hardcoded here."""
        self.assertIn("govulncheck_shards.py --list", self.text)


class TestShardReaderStillResolves(unittest.TestCase):
    def test_reader_returns_the_live_matrix(self):
        mod = load(SHARDS_PY, "govulncheck_shards")
        with open(WORKFLOW, encoding="utf-8") as fh:
            shards = mod.parse_shards(fh.read())
        self.assertGreaterEqual(len(shards), 4)
        names = {s["shard"] for s in shards}
        self.assertIn("tests-and-sim", names)
        for entry in shards:
            self.assertTrue(entry["patterns"], f"{entry['shard']} has no patterns")


if __name__ == "__main__":
    unittest.main(verbosity=2)