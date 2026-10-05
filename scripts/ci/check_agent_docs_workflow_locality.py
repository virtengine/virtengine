#!/usr/bin/env python3
"""Guard: no agent doc may cite the `bosun` npm publisher as a workflow of THIS repo.

Why this exists
---------------
`.github/AGENTS.md` cited `bosun-publish.yaml` as one of this repository's workflows
in four places. No such file has ever existed in virtengine/virtengine:

  git log --all -- .github/workflows/bosun-publish.yaml     # empty
  git ls-tree -r --name-only refs/remotes/origin/develop -- .github/workflows/ | wc -l   # 37
  gh workflow list -R virtengine/virtengine                 # no publisher

The real publisher lives in another repository:
`virtengine/bosun/.github/workflows/publish.yaml` (`name: "Publish to npm"`,
`NODE_VERSION: "24"`, jobs `check` + `publish`, `environment: npm-publish`).

Why a classifier and not a bare grep
------------------------------------
A bare `grep bosun-publish` is VACUOUS here: the *compliant* line is one that MENTIONS
`publish.yaml` in order to DENY that it belongs here. The *violating* line is one that
asserts it unqualified. Both match the same pattern, so grep fires on the fix and stays
silent on the bug. Every mentioning line is therefore classified:

  REFUTATION  - explicitly scopes the workflow to another repository
                 ("NOT in this repository", "virtengine/bosun", "out-of-repo", ...)
                 -> compliant; this is the documented correction shape.
  INSTRUCTION - mentions it with no out-of-repo qualifier, so the reader would
                 conclude this repo runs it -> VIOLATION.

Exit 0 when there are no violations, 1 when there are. `--selftest` proves the
predicate can fire on both shapes, so "0 findings" is not blindness.

Usage:
  python3 scripts/ci/check_agent_docs_workflow_locality.py [paths...]
  python3 scripts/ci/check_agent_docs_workflow_locality.py --selftest
"""

from __future__ import annotations

import argparse
import pathlib
import re
import sys
import tempfile

REPO_ROOT = pathlib.Path(__file__).resolve().parents[2]

# Files that document this repository's CI. Cite-only checks belong to these.
DEFAULT_TARGETS = (
    ".github/AGENTS.md",
    "AGENTS.md",
)

# Matches the phantom `bosun-publish.yaml` plus the real (out-of-repo) `publish.yaml`.
# The negative lookbehind is load-bearing: `sdk-publish.yaml` IS a genuine workflow of
# this repo (its jobs are `publish-python`, `publish-rust`), so it must not match.
TOKEN = re.compile(r"bosun-publish\.yaml|(?<!sdk-)publish\.yaml", re.IGNORECASE)

# Markers that explicitly place the workflow outside virtengine/virtengine.
OUT_OF_REPO = (
    "not in this repository",
    "not in this repo",
    "no npm publisher",
    "another repo",
    "another repository",
    "out-of-repo",
    "out of repo",
    "other repo",
    "virtengine/bosun",
    "bosun repo",
    "does not exist here",
    "never existed",
    "no such workflow",
)

# Markdown syntax stripped before matching so `_docs/...`, `**bold**` and escaped
# underscores cannot hide an out-of-repo qualifier.
MARKUP = re.compile(r"[*_`>|#>-]")


def normalise(line: str) -> str:
    flat = line.replace("\\_", "_").replace("\\`", "`")
    flat = MARKUP.sub(" ", flat)
    return re.sub(r"\s+", " ", flat).strip().lower()


def is_violation(line: str) -> bool:
    """True if the line tells a reader this repo runs the npm publisher."""
    if not TOKEN.search(line):
        return False
    return not any(marker in normalise(line) for marker in OUT_OF_REPO)


def scan(paths: list[pathlib.Path]) -> list[tuple[str, int, str]]:
    violations: list[tuple[str, int, str]] = []
    for path in paths:
        if not path.exists():
            continue
        for lineno, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
            if is_violation(line):
                violations.append((path.as_posix(), lineno, line.strip()))
    return violations


