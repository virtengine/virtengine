"""No workflow in this repo may make a develop push unverifiable.

The defect
----------
``concurrency`` with ``cancel-in-progress: true`` and a group scoped only to
``github.ref`` puts every run for a ref in ONE group. With
``cancel-in-progress: true`` a later commit kills its predecessor -- and a
**cancelled run publishes NO verdict**. The failure mode is therefore not "the
gate is red" but "there is no gate". Measured on this repo:

  * ``ci.yaml``: 35 consecutive develop push runs ended ``cancelled`` over 32h.
  * ``quality-gate.yaml``: 7 develop push runs cancelled, each killed by a run
    created inside its own window (overlaps 34s-658s).
  * ``supply-chain.yaml``: 1 develop push run cancelled, killed 65s in.
  * ``license-compliance.yaml``: 1 cancelled run, NOT explained by overlap.

What is pinned
--------------
1. Every workflow that can see a develop push scopes its push group per-commit
   (``github.sha``), gated on the push event, so no push can cancel another.
2. The pull_request path stays per-ref and cancellable. Cancelling a superseded
   PR run is what makes iteration fast; serialising PR runs would trade a real
   problem for a worse one.
3. Reusable workflows namespace on ``github.workflow``. In a called workflow
   ``github.workflow`` resolves to the CALLER's name, so a group omitting it
   puts every caller in one shared group per ref -- ``ci.yaml``'s push run was
   cancelled by ``quality-gate.yaml``'s run for the same commit.

Exposure is DERIVED, not listed
-------------------------------
The first version of this fix was scoped to ``ci.yaml`` alone and left the
disease in five sibling files. Hand-maintaining that list is how it then missed
two workflows, so this test computes exposure instead:

  * a workflow's own ``on: push`` branch filter, OR
  * for a reusable workflow, the union of the push branch filters of every
    workflow that calls it (``uses: ./.github/workflows/<name>``).

The reusable derivation is the one that catches the real cases. Read from the
``on:`` block alone, ``compatibility.yaml``/``ml-determinism.yaml``/
``veid-e2e.yaml`` look safe because they do not list ``develop`` -- but they are
invoked by workflows that do, and in a reusable call ``github.ref`` is the
caller's ref.

Why these assert the PARSED block
---------------------------------
The affected workflows carry explanatory comments whose vocabulary deliberately
repeats the words ``cancel-in-progress``, ``group`` and ``push``. A substring
search over raw text would match a comment instead of the live setting. Same
class as the 2026-10-05 integration-keeper note on ``x is not`` on tuples.

The ``github.sha`` expression is not evaluated (there is no GitHub context here),
so the group is asserted by STRUCTURE: it must reference ``github.sha`` and gate
that reference on ``github.event_name == 'push'``.
"""

from __future__ import annotations

import os
import re
import unittest
from pathlib import Path

import yaml

# Overridable so the falsifier can drive this suite against a throwaway copy of
# .github/workflows with one real block reverted, instead of mutating the
# checkout. Absent the variable this is the repo root, so CI is unaffected.
REPO_ROOT = Path(
    os.environ.get("CONCURRENCY_TEST_ROOT") or Path(__file__).resolve().parents[2]
)
WORKFLOWS = REPO_ROOT / ".github" / "workflows"

# Workflows that deliberately do NOT cancel in flight. Named so the comment in
# ci.yaml and any future sibling cannot rot into fiction.
UNCANCELLING_WORKFLOWS = (
    "infrastructure.yaml",
    "dr-tools-image.yaml",
    "dr-failover-test.yaml",
)


def _load(name: str) -> dict:
    return yaml.safe_load((WORKFLOWS / name).read_text(encoding="utf-8"))


def _workflow_names() -> list[str]:
    return sorted(p.name for p in WORKFLOWS.glob("*.y*ml"))


def _on(doc: dict):
    # PyYAML resolves the bare key `on` to the boolean True.
    return doc.get(True, doc.get("on"))


def _push_branches(doc: dict) -> set | None:
    """Branches this workflow's own push trigger accepts.

    Returns ``None`` for "every branch", which includes develop.
    """
    on = _on(doc)
    if isinstance(on, list):
        return None if "push" in [str(x) for x in on] else set()
    if isinstance(on, dict):
        push = on.get("push")
        if push is None:
            return set()
        if isinstance(push, dict):
            branches = push.get("branches")
            return set(branches) if branches else None
        return None  # `on: {push: null}` == every branch
    if on == "push":
        return None
    return set()


def _is_reusable(doc: dict) -> bool:
    on = _on(doc)
    return isinstance(on, dict) and "workflow_call" in on


