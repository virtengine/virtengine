#!/usr/bin/env python3
"""Audit every ``uses:`` action in .github/workflows for a node20 JS runtime.

Why this exists
---------------
GitHub emits ``##[warning]Node.js 20 is deprecated`` on any step whose action
declares ``using: node20``.  As of 2026-09-26 every job in this repo carries
that warning, which makes it useless as a signal: the tree was swept
mechanically off the deprecated runtimes, but nothing proved the sweep was
complete, so a straggler added later is invisible.

The audit is deliberately *behavioural* rather than a list of blessed pins:

* it reads each action's real ``action.yml`` at the real ref,
* it resolves COMPOSITE actions recursively, because ``upload-pages-artifact``
  is ``using: composite`` and its embedded ``actions/upload-artifact@v4``
  is the thing that actually runs node20,
* and it compares the action's declared INPUTS against the inputs the
  workflow passes, so a bump is only ever reported as safe when it is.

Usage
-----
    python .github/scripts/audit_node_runtime.py            # audit, exit 1 on node20
    python .github/scripts/audit_node_runtime.py --report   # audit, always exit 0

Requires ``gh`` on PATH and network access. Without either, the audit exits 2
("could not audit") rather than reporting a clean tree -- an unrunnable audit
must never read as a pass.
"""

from __future__ import annotations

import argparse
import base64
import json
import os
import re
import subprocess
import sys
from dataclasses import dataclass, field

REPO_ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
WORKFLOWS = os.path.join(REPO_ROOT, ".github", "workflows")

# `using:` values that are NOT a deprecated JS runtime.
SAFE_USINGS = {"composite", "docker"}

# A `uses: repo[/path]@ref` line. Actions are only meaningful at step level,
# so the `uses:` is at an indentation a job step can have.
USES_RE = re.compile(r"^\s*-?\s*uses:\s*['\"]?(?P<ref>[^'\"\s#]+)")
USES_ANY_RE = re.compile(r"^\s*uses:\s*['\"]?(?P<ref>[^'\"\s#]+)")
# A step may put the action on the dash line: `- uses: owner/repo@v1`. That is
# the form every action's own README uses, and a scanner that only reads the
# two-line form misses a workflow written that way entirely.
INLINE_USES_RE = re.compile(r"^\s*-\s+uses:\s*['\"]?(?P<ref>[^'\"\s#]+)")
# `with:` block entries, used to measure an action's consumed input set.
WITH_KEY_RE = re.compile(r"^\s{4,}(?P<key>[A-Za-z0-9_-]+):")

# Refs that are not upstream actions: local reusable workflows, docker refs
# pinned by digest, and dynamic expressions.
LOCAL_PREFIX = "./"
DYNAMIC = "${{"


@dataclass
class ActionRef:
    """One ``uses:`` occurrence, with where it came from and what it passes."""

    repo: str
    ref: str
    file: str
    line: int
    inputs: list[str] = field(default_factory=list)

    @property
    def spec(self) -> str:
        return f"{self.repo}@{self.ref}"


def iter_action_refs(workflow_file: str) -> list[ActionRef]:
    """Every local (non-composite-step) ``uses:`` in one workflow file.

    A ``uses:`` nested under a step's ``with:`` block is a *reusable workflow
    call*, which is a different mechanism -- recorded separately by the caller
    -- so only top-level step uses are returned here.
    """
    out: list[ActionRef] = []
    with open(workflow_file, encoding="utf-8") as fh:
        lines = fh.read().splitlines()

    in_with = False
    with_indent = 0
    for idx, raw in enumerate(lines):
        stripped = raw.strip()
        if not stripped:
            continue

        indent = len(raw) - len(raw.lstrip())
        if stripped.startswith("with:"):
            in_with = True
            with_indent = indent
            continue
        # Leaving the `with:` block: any dedent past its key ends it.
        if in_with and indent <= with_indent and not stripped.startswith("-"):
            in_with = False

        if in_with:
            m = WITH_KEY_RE.match(raw)
            # Only the step's own inputs count, not a nested map's keys.
            if m and indent > with_indent:
                out_keys = out[-1].inputs if out else []
                if m.group("key") not in out_keys:
                    out_keys.append(m.group("key"))
            continue

        m = USES_ANY_RE.match(raw) or INLINE_USES_RE.match(raw)
        if not m:
            continue
        ref = m.group("ref")
        if ref.startswith(LOCAL_PREFIX) or DYNAMIC in ref:
            continue
        # `uses: owner/repo/.github/workflows/x.yml@ref` is a REUSABLE WORKFLOW
        # call, not an action: it has no action manifest and its steps are the
        # called workflow's own. Auditing it as an action yields a permanent
        # false "unresolvable".
        if ref.rpartition("@")[0].endswith((".yml", ".yaml")):
            continue
        repo, _, version = ref.rpartition("@")
        if not repo or "/" not in repo:
            continue
        out.append(ActionRef(repo=repo, ref=version, file=os.path.basename(workflow_file), line=idx + 1))
    return out


