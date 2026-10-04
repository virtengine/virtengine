# CKV2_AWS_6 and indexed (`count`) resources — a *reproduced* limitation

**Status: the claims in this file are reproducible and mutation-tested. They
replace both the retracted claims of PR #1146 and the interim claims of the
previous version of this file, which asserted that these two findings were
genuine exposures. The gate stays RED. No `#checkov:skip` was written,
`soft_fail` stays `false`, no `skip-check` list was introduced, and no finding
was suppressed.**

The bucket and its public-access-block genuinely exist on `develop` for both
resources named below. That is not an assumption: it is the positive control
that makes this a limitation claim rather than a defect claim, and the probe
asserts it on every run.

## The two findings, and why they are artefacts

Baseline on `origin/develop @ 2191979db`, checkov **3.3.22** (the version
`ghcr.io/bridgecrewio/checkov:3.3.22` ships, which is what
`bridgecrewio/checkov-action@v12` pulls), `--check CKV2_AWS_6`:

```
summary: passed=9  failed=2  resource_count=342  parsing_errors=0

PASS  module.s3.aws_s3_bucket.ml_models      <- unindexed
FAIL  module.s3.aws_s3_bucket.ml_models[0]   <- indexed
FAIL  aws_s3_bucket.logs[0]                  <- indexed
```

`module.s3.aws_s3_bucket.ml_models` is reported **twice in one run, with
opposite verdicts**. That is the whole finding: the indexed copy cannot see a
control that the unindexed copy resolves.

## Evidence

Both resources carry a complete `aws_s3_bucket_public_access_block`, declared
with the same `count` expression as its bucket and referencing `[0].id`:

* `infra/terraform/modules/s3/main.tf:181` — `ml_models`
* `infra/terraform/modules/scaling/main.tf:200` — `logs`

| # | mutation | passed | failed | what it proves |
|---|---|---|---|---|
| T1 | delete `..._public_access_block.ml_models` | 9 → **8** | +`ml_models` | the unindexed verdict **does** track the control — the block is really attached |
| T2a | delete `..._public_access_block.logs` | 9 → **9** | *unchanged* | the indexed verdict is **blind** to the control |
| T2b | same tree as T1 | — | `ml_models[0]` still FAILED | indexed verdict blind in the **second** module too |
| **T3** | remove `count` from `aws_s3_bucket.logs` **and** its block | 9 → **10** | `logs[0]` **gone** | **the index is the cause** |

T3 is decisive: with the identical bucket, identical block, identical module and
identical everything except the `count`, the finding disappears entirely.
`logs[0]` fails whether its block is present (T2a) or deleted, and passes as
soon as the index is removed (T3). A verdict that ignores the control it claims
to test, in both directions, is not reporting the control.

## Reproduction

```bash
python infra/terraform/ckv2_indexed_limitation_probe.py
```

The probe copies the tree once per case, applies one mutation, re-scans, and
asserts each count moves exactly as the table above predicts. It exits non-zero
if any case fails to move, so this document cannot silently drift away from the
tool's behaviour — the failure mode of the previous two versions.

```bash
python infra/terraform/ckv2_indexed_limitation_probe.py --self-test
```

`--self-test` is the negative half, and it runs in **seconds with no checkov**:
it drives the same `verdicts_from` / `require_live` code the full run uses,
against synthetic scans that are deliberately wrong, and asserts each one is
rejected. A green full run only ever shows that a healthy tree looks healthy —
which is exactly what a probe that had quietly stopped measuring would also
print. The self-test is what discriminates, covering: a T3 scan in which the
logs resource fell out entirely (the case this file was hardened for), a T1
failure landing on the wrong resource, a mutated scan that lost a finding, each
of the four ways a scan can lie, a collapse below the resource floor, and the
three mutators' refusal to raise anything `main()` would report as exit 1. It
needs no checkov and no terraform copy, so it is cheap enough to run on every
edit to the probe.

The other half is slower and proves the same thing end to end, by running the
probe's own `main()` against a tree the mutation has deliberately broken:

```bash
python infra/terraform/ckv2_indexed_limitation_probe_negtest.py
```

The probe is not modified; only its `TF_DIR` global is redirected at a
throwaway copy. Costs a full run's worth of scans (~7 min), so it is manual too.
It has two modes:

* `wipe` (the default) — the mutator destroys every `.tf` file it is handed, so
  the mutated scan comes back empty while the **baseline** (scanned from the
  untouched tree) stays perfectly live. This is the "a valid empty result is
  indistinguishable from a verdict" gap, and it must exit 3, never 0.
* `dangling` — informational only, and it does **not** exit 0. This mode
  *replaces `probe.delete_block` wholesale* (so it sabotages T1 and T2a as well
  as running its own mutation), which is why the run reports T1 and T2a
  FALSIFIED and the probe exits **1**. Read that as a fact about this driver's
  design, not as evidence against the document: the whole point of the mode is
  what checkov does with a dangling `logs[0].id`, and that is T3's line, which
  still reads `OK: aws_s3_bucket.logs[0] moved FAILED -> aws_s3_bucket.logs
  PASSED`.

  Measured on checkov 3.3.22 (2026-10-04, `.ckv2probe-venv`), the substantive
  claim is **true**: checkov **tolerates** the dangling `logs[0].id` — it still
  scans the bucket and still reports `aws_s3_bucket.logs` as genuinely PASSING,
  so T3's movement assertion holds rather than being satisfied by a resource
  vanishing. That was previously asserted here as prose on the strength of a run
  that had not been executed; it is now a measured result. Worth recording
  because the inconsistency the reviewer expected to make `logs` vanish does not
  in fact do so on this tool version — the movement assertion is still the right
  protection, but it is protecting against the *scan dropping the resource*, not
  against a half-applied mutation.

