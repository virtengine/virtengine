"""Guard for the DR Tools Image re-pin PR loop.

The DR Tools Image workflow publishes ghcr.io/virtengine/dr-tools on every
build and then re-pins infra/kubernetes/dr/backup-cronjobs.yaml to whatever
digest that same run produced. The pin was minted per-commit:

    branch: dr-tools/repin-${{ github.sha }}

so every publish opened its OWN branch and its own PR instead of updating the
one re-pin PR. That produced a fan of competing stale re-pins:

  #1057  sha256:9f5fcb72   2026-09-29T00:06:47Z
  #1092  sha256:180c19f0   2026-09-29T06:18:59Z
  #1131  sha256:48c95a64   2026-10-01T19:39:06Z

Each one is MERGEABLE and each one names a digest that is merely *older* than
the next, so the merge gate sees no conflict and an unattended merge can land a
DR rollback (an older recovery image than the branch already carries).

The branch name is now the stable `dr-tools/repin`, so there is exactly one
re-pin PR and it always carries the newest digest. This test asserts that
property. Run:

  python -m unittest discover -s .github/tests -p "test_*policy*.py"
"""
import re
import unittest
from pathlib import Path

WF = Path(__file__).resolve().parents[1] / "workflows" / "dr-tools-image.yaml"
CRONJOBS = Path(__file__).resolve().parents[2] / "infra" / "kubernetes" / "dr" / "backup-cronjobs.yaml"

# `branch:` appears in several places (on: push/pull_request, git config).
# The re-pin branch is the one handed to the create-pull-request action.
# NOTE: an OCI digest reference uses `@sha256:`, not `:sha256:` -- getting that
# separator wrong makes this test assert an empty set and pass vacuously.
_BRANCH_LINE = re.compile(r"^\s*branch:\s*(.+?)\s*$", re.M)
_REPIN_PREFIX = "dr-tools/repin"
# GitHub expressions that make the branch vary per build.
_EXPR = re.compile(r"\$\{\{|\$\(")


def read(p):
    return p.read_text(encoding="utf-8")


def repin_branches(text):
    """Branch names minted by the re-pin job for the DR Tools image."""
    out = []
    for b in text.split("- name:"):
        if "create-pull-request" in b:
            for m in _BRANCH_LINE.finditer(b):
                out.append(m.group(1))
    return out


class TestDrToolsPinBranchIsStable(unittest.TestCase):
    def test_workflow_exists(self):
        self.assertTrue(WF.is_file(), f"missing {WF}")

    def test_repin_branch_is_stable(self):
        """The re-pin branch must not embed github.sha (or any other expr).

        A per-commit branch is what fanned one PR per publish and left a pile
        of mergeable-but-stale re-pins competing to roll the DR image back.
        """
        branches = repin_branches(read(WF))
        self.assertTrue(
            branches,
            "no create-pull-request branch found in dr-tools-image.yaml; "
            "the re-pin PR branch name is no longer asserted by this test",
        )
        for b in branches:
            self.assertNotRegex(
                b,
                _EXPR,
                f"re-pin branch {b!r} embeds a GitHub expression, so it varies "
                "per build and mints a competing PR every publish; the branch "
                "must be a stable literal (dr-tools/repin)",
            )

    def test_repin_branch_keeps_prefix(self):
        """The branch must stay under the dr-tools/repin namespace."""
        for b in repin_branches(read(WF)):
            self.assertTrue(
                b.startswith(_REPIN_PREFIX),
                f"re-pin branch {b!r} lost the {_REPIN_PREFIX!r} prefix",
            )

    def test_repin_branch_is_a_literal_not_a_template(self):
        """A literal branch name, so `delete-branch` and reruns converge."""
        branches = repin_branches(read(WF))
        self.assertEqual(
            branches,
            [_REPIN_PREFIX],
            f"expected exactly one stable re-pin branch {_REPIN_PREFIX!r}, "
            f"found {branches}",
        )


class TestCronjobPinIsAFullDigest(unittest.TestCase):
    def test_cronjobs_exist(self):
        self.assertTrue(CRONJOBS.is_file(), f"missing {CRONJOBS}")

    def test_every_dr_tools_image_is_pinned_by_full_digest(self):
        """Every dr-tools reference is a full 64-hex sha256, never a tag.

        A tag or a truncated digest makes the DR recovery image unresolvable,
        which is how the CronJobs ended up in ImagePullBackOff.
        """
        text = read(CRONJOBS)
        images = re.findall(r"ghcr\.io/virtengine/dr-tools@(\S+)", text)
        self.assertTrue(
            images,
            "no ghcr.io/virtengine/dr-tools reference found in backup-cronjobs.yaml",
        )
        for ref in set(images):
            self.assertRegex(
                ref,
                r"^sha256:[0-9a-f]{64}$",
                f"dr-tools image {ref!r} is not a full sha256 digest pin",
            )

    def test_all_dr_tools_pins_agree(self):
        """One digest for the whole DR chain.

        Divergent pins mean the backup, provider-state and test CronJobs can
        run different dr-tools builds, so a DR drill passes against an image
        that recovery would not actually use.
        """
        text = read(CRONJOBS)
        refs = set(
            re.findall(r"ghcr\.io/virtengine/dr-tools@(sha256:[0-9a-f]{64})", text)
        )
        self.assertEqual(
            len(refs),
            1,
            f"DR CronJobs disagree on the dr-tools digest: {sorted(refs)}",
        )


if __name__ == "__main__":
    unittest.main()