def scan_workflows() -> list[ActionRef]:
    """Every action pin across the whole workflow directory."""
    refs: list[ActionRef] = []
    for name in sorted(os.listdir(WORKFLOWS)):
        if name.endswith((".yaml", ".yml")):
            refs.extend(iter_action_refs(os.path.join(WORKFLOWS, name)))
    return refs


_FETCH_CACHE: dict[tuple[str, str, str], str | None] = {}

# An action's manifest is `action.yml` by convention, but the name is not
# enforced: aquasecurity/trivy-action, ludeeus/action-shellcheck and
# benc-uk/workflow-dispatch all ship `action.yaml` only. Probing only one name
# silently marks a third of the tree "unknown", which reads as a pass.
MANIFEST_NAMES = ("action.yml", "action.yaml")

# Sentinel: the pin MAY exist but could not be read this run. Distinct from
# None ("no such manifest"), so a transient API fault is never graded as clean.
TRANSIENT_FAILURE = "\x00TRANSIENT\x00"


def split_ref(spec_repo: str) -> tuple[str, str]:
    """('github/codeql-action/init') -> ('github/codeql-action', 'init')."""
    parts = spec_repo.split("/")
    return "/".join(parts[:2]), "/".join(parts[2:])


def fetch_action_yml(repo_path: str, ref: str) -> str | None:
    """The raw action manifest at ``repo_path@ref``, or None if unreadable.

    Tries each known manifest name in a subdirectory when the pin names one.
    """
    key = (repo_path, ref, "manifest")
    if key in _FETCH_CACHE:
        return _FETCH_CACHE[key]

    owner_repo, subdir = split_ref(repo_path)
    prefixes = [subdir] if subdir else ["", "."]
    saw_other_error = False
    for prefix in prefixes:
        for name in MANIFEST_NAMES:
            path = f"{prefix}/{name}" if prefix else name
            try:
                proc = subprocess.run(
                    ["gh", "api", f"repos/{owner_repo}/contents/{path}?ref={ref}", "--jq", ".content"],
                    capture_output=True,
                    text=True,
                    timeout=60,
                    check=True,
                )
                content = base64.b64decode(proc.stdout.strip()).decode("utf-8", "replace")
            except subprocess.CalledProcessError as exc:
                # 404 means "this layout does not exist" -- keep probing. Any
                # other failure (5xx, auth, rate limit) is a TRANSIENT fault and
                # must not be silently demoted to "no such file", or a network
                # blip reads as a clean audit and the gate goes green blind.
                if "404" not in (exc.stderr or ""):
                    saw_other_error = True
                continue
            except (subprocess.SubprocessError, ValueError, OSError):
                saw_other_error = True
                continue
            _FETCH_CACHE[key] = content
            return content

    _FETCH_CACHE[key] = None if not saw_other_error else TRANSIENT_FAILURE


def parse_action_yml(text: str) -> tuple[str, list[str]]:
    """(using, input names) from an action.yml.

    Inputs are the top-level keys of the ``inputs:`` block only. ``runs.steps``
    keys (``run``, ``shell``, ``with``) must not be counted, or every action
    looks like it accepts every input.
    """
    using = ""
    m = re.search(r"^\s*using:\s*['\"]?([\w-]+)", text, re.M)
    if m:
        using = m.group(1)

    inputs: list[str] = []
    in_inputs = False
    base_indent = 0
    for raw in text.splitlines():
        if re.match(r"^inputs:\s*$", raw):
            in_inputs = True
            base_indent = 0
            continue
        if not in_inputs:
            continue
        stripped = raw.strip()
        if not stripped:
            continue
        indent = len(raw) - len(raw.lstrip())
        if indent == 0:
            break
        if base_indent == 0:
            base_indent = indent
        if indent == base_indent:
            km = re.match(r"^([A-Za-z0-9_-]+):", stripped)
            if km:
                inputs.append(km.group(1))
    return using, inputs


def embedded_uses(text: str) -> list[str]:
    """``uses:`` lines inside a composite action's ``runs.steps``."""
    found = []
    in_steps = False
    for raw in text.splitlines():
        stripped = raw.strip()
        if re.match(r"^\s*steps:\s*$", raw):
            in_steps = True
            continue
        if in_steps and re.match(r"^[A-Za-z_]+:", raw):
            in_steps = False
        if not in_steps:
            continue
        m = USES_RE.match(raw) or USES_ANY_RE.match(raw)
        if m and not m.group("ref").startswith(LOCAL_PREFIX):
            found.append(m.group("ref"))
    return found


