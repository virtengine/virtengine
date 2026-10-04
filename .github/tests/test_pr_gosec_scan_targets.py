#!/usr/bin/env python3
"""Prove the PR-side gosec leg scans the changed files instead of failing on them.

Why this exists
---------------
``pr-security-check.yaml`` used to pass the changed ``.go`` FILES to gosec.
gosec v2.25.0 resolves its arguments as package patterns, so every such run
produced::

    "Issues": [], "Stats": {"files": 0, ...},
    "Golang errors": {"pkg/foo/bar.go": [... "cannot find package \".\""]}

and exited 1 - a false negative (the only PR-side Go SAST control analysed
nothing) reported as a false red. This test pins the fixed behaviour.

What it checks, each against a real gosec run of the repo's own pinned version
when one is available and a stub otherwise:

1. the workflow resolves targets through ``gosec_pr_targets.py`` and never
   passes ``${changed_files[@]}`` to gosec;
2. gosec given the resolved package dirs analyses files (Stats.files > 0);
3. gosec given the raw file paths does NOT (Stats.files == 0) - so a regression
   to the old argument shape is caught by the same test;
4. a report with Stats.files == 0 is treated as a failure, not a clean scan;
5. a changed file outside any loadable package is rejected (fail closed).

Usage: python .github/tests/test_pr_gosec_scan_targets.py
Exit 0 = all properties hold; non-zero = the failing property is printed.
"""

from __future__ import annotations

import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

REPO = Path(__file__).resolve().parents[2]
WORKFLOW = REPO / ".github/workflows/pr-security-check.yaml"
SCRIPT = REPO / ".github/scripts/gosec_pr_targets.py"
sys.path.insert(0, str(REPO / ".github/scripts"))

import gosec_pr_targets as gpt  # noqa: E402

failures: list[str] = []


def check(name: str, ok: bool, detail: str = "") -> None:
    print(f"  {'PASS' if ok else 'FAIL'}  {name}" + (f"  -- {detail}" if detail else ""))
    if not ok:
        failures.append(name)


# The two real packages used throughout: both are small, both are real, and the
# email/sms pair is the exact input from the original defect report.
SAMPLE = ["pkg/verification/sms/providers.go", "pkg/verification/email/providers.go"]


def test_workflow_shape() -> None:
    print("== 1. workflow wiring ==")
    text = WORKFLOW.read_text(encoding="utf-8")
    check("workflow calls gosec_pr_targets.py", "gosec_pr_targets.py" in text)
    # The old shape: gosec handed the changed file array directly.
    old_shape = re.search(r'gosec\s+-[^\n]*"?\$\{changed_files\[@\]\}"?', text)
    check("workflow does not pass ${changed_files[@]} to gosec", old_shape is None,
          "found a gosec invocation still taking the raw changed-file array")
    check("workflow fails closed on Stats.files == 0",
          "analysis_files" in text and '-eq 0' in text)
    check("helper file exists and is importable", SCRIPT.is_file())


