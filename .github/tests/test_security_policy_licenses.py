from __future__ import annotations

import importlib.util
import json
import sys
import tempfile
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SCRIPT_PATH = REPO_ROOT / ".github" / "scripts" / "enforce_license_policy.py"


def load_enforcer_module():
    spec = importlib.util.spec_from_file_location("enforce_license_policy", SCRIPT_PATH)
    module = importlib.util.module_from_spec(spec)
    assert spec is not None and spec.loader is not None
    sys.modules[spec.name] = module
    spec.loader.exec_module(module)
    return module


class LicensePolicyFixTests(unittest.TestCase):
    """Tests for t_f289c053 fixes: JS null license handling and malformed Python reports."""

    @classmethod
    def setUpClass(cls) -> None:
        cls.enforcer = load_enforcer_module()

    def write_json(self, base: Path, name: str, payload: object) -> Path:
        target = base / name
        target.write_text(json.dumps(payload), encoding="utf-8")
        return target

    def fresh_dir(self) -> Path:
        temp_dir = tempfile.TemporaryDirectory()
        self.addCleanup(temp_dir.cleanup)
        return Path(temp_dir.name)

    # --- JS null license fix (t_f289c053) ---

    def test_js_null_license_is_undetermined(self) -> None:
        """Explicit JSON null in JS licenses must be treated as undetermined (fail)."""
        base = self.fresh_dir()
        report_path = base / "js_report.json"
        # license-checker can emit explicit null for packages with no license found
        report_path.write_text(
            json.dumps({"left-pad": {"version": "1.0.0", "licenses": None}}),
            encoding="utf-8",
        )
        rc, errors = self.enforcer.check_javascript_reports(str(report_path))
        self.assertEqual(rc, 1)
        self.assertTrue(any("null license" in e for e in errors))

    def test_js_string_none_is_undetermined(self) -> None:
        """Stringified 'None' must NOT slip through as valid."""
        base = self.fresh_dir()
        report_path = base / "js_report.json"
        # Old bug: json null became string "None" which passed check_package
        report_path.write_text(
            json.dumps({"left-pad": {"version": "1.0.0", "licenses": "None"}}),
            encoding="utf-8",
        )
        rc, errors = self.enforcer.check_javascript_reports(str(report_path))
        self.assertEqual(rc, 1)
        self.assertTrue(any("undetermined" in e.lower() for e in errors))

    # --- malformed Python report fix (t_f289c053) ---

    def test_malformed_python_report_dict_returns_exit_2(self) -> None:
        """A dict-shaped JSON in a Python report must produce ::error:: and exit 2."""
        base = self.fresh_dir()
        report_path = base / "py_report.json"
        # Malformed: dict instead of expected list
        report_path.write_text(
            json.dumps({"not_list": True}),
            encoding="utf-8",
        )
        rc, errors = self.enforcer.check_python_reports(str(report_path))
        self.assertEqual(rc, 2)

    def test_malformed_python_report_trajectory_is_clean(self) -> None:
        """Malformed Python reports must NOT raise AttributeError or traceback."""
        base = self.fresh_dir()
        report_path = base / "py_report.json"
        report_path.write_text(
            json.dumps({"unexpected": "dict"}),
            encoding="utf-8",
        )
        # Should not raise
        rc, _ = self.enforcer.check_python_reports(str(report_path))
        self.assertIn(rc, (0, 1, 2))


if __name__ == "__main__":
    unittest.main()
