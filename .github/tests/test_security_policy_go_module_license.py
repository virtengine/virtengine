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


PROSE = "# sdk/go\n\nGo client for the VirtEngine node API.\n"


def build_fake_repo(root: Path, modules: dict[str, object]) -> Path:
    """Create a minimal git repo with `modules` mapped to its license fixture.

    True writes a real Apache-2.0 LICENSE, "readme" writes a prose README.md
    carrying no license text, and False leaves the module root without either.
    """
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

    for module, fixture in modules.items():
        (repo / module).mkdir(parents=True, exist_ok=True)
        (repo / module / "go.mod").write_text(
            f"module example.com/{module.replace('/', '-')}\n\ngo 1.26.8\n", encoding="utf-8"
        )
        if fixture is True:
            (repo / module / "LICENSE").write_text(APACHE, encoding="utf-8")
        elif fixture == "readme":
            (repo / module / "README.md").write_text(PROSE, encoding="utf-8")

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

    def test_nested_module_with_only_a_prose_readme_fails(self) -> None:
        """go-licenses needs the file's CONTENT to classify, not just its name.

        README matches go-licenses' candidate regexp, but a prose README
        classifies as no license, so every package in such a module reports
        "Failed to find license". A name-only gate would pass here: false green.
        """
        with tempfile.TemporaryDirectory() as tmp:
            repo = build_fake_repo(Path(tmp), {"sdk/newmod": "readme"})
            result = self.run_gate(repo)
        self.assertNotEqual(result.returncode, 0, result.stdout)
        self.assertIn("sdk/newmod", result.stderr)
        self.assertIn("README.md", result.stderr)
        self.assertIn("no recognisable license text", result.stderr)

    def test_alternate_license_filenames_are_accepted(self) -> None:
        """go-licenses matches LICENSE/COPYING/NOTICE names, not just LICENSE."""
        with tempfile.TemporaryDirectory() as tmp:
            repo = build_fake_repo(Path(tmp), {"sdk/go": False})
            (repo / "sdk" / "go" / "COPYING").write_text(APACHE, encoding="utf-8")
            _git(repo, "add", "-A")
            result = self.run_gate(repo)
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_license_file_named_notice_is_accepted_when_it_carries_license_text(
        self,
    ) -> None:
        """A NOTICE holding real license text satisfies go-licenses and the gate."""
        with tempfile.TemporaryDirectory() as tmp:
            repo = build_fake_repo(Path(tmp), {"sdk/go": False})
            (repo / "sdk" / "go" / "NOTICE").write_text(
                "This product is licensed under the Apache License, Version 2.0.\n",
                encoding="utf-8",
            )
            _git(repo, "add", "-A")
            result = self.run_gate(repo)
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_spdx_identifer_file_is_accepted(self) -> None:
        """An SPDX-tagged file is classifiable even without full license text."""
        with tempfile.TemporaryDirectory() as tmp:
            repo = build_fake_repo(Path(tmp), {"sdk/go": False})
            (repo / "sdk" / "go" / "LICENSE").write_text(
                "// SPDX-License-Identifier: Apache-2.0\n", encoding="utf-8"
            )
            _git(repo, "add", "-A")
            result = self.run_gate(repo)
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_license_named_directory_is_not_accepted(self) -> None:
        """A directory called LICENSE is not a license file."""
        with tempfile.TemporaryDirectory() as tmp:
            repo = build_fake_repo(Path(tmp), {"sdk/go": False})
            (repo / "sdk" / "go" / "LICENSE").mkdir()
            _git(repo, "add", "-A")
            result = self.run_gate(repo)
        self.assertNotEqual(result.returncode, 0, result.stdout)
        self.assertIn("sdk/go", result.stderr)

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
