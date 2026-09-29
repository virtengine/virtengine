#!/usr/bin/env python3
"""Regression tests for .github/scripts/filter_allowlist.py.

Two defects are pinned here:

1. REPORTING. When every finding was allowlisted the script printed only a
   count ("Total findings: 3, After allowlist: 0"), never naming the advisories
   it had suppressed, so a reviewer could not tell WHAT the allowlist was
   hiding on a green run.

2. FAIL-OPEN. extract_findings() used to skip past an undecodable object with
   content.find('{', idx+1). A report cut off mid-write - the exact shape left
   by the runner shutdown that stopped govulncheck from ever reaching a verdict
   - therefore parsed as "no findings" and exited 0, i.e. the gate reported
   CLEAN for a scan that never finished. It now exits 4 (incomplete).

Run: python -m unittest discover -s .github/tests -p "test_filter_allowlist*.py"
"""

import importlib.util
import json
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
SCRIPT = REPO_ROOT / ".github" / "scripts" / "filter_allowlist.py"

EXIT_CLEAN = 0
EXIT_USAGE = 2
EXIT_VULNS = 3
EXIT_INCOMPLETE = 4


def load_script():
    spec = importlib.util.spec_from_file_location("filter_allowlist", SCRIPT)
    if spec is None or spec.loader is None:  # import plumbing, not a code path
        raise RuntimeError(f"cannot load {SCRIPT}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


ALLOWLIST_YAML = """
version: 2
policy:
  max_allowlist_age_days: 30
exceptions:
  go:
  - id: GO-2026-4740
    reason: reviewed, time-boxed
  - id: GO-2026-5932
    reason: reviewed, time-boxed
  python: []
  npm: []
"""


def write_report(path, findings, *, tail="", config=True):
    """govulncheck -format json emits a STREAM of concatenated JSON objects."""
    lines = []
    if config:
        lines.append(json.dumps({"config": {"scanner": {"name": "govulncheck"}}}))
    lines.append(json.dumps({"sbom": {"modules": []}}))
    for osv in findings:
        lines.append(json.dumps({
            "finding": {"osv": osv, "fixed_version": "", "trace": [{"module": "m"}]}
        }))
    lines.append(json.dumps({"progress": {"message": "Scanning..."}}))
    body = "\n".join(lines)
    if tail:
        body += "\n" + tail
    Path(path).write_text(body + "\n", encoding="utf-8")


class FilterAllowlistTest(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        cls.tmp = tempfile.TemporaryDirectory()
        cls.allowlist = os.path.join(cls.tmp.name, "allowlist.yaml")
        with open(cls.allowlist, "w", encoding="utf-8") as f:
            f.write(ALLOWLIST_YAML)

    @classmethod
    def tearDownClass(cls):
        cls.tmp.cleanup()

    def run_script(self, report_path):
        proc = subprocess.run(
            [sys.executable, str(SCRIPT), str(report_path), self.allowlist],
            capture_output=True, text=True,
        )
        return proc.returncode, proc.stdout, proc.stderr

    def report(self, name, findings, **kw):
        path = os.path.join(self.tmp.name, name)
        write_report(path, findings, **kw)
        return path

    # --- defect 1: reporting -------------------------------------------

    def test_all_allowlisted_names_every_detected_advisory(self):
        """DONE WHEN #3: the all-allowlisted case must name what it suppressed."""
        code, out, err = self.run_script(
            self.report("all_allowed.json", ["GO-2026-4740", "GO-2026-4740", "GO-2026-5932"])
        )
        self.assertEqual(code, EXIT_CLEAN)
        self.assertEqual(out.strip(), "", "nothing un-allowlisted should reach stdout")
        self.assertIn("GO-2026-4740", err)
        self.assertIn("GO-2026-5932", err)
        self.assertIn("none", err.splitlines()[-1])

    def test_duplicate_findings_collapse_to_unique_ids(self):
        code, _, err = self.run_script(
            self.report("dupes.json", ["GO-2026-4740"] * 9)
        )
        self.assertEqual(code, EXIT_CLEAN)
        self.assertIn("Advisories detected (1)", err)
        self.assertIn("GO-2026-4740", err)

    def test_mixed_reports_separate_detected_from_unallowlisted(self):
        code, out, err = self.run_script(
            self.report("mixed.json", ["GO-2026-4740", "GO-9999-0001", "GO-9999-0001"])
        )
        self.assertEqual(code, EXIT_VULNS)
        self.assertIn("Advisories detected (2)", err)
        self.assertIn("Un-allowlisted advisories (1)", err)
        self.assertIn("GO-9999-0001", err)
        self.assertNotIn('"osv": "GO-2026-4740"', out,
                         "an allowlisted finding must not be re-emitted")

    def test_no_findings_reports_none_cleanly(self):
        code, out, err = self.run_script(self.report("none.json", []))
        self.assertEqual(code, EXIT_CLEAN)
        self.assertEqual(out.strip(), "")
        self.assertIn("none", err)

    # --- defect 2: fail-open -------------------------------------------

    def test_incomplete_report_never_reports_clean(self):
        """A half-written report must not exit 0 - that is the original
        'gate never reached a verdict' defect re-entering through the filter."""
        path = os.path.join(self.tmp.name, "cut_mid_write.json")
        full = self.report("cut_src.json", ["GO-9999-0001"])
        raw = Path(full).read_text(encoding="utf-8").splitlines()
        # Keep everything except the last object, then cut that one in half.
        Path(path).write_text("\n".join(raw[:-1] + [raw[-1][:30]]) + "\n", encoding="utf-8")
        code, _, err = self.run_script(path)
        self.assertEqual(code, EXIT_INCOMPLETE,
                         "an unparseable stream must exit 4, never 0 or 3")
        self.assertIn("INCOMPLETE", err)

    def test_incomplete_report_with_earlier_findings_still_refuses_verdict(self):
        path = os.path.join(self.tmp.name, "cut_after_findings.json")
        raw = Path(self.report("cut_src2.json", ["GO-9999-0001", "GO-2026-4740"])) \
            .read_text(encoding="utf-8").splitlines()
        Path(path).write_text("\n".join(raw[:-1] + [raw[-1][:25]]) + "\n", encoding="utf-8")
        code, _, err = self.run_script(path)
        self.assertEqual(code, EXIT_INCOMPLETE)
        self.assertIn("INCOMPLETE", err)

    def test_incomplete_report_still_names_advisories_it_did_decode(self):
        path = os.path.join(self.tmp.name, "cut_names.json")
        raw = Path(self.report("cut_src3.json", ["GO-9999-0001", "GO-2026-4740"])) \
            .read_text(encoding="utf-8").splitlines()
        # Truncate so both findings decode, then the stream breaks.
        Path(path).write_text("\n".join(raw[:3]) + '\n{"progress": {"mess', encoding="utf-8")
        code, _, err = self.run_script(path)
        self.assertEqual(code, EXIT_INCOMPLETE)
        self.assertIn("GO-9999-0001", err)
        self.assertIn("decoded before the break", err)

    def test_garbage_input_is_incomplete_not_clean(self):
        path = os.path.join(self.tmp.name, "garbage.json")
        Path(path).write_text("this is not json at all\n", encoding="utf-8")
        code, _, err = self.run_script(path)
        self.assertEqual(code, EXIT_INCOMPLETE)
        self.assertIn("INCOMPLETE", err)

    def test_empty_report_is_clean(self):
        """An empty file is a real, complete, zero-finding scan - not a failure."""
        path = os.path.join(self.tmp.name, "empty.json")
        Path(path).write_text("", encoding="utf-8")
        code, _, err = self.run_script(path)
        self.assertEqual(code, EXIT_CLEAN)

    # --- error handling ------------------------------------------------

    def test_missing_report_is_error(self):
        code, _, err = self.run_script(os.path.join(self.tmp.name, "nope.json"))
        self.assertEqual(code, EXIT_USAGE)
        self.assertIn("not found", err)

    def test_missing_allowlist_is_a_readable_error_not_a_crash(self):
        """An unreadable allowlist must produce a clean exit 2 with a message.

        Asserting only "not 0" is too weak to pin anything: the pre-fix script
        satisfies that by letting an unhandled FileNotFoundError traceback exit
        1. The contract is a REPORTED error, not a crash.
        """
        proc = subprocess.run(
            [sys.executable, str(SCRIPT), self.report("w_al.json", ["GO-9999-0001"]),
             os.path.join(self.tmp.name, "no-such-allowlist.yaml")],
            capture_output=True, text=True,
        )
        self.assertEqual(proc.returncode, EXIT_USAGE)
        self.assertIn("allowlist", proc.stderr)
        self.assertNotIn("Traceback", proc.stderr,
                         "a missing allowlist must be reported, not crash")

    def test_no_arguments_is_usage_error(self):
        proc = subprocess.run([sys.executable, str(SCRIPT)], capture_output=True, text=True)
        self.assertEqual(proc.returncode, EXIT_USAGE)
        self.assertIn("usage", proc.stderr)

    # --- unit-level ----------------------------------------------------

    def test_extract_findings_reports_completeness(self):
        mod = load_script()
        findings, complete = mod.extract_findings(
            json.dumps({"finding": {"osv": "GO-1"}}) + "\n" + json.dumps({"progress": {}})
        )
        self.assertTrue(complete)
        self.assertEqual(len(findings), 1)

        _, complete_bad = mod.extract_findings(json.dumps({"finding": {"osv": "GO-1"}})[:-5])
        self.assertFalse(complete_bad,
                         "a half-decodable stream must be marked incomplete")

    def test_unique_ids_skips_findings_without_osv(self):
        mod = load_script()
        self.assertEqual(
            mod.unique_ids([{"osv": "GO-B"}, {"osv": None}, {}, {"osv": "GO-A"}]),
            ["GO-A", "GO-B"],
        )

    def test_load_allowlist_tolerates_missing_exceptions_key(self):
        mod = load_script()
        path = os.path.join(self.tmp.name, "bare.yaml")
        with open(path, "w", encoding="utf-8") as f:
            f.write("version: 2\n")
        self.assertEqual(mod.load_allowlist(Path(path)), set())


if __name__ == "__main__":
    unittest.main(verbosity=2)
