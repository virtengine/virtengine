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
      -> deleting the public-access-block MUST move that bucket PASS -> FAIL
  T2  the indexed verdict is blind to the control
      -> deleting the public-access-block MUST NOT change the failure set
  T3  the index is the cause
      -> removing `count` from the bucket AND its block MUST move that bucket
         from FAILED (as `logs[0]`) to PASSED (as `logs`)

Liveness, which is the property every claim above rests on:

  a scan that returned nothing and a scan of a tree with nothing in it produce
  the same JSON. So EVERY scan here -- baseline and mutated alike -- must
  (a) parse, (b) report a non-zero resource count, (c) report the control check
  CKV_AWS_338 as FAILED, which is known-red on this tree, and, for a MUTATED
  scan, (d) still see at least half the baseline's resources, because a broken
  tree does not necessarily yield 0 -- checkov will happily report the files
  that still parse. A scan that cannot be shown live is REFUSED (exit 3) rather
  than read as a verdict.

That is why T3 asserts MOVEMENT rather than absence. "`logs` is absent from the
failure list" is satisfied both by "it moved to PASSED" and by "it stopped being
scanned at all", and the second reading is the exact failure this harness exists
to catch. So T3 requires the unindexed name to be found in the PASSED list. No
verdict anywhere in this file is inferred from the absence of a failure.

Usage:  python infra/terraform/ckv2_indexed_limitation_probe.py
Exit 0 = all claims hold. 1 = a claim was falsified. 2 = the harness itself
        failed: checkov produced nothing, or a mutation could not be applied.
        3 = a scan was not live, nothing was measured from it.
Only 0 and 1 are about the document. 2 and 3 are about the probe, and a mutator
that loses its grip on the terraform must never be able to exit 1 -- that would
announce the document stale when only the harness broke.

Requires checkov importable (`pip install checkov`). If it is not, exit 2 --
never a silent pass.

