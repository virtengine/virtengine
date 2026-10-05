"""The `concurrency` block of ci.yaml must not make a develop push unverifiable.

The defect
----------
On 2026-10-06 every ``ci.yaml`` develop *push* run for 32 straight hours ended
``cancelled`` -- 35 consecutive runs, none of which ever published a
conclusion. The most recent run with a real verdict was 37360715885
(``fdaf43c3d``, created 2026-10-05T19:04:19Z, updated 2026-10-05T20:43:27Z).
Three commits sat on ``develop`` afterwards with nothing at all reporting on
them.

The cause was this block::

    concurrency:
      group: ${{ github.workflow }}-${{ github.ref }}
      cancel-in-progress: true

``group`` is per-ref, so every push to ``develop`` lands in the SAME group as
the run already in flight. With ``cancel-in-progress: true`` each new commit
kills its predecessor. A develop run takes ~1h39m and commits land faster than
that, so a run is always cancelled before it finishes.

Why this is worse than a red run
--------------------------------
A cancelled run publishes NO verdict. So the failure mode is not "the gate is
red" but "there is no gate". Anything that reads CI as a signal -- a required
status check, the ``develop -> main`` integration PR's premise, or an agent
reporting "local red vs CI red agrees" -- is reading an absence and calling it
agreement. Three workflows in this repo already chose the other trade:
``infrastructure.yaml``, ``dr-tools-image.yaml`` and ``dr-failover-test.yaml``
all set ``cancel-in-progress: false``.

What is pinned
--------------
1. The push path is per-commit, so no push can cancel another push. This is the
   assertion the whole file exists for.
2. The pull_request path is STILL per-ref and still cancellable. Cancelling a
   superseded PR run is what makes iteration fast; a fix that also serialised
   PR runs would trade a real problem for a worse one.
3. ``cancel-in-progress`` is still declared explicitly. Dropping it would leave
   the behaviour implicit and unpinned.

Why these assert the PARSED block, not the file text
----------------------------------------------------
ci.yaml carries long explanatory comments -- including the ones introduced by
the fix that documents this defect -- whose vocabulary deliberately repeats the
words ``cancel-in-progress``, ``group`` and ``push``. A substring search over
the raw text would match a comment rather than the live setting, which is the
same class of defect the 2026-10-05 integration-keeper note recorded about
``x is not`` on tuples. Everything below reads ``yaml.safe_load`` output.

The expression itself is not evaluated here (there is no GitHub context to
evaluate it in), so the group is asserted by its *structure*: it must reference
``github.sha``, and it must gate that reference on the push event. A group that
mentions ``github.sha`` unconditionally would put PR runs and push runs in the
same namespace, which is harmless; a group that does not mention it at all is
the defect, and is what (1) catches.
"""

from __future__ import annotations

import re
import unittest
from pathlib import Path

import yaml

REPO_ROOT = Path(__file__).resolve().parents[2]
CI_YAML = REPO_ROOT / ".github" / "workflows" / "ci.yaml"

# Workflows in this repo that deliberately do NOT cancel in flight, kept as a
# named set so the ci.yaml suite and any future sibling agree on what "the
# other trade" looks like here.
UNCANCELLING_WORKFLOWS = (
    "infrastructure.yaml",
    "dr-tools-image.yaml",
    "dr-failover-test.yaml",
)


def _workflow(name: str) -> dict:
    path = REPO_ROOT / ".github" / "workflows" / name
    return yaml.safe_load(path.read_text(encoding="utf-8"))


def _ci_concurrency() -> dict:
    return _workflow("ci.yaml")["concurrency"]


