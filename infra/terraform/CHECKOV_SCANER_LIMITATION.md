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
2026-10-04 on the development host: **8 checkov runs, ~9 minutes wall clock** end
to end, each run copying the whole `infra/terraform` tree. That is more than the
Infrastructure Checkov job itself costs, and the claims it re-measures only go
stale when checkov or the terraform changes — both of which already re-run the
gate that flags them. **Wire it in only if a claim here has actually been
contradicted in review, not on principle.** This decision is recorded so the
question does not get re-opened on a hunch.

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
  resources whose control is present.
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