"""Tests for .github/scripts/check_go_license_policy.py.

The Go License Check used to key only on `license.deniedTokens`, which made the
policy's own `license.exceptions` list inert: the two reviewed exceptions were
rejected on every run. These tests exercise the shipped script against real CSV
reports and the real policy file, so they pin the failure mode rather than
re-asserting the current tree.
"""

from __future__ import annotations

import csv
import importlib.util
import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
SCRIPT_PATH = REPO_ROOT / ".github" / "scripts" / "check_go_license_policy.py"
POLICY_PATH = REPO_ROOT / "scripts" / "supply-chain" / "go-module-policy.json"


def load_module():
    spec = importlib.util.spec_from_file_location("check_go_license_policy", SCRIPT_PATH)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


MOD = load_module()


def write_report(path: Path, rows: list[tuple[str, str]]) -> None:
    with path.open("w", newline="", encoding="utf-8") as fh:
        writer = csv.writer(fh)
        for pkg, lic in rows:
            writer.writerow([pkg, lic, ""])


class PurePredicateTests(unittest.TestCase):
    def test_exact_exception_is_covered(self) -> None:
        self.assertTrue(
            MOD.is_excepted("github.com/ethereum/go-ethereum",
                            ["github.com/ethereum/go-ethereum"])
        )

    def test_subpackage_of_an_exception_is_covered(self) -> None:
        self.assertTrue(
            MOD.is_excepted("github.com/ethereum/go-ethereum/crypto/keccak",
                            ["github.com/ethereum/go-ethereum"])
        )

    def test_sibling_with_a_shared_prefix_is_NOT_covered(self) -> None:
        # The whole point of matching on a path SEGMENT: a string-prefix test
        # would silently except an unrelated module.
        self.assertFalse(
            MOD.is_excepted("github.com/ethereum/go-ethereum-evil",
                            ["github.com/ethereum/go-ethereum"])
        )

    def test_leading_and_trailing_noise_is_normalised(self) -> None:
        self.assertTrue(
            MOD.is_excepted("./github.com/zondax/hid/libusb/libusb/",
                            ["github.com/zondax/hid/libusb/libusb"])
        )

    def test_unrelated_package_is_not_covered(self) -> None:
        self.assertFalse(
            MOD.is_excepted("github.com/foo/bar",
                            ["github.com/ethereum/go-ethereum"])
        )


class FindingTests(unittest.TestCase):
    def setUp(self) -> None:
        self.policy = json.loads(POLICY_PATH.read_text(encoding="utf-8"))

    def test_reviewed_exception_is_not_a_violation(self) -> None:
        violations, excepted = MOD.find_violations(
            [["github.com/ethereum/go-ethereum", "GPL-3.0", ""]], self.policy
        )
        self.assertEqual(violations, [])
        self.assertEqual(
            excepted, ["github.com/ethereum/go-ethereum: GPL-3.0"]
        )

    def test_zondax_libusb_exception_is_not_a_violation(self) -> None:
        violations, excepted = MOD.find_violations(
            [["github.com/zondax/hid/libusb/libusb", "LGPL-2.1", ""]], self.policy
        )
        self.assertEqual(violations, [])
        self.assertEqual(len(excepted), 1)

    def test_unexcepted_gpl_is_still_a_violation(self) -> None:
        violations, _ = MOD.find_violations(
            [["github.com/not/reviewed", "GPL-3.0", ""]], self.policy
        )
        self.assertEqual(violations, ["github.com/not/reviewed: GPL-3.0"])

    def test_permissive_license_never_flagged(self) -> None:
        violations, excepted = MOD.find_violations(
            [["github.com/foo/bar", "Apache-2.0", ""]], self.policy
        )
        self.assertEqual(violations, [])
        self.assertEqual(excepted, [])

    def test_both_real_exceptions_from_the_live_policy_are_honoured(self) -> None:
        rows = [
            ["github.com/ethereum/go-ethereum", "GPL-3.0", ""],
            ["github.com/zondax/hid/libusb/libusb", "LGPL-2.1", ""],
        ]
        violations, excepted = MOD.find_violations(rows, self.policy)
        self.assertEqual(violations, [], "declared exceptions must not be violations")
        self.assertEqual(len(excepted), 2)


