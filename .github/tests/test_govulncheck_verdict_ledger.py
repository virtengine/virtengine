#!/usr/bin/env python3
"""Regression tests for the fail-closed govulncheck verdict ledger.

`t_76147140` requires that a shard which produced no verdict is reported as a
distinct outcome - neither "clean" nor "advisory found". These tests drive the
real `.github/scripts/govulncheck_verdict_ledger.py` over fixture ledgers, so a
pass here means the shipped script behaves, not that a copy of it would.

Run: python -m unittest discover -s .github/tests -p "test_govulncheck_verdict_ledger*.py"
"""
from __future__ import annotations

import importlib.util
import os
import subprocess
import sys
import tempfile
import unittest

REPO_ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
LEDGER = os.path.join(REPO_ROOT, ".github", "scripts", "govulncheck_verdict_ledger.py")


def load_module():
    """Import the shipped script by path, not via sys.path games."""
    spec = importlib.util.spec_from_file_location("govulncheck_verdict_ledger", LEDGER)
    if spec is None or spec.loader is None:
        raise unittest.SkipTest(f"cannot load {LEDGER}")
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


class LedgerFixture(unittest.TestCase):
    """Builds a ledger directory and asserts the gate's exit + classification."""

    def setUp(self):
        self.mod = load_module()
        self._tmp = tempfile.TemporaryDirectory()
        self.dir = self._tmp.name

    def tearDown(self):
        self._tmp.cleanup()

    def manifest(self, *shards: str) -> None:
        with open(os.path.join(self.dir, "shards.manifest"), "w", encoding="utf-8") as fh:
            fh.write("\n".join(shards) + "\n")

    def verdict(self, shard: str, value: str) -> None:
        with open(os.path.join(self.dir, f"{shard}.verdict"), "w", encoding="utf-8") as fh:
            fh.write(f"verdict={value}\n")

    def empty_file(self, shard: str) -> None:
        open(os.path.join(self.dir, f"{shard}.verdict"), "w", encoding="utf-8").close()

    def run_gate(self, expect: int | None = None) -> int:
        argv = [self.dir]
        if expect is not None:
            argv += ["--expect", str(expect)]
        return self.mod.main(argv)


class TestNoVerdictIsNotClean(LedgerFixture):
    """The card's core requirement, stated as three separate assertions."""

    def test_missing_shard_fails_closed(self):
        self.manifest("core", "platform", "tooling")
        self.verdict("core", "clean")
        self.verdict("platform", "clean")
        # `tooling` never wrote a verdict - it was cancelled mid-scan.
        self.assertEqual(self.run_gate(), 1)

    def test_missing_shard_is_not_reported_clean(self):
        self.manifest("core", "platform")
        self.verdict("core", "clean")
        verdict, note = self.mod.read_shard_ledger(self.dir, "platform")
        self.assertIsNone(verdict, "a verdict-less shard must classify as None")
        self.assertIn("no verdict file", note)

    def test_missing_shard_is_not_reported_as_a_finding(self):
        """UNKNOWN must be distinct from `vulns`: neither clean nor a finding."""
        self.manifest("core", "platform")
        self.verdict("core", "clean")
        verdict, _ = self.mod.read_shard_ledger(self.dir, "platform")
        self.assertNotEqual(verdict, "vulns")
        self.assertNotEqual(verdict, "clean")
        self.assertNotEqual(verdict, "error")

    def test_error_exit_carries_the_no_verdict_annotation(self):
        """The message must name the count, so a red is never read as a finding."""
        import io
        import contextlib

        self.manifest("core", "platform")
        self.verdict("core", "clean")
        err = io.StringIO()
        with contextlib.redirect_stderr(err):
            self.run_gate()
        text = err.getvalue()
        self.assertIn("::error::", text)
        self.assertIn("produced NO", text)
        self.assertIn("NOT clean", text)


class TestVerdictsAreClassified(LedgerFixture):
    def test_all_clean_passes(self):
        self.manifest("core", "platform", "tooling", "tests-and-sim")
        for shard in ("core", "platform", "tooling", "tests-and-sim"):
            self.verdict(shard, "clean")
        self.assertEqual(self.run_gate(), 0)

    def test_each_known_verdict_is_accepted(self):
        for value in ("clean", "vulns", "error", "incomplete"):
            with self.subTest(verdict=value):
                self.setUp()
                self.manifest("core")
                self.verdict("core", value)
                # The ledger's job is completeness, not severity: every shard
                # DID report, so it passes and the matrix result carries the
                # red. It must not silently convert a red into a green.
                self.assertEqual(self.run_gate(), 0)

    def test_unknown_verdict_value_fails_closed(self):
        self.manifest("core")
        self.verdict("core", "probably-fine")
        self.assertEqual(self.run_gate(), 1)

    def test_duplicate_verdict_lines_fail_closed(self):
        """Two verdicts for one shard is ambiguous, so it must not be trusted."""
        self.manifest("core")
        with open(os.path.join(self.dir, "core.verdict"), "w", encoding="utf-8") as fh:
            fh.write("verdict=clean\nverdict=vulns\n")
        self.assertEqual(self.run_gate(), 1)

    def test_empty_verdict_file_is_not_clean(self):
        self.manifest("core")
        self.empty_file("core")
        self.assertEqual(self.run_gate(), 1)

    def test_verdict_file_without_the_key_fails_closed(self):
        self.manifest("core")
        with open(os.path.join(self.dir, "core.verdict"), "w", encoding="utf-8") as fh:
            fh.write("I finished the scan, all good\n")
        self.assertEqual(self.run_gate(), 1)