def _has_pr_trigger(doc: dict) -> bool:
    """Does this workflow ever run on pull_request?

    A workflow with no pull_request trigger has no PR runs to serialise, so an
    unconditional ``github.sha`` in its group is harmless. The PR-shape
    assertions are scoped to workflows where that is not true.
    """
    on = _on(doc)
    if isinstance(on, list):
        return "pull_request" in [str(x) for x in on]
    if isinstance(on, dict):
        return "pull_request" in on
    return False


def _callers_of(target: str, docs: dict) -> set:
    """Workflows that invoke ``target`` as a reusable workflow."""
    out = set()
    for name in docs:
        text = (WORKFLOWS / name).read_text(encoding="utf-8")
        if re.search(rf"uses:\s*\./\.github/workflows/{re.escape(target)}\b", text):
            out.add(name)
    return out


def _sees_develop(name: str, docs: dict) -> bool:
    """Does any run of this workflow execute against the develop ref?"""
    doc = docs[name]
    if _is_reusable(doc):
        callers = _callers_of(name, docs)
        if callers:
            for caller in callers:
                branches = _push_branches(docs[caller])
                if branches is None or "develop" in branches:
                    return True
            return False
    branches = _push_branches(doc)
    return branches is None or "develop" in branches


def _concurrency(doc: dict) -> dict:
    conc = doc.get("concurrency")
    return conc if isinstance(conc, dict) else {}


class PushPathPublishesAVerdictTest(unittest.TestCase):
    """The core defect, asserted against every derived-exposed workflow."""

    def setUp(self):
        self.docs = {n: _load(n) for n in _workflow_names()}

    def _exposed(self) -> dict:
        return {
            n: self.docs[n]
            for n in self.docs
            if _sees_develop(n, self.docs)
            and _concurrency(self.docs[n]).get("cancel-in-progress") is True
        }

    def test_at_least_one_workflow_is_covered(self):
        # If this fails, the exposure derivation silently returned nothing and
        # every other assertion below is vacuous. An empty result set is a
        # suspect, not an all-clear.
        self.assertGreaterEqual(
            len(self._exposed()),
            5,
            "the derivation found fewer exposed workflows than the six measured "
            "on 2026-10-06; if the workflows were deleted that is fine, but say "
            "so rather than passing vacuously",
        )

    def test_no_exposed_workflow_loses_push_verdicts(self):
        offenders = []
        for name, doc in sorted(self._exposed().items()):
            group = str(_concurrency(doc).get("group") or "")
            if "github.sha" not in group:
                offenders.append(f"{name}: {group!r}")
        self.assertEqual(
            offenders,
            [],
            "each of these can see a develop push but scopes its concurrency "
            "group per-ref only, so cancel-in-progress cancels the in-flight "
            "run and NO verdict is published:\n  " + "\n  ".join(offenders),
        )

    def test_the_per_commit_scoping_is_conditional_on_push(self):
        for name, doc in sorted(self._exposed().items()):
            if not _has_pr_trigger(doc):
                continue  # see test_pr_shape_is_only_required_where_pr_runs_exist
            with self.subTest(workflow=name):
                group = str(_concurrency(doc).get("group") or "")
                self.assertIn("github.sha", group, "sanity: derived exposed")
                self.assertIn(
                    "github.event_name",
                    group,
                    "the github.sha component must be gated on github.event_name "
                    "so pull_request runs keep the per-ref, cancellable group",
                )
                self.assertRegex(
                    group,
                    r"event_name\s*==\s*'push'",
                    "the per-commit component must apply to the push event "
                    f"specifically; found: {group!r}",
                )

    def test_pr_shape_is_only_required_where_pr_runs_exist(self):
        # The control for the two assertions above. security.yaml never runs on
        # pull_request, so an UNCONDITIONAL github.sha there serialises nothing
        # and is correct; demanding the push-gated form would be a false
        # positive. This test exists so that narrowing is deliberate: if
        # security.yaml ever gains a pull_request trigger, the narrowing stops
        # applying to it and the strict assertions resume.
        doc = _load("security.yaml")
        self.assertFalse(
            _has_pr_trigger(doc),
            "security.yaml gained a pull_request trigger; the unconditional "
            "github.sha in its group now serialises PR runs and must be "
            "changed to the push-gated form",
        )
        self.assertIn("github.sha", str(_concurrency(doc)["group"]))

    def test_pull_request_path_stays_per_ref_and_cancellable(self):
        for name, doc in sorted(self._exposed().items()):
            if not _has_pr_trigger(doc):
                continue  # see test_pr_shape_is_only_required_where_pr_runs_exist
            with self.subTest(workflow=name):
                group = str(_concurrency(doc).get("group") or "")
                self.assertIn(
                    "'pr'",
                    group,
                    "the non-push branch of the group must yield a constant "
                    "per-ref value, not the sha, so pull_request runs stay in "
                    "one cancellable group per ref",
                )

    def test_reusable_workflows_namespace_on_the_caller(self):
        # A second, independent defect: a called workflow that omits
        # github.workflow shares one group per ref across ALL its callers.
        offenders = []
        for name, doc in sorted(self.docs.items()):
            if not _is_reusable(doc):
                continue
            callers = _callers_of(name, self.docs)
            if not callers:
                continue
            if _concurrency(doc).get("cancel-in-progress") is not True:
                continue
            group = str(_concurrency(doc).get("group") or "")
            if "github.workflow" not in group:
                who = ", ".join(sorted(callers))
                offenders.append(f"{name} (callers: {who}): {group!r}")
        self.assertEqual(
            offenders,
            [],
            "these REUSABLE workflows cancel in progress without namespacing "
            "on github.workflow, so callers cancel each other's runs for the "
            "same commit:\n  " + "\n  ".join(offenders),
        )

    def test_cancel_in_progress_is_declared_explicitly(self):
        for name, doc in sorted(self._exposed().items()):
            with self.subTest(workflow=name):
                self.assertIs(
                    _concurrency(doc).get("cancel-in-progress"),
                    True,
                    "cancel-in-progress must stay an explicit, pinned decision",
                )