class CliTests(unittest.TestCase):
    """Run the shipped script the way the workflow does."""

    def run_script(self, rows, policy_path=POLICY_PATH):
        with tempfile.TemporaryDirectory() as td:
            report = Path(td) / "licenses.csv"
            write_report(report, rows)
            return subprocess.run(
                [sys.executable, str(SCRIPT_PATH), str(report), str(policy_path)],
                capture_output=True, text=True,
            )

    def test_cli_exit_zero_when_only_exceptions_are_denied(self) -> None:
        p = self.run_script([
            ("github.com/ethereum/go-ethereum", "GPL-3.0"),
            ("github.com/zondax/hid/libusb/libusb", "LGPL-2.1"),
        ])
        self.assertEqual(p.returncode, 0, p.stderr)
        self.assertIn("excepted (reviewed", p.stdout)

    def test_cli_exit_one_on_a_real_violation(self) -> None:
        p = self.run_script([("github.com/not/reviewed", "GPL-3.0")])
        self.assertEqual(p.returncode, 1)
        self.assertIn("Disallowed Go licenses detected", p.stderr)
        self.assertIn("github.com/not/reviewed: GPL-3.0", p.stderr)

    def test_cli_reports_both_excepted_and_violations(self) -> None:
        p = self.run_script([
            ("github.com/ethereum/go-ethereum", "GPL-3.0"),
            ("github.com/not/reviewed", "AGPL-3.0"),
        ])
        self.assertEqual(p.returncode, 1)
        self.assertIn("excepted (reviewed", p.stdout)
        self.assertIn("github.com/not/reviewed: AGPL-3.0", p.stderr)

    def test_cli_rejects_a_missing_report(self) -> None:
        p = subprocess.run(
            [sys.executable, str(SCRIPT_PATH), "/nonexistent/licenses.csv",
             str(POLICY_PATH)],
            capture_output=True, text=True,
        )
        self.assertEqual(p.returncode, 2)

    def test_cli_rejects_a_policy_without_denied_tokens(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            bad = Path(td) / "policy.json"
            bad.write_text(json.dumps({"license": {"exceptions": []}}), encoding="utf-8")
            p = self.run_script([("github.com/x/y", "MIT")], policy_path=bad)
            self.assertEqual(p.returncode, 2)
            self.assertIn("deniedTokens", p.stderr)

    def test_cli_treats_a_malformed_exceptions_field_as_an_error(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            bad = Path(td) / "policy.json"
            bad.write_text(
                json.dumps({"license": {"deniedTokens": ["GPL"], "exceptions": "nope"}}),
                encoding="utf-8",
            )
            p = self.run_script([("github.com/x/y", "GPL-3.0")], policy_path=bad)
            self.assertEqual(p.returncode, 2)


class LivePolicyTests(unittest.TestCase):
    def test_live_policy_declares_the_two_reviewed_exceptions(self) -> None:
        policy = json.loads(POLICY_PATH.read_text(encoding="utf-8"))
        self.assertEqual(
            policy["license"]["exceptions"],
            [
                "github.com/ethereum/go-ethereum",
                "github.com/zondax/hid/libusb/libusb",
            ],
        )

    def test_live_policy_denies_gpl_and_lgpl_tokens(self) -> None:
        policy = json.loads(POLICY_PATH.read_text(encoding="utf-8"))
        denied = policy["license"]["deniedTokens"]
        self.assertIn("GPL", denied)
        self.assertIn("LGPL", denied)


if __name__ == "__main__":
    unittest.main()
