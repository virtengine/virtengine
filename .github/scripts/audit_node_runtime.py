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
* and it compares the INPUTS a JS action's manifest declares against the
  inputs the workflow actually passes, so an undeclared ``with:`` key is
  caught by this audit rather than by the job that trips over it.

What this audit does NOT check: composite and docker actions' input names.
``runs.steps`` is not an input manifest and a docker action takes its arguments
through ``args:``, not ``with:``, so neither can be validated from the
manifest alone. Only JS actions are enforced -- those declare a flat
``inputs:`` map that GitHub rejects an unknown key against, with
``input "x" is not defined in action owner/repo@ref``. A composite or docker
pin is reported as ``INPUTS_UNCHECKED`` rather than passed in silence: "cannot
be checked" must never render as "checked and clean".

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
import time
import urllib.error
import urllib.request
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
    # Indent of the FIRST key inside the current `with:` block. Only keys at
    # exactly this indent are the step's own inputs: a deeper line belongs to a
    # nested map or to the body of a multi-line scalar, and counting it invents
    # an input name the workflow never passes.
    with_key_indent = 0
    # The step `uses:` the current `with:` belongs to. Deliberately NOT simply
    # `out[-1]`: a job-level `uses:` (a reusable-workflow call) is filtered out
    # below, so without this reset its `with:` keys would be charged to whatever
    # step happened to come last in the PREVIOUS job -- an invented input on an
    # unrelated pin. Reset at every job boundary, so an unrecorded `with:` is
    # attributed to nobody, which is the truth.
    last_step: ActionRef | None = None
    in_jobs = False

    for idx, raw in enumerate(lines):
        stripped = raw.strip()
        if not stripped:
            continue

        indent = len(raw) - len(raw.lstrip())
        if indent == 0:
            in_jobs = stripped.startswith("jobs:")
            last_step = None
        elif in_jobs and indent == 2 and re.match(r"^[A-Za-z0-9_-]+:\s*$", stripped):
            last_step = None  # a new job starts; the previous job's `uses:` is gone

        if stripped.startswith("with:"):
            in_with = True
            with_indent = indent
            with_key_indent = 0
            continue
        # Leaving the `with:` block. A `with:` takes a mapping, never a list
        # item, so a line starting with `- ` ALWAYS ends the block -- and that
        # includes the inline one-line step form `      - uses: owner/repo@v1`.
        # Without this the block stayed open, the next step was swallowed by the
        # input branch below, and the step was never recorded as a pin AT ALL:
        # a whole action, inputs included, audited by nothing. Only the two-line
        # step form (`- name:` then `uses:`) escaped, which is why the tree
        # looked clean -- the bug was live, just not yet triggered.
        if in_with and (indent <= with_indent or stripped.startswith("-")):
            in_with = False

        if in_with:
            m = WITH_KEY_RE.match(raw)
            if not m:
                continue
            if with_key_indent == 0:
                with_key_indent = indent
            if indent == with_key_indent and last_step is not None:
                key = m.group("key")
                if key not in last_step.inputs:
                    last_step.inputs.append(key)
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
        action = ActionRef(repo=repo, ref=version, file=os.path.basename(workflow_file), line=idx + 1)
        out.append(action)
        # Recorded even for a ref this scanner skips for runtime auditing (a
        # composite or docker pin), because the INPUT question is separate from
        # the runtime one and applies to those too.
        last_step = action
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

# Why a pin could not be read, for the report. Without this the operator sees
# "API error" 43 times and learns nothing; the class is the diagnosis.
_FETCH_ERRORS: dict[tuple[str, str, str], str] = {}

# Paced delay between upstream reads. The whole tree is ~40 manifests and is
# read in seconds, which is what tripped GitHub's abuse detection on the first
# CI run; a 43-read burst against one API is exactly the shape it flags.
_READ_DELAY_SECONDS = 0.4

# Whether a read for this pin has already happened, so the delay is applied
# BETWEEN upstream calls rather than before the first.
_READ_SEEN: dict[tuple[str, str, str], bool] = {}


def split_ref(spec_repo: str) -> tuple[str, str]:
    """('github/codeql-action/init') -> ('github/codeql-action', 'init')."""
    parts = spec_repo.split("/")
    return "/".join(parts[:2]), "/".join(parts[2:])


def _run_gh(owner_repo: str, path: str, ref: str):
    """One authenticated `gh api` contents read."""
    return subprocess.run(
        ["gh", "api", f"repos/{owner_repo}/contents/{path}?ref={ref}", "--jq", ".content"],
        capture_output=True,
        text=True,
        timeout=60,
        check=True,
    )