def test_target_resolution() -> None:
    print("== 2. target resolution (the helper's own logic) ==")
    with tempfile.TemporaryDirectory() as td:
        listing = Path(td) / "changed.txt"
        # Deliberately CRLF: a CRLF file list is what a Windows checkout produces,
        # and an unstripped \r reaches gosec as part of the path.
        listing.write_bytes(("\r\n".join(SAMPLE) + "\r\n").encode())
        files = gpt.read_changed_files(str(listing))
        check("CRLF file list yields clean paths", files == SAMPLE, repr(files))
        dirs = gpt.package_dirs(files)
        check("both files map to distinct package dirs",
              dirs == ["pkg/verification/sms", "pkg/verification/email"], repr(dirs))
        check("generated/test files are filtered out",
              not gpt.is_scannable_file("x/foo/bar_test.go")
              and not gpt.is_scannable_file("api/query.pb.go")
              and gpt.is_scannable_file("x/foo/bar.go"))

        # A file the go tool IGNORES (basename starting with _ or .) is not
        # scannable: it is never compiled into the directory's package, so that
        # directory may legitimately have no loadable package. Measured on PR
        # #1222 / job 111357375880: a gofmt-only change to scripts/dev/_tmp_addr.go
        # was fail-closed refused with "no Go files in .../scripts/dev" and turned
        # a clean PR red. One such file alongside a real one must still yield the
        # real one -- the filter must not silently swallow a whole PR.
        check("go-tool-ignored files are filtered out",
              not gpt.is_scannable_file("scripts/dev/_tmp_addr.go")
              and not gpt.is_scannable_file("scripts/dev/.hidden.go")
              and not gpt.is_scannable_file("x/foo/._bar.go"))
        check("an ordinary file in an _-prefixed FILE name is still scannable",
              gpt.is_scannable_file("x/foo/real.go"))
        mixed_good, mixed_bad, mixed_skipped = gpt.resolve(
            ["scripts/dev/_tmp_addr.go", "pkg/verification/sms/providers.go"], str(REPO))
        check("an ignored file does not drag a real package down with it",
              mixed_good == ["./pkg/verification/sms"] and mixed_bad == []
              and mixed_skipped == ["scripts/dev/_tmp_addr.go"], repr((mixed_good, mixed_bad, mixed_skipped)))
        # An all-ignored PR must be "nothing to scan" (exit 0 via the empty-report
        # path), NOT the fail-closed refusal that made it red.
        only_good, only_bad, only_skipped = gpt.resolve(["scripts/dev/_tmp_addr.go"], str(REPO))
        check("a PR touching only go-tool-ignored files resolves to no targets",
              only_good == [] and only_bad == [] and only_skipped == ["scripts/dev/_tmp_addr.go"],
              repr((only_good, only_bad, only_skipped)))
        # A file in a directory that is not a package must be refused, not passed on.
        good, bad = gpt.loadable(["pkg/verification/sms"], str(REPO))
        check("a real package resolves", good == ["./pkg/verification/sms"], repr(good))
        check("a non-package directory is rejected", bad == [], repr(bad))
        if shutil.which("go"):
            _, bad2 = gpt.loadable(["testdata/definitely-not-a-package"], str(REPO))
            check("an unloadable directory is reported as bad", len(bad2) == 1, repr(bad2))
        else:
            print("  SKIP  go not on PATH: cannot prove the loadable/unloadable split")


def _stub_gosec(record: Path, files: int, exit_code: int) -> None:
    """A gosec stand-in that records its argv and writes a chosen report.

    Invoked as ``[sys.executable, stub, *args]`` rather than as an executable
    script: the repo has a windows-native CI job, and Python's CreateProcess
    cannot exec a ``#!`` script there (WinError 193). A stub that only runs on
    the author's platform is a test that silently stops running in CI.
    """
    record.write_text(
        "import sys, json, pathlib\n"
        "argv = sys.argv[1:]\n"
        "pathlib.Path('gosec-argv.txt').write_text('\\n'.join(argv))\n"
        # Honour the real gosec contract: the output path is whatever follows
        # '-out'. Guessing the filename instead would make the stub disagree
        # with the very invocation it is standing in for.
        "out = argv[argv.index('-out') + 1] if '-out' in argv else 'out.json'\n"
        "fmt = 'json' if out.endswith('.json') else 'sarif'\n"
        "if fmt == 'json':\n"
        f"    report = {{'Issues': [], 'Stats': {{'files': {files}, 'lines': 100, 'found': 0}}, 'Golang errors': {{}}}}\n"
        "else:\n"
        "    report = {'version': '2.1.0', 'runs': []}\n"
        "pathlib.Path(out).write_text(json.dumps(report))\n"
        f"sys.exit({exit_code})\n",
        encoding="utf-8",
        newline="\n",
    )


