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


def run_audit(root: str, full: bool = False, env_extra: dict | None = None, report: bool = True) -> tuple[int, str]:
    """Run the SHIPPED script as a subprocess; return (rc, combined output).

    ``full=False`` audits only the planted probe file. The probe is a copy of the
    production scanner run through the production entry point, so the verdict is
    the real one -- not a reconstruction -- while staying fast enough to run a
    case per network call.

    ``report=False`` drops ``--report`` so the EXIT CODE is the real gate
    behaviour. Any assertion about a check that must fail a job has to go
    through that mode: under ``--report`` every run exits 0 by construction, so a
    text match alone would pass even if the verdict branch were deleted.
    """
    cmd = [sys.executable, os.path.join(root, ".github", "scripts", "audit_node_runtime.py")]
    if report:
        cmd.append("--report")
    if not full:
        cmd += ["--workflow", PROBE]
    env = dict(os.environ)
    if env_extra is not None:
        env.update(env_extra)
    proc = subprocess.run(cmd, capture_output=True, text=True, timeout=600, env=env)
    return proc.returncode, proc.stdout + proc.stderr


def unauthenticated() -> dict:
    """Env that makes `gh` genuinely unusable: no token AND no config file.

    Clearing only the env vars is NOT enough here -- this host's `gh`
    authenticates from its config file, so a tokenless-but-configured run is
    still capable, and the audit is right to proceed. The condition under test
    is "gh cannot authenticate at all", so the config has to go too.
    """
    empty = tempfile.mkdtemp(prefix="gh-empty-")
    return {"GH_TOKEN": "", "GITHUB_TOKEN": "", "GH_CONFIG_DIR": empty}


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

        # The INPUT half of the audit. `consumed_inputs()` shipped as dead code
        # for months while the module docstring promised this exact check, so
        # these cases are the only thing standing between the next refactor and
        # the same silent hole. The bug they guard is not hypothetical: #1272 was
        # a `download-artifact` step passing an input v6 does not declare, caught
        # by CI rather than by this audit.
        #
        # `amondnet/vercel-action@v42` is the node24 pin from the cases above:
        # declaring no runtime finding is precisely the point -- a node24 pin is
        # what a safe bump looks like, so it is the only place a false positive
        # would show up.

        # 1. An input the action does NOT declare must be reported. `fetch-depth`
        # is declared on actions/checkout, not on vercel-action.
        plant(root, WORKFLOW.format(
            name="undeclared-input",
            steps=step("amondnet/vercel-action@v42", extra="        with:\n          fetch-depth: 0\n"),
        ))
        code, out = run_audit(root)
        ok = "UNDECLARED_INPUT" in out and "fetch-depth" in out
        if not ok:
            failures.append(
                "undeclared-input-is-caught: a `with:` key the action does not declare "
                f"must be reported; got rc={code}:\n{out}\n  (why: this is the check the "
                "docstring promises and the one #1272 needed)"
            )
        print(f"  {'ok  ' if ok else 'FAIL'} undeclared-input-is-caught")

        # 2. ...and it must fail the gate, not merely be printed. Under --report
        # every run exits 0, so the exit code has to be checked for real.
        code, out = run_audit(root, report=False)
        ok = code == 1 and "undeclared input" in out
        if not ok:
            failures.append(
                "undeclared-input-fails-the-gate: an undeclared input must exit 1; "
                f"got rc={code}:\n{out}"
            )
        print(f"  {'ok  ' if ok else 'FAIL'} undeclared-input-fails-the-gate")

        # 3. A DECLARED input must stay silent, or the check is pure noise and the
        # gate will be switched off at the first false positive. `target` is
        # declared on vercel-action@v42 -- verified against its real manifest,
        # because asserting on an input name invented from memory is how a check
        # like this rots into a check that asserts nothing.
        plant(root, WORKFLOW.format(
            name="declared-input",
            steps=step("amondnet/vercel-action@v42", extra="        with:\n          target: production\n"),
        ))
        code, out = run_audit(root)
        ok = "UNDECLARED_INPUT" not in out
        if not ok:
            failures.append(f"declared-input-is-silent: got:\n{out}")
        print(f"  {'ok  ' if ok else 'FAIL'} declared-input-is-silent")

        # 4. A COMPOSITE action must be reported as UNCHECKED rather than
        # silently passed. `runs.steps` is not an input manifest, so the audit
        # genuinely cannot validate it -- but "cannot check" must never render as
        # "checked and clean", which is the exact failure this card is about.
        plant(root, WORKFLOW.format(
            name="composite-inputs-unchecked",
            steps=step("actions/upload-pages-artifact@v5", extra="        with:\n          totally-not-an-input: x\n"),
        ))
        code, out = run_audit(root)
        ok = "INPUTS_UNCHECKED" in out and "UNDECLARED_INPUT" not in out
        if not ok:
            failures.append(
                "composite-inputs-are-unchecked-not-clean: a composite pin's inputs cannot "
                f"be validated from the manifest and must be labelled as such; got:\n{out}"
            )
        print(f"  {'ok  ' if ok else 'FAIL'} composite-inputs-are-unchecked-not-clean")

        # 5. A `with:` block must attach to the step it belongs to, and the NEXT step
        # must still be recorded as a pin. Both halves matter: two inline
        # `- uses:` steps, the first passing an input the second's action does
        # not declare. Charging the key to the wrong pin would invent a failure,
        # and swallowing the second step would audit nothing at all -- which is
        # exactly what the scanner did until the `with:`/step boundary was fixed.
        two_steps = (
            step("amondnet/vercel-action@v42", extra="        with:\n          target: production\n")
            + step("amondnet/vercel-action@v42", extra="        with:\n          fetch-depth: 0\n")
        )
        plant(root, WORKFLOW.format(name="attribution", steps=two_steps))
        code, out = run_audit(root)
        ok = "audited 2 `uses:` references" in out
        if not ok:
            failures.append(
                "step-after-a-with-block-is-still-a-pin: the second step was swallowed by "
                f"the first step's `with:` block, so an action went unaudited; got:\n{out}"
            )
        print(f"  {'ok  ' if ok else 'FAIL'} step-after-a-with-block-is-still-a-pin")

        # 6. And the bad key must be reported ONCE, against the step that actually owns
        # it (line 12, the second step) -- not zero (key dropped), not twice, and
        # not against line 9 (the first step, which passes only `target`, a
        # declared input). Charging the key to the wrong step would invent a
        # failure on a correct workflow.
        findings = [ln for ln in out.splitlines() if "UNDECLARED_INPUT" in ln]
        ok = len(findings) == 1 and ":12)" in findings[0] and "fetch-depth" in findings[0]
        if not ok:
            failures.append(
                "inputs-attach-to-their-own-step: expected exactly one finding, on the step "
                f"that owns the key (line 12); got {len(findings)} finding(s):\n{out}"
            )
        print(f"  {'ok  ' if ok else 'FAIL'} inputs-attach-to-their-own-step")

        # 7. A `with:` block containing a MULTI-LINE SCALAR must keep charging
        # keys after the scalar. This is the shape the shipped scanner got wrong:
        # it ended the block on any `- ` line, and the body of `body: |` is full
        # of them, so `branch`, `base`, `labels` and `delete-branch` were dropped
        # and the audit reported a clean tree for a step that would fail at run
        # time. The flat one-line fixtures above cannot see this -- every value
        # they use ends on its own line.
        #
        # Asserted in BOTH directions: the key after the scalar is caught when it
        # is undeclared (it is not silently dropped) AND the `- ` line itself is
        # not mistaken for an input (it is scalar text, not a `with:` entry).
        scalar_with = (
            "        with:\n"
            "          target: production\n"
            "          body: |\n"
            "            - Updated CHANGELOG.md with all changes\n"
            "            - Generated release notes\n"
            "          totally-not-an-input: changelog/x\n"
        )
        plant(root, WORKFLOW.format(
            name="scalar-body",
            steps=step("amondnet/vercel-action@v42", extra=scalar_with),
        ))
        code, out = run_audit(root)
        ok = "UNDECLARED_INPUT" in out and "totally-not-an-input" in out
        if not ok:
            failures.append(
                "key-after-a-multi-line-scalar-is-still-charged: a `- ` line inside a "
                "`body: |` scalar ended the `with:` block, so every key after it went "
                f"unaudited; got rc={code}:\n{out}"
            )
        # The scalar's own lines must not be charged as inputs: they are text
        # the action receives as the VALUE of `body`, not `with:` keys.
        stray = [ln for ln in out.splitlines()
                 if "UNDECLARED_INPUT" in ln and "Updated CHANGELOG.md" in ln]
        if stray:
            failures.append(
                "multi-line-scalar-body-is-not-read-as-inputs: `- ` lines inside a "
                f"scalar body were charged as `with:` keys:\n{out}"
            )
        print(f"  {'ok  ' if ok and not stray else 'FAIL'} key-after-a-multi-line-scalar-is-still-charged")

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

        # The input check must not have made the clean tree quietly incomplete:
        # a run that could not read a manifest is not a pass, so this asserts the
        # INPUT sweep reports incompleteness on its own account.
        code, out = run_audit(root, full=True)
        ok = "INCOMPLETE" not in out
        if not ok:
            failures.append(f"baseline-input-sweep-is-complete: shipped tree reported INCOMPLETE\n{out}")
        print(f"  {'FAIL' if not ok else 'ok  '} baseline-input-sweep-is-complete")

        # An unauthenticated `gh` is the condition that shipped this job red on
        # a clean tree. It must be refused UP FRONT with one actionable line,
        # not surface later as 43 identical per-pin "API error" lines.
        code, out = run_audit(root, full=True, env_extra=unauthenticated())
        ok = code == 2 and "gh auth status" in out and "TRANSIENT" not in out
        if not ok:
            failures.append(
                f"unauthenticated-gh-is-refused-upfront: expected exit 2 naming "
                f"`gh auth status` with no per-pin noise, got rc={code}:\n{out}"
            )
        print(f"  {'FAIL' if not ok else 'ok  '} unauthenticated-gh-is-refused-upfront")
    finally:
        shutil.rmtree(root, ignore_errors=True)

    for f in failures:
        print("FAIL:", f)
    total = len(CASES) + 10
    print(f"{total - len({f.split(':')[0] for f in failures})}/{total} checks passed")
    return 1 if failures else 0


if __name__ == "__main__":
    sys.exit(main())