def _classify(exc: subprocess.CalledProcessError) -> str:
    """Error class from `gh`'s stderr. Split out so the authenticated read and
    the anonymous read classify their failures identically."""
    blob = ((exc.stderr or "") + (exc.stdout or "")).lower()
    if "404" in blob or "not found" in blob:
        return "404"
    if "rate limit" in blob or "429" in blob or "abuse" in blob:
        return "ratelimit"
    if "403" in blob or "forbidden" in blob or "not accessible" in blob:
        return "forbidden"
    return "api-error"


def _classify_http(status: int) -> str:
    """The same classes, from an HTTP status code instead of a stderr string."""
    if status == 404:
        return "404"
    if status == 429:
        return "ratelimit"
    if status == 403:
        return "forbidden"
    return "api-error"


def _anon_contents(owner_repo: str, path: str, ref: str) -> tuple[str | None, str]:
    """(content, error_class) for an UNAUTHENTICATED read of a public file.

    This deliberately does not shell out to `gh`. `gh` refuses to issue an API
    call that carries no credential -- on a runner it exits non-zero with its own
    usage text before any request is made -- so "run gh with the token removed"
    is not an anonymous read at all, it is a guaranteed failure wearing the
    costume of one. The first attempt at this fix did exactly that and moved the
    5 pins from `forbidden` to `api-error` in CI while passing every local
    test, because a workstation's `gh` has a logged-in account to fall back on
    and a runner does not.

    A plain HTTPS GET of a public path on api.github.com needs no credential,
    which is the whole point: the read must not depend on who is asking.
    """
    url = f"https://api.github.com/repos/{owner_repo}/contents/{path}?ref={ref}"
    req = urllib.request.Request(
        url,
        headers={
            "Accept": "application/vnd.github+json",
            "User-Agent": "virtengine-node-runtime-audit",
            "X-GitHub-Api-Version": "2022-11-28",
        },
    )
    try:
        with urllib.request.urlopen(req, timeout=60) as resp:
            payload = json.loads(resp.read().decode("utf-8", "replace"))
    except urllib.error.HTTPError as exc:
        return None, _classify_http(exc.code)
    except (urllib.error.URLError, OSError, ValueError):
        return None, "unavailable"
    try:
        return base64.b64decode(payload["content"]).decode("utf-8", "replace"), ""
    except (KeyError, TypeError, ValueError):
        return None, "api-error"


def _gh_contents(owner_repo: str, path: str, ref: str) -> tuple[str | None, str]:
    """(content, error_class). error_class is "" on success.

    Coarse on purpose: we need to tell "this layout does not exist" (keep
    probing) from "the call did not happen" (transient, retry) -- not to
    diagnose GitHub.

    A `forbidden` verdict is retried ANONYMOUSLY before it is believed. Every
    manifest this audit reads is public, so a permission failure says something
    about the CALLER's token, never about whether the file exists. The CI job
    passes the job's own ``github.token``, whose scope does not cover other
    repositories, and GitHub answers 403 "Resource not accessible by
    integration" for pins that are plainly readable -- which made the audit
    declare itself INCOMPLETE on a tree it had already proven clean on every
    other pin. An unauthenticated read of a public file needs no scope at all,
    so a permission failure that survives it is a real failure and the class is
    returned unchanged.
    """
    try:
        proc = _run_gh(owner_repo, path, ref)
        return base64.b64decode(proc.stdout.strip()).decode("utf-8", "replace"), ""
    except subprocess.CalledProcessError as exc:
        err = _classify(exc)
        if err != "forbidden":
            return None, err
    except (subprocess.SubprocessError, ValueError, OSError):
        return None, "unavailable"

    # Authenticated read was refused. Retry with no credential at all.
    return _anon_contents(owner_repo, path, ref)


# Severity order, most transient first. `>` on this ordering, NOT on
# _ATTEMPTS: `forbidden` and `404` share an attempt budget of 1, so comparing
# budgets let "404" -- the default -- win ties and a permission failure was
# reported as "this repo has no manifest". Two classes with equal retry value
# must still be distinguishable, and only a genuine 404 may be called MISSING.
_SEVERITY = {"404": 0, "forbidden": 1, "api-error": 2, "unavailable": 3, "ratelimit": 4}

# How many attempts each failure class earns. 404 is terminal (this layout does
# not exist); everything else is worth another go, because the whole tree is
# read in a few seconds and GitHub's abuse detection is what actually bites.
_ATTEMPTS = {"404": 1, "forbidden": 2, "api-error": 2, "unavailable": 2, "ratelimit": 4}


