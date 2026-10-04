"""Falsification harness for .github/tests/test_image_size_gate.py.

A green suite that cannot go red is decoration (ESTATE.md's recurring lesson:
"a guard that exercises a PROXY for the production path is decoration wearing a
checkmark"). This mutates a COPY of the gate inside an ISOLATED TREE and
requires the owning test to fail, so each guard is proven able to catch the
defect it exists for.

It never writes to the repository working tree: `scripts/ci/image_size_gate.py`,
`.github/tests/test_image_size_gate.py` and `.github/workflows/ci.yaml` are
copied into a temp tree, and only those copies are mutated. The tree is built by
COPYING THE REAL FILES, so the suite runs against the production bytes.

Run: python falsify_image_size_gate.py
Exit 0 = every mutation caught. Exit 1 = a mutation survived. Exit 2 = harness error.
"""
from __future__ import annotations

import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
WORKTREE = REPO_ROOT
GATE_REL = Path("scripts") / "ci" / "image_size_gate.py"
TESTS_REL = Path(".github") / "tests" / "test_image_size_gate.py"
CI_REL = Path(".github") / "workflows" / "ci.yaml"

# (tag, original, replacement, test that must go red)
# Some defects need TWO coordinated edits to be expressible, so an entry may
# carry an optional `extra_gate_mutation=(original, replacement)` applied on
# top of the primary one, in the same isolated tree, before the suite is run.
MUTATIONS = [
    (
        "M1-ceiling-blanked-to-target",
        "MAX_IMAGE_BYTES = 211 * 1024 * 1024",
        "MAX_IMAGE_BYTES = TARGET_IMAGE_BYTES",
        "test_the_measured_image_passes_the_ratchet",
    ),
    (
        "M2-regression-check-inverted",
        "if image_bytes > max_bytes:",
        "if image_bytes < max_bytes:",
        "test_ratchet_still_fails_a_real_regression",
    ),
    (
        "M3-regression-neutered",
        "if image_bytes > max_bytes:",
        "if False:",
        "test_ratchet_still_fails_a_real_regression",
    ),
    (
        "M4-measurement-made-convenient",
        "MEASURED_IMAGE_BYTES = 216548169",
        "MEASURED_IMAGE_BYTES = 100 * 1024 * 1024",
        "test_the_recorded_measurement_is_the_real_one",
    ),
    (
        "M5-unmeasured-passes",
        "    if image_bytes is None:",
        "    if False:",
        "test_missing_size_fails_closed",
    ),
    (
        "M6-expiry-neutered",
        "    if review_expired(today):",
        "    if False:",
        "test_expired_review_window_fails_the_gate",
    ),
    (
        "M7-expiry-pushed-to-the-far-future",
        "RATCHET_EXPIRES = date(2026, 11, 2)",
        "RATCHET_EXPIRES = date(2099, 1, 1)",
        "test_review_window_is_a_real_window",
    ),
    (
        "M8-tracking-issue-dropped",
        'BUDGET_ISSUE = "virtengine/virtengine#1159"',
        'BUDGET_ISSUE = "nobody"',
        "test_budget_names_a_tracking_issue",
    ),
    (
        "M9-headroom-unbounded",
        "MAX_IMAGE_BYTES = 211 * 1024 * 1024",
        "MAX_IMAGE_BYTES = 10 * 1024 * 1024 * 1024",
        "test_ratchet_is_tight_enough_to_notice_growth",
    ),
    (
        "M10-nonpositive-accepted",
        "    if image_bytes <= 0:",
        "    if False:",
        "test_nonpositive_size_is_rejected",
    ),
    (
        "M11-aspiration-target-deleted",
        "TARGET_IMAGE_BYTES = 100 * 1024 * 1024",
        "TARGET_IMAGE_BYTES = 211 * 1024 * 1024",
        "test_target_is_reported_not_deleted",
    ),
    (
        # Both layers of the fallback are removed together. Removing only the
        # `raw or <huge>` line is NOT caught any more, because the one-sided
        # `value > recorded` guard independently refuses the resulting huge
        # budget -- defence in depth, not decoration. So this mutation asks the
        # sharper question: with BOTH protections gone, does the junk path fail?
        "M12-junk-override-accepted-both-layers-removed",
        '    raw = os.environ.get("IMAGE_SIZE_MAX", "").strip()\n    if not raw:',
        '    raw = os.environ.get("IMAGE_SIZE_MAX", "").strip() or "999999999999"\n    if not raw:',
        "test_junk_and_nonpositive_overrides_fall_back",
        ("    if value > recorded:", "    if False:"),
    ),
    (
        # `if value > 0` refuses EVERY positive override, so it breaks
        # TIGHTENING, not raising. The owning test is the tighten one; asserting
        # the raise test would be a green that proves nothing.
        "M15-raising-override-allowed-again",
        "    if value > recorded:",
        "    if False:",
        "test_override_cannot_raise_the_ceiling",
    ),
    (
        "M16-tightening-broken-by-unconditional-guard",
        "    if value > recorded:",
        "    if value > 0:",
        "test_override_can_tighten_the_ceiling",
    ),
    (
        "M17-dead-workflow-literal-restored",
        "IMAGE_SIZE_BYTES=\"${SIZE_BYTES}\" python3 scripts/ci/image_size_gate.py virtengine:ci",
        "MAX_BYTES=$((100 * 1024 * 1024))",
        "test_workflow_no_longer_compares_against_the_unreachable_literal",
    ),
    (
        "M18-self-referencing-env-restored",
        "          IMAGE_SIZE_MAX: ${{ vars.IMAGE_SIZE_MAX }}",
        "          IMAGE_SIZE_BYTES: ${{ steps.image_size.outputs.size_bytes }}",
        "test_env_does_not_self_reference_the_step_output",
    ),
    # ---- M19-M29: the recorded-measurement guards ------------------------
    # These cover the constants added when MEASURED_BINARY_BYTES was corrected
    # to the workspace-mode build, and the derived runtime layer. They exist
    # because an earlier version of this suite constrained the unstripped
    # constant only by `assertGreater(unstripped, stripped)` -- which admits
    # any LARGER wrong value, so the gate's "stripping already removes 77.03
    # MiB" line could be fabricated while the suite stayed green.
    (
        "M19-unstripped-constant-made-convenient",
        "MEASURED_UNSTRIPPED_BINARY_BYTES = 274068165",
        "MEASURED_UNSTRIPPED_BINARY_BYTES = 200000000",
        "test_recorded_unstripped_binary_is_the_measured_one",
    ),
    (
        # The sharpest version: barely above the stripped size, so stripping
        # reports 0.00 MiB while the whole point is that it removes 77 MiB.
        "M20-unstripped-constant-barely-above-stripped",
        "MEASURED_UNSTRIPPED_BINARY_BYTES = 274068165",
        "MEASURED_UNSTRIPPED_BINARY_BYTES = 193294499",
        "test_recorded_unstripped_binary_is_the_measured_one",
    ),
    (
        "M21-unstripped-print-gutted",
        '''    print(f"Recorded binary unstripped: {MEASURED_UNSTRIPPED_BINARY_BYTES} bytes "
          f"({MEASURED_UNSTRIPPED_BINARY_BYTES / mib:.2f} MiB), so stripping already "
          f"removes {(MEASURED_UNSTRIPPED_BINARY_BYTES - MEASURED_BINARY_BYTES) / mib:.2f} MiB")
''',
        '    print(f"Recorded binary unstripped: {MEASURED_UNSTRIPPED_BINARY_BYTES} bytes")\n',
        "test_derived_runtime_layer_is_printed_with_its_provenance",
    ),
    (
        "M22-docstring-runtime-figure-falsified",
        "23,253,671 B (22.18 MiB), which does fit under the budget on its own",
        "21,000,000 B (20.03 MiB), which does fit under the budget on its own",
        "test_docstring_runtime_figure_matches_the_derived_value",
    ),
    (
        # Zeroing the layer and inverting the subtraction both attribute an
        # impossible (zero / negative) runtime layer to the image.
        "M23-derived-runtime-zeroed",
        "    return MEASURED_IMAGE_BYTES - MEASURED_BINARY_BYTES",
        "    return 0",
        "test_derived_runtime_layer_is_the_subtraction_and_is_bounded",
    ),
    (
        "M24-derived-runtime-subtraction-inverted",
        "    return MEASURED_IMAGE_BYTES - MEASURED_BINARY_BYTES",
        "    return MEASURED_BINARY_BYTES - MEASURED_IMAGE_BYTES",
        "test_derived_runtime_layer_is_the_subtraction_and_is_bounded",
    ),
    (
        "M25-docstring-stripped-figure-falsified",
        "193,294,498 B (184.34 MiB)",
        "193,294,497 B (184.34 MiB)",
        "test_docstring_runtime_figure_matches_the_derived_value",
    ),
    (
        "M26-docstring-unstripped-figure-falsified",
        "274,068,165 B (261.37 MiB)",
        "274,067,495 B (261.37 MiB)",
        "test_docstring_runtime_figure_matches_the_derived_value",
    ),
    (
        # Output assertions cannot catch these two: hardcoding the CURRENTLY
        # correct value produces byte-identical output. The literal then rots
        # silently on the next re-measurement. Caught structurally instead.
        "M27-derived-runtime-print-hardcoded",
        "    derived_runtime = derived_runtime_layer_bytes()",
        "    derived_runtime = 23253671",
        "test_printed_figures_are_derived_not_hardcoded",
    ),
    (
        "M28-unstripped-print-figure-hardcoded",
        'f"removes {(MEASURED_UNSTRIPPED_BINARY_BYTES - MEASURED_BINARY_BYTES) / mib:.2f} MiB")',
        'f"removes 77.03 MiB")',
        "test_printed_figures_are_derived_not_hardcoded",
    ),
    (
        "M29-derived-runtime-binary-ignored",
        "    return MEASURED_IMAGE_BYTES - MEASURED_BINARY_BYTES",
        "    return MEASURED_IMAGE_BYTES",
        "test_derived_runtime_layer_is_the_subtraction_and_is_bounded",
    ),
]


