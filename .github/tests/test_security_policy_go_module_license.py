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
import re
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

# The full Apache-2.0 text, abbreviated in the grant/disclaimer sections but
# carrying every clause the gate keys on. Used where a test needs a body the
# classifier would actually identify, not a one-line mention.
APACHE_FULL = (
    "Apache License\nVersion 2.0, January 2004\n"
    "http://www.apache.org/licenses/LICENSE-2.0\n\n"
    "TERMS AND CONDITIONS FOR USE, REPRODUCTION, AND DISTRIBUTION\n\n"
    "1. Definitions.\n\n"
    '   "License" shall mean the terms and conditions for use, reproduction,\n'
    "   and distribution as defined by Sections 1 through 9 of this document.\n\n"
    '   "Licensor" shall mean the copyright owner or entity authorized by\n'
    "   the copyright owner that is granting the License.\n\n"
    "2. Grant of Copyright License. Subject to the terms and conditions of\n"
    "   this License, each Contributor hereby grants to You a perpetual,\n"
    "   worldwide, non-exclusive, no-charge, royalty-free, irrevocable\n"
    "   copyright license to reproduce, prepare Derivative Works of,\n"
    "   publicly display, publicly perform, sublicense, and distribute the\n"
    "   Work and such Derivative Works in Source or Object form.\n"
)

# Verbatim license texts for the permissive families whose canonical text is
# pure boilerplate with NO self-describing header. A gate keyed on header
# keywords ("BSD 3-Clause License") falsely rejects these - that was the
# measured defect - so each family is pinned by its real text, not a stub.
# Each is the actual text of a module in the Go module cache.
BSD_3_CLAUSE = """\
Copyright (c) 2009 The Go Authors. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
   * Neither the name of Google Inc. nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
"""

ISC = """\
Copyright (c) 2012 The Gorilla Authors. All rights reserved.

Permission to use, copy, modify, and/or distribute this software for any
purpose with or without fee is hereby granted, provided that the above
copyright notice and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
"""

MIT = """\
Copyright (c) 2013-2023 The Cobra Authors

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.
"""

ZLIB = """\
Copyright (c) 2012 Daniel Theophanes

This software is provided 'as-is', without any express or implied
warranty. In no event will the authors be held liable for any damages
arising from the use of this software.

Permission is granted to anyone to use this software for any purpose,
including commercial applications, and to alter it and redistribute it
freely, subject to the following restrictions:

   1. The origin of this software must not be misrepresented; you must not
   claim that you wrote the original software. If you use this software
   in a product, an acknowledgment in the product documentation would be
   appreciated but is not required.
"""

# A file that names licenses without granting any: prose, not a license. This
# is the shape the round-1 false green took and it must stay red, because the
# classifier cannot identify it and the module would still report "Failed to
# find license" for every one of its packages.
MENTIONS_LICENSES = (
    "This product is licensed under the Apache License, Version 2.0.\n"
    "See also the BSD 3-Clause License for bundled parts.\n"
)


POLICY_PATH = REPO_ROOT / "scripts" / "supply-chain" / "go-module-policy.json"


def _git(repo: Path, *args: str) -> None:
    subprocess.run(
        [GIT, *args],
        cwd=repo,
        check=True,
        capture_output=True,
    )


PROSE = "# sdk/go\n\nGo client for the VirtEngine node API.\n"


