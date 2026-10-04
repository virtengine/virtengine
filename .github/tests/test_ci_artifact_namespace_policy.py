"""Guard for the artifact-namespace collision between the two reusable
workflows that quality-gate.yaml calls in the SAME run:

  .github/workflows/veid-conformance.yaml   uploads conformance-evidence-<os>-<arch>
  .github/workflows/ml-determinism.yaml     uploads conformance-evidence-<os>

Actions artifacts are pooled per RUN, and each workflow's verifier downloads
with a wildcard (`conformance-evidence-*`), so each verifier pulled the other's
artifacts and then died on the foreign directories:

  veid: Missing conformance-output.txt in all-evidence/conformance-evidence-macos-latest
  ml:   Missing conformance evidence directories: [...]  (plus foreign dirs with
        an incompatible JSON schema: veid results have no `test_name`/`hash`)

This test asserts the two upload prefixes are disjoint, and that every
download pattern / expected-name set in each workflow is consistent with its own
upload name. Run: python3 -m unittest discover -s .github/tests -p "test_*policy*.py"
"""
import re
import sys
import unittest
from pathlib import Path

WF = Path(__file__).resolve().parents[1] / "workflows"
VEID = WF / "veid-conformance.yaml"
ML = WF / "ml-determinism.yaml"

# The artifact `name:` sits inside the step's `with:` block, indented deeper
# than the step's own `- name:`, and its value may contain spaces
# (e.g. `${{ matrix.os }}`), so capture the whole remainder of the line.
_WITH_NAME = re.compile(r"^\s{9,}name:\s*(.+?)\s*$", re.M)
_PATTERN = re.compile(r"^\s+pattern:\s*(\S+)\s*$", re.M)
_EVIDENCE_LITERAL = re.compile(r'"([a-z][a-z0-9-]*evidence-[a-z0-9-]+)"')


def read(p):
    return p.read_text(encoding="utf-8")


def blocks(text):
    return text.split("- name:")


def upload_names(text):
    """Artifact names of every upload-artifact step that sets `name:`."""
    out = []
    for b in blocks(text):
        if "actions/upload-artifact" in b:
            m = _WITH_NAME.search(b)
            if m:
                out.append(m.group(1))
    return out


def download_patterns(text):
    out = []
    for b in blocks(text):
        if "actions/download-artifact" in b:
            m = _PATTERN.search(b)
            if m:
                out.append(m.group(1))
    return out


def prefix(literal):
    """`conformance-evidence-${{ matrix.os }}` -> `conformance-evidence-`."""
    return literal.split("${{")[0]


def expected_literals(text):
    """Concrete artifact names quoted in the verifier's expected-name set."""
    return sorted(set(_EVIDENCE_LITERAL.findall(text)))


class NamespaceTest(unittest.TestCase):
    def test_upload_names_are_parsed(self):
        """Guards against a silently vacuous suite: if the YAML shape changes
        and parsing returns nothing, every other assertion here would pass
        without checking anything."""
        for path in (VEID, ML):
            self.assertTrue(
                upload_names(read(path)),
                f"{path.name}: parsed zero upload-artifact names — the other "
                f"assertions in this module would be vacuous",
            )
            self.assertTrue(
                download_patterns(read(path)),
                f"{path.name}: parsed zero download-artifact patterns",
            )

    def test_veid_and_ml_upload_prefixes_are_disjoint(self):
        veid = upload_names(read(VEID))
        ml = upload_names(read(ML))

        veid_prefixes = {prefix(n) for n in veid}
        ml_prefixes = {prefix(n) for n in ml}
        self.assertFalse(
            veid_prefixes & ml_prefixes,
            f"artifact namespace collision on {sorted(veid_prefixes & ml_prefixes)}: "
            f"veid uploads {sorted(veid_prefixes)} and ml uploads {sorted(ml_prefixes)} "
            f"into the SAME run's artifact pool, so each wildcard verifier downloads "
            f"the other's artifacts and then fails on the foreign directories",
        )

    def test_each_wildcard_download_matches_only_its_own_workflow(self):
        for path, own_names, other_names in (
            (VEID, upload_names(read(VEID)), upload_names(read(ML))),
            (ML, upload_names(read(ML)), upload_names(read(VEID))),
        ):
            own = {prefix(n) for n in own_names}
            other = {prefix(n) for n in other_names}
            for pat in download_patterns(read(path)):
                if not pat.endswith("*"):
                    self.assertIn(pat, own_names, f"{path.name}: pattern {pat!r} matches nothing it uploads")
                    continue
                lit = pat[:-1]
                self.assertTrue(
                    any(op.startswith(lit) or lit.startswith(op) for op in own),
                    f"{path.name}: download pattern {pat!r} does not match its own "
                    f"uploads {sorted(own)}",
                )
                clash = sorted(o for o in other if o.startswith(lit) or lit.startswith(o))
                self.assertFalse(
                    clash,
                    f"{path.name}: download pattern {pat!r} ALSO matches the other "
                    f"workflow's artifacts {clash} — foreign directories get downloaded "
                    f"and fail the verifier",
                )

    def test_expected_name_sets_use_the_workflows_own_prefix(self):
        for path in (VEID, ML):
            text = read(path)
            own = {prefix(n) for n in upload_names(text)}
            lits = expected_literals(text)
            self.assertTrue(lits, f"{path.name}: parsed zero expected artifact literals")
            for lit in lits:
                self.assertTrue(
                    any(lit.startswith(p) for p in own),
                    f"{path.name}: expected artifact {lit!r} does not start with any "
                    f"prefix this workflow uploads {sorted(own)} — the verifier would "
                    f"wait for an artifact that is never produced",
                )


if __name__ == "__main__":
    sys.exit(0 if unittest.main(exit=False, verbosity=2).result.wasSuccessful() else 1)
