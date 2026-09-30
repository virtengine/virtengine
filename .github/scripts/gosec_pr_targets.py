#!/usr/bin/env python3
"""Resolve a PR's changed Go files into package directories gosec can actually load.

Why this exists
---------------
``pr-security-check.yaml`` scopes gosec to the files a PR changed rather than to the
whole tree, which is the right call for PR latency. It passed those files as
*arguments*::

    gosec -fmt json -out gosec-results.json -exclude-generated "${changed_files[@]}"

gosec v2.25.0 resolves its arguments as **package patterns** through
``golang.org/x/tools/go/packages``. Handed a ``.go`` file it reports, per argument::

    "Golang errors": {
      "pkg/verification/email/providers.go": [
        {"line": 0, "column": 0,
         "error": "importing dir \"pkg/verification/email/providers.go\":
                   cannot find package \".\" in: ..."}
      ]
    },
    "Issues": [], "Stats": {"files": 0, "lines": 0, "nosec": 0, "found": 0}

i.e. it imports zero files and still **exits 1**. Two defects, opposite in
direction, from one argument shape:

* a **false negative** - the only PR-side Go SAST control analysed nothing on
  every run where a Go file changed;
* a **false red** - the non-zero exit was reported to the reviewer as a security
  finding when no scan had taken place.

This script produces the argument list gosec can load (package directories, one
entry per analysable package) and rejects any changed file whose directory is not
a loadable package, so a non-package file is reported rather than silently handed
to gosec as a pattern that cannot resolve.

Usage
-----
    gosec_pr_targets.py --changed-files changed.txt
    # prints one package directory per line; exits non-zero on an unloadable target
"""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
from pathlib import Path

# Directories that never contain a loadable package of this module. A changed file
# under one of them is reported as unresolvable rather than passed to gosec.
NEVER_A_PACKAGE_DIRS = {"testdata", "vendor", "node_modules", ".git"}


def norm(p: str) -> str:
    """Normalise separators so the same path compares equal on Windows and Linux."""
    return p.replace("\\", "/")


def read_changed_files(path: str) -> list[str]:
    text = Path(path).read_text(encoding="utf-8")
    # Strip a trailing CR as well as the newline. The workflow feeds this helper
    # from `git diff --name-only`, and a file list produced on a CRLF checkout
    # (Windows, or this repo's core.autocrlf=true) carries \r. A \r left on a
    # path is not caught by the `go list` check - that check only proves the
    # DIRECTORY resolves - so it survives into the gosec argument, and gosec then
    # reports "importing dir \"./pkg/foo\r\": cannot find package" while exiting 1:
    # exactly the false red being fixed, reintroduced one character wide. Verified:
    # the \r variant of this defect reproduced the 0-files/2-errors report while
    # the clean variant analysed 23 files.
    return [norm(line.strip().strip("\r\n")) for line in text.splitlines() if line.strip()]


def package_dirs(files: list[str]) -> list[str]:
    """Distinct parent directories of the changed Go files, in first-seen order."""
    seen: list[str] = []
    for f in files:
        d = f.rsplit("/", 1)[0] if "/" in f else "."
        if d not in seen:
            seen.append(d)
    return seen


def loadable(dirs: list[str], module_dir: str) -> tuple[list[str], list[tuple[str, str]]]:
    """Split dirs into loadable packages and dirs the go tool cannot load.

    One ``go list -e`` call for all of them: a directory that is a real package
    comes back with a ``.Name``; one that is not (a stray file, a fixture
    directory, a nested module) comes back with an empty name and an error.
    """
    if not dirs:
        return [], []

    # A directory must be handed to the go tool as a pattern with a leading ./ :
    # `go list pkg/foo` is not a pattern, it is a module-relative import path and
    # resolves to nothing (or to another module's package). gosec is given
    # ./pkg/foo, so the enumeration must use the same shape.
    patterns = [d if d.startswith(".") else f"./{d}" for d in dirs]

    env = dict(os.environ)
    env["GOFLAGS"] = "-mod=readonly"
    proc = subprocess.run(
        # `{{with .Error}}` (not `{{.Error.Err}}`) so a nil *load.PackageError does
        # not abort the whole template with "nil pointer evaluating
        # *load.PackageError.Err" and blank out every other field.
        ["go", "list", "-e", "-f", "{{.Name}}|{{.Dir}}|{{with .Error}}{{.Err}}{{end}}", *patterns],
        cwd=module_dir,
        capture_output=True,
        text=True,
        env=env,
    )
    if proc.returncode != 0:
        # `go list` itself failed (missing go, broken go.mod): report and let the
        # caller fail closed rather than assuming the targets are fine.
        err = (proc.stderr or proc.stdout).strip()
        print(f"::error::'go list' could not enumerate the changed packages: {err}", file=sys.stderr)
        return [], [(d, err or "go list failed") for d in dirs]

    good: list[str] = []
    bad: list[tuple[str, str]] = []
    by_dir: dict[str, tuple[str, str]] = {}
    for line in proc.stdout.splitlines():
        name, _, rest = line.partition("|")
        absdir, _, err = rest.partition("|")
        absdir = norm(absdir.strip())
        if not absdir:
            continue
        # `go list` reports the ABSOLUTE package dir while the caller holds the
        # module-relative one, so index on the module-relative suffix as well or
        # every lookup misses and every valid package is reported unloadable.
        entry = (name.strip(), err.strip())
        by_dir[absdir] = entry
        by_dir[absdir.rsplit("/", 1)[-1]] = entry
        parent = absdir.rsplit("/", 1)[0]
        if parent:
            by_dir[parent + "/" + absdir.rsplit("/", 1)[-1]] = entry

    for d in dirs:
        key = norm(d)
        entry = by_dir.get(key) or by_dir.get(key.rsplit("/", 1)[-1])
        if entry is None:
            bad.append((d, "go list returned no result for this path"))
        elif not entry[0]:
            bad.append((d, entry[1] or "not a loadable Go package"))
        else:
            good.append(self_dir(key))
    return good, bad