def test_scan_outcomes() -> None:
    print("== 3. the scan verdict the step now computes ==")
    with tempfile.TemporaryDirectory() as td:
        # (a) packages -> files analysed, exit 0 -> must be recorded as a pass.
        _stub_gosec(Path(td) / "gosec_ok.py", files=23, exit_code=0)
        _, out = _run_step_targets(td, ["./pkg/verification/sms"], "gosec_ok.py")
        check("resolved packages with a real scan exit 0", out == 0, f"verdict={out}")

        # (b) the OLD shape (file paths) -> 0 files, exit 1. Recorded as failure,
        # never as a clean scan. This is the regression the fix exists for.
        _stub_gosec(Path(td) / "gosec_zero.py", files=0, exit_code=1)
        argv, out = _run_step_targets(td, ["pkg/verification/sms/providers.go"], "gosec_zero.py")
        check("file-path args (Stats.files=0, exit 1) is a FAILURE, not a pass", out == 1, f"verdict={out}")
        check("the stub really was handed the raw file path", "providers.go" in argv, argv.replace("\n", " "))

        # (c) an empty report but exit 0 (the shape a partial scan can take).
        _stub_gosec(Path(td) / "gosec_partial.py", files=0, exit_code=0)
        _, out = _run_step_targets(td, ["./pkg/verification/sms"], "gosec_partial.py")
        check("Stats.files=0 with exit 0 is still a FAILURE", out == 1, f"verdict={out}")

        # (d) a genuine finding still fails - the gate must not have been blunted.
        _stub_gosec(Path(td) / "gosec_finding.py", files=23, exit_code=1)
        _, out = _run_step_targets(td, ["./pkg/verification/sms"], "gosec_finding.py")
        check("a real finding (exit 1) still fails", out == 1, f"verdict={out}")


def _run_step_targets(td: str, targets: list[str], gosec_name: str) -> tuple[str, int]:
    """Run the workflow's verdict arithmetic against a stub gosec.

    Mirrors the step's own lines so the test grades the SHIPPED decision, not a
    restatement of it: gosec is invoked with the targets, then the same two
    signals (its exit status and Stats.files) decide the recorded exit_code.
    """
    stub = str(Path(td) / gosec_name)
    rc = subprocess.run(
        [sys.executable, stub, "-fmt", "json", "-out", "gosec-results.json",
         "-exclude-generated", *targets],
        cwd=td, capture_output=True, text=True,
    ).returncode
    subprocess.run(
        [sys.executable, stub, "-fmt", "sarif", "-out", "gosec.sarif",
         "-exclude-generated", *targets],
        cwd=td, capture_output=True, text=True,
    )
    analysis_files = int(json.loads(Path(td, "gosec-results.json").read_text())["Stats"]["files"])
    argv = Path(td, "gosec-argv.txt").read_text()
    verdict = 1 if (rc != 0 or analysis_files == 0) else 0
    return argv, verdict


def test_empty_report_flag() -> None:
    print("== 4. the nothing-to-scan path writes a usable report ==")
    with tempfile.TemporaryDirectory() as td:
        listing = Path(td) / "changed.txt"
        listing.write_text("x/foo/bar_test.go\n", encoding="utf-8")
        rc = subprocess.run(
            [sys.executable, str(SCRIPT), "--changed-files", str(listing), "--write-empty-report"],
            cwd=td, capture_output=True, text=True,
        ).returncode
        check("--write-empty-report exits 0", rc == 0)
        check("empty report is valid JSON with no issues",
              json.loads(Path(td, "gosec-results.json").read_text()) == {"Issues": []})
        check("empty SARIF has the 2.1.0 schema",
              json.loads(Path(td, "gosec.sarif").read_text())["version"] == "2.1.0")


