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
        python infra/terraform/ckv2_indexed_limitation_probe.py --self-test
Exit 0 = all claims hold. 1 = a claim was falsified. 2 = the harness itself
        failed: checkov produced nothing, or a mutation could not be applied.
        3 = a scan was not live, nothing was measured from it.
Only 0 and 1 are about the document. 2 and 3 are about the probe, and a mutator
that loses its grip on the terraform must never be able to exit 1 -- that would
announce the document stale when only the harness broke.

`--self-test` runs the verdict logic against a stubbed scanner, in seconds,
and asserts the exit-code contract: that a claim which stops holding exits 1,
that a dead scan and an unappliable mutation exit 3 and 2 respectively and NOT
1, and that T3's movement assertion rejects a resource which fell out of the
scan. It is the negative half of this file -- the full run above only ever
demonstrates that a healthy tree looks healthy, which is exactly the evidence a
harness that had quietly stopped measuring would also produce.

Requires checkov importable (`pip install checkov`) for a real run. If it is
not, exit 2 -- never a silent pass. `--self-test` needs no checkov.

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


def verdicts_from(base: Scan, t1: Scan, t2: Scan, t3: Scan) -> dict[str, bool]:
    """Decide T1/T2a/T2b/T3 from four live scans. Pure: no I/O, no exceptions.

    Split out of `main` so `--self-test` can drive the DECISION LOGIC itself
    rather than a paraphrase of it. A negative test written against a copy of
    the rules proves only that the copy behaves as the author expected; here
    the thing under test is the code that decides the probe's exit code.

    Every clause is about movement between two named resources, never about a
    count alone and never about the absence of a failure:

      T1   the unindexed verdict tracks the control -- `ml_models` was PASSING
           in the baseline and is FAILING once its block is deleted. Asserting
           only that a new failure appeared would be satisfied by a failure on
           some unrelated resource.
      T2a  the indexed verdict is blind -- deleting the logs block leaves the
           failure SET identical, resource for resource.
      T2b  ...and that blindness is not module-specific: the same mutation
           leaves `ml_models[0]` failing under T1's copy.
      T3   the index is the cause -- the bucket that failed as `logs[0]` is now
           PRESENT IN `t3.passed` under its unindexed name. Membership in the
           passed list is what separates "it passed" from "it stopped being
           scanned"; the latter reports neither name and fails this clause.
    """
    return {
        "T1": (S3_ML in base.passed and S3_ML not in base.failed
               and S3_ML in t1.failed),
        "T2a": sorted(t2.failed) == sorted(base.failed),
        "T2b": S3_ML_INDEXED in t1.failed,
        "T3": (LOGS_INDEXED in base.failed
               and LOGS_UNINDEXED in t3.passed
               and LOGS_INDEXED not in t3.failed
               and LOGS_UNINDEXED not in t3.failed),
    }


def _scan(passed: list[str], failed: list[str],
          resource_count: int = 342, parsing_errors: int = 0) -> Scan:
    return Scan(passed=list(passed), failed=list(failed),
                resource_count=resource_count, parsing_errors=parsing_errors)


def healthy_scans() -> tuple[Scan, Scan, Scan, Scan]:
    """The four scans a healthy tree produces, as of the numbers in the table."""
    base = _scan(
        passed=[S3_ML, "aws_s3_bucket.manifests", "module.s3.aws_s3_bucket.other",
                "aws_s3_bucket.tls", "module.s3.aws_s3_bucket.assets",
                "aws_s3_bucket.access_logs", "module.s3.aws_s3_bucket.data",
                "aws_s3_bucket.frontend", "aws_s3_bucket.backend"],
        failed=[S3_ML_INDEXED, LOGS_INDEXED],
    )
    t1 = _scan(passed=[n for n in base.passed if n != S3_ML],
               failed=base.failed + [S3_ML])          # 8 passed, 3 failed
    t2 = _scan(passed=base.passed, failed=base.failed)  # unchanged: 9, 2
    t3 = _scan(passed=base.passed + [LOGS_UNINDEXED],
               failed=[S3_ML_INDEXED])                 # 10 passed, 1 failed
    return base, t1, t2, t3


def _control_ok() -> Scan:
    """A scan of the control check that shows it still red on this tree."""
    return _scan(passed=["aws_s3_bucket.somewhere_else"],
                 failed=["module.s3.aws_s3_bucket.no_bucket_policy"])


