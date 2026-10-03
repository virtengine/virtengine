#!/usr/bin/env python3
"""END-TO-END NEGATIVE TESTS for ckv2_indexed_limitation_probe.py.

Each mode runs the probe's OWN main() -- unmodified -- against a copy of the
tree whose mutation is deliberately misbehaving, and asserts the probe refuses
to report a PASS. The probe file is never modified: it is imported and its
TF_DIR global is redirected at a throwaway copy, so the code under test is
byte-identical to what ships. Each run costs a full run's scans (~6 min).

  wipe     THE ONE THAT MATTERS. The mutator destroys the whole tree it is
           handed, so the MUTATED scan comes back empty while the baseline --
           scanned from the untouched TF_DIR -- is perfectly live. This is the
           "valid empty result is indistinguishable from a verdict" gap: it
           must exit 3 (nothing measured), and must never exit 0.

  dangling Drops `count` from the logs bucket only, leaving its block still
           saying `logs[0].id`. Recorded as INFORMATIONAL: measured on
           checkov 3.3.22, checkov tolerates the dangling reference, still scans
           the bucket, and still reports `aws_s3_bucket.logs` as genuinely
           PASSING -- so exit 0 is a TRUE verdict, not a false pass. This mode
           asserts the verdict is justified (the unindexed name really is in the
           passed list) instead of asserting a failure.

Usage:  python ckv2_indexed_limitation_probe_negtest.py [wipe|dangling]
Exit:   0 = the probe refused (or, for dangling, its PASS was justified).
        1 = the probe reported an unjustified PASS -- a real defect.
        2 = this driver could not set up or run the probe -- a fault HERE, not
            a verdict about the probe, and kept distinct from 1 on purpose.
"""
from __future__ import annotations

import importlib.util
import io
import shutil
import sys
import tempfile
import traceback
from contextlib import redirect_stdout
from pathlib import Path

MODE = sys.argv[1] if len(sys.argv) > 1 else "wipe"
if MODE not in ("wipe", "dangling"):
    print(f"usage: {Path(__file__).name} [wipe|dangling]")
    sys.exit(2)

HERE = Path(__file__).resolve().parent
probe_path = HERE / "ckv2_indexed_limitation_probe.py"
spec = importlib.util.spec_from_file_location("probe_under_test", probe_path)
probe = importlib.util.module_from_spec(spec)
# Must be registered BEFORE exec_module: @dataclass resolves annotations by
# looking the module up in sys.modules, and without this the import dies with
# `AttributeError: 'NoneType' object has no attribute '__dict__'`. That failure
# is a traceback plus this driver's own non-zero exit -- indistinguishable, by
# exit code alone, from "the probe correctly refused", which is exactly how a
# broken driver gets mistaken for a passing test.
sys.modules[spec.name] = probe
try:
    spec.loader.exec_module(probe)
except Exception:                                       # noqa: BLE001
    print("SETUP FAULT: could not import the probe under test:")
    traceback.print_exc()
    sys.exit(2)

holder = tempfile.TemporaryDirectory(dir=str(probe.REPO.parent))
copy = Path(holder.name) / "terraform"
shutil.copytree(probe.TF_DIR, copy)
probe.TF_DIR = copy            # redirect the real probe at the copy


def wipe(rel: str, name: str):
    """Return a mutator that destroys every .tf file in the copy.

    The mutated scan then finds nothing, while the baseline -- scanned from the
    untouched TF_DIR -- stays perfectly live. The only thing distinguishing a
    real verdict from a vacuous one here is the probe's own liveness control,
    which is exactly what is under test.

    Returns a callable rather than mutating directly because it REPLACES
    `delete_block`, whose contract is a factory: main() calls
    `delete_block(rel, name)` to GET the mutator, then `mutated()` calls it with
    the tree path.
    """
    def apply(tf: Path) -> None:
        for tf_file in tf.rglob("*.tf"):
            tf_file.write_text("this is not terraform {{{\n", encoding="utf-8")
    return apply


def dangling(rel: str, name: str):
    """Return a mutator dropping `count` from the logs bucket ONLY.

    The block keeps saying `logs[0].id`, so the pair is left inconsistent. Also
    a factory, for the same reason as `wipe`. The path argument is unused: this
    mutation is always about the scaling module's logs pair, wherever it is
    copied to.
    """
    def apply(tf: Path) -> None:
        target = tf / "modules/scaling/main.tf"
        text = target.read_text(encoding="utf-8")
        mutated = text.replace(
            'resource "aws_s3_bucket" "logs" {\n'
            "  count = var.enable_global_accelerator ? 1 : 0\n",
            'resource "aws_s3_bucket" "logs" {\n', 1)
        if mutated == text:
            raise probe.HarnessFault("dangling mutator matched nothing (tree moved)")
        target.write_text(mutated, encoding="utf-8")
    return apply


# Both modes sabotage T1's mutator, which runs first, so the run stops there.
probe.delete_block = wipe if MODE == "wipe" else dangling

print("=" * 70)
print(f"NEGATIVE MODE: {MODE}")
if MODE == "wipe":
    print("Mutated tree destroyed; baseline left live. Probe must NOT exit 0.")
else:
    print("Dangling logs[0].id left behind. Probe must not PASS without the")
    print("unindexed resource genuinely appearing in the passed list.")
print("=" * 70)
_captured = io.StringIO()
try:
    with redirect_stdout(_captured):
        code = probe.main()
except BaseException:                                   # noqa: BLE001
    print("\nSETUP FAULT: the probe raised out of main():")
    traceback.print_exc()
    holder.cleanup()
    sys.exit(2)

# Echo what the probe reported, so the record shows what was actually measured
# rather than only this driver's summary of it.
print(_captured.getvalue())

print("=" * 70)
print(f"probe exit code = {code}")

if MODE == "wipe":
    if code == 0:
        print("NEGATIVE TEST FAILED: the probe reported a PASS on a destroyed "
              "tree -- exactly the gap this test exists to catch")
        rc = 1
    else:
        print(f"NEGATIVE TEST PASSED: destroyed tree produced exit {code}, not 0")
        rc = 0
else:
    # Informational: assert only that a PASS is justified, i.e. the probe's own
    # movement assertion really did see the unindexed name pass. The run's
    # stdout above is the record of what checkov did.
    justified = "moved FAILED -> aws_s3_bucket.logs PASSED" in _captured.getvalue()
    if code == 0 and justified:
        print("INFORMATIONAL: checkov tolerated the dangling reference and "
              "genuinely reported the unindexed bucket as PASSING; exit 0 is a "
              "true verdict, not a false pass")
        rc = 0
    elif code == 0:
        print("NEGATIVE TEST FAILED: exit 0 without the movement assertion "
              "firing -- the PASS is not justified")
        rc = 1
    else:
        print(f"INFORMATIONAL: probe refused (exit {code})")
        rc = 0

holder.cleanup()
sys.exit(rc)