def selftest(extra_paths: list[str]) -> int:
    """Prove the PREDICATE fires on an instruction and spares a refutation.

    This asserts harness integrity only. The live tree's state is reported
    separately, because "the repo is currently unfixed" is a fact about the repo,
    not a harness bug - conflating them would make the selftest fail for reasons
    that have nothing to do with the classifier.
    """
    cases: list[tuple[str, bool, str]] = [
        ("| Node (npm OIDC)               | 24      | `bosun-publish.yaml`    | x |", True,
         "toolchain matrix row"),
        ("| `bosun-publish.yaml` | Publish bosun | push main |", True,
         "inventory table row"),
        ("- npm: `bosun-publish.yaml` uses npm provenance (Node 24+ required).", True,
         "OIDC usage bullet"),
        ("**Symptom:** `bosun-publish.yaml` fails on `npm publish` with auth error", True,
         "troubleshooting symptom"),
        ("gh workflow run bosun-publish.yaml", True, "bare command"),
        ("| Node (npm OIDC, bosun repo) | 24 | `virtengine/bosun` `publish.yaml` (NOT in this repo) |",
         False, "corrected matrix row"),
        ("> **SCOPE: this workflow is NOT in this repository.** published by "
         "`virtengine/bosun/.github/workflows/publish.yaml`", False, "blockquote refutation"),
        ("gh workflow run publish.yaml -R virtengine/bosun", False, "corrected command"),
        ("The repository has no npm publisher at all.", False, "bare denial"),
    ]

    ok = True
    for line, want, label in cases:
        got = is_violation(line)
        status = "ok" if got == want else "HARNESS BUG"
        if got != want:
            ok = False
        print(f"  [{status:11}] {label:26} want={want!s:5} got={got!s:5}")

    # Negative control: a planted instruction must be detected.
    with tempfile.TemporaryDirectory() as td:
        planted = pathlib.Path(td) / "AGENTS.md"
        planted.write_text(
            "# T\n\n| C | v | `bosun-publish.yaml` | n |\n", encoding="utf-8")
        n = len(scan([planted]))
        print(f"  [{'ok' if n == 1 else 'HARNESS BUG':11}] "
              f"{'planted instruction':26} want=1     got={n}")
        if n != 1:
            ok = False

    # A qualifying file must be clean, else the qualifier vocabulary is too narrow.
    if extra_paths:
        for path in extra_paths:
            got = scan([pathlib.Path(path)])
            print(f"  [{'ok' if not got else 'HARNESS BUG':11}] "
                  f"{pathlib.Path(path).name:26} want=0     got={len(got)}")
            if got:
                for f, i, l in got:
                    print(f"        {f}:{i}: {l[:100]}")
                ok = False

    print("SELFTEST", "PASS" if ok else "FAIL")
    return 0 if ok else 1


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("paths", nargs="*", help="markdown files to check")
    parser.add_argument("--selftest", action="store_true",
                        help="prove the predicate fires on both shapes; pass "
                             "corrected files to also assert they come out clean")
    args = parser.parse_args()

    if args.selftest:
        return selftest(args.paths)

    targets = [pathlib.Path(p) for p in args.paths] or [
        REPO_ROOT / p for p in DEFAULT_TARGETS]

    violations = scan(targets)
    if violations:
        print(f"RED: {len(violations)} unqualified `publish.yaml` reference(s):")
        for path, lineno, line in violations:
            print(f"  {path}:{lineno}: {line[:110]}")
        print("\nThis repo has no npm publisher. The real one is "
              "virtengine/bosun/.github/workflows/publish.yaml - qualify it as "
              "out-of-repo, do not list it as a local workflow.")
        return 1

    print(f"GREEN: 0 unqualified `publish.yaml` references in {len(targets)} file(s) "
          "(refutations correctly excluded)")
    return 0


if __name__ == "__main__":
    sys.exit(main())