def self_test() -> int:
    """Assert the exit-code contract against synthetic scans. Seconds, not minutes.

    A passing full run only ever shows that a healthy tree looks healthy. That
    is the same output a probe which stopped measuring entirely would produce,
    so these are the cases that actually discriminate: every one of them must
    FAIL loudly rather than report a PASS.
    """
    failures: list[str] = []
    assertions = 0

    # Every assertion goes through here, and the TOTAL is counted by the probe
    # itself and printed at the end. CHECKOV_SCANER_LIMITATION.md quotes that
    # total instead of carrying its own copy: this file used to print a
    # hardcoded "22/22" that had drifted from the real 24, in the one document
    # whose whole argument is that its numbers come from a script that fails
    # when they change. The pass criterion is "0 failures, exit 0" -- the total
    # is REPORTED, never asserted, so adding a case can never make this file
    # wrong. Anything that prints an ok/FAIL line must go through this.
    def check(label: str, ok: bool = True, detail: str = "") -> None:
        nonlocal assertions
        assertions += 1
        if not ok:
            failures.append(label)
        print(f"  {'ok  ' if ok else 'FAIL'}  {label}"
              f"{f' -- {detail}' if detail and not ok else ''}")

    print("VERDICT LOGIC")
    base, t1, t2, t3 = healthy_scans()
    v = verdicts_from(base, t1, t2, t3)
    check("healthy tree satisfies T1", v["T1"])
    check("healthy tree satisfies T2a", v["T2a"])
    check("healthy tree satisfies T2b", v["T2b"])
    check("healthy tree satisfies T3", v["T3"])

    print("\nREGRESSION: the two harness gaps this file was hardened against")

    # GAP 1, restated as a test. Under the OLD rule ("logs absent from
    # failed"), a T3 scan in which the resource was never scanned at all still
    # satisfies the assertion -- both names simply missing. Under the rule in
    # use, it must not.
    vanished = _scan(passed=base.passed, failed=[S3_ML_INDEXED])
    check("T3 rejects a resource that fell out of the scan entirely",
          not verdicts_from(base, t1, t2, vanished)["T3"])
    old_rule_would_have_passed = (LOGS_UNINDEXED not in vanished.failed
                                  and LOGS_INDEXED not in vanished.failed)
    check("  ...and that case is exactly what the old absence-rule missed",
          old_rule_would_have_passed)
    # Partially-dropped: present as unindexed but still failing.
    still_failing = _scan(passed=base.passed, failed=[S3_ML_INDEXED, LOGS_UNINDEXED])
    check("T3 rejects logs renamed but still failing",
          not verdicts_from(base, t1, t2, still_failing)["T3"])

    # GAP 2: T1 must not be satisfied by a failure on an unrelated resource.
    unrelated = _scan(passed=[n for n in base.passed if n != S3_ML],
                      failed=base.failed + ["aws_s3_bucket.some_other_bucket"])
    check("T1 rejects a new failure on the wrong resource",
          not verdicts_from(base, unrelated, t2, t3)["T1"])
    # T2a must not be satisfied by "the failure set shrank" -- it is a SET
    # equality, so losing a finding is a change.
    lost_one = _scan(passed=base.passed, failed=[S3_ML_INDEXED])
    check("T2a rejects a mutated scan that dropped a finding",
          not verdicts_from(base, t1, lost_one, t3)["T2a"])

    print("\nLIVENESS: a scan that cannot be shown live is REFUSED (exit 3)")

    def refused(scan_obj: Scan | None, control_obj: Scan | None,
                floor: int | None, label: str) -> None:
        try:
            require_live(scan_obj, control_obj, label, floor=floor)
        except ScanNotLive:
            check(f"{label}")
        else:
            check(label, False, "it was accepted, so a dead scan would pass")

    refused(None, _control_ok(), None, "no JSON at all is refused")
    refused(_scan([], [], resource_count=0), _control_ok(), None,
            "0 resources (empty tree and broken tree look alike) is refused")
    refused(_scan([], [], resource_count=0, parsing_errors=9), _control_ok(), None,
            "0 resources with parsing errors is refused")
    refused(base, _scan([], []), None,
            "control not red -- scan is not seeing the real finding set")
    refused(_scan(base.passed, base.failed, resource_count=12), _control_ok(),
            171, "collapse below the resource floor is refused")

    print("\nLIVENESS NOT INSUFFICIENT: live-looking scans still pass the gate")
    for label, obj, ctl, floor in (
        ("baseline at full count, no floor", base, _control_ok(), None),
        ("mutated at 9/8 and 9/2 with floor 171", t1, _control_ok(), 171),
    ):
        try:
            require_live(obj, ctl, label, floor=floor)
        except ScanNotLive as exc:
            check(label, False, f"was refused -- {exc}")
        else:
            check(label)

    print("\nMUTATOR GRIP: a mutation that cannot be applied is exit 2, NOT exit 1")
    # Runs against a THROWAWAY copy of the tree, never TF_DIR: these mutators
    # edit real files in place, and a self-test that left `ml_models` without
    # its public-access-block behind would break the very tree it is asserting
    # about. The copy is discarded either way.
    with tempfile.TemporaryDirectory(dir=str(REPO.parent)) as tmp:
        grip_tree = Path(tmp) / "terraform"
        shutil.copytree(TF_DIR, grip_tree)
        grip_failures = 0
        for rel, name in (("modules/s3/main.tf", "ml_models"),
                          ("modules/scaling/main.tf", "logs"),
                          ("modules/s3/main.tf", "no_such_bucket")):
            try:
                delete_block(rel, name)(grip_tree)
            except HarnessFault:
                grip_failures += 1
                continue
            except Exception as exc:                     # noqa: BLE001
                check(f"{rel}/{name} raised {type(exc).__name__}", False,
                      "which main() does NOT catch as a fault -> would exit 1")
                continue
            check(f"{rel}/{name} mutated")
        check("delete_block raises HarnessFault, never ValueError/SystemExit",
              grip_failures == 1)

        # The one target that genuinely exists must still be removable, or the
        # "fault" path above would pass on a mutator that never works at all.
        intact = (grip_tree / "modules/s3/main.tf").read_text(encoding="utf-8")
        check("the real target was removed from the copy, not just located",
              f'resource "{BLOCK}" "ml_models"' not in intact
              and 'resource "aws_s3_bucket" "ml_models"' in intact,
              "ml_models block still present after deletion")

        try:
            unindex_logs(grip_tree)
        except HarnessFault as exc:
            check(f"unindex_logs on a shape-changed tree -> HarnessFault "
                  f"({str(exc)[:48]}...)")
        except Exception as exc:                          # noqa: BLE001
            check(f"unindex_logs raised {type(exc).__name__}", False,
                  "which main() does NOT catch as a fault -> would exit 1")
        else:
            check("unindex_logs matched the tree as it stands")

    print("\nNON-OVERLAPPING EXIT CODES")
    # The falsified cases above must actually flip the summary verdict, or
    # main() would print "All claims hold" and return 0 on them.
    falsified = verdicts_from(base, unrelated, t2, vanished)
    check("a falsified claim flips the run to non-zero",
          not all(falsified.values()),
          f"verdicts={falsified}")
    check("exactly the broken claims flip False, the intact ones stay True",
          sorted(k for k, v in falsified.items() if not v) == ["T1", "T3"],
          f"expected T1 and T3 falsified, got {falsified}")
    check("ScanNotLive and HarnessFault are distinct types",
          not issubclass(ScanNotLive, HarnessFault)
          and not issubclass(HarnessFault, ScanNotLive))

    print()
    if failures:
        print(f"SELF-TEST FAILED: {len(failures)} of {assertions} assertions "
              f"did not hold:")
        for f in failures:
            print(f"    - {f}")
        return 1
    # The total is printed so the document can quote it instead of guessing it,
    # but it is not a pass criterion -- adding a case must not be able to make
    # this fail. See the comment on check().
    print(f"SELF-TEST PASSED: {assertions}/{assertions} assertions held, 0 "
          f"failures -- the probe's decision logic rejects every broken scan "
          f"above and accepts the healthy one.")
    return 0


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
        t2, t2_ctl = mutated(delete_block("modules/scaling/main.tf", "logs"),
                             "T2a delete scaling logs block", floor=floor)
        t3, t3_ctl = mutated(unindex_logs, "T3 remove count from logs pair", floor=floor)
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

    # One implementation of the rules, shared with --self-test. T2b reads T1's
    # copy deliberately: the point is that the OTHER module's indexed resource
    # stayed blind under that same mutation.
    verdicts = verdicts_from(base, t1, t2, t3)

    # T1: the unindexed verdict must TRACK its control. Deleting the block has
    # to produce a NEW failure on the SAME resource the baseline reported as
    # PASSING -- otherwise "a failure appeared" is a claim about something else.
    print(f"T1  delete s3 ml_models block  -> {t1.describe()} "
          f"(control on the same copy: {t1_ctl.describe()})")
    print(f"    {'OK' if verdicts['T1'] else 'FALSIFIED'}: {S3_ML} moved "
          f"PASS -> FAIL")

    print(f"\nT2a delete scaling logs block -> {t2.describe()} "
          f"(control on the same copy: {t2_ctl.describe()})")
    print(f"    {'OK' if verdicts['T2a'] else 'FALSIFIED'}: failure set "
          f"unchanged -> [0] ignores the control")
    print(f"    {'OK' if verdicts['T2b'] else 'FALSIFIED'}: [0] still blind in "
          f"the second module")

    # T3 is the load-bearing claim ("the index is the cause"), so it asserts
    # MOVEMENT rather than absence: removing `count` RENAMES the resource, so
    # the bucket that failed as `logs[0]` must be PRESENT IN PASSED as `logs`.
    # Found and passing, not merely no longer failing.
    print(f"\nT3  remove count from logs pair -> {t3.describe()} "
          f"(control on the same copy: {t3_ctl.describe()})")
    print(f"    {'OK' if verdicts['T3'] else 'FALSIFIED'}: {LOGS_INDEXED} moved "
          f"FAILED -> {LOGS_UNINDEXED} PASSED -> the index is the cause")

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
    if "--self-test" in sys.argv[1:]:
        raise SystemExit(self_test())
    raise SystemExit(main())