class TestFailsClosedOnItsOwnInputs(LedgerFixture):
    """A gate that cannot read its inputs must not report a pass."""

    def test_no_manifest_is_rejected(self):
        self.verdict("core", "clean")
        self.assertEqual(self.run_gate(), 2)

    def test_expect_without_manifest_is_rejected(self):
        """A count cannot be attributed to a shard, so nothing may be credited."""
        self.verdict("core", "clean")
        self.assertEqual(self.run_gate(expect=1), 2)

    def test_manifest_count_mismatch_fails(self):
        self.manifest("core", "platform")
        for shard in ("core", "platform"):
            self.verdict(shard, "clean")
        self.assertEqual(self.run_gate(expect=3), 1)

    def test_missing_directory_is_a_usage_error(self):
        self.assertEqual(self.mod.main([os.path.join(self.dir, "nope")]), 2)

    def test_empty_manifest_is_rejected(self):
        self.manifest()
        self.assertEqual(self.run_gate(), 2)


class TestManifestParsing(LedgerFixture):
    def test_comments_and_blank_lines_are_ignored(self):
        with open(os.path.join(self.dir, "shards.manifest"), "w", encoding="utf-8") as fh:
            fh.write("# declared shards\n\n  core  \nplatform\n")
        names = self.mod.expected_shards(None, self.dir)
        self.assertEqual(names, ["core", "platform"])

    def test_cli_exit_codes(self):
        """Drive the real CLI, not only the importable functions."""
        self.manifest("core", "platform")
        self.verdict("core", "clean")
        script = [sys.executable, LEDGER, self.dir, "--expect", "2"]
        proc = subprocess.run(script, capture_output=True, text=True)
        self.assertEqual(proc.returncode, 1, proc.stderr)
        self.assertIn("NO", proc.stderr)

        self.verdict("platform", "clean")
        proc = subprocess.run(script, capture_output=True, text=True)
        self.assertEqual(proc.returncode, 0, proc.stderr)
        self.assertIn("reported a verdict", proc.stdout)


class TestVerdictsJson(LedgerFixture):
    """The transport the workflow actually uses: matrix job outputs.

    The file-based ledger was superseded because a cross-job artifact download
    failed on run 37292268699 even though every shard succeeded. These pin the
    replacement, including the empty-output case GitHub produces when the whole
    matrix is cancelled.
    """

    def verdicts_json(self, body: str) -> str:
        path = os.path.join(self.dir, "verdicts.json")
        with open(path, "w", encoding="utf-8") as fh:
            fh.write(body)
        return path

    def test_all_verdicts_reported_passes(self):
        self.manifest("core", "platform", "tooling")
        path = self.verdicts_json(
            '{"core":"clean","platform":"clean","tooling":"clean"}'
        )
        self.assertEqual(self.mod.main([self.dir, "--verdicts-json", path,
                                        "--expect", "3"]), 0)

    def test_cancelled_shard_absent_from_the_object_fails_closed(self):
        """The exact recurrence: one shard produced no output at all."""
        self.manifest("core", "platform", "tooling")
        path = self.verdicts_json('{"core":"clean","platform":"clean"}')
        self.assertEqual(self.mod.main([self.dir, "--verdicts-json", path,
                                        "--expect", "3"]), 1)

    def test_empty_output_means_every_shard_is_unknown(self):
        """GitHub emits an EMPTY STRING, not `{}`, when nothing reported."""
        self.manifest("core", "platform")
        path = self.verdicts_json("")
        self.assertEqual(self.mod.main([self.dir, "--verdicts-json", path,
                                        "--expect", "2"]), 1)

    def test_whitespace_only_output_is_treated_as_empty(self):
        self.manifest("core")
        path = self.verdicts_json("  \n ")
        self.assertEqual(self.mod.main([self.dir, "--verdicts-json", path,
                                        "--expect", "1"]), 1)

    def test_no_verdict_value_fails_closed(self):
        """A shard can complete its step and still not reach a verdict."""
        self.manifest("core")
        path = self.verdicts_json('{"core":"no-verdict"}')
        self.assertEqual(self.mod.main([self.dir, "--verdicts-json", path,
                                        "--expect", "1"]), 1)

    def test_malformed_json_is_a_usage_error_not_a_pass(self):
        self.manifest("core")
        path = self.verdicts_json('{"core": ')
        self.assertEqual(self.mod.main([self.dir, "--verdicts-json", path,
                                        "--expect", "1"]), 2)

    def test_json_array_instead_of_object_is_rejected(self):
        self.manifest("core")
        path = self.verdicts_json('["clean"]')
        self.assertEqual(self.mod.main([self.dir, "--verdicts-json", path,
                                        "--expect", "1"]), 2)

    def test_missing_json_file_yields_all_unknown(self):
        self.manifest("core", "platform")
        path = os.path.join(self.dir, "does-not-exist.json")
        self.assertEqual(self.mod.main([self.dir, "--verdicts-json", path,
                                        "--expect", "2"]), 1)

    def test_unknown_verdict_value_in_json_fails_closed(self):
        self.manifest("core")
        path = self.verdicts_json('{"core":"looks-fine"}')
        self.assertEqual(self.mod.main([self.dir, "--verdicts-json", path,
                                        "--expect", "1"]), 1)

    def test_extra_shard_in_json_is_not_credited(self):
        """A shard the workflow no longer declares must not satisfy the count."""
        self.manifest("core")
        path = self.verdicts_json('{"core":"clean","ghost":"clean"}')
        self.assertEqual(self.mod.main([self.dir, "--verdicts-json", path,
                                        "--expect", "1"]), 0)


if __name__ == "__main__":
    unittest.main(verbosity=2)