class SiblingWorkflowTradeIsUnchangedTest(unittest.TestCase):
    """The comments cite these three; do not let the citation rot."""

    def test_the_uncancelling_workflows_are_still_that_way(self):
        for name in UNCANCELLING_WORKFLOWS:
            with self.subTest(workflow=name):
                self.assertIs(
                    _concurrency(_load(name)).get("cancel-in-progress"),
                    False,
                    f"{name} is cited as an existing uncancellable workflow; if "
                    "it changed, the comment is now wrong",
                )


class MutantDiscriminationTest(unittest.TestCase):
    """Proves the assertions can fail. All-INERT guards measure nothing."""

    def _catch(self, group: str) -> bool:
        """Would the push-verdict assertions reject this group? True == caught."""
        ok = (
            "github.sha" in group
            and "github.event_name" in group
            and re.search(r"event_name\s*==\s*'push'", group) is not None
            and "'pr'" in group
        )
        return not ok

    def test_control_unmutated_group_is_accepted(self):
        # The no-op control. If this fails the harness is not measuring the
        # property at all -- it is always answering "caught".
        live = _concurrency(_load("ci.yaml"))["group"]
        self.assertFalse(
            self._catch(live),
            "the live ci.yaml group was rejected by the harness; the harness "
            "does not implement the property it is supposed to test",
        )

    def test_the_pre_fix_group_is_caught(self):
        # The exact configuration measured in production.
        self.assertTrue(
            self._catch("${{ github.workflow }}-${{ github.ref }}"),
            "the pre-fix group must be rejected: it produced 35 consecutive "
            "verdicts-less cancelled runs",
        )

    def test_unconditional_sha_is_caught(self):
        self.assertTrue(
            self._catch("${{ github.workflow }}-${{ github.ref }}-${{ github.sha }}"),
            "an unconditional github.sha must be caught: it looks right and "
            "serialises pull_request runs",
        )

    def test_push_test_inverted_is_caught(self):
        self.assertTrue(
            self._catch(
                "${{ github.workflow }}-${{ github.ref }}-"
                "${{ github.event_name == 'pull_request' && github.sha || 'pr' }}"
            ),
            "an inverted event test must be caught: PR runs would get the "
            "per-commit group and push runs the shared one",
        )


class ExposureDerivationTest(unittest.TestCase):
    """The derivation itself is the load-bearing part of this file."""

    def setUp(self):
        self.docs = {n: _load(n) for n in _workflow_names()}

    def test_reusable_workflow_is_exposed_through_its_caller(self):
        # compatibility.yaml's own on: block does not list develop. It is still
        # exposed because ci.yaml calls it and ci.yaml pushes to develop.
        self.assertNotIn(
            "develop", _push_branches(self.docs["compatibility.yaml"]) or set()
        )
        self.assertTrue(
            _sees_develop("compatibility.yaml", self.docs),
            "a reusable workflow must inherit its callers' push branches",
        )

    def test_workflow_with_no_develop_path_is_not_exposed(self):
        # The control for the derivation: pull_request-only.
        self.assertFalse(
            _sees_develop("labeler.yaml", self.docs),
            "labeler.yaml is pull_request-only and must not be reported",
        )

    def test_own_push_filter_is_used_for_non_reusable(self):
        self.assertTrue(_sees_develop("supply-chain.yaml", self.docs))
        self.assertTrue(_sees_develop("quality-gate.yaml", self.docs))


if __name__ == "__main__":  # pragma: no cover
    unittest.main()