def node20_findings(refs: list[ActionRef]) -> tuple[list[str], list[str], list[str]]:
    """(node20 findings, missing-manifest problems, transient API faults).

    A finding is a node20 runtime, reachable transitively. Composite actions are
    followed: a composite whose own ``using:`` is fine can still execute node20
    in an embedded step, and that is precisely how
    ``upload-pages-artifact@v4`` was missed.

    The three outcomes are kept apart on purpose. Collapsing "upstream has no
    such manifest" and "GitHub was unreachable" into one bucket is what lets a
    blind run report a clean tree.
    """
    findings: list[str] = []
    problems: list[str] = []
    transient: list[str] = []
    seen: set[tuple[str, str]] = set()

    for ref in refs:
        stack: list[tuple[str, str]] = [(ref.repo, ref.ref)]
        local_seen: set[tuple[str, str]] = set()
        while stack:
            repo, version = stack.pop()
            if (repo, version) in local_seen:
                continue
            local_seen.add((repo, version))
            text = fetch_action_yml(repo, version)
            if text is None:
                problems.append(f"{ref.spec} ({ref.file}:{ref.line}): no action manifest found upstream")
                break
            if text == TRANSIENT_FAILURE:
                transient.append(f"{ref.spec} ({ref.file}:{ref.line}): GitHub API error, not a verdict")
                break
            using, _ = parse_action_yml(text)
            if using.startswith("node"):
                node_ver = using.removeprefix("node")
                marker = f"{repo}@{version} -> using: node{node_ver}"
                if node_ver == "20" and (repo, version) not in seen:
                    seen.add((repo, version))
                    findings.append(f"{marker}  [{ref.file}:{ref.line}]")
            if using == "composite":
                for sub in embedded_uses(text):
                    srepo, _, sver = sub.rpartition("@")
                    if srepo and sver:
                        stack.append((srepo, sver))
    return findings, problems, transient


def consumed_inputs(refs: list[ActionRef]) -> dict[str, set[str]]:
    """action spec -> inputs the workflows actually pass."""
    consumed: dict[str, set[str]] = {}
    for ref in refs:
        consumed.setdefault(f"{ref.repo}", set()).update(ref.inputs)
    return consumed


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--report", action="store_true", help="always exit 0")
    ap.add_argument(
        "--workflow",
        action="append",
        default=None,
        metavar="NAME",
        help="audit only these workflow files (repeatable); used by the falsifier",
    )
    args = ap.parse_args()

    if not os.path.isdir(WORKFLOWS):
        print(f"FAIL: no workflow directory at {WORKFLOWS}", file=sys.stderr)
        return 2

    files = sorted(f for f in os.listdir(WORKFLOWS) if f.endswith((".yaml", ".yml")))
    if args.workflow:
        unknown = [w for w in args.workflow if w not in files]
        if unknown:
            print(f"FAIL: no such workflow file(s): {', '.join(unknown)}", file=sys.stderr)
            return 2
        files = [f for f in files if f in set(args.workflow)]

    refs: list[ActionRef] = []
    for name in files:
        refs.extend(iter_action_refs(os.path.join(WORKFLOWS, name)))
    if not refs:
        print("FAIL: no `uses:` references found -- the scanner is broken, not the tree", file=sys.stderr)
        return 2

    findings, problems, transient = node20_findings(refs)
    specs = sorted({r.spec for r in refs})
    print(f"audited {len(refs)} `uses:` references ({len(specs)} distinct action pins)")
    print(f"node20-runtime actions reachable: {len(findings)}")
    for f in findings:
        print(f"  NODE20  {f}")

    if transient:
        print(f"pins unreadable this run (API fault, NOT a verdict): {len(transient)}")
        for t in sorted(set(transient)):
            print(f"  TRANSIENT  {t}")
        print("audit INCOMPLETE -- a blind run is not a pass")
        return 2

    if problems:
        print(f"pins with no upstream action manifest: {len(problems)}")
        for p in sorted(set(problems)):
            print(f"  MISSING  {p}")
        print("audit INCOMPLETE -- an unresolvable pin is not a pass")
        return 2

    if findings:
        print("RESULT: node20 runtimes present")
        return 0 if args.report else 1
    print("RESULT: no node20-runtime action reachable from .github/workflows")
    return 0


if __name__ == "__main__":
    sys.exit(main())