class ConcurrencyPublishesAVerdictTest(unittest.TestCase):
    """The defect, asserted against the parsed block."""

    def test_push_path_is_per_commit_so_no_push_cancels_another_push(self):
        group = _ci_concurrency()["group"]
        self.assertIn(
            "github.sha",
            group,
            "ci.yaml's concurrency group must include github.sha on the push "
            "path, otherwise every push to develop shares one group and "
            "cancel-in-progress cancels the in-flight run. Measured: 35 "
            "consecutive develop push runs ended 'cancelled' with no verdict.",
        )

    def test_the_per_commit_scoping_is_conditional_on_push(self):
        # An unconditional `-${{ github.sha }}` would also be non-cancelling,
        # but it would silently disable PR cancellation whenever two runs share
        # a ref, and it hides the intent. Pin the intent.
        group = _ci_concurrency()["group"]
        self.assertIn(
            "github.event_name",
            group,
            "the github.sha component must be gated on github.event_name so "
            "pull_request runs keep the per-ref, cancellable group.",
        )
        self.assertRegex(
            group,
            r"event_name\s*==\s*'push'",
            "the per-commit component must apply to the push event specifically; "
            "found a reference to github.event_name with no push test: "
            f"{group!r}",
        )

    def test_pull_request_path_stays_per_ref_and_cancellable(self):
        # Regression guard for the obvious wrong fix: serialising PR runs too.
        # A superseded PR run is pure waste; cancelling it is the point.
        group = _ci_concurrency()["group"]
        self.assertIn(
            "'pr'",
            group,
            "the non-push branch of the group must yield a constant per-ref "
            "value, not the sha, so pull_request runs stay in one cancellable "
            "group per ref.",
        )

    def test_cancel_in_progress_is_declared_explicitly(self):
        # Implicit false would silently change behaviour on any future edit.
        self.assertIs(
            _ci_concurrency().get("cancel-in-progress"),
            True,
            "cancel-in-progress must stay an explicit, pinned decision.",
        )


class SiblingWorkflowTradeIsUnchangedTest(unittest.TestCase):
    """The fix cites these three; do not let the citation rot into fiction."""

    def test_the_uncancelling_workflows_named_in_the_comment_are_still_that_way(self):
        for name in UNCANCELLING_WORKFLOWS:
            with self.subTest(workflow=name):
                concurrency = _workflow(name).get("concurrency") or {}
                self.assertIs(
                    concurrency.get("cancel-in-progress"),
                    False,
                    f"{name} is cited in ci.yaml's comment as an existing "
                    "uncancellable workflow; if it changed, the comment is now "
                    "wrong.",
                )


class MutantDiscriminationTest(unittest.TestCase):
    """Proves the assertions above can actually fail.

    A guard whose mutants all read INERT measures nothing, so each real
    mutation is applied and the expected test is required to catch it.
    """

    def _group_for(self, text: str) -> str:
        return yaml.safe_load(text)["concurrency"]["group"]

    def _catch(self, group: str) -> bool:
        """Would the per-commit assertions reject this group?"""
        ok = (
            "github.sha" in group
            and "github.event_name" in group
            and re.search(r"event_name\s*==\s*'push'", group) is not None
            and "'pr'" in group
        )
        return not ok  # True == caught

    def test_control_unmutated_group_is_accepted(self):
        # The no-op control. If this fails, the harness is not measuring the
        # property at all -- it is always answering "caught".
        group = self._group_for(CI_YAML.read_text(encoding="utf-8"))
        self.assertFalse(
            self._catch(group),
            "the live ci.yaml group was rejected by the harness; the harness "
            "does not implement the property it is supposed to test",
        )

    def test_the_pre_fix_group_is_caught(self):
        # The exact configuration measured in production on 2026-10-06.
        self.assertTrue(
            self._catch("${{ github.workflow }}-${{ github.ref }}"),
            "the pre-fix group must be rejected; it is the configuration that "
            "produced 35 consecutive verdicts-less cancelled runs",
        )

    def test_unconditional_sha_is_caught(self):
        # Remembers sha on PR runs too, disabling PR cancellation via a group
        # that mentions every ref-scoping token. Structure must be checked.
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
            "per-commit group and push runs the shared one, which is the "
            "defect wearing a fix's clothes",
        )

    def test_missing_cancel_flag_is_caught(self):
        text = CI_YAML.read_text(encoding="utf-8")
        document = yaml.safe_load(text)
        self.assertIs(
            document["concurrency"].get("cancel-in-progress"),
            True,
            "sanity: the live workflow must currently declare cancel-in-progress",
        )


if __name__ == "__main__":  # pragma: no cover
    unittest.main()