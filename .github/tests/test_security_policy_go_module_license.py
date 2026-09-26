"""Tests for the Go module license gate in verify-go-module-policy.mjs.

go-licenses resolves a package's license by walking up from the package
directory and stopping at the enclosing module root. A nested Go module with no
license file of its own therefore reports "Failed to find license" for every one
of its packages, which is what broke the Go License Check for sdk/go.

These tests run the real gate against a synthetic repository built in a temp
directory, so they exercise the actual failure mode rather than re-asserting the
current tree happens to be correct.
"""

from __future__ import annotations

import json
import shutil
import subprocess
import tempfile
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
GATE_PATH = REPO_ROOT / "scripts" / "supply-chain" / "verify-go-module-policy.mjs"
NODE = shutil.which("node") or shutil.which("nodejs")
GIT = shutil.which("git")

APACHE = "Apache License\nVersion 2.0, January 2004\n"


def _git(repo: Path, *args: str) -> None:
    subprocess.run(
        [GIT, *args],
        cwd=repo,
        check=True,
        capture_output=True,
    )


def build_fake_repo(root: Path, modules: dict[str, bool]) -> Path:
    """Create a minimal git repo with `modules` mapped to has_license."""
    repo = root / "repo"
    (repo / "scripts" / "supply-chain").mkdir(parents=True)
    shutil.copy2(GATE_PATH, repo / "scripts" / "supply-chain" / GATE_PATH.name)

    # A single root module with no replace directives keeps the policy half of
    # the gate satisfied so the license half is what is under test.
    (repo / "go.mod").write_text("module example.com/fake\n\ngo 1.26.8\n", encoding="utf-8")
    (repo / "LICENSE").write_text(APACHE, encoding="utf-8")
    (repo / "scripts" / "supply-chain" / "go-module-policy.json").write_text(
        json.dumps(
            {
                "schemaVersion": 1,
                "license": {
                    "allowedFamilies": ["Apache-2.0"],
                    "deniedTokens": ["GPL"],
                },
                "replaces": {},
            }
        ),
        encoding="utf-8",
    )

    for module, has_license in modules.items():
        (repo / module).mkdir(parents=True, exist_ok=True)
        (repo / module / "go.mod").write_text(
            f"module example.com/{module.replace('/', '-')}\n\ngo 1.26.8\n", encoding="utf-8"
        )
        if has_license:
            (repo / module / "LICENSE").write_text(APACHE, encoding="utf-8")

    _git(repo, "init", "-q")
    _git(repo, "add", "-A")
    return repo


@unittest.skipIf(NODE is None, "node is required to run the Go module license gate")
@unittest.skipIf(GIT is None, "git is required to build the fixture repositories")
class GoModuleLicenseGateTests(unittest.TestCase):
    def run_gate(self, repo: Path) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            [NODE, f"scripts/supply-chain/{GATE_PATH.name}"],
            cwd=repo,
            capture_output=True,
            text=True,
        )

    def test_nested_module_without_license_fails(self) -> None:
        """The regression: sdk/go had no LICENSE and the gate must reject that."""
        with tempfile.TemporaryDirectory() as tmp:
            repo = build_fake_repo(Path(tmp), {"sdk/go": False})
            result = self.run_gate(repo)
        self.assertNotEqual(result.returncode, 0, result.stdout)
        self.assertIn("sdk/go", result.stderr)
        self.assertIn("no license file", result.stderr)

    def test_nested_module_with_license_passes(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            repo = build_fake_repo(Path(tmp), {"sdk/go": True, "sdk/specs": True})
            result = self.run_gate(repo)
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_alternate_license_filenames_are_accepted(self) -> None:
        """go-licenses matches LICENSE/COPYING/NOTICE, not just LICENSE."""
        with tempfile.TemporaryDirectory() as tmp:
            repo = build_fake_repo(Path(tmp), {"sdk/go": False})
            (repo / "sdk" / "go" / "COPYING").write_text(APACHE, encoding="utf-8")
            _git(repo, "add", "-A")
            result = self.run_gate(repo)
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_every_tracked_module_in_this_repo_has_a_license(self) -> None:
        """Guard the real tree, not just the synthetic fixture."""
        modules = subprocess.run(
            [GIT, "ls-files", "**/go.mod", "go.mod"],
            cwd=REPO_ROOT,
            check=True,
            capture_output=True,
            text=True,
        ).stdout.split()
        self.assertGreater(len(modules), 1, "expected nested Go modules in this repo")
        missing = [
            module
            for module in modules
            if not any(
                (REPO_ROOT / Path(module).parent / name).exists()
                for name in ("LICENSE", "LICENSE.md", "COPYING", "NOTICE")
            )
        ]
        self.assertEqual(missing, [], f"Go module roots missing a license file: {missing}")

    def test_gate_is_wired_into_the_license_compliance_workflow(self) -> None:
        workflow = (REPO_ROOT / ".github" / "workflows" / "license-compliance.yaml").read_text(
            encoding="utf-8"
        )
        self.assertIn("scripts/supply-chain/verify-go-module-policy.mjs", workflow)


if __name__ == "__main__":
    unittest.main()
