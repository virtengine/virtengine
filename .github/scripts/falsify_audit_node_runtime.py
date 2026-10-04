#!/usr/bin/env python3
"""Mutation harness: prove the audit's OWN tests can go red.

A green test suite proves nothing unless each test fails when the thing it
guards is broken. This mutates a COPY of `audit_node_runtime.py` once per
mutation class and requires `test_audit_node_runtime.py` to FAIL on the mutant.
The live file is never written.

The mutations target the three ways this audit could go blind, each of which is
a real defect class in this repo's history:

  M1  transitive composite resolution gutted  -> the exact miss that let
      `upload-pages-artifact@v4` (composite wrapping node20) sit in the tree.
  M2  "no manifest" demoted to a clean pass    -> a blind run reads as green.
  M3  `using: composite` no longer followed   -> composite wrapping is missed.

Run:  python .github/scripts/falsify_audit_node_runtime.py
"""

from __future__ import annotations

import os
import re
import shutil
import subprocess
import sys
import tempfile

HERE = os.path.dirname(os.path.abspath(__file__))
AUDIT_NAME = "audit_node_runtime.py"
TEST_NAME = "test_audit_node_runtime.py"
# The classifier's own unit checks. Mutation M5 lives here, so the harness has
# to run this file too: a suite that only ran TEST_NAME would report M5 as
# "mutation did not apply" forever and never notice it had grown a blind spot.
ERRORCLASS_NAME = "test_audit_error_class.py"

# Each mutation: (name, pattern, replacement, why it must be caught)
MUTATIONS = [
    (
        "M1-composite-resolution-gutted",
        r'if using == "composite":',
        'if False:',
        "with composite resolution disabled, upload-pages-artifact@v4's embedded "
        "node20 step is never seen and the one case that matters goes unreported",
    ),
    (
        "M2-missing-manifest-becomes-clean",
        r'problems\.append\(f"\{ref\.spec\} \(\{ref\.file\}:\{ref\.line\}\): no action manifest found upstream"\)',
        'pass  # demoted to a pass',
        "an unreadable pin would then be graded as a clean tree, so the audit "
        "reports success without having checked anything",
    ),
    (
        "M3-composite-branch-never-taken",
        r"if using == \"composite\":\n",
        "if using == 'composite-but-never-matches':\n",
        "the composite branch becomes dead code; only direct pins are resolved",
    ),
    (
        "M4-auth-preflight-removed",
        r'if auth\.returncode != 0:',
        "if False:",
        "without the preflight an unauthenticated run degrades into 43 identical "
        "per-pin 'API error' lines instead of one actionable refusal -- the "
        "exact symptom that shipped this job red on a clean tree",
    ),
    (
        "M5-anonymous-fallback-removed",
        r'        err = _classify\(exc\)\n        if err != "forbidden":\n            return None, err',
        '        return None, _classify(exc)',
        "without the anonymous retry a 403 from a job token scoped to this one "
        "repository is believed outright, so the audit calls itself INCOMPLETE "
        "and reds the build on pins that are plainly public and readable -- the "
        "exact failure that shipped this job red on a clean tree",
    ),
]

# The live file must be byte-identical after the run.
def sha(path: str) -> str:
    import hashlib

    with open(path, "rb") as fh:
        return hashlib.sha256(fh.read()).hexdigest()[:16]


def build_tree() -> str:
    """A full throwaway copy: the two scripts PLUS the real workflow tree.

    The workflow tree is required, not optional. The test suite's last case
    audits the SHIPPED tree, so a tree without workflows makes that case fail
    and the harness's own control reports the baseline as broken -- which
    silently destroys the meaning of every "CAUGHT" in the run.
    """
    root = tempfile.mkdtemp(prefix="auditfalsify-")
    scripts = os.path.join(root, ".github", "scripts")
    os.makedirs(scripts)
    shutil.copy(os.path.join(HERE, AUDIT_NAME), os.path.join(scripts, AUDIT_NAME))
    shutil.copy(os.path.join(HERE, TEST_NAME), os.path.join(scripts, TEST_NAME))
    shutil.copy(os.path.join(HERE, ERRORCLASS_NAME), os.path.join(scripts, ERRORCLASS_NAME))
    shutil.copytree(
        os.path.join(HERE, "..", "workflows"), os.path.join(root, ".github", "workflows")
    )
    return root


def run_tests(root: str) -> tuple[int, str]:
    """Both suites must pass. Either one going red fails the run, so a mutant
    is caught wherever its test happens to live and a broken baseline is
    visible in the control rather than hidden behind the other file."""
    combined = ""
    code = 0
    for name in (TEST_NAME, ERRORCLASS_NAME):
        proc = subprocess.run(
            [sys.executable, os.path.join(root, ".github", "scripts", name)],
            capture_output=True,
            text=True,
            timeout=900,
        )
        combined += f"--- {name} (rc={proc.returncode})\n{proc.stdout}{proc.stderr}\n"
        if proc.returncode != 0:
            code = proc.returncode
    return code, combined


def main() -> int:
    live = os.path.join(HERE, AUDIT_NAME)
    before = sha(live)

    caught, inert = [], []
    failures: list[str] = []

    for name, pattern, repl, why in MUTATIONS:
        root = build_tree()
        try:
            target = os.path.join(root, ".github", "scripts", AUDIT_NAME)
            with open(live, "r", encoding="utf-8", newline="") as fh:
                original = fh.read()
            mutated, n = re.subn(pattern, repl, original, count=1)
            if n != 1:
                inert.append(name)
                failures.append(f"{name}: mutation did not apply (pattern matched {n}x) -- harness is stale")
                print(f"  STALE {name}")
                continue
            with open(target, "w", encoding="utf-8", newline="") as fh:
                fh.write(mutated)

            code, out = run_tests(root)
            if code != 0:
                caught.append(name)
                print(f"  ok   {name}  (tests went RED)")
            else:
                inert.append(name)
                failures.append(
                    f"{name}: mutant SURVIVED -- the test suite is decoration for this class\n"
                    f"  why it matters: {why}\n{out}"
                )
                print(f"  FAIL {name}  (mutant survived)")
        finally:
            shutil.rmtree(root, ignore_errors=True)

    # CONTROL: an unmutated copy must PASS. If the tests fail on the real file
    # too, every "CAUGHT" above is meaningless.
    root = build_tree()
    try:
        code, out = run_tests(root)
        control_ok = code == 0
        print(f"  {'ok  ' if control_ok else 'FAIL'} CONTROL unmutated copy passes")
        if not control_ok:
            failures.append("CONTROL: the suite fails on the UNMUTATED file -- baseline is not green\n" + out)
    finally:
        shutil.rmtree(root, ignore_errors=True)

    after = sha(live)
    if after != before:
        failures.append(f"live file MUTATED: sha {before} -> {after}")
    print(f"  {'ok  ' if after == before else 'FAIL'} live {AUDIT_NAME} unchanged (sha256[:16]={after})")

    for f in failures:
        print("FAIL:", f)
    print(f"{len(caught)}/{len(MUTATIONS)} mutations CAUGHT, {len(inert)} inert, control {'ok' if control_ok else 'BROKEN'}")
    return 1 if failures else 0


if __name__ == "__main__":
    sys.exit(main())
