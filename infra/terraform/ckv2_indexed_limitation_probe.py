#!/usr/bin/env python3
"""Mutation harness behind infra/terraform/CHECKOV_SCANER_LIMITATION.md.

The previous two versions of that document asserted things about checkov that
turned out to be false, and one of them was used to suppress a real finding.
The reason that was possible is that the claims were prose with no executable
backing: nothing failed when they stopped matching the tool.

This is that backing. It re-measures every claim the document makes and exits
non-zero if any of them stops holding, so the document cannot drift away from
checkov's behaviour without this going red.

Claims, and the mutation that falsifies each:

  T1  the unindexed verdict tracks the control
      -> deleting the public-access-block MUST flip PASS -> FAIL
  T2  the indexed verdict is blind to the control
      -> deleting the public-access-block MUST NOT change the failure set
  T3  the index is the cause
      -> removing `count` from the bucket AND its block MUST flip
         logs[0] FAILED -> PASSED

The control case matters most: a "0 failures" result is indistinguishable from a
scan that never ran, which is how a probe once reported a clean tree while
scanning nothing. So before the claims are tested, this asserts that a check
known to be red on the tree (`CKV_AWS_338`, log retention) actually reports
failures. If it does not, the harness exits 3 having reported nothing.

Usage:  python infra/terraform/ckv2_indexed_limitation_probe.py
Exit 0 = all claims hold. 1 = a claim was falsified. 2 = harness error.
        3 = control broken, nothing was measured.
Requires checkov importable (`pip install checkov`). If it is not, exit 2 --
never a silent pass.
"""
from __future__ import annotations

import json
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

REPO = Path(__file__).resolve().parents[2]
TF_DIR = REPO / "infra" / "terraform"
BLOCK = "aws_s3_bucket_public_access_block"

# Known-red on this tree, and deliberately left red. Used only as a liveness
# control, never as a claim.
CONTROL_CHECK = "CKV_AWS_338"


def scan(check: str = "CKV2_AWS_6", tf_dir: Path | None = None):
    """(passed_count, [failed_resource_names]) or None if checkov did not run.

    `tf_dir` MUST be passed by every caller that scans a mutated copy. It
    defaulted to the module-global TF_DIR, so the mutation harness scanned the
    ORIGINAL worktree and every "mutated" result was really the baseline. The
    probe reported FALSIFIED for T1/T3 on that run -- it was measuring the wrong
    tree, not refuting anything. A harness that cannot be pointed at a different
    tree cannot test a mutation at all.
    """
    proc = subprocess.run(
        [sys.executable, "-m", "checkov.main", "-d", str(tf_dir or TF_DIR),
         "--framework", "terraform", "--compact", "--check", check, "-o", "json"],
        capture_output=True, text=True, encoding="utf-8", errors="replace",
    )
    try:
        # checkov emits ONE JSON OBJECT per framework, not an array. A parser
        # that only handles [...] returns nothing and looks like a clean tree.
        payload = json.loads(proc.stdout)
    except Exception:
        return None
    if not isinstance(payload, dict):  # a list here is a different shape again
        return None
    results = payload.get("results", {})
    return (
        len(results.get("passed_checks", [])),
        [c.get("resource", "?") for c in results.get("failed_checks", [])],
    )


def short(names):
    return sorted({n.split(".")[-1] for n in names})


def mutated(mutator):
    """Copy the tree, mutate it, scan THE COPY. Never mutates the worktree."""
    holder = tempfile.TemporaryDirectory(dir=str(REPO.parent))
    try:
        tree = Path(holder.name) / "repo"
        shutil.copytree(TF_DIR, tree / TF_DIR.name)
        copy_tf = tree / TF_DIR.name
        mutator(copy_tf)
        return scan(tf_dir=copy_tf)
    finally:
        holder.cleanup()


def delete_block(rel: str, name: str):
    def apply(tf: Path):
        target = tf / rel
        text = target.read_text(encoding="utf-8")
        start = text.index(f'resource "{BLOCK}" "{name}"')
        end = text.index("\nresource ", start + 10)
        target.write_text(text[:start] + text[end + 1:], encoding="utf-8")
    return apply


