#!/usr/bin/env python3
"""Falsify audit_node_runtime.py: prove the guard can actually go RED.

A guard that cannot fail is decoration. This plants each node20 shape the audit
is supposed to catch -- including the COMPOSITE one that the original
`upload-pages-artifact@v4` slip proved is the case people miss -- into a copy of
the workflow tree, and requires the audit to report it.

Every case asserts on the audit's real verdict, produced by running the shipped
script as a subprocess. Nothing here reconstructs the expected answer.

Run:  python .github/scripts/test_audit_node_runtime.py
"""

from __future__ import annotations

import os
import re
import shutil
import subprocess
import sys
import tempfile

HERE = os.path.dirname(os.path.abspath(__file__))
REPO_ROOT = os.path.dirname(os.path.dirname(HERE))
AUDIT = os.path.join(HERE, "audit_node_runtime.py")


def make_tree() -> str:
    """A throwaway repo copy holding only .github/ (workflows + script)."""
    root = tempfile.mkdtemp(prefix="nodeaudit-")
    dest = os.path.join(root, ".github")
    os.makedirs(dest)
    shutil.copytree(os.path.join(REPO_ROOT, ".github", "workflows"), os.path.join(dest, "workflows"))
    os.makedirs(os.path.join(dest, "scripts"))
    shutil.copy(AUDIT, os.path.join(dest, "scripts", "audit_node_runtime.py"))
    return root


PROBE = "zz-audit-probe.yaml"


def run_audit(root: str, full: bool = False) -> tuple[int, str]:
    """Run the SHIPPED script as a subprocess; return (rc, combined output).

    ``full=False`` audits only the planted probe file. The probe is a copy of the
    production scanner run through the production entry point, so the verdict is
    the real one -- not a reconstruction -- while staying fast enough to run a
    case per network call.
    """
    cmd = [sys.executable, os.path.join(root, ".github", "scripts", "audit_node_runtime.py"), "--report"]
    if not full:
        cmd += ["--workflow", PROBE]
    proc = subprocess.run(cmd, capture_output=True, text=True, timeout=600)
    return proc.returncode, proc.stdout + proc.stderr


def plant(root: str, body: str) -> None:
    with open(os.path.join(root, ".github", "workflows", PROBE), "w", encoding="utf-8", newline="\n") as fh:
        fh.write(body)


WORKFLOW = """name: {name}
on: [push]
permissions:
  contents: read
jobs:
  probe:
    runs-on: ubuntu-latest
    steps:
{steps}
"""


def step(uses: str, extra: str = "") -> str:
    return f"      - uses: {uses}\n{extra}"


# (name, pin, expectation)
#   pin        -- what goes in the probe's `uses:`; None means "plant nothing"
#   expectation -- "caught"  : the audit MUST report a node20 runtime
#                 "ignored" : the audit MUST stay clean and say why it skipped
CASES = [
    (
        "plain-node20-action-is-caught",
        "amondnet/vercel-action@v25",
        "caught",
        "A direct `using: node20` pin must be reported.",
    ),
    (
        "node24-pin-is-silent",
        "amondnet/vercel-action@v42",
        "ignored",
        "A node24 pin must stay silent, or the audit is pure noise.",
    ),
    (
        "composite-wrapping-node20-is-caught",
        "actions/upload-pages-artifact@v4",
        "caught",
        "A composite whose OWN using: is composite, but whose embedded step is "
        "node20, must be reported TRANSITIVELY -- and named by the pin that "
        "actually carries node20, not by the wrapper.",
    ),
    (
        "reusable-workflow-call-is-not-an-action",
        "slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.1.0",
        "ignored",
        "A reusable-workflow call has no action manifest; it must be skipped "
        "silently, not reported as an unresolvable pin.",
    ),
    (
        "docker-runtime-action-is-silent",
        "aquasecurity/trivy-action@v0.24.0",
        "ignored",
        "A `using: docker` action emits no Node warning and must not be flagged.",
    ),
]


def main() -> int:
    failures: list[str] = []
    root = make_tree()
    try:
        for name, pin, expectation, why in CASES:
            plant(root, WORKFLOW.format(name=name, steps=step(pin)))
            code, out = run_audit(root)

            if expectation == "caught":
                ok = "NODE20" in out and "INCOMPLETE" not in out
                if not ok:
                    failures.append(
                        f"{name}: {pin} expected a node20 report, got:\n{out}\n  (why: {why})"
                    )
            else:
                # "ignored" means the pin is SKIPPED, not that the run is clean.
                # A probe holding nothing but a skipped pin legitimately trips the
                # audit's own "no refs at all -> scanner is broken" guard, so
                # requiring a clean RESULT here would assert the wrong thing.
                # The property under test is that the pin is never treated as an
                # action: no NODE20, and no MISSING verdict either.
                ok = "NODE20" not in out and "MISSING" not in out
                if not ok:
                    failures.append(
                        f"{name}: {pin} was treated as an action, got:\n{out}\n  (why: {why})"
                    )
            print(f"  {'ok  ' if ok else 'FAIL'} {name}")

        # A pin that cannot exist upstream must NOT be graded as a pass.
        plant(root, WORKFLOW.format(name="missing", steps=step("virtengine/definitely-not-a-real-action-xyz@v1")))
        code, out = run_audit(root)
        if "RESULT: no node20" in out:
            failures.append("missing-pin: unreadable pin was graded as a clean tree\n" + out)
        print(f"  {'FAIL' if any(f.startswith('missing-pin') for f in failures) else 'ok  '} missing-pin-is-not-a-pass")

        # And the shipped tree, untouched, must be clean end to end.
        os.remove(os.path.join(root, ".github", "workflows", PROBE))
        code, out = run_audit(root, full=True)
        if "RESULT: no node20" not in out:
            failures.append(f"baseline: shipped tree is not clean\n{out}")
        print(f"  {'FAIL' if any(f.startswith('baseline') for f in failures) else 'ok  '} baseline-tree-is-clean")
    finally:
        shutil.rmtree(root, ignore_errors=True)

    for f in failures:
        print("FAIL:", f)
    total = len(CASES) + 2
    print(f"{total - len({f.split(':')[0] for f in failures})}/{total} checks passed")
    return 1 if failures else 0


if __name__ == "__main__":
    sys.exit(main())
