"""Keep Dependabot PRs on develop and prevent the auto-merger from touching main.

The estate's standing merge policy is two-tier: bot work lands on `develop`,
while `main` is human-only. Dependabot defaults to a repository's default
branch (`main`) unless every update entry says otherwise, so this test protects
both the producer configuration and the consumer's independent guard.
"""
from pathlib import Path
import unittest

import yaml

ROOT = Path(__file__).resolve().parents[2]


class DependabotBaseBranchPolicyTests(unittest.TestCase):
    def test_every_dependabot_update_targets_develop(self):
        config = yaml.safe_load((ROOT / ".github/dependabot.yml").read_text(encoding="utf-8"))
        self.assertEqual(config["version"], 2)
        updates = config["updates"]
        self.assertTrue(updates, "Dependabot config has no update entries")
        wrong = [
            f"{entry.get('package-ecosystem')} {entry.get('directory')}: "
            f"target-branch={entry.get('target-branch')!r}"
            for entry in updates
            if entry.get("target-branch") != "develop"
        ]
        self.assertEqual(
            wrong,
            [],
            "Every Dependabot update key must target develop; otherwise Dependabot "
            "defaults to main, which is human-only.",
        )

    def test_auto_merge_is_restricted_to_dependabot_prs_based_on_develop(self):
        workflow = yaml.safe_load(
            (ROOT / ".github/workflows/dependabot-auto-merge.yaml").read_text(encoding="utf-8")
        )
        # PyYAML's YAML 1.1 loader parses the unquoted GitHub `on:` key as True.
        trigger = workflow.get("on", workflow.get(True))
        self.assertEqual(trigger, "pull_request")

        job = workflow["jobs"]["dependabot"]
        condition = job["if"].replace("\n", " ")
        self.assertIn("github.actor == 'dependabot[bot]'", condition)
        self.assertIn("github.base_ref == 'develop'", condition)

        # The guard must cover the whole job, including the approval step, not
        # just the merge command; otherwise a main-based PR may still be approved.
        step_names = [step["name"] for step in job["steps"]]
        self.assertIn("Auto-approve patch & minor updates", step_names)
        self.assertIn("Enable auto-merge (squash)", step_names)


if __name__ == "__main__":
    unittest.main()