def self_dir(d: str) -> str:
    """Return the module-relative pattern to hand to gosec."""
    return d if d.startswith(".") else f"./{d}"


def is_scannable_file(f: str) -> bool:
    """False for files gosec must not be asked to cover (mirrors the workflow's awk)."""
    parts = norm(f).split("/")
    base = parts[-1]
    if base.endswith("_test.go") or base.endswith(".pb.go"):
        return False
    return not any(p in NEVER_A_PACKAGE_DIRS for p in parts[:-1])


def resolve(changed: list[str], module_dir: str) -> tuple[list[str], list[tuple[str, str]], list[str]]:
    scannable = [f for f in changed if is_scannable_file(f)]
    skipped = [f for f in changed if not is_scannable_file(f)]
    good, bad = loadable(package_dirs(scannable), module_dir)
    return good, bad, skipped


def summarise(report: str) -> dict:
    """Read Stats/Golang errors out of a gosec JSON report.

    ``files == 0`` with a non-empty ``Golang errors`` is the signature of a scan
    that could not load what it was given, which must never be reported as a
    clean scan (and never as a finding).
    """
    data = json.loads(Path(report).read_text(encoding="utf-8"))
    stats = data.get("Stats") or {}
    issues = data.get("Issues") or []
    golang_errors = data.get("Golang errors") or {}
    return {
        "files": int(stats.get("files", 0) or 0),
        "lines": int(stats.get("lines", 0) or 0),
        "found": int(stats.get("found", 0) or 0),
        "issues": len(issues),
        "golang_errors": golang_errors,
    }


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--changed-files", required=True,
                    help="newline-delimited changed Go files (git diff --name-only)")
    ap.add_argument("--module-dir", default=".", help="module root to resolve packages in")
    ap.add_argument("--write-empty-report", action="store_true",
                    help="write an empty gosec JSON + SARIF report and exit 0 (nothing to scan)")
    args = ap.parse_args()

    if args.write_empty_report:
        # Written here rather than as a heredoc in the workflow: a heredoc inside a
        # YAML block scalar must have its terminator at column 0, which the step's
        # indentation would break (actionlint/shellcheck SC1073 caught exactly that).
        Path("gosec-results.json").write_text(
            json.dumps({"Issues": []}), encoding="utf-8")
        Path("gosec.sarif").write_text(json.dumps({
            "$schema": "https://json.schemastore.org/sarif-2.1.0.json",
            "version": "2.1.0",
            "runs": [{"tool": {"driver": {"name": "gosec",
                                         "informationUri": "https://github.com/securego/gosec",
                                         "rules": []}},
                      "results": []}],
        }), encoding="utf-8")
        print("no scannable Go packages among the changed files; wrote an empty report", file=sys.stderr)
        return 0

    changed = read_changed_files(args.changed_files)
    good, bad, skipped = resolve(changed, args.module_dir)

    for d, err in bad:
        print(f"::error::changed Go file(s) in '{d}' are not in a loadable package: {err}",
              file=sys.stderr)
    if skipped:
        print(f"note: {len(skipped)} changed file(s) excluded by the scan filter "
              f"(tests, generated .pb.go, vendor/testdata), e.g. {skipped[0]}")

    if bad:
        print("::error::refusing to hand gosec a path it cannot load as a package - that "
              "makes it import 0 files and exit 1, which is a false red and a false "
              "negative at the same time. Fix the paths above.", file=sys.stderr)
        return 1

    for d in good:
        print(d)
    print(f"gosec targets: {len(good)} package(s) from {len(changed)} changed file(s) "
          f"({len(skipped)} filtered out)", file=sys.stderr)
    return 0


if __name__ == "__main__":
    sys.exit(main())
