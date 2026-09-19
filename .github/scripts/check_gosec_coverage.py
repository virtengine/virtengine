#!/usr/bin/env python3
"""Fail a gosec run that silently skipped packages.

Why this exists
---------------
gosec resolves packages through ``golang.org/x/tools/go/packages``, which shells
out to the ``go`` command. When a package cannot be loaded (an untidy ``go.mod``
with the default ``-mod=readonly``, a module boundary, a type error) the loader
returns a ``*packages.Package`` with an empty ``Name`` and the error recorded in
``pkg.Errors``. gosec v2.25.0 then does, in ``analyzer.go``::

    for _, pkg := range pkgs {
        if pkg.Name == "" {
            continue          // <-- silent: no log line, no metric, no error
        }

so the dropped package never reaches the JSON report, ``Golang errors`` stays
empty, and the job still exits 0. That is how the ``gosec Security Scan`` gate
came to analyse only ~582 of the repository's ~3367 Go files while reporting
success.

This script reconstructs the package set from the ``go`` tool itself and
compares it against the files gosec actually looked at (the ``Checking file:``
lines it writes to stderr). Any package that produced no analysed file - unless
it is covered by an explicit ``--exclude-dir`` or every one of its files was
skipped as generated - is reported and fails the build.

Usage
-----
    check_gosec_coverage.py --log gosec.log --module-dir . \
        --exclude-dir vendor --exclude-dir testutil
"""

from __future__ import annotations

import argparse
import os
import re
import subprocess
import sys
from pathlib import Path

IMPORT_RE = re.compile(r"Import directory: (.+)$")
CHECKED_RE = re.compile(r"Checking file: (.+)$")
GENERATED_RE = re.compile(r"Ignoring generated file: (.+)$")


def norm(p: str) -> str:
    """Normalise a path for comparison on any host."""
    return p.replace("\\", "/").rstrip("/")


def parse_log(text: str) -> tuple[set[str], set[str], set[str]]:
    imported: set[str] = set()
    checked: set[str] = set()
    generated: set[str] = set()
    for line in text.splitlines():
        if m := IMPORT_RE.search(line):
            imported.add(norm(m.group(1)))
        elif m := CHECKED_RE.search(line):
            checked.add(norm(m.group(1)))
        elif m := GENERATED_RE.search(line):
            generated.add(norm(m.group(1)))
    return imported, checked, generated


def list_packages(module_dir: str, goflags: str) -> tuple[list[tuple[str, list[str]]], str]:
    """Return [(abs_dir, [go file names])] for the module, and any load error."""
    env = dict(os.environ)
    env["GOFLAGS"] = goflags
    proc = subprocess.run(
        ["go", "list", "-f", "{{.Dir}}|{{join .GoFiles \" \"}}", "./..."],
        cwd=module_dir,
        capture_output=True,
        text=True,
        env=env,
    )
    if proc.returncode != 0:
        return [], (proc.stderr or proc.stdout).strip()
    pkgs: list[tuple[str, list[str]]] = []
    for line in proc.stdout.strip().splitlines():
        d, _, gofiles = line.partition("|")
        pkgs.append((norm(d), [f for f in gofiles.split() if f]))
    return pkgs, ""


def under_excluded(path: str, module_dir: str, excluded: list[str]) -> bool:
    rel = path[len(module_dir) + 1 :] if path.startswith(module_dir + "/") else path
    parts = rel.split("/")
    return any(name in parts for name in excluded)


def find_dropped(
    pkgs: list[tuple[str, list[str]]],
    module_dir: str,
    excluded: list[str],
    checked: set[str],
    generated: set[str],
) -> list[tuple[str, list[str]]]:
    """Packages that produced no analysed file despite having files to analyse.

    A package is *not* dropped when it is covered by an explicit --exclude-dir,
    or when every one of its files was skipped as generated.
    """
    checked_dirs = {f.rsplit("/", 1)[0] for f in checked}
    dropped: list[tuple[str, list[str]]] = []
    for d, gofiles in pkgs:
        if under_excluded(d, module_dir, excluded):
            continue
        files = [f"{d}/{f}" for f in gofiles] if d else list(gofiles)
        outstanding = [f for f in files if f not in checked and f not in generated]
        if outstanding and d not in checked_dirs:
            dropped.append((d, outstanding))
    return dropped


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--log", required=True, help="gosec stderr log (contains 'Checking file:' lines)")
    ap.add_argument("--module-dir", default=".", help="module root to enumerate with 'go list'")
    ap.add_argument("--exclude-dir", action="append", default=[],
                    help="directory name gosec was told to skip (repeatable)")
    ap.add_argument("--goflags", default="-mod=readonly",
                    help="GOFLAGS used for the package enumeration (must match CI)")
    args = ap.parse_args()

    module_dir = norm(str(Path(args.module_dir).resolve()))
    excluded = [norm(e) for e in args.exclude_dir]

    log_path = Path(args.log)
    if not log_path.is_file():
        print(f"::error::gosec log not found: {log_path}", file=sys.stderr)
        return 1
    imported, checked, generated = parse_log(log_path.read_text(encoding="utf-8", errors="replace"))
    if not checked:
        print("::error::no 'Checking file:' lines in the gosec log - the scan produced nothing",
              file=sys.stderr)
        return 1

    pkgs, load_error = list_packages(module_dir, args.goflags)
    if load_error:
        print("::error::the go tool cannot enumerate this module's packages, so the "
              "scan would silently skip them:", file=sys.stderr)
        print(load_error, file=sys.stderr)
        print("::error::run 'go mod tidy' and commit go.mod/go.sum, then re-run", file=sys.stderr)
        return 1

    checked_dirs = {f.rsplit("/", 1)[0] for f in checked}
    dropped = find_dropped(pkgs, module_dir, excluded, checked, generated)

    rel = lambda p: p[len(module_dir) + 1 :] if p.startswith(module_dir + "/") else p
    print(f"gosec coverage: {len(checked)} files analysed across "
          f"{len(checked_dirs)} directories; {len(pkgs)} packages enumerated "
          f"({', '.join(excluded) or 'no'} excluded)")

    explicitly_imported_but_dropped = sorted(
        d for d in imported if d not in checked_dirs and not under_excluded(d, module_dir, excluded)
    )
    if explicitly_imported_but_dropped and not dropped:
        # gosec's directory walk descends into things that are not analysable
        # packages of this module: nested modules, testdata/fixtures, directories
        # with only generated files, and files excluded by build tags for this
        # GOOS. None of them is a package whose files were silently skipped.
        print(f"note: {len(explicitly_imported_but_dropped)} directory(ies) gosec walked produced no "
              f"analysed file and are not analysable packages of this module "
              f"(nested module, testdata/fixture, generated-only, or build-constrained): "
              + ", ".join(rel(d) for d in explicitly_imported_but_dropped[:8])
              + (" ..." if len(explicitly_imported_but_dropped) > 8 else ""))

    if dropped:
        print(f"::error::{len(dropped)} package(s) were not analysed by gosec but the "
              f"job would otherwise pass:", file=sys.stderr)
        for d, files in dropped:
            print(f"  {rel(d)}  ({len(files)} file(s) unscanned, e.g. {rel(files[0])})",
                  file=sys.stderr)
        print("::error::gosec drops packages whose type-check fails without logging it. "
              "Fix the load error (usually 'go mod tidy') or add a documented -exclude-dir "
              "with a reason - do not let the gate pass on a partial scan.", file=sys.stderr)
        return 1

    print("OK: every package in this module was analysed (or is explicitly excluded)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
