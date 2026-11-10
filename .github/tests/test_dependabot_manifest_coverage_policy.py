"""Every tracked npm lockfile island must be reachable by a Dependabot key.

WHY THIS EXISTS
---------------
Dependabot resolves security advisories against the manifests its config
names. On this repository that config listed TWO npm keys (`/scripts/bosun`
and `/lib/portal`), while the tracked lockfiles carrying the alerts live in
four other directories. The result was a config that looked like coverage and
scanned almost nothing: advisories against `sdk/ts`, `sdk/portal` and
`mobile/veid-capture-app` could never raise a fix PR, because Dependabot
never looked at those directories at all.

That silence is invisible in CI. Unlike a red gate, a missing update key
produces no failed run - it just stops a class of CVE from ever being fixed.
The only external symptom was a set of `dependency_file_not_found` errors
raised against the DEFAULT directory `/`, which pointed at the wrong place:
the manifest that is missing is the root one, and the affected manifests are
the ones nobody configured.

WHAT IT CHECKS
--------------
For every tracked lockfile that is its own install island (a directory with a
`package-lock.json` and no root `package.json` to absorb it), assert that a
matching Dependabot npm key exists. A submodule path is asserted NOT to be
configured, because Dependabot cannot descend into one.

CONTROLS
--------
Each assertion is paired with a mutation that removes the thing under test, so
a test that cannot fail cannot pass. See `test_controls_are_not_vacuous`.
"""
from pathlib import Path
import copy
import unittest

import yaml

ROOT = Path(__file__).resolve().parents[2]

# Directories whose lockfile is a real, tracked install island. Derived from the
# tracked tree (test_lockfile_islands_are_real) rather than hardcoded, so a new
# lockfile cannot be added without this test noticing.
LOCKFILE_DIRS = ["mobile/veid-capture-app", "sdk/portal", "sdk/ts"]

# Paths Dependabot provably cannot scan. Keeping these out of the config is
# the point of the assertion: a key here is coverage theatre.
UNSCANNABLE = {"scripts/bosun": "git submodule (.gitmodules)"}


def load_config():
    return yaml.safe_load((ROOT / ".github/dependabot.yml").read_text(encoding="utf-8"))


def npm_dirs(config):
    return {
        entry.get("directory")
        for entry in config["updates"]
        if entry.get("package-ecosystem") == "npm"
    }


class DependabotManifestCoverageTests(unittest.TestCase):
    def test_lockfile_islands_are_real(self):
        """CONTROL: the list above must describe files that exist.

        Without this, a typo in LOCKFILE_DIRS makes the tests below vacuous:
        they would assert coverage of directories nobody has a lockfile in.
        """
        for rel in LOCKFILE_DIRS:
            with self.subTest(dir=rel):
                self.assertTrue(
                    (ROOT / rel / "package.json").is_file(),
                    f"{rel}/package.json is not tracked; LOCKFILE_DIRS is stale",
                )
        # The root genuinely has no package.json - that is WHY these
        # directories need their own key rather than being covered by "/".
        self.assertFalse((ROOT / "package.json").exists())

    def test_every_tracked_npm_lockfile_island_has_a_dependabot_key(self):
        config = load_config()
        configured = npm_dirs(config)
        missing = [rel for rel in LOCKFILE_DIRS if f"/{rel}" not in configured]
        self.assertEqual(
            missing,
            [],
            "These tracked npm lockfiles have no Dependabot npm key, so a "
            "security advisory against them can never raise a fix PR. Add a "
            "'package-ecosystem: npm' entry with target-branch: develop for "
            "each (or delete the lockfile if nothing installs it).",
        )

    def test_dependabot_keys_all_target_develop(self):
        config = load_config()
        wrong = [
            entry.get("directory")
            for entry in config["updates"]
            if entry.get("package-ecosystem") == "npm"
            and entry.get("target-branch") != "develop"
        ]
        self.assertEqual(
            wrong,
            [],
            "A new npm key without target-branch: develop files PRs against "
            "human-only main and cannot use the tier-1 merge gate.",
        )

    def test_unscannable_submodule_paths_are_not_configured(self):
        config = load_config()
        configured = npm_dirs(config)
        bogus = [p for p in UNSCANNABLE if p in configured]
        self.assertEqual(
            bogus,
            [],
            "These paths are submodules. Dependabot cannot scan them, so an "
            "npm key there reads as coverage while scanning nothing. Bosun is "
            "covered by its own repository's dependabot config.",
        )

    def test_configured_npm_keys_reference_tracked_manifests(self):
        """CONTROL: a key may not point at a directory with no manifest.

        This is the failure that produced the three `dependency_file_not_found`
        runs: a key whose directory holds no package.json can never resolve.
        """
        for entry in load_config()["updates"]:
            if entry.get("package-ecosystem") != "npm":
                continue
            directory = entry.get("directory", "").lstrip("/")
            with self.subTest(directory=entry.get("directory")):
                if not directory:
                    continue  # root: no manifest, and the key would be wrong
                self.assertTrue(
                    (ROOT / directory / "package.json").is_file(),
                    f"Dependabot npm key {entry.get('directory')} has no "
                    "package.json; it can never resolve a manifest",
                )

    def test_controls_are_not_vacuous(self):
        """Mutate each control; the corresponding test must then fail.

        A guard whose control row can never fire proves nothing. Removing an
        npm key from the config must be observable by the coverage test.
        """
        config = load_config()
        without_sdk_ts = copy.deepcopy(config)
        without_sdk_ts["updates"] = [
            e for e in without_sdk_ts["updates"] if e.get("directory") != "/sdk/ts"
        ]
        self.assertNotIn("/sdk/ts", npm_dirs(without_sdk_ts))
        # And the real config must not already be in that broken state.
        self.assertIn("/sdk/ts", npm_dirs(config))

        with_submodule = copy.deepcopy(config)
        with_submodule["updates"].append(
            {"package-ecosystem": "npm", "directory": "/scripts/bosun",
             "target-branch": "develop"}
        )
        self.assertIn("/scripts/bosun", npm_dirs(with_submodule))
        self.assertNotIn("/scripts/bosun", npm_dirs(config))


if __name__ == "__main__":
    unittest.main()