Deliberately NOT wired into CI -- see "Running it" in
CHECKOV_SCANER_LIMITATION.md for the measured cost and the decision.
"""
from __future__ import annotations

import json
import shutil
import subprocess
import sys
import tempfile
from dataclasses import dataclass, field
from pathlib import Path

REPO = Path(__file__).resolve().parents[2]
TF_DIR = REPO / "infra" / "terraform"
BLOCK = "aws_s3_bucket_public_access_block"

# Known-red on this tree, and deliberately left red. Used only as a liveness
# control, never as a claim.
CONTROL_CHECK = "CKV_AWS_338"

# The check whose behaviour the document makes claims about.
CLAIM_CHECK = "CKV2_AWS_6"

# Resource identities the claims are about, spelled in full. Shortened forms
# cannot tell `logs` from `logs[0]`, which is precisely the distinction T3 turns
# on, so nothing below compares a truncated name.
S3_ML = "module.s3.aws_s3_bucket.ml_models"
S3_ML_INDEXED = "module.s3.aws_s3_bucket.ml_models[0]"
LOGS_INDEXED = "aws_s3_bucket.logs[0]"
LOGS_UNINDEXED = "aws_s3_bucket.logs"


class ScanNotLive(Exception):
    """A scan returned something unparseable, or proved it was not measuring.

    Never a verdict: a scan in this state says nothing about the tree, so it is
    refused rather than reported. Exit 3 ("nothing was measured") is deliberately
    distinct from exit 1 ("the claims are false") -- conflating the two is how a
    broken probe comes to look like a clean bill of health.
    """


class HarnessFault(Exception):
    """A mutation could not be applied to the tree it was asked to mutate.

    Also not a verdict, and a DIFFERENT not-a-verdict from ScanNotLive: nothing
    was even attempted, as opposed to attempted and refused. Exit 2.

    This exists because both mutators used to signal "the terraform changed
    shape" by raising ValueError out of str.index, or SystemExit("...") -- and
    neither was caught, so the traceback exited 1, the code this file defines as
    "a claim was falsified". A renamed or reordered resource would then have made
    the probe announce that CHECKOV_SCANER_LIMITATION.md is stale, when in fact
    the document had not changed and only the harness had lost its grip. A false
    accusation of the document is as bad as a false clearance of it.
    """


@dataclass
class Scan:
    """One checkov run, addressed by full resource NAME.

    Both halves of the verdict are kept by name, because a claim about which
    resource moved can only be checked against the resource that actually moved,
    and this harness refuses to reason from counts alone.
    """

    passed: list[str] = field(default_factory=list)
    failed: list[str] = field(default_factory=list)
    resource_count: int | None = None
    parsing_errors: int | None = None

    @property
    def names(self) -> list[str]:
        """Every resource name the scan reported, in either verdict."""
        return sorted(self.passed + self.failed)

    def describe(self) -> str:
        return f"passed={len(self.passed)} failed={len(self.failed)}"


def _payload(check: str, tf_dir: Path) -> dict | None:
    """Run checkov once and return its parsed JSON object, or None.

    checkov emits ONE JSON OBJECT per framework, not an array. A parser that
    only handles [...] returns nothing, and "nothing" is indistinguishable from
    "clean" -- the bug this harness was written to stop.

    One call, parsed once. A previous revision re-ran checkov just to lift the
    `summary` block out of a second payload, which doubled an already slow scan
    for no gain: `summary` sits beside `results` at the top level of the very
    same object.
    """
    proc = subprocess.run(
        [sys.executable, "-m", "checkov.main", "-d", str(tf_dir),
         "--framework", "terraform", "--compact", "--check", check, "-o", "json"],
        capture_output=True, text=True, encoding="utf-8", errors="replace",
    )
    try:
        parsed = json.loads(proc.stdout)
    except Exception:
        return None
    return parsed if isinstance(parsed, dict) else None


def scan(check: str, tf_dir: Path) -> Scan | None:
    """Scan `tf_dir` for `check`, or None if checkov produced no usable JSON.

    Both halves of the verdict are kept BY NAME, because every claim here is
    about which resource moved, and a name is the only thing that can carry
    that. Counts alone cannot distinguish "logs passed" from "logs was never
    scanned", which is the confusion this harness exists to prevent.

    `tf_dir` is a REQUIRED argument on purpose. It used to default to the
    module-global TF_DIR, so the mutation harness scanned the ORIGINAL worktree
    and every "mutated" result was really the baseline. The probe reported
    FALSIFIED for T1/T3 on that run -- it was measuring the wrong tree, not
    refuting anything. A harness that cannot be pointed at a different tree
    cannot test a mutation at all.
    """
    payload = _payload(check, tf_dir)
    if payload is None:
        return None
    results = payload.get("results") or {}
    summary = payload.get("summary") or {}
    return Scan(
        passed=[c.get("resource", "?") for c in results.get("passed_checks", [])],
        failed=[c.get("resource", "?") for c in results.get("failed_checks", [])],
        resource_count=summary.get("resource_count"),
        parsing_errors=summary.get("parsing_errors"),
    )


def require_live(claim: Scan | None, control: Scan | None, where: str,
                 floor: int | None = None) -> None:
    """Raise ScanNotLive unless both scans prove they measured this tree.

    Applied to EVERY scan, mutated or not -- not just the baseline. A mutated
    scan is the case that matters: the mutation is expected to change the
    finding set, so an empty or vacuous result from one is exactly the shape a
    "the mutation worked" false PASS arrives in, and a count-only assertion
    (`passed + failed > 0`) cannot tell a real verdict from a scan that stopped
    scanning. The control is what distinguishes them.

    A scan can lie four ways:

      * no JSON          -> one of the two Scans is None
      * 0 resources      -> the tree failed to parse and produced a vacuous "0
                            failures", the shape a broken scan and the shape of a
                            clean tree share
      * control not red  -> the run is not seeing the known-red finding set, so
                            its silence about anything else means nothing
      * collapse         -> `floor` is set (mutated scans, against the baseline's
                            resource count) and the scan came back with far
                            fewer resources than the tree demonstrably has

    The fourth is the one a non-zero count cannot catch. A mutation that breaks
    the terraform does not always yield 0 resources: checkov can happily report
    the handful of files that still parse and call that a clean result. A floor
    of half the baseline is deliberately loose -- T1/T2a legitimately remove one
    resource and T3 renames one -- and still refuses any scan that only saw a
    fraction of the tree.
    """
    if claim is None or control is None:
        missing = CLAIM_CHECK if claim is None else CONTROL_CHECK
        raise ScanNotLive(f"{where}: checkov produced no usable JSON for {missing}")
    if not claim.resource_count:
        raise ScanNotLive(f"{where}: scan reported 0 resources "
                          f"(parsing_errors={claim.parsing_errors}); 'no findings' "
                          f"here means 'nothing was scanned', so no verdict is read "
                          f"from it")
    if floor is not None and claim.resource_count < floor:
        raise ScanNotLive(f"{where}: scan reported {claim.resource_count} resources, "
                          f"below the floor of {floor} (half the baseline's); the "
                          f"tree looks broken and only part of it was scanned, so "
                          f"no verdict is read from it "
                          f"(parsing_errors={claim.parsing_errors})")
    if not control.failed:
        raise ScanNotLive(f"{where}: control {CONTROL_CHECK} is known-red on this tree "
                          f"but reported 0 failures ({control.describe()}); this scan "
                          f"is not measuring the real finding set")


def short(names):
    return sorted({n.split(".")[-1] for n in names})


def mutated(mutator, label: str, floor: int | None = None) -> tuple[Scan, Scan]:
    """Copy the tree, mutate the COPY, scan it. Never mutates the worktree.

    Returns (claim, control) read from the SAME copy. Both scans must describe
    one set of bytes: two separately built copies could differ, and the control
    would then be certifying a tree that nobody measured a claim against.

    `require_live` runs on every pair, so a mutated scan that came back empty
    or vacuously clean is refused here rather than reported as a verdict. `floor`
    is the baseline's half-resource count, so a mutation that leaves the tree
    half-unscannable is refused even though it still reports some resources.
    """
    holder = tempfile.TemporaryDirectory(dir=str(REPO.parent))
    try:
        tree = Path(holder.name) / "repo"
        shutil.copytree(TF_DIR, tree / TF_DIR.name)
        copy_tf = tree / TF_DIR.name
        mutator(copy_tf)
        claim = scan(CLAIM_CHECK, copy_tf)
        control = scan(CONTROL_CHECK, copy_tf)
    finally:
        holder.cleanup()
    require_live(claim, control, label, floor=floor)
    return claim, control


def delete_block(rel: str, name: str):
    """Remove one `..._public_access_block` resource from a module file.

    Every way this can fail to find its target raises HarnessFault (exit 2)
    rather than propagating ValueError/IndexError, because an uncaught traceback
    exits 1 -- indistinguishable, to anything reading this probe's exit code,
    from "a claim was falsified". It must never be able to accuse the document.
    """
    def apply(tf: Path):
        target = tf / rel
        text = target.read_text(encoding="utf-8")
        header = f'resource "{BLOCK}" "{name}"'
        try:
            start = text.index(header)
            end = text.index("\nresource ", start + 10)
        except ValueError as exc:
            raise HarnessFault(
                f"could not locate {header} in {rel} ({exc}); the module "
                f"changed shape, so this mutation cannot be applied and no "
                f"verdict can be drawn from it") from exc
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
    # Requiring BOTH edits to land is what makes this a measurement rather than
    # a hope. Dropping `count` from the bucket while its block still said
    # logs[0].id would leave a broken tree, and a broken tree that happens to
    # scan would otherwise be reported as "the index is the cause".
    if text == before:
        raise HarnessFault(
            "unindex_logs matched nothing -- the scaling module changed shape "
            "(neither the count line on aws_s3_bucket.logs nor the one on its "
            "block was found)")
    if text.count('resource "aws_s3_bucket" "logs" {') != 1:
        raise HarnessFault("unindex_logs did not match exactly once")
    target.write_text(text, encoding="utf-8")


def main() -> int:
    base = scan(CLAIM_CHECK, TF_DIR)
    base_control = scan(CONTROL_CHECK, TF_DIR)
    try:
        require_live(base, base_control, "BASELINE")
    except ScanNotLive as exc:
        print(f"REFUSED: {exc}")
        print("Nothing has been measured.")
        # 2 when checkov produced nothing at all (a harness/environment fault);
        # 3 when it ran but the scan could not be shown live (a measurement
        # fault). Both are "no verdict", and neither may be mistaken for 0.
        return 2 if (base is None or base_control is None) else 3

    print(f"CONTROL {CONTROL_CHECK}: {base_control.describe()} -> scan is live")
    print(f"BASELINE {CLAIM_CHECK}: {base.describe()}")
    for name in short(base.failed):
        print(f"    FAIL {name}")
    print()

    # Every mutated scan must still see most of the tree. T1/T2a legitimately
    # remove one resource and T3 renames one, so half the baseline is a
    # deliberately loose floor -- loose enough to tolerate the mutations,
    # tight enough that a mutation which breaks the tree cannot slip through
    # with the few resources that still parse.
    floor = max(1, base.resource_count // 2)
    print(f"MUTATED-SCAN FLOOR: {floor} resources "
          f"(half the baseline's {base.resource_count})\n")

    verdicts: dict[str, bool] = {}

    try:
        t1, t1_ctl = mutated(delete_block("modules/s3/main.tf", "ml_models"),
                             "T1 delete s3 ml_models block", floor=floor)
        # T1: the unindexed verdict must TRACK its control. Deleting the block
        # has to produce a NEW failure, and for the same unindexed resource the
        # baseline reported as PASSING -- otherwise "a failure appeared" is a
        # claim about some other resource entirely.
        t1_ok = (S3_ML in t1.failed
                 and S3_ML in base.passed
                 and S3_ML not in base.failed)
        verdicts["T1"] = t1_ok
        print(f"T1  delete s3 ml_models block  -> {t1.describe()} "
              f"(control on the same copy: {t1_ctl.describe()})")
        print(f"    {'OK' if t1_ok else 'FALSIFIED'}: {S3_ML} moved PASS -> FAIL")

        t2, t2_ctl = mutated(delete_block("modules/scaling/main.tf", "logs"),
                             "T2a delete scaling logs block", floor=floor)
        t2_ok = sorted(t2.failed) == sorted(base.failed)
        verdicts["T2a"] = t2_ok
        print(f"\nT2a delete scaling logs block -> {t2.describe()} "
              f"(control on the same copy: {t2_ctl.describe()})")
        print(f"    {'OK' if t2_ok else 'FALSIFIED'}: failure set unchanged -> "
              f"[0] ignores the control")

        # T2b reuses T1's copy deliberately: the point is that the OTHER module's
        # indexed resource stayed blind under the same mutation.
        t2b_ok = S3_ML_INDEXED in t1.failed
        verdicts["T2b"] = t2b_ok
        print(f"    {'OK' if t2b_ok else 'FALSIFIED'}: [0] still blind in the "
              f"second module")

        t3, t3_ctl = mutated(unindex_logs, "T3 remove count from logs pair", floor=floor)
        # T3 is the load-bearing claim ("the index is the cause"), so it asserts
        # MOVEMENT rather than absence. Removing `count` RENAMES the resource, so
        # the bucket that failed as `logs[0]` must now be PRESENT IN PASSED as
        # `logs`: found and passing, not merely no longer failing. A scan that
        # dropped the resource altogether reports neither name and so fails the
        # second clause -- which is precisely the ambiguity this replaces.
        t3_ok = (LOGS_INDEXED in base.failed
                 and LOGS_UNINDEXED in t3.passed
                 and LOGS_INDEXED not in t3.failed
                 and LOGS_UNINDEXED not in t3.failed)
        verdicts["T3"] = t3_ok
        print(f"\nT3  remove count from logs pair -> {t3.describe()} "
              f"(control on the same copy: {t3_ctl.describe()})")
        print(f"    {'OK' if t3_ok else 'FALSIFIED'}: {LOGS_INDEXED} moved "
              f"FAILED -> {LOGS_UNINDEXED} PASSED -> the index is the cause")
    except ScanNotLive as exc:
        print(f"\nREFUSED: {exc}")
        print("Nothing was measured from that scan, so no verdict is reported "
              "for it.")
        return 3
    except HarnessFault as exc:
        print(f"\nHARNESS FAULT: {exc}")
        print("No mutation could be applied, so no verdict is reported. This is "
              "NOT evidence that a claim in CHECKOV_SCANER_LIMITATION.md is "
              "false -- the harness lost its grip on the tree, the document is "
              "untested.")
        return 2

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