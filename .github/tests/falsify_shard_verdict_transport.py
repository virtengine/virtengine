#!/usr/bin/env python3
"""Throwaway: prove the transport tests catch a regression to either bad design.

Each mutation restores one of the two transports that FAILED in production, and
the suite must go RED for it.
"""
import pathlib
import subprocess
import sys

ROOT = pathlib.Path(__file__).resolve().parent.parent.parent
WF = ROOT / ".github" / "workflows" / "security.yaml"
BAK = ROOT / ".probe_wf_backup.yaml"

MUTATIONS = [
    (
        "M1-gh-run-download-returns",
        '          ls -1 "$RUNNER_TEMP/verdicts"',
        '          gh run download "$GITHUB_RUN_ID" --name x --dir y\n'
        '          ls -1 "$RUNNER_TEMP/verdicts"',
    ),
    (
        "M2-matrix-output-transport-returns",
        "      - name: Publish this shard's verdict",
        "      - name: Expose shard verdict as an aggregate output\n"
        "        id: shard-verdict-agg\n"
        "        if: always()\n"
        "        run: echo \"agg={}\" >> \"$GITHUB_OUTPUT\"\n"
        "\n"
        "      - name: Publish this shard's verdict",
    ),
    (
        # The regression CI actually hit: `if-no-files-found` is an
        # upload-artifact input, and download-artifact@v6 does not define it.
        "M3-bogus-input-on-download-step",
        "          pattern: govulncheck-verdict-*",
        "          pattern: govulncheck-verdict-*\n"
        "          if-no-files-found: error",
    ),
    (
        "M4-per-shard-artifact-removed",
        "          name: govulncheck-verdict-${{ matrix.shard }}",
        "          name: govulncheck-verdict-shared",
    ),
]


def run() -> int:
    return subprocess.run(
        [sys.executable, "-m", "unittest", "discover", "-s", ".github/tests",
         "-p", "test_shard_verdict_transport*.py"],
        cwd=ROOT, capture_output=True, text=True,
    ).returncode


def main() -> int:
    original = WF.read_bytes()
    BAK.write_bytes(original)
    print("== control: unmutated suite must PASS ==")
    rc = run()
    print(f"  control rc={rc} (expect 0)")
    if rc != 0:
        WF.write_bytes(original); BAK.unlink(); return 1

    caught = 0
    try:
        for name, old, new in MUTATIONS:
            text = original.decode("utf-8")
            if old not in text:
                print(f"  INERT {name}: anchor not found")
                continue
            WF.write_text(text.replace(old, new, 1), encoding="utf-8", newline="")
            rc = run()
            ok = rc != 0
            caught += 1 if ok else 0
            print(f"  {'CAUGHT' if ok else '*** INERT - DECORATION ***'}  {name} (rc={rc})")
            WF.write_bytes(original)
    finally:
        WF.write_bytes(original)
        if BAK.exists():
            BAK.unlink()

    same = WF.read_bytes() == original
    print(f"\n  workflow restored byte-identical: {same}")
    print(f"RESULT: {caught}/{len(MUTATIONS)} mutations CAUGHT")
    return 0 if (caught == len(MUTATIONS) and same) else 1


if __name__ == "__main__":
    sys.exit(main())