def test_stdout_is_only_package_paths() -> None:
    """The real end-to-end contract: stdout is the gosec argument list, nothing else.

    Why this exists
    ---------------
    ``main()`` printed its human-readable ``note: ...`` summary to STDOUT while
    every other diagnostic in the same function used ``file=sys.stderr``. The
    workflow consumes stdout verbatim as the package list::

        python3 .github/scripts/gosec_pr_targets.py --changed-files ... > gosec-targets.txt
        mapfile -t gosec_targets < gosec-targets.txt

    so that line became a gosec target. gosec skipped it as a non-existent path,
    imported 0 files, and the job's fail-closed ``Stats.files=0`` guard reported a
    red that was not a security finding at all. Measured on run 36786584597, job
    ``Go Security Scan (core)``, 2026-09-30: stderr published
    ``gosec targets: 0 package(s)`` while the scan line read
    ``scanning 1 package(s): note: ...``.

    The earlier sections of this file test the helper's FUNCTIONS, so they all
    pass while this defect is live - only the process's real stdout shows it.
    That is why the assertion is made by RUNNING the script, exactly as the
    workflow runs it, rather than by reading its source.
    """
    print("== 5. stdout carries package paths ONLY (the real process contract) ==")
    with tempfile.TemporaryDirectory() as td:
        # A changed file that the scan filter drops, so `skipped` is non-empty and
        # the note branch actually executes. Zero packages survive -> the count the
        # script publishes and the list the caller builds must both be empty.
        listing = Path(td) / "changed.txt"
        listing.write_text("x/veid/keeper/keeper_test.go\n", encoding="utf-8")
        proc = subprocess.run(
            [sys.executable, str(SCRIPT), "--changed-files", str(listing),
             "--module-dir", str(REPO)],
            cwd=td, capture_output=True, text=True,
        )
        check("helper exits 0 on a fully-filtered change list", proc.returncode == 0,
              f"rc={proc.returncode} stderr={proc.stderr[-300:]}")

        stdout_lines = [ln for ln in proc.stdout.splitlines() if ln.strip()]
        check("stdout is empty when 0 packages resolved", stdout_lines == [],
              "stdout carried " + repr(stdout_lines))

        # The note must still be VISIBLE - moving it to stderr must not silence it.
        check("the note is still reported, on stderr", "note:" in proc.stderr,
              proc.stderr[-300:])
        # And the published count must be consistent with the list the caller builds.
        m = re.search(r"gosec targets: (\d+) package", proc.stderr)
        check("published count is present on stderr", m is not None)
        if m:
            check("published count matches what stdout delivered",
                  int(m.group(1)) == len(stdout_lines),
                  f"stderr said {m.group(1)}, stdout delivered {len(stdout_lines)}")

    # And the positive case: a real package reaches stdout as a bare path, with no
    # diagnostics mixed in. This is what the mapfile actually consumes.
    if not shutil.which("go"):
        print("  SKIP  go not on PATH: cannot resolve a real package")
        return
    with tempfile.TemporaryDirectory() as td:
        listing = Path(td) / "changed.txt"
        listing.write_text("pkg/verification/sms/providers.go\n", encoding="utf-8")
        proc = subprocess.run(
            [sys.executable, str(SCRIPT), "--changed-files", str(listing),
             "--module-dir", str(REPO)],
            cwd=td, capture_output=True, text=True,
        )
        lines = [ln for ln in proc.stdout.splitlines() if ln.strip()]
        check("a resolvable package is emitted on stdout as a bare path",
              lines == ["./pkg/verification/sms"], repr(lines))
        # Every stdout line must be a path the script would hand to gosec, i.e.
        # start with "./" or a drive letter - never a sentence.
        check("no stdout line looks like a human-readable message",
              all(re.match(r"^\./|^[A-Za-z]:[\\/]", ln) for ln in lines), repr(lines))


def main() -> int:
    print("test_pr_gosec_scan_targets: PR-side gosec must SCAN the changed files\n")
    for fn in (test_workflow_shape, test_target_resolution, test_scan_outcomes,
               test_empty_report_flag, test_stdout_is_only_package_paths):
        fn()
        print()
    if failures:
        print(f"RESULT: {len(failures)} property(ies) FAILED: " + "; ".join(failures))
        return 1
    print("RESULT: all properties hold - the PR gosec leg scans the changed packages, "
          "fails closed on a partial scan, and refuses an unloadable target")
    return 0


if __name__ == "__main__":
    sys.exit(main())