Two things it refuses to do, because both previously produced a confident wrong
answer:

* **It never infers a pass from the absence of a failure.** T3 asserts that
  `aws_s3_bucket.logs` appears in the *passed* list, because "not failing" is
  satisfied equally by "passed" and by "stopped being scanned". Removing `count`
  renames the resource, so the name to look for is the unindexed one.
* **It never reports a verdict from a scan it could not show was live.** Every
  scan — baseline and mutated alike — must parse, report a non-zero resource
  count, and still report the control check `CKV_AWS_338` as failing. Mutated
  scans must additionally see at least half the baseline's resources: a broken
  tree does not necessarily scan as zero, because checkov reports the files that
  still parse, so a resource-count *floor* is what catches a partly-broken tree
  that a plain non-zero check would wave through.

Exit codes are kept non-overlapping on purpose, because the two ways this can go
wrong are "the document is wrong" and "the probe is wrong", and conflating them
is how a broken probe comes to look like a clean bill of health:

| exit | meaning | about |
|---|---|---|
| 0 | every claim above still holds | the document |
| 1 | a claim no longer holds | the document |
| 2 | checkov produced nothing, or a mutation could not be applied | the probe |
| 3 | a scan could not be shown live, so nothing was measured | the probe |

Only 0 and 1 are evidence about this document. A mutator that loses its grip on
the terraform (renamed resource, reshaped module) exits 2 — never 1, because a
harness fault must not be able to announce the document stale when only the
harness broke.

To see the baseline directly:

```bash
python -m checkov.main -d infra/terraform --framework terraform \
  --compact --check CKV2_AWS_6 -o json
```

### Running it

**Deliberately manual — not wired into CI, and should not be.** Measured on
2026-10-04 on the development host: **8 checkov runs, 431 s (7 min 11 s) wall
clock** end to end — baseline claim + control, then two scans per mutated case,
each after copying the whole `infra/terraform` tree. That is roughly what the
Infrastructure Checkov job itself costs, and the claims it re-measures only go
stale when checkov or the terraform changes — both of which already re-run the
gate that flags them. **Wire it in only if a claim here has actually been
contradicted in review, not on principle.** This decision is recorded so the
question does not get re-opened on a hunch.

The `--self-test` above is the exception and *is* cheap (seconds, no checkov),
because it makes no claim about the tree — it only checks that the probe's own
logic still rejects broken scans. That is the half worth running on every edit.

Requires `pip install checkov` (3.3.22, the version the CI image ships). Note it
must be the interpreter that runs the probe: the probe shells out to
`sys.executable -m checkov.main`, so checkov has to be importable there or the
run exits 2 rather than silently reporting nothing.

## What is NOT being claimed

* **Not** that `count` / `for_each` resources are invisible to `CKV2_*`. They are
  reported in both directions — as PR #1146's retraction also eventually
  conceded.
* **Not** that `count = 0` passes vacuously. It does not.
* **Not** that the other findings are unreal. **110 of the 112 are untouched by
  this.** It concerns exactly the 2 `CKV2_AWS_6` findings that sit on indexed
  resources whose control is present. What happened to those 110 is recorded in
  `CHECKOV_ACCEPTED_RISK.md`: they were dispositioned as accepted risk, per
  resource, on t_ae133441 — not "fixed on their own merits", which is what this
  file originally said. 41 of them are real defects carried deliberately and
  labelled `KNOWN GAP`; the rest are accepted design or tool-shape.
* **Not** a claim that the previous version of this file was right. It was not,
  and it suppressed a real finding — which is why every number above now comes
  from a script that fails when the numbers change.

## Disposition

`CKV2_AWS_6` on indexed resources is red for a reason that is not a defect in
this repository. The honest options are (a) fix checkov, (b) pin a version whose
graph resolves indexed references, or (c) have the gate compare indexed findings
against a resolved plan. **None is a suppression and none is done here** — the
gate is reported red with the reason attached, and the remaining 110 findings
are being fixed on their own merits.

## Traps in reading this gate's output

Both of these produced wrong conclusions here before, so they are recorded:

* checkov **prepends** an ANSI reset to the line after every `Guide:` URL, so a
  `Check:` header is literally `\x1b[0mCheck: ...`. Strip ANSI **first**; a
  `^Check:` anchor otherwise misses every header and keeps stale check ids live.
* `checkov -o json` emits a JSON **object** per framework, not an array. A
  parser handling only `[...]` returns zero findings and is indistinguishable
  from a clean tree.
* Never pair a result with a check id by line proximity — the log interleaves
  `PASSED`/`FAILED` under a moving header; walk it in order.

Refs: card `t_38224eaa` (authorisation to correct: project-steward, 2026-10-02),
Infrastructure run `37065277068` (`develop @ 2191979db`),
PR #1146 (retracted claim), PR #1145/#1151/#1153 (the fixes, undisputed).
The probe's own hardening (assert movement, not absence; control every scan)
came from the project-steward cross-review of PR #1171, card `t_235d925a`, and
card `t_285211f4`.

Re-verified 2026-10-04 on this hardening, checkov 3.3.22: full run exits 0 in
431 s with `passed=9 failed=2` baseline and the table's `8/3`, `9/2`, `9/2`,
`10/1` reproduced exactly, control `CKV_AWS_338` red (5 failures) on the
baseline **and** on every mutated copy, and T3 confirming `aws_s3_bucket.logs`
present in PASSED — movement, not absence. `--self-test` reports 24/24
assertions, 0 failures (the count is derived by the probe and printed by it, so
this line quotes a script's output rather than carrying a number that can drift;
the pass criterion is 0 failures, not the total).