def fetch_action_yml(repo_path: str, ref: str) -> str | None:
    """The raw action manifest at ``repo_path@ref``, or None if unreadable.

    Tries each known manifest name in a subdirectory when the pin names one,
    and retries a genuinely transient failure with backoff. A manifest read is
    cheap; a rate-limited upstream must not decide whether the gate is red.
    """
    key = (repo_path, ref, "manifest")
    if key in _FETCH_CACHE:
        return _FETCH_CACHE[key]

    owner_repo, subdir = split_ref(repo_path)
    # One prefix per pin. A previous version also probed "." for root-level
    # pins, which doubled every 404-bearing root pin's API calls for no chance
    # of a different answer -- `gh` resolves "action.yml" and "./action.yml" to
    # the same object.
    prefixes = [subdir] if subdir else [""]

    best = "404"
    for attempt in range(1, max(_ATTEMPTS.values()) + 1):
        # The MOST SEVERE class seen in a full sweep decides the outcome, and
        # the ranking must be independent of `attempt`: an earlier version
        # compared a class's attempt budget against the sweep number, so a
        # `forbidden` (budget 1) seen on attempt 1 never displaced the "404"
        # default -- a PERMISSION failure was reported as "no such manifest".
        # That is the worst possible direction: a confident falsehood about a
        # repository that plainly exists.
        worst = "404"
        for prefix in prefixes:
            for name in MANIFEST_NAMES:
                path = f"{prefix}/{name}" if prefix else name
                if _READ_SEEN.get(key):
                    time.sleep(_READ_DELAY_SECONDS)
                _READ_SEEN[key] = True
                content, err = _gh_contents(owner_repo, path, ref)
                if err == "":
                    _FETCH_CACHE[key] = content
                    return content
                if _SEVERITY[err] > _SEVERITY[worst]:
                    worst = err
        best = worst
        if _ATTEMPTS[worst] <= attempt:
            break
        time.sleep(3 * attempt if worst == "ratelimit" else attempt)

    if best == "404":
        _FETCH_CACHE[key] = None
    else:
        _FETCH_CACHE[key] = TRANSIENT_FAILURE
        _FETCH_ERRORS[key] = best
    return _FETCH_CACHE[key]


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
                cls = _FETCH_ERRORS.get((repo, version, "manifest"), "unknown")
                transient.append(
                    f"{ref.spec} ({ref.file}:{ref.line}): upstream read failed [{cls}]"
                )
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
    """action spec -> the union of inputs the workflows pass to it.

    Keyed by the full ``repo@ref`` spec, not by repo: the same action is used at
    more than one pin in this tree, and those two manifests can declare different
    input sets, so collapsing them to a repo key would invent or hide a
    difference that is the entire point of the check.
    """
    consumed: dict[str, set[str]] = {}
    for ref in refs:
        if ref.inputs:
            consumed.setdefault(ref.spec, set()).update(ref.inputs)
    return consumed


# The only `using:` values whose manifest declares a flat `inputs:` map that
# GitHub validates a step's `with:` against. A composite action forwards `with:`
# into its own steps, and a docker action takes `args:`, so neither can be
# checked from the manifest alone -- see `parse_action_yml`'s note on
# `runs.steps`. Enforcing those two would manufacture failures, not catch bugs.
ENFORCEABLE_USINGS = ("node16", "node20", "node24", "node12")


