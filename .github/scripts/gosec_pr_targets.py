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

# Subtrees that are SEPARATE Go modules (they carry their own go.mod). `go list`
# run from a module root cannot resolve a package that lives inside a nested
# module - relative or absolute - and reports instead:
#     main module (github.com/virtengine/virtengine) does not contain package
#     github.com/virtengine/virtengine/sdk/go/node/settlement/v1
# This is not a `go list` bug and no flag fixes it: the nested module has its own
# build list, so the parent's resolver legitimately refuses to look inside it.
# Measured against the real tree, 2026-10-02, on a minimal fixture with a root
# module plus a nested `sdk/go/go.mod`: the relative pattern AND the absolute path
# both return that message from the parent root, while the same pattern run with
# cwd=sdk/go returns `v1|<abs>/sdk/go/node/settlement/v1`.
#
# So resolution has to be per-module: a changed file is resolved against the
# nearest ancestor go.mod, not against one module root.
NESTED_MODULE_DIRS = ("sdk/go", "sdk/generation")


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


def module_root_for(dirpath: str, base: str = ".") -> str:
    """Return the module root that OWNS ``dirpath``.

    Resolution walks the path from the repository root and returns the deepest
    prefix that has its own ``go.mod`` (falling back to the repo root when there
    is none). Deeper wins deliberately: a package under ``sdk/go`` belongs to the
    nested module, not to the root module that also happens to contain it.
    """
    parts = [p for p in norm(dirpath).split("/") if p and p != "."]
    for depth in range(len(parts), -1, -1):
        candidate = "/".join(parts[:depth])
        root = os.path.join(base, *parts[:depth]) if candidate else base
        if os.path.isfile(os.path.join(root, "go.mod")):
            return root
    return base


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

    One ``go list -e`` call per OWNING MODULE ROOT, not one per repository: the
    changed packages of a PR routinely span several modules (the root module and
    ``sdk/go``), and a single call from the repository root cannot see into the
    nested ones. A directory that is a real package comes back with a ``.Name``;
    one that is not (a stray file, a fixture directory, a nested module passed
    at the wrong root) comes back with an empty name and an error.
    """
    if not dirs:
        return [], []

    # Group by the module that owns each dir, preserving first-seen order both
    # across the groups and inside them, so the emitted list stays stable.
    grouped: dict[str, list[str]] = {}
    for d in dirs:
        root = norm(module_root_for(d, module_dir))
        grouped.setdefault(root, []).append(d)

    good: list[str] = []
    bad: list[tuple[str, str]] = []
    for root, root_dirs in grouped.items():
        _loadable_in(root, root_dirs, module_dir, good, bad)
    return good, bad


def _loadable_in(root: str, dirs: list[str], module_dir: str,
                 good: list[str], bad: list[tuple[str, str]]) -> None:
    """Append the resolved packages of ``dirs`` (all inside ``root``) to good/bad."""
    # A directory must be handed to the go tool as a pattern with a leading ./ :
    # `go list pkg/foo` is not a pattern, it is a module-relative import path and
    # resolves to nothing (or to another module's package). gosec is given
    # ./pkg/foo, so the enumeration must use the same shape. The pattern is made
    # relative to the module root that owns it, which is what makes a nested
    # module's packages resolvable at all.
    rel_root = norm(os.path.relpath(root, module_dir)).replace("\\", "/")
    if rel_root == ".":
        rel_root = ""

    patterns = []
    for d in dirs:
        key = norm(d)
        if rel_root and (key == rel_root or key.startswith(rel_root + "/")):
            inner = key[len(rel_root):].lstrip("/")
            patterns.append("./" + inner if inner else ".")
        else:
            patterns.append(key if key.startswith(".") else f"./{key}")

    env = dict(os.environ)
    env["GOFLAGS"] = "-mod=readonly"
    proc = subprocess.run(
        # `{{with .Error}}` (not `{{.Error.Err}}`) so a nil *load.PackageError does
        # not abort the whole template with "nil pointer evaluating
        # *load.PackageError.Err" and blank out every other field.
        ["go", "list", "-e", "-f", "{{.Name}}|{{.Dir}}|{{with .Error}}{{.Err}}{{end}}", *patterns],
        cwd=root,
        capture_output=True,
        text=True,
        env=env,
    )
    if proc.returncode != 0:
        # `go list` itself failed (missing go, broken go.mod): report and let the
        # caller fail closed rather than assuming the targets are fine.
        err = (proc.stderr or proc.stdout).strip()
        print(f"::error::'go list' could not enumerate the changed packages in {root}: {err}",
              file=sys.stderr)
        bad.extend((d, err or "go list failed") for d in dirs)
        return

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
            bad.append((d, f"go list returned no result for this path (resolved in module {rel_root or '.'})"))
        elif not entry[0]:
            bad.append((d, entry[1] or "not a loadable Go package"))
        else:
            good.append(self_dir(key))


def self_dir(d: str) -> str:
    """Return the module-relative pattern to hand to gosec."""
    return d if d.startswith(".") else f"./{d}"


def is_scannable_file(f: str) -> bool:
    """False for files gosec must not be asked to cover (mirrors the workflow's awk).

    A file the GO TOOLCHAIN IGNORES is not a scannable target either, and asking
    for one turns a clean PR into a false red. The toolchain ignores any basename
    beginning with ``_`` or ``.`` (``go help build``: files whose names begin with
    ``_`` or ``.`` are ignored by the go tool) -- such a file is a standalone
    ``go run`` scratch program, never compiled into the directory's package, so
    its directory legitimately has no loadable package. ``git diff --name-only
    '*.go'`` does match it, and it is exactly the shape that produced a
    fail-closed refusal on a run that had introduced no security finding at all.

    Measured on PR #1222, run 37175539776, job 111357375880 (2026-10-04), where a
    gofmt-only change to ``scripts/dev/_tmp_addr.go`` (this repo's whole-tree
    gofmt debt, paid off in that PR) was refused by the gate above with::

        ::error::changed Go file(s) in 'scripts/dev' are not in a loadable
        package: no Go files in .../scripts/dev

    ``scripts/dev`` holds no loadable package for a real reason -- ``_tmp_addr.go``
    is the only .go file there and the go tool ignores it. Refusing was CORRECT
    behaviour for the input it was handed, but the input should never have been
    handed to it: gosec cannot scan a file the compiler never sees, so the honest
    verdict is "nothing to scan here", not "refusing, and red".
    """
    parts = norm(f).split("/")
    base = parts[-1]
    # The go tool ignores _foo.go / .foo.go / ..foo.go outright, so they are
    # neither compiled nor scannable.
    if base.startswith(("_", ".")):
        return False
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
        # MUST go to stderr. The workflow consumes stdout as the gosec argument list
        # (`mapfile -t gosec_targets < gosec-targets.txt`), so a human-readable line
        # here is read as a package path: gosec is handed "note: ...", skips it as a
        # non-existent path, imports 0 files, exits 0, and the job's fail-closed
        # `Stats.files=0` guard then reports a red that is not a security finding.
        # Measured on run 36786584597, job 'Go Security Scan (core)', 2026-09-30:
        # stderr said "gosec targets: 0 package(s)" while the scan line read
        # "scanning 1 package(s): note: ..." - the count that stderr published was
        # correct and the array the caller built from stdout was not.
        print(f"note: {len(skipped)} changed file(s) excluded by the scan filter "
              f"(tests, generated .pb.go, vendor/testdata), e.g. {skipped[0]}",
              file=sys.stderr)

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
