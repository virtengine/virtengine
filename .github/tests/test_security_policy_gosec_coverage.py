from __future__ import annotations

import importlib.util
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SCRIPT_PATH = REPO_ROOT / ".github" / "scripts" / "check_gosec_coverage.py"


def load_checker_module():
    spec = importlib.util.spec_from_file_location("check_gosec_coverage", SCRIPT_PATH)
    module = importlib.util.module_from_spec(spec)
    assert spec is not None and spec.loader is not None
    sys.modules[spec.name] = module
    spec.loader.exec_module(module)
    return module


checker = load_checker_module()


class ParseLogTests(unittest.TestCase):
    def test_extracts_all_three_line_kinds_and_normalises_windows_paths(self):
        log = "\n".join(
            [
                "[gosec] 2026/09/19 04:08:10 Import directory: C:\\repo\\x\\veid\\keeper",
                "[gosec] 2026/09/19 04:08:10 Checking package: keeper",
                "[gosec] 2026/09/19 04:08:11 Checking file: C:\\repo\\x\\veid\\keeper\\keeper.go",
                "[gosec] 2026/09/19 04:08:11 Ignoring generated file: C:\\repo\\pkg\\gen\\types.go",
                "[gosec] 2026/09/19 04:08:11 Checking file: /repo/pkg/a/a.go",
            ]
        )
        imported, checked, generated = checker.parse_log(log)
        self.assertEqual(imported, {"C:/repo/x/veid/keeper"})
        self.assertEqual(
            checked, {"C:/repo/x/veid/keeper/keeper.go", "/repo/pkg/a/a.go"}
        )
        self.assertEqual(generated, {"C:/repo/pkg/gen/types.go"})


class FindDroppedTests(unittest.TestCase):
    """The decision that turns a silently-shrunken scan into a red build."""

    mod = "/repo"

    def test_package_with_no_analysed_file_is_dropped(self):
        pkgs = [("/repo/x/veid/keeper", ["keeper.go", "params.go"])]
        dropped = checker.find_dropped(pkgs, self.mod, [], set(), set())
        self.assertEqual(dropped, [("/repo/x/veid/keeper", [
            "/repo/x/veid/keeper/keeper.go", "/repo/x/veid/keeper/params.go",
        ])])

    def test_fully_analysed_package_is_not_dropped(self):
        pkgs = [("/repo/x/veid/keeper", ["keeper.go"])]
        checked = {"/repo/x/veid/keeper/keeper.go"}
        self.assertEqual(checker.find_dropped(pkgs, self.mod, [], checked, set()), [])

    def test_package_whose_files_are_all_generated_is_not_dropped(self):
        pkgs = [("/repo/pkg/gen", ["types.go"])]
        generated = {"/repo/pkg/gen/types.go"}
        self.assertEqual(checker.find_dropped(pkgs, self.mod, [], set(), generated), [])

    def test_explicitly_excluded_directory_is_not_dropped(self):
        pkgs = [("/repo/testutil/sims", ["app.go"])]
        self.assertEqual(checker.find_dropped(pkgs, self.mod, ["testutil"], set(), set()), [])

    def test_exclusion_matches_nested_name_not_only_toplevel(self):
        pkgs = [("/repo/pkg/keymanagement/hsm/testutil", ["softhsm.go"])]
        self.assertEqual(checker.find_dropped(pkgs, self.mod, ["testutil"], set(), set()), [])


@unittest.skipUnless(shutil.which("go"), "go toolchain required for the end-to-end check")
class EndToEndTests(unittest.TestCase):
    """Runs the real script: a partial scan must fail, a complete scan must pass."""

    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.root = Path(self.tmp.name)
        (self.root / "go.mod").write_text("module example.com/m\n\ngo 1.21\n", encoding="utf-8")
        for name in ("pkga", "pkgb"):
            d = self.root / name
            d.mkdir()
            (d / f"{name}.go").write_text(
                f"package {name}\n\nfunc F() int {{ return 1 }}\n", encoding="utf-8"
            )
        # go list emits native separators; the checker normalises both sides.
        self.dir_a = (self.root / "pkga").as_posix()
        self.dir_b = (self.root / "pkgb").as_posix()

    def tearDown(self):
        self.tmp.cleanup()

    def run_checker(self, log_text: str) -> subprocess.CompletedProcess:
        log = self.root / "gosec.log"
        log.write_text(log_text, encoding="utf-8")
        return subprocess.run(
            [sys.executable, str(SCRIPT_PATH), "--log", str(log),
             "--module-dir", str(self.root)],
            capture_output=True, text=True,
        )

    def test_partial_scan_fails_and_names_the_skipped_package(self):
        proc = self.run_checker(f"[gosec] Checking file: {self.dir_a}/pkga.go\n")
        out = proc.stdout + proc.stderr
        self.assertEqual(proc.returncode, 1, out)
        # only the package gosec never analysed is named, not the one it did
        self.assertIn("pkgb", proc.stderr)
        self.assertNotIn("pkga", proc.stderr)

    def test_complete_scan_passes(self):
        log = (
            f"[gosec] Checking file: {self.dir_a}/pkga.go\n"
            f"[gosec] Checking file: {self.dir_b}/pkgb.go\n"
        )
        proc = self.run_checker(log)
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        self.assertIn("OK:", proc.stdout)

    def test_log_without_any_analysed_file_fails(self):
        proc = self.run_checker("[gosec] Including rules: default\n")
        self.assertEqual(proc.returncode, 1)
        self.assertIn("produced nothing", proc.stderr)


if __name__ == "__main__":
    unittest.main()