def unindex_logs(tf: Path):
    """Remove `count` from the scaling logs bucket and its block, together."""
    target = tf / "modules/scaling/main.tf"
    text = target.read_text(encoding="utf-8")
    before = text
    text = text.replace(
        'resource "aws_s3_bucket" "logs" {\n'
        "  count = var.enable_global_accelerator ? 1 : 0\n",
        'resource "aws_s3_bucket" "logs" {\n', 1)
    text = text.replace(
        f'resource "{BLOCK}" "logs" {{\n'
        "  count = var.enable_global_accelerator ? 1 : 0\n"
        "\n  bucket = aws_s3_bucket.logs[0].id",
        f'resource "{BLOCK}" "logs" {{\n  bucket = aws_s3_bucket.logs.id', 1)
    if text == before:
        raise SystemExit("harness error: unindex_logs matched nothing -- "
                         "the scaling module changed shape")
    target.write_text(text, encoding="utf-8")


def main() -> int:
    control = scan(CONTROL_CHECK)
    if control is None:
        print(f"CONTROL: {CONTROL_CHECK} produced no JSON. checkov missing or failed.")
        print("Nothing has been measured.")
        return 2
    if not control[1]:
        print(f"CONTROL BROKEN: {CONTROL_CHECK} is known-red on this tree but reported "
              "0 failures. The scan is not measuring the real finding set; refusing to "
              "report a result from it.")
        return 3
    print(f"CONTROL {CONTROL_CHECK}: passed={control[0]} failed={len(control[1])} "
          f"-> scan is live\n")

    base = scan()
    if base is None:
        print("baseline scan produced no JSON")
        return 2
    print(f"BASELINE CKV2_AWS_6: passed={base[0]} failed={len(base[1])}")
    for name in short(base[1]):
        print(f"    FAIL {name}")
    print()

    verdicts: dict[str, bool] = {}

    t1 = mutated(delete_block("modules/s3/main.tf", "ml_models"))
    t1_ok = t1 is not None and "ml_models" in short(t1[1]) and "ml_models[0]" in short(t1[1])
    verdicts["T1"] = t1_ok
    print(f"T1  delete s3 ml_models block  -> passed={t1[0] if t1 else '?'} "
          f"failed={short(t1[1]) if t1 else '?'}")
    print(f"    {'OK' if t1_ok else 'FALSIFIED'}: unindexed flips PASS->FAIL, [0] holds")

    t2 = mutated(delete_block("modules/scaling/main.tf", "logs"))
    t2_ok = t2 is not None and sorted(t2[1]) == sorted(base[1])
    verdicts["T2a"] = t2_ok
    print(f"\nT2a delete scaling logs block -> passed={t2[0] if t2 else '?'} "
          f"failed={short(t2[1]) if t2 else '?'}")
    print(f"    {'OK' if t2_ok else 'FALSIFIED'}: failure set unchanged -> [0] ignores the control")

    t2b_ok = t1 is not None and "ml_models[0]" in short(t1[1])
    verdicts["T2b"] = t2b_ok
    print(f"    {'OK' if t2b_ok else 'FALSIFIED'}: [0] still blind in the second module")

    t3 = mutated(unindex_logs)
    t3_ok = t3 is not None and "logs" not in short(t3[1]) and "logs[0]" not in short(t3[1])
    verdicts["T3"] = t3_ok
    print(f"\nT3  remove count from logs pair -> passed={t3[0] if t3 else '?'} "
          f"failed={short(t3[1]) if t3 else '?'}")
    print(f"    {'OK' if t3_ok else 'FALSIFIED'}: unindexed logs PASSES -> the index is the cause")

    print()
    for name in ("T1", "T2a", "T2b", "T3"):
        print(f"  {name}: {'holds' if verdicts[name] else 'FALSIFIED'}")
    ok = all(verdicts.values())
    print("\n" + "=" * 66)
    print("All claims in CHECKOV_SCANER_LIMITATION.md still hold." if ok else
          "AT LEAST ONE CLAIM IS NOW FALSE. The document is stale and must be "
          "corrected before anyone relies on it.")
    return 0 if ok else 1


if __name__ == "__main__":
    raise SystemExit(main())