def build_fake_repo(root: Path, modules: dict[str, object], policy: dict | None = None) -> Path:
    """Create a minimal git repo with `modules` mapped to its license fixture.

    True writes a real Apache-2.0 LICENSE, "readme" writes a prose README.md
    carrying no license text, False leaves the module root without either, and a
    string writes that exact text as the module's LICENSE.

    The root module is given a LICENSE of its own: the gate inspects every
    module root including the repository's, so without it every case would fail
    on the root rather than on the fixture under test.
    """
    repo = root / "repo"
    (repo / "scripts" / "supply-chain").mkdir(parents=True)
    shutil.copy2(GATE_PATH, repo / "scripts" / "supply-chain" / GATE_PATH.name)

    # A single root module with no replace directives keeps the policy half of
    # the gate satisfied so the license half is what is under test.
    (repo / "go.mod").write_text("module example.com/fake\n\ngo 1.26.8\n", encoding="utf-8")
    (repo / "LICENSE").write_text(APACHE_FULL, encoding="utf-8")
    if policy is None:
        # Mirror the real policy so a family added to it is exercised here
        # without this file needing to be edited in step.
        policy = json.loads(POLICY_PATH.read_text(encoding="utf-8"))
        policy["replaces"] = {}
    (repo / "scripts" / "supply-chain" / "go-module-policy.json").write_text(
        json.dumps(policy),
        encoding="utf-8",
    )

    for module, fixture in modules.items():
        (repo / module).mkdir(parents=True, exist_ok=True)
        (repo / module / "go.mod").write_text(
            f"module example.com/{module.replace('/', '-')}\n\ngo 1.26.8\n", encoding="utf-8"
        )
        if fixture is True:
            (repo / module / "LICENSE").write_text(APACHE_FULL, encoding="utf-8")
        elif fixture == "readme":
            (repo / module / "README.md").write_text(PROSE, encoding="utf-8")
        elif isinstance(fixture, str):
            (repo / module / "LICENSE").write_text(fixture, encoding="utf-8")

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

    # --- The false-red regression -------------------------------------------
    #
    # A previous revision of the gate keyed on self-describing headers such as
    # "BSD 3-Clause License". That rejected the majority of real BSD/ISC license
    # files, which are pure boilerplate carrying no such header: measured against
    # licenses.NewClassifier(0.9) over the Go module cache, 439 of 2574 module
    # roots (17.1%) were falsely rejected, covering 5 of the 8 families the
    # policy allows. The fixtures below are verbatim real license texts, and each
    # must leave the gate green.

    def test_real_bsd_3_clause_license_text_is_accepted(self) -> None:
        """golang.org/x/text's own LICENSE: BSD-3-Clause with no self-describing
        header. This is the case a header-keyword heuristic falsely rejected.
        """
        with tempfile.TemporaryDirectory() as tmp:
            repo = build_fake_repo(Path(tmp), {"sdk/go": BSD_3_CLAUSE})
            result = self.run_gate(repo)
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_real_isc_license_text_is_accepted(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            repo = build_fake_repo(Path(tmp), {"sdk/go": ISC})
            result = self.run_gate(repo)
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_real_mit_license_text_is_accepted(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            repo = build_fake_repo(Path(tmp), {"sdk/go": MIT})
            result = self.run_gate(repo)
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_real_zlib_license_text_is_accepted(self) -> None:
        """Zlib is not in allowedFamilies, but this gate judges whether a license
        is identifiable, not whether it is allowed - the policy step in the same
        job does that. A module carrying Zlib must not be falsely rejected here,
        or the gate would duplicate and contradict the policy's own split.
        """
        with tempfile.TemporaryDirectory() as tmp:
            repo = build_fake_repo(Path(tmp), {"sdk/go": ZLIB})
            result = self.run_gate(repo)
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_every_allowed_family_is_covered_by_the_gate(self) -> None:
        """A family the policy allows but the gate cannot recognise is a gate
        contradicting the policy: it would block any module under a license this
        repo explicitly permits. license.allowedFamilies therefore drives the
        gate's clause table, and this asserts the two cannot drift.
        """
        gate = GATE_PATH.read_text(encoding="utf-8")
        table = gate.split("licenseFamilyClauses = {", 1)[1].split("\n};", 1)[0]
        families = json.loads(POLICY_PATH.read_text(encoding="utf-8"))["license"][
            "allowedFamilies"
        ]
        self.assertGreater(len(families), 1, "expected a real allowedFamilies list")
        for family in families:
            # Keys may be quoted or bare (ISC is a valid bare JS identifier), so
            # match the family as an object key rather than as a quoted string.
            self.assertRegex(
                table,
                rf'(^|[\s{{])"?{re.escape(family)}"?(\s*:)',
                f"gate has no clause entry for {family}",
            )

    def test_family_the_gate_cannot_recognise_is_a_build_failure(self) -> None:
        """The drift guard actually fires: an allowed family with no clause entry
        must fail the gate rather than silently narrowing what it accepts.
        """
        with tempfile.TemporaryDirectory() as tmp:
            policy = {
                "schemaVersion": 1,
                "license": {
                    "allowedFamilies": ["Apache-2.0", "WTFPL-1.0"],
                    "deniedTokens": ["GPL"],
                },
                "replaces": {},
            }
            repo = build_fake_repo(Path(tmp), {"sdk/go": True}, policy=policy)
            result = self.run_gate(repo)
        self.assertNotEqual(result.returncode, 0, result.stdout)
        self.assertIn("WTFPL-1.0", result.stderr)
        self.assertIn("no clause pattern", result.stderr)

    def test_readme_merely_naming_licenses_is_rejected(self) -> None:
        """Naming a license is not holding one. A README that references the
        Apache License and the BSD 3-Clause License while granting neither
        classifies as no license, so the module would still report "Failed to
        find license" for every package.
        """
        with tempfile.TemporaryDirectory() as tmp:
            repo = build_fake_repo(Path(tmp), {"sdk/go": False})
            (repo / "sdk" / "go" / "README.md").write_text(MENTIONS_LICENSES, encoding="utf-8")
            _git(repo, "add", "-A")
            result = self.run_gate(repo)
        self.assertNotEqual(result.returncode, 0, result.stdout)
        self.assertIn("sdk/go", result.stderr)
        self.assertIn("no recognisable license text", result.stderr)

    def test_license_file_named_notice_is_accepted_when_it_carries_license_text(
        self,
    ) -> None:
        """A NOTICE holding the full license text satisfies go-licenses.

        The name is only a candidate match; what carries it is the content. The
        fixture is the full Apache-2.0 text rather than a one-line mention,
        because the classifier matches whole license templates and reports "no
        license" for a NOTICE that merely names one.
        """
        with tempfile.TemporaryDirectory() as tmp:
            repo = build_fake_repo(Path(tmp), {"sdk/go": False})
            (repo / "sdk" / "go" / "NOTICE").write_text(APACHE_FULL, encoding="utf-8")
            _git(repo, "add", "-A")
            result = self.run_gate(repo)
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_notice_with_only_a_one_line_mention_is_rejected(self) -> None:
        """The other half of the case above, and the exact shape an earlier
        revision of this suite wrongly asserted as passing: a NOTICE naming a
        license without carrying its text.
        """
        with tempfile.TemporaryDirectory() as tmp:
            repo = build_fake_repo(Path(tmp), {"sdk/go": False})
            (repo / "sdk" / "go" / "NOTICE").write_text(MENTIONS_LICENSES, encoding="utf-8")
            _git(repo, "add", "-A")
            result = self.run_gate(repo)
        self.assertNotEqual(result.returncode, 0, result.stdout)
        self.assertIn("no recognisable license text", result.stderr)

    def test_spdx_identifer_file_is_accepted(self) -> None:
        """A LICENSE carrying only an SPDX tag is accepted, deliberately.

        Verified against the real classifier: licenses.NewClassifier(0.9) does
        NOT identify a bare SPDX tag (it matches whole license templates), so
        go-licenses will still report "Failed to find license" here. That is the
        accepted false-green direction - it costs a log line, never a blocked
        merge - and the tag is an unambiguous declaration of the module's own
        license, which is the thing this gate exists to establish.
        """
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
