# checkov CKV2_* graph checks are blind to `count` / `for_each`

**Status: confirmed on checkov 3.3.22 (the version `bridgecrewio/checkov-action@v12` pulls).**
This is a scanner limitation, not a Terraform defect, and it fails in the
dangerous direction — it produces **false negatives**.

## What happens

The `CKV2_*` checks are declared as JSON with a `connection` +
`operator: exists` predicate. For example `CKV2_AWS_6`
(`checkov/terraform/checks/graph_checks/aws/S3BucketHasPublicAccessBlock.json`)
requires an `aws_s3_bucket` connected to an `aws_s3_bucket_public_access_block`.
The resolver does **not** build that graph edge when the reference is indexed,
i.e. `bucket = aws_s3_bucket.y[0].id`.

So a resource declared with `count` or `for_each` is invisible to these checks.

## Proof

A fixture whose wiring is **byte-identical** across rows, differing only in the
`count` expression:

| `count` expression             | CKV2_AWS_6 |
|--------------------------------|------------|
| *(no `count`)*                 | PASSED     |
| `var.flag_false ? 1 : 0`       | PASSED     |
| `0`                            | PASSED     |
| `1`                            | **FAILED** |
| `var.flag_true ? 1 : 0`        | **FAILED** |
| `for_each` on both resources   | **FAILED** |

Reproduce with any Terraform file laid out like that:

```bash
python -m checkov.main -d . --framework terraform --check CKV2_AWS_6 --output cli
```

## The false-negative direction

`count = 0` "passes" **vacuously** — a resource that is not created is trivially
secured. So the check both misses real exposure *and* manufactures a green for
resources that do not exist.

This repo is its own control, in `modules/s3/main.tf`:

| resource                           | `count` var           | default | result   |
|------------------------------------|-----------------------|---------|----------|
| `aws_s3_bucket.ml_models[0]`       | `create_ml_bucket`    | `true`  | FAILED   |
| `aws_s3_bucket.terraform_state[0]` | `create_state_bucket` | `false` | "PASSED" |

Both have a complete `aws_s3_bucket_public_access_block`. The only difference is
whether the resource exists.

## Blast radius

Of the 137 findings on `develop @ 3a7abee10` (137 failed, 74 resources, 38 check
IDs), **72 (52.6%)** sit on a resource carrying `count` or `for_each`.

Affected checks include `CKV2_AWS_6` (S3 public access block), `CKV2_AWS_61`
(lifecycle), `CKV2_AWS_62` (event notifications), `CKV2_AWS_64` (KMS key
policy), `CKV2_AWS_5` (security group attached), `CKV2_AWS_23` (Route53 alias),
`CKV2_AWS_12` (default security group), `CKV2_AWS_38`, `CKV2_AWS_39`,
`CKV2_AWS_57`, `CKV2_AWS_60`.

**Every conditional or optional resource in this estate is currently unscanned
by those checks.**

## Disposition

The gate is left **red**. No `#checkov:skip`, `soft_fail` stays `false`, no
`skip-check` list was introduced.

Rationale (premium panel, 1 of 2 models answered, rc=0): a skip converts "the
scanner cannot see this" into "we assert this is fine", which is strictly worse
than a red gate — the red gate at least says *look*. Suppressing them would also
erase the fact that these resources are unscanned.

The remaining **real** findings (attributes genuinely absent) are tracked
separately and are fixed on their own merits.

Refs: card `t_caefc334`, PR #1145.