def build_tree(root: Path) -> Path:
    """Copy the production files the suite reads into an isolated tree."""
    tree = root / "repo"
    for rel in (GATE_REL, TESTS_REL, CI_REL):
        dest = tree / rel
        dest.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(WORKTREE / rel, dest)
    return tree


def run_suite(tree: Path) -> tuple[int, str]:
    proc = subprocess.run(
        [
            sys.executable,
            "-m",
            "unittest",
            "discover",
            "-s",
            str(tree / TESTS_REL.parent),
            "-p",
            TESTS_REL.name,
        ],
        cwd=tree,
        capture_output=True,
        text=True,
    )
    return proc.returncode, proc.stdout + proc.stderr


def main() -> int:
    with tempfile.TemporaryDirectory() as tmp:
        tmpdir = Path(tmp)
        tree = build_tree(tmpdir)

        rc, out = run_suite(tree)
        if rc != 0:
            print("HARNESS ERROR: baseline suite is not green in the isolated tree")
            print(out[-4000:])
            return 2
        print(f"baseline: GREEN on copied production files in an isolated tree")

        gate_path = tree / GATE_REL
        original_gate = gate_path.read_bytes()
        original_ci = (tree / CI_REL).read_bytes()
        survived = []

        for entry in MUTATIONS:
            tag, original, replacement, owning_test = entry[:4]
            extra_gate = entry[4] if len(entry) > 4 else None
            target = gate_path
            text = target.read_text(encoding="utf-8")
            if original not in text:
                target = tree / CI_REL
                text = target.read_text(encoding="utf-8")
            if original not in text:
                print(f"  SKIP    {tag}: anchor not found -- source drifted")
                survived.append(f"{tag} (anchor missing)")
                continue

            mutated = text.replace(original, replacement, 1)
            if extra_gate is not None:
                extra_original, extra_replacement = extra_gate
                if extra_original not in mutated:
                    print(f"  SKIP    {tag}: extra anchor not found -- source drifted")
                    survived.append(f"{tag} (extra anchor missing)")
                    continue
                if target != gate_path:
                    print(f"  SKIP    {tag}: extra anchor is in the gate, not the workflow")
                    survived.append(f"{tag} (extra anchor wrong file)")
                    continue
                mutated = mutated.replace(extra_original, extra_replacement, 1)

            target.write_text(mutated, encoding="utf-8")
            rc, out = run_suite(tree)
            target.write_bytes(original_gate if target == gate_path else original_ci)

            if rc == 0:
                print(f"  MISSED  {tag}: suite stayed GREEN -> {owning_test} IS DECORATION")
                survived.append(tag)
            else:
                fired = owning_test in out
                print(f"  CAUGHT  {tag}: suite RED, owning test fired = {fired}")
                if not fired:
                    survived.append(f"{tag} (red for the wrong reason)")

        # Hygiene: the copied sources must be byte-identical to production again.
        for rel, original in ((GATE_REL, original_gate), (CI_REL, original_ci)):
            if (tree / rel).read_bytes() != original:
                print(f"HYGIENE FAIL: {rel} copy was left mutated")
                return 1

    if survived:
        print(f"\nSURVIVED (decoration): {survived}")
        return 1
    print(f"\nall {len(MUTATIONS)} mutations CAUGHT; no repo file was written")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())