def undeclared_input_findings(refs: list[ActionRef]) -> tuple[list[str], list[str], list[str]]:
    """(undeclared inputs, skipped pins, transient API faults).

    A workflow that passes ``with: {foo: bar}`` to an action that declares no
    ``foo`` fails that STEP at run time with ``input "foo" is not defined in
    action owner/repo@ref`` -- so the class is a real, already-shipped defect in
    this repo (see #1272), not a hypothetical.

    Three outcomes, kept apart for the reason `node20_findings` keeps its own
    apart: a pin whose manifest could not be read is NOT a verified input set,
    and grading it as one is exactly the blind run this audit exists to prevent.
    """
    findings: list[str] = []
    skipped: list[str] = []
    transient: list[str] = []

    for ref in refs:
        if not ref.inputs:
            continue
        text = fetch_action_yml(ref.repo, ref.ref)
        if text is None:
            # Already reported by node20_findings as a MISSING manifest; the
            # input set of a nonexistent action is not a finding of its own.
            continue
        if text == TRANSIENT_FAILURE:
            cls = _FETCH_ERRORS.get((ref.repo, ref.ref, "manifest"), "unknown")
            transient.append(f"{ref.spec} ({ref.file}:{ref.line}): upstream read failed [{cls}]")
            continue

        using, declared = parse_action_yml(text)
        if not using.startswith("node"):
            skipped.append(f"{ref.spec} ({ref.file}:{ref.line}): using: {using or 'unknown'}")
            continue
        if using not in ENFORCEABLE_USINGS:
            # e.g. a node22 pin added after this list was written. Refusing to
            # guess is correct: an unenforced pin would read as a checked one.
            skipped.append(f"{ref.spec} ({ref.file}:{ref.line}): unenforced runtime {using}")
            continue

        unknown = sorted(set(ref.inputs) - set(declared))
        if unknown:
            findings.append(
                f"{ref.spec} ({ref.file}:{ref.line}): undeclared input(s) "
                f"{', '.join(unknown)}; declares {', '.join(declared) or '(none)'}"
            )
    return findings, skipped, transient


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

    # Preflight the ONE thing that makes every read fail at once. `gh` without
    # usable credentials refuses all calls, which the per-pin path would
    # otherwise report as 43 separate "API error" lines -- technically honest,
    # but it hides a single root cause behind a wall of identical symptoms.
    # This exact case shipped red: the first CI run of this job failed on a
    # CLEAN tree purely because GH_TOKEN was not in the step env.
    #
    # The probe is `gh auth status`, NOT an env-var check. `gh` can authenticate
    # from a config file, so requiring the env vars would refuse a perfectly
    # capable local run while still being the wrong question for CI.
    try:
        auth = subprocess.run(
            ["gh", "auth", "status"], capture_output=True, text=True, timeout=60
        )
    except (OSError, subprocess.SubprocessError) as exc:
        print(f"FAIL: cannot execute `gh` ({exc}). The audit needs it to read manifests.", file=sys.stderr)
        return 2
    if auth.returncode != 0:
        print(
            "FAIL: `gh auth status` failed, so no manifest can be read and this run "
            "could not reach a verdict.\n"
            "  Export a token, e.g. GH_TOKEN=${{ github.token }} in CI.\n"
            f"  gh said: {(auth.stderr or auth.stdout).strip().splitlines()[:3]}",
            file=sys.stderr,
        )
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

    # The input half of the audit. Run against the SAME refs, and only after the
    # runtime sweep has warmed the manifest cache, so the two checks cannot
    # disagree about what a manifest says.
    inputs_with = consumed_inputs(refs)
    bad_inputs, input_skipped, input_transient = undeclared_input_findings(refs)
    if inputs_with:
        print(f"pins passing `with:` inputs: {len(inputs_with)}")
    print(f"undeclared inputs passed to an action: {len(bad_inputs)}")
    for f in bad_inputs:
        print(f"  UNDECLARED_INPUT  {f}")
    for s in sorted(set(input_skipped)):
        print(f"  INPUTS_UNCHECKED  {s}")

    if transient or input_transient:
        # De-duplicate by (pin, class): one unreadable pin used 3 times is ONE
        # fault, and printing it 3 times buries the diagnosis. Both sweeps'
        # faults are merged rather than reported twice: an unreadable manifest is
        # a single fact that invalidates BOTH checks, and a report that listed it
        # once per check would read as two independent problems.
        seen_faults: dict[str, int] = {}
        for t in transient + input_transient:
            seen_faults[t] = seen_faults.get(t, 0) + 1
        print(f"pins unreadable this run (NOT a verdict): {len(transient) + len(input_transient)} occurrence(s), "
              f"{len(seen_faults)} distinct")
        for t, n in sorted(seen_faults.items()):
            suffix = f"  (x{n} uses)" if n > 1 else ""
            print(f"  TRANSIENT  {t}{suffix}")
        print("audit INCOMPLETE -- a blind run is not a pass")
        return 2

    if problems:
        print(f"pins with no upstream action manifest: {len(problems)}")
        for p in sorted(set(problems)):
            print(f"  MISSING  {p}")
        print("audit INCOMPLETE -- an unresolvable pin is not a pass")
        return 2

    # The verdict. Both failing checks are real defects that would fail a job at
    # run time, so both are exit 1 -- a bumped pin that passes an input the new
    # version dropped is exactly as broken as a bumped pin that reintroduced a
    # deprecated runtime, and reporting one without the other would understate
    # the risk the gate exists to carry.
    if findings:
        print("RESULT: node20 runtimes present")
        return 0 if args.report else 1
    if bad_inputs:
        print(f"RESULT: {len(bad_inputs)} undeclared input(s) passed to an action")
        return 0 if args.report else 1
    print("RESULT: no node20-runtime action reachable from .github/workflows; "
          "no undeclared inputs passed")
    return 0


if __name__ == "__main__":
    sys.exit(main())
