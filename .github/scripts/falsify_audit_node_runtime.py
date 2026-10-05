#!/usr/bin/env python3
"""Mutation harness: prove the audit's OWN tests can go red.

A green test suite proves nothing unless each test fails when the thing it
guards is broken. This mutates a COPY of `audit_node_runtime.py` once per
mutation class and requires `test_audit_node_runtime.py` to FAIL on the mutant.
The live file is never written.

The mutations target every way this audit could go blind, each of which is a
real defect class in this repo's history:

  M1  transitive composite resolution gutted  -> the exact miss that let
      `upload-pages-artifact@v4` (composite wrapping node20) sit in the tree.
  M2  "no manifest" demoted to a clean pass    -> a blind run reads as green.
  M3  `using: composite` no longer followed   -> composite wrapping is missed.
  M4  auth preflight removed                  -> 43 identical errors, no cause.
  M5  anonymous fallback removed              -> a 403 believed outright.
  M6  input check unwired again               -> the state this audit SHIPPED
      in: `consumed_inputs()` defined, never called, docstring promising it.
  M7  input finding kept out of the verdict   -> printed, logged, never gated on.
  M8  composite inputs made enforceable       -> 10 false positives on a clean
      tree, i.e. a gate that gets switched off rather than fixed.
  M9  `with:` block swallowing the next step  -> an action audited by nothing.

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
    (
        # M6 is the mutation that reopens THIS card. `consumed_inputs()` shipped
        # as dead code for four months while the docstring promised this check,
        # and it took a reviewer reading the source to notice -- no test, no
        # mutation, nothing failed. Without a mutation class for it, the same
        # silence is available to the next person who touches main().
        "M6-input-check-unwired-again",
        r"    bad_inputs, input_skipped, input_transient = undeclared_input_findings\(refs\)",
        "    bad_inputs, input_skipped, input_transient = [], [], []",
        "with the input sweep stubbed out, an undeclared input is neither reported "
        "nor able to fail the gate -- exactly the state the audit shipped in, and "
        "the state its docstring promised did not exist",
    ),
    (
        "M7-input-finding-not-in-the-verdict",
        r"    if bad_inputs:\n        print\(f\"RESULT: \{len\(bad_inputs\)\} undeclared input",
        "    if False:\n        print(f\"RESULT: {len(bad_inputs)} undeclared input",
        "the finding is computed and printed but never reaches the exit code, so a "
        "bump that passes a dropped input is reported in the log and still grades "
        "the job green -- printing a finding is not gating on it",
    ),
    (
        "M8-composite-inputs-treated-as-enforceable",
        r"        using, declared = parse_action_yml\(text\)",
        '        using, declared = parse_action_yml(text)\n'
        '        if not using.startswith("node"):\n'
        '            using = "node24"',
        "without the composite guard, runs.steps keys are read against the action's "
        "declared inputs and every composite action in the tree is reported as "
        "passing undeclared inputs -- false positives on a clean tree, i.e. a gate "
        "nobody can keep on, which is how the real one was switched off in the first "
        "place. Mutation note: neutering the `if` itself does NOT work here, it just "
        "falls through to the ENFORCEABLE_USINGS guard below, which also rejects "
        "composite. The mutant has to defeat BOTH guards.",
    ),
    (
        "M9-with-block-swallows-the-next-step",
        r"        if in_with and \(indent <= with_indent or stripped\.startswith\(\"-\"\)\):",
        "        if in_with and indent <= with_indent and not stripped.startswith(\"-\"):",
        "the `- uses:` one-line step form after a `with:` block is swallowed, so a "
        "whole action pin is never recorded and is audited by nothing at all -- not "
        "its runtime, not its inputs",
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
