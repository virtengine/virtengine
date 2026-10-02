# checkov CKV2_* graph checks and `count` / `for_each`

**Status: the previous version of this document was WRONG and has been
corrected.** It claimed a checkov blind spot on `count` / `for_each`. That claim
is false, and it was being used to explain away real findings — including
`aws_s3_bucket.logs`, which is a genuine exposure. This file now records what
the evidence actually shows.

The gate stays **red**. No `#checkov:skip`, `soft_fail` stays `false`, no
`skip-check` list was introduced, and no finding was suppressed.

## What the evidence shows

`CKV2_AWS_6` reports resources declared with `count` and `for_each` **in both
directions** — it fails some indexed resources and passes others. A resource
declared under `count` is not invisible to it.

From the repository's own CI (run `37000014741`, job `110816060492`, Security
Scan / Checkov on `develop @ 75276ed9d`), parsed by walking the log in order and
attributing each result to the preceding `Check: <ID>:` header:

| resource | CKV2_AWS_6 |
|---|---|
| `module.s3.aws_s3_bucket.ml_models[0]` | **FAILED** |
| `aws_s3_bucket.logs[0]` | **FAILED** |
| `aws_s3_bucket.dr_results`, `aws_s3_bucket.backup_primary`, `aws_s3_bucket.backup_secondary`, `module.s3.aws_s3_bucket.chain_backups`, `module.s3.aws_s3_bucket.manifests`, `module.s3.aws_s3_bucket.terraform_state`, `module.database.aws_s3_bucket.backups`, `aws_s3_bucket.terraform_state` | PASSED |

Both failures are on **indexed** resources. Across the whole run: 674 indexed
results, 37 FAILED, 637 PASSED — evaluated in both directions.

## Reproduction

Public `aws_s3_bucket` with **no** public-access-block, wiring identical except
for how the resource is declared:

| declaration | CKV2_AWS_6 |
|---|---|
| `count = 1` | FAILED (`aws_s3_bucket.c[0]`) |
| `count = 0` (literal) | FAILED |
| `count = var.n` (no default) | FAILED |
| `count = var.flag ? 1 : 0`, `flag = true` (resource exists, public) | **FAILED** (`aws_s3_bucket.h[0]`) |
| `count = var.flag ? 1 : 0`, `flag = false` (resource absent) | FAILED |
| `for_each = toset(["x"])` | FAILED (`aws_s3_bucket.d["x"]`) |
| bucket **with** a public-access-block, `count = 1` | PASSED |

```bash
checkov -d . --check CKV2_AWS_6 --compact
```

The two claims that the earlier version of this document made, and why they are
retracted:

* *"`count = 0` passes vacuously"* — it does not. `count = 0` fails, exactly
  like `count = 1`.
* *"a resource declared with `count` or `for_each` is invisible to these
  checks"* — it is not. The dangerous direction was tested directly: a public
  bucket declared under `count = var.flag ? 1 : 0` with `flag = true` is
  **reported as a failure**.

## Blast radius — correction

The earlier "72 of 137 (52.6%) sit on a resource carrying `count` or `for_each`"
is **not substantiated** and has been withdrawn. That correlation cannot be
derived from the CI log at all (checkov logs resource names, not their
`count`/`for_each`), and the local reproduction behind it behaved differently
from CI. Findings on conditional resources are real findings.

## Disposition

The gate is left red on purpose. Every remaining finding — including the
indexed ones such as `aws_s3_bucket.logs` — is a real control that is genuinely
absent, and is fixed on its own merits.

Note for anyone parsing this output: checkov interleaves `PASSED`/`FAILED` lines
under a moving `Check:` header. Never pair a result with a check id by line
proximity; walk the log in order.

Refs: card `t_38224eaa` (authorisation to correct: project-steward, 2026-10-02),
PR #1146 (superseded claim), PR #1145, #1151, #1153.